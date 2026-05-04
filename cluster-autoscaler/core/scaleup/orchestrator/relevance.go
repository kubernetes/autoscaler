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

package orchestrator

import (
	"hash/fnv"
	"sort"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/klog/v2"
)

// ApplyRelevanceFilter forks the cluster snapshot, removes nodes that are irrelevant to
// the scheduling of pendingPods, and returns a function to revert the snapshot.
// It is the caller's responsibility to call the returned revert function.
func ApplyRelevanceFilter(snapshot clustersnapshot.ClusterSnapshot, pendingPods []*apiv1.Pod) func() {
	allNodeInfos, err := snapshot.ListNodeInfos()
	if err != nil {
		klog.Errorf("Failed to list NodeInfos for relevance filter: %v", err)
		return func() {}
	}

	relevantNodeInfos := FilterRelevantNodeInfos(pendingPods, allNodeInfos)
	relevantMap := make(map[string]bool)
	for _, ni := range relevantNodeInfos {
		relevantMap[ni.Node().Name] = true
	}

	snapshot.Fork()

	removedCount := 0
	for _, ni := range allNodeInfos {
		if !relevantMap[ni.Node().Name] {
			if err := snapshot.RemoveNodeInfo(ni.Node().Name); err != nil {
				klog.Errorf("Failed to remove irrelevant node %s from snapshot: %v", ni.Node().Name, err)
			} else {
				removedCount++
			}
		}
	}

	klog.V(4).Infof("Relevance Filter: Removed %d irrelevant nodes from ClusterSnapshot (kept %d/%d)", removedCount, len(relevantMap), len(allNodeInfos))

	return func() {
		snapshot.Revert()
	}
}

// FilterRelevantNodeInfos returns a list of NodeInfos that have pods relevant to the scheduling of pendingPods.
// Nodes are deemed relevant if they contain pods that match the affinity, anti-affinity, or topology spread constraints of the pending pods.
func FilterRelevantNodeInfos(pendingPods []*apiv1.Pod, allNodeInfos []*framework.NodeInfo) []*framework.NodeInfo {
	// TODO(x13n): De-duplicate pendingLabels and selectors to avoid redundant matching.
	pendingLabels := make([]labels.Set, 0, len(pendingPods))
	for _, p := range pendingPods {
		pendingLabels = append(pendingLabels, labels.Set(p.Labels))
	}

	var pendingSelectors []labels.Selector
	var selfMatchingHostnameSelectors []labels.Selector
	hasTSC := false

	for _, p := range pendingPods {
		pLabels := labels.Set(p.Labels)
		if p.Spec.Affinity != nil {
			if p.Spec.Affinity.PodAffinity != nil {
				for _, term := range p.Spec.Affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution {
					sel, err := metav1.LabelSelectorAsSelector(term.LabelSelector)
					if err != nil {
						continue
					}
					if term.TopologyKey == apiv1.LabelHostname {
						if sel.Matches(pLabels) {
							selfMatchingHostnameSelectors = append(selfMatchingHostnameSelectors, sel)
						}
						continue
					}
					pendingSelectors = append(pendingSelectors, sel)
				}
			}
			if p.Spec.Affinity.PodAntiAffinity != nil {
				for _, term := range p.Spec.Affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution {
					// Hostname topology anti-affinities don't matter in ScaleUp (new nodes are empty).
					if term.TopologyKey == apiv1.LabelHostname {
						continue
					}
					if term.LabelSelector != nil {
						if sel, err := metav1.LabelSelectorAsSelector(term.LabelSelector); err == nil {
							pendingSelectors = append(pendingSelectors, sel)
						}
					}
				}
			}
		}
		for _, tsc := range p.Spec.TopologySpreadConstraints {
			if tsc.TopologyKey == apiv1.LabelHostname {
				continue
			}
			hasTSC = true
			if tsc.LabelSelector != nil {
				if sel, err := metav1.LabelSelectorAsSelector(tsc.LabelSelector); err == nil {
					pendingSelectors = append(pendingSelectors, sel)
				}
			}
		}
	}

	var relevant []*framework.NodeInfo
	satisfiedSelfMatches := make([]bool, len(selfMatchingHostnameSelectors))
	preservedSignatures := make(map[uint64]bool)

	for _, ni := range allNodeInfos {
		hasRelevantPod := false
		for _, pi := range ni.Pods() {
			p := pi.Pod
			podLabels := labels.Set(p.Labels)

			// 1. Regular non-hostname selectors
			for _, sel := range pendingSelectors {
				if sel.Matches(podLabels) {
					hasRelevantPod = true
					break
				}
			}
			if hasRelevantPod {
				break
			}

			// 2. Self-matching hostname selectors (at least one node per selector)
			for i, sel := range selfMatchingHostnameSelectors {
				if !satisfiedSelfMatches[i] && sel.Matches(podLabels) {
					hasRelevantPod = true
					satisfiedSelfMatches[i] = true
				}
			}

			// 3. Does this node's pod have ANTI-affinity TO our pending pods?
			if p.Spec.Affinity != nil && matchesAnyAntiOptimized(p.Spec.Affinity.PodAntiAffinity, pendingLabels) {
				hasRelevantPod = true
				break
			}
		}

		// 4. Preserve representative nodes for TSC domains including across subsets of nodes.
		// This is required to ensure no topology domains will be pruned.
		isNeededForTSC := false
		var sig uint64
		if hasTSC {
			sig = nodeSignature(ni.Node())
			if !hasRelevantPod && !preservedSignatures[sig] {
				isNeededForTSC = true
			}
		}

		if hasRelevantPod || isNeededForTSC {
			if hasTSC {
				preservedSignatures[sig] = true
			}
			relevant = append(relevant, ni)
		}
	}
	return relevant
}

func matchesAnyAntiOptimized(affinity *apiv1.PodAntiAffinity, labelsList []labels.Set) bool {
	if affinity == nil {
		return false
	}
	for _, term := range affinity.RequiredDuringSchedulingIgnoredDuringExecution {
		// Hostname topology affinities don't actually matter in ScaleUp, so they can be effectively ignored.
		if term.TopologyKey == apiv1.LabelHostname {
			continue
		}
		if term.LabelSelector != nil {
			if sel, err := metav1.LabelSelectorAsSelector(term.LabelSelector); err == nil {
				for _, l := range labelsList {
					if sel.Matches(l) {
						return true
					}
				}
			}
		}
	}
	return false
}

func nodeSignature(node *apiv1.Node) uint64 {
	h := fnv.New64a()

	// Sort labels to ensure deterministic order
	keys := make([]string, 0, len(node.Labels))
	for k := range node.Labels {
		if k == apiv1.LabelHostname {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		h.Write([]byte(k))
		h.Write([]byte{0}) // delimiter
		h.Write([]byte(node.Labels[k]))
		h.Write([]byte{0})
	}

	// Sort taints
	taints := make([]apiv1.Taint, len(node.Spec.Taints))
	copy(taints, node.Spec.Taints)
	sort.Slice(taints, func(i, j int) bool {
		return taints[i].Key < taints[j].Key
	})

	for _, t := range taints {
		h.Write([]byte(t.Key))
		h.Write([]byte{0})
		h.Write([]byte(t.Value))
		h.Write([]byte{0})
		h.Write([]byte(t.Effect))
		h.Write([]byte{0})
	}

	return h.Sum64()
}
