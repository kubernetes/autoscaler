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
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/kubernetes/test/e2e/framework"
	framework_deployment "k8s.io/kubernetes/test/e2e/framework/deployment"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

var _ = AdmissionControllerE2eDescribe("Admission-controller", func() {
	f := framework.NewDefaultFramework("vertical-pod-autoscaling")

	ginkgo.It("starts pods with new recommended request", func() {
		d := NewHamsterDeploymentWithResources(f, ParseQuantityOrDie("100m") /*cpu*/, ParseQuantityOrDie("100Mi") /*memory*/)

		ginkgo.By("Setting up a VPA CRD")
		vpaCRD := NewVPA(f, "hamster-vpa", hamsterTargetRef)
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
		vpaCRD := NewVPA(f, "hamster-vpa", hamsterTargetRef)
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

		ginkgo.By("Verifying hamster deployment")
		for i, pod := range podList.Items {
			podInfo := fmt.Sprintf("pod at index %d", i)
			cpuDescription := fmt.Sprintf("%s: originally Pods had 100m CPU, admission controller should change it to recommended 250m CPU", podInfo)
			gomega.Expect(pod.Spec.Containers[0].Resources.Requests[apiv1.ResourceCPU]).To(gomega.Equal(ParseQuantityOrDie("250m")), cpuDescription)
			memDescription := fmt.Sprintf("%s: originally Pods had 100Mi of memory, admission controller should change it to recommended 200Mi memory", podInfo)
			gomega.Expect(pod.Spec.Containers[0].Resources.Requests[apiv1.ResourceMemory]).To(gomega.Equal(ParseQuantityOrDie("200Mi")), memDescription)
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
		vpaCRD := NewVPA(f, "hamster-vpa", hamsterTargetRef)
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
		vpaCRD := NewVPA(f, "hamster-vpa", hamsterTargetRef)
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

	ginkgo.It("keeps limits unchanged when container controlled values is requests only", func() {
		d := NewHamsterDeploymentWithResourcesAndLimits(f,
			ParseQuantityOrDie("100m") /*cpu request*/, ParseQuantityOrDie("100Mi"), /*memory request*/
			ParseQuantityOrDie("500m") /*cpu limit*/, ParseQuantityOrDie("500Mi") /*memory limit*/)

		ginkgo.By("Setting up a VPA CRD")
		vpaCRD := NewVPA(f, "hamster-vpa", hamsterTargetRef)
		vpaCRD.Status.Recommendation = &vpa_types.RecommendedPodResources{
			ContainerRecommendations: []vpa_types.RecommendedContainerResources{{
				ContainerName: "hamster",
				Target: apiv1.ResourceList{
					apiv1.ResourceCPU:    ParseQuantityOrDie("250m"),
					apiv1.ResourceMemory: ParseQuantityOrDie("200Mi"),
				},
			}},
		}
		containerControlledValuesRequestsOnly := vpa_types.ContainerControlledValuesRequestsOnly
		vpaCRD.Spec.ResourcePolicy = &vpa_types.PodResourcePolicy{
			ContainerPolicies: []vpa_types.ContainerResourcePolicy{{
				ContainerName:    "hamster",
				ControlledValues: &containerControlledValuesRequestsOnly,
			}},
		}
		InstallVPA(f, vpaCRD)

		ginkgo.By("Setting up a hamster deployment")
		podList := startDeploymentPods(f, d)

		// Originally Pods had 100m CPU, 100Mi of memory, but admission controller
		// should change it to 250m CPU and 200Mi of memory. Limits should stay unchanged.
		for _, pod := range podList.Items {
			gomega.Expect(*pod.Spec.Containers[0].Resources.Requests.Cpu()).To(gomega.Equal(ParseQuantityOrDie("250m")))
			gomega.Expect(*pod.Spec.Containers[0].Resources.Requests.Memory()).To(gomega.Equal(ParseQuantityOrDie("200Mi")))
			gomega.Expect(*pod.Spec.Containers[0].Resources.Limits.Cpu()).To(gomega.Equal(ParseQuantityOrDie("500m")))
			gomega.Expect(*pod.Spec.Containers[0].Resources.Limits.Memory()).To(gomega.Equal(ParseQuantityOrDie("500Mi")))
		}
	})

	ginkgo.It("caps request according to container max limit set in LimitRange", func() {
		startCpuRequest := ParseQuantityOrDie("100m")
		startCpuLimit := ParseQuantityOrDie("150m")
		startMemRequest := ParseQuantityOrDie("100Mi")
		startMemLimit := ParseQuantityOrDie("200Mi")
		cpuRecommendation := ParseQuantityOrDie("250m")
		memRecommendation := ParseQuantityOrDie("200Mi")

		d := NewHamsterDeploymentWithResourcesAndLimits(f, startCpuRequest, startMemRequest, startCpuLimit, startMemLimit)

		ginkgo.By("Setting up a VPA CRD")
		vpaCRD := NewVPA(f, "hamster-vpa", hamsterTargetRef)
		vpaCRD.Status.Recommendation = &vpa_types.RecommendedPodResources{
			ContainerRecommendations: []vpa_types.RecommendedContainerResources{{
				ContainerName: "hamster",
				Target: apiv1.ResourceList{
					apiv1.ResourceCPU:    cpuRecommendation,
					apiv1.ResourceMemory: memRecommendation,
				},
			}},
		}
		InstallVPA(f, vpaCRD)

		// Max CPU limit is 300m and ratio is 1.5, so max request is 200m, while
		// recommendation is 250m
		// Max memory limit is 1Gi and ratio is 2., so max request is 0.5Gi
		maxCpu := ParseQuantityOrDie("300m")
		InstallLimitRangeWithMax(f, maxCpu.String(), "1Gi", apiv1.LimitTypeContainer)

		ginkgo.By("Setting up a hamster deployment")
		podList := startDeploymentPods(f, d)

		ginkgo.By("Verifying hamster deployment")
		for i, pod := range podList.Items {
			podInfo := fmt.Sprintf("pod %s at index %d", pod.Name, i)

			cpuRequestMsg := fmt.Sprintf("%s: CPU request didn't increase to the recommendation capped to max limit in LimitRange", podInfo)
			gomega.Expect(*pod.Spec.Containers[0].Resources.Requests.Cpu()).To(gomega.Equal(ParseQuantityOrDie("200m")), cpuRequestMsg)

			cpuLimitMsg := fmt.Sprintf("%s: CPU limit above max in LimitRange", podInfo)
			gomega.Expect(pod.Spec.Containers[0].Resources.Limits.Cpu().MilliValue()).To(gomega.BeNumerically("<=", maxCpu.MilliValue()), cpuLimitMsg)

			cpuRatioMsg := fmt.Sprintf("%s: CPU limit / request ratio isn't approximately equal to the original ratio", podInfo)
			cpuRatio := float64(pod.Spec.Containers[0].Resources.Limits.Cpu().MilliValue()) / float64(pod.Spec.Containers[0].Resources.Requests.Cpu().MilliValue())
			gomega.Expect(cpuRatio).To(gomega.BeNumerically("~", 1.5), cpuRatioMsg)

			memRequestMsg := fmt.Sprintf("%s: memory request didn't increase to the recommendation capped to max limit in LimitRange", podInfo)
			gomega.Expect(pod.Spec.Containers[0].Resources.Requests.Memory().Value()).To(gomega.Equal(memRecommendation.Value()), memRequestMsg)

			memLimitMsg := fmt.Sprintf("%s: memory limit above max limit in LimitRange", podInfo)
			gomega.Expect(pod.Spec.Containers[0].Resources.Limits.Memory().Value()).To(gomega.BeNumerically("<=", 1024*1024*1024), memLimitMsg)

			memRatioMsg := fmt.Sprintf("%s: memory limit / request ratio isn't approximately equal to the original ratio", podInfo)
			memRatio := float64(pod.Spec.Containers[0].Resources.Limits.Memory().Value()) / float64(pod.Spec.Containers[0].Resources.Requests.Memory().Value())
			gomega.Expect(memRatio).To(gomega.BeNumerically("~", 2.), memRatioMsg)
		}
	})

	ginkgo.It("raises request according to container min limit set in LimitRange", func() {
		d := NewHamsterDeploymentWithResourcesAndLimits(f,
			ParseQuantityOrDie("100m") /*cpu request*/, ParseQuantityOrDie("200Mi"), /*memory request*/
			ParseQuantityOrDie("150m") /*cpu limit*/, ParseQuantityOrDie("400Mi") /*memory limit*/)

		ginkgo.By("Setting up a VPA CRD")
		vpaCRD := NewVPA(f, "hamster-vpa", hamsterTargetRef)
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

		// Min CPU from limit range is 50m and ratio is 1.5. Min applies to both limit and request so min
		// request is 50m and min limit is 75
		// Min memory limit is 250Mi and it applies to both limit and request. Recommendation is 100Mi.
		// It should be scaled up to 250Mi.
		InstallLimitRangeWithMin(f, "50m", "250Mi", apiv1.LimitTypeContainer)

		ginkgo.By("Setting up a hamster deployment")
		podList := startDeploymentPods(f, d)

		// Originally Pods had 100m CPU, 200Mi of memory, but admission controller
		// should change it to 250m CPU and 125Mi of memory, since this is the lowest
		// request that limitrange allows.
		// Limit to request ratio should stay unchanged.
		for _, pod := range podList.Items {
			gomega.Expect(*pod.Spec.Containers[0].Resources.Requests.Cpu()).To(gomega.Equal(ParseQuantityOrDie("250m")))
			gomega.Expect(*pod.Spec.Containers[0].Resources.Requests.Memory()).To(gomega.Equal(ParseQuantityOrDie("250Mi")))
			gomega.Expect(pod.Spec.Containers[0].Resources.Limits.Cpu().MilliValue()).To(gomega.BeNumerically(">=", 75))
			gomega.Expect(pod.Spec.Containers[0].Resources.Limits.Memory().Value()).To(gomega.BeNumerically(">=", 250*1024*1024))
			gomega.Expect(float64(pod.Spec.Containers[0].Resources.Limits.Cpu().MilliValue()) / float64(pod.Spec.Containers[0].Resources.Requests.Cpu().MilliValue())).To(gomega.BeNumerically("~", 1.5))
			gomega.Expect(float64(pod.Spec.Containers[0].Resources.Limits.Memory().Value()) / float64(pod.Spec.Containers[0].Resources.Requests.Memory().Value())).To(gomega.BeNumerically("~", 2.))
		}
	})

	ginkgo.It("caps request according to pod max limit set in LimitRange", func() {
		d := NewHamsterDeploymentWithResourcesAndLimits(f,
			ParseQuantityOrDie("100m") /*cpu request*/, ParseQuantityOrDie("100Mi"), /*memory request*/
			ParseQuantityOrDie("150m") /*cpu limit*/, ParseQuantityOrDie("200Mi") /*memory limit*/)
		d.Spec.Template.Spec.Containers = append(d.Spec.Template.Spec.Containers, d.Spec.Template.Spec.Containers[0])
		d.Spec.Template.Spec.Containers[1].Name = "hamster2"
		ginkgo.By("Setting up a VPA CRD")
		vpaCRD := NewVPA(f, "hamster-vpa", hamsterTargetRef)
		vpaCRD.Status.Recommendation = &vpa_types.RecommendedPodResources{
			ContainerRecommendations: []vpa_types.RecommendedContainerResources{
				{
					ContainerName: "hamster",
					Target: apiv1.ResourceList{
						apiv1.ResourceCPU:    ParseQuantityOrDie("250m"),
						apiv1.ResourceMemory: ParseQuantityOrDie("200Mi"),
					},
				},
				{
					ContainerName: "hamster2",
					Target: apiv1.ResourceList{
						apiv1.ResourceCPU:    ParseQuantityOrDie("250m"),
						apiv1.ResourceMemory: ParseQuantityOrDie("200Mi"),
					},
				},
			},
		}
		InstallVPA(f, vpaCRD)

		// Max CPU limit is 600m for pod, 300 per container and ratio is 1.5, so max request is 200m,
		// while recommendation is 250m
		// Max memory limit is 1Gi and ratio is 2., so max request is 0.5Gi
		InstallLimitRangeWithMax(f, "600m", "1Gi", apiv1.LimitTypePod)

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

	ginkgo.It("raises request according to pod min limit set in LimitRange", func() {
		d := NewHamsterDeploymentWithResourcesAndLimits(f,
			ParseQuantityOrDie("100m") /*cpu request*/, ParseQuantityOrDie("200Mi"), /*memory request*/
			ParseQuantityOrDie("150m") /*cpu limit*/, ParseQuantityOrDie("400Mi") /*memory limit*/)
		d.Spec.Template.Spec.Containers = append(d.Spec.Template.Spec.Containers, d.Spec.Template.Spec.Containers[0])
		d.Spec.Template.Spec.Containers[1].Name = "hamster2"
		ginkgo.By("Setting up a VPA CRD")
		vpaCRD := NewVPA(f, "hamster-vpa", hamsterTargetRef)
		vpaCRD.Status.Recommendation = &vpa_types.RecommendedPodResources{
			ContainerRecommendations: []vpa_types.RecommendedContainerResources{
				{
					ContainerName: "hamster",
					Target: apiv1.ResourceList{
						apiv1.ResourceCPU:    ParseQuantityOrDie("120m"),
						apiv1.ResourceMemory: ParseQuantityOrDie("100Mi"), // memory is downscaled
					},
				},
				{
					ContainerName: "hamster2",
					Target: apiv1.ResourceList{
						apiv1.ResourceCPU:    ParseQuantityOrDie("120m"),
						apiv1.ResourceMemory: ParseQuantityOrDie("100Mi"), // memory is downscaled
					},
				},
			},
		}
		InstallVPA(f, vpaCRD)

		// Min CPU from limit range is 100m, 50m per pod and ratio is 1.5. Min applies to both limit and
		// request so min request is 50m and min limit is 75
		// Min memory limit is 500Mi per pod, 250 per container and it applies to both limit and request.
		// Recommendation is 100Mi it should be scaled up to 250Mi.
		InstallLimitRangeWithMin(f, "100m", "500Mi", apiv1.LimitTypePod)

		ginkgo.By("Setting up a hamster deployment")
		podList := startDeploymentPods(f, d)

		// Originally Pods had 100m CPU, 200Mi of memory, but admission controller
		// should change it to 250m CPU and 125Mi of memory, since this is the lowest
		// request that limitrange allows.
		// Limit to request ratio should stay unchanged.
		for _, pod := range podList.Items {
			gomega.Expect(*pod.Spec.Containers[0].Resources.Requests.Cpu()).To(gomega.Equal(ParseQuantityOrDie("120m")))
			gomega.Expect(*pod.Spec.Containers[0].Resources.Requests.Memory()).To(gomega.Equal(ParseQuantityOrDie("250Mi")))
			gomega.Expect(pod.Spec.Containers[0].Resources.Limits.Cpu().MilliValue()).To(gomega.BeNumerically(">=", 75))
			gomega.Expect(pod.Spec.Containers[0].Resources.Limits.Memory().Value()).To(gomega.BeNumerically(">=", 250*1024*1024))
			gomega.Expect(float64(pod.Spec.Containers[0].Resources.Limits.Cpu().MilliValue()) / float64(pod.Spec.Containers[0].Resources.Requests.Cpu().MilliValue())).To(gomega.BeNumerically("~", 1.5))
			gomega.Expect(float64(pod.Spec.Containers[0].Resources.Limits.Memory().Value()) / float64(pod.Spec.Containers[0].Resources.Requests.Memory().Value())).To(gomega.BeNumerically("~", 2.))
		}
	})

	ginkgo.It("caps request to max set in VPA", func() {
		d := NewHamsterDeploymentWithResources(f, ParseQuantityOrDie("100m") /*cpu*/, ParseQuantityOrDie("100Mi") /*memory*/)

		ginkgo.By("Setting up a VPA CRD")
		vpaCRD := NewVPA(f, "hamster-vpa", hamsterTargetRef)
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
		vpaCRD := NewVPA(f, "hamster-vpa", hamsterTargetRef)
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
		vpaCRD := NewVPA(f, "hamster-vpa", hamsterTargetRef)
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
		vpaCRD := NewVPA(f, "hamster-vpa", hamsterTargetRef)
		InstallVPA(f, vpaCRD)

		ginkgo.By("Setting up a hamster deployment")
		podList := startDeploymentPods(f, d)

		// VPA has no recommendation, deployment has no request specified
		for _, pod := range podList.Items {
			gomega.Expect(pod.Spec.Containers[0].Resources.Requests).To(gomega.BeEmpty())
		}
	})

	ginkgo.It("accepts valid and rejects invalid VPA object", func() {
		ginkgo.By("Setting up valid VPA object")
		validVPA := []byte(`{
			"kind": "VerticalPodAutoscaler",
			"apiVersion": "autoscaling.k8s.io/v1",
			"metadata": {"name": "hamster-vpa-valid"},
			"spec": {
				"targetRef": {
					"apiVersion": "apps/v1",
					"kind": "Deployment",
					"name":"hamster"
				},
		   	"resourcePolicy": {
		  		"containerPolicies": [{"containerName": "*", "minAllowed":{"cpu":"50m"}}]
		  	}
		  }
		}`)
		err := InstallRawVPA(f, validVPA)
		gomega.Expect(err).NotTo(gomega.HaveOccurred(), "Valid VPA object rejected")

		ginkgo.By("Setting up invalid VPA object")
		// The invalid object differs by name and minAllowed - there is an invalid "requests" field.
		invalidVPA := []byte(`{
			"kind": "VerticalPodAutoscaler",
			"apiVersion": "autoscaling.k8s.io/v1",
			"metadata": {"name": "hamster-vpa-invalid"},
			"spec": {
				"targetRef": {
					"apiVersion": "apps/v1",
					"kind": "Deployment",
					"name":"hamster"
				},
		   	"resourcePolicy": {
		  		"containerPolicies": [{"containerName": "*", "minAllowed":{"requests":{"cpu":"50m"}}}]
		  	}
		  }
		}`)
		err2 := InstallRawVPA(f, invalidVPA)
		gomega.Expect(err2).To(gomega.HaveOccurred(), "Invalid VPA object accepted")
		gomega.Expect(err2.Error()).To(gomega.MatchRegexp(`.*admission webhook .*vpa.* denied the request: .*`))
	})

})

func startDeploymentPods(f *framework.Framework, deployment *appsv1.Deployment) *apiv1.PodList {
	// Apiserver watch can lag depending on cached object count and apiserver resource usage.
	// We assume that watch can lag up to 5 seconds.
	const apiserverWatchLag = 5 * time.Second
	// In admission controller e2e tests a recommendation is created before deployment.
	// Creating deployment with size greater than 0 would create a race between information
	// about pods and information about deployment getting to the admission controller.
	// Any pods that get processed by AC before it receives information about the deployment
	// don't receive recommendation.
	// To avoid this create deployment with size 0, then scale it up to the desired size.
	desiredPodCount := *deployment.Spec.Replicas
	zero := int32(0)
	deployment.Spec.Replicas = &zero
	c, ns := f.ClientSet, f.Namespace.Name
	deployment, err := c.AppsV1().Deployments(ns).Create(deployment)
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), "when creating deployment with size 0")

	err = framework_deployment.WaitForDeploymentComplete(c, deployment)
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), "when waiting for empty deployment to create")
	// If admission controller receives pod before controller it will not apply recommendation and test will fail.
	// Wait after creating deployment to ensure VPA knows about it, then scale up.
	// Normally watch lag is not a problem in terms of correctness:
	// - Mode "Auto": created pod without assigned resources will be handled by the eviction loop.
	// - Mode "Initial": calculating recommendations takes more than potential ectd lag.
	// - Mode "Off": pods are not handled by the admission controller.
	// In e2e admission controller tests we want to focus on scenarios without considering watch lag.
	// TODO(#2631): Remove sleep when issue is fixed.
	time.Sleep(apiserverWatchLag)

	scale := autoscalingv1.Scale{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deployment.ObjectMeta.Name,
			Namespace: deployment.ObjectMeta.Namespace,
		},
		Spec: autoscalingv1.ScaleSpec{
			Replicas: desiredPodCount,
		},
	}
	afterScale, err := c.AppsV1().Deployments(ns).UpdateScale(deployment.Name, &scale)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(afterScale.Spec.Replicas).To(gomega.Equal(desiredPodCount), fmt.Sprintf("expected %d replicas after scaling", desiredPodCount))

	// After scaling deployment we need to retrieve current version with updated replicas count.
	deployment, err = c.AppsV1().Deployments(ns).Get(deployment.Name, metav1.GetOptions{})
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), "when getting scaled deployment")
	err = framework_deployment.WaitForDeploymentComplete(c, deployment)
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), "when waiting for deployment to resize")

	podList, err := framework_deployment.GetPodsForDeployment(c, deployment)
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), "when listing pods after deployment resize")
	return podList
}
