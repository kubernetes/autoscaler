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
	"strings"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/core/scaleup"
	"k8s.io/autoscaler/cluster-autoscaler/core/utils"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/expander"
	"k8s.io/autoscaler/cluster-autoscaler/metrics"
	ca_processors "k8s.io/autoscaler/cluster-autoscaler/processors"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodegroups"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodegroupset"
	"k8s.io/autoscaler/cluster-autoscaler/processors/status"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	"k8s.io/autoscaler/cluster-autoscaler/utils/klogx"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"

	klog "k8s.io/klog/v2"
)

type skippedReasons struct {
	message []string
}

func (sr *skippedReasons) Reasons() []string {
	return sr.message
}

var (
	backoffReason         = &skippedReasons{[]string{"in backoff after failed scale-up"}}
	maxLimitReachedReason = &skippedReasons{[]string{"max node group size reached"}}
	notReadyReason        = &skippedReasons{[]string{"not ready for scale-up"}}
)

func maxResourceLimitReached(resources []string) *skippedReasons {
	return &skippedReasons{[]string{fmt.Sprintf("max cluster %s limit reached", strings.Join(resources, ", "))}}
}

func computeExpansionOption(context *context.AutoscalingContext, podEquivalenceGroups []*podEquivalenceGroup, nodeGroup cloudprovider.NodeGroup, nodeInfo *schedulerframework.NodeInfo, upcomingNodes []*schedulerframework.NodeInfo) (expander.Option, error) {
	option := expander.Option{
		NodeGroup: nodeGroup,
		Pods:      make([]*apiv1.Pod, 0),
	}

	context.ClusterSnapshot.Fork()

	// add test node to snapshot
	var pods []*apiv1.Pod
	for _, podInfo := range nodeInfo.Pods {
		pods = append(pods, podInfo.Pod)
	}
	if err := context.ClusterSnapshot.AddNodeWithPods(nodeInfo.Node(), pods); err != nil {
		klog.Errorf("Error while adding test Node; %v", err)
		context.ClusterSnapshot.Revert()
		// TODO: Or should I just skip the node group?
		return expander.Option{}, nil
	}

	for _, eg := range podEquivalenceGroups {
		samplePod := eg.pods[0]
		if err := context.PredicateChecker.CheckPredicates(context.ClusterSnapshot, samplePod, nodeInfo.Node().Name); err == nil {
			// add pods to option
			option.Pods = append(option.Pods, eg.pods...)
			// mark pod group as (theoretically) schedulable
			eg.schedulable = true
		} else {
			klog.V(2).Infof("Pod %s can't be scheduled on %s, predicate checking error: %v", samplePod.Name, nodeGroup.Id(), err.VerboseMessage())
			if podCount := len(eg.pods); podCount > 1 {
				klog.V(2).Infof("%d other pods similar to %s can't be scheduled on %s", podCount-1, samplePod.Name, nodeGroup.Id())
			}
			eg.schedulingErrors[nodeGroup.Id()] = err
		}
	}

	context.ClusterSnapshot.Revert()

	if len(option.Pods) > 0 {
		estimator := context.EstimatorBuilder(context.PredicateChecker, context.ClusterSnapshot)
		option.NodeCount, option.Pods = estimator.Estimate(option.Pods, nodeInfo, option.NodeGroup)
	}

	return option, nil
}

func isNodeGroupReadyToScaleUp(nodeGroup cloudprovider.NodeGroup, clusterStateRegistry *clusterstate.ClusterStateRegistry, now time.Time) (bool, *skippedReasons) {
	// Autoprovisioned node groups without nodes are created later so skip check for them.
	if nodeGroup.Exist() && !clusterStateRegistry.IsNodeGroupSafeToScaleUp(nodeGroup, now) {
		// Hack that depends on internals of IsNodeGroupSafeToScaleUp.
		if !clusterStateRegistry.IsNodeGroupHealthy(nodeGroup.Id()) {
			klog.Warningf("Node group %s is not ready for scaleup - unhealthy", nodeGroup.Id())
			return false, notReadyReason
		}
		klog.Warningf("Node group %s is not ready for scaleup - backoff", nodeGroup.Id())
		return false, backoffReason
	}
	return true, nil
}

func isNodeGroupResourceExceeded(ctx *context.AutoscalingContext, resourceManager *scaleup.ResourceManager, resourcesLeft scaleup.ResourcesLimits, nodeGroup cloudprovider.NodeGroup, nodeInfo *schedulerframework.NodeInfo) (bool, *skippedReasons) {
	resourcesDelta, err := resourceManager.DeltaForNode(ctx, nodeInfo, nodeGroup)
	if err != nil {
		klog.Errorf("Skipping node group %s; error getting node group resources: %v", nodeGroup.Id(), err)
		return true, notReadyReason
	}

	checkResult := scaleup.CheckDeltaWithinLimits(resourcesLeft, resourcesDelta)
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
		return true, maxResourceLimitReached(checkResult.ExceededResources)
	}
	return false, nil
}

func getCappedNewNodeCount(context *context.AutoscalingContext, newNodeCount, currentNodeCount int) (int, errors.AutoscalerError) {
	if context.MaxNodes > 0 && newNodeCount+currentNodeCount > context.MaxNodes {
		klog.V(1).Infof("Capping size to max cluster total size (%d)", context.MaxNodes)
		newNodeCount = context.MaxNodes - currentNodeCount
		context.LogRecorder.Eventf(apiv1.EventTypeWarning, "MaxNodesTotalReached", "Max total nodes in cluster reached: %v", context.MaxNodes)
		if newNodeCount < 1 {
			return newNodeCount, errors.NewAutoscalerError(errors.TransientError, "max node total count already reached")
		}
	}
	return newNodeCount, nil
}

// ScaleUp tries to scale the cluster up. Return true if it found a way to increase the size,
// false if it didn't and error if an error occurred. Assumes that all nodes in the cluster are
// ready and in sync with instance groups.
func ScaleUp(context *context.AutoscalingContext, processors *ca_processors.AutoscalingProcessors, clusterStateRegistry *clusterstate.ClusterStateRegistry, resourceManager *scaleup.ResourceManager, unschedulablePods []*apiv1.Pod,
	nodes []*apiv1.Node, daemonSets []*appsv1.DaemonSet, nodeInfos map[string]*schedulerframework.NodeInfo, ignoredTaints taints.TaintKeySet) (*status.ScaleUpStatus, errors.AutoscalerError) {
	// From now on we only care about unschedulable pods that were marked after the newest
	// node became available for the scheduler.
	if len(unschedulablePods) == 0 {
		klog.V(1).Info("No unschedulable pods")
		return &status.ScaleUpStatus{Result: status.ScaleUpNotNeeded}, nil
	}

	loggingQuota := klogx.PodsLoggingQuota()
	for _, pod := range unschedulablePods {
		klogx.V(1).UpTo(loggingQuota).Infof("Pod %s/%s is unschedulable", pod.Namespace, pod.Name)
	}
	klogx.V(1).Over(loggingQuota).Infof("%v other pods are also unschedulable", -loggingQuota.Left())
	podEquivalenceGroups := buildPodEquivalenceGroups(unschedulablePods)

	upcomingNodes := make([]*schedulerframework.NodeInfo, 0)
	for nodeGroup, numberOfNodes := range clusterStateRegistry.GetUpcomingNodes() {
		nodeTemplate, found := nodeInfos[nodeGroup]
		if !found {
			return scaleUpError(&status.ScaleUpStatus{}, errors.NewAutoscalerError(
				errors.InternalError,
				"failed to find template node for node group %s",
				nodeGroup))
		}
		for i := 0; i < numberOfNodes; i++ {
			upcomingNodes = append(upcomingNodes, nodeTemplate)
		}
	}
	klog.V(4).Infof("Upcoming %d nodes", len(upcomingNodes))

	nodeGroups := context.CloudProvider.NodeGroups()
	if processors != nil && processors.NodeGroupListProcessor != nil {
		var errProc error
		nodeGroups, nodeInfos, errProc = processors.NodeGroupListProcessor.Process(context, nodeGroups, nodeInfos, unschedulablePods)
		if errProc != nil {
			return scaleUpError(&status.ScaleUpStatus{}, errors.ToAutoscalerError(errors.InternalError, errProc))
		}
	}

	resourcesLeft, err := resourceManager.ResourcesLeft(context, nodeInfos, nodes)
	if err != nil {
		return scaleUpError(&status.ScaleUpStatus{}, err.AddPrefix("could not compute total resources: "))
	}

	now := time.Now()
	gpuLabel := context.CloudProvider.GPULabel()
	availableGPUTypes := context.CloudProvider.GetAvailableGPUTypes()
	expansionOptions := make(map[string]expander.Option, 0)
	skippedNodeGroups := map[string]status.Reasons{}

	for _, nodeGroup := range nodeGroups {
		if readyToScaleUp, skipReason := isNodeGroupReadyToScaleUp(nodeGroup, clusterStateRegistry, now); !readyToScaleUp {
			if skipReason != nil {
				skippedNodeGroups[nodeGroup.Id()] = skipReason
			}
			continue
		}

		currentTargetSize, err := nodeGroup.TargetSize()
		if err != nil {
			klog.Errorf("Failed to get node group size: %v", err)
			skippedNodeGroups[nodeGroup.Id()] = notReadyReason
			continue
		}
		if currentTargetSize >= nodeGroup.MaxSize() {
			klog.V(4).Infof("Skipping node group %s - max size reached", nodeGroup.Id())
			skippedNodeGroups[nodeGroup.Id()] = maxLimitReachedReason
			continue
		}

		nodeInfo, found := nodeInfos[nodeGroup.Id()]
		if !found {
			klog.Errorf("No node info for: %s", nodeGroup.Id())
			skippedNodeGroups[nodeGroup.Id()] = notReadyReason
			continue
		}

		if exceeded, skipReason := isNodeGroupResourceExceeded(context, resourceManager, resourcesLeft, nodeGroup, nodeInfo); exceeded {
			if skipReason != nil {
				skippedNodeGroups[nodeGroup.Id()] = skipReason
			}
			continue
		}

		option, err := computeExpansionOption(context, podEquivalenceGroups, nodeGroup, nodeInfo, upcomingNodes)
		if err != nil {
			return scaleUpError(&status.ScaleUpStatus{}, errors.ToAutoscalerError(errors.InternalError, err))
		}

		if len(option.Pods) > 0 {
			if option.NodeCount > 0 {
				expansionOptions[nodeGroup.Id()] = option
			} else {
				klog.V(4).Infof("No pod can fit to %s", nodeGroup.Id())
			}
		} else {
			klog.V(4).Infof("No pod can fit to %s", nodeGroup.Id())
		}
	}

	if len(expansionOptions) == 0 {
		klog.V(1).Info("No expansion options")
		return &status.ScaleUpStatus{
			Result:                  status.ScaleUpNoOptionsAvailable,
			PodsRemainUnschedulable: getRemainingPods(podEquivalenceGroups, skippedNodeGroups),
			ConsideredNodeGroups:    nodeGroups,
		}, nil
	}

	// Pick some expansion option.
	options := make([]expander.Option, 0, len(expansionOptions))
	for _, o := range expansionOptions {
		options = append(options, o)
	}
	bestOption := context.ExpanderStrategy.BestOption(options, nodeInfos)
	if bestOption != nil && bestOption.NodeCount > 0 {
		klog.V(1).Infof("Best option to resize: %s", bestOption.NodeGroup.Id())
		if len(bestOption.Debug) > 0 {
			klog.V(1).Info(bestOption.Debug)
		}
		klog.V(1).Infof("Estimated %d nodes needed in %s", bestOption.NodeCount, bestOption.NodeGroup.Id())

		newNodes := bestOption.NodeCount
		newNodeCount, err := getCappedNewNodeCount(context, newNodes, len(nodes)+len(upcomingNodes))
		if err != nil {
			return scaleUpError(&status.ScaleUpStatus{PodsTriggeredScaleUp: bestOption.Pods}, err)
		}
		newNodes = newNodeCount

		createNodeGroupResults := make([]nodegroups.CreateNodeGroupResult, 0)
		if !bestOption.NodeGroup.Exist() {
			oldId := bestOption.NodeGroup.Id()
			createNodeGroupResult, err := processors.NodeGroupManager.CreateNodeGroup(context, bestOption.NodeGroup)
			if err != nil {
				return scaleUpError(
					&status.ScaleUpStatus{FailedCreationNodeGroups: []cloudprovider.NodeGroup{bestOption.NodeGroup}, PodsTriggeredScaleUp: bestOption.Pods},
					err)
			}
			createNodeGroupResults = append(createNodeGroupResults, createNodeGroupResult)
			bestOption.NodeGroup = createNodeGroupResult.MainCreatedNodeGroup

			// If possible replace candidate node-info with node info based on crated node group. The latter
			// one should be more in line with nodes which will be created by node group.
			mainCreatedNodeInfo, err := utils.GetNodeInfoFromTemplate(createNodeGroupResult.MainCreatedNodeGroup, daemonSets, context.PredicateChecker, ignoredTaints)
			if err == nil {
				nodeInfos[createNodeGroupResult.MainCreatedNodeGroup.Id()] = mainCreatedNodeInfo
			} else {
				klog.Warningf("Cannot build node info for newly created main node group %v; balancing similar node groups may not work; err=%v", createNodeGroupResult.MainCreatedNodeGroup.Id(), err)
				// Use node info based on expansion candidate but upadte Id which likely changed when node group was created.
				nodeInfos[bestOption.NodeGroup.Id()] = nodeInfos[oldId]
			}

			if oldId != createNodeGroupResult.MainCreatedNodeGroup.Id() {
				delete(nodeInfos, oldId)
			}

			for _, nodeGroup := range createNodeGroupResult.ExtraCreatedNodeGroups {
				nodeInfo, err := utils.GetNodeInfoFromTemplate(nodeGroup, daemonSets, context.PredicateChecker, ignoredTaints)

				if err != nil {
					klog.Warningf("Cannot build node info for newly created extra node group %v; balancing similar node groups will not work; err=%v", nodeGroup.Id(), err)
					continue
				}
				nodeInfos[nodeGroup.Id()] = nodeInfo

				option, err2 := computeExpansionOption(context, podEquivalenceGroups, nodeGroup, nodeInfo, upcomingNodes)
				if err2 != nil {
					return scaleUpError(&status.ScaleUpStatus{PodsTriggeredScaleUp: bestOption.Pods}, errors.ToAutoscalerError(errors.InternalError, err))
				}

				if len(option.Pods) > 0 && option.NodeCount > 0 {
					expansionOptions[nodeGroup.Id()] = option
				}
			}

			// Update ClusterStateRegistry so similar nodegroups rebalancing works.
			// TODO(lukaszos) when pursuing scalability update this call with one which takes list of changed node groups so we do not
			//                do extra API calls. (the call at the bottom of ScaleUp() could be also changed then)
			clusterStateRegistry.Recalculate()
		}

		nodeInfo, found := nodeInfos[bestOption.NodeGroup.Id()]
		if !found {
			// This should never happen, as we already should have retrieved
			// nodeInfo for any considered nodegroup.
			klog.Errorf("No node info for: %s", bestOption.NodeGroup.Id())
			return scaleUpError(
				&status.ScaleUpStatus{CreateNodeGroupResults: createNodeGroupResults, PodsTriggeredScaleUp: bestOption.Pods},
				errors.NewAutoscalerError(
					errors.CloudProviderError,
					"No node info for best expansion option!"))
		}

		// apply upper limits for CPU and memory
		newNodes, err = resourceManager.ApplyResourcesLimits(context, newNodes, resourcesLeft, nodeInfo, bestOption.NodeGroup)
		if err != nil {
			return scaleUpError(
				&status.ScaleUpStatus{CreateNodeGroupResults: createNodeGroupResults, PodsTriggeredScaleUp: bestOption.Pods},
				err)
		}

		targetNodeGroups := []cloudprovider.NodeGroup{bestOption.NodeGroup}
		if context.BalanceSimilarNodeGroups {
			similarNodeGroups, typedErr := processors.NodeGroupSetProcessor.FindSimilarNodeGroups(context, bestOption.NodeGroup, nodeInfos)
			if typedErr != nil {
				return scaleUpError(
					&status.ScaleUpStatus{CreateNodeGroupResults: createNodeGroupResults, PodsTriggeredScaleUp: bestOption.Pods},
					typedErr.AddPrefix("failed to find matching node groups: "))
			}

			similarNodeGroups = filterNodeGroupsByPods(similarNodeGroups, bestOption.Pods, expansionOptions)
			for _, ng := range similarNodeGroups {
				if clusterStateRegistry.IsNodeGroupSafeToScaleUp(ng, now) {
					targetNodeGroups = append(targetNodeGroups, ng)
				} else {
					// This should never happen, as we will filter out the node group earlier on
					// because of missing entry in podsPassingPredicates, but double checking doesn't
					// really cost us anything
					klog.V(2).Infof("Ignoring node group %s when balancing: group is not ready for scaleup", ng.Id())
				}
			}

			if len(targetNodeGroups) > 1 {
				var names = []string{}
				for _, ng := range targetNodeGroups {
					names = append(names, ng.Id())
				}
				klog.V(1).Infof("Splitting scale-up between %v similar node groups: {%v}", len(targetNodeGroups), strings.Join(names, ", "))
			}
		}

		scaleUpInfos, typedErr := processors.NodeGroupSetProcessor.BalanceScaleUpBetweenGroups(
			context, targetNodeGroups, newNodes)
		if typedErr != nil {
			return scaleUpError(
				&status.ScaleUpStatus{CreateNodeGroupResults: createNodeGroupResults, PodsTriggeredScaleUp: bestOption.Pods},
				typedErr)
		}

		klog.V(1).Infof("Final scale-up plan: %v", scaleUpInfos)
		for _, info := range scaleUpInfos {
			typedErr := executeScaleUp(context, clusterStateRegistry, info, gpu.GetGpuTypeForMetrics(gpuLabel, availableGPUTypes, nodeInfo.Node(), nil), now)
			if typedErr != nil {
				return scaleUpError(
					&status.ScaleUpStatus{
						CreateNodeGroupResults: createNodeGroupResults,
						FailedResizeNodeGroups: []cloudprovider.NodeGroup{info.Group},
						PodsTriggeredScaleUp:   bestOption.Pods,
					},
					typedErr,
				)
			}
		}

		clusterStateRegistry.Recalculate()
		return &status.ScaleUpStatus{
			Result:                  status.ScaleUpSuccessful,
			ScaleUpInfos:            scaleUpInfos,
			PodsRemainUnschedulable: getRemainingPods(podEquivalenceGroups, skippedNodeGroups),
			ConsideredNodeGroups:    nodeGroups,
			CreateNodeGroupResults:  createNodeGroupResults,
			PodsTriggeredScaleUp:    bestOption.Pods,
			PodsAwaitEvaluation:     getPodsAwaitingEvaluation(podEquivalenceGroups, bestOption.NodeGroup.Id()),
		}, nil
	}

	return &status.ScaleUpStatus{
		Result:                  status.ScaleUpNoOptionsAvailable,
		PodsRemainUnschedulable: getRemainingPods(podEquivalenceGroups, skippedNodeGroups),
		ConsideredNodeGroups:    nodeGroups,
	}, nil
}

// ScaleUpToNodeGroupMinSize tries to scale up node groups that have less nodes than the configured min size.
// The source of truth for the current node group size is the TargetSize queried directly from cloud providers.
// Return the scale up status (ScaleUpNotNeeded, ScaleUpSuccessful or FailedResizeNodeGroups) and errors if any.
func ScaleUpToNodeGroupMinSize(context *context.AutoscalingContext, processors *ca_processors.AutoscalingProcessors, clusterStateRegistry *clusterstate.ClusterStateRegistry, resourceManager *scaleup.ResourceManager,
	nodes []*apiv1.Node, nodeInfos map[string]*schedulerframework.NodeInfo) (*status.ScaleUpStatus, errors.AutoscalerError) {
	now := time.Now()
	nodeGroups := context.CloudProvider.NodeGroups()
	gpuLabel := context.CloudProvider.GPULabel()
	availableGPUTypes := context.CloudProvider.GetAvailableGPUTypes()
	scaleUpInfos := make([]nodegroupset.ScaleUpInfo, 0)

	resourcesLeft, err := resourceManager.ResourcesLeft(context, nodeInfos, nodes)
	if err != nil {
		return scaleUpError(&status.ScaleUpStatus{}, err.AddPrefix("could not compute total resources: "))
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

		if readyToScaleUp, skipReason := isNodeGroupReadyToScaleUp(ng, clusterStateRegistry, now); !readyToScaleUp {
			klog.Warningf("ScaleUpToNodeGroupMinSize: node group is ready to scale up: %v", skipReason)
			continue
		}

		nodeInfo, found := nodeInfos[ng.Id()]
		if !found {
			klog.Warningf("ScaleUpToNodeGroupMinSize: no node info for %s", ng.Id())
			continue
		}

		exceeded, skipReason := isNodeGroupResourceExceeded(context, resourceManager, resourcesLeft, ng, nodeInfo)
		if exceeded {
			klog.Warning("ScaleUpToNodeGroupMinSize: node group resource excceded: %v", skipReason)
			continue
		}

		newNodeCount := ng.MinSize() - targetSize
		newNodeCount, err = resourceManager.ApplyResourcesLimits(context, newNodeCount, resourcesLeft, nodeInfo, ng)
		if err != nil {
			klog.Warning("ScaleUpToNodeGroupMinSize: failed to apply resource limits: %v", err)
			continue
		}

		newNodeCount, err = getCappedNewNodeCount(context, newNodeCount, targetSize)
		if err != nil {
			klog.Warning("ScaleUpToNodeGroupMinSize: failed to get capped node count: %v", err)
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
	for _, info := range scaleUpInfos {
		nodeInfo, ok := nodeInfos[info.Group.Id()]
		if !ok {
			klog.Warningf("ScaleUpToNodeGroupMinSize: failed to get node info for node group %s", info.Group.Id())
			continue
		}

		gpuType := gpu.GetGpuTypeForMetrics(gpuLabel, availableGPUTypes, nodeInfo.Node(), nil)
		if err := executeScaleUp(context, clusterStateRegistry, info, gpuType, now); err != nil {
			return scaleUpError(
				&status.ScaleUpStatus{
					FailedResizeNodeGroups: []cloudprovider.NodeGroup{info.Group},
				},
				err,
			)
		}
	}

	clusterStateRegistry.Recalculate()
	return &status.ScaleUpStatus{
		Result:               status.ScaleUpSuccessful,
		ScaleUpInfos:         scaleUpInfos,
		ConsideredNodeGroups: nodeGroups,
	}, nil
}

func getRemainingPods(egs []*podEquivalenceGroup, skipped map[string]status.Reasons) []status.NoScaleUpInfo {
	remaining := []status.NoScaleUpInfo{}
	for _, eg := range egs {
		if eg.schedulable {
			continue
		}
		for _, pod := range eg.pods {
			noScaleUpInfo := status.NoScaleUpInfo{
				Pod:                pod,
				RejectedNodeGroups: eg.schedulingErrors,
				SkippedNodeGroups:  skipped,
			}
			remaining = append(remaining, noScaleUpInfo)
		}
	}
	return remaining
}

func getPodsAwaitingEvaluation(egs []*podEquivalenceGroup, bestOption string) []*apiv1.Pod {
	awaitsEvaluation := []*apiv1.Pod{}
	for _, eg := range egs {
		if eg.schedulable {
			if _, found := eg.schedulingErrors[bestOption]; found {
				// Schedulable, but not yet.
				awaitsEvaluation = append(awaitsEvaluation, eg.pods...)
			}
		}
	}
	return awaitsEvaluation
}

func filterNodeGroupsByPods(
	groups []cloudprovider.NodeGroup,
	podsRequiredToFit []*apiv1.Pod,
	expansionOptions map[string]expander.Option) []cloudprovider.NodeGroup {

	result := make([]cloudprovider.NodeGroup, 0)

	for _, group := range groups {
		option, found := expansionOptions[group.Id()]
		if !found {
			klog.V(1).Infof("No info about pods passing predicates found for group %v, skipping it from scale-up consideration", group.Id())
			continue
		}
		fittingPods := make(map[*apiv1.Pod]bool, len(option.Pods))
		for _, pod := range option.Pods {
			fittingPods[pod] = true
		}
		allFit := true
		for _, pod := range podsRequiredToFit {
			if _, found := fittingPods[pod]; !found {
				klog.V(1).Infof("Group %v, can't fit pod %v/%v, removing from scale-up consideration", group.Id(), pod.Namespace, pod.Name)
				allFit = false
				break
			}
		}
		if allFit {
			result = append(result, group)
		}
	}

	return result
}

func executeScaleUp(context *context.AutoscalingContext, clusterStateRegistry *clusterstate.ClusterStateRegistry, info nodegroupset.ScaleUpInfo, gpuType string, now time.Time) errors.AutoscalerError {
	klog.V(0).Infof("Scale-up: setting group %s size to %d", info.Group.Id(), info.NewSize)
	context.LogRecorder.Eventf(apiv1.EventTypeNormal, "ScaledUpGroup",
		"Scale-up: setting group %s size to %d instead of %d (max: %d)", info.Group.Id(), info.NewSize, info.CurrentSize, info.MaxSize)
	increase := info.NewSize - info.CurrentSize
	if err := info.Group.IncreaseSize(increase); err != nil {
		context.LogRecorder.Eventf(apiv1.EventTypeWarning, "FailedToScaleUpGroup", "Scale-up failed for group %s: %v", info.Group.Id(), err)
		aerr := errors.ToAutoscalerError(errors.CloudProviderError, err).AddPrefix("failed to increase node group size: ")
		clusterStateRegistry.RegisterFailedScaleUp(info.Group, metrics.FailedScaleUpReason(string(aerr.Type())), now)
		return aerr
	}
	clusterStateRegistry.RegisterOrUpdateScaleUp(
		info.Group,
		increase,
		time.Now())
	metrics.RegisterScaleUp(increase, gpuType)
	context.LogRecorder.Eventf(apiv1.EventTypeNormal, "ScaledUpGroup",
		"Scale-up: group %s size set to %d instead of %d (max: %d)", info.Group.Id(), info.NewSize, info.CurrentSize, info.MaxSize)
	return nil
}

func scaleUpError(s *status.ScaleUpStatus, err errors.AutoscalerError) (*status.ScaleUpStatus, errors.AutoscalerError) {
	s.ScaleUpError = &err
	s.Result = status.ScaleUpError
	return s, err
}
