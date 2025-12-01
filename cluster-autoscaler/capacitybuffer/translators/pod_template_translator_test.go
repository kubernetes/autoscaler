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

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	fakeClient "k8s.io/client-go/kubernetes/fake"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/autoscaler/cluster-autoscaler/apis/capacitybuffer/autoscaling.x-k8s.io/v1alpha1"
	cbclient "k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/client"
	"k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/testutil"
)

func TestPodTemplateBufferTranslator(t *testing.T) {
	registeredPodTemplate := &corev1.PodTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name:       testutil.SomePodTemplateRefName,
			Namespace:  "default",
			Generation: 1,
		},
	}
	anotherRegisteredPodTemplate := &corev1.PodTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name:       testutil.AnotherPodTemplateRefName,
			Namespace:  "default",
			Generation: 5,
		},
	}
	fakeClient := fakeClient.NewSimpleClientset(registeredPodTemplate, anotherRegisteredPodTemplate)
	fakeCapacityBuffersClient, _ := cbclient.NewCapacityBufferClient(nil, fakeClient, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	tests := []struct {
		name                   string
		buffers                []*v1.CapacityBuffer
		expectedStatus         []*v1.CapacityBufferStatus
		expectedNumberOfErrors int
	}{
		{
			name: "Test 1 buffer with pod template ref",
			buffers: []*v1.CapacityBuffer{
				testutil.GetPodTemplateRefBuffer(&v1.LocalObjectRef{Name: registeredPodTemplate.Name}, &testutil.SomeNumberOfReplicas),
			},
			expectedStatus: []*v1.CapacityBufferStatus{
				testutil.GetBufferStatus(&v1.LocalObjectRef{Name: registeredPodTemplate.Name}, &testutil.SomeNumberOfReplicas, &registeredPodTemplate.Generation, &testutil.ProvisioningStrategy, testutil.GetConditionReady()),
			},
			expectedNumberOfErrors: 0,
		},
		{
			name: "Test 2 buffers with pod template ref",
			buffers: []*v1.CapacityBuffer{
				testutil.GetPodTemplateRefBuffer(&v1.LocalObjectRef{Name: registeredPodTemplate.Name}, &testutil.SomeNumberOfReplicas),
				testutil.GetPodTemplateRefBuffer(&v1.LocalObjectRef{Name: anotherRegisteredPodTemplate.Name}, &testutil.AnotherNumberOfReplicas),
			},
			expectedStatus: []*v1.CapacityBufferStatus{
				testutil.GetBufferStatus(&v1.LocalObjectRef{Name: registeredPodTemplate.Name}, &testutil.SomeNumberOfReplicas, &registeredPodTemplate.Generation, &testutil.ProvisioningStrategy, testutil.GetConditionReady()),
				testutil.GetBufferStatus(&v1.LocalObjectRef{Name: anotherRegisteredPodTemplate.Name}, &testutil.AnotherNumberOfReplicas, &anotherRegisteredPodTemplate.Generation, &testutil.ProvisioningStrategy, testutil.GetConditionReady()),
			},
			expectedNumberOfErrors: 0,
		},
		{
			name: "Test 2 buffers one with not existing podTemplateRef",
			buffers: []*v1.CapacityBuffer{
				testutil.GetPodTemplateRefBuffer(&v1.LocalObjectRef{Name: registeredPodTemplate.Name}, &testutil.SomeNumberOfReplicas),
				testutil.GetPodTemplateRefBuffer(&v1.LocalObjectRef{Name: "randomRef"}, &testutil.AnotherNumberOfReplicas),
			},
			expectedStatus: []*v1.CapacityBufferStatus{
				testutil.GetBufferStatus(&v1.LocalObjectRef{Name: registeredPodTemplate.Name}, &testutil.SomeNumberOfReplicas, &registeredPodTemplate.Generation, &testutil.ProvisioningStrategy, testutil.GetConditionReady()),
				testutil.GetBufferStatus(nil, nil, nil, &testutil.ProvisioningStrategy, testutil.GetConditionNotReady()),
			},
			expectedNumberOfErrors: 1,
		},
		{
			name: "Test 2 buffers, one with no replicas",
			buffers: []*v1.CapacityBuffer{
				testutil.GetPodTemplateRefBuffer(&v1.LocalObjectRef{Name: registeredPodTemplate.Name}, &testutil.SomeNumberOfReplicas),
				testutil.GetPodTemplateRefBuffer(&v1.LocalObjectRef{Name: anotherRegisteredPodTemplate.Name}, nil),
			},
			expectedStatus: []*v1.CapacityBufferStatus{
				testutil.GetBufferStatus(&v1.LocalObjectRef{Name: registeredPodTemplate.Name}, &testutil.SomeNumberOfReplicas, &registeredPodTemplate.Generation, &testutil.ProvisioningStrategy, testutil.GetConditionReady()),
				testutil.GetBufferStatus(&v1.LocalObjectRef{Name: anotherRegisteredPodTemplate.Name}, nil, &anotherRegisteredPodTemplate.Generation, &testutil.ProvisioningStrategy, testutil.GetConditionNotReady()),
			},
			expectedNumberOfErrors: 0,
		},
		{
			name: "Test 2 buffers, one with nil podTemplateRef",
			buffers: []*v1.CapacityBuffer{
				testutil.GetPodTemplateRefBuffer(&v1.LocalObjectRef{Name: registeredPodTemplate.Name}, &testutil.SomeNumberOfReplicas),
				testutil.GetPodTemplateRefBuffer(nil, &testutil.AnotherNumberOfReplicas),
			},
			expectedStatus: []*v1.CapacityBufferStatus{
				testutil.GetBufferStatus(&v1.LocalObjectRef{Name: registeredPodTemplate.Name}, &testutil.SomeNumberOfReplicas, &registeredPodTemplate.Generation, &testutil.ProvisioningStrategy, testutil.GetConditionReady()),
				testutil.GetBufferStatus(nil, nil, nil, nil, nil),
			},
			expectedNumberOfErrors: 0,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			podTemplateBufferTranslator := NewPodTemplateBufferTranslator(fakeCapacityBuffersClient)
			errors := podTemplateBufferTranslator.Translate(test.buffers)
			assert.Equal(t, len(errors), test.expectedNumberOfErrors)
			assert.ElementsMatch(t, test.expectedStatus, testutil.SanitizeBuffersStatus(test.buffers))
		})
	}
}
