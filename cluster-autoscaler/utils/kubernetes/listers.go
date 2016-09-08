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

	kube_api "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/cache"
	kube_client "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/labels"
)

// UnschedulablePodLister lists unscheduled pods
type UnschedulablePodLister struct {
	podLister *cache.StoreToPodLister
}

// List returns all unscheduled pods.
func (unschedulablePodLister *UnschedulablePodLister) List() ([]*kube_api.Pod, error) {
	var unschedulablePods []*kube_api.Pod
	allPods, err := unschedulablePodLister.podLister.List(labels.Everything())
	if err != nil {
		return unschedulablePods, err
	}
	for _, pod := range allPods {
		_, condition := kube_api.GetPodCondition(&pod.Status, kube_api.PodScheduled)
		if condition != nil && condition.Status == kube_api.ConditionFalse && condition.Reason == "Unschedulable" {
			unschedulablePods = append(unschedulablePods, pod)
		}
	}
	return unschedulablePods, nil
}

// NewUnschedulablePodLister returns a lister providing pods that failed to be scheduled.
func NewUnschedulablePodLister(kubeClient *kube_client.Client, namespace string) *UnschedulablePodLister {
	// watch unscheduled pods
	selector := fields.ParseSelectorOrDie("spec.nodeName==" + "" + ",status.phase!=" +
		string(kube_api.PodSucceeded) + ",status.phase!=" + string(kube_api.PodFailed))
	podListWatch := cache.NewListWatchFromClient(kubeClient, "pods", namespace, selector)
	store := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	podLister := &cache.StoreToPodLister{store}
	podReflector := cache.NewReflector(podListWatch, &kube_api.Pod{}, store, time.Hour)
	podReflector.Run()

	return &UnschedulablePodLister{
		podLister: podLister,
	}
}

// ScheduledPodLister lists scheduled pods.
type ScheduledPodLister struct {
	podLister *cache.StoreToPodLister
}

// List returns all scheduled pods.
func (lister *ScheduledPodLister) List() ([]*kube_api.Pod, error) {
	return lister.podLister.List(labels.Everything())
}

// NewScheduledPodLister builds ScheduledPodLister
func NewScheduledPodLister(kubeClient *kube_client.Client) *ScheduledPodLister {
	// watch unscheduled pods
	selector := fields.ParseSelectorOrDie("spec.nodeName!=" + "" + ",status.phase!=" +
		string(kube_api.PodSucceeded) + ",status.phase!=" + string(kube_api.PodFailed))
	podListWatch := cache.NewListWatchFromClient(kubeClient, "pods", kube_api.NamespaceAll, selector)
	store := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	podLister := &cache.StoreToPodLister{store}
	podReflector := cache.NewReflector(podListWatch, &kube_api.Pod{}, store, time.Hour)
	podReflector.Run()

	return &ScheduledPodLister{
		podLister: podLister,
	}
}

// ReadyNodeLister lists ready nodes.
type ReadyNodeLister struct {
	nodeLister *cache.StoreToNodeLister
}

// List returns ready nodes.
func (readyNodeLister *ReadyNodeLister) List() ([]*kube_api.Node, error) {
	nodes, err := readyNodeLister.nodeLister.List()
	if err != nil {
		return []*kube_api.Node{}, err
	}
	readyNodes := make([]*kube_api.Node, 0, len(nodes.Items))
	for i, node := range nodes.Items {
		if node.Spec.Unschedulable {
			continue
		}
		for _, condition := range node.Status.Conditions {
			if condition.Type == kube_api.NodeReady && condition.Status == kube_api.ConditionTrue {
				readyNodes = append(readyNodes, &nodes.Items[i])
				break
			}
		}
	}
	return readyNodes, nil
}

// NewNodeLister builds a node lister.
func NewNodeLister(kubeClient *kube_client.Client) *ReadyNodeLister {
	listWatcher := cache.NewListWatchFromClient(kubeClient, "nodes", kube_api.NamespaceAll, fields.Everything())
	nodeLister := &cache.StoreToNodeLister{Store: cache.NewStore(cache.MetaNamespaceKeyFunc)}
	reflector := cache.NewReflector(listWatcher, &kube_api.Node{}, nodeLister.Store, time.Hour)
	reflector.Run()
	return &ReadyNodeLister{
		nodeLister: nodeLister,
	}
}
