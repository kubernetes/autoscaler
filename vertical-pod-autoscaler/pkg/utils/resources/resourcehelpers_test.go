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

func TestResourcesEqual(t *testing.T) {
	resA := corev1.ResourceList{
		corev1.ResourceCPU:    resource.MustParse("100m"),
		corev1.ResourceMemory: resource.MustParse("256Mi"),
	}
	resB := corev1.ResourceList{
		corev1.ResourceCPU:    resource.MustParse("100m"),
		corev1.ResourceMemory: resource.MustParse("256Mi"),
	}
	resC := corev1.ResourceList{
		corev1.ResourceCPU:    resource.MustParse("200m"),
		corev1.ResourceMemory: resource.MustParse("256Mi"),
	}

	tests := []struct {
		name     string
		a        corev1.ResourceList
		b        corev1.ResourceList
		expected bool
	}{
		{"Both empty", corev1.ResourceList{}, corev1.ResourceList{}, true},
		{"Identical resources", resA, resB, true},
		{"Different values", resA, resC, false},
		{"Different lengths", resA, corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("100m")}, false},
		{"Missing key", resA, corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("100m"), "disk": resource.MustParse("1Gi")}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ResourcesEqual(tt.a, tt.b); got != tt.expected {
				t.Errorf("ResourcesEqual() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestRecommendationsEqual(t *testing.T) {
	container1 := vpa_types.RecommendedContainerResources{
		ContainerName: "app",
		Target: corev1.ResourceList{
			corev1.ResourceCPU: resource.MustParse("100m"),
		},
	}

	container2 := vpa_types.RecommendedContainerResources{
		ContainerName: "sidecar",
		Target: corev1.ResourceList{
			corev1.ResourceCPU: resource.MustParse("50m"),
		},
	}

	tests := []struct {
		name     string
		a        *vpa_types.RecommendedPodResources
		b        *vpa_types.RecommendedPodResources
		expected bool
	}{
		{"Both nil", nil, nil, true},
		{"One nil", &vpa_types.RecommendedPodResources{}, nil, false},
		{"Both empty lists",
			&vpa_types.RecommendedPodResources{ContainerRecommendations: []vpa_types.RecommendedContainerResources{}},
			&vpa_types.RecommendedPodResources{ContainerRecommendations: []vpa_types.RecommendedContainerResources{}},
			true,
		},
		{"Matching recommendations",
			&vpa_types.RecommendedPodResources{ContainerRecommendations: []vpa_types.RecommendedContainerResources{container1}},
			&vpa_types.RecommendedPodResources{ContainerRecommendations: []vpa_types.RecommendedContainerResources{container1}},
			true,
		},
		{"Different container names",
			&vpa_types.RecommendedPodResources{ContainerRecommendations: []vpa_types.RecommendedContainerResources{container1}},
			&vpa_types.RecommendedPodResources{ContainerRecommendations: []vpa_types.RecommendedContainerResources{container2}},
			false,
		},
		{"Different number of containers",
			&vpa_types.RecommendedPodResources{ContainerRecommendations: []vpa_types.RecommendedContainerResources{container1, container2}},
			&vpa_types.RecommendedPodResources{ContainerRecommendations: []vpa_types.RecommendedContainerResources{container1}},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RecommendationsEqual(tt.a, tt.b); got != tt.expected {
				t.Errorf("RecommendationsEqual() = %v, want %v", got, tt.expected)
			}
		})
	}
}
