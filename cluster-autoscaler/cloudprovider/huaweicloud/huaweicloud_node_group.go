/*
Copyright 2020 The Kubernetes Authors.

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

package huaweicloud

import (
	"fmt"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	klog "k8s.io/klog/v2"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework/v1alpha1"
	"sync"
	"time"
)

const updateWaitTime = 10 * time.Second

// NodeGroup contains configuration info and functions to control a set of nodes that have the same capacity and set of labels.
// Represents a homogeneous collection of nodes within a cluster, which can be dynamically resized between a minimum and maximum number of nodes.
type NodeGroup struct {
	huaweiCloudManager *huaweicloudCloudManager
	nodePoolName       string
	nodePoolId         string
	clusterName        string
	minNodeCount       int
	maxNodeCount       int
	targetSize         *int
	autoscalingEnabled bool

	nodesToDelete       []*apiv1.Node
	deleteWaitTime      time.Duration
	timeIncrement       time.Duration
	nodePoolSizeTmp     int
	getNodePoolSizeTime time.Time

	clusterUpdateMutex *sync.Mutex
}

// MaxSize returns maximum size of the node group.
func (ng *NodeGroup) MaxSize() int {
	return ng.maxNodeCount
}

// MinSize returns minimum size of the node group.
func (ng *NodeGroup) MinSize() int {
	return ng.minNodeCount
}

// TargetSize returns the current target size of the node group. It is possible that the
// number of nodes in Kubernetes is different at the moment but should be equal
// to Size() once everything stabilizes (new nodes finish startup and registration or
// removed nodes are deleted completely).
func (ng *NodeGroup) TargetSize() (int, error) {
	return *ng.targetSize, nil
}

// waitForClusterStatus keeps waiting until the cluster has reached a specified status or timeout occurs.
func (ng *NodeGroup) waitForClusterStatus(status string, timeout time.Duration) error {
	klog.V(2).Infof("Waiting for cluster %s status", status)
	for start := time.Now(); time.Since(start) < timeout; time.Sleep(ng.timeIncrement) {
		clusterStatus, err := ng.huaweiCloudManager.getClusterStatus()
		if err != nil {
			return fmt.Errorf("error waiting for %s status: %v", status, err)
		}
		if clusterStatus == status {
			klog.V(0).Infof("Cluster has reached %s status", status)
			return nil
		}
	}
	return fmt.Errorf("timeout (%v) waiting for %s status", timeout, status)
}

// IncreaseSize increases the size of the node group. To delete a node you need
// to explicitly name it and use DeleteNode. This function should wait until
// node group size is updated.
func (ng *NodeGroup) IncreaseSize(delta int) error {
	ng.clusterUpdateMutex.Lock()
	defer ng.clusterUpdateMutex.Unlock()

	if delta <= 0 {
		return fmt.Errorf("delta for increasing size should be positive")
	}

	currentSize, err := ng.huaweiCloudManager.nodeGroupSize(ng.nodePoolName)
	if err != nil {
		return fmt.Errorf("failed to get the size of the node pool: %v", err)
	}
	if currentSize+delta > ng.MaxSize() {
		return fmt.Errorf("failed to increase the size of the node pool. target size above maximum. target size:%d, maximum size:%d", currentSize+delta, ng.MaxSize())
	}

	canUpdate, status, err := ng.huaweiCloudManager.canUpdate()

	if err != nil {
		return fmt.Errorf("failed to get status of the node pool: %v", err)
	}
	if !canUpdate {
		return fmt.Errorf("cluster is in %s status, cannot increase the size the cluster now", status)
	}
	klog.V(0).Infof("Increasing the size of the node pool by %d: %d->%d", delta, *ng.targetSize, *ng.targetSize+delta)
	*ng.targetSize += delta

	err = ng.huaweiCloudManager.updateNodeCount(ng, *ng.targetSize)
	if err != nil {
		return fmt.Errorf("failed to update the size of the node pool: %v", err)
	}

	// wait until cluster become available
	err = ng.waitForClusterStatus(clusterStatusAvailable, waitForCompleteStatusTimout)
	if err != nil {
		return fmt.Errorf("cluster failed to reach available status: %v", err)
	}

	// update internal cache
	*ng.targetSize = currentSize + delta

	return nil
}

// DeleteNodes deletes nodes from this node group. This function
// should wait until node group size is updated.
func (ng *NodeGroup) DeleteNodes(nodes []*apiv1.Node) error {
	// start the process of deleting nodes
	ng.clusterUpdateMutex.Lock()
	defer ng.clusterUpdateMutex.Unlock()

	// check whether deleting the nodes will cause the size of the node pool below minimum size
	// and update ng.nodesToDelete (as well as ng.nodePoolSizeTmp and ng.getNodePoolSizeTime if necessary)
	currentSize, err := checkAndUpdate(ng, nodes)
	if err != nil {
		return err
	}

	// get the unique ids of the nodes to be deleted
	nodeIDs, finished, err := getNodeIDsToDelete(ng)
	if finished { // nodes have been deleted by other goroutine
		return nil
	}

	if err != nil {
		return err
	}

	// call REST API to delete the nodes
	err = ng.huaweiCloudManager.deleteNodes(ng, nodeIDs, currentSize-len(nodes))
	if err != nil {
		return fmt.Errorf("fail to delete the nodes: %v", err)
	}

	// wait until cluster become available
	err = ng.waitForClusterStatus(clusterStatusAvailable, waitForCompleteStatusTimout)
	if err != nil {
		return fmt.Errorf("cluster failed to reach available status: %v", err)
	}

	// update ng.targetSize
	newSize, err := ng.huaweiCloudManager.nodeGroupSize(ng.nodePoolName)
	if err != nil {
		*ng.targetSize = currentSize - len(nodes)
		return fmt.Errorf("failed to get node pool's size after deleting the nodes: %v", err)
	}
	*ng.targetSize = newSize

	return nil
}

// checkAndUpdate checks whether deleting the nodes will cause the size of the node pool below minimum size,
// and updates ng.nodePoolSizeTmp, ng.getNodePoolSizeTime and ng.nodesToDelete if necessary.
func checkAndUpdate(ng *NodeGroup, nodes []*apiv1.Node) (int, error) {
	// currentSize is used to evaluate whether it's valid to delete the nodes. If the time since last update isn't
	// longer than updateWaitTime, ng.nodePoolSizeTmp will be used; otherwise, latest size of the node pool will be
	// obtained and used by currentSize.
	var currentSize int
	var err error
	if time.Since(ng.getNodePoolSizeTime) > updateWaitTime {
		currentSize, err = ng.huaweiCloudManager.nodeGroupSize(ng.nodePoolName)
		if err != nil {
			return 0, fmt.Errorf("failed to get current node pool's size: %v", err)
		}
		// update ng.nodePoolSizeTmp and ng.getNodePoolSizeTime
		ng.nodePoolSizeTmp = currentSize
		ng.getNodePoolSizeTime = time.Now()
	} else {
		// use ng.nodePoolSizeTmp if the time since last update isn't longer than updateWaitTime
		currentSize = ng.nodePoolSizeTmp
	}

	// evaluate whether it's valid to delete the nodes.
	// make sure that deleting the nodes won't cause the size of node pool below minimum size
	if currentSize-len(ng.nodesToDelete)-len(nodes) < ng.MinSize() {
		return 0, fmt.Errorf("cannot delete the nodes now since the size of the node pool isn't sufficient to retain minimum size")
	}

	// update ng.nodesToDelete
	ng.nodesToDelete = append(ng.nodesToDelete, nodes...)

	return currentSize, nil
}

// getNodeIDsToDelete checks whether there're still nodes waiting for being deleted. If there're no nodes
// to delete, it will return true, representing that the process of deleting the nodes has been finished;
// otherwise, it will return a slice of node ids to be deleted.
func getNodeIDsToDelete(ng *NodeGroup) ([]string, bool, error) {
	// check whether the nodes has already been deleted by other goroutine
	// If current goroutine is not the first one to acquire the ng.clusterUpdateMutex,
	// it's possible that the nodes have already been deleted, which makes ng.nodesToDelete to be empty.
	if len(ng.nodesToDelete) == 0 {
		return nil, true, nil
	}

	// check whether the cluster is available for deleting nodes
	canUpdate, status, err := ng.huaweiCloudManager.canUpdate()
	if err != nil {
		return nil, false, fmt.Errorf("failed to check whether the cluster is available for updating: %v", err)
	}
	if !canUpdate {
		return nil, false, fmt.Errorf("cluster is in %s status, cannot perform node deletion now", status)
	}

	nodeIDs := make([]string, 0)
	for _, node := range ng.nodesToDelete {
		// the node.Spec.ProviderID is the node's uid
		klog.Infof("Delete node with node id: %s", node.Spec.ProviderID)
		nodeIDs = append(nodeIDs, node.Spec.ProviderID)
	}

	ng.nodesToDelete = nil

	return nodeIDs, false, nil
}

// DecreaseTargetSize decreases the target size of the node group. This function
// doesn't permit to delete any existing node and can be used only to reduce the
// request for new nodes that have not been yet fulfilled. Delta should be negative.
// It is assumed that cloud provider will not delete the existing nodes when there
// is an option to just decrease the target.
func (ng *NodeGroup) DecreaseTargetSize(delta int) error {
	if delta >= 0 {
		return fmt.Errorf("delta for decreasing target size should be negative")
	}
	klog.V(0).Infof("Target size is decreased by %d, %d->%d", delta, *ng.targetSize, *ng.targetSize+delta)
	*ng.targetSize += delta
	return ng.huaweiCloudManager.updateNodeCount(ng, *ng.targetSize)
}

// Id returns the node pool's name.
func (ng *NodeGroup) Id() string {
	return ng.nodePoolName
}

// Debug returns a string containing information of the name, minimum size, maximum size and target size of the node pool
func (ng *NodeGroup) Debug() string {
	return fmt.Sprintf("%s min=%d max=%d target=%d", ng.nodePoolName, ng.minNodeCount, ng.maxNodeCount, ng.targetSize)
}

// Nodes returns a list of all nodes that belong to this node group.
// It is required that Instance objects returned by this method have Id field set.
// Other fields are optional.
// This list should include also instances that might have not become a kubernetes node yet.
func (ng *NodeGroup) Nodes() ([]cloudprovider.Instance, error) {
	nodes, err := ng.huaweiCloudManager.getNodes(ng.nodePoolName)
	if err != nil {
		return nil, fmt.Errorf("failed to get the nodes belong to the node pool: %v", err)
	}
	var instances []cloudprovider.Instance
	for _, node := range nodes {
		instances = append(instances, cloudprovider.Instance{Id: node})
	}
	return instances, nil
}

// TemplateNodeInfo returns a schedulerframework.NodeInfo structure of an empty
// (as if just started) node. This will be used in scale-up simulations to
// predict what would a new node look like if a node group was expanded. The returned
// NodeInfo is expected to have a fully populated Node object, with all of the labels,
// capacity and allocatable information as well as all pods that are started on
// the node by default, using manifest (most likely only kube-proxy). Not implemented
func (ng *NodeGroup) TemplateNodeInfo() (*schedulerframework.NodeInfo, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// Exist checks if the node group really exists on the cloud provider side. Currently always returns true.
func (ng *NodeGroup) Exist() bool {
	return true
}

// Create creates the node group on the cloud provider side. Not implemented.
func (ng *NodeGroup) Create() (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrAlreadyExist
}

// Delete deletes the node group on the cloud provider side. Not implemented.
func (ng *NodeGroup) Delete() error {
	return cloudprovider.ErrNotImplemented
}

// Autoprovisioned returns true if the node group is autoprovisioned. An autoprovisioned group
// was created by CA and can be deleted when scaled to 0. Currently always returns false.
func (ng *NodeGroup) Autoprovisioned() bool {
	return false
}
