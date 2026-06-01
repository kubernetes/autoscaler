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

package v1alpha1

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cqv1alpha1 "k8s.io/autoscaler/cluster-autoscaler/apis/capacityquota/autoscaling.x-k8s.io/v1alpha1"
)

var _ = Describe("CapacityQuota Webhook", func() {
	var (
		obj *cqv1alpha1.CapacityQuota
	)

	BeforeEach(func() {
		obj = &cqv1alpha1.CapacityQuota{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-quota",
			},
			Spec: cqv1alpha1.CapacityQuotaSpec{
				Limits: cqv1alpha1.CapacityQuotaLimits{
					Resources: cqv1alpha1.ResourceList{
						cqv1alpha1.ResourceCPU: resource.MustParse("10"),
					},
				},
			},
		}
	})

	Context("When creating or updating CapacityQuota under Validating Webhook", func() {
		It("Should admit creation if selector is missing", func() {
			By("creating a CapacityQuota with no selector")
			Expect(k8sClient.Create(ctx, obj)).To(Succeed())

			Expect(k8sClient.Delete(ctx, obj)).To(Succeed())
		})

		It("Should admit creation if selector has valid match labels", func() {
			By("creating a CapacityQuota with a valid selector")
			obj.Spec.Selector = &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "my-app",
				},
			}
			Expect(k8sClient.Create(ctx, obj)).To(Succeed())

			Expect(k8sClient.Delete(ctx, obj)).To(Succeed())
		})

		It("Should admit creation if selector has valid match expressions", func() {
			By("creating a CapacityQuota with a valid selector")
			obj.Spec.Selector = &metav1.LabelSelector{
				MatchExpressions: []metav1.LabelSelectorRequirement{
					{
						Key:      "env",
						Operator: metav1.LabelSelectorOpIn,
						Values:   []string{"prod", "staging"},
					},
				},
			}
			Expect(k8sClient.Create(ctx, obj)).To(Succeed())

			Expect(k8sClient.Delete(ctx, obj)).To(Succeed())
		})

		It("Should deny creation if selector has invalid match expressions", func() {
			By("trying to create a CapacityQuota with an invalid label selector")
			obj.Spec.Selector = &metav1.LabelSelector{
				MatchExpressions: []metav1.LabelSelectorRequirement{
					{
						Key:      "env",
						Operator: "InvalidOperator",
						Values:   []string{"prod"},
					},
				},
			}
			err := k8sClient.Create(ctx, obj)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("is not a valid label selector operator"))
		})

		It("Should deny update if selector is modified to be invalid", func() {
			By("creating a CapacityQuota with a valid selector")
			obj.Spec.Selector = &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "my-app",
				},
			}
			Expect(k8sClient.Create(ctx, obj)).To(Succeed())

			By("trying to update the CapacityQuota with an invalid selector")
			var updatedObj cqv1alpha1.CapacityQuota
			Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(obj), &updatedObj)).To(Succeed())

			updatedObj.Spec.Selector = &metav1.LabelSelector{
				MatchExpressions: []metav1.LabelSelectorRequirement{
					{
						Key:      "env",
						Operator: "InvalidOperator",
						Values:   []string{"prod"},
					},
				},
			}

			err := k8sClient.Update(ctx, &updatedObj)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("is not a valid label selector operator"))

			Expect(k8sClient.Delete(ctx, obj)).To(Succeed())
		})
	})
})
