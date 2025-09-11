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
	v1 "k8s.io/autoscaler/cluster-autoscaler/apis/capacitybuffer/autoscaling.x-k8s.io/v1"
	"k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/common"
	"k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/testutil"
)

func TestStatusFilter(t *testing.T) {
	tests := []struct {
		name                       string
		conditions                 map[string]string
		buffers                    []*v1.CapacityBuffer
		expectedFilteredBuffers    []*v1.CapacityBuffer
		expectedFilteredOutBuffers []*v1.CapacityBuffer
	}{
		{
			name:       "Empty conditions, filter none",
			conditions: map[string]string{},
			buffers: []*v1.CapacityBuffer{
				testutil.GetPodTemplateRefBuffer(&v1.LocalObjectRef{Name: testutil.SomePodTemplateRefName}, nil),
			},
			expectedFilteredBuffers: []*v1.CapacityBuffer{
				testutil.GetPodTemplateRefBuffer(&v1.LocalObjectRef{Name: testutil.SomePodTemplateRefName}, nil),
			},
			expectedFilteredOutBuffers: []*v1.CapacityBuffer{},
		},
		{
			name:       "Some condition, filter one",
			conditions: map[string]string{common.ReadyForProvisioningCondition: common.ConditionTrue},
			buffers: []*v1.CapacityBuffer{
				testutil.GetBuffer(&testutil.ProvisioningStrategy, &v1.LocalObjectRef{Name: testutil.SomePodTemplateRefName}, nil, nil, nil, testutil.GetConditionReady()),
			},
			expectedFilteredBuffers: []*v1.CapacityBuffer{},
			expectedFilteredOutBuffers: []*v1.CapacityBuffer{
				testutil.GetBuffer(&testutil.ProvisioningStrategy, &v1.LocalObjectRef{Name: testutil.SomePodTemplateRefName}, nil, nil, nil, testutil.GetConditionReady()),
			},
		},
		{
			name:       "Some condition, filter one in and one out",
			conditions: map[string]string{common.ReadyForProvisioningCondition: common.ConditionTrue},
			buffers: []*v1.CapacityBuffer{
				testutil.GetBuffer(&testutil.ProvisioningStrategy, &v1.LocalObjectRef{Name: testutil.SomePodTemplateRefName}, nil, nil, nil, testutil.GetConditionReady()),
				testutil.GetBuffer(&testutil.ProvisioningStrategy, &v1.LocalObjectRef{Name: testutil.AnotherPodTemplateRefName}, nil, nil, nil, testutil.GetConditionNotReady()),
			},
			expectedFilteredBuffers: []*v1.CapacityBuffer{
				testutil.GetBuffer(&testutil.ProvisioningStrategy, &v1.LocalObjectRef{Name: testutil.AnotherPodTemplateRefName}, nil, nil, nil, testutil.GetConditionNotReady()),
			},
			expectedFilteredOutBuffers: []*v1.CapacityBuffer{
				testutil.GetBuffer(&testutil.ProvisioningStrategy, &v1.LocalObjectRef{Name: testutil.SomePodTemplateRefName}, nil, nil, nil, testutil.GetConditionReady()),
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			statusFilter := NewStatusFilter(test.conditions)
			filtered, filteredOut := statusFilter.Filter(test.buffers)
			assert.ElementsMatch(t, test.expectedFilteredBuffers, filtered)
			assert.ElementsMatch(t, test.expectedFilteredOutBuffers, filteredOut)
		})
	}
}
