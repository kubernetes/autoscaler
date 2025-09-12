/*
Copyright 2022 The Kubernetes Authors.

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

package routines

import (
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	vpa_utils "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/vpa"
)

// MemoryPerCPUPostProcessor enforces a fixed memory-per-CPU ratio for each container's recommendation.
// The ratio is defined in the container's policy as MemoryPerCPU (bytes per 1 CPU core).
// Applied to Target, LowerBound, UpperBound, and UncappedTarget.
type MemoryPerCPUPostProcessor struct{}

var _ RecommendationPostProcessor = &MemoryPerCPUPostProcessor{}

// Process applies the memory-per-CPU enforcement to the recommendation if specified in the container policy.
func (p *MemoryPerCPUPostProcessor) Process(
	vpa *vpa_types.VerticalPodAutoscaler,
	recommendation *vpa_types.RecommendedPodResources,
) *vpa_types.RecommendedPodResources {
	if vpa == nil || vpa.Spec.ResourcePolicy == nil || recommendation == nil {
		return recommendation
	}

	amendedRecommendation := recommendation.DeepCopy()

	for _, r := range amendedRecommendation.ContainerRecommendations {
		pol := vpa_utils.GetContainerResourcePolicy(r.ContainerName, vpa.Spec.ResourcePolicy)
		if pol != nil && pol.MemoryPerCPU != nil {
			memPerCPUBytes := pol.MemoryPerCPU.Value()
			r.Target = enforceMemoryPerCPU(r.Target, memPerCPUBytes)
			r.LowerBound = enforceMemoryPerCPU(r.LowerBound, memPerCPUBytes)
			r.UpperBound = enforceMemoryPerCPU(r.UpperBound, memPerCPUBytes)
			r.UncappedTarget = enforceMemoryPerCPU(r.UncappedTarget, memPerCPUBytes)
		}
	}

	return amendedRecommendation
}

// enforceMemoryPerCPU adjusts CPU or Memory to satisfy:
//
//	memory_bytes = cpu_cores * memPerCPUBytes
//
// If memory is too low for the given CPU, increase memory.
// If memory is too high for the given CPU, increase CPU.
// enforceMemoryPerCPU adjusts CPU or Memory to satisfy:
//
//	memory_bytes = cpu_cores * memPerCPUBytes
//
// If memory is too low for the given CPU, increase memory.
// If memory is too high for the given CPU, increase CPU.
func enforceMemoryPerCPU(resources apiv1.ResourceList, bytesPerCore int64) apiv1.ResourceList {
	if bytesPerCore <= 0 {
		return resources
	}

	cpuQty, hasCPU := resources[apiv1.ResourceCPU]
	memQty, hasMem := resources[apiv1.ResourceMemory]
	if !hasCPU || !hasMem || cpuQty.IsZero() || memQty.IsZero() {
		return resources
	}

	// cpuCores = milliCPU / 1000
	cpuMilli := cpuQty.MilliValue()
	memBytes := memQty.Value()

	// Desired memory in bytes = CPU cores * bytes per core
	desiredMem := divCeil(cpuMilli*bytesPerCore, 1000)

	if memBytes < desiredMem {
		// Not enough RAM → increase memory
		resources[apiv1.ResourceMemory] = *resource.NewQuantity(desiredMem, resource.BinarySI)
	} else if memBytes > desiredMem {
		// Too much RAM → increase CPU
		desiredMilli := divCeil(memBytes*1000, bytesPerCore)
		resources[apiv1.ResourceCPU] = *resource.NewMilliQuantity(desiredMilli, resource.DecimalSI)
	}

	return resources
}

func divCeil(a, b int64) int64 {
	return (a + b - 1) / b
}
