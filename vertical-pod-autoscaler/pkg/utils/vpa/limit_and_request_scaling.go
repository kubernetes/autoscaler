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

	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// ContainerResources holds resources request for container
type ContainerResources struct {
	Limits   core.ResourceList
	Requests core.ResourceList
}

func newContainerResources() ContainerResources {
	return ContainerResources{
		Requests: core.ResourceList{},
		Limits:   core.ResourceList{},
	}
}

// GetProportionalLimit returns limit that will be in the same proportion to recommended request as original limit had to original request.
func GetProportionalLimit(originalLimit, originalRequest, recommendation, defaultLimit core.ResourceList) (core.ResourceList, []string) {
	annotations := []string{}
	cpuLimit, annotation := getProportionalResourceLimit(core.ResourceCPU, originalLimit.Cpu(), originalRequest.Cpu(), recommendation.Cpu(), defaultLimit.Cpu())
	if annotation != "" {
		annotations = append(annotations, annotation)
	}
	memLimit, annotation := getProportionalResourceLimit(core.ResourceMemory, originalLimit.Memory(), originalRequest.Memory(), recommendation.Memory(), defaultLimit.Memory())
	if annotation != "" {
		annotations = append(annotations, annotation)
	}
	if memLimit == nil && cpuLimit == nil {
		return nil, []string{}
	}
	result := core.ResourceList{}
	if cpuLimit != nil {
		result[core.ResourceCPU] = *cpuLimit
	}
	if memLimit != nil {
		result[core.ResourceMemory] = *memLimit
	}
	return result, annotations
}

func getProportionalResourceLimit(resourceName core.ResourceName, originalLimit, originalRequest, recommendedRequest, defaultLimit *resource.Quantity) (*resource.Quantity, string) {
	if originalLimit == nil || originalLimit.Value() == 0 && defaultLimit != nil {
		originalLimit = defaultLimit
	}
	// originalLimit not set, don't set limit.
	if originalLimit == nil || originalLimit.Value() == 0 {
		return nil, ""
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
	result, capped := scaleQuantityProportionally( /*scaledQuantity=*/ originalLimit /*scaleBase=*/, originalRequest /*scaleResult=*/, recommendedRequest, noRounding)
	if !capped {
		return result, ""
	}
	return result, fmt.Sprintf(
		"%v: failed to keep limit to request ratio; capping limit to int64", resourceName)
}

// GetBoundaryRequest returns the boundary (min/max) request that can be specified with
// preserving the original limit to request ratio. Returns nil if no boundary exists
func GetBoundaryRequest(originalRequest, originalLimit, boundaryLimit, defaultLimit *resource.Quantity) *resource.Quantity {
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
	result, _ := scaleQuantityProportionally(originalRequest /* scaledQuantity */, originalLimit /*scaleBase*/, boundaryLimit /*scaleResult*/, noRounding)
	return result
}

type roundingMode int

const (
	noRounding roundingMode = iota
	roundUpToFullUnit
	roundDownToFullUnit
)

// scaleQuantityProportionally returns value which has the same proportion to scaledQuantity as scaleResult has to scaleBase
// It also returns a bool indicating if it had to cap result to MaxInt64 milliunits.
func scaleQuantityProportionally(scaledQuantity, scaleBase, scaleResult *resource.Quantity, rounding roundingMode) (*resource.Quantity, bool) {
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
