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
	"fmt"
	"math"
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
	podTemp4mem100cpu := getPodTemplateWithResources("podTemp4mem100cpu", corev1.ResourceList{
		"memory": resource.MustParse("4Gi"),
		"cpu":    resource.MustParse("100m"),
	})
	podTemp8mem200cpu := getPodTemplateWithResources("podTemp8mem200cpu", corev1.ResourceList{
		"memory": resource.MustParse("8Gi"),
		"cpu":    resource.MustParse("200m"),
	})
	podTemp4gpu := getPodTemplateWithResources("podTemp4gpu", corev1.ResourceList{
		"nvidia.com/gpu": resource.MustParse("4"),
		"memory":         resource.MustParse("16Gi"),
		"cpu":            resource.MustParse("1000m"),
	})
	zero := int64(0)
	fakeClient := fakeClient.NewSimpleClientset(registeredPodTemplate, anotherRegisteredPodTemplate, podTemp4mem100cpu, podTemp8mem200cpu, podTemp4gpu)
	fakeCapacityBuffersClient, _ := cbclient.NewCapacityBufferClient(nil, fakeClient, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	noReplicasMsg := "Buffer not ready for provisioning: couldn't get number of replicas for buffer: replicas, percentage and limits are not defined"
	podTemplateNotFoundMsg := "Buffer not ready for provisioning: capacity buffer client can't get pod template: podtemplates %q not found"
	resourcesNotFoundMsg := "Buffer not ready for provisioning: couldn't get number of replicas for buffer: resources in configured limits not found in the pod template"
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
				testutil.GetBufferStatus(nil, nil, nil, &testutil.ProvisioningStrategy, testutil.GetConditionNotReadyWithMessage(fmt.Sprintf(podTemplateNotFoundMsg, "randomRef"))),
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
				testutil.GetBufferStatus(&v1.LocalObjectRef{Name: "capacitybuffer--pod-template"}, nil, &zero, &testutil.ProvisioningStrategy, testutil.GetConditionNotReadyWithMessage(noReplicasMsg)),
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
		{
			name: "Limits exist and no replicas, buffer filtered",
			buffers: []*v1.CapacityBuffer{
				getTestBufferWithLimits(&v1.LocalObjectRef{Name: podTemp4mem100cpu.Name}, nil, &v1.ResourceList{
					"memory": resource.MustParse("5Gi"),
					"cpu":    resource.MustParse("200m"),
				}),
			},
			expectedStatus: []*v1.CapacityBufferStatus{

				getTestBufferStatusWithReplicas(&v1.LocalObjectRef{Name: "capacitybuffer--pod-template"}, 1),
			},
			expectedNumberOfErrors: 0,
		},
		{
			name: "Limits exist and with higher replicas",
			buffers: []*v1.CapacityBuffer{
				getTestBufferWithLimits(&v1.LocalObjectRef{Name: podTemp4mem100cpu.Name}, ptr.To[int32](3), &v1.ResourceList{
					"memory": resource.MustParse("9Gi"),
					"cpu":    resource.MustParse("200m"),
				}),
			},
			expectedStatus: []*v1.CapacityBufferStatus{
				getTestBufferStatusWithReplicas(&v1.LocalObjectRef{Name: "capacitybuffer--pod-template"}, 2),
			},
			expectedNumberOfErrors: 0,
		},
		{
			name: "Limits exist and with lower replicas",
			buffers: []*v1.CapacityBuffer{
				getTestBufferWithLimits(&v1.LocalObjectRef{Name: podTemp4mem100cpu.Name}, ptr.To[int32](1), &v1.ResourceList{
					"memory": resource.MustParse("9Gi"),
					"cpu":    resource.MustParse("200m"),
				}),
			},
			expectedStatus: []*v1.CapacityBufferStatus{
				getTestBufferStatusWithReplicas(&v1.LocalObjectRef{Name: "capacitybuffer--pod-template"}, 1),
			},
			expectedNumberOfErrors: 0,
		},
		{
			name: "Limits with gpu resource",
			buffers: []*v1.CapacityBuffer{
				getTestBufferWithLimits(&v1.LocalObjectRef{Name: podTemp4gpu.Name}, ptr.To[int32](5), &v1.ResourceList{
					"memory":         resource.MustParse("100Gi"),
					"cpu":            resource.MustParse("5000m"),
					"nvidia.com/gpu": resource.MustParse("10"),
				}),
			},
			expectedStatus: []*v1.CapacityBufferStatus{
				getTestBufferStatusWithReplicas(&v1.LocalObjectRef{Name: "capacitybuffer--pod-template"}, 2),
			},
			expectedNumberOfErrors: 0,
		},
		{
			name: "Limits less than pod template requested resources",
			buffers: []*v1.CapacityBuffer{
				getTestBufferWithLimits(&v1.LocalObjectRef{Name: podTemp4gpu.Name}, ptr.To[int32](5), &v1.ResourceList{
					"memory":         resource.MustParse("100Gi"),
					"cpu":            resource.MustParse("5000m"),
					"nvidia.com/gpu": resource.MustParse("2"),
				}),
				getTestBufferWithLimits(&v1.LocalObjectRef{Name: podTemp8mem200cpu.Name}, ptr.To[int32](3), &v1.ResourceList{
					"cpu": resource.MustParse("100m"),
				}),
			},
			expectedStatus: []*v1.CapacityBufferStatus{
				getTestBufferStatusWithReplicas(&v1.LocalObjectRef{Name: "capacitybuffer--pod-template"}, 0),
				getTestBufferStatusWithReplicas(&v1.LocalObjectRef{Name: "capacitybuffer--pod-template"}, 0),
			},
			expectedNumberOfErrors: 0,
		},
		{
			name: "Defined limits doesn't have any of the resources in pod template",
			buffers: []*v1.CapacityBuffer{
				getTestBufferWithLimits(&v1.LocalObjectRef{Name: podTemp4mem100cpu.Name}, ptr.To[int32](3), &v1.ResourceList{
					"nvidia.com/gpu": resource.MustParse("100m"),
				}),
			},
			expectedStatus: []*v1.CapacityBufferStatus{
				testutil.GetBufferStatus(
					&v1.LocalObjectRef{Name: "capacitybuffer--pod-template"}, nil, ptr.To[int64](0), &testutil.ProvisioningStrategy, testutil.GetConditionNotReadyWithMessage(resourcesNotFoundMsg)),
			},
			expectedNumberOfErrors: 0,
		},
		{
			name: "Replicas set to MaxInt32",
			buffers: []*v1.CapacityBuffer{
				testutil.GetPodTemplateRefBuffer(&v1.LocalObjectRef{Name: registeredPodTemplate.Name}, ptr.To[int32](math.MaxInt32)),
			},
			expectedStatus: []*v1.CapacityBufferStatus{
				testutil.GetBufferStatus(
					&v1.LocalObjectRef{Name: "capacitybuffer--pod-template"}, ptr.To[int32](math.MaxInt32), ptr.To[int64](0), &testutil.ProvisioningStrategy, testutil.GetConditionReady(),
				),
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
			for i, buffer := range test.buffers {
				wantStatus := test.expectedStatus[i]
				if diff := cmp.Diff(wantStatus, &buffer.Status, cmpopts.IgnoreFields(metav1.Condition{}, "LastTransitionTime")); diff != "" {
					t.Errorf("buffer status mismatch (-want +got):\n%s", diff)
				}
			}
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

func getTestBufferWithLimits(podTemplateRef *v1.LocalObjectRef, replicas *int32, limits *v1.ResourceList) *v1.CapacityBuffer {
	opts := []testutil.BufferOption{testutil.WithActiveProvisioningStrategy()}
	if podTemplateRef != nil {
		opts = append(opts, testutil.WithPodTemplateRef(podTemplateRef.Name))
	}
	if replicas != nil {
		opts = append(opts, testutil.WithReplicas(*replicas))
	}
	if limits != nil {
		opts = append(opts, testutil.WithLimits(*limits))
	}
	buffer := testutil.NewBuffer(opts...)
	return buffer
}

func getTestBufferStatusWithReplicas(podTemplateRef *v1.LocalObjectRef, replicas int32) *v1.CapacityBufferStatus {
	return testutil.GetBufferStatus(podTemplateRef, &replicas, ptr.To[int64](0), &testutil.ProvisioningStrategy, testutil.GetConditionReady())
}

func getPodTemplateWithResources(name string, resources corev1.ResourceList) *corev1.PodTemplate {
	return &corev1.PodTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name:       name,
			Namespace:  "default",
			Generation: 1,
		},
		Template: corev1.PodTemplateSpec{
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Resources: corev1.ResourceRequirements{
							Requests: resources,
						},
					},
				},
			},
		},
	}
}
