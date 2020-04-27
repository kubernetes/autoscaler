package clusters

import (
	"fmt"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huawei-cloud-sdk-go"
	"net/http"
)

// GetCCECluster calls CCE REST API to get cce cluster's information.
func GetCCECluster(client *huawei_cloud_sdk_go.ServiceClient, clusterId string) (r GetCCEClusterResult) {
	clusterURL := getClusterURL(client, clusterId)
	_, r.Err = client.Get(clusterURL, &r.Body, nil)
	return
}

// GetNodePools calls CCE REST API to get the information of all node pools in a designated cluster.
func GetNodePools(client *huawei_cloud_sdk_go.ServiceClient, clusterId string) (r GetCCENodePoolsResult) {
	nodePoolsURL := getNodePoolsURL(client, clusterId)
	_, r.Err = client.Get(nodePoolsURL, &r.Body, nil)
	return
}

// DeleteNode calls CCE REST API to delete a node with UID of nodeId in a designated cluster.
func DeleteNode(client *huawei_cloud_sdk_go.ServiceClient, clusterId string, nodeId string) (r DeleteNodeResult) {
	deleteNodeURL := deleteNodeURL(client, clusterId, nodeId)
	_, r.Err = client.Delete(deleteNodeURL, nil)
	return 
}

// UpdateOptsBuilder is used to build request body for updating the size of a node pool.
// It allows extensions to add additional parameters to the Update request.
type UpdateOptsBuilder interface {
	ToClustersUpdateMap() (map[string]interface{}, error)
}

// UpdateNodePool calls CCE REST API to update the size of a node pool with UID of nodepoolID in a designated cluster.
func UpdateNodePool(client *huawei_cloud_sdk_go.ServiceClient, clusterId string, nodepoolId string,  opt UpdateOptsBuilder) (r UpdateResult) {
	b, err := opt.ToClustersUpdateMap()
	if err != nil {
		r.Err = err
		return r
	}

	var result *http.Response

	result, r.Err = client.Put(updateNodePoolURL(client, clusterId, nodepoolId), b, &r.Body, &huawei_cloud_sdk_go.RequestOpts{
		OkCodes: []int{200, 202},
	})

	if r.Err == nil {
		r.Header = result.Header
	} else {
		fmt.Printf("Failed to Update NodePool Size huaweicloud cloud provider: %v\n", r.Err)
	}

	return
}