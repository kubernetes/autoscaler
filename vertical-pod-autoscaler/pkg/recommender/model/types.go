/*
Copyright 2017 The Kubernetes Authors.

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

package model

import (
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/klog"
)

// ResourceName represents the name of the resource monitored by recommender.
type ResourceName string

// ResourceAmount represents quantity of a certain resource within a container.
// Note this keeps CPU in millicores (which is not a standard unit in APIs)
// and memory in bytes.
// Allowed values are in the range from 0 to MaxResourceAmount.
type ResourceAmount int64

// Resources is a map from resource name to the corresponding ResourceAmount.
type Resources map[ResourceName]ResourceAmount

const (
	// ResourceCPU represents CPU in millicores (1core = 1000millicores).
	ResourceCPU ResourceName = "cpu"
	// ResourceMemory represents memory, in bytes. (500Gi = 500GiB = 500 * 1024 * 1024 * 1024).
	ResourceMemory ResourceName = "memory"
	// MaxResourceAmount is the maximum allowed value of resource amount.
	MaxResourceAmount = ResourceAmount(1e14)
)

// CPUAmountFromCores converts CPU cores to a ResourceAmount.
func CPUAmountFromCores(cores float64) ResourceAmount {
	return resourceAmountFromFloat(cores * 1000.0)
}

// CoresFromCPUAmount converts ResourceAmount to number of cores expressed as float64.
func CoresFromCPUAmount(cpuAmount ResourceAmount) float64 {
	return float64(cpuAmount) / 1000.0
}

// QuantityFromCPUAmount converts CPU ResourceAmount to a resource.Quantity.
func QuantityFromCPUAmount(cpuAmount ResourceAmount) resource.Quantity {
	return *resource.NewScaledQuantity(int64(cpuAmount), -3)
}

// MemoryAmountFromBytes converts memory bytes to a ResourceAmount.
func MemoryAmountFromBytes(bytes float64) ResourceAmount {
	return resourceAmountFromFloat(bytes)
}

// BytesFromMemoryAmount converts ResourceAmount to number of bytes expressed as float64.
func BytesFromMemoryAmount(memoryAmount ResourceAmount) float64 {
	return float64(memoryAmount)
}

// QuantityFromMemoryAmount converts memory ResourceAmount to a resource.Quantity.
func QuantityFromMemoryAmount(memoryAmount ResourceAmount) resource.Quantity {
	return *resource.NewScaledQuantity(int64(memoryAmount), 0)
}

// ScaleResource returns the resource amount multiplied by a given factor.
func ScaleResource(amount ResourceAmount, factor float64) ResourceAmount {
	return resourceAmountFromFloat(float64(amount) * factor)
}

// ResourcesAsResourceList converts internal Resources representation to ResourcesList.
func ResourcesAsResourceList(resources Resources) apiv1.ResourceList {
	result := make(apiv1.ResourceList)
	for key, resourceAmount := range resources {
		var newKey apiv1.ResourceName
		var quantity resource.Quantity
		switch key {
		case ResourceCPU:
			newKey = apiv1.ResourceCPU
			quantity = QuantityFromCPUAmount(resourceAmount)
		case ResourceMemory:
			newKey = apiv1.ResourceMemory
			quantity = QuantityFromMemoryAmount(resourceAmount)
		default:
			klog.Errorf("Cannot translate %v resource name", key)
			continue
		}
		result[newKey] = quantity
	}
	return result
}

// ResourceNamesApiToModel converts an array of resource names expressed in API types into model types.
func ResourceNamesApiToModel(resources []apiv1.ResourceName) *[]ResourceName {
	result := make([]ResourceName, 0, len(resources))
	for _, resource := range resources {
		switch resource {
		case apiv1.ResourceCPU:
			result = append(result, ResourceCPU)
		case apiv1.ResourceMemory:
			result = append(result, ResourceMemory)
		default:
			klog.Errorf("Cannot translate %v resource name", resource)
			continue
		}
	}
	return &result
}

// RoundResourceAmount returns the given resource amount rounded down to the
// whole multiple of another resource amount (unit).
func RoundResourceAmount(amount, unit ResourceAmount) ResourceAmount {
	return ResourceAmount(int64(amount) - int64(amount)%int64(unit))
}

// ResourceAmountMax returns the larger of two resource amounts.
func ResourceAmountMax(amount1, amount2 ResourceAmount) ResourceAmount {
	if amount1 > amount2 {
		return amount1
	}
	return amount2
}

func resourceAmountFromFloat(amount float64) ResourceAmount {
	if amount < 0 {
		return ResourceAmount(0)
	} else if amount > float64(MaxResourceAmount) {
		return MaxResourceAmount
	} else {
		return ResourceAmount(amount)
	}
}

// PodID contains information needed to identify a Pod within a cluster.
type PodID struct {
	// Namespaces where the Pod is defined.
	Namespace string
	// PodName is the name of the pod unique within a namespace.
	PodName string
}

// ContainerID contains information needed to identify a Container within a cluster.
type ContainerID struct {
	PodID
	// ContainerName is the name of the container, unique within a pod.
	ContainerName string
}

// VpaID contains information needed to identify a VPA API object within a cluster.
type VpaID struct {
	Namespace string
	VpaName   string
}
