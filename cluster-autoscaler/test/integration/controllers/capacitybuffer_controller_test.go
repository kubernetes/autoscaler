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

package controllers

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cbapi "k8s.io/autoscaler/cluster-autoscaler/capacitybuffer"
	"k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/testutil"
	"k8s.io/utils/ptr"
)

var _ = Describe("CapacityBuffer Controller", func() {
	var namespace string

	Context("ResourceQuotas Integration", func() {
		SetDefaultEventuallyTimeout(5 * time.Second)
		SetDefaultEventuallyPollingInterval(100 * time.Millisecond)

		BeforeEach(func() {
			By("creating a test namespace")
			testNS := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "test-ns-",
				},
			}
			testNS, err := k8sClient.CoreV1().Namespaces().Create(ctx, testNS, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())

			namespace = testNS.Name

			By("creating a pod template")
			podTemp := testutil.NewPodTemplate(
				testutil.WithPodTemplateName("pod-temp"),
				testutil.WithPodTemplateResources(corev1.ResourceList{"cpu": resource.MustParse("1")}, nil),
			)
			podTemp.Template.Spec.Containers[0].Image = "nginx"
			podTemp.Namespace = namespace
			_, err = k8sClient.CoreV1().PodTemplates(namespace).Create(ctx, podTemp, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())

			By("creating a resource quota")
			rq := testutil.NewResourceQuota(
				testutil.WithResourceQuotaName("quota"),
				testutil.WithResourceQuotaHard(corev1.ResourceList{"cpu": resource.MustParse("5")}),
			)
			rq.Namespace = namespace
			rq, err = k8sClient.CoreV1().ResourceQuotas(namespace).Create(ctx, rq, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())

			By("setting resource quota's status")
			rq.Status.Hard = rq.Spec.Hard
			rq.Status.Used = corev1.ResourceList{"cpu": resource.MustParse("0")}
			_, err = k8sClient.CoreV1().ResourceQuotas(namespace).UpdateStatus(ctx, rq, metav1.UpdateOptions{})
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			By("cleaning up test resources")
			_ = k8sClient.CoreV1().PodTemplates(namespace).DeleteCollection(ctx, metav1.DeleteOptions{}, metav1.ListOptions{})
			_ = k8sClient.CoreV1().ResourceQuotas(namespace).DeleteCollection(ctx, metav1.DeleteOptions{}, metav1.ListOptions{})
			_ = buffersClient.AutoscalingV1beta1().CapacityBuffers(namespace).DeleteCollection(ctx, metav1.DeleteOptions{}, metav1.ListOptions{})
			_ = k8sClient.CoreV1().Namespaces().Delete(ctx, namespace, metav1.DeleteOptions{})
		})

		It("should respect k8s ResourceQuotas", func() {
			By("creating a buffer that does not exceed remaining quota")
			b1 := testutil.NewBuffer(
				testutil.WithName("b1"),
				testutil.WithPodTemplateRef("pod-temp"),
				testutil.WithReplicas(2),
				testutil.WithActiveProvisioningStrategy(),
			)
			b1.Namespace = namespace
			_, err := buffersClient.AutoscalingV1beta1().CapacityBuffers(namespace).Create(ctx, b1, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())

			By("checking that the buffer's replicas were not limited")
			Eventually(func(g Gomega) {
				b, err := buffersClient.AutoscalingV1beta1().CapacityBuffers(namespace).Get(ctx, "b1", metav1.GetOptions{})
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(b.Status.Replicas).To(Equal(ptr.To[int32](2)))
				g.Expect(meta.IsStatusConditionFalse(b.Status.Conditions, cbapi.LimitedByQuotasCondition)).To(BeTrue())
			}).Should(Succeed())

			By("creating a buffer that exceeds remaining quota")
			b2 := testutil.NewBuffer(
				testutil.WithName("b2"),
				testutil.WithPodTemplateRef("pod-temp"),
				testutil.WithReplicas(4),
				testutil.WithActiveProvisioningStrategy(),
			)
			b2.Namespace = namespace
			_, err = buffersClient.AutoscalingV1beta1().CapacityBuffers(namespace).Create(ctx, b2, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())

			By("checking that the new buffer's replicas were limited")
			Eventually(func(g Gomega) {
				b, err := buffersClient.AutoscalingV1beta1().CapacityBuffers(namespace).Get(ctx, "b2", metav1.GetOptions{})
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(b.Status.Replicas).To(Equal(ptr.To[int32](3)))
				g.Expect(meta.IsStatusConditionTrue(b.Status.Conditions, cbapi.LimitedByQuotasCondition)).To(BeTrue())
			}).Should(Succeed())

			By("increasing b1's replicas from 2 to 4")
			Eventually(func(g Gomega) {
				b, err := buffersClient.AutoscalingV1beta1().CapacityBuffers(namespace).Get(ctx, "b1", metav1.GetOptions{})
				g.Expect(err).NotTo(HaveOccurred())
				b.Spec.Replicas = ptr.To[int32](4)
				_, err = buffersClient.AutoscalingV1beta1().CapacityBuffers(namespace).Update(ctx, b, metav1.UpdateOptions{})
				g.Expect(err).NotTo(HaveOccurred())
			}).Should(Succeed())

			By("checking that b1's replicas are still not limited")
			Eventually(func(g Gomega) {
				b, err := buffersClient.AutoscalingV1beta1().CapacityBuffers(namespace).Get(ctx, "b1", metav1.GetOptions{})
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(b.Status.Replicas).To(Equal(ptr.To[int32](4)))
				g.Expect(meta.IsStatusConditionFalse(b.Status.Conditions, cbapi.LimitedByQuotasCondition)).To(BeTrue())
			}).Should(Succeed())

			// buffers are ordered in the order of creation
			By("checking that b2's replicas dropped from 3 to 1")
			Eventually(func(g Gomega) {
				b, err := buffersClient.AutoscalingV1beta1().CapacityBuffers(namespace).Get(ctx, "b2", metav1.GetOptions{})
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(b.Status.Replicas).To(Equal(ptr.To[int32](1)))
				g.Expect(meta.IsStatusConditionTrue(b.Status.Conditions, cbapi.LimitedByQuotasCondition)).To(BeTrue())
			}).Should(Succeed())

			By("increasing quota from 5 to 10")
			Eventually(func(g Gomega) {
				rq, err := k8sClient.CoreV1().ResourceQuotas(namespace).Get(ctx, "quota", metav1.GetOptions{})
				g.Expect(err).NotTo(HaveOccurred())
				rq.Spec.Hard["cpu"] = resource.MustParse("10")
				rq, err = k8sClient.CoreV1().ResourceQuotas(namespace).Update(ctx, rq, metav1.UpdateOptions{})
				g.Expect(err).NotTo(HaveOccurred())
				rq.Status.Hard["cpu"] = resource.MustParse("10")
				g.Expect(k8sClient.CoreV1().ResourceQuotas(namespace).UpdateStatus(ctx, rq, metav1.UpdateOptions{})).Error().NotTo(HaveOccurred())
			}).Should(Succeed())

			By("checking that b1's replicas are still not limited")
			Eventually(func(g Gomega) {
				b, err := buffersClient.AutoscalingV1beta1().CapacityBuffers(namespace).Get(ctx, "b1", metav1.GetOptions{})
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(b.Status.Replicas).To(Equal(ptr.To[int32](4)))

			}).Should(Succeed())

			By("checking that b2 used the newly available quota")
			Eventually(func(g Gomega) {
				b, err := buffersClient.AutoscalingV1beta1().CapacityBuffers(namespace).Get(ctx, "b2", metav1.GetOptions{})
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(b.Status.Replicas).To(Equal(ptr.To[int32](4)))
				g.Expect(meta.IsStatusConditionFalse(b.Status.Conditions, cbapi.LimitedByQuotasCondition)).To(BeTrue())
			}).Should(Succeed())
		})
	})
})
