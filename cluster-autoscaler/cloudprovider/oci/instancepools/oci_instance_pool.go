/*
Copyright 2021-2023 Oracle and/or its affiliates.
*/

package instancepools

import (
	"fmt"

	"github.com/pkg/errors"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/common"
	ocicommon "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/common"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

// InstancePoolNodeGroup implements the NodeGroup interface using OCI instance pools.
type InstancePoolNodeGroup struct {
	manager    InstancePoolManager
	kubeClient kubernetes.Interface
	id         string
	minSize    int
	maxSize    int
}

// MaxSize returns maximum size of the instance-pool based node group.
func (ip *InstancePoolNodeGroup) MaxSize() int {
	return ip.maxSize
}

// MinSize returns minimum size of the instance-pool based node group.
func (ip *InstancePoolNodeGroup) MinSize() int {
	return ip.minSize
}

// TargetSize returns the current target size of the instance-pool based node group. It is possible that the
// number of nodes in Kubernetes is different at the moment but should be equal
// to Size() once everything stabilizes (new nodes finish startup and registration or
// removed nodes are deleted completely). Implementation required.
func (ip *InstancePoolNodeGroup) TargetSize() (int, error) {
	return ip.manager.GetInstancePoolSize(*ip)
}

// IncreaseSize increases the size of the instance-pool based node group. To delete a node you need
// to explicitly name it and use DeleteNode. This function should wait until
// instance-pool size is updated. Implementation required.
func (ip *InstancePoolNodeGroup) IncreaseSize(delta int) error {
	if delta <= 0 {
		return fmt.Errorf("size increase must be positive")
	}

	size, err := ip.manager.GetInstancePoolSize(*ip)
	if err != nil {
		return err
	}

	if size+delta > ip.MaxSize() {
		return fmt.Errorf("size increase too large - desired:%d max:%d", size+delta, ip.MaxSize())
	}

	return ip.manager.SetInstancePoolSize(*ip, size+delta)
}

// AtomicIncreaseSize is not implemented.
func (ip *InstancePoolNodeGroup) AtomicIncreaseSize(delta int) error {
	return cloudprovider.ErrNotImplemented
}

// DeleteNodes deletes nodes from this instance-pool. Error is returned either on
// failure or if the given node doesn't belong to this instance-pool. This function
// should wait until instance-pool size is updated. Implementation required.
func (ip *InstancePoolNodeGroup) DeleteNodes(nodes []*apiv1.Node) error {

	// FYI, unregistered nodes come in as the provider id as node name.

	klog.Infof("DeleteNodes called with %d node(s)", len(nodes))

	size, err := ip.manager.GetInstancePoolSize(*ip)
	if err != nil {
		return err
	}

	if size <= ip.MinSize() {
		return fmt.Errorf("min size reached, nodes will not be deleted")
	}

	refs := make([]common.OciRef, 0, len(nodes))
	for _, node := range nodes {
		belongs, err := ip.Belongs(node)
		if err != nil {
			return err
		}
		if !belongs {
			return fmt.Errorf("%s belong to a different instance-pool than %s", node.Name, ip.Id())
		}
		ociRef, err := ocicommon.NodeToOciRef(node)
		if err != nil {
			return err
		}

		refs = append(refs, ociRef)
	}

	return ip.manager.DeleteInstances(*ip, refs)
}

// DecreaseTargetSize decreases the target size of the instance-pool based node group. This function
// doesn't permit to delete any existing node and can be used only to reduce the
// request for new nodes that have not been yet fulfilled. Delta should be negative.
// It is assumed that cloud provider will not delete the existing nodes when there
// is an option to just decrease the target. Implementation required.
func (ip *InstancePoolNodeGroup) DecreaseTargetSize(delta int) error {
	if delta >= 0 {
		return fmt.Errorf("size decrease must be negative")
	}

	size, err := ip.manager.GetInstancePoolSize(*ip)
	if err != nil {
		return err
	}

	nodes, err := ip.manager.GetInstancePoolNodes(*ip)
	if err != nil {
		return err
	}

	if size+delta < len(nodes) {
		return fmt.Errorf("attempt to delete existing nodes targetSize:%d delta:%d existingNodes: %d",
			size, delta, len(nodes))
	}

	return ip.manager.SetInstancePoolSize(*ip, size+delta)
}

// Belongs returns true if the given node belongs to the InstancePoolNodeGroup.
func (ip *InstancePoolNodeGroup) Belongs(node *apiv1.Node) (bool, error) {
	ref, err := ocicommon.NodeToOciRef(node)
	if err != nil {
		return false, err
	}

	targetInstancePool, err := ip.manager.GetInstancePoolForInstance(ref)
	if err != nil {
		return false, err
	}

	if targetInstancePool == nil {
		return false, fmt.Errorf("%s doesn't belong to a known instance-pool", node.Name)
	}

	return targetInstancePool.Id() == ip.Id(), nil
}

// Id returns an unique identifier of the instance-pool based node group.
func (ip *InstancePoolNodeGroup) Id() string {
	return ip.id
}

// Debug returns a string containing all information regarding this instance-pool.
func (ip *InstancePoolNodeGroup) Debug() string {
	return fmt.Sprintf("%s (%d:%d)", ip.Id(), ip.MinSize(), ip.MaxSize())
}

// Nodes returns a list of all nodes that belong to this instance-pool.
// It is required that Instance objects returned by this method have Id field set.
// Other fields are optional.
// This list should include also instances that might have not become a kubernetes node yet.
func (ip *InstancePoolNodeGroup) Nodes() ([]cloudprovider.Instance, error) {
	return ip.manager.GetInstancePoolNodes(*ip)
}

// TemplateNodeInfo returns a framework.NodeInfo structure of an empty
// (as if just started) node. This will be used in scale-up simulations to
// predict what would a new node look like if a instance-pool was expanded. The returned
// NodeInfo is expected to have a fully populated Node object, with all of the labels,
// capacity and allocatable information as well as all pods that are started on
// the node by default, using manifest (most likely only kube-proxy). Implementation optional.
func (ip *InstancePoolNodeGroup) TemplateNodeInfo() (*framework.NodeInfo, error) {
	node, err := ip.manager.GetInstancePoolTemplateNode(*ip)
	if err != nil {
		return nil, errors.Wrap(err, "unable to build node info template")
	}

	nodeInfo := framework.NewNodeInfo(
		node, nil,
		&framework.PodInfo{Pod: cloudprovider.BuildKubeProxy(ip.id)},
		&framework.PodInfo{Pod: ocicommon.BuildCSINodePod()},
	)
	return nodeInfo, nil
}

// Exist checks if the instance-pool based node group really exists on the cloud provider side. Allows to tell the
// theoretical instance-pool from the real one. Implementation required.
func (ip *InstancePoolNodeGroup) Exist() bool {
	return true
}

// Create creates the instance-pool based node group on the cloud provider side. Implementation optional.
func (ip *InstancePoolNodeGroup) Create() (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// Delete deletes the instance-pool based node group on the cloud provider side.
// This will be executed only for autoprovisioned instance-pools, once their size drops to 0.
// Implementation optional.
func (ip *InstancePoolNodeGroup) Delete() error {
	return cloudprovider.ErrNotImplemented
}

// GetOptions returns NodeGroupAutoscalingOptions that should be used for this particular
// InstancePoolNodeGroup. Returning a nil will result in using default options.
func (ip *InstancePoolNodeGroup) GetOptions(defaults config.NodeGroupAutoscalingOptions) (*config.NodeGroupAutoscalingOptions, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// Autoprovisioned returns true if the instance-pool based node group is autoprovisioned. An autoprovisioned group
// was created by CA and can be deleted when scaled to 0.
func (ip *InstancePoolNodeGroup) Autoprovisioned() bool {
	return false
}
