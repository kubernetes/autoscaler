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

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown"
)

// NodeGroupView is a subset of nodes from a given NodeGroup
type NodeGroupView struct {
	Group cloudprovider.NodeGroup
	Nodes []*apiv1.Node
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
func (bp *ScaleDownBudgetProcessor) CropNodes(as scaledown.ActuationStatus, empty, drain []*apiv1.Node) (emptyToDelete, drainToDelete []*NodeGroupView) {
	emptyIndividual, emptyAtomic := bp.categorize(bp.group(empty))
	drainIndividual, drainAtomic := bp.categorize(bp.group(drain))

	emptyInProgress, drainInProgress := as.DeletionsInProgress()
	parallelismBudget := bp.ctx.MaxScaleDownParallelism - len(emptyInProgress) - len(drainInProgress)
	drainBudget := bp.ctx.MaxDrainParallelism - len(drainInProgress)

	emptyToDelete, allowedCount, canOverflow := cropAtomicNodes(emptyAtomic, parallelismBudget, true)
	parallelismBudget -= allowedCount

	drainBudget = min(parallelismBudget, drainBudget)
	drainToDelete, allowedCount, _ = cropAtomicNodes(drainAtomic, drainBudget, canOverflow)
	parallelismBudget -= allowedCount
	drainBudget -= allowedCount

	emptyToDelete, allowedCount = cropIndividualNodes(emptyToDelete, emptyIndividual, parallelismBudget)
	parallelismBudget -= allowedCount
	drainBudget = min(parallelismBudget, drainBudget)

	drainToDelete, _ = cropIndividualNodes(drainToDelete, drainIndividual, drainBudget)

	return emptyToDelete, drainToDelete
}

// cropAtomicNodes returns three values:
// * nodes selected for deletion
// * the number of nodes planned for deletion in this invocation
// * whether a budget overflow is still allowed.
func cropAtomicNodes(groups []*NodeGroupView, budget int, canOverflow bool) ([]*NodeGroupView, int, bool) {
	toDelete := []*NodeGroupView{}
	remainingBudget := budget
	for _, bucket := range groups {
		if remainingBudget < len(bucket.Nodes) {
			// One pod slice can sneak in even if it would exceed parallelism budget.
			// This is to help avoid starvation of pod slices by regular nodes,
			// also larger pod slices will immediately exceed parallelism budget.
			if remainingBudget == 0 || (len(bucket.Nodes) > 0 && !canOverflow) {
				break
			}
		}
		toDelete = append(toDelete, bucket)
		remainingBudget -= len(bucket.Nodes)
		canOverflow = false
	}
	return toDelete, budget - remainingBudget, canOverflow
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
	groupMap := map[cloudprovider.NodeGroup]int{}
	grouped := []*NodeGroupView{}
	for _, node := range nodes {
		nodeGroup, err := bp.ctx.CloudProvider.NodeGroupForNode(node)
		if err != nil || nodeGroup == nil || reflect.ValueOf(nodeGroup).IsNil() {
			klog.Errorf("Failed to find node group for %s: %v", node.Name, err)
			continue
		}
		if idx, ok := groupMap[nodeGroup]; ok {
			grouped[idx].Nodes = append(grouped[idx].Nodes, node)
		} else {
			groupMap[nodeGroup] = len(grouped)
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
		if err != nil {
			klog.Errorf("Failed to get autoscaling options for node group %s: %v", view.Group.Id(), err)
			continue
		}
		if autoscalingOptions != nil && autoscalingOptions.AtomicScaling {
			atomic = append(atomic, view)
		} else {
			individual = append(individual, view)
		}
	}
	return individual, atomic
}

func (bp *ScaleDownBudgetProcessor) groupByNodeGroup(nodes []*apiv1.Node) (individual, atomic []*NodeGroupView) {
	individualGroup, atomicGroup := map[cloudprovider.NodeGroup]int{}, map[cloudprovider.NodeGroup]int{}
	individual, atomic = []*NodeGroupView{}, []*NodeGroupView{}
	for _, node := range nodes {
		nodeGroup, err := bp.ctx.CloudProvider.NodeGroupForNode(node)
		if err != nil || nodeGroup == nil || reflect.ValueOf(nodeGroup).IsNil() {
			klog.Errorf("Failed to find node group for %s: %v", node.Name, err)
			continue
		}
		autoscalingOptions, err := nodeGroup.GetOptions(bp.ctx.NodeGroupDefaults)
		if err != nil {
			klog.Errorf("Failed to get autoscaling options for node group %s: %v", nodeGroup.Id(), err)
			continue
		}
		if autoscalingOptions != nil && autoscalingOptions.AtomicScaling {
			if idx, ok := atomicGroup[nodeGroup]; ok {
				atomic[idx].Nodes = append(atomic[idx].Nodes, node)
			} else {
				atomicGroup[nodeGroup] = len(atomic)
				atomic = append(atomic, &NodeGroupView{
					Group: nodeGroup,
					Nodes: []*apiv1.Node{node},
				})
			}
		} else {
			if idx, ok := individualGroup[nodeGroup]; ok {
				individual[idx].Nodes = append(individual[idx].Nodes, node)
			} else {
				individualGroup[nodeGroup] = len(individual)
				individual = append(individual, &NodeGroupView{
					Group: nodeGroup,
					Nodes: []*apiv1.Node{node},
				})
			}
		}
	}
	return individual, atomic
}

func min(x, y int) int {
	if x <= y {
		return x
	}
	return y
}
