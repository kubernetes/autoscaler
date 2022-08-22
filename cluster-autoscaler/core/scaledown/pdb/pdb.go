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
	"fmt"

	apiv1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/autoscaler/cluster-autoscaler/utils/drain"
	"k8s.io/klog/v2"
)

// PdbRemainingDisruptions stores how many discuptiption is left for pdb.
type PdbRemainingDisruptions struct {
	pdbs []*policyv1.PodDisruptionBudget
}

// NewPdbRemainingDisruptions initialize PdbRemainingDisruptions.
func NewPdbRemainingDisruptions(pdbs []*policyv1.PodDisruptionBudget) *PdbRemainingDisruptions {
	pdbsCopy := make([]*policyv1.PodDisruptionBudget, len(pdbs))
	for i, pdb := range pdbs {
		pdbsCopy[i] = pdb.DeepCopy()
	}
	return &PdbRemainingDisruptions{pdbsCopy}
}

// CanDisrupt return if the pod can be removed.
func (p *PdbRemainingDisruptions) CanDisrupt(pods []*apiv1.Pod) (bool, *drain.BlockingPod) {
	for _, pdb := range p.pdbs {
		selector, err := metav1.LabelSelectorAsSelector(pdb.Spec.Selector)
		if err != nil {
			klog.Errorf("Can't get selector for pdb %s", pdb.GetNamespace()+" "+pdb.GetName())
			return false, nil
		}
		count := int32(0)
		for _, pod := range pods {
			if pod.Namespace == pdb.Namespace && selector.Matches(labels.Set(pod.Labels)) {
				count += 1
				if pdb.Status.DisruptionsAllowed < count {
					return false, &drain.BlockingPod{Pod: pod, Reason: drain.NotEnoughPdb}
				}
			}
		}
	}
	return true, nil
}

// Update make updates the remaining disruptions for pdb.
func (p *PdbRemainingDisruptions) Update(pods []*apiv1.Pod) error {
	for _, pdb := range p.pdbs {
		selector, err := metav1.LabelSelectorAsSelector(pdb.Spec.Selector)
		if err != nil {
			return err
		}
		for _, pod := range pods {
			if pod.Namespace == pdb.Namespace && selector.Matches(labels.Set(pod.Labels)) {
				if pdb.Status.DisruptionsAllowed < 1 {
					return fmt.Errorf("Pod can't be removed, pdb is blocking by pdb %s, disruptionsAllowed: %v", pdb.GetNamespace()+"/"+pdb.GetName(), pdb.Status.DisruptionsAllowed)
				}
				pdb.Status.DisruptionsAllowed -= 1
			}
		}
	}
	return nil
}
