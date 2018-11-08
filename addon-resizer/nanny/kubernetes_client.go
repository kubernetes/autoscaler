/*
Copyright 2016 The Kubernetes Authors.

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

package nanny

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/client-go/kubernetes"
	scheme "k8s.io/client-go/kubernetes/scheme"
)

type kubernetesClient struct {
	namespace  string
	deployment string
	pod        string
	container  string
	clientset  *kubernetes.Clientset
}

func (k *kubernetesClient) CountNodes() (uint64, error) {
	// Set ResourceVersion = 0 to use cached versions.
	options := metav1.ListOptions{
		ResourceVersion: "0",
	}
	result := &metav1beta1.PartialObjectMetadataList{}
	err := k.clientset.
		Core().
		RESTClient().
		Get().
		Resource("nodes").
		// Set as=PartialObjectMetadataList to fetch only nodes metadata.
		SetHeader("Accept", "application/vnd.kubernetes.protobuf;as=PartialObjectMetadataList;g=meta.k8s.io;v=v1beta1").
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return uint64(len(result.Items)), err
}

func (k *kubernetesClient) ContainerResources() (*corev1.ResourceRequirements, error) {
	pod, err := k.clientset.Core().Pods(k.namespace).Get(k.pod, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	for _, container := range pod.Spec.Containers {
		if container.Name == k.container {
			return &container.Resources, nil
		}
	}
	return nil, fmt.Errorf("Container %s was not found in deployment %s in namespace %s.", k.container, k.deployment, k.namespace)
}

func (k *kubernetesClient) UpdateDeployment(resources *corev1.ResourceRequirements) error {
	// First, get the Deployment.
	dep, err := k.clientset.Extensions().Deployments(k.namespace).Get(k.deployment, metav1.GetOptions{})
	if err != nil {
		return err
	}

	// Modify the Deployment object with our ResourceRequirements.
	for i, container := range dep.Spec.Template.Spec.Containers {
		if container.Name == k.container {
			// Update the deployment.
			dep.Spec.Template.Spec.Containers[i].Resources = *resources
			_, err = k.clientset.Extensions().Deployments(k.namespace).Update(dep)
			return err
		}
	}

	return fmt.Errorf("Container %s was not found in the deployment %s in namespace %s.", k.container, k.deployment, k.namespace)
}

// NewKubernetesClient gives a KubernetesClient with the given dependencies.
func NewKubernetesClient(namespace, deployment, pod, container string, clientset *kubernetes.Clientset) KubernetesClient {
	result := &kubernetesClient{
		namespace:  namespace,
		deployment: deployment,
		pod:        pod,
		container:  container,
		clientset:  clientset,
	}
	return result
}
