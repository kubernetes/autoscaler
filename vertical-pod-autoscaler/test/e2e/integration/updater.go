/*
Copyright The Kubernetes Authors.

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
	"time"

	apiv1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/test/e2e/utils"

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/test"
	"k8s.io/kubernetes/test/e2e/framework"
	podsecurity "k8s.io/pod-security-admission/api"

	ginkgo "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

// evictionTimeout is how long to wait for the updater to evict a pod. It
// covers a few runs of an updater started with --updater-interval=10s.
const evictionTimeout = 3 * time.Minute

var _ = utils.UpdaterE2eDescribe("Flags", func() {
	f := framework.NewDefaultFramework("vertical-pod-autoscaling")
	f.NamespacePodSecurityLevel = podsecurity.LevelBaseline

	var hamsterNamespace string

	ginkgo.BeforeEach(func() {
		hamsterNamespace = f.Namespace.Name
	})

	ginkgo.AfterEach(func() {
		// A test that fails while deploying the updater leaves the framework
		// namespace set to utils.VpaNamespace. Restore it so that the
		// framework deletes the hamster namespace.
		f.Namespace.Name = hamsterNamespace
		f.ClientSet.AppsV1().Deployments(utils.VpaNamespace).Delete(context.TODO(), utils.UpdaterDeploymentName, metav1.DeleteOptions{})
	})

	ginkgo.It("starts updater with --min-replicas parameter", func() {
		ginkgo.By("Setting up VPA deployment")
		f.Namespace.Name = utils.VpaNamespace
		// No admission controller runs in the integration cluster, so the
		// updater must not wait for its status before evicting.
		vpaDeployment := utils.NewVPAComponentDeployment(f, utils.UpdaterComponentConfig(
			"--updater-interval=10s",
			"--use-admission-controller-status=false",
			"--min-replicas=1",
		))
		utils.StartDeploymentPods(f, vpaDeployment)

		// With the default --min-replicas=2 a single replica is never evicted.
		f.Namespace.Name = hamsterNamespace
		podList := setupHamsterWithVPA(f, 1, "200m", "200Mi")

		ginkgo.By("Waiting for the pod outside the recommended range to be evicted")
		waitForAnyPodEvicted(f, podList)
	})

	ginkgo.It("starts updater with --in-recommendation-bounds-eviction-lifetime-threshold parameter", func() {
		ginkgo.By("Setting up VPA deployment")
		f.Namespace.Name = utils.VpaNamespace
		vpaDeployment := utils.NewVPAComponentDeployment(f, utils.UpdaterComponentConfig(
			"--updater-interval=10s",
			"--use-admission-controller-status=false",
			"--in-recommendation-bounds-eviction-lifetime-threshold=1m",
		))
		utils.StartDeploymentPods(f, vpaDeployment)

		// Requests are within the recommended range, so with the default
		// threshold of 12h the pods are never evicted.
		f.Namespace.Name = hamsterNamespace
		podList := setupHamsterWithVPA(f, utils.DefaultHamsterReplicas, "50m", "50Mi")

		ginkgo.By("Waiting for a pod within the recommended range to be evicted after the lifetime threshold")
		waitForAnyPodEvicted(f, podList)
	})
})

// setupHamsterWithVPA starts a hamster deployment requesting 100m CPU and
// 100Mi memory, and installs a VPA in Recreate mode recommending 200m and
// 200Mi with the given lower bound and an upper bound of 300m and 300Mi.
func setupHamsterWithVPA(f *framework.Framework, replicas int32, lowerBoundCPU, lowerBoundMemory string) *apiv1.PodList {
	ginkgo.By("Setting up a hamster deployment")
	d := utils.NewNHamstersDeployment(f, 1)
	d.Spec.Replicas = &replicas
	d.Spec.Template.Spec.Containers[0].Resources.Requests = apiv1.ResourceList{
		apiv1.ResourceCPU:    resource.MustParse("100m"),
		apiv1.ResourceMemory: resource.MustParse("100Mi"),
	}
	podList := utils.StartDeploymentPods(f, d)

	ginkgo.By("Setting up VPA")
	vpaCRD := test.VerticalPodAutoscaler().
		WithName("hamster-vpa").
		WithNamespace(f.Namespace.Name).
		WithTargetRef(utils.HamsterTargetRef).
		WithUpdateMode(vpa_types.UpdateModeRecreate).
		WithContainer(utils.GetHamsterContainerNameByIndex(0)).
		WithTarget("200m", "200Mi").
		WithLowerBound(lowerBoundCPU, lowerBoundMemory).
		WithUpperBound("300m", "300Mi").
		Get()
	utils.InstallVPA(f, vpaCRD)

	return podList
}

// waitForAnyPodEvicted waits until at least one pod of podList is deleted or
// being deleted.
func waitForAnyPodEvicted(f *framework.Framework, podList *apiv1.PodList) {
	gomega.Eventually(func() bool {
		for _, pod := range podList.Items {
			current, err := f.ClientSet.CoreV1().Pods(pod.Namespace).Get(context.TODO(), pod.Name, metav1.GetOptions{})
			if apierrors.IsNotFound(err) || (err == nil && current.DeletionTimestamp != nil) {
				return true
			}
		}
		return false
	}, evictionTimeout, utils.PollInterval).Should(gomega.BeTrue(), "expected the updater to evict a hamster pod")
}
