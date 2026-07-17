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

package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	buffersfake "k8s.io/autoscaler/cluster-autoscaler/apis/capacitybuffer/client/clientset/versioned/fake"
	fakeclient "k8s.io/client-go/kubernetes/fake"
)

func TestClientGetPodTemplate(t *testing.T) {
	podTemplate1 := &corev1.PodTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "podTemp1",
			Namespace:  "default",
			Generation: 3,
		},
	}
	podTemplate2 := &corev1.PodTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "podTemp2",
			Namespace:  "default",
			Generation: 4,
		},
	}
	tests := []struct {
		name                      string
		objectsInKubernetesClient []runtime.Object
		objectName                string
		expectedValue             *corev1.PodTemplate
		expectError               bool
	}{
		{
			name:                      "simple existing pod template",
			objectsInKubernetesClient: []runtime.Object{podTemplate1, podTemplate2},
			objectName:                podTemplate1.Name,
			expectedValue:             podTemplate1,
			expectError:               false,
		},
		{
			name:                      "simple non existing pod template",
			objectsInKubernetesClient: []runtime.Object{podTemplate1, podTemplate2},
			objectName:                "RandomName",
			expectedValue:             nil,
			expectError:               true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fakeKubernetesClient := fakeclient.NewSimpleClientset(test.objectsInKubernetesClient...)
			fakeBuffersClient := buffersfake.NewSimpleClientset()
			fakeCapacityBuffersClient, _ := NewCapacityBufferClientFromClients(fakeBuffersClient, fakeKubernetesClient, nil, nil)
			pt, err := fakeCapacityBuffersClient.GetPodTemplate("default", test.objectName)
			assert.Equal(t, err != nil, test.expectError)
			assert.Equal(t, pt, test.expectedValue)
		})
	}
}
