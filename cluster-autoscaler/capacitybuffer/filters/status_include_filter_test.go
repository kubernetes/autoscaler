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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/autoscaler/cluster-autoscaler/apis/capacitybuffer/autoscaling.x-k8s.io/v1alpha1"
	"k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/common"
	"k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/testutil"
)

func TestStatusIncludeFilter(t *testing.T) {
	tests := []struct {
		name                    string
		conditionsToInclude     map[string]string
		buffers                 []*v1.CapacityBuffer
		expectedIncludedBuffers []*v1.CapacityBuffer
		expectedExcludedBuffers []*v1.CapacityBuffer
	}{
		{
			name:                "Empty conditions map, include all",
			conditionsToInclude: map[string]string{},
			buffers: []*v1.CapacityBuffer{
				getTestBufferWithCondition(testutil.SomePodTemplateRefName, testutil.GetConditionReady()),
			},
			expectedIncludedBuffers: []*v1.CapacityBuffer{
				getTestBufferWithCondition(testutil.SomePodTemplateRefName, testutil.GetConditionReady()),
			},
			expectedExcludedBuffers: []*v1.CapacityBuffer{},
		},
		{
			name:                "One buffer, include only ready for provisioning",
			conditionsToInclude: map[string]string{common.ReadyForProvisioningCondition: common.ConditionTrue},
			buffers: []*v1.CapacityBuffer{
				getTestBufferWithCondition(testutil.SomePodTemplateRefName, testutil.GetConditionReady()),
			},
			expectedIncludedBuffers: []*v1.CapacityBuffer{
				getTestBufferWithCondition(testutil.SomePodTemplateRefName, testutil.GetConditionReady()),
			},
			expectedExcludedBuffers: []*v1.CapacityBuffer{},
		},
		{
			name:                "Two buffers, one included",
			conditionsToInclude: map[string]string{common.ReadyForProvisioningCondition: common.ConditionTrue},
			buffers: []*v1.CapacityBuffer{
				getTestBufferWithCondition(testutil.SomePodTemplateRefName, testutil.GetConditionReady()),
				getTestBufferWithCondition(testutil.AnotherPodTemplateRefName, testutil.GetConditionNotReady()),
			},
			expectedIncludedBuffers: []*v1.CapacityBuffer{
				getTestBufferWithCondition(testutil.SomePodTemplateRefName, testutil.GetConditionReady()),
			},
			expectedExcludedBuffers: []*v1.CapacityBuffer{
				getTestBufferWithCondition(testutil.AnotherPodTemplateRefName, testutil.GetConditionNotReady()),
			},
		},
		{
			name:                "Buffer with no conditions, excluded",
			conditionsToInclude: map[string]string{common.ReadyForProvisioningCondition: common.ConditionTrue},
			buffers: []*v1.CapacityBuffer{
				getTestBufferWithCondition(testutil.SomePodTemplateRefName, []metav1.Condition{}),
			},
			expectedIncludedBuffers: []*v1.CapacityBuffer{},
			expectedExcludedBuffers: []*v1.CapacityBuffer{
				getTestBufferWithCondition(testutil.SomePodTemplateRefName, []metav1.Condition{}),
			},
		},
		{
			name: "Multiple conditions required, all must match",
			conditionsToInclude: map[string]string{
				common.ReadyForProvisioningCondition: common.ConditionTrue,
				common.ProvisioningCondition:         common.ConditionTrue,
			},
			buffers: []*v1.CapacityBuffer{
				getTestBufferWithCondition(testutil.SomePodTemplateRefName, []metav1.Condition{
					{Type: common.ReadyForProvisioningCondition, Status: common.ConditionTrue},
					{Type: common.ProvisioningCondition, Status: common.ConditionTrue},
				}),
				getTestBufferWithCondition(testutil.AnotherPodTemplateRefName, testutil.GetConditionReady()),
			},
			expectedIncludedBuffers: []*v1.CapacityBuffer{
				getTestBufferWithCondition(testutil.SomePodTemplateRefName, []metav1.Condition{
					{Type: common.ReadyForProvisioningCondition, Status: common.ConditionTrue},
					{Type: common.ProvisioningCondition, Status: common.ConditionTrue},
				}),
			},
			expectedExcludedBuffers: []*v1.CapacityBuffer{
				getTestBufferWithCondition(testutil.AnotherPodTemplateRefName, testutil.GetConditionReady()),
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			statusIncludeFilter := NewStatusIncludeFilter(test.conditionsToInclude)
			included, excluded := statusIncludeFilter.Filter(test.buffers)
			assert.ElementsMatch(t, test.expectedIncludedBuffers, included)
			assert.ElementsMatch(t, test.expectedExcludedBuffers, excluded)
		})
	}
}
