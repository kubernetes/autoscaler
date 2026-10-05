/*
Copyright 2015 The Kubernetes Authors.

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

// This is a cut down fork of k8s.io/kubernetes/test/e2e/common/autoscaling_utils.go

package autoscaling

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/autoscaler/vertical-pod-autoscaler/test/e2e/utils"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/test/e2e/framework"
	e2edebug "k8s.io/kubernetes/test/e2e/framework/debug"
	e2ekubectl "k8s.io/kubernetes/test/e2e/framework/kubectl"
	e2erc "k8s.io/kubernetes/test/e2e/framework/rc"
	"k8s.io/kubernetes/test/e2e/framework/resource"
	testutils "k8s.io/kubernetes/test/utils"

	ginkgo "github.com/onsi/ginkgo/v2"

	imageutils "k8s.io/kubernetes/test/utils/image"
)

const (
	dynamicConsumptionTimeInSeconds = 30
	targetPort                      = 8080
	timeoutRC                       = 120 * time.Second
	invalidKind                     = "ERROR: invalid workload kind for resource consumer"
	serviceInitializationTimeout    = 2 * time.Minute
	serviceInitializationInterval   = 15 * time.Second
	stressImage                     = "registry.k8s.io/e2e-test-images/agnhost:2.53"
)

var (
	resourceConsumerImage = imageutils.GetE2EImage(imageutils.ResourceConsumer)
)

var (
	// KindRC is the GVK for ReplicationController
	KindRC = schema.GroupVersionKind{Version: "v1", Kind: "ReplicationController"}
	// KindDeployment is the GVK for Deployment
	KindDeployment = schema.GroupVersionKind{Group: "apps", Version: "v1beta2", Kind: "Deployment"}
	// KindReplicaSet is the GVK for ReplicaSet
	KindReplicaSet = schema.GroupVersionKind{Group: "apps", Version: "v1beta2", Kind: "ReplicaSet"}
)

/*
ResourceConsumer is a tool for testing. It helps create specified usage of CPU or memory.
Load is sent directly to each pod via the Kubernetes pod proxy API, so no
service, kube-proxy or resource-consumer-controller pod is involved.
typical use case:
rc.ConsumeCPUPerPod(600)
// ... check your assumption here
rc.ConsumeCPUPerPod(300)
// ... check your assumption here
*/
type ResourceConsumer struct {
	name                     string
	kind                     schema.GroupVersionKind
	nsName                   string
	clientSet                clientset.Interface
	cpuPerPod                chan int
	memPerPod                chan int
	stopCPUPerPod            chan int
	stopMemPerPod            chan int
	stopWaitGroup            sync.WaitGroup
	consumptionTimeInSeconds int
	sleepTime                time.Duration
}

// NewDynamicResourceConsumer is a wrapper to create a new dynamic ResourceConsumer.
func NewDynamicResourceConsumer(name, nsName string, kind schema.GroupVersionKind, replicas, initCPUTotal, initMemoryTotal int, cpuLimit, memLimit int64, clientset clientset.Interface) *ResourceConsumer {
	return newResourceConsumer(name, nsName, kind, replicas, initCPUTotal, initMemoryTotal, dynamicConsumptionTimeInSeconds,
		cpuLimit, memLimit, clientset, nil)
}

/*
newResourceConsumer creates new ResourceConsumer
initCPUTotal argument is in millicores
initMemoryTotal argument is in megabytes
memLimit argument is in megabytes, memLimit is a maximum amount of memory that can be consumed by a single pod
cpuLimit argument is in millicores, cpuLimit is a maximum amount of cpu that can be consumed by a single pod
*/
func newResourceConsumer(name, nsName string, kind schema.GroupVersionKind, replicas, initCPUTotal, initMemoryTotal, consumptionTimeInSeconds int,
	cpuLimit, memLimit int64, clientset clientset.Interface, podAnnotations map[string]string) *ResourceConsumer {
	if podAnnotations == nil {
		podAnnotations = make(map[string]string)
	}
	runWorkloadForResourceConsumer(clientset, nsName, name, kind, replicas, cpuLimit, memLimit, podAnnotations)
	rc := &ResourceConsumer{
		name:                     name,
		kind:                     kind,
		nsName:                   nsName,
		clientSet:                clientset,
		cpuPerPod:                make(chan int),
		memPerPod:                make(chan int),
		stopCPUPerPod:            make(chan int),
		stopMemPerPod:            make(chan int),
		consumptionTimeInSeconds: consumptionTimeInSeconds,
		sleepTime:                time.Duration(consumptionTimeInSeconds) * time.Second,
	}

	rc.stopWaitGroup.Add(1)
	go rc.makeConsumeCPUPerPodRequests()
	rc.ConsumeCPUPerPod(initCPUTotal)
	rc.stopWaitGroup.Add(1)
	go rc.makeConsumeMemPerPodRequests()
	rc.ConsumeMemPerPod(initMemoryTotal)
	return rc
}

// ConsumeCPUPerPod sends CPU load directly to each consumer pod via the
// Kubernetes pod proxy API, bypassing kube-proxy's non-deterministic load
// balancing. millicoresTotal is divided evenly across all running pods,
// guaranteeing each pod receives an equal share.
func (rc *ResourceConsumer) ConsumeCPUPerPod(millicoresTotal int) {
	framework.Logf("RC %s: consume %v millicores in total (evenly distributed per pod)", rc.name, millicoresTotal)
	rc.cpuPerPod <- millicoresTotal
}

// ConsumeMemPerPod is the memory equivalent of ConsumeCPUPerPod:
// megabytesTotal is divided evenly across all running pods and sent directly
// via the Kubernetes pod proxy API.
func (rc *ResourceConsumer) ConsumeMemPerPod(megabytesTotal int) {
	framework.Logf("RC %s: consume %v MB in total (evenly distributed per pod)", rc.name, megabytesTotal)
	rc.memPerPod <- megabytesTotal
}

func (rc *ResourceConsumer) makeConsumeCPUPerPodRequests() {
	defer ginkgo.GinkgoRecover()
	defer rc.stopWaitGroup.Done()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// Cancel in-flight requests as soon as the stop channel closes, so CleanUp
	// isn't blocked waiting for a long poll to finish.
	go func() {
		<-rc.stopCPUPerPod
		cancel()
	}()
	sleepTime := time.Duration(0)
	millicoresTotal := 0
	for {
		select {
		case millicoresTotal = <-rc.cpuPerPod:
			framework.Logf("RC %s: setting per-pod CPU to %v millicores total", rc.name, millicoresTotal)
		case <-time.After(sleepTime):
			if millicoresTotal != 0 {
				framework.Logf("RC %s: sending per-pod CPU request: %d millicores total", rc.name, millicoresTotal)
				rc.sendConsumeCPUPerPodRequest(ctx, millicoresTotal)
			}
			sleepTime = rc.sleepTime
		case <-rc.stopCPUPerPod:
			framework.Logf("RC %s: stopping per-pod CPU consumer", rc.name)
			return
		}
	}
}

func (rc *ResourceConsumer) makeConsumeMemPerPodRequests() {
	defer ginkgo.GinkgoRecover()
	defer rc.stopWaitGroup.Done()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// Cancel in-flight requests as soon as the stop channel closes, so CleanUp
	// isn't blocked waiting for a long poll to finish.
	go func() {
		<-rc.stopMemPerPod
		cancel()
	}()
	sleepTime := time.Duration(0)
	megabytesTotal := 0
	for {
		select {
		case megabytesTotal = <-rc.memPerPod:
			framework.Logf("RC %s: setting per-pod mem to %v MB total", rc.name, megabytesTotal)
		case <-time.After(sleepTime):
			if megabytesTotal != 0 {
				framework.Logf("RC %s: sending per-pod mem request: %d MB total", rc.name, megabytesTotal)
				rc.sendConsumeMemPerPodRequest(ctx, megabytesTotal)
			}
			sleepTime = rc.sleepTime
		case <-rc.stopMemPerPod:
			framework.Logf("RC %s: stopping per-pod mem consumer", rc.name)
			return
		}
	}
}

func (rc *ResourceConsumer) sendConsumeCPUPerPodRequest(ctx context.Context, millicoresTotal int) {
	rc.sendConsumePerPodRequests(ctx, "ConsumeCPU", "millicores", millicoresTotal)
}

func (rc *ResourceConsumer) sendConsumeMemPerPodRequest(ctx context.Context, megabytesTotal int) {
	rc.sendConsumePerPodRequests(ctx, "ConsumeMem", "megabytes", megabytesTotal)
}

// sendConsumePerPodRequests distributes load evenly across all running pods
// by sending requests directly via the Kubernetes pod proxy API. This
// bypasses kube-proxy load balancing, guaranteeing each pod receives exactly
// its share.
func (rc *ResourceConsumer) sendConsumePerPodRequests(ctx context.Context, endpoint, valueParam string, total int) {
	ctx, cancel := context.WithTimeout(ctx, framework.SingleCallTimeout)
	defer cancel()

	var readyPods []string
	err := wait.PollUntilContextTimeout(ctx, serviceInitializationInterval, serviceInitializationTimeout, true, func(ctx context.Context) (done bool, err error) {
		readyPods = nil
		pods, err := rc.clientSet.CoreV1().Pods(rc.nsName).List(ctx, metav1.ListOptions{
			LabelSelector: fmt.Sprintf("name=%s", rc.name),
		})
		if err != nil {
			framework.Logf("%s per pod: failed to list pods: %v", endpoint, err)
			return false, nil
		}
		for i := range pods.Items {
			pod := &pods.Items[i]
			if pod.Status.Phase != v1.PodRunning || pod.DeletionTimestamp != nil {
				continue
			}
			for _, cond := range pod.Status.Conditions {
				if cond.Type == v1.PodReady && cond.Status == v1.ConditionTrue {
					readyPods = append(readyPods, pod.Name)
					break
				}
			}
		}
		if len(readyPods) == 0 {
			framework.Logf("%s per pod: no running pods labeled name=%s", endpoint, rc.name)
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		// The pods may be under eviction by the VPA updater. Skip this tick and
		// retry after the next sleep interval.
		framework.Logf("%s per pod: no ready pods to send load to: %v", endpoint, err)
		return
	}

	perPodValue := total / len(readyPods)
	if perPodValue == 0 {
		perPodValue = 1
	}

	framework.Logf("%s per pod: distributing %d %s across %d pods (%d per pod)",
		endpoint, total, valueParam, len(readyPods), perPodValue)

	var wg sync.WaitGroup
	for _, podName := range readyPods {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			// Pod proxy URL: /api/v1/namespaces/{ns}/pods/{podname}:{port}/proxy/{path}
			// Both service and pod proxy support the name:port format. Without an explicit
			// port the API server defaults to port 80, but resource-consumer listens on
			// targetPort (8080), so the port must be specified.
			err := wait.PollUntilContextTimeout(ctx, serviceInitializationInterval, serviceInitializationTimeout, true, func(ctx context.Context) (done bool, err error) {
				_, podErr := rc.clientSet.CoreV1().RESTClient().Post().
					Resource("pods").
					Namespace(rc.nsName).
					Name(fmt.Sprintf("%s:%d", name, targetPort)).
					SubResource("proxy").
					Suffix(endpoint).
					Param(valueParam, strconv.Itoa(perPodValue)).
					Param("durationSec", strconv.Itoa(rc.consumptionTimeInSeconds)).
					DoRaw(ctx)
				if podErr != nil {
					framework.Logf("%s per pod: error sending to pod %s: %v", endpoint, name, podErr)
					return false, nil
				}
				return true, nil
			})
			if err != nil {
				// The pod may have been evicted or recreated by the VPA updater;
				// its replacement will get its share on the next tick.
				framework.Logf("%s per pod: giving up on pod %s: %v", endpoint, name, err)
			}
		}(podName)
	}
	wg.Wait()
}

// CleanUp clean up the background goroutines responsible for consuming resources.
func (rc *ResourceConsumer) CleanUp() {
	ginkgo.By(fmt.Sprintf("Removing consuming RC %s", rc.name))
	close(rc.stopCPUPerPod)
	close(rc.stopMemPerPod)
	rc.stopWaitGroup.Wait()
	kind := rc.kind.GroupKind()
	framework.ExpectNoError(resource.DeleteResourceAndWaitForGC(context.TODO(), rc.clientSet, kind, rc.nsName, rc.name))
}

func runWorkloadForResourceConsumer(c clientset.Interface, ns, name string, kind schema.GroupVersionKind, replicas int, cpuRequestMillis, memRequestMb int64, podAnnotations map[string]string) {
	ginkgo.By(fmt.Sprintf("Running consuming RC %s via %s with %v replicas", name, kind, replicas))

	rcConfig := testutils.RCConfig{
		Client:      c,
		Image:       resourceConsumerImage,
		Name:        name,
		Namespace:   ns,
		Timeout:     timeoutRC,
		Replicas:    replicas,
		CPURequest:  cpuRequestMillis,
		MemRequest:  memRequestMb * 1024 * 1024, // MemRequest is in bytes
		Annotations: podAnnotations,
	}

	switch kind {
	case KindRC:
		framework.ExpectNoError(e2erc.RunRC(context.TODO(), rcConfig))
	case KindDeployment:
		dpConfig := testutils.DeploymentConfig{
			RCConfig: rcConfig,
		}
		ginkgo.By(fmt.Sprintf("creating deployment %s in namespace %s", dpConfig.Name, dpConfig.Namespace))
		dpConfig.NodeDumpFunc = e2edebug.DumpNodeDebugInfo
		dpConfig.ContainerDumpFunc = e2ekubectl.LogFailedContainers
		framework.ExpectNoError(testutils.RunDeployment(context.TODO(), dpConfig))
	case KindReplicaSet:
		rsConfig := testutils.ReplicaSetConfig{
			RCConfig: rcConfig,
		}
		ginkgo.By(fmt.Sprintf("creating replicaset %s in namespace %s", rsConfig.Name, rsConfig.Namespace))
		framework.ExpectNoError(runReplicaSet(rsConfig))
	default:
		framework.Failf(invalidKind)
	}
}

// runReplicaSet launches (and verifies correctness) of a replicaset.
func runReplicaSet(config testutils.ReplicaSetConfig) error {
	ginkgo.By(fmt.Sprintf("creating replicaset %s in namespace %s", config.Name, config.Namespace))
	config.NodeDumpFunc = e2edebug.DumpNodeDebugInfo
	config.ContainerDumpFunc = e2ekubectl.LogFailedContainers
	return testutils.RunReplicaSet(context.TODO(), config)
}

func runOomingReplicationController(c clientset.Interface, ns, name string, replicas int) {
	ginkgo.By(fmt.Sprintf("Running OOMing RC %s with %v replicas", name, replicas))

	rcConfig := testutils.RCConfig{
		Client: c,
		Image:  stressImage,
		// request exactly 1025 MiB, in a single chunk (1 MiB above the limit)
		Command:     []string{"/agnhost", "stress", "--mem-total", "1074790400", "--mem-alloc-size", "1074790400"},
		Name:        name,
		Namespace:   ns,
		Timeout:     timeoutRC,
		Replicas:    replicas,
		Labels:      utils.OOMLabels,
		Annotations: make(map[string]string),
		MemRequest:  1024 * 1024 * 1024,
		MemLimit:    1024 * 1024 * 1024,
	}

	dpConfig := testutils.DeploymentConfig{
		RCConfig: rcConfig,
	}
	ginkgo.By(fmt.Sprintf("Creating deployment %s in namespace %s", dpConfig.Name, dpConfig.Namespace))
	dpConfig.NodeDumpFunc = e2edebug.DumpNodeDebugInfo
	dpConfig.ContainerDumpFunc = e2ekubectl.LogFailedContainers
	// Allow containers to fail (they should be OOM-killed).
	failures := 999
	dpConfig.MaxContainerFailures = &failures
	// Decrease the timeout since the containers are note expected to actually get up.
	dpConfig.Timeout = 10 * time.Second
	dpConfig.PollInterval = 5 * time.Second
	err := testutils.RunDeployment(context.TODO(), dpConfig)
	// Only ignore an error about Pods not starting properly - they're not expected to.
	if err != nil && !strings.Contains(err.Error(), "pods started out of") {
		framework.ExpectNoError(err)
	}
}
