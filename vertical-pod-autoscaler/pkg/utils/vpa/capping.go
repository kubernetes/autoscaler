/*
Copyright 2018 The Kubernetes Authors.

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

package api

import (
	"errors"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/klog/v2"

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/features"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/limitrange"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/limits"
	resourcehelpers "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/resources"
)

// NewCappingRecommendationProcessor constructs new RecommendationsProcessor that adjusts recommendation
// for given pod to obey VPA resources policy and container limits
func NewCappingRecommendationProcessor(limitsRangeCalculator limitrange.LimitRangeCalculator) RecommendationProcessor {
	return &cappingRecommendationProcessor{limitsRangeCalculator: limitsRangeCalculator}
}

type cappingAction string

var (
	cappedToMinAllowed             cappingAction = "capped to minAllowed"
	cappedToMaxAllowed             cappingAction = "capped to maxAllowed"
	cappedToLimit                  cappingAction = "capped to container limit"
	cappedProportionallyToMaxLimit cappingAction = "capped to fit Max in container LimitRange"
	cappedProportionallyToMinLimit cappingAction = "capped to fit Min in container LimitRange"
	// pod level annotations
	cappedToPodLevelMinAllowed cappingAction = "capped to pod level minAllowed"
	cappedToPodLevelMaxAllowed cappingAction = "capped to pod level maxAllowed"
	cappedToPodLimitRangeMax   cappingAction = "capped to fit Max in Pod LimitRange"
	cappedToPodLimitRangeMin   cappingAction = "capped to fit Min in Pod LimitRange"
)

func toCappingAnnotation(isPodLevel bool, resourceName corev1.ResourceName, action cappingAction) string {
	if isPodLevel {
		return fmt.Sprintf("pod level %s %s", resourceName, action)
	}
	return fmt.Sprintf("%s %s", resourceName, action)
}

type cappingRecommendationProcessor struct {
	limitsRangeCalculator limitrange.LimitRangeCalculator
}

// Apply returns a recommendation for the given pod, adjusted to obey policy and limits.
func (c *cappingRecommendationProcessor) Apply(
	vpa *vpa_types.VerticalPodAutoscaler,
	pod *corev1.Pod) (*vpa_types.RecommendedPodResources, *utils.Annotations, error) {
	// TODO: Annotate if request enforced by maintaining proportion with limit and allowed limit range is in conflict with policy.

	if vpa == nil {
		return nil, nil, errors.New("cannot process nil vpa")
	}
	if pod == nil {
		return nil, nil, errors.New("cannot process nil pod")
	}

	policy := vpa.Spec.ResourcePolicy
	podRecommendation := vpa.Status.Recommendation

	if podRecommendation == nil && policy == nil {
		// If there is no recommendation and no policies have been defined then no recommendation can be computed.
		return nil, nil, nil
	}
	if podRecommendation == nil {
		// Policies have been specified. Create an empty recommendation so that the policies can be applied correctly.
		podRecommendation = new(vpa_types.RecommendedPodResources)
	}

	containerLimitRange, err := c.limitsRangeCalculator.GetContainerLimitRangeItem(pod.Namespace)
	if err != nil {
		klog.V(0).InfoS("Failed to fetch LimitRange for namespace", "namespace", pod.Namespace)
	}

	updatedRecommendations := []vpa_types.RecommendedContainerResources{}
	containerToAnnotationsMap := utils.ContainerToAnnotationsMap{}
	limitAdjustedRecommendation, err := c.capProportionallyToPodLimitRange(podRecommendation.ContainerRecommendations, pod)
	if err != nil {
		return nil, nil, err
	}

	for _, containerRecommendation := range limitAdjustedRecommendation {
		container := getContainer(containerRecommendation.ContainerName, pod)

		if container == nil {
			klog.V(2).InfoS("No matching Container found for recommendation", "containerName", containerRecommendation.ContainerName, "vpa", klog.KObj(vpa))
			continue
		}

		updatedContainerResources, containerAnnotations, err := getCappedRecommendationForContainer(
			pod, *container, &containerRecommendation, policy, containerLimitRange)

		if len(containerAnnotations) != 0 {
			containerToAnnotationsMap[containerRecommendation.ContainerName] = containerAnnotations
		}

		if err != nil {
			return nil, nil, fmt.Errorf("cannot update recommendation for container name %v", container.Name)
		}
		updatedRecommendations = append(updatedRecommendations, *updatedContainerResources)
	}

	var updatedPodRecommendation *vpa_types.PodRecommendations
	annotations := utils.Annotations{
		Container: containerToAnnotationsMap,
		Pod:       nil,
	}

	if features.Enabled(features.VPAPodLevelResources) && IsPodLevelScalingModeEnabled(vpa.Spec.ResourcePolicy) {
		var podPolicyAnnotations []string
		updatedRecommendations, podPolicyAnnotations = capRecommendationByPodConstraints(updatedRecommendations, policy.PodPolicies)
		annotations.Pod = append(annotations.Pod, podPolicyAnnotations...)
		if containerLimitRange != nil {
			klog.V(0).InfoS(
				"Container LimitRange object exists while VPA is configured with Pod-level scaling enabled. It is recommended to remove it from the namespace as described in KEP-2837.",
				"namespace", pod.Namespace,
			)
		}
		updatedPodRecommendation = resourcehelpers.SumContainerLevelRecommendations(updatedRecommendations)
		var podLimitRangeAnnotations []string
		updatedPodRecommendation, podLimitRangeAnnotations, _ = c.capPodLevelRecommendationToPodLimitRange(updatedPodRecommendation, pod)
		annotations.Pod = append(annotations.Pod, podLimitRangeAnnotations...)
	}
	return &vpa_types.RecommendedPodResources{
		ContainerRecommendations: updatedRecommendations,
		PodRecommendations:       updatedPodRecommendation}, &annotations, nil
}

// capPodLevelRecommendationToPodLimitRange processes pod-level recommendations
// and ensures they comply with the Min and Max constraints defined in the LimitRange object for Pod type.
// It also returns annotations that describe any modifications applied to the pod-level recommendations.
func (c *cappingRecommendationProcessor) capPodLevelRecommendationToPodLimitRange(
	podRecommendations *vpa_types.PodRecommendations,
	pod *corev1.Pod,
) (*vpa_types.PodRecommendations, []string, error) {
	podLimitRange, err := c.limitsRangeCalculator.GetPodLimitRangeItem(pod.Namespace)
	if err != nil {
		return nil, nil, err
	}
	if podLimitRange == nil {
		return podRecommendations, nil, nil
	}
	return capPodLevelRecommendationToPodLimitRange(podRecommendations, pod, podLimitRange)
}

func capPodLevelRecommendationToPodLimitRange(
	podRecommendations *vpa_types.PodRecommendations, pod *corev1.Pod, podLimitRange *corev1.LimitRangeItem) (*vpa_types.PodRecommendations, []string, error) {
	if podRecommendations == nil {
		return nil, nil, errors.New("error obtaining pod level recommendations")
	}
	if podLimitRange == nil {
		return podRecommendations, nil, nil
	}

	requests, limits := resourcehelpers.PodRequestsAndLimits(pod)
	cappingAnnotations := make([]string, 0)

	process := func(recommendation corev1.ResourceList, genAnnotations bool) {
		annotations := applyContainerLimitRange(true, recommendation, requests, limits, podLimitRange)
		if genAnnotations {
			cappingAnnotations = append(cappingAnnotations, annotations...)
		}
	}

	process(podRecommendations.LowerBound, false)
	process(podRecommendations.Target, true)
	process(podRecommendations.UpperBound, false)

	return podRecommendations, cappingAnnotations, nil
}

// getCappedRecommendationForContainer returns a recommendation for the given container,
// adjusted to obey the following:
//
// - container-level policies defined in the VPA object (minAllowed and maxAllowed)
//
// - limits specified in a LimitRange object of type Container
func getCappedRecommendationForContainer(
	pod *corev1.Pod,
	container corev1.Container,
	containerRecommendation *vpa_types.RecommendedContainerResources,
	policy *vpa_types.PodResourcePolicy, limitRange *corev1.LimitRangeItem) (*vpa_types.RecommendedContainerResources, []string, error) {
	if containerRecommendation == nil {
		return nil, nil, fmt.Errorf("no recommendation available for container name %v", container.Name)
	}
	// We skip containers with RecommendationOnly mode, as their resource stanzas are not managed,
	// only when the VPAPodLevelResources feature gate is enabled.
	// Otherwise, we fall back from RecommendationOnly container mode to Auto.
	// if features.Enabled(features.VPAPodLevelResources) && IsContainerScalingModeRecsOnly(container.Name, policy) {
	// 	return containerRecommendation, nil, nil
	// }

	// containerPolicy can be nil (user does not have to configure it).
	containerPolicy := GetContainerResourcePolicy(container.Name, policy)
	containerControlledValues := GetContainerControlledValues(container.Name, policy)

	cappedRecommendations := containerRecommendation.DeepCopy()

	cappingAnnotations := make([]string, 0)

	process := func(recommendation corev1.ResourceList, genAnnotations bool) {
		containerRequests, containerLimits := resourcehelpers.ContainerRequestsAndLimits(container.Name, pod)
		limitAnnotations := applyContainerLimitRange(false, recommendation, containerRequests, containerLimits, limitRange)
		annotations := applyVPAPolicy(recommendation, containerPolicy)
		if genAnnotations {
			cappingAnnotations = append(cappingAnnotations, limitAnnotations...)
			cappingAnnotations = append(cappingAnnotations, annotations...)
		}
		// TODO: If limits and policy are conflicting, set some condition on the VPA.
		if containerControlledValues == vpa_types.ContainerControlledValuesRequestsOnly {
			annotations = capRecommendationToContainerLimit(recommendation, containerLimits)
			if genAnnotations {
				cappingAnnotations = append(cappingAnnotations, annotations...)
			}
		}
	}

	process(cappedRecommendations.Target, true)
	process(cappedRecommendations.LowerBound, false)
	process(cappedRecommendations.UpperBound, false)

	return cappedRecommendations, cappingAnnotations, nil
}

// capRecommendationByPodConstraints operates on container-level recommendations.
// It returns updated container-level recommendations that comply with the
// pod-level minAllowed and maxAllowed. It also returns annotations when
// violations occur.
func capRecommendationByPodConstraints(containerRecommendations []vpa_types.RecommendedContainerResources,
	podPolicy *vpa_types.PodResourcePolicies) ([]vpa_types.RecommendedContainerResources, []string) {
	// Due to no constraints, return the unmodified container-level recommendations
	if podPolicy == nil || (podPolicy.MinAllowed == nil && podPolicy.MaxAllowed == nil) {
		return containerRecommendations, nil
	}

	podRecommendations := resourcehelpers.SumContainerLevelRecommendations(containerRecommendations)
	updatedContainerRecommendations := applyPodLevelBoundsToContainers(
		*podRecommendations,
		containerRecommendations,
		podPolicy.MinAllowed,
		podPolicy.MaxAllowed,
	)
	updatedContainerRecommendations = ensureBoundsAreValid(updatedContainerRecommendations)
	annotations := getPodLevelVPAPolicyAnnotations(podRecommendations.Target, podPolicy)

	return updatedContainerRecommendations, annotations
}

// getPodLevelVPAPolicyAnnotations returns annotations based on the
// pod-level VPA policies. This function should only run on target-level
// recommendations, as they are equivalent to resource requests.
func getPodLevelVPAPolicyAnnotations(recommendation corev1.ResourceList, policy *vpa_types.PodResourcePolicies) []string {
	if policy == nil {
		return nil
	}
	annotations := make([]string, 0)
	for resourceName, recommended := range recommendation {
		_, isCapped := maybeCapToPodLevelPolicyMin(recommended, resourceName, policy)
		if isCapped {
			annotations = append(annotations, toCappingAnnotation(true, resourceName, cappedToPodLevelMinAllowed))
		}
		_, isCapped = maybeCapToPodLevelPolicyMax(recommended, resourceName, policy)
		if isCapped {
			annotations = append(annotations, toCappingAnnotation(true, resourceName, cappedToPodLevelMaxAllowed))
		}
	}
	return annotations
}

// capRecommendationToContainerLimit makes sure recommendation is not above current limit for the container.
// If this function makes adjustments appropriate annotations are returned.
func capRecommendationToContainerLimit(recommendation corev1.ResourceList, containerLimits corev1.ResourceList) []string {
	annotations := make([]string, 0)
	// Iterate over limits set in the container. Unset means Infinite limit.
	for resourceName, limit := range containerLimits {
		recommendedValue, found := recommendation[resourceName]
		if found && recommendedValue.MilliValue() > limit.MilliValue() {
			recommendation[resourceName] = limit
			annotations = append(annotations, toCappingAnnotation(false, resourceName, cappedToLimit))
		}
	}
	return annotations
}

// applyVPAPolicy updates recommendation if recommended resources are outside of limits defined in VPA resources policy
func applyVPAPolicy(recommendation corev1.ResourceList, policy *vpa_types.ContainerResourcePolicy) []string {
	if policy == nil {
		return nil
	}
	annotations := make([]string, 0)
	for resourceName, recommended := range recommendation {
		cappedToMin, isCapped := maybeCapToPolicyMin(recommended, resourceName, policy)
		recommendation[resourceName] = cappedToMin
		if isCapped {
			annotations = append(annotations, toCappingAnnotation(false, resourceName, cappedToMinAllowed))
		}
		cappedToMax, isCapped := maybeCapToPolicyMax(cappedToMin, resourceName, policy)
		recommendation[resourceName] = cappedToMax
		if isCapped {
			annotations = append(annotations, toCappingAnnotation(false, resourceName, cappedToMaxAllowed))
		}
	}
	return annotations
}

func applyVPAPolicyForContainer(containerName string,
	containerRecommendation *vpa_types.RecommendedContainerResources,
	policy *vpa_types.PodResourcePolicy,
	globalMaxAllowed corev1.ResourceList) (*vpa_types.RecommendedContainerResources, error) {
	if containerRecommendation == nil {
		return nil, fmt.Errorf("no recommendation available for container name %v", containerName)
	}
	cappedRecommendations := containerRecommendation.DeepCopy()
	containerPolicy := GetContainerResourcePolicy(containerName, policy)

	var minAllowed corev1.ResourceList
	if containerPolicy != nil {
		minAllowed = containerPolicy.MinAllowed
	}

	var maxAllowed corev1.ResourceList
	if containerPolicy != nil {
		// Deep copy containerPolicy.MaxAllowed as maxAllowed can later on be merged with globalMaxAllowed.
		// Deep copy is needed to prevent unwanted modifications to containerPolicy.MaxAllowed.
		maxAllowed = containerPolicy.MaxAllowed.DeepCopy()
	}
	if maxAllowed == nil {
		maxAllowed = globalMaxAllowed
	} else {
		// Set resources from the global max allowed if the VPA max allowed is missing them.
		for resourceName, quantity := range globalMaxAllowed {
			if _, ok := maxAllowed[resourceName]; !ok {
				maxAllowed[resourceName] = quantity
			}
		}
	}

	process := func(recommendation corev1.ResourceList) {
		for resourceName := range recommendation {
			if minAllowed != nil {
				cappedToMin, _ := maybeCapToMin(recommendation[resourceName], resourceName, minAllowed)
				recommendation[resourceName] = cappedToMin
			}
			if maxAllowed != nil {
				cappedToMax, _ := maybeCapToMax(recommendation[resourceName], resourceName, maxAllowed)
				recommendation[resourceName] = cappedToMax
			}
		}
	}

	process(cappedRecommendations.Target)
	process(cappedRecommendations.LowerBound)
	process(cappedRecommendations.UpperBound)

	return cappedRecommendations, nil
}

func maybeCapToPolicyMin(recommended resource.Quantity, resourceName corev1.ResourceName,
	containerPolicy *vpa_types.ContainerResourcePolicy) (resource.Quantity, bool) {
	return maybeCapToMin(recommended, resourceName, containerPolicy.MinAllowed)
}

func maybeCapToPodLevelPolicyMin(recommended resource.Quantity, resourceName corev1.ResourceName,
	podPolicy *vpa_types.PodResourcePolicies) (resource.Quantity, bool) {
	return maybeCapToMin(recommended, resourceName, podPolicy.MinAllowed)
}

func maybeCapToPolicyMax(recommended resource.Quantity, resourceName corev1.ResourceName,
	containerPolicy *vpa_types.ContainerResourcePolicy) (resource.Quantity, bool) {
	return maybeCapToMax(recommended, resourceName, containerPolicy.MaxAllowed)
}

func maybeCapToPodLevelPolicyMax(recommended resource.Quantity, resourceName corev1.ResourceName,
	podPolicy *vpa_types.PodResourcePolicies) (resource.Quantity, bool) {
	return maybeCapToMax(recommended, resourceName, podPolicy.MaxAllowed)
}

func maybeCapToMax(recommended resource.Quantity, resourceName corev1.ResourceName,
	maxVal corev1.ResourceList) (resource.Quantity, bool) {
	val, found := maxVal[resourceName]
	if found && !val.IsZero() && recommended.Cmp(val) > 0 {
		return val, true
	}
	return recommended, false
}

func maybeCapToMin(recommended resource.Quantity, resourceName corev1.ResourceName,
	minAllowed corev1.ResourceList) (resource.Quantity, bool) {
	minResource, found := minAllowed[resourceName]
	if found && !minResource.IsZero() && recommended.Cmp(minResource) < 0 {
		return minResource, true
	}
	return recommended, false
}

// ApplyVPAPolicy returns a recommendation adjusted to obey policies
// defined at both the container and Pod level, for example:
//
// - recommender flags that set maximum constraints
//
// - container and Pod-level minAllowed and maxAllowed values defined in the VPA object
func ApplyVPAPolicy(podLevel bool, podRecommendation *vpa_types.RecommendedPodResources,
	policy *vpa_types.PodResourcePolicy, globalMaxAllowed limits.GlobalMaxAllowed) (*vpa_types.RecommendedPodResources, error) {
	if podRecommendation == nil {
		return nil, nil
	}

	// Adjust container-level recommendations according to container-level policies
	updatedRecommendations := []vpa_types.RecommendedContainerResources{}
	for _, containerRecommendation := range podRecommendation.ContainerRecommendations {
		containerName := containerRecommendation.ContainerName
		updatedContainerResources, err := applyVPAPolicyForContainer(containerName,
			&containerRecommendation, policy, globalMaxAllowed.Container)
		if err != nil {
			return nil, fmt.Errorf("cannot apply policy on recommendation for container name %v", containerName)
		}
		updatedRecommendations = append(updatedRecommendations, *updatedContainerResources)
	}

	// Adjust container-level recommendations according to Pod-level policies
	if podLevel {
		podLevelRecommendations := resourcehelpers.SumContainerLevelRecommendations(updatedRecommendations)

		if policy == nil {
			policy = &vpa_types.PodResourcePolicy{
				PodPolicies: &vpa_types.PodResourcePolicies{},
			}
		}
		podPolicies := policy.PodPolicies
		var minAllowed, maxAllowed corev1.ResourceList
		if podPolicies != nil {
			minAllowed = podPolicies.MinAllowed
			maxAllowed = podPolicies.MaxAllowed
		}

		if maxAllowed == nil {
			maxAllowed = globalMaxAllowed.Pod
		} else {
			// Set resources from the global max allowed if the VPA max allowed is missing them
			for resourceName, quantity := range globalMaxAllowed.Pod {
				if _, ok := maxAllowed[resourceName]; !ok {
					maxAllowed[resourceName] = quantity
				}
			}
		}
		updatedRecommendations = applyPodLevelBoundsToContainers(
			*podLevelRecommendations,
			updatedRecommendations,
			minAllowed,
			maxAllowed,
		)
		updatedRecommendations = ensureBoundsAreValid(updatedRecommendations)
	}

	return &vpa_types.RecommendedPodResources{
		ContainerRecommendations: updatedRecommendations,
		PodRecommendations:       nil,
	}, nil
}

// applyPodLevelBoundsToContainers adjusts container-level recommendations
// based on Pod-level constraints, such as minAllowed and maxAllowed,
// or constraints passed to the recommender as flags.
func applyPodLevelBoundsToContainers(
	podLevelRecs vpa_types.PodRecommendations,
	containerLevelRecs []vpa_types.RecommendedContainerResources,
	minAllowed, maxAllowed corev1.ResourceList,
) []vpa_types.RecommendedContainerResources {
	if minAllowed == nil && maxAllowed == nil {
		return containerLevelRecs
	}

	containerLevelRecs = processContainerLevelRecs(podLevelRecs.LowerBound, containerLevelRecs, minAllowed, maxAllowed, getLower)
	containerLevelRecs = processContainerLevelRecs(podLevelRecs.Target, containerLevelRecs, minAllowed, maxAllowed, getTarget)
	containerLevelRecs = processContainerLevelRecs(podLevelRecs.UpperBound, containerLevelRecs, minAllowed, maxAllowed, getUpper)
	return containerLevelRecs
}

// processContainerLevelRecs adjusts a single type of container-level
// recommendation, such as LowerBound, based on Pod-level boundaries.
func processContainerLevelRecs(
	podLevelRec corev1.ResourceList,
	containerLevelRecs []vpa_types.RecommendedContainerResources,
	mins corev1.ResourceList,
	maxs corev1.ResourceList,
	fieldGetter func(vpa_types.RecommendedContainerResources) *corev1.ResourceList,
) []vpa_types.RecommendedContainerResources {
	scaleAll := func(resourceName corev1.ResourceName, roundingMode roundingMode, podRec, target resource.Quantity) {
		for i := range containerLevelRecs {
			current := (*fieldGetter(containerLevelRecs[i]))[resourceName]

			if current.IsZero() {
				continue
			}

			var scaled *resource.Quantity
			if resourceName == corev1.ResourceMemory {
				scaled, _ = scaleQuantityProportionallyMem(&current, &podRec, &target, roundUpToFullUnit)
			} else {
				scaled, _ = scaleQuantityProportionallyCPU(&current, &podRec, &target, roundingMode)
			}
			(*fieldGetter(containerLevelRecs[i]))[resourceName] = *scaled
		}
	}

	applyBound := func(bounds corev1.ResourceList, roundingMode roundingMode, shouldScale func(bound, podRec resource.Quantity) bool) {
		for resourceName, bound := range bounds {
			podRec, ok := podLevelRec[resourceName]
			if !ok {
				continue
			}
			if shouldScale(bound, podRec) {
				scaleAll(resourceName, roundingMode, podRec, bound)
			}
		}
	}

	// If the Pod-level recommendation is below min, scale up container-level recommendations to satisfy the minimum.
	applyBound(mins, roundUpToFullUnit, func(bound, podRec resource.Quantity) bool { return podRec.Cmp(bound) < 0 })

	// If the Pod-level recommendation is above max, scale down container-level recommendations to satisfy the max.
	applyBound(maxs, roundDownToFullUnit, func(bound, podRec resource.Quantity) bool { return podRec.Cmp(bound) > 0 })

	return containerLevelRecs
}

// ensureBoundsAreValid ensures that for each container-level recommendation,
// LowerBound <= Target <= UpperBound always holds.
//
// The function treats Target as the reference value and adjusts LowerBound
// and UpperBound as needed to maintain valid bounds.
//
// When a violation occurs, the function redistributes freed values
// across containers where violations do not occur. It may also take values
// from containers without violations and assign them to containers that
// lack sufficient values to resolve the violation.
func ensureBoundsAreValid(containerLevelRecs []vpa_types.RecommendedContainerResources) []vpa_types.RecommendedContainerResources {
	validateBounds := func(
		getBound func(vpa_types.RecommendedContainerResources) *corev1.ResourceList,
		boundValid func(bound, floor resource.Quantity) bool) {
		for _, resourceName := range []corev1.ResourceName{corev1.ResourceCPU, corev1.ResourceMemory} {
			var valuesToRedistribute resource.Quantity
			var sumEligible resource.Quantity
			eligible := make([]int, 0, len(containerLevelRecs))

			for i := range containerLevelRecs {
				targetRL := getTarget(containerLevelRecs[i])
				boundRL := getBound(containerLevelRecs[i])

				target, targetOK := (*targetRL)[resourceName]
				bound, boundOK := (*boundRL)[resourceName]
				if !targetOK || !boundOK {
					continue
				}

				// violated, bound needs to be set to target and reditribute totalAdded
				if !boundValid(target, bound) {
					added := target.DeepCopy()
					added.Sub(bound) // target - bound
					valuesToRedistribute.Add(added)
					(*boundRL)[resourceName] = target
					continue
				}

				eligible = append(eligible, i)
				sumEligible.Add(bound)
			}

			if valuesToRedistribute.IsZero() || sumEligible.IsZero() || len(eligible) == 0 {
				continue
			}

			for _, idx := range eligible {
				targetRL := getTarget(containerLevelRecs[idx])
				boundRL := getBound(containerLevelRecs[idx])

				target, targetOK := (*targetRL)[resourceName]
				bound, boundOK := (*boundRL)[resourceName]
				if !targetOK || !boundOK {
					continue
				}
				var toSubtract *resource.Quantity
				if resourceName == corev1.ResourceMemory {
					toSubtract, _ = scaleQuantityProportionallyMem(&bound, &sumEligible, &valuesToRedistribute, roundUpToFullUnit)
				} else {
					toSubtract, _ = scaleQuantityProportionallyCPU(&bound, &sumEligible, &valuesToRedistribute, noRounding)
				}

				updated := bound.DeepCopy()
				updated.Sub(*toSubtract)

				if !boundValid(updated, target) {
					(*boundRL)[resourceName] = updated
				} else {
					(*boundRL)[resourceName] = target
				}
			}
		}
	}

	// enforces "target >= LowerBound" (per-container)
	validateBounds(getLower, func(target, lowerBound resource.Quantity) bool { return target.Cmp(lowerBound) >= 0 })

	// enforces "target <= UpperBound" (per-container)
	validateBounds(getUpper, func(target, upperBound resource.Quantity) bool { return target.Cmp(upperBound) <= 0 })

	return containerLevelRecs
}

func getRecommendationForContainer(containerName string, resources []vpa_types.RecommendedContainerResources) *vpa_types.RecommendedContainerResources {
	for _, containerRec := range resources {
		if containerRec.ContainerName == containerName {
			return &containerRec
		}
	}
	return nil
}

// GetRecommendationForContainer returns recommendation for given container name
func GetRecommendationForContainer(containerName string, recommendation *vpa_types.RecommendedPodResources) *vpa_types.RecommendedContainerResources {
	if recommendation != nil {
		if recommendationForContainer := getRecommendationForContainer(containerName, recommendation.ContainerRecommendations); recommendationForContainer != nil {
			result := *recommendationForContainer
			return &result
		}
	}
	return nil
}

func getContainer(containerName string, pod *corev1.Pod) *corev1.Container {
	for i, container := range pod.Spec.Containers {
		if container.Name == containerName {
			return &pod.Spec.Containers[i]
		}
	}
	return nil
}

// applyContainerLimitRange updates recommendation if recommended resources are outside of limits defined in VPA resources policy
func applyContainerLimitRange(isPodLevel bool, recommendation corev1.ResourceList,
	containerRequests corev1.ResourceList, containerLimits corev1.ResourceList,
	limitRange *corev1.LimitRangeItem) []string {
	annotations := make([]string, 0)
	if limitRange == nil {
		return annotations
	}
	podlevel := false
	cappedMinAction := cappedProportionallyToMinLimit
	cappedMaxAction := cappedProportionallyToMaxLimit
	if isPodLevel {
		podlevel = true
		cappedMinAction = cappedToPodLimitRangeMin
		cappedMaxAction = cappedToPodLimitRangeMax
	}

	maxAllowedRecommendation := getMaxAllowedRecommendation(recommendation, containerRequests, containerLimits, limitRange)
	minAllowedRecommendation := getMinAllowedRecommendation(recommendation, containerRequests, containerLimits, limitRange)
	for resourceName, recommended := range recommendation {
		cappedToMin, isCapped := maybeCapToMin(recommended, resourceName, minAllowedRecommendation)
		recommendation[resourceName] = cappedToMin
		if isCapped {
			annotations = append(annotations, toCappingAnnotation(podlevel, resourceName, cappedMinAction))
		}
		cappedToMax, isCapped := maybeCapToMax(cappedToMin, resourceName, maxAllowedRecommendation)
		recommendation[resourceName] = cappedToMax
		if isCapped {
			annotations = append(annotations, toCappingAnnotation(podlevel, resourceName, cappedMaxAction))
		}
	}
	return annotations
}

func getMaxAllowedRecommendation(recommendation corev1.ResourceList,
	containerRequests corev1.ResourceList, containerLimits corev1.ResourceList,
	podLimitRange *corev1.LimitRangeItem) corev1.ResourceList {
	if podLimitRange == nil {
		return corev1.ResourceList{}
	}
	return getBoundaryRecommendation(recommendation, containerRequests, containerLimits, podLimitRange.Max, podLimitRange.Default)
}

func getMinAllowedRecommendation(recommendation corev1.ResourceList,
	containerRequests corev1.ResourceList, containerLimits corev1.ResourceList,
	podLimitRange *corev1.LimitRangeItem) corev1.ResourceList {
	// Both limit and request must be higher than min set in the limit range:
	// https://github.com/kubernetes/kubernetes/blob/016e9d5c06089774c6286fd825302cbae661a446/plugin/pkg/admission/limitranger/admission.go#L303
	if podLimitRange == nil {
		return corev1.ResourceList{}
	}
	minForLimit := getBoundaryRecommendation(recommendation, containerRequests, containerLimits, podLimitRange.Min, podLimitRange.Default)
	minForRequest := podLimitRange.Min
	if minForRequest == nil {
		return minForLimit
	}
	result := minForLimit
	if minForRequest.Cpu() != nil && minForRequest.Cpu().Cmp(*minForLimit.Cpu()) > 0 {
		result[corev1.ResourceCPU] = *minForRequest.Cpu()
	}
	if minForRequest.Memory() != nil && minForRequest.Memory().Cmp(*minForLimit.Memory()) > 0 {
		result[corev1.ResourceMemory] = *minForRequest.Memory()
	}
	return result
}

func getBoundaryRecommendation(recommendation corev1.ResourceList,
	containerRequests corev1.ResourceList, containerLimits corev1.ResourceList,
	boundaryLimit, defaultLimit corev1.ResourceList) corev1.ResourceList {
	if boundaryLimit == nil {
		return corev1.ResourceList{}
	}
	boundaryCpu := GetBoundaryRequest(corev1.ResourceCPU, containerRequests.Cpu(), containerLimits.Cpu(), boundaryLimit.Cpu(), defaultLimit.Cpu())
	boundaryMem := GetBoundaryRequest(corev1.ResourceMemory, containerRequests.Memory(), containerLimits.Memory(), boundaryLimit.Memory(), defaultLimit.Memory())
	return corev1.ResourceList{
		corev1.ResourceCPU:    *boundaryCpu,
		corev1.ResourceMemory: *boundaryMem,
	}
}

type containerWithRecommendation struct {
	container      *corev1.Container
	recommendation *vpa_types.RecommendedContainerResources
}

func zipContainersWithRecommendations(resources []vpa_types.RecommendedContainerResources, pod *corev1.Pod) []containerWithRecommendation {
	result := make([]containerWithRecommendation, 0)
	for _, container := range pod.Spec.Containers {
		recommendation := getRecommendationForContainer(container.Name, resources)
		result = append(result, containerWithRecommendation{container: &container, recommendation: recommendation})
	}
	return result
}

func applyPodLimitRange(resources []vpa_types.RecommendedContainerResources,
	pod *corev1.Pod, limitRange corev1.LimitRangeItem, resourceName corev1.ResourceName,
	fieldGetter func(vpa_types.RecommendedContainerResources) *corev1.ResourceList) []vpa_types.RecommendedContainerResources {
	minLimit := limitRange.Min[resourceName]
	maxLimit := limitRange.Max[resourceName]
	defaultLimit := limitRange.Default[resourceName]

	containersWithRecommendations := zipContainersWithRecommendations(resources, pod)
	var sumLimit, sumRecommendation resource.Quantity
	for _, containerWithRecommendation := range containersWithRecommendations {
		container := containerWithRecommendation.container
		requests, limits := resourcehelpers.ContainerRequestsAndLimits(container.Name, pod)
		limit := limits[resourceName]
		request := requests[resourceName]
		var recommendation resource.Quantity
		if containerWithRecommendation.recommendation == nil {
			// No recommendation, don't change the container
			recommendation = request
		} else {
			recommendation = (*fieldGetter(*containerWithRecommendation.recommendation))[resourceName]
		}
		containerLimit, _ := getProportionalResourceLimit(resourceName, &limit, &request, &recommendation, &defaultLimit)
		if containerLimit != nil {
			sumLimit.Add(*containerLimit)
		}
		sumRecommendation.Add(recommendation)
	}

	if sumRecommendation.IsZero() {
		return resources
	}

	if minLimit.Cmp(sumLimit) <= 0 && minLimit.Cmp(sumRecommendation) <= 0 && (maxLimit.IsZero() || maxLimit.Cmp(sumLimit) >= 0) {
		return resources
	}

	// if minLimit.Cmp(sumRecommendation) > 0 && !sumLimit.IsZero() {
	if minLimit.Cmp(sumRecommendation) > 0 {
		for _, containerWithRecommendation := range containersWithRecommendations {
			if containerWithRecommendation.recommendation == nil {
				continue
			}
			request := (*fieldGetter(*containerWithRecommendation.recommendation))[resourceName]
			var cappedContainerRequest *resource.Quantity
			if resourceName == corev1.ResourceMemory {
				cappedContainerRequest, _ = scaleQuantityProportionallyMem(&request, &sumRecommendation, &minLimit, roundUpToFullUnit)
			} else {
				cappedContainerRequest, _ = scaleQuantityProportionallyCPU(&request, &sumRecommendation, &minLimit, roundUpToFullUnit)
			}
			(*fieldGetter(*containerWithRecommendation.recommendation))[resourceName] = *cappedContainerRequest
		}
		return resources
	}

	// if sumLimit.IsZero() {
	// 	return resources
	// }

	var targetTotalLimit resource.Quantity
	if minLimit.Cmp(sumLimit) > 0 {
		targetTotalLimit = minLimit
	}
	if !maxLimit.IsZero() && maxLimit.Cmp(sumLimit) < 0 {
		targetTotalLimit = maxLimit
	}
	for _, containerWithRecommendation := range containersWithRecommendations {
		var limit resource.Quantity
		if containerWithRecommendation.recommendation == nil {
			// No recommendation, don't change the container
			_, limits := resourcehelpers.ContainerRequestsAndLimits(containerWithRecommendation.container.Name, pod)
			limit = limits[resourceName]
		} else {
			limit = (*fieldGetter(*containerWithRecommendation.recommendation))[resourceName]
		}

		var cappedContainerRequest *resource.Quantity
		if resourceName == corev1.ResourceMemory {
			cappedContainerRequest, _ = scaleQuantityProportionallyMem(&limit, &sumLimit, &targetTotalLimit, roundDownToFullUnit)
		} else {
			cappedContainerRequest, _ = scaleQuantityProportionallyCPU(&limit, &sumLimit, &targetTotalLimit, roundDownToFullUnit)
		}
		(*fieldGetter(*containerWithRecommendation.recommendation))[resourceName] = *cappedContainerRequest
	}
	return resources
}

func recommendationForContainerExists(containerName string, containerRecommendations []vpa_types.RecommendedContainerResources) bool {
	for _, recommendation := range containerRecommendations {
		if containerName == recommendation.ContainerName {
			return true
		}
	}
	return false
}

func insertRequestsForMissingRecommendations(containerRecommendations []vpa_types.RecommendedContainerResources, pod *corev1.Pod) []vpa_types.RecommendedContainerResources {
	result := make([]vpa_types.RecommendedContainerResources, 0)
	for _, r := range containerRecommendations {
		result = append(result, *r.DeepCopy())
	}
	for _, container := range pod.Spec.Containers {
		if recommendationForContainerExists(container.Name, containerRecommendations) {
			continue
		}
		requests, _ := resourcehelpers.ContainerRequestsAndLimits(container.Name, pod)
		if len(requests) == 0 {
			continue
		}
		result = append(result, vpa_types.RecommendedContainerResources{
			ContainerName: container.Name,
			Target:        requests,
		})
	}
	return result
}

// capProportionallyToPodLimitRange adjusts container-level recommendations to comply with
// constraints set by LimitRange objects of type Pod.
func (c *cappingRecommendationProcessor) capProportionallyToPodLimitRange(
	containerRecommendations []vpa_types.RecommendedContainerResources, pod *corev1.Pod) ([]vpa_types.RecommendedContainerResources, error) {
	podLimitRange, err := c.limitsRangeCalculator.GetPodLimitRangeItem(pod.Namespace)
	if err != nil {
		return nil, fmt.Errorf("error obtaining limit range: %s", err)
	}
	if podLimitRange == nil {
		return containerRecommendations, nil
	}
	containerRecommendations = insertRequestsForMissingRecommendations(containerRecommendations, pod)
	containerRecommendations = applyPodLimitRange(containerRecommendations, pod, *podLimitRange, corev1.ResourceCPU, getUpper)
	containerRecommendations = applyPodLimitRange(containerRecommendations, pod, *podLimitRange, corev1.ResourceMemory, getUpper)

	containerRecommendations = applyPodLimitRange(containerRecommendations, pod, *podLimitRange, corev1.ResourceCPU, getTarget)
	containerRecommendations = applyPodLimitRange(containerRecommendations, pod, *podLimitRange, corev1.ResourceMemory, getTarget)

	containerRecommendations = applyPodLimitRange(containerRecommendations, pod, *podLimitRange, corev1.ResourceCPU, getLower)
	containerRecommendations = applyPodLimitRange(containerRecommendations, pod, *podLimitRange, corev1.ResourceMemory, getLower)
	return containerRecommendations, nil
}

func getLower(rl vpa_types.RecommendedContainerResources) *corev1.ResourceList  { return &rl.LowerBound }
func getTarget(rl vpa_types.RecommendedContainerResources) *corev1.ResourceList { return &rl.Target }
func getUpper(rl vpa_types.RecommendedContainerResources) *corev1.ResourceList  { return &rl.UpperBound }
