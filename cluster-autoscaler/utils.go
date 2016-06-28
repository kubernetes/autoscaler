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
	kube_api_unversioned "k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/client/cache"
	kube_client "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/plugin/pkg/scheduler/schedulercache"

	"github.com/golang/glog"
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
func NewUnschedulablePodLister(kubeClient *kube_client.Client) *UnschedulablePodLister {
	// watch unscheduled pods
	selector := fields.ParseSelectorOrDie("spec.nodeName==" + "" + ",status.phase!=" +
		string(kube_api.PodSucceeded) + ",status.phase!=" + string(kube_api.PodFailed))
	podListWatch := cache.NewListWatchFromClient(kubeClient, "pods", kube_api.NamespaceAll, selector)
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

// GetAllNodesAvailableTime returns time when the newest node became available for scheduler.
// TODO: This function should use LastTransitionTime from NodeReady condition.
func GetAllNodesAvailableTime(nodes []*kube_api.Node) time.Time {
	var result time.Time
	for _, node := range nodes {
		if node.CreationTimestamp.After(result) {
			result = node.CreationTimestamp.Time
		}
	}
	return result.Add(1 * time.Minute)
}

// SlicePodsByPodScheduledTime slices given pod array into those where PodScheduled condition
// have been updated after the thresold and others.
// Each pod must be in condition "Scheduled: False; Reason: Unschedulable"
func SlicePodsByPodScheduledTime(pods []*kube_api.Pod, threshold time.Time) (oldPods []*kube_api.Pod, newPods []*kube_api.Pod) {
	for _, pod := range pods {
		_, condition := kube_api.GetPodCondition(&pod.Status, kube_api.PodScheduled)
		if condition != nil {
			if condition.LastTransitionTime.After(threshold) {
				newPods = append(newPods, pod)
			} else {
				oldPods = append(oldPods, pod)
			}
		}
	}
	return
}

// ResetPodScheduledCondition resets pod condition PodScheduled to "unknown" for all the pods with LastTransitionTime
// not after the threshold time.
func ResetPodScheduledCondition(kubeClient *kube_client.Client, pods []*kube_api.Pod) {
	for _, pod := range pods {
		if err := resetPodScheduledConditionForPod(kubeClient, pod); err != nil {
			glog.Errorf("Error during reseting pod condition for %s/%s: %v", pod.Namespace, pod.Name, err)
		}
	}
}

func resetPodScheduledConditionForPod(kubeClient *kube_client.Client, pod *kube_api.Pod) error {
	_, condition := kube_api.GetPodCondition(&pod.Status, kube_api.PodScheduled)
	if condition != nil {
		glog.V(4).Infof("Reseting pod condition for %s/%s, last transition: %s",
			pod.Namespace, pod.Name, condition.LastTransitionTime.Time.String())
		condition.Status = kube_api.ConditionUnknown
		condition.LastTransitionTime = kube_api_unversioned.Now()
		_, err := kubeClient.Pods(pod.Namespace).UpdateStatus(pod)
		return err
	}
	return fmt.Errorf("Expected condition PodScheduled")
}

// FilterOutSchedulable checks whether pods from <unschedulableCandidates> marked as unschedulable
// by Scheduler actually can't be scheduled on any node and filter out the ones that can.
func FilterOutSchedulable(unschedulableCandidates []*kube_api.Pod, nodes []*kube_api.Node, allPods []*kube_api.Pod, predicateChecker *simulator.PredicateChecker) []*kube_api.Pod {
	unschedulablePods := []*kube_api.Pod{}
	nodeNameToNodeInfo := createNodeNameToInfoMap(allPods, nodes)

	for _, pod := range unschedulableCandidates {
		if nodeName, err := predicateChecker.FitsAny(pod, nodeNameToNodeInfo); err == nil {
			glog.Warningf("Pod %s marked as unschedulable can be scheduled on %s. Ignoring in scale up.", pod.Name, nodeName)
		} else {
			unschedulablePods = append(unschedulablePods, pod)
		}
	}

	return unschedulablePods
}

// TODO: move this function to scheduler utils.
func createNodeNameToInfoMap(pods []*kube_api.Pod, nodes []*kube_api.Node) map[string]*schedulercache.NodeInfo {
	nodeNameToNodeInfo := schedulercache.CreateNodeNameToInfoMap(pods)
	for _, node := range nodes {
		if nodeInfo, found := nodeNameToNodeInfo[node.Name]; found {
			nodeInfo.SetNode(node)
		}
	}

	// Some pods may be out of sync with node lists. Removing incomplete node infos.
	keysToRemove := make([]string, 0)
	for key, nodeInfo := range nodeNameToNodeInfo {
		if nodeInfo.Node() == nil {
			keysToRemove = append(keysToRemove, key)
		}
	}
	for _, key := range keysToRemove {
		delete(nodeNameToNodeInfo, key)
	}

	return nodeNameToNodeInfo
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
		if migConfig == nil {
			continue
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
// TODO(mwielgus): This returns map keyed by url, while most code (including scheduler) uses node.Name for a key.
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
