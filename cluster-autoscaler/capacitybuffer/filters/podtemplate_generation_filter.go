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

package filter

import (
	v1 "k8s.io/autoscaler/cluster-autoscaler/apis/capacitybuffer/autoscaling.x-k8s.io/v1alpha1"
	cbclient "k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/client"
	"k8s.io/klog/v2"
)

// podTemplateGenerationChangedFilter filters in buffers that has pod template that its generation changeed
type podTemplateGenerationChangedFilter struct {
	client *cbclient.CapacityBufferClient
}

// NewPodTemplateGenerationChangedFilter creates an instance of podTemplateGenerationChangedFilter that filters the buffers with pod templates that needs to be updated.
func NewPodTemplateGenerationChangedFilter(client *cbclient.CapacityBufferClient) *podTemplateGenerationChangedFilter {
	return &podTemplateGenerationChangedFilter{
		client: client,
	}
}

// Filter filters the passed buffers based on buffer status conditions
func (f *podTemplateGenerationChangedFilter) Filter(buffersToFilter []*v1.CapacityBuffer) ([]*v1.CapacityBuffer, []*v1.CapacityBuffer) {
	var buffers []*v1.CapacityBuffer
	var filteredOutBuffers []*v1.CapacityBuffer

	for _, buffer := range buffersToFilter {
		if f.podTemplateGenerationChanged(buffer) {
			buffers = append(buffers, buffer)
		} else {
			filteredOutBuffers = append(filteredOutBuffers, buffer)
		}
	}
	return buffers, filteredOutBuffers
}

func (f *podTemplateGenerationChangedFilter) podTemplateGenerationChanged(buffer *v1.CapacityBuffer) bool {
	if buffer.Status.PodTemplateRef == nil || buffer.Status.PodTemplateGeneration == nil {
		return false
	}

	podTemplate, err := f.client.GetPodTemplate(buffer.Namespace, buffer.Status.PodTemplateRef.Name)

	if err != nil {
		klog.Errorf("Couldn't get pod template defined in buffer %v, with error: %v", buffer.Name, err.Error())
		return false
	}
	podTemplateGeneration := podTemplate.Generation
	bufferTemplateGeneration := *buffer.Status.PodTemplateGeneration

	if podTemplateGeneration != bufferTemplateGeneration {
		return true
	}
	return false
}

// CleanUp cleans up the filter's internal structures.
func (f *podTemplateGenerationChangedFilter) CleanUp() {
}
