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
	"context"
	"fmt"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	ca_context "k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/eligibility"

	"k8s.io/autoscaler/cluster-autoscaler/processors/nodes"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	"k8s.io/autoscaler/cluster-autoscaler/utils"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"

	klog "k8s.io/klog/v2"
)

// Nodes tracks the state of cluster nodes that are not needed.
type Nodes struct {
	sdtg scaleDownTimeGetter

	cachedList []*scaledown.UnneededNode
	byName     map[string]*node
}

type node struct {
	ntbr                     simulator.NodeToBeRemoved
	since                    time.Time
	removalThreshold         time.Duration
	thresholdRetrievalFailed bool
}

type scaleDownTimeGetter interface {
	// GetScaleDownUnneededTime returns ScaleDownUnneededTime value that should be used for a given NodeGroup.
	GetScaleDownUnneededTime(nodeGroup cloudprovider.NodeGroup) (time.Duration, error)
	// GetScaleDownUnreadyTime returns ScaleDownUnreadyTime value that should be used for a given NodeGroup.
	GetScaleDownUnreadyTime(nodeGroup cloudprovider.NodeGroup) (time.Duration, error)
}

// NewNodes returns a new initialized Nodes object.
func NewNodes(sdtg scaleDownTimeGetter) *Nodes {
	return &Nodes{
		sdtg: sdtg,
	}
}

// LoadFromExistingTaints loads any existing DeletionCandidateTaint taints from the kubernetes cluster. given a TTL for the taint
func (n *Nodes) LoadFromExistingTaints(ctx context.Context, autoscalingCtx *ca_context.AutoscalingContext, ts time.Time) error {
	logger := klog.FromContext(ctx)
	allNodes, err := autoscalingCtx.ListerRegistry.AllNodeLister().List()
	if err != nil {
		return fmt.Errorf("failed to list nodes when initializing unneeded nodes: %v", err)
	}

	var deletionCandidateStalenessTTL = autoscalingCtx.AutoscalingOptions.NodeDeletionCandidateTTL
	var nodesWithTaints []simulator.NodeToBeRemoved
	for _, node := range allNodes {
		since, err := taints.GetDeletionCandidateTime(node)
		if err != nil {
			logger.Error(err, "Failed to get pods to move for node", "node", node.Name)
			continue
		}
		if since == nil {
			continue
		}
		if since.Add(deletionCandidateStalenessTTL).Before(ts) {
			logger.V(4).Info("Skipping node with deletion candidate taint , since it is older than TTL", "node", node.Name, "string", since.String(), "string", deletionCandidateStalenessTTL.String())
			continue
		}
		nodeToBeRemoved := simulator.NodeToBeRemoved{
			Node: node,
		}
		nodesWithTaints = append(nodesWithTaints, nodeToBeRemoved)
		logger.V(4).Info("Found node with deletion candidate taint", "node", node.Name, "string", since.String())
	}

	if len(nodesWithTaints) > 0 {
		logger.V(1).Info("Initializing unneeded nodes nodes that have deletion candidate taints", "nodesWithTaintsCount", len(nodesWithTaints))
		n.initialize(ctx, autoscalingCtx, nodesWithTaints, ts)
	}

	return nil
}

// initialize initializes the Nodes object with the given node list.
// It sets the initial state of unneeded nodes reflect the taint status of nodes in the cluster.
// This is in order the avoid state loss between deployment restarts.
func (n *Nodes) initialize(ctx context.Context, autoscalingCtx *ca_context.AutoscalingContext, nodes []simulator.NodeToBeRemoved, ts time.Time) {
	logger := klog.FromContext(ctx)
	n.updateInternalState(ctx, autoscalingCtx, nodes, ts, func(nn simulator.NodeToBeRemoved) *time.Time {
		name := nn.Node.Name
		if since, err := taints.GetDeletionCandidateTime(nn.Node); err == nil {
			logger.V(4).Info("Found node with deletion candidate taint", "name", name, "string", since.String())
			return since
		} else if since == nil {
			logger.Error(err, "Failed to get deletion candidate taint time for node", "name", name)
			return nil
		}
		logger.V(4).Info("Found node with deletion candidate taint from now", "name", name)
		return nil
	})
}

// Update stores nodes along with a time at which they were found to be
// unneeded. Previously existing timestamps are preserved.
func (n *Nodes) Update(ctx context.Context, autoscalingCtx *ca_context.AutoscalingContext, nodes []simulator.NodeToBeRemoved, ts time.Time) {
	n.updateInternalState(ctx, autoscalingCtx, nodes, ts, func(nn simulator.NodeToBeRemoved) *time.Time {
		return nil
	})
}

func (n *Nodes) updateInternalState(ctx context.Context, autoscalingCtx *ca_context.AutoscalingContext, nodes []simulator.NodeToBeRemoved, ts time.Time, timestampGetter func(simulator.NodeToBeRemoved) *time.Time) {
	logger := klog.FromContext(ctx)
	updated := make(map[string]*node, len(nodes))
	for _, nn := range nodes {
		name := nn.Node.Name

		val, found := n.byName[name]
		if found {
			newNodeState := &node{
				ntbr:  nn,
				since: val.since,
			}
			n.lookupAndSetRemovalThreshold(ctx, newNodeState, autoscalingCtx.CloudProvider)
			updated[name] = newNodeState
		} else {
			updated[name] = n.newNode(ctx, nn, timestampGetter, ts, autoscalingCtx.CloudProvider)
		}
	}
	n.byName = updated
	n.cachedList = nil
	if klog.V(4).Enabled() {
		for k, v := range n.byName {
			logger.Info("is unneeded since duration", "k", k, "since", v.since, "sub", ts.Sub(v.since))
		}
	}
}

func (n *Nodes) newNode(ctx context.Context, nn simulator.NodeToBeRemoved, timestampGetter func(simulator.NodeToBeRemoved) *time.Time, ts time.Time, cp cloudprovider.CloudProvider) *node {
	var since time.Time
	if existingts := timestampGetter(nn); existingts != nil {
		since = *existingts
	} else {
		since = ts
	}

	newNode := &node{
		ntbr:  nn,
		since: since,
	}

	n.lookupAndSetRemovalThreshold(ctx, newNode, cp)

	return newNode
}

// Clear resets the internal state, dropping information about all tracked nodes.
func (n *Nodes) Clear(ctx context.Context) {
	n.Update(ctx, nil, nil, time.Time{})
}

// Contains returns true iff a given node is unneeded.
func (n *Nodes) Contains(nodeName string) bool {
	_, found := n.byName[nodeName]
	return found
}

// AsList returns a slice of unneeded Node objects.
func (n *Nodes) AsList() []*scaledown.UnneededNode {
	if n.cachedList == nil {
		n.cachedList = make([]*scaledown.UnneededNode, 0, len(n.byName))
		for _, v := range n.byName {
			n.cachedList = append(n.cachedList, &scaledown.UnneededNode{
				Node:             v.ntbr.Node,
				RemovalThreshold: v.removalThreshold,
			})
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
func (n *Nodes) RemovableAt(ctx context.Context, autoscalingCtx *ca_context.AutoscalingContext, scaleDownContext nodes.ScaleDownContext, ts time.Time) (empty, needDrain []simulator.NodeToBeRemoved, unremovable []simulator.UnremovableNode) {
	logger := klog.FromContext(ctx)
	nodeGroupSize := utils.GetNodeGroupSizeMap(autoscalingCtx.CloudProvider)
	emptyNodes, drainNodes := n.splitEmptyAndNonEmptyNodes()

	for nodeName, v := range emptyNodes {
		logger.V(2).Info("was unneeded", "nodeName", nodeName, "string", ts.Sub(v.since).String())
		if r := n.unremovableReason(ctx, autoscalingCtx, scaleDownContext, v, ts, nodeGroupSize); r != simulator.NoReason {
			unremovable = append(unremovable, simulator.UnremovableNode{Node: v.ntbr.Node, Reason: r})
			continue
		}
		empty = append(empty, v.ntbr)
	}
	for nodeName, v := range drainNodes {
		logger.V(2).Info("was unneeded", "nodeName", nodeName, "string", ts.Sub(v.since).String())
		if r := n.unremovableReason(ctx, autoscalingCtx, scaleDownContext, v, ts, nodeGroupSize); r != simulator.NoReason {
			unremovable = append(unremovable, simulator.UnremovableNode{Node: v.ntbr.Node, Reason: r})
			continue
		}
		needDrain = append(needDrain, v.ntbr)
	}
	return
}

// lookupAndSetRemovalThreshold gets the unneeded/unready time for a node and updates the node struct.
func (n *Nodes) lookupAndSetRemovalThreshold(ctx context.Context, v *node, cp cloudprovider.CloudProvider) {
	logger := klog.FromContext(ctx)
	nodeGroup, err := cp.NodeGroupForNode(v.ntbr.Node)
	if err != nil {
		logger.Error(err, "Error determining node group", "name", v.ntbr.Node.Name)
		return
	}
	if nodeGroup == nil {
		logger.V(4).Info("Skipping - no node group config", "name", v.ntbr.Node.Name)
		return
	}

	readiness, _ := kube_util.GetNodeReadiness(v.ntbr.Node)
	var removalThreshold time.Duration

	if readiness.Ready {
		removalThreshold, err = n.sdtg.GetScaleDownUnneededTime(nodeGroup)
	} else {
		removalThreshold, err = n.sdtg.GetScaleDownUnreadyTime(nodeGroup)
	}

	if err != nil {
		logger.Error(err, "Failed to get scale down unneeded/unready time", "name", v.ntbr.Node.Name)
		v.thresholdRetrievalFailed = true
		return
	}

	v.removalThreshold = removalThreshold
}

func (n *Nodes) unremovableReason(ctx context.Context, autoscalingCtx *ca_context.AutoscalingContext, scaleDownContext nodes.ScaleDownContext, v *node, ts time.Time, nodeGroupSize map[string]int) simulator.UnremovableReason {
	logger := klog.FromContext(ctx)
	if v.thresholdRetrievalFailed {
		return simulator.UnexpectedError
	}
	node := v.ntbr.Node
	// Check if node is marked with no scale down annotation.
	if eligibility.HasNoScaleDownAnnotation(node) {
		logger.V(4).Info("Skipping - scale down disabled annotation found", "node", node.Name)
		return simulator.ScaleDownDisabledAnnotation
	}

	readiness, _ := kube_util.GetNodeReadiness(v.ntbr.Node)

	if v.removalThreshold > 0 && !v.since.Add(v.removalThreshold).Before(ts) {
		if readiness.Ready {
			return simulator.NotUnneededLongEnough
		}
		return simulator.NotUnreadyLongEnough
	}

	nodeGroup, err := autoscalingCtx.CloudProvider.NodeGroupForNode(node)
	if err != nil || nodeGroup == nil {
		logger.V(4).Info("Skipping - no node group config", "node", node.Name)
		return simulator.NotAutoscaled
	}

	if reason := verifyMinSize(ctx, node.Name, nodeGroup, nodeGroupSize, scaleDownContext.ActuationStatus); reason != simulator.NoReason {
		return reason
	}

	checkResult, err := scaleDownContext.Tracker.ConsumeQuota(autoscalingCtx, nodeGroup, node, 1)
	if err != nil {
		logger.Error(err, "Failed to check limits", "node", node.Name)
		return simulator.UnexpectedError
	}

	if checkResult.Exceeded() {
		logger.V(2).Info("Skipping scale down in this batch - would violate minimum limits", "node", node.Name)
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

func verifyMinSize(ctx context.Context, nodeName string, nodeGroup cloudprovider.NodeGroup, nodeGroupSize map[string]int, as scaledown.ActuationStatus) simulator.UnremovableReason {
	logger := klog.FromContext(ctx)
	size, found := nodeGroupSize[nodeGroup.Id()]
	if !found {
		logger.Error(nil, "Error while checking node group size : group size not found in cache", "nodeGroupId", nodeGroup.Id())
		return simulator.UnexpectedError
	}
	deletionsInProgress := as.DeletionsCount(nodeGroup.Id())
	if size-deletionsInProgress <= nodeGroup.MinSize() {
		logger.V(1).Info("Skipping - node group min size reached", "nodeName", nodeName)
		return simulator.NodeGroupMinSizeReached
	}
	return simulator.NoReason
}
