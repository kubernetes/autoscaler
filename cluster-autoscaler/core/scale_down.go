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
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/metrics"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	"k8s.io/autoscaler/cluster-autoscaler/utils/deletetaint"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	scheduler_util "k8s.io/autoscaler/cluster-autoscaler/utils/scheduler"

	apiv1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1beta1"
	kube_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_client "k8s.io/client-go/kubernetes"
	kube_record "k8s.io/client-go/tools/record"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/autoscaler/cluster-autoscaler/processors/status"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	"k8s.io/klog"
)

const (
	// ScaleDownDisabledKey is the name of annotation marking node as not eligible for scale down.
	ScaleDownDisabledKey = "cluster-autoscaler.kubernetes.io/scale-down-disabled"
	// DelayDeletionAnnotationPrefix is the prefix of annotation marking node as it needs to wait
	// for other K8s components before deleting node.
	DelayDeletionAnnotationPrefix = "delay-deletion.cluster-autoscaler.kubernetes.io/"
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
)

// NodeDeletionTracker keeps track of node deletions.
type NodeDeletionTracker struct {
	sync.Mutex
	nonEmptyNodeDeleteInProgress bool
	// A map of node delete results by node name. It's being constantly emptied into ScaleDownStatus
	// objects in order to notify the ScaleDownStatusProcessor that the node drain has ended or that
	// an error occurred during the deletion process.
	nodeDeleteResults map[string]status.NodeDeleteResult
	// A map which keeps track of deletions in progress for nodepools.
	// Key is a node group id and value is a number of node deletions in progress.
	deletionsInProgress map[string]int
}

// Get current time. Proxy for unit tests.
var now func() time.Time = time.Now

// NewNodeDeletionTracker creates new NodeDeletionTracker.
func NewNodeDeletionTracker() *NodeDeletionTracker {
	return &NodeDeletionTracker{
		nodeDeleteResults:   make(map[string]status.NodeDeleteResult),
		deletionsInProgress: make(map[string]int),
	}
}

// IsNonEmptyNodeDeleteInProgress returns true if a non empty node is being deleted.
func (n *NodeDeletionTracker) IsNonEmptyNodeDeleteInProgress() bool {
	n.Lock()
	defer n.Unlock()
	return n.nonEmptyNodeDeleteInProgress
}

// SetNonEmptyNodeDeleteInProgress sets non empty node deletion in progress status.
func (n *NodeDeletionTracker) SetNonEmptyNodeDeleteInProgress(status bool) {
	n.Lock()
	defer n.Unlock()
	n.nonEmptyNodeDeleteInProgress = status
}

// StartDeletion increments node deletion in progress counter for the given nodegroup.
func (n *NodeDeletionTracker) StartDeletion(nodeGroupId string) {
	n.Lock()
	defer n.Unlock()
	n.deletionsInProgress[nodeGroupId]++
}

// EndDeletion decrements node deletion in progress counter for the given nodegroup.
func (n *NodeDeletionTracker) EndDeletion(nodeGroupId string) {
	n.Lock()
	defer n.Unlock()

	value, found := n.deletionsInProgress[nodeGroupId]
	if !found {
		klog.Errorf("This should never happen, counter for %s in DelayedNodeDeletionStatus wasn't found", nodeGroupId)
		return
	}
	if value <= 0 {
		klog.Errorf("This should never happen, counter for %s in DelayedNodeDeletionStatus isn't greater than 0, counter value is %d", nodeGroupId, value)
	}
	n.deletionsInProgress[nodeGroupId]--
	if n.deletionsInProgress[nodeGroupId] <= 0 {
		delete(n.deletionsInProgress, nodeGroupId)
	}
}

// GetDeletionsInProgress returns the number of deletions in progress for the given node group.
func (n *NodeDeletionTracker) GetDeletionsInProgress(nodeGroupId string) int {
	n.Lock()
	defer n.Unlock()
	return n.deletionsInProgress[nodeGroupId]
}

// AddNodeDeleteResult adds a node delete result to the result map.
func (n *NodeDeletionTracker) AddNodeDeleteResult(nodeName string, result status.NodeDeleteResult) {
	n.Lock()
	defer n.Unlock()
	n.nodeDeleteResults[nodeName] = result
}

// GetAndClearNodeDeleteResults returns the whole result map and replaces it with a new empty one.
func (n *NodeDeletionTracker) GetAndClearNodeDeleteResults() map[string]status.NodeDeleteResult {
	n.Lock()
	defer n.Unlock()
	results := n.nodeDeleteResults
	n.nodeDeleteResults = make(map[string]status.NodeDeleteResult)
	return results
}

type scaleDownResourcesLimits map[string]int64
type scaleDownResourcesDelta map[string]int64

// used as a value in scaleDownResourcesLimits if actual limit could not be obtained due to errors talking to cloud provider
const scaleDownLimitUnknown = math.MinInt64

func computeScaleDownResourcesLeftLimits(nodes []*apiv1.Node, resourceLimiter *cloudprovider.ResourceLimiter, cp cloudprovider.CloudProvider, timestamp time.Time) scaleDownResourcesLimits {
	totalCores, totalMem := calculateScaleDownCoresMemoryTotal(nodes, timestamp)

	var totalGpus map[string]int64
	var totalGpusErr error
	if cloudprovider.ContainsGpuResources(resourceLimiter.GetResources()) {
		totalGpus, totalGpusErr = calculateScaleDownGpusTotal(nodes, cp, timestamp)
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
			case cloudprovider.IsGpuResource(resource):
				if totalGpusErr != nil {
					resultScaleDownLimits[resource] = scaleDownLimitUnknown
				} else {
					resultScaleDownLimits[resource] = computeAboveMin(totalGpus[resource], min)
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
		if isNodeBeingDeleted(node, timestamp) {
			// Nodes being deleted do not count towards total cluster resources
			continue
		}
		cores, memory := getNodeCoresAndMemory(node)

		coresTotal += cores
		memoryTotal += memory
	}

	return coresTotal, memoryTotal
}

func calculateScaleDownGpusTotal(nodes []*apiv1.Node, cp cloudprovider.CloudProvider, timestamp time.Time) (map[string]int64, error) {
	type gpuInfo struct {
		name  string
		count int64
	}

	result := make(map[string]int64)
	ngCache := make(map[string]gpuInfo)
	for _, node := range nodes {
		if isNodeBeingDeleted(node, timestamp) {
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

		var gpuType string
		var gpuCount int64

		var cached gpuInfo
		var cacheHit bool
		if nodeGroup != nil {
			cached, cacheHit = ngCache[nodeGroup.Id()]
			if cacheHit {
				gpuType = cached.name
				gpuCount = cached.count
			}
		}
		if !cacheHit {
			gpuType, gpuCount, err = gpu.GetNodeTargetGpus(cp.GPULabel(), node, nodeGroup)
			if err != nil {
				return nil, errors.ToAutoscalerError(errors.CloudProviderError, err).AddPrefix("can not get gpu count for node %v when calculating cluster gpu usage")
			}
			if nodeGroup != nil {
				ngCache[nodeGroup.Id()] = gpuInfo{name: gpuType, count: gpuCount}
			}
		}
		if gpuType == "" || gpuCount == 0 {
			continue
		}
		result[gpuType] += gpuCount
	}

	return result, nil
}

func isNodeBeingDeleted(node *apiv1.Node, timestamp time.Time) bool {
	deleteTime, _ := deletetaint.GetToBeDeletedTime(node)
	return deleteTime != nil && (timestamp.Sub(*deleteTime) < MaxCloudProviderNodeDeletionTime || timestamp.Sub(*deleteTime) < MaxKubernetesEmptyNodeDeletionTime)
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

func computeScaleDownResourcesDelta(cp cloudprovider.CloudProvider, node *apiv1.Node, nodeGroup cloudprovider.NodeGroup, resourcesWithLimits []string) (scaleDownResourcesDelta, errors.AutoscalerError) {
	resultScaleDownDelta := make(scaleDownResourcesDelta)

	nodeCPU, nodeMemory := getNodeCoresAndMemory(node)
	resultScaleDownDelta[cloudprovider.ResourceNameCores] = nodeCPU
	resultScaleDownDelta[cloudprovider.ResourceNameMemory] = nodeMemory

	if cloudprovider.ContainsGpuResources(resourcesWithLimits) {
		gpuType, gpuCount, err := gpu.GetNodeTargetGpus(cp.GPULabel(), node, nodeGroup)
		if err != nil {
			return scaleDownResourcesDelta{}, errors.ToAutoscalerError(errors.CloudProviderError, err).AddPrefix("Failed to get node %v gpu: %v", node.Name)
		}
		resultScaleDownDelta[gpuType] = gpuCount
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
	clusterStateRegistry *clusterstate.ClusterStateRegistry
	unneededNodes        map[string]time.Time
	unneededNodesList    []*apiv1.Node
	unremovableNodes     map[string]time.Time
	podLocationHints     map[string]string
	nodeUtilizationMap   map[string]simulator.UtilizationInfo
	usageTracker         *simulator.UsageTracker
	nodeDeletionTracker  *NodeDeletionTracker
}

// NewScaleDown builds new ScaleDown object.
func NewScaleDown(context *context.AutoscalingContext, clusterStateRegistry *clusterstate.ClusterStateRegistry) *ScaleDown {
	return &ScaleDown{
		context:              context,
		clusterStateRegistry: clusterStateRegistry,
		unneededNodes:        make(map[string]time.Time),
		unremovableNodes:     make(map[string]time.Time),
		podLocationHints:     make(map[string]string),
		nodeUtilizationMap:   make(map[string]simulator.UtilizationInfo),
		usageTracker:         simulator.NewUsageTracker(),
		unneededNodesList:    make([]*apiv1.Node, 0),
		nodeDeletionTracker:  NewNodeDeletionTracker(),
	}
}

// CleanUp cleans up the internal ScaleDown state.
func (sd *ScaleDown) CleanUp(timestamp time.Time) {
	sd.usageTracker.CleanUp(timestamp.Add(-sd.context.ScaleDownUnneededTime))
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
	// Only scheduled non expendable pods and pods waiting for lower priority pods preemption can prevent node delete.
	nonExpendablePods := filterOutExpendablePods(pods, sd.context.ExpendablePodsPriorityCutoff)
	nodeNameToNodeInfo := scheduler_util.CreateNodeNameToInfoMap(nonExpendablePods, nodes)
	utilizationMap := make(map[string]simulator.UtilizationInfo)

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
		klog.V(1).Infof("Scale-down calculation: ignoring %v nodes unremovable in the last %v", skipped, sd.context.AutoscalingOptions.UnremovableNodeRecheckTimeout)
	}

	// Phase1 - look at the nodes utilization. Calculate the utilization
	// only for the managed nodes.
	for _, node := range filteredNodesToCheck {

		// Skip nodes marked to be deleted, if they were marked recently.
		// Old-time marked nodes are again eligible for deletion - something went wrong with them
		// and they have not been deleted.
		if isNodeBeingDeleted(node, timestamp) {
			klog.V(1).Infof("Skipping %s from delete considerations - the node is currently being deleted", node.Name)
			continue
		}

		// Skip nodes marked with no scale down annotation
		if hasNoScaleDownAnnotation(node) {
			klog.V(1).Infof("Skipping %s from delete consideration - the node is marked as no scale down", node.Name)
			continue
		}

		nodeInfo, found := nodeNameToNodeInfo[node.Name]
		if !found {
			klog.Errorf("Node info for %s not found", node.Name)
			continue
		}

		utilInfo, err := simulator.CalculateUtilization(node, nodeInfo, sd.context.IgnoreDaemonSetsUtilization, sd.context.IgnoreMirrorPodsUtilization, sd.context.CloudProvider.GPULabel())
		if err != nil {
			klog.Warningf("Failed to calculate utilization for %s: %v", node.Name, err)
		}
		klog.V(4).Infof("Node %s - %s utilization %f", node.Name, utilInfo.ResourceName, utilInfo.Utilization)
		utilizationMap[node.Name] = utilInfo

		if !sd.isNodeBelowUtilzationThreshold(node, utilInfo) {
			klog.V(4).Infof("Node %s is not suitable for removal - %s utilization too big (%f)", node.Name, utilInfo.ResourceName, utilInfo.Utilization)
			continue
		}
		currentlyUnneededNodes = append(currentlyUnneededNodes, node)
	}

	emptyNodes := make(map[string]bool)

	emptyNodesList := sd.getEmptyNodesNoResourceLimits(currentlyUnneededNodes, pods, len(currentlyUnneededNodes))
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
		currentCandidates, nodes, nonExpendablePods, nil, sd.context.PredicateChecker,
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
		// Look for additional nodes to remove among the rest of nodes.
		klog.V(3).Infof("Finding additional %v candidates for scale down.", additionalCandidatesCount)
		additionalNodesToRemove, additionalUnremovable, additionalNewHints, simulatorErr :=
			simulator.FindNodesToRemove(currentNonCandidates[:additionalCandidatesPoolSize], nodes, nonExpendablePods, nil,
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
		unremovableTimeout := timestamp.Add(sd.context.AutoscalingOptions.UnremovableNodeRecheckTimeout)
		for _, node := range unremovable {
			sd.unremovableNodes[node.Name] = unremovableTimeout
		}
		klog.V(1).Infof("%v nodes found to be unremovable in simulation, will re-check them at %v", len(unremovable), unremovableTimeout)
	}

	// Update state and metrics
	sd.unneededNodesList = unneededNodesList
	sd.unneededNodes = result
	sd.podLocationHints = newHints
	sd.nodeUtilizationMap = utilizationMap
	sd.clusterStateRegistry.UpdateScaleDownCandidates(sd.unneededNodesList, timestamp)
	metrics.UpdateUnneededNodesCount(len(sd.unneededNodesList))
	return nil
}

// isNodeBelowUtilzationThreshold determintes if a given node utilization is blow threshold.
func (sd *ScaleDown) isNodeBelowUtilzationThreshold(node *apiv1.Node, utilInfo simulator.UtilizationInfo) bool {
	if gpu.NodeHasGpu(sd.context.CloudProvider.GPULabel(), node) {
		if utilInfo.Utilization >= sd.context.ScaleDownGpuUtilizationThreshold {
			return false
		}
	} else {
		if utilInfo.Utilization >= sd.context.ScaleDownUtilizationThreshold {
			return false
		}
	}
	return true
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
// down state and returning an appropriate error.
func (sd *ScaleDown) markSimulationError(simulatorErr errors.AutoscalerError,
	timestamp time.Time) errors.AutoscalerError {
	klog.Errorf("Error while simulating node drains: %v", simulatorErr)
	sd.unneededNodesList = make([]*apiv1.Node, 0)
	sd.unneededNodes = make(map[string]time.Time)
	sd.nodeUtilizationMap = make(map[string]simulator.UtilizationInfo)
	sd.clusterStateRegistry.UpdateScaleDownCandidates(sd.unneededNodesList, timestamp)
	return simulatorErr.AddPrefix("error while simulating node drains: ")
}

// chooseCandidates splits nodes into current candidates for scale-down and the
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

// SoftTaintUnneededNodes manage soft taints of unneeded nodes.
func (sd *ScaleDown) SoftTaintUnneededNodes(allNodes []*apiv1.Node) (errors []error) {
	defer metrics.UpdateDurationFromStart(metrics.ScaleDownSoftTaintUnneeded, time.Now())
	apiCallBudget := sd.context.AutoscalingOptions.MaxBulkSoftTaintCount
	timeBudget := sd.context.AutoscalingOptions.MaxBulkSoftTaintTime
	skippedNodes := 0
	startTime := now()
	for _, node := range allNodes {
		if deletetaint.HasToBeDeletedTaint(node) {
			// Do not consider nodes that are scheduled to be deleted
			continue
		}
		alreadyTainted := deletetaint.HasDeletionCandidateTaint(node)
		_, unneeded := sd.unneededNodes[node.Name]

		// Check if expected taints match existing taints
		if unneeded != alreadyTainted {
			if apiCallBudget <= 0 || now().Sub(startTime) >= timeBudget {
				skippedNodes++
				continue
			}
			apiCallBudget--
			if unneeded && !alreadyTainted {
				err := deletetaint.MarkDeletionCandidate(node, sd.context.ClientSet)
				if err != nil {
					errors = append(errors, err)
					klog.Warningf("Soft taint on %s adding error %v", node.Name, err)
				}
			}
			if !unneeded && alreadyTainted {
				_, err := deletetaint.CleanDeletionCandidate(node, sd.context.ClientSet)
				if err != nil {
					errors = append(errors, err)
					klog.Warningf("Soft taint on %s removal error %v", node.Name, err)
				}
			}
		}
	}
	if skippedNodes > 0 {
		klog.V(4).Infof("Skipped adding/removing soft taints on %v nodes - API call limit exceeded", skippedNodes)
	}
	return
}

// TryToScaleDown tries to scale down the cluster. It returns a result inside a ScaleDownStatus indicating if any node was
// removed and error if such occurred.
func (sd *ScaleDown) TryToScaleDown(allNodes []*apiv1.Node, pods []*apiv1.Pod, pdbs []*policyv1.PodDisruptionBudget, currentTime time.Time) (*status.ScaleDownStatus, errors.AutoscalerError) {
	scaleDownStatus := &status.ScaleDownStatus{NodeDeleteResults: sd.nodeDeletionTracker.GetAndClearNodeDeleteResults()}
	nodeDeletionDuration := time.Duration(0)
	findNodesToRemoveDuration := time.Duration(0)
	defer updateScaleDownMetrics(time.Now(), &findNodesToRemoveDuration, &nodeDeletionDuration)
	nodesWithoutMaster := filterOutMasters(allNodes, pods)
	candidates := make([]*apiv1.Node, 0)
	readinessMap := make(map[string]bool)
	candidateNodeGroups := make(map[string]cloudprovider.NodeGroup)
	gpuLabel := sd.context.CloudProvider.GPULabel()
	availableGPUTypes := sd.context.CloudProvider.GetAvailableGPUTypes()

	resourceLimiter, errCP := sd.context.CloudProvider.GetResourceLimiter()
	if errCP != nil {
		scaleDownStatus.Result = status.ScaleDownError
		return scaleDownStatus, errors.ToAutoscalerError(errors.CloudProviderError, errCP)
	}

	scaleDownResourcesLeft := computeScaleDownResourcesLeftLimits(nodesWithoutMaster, resourceLimiter, sd.context.CloudProvider, currentTime)

	nodeGroupSize := getNodeGroupSizeMap(sd.context.CloudProvider)
	resourcesWithLimits := resourceLimiter.GetResources()
	for _, node := range nodesWithoutMaster {
		if val, found := sd.unneededNodes[node.Name]; found {

			klog.V(2).Infof("%s was unneeded for %s", node.Name, currentTime.Sub(val).String())

			// Check if node is marked with no scale down annotation.
			if hasNoScaleDownAnnotation(node) {
				klog.V(4).Infof("Skipping %s - scale down disabled annotation found", node.Name)
				continue
			}

			ready, _, _ := kube_util.GetReadinessState(node)
			readinessMap[node.Name] = ready

			// Check how long the node was underutilized.
			if ready && !val.Add(sd.context.ScaleDownUnneededTime).Before(currentTime) {
				continue
			}

			// Unready nodes may be deleted after a different time than underutilized nodes.
			if !ready && !val.Add(sd.context.ScaleDownUnreadyTime).Before(currentTime) {
				continue
			}

			nodeGroup, err := sd.context.CloudProvider.NodeGroupForNode(node)
			if err != nil {
				klog.Errorf("Error while checking node group for %s: %v", node.Name, err)
				continue
			}
			if nodeGroup == nil || reflect.ValueOf(nodeGroup).IsNil() {
				klog.V(4).Infof("Skipping %s - no node group config", node.Name)
				continue
			}

			size, found := nodeGroupSize[nodeGroup.Id()]
			if !found {
				klog.Errorf("Error while checking node group size %s: group size not found in cache", nodeGroup.Id())
				continue
			}

			if size-sd.nodeDeletionTracker.GetDeletionsInProgress(nodeGroup.Id()) <= nodeGroup.MinSize() {
				klog.V(1).Infof("Skipping %s - node group min size reached", node.Name)
				continue
			}

			scaleDownResourcesDelta, err := computeScaleDownResourcesDelta(sd.context.CloudProvider, node, nodeGroup, resourcesWithLimits)
			if err != nil {
				klog.Errorf("Error getting node resources: %v", err)
				continue
			}

			checkResult := scaleDownResourcesLeft.checkScaleDownDeltaWithinLimits(scaleDownResourcesDelta)
			if checkResult.exceeded {
				klog.V(4).Infof("Skipping %s - minimal limit exceeded for %v", node.Name, checkResult.exceededResources)
				continue
			}

			candidates = append(candidates, node)
			candidateNodeGroups[node.Name] = nodeGroup
		}
	}
	if len(candidates) == 0 {
		klog.V(1).Infof("No candidates for scale down")
		scaleDownStatus.Result = status.ScaleDownNoUnneeded
		return scaleDownStatus, nil
	}

	// Trying to delete empty nodes in bulk. If there are no empty nodes then CA will
	// try to delete not-so-empty nodes, possibly killing some pods and allowing them
	// to recreate on other nodes.
	emptyNodes := sd.getEmptyNodes(candidates, pods, sd.context.MaxEmptyBulkDelete, scaleDownResourcesLeft)
	if len(emptyNodes) > 0 {
		nodeDeletionStart := time.Now()
		deletedNodes, err := sd.scheduleDeleteEmptyNodes(emptyNodes, sd.context.ClientSet, sd.context.Recorder, readinessMap, candidateNodeGroups)
		nodeDeletionDuration = time.Now().Sub(nodeDeletionStart)

		// TODO: Give the processor some information about the nodes that failed to be deleted.
		scaleDownStatus.ScaledDownNodes = sd.mapNodesToStatusScaleDownNodes(deletedNodes, candidateNodeGroups, make(map[string][]*apiv1.Pod))
		if len(deletedNodes) > 0 {
			scaleDownStatus.Result = status.ScaleDownNodeDeleteStarted
		} else {
			scaleDownStatus.Result = status.ScaleDownError
		}
		if err != nil {
			return scaleDownStatus, err.AddPrefix("failed to delete at least one empty node: ")
		}
		return scaleDownStatus, nil
	}

	findNodesToRemoveStart := time.Now()
	// Only scheduled non expendable pods are taken into account and have to be moved.
	nonExpendablePods := filterOutExpendablePods(pods, sd.context.ExpendablePodsPriorityCutoff)
	// We look for only 1 node so new hints may be incomplete.
	nodesToRemove, _, _, err := simulator.FindNodesToRemove(candidates, nodesWithoutMaster, nonExpendablePods, sd.context.ListerRegistry,
		sd.context.PredicateChecker, 1, false,
		sd.podLocationHints, sd.usageTracker, time.Now(), pdbs)
	findNodesToRemoveDuration = time.Now().Sub(findNodesToRemoveStart)

	if err != nil {
		scaleDownStatus.Result = status.ScaleDownError
		return scaleDownStatus, err.AddPrefix("Find node to remove failed: ")
	}
	if len(nodesToRemove) == 0 {
		klog.V(1).Infof("No node to remove")
		scaleDownStatus.Result = status.ScaleDownNoNodeDeleted
		return scaleDownStatus, nil
	}
	toRemove := nodesToRemove[0]
	utilization := sd.nodeUtilizationMap[toRemove.Node.Name]
	podNames := make([]string, 0, len(toRemove.PodsToReschedule))
	for _, pod := range toRemove.PodsToReschedule {
		podNames = append(podNames, pod.Namespace+"/"+pod.Name)
	}
	klog.V(0).Infof("Scale-down: removing node %s, utilization: %v, pods to reschedule: %s", toRemove.Node.Name, utilization,
		strings.Join(podNames, ","))
	sd.context.LogRecorder.Eventf(apiv1.EventTypeNormal, "ScaleDown", "Scale-down: removing node %s, utilization: %v, pods to reschedule: %s",
		toRemove.Node.Name, utilization, strings.Join(podNames, ","))

	// Nothing super-bad should happen if the node is removed from tracker prematurely.
	simulator.RemoveNodeFromTracker(sd.usageTracker, toRemove.Node.Name, sd.unneededNodes)
	nodeDeletionStart := time.Now()

	// Starting deletion.
	nodeDeletionDuration = time.Now().Sub(nodeDeletionStart)
	sd.nodeDeletionTracker.SetNonEmptyNodeDeleteInProgress(true)

	go func() {
		// Finishing the delete process once this goroutine is over.
		var result status.NodeDeleteResult
		defer func() { sd.nodeDeletionTracker.AddNodeDeleteResult(toRemove.Node.Name, result) }()
		defer sd.nodeDeletionTracker.SetNonEmptyNodeDeleteInProgress(false)
		nodeGroup, found := candidateNodeGroups[toRemove.Node.Name]
		if !found {
			result = status.NodeDeleteResult{ResultType: status.NodeDeleteErrorFailedToDelete, Err: errors.NewAutoscalerError(
				errors.InternalError, "failed to find node group for %s", toRemove.Node.Name)}
			return
		}
		result = sd.deleteNode(toRemove.Node, toRemove.PodsToReschedule, nodeGroup)
		if result.ResultType != status.NodeDeleteOk {
			klog.Errorf("Failed to delete %s: %v", toRemove.Node.Name, result.Err)
			return
		}
		if readinessMap[toRemove.Node.Name] {
			metrics.RegisterScaleDown(1, gpu.GetGpuTypeForMetrics(gpuLabel, availableGPUTypes, toRemove.Node, nodeGroup), metrics.Underutilized)
		} else {
			metrics.RegisterScaleDown(1, gpu.GetGpuTypeForMetrics(gpuLabel, availableGPUTypes, toRemove.Node, nodeGroup), metrics.Unready)
		}
	}()

	scaleDownStatus.ScaledDownNodes = sd.mapNodesToStatusScaleDownNodes([]*apiv1.Node{toRemove.Node}, candidateNodeGroups, map[string][]*apiv1.Pod{toRemove.Node.Name: toRemove.PodsToReschedule})
	scaleDownStatus.Result = status.ScaleDownNodeDeleteStarted
	return scaleDownStatus, nil
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

func (sd *ScaleDown) getEmptyNodesNoResourceLimits(candidates []*apiv1.Node, pods []*apiv1.Pod, maxEmptyBulkDelete int) []*apiv1.Node {
	return sd.getEmptyNodes(candidates, pods, maxEmptyBulkDelete, noScaleDownLimitsOnResources())
}

// This functions finds empty nodes among passed candidates and returns a list of empty nodes
// that can be deleted at the same time.
func (sd *ScaleDown) getEmptyNodes(candidates []*apiv1.Node, pods []*apiv1.Pod, maxEmptyBulkDelete int,
	resourcesLimits scaleDownResourcesLimits) []*apiv1.Node {

	emptyNodes := simulator.FindEmptyNodesToRemove(candidates, pods)
	availabilityMap := make(map[string]int)
	result := make([]*apiv1.Node, 0)
	resourcesLimitsCopy := copyScaleDownResourcesLimits(resourcesLimits) // we do not want to modify input parameter
	resourcesNames := sets.StringKeySet(resourcesLimits).List()
	for _, node := range emptyNodes {
		nodeGroup, err := sd.context.CloudProvider.NodeGroupForNode(node)
		if err != nil {
			klog.Errorf("Failed to get group for %s", node.Name)
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
			available = size - nodeGroup.MinSize() - sd.nodeDeletionTracker.GetDeletionsInProgress(nodeGroup.Id())
			if available < 0 {
				available = 0
			}
			availabilityMap[nodeGroup.Id()] = available
		}
		if available > 0 {
			resourcesDelta, err := computeScaleDownResourcesDelta(sd.context.CloudProvider, node, nodeGroup, resourcesNames)
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
			result = append(result, node)
		}
	}
	limit := maxEmptyBulkDelete
	if len(result) < limit {
		limit = len(result)
	}
	return result[:limit]
}

func (sd *ScaleDown) scheduleDeleteEmptyNodes(emptyNodes []*apiv1.Node, client kube_client.Interface,
	recorder kube_record.EventRecorder, readinessMap map[string]bool,
	candidateNodeGroups map[string]cloudprovider.NodeGroup) ([]*apiv1.Node, errors.AutoscalerError) {
	deletedNodes := []*apiv1.Node{}
	for _, node := range emptyNodes {
		klog.V(0).Infof("Scale-down: removing empty node %s", node.Name)
		sd.context.LogRecorder.Eventf(apiv1.EventTypeNormal, "ScaleDownEmpty", "Scale-down: removing empty node %s", node.Name)
		simulator.RemoveNodeFromTracker(sd.usageTracker, node.Name, sd.unneededNodes)
		nodeGroup, found := candidateNodeGroups[node.Name]
		if !found {
			return deletedNodes, errors.NewAutoscalerError(
				errors.CloudProviderError, "failed to find node group for %s", node.Name)
		}
		taintErr := deletetaint.MarkToBeDeleted(node, client)
		if taintErr != nil {
			recorder.Eventf(node, apiv1.EventTypeWarning, "ScaleDownFailed", "failed to mark the node as toBeDeleted/unschedulable: %v", taintErr)
			return deletedNodes, errors.ToAutoscalerError(errors.ApiCallError, taintErr)
		}
		deletedNodes = append(deletedNodes, node)
		go func(nodeToDelete *apiv1.Node, nodeGroupForDeletedNode cloudprovider.NodeGroup) {
			sd.nodeDeletionTracker.StartDeletion(nodeGroupForDeletedNode.Id())
			defer sd.nodeDeletionTracker.EndDeletion(nodeGroupForDeletedNode.Id())
			var result status.NodeDeleteResult
			defer func() { sd.nodeDeletionTracker.AddNodeDeleteResult(nodeToDelete.Name, result) }()

			var deleteErr errors.AutoscalerError
			// If we fail to delete the node we want to remove delete taint
			defer func() {
				if deleteErr != nil {
					deletetaint.CleanToBeDeleted(nodeToDelete, client)
					recorder.Eventf(nodeToDelete, apiv1.EventTypeWarning, "ScaleDownFailed", "failed to delete empty node: %v", deleteErr)
				} else {
					sd.context.LogRecorder.Eventf(apiv1.EventTypeNormal, "ScaleDownEmpty", "Scale-down: empty node %s removed", nodeToDelete.Name)
				}
			}()

			deleteErr = waitForDelayDeletion(nodeToDelete, sd.context.ListerRegistry.AllNodeLister(), sd.context.AutoscalingOptions.NodeDeletionDelayTimeout)
			if deleteErr != nil {
				klog.Errorf("Problem with empty node deletion: %v", deleteErr)
				result = status.NodeDeleteResult{ResultType: status.NodeDeleteErrorFailedToDelete, Err: deleteErr}
				return
			}
			deleteErr = deleteNodeFromCloudProvider(nodeToDelete, sd.context.CloudProvider,
				sd.context.Recorder, sd.clusterStateRegistry)
			if deleteErr != nil {
				klog.Errorf("Problem with empty node deletion: %v", deleteErr)
				result = status.NodeDeleteResult{ResultType: status.NodeDeleteErrorFailedToDelete, Err: deleteErr}
				return
			}
			if readinessMap[nodeToDelete.Name] {
				metrics.RegisterScaleDown(1, gpu.GetGpuTypeForMetrics(sd.context.CloudProvider.GPULabel(), sd.context.CloudProvider.GetAvailableGPUTypes(), nodeToDelete, nodeGroupForDeletedNode), metrics.Empty)
			} else {
				metrics.RegisterScaleDown(1, gpu.GetGpuTypeForMetrics(sd.context.CloudProvider.GPULabel(), sd.context.CloudProvider.GetAvailableGPUTypes(), nodeToDelete, nodeGroupForDeletedNode), metrics.Unready)
			}
			result = status.NodeDeleteResult{ResultType: status.NodeDeleteOk}
		}(node, nodeGroup)
	}
	return deletedNodes, nil
}

func (sd *ScaleDown) deleteNode(node *apiv1.Node, pods []*apiv1.Pod,
	nodeGroup cloudprovider.NodeGroup) status.NodeDeleteResult {
	deleteSuccessful := false
	drainSuccessful := false

	if err := deletetaint.MarkToBeDeleted(node, sd.context.ClientSet); err != nil {
		sd.context.Recorder.Eventf(node, apiv1.EventTypeWarning, "ScaleDownFailed", "failed to mark the node as toBeDeleted/unschedulable: %v", err)
		return status.NodeDeleteResult{ResultType: status.NodeDeleteErrorFailedToMarkToBeDeleted, Err: errors.ToAutoscalerError(errors.ApiCallError, err)}
	}

	sd.nodeDeletionTracker.StartDeletion(nodeGroup.Id())
	defer sd.nodeDeletionTracker.EndDeletion(nodeGroup.Id())

	// If we fail to evict all the pods from the node we want to remove delete taint
	defer func() {
		if !deleteSuccessful {
			deletetaint.CleanToBeDeleted(node, sd.context.ClientSet)
			if !drainSuccessful {
				sd.context.Recorder.Eventf(node, apiv1.EventTypeWarning, "ScaleDownFailed", "failed to drain the node, aborting ScaleDown")
			} else {
				sd.context.Recorder.Eventf(node, apiv1.EventTypeWarning, "ScaleDownFailed", "failed to delete the node")
			}
		}
	}()

	sd.context.Recorder.Eventf(node, apiv1.EventTypeNormal, "ScaleDown", "marked the node as toBeDeleted/unschedulable")

	// attempt drain
	evictionResults, err := drainNode(node, pods, sd.context.ClientSet, sd.context.Recorder, sd.context.MaxGracefulTerminationSec, MaxPodEvictionTime, EvictionRetryTime, PodEvictionHeadroom)
	if err != nil {
		return status.NodeDeleteResult{ResultType: status.NodeDeleteErrorFailedToEvictPods, Err: err, PodEvictionResults: evictionResults}
	}
	drainSuccessful = true

	if typedErr := waitForDelayDeletion(node, sd.context.ListerRegistry.AllNodeLister(), sd.context.AutoscalingOptions.NodeDeletionDelayTimeout); typedErr != nil {
		return status.NodeDeleteResult{ResultType: status.NodeDeleteErrorFailedToDelete, Err: typedErr}
	}

	// attempt delete from cloud provider

	if typedErr := deleteNodeFromCloudProvider(node, sd.context.CloudProvider, sd.context.Recorder, sd.clusterStateRegistry); typedErr != nil {
		return status.NodeDeleteResult{ResultType: status.NodeDeleteErrorFailedToDelete, Err: typedErr}
	}

	deleteSuccessful = true // Let the deferred function know there is no need to cleanup
	return status.NodeDeleteResult{ResultType: status.NodeDeleteOk}
}

func evictPod(podToEvict *apiv1.Pod, client kube_client.Interface, recorder kube_record.EventRecorder,
	maxGracefulTerminationSec int, retryUntil time.Time, waitBetweenRetries time.Duration) status.PodEvictionResult {
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
			return status.PodEvictionResult{Pod: podToEvict, TimedOut: false, Err: nil}
		}
	}
	klog.Errorf("Failed to evict pod %s, error: %v", podToEvict.Name, lastError)
	recorder.Eventf(podToEvict, apiv1.EventTypeWarning, "ScaleDownFailed", "failed to delete pod for ScaleDown")
	return status.PodEvictionResult{Pod: podToEvict, TimedOut: true, Err: fmt.Errorf("failed to evict pod %s/%s within allowed timeout (last error: %v)", podToEvict.Namespace, podToEvict.Name, lastError)}
}

// Performs drain logic on the node. Marks the node as unschedulable and later removes all pods, giving
// them up to MaxGracefulTerminationTime to finish.
func drainNode(node *apiv1.Node, pods []*apiv1.Pod, client kube_client.Interface, recorder kube_record.EventRecorder,
	maxGracefulTerminationSec int, maxPodEvictionTime time.Duration, waitBetweenRetries time.Duration,
	podEvictionHeadroom time.Duration) (evictionResults map[string]status.PodEvictionResult, err error) {

	evictionResults = make(map[string]status.PodEvictionResult)
	toEvict := len(pods)
	retryUntil := time.Now().Add(maxPodEvictionTime)
	confirmations := make(chan status.PodEvictionResult, toEvict)
	for _, pod := range pods {
		evictionResults[pod.Name] = status.PodEvictionResult{Pod: pod, TimedOut: true, Err: nil}
		go func(podToEvict *apiv1.Pod) {
			confirmations <- evictPod(podToEvict, client, recorder, maxGracefulTerminationSec, retryUntil, waitBetweenRetries)
		}(pod)
	}

	for range pods {
		select {
		case evictionResult := <-confirmations:
			evictionResults[evictionResult.Pod.Name] = evictionResult
			if evictionResult.WasEvictionSuccessful() {
				metrics.RegisterEvictions(1)
			}
		case <-time.After(retryUntil.Sub(time.Now()) + 5*time.Second):
			// All pods initially had results with TimedOut set to true, so the ones that didn't receive an actual result are correctly marked as timed out.
			return evictionResults, errors.NewAutoscalerError(errors.ApiCallError, "Failed to drain node %s/%s: timeout when waiting for creating evictions", node.Namespace, node.Name)
		}
	}

	evictionErrs := make([]error, 0)
	for _, result := range evictionResults {
		if !result.WasEvictionSuccessful() {
			evictionErrs = append(evictionErrs, result.Err)
		}
	}
	if len(evictionErrs) != 0 {
		return evictionResults, errors.NewAutoscalerError(errors.ApiCallError, "Failed to drain node %s/%s, due to following errors: %v", node.Namespace, node.Name, evictionErrs)
	}

	// Evictions created successfully, wait maxGracefulTerminationSec + podEvictionHeadroom to see if pods really disappeared.
	allGone := true
	for start := time.Now(); time.Now().Sub(start) < time.Duration(maxGracefulTerminationSec)*time.Second+podEvictionHeadroom; time.Sleep(5 * time.Second) {
		allGone = true
		for _, pod := range pods {
			podreturned, err := client.CoreV1().Pods(pod.Namespace).Get(pod.Name, metav1.GetOptions{})
			if err == nil && (podreturned == nil || podreturned.Spec.NodeName == node.Name) {
				klog.Errorf("Not deleted yet %v", podreturned)
				allGone = false
				break
			}
			if err != nil && !kube_errors.IsNotFound(err) {
				klog.Errorf("Failed to check pod %s/%s: %v", pod.Namespace, pod.Name, err)
				allGone = false
				break
			}
		}
		if allGone {
			klog.V(1).Infof("All pods removed from %s", node.Name)
			// Let the deferred function know there is no need for cleanup
			return evictionResults, nil
		}
	}

	for _, pod := range pods {
		podReturned, err := client.CoreV1().Pods(pod.Namespace).Get(pod.Name, metav1.GetOptions{})
		if err == nil && (podReturned == nil || podReturned.Spec.NodeName == node.Name) {
			evictionResults[pod.Name] = status.PodEvictionResult{Pod: pod, TimedOut: true, Err: nil}
		} else if err != nil && !kube_errors.IsNotFound(err) {
			evictionResults[pod.Name] = status.PodEvictionResult{Pod: pod, TimedOut: true, Err: err}
		} else {
			evictionResults[pod.Name] = status.PodEvictionResult{Pod: pod, TimedOut: false, Err: nil}
		}
	}

	return evictionResults, errors.NewAutoscalerError(errors.TransientError, "Failed to drain node %s/%s: pods remaining after timeout", node.Namespace, node.Name)
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
		NodeGroup:          nodeGroup,
		NodeName:           node.Name,
		Time:               time.Now(),
		ExpectedDeleteTime: time.Now().Add(MaxCloudProviderNodeDeletionTime),
	})
	return nil
}

func waitForDelayDeletion(node *apiv1.Node, nodeLister kubernetes.NodeLister, timeout time.Duration) errors.AutoscalerError {
	if timeout != 0 && hasDelayDeletionAnnotation(node) {
		klog.V(1).Infof("Wait for removing %s annotations on node %v", DelayDeletionAnnotationPrefix, node.Name)
		err := wait.Poll(5*time.Second, timeout, func() (bool, error) {
			klog.V(5).Infof("Waiting for removing %s annotations on node %v", DelayDeletionAnnotationPrefix, node.Name)
			freshNode, err := nodeLister.Get(node.Name)
			if err != nil || freshNode == nil {
				return false, fmt.Errorf("failed to get node %v: %v", node.Name, err)
			}
			return !hasDelayDeletionAnnotation(freshNode), nil
		})
		if err != nil && err != wait.ErrWaitTimeout {
			return errors.ToAutoscalerError(errors.ApiCallError, err)
		}
		if err == wait.ErrWaitTimeout {
			klog.Warningf("Delay node deletion timed out for node %v, delay deletion annotation wasn't removed within %v, this might slow down scale down.", node.Name, timeout)
		} else {
			klog.V(2).Infof("Annotation %s removed from node %v", DelayDeletionAnnotationPrefix, node.Name)
		}
	}
	return nil
}

func hasDelayDeletionAnnotation(node *apiv1.Node) bool {
	for annotation := range node.Annotations {
		if strings.HasPrefix(annotation, DelayDeletionAnnotationPrefix) {
			return true
		}
	}
	return false
}

func hasNoScaleDownAnnotation(node *apiv1.Node) bool {
	return node.Annotations[ScaleDownDisabledKey] == "true"
}

const (
	apiServerLabelKey   = "component"
	apiServerLabelValue = "kube-apiserver"
)

func filterOutMasters(nodes []*apiv1.Node, pods []*apiv1.Pod) []*apiv1.Node {
	masters := make(map[string]bool)
	for _, pod := range pods {
		if pod.Namespace == metav1.NamespaceSystem && pod.Labels[apiServerLabelKey] == apiServerLabelValue {
			masters[pod.Spec.NodeName] = true
		}
	}

	// if masters aren't on the list of nodes, capacity will be increased on overflowing append
	others := make([]*apiv1.Node, 0, len(nodes)-len(masters))
	for _, node := range nodes {
		if !masters[node.Name] {
			others = append(others, node)
		}
	}

	return others
}
