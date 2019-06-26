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

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1beta1"
	"k8s.io/kubernetes/test/e2e/framework"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

var _ = AdmissionControllerE2eDescribe("Admission-controller", func() {
	f := framework.NewDefaultFramework("vertical-pod-autoscaling")

	ginkgo.It("starts pods with new recommended request", func() {
		d := NewHamsterDeploymentWithResources(f, ParseQuantityOrDie("100m") /*cpu*/, ParseQuantityOrDie("100Mi") /*memory*/)

		ginkgo.By("Setting up a VPA CRD")
		vpaCRD := NewVPA(f, "hamster-vpa", &metav1.LabelSelector{
			MatchLabels: d.Spec.Template.Labels,
		})
		vpaCRD.Status.Recommendation = &vpa_types.RecommendedPodResources{
			ContainerRecommendations: []vpa_types.RecommendedContainerResources{{
				ContainerName: "hamster",
				Target: apiv1.ResourceList{
					apiv1.ResourceCPU:    ParseQuantityOrDie("250m"),
					apiv1.ResourceMemory: ParseQuantityOrDie("200Mi"),
				},
			}},
		}
		InstallVPA(f, vpaCRD)

		ginkgo.By("Setting up a hamster deployment")
		podList := startDeploymentPods(f, d)

		// Originally Pods had 100m CPU, 100Mi of memory, but admission controller
		// should change it to recommended 250m CPU and 200Mi of memory.
		for _, pod := range podList.Items {
			gomega.Expect(pod.Spec.Containers[0].Resources.Requests[apiv1.ResourceCPU]).To(gomega.Equal(ParseQuantityOrDie("250m")))
			gomega.Expect(pod.Spec.Containers[0].Resources.Requests[apiv1.ResourceMemory]).To(gomega.Equal(ParseQuantityOrDie("200Mi")))
		}
	})

	ginkgo.It("doesn't block patches", func() {
		d := NewHamsterDeploymentWithResources(f, ParseQuantityOrDie("100m") /*cpu*/, ParseQuantityOrDie("100Mi") /*memory*/)

		ginkgo.By("Setting up a VPA CRD")
		vpaCRD := NewVPA(f, "hamster-vpa", &metav1.LabelSelector{
			MatchLabels: d.Spec.Template.Labels,
		})
		vpaCRD.Status.Recommendation = &vpa_types.RecommendedPodResources{
			ContainerRecommendations: []vpa_types.RecommendedContainerResources{{
				ContainerName: "hamster",
				Target: apiv1.ResourceList{
					apiv1.ResourceCPU:    ParseQuantityOrDie("250m"),
					apiv1.ResourceMemory: ParseQuantityOrDie("200Mi"),
				},
			}},
		}
		InstallVPA(f, vpaCRD)

		ginkgo.By("Setting up a hamster deployment")
		podList := startDeploymentPods(f, d)

		// Originally Pods had 100m CPU, 100Mi of memory, but admission controller
		// should change it to recommended 250m CPU and 200Mi of memory.
		for _, pod := range podList.Items {
			gomega.Expect(pod.Spec.Containers[0].Resources.Requests[apiv1.ResourceCPU]).To(gomega.Equal(ParseQuantityOrDie("250m")))
			gomega.Expect(pod.Spec.Containers[0].Resources.Requests[apiv1.ResourceMemory]).To(gomega.Equal(ParseQuantityOrDie("200Mi")))
		}

		ginkgo.By("Modifying recommendation.")
		PatchVpaRecommendation(f, vpaCRD, &vpa_types.RecommendedPodResources{
			ContainerRecommendations: []vpa_types.RecommendedContainerResources{{
				ContainerName: "hamster",
				Target: apiv1.ResourceList{
					apiv1.ResourceCPU:    ParseQuantityOrDie("100m"),
					apiv1.ResourceMemory: ParseQuantityOrDie("100Mi"),
				},
			}},
		})

		podName := podList.Items[0].Name
		ginkgo.By(fmt.Sprintf("Modifying pod %v.", podName))
		AnnotatePod(f, podName, "someAnnotation", "someValue")
	})

	ginkgo.It("keeps limits equal to request", func() {
		d := NewHamsterDeploymentWithGuaranteedResources(f, ParseQuantityOrDie("100m") /*cpu*/, ParseQuantityOrDie("100Mi") /*memory*/)

		ginkgo.By("Setting up a VPA CRD")
		vpaCRD := NewVPA(f, "hamster-vpa", &metav1.LabelSelector{
			MatchLabels: d.Spec.Template.Labels,
		})
		vpaCRD.Status.Recommendation = &vpa_types.RecommendedPodResources{
			ContainerRecommendations: []vpa_types.RecommendedContainerResources{{
				ContainerName: "hamster",
				Target: apiv1.ResourceList{
					apiv1.ResourceCPU:    ParseQuantityOrDie("250m"),
					apiv1.ResourceMemory: ParseQuantityOrDie("200Mi"),
				},
			}},
		}
		InstallVPA(f, vpaCRD)

		ginkgo.By("Setting up a hamster deployment")
		podList := startDeploymentPods(f, d)

		// Originally Pods had 100m CPU, 100Mi of memory, but admission controller
		// should change it to 250m CPU and 200Mi of memory. Limits and requests should stay equal.
		for _, pod := range podList.Items {
			gomega.Expect(pod.Spec.Containers[0].Resources.Requests[apiv1.ResourceCPU]).To(gomega.Equal(ParseQuantityOrDie("250m")))
			gomega.Expect(pod.Spec.Containers[0].Resources.Requests[apiv1.ResourceMemory]).To(gomega.Equal(ParseQuantityOrDie("200Mi")))
			gomega.Expect(pod.Spec.Containers[0].Resources.Limits[apiv1.ResourceCPU]).To(gomega.Equal(ParseQuantityOrDie("250m")))
			gomega.Expect(pod.Spec.Containers[0].Resources.Limits[apiv1.ResourceMemory]).To(gomega.Equal(ParseQuantityOrDie("200Mi")))
		}
	})

	ginkgo.It("keeps limits to request ratio constant", func() {
		d := NewHamsterDeploymentWithResourcesAndLimits(f,
			ParseQuantityOrDie("100m") /*cpu request*/, ParseQuantityOrDie("100Mi"), /*memory request*/
			ParseQuantityOrDie("150m") /*cpu limit*/, ParseQuantityOrDie("200Mi") /*memory limit*/)

		ginkgo.By("Setting up a VPA CRD")
		vpaCRD := NewVPA(f, "hamster-vpa", &metav1.LabelSelector{
			MatchLabels: d.Spec.Template.Labels,
		})
		vpaCRD.Status.Recommendation = &vpa_types.RecommendedPodResources{
			ContainerRecommendations: []vpa_types.RecommendedContainerResources{{
				ContainerName: "hamster",
				Target: apiv1.ResourceList{
					apiv1.ResourceCPU:    ParseQuantityOrDie("250m"),
					apiv1.ResourceMemory: ParseQuantityOrDie("200Mi"),
				},
			}},
		}
		InstallVPA(f, vpaCRD)

		ginkgo.By("Setting up a hamster deployment")
		podList := startDeploymentPods(f, d)

		// Originally Pods had 100m CPU, 100Mi of memory, but admission controller
		// should change it to 250m CPU and 200Mi of memory. Limits to request ratio should stay unchanged.
		for _, pod := range podList.Items {
			gomega.Expect(*pod.Spec.Containers[0].Resources.Requests.Cpu()).To(gomega.Equal(ParseQuantityOrDie("250m")))
			gomega.Expect(*pod.Spec.Containers[0].Resources.Requests.Memory()).To(gomega.Equal(ParseQuantityOrDie("200Mi")))
			gomega.Expect(float64(pod.Spec.Containers[0].Resources.Limits.Cpu().MilliValue()) / float64(pod.Spec.Containers[0].Resources.Requests.Cpu().MilliValue())).To(gomega.BeNumerically("~", 1.5))
			gomega.Expect(float64(pod.Spec.Containers[0].Resources.Limits.Memory().Value()) / float64(pod.Spec.Containers[0].Resources.Requests.Memory().Value())).To(gomega.BeNumerically("~", 2.))
		}
	})

	ginkgo.It("caps request according to max limit set in LimitRange", func() {
		d := NewHamsterDeploymentWithResourcesAndLimits(f,
			ParseQuantityOrDie("100m") /*cpu request*/, ParseQuantityOrDie("100Mi"), /*memory request*/
			ParseQuantityOrDie("150m") /*cpu limit*/, ParseQuantityOrDie("200Mi") /*memory limit*/)

		ginkgo.By("Setting up a VPA CRD")
		vpaCRD := NewVPA(f, "hamster-vpa", &metav1.LabelSelector{
			MatchLabels: d.Spec.Template.Labels,
		})
		vpaCRD.Status.Recommendation = &vpa_types.RecommendedPodResources{
			ContainerRecommendations: []vpa_types.RecommendedContainerResources{{
				ContainerName: "hamster",
				Target: apiv1.ResourceList{
					apiv1.ResourceCPU:    ParseQuantityOrDie("250m"),
					apiv1.ResourceMemory: ParseQuantityOrDie("200Mi"),
				},
			}},
		}
		InstallVPA(f, vpaCRD)

		// Max CPU limit is 300m and ratio is 1.5, so max request is 200m, while
		// recommendation is 250m
		// Max memory limit is 1Gi and ratio is 2., so max request is 0.5Gi
		InstallLimitRangeWithMax(f, "300m", "1Gi")

		ginkgo.By("Setting up a hamster deployment")
		podList := startDeploymentPods(f, d)

		// Originally Pods had 100m CPU, 100Mi of memory, but admission controller
		// should change it to 200m CPU (as this is the recommendation
		// capped according to max limit in LimitRange) and 200Mi of memory,
		// which is uncapped. Limit to request ratio should stay unchanged.
		for _, pod := range podList.Items {
			gomega.Expect(*pod.Spec.Containers[0].Resources.Requests.Cpu()).To(gomega.Equal(ParseQuantityOrDie("200m")))
			gomega.Expect(*pod.Spec.Containers[0].Resources.Requests.Memory()).To(gomega.Equal(ParseQuantityOrDie("200Mi")))
			gomega.Expect(pod.Spec.Containers[0].Resources.Limits.Cpu().MilliValue()).To(gomega.BeNumerically("<=", 300))
			gomega.Expect(pod.Spec.Containers[0].Resources.Limits.Memory().Value()).To(gomega.BeNumerically("<=", 1024*1024*1024))
			gomega.Expect(float64(pod.Spec.Containers[0].Resources.Limits.Cpu().MilliValue()) / float64(pod.Spec.Containers[0].Resources.Requests.Cpu().MilliValue())).To(gomega.BeNumerically("~", 1.5))
			gomega.Expect(float64(pod.Spec.Containers[0].Resources.Limits.Memory().Value()) / float64(pod.Spec.Containers[0].Resources.Requests.Memory().Value())).To(gomega.BeNumerically("~", 2.))
		}
	})

	ginkgo.It("raises request according to min limit set in LimitRange", func() {
		d := NewHamsterDeploymentWithResourcesAndLimits(f,
			ParseQuantityOrDie("100m") /*cpu request*/, ParseQuantityOrDie("200Mi"), /*memory request*/
			ParseQuantityOrDie("150m") /*cpu limit*/, ParseQuantityOrDie("400Mi") /*memory limit*/)

		ginkgo.By("Setting up a VPA CRD")
		vpaCRD := NewVPA(f, "hamster-vpa", &metav1.LabelSelector{
			MatchLabels: d.Spec.Template.Labels,
		})
		vpaCRD.Status.Recommendation = &vpa_types.RecommendedPodResources{
			ContainerRecommendations: []vpa_types.RecommendedContainerResources{{
				ContainerName: "hamster",
				Target: apiv1.ResourceList{
					apiv1.ResourceCPU:    ParseQuantityOrDie("250m"),
					apiv1.ResourceMemory: ParseQuantityOrDie("100Mi"), // memory is downscaled
				},
			}},
		}
		InstallVPA(f, vpaCRD)

		// Min CPU limit is 75m and ratio is 1.5, so min request is 50m
		// Min memory limit is 250Mi and ratio is 2., so min request is 125Mi, while
		// recommendation is 100Mi.
		InstallLimitRangeWithMin(f, "75m", "250Mi")

		ginkgo.By("Setting up a hamster deployment")
		podList := startDeploymentPods(f, d)

		// Originally Pods had 100m CPU, 200Mi of memory, but admission controller
		// should change it to 250m CPU and 125Mi of memory, since this is the lowest
		// request that limitrange allows.
		// Limit to request ratio should stay unchanged.
		for _, pod := range podList.Items {
			gomega.Expect(*pod.Spec.Containers[0].Resources.Requests.Cpu()).To(gomega.Equal(ParseQuantityOrDie("250m")))
			gomega.Expect(*pod.Spec.Containers[0].Resources.Requests.Memory()).To(gomega.Equal(ParseQuantityOrDie("125Mi")))
			gomega.Expect(pod.Spec.Containers[0].Resources.Limits.Cpu().MilliValue()).To(gomega.BeNumerically(">=", 75))
			gomega.Expect(pod.Spec.Containers[0].Resources.Limits.Memory().Value()).To(gomega.BeNumerically(">=", 250*1024*1024))
			gomega.Expect(float64(pod.Spec.Containers[0].Resources.Limits.Cpu().MilliValue()) / float64(pod.Spec.Containers[0].Resources.Requests.Cpu().MilliValue())).To(gomega.BeNumerically("~", 1.5))
			gomega.Expect(float64(pod.Spec.Containers[0].Resources.Limits.Memory().Value()) / float64(pod.Spec.Containers[0].Resources.Requests.Memory().Value())).To(gomega.BeNumerically("~", 2.))
		}
	})

	ginkgo.It("caps request to max set in VPA", func() {
		d := NewHamsterDeploymentWithResources(f, ParseQuantityOrDie("100m") /*cpu*/, ParseQuantityOrDie("100Mi") /*memory*/)

		ginkgo.By("Setting up a VPA CRD")
		vpaCRD := NewVPA(f, "hamster-vpa", &metav1.LabelSelector{
			MatchLabels: d.Spec.Template.Labels,
		})
		vpaCRD.Status.Recommendation = &vpa_types.RecommendedPodResources{
			ContainerRecommendations: []vpa_types.RecommendedContainerResources{{
				ContainerName: "hamster",
				Target: apiv1.ResourceList{
					apiv1.ResourceCPU:    ParseQuantityOrDie("250m"),
					apiv1.ResourceMemory: ParseQuantityOrDie("200Mi"),
				},
			}},
		}
		vpaCRD.Spec.ResourcePolicy = &vpa_types.PodResourcePolicy{
			ContainerPolicies: []vpa_types.ContainerResourcePolicy{{
				ContainerName: "hamster",
				MaxAllowed: apiv1.ResourceList{
					apiv1.ResourceCPU:    ParseQuantityOrDie("233m"),
					apiv1.ResourceMemory: ParseQuantityOrDie("150Mi"),
				},
			}},
		}
		InstallVPA(f, vpaCRD)

		ginkgo.By("Setting up a hamster deployment")
		podList := startDeploymentPods(f, d)

		// Originally Pods had 100m CPU, 100Mi of memory, but admission controller
		// should change it to 233m CPU and 150Mi of memory (as this is the recommendation
		// capped to max specified in VPA)
		for _, pod := range podList.Items {
			gomega.Expect(pod.Spec.Containers[0].Resources.Requests[apiv1.ResourceCPU]).To(gomega.Equal(ParseQuantityOrDie("233m")))
			gomega.Expect(pod.Spec.Containers[0].Resources.Requests[apiv1.ResourceMemory]).To(gomega.Equal(ParseQuantityOrDie("150Mi")))
		}
	})

	ginkgo.It("raises request to min set in VPA", func() {
		d := NewHamsterDeploymentWithResources(f, ParseQuantityOrDie("100m") /*cpu*/, ParseQuantityOrDie("100Mi") /*memory*/)

		ginkgo.By("Setting up a VPA CRD")
		vpaCRD := NewVPA(f, "hamster-vpa", &metav1.LabelSelector{
			MatchLabels: d.Spec.Template.Labels,
		})
		vpaCRD.Status.Recommendation = &vpa_types.RecommendedPodResources{
			ContainerRecommendations: []vpa_types.RecommendedContainerResources{{
				ContainerName: "hamster",
				Target: apiv1.ResourceList{
					apiv1.ResourceCPU:    ParseQuantityOrDie("50m"),
					apiv1.ResourceMemory: ParseQuantityOrDie("60Mi"),
				},
			}},
		}
		vpaCRD.Spec.ResourcePolicy = &vpa_types.PodResourcePolicy{
			ContainerPolicies: []vpa_types.ContainerResourcePolicy{{
				ContainerName: "hamster",
				MinAllowed: apiv1.ResourceList{
					apiv1.ResourceCPU:    ParseQuantityOrDie("90m"),
					apiv1.ResourceMemory: ParseQuantityOrDie("80Mi"),
				},
			}},
		}
		InstallVPA(f, vpaCRD)

		ginkgo.By("Setting up a hamster deployment")
		podList := startDeploymentPods(f, d)

		// Originally Pods had 100m CPU, 100Mi of memory, but admission controller
		// should change it to recommended 90m CPU and 800Mi of memory (as this the
		// recommendation raised to min specified in VPA)
		for _, pod := range podList.Items {
			gomega.Expect(pod.Spec.Containers[0].Resources.Requests[apiv1.ResourceCPU]).To(gomega.Equal(ParseQuantityOrDie("90m")))
			gomega.Expect(pod.Spec.Containers[0].Resources.Requests[apiv1.ResourceMemory]).To(gomega.Equal(ParseQuantityOrDie("80Mi")))
		}
	})

	ginkgo.It("leaves users request when no recommendation", func() {
		d := NewHamsterDeploymentWithResources(f, ParseQuantityOrDie("100m") /*cpu*/, ParseQuantityOrDie("100Mi") /*memory*/)

		ginkgo.By("Setting up a VPA CRD")
		vpaCRD := NewVPA(f, "hamster-vpa", &metav1.LabelSelector{
			MatchLabels: d.Spec.Template.Labels,
		})
		InstallVPA(f, vpaCRD)

		ginkgo.By("Setting up a hamster deployment")
		podList := startDeploymentPods(f, d)

		// VPA has no recommendation, so user's request is passed through
		for _, pod := range podList.Items {
			gomega.Expect(pod.Spec.Containers[0].Resources.Requests[apiv1.ResourceCPU]).To(gomega.Equal(ParseQuantityOrDie("100m")))
			gomega.Expect(pod.Spec.Containers[0].Resources.Requests[apiv1.ResourceMemory]).To(gomega.Equal(ParseQuantityOrDie("100Mi")))
		}
	})

	ginkgo.It("passes empty request when no recommendation and no user-specified request", func() {
		d := NewHamsterDeployment(f)

		ginkgo.By("Setting up a VPA CRD")
		vpaCRD := NewVPA(f, "hamster-vpa", &metav1.LabelSelector{
			MatchLabels: d.Spec.Template.Labels,
		})
		InstallVPA(f, vpaCRD)

		ginkgo.By("Setting up a hamster deployment")
		podList := startDeploymentPods(f, d)

		// VPA has no recommendation, deployment has no request specified
		for _, pod := range podList.Items {
			gomega.Expect(pod.Spec.Containers[0].Resources.Requests).To(gomega.BeEmpty())
		}
	})

})

func startDeploymentPods(f *framework.Framework, deployment *appsv1.Deployment) *apiv1.PodList {
	c, ns := f.ClientSet, f.Namespace.Name
	deployment, err := c.AppsV1().Deployments(ns).Create(deployment)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	err = framework.WaitForDeploymentComplete(c, deployment)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	podList, err := framework.GetPodsForDeployment(c, deployment)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	return podList
}
