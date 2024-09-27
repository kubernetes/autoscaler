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

package cherryservers

import (
	"fmt"
	"sync"
	"time"

	apiv1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	klog "k8s.io/klog/v2"
)

const (
	waitForStatusTimeStep       = 30 * time.Second
	waitForUpdateStatusTimeout  = 2 * time.Minute
	waitForCompleteStatusTimout = 10 * time.Minute
	scaleToZeroSupported        = true

	// Time that the goroutine that first acquires clusterUpdateMutex
	// in deleteNodes should wait for other synchronous calls to deleteNodes.
	deleteNodesBatchingDelay = 2 * time.Second
)

// cherryNodeGroup implements NodeGroup interface from cluster-autoscaler/cloudprovider.
//
// Represents a homogeneous collection of nodes within a cluster,
// which can be dynamically resized between a minimum and maximum
// number of nodes.
type cherryNodeGroup struct {
	cherryManager cherryManager
	id            string

	clusterUpdateMutex *sync.Mutex

	minSize int
	maxSize int
	// Stored as a pointer so that when autoscaler copies the nodegroup it can still update the target size
	targetSize int

	nodesToDelete      []*apiv1.Node
	nodesToDeleteMutex *sync.Mutex

	waitTimeStep        time.Duration
	deleteBatchingDelay time.Duration

	// Used so that only one DeleteNodes goroutine has to get the node group size at the start of the deletion
	deleteNodesCachedSize   int
	deleteNodesCachedSizeAt time.Time
}

func newCherryNodeGroup(manager cherryManager, name string, minSize, maxSize, targetSize int, wait, deleteBatching time.Duration) cherryNodeGroup {
	ng := cherryNodeGroup{
		cherryManager:       manager,
		id:                  name,
		clusterUpdateMutex:  &sync.Mutex{},
		nodesToDeleteMutex:  &sync.Mutex{},
		minSize:             minSize,
		maxSize:             maxSize,
		targetSize:          targetSize,
		waitTimeStep:        wait,
		deleteBatchingDelay: deleteBatching,
	}
	return ng
}

// IncreaseSize increases the number of nodes by replacing the cluster's node_count.
//
// Takes precautions so that the cluster is not modified while in an UPDATE_IN_PROGRESS state.
// Blocks until the cluster has reached UPDATE_COMPLETE.
func (ng *cherryNodeGroup) IncreaseSize(delta int) error {
	ng.clusterUpdateMutex.Lock()
	defer ng.clusterUpdateMutex.Unlock()

	if delta <= 0 {
		return fmt.Errorf("size increase must be positive")
	}

	size, err := ng.cherryManager.nodeGroupSize(ng.id)
	if err != nil {
		return fmt.Errorf("could not check current nodegroup size: %v", err)
	}
	if size+delta > ng.MaxSize() {
		return fmt.Errorf("size increase too large, desired:%d max:%d", size+delta, ng.MaxSize())
	}

	klog.V(0).Infof("Increasing size by %d, %d->%d", delta, ng.targetSize, ng.targetSize+delta)
	ng.targetSize += delta

	err = ng.cherryManager.createNodes(ng.id, delta)
	if err != nil {
		return fmt.Errorf("could not increase cluster size: %v", err)
	}

	return nil
}

// AtomicIncreaseSize is not implemented.
func (ng *cherryNodeGroup) AtomicIncreaseSize(delta int) error {
	return cloudprovider.ErrNotImplemented
}

// DeleteNodes deletes a set of nodes chosen by the autoscaler.
func (ng *cherryNodeGroup) DeleteNodes(nodes []*apiv1.Node) error {
	// Batch simultaneous deletes on individual nodes
	if err := ng.addNodesToDelete(nodes); err != nil {
		return err
	}
	cachedSize := ng.deleteNodesCachedSize

	// The first of the parallel delete calls to obtain this lock will be the one to actually perform the deletion
	ng.clusterUpdateMutex.Lock()
	defer ng.clusterUpdateMutex.Unlock()

	// This goroutine has the clusterUpdateMutex, so will be the one
	// to actually delete the nodes. While this goroutine waits, others
	// will add their nodes to nodesToDelete and block at acquiring
	// the clusterUpdateMutex lock. Once they get it, the deletion will
	// already be done and they will return above at the check
	// for len(ng.nodesToDelete) == 0.
	time.Sleep(ng.deleteBatchingDelay)

	nodes = ng.getNodesToDelete()
	if len(nodes) == 0 {
		// Deletion was handled by another goroutine
		return nil
	}

	var nodeNames []string
	for _, node := range nodes {
		nodeNames = append(nodeNames, node.Name)
	}

	// Double check that the total number of batched nodes for deletion will not take the node group below its minimum size
	if cachedSize-len(nodes) < ng.MinSize() {
		return fmt.Errorf("size decrease too large, desired:%d min:%d", cachedSize-len(nodes), ng.MinSize())
	}
	klog.V(0).Infof("Deleting nodes: %v", nodeNames)

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

	if err := ng.cherryManager.deleteNodes(ng.id, nodeRefs, cachedSize-len(nodes)); err != nil {
		return fmt.Errorf("manager error deleting nodes: %v", err)
	}

	// Check the new node group size and store that as the new target
	newSize, err := ng.cherryManager.nodeGroupSize(ng.id)
	if err != nil {
		// Set to the expected size as a fallback
		ng.targetSize = cachedSize - len(nodes)
		return fmt.Errorf("could not check new cluster size after scale down: %v", err)
	}
	ng.targetSize = newSize

	return nil
}

// getNodesToDelete safely gets all of the nodes added to the delete queue.
// "safely", as in it locks, gets and then releases the queue.
func (ng *cherryNodeGroup) getNodesToDelete() []*apiv1.Node {
	ng.nodesToDeleteMutex.Lock()
	defer ng.nodesToDeleteMutex.Unlock()
	nodes := make([]*apiv1.Node, len(ng.nodesToDelete))
	copy(nodes, ng.nodesToDelete)
	ng.nodesToDelete = nil
	return nodes
}

// addNodesToDelete safely adds nodes to the delete queue.
// "safely", as in it locks, adds, and then releases the queue.
func (ng *cherryNodeGroup) addNodesToDelete(nodes []*v1.Node) error {
	// Batch simultaneous deletes on individual nodes
	ng.nodesToDeleteMutex.Lock()
	defer ng.nodesToDeleteMutex.Unlock()

	// First get the node group size and store the value, so that any other parallel delete calls can use it
	// without having to make the get request themselves.
	// cachedSize keeps a local copy for this goroutine, so that ng.deleteNodesCachedSize is used
	// only within the ng.nodesToDeleteMutex.
	var (
		cachedSize int = ng.deleteNodesCachedSize
		err        error
	)
	// if the cache is more than 10 seconds old, refresh it
	if time.Since(ng.deleteNodesCachedSizeAt) > time.Second*10 {
		cachedSize, err = ng.cherryManager.nodeGroupSize(ng.id)
		if err != nil {
			return fmt.Errorf("could not get current node count: %v", err)
		}
		ng.deleteNodesCachedSize = cachedSize
		ng.deleteNodesCachedSizeAt = time.Now()
	}

	// Check that these nodes would not make the batch delete more nodes than the minimum would allow
	if cachedSize-len(ng.nodesToDelete)-len(nodes) < ng.MinSize() {
		return fmt.Errorf("deleting nodes would take nodegroup below minimum size %d", ng.minSize)
	}
	// otherwise, add the nodes to the batch and release the lock
	ng.nodesToDelete = append(ng.nodesToDelete, nodes...)

	return nil
}

// DecreaseTargetSize decreases the cluster node_count in Cherry Servers.
func (ng *cherryNodeGroup) DecreaseTargetSize(delta int) error {
	if delta >= 0 {
		return fmt.Errorf("size decrease must be negative")
	}
	klog.V(0).Infof("Decreasing target size by %d, %d->%d", delta, ng.targetSize, ng.targetSize+delta)
	ng.targetSize += delta
	return fmt.Errorf("could not decrease target size") /*ng.cherryManager.updateNodeCount(ng.id, ng.targetSize)*/
}

// Id returns the node group ID
func (ng *cherryNodeGroup) Id() string {
	return ng.id
}

// Debug returns a string formatted with the node group's min, max and target sizes.
func (ng *cherryNodeGroup) Debug() string {
	return fmt.Sprintf("%s min=%d max=%d target=%d", ng.id, ng.minSize, ng.maxSize, ng.targetSize)
}

// Nodes returns a list of nodes that belong to this node group.
func (ng *cherryNodeGroup) Nodes() ([]cloudprovider.Instance, error) {
	nodes, err := ng.cherryManager.getNodes(ng.id)
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
func (ng *cherryNodeGroup) TemplateNodeInfo() (*framework.NodeInfo, error) {
	return ng.cherryManager.templateNodeInfo(ng.id)
}

// Exist returns if this node group exists.
// Currently always returns true.
func (ng *cherryNodeGroup) Exist() bool {
	return true
}

// Create creates the node group on the cloud provider side.
func (ng *cherryNodeGroup) Create() (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrAlreadyExist
}

// Delete deletes the node group on the cloud provider side.
func (ng *cherryNodeGroup) Delete() error {
	return cloudprovider.ErrNotImplemented
}

// Autoprovisioned returns if the nodegroup is autoprovisioned.
func (ng *cherryNodeGroup) Autoprovisioned() bool {
	return false
}

// MaxSize returns the maximum allowed size of the node group.
func (ng *cherryNodeGroup) MaxSize() int {
	return ng.maxSize
}

// MinSize returns the minimum allowed size of the node group.
func (ng *cherryNodeGroup) MinSize() int {
	return ng.minSize
}

// TargetSize returns the target size of the node group.
func (ng *cherryNodeGroup) TargetSize() (int, error) {
	return ng.targetSize, nil
}

// GetOptions returns NodeGroupAutoscalingOptions that should be used for this particular
// NodeGroup. Returning a nil will result in using default options.
func (ng *cherryNodeGroup) GetOptions(defaults config.NodeGroupAutoscalingOptions) (*config.NodeGroupAutoscalingOptions, error) {
	return nil, cloudprovider.ErrNotImplemented
}
