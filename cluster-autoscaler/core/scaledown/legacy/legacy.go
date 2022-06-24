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
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/actuation"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/deletiontracker"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/unremovable"
	core_utils "k8s.io/autoscaler/cluster-autoscaler/core/utils"
	"k8s.io/autoscaler/cluster-autoscaler/metrics"
	"k8s.io/autoscaler/cluster-autoscaler/processors"
	"k8s.io/autoscaler/cluster-autoscaler/processors/customresources"
	"k8s.io/autoscaler/cluster-autoscaler/processors/status"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/utilization"
	"k8s.io/autoscaler/cluster-autoscaler/utils"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"

	apiv1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	klog "k8s.io/klog/v2"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
)

const (
	// ScaleDownDisabledKey is the name of annotation marking node as not eligible for scale down.
	ScaleDownDisabledKey = "cluster-autoscaler.kubernetes.io/scale-down-disabled"
)

type scaleDownResourcesLimits map[string]int64
type scaleDownResourcesDelta map[string]int64

// used as a value in scaleDownResourcesLimits if actual limit could not be obtained due to errors talking to cloud provider
const scaleDownLimitUnknown = math.MinInt64

func (sd *ScaleDown) computeScaleDownResourcesLeftLimits(nodes []*apiv1.Node, resourceLimiter *cloudprovider.ResourceLimiter, cp cloudprovider.CloudProvider, timestamp time.Time) scaleDownResourcesLimits {
	totalCores, totalMem := calculateScaleDownCoresMemoryTotal(nodes, timestamp)

	var totalResources map[string]int64
	var totalResourcesErr error
	if cloudprovider.ContainsCustomResources(resourceLimiter.GetResources()) {
		totalResources, totalResourcesErr = sd.calculateScaleDownCustomResourcesTotal(nodes, cp, timestamp)
	}

	resultScaleDownLimits := make(scaleDownResourcesLimits)
	for _, resource := range resourceLimiter.GetResources() {
		min := resourceLimiter.GetMin(resource)

		// we put only actual limits into final map. No entry means no limit.
		if min > 0 {
			switch {
			case resource == cloudprovider.ResourceNameCores:
				resultScaleDownLimits[resource] = computeAboveMin(totalCores, min)
			case resource == cloudprovider.ResourceNameMemory:
				resultScaleDownLimits[resource] = computeAboveMin(totalMem, min)
			case cloudprovider.IsCustomResource(resource):
				if totalResourcesErr != nil {
					resultScaleDownLimits[resource] = scaleDownLimitUnknown
				} else {
					resultScaleDownLimits[resource] = computeAboveMin(totalResources[resource], min)
				}
			default:
				klog.Errorf("Scale down limits defined for unsupported resource '%s'", resource)
			}
		}
	}
	return resultScaleDownLimits
}

func computeAboveMin(total int64, min int64) int64 {
	if total > min {
		return total - min
	}
	return 0

}

func calculateScaleDownCoresMemoryTotal(nodes []*apiv1.Node, timestamp time.Time) (int64, int64) {
	var coresTotal, memoryTotal int64
	for _, node := range nodes {
		if actuation.IsNodeBeingDeleted(node, timestamp) {
			// Nodes being deleted do not count towards total cluster resources
			continue
		}
		cores, memory := core_utils.GetNodeCoresAndMemory(node)

		coresTotal += cores
		memoryTotal += memory
	}

	return coresTotal, memoryTotal
}

func (sd *ScaleDown) calculateScaleDownCustomResourcesTotal(nodes []*apiv1.Node, cp cloudprovider.CloudProvider, timestamp time.Time) (map[string]int64, error) {
	result := make(map[string]int64)
	ngCache := make(map[string][]customresources.CustomResourceTarget)
	for _, node := range nodes {
		if actuation.IsNodeBeingDeleted(node, timestamp) {
			// Nodes being deleted do not count towards total cluster resources
			continue
		}
		nodeGroup, err := cp.NodeGroupForNode(node)
		if err != nil {
			return nil, errors.ToAutoscalerError(errors.CloudProviderError, err).AddPrefix("can not get node group for node %v when calculating cluster gpu usage", node.Name)
		}
		if nodeGroup == nil || reflect.ValueOf(nodeGroup).IsNil() {
			// We do not trust cloud providers to return properly constructed nil for interface type - hence the reflection check.
			// See https://golang.org/doc/faq#nil_error
			// TODO[lukaszos] consider creating cloud_provider sanitizer which will wrap cloud provider and ensure sane behaviour.
			nodeGroup = nil
		}

		var resourceTargets []customresources.CustomResourceTarget
		var cacheHit bool

		if nodeGroup != nil {
			resourceTargets, cacheHit = ngCache[nodeGroup.Id()]
		}
		if !cacheHit {
			resourceTargets, err = sd.processors.CustomResourcesProcessor.GetNodeResourceTargets(sd.context, node, nodeGroup)
			if err != nil {
				return nil, errors.ToAutoscalerError(errors.CloudProviderError, err).AddPrefix("can not get gpu count for node %v when calculating cluster gpu usage")
			}
			if nodeGroup != nil {
				ngCache[nodeGroup.Id()] = resourceTargets
			}
		}

		for _, resourceTarget := range resourceTargets {
			if resourceTarget.ResourceType == "" || resourceTarget.ResourceCount == 0 {
				continue
			}
			result[resourceTarget.ResourceType] += resourceTarget.ResourceCount
		}
	}

	return result, nil
}

func noScaleDownLimitsOnResources() scaleDownResourcesLimits {
	return nil
}

func copyScaleDownResourcesLimits(source scaleDownResourcesLimits) scaleDownResourcesLimits {
	copy := scaleDownResourcesLimits{}
	for k, v := range source {
		copy[k] = v
	}
	return copy
}

func (sd *ScaleDown) computeScaleDownResourcesDelta(cp cloudprovider.CloudProvider, node *apiv1.Node, nodeGroup cloudprovider.NodeGroup, resourcesWithLimits []string) (scaleDownResourcesDelta, errors.AutoscalerError) {
	resultScaleDownDelta := make(scaleDownResourcesDelta)

	nodeCPU, nodeMemory := core_utils.GetNodeCoresAndMemory(node)
	resultScaleDownDelta[cloudprovider.ResourceNameCores] = nodeCPU
	resultScaleDownDelta[cloudprovider.ResourceNameMemory] = nodeMemory

	if cloudprovider.ContainsCustomResources(resourcesWithLimits) {
		resourceTargets, err := sd.processors.CustomResourcesProcessor.GetNodeResourceTargets(sd.context, node, nodeGroup)
		if err != nil {
			return scaleDownResourcesDelta{}, errors.ToAutoscalerError(errors.CloudProviderError, err).AddPrefix("Failed to get node %v custom resources: %v", node.Name)
		}
		for _, resourceTarget := range resourceTargets {
			resultScaleDownDelta[resourceTarget.ResourceType] = resourceTarget.ResourceCount
		}
	}
	return resultScaleDownDelta, nil
}

type scaleDownLimitsCheckResult struct {
	exceeded          bool
	exceededResources []string
}

func scaleDownLimitsNotExceeded() scaleDownLimitsCheckResult {
	return scaleDownLimitsCheckResult{false, []string{}}
}

func (limits *scaleDownResourcesLimits) checkScaleDownDeltaWithinLimits(delta scaleDownResourcesDelta) scaleDownLimitsCheckResult {
	exceededResources := sets.NewString()
	for resource, resourceDelta := range delta {
		resourceLeft, found := (*limits)[resource]
		if found {
			if (resourceDelta > 0) && (resourceLeft == scaleDownLimitUnknown || resourceDelta > resourceLeft) {
				exceededResources.Insert(resource)
			}
		}
	}
	if len(exceededResources) > 0 {
		return scaleDownLimitsCheckResult{true, exceededResources.List()}
	}

	return scaleDownLimitsNotExceeded()
}

func (limits *scaleDownResourcesLimits) tryDecrementLimitsByDelta(delta scaleDownResourcesDelta) scaleDownLimitsCheckResult {
	result := limits.checkScaleDownDeltaWithinLimits(delta)
	if result.exceeded {
		return result
	}
	for resource, resourceDelta := range delta {
		resourceLeft, found := (*limits)[resource]
		if found {
			(*limits)[resource] = resourceLeft - resourceDelta
		}
	}
	return scaleDownLimitsNotExceeded()
}

// ScaleDown is responsible for maintaining the state needed to perform unneeded node removals.
type ScaleDown struct {
	context              *context.AutoscalingContext
	processors           *processors.AutoscalingProcessors
	clusterStateRegistry *clusterstate.ClusterStateRegistry
	unneededNodes        map[string]time.Time
	unneededNodesList    []*apiv1.Node
	unremovableNodes     *unremovable.Nodes
	podLocationHints     map[string]string
	nodeUtilizationMap   map[string]utilization.Info
	usageTracker         *simulator.UsageTracker
	nodeDeletionTracker  *deletiontracker.NodeDeletionTracker
	removalSimulator     *simulator.RemovalSimulator
}

// NewScaleDown builds new ScaleDown object.
func NewScaleDown(context *context.AutoscalingContext, processors *processors.AutoscalingProcessors, clusterStateRegistry *clusterstate.ClusterStateRegistry, ndt *deletiontracker.NodeDeletionTracker) *ScaleDown {
	usageTracker := simulator.NewUsageTracker()
	removalSimulator := simulator.NewRemovalSimulator(context.ListerRegistry, context.ClusterSnapshot, context.PredicateChecker, usageTracker)
	return &ScaleDown{
		context:              context,
		processors:           processors,
		clusterStateRegistry: clusterStateRegistry,
		unneededNodes:        make(map[string]time.Time),
		unremovableNodes:     unremovable.NewNodes(),
		podLocationHints:     make(map[string]string),
		nodeUtilizationMap:   make(map[string]utilization.Info),
		usageTracker:         usageTracker,
		unneededNodesList:    make([]*apiv1.Node, 0),
		nodeDeletionTracker:  ndt,
		removalSimulator:     removalSimulator,
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
	sd.unneededNodesList = make([]*apiv1.Node, 0)
	sd.unneededNodes = make(map[string]time.Time)
}

// UnneededNodes returns a list of nodes that can potentially be scaled down.
func (sd *ScaleDown) UnneededNodes() []*apiv1.Node {
	return sd.unneededNodesList
}

func (sd *ScaleDown) checkNodeUtilization(timestamp time.Time, node *apiv1.Node, nodeInfo *schedulerframework.NodeInfo) (simulator.UnremovableReason, *utilization.Info) {
	// Skip nodes that were recently checked.
	if sd.unremovableNodes.IsRecent(node.Name) {
		return simulator.RecentlyUnremovable, nil
	}

	// Skip nodes marked to be deleted, if they were marked recently.
	// Old-time marked nodes are again eligible for deletion - something went wrong with them
	// and they have not been deleted.
	if actuation.IsNodeBeingDeleted(node, timestamp) {
		klog.V(1).Infof("Skipping %s from delete consideration - the node is currently being deleted", node.Name)
		return simulator.CurrentlyBeingDeleted, nil
	}

	// Skip nodes marked with no scale down annotation
	if hasNoScaleDownAnnotation(node) {
		klog.V(1).Infof("Skipping %s from delete consideration - the node is marked as no scale down", node.Name)
		return simulator.ScaleDownDisabledAnnotation, nil
	}

	utilInfo, err := utilization.Calculate(node, nodeInfo, sd.context.IgnoreDaemonSetsUtilization, sd.context.IgnoreMirrorPodsUtilization, sd.context.CloudProvider.GPULabel(), timestamp)
	if err != nil {
		klog.Warningf("Failed to calculate utilization for %s: %v", node.Name, err)
	}

	nodeGroup, err := sd.context.CloudProvider.NodeGroupForNode(node)
	if err != nil {
		return simulator.UnexpectedError, nil
	}
	if nodeGroup == nil || reflect.ValueOf(nodeGroup).IsNil() {
		// We should never get here as non-autoscaled nodes should not be included in scaleDownCandidates list
		// (and the default PreFilteringScaleDownNodeProcessor would indeed filter them out).
		klog.Warningf("Skipped %s from delete consideration - the node is not autoscaled", node.Name)
		return simulator.NotAutoscaled, nil
	}

	underutilized, err := sd.isNodeBelowUtilizationThreshold(node, nodeGroup, utilInfo)
	if err != nil {
		klog.Warningf("Failed to check utilization thresholds for %s: %v", node.Name, err)
		return simulator.UnexpectedError, nil
	}
	if !underutilized {
		klog.V(4).Infof("Node %s is not suitable for removal - %s utilization too big (%f)", node.Name, utilInfo.ResourceName, utilInfo.Utilization)
		return simulator.NotUnderutilized, &utilInfo
	}

	klog.V(4).Infof("Node %s - %s utilization %f", node.Name, utilInfo.ResourceName, utilInfo.Utilization)

	return simulator.NoReason, &utilInfo
}

// UpdateUnneededNodes calculates which nodes are not needed, i.e. all pods can be scheduled somewhere else,
// and updates unneededNodes map accordingly. It also computes information where pods can be rescheduled and
// node utilization level. The computations are made only for the nodes managed by CA.
// * destinationNodes are the nodes that can potentially take in any pods that are evicted because of a scale down.
// * scaleDownCandidates are the nodes that are being considered for scale down.
// * timestamp is the current timestamp.
// * pdbs is a list of pod disruption budgets.
func (sd *ScaleDown) UpdateUnneededNodes(
	destinationNodes []*apiv1.Node,
	scaleDownCandidates []*apiv1.Node,
	timestamp time.Time,
	pdbs []*policyv1.PodDisruptionBudget,
) errors.AutoscalerError {

	// Only scheduled non expendable pods and pods waiting for lower priority pods preemption can prevent node delete.
	// Extract cluster state from snapshot for initial analysis
	allNodeInfos, err := sd.context.ClusterSnapshot.NodeInfos().List()
	if err != nil {
		// This should never happen, List() returns err only because scheduler interface requires it.
		return errors.ToAutoscalerError(errors.InternalError, err)
	}

	sd.unremovableNodes.Update(sd.context.ClusterSnapshot.NodeInfos(), timestamp)

	skipped := 0
	utilizationMap := make(map[string]utilization.Info)
	currentlyUnneededNodeNames := make([]string, 0, len(scaleDownCandidates))

	// Phase1 - look at the nodes utilization. Calculate the utilization
	// only for the managed nodes.
	for _, node := range scaleDownCandidates {
		nodeInfo, err := sd.context.ClusterSnapshot.NodeInfos().Get(node.Name)
		if err != nil {
			klog.Errorf("Can't retrieve scale-down candidate %s from snapshot, err: %v", node.Name, err)
			sd.unremovableNodes.AddReason(node, simulator.UnexpectedError)
			continue
		}

		reason, utilInfo := sd.checkNodeUtilization(timestamp, node, nodeInfo)
		if utilInfo != nil {
			utilizationMap[node.Name] = *utilInfo
		}
		if reason != simulator.NoReason {
			// For logging purposes.
			if reason == simulator.RecentlyUnremovable {
				skipped++
			}

			sd.unremovableNodes.AddReason(node, reason)
			continue
		}

		currentlyUnneededNodeNames = append(currentlyUnneededNodeNames, node.Name)
	}

	if skipped > 0 {
		klog.V(1).Infof("Scale-down calculation: ignoring %v nodes unremovable in the last %v", skipped, sd.context.AutoscalingOptions.UnremovableNodeRecheckTimeout)
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
	nodesToRemove, unremovable, newHints, simulatorErr := sd.removalSimulator.FindNodesToRemove(
		currentCandidates,
		destinations,
		sd.podLocationHints,
		timestamp,
		pdbs)
	if simulatorErr != nil {
		return sd.markSimulationError(simulatorErr, timestamp)
	}

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
		additionalNodesToRemove, additionalUnremovable, additionalNewHints, simulatorErr :=
			sd.removalSimulator.FindNodesToRemove(
				currentNonCandidates[:additionalCandidatesPoolSize],
				destinations,
				sd.podLocationHints,
				timestamp,
				pdbs)
		if simulatorErr != nil {
			return sd.markSimulationError(simulatorErr, timestamp)
		}
		if len(additionalNodesToRemove) > additionalCandidatesCount {
			additionalNodesToRemove = additionalNodesToRemove[:additionalCandidatesCount]
		}
		nodesToRemove = append(nodesToRemove, additionalNodesToRemove...)
		unremovable = append(unremovable, additionalUnremovable...)
		for key, value := range additionalNewHints {
			newHints[key] = value
		}
	}

	for _, empty := range emptyNodesToRemove {
		nodesToRemove = append(nodesToRemove, simulator.NodeToBeRemoved{Node: empty.Node, PodsToReschedule: []*apiv1.Pod{}})
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
		unremovableTimeout := timestamp.Add(sd.context.AutoscalingOptions.UnremovableNodeRecheckTimeout)
		for _, unremovableNode := range unremovable {
			sd.unremovableNodes.AddTimeout(unremovableNode, unremovableTimeout)
		}
		klog.V(1).Infof("%v nodes found to be unremovable in simulation, will re-check them at %v", len(unremovable), unremovableTimeout)
	}

	// This method won't always check all nodes, so let's give a generic reason for all nodes that weren't checked.
	for _, node := range scaleDownCandidates {
		unremovableReasonProvided := sd.unremovableNodes.HasReason(node.Name)
		_, unneeded := result[node.Name]
		if !unneeded && !unremovableReasonProvided {
			sd.unremovableNodes.AddReason(node, simulator.NotUnneededOtherReason)
		}
	}

	// Update state and metrics
	sd.unneededNodesList = unneededNodesList
	sd.unneededNodes = result
	sd.podLocationHints = newHints
	sd.nodeUtilizationMap = utilizationMap
	sd.clusterStateRegistry.UpdateScaleDownCandidates(sd.unneededNodesList, timestamp)
	metrics.UpdateUnneededNodesCount(len(sd.unneededNodesList))
	if klog.V(4).Enabled() {
		for key, val := range sd.unneededNodes {
			klog.Infof("%s is unneeded since %s duration %s", key, val.String(), timestamp.Sub(val).String())
		}
	}
	return nil
}

// NodeUtilizationMap returns the most recent mapping from node names to utilization info.
func (sd *ScaleDown) NodeUtilizationMap() map[string]utilization.Info {
	return sd.nodeUtilizationMap
}

// isNodeBelowUtilizationThreshold determines if a given node utilization is below threshold.
func (sd *ScaleDown) isNodeBelowUtilizationThreshold(node *apiv1.Node, nodeGroup cloudprovider.NodeGroup, utilInfo utilization.Info) (bool, error) {
	var threshold float64
	var err error
	if gpu.NodeHasGpu(sd.context.CloudProvider.GPULabel(), node) {
		threshold, err = sd.processors.NodeGroupConfigProcessor.GetScaleDownGpuUtilizationThreshold(sd.context, nodeGroup)
		if err != nil {
			return false, err
		}
	} else {
		threshold, err = sd.processors.NodeGroupConfigProcessor.GetScaleDownUtilizationThreshold(sd.context, nodeGroup)
		if err != nil {
			return false, err
		}
	}
	if utilInfo.Utilization >= threshold {
		return false, nil
	}
	return true, nil
}

// UnremovableNodes returns a list of nodes that cannot be removed according to
// the scale down algorithm.
func (sd *ScaleDown) UnremovableNodes() []*simulator.UnremovableNode {
	return sd.unremovableNodes.AsList()
}

// markSimulationError indicates a simulation error by clearing  relevant scale
// down state and returning an appropriate error.
func (sd *ScaleDown) markSimulationError(simulatorErr errors.AutoscalerError,
	timestamp time.Time) errors.AutoscalerError {
	klog.Errorf("Error while simulating node drains: %v", simulatorErr)
	sd.unneededNodesList = make([]*apiv1.Node, 0)
	sd.unneededNodes = make(map[string]time.Time)
	sd.nodeUtilizationMap = make(map[string]utilization.Info)
	sd.clusterStateRegistry.UpdateScaleDownCandidates(sd.unneededNodesList, timestamp)
	return simulatorErr.AddPrefix("error while simulating node drains: ")
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
		if _, found := sd.unneededNodes[node]; found {
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
func (sd *ScaleDown) NodesToDelete(currentTime time.Time, pdbs []*policyv1.PodDisruptionBudget) (empty, drain []*apiv1.Node, res status.ScaleDownResult, err errors.AutoscalerError) {
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

	nodesWithoutMaster := filterOutMasters(allNodeInfos)
	nodesWithoutMasterNames := make([]string, 0, len(nodesWithoutMaster))
	for _, node := range nodesWithoutMaster {
		nodesWithoutMasterNames = append(nodesWithoutMasterNames, node.Name)
	}

	candidateNames := make([]string, 0)
	readinessMap := make(map[string]bool)
	candidateNodeGroups := make(map[string]cloudprovider.NodeGroup)

	resourceLimiter, errCP := sd.context.CloudProvider.GetResourceLimiter()
	if errCP != nil {
		return nil, nil, status.ScaleDownError, errors.ToAutoscalerError(errors.CloudProviderError, errCP)
	}

	scaleDownResourcesLeft := sd.computeScaleDownResourcesLeftLimits(nodesWithoutMaster, resourceLimiter, sd.context.CloudProvider, currentTime)

	nodeGroupSize := utils.GetNodeGroupSizeMap(sd.context.CloudProvider)
	resourcesWithLimits := resourceLimiter.GetResources()
	for nodeName, unneededSince := range sd.unneededNodes {
		klog.V(2).Infof("%s was unneeded for %s", nodeName, currentTime.Sub(unneededSince).String())

		nodeInfo, err := sd.context.ClusterSnapshot.NodeInfos().Get(nodeName)
		if err != nil {
			klog.Errorf("Can't retrieve unneeded node %s from snapshot, err: %v", nodeName, err)
			continue
		}

		node := nodeInfo.Node()

		// Check if node is marked with no scale down annotation.
		if hasNoScaleDownAnnotation(node) {
			klog.V(4).Infof("Skipping %s - scale down disabled annotation found", node.Name)
			sd.unremovableNodes.AddReason(node, simulator.ScaleDownDisabledAnnotation)
			continue
		}

		ready, _, _ := kube_util.GetReadinessState(node)
		readinessMap[node.Name] = ready

		nodeGroup, err := sd.context.CloudProvider.NodeGroupForNode(node)
		if err != nil {
			klog.Errorf("Error while checking node group for %s: %v", node.Name, err)
			sd.unremovableNodes.AddReason(node, simulator.UnexpectedError)
			continue
		}
		if nodeGroup == nil || reflect.ValueOf(nodeGroup).IsNil() {
			klog.V(4).Infof("Skipping %s - no node group config", node.Name)
			sd.unremovableNodes.AddReason(node, simulator.NotAutoscaled)
			continue
		}

		if ready {
			// Check how long a ready node was underutilized.
			unneededTime, err := sd.processors.NodeGroupConfigProcessor.GetScaleDownUnneededTime(sd.context, nodeGroup)
			if err != nil {
				klog.Errorf("Error trying to get ScaleDownUnneededTime for node %s (in group: %s)", node.Name, nodeGroup.Id())
				continue
			}
			if !unneededSince.Add(unneededTime).Before(currentTime) {
				sd.unremovableNodes.AddReason(node, simulator.NotUnneededLongEnough)
				continue
			}
		} else {
			// Unready nodes may be deleted after a different time than underutilized nodes.
			unreadyTime, err := sd.processors.NodeGroupConfigProcessor.GetScaleDownUnreadyTime(sd.context, nodeGroup)
			if err != nil {
				klog.Errorf("Error trying to get ScaleDownUnreadyTime for node %s (in group: %s)", node.Name, nodeGroup.Id())
				continue
			}
			if !unneededSince.Add(unreadyTime).Before(currentTime) {
				sd.unremovableNodes.AddReason(node, simulator.NotUnreadyLongEnough)
				continue
			}
		}

		size, found := nodeGroupSize[nodeGroup.Id()]
		if !found {
			klog.Errorf("Error while checking node group size %s: group size not found in cache", nodeGroup.Id())
			sd.unremovableNodes.AddReason(node, simulator.UnexpectedError)
			continue
		}

		deletionsInProgress := sd.nodeDeletionTracker.DeletionsCount(nodeGroup.Id())
		if size-deletionsInProgress <= nodeGroup.MinSize() {
			klog.V(1).Infof("Skipping %s - node group min size reached", node.Name)
			sd.unremovableNodes.AddReason(node, simulator.NodeGroupMinSizeReached)
			continue
		}

		scaleDownResourcesDelta, err := sd.computeScaleDownResourcesDelta(sd.context.CloudProvider, node, nodeGroup, resourcesWithLimits)
		if err != nil {
			klog.Errorf("Error getting node resources: %v", err)
			sd.unremovableNodes.AddReason(node, simulator.UnexpectedError)
			continue
		}

		checkResult := scaleDownResourcesLeft.checkScaleDownDeltaWithinLimits(scaleDownResourcesDelta)
		if checkResult.exceeded {
			klog.V(4).Infof("Skipping %s - minimal limit exceeded for %v", node.Name, checkResult.exceededResources)
			sd.unremovableNodes.AddReason(node, simulator.MinimalResourceLimitExceeded)
			continue
		}

		candidateNames = append(candidateNames, node.Name)
		candidateNodeGroups[node.Name] = nodeGroup
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
			simulator.RemoveNodeFromTracker(sd.usageTracker, node.Node.Name, sd.unneededNodes)
			nodes = append(nodes, node.Node)
		}
		return nodes, nil, status.ScaleDownNodeDeleteStarted, nil
	}

	findNodesToRemoveStart := time.Now()
	// We look for only 1 node so new hints may be incomplete.
	nodesToRemove, unremovable, _, err := sd.removalSimulator.FindNodesToRemove(
		candidateNames,
		nodesWithoutMasterNames,
		sd.podLocationHints,
		time.Now(),
		pdbs)
	findNodesToRemoveDuration = time.Now().Sub(findNodesToRemoveStart)

	for _, unremovableNode := range unremovable {
		sd.unremovableNodes.Add(unremovableNode)
	}
	if err != nil {
		return nil, nil, status.ScaleDownError, err.AddPrefix("Find node to remove failed: ")
	}

	nodesToRemove = sd.processors.ScaleDownSetProcessor.GetNodesToRemove(sd.context, nodesToRemove, 1)
	if len(nodesToRemove) == 0 {
		klog.V(1).Infof("No node to remove")
		return nil, nil, status.ScaleDownNoNodeDeleted, nil
	}
	toRemove := nodesToRemove[0]
	// Nothing super-bad should happen if the node is removed from tracker prematurely.
	simulator.RemoveNodeFromTracker(sd.usageTracker, toRemove.Node.Name, sd.unneededNodes)
	return nil, []*apiv1.Node{toRemove.Node}, status.ScaleDownNodeDeleteStarted, nil
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
	return sd.getEmptyNodesToRemove(candidates, noScaleDownLimitsOnResources(), timestamp)
}

// This functions finds empty nodes among passed candidates and returns a list of empty nodes
// that can be deleted at the same time.
func (sd *ScaleDown) getEmptyNodesToRemove(candidates []string, resourcesLimits scaleDownResourcesLimits,
	timestamp time.Time) []simulator.NodeToBeRemoved {

	emptyNodes := sd.removalSimulator.FindEmptyNodesToRemove(candidates, timestamp)
	availabilityMap := make(map[string]int)
	nodesToRemove := make([]simulator.NodeToBeRemoved, 0)
	resourcesLimitsCopy := copyScaleDownResourcesLimits(resourcesLimits) // we do not want to modify input parameter
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
			resourcesDelta, err := sd.computeScaleDownResourcesDelta(sd.context.CloudProvider, node, nodeGroup, resourcesNames)
			if err != nil {
				klog.Errorf("Error: %v", err)
				continue
			}
			checkResult := resourcesLimitsCopy.tryDecrementLimitsByDelta(resourcesDelta)
			if checkResult.exceeded {
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

func hasNoScaleDownAnnotation(node *apiv1.Node) bool {
	return node.Annotations[ScaleDownDisabledKey] == "true"
}

const (
	apiServerLabelKey   = "component"
	apiServerLabelValue = "kube-apiserver"
)

func isMasterNode(nodeInfo *schedulerframework.NodeInfo) bool {
	for _, podInfo := range nodeInfo.Pods {
		if podInfo.Pod.Namespace == metav1.NamespaceSystem && podInfo.Pod.Labels[apiServerLabelKey] == apiServerLabelValue {
			return true
		}
	}
	return false
}

func filterOutMasters(nodeInfos []*schedulerframework.NodeInfo) []*apiv1.Node {
	result := make([]*apiv1.Node, 0, len(nodeInfos))
	for _, nodeInfo := range nodeInfos {
		if !isMasterNode(nodeInfo) {
			result = append(result, nodeInfo.Node())
		}
	}
	return result
}
