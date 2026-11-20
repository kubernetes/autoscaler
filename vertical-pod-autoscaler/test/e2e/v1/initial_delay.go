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

package autoscaling

import (
	"context"
	"fmt"
	"time"

	ginkgo "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/kubernetes/test/e2e/framework"
	podsecurity "k8s.io/pod-security-admission/api"

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/features"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/status"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/test"
	"k8s.io/autoscaler/vertical-pod-autoscaler/test/e2e/utils"
)

const (
	initialDelayVpaName = "hamster-vpa"
	// longInitialDelay keeps the window open for the whole test.
	longInitialDelay = int32(3600)
	// expiredInitialDelay ends the window, since the VPA is older than 1s by the time it is patched.
	expiredInitialDelay = int32(1)
)

var _ = UpdaterE2eDescribe("Updater with VPAInitialDelay", func() {
	f := framework.NewDefaultFramework("vertical-pod-autoscaling")
	f.NamespacePodSecurityLevel = podsecurity.LevelBaseline

	for _, mode := range []vpa_types.UpdateMode{vpa_types.UpdateModeRecreate, vpa_types.UpdateModeInPlaceOrRecreate, vpa_types.UpdateModeInPlace} {
		args := []any{fmt.Sprintf("does not actuate pods with %s mode until the initial delay window ends", mode), framework.WithSerial(), framework.WithFeatureGate(features.VPAInitialDelay)}
		if mode == vpa_types.UpdateModeInPlace {
			args = append(args, framework.WithFeatureGate(features.InPlace))
		}
		args = append(args, func() {
			defer startAdmissionControllerStatus(f)()

			podList := setupPodsForInitialDelay(f, mode, longInitialDelay, true)
			initialPods := podList.DeepCopy()

			ginkgo.By(fmt.Sprintf("Waiting for pods to be actuated, hoping it won't happen, sleep for %s", VpaEvictionTimeout.String()))
			checkNoPodsActuated(f, initialPods)

			ginkgo.By("Ending the initial delay window")
			patchInitialDelaySeconds(f, expiredInitialDelay)

			if mode == vpa_types.UpdateModeRecreate {
				ginkgo.By("Waiting for pods to be evicted")
				gomega.Expect(WaitForPodsEvicted(f, podList)).To(gomega.Succeed())
			} else {
				ginkgo.By("Waiting for pods to be in-place updated")
				gomega.Expect(WaitForPodsUpdatedWithoutEviction(f, initialPods)).To(gomega.Succeed())
			}
		})
		framework.It(args...)
	}

	framework.It("stops actuating pods when the initial delay window is re-armed", framework.WithSerial(), framework.WithFeatureGate(features.VPAInitialDelay), func() {
		defer startAdmissionControllerStatus(f)()

		podList := setupPodsForInitialDelay(f, vpa_types.UpdateModeRecreate, expiredInitialDelay, false)
		initialPods := podList.DeepCopy()

		ginkgo.By("Re-arming the initial delay window")
		time.Sleep(time.Duration(expiredInitialDelay) * time.Second)
		patchInitialDelaySeconds(f, longInitialDelay)

		ginkgo.By("Setting a recommendation that would trigger eviction")
		vpa, err := utils.GetVpaClientSet(f).AutoscalingV1().VerticalPodAutoscalers(f.Namespace.Name).Get(context.TODO(), initialDelayVpaName, metav1.GetOptions{})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		utils.PatchVpaRecommendation(f, vpa, upscalingRecommendation())

		ginkgo.By(fmt.Sprintf("Waiting for pods to be evicted, hoping it won't happen, sleep for %s", VpaEvictionTimeout.String()))
		checkNoPodsActuated(f, initialPods)
	})
})

var _ = AdmissionControllerE2eDescribe("Admission-controller with VPAInitialDelay", func() {
	f := framework.NewDefaultFramework("vertical-pod-autoscaling")
	f.NamespacePodSecurityLevel = podsecurity.LevelBaseline

	ginkgo.BeforeEach(func() {
		waitForVpaWebhookRegistration(f)
	})

	for _, mode := range []vpa_types.UpdateMode{vpa_types.UpdateModeRecreate, vpa_types.UpdateModeInPlaceOrRecreate, vpa_types.UpdateModeInPlace, vpa_types.UpdateModeInitial} {
		args := []any{fmt.Sprintf("does not apply recommendations with %s mode until the initial delay window ends", mode), framework.WithFeatureGate(features.VPAInitialDelay)}
		if mode == vpa_types.UpdateModeInPlace {
			args = append(args, framework.WithFeatureGate(features.InPlace))
		}
		args = append(args, func() {
			d := NewHamsterDeploymentWithResources(f, ParseQuantityOrDie("100m"), ParseQuantityOrDie("100Mi"))

			ginkgo.By("Setting up a VPA in its initial delay window")
			containerName := utils.GetHamsterContainerNameByIndex(0)
			vpaCRD := test.VerticalPodAutoscaler().
				WithName(initialDelayVpaName).
				WithNamespace(f.Namespace.Name).
				WithTargetRef(utils.HamsterTargetRef).
				WithUpdateMode(mode).
				WithInitialDelaySeconds(longInitialDelay).
				WithContainer(containerName).
				AppendRecommendation(
					test.Recommendation().
						WithContainer(containerName).
						WithTarget("250m", "200Mi").
						GetContainerResources()).
				Get()
			utils.InstallVPA(f, vpaCRD)

			ginkgo.By("Starting pods and verifying they keep their template resources")
			podList := utils.StartDeploymentPods(f, d)
			for _, pod := range podList.Items {
				gomega.Expect(pod.Spec.Containers[0].Resources.Requests[apiv1.ResourceCPU]).To(gomega.Equal(ParseQuantityOrDie("100m")))
				gomega.Expect(pod.Spec.Containers[0].Resources.Requests[apiv1.ResourceMemory]).To(gomega.Equal(ParseQuantityOrDie("100Mi")))
			}

			ginkgo.By("Ending the initial delay window")
			patchInitialDelaySeconds(f, expiredInitialDelay)
			// Give the admission controller time to observe the patched VPA.
			time.Sleep(5 * time.Second)

			ginkgo.By("Scaling up and verifying the new pod gets the recommendation")
			pod := scaleUpAndGetNewPod(f, d.Name, podList)
			gomega.Expect(pod.Spec.Containers[0].Resources.Requests[apiv1.ResourceCPU]).To(gomega.Equal(ParseQuantityOrDie("250m")))
			gomega.Expect(pod.Spec.Containers[0].Resources.Requests[apiv1.ResourceMemory]).To(gomega.Equal(ParseQuantityOrDie("200Mi")))
		})
		f.It(args...)
	}
})

var _ = FullVpaE2eDescribe("Pods under VPA with VPAInitialDelay", func() {
	f := framework.NewDefaultFramework("vertical-pod-autoscaling")
	f.NamespacePodSecurityLevel = podsecurity.LevelBaseline

	f.It("reports the initial delay window in the InitialDelayActive condition", framework.WithFeatureGate(features.VPAInitialDelay), func() {
		SetupHamsterDeployment(f, "100m", "100Mi", 1)

		ginkgo.By("Setting up a VPA in its initial delay window")
		vpaCRD := test.VerticalPodAutoscaler().
			WithName(initialDelayVpaName).
			WithNamespace(f.Namespace.Name).
			WithTargetRef(utils.HamsterTargetRef).
			WithUpdateMode(vpa_types.UpdateModeRecreate).
			WithInitialDelaySeconds(longInitialDelay).
			WithContainer(utils.GetHamsterContainerNameByIndex(0)).
			Get()
		utils.InstallVPA(f, vpaCRD)
		vpaClientSet := utils.GetVpaClientSet(f)

		ginkgo.By("Waiting for InitialDelayActive=True")
		_, err := utils.WaitForVPAMatch(vpaClientSet, vpaCRD, initialDelayConditionIs(apiv1.ConditionTrue))
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		ginkgo.By("Ending the initial delay window and waiting for InitialDelayActive=False")
		patchInitialDelaySeconds(f, expiredInitialDelay)
		_, err = utils.WaitForVPAMatch(vpaClientSet, vpaCRD, initialDelayConditionIs(apiv1.ConditionFalse))
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		ginkgo.By("Re-arming the initial delay window and waiting for InitialDelayActive=True")
		patchInitialDelaySeconds(f, longInitialDelay)
		_, err = utils.WaitForVPAMatch(vpaClientSet, vpaCRD, initialDelayConditionIs(apiv1.ConditionTrue))
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	})
})

// startAdmissionControllerStatus fakes a live Admission Controller so the
// Updater actuates pods. It returns a function that removes the status.
func startAdmissionControllerStatus(f *framework.Framework) func() {
	ginkgo.By("Setting up the Admission Controller status")
	stopCh := make(chan struct{})
	statusUpdater := status.NewUpdater(
		f.ClientSet,
		status.AdmissionControllerStatusName,
		utils.VpaNamespace,
		10*time.Second,
		"e2e test",
	)
	statusUpdater.Run(stopCh)
	return func() {
		ginkgo.By("Deleting the Admission Controller status")
		close(stopCh)
		err := f.ClientSet.CoordinationV1().Leases(utils.VpaNamespace).
			Delete(context.TODO(), status.AdmissionControllerStatusName, metav1.DeleteOptions{})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	}
}

func upscalingRecommendation() *vpa_types.RecommendedPodResources {
	return &vpa_types.RecommendedPodResources{
		ContainerRecommendations: []vpa_types.RecommendedContainerResources{
			test.Recommendation().
				WithContainer(utils.GetHamsterContainerNameByIndex(0)).
				WithTarget("200m", "200Mi").
				WithLowerBound("200m", "200Mi").
				WithUpperBound("200m", "200Mi").
				GetContainerResources(),
		},
	}
}

func setupPodsForInitialDelay(f *framework.Framework, updateMode vpa_types.UpdateMode, initialDelaySeconds int32, withRecommendation bool) *apiv1.PodList {
	ginkgo.By("Setting up a hamster Deployment")
	setupHamsterController(f, "Deployment", "100m", "100Mi", utils.DefaultHamsterReplicas)
	podList, err := GetHamsterPods(f)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	ginkgo.By(fmt.Sprintf("Setting up a VPA with initialDelaySeconds=%d", initialDelaySeconds))
	vpaBuilder := test.VerticalPodAutoscaler().
		WithName(initialDelayVpaName).
		WithNamespace(f.Namespace.Name).
		WithTargetRef(utils.HamsterTargetRef).
		WithUpdateMode(updateMode).
		WithInitialDelaySeconds(initialDelaySeconds).
		WithContainer(utils.GetHamsterContainerNameByIndex(0))
	if withRecommendation {
		vpaBuilder = vpaBuilder.AppendRecommendation(upscalingRecommendation().ContainerRecommendations[0])
	}
	utils.InstallVPA(f, vpaBuilder.Get())

	return podList
}

func patchInitialDelaySeconds(f *framework.Framework, seconds int32) {
	patch := fmt.Appendf(nil, `{"spec":{"updatePolicy":{"initialDelaySeconds":%d}}}`, seconds)
	_, err := utils.GetVpaClientSet(f).AutoscalingV1().VerticalPodAutoscalers(f.Namespace.Name).
		Patch(context.TODO(), initialDelayVpaName, types.MergePatchType, patch, metav1.PatchOptions{})
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), "Failed to patch initialDelaySeconds.")
}

// checkNoPodsActuated waits long enough for the Updater to act, then checks
// that no pod was evicted or resized in place.
func checkNoPodsActuated(f *framework.Framework, initialPods *apiv1.PodList) {
	time.Sleep(VpaEvictionTimeout)
	currentPods, err := GetHamsterPods(f)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(GetEvictedPodsCount(MakePodSet(currentPods), MakePodSet(initialPods))).To(gomega.Equal(0), "there should be no pod evictions")
	for _, initialPod := range initialPods.Items {
		for _, pod := range currentPods.Items {
			if pod.UID != initialPod.UID {
				continue
			}
			for i, container := range pod.Spec.Containers {
				gomega.Expect(container.Resources.Requests).To(gomega.Equal(initialPod.Spec.Containers[i].Resources.Requests),
					"pod %s should not be resized in place", pod.Name)
			}
		}
	}
}

func scaleUpAndGetNewPod(f *framework.Framework, deploymentName string, existing *apiv1.PodList) *apiv1.Pod {
	deployments := f.ClientSet.AppsV1().Deployments(f.Namespace.Name)
	scale, err := deployments.GetScale(context.TODO(), deploymentName, metav1.GetOptions{})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	scale.Spec.Replicas++
	_, err = deployments.UpdateScale(context.TODO(), deploymentName, scale, metav1.UpdateOptions{})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	existingPods := MakePodSet(existing)
	var newPod *apiv1.Pod
	err = wait.PollUntilContextTimeout(context.TODO(), utils.PollInterval, utils.PollTimeout, true, func(ctx context.Context) (bool, error) {
		podList, err := GetHamsterPods(f)
		if err != nil {
			return false, err
		}
		for i := range podList.Items {
			if _, found := existingPods[podList.Items[i].Name]; !found {
				newPod = &podList.Items[i]
				return true, nil
			}
		}
		return false, nil
	})
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), "waiting for the new pod")
	return newPod
}

func initialDelayConditionIs(conditionStatus apiv1.ConditionStatus) func(*vpa_types.VerticalPodAutoscaler) bool {
	return func(vpa *vpa_types.VerticalPodAutoscaler) bool {
		for _, condition := range vpa.Status.Conditions {
			if condition.Type == vpa_types.InitialDelayActive {
				return condition.Status == conditionStatus
			}
		}
		return false
	}
}
