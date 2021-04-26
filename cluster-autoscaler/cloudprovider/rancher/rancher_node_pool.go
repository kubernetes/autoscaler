/*
Copyright 2021 The Kubernetes Authors.

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

package rancher

import (
	"errors"
	"fmt"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"math"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/rancher/rancher"
	"k8s.io/klog/v2"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
)

// NodePool implements cloudprovider.NodeGroup interface. NodePool contains
// configuration info and functions to control a set of nodes that have the
// same capacity and set of labels.
type NodePool struct {
	manager   *manager
	rancherNP rancher.NodePool
	id        string
	minSize   int
	maxSize   int
}

// MaxSize returns maximum size of the node group.
func (n *NodePool) MaxSize() int {
	return n.maxSize
}

// MinSize returns minimum size of the node group.
func (n *NodePool) MinSize() int {
	return n.minSize
}

// TargetSize returns the current target size of the node group. It is possible that the
// number of nodes in Kubernetes is different at the moment but should be equal
// to Size() once everything stabilizes (new nodes finish startup and registration or
// removed nodes are deleted completely). Implementation required.
func (n *NodePool) TargetSize() (int, error) {
	return n.rancherNP.Quantity, nil
}

// IncreaseSize increases the size of the node group. To delete a node you need
// to explicitly name it and use DeleteNode. This function should wait until
// node group size is updated. Implementation required.
func (n *NodePool) IncreaseSize(delta int) error {
	klog.Infof("increasing NodePoolID: %q by %d", n.id, delta)
	if delta <= 0 {
		return fmt.Errorf("delta must be positive, have: %d", delta)
	}

	targetSize := n.rancherNP.Quantity + delta
	if targetSize > n.MaxSize() {
		return fmt.Errorf("size increase is too large - desired:%d max:%d", targetSize, n.maxSize)
	}

	updatedNodePool, err := n.manager.client.ResizeNodePool(n.id, targetSize)
	if err != nil {
		return err
	}

	if updatedNodePool.Quantity != targetSize {
		return fmt.Errorf("couldn't increase size to %d (delta: %d). Current size is: %d",
			targetSize, delta, n.rancherNP.Quantity)
	}

	n.rancherNP.Quantity = targetSize
	return nil
}

// DeleteNodes deletes nodes from this node group. Error is returned either on
// failure or if the given node doesn't belong to this node group. This function
// should wait until node group size is updated. Implementation required.
func (n *NodePool) DeleteNodes(nodes []*apiv1.Node) error {
	klog.Infof("Deleting %d nodes from %q\n", len(nodes), n.Id())

	size, err := n.TargetSize()
	if err != nil {
		klog.Errorf("failed to get node pool size:  %s", err.Error())
		return err
	}

	if size-len(nodes) < n.MinSize() {
		return fmt.Errorf("unable to delete nodes, size decrease is too small. current: %d desired: %d min: %d",
			size, size-len(nodes), n.minSize)
	}

	for _, node := range nodes {
		rn, err := n.manager.getNode(node)
		if err != nil {
			return err
		}

		klog.Infof("Deleting node %q - %q", rn.Name, rn.ID)
		if rn.NodePoolID != n.id {
			return fmt.Errorf("node: %s doesn't belong to the nodePool: %s", node.Name, n.id)
		}

		if err := n.manager.client.ScaleDownNode(rn.ID); err != nil {
			return err
		}

		n.rancherNP.Quantity--
	}

	return nil
}

// DecreaseTargetSize decreases the target size of the node group. This function
// doesn't permit to delete any existing node and can be used only to reduce the
// request for new nodes that have not been yet fulfilled. Delta should be negative.
// It is assumed that cloud provider will not delete the existing nodes when there
// is an option to just decrease the target. Implementation required.
func (n *NodePool) DecreaseTargetSize(delta int) error {
	delta = int(-math.Abs(float64(delta)))
	if delta == 0 {
		return errors.New("delta cannot be 0")
	}

	klog.Infof("Decreasing NodePoolID: %q by %d", n.id, delta)
	targetSize := n.rancherNP.Quantity + delta
	if targetSize < n.MinSize() {
		return fmt.Errorf("size decrease is too small. current: %d desired: %d min: %d",
			n.rancherNP.Quantity, targetSize, n.minSize)
	}

	updatedNodePool, err := n.manager.client.ResizeNodePool(n.id, targetSize)
	if err != nil {
		return err
	}

	if updatedNodePool.Quantity != targetSize {
		return fmt.Errorf("couldn't decrease size to %d (delta: %d). Current size is: %d",
			targetSize, delta, n.rancherNP.Quantity)
	}

	n.rancherNP.Quantity = targetSize
	return nil
}

// Id returns an unique identifier of the node group.
func (n *NodePool) Id() string {
	return n.id
}

// Debug returns a string containing all information regarding this node group.
func (n *NodePool) Debug() string {
	return fmt.Sprintf("id: %s (min:%d max:%d)", n.id, n.minSize, n.maxSize)
}

// Nodes returns a list of all nodes that belong to this node group.
// It is required that Instance objects returned by this method have Id field set.
// Other fields are optional.
// This list should include also instances that might have not become a kubernetes node yet.
func (n *NodePool) Nodes() ([]cloudprovider.Instance, error) {
	klog.V(4).Infof("Getting nodes from NodePool: %q", n.id)
	nodes, err := n.manager.client.NodesByNodePool(n.id)
	if err != nil {
		return nil, err
	}

	out := make([]cloudprovider.Instance, len(nodes))
	for i, n := range nodes {
		if n.State != "active" {
			continue
		}
		out[i] = cloudprovider.Instance{Id: n.ProviderID}
	}
	return out, nil
}

// TemplateNodeInfo returns a schedulerframework.NodeInfo structure of an empty
// (as if just started) node. This will be used in scale-up simulations to
// predict what would a new node look like if a node group was expanded. The returned
// NodeInfo is expected to have a fully populated Node object, with all of the labels,
// capacity and allocatable information as well as all pods that are started on
// the node by default, using manifest (most likely only kube-proxy). Implementation optional.
func (n *NodePool) TemplateNodeInfo() (*schedulerframework.NodeInfo, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// Exist checks if the node group really exists on the cloud provider side. Allows to tell the
// theoretical node group from the real one. Implementation required.
func (n *NodePool) Exist() bool {
	if _, err := n.manager.client.NodePoolByID(n.id); err != nil {
		return false
	}

	return true
}

// Create creates the node group on the cloud provider side. Implementation optional.
func (n *NodePool) Create() (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// Delete deletes the node group on the cloud provider side.
// This will be executed only for autoprovisioned node groups, once their size drops to 0.
// Implementation optional.
func (n *NodePool) Delete() error {
	return cloudprovider.ErrNotImplemented
}

// Autoprovisioned returns true if the node group is autoprovisioned. An autoprovisioned group
// was created by CA and can be deleted when scaled to 0.
func (n *NodePool) Autoprovisioned() bool {
	return false
}

// GetOptions returns NodeGroupAutoscalingOptions that should be used for this particular
// NodeGroup. Returning a nil will result in using default options.
// Implementation optional.
func (n *NodePool) GetOptions(defaults config.NodeGroupAutoscalingOptions) (*config.NodeGroupAutoscalingOptions, error) {
	return nil, cloudprovider.ErrNotImplemented
}
