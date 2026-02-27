/*
Copyright 2025 The Kubernetes Authors.

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
	"time"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/e2e/utils"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/features"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/status"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/test"
	"k8s.io/kubernetes/test/e2e/framework"
	framework_deployment "k8s.io/kubernetes/test/e2e/framework/deployment"
	podsecurity "k8s.io/pod-security-admission/api"

	ginkgo "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

const (
	sidecarContainerName = "sidecar"
	mainContainerName    = "hamster"
)

var _ = AdmissionControllerE2eDescribe("Admission-controller native sidecar", func() {
	f := framework.NewDefaultFramework("vertical-pod-autoscaling")
	f.NamespacePodSecurityLevel = podsecurity.LevelBaseline

	ginkgo.BeforeEach(func() {
		waitForVpaWebhookRegistration(f)
	})

	f.It("patches native sidecar init container resources on pod admission",
		framework.WithFeatureGate(features.NativeSidecar), func() {
			d := newDeploymentWithNativeSidecar(f,
				"100m", "100Mi", // main container
				"50m", "50Mi", // sidecar
			)

			ginkgo.By("Setting up a VPA CRD with recommendations for both containers")
			vpaCRD := test.VerticalPodAutoscaler().
				WithName("hamster-vpa").
				WithNamespace(f.Namespace.Name).
				WithTargetRef(utils.HamsterTargetRef).
				WithContainer(mainContainerName).
				WithContainer(sidecarContainerName).
				AppendRecommendation(
					test.Recommendation().
						WithContainer(mainContainerName).
						WithTarget("250m", "200Mi").
						WithLowerBound("250m", "200Mi").
						WithUpperBound("250m", "200Mi").
						GetContainerResources()).
				AppendRecommendation(
					test.Recommendation().
						WithContainer(sidecarContainerName).
						WithTarget("150m", "150Mi").
						WithLowerBound("150m", "150Mi").
						WithUpperBound("150m", "150Mi").
						GetContainerResources()).
				Get()

			utils.InstallVPA(f, vpaCRD)

			ginkgo.By("Setting up a deployment with a native sidecar")
			podList := utils.StartDeploymentPods(f, d)

			ginkgo.By("Verifying main container was patched with recommendations")
			for _, pod := range podList.Items {
				gomega.Expect(pod.Spec.Containers[0].Resources.Requests[apiv1.ResourceCPU]).
					To(gomega.Equal(ParseQuantityOrDie("250m")),
						"main container CPU should be updated to recommendation")
				gomega.Expect(pod.Spec.Containers[0].Resources.Requests[apiv1.ResourceMemory]).
					To(gomega.Equal(ParseQuantityOrDie("200Mi")),
						"main container memory should be updated to recommendation")
			}

			ginkgo.By("Verifying native sidecar init container was patched with recommendations")
			for _, pod := range podList.Items {
				if !gomega.Expect(len(pod.Spec.InitContainers)).To(gomega.BeNumerically(">=", 1),
					"pod should have at least one init container") {
					continue
				}
				gomega.Expect(pod.Spec.InitContainers[0].Resources.Requests[apiv1.ResourceCPU]).
					To(gomega.Equal(ParseQuantityOrDie("150m")),
						"sidecar CPU should be updated to recommendation")
				gomega.Expect(pod.Spec.InitContainers[0].Resources.Requests[apiv1.ResourceMemory]).
					To(gomega.Equal(ParseQuantityOrDie("150Mi")),
						"sidecar memory should be updated to recommendation")
			}
		})

	f.It("does not patch native sidecar when feature gate is disabled", func() {
		// This test runs without the NativeSidecar feature gate, so the sidecar
		// init container should NOT receive recommendations.
		d := newDeploymentWithNativeSidecar(f,
			"100m", "100Mi",
			"50m", "50Mi",
		)

		ginkgo.By("Setting up a VPA CRD with recommendations for both containers")
		vpaCRD := test.VerticalPodAutoscaler().
			WithName("hamster-vpa").
			WithNamespace(f.Namespace.Name).
			WithTargetRef(utils.HamsterTargetRef).
			WithContainer(mainContainerName).
			WithContainer(sidecarContainerName).
			AppendRecommendation(
				test.Recommendation().
					WithContainer(mainContainerName).
					WithTarget("250m", "200Mi").
					WithLowerBound("250m", "200Mi").
					WithUpperBound("250m", "200Mi").
					GetContainerResources()).
			AppendRecommendation(
				test.Recommendation().
					WithContainer(sidecarContainerName).
					WithTarget("150m", "150Mi").
					WithLowerBound("150m", "150Mi").
					WithUpperBound("150m", "150Mi").
					GetContainerResources()).
			Get()

		utils.InstallVPA(f, vpaCRD)

		ginkgo.By("Setting up a deployment with a native sidecar")
		podList := utils.StartDeploymentPods(f, d)

		ginkgo.By("Verifying main container was patched (feature gate doesn't affect regular containers)")
		for _, pod := range podList.Items {
			gomega.Expect(pod.Spec.Containers[0].Resources.Requests[apiv1.ResourceCPU]).
				To(gomega.Equal(ParseQuantityOrDie("250m")))
		}

		ginkgo.By("Verifying sidecar init container was NOT patched (gate disabled)")
		for _, pod := range podList.Items {
			if len(pod.Spec.InitContainers) == 0 {
				continue
			}
			gomega.Expect(pod.Spec.InitContainers[0].Resources.Requests[apiv1.ResourceCPU]).
				To(gomega.Equal(ParseQuantityOrDie("50m")),
					"sidecar CPU should remain unchanged when gate is disabled")
			gomega.Expect(pod.Spec.InitContainers[0].Resources.Requests[apiv1.ResourceMemory]).
				To(gomega.Equal(ParseQuantityOrDie("50Mi")),
					"sidecar memory should remain unchanged when gate is disabled")
		}
	})

	f.It("observed containers annotation includes native sidecar name",
		framework.WithFeatureGate(features.NativeSidecar), func() {
			d := newDeploymentWithNativeSidecar(f,
				"100m", "100Mi",
				"50m", "50Mi",
			)

			ginkgo.By("Setting up a VPA CRD with recommendations for both containers")
			vpaCRD := test.VerticalPodAutoscaler().
				WithName("hamster-vpa").
				WithNamespace(f.Namespace.Name).
				WithTargetRef(utils.HamsterTargetRef).
				WithContainer(mainContainerName).
				WithContainer(sidecarContainerName).
				AppendRecommendation(
					test.Recommendation().
						WithContainer(mainContainerName).
						WithTarget("250m", "200Mi").
						WithLowerBound("250m", "200Mi").
						WithUpperBound("250m", "200Mi").
						GetContainerResources()).
				AppendRecommendation(
					test.Recommendation().
						WithContainer(sidecarContainerName).
						WithTarget("150m", "150Mi").
						WithLowerBound("150m", "150Mi").
						WithUpperBound("150m", "150Mi").
						GetContainerResources()).
				Get()

			utils.InstallVPA(f, vpaCRD)

			ginkgo.By("Setting up a deployment with a native sidecar")
			podList := utils.StartDeploymentPods(f, d)

			ginkgo.By("Verifying observed containers annotation includes sidecar")
			for _, pod := range podList.Items {
				ann := pod.GetAnnotations()["vpaObservedContainers"]
				gomega.Expect(ann).To(gomega.ContainSubstring(sidecarContainerName),
					"vpaObservedContainers annotation should include the native sidecar name")
				gomega.Expect(ann).To(gomega.ContainSubstring(mainContainerName),
					"vpaObservedContainers annotation should include the main container name")
			}
		})
})

var _ = UpdaterE2eDescribe("Updater native sidecar", func() {
	f := framework.NewDefaultFramework("vertical-pod-autoscaling")
	f.NamespacePodSecurityLevel = podsecurity.LevelBaseline

	f.It("evicts pods when native sidecar resources are outside recommended range",
		framework.WithFeatureGate(features.NativeSidecar), framework.WithSerial(), func() {
			const statusUpdateInterval = 10 * time.Second

			ginkgo.By("Setting up the Admission Controller status")
			stopCh := make(chan struct{})
			statusUpdater := status.NewUpdater(
				f.ClientSet,
				status.AdmissionControllerStatusName,
				utils.VpaNamespace,
				statusUpdateInterval,
				"e2e test",
			)
			defer func() {
				ginkgo.By("Deleting the Admission Controller status")
				close(stopCh)
				err := f.ClientSet.CoordinationV1().Leases(utils.VpaNamespace).
					Delete(context.TODO(), status.AdmissionControllerStatusName, metav1.DeleteOptions{})
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			}()
			statusUpdater.Run(stopCh)

			podList := setupNativeSidecarPodsForEviction(f)

			ginkgo.By("Waiting for pods to be evicted")
			err := WaitForPodsEvicted(f, podList)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		})
})

// newDeploymentWithNativeSidecar creates a deployment that has a main container
// and a native sidecar init container (restartPolicy: Always).
func newDeploymentWithNativeSidecar(f *framework.Framework, mainCPU, mainMemory, sidecarCPU, sidecarMemory string) *appsv1.Deployment {
	d := NewHamsterDeploymentWithResources(f,
		ParseQuantityOrDie(mainCPU),
		ParseQuantityOrDie(mainMemory),
	)

	restartAlways := apiv1.ContainerRestartPolicyAlways
	d.Spec.Template.Spec.InitContainers = []apiv1.Container{
		{
			Name:          sidecarContainerName,
			Image:         "registry.k8s.io/ubuntu-slim:0.14",
			RestartPolicy: &restartAlways,
			Resources: apiv1.ResourceRequirements{
				Requests: apiv1.ResourceList{
					apiv1.ResourceCPU:    ParseQuantityOrDie(sidecarCPU),
					apiv1.ResourceMemory: ParseQuantityOrDie(sidecarMemory),
				},
			},
			Command: []string{"/bin/sh"},
			Args:    []string{"-c", "while true; do sleep 10; done"},
		},
	}

	return d
}

// setupNativeSidecarPodsForEviction sets up a deployment with a native sidecar
// whose current requests are significantly below the VPA recommendation, causing
// the updater to evict the pods.
func setupNativeSidecarPodsForEviction(f *framework.Framework) *apiv1.PodList {
	ginkgo.By("Setting up a deployment with a native sidecar")
	d := newDeploymentWithNativeSidecar(f,
		"100m", "100Mi", // main container - under-provisioned
		"50m", "50Mi", // sidecar - under-provisioned
	)
	d.Spec.Replicas = &[]int32{utils.DefaultHamsterReplicas}[0]
	d, err := f.ClientSet.AppsV1().Deployments(f.Namespace.Name).Create(context.TODO(), d, metav1.CreateOptions{})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	err = framework_deployment.WaitForDeploymentComplete(f.ClientSet, d)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	podList, err := GetHamsterPods(f)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	ginkgo.By("Setting up a VPA CRD with higher recommendations")
	vpaCRD := test.VerticalPodAutoscaler().
		WithName("hamster-vpa").
		WithNamespace(f.Namespace.Name).
		WithTargetRef(utils.HamsterTargetRef).
		WithUpdateMode(vpa_types.UpdateModeRecreate).
		WithContainer(mainContainerName).
		WithContainer(sidecarContainerName).
		AppendRecommendation(
			test.Recommendation().
				WithContainer(mainContainerName).
				WithTarget("200m", "200Mi").
				WithLowerBound("200m", "200Mi").
				WithUpperBound("200m", "200Mi").
				GetContainerResources()).
		AppendRecommendation(
			test.Recommendation().
				WithContainer(sidecarContainerName).
				WithTarget("200m", "200Mi").
				WithLowerBound("200m", "200Mi").
				WithUpperBound("200m", "200Mi").
				GetContainerResources()).
		Get()

	utils.InstallVPA(f, vpaCRD)

	return podList
}
