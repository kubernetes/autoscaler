/*
Copyright 2026 The Kubernetes Authors.

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

package karpenter

import (
	"context"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	podutils "sigs.k8s.io/karpenter/pkg/utils/pod"
)

// HydrateClusterState filters relevant nodes and pods from the ClusterSnapshot.
// It performs Topology Pruning to reduce the number of pods and nodes Karpenter needs to evaluate.
func HydrateClusterState(ctx context.Context, snapshot clustersnapshot.ClusterSnapshot, nodeInfos []*framework.NodeInfo, podsToSchedule []*apiv1.Pod) ([]*apiv1.Pod, []*apiv1.Node, error) {
	relevantSelectors := []labels.Selector{}
	relevantLabels := make(map[string]sets.Set[string])

	// 1. Identify relevant labels and selectors from pods to schedule
	for _, p := range podsToSchedule {
		for k, v := range p.Labels {
			if _, ok := relevantLabels[k]; !ok {
				relevantLabels[k] = sets.New[string]()
			}
			relevantLabels[k].Insert(v)
		}
		if p.Spec.Affinity != nil && p.Spec.Affinity.PodAntiAffinity != nil {
			for _, term := range p.Spec.Affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution {
				sel, err := metav1.LabelSelectorAsSelector(term.LabelSelector)
				if err == nil {
					relevantSelectors = append(relevantSelectors, sel)
				}
			}
		}
		for _, tsc := range p.Spec.TopologySpreadConstraints {
			sel, err := metav1.LabelSelectorAsSelector(tsc.LabelSelector)
			if err == nil {
				relevantSelectors = append(relevantSelectors, sel)
			}
		}
	}

	var allRelevantPods []*apiv1.Pod
	var allRelevantNodes []*apiv1.Node

	// 2. Filter only relevant nodes and pods
	for _, ni := range nodeInfos {
		node := ni.Node()
		var relevantPodsOnNode []*apiv1.Pod
		hasRelevantPod := false

		for _, pi := range ni.Pods() {
			pod := pi.Pod
			isRelevant := false

			for _, sel := range relevantSelectors {
				if sel.Matches(labels.Set(pod.Labels)) {
					isRelevant = true
					break
				}
			}

			if !isRelevant && (podutils.HasRequiredPodAntiAffinity(pod) || hasTopologySpread(pod)) {
				if podMatchesSurgeLabels(pod, relevantLabels) {
					isRelevant = true
				}
			}

			if isRelevant {
				relevantPodsOnNode = append(relevantPodsOnNode, pod)
				hasRelevantPod = true
			}
		}

		if hasRelevantPod || !nodeIsFull(ni) {
			allRelevantNodes = append(allRelevantNodes, node)
			for _, pi := range ni.Pods() {
				allRelevantPods = append(allRelevantPods, pi.Pod)
			}
		}
	}

	return allRelevantPods, allRelevantNodes, nil
}

func hasTopologySpread(p *apiv1.Pod) bool {
	return len(p.Spec.TopologySpreadConstraints) > 0
}

func podMatchesSurgeLabels(p *apiv1.Pod, surgeLabels map[string]sets.Set[string]) bool {
	if p.Spec.Affinity != nil && p.Spec.Affinity.PodAntiAffinity != nil {
		for _, term := range p.Spec.Affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution {
			if termMatchesSurgeLabels(term.LabelSelector, surgeLabels) {
				return true
			}
		}
	}
	for _, tsc := range p.Spec.TopologySpreadConstraints {
		if termMatchesSurgeLabels(tsc.LabelSelector, surgeLabels) {
			return true
		}
	}
	return false
}

func termMatchesSurgeLabels(ls *metav1.LabelSelector, surgeLabels map[string]sets.Set[string]) bool {
	if ls == nil {
		return true // matches everything
	}
	for k := range ls.MatchLabels {
		if _, ok := surgeLabels[k]; ok {
			return true
		}
	}
	for _, expr := range ls.MatchExpressions {
		if _, ok := surgeLabels[expr.Key]; ok {
			return true
		}
	}
	return false
}

func nodeIsFull(ni *framework.NodeInfo) bool {
	if ni == nil {
		return true
	}
	node := ni.Node()
	if node == nil {
		return true
	}
	res := ni.Requested
	if res == nil {
		return false
	}
	cap := node.Status.Allocatable
	
	if cap.Cpu().MilliValue() > 0 && (cap.Cpu().MilliValue() - res.MilliCPU) > cap.Cpu().MilliValue() / 100 {
		return false
	}
	if cap.Memory().Value() > 0 && (cap.Memory().Value() - res.Memory) > cap.Memory().Value() / 100 {
		return false
	}
	return true
}
