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

package vultr

import (
	"context"
	"errors"
	"fmt"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/vultr/govultr"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
)

const (
	vkeLabel    = "vke.vultr.com"
	nodeIDLabel = vkeLabel + "/node-id"
)

// NodeGroup implements cloudprovider.NodeGroup interface. NodeGroup contains
// configuration info and functions to control a set of nodes that have the
// same capacity and set of labels.
type NodeGroup struct {
	id        string
	clusterID string
	client    vultrClient
	nodePool  *govultr.NodePool

	minSize int
	maxSize int
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
// registration or removed nodes are deleted completely). Implementation
// required.
func (n *NodeGroup) TargetSize() (int, error) {
	return n.nodePool.NodeQuantity, nil
}

// IncreaseSize increases the size of the node group. To delete a node you need
// to explicitly name it and use DeleteNode. This function should wait until
// node group size is updated.
func (n *NodeGroup) IncreaseSize(delta int) error {
	if delta <= 0 {
		return fmt.Errorf("delta must be positive, have: %d", delta)
	}

	targetSize := n.nodePool.NodeQuantity + delta
	if targetSize > n.MaxSize() {
		return fmt.Errorf("size increase is too large. current: %d desired: %d max: %d",
			n.nodePool.NodeQuantity, targetSize, n.MaxSize())
	}

	req := &govultr.NodePoolReqUpdate{NodeQuantity: targetSize}

	updatedNodePool, err := n.client.UpdateNodePool(context.Background(), n.clusterID, n.id, req)
	if err != nil {
		return err
	}

	if updatedNodePool.NodeQuantity != targetSize {
		return fmt.Errorf("couldn't increase size to %d (delta: %d). Current size is: %d",
			targetSize, delta, updatedNodePool.NodeQuantity)
	}

	// update internal cache
	n.nodePool.NodeQuantity = targetSize
	return nil
}

// AtomicIncreaseSize is not implemented.
func (n *NodeGroup) AtomicIncreaseSize(delta int) error {
	return cloudprovider.ErrNotImplemented
}

// DeleteNodes deletes nodes from this node group (and also increasing the size
// of the node group with that). Error is returned either on failure or if the
// given node doesn't belong to this node group. This function should wait
// until node group size is updated. Implementation required.
func (n *NodeGroup) DeleteNodes(nodes []*apiv1.Node) error {
	for _, node := range nodes {
		nodeID, ok := node.Labels[nodeIDLabel]
		providerID := node.Spec.ProviderID

		if !ok {
			if providerID == "" {
				return fmt.Errorf("cannot delete node %q on node pool %q: missing provider ID and node ID label %q", node.Name, n.id, nodeIDLabel)
			}
			nodeID = toNodeID(providerID)
		}

		err := n.client.DeleteNodePoolInstance(context.Background(), n.clusterID, n.id, nodeID)
		if err != nil {
			return fmt.Errorf("deleting node failed for cluster: %q node pool: %q node: %q: %s",
				n.clusterID, n.id, nodeID, err)
		}

		n.nodePool.NodeQuantity--
	}

	return nil
}

// DecreaseTargetSize decreases the target size of the node group. This function
// doesn't permit to delete any existing node and can be used only to reduce the
// request for new nodes that have not been yet fulfilled. Delta should be negative.
// It is assumed that cloud provider will not delete the existing nodes when there
// is an option to just decrease the target.
func (n *NodeGroup) DecreaseTargetSize(delta int) error {
	if delta >= 0 {
		return fmt.Errorf("delta must be negative, have: %d", delta)
	}

	targetSize := n.nodePool.NodeQuantity + delta
	if targetSize < n.MinSize() {
		return fmt.Errorf("size decrease is too small. current: %d desired: %d min: %d",
			n.nodePool.NodeQuantity, targetSize, n.MinSize())
	}

	req := &govultr.NodePoolReqUpdate{NodeQuantity: targetSize}
	updatedNodePool, err := n.client.UpdateNodePool(context.Background(), n.clusterID, n.id, req)
	if err != nil {
		return err
	}

	if updatedNodePool.NodeQuantity != targetSize {
		return fmt.Errorf("couldn't increase size to %d (delta: %d). Current size is: %d",
			targetSize, delta, updatedNodePool.NodeQuantity)
	}

	// update internal cache
	n.nodePool.NodeQuantity = targetSize
	return nil
}

// Id returns an unique identifier of the node group.
func (n *NodeGroup) Id() string {
	return n.id
}

// Debug returns a string containing all information regarding this node group.
func (n *NodeGroup) Debug() string {
	return fmt.Sprintf("node group ID: %s (min:%d max:%d)", n.Id(), n.MinSize(), n.MaxSize())
}

// Nodes returns a list of all nodes that belong to this node group.  It is
// required that Instance objects returned by this method have ID field set.
// Other fields are optional.
func (n *NodeGroup) Nodes() ([]cloudprovider.Instance, error) {
	if n.nodePool == nil {
		return nil, errors.New("node pool instance is not created")
	}

	nodes := n.nodePool.Nodes
	instances := make([]cloudprovider.Instance, 0, len(nodes))
	for _, nd := range nodes {

		i := cloudprovider.Instance{
			Id: toProviderID(nd.ID),
		}

		instances = append(instances, i)
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
// Allows to tell the theoretical node group from the real one. Implementation
// required.
func (n *NodeGroup) Exist() bool {
	return n.nodePool != nil
}

// Create creates the node group on the cloud provider side. Implementation
// optional.
func (n *NodeGroup) Create() (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// Delete deletes the node group on the cloud provider side.  This will be
// executed only for autoprovisioned node groups, once their size drops to 0.
// Implementation optional.
func (n *NodeGroup) Delete() error {
	return cloudprovider.ErrNotImplemented
}

// Autoprovisioned returns true if the node group is autoprovisioned. An
// autoprovisioned group was created by CA and can be deleted when scaled to 0.
func (n *NodeGroup) Autoprovisioned() bool {
	return false
}

// GetOptions returns NodeGroupAutoscalingOptions that should be used for this particular
// NodeGroup. Returning a nil will result in using default options.
func (n *NodeGroup) GetOptions(defaults config.NodeGroupAutoscalingOptions) (*config.NodeGroupAutoscalingOptions, error) {
	return nil, cloudprovider.ErrNotImplemented
}
