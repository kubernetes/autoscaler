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

package exoscale

import (
	"context"
	"fmt"
	"sync"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	egoscale "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/exoscale/internal/github.com/exoscale/egoscale/v2"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
)

const (
	scaleToZeroSupported = false
)

// sksNodepoolNodeGroup implements cloudprovider.NodeGroup interface for Exoscale SKS Nodepools.
type sksNodepoolNodeGroup struct {
	sksNodepool *egoscale.SKSNodepool
	sksCluster  *egoscale.SKSCluster

	m *Manager

	sync.Mutex

	minSize int
	maxSize int
}

// MaxSize returns maximum size of the node group.
func (n *sksNodepoolNodeGroup) MaxSize() int {
	return n.maxSize
}

// MinSize returns minimum size of the node group.
func (n *sksNodepoolNodeGroup) MinSize() int {
	return n.minSize
}

// TargetSize returns the current target size of the node group. It is possible that the
// number of nodes in Kubernetes is different at the moment but should be equal
// to Size() once everything stabilizes (new nodes finish startup and registration or
// removed nodes are deleted completely). Implementation required.
func (n *sksNodepoolNodeGroup) TargetSize() (int, error) {
	return int(*n.sksNodepool.Size), nil
}

// IncreaseSize increases the size of the node group. To delete a node you need
// to explicitly name it and use DeleteNode. This function should wait until
// node group size is updated. Implementation required.
func (n *sksNodepoolNodeGroup) IncreaseSize(delta int) error {
	if delta <= 0 {
		return fmt.Errorf("delta must be positive, have: %d", delta)
	}

	targetSize := *n.sksNodepool.Size + int64(delta)

	if targetSize > int64(n.MaxSize()) {
		return fmt.Errorf("size increase is too large (current: %d desired: %d max: %d)",
			*n.sksNodepool.Size, targetSize, n.MaxSize())
	}

	infof("scaling SKS Nodepool %s to size %d", *n.sksNodepool.ID, targetSize)

	if err := n.m.client.ScaleSKSNodepool(n.m.ctx, n.m.zone, n.sksCluster, n.sksNodepool, targetSize); err != nil {
		errorf("unable to scale SKS Nodepool %s: %v", *n.sksNodepool.ID, err)
		return err
	}

	if err := n.waitUntilRunning(n.m.ctx); err != nil {
		return err
	}

	n.sksNodepool.Size = &targetSize

	return nil
}

// AtomicIncreaseSize is not implemented.
func (n *sksNodepoolNodeGroup) AtomicIncreaseSize(delta int) error {
	return cloudprovider.ErrNotImplemented
}

// DeleteNodes deletes nodes from this node group. Error is returned either on
// failure or if the given node doesn't belong to this node group. This function
// should wait until node group size is updated. Implementation required.
func (n *sksNodepoolNodeGroup) DeleteNodes(nodes []*apiv1.Node) error {
	n.Lock()
	defer n.Unlock()

	if err := n.waitUntilRunning(n.m.ctx); err != nil {
		return err
	}

	instanceIDs := make([]string, len(nodes))
	for i, node := range nodes {
		instanceIDs[i] = toNodeID(node.Spec.ProviderID)
	}

	infof("evicting SKS Nodepool %s members: %v", *n.sksNodepool.ID, instanceIDs)

	if err := n.m.client.EvictSKSNodepoolMembers(
		n.m.ctx,
		n.m.zone,
		n.sksCluster,
		n.sksNodepool,
		instanceIDs,
	); err != nil {
		errorf("unable to evict instances from SKS Nodepool %s: %v", *n.sksNodepool.ID, err)
		return err
	}

	if err := n.waitUntilRunning(n.m.ctx); err != nil {
		return err
	}

	newSize := *n.sksNodepool.Size - int64(len(instanceIDs))
	n.sksNodepool.Size = &newSize

	return nil
}

// ForceDeleteNodes deletes nodes from the group regardless of constraints.
func (n *sksNodepoolNodeGroup) ForceDeleteNodes(nodes []*apiv1.Node) error {
	return cloudprovider.ErrNotImplemented
}

// DecreaseTargetSize decreases the target size of the node group. This function
// doesn't permit to delete any existing node and can be used only to reduce the
// request for new nodes that have not been yet fulfilled. Delta should be negative.
// It is assumed that cloud provider will not delete the existing nodes when there
// is an option to just decrease the target. Implementation required.
func (n *sksNodepoolNodeGroup) DecreaseTargetSize(_ int) error {
	// Exoscale Instance Pools don't support down-sizing without deleting members,
	// so it is not possible to implement it according to the documented behavior.
	return nil
}

// Id returns an unique identifier of the node group.
func (n *sksNodepoolNodeGroup) Id() string {
	return *n.sksNodepool.ID
}

// Debug returns a string containing all information regarding this node group.
func (n *sksNodepoolNodeGroup) Debug() string {
	return fmt.Sprintf("Node group ID: %s (min:%d max:%d)", n.Id(), n.MinSize(), n.MaxSize())
}

// Nodes returns a list of all nodes that belong to this node group.
// It is required that Instance objects returned by this method have Id field set.
// Other fields are optional.
// This list should include also instances that might have not become a kubernetes node yet.
func (n *sksNodepoolNodeGroup) Nodes() ([]cloudprovider.Instance, error) {
	instancePool, err := n.m.client.GetInstancePool(n.m.ctx, n.m.zone, *n.sksNodepool.InstancePoolID)
	if err != nil {
		errorf(
			"unable to retrieve Instance Pool %s managed by SKS Nodepool %s",
			*n.sksNodepool.InstancePoolID,
			*n.sksNodepool.ID,
		)
		return nil, err
	}

	nodes := make([]cloudprovider.Instance, len(*instancePool.InstanceIDs))
	for i, id := range *instancePool.InstanceIDs {
		instance, err := n.m.client.GetInstance(n.m.ctx, n.m.zone, id)
		if err != nil {
			errorf("unable to retrieve Compute instance %s: %v", id, err)
			return nil, err
		}
		nodes[i] = toInstance(instance)
	}

	return nodes, nil
}

// TemplateNodeInfo returns a framework.NodeInfo structure of an empty
// (as if just started) node. This will be used in scale-up simulations to
// predict what would a new node look like if a node group was expanded. The returned
// NodeInfo is expected to have a fully populated Node object, with all of the labels,
// capacity and allocatable information as well as all pods that are started on
// the node by default, using manifest (most likely only kube-proxy). Implementation optional.
func (n *sksNodepoolNodeGroup) TemplateNodeInfo() (*framework.NodeInfo, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// Exist checks if the node group really exists on the cloud provider side. Allows to tell the
// theoretical node group from the real one. Implementation required.
func (n *sksNodepoolNodeGroup) Exist() bool {
	return n.sksNodepool != nil
}

// Create creates the node group on the cloud provider side. Implementation optional.
func (n *sksNodepoolNodeGroup) Create() (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// Delete deletes the node group on the cloud provider side.
// This will be executed only for autoprovisioned node groups, once their size drops to 0.
// Implementation optional.
func (n *sksNodepoolNodeGroup) Delete() error {
	return cloudprovider.ErrNotImplemented
}

// Autoprovisioned returns true if the node group is autoprovisioned. An autoprovisioned group
// was created by CA and can be deleted when scaled to 0.
func (n *sksNodepoolNodeGroup) Autoprovisioned() bool {
	return false
}

// GetOptions returns NodeGroupAutoscalingOptions that should be used for this particular
// sksNodepoolNodeGroup. Returning a nil will result in using default options.
func (n *sksNodepoolNodeGroup) GetOptions(_ config.NodeGroupAutoscalingOptions) (*config.NodeGroupAutoscalingOptions, error) {
	return nil, cloudprovider.ErrNotImplemented
}

func (n *sksNodepoolNodeGroup) waitUntilRunning(ctx context.Context) error {
	return pollCmd(ctx, func() (bool, error) {
		instancePool, err := n.m.client.GetInstancePool(ctx, n.m.zone, *n.sksNodepool.InstancePoolID)
		if err != nil {
			errorf("unable to retrieve Instance Pool %s: %s", *n.sksNodepool.InstancePoolID, err)
			return false, err
		}

		if *instancePool.State == "running" {
			return true, nil
		}

		return false, nil
	})
}
