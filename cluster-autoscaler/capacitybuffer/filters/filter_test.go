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
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/autoscaler/cluster-autoscaler/apis/capacitybuffer/autoscaling.x-k8s.io/v1alpha1"
)

func TestCombinedAnyFilter(t *testing.T) {
	tests := []struct {
		name                       string
		filters                    []Filter
		buffers                    []*v1.CapacityBuffer
		expectedFilteredBuffers    []*v1.CapacityBuffer
		expectedFilteredOutBuffers []*v1.CapacityBuffer
	}{
		{
			name:    "buffer to be filtered",
			filters: []Filter{getTestBufferGenerationFilter(1)},
			buffers: []*v1.CapacityBuffer{
				getTestBufferWithGeneration("someBuffer", 1),
			},
			expectedFilteredBuffers: []*v1.CapacityBuffer{
				getTestBufferWithGeneration("someBuffer", 1),
			},
			expectedFilteredOutBuffers: []*v1.CapacityBuffer{},
		},
		{
			name:    "buffer to be filtered out",
			filters: []Filter{getTestBufferGenerationFilter(1)},

			buffers: []*v1.CapacityBuffer{
				getTestBufferWithGeneration("someBuffer", 2),
			},
			expectedFilteredBuffers: []*v1.CapacityBuffer{},
			expectedFilteredOutBuffers: []*v1.CapacityBuffer{
				getTestBufferWithGeneration("someBuffer", 2),
			},
		},
		{
			name:    "1 buffer to be filtered and 1 filtered out",
			filters: []Filter{getTestBufferGenerationFilter(1)},

			buffers: []*v1.CapacityBuffer{
				getTestBufferWithGeneration("Buffer1", 1),
				getTestBufferWithGeneration("Buffer2", 2),
			},
			expectedFilteredBuffers: []*v1.CapacityBuffer{
				getTestBufferWithGeneration("Buffer1", 1),
			},
			expectedFilteredOutBuffers: []*v1.CapacityBuffer{
				getTestBufferWithGeneration("Buffer2", 2),
			},
		},
		{
			name: "multiple filters multiple buffers",
			filters: []Filter{
				getTestBufferGenerationFilter(1),
				getTestBufferGenerationFilter(2),
				getTestBufferGenerationFilter(3),
			},

			buffers: []*v1.CapacityBuffer{
				getTestBufferWithGeneration("Buffer1", 1),
				getTestBufferWithGeneration("Buffer2", 2),
				getTestBufferWithGeneration("Buffer3_A", 3),
				getTestBufferWithGeneration("Buffer3_B", 3),
				getTestBufferWithGeneration("Buffer4_A", 4),
				getTestBufferWithGeneration("Buffer4_B", 4),
				getTestBufferWithGeneration("Buffer5", 5),
			},
			expectedFilteredBuffers: []*v1.CapacityBuffer{
				getTestBufferWithGeneration("Buffer1", 1),
				getTestBufferWithGeneration("Buffer2", 2),
				getTestBufferWithGeneration("Buffer3_A", 3),
				getTestBufferWithGeneration("Buffer3_B", 3),
			},
			expectedFilteredOutBuffers: []*v1.CapacityBuffer{
				getTestBufferWithGeneration("Buffer4_A", 4),
				getTestBufferWithGeneration("Buffer4_B", 4),
				getTestBufferWithGeneration("Buffer5", 5),
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			filter := NewCombinedAnyFilter(test.filters)
			filtered, filteredOut := filter.Filter(test.buffers)
			assert.ElementsMatch(t, test.expectedFilteredBuffers, filtered)
			assert.ElementsMatch(t, test.expectedFilteredOutBuffers, filteredOut)
		})
	}
}

func getTestBufferGenerationFilter(generation int64) *testGenerationFilter {
	return &testGenerationFilter{filerGeneration: generation}
}

type testGenerationFilter struct {
	filerGeneration int64
}

func (f *testGenerationFilter) Filter(buffers []*v1.CapacityBuffer) ([]*v1.CapacityBuffer, []*v1.CapacityBuffer) {
	filteredBuffers := []*v1.CapacityBuffer{}
	filteredOutBuffers := []*v1.CapacityBuffer{}
	for _, buffer := range buffers {
		if buffer.Generation == f.filerGeneration {
			filteredBuffers = append(filteredBuffers, buffer)
		} else {
			filteredOutBuffers = append(filteredOutBuffers, buffer)
		}
	}
	return filteredBuffers, filteredOutBuffers
}

func (f *testGenerationFilter) CleanUp() {}
