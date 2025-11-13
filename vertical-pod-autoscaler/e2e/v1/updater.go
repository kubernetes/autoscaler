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
	"k8s.io/autoscaler/vertical-pod-autoscaler/e2e/utils"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/features"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/annotations"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/status"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/test"
	"k8s.io/kubernetes/test/e2e/framework"
	podsecurity "k8s.io/pod-security-admission/api"

	ginkgo "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = UpdaterE2eDescribe("Updater", func() {
	f := framework.NewDefaultFramework("vertical-pod-autoscaling")
	f.NamespacePodSecurityLevel = podsecurity.LevelBaseline

	// Sets up a lease object updated periodically to signal - requires WithSerial()
	framework.It("evicts pods when Admission Controller status available", framework.WithSerial(), func() {
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

	// Sets up a lease object updated periodically to signal - requires WithSerial()
	framework.It("evicts pods for downscaling", framework.WithSerial(), func() {
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
	// Sets up a lease object updated periodically to signal - requires WithSerial()
	framework.It("does not evict pods for downscaling when EvictionRequirement prevents it", framework.WithSerial(), func() {
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
	// FIXME todo(adrianmoisey): This test seems to be flaky after running in parallel, unsure why, see if it's possible to fix
	framework.It("doesn't evict pods when Admission Controller status unavailable", framework.WithSerial(), func() {
		podList := setupPodsForUpscalingEviction(f)

		ginkgo.By(fmt.Sprintf("Waiting for pods to be evicted, hoping it won't happen, sleep for %s", VpaEvictionTimeout.String()))
		CheckNoPodsEvicted(f, MakePodSet(podList))
	})

	// Sets up a lease object updated periodically to signal - requires WithSerial()
	framework.It("In-place update pods when Admission Controller status available", framework.WithSerial(), func() {
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
		initialPods := podList.DeepCopy()

		ginkgo.By("Waiting for pods to be in-place updated")
		err := WaitForPodsUpdatedWithoutEviction(f, initialPods)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	})

	// Sets up a lease object updated periodically to signal - requires WithSerial()
	framework.It("Does not evict pods for downscaling in-place", framework.WithSerial(), func() {
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
		initialPods := podList.DeepCopy()

		ginkgo.By("Waiting for pods to be in-place downscaled")
		err := WaitForPodsUpdatedWithoutEviction(f, initialPods)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	})
})

var _ = UpdaterE2eDescribe("Updater", func() {
	f := framework.NewDefaultFramework("vertical-pod-autoscaling")
	f.NamespacePodSecurityEnforceLevel = podsecurity.LevelBaseline

	f.It("Unboost pods when they become Ready", framework.WithFeatureGate(features.CPUStartupBoost), func() {
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

		podList := setupPodsForCPUBoost(f, "100m", "100Mi")
		initialPods := podList.DeepCopy()

		ginkgo.By("Waiting for pods to be in-place updated")
		err := WaitForPodsUpdatedWithoutEviction(f, initialPods)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	})

})

func setupPodsForCPUBoost(f *framework.Framework, hamsterCPU, hamsterMemory string) *apiv1.PodList {
	controller := &autoscaling.CrossVersionObjectReference{
		APIVersion: "apps/v1",
		Kind:       "Deployment",
		Name:       "hamster-deployment",
	}
	ginkgo.By(fmt.Sprintf("Setting up a hamster %v", controller.Kind))
	// Create pods with boosted CPU, which is 2x the target recommendation
	boostedCPU := "200m"
	setupHamsterController(f, controller.Kind, boostedCPU, hamsterMemory, utils.DefaultHamsterReplicas)
	podList, err := GetHamsterPods(f)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	ginkgo.By("Setting up a VPA CRD")
	containerName := utils.GetHamsterContainerNameByIndex(0)
	factor := int32(2)
	vpaCRD := test.VerticalPodAutoscaler().
		WithName("hamster-vpa").
		WithNamespace(f.Namespace.Name).
		WithTargetRef(controller).
		WithUpdateMode(vpa_types.UpdateModeAuto).
		WithContainer(containerName).
		WithCPUStartupBoost(vpa_types.FactorStartupBoostType, &factor, nil, "1s").
		AppendRecommendation(
			test.Recommendation().
				WithContainer(containerName).
				WithTarget(hamsterCPU, hamsterMemory).
				GetContainerResources(),
		).
		Get()

	utils.InstallVPA(f, vpaCRD)

	ginkgo.By("Annotating pods with boost annotation")
	for _, pod := range podList.Items {
		original, err := annotations.GetOriginalResourcesAnnotationValue(&pod.Spec.Containers[0])
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		AnnotatePod(f, pod.Name, annotations.StartupCPUBoostAnnotation, original)
	}
	return podList
}

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
	setupHamsterController(f, controller.Kind, hamsterCPU, hamsterMemory, utils.DefaultHamsterReplicas)
	podList, err := GetHamsterPods(f)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	ginkgo.By("Setting up a VPA CRD")
	containerName := utils.GetHamsterContainerNameByIndex(0)
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

	utils.InstallVPA(f, vpaCRD)

	return podList
}

func setupPodsForUpscalingInPlace(f *framework.Framework) *apiv1.PodList {
	return setupPodsForInPlace(f, "100m", "100Mi", nil, true)
}

func setupPodsForDownscalingInPlace(f *framework.Framework, er []*vpa_types.EvictionRequirement) *apiv1.PodList {
	return setupPodsForInPlace(f, "500m", "500Mi", er, true)
}

func setupPodsForInPlace(f *framework.Framework, hamsterCPU, hamsterMemory string, er []*vpa_types.EvictionRequirement, withRecommendation bool) *apiv1.PodList {
	controller := &autoscaling.CrossVersionObjectReference{
		APIVersion: "apps/v1",
		Kind:       "Deployment",
		Name:       "hamster-deployment",
	}
	ginkgo.By(fmt.Sprintf("Setting up a hamster %v", controller.Kind))
	setupHamsterController(f, controller.Kind, hamsterCPU, hamsterMemory, utils.DefaultHamsterReplicas)
	podList, err := GetHamsterPods(f)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	ginkgo.By("Setting up a VPA CRD")
	containerName := utils.GetHamsterContainerNameByIndex(0)
	vpaBuilder := test.VerticalPodAutoscaler().
		WithName("hamster-vpa").
		WithNamespace(f.Namespace.Name).
		WithTargetRef(controller).
		WithUpdateMode(vpa_types.UpdateModeInPlaceOrRecreate).
		WithEvictionRequirements(er).
		WithContainer(containerName)

	if withRecommendation {
		vpaBuilder = vpaBuilder.AppendRecommendation(
			test.Recommendation().
				WithContainer(containerName).
				WithTarget(containerName, "200m").
				WithLowerBound(containerName, "200m").
				WithUpperBound(containerName, "200m").
				GetContainerResources())
	}

	vpaCRD := vpaBuilder.Get()
	utils.InstallVPA(f, vpaCRD)

	return podList
}
