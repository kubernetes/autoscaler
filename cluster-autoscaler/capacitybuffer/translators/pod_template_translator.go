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

	v1 "k8s.io/autoscaler/cluster-autoscaler/apis/capacitybuffer/autoscaling.x-k8s.io/v1"
)

// podTemplateBufferTranslator translates podTemplateRef buffers specs to fill their status.
type podTemplateBufferTranslator struct {
}

// NewPodTemplateBufferTranslator creates an instance of podTemplateBufferTranslator.
func NewPodTemplateBufferTranslator() *podTemplateBufferTranslator {
	return &podTemplateBufferTranslator{}
}

// Translate translates buffers processors into pod capacity.
func (t *podTemplateBufferTranslator) Translate(buffers []*v1.CapacityBuffer) []error {
	errors := []error{}
	for _, buffer := range buffers {
		if isPodTemplateBasedBuffer(buffer) {
			podTemplateRef, numberOfPods, err := t.translate(buffer)
			if err != nil {
				setBufferAsNotReadyForProvisioning(buffer, err.Error())
				errors = append(errors, err)
			} else {
				setBufferAsReadyForProvisioning(buffer, podTemplateRef.Name, numberOfPods)
			}
		}
	}
	return errors
}

func (t *podTemplateBufferTranslator) translate(buffer *v1.CapacityBuffer) (*v1.LocalObjectRef, int32, error) {
	// Fixed Replicas will be used if both Replicas and Percent are defined
	if buffer.Spec.Replicas != nil {
		return buffer.Spec.PodTemplateRef, max(1, int32(*buffer.Spec.Replicas)), nil
	}
	return nil, 0, fmt.Errorf("Failed to translate buffer %v, Replicas should have a value when PodTemplateRef is set", buffer.Name)
}

func isPodTemplateBasedBuffer(buffer *v1.CapacityBuffer) bool {
	return buffer.Spec.PodTemplateRef != nil
}

// CleanUp cleans up the translator's internal structures.
func (t *podTemplateBufferTranslator) CleanUp() {
}
