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

package simulator

import (
	"fmt"
	"math"

	kube_api "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/resource"
	kube_client "k8s.io/kubernetes/pkg/client/unversioned"
	cmd "k8s.io/kubernetes/pkg/kubectl/cmd"
	"k8s.io/kubernetes/plugin/pkg/scheduler/schedulercache"

	"github.com/golang/glog"
)

// FindNodesToRemove finds nodes that can be removed.
func FindNodesToRemove(candidates []*kube_api.Node, allNodes []*kube_api.Node, pods []*kube_api.Pod,
	client *kube_client.Client, predicateChecker *PredicateChecker, maxCount int,
	fastCheck bool) ([]*kube_api.Node, error) {

	nodeNameToNodeInfo := schedulercache.CreateNodeNameToInfoMap(pods)
	for _, node := range allNodes {
		if nodeInfo, found := nodeNameToNodeInfo[node.Name]; found {
			nodeInfo.SetNode(node)
		}
	}
	result := make([]*kube_api.Node, 0)

	for _, node := range candidates {
		glog.V(2).Infof("Considering %s for removal", node.Name)

		var podsToRemove []*kube_api.Pod
		var err error

		if fastCheck {
			if nodeInfo, found := nodeNameToNodeInfo[node.Name]; found {
				podsToRemove, err = FastGetPodsToMove(nodeInfo, false, true, kube_api.Codecs.UniversalDecoder())
			}
		} else {
			drainResult, _, _, err := cmd.GetPodsForDeletionOnNodeDrain(client, node.Name,
				kube_api.Codecs.UniversalDecoder(), false, true)

			if err != nil {
				glog.V(2).Infof("Node %s cannot be removed: %v", node.Name, err)
				continue
			}
			podsToRemove = make([]*kube_api.Pod, 0, len(drainResult))
			for i := range drainResult {
				podsToRemove = append(podsToRemove, &drainResult[i])
			}
		}

		findProblems := findPlaceFor(node.Name, podsToRemove, allNodes, nodeNameToNodeInfo, predicateChecker)
		if findProblems == nil {
			result = append(result, node)
			if len(result) >= maxCount {
				break
			}
		}
		glog.V(2).Infof("Node %s is not suitable for removal %v", node.Name, err)
	}
	return result, nil
}

// CalculateReservation calculates reservation of a node.
func CalculateReservation(node *kube_api.Node, nodeInfo *schedulercache.NodeInfo) (float64, error) {
	cpu, err := calculateReservationOfResource(node, nodeInfo, kube_api.ResourceCPU)
	if err != nil {
		return 0, err
	}
	mem, err := calculateReservationOfResource(node, nodeInfo, kube_api.ResourceMemory)
	if err != nil {
		return 0, err
	}
	return math.Max(cpu, mem), nil
}

func calculateReservationOfResource(node *kube_api.Node, nodeInfo *schedulercache.NodeInfo, resourceName kube_api.ResourceName) (float64, error) {
	nodeCapacity, found := node.Status.Capacity[resourceName]
	if !found {
		return 0, fmt.Errorf("Failed to get %v from %s", resourceName, node.Name)
	}
	if nodeCapacity.MilliValue() == 0 {
		return 0, fmt.Errorf("%v is 0 at %s", resourceName, node.Name)
	}
	podsRequest := resource.MustParse("0")
	for _, pod := range nodeInfo.Pods() {
		for _, container := range pod.Spec.Containers {
			if resourceValue, found := container.Resources.Requests[resourceName]; found {
				podsRequest.Add(resourceValue)
			}
		}
	}
	return float64(podsRequest.MilliValue()) / float64(nodeCapacity.MilliValue()), nil
}

// TODO: We don't need to pass list of nodes here as they are already available in nodeInfos.
func findPlaceFor(bannedNode string, pods []*kube_api.Pod, nodes []*kube_api.Node, nodeInfos map[string]*schedulercache.NodeInfo,
	predicateChecker *PredicateChecker) error {

	newNodeInfos := make(map[string]*schedulercache.NodeInfo)

	for _, podptr := range pods {
		newpod := *podptr
		newpod.Spec.NodeName = ""
		pod := &newpod

		foundPlace := false
		glog.V(4).Infof("Looking for place for %s/%s", pod.Namespace, pod.Name)
		podKey := fmt.Sprintf("%s/%s", pod.Namespace, pod.Name)

		// TODO: Sort nodes by reservation
	nodeloop:
		for _, node := range nodes {
			if node.Name == bannedNode {
				continue
			}

			node.Status.Allocatable = node.Status.Capacity
			nodeInfo, found := newNodeInfos[node.Name]
			if !found {
				nodeInfo, found = nodeInfos[node.Name]
			}

			if found {
				err := predicateChecker.CheckPredicates(pod, nodeInfo)
				glog.V(4).Infof("Evaluation %s for %s -> %v", node.Name, podKey, err)
				if err == nil {
					foundPlace = true
					// TODO(mwielgus): Optimize it.
					podsOnNode := nodeInfo.Pods()
					podsOnNode = append(podsOnNode, pod)
					newNodeInfo := schedulercache.NewNodeInfo(podsOnNode...)
					newNodeInfo.SetNode(node)
					newNodeInfos[node.Name] = newNodeInfo
					break nodeloop
				}
			}
		}
		if !foundPlace {
			return fmt.Errorf("failed to find place for %s", podKey)
		}
	}
	return nil
}
