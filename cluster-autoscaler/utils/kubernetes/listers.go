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
	"context"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	apiv1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	dynamiclister "k8s.io/client-go/dynamic/dynamiclister"
	client "k8s.io/client-go/kubernetes"
	v1appslister "k8s.io/client-go/listers/apps/v1"
	v1batchlister "k8s.io/client-go/listers/batch/v1"
	v1lister "k8s.io/client-go/listers/core/v1"
	v1policylister "k8s.io/client-go/listers/policy/v1"
	"k8s.io/client-go/tools/cache"
	klog "k8s.io/klog/v2"
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
	GetLister(gvr schema.GroupVersionKind, namespace string) dynamiclister.Lister
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
	stopCh                      <-chan struct{}
	listersMap                  map[string]dynamiclister.Lister
	dynamicClient               *dynamic.DynamicClient
	// Only used to find plural names for resource Kinds
	// by looking up the CRD's `spec.names.plural` field
	crdLister dynamiclister.Lister
}

// GetLister returns the lister for a particular GVR
func (g listerRegistryImpl) GetLister(gvr schema.GroupVersionKind, namespace string) dynamiclister.Lister {
	crd, err := g.crdLister.Get(fmt.Sprintf("%s/%s", gvr.Group, gvr.Kind))
	if err != nil {
		fmt.Println(fmt.Errorf("crd not found: %v", err))
	}

	resource, found, err := unstructured.NestedString(crd.Object, "spec", "names", "plural")
	if !found {
		fmt.Println(fmt.Errorf("couldn't find the field 'spec.names.plural' on the CRD '%s'", crd.GetName()))
	}
	if err != nil {
		fmt.Errorf("error retrieving the field `spec.names.plural` on the CRD '%s': %v", crd.GetName(), err)
	}

	return NewGenericLister(g.dynamicClient, g.listersMap, g.stopCh, schema.GroupVersionResource{Group: gvr.Group,
		Version:  gvr.Version,
		Resource: resource}, namespace)
}

// NewListerRegistry returns a registry providing various listers to list pods or nodes matching conditions
func NewListerRegistry(allNode NodeLister, readyNode NodeLister, scheduledPod PodLister,
	unschedulablePod PodLister, podDisruptionBudgetLister PodDisruptionBudgetLister,
	daemonSetLister v1appslister.DaemonSetLister, replicationControllerLister v1lister.ReplicationControllerLister,
	jobLister v1batchlister.JobLister, replicaSetLister v1appslister.ReplicaSetLister,
	statefulSetLister v1appslister.StatefulSetLister,
	// genericListerFactory GenericListerFactory) ListerRegistry {
	stopCh <-chan struct{},
	listersMap map[string]dynamiclister.Lister,
	dynamicClient *dynamic.DynamicClient,
	crdLister dynamiclister.Lister) ListerRegistry {
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
		// genericListerFactory:        genericListerFactory,
		stopCh:        stopCh,
		listersMap:    listersMap,
		dynamicClient: dynamicClient,
		crdLister:     crdLister,
	}
}

// NewListerRegistryWithDefaultListers returns a registry filled with listers of the default implementations
func NewListerRegistryWithDefaultListers(kubeClient client.Interface, dynamicClient *dynamic.DynamicClient, stopChannel <-chan struct{}) ListerRegistry {
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
	listersMap := make(map[string]dynamiclister.Lister)
	crdLister := NewDynamicCRDLister(dynamicClient, stopChannel)
	// genericListerFactory := NewGenericListerFactory(dynamicClient, stopChannel)
	return NewListerRegistry(allNodeLister, readyNodeLister, scheduledPodLister,
		unschedulablePodLister, podDisruptionBudgetLister, daemonSetLister,
		replicationControllerLister, jobLister, replicaSetLister, statefulSetLister,
		//  genericListerFactory
		stopChannel,
		listersMap,
		dynamicClient,
		crdLister,
	)
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

// NewDynamicCRDLister returns a lister providing CRDs with key `<api-group>/<Kind>` instead of `<namespace>/<name>` for
// easy querying based on Kind
func NewDynamicCRDLister(dClient *dynamic.DynamicClient, stopChannel <-chan struct{}) dynamiclister.Lister {

	var lister func(ctx context.Context, opts metav1.ListOptions) (*unstructured.UnstructuredList, error)
	var watcher func(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)

	gvr := schema.GroupVersionResource{Group: "apiextensions.k8s.io", Version: "v1", Resource: "customresourcedefinitions"}

	lister = dClient.Resource(gvr).List
	watcher = dClient.Resource(gvr).Watch
	store := cache.NewIndexer( /* Key Func*/ func(obj interface{}) (string, error) {
		uo := obj.(*unstructured.Unstructured)
		o := uo.Object
		group, found, err := unstructured.NestedString(o, "spec", "group")
		if !found {
			fmt.Printf("didn't find value on %v", uo.GetName())
		}
		if err != nil {
			fmt.Printf("err: %v", err)
		}

		names, found, err := unstructured.NestedStringMap(o, "spec", "names")
		if !found {
			fmt.Printf("didn't find value on %v", uo.GetName())
		}
		if err != nil {
			fmt.Printf("err: %v", err)
		}

		// Key is <group>/<Kind> as opposed to <namespace>/name
		// This is so that you can find CRD just using Kind and API Group
		// instead of knowing the name
		return group + "/" + names["kind"], nil
	}, cache.Indexers{"group": /* Index Func */ func(obj interface{}) ([]string, error) {
		uo := obj.(*unstructured.Unstructured)
		o := uo.Object
		group, found, err := unstructured.NestedString(o, "spec", "group")
		if !found {
			fmt.Printf("didn't find value on %v", uo.GetName())
		}
		if err != nil {
			return []string{""}, fmt.Errorf("err: %v", err)
		}
		/* Index by APi Group of the CRD */
		return []string{group}, nil
	}})

	lw := &cache.ListWatch{
		ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
			return lister(context.Background(), options)
		},
		WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
			return watcher(context.Background(), options)
		},
	}

	reflector := cache.NewReflector(lw, unstructured.Unstructured{}, store, time.Hour)

	crdLister := dynamiclister.New(store, gvr)

	// Run reflector in the background so that we get new updates from the api-server
	go reflector.Run(stopChannel)

	// Wait for reflector to sync the cache for the first time
	// Note: Based on the docs WaitForNamedCacheSync seems to be used to check if an informer has synced
	// but the function is generic enough so we can use
	// it for reflectors as well
	synced := cache.WaitForNamedCacheSync(fmt.Sprintf("generic-%s-lister", gvr.Resource), stopChannel, func() bool {
		no, err := crdLister.List(labels.Everything())
		if err != nil {
			klog.Error("err", err)
		}
		return len(no) > 0
	})
	if !synced {
		klog.Error("couldn't sync cache")
	}

	return crdLister
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
func NewPodDisruptionBudgetLister(kubeClient client.Interface, stopchannel <-chan struct{}) PodDisruptionBudgetLister {
	listWatcher := cache.NewListWatchFromClient(kubeClient.PolicyV1().RESTClient(), "poddisruptionbudgets", apiv1.NamespaceAll, fields.Everything())
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

// NewGenericLister is a helper which returns a generic lister given the right gvr and namespace
// This is meant to be a more generic version of GetLister()
func NewGenericLister(dClient dynamic.Interface, listersMap map[string]dynamiclister.Lister, stopCh <-chan struct{}, gvr schema.GroupVersionResource, namespace string) dynamiclister.Lister {
	key := fmt.Sprintf("%s_%s_%s_%s", gvr.Group, gvr.Version, gvr.Resource, namespace)
	if listersMap[key] != nil {
		return listersMap[key]
	}

	var lister func(ctx context.Context, opts metav1.ListOptions) (*unstructured.UnstructuredList, error)
	var watcher func(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)

	if namespace == apiv1.NamespaceAll {
		lister = dClient.Resource(gvr).List
		watcher = dClient.Resource(gvr).Watch
	} else {
		// For lister limited to a particular namespace
		lister = dClient.Resource(gvr).Namespace(namespace).List
		watcher = dClient.Resource(gvr).Namespace(namespace).Watch
	}

	// NewNamespaceKeyedIndexerAndReflector can be
	// used for both namespace and cluster scoped resources
	store, reflector := cache.NewNamespaceKeyedIndexerAndReflector(&cache.ListWatch{
		ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
			return lister(context.Background(), options)
		},
		WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
			return watcher(context.Background(), options)
		},
	}, unstructured.Unstructured{}, time.Hour)
	l := dynamiclister.New(store, gvr)

	// Run reflector in the background so that we get new updates from the api-server
	go reflector.Run(stopCh)

	// Wait for reflector to sync the cache for the first time
	// Note: Based on the docs WaitForNamedCacheSync seems to be used to check if an informer has synced
	// but the function is generic enough so we can use
	// it for reflectors as well
	synced := cache.WaitForNamedCacheSync(fmt.Sprintf("generic-%s-lister", gvr.Resource), stopCh, func() bool {
		no, err := l.List(labels.Everything())
		if err != nil {
			klog.Error("err", err)
		}
		return len(no) > 0
	})
	if !synced {
		// don't return an error but don't add
		// this lister to listers map
		// so that another attempt is made
		// to create the lister and sync cache
		klog.Error("couldn't sync cache")
	} else {
		// make the lister available in the listers map through GetDynamicListerMap (maps are passed by reference)
		// for the next time something requests a lister for the same GVR and namespace
		listersMap[key] = l
	}

	return l
}
