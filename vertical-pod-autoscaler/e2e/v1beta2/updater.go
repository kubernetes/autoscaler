/*
Copyright 2020 The Kubernetes Authors.

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
	"fmt"
	"time"

	autoscaling "k8s.io/api/autoscaling/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1beta2"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/status"
	"k8s.io/kubernetes/test/e2e/framework"
	podsecurity "k8s.io/pod-security-admission/api"

	ginkgo "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = UpdaterE2eDescribe("Updater", func() {
	f := framework.NewDefaultFramework("vertical-pod-autoscaling")
	f.NamespacePodSecurityEnforceLevel = podsecurity.LevelBaseline

	ginkgo.It("evicts pods when Admission Controller status available", func() {
		const statusUpdateInterval = 10 * time.Second

		ginkgo.By("Setting up the Admission Controller status")
		stopCh := make(chan struct{})
		statusUpdater := status.NewUpdater(
			f.ClientSet,
			status.AdmissionControllerStatusName,
			status.AdmissionControllerStatusNamespace,
			statusUpdateInterval,
			"e2e test",
		)
		defer func() {
			// Schedule a cleanup of the Admission Controller status.
			// Status is created outside the test namespace.
			ginkgo.By("Deleting the Admission Controller status")
			close(stopCh)
			err := f.ClientSet.CoordinationV1().Leases(status.AdmissionControllerStatusNamespace).
				Delete(context.TODO(), status.AdmissionControllerStatusName, metav1.DeleteOptions{})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		}()
		statusUpdater.Run(stopCh)

		podList := setupPodsForEviction(f)

		ginkgo.By("Waiting for pods to be evicted")
		err := WaitForPodsEvicted(f, podList)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	})

	ginkgo.It("doesn't evict pods when Admission Controller status unavailable", func() {
		podList := setupPodsForEviction(f)

		ginkgo.By(fmt.Sprintf("Waiting for pods to be evicted, hoping it won't happen, sleep for %s", VpaEvictionTimeout.String()))
		CheckNoPodsEvicted(f, MakePodSet(podList))
	})
})

func setupPodsForEviction(f *framework.Framework) *apiv1.PodList {
	controller := &autoscaling.CrossVersionObjectReference{
		APIVersion: "apps/v1",
		Kind:       "Deployment",
		Name:       "hamster-deployment",
	}
	ginkgo.By(fmt.Sprintf("Setting up a hamster %v", controller.Kind))
	setupHamsterController(f, controller.Kind, "100m", "100Mi", defaultHamsterReplicas)
	podList, err := GetHamsterPods(f)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	ginkgo.By("Setting up a VPA CRD")
	SetupVPA(f, "200m", vpa_types.UpdateModeRecreate, controller)

	return podList
}
