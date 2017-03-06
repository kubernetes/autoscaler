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

package kubernetes

import (
	"time"

	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
	apiv1 "k8s.io/kubernetes/pkg/api/v1"
	policyv1 "k8s.io/kubernetes/pkg/apis/policy/v1beta1"
	client "k8s.io/kubernetes/pkg/client/clientset_generated/clientset"
	v1lister "k8s.io/kubernetes/pkg/client/listers/core/v1"
	v1policylister "k8s.io/kubernetes/pkg/client/listers/policy/v1beta1"
)

// ListerRegistry is a registry providing various listers to list pods or nodes matching conditions
type ListerRegistry interface {
	AllNodeLister() *AllNodeLister
	ReadyNodeLister() *ReadyNodeLister
	ScheduledPodLister() *ScheduledPodLister
	UnschedulablePodLister() *UnschedulablePodLister
	PodDisruptionBudgetLister() *PodDisruptionBudgetLister
}

type listerRegistryImpl struct {
	allNodeLister             *AllNodeLister
	readyNodeLister           *ReadyNodeLister
	scheduledPodLister        *ScheduledPodLister
	unschedulablePodLister    *UnschedulablePodLister
	podDisruptionBudgetLister *PodDisruptionBudgetLister
}

// NewListerRegistry returns a registry providing various listers to list pods or nodes matching conditions
func NewListerRegistry(allNode *AllNodeLister, readyNode *ReadyNodeLister, scheduledPod *ScheduledPodLister,
	unschedulablePod *UnschedulablePodLister, podDisruptionBudgetLister *PodDisruptionBudgetLister) ListerRegistry {
	return listerRegistryImpl{
		allNodeLister:             allNode,
		readyNodeLister:           readyNode,
		scheduledPodLister:        scheduledPod,
		unschedulablePodLister:    unschedulablePod,
		podDisruptionBudgetLister: podDisruptionBudgetLister,
	}
}

// NewListerRegistryWithDefaultListers returns a registry filled with listers of the default implementations
func NewListerRegistryWithDefaultListers(kubeClient client.Interface, stopChannel <-chan struct{}) ListerRegistry {
	unschedulablePodLister := NewUnschedulablePodLister(kubeClient, stopChannel)
	scheduledPodLister := NewScheduledPodLister(kubeClient, stopChannel)
	readyNodeLister := NewReadyNodeLister(kubeClient, stopChannel)
	allNodeLister := NewAllNodeLister(kubeClient, stopChannel)
	podDisruptionBudgetLister := NewPodDisruptionBudgetLister(kubeClient, stopChannel)
	return NewListerRegistry(allNodeLister, readyNodeLister, scheduledPodLister,
		unschedulablePodLister, podDisruptionBudgetLister)
}

// AllNodeLister returns the AllNodeLister registered to this registry
func (r listerRegistryImpl) AllNodeLister() *AllNodeLister {
	return r.allNodeLister
}

// ReadyNodeLister returns the ReadyNodeLister registered to this registry
func (r listerRegistryImpl) ReadyNodeLister() *ReadyNodeLister {
	return r.readyNodeLister
}

// ScheduledPodLister returns the ScheduledPodLister registered to this registry
func (r listerRegistryImpl) ScheduledPodLister() *ScheduledPodLister {
	return r.scheduledPodLister
}

// UnschedulablePodLister returns the UnschedulablePodLister registered to this registry
func (r listerRegistryImpl) UnschedulablePodLister() *UnschedulablePodLister {
	return r.unschedulablePodLister
}

// PodDisruptionBudgetLister returns the podDisruptionBudgetLister registered to this registry
func (r listerRegistryImpl) PodDisruptionBudgetLister() *PodDisruptionBudgetLister {
	return r.podDisruptionBudgetLister
}

// UnschedulablePodLister lists unscheduled pods
type UnschedulablePodLister struct {
	podLister v1lister.PodLister
}

// List returns all unscheduled pods.
func (unschedulablePodLister *UnschedulablePodLister) List() ([]*apiv1.Pod, error) {
	var unschedulablePods []*apiv1.Pod
	allPods, err := unschedulablePodLister.podLister.List(labels.Everything())
	if err != nil {
		return unschedulablePods, err
	}
	for _, pod := range allPods {
		_, condition := apiv1.GetPodCondition(&pod.Status, apiv1.PodScheduled)
		if condition != nil && condition.Status == apiv1.ConditionFalse && condition.Reason == "Unschedulable" {
			unschedulablePods = append(unschedulablePods, pod)
		}
	}
	return unschedulablePods, nil
}

// NewUnschedulablePodLister returns a lister providing pods that failed to be scheduled.
func NewUnschedulablePodLister(kubeClient client.Interface, stopchannel <-chan struct{}) *UnschedulablePodLister {
	return NewUnschedulablePodInNamespaceLister(kubeClient, apiv1.NamespaceAll, stopchannel)
}

// NewUnschedulablePodInNamespaceLister returns a lister providing pods that failed to be scheduled in the given namespace.
func NewUnschedulablePodInNamespaceLister(kubeClient client.Interface, namespace string, stopchannel <-chan struct{}) *UnschedulablePodLister {
	// watch unscheduled pods
	selector := fields.ParseSelectorOrDie("spec.nodeName==" + "" + ",status.phase!=" +
		string(apiv1.PodSucceeded) + ",status.phase!=" + string(apiv1.PodFailed))
	podListWatch := cache.NewListWatchFromClient(kubeClient.Core().RESTClient(), "pods", namespace, selector)
	store := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	podLister := v1lister.NewPodLister(store)
	podReflector := cache.NewReflector(podListWatch, &apiv1.Pod{}, store, time.Hour)
	podReflector.RunUntil(stopchannel)
	return &UnschedulablePodLister{
		podLister: podLister,
	}
}

// ScheduledPodLister lists scheduled pods.
type ScheduledPodLister struct {
	podLister v1lister.PodLister
}

// List returns all scheduled pods.
func (lister *ScheduledPodLister) List() ([]*apiv1.Pod, error) {
	return lister.podLister.List(labels.Everything())
}

// NewScheduledPodLister builds ScheduledPodLister
func NewScheduledPodLister(kubeClient client.Interface, stopchannel <-chan struct{}) *ScheduledPodLister {
	// watch unscheduled pods
	selector := fields.ParseSelectorOrDie("spec.nodeName!=" + "" + ",status.phase!=" +
		string(apiv1.PodSucceeded) + ",status.phase!=" + string(apiv1.PodFailed))
	podListWatch := cache.NewListWatchFromClient(kubeClient.Core().RESTClient(), "pods", apiv1.NamespaceAll, selector)
	store := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	podLister := v1lister.NewPodLister(store)
	podReflector := cache.NewReflector(podListWatch, &apiv1.Pod{}, store, time.Hour)
	podReflector.RunUntil(stopchannel)

	return &ScheduledPodLister{
		podLister: podLister,
	}
}

// ReadyNodeLister lists ready nodes.
type ReadyNodeLister struct {
	nodeLister v1lister.NodeLister
}

// List returns ready nodes.
func (readyNodeLister *ReadyNodeLister) List() ([]*apiv1.Node, error) {
	nodes, err := readyNodeLister.nodeLister.List(labels.Everything())
	if err != nil {
		return []*apiv1.Node{}, err
	}
	readyNodes := make([]*apiv1.Node, 0, len(nodes))
	for _, node := range nodes {
		if IsNodeReadyAndSchedulable(node) {
			readyNodes = append(readyNodes, node)
			break
		}
	}
	return readyNodes, nil
}

// NewReadyNodeLister builds a node lister.
func NewReadyNodeLister(kubeClient client.Interface, stopChannel <-chan struct{}) *ReadyNodeLister {
	listWatcher := cache.NewListWatchFromClient(kubeClient.Core().RESTClient(), "nodes", apiv1.NamespaceAll, fields.Everything())
	store := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	nodeLister := v1lister.NewNodeLister(store)
	reflector := cache.NewReflector(listWatcher, &apiv1.Node{}, store, time.Hour)
	reflector.RunUntil(stopChannel)
	return &ReadyNodeLister{
		nodeLister: nodeLister,
	}
}

// AllNodeLister lists all nodes
type AllNodeLister struct {
	nodeLister v1lister.NodeLister
}

// List returns all nodes
func (allNodeLister *AllNodeLister) List() ([]*apiv1.Node, error) {
	nodes, err := allNodeLister.nodeLister.List(labels.Everything())
	if err != nil {
		return []*apiv1.Node{}, err
	}
	allNodes := make([]*apiv1.Node, 0, len(nodes))
	for _, node := range nodes {
		allNodes = append(allNodes, node)
	}
	return allNodes, nil
}

// NewAllNodeLister builds a node lister that returns all nodes (ready and unready)
func NewAllNodeLister(kubeClient client.Interface, stopchannel <-chan struct{}) *AllNodeLister {
	listWatcher := cache.NewListWatchFromClient(kubeClient.Core().RESTClient(), "nodes", apiv1.NamespaceAll, fields.Everything())
	store := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	nodeLister := v1lister.NewNodeLister(store)
	reflector := cache.NewReflector(listWatcher, &apiv1.Node{}, store, time.Hour)
	reflector.RunUntil(stopchannel)
	return &AllNodeLister{
		nodeLister: nodeLister,
	}
}

// PodDisruptionBudgetLister lists all pod disruption budgets
type PodDisruptionBudgetLister struct {
	pdbLister v1policylister.PodDisruptionBudgetLister
}

// List returns all nodes
func (lister *PodDisruptionBudgetLister) List() ([]*policyv1.PodDisruptionBudget, error) {
	return lister.pdbLister.List(labels.Everything())
}

// NewPodDisruptionBudgetLister builds a pod disruption budget lister.
func NewPodDisruptionBudgetLister(kubeClient client.Interface, stopchannel <-chan struct{}) *PodDisruptionBudgetLister {
	listWatcher := cache.NewListWatchFromClient(kubeClient.Policy().RESTClient(), "podpisruptionbudgets", apiv1.NamespaceAll, fields.Everything())
	store := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	pdbLister := v1policylister.NewPodDisruptionBudgetLister(store)
	reflector := cache.NewReflector(listWatcher, &policyv1.PodDisruptionBudget{}, store, time.Hour)
	reflector.RunUntil(stopchannel)
	return &PodDisruptionBudgetLister{
		pdbLister: pdbLister,
	}
}
