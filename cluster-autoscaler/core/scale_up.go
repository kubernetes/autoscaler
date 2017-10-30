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
	"time"

	apiv1 "k8s.io/api/core/v1"
	extensionsv1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate"
	"k8s.io/autoscaler/cluster-autoscaler/estimator"
	"k8s.io/autoscaler/cluster-autoscaler/expander"
	"k8s.io/autoscaler/cluster-autoscaler/metrics"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/labels"
	"k8s.io/autoscaler/cluster-autoscaler/utils/nodegroupset"
	"k8s.io/kubernetes/plugin/pkg/scheduler/schedulercache"

	"github.com/golang/glog"
)

// ScaleUp tries to scale the cluster up. Return true if it found a way to increase the size,
// false if it didn't and error if an error occurred. Assumes that all nodes in the cluster are
// ready and in sync with instance groups.
func ScaleUp(context *AutoscalingContext, unschedulablePods []*apiv1.Pod, nodes []*apiv1.Node,
	daemonSets []*extensionsv1.DaemonSet) (bool, errors.AutoscalerError) {
	// From now on we only care about unschedulable pods that were marked after the newest
	// node became available for the scheduler.
	if len(unschedulablePods) == 0 {
		glog.V(1).Info("No unschedulable pods")
		return false, nil
	}

	now := time.Now()

	for _, pod := range unschedulablePods {
		glog.V(1).Infof("Pod %s/%s is unschedulable", pod.Namespace, pod.Name)
	}
	nodeInfos, err := GetNodeInfosForGroups(nodes, context.CloudProvider, context.ClientSet,
		daemonSets, context.PredicateChecker)
	if err != nil {
		return false, err.AddPrefix("failed to build node infos for node groups: ")
	}

	nodeGroups := context.CloudProvider.NodeGroups()

	resourceLimiter, errCP := context.CloudProvider.GetResourceLimiter()
	if errCP != nil {
		return false, errors.ToAutoscalerError(
			errors.CloudProviderError,
			errCP)
	}
	// calculate current cores & gigabytes of memory
	coresTotal, memoryTotal := calculateClusterCoresMemoryTotal(nodeGroups, nodeInfos)

	upcomingNodes := make([]*schedulercache.NodeInfo, 0)
	for nodeGroup, numberOfNodes := range context.ClusterStateRegistry.GetUpcomingNodes() {
		nodeTemplate, found := nodeInfos[nodeGroup]
		if !found {
			return false, errors.NewAutoscalerError(
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
	podsRemainUnschedulable := make(map[*apiv1.Pod]bool)
	expansionOptions := make([]expander.Option, 0)

	if context.AutoscalingOptions.NodeAutoprovisioningEnabled {
		nodeGroups, nodeInfos = addAutoprovisionedCandidates(context, nodeGroups, nodeInfos, unschedulablePods)
	}

	for _, nodeGroup := range nodeGroups {
		// Autoprovisioned node groups without nodes are created later so skip check for them.
		if nodeGroup.Exist() && !context.ClusterStateRegistry.IsNodeGroupSafeToScaleUp(nodeGroup.Id(), now) {
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

		nodeCPU, nodeMemory, err := getNodeInfoCoresAndMemory(nodeInfo)
		if err != nil {
			glog.Errorf("Failed to get node resources: %v", err)
		}
		if nodeCPU > (resourceLimiter.GetMax(cloudprovider.ResourceNameCores) - coresTotal) {
			// skip this node group
			glog.V(4).Infof("Skipping node group %s - not enough cores limit left", nodeGroup.Id())
			continue
		}
		if nodeMemory > (resourceLimiter.GetMax(cloudprovider.ResourceNameMemory) - memoryTotal) {
			// skip this node group
			glog.V(4).Infof("Skipping node group %s - not enough memory limit left", nodeGroup.Id())
			continue
		}

		option := expander.Option{
			NodeGroup: nodeGroup,
			Pods:      make([]*apiv1.Pod, 0),
		}

		for _, pod := range unschedulablePods {
			err = context.PredicateChecker.CheckPredicates(pod, nil, nodeInfo, simulator.ReturnVerboseError)
			if err == nil {
				option.Pods = append(option.Pods, pod)
				podsRemainUnschedulable[pod] = false
			} else {
				glog.V(2).Infof("Scale-up predicate failed: %v", err)
				if _, exists := podsRemainUnschedulable[pod]; !exists {
					podsRemainUnschedulable[pod] = true
				}
			}
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
		for pod, unschedulable := range podsRemainUnschedulable {
			if unschedulable {
				context.Recorder.Event(pod, apiv1.EventTypeNormal, "NotTriggerScaleUp",
					"pod didn't trigger scale-up (it wouldn't fit if a new node is added)")
			}
		}
		return false, nil
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
				return false, errors.NewAutoscalerError(
					errors.TransientError,
					"max node total count already reached")
			}
		}
		if context.AutoscalingOptions.NodeAutoprovisioningEnabled {
			if !bestOption.NodeGroup.Exist() {
				// Node group id may change when we create node group and we need to update
				// our data structures
				oldId := bestOption.NodeGroup.Id()
				err := bestOption.NodeGroup.Create()
				if err != nil {
					context.LogRecorder.Eventf(apiv1.EventTypeWarning, "FailedToCreateNodeGroup",
						"NodeAutoprovisioning: attempt to create node group %v failed: %v", oldId, err)
					return false, errors.ToAutoscalerError(errors.CloudProviderError, err)
				}
				newId := bestOption.NodeGroup.Id()
				if newId != oldId {
					glog.V(2).Infof("Created node group %s based on template node group %s, will use new node group in scale-up", newId, oldId)
					podsPassingPredicates[newId] = podsPassingPredicates[oldId]
					delete(podsPassingPredicates, oldId)
					nodeInfos[newId] = nodeInfos[oldId]
					delete(nodeInfos, oldId)
				}
				context.LogRecorder.Eventf(apiv1.EventTypeNormal, "CreatedNodeGroup",
					"NodeAutoprovisioning: created new node group %v", newId)

			}
		}

		nodeInfo, found := nodeInfos[bestOption.NodeGroup.Id()]
		if !found {
			// This should never happen, as we already should have retrieved
			// nodeInfo for any considered nodegroup.
			glog.Errorf("No node info for: %s", bestOption.NodeGroup.Id())
			return false, errors.NewAutoscalerError(
				errors.CloudProviderError,
				"No node info for best expansion option!")
		}

		// apply upper limits for CPU and memory
		newNodes, err = applyMaxClusterCoresMemoryLimits(newNodes, coresTotal, memoryTotal, resourceLimiter.GetMax(cloudprovider.ResourceNameCores), resourceLimiter.GetMax(cloudprovider.ResourceNameMemory), nodeInfo)
		if err != nil {
			return false, err
		}

		targetNodeGroups := []cloudprovider.NodeGroup{bestOption.NodeGroup}
		if context.BalanceSimilarNodeGroups {
			similarNodeGroups, typedErr := nodegroupset.FindSimilarNodeGroups(bestOption.NodeGroup, context.CloudProvider, nodeInfos)
			if typedErr != nil {
				return false, typedErr.AddPrefix("Failed to find matching node groups: ")
			}
			similarNodeGroups = filterNodeGroupsByPods(similarNodeGroups, bestOption.Pods, podsPassingPredicates)
			for _, ng := range similarNodeGroups {
				if context.ClusterStateRegistry.IsNodeGroupSafeToScaleUp(ng.Id(), now) {
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
			return false, typedErr
		}
		glog.V(1).Infof("Final scale-up plan: %v", scaleUpInfos)
		for _, info := range scaleUpInfos {
			typedErr := executeScaleUp(context, info)
			if typedErr != nil {
				return false, typedErr
			}
		}

		for _, pod := range bestOption.Pods {
			context.Recorder.Eventf(pod, apiv1.EventTypeNormal, "TriggeredScaleUp",
				"pod triggered scale-up: %v", scaleUpInfos)
		}

		context.ClusterStateRegistry.Recalculate()
		return true, nil
	}
	for pod, unschedulable := range podsRemainUnschedulable {
		if unschedulable {
			context.Recorder.Event(pod, apiv1.EventTypeNormal, "NotTriggerScaleUp",
				"pod didn't trigger scale-up (it wouldn't fit if a new node is added)")
		}
	}

	return false, nil
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

func executeScaleUp(context *AutoscalingContext, info nodegroupset.ScaleUpInfo) errors.AutoscalerError {
	glog.V(0).Infof("Scale-up: setting group %s size to %d", info.Group.Id(), info.NewSize)
	increase := info.NewSize - info.CurrentSize
	if err := info.Group.IncreaseSize(increase); err != nil {
		context.LogRecorder.Eventf(apiv1.EventTypeWarning, "FailedToScaleUpGroup", "Scale-up failed for group %s: %v", info.Group.Id(), err)
		context.ClusterStateRegistry.RegisterFailedScaleUp(info.Group.Id(), metrics.APIError)
		return errors.NewAutoscalerError(errors.CloudProviderError,
			"failed to increase node group size: %v", err)
	}
	context.ClusterStateRegistry.RegisterScaleUp(
		&clusterstate.ScaleUpRequest{
			NodeGroupName:   info.Group.Id(),
			Increase:        increase,
			Time:            time.Now(),
			ExpectedAddTime: time.Now().Add(context.MaxNodeProvisionTime),
		})
	metrics.RegisterScaleUp(increase)
	context.LogRecorder.Eventf(apiv1.EventTypeNormal, "ScaledUpGroup",
		"Scale-up: group %s size set to %d", info.Group.Id(), info.NewSize)
	return nil
}

func addAutoprovisionedCandidates(context *AutoscalingContext, nodeGroups []cloudprovider.NodeGroup,
	nodeInfos map[string]*schedulercache.NodeInfo, unschedulablePods []*apiv1.Pod) ([]cloudprovider.NodeGroup,
	map[string]*schedulercache.NodeInfo) {

	autoprovisionedNodeGroupCount := 0
	for _, group := range nodeGroups {
		if group.Autoprovisioned() {
			autoprovisionedNodeGroupCount++
		}
	}
	if autoprovisionedNodeGroupCount >= context.MaxAutoprovisionedNodeGroupCount {
		glog.V(4).Infof("Max autoprovisioned node group count reached")
		return nodeGroups, nodeInfos
	}

	machines, err := context.CloudProvider.GetAvailableMachineTypes()
	if err != nil {
		glog.Warningf("Failed to get machine types: %v", err)
	} else {
		bestLabels := labels.BestLabelSet(unschedulablePods)
		for _, machineType := range machines {
			nodeGroup, err := context.CloudProvider.NewNodeGroup(machineType, bestLabels, nil)
			if err != nil {
				glog.Warningf("Unable to build temporary node group for %s: %v", machineType, err)
				continue
			}
			nodeInfo, err := nodeGroup.TemplateNodeInfo()
			if err != nil {
				glog.Warningf("Unable to build template for node group for %s: %v", nodeGroup.Id(), err)
				continue
			}
			nodeInfos[nodeGroup.Id()] = nodeInfo
			nodeGroups = append(nodeGroups, nodeGroup)
		}
	}
	return nodeGroups, nodeInfos
}

func calculateClusterCoresMemoryTotal(nodeGroups []cloudprovider.NodeGroup, nodeInfos map[string]*schedulercache.NodeInfo) (int64, int64) {
	var coresTotal int64
	var memoryTotal int64
	for _, nodeGroup := range nodeGroups {
		currentSize, err := nodeGroup.TargetSize()
		if err != nil {
			glog.Errorf("Failed to get node group size of %v: %v", nodeGroup.Id(), err)
			continue
		}
		nodeInfo, found := nodeInfos[nodeGroup.Id()]
		if !found {
			glog.Errorf("No node info for: %s", nodeGroup.Id())
			continue
		}
		if currentSize > 0 {
			nodeCPU, nodeMemory, err := getNodeInfoCoresAndMemory(nodeInfo)
			if err != nil {
				glog.Errorf("Failed to get node resources: %v", err)
				continue
			}
			coresTotal = coresTotal + int64(currentSize)*nodeCPU
			memoryTotal = memoryTotal + int64(currentSize)*nodeMemory
		}
	}

	return coresTotal, memoryTotal
}

func applyMaxClusterCoresMemoryLimits(newNodes int, coresTotal, memoryTotal, maxCoresTotal, maxMemoryTotal int64, nodeInfo *schedulercache.NodeInfo) (int, errors.AutoscalerError) {
	newNodeCPU, newNodeMemory, err := getNodeInfoCoresAndMemory(nodeInfo)
	if err != nil {
		// This is not very elegant, but it allows us to proceed even if we're
		// unable to compute cpu/memory limits (not breaking current functionality)
		glog.Errorf("Failed to get node resources: %v", err)
		return newNodes, nil
	}
	if coresTotal+newNodeCPU*int64(newNodes) > maxCoresTotal {
		glog.V(1).Infof("Capping size to max cluster cores (%d)", maxCoresTotal)
		newNodes = int((maxCoresTotal - coresTotal) / newNodeCPU)
		if newNodes < 1 {
			// This should never happen, as we already check that
			// at least one node will fit when considering nodegroup
			return 0, errors.NewAutoscalerError(
				errors.TransientError,
				"max cores already reached")
		}
	}
	if memoryTotal+newNodeMemory*int64(newNodes) > maxMemoryTotal {
		glog.V(1).Infof("Capping size to max cluster memory allowed (%d)", maxMemoryTotal)
		newNodes = int((maxMemoryTotal - memoryTotal) / newNodeMemory)
		if newNodes < 1 {
			// This should never happen, as we already check that
			// at least one node will fit when considering nodegroup
			return 0, errors.NewAutoscalerError(
				errors.TransientError,
				"max memory already reached")
		}
	}
	return newNodes, nil
}

func getNodeInfoCoresAndMemory(nodeInfo *schedulercache.NodeInfo) (int64, int64, error) {
	return getNodeCoresAndMemory(nodeInfo.Node())
}
