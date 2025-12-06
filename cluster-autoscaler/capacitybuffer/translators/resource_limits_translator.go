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
	"fmt"

	corev1 "k8s.io/api/core/v1"
	api_v1 "k8s.io/autoscaler/cluster-autoscaler/apis/capacitybuffer/autoscaling.x-k8s.io/v1alpha1"
	cbclient "k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/client"
	"k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/common"
)

// resourceLimitsTranslator translates buffers processors into pod capacity.
type resourceLimitsTranslator struct {
	client *cbclient.CapacityBufferClient
}

// NewResourceLimitsTranslator creates an instance of resourceLimitsTranslator.
func NewResourceLimitsTranslator(client *cbclient.CapacityBufferClient) *resourceLimitsTranslator {
	return &resourceLimitsTranslator{
		client: client,
	}
}

// Translate translates buffers processors into pod capacity.
func (t *resourceLimitsTranslator) Translate(buffers []*api_v1.CapacityBuffer) []error {
	errors := []error{}
	for _, buffer := range buffers {
		if isResourcesLimitsDefinedInBuffer(buffer) {
			if buffer.Status.PodTemplateRef == nil {
				err := fmt.Errorf("Can't get pod template, PodTemplateRef is nil")
				common.SetBufferAsNotReadyForProvisioning(buffer, nil, nil, nil, buffer.Spec.ProvisioningStrategy, err)
				errors = append(errors, err)
				continue
			}
			podTemplate, err := t.client.GetPodTemplate(buffer.Namespace, buffer.Status.PodTemplateRef.Name)
			if err != nil {
				err = fmt.Errorf("Couldn't get pod template, error: %v", err)
				common.SetBufferAsNotReadyForProvisioning(buffer, nil, nil, nil, buffer.Spec.ProvisioningStrategy, err)
				errors = append(errors, err)
				continue
			}
			numberOfPods := limitNumberOfPodsForResource(podTemplate, *buffer.Spec.Limits)
			if numberOfPods == nil {
				err := fmt.Errorf("Couldn't calculate number of pods for buffer %v based on provided resource limits %v", buffer.Name, *buffer.Spec.Limits)
				common.SetBufferAsNotReadyForProvisioning(buffer, nil, nil, nil, buffer.Spec.ProvisioningStrategy, err)
				errors = append(errors, err)
				continue
			}
			if buffer.Status.Replicas != nil {
				numberOfPods = pointerToInt32(min(*buffer.Status.Replicas, *numberOfPods))
			}
			common.SetBufferAsReadyForProvisioning(buffer, buffer.Status.PodTemplateRef, &podTemplate.Generation, numberOfPods, buffer.Spec.ProvisioningStrategy)
		}
	}
	return errors
}

func limitNumberOfPodsForResource(podTemplate *corev1.PodTemplate, limits api_v1.ResourceList) *int32 {
	var maximumNumberOfPods *int32
	podResourcesValues := map[string]int64{}

	for _, container := range podTemplate.Template.Spec.Containers {
		for resourceName, quantity := range container.Resources.Requests {
			if _, found := limits[api_v1.ResourceName(resourceName.String())]; found {
				podResourcesValues[resourceName.String()] += quantity.MilliValue()
			}
		}
	}
	for resourceName, quantityMilliValue := range podResourcesValues {
		if quantityMilliValue <= 0 {
			continue
		}
		if limitQuantity, found := limits[api_v1.ResourceName(resourceName)]; found {
			maxPods := limitQuantity.MilliValue() / quantityMilliValue
			if maxPods < 0 {
				continue
			}
			if maximumNumberOfPods == nil {
				maximumNumberOfPods = pointerToInt32(int32(maxPods))
			} else {
				maximumNumberOfPods = pointerToInt32(int32(min(*maximumNumberOfPods, int32(maxPods))))
			}
		}
	}

	return maximumNumberOfPods
}

func isResourcesLimitsDefinedInBuffer(buffer *api_v1.CapacityBuffer) bool {
	return buffer.Spec.Limits != nil
}

// CleanUp cleans up the translator's internal structures.
func (t *resourceLimitsTranslator) CleanUp() {
}

func pointerToInt32(number int32) *int32 { return &number }
func pointerToInt64(number int64) *int64 { return &number }
