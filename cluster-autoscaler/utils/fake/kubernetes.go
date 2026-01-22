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

package fake

import (
	"context"
	apiv1 "k8s.io/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
)

// Kubernetes encapsulates a fake Kubernetes client and its corresponding listers.
type Kubernetes struct {
	Client          *fake.Clientset
	InformerFactory informers.SharedInformerFactory
}

// NewKubernetes creates a new, fully wired fake Kubernetes simulation.
func NewKubernetes(client *fake.Clientset, factory informers.SharedInformerFactory) *Kubernetes {
	return &Kubernetes{
		Client:          client,
		InformerFactory: factory,
	}
}

// AddNode adds a node to the fake client.
func (k *Kubernetes) AddNode(node *apiv1.Node) {
	_, _ = k.Client.CoreV1().Nodes().Create(context.TODO(), node, metav1.CreateOptions{})
}

// UpdateNode updates a node.
func (k *Kubernetes) UpdateNode(node *apiv1.Node) {
	_, _ = k.Client.CoreV1().Nodes().Update(context.TODO(), node, metav1.UpdateOptions{})
}

// DeleteNode deletes a node.
func (k *Kubernetes) DeleteNode(name string) {
	_ = k.Client.CoreV1().Nodes().Delete(context.TODO(), name, metav1.DeleteOptions{})
}

// Nodes lists all available nodes.
func (k *Kubernetes) Nodes() *corev1.NodeList {
	nodes, _ := k.Client.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	return nodes
}

// AddPod adds a pod to the fake client.
func (k *Kubernetes) AddPod(pod *apiv1.Pod) {
	_, _ = k.Client.CoreV1().Pods(pod.Namespace).Create(context.TODO(), pod, metav1.CreateOptions{})
}

// DeletePod deletes a pod.
func (k *Kubernetes) DeletePod(namespace, name string) {
	_ = k.Client.CoreV1().Pods(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
}

// ListerRegistry returns a real ListerRegistry populated with the fake listers.
func (k *Kubernetes) ListerRegistry() kubernetes.ListerRegistry {
	return kubernetes.NewListerRegistryWithDefaultListers(k.InformerFactory)
}
