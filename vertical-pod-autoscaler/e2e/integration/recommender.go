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

package integration

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/e2e/utils"

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	vpa_clientset "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/test"
	"k8s.io/kubernetes/test/e2e/framework"
	podsecurity "k8s.io/pod-security-admission/api"

	ginkgo "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = utils.RecommenderE2eDescribe("Flags", func() {
	f := framework.NewDefaultFramework("vertical-pod-autoscaling")
	f.NamespacePodSecurityEnforceLevel = podsecurity.LevelBaseline

	var vpaClientSet vpa_clientset.Interface
	var hamsterNamespace string

	ginkgo.BeforeEach(func() {
		vpaClientSet = utils.GetVpaClientSet(f)
		hamsterNamespace = f.Namespace.Name
	})

	ginkgo.AfterEach(func() {
		f.ClientSet.AppsV1().Deployments(utils.RecommenderNamespace).Delete(context.TODO(), utils.RecommenderDeploymentName, metav1.DeleteOptions{})
	})

	ginkgo.It("starts recommender with --vpa-object-namespace parameter", func() {
		ginkgo.By("Setting up VPA deployment")
		ignoredNamespace, err := f.CreateNamespace(context.TODO(), "ignored-namespace", nil)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		f.Namespace.Name = utils.RecommenderNamespace
		vpaDeployment := utils.NewVPADeployment(f, []string{
			"--recommender-interval=10s",
			fmt.Sprintf("--vpa-object-namespace=%s", hamsterNamespace),
		})
		utils.StartDeploymentPods(f, vpaDeployment)

		testIncludedAndIgnoredNamespaces(f, vpaClientSet, hamsterNamespace, ignoredNamespace.Name)
	})

	ginkgo.It("starts recommender with --ignored-vpa-object-namespaces parameter", func() {
		ginkgo.By("Setting up VPA deployment")
		ignoredNamespace, err := f.CreateNamespace(context.TODO(), "ignored-namespace", nil)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		f.Namespace.Name = utils.RecommenderNamespace
		vpaDeployment := utils.NewVPADeployment(f, []string{
			"--recommender-interval=10s",
			fmt.Sprintf("--ignored-vpa-object-namespaces=%s", ignoredNamespace.Name),
		})
		utils.StartDeploymentPods(f, vpaDeployment)

		testIncludedAndIgnoredNamespaces(f, vpaClientSet, hamsterNamespace, ignoredNamespace.Name)
	})
})

// Create VPA and deployment in 2 namespaces, 1 should be ignored
// Ignored namespace VPA and deployment are intentionally created first
// so that by the time included namespace has recommendation generated,
// we know that ignored namespace has been waiting long enough.
func testIncludedAndIgnoredNamespaces(f *framework.Framework, vpaClientSet vpa_clientset.Interface, includedNamespace, ignoredNamespace string) {
	ginkgo.By("Setting up a hamster deployment in ignored namespace")
	f.Namespace.Name = ignoredNamespace
	d := utils.NewNHamstersDeployment(f, 2)
	_ = utils.StartDeploymentPods(f, d)

	ginkgo.By("Setting up VPA for ignored namespace")
	container1Name := utils.GetHamsterContainerNameByIndex(0)
	container2Name := utils.GetHamsterContainerNameByIndex(1)
	ignoredVpaCRD := test.VerticalPodAutoscaler().
		WithName("hamster-vpa").
		WithNamespace(ignoredNamespace).
		WithTargetRef(utils.HamsterTargetRef).
		WithContainer(container1Name).
		WithScalingMode(container1Name, vpa_types.ContainerScalingModeOff).
		WithContainer(container2Name).
		Get()
	f.Namespace.Name = ignoredNamespace
	utils.InstallVPA(f, ignoredVpaCRD)

	ginkgo.By("Setting up a hamster deployment in included namespace")
	f.Namespace.Name = includedNamespace
	d = utils.NewNHamstersDeployment(f, 2)
	_ = utils.StartDeploymentPods(f, d)

	ginkgo.By("Setting up VPA for included namespace")
	vpaCRD := test.VerticalPodAutoscaler().
		WithName("hamster-vpa").
		WithNamespace(includedNamespace).
		WithTargetRef(utils.HamsterTargetRef).
		WithContainer(container1Name).
		WithScalingMode(container1Name, vpa_types.ContainerScalingModeOff).
		WithContainer(container2Name).
		Get()

	f.Namespace.Name = includedNamespace
	utils.InstallVPA(f, vpaCRD)

	ginkgo.By("Waiting for recommendation to be filled for just one container")
	vpa, err := utils.WaitForRecommendationPresent(vpaClientSet, vpaCRD)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	errMsg := fmt.Sprintf("%s container has recommendations turned off. We expect expect only recommendations for %s",
		utils.GetHamsterContainerNameByIndex(0),
		utils.GetHamsterContainerNameByIndex(1))
	gomega.Expect(vpa.Status.Recommendation.ContainerRecommendations).Should(gomega.HaveLen(1), errMsg)
	gomega.Expect(vpa.Status.Recommendation.ContainerRecommendations[0].ContainerName).To(gomega.Equal(utils.GetHamsterContainerNameByIndex(1)), errMsg)

	ginkgo.By("Ignored namespace should not be recommended")
	ignoredVpa, err := vpaClientSet.AutoscalingV1().VerticalPodAutoscalers(ignoredNamespace).Get(context.TODO(), ignoredVpaCRD.Name, metav1.GetOptions{})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(ignoredVpa.Status.Conditions).Should(gomega.HaveLen(0))
}
