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

package cloudprovider

import (
	"bytes"
	"fmt"
	"k8s.io/apimachinery/pkg/util/sets"
	"math"
)

// ResourceLimiter contains limits (max, min) for resources (cores, memory etc.).
type ResourceLimiter struct {
	minLimits map[string]int64
	maxLimits map[string]int64
}

// NewResourceLimiter creates new ResourceLimiter for map. Maps are deep copied.
func NewResourceLimiter(minLimits map[string]int64, maxLimits map[string]int64) *ResourceLimiter {
	minLimitsCopy := make(map[string]int64)
	maxLimitsCopy := make(map[string]int64)
	for key, value := range minLimits {
		if value > 0 {
			minLimitsCopy[key] = value
		}
	}
	for key, value := range maxLimits {
		maxLimitsCopy[key] = value
	}
	return &ResourceLimiter{minLimitsCopy, maxLimitsCopy}
}

// GetMin returns minimal number of resources for a given resource type.
func (r *ResourceLimiter) GetMin(resourceName string) int64 {
	result, found := r.minLimits[resourceName]
	if found {
		return result
	}
	return 0
}

// GetMax returns maximal number of resources for a given resource type.
func (r *ResourceLimiter) GetMax(resourceName string) int64 {
	result, found := r.maxLimits[resourceName]
	if found {
		return result
	}
	return math.MaxInt64
}

// GetResources returns list of all resource names for which min or max limits are defined
func (r *ResourceLimiter) GetResources() []string {
	minResources := sets.StringKeySet(r.minLimits)
	maxResources := sets.StringKeySet(r.maxLimits)
	return minResources.Union(maxResources).List()
}

// HasMinLimitSet returns true iff minimal limit is set for given resource.
func (r *ResourceLimiter) HasMinLimitSet(resourceName string) bool {
	_, found := r.minLimits[resourceName]
	return found
}

// HasMaxLimitSet returns true iff maximal limit is set for given resource.
func (r *ResourceLimiter) HasMaxLimitSet(resourceName string) bool {
	_, found := r.maxLimits[resourceName]
	return found
}

func (r *ResourceLimiter) String() string {
	var buffer bytes.Buffer
	for _, name := range r.GetResources() {
		if buffer.Len() > 0 {
			buffer.WriteString(", ")
		}
		buffer.WriteString(fmt.Sprintf("{%s : %d - %d}", name, r.GetMin(name), r.GetMax(name)))
	}
	return buffer.String()
}
