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

	v1 "k8s.io/autoscaler/cluster-autoscaler/apis/capacitybuffer/autoscaling.x-k8s.io/v1alpha1"
	cbclient "k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/client"
	"k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/common"
)

// podTemplateBufferTranslator translates podTemplateRef buffers specs to fill their status.
type podTemplateBufferTranslator struct {
	client *cbclient.CapacityBufferClient
}

// NewPodTemplateBufferTranslator creates an instance of podTemplateBufferTranslator.
func NewPodTemplateBufferTranslator(client *cbclient.CapacityBufferClient) *podTemplateBufferTranslator {
	return &podTemplateBufferTranslator{
		client: client,
	}
}

// Translate translates buffers processors into pod capacity.
func (t *podTemplateBufferTranslator) Translate(buffers []*v1.CapacityBuffer) []error {
	errors := []error{}
	var numberOfPods *int32
	var podTemplateRef *v1.LocalObjectRef
	for _, buffer := range buffers {
		if isPodTemplateBasedBuffer(buffer) {
			podTemplateRef = buffer.Spec.PodTemplateRef
			podTemplate, err := t.client.GetPodTemplate(buffer.Namespace, podTemplateRef.Name)
			if err != nil {
				common.SetBufferAsNotReadyForProvisioning(buffer, nil, nil, nil, buffer.Spec.ProvisioningStrategy, err)
				errors = append(errors, err)
				continue
			}
			numberOfPods = t.getNumberOfReplicas(buffer)
			if numberOfPods == nil {
				common.SetBufferAsNotReadyForProvisioning(buffer, podTemplateRef, &podTemplate.Generation, nil, buffer.Spec.ProvisioningStrategy, fmt.Errorf("Failed to get buffer's number of pods"))
				continue
			}
			common.SetBufferAsReadyForProvisioning(buffer, podTemplateRef, &podTemplate.Generation, numberOfPods, buffer.Spec.ProvisioningStrategy)
		}
	}
	return errors
}

func (t *podTemplateBufferTranslator) getNumberOfReplicas(buffer *v1.CapacityBuffer) *int32 {
	if buffer.Spec.Replicas != nil {
		replicas := max(0, int32(*buffer.Spec.Replicas))
		return &replicas
	}
	return nil
}

func isPodTemplateBasedBuffer(buffer *v1.CapacityBuffer) bool {
	return buffer.Spec.PodTemplateRef != nil
}

// CleanUp cleans up the translator's internal structures.
func (t *podTemplateBufferTranslator) CleanUp() {
}
