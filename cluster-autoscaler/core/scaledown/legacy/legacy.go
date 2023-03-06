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

package legacy

import (
	"math"
	"reflect"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/deletiontracker"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/eligibility"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/resource"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/status"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/unneeded"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/unremovable"
	"k8s.io/autoscaler/cluster-autoscaler/metrics"
	"k8s.io/autoscaler/cluster-autoscaler/processors"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/utilization"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	klog "k8s.io/klog/v2"
)

// ScaleDown is responsible for maintaining the state needed to perform unneeded node removals.
type ScaleDown struct {
	context              *context.AutoscalingContext
	processors           *processors.AutoscalingProcessors
	unremovableNodes     *unremovable.Nodes
	unneededNodes        *unneeded.Nodes
	nodeUtilizationMap   map[string]utilization.Info
	usageTracker         *simulator.UsageTracker
	nodeDeletionTracker  *deletiontracker.NodeDeletionTracker
	removalSimulator     *simulator.RemovalSimulator
	eligibilityChecker   *eligibility.Checker
	resourceLimitsFinder *resource.LimitsFinder
}

// NewScaleDown builds new ScaleDown object.
func NewScaleDown(context *context.AutoscalingContext, processors *processors.AutoscalingProcessors, ndt *deletiontracker.NodeDeletionTracker, deleteOptions simulator.NodeDeleteOptions) *ScaleDown {
	usageTracker := simulator.NewUsageTracker()
	removalSimulator := simulator.NewRemovalSimulator(context.ListerRegistry, context.ClusterSnapshot, context.PredicateChecker, usageTracker, deleteOptions, false)
	unremovableNodes := unremovable.NewNodes()
	resourceLimitsFinder := resource.NewLimitsFinder(processors.CustomResourcesProcessor)
	return &ScaleDown{
		context:              context,
		processors:           processors,
		unremovableNodes:     unremovableNodes,
		unneededNodes:        unneeded.NewNodes(processors.NodeGroupConfigProcessor, resourceLimitsFinder),
		nodeUtilizationMap:   make(map[string]utilization.Info),
		usageTracker:         usageTracker,
		nodeDeletionTracker:  ndt,
		removalSimulator:     removalSimulator,
		eligibilityChecker:   eligibility.NewChecker(processors.NodeGroupConfigProcessor),
		resourceLimitsFinder: resourceLimitsFinder,
	}
}

// CleanUp cleans up the internal ScaleDown state.
func (sd *ScaleDown) CleanUp(timestamp time.Time) {
	// Use default ScaleDownUnneededTime as in this context the value
	// doesn't apply to any specific NodeGroup.
	sd.usageTracker.CleanUp(timestamp.Add(-sd.context.NodeGroupDefaults.ScaleDownUnneededTime))
}

// CleanUpUnneededNodes clears the list of unneeded nodes.
func (sd *ScaleDown) CleanUpUnneededNodes() {
	sd.unneededNodes.Clear()
}

// UnneededNodes returns a list of nodes that can potentially be scaled down.
func (sd *ScaleDown) UnneededNodes() []*apiv1.Node {
	return sd.unneededNodes.AsList()
}

// UpdateUnneededNodes calculates which nodes are not needed, i.e. all pods can be scheduled somewhere else,
// and updates unneededNodes accordingly. It also computes information where pods can be rescheduled and
// node utilization level. The computations are made only for the nodes managed by CA.
// * destinationNodes are the nodes that can potentially take in any pods that are evicted because of a scale down.
// * scaleDownCandidates are the nodes that are being considered for scale down.
// * timestamp is the current timestamp.
// * pdbs is a list of pod disruption budgets.
func (sd *ScaleDown) UpdateUnneededNodes(
	destinationNodes []*apiv1.Node,
	scaleDownCandidates []*apiv1.Node,
	timestamp time.Time,
) errors.AutoscalerError {

	// Only scheduled non expendable pods and pods waiting for lower priority pods preemption can prevent node delete.
	// Extract cluster state from snapshot for initial analysis
	allNodeInfos, err := sd.context.ClusterSnapshot.NodeInfos().List()
	if err != nil {
		// This should never happen, List() returns err only because scheduler interface requires it.
		return errors.ToAutoscalerError(errors.InternalError, err)
	}

	// Phase1 - look at the nodes utilization. Calculate the utilization
	// only for the managed nodes.
	sd.unremovableNodes.Update(sd.context.ClusterSnapshot.NodeInfos(), timestamp)
	currentlyUnneededNodeNames, utilizationMap, ineligible := sd.eligibilityChecker.FilterOutUnremovable(sd.context, scaleDownCandidates, timestamp, sd.unremovableNodes)
	for _, n := range ineligible {
		sd.unremovableNodes.Add(n)
	}

	emptyNodesToRemove := sd.getEmptyNodesToRemoveNoResourceLimits(currentlyUnneededNodeNames, timestamp)

	emptyNodes := make(map[string]bool)
	for _, empty := range emptyNodesToRemove {
		emptyNodes[empty.Node.Name] = true
	}

	currentlyUnneededNonEmptyNodes := make([]string, 0, len(currentlyUnneededNodeNames))
	for _, node := range currentlyUnneededNodeNames {
		if !emptyNodes[node] {
			currentlyUnneededNonEmptyNodes = append(currentlyUnneededNonEmptyNodes, node)
		}
	}

	// Phase2 - check which nodes can be probably removed using fast drain.
	currentCandidates, currentNonCandidates := sd.chooseCandidates(currentlyUnneededNonEmptyNodes)

	destinations := make([]string, 0, len(destinationNodes))
	for _, destinationNode := range destinationNodes {
		destinations = append(destinations, destinationNode.Name)
	}

	// Look for nodes to remove in the current candidates
	nodesToRemove, unremovable := sd.removalSimulator.FindNodesToRemove(
		currentCandidates,
		destinations,
		timestamp,
		sd.context.RemainingPdbTracker.GetPdbs())

	additionalCandidatesCount := sd.context.ScaleDownNonEmptyCandidatesCount - len(nodesToRemove)
	if additionalCandidatesCount > len(currentNonCandidates) {
		additionalCandidatesCount = len(currentNonCandidates)
	}
	// Limit the additional candidates pool size for better performance.
	additionalCandidatesPoolSize := int(math.Ceil(float64(len(allNodeInfos)) * sd.context.ScaleDownCandidatesPoolRatio))
	if additionalCandidatesPoolSize < sd.context.ScaleDownCandidatesPoolMinCount {
		additionalCandidatesPoolSize = sd.context.ScaleDownCandidatesPoolMinCount
	}
	if additionalCandidatesPoolSize > len(currentNonCandidates) {
		additionalCandidatesPoolSize = len(currentNonCandidates)
	}
	if additionalCandidatesCount > 0 {
		// Look for additional nodes to remove among the rest of nodes.
		klog.V(3).Infof("Finding additional %v candidates for scale down.", additionalCandidatesCount)
		additionalNodesToRemove, additionalUnremovable :=
			sd.removalSimulator.FindNodesToRemove(
				currentNonCandidates[:additionalCandidatesPoolSize],
				destinations,
				timestamp,
				sd.context.RemainingPdbTracker.GetPdbs())
		if len(additionalNodesToRemove) > additionalCandidatesCount {
			additionalNodesToRemove = additionalNodesToRemove[:additionalCandidatesCount]
		}
		nodesToRemove = append(nodesToRemove, additionalNodesToRemove...)
		unremovable = append(unremovable, additionalUnremovable...)
	}

	for _, empty := range emptyNodesToRemove {
		nodesToRemove = append(nodesToRemove, simulator.NodeToBeRemoved{Node: empty.Node, PodsToReschedule: []*apiv1.Pod{}})
	}

	// Update the timestamp map.
	sd.unneededNodes.Update(nodesToRemove, timestamp)

	// Add nodes to unremovable map
	if len(unremovable) > 0 {
		unremovableTimeout := timestamp.Add(sd.context.AutoscalingOptions.UnremovableNodeRecheckTimeout)
		for _, unremovableNode := range unremovable {
			sd.unremovableNodes.AddTimeout(unremovableNode, unremovableTimeout)
		}
		klog.V(1).Infof("%v nodes found to be unremovable in simulation, will re-check them at %v", len(unremovable), unremovableTimeout)
	}

	// This method won't always check all nodes, so let's give a generic reason for all nodes that weren't checked.
	for _, node := range scaleDownCandidates {
		unremovableReasonProvided := sd.unremovableNodes.HasReason(node.Name)
		unneeded := sd.unneededNodes.Contains(node.Name)
		if !unneeded && !unremovableReasonProvided {
			sd.unremovableNodes.AddReason(node, simulator.NotUnneededOtherReason)
		}
	}

	// Update state and metrics
	sd.removalSimulator.DropOldHints()
	sd.nodeUtilizationMap = utilizationMap
	return nil
}

// NodeUtilizationMap returns the most recent mapping from node names to utilization info.
func (sd *ScaleDown) NodeUtilizationMap() map[string]utilization.Info {
	return sd.nodeUtilizationMap
}

// UnremovableNodes returns a list of nodes that cannot be removed according to
// the scale down algorithm.
func (sd *ScaleDown) UnremovableNodes() []*simulator.UnremovableNode {
	return sd.unremovableNodes.AsList()
}

// chooseCandidates splits nodes into current candidates for scale-down and the
// rest. Current candidates are unneeded nodes from the previous run that are
// still in the nodes list.
func (sd *ScaleDown) chooseCandidates(nodes []string) (candidates []string, nonCandidates []string) {
	// Number of candidates should not be capped. We will look for nodes to remove
	// from the whole set of nodes.
	if sd.context.ScaleDownNonEmptyCandidatesCount <= 0 {
		return nodes, nil
	}
	for _, node := range nodes {
		if sd.unneededNodes.Contains(node) {
			candidates = append(candidates, node)
		} else {
			nonCandidates = append(nonCandidates, node)
		}
	}
	return candidates, nonCandidates
}

func (sd *ScaleDown) mapNodesToStatusScaleDownNodes(nodes []*apiv1.Node, nodeGroups map[string]cloudprovider.NodeGroup, evictedPodLists map[string][]*apiv1.Pod) []*status.ScaleDownNode {
	var result []*status.ScaleDownNode
	for _, node := range nodes {
		result = append(result, &status.ScaleDownNode{
			Node:        node,
			NodeGroup:   nodeGroups[node.Name],
			UtilInfo:    sd.nodeUtilizationMap[node.Name],
			EvictedPods: evictedPodLists[node.Name],
		})
	}
	return result
}

// NodesToDelete selects the nodes to delete for scale down.
func (sd *ScaleDown) NodesToDelete(currentTime time.Time) (_, drain []*apiv1.Node, res status.ScaleDownResult, err errors.AutoscalerError) {
	_, drained := sd.nodeDeletionTracker.DeletionsInProgress()
	if len(drained) > 0 {
		return nil, nil, status.ScaleDownInProgress, nil
	}

	findNodesToRemoveDuration := time.Duration(0)
	defer updateScaleDownMetrics(time.Now(), &findNodesToRemoveDuration)

	allNodeInfos, errSnapshot := sd.context.ClusterSnapshot.NodeInfos().List()
	if errSnapshot != nil {
		// This should never happen, List() returns err only because scheduler interface requires it.
		return nil, nil, status.ScaleDownError, errors.ToAutoscalerError(errors.InternalError, errSnapshot)
	}

	allNodes := make([]*apiv1.Node, 0, len(allNodeInfos))
	allNodeNames := make([]string, 0, len(allNodeInfos))
	for _, ni := range allNodeInfos {
		allNodes = append(allNodes, ni.Node())
		allNodeNames = append(allNodeNames, ni.Node().Name)
	}

	resourceLimiter, errCP := sd.context.CloudProvider.GetResourceLimiter()
	if errCP != nil {
		return nil, nil, status.ScaleDownError, errors.ToAutoscalerError(errors.CloudProviderError, errCP)
	}

	scaleDownResourcesLeft := sd.resourceLimitsFinder.LimitsLeft(sd.context, allNodes, resourceLimiter, currentTime)
	empty, nonEmpty, unremovable := sd.unneededNodes.RemovableAt(sd.context, currentTime, scaleDownResourcesLeft, resourceLimiter.GetResources(), sd.nodeDeletionTracker)
	for _, u := range unremovable {
		sd.unremovableNodes.Add(u)
	}
	candidateNames := make([]string, 0, len(empty)+len(nonEmpty))
	for _, n := range empty {
		candidateNames = append(candidateNames, n.Node.Name)
	}
	for _, n := range nonEmpty {
		candidateNames = append(candidateNames, n.Node.Name)
	}

	if len(candidateNames) == 0 {
		klog.V(1).Infof("No candidates for scale down")
		return nil, nil, status.ScaleDownNoUnneeded, nil
	}

	// Trying to delete empty nodes in bulk. If there are no empty nodes then CA will
	// try to delete not-so-empty nodes, possibly killing some pods and allowing them
	// to recreate on other nodes.
	emptyNodesToRemove := sd.getEmptyNodesToRemove(candidateNames, scaleDownResourcesLeft, currentTime)
	emptyNodesToRemove = sd.processors.ScaleDownSetProcessor.GetNodesToRemove(sd.context, emptyNodesToRemove, sd.context.MaxEmptyBulkDelete)
	if len(emptyNodesToRemove) > 0 {
		var nodes []*apiv1.Node
		for _, node := range emptyNodesToRemove {
			// Nothing super-bad should happen if the node is removed from tracker prematurely.
			sd.removeNodeFromTracker(node.Node.Name)
			nodes = append(nodes, node.Node)
		}
		return nodes, nil, status.ScaleDownNodeDeleteStarted, nil
	}

	findNodesToRemoveStart := time.Now()
	nodesToRemove, unremovable := sd.removalSimulator.FindNodesToRemove(
		candidateNames,
		allNodeNames,
		time.Now(),
		sd.context.RemainingPdbTracker.GetPdbs())
	findNodesToRemoveDuration = time.Now().Sub(findNodesToRemoveStart)

	for _, unremovableNode := range unremovable {
		sd.unremovableNodes.Add(unremovableNode)
	}

	nodesToRemove = sd.processors.ScaleDownSetProcessor.GetNodesToRemove(sd.context, nodesToRemove, 1)
	if len(nodesToRemove) == 0 {
		klog.V(1).Infof("No node to remove")
		return nil, nil, status.ScaleDownNoNodeDeleted, nil
	}
	toRemove := nodesToRemove[0]
	// Nothing super-bad should happen if the node is removed from tracker prematurely.
	sd.removeNodeFromTracker(toRemove.Node.Name)
	return nil, []*apiv1.Node{toRemove.Node}, status.ScaleDownNodeDeleteStarted, nil
}

func (sd *ScaleDown) removeNodeFromTracker(node string) {
	unneeded := make([]string, 0, len(sd.unneededNodes.AsList()))
	for _, n := range sd.unneededNodes.AsList() {
		unneeded = append(unneeded, n.Name)
	}
	toRemove := simulator.RemoveNodeFromTracker(sd.usageTracker, node, unneeded)
	for _, n := range toRemove {
		sd.unneededNodes.Drop(n)
	}
}

// updateScaleDownMetrics registers duration of different parts of scale down.
// Separates time spent on finding nodes to remove, deleting nodes and other operations.
func updateScaleDownMetrics(scaleDownStart time.Time, findNodesToRemoveDuration *time.Duration) {
	stop := time.Now()
	miscDuration := stop.Sub(scaleDownStart) - *findNodesToRemoveDuration
	metrics.UpdateDuration(metrics.ScaleDownFindNodesToRemove, *findNodesToRemoveDuration)
	metrics.UpdateDuration(metrics.ScaleDownMiscOperations, miscDuration)
}

func (sd *ScaleDown) getEmptyNodesToRemoveNoResourceLimits(candidates []string, timestamp time.Time) []simulator.NodeToBeRemoved {
	return sd.getEmptyNodesToRemove(candidates, resource.NoLimits(), timestamp)
}

// This functions finds empty nodes among passed candidates and returns a list of empty nodes
// that can be deleted at the same time.
func (sd *ScaleDown) getEmptyNodesToRemove(candidates []string, resourcesLimits resource.Limits,
	timestamp time.Time) []simulator.NodeToBeRemoved {

	emptyNodes := sd.removalSimulator.FindEmptyNodesToRemove(candidates, timestamp)
	availabilityMap := make(map[string]int)
	nodesToRemove := make([]simulator.NodeToBeRemoved, 0)
	resourcesLimitsCopy := resourcesLimits.DeepCopy() // we do not want to modify input parameter
	resourcesNames := sets.StringKeySet(resourcesLimits).List()
	for _, nodeName := range emptyNodes {
		nodeInfo, err := sd.context.ClusterSnapshot.NodeInfos().Get(nodeName)
		if err != nil {
			klog.Errorf("Can't retrieve node %s from snapshot, err: %v", nodeName, err)
			continue
		}
		node := nodeInfo.Node()
		nodeGroup, err := sd.context.CloudProvider.NodeGroupForNode(node)
		if err != nil {
			klog.Errorf("Failed to get group for %s", nodeName)
			continue
		}
		if nodeGroup == nil || reflect.ValueOf(nodeGroup).IsNil() {
			continue
		}
		var available int
		var found bool
		if available, found = availabilityMap[nodeGroup.Id()]; !found {
			// Will be cached.
			size, err := nodeGroup.TargetSize()
			if err != nil {
				klog.Errorf("Failed to get size for %s: %v ", nodeGroup.Id(), err)
				continue
			}
			deletionsInProgress := sd.nodeDeletionTracker.DeletionsCount(nodeGroup.Id())
			available = size - nodeGroup.MinSize() - deletionsInProgress
			if available < 0 {
				available = 0
			}
			availabilityMap[nodeGroup.Id()] = available
		}
		if available > 0 {
			resourcesDelta, err := sd.resourceLimitsFinder.DeltaForNode(sd.context, node, nodeGroup, resourcesNames)
			if err != nil {
				klog.Errorf("Error: %v", err)
				continue
			}
			checkResult := resourcesLimitsCopy.TryDecrementBy(resourcesDelta)
			if checkResult.Exceeded() {
				continue
			}
			available--
			availabilityMap[nodeGroup.Id()] = available
			nodesToRemove = append(nodesToRemove, simulator.NodeToBeRemoved{
				Node: node,
			})
		}
	}

	return nodesToRemove
}
