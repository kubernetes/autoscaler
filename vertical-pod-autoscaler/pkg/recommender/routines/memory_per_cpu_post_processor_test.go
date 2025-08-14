/*
Copyright 2022 The Kubernetes Authors.

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

package routines

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/test"
)

func TestMemoryPerCPUPostProcessor_Process(t *testing.T) {
	const Gi = int64(1024 * 1024 * 1024)

	tests := []struct {
		name           string
		vpa            *vpa_types.VerticalPodAutoscaler
		recommendation *vpa_types.RecommendedPodResources
		want           *vpa_types.RecommendedPodResources
	}{
		{
			name: "No policy defined - no change",
			vpa:  &vpa_types.VerticalPodAutoscaler{},
			recommendation: &vpa_types.RecommendedPodResources{
				ContainerRecommendations: []vpa_types.RecommendedContainerResources{
					test.Recommendation().WithContainer("c1").WithTarget("1", "4Gi").GetContainerResources(),
				},
			},
			want: &vpa_types.RecommendedPodResources{
				ContainerRecommendations: []vpa_types.RecommendedContainerResources{
					test.Recommendation().WithContainer("c1").WithTarget("1", "4Gi").GetContainerResources(),
				},
			},
		},
		{
			name: "Policy matches - too much RAM -> increase CPU",
			vpa: &vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					ResourcePolicy: &vpa_types.PodResourcePolicy{
						ContainerPolicies: []vpa_types.ContainerResourcePolicy{
							{
								ContainerName: "c1",
								MemoryPerCPU:  resource.NewQuantity(4*Gi, resource.BinarySI), // 1 core -> 4Gi
							},
						},
					},
				},
			},
			recommendation: &vpa_types.RecommendedPodResources{
				ContainerRecommendations: []vpa_types.RecommendedContainerResources{
					test.Recommendation().WithContainer("c1").WithTarget("1", "8Gi").GetContainerResources(),
				},
			},
			want: &vpa_types.RecommendedPodResources{
				ContainerRecommendations: []vpa_types.RecommendedContainerResources{
					test.Recommendation().WithContainer("c1").WithTarget("2", "8Gi").GetContainerResources(),
				},
			},
		},
		{
			name: "Policy matches - not enough RAM -> increase Memory",
			vpa: &vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					ResourcePolicy: &vpa_types.PodResourcePolicy{
						ContainerPolicies: []vpa_types.ContainerResourcePolicy{
							{
								ContainerName: "c1",
								MemoryPerCPU:  resource.NewQuantity(4*Gi, resource.BinarySI),
							},
						},
					},
				},
			},
			recommendation: &vpa_types.RecommendedPodResources{
				ContainerRecommendations: []vpa_types.RecommendedContainerResources{
					test.Recommendation().WithContainer("c1").WithTarget("4", "8Gi").GetContainerResources(),
				},
			},
			want: &vpa_types.RecommendedPodResources{
				ContainerRecommendations: []vpa_types.RecommendedContainerResources{
					test.Recommendation().WithContainer("c1").WithTarget("4", "16Gi").GetContainerResources(),
				},
			},
		},
		{
			name: "Missing CPU or Memory - no change",
			vpa: &vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					ResourcePolicy: &vpa_types.PodResourcePolicy{
						ContainerPolicies: []vpa_types.ContainerResourcePolicy{
							{
								ContainerName: "c1",
								MemoryPerCPU:  resource.NewQuantity(4*Gi, resource.BinarySI),
							},
						},
					},
				},
			},
			recommendation: &vpa_types.RecommendedPodResources{
				ContainerRecommendations: []vpa_types.RecommendedContainerResources{
					{
						ContainerName: "c1",
						Target: v1.ResourceList{
							v1.ResourceCPU: *resource.NewMilliQuantity(1000, resource.DecimalSI),
						},
					},
				},
			},
			want: &vpa_types.RecommendedPodResources{
				ContainerRecommendations: []vpa_types.RecommendedContainerResources{
					{
						ContainerName: "c1",
						Target: v1.ResourceList{
							v1.ResourceCPU: *resource.NewMilliQuantity(1000, resource.DecimalSI),
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := MemoryPerCPUPostProcessor{}
			got := c.Process(tt.vpa, tt.recommendation)
			assert.True(t, equalRecommendedPodResources(tt.want, got), "Process(%v, %v)", tt.vpa, tt.recommendation)
		})
	}
}
