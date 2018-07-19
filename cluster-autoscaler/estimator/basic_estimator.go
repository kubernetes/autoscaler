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

package estimator

import (
	"bytes"
	"fmt"
	"math"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	schedulercache "k8s.io/kubernetes/pkg/scheduler/cache"
)

// BasicNodeEstimator estimates the number of needed nodes to handle the given amount of pods.
// It will never overestimate the number of nodes but is quite likely to provide a number that
// is too small.
type BasicNodeEstimator struct {
	cpuSum      resource.Quantity
	memorySum   resource.Quantity
	portSum     map[int32]int
	FittingPods map[*apiv1.Pod]struct{}
}

// NewBasicNodeEstimator builds BasicNodeEstimator.
func NewBasicNodeEstimator() *BasicNodeEstimator {
	return &BasicNodeEstimator{
		portSum:     make(map[int32]int),
		FittingPods: make(map[*apiv1.Pod]struct{}),
	}
}

// Add adds Pod to the estimation.
func (basicEstimator *BasicNodeEstimator) Add(pod *apiv1.Pod) error {
	ports := make(map[int32]struct{})
	for _, container := range pod.Spec.Containers {
		if request, ok := container.Resources.Requests[apiv1.ResourceCPU]; ok {
			basicEstimator.cpuSum.Add(request)
		}
		if request, ok := container.Resources.Requests[apiv1.ResourceMemory]; ok {
			basicEstimator.memorySum.Add(request)
		}
		for _, port := range container.Ports {
			if port.HostPort > 0 {
				ports[port.HostPort] = struct{}{}
			}
		}
	}
	for port := range ports {
		if sum, ok := basicEstimator.portSum[port]; ok {
			basicEstimator.portSum[port] = sum + 1
		} else {
			basicEstimator.portSum[port] = 1
		}
	}
	basicEstimator.FittingPods[pod] = struct{}{}
	return nil
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// GetDebug returns debug information about the current state of BasicNodeEstimator
func (basicEstimator *BasicNodeEstimator) GetDebug() string {
	var buffer bytes.Buffer
	buffer.WriteString("Resources needed:\n")
	buffer.WriteString(fmt.Sprintf("CPU: %s\n", basicEstimator.cpuSum.String()))
	buffer.WriteString(fmt.Sprintf("Mem: %s\n", basicEstimator.memorySum.String()))
	for port, count := range basicEstimator.portSum {
		buffer.WriteString(fmt.Sprintf("Port %d: %d\n", port, count))
	}
	return buffer.String()
}

// Estimate estimates the number needed of nodes of the given shape.
func (basicEstimator *BasicNodeEstimator) Estimate(pods []*apiv1.Pod, nodeInfo *schedulercache.NodeInfo, upcomingNodes []*schedulercache.NodeInfo) int {
	for _, pod := range pods {
		basicEstimator.Add(pod)
	}

	result := 0
	resources := apiv1.ResourceList{}
	for _, node := range upcomingNodes {
		cpu := resources[apiv1.ResourceCPU]
		cpu.Add(node.Node().Status.Capacity[apiv1.ResourceCPU])
		resources[apiv1.ResourceCPU] = cpu

		mem := resources[apiv1.ResourceMemory]
		mem.Add(node.Node().Status.Capacity[apiv1.ResourceMemory])
		resources[apiv1.ResourceMemory] = mem

		pods := resources[apiv1.ResourcePods]
		pods.Add(node.Node().Status.Capacity[apiv1.ResourcePods])
		resources[apiv1.ResourcePods] = pods
	}

	node := nodeInfo.Node()
	if cpuCapacity, ok := node.Status.Capacity[apiv1.ResourceCPU]; ok {
		comingCpu := resources[apiv1.ResourceCPU]
		prop := int(math.Ceil(float64(
			basicEstimator.cpuSum.MilliValue()-comingCpu.MilliValue()) /
			float64(cpuCapacity.MilliValue())))

		result = maxInt(result, prop)
	}
	if memCapacity, ok := node.Status.Capacity[apiv1.ResourceMemory]; ok {
		comingMem := resources[apiv1.ResourceMemory]
		prop := int(math.Ceil(float64(
			basicEstimator.memorySum.Value()-comingMem.Value()) /
			float64(memCapacity.Value())))
		result = maxInt(result, prop)
	}
	if podCapacity, ok := node.Status.Capacity[apiv1.ResourcePods]; ok {
		comingPods := resources[apiv1.ResourcePods]
		prop := int(math.Ceil(float64(basicEstimator.GetCount()-int(comingPods.Value())) /
			float64(podCapacity.Value())))
		result = maxInt(result, prop)
	}
	for _, count := range basicEstimator.portSum {
		result = maxInt(result, count-len(upcomingNodes))
	}
	return result
}

// GetCount returns number of pods included in the estimation.
func (basicEstimator *BasicNodeEstimator) GetCount() int {
	return len(basicEstimator.FittingPods)
}
