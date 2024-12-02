/*
Copyright 2019 The Kubernetes Authors.

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

package mpa

import (
	mpa_types "k8s.io/autoscaler/multidimensional-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1alpha1"
)

// PreProcessor processes the MPAs before applying default .
type PreProcessor interface {
	Process(mpa *mpa_types.MultidimPodAutoscaler, isCreate bool) (*mpa_types.MultidimPodAutoscaler, error)
}

// noopPreProcessor leaves pods unchanged when processing
type noopPreProcessor struct{}

// Process leaves the pod unchanged
func (p *noopPreProcessor) Process(mpa *mpa_types.MultidimPodAutoscaler, isCreate bool) (*mpa_types.MultidimPodAutoscaler, error) {
	return mpa, nil
}

// NewDefaultPreProcessor creates a PreProcessor that leaves MPAs unchanged and returns no error
func NewDefaultPreProcessor() PreProcessor {
	return &noopPreProcessor{}
}
