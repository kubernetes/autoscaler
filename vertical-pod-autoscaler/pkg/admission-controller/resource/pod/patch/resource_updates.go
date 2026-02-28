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

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	resource_admission "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/admission-controller/resource"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/admission-controller/resource/pod/recommendation"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/features"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/annotations"
	resourcehelpers "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/resources"
	vpa_api_util "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/vpa"
)

const (
	// ResourceUpdatesAnnotation is the name of annotation
	// containing resource updates performed by VPA.
	ResourceUpdatesAnnotation = "vpaUpdates"
)

type resourcesUpdatesPatchCalculator struct {
	recommendationProvider recommendation.Provider
	maxAllowedCPUBoost     resource.Quantity
}

// NewResourceUpdatesCalculator returns a calculator for
// resource update patches.
func NewResourceUpdatesCalculator(recommendationProvider recommendation.Provider, maxAllowedCPUBoost resource.QuantityValue) Calculator {
	return &resourcesUpdatesPatchCalculator{
		recommendationProvider: recommendationProvider,
		maxAllowedCPUBoost:     maxAllowedCPUBoost.Quantity,
	}
}

func (*resourcesUpdatesPatchCalculator) PatchResourceTarget() PatchResourceTarget {
	return Pod
}

func (c *resourcesUpdatesPatchCalculator) CalculatePatches(pod *corev1.Pod, vpa *vpa_types.VerticalPodAutoscaler) ([]resource_admission.PatchRecord, error) {
	result := []resource_admission.PatchRecord{}

	containersResources, annotationsPerContainer, err := c.recommendationProvider.GetContainersResourcesForPod(pod, vpa)
	if err != nil {
		return []resource_admission.PatchRecord{}, fmt.Errorf("failed to calculate resource patch for pod %s/%s: %v", pod.Namespace, pod.Name, err)
	}

	if vpa_api_util.GetUpdateMode(vpa) == vpa_types.UpdateModeOff {
		// If update mode is "Off", we don't want to apply any recommendations,
		// but we still want to apply startup boost.
		for i := range containersResources {
			containersResources[i].Requests = nil
			containersResources[i].Limits = nil
		}
		annotationsPerContainer = vpa_api_util.ContainerToAnnotationsMap{}
	}

	if annotationsPerContainer == nil {
		annotationsPerContainer = vpa_api_util.ContainerToAnnotationsMap{}
	}

	updatesAnnotation := []string{}
	cpuStartupBoostEnabled := features.Enabled(features.CPUStartupBoost)
	for i := range containersResources {
		// Apply startup boost if configured
		if cpuStartupBoostEnabled {
			// Get the container resource policy to check for scaling mode.
			policy := vpa_api_util.GetContainerResourcePolicy(pod.Spec.Containers[i].Name, vpa.Spec.ResourcePolicy)
			if policy != nil && policy.Mode != nil && *policy.Mode == vpa_types.ContainerScalingModeOff {
				continue
			}
			boostPatches, err := c.applyCPUStartupBoost(&pod.Spec.Containers[i], vpa, &containersResources[i])
			if err != nil {
				return nil, err
			}
			result = append(result, boostPatches...)
		}

		newPatches, newUpdatesAnnotation := getContainerPatch(pod, i, annotationsPerContainer, containersResources[i])
		if len(newPatches) > 0 {
			result = append(result, newPatches...)
			updatesAnnotation = append(updatesAnnotation, newUpdatesAnnotation)
		}
	}

	if len(updatesAnnotation) > 0 {
		vpaAnnotationValue := fmt.Sprintf("Pod resources updated by %s: %s", vpa.Name, strings.Join(updatesAnnotation, "; "))
		result = append(result, GetAddAnnotationPatch(ResourceUpdatesAnnotation, vpaAnnotationValue))
	}
	return result, nil
}

func getContainerPatch(pod *corev1.Pod, i int, annotationsPerContainer vpa_api_util.ContainerToAnnotationsMap, containerResources vpa_api_util.ContainerResources) ([]resource_admission.PatchRecord, string) {
	var patches []resource_admission.PatchRecord
	// Add empty resources object if missing.
	requests, limits := resourcehelpers.ContainerRequestsAndLimits(pod.Spec.Containers[i].Name, pod)
	if limits == nil && requests == nil {
		patches = append(patches, GetPatchInitializingEmptyResources(i))
	}

	annotations, found := annotationsPerContainer[pod.Spec.Containers[i].Name]
	if !found {
		annotations = make([]string, 0)
	}

	patches, annotations = appendPatchesAndAnnotations(patches, annotations, requests, i, containerResources.Requests, "requests", "request")
	patches, annotations = appendPatchesAndAnnotations(patches, annotations, limits, i, containerResources.Limits, "limits", "limit")

	updatesAnnotation := fmt.Sprintf("container %d: ", i) + strings.Join(annotations, ", ")
	return patches, updatesAnnotation
}

func appendPatchesAndAnnotations(patches []resource_admission.PatchRecord, annotations []string, current corev1.ResourceList, containerIndex int, resources corev1.ResourceList, fieldName, resourceName string) ([]resource_admission.PatchRecord, []string) {
	// Add empty object if it's missing and we're about to fill it.
	if current == nil && len(resources) > 0 {
		patches = append(patches, GetPatchInitializingEmptyResourcesSubfield(containerIndex, fieldName))
	}
	for resource, request := range resources {
		patches = append(patches, GetAddResourceRequirementValuePatch(containerIndex, fieldName, resource, request))
		annotations = append(annotations, fmt.Sprintf("%s %s", resource, resourceName))
	}
	return patches, annotations
}

func (c *resourcesUpdatesPatchCalculator) applyCPUStartupBoost(container *corev1.Container, vpa *vpa_types.VerticalPodAutoscaler, containerResources *vpa_api_util.ContainerResources) ([]resource_admission.PatchRecord, error) {
	var patches []resource_admission.PatchRecord

	startupBoostPolicy := getContainerStartupBoostPolicy(container, vpa)
	if startupBoostPolicy == nil {
		return nil, nil
	}

	err := c.applyControlledCPUResources(container, vpa, containerResources, startupBoostPolicy)
	if err != nil {
		return nil, err
	}

	originalResources, err := annotations.GetOriginalResourcesAnnotationValue(container)
	if err != nil {
		return nil, err
	}
	patches = append(patches, GetAddAnnotationPatch(annotations.StartupCPUBoostAnnotation, originalResources))

	return patches, nil
}

func getContainerStartupBoostPolicy(container *corev1.Container, vpa *vpa_types.VerticalPodAutoscaler) *vpa_types.StartupBoost {
	policy := vpa_api_util.GetContainerResourcePolicy(container.Name, vpa.Spec.ResourcePolicy)
	startupBoost := vpa.Spec.StartupBoost
	if policy != nil && policy.StartupBoost != nil {
		startupBoost = policy.StartupBoost
	}
	return startupBoost
}

func (c *resourcesUpdatesPatchCalculator) calculateBoostedCPUValue(baseCPU resource.Quantity, startupBoost *vpa_types.StartupBoost) (*resource.Quantity, error) {
	boostType := startupBoost.CPU.Type
	if boostType == "" {
		boostType = vpa_types.FactorStartupBoostType
	}

	switch boostType {
	case vpa_types.FactorStartupBoostType:
		if startupBoost.CPU.Factor == nil {
			return nil, errors.New("startupBoost.CPU.Factor is required when Type is Factor or not specified")
		}
		factor := *startupBoost.CPU.Factor
		if factor < 1 {
			return nil, errors.New("boost factor must be >= 1")
		}
		boostedCPUMilli := baseCPU.MilliValue()
		boostedCPUMilli = int64(float64(boostedCPUMilli) * float64(factor))
		return resource.NewMilliQuantity(boostedCPUMilli, resource.DecimalSI), nil
	case vpa_types.QuantityStartupBoostType:
		if startupBoost.CPU.Quantity == nil {
			return nil, errors.New("startupBoost.CPU.Quantity is required when Type is Quantity")
		}
		quantity := *startupBoost.CPU.Quantity
		boostedCPUMilli := baseCPU.MilliValue() + quantity.MilliValue()
		return resource.NewMilliQuantity(boostedCPUMilli, resource.DecimalSI), nil
	default:
		return nil, fmt.Errorf("unsupported startup boost type: %s", startupBoost.CPU.Type)
	}
}

func (c *resourcesUpdatesPatchCalculator) calculateBoostedCPU(recommendedCPU, originalCPU resource.Quantity, startupBoost *vpa_types.StartupBoost) (*resource.Quantity, error) {
	baseCPU := recommendedCPU
	if baseCPU.IsZero() {
		baseCPU = originalCPU
	}

	if startupBoost == nil {
		return &baseCPU, nil
	}

	boostedCPU, err := c.calculateBoostedCPUValue(baseCPU, startupBoost)
	if err != nil {
		return nil, err
	}

	if !c.maxAllowedCPUBoost.IsZero() && boostedCPU.Cmp(c.maxAllowedCPUBoost) > 0 {
		return &c.maxAllowedCPUBoost, nil
	}
	return boostedCPU, nil
}

// capStartupBoostToContainerLimit makes sure startup boost recommendation is not above current limit for the container for CPU.
// It attempts to keep the request 1m below the limit to maintain QoS.
func capStartupBoostToContainerLimit(recommendation corev1.ResourceList, containerLimits corev1.ResourceList) {
	limit, found := containerLimits[corev1.ResourceCPU]
	if !found {
		return
	}

	recommendedValue, found := recommendation[corev1.ResourceCPU]
	if found && recommendedValue.MilliValue() > limit.MilliValue() {
		newRecommended := limit.DeepCopy()
		if limit.Cmp(resource.MustParse("1m")) > 0 {
			newRecommended.Sub(resource.MustParse("1m"))
		}
		recommendation[corev1.ResourceCPU] = newRecommended
	}
}

func (c *resourcesUpdatesPatchCalculator) applyControlledCPUResources(container *corev1.Container, vpa *vpa_types.VerticalPodAutoscaler, containerResources *vpa_api_util.ContainerResources, startupBoostPolicy *vpa_types.StartupBoost) error {
	controlledValues := vpa_api_util.GetContainerControlledValues(container.Name, vpa.Spec.ResourcePolicy)

	recommendedRequest := containerResources.Requests[corev1.ResourceCPU]
	originalRequest := container.Resources.Requests[corev1.ResourceCPU]
	boostedRequest, err := c.calculateBoostedCPU(recommendedRequest, originalRequest, startupBoostPolicy)
	if err != nil {
		return err
	}

	if containerResources.Requests == nil {
		containerResources.Requests = corev1.ResourceList{}
	}
	containerResources.Requests[corev1.ResourceCPU] = *boostedRequest

	switch controlledValues {
	case vpa_types.ContainerControlledValuesRequestsOnly:
		capStartupBoostToContainerLimit(containerResources.Requests, container.Resources.Limits)
	case vpa_types.ContainerControlledValuesRequestsAndLimits:
		if containerResources.Limits == nil {
			containerResources.Limits = corev1.ResourceList{}
		}
		newLimits, _ := vpa_api_util.GetProportionalLimit(
			container.Resources.Limits,                               // originalLimits
			container.Resources.Requests,                             // originalRequests
			corev1.ResourceList{corev1.ResourceCPU: *boostedRequest}, // newRequests
			corev1.ResourceList{},                                    // defaultLimit
		)
		if newLimit, ok := newLimits[corev1.ResourceCPU]; ok {
			containerResources.Limits[corev1.ResourceCPU] = newLimit
		}
	default:
		// Do nothing
	}
	return nil
}
