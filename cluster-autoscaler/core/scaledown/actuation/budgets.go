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

package actuation

import (
	"reflect"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/deletiontracker"
)

type nodeBucket struct {
	group cloudprovider.NodeGroup
	nodes []*apiv1.Node
}

// ScaleDownBudgetProcessor is responsible for keeping the number of nodes deleted in parallel within defined limits.
type ScaleDownBudgetProcessor struct {
	ctx                 *context.AutoscalingContext
	nodeDeletionTracker *deletiontracker.NodeDeletionTracker
}

// NewScaleDownBudgetProcessor creates a ScaleDownBudgetProcessor instance.
func NewScaleDownBudgetProcessor(ctx *context.AutoscalingContext, ndt *deletiontracker.NodeDeletionTracker) *ScaleDownBudgetProcessor {
	return &ScaleDownBudgetProcessor{
		ctx:                 ctx,
		nodeDeletionTracker: ndt,
	}
}

// CropNodes crops the provided node lists to respect scale-down max parallelism budgets.
// The returned nodes are grouped by a node group.
func (bp *ScaleDownBudgetProcessor) CropNodes(empty, drain []*apiv1.Node) (emptyToDelete, drainToDelete []*nodeBucket) {
	emptyIndividual, emptyAtomic := bp.groupByNodeGroup(empty)
	drainIndividual, drainAtomic := bp.groupByNodeGroup(drain)

	emptyInProgress, drainInProgress := bp.nodeDeletionTracker.DeletionsInProgress()
	parallelismBudget := bp.ctx.MaxScaleDownParallelism - len(emptyInProgress) - len(drainInProgress)
	drainBudget := bp.ctx.MaxDrainParallelism - len(drainInProgress)

	emptyToDelete = []*nodeBucket{}
	for _, bucket := range emptyAtomic {
		if parallelismBudget < len(bucket.nodes) {
			// One pod slice can sneak in even if it would exceed parallelism budget.
			// This is to help avoid starvation of pod slices by regular nodes,
			// also larger pod slices will immediately exceed parallelism budget.
			if parallelismBudget == 0 || len(emptyToDelete) > 0 {
				break
			}
		}
		emptyToDelete = append(emptyToDelete, bucket)
		parallelismBudget -= len(bucket.nodes)
	}

	drainBudget = min(parallelismBudget, drainBudget)

	drainToDelete = []*nodeBucket{}
	for _, bucket := range drainAtomic {
		if drainBudget < len(bucket.nodes) {
			// One pod slice can sneak in even if it would exceed parallelism budget.
			// This is to help avoid starvation of pod slices by regular nodes,
			// also larger pod slices will immediately exceed parallelism budget.
			if drainBudget == 0 || len(emptyToDelete) > 0 || len(drainToDelete) > 0 {
				break
			}
		}
		drainToDelete = append(drainToDelete, bucket)
		drainBudget -= len(bucket.nodes)
		parallelismBudget -= len(bucket.nodes)
	}

	for _, bucket := range emptyIndividual {
		if parallelismBudget < 1 {
			break
		}
		if parallelismBudget < len(bucket.nodes) {
			bucket.nodes = bucket.nodes[:parallelismBudget]
		}
		emptyToDelete = append(emptyToDelete, bucket)
		parallelismBudget -= len(bucket.nodes)
	}

	drainBudget = min(parallelismBudget, drainBudget)

	for _, bucket := range drainIndividual {
		if drainBudget < 1 {
			break
		}
		if drainBudget < len(bucket.nodes) {
			bucket.nodes = bucket.nodes[:drainBudget]
		}
		drainToDelete = append(drainToDelete, bucket)
		drainBudget -= 1
	}

	return emptyToDelete, drainToDelete
}

func (bp *ScaleDownBudgetProcessor) groupByNodeGroup(nodes []*apiv1.Node) (individual, atomic []*nodeBucket) {
	individualGroup, atomicGroup := map[cloudprovider.NodeGroup]int{}, map[cloudprovider.NodeGroup]int{}
	individual, atomic = []*nodeBucket{}, []*nodeBucket{}
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
		if autoscalingOptions != nil && autoscalingOptions.AtomicScaleDown {
			if idx, ok := atomicGroup[nodeGroup]; ok {
				atomic[idx].nodes = append(atomic[idx].nodes, node)
			} else {
				atomicGroup[nodeGroup] = len(atomic)
				atomic = append(atomic, &nodeBucket{
					group: nodeGroup,
					nodes: []*apiv1.Node{node},
				})
			}
		} else {
			if idx, ok := individualGroup[nodeGroup]; ok {
				individual[idx].nodes = append(individual[idx].nodes, node)
			} else {
				individualGroup[nodeGroup] = len(individual)
				individual = append(individual, &nodeBucket{
					group: nodeGroup,
					nodes: []*apiv1.Node{node},
				})
			}
		}
	}
	return individual, atomic
}
