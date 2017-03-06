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

package core

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"k8s.io/contrib/cluster-autoscaler/cloudprovider"
	"k8s.io/contrib/cluster-autoscaler/clusterstate"
	"k8s.io/contrib/cluster-autoscaler/simulator"
	"k8s.io/contrib/cluster-autoscaler/utils/deletetaint"
	kube_util "k8s.io/contrib/cluster-autoscaler/utils/kubernetes"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_record "k8s.io/client-go/tools/record"
	apiv1 "k8s.io/kubernetes/pkg/api/v1"
	policyv1 "k8s.io/kubernetes/pkg/apis/policy/v1beta1"
	kube_client "k8s.io/kubernetes/pkg/client/clientset_generated/clientset"
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
	// MaxKubernetesEmptyNodeDeletionTime is the maximum time needed by Kubernetes to delete an empty node.
	MaxKubernetesEmptyNodeDeletionTime = 3 * time.Minute
	// MaxCloudProviderNodeDeletionTime is the maximum time needed by cloud provider to delete a node.
	MaxCloudProviderNodeDeletionTime = 5 * time.Minute
)

// ScaleDown is responsible for maintaining the state needed to perform unneded node removals.
type ScaleDown struct {
	context            *AutoscalingContext
	unneededNodes      map[string]time.Time
	unneededNodesList  []*apiv1.Node
	podLocationHints   map[string]string
	nodeUtilizationMap map[string]float64
	usageTracker       *simulator.UsageTracker
}

// NewScaleDown builds new ScaleDown object.
func NewScaleDown(context *AutoscalingContext) *ScaleDown {
	return &ScaleDown{
		context:            context,
		unneededNodes:      make(map[string]time.Time),
		podLocationHints:   make(map[string]string),
		nodeUtilizationMap: make(map[string]float64),
		usageTracker:       simulator.NewUsageTracker(),
		unneededNodesList:  make([]*apiv1.Node, 0),
	}
}

// CleanUp cleans up the internal ScaleDown state.
func (sd *ScaleDown) CleanUp(timestamp time.Time) {
	sd.usageTracker.CleanUp(time.Now().Add(-(sd.context.ScaleDownUnneededTime)))
}

// GetCandidatesForScaleDown gets candidates for scale down.
func (sd *ScaleDown) GetCandidatesForScaleDown() []*apiv1.Node {
	return sd.unneededNodesList
}

// UpdateUnneededNodes calculates which nodes are not needed, i.e. all pods can be scheduled somewhere else,
// and updates unneededNodes map accordingly. It also computes information where pods can be rescheduled and
// node utilization level. Timestamp is the current timestamp.
func (sd *ScaleDown) UpdateUnneededNodes(
	nodes []*apiv1.Node,
	pods []*apiv1.Pod,
	timestamp time.Time,
	pdbs []*policyv1.PodDisruptionBudget) error {

	currentlyUnneededNodes := make([]*apiv1.Node, 0)
	nodeNameToNodeInfo := schedulercache.CreateNodeNameToInfoMap(pods, nodes)
	utilizationMap := make(map[string]float64)

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

		if utilization >= sd.context.ScaleDownUtilizationThreshold {
			glog.V(4).Infof("Node %s is not suitable for removal - utilization too big (%f)", node.Name, utilization)
			continue
		}
		currentlyUnneededNodes = append(currentlyUnneededNodes, node)
	}

	// Phase2 - check which nodes can be probably removed using fast drain.
	nodesToRemove, newHints, err := simulator.FindNodesToRemove(currentlyUnneededNodes, nodes, pods,
		nil, sd.context.PredicateChecker,
		len(currentlyUnneededNodes), true, sd.podLocationHints, sd.usageTracker, timestamp, pdbs)
	if err != nil {
		glog.Errorf("Error while simulating node drains: %v", err)

		sd.unneededNodesList = make([]*apiv1.Node, 0)
		sd.unneededNodes = make(map[string]time.Time)
		sd.nodeUtilizationMap = make(map[string]float64)

		return fmt.Errorf("error while simulating node drains: %v", err)
	}

	// Update the timestamp map.
	result := make(map[string]time.Time)
	unneadedNodeList := make([]*apiv1.Node, 0, len(nodesToRemove))
	for _, node := range nodesToRemove {
		name := node.Node.Name
		unneadedNodeList = append(unneadedNodeList, node.Node)
		if val, found := sd.unneededNodes[name]; !found {
			result[name] = timestamp
		} else {
			result[name] = val
		}
	}

	sd.unneededNodesList = unneadedNodeList
	sd.unneededNodes = result
	sd.podLocationHints = newHints
	sd.nodeUtilizationMap = utilizationMap
	return nil
}

// TryToScaleDown tries to scale down the cluster. It returns ScaleDownResult indicating if any node was
// removed and error if such occured.
func (sd *ScaleDown) TryToScaleDown(nodes []*apiv1.Node, pods []*apiv1.Pod, pdbs []*policyv1.PodDisruptionBudget) (ScaleDownResult, error) {

	now := time.Now()
	candidates := make([]*apiv1.Node, 0)
	for _, node := range nodes {
		if val, found := sd.unneededNodes[node.Name]; found {

			glog.V(2).Infof("%s was unneeded for %s", node.Name, now.Sub(val).String())

			ready, _, _ := kube_util.GetReadinessState(node)

			// Check how long the node was underutilized.
			if ready && !val.Add(sd.context.ScaleDownUnneededTime).Before(now) {
				continue
			}

			// Unready nodes may be deleted after a different time than unrerutilized.
			if !ready && !val.Add(sd.context.ScaleDownUnreadyTime).Before(now) {
				continue
			}

			nodeGroup, err := sd.context.CloudProvider.NodeGroupForNode(node)
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
	emptyNodes := getEmptyNodes(candidates, pods, sd.context.MaxEmptyBulkDelete, sd.context.CloudProvider)
	if len(emptyNodes) > 0 {
		confirmation := make(chan error, len(emptyNodes))
		for _, node := range emptyNodes {
			glog.V(0).Infof("Scale-down: removing empty node %s", node.Name)
			sd.context.LogRecorder.Eventf(apiv1.EventTypeNormal, "ScaleDownEmpty", "Scale-down: removing empty node %s", node.Name)
			simulator.RemoveNodeFromTracker(sd.usageTracker, node.Name, sd.unneededNodes)
			go func(nodeToDelete *apiv1.Node) {
				confirmation <- deleteNodeFromCloudProvider(nodeToDelete, sd.context.CloudProvider,
					sd.context.Recorder, sd.context.ClusterStateRegistry)
			}(node)
		}
		var finalError error

		startTime := time.Now()
		for range emptyNodes {
			timeElapsed := time.Now().Sub(startTime)
			timeLeft := MaxCloudProviderNodeDeletionTime - timeElapsed
			if timeLeft < 0 {
				finalError = fmt.Errorf("Failed to delete nodes in time")
				break
			}
			select {
			case err := <-confirmation:
				if err != nil {
					glog.Errorf("Problem with empty node deletion: %v", err)
					finalError = err
				}
			case <-time.After(timeLeft):
				finalError = fmt.Errorf("Failed to delete nodes in time")
			}
		}
		if finalError == nil {
			return ScaleDownNodeDeleted, nil
		}
		return ScaleDownError, fmt.Errorf("failed to delete at least one empty node: %v", finalError)
	}

	// We look for only 1 node so new hints may be incomplete.
	nodesToRemove, _, err := simulator.FindNodesToRemove(candidates, nodes, pods, sd.context.ClientSet,
		sd.context.PredicateChecker, 1, false,
		sd.podLocationHints, sd.usageTracker, time.Now(), pdbs)

	if err != nil {
		return ScaleDownError, fmt.Errorf("Find node to remove failed: %v", err)
	}
	if len(nodesToRemove) == 0 {
		glog.V(1).Infof("No node to remove")
		return ScaleDownNoNodeDeleted, nil
	}
	toRemove := nodesToRemove[0]
	utilization := sd.nodeUtilizationMap[toRemove.Node.Name]
	podNames := make([]string, 0, len(toRemove.PodsToReschedule))
	for _, pod := range toRemove.PodsToReschedule {
		podNames = append(podNames, pod.Namespace+"/"+pod.Name)
	}
	glog.V(0).Infof("Scale-down: removing node %s, utilization: %v, pods to reschedule: %s", toRemove.Node.Name, utilization,
		strings.Join(podNames, ","))
	sd.context.LogRecorder.Eventf(apiv1.EventTypeNormal, "ScaleDown", "Scale-down: removing node %s, utilization: %v, pods to reschedule: %s",
		toRemove.Node.Name, utilization, strings.Join(podNames, ","))

	// Nothing super-bad should happen if the node is removed from tracker prematurely.
	simulator.RemoveNodeFromTracker(sd.usageTracker, toRemove.Node.Name, sd.unneededNodes)
	err = deleteNode(sd.context, toRemove.Node, toRemove.PodsToReschedule)
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

func deleteNode(context *AutoscalingContext, node *apiv1.Node, pods []*apiv1.Pod) error {
	if err := drainNode(node, pods, context.ClientSet, context.Recorder, context.MaxGratefulTerminationSec); err != nil {
		return err
	}
	return deleteNodeFromCloudProvider(node, context.CloudProvider, context.Recorder, context.ClusterStateRegistry)
}

// Performs drain logic on the node. Marks the node as unschedulable and later removes all pods, giving
// them up to MaxGracefulTerminationTime to finish.
func drainNode(node *apiv1.Node, pods []*apiv1.Pod, client kube_client.Interface, recorder kube_record.EventRecorder,
	maxGratefulTerminationSec int) error {
	if err := deletetaint.MarkToBeDeleted(node, client); err != nil {
		recorder.Eventf(node, apiv1.EventTypeWarning, "ScaleDown", "failed to mark the node as toBeDeleted/unschedulable: %v", err)
		return err
	}
	recorder.Eventf(node, apiv1.EventTypeNormal, "ScaleDown", "marked the node as toBeDeleted/unschedulable")

	maxGraceful64 := int64(maxGratefulTerminationSec)
	for _, pod := range pods {
		recorder.Eventf(pod, apiv1.EventTypeNormal, "ScaleDown", "deleting pod for node scale down")
		err := client.Core().Pods(pod.Namespace).Delete(pod.Name, &metav1.DeleteOptions{
			GracePeriodSeconds: &maxGraceful64,
		})
		if err != nil {
			glog.Errorf("Failed to delete %s/%s: %v", pod.Namespace, pod.Name, err)
		}
	}
	allGone := true

	// Wait up to MaxGracefulTerminationTime.
	for start := time.Now(); time.Now().Sub(start) < time.Duration(maxGratefulTerminationSec)*time.Second; time.Sleep(5 * time.Second) {
		allGone = true
		for _, pod := range pods {
			podreturned, err := client.Core().Pods(pod.Namespace).Get(pod.Name, metav1.GetOptions{})
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

// cleanToBeDeleted cleans ToBeDeleted taints.
func cleanToBeDeleted(nodes []*apiv1.Node, client kube_client.Interface, recorder kube_record.EventRecorder) {
	for _, node := range nodes {
		cleaned, err := deletetaint.CleanToBeDeleted(node, client)
		if err != nil {
			glog.Warningf("Error while releasing taints on node %v: %v", node.Name, err)
			recorder.Eventf(node, apiv1.EventTypeWarning, "ClusterAutoscalerCleanup",
				"failed to clean toBeDeletedTaint: %v", err)
		} else if cleaned {
			glog.V(1).Infof("Successfully released toBeDeletedTaint on node %v", node.Name)
			recorder.Eventf(node, apiv1.EventTypeNormal, "ClusterAutoscalerCleanup", "marking the node as schedulable")
		}
	}
}

// Removes the given node from cloud provider. No extra pre-deletion actions are executed on
// the Kubernetes side.
func deleteNodeFromCloudProvider(node *apiv1.Node, cloudProvider cloudprovider.CloudProvider,
	recorder kube_record.EventRecorder, registry *clusterstate.ClusterStateRegistry) error {
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
	registry.RegisterScaleDown(&clusterstate.ScaleDownRequest{
		NodeGroupName:      nodeGroup.Id(),
		NodeName:           node.Name,
		Time:               time.Now(),
		ExpectedDeleteTime: time.Now().Add(MaxCloudProviderNodeDeletionTime),
	})
	return nil
}
