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
	"k8s.io/apimachinery/pkg/api/resource"
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
			utils.VpaNamespace,
			statusUpdateInterval,
			"e2e test",
		)
		defer func() {
			// Schedule a cleanup of the Admission Controller status.
			// Status is created outside the test namespace.
			ginkgo.By("Deleting the Admission Controller status")
			close(stopCh)
			err := f.ClientSet.CoordinationV1().Leases(utils.VpaNamespace).
				Delete(context.TODO(), status.AdmissionControllerStatusName, metav1.DeleteOptions{})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		}()
		statusUpdater.Run(stopCh)

		podList := setupPodsForUpscalingEviction(f, vpa_types.UpdateModeRecreate)

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
			utils.VpaNamespace,
			statusUpdateInterval,
			"e2e test",
		)
		defer func() {
			// Schedule a cleanup of the Admission Controller status.
			// Status is created outside the test namespace.
			ginkgo.By("Deleting the Admission Controller status")
			close(stopCh)
			err := f.ClientSet.CoordinationV1().Leases(utils.VpaNamespace).
				Delete(context.TODO(), status.AdmissionControllerStatusName, metav1.DeleteOptions{})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		}()
		statusUpdater.Run(stopCh)

		podList := setupPodsForDownscalingEviction(f, nil, vpa_types.UpdateModeRecreate)

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
			utils.VpaNamespace,
			statusUpdateInterval,
			"e2e test",
		)
		defer func() {
			// Schedule a cleanup of the Admission Controller status.
			// Status is created outside the test namespace.
			ginkgo.By("Deleting the Admission Controller status")
			close(stopCh)
			err := f.ClientSet.CoordinationV1().Leases(utils.VpaNamespace).
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
		podList := setupPodsForDownscalingEviction(f, er, vpa_types.UpdateModeRecreate)

		ginkgo.By(fmt.Sprintf("Waiting for pods to be evicted, hoping it won't happen, sleep for %s", VpaEvictionTimeout.String()))
		CheckNoPodsEvicted(f, MakePodSet(podList))
	})
	// FIXME todo(adrianmoisey): This test seems to be flaky after running in parallel, unsure why, see if it's possible to fix
	framework.It("doesn't evict pods when Admission Controller status unavailable", framework.WithSerial(), func() {
		podList := setupPodsForUpscalingEviction(f, vpa_types.UpdateModeInPlaceOrRecreate)

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
			utils.VpaNamespace,
			statusUpdateInterval,
			"e2e test",
		)
		defer func() {
			// Schedule a cleanup of the Admission Controller status.
			// Status is created outside the test namespace.
			ginkgo.By("Deleting the Admission Controller status")
			close(stopCh)
			err := f.ClientSet.CoordinationV1().Leases(utils.VpaNamespace).
				Delete(context.TODO(), status.AdmissionControllerStatusName, metav1.DeleteOptions{})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		}()
		statusUpdater.Run(stopCh)

		podList := setupPodsForUpscalingInPlace(f, vpa_types.UpdateModeInPlaceOrRecreate)
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
			utils.VpaNamespace,
			statusUpdateInterval,
			"e2e test",
		)
		defer func() {
			// Schedule a cleanup of the Admission Controller status.
			// Status is created outside the test namespace.
			ginkgo.By("Deleting the Admission Controller status")
			close(stopCh)
			err := f.ClientSet.CoordinationV1().Leases(utils.VpaNamespace).
				Delete(context.TODO(), status.AdmissionControllerStatusName, metav1.DeleteOptions{})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		}()
		statusUpdater.Run(stopCh)

		podList := setupPodsForDownscalingInPlace(f, nil, vpa_types.UpdateModeInPlaceOrRecreate)
		initialPods := podList.DeepCopy()

		ginkgo.By("Waiting for pods to be in-place downscaled")
		err := WaitForPodsUpdatedWithoutEviction(f, initialPods)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	})

	// Sets up a lease object updated periodically to signal - requires WithSerial()
	framework.It("In-place updates pods with InPlace mode when update succeeds", framework.WithSerial(), framework.WithFeatureGate(features.InPlace), func() {
		const statusUpdateInterval = 10 * time.Second

		ginkgo.By("Setting up the Admission Controller status")
		stopCh := make(chan struct{})
		statusUpdater := status.NewUpdater(
			f.ClientSet,
			status.AdmissionControllerStatusName,
			utils.VpaNamespace,
			statusUpdateInterval,
			"e2e test",
		)
		defer func() {
			// Schedule a cleanup of the Admission Controller status.
			// Status is created outside the test namespace.
			ginkgo.By("Deleting the Admission Controller status")
			close(stopCh)
			err := f.ClientSet.CoordinationV1().Leases(utils.VpaNamespace).
				Delete(context.TODO(), status.AdmissionControllerStatusName, metav1.DeleteOptions{})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		}()
		statusUpdater.Run(stopCh)

		podList := setupPodsForUpscalingInPlace(f, vpa_types.UpdateModeInPlace)
		initialPods := podList.DeepCopy()

		ginkgo.By("Waiting for pods to be in-place updated with InPlace mode")
		err := WaitForPodsUpdatedWithoutEviction(f, initialPods)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	})

	framework.It("does not evicts pods for downscaling with InPlace mode", framework.WithSerial(), func() {
		const statusUpdateInterval = 10 * time.Second

		ginkgo.By("Setting up the Admission Controller status")
		stopCh := make(chan struct{})
		statusUpdater := status.NewUpdater(
			f.ClientSet,
			status.AdmissionControllerStatusName,
			utils.VpaNamespace,
			statusUpdateInterval,
			"e2e test",
		)
		defer func() {
			// Schedule a cleanup of the Admission Controller status.
			// Status is created outside the test namespace.
			ginkgo.By("Deleting the Admission Controller status")
			close(stopCh)
			err := f.ClientSet.CoordinationV1().Leases(utils.VpaNamespace).
				Delete(context.TODO(), status.AdmissionControllerStatusName, metav1.DeleteOptions{})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		}()
		statusUpdater.Run(stopCh)

		podList := setupPodsForDownscalingEviction(f, nil, vpa_types.UpdateModeInPlace)

		ginkgo.By(fmt.Sprintf("Waiting for pods to be evicted, hoping it won't happen, sleep for %s", VpaEvictionTimeout.String()))
		CheckNoPodsEvicted(f, MakePodSet(podList))
	})
	framework.It("InPlace mode retries when recommendations change", framework.WithSerial(), framework.WithFeatureGate(features.InPlace), func() {
		const statusUpdateInterval = 10 * time.Second

		ginkgo.By("Setting up the Admission Controller status")
		stopCh := make(chan struct{})
		statusUpdater := status.NewUpdater(
			f.ClientSet,
			status.AdmissionControllerStatusName,
			utils.VpaNamespace,
			statusUpdateInterval,
			"e2e test",
		)
		defer func() {
			ginkgo.By("Deleting the Admission Controller status")
			close(stopCh)
			err := f.ClientSet.CoordinationV1().Leases(utils.VpaNamespace).
				Delete(context.TODO(), status.AdmissionControllerStatusName, metav1.DeleteOptions{})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		}()
		statusUpdater.Run(stopCh)

		// Set up pods with initial recommendation (100m -> 200m)
		podList := setupPodsForUpscalingInPlace(f, vpa_types.UpdateModeInPlace)
		initialPodSet := MakePodSet(podList)
		initialPods := podList.DeepCopy()

		ginkgo.By("Waiting for initial in-place update")
		err := WaitForPodsUpdatedWithoutEviction(f, initialPods)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		ginkgo.By("Recording current pod state after first update")
		podListAfterFirstUpdate, err := GetHamsterPods(f)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		podsAfterFirstUpdate := podListAfterFirstUpdate.DeepCopy()

		ginkgo.By("Updating VPA with new recommendations (300m)")
		containerName := utils.GetHamsterContainerNameByIndex(0)
		vpaClientSet := utils.GetVpaClientSet(f)

		vpaCRD, err := vpaClientSet.AutoscalingV1().VerticalPodAutoscalers(f.Namespace.Name).
			Get(context.TODO(), "hamster-vpa", metav1.GetOptions{})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		vpaCRD.Status.Recommendation = &vpa_types.RecommendedPodResources{
			ContainerRecommendations: []vpa_types.RecommendedContainerResources{
				{
					ContainerName: containerName,
					Target: apiv1.ResourceList{
						apiv1.ResourceCPU:    resource.MustParse("300m"),
						apiv1.ResourceMemory: resource.MustParse("200Mi"),
					},
					LowerBound: apiv1.ResourceList{
						apiv1.ResourceCPU:    resource.MustParse("300m"),
						apiv1.ResourceMemory: resource.MustParse("200Mi"),
					},
					UpperBound: apiv1.ResourceList{
						apiv1.ResourceCPU:    resource.MustParse("300m"),
						apiv1.ResourceMemory: resource.MustParse("200Mi"),
					},
				},
			},
		}

		_, err = vpaClientSet.AutoscalingV1().VerticalPodAutoscalers(f.Namespace.Name).
			UpdateStatus(context.TODO(), vpaCRD, metav1.UpdateOptions{})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		ginkgo.By("Waiting for pods to be updated again with new recommendations")
		err = WaitForPodsUpdatedWithoutEviction(f, podsAfterFirstUpdate)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		ginkgo.By("Verifying no pods were evicted during the process")
		currentPods, err := GetHamsterPods(f)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		currentPodSet := MakePodSet(currentPods)

		// Verify no pods were evicted by checking UIDs remain the same
		evictedCount := GetEvictedPodsCount(currentPodSet, initialPodSet)
		gomega.Expect(evictedCount).To(gomega.Equal(0),
			"No pods should be evicted when using InPlace mode")
	})
})

func setupPodsForUpscalingEviction(f *framework.Framework, updateMode vpa_types.UpdateMode) *apiv1.PodList {
	return setupPodsForEviction(f, "100m", "100Mi", nil, updateMode)
}

var _ = UpdaterE2eDescribe("Updater with PerVPAConfig", func() {
	const replicas = 3
	const statusUpdateInterval = 10 * time.Second
	f := framework.NewDefaultFramework("vertical-pod-autoscaling")
	f.NamespacePodSecurityEnforceLevel = podsecurity.LevelBaseline

	f.It("does not evict pods when OOM duration exceeds threshold", framework.WithFeatureGate(features.PerVPAConfig), func() {
		ginkgo.By("Setting up the Admission Controller status")
		stopCh := make(chan struct{})
		statusUpdater := status.NewUpdater(
			f.ClientSet,
			status.AdmissionControllerStatusName,
			utils.VpaNamespace,
			statusUpdateInterval,
			"e2e test",
		)
		defer func() {
			ginkgo.By("Deleting the Admission Controller status")
			close(stopCh)
			err := f.ClientSet.CoordinationV1().Leases(utils.VpaNamespace).
				Delete(context.TODO(), status.AdmissionControllerStatusName, metav1.DeleteOptions{})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		}()
		statusUpdater.Run(stopCh)

		// Use a very short threshold (1 second)
		// Pods that take longer than 1 second to OOM will NOT be considered "quick OOM"
		threshold := int32(1)

		ginkgo.By("Setting up VPA with very short evictAfterOOMSeconds (1 second)")
		containerName := utils.GetHamsterContainerNameByIndex(0)
		vpaCRD := newOOMTestVPA(f.Namespace.Name, "hamster-vpa", "hamster", containerName, threshold)
		utils.InstallVPA(f, vpaCRD)

		ginkgo.By("Creating deployment that will OOM (takes >1 second to OOM)")
		runOomingReplicationController(f.ClientSet, f.Namespace.Name, "hamster", replicas)

		ginkgo.By("Waiting for pods to OOM and restart")
		gomega.Eventually(func() bool {
			podList, err := GetOOMPods(f)
			if err != nil {
				return false
			}
			for _, pod := range podList.Items {
				for _, cs := range pod.Status.ContainerStatuses {
					if cs.LastTerminationState.Terminated != nil &&
						cs.LastTerminationState.Terminated.Reason == "OOMKilled" {
						framework.Logf("Pod %s container %s was OOMKilled", pod.Name, cs.Name)
						return true
					}
				}
			}
			return false
		}, 60*time.Second, 5*time.Second).Should(gomega.BeTrue(), "At least one pod should have OOMed")

		podList, err := GetOOMPods(f)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		initialPodSet := MakePodSet(podList)

		ginkgo.By("Waiting to verify pods are NOT evicted (OOM took longer than 1s threshold)")
		CheckNoPodsEvictedOOM(f, initialPodSet)
	})

	f.It("evicts pods when OOM occurs within threshold", framework.WithFeatureGate(features.PerVPAConfig), func() {
		ginkgo.By("Setting up the Admission Controller status")
		stopCh := make(chan struct{})
		statusUpdater := status.NewUpdater(
			f.ClientSet,
			status.AdmissionControllerStatusName,
			utils.VpaNamespace,
			statusUpdateInterval,
			"e2e test",
		)
		defer func() {
			ginkgo.By("Deleting the Admission Controller status")
			close(stopCh)
			err := f.ClientSet.CoordinationV1().Leases(utils.VpaNamespace).
				Delete(context.TODO(), status.AdmissionControllerStatusName, metav1.DeleteOptions{})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		}()
		statusUpdater.Run(stopCh)

		// Long threshold - OOM (~5s) < threshold → quickOOM = true → evict
		threshold := int32(600)

		ginkgo.By("Setting up VPA with very short evictAfterOOMSeconds (1 second)")
		containerName := utils.GetHamsterContainerNameByIndex(0)
		vpaCRD := newOOMTestVPA(f.Namespace.Name, "hamster-vpa", "hamster", containerName, threshold)
		utils.InstallVPA(f, vpaCRD)

		ginkgo.By("Creating deployment that will OOM (takes >1 second to OOM)")
		runOomingReplicationController(f.ClientSet, f.Namespace.Name, "hamster", replicas)

		ginkgo.By("Waiting for pods to OOM and restart")
		gomega.Eventually(func() bool {
			podList, err := GetOOMPods(f)
			if err != nil {
				return false
			}
			for _, pod := range podList.Items {
				for _, cs := range pod.Status.ContainerStatuses {
					if cs.LastTerminationState.Terminated != nil &&
						cs.LastTerminationState.Terminated.Reason == "OOMKilled" {
						framework.Logf("Pod %s container %s was OOMKilled", pod.Name, cs.Name)
						return true
					}
				}
			}
			return false
		}, 60*time.Second, 5*time.Second).Should(gomega.BeTrue(), "At least one pod should have OOMed")

		podList, err := GetOOMPods(f)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		ginkgo.By("Waiting for pods to be evicted")
		err = WaitForPodsEvictedOOM(f, podList)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	})
})

var _ = UpdaterE2eDescribe("Updater", func() {
	f := framework.NewDefaultFramework("vertical-pod-autoscaling")
	f.NamespacePodSecurityLevel = podsecurity.LevelBaseline

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
		WithCPUStartupBoost(vpa_types.FactorStartupBoostType, &factor, nil, 1).
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

func setupPodsForDownscalingEviction(f *framework.Framework, er []*vpa_types.EvictionRequirement, updateMode vpa_types.UpdateMode) *apiv1.PodList {
	return setupPodsForEviction(f, "500m", "500Mi", er, updateMode)
}

func setupPodsForEviction(f *framework.Framework, hamsterCPU, hamsterMemory string, er []*vpa_types.EvictionRequirement, updateMode vpa_types.UpdateMode) *apiv1.PodList {
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
		WithUpdateMode(updateMode).
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

func setupPodsForUpscalingInPlace(f *framework.Framework, updateMode vpa_types.UpdateMode) *apiv1.PodList {
	return setupPodsForInPlace(f, "100m", "100Mi", nil, true, updateMode)
}

func setupPodsForDownscalingInPlace(f *framework.Framework, er []*vpa_types.EvictionRequirement, updateMode vpa_types.UpdateMode) *apiv1.PodList {
	return setupPodsForInPlace(f, "500m", "500Mi", er, true, updateMode)
}

func setupPodsForInPlace(f *framework.Framework, hamsterCPU, hamsterMemory string, er []*vpa_types.EvictionRequirement, withRecommendation bool, updateMode vpa_types.UpdateMode) *apiv1.PodList {
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
		WithUpdateMode(updateMode).
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

// newOOMTestVPA creates a VPA for OOM testing with memory-only recommendations.
// Memory bounds are set so that 1Gi request is within range (no OutsideRecommendedRange eviction)
// but target differs from request (ResourceDiff > 0, enabling eviction when quickOOM = true).
func newOOMTestVPA(namespace, name, deploymentName, containerName string, oomThreshold int32) *vpa_types.VerticalPodAutoscaler {
	updateMode := vpa_types.UpdateModeRecreate
	return &vpa_types.VerticalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: vpa_types.VerticalPodAutoscalerSpec{
			TargetRef: &autoscaling.CrossVersionObjectReference{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
				Name:       deploymentName,
			},
			UpdatePolicy: &vpa_types.PodUpdatePolicy{
				UpdateMode:           &updateMode,
				EvictAfterOOMSeconds: &oomThreshold,
			},
			ResourcePolicy: &vpa_types.PodResourcePolicy{
				ContainerPolicies: []vpa_types.ContainerResourcePolicy{{
					ContainerName: containerName,
				}},
			},
		},
		Status: vpa_types.VerticalPodAutoscalerStatus{
			Recommendation: &vpa_types.RecommendedPodResources{
				ContainerRecommendations: []vpa_types.RecommendedContainerResources{{
					ContainerName: containerName,
					Target: apiv1.ResourceList{
						apiv1.ResourceMemory: resource.MustParse("2Gi"), // Different from 1Gi request
					},
					LowerBound: apiv1.ResourceList{
						apiv1.ResourceMemory: resource.MustParse("500Mi"), // 1Gi is within [500Mi, 3Gi]
					},
					UpperBound: apiv1.ResourceList{
						apiv1.ResourceMemory: resource.MustParse("3Gi"),
					},
				}},
			},
		},
	}
}
