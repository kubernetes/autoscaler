/*
Copyright 2025 The Kubernetes Authors.

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

package filter

import (
	v1 "k8s.io/autoscaler/cluster-autoscaler/apis/capacitybuffer/autoscaling.x-k8s.io/v1alpha1"
)

// statusIncludeFilter includes buffers with the defined conditions (opposite of statusFilter)
type statusIncludeFilter struct {
	conditions map[string]string
}

// NewStatusIncludeFilter creates an instance of statusIncludeFilter that includes only the buffers with conditions in passed conditions.
func NewStatusIncludeFilter(conditionsToInclude map[string]string) *statusIncludeFilter {
	return &statusIncludeFilter{
		conditions: conditionsToInclude,
	}
}

// Filter filters the passed buffers to include only those with the specified status conditions
func (f *statusIncludeFilter) Filter(buffersToFilter []*v1.CapacityBuffer) ([]*v1.CapacityBuffer, []*v1.CapacityBuffer) {
	var included []*v1.CapacityBuffer
	var excluded []*v1.CapacityBuffer

	for _, buffer := range buffersToFilter {
		if f.hasAllConditions(buffer) {
			included = append(included, buffer)
		} else {
			excluded = append(excluded, buffer)
		}
	}
	return included, excluded
}

// hasAllConditions checks if the buffer has all the specified conditions
func (f *statusIncludeFilter) hasAllConditions(buffer *v1.CapacityBuffer) bool {
	bufferConditions := buffer.Status.Conditions
	matchCount := 0

	for condType, condStatus := range f.conditions {
		for _, condition := range bufferConditions {
			if condition.Type == condType && string(condition.Status) == condStatus {
				matchCount++
				break
			}
		}
	}

	// All specified conditions must be present
	return matchCount == len(f.conditions)
}

// CleanUp cleans up the filter's internal structures.
func (f *statusIncludeFilter) CleanUp() {
}
