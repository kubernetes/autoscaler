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

	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/test"
	"k8s.io/utils/pointer"

	appsv1 "k8s.io/api/apps/v1"
	autoscaling "k8s.io/api/autoscaling/v1"
	apiv1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/autoscaler/vertical-pod-autoscaler/e2e/utils"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/annotations"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/test/e2e/framework"
	framework_deployment "k8s.io/kubernetes/test/e2e/framework/deployment"
	framework_job "k8s.io/kubernetes/test/e2e/framework/job"
	framework_rc "k8s.io/kubernetes/test/e2e/framework/rc"
	framework_rs "k8s.io/kubernetes/test/e2e/framework/replicaset"
	framework_ss "k8s.io/kubernetes/test/e2e/framework/statefulset"
	testutils "k8s.io/kubernetes/test/utils"
	podsecurity "k8s.io/pod-security-admission/api"

	ginkgo "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ActuationSuiteE2eDescribe("Actuation", func() {
	f := framework.NewDefaultFramework("vertical-pod-autoscaling")
	f.NamespacePodSecurityEnforceLevel = podsecurity.LevelBaseline

	ginkgo.It("stops when pods get pending", func() {

		ginkgo.By("Setting up a hamster deployment")
		d := SetupHamsterDeployment(f, "100m", "100Mi", defaultHamsterReplicas)

		ginkgo.By("Setting up a VPA CRD with ridiculous request")
		containerName := GetHamsterContainerNameByIndex(0)
		vpaCRD := test.VerticalPodAutoscaler().
			WithName("hamster-vpa").
			WithNamespace(f.Namespace.Name).
			WithTargetRef(hamsterTargetRef).
			WithContainer(containerName).
			AppendRecommendation(
				test.Recommendation().
					WithContainer(containerName).
					WithTarget("9999", ""). // Request 9999 CPUs to make POD pending
					WithLowerBound("9999", "").
					WithUpperBound("9999", "").
					GetContainerResources()).
			Get()

		InstallVPA(f, vpaCRD)

		ginkgo.By("Waiting for pods to be restarted and stuck pending")
		err := assertPodsPendingForDuration(f.ClientSet, d, 1, 2*time.Minute)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

	})

	ginkgo.It("never applies recommendations when update mode is Off", func() {
		ginkgo.By("Setting up a hamster deployment")
		d := SetupHamsterDeployment(f, "100m", "100Mi", defaultHamsterReplicas)
		cpuRequest := getCPURequest(d.Spec.Template.Spec)
		podList, err := GetHamsterPods(f)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		podSet := MakePodSet(podList)

		ginkgo.By("Setting up a VPA CRD in mode Off")
		containerName := GetHamsterContainerNameByIndex(0)
		vpaCRD := test.VerticalPodAutoscaler().
			WithName("hamster-vpa").
			WithNamespace(f.Namespace.Name).
			WithTargetRef(hamsterTargetRef).
			WithUpdateMode(vpa_types.UpdateModeOff).
			WithContainer(containerName).
			AppendRecommendation(
				test.Recommendation().
					WithContainer(containerName).
					WithTarget("200m", "").
					WithLowerBound("200m", "").
					WithUpperBound("200m", "").
					GetContainerResources()).
			Get()

		InstallVPA(f, vpaCRD)

		ginkgo.By(fmt.Sprintf("Waiting for pods to be evicted, hoping it won't happen, sleep for %s", VpaEvictionTimeout.String()))
		CheckNoPodsEvicted(f, podSet)
		ginkgo.By("Forcefully killing one pod")
		killPod(f, podList)

		ginkgo.By("Checking the requests were not modified")
		updatedPodList, err := GetHamsterPods(f)
		for _, pod := range updatedPodList.Items {
			gomega.Expect(getCPURequest(pod.Spec)).To(gomega.Equal(cpuRequest))
		}
	})

	ginkgo.It("applies recommendations only on restart when update mode is Initial", func() {
		ginkgo.By("Setting up a hamster deployment")
		SetupHamsterDeployment(f, "100m", "100Mi", defaultHamsterReplicas)
		podList, err := GetHamsterPods(f)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		podSet := MakePodSet(podList)

		ginkgo.By("Setting up a VPA CRD in mode Initial")
		containerName := GetHamsterContainerNameByIndex(0)
		vpaCRD := test.VerticalPodAutoscaler().
			WithName("hamster-vpa").
			WithNamespace(f.Namespace.Name).
			WithTargetRef(hamsterTargetRef).
			WithUpdateMode(vpa_types.UpdateModeInitial).
			WithContainer(containerName).
			AppendRecommendation(
				test.Recommendation().
					WithContainer(containerName).
					WithTarget("200m", "").
					WithLowerBound("200m", "").
					WithUpperBound("200m", "").
					GetContainerResources()).
			Get()

		InstallVPA(f, vpaCRD)
		updatedCPURequest := ParseQuantityOrDie("200m")

		ginkgo.By(fmt.Sprintf("Waiting for pods to be evicted, hoping it won't happen, sleep for %s", VpaEvictionTimeout.String()))
		CheckNoPodsEvicted(f, podSet)
		ginkgo.By("Forcefully killing one pod")
		killPod(f, podList)

		ginkgo.By("Checking that request was modified after forceful restart")
		updatedPodList, err := GetHamsterPods(f)
		foundUpdated := 0
		for _, pod := range updatedPodList.Items {
			podRequest := getCPURequest(pod.Spec)
			framework.Logf("podReq: %v", podRequest)
			if podRequest.Cmp(updatedCPURequest) == 0 {
				foundUpdated += 1
			}
		}
		gomega.Expect(foundUpdated).To(gomega.Equal(1))
	})

	perControllerTests := []struct {
		apiVersion string
		kind       string
		name       string
	}{
		{
			apiVersion: "apps/v1",
			kind:       "Deployment",
			name:       "hamster-deployment",
		},
		{
			apiVersion: "v1",
			kind:       "ReplicationController",
			name:       "hamster-rc",
		},
		{
			apiVersion: "batch/v1",
			kind:       "Job",
			name:       "hamster-job",
		},
		{
			apiVersion: "batch/v1",
			kind:       "CronJob",
			name:       "hamster-cronjob",
		},
		{
			apiVersion: "apps/v1",
			kind:       "ReplicaSet",
			name:       "hamster-rs",
		},
		{
			apiVersion: "apps/v1",
			kind:       "StatefulSet",
			name:       "hamster-stateful",
		},
	}
	for _, tc := range perControllerTests {
		ginkgo.It("evicts pods in a multiple-replica "+tc.kind, func() {
			testEvictsReplicatedPods(f, &autoscaling.CrossVersionObjectReference{
				APIVersion: tc.apiVersion,
				Kind:       tc.kind,
				Name:       tc.name,
			})
		})
		ginkgo.It("by default does not evict pods in a 1-Pod "+tc.kind, func() {
			testDoesNotEvictSingletonPodByDefault(f, &autoscaling.CrossVersionObjectReference{
				APIVersion: tc.apiVersion,
				Kind:       tc.kind,
				Name:       tc.name,
			})
		})
		ginkgo.It("when configured, evicts pods in a 1-Pod "+tc.kind, func() {
			testEvictsSingletonPodWhenConfigured(f, &autoscaling.CrossVersionObjectReference{
				APIVersion: tc.apiVersion,
				Kind:       tc.kind,
				Name:       tc.name,
			})
		})
	}

	ginkgo.It("observes pod disruption budget", func() {

		ginkgo.By("Setting up a hamster deployment")
		c := f.ClientSet
		ns := f.Namespace.Name

		SetupHamsterDeployment(f, "10m", "10Mi", 10)
		podList, err := GetHamsterPods(f)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		podSet := MakePodSet(podList)

		ginkgo.By("Setting up prohibitive PDB for hamster deployment")
		pdb := setupPDB(f, "hamster-pdb", 0 /* maxUnavailable */)

		ginkgo.By("Setting up a VPA CRD")
		containerName := GetHamsterContainerNameByIndex(0)
		vpaCRD := test.VerticalPodAutoscaler().
			WithName("hamster-vpa").
			WithNamespace(f.Namespace.Name).
			WithTargetRef(hamsterTargetRef).
			WithContainer(containerName).
			AppendRecommendation(
				test.Recommendation().
					WithContainer(containerName).
					WithTarget("25m", "").
					WithLowerBound("25m", "").
					WithUpperBound("25m", "").
					GetContainerResources()).
			Get()

		InstallVPA(f, vpaCRD)

		ginkgo.By(fmt.Sprintf("Waiting for pods to be evicted, hoping it won't happen, sleep for %s", VpaEvictionTimeout.String()))
		CheckNoPodsEvicted(f, podSet)

		ginkgo.By("Updating the PDB to allow for multiple pods to be evicted")
		// We will check that 7 replicas are evicted in 3 minutes, which translates
		// to 3 updater loops. This gives us relatively good confidence that updater
		// evicts more than one pod in a loop if PDB allows it.
		permissiveMaxUnavailable := 7
		// Creating new PDB and removing old one, since PDBs are immutable at the moment
		setupPDB(f, "hamster-pdb-2", permissiveMaxUnavailable)
		err = c.PolicyV1().PodDisruptionBudgets(ns).Delete(context.TODO(), pdb.Name, metav1.DeleteOptions{})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		ginkgo.By(fmt.Sprintf("Waiting for pods to be evicted, sleep for %s", VpaEvictionTimeout.String()))
		time.Sleep(VpaEvictionTimeout)
		ginkgo.By("Checking enough pods were evicted.")
		currentPodList, err := GetHamsterPods(f)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		evictedCount := GetEvictedPodsCount(MakePodSet(currentPodList), podSet)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(evictedCount >= permissiveMaxUnavailable).To(gomega.BeTrue())
	})

	ginkgo.It("observes container max in LimitRange", func() {
		ginkgo.By("Setting up a hamster deployment")
		d := NewHamsterDeploymentWithResourcesAndLimits(f,
			ParseQuantityOrDie("100m") /*cpu request*/, ParseQuantityOrDie("200Mi"), /*memory request*/
			ParseQuantityOrDie("300m") /*cpu limit*/, ParseQuantityOrDie("400Mi") /*memory limit*/)
		podList := startDeploymentPods(f, d)

		ginkgo.By("Setting up a VPA CRD")
		containerName := GetHamsterContainerNameByIndex(0)
		vpaCRD := test.VerticalPodAutoscaler().
			WithName("hamster-vpa").
			WithNamespace(f.Namespace.Name).
			WithTargetRef(hamsterTargetRef).
			WithContainer(containerName).
			AppendRecommendation(
				test.Recommendation().
					WithContainer(containerName).
					WithTarget("200m", "").
					WithLowerBound("200m", "").
					WithUpperBound("200m", "").
					GetContainerResources()).
			Get()

		InstallVPA(f, vpaCRD)

		// Max CPU limit is 300m and ratio is 3., so max request is 100m, while
		// recommendation is 200m
		// Max memory limit is 1T and ratio is 2., so max request is 0.5T
		InstallLimitRangeWithMax(f, "300m", "1T", apiv1.LimitTypeContainer)

		ginkgo.By(fmt.Sprintf("Waiting for pods to be evicted, hoping it won't happen, sleep for %s", VpaEvictionTimeout.String()))
		CheckNoPodsEvicted(f, MakePodSet(podList))
	})

	ginkgo.It("observes container min in LimitRange", func() {
		ginkgo.By("Setting up a hamster deployment")
		d := NewHamsterDeploymentWithResourcesAndLimits(f,
			ParseQuantityOrDie("100m") /*cpu request*/, ParseQuantityOrDie("200Mi"), /*memory request*/
			ParseQuantityOrDie("300m") /*cpu limit*/, ParseQuantityOrDie("400Mi") /*memory limit*/)
		podList := startDeploymentPods(f, d)

		ginkgo.By("Setting up a VPA CRD")
		containerName := GetHamsterContainerNameByIndex(0)
		vpaCRD := test.VerticalPodAutoscaler().
			WithName("hamster-vpa").
			WithNamespace(f.Namespace.Name).
			WithTargetRef(hamsterTargetRef).
			WithContainer(containerName).
			AppendRecommendation(
				test.Recommendation().
					WithContainer(containerName).
					WithTarget("50m", "").
					WithLowerBound("50m", "").
					WithUpperBound("50m", "").
					GetContainerResources()).
			Get()

		InstallVPA(f, vpaCRD)

		// Min CPU from limit range is 100m and ratio is 3. Min applies both to limit and request so min
		// request is 100m request and 300m limit
		// Min memory limit is 0 and ratio is 2., so min request is 0
		InstallLimitRangeWithMin(f, "100m", "0", apiv1.LimitTypeContainer)

		ginkgo.By(fmt.Sprintf("Waiting for pods to be evicted, hoping it won't happen, sleep for %s", VpaEvictionTimeout.String()))
		CheckNoPodsEvicted(f, MakePodSet(podList))
	})

	ginkgo.It("observes pod max in LimitRange", func() {
		ginkgo.By("Setting up a hamster deployment")
		d := NewHamsterDeploymentWithResourcesAndLimits(f,
			ParseQuantityOrDie("100m") /*cpu request*/, ParseQuantityOrDie("200Mi"), /*memory request*/
			ParseQuantityOrDie("300m") /*cpu limit*/, ParseQuantityOrDie("400Mi") /*memory limit*/)
		d.Spec.Template.Spec.Containers = append(d.Spec.Template.Spec.Containers, d.Spec.Template.Spec.Containers[0])
		d.Spec.Template.Spec.Containers[1].Name = "hamster2"
		podList := startDeploymentPods(f, d)

		ginkgo.By("Setting up a VPA CRD")
		container1Name := GetHamsterContainerNameByIndex(0)
		container2Name := GetHamsterContainerNameByIndex(1)
		vpaCRD := test.VerticalPodAutoscaler().
			WithName("hamster-vpa").
			WithNamespace(f.Namespace.Name).
			WithTargetRef(hamsterTargetRef).
			WithContainer(container1Name).
			AppendRecommendation(
				test.Recommendation().
					WithContainer(container1Name).
					WithTarget("200m", "").
					WithLowerBound("200m", "").
					WithUpperBound("200m", "").
					GetContainerResources()).
			WithContainer(container2Name).
			AppendRecommendation(
				test.Recommendation().
					WithContainer(container2Name).
					WithTarget("200m", "").
					WithLowerBound("200m", "").
					WithUpperBound("200m", "").
					GetContainerResources()).
			Get()

		InstallVPA(f, vpaCRD)

		// Max CPU limit is 600m per pod, 300m per container and ratio is 3., so max request is 100m,
		// while recommendation is 200m
		// Max memory limit is 2T per pod, 1T per container and ratio is 2., so max request is 0.5T
		InstallLimitRangeWithMax(f, "600m", "2T", apiv1.LimitTypePod)

		ginkgo.By(fmt.Sprintf("Waiting for pods to be evicted, hoping it won't happen, sleep for %s", VpaEvictionTimeout.String()))
		CheckNoPodsEvicted(f, MakePodSet(podList))
	})

	ginkgo.It("observes pod min in LimitRange", func() {
		ginkgo.By("Setting up a hamster deployment")
		d := NewHamsterDeploymentWithResourcesAndLimits(f,
			ParseQuantityOrDie("100m") /*cpu request*/, ParseQuantityOrDie("200Mi"), /*memory request*/
			ParseQuantityOrDie("300m") /*cpu limit*/, ParseQuantityOrDie("400Mi") /*memory limit*/)
		d.Spec.Template.Spec.Containers = append(d.Spec.Template.Spec.Containers, d.Spec.Template.Spec.Containers[0])
		container2Name := "hamster2"
		d.Spec.Template.Spec.Containers[1].Name = container2Name
		podList := startDeploymentPods(f, d)

		ginkgo.By("Setting up a VPA CRD")
		container1Name := GetHamsterContainerNameByIndex(0)
		vpaCRD := test.VerticalPodAutoscaler().
			WithName("hamster-vpa").
			WithNamespace(f.Namespace.Name).
			WithTargetRef(hamsterTargetRef).
			WithContainer(container1Name).
			AppendRecommendation(
				test.Recommendation().
					WithContainer(container1Name).
					WithTarget("50m", "").
					WithLowerBound("50m", "").
					WithUpperBound("50m", "").
					GetContainerResources()).
			WithContainer(container2Name).
			AppendRecommendation(
				test.Recommendation().
					WithContainer(container2Name).
					WithTarget("50m", "").
					WithLowerBound("50m", "").
					WithUpperBound("50m", "").
					GetContainerResources()).
			Get()

		InstallVPA(f, vpaCRD)

		// Min CPU from limit range is 200m per pod, 100m per container and ratio is 3. Min applies both
		// to limit and request so min request is 100m request and 300m limit
		// Min memory limit is 0 and ratio is 2., so min request is 0
		InstallLimitRangeWithMin(f, "200m", "0", apiv1.LimitTypePod)

		ginkgo.By(fmt.Sprintf("Waiting for pods to be evicted, hoping it won't happen, sleep for %s", VpaEvictionTimeout.String()))
		CheckNoPodsEvicted(f, MakePodSet(podList))
	})

	ginkgo.It("does not act on injected sidecars", func() {
		const (
			agnhostImage  = "registry.k8s.io/e2e-test-images/agnhost:2.40"
			sidecarParam  = "--sidecar-image=registry.k8s.io/pause:3.1"
			servicePort   = int32(8443)
			containerPort = int32(8444)
		)

		ginkgo.By("Setting up Webhook for sidecar injection")

		client := f.ClientSet
		namespaceName := f.Namespace.Name
		defer utils.CleanWebhookTest(client, namespaceName)

		// Make sure the namespace created for the test is labeled to be selected by the webhooks.
		utils.LabelNamespace(f, f.Namespace.Name)
		utils.CreateWebhookConfigurationReadyNamespace(f)

		ginkgo.By("Setting up server cert")
		context := utils.SetupWebhookCert(namespaceName)
		utils.CreateAuthReaderRoleBinding(f, namespaceName)

		utils.DeployWebhookAndService(f, agnhostImage, context, servicePort, containerPort, sidecarParam)

		// Webhook must be placed after vpa webhook. Webhooks are registered alphabetically.
		// Use name that starts with "z".
		webhookCleanup := utils.RegisterMutatingWebhookForPod(f, "z-sidecar-injection-webhook", context, servicePort)
		defer webhookCleanup()

		ginkgo.By("Setting up a hamster vpa")
		container1Name := GetHamsterContainerNameByIndex(0)
		container2Name := GetHamsterContainerNameByIndex(1)
		vpaCRD := test.VerticalPodAutoscaler().
			WithName("hamster-vpa").
			WithNamespace(f.Namespace.Name).
			WithTargetRef(hamsterTargetRef).
			WithContainer(container1Name).
			AppendRecommendation(
				test.Recommendation().
					WithContainer(container1Name).
					WithTarget("100m", "").
					WithLowerBound("100m", "").
					WithUpperBound("100m", "").
					GetContainerResources()).
			WithContainer(container2Name).
			AppendRecommendation(
				test.Recommendation().
					WithContainer(container2Name).
					WithTarget("5000m", "").
					WithLowerBound("5000m", "").
					WithUpperBound("5000m", "").
					GetContainerResources()).
			Get()

		InstallVPA(f, vpaCRD)

		ginkgo.By("Setting up a hamster deployment")

		d := NewHamsterDeploymentWithResources(f, ParseQuantityOrDie("100m"), ParseQuantityOrDie("100Mi"))
		podList := startDeploymentPods(f, d)
		for _, pod := range podList.Items {
			observedContainers, ok := pod.GetAnnotations()[annotations.VpaObservedContainersLabel]
			gomega.Expect(ok).To(gomega.Equal(true))
			containers, err := annotations.ParseVpaObservedContainersValue(observedContainers)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(containers).To(gomega.HaveLen(1))
			gomega.Expect(pod.Spec.Containers).To(gomega.HaveLen(2))
		}

		podSet := MakePodSet(podList)
		ginkgo.By(fmt.Sprintf("Waiting for pods to be evicted, hoping it won't happen, sleep for %s", VpaEvictionTimeout.String()))
		CheckNoPodsEvicted(f, podSet)
	})
})

func getCPURequest(podSpec apiv1.PodSpec) resource.Quantity {
	return podSpec.Containers[0].Resources.Requests[apiv1.ResourceCPU]
}

func killPod(f *framework.Framework, podList *apiv1.PodList) {
	f.ClientSet.CoreV1().Pods(f.Namespace.Name).Delete(context.TODO(), podList.Items[0].Name, metav1.DeleteOptions{})
	err := WaitForPodsRestarted(f, podList)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
}

// assertPodsPendingForDuration checks that at most pendingPodsNum pods are pending for pendingDuration
func assertPodsPendingForDuration(c clientset.Interface, deployment *appsv1.Deployment, pendingPodsNum int, pendingDuration time.Duration) error {

	pendingPods := make(map[string]time.Time)

	err := wait.PollUntilContextTimeout(context.Background(), pollInterval, pollTimeout, true, func(ctx context.Context) (done bool, err error) {
		currentPodList, err := framework_deployment.GetPodsForDeployment(ctx, c, deployment)
		if err != nil {
			return false, err
		}

		missingPods := make(map[string]bool)
		for podName := range pendingPods {
			missingPods[podName] = true
		}

		now := time.Now()
		for _, pod := range currentPodList.Items {
			delete(missingPods, pod.Name)
			switch pod.Status.Phase {
			case apiv1.PodPending:
				_, ok := pendingPods[pod.Name]
				if !ok {
					pendingPods[pod.Name] = now
				}
			default:
				delete(pendingPods, pod.Name)
			}
		}

		for missingPod := range missingPods {
			delete(pendingPods, missingPod)
		}

		if len(pendingPods) < pendingPodsNum {
			return false, nil
		}

		if len(pendingPods) > pendingPodsNum {
			return false, fmt.Errorf("%v pending pods seen - expecting %v", len(pendingPods), pendingPodsNum)
		}

		for p, t := range pendingPods {
			fmt.Println("task", now, p, t, now.Sub(t), pendingDuration)
			if now.Sub(t) < pendingDuration {
				return false, nil
			}
		}

		return true, nil
	})

	if err != nil {
		return fmt.Errorf("assertion failed for pending pods in %v: %v", deployment.Name, err)
	}
	return nil
}

func testEvictsReplicatedPods(f *framework.Framework, controller *autoscaling.CrossVersionObjectReference) {
	ginkgo.By(fmt.Sprintf("Setting up a hamster %v", controller.Kind))
	setupHamsterController(f, controller.Kind, "100m", "100Mi", defaultHamsterReplicas)
	podList, err := GetHamsterPods(f)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	ginkgo.By("Setting up a VPA CRD")
	containerName := GetHamsterContainerNameByIndex(0)
	vpaCRD := test.VerticalPodAutoscaler().
		WithName("hamster-vpa").
		WithNamespace(f.Namespace.Name).
		WithTargetRef(controller).
		WithContainer(containerName).
		AppendRecommendation(
			test.Recommendation().
				WithContainer(containerName).
				WithTarget("200m", "").
				WithLowerBound("200m", "").
				WithUpperBound("200m", "").
				GetContainerResources()).
		Get()

	InstallVPA(f, vpaCRD)

	ginkgo.By("Waiting for pods to be evicted")
	err = WaitForPodsEvicted(f, podList)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
}

func testDoesNotEvictSingletonPodByDefault(f *framework.Framework, controller *autoscaling.CrossVersionObjectReference) {
	ginkgo.By(fmt.Sprintf("Setting up a hamster %v", controller.Kind))
	setupHamsterController(f, controller.Kind, "100m", "100Mi", 1 /*replicas*/)
	podList, err := GetHamsterPods(f)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	ginkgo.By("Setting up a VPA CRD")
	containerName := GetHamsterContainerNameByIndex(0)
	vpaCRD := test.VerticalPodAutoscaler().
		WithName("hamster-vpa").
		WithNamespace(f.Namespace.Name).
		WithTargetRef(controller).
		WithContainer(containerName).
		AppendRecommendation(
			test.Recommendation().
				WithContainer(containerName).
				WithTarget("200m", "").
				WithLowerBound("200m", "").
				WithUpperBound("200m", "").
				GetContainerResources()).
		Get()

	InstallVPA(f, vpaCRD)

	// No eviction is expected with the default settings of VPA object
	ginkgo.By(fmt.Sprintf("Waiting for pods to be evicted, hoping it won't happen, sleep for %s", VpaEvictionTimeout.String()))
	CheckNoPodsEvicted(f, MakePodSet(podList))
}

func testEvictsSingletonPodWhenConfigured(f *framework.Framework, controller *autoscaling.CrossVersionObjectReference) {
	ginkgo.By(fmt.Sprintf("Setting up a hamster %v", controller.Kind))
	setupHamsterController(f, controller.Kind, "100m", "100Mi", 1 /*replicas*/)
	podList, err := GetHamsterPods(f)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	// Prepare the VPA to allow single-Pod eviction.
	ginkgo.By("Setting up a VPA CRD")
	containerName := GetHamsterContainerNameByIndex(0)
	vpaCRD := test.VerticalPodAutoscaler().
		WithName("hamster-vpa").
		WithNamespace(f.Namespace.Name).
		WithTargetRef(controller).
		WithMinReplicas(pointer.Int32(1)).
		WithContainer(containerName).
		AppendRecommendation(
			test.Recommendation().
				WithContainer(containerName).
				WithTarget("200m", "").
				WithLowerBound("200m", "").
				WithUpperBound("200m", "").
				GetContainerResources()).
		Get()

	InstallVPA(f, vpaCRD)

	ginkgo.By("Waiting for pods to be evicted")
	err = WaitForPodsEvicted(f, podList)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
}

func setupHamsterController(f *framework.Framework, controllerKind, cpu, memory string, replicas int32) *apiv1.PodList {
	switch controllerKind {
	case "Deployment":
		SetupHamsterDeployment(f, cpu, memory, replicas)
	case "ReplicationController":
		setupHamsterReplicationController(f, cpu, memory, replicas)
	case "Job":
		setupHamsterJob(f, cpu, memory, replicas)
	case "CronJob":
		SetupHamsterCronJob(f, "*/2 * * * *", cpu, memory, replicas)
	case "ReplicaSet":
		setupHamsterRS(f, cpu, memory, replicas)
	case "StatefulSet":
		setupHamsterStateful(f, cpu, memory, replicas)
	default:
		framework.Failf("Unknown controller kind: %v", controllerKind)
		return nil
	}
	pods, err := GetHamsterPods(f)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	return pods
}

func setupHamsterReplicationController(f *framework.Framework, cpu, memory string, replicas int32) {
	hamsterContainer := SetupHamsterContainer(cpu, memory)
	rc := framework_rc.ByNameContainer("hamster-rc", replicas, hamsterLabels, hamsterContainer, nil)

	rc.Namespace = f.Namespace.Name
	err := testutils.CreateRCWithRetries(f.ClientSet, f.Namespace.Name, rc)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	err = waitForRCPodsRunning(f, rc)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
}

func waitForRCPodsRunning(f *framework.Framework, rc *apiv1.ReplicationController) error {
	return wait.PollUntilContextTimeout(context.Background(), pollInterval, pollTimeout, true, func(ctx context.Context) (done bool, err error) {
		podList, err := GetHamsterPods(f)
		if err != nil {
			framework.Logf("Error listing pods, retrying: %v", err)
			return false, nil
		}
		podsRunning := int32(0)
		for _, pod := range podList.Items {
			if pod.Status.Phase == apiv1.PodRunning {
				podsRunning += 1
			}
		}
		return podsRunning == *rc.Spec.Replicas, nil
	})
}

func setupHamsterJob(f *framework.Framework, cpu, memory string, replicas int32) {
	job := framework_job.NewTestJob("notTerminate", "hamster-job", apiv1.RestartPolicyOnFailure,
		replicas, replicas, nil, 10)
	job.Spec.Template.Spec.Containers[0] = SetupHamsterContainer(cpu, memory)
	for label, value := range hamsterLabels {
		job.Spec.Template.Labels[label] = value
	}
	_, err := framework_job.CreateJob(context.TODO(), f.ClientSet, f.Namespace.Name, job)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	err = framework_job.WaitForJobPodsRunning(context.TODO(), f.ClientSet, f.Namespace.Name, job.Name, replicas)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
}

func setupHamsterRS(f *framework.Framework, cpu, memory string, replicas int32) {
	rs := newReplicaSet("hamster-rs", f.Namespace.Name, replicas, hamsterLabels, "", "")
	rs.Spec.Template.Spec.Containers[0] = SetupHamsterContainer(cpu, memory)
	err := createReplicaSetWithRetries(f.ClientSet, f.Namespace.Name, rs)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	err = framework_rs.WaitForReadyReplicaSet(context.TODO(), f.ClientSet, f.Namespace.Name, rs.Name)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
}

func setupHamsterStateful(f *framework.Framework, cpu, memory string, replicas int32) {
	stateful := framework_ss.NewStatefulSet("hamster-stateful", f.Namespace.Name,
		"hamster-service", replicas, nil, nil, hamsterLabels)

	stateful.Spec.Template.Spec.Containers[0] = SetupHamsterContainer(cpu, memory)
	err := createStatefulSetSetWithRetries(f.ClientSet, f.Namespace.Name, stateful)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	framework_ss.WaitForRunningAndReady(context.TODO(), f.ClientSet, *stateful.Spec.Replicas, stateful)
}

func setupPDB(f *framework.Framework, name string, maxUnavailable int) *policyv1.PodDisruptionBudget {
	maxUnavailableIntstr := intstr.FromInt(maxUnavailable)
	pdb := &policyv1.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: policyv1.PodDisruptionBudgetSpec{
			MaxUnavailable: &maxUnavailableIntstr,
			Selector: &metav1.LabelSelector{
				MatchLabels: hamsterLabels,
			},
		},
	}
	_, err := f.ClientSet.PolicyV1().PodDisruptionBudgets(f.Namespace.Name).Create(context.TODO(), pdb, metav1.CreateOptions{})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	return pdb
}

func getCurrentPodSetForDeployment(c clientset.Interface, d *appsv1.Deployment) PodSet {
	podList, err := framework_deployment.GetPodsForDeployment(context.TODO(), c, d)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	return MakePodSet(podList)
}

func createReplicaSetWithRetries(c clientset.Interface, namespace string, obj *appsv1.ReplicaSet) error {
	if obj == nil {
		return fmt.Errorf("object provided to create is empty")
	}
	createFunc := func() (bool, error) {
		_, err := c.AppsV1().ReplicaSets(namespace).Create(context.TODO(), obj, metav1.CreateOptions{})
		if err == nil || apierrs.IsAlreadyExists(err) {
			return true, nil
		}
		return false, fmt.Errorf("failed to create object with non-retriable error: %v", err)
	}
	return testutils.RetryWithExponentialBackOff(createFunc)
}

func createStatefulSetSetWithRetries(c clientset.Interface, namespace string, obj *appsv1.StatefulSet) error {
	if obj == nil {
		return fmt.Errorf("object provided to create is empty")
	}
	createFunc := func() (bool, error) {
		_, err := c.AppsV1().StatefulSets(namespace).Create(context.TODO(), obj, metav1.CreateOptions{})
		if err == nil || apierrs.IsAlreadyExists(err) {
			return true, nil
		}
		return false, fmt.Errorf("failed to create object with non-retriable error: %v", err)
	}
	return testutils.RetryWithExponentialBackOff(createFunc)
}

// newReplicaSet returns a new ReplicaSet.
func newReplicaSet(name, namespace string, replicas int32, podLabels map[string]string, imageName, image string) *appsv1.ReplicaSet {
	return &appsv1.ReplicaSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ReplicaSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Spec: appsv1.ReplicaSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: podLabels,
			},
			Replicas: &replicas,
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: podLabels,
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:            imageName,
							Image:           image,
							SecurityContext: &apiv1.SecurityContext{},
						},
					},
				},
			},
		},
	}
}
