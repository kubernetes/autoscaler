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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	v1 "k8s.io/autoscaler/cluster-autoscaler/apis/capacitybuffer/autoscaling.x-k8s.io/v1alpha1"
	cbclient "k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/client"
	"k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/common"
	"k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/testutil"
	fakeClient "k8s.io/client-go/kubernetes/fake"
)

func TestResourceQuotasTranslator(t *testing.T) {
	podTemp := &corev1.PodTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "podTemp",
			Namespace: "default",
		},
		Template: corev1.PodTemplateSpec{
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								"cpu":    resource.MustParse("1"),
								"memory": resource.MustParse("1Gi"),
							},
						},
					},
				},
			},
		},
	}

	rqBase := &corev1.ResourceQuota{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "quota",
			Namespace: "default",
			UID:       types.UID("quota-uid"),
		},
		Status: corev1.ResourceQuotaStatus{
			Hard: corev1.ResourceList{
				"cpu":  resource.MustParse("10"),
				"pods": resource.MustParse("10"),
			},
			Used: corev1.ResourceList{
				"cpu":  resource.MustParse("0"),
				"pods": resource.MustParse("0"),
			},
		},
	}

	tests := []struct {
		name             string
		quotas           []*corev1.ResourceQuota
		buffers          []*v1.CapacityBuffer
		expectedReplicas []int32
		expectedLimited  []bool
	}{
		{
			name:   "single buffer fits within quota",
			quotas: []*corev1.ResourceQuota{rqBase},
			buffers: []*v1.CapacityBuffer{
				testutil.NewBuffer(
					testutil.WithStatusPodTemplateRef("podTemp"),
					testutil.WithStatusReplicas(5),
				),
			},
			expectedReplicas: []int32{5},
			expectedLimited:  []bool{false},
		},
		{
			name:   "multiple buffers, second limited by quota",
			quotas: []*corev1.ResourceQuota{rqBase},
			buffers: []*v1.CapacityBuffer{
				testutil.NewBuffer(
					testutil.WithStatusPodTemplateRef("podTemp"),
					testutil.WithStatusReplicas(5),
				),
				testutil.NewBuffer(
					testutil.WithStatusPodTemplateRef("podTemp"),
					testutil.WithStatusReplicas(10),
				),
			},
			expectedReplicas: []int32{5, 5}, // 10 total capacity - 5 used by buffer1 = 5 available for buffer2
			expectedLimited:  []bool{false, true},
		},
		{
			name: "buffer limited by existing usage",
			quotas: func() []*corev1.ResourceQuota {
				rq := rqBase.DeepCopy()
				rq.Status.Used["cpu"] = resource.MustParse("8")
				return []*corev1.ResourceQuota{rq}
			}(),
			buffers: []*v1.CapacityBuffer{
				testutil.NewBuffer(
					testutil.WithStatusPodTemplateRef("podTemp"),
					testutil.WithStatusReplicas(5),
				),
			},
			expectedReplicas: []int32{2}, // 10 total - 8 used = 2 available
			expectedLimited:  []bool{true},
		},
		{
			name: "buffer limited by pods quota",
			quotas: []*corev1.ResourceQuota{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "quota-pods",
						Namespace: "default",
						UID:       types.UID("quota-pods-uid"),
					},
					Status: corev1.ResourceQuotaStatus{
						Hard: corev1.ResourceList{
							"pods": resource.MustParse("5"),
						},
						Used: corev1.ResourceList{
							"pods": resource.MustParse("0"),
						},
					},
				},
			},
			buffers: []*v1.CapacityBuffer{
				testutil.NewBuffer(
					testutil.WithStatusPodTemplateRef("podTemp"),
					testutil.WithStatusReplicas(3),
				),
				testutil.NewBuffer(
					testutil.WithStatusPodTemplateRef("podTemp"),
					testutil.WithStatusReplicas(5),
				),
			},
			expectedReplicas: []int32{3, 2}, // 5 total - 3 used = 2 available
			expectedLimited:  []bool{false, true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			objs := []runtime.Object{podTemp}
			for _, q := range tt.quotas {
				objs = append(objs, q)
			}
			fakeK8s := fakeClient.NewSimpleClientset(objs...)
			client, _ := cbclient.NewCapacityBufferClient(nil, fakeK8s, nil, nil, nil, nil, nil, nil, nil, nil, nil)
			translator := NewResourceQuotasTranslator(client)

			// Assign namespace and names to buffers
			for i, buffer := range tt.buffers {
				buffer.Namespace = "default"
				if buffer.Name == "" {
					buffer.Name = "buffer-" + string(rune(i))
				}
			}

			errs := translator.Translate(tt.buffers)
			assert.Empty(t, errs)

			for i, buffer := range tt.buffers {
				assert.Equal(t, tt.expectedReplicas[i], *buffer.Status.Replicas, "Buffer %d replicas mismatch", i)

				hasLimitedCondition := false
				for _, c := range buffer.Status.Conditions {
					if c.Type == common.LimitedByQuotasCondition {
						if c.Status == metav1.ConditionTrue {
							hasLimitedCondition = true
						}
					}
				}
				assert.Equal(t, tt.expectedLimited[i], hasLimitedCondition, "Buffer %d LimitedByQuotas condition mismatch", i)
			}
		})
	}
}
