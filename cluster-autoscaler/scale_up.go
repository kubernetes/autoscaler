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

package main

import (
	"fmt"

	"k8s.io/contrib/cluster-autoscaler/estimator"
	"k8s.io/contrib/cluster-autoscaler/expander"
	apiv1 "k8s.io/kubernetes/pkg/api/v1"

	"github.com/golang/glog"
)

// ScaleUp tries to scale the cluster up. Return true if it found a way to increase the size,
// false if it didn't and error if an error occured. Assumes that all nodes in the cluster are
// ready and in sync with instance groups.
func ScaleUp(context *AutoscalingContext, unschedulablePods []*apiv1.Pod, nodes []*apiv1.Node) (bool, error) {
	// From now on we only care about unschedulable pods that were marked after the newest
	// node became available for the scheduler.
	if len(unschedulablePods) == 0 {
		glog.V(1).Info("No unschedulable pods")
		return false, nil
	}

	for _, pod := range unschedulablePods {
		glog.V(1).Infof("Pod %s/%s is unschedulable", pod.Namespace, pod.Name)
	}

	nodeInfos, err := GetNodeInfosForGroups(nodes, context.CloudProvider, context.ClientSet)
	expansionOptions := make([]expander.Option, 0)
	if err != nil {
		return false, fmt.Errorf("failed to build node infos for node groups: %v", err)
	}

	podsRemainUnshedulable := make(map[*apiv1.Pod]struct{})
	for _, nodeGroup := range context.CloudProvider.NodeGroups() {

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
			} else {
				glog.V(2).Infof("Scale-up predicate failed: %v", err)
				podsRemainUnshedulable[pod] = struct{}{}
			}
		}
		if len(option.Pods) > 0 {
			if context.EstimatorName == BinpackingEstimatorName {
				binpackingEstimator := estimator.NewBinpackingNodeEstimator(context.PredicateChecker)
				option.NodeCount = binpackingEstimator.Estimate(option.Pods, nodeInfo)
			} else if context.EstimatorName == BasicEstimatorName {
				basicEstimator := estimator.NewBasicNodeEstimator()
				for _, pod := range option.Pods {
					basicEstimator.Add(pod)
				}
				option.NodeCount, option.Debug = basicEstimator.Estimate(nodeInfo.Node())
			} else {
				glog.Fatalf("Unrecognized estimator: %s", context.EstimatorName)
			}
			expansionOptions = append(expansionOptions, option)
		}
	}

	if len(expansionOptions) == 0 {
		glog.V(1).Info("No node group can help with pending pods.")
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

		currentSize, err := bestOption.NodeGroup.TargetSize()
		if err != nil {
			return false, fmt.Errorf("failed to get node group size: %v", err)
		}
		newSize := currentSize + bestOption.NodeCount
		if newSize >= bestOption.NodeGroup.MaxSize() {
			glog.V(1).Infof("Capping size to MAX (%d)", bestOption.NodeGroup.MaxSize())
			newSize = bestOption.NodeGroup.MaxSize()
		}

		if context.MaxNodesTotal > 0 && len(nodes)+(newSize-currentSize) > context.MaxNodesTotal {
			glog.V(1).Infof("Capping size to max cluster total size (%d)", context.MaxNodesTotal)
			newSize = context.MaxNodesTotal - len(nodes) + currentSize
			if newSize < currentSize {
				return false, fmt.Errorf("max node total count already reached")
			}
		}

		glog.V(0).Infof("Scale-up: setting group %s size to %d", bestOption.NodeGroup.Id(), newSize)

		if err := bestOption.NodeGroup.IncreaseSize(newSize - currentSize); err != nil {
			return false, fmt.Errorf("failed to increase node group size: %v", err)
		}

		for _, pod := range bestOption.Pods {
			context.Recorder.Eventf(pod, apiv1.EventTypeNormal, "TriggeredScaleUp",
				"pod triggered scale-up, group: %s, sizes (current/new): %d/%d", bestOption.NodeGroup.Id(), currentSize, newSize)
		}

		return true, nil
	}
	for pod := range podsRemainUnshedulable {
		context.Recorder.Event(pod, apiv1.EventTypeNormal, "NotTriggerScaleUp",
			"pod didn't trigger scale-up (it wouldn't fit if a new node is added)")
	}

	return false, nil
}
