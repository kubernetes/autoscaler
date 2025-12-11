/*
Copyright 2025 The Kubernetes Authors.

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

package capacityquota

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	cqv1alpha1 "k8s.io/autoscaler/cluster-autoscaler/apis/capacityquota/autoscaling.x-k8s.io/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestProvider_Quotas(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = cqv1alpha1.AddToScheme(scheme)

	testCases := []struct {
		name        string
		existingCQs []client.Object
		wantQuotas  []string
	}{
		{
			name:        "no-capacity-quotas",
			existingCQs: []client.Object{},
			wantQuotas:  nil,
		},
		{
			name: "single-capacity-quota",
			existingCQs: []client.Object{
				&cqv1alpha1.CapacityQuota{
					ObjectMeta: metav1.ObjectMeta{Name: "cq1"},
					Spec: cqv1alpha1.CapacityQuotaSpec{
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{"foo": "bar"}},
						Limits: cqv1alpha1.CapacityQuotaLimits{
							Resources: cqv1alpha1.ResourceList{
								cqv1alpha1.ResourceCPU:    resource.MustParse("10"),
								cqv1alpha1.ResourceMemory: resource.MustParse("20Gi"),
							},
						},
					},
				},
			},
			wantQuotas: []string{"CapacityQuota/cq1"},
		},
		{
			name: "multiple-capacity-quotas",
			existingCQs: []client.Object{
				&cqv1alpha1.CapacityQuota{
					ObjectMeta: metav1.ObjectMeta{Name: "cq1"},
					Spec: cqv1alpha1.CapacityQuotaSpec{
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{"foo": "bar"},
						},
						Limits: cqv1alpha1.CapacityQuotaLimits{
							Resources: cqv1alpha1.ResourceList{
								cqv1alpha1.ResourceCPU: resource.MustParse("10"),
							},
						},
					},
				},
				&cqv1alpha1.CapacityQuota{
					ObjectMeta: metav1.ObjectMeta{Name: "cq2"},
					Spec: cqv1alpha1.CapacityQuotaSpec{
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{"baz": "qux"},
						},
						Limits: cqv1alpha1.CapacityQuotaLimits{
							Resources: cqv1alpha1.ResourceList{
								cqv1alpha1.ResourceMemory: resource.MustParse("5Gi"),
							},
						},
					},
				},
			},
			wantQuotas: []string{"CapacityQuota/cq1", "CapacityQuota/cq2"},
		},
		{
			name: "capacity-quota-with-invalid-selector-is-skipped",
			existingCQs: []client.Object{
				&cqv1alpha1.CapacityQuota{
					ObjectMeta: metav1.ObjectMeta{Name: "cq_invalid"},
					Spec: cqv1alpha1.CapacityQuotaSpec{
						Selector: &metav1.LabelSelector{
							MatchExpressions: []metav1.LabelSelectorRequirement{
								{
									Key:      "invalidKey!!!!",
									Operator: metav1.LabelSelectorOpExists,
								},
							},
						},
						Limits: cqv1alpha1.CapacityQuotaLimits{
							Resources: cqv1alpha1.ResourceList{
								cqv1alpha1.ResourceCPU: resource.MustParse("10"),
							},
						},
					},
				},
				&cqv1alpha1.CapacityQuota{
					ObjectMeta: metav1.ObjectMeta{Name: "cq_valid"},
					Spec: cqv1alpha1.CapacityQuotaSpec{
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{"valid": "true"},
						},
						Limits: cqv1alpha1.CapacityQuotaLimits{
							Resources: cqv1alpha1.ResourceList{
								cqv1alpha1.ResourceMemory: resource.MustParse("1Gi"),
							},
						},
					},
				},
			},
			wantQuotas: []string{"CapacityQuota/cq_valid"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(tc.existingCQs...).Build()
			p := &Provider{kubeClient: fakeClient}

			gotQuotas, err := p.Quotas()
			if err != nil {
				t.Fatalf("Provider.Quotas() unexpected error: %v", err)
			}

			var gotIDs []string
			for _, q := range gotQuotas {
				gotIDs = append(gotIDs, q.ID())
			}

			if diff := cmp.Diff(tc.wantQuotas, gotIDs, cmpopts.SortSlices(func(a, b string) bool { return a < b })); diff != "" {
				t.Errorf("Provider.Quotas() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestCapacityQuota_Selector(t *testing.T) {
	testCases := []struct {
		name          string
		cq            cqv1alpha1.CapacityQuota
		node          *corev1.Node
		wantAppliesTo bool
		wantErrMsg    string
	}{
		{
			name: "matchLabels-matches",
			cq: cqv1alpha1.CapacityQuota{
				ObjectMeta: metav1.ObjectMeta{Name: "cq1"},
				Spec: cqv1alpha1.CapacityQuotaSpec{
					Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"foo": "bar"}},
					Limits:   cqv1alpha1.CapacityQuotaLimits{Resources: cqv1alpha1.ResourceList{cqv1alpha1.ResourceCPU: resource.MustParse("1")}},
				},
			},
			node:          &corev1.Node{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"foo": "bar"}}},
			wantAppliesTo: true,
		},
		{
			name: "matchLabels-does-not-match",
			cq: cqv1alpha1.CapacityQuota{
				ObjectMeta: metav1.ObjectMeta{Name: "cq1"},
				Spec: cqv1alpha1.CapacityQuotaSpec{
					Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"foo": "bar"}},
					Limits:   cqv1alpha1.CapacityQuotaLimits{Resources: cqv1alpha1.ResourceList{cqv1alpha1.ResourceCPU: resource.MustParse("1")}},
				},
			},
			node:          &corev1.Node{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"foo": "baz"}}},
			wantAppliesTo: false,
		},
		{
			name: "nil-selector-matches-everything",
			cq: cqv1alpha1.CapacityQuota{
				ObjectMeta: metav1.ObjectMeta{Name: "cq1"},
				Spec: cqv1alpha1.CapacityQuotaSpec{
					Selector: nil,
					Limits:   cqv1alpha1.CapacityQuotaLimits{Resources: cqv1alpha1.ResourceList{cqv1alpha1.ResourceCPU: resource.MustParse("1")}},
				},
			},
			node:          &corev1.Node{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"any": "label"}}},
			wantAppliesTo: true,
		},
		{
			name: "invalid-selector",
			cq: cqv1alpha1.CapacityQuota{
				ObjectMeta: metav1.ObjectMeta{Name: "cq1"},
				Spec: cqv1alpha1.CapacityQuotaSpec{
					Selector: &metav1.LabelSelector{
						MatchExpressions: []metav1.LabelSelectorRequirement{
							{
								Key:      "invalidKey!!!!",
								Operator: metav1.LabelSelectorOpExists,
							},
						},
					},
				},
			},
			wantErrMsg: "invalid label selector",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			q, err := newFromCapacityQuota(tc.cq)
			if tc.wantErrMsg == "" {
				if err != nil {
					t.Fatalf("newFromCapacityQuota() unexpected error: %v", err)
				}
				if got := q.AppliesTo(tc.node); got != tc.wantAppliesTo {
					t.Errorf("AppliesTo() = %v, want %v", got, tc.wantAppliesTo)
				}
			} else {
				if err == nil || !strings.Contains(err.Error(), tc.wantErrMsg) {
					t.Errorf("newFromCapacityQuota() want err containing %q, got %v", tc.wantErrMsg, err)
				}
			}
		})
	}
}

func TestCapacityQuota_Limits(t *testing.T) {
	testCases := []struct {
		name       string
		cq         cqv1alpha1.CapacityQuota
		wantLimits map[string]int64
	}{
		{
			name: "cpu-and-memory-limits",
			cq: cqv1alpha1.CapacityQuota{
				ObjectMeta: metav1.ObjectMeta{Name: "cq1"},
				Spec: cqv1alpha1.CapacityQuotaSpec{
					Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"foo": "bar"}},
					Limits: cqv1alpha1.CapacityQuotaLimits{
						Resources: cqv1alpha1.ResourceList{
							cqv1alpha1.ResourceCPU:    resource.MustParse("5"),
							cqv1alpha1.ResourceMemory: resource.MustParse("10Gi"),
						},
					},
				},
			},
			wantLimits: map[string]int64{"cpu": 5, "memory": 10 * 1024 * 1024 * 1024},
		},
		{
			name: "empty-limits",
			cq: cqv1alpha1.CapacityQuota{
				ObjectMeta: metav1.ObjectMeta{Name: "cq1"},
				Spec: cqv1alpha1.CapacityQuotaSpec{
					Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"foo": "bar"}},
					Limits: cqv1alpha1.CapacityQuotaLimits{
						Resources: cqv1alpha1.ResourceList{},
					},
				},
			},
			wantLimits: map[string]int64{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			q, err := newFromCapacityQuota(tc.cq)
			if err != nil {
				t.Fatalf("newFromCapacityQuota() unexpected error: %v", err)
			}
			if diff := cmp.Diff(tc.wantLimits, q.Limits()); diff != "" {
				t.Errorf("Limits() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
