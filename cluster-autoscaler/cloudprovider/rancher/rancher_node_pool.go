package rancher

import (
	"fmt"
	"k8s.io/kubernetes/pkg/scheduler/schedulercache"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/config/dynamic"
)

// NodeRefFromNode creates InstanceConfig object from Node
func NodeRefFromNode(id string) (*RancherRef, error) {
	return &RancherRef{
		Name: id,
	}, nil
}

// NewNodePool creates a new NewScaleSet.
func NewNodePool(spec *dynamic.NodeGroupSpec, rm *RancherManager) (*NodePool, error) {
	nodePool := &NodePool{
		RancherRef: RancherRef{
			Name: spec.Name,
		},
		minSize: spec.MinSize,
		maxSize: spec.MaxSize,
		rancherManager: rm,
	}

	return nodePool, nil
}

// RancherRef contains a reference to some entity in Rancher.
type RancherRef struct {
	Name string
}

// NodePool implements NodeGroup interface.
type NodePool struct {
	RancherRef

	rancherManager *RancherManager

	minSize int
	maxSize int
}

// MaxSize returns maximum size of the node group.
func (nodePool *NodePool) MaxSize() int {
	return nodePool.maxSize
}

// MinSize returns minimum size of the node group.
func (nodePool *NodePool) MinSize() int {
	return nodePool.minSize
}

// TargetSize returns the current TARGET size of the node group. It is possible that the
// number is different from the number of nodes registered in Kubernetes.
func (nodePool *NodePool) TargetSize() (int, error) {
	size, err := nodePool.rancherManager.GetNodePoolSize(nodePool)
	return int(size), err
}

// Exist checks if the node group really exists on the cloud provider side. Allows to tell the
// theoretical node group from the real one.
func (nodePool *NodePool) Exist() bool {
	return true
}

// Create creates the node group on the cloud provider side.
func (nodePool *NodePool) Create() error {
	return cloudprovider.ErrAlreadyExist
}

// Autoprovisioned returns true if the node group is autoprovisioned.
func (nodePool *NodePool) Autoprovisioned() bool {
	return false
}

// Delete deletes the node group on the cloud provider side.
// This will be executed only for autoprovisioned node groups, once their size drops to 0.
func (nodePool *NodePool) Delete() error {
	return cloudprovider.ErrNotImplemented
}

// IncreaseSize increases NodePool size
func (nodePool *NodePool) IncreaseSize(delta int) error {
	if delta <= 0 {
		return fmt.Errorf("size increase must be positive")
	}
	size, err := nodePool.rancherManager.GetNodePoolSize(nodePool)
	if err != nil {
		return err
	}
	if int(size)+delta > nodePool.MaxSize() {
		return fmt.Errorf("size increase too large - desired:%d max:%d", int(size)+delta, nodePool.MaxSize())
	}
	return nodePool.rancherManager.SetNodePoolSize(nodePool, size+int64(delta))
}

// DecreaseTargetSize decreases the target size of the node group. This function
// doesn't permit to delete any existing node and can be used only to reduce the
// request for new nodes that have not been yet fulfilled. Delta should be negative.
// It is assumed that cloud provider will not delete the existing nodes if the size
// when there is an option to just decrease the target.
func (nodePool *NodePool) DecreaseTargetSize(delta int) error {
	if delta >= 0 {
		return fmt.Errorf("size decrease size must be negative")
	}
	size, err := nodePool.rancherManager.GetNodePoolSize(nodePool)
	if err != nil {
		return err
	}
	nodes, err := nodePool.rancherManager.GetNodePoolNodes(nodePool)
	if err != nil {
		return err
	}
	if int(size)+delta < len(nodes) {
		return fmt.Errorf("attempt to delete existing nodes targetSize:%d delta:%d existingNodes: %d",
			size, delta, len(nodes))
	}
	return nodePool.rancherManager.SetNodePoolSize(nodePool, size+int64(delta))
}

// DeleteNodes deletes the nodes from the group.
func (nodePool *NodePool) DeleteNodes(nodes []*apiv1.Node) error {
	size, err := nodePool.rancherManager.GetNodePoolSize(nodePool)
	if err != nil {
		return err
	}
	if int(size) <= nodePool.MinSize() {
		return fmt.Errorf("min size reached, nodes will not be deleted")
	}
	refs := make([]*RancherRef, 0, len(nodes))
	for _, node := range nodes {
		rancherref, err := NodeRefFromNode(node.Name)
		if err != nil {
			return err
		}
		refs = append(refs, rancherref)
	}
	return nodePool.rancherManager.DeleteInstances(refs)
}

// Id returns nodePool id.
func (nodePool *NodePool) Id() string {
	return nodePool.Name
}

// Debug returns a debug string for the NodePool.
func (nodePool *NodePool) Debug() string {
	return fmt.Sprintf("%s (%d:%d)", nodePool.Id(), nodePool.MinSize(), nodePool.MaxSize())
}

// Nodes returns a list of all nodes that belong to this node group.
func (nodePool *NodePool) Nodes() ([]string, error) {
	return nodePool.rancherManager.GetNodePoolNodes(nodePool)
}

// TemplateNodeInfo returns a node template for this node group.
func (nodePool *NodePool) TemplateNodeInfo() (*schedulercache.NodeInfo, error) {
	return nil, fmt.Errorf("not implemented")
}