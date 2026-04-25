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

package resourcehelpers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/test"
)

func TestContainerRequestsAndLimits(t *testing.T) {
	testCases := []struct {
		desc          string
		containerName string
		pod           *corev1.Pod
		wantRequests  corev1.ResourceList
		wantLimits    corev1.ResourceList
	}{
		{
			desc:          "Prefer resource requests from container status",
			containerName: "container",
			pod: test.Pod().AddContainer(
				test.Container().WithName("container").
					WithCPURequest(resource.MustParse("1")).
					WithMemRequest(resource.MustParse("10Mi")).
					WithCPULimit(resource.MustParse("2")).
					WithMemLimit(resource.MustParse("20Mi")).Get()).
				AddContainerStatus(
					test.ContainerStatus().WithName("container").
						WithCPURequest(resource.MustParse("3")).
						WithMemRequest(resource.MustParse("30Mi")).
						WithCPULimit(resource.MustParse("4")).
						WithMemLimit(resource.MustParse("40Mi")).Get()).Get(),
			wantRequests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("3"),
				corev1.ResourceMemory: resource.MustParse("30Mi"),
			},
			wantLimits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("4"),
				corev1.ResourceMemory: resource.MustParse("40Mi"),
			},
		},
		{
			desc:          "No container status, get resources from pod spec",
			containerName: "container",
			pod: test.Pod().AddContainer(
				test.Container().WithName("container").
					WithCPURequest(resource.MustParse("1")).
					WithMemRequest(resource.MustParse("10Mi")).
					WithCPULimit(resource.MustParse("2")).
					WithMemLimit(resource.MustParse("20Mi")).Get()).Get(),
			wantRequests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("1"),
				corev1.ResourceMemory: resource.MustParse("10Mi"),
			},
			wantLimits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("2"),
				corev1.ResourceMemory: resource.MustParse("20Mi"),
			},
		},
		{
			desc:          "Only containerStatus, get resources from containerStatus",
			containerName: "container",
			pod: test.Pod().AddContainerStatus(
				test.ContainerStatus().WithName("container").
					WithCPURequest(resource.MustParse("0")).
					WithMemRequest(resource.MustParse("30Mi")).
					WithCPULimit(resource.MustParse("4")).
					WithMemLimit(resource.MustParse("40Mi")).Get()).Get(),
			wantRequests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("0"),
				corev1.ResourceMemory: resource.MustParse("30Mi"),
			},
			wantLimits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("4"),
				corev1.ResourceMemory: resource.MustParse("40Mi"),
			},
		},
		{
			desc:          "Inexistent container",
			containerName: "inexistent-container",
			pod: test.Pod().AddContainer(
				test.Container().WithName("container").
					WithCPURequest(resource.MustParse("1")).
					WithMemRequest(resource.MustParse("10Mi")).
					WithCPULimit(resource.MustParse("2")).
					WithMemLimit(resource.MustParse("20Mi")).Get()).Get(),
			wantRequests: nil,
			wantLimits:   nil,
		},
		{
			desc:          "Init container with the same name as the container is ignored",
			containerName: "container-1",
			pod: test.Pod().AddInitContainer(
				test.Container().WithName("container-1").
					WithCPURequest(resource.MustParse("1")).
					WithMemRequest(resource.MustParse("10Mi")).
					WithCPULimit(resource.MustParse("2")).
					WithMemLimit(resource.MustParse("20Mi")).Get()).
				AddContainer(
					test.Container().WithName("container-1").
						WithCPURequest(resource.MustParse("4")).
						WithMemRequest(resource.MustParse("40Mi")).
						WithCPULimit(resource.MustParse("5")).
						WithMemLimit(resource.MustParse("50Mi")).Get()).
				Get(),
			wantRequests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("4"),
				corev1.ResourceMemory: resource.MustParse("40Mi"),
			},
			wantLimits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("5"),
				corev1.ResourceMemory: resource.MustParse("50Mi"),
			},
		},
		{
			desc:          "Container with no requests or limits returns non-nil resources",
			containerName: "container",
			pod:           test.Pod().AddContainer(test.Container().WithName("container").Get()).Get(),
			wantRequests:  corev1.ResourceList{},
			wantLimits:    corev1.ResourceList{},
		},
		{
			desc:          "2 containers",
			containerName: "container-1",
			pod: test.Pod().AddContainer(
				test.Container().WithName("container-1").
					WithCPURequest(resource.MustParse("1")).
					WithMemRequest(resource.MustParse("10Mi")).
					WithCPULimit(resource.MustParse("2")).
					WithMemLimit(resource.MustParse("20Mi")).Get()).
				AddContainerStatus(
					test.ContainerStatus().WithName("container-1").
						WithCPURequest(resource.MustParse("3")).
						WithMemRequest(resource.MustParse("30Mi")).
						WithCPULimit(resource.MustParse("4")).
						WithMemLimit(resource.MustParse("40Mi")).Get()).
				AddContainer(
					test.Container().WithName("container-2").
						WithCPURequest(resource.MustParse("5")).
						WithMemRequest(resource.MustParse("5Mi")).
						WithCPULimit(resource.MustParse("5")).
						WithMemLimit(resource.MustParse("5Mi")).Get()).
				AddContainerStatus(
					test.ContainerStatus().WithName("container-2").
						WithCPURequest(resource.MustParse("5")).
						WithMemRequest(resource.MustParse("5Mi")).
						WithCPULimit(resource.MustParse("5")).
						WithMemLimit(resource.MustParse("5Mi")).Get()).
				Get(),
			wantRequests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("3"),
				corev1.ResourceMemory: resource.MustParse("30Mi"),
			},
			wantLimits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("4"),
				corev1.ResourceMemory: resource.MustParse("40Mi"),
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			gotRequests, gotLimits := ContainerRequestsAndLimits(tc.containerName, tc.pod)
			assert.Equal(t, tc.wantRequests, gotRequests, "requests don't match")
			assert.Equal(t, tc.wantLimits, gotLimits, "limits don't match")
		})
	}
}

func TestInitContainerRequestsAndLimits(t *testing.T) {
	testCases := []struct {
		desc              string
		initContainerName string
		pod               *corev1.Pod
		wantRequests      corev1.ResourceList
		wantLimits        corev1.ResourceList
	}{
		{
			desc:              "Prefer resource requests from initContainer status",
			initContainerName: "init-container",
			pod: test.Pod().AddInitContainer(
				test.Container().WithName("init-container").
					WithCPURequest(resource.MustParse("1")).
					WithMemRequest(resource.MustParse("10Mi")).
					WithCPULimit(resource.MustParse("2")).
					WithMemLimit(resource.MustParse("20Mi")).Get()).
				AddInitContainerStatus(
					test.ContainerStatus().WithName("init-container").
						WithCPURequest(resource.MustParse("3")).
						WithMemRequest(resource.MustParse("30Mi")).
						WithCPULimit(resource.MustParse("4")).
						WithMemLimit(resource.MustParse("40Mi")).Get()).Get(),
			wantRequests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("3"),
				corev1.ResourceMemory: resource.MustParse("30Mi"),
			},
			wantLimits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("4"),
				corev1.ResourceMemory: resource.MustParse("40Mi"),
			},
		},
		{
			desc:              "No initContainer status, get resources from pod spec",
			initContainerName: "init-container",
			pod: test.Pod().AddInitContainer(
				test.Container().WithName("init-container").
					WithCPURequest(resource.MustParse("1")).
					WithMemRequest(resource.MustParse("10Mi")).
					WithCPULimit(resource.MustParse("2")).
					WithMemLimit(resource.MustParse("20Mi")).Get()).Get(),
			wantRequests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("1"),
				corev1.ResourceMemory: resource.MustParse("10Mi"),
			},
			wantLimits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("2"),
				corev1.ResourceMemory: resource.MustParse("20Mi"),
			},
		},
		{
			desc:              "Only initContainerStatus, get resources from initContainerStatus",
			initContainerName: "init-container",
			pod: test.Pod().AddInitContainerStatus(
				test.ContainerStatus().WithName("init-container").
					WithCPURequest(resource.MustParse("0")).
					WithMemRequest(resource.MustParse("30Mi")).
					WithCPULimit(resource.MustParse("4")).
					WithMemLimit(resource.MustParse("40Mi")).Get()).Get(),
			wantRequests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("0"),
				corev1.ResourceMemory: resource.MustParse("30Mi"),
			},
			wantLimits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("4"),
				corev1.ResourceMemory: resource.MustParse("40Mi"),
			},
		},
		{
			desc:              "Inexistent initContainer",
			initContainerName: "inexistent-init-container",
			pod: test.Pod().AddInitContainer(
				test.Container().WithName("init-container").
					WithCPURequest(resource.MustParse("1")).
					WithMemRequest(resource.MustParse("10Mi")).
					WithCPULimit(resource.MustParse("2")).
					WithMemLimit(resource.MustParse("20Mi")).Get()).Get(),
			wantRequests: nil,
			wantLimits:   nil,
		},
		{
			desc:              "Container with the same name as the initContainer is ignored",
			initContainerName: "container-1",
			pod: test.Pod().AddInitContainer(
				test.Container().WithName("container-1").
					WithCPURequest(resource.MustParse("1")).
					WithMemRequest(resource.MustParse("10Mi")).
					WithCPULimit(resource.MustParse("2")).
					WithMemLimit(resource.MustParse("20Mi")).Get()).
				AddContainer(
					test.Container().WithName("container-1").
						WithCPURequest(resource.MustParse("4")).
						WithMemRequest(resource.MustParse("40Mi")).
						WithCPULimit(resource.MustParse("5")).
						WithMemLimit(resource.MustParse("50Mi")).Get()).
				Get(),
			wantRequests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("1"),
				corev1.ResourceMemory: resource.MustParse("10Mi"),
			},
			wantLimits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("2"),
				corev1.ResourceMemory: resource.MustParse("20Mi"),
			},
		},
		{
			desc:              "InitContainer with no requests or limits returns non-nil resources",
			initContainerName: "init-container",
			pod:               test.Pod().AddInitContainer(test.Container().WithName("init-container").Get()).Get(),
			wantRequests:      corev1.ResourceList{},
			wantLimits:        corev1.ResourceList{},
		},
		{
			desc:              "2 init containers",
			initContainerName: "init-container-1",
			pod: test.Pod().AddInitContainer(
				test.Container().WithName("init-container-1").
					WithCPURequest(resource.MustParse("1")).
					WithMemRequest(resource.MustParse("10Mi")).
					WithCPULimit(resource.MustParse("2")).
					WithMemLimit(resource.MustParse("20Mi")).Get()).
				AddInitContainerStatus(
					test.ContainerStatus().WithName("init-container-1").
						WithCPURequest(resource.MustParse("3")).
						WithMemRequest(resource.MustParse("30Mi")).
						WithCPULimit(resource.MustParse("4")).
						WithMemLimit(resource.MustParse("40Mi")).Get()).
				AddInitContainer(
					test.Container().WithName("init-container-2").
						WithCPURequest(resource.MustParse("5")).
						WithMemRequest(resource.MustParse("5Mi")).
						WithCPULimit(resource.MustParse("5")).
						WithMemLimit(resource.MustParse("5Mi")).Get()).
				AddInitContainerStatus(
					test.ContainerStatus().WithName("init-container-2").
						WithCPURequest(resource.MustParse("5")).
						WithMemRequest(resource.MustParse("5Mi")).
						WithCPULimit(resource.MustParse("5")).
						WithMemLimit(resource.MustParse("5Mi")).Get()).
				Get(),
			wantRequests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("3"),
				corev1.ResourceMemory: resource.MustParse("30Mi"),
			},
			wantLimits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("4"),
				corev1.ResourceMemory: resource.MustParse("40Mi"),
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			gotRequests, gotLimits := InitContainerRequestsAndLimits(tc.initContainerName, tc.pod)
			assert.Equal(t, tc.wantRequests, gotRequests, "requests don't match")
			assert.Equal(t, tc.wantLimits, gotLimits, "limits don't match")
		})
	}
}

func TestSumContainerLevelRecommendations(t *testing.T) {
	cases := []struct {
		name                       string
		containerRecommendations   []vpa_types.RecommendedContainerResources
		expectedPodRecommendations *vpa_types.PodRecommendations
	}{
		{
			name:                     "no container recommendations",
			containerRecommendations: nil,
		},
		{
			name:                     "no container recommendations",
			containerRecommendations: []vpa_types.RecommendedContainerResources{},
		},
		{
			name: "one container resource does not include a recommendations",
			containerRecommendations: []vpa_types.RecommendedContainerResources{
				test.Recommendation().
					WithContainer("container1").
					WithTarget("2m", "10Mi").
					WithLowerBound("2m", "10Mi").
					WithUpperBound("2m", "10Mi").
					GetContainerResources(),
				test.Recommendation().
					WithContainer("container2").
					WithTarget("8m", "").
					WithLowerBound("8m", "").
					WithUpperBound("8m", "").
					GetContainerResources(),
			},
			expectedPodRecommendations: &vpa_types.PodRecommendations{
				Target: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("10m"),
					corev1.ResourceMemory: resource.MustParse("10Mi"),
				},
				LowerBound: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("10m"),
					corev1.ResourceMemory: resource.MustParse("10Mi"),
				},
				UpperBound: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("10m"),
					corev1.ResourceMemory: resource.MustParse("10Mi"),
				},
				UncappedTarget: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("10m"),
					corev1.ResourceMemory: resource.MustParse("10Mi"),
				},
			},
		},
		{
			name: "all containers contain recommendations",
			containerRecommendations: []vpa_types.RecommendedContainerResources{
				test.Recommendation().
					WithContainer("container1").
					WithTarget("4m", "4Mi").
					WithLowerBound("4m", "4Mi").
					WithUpperBound("4m", "4Mi").
					GetContainerResources(),
				test.Recommendation().
					WithContainer("container2").
					WithTarget("4m", "4Mi").
					WithLowerBound("4m", "4Mi").
					WithUpperBound("4m", "4Mi").
					GetContainerResources(),
				test.Recommendation().
					WithContainer("container3").
					WithTarget("2m", "2Mi").
					WithLowerBound("2m", "2Mi").
					WithUpperBound("2m", "2Mi").
					GetContainerResources(),
			},
			expectedPodRecommendations: &vpa_types.PodRecommendations{
				Target: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("10m"),
					corev1.ResourceMemory: resource.MustParse("10Mi"),
				},
				LowerBound: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("10m"),
					corev1.ResourceMemory: resource.MustParse("10Mi"),
				},
				UpperBound: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("10m"),
					corev1.ResourceMemory: resource.MustParse("10Mi"),
				},
				UncappedTarget: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("10m"),
					corev1.ResourceMemory: resource.MustParse("10Mi"),
				},
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			outRecommendations := SumContainerLevelRecommendations(tc.containerRecommendations)
			if outRecommendations != nil {
				assert.Equal(t, tc.expectedPodRecommendations.Target.Cpu().MilliValue(), outRecommendations.Target.Cpu().MilliValue())
				assert.Equal(t, tc.expectedPodRecommendations.Target.Memory().Value(), outRecommendations.Target.Memory().Value())

				assert.Equal(t, tc.expectedPodRecommendations.LowerBound.Cpu().MilliValue(), outRecommendations.LowerBound.Cpu().MilliValue())
				assert.Equal(t, tc.expectedPodRecommendations.LowerBound.Memory().Value(), outRecommendations.LowerBound.Memory().Value())

				assert.Equal(t, tc.expectedPodRecommendations.UpperBound.Cpu().MilliValue(), outRecommendations.UpperBound.Cpu().MilliValue())
				assert.Equal(t, tc.expectedPodRecommendations.UpperBound.Memory().Value(), outRecommendations.UpperBound.Memory().Value())

				assert.Equal(t, tc.expectedPodRecommendations.UncappedTarget.Cpu().MilliValue(), outRecommendations.UncappedTarget.Cpu().MilliValue())
				assert.Equal(t, tc.expectedPodRecommendations.UncappedTarget.Memory().Value(), outRecommendations.UncappedTarget.Memory().Value())
			} else {
				assert.Nil(t, outRecommendations, "No recommendations are expected")
			}
		})
	}
}

func TestPodRequestsAndLimits(t *testing.T) {
	testCases := []struct {
		name         string
		pod          *corev1.Pod
		wantRequests corev1.ResourceList
		wantLimits   corev1.ResourceList
	}{
		{
			name:         "resources in the pod spec and the resources in the pod status are both omitted",
			pod:          test.Pod().Get(),
			wantRequests: corev1.ResourceList{},
			wantLimits:   corev1.ResourceList{},
		},
		{
			name: "resources in the pod spec is set and no resources in the pod status",
			pod: test.Pod().
				WithCPURequest(resource.MustParse("100m")).
				WithMemRequest(resource.MustParse("100Mi")).
				WithCPULimit(resource.MustParse("120m")).
				WithMemLimit(resource.MustParse("120Mi")).
				Get(),
			wantRequests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("100m"),
				corev1.ResourceMemory: resource.MustParse("100Mi"),
			},
			wantLimits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("120m"),
				corev1.ResourceMemory: resource.MustParse("120Mi"),
			},
		},
		// TODO: Comment out this part once https://github.com/kubernetes/kubernetes/issues/137628 is fixed and backported to Kubernetes versions.
		// {
		// 	name: "resources in the pod spec and the pod status are both set but the latter is preferred",
		// 	pod: test.Pod().
		// 		WithCPURequest(resource.MustParse("100m")).
		// 		WithMemRequest(resource.MustParse("100Mi")).
		// 		WithCPULimit(resource.MustParse("120m")).
		// 		WithMemLimit(resource.MustParse("120Mi")).
		// 		AddPodStatus(
		// 			test.PodStatus().
		// 				WithCPURequest(resource.MustParse("111m")).
		// 				WithMemRequest(resource.MustParse("111Mi")).
		// 				WithCPULimit(resource.MustParse("133m")).
		// 				WithMemLimit(resource.MustParse("133Mi")).Get()).
		// 		Get(),
		// 	wantRequests: corev1.ResourceList{
		// 		corev1.ResourceCPU:    resource.MustParse("111m"),
		// 		corev1.ResourceMemory: resource.MustParse("111Mi"),
		// 	},
		// 	wantLimits: corev1.ResourceList{
		// 		corev1.ResourceCPU:    resource.MustParse("133m"),
		// 		corev1.ResourceMemory: resource.MustParse("133Mi"),
		// 	},
		// },
		// TODO: Comment out this part once https://github.com/kubernetes/kubernetes/issues/137628 is fixed and backported to Kubernetes versions.
		// {
		// 	name: "resources in the pod spec and the pod status are both set but the latter is preferred only cpu",
		// 	pod: test.Pod().
		// 		WithCPURequest(resource.MustParse("100m")).
		// 		WithCPULimit(resource.MustParse("120m")).
		// 		AddPodStatus(
		// 			test.PodStatus().
		// 				WithCPURequest(resource.MustParse("111m")).
		// 				WithCPULimit(resource.MustParse("133m")).Get()).
		// 		Get(),
		// 	wantRequests: corev1.ResourceList{
		// 		corev1.ResourceCPU: resource.MustParse("111m"),
		// 	},
		// 	wantLimits: corev1.ResourceList{
		// 		corev1.ResourceCPU: resource.MustParse("133m"),
		// 	},
		// },
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gotRequests, gotLimits := PodRequestsAndLimits(tc.pod)
			assert.Equal(t, tc.wantRequests, gotRequests, "Pod-level requests do not match")
			assert.Equal(t, tc.wantLimits, gotLimits, "Pod-level limits do not match")
		})
	}
}
