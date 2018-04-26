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

package pods

import (
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/context"
)

// PodListProcessor processes lists of unschedulable and sheduled pods before scaling of the cluster.
type PodListProcessor interface {
	Process(context *context.AutoscalingContext, unschedulablePods []*apiv1.Pod, allScheduled []*apiv1.Pod, nodes []*apiv1.Node) ([]*apiv1.Pod, []*apiv1.Pod, error)
}

// NoOpPodListProcessor is returning pod lists without processing them.
type NoOpPodListProcessor struct {
}

// NewDefaultPodListProcessor creates an instance of PodListProcessor.
func NewDefaultPodListProcessor() PodListProcessor {
	return &NoOpPodListProcessor{}
}

// Process processes lists of unschedulable and sheduled pods before scaling of the cluster.
func (p *NoOpPodListProcessor) Process(context *context.AutoscalingContext, unschedulablePods []*apiv1.Pod, allScheduled []*apiv1.Pod, nodes []*apiv1.Node) ([]*apiv1.Pod, []*apiv1.Pod, error) {
	return unschedulablePods, allScheduled, nil
}
