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
	"fmt"
	"io/ioutil"
	"os"
	"time"

	flag "github.com/spf13/pflag"

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/addon-resizer/nanny"

	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/autoscaler/addon-resizer/nanny/apis/nannyconfig"
	nannyscheme "k8s.io/autoscaler/addon-resizer/nanny/apis/nannyconfig/scheme"
	nannyconfigalpha "k8s.io/autoscaler/addon-resizer/nanny/apis/nannyconfig/v1alpha1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"path/filepath"
)

var (
	// Flags to define the resource requirements.
	configDir = flag.String("config-dir", nannyconfig.NoValue, "Path of configuration containing base resource requirements.")
	// Following empty values ("") will be overwritten by defaults specified in apis/nannyconfig/v1alpha1/defaults.go
	baseCPU        = flag.String("cpu", "", "The base CPU resource requirement.")
	cpuPerNode     = flag.String("extra-cpu", "", "The amount of CPU to add per node.")
	baseMemory     = flag.String("memory", "", "The base memory resource requirement.")
	memoryPerNode  = flag.String("extra-memory", "", "The amount of memory to add per node.")
	baseStorage    = flag.String("storage", nannyconfig.NoValue, "The base storage resource requirement.")
	storagePerNode = flag.String("extra-storage", "0Gi", "The amount of storage to add per node.")
	threshold      = flag.Int("threshold", 0, "A number between 0-100. The dependent's resources are rewritten when they deviate from expected by more than threshold.")
	// Flags to identify the container to nanny.
	podNamespace  = flag.String("namespace", os.Getenv("MY_POD_NAMESPACE"), "The namespace of the ward. This defaults to the nanny pod's own namespace.")
	deployment    = flag.String("deployment", "", "The name of the deployment being monitored. This is required.")
	podName       = flag.String("pod", os.Getenv("MY_POD_NAME"), "The name of the pod to watch. This defaults to the nanny's own pod.")
	containerName = flag.String("container", "pod-nanny", "The name of the container to watch. This defaults to the nanny itself.")
	// Flags to control runtime behavior.
	pollPeriod     = time.Millisecond * time.Duration(*flag.Int("poll-period", 10000, "The time, in milliseconds, to poll the dependent container."))
	estimator      = flag.String("estimator", "linear", "The estimator to use. Currently supported: linear, exponential")
	minClusterSize = flag.Uint64("minClusterSize", 16, "The smallest number of nodes resources will be scaled to. Must be > 1. This flag is used only when an exponential estimator is used.")
)

func main() {
	// First log our starting config, and then set up.
	glog.Infof("Invoked by %v", os.Args)
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
	glog.Infof("storage: %s, extra_storage: %s", *baseStorage, *storagePerNode)

	// Set up work objects.
	config, err := rest.InClusterConfig()
	if err != nil {
		glog.Fatal(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		glog.Fatal(err)
	}
	// Use protobufs to improve performance.
	config.ContentType = "application/vnd.kubernetes.protobuf"

	k8s := nanny.NewKubernetesClient(*podNamespace, *deployment, *podName, *containerName, clientset)

	nannyConfigurationFromFlags := &nannyconfigalpha.NannyConfiguration{
		BaseCPU:       *baseCPU,
		CPUPerNode:    *cpuPerNode,
		BaseMemory:    *baseMemory,
		MemoryPerNode: *memoryPerNode,
	}
	nannycfg, err := loadNannyConfiguration(*configDir, nannyConfigurationFromFlags)
	if err != nil {
		glog.Fatal(err)
	}
	glog.Infof("cpu: %s, extra_cpu: %s, memory: %s, extra_memory: %s", nannycfg.BaseCPU, nannycfg.CPUPerNode, nannycfg.BaseMemory, nannycfg.MemoryPerNode)

	var resources []nanny.Resource

	// Monitor only the resources specified.
	if nannycfg.BaseCPU != nannyconfig.NoValue {
		resources = append(resources, nanny.Resource{
			Base:         resource.MustParse(nannycfg.BaseCPU),
			ExtraPerNode: resource.MustParse(nannycfg.CPUPerNode),
			Name:         "cpu",
		})
	}

	if nannycfg.BaseMemory != nannyconfig.NoValue {
		resources = append(resources, nanny.Resource{
			Base:         resource.MustParse(nannycfg.BaseMemory),
			ExtraPerNode: resource.MustParse(nannycfg.MemoryPerNode),
			Name:         "memory",
		})
	}

	if *baseStorage != nannyconfig.NoValue {
		resources = append(resources, nanny.Resource{
			Base:         resource.MustParse(*baseStorage),
			ExtraPerNode: resource.MustParse(nannycfg.MemoryPerNode),
			Name:         "storage",
		})
	}

	glog.Infof("Resources: %+v", resources)

	var est nanny.ResourceEstimator
	if *estimator == "linear" {
		est = nanny.LinearEstimator{
			Resources: resources,
		}
	} else if *estimator == "exponential" {
		est = nanny.ExponentialEstimator{
			Resources:      resources,
			ScaleFactor:    1.5,
			MinClusterSize: *minClusterSize,
		}
	} else {
		glog.Fatalf("Estimator %s not supported", *estimator)
	}

	// Begin nannying.
	nanny.PollAPIServer(k8s, est, *containerName, pollPeriod, uint64(*threshold))
}

func loadNannyConfiguration(configDir string, defaultConfig *nannyconfigalpha.NannyConfiguration) (*nannyconfig.NannyConfiguration, error) {
	path := filepath.Join(configDir, "NannyConfiguration")
	scheme, codecs, err := nannyscheme.NewSchemeAndCodecs()
	if err != nil {
		return nil, err
	}
	// overwrite defaults with flag-specified parameters
	nannyconfigalpha.SetDefaults_NannyConfiguration(defaultConfig)
	// retrieve config map parameters if present
	configMapConfig := &nannyconfigalpha.NannyConfiguration{}
	data, err := ioutil.ReadFile(path)
	if err != nil {
		glog.V(0).Infof("Failed to read data from config file %q: %v, using default parameters", path, err)
	} else if configMapConfig, err = decodeNannyConfiguration(data, scheme, codecs); err != nil {
		configMapConfig = &nannyconfigalpha.NannyConfiguration{}
		glog.V(0).Infof("Unable to decode Nanny Configuration from config map, using default parameters")
	}
	nannyconfigalpha.SetDefaults_NannyConfiguration(configMapConfig)
	// overwrite defaults with config map parameters
	nannyconfigalpha.FillInDefaults_NannyConfiguration(configMapConfig, defaultConfig)
	return convertNannyConfiguration(configMapConfig, scheme)
}

func convertNannyConfiguration(configAlpha *nannyconfigalpha.NannyConfiguration, scheme *runtime.Scheme) (*nannyconfig.NannyConfiguration, error) {
	config := &nannyconfig.NannyConfiguration{}
	err := scheme.Convert(configAlpha, config, nil)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func decodeNannyConfiguration(data []byte, scheme *runtime.Scheme, codecs *serializer.CodecFactory) (*nannyconfigalpha.NannyConfiguration, error) {
	obj, err := runtime.Decode(codecs.UniversalDecoder(nannyconfigalpha.SchemeGroupVersion), data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode, error: %v", err)
	}
	externalHC, ok := obj.(*nannyconfigalpha.NannyConfiguration)
	if !ok {
		return nil, fmt.Errorf("failed to cast object to NannyConfiguration, object: %#v", obj)
	}
	return externalHC, nil
}
