/*
Copyright 2023 The Kubernetes Authors.

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

package budgets

import (
	"reflect"
	"strconv"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/utils/annotations"
)

// NodeGroupView is a subset of nodes from a given NodeGroup
type NodeGroupView struct {
	Group cloudprovider.NodeGroup
	Nodes []*apiv1.Node
	// BatchSize allows overriding the number of nodes needed to trigger deletion.
	// Useful for node groups which only scale between zero and max size.
	BatchSize int
}

// ScaleDownBudgetProcessor is responsible for keeping the number of nodes deleted in parallel within defined limits.
type ScaleDownBudgetProcessor struct {
	ctx *context.AutoscalingContext
}

// NewScaleDownBudgetProcessor creates a ScaleDownBudgetProcessor instance.
func NewScaleDownBudgetProcessor(ctx *context.AutoscalingContext) *ScaleDownBudgetProcessor {
	return &ScaleDownBudgetProcessor{
		ctx: ctx,
	}
}

// CropNodes crops the provided node lists to respect scale-down max parallelism budgets.
// The returned nodes are grouped by a node group.
// This function assumes that each node group may occur at most once in each of the "empty" and "drain" lists.
func (bp *ScaleDownBudgetProcessor) CropNodes(as scaledown.ActuationStatus, empty, drain []*apiv1.Node) (emptyToDelete, drainToDelete []*NodeGroupView) {
	emptyIndividual, emptyAtomic := bp.categorize(bp.group(empty))
	drainIndividual, drainAtomic := bp.categorize(bp.group(drain))

	emptyAtomicMap := groupBuckets(emptyAtomic)
	drainAtomicMap := groupBuckets(drainAtomic)

	emptyInProgress, drainInProgress := as.DeletionsInProgress()
	parallelismBudget := bp.ctx.MaxScaleDownParallelism - len(emptyInProgress) - len(drainInProgress)
	drainBudget := bp.ctx.MaxDrainParallelism - len(drainInProgress)

	var err error
	canOverflow := true
	emptyToDelete, drainToDelete = []*NodeGroupView{}, []*NodeGroupView{}

	allNodes, err := allNodes(bp.ctx.ClusterSnapshot)
	if err != nil {
		klog.Errorf("failed to read all nodes from the cluster snapshot for nodes cropping, err: %s", err)
	}

	for _, bucket := range emptyAtomic {
		drainNodes := []*apiv1.Node{}
		drainBucket, drainFound := drainAtomicMap[bucket.Group.Id()]
		if drainFound {
			drainNodes = drainBucket.Nodes
		}
		// For node groups using atomic scaling, skip them if either the total number
		// of empty and drain nodes exceeds the parallelism budget,
		// or if the number of drain nodes exceeds the drain budget.
		if parallelismBudget < len(bucket.Nodes)+len(drainNodes) ||
			drainBudget < len(drainNodes) {
			// One pod slice can sneak in even if it would exceed parallelism budget.
			// This is to help avoid starvation of pod slices by regular nodes,
			// also larger pod slices will immediately exceed parallelism budget.
			if parallelismBudget == 0 || (len(drainNodes) > 0 && drainBudget == 0) || !canOverflow {
				break
			}
		}
		var targetSize int
		if targetSize, err = bucket.Group.TargetSize(); err != nil {
			// Very unlikely to happen, as we've got this far with this group.
			klog.Errorf("not scaling atomically scaled group %v: can't get target size, err: %v", bucket.Group.Id(), err)
			continue
		}
		bucket.BatchSize = targetSize

		// If available we consider only registered nodes for the scale down,
		// excluding failed instances unable to register as K8s nodes. Waiting
		// for such instances could block scale down indefinitely.
		registeredNodes, err := bp.getAllRegisteredNodesForNodeGroup(allNodes, bucket.Group)
		if err != nil {
			klog.Errorf("failed to get registered nodes for node group %s: %v", bucket.Group.Id(), err)
		}
		currentSize := len(registeredNodes)
		if len(bucket.Nodes) == currentSize {
			bucket.BatchSize = currentSize
		}

		if len(bucket.Nodes)+len(drainNodes) != targetSize && len(bucket.Nodes)+len(drainNodes) != currentSize {
			// We can't only partially scale down atomic group.
			klog.Errorf("not scaling atomic group %v because not all nodes are candidates, target size: %v, current size: %v empty: %v, drainable: %v", bucket.Group.Id(), targetSize, currentSize, len(bucket.Nodes), len(drainNodes))
			continue
		}
		emptyToDelete = append(emptyToDelete, bucket)
		if drainFound {
			drainBucket.BatchSize = bucket.BatchSize
			drainToDelete = append(drainToDelete, drainBucket)
		}
		parallelismBudget -= len(bucket.Nodes) + len(drainNodes)
		drainBudget -= len(drainNodes)
		canOverflow = false
	}

	drainBudget = min(parallelismBudget, drainBudget)
	for _, bucket := range drainAtomic {
		if _, found := emptyAtomicMap[bucket.Group.Id()]; found {
			// This atomically-scaled node group should have been already processed
			// in the previous loop.
			continue
		}
		if drainBudget < len(bucket.Nodes) {
			// One pod slice can sneak in even if it would exceed parallelism budget.
			// This is to help avoid starvation of pod slices by regular nodes,
			// also larger pod slices will immediately exceed parallelism budget.
			if drainBudget == 0 || !canOverflow {
				break
			}
		}
		var targetSize int
		if targetSize, err = bucket.Group.TargetSize(); err != nil {
			// Very unlikely to happen, as we've got this far with this group.
			klog.Errorf("not scaling atomically scaled group %v: can't get target size, err: %v", bucket.Group.Id(), err)
			continue
		}
		bucket.BatchSize = targetSize

		// If available we consider only registered nodes for the scale down,
		// excluding failed instances unable to register as K8s nodes. Waiting
		// for such instances could block scale down indefinitely.
		registeredNodes, err := bp.getAllRegisteredNodesForNodeGroup(allNodes, bucket.Group)
		if err != nil {
			klog.Errorf("Failed to get registered nodes for node group %s: %v", bucket.Group.Id(), err)
		}
		currentSize := len(registeredNodes)
		if len(bucket.Nodes) == currentSize {
			bucket.BatchSize = currentSize
		}

		if len(bucket.Nodes) != targetSize && len(bucket.Nodes) != currentSize {
			// We can't only partially scale down atomic group.
			klog.Errorf("not scaling atomic group %v because not all nodes are candidates, target size: %v, current size: %v, empty: none, drainable: %v", bucket.Group.Id(), targetSize, currentSize, len(bucket.Nodes))
			continue
		}
		drainToDelete = append(drainToDelete, bucket)
		parallelismBudget -= len(bucket.Nodes)
		drainBudget -= len(bucket.Nodes)
		canOverflow = false
	}

	emptyToDelete, allowedCount := cropIndividualNodes(emptyToDelete, emptyIndividual, parallelismBudget)
	parallelismBudget -= allowedCount
	drainBudget = min(parallelismBudget, drainBudget)

	drainToDelete, _ = cropIndividualNodes(drainToDelete, drainIndividual, drainBudget)

	return emptyToDelete, drainToDelete
}

func groupBuckets(buckets []*NodeGroupView) map[string]*NodeGroupView {
	grouped := map[string]*NodeGroupView{}
	for _, bucket := range buckets {
		grouped[bucket.Group.Id()] = bucket
	}
	return grouped
}

// cropIndividualNodes returns two values:
// * nodes selected for deletion
// * the number of nodes planned for deletion in this invocation
func cropIndividualNodes(toDelete []*NodeGroupView, groups []*NodeGroupView, budget int) ([]*NodeGroupView, int) {
	remainingBudget := budget
	for _, bucket := range groups {
		if remainingBudget < 1 {
			break
		}
		if remainingBudget < len(bucket.Nodes) {
			bucket.Nodes = bucket.Nodes[:remainingBudget]
		}
		toDelete = append(toDelete, bucket)
		remainingBudget -= len(bucket.Nodes)
	}
	return toDelete, budget - remainingBudget
}

func (bp *ScaleDownBudgetProcessor) group(nodes []*apiv1.Node) []*NodeGroupView {
	groupMap := map[string]int{}
	grouped := []*NodeGroupView{}
	for _, node := range nodes {
		nodeGroup, err := bp.ctx.CloudProvider.NodeGroupForNode(node)
		if err != nil || nodeGroup == nil || reflect.ValueOf(nodeGroup).IsNil() {
			klog.Errorf("Failed to find node group for %s: %v", node.Name, err)
			continue
		}
		if idx, ok := groupMap[nodeGroup.Id()]; ok {
			grouped[idx].Nodes = append(grouped[idx].Nodes, node)
		} else {
			groupMap[nodeGroup.Id()] = len(grouped)
			grouped = append(grouped, &NodeGroupView{
				Group: nodeGroup,
				Nodes: []*apiv1.Node{node},
			})
		}
	}
	return grouped
}

func (bp *ScaleDownBudgetProcessor) categorize(groups []*NodeGroupView) (individual, atomic []*NodeGroupView) {
	for _, view := range groups {
		autoscalingOptions, err := view.Group.GetOptions(bp.ctx.NodeGroupDefaults)
		if err != nil && err != cloudprovider.ErrNotImplemented {
			klog.Errorf("Failed to get autoscaling options for node group %s: %v", view.Group.Id(), err)
			continue
		}
		if autoscalingOptions != nil && autoscalingOptions.ZeroOrMaxNodeScaling {
			atomic = append(atomic, view)
		} else {
			individual = append(individual, view)
		}
	}
	return individual, atomic
}

func allNodes(s clustersnapshot.ClusterSnapshot) ([]*apiv1.Node, error) {
	nodeInfos, err := s.ListNodeInfos()
	if err != nil {
		// This should never happen, List() returns err only because scheduler interface requires it.
		return nil, err
	}
	nodes := make([]*apiv1.Node, len(nodeInfos))
	for i, ni := range nodeInfos {
		nodes[i] = ni.Node()
	}
	return nodes, nil
}

func (bp *ScaleDownBudgetProcessor) getAllRegisteredNodesForNodeGroup(allNodes []*apiv1.Node, nodeGroup cloudprovider.NodeGroup) ([]*apiv1.Node, error) {
	allNodesInNodeGroup, err := nodeGroup.Nodes()
	if err != nil {
		return nil, err
	}
	nodeByNodeName := map[string]cloudprovider.Instance{}
	for _, node := range allNodesInNodeGroup {
		nodeByNodeName[node.Id] = node
	}
	var registeredNodesForNodeGroup []*apiv1.Node
	for _, node := range allNodes {
		if val, ok := node.Annotations[annotations.NodeUpcomingAnnotation]; ok {
			if res, ok := strconv.ParseBool(val); ok == nil && res {
				continue
			}
		}
		if _, ok := nodeByNodeName[node.Spec.ProviderID]; ok {
			registeredNodesForNodeGroup = append(registeredNodesForNodeGroup, node)
		}
	}
	return registeredNodesForNodeGroup, nil
}
