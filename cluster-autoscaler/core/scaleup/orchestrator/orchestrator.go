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

package orchestrator

import (
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaleup/equivalence"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaleup/resource"
	"k8s.io/autoscaler/cluster-autoscaler/estimator"
	"k8s.io/autoscaler/cluster-autoscaler/expander"
	"k8s.io/autoscaler/cluster-autoscaler/metrics"
	ca_processors "k8s.io/autoscaler/cluster-autoscaler/processors"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodegroups"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodegroupset"
	"k8s.io/autoscaler/cluster-autoscaler/processors/status"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/klogx"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"
	"k8s.io/klog/v2"
)

// ScaleUpOrchestrator implements scaleup.Orchestrator interface.
type ScaleUpOrchestrator struct {
	autoscalingContext   *context.AutoscalingContext
	processors           *ca_processors.AutoscalingProcessors
	resourceManager      *resource.Manager
	clusterStateRegistry *clusterstate.ClusterStateRegistry
	scaleUpExecutor      *scaleUpExecutor
	estimatorBuilder     estimator.EstimatorBuilder
	taintConfig          taints.TaintConfig
	initialized          bool
}

// New returns new instance of scale up Orchestrator.
func New() *ScaleUpOrchestrator {
	return &ScaleUpOrchestrator{
		initialized: false,
	}
}

// Initialize initializes the orchestrator object with required fields.
func (o *ScaleUpOrchestrator) Initialize(
	autoscalingContext *context.AutoscalingContext,
	processors *ca_processors.AutoscalingProcessors,
	clusterStateRegistry *clusterstate.ClusterStateRegistry,
	estimatorBuilder estimator.EstimatorBuilder,
	taintConfig taints.TaintConfig,
) {
	o.autoscalingContext = autoscalingContext
	o.processors = processors
	o.clusterStateRegistry = clusterStateRegistry
	o.estimatorBuilder = estimatorBuilder
	o.taintConfig = taintConfig
	o.resourceManager = resource.NewManager(processors.CustomResourcesProcessor)
	o.scaleUpExecutor = newScaleUpExecutor(autoscalingContext, processors.ScaleStateNotifier, o.processors.AsyncNodeGroupStateChecker)
	o.initialized = true
}

// ScaleUp tries to scale the cluster up. Returns appropriate status or error if
// an unexpected error occurred. Assumes that all nodes in the cluster are ready
// and in sync with instance groups.
func (o *ScaleUpOrchestrator) ScaleUp(
	unschedulablePods []*apiv1.Pod,
	nodes []*apiv1.Node,
	daemonSets []*appsv1.DaemonSet,
	nodeInfos map[string]*framework.NodeInfo,
	allOrNothing bool, // Either request enough capacity for all unschedulablePods, or don't request it at all.
) (*status.ScaleUpStatus, errors.AutoscalerError) {
	if !o.initialized {
		return status.UpdateScaleUpError(&status.ScaleUpStatus{}, errors.NewAutoscalerError(errors.InternalError, "ScaleUpOrchestrator is not initialized"))
	}

	loggingQuota := klogx.PodsLoggingQuota()
	for _, pod := range unschedulablePods {
		klogx.V(1).UpTo(loggingQuota).Infof("Pod %s/%s is unschedulable", pod.Namespace, pod.Name)
	}
	klogx.V(1).Over(loggingQuota).Infof("%v other pods are also unschedulable", -loggingQuota.Left())

	buildPodEquivalenceGroupsStart := time.Now()
	podEquivalenceGroups := equivalence.BuildPodGroups(unschedulablePods)
	metrics.UpdateDurationFromStart(metrics.BuildPodEquivalenceGroups, buildPodEquivalenceGroupsStart)

	upcomingNodes, aErr := o.UpcomingNodes(nodeInfos)
	if aErr != nil {
		return status.UpdateScaleUpError(&status.ScaleUpStatus{}, aErr.AddPrefix("could not get upcoming nodes: "))
	}
	klog.V(4).Infof("Upcoming %d nodes", len(upcomingNodes))

	nodeGroups := o.autoscalingContext.CloudProvider.NodeGroups()
	if o.processors != nil && o.processors.NodeGroupListProcessor != nil {
		var err error
		nodeGroups, nodeInfos, err = o.processors.NodeGroupListProcessor.Process(o.autoscalingContext, nodeGroups, nodeInfos, unschedulablePods)
		if err != nil {
			return status.UpdateScaleUpError(&status.ScaleUpStatus{}, errors.ToAutoscalerError(errors.InternalError, err))
		}
	}

	// Initialise binpacking limiter.
	o.processors.BinpackingLimiter.InitBinpacking(o.autoscalingContext, nodeGroups)

	resourcesLeft, aErr := o.resourceManager.ResourcesLeft(o.autoscalingContext, nodeInfos, nodes)
	if aErr != nil {
		return status.UpdateScaleUpError(&status.ScaleUpStatus{}, aErr.AddPrefix("could not compute total resources: "))
	}

	now := time.Now()

	// Filter out invalid node groups
	validNodeGroups, skippedNodeGroups := o.filterValidScaleUpNodeGroups(nodeGroups, nodeInfos, resourcesLeft, len(nodes)+len(upcomingNodes), now)

	// Mark skipped node groups as processed.
	for nodegroupID := range skippedNodeGroups {
		o.processors.BinpackingLimiter.MarkProcessed(o.autoscalingContext, nodegroupID)
	}

	// Calculate expansion options
	schedulablePodGroups := map[string][]estimator.PodEquivalenceGroup{}
	var options []expander.Option

	for _, nodeGroup := range validNodeGroups {
		schedulablePodGroups[nodeGroup.Id()] = o.SchedulablePodGroups(podEquivalenceGroups, nodeGroup, nodeInfos[nodeGroup.Id()])
	}

	for _, nodeGroup := range validNodeGroups {
		option := o.ComputeExpansionOption(nodeGroup, schedulablePodGroups, nodeInfos, len(nodes)+len(upcomingNodes), now, allOrNothing)
		o.processors.BinpackingLimiter.MarkProcessed(o.autoscalingContext, nodeGroup.Id())

		if len(option.Pods) == 0 || option.NodeCount == 0 {
			klog.V(4).Infof("No pod can fit to %s", nodeGroup.Id())
		} else if allOrNothing && len(option.Pods) < len(unschedulablePods) {
			klog.V(4).Infof("Some pods can't fit to %s, giving up due to all-or-nothing scale-up strategy", nodeGroup.Id())
		} else {
			options = append(options, option)
		}

		if o.processors.BinpackingLimiter.StopBinpacking(o.autoscalingContext, options) {
			break
		}
	}

	// Finalize binpacking limiter.
	o.processors.BinpackingLimiter.FinalizeBinpacking(o.autoscalingContext, options)

	if len(options) == 0 {
		klog.V(1).Info("No expansion options")
		return &status.ScaleUpStatus{
			Result:                  status.ScaleUpNoOptionsAvailable,
			PodsRemainUnschedulable: GetRemainingPods(podEquivalenceGroups, skippedNodeGroups),
			ConsideredNodeGroups:    nodeGroups,
		}, nil
	}

	// Pick some expansion option.
	bestOption := o.autoscalingContext.ExpanderStrategy.BestOption(options, nodeInfos)
	if bestOption == nil || bestOption.NodeCount <= 0 {
		return &status.ScaleUpStatus{
			Result:                  status.ScaleUpNoOptionsAvailable,
			PodsRemainUnschedulable: GetRemainingPods(podEquivalenceGroups, skippedNodeGroups),
			ConsideredNodeGroups:    nodeGroups,
		}, nil
	}
	klog.V(1).Infof("Best option to resize: %s", bestOption.NodeGroup.Id())
	if len(bestOption.Debug) > 0 {
		klog.V(1).Info(bestOption.Debug)
	}
	klog.V(1).Infof("Estimated %d nodes needed in %s", bestOption.NodeCount, bestOption.NodeGroup.Id())

	// Cap new nodes to supported number of nodes in the cluster.
	newNodes, aErr := o.GetCappedNewNodeCount(bestOption.NodeCount, len(nodes)+len(upcomingNodes))
	if aErr != nil {
		return status.UpdateScaleUpError(&status.ScaleUpStatus{PodsTriggeredScaleUp: bestOption.Pods}, aErr)
	}

	newNodes, aErr = o.applyLimits(newNodes, resourcesLeft, bestOption.NodeGroup, nodeInfos)
	if aErr != nil {
		return status.UpdateScaleUpError(
			&status.ScaleUpStatus{PodsTriggeredScaleUp: bestOption.Pods},
			aErr)
	}

	if newNodes < bestOption.NodeCount {
		klog.V(1).Infof("Only %d nodes can be added to %s due to cluster-wide limits", newNodes, bestOption.NodeGroup.Id())
		if allOrNothing {
			// Can't execute a scale-up that will accommodate all pods, so nothing is considered schedulable.
			klog.V(1).Info("Not attempting scale-up due to all-or-nothing strategy: not all pods would be accommodated")
			markedEquivalenceGroups := markAllGroupsAsUnschedulable(podEquivalenceGroups, AllOrNothingReason)
			return buildNoOptionsAvailableStatus(markedEquivalenceGroups, skippedNodeGroups, nodeGroups), nil
		}
	}

	// If necessary, create the node group. This is no longer simulation, an empty node group will be created by cloud provider if supported.
	createNodeGroupResults := make([]nodegroups.CreateNodeGroupResult, 0)
	if !bestOption.NodeGroup.Exist() && !o.processors.AsyncNodeGroupStateChecker.IsUpcoming(bestOption.NodeGroup) {
		if allOrNothing && bestOption.NodeGroup.MaxSize() < newNodes {
			klog.V(1).Infof("Can only create a new node group with max %d nodes, need %d nodes", bestOption.NodeGroup.MaxSize(), newNodes)
			// Can't execute a scale-up that will accommodate all pods, so nothing is considered schedulable.
			klog.V(1).Info("Not attempting scale-up due to all-or-nothing strategy: not all pods would be accommodated")
			markedEquivalenceGroups := markAllGroupsAsUnschedulable(podEquivalenceGroups, AllOrNothingReason)
			return buildNoOptionsAvailableStatus(markedEquivalenceGroups, skippedNodeGroups, nodeGroups), nil
		}
		var scaleUpStatus *status.ScaleUpStatus
		createNodeGroupResults, scaleUpStatus, aErr = o.CreateNodeGroup(bestOption, nodeInfos, schedulablePodGroups, podEquivalenceGroups, daemonSets, allOrNothing)
		if aErr != nil {
			return scaleUpStatus, aErr
		}
	}

	scaleUpInfos, aErr := o.balanceScaleUps(now, bestOption.NodeGroup, newNodes, nodeInfos, schedulablePodGroups)
	if aErr != nil {
		return status.UpdateScaleUpError(
			&status.ScaleUpStatus{CreateNodeGroupResults: createNodeGroupResults, PodsTriggeredScaleUp: bestOption.Pods},
			aErr)
	}

	// Last check before scale-up. Node group capacity (both due to max size limits & current size) is only checked when balancing.
	totalCapacity := 0
	for _, sui := range scaleUpInfos {
		totalCapacity += sui.NewSize - sui.CurrentSize
	}
	if totalCapacity < newNodes {
		klog.V(1).Infof("Can only add %d nodes due to node group limits, need %d nodes", totalCapacity, newNodes)
		if allOrNothing {
			// Can't execute a scale-up that will accommodate all pods, so nothing is considered schedulable.
			klog.V(1).Info("Not attempting scale-up due to all-or-nothing strategy: not all pods would be accommodated")
			markedEquivalenceGroups := markAllGroupsAsUnschedulable(podEquivalenceGroups, AllOrNothingReason)
			return buildNoOptionsAvailableStatus(markedEquivalenceGroups, skippedNodeGroups, nodeGroups), nil
		}
	}

	// Execute scale up.
	klog.V(1).Infof("Final scale-up plan: %v", scaleUpInfos)
	aErr, failedNodeGroups := o.scaleUpExecutor.ExecuteScaleUps(scaleUpInfos, nodeInfos, now, allOrNothing)
	if aErr != nil {
		return status.UpdateScaleUpError(
			&status.ScaleUpStatus{
				CreateNodeGroupResults: createNodeGroupResults,
				FailedResizeNodeGroups: failedNodeGroups,
				PodsTriggeredScaleUp:   bestOption.Pods,
			},
			aErr,
		)
	}

	o.clusterStateRegistry.Recalculate()
	return &status.ScaleUpStatus{
		Result:                  status.ScaleUpSuccessful,
		ScaleUpInfos:            scaleUpInfos,
		PodsRemainUnschedulable: GetRemainingPods(podEquivalenceGroups, skippedNodeGroups),
		ConsideredNodeGroups:    nodeGroups,
		CreateNodeGroupResults:  createNodeGroupResults,
		PodsTriggeredScaleUp:    bestOption.Pods,
		PodsAwaitEvaluation:     GetPodsAwaitingEvaluation(podEquivalenceGroups, bestOption.NodeGroup.Id()),
	}, nil
}

func (o *ScaleUpOrchestrator) applyLimits(newNodes int, resourcesLeft resource.Limits, nodeGroup cloudprovider.NodeGroup, nodeInfos map[string]*framework.NodeInfo) (int, errors.AutoscalerError) {
	nodeInfo, found := nodeInfos[nodeGroup.Id()]
	if !found {
		// This should never happen, as we already should have retrieved nodeInfo for any considered nodegroup.
		klog.Errorf("No node info for: %s", nodeGroup.Id())
		return 0, errors.NewAutoscalerError(errors.CloudProviderError, "No node info for best expansion option!")
	}
	return o.resourceManager.ApplyLimits(o.autoscalingContext, newNodes, resourcesLeft, nodeInfo, nodeGroup)
}

// ScaleUpToNodeGroupMinSize tries to scale up node groups that have less nodes
// than the configured min size. The source of truth for the current node group
// size is the TargetSize queried directly from cloud providers. Returns
// appropriate status or error if an unexpected error occurred.
func (o *ScaleUpOrchestrator) ScaleUpToNodeGroupMinSize(
	nodes []*apiv1.Node,
	nodeInfos map[string]*framework.NodeInfo,
) (*status.ScaleUpStatus, errors.AutoscalerError) {
	if !o.initialized {
		return status.UpdateScaleUpError(&status.ScaleUpStatus{}, errors.NewAutoscalerError(errors.InternalError, "ScaleUpOrchestrator is not initialized"))
	}

	now := time.Now()
	nodeGroups := o.autoscalingContext.CloudProvider.NodeGroups()
	scaleUpInfos := make([]nodegroupset.ScaleUpInfo, 0)

	resourcesLeft, aErr := o.resourceManager.ResourcesLeft(o.autoscalingContext, nodeInfos, nodes)
	if aErr != nil {
		return status.UpdateScaleUpError(&status.ScaleUpStatus{}, aErr.AddPrefix("could not compute total resources: "))
	}

	for _, ng := range nodeGroups {
		if !ng.Exist() {
			klog.Warningf("ScaleUpToNodeGroupMinSize: NodeGroup %s does not exist", ng.Id())
			continue
		}

		targetSize, err := ng.TargetSize()
		if err != nil {
			klog.Warningf("ScaleUpToNodeGroupMinSize: failed to get target size of node group %s", ng.Id())
			continue
		}

		klog.V(4).Infof("ScaleUpToNodeGroupMinSize: NodeGroup %s, TargetSize %d, MinSize %d, MaxSize %d", ng.Id(), targetSize, ng.MinSize(), ng.MaxSize())
		if targetSize >= ng.MinSize() {
			continue
		}

		if skipReason := o.IsNodeGroupReadyToScaleUp(ng, now); skipReason != nil {
			klog.Warningf("ScaleUpToNodeGroupMinSize: node group is ready to scale up: %v", skipReason)
			continue
		}

		nodeInfo, found := nodeInfos[ng.Id()]
		if !found {
			klog.Warningf("ScaleUpToNodeGroupMinSize: no node info for %s", ng.Id())
			continue
		}

		if skipReason := o.IsNodeGroupResourceExceeded(resourcesLeft, ng, nodeInfo, 1); skipReason != nil {
			klog.Warningf("ScaleUpToNodeGroupMinSize: node group resource excceded: %v", skipReason)
			continue
		}

		newNodeCount := ng.MinSize() - targetSize
		newNodeCount, err = o.resourceManager.ApplyLimits(o.autoscalingContext, newNodeCount, resourcesLeft, nodeInfo, ng)
		if err != nil {
			klog.Warningf("ScaleUpToNodeGroupMinSize: failed to apply resource limits: %v", err)
			continue
		}

		newNodeCount, err = o.GetCappedNewNodeCount(newNodeCount, targetSize)
		if err != nil {
			klog.Warningf("ScaleUpToNodeGroupMinSize: failed to get capped node count: %v", err)
			continue
		}

		info := nodegroupset.ScaleUpInfo{
			Group:       ng,
			CurrentSize: targetSize,
			NewSize:     targetSize + newNodeCount,
			MaxSize:     ng.MaxSize(),
		}
		scaleUpInfos = append(scaleUpInfos, info)
	}

	if len(scaleUpInfos) == 0 {
		klog.V(1).Info("ScaleUpToNodeGroupMinSize: scale up not needed")
		return &status.ScaleUpStatus{Result: status.ScaleUpNotNeeded}, nil
	}

	klog.V(1).Infof("ScaleUpToNodeGroupMinSize: final scale-up plan: %v", scaleUpInfos)
	aErr, failedNodeGroups := o.scaleUpExecutor.ExecuteScaleUps(scaleUpInfos, nodeInfos, now, false /* allOrNothing disabled */)
	if aErr != nil {
		return status.UpdateScaleUpError(
			&status.ScaleUpStatus{
				FailedResizeNodeGroups: failedNodeGroups,
			},
			aErr,
		)
	}

	o.clusterStateRegistry.Recalculate()
	return &status.ScaleUpStatus{
		Result:               status.ScaleUpSuccessful,
		ScaleUpInfos:         scaleUpInfos,
		ConsideredNodeGroups: nodeGroups,
	}, nil
}

// filterValidScaleUpNodeGroups filters the node groups that are valid for scale-up
func (o *ScaleUpOrchestrator) filterValidScaleUpNodeGroups(
	nodeGroups []cloudprovider.NodeGroup,
	nodeInfos map[string]*framework.NodeInfo,
	resourcesLeft resource.Limits,
	currentNodeCount int,
	now time.Time,
) ([]cloudprovider.NodeGroup, map[string]status.Reasons) {
	var validNodeGroups []cloudprovider.NodeGroup
	skippedNodeGroups := map[string]status.Reasons{}

	for _, nodeGroup := range nodeGroups {
		if skipReason := o.IsNodeGroupReadyToScaleUp(nodeGroup, now); skipReason != nil {
			skippedNodeGroups[nodeGroup.Id()] = skipReason
			continue
		}

		currentTargetSize, err := nodeGroup.TargetSize()
		if err != nil {
			klog.Errorf("Failed to get node group size: %v", err)
			skippedNodeGroups[nodeGroup.Id()] = NotReadyReason
			continue
		}
		if currentTargetSize >= nodeGroup.MaxSize() {
			klog.V(4).Infof("Skipping node group %s - max size reached", nodeGroup.Id())
			skippedNodeGroups[nodeGroup.Id()] = MaxLimitReachedReason
			continue
		}
		autoscalingOptions, err := nodeGroup.GetOptions(o.autoscalingContext.NodeGroupDefaults)
		if err != nil && err != cloudprovider.ErrNotImplemented {
			klog.Errorf("Couldn't get autoscaling options for ng: %v", nodeGroup.Id())
		}
		numNodes := 1
		if autoscalingOptions != nil && autoscalingOptions.ZeroOrMaxNodeScaling {
			numNodes = nodeGroup.MaxSize() - currentTargetSize
			if o.autoscalingContext.MaxNodesTotal != 0 && currentNodeCount+numNodes > o.autoscalingContext.MaxNodesTotal {
				klog.V(4).Infof("Skipping node group %s - atomic scale-up exceeds cluster node count limit", nodeGroup.Id())
				skippedNodeGroups[nodeGroup.Id()] = NewSkippedReasons("atomic scale-up exceeds cluster node count limit")
				continue
			}
		}

		nodeInfo, found := nodeInfos[nodeGroup.Id()]
		if !found {
			klog.Errorf("No node info for: %s", nodeGroup.Id())
			skippedNodeGroups[nodeGroup.Id()] = NotReadyReason
			continue
		}
		if skipReason := o.IsNodeGroupResourceExceeded(resourcesLeft, nodeGroup, nodeInfo, numNodes); skipReason != nil {
			skippedNodeGroups[nodeGroup.Id()] = skipReason
			continue
		}

		validNodeGroups = append(validNodeGroups, nodeGroup)
	}
	return validNodeGroups, skippedNodeGroups
}

// ComputeExpansionOption computes expansion option based on pending pods and cluster state.
func (o *ScaleUpOrchestrator) ComputeExpansionOption(
	nodeGroup cloudprovider.NodeGroup,
	schedulablePodGroups map[string][]estimator.PodEquivalenceGroup,
	nodeInfos map[string]*framework.NodeInfo,
	currentNodeCount int,
	now time.Time,
	allOrNothing bool,
) expander.Option {
	option := expander.Option{NodeGroup: nodeGroup}
	podGroups := schedulablePodGroups[nodeGroup.Id()]
	nodeInfo := nodeInfos[nodeGroup.Id()]

	if len(podGroups) == 0 {
		return option
	}

	option.SimilarNodeGroups = o.ComputeSimilarNodeGroups(nodeGroup, nodeInfos, schedulablePodGroups, now)

	estimateStart := time.Now()
	expansionEstimator := o.estimatorBuilder(
		o.autoscalingContext.PredicateChecker,
		o.autoscalingContext.ClusterSnapshot,
		estimator.NewEstimationContext(o.autoscalingContext.MaxNodesTotal, option.SimilarNodeGroups, currentNodeCount),
	)
	option.NodeCount, option.Pods = expansionEstimator.Estimate(podGroups, nodeInfo, nodeGroup)
	metrics.UpdateDurationFromStart(metrics.Estimate, estimateStart)

	autoscalingOptions, err := nodeGroup.GetOptions(o.autoscalingContext.NodeGroupDefaults)
	if err != nil && err != cloudprovider.ErrNotImplemented {
		klog.Errorf("Failed to get autoscaling options for node group %s: %v", nodeGroup.Id(), err)
	}

	// Special handling for groups that only scale from zero to max.
	if autoscalingOptions != nil && autoscalingOptions.ZeroOrMaxNodeScaling {
		// For zero-or-max scaling groups, the only valid value of node count is node group's max size.
		if allOrNothing && option.NodeCount > nodeGroup.MaxSize() {
			// We would have to cap the node count, which means not all pods will be
			// accommodated. This violates the principle of all-or-nothing strategy.
			option.Pods = nil
			option.NodeCount = 0
		}
		if option.NodeCount > 0 {
			// Cap or increase the number of nodes to the only valid value - node group's max size.
			option.NodeCount = nodeGroup.MaxSize()
		}
	}

	return option
}

// CreateNodeGroup will try to create a new node group based on the initialOption.
func (o *ScaleUpOrchestrator) CreateNodeGroup(
	initialOption *expander.Option,
	nodeInfos map[string]*framework.NodeInfo,
	schedulablePodGroups map[string][]estimator.PodEquivalenceGroup,
	podEquivalenceGroups []*equivalence.PodGroup,
	daemonSets []*appsv1.DaemonSet,
	allOrNothing bool,
) ([]nodegroups.CreateNodeGroupResult, *status.ScaleUpStatus, errors.AutoscalerError) {
	createNodeGroupResults := make([]nodegroups.CreateNodeGroupResult, 0)

	oldId := initialOption.NodeGroup.Id()
	var createNodeGroupResult nodegroups.CreateNodeGroupResult
	var aErr errors.AutoscalerError
	if o.autoscalingContext.AsyncNodeGroupsEnabled {
		initializer := newAsyncNodeGroupInitializer(initialOption.NodeGroup, nodeInfos[oldId], o.scaleUpExecutor, o.taintConfig, daemonSets, o.processors.ScaleUpStatusProcessor, o.autoscalingContext, allOrNothing)
		createNodeGroupResult, aErr = o.processors.NodeGroupManager.CreateNodeGroupAsync(o.autoscalingContext, initialOption.NodeGroup, initializer)
	} else {
		createNodeGroupResult, aErr = o.processors.NodeGroupManager.CreateNodeGroup(o.autoscalingContext, initialOption.NodeGroup)
	}
	if aErr != nil {
		status, err := status.UpdateScaleUpError(
			&status.ScaleUpStatus{FailedCreationNodeGroups: []cloudprovider.NodeGroup{initialOption.NodeGroup}, PodsTriggeredScaleUp: initialOption.Pods},
			aErr)
		return createNodeGroupResults, status, err
	}

	createNodeGroupResults = append(createNodeGroupResults, createNodeGroupResult)
	initialOption.NodeGroup = createNodeGroupResult.MainCreatedNodeGroup

	// If possible replace candidate node-info with node info based on crated node group. The latter
	// one should be more in line with nodes which will be created by node group.
	mainCreatedNodeInfo, aErr := simulator.TemplateNodeInfoFromNodeGroupTemplate(createNodeGroupResult.MainCreatedNodeGroup, daemonSets, o.taintConfig)
	if aErr == nil {
		nodeInfos[createNodeGroupResult.MainCreatedNodeGroup.Id()] = mainCreatedNodeInfo
		schedulablePodGroups[createNodeGroupResult.MainCreatedNodeGroup.Id()] = o.SchedulablePodGroups(podEquivalenceGroups, createNodeGroupResult.MainCreatedNodeGroup, mainCreatedNodeInfo)
	} else {
		klog.Warningf("Cannot build node info for newly created main node group %v; balancing similar node groups may not work; err=%v", createNodeGroupResult.MainCreatedNodeGroup.Id(), aErr)
		// Use node info based on expansion candidate but update Id which likely changed when node group was created.
		nodeInfos[createNodeGroupResult.MainCreatedNodeGroup.Id()] = nodeInfos[oldId]
		schedulablePodGroups[createNodeGroupResult.MainCreatedNodeGroup.Id()] = schedulablePodGroups[oldId]
	}
	if oldId != createNodeGroupResult.MainCreatedNodeGroup.Id() {
		delete(nodeInfos, oldId)
		delete(schedulablePodGroups, oldId)
	}
	for _, nodeGroup := range createNodeGroupResult.ExtraCreatedNodeGroups {
		nodeInfo, aErr := simulator.TemplateNodeInfoFromNodeGroupTemplate(nodeGroup, daemonSets, o.taintConfig)
		if aErr != nil {
			klog.Warningf("Cannot build node info for newly created extra node group %v; balancing similar node groups will not work; err=%v", nodeGroup.Id(), aErr)
			continue
		}
		nodeInfos[nodeGroup.Id()] = nodeInfo
		schedulablePodGroups[nodeGroup.Id()] = o.SchedulablePodGroups(podEquivalenceGroups, nodeGroup, nodeInfo)
	}

	// Update ClusterStateRegistry so similar nodegroups rebalancing works.
	// TODO(lukaszos) when pursuing scalability update this call with one which takes list of changed node groups so we do not
	//                do extra API calls. (the call at the bottom of ScaleUp() could be also changed then)
	o.clusterStateRegistry.Recalculate()
	return createNodeGroupResults, nil, nil
}

// SchedulablePodGroups returns a list of pods that could be scheduled
// in a given node group after a scale up.
func (o *ScaleUpOrchestrator) SchedulablePodGroups(
	podEquivalenceGroups []*equivalence.PodGroup,
	nodeGroup cloudprovider.NodeGroup,
	nodeInfo *framework.NodeInfo,
) []estimator.PodEquivalenceGroup {
	o.autoscalingContext.ClusterSnapshot.Fork()
	defer o.autoscalingContext.ClusterSnapshot.Revert()

	// Add test node to snapshot.
	if err := o.autoscalingContext.ClusterSnapshot.AddNodeInfo(nodeInfo); err != nil {
		klog.Errorf("Error while adding test Node: %v", err)
		return []estimator.PodEquivalenceGroup{}
	}

	var schedulablePodGroups []estimator.PodEquivalenceGroup
	for _, eg := range podEquivalenceGroups {
		samplePod := eg.Pods[0]
		if _, err := o.autoscalingContext.PredicateChecker.CheckPredicates(o.autoscalingContext.ClusterSnapshot, samplePod, nodeInfo.Node().Name); err == nil {
			// Add pods to option.
			schedulablePodGroups = append(schedulablePodGroups, estimator.PodEquivalenceGroup{
				Pods: eg.Pods,
			})
			// Mark pod group as (theoretically) schedulable.
			eg.Schedulable = true
			eg.SchedulableGroups = append(eg.SchedulableGroups, nodeGroup.Id())
		} else {
			klog.V(2).Infof("Pod %s/%s can't be scheduled on %s, predicate checking error: %v", samplePod.Namespace, samplePod.Name, nodeGroup.Id(), err.VerboseMessage())
			if podCount := len(eg.Pods); podCount > 1 {
				klog.V(2).Infof("%d other pods similar to %s can't be scheduled on %s", podCount-1, samplePod.Name, nodeGroup.Id())
			}
			eg.SchedulingErrors[nodeGroup.Id()] = err
		}
	}

	return schedulablePodGroups
}

// UpcomingNodes returns a list of nodes that are not ready but should be.
func (o *ScaleUpOrchestrator) UpcomingNodes(nodeInfos map[string]*framework.NodeInfo) ([]*framework.NodeInfo, errors.AutoscalerError) {
	upcomingCounts, _ := o.clusterStateRegistry.GetUpcomingNodes()
	upcomingNodes := make([]*framework.NodeInfo, 0)
	for nodeGroup, numberOfNodes := range upcomingCounts {
		nodeTemplate, found := nodeInfos[nodeGroup]
		if !found {
			return nil, errors.NewAutoscalerError(errors.InternalError, "failed to find template node for node group %s", nodeGroup)
		}
		for i := 0; i < numberOfNodes; i++ {
			upcomingNodes = append(upcomingNodes, nodeTemplate)
		}
	}
	return upcomingNodes, nil
}

// IsNodeGroupReadyToScaleUp returns nil if node group is ready to be scaled up, otherwise a reason is provided.
func (o *ScaleUpOrchestrator) IsNodeGroupReadyToScaleUp(nodeGroup cloudprovider.NodeGroup, now time.Time) *SkippedReasons {
	// Non-existing node groups are created later so skip check for them.
	if !nodeGroup.Exist() {
		return nil
	}
	if scaleUpSafety := o.clusterStateRegistry.NodeGroupScaleUpSafety(nodeGroup, now); !scaleUpSafety.SafeToScale {
		if !scaleUpSafety.Healthy {
			klog.Warningf("Node group %s is not ready for scaleup - unhealthy", nodeGroup.Id())
			return NotReadyReason
		}
		klog.Warningf("Node group %s is not ready for scaleup - backoff with status: %v", nodeGroup.Id(), scaleUpSafety.BackoffStatus)
		return BackoffReason
	}
	return nil
}

// IsNodeGroupResourceExceeded returns nil if node group resource limits are not exceeded, otherwise a reason is provided.
func (o *ScaleUpOrchestrator) IsNodeGroupResourceExceeded(resourcesLeft resource.Limits, nodeGroup cloudprovider.NodeGroup, nodeInfo *framework.NodeInfo, numNodes int) status.Reasons {
	resourcesDelta, err := o.resourceManager.DeltaForNode(o.autoscalingContext, nodeInfo, nodeGroup)
	if err != nil {
		klog.Errorf("Skipping node group %s; error getting node group resources: %v", nodeGroup.Id(), err)
		return NotReadyReason
	}

	for resource, delta := range resourcesDelta {
		resourcesDelta[resource] = delta * int64(numNodes)
	}

	checkResult := resource.CheckDeltaWithinLimits(resourcesLeft, resourcesDelta)
	if checkResult.Exceeded {
		klog.V(4).Infof("Skipping node group %s; maximal limit exceeded for %v", nodeGroup.Id(), checkResult.ExceededResources)
		for _, resource := range checkResult.ExceededResources {
			switch resource {
			case cloudprovider.ResourceNameCores:
				metrics.RegisterSkippedScaleUpCPU()
			case cloudprovider.ResourceNameMemory:
				metrics.RegisterSkippedScaleUpMemory()
			default:
				continue
			}
		}
		return NewMaxResourceLimitReached(checkResult.ExceededResources)
	}
	return nil
}

// GetCappedNewNodeCount caps resize according to cluster wide node count limit.
func (o *ScaleUpOrchestrator) GetCappedNewNodeCount(newNodeCount, currentNodeCount int) (int, errors.AutoscalerError) {
	if o.autoscalingContext.MaxNodesTotal > 0 && newNodeCount+currentNodeCount > o.autoscalingContext.MaxNodesTotal {
		klog.V(1).Infof("Capping size to max cluster total size (%d)", o.autoscalingContext.MaxNodesTotal)
		newNodeCount = o.autoscalingContext.MaxNodesTotal - currentNodeCount
		o.autoscalingContext.LogRecorder.Eventf(apiv1.EventTypeWarning, "MaxNodesTotalReached", "Max total nodes in cluster reached: %v", o.autoscalingContext.MaxNodesTotal)
		if newNodeCount < 1 {
			return newNodeCount, errors.NewAutoscalerError(errors.TransientError, "max node total count already reached")
		}
	}
	return newNodeCount, nil
}

func (o *ScaleUpOrchestrator) balanceScaleUps(
	now time.Time,
	nodeGroup cloudprovider.NodeGroup,
	newNodes int,
	nodeInfos map[string]*framework.NodeInfo,
	schedulablePodGroups map[string][]estimator.PodEquivalenceGroup,
) ([]nodegroupset.ScaleUpInfo, errors.AutoscalerError) {
	// Recompute similar node groups in case they need to be updated
	similarNodeGroups := o.ComputeSimilarNodeGroups(nodeGroup, nodeInfos, schedulablePodGroups, now)
	if similarNodeGroups != nil {
		// if similar node groups are found, log about them
		similarNodeGroupIds := make([]string, 0)
		for _, sng := range similarNodeGroups {
			similarNodeGroupIds = append(similarNodeGroupIds, sng.Id())
		}
		klog.V(2).Infof("Found %d similar node groups: %v", len(similarNodeGroups), similarNodeGroupIds)
	} else if o.autoscalingContext.BalanceSimilarNodeGroups {
		// if no similar node groups are found and the flag is enabled, log about it
		klog.V(2).Info("No similar node groups found")
	}

	targetNodeGroups := []cloudprovider.NodeGroup{nodeGroup}
	for _, ng := range similarNodeGroups {
		targetNodeGroups = append(targetNodeGroups, ng)
	}

	if len(targetNodeGroups) > 1 {
		var names []string
		for _, ng := range targetNodeGroups {
			names = append(names, ng.Id())
		}
		klog.V(1).Infof("Splitting scale-up between %v similar node groups: {%v}", len(targetNodeGroups), strings.Join(names, ", "))
	}
	return o.processors.NodeGroupSetProcessor.BalanceScaleUpBetweenGroups(o.autoscalingContext, targetNodeGroups, newNodes)
}

// ComputeSimilarNodeGroups finds similar node groups which can schedule the same
// set of pods as the main node group.
func (o *ScaleUpOrchestrator) ComputeSimilarNodeGroups(
	nodeGroup cloudprovider.NodeGroup,
	nodeInfos map[string]*framework.NodeInfo,
	schedulablePodGroups map[string][]estimator.PodEquivalenceGroup,
	now time.Time,
) []cloudprovider.NodeGroup {
	if !o.autoscalingContext.BalanceSimilarNodeGroups {
		return nil
	}

	autoscalingOptions, err := nodeGroup.GetOptions(o.autoscalingContext.NodeGroupDefaults)
	if err != nil && err != cloudprovider.ErrNotImplemented {
		klog.Errorf("Failed to get autoscaling options for node group %s: %v", nodeGroup.Id(), err)
	}
	if autoscalingOptions != nil && autoscalingOptions.ZeroOrMaxNodeScaling {
		return nil
	}

	podGroups, found := schedulablePodGroups[nodeGroup.Id()]
	if !found || len(podGroups) == 0 {
		return nil
	}

	similarNodeGroups, err := o.processors.NodeGroupSetProcessor.FindSimilarNodeGroups(o.autoscalingContext, nodeGroup, nodeInfos)
	if err != nil {
		klog.Errorf("Failed to find similar node groups: %v", err)
		return nil
	}

	var validSimilarNodeGroups []cloudprovider.NodeGroup
	for _, ng := range similarNodeGroups {
		// Non-existing node groups are created later so skip check for them.
		if ng.Exist() && !o.clusterStateRegistry.NodeGroupScaleUpSafety(ng, now).SafeToScale {
			klog.V(2).Infof("Ignoring node group %s when balancing: group is not ready for scaleup", ng.Id())
		} else if similarPodGroups, found := schedulablePodGroups[ng.Id()]; found && matchingSchedulablePodGroups(podGroups, similarPodGroups) {
			validSimilarNodeGroups = append(validSimilarNodeGroups, ng)
		}
	}

	return validSimilarNodeGroups
}

func matchingSchedulablePodGroups(podGroups []estimator.PodEquivalenceGroup, similarPodGroups []estimator.PodEquivalenceGroup) bool {
	schedulableSamplePods := make(map[*apiv1.Pod]bool)
	for _, podGroup := range similarPodGroups {
		schedulableSamplePods[podGroup.Exemplar()] = true
	}
	for _, podGroup := range podGroups {
		if _, found := schedulableSamplePods[podGroup.Exemplar()]; !found {
			return false
		}
	}
	return true
}

func markAllGroupsAsUnschedulable(egs []*equivalence.PodGroup, reason status.Reasons) []*equivalence.PodGroup {
	for _, eg := range egs {
		if eg.Schedulable {
			if eg.SchedulingErrors == nil {
				eg.SchedulingErrors = map[string]status.Reasons{}
			}
			for _, sg := range eg.SchedulableGroups {
				eg.SchedulingErrors[sg] = reason
			}
			eg.Schedulable = false
		}
	}
	return egs
}

func buildNoOptionsAvailableStatus(egs []*equivalence.PodGroup, skipped map[string]status.Reasons, ngs []cloudprovider.NodeGroup) *status.ScaleUpStatus {
	return &status.ScaleUpStatus{
		Result:                  status.ScaleUpNoOptionsAvailable,
		PodsRemainUnschedulable: GetRemainingPods(egs, skipped),
		ConsideredNodeGroups:    ngs,
	}
}

// GetRemainingPods returns information about pods which CA is unable to help
// at this moment.
func GetRemainingPods(egs []*equivalence.PodGroup, skipped map[string]status.Reasons) []status.NoScaleUpInfo {
	remaining := []status.NoScaleUpInfo{}
	for _, eg := range egs {
		if eg.Schedulable {
			continue
		}
		for _, pod := range eg.Pods {
			noScaleUpInfo := status.NoScaleUpInfo{
				Pod:                pod,
				RejectedNodeGroups: eg.SchedulingErrors,
				SkippedNodeGroups:  skipped,
			}
			remaining = append(remaining, noScaleUpInfo)
		}
	}
	return remaining
}

// GetPodsAwaitingEvaluation returns list of pods for which CA was unable to help
// this scale up loop (but should be able to help).
func GetPodsAwaitingEvaluation(egs []*equivalence.PodGroup, bestOption string) []*apiv1.Pod {
	awaitsEvaluation := []*apiv1.Pod{}
	for _, eg := range egs {
		if eg.Schedulable {
			if _, found := eg.SchedulingErrors[bestOption]; found {
				// Schedulable, but not yet.
				awaitsEvaluation = append(awaitsEvaluation, eg.Pods...)
			}
		}
	}
	return awaitsEvaluation
}
