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

// strategyFilter filters out buffers with provisioning strategies not defined in strategiesToUse
// and defaults nil values of provisioningStrategy to empty string
type strategyFilter struct {
	strategiesToUse map[string]bool
}

// NewStrategyFilter creates an instance of strategyFilter.
func NewStrategyFilter(strategiesToUse []string) *strategyFilter {
	strategiesToUseMap := map[string]bool{}
	for _, strategy := range strategiesToUse {
		strategiesToUseMap[strategy] = true
	}
	return &strategyFilter{
		strategiesToUse: strategiesToUseMap,
	}
}

// Filter filters out buffers with provisioning strategies not defined in strategiesToUseMap
func (f *strategyFilter) Filter(buffers []*v1.CapacityBuffer) ([]*v1.CapacityBuffer, []*v1.CapacityBuffer) {

	var filteredBuffers []*v1.CapacityBuffer
	var filteredOutBuffers []*v1.CapacityBuffer

	for _, buffer := range buffers {
		if f.isAllowedProvisioningStrategy(buffer) {
			filteredBuffers = append(filteredBuffers, buffer)
		} else {
			filteredOutBuffers = append(filteredOutBuffers, buffer)
		}
	}
	return filteredBuffers, filteredOutBuffers
}

func (f *strategyFilter) isAllowedProvisioningStrategy(buffer *v1.CapacityBuffer) bool {
	provisioningStrategy := ""
	if buffer.Spec.ProvisioningStrategy != nil {
		provisioningStrategy = *buffer.Spec.ProvisioningStrategy
	}

	if useStrategy, found := f.strategiesToUse[provisioningStrategy]; found && useStrategy {
		return true
	}

	return false
}

// CleanUp cleans up the filter's internal structures.
func (f *strategyFilter) CleanUp() {
}
