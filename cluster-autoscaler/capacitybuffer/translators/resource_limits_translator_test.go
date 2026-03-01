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
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/autoscaler/cluster-autoscaler/apis/capacitybuffer/autoscaling.x-k8s.io/v1beta1"
	cbclient "k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/client"
	"k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/testutil"
	fakeClient "k8s.io/client-go/kubernetes/fake"
	"k8s.io/utils/ptr"
)

func TestResourceLimitsTranslator(t *testing.T) {
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
	fakeClient := fakeClient.NewSimpleClientset(podTemp4mem100cpu, podTemp8mem200cpu, podTemp4gpu)
	fakeCapacityBuffersClient, _ := cbclient.NewCapacityBufferClient(nil, fakeClient, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	noResourcesSetMessage := "couldn't calculate number of pods for buffer based on provided resource limits. Check if the pod template requests at least one limited resource"

	tests := []struct {
		name                   string
		buffers                []*v1.CapacityBuffer
		expectedStatus         []*v1.CapacityBufferStatus
		expectedNumberOfErrors int
	}{
		{
			name: "Limits set to nil, buffer filtered out",
			buffers: []*v1.CapacityBuffer{
				getTestBufferWithLimits(nil, nil, nil),
			},
			expectedStatus: []*v1.CapacityBufferStatus{
				testutil.GetBufferStatus(nil, nil, nil, &testutil.ProvisioningStrategy, nil),
			},
			expectedNumberOfErrors: 0,
		},
		{
			name: "Limits exist, podTemplateRef is nil",
			buffers: []*v1.CapacityBuffer{
				getTestBufferWithLimits(nil, nil, &v1.ResourceList{
					"cpu": resource.MustParse("500m"),
				}),
			},
			expectedStatus: []*v1.CapacityBufferStatus{
				testutil.GetBufferStatus(nil, nil, nil, &testutil.ProvisioningStrategy, nil),
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
				getTestBufferStatusWithReplicas(&v1.LocalObjectRef{Name: podTemp4mem100cpu.Name}, 1),
			},
			expectedNumberOfErrors: 0,
		},
		{
			name: "Limits exist and no replicas, buffer filtered",
			buffers: []*v1.CapacityBuffer{
				getTestBufferWithLimits(&v1.LocalObjectRef{Name: podTemp4mem100cpu.Name}, nil, &v1.ResourceList{
					"memory": resource.MustParse("9Gi"),
					"cpu":    resource.MustParse("200m"),
				}),
			},
			expectedStatus: []*v1.CapacityBufferStatus{
				getTestBufferStatusWithReplicas(&v1.LocalObjectRef{Name: podTemp4mem100cpu.Name}, 2),
			},
			expectedNumberOfErrors: 0,
		},
		{
			name: "Limits exist and with bigger replicas",
			buffers: []*v1.CapacityBuffer{
				getTestBufferWithLimits(&v1.LocalObjectRef{Name: podTemp4mem100cpu.Name}, pointerToInt32(3), &v1.ResourceList{
					"memory": resource.MustParse("9Gi"),
					"cpu":    resource.MustParse("200m"),
				}),
			},
			expectedStatus: []*v1.CapacityBufferStatus{
				getTestBufferStatusWithReplicas(&v1.LocalObjectRef{Name: podTemp4mem100cpu.Name}, 2),
			},
			expectedNumberOfErrors: 0,
		},
		{
			name: "Limits exist and with smaller replicas",
			buffers: []*v1.CapacityBuffer{
				getTestBufferWithLimits(&v1.LocalObjectRef{Name: podTemp4mem100cpu.Name}, pointerToInt32(1), &v1.ResourceList{
					"memory": resource.MustParse("9Gi"),
					"cpu":    resource.MustParse("200m"),
				}),
			},
			expectedStatus: []*v1.CapacityBufferStatus{
				getTestBufferStatusWithReplicas(&v1.LocalObjectRef{Name: podTemp4mem100cpu.Name}, 1),
			},
			expectedNumberOfErrors: 0,
		},
		{
			name: "Limits exist and no replicas, buffer filtered",
			buffers: []*v1.CapacityBuffer{
				getTestBufferWithLimits(&v1.LocalObjectRef{Name: podTemp4mem100cpu.Name}, pointerToInt32(5), &v1.ResourceList{
					"memory": resource.MustParse("10Gi"),
					"cpu":    resource.MustParse("1000m"),
				}),
				getTestBufferWithLimits(&v1.LocalObjectRef{Name: podTemp8mem200cpu.Name}, pointerToInt32(3), &v1.ResourceList{
					"memory": resource.MustParse("100Gi"),
					"cpu":    resource.MustParse("10000m"),
				}),
			},
			expectedStatus: []*v1.CapacityBufferStatus{
				getTestBufferStatusWithReplicas(&v1.LocalObjectRef{Name: podTemp4mem100cpu.Name}, 2),
				getTestBufferStatusWithReplicas(&v1.LocalObjectRef{Name: podTemp8mem200cpu.Name}, 3),
			},
			expectedNumberOfErrors: 0,
		},
		{
			name: "Limits with gpu resource",
			buffers: []*v1.CapacityBuffer{
				getTestBufferWithLimits(&v1.LocalObjectRef{Name: podTemp4gpu.Name}, pointerToInt32(5), &v1.ResourceList{
					"memory":         resource.MustParse("100Gi"),
					"cpu":            resource.MustParse("5000m"),
					"nvidia.com/gpu": resource.MustParse("10"),
				}),
			},
			expectedStatus: []*v1.CapacityBufferStatus{
				getTestBufferStatusWithReplicas(&v1.LocalObjectRef{Name: podTemp4gpu.Name}, 2),
			},
			expectedNumberOfErrors: 0,
		},
		{
			name: "Limits less than pod template requested resources",
			buffers: []*v1.CapacityBuffer{
				getTestBufferWithLimits(&v1.LocalObjectRef{Name: podTemp4gpu.Name}, pointerToInt32(5), &v1.ResourceList{
					"memory":         resource.MustParse("100Gi"),
					"cpu":            resource.MustParse("5000m"),
					"nvidia.com/gpu": resource.MustParse("2"),
				}),
				getTestBufferWithLimits(&v1.LocalObjectRef{Name: podTemp8mem200cpu.Name}, pointerToInt32(3), &v1.ResourceList{
					"cpu": resource.MustParse("100m"),
				}),
			},
			expectedStatus: []*v1.CapacityBufferStatus{
				getTestBufferStatusWithReplicas(&v1.LocalObjectRef{Name: podTemp4gpu.Name}, 0),
				getTestBufferStatusWithReplicas(&v1.LocalObjectRef{Name: podTemp8mem200cpu.Name}, 0),
			},
			expectedNumberOfErrors: 0,
		},
		{
			name: "Defined limits doesn't have any of the resources in pod template",
			buffers: []*v1.CapacityBuffer{
				getTestBufferWithLimits(&v1.LocalObjectRef{Name: podTemp4mem100cpu.Name}, pointerToInt32(3), &v1.ResourceList{
					"nvidia.com/gpu": resource.MustParse("100m"),
				}),
			},
			expectedStatus: []*v1.CapacityBufferStatus{
				testutil.GetBufferStatus(
					nil, nil, nil, &testutil.ProvisioningStrategy, testutil.GetConditionNotReadyWithMessage(noResourcesSetMessage)),
			},
			expectedNumberOfErrors: 0,
		},
		{
			name: "conditions are not overridden if buffer is not ready",
			buffers: []*v1.CapacityBuffer{
				testutil.NewBuffer(
					testutil.WithPodTemplateRef(podTemp4mem100cpu.Name),
					testutil.WithLimits(v1.ResourceList{
						"nvidia.com/gpu": resource.MustParse("100m"),
					}),
					testutil.WithActiveProvisioningStrategy(),
					func(buffer *v1.CapacityBuffer) {
						buffer.Status.Conditions = testutil.GetConditionNotReadyWithMessage("test error")
					},
				),
			},
			expectedStatus: []*v1.CapacityBufferStatus{
				testutil.GetBufferStatus(nil, nil, nil, nil, testutil.GetConditionNotReadyWithMessage("test error")),
			},
			expectedNumberOfErrors: 0,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			translator := NewResourceLimitsTranslator(fakeCapacityBuffersClient)
			errors := translator.Translate(test.buffers)
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

func getTestBufferWithLimits(podTemplateRef *v1.LocalObjectRef, replicas *int32, limits *v1.ResourceList) *v1.CapacityBuffer {
	var podTemplateGeneration *int64
	if podTemplateRef != nil {
		podTemplateGeneration = ptr.To[int64](1)
	}
	return testutil.GetBuffer(&testutil.ProvisioningStrategy, nil, nil, podTemplateRef, replicas, podTemplateGeneration, nil, limits)
}

func getTestBufferStatusWithReplicas(podTemplateRef *v1.LocalObjectRef, replicas int32) *v1.CapacityBufferStatus {
	return testutil.GetBufferStatus(podTemplateRef, &replicas, ptr.To[int64](1), &testutil.ProvisioningStrategy, testutil.GetConditionReadyWithMessage("ready"))
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
