/*
Copyright The Kubernetes Authors.

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

package listers

import (
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	kube_client "k8s.io/client-go/kubernetes"
	v1lister "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
)

type nodeListerImpl struct {
	nodeLister v1lister.NodeLister
	filters    []func(*corev1.Node) bool
}

// NewNodeLister builds a node lister.
func NewNodeLister(nl v1lister.NodeLister, filters []func(*corev1.Node) bool) v1lister.NodeLister {
	return &nodeListerImpl{
		nodeLister: nl,
		filters:    filters,
	}
}

// List returns list of nodes.
func (l *nodeListerImpl) List(labels labels.Selector) ([]*corev1.Node, error) {
	nodes, err := l.nodeLister.List(labels)
	if err != nil {
		return []*corev1.Node{}, err
	}

	if len(l.filters) != 0 {
		nodes = filterNodes(nodes, l.filters)
	}

	return nodes, nil
}

// Get returns the node with the given name.
func (l *nodeListerImpl) Get(name string) (*corev1.Node, error) {
	node, err := l.nodeLister.Get(name)
	if err != nil {
		return nil, err
	}
	return node, nil
}

func filterNodes(nodes []*corev1.Node, filters []func(*corev1.Node) bool) []*corev1.Node {
	if len(filters) == 0 {
		return nodes
	}
	// TODO: if we will need some filtering in the future this is where we should filter.
	return nodes
}

// NewPodLister creates a new PodLister that lists pods based on the provided kubeClient and namespace.
// It filters out pods that are not scheduled (spec.nodeName is empty) and pods that are in Succeeded or Failed phase, as these pods are not relevant for eviction or in-place updates.
func NewPodLister(kubeClient kube_client.Interface, namespace string, stopCh <-chan struct{}) v1lister.PodLister {
	selector := fields.ParseSelectorOrDie("spec.nodeName!=" + "" + ",status.phase!=" +
		string(corev1.PodSucceeded) + ",status.phase!=" + string(corev1.PodFailed))
	podListWatch := cache.NewListWatchFromClient(kubeClient.CoreV1().RESTClient(), "pods", namespace, selector)
	store := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	podLister := v1lister.NewPodLister(store)
	podReflector := cache.NewReflector(podListWatch, &corev1.Pod{}, store, time.Hour)
	go podReflector.Run(stopCh)

	return podLister
}
