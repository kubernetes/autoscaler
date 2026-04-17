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

package utho

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strconv"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/utho/utho-go"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/klog/v2"
)

const (
	uthoLabel   = "utho.com"
	nodeIDLabel = "node_id"
)

var (
	// ErrNodePoolNotExist is return if no node pool exists for a given cluster ID
	ErrNodePoolNotExist = errors.New("node pool does not exist")
)

// NodeGroup implements cloudprovider.NodeGroup interface. NodeGroup contains
// configuration info and functions to control a set of nodes that have the
// same capacity and set of labels.
type NodeGroup struct {
	id        string
	clusterID int
	client    nodeGroupClient
	nodePool  *utho.NodepoolDetails

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
	return n.nodePool.Count, nil
}

// IncreaseSize increases the size of the node group. To delete a node you need
// to explicitly name it and use DeleteNode. This function should wait until
// node group size is updated. Implementation required.
func (n *NodeGroup) IncreaseSize(delta int) error {
	klog.V(4).Infof("IncreaseSize: requested delta=%d for node group %s", delta, n.id)

	if delta <= 0 {
		klog.Errorf("IncreaseSize: delta must be positive, got %d", delta)
		return fmt.Errorf("delta must be positive, have: %d", delta)
	}

	targetSize := n.nodePool.Count + delta

	if targetSize > n.MaxSize() {
		klog.Errorf("IncreaseSize: size increase too large for node group %s. current: %d, desired: %d, max: %d",
			n.id, n.nodePool.Count, targetSize, n.MaxSize())
		return fmt.Errorf("size increase is too large. current: %d desired: %d max: %d",
			n.nodePool.Count, targetSize, n.MaxSize())
	}

	param := utho.UpdateKubernetesAutoscaleNodepool{
		ClusterId:  n.clusterID,
		NodePoolId: n.id,
		Count:      strconv.Itoa(targetSize),
	}
	ctx := context.Background()
	klog.V(4).Infof("IncreaseSize: calling UpdateNodePool with targetSize=%d for node group %s", targetSize, n.id)
	_, err := n.client.UpdateNodePool(ctx, param)
	if err != nil {
		klog.Errorf("IncreaseSize: UpdateNodePool API error for node group %s: %v", n.id, err)
		return err
	}

	nodePool, err := n.client.ReadNodePool(ctx, n.clusterID, n.id)
	if err != nil {
		klog.Errorf("IncreaseSize: ReadNodePool API error for node group %s: %v", n.id, err)
		return fmt.Errorf("failed to read node pool after update for node group %s: %w", n.id, err)
	}

	if nodePool.Count != targetSize {
		klog.Errorf("IncreaseSize: couldn't increase size to %d (delta: %d). Current size is: %d for node group %s",
			targetSize, delta, nodePool.Count, n.id)
		return fmt.Errorf("couldn't increase size to %d (delta: %d). Current size is: %d",
			targetSize, delta, nodePool.Count)
	}

	// update internal cache
	n.nodePool.Count = targetSize
	klog.V(4).Infof("IncreaseSize: node group %s count updated, new Count=%d", n.id, n.nodePool.Count)
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
	klog.V(4).Infof("DeleteNodes: requested deletion of %d nodes from pool %s", len(nodes), n.id)

	ctx := context.Background()

	for _, node := range nodes {
		// Find the node ID label
		nodeID, ok := node.Labels["node_id"]
		if !ok {
			// CA creates fake node objects to represent upcoming VMs that
			// haven't registered as nodes yet. We cannot delete the node at
			// this point.
			klog.V(4).Infof("DeleteNodes: node %q (providerID=%q) missing node_id label",
				node.Name, node.Spec.ProviderID)
			return fmt.Errorf("cannot delete node %q with provider ID %q on node pool %q: node ID label is missing",
				node.Name, node.Spec.ProviderID, n.id)
		}
		klog.V(4).Infof("DeleteNodes: resolved nodeID=%s for node %q in pool %s", nodeID, node.Name, n.id)

		// Call provider API
		param := utho.DeleteNodeParams{
			ClusterId: n.clusterID,
			PoolId:    n.id,
			NodeId:    nodeID,
		}
		klog.V(4).Infof("DeleteNodes: calling DeleteNode(cluster=%d, pool=%s, node=%s)",
			n.clusterID, n.id, nodeID)

		if _, err := n.client.DeleteNode(ctx, param); err != nil {
			klog.Errorf("DeleteNodes: API error for cluster %d pool %s node %s: %v",
				n.clusterID, n.id, nodeID, err)
			return fmt.Errorf("deleting node failed for cluster %q node pool %q node %q: %w",
				n.clusterID, n.id, nodeID, err)
		}
		klog.V(4).Infof("DeleteNodes: provider confirmed deletion of node %s in pool %s", nodeID, n.id)

		// decrement the count by one  after a successful delete
		n.nodePool.Count--
		klog.V(4).Infof("DeleteNodes: pool %s count decremented, new Count=%d", n.id, n.nodePool.Count)
	}

	klog.V(4).Infof("DeleteNodes: completed deletion for pool %s", n.id)
	return nil
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
	klog.V(4).Infof("DecreaseTargetSize: requested delta=%d for node group %s", delta, n.id)

	if delta >= 0 {
		klog.Errorf("DecreaseTargetSize: delta must be negative, got %d", delta)
		return fmt.Errorf("delta must be negative, have: %d", delta)
	}

	targetSize := n.nodePool.Count + delta
	if targetSize < n.MinSize() {
		klog.Errorf("DecreaseTargetSize: size decrease too small for node group %s. current: %d, desired: %d, min: %d",
			n.Id(), n.nodePool.Count, targetSize, n.MinSize())
		return fmt.Errorf("node group %s: size decrease is too small. current size: %d, desired size: %d, minimum size: %d",
			n.Id(), n.nodePool.Count, targetSize, n.MinSize())
	}

	req := utho.UpdateKubernetesAutoscaleNodepool{
		Count:      strconv.Itoa(targetSize),
		ClusterId:  n.clusterID,
		NodePoolId: n.id,
		Label:      uthoLabel,
		Size:       strconv.Itoa(targetSize),
	}
	ctx := context.Background()
	klog.V(4).Infof("DecreaseTargetSize: calling UpdateNodePool with targetSize=%d for node group %s", targetSize, n.id)
	updatedNodePool, err := n.client.UpdateNodePool(ctx, req)
	if err != nil {
		klog.Errorf("DecreaseTargetSize: UpdateNodePool API error for node group %s: %v", n.id, err)
		return err
	}

	if updatedNodePool.Count != targetSize {
		klog.Errorf("DecreaseTargetSize: couldn't decrease size to %d (delta: %d). Current size is: %d for node group %s",
			targetSize, delta, updatedNodePool.Count, n.id)
		return fmt.Errorf("couldn't decrease size to %d (delta: %d). Current size is: %d",
			targetSize, delta, updatedNodePool.Count)
	}

	// update internal cache
	n.nodePool.Count = targetSize
	klog.V(4).Infof("DecreaseTargetSize: node group %s count updated, new Count=%d", n.id, n.nodePool.Count)
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
// required that Instance objects returned by this method have Id field set.
// Other fields are optional.
func (n *NodeGroup) Nodes() ([]cloudprovider.Instance, error) {
	klog.V(5).Infof("Checking nodes for node group %s", n.Id())

	if n.nodePool == nil {
		klog.Errorf("Node pool object for group %s is nil", n.Id())
		return nil, errors.New("node pool instance is not created")
	}

	nodes := n.nodePool.Workers
	klog.V(5).Infof("Node group %s has %d workers in its data cache", n.Id(), len(nodes))

	instances := make([]cloudprovider.Instance, 0, len(nodes))
	seenIDs := make(map[string]bool)

	for _, nd := range nodes {
		nodeID := strconv.Itoa(nd.ID)
		if nodeID == "0" || seenIDs[nodeID] {
			klog.V(4).Infof("Skipping invalid or duplicate node ID: %s", nodeID)
			continue
		}
		seenIDs[nodeID] = true

		instances = append(instances, cloudprovider.Instance{
			Id: toProviderID(nodeID),
		})
	}

	klog.V(5).Infof("Returning %d valid instances for node group %s", len(instances), n.Id())
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
	klog.V(4).Infof("TemplateNodeInfo: start for node-group %s", n.id)

	if n.nodePool == nil {
		return nil, fmt.Errorf("node pool not initialised for group %s", n.id)
	}
	if len(n.nodePool.Workers) == 0 {
		return nil, fmt.Errorf("node pool %s has no example worker to derive resources", n.id)
	}
	klog.V(5).Infof("TemplateNodeInfo: using first worker of pool %s as spec template", n.id)

	// Use the first worker as a spec template
	w := n.nodePool.Workers[0]

	// Build a synthetic *v1.Node
	name := fmt.Sprintf("%s-template-%d", n.id, rand.Int63())
	node := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: map[string]string{},
		},
	}

	// Capacity / Allocatable
	cpuQty := resource.NewQuantity(int64(w.Cpu), resource.DecimalSI)
	memQty := resource.NewQuantity(int64(w.Ram)*1024*1024, resource.DecimalSI) // MB→bytes
	node.Status.Capacity = apiv1.ResourceList{
		apiv1.ResourceCPU:    *cpuQty,
		apiv1.ResourceMemory: *memQty,
		apiv1.ResourcePods:   *resource.NewQuantity(110, resource.DecimalSI),
	}
	node.Status.Allocatable = node.Status.Capacity
	klog.V(4).Infof("TemplateNodeInfo: capacity set CPU=%s  Mem=%s", cpuQty.String(), memQty.String())

	// Generic + provider labels
	node.Labels = join(node.Labels, map[string]string{
		"kubernetes.io/os":                 "linux",
		"kubernetes.io/arch":               "amd64",
		"node.kubernetes.io/instance-type": n.nodePool.Size,
		"topology.kubernetes.io/zone":      n.nodePool.Ip,
		"node_id":                          strconv.Itoa(w.ID),
	})
	klog.V(5).Infof("TemplateNodeInfo: labels populated for node %s", name)

	// Mark Ready
	node.Status.Conditions = readyConditions()

	// Wrap in NodeInfo, add kube-proxy pod
	ni := framework.NewNodeInfo(node, nil, framework.NewPodInfo(buildKubeProxy(n.id), nil))
	klog.V(4).Infof("TemplateNodeInfo: completed for node-group %s, returning NodeInfo", n.id)
	return ni, nil
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
