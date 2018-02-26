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

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/poc.autoscaling.k8s.io/v1alpha1"
	"k8s.io/kubernetes/test/e2e/framework"
)

const (
	recommenderComponent         = "recommender"
	updateComponent              = "updater"
	admissionControllerComponent = "admission-controller"
	fullVpaSuite                 = "full-vpa"
	actuationSuite               = "actuation"
	pollInterval                 = framework.Poll
	pollTimeout                  = 5 * time.Minute
)

func e2eDescribe(scenario, name string, body func()) bool {
	return ginkgo.Describe(fmt.Sprintf("[VPA] [%s] %s", scenario, name), body)
}

func recommenderE2eDescribe(name string, body func()) bool {
	return e2eDescribe(recommenderComponent, name, body)
}

func updaterE2eDescribe(name string, body func()) bool {
	return e2eDescribe(updateComponent, name, body)
}

func admissionControllerE2eDescribe(name string, body func()) bool {
	return e2eDescribe(admissionControllerComponent, name, body)
}

func fullVpaE2eDescribe(name string, body func()) bool {
	return e2eDescribe(fullVpaSuite, name, body)
}

func actuationSuiteE2eDescribe(name string, body func()) bool {
	return e2eDescribe(actuationSuite, name, body)
}

func hamsterDeployment(f *framework.Framework, cpuQuantity, memoryQuantity resource.Quantity) *extensions.Deployment {
	d := framework.NewDeployment("hamster-deployment", 3, map[string]string{"app": "hamster"}, "hamster", "gcr.io/google_containers/ubuntu-slim:0.1", extensions.RollingUpdateDeploymentStrategyType)
	d.ObjectMeta.Namespace = f.Namespace.Name
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
			UpdatePolicy: vpa_types.PodUpdatePolicy{
				UpdateMode: vpa_types.UpdateModeAuto,
			},
			ResourcePolicy: vpa_types.PodResourcePolicy{
				ContainerPolicies: []vpa_types.ContainerResourcePolicy{},
			},
		},
	}
	return &vpa
}

func parseQuantityOrDie(text string) resource.Quantity {
	quantity, err := resource.ParseQuantity(text)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	return quantity
}
