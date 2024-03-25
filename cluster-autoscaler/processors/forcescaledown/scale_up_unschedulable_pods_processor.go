/*
Copyright 2024 The Kubernetes Authors.

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

package forcescaledown

import (
	"fmt"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/scheduling"
	pod_util "k8s.io/autoscaler/cluster-autoscaler/utils/pod"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"
	"k8s.io/klog/v2"
)

// ScaleUpUnschedulablePodsProcessor is a processor to handle unschedulable pods on force-scale-down nodes.
type ScaleUpUnschedulablePodsProcessor struct{}

// NewScaleUpUnschedulablePodsProcessor returns a new processor adding pods
// from currently force-scale-down nodes to the unschedulable pods.
func NewScaleUpUnschedulablePodsProcessor() *ScaleUpUnschedulablePodsProcessor {
	return &ScaleUpUnschedulablePodsProcessor{}
}

// Process adds recreatable and unschedulable pods from currently force-scale-down nodes
func (p *ScaleUpUnschedulablePodsProcessor) Process(context *context.AutoscalingContext, unschedulablePods []*apiv1.Pod) ([]*apiv1.Pod, error) {
	klog.V(4).Infof("Processing currently force-scale-down nodes and pods")
	forceScaleDownPods, err := getCurrentlyForceScaleDownPods(context)
	if err != nil {
		return nil, fmt.Errorf("failed to get currently force-scale-down pods: %w", err)
	}
	forceScaleDownPods = pod_util.ClearPodNodeNames(forceScaleDownPods)
	recreatablePods := pod_util.FilterRecreatablePods(forceScaleDownPods)
	unschedulableForceScaleDownPods, err := filterOutSchedulablePodsToAvoidNodeChurn(context, recreatablePods)
	if err != nil {
		return nil, fmt.Errorf("failed to filter out schedulable force-scale-down pods: %w", err)
	}
	allUnschedulablePods := appendAdditionalPods(unschedulablePods, unschedulableForceScaleDownPods)
	klog.V(4).Infof("Pod count summary: all unschedulable %d, initial unschedulable %d, all force-scale-down %d, recreatable force-scale-down %d, unschedulable force-scale-down %d",
		len(allUnschedulablePods), len(unschedulablePods), len(forceScaleDownPods), len(recreatablePods), len(unschedulableForceScaleDownPods))
	return allUnschedulablePods, nil
}

// CleanUp cleans up the processor's internal structures.
func (p *ScaleUpUnschedulablePodsProcessor) CleanUp() {}

func getCurrentlyForceScaleDownPods(context *context.AutoscalingContext) ([]*apiv1.Pod, error) {
	var pods []*apiv1.Pod
	nodes, err := context.AllNodeLister().List()
	if err != nil {
		return nil, fmt.Errorf("failed to list all nodes when getting force-scale-down nodes: %w", err)
	}
	for _, node := range nodes {
		if !taints.HasForceScaleDownTaint(node) {
			continue
		}
		nodeInfo, err := context.ClusterSnapshot.NodeInfos().Get(node.Name)
		if err != nil {
			klog.Warningf("Couldn't get node %s info, assuming the node got deleted already: %v", node.Name, err)
			continue
		}
		klog.V(2).Infof("Found %d pods on force-scale-down node %s", len(nodeInfo.Pods), node.Name)
		for _, podInfo := range nodeInfo.Pods {
			if podInfo.Pod.DeletionTimestamp == nil {
				pods = append(pods, podInfo.Pod)
			}
		}
	}
	return pods, nil
}

func filterOutSchedulablePodsToAvoidNodeChurn(context *context.AutoscalingContext, pods []*apiv1.Pod) ([]*apiv1.Pod, error) {
	context.ClusterSnapshot.Fork()
	defer context.ClusterSnapshot.Revert()
	simulator := scheduling.NewHintingSimulator(context.PredicateChecker)
	statuses, _, err := simulator.TrySchedulePods(context.ClusterSnapshot, pods, scheduling.ScheduleAnywhere, false)
	if err != nil {
		return nil, fmt.Errorf("failed to simulate scheduling force-scale-down pods: %w", err)
	}
	scheduledPodUIDs := map[types.UID]bool{}
	for _, status := range statuses {
		scheduledPodUIDs[status.Pod.UID] = true
	}
	unschedulablePods := []*apiv1.Pod{}
	for _, pod := range pods {
		if scheduledPodUIDs[pod.UID] {
			continue
		}
		unschedulablePods = append(unschedulablePods, pod)
	}
	return unschedulablePods, nil
}

func appendAdditionalPods(existingPods, additionalPods []*apiv1.Pod) []*apiv1.Pod {
	result := []*apiv1.Pod{}
	result = append(result, existingPods...)
	existingPodUIDs := map[types.UID]bool{}
	for _, pod := range existingPods {
		existingPodUIDs[pod.UID] = true
	}
	for _, pod := range additionalPods {
		if existingPodUIDs[pod.UID] {
			continue
		}
		result = append(result, pod)
	}
	return result
}
