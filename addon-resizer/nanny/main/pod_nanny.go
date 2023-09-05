/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"os"
	"time"

	flag "github.com/spf13/pflag"

	"k8s.io/autoscaler/addon-resizer/nanny"

	"path/filepath"

	"github.com/golang/glog"
	"k8s.io/autoscaler/addon-resizer/healthcheck"
	"k8s.io/autoscaler/addon-resizer/nanny/apis/nannyconfig"
	nannyconfigalpha "k8s.io/autoscaler/addon-resizer/nanny/apis/nannyconfig/v1alpha1"
	"k8s.io/autoscaler/addon-resizer/nanny/utils"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	// Flags to define the resource requirements.
	configDir = flag.String("config-dir", nannyconfig.NoValue, "Path of configuration containing base resource requirements.")
	// Following empty values ("") will be overwritten by defaults specified in apis/nannyconfig/v1alpha1/defaults.go
	baseCPU            = flag.String("cpu", "", "The base CPU resource requirement.")
	cpuPerResource     = flag.String("extra-cpu", "", "The amount of CPU to add per resource.")
	baseMemory         = flag.String("memory", "", "The base memory resource requirement.")
	memoryPerResource  = flag.String("extra-memory", "", "The amount of memory to add per resource.")
	baseStorage        = flag.String("storage", nannyconfig.NoValue, "The base storage resource requirement.")
	storagePerResource = flag.String("extra-storage", "0Gi", "The amount of storage to add per resource.")
	scaleDownDelay     = flag.Duration("scale-down-delay", time.Duration(0), "The time to wait after the addon-resizer start or last scaling operation before the scale down can be performed.")
	scaleUpDelay       = flag.Duration("scale-up-delay", time.Duration(0), "The time to wait after the addon-resizer start or last scaling operation before the scale up can be performed.")
	threshold          = flag.Int("threshold", 0, "A number between 0-100. The dependent's resources are rewritten when they deviate from expected by more than threshold.")
	// Flags to identify the container to nanny.
	podNamespace  = flag.String("namespace", os.Getenv("MY_POD_NAMESPACE"), "The namespace of the ward. This defaults to the nanny pod's own namespace.")
	deployment    = flag.String("deployment", "", "The name of the deployment being monitored. This is required.")
	podName       = flag.String("pod", os.Getenv("MY_POD_NAME"), "The name of the pod to watch. This defaults to the nanny's own pod.")
	containerName = flag.String("container", "pod-nanny", "The name of the container to watch. This defaults to the nanny itself.")
	// Flags to control runtime behavior.
	pollPeriod     = flag.Int("poll-period", 10000, "The time, in milliseconds, to poll the dependent container.")
	estimator      = flag.String("estimator", "linear", "The estimator to use. Currently supported: linear, exponential")
	minClusterSize = flag.Uint64("minClusterSize", 16, "The smallest number of resources will be scaled to. Must be > 1. This flag is used only when an exponential estimator is used.")
	useMetrics     = flag.Bool("use-metrics", false, "Whether to use apiserver metrics to detect cluster size instead of the default method of listing objects from the Kubernetes API.")
	hcAddress      = flag.String("healthcheck-address", ":8080", "The address to expose an HTTP health-check on.")
	scalingMode    = flag.String("scaling-mode", nanny.NodeProportional, "The mode of scaling to be used. Possible values: 'node-proportional' or 'container-proportional'")
)

func main() {
	// First log our starting config, and then set up.
	glog.Infof("Invoked by %v", os.Args)
	glog.Infof("Version: %s", nanny.AddonResizerVersion)
	flag.Parse()

	// Perform further validation of flags.
	if *deployment == "" {
		glog.Fatal("Must specify a deployment.")
	}

	if *threshold < 0 || *threshold > 100 {
		glog.Fatalf("Threshold must be between 0 and 100 inclusive. It is %d.", *threshold)
	}

	if *minClusterSize < 2 {
		glog.Fatalf("minClusterSize must be greater than 1. It is set to %d.", *minClusterSize)
	}

	glog.Infof("Watching namespace: %s, pod: %s, container: %s.", *podNamespace, *podName, *containerName)
	glog.Infof("storage: %s, extra_storage: %s", *baseStorage, *storagePerResource)

	// Set up work objects.
	config, err := rest.InClusterConfig()
	if err != nil {
		glog.Fatal(err)
	}
	config.UserAgent = userAgent()
	// Use protobufs for communication with apiserver
	config.ContentType = "application/vnd.kubernetes.protobuf"

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		glog.Fatal(err)
	}
	// Use protobufs to improve performance.
	config.ContentType = "application/vnd.kubernetes.protobuf"

	k8s := nanny.NewKubernetesClient(*podNamespace, *deployment, *podName, *containerName, clientset, *useMetrics)

	nannyConfigurationFromFlags := &nannyconfigalpha.NannyConfiguration{
		BaseCPU:       *baseCPU,
		CPUPerNode:    *cpuPerResource,
		BaseMemory:    *baseMemory,
		MemoryPerNode: *memoryPerResource,
	}

	period := time.Duration(*pollPeriod) * time.Millisecond
	hc := healthcheck.NewHealthCheck(*hcAddress, period*5)
	hc.Serve()

	n := &nanny.Nanny{
		StopChan:       make(chan struct{}, 1),
		ScaleDownDelay: *scaleDownDelay,
		ScaleUpDelay:   *scaleUpDelay,
		PollPeriod:     period,
		HealthCheck:    hc,
		Threshold:      uint64(*threshold),
		Client:         k8s,
		ScalingMode:    *scalingMode,
		EstimatorType:  *estimator,
		MinClusterSize: *minClusterSize,
		ConfigurationFlags:  *nannyConfigurationFromFlags,
		ConfigDir:      *configDir,
		BaseStorage:    *baseStorage,
	}

	n.Run()

	configWatch, err := utils.CreateFsWatcher(time.Second, filepath.Join(*configDir, "NannyConfiguration"))
	if err != nil {
		glog.Fatal(err)
	}

	for {
		select {
		case <-configWatch.Events:
			// stop nanny
			glog.Info("stopping nanny")
			n.Stop()
			// run nanny
			glog.Info("reloading configuration and starting nanny")
			n.Run()
		}
	}
}

func userAgent() string {
	command := ""
	if len(os.Args) > 0 && len(os.Args[0]) > 0 {
		command = filepath.Base(os.Args[0])
	}
	if len(command) == 0 {
		command = "addon-resizer"
	}
	return command + "/" + nanny.AddonResizerVersion
}
