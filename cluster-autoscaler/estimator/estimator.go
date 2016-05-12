/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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
	"math"

	kube_api "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/resource"
)

// BasicNodeEstimator estimates the number of needed nodes to handle the given amount of pods.
// It will never overestimate the number of nodes but is quite likekly to provide a number that
// is too small.
type BasicNodeEstimator struct {
	count     int
	cpuSum    resource.Quantity
	memorySum resource.Quantity
	portSum   map[int32]int
}

// NewBasicNodeEstimator builds BasicNodeEstimator.
func NewBasicNodeEstimator() *BasicNodeEstimator {
	return &BasicNodeEstimator{
		portSum: make(map[int32]int),
	}
}

// Add adds Pod to the estimation.
func (basicEstimator *BasicNodeEstimator) Add(pod *kube_api.Pod) error {
	ports := make(map[int32]struct{})
	for _, container := range pod.Spec.Containers {
		if request, ok := container.Resources.Requests[kube_api.ResourceCPU]; ok {
			if err := basicEstimator.cpuSum.Add(request); err != nil {
				return err
			}
		}
		if request, ok := container.Resources.Requests[kube_api.ResourceMemory]; ok {
			if err := basicEstimator.memorySum.Add(request); err != nil {
				return err
			}
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
	basicEstimator.count++
	return nil
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Estimate estimates the number needed of nodes of the given shape.
func (basicEstimator *BasicNodeEstimator) Estimate(node *kube_api.Node) int {
	result := 0
	if cpuCapcaity, ok := node.Status.Capacity[kube_api.ResourceCPU]; ok {
		result = maxInt(result,
			int(math.Ceil(float64(basicEstimator.cpuSum.MilliValue())/float64(cpuCapcaity.MilliValue()))))
	}
	if memCapcaity, ok := node.Status.Capacity[kube_api.ResourceMemory]; ok {
		result = maxInt(result,
			int(math.Ceil(float64(basicEstimator.memorySum.Value())/float64(memCapcaity.Value()))))
	}
	if podCapcaity, ok := node.Status.Capacity[kube_api.ResourcePods]; ok {
		result = maxInt(result,
			int(math.Ceil(float64(basicEstimator.count)/float64(podCapcaity.Value()))))
	}
	for _, count := range basicEstimator.portSum {
		result = maxInt(result, count)
	}
	return result
}
