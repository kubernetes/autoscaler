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

package predicate

import (
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/scheduler/framework/plugins/interpodaffinity"
	"k8s.io/kubernetes/pkg/scheduler/framework/plugins/podtopologyspread"
)

type FastPredicateState struct {
	podAffinitySelectors     []labels.Selector
	podAntiAffinitySelectors []labels.Selector
	topologySpreadSelectors  []labels.Selector
}

func (p *SchedulerPluginRunner) computeFastPredicateState(pod *apiv1.Pod) (*FastPredicateState, error) {
	state := &FastPredicateState{}

	affinity := pod.Spec.Affinity
	if affinity != nil {
		if affinity.PodAffinity != nil {
			for _, term := range affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution {
				sel, err := metav1.LabelSelectorAsSelector(term.LabelSelector)
				if err != nil {
					return nil, err
				}
				state.podAffinitySelectors = append(state.podAffinitySelectors, sel)
			}
		}
		if affinity.PodAntiAffinity != nil {
			for _, term := range affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution {
				sel, err := metav1.LabelSelectorAsSelector(term.LabelSelector)
				if err != nil {
					return nil, err
				}
				state.podAntiAffinitySelectors = append(state.podAntiAffinitySelectors, sel)
			}
		}
	}

	for _, constraint := range pod.Spec.TopologySpreadConstraints {
		if constraint.WhenUnsatisfiable == apiv1.DoNotSchedule {
			sel, err := metav1.LabelSelectorAsSelector(constraint.LabelSelector)
			if err != nil {
				return nil, err
			}
			state.topologySpreadSelectors = append(state.topologySpreadSelectors, sel)
		}
	}

	return state, nil
}

func (p *SchedulerPluginRunner) fastCheckPredicates(pod *apiv1.Pod, nodeInfo *framework.NodeInfo, state *FastPredicateState) clustersnapshot.SchedulingError {
	lister := p.snapshot.FastPredicateLister()
	if lister == nil {
		klog.V(4).Infof("FastPredicateLister is nil, skipping fast predicates")
		return nil
	}

	if err := p.fastCheckPodAffinity(pod, nodeInfo, lister, state); err != nil {
		klog.V(4).Infof("fastCheckPodAffinity failed: %v", err)
		return err
	}

	if err := p.fastCheckPodTopologySpread(pod, nodeInfo, lister, state); err != nil {
		klog.V(4).Infof("fastCheckPodTopologySpread failed: %v", err)
		return err
	}

	return nil
}

func (p *SchedulerPluginRunner) fastCheckPodAffinity(pod *apiv1.Pod, nodeInfo *framework.NodeInfo, lister clustersnapshot.FastPredicateLister, state *FastPredicateState) clustersnapshot.SchedulingError {
	affinity := pod.Spec.Affinity
	if affinity == nil || (affinity.PodAffinity == nil && affinity.PodAntiAffinity == nil) {
		return nil
	}

	node := nodeInfo.Node()

	if affinity.PodAffinity != nil {
		for i, term := range affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution {
			topoValue, ok := node.Labels[term.TopologyKey]
			if !ok {
				continue
			}
			var selector labels.Selector
			var err error
			if state != nil && i < len(state.podAffinitySelectors) {
				selector = state.podAffinitySelectors[i]
			} else {
				selector, err = metav1.LabelSelectorAsSelector(term.LabelSelector)
				if err != nil {
					return clustersnapshot.NewSchedulingInternalError(pod, err.Error())
				}
			}
			count := lister.PodAffinityCount(term.TopologyKey, topoValue, selector)
			if count == 0 {
				return clustersnapshot.NewFailingPredicateError(pod, "InterPodAffinity", []string{interpodaffinity.ErrReasonAffinityRulesNotMatch}, "", "")
			}
		}
	}

	if affinity.PodAntiAffinity != nil {
		for i, term := range affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution {
			topoValue, ok := node.Labels[term.TopologyKey]
			if !ok {
				continue
			}
			var selector labels.Selector
			var err error
			if state != nil && i < len(state.podAntiAffinitySelectors) {
				selector = state.podAntiAffinitySelectors[i]
			} else {
				selector, err = metav1.LabelSelectorAsSelector(term.LabelSelector)
				if err != nil {
					return clustersnapshot.NewSchedulingInternalError(pod, err.Error())
				}
			}
			count := lister.PodAffinityCount(term.TopologyKey, topoValue, selector)
			if count > 0 {
				return clustersnapshot.NewFailingPredicateError(pod, "InterPodAffinity", []string{interpodaffinity.ErrReasonAntiAffinityRulesNotMatch}, "", "")
			}
		}
	}

	return nil
}

func (p *SchedulerPluginRunner) fastCheckPodTopologySpread(pod *apiv1.Pod, nodeInfo *framework.NodeInfo, lister clustersnapshot.FastPredicateLister, state *FastPredicateState) clustersnapshot.SchedulingError {
	if len(pod.Spec.TopologySpreadConstraints) == 0 {
		return nil
	}

	node := nodeInfo.Node()

	spreadIdx := 0
	for _, constraint := range pod.Spec.TopologySpreadConstraints {
		if constraint.WhenUnsatisfiable != apiv1.DoNotSchedule {
			continue
		}

		topoValue, ok := node.Labels[constraint.TopologyKey]
		if !ok {
			// If node doesn't have the topology key, it's usually skipped or fails depending on configuration.
			spreadIdx++
			continue
		}

		var selector labels.Selector
		var err error
		if state != nil && spreadIdx < len(state.topologySpreadSelectors) {
			selector = state.topologySpreadSelectors[spreadIdx]
		} else {
			selector, err = metav1.LabelSelectorAsSelector(constraint.LabelSelector)
			if err != nil {
				return clustersnapshot.NewSchedulingInternalError(pod, err.Error())
			}
		}
		spreadIdx++

		count := lister.PodAffinityCount(constraint.TopologyKey, topoValue, selector)

		// To implement maxSkew, we need minCount across all domains.
		minCount := p.getMinTopologyCount(constraint.TopologyKey, selector, lister)

		if int32(count+1-minCount) > constraint.MaxSkew {
			return clustersnapshot.NewFailingPredicateError(pod, "PodTopologySpread", []string{podtopologyspread.ErrReasonConstraintsNotMatch}, "", "")
		}
	}

	return nil
}

func (p *SchedulerPluginRunner) getMinTopologyCount(topoKey string, selector labels.Selector, lister clustersnapshot.FastPredicateLister) int {
	domains := lister.TopologyDomains(topoKey)
	if len(domains) == 0 {
		return 0
	}

	minCount := -1
	for _, domain := range domains {
		count := lister.PodAffinityCount(topoKey, domain, selector)
		if minCount == -1 || count < minCount {
			minCount = count
		}
		if minCount == 0 {
			break
		}
	}

	if minCount == -1 {
		return 0
	}
	return minCount
}
