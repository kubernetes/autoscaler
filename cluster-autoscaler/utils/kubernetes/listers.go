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

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	apiv1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1beta1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	client "k8s.io/client-go/kubernetes"
	v1appslister "k8s.io/client-go/listers/apps/v1"
	v1batchlister "k8s.io/client-go/listers/batch/v1"
	v1lister "k8s.io/client-go/listers/core/v1"
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
	DaemonSetLister() v1appslister.DaemonSetLister
	ReplicationControllerLister() v1lister.ReplicationControllerLister
	JobLister() v1batchlister.JobLister
	ReplicaSetLister() v1appslister.ReplicaSetLister
	StatefulSetLister() v1appslister.StatefulSetLister
}

type listerRegistryImpl struct {
	allNodeLister               NodeLister
	readyNodeLister             NodeLister
	scheduledPodLister          PodLister
	unschedulablePodLister      PodLister
	podDisruptionBudgetLister   PodDisruptionBudgetLister
	daemonSetLister             v1appslister.DaemonSetLister
	replicationControllerLister v1lister.ReplicationControllerLister
	jobLister                   v1batchlister.JobLister
	replicaSetLister            v1appslister.ReplicaSetLister
	statefulSetLister           v1appslister.StatefulSetLister
}

// NewListerRegistry returns a registry providing various listers to list pods or nodes matching conditions
func NewListerRegistry(allNode NodeLister, readyNode NodeLister, scheduledPod PodLister,
	unschedulablePod PodLister, podDisruptionBudgetLister PodDisruptionBudgetLister,
	daemonSetLister v1appslister.DaemonSetLister, replicationControllerLister v1lister.ReplicationControllerLister,
	jobLister v1batchlister.JobLister, replicaSetLister v1appslister.ReplicaSetLister,
	statefulSetLister v1appslister.StatefulSetLister) ListerRegistry {
	return listerRegistryImpl{
		allNodeLister:               allNode,
		readyNodeLister:             readyNode,
		scheduledPodLister:          scheduledPod,
		unschedulablePodLister:      unschedulablePod,
		podDisruptionBudgetLister:   podDisruptionBudgetLister,
		daemonSetLister:             daemonSetLister,
		replicationControllerLister: replicationControllerLister,
		jobLister:                   jobLister,
		replicaSetLister:            replicaSetLister,
		statefulSetLister:           statefulSetLister,
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
	replicationControllerLister := NewReplicationControllerLister(kubeClient, stopChannel)
	jobLister := NewJobLister(kubeClient, stopChannel)
	replicaSetLister := NewReplicaSetLister(kubeClient, stopChannel)
	statefulSetLister := NewStatefulSetLister(kubeClient, stopChannel)
	return NewListerRegistry(allNodeLister, readyNodeLister, scheduledPodLister,
		unschedulablePodLister, podDisruptionBudgetLister, daemonSetLister,
		replicationControllerLister, jobLister, replicaSetLister, statefulSetLister)
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
func (r listerRegistryImpl) DaemonSetLister() v1appslister.DaemonSetLister {
	return r.daemonSetLister
}

// ReplicationControllerLister returns the replicationControllerLister registered to this registry
func (r listerRegistryImpl) ReplicationControllerLister() v1lister.ReplicationControllerLister {
	return r.replicationControllerLister
}

// JobLister returns the jobLister registered to this registry
func (r listerRegistryImpl) JobLister() v1batchlister.JobLister {
	return r.jobLister
}

// ReplicaSetLister returns the replicaSetLister registered to this registry
func (r listerRegistryImpl) ReplicaSetLister() v1appslister.ReplicaSetLister {
	return r.replicaSetLister
}

// StatefulSetLister returns the statefulSetLister registered to this registry
func (r listerRegistryImpl) StatefulSetLister() v1appslister.StatefulSetLister {
	return r.statefulSetLister
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
	store, reflector := cache.NewNamespaceKeyedIndexerAndReflector(podListWatch, &apiv1.Pod{}, time.Hour)
	podLister := v1lister.NewPodLister(store)
	go reflector.Run(stopchannel)
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
	store, reflector := cache.NewNamespaceKeyedIndexerAndReflector(podListWatch, &apiv1.Pod{}, time.Hour)
	podLister := v1lister.NewPodLister(store)
	go reflector.Run(stopchannel)

	return &ScheduledPodLister{
		podLister: podLister,
	}
}

// NodeLister lists nodes.
type NodeLister interface {
	List() ([]*apiv1.Node, error)
	Get(name string) (*apiv1.Node, error)
}

// nodeLister implementation.
type nodeListerImpl struct {
	nodeLister v1lister.NodeLister
	filter     func(*apiv1.Node) bool
}

// NewReadyNodeLister builds a node lister that returns only ready nodes.
func NewReadyNodeLister(kubeClient client.Interface, stopChannel <-chan struct{}) NodeLister {
	return NewNodeLister(kubeClient, IsNodeReadyAndSchedulable, stopChannel)
}

// NewAllNodeLister builds a node lister that returns all nodes (ready and unready).
func NewAllNodeLister(kubeClient client.Interface, stopChannel <-chan struct{}) NodeLister {
	return NewNodeLister(kubeClient, nil, stopChannel)
}

// NewNodeLister builds a node lister.
func NewNodeLister(kubeClient client.Interface, filter func(*apiv1.Node) bool, stopChannel <-chan struct{}) NodeLister {
	listWatcher := cache.NewListWatchFromClient(kubeClient.CoreV1().RESTClient(), "nodes", apiv1.NamespaceAll, fields.Everything())
	store, reflector := cache.NewNamespaceKeyedIndexerAndReflector(listWatcher, &apiv1.Node{}, time.Hour)
	nodeLister := v1lister.NewNodeLister(store)
	go reflector.Run(stopChannel)
	return &nodeListerImpl{
		nodeLister: nodeLister,
		filter:     filter,
	}
}

// List returns list of nodes.
func (l *nodeListerImpl) List() ([]*apiv1.Node, error) {
	nodes, err := l.nodeLister.List(labels.Everything())
	if err != nil {
		return []*apiv1.Node{}, err
	}
	results := make([]*apiv1.Node, 0, len(nodes))
	if l.filter != nil {
		for _, node := range nodes {
			if l.filter(node) {
				results = append(results, node)
			}
		}
	} else {
		results = append(results, nodes...)
	}
	return nodes, nil
}

// Get returns the node with the given name.
func (l *nodeListerImpl) Get(name string) (*apiv1.Node, error) {
	node, err := l.nodeLister.Get(name)
	if err != nil {
		return nil, err
	}
	return node, nil
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
	listWatcher := cache.NewListWatchFromClient(kubeClient.PolicyV1beta1().RESTClient(), "poddisruptionbudgets", apiv1.NamespaceAll, fields.Everything())
	store, reflector := cache.NewNamespaceKeyedIndexerAndReflector(listWatcher, &policyv1.PodDisruptionBudget{}, time.Hour)
	pdbLister := v1policylister.NewPodDisruptionBudgetLister(store)
	go reflector.Run(stopchannel)
	return &PodDisruptionBudgetListerImpl{
		pdbLister: pdbLister,
	}
}

// NewDaemonSetLister builds a daemonset lister.
func NewDaemonSetLister(kubeClient client.Interface, stopchannel <-chan struct{}) v1appslister.DaemonSetLister {
	listWatcher := cache.NewListWatchFromClient(kubeClient.AppsV1().RESTClient(), "daemonsets", apiv1.NamespaceAll, fields.Everything())
	store, reflector := cache.NewNamespaceKeyedIndexerAndReflector(listWatcher, &appsv1.DaemonSet{}, time.Hour)
	lister := v1appslister.NewDaemonSetLister(store)
	go reflector.Run(stopchannel)
	return lister
}

// NewReplicationControllerLister builds a replicationcontroller lister.
func NewReplicationControllerLister(kubeClient client.Interface, stopchannel <-chan struct{}) v1lister.ReplicationControllerLister {
	listWatcher := cache.NewListWatchFromClient(kubeClient.CoreV1().RESTClient(), "replicationcontrollers", apiv1.NamespaceAll, fields.Everything())
	store, reflector := cache.NewNamespaceKeyedIndexerAndReflector(listWatcher, &apiv1.ReplicationController{}, time.Hour)
	lister := v1lister.NewReplicationControllerLister(store)
	go reflector.Run(stopchannel)
	return lister
}

// NewJobLister builds a job lister.
func NewJobLister(kubeClient client.Interface, stopchannel <-chan struct{}) v1batchlister.JobLister {
	listWatcher := cache.NewListWatchFromClient(kubeClient.BatchV1().RESTClient(), "jobs", apiv1.NamespaceAll, fields.Everything())
	store, reflector := cache.NewNamespaceKeyedIndexerAndReflector(listWatcher, &batchv1.Job{}, time.Hour)
	lister := v1batchlister.NewJobLister(store)
	go reflector.Run(stopchannel)
	return lister
}

// NewReplicaSetLister builds a replicaset lister.
func NewReplicaSetLister(kubeClient client.Interface, stopchannel <-chan struct{}) v1appslister.ReplicaSetLister {
	listWatcher := cache.NewListWatchFromClient(kubeClient.AppsV1().RESTClient(), "replicasets", apiv1.NamespaceAll, fields.Everything())
	store, reflector := cache.NewNamespaceKeyedIndexerAndReflector(listWatcher, &appsv1.ReplicaSet{}, time.Hour)
	lister := v1appslister.NewReplicaSetLister(store)
	go reflector.Run(stopchannel)
	return lister
}

// NewStatefulSetLister builds a statefulset lister.
func NewStatefulSetLister(kubeClient client.Interface, stopchannel <-chan struct{}) v1appslister.StatefulSetLister {
	listWatcher := cache.NewListWatchFromClient(kubeClient.AppsV1().RESTClient(), "statefulsets", apiv1.NamespaceAll, fields.Everything())
	store, reflector := cache.NewNamespaceKeyedIndexerAndReflector(listWatcher, &appsv1.StatefulSet{}, time.Hour)
	lister := v1appslister.NewStatefulSetLister(store)
	go reflector.Run(stopchannel)
	return lister
}

// NewConfigMapListerForNamespace builds a configmap lister for the passed namespace (including all).
func NewConfigMapListerForNamespace(kubeClient client.Interface, stopchannel <-chan struct{},
	namespace string) v1lister.ConfigMapLister {
	listWatcher := cache.NewListWatchFromClient(kubeClient.CoreV1().RESTClient(), "configmaps", namespace, fields.Everything())
	store, reflector := cache.NewNamespaceKeyedIndexerAndReflector(listWatcher, &apiv1.ConfigMap{}, time.Hour)
	lister := v1lister.NewConfigMapLister(store)
	go reflector.Run(stopchannel)
	return lister
}
