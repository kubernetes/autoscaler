/*
Copyright 2019 The Kubernetes Authors.

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

package pod

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/api/admission/v1beta1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	resource_admission "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/admission-controller/resource"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/admission-controller/resource/pod/patch"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/annotations"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/test"
	vpa_api_util "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/vpa"
)

const (
	cpu        = "cpu"
	unobtanium = "unobtanium"
	limit      = "limit"
	request    = "request"
)

// TODO(bskiba): split these tests into correct functional tests
// of patch calculators.

type fakePodPreProcessor struct {
	e error
}

func (fpp *fakePodPreProcessor) Process(pod apiv1.Pod) (apiv1.Pod, error) {
	return pod, fpp.e
}

type fakeRecommendationProvider struct {
	resources              []vpa_api_util.ContainerResources
	containerToAnnotations vpa_api_util.ContainerToAnnotationsMap
	e                      error
}

func (frp *fakeRecommendationProvider) GetContainersResourcesForPod(pod *apiv1.Pod, vpa *vpa_types.VerticalPodAutoscaler) ([]vpa_api_util.ContainerResources, vpa_api_util.ContainerToAnnotationsMap, error) {
	return frp.resources, frp.containerToAnnotations, frp.e
}

type fakeVpaMatcher struct{}

func (m fakeVpaMatcher) GetMatchingVPA(pod *apiv1.Pod) *vpa_types.VerticalPodAutoscaler {
	return test.VerticalPodAutoscaler().WithName("name").WithContainer("testy-container").Get()
}

func addResourcesPatch(idx int) resource_admission.PatchRecord {
	return resource_admission.PatchRecord{
		"add",
		fmt.Sprintf("/spec/containers/%d/resources", idx),
		apiv1.ResourceRequirements{},
	}
}

func addRequestsPatch(idx int) resource_admission.PatchRecord {
	return resource_admission.PatchRecord{
		"add",
		fmt.Sprintf("/spec/containers/%d/resources/requests", idx),
		apiv1.ResourceList{},
	}
}

func addLimitsPatch(idx int) resource_admission.PatchRecord {
	return resource_admission.PatchRecord{
		"add",
		fmt.Sprintf("/spec/containers/%d/resources/limits", idx),
		apiv1.ResourceList{},
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
	return patch.GetAddAnnotationPatch(patch.ResourceUpdatesAnnotation, vpaUpdates)
}

func addVpaObservedContainersPatch(conetinerNames []string) resource_admission.PatchRecord {
	return patch.GetAddAnnotationPatch(
		annotations.VpaObservedContainersLabel,
		strings.Join(conetinerNames, ", "),
	)
}

func assertPatchOneOf(t *testing.T, got resource_admission.PatchRecord, want []resource_admission.PatchRecord) {
	for _, wanted := range want {
		if patch.EqPatch(got, wanted) {
			return
		}
	}
	msg := fmt.Sprintf("got: %+v, expected one of %+v", got, want)
	assert.Fail(t, msg)
}

func TestGetPatches(t *testing.T) {
	tests := []struct {
		name                 string
		podJson              []byte
		namespace            string
		podPreProcessorError error
		recommendResources   []vpa_api_util.ContainerResources
		recommendAnnotations vpa_api_util.ContainerToAnnotationsMap
		recommendName        string
		recommendError       error
		expectPatches        []resource_admission.PatchRecord
		expectError          error
	}{
		{
			name:                 "invalid JSON",
			podJson:              []byte("{"),
			namespace:            "default",
			podPreProcessorError: nil,
			recommendResources:   []vpa_api_util.ContainerResources{},
			recommendAnnotations: vpa_api_util.ContainerToAnnotationsMap{},
			expectError:          fmt.Errorf("unexpected end of JSON input"),
		},
		{
			name:                 "invalid pod",
			podJson:              []byte("{}"),
			namespace:            "default",
			podPreProcessorError: fmt.Errorf("bad pod"),
			recommendResources:   []vpa_api_util.ContainerResources{},
			recommendAnnotations: vpa_api_util.ContainerToAnnotationsMap{},
			expectError:          fmt.Errorf("bad pod"),
		},
		{
			name: "new cpu recommendation",
			podJson: []byte(
				`{
					"spec": {
						"containers": [{}]
					}
				}`),
			namespace: "default",
			recommendResources: []vpa_api_util.ContainerResources{
				{
					Requests: apiv1.ResourceList{
						cpu: resource.MustParse("1"),
					},
				},
			},
			recommendAnnotations: vpa_api_util.ContainerToAnnotationsMap{},
			expectPatches: []resource_admission.PatchRecord{
				patch.GetAddEmptyAnnotationsPatch(),
				addResourcesPatch(0),
				addRequestsPatch(0),
				addResourceRequestPatch(0, cpu, "1"),
				addAnnotationRequest([][]string{{cpu}}, request),
				addVpaObservedContainersPatch([]string{}),
			},
		},
		{
			name: "replacement cpu recommendation",
			podJson: []byte(
				`{
					"spec": {
						"containers": [
							{
								"resources": {
									"requests": {
										"cpu": "0"
									}
								}
							}
						]
					}
				}`),
			namespace: "default",
			recommendResources: []vpa_api_util.ContainerResources{
				{
					Requests: apiv1.ResourceList{
						cpu: resource.MustParse("1"),
					},
				},
			},
			recommendAnnotations: vpa_api_util.ContainerToAnnotationsMap{},
			expectPatches: []resource_admission.PatchRecord{
				patch.GetAddEmptyAnnotationsPatch(),
				addResourceRequestPatch(0, cpu, "1"),
				addAnnotationRequest([][]string{{cpu}}, request),
				addVpaObservedContainersPatch([]string{}),
			},
		},
		{
			name: "two containers",
			podJson: []byte(
				`{
					"spec": {
						"containers": [
							{
								"resources": {
									"requests": {
										"cpu": "0"
									}
								}
							},
							{}
						]
					}
				}`),
			namespace: "default",
			recommendResources: []vpa_api_util.ContainerResources{
				{
					Requests: apiv1.ResourceList{
						cpu: resource.MustParse("1"),
					},
				},
				{
					Requests: apiv1.ResourceList{
						cpu: resource.MustParse("2"),
					},
				},
			},
			recommendAnnotations: vpa_api_util.ContainerToAnnotationsMap{},
			expectPatches: []resource_admission.PatchRecord{
				patch.GetAddEmptyAnnotationsPatch(),
				addResourceRequestPatch(0, cpu, "1"),
				addResourcesPatch(1),
				addRequestsPatch(1),
				addResourceRequestPatch(1, cpu, "2"),
				addAnnotationRequest([][]string{{cpu}, {cpu}}, request),
				addVpaObservedContainersPatch([]string{"", ""}),
			},
		},
		{
			name: "new cpu limit",
			podJson: []byte(
				`{
					"spec": {
						"containers": [{}]
					}
				}`),
			namespace: "default",
			recommendResources: []vpa_api_util.ContainerResources{
				{
					Limits: apiv1.ResourceList{
						cpu: resource.MustParse("1"),
					},
				},
			},
			recommendAnnotations: vpa_api_util.ContainerToAnnotationsMap{},
			expectPatches: []resource_admission.PatchRecord{
				patch.GetAddEmptyAnnotationsPatch(),
				addResourcesPatch(0),
				addLimitsPatch(0),
				addResourceLimitPatch(0, cpu, "1"),
				addAnnotationRequest([][]string{{cpu}}, limit),
				addVpaObservedContainersPatch([]string{}),
			},
		},
		{
			name: "replacement cpu limit",
			podJson: []byte(
				`{
					"spec": {
						"containers": [
							{
								"resources": {
									"limits": {
										"cpu": "0"
									}
								}
							}
						]
					}
				}`),
			namespace: "default",
			recommendResources: []vpa_api_util.ContainerResources{
				{
					Limits: apiv1.ResourceList{
						cpu: resource.MustParse("1"),
					},
				},
			},
			recommendAnnotations: vpa_api_util.ContainerToAnnotationsMap{},
			expectPatches: []resource_admission.PatchRecord{
				patch.GetAddEmptyAnnotationsPatch(),
				addResourceLimitPatch(0, cpu, "1"),
				addAnnotationRequest([][]string{{cpu}}, limit),
				addVpaObservedContainersPatch([]string{}),
			},
		},
	}
	for _, tc := range tests {
		t.Run(fmt.Sprintf("test case: %s", tc.name), func(t *testing.T) {
			fppp := fakePodPreProcessor{e: tc.podPreProcessorError}
			fvm := fakeVpaMatcher{}
			frp := fakeRecommendationProvider{tc.recommendResources, tc.recommendAnnotations, tc.recommendError}
			cs := []patch.Calculator{patch.NewResourceUpdatesCalculator(&frp), patch.NewObservedContainersCalculator()}
			h := NewResourceHandler(&fppp, &fvm, cs)
			patches, err := h.GetPatches(&v1beta1.AdmissionRequest{
				Resource: v1.GroupVersionResource{
					Version: "v1",
				},
				Namespace: tc.namespace,
				Object: runtime.RawExtension{
					Raw: tc.podJson,
				},
			})
			if tc.expectError == nil {
				assert.NoError(t, err)
			} else {
				if assert.Error(t, err) {
					assert.Equal(t, tc.expectError.Error(), err.Error())
				}
			}
			if assert.Equal(t, len(tc.expectPatches), len(patches), fmt.Sprintf("got %+v, want %+v", patches, tc.expectPatches)) {
				for i, gotPatch := range patches {
					if !patch.EqPatch(gotPatch, tc.expectPatches[i]) {
						t.Errorf("Expected patch at position %d to be %+v, got %+v", i, tc.expectPatches[i], gotPatch)
					}
				}
			}
		})
	}
}

func TestGetPatches_TwoReplacementResources(t *testing.T) {
	fppp := fakePodPreProcessor{}
	fvm := fakeVpaMatcher{}
	recommendResources := []vpa_api_util.ContainerResources{
		{
			Requests: apiv1.ResourceList{
				cpu:        resource.MustParse("1"),
				unobtanium: resource.MustParse("2"),
			},
		},
	}
	podJson := []byte(
		`{
					"spec": {
						"containers": [
							{
								"resources": {
									"requests": {
										"cpu": "0"
									}
								}
							}
						]
					}
				}`)
	recommendAnnotations := vpa_api_util.ContainerToAnnotationsMap{}
	frp := fakeRecommendationProvider{recommendResources, recommendAnnotations, nil}
	cs := []patch.Calculator{patch.NewResourceUpdatesCalculator(&frp), patch.NewObservedContainersCalculator()}
	h := NewResourceHandler(&fppp, &fvm, cs)
	patches, err := h.GetPatches(&v1beta1.AdmissionRequest{
		Namespace: "default",
		Resource: v1.GroupVersionResource{
			Version: "v1",
		},
		Object: runtime.RawExtension{
			Raw: podJson,
		},
	})
	assert.NoError(t, err)
	// Order of updates for cpu and unobtanium depends on order of iterating a map, both possible results are valid.
	if assert.Equal(t, len(patches), 5) {
		cpuUpdate := addResourceRequestPatch(0, cpu, "1")
		unobtaniumUpdate := addResourceRequestPatch(0, unobtanium, "2")
		patch.AssertEqPatch(t, patches[0], patch.GetAddEmptyAnnotationsPatch())
		assertPatchOneOf(t, patches[1], []resource_admission.PatchRecord{cpuUpdate, unobtaniumUpdate})
		assertPatchOneOf(t, patches[2], []resource_admission.PatchRecord{cpuUpdate, unobtaniumUpdate})
		assert.False(t, patch.EqPatch(patches[1], patches[2]))
		cpuFirstUnobtaniumSecond := addAnnotationRequest([][]string{{cpu, unobtanium}}, request)
		unobtaniumFirstCpuSecond := addAnnotationRequest([][]string{{unobtanium, cpu}}, request)
		assertPatchOneOf(t, patches[3], []resource_admission.PatchRecord{cpuFirstUnobtaniumSecond, unobtaniumFirstCpuSecond})
		patch.AssertEqPatch(t, patches[4], addVpaObservedContainersPatch([]string{}))
	}
}
