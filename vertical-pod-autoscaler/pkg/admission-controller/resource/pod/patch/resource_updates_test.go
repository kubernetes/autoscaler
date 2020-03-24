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

package patch

import (
	"fmt"
	"strings"
	"testing"

	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	resource_admission "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/admission-controller/resource"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/test"
	vpa_api_util "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/vpa"

	"github.com/stretchr/testify/assert"
)

const (
	cpu        = "cpu"
	unobtanium = "unobtanium"
	limit      = "limit"
	request    = "request"
)

type fakeRecommendationProvider struct {
	resources              []vpa_api_util.ContainerResources
	containerToAnnotations vpa_api_util.ContainerToAnnotationsMap
	e                      error
}

func (frp *fakeRecommendationProvider) GetContainersResourcesForPod(pod *core.Pod, vpa *vpa_types.VerticalPodAutoscaler) ([]vpa_api_util.ContainerResources, vpa_api_util.ContainerToAnnotationsMap, error) {
	return frp.resources, frp.containerToAnnotations, frp.e
}

func addResourcesPatch(idx int) resource_admission.PatchRecord {
	return resource_admission.PatchRecord{
		"add",
		fmt.Sprintf("/spec/containers/%d/resources", idx),
		core.ResourceRequirements{},
	}
}

func addRequestsPatch(idx int) resource_admission.PatchRecord {
	return resource_admission.PatchRecord{
		"add",
		fmt.Sprintf("/spec/containers/%d/resources/requests", idx),
		core.ResourceList{},
	}
}

func addLimitsPatch(idx int) resource_admission.PatchRecord {
	return resource_admission.PatchRecord{
		"add",
		fmt.Sprintf("/spec/containers/%d/resources/limits", idx),
		core.ResourceList{},
	}
}

func addResourceRequestPatch(index int, res, amount string) resource_admission.PatchRecord {
	return resource_admission.PatchRecord{
		"add",
		fmt.Sprintf("/spec/containers/%d/resources/requests/%s", index, res),
		resource.MustParse(amount),
	}
}

func addResourceLimitPatch(index int, res, amount string) resource_admission.PatchRecord {
	return resource_admission.PatchRecord{
		"add",
		fmt.Sprintf("/spec/containers/%d/resources/limits/%s", index, res),
		resource.MustParse(amount),
	}
}

func addAnnotationRequest(updateResources [][]string, kind string) resource_admission.PatchRecord {
	requests := make([]string, 0)
	for idx, podResources := range updateResources {
		podRequests := make([]string, 0)
		for _, resource := range podResources {
			podRequests = append(podRequests, resource+" "+kind)
		}
		requests = append(requests, fmt.Sprintf("container %d: %s", idx, strings.Join(podRequests, ", ")))
	}

	vpaUpdates := fmt.Sprintf("Pod resources updated by name: %s", strings.Join(requests, "; "))
	return GetAddAnnotationPatch(ResourceUpdatesAnnotation, vpaUpdates)
}

func TestClalculatePatches_ResourceUpdates(t *testing.T) {
	tests := []struct {
		name                 string
		pod                  *core.Pod
		namespace            string
		recommendResources   []vpa_api_util.ContainerResources
		recommendAnnotations vpa_api_util.ContainerToAnnotationsMap
		recommendError       error
		expectPatches        []resource_admission.PatchRecord
		expectError          error
	}{
		{
			name: "new cpu recommendation",
			pod: &core.Pod{
				Spec: core.PodSpec{
					Containers: []core.Container{{}},
				},
			},
			namespace: "default",
			recommendResources: []vpa_api_util.ContainerResources{
				{
					Requests: core.ResourceList{
						cpu: resource.MustParse("1"),
					},
				},
			},
			recommendAnnotations: vpa_api_util.ContainerToAnnotationsMap{},
			expectPatches: []resource_admission.PatchRecord{
				addResourcesPatch(0),
				addRequestsPatch(0),
				addResourceRequestPatch(0, cpu, "1"),
				addAnnotationRequest([][]string{{cpu}}, request),
			},
		},
		{
			name: "replacement cpu recommendation",
			pod: &core.Pod{
				Spec: core.PodSpec{
					Containers: []core.Container{{
						Resources: core.ResourceRequirements{
							Requests: core.ResourceList{
								cpu: resource.MustParse("0"),
							},
						},
					}},
				},
			},
			namespace: "default",
			recommendResources: []vpa_api_util.ContainerResources{
				{
					Requests: core.ResourceList{
						cpu: resource.MustParse("1"),
					},
				},
			},
			recommendAnnotations: vpa_api_util.ContainerToAnnotationsMap{},
			expectPatches: []resource_admission.PatchRecord{
				addResourceRequestPatch(0, cpu, "1"),
				addAnnotationRequest([][]string{{cpu}}, request),
			},
		},
		{
			name: "two containers",
			pod: &core.Pod{
				Spec: core.PodSpec{
					Containers: []core.Container{{
						Resources: core.ResourceRequirements{
							Requests: core.ResourceList{
								cpu: resource.MustParse("0"),
							},
						},
					}, {}},
				},
			},
			namespace: "default",
			recommendResources: []vpa_api_util.ContainerResources{
				{
					Requests: core.ResourceList{
						cpu: resource.MustParse("1"),
					},
				},
				{
					Requests: core.ResourceList{
						cpu: resource.MustParse("2"),
					},
				},
			},
			recommendAnnotations: vpa_api_util.ContainerToAnnotationsMap{},
			expectPatches: []resource_admission.PatchRecord{
				addResourceRequestPatch(0, cpu, "1"),
				addResourcesPatch(1),
				addRequestsPatch(1),
				addResourceRequestPatch(1, cpu, "2"),
				addAnnotationRequest([][]string{{cpu}, {cpu}}, request),
			},
		},
		{
			name: "new cpu limit",
			pod: &core.Pod{
				Spec: core.PodSpec{
					Containers: []core.Container{{}},
				},
			},
			namespace: "default",
			recommendResources: []vpa_api_util.ContainerResources{
				{
					Limits: core.ResourceList{
						cpu: resource.MustParse("1"),
					},
				},
			},
			recommendAnnotations: vpa_api_util.ContainerToAnnotationsMap{},
			expectPatches: []resource_admission.PatchRecord{
				addResourcesPatch(0),
				addLimitsPatch(0),
				addResourceLimitPatch(0, cpu, "1"),
				addAnnotationRequest([][]string{{cpu}}, limit),
			},
		},
		{
			name: "replacement cpu limit",
			pod: &core.Pod{
				Spec: core.PodSpec{
					Containers: []core.Container{{
						Resources: core.ResourceRequirements{
							Limits: core.ResourceList{
								cpu: resource.MustParse("0"),
							},
						},
					}},
				},
			},
			namespace: "default",
			recommendResources: []vpa_api_util.ContainerResources{
				{
					Limits: core.ResourceList{
						cpu: resource.MustParse("1"),
					},
				},
			},
			recommendAnnotations: vpa_api_util.ContainerToAnnotationsMap{},
			expectPatches: []resource_admission.PatchRecord{
				addResourceLimitPatch(0, cpu, "1"),
				addAnnotationRequest([][]string{{cpu}}, limit),
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			frp := fakeRecommendationProvider{tc.recommendResources, tc.recommendAnnotations, tc.recommendError}
			c := NewResourceUpdatesCalculator(&frp)
			patches, err := c.CalculatePatches(tc.pod, test.VerticalPodAutoscaler().WithContainer("test").WithName("name").Get())
			if tc.expectError == nil {
				assert.NoError(t, err)
			} else {
				if assert.Error(t, err) {
					assert.Equal(t, tc.expectError.Error(), err.Error())
				}
			}
			if assert.Len(t, patches, len(tc.expectPatches), fmt.Sprintf("got %+v, want %+v", patches, tc.expectPatches)) {
				for i, gotPatch := range patches {
					if !EqPatch(gotPatch, tc.expectPatches[i]) {
						t.Errorf("Expected patch at position %d to be %+v, got %+v", i, tc.expectPatches[i], gotPatch)
					}
				}
			}
		})
	}
}

func TestGetPatches_TwoReplacementResources(t *testing.T) {
	recommendResources := []vpa_api_util.ContainerResources{
		{
			Requests: core.ResourceList{
				cpu:        resource.MustParse("1"),
				unobtanium: resource.MustParse("2"),
			},
		},
	}
	pod := &core.Pod{
		Spec: core.PodSpec{
			Containers: []core.Container{{
				Resources: core.ResourceRequirements{
					Requests: core.ResourceList{
						cpu: resource.MustParse("0"),
					},
				},
			}},
		},
	}
	recommendAnnotations := vpa_api_util.ContainerToAnnotationsMap{}
	frp := fakeRecommendationProvider{recommendResources, recommendAnnotations, nil}
	c := NewResourceUpdatesCalculator(&frp)
	patches, err := c.CalculatePatches(pod, test.VerticalPodAutoscaler().WithName("name").WithContainer("test").Get())
	assert.NoError(t, err)
	// Order of updates for cpu and unobtanium depends on order of iterating a map, both possible results are valid.
	if assert.Len(t, patches, 3, "unexpected number of patches") {
		cpuUpdate := addResourceRequestPatch(0, cpu, "1")
		unobtaniumUpdate := addResourceRequestPatch(0, unobtanium, "2")
		AssertPatchOneOf(t, patches[0], []resource_admission.PatchRecord{cpuUpdate, unobtaniumUpdate})
		AssertPatchOneOf(t, patches[1], []resource_admission.PatchRecord{cpuUpdate, unobtaniumUpdate})
		assert.False(t, EqPatch(patches[0], patches[1]))
		cpuFirstUnobtaniumSecond := addAnnotationRequest([][]string{{cpu, unobtanium}}, request)
		unobtaniumFirstCpuSecond := addAnnotationRequest([][]string{{unobtanium, cpu}}, request)
		AssertPatchOneOf(t, patches[2], []resource_admission.PatchRecord{cpuFirstUnobtaniumSecond, unobtaniumFirstCpuSecond})
	}
}
