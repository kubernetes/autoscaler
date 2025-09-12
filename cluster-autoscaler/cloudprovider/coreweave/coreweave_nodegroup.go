/*
Copyright 2025 The Kubernetes Authors.

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

package coreweave

import (
	"fmt"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/klog/v2"
	"sync"
)

// CoreWeaveNodeGroup represents a node group in the CoreWeave cloud provider.
type CoreWeaveNodeGroup struct {
	Name     string
	nodepool *CoreWeaveNodePool
	mutex    sync.Mutex
}

// NewCoreWeaveNodeGroup creates a new CoreWeaveNodeGroup instance.
// It takes a CoreWeaveNodePool as an argument and constructs the node group name
func NewCoreWeaveNodeGroup(nodepool *CoreWeaveNodePool) *CoreWeaveNodeGroup {
	return &CoreWeaveNodeGroup{
		Name:     nodepool.GetName(),
		nodepool: nodepool,
	}
}

// MaxSize returns the maximum size of the node group.
func (ng *CoreWeaveNodeGroup) MaxSize() int {
	return ng.nodepool.GetMaxSize()
}

// MinSize returns the minimum size of the node group.
func (ng *CoreWeaveNodeGroup) MinSize() int {
	return ng.nodepool.GetMinSize()
}

// TargetSize returns the current target size of the node group.
func (ng *CoreWeaveNodeGroup) TargetSize() (int, error) {
	return ng.nodepool.GetTargetSize(), nil
}

// IncreaseSize increases the size of the node group by the specified delta.
func (ng *CoreWeaveNodeGroup) IncreaseSize(delta int) error {
	klog.V(4).Infof("Increasing size of node group %s by %d", ng.Name, delta)
	return ng.nodepool.SetSize(ng.nodepool.GetTargetSize() + delta)
}

// AtomicIncreaseSize atomically increases the size of the node group by the specified delta.
// This method is not implemented for CoreWeaveNodeGroup.
func (ng *CoreWeaveNodeGroup) AtomicIncreaseSize(delta int) error {
	return cloudprovider.ErrNotImplemented
}

// DeleteNodes deletes the specified nodes from the node group.
// It verifies that the nodes belong to this node group and that deleting them does not violate the
// minimum size constraint. If the checks pass, it marks the nodes for removal and updates the
// target size of the node group accordingly.
// If any check fails, it returns an error with a descriptive message.
// If the nodes are successfully marked for removal, it returns nil.
func (ng *CoreWeaveNodeGroup) DeleteNodes(nodes []*apiv1.Node) error {
	ng.mutex.Lock()
	defer ng.mutex.Unlock()
	// Validate that the nodes belong to this node group
	err := ng.nodepool.ValidateNodes(nodes)
	if err != nil {
		return fmt.Errorf("some nodes do not belong to node group %s: %v", ng.Name, err)
	}
	// If we reach here, it means we can delete the nodes
	for _, node := range nodes {
		// Mark the node for removal
		if err := ng.nodepool.MarkNodeForRemoval(node); err != nil {
			return fmt.Errorf("failed to mark node %s for removal: %v", node.Name, err)
		}
	}
	//update target size
	if err := ng.nodepool.SetSize(ng.nodepool.GetTargetSize() - len(nodes)); err != nil {
		return fmt.Errorf("failed to update target size after marking nodes for removal: %v", err)
	}
	return nil
}

// ForceDeleteNodes is not implemented for CoreWeaveNodeGroup.
// node groups in CoreWeave do not support force deletion of nodes.
func (ng *CoreWeaveNodeGroup) ForceDeleteNodes(nodes []*apiv1.Node) error {
	return cloudprovider.ErrNotImplemented
}

// DecreaseTargetSize decreases the target size of the node group by the specified delta.
func (ng *CoreWeaveNodeGroup) DecreaseTargetSize(delta int) error {
	klog.V(4).Infof("Decreasing target size of node group %s by %d", ng.Name, delta)
	return ng.nodepool.SetSize(ng.nodepool.GetTargetSize() - delta)
}

// Id returns the unique identifier of the node group.
func (ng *CoreWeaveNodeGroup) Id() string {
	return ng.nodepool.GetUID()
}

// Debug returns a debug string for the node group.
func (ng *CoreWeaveNodeGroup) Debug() string {
	return ""
}

// Nodes returns the list of nodes in the node group.
func (ng *CoreWeaveNodeGroup) Nodes() ([]cloudprovider.Instance, error) {
	// Return empty slice to avoid "not registered" warnings
	return []cloudprovider.Instance{}, nil
}

// TemplateNodeInfo returns a template NodeInfo for the node group.
// This method is not implemented for CoreWeaveNodeGroup.
func (ng *CoreWeaveNodeGroup) TemplateNodeInfo() (*framework.NodeInfo, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// Exist checks if the node group exists.
func (ng *CoreWeaveNodeGroup) Exist() bool { return true }

// Create is not implemented for CoreWeaveNodeGroup.
func (ng *CoreWeaveNodeGroup) Create() (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// Delete is not implemented for CoreWeaveNodeGroup.
func (ng *CoreWeaveNodeGroup) Delete() error {
	return cloudprovider.ErrNotImplemented
}

// Autoprovisioned returns whether the node group is autoprovisioned.
// In CoreWeave, node groups are not autoprovisioned, so it returns false
func (ng *CoreWeaveNodeGroup) Autoprovisioned() bool { return false }

// GetOptions returns the autoscaling options for the node group.
func (ng *CoreWeaveNodeGroup) GetOptions(defaults config.NodeGroupAutoscalingOptions) (*config.NodeGroupAutoscalingOptions, error) {
	return nil, cloudprovider.ErrNotImplemented
}
