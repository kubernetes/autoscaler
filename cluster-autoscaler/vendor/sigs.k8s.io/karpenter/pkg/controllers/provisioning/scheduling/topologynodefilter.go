/*
Copyright The Kubernetes Authors.

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

package scheduling

import (
	"github.com/awslabs/operatorpkg/option"
	corev1 "k8s.io/api/core/v1"

	"sigs.k8s.io/karpenter/pkg/scheduling"
)

// TopologyNodeFilter is used to determine if a given actual node or scheduling node matches the pod's node selectors
// and required node affinity terms.  This is used with topology spread constraints to determine if the node should be
// included for topology counting purposes. This is only used with topology spread constraints as affinities/anti-affinities
// always count across all nodes. A nil or zero-value TopologyNodeFilter behaves well and the filter returns true for
// all nodes.
type TopologyNodeFilter struct {
	Requirements   []scheduling.Requirements
	TaintPolicy    corev1.NodeInclusionPolicy
	AffinityPolicy corev1.NodeInclusionPolicy
	Tolerations    []corev1.Toleration
}

func MakeTopologyNodeFilter(p *corev1.Pod, taintPolicy corev1.NodeInclusionPolicy, affinityPolicy corev1.NodeInclusionPolicy) TopologyNodeFilter {
	nodeSelectorRequirements := scheduling.NewLabelRequirements(p.Spec.NodeSelector)
	// if we only have a label selector, that's the only requirement that must match
	if p.Spec.Affinity == nil || p.Spec.Affinity.NodeAffinity == nil || p.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution == nil {
		return TopologyNodeFilter{
			Requirements:   []scheduling.Requirements{nodeSelectorRequirements},
			TaintPolicy:    taintPolicy,
			AffinityPolicy: affinityPolicy,
			Tolerations:    p.Spec.Tolerations,
		}
	}

	// otherwise, we need to match the combination of label selector and any term of the required node affinities since
	// those terms are OR'd together
	filter := TopologyNodeFilter{
		TaintPolicy:    taintPolicy,
		AffinityPolicy: affinityPolicy,
		Tolerations:    p.Spec.Tolerations,
	}
	for _, term := range p.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms {
		requirements := scheduling.NewRequirements()
		requirements.Add(nodeSelectorRequirements.Values()...)
		requirements.Add(scheduling.NewNodeSelectorRequirements(term.MatchExpressions...).Values()...)
		filter.Requirements = append(filter.Requirements, requirements)
	}

	return filter
}

// Matches returns true if the TopologyNodeFilter doesn't prohibit node from the participating in the topology
func (t TopologyNodeFilter) Matches(taints []corev1.Taint, requirements scheduling.Requirements, compatibilityOptions ...option.Function[scheduling.CompatibilityOptions]) bool {
	matchesAffinity := true
	if t.AffinityPolicy == corev1.NodeInclusionPolicyHonor {
		matchesAffinity = t.matchesRequirements(requirements)
	}
	matchesTaints := true
	if t.TaintPolicy == corev1.NodeInclusionPolicyHonor {
		if err := scheduling.Taints(taints).Tolerates(t.Tolerations); err != nil {
			matchesTaints = false
		}
	}
	return matchesAffinity && matchesTaints
}

// MatchesRequirements returns true if the TopologyNodeFilter doesn't prohibit a node with the requirements from
// participating in the topology. This method allows checking the requirements from a scheduling.NodeClaim to see if the
// node we will soon create participates in this topology.
func (t TopologyNodeFilter) matchesRequirements(requirements scheduling.Requirements, compatibilityOptions ...option.Function[scheduling.CompatibilityOptions]) bool {
	// no requirements, so it always matches
	if len(t.Requirements) == 0 || t.AffinityPolicy == corev1.NodeInclusionPolicyIgnore {
		return true
	}
	// these are an OR, so if any passes the filter passes
	for _, req := range t.Requirements {
		if err := requirements.Compatible(req, compatibilityOptions...); err == nil {
			return true
		}
	}
	return false
}
