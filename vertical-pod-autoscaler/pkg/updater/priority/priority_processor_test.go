/*
Copyright 2020 The Kubernetes Authors.

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

package priority

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/annotations"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/test"
)

func TestGetUpdatePriority(t *testing.T) {
	containerName := "test-container"
	containerName2 := "test-container2"
	testCases := []struct {
		name         string
		pod          *corev1.Pod
		vpa          *vpa_types.VerticalPodAutoscaler
		expectedPrio PodPriority
	}{
		{
			name: "simple scale up",
			pod:  test.Pod().WithName("POD1").AddContainer(test.Container().WithName(containerName).WithCPURequest(resource.MustParse("2")).Get()).Get(),
			vpa:  test.VerticalPodAutoscaler().WithContainer(containerName).WithTarget("10", "").Get(),
			expectedPrio: PodPriority{
				OutsideRecommendedRange: false,
				ResourceDiff:            4.0,
				ScaleUp:                 true,
			},
		}, {
			name: "simple scale up, resources from containerStatus have higher priority",
			pod: test.Pod().WithName("POD1").AddContainer(test.Container().WithName(containerName).WithCPURequest(resource.MustParse("10000")).Get()).
				AddContainerStatus(test.ContainerStatus().WithName(containerName).WithCPURequest(resource.MustParse("2")).Get()).Get(),
			vpa: test.VerticalPodAutoscaler().WithContainer(containerName).WithTarget("10", "").Get(),
			expectedPrio: PodPriority{
				OutsideRecommendedRange: false,
				ResourceDiff:            4.0,
				ScaleUp:                 true,
			},
		}, {
			name: "simple scale down",
			pod:  test.Pod().WithName("POD1").AddContainer(test.Container().WithName(containerName).WithCPURequest(resource.MustParse("4")).Get()).Get(),
			vpa:  test.VerticalPodAutoscaler().WithContainer(containerName).WithTarget("2", "").Get(),
			expectedPrio: PodPriority{
				OutsideRecommendedRange: false,
				ResourceDiff:            0.5,
				ScaleUp:                 false,
			},
		}, {
			name: "no resource diff",
			pod:  test.Pod().WithName("POD1").AddContainer(test.Container().WithName(containerName).WithCPURequest(resource.MustParse("2")).Get()).Get(),
			vpa:  test.VerticalPodAutoscaler().WithContainer(containerName).WithTarget("2", "").Get(),
			expectedPrio: PodPriority{
				OutsideRecommendedRange: false,
				ResourceDiff:            0.0,
				ScaleUp:                 false,
			},
		}, {
			name: "scale up on milliquanitites",
			pod:  test.Pod().WithName("POD1").AddContainer(test.Container().WithName(containerName).WithCPURequest(resource.MustParse("10m")).Get()).Get(),
			vpa:  test.VerticalPodAutoscaler().WithContainer(containerName).WithTarget("900m", "").Get(),
			expectedPrio: PodPriority{
				OutsideRecommendedRange: false,
				ResourceDiff:            89.0,
				ScaleUp:                 true,
			},
		}, {
			name: "scale up outside recommended range",
			pod:  test.Pod().WithName("POD1").AddContainer(test.Container().WithName(containerName).WithCPURequest(resource.MustParse("4")).Get()).Get(),
			vpa: test.VerticalPodAutoscaler().WithContainer(containerName).
				WithTarget("10", "").
				WithLowerBound("6", "").
				WithUpperBound("14", "").Get(),
			expectedPrio: PodPriority{
				OutsideRecommendedRange: true,
				ResourceDiff:            1.5,
				ScaleUp:                 true,
			},
		}, {
			name: "scale down outside recommended range",
			pod:  test.Pod().WithName("POD1").AddContainer(test.Container().WithName(containerName).WithCPURequest(resource.MustParse("8")).Get()).Get(),
			vpa: test.VerticalPodAutoscaler().WithContainer(containerName).
				WithTarget("2", "").
				WithLowerBound("1", "").
				WithUpperBound("3", "").Get(),
			expectedPrio: PodPriority{
				OutsideRecommendedRange: true,
				ResourceDiff:            0.75,
				ScaleUp:                 false,
			},
		}, {
			name: "scale up with multiple quantities",
			pod:  test.Pod().WithName("POD1").AddContainer(test.Container().WithName(containerName).WithCPURequest(resource.MustParse("2")).Get()).Get(),
			vpa:  test.VerticalPodAutoscaler().WithContainer(containerName).WithTarget("10", "").Get(),
			expectedPrio: PodPriority{
				OutsideRecommendedRange: false,
				ResourceDiff:            4.0,
				ScaleUp:                 true,
			},
		}, {
			name: "multiple resources, both scale up",
			pod:  test.Pod().WithName("POD1").AddContainer(test.Container().WithName(containerName).WithCPURequest(resource.MustParse("3")).WithMemRequest(resource.MustParse("10M")).Get()).Get(),
			vpa:  test.VerticalPodAutoscaler().WithContainer(containerName).WithTarget("6", "20M").Get(),
			expectedPrio: PodPriority{
				OutsideRecommendedRange: false,
				ResourceDiff:            1.0 + 1.0, // summed relative diffs for resources
				ScaleUp:                 true,
			},
		}, {
			name: "multiple resources, only one scale up",
			pod:  test.Pod().WithName("POD1").AddContainer(test.Container().WithName(containerName).WithCPURequest(resource.MustParse("4")).WithMemRequest(resource.MustParse("10M")).Get()).Get(),
			vpa:  test.VerticalPodAutoscaler().WithContainer(containerName).WithTarget("2", "20M").Get(),
			expectedPrio: PodPriority{
				OutsideRecommendedRange: false,
				ResourceDiff:            0.5 + 1, // summed relative diffs for resources
				ScaleUp:                 true,
			},
		}, {
			name: "multiple resources, both scale down",
			pod:  test.Pod().WithName("POD1").AddContainer(test.Container().WithName(containerName).WithCPURequest(resource.MustParse("4")).WithMemRequest(resource.MustParse("20M")).Get()).Get(),
			vpa:  test.VerticalPodAutoscaler().WithContainer(containerName).WithTarget("2", "10M").Get(),
			expectedPrio: PodPriority{
				OutsideRecommendedRange: false,
				ResourceDiff:            0.5 + 0.5, // summed relative diffs for resources
				ScaleUp:                 false,
			},
		}, {
			name: "multiple resources, one outside recommended range",
			pod:  test.Pod().WithName("POD1").AddContainer(test.Container().WithName(containerName).WithCPURequest(resource.MustParse("4")).WithMemRequest(resource.MustParse("20M")).Get()).Get(),
			vpa: test.VerticalPodAutoscaler().WithContainer(containerName).
				WithTarget("2", "10M").
				WithLowerBound("1", "5M").
				WithUpperBound("3", "30M").Get(),
			expectedPrio: PodPriority{
				OutsideRecommendedRange: true,
				ResourceDiff:            0.5 + 0.5, // summed relative diffs for resources
				ScaleUp:                 false,
			},
		}, {
			name: "multiple containers, both scale up",
			pod: test.Pod().WithName("POD1").AddContainer(test.Container().WithName(containerName).WithCPURequest(resource.MustParse("1")).Get()).
				AddContainer(test.Container().WithName("test-container-2").WithCPURequest(resource.MustParse("2")).Get()).Get(),
			vpa: test.VerticalPodAutoscaler().WithContainer(containerName).
				WithTarget("4", "").AppendRecommendation(
				test.Recommendation().
					WithContainer("test-container-2").
					WithTarget("8", "").GetContainerResources()).Get(),
			expectedPrio: PodPriority{
				OutsideRecommendedRange: false,
				ResourceDiff:            3.0, // relative diff between summed requests and summed recommendations
				ScaleUp:                 true,
			},
		}, {
			name: "multiple containers, both scale down",
			pod: test.Pod().WithName("POD1").AddContainer(test.Container().WithName(containerName).WithCPURequest(resource.MustParse("3")).Get()).
				AddContainer(test.Container().WithName("test-container-2").WithCPURequest(resource.MustParse("7")).Get()).Get(),
			vpa: test.VerticalPodAutoscaler().WithContainer(containerName).
				WithTarget("1", "").AppendRecommendation(
				test.Recommendation().
					WithContainer("test-container-2").
					WithTarget("2", "").GetContainerResources()).Get(),
			expectedPrio: PodPriority{
				OutsideRecommendedRange: false,
				ResourceDiff:            0.7, // relative diff between summed requests and summed recommendations
				ScaleUp:                 false,
			},
		}, {
			name: "multiple containers, both scale up, one outside range",
			pod: test.Pod().WithName("POD1").AddContainer(test.Container().WithName(containerName).WithCPURequest(resource.MustParse("1")).Get()).
				AddContainer(test.Container().WithName("test-container-2").WithCPURequest(resource.MustParse("2")).Get()).Get(),
			vpa: test.VerticalPodAutoscaler().WithContainer(containerName).
				WithTarget("4", "").
				WithLowerBound("1", "").AppendRecommendation(
				test.Recommendation().
					WithContainer("test-container-2").
					WithTarget("8", "").
					WithLowerBound("3", "").
					WithUpperBound("10", "").GetContainerResources()).Get(),
			expectedPrio: PodPriority{
				OutsideRecommendedRange: true,
				ResourceDiff:            3.0, // relative diff between summed requests and summed recommendations
				ScaleUp:                 true,
			},
		}, {
			name: "multiple containers, multiple resources",
			//   container1: request={6 CPU, 10 MB}, recommended={8 CPU, 20 MB}
			//   container2: request={4 CPU, 30 MB}, recommended={7 CPU, 30 MB}
			//   total:      request={10 CPU, 40 MB}, recommended={15 CPU, 50 MB}
			pod: test.Pod().WithName("POD1").AddContainer(test.Container().WithName(containerName).WithCPURequest(resource.MustParse("6")).WithMemRequest(resource.MustParse("10M")).Get()).
				AddContainer(test.Container().WithName("test-container-2").WithCPURequest(resource.MustParse("4")).WithMemRequest(resource.MustParse("30M")).Get()).Get(),
			vpa: test.VerticalPodAutoscaler().WithContainer(containerName).
				WithTarget("8", "20M").AppendRecommendation(
				test.Recommendation().
					WithContainer("test-container-2").
					WithTarget("7", "30M").GetContainerResources()).Get(),
			expectedPrio: PodPriority{
				OutsideRecommendedRange: false,
				// relative diff between summed requests and summed recommendations, summed over resources
				ResourceDiff: 0.5 + 0.25,
				ScaleUp:      true,
			},
		},
		{
			name: "VPA with an empty status stanza",
			pod: test.Pod().
				WithName("POD1").
				Get(),
			vpa: &vpa_types.VerticalPodAutoscaler{
				Status: vpa_types.VerticalPodAutoscalerStatus{},
			},
			expectedPrio: PodPriority{
				OutsideRecommendedRange: false,
				ResourceDiff:            0,
				ScaleUp:                 false,
			},
		},
		{
			name: "pod level simple scale up",
			pod: test.Pod().
				WithName("POD1").
				AddContainer(
					test.Container().
						WithName(containerName).
						Get()).
				WithCPURequest(resource.MustParse("2")).
				Get(),
			vpa: test.VerticalPodAutoscaler().
				WithPodLevelTarget("10", "").
				Get(),
			expectedPrio: PodPriority{
				OutsideRecommendedRange: false,
				ResourceDiff:            4.0, // Abs(2-10)/2
				ScaleUp:                 true,
			},
		},
		{
			name: "pod level simple scale down",
			pod: test.Pod().
				WithName("POD1").
				AddContainer(
					test.Container().
						WithName(containerName).
						Get()).
				WithCPURequest(resource.MustParse("4")).
				Get(),
			vpa: test.VerticalPodAutoscaler().
				WithPodLevelTarget("2", "").
				Get(),
			expectedPrio: PodPriority{
				OutsideRecommendedRange: false,
				ResourceDiff:            0.5, // Abs(4-2)/4
				ScaleUp:                 false,
			},
		},
		{
			name: "pod level no resource diff",
			pod: test.Pod().
				WithName("POD1").
				AddContainer(
					test.Container().
						WithName(containerName).
						Get()).
				WithCPURequest(resource.MustParse("2")).
				Get(),
			vpa: test.VerticalPodAutoscaler().
				WithPodLevelTarget("2", "").
				Get(),
			expectedPrio: PodPriority{
				OutsideRecommendedRange: false,
				ResourceDiff:            0, // Abs(2-2)/2
				ScaleUp:                 false,
			},
		},
		{
			name: "pod level no resource diff at either container or pod level",
			pod: test.Pod().
				WithName("POD1").
				AddContainer(
					test.Container().
						WithName(containerName).
						WithCPURequest(resource.MustParse("1")).
						Get()).
				WithCPURequest(resource.MustParse("2")).
				Get(),
			vpa: test.VerticalPodAutoscaler().
				WithContainer(containerName).
				AppendRecommendation(
					test.Recommendation().
						WithContainer(containerName).
						WithTarget("1", "").
						GetContainerResources()).
				WithPodLevelTarget("2", "").
				Get(),
			expectedPrio: PodPriority{
				OutsideRecommendedRange: false,
				ResourceDiff:            0, // Abs(1-1)/1 + Abs(2-2)/2
				ScaleUp:                 false,
			},
		},
		{
			name: "pod level scale up outside recommended range",
			pod: test.Pod().
				WithName("POD1").
				AddContainer(
					test.Container().
						WithName(containerName).
						Get()).
				WithCPURequest(resource.MustParse("2")).
				Get(),
			vpa: test.VerticalPodAutoscaler().
				WithPodLevelLowerBound("3", "").
				WithPodLevelTarget("4", "").
				WithPodLevelUpperBound("5", "").
				Get(),
			expectedPrio: PodPriority{
				OutsideRecommendedRange: true,
				ResourceDiff:            1, // Abs(2-4)/2
				ScaleUp:                 true,
			},
		},
		{
			name: "pod level scale down outside recommended range",
			pod: test.Pod().
				WithName("POD1").
				AddContainer(
					test.Container().
						WithName(containerName).
						Get()).
				WithCPURequest(resource.MustParse("8")).
				Get(),
			vpa: test.VerticalPodAutoscaler().
				WithPodLevelLowerBound("1", "").
				WithPodLevelTarget("2", "").
				WithPodLevelUpperBound("3", "").
				Get(),
			expectedPrio: PodPriority{
				OutsideRecommendedRange: true,
				ResourceDiff:            0.75, // Abs(8-2)/8
				ScaleUp:                 false,
			},
		},
		{
			name: "pod level scale up outside recommended range with one container recommendation",
			pod: test.Pod().
				WithName("POD1").
				AddContainer(
					test.Container().
						WithName(containerName).
						WithCPURequest(resource.MustParse("5")).
						Get()).
				WithCPURequest(resource.MustParse("6")).
				Get(),
			vpa: test.VerticalPodAutoscaler().
				WithContainer(containerName).
				AppendRecommendation(
					test.Recommendation().
						WithContainer(containerName).
						WithLowerBound("3", "").
						WithTarget("4", "").
						WithUpperBound("5", "").
						GetContainerResources()).
				WithPodLevelLowerBound("10", "").
				WithPodLevelTarget("12", "").
				WithPodLevelUpperBound("14", "").
				Get(),
			expectedPrio: PodPriority{
				OutsideRecommendedRange: true,
				ResourceDiff:            0.2 + 1, // Abs(5-4)/5 + Abs(6−12)/6
				ScaleUp:                 true,
			},
		},
		{
			name: "pod level scale up outside recommended range with two container recommendations",
			pod: test.Pod().
				WithName("POD1").
				AddContainer(
					test.Container().
						WithName(containerName).
						WithCPURequest(resource.MustParse("5")).
						Get()).
				AddContainer(
					test.Container().
						WithName(containerName2).
						WithCPURequest(resource.MustParse("7")).
						Get()).
				WithCPURequest(resource.MustParse("6")).
				Get(),
			vpa: test.VerticalPodAutoscaler().
				WithContainer(containerName).
				AppendRecommendation(
					test.Recommendation().
						WithContainer(containerName).
						WithLowerBound("3", "").
						WithTarget("4", "").
						WithUpperBound("5", "").
						GetContainerResources()).
				WithContainer(containerName2).
				AppendRecommendation(
					test.Recommendation().
						WithContainer(containerName2).
						WithLowerBound("5", "").
						WithTarget("5", "").
						WithUpperBound("7", "").
						GetContainerResources()).
				WithPodLevelLowerBound("10", "").
				WithPodLevelTarget("12", "").
				WithPodLevelUpperBound("14", "").
				Get(),
			expectedPrio: PodPriority{
				OutsideRecommendedRange: true,
				ResourceDiff:            0.25 + 1, // Abs(12-9)/12 + Abs(6−12)/6
				ScaleUp:                 true,
			},
		},
		{
			name: "pod level scale down outside recommended range with one container recommendation",
			pod: test.Pod().
				WithName("POD1").
				AddContainer(
					test.Container().
						WithName(containerName).
						WithCPURequest(resource.MustParse("5")).
						Get()).
				WithCPURequest(resource.MustParse("14")).
				Get(),
			vpa: test.VerticalPodAutoscaler().
				WithContainer(containerName).
				AppendRecommendation(
					test.Recommendation().
						WithContainer(containerName).
						WithLowerBound("3", "").
						WithTarget("4", "").
						WithUpperBound("5", "").
						GetContainerResources()).
				WithPodLevelLowerBound("6", "").
				WithPodLevelTarget("7", "").
				WithPodLevelUpperBound("8", "").
				Get(),
			expectedPrio: PodPriority{
				OutsideRecommendedRange: true,
				ResourceDiff:            0.2 + 0.5, // Abs(5-4)/5 + Abs(14-7)/14
				ScaleUp:                 false,
			},
		},
		{
			name: "container level has ResourceDiff while the Pod level does not",
			pod: test.Pod().
				WithName("POD1").
				AddContainer(
					test.Container().
						WithName(containerName).
						WithMemRequest(resource.MustParse("10Mi")).
						Get()).
				WithMemRequest(resource.MustParse("100Mi")).
				Get(),
			vpa: test.VerticalPodAutoscaler().
				WithContainer(containerName).
				AppendRecommendation(
					test.Recommendation().
						WithContainer(containerName).
						WithLowerBound("", "50Mi").
						WithTarget("", "50Mi").
						WithUpperBound("", "50Mi").
						GetContainerResources()).
				WithPodLevelLowerBound("", "100Mi").
				WithPodLevelTarget("", "100Mi").
				WithPodLevelUpperBound("", "100Mi").
				Get(),
			expectedPrio: PodPriority{
				OutsideRecommendedRange: true,
				ResourceDiff:            4, // Abs(10485760000−52428800000)÷10485760000 + 0
				ScaleUp:                 true,
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			processor := NewProcessor()
			prio := processor.GetUpdatePriority(tc.pod, tc.vpa, tc.vpa.Status.Recommendation)
			assert.Equal(t, tc.expectedPrio, prio)
		})
	}
}

// Verify GetUpdatePriority does not encounter a NPE when there is no
// recommendation for a container.
func TestGetUpdatePriority_NoRecommendationForContainer(t *testing.T) {
	p := NewProcessor()
	pod := test.Pod().WithName("POD1").AddContainer(test.Container().WithName("test-container").WithCPURequest(resource.MustParse("5")).WithMemRequest(resource.MustParse("10")).Get()).Get()
	vpa := test.VerticalPodAutoscaler().WithName("test-vpa").WithContainer("test-container").Get()
	result := p.GetUpdatePriority(pod, vpa, nil)
	assert.NotNil(t, result)
}

func TestGetUpdatePriority_VpaObservedContainers(t *testing.T) {
	const (
		// There is no VpaObservedContainers annotation
		// or the container is listed in the annotation.
		optedInContainerDiff = 9
		// There is VpaObservedContainers annotation
		// and the container is not listed in.
		optedOutContainerDiff = 0
	)
	testVpa := test.VerticalPodAutoscaler().WithName("test-vpa").WithContainer(containerName).Get()
	tests := []struct {
		name           string
		pod            *corev1.Pod
		recommendation *vpa_types.RecommendedPodResources
		want           float64
	}{
		{
			name:           "with no VpaObservedContainers annotation",
			pod:            test.Pod().WithName("POD1").AddContainer(test.Container().WithName(containerName).WithCPURequest(resource.MustParse("1")).Get()).Get(),
			recommendation: test.Recommendation().WithContainer(containerName).WithTarget("10", "").Get(),
			want:           optedInContainerDiff,
		},
		{
			name: "with container listed in VpaObservedContainers annotation",
			pod: test.Pod().WithAnnotations(map[string]string{annotations.VpaObservedContainersLabel: containerName}).
				WithName("POD1").AddContainer(test.Container().WithName(containerName).WithCPURequest(resource.MustParse("1")).Get()).Get(),
			recommendation: test.Recommendation().WithContainer(containerName).WithTarget("10", "").Get(),
			want:           optedInContainerDiff,
		},
		{
			name: "with container not listed in VpaObservedContainers annotation",
			pod: test.Pod().WithAnnotations(map[string]string{annotations.VpaObservedContainersLabel: ""}).
				WithName("POD1").AddContainer(test.Container().WithName(containerName).WithCPURequest(resource.MustParse("1")).Get()).Get(),
			recommendation: test.Recommendation().WithContainer(containerName).WithTarget("10", "").Get(),
			want:           optedOutContainerDiff,
		},
		{
			name: "with incorrect VpaObservedContainers annotation",
			pod: test.Pod().WithAnnotations(map[string]string{annotations.VpaObservedContainersLabel: "abcd;';"}).
				WithName("POD1").AddContainer(test.Container().WithName(containerName).WithCPURequest(resource.MustParse("1")).Get()).Get(),
			recommendation: test.Recommendation().WithContainer(containerName).WithTarget("10", "").Get(),
			want:           optedInContainerDiff,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := NewProcessor()
			result := p.GetUpdatePriority(tc.pod, testVpa, tc.recommendation)
			assert.NotNil(t, result)
			// The resourceDiff should be a difference between container resources
			// and container resource recommendations. Containers not listed
			// in an existing vpaObservedContainers annotations shouldn't be taken
			// into account during calculations.
			assert.InDelta(t, result.ResourceDiff, tc.want, 0.0001)
		})
	}
}
