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
	"reflect"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/poc.autoscaling.k8s.io/v1alpha1"
	vpa_clientset "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned"
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
		ns := f.Namespace.Name

		cpuQuantity := ParseQuantityOrDie("100m")
		memoryQuantity := ParseQuantityOrDie("100Mi")

		d := NewHamsterDeploymentWithResources(f, cpuQuantity, memoryQuantity)
		d, err := c.AppsV1().Deployments(ns).Create(d)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		err = framework.WaitForDeploymentComplete(c, d)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		podList, err := framework.GetPodsForDeployment(c, d)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		ginkgo.By("Setting up a VPA CRD")
		config, err := framework.LoadConfig()
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		vpaCRD := NewVPA(f, "hamster-vpa", &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"app": "hamster",
			},
		})

		newCPUQuantity, err := resource.ParseQuantity("200m")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		resourceList := apiv1.ResourceList{apiv1.ResourceCPU: newCPUQuantity}

		vpaCRD.Status.Recommendation = &vpa_types.RecommendedPodResources{
			ContainerRecommendations: []vpa_types.RecommendedContainerResources{{
				ContainerName: "hamster",
				Target:        resourceList,
				LowerBound:    resourceList,
				UpperBound:    resourceList,
			}},
		}

		vpaClientSet := vpa_clientset.NewForConfigOrDie(config)
		vpaClient := vpaClientSet.PocV1alpha1()
		_, err = vpaClient.VerticalPodAutoscalers(ns).Create(vpaCRD)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		ginkgo.By("Waiting for pods to be restarted")

		err = waitForPodSetChangedInDeployment(c, d, podList)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

	})
})

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
		var err error
		currentPodList, err := framework.GetPodsForDeployment(c, deployment)
		if err != nil {
			return false, err
		}

		currentPodSet := makePodSet(currentPodList)

		return !reflect.DeepEqual(initialPodSet, currentPodSet), nil

	})

	if err != nil {
		return fmt.Errorf("Waiting for set of pods changed in %v: %v", deployment.Name, err)
	}
	return nil
}
