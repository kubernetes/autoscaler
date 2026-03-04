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

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/capacitybuffer"
	"k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/fakepods"
	fakeClient "k8s.io/client-go/kubernetes/fake"
	"k8s.io/utils/ptr"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/autoscaler/cluster-autoscaler/apis/capacitybuffer/autoscaling.x-k8s.io/v1beta1"
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
	zero := int64(0)
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
				testutil.GetBufferStatus(&v1.LocalObjectRef{Name: "capacitybuffer--pod-template"}, &testutil.SomeNumberOfReplicas, &zero, &testutil.ProvisioningStrategy, testutil.GetConditionReady()),
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
				testutil.GetBufferStatus(&v1.LocalObjectRef{Name: "capacitybuffer--pod-template"}, &testutil.SomeNumberOfReplicas, &zero, &testutil.ProvisioningStrategy, testutil.GetConditionReady()),
				testutil.GetBufferStatus(&v1.LocalObjectRef{Name: "capacitybuffer--pod-template"}, &testutil.AnotherNumberOfReplicas, &zero, &testutil.ProvisioningStrategy, testutil.GetConditionReady()),
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
				testutil.GetBufferStatus(&v1.LocalObjectRef{Name: "capacitybuffer--pod-template"}, &testutil.SomeNumberOfReplicas, &zero, &testutil.ProvisioningStrategy, testutil.GetConditionReady()),
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
				testutil.GetBufferStatus(&v1.LocalObjectRef{Name: "capacitybuffer--pod-template"}, &testutil.SomeNumberOfReplicas, &zero, &testutil.ProvisioningStrategy, testutil.GetConditionReady()),
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
				testutil.GetBufferStatus(&v1.LocalObjectRef{Name: "capacitybuffer--pod-template"}, &testutil.SomeNumberOfReplicas, &zero, &testutil.ProvisioningStrategy, testutil.GetConditionReady()),
				testutil.GetBufferStatus(nil, nil, nil, nil, nil),
			},
			expectedNumberOfErrors: 0,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resolver := fakepods.NewDefaultingResolver()
			podTemplateBufferTranslator := NewPodTemplateBufferTranslator(fakeCapacityBuffersClient, resolver)
			errors := podTemplateBufferTranslator.Translate(test.buffers)
			assert.Equal(t, len(errors), test.expectedNumberOfErrors)
			assert.ElementsMatch(t, test.expectedStatus, testutil.SanitizeBuffersStatus(test.buffers))
		})
	}
}

func TestPodTemplateBufferTranslator_ManagedPodTemplate(t *testing.T) {
	bufferPodTemplate := testutil.NewPodTemplate(
		testutil.WithPodTemplateName("test-pod-template"),
		withResources(nil, corev1.ResourceList{
			"cpu": resource.MustParse("1000m"),
		}),
	)
	buffer := testutil.NewBuffer(
		testutil.WithName("test-buffer"),
		testutil.WithPodTemplateRef("test-pod-template"),
	)

	wantManagedTemplate := testutil.NewPodTemplate(
		testutil.WithPodTemplateName("capacitybuffer-test-buffer-pod-template"),
		withResources(
			corev1.ResourceList{
				"cpu": resource.MustParse("1000m"),
			}, corev1.ResourceList{
				"cpu": resource.MustParse("1000m"),
			},
		),
		func(template *corev1.PodTemplate) {
			template.OwnerReferences = []metav1.OwnerReference{
				{
					APIVersion: capacitybuffer.CapacityBufferApiVersion,
					Kind:       capacitybuffer.CapacityBufferKind,
					Name:       buffer.Name,
					UID:        buffer.UID,
					Controller: ptr.To(true),
				},
			}
		},
	)

	fakeClient := fakeClient.NewSimpleClientset(bufferPodTemplate)
	fakeCapacityBuffersClient, _ := cbclient.NewCapacityBufferClient(nil, fakeClient, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	resolver := fakepods.NewDefaultingResolver()
	podTemplateBufferTranslator := NewPodTemplateBufferTranslator(fakeCapacityBuffersClient, resolver)
	buffers := []*v1.CapacityBuffer{buffer}
	errors := podTemplateBufferTranslator.Translate(buffers)
	assert.Equal(t, 0, len(errors))

	gotManagedPodTemplate, err := fakeClient.CoreV1().PodTemplates("default").Get(t.Context(), "capacitybuffer-test-buffer-pod-template", metav1.GetOptions{})
	assert.NoError(t, err)
	if diff := cmp.Diff(wantManagedTemplate, gotManagedPodTemplate, cmpopts.EquateEmpty()); diff != "" {
		t.Errorf("managed pod template mismatch (-want +got):\n%s", diff)
	}
}

func withResources(requests, limits corev1.ResourceList) func(*corev1.PodTemplate) {
	return func(template *corev1.PodTemplate) {
		template.Template.Spec.Containers = []corev1.Container{
			{
				Name:  "test",
				Image: "test",
				Resources: corev1.ResourceRequirements{
					Limits:   limits,
					Requests: requests,
				},
			},
		}
	}
}
