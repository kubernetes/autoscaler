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

package translator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/autoscaler/cluster-autoscaler/apis/capacitybuffer/autoscaling.x-k8s.io/v1"
	"k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/testutil"
)

func TestPodTemplateBufferTranslator(t *testing.T) {
	podTemplateBufferTranslator := NewPodTemplateBufferTranslator()
	tests := []struct {
		name                   string
		buffers                []*v1.CapacityBuffer
		expectedStatus         []*v1.CapacityBufferStatus
		expectedNumberOfErrors int
	}{
		{
			name: "Test 1 buffer with pod template ref",
			buffers: []*v1.CapacityBuffer{
				testutil.GetPodTemplateRefBuffer(&v1.LocalObjectRef{Name: testutil.SomePodTemplateRefName}, &testutil.SomeNumberOfReplicas),
			},
			expectedStatus: []*v1.CapacityBufferStatus{
				testutil.GetBufferStatus(&v1.LocalObjectRef{Name: testutil.SomePodTemplateRefName}, &testutil.SomeNumberOfReplicas, testutil.GetConditionReady()),
			},
			expectedNumberOfErrors: 0,
		},
		{
			name: "Test 2 buffers with pod template ref",
			buffers: []*v1.CapacityBuffer{
				testutil.GetPodTemplateRefBuffer(&v1.LocalObjectRef{Name: testutil.SomePodTemplateRefName}, &testutil.SomeNumberOfReplicas),
				testutil.GetPodTemplateRefBuffer(&v1.LocalObjectRef{Name: testutil.AnotherPodTemplateRefName}, &testutil.AnotherNumberOfReplicas),
			},
			expectedStatus: []*v1.CapacityBufferStatus{
				testutil.GetBufferStatus(&v1.LocalObjectRef{Name: testutil.SomePodTemplateRefName}, &testutil.SomeNumberOfReplicas, testutil.GetConditionReady()),
				testutil.GetBufferStatus(&v1.LocalObjectRef{Name: testutil.AnotherPodTemplateRefName}, &testutil.AnotherNumberOfReplicas, testutil.GetConditionReady()),
			},
			expectedNumberOfErrors: 0,
		},
		{
			name: "Test 2 buffers, one with no replicas",
			buffers: []*v1.CapacityBuffer{
				testutil.GetPodTemplateRefBuffer(&v1.LocalObjectRef{Name: testutil.SomePodTemplateRefName}, &testutil.SomeNumberOfReplicas),
				testutil.GetPodTemplateRefBuffer(&v1.LocalObjectRef{Name: testutil.AnotherPodTemplateRefName}, nil),
			},
			expectedStatus: []*v1.CapacityBufferStatus{
				testutil.GetBufferStatus(&v1.LocalObjectRef{Name: testutil.SomePodTemplateRefName}, &testutil.SomeNumberOfReplicas, testutil.GetConditionReady()),
				testutil.GetBufferStatus(nil, nil, testutil.GetConditionNotReady()),
			},
			expectedNumberOfErrors: 1,
		},
		{
			name: "Test 2 buffers, one with no pod template ref",
			buffers: []*v1.CapacityBuffer{
				testutil.GetPodTemplateRefBuffer(&v1.LocalObjectRef{Name: testutil.SomePodTemplateRefName}, &testutil.SomeNumberOfReplicas),
				testutil.GetPodTemplateRefBuffer(nil, &testutil.AnotherNumberOfReplicas),
			},
			expectedStatus: []*v1.CapacityBufferStatus{
				testutil.GetBufferStatus(&v1.LocalObjectRef{Name: testutil.SomePodTemplateRefName}, &testutil.SomeNumberOfReplicas, testutil.GetConditionReady()),
				testutil.GetBufferStatus(nil, nil, nil),
			},
			expectedNumberOfErrors: 0,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			errors := podTemplateBufferTranslator.Translate(test.buffers)
			assert.Equal(t, len(errors), test.expectedNumberOfErrors)
			assert.ElementsMatch(t, test.expectedStatus, testutil.SanitizeBuffersStatus(test.buffers))
		})
	}
}
