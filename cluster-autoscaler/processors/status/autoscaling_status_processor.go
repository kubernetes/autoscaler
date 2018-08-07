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
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"time"
)

// AutoscalingStatusProcessor processes the status of the cluster after each autoscaling iteration.
// It's triggered at the end of Autoscaler's RunOnce method.
type AutoscalingStatusProcessor interface {
	Process(context *context.AutoscalingContext, csr *clusterstate.ClusterStateRegistry, now time.Time) error
	CleanUp()
}

// NewDefaultAutoscalingStatusProcessor creates a default instance of AutoscalingStatusProcessor.
func NewDefaultAutoscalingStatusProcessor() AutoscalingStatusProcessor {
	return &NoOpAutoscalingStatusProcessor{}
}

// NoOpAutoscalingStatusProcessor is an AutoscalingStatusProcessor implementation useful for testing.
type NoOpAutoscalingStatusProcessor struct{}

// Process processes the status of the cluster after an autoscaling iteration.
func (p *NoOpAutoscalingStatusProcessor) Process(context *context.AutoscalingContext, csr *clusterstate.ClusterStateRegistry, now time.Time) error {
	return nil
}

// CleanUp cleans up the processor's internal structures.
func (p *NoOpAutoscalingStatusProcessor) CleanUp() {
}
