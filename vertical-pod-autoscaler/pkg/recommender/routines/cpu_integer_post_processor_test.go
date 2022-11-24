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
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/test"
	"testing"
)

func TestExtractContainerName(t *testing.T) {
	tests := []struct {
		name   string
		key    string
		prefix string
		suffix string
		want   string
	}{
		{
			name:   "empty",
			key:    "",
			prefix: "",
			suffix: "",
			want:   "",
		},
		{
			name:   "no match",
			key:    "abc",
			prefix: "z",
			suffix: "x",
			want:   "",
		},
		{
			name:   "match",
			key:    "abc",
			prefix: "a",
			suffix: "c",
			want:   "b",
		},
		{
			name:   "real",
			key:    vpaPostProcessorPrefix + "kafka" + vpaPostProcessorIntegerCPUSuffix,
			prefix: vpaPostProcessorPrefix,
			suffix: vpaPostProcessorIntegerCPUSuffix,
			want:   "kafka",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, extractContainerName(tt.key, tt.prefix, tt.suffix), "extractContainerName(%v, %v, %v)", tt.key, tt.prefix, tt.suffix)
		})
	}
}

func TestIntegerCPUPostProcessor_Process(t *testing.T) {
	tests := []struct {
		name           string
		vpa            *model.Vpa
		recommendation *vpa_types.RecommendedPodResources
		want           *vpa_types.RecommendedPodResources
	}{
		{
			name: "No containers match",
			vpa: &model.Vpa{Annotations: map[string]string{
				vpaPostProcessorPrefix + "container-other" + vpaPostProcessorIntegerCPUSuffix: vpaPostProcessorIntegerCPUValue,
			}},
			recommendation: &vpa_types.RecommendedPodResources{
				ContainerRecommendations: []vpa_types.RecommendedContainerResources{
					test.Recommendation().WithContainer("container1").WithTarget("8.6", "200Mi").GetContainerResources(),
					test.Recommendation().WithContainer("container2").WithTarget("8.2", "300Mi").GetContainerResources(),
				},
			},
			want: &vpa_types.RecommendedPodResources{
				ContainerRecommendations: []vpa_types.RecommendedContainerResources{
					test.Recommendation().WithContainer("container1").WithTarget("8.6", "200Mi").GetContainerResources(),
					test.Recommendation().WithContainer("container2").WithTarget("8.2", "300Mi").GetContainerResources(),
				},
			},
		},
		{
			name: "2 containers, 1 matching only",
			vpa: &model.Vpa{Annotations: map[string]string{
				vpaPostProcessorPrefix + "container1" + vpaPostProcessorIntegerCPUSuffix: vpaPostProcessorIntegerCPUValue,
			}},
			recommendation: &vpa_types.RecommendedPodResources{
				ContainerRecommendations: []vpa_types.RecommendedContainerResources{
					test.Recommendation().WithContainer("container1").WithTarget("8.6", "200Mi").GetContainerResources(),
					test.Recommendation().WithContainer("container2").WithTarget("8.2", "300Mi").GetContainerResources(),
				},
			},
			want: &vpa_types.RecommendedPodResources{
				ContainerRecommendations: []vpa_types.RecommendedContainerResources{
					test.Recommendation().WithContainer("container1").WithTarget("9", "200Mi").GetContainerResources(),
					test.Recommendation().WithContainer("container2").WithTarget("8.2", "300Mi").GetContainerResources(),
				},
			},
		},
		{
			name: "2 containers, 2 matching",
			vpa: &model.Vpa{Annotations: map[string]string{
				vpaPostProcessorPrefix + "container1" + vpaPostProcessorIntegerCPUSuffix: vpaPostProcessorIntegerCPUValue,
				vpaPostProcessorPrefix + "container2" + vpaPostProcessorIntegerCPUSuffix: vpaPostProcessorIntegerCPUValue,
			}},
			recommendation: &vpa_types.RecommendedPodResources{
				ContainerRecommendations: []vpa_types.RecommendedContainerResources{
					test.Recommendation().WithContainer("container1").WithTarget("8.6", "200Mi").GetContainerResources(),
					test.Recommendation().WithContainer("container2").WithTarget("5.2", "300Mi").GetContainerResources(),
				},
			},
			want: &vpa_types.RecommendedPodResources{
				ContainerRecommendations: []vpa_types.RecommendedContainerResources{
					test.Recommendation().WithContainer("container1").WithTarget("9", "200Mi").GetContainerResources(),
					test.Recommendation().WithContainer("container2").WithTarget("6", "300Mi").GetContainerResources(),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := IntegerCPUPostProcessor{}
			got := c.Process(tt.vpa, tt.recommendation, nil)
			assert.True(t, equalRecommendedPodResources(tt.want, got), "Process(%v, %v, nil)", tt.vpa, tt.recommendation)
		})
	}
}

func equalRecommendedPodResources(a, b *vpa_types.RecommendedPodResources) bool {
	if len(a.ContainerRecommendations) != len(b.ContainerRecommendations) {
		return false
	}

	for i := range a.ContainerRecommendations {
		if !equalResourceList(a.ContainerRecommendations[i].LowerBound, b.ContainerRecommendations[i].LowerBound) {
			return false
		}
		if !equalResourceList(a.ContainerRecommendations[i].Target, b.ContainerRecommendations[i].Target) {
			return false
		}
		if !equalResourceList(a.ContainerRecommendations[i].UncappedTarget, b.ContainerRecommendations[i].UncappedTarget) {
			return false
		}
		if !equalResourceList(a.ContainerRecommendations[i].UpperBound, b.ContainerRecommendations[i].UpperBound) {
			return false
		}
	}
	return true
}

func equalResourceList(rla, rlb v1.ResourceList) bool {
	if len(rla) != len(rlb) {
		return false
	}
	for k := range rla {
		q := rla[k]
		if q.Cmp(rlb[k]) != 0 {
			return false
		}
	}
	for k := range rlb {
		q := rlb[k]
		if q.Cmp(rla[k]) != 0 {
			return false
		}
	}
	return true
}

func TestSetIntegerCPURecommendation(t *testing.T) {
	tests := []struct {
		name                   string
		recommendation         v1.ResourceList
		expectedRecommendation v1.ResourceList
	}{
		{
			name: "unchanged",
			recommendation: map[v1.ResourceName]resource.Quantity{
				v1.ResourceCPU:    resource.MustParse("8"),
				v1.ResourceMemory: resource.MustParse("6Gi"),
			},
			expectedRecommendation: map[v1.ResourceName]resource.Quantity{
				v1.ResourceCPU:    resource.MustParse("8"),
				v1.ResourceMemory: resource.MustParse("6Gi"),
			},
		},
		{
			name: "round up from 0.1",
			recommendation: map[v1.ResourceName]resource.Quantity{
				v1.ResourceCPU:    resource.MustParse("8.1"),
				v1.ResourceMemory: resource.MustParse("6Gi"),
			},
			expectedRecommendation: map[v1.ResourceName]resource.Quantity{
				v1.ResourceCPU:    resource.MustParse("9"),
				v1.ResourceMemory: resource.MustParse("6Gi"),
			},
		},
		{
			name: "round up from 0.9",
			recommendation: map[v1.ResourceName]resource.Quantity{
				v1.ResourceCPU:    resource.MustParse("8.9"),
				v1.ResourceMemory: resource.MustParse("6Gi"),
			},
			expectedRecommendation: map[v1.ResourceName]resource.Quantity{
				v1.ResourceCPU:    resource.MustParse("9"),
				v1.ResourceMemory: resource.MustParse("6Gi"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setIntegerCPURecommendation(tt.recommendation)
			assert.True(t, equalResourceList(tt.recommendation, tt.expectedRecommendation))
		})
	}
}
