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

package estimator

import (
	"fmt"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/predicatechecker"
	"k8s.io/autoscaler/cluster-autoscaler/utils/scheduler"
	klog "k8s.io/klog/v2"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
)

// BinpackingNodeEstimator estimates the number of needed nodes to handle the given amount of pods.
type BinpackingNodeEstimator struct {
	predicateChecker predicatechecker.PredicateChecker
	clusterSnapshot  clustersnapshot.ClusterSnapshot
	limiter          EstimationLimiter
	podOrderer       EstimationPodOrderer
}

// NewBinpackingNodeEstimator builds a new BinpackingNodeEstimator.
func NewBinpackingNodeEstimator(
	predicateChecker predicatechecker.PredicateChecker,
	clusterSnapshot clustersnapshot.ClusterSnapshot,
	limiter EstimationLimiter,
	podOrderer EstimationPodOrderer) *BinpackingNodeEstimator {
	return &BinpackingNodeEstimator{
		predicateChecker: predicateChecker,
		clusterSnapshot:  clusterSnapshot,
		limiter:          limiter,
		podOrderer:       podOrderer,
	}
}

// Estimate implements First-Fit bin-packing approximation algorithm
// The ordering of the pods depend on the EstimatePodOrderer, the default
// order is DecreasingPodOrderer
// First-Fit Decreasing bin-packing approximation algorithm.
// See https://en.wikipedia.org/wiki/Bin_packing_problem for more details.
// While it is a multi-dimensional bin packing (cpu, mem, ports) in most cases the main dimension
// will be cpu thus the estimated overprovisioning of 11/9 * optimal + 6/9 should be
// still be maintained.
// It is assumed that all pods from the given list can fit to nodeTemplate.
// Returns the number of nodes needed to accommodate all pods from the list.
func (e *BinpackingNodeEstimator) Estimate(
	pods []*apiv1.Pod,
	nodeTemplate *schedulerframework.NodeInfo,
	nodeGroup cloudprovider.NodeGroup) (int, []*apiv1.Pod) {

	e.limiter.StartEstimation(pods, nodeGroup)
	defer e.limiter.EndEstimation()

	pods = e.podOrderer.Order(pods, nodeTemplate, nodeGroup)

	newNodeNames := make(map[string]bool)
	newNodesWithPods := make(map[string]bool)

	e.clusterSnapshot.Fork()
	defer func() {
		e.clusterSnapshot.Revert()
	}()

	newNodeNameIndex := 0
	scheduledPods := []*apiv1.Pod{}
	lastNodeName := ""

	for _, pod := range pods {
		found := false

		nodeName, err := e.predicateChecker.FitsAnyNodeMatching(e.clusterSnapshot, pod, func(nodeInfo *schedulerframework.NodeInfo) bool {
			return newNodeNames[nodeInfo.Node().Name]
		})
		if err == nil {
			found = true
			if err := e.clusterSnapshot.AddPod(pod, nodeName); err != nil {
				klog.Errorf("Error adding pod %v.%v to node %v in ClusterSnapshot; %v", pod.Namespace, pod.Name, nodeName, err)
				return 0, nil
			}
			scheduledPods = append(scheduledPods, pod)
			newNodesWithPods[nodeName] = true
		}

		if !found {
			// Stop binpacking if we reach the limit of nodes we can add.
			// We return the result of the binpacking that we already performed.
			if !e.limiter.PermissionToAddNode() {
				break
			}

			// If the last node we've added is empty and the pod couldn't schedule on it, it wouldn't be able to schedule
			// on a new node either. There is no point adding more nodes to snapshot in such case, especially because of
			// performance cost each extra node adds to future FitsAnyNodeMatching calls.
			if lastNodeName != "" && !newNodesWithPods[lastNodeName] {
				continue
			}

			// Add new node
			newNodeName, err := e.addNewNodeToSnapshot(nodeTemplate, newNodeNameIndex)
			if err != nil {
				klog.Errorf("Error while adding new node for template to ClusterSnapshot; %v", err)
				return 0, nil
			}
			newNodeNameIndex++
			newNodeNames[newNodeName] = true
			lastNodeName = newNodeName

			// And try to schedule pod to it.
			// Note that this may still fail (ex. if topology spreading with zonal topologyKey is used);
			// in this case we can't help the pending pod. We keep the node in clusterSnapshot to avoid
			// adding and removing node to snapshot for each such pod.
			if err := e.predicateChecker.CheckPredicates(e.clusterSnapshot, pod, newNodeName); err != nil {
				continue
			}
			if err := e.clusterSnapshot.AddPod(pod, newNodeName); err != nil {
				klog.Errorf("Error adding pod %v.%v to node %v in ClusterSnapshot; %v", pod.Namespace, pod.Name, newNodeName, err)
				return 0, nil
			}
			newNodesWithPods[newNodeName] = true
			scheduledPods = append(scheduledPods, pod)
		}
	}
	return len(newNodesWithPods), scheduledPods
}

func (e *BinpackingNodeEstimator) addNewNodeToSnapshot(
	template *schedulerframework.NodeInfo,
	nameIndex int) (string, error) {

	newNodeInfo := scheduler.DeepCopyTemplateNode(template, fmt.Sprintf("e-%d", nameIndex))
	var pods []*apiv1.Pod
	for _, podInfo := range newNodeInfo.Pods {
		pods = append(pods, podInfo.Pod)
	}
	if err := e.clusterSnapshot.AddNodeWithPods(newNodeInfo.Node(), pods); err != nil {
		return "", err
	}
	return newNodeInfo.Node().Name, nil
}
