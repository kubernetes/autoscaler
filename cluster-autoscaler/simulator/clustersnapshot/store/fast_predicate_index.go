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

package store

import (
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/common"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
)

type podCountKey struct {
	labelSetHash  string
	topologyKey   string
	topologyValue string
}

type nodeCountKey struct {
	topologyKey   string
	topologyValue string
}

type fastPredicateIndex struct {
	// (labelSetHash, topologyKey, topologyValue) -> count
	podCounts *common.PatchSet[podCountKey, int]
	// labelSetHash -> labels (immutable)
	// Ref-counted: entries are removed when no more pods use them.
	hashToLabels *common.PatchSet[string, map[string]string]
	// labelSetHash -> total count of pods across all nodes
	labelSetCounts *common.PatchSet[string, int]
	// (topologyKey, topologyValue) -> count of nodes
	nodeCounts *common.PatchSet[nodeCountKey, int]
	// topologyKey -> set of topology values
	topologyDomains *common.PatchSet[string, map[string]bool]
}

func newFastPredicateIndex() *fastPredicateIndex {
	return &fastPredicateIndex{
		podCounts:       common.NewPatchSet(common.NewPatch[podCountKey, int]()),
		hashToLabels:    common.NewPatchSet(common.NewPatch[string, map[string]string]()),
		labelSetCounts:  common.NewPatchSet(common.NewPatch[string, int]()),
		nodeCounts:      common.NewPatchSet(common.NewPatch[nodeCountKey, int]()),
		topologyDomains: common.NewPatchSet(common.NewPatch[string, map[string]bool]()),
	}
}

func (idx *fastPredicateIndex) Fork() {
	idx.podCounts.Fork()
	idx.hashToLabels.Fork()
	idx.labelSetCounts.Fork()
	idx.nodeCounts.Fork()
	idx.topologyDomains.Fork()
}

func (idx *fastPredicateIndex) Revert() {
	idx.podCounts.Revert()
	idx.hashToLabels.Revert()
	idx.labelSetCounts.Revert()
	idx.nodeCounts.Revert()
	idx.topologyDomains.Revert()
}

func (idx *fastPredicateIndex) Commit() {
	idx.podCounts.Commit()
	idx.hashToLabels.Commit()
	idx.labelSetCounts.Commit()
	idx.nodeCounts.Commit()
	idx.topologyDomains.Commit()
}

func (idx *fastPredicateIndex) addPod(pod *apiv1.Pod, node *apiv1.Node) {
	if pod == nil || node == nil {
		return
	}
	labelSetHash := framework.GetPodLabelSetHash(pod)

	// Global ref-counting for labels to avoid unbounded growth.
	globalCount, _ := idx.labelSetCounts.FindValue(labelSetHash)
	if globalCount == 0 {
		labels := make(map[string]string)
		for k, v := range pod.Labels {
			labels[k] = v
		}
		idx.hashToLabels.SetCurrent(labelSetHash, labels)
	}
	idx.labelSetCounts.SetCurrent(labelSetHash, globalCount+1)

	topologyKeys := make(map[string]string)
	for k, v := range node.Labels {
		topologyKeys[k] = v
	}
	topologyKeys[apiv1.LabelHostname] = node.Name

	for tk, tv := range topologyKeys {
		key := podCountKey{labelSetHash: labelSetHash, topologyKey: tk, topologyValue: tv}
		count, _ := idx.podCounts.FindValue(key)
		idx.podCounts.SetCurrent(key, count+1)
	}
}

func (idx *fastPredicateIndex) removePod(pod *apiv1.Pod, node *apiv1.Node) {
	if pod == nil || node == nil {
		return
	}
	labelSetHash := framework.GetPodLabelSetHash(pod)

	// Global ref-counting for labels.
	globalCount, found := idx.labelSetCounts.FindValue(labelSetHash)
	if found && globalCount > 0 {
		if globalCount == 1 {
			idx.labelSetCounts.DeleteCurrent(labelSetHash)
			idx.hashToLabels.DeleteCurrent(labelSetHash)
		} else {
			idx.labelSetCounts.SetCurrent(labelSetHash, globalCount-1)
		}
	}

	topologyKeys := make(map[string]string)
	for k, v := range node.Labels {
		topologyKeys[k] = v
	}
	topologyKeys[apiv1.LabelHostname] = node.Name

	for tk, tv := range topologyKeys {
		key := podCountKey{labelSetHash: labelSetHash, topologyKey: tk, topologyValue: tv}
		count, found := idx.podCounts.FindValue(key)
		if found && count > 0 {
			if count == 1 {
				idx.podCounts.DeleteCurrent(key)
			} else {
				idx.podCounts.SetCurrent(key, count-1)
			}
		}
	}
}

func (idx *fastPredicateIndex) addNode(node *apiv1.Node) {
	if node == nil {
		return
	}

	topologyKeys := make(map[string]string)
	for k, v := range node.Labels {
		topologyKeys[k] = v
	}
	topologyKeys[apiv1.LabelHostname] = node.Name

	for tk, tv := range topologyKeys {
		key := nodeCountKey{topologyKey: tk, topologyValue: tv}
		count, _ := idx.nodeCounts.FindValue(key)
		idx.nodeCounts.SetCurrent(key, count+1)

		if count == 0 {
			domains, _ := idx.topologyDomains.FindValue(tk)
			newDomains := make(map[string]bool)
			for d := range domains {
				newDomains[d] = true
			}
			newDomains[tv] = true
			idx.topologyDomains.SetCurrent(tk, newDomains)
		}
	}
}

func (idx *fastPredicateIndex) removeNode(node *apiv1.Node) {
	if node == nil {
		return
	}

	topologyKeys := make(map[string]string)
	for k, v := range node.Labels {
		topologyKeys[k] = v
	}
	topologyKeys[apiv1.LabelHostname] = node.Name

	for tk, tv := range topologyKeys {
		key := nodeCountKey{topologyKey: tk, topologyValue: tv}
		count, found := idx.nodeCounts.FindValue(key)
		if found && count > 0 {
			if count == 1 {
				idx.nodeCounts.DeleteCurrent(key)

				domains, _ := idx.topologyDomains.FindValue(tk)
				newDomains := make(map[string]bool)
				for d := range domains {
					newDomains[d] = true
				}
				delete(newDomains, tv)
				if len(newDomains) == 0 {
					idx.topologyDomains.DeleteCurrent(tk)
				} else {
					idx.topologyDomains.SetCurrent(tk, newDomains)
				}
			} else {
				idx.nodeCounts.SetCurrent(key, count-1)
			}
		}
	}
}

func (idx *fastPredicateIndex) PodAffinityCount(topologyKey, topologyValue string, selector labels.Selector) int {
	count := 0
	// We iterate over the current view of unique label sets.
	// Size is O(UniqueLabelSets), which is usually small.
	for hash, labelsMap := range idx.hashToLabels.AsMap() {
		if selector.Matches(labels.Set(labelsMap)) {
			key := podCountKey{labelSetHash: hash, topologyKey: topologyKey, topologyValue: topologyValue}
			c, _ := idx.podCounts.FindValue(key)
			count += c
		}
	}
	return count
}

func (idx *fastPredicateIndex) TopologyValueCount(topologyKey string) int {
	domains, found := idx.topologyDomains.FindValue(topologyKey)
	if !found {
		return 0
	}
	return len(domains)
}

func (idx *fastPredicateIndex) TopologyDomains(topologyKey string) []string {
	domainsMap, found := idx.topologyDomains.FindValue(topologyKey)
	if !found {
		return nil
	}
	domains := make([]string, 0, len(domainsMap))
	for d := range domainsMap {
		domains = append(domains, d)
	}
	return domains
}

func (idx *fastPredicateIndex) clone() *fastPredicateIndex {
	newIdx := &fastPredicateIndex{
		podCounts: common.ClonePatchSet(idx.podCounts,
			func(k podCountKey) podCountKey { return k },
			func(v int) int { return v }),
		hashToLabels: common.ClonePatchSet(idx.hashToLabels,
			func(k string) string { return k },
			func(v map[string]string) map[string]string {
				newMap := make(map[string]string)
				for k1, v1 := range v {
					newMap[k1] = v1
				}
				return newMap
			}),
		labelSetCounts: common.ClonePatchSet(idx.labelSetCounts,
			func(k string) string { return k },
			func(v int) int { return v }),
		nodeCounts: common.ClonePatchSet(idx.nodeCounts,
			func(k nodeCountKey) nodeCountKey { return k },
			func(v int) int { return v }),
		topologyDomains: common.ClonePatchSet(idx.topologyDomains,
			func(k string) string { return k },
			func(v map[string]bool) map[string]bool {
				newMap := make(map[string]bool)
				for k1, v1 := range v {
					newMap[k1] = v1
				}
				return newMap
			}),
	}
	return newIdx
}

var _ clustersnapshot.FastPredicateLister = &fastPredicateIndex{}
