/*
Copyright 2019 The Kubernetes Authors.

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

package api

import (
	"fmt"
	"math"
	"math/big"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// ContainerResources holds resources request for container
type ContainerResources struct {
	Limits   corev1.ResourceList
	Requests corev1.ResourceList
}

// GetProportionalLimit returns limit that will be in the same proportion to recommended request as original limit had to original request.
func GetProportionalLimit(originalLimit, originalRequest, recommendation, defaultLimit corev1.ResourceList) (corev1.ResourceList, []string) {
	annotations := []string{}
	cpuLimit, annotation := getProportionalResourceLimit(corev1.ResourceCPU, originalLimit.Cpu(), originalRequest.Cpu(), recommendation.Cpu(), defaultLimit.Cpu())
	if annotation != "" {
		annotations = append(annotations, annotation)
	}
	memLimit, annotation := getProportionalResourceLimit(corev1.ResourceMemory, originalLimit.Memory(), originalRequest.Memory(), recommendation.Memory(), defaultLimit.Memory())
	if annotation != "" {
		annotations = append(annotations, annotation)
	}
	if memLimit == nil && cpuLimit == nil {
		return nil, []string{}
	}
	result := corev1.ResourceList{}
	if cpuLimit != nil {
		result[corev1.ResourceCPU] = *cpuLimit
	}
	if memLimit != nil {
		result[corev1.ResourceMemory] = *memLimit
	}
	return result, annotations
}

func getProportionalResourceLimit(resourceName corev1.ResourceName, originalLimit, originalRequest, recommendedRequest, defaultLimit *resource.Quantity) (*resource.Quantity, string) {
	if originalLimit == nil || originalLimit.Value() == 0 && defaultLimit != nil {
		originalLimit = defaultLimit
	}
	// originalLimit not set, don't set limit.
	if originalLimit == nil || originalLimit.Value() == 0 {
		return nil, fmt.Sprintf("%v: limit NOT set since originalLimit is nil or 0", resourceName)
	}
	// recommendedRequest not set, don't set limit.
	if recommendedRequest == nil || recommendedRequest.Value() == 0 {
		return nil, fmt.Sprintf("%v: limit NOT set since recommendedRequest is nil or 0", resourceName)
	}
	// originalLimit set but originalRequest not set - K8s will treat the pod as if they were equal,
	// recommend limit equal to request
	if originalRequest == nil || originalRequest.Value() == 0 {
		result := *recommendedRequest
		return &result, ""
	}
	// originalLimit and originalRequest are set. If they are equal recommend limit equal to request.
	if originalRequest.MilliValue() == originalLimit.MilliValue() {
		result := *recommendedRequest
		return &result, ""
	}
	if resourceName == corev1.ResourceCPU {
		result, capped := scaleQuantityProportionallyCPU( /* scaledQuantity= */ originalLimit /* scaleBase= */, originalRequest /* scaleResult= */, recommendedRequest, noRounding)
		if !capped {
			return result, ""
		}
		return result, fmt.Sprintf(
			"%v: failed to keep limit to request ratio; capping limit to int64", resourceName)
	}
	result, capped := scaleQuantityProportionallyMem( /* scaledQuantity= */ originalLimit /* scaleBase= */, originalRequest /* scaleResult= */, recommendedRequest, noRounding)
	if !capped {
		return result, ""
	}
	return result, fmt.Sprintf(
		"%v: failed to keep limit to request ratio; capping limit to int64", resourceName)
}

// GetBoundaryRequest returns the boundary (min/max) request that can be specified with
// preserving the original limit to request ratio. Returns nil if no boundary exists
func GetBoundaryRequest(resourceName corev1.ResourceName, originalRequest, originalLimit, boundaryLimit, defaultLimit *resource.Quantity) *resource.Quantity {
	if originalLimit == nil || originalLimit.Value() == 0 && defaultLimit != nil {
		originalLimit = defaultLimit
	}
	// originalLimit not set, no boundary
	if originalLimit == nil || originalLimit.Value() == 0 {
		return &resource.Quantity{}
	}
	// originalLimit set but originalRequest not set - K8s will treat the pod as if they were equal
	if originalRequest == nil || originalRequest.Value() == 0 {
		return boundaryLimit
	}

	// Determine which scaling function to use based on resource type.
	var result *resource.Quantity
	if resourceName == corev1.ResourceCPU {
		result, _ = scaleQuantityProportionallyCPU(originalRequest /* scaledQuantity */, originalLimit /* scaleBase */, boundaryLimit /* scaleResult */, noRounding)
		return result
	}
	result, _ = scaleQuantityProportionallyMem(originalRequest /* scaledQuantity */, originalLimit /* scaleBase */, boundaryLimit /* scaleResult */, noRounding)
	return result
}

type roundingMode int

const (
	noRounding roundingMode = iota
	roundUpToFullUnit
	roundDownToFullUnit
)

// scaleQuantityProportionallyCPU returns a value in milliunits which has the same proportion to scaledQuantity as scaleResult has to scaleBase.
// It also returns a bool indicating if it had to cap result to MaxInt64 milliunits.
func scaleQuantityProportionallyCPU(scaledQuantity, scaleBase, scaleResult *resource.Quantity, rounding roundingMode) (*resource.Quantity, bool) {
	originalMilli := big.NewInt(scaledQuantity.MilliValue())
	scaleBaseMilli := big.NewInt(scaleBase.MilliValue())
	scaleResultMilli := big.NewInt(scaleResult.MilliValue())
	var scaledOriginal big.Int
	scaledOriginal.Mul(originalMilli, scaleResultMilli)
	scaledOriginal.Div(&scaledOriginal, scaleBaseMilli)
	if scaledOriginal.IsInt64() {
		result := resource.NewMilliQuantity(scaledOriginal.Int64(), scaledQuantity.Format)
		if rounding == roundUpToFullUnit {
			result.RoundUp(resource.Scale(0))
		}
		if rounding == roundDownToFullUnit {
			result.Sub(*resource.NewMilliQuantity(999, result.Format))
			result.RoundUp(resource.Scale(0))
		}
		return result, false
	}
	return resource.NewMilliQuantity(math.MaxInt64, scaledQuantity.Format), true
}

// scaleQuantityProportionallyMem returns a value in whole units which has the same proportion to scaledQuantity as scaleResult has to scaleBase.
// It also returns a bool indicating if it had to cap result to MaxInt64 units.
func scaleQuantityProportionallyMem(scaledQuantity, scaleBase, scaleResult *resource.Quantity, rounding roundingMode) (*resource.Quantity, bool) {
	originalValue := big.NewInt(scaledQuantity.Value())
	scaleBaseValue := big.NewInt(scaleBase.Value())
	scaleResultValue := big.NewInt(scaleResult.Value())
	var scaledOriginal big.Int
	scaledOriginal.Mul(originalValue, scaleResultValue)
	scaledOriginal.Div(&scaledOriginal, scaleBaseValue)
	if scaledOriginal.IsInt64() {
		result := resource.NewQuantity(scaledOriginal.Int64(), scaledQuantity.Format)
		if rounding == roundUpToFullUnit {
			result.RoundUp(resource.Scale(0))
		}
		if rounding == roundDownToFullUnit {
			result.Sub(*resource.NewMilliQuantity(999, result.Format))
			result.RoundUp(resource.Scale(0))
		}
		return result, false
	}
	return resource.NewQuantity(math.MaxInt64, scaledQuantity.Format), true
}

// RemoveEmptyResourceKeyIfAny ensure that we are not pushing a resource with an empty key. Return true if an empty key was eliminated
func (cr *ContainerResources) RemoveEmptyResourceKeyIfAny() bool {
	var found bool
	if _, foundEmptyKey := cr.Limits[""]; foundEmptyKey {
		delete(cr.Limits, "")
		found = true
	}
	if _, foundEmptyKey := cr.Requests[""]; foundEmptyKey {
		delete(cr.Requests, "")
		found = true
	}
	return found
}
