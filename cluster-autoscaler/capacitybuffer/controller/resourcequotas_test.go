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

package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	v1 "k8s.io/autoscaler/cluster-autoscaler/apis/capacitybuffer/autoscaling.x-k8s.io/v1beta1"
	"k8s.io/autoscaler/cluster-autoscaler/capacitybuffer"
	cbclient "k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/client"
	"k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/testutil"
	podutils "k8s.io/autoscaler/cluster-autoscaler/utils/pod"
	fakeClient "k8s.io/client-go/kubernetes/fake"
)

func TestResourceQuotaAllocator(t *testing.T) {
	podTemp := testutil.NewPodTemplate(
		testutil.WithPodTemplateName("podTemp"),
		testutil.WithPodTemplateResources(
			corev1.ResourceList{
				"cpu":    resource.MustParse("1"),
				"memory": resource.MustParse("1Gi"),
			},
			nil,
		),
	)

	rqBase := testutil.NewResourceQuota(
		testutil.WithResourceQuotaName("quota"),
		testutil.WithResourceQuotaHard(corev1.ResourceList{
			"cpu":    resource.MustParse("10"),
			"memory": resource.MustParse("10Gi"),
		}),
		testutil.WithResourceQuotaUsed(corev1.ResourceList{
			"cpu":    resource.MustParse("0"),
			"memory": resource.MustParse("0"),
		}),
	)

	tests := []struct {
		name               string
		podTemplates       []*corev1.PodTemplate
		quotas             []*corev1.ResourceQuota
		buffers            []*v1.CapacityBuffer
		wantReplicas       []int32
		wantExceededQuotas [][]string
	}{
		{
			name: "no quotas",
			buffers: []*v1.CapacityBuffer{
				testutil.NewBuffer(
					testutil.WithStatusPodTemplateRef("podTemp"),
					testutil.WithStatusPodTemplateGeneration(0),
					testutil.WithStatusReplicas(5),
				),
			},
			wantReplicas:       []int32{5},
			wantExceededQuotas: [][]string{nil},
		},
		{
			name:   "single buffer fits within quota",
			quotas: []*corev1.ResourceQuota{rqBase},
			buffers: []*v1.CapacityBuffer{
				testutil.NewBuffer(
					testutil.WithStatusPodTemplateRef("podTemp"),
					testutil.WithStatusPodTemplateGeneration(0),
					testutil.WithStatusReplicas(5),
				),
			},
			wantReplicas:       []int32{5},
			wantExceededQuotas: [][]string{nil},
		},
		{
			name:   "multiple buffers, second limited by quota",
			quotas: []*corev1.ResourceQuota{rqBase},
			buffers: []*v1.CapacityBuffer{
				testutil.NewBuffer(
					testutil.WithStatusPodTemplateRef("podTemp"),
					testutil.WithStatusPodTemplateGeneration(0),
					testutil.WithStatusReplicas(5),
				),
				testutil.NewBuffer(
					testutil.WithStatusPodTemplateRef("podTemp"),
					testutil.WithStatusPodTemplateGeneration(0),
					testutil.WithStatusReplicas(10),
				),
			},
			wantReplicas:       []int32{5, 5}, // 10 total capacity - 5 used by buffer1 = 5 available for buffer2
			wantExceededQuotas: [][]string{nil, {"quota"}},
		},
		{
			name: "buffer limited by existing usage",
			quotas: []*corev1.ResourceQuota{
				testutil.NewResourceQuota(
					testutil.WithResourceQuotaName("quota-usage"),
					testutil.WithResourceQuotaHard(corev1.ResourceList{"cpu": resource.MustParse("10")}),
					testutil.WithResourceQuotaUsed(corev1.ResourceList{"cpu": resource.MustParse("8")}),
				),
			},
			buffers: []*v1.CapacityBuffer{
				testutil.NewBuffer(
					testutil.WithStatusPodTemplateRef("podTemp"),
					testutil.WithStatusPodTemplateGeneration(0),
					testutil.WithStatusReplicas(5),
				),
			},
			wantReplicas:       []int32{2}, // 10 total - 8 used = 2 available
			wantExceededQuotas: [][]string{{"quota-usage"}},
		},
		{
			name: "multiple quotas on the same resource, both limit the buffer",
			quotas: []*corev1.ResourceQuota{
				testutil.NewResourceQuota(
					testutil.WithResourceQuotaName("q1"),
					testutil.WithResourceQuotaHard(corev1.ResourceList{"cpu": resource.MustParse("5")}),
					testutil.WithResourceQuotaUsed(corev1.ResourceList{"cpu": resource.MustParse("4")}),
				),
				testutil.NewResourceQuota(
					testutil.WithResourceQuotaName("q2"),
					testutil.WithResourceQuotaHard(corev1.ResourceList{"cpu": resource.MustParse("10")}),
					testutil.WithResourceQuotaUsed(corev1.ResourceList{"cpu": resource.MustParse("8")}),
				),
			},
			buffers: []*v1.CapacityBuffer{
				testutil.NewBuffer(
					testutil.WithStatusPodTemplateRef("podTemp"),
					testutil.WithStatusPodTemplateGeneration(0),
					testutil.WithStatusReplicas(5),
				),
			},
			wantReplicas:       []int32{1},
			wantExceededQuotas: [][]string{{"q1", "q2"}},
		},
		{
			name: "requests prefixes are handled",
			quotas: []*corev1.ResourceQuota{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "quota-requests",
						Namespace: "default",
						UID:       types.UID("quota-uid"),
					},
					Status: corev1.ResourceQuotaStatus{
						Hard: corev1.ResourceList{
							"requests.cpu": resource.MustParse("10"),
						},
						Used: corev1.ResourceList{
							"requests.cpu": resource.MustParse("8"),
						},
					},
				},
			},
			buffers: []*v1.CapacityBuffer{
				testutil.NewBuffer(
					testutil.WithStatusPodTemplateRef("podTemp"),
					testutil.WithStatusPodTemplateGeneration(0),
					testutil.WithStatusReplicas(5),
				),
			},
			wantReplicas:       []int32{2}, // 10 total - 8 used = 2 available
			wantExceededQuotas: [][]string{{"quota-requests"}},
		},
		{
			name: "pod defaulting works: limits imply requests",
			podTemplates: []*corev1.PodTemplate{
				testutil.NewPodTemplate(
					testutil.WithPodTemplateName("pod-with-limits"),
					testutil.WithPodTemplateResources(
						nil, // No requests
						corev1.ResourceList{"cpu": resource.MustParse("1")},
					),
				),
			},
			quotas: []*corev1.ResourceQuota{
				testutil.NewResourceQuota(
					testutil.WithResourceQuotaName("quota-reqs"),
					testutil.WithResourceQuotaHard(corev1.ResourceList{"requests.cpu": resource.MustParse("2")}),
					testutil.WithResourceQuotaUsed(corev1.ResourceList{"requests.cpu": resource.MustParse("0")}),
				),
			},
			buffers: []*v1.CapacityBuffer{
				testutil.NewBuffer(
					testutil.WithStatusPodTemplateRef("pod-with-limits"),
					testutil.WithStatusPodTemplateGeneration(0),
					testutil.WithStatusReplicas(5),
				),
			},
			wantReplicas:       []int32{2}, // 5 requested, but only 2 fit
			wantExceededQuotas: [][]string{{"quota-reqs"}},
		},
		{
			name: "quota scopes: terminating scope ignored",
			podTemplates: []*corev1.PodTemplate{
				testutil.NewPodTemplate(
					testutil.WithPodTemplateName("buffer-pod"),
					testutil.WithPodTemplateResources(corev1.ResourceList{"cpu": resource.MustParse("1")}, nil),
				),
			},
			quotas: []*corev1.ResourceQuota{
				testutil.NewResourceQuota(
					testutil.WithResourceQuotaName("quota-terminating"),
					testutil.WithResourceQuotaHard(corev1.ResourceList{"cpu": resource.MustParse("2")}),
					testutil.WithResourceQuotaScopes([]corev1.ResourceQuotaScope{corev1.ResourceQuotaScopeTerminating}),
				),
			},
			buffers: []*v1.CapacityBuffer{
				testutil.NewBuffer(
					testutil.WithStatusPodTemplateRef("buffer-pod"),
					testutil.WithStatusPodTemplateGeneration(0),
					testutil.WithStatusReplicas(5),
				),
			},
			wantReplicas:       []int32{5}, // Quota ignored because it's scoped to Terminating, and buffers are treated as NotTerminating
			wantExceededQuotas: [][]string{nil},
		},
		{
			name: "quota scopes: not terminating scope applies",
			podTemplates: []*corev1.PodTemplate{
				testutil.NewPodTemplate(
					testutil.WithPodTemplateName("pod"),
					testutil.WithPodTemplateResources(corev1.ResourceList{"cpu": resource.MustParse("1")}, nil),
				),
			},
			quotas: []*corev1.ResourceQuota{
				testutil.NewResourceQuota(
					testutil.WithResourceQuotaName("quota-not-terminating"),
					testutil.WithResourceQuotaHard(corev1.ResourceList{"cpu": resource.MustParse("2")}),
					testutil.WithResourceQuotaScopes([]corev1.ResourceQuotaScope{corev1.ResourceQuotaScopeNotTerminating}),
				),
			},
			buffers: []*v1.CapacityBuffer{
				testutil.NewBuffer(
					testutil.WithStatusPodTemplateRef("pod"),
					testutil.WithStatusPodTemplateGeneration(0),
					testutil.WithStatusReplicas(5),
				),
			},
			wantReplicas:       []int32{2}, // Quota applies
			wantExceededQuotas: [][]string{{"quota-not-terminating"}},
		},
		{
			name: "quota scopes: best effort scope applies to best effort pod",
			podTemplates: []*corev1.PodTemplate{
				testutil.NewPodTemplate(
					testutil.WithPodTemplateName("best-effort-pod"),
					testutil.WithPodTemplateResources(corev1.ResourceList{"nvidia.com/gpu": resource.MustParse("1")}, nil),
					// No CPU or memory -> BestEffort
				),
			},
			quotas: []*corev1.ResourceQuota{
				testutil.NewResourceQuota(
					testutil.WithResourceQuotaName("quota-best-effort"),
					testutil.WithResourceQuotaHard(corev1.ResourceList{"requests.nvidia.com/gpu": resource.MustParse("2")}),
					testutil.WithResourceQuotaScopes([]corev1.ResourceQuotaScope{corev1.ResourceQuotaScopeBestEffort}),
				),
			},
			buffers: []*v1.CapacityBuffer{
				testutil.NewBuffer(
					testutil.WithStatusPodTemplateRef("best-effort-pod"),
					testutil.WithStatusPodTemplateGeneration(0),
					testutil.WithStatusReplicas(5),
				),
			},
			wantReplicas:       []int32{2}, // Quota applies
			wantExceededQuotas: [][]string{{"quota-best-effort"}},
		},
		{
			name: "quota scopes: best effort scope ignored for non-best effort pod",
			podTemplates: []*corev1.PodTemplate{
				testutil.NewPodTemplate(
					testutil.WithPodTemplateName("burstable-pod"),
					testutil.WithPodTemplateResources(corev1.ResourceList{"cpu": resource.MustParse("1")}, nil),
				),
			},
			quotas: []*corev1.ResourceQuota{
				testutil.NewResourceQuota(
					testutil.WithResourceQuotaName("quota-best-effort"),
					testutil.WithResourceQuotaHard(corev1.ResourceList{"cpu": resource.MustParse("2")}),
					testutil.WithResourceQuotaScopes([]corev1.ResourceQuotaScope{corev1.ResourceQuotaScopeBestEffort}),
				),
			},
			buffers: []*v1.CapacityBuffer{
				testutil.NewBuffer(
					testutil.WithStatusPodTemplateRef("burstable-pod"),
					testutil.WithStatusPodTemplateGeneration(0),
					testutil.WithStatusReplicas(5),
				),
			},
			wantReplicas:       []int32{5}, // Quota ignored
			wantExceededQuotas: [][]string{nil},
		},
		{
			name: "quota scopes: priority class scope applies",
			podTemplates: []*corev1.PodTemplate{
				testutil.NewPodTemplate(
					testutil.WithPodTemplateName("priority-pod"),
					testutil.WithPodTemplatePriorityClassName("high-priority"),
					testutil.WithPodTemplateResources(corev1.ResourceList{"cpu": resource.MustParse("1")}, nil),
				),
				testutil.NewPodTemplate(
					testutil.WithPodTemplateName("regular-pod"),
					testutil.WithPodTemplateResources(corev1.ResourceList{"cpu": resource.MustParse("1")}, nil),
				),
			},
			quotas: []*corev1.ResourceQuota{
				testutil.NewResourceQuota(
					testutil.WithResourceQuotaName("quota-priority"),
					testutil.WithResourceQuotaHard(corev1.ResourceList{"cpu": resource.MustParse("2")}),
					testutil.WithResourceQuotaScopeSelector(&corev1.ScopeSelector{
						MatchExpressions: []corev1.ScopedResourceSelectorRequirement{
							{
								ScopeName: corev1.ResourceQuotaScopePriorityClass,
								Operator:  corev1.ScopeSelectorOpIn,
								Values:    []string{"high-priority"},
							},
						},
					}),
				),
			},
			buffers: []*v1.CapacityBuffer{
				// we want the regular to be processed first to ensure that we don't update usages
				// of not matching quotas.
				testutil.NewBuffer(
					testutil.WithStatusPodTemplateRef("regular-pod"),
					testutil.WithStatusPodTemplateGeneration(0),
					testutil.WithStatusReplicas(5),
				),
				testutil.NewBuffer(
					testutil.WithStatusPodTemplateRef("priority-pod"),
					testutil.WithStatusPodTemplateGeneration(0),
					testutil.WithStatusReplicas(5),
				),
			},
			wantReplicas:       []int32{5, 2}, // First is ignored, the second one matches the scope and is limited
			wantExceededQuotas: [][]string{nil, {"quota-priority"}},
		},
		{
			name: "multiple quotas: buffers limited by different quotas",
			podTemplates: []*corev1.PodTemplate{
				testutil.NewPodTemplate(
					testutil.WithPodTemplateName("pod-a"),
					testutil.WithPodTemplateResources(corev1.ResourceList{"cpu": resource.MustParse("1")}, nil),
				),
				testutil.NewPodTemplate(
					testutil.WithPodTemplateName("pod-b"),
					testutil.WithPodTemplateResources(corev1.ResourceList{"memory": resource.MustParse("1Gi")}, nil),
				),
			},
			quotas: []*corev1.ResourceQuota{
				testutil.NewResourceQuota(
					testutil.WithResourceQuotaName("quota-cpu"),
					testutil.WithResourceQuotaHard(corev1.ResourceList{"cpu": resource.MustParse("2")}),
				),
				testutil.NewResourceQuota(
					testutil.WithResourceQuotaName("quota-mem"),
					testutil.WithResourceQuotaHard(corev1.ResourceList{"memory": resource.MustParse("2Gi")}),
				),
			},
			buffers: []*v1.CapacityBuffer{
				testutil.NewBuffer(
					testutil.WithStatusPodTemplateRef("pod-a"),
					testutil.WithStatusPodTemplateGeneration(0),
					testutil.WithStatusReplicas(5),
				),
				testutil.NewBuffer(
					testutil.WithStatusPodTemplateRef("pod-b"),
					testutil.WithStatusPodTemplateGeneration(0),
					testutil.WithStatusReplicas(5),
				),
			},
			wantReplicas:       []int32{2, 2}, // pod-a limited by CPU quota, pod-b limited by Memory quota
			wantExceededQuotas: [][]string{{"quota-cpu"}, {"quota-mem"}},
		},
		{
			name: "quota scopes: cross-namespace pod affinity scope applies",
			podTemplates: []*corev1.PodTemplate{
				testutil.NewPodTemplate(
					testutil.WithPodTemplateName("affinity-pod"),
					testutil.WithPodTemplateResources(corev1.ResourceList{"cpu": resource.MustParse("1")}, nil),
					testutil.WithPodTemplateAffinity(&corev1.Affinity{
						PodAffinity: &corev1.PodAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
								{
									LabelSelector: &metav1.LabelSelector{
										MatchLabels: map[string]string{"foo": "bar"},
									},
									Namespaces:  []string{"other-ns"},
									TopologyKey: "kubernetes.io/hostname",
								},
							},
						},
					}),
				),
				testutil.NewPodTemplate(
					testutil.WithPodTemplateName("anti-affinity-pod"),
					testutil.WithPodTemplateResources(corev1.ResourceList{"cpu": resource.MustParse("1")}, nil),
					testutil.WithPodTemplateAffinity(&corev1.Affinity{
						PodAntiAffinity: &corev1.PodAntiAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
								{
									LabelSelector: &metav1.LabelSelector{
										MatchLabels: map[string]string{"foo": "bar"},
									},
									Namespaces:  []string{"other-ns"},
									TopologyKey: "kubernetes.io/hostname",
								},
							},
						},
					}),
				),
				testutil.NewPodTemplate(
					testutil.WithPodTemplateName("regular-pod"),
					testutil.WithPodTemplateResources(corev1.ResourceList{"cpu": resource.MustParse("1")}, nil),
				),
			},
			quotas: []*corev1.ResourceQuota{
				testutil.NewResourceQuota(
					testutil.WithResourceQuotaName("quota-cross-ns"),
					testutil.WithResourceQuotaHard(corev1.ResourceList{"cpu": resource.MustParse("2")}),
					testutil.WithResourceQuotaScopes([]corev1.ResourceQuotaScope{corev1.ResourceQuotaScopeCrossNamespacePodAffinity}),
				),
			},
			buffers: []*v1.CapacityBuffer{
				testutil.NewBuffer(
					testutil.WithStatusPodTemplateRef("affinity-pod"),
					testutil.WithStatusPodTemplateGeneration(0),
					testutil.WithStatusReplicas(5),
				),
				testutil.NewBuffer(
					testutil.WithStatusPodTemplateRef("anti-affinity-pod"),
					testutil.WithStatusPodTemplateGeneration(0),
					testutil.WithStatusReplicas(2),
				),
				testutil.NewBuffer(
					testutil.WithStatusPodTemplateRef("regular-pod"),
					testutil.WithStatusPodTemplateGeneration(0),
					testutil.WithStatusReplicas(5),
				),
			},
			wantReplicas:       []int32{2, 0, 5}, // Affinity pod limited, regular pod ignored
			wantExceededQuotas: [][]string{{"quota-cross-ns"}, {"quota-cross-ns"}, nil},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var objs []runtime.Object
			if len(tt.podTemplates) > 0 {
				for _, pt := range tt.podTemplates {
					objs = append(objs, pt)
				}
			} else {
				objs = append(objs, podTemp)
			}
			for _, q := range tt.quotas {
				objs = append(objs, q)
			}

			// Pre-populate PodTemplates with defaulted requests (simulating Admission/DryRun)
			for _, obj := range objs {
				if pt, ok := obj.(*corev1.PodTemplate); ok {
					// Simple defaulting for tests: if limits set and requests not, copy limits to requests
					// This mimics what dry-run would do for the fields relevant to tests
					for i := range pt.Template.Spec.Containers {
						if pt.Template.Spec.Containers[i].Resources.Requests == nil {
							pt.Template.Spec.Containers[i].Resources.Requests = make(corev1.ResourceList)
						}
						for k, v := range pt.Template.Spec.Containers[i].Resources.Limits {
							if _, exists := pt.Template.Spec.Containers[i].Resources.Requests[k]; !exists {
								pt.Template.Spec.Containers[i].Resources.Requests[k] = v
							}
						}
					}
				}
			}

			fakeK8s := fakeClient.NewSimpleClientset(objs...)
			client, _ := cbclient.NewCapacityBufferClient(nil, fakeK8s, nil, nil, nil, nil, nil, nil, nil, nil, nil)
			allocator := newResourceQuotaAllocator(client)

			// Assign namespace and names to buffers
			for i, buffer := range tt.buffers {
				buffer.Namespace = "default"
				if buffer.Name == "" {
					buffer.Name = "buffer-" + string(rune(i))
				}
			}

			errs := allocator.Allocate("default", tt.buffers)
			assert.Empty(t, errs)

			for i, buffer := range tt.buffers {
				assert.Equal(t, tt.wantReplicas[i], *buffer.Status.Replicas, "Buffer %d replicas mismatch", i)

				var limitCondition *metav1.Condition
				for _, c := range buffer.Status.Conditions {
					if c.Type == capacitybuffer.LimitedByQuotasCondition {
						limitCondition = c.DeepCopy()
						break
					}
				}

				if len(tt.wantExceededQuotas[i]) == 0 {
					if limitCondition != nil {
						assert.Equal(t, metav1.ConditionFalse, limitCondition.Status, "Buffer %d should not be limited", i)
					}
				} else {
					assert.NotNil(t, limitCondition, "Buffer %q should be limited", buffer.Name)
					assert.Equal(t, metav1.ConditionTrue, limitCondition.Status, "Buffer %q LimitedByQuotas condition status mismatch", buffer.Name)
					// Check if message contains all expected quotas
					for _, q := range tt.wantExceededQuotas[i] {
						assert.Contains(t, limitCondition.Message, q, "Buffer %d message should contain quota %s", i, q)
					}
				}
			}
		})
	}
}

func TestPodMatchesQuotaScope(t *testing.T) {
	tests := []struct {
		name        string
		podTemplate *corev1.PodTemplate
		quota       *corev1.ResourceQuota
		wantMatch   bool
		wantErr     bool
	}{
		{
			name:        "scope NotTerminating matches",
			podTemplate: testutil.NewPodTemplate(),
			quota: testutil.NewResourceQuota(
				testutil.WithResourceQuotaScopes([]corev1.ResourceQuotaScope{corev1.ResourceQuotaScopeNotTerminating}),
			),
			wantMatch: true,
		},
		{
			name:        "scope Terminating does not match",
			podTemplate: testutil.NewPodTemplate(),
			quota: testutil.NewResourceQuota(
				testutil.WithResourceQuotaScopes([]corev1.ResourceQuotaScope{corev1.ResourceQuotaScopeTerminating}),
			),
			wantMatch: false,
		},
		{
			name:        "scope BestEffort matches BestEffort pod",
			podTemplate: testutil.NewPodTemplate(), // No resources -> BestEffort
			quota: testutil.NewResourceQuota(
				testutil.WithResourceQuotaScopes([]corev1.ResourceQuotaScope{corev1.ResourceQuotaScopeBestEffort}),
			),
			wantMatch: true,
		},
		{
			name: "scope BestEffort does not match Burstable pod",
			podTemplate: testutil.NewPodTemplate(
				testutil.WithPodTemplateResources(corev1.ResourceList{"cpu": resource.MustParse("100m")}, nil),
			),
			quota: testutil.NewResourceQuota(
				testutil.WithResourceQuotaScopes([]corev1.ResourceQuotaScope{corev1.ResourceQuotaScopeBestEffort}),
			),
			wantMatch: false,
		},
		{
			name: "scope NotBestEffort matches Burstable pod",
			podTemplate: testutil.NewPodTemplate(
				testutil.WithPodTemplateResources(corev1.ResourceList{"cpu": resource.MustParse("100m")}, nil),
			),
			quota: testutil.NewResourceQuota(
				testutil.WithResourceQuotaScopes([]corev1.ResourceQuotaScope{corev1.ResourceQuotaScopeNotBestEffort}),
			),
			wantMatch: true,
		},
		{
			name:        "scope NotBestEffort does not match BestEffort pod",
			podTemplate: testutil.NewPodTemplate(),
			quota: testutil.NewResourceQuota(
				testutil.WithResourceQuotaScopes([]corev1.ResourceQuotaScope{corev1.ResourceQuotaScopeNotBestEffort}),
			),
			wantMatch: false,
		},
		{
			name: "ScopeSelector PriorityClass OpIn matches",
			podTemplate: testutil.NewPodTemplate(
				testutil.WithPodTemplatePriorityClassName("high"),
			),
			quota: testutil.NewResourceQuota(
				testutil.WithResourceQuotaScopeSelector(&corev1.ScopeSelector{
					MatchExpressions: []corev1.ScopedResourceSelectorRequirement{
						{
							ScopeName: corev1.ResourceQuotaScopePriorityClass,
							Operator:  corev1.ScopeSelectorOpIn,
							Values:    []string{"high", "medium"},
						},
					},
				}),
			),
			wantMatch: true,
		},
		{
			name: "ScopeSelector PriorityClass OpIn does not match",
			podTemplate: testutil.NewPodTemplate(
				testutil.WithPodTemplatePriorityClassName("low"),
			),
			quota: testutil.NewResourceQuota(
				testutil.WithResourceQuotaScopeSelector(&corev1.ScopeSelector{
					MatchExpressions: []corev1.ScopedResourceSelectorRequirement{
						{
							ScopeName: corev1.ResourceQuotaScopePriorityClass,
							Operator:  corev1.ScopeSelectorOpIn,
							Values:    []string{"high", "medium"},
						},
					},
				}),
			),
			wantMatch: false,
		},
		{
			name: "ScopeSelector PriorityClass OpNotIn matches",
			podTemplate: testutil.NewPodTemplate(
				testutil.WithPodTemplatePriorityClassName("low"),
			),
			quota: testutil.NewResourceQuota(
				testutil.WithResourceQuotaScopeSelector(&corev1.ScopeSelector{
					MatchExpressions: []corev1.ScopedResourceSelectorRequirement{
						{
							ScopeName: corev1.ResourceQuotaScopePriorityClass,
							Operator:  corev1.ScopeSelectorOpNotIn,
							Values:    []string{"high", "medium"},
						},
					},
				}),
			),
			wantMatch: true,
		},
		{
			name: "ScopeSelector PriorityClass OpNotIn does not match",
			podTemplate: testutil.NewPodTemplate(
				testutil.WithPodTemplatePriorityClassName("high"),
			),
			quota: testutil.NewResourceQuota(
				testutil.WithResourceQuotaScopeSelector(&corev1.ScopeSelector{
					MatchExpressions: []corev1.ScopedResourceSelectorRequirement{
						{
							ScopeName: corev1.ResourceQuotaScopePriorityClass,
							Operator:  corev1.ScopeSelectorOpNotIn,
							Values:    []string{"high", "medium"},
						},
					},
				}),
			),
			wantMatch: false,
		},
		{
			name: "ScopeSelector PriorityClass OpExists matches",
			podTemplate: testutil.NewPodTemplate(
				testutil.WithPodTemplatePriorityClassName("any"),
			),
			quota: testutil.NewResourceQuota(
				testutil.WithResourceQuotaScopeSelector(&corev1.ScopeSelector{
					MatchExpressions: []corev1.ScopedResourceSelectorRequirement{
						{
							ScopeName: corev1.ResourceQuotaScopePriorityClass,
							Operator:  corev1.ScopeSelectorOpExists,
						},
					},
				}),
			),
			wantMatch: true,
		},
		{
			name:        "ScopeSelector PriorityClass OpExists does not match",
			podTemplate: testutil.NewPodTemplate(), // No PC
			quota: testutil.NewResourceQuota(
				testutil.WithResourceQuotaScopeSelector(&corev1.ScopeSelector{
					MatchExpressions: []corev1.ScopedResourceSelectorRequirement{
						{
							ScopeName: corev1.ResourceQuotaScopePriorityClass,
							Operator:  corev1.ScopeSelectorOpExists,
						},
					},
				}),
			),
			wantMatch: false,
		},
		{
			name:        "ScopeSelector PriorityClass OpDoesNotExist matches",
			podTemplate: testutil.NewPodTemplate(), // No PC
			quota: testutil.NewResourceQuota(
				testutil.WithResourceQuotaScopeSelector(&corev1.ScopeSelector{
					MatchExpressions: []corev1.ScopedResourceSelectorRequirement{
						{
							ScopeName: corev1.ResourceQuotaScopePriorityClass,
							Operator:  corev1.ScopeSelectorOpDoesNotExist,
						},
					},
				}),
			),
			wantMatch: true,
		},
		{
			name: "ScopeSelector PriorityClass OpDoesNotExist does not match",
			podTemplate: testutil.NewPodTemplate(
				testutil.WithPodTemplatePriorityClassName("any"),
			),
			quota: testutil.NewResourceQuota(
				testutil.WithResourceQuotaScopeSelector(&corev1.ScopeSelector{
					MatchExpressions: []corev1.ScopedResourceSelectorRequirement{
						{
							ScopeName: corev1.ResourceQuotaScopePriorityClass,
							Operator:  corev1.ScopeSelectorOpDoesNotExist,
						},
					},
				}),
			),
			wantMatch: false,
		},
		{
			name:        "ScopeSelector PriorityClass Invalid Operator error",
			podTemplate: testutil.NewPodTemplate(),
			quota: testutil.NewResourceQuota(
				testutil.WithResourceQuotaScopeSelector(&corev1.ScopeSelector{
					MatchExpressions: []corev1.ScopedResourceSelectorRequirement{
						{
							ScopeName: corev1.ResourceQuotaScopePriorityClass,
							Operator:  "invalid",
							Values:    []string{"foo"},
						},
					},
				}),
			),
			wantErr: true,
		},
		{
			name:        "ScopeSelector PriorityClass Error from labels.NewRequirement (OpIn empty values)",
			podTemplate: testutil.NewPodTemplate(),
			quota: testutil.NewResourceQuota(
				testutil.WithResourceQuotaScopeSelector(&corev1.ScopeSelector{
					MatchExpressions: []corev1.ScopedResourceSelectorRequirement{
						{
							ScopeName: corev1.ResourceQuotaScopePriorityClass,
							Operator:  corev1.ScopeSelectorOpIn,
							Values:    []string{}, // Empty values for In should error
						},
					},
				}),
			),
			wantErr: true,
		},
		{
			name: "scope CrossNamespacePodAffinity matches",
			podTemplate: testutil.NewPodTemplate(
				testutil.WithPodTemplateAffinity(&corev1.Affinity{
					PodAffinity: &corev1.PodAffinity{
						RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
							{
								Namespaces: []string{"other"},
							},
						},
					},
				}),
			),
			quota: testutil.NewResourceQuota(
				testutil.WithResourceQuotaScopes([]corev1.ResourceQuotaScope{corev1.ResourceQuotaScopeCrossNamespacePodAffinity}),
			),
			wantMatch: true,
		},
		{
			name:        "scope CrossNamespacePodAffinity does not match",
			podTemplate: testutil.NewPodTemplate(),
			quota: testutil.NewResourceQuota(
				testutil.WithResourceQuotaScopes([]corev1.ResourceQuotaScope{corev1.ResourceQuotaScopeCrossNamespacePodAffinity}),
			),
			wantMatch: false,
		},
		{
			name: "scope CrossNamespacePodAffinity matches (PodAffinity Preferred)",
			podTemplate: testutil.NewPodTemplate(
				testutil.WithPodTemplateAffinity(&corev1.Affinity{
					PodAffinity: &corev1.PodAffinity{
						PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
							{
								Weight: 1,
								PodAffinityTerm: corev1.PodAffinityTerm{
									Namespaces: []string{"other"},
								},
							},
						},
					},
				}),
			),
			quota: testutil.NewResourceQuota(
				testutil.WithResourceQuotaScopes([]corev1.ResourceQuotaScope{corev1.ResourceQuotaScopeCrossNamespacePodAffinity}),
			),
			wantMatch: true,
		},
		{
			name: "scope CrossNamespacePodAffinity matches (PodAntiAffinity Required)",
			podTemplate: testutil.NewPodTemplate(
				testutil.WithPodTemplateAffinity(&corev1.Affinity{
					PodAntiAffinity: &corev1.PodAntiAffinity{
						RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
							{
								Namespaces: []string{"other"},
							},
						},
					},
				}),
			),
			quota: testutil.NewResourceQuota(
				testutil.WithResourceQuotaScopes([]corev1.ResourceQuotaScope{corev1.ResourceQuotaScopeCrossNamespacePodAffinity}),
			),
			wantMatch: true,
		},
		{
			name: "scope CrossNamespacePodAffinity matches (PodAntiAffinity Preferred)",
			podTemplate: testutil.NewPodTemplate(
				testutil.WithPodTemplateAffinity(&corev1.Affinity{
					PodAntiAffinity: &corev1.PodAntiAffinity{
						PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
							{
								Weight: 1,
								PodAffinityTerm: corev1.PodAffinityTerm{
									Namespaces: []string{"other"},
								},
							},
						},
					},
				}),
			),
			quota: testutil.NewResourceQuota(
				testutil.WithResourceQuotaScopes([]corev1.ResourceQuotaScope{corev1.ResourceQuotaScopeCrossNamespacePodAffinity}),
			),
			wantMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pod := podutils.GetPodFromTemplate(&tt.podTemplate.Template)
			pod.Namespace = tt.podTemplate.Namespace
			gotMatch, err := podMatchesQuotaScope(pod, tt.quota)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.wantMatch, gotMatch)
		})
	}
}
