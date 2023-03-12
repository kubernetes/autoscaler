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

package status

import (
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/status"
)

// ScaleDownStatusProcessor processes the status of the cluster after a scale-down.
type ScaleDownStatusProcessor interface {
	Process(context *context.AutoscalingContext, status *status.ScaleDownStatus)
	CleanUp()
}

// NewDefaultScaleDownStatusProcessor creates a default instance of ScaleUpStatusProcessor.
func NewDefaultScaleDownStatusProcessor() ScaleDownStatusProcessor {
	return &NoOpScaleDownStatusProcessor{}
}

// NoOpScaleDownStatusProcessor is a ScaleDownStatusProcessor implementations useful for testing.
type NoOpScaleDownStatusProcessor struct{}

// Process processes the status of the cluster after a scale-down.
func (p *NoOpScaleDownStatusProcessor) Process(context *context.AutoscalingContext, status *status.ScaleDownStatus) {
}

// CleanUp cleans up the processor's internal structures.
func (p *NoOpScaleDownStatusProcessor) CleanUp() {
}
