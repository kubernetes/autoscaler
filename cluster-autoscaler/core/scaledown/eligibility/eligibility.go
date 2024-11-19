/*
Copyright 2022 The Kubernetes Authors.

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

package eligibility

import (
	"reflect"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/actuation"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/unremovable"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/utilization"
	"k8s.io/autoscaler/cluster-autoscaler/utils/klogx"

	apiv1 "k8s.io/api/core/v1"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	"k8s.io/klog/v2"
)

const (
	// ScaleDownDisabledKey is the name of annotation marking node as not eligible for scale down.
	ScaleDownDisabledKey = "cluster-autoscaler.kubernetes.io/scale-down-disabled"
)

// Checker is responsible for deciding which nodes pass the criteria for scale down.
type Checker struct {
	configGetter nodeGroupConfigGetter
}

type nodeGroupConfigGetter interface {
	// GetScaleDownUtilizationThreshold returns ScaleDownUtilizationThreshold value that should be used for a given NodeGroup.
	GetScaleDownUtilizationThreshold(nodeGroup cloudprovider.NodeGroup) (float64, error)
	// GetScaleDownGpuUtilizationThreshold returns ScaleDownGpuUtilizationThreshold value that should be used for a given NodeGroup.
	GetScaleDownGpuUtilizationThreshold(nodeGroup cloudprovider.NodeGroup) (float64, error)
	// GetIgnoreDaemonSetsUtilization returns IgnoreDaemonSetsUtilization value that should be used for a given NodeGroup.
	GetIgnoreDaemonSetsUtilization(nodeGroup cloudprovider.NodeGroup) (bool, error)
}

// NewChecker creates a new Checker object.
func NewChecker(configGetter nodeGroupConfigGetter) *Checker {
	return &Checker{
		configGetter: configGetter,
	}
}

// FilterOutUnremovable accepts a list of nodes that are candidates for
// scale down and filters out nodes that cannot be removed, along with node
// utilization info.
// TODO(x13n): Node utilization could actually be calculated independently for
// all nodes and just used here. Next refactor...
func (c *Checker) FilterOutUnremovable(context *context.AutoscalingContext, scaleDownCandidates []*apiv1.Node, timestamp time.Time, unremovableNodes *unremovable.Nodes) ([]string, map[string]utilization.Info, []*simulator.UnremovableNode) {
	ineligible := []*simulator.UnremovableNode{}
	skipped := 0
	utilizationMap := make(map[string]utilization.Info)
	currentlyUnneededNodeNames := make([]string, 0, len(scaleDownCandidates))
	utilLogsQuota := klogx.NewLoggingQuota(20)

	for _, node := range scaleDownCandidates {
		nodeInfo, err := context.ClusterSnapshot.GetNodeInfo(node.Name)
		if err != nil {
			klog.Errorf("Can't retrieve scale-down candidate %s from snapshot, err: %v", node.Name, err)
			ineligible = append(ineligible, &simulator.UnremovableNode{Node: node, Reason: simulator.UnexpectedError})
			continue
		}

		// Skip nodes that were recently checked.
		if unremovableNodes.IsRecent(node.Name) {
			ineligible = append(ineligible, &simulator.UnremovableNode{Node: node, Reason: simulator.RecentlyUnremovable})
			skipped++
			continue
		}

		reason, utilInfo := c.unremovableReasonAndNodeUtilization(context, timestamp, nodeInfo, utilLogsQuota)
		if utilInfo != nil {
			utilizationMap[node.Name] = *utilInfo
		}
		if reason != simulator.NoReason {
			ineligible = append(ineligible, &simulator.UnremovableNode{Node: node, Reason: reason})
			continue
		}

		currentlyUnneededNodeNames = append(currentlyUnneededNodeNames, node.Name)
	}

	klogx.V(4).Over(utilLogsQuota).Infof("Skipped logging utilization for %d other nodes", -utilLogsQuota.Left())
	if skipped > 0 {
		klog.V(1).Infof("Scale-down calculation: ignoring %v nodes unremovable in the last %v", skipped, context.AutoscalingOptions.UnremovableNodeRecheckTimeout)
	}
	return currentlyUnneededNodeNames, utilizationMap, ineligible
}

func (c *Checker) unremovableReasonAndNodeUtilization(context *context.AutoscalingContext, timestamp time.Time, nodeInfo *framework.NodeInfo, utilLogsQuota *klogx.Quota) (simulator.UnremovableReason, *utilization.Info) {
	node := nodeInfo.Node()

	if actuation.IsNodeBeingDeleted(node, timestamp) {
		klog.V(1).Infof("Skipping %s from delete consideration - the node is currently being deleted", node.Name)
		return simulator.CurrentlyBeingDeleted, nil
	}

	// Skip nodes marked with no scale down annotation
	if HasNoScaleDownAnnotation(node) {
		klog.V(1).Infof("Skipping %s from delete consideration - the node is marked as no scale down", node.Name)
		return simulator.ScaleDownDisabledAnnotation, nil
	}

	nodeGroup, err := context.CloudProvider.NodeGroupForNode(node)
	if err != nil {
		klog.Warningf("Node group not found for node %v: %v", node.Name, err)
		return simulator.UnexpectedError, nil
	}
	if nodeGroup == nil || reflect.ValueOf(nodeGroup).IsNil() {
		// We should never get here as non-autoscaled nodes should not be included in scaleDownCandidates list
		// (and the default PreFilteringScaleDownNodeProcessor would indeed filter them out).
		klog.Warningf("Skipped %s from delete consideration - the node is not autoscaled", node.Name)
		return simulator.NotAutoscaled, nil
	}

	ignoreDaemonSetsUtilization, err := c.configGetter.GetIgnoreDaemonSetsUtilization(nodeGroup)
	if err != nil {
		klog.Warningf("Couldn't retrieve `IgnoreDaemonSetsUtilization` option for node %v: %v", node.Name, err)
		return simulator.UnexpectedError, nil
	}

	gpuConfig := context.CloudProvider.GetNodeGpuConfig(node)
	utilInfo, err := utilization.Calculate(nodeInfo, ignoreDaemonSetsUtilization, context.IgnoreMirrorPodsUtilization, gpuConfig, timestamp)
	if err != nil {
		klog.Warningf("Failed to calculate utilization for %s: %v", node.Name, err)
	}

	// If scale down of unready nodes is disabled, skip the node if it is unready
	if !context.ScaleDownUnreadyEnabled {
		ready, _, _ := kube_util.GetReadinessState(node)
		if !ready {
			klog.V(4).Infof("Skipping unready node %s from delete consideration - scale-down of unready nodes is disabled", node.Name)
			return simulator.ScaleDownUnreadyDisabled, nil
		}
	}

	underutilized, err := c.isNodeBelowUtilizationThreshold(context, node, nodeGroup, utilInfo)
	if err != nil {
		klog.Warningf("Failed to check utilization thresholds for %s: %v", node.Name, err)
		return simulator.UnexpectedError, nil
	}
	if !underutilized {
		klog.V(4).Infof("Node %s unremovable: %s requested (%.6g%% of allocatable) is above the scale-down utilization threshold", node.Name, utilInfo.ResourceName, utilInfo.Utilization*100)
		return simulator.NotUnderutilized, &utilInfo
	}

	klogx.V(4).UpTo(utilLogsQuota).Infof("Node %s - %s requested is %.6g%% of allocatable", node.Name, utilInfo.ResourceName, utilInfo.Utilization*100)

	return simulator.NoReason, &utilInfo
}

// isNodeBelowUtilizationThreshold determines if a given node utilization is below threshold.
func (c *Checker) isNodeBelowUtilizationThreshold(context *context.AutoscalingContext, node *apiv1.Node, nodeGroup cloudprovider.NodeGroup, utilInfo utilization.Info) (bool, error) {
	var threshold float64
	var err error
	gpuConfig := context.CloudProvider.GetNodeGpuConfig(node)
	if gpuConfig != nil {
		threshold, err = c.configGetter.GetScaleDownGpuUtilizationThreshold(nodeGroup)
		if err != nil {
			return false, err
		}
	} else {
		threshold, err = c.configGetter.GetScaleDownUtilizationThreshold(nodeGroup)
		if err != nil {
			return false, err
		}
	}
	if utilInfo.Utilization >= threshold {
		return false, nil
	}
	return true, nil
}

// HasNoScaleDownAnnotation checks whether the node has an annotation blocking it from being scaled down.
func HasNoScaleDownAnnotation(node *apiv1.Node) bool {
	return node.Annotations[ScaleDownDisabledKey] == "true"
}
