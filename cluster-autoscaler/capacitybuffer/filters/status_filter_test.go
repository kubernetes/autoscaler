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

func TestStatusFilter(t *testing.T) {
	tests := []struct {
		name                       string
		conditionsToFilterOut      map[string]string
		buffers                    []*v1.CapacityBuffer
		expectedFilteredBuffers    []*v1.CapacityBuffer
		expectedFilteredOutBuffers []*v1.CapacityBuffer
	}{
		{
			name:                  "Empty conditions map, filter all",
			conditionsToFilterOut: map[string]string{},
			buffers: []*v1.CapacityBuffer{
				getTestBufferWithCondition(testutil.SomePodTemplateRefName, testutil.GetConditionReady()),
			},
			expectedFilteredBuffers: []*v1.CapacityBuffer{
				getTestBufferWithCondition(testutil.SomePodTemplateRefName, testutil.GetConditionReady()),
			},
			expectedFilteredOutBuffers: []*v1.CapacityBuffer{},
		},
		{
			name:                  "One buffer, filter out ready for provisioning",
			conditionsToFilterOut: map[string]string{common.ReadyForProvisioningCondition: common.ConditionTrue},
			buffers: []*v1.CapacityBuffer{
				getTestBufferWithCondition(testutil.SomePodTemplateRefName, testutil.GetConditionReady()),
			},
			expectedFilteredBuffers: []*v1.CapacityBuffer{},
			expectedFilteredOutBuffers: []*v1.CapacityBuffer{
				getTestBufferWithCondition(testutil.SomePodTemplateRefName, testutil.GetConditionReady()),
			},
		},
		{
			name:                  "Two buffers, one Filtered",
			conditionsToFilterOut: map[string]string{common.ReadyForProvisioningCondition: common.ConditionTrue},
			buffers: []*v1.CapacityBuffer{
				getTestBufferWithCondition(testutil.SomePodTemplateRefName, testutil.GetConditionReady()),
				getTestBufferWithCondition(testutil.AnotherPodTemplateRefName, testutil.GetConditionNotReady()),
			},
			expectedFilteredBuffers: []*v1.CapacityBuffer{
				getTestBufferWithCondition(testutil.AnotherPodTemplateRefName, testutil.GetConditionNotReady()),
			},
			expectedFilteredOutBuffers: []*v1.CapacityBuffer{
				getTestBufferWithCondition(testutil.SomePodTemplateRefName, testutil.GetConditionReady()),
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			statusFilter := NewStatusFilter(test.conditionsToFilterOut)
			filtered, filteredOut := statusFilter.Filter(test.buffers)
			assert.ElementsMatch(t, test.expectedFilteredBuffers, filtered)
			assert.ElementsMatch(t, test.expectedFilteredOutBuffers, filteredOut)
		})
	}
}

func getTestBufferWithCondition(podTemplateRefName string, condition []metav1.Condition) *v1.CapacityBuffer {
	return testutil.GetBuffer(nil, &v1.LocalObjectRef{Name: podTemplateRefName}, nil, nil, nil, nil, condition, nil)
}
