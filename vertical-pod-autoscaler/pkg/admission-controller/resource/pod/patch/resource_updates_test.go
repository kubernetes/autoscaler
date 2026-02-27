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
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	featuregatetesting "k8s.io/component-base/featuregate/testing"

	resource_admission "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/admission-controller/resource"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/features"
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

type fakeRecommendationProvider struct {
	resources              []vpa_api_util.ContainerResources
	initResources          []vpa_api_util.ContainerResources
	containerToAnnotations vpa_api_util.ContainerToAnnotationsMap
	e                      error
}

func (frp *fakeRecommendationProvider) GetContainersResourcesForPod(pod *corev1.Pod, vpa *vpa_types.VerticalPodAutoscaler) ([]vpa_api_util.ContainerResources, []vpa_api_util.ContainerResources, vpa_api_util.ContainerToAnnotationsMap, error) {
	return frp.resources, frp.initResources, frp.containerToAnnotations, frp.e
}

func addResourcesPatch(idx int) resource_admission.PatchRecord {
	return addResourcesPatchForPath(ContainersPath, idx)
}

func addResourcesPatchForPath(basePath string, idx int) resource_admission.PatchRecord {
	return resource_admission.PatchRecord{
		Op:    "add",
		Path:  fmt.Sprintf("%s/%d/resources", basePath, idx),
		Value: corev1.ResourceRequirements{},
	}
}

func addRequestsPatch(idx int) resource_admission.PatchRecord {
	return addRequestsPatchForPath(ContainersPath, idx)
}

func addRequestsPatchForPath(basePath string, idx int) resource_admission.PatchRecord {
	return resource_admission.PatchRecord{
		Op:    "add",
		Path:  fmt.Sprintf("%s/%d/resources/requests", basePath, idx),
		Value: corev1.ResourceList{},
	}
}

func addLimitsPatch(idx int) resource_admission.PatchRecord {
	return addLimitsPatchForPath(ContainersPath, idx)
}

func addLimitsPatchForPath(basePath string, idx int) resource_admission.PatchRecord {
	return resource_admission.PatchRecord{
		Op:    "add",
		Path:  fmt.Sprintf("%s/%d/resources/limits", basePath, idx),
		Value: corev1.ResourceList{},
	}
}

func addResourceRequestPatch(index int, res, amount string) resource_admission.PatchRecord {
	return addResourceRequestPatchForPath(ContainersPath, index, res, amount)
}

func addResourceRequestPatchForPath(basePath string, index int, res, amount string) resource_admission.PatchRecord {
	return resource_admission.PatchRecord{
		Op:    "add",
		Path:  fmt.Sprintf("%s/%d/resources/requests/%s", basePath, index, res),
		Value: resource.MustParse(amount),
	}
}

func addResourceLimitPatch(index int, res, amount string) resource_admission.PatchRecord {
	return addResourceLimitPatchForPath(ContainersPath, index, res, amount)
}

func addResourceLimitPatchForPath(basePath string, index int, res, amount string) resource_admission.PatchRecord {
	return resource_admission.PatchRecord{
		Op:    "add",
		Path:  fmt.Sprintf("%s/%d/resources/limits/%s", basePath, index, res),
		Value: resource.MustParse(amount),
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

func TestCalculatePatches_ResourceUpdates(t *testing.T) {
	tests := []struct {
		name                 string
		pod                  *corev1.Pod
		namespace            string
		recommendResources   []vpa_api_util.ContainerResources
		recommendAnnotations vpa_api_util.ContainerToAnnotationsMap
		recommendError       error
		expectPatches        []resource_admission.PatchRecord
		expectError          error
	}{
		{
			name: "new cpu recommendation",
			pod: &corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{}},
				},
			},
			namespace: "default",
			recommendResources: []vpa_api_util.ContainerResources{
				{
					Requests: corev1.ResourceList{
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
			pod: &corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								cpu: resource.MustParse("0"),
							},
						},
					}},
				},
			},
			namespace: "default",
			recommendResources: []vpa_api_util.ContainerResources{
				{
					Requests: corev1.ResourceList{
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
			name: "replacement cpu request recommendation from container status",
			pod: test.Pod().
				AddContainer(corev1.Container{}).
				AddContainerStatus(test.ContainerStatus().
					WithCPURequest(resource.MustParse("0")).Get()).Get(),
			namespace: "default",
			recommendResources: []vpa_api_util.ContainerResources{
				{
					Requests: corev1.ResourceList{
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
			pod: &corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name: "container-1",
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								cpu: resource.MustParse("0"),
							},
						},
					}, {}},
				},
			},
			namespace: "default",
			recommendResources: []vpa_api_util.ContainerResources{
				{
					Requests: corev1.ResourceList{
						cpu: resource.MustParse("1"),
					},
				},
				{
					Requests: corev1.ResourceList{
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
			pod: &corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{}},
				},
			},
			namespace: "default",
			recommendResources: []vpa_api_util.ContainerResources{
				{
					Limits: corev1.ResourceList{
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
			pod: &corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Resources: corev1.ResourceRequirements{
							Limits: corev1.ResourceList{
								cpu: resource.MustParse("0"),
							},
						},
					}},
				},
			},
			namespace: "default",
			recommendResources: []vpa_api_util.ContainerResources{
				{
					Limits: corev1.ResourceList{
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
		{
			name: "replacement cpu limit from container status",
			pod: test.Pod().
				AddContainer(corev1.Container{}).
				AddContainerStatus(test.ContainerStatus().
					WithCPULimit(resource.MustParse("0")).Get()).Get(),
			namespace: "default",
			recommendResources: []vpa_api_util.ContainerResources{
				{
					Limits: corev1.ResourceList{
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
		{
			name: "no recommendation present",
			pod: test.Pod().
				AddContainer(corev1.Container{}).
				AddContainerStatus(test.ContainerStatus().
					WithCPULimit(resource.MustParse("0")).Get()).Get(),
			namespace:            "default",
			recommendResources:   make([]vpa_api_util.ContainerResources, 1),
			recommendAnnotations: vpa_api_util.ContainerToAnnotationsMap{},
			expectPatches:        []resource_admission.PatchRecord{},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			frp := fakeRecommendationProvider{tc.recommendResources, nil, tc.recommendAnnotations, tc.recommendError}
			c := NewResourceUpdatesCalculator(&frp, resource.QuantityValue{})
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
			Requests: corev1.ResourceList{
				cpu:        resource.MustParse("1"),
				unobtanium: resource.MustParse("2"),
			},
		},
	}
	pod := &corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						cpu: resource.MustParse("0"),
					},
				},
			}},
		},
	}
	recommendAnnotations := vpa_api_util.ContainerToAnnotationsMap{}
	frp := fakeRecommendationProvider{recommendResources, nil, recommendAnnotations, nil}
	c := NewResourceUpdatesCalculator(&frp, resource.QuantityValue{})
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

func TestCalculatePatches_StartupBoost(t *testing.T) {
	factor2 := int32(2)
	factor3 := int32(3)
	quantity := resource.MustParse("500m")
	invalidFactor := int32(0)
	invalidQuantity := resource.MustParse("200m")
	tests := []struct {
		name                 string
		pod                  *corev1.Pod
		vpa                  *vpa_types.VerticalPodAutoscaler
		recommendResources   []vpa_api_util.ContainerResources
		recommendAnnotations vpa_api_util.ContainerToAnnotationsMap
		recommendError       error
		maxAllowedCpu        resource.QuantityValue
		expectPatches        []resource_admission.PatchRecord
		expectError          error
		featureGateEnabled   bool
	}{
		{
			name: "startup boost factor",
			pod: &corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "container1",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									cpu: resource.MustParse("1m"),
								},
								Limits: corev1.ResourceList{
									cpu: resource.MustParse("1m"),
								},
							},
						},
					},
				},
			},
			vpa: test.VerticalPodAutoscaler().WithName("name").WithContainer("container1").WithCPUStartupBoost(vpa_types.FactorStartupBoostType, &factor2, nil, 10).Get(),
			recommendResources: []vpa_api_util.ContainerResources{
				{
					Requests: corev1.ResourceList{
						cpu: resource.MustParse("100m"),
					},
					Limits: corev1.ResourceList{
						cpu: resource.MustParse("100m"),
					},
				},
			},
			maxAllowedCpu:      resource.QuantityValue{},
			featureGateEnabled: true,
			expectPatches: []resource_admission.PatchRecord{
				GetAddAnnotationPatch(annotations.StartupCPUBoostAnnotation, "{\"requests\":{\"cpu\":\"1m\"},\"limits\":{\"cpu\":\"1m\"}}"),
				addResourceRequestPatch(0, cpu, "200m"),
				addResourceLimitPatch(0, cpu, "200m"),
				GetAddAnnotationPatch(ResourceUpdatesAnnotation, "Pod resources updated by name: container 0: cpu request, cpu limit"),
			},
		},
		{
			name: "startup boost factor with 0s duration",
			pod: &corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "container1",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									cpu: resource.MustParse("1m"),
								},
								Limits: corev1.ResourceList{
									cpu: resource.MustParse("1m"),
								},
							},
						},
					},
				},
			},
			vpa: test.VerticalPodAutoscaler().WithName("name").WithContainer("container1").WithCPUStartupBoost(vpa_types.FactorStartupBoostType, &factor2, nil, 0).Get(),
			recommendResources: []vpa_api_util.ContainerResources{
				{
					Requests: corev1.ResourceList{
						cpu: resource.MustParse("100m"),
					},
					Limits: corev1.ResourceList{
						cpu: resource.MustParse("100m"),
					},
				},
			},
			maxAllowedCpu:      resource.QuantityValue{},
			featureGateEnabled: true,
			expectPatches: []resource_admission.PatchRecord{
				GetAddAnnotationPatch(annotations.StartupCPUBoostAnnotation, "{\"requests\":{\"cpu\":\"1m\"},\"limits\":{\"cpu\":\"1m\"}}"),
				addResourceRequestPatch(0, cpu, "200m"),
				addResourceLimitPatch(0, cpu, "200m"),
				GetAddAnnotationPatch(ResourceUpdatesAnnotation, "Pod resources updated by name: container 0: cpu request, cpu limit"),
			},
		},
		{
			name: "startup boost quantity",
			pod: &corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "container1",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									cpu: resource.MustParse("1m"),
								},
								Limits: corev1.ResourceList{
									cpu: resource.MustParse("1m"),
								},
							},
						},
					},
				},
			},
			vpa: test.VerticalPodAutoscaler().WithName("name").WithContainer("container1").WithCPUStartupBoost(vpa_types.QuantityStartupBoostType, nil, &quantity, 10).Get(),
			recommendResources: []vpa_api_util.ContainerResources{
				{
					Requests: corev1.ResourceList{
						cpu: resource.MustParse("100m"),
					},
					Limits: corev1.ResourceList{
						cpu: resource.MustParse("100m"),
					},
				},
			},
			maxAllowedCpu:      resource.QuantityValue{},
			featureGateEnabled: true,
			expectPatches: []resource_admission.PatchRecord{
				GetAddAnnotationPatch(annotations.StartupCPUBoostAnnotation, "{\"requests\":{\"cpu\":\"1m\"},\"limits\":{\"cpu\":\"1m\"}}"),
				addResourceRequestPatch(0, cpu, "600m"),
				addResourceLimitPatch(0, cpu, "600m"),
				GetAddAnnotationPatch(ResourceUpdatesAnnotation, "Pod resources updated by name: container 0: cpu request, cpu limit"),
			},
		},
		{
			name: "feature gate disabled",
			pod: &corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "container1",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									cpu: resource.MustParse("100m"),
								},
							},
						},
					},
				},
			},
			vpa: test.VerticalPodAutoscaler().WithName("name").WithContainer("container1").WithCPUStartupBoost(vpa_types.FactorStartupBoostType, &factor2, nil, 10).Get(),
			recommendResources: []vpa_api_util.ContainerResources{
				{
					Requests: corev1.ResourceList{
						cpu: resource.MustParse("100m"),
					},
				},
			},
			maxAllowedCpu:      resource.QuantityValue{},
			featureGateEnabled: false,
			expectPatches: []resource_admission.PatchRecord{
				addResourceRequestPatch(0, cpu, "100m"),
				addAnnotationRequest([][]string{{cpu}}, "request"),
			},
		},
		{
			name: "invalid factor",
			pod: &corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "container1",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									cpu: resource.MustParse("1m"),
								},
							},
						},
					},
				},
			},
			vpa: test.VerticalPodAutoscaler().WithName("name").WithContainer("container1").WithCPUStartupBoost(vpa_types.FactorStartupBoostType, &invalidFactor, nil, 10).Get(),
			recommendResources: []vpa_api_util.ContainerResources{
				{
					Requests: corev1.ResourceList{
						cpu: resource.MustParse("100m"),
					},
				},
			},
			maxAllowedCpu:      resource.QuantityValue{},
			featureGateEnabled: true,
			expectError:        errors.New("boost factor must be >= 1"),
		},
		{
			name: "quantity less than request",
			pod: &corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "container1",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									cpu: resource.MustParse("400m"),
								},
								Limits: corev1.ResourceList{
									cpu: resource.MustParse("400m"),
								},
							},
						},
					},
				},
			},
			vpa: test.VerticalPodAutoscaler().WithName("name").WithContainer("container1").WithCPUStartupBoost(vpa_types.QuantityStartupBoostType, nil, &invalidQuantity, 10).Get(),
			recommendResources: []vpa_api_util.ContainerResources{
				{
					Requests: corev1.ResourceList{
						cpu: resource.MustParse("500m"),
					},
					Limits: corev1.ResourceList{
						cpu: resource.MustParse("500m"),
					},
				},
			},
			maxAllowedCpu:      resource.QuantityValue{},
			featureGateEnabled: true,
			expectPatches: []resource_admission.PatchRecord{
				GetAddAnnotationPatch(annotations.StartupCPUBoostAnnotation, "{\"requests\":{\"cpu\":\"400m\"},\"limits\":{\"cpu\":\"400m\"}}"),
				addResourceRequestPatch(0, cpu, "700m"),
				addResourceLimitPatch(0, cpu, "700m"),
				GetAddAnnotationPatch(ResourceUpdatesAnnotation, "Pod resources updated by name: container 0: cpu request, cpu limit"),
			},
		},
		{
			name: "startup boost capped",
			pod: &corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "container1",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									cpu: resource.MustParse("1m"),
								},
								Limits: corev1.ResourceList{
									cpu: resource.MustParse("1m"),
								},
							},
						},
					},
				},
			},
			vpa: test.VerticalPodAutoscaler().WithName("name").WithContainer("container1").WithCPUStartupBoost(vpa_types.FactorStartupBoostType, &factor3, nil, 1).Get(),
			recommendResources: []vpa_api_util.ContainerResources{
				{
					Requests: corev1.ResourceList{
						cpu: resource.MustParse("20m"),
					},
					Limits: corev1.ResourceList{
						cpu: resource.MustParse("20m"),
					},
				},
			},
			maxAllowedCpu:      resource.QuantityValue{Quantity: resource.MustParse("40m")},
			featureGateEnabled: true,
			expectPatches: []resource_admission.PatchRecord{
				GetAddAnnotationPatch(annotations.StartupCPUBoostAnnotation, "{\"requests\":{\"cpu\":\"1m\"},\"limits\":{\"cpu\":\"1m\"}}"),
				addResourceRequestPatch(0, cpu, "40m"),
				addResourceLimitPatch(0, cpu, "40m"),
				GetAddAnnotationPatch(ResourceUpdatesAnnotation, "Pod resources updated by name: container 0: cpu request, cpu limit"),
			},
		},
		{
			name: "startup boost with scaling mode off",
			pod: &corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "container1",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									cpu: resource.MustParse("100m"),
								},
							},
						},
					},
				},
			},
			vpa: test.VerticalPodAutoscaler().WithName("name").WithContainer("container1").WithCPUStartupBoost(vpa_types.FactorStartupBoostType, &factor2, nil, 10).WithScalingMode("container1", vpa_types.ContainerScalingModeOff).Get(),
			recommendResources: []vpa_api_util.ContainerResources{
				{
					Requests: corev1.ResourceList{
						cpu: resource.MustParse("1"),
					},
				},
			},
			maxAllowedCpu:      resource.QuantityValue{},
			featureGateEnabled: true,
			expectPatches:      []resource_admission.PatchRecord{},
		},
		{
			name: "startup boost no recommendation",
			pod: &corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "container1",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									cpu: resource.MustParse("100m"),
								},
								Limits: corev1.ResourceList{
									cpu: resource.MustParse("100m"),
								},
							},
						},
					},
				},
			},
			vpa:                test.VerticalPodAutoscaler().WithName("name").WithContainer("container1").WithCPUStartupBoost(vpa_types.FactorStartupBoostType, &factor2, nil, 10).Get(),
			recommendResources: make([]vpa_api_util.ContainerResources, 1),
			maxAllowedCpu:      resource.QuantityValue{},
			featureGateEnabled: true,
			expectPatches: []resource_admission.PatchRecord{
				GetAddAnnotationPatch(annotations.StartupCPUBoostAnnotation, "{\"requests\":{\"cpu\":\"100m\"},\"limits\":{\"cpu\":\"100m\"}}"),
				addResourceRequestPatch(0, cpu, "200m"),
				addResourceLimitPatch(0, cpu, "200m"),
				GetAddAnnotationPatch(ResourceUpdatesAnnotation, "Pod resources updated by name: container 0: cpu request, cpu limit"),
			},
		},
		{
			name: "startup boost with ControlledValues=RequestsOnly",
			pod: &corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "container1",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									cpu: resource.MustParse("100m"),
								},
								Limits: corev1.ResourceList{
									cpu: resource.MustParse("300m"),
								},
							},
						},
					},
				},
			},
			vpa: test.VerticalPodAutoscaler().WithName("name").WithContainer("container1").WithCPUStartupBoost(vpa_types.FactorStartupBoostType, &factor2, nil, 10).WithControlledValues("container1", vpa_types.ContainerControlledValuesRequestsOnly).Get(),
			recommendResources: []vpa_api_util.ContainerResources{
				{
					Requests: corev1.ResourceList{
						cpu: resource.MustParse("100m"),
					},
				},
			},
			maxAllowedCpu:      resource.QuantityValue{},
			featureGateEnabled: true,
			expectPatches: []resource_admission.PatchRecord{
				GetAddAnnotationPatch(annotations.StartupCPUBoostAnnotation, "{\"requests\":{\"cpu\":\"100m\"},\"limits\":{\"cpu\":\"300m\"}}"),
				addResourceRequestPatch(0, cpu, "200m"),
				GetAddAnnotationPatch(ResourceUpdatesAnnotation, "Pod resources updated by name: container 0: cpu request"),
			},
		},
		{
			name: "startup boost with RequestsOnly - capped below limit to preserve pod Qos",
			pod: &corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "container1",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU: resource.MustParse("100m"),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU: resource.MustParse("150m"),
								},
							},
						},
					},
				},
			},
			vpa: test.VerticalPodAutoscaler().WithName("name").WithContainer("container1").WithCPUStartupBoost(vpa_types.FactorStartupBoostType, &factor2, nil, 10).WithControlledValues("container1", vpa_types.ContainerControlledValuesRequestsOnly).Get(),
			recommendResources: []vpa_api_util.ContainerResources{
				{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU: resource.MustParse("100m"),
					},
				},
			},
			maxAllowedCpu:      resource.QuantityValue{},
			featureGateEnabled: true,
			expectPatches: []resource_admission.PatchRecord{
				GetAddAnnotationPatch(annotations.StartupCPUBoostAnnotation, "{\"requests\":{\"cpu\":\"100m\"},\"limits\":{\"cpu\":\"150m\"}}"),
				addResourceRequestPatch(0, cpu, "149m"),
				GetAddAnnotationPatch(ResourceUpdatesAnnotation, "Pod resources updated by name: container 0: cpu request"),
			},
		},
		{
			name: "startup boost with ControlledValues=RequestsandLimits and limits set",
			pod: &corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "container1",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									cpu: resource.MustParse("100m"),
								},
								Limits: corev1.ResourceList{
									cpu: resource.MustParse("300m"),
								},
							},
						},
					},
				},
			},
			vpa: test.VerticalPodAutoscaler().WithName("name").WithContainer("container1").WithCPUStartupBoost(vpa_types.FactorStartupBoostType, &factor2, nil, 10).WithControlledValues("container1", vpa_types.ContainerControlledValuesRequestsAndLimits).Get(),
			recommendResources: []vpa_api_util.ContainerResources{
				{
					Requests: corev1.ResourceList{
						cpu: resource.MustParse("150m"),
					},
				},
			},
			maxAllowedCpu:      resource.QuantityValue{},
			featureGateEnabled: true,
			expectPatches: []resource_admission.PatchRecord{
				GetAddAnnotationPatch(annotations.StartupCPUBoostAnnotation, "{\"requests\":{\"cpu\":\"100m\"},\"limits\":{\"cpu\":\"300m\"}}"),
				addResourceRequestPatch(0, cpu, "300m"),
				addResourceLimitPatch(0, cpu, "900m"),
				GetAddAnnotationPatch(ResourceUpdatesAnnotation, "Pod resources updated by name: container 0: cpu request, cpu limit"),
			},
		},
		{
			name: "startup boost with ControlledValues=RequestsandLimits and limits not set",
			pod: &corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "container1",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									cpu: resource.MustParse("100m"),
								},
							},
						},
					},
				},
			},
			vpa: test.VerticalPodAutoscaler().WithName("name").WithContainer("container1").WithCPUStartupBoost(vpa_types.FactorStartupBoostType, &factor2, nil, 10).WithControlledValues("container1", vpa_types.ContainerControlledValuesRequestsAndLimits).Get(),
			recommendResources: []vpa_api_util.ContainerResources{
				{
					Requests: corev1.ResourceList{
						cpu: resource.MustParse("150m"),
					},
				},
			},
			maxAllowedCpu:      resource.QuantityValue{},
			featureGateEnabled: true,
			expectPatches: []resource_admission.PatchRecord{
				GetAddAnnotationPatch(annotations.StartupCPUBoostAnnotation, "{\"requests\":{\"cpu\":\"100m\"},\"limits\":{}}"),
				addResourceRequestPatch(0, cpu, "300m"),
				GetAddAnnotationPatch(ResourceUpdatesAnnotation, "Pod resources updated by name: container 0: cpu request"),
			},
		},
		{
			name: "startup boost invalid type",
			pod: &corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "container1",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									cpu: resource.MustParse("1m"),
								},
							},
						},
					},
				},
			},
			vpa: test.VerticalPodAutoscaler().WithName("name").WithContainer("container1").WithCPUStartupBoost("Invalid", &factor2, nil, 10).Get(),
			recommendResources: []vpa_api_util.ContainerResources{
				{
					Requests: corev1.ResourceList{
						cpu: resource.MustParse("100m"),
					},
				},
			},
			maxAllowedCpu:      resource.QuantityValue{},
			featureGateEnabled: true,
			expectError:        errors.New("unsupported startup boost type: Invalid"),
		},
		{
			name: "startup boost container policy takes precedence",
			pod: &corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "container1",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									cpu: resource.MustParse("1m"),
								},
								Limits: corev1.ResourceList{
									cpu: resource.MustParse("1m"),
								},
							},
						},
					},
				},
			},
			vpa: test.VerticalPodAutoscaler().WithName("name").WithContainer("container1").
				WithCPUStartupBoost(vpa_types.FactorStartupBoostType, &factor2, nil, 10).
				WithContainerCPUStartupBoost("container1", vpa_types.FactorStartupBoostType, &factor3, nil, 10).Get(),
			recommendResources: []vpa_api_util.ContainerResources{
				{
					Requests: corev1.ResourceList{
						cpu: resource.MustParse("100m"),
					},
					Limits: corev1.ResourceList{
						cpu: resource.MustParse("100m"),
					},
				},
			},
			maxAllowedCpu:      resource.QuantityValue{},
			featureGateEnabled: true,
			expectPatches: []resource_admission.PatchRecord{
				GetAddAnnotationPatch(annotations.StartupCPUBoostAnnotation, "{\"requests\":{\"cpu\":\"1m\"},\"limits\":{\"cpu\":\"1m\"}}"),
				addResourceRequestPatch(0, cpu, "300m"),
				addResourceLimitPatch(0, cpu, "300m"),
				GetAddAnnotationPatch(ResourceUpdatesAnnotation, "Pod resources updated by name: container 0: cpu request, cpu limit"),
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			featuregatetesting.SetFeatureGateDuringTest(t, features.MutableFeatureGate, features.CPUStartupBoost, tc.featureGateEnabled)

			frp := fakeRecommendationProvider{tc.recommendResources, nil, tc.recommendAnnotations, tc.recommendError}
			c := NewResourceUpdatesCalculator(&frp, tc.maxAllowedCpu)
			patches, err := c.CalculatePatches(tc.pod, tc.vpa)
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
