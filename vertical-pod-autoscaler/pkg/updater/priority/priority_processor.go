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

func (*defaultPriorityProcessor) GetUpdatePriority(pod *apiv1.Pod, vpa *vpa_types.VerticalPodAutoscaler,
	recommendation *vpa_types.RecommendedPodResources) PodPriority {
	outsideRecommendedRange := false
	scaleUp := false
	// Sum of requests over all containers, per resource type.
	totalRequestPerResource := make(map[apiv1.ResourceName]int64)
	// Sum of recommendations over all containers, per resource type.
	totalRecommendedPerResource := make(map[apiv1.ResourceName]int64)

	hasObservedContainers, vpaContainerSet := parseVpaObservedContainers(pod)

	for num, podContainer := range pod.Spec.Containers {
		if hasObservedContainers && !vpaContainerSet.Has(podContainer.Name) {
			klog.V(4).InfoS("Not listed in VPA observed containers label. Skipping container priority calculations", "label", annotations.VpaObservedContainersLabel, "observedContainers", pod.GetAnnotations()[annotations.VpaObservedContainersLabel], "containerName", podContainer.Name, "vpa", klog.KObj(vpa))
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
				if recommended.MilliValue() > request.MilliValue() {
					scaleUp = true
				}
				if (hasLowerBound && request.Cmp(lowerBound) < 0) ||
					(hasUpperBound && request.Cmp(upperBound) > 0) {
					outsideRecommendedRange = true
				}

				// TODO(jkyros): I think we're picking up early zeroes here from the VPA when it has no recommendation, I think that's why I have to wait
				// for the recommendation later before I try to scale in-place
				// TODO(jkyros): For in place VPA, this might be gross, but we need this pod to be in the eviction list because it doesn't actually have
				// the resources it asked for even if the spec is right, and we might need to fall back to evicting it
				// TODO(jkyros): Can we have empty container status at this point for real? It's at least failing the tests if we don't check, but
				// we could just populate the status in the tests
				// Statuses can be missing, or status resources can be nil
				if len(pod.Status.ContainerStatuses) > num && pod.Status.ContainerStatuses[num].Resources != nil {
					if statusRequest, hasStatusRequest := pod.Status.ContainerStatuses[num].Resources.Requests[resourceName]; hasStatusRequest {
						// If we're updating, but we still don't have what we asked for, we may still need to act on this pod
						if request.MilliValue() > statusRequest.MilliValue() {
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
				// Note: if the request is not specified, the container will use the
				// namespace default request. Currently we ignore it and treat such
				// containers as if they had 0 request. A more correct approach would
				// be to always calculate the 'effective' request.
				scaleUp = true
				outsideRecommendedRange = true
			}
		}
	}

	// TODO(jkyros): hmm this gets hairy here because if "status" is what let us into the list,
	// we probably need to do these calculations vs the status rather than the spec, because the
	// spec is a "lie"
	resourceDiff := 0.0
	for resource, totalRecommended := range totalRecommendedPerResource {
		totalRequest := math.Max(float64(totalRequestPerResource[resource]), 1.0)
		resourceDiff += math.Abs(totalRequest-float64(totalRecommended)) / totalRequest
	}
	return PodPriority{
		OutsideRecommendedRange: outsideRecommendedRange,
		ScaleUp:                 scaleUp,
		ResourceDiff:            resourceDiff,
	}
}
