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

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate"
	"k8s.io/autoscaler/cluster-autoscaler/metrics"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	"k8s.io/autoscaler/cluster-autoscaler/utils/deletetaint"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"

	kube_errors "k8s.io/apimachinery/pkg/api/errors"
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
	// MaxPodEvictionTime is the maximum time CA tries to evict a pod before giving up.
	MaxPodEvictionTime = 2 * time.Minute
	// EvictionRetryTime is the time after CA retries failed pod eviction.
	EvictionRetryTime = 10 * time.Second
	// PodEvictionHeadroom is the extra time we wait to catch situations when the pod is ignoring SIGTERM and
	// is killed with SIGKILL after MaxGracefulTerminationTime
	PodEvictionHeadroom = 20 * time.Second
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

// CleanUpUnneededNodes clears the list of unneeded nodes.
func (sd *ScaleDown) CleanUpUnneededNodes() {
	sd.unneededNodesList = make([]*apiv1.Node, 0)
	sd.unneededNodes = make(map[string]time.Time)
}

// UpdateUnneededNodes calculates which nodes are not needed, i.e. all pods can be scheduled somewhere else,
// and updates unneededNodes map accordingly. It also computes information where pods can be rescheduled and
// node utilization level. Timestamp is the current timestamp.
func (sd *ScaleDown) UpdateUnneededNodes(
	nodes []*apiv1.Node,
	pods []*apiv1.Pod,
	timestamp time.Time,
	pdbs []*policyv1.PodDisruptionBudget) errors.AutoscalerError {

	currentlyUnneededNodes := make([]*apiv1.Node, 0)
	nodeNameToNodeInfo := schedulercache.CreateNodeNameToInfoMap(pods, nodes)
	utilizationMap := make(map[string]float64)

	// Phase1 - look at the nodes utilization.
	for _, node := range nodes {

		// Skip nodes marked to be deleted, if they were marked recently.
		// Old-time marked nodes are again eligible for deletion - something went wrong with them
		// and they have not been deleted.
		deleteTime, _ := deletetaint.GetToBeDeletedTime(node)
		if deleteTime != nil && (timestamp.Sub(*deleteTime) < MaxCloudProviderNodeDeletionTime || timestamp.Sub(*deleteTime) < MaxKubernetesEmptyNodeDeletionTime) {
			glog.V(1).Info("Skipping %s from delete considerations - the node is currently being deleted", node.Name)
			continue
		}

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
	nodesToRemove, newHints, simulatorErr := simulator.FindNodesToRemove(currentlyUnneededNodes, nodes, pods,
		nil, sd.context.PredicateChecker,
		len(currentlyUnneededNodes), true, sd.podLocationHints, sd.usageTracker, timestamp, pdbs)
	if simulatorErr != nil {
		glog.Errorf("Error while simulating node drains: %v", simulatorErr)

		sd.unneededNodesList = make([]*apiv1.Node, 0)
		sd.unneededNodes = make(map[string]time.Time)
		sd.nodeUtilizationMap = make(map[string]float64)
		sd.context.ClusterStateRegistry.UpdateScaleDownCandidates(sd.unneededNodesList, timestamp)

		return simulatorErr.AddPrefix("error while simulating node drains: ")
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
	sd.context.ClusterStateRegistry.UpdateScaleDownCandidates(sd.unneededNodesList, timestamp)
	metrics.UpdateUnneededNodesCount(len(sd.unneededNodesList))
	return nil
}

// TryToScaleDown tries to scale down the cluster. It returns ScaleDownResult indicating if any node was
// removed and error if such occured.
func (sd *ScaleDown) TryToScaleDown(nodes []*apiv1.Node, pods []*apiv1.Pod, pdbs []*policyv1.PodDisruptionBudget) (ScaleDownResult, errors.AutoscalerError) {

	now := time.Now()
	candidates := make([]*apiv1.Node, 0)
	readinessMap := make(map[string]bool)

	for _, node := range nodes {
		if val, found := sd.unneededNodes[node.Name]; found {

			glog.V(2).Infof("%s was unneeded for %s", node.Name, now.Sub(val).String())

			ready, _, _ := kube_util.GetReadinessState(node)
			readinessMap[node.Name] = ready

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
		confirmation := make(chan errors.AutoscalerError, len(emptyNodes))
		for _, node := range emptyNodes {
			glog.V(0).Infof("Scale-down: removing empty node %s", node.Name)
			sd.context.LogRecorder.Eventf(apiv1.EventTypeNormal, "ScaleDownEmpty", "Scale-down: removing empty node %s", node.Name)
			simulator.RemoveNodeFromTracker(sd.usageTracker, node.Name, sd.unneededNodes)
			go func(nodeToDelete *apiv1.Node) {

				deleteErr := deleteNodeFromCloudProvider(nodeToDelete, sd.context.CloudProvider,
					sd.context.Recorder, sd.context.ClusterStateRegistry)
				if deleteErr == nil {
					if readinessMap[nodeToDelete.Name] {
						metrics.RegisterScaleDown(1, metrics.Empty)
					} else {
						metrics.RegisterScaleDown(1, metrics.Unready)
					}
				}
				confirmation <- deleteErr
			}(node)
		}
		var finalError errors.AutoscalerError

		startTime := time.Now()
		for range emptyNodes {
			timeElapsed := time.Now().Sub(startTime)
			timeLeft := MaxCloudProviderNodeDeletionTime - timeElapsed
			if timeLeft < 0 {
				finalError = errors.NewAutoscalerError(errors.TransientError, "Failed to delete nodes in time")
				break
			}
			select {
			case err := <-confirmation:
				if err != nil {
					glog.Errorf("Problem with empty node deletion: %v", err)
					finalError = err
				}
			case <-time.After(timeLeft):
				finalError = errors.NewAutoscalerError(errors.TransientError, "Failed to delete nodes in time")
			}
		}
		if finalError == nil {
			return ScaleDownNodeDeleted, nil
		}
		return ScaleDownError, finalError.AddPrefix("failed to delete at least one empty node: ")
	}

	// We look for only 1 node so new hints may be incomplete.
	nodesToRemove, _, err := simulator.FindNodesToRemove(candidates, nodes, pods, sd.context.ClientSet,
		sd.context.PredicateChecker, 1, false,
		sd.podLocationHints, sd.usageTracker, time.Now(), pdbs)

	if err != nil {
		return ScaleDownError, err.AddPrefix("Find node to remove failed: ")
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
		return ScaleDownError, err.AddPrefix("Failed to delete %s: ", toRemove.Node.Name)
	}
	if readinessMap[toRemove.Node.Name] {
		metrics.RegisterScaleDown(1, metrics.Underutilized)
	} else {
		metrics.RegisterScaleDown(1, metrics.Unready)
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

func deleteNode(context *AutoscalingContext, node *apiv1.Node, pods []*apiv1.Pod) errors.AutoscalerError {
	if err := drainNode(node, pods, context.ClientSet, context.Recorder, context.MaxGratefulTerminationSec,
		MaxPodEvictionTime, EvictionRetryTime); err != nil {
		return err
	}
	return deleteNodeFromCloudProvider(node, context.CloudProvider, context.Recorder, context.ClusterStateRegistry)
}

func evictPod(podToEvict *apiv1.Pod, client kube_client.Interface, recorder kube_record.EventRecorder,
	maxGratefulTerminationSec int, retryUntil time.Time, waitBetweenRetries time.Duration) error {
	recorder.Eventf(podToEvict, apiv1.EventTypeNormal, "ScaleDown", "deleting pod for node scale down")
	maxGraceful64 := int64(maxGratefulTerminationSec)
	var lastError error
	for first := true; first || time.Now().Before(retryUntil); time.Sleep(waitBetweenRetries) {
		first = false
		eviction := &policyv1.Eviction{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: podToEvict.Namespace,
				Name:      podToEvict.Name,
			},
			DeleteOptions: &metav1.DeleteOptions{
				GracePeriodSeconds: &maxGraceful64,
			},
		}
		lastError = client.Core().Pods(podToEvict.Namespace).Evict(eviction)
		if lastError == nil {
			return nil
		}
	}
	glog.Errorf("Failed to evict pod %s, error: %v", podToEvict.Name, lastError)
	recorder.Eventf(podToEvict, apiv1.EventTypeWarning, "ScaleDownFailed", "failed to delete pod for ScaleDown")
	return fmt.Errorf("Failed to evict pod %s/%s within allowed timeout (last error: %v)", podToEvict.Namespace, podToEvict.Name, lastError)
}

// Performs drain logic on the node. Marks the node as unschedulable and later removes all pods, giving
// them up to MaxGracefulTerminationTime to finish.
func drainNode(node *apiv1.Node, pods []*apiv1.Pod, client kube_client.Interface, recorder kube_record.EventRecorder,
	maxGratefulTerminationSec int, maxPodEvictionTime time.Duration, waitBetweenRetries time.Duration) errors.AutoscalerError {

	drainSuccessful := false
	toEvict := len(pods)
	if err := deletetaint.MarkToBeDeleted(node, client); err != nil {
		recorder.Eventf(node, apiv1.EventTypeWarning, "ScaleDownFailed", "failed to mark the node as toBeDeleted/unschedulable: %v", err)
		return errors.ToAutoscalerError(errors.ApiCallError, err)
	}

	// If we fail to evict all the pods from the node we want to remove delete taint
	defer func() {
		if !drainSuccessful {
			deletetaint.CleanToBeDeleted(node, client)
			recorder.Eventf(node, apiv1.EventTypeWarning, "ScaleDownFailed", "failed to drain the node, aborting ScaleDown")
		}
	}()

	recorder.Eventf(node, apiv1.EventTypeNormal, "ScaleDown", "marked the node as toBeDeleted/unschedulable")

	retryUntil := time.Now().Add(maxPodEvictionTime)
	confirmations := make(chan error, toEvict)
	for _, pod := range pods {
		go func(podToEvict *apiv1.Pod) {
			confirmations <- evictPod(podToEvict, client, recorder, maxGratefulTerminationSec, retryUntil, waitBetweenRetries)
		}(pod)
	}

	evictionErrs := make([]error, 0)

	for range pods {
		select {
		case err := <-confirmations:
			if err != nil {
				evictionErrs = append(evictionErrs, err)
			} else {
				metrics.RegisterEvictions(1)
			}
		case <-time.After(retryUntil.Sub(time.Now()) + 5*time.Second):
			return errors.NewAutoscalerError(
				errors.ApiCallError, "Failed to drain node %s/%s: timeout when waiting for creating evictions", node.Namespace, node.Name)
		}
	}
	if len(evictionErrs) != 0 {
		return errors.NewAutoscalerError(
			errors.ApiCallError, "Failed to drain node %s/%s, due to following errors: %v", node.Namespace, node.Name, evictionErrs)
	}

	// Evictions created successfully, wait maxGratefulTerminationSec + PodEvictionHeadroom to see if pods really disappeared.
	allGone := true
	for start := time.Now(); time.Now().Sub(start) < time.Duration(maxGratefulTerminationSec)*time.Second+PodEvictionHeadroom; time.Sleep(5 * time.Second) {
		allGone = true
		for _, pod := range pods {
			podreturned, err := client.Core().Pods(pod.Namespace).Get(pod.Name, metav1.GetOptions{})
			if err == nil {
				glog.Errorf("Not deleted yet %v", podreturned)
				allGone = false
				break
			}
			if !kube_errors.IsNotFound(err) {
				glog.Errorf("Failed to check pod %s/%s: %v", pod.Namespace, pod.Name, err)
				allGone = false
			}
		}
		if allGone {
			glog.V(1).Infof("All pods removed from %s", node.Name)
			// Let the defered function know there is no need for cleanup
			drainSuccessful = true
			return nil
		}
	}
	return errors.NewAutoscalerError(
		errors.TransientError, "Failed to drain node %s/%s: pods remaining after timeout", node.Namespace, node.Name)
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
	recorder kube_record.EventRecorder, registry *clusterstate.ClusterStateRegistry) errors.AutoscalerError {
	nodeGroup, err := cloudProvider.NodeGroupForNode(node)
	if err != nil {
		return errors.NewAutoscalerError(
			errors.CloudProviderError, "failed to find node group for %s: %v", node.Name, err)
	}
	if nodeGroup == nil || reflect.ValueOf(nodeGroup).IsNil() {
		return errors.NewAutoscalerError(errors.InternalError, "picked node that doesn't belong to a node group: %s", node.Name)
	}
	if err = nodeGroup.DeleteNodes([]*apiv1.Node{node}); err != nil {
		return errors.NewAutoscalerError(errors.CloudProviderError, "failed to delete %s: %v", node.Name, err)
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
