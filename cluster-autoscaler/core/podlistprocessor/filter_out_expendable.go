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
	"fmt"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	core_utils "k8s.io/autoscaler/cluster-autoscaler/core/utils"
	caerrors "k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/klog/v2"
)

type filterOutExpendable struct {
}

// NewFilterOutExpendablePodListProcessor creates a PodListProcessor filtering out expendable pods
func NewFilterOutExpendablePodListProcessor() *filterOutExpendable {
	return &filterOutExpendable{}
}

// Process filters out pods which are expendable and adds pods which is waiting for lower priority pods preemption to the cluster snapshot
func (p *filterOutExpendable) Process(context *context.AutoscalingContext, pods []*apiv1.Pod) ([]*apiv1.Pod, error) {
	nodes, err := context.AllNodeLister().List()
	if err != nil {
		return nil, fmt.Errorf("Failed to list all nodes while filtering expendable pods: %v", err)
	}
	expendablePodsPriorityCutoff := context.AutoscalingOptions.ExpendablePodsPriorityCutoff

	unschedulablePods, waitingForLowerPriorityPreemption := core_utils.FilterOutExpendableAndSplit(pods, nodes, expendablePodsPriorityCutoff)
	if err = p.addPreemptingPodsToSnapshot(waitingForLowerPriorityPreemption, context); err != nil {
		klog.Warningf("Failed to add preempting pods to snapshot: %v", err)
		return nil, err
	}

	return unschedulablePods, nil
}

// addPreemptingPodsToSnapshot modifies the snapshot simulating scheduling of pods waiting for preemption.
// this is not strictly correct as we are not simulating preemption itself but it matches
// CA logic from before migration to scheduler framework. So let's keep it for now
func (p *filterOutExpendable) addPreemptingPodsToSnapshot(pods []*apiv1.Pod, ctx *context.AutoscalingContext) error {
	for _, p := range pods {
		// TODO(DRA): Figure out if/how to use the predicate-checking SchedulePod() here instead - otherwise this doesn't work with DRA pods.
		if err := ctx.ClusterSnapshot.ForceAddPod(p, p.Status.NominatedNodeName); err != nil {
			klog.Errorf("Failed to update snapshot with pod %s/%s waiting for preemption: %v", p.Namespace, p.Name, err)
			return caerrors.ToAutoscalerError(caerrors.InternalError, err)
		}
	}
	return nil
}

func (p *filterOutExpendable) CleanUp() {
}
