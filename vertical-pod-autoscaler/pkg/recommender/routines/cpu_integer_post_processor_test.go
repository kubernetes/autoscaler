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
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/test"
	"testing"
)

func Test_extractContainerName(t *testing.T) {
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

func Test_integerCPUPostProcessor_Process(t *testing.T) {
	tests := []struct {
		name           string
		vpa            *model.Vpa
		recommendation *vpa_types.RecommendedPodResources
		want           *vpa_types.RecommendedPodResources
	}{
		{
			name: "No containers match",
			vpa: &model.Vpa{Annotations: map[string]string{
				vpaPostProcessorPrefix + "container-other" + vpaPostProcessorIntegerCPUSuffix: "true",
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
				vpaPostProcessorPrefix + "container1" + vpaPostProcessorIntegerCPUSuffix: "true",
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
				vpaPostProcessorPrefix + "container1" + vpaPostProcessorIntegerCPUSuffix: "true",
				vpaPostProcessorPrefix + "container2" + vpaPostProcessorIntegerCPUSuffix: "true",
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
			assert.Equalf(t, reScale3(tt.want), reScale3(got), "Process(%v, %v, nil)", tt.vpa, tt.recommendation)
		})
	}
}

func reScale3(recommended *vpa_types.RecommendedPodResources) *vpa_types.RecommendedPodResources {

	scale3 := func(rl v1.ResourceList) {
		for k, v := range rl {
			v.SetMilli(v.MilliValue())
			rl[k] = v
		}
	}

	for _, r := range recommended.ContainerRecommendations {
		scale3(r.LowerBound)
		scale3(r.Target)
		scale3(r.UncappedTarget)
		scale3(r.UpperBound)
	}
	return recommended
}
