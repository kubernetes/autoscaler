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

// bufferGenerationChengedFilter filters in buffers that its generation changeed
type bufferGenerationChengedFilter struct {
	buffersGenerations map[string]int64
}

// NewBufferGenerationChangedFilter creates an instance of bufferGenerationChengedFilter that filters the buffers that need to be updated.
func NewBufferGenerationChangedFilter() *bufferGenerationChengedFilter {
	return &bufferGenerationChengedFilter{
		buffersGenerations: map[string]int64{},
	}
}

// Filter returns the buffers that has a generation change as the filtered buffers
func (f *bufferGenerationChengedFilter) Filter(buffersToFilter []*v1.CapacityBuffer) ([]*v1.CapacityBuffer, []*v1.CapacityBuffer) {
	var buffers []*v1.CapacityBuffer
	var filteredOutBuffers []*v1.CapacityBuffer

	for _, buffer := range buffersToFilter {
		if f.generationChanged(buffer) {
			buffers = append(buffers, buffer)
		} else {
			filteredOutBuffers = append(filteredOutBuffers, buffer)
		}
	}
	return buffers, filteredOutBuffers
}

func (f *bufferGenerationChengedFilter) generationChanged(buffer *v1.CapacityBuffer) bool {
	newGeneration := buffer.Generation
	oldGeneration, found := f.buffersGenerations[buffer.Name]
	f.buffersGenerations[buffer.Name] = newGeneration
	if found && oldGeneration == newGeneration {
		return false
	}
	return true
}

// CleanUp cleans up the filter's internal structures.
func (f *bufferGenerationChengedFilter) CleanUp() {
}
