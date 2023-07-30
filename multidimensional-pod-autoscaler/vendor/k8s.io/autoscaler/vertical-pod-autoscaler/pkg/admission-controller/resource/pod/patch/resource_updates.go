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

	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	resource_admission "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/admission-controller/resource"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/admission-controller/resource/pod/recommendation"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	vpa_api_util "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/vpa"
)

const (
	// ResourceUpdatesAnnotation is the name of annotation
	// containing resource updates performed by VPA.
	ResourceUpdatesAnnotation = "vpaUpdates"
)

type resourcesUpdatesPatchCalculator struct {
	recommendationProvider recommendation.Provider
}

// NewResourceUpdatesCalculator returns a calculator for
// resource update patches.
func NewResourceUpdatesCalculator(recommendationProvider recommendation.Provider) Calculator {
	return &resourcesUpdatesPatchCalculator{
		recommendationProvider: recommendationProvider,
	}
}

func (c *resourcesUpdatesPatchCalculator) CalculatePatches(pod *core.Pod, vpa *vpa_types.VerticalPodAutoscaler) ([]resource_admission.PatchRecord, error) {
	result := []resource_admission.PatchRecord{}

	containersResources, annotationsPerContainer, err := c.recommendationProvider.GetContainersResourcesForPod(pod, vpa)
	if err != nil {
		return []resource_admission.PatchRecord{}, fmt.Errorf("Failed to calculate resource patch for pod %v/%v: %v", pod.Namespace, pod.Name, err)
	}

	if annotationsPerContainer == nil {
		annotationsPerContainer = vpa_api_util.ContainerToAnnotationsMap{}
	}

	updatesAnnotation := []string{}
	for i, containerResources := range containersResources {
		newPatches, newUpdatesAnnotation := getContainerPatch(pod, i, annotationsPerContainer, containerResources)
		result = append(result, newPatches...)
		updatesAnnotation = append(updatesAnnotation, newUpdatesAnnotation)
	}

	if len(updatesAnnotation) > 0 {
		vpaAnnotationValue := fmt.Sprintf("Pod resources updated by %s: %s", vpa.Name, strings.Join(updatesAnnotation, "; "))
		result = append(result, GetAddAnnotationPatch(ResourceUpdatesAnnotation, vpaAnnotationValue))
	}
	return result, nil
}

func getContainerPatch(pod *core.Pod, i int, annotationsPerContainer vpa_api_util.ContainerToAnnotationsMap, containerResources vpa_api_util.ContainerResources) ([]resource_admission.PatchRecord, string) {
	var patches []resource_admission.PatchRecord
	// Add empty resources object if missing.
	if pod.Spec.Containers[i].Resources.Limits == nil &&
		pod.Spec.Containers[i].Resources.Requests == nil {
		patches = append(patches, getPatchInitializingEmptyResources(i))
	}

	annotations, found := annotationsPerContainer[pod.Spec.Containers[i].Name]
	if !found {
		annotations = make([]string, 0)
	}

	patches, annotations = appendPatchesAndAnnotations(patches, annotations, pod.Spec.Containers[i].Resources.Requests, i, containerResources.Requests, "requests", "request")
	patches, annotations = appendPatchesAndAnnotations(patches, annotations, pod.Spec.Containers[i].Resources.Limits, i, containerResources.Limits, "limits", "limit")

	updatesAnnotation := fmt.Sprintf("container %d: ", i) + strings.Join(annotations, ", ")
	return patches, updatesAnnotation
}

func appendPatchesAndAnnotations(patches []resource_admission.PatchRecord, annotations []string, current core.ResourceList, containerIndex int, resources core.ResourceList, fieldName, resourceName string) ([]resource_admission.PatchRecord, []string) {
	// Add empty object if it's missing and we're about to fill it.
	if current == nil && len(resources) > 0 {
		patches = append(patches, getPatchInitializingEmptyResourcesSubfield(containerIndex, fieldName))
	}
	for resource, request := range resources {
		patches = append(patches, getAddResourceRequirementValuePatch(containerIndex, fieldName, resource, request))
		annotations = append(annotations, fmt.Sprintf("%s %s", resource, resourceName))
	}
	return patches, annotations
}

func getAddResourceRequirementValuePatch(i int, kind string, resource core.ResourceName, quantity resource.Quantity) resource_admission.PatchRecord {
	return resource_admission.PatchRecord{
		Op:    "add",
		Path:  fmt.Sprintf("/spec/containers/%d/resources/%s/%s", i, kind, resource),
		Value: quantity.String()}
}

func getPatchInitializingEmptyResources(i int) resource_admission.PatchRecord {
	return resource_admission.PatchRecord{
		Op:    "add",
		Path:  fmt.Sprintf("/spec/containers/%d/resources", i),
		Value: core.ResourceRequirements{},
	}
}

func getPatchInitializingEmptyResourcesSubfield(i int, kind string) resource_admission.PatchRecord {
	return resource_admission.PatchRecord{
		Op:    "add",
		Path:  fmt.Sprintf("/spec/containers/%d/resources/%s", i, kind),
		Value: core.ResourceList{},
	}
}
