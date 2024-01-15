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
	"k8s.io/autoscaler/cluster-autoscaler/processors/pods"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/predicatechecker"
)

type defaultPodListProcessor struct {
	processors map[string]pods.PodListProcessor
}

// NewDefaultPodListProcessor returns a default implementation of the pod list
// processor, which wraps and sequentially runs other sub-processors.
func NewDefaultPodListProcessor(predicateChecker predicatechecker.PredicateChecker) (*defaultPodListProcessor, error) {

	f, err := NewFilterOutNoPreemptionPodsListProcessor()
	if err != nil {
		return nil, err
	}

	processors := map[string]pods.PodListProcessor{
		CurrentlyDrainedNodes: NewCurrentlyDrainedNodesPodListProcessor(),
		FilterOutSchedulable:  NewFilterOutSchedulablePodListProcessor(predicateChecker),
		FilterOutDaemonSet:    NewFilterOutDaemonSetPodListProcessor(),
		FilterOutNoPreemption: f,
	}

	return &defaultPodListProcessor{
		processors: processors,
	}, nil
}

// Update updates the state of defaultPodListProcessor with extra info which can be passed
// down to the pod list processors (if they need it)
func (p *defaultPodListProcessor) Update(scheduledPods []*apiv1.Pod, allNodes []*apiv1.Node) error {
	// as of now, only FilterOutNoPreemptionPodsListProcessor needs scheduledPods and allNodes
	if _, ok := p.processors[FilterOutNoPreemption]; ok {
		if err := p.processors[FilterOutNoPreemption].Update(scheduledPods, allNodes); err != nil {
			return err
		}
	}

	return nil
}

// Process runs sub-processors sequentially
func (p *defaultPodListProcessor) Process(ctx *context.AutoscalingContext, unschedulablePods []*apiv1.Pod) ([]*apiv1.Pod, error) {
	var err error
	for _, processor := range p.processors {
		unschedulablePods, err = processor.Process(ctx, unschedulablePods)
		if err != nil {
			return nil, err
		}
	}
	return unschedulablePods, nil
}

func (p *defaultPodListProcessor) CleanUp() {
	for _, processor := range p.processors {
		processor.CleanUp()
	}
}
