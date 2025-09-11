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

package common

import (
	"context"

	v1 "k8s.io/autoscaler/cluster-autoscaler/apis/capacitybuffer/autoscaling.x-k8s.io/v1"
	client "k8s.io/autoscaler/cluster-autoscaler/apis/capacitybuffer/client/clientset/versioned"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Constants to use in Capacity Buffers objects
const (
	ActiveProvisioningStrategy    = "buffer.x-k8s.io/active-capacity"
	ReadyForProvisioningCondition = "ReadyForProvisioning"
	ProvisioningCondition         = "Provisioning"
	ConditionTrue                 = "True"
	ConditionFalse                = "False"
	DefaultNamespace              = "default"
)

// CreatePodTemplate creates a pod template object by calling API server
func CreatePodTemplate(client *kubernetes.Clientset, podTemplate *corev1.PodTemplate) (*corev1.PodTemplate, error) {
	return client.CoreV1().PodTemplates(DefaultNamespace).Create(context.TODO(), podTemplate, metav1.CreateOptions{})
}

// UpdateBufferStatus updates the passed buffer object with its defined status
func UpdateBufferStatus(buffersClient client.Interface, buffer *v1.CapacityBuffer) error {
	_, err := buffersClient.AutoscalingV1().CapacityBuffers(DefaultNamespace).UpdateStatus(context.TODO(), buffer, metav1.UpdateOptions{})
	return err
}
