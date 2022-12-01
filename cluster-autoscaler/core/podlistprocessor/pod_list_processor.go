/*
Copyright 2022 The Kubernetes Authors.

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

package podlistprocessor

import (
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/context"
)

type defaultPodListProcessor struct {
	currentlyDrainedNodes *currentlyDrainedNodesPodListProcessor
	filterOutSchedulable  *filterOutSchedulablePodListProcessor
}

// NewDefaultPodListProcessor returns a default implementation of the pod list
// processor, which wraps and sequentially runs other sub-processors.
func NewDefaultPodListProcessor(currentlyDrainedNodes *currentlyDrainedNodesPodListProcessor, filterOutSchedulable *filterOutSchedulablePodListProcessor) *defaultPodListProcessor {
	return &defaultPodListProcessor{
		currentlyDrainedNodes: currentlyDrainedNodes,
		filterOutSchedulable:  filterOutSchedulable,
	}
}

// Process runs sub-processors sequentially
func (p *defaultPodListProcessor) Process(ctx *context.AutoscalingContext, unschedulablePods []*apiv1.Pod) ([]*apiv1.Pod, error) {
	unschedulablePods, err := p.currentlyDrainedNodes.Process(ctx, unschedulablePods)
	if err != nil {
		return nil, err
	}

	return p.filterOutSchedulable.Process(ctx, unschedulablePods)
}

func (p *defaultPodListProcessor) CleanUp() {
	p.currentlyDrainedNodes.CleanUp()
	p.filterOutSchedulable.CleanUp()
}
