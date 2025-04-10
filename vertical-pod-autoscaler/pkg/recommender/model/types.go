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
	"fmt"
	"math"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/klog/v2"
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
func ResourcesAsResourceList(resources Resources, humanizeMemory bool, roundCPUMillicores int) apiv1.ResourceList {
	result := make(apiv1.ResourceList)
	for key, resourceAmount := range resources {
		var newKey apiv1.ResourceName
		var quantity resource.Quantity
		switch key {
		case ResourceCPU:
			newKey = apiv1.ResourceCPU
			quantity = QuantityFromCPUAmount(resourceAmount)
			if roundCPUMillicores != 1 && !quantity.IsZero() {
				roundedValues, err := RoundUpToScale(resourceAmount, roundCPUMillicores)
				if err != nil {
					klog.V(4).InfoS("Error rounding CPU value; leaving unchanged", "rawValue", resourceAmount, "scale", roundCPUMillicores, "error", err)
				} else {
					klog.V(4).InfoS("Successfully rounded CPU value", "rawValue", resourceAmount, "roundedValue", roundedValues)
				}
				quantity = QuantityFromCPUAmount(roundedValues)
			}
		case ResourceMemory:
			newKey = apiv1.ResourceMemory
			if humanizeMemory && resourceAmount > 0 {
				humanizedValue := HumanizeMemoryQuantity(resourceAmount)
				klog.V(4).InfoS("Converting raw value to humanized value", "resourceAmount", resourceAmount, "humanizedValue", humanizedValue)
				quantity = resource.MustParse(humanizedValue)
				klog.V(4).InfoS("after quantity creation", "quantity", quantity.String())
			} else {
				quantity = QuantityFromMemoryAmount(resourceAmount)
			}
		default:
			klog.ErrorS(nil, "Cannot translate resource name", "resourceName", key)
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
			klog.ErrorS(nil, "Cannot translate resource name", "resourceName", resource)
			continue
		}
	}
	return &result
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

// HumanizeMemoryQuantity converts bytes to a human-readable string using binary units (KiB, MiB, GiB, TiB), rounding to the nearest whole number.
func HumanizeMemoryQuantity(bytes ResourceAmount) string {
	const (
		Ki ResourceAmount = 1024
		Mi                = 1024 * Ki
		Gi                = 1024 * Mi
		Ti                = 1024 * Gi
	)

	switch {
	case bytes >= Ti:
		return fmt.Sprintf("%.0fTi", math.Ceil(float64(bytes)/float64(Ti)))
	case bytes >= Gi:
		return fmt.Sprintf("%.0fGi", math.Ceil(float64(bytes)/float64(Gi)))
	case bytes >= Mi:
		return fmt.Sprintf("%.0fMi", math.Ceil(float64(bytes)/float64(Mi)))
	case bytes >= Ki:
		return fmt.Sprintf("%.0fKi", math.Ceil(float64(bytes)/float64(Ki)))
	default:
		return fmt.Sprintf("%d", bytes)
	}
}

// RoundUpToScale rounds the value to the nearest multiple of scale, rounding up
func RoundUpToScale(value ResourceAmount, scale int) (ResourceAmount, error) {
	if scale <= 0 {
		return value, fmt.Errorf("scale must be greater than zero")
	}
	scale64 := int64(scale)
	roundedValue := int64(math.Ceil(float64(value)/float64(scale64))) * scale64
	return ResourceAmount(roundedValue), nil
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
