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

package controllers

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/autoscaler/cluster-autoscaler/core/utils"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cqv1alpha1 "k8s.io/autoscaler/cluster-autoscaler/apis/capacityquota/autoscaling.x-k8s.io/v1alpha1"
	cqtest "k8s.io/autoscaler/cluster-autoscaler/resourcequotas/capacityquota/testutil"
	testutils "k8s.io/autoscaler/cluster-autoscaler/utils/test"
	"k8s.io/autoscaler/cluster-autoscaler/utils/units"
)

var _ = Describe("CapacityQuota Controller", func() {
	SetDefaultEventuallyTimeout(5 * time.Second)
	SetDefaultEventuallyPollingInterval(100 * time.Millisecond)

	BeforeEach(func() {
		By("creating a few initial nodes")
		node1 := testutils.BuildTestNode("test-node-1", 2000, 8*units.GiB, testutils.WithNodeLabels(
			map[string]string{"node-pool": "test-pool"},
		))
		Expect(crClient.Create(ctx, node1)).To(Succeed())
		Expect(crClient.Status().Update(ctx, node1)).To(Succeed())

		node2 := testutils.BuildTestNode("test-node-2", 4000, 16*units.GiB, testutils.WithNodeLabels(
			map[string]string{"node-pool": "test-pool"},
		))
		Expect(crClient.Create(ctx, node2)).To(Succeed())
		Expect(crClient.Status().Update(ctx, node2)).To(Succeed())

		node3 := testutils.BuildTestNode("other-node-1", 8000, 32*units.GiB, testutils.WithNodeLabels(
			map[string]string{"node-pool": "other-pool"},
		))
		Expect(crClient.Create(ctx, node3)).To(Succeed())
		Expect(crClient.Status().Update(ctx, node3)).To(Succeed())
	})

	AfterEach(func() {
		By("cleaning up nodes")
		err := crClient.DeleteAllOf(ctx, &corev1.Node{}, client.HasLabels{"node-pool"})
		Expect(err).NotTo(HaveOccurred())

		By("cleaning up quotas")
		err = crClient.DeleteAllOf(ctx, &cqv1alpha1.CapacityQuota{})
		Expect(err).NotTo(HaveOccurred())
	})

	It("should correctly sum capacities of all nodes with nil selector", func() {
		By("creating a CapacityQuota without label selector")
		cq := cqtest.NewCapacityQuota("test-quota-all",
			cqtest.WithLimits(cqv1alpha1.ResourceList{
				cqv1alpha1.ResourceCPU: resource.MustParse("20"),
			}),
		)
		Expect(crClient.Create(ctx, cq)).To(Succeed())

		By("waiting for the CapacityQuota to be reconciled with correct usage")
		cqKey := types.NamespacedName{Name: "test-quota-all"}
		Eventually(func(g Gomega) {
			assertQuotaReconciled(ctx, g, cqKey, cqv1alpha1.ResourceList{
				cqv1alpha1.ResourceCPU: resource.MustParse("14"),
			})
		}).Should(Succeed())
	})

	It("should correctly sum capacities of matching nodes", func() {
		By("creating a CapacityQuota")
		cq := cqtest.NewCapacityQuota("test-quota-pool",
			cqtest.WithLabelSelector(map[string]string{"node-pool": "test-pool"}),
			cqtest.WithLimits(cqv1alpha1.ResourceList{
				cqv1alpha1.ResourceCPU:    resource.MustParse("10"),
				cqv1alpha1.ResourceMemory: resource.MustParse("32Gi"),
				cqv1alpha1.ResourceNodes:  resource.MustParse("5"),
				"nvidia.com/gpu":          resource.MustParse("2"),
			}),
		)
		Expect(crClient.Create(ctx, cq)).To(Succeed())

		cqKey := types.NamespacedName{Name: "test-quota-pool"}

		By("waiting for the CapacityQuota to be reconciled with correct usage")
		Eventually(func(g Gomega) {
			assertQuotaReconciled(ctx, g, cqKey, cqv1alpha1.ResourceList{
				cqv1alpha1.ResourceCPU:    resource.MustParse("6"),
				cqv1alpha1.ResourceMemory: resource.MustParse("24Gi"),
				cqv1alpha1.ResourceNodes:  resource.MustParse("2"),
				"nvidia.com/gpu":          resource.MustParse("0"),
			})
		}).Should(Succeed())
	})

	It("should correctly update capacities when a new matching node is added", func() {
		By("creating a CapacityQuota")
		cq := cqtest.NewCapacityQuota("test-quota-pool-add",
			cqtest.WithLabelSelector(map[string]string{"node-pool": "test-pool"}),
			cqtest.WithLimits(cqv1alpha1.ResourceList{
				cqv1alpha1.ResourceCPU: resource.MustParse("10"),
			}),
		)
		Expect(crClient.Create(ctx, cq)).To(Succeed())

		cqKey := types.NamespacedName{Name: "test-quota-pool-add"}

		By("waiting for initial reconciliation")
		Eventually(func(g Gomega) {
			assertQuotaReconciled(ctx, g, cqKey, cqv1alpha1.ResourceList{
				cqv1alpha1.ResourceCPU: resource.MustParse("6"),
			})
		}).Should(Succeed())

		By("adding a new matching node")
		newNode := testutils.BuildTestNode("test-node-3", 2000, 0, testutils.WithNodeLabels(
			map[string]string{"node-pool": "test-pool"},
		))
		Expect(crClient.Create(ctx, newNode)).To(Succeed())
		Expect(crClient.Status().Update(ctx, newNode)).To(Succeed())

		By("ensuring quota was updated")
		Eventually(func(g Gomega) {
			assertQuotaReconciled(ctx, g, cqKey, cqv1alpha1.ResourceList{
				cqv1alpha1.ResourceCPU: resource.MustParse("8"),
			})
		}).Should(Succeed())
	})

	It("should correctly update capacities when a matching node is deleted", func() {
		By("creating a CapacityQuota")
		cq := cqtest.NewCapacityQuota("test-quota-pool-delete",
			cqtest.WithLabelSelector(map[string]string{"node-pool": "test-pool"}),
			cqtest.WithLimits(cqv1alpha1.ResourceList{
				cqv1alpha1.ResourceCPU: resource.MustParse("10"),
			}),
		)
		Expect(crClient.Create(ctx, cq)).To(Succeed())

		cqKey := types.NamespacedName{Name: "test-quota-pool-delete"}

		By("waiting for initial reconciliation")
		Eventually(func(g Gomega) {
			assertQuotaReconciled(ctx, g, cqKey, cqv1alpha1.ResourceList{
				cqv1alpha1.ResourceCPU: resource.MustParse("6"),
			})
		}).Should(Succeed())

		By("deleting a matching node")
		var fetchedNode corev1.Node
		Expect(crClient.Get(ctx, types.NamespacedName{Name: "test-node-1"}, &fetchedNode)).To(Succeed())
		Expect(crClient.Delete(ctx, &fetchedNode)).To(Succeed())

		By("ensuring quota was updated")
		Eventually(func(g Gomega) {
			assertQuotaReconciled(ctx, g, cqKey, cqv1alpha1.ResourceList{
				cqv1alpha1.ResourceCPU: resource.MustParse("4"),
			})
		}).Should(Succeed())
	})

	It("should correctly update capacities when a node's label is removed, causing it to stop matching", func() {
		By("creating the first CapacityQuota")
		cq1 := cqtest.NewCapacityQuota("test-quota-pool-1",
			cqtest.WithLabelSelector(map[string]string{"node-pool": "test-pool"}),
			cqtest.WithLimits(cqv1alpha1.ResourceList{
				cqv1alpha1.ResourceCPU: resource.MustParse("10"),
			}),
		)
		Expect(crClient.Create(ctx, cq1)).To(Succeed())

		By("creating the second CapacityQuota")
		cq2 := cqtest.NewCapacityQuota("test-quota-pool-2",
			cqtest.WithLabelSelector(map[string]string{"node-pool-2": "test-pool-2"}),
			cqtest.WithLimits(cqv1alpha1.ResourceList{
				cqv1alpha1.ResourceCPU: resource.MustParse("10"),
			}),
		)
		Expect(crClient.Create(ctx, cq2)).To(Succeed())

		cqKey1 := types.NamespacedName{Name: "test-quota-pool-1"}
		cqKey2 := types.NamespacedName{Name: "test-quota-pool-2"}

		By("creating a node that matches both quotas")
		newNode := testutils.BuildTestNode("test-node-5", 5000, 0, testutils.WithNodeLabels(
			map[string]string{
				"node-pool":   "test-pool",
				"node-pool-2": "test-pool-2",
			},
		))
		Expect(crClient.Create(ctx, newNode)).To(Succeed())
		Expect(crClient.Status().Update(ctx, newNode)).To(Succeed())

		By("waiting for initial reconciliation")
		Eventually(func(g Gomega) {
			assertQuotaReconciled(ctx, g, cqKey1, cqv1alpha1.ResourceList{
				cqv1alpha1.ResourceCPU: resource.MustParse("11"),
			})
			assertQuotaReconciled(ctx, g, cqKey2, cqv1alpha1.ResourceList{
				cqv1alpha1.ResourceCPU: resource.MustParse("5"),
			})
		}).Should(Succeed())

		By("removing the label from the node, so it no longer matches cq1")
		var fetchedNode corev1.Node
		Expect(crClient.Get(ctx, types.NamespacedName{Name: "test-node-5"}, &fetchedNode)).To(Succeed())
		delete(fetchedNode.Labels, "node-pool")
		Expect(crClient.Update(ctx, &fetchedNode)).To(Succeed())

		By("ensuring cq1 drops the capacity of node5 (11 -> 6), while cq2 remains unchanged (5)")
		Eventually(func(g Gomega) {
			assertQuotaReconciled(ctx, g, cqKey1, cqv1alpha1.ResourceList{
				cqv1alpha1.ResourceCPU: resource.MustParse("6"),
			})
			assertQuotaReconciled(ctx, g, cqKey2, cqv1alpha1.ResourceList{
				cqv1alpha1.ResourceCPU: resource.MustParse("5"),
			})
		}).Should(Succeed())
	})

	It("should exclude nodes matching the configured NodeFilter", func() {
		By("adding a virtual kubelet node that matches the quota's selector")
		// the controller has been initialized with VirtualKubeletNodeFilter, so virtual kubelet nodes should be excluded
		virtualKubeletNode := testutils.BuildTestNode("test-node-virtual", 1000000, 0, testutils.WithNodeLabels(
			map[string]string{
				"node-pool": "test-pool",
				"type":      utils.VirtualKubeletNodeLabelValue,
			},
		))
		Expect(crClient.Create(ctx, virtualKubeletNode)).To(Succeed())
		Expect(crClient.Status().Update(ctx, virtualKubeletNode)).To(Succeed())

		By("creating a CapacityQuota")
		cq := cqtest.NewCapacityQuota("test-quota-pool-virtual",
			cqtest.WithLabelSelector(map[string]string{"node-pool": "test-pool"}),
			cqtest.WithLimits(cqv1alpha1.ResourceList{
				cqv1alpha1.ResourceCPU: resource.MustParse("10"),
			}),
		)
		Expect(crClient.Create(ctx, cq)).To(Succeed())

		cqKey := types.NamespacedName{Name: "test-quota-pool-virtual"}

		By("ensuring quota does not include the virtual node's capacity")
		Eventually(func(g Gomega) {
			assertQuotaReconciled(ctx, g, cqKey, cqv1alpha1.ResourceList{
				cqv1alpha1.ResourceCPU: resource.MustParse("6"),
			})
		}).Should(Succeed())
	})

	It("should report a failure condition if the selector is invalid", func() {
		By("creating a CapacityQuota with an invalid selector")
		cq := cqtest.NewCapacityQuota("test-quota-pool-invalid",
			func(cq *cqv1alpha1.CapacityQuota) {
				cq.Spec.Selector = &metav1.LabelSelector{
					MatchExpressions: []metav1.LabelSelectorRequirement{
						{
							Key:      "node-pool",
							Operator: "InvalidOperator",
						},
					},
				}
			},
			cqtest.WithLimits(cqv1alpha1.ResourceList{
				cqv1alpha1.ResourceCPU: resource.MustParse("10"),
			}),
		)
		Expect(crClient.Create(ctx, cq)).To(Succeed())

		cqKey := types.NamespacedName{Name: "test-quota-pool-invalid"}

		By("waiting for the CapacityQuota to be reconciled with a failed condition")
		Eventually(func(g Gomega) {
			var fetchedCQ cqv1alpha1.CapacityQuota
			g.Expect(crClient.Get(ctx, cqKey, &fetchedCQ)).To(Succeed())

			g.Expect(meta.IsStatusConditionFalse(fetchedCQ.Status.Conditions, cqv1alpha1.ValidCondition)).To(BeTrue())
			cond := meta.FindStatusCondition(fetchedCQ.Status.Conditions, cqv1alpha1.ValidCondition)
			g.Expect(cond).NotTo(BeNil())
			g.Expect(cond.Reason).To(Equal("ValidationFailed"))
			g.Expect(cond.Message).To(ContainSubstring("is not a valid label selector operator"))

			g.Expect(meta.IsStatusConditionFalse(fetchedCQ.Status.Conditions, cqv1alpha1.ReconciledCondition)).To(BeTrue())
		}).Should(Succeed())
	})
})

func assertQuotaReconciled(ctx context.Context, g Gomega, cqKey types.NamespacedName, wantResources cqv1alpha1.ResourceList) {
	var fetchedCQ cqv1alpha1.CapacityQuota
	g.Expect(crClient.Get(ctx, cqKey, &fetchedCQ)).To(Succeed())

	g.Expect(meta.IsStatusConditionTrue(fetchedCQ.Status.Conditions, cqv1alpha1.ValidCondition)).To(BeTrue())
	g.Expect(meta.IsStatusConditionTrue(fetchedCQ.Status.Conditions, cqv1alpha1.ReconciledCondition)).To(BeTrue())
	g.Expect(fetchedCQ.Status.Used).ToNot(BeNil())
	g.Expect(apiequality.Semantic.DeepEqual(fetchedCQ.Status.Used.Resources, wantResources)).To(BeTrue())
}
