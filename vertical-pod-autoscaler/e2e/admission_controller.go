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
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/poc.autoscaling.k8s.io/v1alpha1"
	vpa_clientset "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned"
	"k8s.io/kubernetes/test/e2e/framework"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

var _ = admissionControllerE2eDescribe("Admission-controller", func() {
	f := framework.NewDefaultFramework("vertical-pod-autoscaling")

	ginkgo.It("starts pods with new request", func() {
		c := f.ClientSet
		ns := f.Namespace.Name

		ginkgo.By("Setting up a VPA CRD")
		config, err := framework.LoadConfig()
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		vpaCRD := newVPA(f, "hamster-vpa", &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"app": "hamster",
			},
		})
		newCPUQuantity := parseQuantityOrDie("200m")
		vpaCRD.Status.Recommendation.ContainerRecommendations = []vpa_types.RecommendedContainerResources{
			{
				Name:   "hamster",
				Target: apiv1.ResourceList{apiv1.ResourceCPU: newCPUQuantity},
			},
		}

		vpaClientSet := vpa_clientset.NewForConfigOrDie(config)
		vpaClient := vpaClientSet.PocV1alpha1()
		_, err = vpaClient.VerticalPodAutoscalers(ns).Create(vpaCRD)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		ginkgo.By("Setting up a hamster deployment")

		cpuQuantity := parseQuantityOrDie("100m")
		memoryQuantity := parseQuantityOrDie("100Mi")

		d := hamsterDeployment(f, cpuQuantity, memoryQuantity)
		d, err = c.ExtensionsV1beta1().Deployments(ns).Create(d)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		err = framework.WaitForDeploymentComplete(c, d)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		podList, err := framework.GetPodsForDeployment(c, d)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		// Originally Pods had 100m CPU, but admission controller should change it to recommended 200m CPU
		for _, pod := range podList.Items {
			gomega.Ω(pod.Spec.Containers[0].Resources.Requests[apiv1.ResourceCPU]).Should(gomega.Equal(parseQuantityOrDie("200m")))
		}

	})
})
