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
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"math"
	"math/big"
)

// ContainerResources holds resources request for container
type ContainerResources struct {
	Limits   v1.ResourceList
	Requests v1.ResourceList
}

func newContainerResources() ContainerResources {
	return ContainerResources{
		Requests: v1.ResourceList{},
		Limits:   v1.ResourceList{},
	}
}

// GetProportionalLimit returns limit that will be in the same proportion to recommended request as original limit had to original request.
func GetProportionalLimit(originalLimit, originalRequest, recommendedRequest, defaultLimit *resource.Quantity) (*resource.Quantity, string) {
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
	result, capped := scaleQuantityProportionally( /*scaledQuantity=*/ originalLimit /*scaleBase=*/, originalRequest /*scaleResult=*/, recommendedRequest)
	if capped {
		return result, ""
	}
	return result, fmt.Sprintf(
		"failed to keep limit to request proportion of %s to %s with recommended request of %s; doesn't fit in int64. Capping limit to MaxInt64 milliunits",
		originalLimit, originalRequest, recommendedRequest)
}

// scaleQuantityProportionally returns value which has the same proportion to scaledQuantity as scaleResult has to scaleBase
// It also returns a bool indicating if it had to cap result to MaxInt64 milliunits.
func scaleQuantityProportionally(scaledQuantity, scaleBase, scaleResult *resource.Quantity) (*resource.Quantity, bool) {
	originalMilli := big.NewInt(scaledQuantity.MilliValue())
	scaleBaseMilli := big.NewInt(scaleBase.MilliValue())
	scaleResultMilli := big.NewInt(scaleResult.MilliValue())
	var scaledOriginal big.Int
	scaledOriginal.Mul(originalMilli, scaleResultMilli)
	scaledOriginal.Div(&scaledOriginal, scaleBaseMilli)
	if scaledOriginal.IsInt64() {
		return resource.NewMilliQuantity(scaledOriginal.Int64(), scaledQuantity.Format), false
	}
	return resource.NewMilliQuantity(math.MaxInt64, scaledQuantity.Format), true
}

func proportionallyCapLimitToMax(recommendedRequest, recommendedLimit, maxLimit *resource.Quantity) (request, limit *resource.Quantity) {
	if recommendedLimit == nil || maxLimit == nil || maxLimit.IsZero() {
		return recommendedRequest, recommendedLimit
	}
	if recommendedLimit.Cmp(*maxLimit) <= 0 {
		return recommendedRequest, recommendedLimit
	}
	scaledRequest, _ := scaleQuantityProportionally(recommendedRequest, recommendedLimit, maxLimit)
	return scaledRequest, maxLimit
}

// ProportionallyCapResourcesToMaxLimit caps CPU and memory limit to maximu and scales requests to maintain limit/request ratio.
func ProportionallyCapResourcesToMaxLimit(recommendedRequests v1.ResourceList, cpuLimit, memLimit, maxCpuLimit, maxMemLimit *resource.Quantity) ContainerResources {
	scaledCpuRequest, scaledCpuLimit := proportionallyCapLimitToMax(recommendedRequests.Cpu(), cpuLimit, maxCpuLimit)
	scaledMemRequest, scaledMemLimit := proportionallyCapLimitToMax(recommendedRequests.Memory(), memLimit, maxMemLimit)
	result := newContainerResources()

	result.Requests[v1.ResourceCPU] = *scaledCpuRequest
	result.Requests[v1.ResourceMemory] = *scaledMemRequest
	if scaledCpuLimit != nil {
		result.Limits[v1.ResourceCPU] = *scaledCpuLimit
	}
	if scaledMemLimit != nil {
		result.Limits[v1.ResourceMemory] = *scaledMemLimit
	}
	return result
}
