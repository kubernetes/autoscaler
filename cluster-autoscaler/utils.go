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

package main

import (
	"fmt"
	"math/rand"
	"reflect"
	"time"

	"k8s.io/contrib/cluster-autoscaler/cloudprovider"
	"k8s.io/contrib/cluster-autoscaler/simulator"

	kube_api "k8s.io/kubernetes/pkg/api"
	kube_api_unversioned "k8s.io/kubernetes/pkg/api/unversioned"
	kube_client "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/plugin/pkg/scheduler/schedulercache"

	"github.com/golang/glog"
)

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

// CheckGroupsAndNodes checks if all node groups have all required nodes.
func CheckGroupsAndNodes(nodes []*kube_api.Node, cloudProvider cloudprovider.CloudProvider) error {
	groupCount := make(map[string]int)
	for _, node := range nodes {

		group, err := cloudProvider.NodeGroupForNode(node)
		if err != nil {
			return err
		}
		if group == nil || reflect.ValueOf(group).IsNil() {
			continue
		}
		id := group.Id()
		count, _ := groupCount[id]
		groupCount[id] = count + 1
	}
	for _, nodeGroup := range cloudProvider.NodeGroups() {
		size, err := nodeGroup.TargetSize()
		if err != nil {
			return err
		}
		count := groupCount[nodeGroup.Id()]
		if size != count {
			return fmt.Errorf("wrong number of nodes for node group: %s expected: %d actual: %d", nodeGroup.Id(), size, count)
		}
	}
	return nil
}

// GetNodeInfosForGroups finds NodeInfos for all node groups used to manage the given nodes. It also returns a node group to sample node mapping.
// TODO(mwielgus): This returns map keyed by url, while most code (including scheduler) uses node.Name for a key.
func GetNodeInfosForGroups(nodes []*kube_api.Node, cloudProvider cloudprovider.CloudProvider, kubeClient *kube_client.Client) (map[string]*schedulercache.NodeInfo, error) {
	result := make(map[string]*schedulercache.NodeInfo)
	for _, node := range nodes {

		nodeGroup, err := cloudProvider.NodeGroupForNode(node)
		if err != nil {
			return map[string]*schedulercache.NodeInfo{}, err
		}
		if nodeGroup == nil || reflect.ValueOf(nodeGroup).IsNil() {
			continue
		}
		id := nodeGroup.Id()
		if _, found := result[id]; !found {
			nodeInfo, err := simulator.BuildNodeInfoForNode(node, kubeClient)
			if err != nil {
				return map[string]*schedulercache.NodeInfo{}, err
			}
			result[id] = nodeInfo
		}
	}
	return result, nil
}

// BestExpansionOption picks the best cluster expansion option.
func BestExpansionOption(expansionOptions []ExpansionOption) *ExpansionOption {
	if len(expansionOptions) > 0 {
		pos := rand.Int31n(int32(len(expansionOptions)))
		return &expansionOptions[pos]
	}
	return nil
}
