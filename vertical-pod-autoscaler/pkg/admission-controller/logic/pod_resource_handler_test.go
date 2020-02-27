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

package logic

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/api/admission/v1beta1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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

func addResourcesPatch(idx int) patchRecord {
	return patchRecord{
		"add",
		fmt.Sprintf("/spec/containers/%d/resources", idx),
		apiv1.ResourceRequirements{},
	}
}

func addRequestsPatch(idx int) patchRecord {
	return patchRecord{
		"add",
		fmt.Sprintf("/spec/containers/%d/resources/requests", idx),
		apiv1.ResourceList{},
	}
}

func addLimitsPatch(idx int) patchRecord {
	return patchRecord{
		"add",
		fmt.Sprintf("/spec/containers/%d/resources/limits", idx),
		apiv1.ResourceList{},
	}
}

func addResourceRequestPatch(index int, res, amount string) patchRecord {
	return patchRecord{
		"add",
		fmt.Sprintf("/spec/containers/%d/resources/requests/%s", index, res),
		resource.MustParse(amount),
	}
}

func addResourceLimitPatch(index int, res, amount string) patchRecord {
	return patchRecord{
		"add",
		fmt.Sprintf("/spec/containers/%d/resources/limits/%s", index, res),
		resource.MustParse(amount),
	}
}

func addAnnotationRequest(updateResources [][]string, kind string) patchRecord {
	requests := make([]string, 0)
	for idx, podResources := range updateResources {
		podRequests := make([]string, 0)
		for _, resource := range podResources {
			podRequests = append(podRequests, resource+" "+kind)
		}
		requests = append(requests, fmt.Sprintf("container %d: %s", idx, strings.Join(podRequests, ", ")))
	}

	vpaUpdates := fmt.Sprintf("Pod resources updated by name: %s", strings.Join(requests, "; "))
	return getAddAnnotationPatch(vpaAnnotationLabel, vpaUpdates)
}

func addVpaObservedContainersPatch(conetinerNames []string) patchRecord {
	return getAddAnnotationPatch(
		annotations.VpaObservedContainersLabel,
		strings.Join(conetinerNames, ", "),
	)
}

func eqPatch(a, b patchRecord) bool {
	aJson, aErr := json.Marshal(a)
	bJson, bErr := json.Marshal(b)
	return string(aJson) == string(bJson) && aErr == bErr
}

func assertEqPatch(t *testing.T, got, want patchRecord) {
	assert.True(t, eqPatch(got, want), "got %+v, want: %+v", got, want)
}

func assertPatchOneOf(t *testing.T, got patchRecord, want []patchRecord) {
	for _, wanted := range want {
		if eqPatch(got, wanted) {
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
		expectPatches        []patchRecord
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
			expectPatches: []patchRecord{
				addResourcesPatch(0),
				addRequestsPatch(0),
				addResourceRequestPatch(0, cpu, "1"),
				getAddEmptyAnnotationsPatch(),
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
			expectPatches: []patchRecord{
				addResourceRequestPatch(0, cpu, "1"),
				getAddEmptyAnnotationsPatch(),
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
			expectPatches: []patchRecord{
				addResourceRequestPatch(0, cpu, "1"),
				addResourcesPatch(1),
				addRequestsPatch(1),
				addResourceRequestPatch(1, cpu, "2"),
				getAddEmptyAnnotationsPatch(),
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
			expectPatches: []patchRecord{
				addResourcesPatch(0),
				addLimitsPatch(0),
				addResourceLimitPatch(0, cpu, "1"),
				getAddEmptyAnnotationsPatch(),
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
			expectPatches: []patchRecord{
				addResourceLimitPatch(0, cpu, "1"),
				getAddEmptyAnnotationsPatch(),
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
			h := newPodResourceHandler(&fppp, &frp, &fvm)
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
					if !eqPatch(gotPatch, tc.expectPatches[i]) {
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
	h := newPodResourceHandler(&fppp, &frp, &fvm)
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
		assertPatchOneOf(t, patches[0], []patchRecord{cpuUpdate, unobtaniumUpdate})
		assertPatchOneOf(t, patches[1], []patchRecord{cpuUpdate, unobtaniumUpdate})
		assert.False(t, eqPatch(patches[0], patches[1]))
		assertEqPatch(t, patches[2], getAddEmptyAnnotationsPatch())
		cpuFirstUnobtaniumSecond := addAnnotationRequest([][]string{{cpu, unobtanium}}, request)
		unobtaniumFirstCpuSecond := addAnnotationRequest([][]string{{unobtanium, cpu}}, request)
		assertPatchOneOf(t, patches[3], []patchRecord{cpuFirstUnobtaniumSecond, unobtaniumFirstCpuSecond})
		assertEqPatch(t, patches[4], addVpaObservedContainersPatch([]string{}))
	}
}

func TestGetPatches_VpaObservedContainers(t *testing.T) {
	tests := []struct {
		name          string
		podJson       []byte
		expectPatches []patchRecord
	}{
		{
			name: "create vpa observed containers annotation",
			podJson: []byte(
				`{
					"spec": {
						"containers": [
							{
								"Name": "test1"
							},
							{
								"Name": "test2"
							}
						]
					}
				}`),
			expectPatches: []patchRecord{
				getAddEmptyAnnotationsPatch(),
				addVpaObservedContainersPatch([]string{"test1", "test2"}),
			},
		},
		{
			name: "create vpa observed containers annotation with no containers",
			podJson: []byte(
				`{
					"spec": {
						"containers": []
					}
				}`),
			expectPatches: []patchRecord{
				getAddEmptyAnnotationsPatch(),
				addVpaObservedContainersPatch([]string{}),
			},
		},
	}
	for _, tc := range tests {
		t.Run(fmt.Sprintf("test case: %s", tc.name), func(t *testing.T) {
			fppp := fakePodPreProcessor{}
			fvm := fakeVpaMatcher{}
			frp := fakeRecommendationProvider{[]vpa_api_util.ContainerResources{}, vpa_api_util.ContainerToAnnotationsMap{}, nil}
			h := newPodResourceHandler(&fppp, &frp, &fvm)
			patches, err := h.GetPatches(&v1beta1.AdmissionRequest{
				Namespace: "default",
				Resource: v1.GroupVersionResource{
					Version: "v1",
				},
				Object: runtime.RawExtension{
					Raw: tc.podJson,
				},
			})
			assert.NoError(t, err)
			if assert.Len(t, patches, len(tc.expectPatches)) {
				for i, gotPatch := range patches {
					if !eqPatch(gotPatch, tc.expectPatches[i]) {
						t.Errorf("Expected patch at position %d to be %+v, got %+v", i, tc.expectPatches[i], gotPatch)
					}
				}
			}
		})
	}
}
