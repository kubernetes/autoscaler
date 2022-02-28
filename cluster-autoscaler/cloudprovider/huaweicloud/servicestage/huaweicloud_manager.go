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
	"io"

	huaweicloudsdkcce "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/cce/v3"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/cce/v3/model"
	huaweicloudsdkecs "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ecs/v2"
	"github.com/pkg/errors"
	"gopkg.in/gcfg.v1"
	"k8s.io/autoscaler/cluster-autoscaler/config"

	"time"

	"k8s.io/apimachinery/pkg/util/sets"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	klog "k8s.io/klog/v2"
)

const (
	clusterStatusAvailable   = "Available"
	clusterStatusUnavailable = "Unavailable"
	// clusterStatusScalingUp   = "ScalingUp"
	clusterStatusScalingDown = "ScalingDown"
	clusterStatusCreating    = "Creating"
	clusterStatusDeleting    = "Deleting"
	clusterStatusUpgrading   = "Upgrading"
	clusterStatusResizing    = "Resizing"
	clusterStatusEmpty       = "Empty"
	clusterStatusHibernation = "Hibernation"

	waitForStatusTimeIncrement  = 30 * time.Second
	waitForUpdateStatusTimeout  = 2 * time.Minute
	waitForCompleteStatusTimout = 10 * time.Minute

	deleteOperationSucceedCode    = "Com.200"
	deleteOperationSucceedMessage = ""
	// deleteOperationFailure = "CCE_CM.0003"
)

// availableStatuses is a set of statuses that would prevent the cluster from successfully scaling.
var availableStatuses = sets.NewString(
	clusterStatusUnavailable,
	clusterStatusCreating,
	clusterStatusDeleting,
	clusterStatusUpgrading,
	clusterStatusResizing,
	clusterStatusEmpty,
	clusterStatusHibernation,
)

type huaweicloudCloudManager struct {
	clusterClient *huaweicloudsdkcce.CceClient
	ecsClient     *huaweicloudsdkecs.EcsClient
	clusterName   string // this is the id of the cluster
	timeIncrement time.Duration
}

func buildManager(configReader io.Reader, discoverOpts cloudprovider.NodeGroupDiscoveryOptions, opts config.AutoscalingOptions) (*huaweicloudCloudManager, error) {
	// read from cloud-config file
	var cfg CloudConfig
	if configReader != nil {
		if err := gcfg.ReadInto(&cfg, configReader); err != nil {
			klog.Errorf("failed to read from cloud-config file: %v", err)
			return nil, err
		}
	}

	if cfg.Global.AccessKey == "" {
		return nil, errors.New("couldn't find access key. An access key must be provided.")
	}
	if cfg.Global.SecretKey == "" {
		return nil, errors.New("couldn't find secret key. A secret key must be provided.")
	}
	// create Huawei CCE Client
	clusterClient := cfg.getCCEClient()
	ecsClient := cfg.getECSClient()

	manager := huaweicloudCloudManager{
		clusterClient: clusterClient,
		ecsClient:     ecsClient,
		clusterName:   opts.ClusterName,
		timeIncrement: waitForStatusTimeIncrement,
	}

	// make sure that the cluster exists
	showClusterRequest := &model.ShowClusterRequest{
		ClusterId: opts.ClusterName,
	}
	cceCluster, err := clusterClient.ShowCluster(showClusterRequest)

	if err != nil {
		return nil, fmt.Errorf("unable to access cluster (%s): %v", manager.clusterName, err)
	}

	if len(*cceCluster.Metadata.Uid) == 0 {
		return nil, fmt.Errorf("cluster: (%s): doesn't exist %v", manager.clusterName, err)
	}

	return &manager, nil
}

// nodeGroupSize gets the size of the node pool with the name of nodeGroupName attached to current cluster.
func (mgr *huaweicloudCloudManager) nodeGroupSize(nodeGroupName string) (int, error) {
	listNodePoolReq := &model.ListNodePoolsRequest{
		ClusterId: mgr.clusterName,
	}
	nodePools, err := mgr.clusterClient.ListNodePools(listNodePoolReq)
	if err != nil {
		return 0, fmt.Errorf("could not retrieve node pools from CCE cluster: %v", err)
	}
	for _, nodePool := range *nodePools.Items {
		if nodePool.Metadata.Name == nodeGroupName {
			return int(*nodePool.Status.CurrentNode), nil
		}
	}
	return 0, fmt.Errorf("could not find node pool with given name: %v", err)
}

/*
   updateNodeCount updates the the number of nodes in a certain node pool associated with current cluster.
   Request body of updating a node pool's size:
{
	"metadata": {
		"name": {NODE POOL NAME}
		"uid": {NODE POOL UID}
    },
	"spec": {
		"initialNodeCount": {TARGET NODE POOL SIZE}
		"autoscaling": {
			"enable": true,
			"minNodeCount": {MIN NODE COUNT}
			"maxNodeCount": {MAX NODE COUNT}
		}
	}
}
*/
// Call CCE REST API to change the size of a node pool.
func (mgr *huaweicloudCloudManager) updateNodeCount(nodepool *NodeGroup, nodes int) error {
	nodeNum := int32(nodes)
	nodepool.nodePoolSpec.InitialNodeCount = &nodeNum
	body := &model.NodePool{
		Metadata: &model.NodePoolMetadata{
			Name: nodepool.nodePoolName,
			Uid:  &nodepool.nodePoolId,
		},
		Spec: nodepool.nodePoolSpec,
	}

	updateNodePoolReq := &model.UpdateNodePoolRequest{
		ClusterId:  mgr.clusterName,
		NodepoolId: nodepool.nodePoolId,
		Body:       body,
	}
	_, err := mgr.clusterClient.UpdateNodePool(updateNodePoolReq)
	if err != nil {
		return fmt.Errorf("could not update the size of this node pool of current cluster: %v", err)
	}
	return nil
}

// getNodes is implemented currently (it should return the unique ids of all nodes in a nodegroup).
func (mgr *huaweicloudCloudManager) getNodes(nodegroup string) ([]string, error) {
	listNodeReq := &model.ListNodesRequest{
		ClusterId: mgr.clusterName,
	}
	allNodes, err := mgr.clusterClient.ListNodes(listNodeReq)
	if err != nil {
		klog.Errorf("failed to get nodes information of a cluster: %v\n", err)
		return []string{}, nil
	}
	res := make([]string, 0)
	for _, node := range *allNodes.Items {
		nodePoolUID := getNodePoolID(node)
		if nodePoolUID == nodegroup {
			res = append(res, *node.Metadata.Uid)
		}
	}
	return res, nil
}

// deleteNodes deletes nodes by passing the node pool, a list of ids of the nodes to be deleted,
// and simultaneously sets the node pool size to be updatedNodeCount.
func (mgr *huaweicloudCloudManager) deleteNodes(nodepool *NodeGroup, nodeIds []string, updatedNodeCount int) error {
	// Step 1: delete nodes by their UIDs
	err := mgr.deleteNodesHelper(nodeIds)

	if err != nil {
		return fmt.Errorf("delete nodes failed: %v", err)
	}

	// Step 2: wait for the cluster to scale down
	err = mgr.waitForClusterStatus(clusterStatusScalingDown, waitForUpdateStatusTimeout)
	if err != nil {
		fmt.Printf("cluster failed to reach %s status: %v. (May not be an error. There's a possibility that the cluster has already turned to Available status before this checking)", clusterStatusScalingDown, err)
	}

	// Step 3: wait for the cluster to be available
	err = mgr.waitForClusterStatus(clusterStatusAvailable, waitForCompleteStatusTimout)
	if err != nil {
		return fmt.Errorf("cluster failed to reach %s status: %v", clusterStatusAvailable, err)
	}

	//// Step 4: decrease the size of the node pool after deleting the nodes
	//err = mgr.updateNodeCount(nodepool, updatedNodeCount)
	//if err != nil {
	//	return fmt.Errorf("failed to update the size of the node pool attached to current cluster: %v", err)
	//}
	return nil
}

// deleteNodesHelper calls CCE REST API to remove a set of nodes from a cluster.
func (mgr *huaweicloudCloudManager) deleteNodesHelper(nodeIds []string) error {
	for _, nodeId := range nodeIds {
		deleteNodeReq := &model.DeleteNodeRequest{
			ClusterId: mgr.clusterName,
			NodeId:    nodeId,
		}
		_, err := mgr.clusterClient.DeleteNode(deleteNodeReq)
		if err != nil {
			return fmt.Errorf("delete nodes failed: %v", err)
		}
	}
	return nil
}

// waitForClusterStatus keeps waiting until the cluster has reached a specified status or timeout occurs.
func (mgr *huaweicloudCloudManager) waitForClusterStatus(status string, timeout time.Duration) error {
	klog.V(2).Infof("Waiting for cluster to reach %s status", status)
	for start := time.Now(); time.Since(start) < timeout; time.Sleep(mgr.timeIncrement) {
		currentStatus, err := mgr.getClusterStatus()
		if err != nil {
			return fmt.Errorf("cluster failed to reach %v status: %v", status, err)
		}
		if currentStatus == status {
			klog.V(0).Infof("cluster has reached %s status", status)
			return nil
		}
	}
	return fmt.Errorf("waited for %v. timeout when waiting for cluster status %s", timeout, status)
}

// getClusterStatus returns the current status of the cce cluster.
func (mgr *huaweicloudCloudManager) getClusterStatus() (string, error) {
	// make sure that the cluster exists
	showClusterRequest := &model.ShowClusterRequest{
		ClusterId: mgr.clusterName,
	}
	cluster, err := mgr.clusterClient.ShowCluster(showClusterRequest)
	if err != nil {
		return "", fmt.Errorf("could not get cluster: %v", err)
	}
	return *cluster.Status.Phase, nil
}

// canUpdate returns true if the status of current status is not one of the statuses that prevent the cluster from being updated.
func (mgr *huaweicloudCloudManager) canUpdate() (bool, string, error) {
	clusterStatus, err := mgr.getClusterStatus()
	if err != nil {
		return false, "", fmt.Errorf("could not get cluster status: %v", err)
	}
	return !availableStatuses.Has(clusterStatus), clusterStatus, nil
}
