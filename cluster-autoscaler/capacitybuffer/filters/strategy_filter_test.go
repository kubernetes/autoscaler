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
	"k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/testutil"
)

func TestStrategyFilter(t *testing.T) {
	someRandomStrategy := "someStrategy"
	tests := []struct {
		name                       string
		buffers                    []*v1.CapacityBuffer
		strategiesToConsider       []string
		expectedFilteredBuffers    []*v1.CapacityBuffer
		expectedFilteredOutBuffers []*v1.CapacityBuffer
	}{
		{
			name: "Single buffer with accepted strategy",
			buffers: []*v1.CapacityBuffer{
				getTestBufferWithStrategy(&testutil.ProvisioningStrategy),
			},
			strategiesToConsider: []string{testutil.ProvisioningStrategy},
			expectedFilteredBuffers: []*v1.CapacityBuffer{
				getTestBufferWithStrategy(&testutil.ProvisioningStrategy),
			},
			expectedFilteredOutBuffers: []*v1.CapacityBuffer{},
		},
		{
			name: "Nil strategy defaulting to empty",
			buffers: []*v1.CapacityBuffer{
				getTestBufferWithStrategy(nil),
			},
			strategiesToConsider: []string{""},
			expectedFilteredBuffers: []*v1.CapacityBuffer{
				getTestBufferWithStrategy(nil),
			},
			expectedFilteredOutBuffers: []*v1.CapacityBuffer{},
		},
		{
			name: "Single buffer with rejected strategy",
			buffers: []*v1.CapacityBuffer{
				getTestBufferWithStrategy(&someRandomStrategy),
			},
			strategiesToConsider:    []string{testutil.ProvisioningStrategy},
			expectedFilteredBuffers: []*v1.CapacityBuffer{},
			expectedFilteredOutBuffers: []*v1.CapacityBuffer{
				getTestBufferWithStrategy(&someRandomStrategy),
			},
		},
		{
			name: "Multiple buffers different strategies",
			buffers: []*v1.CapacityBuffer{
				getTestBufferWithStrategy(&someRandomStrategy),
				getTestBufferWithStrategy(&testutil.ProvisioningStrategy),
				getTestBufferWithStrategy(nil),
			},
			strategiesToConsider: []string{testutil.ProvisioningStrategy, ""},
			expectedFilteredBuffers: []*v1.CapacityBuffer{
				getTestBufferWithStrategy(&testutil.ProvisioningStrategy),
				getTestBufferWithStrategy(nil),
			},
			expectedFilteredOutBuffers: []*v1.CapacityBuffer{
				getTestBufferWithStrategy(&someRandomStrategy),
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			strategyFilter := NewStrategyFilter(test.strategiesToConsider)
			filtered, filteredOut := strategyFilter.Filter(test.buffers)
			assert.ElementsMatch(t, test.expectedFilteredBuffers, filtered)
			assert.ElementsMatch(t, test.expectedFilteredOutBuffers, filteredOut)
		})
	}
}

func getTestBufferWithStrategy(provisioningStrategy *string) *v1.CapacityBuffer {
	return testutil.GetBuffer(provisioningStrategy, nil, nil, nil, nil, nil, nil, nil)
}
