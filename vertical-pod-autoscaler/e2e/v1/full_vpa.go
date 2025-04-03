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
	"context"
	"fmt"
	"time"

	autoscaling "k8s.io/api/autoscaling/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/test"
	"k8s.io/kubernetes/test/e2e/framework"
	podsecurity "k8s.io/pod-security-admission/api"

	ginkgo "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

const (
	minimalCPULowerBound    = "0m"
	minimalCPUUpperBound    = "100m"
	minimalMemoryLowerBound = "0Mi"
	minimalMemoryUpperBound = "300Mi"
	// the initial values should be outside minimal bounds
	initialCPU     = int64(10) // mCPU
	initialMemory  = int64(10) // MB
	oomTestTimeout = 8 * time.Minute
)

var _ = FullVpaE2eDescribe("Pods under VPA", func() {
	var (
		rc *ResourceConsumer
	)
	replicas := 3

	ginkgo.AfterEach(func() {
		rc.CleanUp()
	})

	// This schedules AfterEach block that needs to run after the AfterEach above and
	// BeforeEach that needs to run before the BeforeEach below - thus the order of these matters.
	f := framework.NewDefaultFramework("vertical-pod-autoscaling")
	f.NamespacePodSecurityEnforceLevel = podsecurity.LevelBaseline

	ginkgo.BeforeEach(func() {
		ns := f.Namespace.Name
		ginkgo.By("Setting up a hamster deployment")
		rc = NewDynamicResourceConsumer("hamster", ns, KindDeployment,
			replicas,
			1,             /*initCPUTotal*/
			10,            /*initMemoryTotal*/
			1,             /*initCustomMetric*/
			initialCPU,    /*cpuRequest*/
			initialMemory, /*memRequest*/
			f.ClientSet,
			f.ScalesGetter)

		ginkgo.By("Setting up a VPA CRD")
		targetRef := &autoscaling.CrossVersionObjectReference{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
			Name:       "hamster",
		}

		containerName := GetHamsterContainerNameByIndex(0)
		vpaCRD := test.VerticalPodAutoscaler().
			WithName("hamster-vpa").
			WithNamespace(f.Namespace.Name).
			WithTargetRef(targetRef).
			WithContainer(containerName).
			AppendRecommendation(
				test.Recommendation().
					WithContainer(containerName).
					WithTarget("250m", "200Mi").
					WithLowerBound("250m", "200Mi").
					WithUpperBound("250m", "200Mi").
					GetContainerResources()).
			Get()

		InstallVPA(f, vpaCRD)
	})

	ginkgo.It("have cpu requests growing with usage", func() {
		// initial CPU usage is low so a minimal recommendation is expected
		err := waitForResourceRequestInRangeInPods(
			f, pollTimeout, metav1.ListOptions{LabelSelector: "name=hamster"}, apiv1.ResourceCPU,
			ParseQuantityOrDie(minimalCPULowerBound), ParseQuantityOrDie(minimalCPUUpperBound))
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		// consume more CPU to get a higher recommendation
		rc.ConsumeCPU(600 * replicas)
		err = waitForResourceRequestInRangeInPods(
			f, pollTimeout, metav1.ListOptions{LabelSelector: "name=hamster"}, apiv1.ResourceCPU,
			ParseQuantityOrDie("500m"), ParseQuantityOrDie("1300m"))
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	})

	ginkgo.It("have memory requests growing with usage", func() {
		// initial memory usage is low so a minimal recommendation is expected
		err := waitForResourceRequestInRangeInPods(
			f, pollTimeout, metav1.ListOptions{LabelSelector: "name=hamster"}, apiv1.ResourceMemory,
			ParseQuantityOrDie(minimalMemoryLowerBound), ParseQuantityOrDie(minimalMemoryUpperBound))
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		// consume more memory to get a higher recommendation
		// NOTE: large range given due to unpredictability of actual memory usage
		rc.ConsumeMem(1024 * replicas)
		err = waitForResourceRequestInRangeInPods(
			f, pollTimeout, metav1.ListOptions{LabelSelector: "name=hamster"}, apiv1.ResourceMemory,
			ParseQuantityOrDie("900Mi"), ParseQuantityOrDie("4000Mi"))
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	})
})

var _ = FullVpaE2eDescribe("Pods under VPA with default recommender explicitly configured", func() {
	var (
		rc *ResourceConsumer
	)
	replicas := 3

	ginkgo.AfterEach(func() {
		rc.CleanUp()
	})

	// This schedules AfterEach block that needs to run after the AfterEach above and
	// BeforeEach that needs to run before the BeforeEach below - thus the order of these matters.
	f := framework.NewDefaultFramework("vertical-pod-autoscaling")
	f.NamespacePodSecurityEnforceLevel = podsecurity.LevelBaseline

	ginkgo.BeforeEach(func() {
		ns := f.Namespace.Name
		ginkgo.By("Setting up a hamster deployment")
		rc = NewDynamicResourceConsumer("hamster", ns, KindDeployment,
			replicas,
			1,             /*initCPUTotal*/
			10,            /*initMemoryTotal*/
			1,             /*initCustomMetric*/
			initialCPU,    /*cpuRequest*/
			initialMemory, /*memRequest*/
			f.ClientSet,
			f.ScalesGetter)

		ginkgo.By("Setting up a VPA CRD with Recommender explicitly configured")
		targetRef := &autoscaling.CrossVersionObjectReference{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
			Name:       "hamster",
		}

		containerName := GetHamsterContainerNameByIndex(0)
		vpaCRD := test.VerticalPodAutoscaler().
			WithName("hamster-vpa").
			WithNamespace(f.Namespace.Name).
			WithTargetRef(targetRef).
			WithContainer(containerName).
			AppendRecommendation(
				test.Recommendation().
					WithContainer(containerName).
					WithTarget("250m", "200Mi").
					WithLowerBound("250m", "200Mi").
					WithUpperBound("250m", "200Mi").
					GetContainerResources()).
			Get()

		InstallVPA(f, vpaCRD)

	})

	ginkgo.It("have cpu requests growing with usage", func() {
		// initial CPU usage is low so a minimal recommendation is expected
		err := waitForResourceRequestInRangeInPods(
			f, pollTimeout, metav1.ListOptions{LabelSelector: "name=hamster"}, apiv1.ResourceCPU,
			ParseQuantityOrDie(minimalCPULowerBound), ParseQuantityOrDie(minimalCPUUpperBound))
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		// consume more CPU to get a higher recommendation
		rc.ConsumeCPU(600 * replicas)
		err = waitForResourceRequestInRangeInPods(
			f, pollTimeout, metav1.ListOptions{LabelSelector: "name=hamster"}, apiv1.ResourceCPU,
			ParseQuantityOrDie("500m"), ParseQuantityOrDie("1300m"))
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	})
})

var _ = FullVpaE2eDescribe("Pods under VPA with non-recognized recommender explicitly configured", func() {
	var (
		rc *ResourceConsumer
	)
	replicas := 3

	ginkgo.AfterEach(func() {
		rc.CleanUp()
	})

	// This schedules AfterEach block that needs to run after the AfterEach above and
	// BeforeEach that needs to run before the BeforeEach below - thus the order of these matters.
	f := framework.NewDefaultFramework("vertical-pod-autoscaling")
	f.NamespacePodSecurityEnforceLevel = podsecurity.LevelBaseline

	ginkgo.BeforeEach(func() {
		ns := f.Namespace.Name
		ginkgo.By("Setting up a hamster deployment")
		rc = NewDynamicResourceConsumer("hamster", ns, KindDeployment,
			replicas,
			1,             /*initCPUTotal*/
			10,            /*initMemoryTotal*/
			1,             /*initCustomMetric*/
			initialCPU,    /*cpuRequest*/
			initialMemory, /*memRequest*/
			f.ClientSet,
			f.ScalesGetter)

		ginkgo.By("Setting up a VPA CRD with Recommender explicitly configured")
		targetRef := &autoscaling.CrossVersionObjectReference{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
			Name:       "hamster",
		}

		containerName := GetHamsterContainerNameByIndex(0)
		vpaCRD := test.VerticalPodAutoscaler().
			WithName("hamster-vpa").
			WithRecommender("non-recognized").
			WithNamespace(f.Namespace.Name).
			WithTargetRef(targetRef).
			WithContainer(containerName).
			AppendRecommendation(
				test.Recommendation().
					WithContainer(containerName).
					WithTarget("250m", "200Mi").
					WithLowerBound("250m", "200Mi").
					WithUpperBound("250m", "200Mi").
					GetContainerResources()).
			Get()

		InstallVPA(f, vpaCRD)

	})

	ginkgo.It("deployment not updated by non-recognized recommender", func() {
		err := waitForResourceRequestInRangeInPods(
			f, pollTimeout, metav1.ListOptions{LabelSelector: "name=hamster"}, apiv1.ResourceCPU,
			ParseQuantityOrDie(minimalCPULowerBound), ParseQuantityOrDie(minimalCPUUpperBound))
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		// consume more CPU to get a higher recommendation
		rc.ConsumeCPU(600 * replicas)
		err = waitForResourceRequestInRangeInPods(
			f, pollTimeout, metav1.ListOptions{LabelSelector: "name=hamster"}, apiv1.ResourceCPU,
			ParseQuantityOrDie("500m"), ParseQuantityOrDie("1000m"))
		gomega.Expect(err).To(gomega.HaveOccurred())
	})
})

var _ = FullVpaE2eDescribe("OOMing pods under VPA", func() {
	const replicas = 3

	f := framework.NewDefaultFramework("vertical-pod-autoscaling")
	f.NamespacePodSecurityEnforceLevel = podsecurity.LevelBaseline

	ginkgo.BeforeEach(func() {
		ns := f.Namespace.Name
		ginkgo.By("Setting up a hamster deployment")

		runOomingReplicationController(
			f.ClientSet,
			ns,
			"hamster",
			replicas)
		ginkgo.By("Setting up a VPA CRD")
		targetRef := &autoscaling.CrossVersionObjectReference{
			APIVersion: "v1",
			Kind:       "Deployment",
			Name:       "hamster",
		}

		containerName := GetHamsterContainerNameByIndex(0)
		vpaCRD := test.VerticalPodAutoscaler().
			WithName("hamster-vpa").
			WithNamespace(f.Namespace.Name).
			WithTargetRef(targetRef).
			WithContainer(containerName).
			Get()

		InstallVPA(f, vpaCRD)
	})

	ginkgo.It("have memory requests growing with OOMs", func() {
		listOptions := metav1.ListOptions{
			LabelSelector: "name=hamster",
			FieldSelector: getPodSelectorExcludingDonePodsOrDie(),
		}
		err := waitForResourceRequestInRangeInPods(
			f, oomTestTimeout, listOptions, apiv1.ResourceMemory,
			ParseQuantityOrDie("1400Mi"), ParseQuantityOrDie("10000Mi"))
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	})
})

func waitForPodsMatch(f *framework.Framework, timeout time.Duration, listOptions metav1.ListOptions, matcher func(pod apiv1.Pod) bool) error {
	return wait.PollUntilContextTimeout(context.Background(), pollInterval, timeout, true, func(ctx context.Context) (done bool, err error) {
		ns := f.Namespace.Name
		c := f.ClientSet

		podList, err := c.CoreV1().Pods(ns).List(ctx, listOptions)
		if err != nil {
			return false, err
		}

		if len(podList.Items) == 0 {
			return false, nil
		}

		// Run matcher on all pods, even if we find pod that doesn't match early.
		// This allows the matcher to write logs for all pods. This in turns makes
		// it easier to spot some problems (for example unexpected pods in the list
		// results).
		result := true
		for _, pod := range podList.Items {
			if !matcher(pod) {
				result = false
			}
		}
		return result, nil

	})
}

func waitForResourceRequestInRangeInPods(f *framework.Framework, timeout time.Duration, listOptions metav1.ListOptions, resourceName apiv1.ResourceName, lowerBound, upperBound resource.Quantity) error {
	err := waitForPodsMatch(f, timeout, listOptions,
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
