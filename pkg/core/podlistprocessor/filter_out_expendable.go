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

package podlistprocessor

import (
	apiv1 "k8s.io/api/core/v1"
	ca_context "k8s.io/autoscaler/cluster-autoscaler/context"
	core_utils "k8s.io/autoscaler/cluster-autoscaler/core/utils"
)

type filterOutExpendable struct {
}

// NewFilterOutExpendablePodListProcessor creates a PodListProcessor filtering out expendable pods
func NewFilterOutExpendablePodListProcessor() *filterOutExpendable {
	return &filterOutExpendable{}
}

// Process filters out pods which are expendable and adds pods which is waiting for lower priority pods preemption to the cluster snapshot
func (p *filterOutExpendable) Process(autoscalingCtx *ca_context.AutoscalingContext, pods []*apiv1.Pod) ([]*apiv1.Pod, error) {
	expendablePodsPriorityCutoff := autoscalingCtx.AutoscalingOptions.ExpendablePodsPriorityCutoff

	unschedulablePods := core_utils.FilterOutExpendablePods(pods, expendablePodsPriorityCutoff)

	return unschedulablePods, nil
}

func (p *filterOutExpendable) CleanUp() {
}
