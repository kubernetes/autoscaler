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
	"testing"

	apiv1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/pdb"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/drainability"
	"k8s.io/autoscaler/cluster-autoscaler/utils/drain"

	"github.com/stretchr/testify/assert"
)

func TestDrainable(t *testing.T) {
	one := intstr.FromInt(1)

	for desc, tc := range map[string]struct {
		pod         *apiv1.Pod
		pdbs        []*policyv1.PodDisruptionBudget
		wantOutcome drainability.OutcomeType
		wantReason  drain.BlockingPodReason
	}{
		"no pdbs": {
			pod: &apiv1.Pod{},
		},
		"no matching pdbs": {
			pod: &apiv1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "happy",
					Namespace: "good",
					Labels: map[string]string{
						"label": "true",
					},
				},
			},
			pdbs: []*policyv1.PodDisruptionBudget{
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "bad",
					},
					Spec: policyv1.PodDisruptionBudgetSpec{
						MinAvailable: &one,
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"label": "true",
							},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "good",
					},
					Spec: policyv1.PodDisruptionBudgetSpec{
						MinAvailable: &one,
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"label": "false",
							},
						},
					},
				},
			},
		},
		"pdb prevents scale-down": {
			pod: &apiv1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "sad",
					Namespace: "good",
					Labels: map[string]string{
						"label": "true",
					},
				},
			},
			pdbs: []*policyv1.PodDisruptionBudget{
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "bad",
					},
					Spec: policyv1.PodDisruptionBudgetSpec{
						MinAvailable: &one,
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"label": "true",
							},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "good",
					},
					Spec: policyv1.PodDisruptionBudgetSpec{
						MinAvailable: &one,
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"label": "true",
							},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "good",
					},
					Spec: policyv1.PodDisruptionBudgetSpec{
						MinAvailable: &one,
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"label": "false",
							},
						},
					},
				},
			},
			wantOutcome: drainability.BlockDrain,
			wantReason:  drain.NotEnoughPdb,
		},
	} {
		t.Run(desc, func(t *testing.T) {
			tracker := pdb.NewBasicRemainingPdbTracker()
			tracker.SetPdbs(tc.pdbs)
			drainCtx := &drainability.DrainContext{
				RemainingPdbTracker: tracker,
			}

			got := New().Drainable(drainCtx, tc.pod, nil)
			assert.Equal(t, tc.wantReason, got.BlockingReason)
			assert.Equal(t, tc.wantOutcome, got.Outcome)
		})
	}
}
