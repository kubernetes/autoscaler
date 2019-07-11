/*
Copyright 2019 The Kubernetes Authors.

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

package autoscaling

import (
	"fmt"

	autoscaling "k8s.io/api/autoscaling/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	vpa_clientset "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned"
	e2e_common "k8s.io/kubernetes/test/e2e/common"
	"k8s.io/kubernetes/test/e2e/framework"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

const (
	minimalCPULowerBound    = "20m"
	minimalCPUUpperBound    = "100m"
	minimalMemoryLowerBound = "20Mi"
	minimalMemoryUpperBound = "300Mi"
	// the initial values should be outside minimal bounds
	initialCPU    = "10m"
	initialMemory = "10Mi"
)

var _ = FullVpaE2eDescribe("Pods under VPA", func() {
	var (
		rc           *ResourceConsumer
		vpaClientSet *vpa_clientset.Clientset
		vpaCRD       *vpa_types.VerticalPodAutoscaler
	)
	replicas := 3

	ginkgo.AfterEach(func() {
		rc.CleanUp()
	})

	// This schedules AfterEach block that needs to run after the AfterEach above and
	// BeforeEach that needs to run before the BeforeEach below - thus the order of these matters.
	f := framework.NewDefaultFramework("vertical-pod-autoscaling")

	ginkgo.BeforeEach(func() {
		ns := f.Namespace.Name
		ginkgo.By("Setting up a hamster deployment")
		rc = NewDynamicResourceConsumer("hamster", ns, e2e_common.KindDeployment,
			replicas,
			1,                                 /*initCPUTotal*/
			10,                                /*initMemoryTotal*/
			1,                                 /*initCustomMetric*/
			ParseQuantityOrDie(initialCPU),    /*cpuRequest*/
			ParseQuantityOrDie(initialMemory), /*memRequest*/
			f.ClientSet,
			f.InternalClientset)

		ginkgo.By("Setting up a VPA CRD")
		config, err := framework.LoadConfig()
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		vpaCRD = NewVPA(f, "hamster-vpa", &autoscaling.CrossVersionObjectReference{
			APIVersion: "extensions/v1beta1",
			Kind:       "Deployment",
			Name:       "hamster",
		})

		vpaClientSet = vpa_clientset.NewForConfigOrDie(config)
		vpaClient := vpaClientSet.AutoscalingV1()
		_, err = vpaClient.VerticalPodAutoscalers(ns).Create(vpaCRD)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

	})

	ginkgo.It("have cpu requests growing with usage", func() {
		// initial CPU usage is low so a minimal recommendation is expected
		err := waitForResourceRequestInRangeInPods(
			f, metav1.ListOptions{LabelSelector: "name=hamster"}, apiv1.ResourceCPU,
			ParseQuantityOrDie(minimalCPULowerBound), ParseQuantityOrDie(minimalCPUUpperBound))
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		// consume more CPU to get a higher recommendation
		rc.ConsumeCPU(600 * replicas)
		err = waitForResourceRequestInRangeInPods(
			f, metav1.ListOptions{LabelSelector: "name=hamster"}, apiv1.ResourceCPU,
			ParseQuantityOrDie("500m"), ParseQuantityOrDie("900m"))
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	})

	ginkgo.It("have memory requests growing with usage", func() {
		// initial memory usage is low so a minimal recommendation is expected
		err := waitForResourceRequestInRangeInPods(
			f, metav1.ListOptions{LabelSelector: "name=hamster"}, apiv1.ResourceMemory,
			ParseQuantityOrDie(minimalMemoryLowerBound), ParseQuantityOrDie(minimalMemoryUpperBound))
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		// consume more memory to get a higher recommendation
		// NOTE: large range given due to unpredictability of actual memory usage
		rc.ConsumeMem(1024 * replicas)
		err = waitForResourceRequestInRangeInPods(
			f, metav1.ListOptions{LabelSelector: "name=hamster"}, apiv1.ResourceMemory,
			ParseQuantityOrDie("900Mi"), ParseQuantityOrDie("4000Mi"))
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	})
})

func waitForPodsMatch(f *framework.Framework, listOptions metav1.ListOptions, matcher func(pod apiv1.Pod) bool) error {
	return wait.PollImmediate(pollInterval, pollTimeout, func() (bool, error) {

		ns := f.Namespace.Name
		c := f.ClientSet

		podList, err := c.CoreV1().Pods(ns).List(listOptions)
		if err != nil {
			return false, err
		}

		if len(podList.Items) == 0 {
			return false, nil
		}

		for _, pod := range podList.Items {
			if !matcher(pod) {
				return false, nil
			}
		}
		return true, nil

	})
}

func waitForResourceRequestInRangeInPods(f *framework.Framework, listOptions metav1.ListOptions, resourceName apiv1.ResourceName, lowerBound, upperBound resource.Quantity) error {
	err := waitForPodsMatch(f, listOptions,
		func(pod apiv1.Pod) bool {
			resourceRequest, found := pod.Spec.Containers[0].Resources.Requests[resourceName]
			framework.Logf("Comparing %v request %v against range of (%v, %v)", resourceName, resourceRequest, lowerBound, upperBound)
			return found && resourceRequest.MilliValue() > lowerBound.MilliValue() && resourceRequest.MilliValue() < upperBound.MilliValue()
		})

	if err != nil {
		return fmt.Errorf("error waiting for %s request in range of (%v,%v) for pods: %+v", resourceName, lowerBound, upperBound, listOptions)
	}
	return nil
}
