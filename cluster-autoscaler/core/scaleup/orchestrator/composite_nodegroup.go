/*
Copyright 2024 The Kubernetes Authors.

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

package orchestrator

import (
	"fmt"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodegroupset"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
)

// CompositeNodeGroup is a virtual node group that represents a collection of real node groups
// being scaled up together as part of a single Karpenter simulation result.
// NOTE: This is a hack to avoid balancing logic from messing up zone aware
// simulation results. A better solution might be to make all expanders support
// expansion options with multiple node groups and do balancing before
// expander.
type CompositeNodeGroup struct {
	cloudprovider.NodeGroup
	Plan []nodegroupset.ScaleUpInfo
}

// Id returns a composite ID.
func (c *CompositeNodeGroup) Id() string {
	return fmt.Sprintf("composite:%s", c.NodeGroup.Id())
}

// Debug returns debug info about the plan.
func (c *CompositeNodeGroup) Debug() string {
	return fmt.Sprintf("CompositeNodeGroup with plan: %v", c.Plan)
}

// TemplateNodeInfo returns the template of the representative group.
func (c *CompositeNodeGroup) TemplateNodeInfo() (*framework.NodeInfo, error) {
	return c.NodeGroup.TemplateNodeInfo()
}

// Nodes returns an empty list, as this is a virtual group.
func (c *CompositeNodeGroup) Nodes() ([]cloudprovider.Instance, error) {
	return []cloudprovider.Instance{}, nil
}

// Exist always returns true for the virtual group.
func (c *CompositeNodeGroup) Exist() bool {
	return true
}

// The following methods are not expected to be called on a CompositeNodeGroup
// as the Orchestrator should recognize it and handle its Plan directly.

func (c *CompositeNodeGroup) IncreaseSize(delta int) error {
	return fmt.Errorf("IncreaseSize should not be called on CompositeNodeGroup")
}

func (c *CompositeNodeGroup) AtomicIncreaseSize(delta int) error {
	return fmt.Errorf("AtomicIncreaseSize should not be called on CompositeNodeGroup")
}

func (c *CompositeNodeGroup) DeleteNodes(nodes []*apiv1.Node) error {
	return fmt.Errorf("DeleteNodes should not be called on CompositeNodeGroup")
}

func (c *CompositeNodeGroup) planNodes() int {
	total := 0
	for _, sui := range c.Plan {
		total += sui.NewSize - sui.CurrentSize
	}
	return total
}

func (c *CompositeNodeGroup) cappedPlan(limit int) []nodegroupset.ScaleUpInfo {
	total := c.planNodes()
	if total == 0 {
		return c.Plan
	}

	newPlan := make([]nodegroupset.ScaleUpInfo, len(c.Plan))
	copy(newPlan, c.Plan)

	added := 0
	for i := range newPlan {
		delta := newPlan[i].NewSize - newPlan[i].CurrentSize
		cappedDelta := (delta * limit) / total
		newPlan[i].NewSize = newPlan[i].CurrentSize + cappedDelta
		added += cappedDelta
	}

	// Distribute the remainder
	for i := 0; i < limit-added && i < len(newPlan); i++ {
		if newPlan[i].NewSize < newPlan[i].MaxSize {
			newPlan[i].NewSize++
		}
	}

	return newPlan
}
