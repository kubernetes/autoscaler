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

package nebius

import (
	"fmt"

	mk8sv1 "github.com/nebius/gosdk/proto/nebius/mk8s/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/klog/v2"
)

// NodeGroup implements cloudprovider.NodeGroup interface. NodeGroup contains
// configuration info and functions to control a set of nodes that have the
// same capacity and set of labels.
type NodeGroup struct {
	id        string
	manager   *Manager
	nodeGroup *mk8sv1.NodeGroup
	minSize   int
	maxSize   int
	instances map[string]struct{} // cached set of provider IDs
}

// MaxSize returns maximum size of the node group.
func (n *NodeGroup) MaxSize() int {
	return n.maxSize
}

// MinSize returns minimum size of the node group.
func (n *NodeGroup) MinSize() int {
	return n.minSize
}

// TargetSize returns the current target size of the node group. It is possible
// that the number of nodes in Kubernetes is different at the moment but should
// be equal to Size() once everything stabilizes (new nodes finish startup and
// registration or removed nodes are deleted completely). Implementation required.
func (n *NodeGroup) TargetSize() (int, error) {
	status := n.nodeGroup.GetStatus()
	if status == nil {
		return 0, nil
	}
	return int(status.GetTargetNodeCount()), nil
}

// IncreaseSize increases the size of the node group. To delete a node you need
// to explicitly name it and use DeleteNode. This function should wait until
// node group size is updated. Implementation required.
func (n *NodeGroup) IncreaseSize(delta int) error {
	if delta <= 0 {
		return fmt.Errorf("delta must be positive, have: %d", delta)
	}

	currentSize, err := n.TargetSize()
	if err != nil {
		return err
	}

	targetSize := currentSize + delta
	if targetSize > n.MaxSize() {
		return fmt.Errorf("size increase is too large. current: %d desired: %d max: %d",
			currentSize, targetSize, n.MaxSize())
	}

	klog.V(4).Infof("Increasing node group %s from %d to %d", n.id, currentSize, targetSize)
	return n.manager.setNodeGroupSize(n.id, targetSize)
}

// AtomicIncreaseSize is not implemented.
func (n *NodeGroup) AtomicIncreaseSize(delta int) error {
	return cloudprovider.ErrNotImplemented
}

// DeleteNodes deletes nodes from this node group. Each node is deleted via the
// Nebius Compute API by its specific instance ID, then the node group target
// size is updated to reflect the removal.
func (n *NodeGroup) DeleteNodes(nodes []*apiv1.Node) error {
	currentSize, err := n.TargetSize()
	if err != nil {
		return err
	}

	newSize := currentSize - len(nodes)
	if newSize < n.MinSize() {
		return fmt.Errorf("attempt to delete nodes below minimum size: current=%d, deleteCount=%d, min=%d",
			currentSize, len(nodes), n.MinSize())
	}

	var providerIDs []string
	for _, node := range nodes {
		if node.Spec.ProviderID == "" {
			return fmt.Errorf("node %s has no provider ID", node.Name)
		}
		providerIDs = append(providerIDs, node.Spec.ProviderID)
	}

	klog.V(4).Infof("Deleting %d nodes from node group %s (new size: %d)", len(nodes), n.id, newSize)
	return n.manager.deleteInstances(n.id, providerIDs, newSize)
}

// ForceDeleteNodes deletes nodes from the group regardless of constraints.
func (n *NodeGroup) ForceDeleteNodes(nodes []*apiv1.Node) error {
	return cloudprovider.ErrNotImplemented
}

// DecreaseTargetSize decreases the target size of the node group. This function
// doesn't permit to delete any existing node and can be used only to reduce the
// request for new nodes that have not been yet fulfilled. Delta should be negative.
// It is assumed that cloud provider will not delete the existing nodes when there
// is an option to just decrease the target. Implementation required.
func (n *NodeGroup) DecreaseTargetSize(delta int) error {
	if delta >= 0 {
		return fmt.Errorf("delta must be negative, have: %d", delta)
	}

	currentSize, err := n.TargetSize()
	if err != nil {
		return err
	}

	newSize := currentSize + delta
	if newSize < n.MinSize() {
		return fmt.Errorf("attempt to decrease target size below minimum: current=%d, delta=%d, min=%d",
			currentSize, delta, n.MinSize())
	}

	klog.V(4).Infof("Decreasing node group %s target from %d to %d", n.id, currentSize, newSize)
	return n.manager.setNodeGroupSize(n.id, newSize)
}

// Id returns an unique identifier of the node group.
func (n *NodeGroup) Id() string {
	return n.id
}

// Debug returns a string containing all information regarding this node group.
func (n *NodeGroup) Debug() string {
	return fmt.Sprintf("%s (%d:%d)", n.id, n.MinSize(), n.MaxSize())
}

// Nodes returns a list of all nodes that belong to this node group. It is
// required that Instance objects returned by this method have Id field set.
// Other fields are optional.
func (n *NodeGroup) Nodes() ([]cloudprovider.Instance, error) {
	instances := make([]cloudprovider.Instance, 0, len(n.instances))
	for providerID := range n.instances {
		instances = append(instances, cloudprovider.Instance{
			Id: providerID,
		})
	}
	return instances, nil
}

// TemplateNodeInfo returns a framework.NodeInfo structure of an empty
// (as if just started) node. This will be used in scale-up simulations to
// predict what would a new node look like if a node group was expanded. The
// returned NodeInfo is expected to have a fully populated Node object, with
// all of the labels, capacity and allocatable information as well as all pods
// that are started on the node by default, using manifest (most likely only
// kube-proxy). Implementation optional.
func (n *NodeGroup) TemplateNodeInfo() (*framework.NodeInfo, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// Exist checks if the node group really exists on the cloud provider side.
// Allows to tell the theoretical node group from the real one. Implementation required.
func (n *NodeGroup) Exist() bool {
	return n.nodeGroup != nil
}

// Create creates the node group on the cloud provider side. Implementation optional.
func (n *NodeGroup) Create() (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// Delete deletes the node group on the cloud provider side. Implementation optional.
func (n *NodeGroup) Delete() error {
	return cloudprovider.ErrNotImplemented
}

// Autoprovisioned returns true if the node group is autoprovisioned. An autoprovisioned group
// was created by CA and can be deleted when scaled to 0.
func (n *NodeGroup) Autoprovisioned() bool {
	return false
}

// GetOptions returns NodeGroupAutoscalingOptions for this NodeGroup. Implementation optional.
func (n *NodeGroup) GetOptions(defaults config.NodeGroupAutoscalingOptions) (*config.NodeGroupAutoscalingOptions, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// hasInstance returns true if the node group contains the given provider ID.
func (n *NodeGroup) hasInstance(providerID string) bool {
	_, ok := n.instances[providerID]
	return ok
}
