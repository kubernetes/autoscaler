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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/autoscaler/cluster-autoscaler/apis/capacitybuffer/autoscaling.x-k8s.io/v1beta1"
	buffersfake "k8s.io/autoscaler/cluster-autoscaler/apis/capacitybuffer/client/clientset/versioned/fake"
	cbclient "k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/client"
	"k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/fakepods"
	"k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/testutil"
	fakeclient "k8s.io/client-go/kubernetes/fake"
	"k8s.io/utils/ptr"
)

const defaultNamespace = "default"

func TestScalableObjectsTranslator(t *testing.T) {
	podTemplate1 := &corev1.PodTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "podTemp1",
			Namespace:  defaultNamespace,
			Generation: 3,
		},
	}
	podTemplate2 := &corev1.PodTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "podTemp2",
			Namespace:  defaultNamespace,
			Generation: 4,
		},
	}
	replicaSet1 := &appsv1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "replicaSet1",
			Namespace: defaultNamespace,
		},
		Spec: appsv1.ReplicaSetSpec{
			Replicas: ptr.To[int32](10),
		},
	}
	replicaSetWithResources := &appsv1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "replicaSetWithResources",
			Namespace: defaultNamespace,
		},
		Spec: appsv1.ReplicaSetSpec{
			Replicas: ptr.To[int32](10),
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									"cpu":    resource.MustParse("100m"),
									"memory": resource.MustParse("4Gi"),
								},
							},
						},
					},
				},
			},
		},
	}
	fakeKubernetesClient := fakeclient.NewSimpleClientset(podTemplate1, podTemplate2, replicaSet1, replicaSetWithResources)
	fakeBuffersClient := buffersfake.NewSimpleClientset()
	fakeCapacityBuffersClient, _ := cbclient.NewCapacityBufferClientFromClients(fakeBuffersClient, fakeKubernetesClient, nil, nil)
	notReadyMsg := "Buffer not ready for provisioning: couldn't get number of replicas for buffer: replicas, percentage and limits are not defined"
	tests := []struct {
		name                   string
		buffers                []*v1.CapacityBuffer
		expectedStatus         []*v1.CapacityBufferStatus
		expectedNumberOfErrors int
	}{
		{
			name: "no scalable ref",
			buffers: []*v1.CapacityBuffer{
				getTestBufferWithScalableAttributes("", nil, nil, nil),
			},
			expectedStatus: []*v1.CapacityBufferStatus{
				testutil.GetBufferStatus(nil, nil, nil, nil, nil),
			},
			expectedNumberOfErrors: 0,
		},
		{
			name: "A buffer referencing replica set with fixed number of replicas",
			buffers: []*v1.CapacityBuffer{
				getTestBufferWithScalableAttributes("buffer1", &v1.ScalableRef{
					Name:     "replicaSet1",
					Kind:     "ReplicaSet",
					APIGroup: "apps",
				}, nil, ptr.To[int32](2)),
			},
			expectedStatus: []*v1.CapacityBufferStatus{
				testutil.GetBufferStatus(&v1.LocalObjectRef{Name: "capacitybuffer-buffer1-pod-template"}, ptr.To[int32](2), ptr.To[int64](0), nil, testutil.GetConditionReady()),
			},
			expectedNumberOfErrors: 0,
		},
		{
			name: "A buffer referencing replica set with percentage 50%",
			buffers: []*v1.CapacityBuffer{
				getTestBufferWithScalableAttributes("buffer1", &v1.ScalableRef{
					Name:     "replicaSet1",
					Kind:     "ReplicaSet",
					APIGroup: "apps",
				}, ptr.To[int32](50), nil),
			},
			expectedStatus: []*v1.CapacityBufferStatus{
				testutil.GetBufferStatus(&v1.LocalObjectRef{Name: "capacitybuffer-buffer1-pod-template"}, ptr.To[int32](5), ptr.To[int64](0), nil, testutil.GetConditionReady()),
			},
			expectedNumberOfErrors: 0,
		},
		{
			name: "A buffer referencing replica set with percentage 200%",
			buffers: []*v1.CapacityBuffer{
				getTestBufferWithScalableAttributes("buffer1", &v1.ScalableRef{
					Name:     "replicaSet1",
					Kind:     "ReplicaSet",
					APIGroup: "apps",
				}, ptr.To[int32](200), nil),
			},
			expectedStatus: []*v1.CapacityBufferStatus{
				testutil.GetBufferStatus(&v1.LocalObjectRef{Name: "capacitybuffer-buffer1-pod-template"}, ptr.To[int32](20), ptr.To[int64](0), nil, testutil.GetConditionReady()),
			},
			expectedNumberOfErrors: 0,
		},
		{
			name: "A buffer referencing replica set with percentage 200% and replicas 15",
			buffers: []*v1.CapacityBuffer{
				getTestBufferWithScalableAttributes("buffer1", &v1.ScalableRef{
					Name:     "replicaSet1",
					Kind:     "ReplicaSet",
					APIGroup: "apps",
				}, ptr.To[int32](200), ptr.To[int32](15)),
			},
			expectedStatus: []*v1.CapacityBufferStatus{
				testutil.GetBufferStatus(&v1.LocalObjectRef{Name: "capacitybuffer-buffer1-pod-template"}, ptr.To[int32](15), ptr.To[int64](0), nil, testutil.GetConditionReady()),
			},
			expectedNumberOfErrors: 0,
		},
		{
			name: "A buffer referencing valid replica set with nil percentage and replicas",
			buffers: []*v1.CapacityBuffer{
				getTestBufferWithScalableAttributes("buffer1", &v1.ScalableRef{
					Name:     "replicaSet1",
					Kind:     "ReplicaSet",
					APIGroup: "apps",
				}, nil, nil),
			},
			expectedStatus: []*v1.CapacityBufferStatus{
				testutil.GetBufferStatus(
					&v1.LocalObjectRef{Name: "capacitybuffer-buffer1-pod-template"}, nil, ptr.To[int64](0), nil, testutil.GetConditionNotReadyWithMessage(notReadyMsg),
				),
			},
			expectedNumberOfErrors: 0,
		},
		{
			name: "A buffer referencing replica set with percentage 50% and limits",
			buffers: []*v1.CapacityBuffer{
				getTestBufferWithScalableAttributesAndLimits("buffer1", &v1.ScalableRef{
					Name:     "replicaSetWithResources",
					Kind:     "ReplicaSet",
					APIGroup: "apps",
				}, ptr.To[int32](50), nil, &v1.ResourceList{
					"memory": resource.MustParse("9Gi"),
					"cpu":    resource.MustParse("200m"),
				}),
			},
			expectedStatus: []*v1.CapacityBufferStatus{
				testutil.GetBufferStatus(&v1.LocalObjectRef{Name: "capacitybuffer-buffer1-pod-template"}, ptr.To[int32](2), ptr.To[int64](0), nil, testutil.GetConditionReady()),
			},
			expectedNumberOfErrors: 0,
		},
		{
			name: "Limits exist and no replicas",
			buffers: []*v1.CapacityBuffer{
				getTestBufferWithScalableAttributesAndLimits("buffer1", &v1.ScalableRef{
					Name:     "replicaSetWithResources",
					Kind:     "ReplicaSet",
					APIGroup: "apps",
				}, nil, nil, &v1.ResourceList{
					"memory": resource.MustParse("9Gi"),
					"cpu":    resource.MustParse("200m"),
				}),
			},
			expectedStatus: []*v1.CapacityBufferStatus{
				testutil.GetBufferStatus(&v1.LocalObjectRef{Name: "capacitybuffer-buffer1-pod-template"}, ptr.To[int32](2), ptr.To[int64](0), nil, testutil.GetConditionReady()),
			},
			expectedNumberOfErrors: 0,
		},
		{
			name: "Limits exist with smaller replicas",
			buffers: []*v1.CapacityBuffer{
				getTestBufferWithScalableAttributesAndLimits("buffer1", &v1.ScalableRef{
					Name:     "replicaSetWithResources",
					Kind:     "ReplicaSet",
					APIGroup: "apps",
				}, nil, ptr.To[int32](1), &v1.ResourceList{
					"memory": resource.MustParse("9Gi"),
					"cpu":    resource.MustParse("200m"),
				}),
			},
			expectedStatus: []*v1.CapacityBufferStatus{
				testutil.GetBufferStatus(&v1.LocalObjectRef{Name: "capacitybuffer-buffer1-pod-template"}, ptr.To[int32](1), ptr.To[int64](0), nil, testutil.GetConditionReady()),
			},
			expectedNumberOfErrors: 0,
		},
		{
			name: "Limits exist with bigger replicas",
			buffers: []*v1.CapacityBuffer{
				getTestBufferWithScalableAttributesAndLimits("buffer1", &v1.ScalableRef{
					Name:     "replicaSetWithResources",
					Kind:     "ReplicaSet",
					APIGroup: "apps",
				}, nil, ptr.To[int32](5), &v1.ResourceList{
					"memory": resource.MustParse("9Gi"),
					"cpu":    resource.MustParse("200m"),
				}),
			},
			expectedStatus: []*v1.CapacityBufferStatus{
				testutil.GetBufferStatus(&v1.LocalObjectRef{Name: "capacitybuffer-buffer1-pod-template"}, ptr.To[int32](2), ptr.To[int64](0), nil, testutil.GetConditionReady()),
			},
			expectedNumberOfErrors: 0,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resolver := fakepods.NewDefaultingResolver()
			translator := NewDefaultScalableObjectsTranslator(fakeCapacityBuffersClient, resolver)
			errors := translator.Translate(test.buffers)
			assert.Equal(t, test.expectedNumberOfErrors, len(errors))
			for i, buffer := range test.buffers {
				wantStatus := test.expectedStatus[i]
				if diff := cmp.Diff(wantStatus, &buffer.Status, cmpopts.IgnoreFields(metav1.Condition{}, "LastTransitionTime")); diff != "" {
					t.Errorf("buffer status mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func getTestBufferWithScalableAttributesAndLimits(bufferName string, scalableRef *v1.ScalableRef, percentage *int32, replicas *int32, limits *v1.ResourceList) *v1.CapacityBuffer {
	buffer := getTestBufferWithScalableAttributes(bufferName, scalableRef, percentage, replicas)
	buffer.Spec.Limits = limits
	return buffer
}

func getTestBufferWithScalableAttributes(bufferName string, scalableRef *v1.ScalableRef, percentage *int32, replicas *int32) *v1.CapacityBuffer {
	buffer := &v1.CapacityBuffer{}
	buffer.Name = bufferName
	buffer.Namespace = defaultNamespace
	buffer.Spec.ScalableRef = scalableRef
	buffer.Spec.Percentage = percentage
	buffer.Spec.Replicas = replicas
	return buffer
}
