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
	v1 "k8s.io/autoscaler/cluster-autoscaler/apis/capacitybuffer/autoscaling.x-k8s.io/v1"
)

// Filter filters CapacityBuffer based on some criteria.
type Filter interface {
	Filter(buffers []*v1.CapacityBuffer) ([]*v1.CapacityBuffer, []*v1.CapacityBuffer)
	CleanUp()
}

// combinedFilter is a list of Filter
type combinedFilter struct {
	filters []Filter
}

// NewCombinedFilter construct combinedFilter.
func NewCombinedFilter(filters []Filter) *combinedFilter {
	return &combinedFilter{filters}
}

// AddFilter append a filter to the list.
func (f *combinedFilter) AddFilter(filter Filter) {
	f.filters = append(f.filters, filter)
}

// Filter runs sub-filters sequentially
func (f *combinedFilter) Filter(buffers []*v1.CapacityBuffer) ([]*v1.CapacityBuffer, []*v1.CapacityBuffer) {
	var totalFilteredOutBuffers []*v1.CapacityBuffer
	for _, buffersFilter := range f.filters {
		updatedBuffersList, filteredOutBuffers := buffersFilter.Filter(buffers)
		buffers = updatedBuffersList
		totalFilteredOutBuffers = append(totalFilteredOutBuffers, filteredOutBuffers...)
	}
	return buffers, totalFilteredOutBuffers
}

// CleanUp cleans up the filter's internal structures.
func (f *combinedFilter) CleanUp() {
	for _, filter := range f.filters {
		filter.CleanUp()
	}
}
