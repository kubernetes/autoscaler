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

package podlistprocessor

import (
	"sort"
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/metrics"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/scheduling"
	corev1helpers "k8s.io/component-helpers/scheduling/corev1"
	klog "k8s.io/klog/v2"
)

type filterOutSchedulablePodListProcessor struct {
	schedulingSimulator *scheduling.HintingSimulator
	nodeFilter          func(*framework.NodeInfo) bool
}

// NewFilterOutSchedulablePodListProcessor creates a PodListProcessor filtering out schedulable pods
func NewFilterOutSchedulablePodListProcessor(nodeFilter func(*framework.NodeInfo) bool) *filterOutSchedulablePodListProcessor {
	return &filterOutSchedulablePodListProcessor{
		schedulingSimulator: scheduling.NewHintingSimulator(),
		nodeFilter:          nodeFilter,
	}
}

// Process filters out pods which are schedulable from list of unschedulable pods.
func (p *filterOutSchedulablePodListProcessor) Process(context *context.AutoscalingContext, unschedulablePods []*apiv1.Pod) ([]*apiv1.Pod, error) {
	// We need to check whether pods marked as unschedulable are actually unschedulable.
	// It's likely we added a new node and the scheduler just haven't managed to put the
	// pod on in yet. In this situation we don't want to trigger another scale-up.
	//
	// It's also important to prevent uncontrollable cluster growth if CA's simulated
	// scheduler differs in opinion with real scheduler. Example of such situation:
	// - CA and Scheduler has slightly different configuration
	// - Scheduler can't schedule a pod and marks it as unschedulable
	// - CA added a node which should help the pod
	// - Scheduler doesn't schedule the pod on the new node
	//   because according to it logic it doesn't fit there
	// - CA see the pod is still unschedulable, so it adds another node to help it
	//
	// With the check enabled the last point won't happen because CA will ignore a pod
	// which is supposed to schedule on an existing node.

	klog.V(4).Infof("Filtering out schedulables")
	filterOutSchedulableStart := time.Now()

	unschedulablePodsToHelp, err := p.filterOutSchedulableByPacking(unschedulablePods, context.ClusterSnapshot)

	if err != nil {
		return nil, err
	}

	metrics.UpdateDurationFromStart(metrics.FilterOutSchedulable, filterOutSchedulableStart)

	if len(unschedulablePodsToHelp) != len(unschedulablePods) {
		klog.V(2).Info("Schedulable pods present")

		if context.DebuggingSnapshotter.IsDataCollectionAllowed() {
			schedulablePods := findSchedulablePods(unschedulablePods, unschedulablePodsToHelp)
			context.DebuggingSnapshotter.SetUnscheduledPodsCanBeScheduled(schedulablePods)
		}

	} else {
		klog.V(4).Info("No schedulable pods")
	}
	return unschedulablePodsToHelp, nil
}

func (p *filterOutSchedulablePodListProcessor) CleanUp() {
}

// filterOutSchedulableByPacking checks whether pods from <unschedulableCandidates> marked as
// unschedulable can be scheduled on free capacity on existing nodes by trying to pack the pods. It
// tries to pack the higher priority pods first. It takes into account pods that are bound to node
// and will be scheduled after lower priority pod preemption.
func (p *filterOutSchedulablePodListProcessor) filterOutSchedulableByPacking(unschedulableCandidates []*apiv1.Pod, clusterSnapshot clustersnapshot.ClusterSnapshot) ([]*apiv1.Pod, error) {
	// Sort unschedulable pods by importance
	sort.Slice(unschedulableCandidates, func(i, j int) bool {
		return corev1helpers.PodPriority(unschedulableCandidates[i]) > corev1helpers.PodPriority(unschedulableCandidates[j])
	})

	statuses, overflowingControllerCount, err := p.schedulingSimulator.TrySchedulePods(clusterSnapshot, unschedulableCandidates, p.nodeFilter, false)
	if err != nil {
		return nil, err
	}

	scheduledPods := make(map[types.UID]bool)
	for _, status := range statuses {
		scheduledPods[status.Pod.UID] = true
	}

	// Pods that remain unschedulable
	var unschedulablePods []*apiv1.Pod
	for _, pod := range unschedulableCandidates {
		if !scheduledPods[pod.UID] {
			unschedulablePods = append(unschedulablePods, pod)
		}
	}

	metrics.UpdateOverflowingControllers(overflowingControllerCount)
	klog.V(4).Infof("%v pods marked as unschedulable can be scheduled.", len(unschedulableCandidates)-len(unschedulablePods))

	p.schedulingSimulator.DropOldHints()
	return unschedulablePods, nil
}

func findSchedulablePods(allUnschedulablePods, podsStillUnschedulable []*apiv1.Pod) []*apiv1.Pod {
	podsStillUnschedulableMap := make(map[*apiv1.Pod]struct{}, len(podsStillUnschedulable))
	for _, x := range podsStillUnschedulable {
		podsStillUnschedulableMap[x] = struct{}{}
	}
	var schedulablePods []*apiv1.Pod
	for _, x := range allUnschedulablePods {
		if _, found := podsStillUnschedulableMap[x]; !found {
			schedulablePods = append(schedulablePods, x)
		}
	}
	return schedulablePods
}
