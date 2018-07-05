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
	"bytes"
	"fmt"
	"math"
	"time"

	apiv1 "k8s.io/api/core/v1"
	extensionsv1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/estimator"
	"k8s.io/autoscaler/cluster-autoscaler/expander"
	"k8s.io/autoscaler/cluster-autoscaler/metrics"
	ca_processors "k8s.io/autoscaler/cluster-autoscaler/processors"
	"k8s.io/autoscaler/cluster-autoscaler/processors/status"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/glogx"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	"k8s.io/autoscaler/cluster-autoscaler/utils/nodegroupset"
	schedulercache "k8s.io/kubernetes/pkg/scheduler/cache"

	"github.com/golang/glog"
)

type scaleUpResourcesLimits map[string]int64
type scaleUpResourcesDelta map[string]int64

// used as a value in scaleUpResourcesLimits if actual limit could not be obtained due to errors talking to cloud provider
const scaleUpLimitUnknown = math.MaxInt64

func computeScaleUpResourcesLeftLimits(
	nodeGroups []cloudprovider.NodeGroup,
	nodeInfos map[string]*schedulercache.NodeInfo,
	nodesFromNotAutoscaledGroups []*apiv1.Node,
	resourceLimiter *cloudprovider.ResourceLimiter) (scaleUpResourcesLimits, errors.AutoscalerError) {
	totalCores, totalMem, errCoresMem := calculateScaleUpCoresMemoryTotal(nodeGroups, nodeInfos, nodesFromNotAutoscaledGroups)

	var totalGpus map[string]int64
	var totalGpusErr error
	if cloudprovider.ContainsGpuResources(resourceLimiter.GetResources()) {
		totalGpus, totalGpusErr = calculateScaleUpGpusTotal(nodeGroups, nodeInfos, nodesFromNotAutoscaledGroups)
	}

	resultScaleUpLimits := make(scaleUpResourcesLimits)
	for _, resource := range resourceLimiter.GetResources() {
		max := resourceLimiter.GetMax(resource)

		// we put only actual limits into final map. No entry means no limit.
		if max > 0 {
			if (resource == cloudprovider.ResourceNameCores || resource == cloudprovider.ResourceNameMemory) && errCoresMem != nil {
				// core resource info missing - no reason to proceed with scale up
				return scaleUpResourcesLimits{}, errCoresMem
			}
			switch {
			case resource == cloudprovider.ResourceNameCores:
				if errCoresMem != nil {
					resultScaleUpLimits[resource] = scaleUpLimitUnknown
				} else {
					resultScaleUpLimits[resource] = computeBelowMax(totalCores, max)
				}

			case resource == cloudprovider.ResourceNameMemory:
				if errCoresMem != nil {
					resultScaleUpLimits[resource] = scaleUpLimitUnknown
				} else {
					resultScaleUpLimits[resource] = computeBelowMax(totalMem, max)
				}

			case cloudprovider.IsGpuResource(resource):
				if totalGpusErr != nil {
					resultScaleUpLimits[resource] = scaleUpLimitUnknown
				} else {
					resultScaleUpLimits[resource] = computeBelowMax(totalGpus[resource], max)
				}

			default:
				glog.Errorf("Scale up limits defined for unsupported resource '%s'", resource)
			}
		}
	}

	return resultScaleUpLimits, nil
}

func calculateScaleUpCoresMemoryTotal(
	nodeGroups []cloudprovider.NodeGroup,
	nodeInfos map[string]*schedulercache.NodeInfo,
	nodesFromNotAutoscaledGroups []*apiv1.Node) (int64, int64, errors.AutoscalerError) {
	var coresTotal int64
	var memoryTotal int64

	for _, nodeGroup := range nodeGroups {
		currentSize, err := nodeGroup.TargetSize()
		if err != nil {
			return 0, 0, errors.ToAutoscalerError(errors.CloudProviderError, err).AddPrefix("Failed to get node group size of %v:", nodeGroup.Id())
		}
		nodeInfo, found := nodeInfos[nodeGroup.Id()]
		if !found {
			return 0, 0, errors.ToAutoscalerError(errors.CloudProviderError, err).AddPrefix("No node info for: %s", nodeGroup.Id())
		}
		if currentSize > 0 {
			nodeCPU, nodeMemory := getNodeInfoCoresAndMemory(nodeInfo)
			coresTotal = coresTotal + int64(currentSize)*nodeCPU
			memoryTotal = memoryTotal + int64(currentSize)*nodeMemory
		}
	}

	for _, node := range nodesFromNotAutoscaledGroups {
		cores, memory := getNodeCoresAndMemory(node)
		coresTotal += cores
		memoryTotal += memory
	}

	return coresTotal, memoryTotal, nil
}

func calculateScaleUpGpusTotal(
	nodeGroups []cloudprovider.NodeGroup,
	nodeInfos map[string]*schedulercache.NodeInfo,
	nodesFromNotAutoscaledGroups []*apiv1.Node) (map[string]int64, errors.AutoscalerError) {

	result := make(map[string]int64)
	for _, nodeGroup := range nodeGroups {
		currentSize, err := nodeGroup.TargetSize()
		if err != nil {
			return nil, errors.ToAutoscalerError(errors.CloudProviderError, err).AddPrefix("Failed to get node group size of %v:", nodeGroup.Id())
		}
		nodeInfo, found := nodeInfos[nodeGroup.Id()]
		if !found {
			return nil, errors.ToAutoscalerError(errors.CloudProviderError, err).AddPrefix("No node info for: %s", nodeGroup.Id())
		}
		if currentSize > 0 {
			gpuType, gpuCount, err := gpu.GetNodeTargetGpus(nodeInfo.Node(), nodeGroup)
			if err != nil {
				return nil, errors.ToAutoscalerError(errors.CloudProviderError, err).AddPrefix("Failed to get target gpu for node group %v:", nodeGroup.Id())
			}
			if gpuType == "" {
				continue
			}
			result[gpuType] += gpuCount * int64(currentSize)
		}
	}

	for _, node := range nodesFromNotAutoscaledGroups {
		gpuType, gpuCount, err := gpu.GetNodeTargetGpus(node, nil)
		if err != nil {
			return nil, errors.ToAutoscalerError(errors.CloudProviderError, err).AddPrefix("Failed to get target gpu for node gpus count for node %v:", node.Name)
		}
		result[gpuType] += gpuCount
	}

	return result, nil
}

func computeBelowMax(total int64, max int64) int64 {
	if total < max {
		return max - total
	}
	return 0
}

func computeScaleUpResourcesDelta(nodeInfo *schedulercache.NodeInfo, nodeGroup cloudprovider.NodeGroup, resourceLimiter *cloudprovider.ResourceLimiter) (scaleUpResourcesDelta, errors.AutoscalerError) {
	resultScaleUpDelta := make(scaleUpResourcesDelta)

	nodeCPU, nodeMemory := getNodeInfoCoresAndMemory(nodeInfo)
	resultScaleUpDelta[cloudprovider.ResourceNameCores] = nodeCPU
	resultScaleUpDelta[cloudprovider.ResourceNameMemory] = nodeMemory

	if cloudprovider.ContainsGpuResources(resourceLimiter.GetResources()) {
		gpuType, gpuCount, err := gpu.GetNodeTargetGpus(nodeInfo.Node(), nodeGroup)
		if err != nil {
			return scaleUpResourcesDelta{}, errors.ToAutoscalerError(errors.CloudProviderError, err).AddPrefix("Failed to get target gpu for node group %v:", nodeGroup.Id())
		}
		resultScaleUpDelta[gpuType] = gpuCount
	}

	return resultScaleUpDelta, nil
}

type scaleUpLimitsCheckResult struct {
	exceeded          bool
	exceededResources []string
}

func scaleUpLimitsNotExceeded() scaleUpLimitsCheckResult {
	return scaleUpLimitsCheckResult{false, []string{}}
}

func (limits *scaleUpResourcesLimits) checkScaleUpDeltaWithinLimits(delta scaleUpResourcesDelta) scaleUpLimitsCheckResult {
	exceededResources := sets.NewString()
	for resource, resourceDelta := range delta {
		resourceLeft, found := (*limits)[resource]
		if found {
			if (resourceDelta > 0) && (resourceLeft == scaleUpLimitUnknown || resourceDelta > resourceLeft) {
				exceededResources.Insert(resource)
			}
		}
	}
	if len(exceededResources) > 0 {
		return scaleUpLimitsCheckResult{true, exceededResources.List()}
	}

	return scaleUpLimitsNotExceeded()
}

func getNodeInfoCoresAndMemory(nodeInfo *schedulercache.NodeInfo) (int64, int64) {
	return getNodeCoresAndMemory(nodeInfo.Node())
}

// ScaleUp tries to scale the cluster up. Return true if it found a way to increase the size,
// false if it didn't and error if an error occurred. Assumes that all nodes in the cluster are
// ready and in sync with instance groups.
func ScaleUp(context *context.AutoscalingContext, processors *ca_processors.AutoscalingProcessors, clusterStateRegistry *clusterstate.ClusterStateRegistry, unschedulablePods []*apiv1.Pod,
	nodes []*apiv1.Node, daemonSets []*extensionsv1.DaemonSet) (*status.ScaleUpStatus, errors.AutoscalerError) {
	// From now on we only care about unschedulable pods that were marked after the newest
	// node became available for the scheduler.
	if len(unschedulablePods) == 0 {
		glog.V(1).Info("No unschedulable pods")
		return &status.ScaleUpStatus{ScaledUp: false}, nil
	}

	now := time.Now()

	loggingQuota := glogx.PodsLoggingQuota()

	podsRemainUnschedulable := make(map[*apiv1.Pod]bool)

	for _, pod := range unschedulablePods {
		glogx.V(1).UpTo(loggingQuota).Infof("Pod %s/%s is unschedulable", pod.Namespace, pod.Name)
		podsRemainUnschedulable[pod] = true
	}
	glogx.V(1).Over(loggingQuota).Infof("%v other pods are also unschedulable", -loggingQuota.Left())
	nodeInfos, err := GetNodeInfosForGroups(nodes, context.CloudProvider, context.ClientSet,
		daemonSets, context.PredicateChecker)
	if err != nil {
		return nil, err.AddPrefix("failed to build node infos for node groups: ")
	}

	nodesFromNotAutoscaledGroups, err := FilterOutNodesFromNotAutoscaledGroups(nodes, context.CloudProvider)
	if err != nil {
		return nil, err.AddPrefix("failed to filter out nodes which are from not autoscaled groups: ")
	}

	nodeGroups := context.CloudProvider.NodeGroups()

	resourceLimiter, errCP := context.CloudProvider.GetResourceLimiter()
	if errCP != nil {
		return nil, errors.ToAutoscalerError(
			errors.CloudProviderError,
			errCP)
	}

	scaleUpResourcesLeft, errLimits := computeScaleUpResourcesLeftLimits(nodeGroups, nodeInfos, nodesFromNotAutoscaledGroups, resourceLimiter)
	if errLimits != nil {
		return nil, errLimits.AddPrefix("Could not compute total resources: ")
	}

	upcomingNodes := make([]*schedulercache.NodeInfo, 0)
	for nodeGroup, numberOfNodes := range clusterStateRegistry.GetUpcomingNodes() {
		nodeTemplate, found := nodeInfos[nodeGroup]
		if !found {
			return nil, errors.NewAutoscalerError(
				errors.InternalError,
				"failed to find template node for node group %s",
				nodeGroup)
		}
		for i := 0; i < numberOfNodes; i++ {
			upcomingNodes = append(upcomingNodes, nodeTemplate)
		}
	}
	glog.V(4).Infof("Upcoming %d nodes", len(upcomingNodes))

	podsPassingPredicates := make(map[string][]*apiv1.Pod)
	expansionOptions := make([]expander.Option, 0)

	if processors != nil && processors.NodeGroupListProcessor != nil {
		var errProc error
		nodeGroups, nodeInfos, errProc = processors.NodeGroupListProcessor.Process(context, nodeGroups, nodeInfos, unschedulablePods)
		if errProc != nil {
			return nil, errors.ToAutoscalerError(errors.InternalError, errProc)
		}
	}

	for _, nodeGroup := range nodeGroups {
		// Autoprovisioned node groups without nodes are created later so skip check for them.
		if nodeGroup.Exist() && !clusterStateRegistry.IsNodeGroupSafeToScaleUp(nodeGroup.Id(), now) {
			glog.Warningf("Node group %s is not ready for scaleup", nodeGroup.Id())
			continue
		}

		currentTargetSize, err := nodeGroup.TargetSize()
		if err != nil {
			glog.Errorf("Failed to get node group size: %v", err)
			continue
		}
		if currentTargetSize >= nodeGroup.MaxSize() {
			// skip this node group.
			glog.V(4).Infof("Skipping node group %s - max size reached", nodeGroup.Id())
			continue
		}

		nodeInfo, found := nodeInfos[nodeGroup.Id()]
		if !found {
			glog.Errorf("No node info for: %s", nodeGroup.Id())
			continue
		}

		scaleUpResourcesDelta, err := computeScaleUpResourcesDelta(nodeInfo, nodeGroup, resourceLimiter)
		if err != nil {
			glog.Errorf("Skipping node group %s; error getting node group resources: %v", nodeGroup.Id(), err)
			continue
		}
		checkResult := scaleUpResourcesLeft.checkScaleUpDeltaWithinLimits(scaleUpResourcesDelta)
		if checkResult.exceeded {
			glog.V(4).Infof("Skipping node group %s; maximal limit exceeded for %v", nodeGroup.Id(), checkResult.exceededResources)
			continue
		}

		option := expander.Option{
			NodeGroup: nodeGroup,
			Pods:      make([]*apiv1.Pod, 0),
		}

		option.Pods = FilterSchedulablePodsForNode(context, unschedulablePods, nodeGroup.Id(), nodeInfo)
		for _, pod := range option.Pods {
			podsRemainUnschedulable[pod] = false
		}
		passingPods := make([]*apiv1.Pod, len(option.Pods))
		copy(passingPods, option.Pods)
		podsPassingPredicates[nodeGroup.Id()] = passingPods

		if len(option.Pods) > 0 {
			if context.EstimatorName == estimator.BinpackingEstimatorName {
				binpackingEstimator := estimator.NewBinpackingNodeEstimator(context.PredicateChecker)
				option.NodeCount = binpackingEstimator.Estimate(option.Pods, nodeInfo, upcomingNodes)
			} else if context.EstimatorName == estimator.BasicEstimatorName {
				basicEstimator := estimator.NewBasicNodeEstimator()
				for _, pod := range option.Pods {
					basicEstimator.Add(pod)
				}
				option.NodeCount, option.Debug = basicEstimator.Estimate(nodeInfo.Node(), upcomingNodes)
			} else {
				glog.Fatalf("Unrecognized estimator: %s", context.EstimatorName)
			}
			if option.NodeCount > 0 {
				expansionOptions = append(expansionOptions, option)
			} else {
				glog.V(2).Infof("No need for any nodes in %s", nodeGroup.Id())
			}
		} else {
			glog.V(4).Infof("No pod can fit to %s", nodeGroup.Id())
		}
	}

	if len(expansionOptions) == 0 {
		glog.V(1).Info("No expansion options")
		return &status.ScaleUpStatus{ScaledUp: false, PodsRemainUnschedulable: getRemainingPods(podsRemainUnschedulable)}, nil
	}

	// Pick some expansion option.
	bestOption := context.ExpanderStrategy.BestOption(expansionOptions, nodeInfos)
	if bestOption != nil && bestOption.NodeCount > 0 {
		glog.V(1).Infof("Best option to resize: %s", bestOption.NodeGroup.Id())
		if len(bestOption.Debug) > 0 {
			glog.V(1).Info(bestOption.Debug)
		}
		glog.V(1).Infof("Estimated %d nodes needed in %s", bestOption.NodeCount, bestOption.NodeGroup.Id())

		newNodes := bestOption.NodeCount

		if context.MaxNodesTotal > 0 && len(nodes)+newNodes > context.MaxNodesTotal {
			glog.V(1).Infof("Capping size to max cluster total size (%d)", context.MaxNodesTotal)
			newNodes = context.MaxNodesTotal - len(nodes)
			if newNodes < 1 {
				return nil, errors.NewAutoscalerError(
					errors.TransientError,
					"max node total count already reached")
			}
		}

		if !bestOption.NodeGroup.Exist() {
			oldId := bestOption.NodeGroup.Id()
			bestOption.NodeGroup, err = processors.NodeGroupManager.CreateNodeGroup(context, bestOption.NodeGroup)
			if err != nil {
				return nil, err
			}
			// Node group id may change when we create node group and we need to update
			// our data structures.
			if oldId != bestOption.NodeGroup.Id() {
				podsPassingPredicates[bestOption.NodeGroup.Id()] = podsPassingPredicates[oldId]
				delete(podsPassingPredicates, oldId)
				nodeInfos[bestOption.NodeGroup.Id()] = nodeInfos[oldId]
				delete(nodeInfos, oldId)
			}
		}

		nodeInfo, found := nodeInfos[bestOption.NodeGroup.Id()]
		if !found {
			// This should never happen, as we already should have retrieved
			// nodeInfo for any considered nodegroup.
			glog.Errorf("No node info for: %s", bestOption.NodeGroup.Id())
			return nil, errors.NewAutoscalerError(
				errors.CloudProviderError,
				"No node info for best expansion option!")
		}

		// apply upper limits for CPU and memory
		newNodes, err = applyScaleUpResourcesLimits(newNodes, scaleUpResourcesLeft, nodeInfo, bestOption.NodeGroup, resourceLimiter)
		if err != nil {
			return nil, err
		}

		targetNodeGroups := []cloudprovider.NodeGroup{bestOption.NodeGroup}
		if context.BalanceSimilarNodeGroups {
			similarNodeGroups, typedErr := nodegroupset.FindSimilarNodeGroups(bestOption.NodeGroup, context.CloudProvider, nodeInfos)
			if typedErr != nil {
				return nil, typedErr.AddPrefix("Failed to find matching node groups: ")
			}
			similarNodeGroups = filterNodeGroupsByPods(similarNodeGroups, bestOption.Pods, podsPassingPredicates)
			for _, ng := range similarNodeGroups {
				if clusterStateRegistry.IsNodeGroupSafeToScaleUp(ng.Id(), now) {
					targetNodeGroups = append(targetNodeGroups, ng)
				} else {
					// This should never happen, as we will filter out the node group earlier on
					// because of missing entry in podsPassingPredicates, but double checking doesn't
					// really cost us anything
					glog.V(2).Infof("Ignoring node group %s when balancing: group is not ready for scaleup", ng.Id())
				}
			}
			if len(targetNodeGroups) > 1 {
				var buffer bytes.Buffer
				for i, ng := range targetNodeGroups {
					if i > 0 {
						buffer.WriteString(", ")
					}
					buffer.WriteString(ng.Id())
				}
				glog.V(1).Infof("Splitting scale-up between %v similar node groups: {%v}", len(targetNodeGroups), buffer.String())
			}
		}
		scaleUpInfos, typedErr := nodegroupset.BalanceScaleUpBetweenGroups(
			targetNodeGroups, newNodes)
		if typedErr != nil {
			return nil, typedErr
		}
		glog.V(1).Infof("Final scale-up plan: %v", scaleUpInfos)
		for _, info := range scaleUpInfos {
			typedErr := executeScaleUp(context, clusterStateRegistry, info, gpu.GetGpuTypeForMetrics(nodeInfo.Node(), nil))
			if typedErr != nil {
				return nil, typedErr
			}
		}

		clusterStateRegistry.Recalculate()
		return &status.ScaleUpStatus{
				ScaledUp:                true,
				ScaleUpInfos:            scaleUpInfos,
				PodsRemainUnschedulable: getRemainingPods(podsRemainUnschedulable),
				PodsTriggeredScaleUp:    bestOption.Pods,
				PodsAwaitEvaluation:     getPodsAwaitingEvaluation(unschedulablePods, podsRemainUnschedulable, bestOption.Pods)},
			nil
	}

	return &status.ScaleUpStatus{ScaledUp: false, PodsRemainUnschedulable: getRemainingPods(podsRemainUnschedulable)}, nil
}

func getRemainingPods(podSet map[*apiv1.Pod]bool) []*apiv1.Pod {
	result := make([]*apiv1.Pod, 0)
	for pod, val := range podSet {
		if val {
			result = append(result, pod)
		}
	}
	return result
}

func getPodsAwaitingEvaluation(allPods []*apiv1.Pod, unschedulable map[*apiv1.Pod]bool, bestOption []*apiv1.Pod) []*apiv1.Pod {
	result := make(map[*apiv1.Pod]bool, len(allPods))
	for _, pod := range allPods {
		if !unschedulable[pod] {
			result[pod] = true
		}
	}
	for _, pod := range bestOption {
		result[pod] = false
	}
	return getRemainingPods(result)
}

func filterNodeGroupsByPods(groups []cloudprovider.NodeGroup, podsRequiredToFit []*apiv1.Pod,
	fittingPodsPerNodeGroup map[string][]*apiv1.Pod) []cloudprovider.NodeGroup {
	result := make([]cloudprovider.NodeGroup, 0)
groupsloop:
	for _, group := range groups {
		fittingPods, found := fittingPodsPerNodeGroup[group.Id()]
		if !found {
			glog.V(1).Infof("No info about pods passing predicates found for group %v, skipping it from scale-up consideration", group.Id())
			continue
		}
		podSet := make(map[*apiv1.Pod]bool, len(fittingPods))
		for _, pod := range fittingPods {
			podSet[pod] = true
		}
		for _, pod := range podsRequiredToFit {
			if _, found = podSet[pod]; !found {
				glog.V(1).Infof("Group %v, can't fit pod %v/%v, removing from scale-up consideration", group.Id(), pod.Namespace, pod.Name)
				continue groupsloop
			}
		}
		result = append(result, group)
	}
	return result
}

func executeScaleUp(context *context.AutoscalingContext, clusterStateRegistry *clusterstate.ClusterStateRegistry, info nodegroupset.ScaleUpInfo, gpuType string) errors.AutoscalerError {
	glog.V(0).Infof("Scale-up: setting group %s size to %d", info.Group.Id(), info.NewSize)
	increase := info.NewSize - info.CurrentSize
	if err := info.Group.IncreaseSize(increase); err != nil {
		context.LogRecorder.Eventf(apiv1.EventTypeWarning, "FailedToScaleUpGroup", "Scale-up failed for group %s: %v", info.Group.Id(), err)
		clusterStateRegistry.RegisterFailedScaleUp(info.Group.Id(), metrics.APIError)
		return errors.NewAutoscalerError(errors.CloudProviderError,
			"failed to increase node group size: %v", err)
	}
	clusterStateRegistry.RegisterScaleUp(
		&clusterstate.ScaleUpRequest{
			NodeGroupName:   info.Group.Id(),
			Increase:        increase,
			Time:            time.Now(),
			ExpectedAddTime: time.Now().Add(context.MaxNodeProvisionTime),
		})
	metrics.RegisterScaleUp(increase, gpuType)
	context.LogRecorder.Eventf(apiv1.EventTypeNormal, "ScaledUpGroup",
		"Scale-up: group %s size set to %d", info.Group.Id(), info.NewSize)
	return nil
}

func applyScaleUpResourcesLimits(
	newNodes int,
	scaleUpResourcesLeft scaleUpResourcesLimits,
	nodeInfo *schedulercache.NodeInfo,
	nodeGroup cloudprovider.NodeGroup,
	resourceLimiter *cloudprovider.ResourceLimiter) (int, errors.AutoscalerError) {

	delta, err := computeScaleUpResourcesDelta(nodeInfo, nodeGroup, resourceLimiter)
	if err != nil {
		return 0, err
	}

	for resource, resourceDelta := range delta {
		limit, limitFound := scaleUpResourcesLeft[resource]
		if !limitFound {
			continue
		}
		if limit == scaleUpLimitUnknown {
			// should never happen - checked before
			return 0, errors.NewAutoscalerError(
				errors.InternalError,
				fmt.Sprintf("limit unknown for resource %s", resource))
		}
		if int64(newNodes)*resourceDelta <= limit {
			// no capping required
			continue
		}

		newNodes = int(limit / resourceDelta)
		glog.V(1).Infof("Capping scale-up size due to limit for resource %s", resource)
		if newNodes < 1 {
			// should never happen - checked before
			return 0, errors.NewAutoscalerError(
				errors.InternalError,
				fmt.Sprintf("cannot create any node; max limit for resource %s reached", resource))
		}
	}
	return newNodes, nil
}
