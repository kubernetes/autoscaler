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
	"time"

	extensions "k8s.io/api/extensions/v1beta1"
	"k8s.io/kubernetes/test/e2e/framework"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func SIGDescribe(text string, body func()) bool {
	return Describe("[vpa] "+text, body)
}

var _ = SIGDescribe("[VPA] Vertical pod autoscaling dummy test", func() {
	f := framework.NewDefaultFramework("vertical-pod-autoscaling")

	SIGDescribe("Deployment", func() {
		It("test", func() {
			c := f.ClientSet
			ns := f.Namespace.Name
			d := framework.NewDeployment("vpa-recommender", 1, map[string]string{"test": "app"}, "recommender", "eu.gcr.io/kubernetes-schylek-vpa/recommender:0.0.1", extensions.RollingUpdateDeploymentStrategyType)
			_, err := c.ExtensionsV1beta1().Deployments(ns).Create(d)
			Expect(err).NotTo(HaveOccurred())

			time.Sleep(1 * time.Minute)

		})
	})
})
