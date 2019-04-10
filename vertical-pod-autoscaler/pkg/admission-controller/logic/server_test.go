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
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	vpa_api_util "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/vpa"
)

const (
	cpu        = "cpu"
	unobtanium = "unobtanium"
)

type fakePodPreProcessor struct {
	e error
}

func (fpp *fakePodPreProcessor) Process(pod apiv1.Pod) (apiv1.Pod, error) {
	return pod, fpp.e
}

type fakeRecommendationProvider struct {
	resources              []ContainerResources
	containerToAnnotations vpa_api_util.ContainerToAnnotationsMap
	name                   string
	e                      error
}

func (frp *fakeRecommendationProvider) GetContainersResourcesForPod(pod *apiv1.Pod) ([]ContainerResources, vpa_api_util.ContainerToAnnotationsMap, string, error) {
	return frp.resources, frp.containerToAnnotations, frp.name, frp.e
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

func addResourceRequestPatch(index int, res, amount string) patchRecord {
	return patchRecord{
		"add",
		fmt.Sprintf("/spec/containers/%d/resources/requests/%s", index, res),
		resource.MustParse(amount),
	}
}

func addAnnotationRequest(updateResources [][]string) patchRecord {
	requests := make([]string, 0)
	for idx, podResources := range updateResources {
		podRequests := make([]string, 0)
		for _, resource := range podResources {
			podRequests = append(podRequests, resource+" request")
		}
		requests = append(requests, fmt.Sprintf("container %d: %s", idx, strings.Join(podRequests, ", ")))
	}

	vpaUpdates := fmt.Sprintf("Pod resources updated by name: %s", strings.Join(requests, "; "))
	return patchRecord{
		"add",
		"/metadata/annotations",
		map[string]string{
			"vpaUpdates": vpaUpdates,
		},
	}
}

func eqPatch(a, b patchRecord) bool {
	aJson, aErr := json.Marshal(a)
	bJson, bErr := json.Marshal(b)
	return string(aJson) == string(bJson) && aErr == bErr
}

func TestGetPatchesForResourceRequest(t *testing.T) {
	tests := []struct {
		name                 string
		podJson              []byte
		namespace            string
		preProcessorError    error
		recommendResources   []ContainerResources
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
			preProcessorError:    nil,
			recommendResources:   []ContainerResources{},
			recommendAnnotations: vpa_api_util.ContainerToAnnotationsMap{},
			recommendName:        "name",
			expectError:          fmt.Errorf("unexpected end of JSON input"),
		},
		{
			name:                 "invalid pod",
			podJson:              []byte("{}"),
			namespace:            "default",
			preProcessorError:    fmt.Errorf("bad pod"),
			recommendResources:   []ContainerResources{},
			recommendAnnotations: vpa_api_util.ContainerToAnnotationsMap{},
			recommendName:        "name",
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
			recommendResources: []ContainerResources{
				{
					apiv1.ResourceList{
						cpu: resource.MustParse("1"),
					},
				},
			},
			recommendAnnotations: vpa_api_util.ContainerToAnnotationsMap{},
			recommendName:        "name",
			expectPatches: []patchRecord{
				addResourcesPatch(0),
				addRequestsPatch(0),
				addResourceRequestPatch(0, cpu, "1"),
				addAnnotationRequest([][]string{{cpu}}),
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
			recommendResources: []ContainerResources{
				{
					apiv1.ResourceList{
						cpu: resource.MustParse("1"),
					},
				},
			},
			recommendAnnotations: vpa_api_util.ContainerToAnnotationsMap{},
			recommendName:        "name",
			expectPatches: []patchRecord{
				addResourceRequestPatch(0, cpu, "1"),
				addAnnotationRequest([][]string{{cpu}}),
			},
		},
		{
			name: "two replacement resources",
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
			recommendResources: []ContainerResources{
				{
					apiv1.ResourceList{
						cpu:        resource.MustParse("1"),
						unobtanium: resource.MustParse("2"),
					},
				},
			},
			recommendAnnotations: vpa_api_util.ContainerToAnnotationsMap{},
			recommendName:        "name",
			expectPatches: []patchRecord{
				addResourceRequestPatch(0, cpu, "1"),
				addResourceRequestPatch(0, unobtanium, "2"),
				addAnnotationRequest([][]string{{cpu, unobtanium}}),
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
			recommendResources: []ContainerResources{
				{
					apiv1.ResourceList{
						cpu: resource.MustParse("1"),
					},
				},
				{
					apiv1.ResourceList{
						cpu: resource.MustParse("2"),
					},
				},
			},
			recommendAnnotations: vpa_api_util.ContainerToAnnotationsMap{},
			recommendName:        "name",
			expectPatches: []patchRecord{
				addResourceRequestPatch(0, cpu, "1"),
				addResourcesPatch(1),
				addRequestsPatch(1),
				addResourceRequestPatch(1, cpu, "2"),
				addAnnotationRequest([][]string{{cpu}, {cpu}}),
			},
		},
	}
	for _, tc := range tests {
		t.Run(fmt.Sprintf("test case: %s", tc.name), func(t *testing.T) {
			fppp := fakePodPreProcessor{e: tc.preProcessorError}
			frp := fakeRecommendationProvider{tc.recommendResources, tc.recommendAnnotations, tc.recommendName, tc.recommendError}
			s := NewAdmissionServer(&frp, &fppp)
			patches, err := s.getPatchesForPodResourceRequest(tc.podJson, tc.namespace)
			if tc.expectError == nil {
				assert.Nil(t, err)
			} else {
				assert.Errorf(t, err, tc.expectError.Error())
			}
			if assert.Equal(t, len(tc.expectPatches), len(patches)) {
				for i, gotPatch := range patches {
					if !eqPatch(gotPatch, tc.expectPatches[i]) {
						t.Errorf("Expected patch at position %d to be %+v, got %+v", i, tc.expectPatches[i], gotPatch)
					}
				}
			}
		})
	}
}
