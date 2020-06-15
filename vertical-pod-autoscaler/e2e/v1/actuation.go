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

	appsv1 "k8s.io/api/apps/v1"
	autoscaling "k8s.io/api/autoscaling/v1"
	apiv1 "k8s.io/api/core/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
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

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

var _ = ActuationSuiteE2eDescribe("Actuation", func() {
	f := framework.NewDefaultFramework("vertical-pod-autoscaling")

	ginkgo.It("stops when pods get pending", func() {

		ginkgo.By("Setting up a hamster deployment")
		d := SetupHamsterDeployment(f, "100m", "100Mi", defaultHamsterReplicas)

		ginkgo.By("Setting up a VPA CRD with ridiculous request")
		SetupVPA(f, "9999", vpa_types.UpdateModeAuto, hamsterTargetRef) // Request 9999 CPUs to make POD pending

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
		SetupVPA(f, "200m", vpa_types.UpdateModeOff, hamsterTargetRef)

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
		SetupVPA(f, "200m", vpa_types.UpdateModeInitial, hamsterTargetRef)
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

	ginkgo.It("evicts pods in a Deployment", func() {
		testEvictsPods(f, &autoscaling.CrossVersionObjectReference{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
			Name:       "hamster-deployment",
		})
	})

	ginkgo.It("evicts pods in a Replication Controller", func() {
		testEvictsPods(f, &autoscaling.CrossVersionObjectReference{
			APIVersion: "v1",
			Kind:       "ReplicationController",
			Name:       "hamster-rc",
		})
	})

	ginkgo.It("evicts pods in a Job", func() {
		testEvictsPods(f, &autoscaling.CrossVersionObjectReference{
			APIVersion: "batch/v1",
			Kind:       "Job",
			Name:       "hamster-job",
		})
	})

	ginkgo.It("evicts pods in a CronJob", func() {
		testEvictsPods(f, &autoscaling.CrossVersionObjectReference{
			APIVersion: "batch/v1",
			Kind:       "CronJob",
			Name:       "hamster-cronjob",
		})
	})

	ginkgo.It("evicts pods in a ReplicaSet", func() {
		testEvictsPods(f, &autoscaling.CrossVersionObjectReference{
			APIVersion: "apps/v1",
			Kind:       "ReplicaSet",
			Name:       "hamster-rs",
		})
	})

	ginkgo.It("evicts pods in a StatefulSet", func() {
		testEvictsPods(f, &autoscaling.CrossVersionObjectReference{
			APIVersion: "apps/v1",
			Kind:       "StatefulSet",
			Name:       "hamster-stateful",
		})
	})

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
		SetupVPA(f, "25m", vpa_types.UpdateModeAuto, hamsterTargetRef)

		ginkgo.By(fmt.Sprintf("Waiting for pods to be evicted, hoping it won't happen, sleep for %s", VpaEvictionTimeout.String()))
		CheckNoPodsEvicted(f, podSet)

		ginkgo.By("Updating the PDB to allow for multiple pods to be evicted")
		// We will check that 7 replicas are evicted in 3 minutes, which translates
		// to 3 updater loops. This gives us relatively good confidence that updater
		// evicts more than one pod in a loop if PDB allows it.
		permissiveMaxUnavailable := 7
		// Creating new PDB and removing old one, since PDBs are immutable at the moment
		setupPDB(f, "hamster-pdb-2", permissiveMaxUnavailable)
		err = c.PolicyV1beta1().PodDisruptionBudgets(ns).Delete(context.TODO(), pdb.Name, metav1.DeleteOptions{})
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
		SetupVPA(f, "200m", vpa_types.UpdateModeAuto, hamsterTargetRef)

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
		SetupVPA(f, "50m", vpa_types.UpdateModeAuto, hamsterTargetRef)

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
		SetupVPAForNHamsters(f, 2, "200m", vpa_types.UpdateModeAuto, hamsterTargetRef)

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
		d.Spec.Template.Spec.Containers[1].Name = "hamster2"
		podList := startDeploymentPods(f, d)

		ginkgo.By("Setting up a VPA CRD")
		SetupVPAForNHamsters(f, 2, "50m", vpa_types.UpdateModeAuto, hamsterTargetRef)

		// Min CPU from limit range is 200m per pod, 100m per container and ratio is 3. Min applies both
		// to limit and request so min request is 100m request and 300m limit
		// Min memory limit is 0 and ratio is 2., so min request is 0
		InstallLimitRangeWithMin(f, "200m", "0", apiv1.LimitTypePod)

		ginkgo.By(fmt.Sprintf("Waiting for pods to be evicted, hoping it won't happen, sleep for %s", VpaEvictionTimeout.String()))
		CheckNoPodsEvicted(f, MakePodSet(podList))
	})

	ginkgo.It("does not act on injected sidecars", func() {
		const (
			// TODO(krzysied): Update the image url when the agnhost:2.10 image
			// is promoted to the k8s-e2e-test-images repository.
			agnhostImage  = "gcr.io/k8s-staging-e2e-test-images/agnhost:2.10"
			sidecarParam  = "--sidecar-image=k8s.gcr.io/pause:3.1"
			sidecarName   = "webhook-added-sidecar"
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

		mode := vpa_types.UpdateModeAuto
		hamsterResourceList := apiv1.ResourceList{apiv1.ResourceCPU: ParseQuantityOrDie("100m")}
		sidecarResourceList := apiv1.ResourceList{apiv1.ResourceCPU: ParseQuantityOrDie("5000m")}

		vpaCRD := NewVPA(f, "hamster-vpa", hamsterTargetRef)
		vpaCRD.Spec.UpdatePolicy.UpdateMode = &mode

		vpaCRD.Status.Recommendation = &vpa_types.RecommendedPodResources{
			ContainerRecommendations: []vpa_types.RecommendedContainerResources{
				{
					ContainerName: GetHamsterContainerNameByIndex(0),
					Target:        hamsterResourceList,
					LowerBound:    hamsterResourceList,
					UpperBound:    hamsterResourceList,
				},
				{
					ContainerName: sidecarName,
					Target:        sidecarResourceList,
					LowerBound:    sidecarResourceList,
					UpperBound:    sidecarResourceList,
				},
			},
		}

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

	err := wait.PollImmediate(pollInterval, pollTimeout+pendingDuration, func() (bool, error) {
		var err error
		currentPodList, err := framework_deployment.GetPodsForDeployment(c, deployment)
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

func testEvictsPods(f *framework.Framework, controller *autoscaling.CrossVersionObjectReference) {
	ginkgo.By(fmt.Sprintf("Setting up a hamster %v", controller.Kind))
	setupHamsterController(f, controller.Kind, "100m", "100Mi", defaultHamsterReplicas)
	podList, err := GetHamsterPods(f)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	ginkgo.By("Setting up a VPA CRD")
	SetupVPA(f, "200m", vpa_types.UpdateModeAuto, controller)

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
	return wait.PollImmediate(pollInterval, pollTimeout, func() (bool, error) {
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
	err := testutils.CreateJobWithRetries(f.ClientSet, f.Namespace.Name, job)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	err = framework_job.WaitForAllJobPodsRunning(f.ClientSet, f.Namespace.Name, job.Name, replicas)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
}

func setupHamsterRS(f *framework.Framework, cpu, memory string, replicas int32) {
	rs := newReplicaSet("hamster-rs", f.Namespace.Name, replicas, hamsterLabels, "", "")
	rs.Spec.Template.Spec.Containers[0] = SetupHamsterContainer(cpu, memory)
	err := createReplicaSetWithRetries(f.ClientSet, f.Namespace.Name, rs)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	err = framework_rs.WaitForReadyReplicaSet(f.ClientSet, f.Namespace.Name, rs.Name)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
}

func setupHamsterStateful(f *framework.Framework, cpu, memory string, replicas int32) {
	stateful := framework_ss.NewStatefulSet("hamster-stateful", f.Namespace.Name,
		"hamster-service", replicas, nil, nil, hamsterLabels)

	stateful.Spec.Template.Spec.Containers[0] = SetupHamsterContainer(cpu, memory)
	err := createStatefulSetSetWithRetries(f.ClientSet, f.Namespace.Name, stateful)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	framework_ss.WaitForRunningAndReady(f.ClientSet, *stateful.Spec.Replicas, stateful)
}

func setupPDB(f *framework.Framework, name string, maxUnavailable int) *policyv1beta1.PodDisruptionBudget {
	maxUnavailableIntstr := intstr.FromInt(maxUnavailable)
	pdb := &policyv1beta1.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: policyv1beta1.PodDisruptionBudgetSpec{
			MaxUnavailable: &maxUnavailableIntstr,
			Selector: &metav1.LabelSelector{
				MatchLabels: hamsterLabels,
			},
		},
	}
	_, err := f.ClientSet.PolicyV1beta1().PodDisruptionBudgets(f.Namespace.Name).Create(context.TODO(), pdb, metav1.CreateOptions{})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	return pdb
}

func getCurrentPodSetForDeployment(c clientset.Interface, d *appsv1.Deployment) PodSet {
	podList, err := framework_deployment.GetPodsForDeployment(c, d)
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
		if testutils.IsRetryableAPIError(err) {
			return false, nil
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
		if testutils.IsRetryableAPIError(err) {
			return false, nil
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
