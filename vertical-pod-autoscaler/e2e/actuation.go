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
	"time"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/poc.autoscaling.k8s.io/v1alpha1"
	vpa_clientset "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/test/e2e/framework"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

var _ = ActuationSuiteE2eDescribe("Actuation", func() {
	f := framework.NewDefaultFramework("vertical-pod-autoscaling")

	ginkgo.It("stops when pods get pending", func() {

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

		ginkgo.By("Setting up a VPA CRD with ridiculous request")
		config, err := framework.LoadConfig()
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		vpaCRD := NewVPA(f, "hamster-vpa", &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"app": "hamster",
			},
		})

		resourceList := apiv1.ResourceList{
			apiv1.ResourceCPU: ParseQuantityOrDie("9999"), // Request 9999 CPUs to make POD pending
		}

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

		ginkgo.By("Waiting for pods to be restarted and stuck pending")

		err = assertPodsPendingForDuration(c, d, 1, 2*time.Minute)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

	})
})

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
		return fmt.Errorf("Assertion failed for pending pods in %v: %v", deployment.Name, err)
	}
	return nil
}
