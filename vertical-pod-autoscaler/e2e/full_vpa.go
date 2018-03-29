/*
Copyright 2018 The Kubernetes Authors.

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

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/poc.autoscaling.k8s.io/v1alpha1"
	vpa_clientset "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned"
	e2e_common "k8s.io/kubernetes/test/e2e/common"
	"k8s.io/kubernetes/test/e2e/framework"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

const (
	minimalCPU = "50m"
)

var _ = fullVpaE2eDescribe("Pods under VPA", func() {
	f := framework.NewDefaultFramework("vertical-pod-autoscaling")

	var (
		rc           *ResourceConsumer
		vpaClientSet *vpa_clientset.Clientset
		vpaCRD       *vpa_types.VerticalPodAutoscaler
	)
	replicas := 3

	ginkgo.BeforeEach(func() {
		ns := f.Namespace.Name
		ginkgo.By("Setting up a hamster deployment")
		rc = NewDynamicResourceConsumer("hamster", ns, e2e_common.KindDeployment,
			replicas,
			1,  /*initCPUTotal*/
			10, /*initMemoryTotal*/
			1,  /*initCustomMetric*/
			parseQuantityOrDie("100m"), /*cpuRequest*/
			parseQuantityOrDie("10Mi"), /*memRequest*/
			f.ClientSet,
			f.InternalClientset)

		ginkgo.By("Setting up a VPA CRD")
		config, err := framework.LoadConfig()
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		vpaCRD = newVPA(f, "hamster-vpa", &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"name": "hamster",
			},
		})

		vpaClientSet = vpa_clientset.NewForConfigOrDie(config)
		vpaClient := vpaClientSet.PocV1alpha1()
		_, err = vpaClient.VerticalPodAutoscalers(ns).Create(vpaCRD)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

	})

	ginkgo.It("stabilize at minimum CPU if doing nothing", func() {
		err := waitForResourceRequestAboveThresholdInPods(
			f, metav1.ListOptions{LabelSelector: "name=hamster"}, apiv1.ResourceCPU, parseQuantityOrDie(minimalCPU))
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	})

	ginkgo.It("have cpu requests growing with usage", func() {
		rc.ConsumeCPU(600 * replicas)
		err := waitForResourceRequestAboveThresholdInPods(
			f, metav1.ListOptions{LabelSelector: "name=hamster"}, apiv1.ResourceCPU, parseQuantityOrDie("500m"))
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	})

	ginkgo.It("have memory requests growing with OOMs", func() {
		// Wait for any recommendation
		err := waitForRecommendationPresent(vpaClientSet, vpaCRD)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		// Restart pods to have limits populated
		err = deletePods(f, metav1.ListOptions{LabelSelector: "name=hamster"})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		rc.ConsumeMem(1024 * replicas)
		err = waitForResourceRequestAboveThresholdInPods(
			f, metav1.ListOptions{LabelSelector: "name=hamster"}, apiv1.ResourceMemory, parseQuantityOrDie("600Mi"))
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

func deletePods(f *framework.Framework, listOptions metav1.ListOptions) error {
	ns := f.Namespace.Name
	c := f.ClientSet
	podList, err := c.CoreV1().Pods(ns).List(listOptions)
	if err != nil {
		return err
	}
	for _, pod := range podList.Items {
		err := c.CoreV1().Pods(ns).Delete(pod.Name, &metav1.DeleteOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}

func waitForResourceRequestAboveThresholdInPods(f *framework.Framework, listOptions metav1.ListOptions, resourceName apiv1.ResourceName, threshold resource.Quantity) error {
	err := waitForPodsMatch(f, listOptions,
		func(pod apiv1.Pod) bool {
			resourceRequest, found := pod.Spec.Containers[0].Resources.Requests[resourceName]
			framweork.Logf("Comparing %v request %v against minimum threshold of %v", resourceName, resourceRequest, threshold)
			return found && resourceRequest.MilliValue() > threshold.MilliValue()
		})

	if err != nil {
		return fmt.Errorf("error waiting for %s request above %v for pods: %+v", resourceName, threshold, listOptions)
	}
	return nil
}
