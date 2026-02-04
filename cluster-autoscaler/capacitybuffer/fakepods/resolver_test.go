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

package fakepods

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	core "k8s.io/client-go/testing"
)

func TestResolver(t *testing.T) {
	namespace := "default"
	templateName := "test-template"
	podName := "test-pod"

	podTemplate := &corev1.PodTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name:       templateName,
			Namespace:  namespace,
			Generation: 1,
		},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{"app": "test"},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "nginx",
						Image: "nginx",
						Resources: corev1.ResourceRequirements{
							Limits: corev1.ResourceList{
								corev1.ResourceCPU: resource.MustParse("1"),
							},
						},
					},
				},
			},
		},
	}

	fakeClient := fake.NewClientset()
	fakeClient.PrependReactor("create", "pods", func(action core.Action) (handled bool, ret runtime.Object, err error) {
		createAction, ok := action.(core.CreateActionImpl)
		if !ok {
			return false, nil, nil
		}
		pod := createAction.GetObject().(*corev1.Pod)

		// Verify dry run
		dryRun := createAction.GetCreateOptions().DryRun
		if len(dryRun) != 1 || dryRun[0] != metav1.DryRunAll {
			return false, nil, nil
		}

		// Simulate defaulting/mutation
		pod.Name = podName
		var newContainers []corev1.Container
		for _, c := range pod.Spec.Containers {
			for res, quantity := range c.Resources.Limits {
				if _, ok := c.Resources.Requests[res]; !ok {
					if c.Resources.Requests == nil {
						c.Resources.Requests = make(corev1.ResourceList)
					}
					c.Resources.Requests[res] = quantity
				}
			}
			newContainers = append(newContainers, c)
		}
		pod.Spec.Containers = newContainers
		return true, pod, nil
	})

	resolver := NewResolver(fakeClient)

	pod, err := resolver.Resolve(t.Context(), podTemplate.Namespace, &podTemplate.Template)
	assert.NoError(t, err)
	assert.Equal(t, podName, pod.Name)
	assert.Equal(t, corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("1")}, pod.Spec.Containers[0].Resources.Requests)
}
