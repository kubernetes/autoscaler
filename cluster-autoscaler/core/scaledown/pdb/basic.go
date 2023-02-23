/*
Copyright 2023 The Kubernetes Authors.

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

type pdbInfo struct {
	pdb      *policyv1.PodDisruptionBudget
	selector labels.Selector
}

// basicRemainingPdbTracker is the basic implementation of RemainingPdbTracker
type basicRemainingPdbTracker struct {
	pdbInfos []*pdbInfo
}

// NewBasicRemainingPdbTracker returns a new instance of basicRemainingPdbTracker
func NewBasicRemainingPdbTracker() *basicRemainingPdbTracker {
	return &basicRemainingPdbTracker{}
}

func (t *basicRemainingPdbTracker) SetPdbs(pdbs []*policyv1.PodDisruptionBudget) error {
	t.Clear()
	for _, pdb := range pdbs {
		pdbCopy := pdb.DeepCopy()
		selector, err := metav1.LabelSelectorAsSelector(pdbCopy.Spec.Selector)
		if err != nil {
			return err
		}
		t.pdbInfos = append(t.pdbInfos, &pdbInfo{
			pdb:      pdbCopy,
			selector: selector,
		})
	}
	return nil
}

func (t *basicRemainingPdbTracker) GetPdbs() []*policyv1.PodDisruptionBudget {
	var pdbs []*policyv1.PodDisruptionBudget
	for _, pdbInfo := range t.pdbInfos {
		pdbs = append(pdbs, pdbInfo.pdb)
	}
	return pdbs
}

func (t *basicRemainingPdbTracker) CanRemovePods(pods []*apiv1.Pod) (canRemove, inParallel bool, blockingPod *drain.BlockingPod) {
	inParallel = true
	for _, pdbInfo := range t.pdbInfos {
		count := int32(0)
		for _, pod := range pods {
			if pod.Namespace == pdbInfo.pdb.Namespace && pdbInfo.selector.Matches(labels.Set(pod.Labels)) {
				count += 1
				if pdbInfo.pdb.Status.DisruptionsAllowed < 1 {
					return false, false, &drain.BlockingPod{Pod: pod, Reason: drain.NotEnoughPdb}
				}
				if pdbInfo.pdb.Status.DisruptionsAllowed < count {
					inParallel = false
					blockingPod = &drain.BlockingPod{Pod: pod, Reason: drain.NotEnoughPdb}
				}
			}
		}
	}
	return true, inParallel, blockingPod
}

func (t *basicRemainingPdbTracker) RemovePods(pods []*apiv1.Pod) {
	for _, pdbInfo := range t.pdbInfos {
		for _, pod := range pods {
			if pod.Namespace == pdbInfo.pdb.Namespace && pdbInfo.selector.Matches(labels.Set(pod.Labels)) {
				pdbInfo.pdb.Status.DisruptionsAllowed -= 1
			}
		}
	}
}

func (t *basicRemainingPdbTracker) Clear() {
	t.pdbInfos = nil
}
