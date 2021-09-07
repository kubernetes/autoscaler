/*
Copyright 2019 The Kubernetes Authors.

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

package linode

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/linode/linodego"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
)

const (
	providerIDPrefix = "linode://"
)

// NodeGroup implements cloudprovider.NodeGroup interface. NodeGroup contains
// configuration info and functions to control a set of nodes that have the
// same capacity and set of labels.
//
// Receivers assume all fields are initialized (i.e. not nil).
//
// We cannot use an LKE pool as node group because LKE does not provide a way to
// delete a specific linode in a pool, but provides an API to selectively delete
// a LKE pool. To get around this issue, we build a NodeGroup with multiple LKE pools,
// each with a single linode in them.
type NodeGroup struct {
	clusterID int
	client    linodeAPIClient

	pool *linodego.LKEClusterPool // internal cache

	mtx sync.RWMutex
}

// MaxSize returns maximum size of the node group.
func (n *NodeGroup) MaxSize() int {
	n.mtx.RLock()
	defer n.mtx.RUnlock()

	return n.pool.Autoscaler.Max
}

// MinSize returns minimum size of the node group.
func (n *NodeGroup) MinSize() int {
	n.mtx.RLock()
	defer n.mtx.RUnlock()

	return n.pool.Autoscaler.Min
}

// TargetSize returns the current target size of the node group. It is possible
// that the number of nodes in Kubernetes is different at the moment but should
// be equal to Size() once everything stabilizes (new nodes finish startup and
// registration or removed nodes are deleted completely). Implementation
// required.
func (n *NodeGroup) TargetSize() (int, error) {
	n.mtx.RLock()
	defer n.mtx.RUnlock()

	return len(n.pool.Linodes), nil
}

// IncreaseSize increases the size of the node group. To delete a node you need
// to explicitly name it and use DeleteNode. This function should wait until
// node group size is updated. Implementation required.
func (n *NodeGroup) IncreaseSize(delta int) error {
	n.mtx.Lock()
	defer n.mtx.Unlock()

	if delta <= 0 {
		return fmt.Errorf("delta must be positive, have: %d", delta)
	}

	targetSize := n.pool.Count + delta
	if targetSize > n.pool.Autoscaler.Max {
		return fmt.Errorf("size increase is too large. current: %d desired: %d max: %d",
			n.pool.Count, targetSize, n.pool.Autoscaler.Max)
	}

	updateOpts := linodego.LKEClusterPoolUpdateOptions{Count: targetSize}
	nodePool, err := n.client.UpdateLKEClusterPool(context.Background(), n.clusterID, n.pool.ID, updateOpts)
	if err != nil {
		return fmt.Errorf("failed to scale up cluster (%d) pool (%d): %w", n.pool.ID, n.clusterID, err)
	}

	n.pool = nodePool
	return nil
}

// DeleteNodes deletes nodes from this node group (and also increasing the size
// of the node group with that). Error is returned either on failure or if the
// given node doesn't belong to this node group. This function should wait
// until node group size is updated. Implementation required.
func (n *NodeGroup) DeleteNodes(nodes []*apiv1.Node) error {
	n.mtx.Lock()
	defer n.mtx.Unlock()

	for _, node := range nodes {
		instanceID, err := parseProviderID(node.Spec.ProviderID)
		if err != nil {
			return err
		}

		var nodeID string
		var nodeIndex int
		for i, linode := range n.pool.Linodes {
			if instanceID == linode.InstanceID {
				nodeID = linode.ID
				nodeIndex = i
				break
			}
		}

		if nodeID == "" {
			return fmt.Errorf("failed to delete node: unable to find instance %d in pool %d", n.pool.ID, instanceID)
		}

		if err := n.client.DeleteLKEClusterPoolNode(context.Background(), n.clusterID, nodeID); err != nil {
			return fmt.Errorf("failed to delete node %s in pool %d: %v", nodeID, n.pool.ID, err)
		}

		copy(n.pool.Linodes[nodeIndex:], n.pool.Linodes[nodeIndex+1:])
		n.pool.Linodes = n.pool.Linodes[:len(n.pool.Linodes)-1]
	}
	return nil
}

// DecreaseTargetSize decreases the target size of the node group. This function
// doesn't permit to delete any existing node and can be used only to reduce the
// request for new nodes that have not been yet fulfilled. Delta should be negative.
// It is assumed that cloud provider will not delete the existing nodes when there
// is an option to just decrease the target. Implementation required.
func (n *NodeGroup) DecreaseTargetSize(delta int) error {
	// requests for new nodes are always fulfilled so we cannot
	// decrease the size without actually deleting nodes
	return cloudprovider.ErrNotImplemented
}

// Id returns an unique identifier of the node group.
func (n *NodeGroup) Id() string {
	return strconv.Itoa(n.pool.ID)
}

// Debug returns a string containing all information regarding this node group.
func (n *NodeGroup) Debug() string {
	return fmt.Sprintf("node group ID: %s (min:%d max:%d)", n.Id(), n.MinSize(), n.MaxSize())
}

// Nodes returns a list of all nodes that belong to this node group. It is
// required that Instance objects returned by this method have Id field set.
// Other fields are optional.
func (n *NodeGroup) Nodes() ([]cloudprovider.Instance, error) {
	n.mtx.RLock()
	defer n.mtx.RUnlock()

	nodes := make([]cloudprovider.Instance, len(n.pool.Linodes))
	for i, linode := range n.pool.Linodes {
		nodes[i] = cloudprovider.Instance{
			Id: providerIDPrefix + strconv.Itoa(linode.InstanceID),
		}
	}
	return nodes, nil
}

// TemplateNodeInfo returns a schedulerframework.NodeInfo structure of an empty
// (as if just started) node. This will be used in scale-up simulations to
// predict what would a new node look like if a node group was expanded. The
// returned NodeInfo is expected to have a fully populated Node object, with
// all of the labels, capacity and allocatable information as well as all pods
// that are started on the node by default, using manifest (most likely only
// kube-proxy). Implementation optional.
func (n *NodeGroup) TemplateNodeInfo() (*schedulerframework.NodeInfo, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// Exist checks if the node group really exists on the cloud provider side.
// Allows to tell the theoretical node group from the real one. Implementation
// required.
func (n *NodeGroup) Exist() bool {
	return true
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

func parseProviderID(id string) (int, error) {
	id = strings.TrimPrefix(id, providerIDPrefix)
	instanceID, err := strconv.Atoi(id)
	if err != nil {
		return 0, fmt.Errorf("failed to parse instance ID %q", id)
	}
	return instanceID, nil
}

func nodeGroupFromPool(client linodeAPIClient, clusterID int, pool *linodego.LKEClusterPool) *NodeGroup {
	if !pool.Autoscaler.Enabled {
		return nil
	}

	return &NodeGroup{
		clusterID: clusterID,
		client:    client,
		pool:      pool,
	}
}
