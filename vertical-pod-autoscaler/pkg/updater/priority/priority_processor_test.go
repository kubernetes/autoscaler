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

	corev1 "k8s.io/api/core/v1"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/annotations"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/test"

	"github.com/stretchr/testify/assert"
)

func TestGetUpdatePriority(t *testing.T) {
	containerName := "test-container"
	testCases := []struct {
		name         string
		pod          *corev1.Pod
		vpa          *vpa_types.VerticalPodAutoscaler
		expectedPrio PodPriority
	}{
		{
			name: "simple scale up",
			pod:  test.Pod().WithName("POD1").AddContainer(test.BuildTestContainer(containerName, "2", "")).Get(),
			vpa:  test.VerticalPodAutoscaler().WithContainer(containerName).WithTarget("10", "").Get(),
			expectedPrio: PodPriority{
				OutsideRecommendedRange: false,
				ResourceDiff:            4.0,
				ScaleUp:                 true,
			},
		}, {
			name: "simple scale down",
			pod:  test.Pod().WithName("POD1").AddContainer(test.BuildTestContainer(containerName, "4", "")).Get(),
			vpa:  test.VerticalPodAutoscaler().WithContainer(containerName).WithTarget("2", "").Get(),
			expectedPrio: PodPriority{
				OutsideRecommendedRange: false,
				ResourceDiff:            0.5,
				ScaleUp:                 false,
			},
		}, {
			name: "no resource diff",
			pod:  test.Pod().WithName("POD1").AddContainer(test.BuildTestContainer(containerName, "2", "")).Get(),
			vpa:  test.VerticalPodAutoscaler().WithContainer(containerName).WithTarget("2", "").Get(),
			expectedPrio: PodPriority{
				OutsideRecommendedRange: false,
				ResourceDiff:            0.0,
				ScaleUp:                 false,
			},
		}, {
			name: "scale up on milliquanitites",
			pod:  test.Pod().WithName("POD1").AddContainer(test.BuildTestContainer(containerName, "10m", "")).Get(),
			vpa:  test.VerticalPodAutoscaler().WithContainer(containerName).WithTarget("900m", "").Get(),
			expectedPrio: PodPriority{
				OutsideRecommendedRange: false,
				ResourceDiff:            89.0,
				ScaleUp:                 true,
			},
		}, {
			name: "scale up outside recommended range",
			pod:  test.Pod().WithName("POD1").AddContainer(test.BuildTestContainer(containerName, "4", "")).Get(),
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
			pod:  test.Pod().WithName("POD1").AddContainer(test.BuildTestContainer(containerName, "8", "")).Get(),
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
			pod:  test.Pod().WithName("POD1").AddContainer(test.BuildTestContainer(containerName, "2", "")).Get(),
			vpa:  test.VerticalPodAutoscaler().WithContainer(containerName).WithTarget("10", "").Get(),
			expectedPrio: PodPriority{
				OutsideRecommendedRange: false,
				ResourceDiff:            4.0,
				ScaleUp:                 true,
			},
		}, {
			name: "multiple resources, both scale up",
			pod:  test.Pod().WithName("POD1").AddContainer(test.BuildTestContainer(containerName, "3", "10M")).Get(),
			vpa:  test.VerticalPodAutoscaler().WithContainer(containerName).WithTarget("6", "20M").Get(),
			expectedPrio: PodPriority{
				OutsideRecommendedRange: false,
				ResourceDiff:            1.0 + 1.0, // summed relative diffs for resources
				ScaleUp:                 true,
			},
		}, {
			name: "multiple resources, only one scale up",
			pod:  test.Pod().WithName("POD1").AddContainer(test.BuildTestContainer(containerName, "4", "10M")).Get(),
			vpa:  test.VerticalPodAutoscaler().WithContainer(containerName).WithTarget("2", "20M").Get(),
			expectedPrio: PodPriority{
				OutsideRecommendedRange: false,
				ResourceDiff:            1.5 + 0.0, // summed relative diffs for resources
				ScaleUp:                 true,
			},
		}, {
			name: "multiple resources, both scale down",
			pod:  test.Pod().WithName("POD1").AddContainer(test.BuildTestContainer(containerName, "4", "20M")).Get(),
			vpa:  test.VerticalPodAutoscaler().WithContainer(containerName).WithTarget("2", "10M").Get(),
			expectedPrio: PodPriority{
				OutsideRecommendedRange: false,
				ResourceDiff:            0.5 + 0.5, // summed relative diffs for resources
				ScaleUp:                 false,
			},
		}, {
			name: "multiple resources, one outside recommended range",
			pod:  test.Pod().WithName("POD1").AddContainer(test.BuildTestContainer(containerName, "4", "20M")).Get(),
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
			pod: test.Pod().WithName("POD1").AddContainer(test.BuildTestContainer(containerName, "1", "")).
				AddContainer(test.BuildTestContainer("test-container-2", "2", "")).Get(),
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
			pod: test.Pod().WithName("POD1").AddContainer(test.BuildTestContainer(containerName, "3", "")).
				AddContainer(test.BuildTestContainer("test-container-2", "7", "")).Get(),
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
			pod: test.Pod().WithName("POD1").AddContainer(test.BuildTestContainer(containerName, "1", "")).
				AddContainer(test.BuildTestContainer("test-container-2", "2", "")).Get(),
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
			pod: test.Pod().WithName("POD1").AddContainer(test.BuildTestContainer(containerName, "6", "10M")).
				AddContainer(test.BuildTestContainer("test-container-2", "4", "30M")).Get(),
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
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			processor := NewProcessor()
			prio := processor.GetUpdatePriority(tc.pod, tc.vpa, tc.vpa.Status.Recommendation)
			assert.Equal(t, tc.expectedPrio, prio)
		})
	}
}

// Verify GetUpdatePriorty does not encounter a NPE when there is no
// recommendation for a container.
func TestGetUpdatePriority_NoRecommendationForContainer(t *testing.T) {
	p := NewProcessor()
	pod := test.Pod().WithName("POD1").AddContainer(test.BuildTestContainer("test-container", "5", "10")).Get()
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
			pod:            test.Pod().WithName("POD1").AddContainer(test.BuildTestContainer(containerName, "1", "")).Get(),
			recommendation: test.Recommendation().WithContainer(containerName).WithTarget("10", "").Get(),
			want:           optedInContainerDiff,
		},
		{
			name: "with container listed in VpaObservedContainers annotation",
			pod: test.Pod().WithAnnotations(map[string]string{annotations.VpaObservedContainersLabel: containerName}).
				WithName("POD1").AddContainer(test.BuildTestContainer(containerName, "1", "")).Get(),
			recommendation: test.Recommendation().WithContainer(containerName).WithTarget("10", "").Get(),
			want:           optedInContainerDiff,
		},
		{
			name: "with container not listed in VpaObservedContainers annotation",
			pod: test.Pod().WithAnnotations(map[string]string{annotations.VpaObservedContainersLabel: ""}).
				WithName("POD1").AddContainer(test.BuildTestContainer(containerName, "1", "")).Get(),
			recommendation: test.Recommendation().WithContainer(containerName).WithTarget("10", "").Get(),
			want:           optedOutContainerDiff,
		},
		{
			name: "with incorrect VpaObservedContainers annotation",
			pod: test.Pod().WithAnnotations(map[string]string{annotations.VpaObservedContainersLabel: "abcd;';"}).
				WithName("POD1").AddContainer(test.BuildTestContainer(containerName, "1", "")).Get(),
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
