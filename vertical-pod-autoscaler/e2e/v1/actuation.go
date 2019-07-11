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
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/test/e2e/framework"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

var _ = ActuationSuiteE2eDescribe("Actuation", func() {
	f := framework.NewDefaultFramework("vertical-pod-autoscaling")

	ginkgo.It("stops when pods get pending", func() {

		ginkgo.By("Setting up a hamster deployment")
		d := SetupHamsterDeployment(f, "100m", "100Mi", defaultHamsterReplicas)

		ginkgo.By("Setting up a VPA CRD with ridiculous request")
		SetupVPA(f, "9999", vpa_types.UpdateModeAuto, hamsterTargetRef) // Request 9999 CPUs to make POD pending

		ginkgo.By("Waiting for pods to be restarted and stuck pending")
		err := assertPodsPendingForDuration(f.ClientSet, d, 1, 2*time.Minute)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

	})

	ginkgo.It("never applies recommendations when update mode is Off", func() {
		ginkgo.By("Setting up a hamster deployment")
		d := SetupHamsterDeployment(f, "100m", "100Mi", defaultHamsterReplicas)
		cpuRequest := getCPURequest(d.Spec.Template.Spec)
		podList, err := GetHamsterPods(f)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		podSet := MakePodSet(podList)

		ginkgo.By("Setting up a VPA CRD in mode Off")
		SetupVPA(f, "200m", vpa_types.UpdateModeOff, hamsterTargetRef)

		ginkgo.By(fmt.Sprintf("Waiting for pods to be evicted, hoping it won't happen, sleep for %s", VpaEvictionTimeout.String()))
		CheckNoPodsEvicted(f, podSet)
		ginkgo.By("Forcefully killing one pod")
		killPod(f, podList)

		ginkgo.By("Checking the requests were not modified")
		updatedPodList, err := GetHamsterPods(f)
		for _, pod := range updatedPodList.Items {
			gomega.Expect(getCPURequest(pod.Spec)).To(gomega.Equal(cpuRequest))
		}
	})

	ginkgo.It("applies recommendations only on restart when update mode is Initial", func() {
		ginkgo.By("Setting up a hamster deployment")
		SetupHamsterDeployment(f, "100m", "100Mi", defaultHamsterReplicas)
		podList, err := GetHamsterPods(f)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		podSet := MakePodSet(podList)

		ginkgo.By("Setting up a VPA CRD in mode Initial")
		SetupVPA(f, "200m", vpa_types.UpdateModeInitial, hamsterTargetRef)
		updatedCPURequest := ParseQuantityOrDie("200m")

		ginkgo.By(fmt.Sprintf("Waiting for pods to be evicted, hoping it won't happen, sleep for %s", VpaEvictionTimeout.String()))
		CheckNoPodsEvicted(f, podSet)
		ginkgo.By("Forcefully killing one pod")
		killPod(f, podList)

		ginkgo.By("Checking that request was modified after forceful restart")
		updatedPodList, err := GetHamsterPods(f)
		foundUpdated := 0
		for _, pod := range updatedPodList.Items {
			podRequest := getCPURequest(pod.Spec)
			framework.Logf("podReq: %v", podRequest)
			if podRequest.Cmp(updatedCPURequest) == 0 {
				foundUpdated += 1
			}
		}
		gomega.Expect(foundUpdated).To(gomega.Equal(1))
	})
})

func getCPURequest(podSpec apiv1.PodSpec) resource.Quantity {
	return podSpec.Containers[0].Resources.Requests[apiv1.ResourceCPU]
}

func killPod(f *framework.Framework, podList *apiv1.PodList) {
	f.ClientSet.CoreV1().Pods(f.Namespace.Name).Delete(podList.Items[0].Name, &metav1.DeleteOptions{})
	err := WaitForPodsRestarted(f, podList)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
}

// assertPodsPendingForDuration checks that at most pendingPodsNum pods are pending for pendingDuration
func assertPodsPendingForDuration(c clientset.Interface, deployment *appsv1.Deployment, pendingPodsNum int, pendingDuration time.Duration) error {

	pendingPods := make(map[string]time.Time)

	err := wait.PollImmediate(pollInterval, pollTimeout+pendingDuration, func() (bool, error) {
		var err error
		currentPodList, err := framework.GetPodsForDeployment(c, deployment)
		if err != nil {
			return false, err
		}

		missingPods := make(map[string]bool)
		for podName := range pendingPods {
			missingPods[podName] = true
		}

		now := time.Now()
		for _, pod := range currentPodList.Items {
			delete(missingPods, pod.Name)
			switch pod.Status.Phase {
			case apiv1.PodPending:
				_, ok := pendingPods[pod.Name]
				if !ok {
					pendingPods[pod.Name] = now
				}
			default:
				delete(pendingPods, pod.Name)
			}
		}

		for missingPod := range missingPods {
			delete(pendingPods, missingPod)
		}

		if len(pendingPods) < pendingPodsNum {
			return false, nil
		}

		if len(pendingPods) > pendingPodsNum {
			return false, fmt.Errorf("%v pending pods seen - expecting %v", len(pendingPods), pendingPodsNum)
		}

		for p, t := range pendingPods {
			fmt.Println("task", now, p, t, now.Sub(t), pendingDuration)
			if now.Sub(t) < pendingDuration {
				return false, nil
			}
		}

		return true, nil
	})

	if err != nil {
		return fmt.Errorf("assertion failed for pending pods in %v: %v", deployment.Name, err)
	}
	return nil
}
