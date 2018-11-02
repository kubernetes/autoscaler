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
	apiv1 "k8s.io/api/core/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1beta1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/test/e2e/framework"
	testutils "k8s.io/kubernetes/test/utils"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

var _ = UpdaterE2eDescribe("Updater", func() {
	f := framework.NewDefaultFramework("vertical-pod-autoscaling")

	ginkgo.It("restarts pods in a Deployment", func() {
		testRestartsPods(f, "Deployment")
	})

	ginkgo.It("restarts pods in a Replication Controller", func() {
		testRestartsPods(f, "ReplicationController")
	})

	ginkgo.It("restarts pods in a Job", func() {
		testRestartsPods(f, "Job")
	})

	ginkgo.It("restarts pods in a ReplicaSet", func() {
		testRestartsPods(f, "ReplicaSet")
	})

	ginkgo.It("restarts pods in a StatefulSet", func() {
		testRestartsPods(f, "StatefulSet")
	})

	ginkgo.It("observes pod disruption budget", func() {

		ginkgo.By("Setting up a hamster deployment")
		c := f.ClientSet
		ns := f.Namespace.Name

		setupHamsterDeployment(f, "10m", "10Mi", 10)
		podList, err := getHamsterPods(f)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		podSet := makePodSet(podList)

		ginkgo.By("Setting up prohibitive PDB for hamster deployment")
		pdb := setupPDB(f, "hamster-pdb", 0 /* maxUnavailable */)

		ginkgo.By("Setting up a VPA CRD")
		setupVPA(f, "25m")

		ginkgo.By(fmt.Sprintf("Waiting for pods to be restarted, hoping it won't happen, sleep for %s", VpaEvictionTimeout.String()))
		time.Sleep(VpaEvictionTimeout)
		currentPodList, err := getHamsterPods(f)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		restarted := werePodsRestarted(c, makePodSet(currentPodList), podSet)
		gomega.Expect(restarted).To(gomega.BeFalse())

		ginkgo.By("Updating the PDB to allow for multiple pods to be evicted")
		// We will check that 7 replicas are evicted in 3 minutes, which translates
		// to 3 updater loops. This gives us relatively good confidence that updater
		// evicts more than one pod in a loop if PDB allows it.
		permissiveMaxUnavailable := 7
		// Creating new PDB and removing old one, since PDBs are immutable at the moment
		setupPDB(f, "hamster-pdb-2", permissiveMaxUnavailable)
		err = c.PolicyV1beta1().PodDisruptionBudgets(ns).Delete(pdb.Name, &metav1.DeleteOptions{})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		ginkgo.By(fmt.Sprintf("Waiting for pods to be restarted, sleep for %s", VpaEvictionTimeout.String()))
		time.Sleep(VpaEvictionTimeout)
		ginkgo.By("Checking enough pods were restarted.")
		currentPodList, err = getHamsterPods(f)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		restartedCount := getRestartedPodsCount(c, makePodSet(currentPodList), podSet)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(restartedCount >= permissiveMaxUnavailable).To(gomega.BeTrue())
	})
})

func testRestartsPods(f *framework.Framework, controllerKind string) {
	ginkgo.By(fmt.Sprintf("Setting up a hamster %v", controllerKind))
	setupHamsterController(f, controllerKind, "100m", "100Mi", defaultHamsterReplicas)
	podList, err := getHamsterPods(f)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	ginkgo.By("Setting up a VPA CRD")
	setupVPA(f, "200m")

	ginkgo.By("Waiting for pods to be restarted")
	err = waitForPodSetChanged(f, podList)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
}

func setupHamsterController(f *framework.Framework, controllerKind, cpu, memory string, replicas int32) *apiv1.PodList {
	switch controllerKind {
	case "Deployment":
		setupHamsterDeployment(f, cpu, memory, replicas)
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
	pods, err := getHamsterPods(f)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	return pods
}

func getHamsterPods(f *framework.Framework) (*apiv1.PodList, error) {
	label := labels.SelectorFromSet(labels.Set(hamsterLabels))
	selector := fields.ParseSelectorOrDie("status.phase!=" + string(apiv1.PodSucceeded) +
		",status.phase!=" + string(apiv1.PodFailed))
	options := metav1.ListOptions{LabelSelector: label.String(), FieldSelector: selector.String()}
	return f.ClientSet.CoreV1().Pods(f.Namespace.Name).List(options)
}

func setupHamsterDeployment(f *framework.Framework, cpu, memory string, replicas int32) {
	cpuQuantity := ParseQuantityOrDie(cpu)
	memoryQuantity := ParseQuantityOrDie(memory)

	d := NewHamsterDeploymentWithResources(f, cpuQuantity, memoryQuantity)
	d.Spec.Replicas = &replicas
	d, err := f.ClientSet.AppsV1().Deployments(f.Namespace.Name).Create(d)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	err = framework.WaitForDeploymentComplete(f.ClientSet, d)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
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
		podList, err := getHamsterPods(f)
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

func setupVPA(f *framework.Framework, cpu string) {
	vpaCRD := NewVPA(f, "hamster-vpa", &metav1.LabelSelector{
		MatchLabels: map[string]string{
			"app": "hamster",
		},
	})

	cpuQuantity := ParseQuantityOrDie(cpu)
	resourceList := apiv1.ResourceList{apiv1.ResourceCPU: cpuQuantity}

	vpaCRD.Status.Recommendation = &vpa_types.RecommendedPodResources{
		ContainerRecommendations: []vpa_types.RecommendedContainerResources{{
			ContainerName: "hamster",
			Target:        resourceList,
			LowerBound:    resourceList,
			UpperBound:    resourceList,
		}},
	}
	InstallVPA(f, vpaCRD)
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

type podSet map[string]types.UID

func makePodSet(pods *apiv1.PodList) podSet {
	result := make(podSet)
	if pods == nil {
		return result
	}
	for _, p := range pods.Items {
		result[p.Name] = p.UID
	}
	return result
}

func getCurrentPodSetForDeployment(c clientset.Interface, d *appsv1.Deployment) podSet {
	podList, err := framework.GetPodsForDeployment(c, d)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	return makePodSet(podList)
}

func waitForPodSetChanged(f *framework.Framework, podList *apiv1.PodList) error {
	initialPodSet := makePodSet(podList)

	err := wait.PollImmediate(pollInterval, pollTimeout, func() (bool, error) {
		currentPodList, err := getHamsterPods(f)
		if err != nil {
			return false, err
		}
		currentPodSet := makePodSet(currentPodList)
		return werePodsRestarted(f.ClientSet, currentPodSet, initialPodSet), nil
	})

	if err != nil {
		return fmt.Errorf("Waiting for set of pods changed: %v", err)
	}
	return nil
}

func werePodsRestarted(c clientset.Interface, currentPodSet podSet, initialPodSet podSet) bool {
	return getRestartedPodsCount(c, currentPodSet, initialPodSet) > 0
}

func getRestartedPodsCount(c clientset.Interface, currentPodSet podSet, initialPodSet podSet) int {
	diffs := 0
	for name, initialUID := range initialPodSet {
		currentUID, inCurrent := currentPodSet[name]
		if !inCurrent {
			diffs += 1
		} else if initialUID != currentUID {
			diffs += 1
		}
	}
	return diffs
}

func createReplicaSetWithRetries(c clientset.Interface, namespace string, obj *appsv1.ReplicaSet) error {
	if obj == nil {
		return fmt.Errorf("Object provided to create is empty")
	}
	createFunc := func() (bool, error) {
		_, err := c.AppsV1().ReplicaSets(namespace).Create(obj)
		if err == nil || apierrs.IsAlreadyExists(err) {
			return true, nil
		}
		if testutils.IsRetryableAPIError(err) {
			return false, nil
		}
		return false, fmt.Errorf("Failed to create object with non-retriable error: %v", err)
	}
	return testutils.RetryWithExponentialBackOff(createFunc)
}

func createStatefulSetSetWithRetries(c clientset.Interface, namespace string, obj *appsv1.StatefulSet) error {
	if obj == nil {
		return fmt.Errorf("Object provided to create is empty")
	}
	createFunc := func() (bool, error) {
		_, err := c.AppsV1().StatefulSets(namespace).Create(obj)
		if err == nil || apierrs.IsAlreadyExists(err) {
			return true, nil
		}
		if testutils.IsRetryableAPIError(err) {
			return false, nil
		}
		return false, fmt.Errorf("Failed to create object with non-retriable error: %v", err)
	}
	return testutils.RetryWithExponentialBackOff(createFunc)
}
