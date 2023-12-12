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
	policyv1 "k8s.io/api/policy/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	client "k8s.io/client-go/kubernetes"
	v1appslister "k8s.io/client-go/listers/apps/v1"
	v1batchlister "k8s.io/client-go/listers/batch/v1"
	v1lister "k8s.io/client-go/listers/core/v1"
	v1policylister "k8s.io/client-go/listers/policy/v1"
	"k8s.io/client-go/tools/cache"
	podv1 "k8s.io/kubernetes/pkg/api/v1/pod"
)

// ListerRegistry is a registry providing various listers to list pods or nodes matching conditions
type ListerRegistry interface {
	AllNodeLister() NodeLister
	ReadyNodeLister() NodeLister
	AllPodLister() PodLister
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
	allPodLister                PodLister
	podDisruptionBudgetLister   PodDisruptionBudgetLister
	daemonSetLister             v1appslister.DaemonSetLister
	replicationControllerLister v1lister.ReplicationControllerLister
	jobLister                   v1batchlister.JobLister
	replicaSetLister            v1appslister.ReplicaSetLister
	statefulSetLister           v1appslister.StatefulSetLister
}

// NewListerRegistry returns a registry providing various listers to list pods or nodes matching conditions
func NewListerRegistry(allNode NodeLister, readyNode NodeLister, allPodLister PodLister, podDisruptionBudgetLister PodDisruptionBudgetLister,
	daemonSetLister v1appslister.DaemonSetLister, replicationControllerLister v1lister.ReplicationControllerLister,
	jobLister v1batchlister.JobLister, replicaSetLister v1appslister.ReplicaSetLister,
	statefulSetLister v1appslister.StatefulSetLister) ListerRegistry {
	return listerRegistryImpl{
		allNodeLister:               allNode,
		readyNodeLister:             readyNode,
		allPodLister:                allPodLister,
		podDisruptionBudgetLister:   podDisruptionBudgetLister,
		daemonSetLister:             daemonSetLister,
		replicationControllerLister: replicationControllerLister,
		jobLister:                   jobLister,
		replicaSetLister:            replicaSetLister,
		statefulSetLister:           statefulSetLister,
	}
}

// NewListerRegistryWithDefaultListers returns a registry filled with listers of the default implementations
func NewListerRegistryWithDefaultListers(informerFactory informers.SharedInformerFactory) ListerRegistry {
	allPodLister := NewAllPodLister(informerFactory.Core().V1().Pods().Lister())
	readyNodeLister := NewReadyNodeLister(informerFactory.Core().V1().Nodes().Lister())
	allNodeLister := NewAllNodeLister(informerFactory.Core().V1().Nodes().Lister())

	podDisruptionBudgetLister := NewPodDisruptionBudgetLister(informerFactory.Policy().V1().PodDisruptionBudgets().Lister())
	daemonSetLister := informerFactory.Apps().V1().DaemonSets().Lister()
	replicationControllerLister := informerFactory.Core().V1().ReplicationControllers().Lister()
	jobLister := informerFactory.Batch().V1().Jobs().Lister()
	replicaSetLister := informerFactory.Apps().V1().ReplicaSets().Lister()
	statefulSetLister := informerFactory.Apps().V1().StatefulSets().Lister()
	return NewListerRegistry(allNodeLister, readyNodeLister, allPodLister,
		podDisruptionBudgetLister, daemonSetLister, replicationControllerLister,
		jobLister, replicaSetLister, statefulSetLister)
}

// AllPodLister returns the AllPodLister registered to this registry
func (r listerRegistryImpl) AllPodLister() PodLister {
	return r.allPodLister
}

// AllNodeLister returns the AllNodeLister registered to this registry
func (r listerRegistryImpl) AllNodeLister() NodeLister {
	return r.allNodeLister
}

// ReadyNodeLister returns the ReadyNodeLister registered to this registry
func (r listerRegistryImpl) ReadyNodeLister() NodeLister {
	return r.readyNodeLister
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

// PodLister lists all pods.
// To filter out the scheduled or unschedulable pods the helper methods ScheduledPods and UnschedulablePods should be used.
type PodLister interface {
	List() ([]*apiv1.Pod, error)
}

// isScheduled checks whether a pod is scheduled on a node or not
// This method doesn't check for nil ptr, it's the responsibility of the caller
func isScheduled(pod *apiv1.Pod) bool {
	return pod.Spec.NodeName != ""
}

// isDeleted checks whether a pod is deleted not
// This method doesn't check for nil ptr, it's the responsibility of the caller
func isDeleted(pod *apiv1.Pod) bool {
	return pod.GetDeletionTimestamp() != nil
}

// isUnschedulable checks whether a pod is unschedulable or not
// This method doesn't check for nil ptr, it's the responsibility of the caller
func isUnschedulable(pod *apiv1.Pod) bool {
	if isScheduled(pod) || isDeleted(pod) {
		return false
	}
	_, condition := podv1.GetPodCondition(&pod.Status, apiv1.PodScheduled)
	if condition == nil || condition.Status != apiv1.ConditionFalse || condition.Reason != apiv1.PodReasonUnschedulable {
		return false
	}
	return true
}

// ScheduledPods is a helper method that returns all scheduled pods from given pod list.
func ScheduledPods(allPods []*apiv1.Pod) []*apiv1.Pod {
	var scheduledPods []*apiv1.Pod
	for _, pod := range allPods {
		if isScheduled(pod) {
			scheduledPods = append(scheduledPods, pod)
			continue
		}
	}
	return scheduledPods
}

// SchedulerUnprocessedPods is a helper method that returns all pods which are not yet processed by the specified bypassed schedulers
func SchedulerUnprocessedPods(allPods []*apiv1.Pod, bypassedSchedulers map[string]bool) []*apiv1.Pod {
	var unprocessedPods []*apiv1.Pod

	for _, pod := range allPods {
		if canBypass := bypassedSchedulers[pod.Spec.SchedulerName]; !canBypass {
			continue
		}
		// Make sure it's not scheduled or deleted
		if isScheduled(pod) || isDeleted(pod) || isUnschedulable(pod) {
			continue
		}
		// Make sure that if it's not scheduled it's either
		// Not processed (condition is nil)
		// Or Reason is empty (not schedulerError, terminated, ...etc)
		_, condition := podv1.GetPodCondition(&pod.Status, apiv1.PodScheduled)
		if condition == nil || (condition.Status == apiv1.ConditionFalse && condition.Reason == "") {
			unprocessedPods = append(unprocessedPods, pod)
		}
	}
	return unprocessedPods
}

// UnschedulablePods is a helper method that returns all unschedulable pods from given pod list.
func UnschedulablePods(allPods []*apiv1.Pod) []*apiv1.Pod {
	var unschedulablePods []*apiv1.Pod
	for _, pod := range allPods {
		if !isUnschedulable(pod) {
			continue
		}
		unschedulablePods = append(unschedulablePods, pod)
	}
	return unschedulablePods
}

// AllPodLister lists all pods.
type AllPodLister struct {
	podLister v1lister.PodLister
}

// List returns all scheduled pods.
func (lister *AllPodLister) List() ([]*apiv1.Pod, error) {
	var pods []*apiv1.Pod

	allPods, err := lister.podLister.List(labels.Everything())
	if err != nil {
		return pods, err
	}
	for _, p := range allPods {
		if p.Status.Phase != apiv1.PodSucceeded && p.Status.Phase != apiv1.PodFailed {
			pods = append(pods, p)
		}
	}
	return pods, nil
}

// NewAllPodLister builds AllPodLister
func NewAllPodLister(pl v1lister.PodLister) PodLister {
	return &AllPodLister{
		podLister: pl,
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

// NewAllNodeLister builds a node lister that returns all nodes (ready and unready).
func NewAllNodeLister(nl v1lister.NodeLister) NodeLister {
	return NewNodeLister(nl, nil)
}

// NewReadyNodeLister builds a node lister that returns only ready nodes.
func NewReadyNodeLister(nl v1lister.NodeLister) NodeLister {
	return NewNodeLister(nl, IsNodeReadyAndSchedulable)
}

// NewNodeLister builds a node lister.
func NewNodeLister(nl v1lister.NodeLister, filter func(*apiv1.Node) bool) NodeLister {
	return &nodeListerImpl{
		nodeLister: nl,
		filter:     filter,
	}
}

// List returns list of nodes.
func (l *nodeListerImpl) List() ([]*apiv1.Node, error) {
	var nodes []*apiv1.Node
	var err error

	nodes, err = l.nodeLister.List(labels.Everything())
	if err != nil {
		return []*apiv1.Node{}, err
	}

	if l.filter != nil {
		nodes = filterNodes(nodes, l.filter)
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

func filterNodes(nodes []*apiv1.Node, predicate func(*apiv1.Node) bool) []*apiv1.Node {
	var filtered []*apiv1.Node
	for i := range nodes {
		if predicate(nodes[i]) {
			filtered = append(filtered, nodes[i])
		}
	}
	return filtered
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
func NewPodDisruptionBudgetLister(pdbLister v1policylister.PodDisruptionBudgetLister) PodDisruptionBudgetLister {
	return &PodDisruptionBudgetListerImpl{
		pdbLister: pdbLister,
	}
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
