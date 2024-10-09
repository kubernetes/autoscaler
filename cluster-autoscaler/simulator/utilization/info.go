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

package utilization

import (
	"fmt"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/dynamicresources"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/utils/drain"
	pod_util "k8s.io/autoscaler/cluster-autoscaler/utils/pod"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	resourcehelper "k8s.io/kubernetes/pkg/api/v1/resource"

	klog "k8s.io/klog/v2"
)

// Info contains utilization information for a node.
type Info struct {
	CpuUtil             float64
	MemUtil             float64
	GpuUtil             float64
	DynamicResourceUtil float64
	// Resource name of highest utilization resource
	ResourceName apiv1.ResourceName
	// Max(CpuUtil, MemUtil) or GpuUtils
	Utilization float64
}

// Calculate calculates utilization of a node, defined as maximum of (cpu,
// memory) or gpu utilization based on if the node has GPU or not. Per resource
// utilization is the sum of requests for it divided by allocatable. It also
// returns the individual cpu, memory and gpu utilization.
func Calculate(nodeInfo *framework.NodeInfo, skipDaemonSetPods, skipMirrorPods, draEnabled bool, gpuConfig *cloudprovider.GpuConfig, currentTime time.Time) (utilInfo Info, err error) {
	if gpuConfig != nil {
		gpuUtil, err := CalculateUtilizationOfResource(nodeInfo, gpuConfig.ResourceName, skipDaemonSetPods, skipMirrorPods, currentTime)
		if err != nil {
			klog.V(3).Infof("node %s has unready GPU resource: %s", nodeInfo.Node().Name, gpuConfig.ResourceName)
			// Return 0 if GPU is unready. This will guarantee we can still scale down a node with unready GPU.
			return Info{GpuUtil: 0, ResourceName: gpuConfig.ResourceName, Utilization: 0}, nil
		}
		// Skips cpu and memory utilization calculation for node with GPU.
		return Info{GpuUtil: gpuUtil, ResourceName: gpuConfig.ResourceName, Utilization: gpuUtil}, err
	}

	if draEnabled && len(nodeInfo.LocalResourceSlices) > 0 {
		resourceName, highestUtil, err := dynamicresources.HighestDynamicResourceUtil(nodeInfo)
		if err != nil {
			return Info{}, err
		}
		return Info{DynamicResourceUtil: highestUtil, Utilization: highestUtil, ResourceName: resourceName}, nil
	}

	cpu, err := CalculateUtilizationOfResource(nodeInfo, apiv1.ResourceCPU, skipDaemonSetPods, skipMirrorPods, currentTime)
	if err != nil {
		return Info{}, err
	}
	mem, err := CalculateUtilizationOfResource(nodeInfo, apiv1.ResourceMemory, skipDaemonSetPods, skipMirrorPods, currentTime)
	if err != nil {
		return Info{}, err
	}

	utilization := Info{CpuUtil: cpu, MemUtil: mem}

	if cpu > mem {
		utilization.ResourceName = apiv1.ResourceCPU
		utilization.Utilization = cpu
	} else {
		utilization.ResourceName = apiv1.ResourceMemory
		utilization.Utilization = mem
	}

	return utilization, nil
}

// CalculateUtilizationOfResource calculates utilization of a given resource for a node.
func CalculateUtilizationOfResource(nodeInfo *framework.NodeInfo, resourceName apiv1.ResourceName, skipDaemonSetPods, skipMirrorPods bool, currentTime time.Time) (float64, error) {
	nodeAllocatable, found := nodeInfo.Node().Status.Allocatable[resourceName]
	if !found {
		return 0, fmt.Errorf("failed to get %v from %s", resourceName, nodeInfo.Node().Name)
	}
	if nodeAllocatable.MilliValue() == 0 {
		return 0, fmt.Errorf("%v is 0 at %s", resourceName, nodeInfo.Node().Name)
	}

	opts := resourcehelper.PodResourcesOptions{}

	// if skipDaemonSetPods = True, DaemonSet pods resourses will be subtracted
	// from the node allocatable and won't be added to pods requests
	// the same with the Mirror pod.
	podsRequest := resource.MustParse("0")
	daemonSetAndMirrorPodsUtilization := resource.MustParse("0")
	for _, podInfo := range nodeInfo.Pods {
		requestedResourceList := resourcehelper.PodRequests(podInfo.Pod, opts)
		resourceValue := requestedResourceList[resourceName]

		// factor daemonset pods out of the utilization calculations
		if skipDaemonSetPods && pod_util.IsDaemonSetPod(podInfo.Pod) {
			daemonSetAndMirrorPodsUtilization.Add(resourceValue)
			continue
		}

		// factor mirror pods out of the utilization calculations
		if skipMirrorPods && pod_util.IsMirrorPod(podInfo.Pod) {
			daemonSetAndMirrorPodsUtilization.Add(resourceValue)
			continue
		}

		// ignore Pods that should be terminated
		if drain.IsPodLongTerminating(podInfo.Pod, currentTime) {
			continue
		}

		podsRequest.Add(resourceValue)
	}

	return float64(podsRequest.MilliValue()) / float64(nodeAllocatable.MilliValue()-daemonSetAndMirrorPodsUtilization.MilliValue()), nil
}
