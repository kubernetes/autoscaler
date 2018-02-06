/*
Copyright 2015 The Kubernetes Authors.

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

	"k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/poc.autoscaling.k8s.io/v1alpha1"
	vpa_clientset "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned"
	"k8s.io/kubernetes/test/e2e/framework"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

func testDescribe(text string, body func()) bool {
	return ginkgo.Describe("[VPA] "+text, body)
}

const (
	pollInterval = framework.Poll
	pollTimeout  = 5 * time.Minute
)

var _ = testDescribe("VPA CRD object", func() {
	f := framework.NewDefaultFramework("vertical-pod-autoscaling")

	ginkgo.It("serves recommendation", func() {

		ginkgo.By("Setting up a hamster deployment")
		c := f.ClientSet
		ns := f.Namespace.Name

		d := hamsterDeployment(f)
		_, err := c.ExtensionsV1beta1().Deployments(ns).Create(d)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		err = framework.WaitForDeploymentComplete(c, d)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		ginkgo.By("Setting up a VPA CRD")
		config, err := framework.LoadConfig()
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		vpaCRD := newVPA(f, "hamster-vpa", &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"app": "hamster",
			},
		})

		vpaClientSet := vpa_clientset.NewForConfigOrDie(config)
		vpaClient := vpaClientSet.PocV1alpha1()
		_, err = vpaClient.VerticalPodAutoscalers(ns).Create(vpaCRD)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		ginkgo.By("Waiting for recommendation to be filled")
		err = waitForRecommendationPresent(vpaClientSet, vpaCRD)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

	})
})

func hamsterDeployment(f *framework.Framework) *extensions.Deployment {
	d := framework.NewDeployment("hamster-deployment", 3, map[string]string{"app": "hamster"}, "hamster", "gcr.io/google_containers/ubuntu-slim:0.1", extensions.RollingUpdateDeploymentStrategyType)
	d.ObjectMeta.Namespace = f.Namespace.Name
	cpuQuantity, err := resource.ParseQuantity("100m")
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	memoryQuantity, err := resource.ParseQuantity("100Mi")
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	d.Spec.Template.Spec.Containers[0].Resources.Requests = v1.ResourceList{
		v1.ResourceCPU:    cpuQuantity,
		v1.ResourceMemory: memoryQuantity,
	}
	d.Spec.Template.Spec.Containers[0].Command = []string{"/bin/sh"}
	d.Spec.Template.Spec.Containers[0].Args = []string{"-c", "/usr/bin/yes >/dev/null"}
	return d
}

func newVPA(f *framework.Framework, name string, selector *metav1.LabelSelector) *vpa_types.VerticalPodAutoscaler {
	vpa := vpa_types.VerticalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: f.Namespace.Name,
		},
		Spec: vpa_types.VerticalPodAutoscalerSpec{
			Selector: selector,
			ResourcePolicy: vpa_types.PodResourcePolicy{
				ContainerPolicies: []vpa_types.ContainerResourcePolicy{},
			},
		},
	}
	return &vpa
}

func waitForRecommendationPresent(c *vpa_clientset.Clientset, vpa *vpa_types.VerticalPodAutoscaler) error {
	err := wait.PollImmediate(pollInterval, pollTimeout, func() (bool, error) {
		var err error
		polledVpa, err := c.PocV1alpha1().VerticalPodAutoscalers(vpa.Namespace).Get(vpa.Name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}

		if len(polledVpa.Status.Recommendation.ContainerRecommendations) != 0 {
			return true, nil
		}

		return false, nil
	})

	if err != nil {
		return fmt.Errorf("error waiting for recommendation present in %v: %v", vpa.Name, err)
	}
	return nil
}
