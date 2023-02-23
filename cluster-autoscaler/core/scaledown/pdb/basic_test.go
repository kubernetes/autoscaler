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
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
)

var (
	one    = intstr.FromInt(1)
	label1 = "label-1"
	label2 = "label-2"
	pdb1   = &policyv1.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "ns",
		},
		Spec: policyv1.PodDisruptionBudgetSpec{
			MinAvailable: &one,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					label1: "true",
				},
			},
		},
		Status: policyv1.PodDisruptionBudgetStatus{
			DisruptionsAllowed: 1,
		},
	}
	pdb2 = &policyv1.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "bar",
			Namespace: "ns",
		},
		Spec: policyv1.PodDisruptionBudgetSpec{
			MinAvailable: &one,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					label2: "true",
				},
			},
		},
		Status: policyv1.PodDisruptionBudgetStatus{
			DisruptionsAllowed: 2,
		},
	}
	pdb1Copy = &policyv1.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "ns",
		},
		Spec: policyv1.PodDisruptionBudgetSpec{
			MinAvailable: &one,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					label1: "true",
				},
			},
		},
		Status: policyv1.PodDisruptionBudgetStatus{
			DisruptionsAllowed: 1,
		},
	}
	pdb2Copy = &policyv1.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "bar",
			Namespace: "ns",
		},
		Spec: policyv1.PodDisruptionBudgetSpec{
			MinAvailable: &one,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					label2: "true",
				},
			},
		},
		Status: policyv1.PodDisruptionBudgetStatus{
			DisruptionsAllowed: 2,
		},
	}
)

func TestBasicCanRemovePods(t *testing.T) {
	testCases := []struct {
		name            string
		podsLabel1      int
		podsLabel2      int
		podsBothLabels  int
		pdbs            []*policyv1.PodDisruptionBudget
		pdbsDisruptions [2]int32
		canDisrupt      bool
		inParallel      bool
	}{
		{
			name:       "No pdbs",
			podsLabel1: 2,
			podsLabel2: 1,
			canDisrupt: true,
			inParallel: true,
		},
		{
			name:            "Not enough pod disruption budgets",
			podsLabel1:      2,
			podsLabel2:      1,
			pdbs:            []*policyv1.PodDisruptionBudget{pdb1, pdb2},
			pdbsDisruptions: [2]int32{1, 0},
			canDisrupt:      false,
			inParallel:      false,
		},
		{
			name:            "Pod disruption budgets is at risk",
			podsLabel1:      2,
			podsLabel2:      1,
			pdbs:            []*policyv1.PodDisruptionBudget{pdb1, pdb2},
			pdbsDisruptions: [2]int32{1, 2},
			canDisrupt:      true,
		},
		{
			name:            "Enough pod disruption budgets",
			podsLabel1:      2,
			podsLabel2:      3,
			pdbs:            []*policyv1.PodDisruptionBudget{pdb1, pdb2},
			pdbsDisruptions: [2]int32{2, 4},
			canDisrupt:      true,
			inParallel:      true,
		},
		{
			name:            "Pod covered with both PDBs can be moved",
			podsLabel1:      1,
			podsLabel2:      1,
			podsBothLabels:  1,
			pdbs:            []*policyv1.PodDisruptionBudget{pdb1, pdb2},
			pdbsDisruptions: [2]int32{1, 1},
			canDisrupt:      true,
			inParallel:      true,
		},
		{
			name:            "Pod covered with both PDBs, is risky",
			podsLabel1:      2,
			podsLabel2:      2,
			podsBothLabels:  1,
			pdbs:            []*policyv1.PodDisruptionBudget{pdb1, pdb2},
			pdbsDisruptions: [2]int32{2, 1},
			canDisrupt:      true,
			inParallel:      false,
		},
	}
	for _, test := range testCases {
		pdb1.Status.DisruptionsAllowed = test.pdbsDisruptions[0]
		pdb2.Status.DisruptionsAllowed = test.pdbsDisruptions[1]
		tracker := NewBasicRemainingPdbTracker()
		assert.NoError(t, tracker.SetPdbs(test.pdbs))
		pods := makePodsWithLabel(label1, test.podsLabel1)
		pods2 := makePodsWithLabel(label2, test.podsLabel2-test.podsBothLabels)
		if test.podsBothLabels > 0 {
			addLabelToPods(pods[:test.podsBothLabels], label2)
		}
		pods = append(pods, pods2...)
		gotDisrupt, inParallel, _ := tracker.CanRemovePods(pods)
		if gotDisrupt != test.canDisrupt || inParallel != test.inParallel {
			t.Errorf("%s: CanDisrupt() return %v, %v, want %v, %v", test.name, gotDisrupt, inParallel, test.canDisrupt, test.inParallel)
		}
	}
}

func TestBasicRemovePods(t *testing.T) {
	testCases := []struct {
		name                   string
		podsLabel1             int
		podsLabel2             int
		podsBothLabels         int
		pdbs                   []*policyv1.PodDisruptionBudget
		updatedPdbs            []*policyv1.PodDisruptionBudget
		pdbsDisruptions        [2]int32
		updatedPdbsDisruptions [2]int32
	}{
		{
			name:                   "Pod covered with both PDBs",
			podsLabel1:             1,
			podsLabel2:             1,
			podsBothLabels:         1,
			pdbs:                   []*policyv1.PodDisruptionBudget{pdb1, pdb2},
			updatedPdbs:            []*policyv1.PodDisruptionBudget{pdb1Copy, pdb2Copy},
			pdbsDisruptions:        [2]int32{1, 1},
			updatedPdbsDisruptions: [2]int32{0, 0},
		},
		{
			name:           "No PDBs",
			pdbs:           []*policyv1.PodDisruptionBudget{},
			updatedPdbs:    []*policyv1.PodDisruptionBudget{},
			podsLabel1:     2,
			podsLabel2:     3,
			podsBothLabels: 1,
		},
	}
	for _, test := range testCases {
		pdb1.Status.DisruptionsAllowed = test.pdbsDisruptions[0]
		pdb2.Status.DisruptionsAllowed = test.pdbsDisruptions[1]
		tracker := NewBasicRemainingPdbTracker()
		assert.NoError(t, tracker.SetPdbs(test.pdbs))
		pods := makePodsWithLabel(label1, test.podsLabel1)
		pods2 := makePodsWithLabel(label2, test.podsLabel2-test.podsBothLabels)
		if test.podsBothLabels > 0 {
			addLabelToPods(pods[:test.podsBothLabels], label2)
		}
		pods = append(pods, pods2...)

		pdb1Copy.Status.DisruptionsAllowed = test.updatedPdbsDisruptions[0]
		pdb2Copy.Status.DisruptionsAllowed = test.updatedPdbsDisruptions[1]
		wantTracker := NewBasicRemainingPdbTracker()
		assert.NoError(t, wantTracker.SetPdbs(test.updatedPdbs))
		tracker.RemovePods(pods)
		if diff := cmp.Diff(wantTracker.GetPdbs(), tracker.GetPdbs()); diff != "" {
			t.Errorf("Update() diff (-want +got):\n%s", diff)
		}
	}
}

func makePodsWithLabel(label string, amount int) []*apiv1.Pod {
	pods := []*apiv1.Pod{}
	for i := 0; i < amount; i++ {
		pod := &apiv1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:            fmt.Sprintf("pod-1-%d", i),
				Namespace:       "ns",
				OwnerReferences: GenerateOwnerReferences("rs", "ReplicaSet", "extensions/v1beta1", ""),
				Labels: map[string]string{
					label: "true",
				},
			},
			Spec: apiv1.PodSpec{},
		}
		pods = append(pods, pod)
	}
	return pods
}

func addLabelToPods(pods []*apiv1.Pod, label string) {
	for _, pod := range pods {
		pod.ObjectMeta.Labels[label] = "true"
	}
}
