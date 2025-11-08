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
	"strconv"
	"strings"
	"time"

	autoscaling "k8s.io/api/autoscaling/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/e2e/utils"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/status"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/test"
	"k8s.io/kubernetes/test/e2e/framework"
	podsecurity "k8s.io/pod-security-admission/api"
	"k8s.io/utils/ptr"

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

	framework.It("filters pods using the --pod-label-selectors flag", framework.WithSerial(), func() {
		const (
			testLabelKey        = "vpa-updater-test"
			testLabelValueMatch = "enabled"
			matchingReplicas    = 3
			nonMatchingReplicas = 2
		)
		testNamespace := f.Namespace.Name

		ginkgo.By("Creating pods with non-matching labels")
		nonMatchingDeployment := utils.NewNHamstersDeployment(f, 1)
		nonMatchingDeployment.Name = "non-matching-hamster"
		nonMatchingDeployment.Spec.Replicas = ptr.To(int32(nonMatchingReplicas))
		nonMatchingDeployment.Spec.Template.Labels[testLabelKey] = "disabled"
		nonMatchingDeployment.Spec.Template.Labels["app"] = "non-matching"
		nonMatchingDeployment.Spec.Selector.MatchLabels[testLabelKey] = "disabled"
		nonMatchingDeployment.Spec.Selector.MatchLabels["app"] = "non-matching"
		utils.StartDeploymentPods(f, nonMatchingDeployment)

		ginkgo.By("Creating pods with matching labels")
		matchingDeployment := utils.NewNHamstersDeployment(f, 1)
		matchingDeployment.Name = "matching-hamster"
		matchingDeployment.Spec.Replicas = ptr.To(int32(matchingReplicas))
		matchingDeployment.Spec.Template.Labels[testLabelKey] = testLabelValueMatch
		matchingDeployment.Spec.Template.Labels["app"] = "matching"
		matchingDeployment.Spec.Selector.MatchLabels[testLabelKey] = testLabelValueMatch
		matchingDeployment.Spec.Selector.MatchLabels["app"] = "matching"
		utils.StartDeploymentPods(f, matchingDeployment)

		ginkgo.By("Creating VPAs for both deployments")
		containerName := utils.GetHamsterContainerNameByIndex(0)
		nonMatchingVPA := test.VerticalPodAutoscaler().
			WithName("non-matching-vpa").
			WithNamespace(testNamespace).
			WithTargetRef(&autoscaling.CrossVersionObjectReference{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
				Name:       nonMatchingDeployment.Name,
			}).
			WithContainer(containerName).
			WithUpdateMode(vpa_types.UpdateModeRecreate).
			Get()
		utils.InstallVPA(f, nonMatchingVPA)

		matchingVPA := test.VerticalPodAutoscaler().
			WithName("matching-vpa").
			WithNamespace(testNamespace).
			WithTargetRef(&autoscaling.CrossVersionObjectReference{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
				Name:       matchingDeployment.Name,
			}).
			WithContainer(containerName).
			WithUpdateMode(vpa_types.UpdateModeRecreate).
			Get()
		utils.InstallVPA(f, matchingVPA)

		ginkgo.By("Setting up custom updater deployment with --pod-label-selectors flag")
		// we swap the namespace to kube-system and then back to the test namespace
		// so our custom updater deployment can use the deployed RBAC
		originalNamespace := f.Namespace.Name
		f.Namespace.Name = utils.UpdaterNamespace
		deploymentName := "vpa-updater-with-pod-label-selectors"
		updaterDeployment := utils.NewUpdaterDeployment(f, deploymentName, []string{
			"--updater-interval=10s",
			"--use-admission-controller-status=false",
			fmt.Sprintf("--pod-label-selectors=%s=%s", testLabelKey, testLabelValueMatch),
			fmt.Sprintf("--vpa-object-namespace=%s", testNamespace),
		})
		utils.StartDeploymentPods(f, updaterDeployment)
		f.Namespace.Name = originalNamespace

		defer func() {
			ginkgo.By("Cleaning up custom updater deployment")
			f.ClientSet.AppsV1().Deployments(utils.UpdaterNamespace).Delete(context.TODO(), deploymentName, metav1.DeleteOptions{})
		}()

		ginkgo.By("Waiting for custom updater to report controlled pods count via metrics")
		gomega.Eventually(func() (float64, error) {
			return getMetricValue(f, utils.UpdaterNamespace, "vpa_updater_controlled_pods_total", map[string]string{
				"update_mode": string(vpa_types.UpdateModeRecreate),
			})
		}, 2*time.Minute, 5*time.Second).Should(gomega.Equal(float64(matchingReplicas)),
			"Custom updater should only see %d matching pods (not the %d non-matching pods)",
			matchingReplicas, nonMatchingReplicas)
	})
})

func getMetricValue(f *framework.Framework, namespace, metricName string, labels map[string]string) (float64, error) {
	// Port forward to the updater pod
	pods, err := f.ClientSet.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: "app=custom-vpa-updater",
	})
	if err != nil || len(pods.Items) == 0 {
		return 0, fmt.Errorf("updater pod not found: %v", err)
	}

	// Use kubectl port-forward via exec in the pod
	req := f.ClientSet.CoreV1().RESTClient().Get().
		Namespace(namespace).
		Resource("pods").
		Name(pods.Items[0].Name).
		SubResource("proxy").
		Suffix("metrics")

	result := req.Do(context.TODO())
	body, err := result.Raw()
	if err != nil {
		return 0, fmt.Errorf("failed to get metrics: %v", err)
	}

	// Parse Prometheus metrics format
	lines := strings.Split(string(body), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "#") {
			continue
		}
		if !strings.HasPrefix(line, metricName) {
			continue
		}

		// Match labels
		if len(labels) > 0 {
			allLabelsMatch := true
			for k, v := range labels {
				labelPattern := fmt.Sprintf(`%s="%s"`, k, v)
				if !strings.Contains(line, labelPattern) {
					allLabelsMatch = false
					break
				}
			}
			if !allLabelsMatch {
				continue
			}
		}

		// Extract value from end of line
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			value, err := strconv.ParseFloat(parts[len(parts)-1], 64)
			if err == nil {
				return value, nil
			}
		}
	}

	return 0, fmt.Errorf("metric %s not found", metricName)
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
