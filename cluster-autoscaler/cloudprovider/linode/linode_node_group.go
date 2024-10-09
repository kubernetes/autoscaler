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

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/linode/linodego"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	klog "k8s.io/klog/v2"
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
	client       linodeAPIClient
	lkePools     map[int]*linodego.LKEClusterPool // key: LKEClusterPool.ID
	poolOpts     linodego.LKEClusterPoolCreateOptions
	lkeClusterID int
	minSize      int
	maxSize      int
	id           string // this is a LKEClusterPool Type
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
	return len(n.lkePools), nil
}

// IncreaseSize increases the size of the node group. To delete a node you need
// to explicitly name it and use DeleteNode. This function should wait until
// node group size is updated. Implementation required.
func (n *NodeGroup) IncreaseSize(delta int) error {
	if delta <= 0 {
		return fmt.Errorf("delta must be positive, have: %d", delta)
	}

	currentSize := len(n.lkePools)
	targetSize := currentSize + delta
	if targetSize > n.MaxSize() {
		return fmt.Errorf("size increase is too large. current: %d desired: %d max: %d",
			currentSize, targetSize, n.MaxSize())
	}

	for i := 0; i < delta; i++ {
		err := n.addNewLKEPool()
		if err != nil {
			return err
		}
	}

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
		pool, err := n.findLKEPoolForNode(node)
		if err != nil {
			return err
		}
		if pool == nil {
			return fmt.Errorf("Failed to delete node %q with provider ID %q: cannot find this node in the node group",
				node.Name, node.Spec.ProviderID)
		}
		err = n.deleteLKEPool(pool.ID)
		if err != nil {
			return fmt.Errorf("Failed to delete node %q with provider ID %q: %v",
				node.Name, node.Spec.ProviderID, err)
		}
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
	return n.id
}

// Debug returns a string containing all information regarding this node group.
func (n *NodeGroup) Debug() string {
	return fmt.Sprintf("node group ID: %s (min:%d max:%d)", n.Id(), n.MinSize(), n.MaxSize())
}

// extendendDebug returns a string containing detailed information regarding this node group.
func (n *NodeGroup) extendedDebug() string {
	lkePoolsList := make([]string, 0)
	for _, p := range n.lkePools {
		linodesList := make([]string, 0)
		for _, l := range p.Linodes {
			linode := fmt.Sprintf("ID: %q, instanceID: %d", l.ID, l.InstanceID)
			linodesList = append(linodesList, linode)
		}
		linodes := strings.Join(linodesList, "; ")
		lkePool := fmt.Sprintf("{ poolID: %d, count: %d, type: %s, associated linodes: [%s] }",
			p.ID, p.Count, p.Type, linodes)
		lkePoolsList = append(lkePoolsList, lkePool)
	}
	lkePools := strings.Join(lkePoolsList, ", ")
	return fmt.Sprintf("node group ID %s := min: %d, max: %d, LKEClusterID: %d, poolOpts: %+v, associated LKE pools: %s",
		n.id, n.minSize, n.maxSize, n.lkeClusterID, n.poolOpts, lkePools)
}

// Nodes returns a list of all nodes that belong to this node group. It is
// required that Instance objects returned by this method have Id field set.
// Other fields are optional.
func (n *NodeGroup) Nodes() ([]cloudprovider.Instance, error) {
	nodes := make([]cloudprovider.Instance, 0)
	for _, pool := range n.lkePools {
		linodesCount := len(pool.Linodes)
		if linodesCount != 1 {
			klog.V(2).Infof("Number of linodes in LKE pool %d is not exactly 1 (count: %d)", pool.ID, linodesCount)
		}
		for _, linode := range pool.Linodes {
			instance := cloudprovider.Instance{
				Id: "linode://" + strconv.Itoa(linode.InstanceID),
			}
			nodes = append(nodes, instance)
		}
	}
	return nodes, nil
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

// addNewLKEPool creates a new LKE Pool with a single linode in it and add it
// to the pools of this node group
func (n *NodeGroup) addNewLKEPool() error {
	ctx := context.Background()
	newPool, err := n.client.CreateLKEClusterPool(ctx, n.lkeClusterID, n.poolOpts)
	if err != nil {
		return fmt.Errorf("error on creating new LKE pool for LKE clusterID: %d", n.lkeClusterID)
	}
	n.lkePools[newPool.ID] = newPool
	return nil
}

// deleteLKEPool deletes a pool given its pool id and remove it from the pools
// of this node group
func (n *NodeGroup) deleteLKEPool(id int) error {
	_, ok := n.lkePools[id]
	if !ok {
		return fmt.Errorf("cannot delete LKE pool %d, this pool is not one we are managing", id)
	}
	ctx := context.Background()
	err := n.client.DeleteLKEClusterPool(ctx, n.lkeClusterID, id)
	if err != nil {
		return fmt.Errorf("error on deleting LKE pool %d, Linode API said: %v", id, err)
	}
	delete(n.lkePools, id)
	return nil
}

// findLKEPoolForNode returns the LKE pool where this node is, nil otherwise
func (n *NodeGroup) findLKEPoolForNode(node *apiv1.Node) (*linodego.LKEClusterPool, error) {
	providerID := node.Spec.ProviderID
	instanceIDStr := strings.TrimPrefix(providerID, providerIDPrefix)
	instanceID, err := strconv.Atoi(instanceIDStr)
	if err != nil {
		return nil, fmt.Errorf("Cannot convert ProviderID %q to linode istance id (must be of type %s<integer>)",
			providerID, providerIDPrefix)
	}
	for _, pool := range n.lkePools {
		for _, linode := range pool.Linodes {
			if linode.InstanceID == instanceID {
				return pool, nil
			}
		}
	}
	return nil, nil
}
