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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1beta1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/test/e2e/framework"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

var _ = UpdaterE2eDescribe("Updater", func() {
	f := framework.NewDefaultFramework("vertical-pod-autoscaling")

	ginkgo.It("restarts pods", func() {

		ginkgo.By("Setting up a hamster deployment")
		c := f.ClientSet
		d, podList := setupHamsterDeployment(f, "100m", "100Mi", nil)

		ginkgo.By("Setting up a VPA CRD")
		setupVPA(f, "200m")

		ginkgo.By("Waiting for pods to be restarted")

		err := waitForPodSetChangedInDeployment(c, d, podList)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

	})

	ginkgo.It("observes pod disruption budget", func() {

		ginkgo.By("Setting up a hamster deployment")
		c := f.ClientSet
		ns := f.Namespace.Name

		replicas := int32(10)
		d, podList := setupHamsterDeployment(f, "10m", "10Mi", &replicas)
		podSet := makePodSet(podList)

		ginkgo.By("Setting up prohibitive PDB for hamster deployment")
		pdb := setupPDB(f, "hamster-pdb", 0 /* maxUnavailable */)

		ginkgo.By("Setting up a VPA CRD")
		setupVPA(f, "25m")

		ginkgo.By(fmt.Sprintf("Waiting for pods to be restarted, hoping it won't happen, sleep for %s", VpaEvictionTimeout.String()))
		time.Sleep(VpaEvictionTimeout)
		restarted, err := isPodSetChanged(c, d, podSet)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
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
		restartedCount, err := getPodSetChanges(c, d, podSet)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(restartedCount >= permissiveMaxUnavailable).To(gomega.BeTrue())
	})
})

func setupHamsterDeployment(f *framework.Framework, cpu, memory string, replicas *int32) (*appsv1.Deployment, *apiv1.PodList) {
	cpuQuantity := ParseQuantityOrDie(cpu)
	memoryQuantity := ParseQuantityOrDie(memory)

	d := NewHamsterDeploymentWithResources(f, cpuQuantity, memoryQuantity)
	if replicas != nil {
		d.Spec.Replicas = replicas
	}
	d, err := f.ClientSet.AppsV1().Deployments(f.Namespace.Name).Create(d)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	err = framework.WaitForDeploymentComplete(f.ClientSet, d)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	podList, err := framework.GetPodsForDeployment(f.ClientSet, d)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	return d, podList
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
				MatchLabels: map[string]string{"app": "hamster"},
			},
		},
	}
	_, err := f.ClientSet.PolicyV1beta1().PodDisruptionBudgets(f.Namespace.Name).Create(pdb)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	return pdb
}

func makePodSet(pods *apiv1.PodList) map[string]bool {
	result := make(map[string]bool)
	for _, p := range pods.Items {
		result[p.Name] = true
	}
	return result
}

func waitForPodSetChangedInDeployment(c clientset.Interface, deployment *appsv1.Deployment, podList *apiv1.PodList) error {
	initialPodSet := makePodSet(podList)

	err := wait.PollImmediate(pollInterval, pollTimeout, func() (bool, error) {
		return isPodSetChanged(c, deployment, initialPodSet)

	})

	if err != nil {
		return fmt.Errorf("Waiting for set of pods changed in %v: %v", deployment.Name, err)
	}
	return nil
}

func isPodSetChanged(c clientset.Interface, deployment *appsv1.Deployment, initialPodSet map[string]bool) (bool, error) {
	count, err := getPodSetChanges(c, deployment, initialPodSet)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func getPodSetChanges(c clientset.Interface, deployment *appsv1.Deployment, initialPodSet map[string]bool) (int, error) {
	currentPodList, err := framework.GetPodsForDeployment(c, deployment)
	if err != nil {
		return 0, err
	}
	currentPodSet := makePodSet(currentPodList)
	diffs := 0
	for name, inInitial := range initialPodSet {
		inCurrent := currentPodSet[name]
		if inInitial && !inCurrent {
			diffs += 1
		}
	}
	return diffs, nil
}
