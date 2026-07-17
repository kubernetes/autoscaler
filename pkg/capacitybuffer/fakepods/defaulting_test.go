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

package fakepods

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestDefaultingResolver(t *testing.T) {
	tests := []struct {
		name        string
		podTemplate *corev1.PodTemplateSpec
		want        *corev1.Pod
	}{
		{
			name: "no resources",
			podTemplate: &corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "c1"},
					},
				},
			},
			want: &corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "c1",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{},
							},
						},
					},
				},
			},
		},
		{
			name: "container limits imply requests",
			podTemplate: &corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "c1",
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU: resource.MustParse("1"),
								},
							},
						},
					},
				},
			},
			want: &corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "c1",
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU: resource.MustParse("1"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU: resource.MustParse("1"),
								},
							},
						},
					},
				},
			},
		},
		{
			name: "init container limits imply requests",
			podTemplate: &corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					InitContainers: []corev1.Container{
						{
							Name: "ic1",
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceMemory: resource.MustParse("1Gi"),
								},
							},
						},
					},
					Containers: []corev1.Container{
						{Name: "c1"},
					},
				},
			},
			want: &corev1.Pod{
				Spec: corev1.PodSpec{
					InitContainers: []corev1.Container{
						{
							Name: "ic1",
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceMemory: resource.MustParse("1Gi"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceMemory: resource.MustParse("1Gi"),
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name: "c1",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{},
							},
						},
					},
				},
			},
		},
		{
			name: "requests already set",
			podTemplate: &corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "c1",
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU: resource.MustParse("1"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU: resource.MustParse("500m"),
								},
							},
						},
					},
				},
			},
			want: &corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "c1",
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU: resource.MustParse("1"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU: resource.MustParse("500m"),
								},
							},
						},
					},
				},
			},
		},
		{
			name: "mixed resources",
			podTemplate: &corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "c1",
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("1"),
									corev1.ResourceMemory: resource.MustParse("1Gi"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU: resource.MustParse("500m"),
								},
							},
						},
					},
				},
			},
			want: &corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "c1",
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("1"),
									corev1.ResourceMemory: resource.MustParse("1Gi"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("500m"),
									corev1.ResourceMemory: resource.MustParse("1Gi"),
								},
							},
						},
					},
				},
			},
		},
		{
			name: "pod level limits imply requests",
			podTemplate: &corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Resources: &corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							corev1.ResourceCPU: resource.MustParse("1"),
						},
					},
					Containers: []corev1.Container{
						{Name: "c1"},
					},
				},
			},
			want: &corev1.Pod{
				Spec: corev1.PodSpec{
					Resources: &corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							corev1.ResourceCPU: resource.MustParse("1"),
						},
						Requests: corev1.ResourceList{
							corev1.ResourceCPU: resource.MustParse("1"),
						},
					},
					Containers: []corev1.Container{
						{
							Name: "c1",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{},
							},
						},
					},
				},
			},
		},
		{
			name: "pod level limits is empty",
			podTemplate: &corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Resources: &corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU: resource.MustParse("100m"),
						},
					},
					Containers: []corev1.Container{
						{Name: "c1"},
					},
				},
			},
			want: &corev1.Pod{
				Spec: corev1.PodSpec{
					Resources: &corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU: resource.MustParse("100m"),
						},
					},
					Containers: []corev1.Container{
						{
							Name: "c1",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{},
							},
						},
					},
				},
			},
		},
		{
			name: "default pod requests for overcommittable resources",
			podTemplate: &corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Resources: &corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							corev1.ResourceCPU: resource.MustParse("1"),
						},
					},
					Containers: []corev1.Container{
						{
							Name: "c1",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceHugePagesPrefix + "2Mi": resource.MustParse("1Gi"),
								},
							},
						},
					},
				},
			},
			want: &corev1.Pod{
				Spec: corev1.PodSpec{
					Resources: &corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							corev1.ResourceCPU: resource.MustParse("1"),
						},
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:                     resource.MustParse("1"),
							corev1.ResourceHugePagesPrefix + "2Mi": resource.MustParse("1Gi"),
						},
					},
					Containers: []corev1.Container{
						{
							Name: "c1",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceHugePagesPrefix + "2Mi": resource.MustParse("1Gi"),
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := NewDefaultingResolver()
			pod, err := resolver.Resolve(t.Context(), "", tt.podTemplate)
			assert.NoError(t, err)
			if diff := cmp.Diff(tt.want, pod, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("DefaultingResolver.Resolve() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
