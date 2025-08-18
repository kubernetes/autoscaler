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
	"k8s.io/klog/v2"

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
	maxAllowedCpu          resource.Quantity
}

// NewResourceUpdatesCalculator returns a calculator for
// resource update patches.
func NewResourceUpdatesCalculator(recommendationProvider recommendation.Provider, maxAllowedCpu string) Calculator {
	var maxAllowedCpuQuantity resource.Quantity
	if maxAllowedCpu != "" {
		maxAllowedCpuQuantity = resource.MustParse(maxAllowedCpu)
	}
	return &resourcesUpdatesPatchCalculator{
		recommendationProvider: recommendationProvider,
		maxAllowedCpu:          maxAllowedCpuQuantity,
	}
}

func (*resourcesUpdatesPatchCalculator) PatchResourceTarget() PatchResourceTarget {
	return Pod
}

func (c *resourcesUpdatesPatchCalculator) CalculatePatches(pod *core.Pod, vpa *vpa_types.VerticalPodAutoscaler) ([]resource_admission.PatchRecord, error) {
	klog.Infof("Calculating patches for pod %s/%s with VPA %s", pod.Namespace, pod.Name, vpa.Name)
	result := []resource_admission.PatchRecord{}

	containersResources, annotationsPerContainer, err := c.recommendationProvider.GetContainersResourcesForPod(pod, vpa)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate resource patch for pod %s/%s: %v", pod.Namespace, pod.Name, err)
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
	for i, containerResources := range containersResources {
		// Apply startup boost if configured
		if features.Enabled(features.CPUStartupBoost) {
			policy := vpa_api_util.GetContainerResourcePolicy(pod.Spec.Containers[i].Name, vpa.Spec.ResourcePolicy)
			if policy != nil && policy.Mode != nil && *policy.Mode == vpa_types.ContainerScalingModeOff {
				klog.V(4).Infof("Not applying startup boost for container %s since its scaling mode is Off", pod.Spec.Containers[i].Name)
				continue
			} else {
				boost, err := getStartupBoost(&pod.Spec.Containers[i], vpa)
				if err != nil {
					return nil, err
				}
				if boost != nil {
					if !c.maxAllowedCpu.IsZero() && boost.Cmp(c.maxAllowedCpu) > 0 {
						cappedBoost := c.maxAllowedCpu
						boost = &cappedBoost
					}
					if containerResources.Requests == nil {
						containerResources.Requests = core.ResourceList{}
					}
					containerResources.Requests[core.ResourceCPU] = *boost
					if containerResources.Limits == nil {
						containerResources.Limits = core.ResourceList{}
					}
					containerResources.Limits[core.ResourceCPU] = *boost
					originalResources, err := annotations.GetOriginalResourcesAnnotationValue(&pod.Spec.Containers[i])
					if err != nil {
						return nil, err
					}
					result = append(result, GetAddAnnotationPatch(annotations.StartupCPUBoostAnnotation, originalResources))
				}
			}
		}

		newPatches, newUpdatesAnnotation := getContainerPatch(pod, i, annotationsPerContainer, containerResources)
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

func getStartupBoost(container *core.Container, vpa *vpa_types.VerticalPodAutoscaler) (*resource.Quantity, error) {
	policy := vpa_api_util.GetContainerResourcePolicy(container.Name, vpa.Spec.ResourcePolicy)
	startupBoost := vpa.Spec.StartupBoost
	if policy != nil && policy.StartupBoost != nil {
		startupBoost = policy.StartupBoost
	}
	if startupBoost == nil {
		return nil, nil
	}

	cpuRequest := container.Resources.Requests[core.ResourceCPU]
	boostType := startupBoost.CPU.Type
	if boostType == "" {
		boostType = vpa_types.FactorStartupBoostType
	}

	switch boostType {
	case vpa_types.FactorStartupBoostType:
		if startupBoost.CPU.Factor == nil {
			return nil, fmt.Errorf("startupBoost.CPU.Factor is required when Type is Factor or not specified")
		}
		factor := *startupBoost.CPU.Factor
		if factor < 1 {
			return nil, fmt.Errorf("boost factor must be >= 1")
		}
		boostedCPU := cpuRequest.MilliValue()
		boostedCPU = int64(float64(boostedCPU) * float64(factor))
		return resource.NewMilliQuantity(boostedCPU, resource.DecimalSI), nil
	case vpa_types.QuantityStartupBoostType:
		if startupBoost.CPU.Quantity == nil {
			return nil, fmt.Errorf("startupBoost.CPU.Quantity is required when Type is Quantity")
		}
		quantity := *startupBoost.CPU.Quantity
		if quantity.Cmp(cpuRequest) < 0 {
			return nil, fmt.Errorf("boost quantity %s is less than container's request %s", quantity.String(), cpuRequest.String())
		}
		return &quantity, nil
	default:
		return nil, fmt.Errorf("unsupported startup boost type: %s", startupBoost.CPU.Type)
	}
}

func getContainerPatch(pod *core.Pod, i int, annotationsPerContainer vpa_api_util.ContainerToAnnotationsMap, containerResources vpa_api_util.ContainerResources) ([]resource_admission.PatchRecord, string) {
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

func appendPatchesAndAnnotations(patches []resource_admission.PatchRecord, annotations []string, current core.ResourceList, containerIndex int, resources core.ResourceList, fieldName, resourceName string) ([]resource_admission.PatchRecord, []string) {
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
