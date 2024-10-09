/*
Copyright 2020-2023 Oracle and/or its affiliates.
*/

package nodepools

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/pkg/errors"
	apiv1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/client-go/kubernetes"
	klog "k8s.io/klog/v2"

	ocicommon "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/common"
)

var (
	// This mutex guarantees that multiple node pool actions aren't happening at the same time
	// Note that the actual wait for nodes to come up or delete is asynchronous.
	// This mutex is only around the api operations.
	nodePoolDeleteMutex sync.Mutex
)

// NodePool implements the NodeGroup interface via an OCI Node Pool
type NodePool interface {
	cloudprovider.NodeGroup
}

type nodePool struct {
	manager    NodePoolManager
	kubeClient kubernetes.Interface

	id      string
	minSize int
	maxSize int
}

// MaxSize returns maximum size of the node group.
func (np *nodePool) MaxSize() int {
	return np.maxSize
}

// MinSize returns minimum size of the node group.
func (np *nodePool) MinSize() int {
	return np.minSize
}

var defaultBackoff = wait.Backoff{
	Steps:    4,
	Duration: 10 * time.Millisecond,
	Factor:   5.0,
	Jitter:   0.1,
}

var retryableKubeError = func(err error) bool {
	// if it's a generic net timeout then retry
	if e, ok := err.(net.Error); ok && e.Timeout() {
		return true
	}

	// if it's a kube specific error we want to retry then retry.
	return kerrors.IsInternalError(err) ||
		kerrors.IsTimeout(err) ||
		kerrors.IsServerTimeout(err) ||
		kerrors.IsTooManyRequests(err) ||
		kerrors.IsServiceUnavailable(err) ||
		kerrors.IsUnauthorized(err)
}

// TargetSize returns the current target size of the node group. It is possible that the
// number of nodes in Kubernetes is different at the moment but should be equal
// to Size() once everything stabilizes (new nodes finish startup and registration or
// removed nodes are deleted completely). Implementation required.
func (np *nodePool) TargetSize() (int, error) {
	return np.manager.GetNodePoolSize(np)
}

// IncreaseSize increases the size of the node group. To delete a node you need
// to explicitly name it and use DeleteNode. This function should wait until
// node group size is updated. Implementation required.
func (np *nodePool) IncreaseSize(delta int) error {
	if delta <= 0 {
		return fmt.Errorf("size increase must be positive")
	}
	nodePoolDeleteMutex.Lock()
	defer nodePoolDeleteMutex.Unlock()

	size, err := np.manager.GetNodePoolSize(np)
	if err != nil {
		return err
	}

	if size+delta > np.MaxSize() {
		return fmt.Errorf("size increase too large - desired:%d max:%d", int(size)+delta, np.MaxSize())
	}

	return np.manager.SetNodePoolSize(np, size+delta)
}

// AtomicIncreaseSize is not implemented.
func (np *nodePool) AtomicIncreaseSize(delta int) error {
	return cloudprovider.ErrNotImplemented
}

// DeleteNodes deletes nodes from this node group. Error is returned either on
// failure or if the given node doesn't belong to this node group. This function
// should wait until node group size is updated. Implementation required.
func (np *nodePool) DeleteNodes(nodes []*apiv1.Node) (err error) {
	// Unregistered nodes come in as the provider id as node name.

	// although technically we only need the mutex around the api calls, we should wrap the mutex
	// around when we mark the node to be deleted as well. That way we don't mark a bunch of nodes
	// to be deleted, but have the scale down calls potentially happen seconds later.
	nodePoolDeleteMutex.Lock()
	defer nodePoolDeleteMutex.Unlock()

	klog.Infof("DeleteNodes called with %d nodes", len(nodes))

	size, err := np.manager.GetNodePoolSize(np)
	if err != nil {
		return err
	}

	klog.Infof("Nodepool %s has size %d", np.id, size)
	if int(size) <= np.MinSize() {
		return fmt.Errorf("min size reached, nodes will not be deleted")
	}

	refs := make([]ocicommon.OciRef, 0, len(nodes))

	// even though the nodes param is an array, in reality, nodes only contains a single node
	// Each node is deleted in its own DeleteNodes call, and all the calls are in parallel
	// we will still loop through just to future proof this function
	for _, node := range nodes {
		belongs, err := np.Belongs(node)
		if err != nil {
			return err
		}
		if !belongs {
			return fmt.Errorf("%s belong to a different nodepool than %s", node.Name, np.Id())
		}
		ociRef, err := ocicommon.NodeToOciRef(node)
		if err != nil {
			return err
		}

		refs = append(refs, ociRef)
	}

	if len(refs) == 0 {
		return nil
	}
	deleteInstancesErr := np.manager.DeleteInstances(np, refs)
	if deleteInstancesErr == nil {
		// this will add taints to all the nodes. For now, we have only a single node deleted in a given call, but the implementation might change in the future
		np.manager.TaintToPreventFurtherSchedulingOnRestart(nodes, np.kubeClient)
	} else {
		klog.Warning("Error deleting instances", deleteInstancesErr)
	}
	return deleteInstancesErr
}

// DecreaseTargetSize decreases the target size of the node group. This function
// doesn't permit to delete any existing node and can be used only to reduce the
// request for new nodes that have not been yet fulfilled. Delta should be negative.
// It is assumed that cloud provider will not delete the existing nodes when there
// is an option to just decrease the target. Implementation required.
func (np *nodePool) DecreaseTargetSize(delta int) error {
	if delta >= 0 {
		return fmt.Errorf("size decrease must be negative")
	}

	size, err := np.manager.GetNodePoolSize(np)
	if err != nil {
		return err
	}

	decreaseTargetCheckViaComputeString, ok := os.LookupEnv("DECREASE_TARGET_CHECK_VIA_COMPUTE")
	decreaseTargetCheckViaComputeBool := false
	if !ok {
		klog.V(5).Infof("DECREASE_TARGET_CHECK_VIA_COMPUTE is not present. Using GetNodePoolNodes to check non-terminated nodes")
	} else {
		if val, err := strconv.ParseBool(decreaseTargetCheckViaComputeString); err != nil {
			klog.V(5).Error(err, "Invalid value for environment variable DECREASE_TARGET_CHECK_VIA_COMPUTE set. Should be a boolean (true/false)")
		} else {
			decreaseTargetCheckViaComputeBool = val
		}
	}
	klog.V(4).Infof("DECREASE_TARGET_CHECK_VIA_COMPUTE: %v", decreaseTargetCheckViaComputeBool)
	var nodesLen int
	if decreaseTargetCheckViaComputeBool {
		nodesLen, err = np.manager.GetExistingNodePoolSizeViaCompute(np)
		if err != nil {
			klog.V(4).Error(err, "error while performing GetExistingNodePoolSizeViaCompute call")
			return err
		}
	} else {
		np.manager.InvalidateAndRefreshCache()
		nodes, err := np.manager.GetNodePoolNodes(np)
		if err != nil {
			klog.V(4).Error(err, "error while performing GetNodePoolNodes call")
			return err
		}
		nodesLen = len(nodes)
	}

	klog.V(5).Infof("Found %d non terminated/stopped nodes. size: %d, delta: %d", nodesLen, size, delta)

	if size+delta < nodesLen {
		return fmt.Errorf("attempt to delete existing nodes targetSize:%d delta:%d existingNodes: %d",
			size, delta, nodesLen)
	}

	return np.manager.SetNodePoolSize(np, size+delta)
}

// Belongs returns true if the given node belongs to the NodeGroup.
func (np *nodePool) Belongs(node *apiv1.Node) (bool, error) {
	ref, err := ocicommon.NodeToOciRef(node)
	if err != nil {
		return false, err
	}

	targetNodePool, err := np.manager.GetNodePoolForInstance(ref)
	if err != nil {
		return false, err
	}

	if targetNodePool == nil {
		return false, fmt.Errorf("%s doesn't belong to a known nodepool", node.Name)
	}

	return targetNodePool.Id() == np.Id(), nil
}

// Id returns an unique identifier of the node group.
func (np *nodePool) Id() string {
	return np.id
}

// Debug returns a string containing all information regarding this node group.
func (np *nodePool) Debug() string {
	return fmt.Sprintf("%s (%d:%d)", np.Id(), np.MinSize(), np.MaxSize())
}

// Nodes returns a list of all nodes that belong to this node group.
// It is required that Instance objects returned by this method have Id field set.
// Other fields are optional.
// This list should include also instances that might have not become a kubernetes node yet.
func (np *nodePool) Nodes() ([]cloudprovider.Instance, error) {
	return np.manager.GetNodePoolNodes(np)
}

// TemplateNodeInfo returns a framework.NodeInfo structure of an empty
// (as if just started) node. This will be used in scale-up simulations to
// predict what would a new node look like if a node group was expanded. The returned
// NodeInfo is expected to have a fully populated Node object, with all of the labels,
// capacity and allocatable information as well as all pods that are started on
// the node by default, using manifest (most likely only kube-proxy). Implementation optional.
func (np *nodePool) TemplateNodeInfo() (*framework.NodeInfo, error) {
	node, err := np.manager.GetNodePoolTemplateNode(np)
	if err != nil {
		return nil, errors.Wrap(err, "unable to build node pool template")
	}

	nodeInfo := framework.NewNodeInfo(
		node, nil,
		&framework.PodInfo{Pod: cloudprovider.BuildKubeProxy(np.id)},
		&framework.PodInfo{Pod: ocicommon.BuildFlannelPod()},
		&framework.PodInfo{Pod: ocicommon.BuildProxymuxClientPod()},
	)
	return nodeInfo, nil
}

// Exist checks if the node group really exists on the cloud provider side. Allows to tell the
// theoretical node group from the real one. Implementation required.
func (np *nodePool) Exist() bool {
	return true
}

// Create creates the node group on the cloud provider side. Implementation optional.
func (np *nodePool) Create() (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// Delete deletes the node group on the cloud provider side.
// This will be executed only for autoprovisioned node groups, once their size drops to 0.
// Implementation optional.
func (np *nodePool) Delete() error {
	return cloudprovider.ErrNotImplemented
}

// Autoprovisioned returns true if the node group is autoprovisioned. An autoprovisioned group
// was created by CA and can be deleted when scaled to 0.
func (np *nodePool) Autoprovisioned() bool {
	return false
}

// GetOptions returns NodeGroupAutoscalingOptions that should be used for this particular
// NodeGroup. Returning a nil will result in using default options.
// Implementation optional.
func (np *nodePool) GetOptions(defaults config.NodeGroupAutoscalingOptions) (*config.NodeGroupAutoscalingOptions, error) {
	return nil, cloudprovider.ErrNotImplemented
}
