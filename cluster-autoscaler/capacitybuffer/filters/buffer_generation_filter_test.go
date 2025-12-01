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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/autoscaler/cluster-autoscaler/apis/capacitybuffer/autoscaling.x-k8s.io/v1alpha1"
)

func TestBufferGenerationChengedFilter(t *testing.T) {
	tests := []struct {
		name                       string
		buffers                    []*v1.CapacityBuffer
		cachedBufferGenerations    map[string]int64
		expectedFilteredBuffers    []*v1.CapacityBuffer
		expectedFilteredOutBuffers []*v1.CapacityBuffer
	}{
		{
			name: "buffer not in generations cache",
			buffers: []*v1.CapacityBuffer{
				getTestBufferWithGeneration("someBuffer", 1),
			},
			cachedBufferGenerations: map[string]int64{},
			expectedFilteredBuffers: []*v1.CapacityBuffer{
				getTestBufferWithGeneration("someBuffer", 1),
			},
			expectedFilteredOutBuffers: []*v1.CapacityBuffer{},
		},
		{
			name: "buffer in generations cache same generation",
			buffers: []*v1.CapacityBuffer{
				getTestBufferWithGeneration("someBuffer", 1),
			},
			cachedBufferGenerations: map[string]int64{"someBuffer": 1},
			expectedFilteredBuffers: []*v1.CapacityBuffer{},
			expectedFilteredOutBuffers: []*v1.CapacityBuffer{
				getTestBufferWithGeneration("someBuffer", 1),
			},
		},
		{
			name: "buffer in generations cache different generation",
			buffers: []*v1.CapacityBuffer{
				getTestBufferWithGeneration("someBuffer", 2),
			},
			cachedBufferGenerations: map[string]int64{"someBuffer": 1},
			expectedFilteredBuffers: []*v1.CapacityBuffer{
				getTestBufferWithGeneration("someBuffer", 2),
			},
			expectedFilteredOutBuffers: []*v1.CapacityBuffer{},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bufferGenerationChengedFilter := &bufferGenerationChengedFilter{
				buffersGenerations: test.cachedBufferGenerations,
			}
			filtered, filteredOut := bufferGenerationChengedFilter.Filter(test.buffers)
			assert.ElementsMatch(t, test.expectedFilteredBuffers, filtered)
			assert.ElementsMatch(t, test.expectedFilteredOutBuffers, filteredOut)
		})
	}
}

func TestBufferGenerationChengedFilterCaching(t *testing.T) {
	tests := []struct {
		name                       string
		buffersIterations          [][]*v1.CapacityBuffer
		expectedFilteredBuffers    [][]*v1.CapacityBuffer
		expectedFilteredOutBuffers [][]*v1.CapacityBuffer
	}{
		{
			name: "2 iterations, one buffer getting cached",
			buffersIterations: [][]*v1.CapacityBuffer{
				{getTestBufferWithGeneration("someBuffer", 1)},
				{getTestBufferWithGeneration("someBuffer", 1)},
			},
			expectedFilteredBuffers: [][]*v1.CapacityBuffer{
				{getTestBufferWithGeneration("someBuffer", 1)},
				{},
			},
			expectedFilteredOutBuffers: [][]*v1.CapacityBuffer{
				{},
				{getTestBufferWithGeneration("someBuffer", 1)},
			},
		},
		{
			name: "3 iterations, one buffer generation changes",
			buffersIterations: [][]*v1.CapacityBuffer{
				{getTestBufferWithGeneration("someBuffer", 1)},
				{getTestBufferWithGeneration("someBuffer", 1)},
				{getTestBufferWithGeneration("someBuffer", 3)},
			},
			expectedFilteredBuffers: [][]*v1.CapacityBuffer{
				{getTestBufferWithGeneration("someBuffer", 1)},
				{},
				{getTestBufferWithGeneration("someBuffer", 3)},
			},
			expectedFilteredOutBuffers: [][]*v1.CapacityBuffer{
				{},
				{getTestBufferWithGeneration("someBuffer", 1)},
				{},
			},
		},
		{
			name: "3 iterations, multiple buffers generation changes",
			buffersIterations: [][]*v1.CapacityBuffer{
				{getTestBufferWithGeneration("someBuffer", 1), getTestBufferWithGeneration("anotherBuffer", 3)},
				{getTestBufferWithGeneration("someBuffer", 2), getTestBufferWithGeneration("anotherBuffer", 3)},
				{getTestBufferWithGeneration("someBuffer", 5)},
			},
			expectedFilteredBuffers: [][]*v1.CapacityBuffer{
				{getTestBufferWithGeneration("someBuffer", 1), getTestBufferWithGeneration("anotherBuffer", 3)},
				{getTestBufferWithGeneration("someBuffer", 2)},
				{getTestBufferWithGeneration("someBuffer", 5)},
			},
			expectedFilteredOutBuffers: [][]*v1.CapacityBuffer{
				{},
				{getTestBufferWithGeneration("anotherBuffer", 3)},
				{},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bufferGenerationChengedFilter := NewBufferGenerationChangedFilter()
			for index, buffers := range test.buffersIterations {
				filtered, filteredOut := bufferGenerationChengedFilter.Filter(buffers)
				assert.ElementsMatch(t, test.expectedFilteredBuffers[index], filtered)
				assert.ElementsMatch(t, test.expectedFilteredOutBuffers[index], filteredOut)
			}

		})
	}
}

func getTestBufferWithGeneration(bufferName string, generation int64) *v1.CapacityBuffer {
	return &v1.CapacityBuffer{
		ObjectMeta: metav1.ObjectMeta{
			Name:       bufferName,
			Namespace:  "default",
			Generation: generation,
		},
	}
}
