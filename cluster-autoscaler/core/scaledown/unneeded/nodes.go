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

package unneeded

import (
	"reflect"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/eligibility"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/resource"
	"k8s.io/autoscaler/cluster-autoscaler/metrics"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	"k8s.io/autoscaler/cluster-autoscaler/utils"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"

	apiv1 "k8s.io/api/core/v1"
	klog "k8s.io/klog/v2"
)

// Nodes tracks the state of cluster nodes that are not needed.
type Nodes struct {
	sdtg         scaleDownTimeGetter
	limitsFinder *resource.LimitsFinder
	cachedList   []*apiv1.Node
	byName       map[string]*node
}

type node struct {
	ntbr  simulator.NodeToBeRemoved
	since time.Time
}

type scaleDownTimeGetter interface {
	// GetScaleDownUnneededTime returns ScaleDownUnneededTime value that should be used for a given NodeGroup.
	GetScaleDownUnneededTime(context *context.AutoscalingContext, nodeGroup cloudprovider.NodeGroup) (time.Duration, error)
	// GetScaleDownUnreadyTime returns ScaleDownUnreadyTime value that should be used for a given NodeGroup.
	GetScaleDownUnreadyTime(context *context.AutoscalingContext, nodeGroup cloudprovider.NodeGroup) (time.Duration, error)
}

// NewNodes returns a new initialized Nodes object.
func NewNodes(sdtg scaleDownTimeGetter, limitsFinder *resource.LimitsFinder) *Nodes {
	return &Nodes{
		sdtg:         sdtg,
		limitsFinder: limitsFinder,
	}
}

// Update stores nodes along with a time at which they were found to be
// unneeded. Previously existing timestamps are preserved.
func (n *Nodes) Update(nodes []simulator.NodeToBeRemoved, ts time.Time) {
	updated := make(map[string]*node, len(nodes))
	for _, nn := range nodes {
		name := nn.Node.Name
		updated[name] = &node{
			ntbr: nn,
		}
		if val, found := n.byName[name]; found {
			updated[name].since = val.since
		} else {
			updated[name].since = ts
		}
	}
	n.byName = updated
	n.cachedList = nil
	if klog.V(4).Enabled() {
		for k, v := range n.byName {
			klog.Infof("%s is unneeded since %s duration %s", k, v.since, ts.Sub(v.since))
		}
	}
}

// Clear resets the internal state, dropping information about all tracked nodes.
func (n *Nodes) Clear() {
	n.Update(nil, time.Time{})
}

// Contains returns true iff a given node is unneeded.
func (n *Nodes) Contains(nodeName string) bool {
	_, found := n.byName[nodeName]
	return found
}

// AsList returns a slice of unneeded Node objects.
func (n *Nodes) AsList() []*apiv1.Node {
	if n.cachedList == nil {
		n.cachedList = make([]*apiv1.Node, 0, len(n.byName))
		for _, v := range n.byName {
			n.cachedList = append(n.cachedList, v.ntbr.Node)
		}
	}
	return n.cachedList
}

// Drop stops tracking a specified node.
func (n *Nodes) Drop(node string) {
	delete(n.byName, node)
	n.cachedList = nil
}

// RemovableAt returns all nodes that can be removed at a given time, divided
// into empty and non-empty node lists, as well as a list of nodes that were
// unneeded, but are not removable, annotated by reason.
func (n *Nodes) RemovableAt(context *context.AutoscalingContext, ts time.Time, resourcesLeft resource.Limits, resourcesWithLimits []string, as scaledown.ActuationStatus) (empty, needDrain []simulator.NodeToBeRemoved, unremovable []*simulator.UnremovableNode) {
	nodeGroupSize := utils.GetNodeGroupSizeMap(context.CloudProvider)
	resourcesLeftCopy := resourcesLeft.DeepCopy()
	emptyNodes, drainNodes := n.splitEmptyAndNonEmptyNodes()

	for nodeName, v := range emptyNodes {
		klog.V(2).Infof("%s was unneeded for %s", nodeName, ts.Sub(v.since).String())
		if r := n.unremovableReason(context, v, ts, nodeGroupSize, resourcesLeftCopy, resourcesWithLimits, as); r != simulator.NoReason {
			unremovable = append(unremovable, &simulator.UnremovableNode{Node: v.ntbr.Node, Reason: r})
			continue
		}
		empty = append(empty, v.ntbr)
	}
	for nodeName, v := range drainNodes {
		klog.V(2).Infof("%s was unneeded for %s", nodeName, ts.Sub(v.since).String())
		if r := n.unremovableReason(context, v, ts, nodeGroupSize, resourcesLeftCopy, resourcesWithLimits, as); r != simulator.NoReason {
			unremovable = append(unremovable, &simulator.UnremovableNode{Node: v.ntbr.Node, Reason: r})
			continue
		}
		needDrain = append(needDrain, v.ntbr)
	}
	return
}

func (n *Nodes) unremovableReason(context *context.AutoscalingContext, v *node, ts time.Time, nodeGroupSize map[string]int, resourcesLeft resource.Limits, resourcesWithLimits []string, as scaledown.ActuationStatus) simulator.UnremovableReason {
	node := v.ntbr.Node
	// Check if node is marked with no scale down annotation.
	if eligibility.HasNoScaleDownAnnotation(node) {
		klog.V(4).Infof("Skipping %s - scale down disabled annotation found", node.Name)
		return simulator.ScaleDownDisabledAnnotation
	}
	ready, _, _ := kube_util.GetReadinessState(node)

	nodeGroup, err := context.CloudProvider.NodeGroupForNode(node)
	if err != nil {
		klog.Errorf("Error while checking node group for %s: %v", node.Name, err)
		return simulator.UnexpectedError
	}
	if nodeGroup == nil || reflect.ValueOf(nodeGroup).IsNil() {
		klog.V(4).Infof("Skipping %s - no node group config", node.Name)
		return simulator.NotAutoscaled
	}

	if ready {
		// Check how long a ready node was underutilized.
		unneededTime, err := n.sdtg.GetScaleDownUnneededTime(context, nodeGroup)
		if err != nil {
			klog.Errorf("Error trying to get ScaleDownUnneededTime for node %s (in group: %s)", node.Name, nodeGroup.Id())
			return simulator.UnexpectedError
		}
		if !v.since.Add(unneededTime).Before(ts) {
			return simulator.NotUnneededLongEnough
		}
	} else {
		// Unready nodes may be deleted after a different time than underutilized nodes.
		unreadyTime, err := n.sdtg.GetScaleDownUnreadyTime(context, nodeGroup)
		if err != nil {
			klog.Errorf("Error trying to get ScaleDownUnreadyTime for node %s (in group: %s)", node.Name, nodeGroup.Id())
			return simulator.UnexpectedError
		}
		if !v.since.Add(unreadyTime).Before(ts) {
			return simulator.NotUnreadyLongEnough
		}
	}

	if reason := verifyMinSize(node.Name, nodeGroup, nodeGroupSize, as); reason != simulator.NoReason {
		return reason
	}

	resourceDelta, err := n.limitsFinder.DeltaForNode(context, node, nodeGroup, resourcesWithLimits)
	if err != nil {
		klog.Errorf("Error getting node resources: %v", err)
		return simulator.UnexpectedError
	}

	checkResult := resourcesLeft.TryDecrementBy(resourceDelta)
	if checkResult.Exceeded() {
		klog.V(4).Infof("Skipping %s - minimal limit exceeded for %v", node.Name, checkResult.ExceededResources)
		for _, resource := range checkResult.ExceededResources {
			switch resource {
			case cloudprovider.ResourceNameCores:
				metrics.RegisterSkippedScaleDownCPU()
			case cloudprovider.ResourceNameMemory:
				metrics.RegisterSkippedScaleDownMemory()
			default:
				continue
			}
		}
		return simulator.MinimalResourceLimitExceeded
	}

	nodeGroupSize[nodeGroup.Id()]--
	return simulator.NoReason
}

func (n *Nodes) splitEmptyAndNonEmptyNodes() (empty, needDrain map[string]*node) {
	empty = make(map[string]*node)
	needDrain = make(map[string]*node)
	for name, v := range n.byName {
		if len(v.ntbr.PodsToReschedule) > 0 {
			needDrain[name] = v
		} else {
			empty[name] = v
		}
	}
	return
}

func verifyMinSize(nodeName string, nodeGroup cloudprovider.NodeGroup, nodeGroupSize map[string]int, as scaledown.ActuationStatus) simulator.UnremovableReason {
	size, found := nodeGroupSize[nodeGroup.Id()]
	if !found {
		klog.Errorf("Error while checking node group size %s: group size not found in cache", nodeGroup.Id())
		return simulator.UnexpectedError
	}
	deletionsInProgress := as.DeletionsCount(nodeGroup.Id())
	if size-deletionsInProgress <= nodeGroup.MinSize() {
		klog.V(1).Infof("Skipping %s - node group min size reached", nodeName)
		return simulator.NodeGroupMinSizeReached
	}
	return simulator.NoReason
}
