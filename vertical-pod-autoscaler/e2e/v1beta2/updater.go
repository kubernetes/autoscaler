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
	"time"

	appsv1 "k8s.io/api/apps/v1"
	autoscaling "k8s.io/api/autoscaling/v1"
	apiv1 "k8s.io/api/core/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1beta2"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/test/e2e/framework"
	testutils "k8s.io/kubernetes/test/utils"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

var _ = UpdaterE2eDescribe("Updater", func() {
	f := framework.NewDefaultFramework("vertical-pod-autoscaling")

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
		err = c.PolicyV1beta1().PodDisruptionBudgets(ns).Delete(pdb.Name, &metav1.DeleteOptions{})
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
		SetupVPAForTwoHamsters(f, "200m", vpa_types.UpdateModeAuto, hamsterTargetRef)

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
		SetupVPAForTwoHamsters(f, "50m", vpa_types.UpdateModeAuto, hamsterTargetRef)

		// Min CPU from limit range is 200m per pod, 100m per container and ratio is 3. Min applies both
		// to limit and request so min request is 100m request and 300m limit
		// Min memory limit is 0 and ratio is 2., so min request is 0
		InstallLimitRangeWithMin(f, "200m", "0", apiv1.LimitTypePod)

		ginkgo.By(fmt.Sprintf("Waiting for pods to be evicted, hoping it won't happen, sleep for %s", VpaEvictionTimeout.String()))
		CheckNoPodsEvicted(f, MakePodSet(podList))
	})
})

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
	hamsterContainer := setupHamsterContainer(cpu, memory)
	rc := framework.RcByNameContainer("hamster-rc", replicas, "k8s.gcr.io/ubuntu-slim:0.1",
		hamsterLabels, hamsterContainer, nil)

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
	job := framework.NewTestJob("notTerminate", "hamster-job", apiv1.RestartPolicyOnFailure,
		replicas, replicas, nil, 10)
	job.Spec.Template.Spec.Containers[0] = setupHamsterContainer(cpu, memory)
	for label, value := range hamsterLabels {
		job.Spec.Template.Labels[label] = value
	}
	err := testutils.CreateJobWithRetries(f.ClientSet, f.Namespace.Name, job)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	err = framework.WaitForAllJobPodsRunning(f.ClientSet, f.Namespace.Name, job.Name, replicas)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
}

func setupHamsterRS(f *framework.Framework, cpu, memory string, replicas int32) {
	rs := framework.NewReplicaSet("hamster-rs", f.Namespace.Name, replicas,
		hamsterLabels, "", "")
	rs.Spec.Template.Spec.Containers[0] = setupHamsterContainer(cpu, memory)
	err := createReplicaSetWithRetries(f.ClientSet, f.Namespace.Name, rs)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	err = framework.WaitForReadyReplicaSet(f.ClientSet, f.Namespace.Name, rs.Name)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
}

func setupHamsterStateful(f *framework.Framework, cpu, memory string, replicas int32) {
	stateful := framework.NewStatefulSet("hamster-stateful", f.Namespace.Name,
		"hamster-service", replicas, nil, nil, hamsterLabels)

	stateful.Spec.Template.Spec.Containers[0] = setupHamsterContainer(cpu, memory)
	err := createStatefulSetSetWithRetries(f.ClientSet, f.Namespace.Name, stateful)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	tester := framework.NewStatefulSetTester(f.ClientSet)
	tester.WaitForRunningAndReady(*stateful.Spec.Replicas, stateful)
}

func setupHamsterContainer(cpu, memory string) apiv1.Container {
	cpuQuantity := ParseQuantityOrDie(cpu)
	memoryQuantity := ParseQuantityOrDie(memory)

	return apiv1.Container{
		Name:  "hamster",
		Image: "k8s.gcr.io/ubuntu-slim:0.1",
		Resources: apiv1.ResourceRequirements{
			Requests: apiv1.ResourceList{
				apiv1.ResourceCPU:    cpuQuantity,
				apiv1.ResourceMemory: memoryQuantity,
			},
		},
		Command: []string{"/bin/sh"},
		Args:    []string{"-c", "while true; do sleep 10 ; done"},
	}
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
	_, err := f.ClientSet.PolicyV1beta1().PodDisruptionBudgets(f.Namespace.Name).Create(pdb)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	return pdb
}

func getCurrentPodSetForDeployment(c clientset.Interface, d *appsv1.Deployment) PodSet {
	podList, err := framework.GetPodsForDeployment(c, d)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	return MakePodSet(podList)
}

func createReplicaSetWithRetries(c clientset.Interface, namespace string, obj *appsv1.ReplicaSet) error {
	if obj == nil {
		return fmt.Errorf("object provided to create is empty")
	}
	createFunc := func() (bool, error) {
		_, err := c.AppsV1().ReplicaSets(namespace).Create(obj)
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
		_, err := c.AppsV1().StatefulSets(namespace).Create(obj)
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
