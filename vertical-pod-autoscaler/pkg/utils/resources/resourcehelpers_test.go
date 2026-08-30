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

func TestHasLowerResource(t *testing.T) {
	tests := []struct {
		name     string
		a        corev1.ResourceList
		b        corev1.ResourceList
		expected bool
	}{
		{"Both empty", corev1.ResourceList{}, corev1.ResourceList{}, false},
		{"Identical resources", corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("100m"),
			corev1.ResourceMemory: resource.MustParse("256Mi"),
		}, corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("100m"),
			corev1.ResourceMemory: resource.MustParse("256Mi"),
		}, false},
		{"b has lower CPU", corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("200m"),
			corev1.ResourceMemory: resource.MustParse("256Mi"),
		}, corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("100m"),
			corev1.ResourceMemory: resource.MustParse("256Mi"),
		}, true},
		{"b has lower memory", corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("100m"),
			corev1.ResourceMemory: resource.MustParse("512Mi"),
		}, corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("100m"),
			corev1.ResourceMemory: resource.MustParse("256Mi"),
		}, true},
		{"b has higher resources", corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("100m"),
			corev1.ResourceMemory: resource.MustParse("256Mi"),
		}, corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("200m"),
			corev1.ResourceMemory: resource.MustParse("512Mi"),
		}, false},
		{"b missing key from a", corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("100m"),
			corev1.ResourceMemory: resource.MustParse("256Mi"),
		}, corev1.ResourceList{
			corev1.ResourceCPU: resource.MustParse("50m"),
		}, true},
		{"b has extra key not in a", corev1.ResourceList{
			corev1.ResourceCPU: resource.MustParse("100m"),
		}, corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("100m"),
			corev1.ResourceMemory: resource.MustParse("256Mi"),
		}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HasLowerResource(tt.a, tt.b); got != tt.expected {
				t.Errorf("HasLowerResource() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestRecommendationHasLowerResource(t *testing.T) {
	containerHigh := vpa_types.RecommendedContainerResources{
		ContainerName: "app",
		Target: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("200m"),
			corev1.ResourceMemory: resource.MustParse("512Mi"),
		},
	}

	containerLowCPU := vpa_types.RecommendedContainerResources{
		ContainerName: "app",
		Target: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("100m"),
			corev1.ResourceMemory: resource.MustParse("512Mi"),
		},
	}

	containerSame := vpa_types.RecommendedContainerResources{
		ContainerName: "app",
		Target: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("200m"),
			corev1.ResourceMemory: resource.MustParse("512Mi"),
		},
	}

	containerDiffName := vpa_types.RecommendedContainerResources{
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
		{"Both nil", nil, nil, false},
		{"One nil", &vpa_types.RecommendedPodResources{}, nil, false},
		{"Both empty lists",
			&vpa_types.RecommendedPodResources{ContainerRecommendations: []vpa_types.RecommendedContainerResources{}},
			&vpa_types.RecommendedPodResources{ContainerRecommendations: []vpa_types.RecommendedContainerResources{}},
			false,
		},
		{"Same recommendations",
			&vpa_types.RecommendedPodResources{ContainerRecommendations: []vpa_types.RecommendedContainerResources{containerHigh}},
			&vpa_types.RecommendedPodResources{ContainerRecommendations: []vpa_types.RecommendedContainerResources{containerSame}},
			false,
		},
		{"b has lower CPU for matching container",
			&vpa_types.RecommendedPodResources{ContainerRecommendations: []vpa_types.RecommendedContainerResources{containerHigh}},
			&vpa_types.RecommendedPodResources{ContainerRecommendations: []vpa_types.RecommendedContainerResources{containerLowCPU}},
			true,
		},
		{"No matching container names",
			&vpa_types.RecommendedPodResources{ContainerRecommendations: []vpa_types.RecommendedContainerResources{containerHigh}},
			&vpa_types.RecommendedPodResources{ContainerRecommendations: []vpa_types.RecommendedContainerResources{containerDiffName}},
			false,
		},
		{"Multiple containers one has lower resource",
			&vpa_types.RecommendedPodResources{ContainerRecommendations: []vpa_types.RecommendedContainerResources{containerHigh, containerDiffName}},
			&vpa_types.RecommendedPodResources{ContainerRecommendations: []vpa_types.RecommendedContainerResources{containerLowCPU, containerDiffName}},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RecommendationHasLowerResource(tt.a, tt.b); got != tt.expected {
				t.Errorf("RecommendationHasLowerResource() = %v, want %v", got, tt.expected)
			}
		})
	}
}
