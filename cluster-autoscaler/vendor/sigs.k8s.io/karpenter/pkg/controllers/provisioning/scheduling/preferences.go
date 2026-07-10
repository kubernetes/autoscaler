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
	"context"
	"fmt"
	"sort"

	"github.com/samber/lo"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"sigs.k8s.io/karpenter/pkg/utils/pretty"
)

type Preferences struct {
	// ToleratePreferNoSchedule controls if preference relaxation adds a toleration for PreferNoSchedule taints.  This only
	// helps if there is a corresponding taint, so we don't always add it.
	ToleratePreferNoSchedule bool
}

func (p *Preferences) Relax(ctx context.Context, pod *v1.Pod) bool {
	relaxations := []func(*v1.Pod) *string{
		p.removeRequiredNodeAffinityTerm,
		p.removePreferredPodAffinityTerm,
		p.removePreferredPodAntiAffinityTerm,
		p.removePreferredNodeAffinityTerm,
		p.removeTopologySpreadScheduleAnyway}

	if p.ToleratePreferNoSchedule {
		relaxations = append(relaxations, p.toleratePreferNoScheduleTaints)
	}

	for _, relaxFunc := range relaxations {
		if reason := relaxFunc(pod); reason != nil {
			log.FromContext(ctx).WithValues("Pod", klog.KObj(pod)).V(1).Info("relaxing soft constraints for pod since it previously failed to schedule", "reason", lo.FromPtr(reason))
			return true
		}
	}
	return false
}

func (p *Preferences) removePreferredNodeAffinityTerm(pod *v1.Pod) *string {
	if pod.Spec.Affinity == nil || pod.Spec.Affinity.NodeAffinity == nil || len(pod.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution) == 0 {
		return nil
	}
	terms := pod.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution
	// Remove the terms if there are any (terms are an OR semantic)
	if len(terms) > 0 {
		// Sort descending by weight to remove heaviest preferences to try lighter ones
		sort.SliceStable(terms, func(i, j int) bool { return terms[i].Weight > terms[j].Weight })
		pod.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution = terms[1:]
		return new(fmt.Sprintf("removing: spec.affinity.nodeAffinity.preferredDuringSchedulingIgnoredDuringExecution[0]=%s", pretty.Concise(terms[0])))
	}
	return nil
}

func (p *Preferences) removeRequiredNodeAffinityTerm(pod *v1.Pod) *string {
	if pod.Spec.Affinity == nil ||
		pod.Spec.Affinity.NodeAffinity == nil ||
		pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution == nil ||
		len(pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms) == 0 {
		return nil
	}
	terms := pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
	// Remove the first term if there's more than one (terms are an OR semantic), Unlike preferred affinity, we cannot remove all terms
	if len(terms) > 1 {
		pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms = terms[1:]
		return new(fmt.Sprintf("removing: spec.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution[0]=%s", pretty.Concise(terms[0])))
	}
	return nil
}

func (p *Preferences) removeTopologySpreadScheduleAnyway(pod *v1.Pod) *string {
	for i, tsc := range pod.Spec.TopologySpreadConstraints {
		if tsc.WhenUnsatisfiable == v1.ScheduleAnyway {
			msg := fmt.Sprintf("removing: spec.topologySpreadConstraints = %s", pretty.Concise(tsc))
			pod.Spec.TopologySpreadConstraints[i] = pod.Spec.TopologySpreadConstraints[len(pod.Spec.TopologySpreadConstraints)-1]
			pod.Spec.TopologySpreadConstraints = pod.Spec.TopologySpreadConstraints[:len(pod.Spec.TopologySpreadConstraints)-1]
			return new(msg)
		}
	}
	return nil
}

func (p *Preferences) removePreferredPodAffinityTerm(pod *v1.Pod) *string {
	if pod.Spec.Affinity == nil || pod.Spec.Affinity.PodAffinity == nil || len(pod.Spec.Affinity.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution) == 0 {
		return nil
	}
	terms := pod.Spec.Affinity.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution
	// Remove the all the terms
	if len(terms) > 0 {
		// Sort descending by weight to remove heaviest preferences to try lighter ones
		sort.SliceStable(terms, func(i, j int) bool { return terms[i].Weight > terms[j].Weight })
		pod.Spec.Affinity.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution = terms[1:]
		return new(fmt.Sprintf("removing: spec.affinity.podAffinity.preferredDuringSchedulingIgnoredDuringExecution[0]=%s", pretty.Concise(terms[0])))
	}
	return nil
}

func (p *Preferences) removePreferredPodAntiAffinityTerm(pod *v1.Pod) *string {
	if pod.Spec.Affinity == nil || pod.Spec.Affinity.PodAntiAffinity == nil || len(pod.Spec.Affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution) == 0 {
		return nil
	}
	terms := pod.Spec.Affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution
	// Remove the all the terms
	if len(terms) > 0 {
		// Sort descending by weight to remove heaviest preferences to try lighter ones
		sort.SliceStable(terms, func(i, j int) bool { return terms[i].Weight > terms[j].Weight })
		pod.Spec.Affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution = terms[1:]
		return new(fmt.Sprintf("removing: spec.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution[0]=%s", pretty.Concise(terms[0])))
	}
	return nil
}

func (p *Preferences) toleratePreferNoScheduleTaints(pod *v1.Pod) *string {
	// Tolerate all Taints with PreferNoSchedule effect
	toleration := v1.Toleration{
		Operator: v1.TolerationOpExists,
		Effect:   v1.TaintEffectPreferNoSchedule,
	}
	for _, t := range pod.Spec.Tolerations {
		if t.MatchToleration(&toleration) {
			return nil
		}
	}
	tolerations := append(pod.Spec.Tolerations, toleration)
	pod.Spec.Tolerations = tolerations
	return new("adding: toleration for PreferNoSchedule taints")
}
