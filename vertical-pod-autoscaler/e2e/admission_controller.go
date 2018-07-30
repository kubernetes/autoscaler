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
	extensions "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/poc.autoscaling.k8s.io/v1alpha1"
	"k8s.io/kubernetes/test/e2e/framework"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

var _ = admissionControllerE2eDescribe("Admission-controller", func() {
	f := framework.NewDefaultFramework("vertical-pod-autoscaling")

	ginkgo.It("starts pods with new recommended request", func() {
		d := newHamsterDeploymentWithResources(f, parseQuantityOrDie("100m") /*cpu*/, parseQuantityOrDie("100Mi") /*memory*/)

		ginkgo.By("Setting up a VPA CRD")
		vpaCRD := newVPA(f, "hamster-vpa", &metav1.LabelSelector{
			MatchLabels: d.Spec.Template.Labels,
		})
		vpaCRD.Status.Recommendation = &vpa_types.RecommendedPodResources{
			ContainerRecommendations: []vpa_types.RecommendedContainerResources{{
				ContainerName: "hamster",
				Target: apiv1.ResourceList{
					apiv1.ResourceCPU:    parseQuantityOrDie("250m"),
					apiv1.ResourceMemory: parseQuantityOrDie("200Mi"),
				},
			}},
		}
		installVPA(f, vpaCRD)

		ginkgo.By("Setting up a hamster deployment")
		podList := startDeploymentPods(f, d)

		// Originally Pods had 100m CPU, 100Mi of memory, but admission controller
		// should change it to recommended 250m CPU and 200Mi of memory.
		for _, pod := range podList.Items {
			gomega.Expect(pod.Spec.Containers[0].Resources.Requests[apiv1.ResourceCPU]).To(gomega.Equal(parseQuantityOrDie("250m")))
			gomega.Expect(pod.Spec.Containers[0].Resources.Requests[apiv1.ResourceMemory]).To(gomega.Equal(parseQuantityOrDie("200Mi")))
		}
	})

	ginkgo.It("caps request to limit set by the user", func() {
		d := newHamsterDeploymentWithResources(f, parseQuantityOrDie("100m") /*cpu*/, parseQuantityOrDie("100Mi") /*memory*/)
		d.Spec.Template.Spec.Containers[0].Resources.Limits = apiv1.ResourceList{
			apiv1.ResourceCPU:    parseQuantityOrDie("222m"),
			apiv1.ResourceMemory: parseQuantityOrDie("123Mi"),
		}

		ginkgo.By("Setting up a VPA CRD")
		vpaCRD := newVPA(f, "hamster-vpa", &metav1.LabelSelector{
			MatchLabels: d.Spec.Template.Labels,
		})
		vpaCRD.Status.Recommendation = &vpa_types.RecommendedPodResources{
			ContainerRecommendations: []vpa_types.RecommendedContainerResources{{
				ContainerName: "hamster",
				Target: apiv1.ResourceList{
					apiv1.ResourceCPU:    parseQuantityOrDie("250m"),
					apiv1.ResourceMemory: parseQuantityOrDie("200Mi"),
				},
			}},
		}
		installVPA(f, vpaCRD)

		ginkgo.By("Setting up a hamster deployment")
		podList := startDeploymentPods(f, d)

		// Originally Pods had 100m CPU, 100Mi of memory, but admission controller
		// should change it to 222m CPU and 123Mi of memory (as this is the recommendation
		// capped to the limit set by the user)
		for _, pod := range podList.Items {
			gomega.Expect(pod.Spec.Containers[0].Resources.Requests[apiv1.ResourceCPU]).To(gomega.Equal(parseQuantityOrDie("222m")))
			gomega.Expect(pod.Spec.Containers[0].Resources.Requests[apiv1.ResourceMemory]).To(gomega.Equal(parseQuantityOrDie("123Mi")))
		}
	})

	ginkgo.It("caps request to max set in VPA", func() {
		d := newHamsterDeploymentWithResources(f, parseQuantityOrDie("100m") /*cpu*/, parseQuantityOrDie("100Mi") /*memory*/)

		ginkgo.By("Setting up a VPA CRD")
		vpaCRD := newVPA(f, "hamster-vpa", &metav1.LabelSelector{
			MatchLabels: d.Spec.Template.Labels,
		})
		vpaCRD.Status.Recommendation = &vpa_types.RecommendedPodResources{
			ContainerRecommendations: []vpa_types.RecommendedContainerResources{{
				ContainerName: "hamster",
				Target: apiv1.ResourceList{
					apiv1.ResourceCPU:    parseQuantityOrDie("250m"),
					apiv1.ResourceMemory: parseQuantityOrDie("200Mi"),
				},
			}},
		}
		vpaCRD.Spec.ResourcePolicy = &vpa_types.PodResourcePolicy{
			ContainerPolicies: []vpa_types.ContainerResourcePolicy{{
				ContainerName: "hamster",
				MaxAllowed: apiv1.ResourceList{
					apiv1.ResourceCPU:    parseQuantityOrDie("233m"),
					apiv1.ResourceMemory: parseQuantityOrDie("150Mi"),
				},
			}},
		}
		installVPA(f, vpaCRD)

		ginkgo.By("Setting up a hamster deployment")
		podList := startDeploymentPods(f, d)

		// Originally Pods had 100m CPU, 100Mi of memory, but admission controller
		// should change it to 233m CPU and 150Mi of memory (as this is the recommendation
		// capped to max specified in VPA)
		for _, pod := range podList.Items {
			gomega.Expect(pod.Spec.Containers[0].Resources.Requests[apiv1.ResourceCPU]).To(gomega.Equal(parseQuantityOrDie("233m")))
			gomega.Expect(pod.Spec.Containers[0].Resources.Requests[apiv1.ResourceMemory]).To(gomega.Equal(parseQuantityOrDie("150Mi")))
		}
	})

	ginkgo.It("raises request to min set in VPA", func() {
		d := newHamsterDeploymentWithResources(f, parseQuantityOrDie("100m") /*cpu*/, parseQuantityOrDie("100Mi") /*memory*/)

		ginkgo.By("Setting up a VPA CRD")
		vpaCRD := newVPA(f, "hamster-vpa", &metav1.LabelSelector{
			MatchLabels: d.Spec.Template.Labels,
		})
		vpaCRD.Status.Recommendation = &vpa_types.RecommendedPodResources{
			ContainerRecommendations: []vpa_types.RecommendedContainerResources{{
				ContainerName: "hamster",
				Target: apiv1.ResourceList{
					apiv1.ResourceCPU:    parseQuantityOrDie("50m"),
					apiv1.ResourceMemory: parseQuantityOrDie("60Mi"),
				},
			}},
		}
		vpaCRD.Spec.ResourcePolicy = &vpa_types.PodResourcePolicy{
			ContainerPolicies: []vpa_types.ContainerResourcePolicy{{
				ContainerName: "hamster",
				MinAllowed: apiv1.ResourceList{
					apiv1.ResourceCPU:    parseQuantityOrDie("90m"),
					apiv1.ResourceMemory: parseQuantityOrDie("80Mi"),
				},
			}},
		}
		installVPA(f, vpaCRD)

		ginkgo.By("Setting up a hamster deployment")
		podList := startDeploymentPods(f, d)

		// Originally Pods had 100m CPU, 100Mi of memory, but admission controller
		// should change it to recommended 90m CPU and 800Mi of memory (as this the
		// recommendation raised to min specified in VPA)
		for _, pod := range podList.Items {
			gomega.Expect(pod.Spec.Containers[0].Resources.Requests[apiv1.ResourceCPU]).To(gomega.Equal(parseQuantityOrDie("90m")))
			gomega.Expect(pod.Spec.Containers[0].Resources.Requests[apiv1.ResourceMemory]).To(gomega.Equal(parseQuantityOrDie("80Mi")))
		}
	})

	ginkgo.It("leaves users request when no recommendation", func() {
		d := newHamsterDeploymentWithResources(f, parseQuantityOrDie("100m") /*cpu*/, parseQuantityOrDie("100Mi") /*memory*/)

		ginkgo.By("Setting up a VPA CRD")
		vpaCRD := newVPA(f, "hamster-vpa", &metav1.LabelSelector{
			MatchLabels: d.Spec.Template.Labels,
		})
		installVPA(f, vpaCRD)

		ginkgo.By("Setting up a hamster deployment")
		podList := startDeploymentPods(f, d)

		// VPA has no recommendation, so user's request is passed through
		for _, pod := range podList.Items {
			gomega.Expect(pod.Spec.Containers[0].Resources.Requests[apiv1.ResourceCPU]).To(gomega.Equal(parseQuantityOrDie("100m")))
			gomega.Expect(pod.Spec.Containers[0].Resources.Requests[apiv1.ResourceMemory]).To(gomega.Equal(parseQuantityOrDie("100Mi")))
		}
	})

	ginkgo.It("passes empty request when no recommendation and no user-specified request", func() {
		d := newHamsterDeployment(f)

		ginkgo.By("Setting up a VPA CRD")
		vpaCRD := newVPA(f, "hamster-vpa", &metav1.LabelSelector{
			MatchLabels: d.Spec.Template.Labels,
		})
		installVPA(f, vpaCRD)

		ginkgo.By("Setting up a hamster deployment")
		podList := startDeploymentPods(f, d)

		// VPA has no recommendation, deployment has no request specified
		for _, pod := range podList.Items {
			gomega.Expect(pod.Spec.Containers[0].Resources.Requests).To(gomega.BeEmpty())
		}
	})

})

func startDeploymentPods(f *framework.Framework, deployment *extensions.Deployment) *apiv1.PodList {
	c, ns := f.ClientSet, f.Namespace.Name
	deployment, err := c.ExtensionsV1beta1().Deployments(ns).Create(deployment)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	err = framework.WaitForDeploymentComplete(c, deployment)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	podList, err := framework.GetPodsForDeployment(c, deployment)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	return podList
}
