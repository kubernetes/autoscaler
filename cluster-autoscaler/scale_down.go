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
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"k8s.io/contrib/cluster-autoscaler/cloudprovider"
	"k8s.io/contrib/cluster-autoscaler/simulator"
	"k8s.io/kubernetes/pkg/api/errors"
	apiv1 "k8s.io/kubernetes/pkg/api/v1"
	kube_client "k8s.io/kubernetes/pkg/client/clientset_generated/release_1_5"
	kube_record "k8s.io/kubernetes/pkg/client/record"
	"k8s.io/kubernetes/plugin/pkg/scheduler/schedulercache"

	"github.com/golang/glog"
)

// ScaleDownResult represents the state of scale down.
type ScaleDownResult int

const (
	// ScaleDownError - scale down finished with error.
	ScaleDownError ScaleDownResult = iota
	// ScaleDownNoUnneeded - no unneeded nodes and no errors.
	ScaleDownNoUnneeded ScaleDownResult = iota
	// ScaleDownNoNodeDeleted - unneeded nodes present but not available for deletion.
	ScaleDownNoNodeDeleted ScaleDownResult = iota
	// ScaleDownNodeDeleted - a node was deleted.
	ScaleDownNodeDeleted ScaleDownResult = iota
)

const (
	// ToBeDeletedTaint is a taint used to make the node unschedulable.
	ToBeDeletedTaint = "ToBeDeletedByClusterAutoscaler"

	// MaxGracefulTerminationTime is max gracefull termination time used by CA.
	MaxGracefulTerminationTime = time.Minute
)

// FindUnneededNodes calculates which nodes are not needed, i.e. all pods can be scheduled somewhere else,
// and updates unneededNodes map accordingly. It also returns information where pods can be rescheduld and
// node utilization level.
func FindUnneededNodes(nodes []*apiv1.Node,
	unneededNodes map[string]time.Time,
	utilizationThreshold float64,
	pods []*apiv1.Pod,
	predicateChecker *simulator.PredicateChecker,
	oldHints map[string]string,
	tracker *simulator.UsageTracker,
	timestamp time.Time) (unnededTimeMap map[string]time.Time, podReschedulingHints map[string]string, utilizationMap map[string]float64) {

	currentlyUnneededNodes := make([]*apiv1.Node, 0)
	nodeNameToNodeInfo := schedulercache.CreateNodeNameToInfoMap(pods, nodes)
	utilizationMap = make(map[string]float64)

	// Phase1 - look at the nodes utilization.
	for _, node := range nodes {
		nodeInfo, found := nodeNameToNodeInfo[node.Name]
		if !found {
			glog.Errorf("Node info for %s not found", node.Name)
			continue
		}
		utilization, err := simulator.CalculateUtilization(node, nodeInfo)

		if err != nil {
			glog.Warningf("Failed to calculate utilization for %s: %v", node.Name, err)
		}
		glog.V(4).Infof("Node %s - utilization %f", node.Name, utilization)
		utilizationMap[node.Name] = utilization

		if utilization >= utilizationThreshold {
			glog.V(4).Infof("Node %s is not suitable for removal - utilization too big (%f)", node.Name, utilization)
			continue
		}
		currentlyUnneededNodes = append(currentlyUnneededNodes, node)
	}

	// Phase2 - check which nodes can be probably removed using fast drain.
	nodesToRemove, newHints, err := simulator.FindNodesToRemove(currentlyUnneededNodes, nodes, pods,
		nil, predicateChecker,
		len(currentlyUnneededNodes), true, oldHints, tracker, timestamp)
	if err != nil {
		glog.Errorf("Error while simulating node drains: %v", err)
		return map[string]time.Time{}, oldHints, map[string]float64{}
	}

	// Update the timestamp map.
	now := time.Now()
	result := make(map[string]time.Time)
	for _, node := range nodesToRemove {
		name := node.Node.Name
		if val, found := unneededNodes[name]; !found {
			result[name] = now
		} else {
			result[name] = val
		}
	}
	return result, newHints, utilizationMap
}

// ScaleDown tries to scale down the cluster. It returns ScaleDownResult indicating if any node was
// removed and error if such occured.
func ScaleDown(
	nodes []*apiv1.Node,
	lastUtilizationMap map[string]float64,
	unneededNodes map[string]time.Time,
	unneededTime time.Duration,
	pods []*apiv1.Pod,
	cloudProvider cloudprovider.CloudProvider,
	client kube_client.Interface,
	predicateChecker *simulator.PredicateChecker,
	oldHints map[string]string,
	usageTracker *simulator.UsageTracker,
	recorder kube_record.EventRecorder,
	maxEmptyBulkDelete int) (ScaleDownResult, error) {

	now := time.Now()
	candidates := make([]*apiv1.Node, 0)
	for _, node := range nodes {
		if val, found := unneededNodes[node.Name]; found {

			glog.V(2).Infof("%s was unneeded for %s", node.Name, now.Sub(val).String())

			// Check how long the node was underutilized.
			if !val.Add(unneededTime).Before(now) {
				continue
			}

			nodeGroup, err := cloudProvider.NodeGroupForNode(node)
			if err != nil {
				glog.Errorf("Error while checking node group for %s: %v", node.Name, err)
				continue
			}
			if nodeGroup == nil || reflect.ValueOf(nodeGroup).IsNil() {
				glog.V(4).Infof("Skipping %s - no node group config", node.Name)
				continue
			}

			size, err := nodeGroup.TargetSize()
			if err != nil {
				glog.Errorf("Error while checking node group size %s: %v", nodeGroup.Id(), err)
				continue
			}

			if size <= nodeGroup.MinSize() {
				glog.V(1).Infof("Skipping %s - node group min size reached", node.Name)
				continue
			}

			candidates = append(candidates, node)
		}
	}
	if len(candidates) == 0 {
		glog.Infof("No candidates for scale down")
		return ScaleDownNoUnneeded, nil
	}

	// Trying to delete empty nodes in bulk. If there are no empty nodes then CA will
	// try to delete not-so-empty nodes, possibly killing some pods and allowing them
	// to recreate on other nodes.
	emptyNodes := getEmptyNodes(candidates, pods, maxEmptyBulkDelete, cloudProvider)
	if len(emptyNodes) > 0 {
		confirmation := make(chan error, len(emptyNodes))
		for _, node := range emptyNodes {
			glog.V(0).Infof("Scale-down: removing empty node %s", node.Name)
			simulator.RemoveNodeFromTracker(usageTracker, node.Name, unneededNodes)
			go func(nodeToDelete *apiv1.Node) {
				confirmation <- deleteNodeFromCloudProvider(nodeToDelete, cloudProvider, recorder)
			}(node)
		}
		var finalError error
		for range emptyNodes {
			if err := <-confirmation; err != nil {
				glog.Errorf("Problem with empty node deletion: %v", err)
				finalError = err
			}
		}
		if finalError == nil {
			return ScaleDownNodeDeleted, nil
		}
		return ScaleDownError, fmt.Errorf("failed to delete at least one empty node: %v", finalError)
	}

	// We look for only 1 node so new hints may be incomplete.
	nodesToRemove, _, err := simulator.FindNodesToRemove(candidates, nodes, pods, client, predicateChecker, 1, false,
		oldHints, usageTracker, time.Now())

	if err != nil {
		return ScaleDownError, fmt.Errorf("Find node to remove failed: %v", err)
	}
	if len(nodesToRemove) == 0 {
		glog.V(1).Infof("No node to remove")
		return ScaleDownNoNodeDeleted, nil
	}
	toRemove := nodesToRemove[0]
	utilization := lastUtilizationMap[toRemove.Node.Name]
	podNames := make([]string, 0, len(toRemove.PodsToReschedule))
	for _, pod := range toRemove.PodsToReschedule {
		podNames = append(podNames, pod.Namespace+"/"+pod.Name)
	}
	glog.V(0).Infof("Scale-down: removing node %s, utilization: %v, pods to reschedule: ", toRemove.Node.Name, utilization,
		strings.Join(podNames, ","))

	// Nothing super-bad should happen if the node is removed from tracker prematurely.
	simulator.RemoveNodeFromTracker(usageTracker, toRemove.Node.Name, unneededNodes)

	err = deleteNode(toRemove.Node, toRemove.PodsToReschedule, client, cloudProvider, recorder)
	if err != nil {
		return ScaleDownError, fmt.Errorf("Failed to delete %s: %v", toRemove.Node.Name, err)
	}

	return ScaleDownNodeDeleted, nil
}

// This functions finds empty nodes among passed candidates and returns a list of empty nodes
// that can be deleted at the same time.
func getEmptyNodes(candidates []*apiv1.Node, pods []*apiv1.Pod, maxEmptyBulkDelete int, cloudProvider cloudprovider.CloudProvider) []*apiv1.Node {
	emptyNodes := simulator.FindEmptyNodesToRemove(candidates, pods)
	availabilityMap := make(map[string]int)
	result := make([]*apiv1.Node, 0)
	for _, node := range emptyNodes {
		nodeGroup, err := cloudProvider.NodeGroupForNode(node)
		if err != nil {
			glog.Errorf("Failed to get group for %s", node.Name)
			continue
		}
		if nodeGroup == nil || reflect.ValueOf(nodeGroup).IsNil() {
			continue
		}
		var available int
		var found bool
		if _, found = availabilityMap[nodeGroup.Id()]; !found {
			size, err := nodeGroup.TargetSize()
			if err != nil {
				glog.Errorf("Failed to get size for %s: %v ", nodeGroup.Id(), err)
				continue
			}
			available = size - nodeGroup.MinSize()
			if available < 0 {
				available = 0
			}
			availabilityMap[nodeGroup.Id()] = available
		}
		if available > 0 {
			available -= 1
			availabilityMap[nodeGroup.Id()] = available
			result = append(result, node)
		}
	}
	limit := maxEmptyBulkDelete
	if len(result) < limit {
		limit = len(result)
	}
	return result[:limit]
}

func deleteNode(node *apiv1.Node, pods []*apiv1.Pod, client kube_client.Interface, cloudProvider cloudprovider.CloudProvider, recorder kube_record.EventRecorder) error {
	if err := drainNode(node, pods, client, recorder); err != nil {
		return err
	}
	return deleteNodeFromCloudProvider(node, cloudProvider, recorder)
}

// Performs drain logic on the node. Marks the node as unschedulable and later removes all pods, giving
// them up to MaxGracefulTerminationTime to finish.
func drainNode(node *apiv1.Node, pods []*apiv1.Pod, client kube_client.Interface, recorder kube_record.EventRecorder) error {
	if err := markToBeDeleted(node, client, recorder); err != nil {
		return err
	}

	seconds := int64(MaxGracefulTerminationTime.Seconds())
	for _, pod := range pods {
		recorder.Eventf(pod, apiv1.EventTypeNormal, "ScaleDown", "deleting pod for node scale down")
		err := client.Core().Pods(pod.Namespace).Delete(pod.Name, &apiv1.DeleteOptions{
			GracePeriodSeconds: &seconds,
		})
		if err != nil {
			glog.Errorf("Failed to delete %s/%s: %v", pod.Namespace, pod.Name, err)
		}
	}
	allGone := true

	// Wait up to MaxGracefulTerminationTime.
	for start := time.Now(); time.Now().Sub(start) < MaxGracefulTerminationTime; time.Sleep(5 * time.Second) {
		allGone = true
		for _, pod := range pods {
			podreturned, err := client.Core().Pods(pod.Namespace).Get(pod.Name)
			if err == nil {
				glog.Errorf("Not deleted yet %v", podreturned)
				allGone = false
				break
			}
			if !errors.IsNotFound(err) {
				glog.Errorf("Failed to check pod %s/%s: %v", pod.Namespace, pod.Name, err)
				allGone = false
			}
		}
		if allGone {
			glog.V(1).Infof("All pods removed from %s", node.Name)
			break
		}
	}
	if !allGone {
		glog.Warningf("Not all pods were removed from %s, proceeding anyway", node.Name)
	}
	return nil
}

// Sets unschedulable=true and adds an annotation.
func markToBeDeleted(node *apiv1.Node, client kube_client.Interface, recorder kube_record.EventRecorder) error {
	// Get the newest version of the node.
	freshNode, err := client.Core().Nodes().Get(node.Name)
	if err != nil || freshNode == nil {
		return fmt.Errorf("failed to get node %v: %v", node.Name, err)
	}

	added, err := addToBeDeletedTaint(freshNode)
	if added == false {
		return err
	}
	_, err = client.Core().Nodes().Update(freshNode)
	if err != nil {
		glog.Warningf("Error while adding taints on node %v: %v", node.Name, err)
		return err
	}
	glog.V(1).Infof("Successfully added toBeDeletedTaint on node %v", node.Name)
	recorder.Eventf(node, apiv1.EventTypeNormal, "ScaleDown", "marking the node as unschedulable")
	return nil
}

func addToBeDeletedTaint(node *apiv1.Node) (bool, error) {
	taints, err := apiv1.GetTaintsFromNodeAnnotations(node.Annotations)
	if err != nil {
		glog.Warningf("Error while getting Taints for node %v: %v", node.Name, err)
		return false, err
	}
	for _, taint := range taints {
		if taint.Key == ToBeDeletedTaint {
			glog.Infof("ToBeDeletedTaint already present on on node %v", taint, node.Name)
			return false, nil
		}
	}
	taints = append(taints, apiv1.Taint{
		Key:    ToBeDeletedTaint,
		Value:  time.Now().String(),
		Effect: apiv1.TaintEffectNoSchedule,
	})
	taintsJson, err := json.Marshal(taints)
	if err != nil {
		glog.Warningf("Error while adding taints on node %v: %v", node.Name, err)
		return false, err
	}
	if node.Annotations == nil {
		node.Annotations = make(map[string]string)
	}
	node.Annotations[apiv1.TaintsAnnotationKey] = string(taintsJson)
	return true, nil
}

// cleanToBeDeleted clean ToBeDeleted taints.
func cleanToBeDeleted(nodes []*apiv1.Node, client kube_client.Interface, recorder kube_record.EventRecorder) error {
	for _, node := range nodes {

		taints, err := apiv1.GetTaintsFromNodeAnnotations(node.Annotations)
		if err != nil {
			glog.Warningf("Error while getting Taints for node %v: %v", node.Name, err)
			continue
		}

		newTaints := make([]apiv1.Taint, 0)
		for _, taint := range taints {
			if taint.Key == ToBeDeletedTaint {
				glog.Infof("Releasing taint %+v on node %v", taint, node.Name)
			} else {
				newTaints = append(newTaints, taint)
			}
		}

		if len(newTaints) != len(taints) {
			taintsJson, err := json.Marshal(newTaints)
			if err != nil {
				glog.Warningf("Error while releasing taints on node %v: %v", node.Name, err)
				continue
			}
			if node.Annotations == nil {
				node.Annotations = make(map[string]string)
			}
			node.Annotations[apiv1.TaintsAnnotationKey] = string(taintsJson)
			_, err = client.Core().Nodes().Update(node)
			if err != nil {
				glog.Warningf("Error while releasing taints on node %v: %v", node.Name, err)
			} else {
				glog.V(1).Infof("Successfully released toBeDeletedTaint on node %v", node.Name)
				recorder.Eventf(node, apiv1.EventTypeNormal, "ClusterAutoscalerCleanup", "marking the node as schedulable")
			}
		}
	}
	return nil
}

// Removes the given node from cloud provider. No extra pre-deletion actions are executed on
// the Kubernetes side.
func deleteNodeFromCloudProvider(node *apiv1.Node, cloudProvider cloudprovider.CloudProvider, recorder kube_record.EventRecorder) error {
	nodeGroup, err := cloudProvider.NodeGroupForNode(node)
	if err != nil {
		return fmt.Errorf("failed to node group for %s: %v", node.Name, err)
	}
	if nodeGroup == nil || reflect.ValueOf(nodeGroup).IsNil() {
		return fmt.Errorf("picked node that doesn't belong to a node group: %s", node.Name)
	}
	if err = nodeGroup.DeleteNodes([]*apiv1.Node{node}); err != nil {
		return fmt.Errorf("failed to delete %s: %v", node.Name, err)
	}
	recorder.Eventf(node, apiv1.EventTypeNormal, "ScaleDown", "node removed by cluster autoscaler")
	return nil
}
