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

package magnum

import (
	"fmt"
	"sync"
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	klog "k8s.io/klog/v2"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework/v1alpha1"
)

// magnumNodeGroup implements NodeGroup interface from cluster-autoscaler/cloudprovider.
//
// Represents a homogeneous collection of nodes within a cluster,
// which can be dynamically resized between a minimum and maximum
// number of nodes.
type magnumNodeGroup struct {
	magnumManager magnumManager
	id            string

	clusterUpdateMutex *sync.Mutex

	minSize int
	maxSize int
	// Stored as a pointer so that when autoscaler copies the nodegroup it can still update the target size
	targetSize *int

	nodesToDelete      []*apiv1.Node
	nodesToDeleteMutex sync.Mutex

	waitTimeStep        time.Duration
	deleteBatchingDelay time.Duration

	// Used so that only one DeleteNodes goroutine has to get the node group size at the start of the deletion
	deleteNodesCachedSize   int
	deleteNodesCachedSizeAt time.Time
}

// waitForClusterStatus checks periodically to see if the cluster has entered a given status.
// Returns when the status is observed or the timeout is reached.
func (ng *magnumNodeGroup) waitForClusterStatus(status string, timeout time.Duration) error {
	klog.V(2).Infof("Waiting for cluster %s status", status)
	for start := time.Now(); time.Since(start) < timeout; time.Sleep(ng.waitTimeStep) {
		clusterStatus, err := ng.magnumManager.getClusterStatus()
		if err != nil {
			return fmt.Errorf("error waiting for %s status: %v", status, err)
		}
		if clusterStatus == status {
			klog.V(0).Infof("Waited for cluster %s status", status)
			return nil
		}
	}
	return fmt.Errorf("timeout (%v) waiting for %s status", timeout, status)
}

// IncreaseSize increases the number of nodes by replacing the cluster's node_count.
//
// Takes precautions so that the cluster is not modified while in an UPDATE_IN_PROGRESS state.
// Blocks until the cluster has reached UPDATE_COMPLETE.
func (ng *magnumNodeGroup) IncreaseSize(delta int) error {
	ng.clusterUpdateMutex.Lock()
	defer ng.clusterUpdateMutex.Unlock()

	if delta <= 0 {
		return fmt.Errorf("size increase must be positive")
	}

	size, err := ng.magnumManager.nodeGroupSize(ng.id)
	if err != nil {
		return fmt.Errorf("could not check current nodegroup size: %v", err)
	}
	if size+delta > ng.MaxSize() {
		return fmt.Errorf("size increase too large, desired:%d max:%d", size+delta, ng.MaxSize())
	}

	updatePossible, currentStatus, err := ng.magnumManager.canUpdate()
	if err != nil {
		return fmt.Errorf("can not increase node count: %v", err)
	}
	if !updatePossible {
		return fmt.Errorf("can not add nodes, cluster is in %s status", currentStatus)
	}
	klog.V(0).Infof("Increasing size by %d, %d->%d", delta, *ng.targetSize, *ng.targetSize+delta)
	*ng.targetSize += delta

	err = ng.magnumManager.updateNodeCount(ng.id, *ng.targetSize)
	if err != nil {
		return fmt.Errorf("could not increase cluster size: %v", err)
	}

	// Block until cluster has gone into update status and then completed
	err = ng.waitForClusterStatus(clusterStatusUpdateInProgress, waitForUpdateStatusTimeout)
	if err != nil {
		return fmt.Errorf("wait for cluster status failed: %v", err)
	}
	err = ng.waitForClusterStatus(clusterStatusUpdateComplete, waitForCompleteStatusTimout)
	if err != nil {
		return fmt.Errorf("wait for cluster status failed: %v", err)
	}
	return nil
}

// deleteNodes deletes a set of nodes chosen by the autoscaler.
//
// The process of deletion depends on the implementation of magnumManager,
// but this function handles what should be common between all implementations:
//   - simultaneous but separate calls from the autoscaler are batched together
//   - does not allow scaling while the cluster is already in an UPDATE_IN_PROGRESS state
//   - after scaling down, blocks until the cluster has reached UPDATE_COMPLETE
func (ng *magnumNodeGroup) DeleteNodes(nodes []*apiv1.Node) error {

	// Batch simultaneous deletes on individual nodes
	ng.nodesToDeleteMutex.Lock()

	// First get the node group size and store the value, so that any other parallel delete calls can use it
	// without having to make the get request for the cluster themselves.
	// cachedSize keeps a local copy for this goroutine, so that ng.deleteNodesCachedSize is used
	// only within the ng.nodesToDeleteMutex.
	var cachedSize int
	var err error
	if time.Since(ng.deleteNodesCachedSizeAt) > time.Second*10 {
		cachedSize, err = ng.magnumManager.nodeGroupSize(ng.id)
		if err != nil {
			ng.nodesToDeleteMutex.Unlock()
			return fmt.Errorf("could not get current node count: %v", err)
		}
		ng.deleteNodesCachedSize = cachedSize
		ng.deleteNodesCachedSizeAt = time.Now()
	} else {
		cachedSize = ng.deleteNodesCachedSize
	}

	// Check that these nodes would not make the batch delete more nodes than the minimum would allow
	if cachedSize-len(ng.nodesToDelete)-len(nodes) < ng.MinSize() {
		ng.nodesToDeleteMutex.Unlock()
		return fmt.Errorf("deleting nodes would take nodegroup below minimum size")
	}
	// otherwise, add the nodes to the batch and release the lock
	ng.nodesToDelete = append(ng.nodesToDelete, nodes...)
	ng.nodesToDeleteMutex.Unlock()

	// The first of the parallel delete calls to obtain this lock will be the one to actually perform the deletion
	ng.clusterUpdateMutex.Lock()
	defer ng.clusterUpdateMutex.Unlock()

	ng.nodesToDeleteMutex.Lock()
	if len(ng.nodesToDelete) == 0 {
		// Deletion was handled by another goroutine
		ng.nodesToDeleteMutex.Unlock()
		return nil
	}
	ng.nodesToDeleteMutex.Unlock()

	// This goroutine has the clusterUpdateMutex, so will be the one
	// to actually delete the nodes. While this goroutine waits, others
	// will add their nodes to nodesToDelete and block at acquiring
	// the clusterUpdateMutex lock. One they get it, the deletion will
	// already be done and they will return above at the check
	// for len(ng.nodesToDelete) == 0.
	time.Sleep(ng.deleteBatchingDelay)

	ng.nodesToDeleteMutex.Lock()
	nodes = make([]*apiv1.Node, len(ng.nodesToDelete))
	copy(nodes, ng.nodesToDelete)
	ng.nodesToDelete = nil
	ng.nodesToDeleteMutex.Unlock()

	var nodeNames []string
	for _, node := range nodes {
		nodeNames = append(nodeNames, node.Name)
	}
	klog.V(1).Infof("Deleting nodes: %v", nodeNames)

	updatePossible, currentStatus, err := ng.magnumManager.canUpdate()
	if err != nil {
		return fmt.Errorf("could not check if cluster is ready to delete nodes: %v", err)
	}
	if !updatePossible {
		return fmt.Errorf("can not delete nodes, cluster is in %s status", currentStatus)
	}

	// Double check that the total number of batched nodes for deletion will not take the node group below its minimum size
	if cachedSize-len(nodes) < ng.MinSize() {
		return fmt.Errorf("size decrease too large, desired:%d min:%d", cachedSize-len(nodes), ng.MinSize())
	}

	var nodeRefs []NodeRef
	for _, node := range nodes {

		// Find node IPs, can be multiple (IPv4 and IPv6)
		var IPs []string
		for _, addr := range node.Status.Addresses {
			if addr.Type == apiv1.NodeInternalIP {
				IPs = append(IPs, addr.Address)
			}
		}
		nodeRefs = append(nodeRefs, NodeRef{
			Name:       node.Name,
			MachineID:  node.Status.NodeInfo.MachineID,
			ProviderID: node.Spec.ProviderID,
			IPs:        IPs,
		})
	}

	err = ng.magnumManager.deleteNodes(ng.id, nodeRefs, cachedSize-len(nodes))
	if err != nil {
		return fmt.Errorf("manager error deleting nodes: %v", err)
	}

	// Block until cluster has gone into update status and then completed
	err = ng.waitForClusterStatus(clusterStatusUpdateInProgress, waitForUpdateStatusTimeout)
	if err != nil {
		return fmt.Errorf("wait for cluster status failed: %v", err)
	}
	err = ng.waitForClusterStatus(clusterStatusUpdateComplete, waitForCompleteStatusTimout)
	if err != nil {
		return fmt.Errorf("wait for cluster status failed: %v", err)
	}

	// Check the new node group size and store that as the new target
	newSize, err := ng.magnumManager.nodeGroupSize(ng.id)
	if err != nil {
		// Set to the expected size as a fallback
		*ng.targetSize = cachedSize - len(nodes)
		return fmt.Errorf("could not check new cluster size after scale down: %v", err)
	}
	*ng.targetSize = newSize

	return nil
}

// DecreaseTargetSize decreases the cluster node_count in magnum.
func (ng *magnumNodeGroup) DecreaseTargetSize(delta int) error {
	if delta >= 0 {
		return fmt.Errorf("size decrease must be negative")
	}
	klog.V(0).Infof("Decreasing target size by %d, %d->%d", delta, *ng.targetSize, *ng.targetSize+delta)
	*ng.targetSize += delta
	return ng.magnumManager.updateNodeCount(ng.id, *ng.targetSize)
}

// Id returns the node group ID
func (ng *magnumNodeGroup) Id() string {
	return ng.id
}

// Debug returns a string formatted with the node group's min, max and target sizes.
func (ng *magnumNodeGroup) Debug() string {
	return fmt.Sprintf("%s min=%d max=%d target=%d", ng.id, ng.minSize, ng.maxSize, *ng.targetSize)
}

// Nodes returns a list of nodes that belong to this node group.
func (ng *magnumNodeGroup) Nodes() ([]cloudprovider.Instance, error) {
	nodes, err := ng.magnumManager.getNodes(ng.id)
	if err != nil {
		return nil, fmt.Errorf("could not get nodes: %v", err)
	}
	var instances []cloudprovider.Instance
	for _, node := range nodes {
		instances = append(instances, cloudprovider.Instance{Id: node})
	}
	return instances, nil
}

// TemplateNodeInfo returns a node template for this node group.
func (ng *magnumNodeGroup) TemplateNodeInfo() (*schedulerframework.NodeInfo, error) {
	return ng.magnumManager.templateNodeInfo(ng.id)
}

// Exist returns if this node group exists.
// Currently always returns true.
func (ng *magnumNodeGroup) Exist() bool {
	return true
}

// Create creates the node group on the cloud provider side.
func (ng *magnumNodeGroup) Create() (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrAlreadyExist
}

// Delete deletes the node group on the cloud provider side.
func (ng *magnumNodeGroup) Delete() error {
	return cloudprovider.ErrNotImplemented
}

// Autoprovisioned returns if the nodegroup is autoprovisioned.
func (ng *magnumNodeGroup) Autoprovisioned() bool {
	return false
}

// MaxSize returns the maximum allowed size of the node group.
func (ng *magnumNodeGroup) MaxSize() int {
	return ng.maxSize
}

// MinSize returns the minimum allowed size of the node group.
func (ng *magnumNodeGroup) MinSize() int {
	return ng.minSize
}

// TargetSize returns the target size of the node group.
func (ng *magnumNodeGroup) TargetSize() (int, error) {
	return *ng.targetSize, nil
}
