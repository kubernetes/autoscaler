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

var _ = InPlaceE2eDescribe("In-Place", func() {

	f := framework.NewDefaultFramework("vertical-pod-autoscaling")
	f.NamespacePodSecurityEnforceLevel = podsecurity.LevelBaseline

	// TODO(jkyros): clean this up, some kind of helper, stash the InPlacePodVerticalScalingInUse function somewhere
	// useful
	// TODO(jkyros); the in-place tests check first to see if in-place is in use, and if it's not, there's nothing to test. I bet there's
	// precedent on how to test a gated feature with ginkgo, I should find out what it is
	var InPlacePodVerticalScalingNotInUse bool
	ginkgo.It("Should have InPlacePodVerticalScaling in-use", func() {

		ginkgo.By("Verifying the existence of container ResizePolicy")
		checkPod := &apiv1.Pod{}
		checkPod.Name = "inplace"
		checkPod.Namespace = f.Namespace.Name
		checkPod.Spec.Containers = append(checkPod.Spec.Containers, SetupHamsterContainer("100m", "10Mi"))
		_, err := f.ClientSet.CoreV1().Pods(f.Namespace.Name).Create(context.Background(), checkPod, metav1.CreateOptions{})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		checkPod, err = f.ClientSet.CoreV1().Pods(f.Namespace.Name).Get(context.Background(), checkPod.Name, metav1.GetOptions{})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		if !InPlacePodVerticalScalingInUse(checkPod) {
			InPlacePodVerticalScalingNotInUse = true
			ginkgo.Skip("InPlacePodVerticalScaling was not in use (containers had no ResizePolicy)")
		}
	})

	ginkgo.It("In-place update pods when Admission Controller status available", func() {
		if InPlacePodVerticalScalingNotInUse {
			ginkgo.Skip("InPlacePodVerticalScaling was not in use (containers had no ResizePolicy)")
		}
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
		if InPlacePodVerticalScalingNotInUse {
			ginkgo.Skip("InPlacePodVerticalScaling was not in use (containers had no ResizePolicy)")
		}
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
		if InPlacePodVerticalScalingNotInUse {
			ginkgo.Skip("InPlacePodVerticalScaling was not in use (containers had no ResizePolicy)")
		}

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

	// TODO(jkyros):
	// - It falls back to eviction when in-place is Deferred for a minute (hard to fake, depends on node size)
	// - It falls back to eviction when in-place is Infeasible (easy to fake)
	// - It falls back to eviction when InProgress for more than an hour (maybe fake with annotation?)

})

func setupPodsForUpscalingInPlace(f *framework.Framework) *apiv1.PodList {
	return setupPodsForInPlace(f, "100m", "100Mi", "200m", "200Mi", nil)
}

func setupPodsForDownscalingInPlace(f *framework.Framework, er []*vpa_types.EvictionRequirement) *apiv1.PodList {
	return setupPodsForInPlace(f, "500m", "500Mi", "600m", "600Mi", er)
}

func setupPodsForInPlace(f *framework.Framework, hamsterCPU, hamsterMemory, hamsterCPULimit, hamsterMemoryLimit string, er []*vpa_types.EvictionRequirement) *apiv1.PodList {
	controller := &autoscaling.CrossVersionObjectReference{
		APIVersion: "apps/v1",
		Kind:       "Deployment",
		Name:       "hamster-deployment",
	}
	ginkgo.By(fmt.Sprintf("Setting up a hamster %v", controller.Kind))
	// TODO(jkyros): I didn't want to mangle all the plumbing just yet
	//setupHamsterController(f, controller.Kind, hamsterCPU, hamsterMemory, defaultHamsterReplicas)

	// TODO(jkyros): we can't in-place scale without limits right now because of
	// https://github.com/kubernetes/kubernetes/blob/f4e246bc93ffb68b33ed67c7896c379efa4207e7/pkg/kubelet/kuberuntime/kuberuntime_manager.go#L550,
	// so if we want this to work, we need to add limits for now until we adjust that (assuming we can)
	SetupHamsterDeploymentWithLimits(f, hamsterCPU, hamsterMemory, hamsterCPULimit, hamsterMemoryLimit, defaultHamsterReplicas)
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

// InPlacePodVerticalScalingInUse returns true if pod spec is non-nil and ResizePolicy is set
func InPlacePodVerticalScalingInUse(pod *apiv1.Pod) bool {
	if pod == nil {
		return false
	}
	for _, container := range pod.Spec.Containers {
		if len(container.ResizePolicy) > 0 {
			return true
		}
	}
	return false
}
