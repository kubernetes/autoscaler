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

package core

import (
	"sort"
	"time"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/core/utils"
	"k8s.io/autoscaler/cluster-autoscaler/metrics"
	"k8s.io/autoscaler/cluster-autoscaler/processors/pods"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/klog"
	"k8s.io/kubernetes/pkg/api/v1/pod"
)

type filterOutSchedulablePodListProcessor struct{}

// NewFilterOutSchedulablePodListProcessor creates a PodListProcessor filtering out schedulable pods
func NewFilterOutSchedulablePodListProcessor() pods.PodListProcessor {
	return &filterOutSchedulablePodListProcessor{}
}

// Process filters out pods which are schedulable from list of unschedulable pods.
func (filterOutSchedulablePodListProcessor) Process(
	context *context.AutoscalingContext,
	unschedulablePods []*apiv1.Pod) ([]*apiv1.Pod, error) {
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
	var unschedulablePodsToHelp []*apiv1.Pod

	unschedulablePodsToHelp, err := filterOutSchedulableByPacking(unschedulablePods, context.ClusterSnapshot,
		context.PredicateChecker, context.ExpendablePodsPriorityCutoff)

	if err != nil {
		return nil, err
	}

	metrics.UpdateDurationFromStart(metrics.FilterOutSchedulable, filterOutSchedulableStart)

	if len(unschedulablePodsToHelp) != len(unschedulablePods) {
		klog.V(2).Info("Schedulable pods present")
		context.ProcessorCallbacks.DisableScaleDownForLoop()
	} else {
		klog.V(4).Info("No schedulable pods")
	}
	return unschedulablePodsToHelp, nil
}

func (filterOutSchedulablePodListProcessor) CleanUp() {
}

// filterOutSchedulableByPacking checks whether pods from <unschedulableCandidates> marked as
// unschedulable can be scheduled on free capacity on existing nodes by trying to pack the pods. It
// tries to pack the higher priority pods first. It takes into account pods that are bound to node
// and will be scheduled after lower priority pod preemption.
func filterOutSchedulableByPacking(
	unschedulableCandidates []*apiv1.Pod,
	clusterSnapshot simulator.ClusterSnapshot,
	predicateChecker simulator.PredicateChecker,
	expendablePodsPriorityCutoff int) ([]*apiv1.Pod, error) {

	allScheduled, err := clusterSnapshot.Pods().List(labels.Everything())
	if err != nil {
		return nil, err
	}

	// Sort unschedulable pods by importance
	sort.Slice(unschedulableCandidates, func(i, j int) bool {
		return moreImportantPod(unschedulableCandidates[i], unschedulableCandidates[j])
	})

	// Remove all expendable pods snapshot
	// TODO(scheduler_framework_integration) distinguish move to separate pod list processor
	for _, pod := range allScheduled {
		if utils.IsExpendablePod(pod, expendablePodsPriorityCutoff) {
			klog.V(4).Infof("Removing expandable pod %s.%s", pod.Namespace, pod.Name)
			if err := clusterSnapshot.RemovePod(pod.Namespace, pod.Name, pod.Spec.NodeName); err != nil {
				return nil, err
			}
		}
	}

	// Pods which remain unschedulable
	var unschedulablePods []*apiv1.Pod

	// Bin pack
	for _, pod := range unschedulableCandidates {
		nodeName, err := predicateChecker.FitsAnyNode(clusterSnapshot, pod)
		if err == nil {
			klog.V(4).Infof("Pod %s.%s marked as unschedulable can be scheduled on node %s. Ignoring"+
				" in scale up.", pod.Namespace, pod.Name, nodeName)
			if err := clusterSnapshot.AddPod(pod, nodeName); err != nil {
				return nil, err
			}
		} else {
			unschedulablePods = append(unschedulablePods, pod)
		}
	}

	klog.V(4).Infof("%v pods marked as unschedulable can be scheduled.", len(unschedulableCandidates)-len(unschedulablePods))
	return unschedulablePods, nil
}

func moreImportantPod(pod1, pod2 *apiv1.Pod) bool {
	// based on schedulers MoreImportantPod but does not compare Pod.Status.StartTime which does not make sense
	// for unschedulable pods
	p1 := pod.GetPodPriority(pod1)
	p2 := pod.GetPodPriority(pod2)
	return p1 > p2
}
