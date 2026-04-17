/*
Copyright 2025 The Kubernetes Authors.

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

package translator

import (
	"errors"
	"math"

	corev1 "k8s.io/api/core/v1"
	apiv1 "k8s.io/autoscaler/cluster-autoscaler/apis/capacitybuffer/autoscaling.x-k8s.io/v1beta1"
	podutils "k8s.io/autoscaler/cluster-autoscaler/utils/pod"
)

// getBufferNumberOfPods calculates the desired number of pods for a buffer.
// scalableReplicas is only provided if the buffer uses a scalable object.
func getBufferNumberOfPods(buffer *apiv1.CapacityBuffer, podTemplate corev1.PodTemplateSpec, scalableReplicas *int32) (int32, error) {
	var resolved bool
	replicas := int32(math.MaxInt32)

	if buffer.Spec.Replicas != nil {
		specReplicas := max(0, *buffer.Spec.Replicas)
		replicas = min(replicas, specReplicas)
		resolved = true
	}

	if buffer.Spec.Percentage != nil && scalableReplicas != nil {
		replicas = min(replicas, replicasFromPercentage(*buffer.Spec.Percentage, *scalableReplicas))
		resolved = true
	}

	if buffer.Spec.Limits != nil {
		limitsReplicas, err := limitNumberOfPodsForResource(podTemplate, *buffer.Spec.Limits)
		if err != nil {
			return 0, err
		}
		replicas = min(replicas, limitsReplicas)
		resolved = true
	}

	if !resolved {
		return 0, errors.New("replicas, percentage and limits are not defined")
	}
	return replicas, nil
}

func replicasFromPercentage(percentage int32, scalableReplicas int32) int32 {
	return max(0, (percentage)*(scalableReplicas)/100)
}

func limitNumberOfPodsForResource(podTemplate corev1.PodTemplateSpec, limits apiv1.ResourceList) (int32, error) {
	pod := podutils.GetPodFromTemplate(&podTemplate)
	podRequests := podutils.PodRequests(pod)
	return calculateMaxPodsFromLimits(podRequests, limits)
}

func calculateMaxPodsFromLimits(podRequests corev1.ResourceList, limits apiv1.ResourceList) (int32, error) {
	maxPodsFromLimits := int32(math.MaxInt32)
	var resourceFound bool
	for resourceName, quantity := range podRequests {
		quantityMilliValue := quantity.MilliValue()
		if quantityMilliValue <= 0 {
			continue
		}
		if limitQuantity, found := limits[apiv1.ResourceName(resourceName)]; found {
			maxPods := int32(limitQuantity.MilliValue() / quantityMilliValue)
			if maxPods < 0 {
				continue
			}
			resourceFound = true
			maxPodsFromLimits = min(maxPodsFromLimits, maxPods)
		}
	}
	if !resourceFound {
		return 0, errors.New("resources in configured limits not found in the pod template")
	}
	return maxPodsFromLimits, nil
}
