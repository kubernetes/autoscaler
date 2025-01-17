/*
Copyright 2020 The Kubernetes Authors.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/status"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/test"
	"k8s.io/kubernetes/test/e2e/framework"
	podsecurity "k8s.io/pod-security-admission/api"

	ginkgo "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = UpdaterE2eDescribe("Updater", func() {
	f := framework.NewDefaultFramework("vertical-pod-autoscaling")
	f.NamespacePodSecurityEnforceLevel = podsecurity.LevelBaseline

	// TODO(jkyros): it should only evict here if the feature gate is off, so we need to
	// check behavior by making sure it aligns with the feature gate. e.g. if it's on, then do this test, if it's not, then skip it

	// 1. check if we have resize policies, if we do, do the test with in-place
	// 2. if we don't, then do it the old way

	ginkgo.It("In-place update pods when Admission Controller status available", func() {
		const statusUpdateInterval = 10 * time.Second

		ginkgo.By("Setting up the Admission Controller status")
		stopCh := make(chan struct{})
		statusUpdater := status.NewUpdater(
			f.ClientSet,
			status.AdmissionControllerStatusName,
			status.AdmissionControllerStatusNamespace,
			statusUpdateInterval,
			"e2e test",
		)
		defer func() {
			// Schedule a cleanup of the Admission Controller status.
			// Status is created outside the test namespace.
			ginkgo.By("Deleting the Admission Controller status")
			close(stopCh)
			err := f.ClientSet.CoordinationV1().Leases(status.AdmissionControllerStatusNamespace).
				Delete(context.TODO(), status.AdmissionControllerStatusName, metav1.DeleteOptions{})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		}()
		statusUpdater.Run(stopCh)

		podList := setupPodsForUpscalingInPlace(f)
		if len(podList.Items[0].Spec.Containers[0].ResizePolicy) <= 0 {
			// Feature is probably not working here
			ginkgo.Skip("Skipping test, InPlacePodVerticalScaling not available")
		}

		initialPods := podList.DeepCopy()
		// 1. Take initial pod list
		// 2. Loop through and compare all the resource values
		// 3. When they change, it's good

		ginkgo.By("Waiting for pods to be in-place updated")

		//gomega.Expect(err).NotTo(gomega.HaveOccurred())
		err := WaitForPodsUpdatedWithoutEviction(f, initialPods, podList)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	})

	ginkgo.It("Does not evict pods for downscaling in-place", func() {
		const statusUpdateInterval = 10 * time.Second

		ginkgo.By("Setting up the Admission Controller status")
		stopCh := make(chan struct{})
		statusUpdater := status.NewUpdater(
			f.ClientSet,
			status.AdmissionControllerStatusName,
			status.AdmissionControllerStatusNamespace,
			statusUpdateInterval,
			"e2e test",
		)
		defer func() {
			// Schedule a cleanup of the Admission Controller status.
			// Status is created outside the test namespace.
			ginkgo.By("Deleting the Admission Controller status")
			close(stopCh)
			err := f.ClientSet.CoordinationV1().Leases(status.AdmissionControllerStatusNamespace).
				Delete(context.TODO(), status.AdmissionControllerStatusName, metav1.DeleteOptions{})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		}()
		statusUpdater.Run(stopCh)

		podList := setupPodsForDownscalingInPlace(f, nil)
		if len(podList.Items[0].Spec.Containers[0].ResizePolicy) <= 0 {
			// Feature is probably not working here
			ginkgo.Skip("Skipping test, InPlacePodVerticalScaling not available")
		}
		initialPods := podList.DeepCopy()

		ginkgo.By("Waiting for pods to be in-place downscaled")
		err := WaitForPodsUpdatedWithoutEviction(f, initialPods, podList)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	})

	ginkgo.It("evicts pods when Admission Controller status available", func() {
		const statusUpdateInterval = 10 * time.Second

		ginkgo.By("Setting up the Admission Controller status")
		stopCh := make(chan struct{})
		statusUpdater := status.NewUpdater(
			f.ClientSet,
			status.AdmissionControllerStatusName,
			status.AdmissionControllerStatusNamespace,
			statusUpdateInterval,
			"e2e test",
		)
		defer func() {
			// Schedule a cleanup of the Admission Controller status.
			// Status is created outside the test namespace.
			ginkgo.By("Deleting the Admission Controller status")
			close(stopCh)
			err := f.ClientSet.CoordinationV1().Leases(status.AdmissionControllerStatusNamespace).
				Delete(context.TODO(), status.AdmissionControllerStatusName, metav1.DeleteOptions{})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		}()
		statusUpdater.Run(stopCh)

		podList := setupPodsForUpscalingEviction(f)

		ginkgo.By("Waiting for pods to be evicted")
		err := WaitForPodsEvicted(f, podList)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	})

	ginkgo.It("evicts pods for downscaling", func() {
		const statusUpdateInterval = 10 * time.Second

		ginkgo.By("Setting up the Admission Controller status")
		stopCh := make(chan struct{})
		statusUpdater := status.NewUpdater(
			f.ClientSet,
			status.AdmissionControllerStatusName,
			status.AdmissionControllerStatusNamespace,
			statusUpdateInterval,
			"e2e test",
		)
		defer func() {
			// Schedule a cleanup of the Admission Controller status.
			// Status is created outside the test namespace.
			ginkgo.By("Deleting the Admission Controller status")
			close(stopCh)
			err := f.ClientSet.CoordinationV1().Leases(status.AdmissionControllerStatusNamespace).
				Delete(context.TODO(), status.AdmissionControllerStatusName, metav1.DeleteOptions{})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		}()
		statusUpdater.Run(stopCh)

		podList := setupPodsForDownscalingEviction(f, nil)

		ginkgo.By("Waiting for pods to be evicted")
		err := WaitForPodsEvicted(f, podList)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	})

	ginkgo.It("does not evict pods for downscaling when EvictionRequirement prevents it", func() {
		const statusUpdateInterval = 10 * time.Second

		ginkgo.By("Setting up the Admission Controller status")
		stopCh := make(chan struct{})
		statusUpdater := status.NewUpdater(
			f.ClientSet,
			status.AdmissionControllerStatusName,
			status.AdmissionControllerStatusNamespace,
			statusUpdateInterval,
			"e2e test",
		)
		defer func() {
			// Schedule a cleanup of the Admission Controller status.
			// Status is created outside the test namespace.
			ginkgo.By("Deleting the Admission Controller status")
			close(stopCh)
			err := f.ClientSet.CoordinationV1().Leases(status.AdmissionControllerStatusNamespace).
				Delete(context.TODO(), status.AdmissionControllerStatusName, metav1.DeleteOptions{})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		}()
		statusUpdater.Run(stopCh)
		er := []*vpa_types.EvictionRequirement{
			{
				Resources:         []apiv1.ResourceName{apiv1.ResourceCPU},
				ChangeRequirement: vpa_types.TargetHigherThanRequests,
			},
		}
		podList := setupPodsForDownscalingEviction(f, er)

		ginkgo.By(fmt.Sprintf("Waiting for pods to be evicted, hoping it won't happen, sleep for %s", VpaEvictionTimeout.String()))
		CheckNoPodsEvicted(f, MakePodSet(podList))
	})

	ginkgo.It("doesn't evict pods when Admission Controller status unavailable", func() {
		podList := setupPodsForUpscalingEviction(f)

		ginkgo.By(fmt.Sprintf("Waiting for pods to be evicted, hoping it won't happen, sleep for %s", VpaEvictionTimeout.String()))
		CheckNoPodsEvicted(f, MakePodSet(podList))
	})
})

func setupPodsForUpscalingEviction(f *framework.Framework) *apiv1.PodList {
	return setupPodsForEviction(f, "100m", "100Mi", nil)
}

func setupPodsForDownscalingEviction(f *framework.Framework, er []*vpa_types.EvictionRequirement) *apiv1.PodList {
	return setupPodsForEviction(f, "500m", "500Mi", er)
}

func setupPodsForEviction(f *framework.Framework, hamsterCPU, hamsterMemory string, er []*vpa_types.EvictionRequirement) *apiv1.PodList {
	controller := &autoscaling.CrossVersionObjectReference{
		APIVersion: "apps/v1",
		Kind:       "Deployment",
		Name:       "hamster-deployment",
	}
	ginkgo.By(fmt.Sprintf("Setting up a hamster %v", controller.Kind))
	setupHamsterController(f, controller.Kind, hamsterCPU, hamsterMemory, defaultHamsterReplicas)
	podList, err := GetHamsterPods(f)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	ginkgo.By("Setting up a VPA CRD")
	containerName := GetHamsterContainerNameByIndex(0)
	vpaCRD := test.VerticalPodAutoscaler().
		WithName("hamster-vpa").
		WithNamespace(f.Namespace.Name).
		WithTargetRef(controller).
		WithUpdateMode(vpa_types.UpdateModeRecreate).
		WithEvictionRequirements(er).
		WithContainer(containerName).
		AppendRecommendation(
			test.Recommendation().
				WithContainer(containerName).
				WithTarget(containerName, "200m").
				WithLowerBound(containerName, "200m").
				WithUpperBound(containerName, "200m").
				GetContainerResources()).
		Get()

	InstallVPA(f, vpaCRD)

	return podList
}

func setupPodsForUpscalingInPlace(f *framework.Framework) *apiv1.PodList {
	return setupPodsForInPlace(f, "100m", "100Mi", nil)
}

func setupPodsForDownscalingInPlace(f *framework.Framework, er []*vpa_types.EvictionRequirement) *apiv1.PodList {
	return setupPodsForInPlace(f, "500m", "500Mi", er)
}

func setupPodsForInPlace(f *framework.Framework, hamsterCPU, hamsterMemory string, er []*vpa_types.EvictionRequirement) *apiv1.PodList {
	controller := &autoscaling.CrossVersionObjectReference{
		APIVersion: "apps/v1",
		Kind:       "Deployment",
		Name:       "hamster-deployment",
	}
	ginkgo.By(fmt.Sprintf("Setting up a hamster %v", controller.Kind))
	setupHamsterController(f, controller.Kind, hamsterCPU, hamsterMemory, defaultHamsterReplicas)
	podList, err := GetHamsterPods(f)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	ginkgo.By("Setting up a VPA CRD")
	containerName := GetHamsterContainerNameByIndex(0)
	vpaCRD := test.VerticalPodAutoscaler().
		WithName("hamster-vpa").
		WithNamespace(f.Namespace.Name).
		WithTargetRef(controller).
		WithUpdateMode(vpa_types.UpdateModeInPlaceOrRecreate).
		WithEvictionRequirements(er).
		WithContainer(containerName).
		AppendRecommendation(
			test.Recommendation().
				WithContainer(containerName).
				WithTarget(containerName, "200m").
				WithLowerBound(containerName, "200m").
				WithUpperBound(containerName, "200m").
				GetContainerResources()).
		Get()

	InstallVPA(f, vpaCRD)

	return podList
}
