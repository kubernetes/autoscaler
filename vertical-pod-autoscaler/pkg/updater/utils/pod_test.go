/*
Copyright The Kubernetes Authors.

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

package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
)

func TestIsNonDisruptiveResize(t *testing.T) {
	testCases := []struct {
		name     string
		pod      *corev1.Pod
		expected bool
	}{
		{
			name: "No resize policy - defaults to NotRequired",
			pod: &corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "container1"},
					},
				},
			},
			expected: true,
		},
		{
			name: "All NotRequired",
			pod: &corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "container1",
							ResizePolicy: []corev1.ContainerResizePolicy{
								{ResourceName: corev1.ResourceCPU, RestartPolicy: corev1.NotRequired},
								{ResourceName: corev1.ResourceMemory, RestartPolicy: corev1.NotRequired},
							},
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "One RestartContainer",
			pod: &corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "container1",
							ResizePolicy: []corev1.ContainerResizePolicy{
								{ResourceName: corev1.ResourceCPU, RestartPolicy: corev1.NotRequired},
								{ResourceName: corev1.ResourceMemory, RestartPolicy: corev1.RestartContainer},
							},
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "Multiple containers - one with RestartContainer",
			pod: &corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "container1",
							ResizePolicy: []corev1.ContainerResizePolicy{
								{ResourceName: corev1.ResourceCPU, RestartPolicy: corev1.NotRequired},
							},
						},
						{
							Name: "container2",
							ResizePolicy: []corev1.ContainerResizePolicy{
								{ResourceName: corev1.ResourceCPU, RestartPolicy: corev1.RestartContainer},
							},
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "Multiple containers - all NotRequired",
			pod: &corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "container1",
							ResizePolicy: []corev1.ContainerResizePolicy{
								{ResourceName: corev1.ResourceCPU, RestartPolicy: corev1.NotRequired},
							},
						},
						{
							Name: "container2",
							ResizePolicy: []corev1.ContainerResizePolicy{
								{ResourceName: corev1.ResourceMemory, RestartPolicy: corev1.NotRequired},
							},
						},
					},
				},
			},
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsNonDisruptiveResize(tc.pod)
			assert.Equal(t, tc.expected, result)
		})
	}
}
