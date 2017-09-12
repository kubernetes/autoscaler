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
	"math"
	"reflect"
	"strings"
	"sync"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate"
	"k8s.io/autoscaler/cluster-autoscaler/metrics"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	"k8s.io/autoscaler/cluster-autoscaler/utils/deletetaint"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"

	apiv1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1beta1"
	kube_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_client "k8s.io/client-go/kubernetes"
	kube_record "k8s.io/client-go/tools/record"
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
	// ScaleDownNodeDeleteStarted - a node deletion process was started.
	ScaleDownNodeDeleteStarted ScaleDownResult = iota
	// ScaleDownDisabledKey is the name of annotation marking node as not eligible for scale down.
	ScaleDownDisabledKey = "cluster-autoscaler.kubernetes.io/scale-down-disabled"
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
	PodEvictionHeadroom = 30 * time.Second
	// UnremovableNodeRecheckTimeout is the timeout before we check again a node that couldn't be removed before
	UnremovableNodeRecheckTimeout = 5 * time.Minute
)

// NodeDeleteStatus tells whether a node is being deleted right now.
type NodeDeleteStatus struct {
	sync.Mutex
	deleteInProgress bool
}

// IsDeleteInProgress returns true if a node is being deleted.
func (n *NodeDeleteStatus) IsDeleteInProgress() bool {
	n.Lock()
	defer n.Unlock()
	return n.deleteInProgress
}

// SetDeleteInProgress sets deletion process status
func (n *NodeDeleteStatus) SetDeleteInProgress(status bool) {
	n.Lock()
	defer n.Unlock()
	n.deleteInProgress = status
}

// ScaleDown is responsible for maintaining the state needed to perform unneded node removals.
type ScaleDown struct {
	context            *AutoscalingContext
	unneededNodes      map[string]time.Time
	unneededNodesList  []*apiv1.Node
	unremovableNodes   map[string]time.Time
	podLocationHints   map[string]string
	nodeUtilizationMap map[string]float64
	usageTracker       *simulator.UsageTracker
	nodeDeleteStatus   *NodeDeleteStatus
}

// NewScaleDown builds new ScaleDown object.
func NewScaleDown(context *AutoscalingContext) *ScaleDown {
	return &ScaleDown{
		context:            context,
		unneededNodes:      make(map[string]time.Time),
		unremovableNodes:   make(map[string]time.Time),
		podLocationHints:   make(map[string]string),
		nodeUtilizationMap: make(map[string]float64),
		usageTracker:       simulator.NewUsageTracker(),
		unneededNodesList:  make([]*apiv1.Node, 0),
		nodeDeleteStatus:   &NodeDeleteStatus{},
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
// node utilization level. Timestamp is the current timestamp. The computations are made only for the nodes
// managed by CA.
func (sd *ScaleDown) UpdateUnneededNodes(
	nodes []*apiv1.Node,
	nodesToCheck []*apiv1.Node,
	pods []*apiv1.Pod,
	timestamp time.Time,
	pdbs []*policyv1.PodDisruptionBudget) errors.AutoscalerError {

	currentlyUnneededNodes := make([]*apiv1.Node, 0)
	nodeNameToNodeInfo := schedulercache.CreateNodeNameToInfoMap(pods, nodes)
	utilizationMap := make(map[string]float64)

	sd.updateUnremovableNodes(nodes)
	// Filter out nodes that were recently checked
	filteredNodesToCheck := make([]*apiv1.Node, 0)
	for _, node := range nodesToCheck {
		if unremovableTimestamp, found := sd.unremovableNodes[node.Name]; found {
			if unremovableTimestamp.After(timestamp) {
				continue
			}
			delete(sd.unremovableNodes, node.Name)
		}
		filteredNodesToCheck = append(filteredNodesToCheck, node)
	}
	skipped := len(nodesToCheck) - len(filteredNodesToCheck)
	if skipped > 0 {
		glog.V(1).Infof("Scale-down calculation: ignoring %v nodes, that were unremovable in the last %v", skipped, UnremovableNodeRecheckTimeout)
	}

	// Phase1 - look at the nodes utilization. Calculate the utilization
	// only for the managed nodes.
	for _, node := range filteredNodesToCheck {

		// Skip nodes marked to be deleted, if they were marked recently.
		// Old-time marked nodes are again eligible for deletion - something went wrong with them
		// and they have not been deleted.
		deleteTime, _ := deletetaint.GetToBeDeletedTime(node)
		if deleteTime != nil && (timestamp.Sub(*deleteTime) < MaxCloudProviderNodeDeletionTime || timestamp.Sub(*deleteTime) < MaxKubernetesEmptyNodeDeletionTime) {
			glog.V(1).Infof("Skipping %s from delete considerations - the node is currently being deleted", node.Name)
			continue
		}

		// Skip nodes marked with no scale down annotation
		if hasNoScaleDownAnnotation(node) {
			glog.V(1).Infof("Skipping %s from delete consideration - the node is marked as no scale down", node.Name)
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

	emptyNodes := make(map[string]bool)

	emptyNodesList := getEmptyNodes(currentlyUnneededNodes, pods, len(currentlyUnneededNodes), sd.context.CloudProvider)
	for _, node := range emptyNodesList {
		emptyNodes[node.Name] = true
	}

	currentlyUnneededNonEmptyNodes := make([]*apiv1.Node, 0, len(currentlyUnneededNodes))
	for _, node := range currentlyUnneededNodes {
		if !emptyNodes[node.Name] {
			currentlyUnneededNonEmptyNodes = append(currentlyUnneededNonEmptyNodes, node)
		}
	}

	// Phase2 - check which nodes can be probably removed using fast drain.
	currentCandidates, currentNonCandidates := sd.chooseCandidates(currentlyUnneededNonEmptyNodes)

	// Look for nodes to remove in the current candidates
	nodesToRemove, unremovable, newHints, simulatorErr := simulator.FindNodesToRemove(
		currentCandidates, nodes, pods, nil, sd.context.PredicateChecker,
		len(currentCandidates), true, sd.podLocationHints, sd.usageTracker, timestamp, pdbs)
	if simulatorErr != nil {
		return sd.markSimulationError(simulatorErr, timestamp)
	}

	additionalCandidatesCount := sd.context.ScaleDownNonEmptyCandidatesCount - len(nodesToRemove)
	if additionalCandidatesCount > len(currentNonCandidates) {
		additionalCandidatesCount = len(currentNonCandidates)
	}
	// Limit the additional candidates pool size for better performance.
	additionalCandidatesPoolSize := int(math.Ceil(float64(len(nodes)) * sd.context.ScaleDownCandidatesPoolRatio))
	if additionalCandidatesPoolSize < sd.context.ScaleDownCandidatesPoolMinCount {
		additionalCandidatesPoolSize = sd.context.ScaleDownCandidatesPoolMinCount
	}
	if additionalCandidatesPoolSize > len(currentNonCandidates) {
		additionalCandidatesPoolSize = len(currentNonCandidates)
	}
	if additionalCandidatesCount > 0 {
		// Look for addidtional nodes to remove among the rest of nodes
		glog.V(3).Infof("Finding additional %v candidates for scale down.", additionalCandidatesCount)
		additionalNodesToRemove, additionalUnremovable, additionalNewHints, simulatorErr :=
			simulator.FindNodesToRemove(currentNonCandidates[:additionalCandidatesPoolSize], nodes, pods, nil,
				sd.context.PredicateChecker, additionalCandidatesCount, true,
				sd.podLocationHints, sd.usageTracker, timestamp, pdbs)
		if simulatorErr != nil {
			return sd.markSimulationError(simulatorErr, timestamp)
		}
		nodesToRemove = append(nodesToRemove, additionalNodesToRemove...)
		unremovable = append(unremovable, additionalUnremovable...)
		for key, value := range additionalNewHints {
			newHints[key] = value
		}
	}

	for _, node := range emptyNodesList {
		nodesToRemove = append(nodesToRemove, simulator.NodeToBeRemoved{Node: node, PodsToReschedule: []*apiv1.Pod{}})
	}
	// Update the timestamp map.
	result := make(map[string]time.Time)
	unneededNodesList := make([]*apiv1.Node, 0, len(nodesToRemove))
	for _, node := range nodesToRemove {
		name := node.Node.Name
		unneededNodesList = append(unneededNodesList, node.Node)
		if val, found := sd.unneededNodes[name]; !found {
			result[name] = timestamp
		} else {
			result[name] = val
		}
	}

	// Add nodes to unremovable map
	if len(unremovable) > 0 {
		unremovableTimeout := timestamp.Add(UnremovableNodeRecheckTimeout)
		for _, node := range unremovable {
			sd.unremovableNodes[node.Name] = unremovableTimeout
		}
		glog.V(1).Infof("%v nodes found unremovable in simulation, will re-check them at %v", len(unremovable), unremovableTimeout)
	}

	// Update state and metrics
	sd.unneededNodesList = unneededNodesList
	sd.unneededNodes = result
	sd.podLocationHints = newHints
	sd.nodeUtilizationMap = utilizationMap
	sd.context.ClusterStateRegistry.UpdateScaleDownCandidates(sd.unneededNodesList, timestamp)
	metrics.UpdateUnneededNodesCount(len(sd.unneededNodesList))
	return nil
}

// updateUnremovableNodes updates unremovableNodes map according to current
// state of the cluster. Removes from the map nodes that are no longer in the
// nodes list.
func (sd *ScaleDown) updateUnremovableNodes(nodes []*apiv1.Node) {
	if len(sd.unremovableNodes) <= 0 {
		return
	}
	// A set of nodes to delete from unremovableNodes map.
	nodesToDelete := make(map[string]struct{}, len(sd.unremovableNodes))
	for name := range sd.unremovableNodes {
		nodesToDelete[name] = struct{}{}
	}
	// Nodes that are in the cluster should not be deleted.
	for _, node := range nodes {
		if _, ok := nodesToDelete[node.Name]; ok {
			delete(nodesToDelete, node.Name)
		}
	}
	for nodeName := range nodesToDelete {
		delete(sd.unremovableNodes, nodeName)
	}
}

// markSimulationError indicates a simulation error by clearing  relevant scale
// down state and returning an apropriate error.
func (sd *ScaleDown) markSimulationError(simulatorErr errors.AutoscalerError,
	timestamp time.Time) errors.AutoscalerError {
	glog.Errorf("Error while simulating node drains: %v", simulatorErr)
	sd.unneededNodesList = make([]*apiv1.Node, 0)
	sd.unneededNodes = make(map[string]time.Time)
	sd.nodeUtilizationMap = make(map[string]float64)
	sd.context.ClusterStateRegistry.UpdateScaleDownCandidates(sd.unneededNodesList, timestamp)
	return simulatorErr.AddPrefix("error while simulating node drains: ")
}

// chooseCandidates splits nodes into current candidates for scaledown and the
// rest. Current candidates are unneeded nodes from the previous run that are
// still in the nodes list.
func (sd *ScaleDown) chooseCandidates(nodes []*apiv1.Node) ([]*apiv1.Node, []*apiv1.Node) {
	// Number of candidates should not be capped. We will look for nodes to remove
	// from the whole set of nodes.
	if sd.context.ScaleDownNonEmptyCandidatesCount <= 0 {
		return nodes, []*apiv1.Node{}
	}
	currentCandidates := make([]*apiv1.Node, 0, len(sd.unneededNodesList))
	currentNonCandidates := make([]*apiv1.Node, 0, len(nodes))
	for _, node := range nodes {
		if _, found := sd.unneededNodes[node.Name]; found {
			currentCandidates = append(currentCandidates, node)
		} else {
			currentNonCandidates = append(currentNonCandidates, node)
		}
	}
	return currentCandidates, currentNonCandidates
}

// TryToScaleDown tries to scale down the cluster. It returns ScaleDownResult indicating if any node was
// removed and error if such occurred.
func (sd *ScaleDown) TryToScaleDown(nodes []*apiv1.Node, pods []*apiv1.Pod, pdbs []*policyv1.PodDisruptionBudget) (ScaleDownResult, errors.AutoscalerError) {
	now := time.Now()
	nodeDeletionDuration := time.Duration(0)
	findNodesToRemoveDuration := time.Duration(0)
	defer updateScaleDownMetrics(now, &findNodesToRemoveDuration, &nodeDeletionDuration)
	candidates := make([]*apiv1.Node, 0)
	readinessMap := make(map[string]bool)

	for _, node := range nodes {
		if val, found := sd.unneededNodes[node.Name]; found {

			glog.V(2).Infof("%s was unneeded for %s", node.Name, now.Sub(val).String())

			// Check if node is marked with no scale down annotation.
			if hasNoScaleDownAnnotation(node) {
				glog.V(4).Infof("Skipping %s - scale down disabled annotation found", node.Name)
				continue
			}

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
		glog.V(1).Infof("No candidates for scale down")
		return ScaleDownNoUnneeded, nil
	}

	// Trying to delete empty nodes in bulk. If there are no empty nodes then CA will
	// try to delete not-so-empty nodes, possibly killing some pods and allowing them
	// to recreate on other nodes.
	emptyNodes := getEmptyNodes(candidates, pods, sd.context.MaxEmptyBulkDelete, sd.context.CloudProvider)
	if len(emptyNodes) > 0 {
		nodeDeletionStart := time.Now()
		confirmation := make(chan errors.AutoscalerError, len(emptyNodes))
		sd.scheduleDeleteEmptyNodes(emptyNodes, readinessMap, confirmation)
		err := sd.waitForEmptyNodesDeleted(emptyNodes, confirmation)
		nodeDeletionDuration = time.Now().Sub(nodeDeletionStart)
		if err == nil {
			return ScaleDownNodeDeleted, nil
		}
		return ScaleDownError, err.AddPrefix("failed to delete at least one empty node: ")
	}

	findNodesToRemoveStart := time.Now()
	// We look for only 1 node so new hints may be incomplete.
	nodesToRemove, _, _, err := simulator.FindNodesToRemove(candidates, nodes, pods, sd.context.ClientSet,
		sd.context.PredicateChecker, 1, false,
		sd.podLocationHints, sd.usageTracker, time.Now(), pdbs)
	findNodesToRemoveDuration = time.Now().Sub(findNodesToRemoveStart)

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
	nodeDeletionStart := time.Now()

	// Starting deletion.
	nodeDeletionDuration = time.Now().Sub(nodeDeletionStart)
	sd.nodeDeleteStatus.SetDeleteInProgress(true)

	go func() {
		// Finishing the delete probess once this goroutine is over.
		defer sd.nodeDeleteStatus.SetDeleteInProgress(false)
		err := deleteNode(sd.context, toRemove.Node, toRemove.PodsToReschedule)
		if err != nil {
			glog.Errorf("Failed to delete %s: %v", toRemove.Node.Name, err)
			return
		}
		if readinessMap[toRemove.Node.Name] {
			metrics.RegisterScaleDown(1, metrics.Underutilized)
		} else {
			metrics.RegisterScaleDown(1, metrics.Unready)
		}
	}()

	return ScaleDownNodeDeleteStarted, nil
}

// updateScaleDownMetrics registers duration of different parts of scale down.
// Separates time spent on finding nodes to remove, deleting nodes and other operations.
func updateScaleDownMetrics(scaleDownStart time.Time, findNodesToRemoveDuration *time.Duration, nodeDeletionDuration *time.Duration) {
	stop := time.Now()
	miscDuration := stop.Sub(scaleDownStart) - *nodeDeletionDuration - *findNodesToRemoveDuration
	metrics.UpdateDuration(metrics.ScaleDownNodeDeletion, *nodeDeletionDuration)
	metrics.UpdateDuration(metrics.ScaleDownFindNodesToRemove, *findNodesToRemoveDuration)
	metrics.UpdateDuration(metrics.ScaleDownMiscOperations, miscDuration)
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
		if available, found = availabilityMap[nodeGroup.Id()]; !found {
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

func (sd *ScaleDown) scheduleDeleteEmptyNodes(emptyNodes []*apiv1.Node,
	readinessMap map[string]bool, confirmation chan errors.AutoscalerError) {
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
}

func (sd *ScaleDown) waitForEmptyNodesDeleted(emptyNodes []*apiv1.Node, confirmation chan errors.AutoscalerError) errors.AutoscalerError {
	var finalError errors.AutoscalerError

	startTime := time.Now()
	for range emptyNodes {
		timeElapsed := time.Now().Sub(startTime)
		timeLeft := MaxCloudProviderNodeDeletionTime - timeElapsed
		if timeLeft < 0 {
			return errors.NewAutoscalerError(errors.TransientError, "Failed to delete nodes in time")
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
	return finalError
}

func deleteNode(context *AutoscalingContext, node *apiv1.Node, pods []*apiv1.Pod) errors.AutoscalerError {
	if err := drainNode(node, pods, context.ClientSet, context.Recorder, context.MaxGracefulTerminationSec,
		MaxPodEvictionTime, EvictionRetryTime); err != nil {
		return err
	}
	return deleteNodeFromCloudProvider(node, context.CloudProvider, context.Recorder, context.ClusterStateRegistry)
}

func evictPod(podToEvict *apiv1.Pod, client kube_client.Interface, recorder kube_record.EventRecorder,
	maxGracefulTerminationSec int, retryUntil time.Time, waitBetweenRetries time.Duration) error {
	recorder.Eventf(podToEvict, apiv1.EventTypeNormal, "ScaleDown", "deleting pod for node scale down")

	maxTermination := int64(apiv1.DefaultTerminationGracePeriodSeconds)
	if podToEvict.Spec.TerminationGracePeriodSeconds != nil {
		if *podToEvict.Spec.TerminationGracePeriodSeconds < int64(maxGracefulTerminationSec) {
			maxTermination = *podToEvict.Spec.TerminationGracePeriodSeconds
		} else {
			maxTermination = int64(maxGracefulTerminationSec)
		}
	}

	var lastError error
	for first := true; first || time.Now().Before(retryUntil); time.Sleep(waitBetweenRetries) {
		first = false
		eviction := &policyv1.Eviction{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: podToEvict.Namespace,
				Name:      podToEvict.Name,
			},
			DeleteOptions: &metav1.DeleteOptions{
				GracePeriodSeconds: &maxTermination,
			},
		}
		lastError = client.CoreV1().Pods(podToEvict.Namespace).Evict(eviction)
		if lastError == nil || kube_errors.IsNotFound(lastError) {
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
	maxGracefulTerminationSec int, maxPodEvictionTime time.Duration, waitBetweenRetries time.Duration) errors.AutoscalerError {

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
			confirmations <- evictPod(podToEvict, client, recorder, maxGracefulTerminationSec, retryUntil, waitBetweenRetries)
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

	// Evictions created successfully, wait maxGracefulTerminationSec + PodEvictionHeadroom to see if pods really disappeared.
	allGone := true
	for start := time.Now(); time.Now().Sub(start) < time.Duration(maxGracefulTerminationSec)*time.Second+PodEvictionHeadroom; time.Sleep(5 * time.Second) {
		allGone = true
		for _, pod := range pods {
			podreturned, err := client.CoreV1().Pods(pod.Namespace).Get(pod.Name, metav1.GetOptions{})
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
			// Let the deferred function know there is no need for cleanup
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

func hasNoScaleDownAnnotation(node *apiv1.Node) bool {
	return node.Annotations[ScaleDownDisabledKey] == "true"
}

func cleanUpNodeAutoprovisionedGroups(cloudProvider cloudprovider.CloudProvider) error {
	nodeGroups := cloudProvider.NodeGroups()
	for _, nodeGroup := range nodeGroups {
		if !nodeGroup.Autoprovisioned() {
			continue
		}
		size, err := nodeGroup.TargetSize()
		if err != nil {
			return err
		}
		if size == 0 {
			if err := nodeGroup.Delete(); err != nil {
				return err
			}
		}
	}
	return nil
}
