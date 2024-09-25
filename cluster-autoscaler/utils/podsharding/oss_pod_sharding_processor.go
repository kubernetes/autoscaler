/*
Copyright 2023 The Kubernetes Authors.

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

package podsharding

import (
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	klog "k8s.io/klog/v2"
)

const (
	selectPodShardContextKey = "selected-pod-shard.podsharding.cluster-autoscaler"
)

// PodShardingProcessor is a processor for pre-sharding unschedulable pods. It uses given PodSharder
// and PodShardSelector to group pods into independent scale-up groups and select one of those for given loop iteration.
type PodShardingProcessor struct {
	podSharder       PodSharder
	podShardSelector PodShardSelector
	podShardFilter   PodShardFilter
}

// GetSelectedPodShard returns selected pod shard.
func GetSelectedPodShard(context *context.AutoscalingContext) *PodShard {
	value, found := context.ProcessorCallbacks.GetExtraValue(selectPodShardContextKey)
	if !found {
		return nil
	}
	shard, ok := value.(*PodShard)
	if !ok {
		klog.Errorf("Expected *PodShard as value for %v; got %T", selectPodShardContextKey, value)
		return nil
	}
	return shard
}

// NewPodShardingProcessor creates new instance of PodShardingProcessor
func NewPodShardingProcessor(
	podSharder PodSharder,
	podShardSelector PodShardSelector,
	podShardFilter PodShardFilter) *PodShardingProcessor {
	return &PodShardingProcessor{
		podSharder:       podSharder,
		podShardSelector: podShardSelector,
		podShardFilter:   podShardFilter,
	}
}

// Process executes pod sharding logic for passed unschedulablePods. Pods are sharded and one of the shards is selected.
// The allScheduledPods slice is returned not changed.
func (p *PodShardingProcessor) Process(context *context.AutoscalingContext,
	unschedulablePods []*apiv1.Pod) ([]*apiv1.Pod, error) {

	if len(unschedulablePods) == 0 {
		// nothing to be done
		return unschedulablePods, nil
	}

	podShards := p.podSharder.ComputePodShards(unschedulablePods)

	podShard := p.podShardSelector.SelectPodShard(podShards)

	if podShard == nil {
		return []*apiv1.Pod{}, errors.NewAutoscalerError(errors.InternalError, "No shard selected")
	}

	filteringResult, err := p.podShardFilter.FilterPods(context, podShard, podShards, unschedulablePods)
	if err != nil {
		return []*apiv1.Pod{}, errors.ToAutoscalerError(errors.InternalError, err)
	}
	klog.Infof("Selected pods shard %v; NodeGroupDescriptor=%#v; shardPodsCount=%v; extendedPodsCount=%v",
		podShard.Signature(), podShard.NodeGroupDescriptor, len(podShard.PodUids), len(filteringResult.Pods))

	context.ProcessorCallbacks.SetExtraValue(selectPodShardContextKey, podShard)

	return filteringResult.Pods, nil
}

// CleanUp does nothing for PodShardingProcessor
func (p *PodShardingProcessor) CleanUp() {
}
