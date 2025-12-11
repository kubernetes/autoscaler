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

// statusFilter filters out buffers with the defined conditions
type statusFilter struct {
	conditions map[string]string
}

// NewStatusFilter creates an instance of statusFilter that filters out the buffers with condition in passed conditions.
func NewStatusFilter(conditionsToFilterOut map[string]string) *statusFilter {
	return &statusFilter{
		conditions: conditionsToFilterOut,
	}
}

// Filter filters the passed buffers based on buffer status conditions
func (f *statusFilter) Filter(buffersToFilter []*v1.CapacityBuffer) ([]*v1.CapacityBuffer, []*v1.CapacityBuffer) {
	var buffers []*v1.CapacityBuffer
	var filteredOutBuffers []*v1.CapacityBuffer

	for _, buffer := range buffersToFilter {
		if !f.hasCondition(buffer) {
			buffers = append(buffers, buffer)
		} else {
			filteredOutBuffers = append(filteredOutBuffers, buffer)
		}
	}
	return buffers, filteredOutBuffers
}

func (f *statusFilter) hasCondition(buffer *v1.CapacityBuffer) bool {
	bufferConditions := buffer.Status.Conditions
	for _, condition := range bufferConditions {
		if val, found := f.conditions[condition.Type]; found && val == string(condition.Status) {
			return true
		}
	}
	return false
}

// CleanUp cleans up the filter's internal structures.
func (f *statusFilter) CleanUp() {
}
