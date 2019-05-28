/*
Copyright 2018 The Kubernetes Authors.

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
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	apiv1 "k8s.io/api/core/v1"
	v1appslister "k8s.io/client-go/listers/apps/v1"
	v1batchlister "k8s.io/client-go/listers/batch/v1"
	v1lister "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
)

// TestPodLister is used in tests involving listers
type TestPodLister struct {
	pods []*apiv1.Pod
}

// List returns all pods in test lister.
func (lister TestPodLister) List() ([]*apiv1.Pod, error) {
	return lister.pods, nil
}

// NewTestPodLister returns a lister that returns provided pods
func NewTestPodLister(pods []*apiv1.Pod) PodLister {
	return TestPodLister{pods: pods}
}

// TestNodeLister is used in tests involving listers
type TestNodeLister struct {
	nodes []*apiv1.Node
}

// List returns all nodes in test lister.
func (l *TestNodeLister) List() ([]*apiv1.Node, error) {
	return l.nodes, nil
}

// Get returns node from test lister.
func (l *TestNodeLister) Get(name string) (*apiv1.Node, error) {
	for _, node := range l.nodes {
		if node.Name == name {
			return node, nil
		}
	}
	return nil, fmt.Errorf("Node %s not found", name)
}

// SetNodes sets nodes in test lister.
func (l *TestNodeLister) SetNodes(nodes []*apiv1.Node) {
	l.nodes = nodes
}

// NewTestNodeLister returns a lister that returns provided nodes
func NewTestNodeLister(nodes []*apiv1.Node) *TestNodeLister {
	return &TestNodeLister{nodes: nodes}
}

// NewTestDaemonSetLister returns a lister that returns provided DaemonSets
func NewTestDaemonSetLister(dss []*appsv1.DaemonSet) (v1appslister.DaemonSetLister, error) {
	store := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	for _, ds := range dss {
		err := store.Add(ds)
		if err != nil {
			return nil, fmt.Errorf("Error adding object to cache: %v", err)
		}
	}
	return v1appslister.NewDaemonSetLister(store), nil
}

// NewTestReplicationControllerLister returns a lister that returns provided ReplicationControllers
func NewTestReplicationControllerLister(rcs []*apiv1.ReplicationController) (v1lister.ReplicationControllerLister, error) {
	store := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	for _, rc := range rcs {
		err := store.Add(rc)
		if err != nil {
			return nil, fmt.Errorf("Error adding object to cache: %v", err)
		}
	}
	return v1lister.NewReplicationControllerLister(store), nil
}

// NewTestJobLister returns a lister that returns provided Jobs
func NewTestJobLister(jobs []*batchv1.Job) (v1batchlister.JobLister, error) {
	store := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	for _, job := range jobs {
		err := store.Add(job)
		if err != nil {
			return nil, fmt.Errorf("Error adding object to cache: %v", err)
		}
	}
	return v1batchlister.NewJobLister(store), nil
}

// NewTestReplicaSetLister returns a lister that returns provided ReplicaSets
func NewTestReplicaSetLister(rss []*appsv1.ReplicaSet) (v1appslister.ReplicaSetLister, error) {
	store := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	for _, rs := range rss {
		err := store.Add(rs)
		if err != nil {
			return nil, fmt.Errorf("Error adding object to cache: %v", err)
		}
	}
	return v1appslister.NewReplicaSetLister(store), nil
}

// NewTestStatefulSetLister returns a lister that returns provided StatefulSets
func NewTestStatefulSetLister(sss []*appsv1.StatefulSet) (v1appslister.StatefulSetLister, error) {
	store := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	for _, ss := range sss {
		err := store.Add(ss)
		if err != nil {
			return nil, fmt.Errorf("Error adding object to cache: %v", err)
		}
	}
	return v1appslister.NewStatefulSetLister(store), nil
}

// NewTestConfigMapLister returns a lister that returns provided ConfigMaps
func NewTestConfigMapLister(cms []*apiv1.ConfigMap) (v1lister.ConfigMapLister, error) {
	store := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	for _, cm := range cms {
		err := store.Add(cm)
		if err != nil {
			return nil, fmt.Errorf("Error adding object to cache: %v", err)
		}
	}
	return v1lister.NewConfigMapLister(store), nil
}
