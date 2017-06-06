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

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate"
	"k8s.io/autoscaler/cluster-autoscaler/estimator"
	"k8s.io/autoscaler/cluster-autoscaler/expander"
	"k8s.io/autoscaler/cluster-autoscaler/metrics"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/nodegroupset"
	apiv1 "k8s.io/kubernetes/pkg/api/v1"
	extensionsv1 "k8s.io/kubernetes/pkg/apis/extensions/v1beta1"
	"k8s.io/kubernetes/plugin/pkg/scheduler/schedulercache"

	"github.com/golang/glog"
)

// ScaleUp tries to scale the cluster up. Return true if it found a way to increase the size,
// false if it didn't and error if an error occured. Assumes that all nodes in the cluster are
// ready and in sync with instance groups.
func ScaleUp(context *AutoscalingContext, unschedulablePods []*apiv1.Pod, nodes []*apiv1.Node,
	daemonSets []*extensionsv1.DaemonSet) (bool, *errors.AutoscalerError) {
	// From now on we only care about unschedulable pods that were marked after the newest
	// node became available for the scheduler.
	if len(unschedulablePods) == 0 {
		glog.V(1).Info("No unschedulable pods")
		return false, nil
	}

	for _, pod := range unschedulablePods {
		glog.V(1).Infof("Pod %s/%s is unschedulable", pod.Namespace, pod.Name)
	}
	nodeInfos, err := GetNodeInfosForGroups(nodes, context.CloudProvider, context.ClientSet,
		daemonSets, context.PredicateChecker)
	if err != nil {
		return false, err.AddPrefix("failed to build node infos for node groups: ")
	}

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

	for _, nodeGroup := range context.CloudProvider.NodeGroups() {

		if !context.ClusterStateRegistry.IsNodeGroupHealthy(nodeGroup.Id()) {
			glog.Warningf("Node group %s is unhealthy", nodeGroup.Id())
			continue
		}

		currentSize, err := nodeGroup.TargetSize()
		if err != nil {
			glog.Errorf("Failed to get node group size: %v", err)
			continue
		}
		if currentSize >= nodeGroup.MaxSize() {
			// skip this node group.
			glog.V(4).Infof("Skipping node group %s - max size reached", nodeGroup.Id())
			continue
		}

		option := expander.Option{
			NodeGroup: nodeGroup,
			Pods:      make([]*apiv1.Pod, 0),
		}

		nodeInfo, found := nodeInfos[nodeGroup.Id()]
		if !found {
			glog.Errorf("No node info for: %s", nodeGroup.Id())
			continue
		}

		for _, pod := range unschedulablePods {
			err = context.PredicateChecker.CheckPredicates(pod, nodeInfo)
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
			glog.V(4).Info("No pod can fit to %s", nodeGroup.Id())
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

		targetNodeGroups := []cloudprovider.NodeGroup{bestOption.NodeGroup}
		if context.BalanceSimilarNodeGroups {
			similarNodeGroups, typedErr := nodegroupset.FindSimilarNodeGroups(bestOption.NodeGroup, context.CloudProvider, nodeInfos)
			if typedErr != nil {
				return false, typedErr.AddPrefix("Failed to find matching node groups: ")
			}
			similarNodeGroups = filterNodeGroupsByPods(similarNodeGroups, bestOption.Pods, podsPassingPredicates)
			targetNodeGroups = append(targetNodeGroups, similarNodeGroups...)
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

func executeScaleUp(context *AutoscalingContext, info nodegroupset.ScaleUpInfo) *errors.AutoscalerError {
	glog.V(0).Infof("Scale-up: setting group %s size to %d", info.Group.Id(), info.NewSize)
	increase := info.NewSize - info.CurrentSize
	if err := info.Group.IncreaseSize(increase); err != nil {
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
