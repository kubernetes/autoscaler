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
// Limits which should be left unchanged are absent from the returned resource list.
func GetProportionalLimit(originalLimits, originalRequests, recommendation, defaultLimits corev1.ResourceList) (corev1.ResourceList, []string) {
	annotations := []string{}
	cpuLimit, annotation := getProportionalResourceLimit(corev1.ResourceCPU, originalLimits, originalRequests, recommendation, defaultLimits)
	if annotation != "" {
		annotations = append(annotations, annotation)
	}
	memLimit, annotation := getProportionalResourceLimit(corev1.ResourceMemory, originalLimits, originalRequests, recommendation, defaultLimits)
	if annotation != "" {
		annotations = append(annotations, annotation)
	}
	if memLimit == nil && cpuLimit == nil {
		return nil, annotations
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

// getProportionalResourceLimit returns the limit which should be applied to a container for the
// given resource, keeping it in the same proportion to the recommended request as the original
// limit was to the original request. A nil limit means the limit should be left unchanged.
func getProportionalResourceLimit(resourceName corev1.ResourceName, originalLimits, originalRequests, recommendation, defaultLimits corev1.ResourceList) (*resource.Quantity, string) {
	originalLimit, hasOriginalLimit := getOriginalLimit(resourceName, originalLimits, defaultLimits)
	// originalLimit not set, don't set limit.
	if !hasOriginalLimit {
		return nil, ""
	}
	recommendedRequest, hasRecommendation := recommendation[resourceName]
	// recommendedRequest not set, don't set limit.
	if !hasRecommendation || recommendedRequest.IsZero() {
		return nil, fmt.Sprintf("%v: limit NOT set since recommendedRequest is nil or 0", resourceName)
	}
	originalRequest, hasOriginalRequest := originalRequests[resourceName]
	// originalLimit set but originalRequest not set - K8s will treat the pod as if they were equal,
	// recommend limit equal to request
	if !hasOriginalRequest {
		result := recommendedRequest.DeepCopy()
		return &result, ""
	}
	// originalRequest is explicitly set to 0, which K8s does not default to the limit, so there is
	// no limit to request ratio to keep. Never lower the limit in this case, only raise it to the
	// recommended request if the recommendation doesn't fit within the original limit.
	if originalRequest.IsZero() {
		if recommendedRequest.Cmp(originalLimit) <= 0 {
			return nil, fmt.Sprintf("%v: limit left unchanged since originalRequest is 0", resourceName)
		}
		result := recommendedRequest.DeepCopy()
		return &result, fmt.Sprintf("%v: limit raised to recommendedRequest since originalRequest is 0", resourceName)
	}
	// originalLimit and originalRequest are set. If they are equal recommend limit equal to request.
	if originalRequest.MilliValue() == originalLimit.MilliValue() {
		result := recommendedRequest.DeepCopy()
		return &result, ""
	}
	if resourceName == corev1.ResourceCPU {
		result, isCapped := scaleQuantityProportionallyCPU( /* scaledQuantity= */ &originalLimit /* scaleBase= */, &originalRequest /* scaleResult= */, &recommendedRequest, noRounding)
		if isCapped == capped {
			return result, fmt.Sprintf(
				"%v: failed to keep limit to request ratio; capping limit to int64", resourceName)
		}
		return result, ""
	}
	result, capped := scaleQuantityProportionallyMem( /* scaledQuantity= */ &originalLimit /* scaleBase= */, &originalRequest /* scaleResult= */, &recommendedRequest, noRounding)
	if !capped {
		return result, ""
	}
	return result, fmt.Sprintf(
		"%v: failed to keep limit to request ratio; capping limit to int64", resourceName)
}

// getOriginalLimit returns the limit a container effectively has for the given resource and whether
// such a limit exists at all.
func getOriginalLimit(resourceName corev1.ResourceName, originalLimits, defaultLimits corev1.ResourceList) (resource.Quantity, bool) {
	originalLimit, hasOriginalLimit := originalLimits[resourceName]
	// A LimitRange default limit is only applied to containers which don't specify a limit for the
	// resource, so only fall back to it if the limit is unset.
	if !hasOriginalLimit {
		originalLimit, hasOriginalLimit = defaultLimits[resourceName]
	}
	// A limit of 0 is treated as no limit by the kubelet.
	if hasOriginalLimit && originalLimit.IsZero() {
		return originalLimit, false
	}
	return originalLimit, hasOriginalLimit
}

// GetBoundaryRequest returns the boundary (min/max) request that can be specified with
// preserving the original limit to request ratio. Returns nil if no boundary exists
func GetBoundaryRequest(resourceName corev1.ResourceName, originalRequests, originalLimits, boundaryLimits, defaultLimits corev1.ResourceList) *resource.Quantity {
	originalLimit, hasOriginalLimit := getOriginalLimit(resourceName, originalLimits, defaultLimits)
	// originalLimit not set, no boundary
	if !hasOriginalLimit {
		return &resource.Quantity{}
	}
	boundaryLimit := boundaryLimits[resourceName]
	originalRequest, hasOriginalRequest := originalRequests[resourceName]
	// originalLimit set but originalRequest not set - K8s will treat the pod as if they were equal.
	// An originalRequest explicitly set to 0 gives no limit to request ratio to keep either: the
	// limit is only ever raised to the recommended request, so the boundary is the boundary limit.
	if !hasOriginalRequest || originalRequest.IsZero() {
		return &boundaryLimit
	}

	// Determine which scaling function to use based on resource type.
	var result *resource.Quantity
	if resourceName == corev1.ResourceCPU {
		result, _ = scaleQuantityProportionallyCPU(&originalRequest /* scaledQuantity */, &originalLimit /* scaleBase */, &boundaryLimit /* scaleResult */, noRounding)
		return result
	}
	result, _ = scaleQuantityProportionallyMem(&originalRequest /* scaledQuantity */, &originalLimit /* scaleBase */, &boundaryLimit /* scaleResult */, noRounding)
	return result
}

type roundingMode int

const (
	noRounding roundingMode = iota
	roundUpToFullUnit
	roundDownToFullUnit
)

type scalingResultType int

const (
	success scalingResultType = iota
	capped
	divisionByZero
)

// scaleQuantityProportionallyCPU returns two values:
//  1. The first return value is in milliunits and has the same proportion to
//     scaledQuantity as scaleResult has to scaleBase.
//     It is calculated as (scaledQuantity * scaleResult) / scaleBase
//  2. The second return value describes the type of the first return value.
func scaleQuantityProportionallyCPU(scaledQuantity, scaleBase, scaleResult *resource.Quantity, rounding roundingMode) (*resource.Quantity, scalingResultType) {
	if scaleBase.IsZero() {
		return scaledQuantity, divisionByZero
	}

	originalMilli := big.NewInt(scaledQuantity.MilliValue())
	scaleBaseMilli := big.NewInt(scaleBase.MilliValue())
	scaleResultMilli := big.NewInt(scaleResult.MilliValue())

	var result big.Int
	result.Mul(originalMilli, scaleResultMilli)
	// If the division produces a remainder:
	// - with roundUpToFullUnit, we apply ceiling to the value
	// - with noRounding or roundDownToFullUnit, we apply floor to the value.
	//   Note: In other words, the remainder is discarded. This is acceptable because the difference is at most 1 millicore.
	// TODO(iamzili) - I think we eventually want to get rid of the noRounding mode.
	quotient := new(big.Int)
	remainder := new(big.Int)
	quotient.DivMod(&result, scaleBaseMilli, remainder)
	if quotient.IsInt64() {
		if remainder.Sign() != 0 && rounding == roundUpToFullUnit {
			quotient.Add(quotient, big.NewInt(1))
		}
		return resource.NewMilliQuantity(quotient.Int64(), scaledQuantity.Format), success
	}
	return resource.NewMilliQuantity(math.MaxInt64, scaledQuantity.Format), capped
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
