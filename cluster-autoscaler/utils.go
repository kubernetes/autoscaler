/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

package main

import (
	"fmt"
	"time"

	"k8s.io/contrib/cluster-autoscaler/config"
	"k8s.io/contrib/cluster-autoscaler/simulator"
	"k8s.io/contrib/cluster-autoscaler/utils/gce"

	kube_api "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/cache"
	kube_client "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/plugin/pkg/scheduler/schedulercache"
)

// UnscheduledPodLister list unscheduled pods
type UnscheduledPodLister struct {
	podLister *cache.StoreToPodLister
}

// List returns all unscheduled pods.
func (unscheduledPodLister *UnscheduledPodLister) List() ([]*kube_api.Pod, error) {
	//TODO: Extra filter based on pod condition.
	return unscheduledPodLister.podLister.List(labels.Everything())
}

// NewUnscheduledPodLister returns a lister providing pods that failed to be scheduled.
func NewUnscheduledPodLister(kubeClient *kube_client.Client) *UnscheduledPodLister {
	// watch unscheduled pods
	selector := fields.ParseSelectorOrDie("spec.nodeName==" + "" + ",status.phase!=" +
		string(kube_api.PodSucceeded) + ",status.phase!=" + string(kube_api.PodFailed))
	podListWatch := cache.NewListWatchFromClient(kubeClient, "pods", kube_api.NamespaceAll, selector)
	store := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	podLister := &cache.StoreToPodLister{store}
	podReflector := cache.NewReflector(podListWatch, &kube_api.Pod{}, store, time.Hour)
	podReflector.Run()

	return &UnscheduledPodLister{
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

// GetNewestNode returns the newest node from the given list.
func GetNewestNode(nodes []*kube_api.Node) *kube_api.Node {
	var result *kube_api.Node
	for _, node := range nodes {
		if result == nil || node.CreationTimestamp.After(result.CreationTimestamp.Time) {
			result = node
		}
	}
	return result
}

// GetOldestFailedSchedulingTrail returns the oldest time when a pod from the given list failed to
// be scheduled.
func GetOldestFailedSchedulingTrail(pods []*kube_api.Pod) *time.Time {
	// Dummy implementation.
	//TODO: Implement once pod condition is there.
	now := time.Now()
	return &now
}

// CheckMigsAndNodes checks if all migs have all required nodes.
func CheckMigsAndNodes(nodes []*kube_api.Node, gceManager *gce.GceManager) error {
	migCount := make(map[string]int)
	migs := make(map[string]*config.MigConfig)
	for _, node := range nodes {
		instanceConfig, err := config.InstanceConfigFromProviderId(node.Spec.ProviderID)
		if err != nil {
			return err
		}

		migConfig, err := gceManager.GetMigForInstance(instanceConfig)
		if err != nil {
			return err
		}
		url := migConfig.Url()
		count, _ := migCount[url]
		migCount[url] = count + 1
		migs[url] = migConfig
	}
	for url, mig := range migs {
		size, err := gceManager.GetMigSize(mig)
		if err != nil {
			return err
		}
		count := migCount[url]
		if size != int64(count) {
			return fmt.Errorf("wrong number of nodes for mig: %s expected: %d actual: %d", url, size, count)
		}
	}
	return nil
}

// GetNodeInfosForMigs finds NodeInfos for all migs used to manage the given nodes. It also returns a mig to sample node mapping.
// TODO(mwielgus): This return map keyed by url, while most code (including scheduler) uses node.Name for a key.
func GetNodeInfosForMigs(nodes []*kube_api.Node, gceManager *gce.GceManager, kubeClient *kube_client.Client) (map[string]*schedulercache.NodeInfo, error) {
	result := make(map[string]*schedulercache.NodeInfo)
	for _, node := range nodes {
		instanceConfig, err := config.InstanceConfigFromProviderId(node.Spec.ProviderID)
		if err != nil {
			return map[string]*schedulercache.NodeInfo{}, err
		}

		migConfig, err := gceManager.GetMigForInstance(instanceConfig)
		if err != nil {
			return map[string]*schedulercache.NodeInfo{}, err
		}
		url := migConfig.Url()

		nodeInfo, err := simulator.BuildNodeInfoForNode(node, kubeClient)
		if err != nil {
			return map[string]*schedulercache.NodeInfo{}, err
		}

		result[url] = nodeInfo
	}
	return result, nil
}

// BestExpansionOption picks the best cluster expansion option.
func BestExpansionOption(expansionOptions []ExpansionOption) *ExpansionOption {
	if len(expansionOptions) > 0 {
		return &expansionOptions[0]
	}
	return nil
}
