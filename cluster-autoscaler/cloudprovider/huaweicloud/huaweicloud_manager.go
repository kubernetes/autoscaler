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
	"github.com/pkg/errors"
	"gopkg.in/gcfg.v1"
	"io"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huawei-cloud-sdk-go"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huawei-cloud-sdk-go/openstack"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huawei-cloud-sdk-go/openstack/cce/v3/clusters"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/version"

	"k8s.io/apimachinery/pkg/util/sets"
	"time"

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
	clusterClient *huaweicloudsdk.ServiceClient
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

	akskOptions := toAKSKOptions(cfg)

	// create the authenticated provider client.
	provider, authErr := openstack.AuthenticatedClient(akskOptions)

	if authErr != nil {
		fmt.Println("Failed to get the AuthenticatedClient for Huawei Cloud: ", authErr)
		return nil, authErr
	}

	userAgent := huaweicloudsdk.UserAgent{}
	userAgent.Prepend(fmt.Sprintf("cluster-autoscaler/%s", version.ClusterAutoscalerVersion))
	userAgent.Prepend(fmt.Sprintf("cluster/%s", opts.ClusterName))
	provider.UserAgent = userAgent
	klog.V(5).Infof("Using user-agent %s", userAgent.Join())

	// create Huawei CCE Client
	clusterClient, clientErr := openstack.NewCCEV3(provider, huaweicloudsdk.EndpointOpts{})
	if clientErr != nil {
		fmt.Println("Failed to get the CCEV3 client: ", clientErr)
		return nil, clientErr
	}

	manager := huaweicloudCloudManager{
		clusterClient: clusterClient,
		clusterName:   opts.ClusterName,
		timeIncrement: waitForStatusTimeIncrement,
	}

	// add more headers. without this setting, you will get "not suppport content type issue"
	headerMap := make(map[string]string)
	headerMap["Content-Type"] = "application/json;charset=utf-8"
	clusterClient.MoreHeaders = headerMap

	// make sure that the cluster exists
	cceCluster, err := clusters.GetCCECluster(clusterClient, opts.ClusterName).Extract()

	if err != nil {
		return nil, fmt.Errorf("unable to access cluster (%s): %v", manager.clusterName, err)
	}

	if len(cceCluster.Metadata.Uid) == 0 {
		return nil, fmt.Errorf("cluster: (%s): doesn't exist %v", manager.clusterName, err)
	}

	return &manager, nil
}

// nodeGroupSize gets the size of the node pool with the name of nodeGroupName attached to current cluster.
func (mgr *huaweicloudCloudManager) nodeGroupSize(nodeGroupName string) (int, error) {
	nodePools, err := clusters.GetNodePools(mgr.clusterClient, mgr.clusterName).Extract()
	if err != nil {
		return 0, fmt.Errorf("could not retrieve node pools from CCE cluster: %v", err)
	}
	for _, nodePool := range nodePools.Items {
		if nodePool.Metadata.Name == nodeGroupName {
			return nodePool.NodePoolStatus.CurrentNode, nil
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
	updateOpts := clusters.RequestBody{
		Metadata: clusters.Metadata{
			Name: nodepool.nodePoolName,
			Uid:  nodepool.nodePoolId,
		},
		Spec: clusters.Spec{
			InitialNodeCount: nodes,
			Autoscaling: clusters.Autoscaling{
				Enable:       true,
				MinNodeCount: nodepool.minNodeCount,
				MaxNodeCount: nodepool.maxNodeCount,
			},
		},
	}

	_, err := clusters.UpdateNodePool(mgr.clusterClient, mgr.clusterName, nodepool.nodePoolId, updateOpts).Extract()
	if err != nil {
		return fmt.Errorf("could not update the size of this node pool of current cluster: %v", err)
	}
	return nil
}

// getNodes is not implemented currently (it should return the unique ids of all nodes in a nodegroup).
func (mgr *huaweicloudCloudManager) getNodes(nodegroup string) ([]string, error) {
	return []string{}, nil
}

// deleteNodes deletes nodes by passing the node pool, a list of ids of the nodes to be deleted,
// and simultaneously sets the node pool size to be updatedNodeCount.
func (mgr *huaweicloudCloudManager) deleteNodes(nodepool *NodeGroup, nodeIds []string, updatedNodeCount int) error {
	// Step 1: delete nodes by their UIDs
	err := mgr.deleteNodesHelper(mgr.clusterClient, mgr.clusterName, nodeIds)

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

	// Step 4: decrease the size of the node pool after deleting the nodes
	err = mgr.updateNodeCount(nodepool, updatedNodeCount)
	if err != nil {
		return fmt.Errorf("failed to update the size of the node pool attached to current cluster: %v", err)
	}
	return nil
}

// deleteNodesHelper calls CCE REST API to remove a set of nodes from a cluster.
func (mgr *huaweicloudCloudManager) deleteNodesHelper(client *huaweicloudsdk.ServiceClient, clusterId string, nodeIds []string) error {
	for _, nodeId := range nodeIds {
		err := clusters.DeleteNode(client, clusterId, nodeId).ExtractErr()
		if err != nil {
			/*
				Due to the CCE DELETE Node API Response Issue, we need to do special handling here.
				The CCE issue is:  Even DELETE operation succeeds, CCE still return something like this: {"ErrorCode":"Com.200","Message":""}
				Example response of real DELETE operation failure: {"ErrorCode":"CCE_CM.0003","Message":"Resource not found"}

			*/
			ue := err.(*huaweicloudsdk.UnifiedError)
			if ue.ErrorCode() == deleteOperationSucceedCode && ue.Message() == deleteOperationSucceedMessage {
				continue
			}

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
	cluster, err := clusters.GetCCECluster(mgr.clusterClient, mgr.clusterName).Extract()
	if err != nil {
		return "", fmt.Errorf("could not get cluster: %v", err)
	}
	return cluster.Status.Phase, nil
}

// canUpdate returns true if the status of current status is not one of the statuses that prevent the cluster from being updated.
func (mgr *huaweicloudCloudManager) canUpdate() (bool, string, error) {
	clusterStatus, err := mgr.getClusterStatus()
	if err != nil {
		return false, "", fmt.Errorf("could not get cluster status: %v", err)
	}
	return !availableStatuses.Has(clusterStatus), clusterStatus, nil
}
