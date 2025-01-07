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

package inplace

import (
	"fmt"

	core "k8s.io/api/core/v1"
	resource_admission "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/admission-controller/resource"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/admission-controller/resource/pod/patch"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/admission-controller/resource/pod/recommendation"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	vpa_api_util "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/vpa"
)

type resourcesInplaceUpdatesPatchCalculator struct {
	recommendationProvider recommendation.Provider
}

// NewResourceInPlaceUpdatesCalculator returns a calculator for
// resource update patches.
func NewResourceInPlaceUpdatesCalculator(recommendationProvider recommendation.Provider) patch.Calculator {
	return &resourcesInplaceUpdatesPatchCalculator{
		recommendationProvider: recommendationProvider,
	}
}

// TODO(maxcao13): this calculator's patches should only be marshalled as a JSON patch to the pod "resize" subresource. it won't be able to patch pod annotations
// if we DO need to calculate patches to add annotations, we can create a separate calculator to add that
func (c *resourcesInplaceUpdatesPatchCalculator) CalculatePatches(pod *core.Pod, vpa *vpa_types.VerticalPodAutoscaler) ([]resource_admission.PatchRecord, error) {
	result := []resource_admission.PatchRecord{}

	containersResources, _, err := c.recommendationProvider.GetContainersResourcesForPod(pod, vpa)
	if err != nil {
		return []resource_admission.PatchRecord{}, fmt.Errorf("Failed to calculate resource patch for pod %s/%s: %v", pod.Namespace, pod.Name, err)
	}

	for i, containerResources := range containersResources {
		newPatches := getContainerPatch(pod, i, containerResources)
		result = append(result, newPatches...)
	}

	return result, nil
}

func getContainerPatch(pod *core.Pod, i int, containerResources vpa_api_util.ContainerResources) []resource_admission.PatchRecord {
	var patches []resource_admission.PatchRecord
	// Add empty resources object if missing.
	if pod.Spec.Containers[i].Resources.Limits == nil &&
		pod.Spec.Containers[i].Resources.Requests == nil {
		patches = append(patches, patch.GetPatchInitializingEmptyResources(i))
	}

	patches = appendPatches(patches, pod.Spec.Containers[i].Resources.Requests, i, containerResources.Requests, "requests")
	patches = appendPatches(patches, pod.Spec.Containers[i].Resources.Limits, i, containerResources.Limits, "limits")

	return patches
}

func appendPatches(patches []resource_admission.PatchRecord, current core.ResourceList, containerIndex int, resources core.ResourceList, fieldName string) []resource_admission.PatchRecord {
	// Add empty object if it's missing and we're about to fill it.
	if current == nil && len(resources) > 0 {
		patches = append(patches, patch.GetPatchInitializingEmptyResourcesSubfield(containerIndex, fieldName))
	}
	for resource, request := range resources {
		patches = append(patches, patch.GetAddResourceRequirementValuePatch(containerIndex, fieldName, resource, request))
	}
	return patches
}
