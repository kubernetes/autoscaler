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

package priority

import (
	"math"

	apiv1 "k8s.io/api/core/v1"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/annotations"
	vpa_api_util "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/vpa"
	"k8s.io/klog/v2"
)

// PriorityProcessor calculates priority for pod updates.
type PriorityProcessor interface {
	GetUpdatePriority(pod *apiv1.Pod, vpa *vpa_types.VerticalPodAutoscaler,
		recommendation *vpa_types.RecommendedPodResources) PodPriority
}

// NewProcessor creates a new default PriorityProcessor.
func NewProcessor() PriorityProcessor {
	return &defaultPriorityProcessor{}
}

type defaultPriorityProcessor struct {
}

func (*defaultPriorityProcessor) GetUpdatePriority(pod *apiv1.Pod, _ *vpa_types.VerticalPodAutoscaler,
	recommendation *vpa_types.RecommendedPodResources) PodPriority {
	outsideRecommendedRange := false
	scaleUp := false
	// Sum of requests over all containers, per resource type.
	totalRequestPerResource := make(map[apiv1.ResourceName]int64)
	// Sum of recommendations over all containers, per resource type.
	totalRecommendedPerResource := make(map[apiv1.ResourceName]int64)

	// The denominator doesn't work out if we have all the resources in here
	totalRecommendedPerStatusResource := make(map[apiv1.ResourceName]int64)
	// Sum of status requests in containers where the requests already match (e.g. in-place resize is InProgress)
	totalRequestPerStatusResource := make(map[apiv1.ResourceName]int64)

	hasObservedContainers, vpaContainerSet := parseVpaObservedContainers(pod)

	for num, podContainer := range pod.Spec.Containers {
		if hasObservedContainers && !vpaContainerSet.Has(podContainer.Name) {
			klog.V(4).Infof("Not listed in %s:%s. Skipping container %s priority calculations",
				annotations.VpaObservedContainersLabel, pod.GetAnnotations()[annotations.VpaObservedContainersLabel], podContainer.Name)
			continue
		}
		recommendedRequest := vpa_api_util.GetRecommendationForContainer(podContainer.Name, recommendation)
		if recommendedRequest == nil {
			continue
		}
		for resourceName, recommended := range recommendedRequest.Target {
			totalRecommendedPerResource[resourceName] += recommended.MilliValue()
			lowerBound, hasLowerBound := recommendedRequest.LowerBound[resourceName]
			upperBound, hasUpperBound := recommendedRequest.UpperBound[resourceName]
			if request, hasRequest := podContainer.Resources.Requests[resourceName]; hasRequest {
				totalRequestPerResource[resourceName] += request.MilliValue()

				// If our resources are equal, we know it's an in-place resize so we need to calculate against status
				if recommended.MilliValue() == request.MilliValue() {
					if len(pod.Status.ContainerStatuses) > num && pod.Status.ContainerStatuses[num].Resources != nil {
						if statusRequest, hasStatusRequest := pod.Status.ContainerStatuses[num].Resources.Requests[resourceName]; hasStatusRequest {

							// TODO(jkyros): Ughhhh I need this for the denominator to work out, otherwise I'm potentially dividing
							// values across different number of resource categories (e.g. mem is in sync, cpu isn't yet I divide
							totalRecommendedPerStatusResource[resourceName] += recommended.MilliValue()

							totalRequestPerStatusResource[resourceName] += statusRequest.MilliValue()
							if recommended.MilliValue() > statusRequest.MilliValue() {
								scaleUp = true
								// It's okay if we're actually still resizing, but if we can't now or we're stuck, make sure the pod
								// is still in the list so we can evict it to go live on a fatter node or something
								if pod.Status.Resize == apiv1.PodResizeStatusDeferred || pod.Status.Resize == apiv1.PodResizeStatusInfeasible {
									klog.V(4).Infof("Pod %s looks like it's stuck scaling up (%v state), leaving it in for eviction", pod.Name, pod.Status.Resize)
								} else {
									klog.V(4).Infof("Pod %s is in the process of scaling up (%v state), leaving it in so we can see if it's taking too long", pod.Name, pod.Status.Resize)
								}
							}
							// I guess if it's not outside of compliance, it's probably okay it's stuck here?
							if (hasLowerBound && statusRequest.Cmp(lowerBound) < 0) ||
								(hasUpperBound && statusRequest.Cmp(upperBound) > 0) {
								outsideRecommendedRange = true
							}
						}
					}
				} else {
					if recommended.MilliValue() > request.MilliValue() {
						scaleUp = true
					}
					if (hasLowerBound && request.Cmp(lowerBound) < 0) ||
						(hasUpperBound && request.Cmp(upperBound) > 0) {
						outsideRecommendedRange = true
					}
				}
			} else {
				// Note: if the request is not specified, the container will use the
				// namespace default request. Currently we ignore it and treat such
				// containers as if they had 0 request. A more correct approach would
				// be to always calculate the 'effective' request.
				scaleUp = true
				outsideRecommendedRange = true
			}
		}
	}
	resourceDiff := 0.0
	resourceDiffStatus := 0.0
	for resource, totalRecommended := range totalRecommendedPerResource {
		totalRequest := math.Max(float64(totalRequestPerResource[resource]), 1.0)
		resourceDiff += math.Abs(totalRequest-float64(totalRecommended)) / totalRequest
	}

	// TODO(jkyros): Calculate the skew for status for in-place updates, we can probably merge these calculations
	// someday when they work :)
	// Ohhhh are we dividing by the wrong number of resources because they aren't populated for status
	// that still doesn't work though
	for resource, totalRecommended := range totalRecommendedPerStatusResource {
		totalStatusRequest := math.Max(float64(totalRequestPerStatusResource[resource]), 1.0)
		resourceDiffStatus += math.Abs(totalStatusRequest-float64(totalRecommended)) / totalStatusRequest

	}

	resourceDiff = resourceDiff + resourceDiffStatus

	return PodPriority{
		OutsideRecommendedRange: outsideRecommendedRange,
		ScaleUp:                 scaleUp,
		ResourceDiff:            resourceDiff,
	}

}
