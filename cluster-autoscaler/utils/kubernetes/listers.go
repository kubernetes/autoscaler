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

	apiv1 "k8s.io/api/core/v1"
	extensionsv1 "k8s.io/api/extensions/v1beta1"
	policyv1 "k8s.io/api/policy/v1beta1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	client "k8s.io/client-go/kubernetes"
	v1lister "k8s.io/client-go/listers/core/v1"
	v1extensionslister "k8s.io/client-go/listers/extensions/v1beta1"
	v1policylister "k8s.io/client-go/listers/policy/v1beta1"
	"k8s.io/client-go/tools/cache"
	podv1 "k8s.io/kubernetes/pkg/api/v1/pod"
)

// ListerRegistry is a registry providing various listers to list pods or nodes matching conditions
type ListerRegistry interface {
	AllNodeLister() NodeLister
	ReadyNodeLister() NodeLister
	ScheduledPodLister() PodLister
	UnschedulablePodLister() PodLister
	PodDisruptionBudgetLister() PodDisruptionBudgetLister
	DaemonSetLister() DaemonSetLister
}

type listerRegistryImpl struct {
	allNodeLister             NodeLister
	readyNodeLister           NodeLister
	scheduledPodLister        PodLister
	unschedulablePodLister    PodLister
	podDisruptionBudgetLister PodDisruptionBudgetLister
	daemonSetLister           DaemonSetLister
}

// NewListerRegistry returns a registry providing various listers to list pods or nodes matching conditions
func NewListerRegistry(allNode NodeLister, readyNode NodeLister, scheduledPod PodLister,
	unschedulablePod PodLister, podDisruptionBudgetLister PodDisruptionBudgetLister,
	daemonSetLister DaemonSetLister) ListerRegistry {
	return listerRegistryImpl{
		allNodeLister:             allNode,
		readyNodeLister:           readyNode,
		scheduledPodLister:        scheduledPod,
		unschedulablePodLister:    unschedulablePod,
		podDisruptionBudgetLister: podDisruptionBudgetLister,
		daemonSetLister:           daemonSetLister,
	}
}

// NewListerRegistryWithDefaultListers returns a registry filled with listers of the default implementations
func NewListerRegistryWithDefaultListers(kubeClient client.Interface, stopChannel <-chan struct{}) ListerRegistry {
	unschedulablePodLister := NewUnschedulablePodLister(kubeClient, stopChannel)
	scheduledPodLister := NewScheduledPodLister(kubeClient, stopChannel)
	readyNodeLister := NewReadyNodeLister(kubeClient, stopChannel)
	allNodeLister := NewAllNodeLister(kubeClient, stopChannel)
	podDisruptionBudgetLister := NewPodDisruptionBudgetLister(kubeClient, stopChannel)
	daemonSetLister := NewDaemonSetLister(kubeClient, stopChannel)
	return NewListerRegistry(allNodeLister, readyNodeLister, scheduledPodLister,
		unschedulablePodLister, podDisruptionBudgetLister, daemonSetLister)
}

// AllNodeLister returns the AllNodeLister registered to this registry
func (r listerRegistryImpl) AllNodeLister() NodeLister {
	return r.allNodeLister
}

// ReadyNodeLister returns the ReadyNodeLister registered to this registry
func (r listerRegistryImpl) ReadyNodeLister() NodeLister {
	return r.readyNodeLister
}

// ScheduledPodLister returns the ScheduledPodLister registered to this registry
func (r listerRegistryImpl) ScheduledPodLister() PodLister {
	return r.scheduledPodLister
}

// UnschedulablePodLister returns the UnschedulablePodLister registered to this registry
func (r listerRegistryImpl) UnschedulablePodLister() PodLister {
	return r.unschedulablePodLister
}

// PodDisruptionBudgetLister returns the podDisruptionBudgetLister registered to this registry
func (r listerRegistryImpl) PodDisruptionBudgetLister() PodDisruptionBudgetLister {
	return r.podDisruptionBudgetLister
}

// DaemonSetLister returns the daemonSetLister registered to this registry
func (r listerRegistryImpl) DaemonSetLister() DaemonSetLister {
	return r.daemonSetLister
}

// PodLister lists pods.
type PodLister interface {
	List() ([]*apiv1.Pod, error)
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
		_, condition := podv1.GetPodCondition(&pod.Status, apiv1.PodScheduled)
		if condition != nil && condition.Status == apiv1.ConditionFalse && condition.Reason == apiv1.PodReasonUnschedulable {
			unschedulablePods = append(unschedulablePods, pod)
		}
	}
	return unschedulablePods, nil
}

// NewUnschedulablePodLister returns a lister providing pods that failed to be scheduled.
func NewUnschedulablePodLister(kubeClient client.Interface, stopchannel <-chan struct{}) PodLister {
	return NewUnschedulablePodInNamespaceLister(kubeClient, apiv1.NamespaceAll, stopchannel)
}

// NewUnschedulablePodInNamespaceLister returns a lister providing pods that failed to be scheduled in the given namespace.
func NewUnschedulablePodInNamespaceLister(kubeClient client.Interface, namespace string, stopchannel <-chan struct{}) PodLister {
	// watch unscheduled pods
	selector := fields.ParseSelectorOrDie("spec.nodeName==" + "" + ",status.phase!=" +
		string(apiv1.PodSucceeded) + ",status.phase!=" + string(apiv1.PodFailed))
	podListWatch := cache.NewListWatchFromClient(kubeClient.CoreV1().RESTClient(), "pods", namespace, selector)
	store := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	podLister := v1lister.NewPodLister(store)
	podReflector := cache.NewReflector(podListWatch, &apiv1.Pod{}, store, time.Hour)
	go podReflector.Run(stopchannel)
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
func NewScheduledPodLister(kubeClient client.Interface, stopchannel <-chan struct{}) PodLister {
	// watch unscheduled pods
	selector := fields.ParseSelectorOrDie("spec.nodeName!=" + "" + ",status.phase!=" +
		string(apiv1.PodSucceeded) + ",status.phase!=" + string(apiv1.PodFailed))
	podListWatch := cache.NewListWatchFromClient(kubeClient.CoreV1().RESTClient(), "pods", apiv1.NamespaceAll, selector)
	store := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	podLister := v1lister.NewPodLister(store)
	podReflector := cache.NewReflector(podListWatch, &apiv1.Pod{}, store, time.Hour)
	go podReflector.Run(stopchannel)

	return &ScheduledPodLister{
		podLister: podLister,
	}
}

// NodeLister lists nodes.
type NodeLister interface {
	List() ([]*apiv1.Node, error)
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
		}
	}
	return readyNodes, nil
}

// NewReadyNodeLister builds a node lister.
func NewReadyNodeLister(kubeClient client.Interface, stopChannel <-chan struct{}) NodeLister {
	listWatcher := cache.NewListWatchFromClient(kubeClient.CoreV1().RESTClient(), "nodes", apiv1.NamespaceAll, fields.Everything())
	store := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	nodeLister := v1lister.NewNodeLister(store)
	reflector := cache.NewReflector(listWatcher, &apiv1.Node{}, store, time.Hour)
	go reflector.Run(stopChannel)
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
	allNodes := append(make([]*apiv1.Node, 0, len(nodes)), nodes...)
	return allNodes, nil
}

// NewAllNodeLister builds a node lister that returns all nodes (ready and unready)
func NewAllNodeLister(kubeClient client.Interface, stopchannel <-chan struct{}) NodeLister {
	listWatcher := cache.NewListWatchFromClient(kubeClient.CoreV1().RESTClient(), "nodes", apiv1.NamespaceAll, fields.Everything())
	store := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	nodeLister := v1lister.NewNodeLister(store)
	reflector := cache.NewReflector(listWatcher, &apiv1.Node{}, store, time.Hour)
	go reflector.Run(stopchannel)
	return &AllNodeLister{
		nodeLister: nodeLister,
	}
}

// PodDisruptionBudgetLister lists pod disruption budgets.
type PodDisruptionBudgetLister interface {
	List() ([]*policyv1.PodDisruptionBudget, error)
}

// PodDisruptionBudgetListerImpl lists pod disruption budgets
type PodDisruptionBudgetListerImpl struct {
	pdbLister v1policylister.PodDisruptionBudgetLister
}

// List returns all pdbs
func (lister *PodDisruptionBudgetListerImpl) List() ([]*policyv1.PodDisruptionBudget, error) {
	return lister.pdbLister.List(labels.Everything())
}

// NewPodDisruptionBudgetLister builds a pod disruption budget lister.
func NewPodDisruptionBudgetLister(kubeClient client.Interface, stopchannel <-chan struct{}) PodDisruptionBudgetLister {
	listWatcher := cache.NewListWatchFromClient(kubeClient.Policy().RESTClient(), "poddisruptionbudgets", apiv1.NamespaceAll, fields.Everything())
	store := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	pdbLister := v1policylister.NewPodDisruptionBudgetLister(store)
	reflector := cache.NewReflector(listWatcher, &policyv1.PodDisruptionBudget{}, store, time.Hour)
	go reflector.Run(stopchannel)
	return &PodDisruptionBudgetListerImpl{
		pdbLister: pdbLister,
	}
}

// DaemonSetLister lists daemonsets.
type DaemonSetLister interface {
	List() ([]*extensionsv1.DaemonSet, error)
}

// DaemonSetListerImpl lists all daemonsets.
type DaemonSetListerImpl struct {
	daemonSetLister v1extensionslister.DaemonSetLister
}

// List returns all daemon sets
func (lister *DaemonSetListerImpl) List() ([]*extensionsv1.DaemonSet, error) {
	return lister.daemonSetLister.List(labels.Everything())
}

// NewDaemonSetLister builds a daemonset lister.
func NewDaemonSetLister(kubeClient client.Interface, stopchannel <-chan struct{}) DaemonSetLister {
	listWatcher := cache.NewListWatchFromClient(kubeClient.Extensions().RESTClient(), "daemonsets", apiv1.NamespaceAll, fields.Everything())
	store := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	lister := v1extensionslister.NewDaemonSetLister(store)
	reflector := cache.NewReflector(listWatcher, &extensionsv1.DaemonSet{}, store, time.Hour)
	go reflector.Run(stopchannel)
	return &DaemonSetListerImpl{
		daemonSetLister: lister,
	}
}
