/*
Copyright 2022 The Kubernetes Authors.

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

package pdb

import (
	apiv1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/autoscaler/cluster-autoscaler/utils/drain"
)

// PdbRemainingDisruptions stores how many discuptiption is left for pdb.
type PdbRemainingDisruptions struct {
	pdbs      []*policyv1.PodDisruptionBudget
	selectors map[*policyv1.PodDisruptionBudget]labels.Selector
}

// NewPdbRemainingDisruptions initialize PdbRemainingDisruptions.
func NewPdbRemainingDisruptions(pdbs []*policyv1.PodDisruptionBudget) (*PdbRemainingDisruptions, error) {
	pdbsCopy := make([]*policyv1.PodDisruptionBudget, len(pdbs))
	selectors := make(map[*policyv1.PodDisruptionBudget]labels.Selector)
	for i, pdb := range pdbs {
		pdbsCopy[i] = pdb.DeepCopy()
		selector, err := metav1.LabelSelectorAsSelector(pdb.Spec.Selector)
		if err != nil {
			return nil, err
		}
		selectors[pdbsCopy[i]] = selector
	}
	return &PdbRemainingDisruptions{pdbsCopy, selectors}, nil
}

// CanDisrupt return if the set of pods can be removed.
// inParallel indicates that the pods could not be removed in parallel.
// If inParallel == false, evicting this set of pods from node could fail due to drain timeout.
func (p *PdbRemainingDisruptions) CanDisrupt(pods []*apiv1.Pod) (canRemove, inParallel bool, blockingPod *drain.BlockingPod) {
	inParallel = true
	for _, pdb := range p.pdbs {
		selector := p.selectors[pdb]
		count := int32(0)
		for _, pod := range pods {
			if pod.Namespace == pdb.Namespace && selector.Matches(labels.Set(pod.Labels)) {
				count += 1
				if pdb.Status.DisruptionsAllowed < 1 {
					return false, false, &drain.BlockingPod{Pod: pod, Reason: drain.NotEnoughPdb}
				}
				if pdb.Status.DisruptionsAllowed < count {
					inParallel = false
					blockingPod = &drain.BlockingPod{Pod: pod, Reason: drain.NotEnoughPdb}
				}
			}
		}
	}
	return true, inParallel, blockingPod
}

// Update make updates the remaining disruptions for pdb.
func (p *PdbRemainingDisruptions) Update(pods []*apiv1.Pod) {
	for _, pdb := range p.pdbs {
		selector := p.selectors[pdb]
		for _, pod := range pods {
			if pod.Namespace == pdb.Namespace && selector.Matches(labels.Set(pod.Labels)) {
				pdb.Status.DisruptionsAllowed -= 1
			}
		}
	}
}

// GetPdbs return pdb list.
func (p *PdbRemainingDisruptions) GetPdbs() []*policyv1.PodDisruptionBudget {
	return p.pdbs
}
