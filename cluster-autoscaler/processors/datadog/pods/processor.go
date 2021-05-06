/*
Copyright 2021 The Kubernetes Authors.

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
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/metrics"
	proc "k8s.io/autoscaler/cluster-autoscaler/processors/pods"

	apiv1 "k8s.io/api/core/v1"
	klog "k8s.io/klog/v2"
)

type filteringPodListProcessor struct {
	transforms []proc.PodListProcessor
	filters    []proc.PodListProcessor
}

// NewFilteringPodListProcessor returns an aggregated podlist processor
func NewFilteringPodListProcessor() *filteringPodListProcessor {
	return &filteringPodListProcessor{
		transforms: []proc.PodListProcessor{
			NewTransformLocalData(),
		},
		filters: []proc.PodListProcessor{
			NewFilterOutSchedulablePodListProcessor(),
		},
	}
}

// CleanUp tears down all aggreated podlist processors
func (p *filteringPodListProcessor) CleanUp() {
	for _, transform := range p.transforms {
		transform.CleanUp()
	}
	for _, filter := range p.filters {
		filter.CleanUp()
	}
}

// Process runs all podlists processors
func (p *filteringPodListProcessor) Process(ctx *context.AutoscalingContext, pending []*apiv1.Pod) ([]*apiv1.Pod, error) {
	klog.V(4).Infof("Filtering pending pods")
	start := time.Now()

	var err error

	for _, transform := range p.transforms {
		pending, err = transform.Process(ctx, pending)
		if err != nil {
			return nil, err
		}
	}

	unschedulablePodsToHelp := make([]*apiv1.Pod, len(pending))
	copy(unschedulablePodsToHelp, pending)
	for _, filter := range p.filters {
		unschedulablePodsToHelp, err = filter.Process(ctx, unschedulablePodsToHelp)
		if err != nil {
			return nil, err
		}
	}

	metrics.UpdateDurationFromStart(metrics.FilterOutSchedulable, start)
	return unschedulablePodsToHelp, nil
}
