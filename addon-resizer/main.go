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
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/golang/glog"

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/addon-resizer/nanny"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const noValue = "MISSING"

var (
	// Flags to define the resource requirements.
	baseCPU              = flag.String("cpu", noValue, "The base CPU resource requirement.")
	cpuPerNode           = flag.String("extra-cpu", "0", "The amount of CPU to add per node.")
	baseMemory           = flag.String("memory", noValue, "The base memory resource requirement.")
	memoryPerNode        = flag.String("extra-memory", "0Mi", "The amount of memory to add per node.")
	baseStorage          = flag.String("storage", noValue, "The base storage resource requirement.")
	storagePerNode       = flag.String("extra-storage", "0Gi", "The amount of storage to add per node.")
	recommendationOffset = flag.Int("recommendation-offset", 10, "A number from range 0-100. When the dependent's resources are rewritten, they are set to the closer end of the range defined by this percentage threshold.")
	acceptanceOffset     = flag.Int("acceptance-offset", 20, "A number from range 0-100. The dependent's resources are rewritten when they deviate from expected by a percentage that is higher than this threshold. Can't be lower than recommendation-offset.")
	// Flags to identify the container to nanny.
	podNamespace  = flag.String("namespace", os.Getenv("MY_POD_NAMESPACE"), "The namespace of the ward. This defaults to the nanny pod's own namespace.")
	deployment    = flag.String("deployment", "", "The name of the deployment being monitored. This is required.")
	podName       = flag.String("pod", os.Getenv("MY_POD_NAME"), "The name of the pod to watch. This defaults to the nanny's own pod.")
	containerName = flag.String("container", "pod-nanny", "The name of the container to watch. This defaults to the nanny itself.")
	// Flags to control runtime behavior.
	pollPeriodMillis = flag.Int("poll-period", 10000, "The time, in milliseconds, to poll the dependent container.")
)

func checkPercentageFlagBounds(flagName string, flagValue int) {
	if flagValue < 0 || flagValue > 100 {
		log.Fatalf("%s flag must be between 0 and 100 inclusively, was %d.", flagName, flagValue)
	}
}

// GetClientOrDie returns a k8s clientset to the request from inside of cluster
func GetClientOrDie() kubernetes.Interface {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("Can not get kubernetes config: %v", err)
	}

	return kubernetes.NewForConfigOrDie(config)
}

func buildOutOfClusterConfig() (*rest.Config, error) {
	kubeconfigPath := os.Getenv("KUBECONFIG")
	if kubeconfigPath == "" {
		kubeconfigPath = os.Getenv("HOME") + "/.kube/config"
	}
	return clientcmd.BuildConfigFromFlags("", kubeconfigPath)
}

// GetClientOutOfClusterOrDie returns a k8s clientset to the request from outside of cluster
func GetClientOutOfClusterOrDie() kubernetes.Interface {
	config, err := buildOutOfClusterConfig()
	if err != nil {
		log.Fatalf("Can not get kubernetes config: %v", err)
	}

	clientset := kubernetes.NewForConfigOrDie(config)

	return clientset
}

func main() {
	// Add standard go flags to the flag set, to enable e.g. setting glog flags.
	flag.Parse()

	// Perform further validation of flags.
	if *deployment == "" {
		log.Fatal("Must specify a deployment.")
	}

	checkPercentageFlagBounds("recommendation-offset", *recommendationOffset)
	checkPercentageFlagBounds("acceptance-offset", *acceptanceOffset)

	pollPeriod := time.Duration(int64(*pollPeriodMillis) * int64(time.Millisecond))
	log.Infof("Poll period: %+v", pollPeriod)
	log.Infof("Watching namespace: %s, pod: %s, container: %s.", *podNamespace, *podName, *containerName)
	log.Infof("cpu: %s, extra_cpu: %s, memory: %s, extra_memory: %s, storage: %s, extra_storage: %s", *baseCPU, *cpuPerNode, *baseMemory, *memoryPerNode, *baseStorage, *storagePerNode)
	log.Infof("Accepted range +/-%d%%", *acceptanceOffset)
	log.Infof("Recommended range +/-%d%%", *recommendationOffset)

	var kubeClient kubernetes.Interface
	_, err := rest.InClusterConfig()
	if err != nil {
		kubeClient = GetClientOutOfClusterOrDie()
	} else {
		kubeClient = GetClientOrDie()
	}

	k8s := nanny.NewKubernetesClient(kubeClient, *podNamespace, *deployment, *podName, *containerName)

	var resources []nanny.Resource

	// Monitor only the resources specified.
	if *baseCPU != noValue {
		resources = append(resources, nanny.Resource{
			Base:         resource.MustParse(*baseCPU),
			ExtraPerNode: resource.MustParse(*cpuPerNode),
			Name:         "cpu",
		})
	}

	if *baseMemory != noValue {
		resources = append(resources, nanny.Resource{
			Base:         resource.MustParse(*baseMemory),
			ExtraPerNode: resource.MustParse(*memoryPerNode),
			Name:         "memory",
		})
	}

	if *baseStorage != noValue {
		resources = append(resources, nanny.Resource{
			Base:         resource.MustParse(*baseStorage),
			ExtraPerNode: resource.MustParse(*memoryPerNode),
			Name:         "storage",
		})
	}

	log.Infof("Resources: %+v", resources)

	// handle termination info
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ch
		log.Infof("Received termination, signaling shutdown")
		k8s.Stop()
		os.Exit(0)
	}()

	// Begin nannying.
	nanny.PollAPIServer(
		k8s,
		nanny.Estimator{
			AcceptanceOffset:     int64(*acceptanceOffset),
			RecommendationOffset: int64(*recommendationOffset),
			Resources:            resources,
		},
		pollPeriod)
}
