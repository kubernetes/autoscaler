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

package clusters

import (
	"fmt"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huawei-cloud-sdk-go"
	"net/http"
)

// GetCCECluster calls CCE REST API to get cce cluster's information.
func GetCCECluster(client *huaweicloudsdk.ServiceClient, clusterId string) (r GetCCEClusterResult) {
	clusterURL := getClusterURL(client, clusterId)
	_, r.Err = client.Get(clusterURL, &r.Body, nil)
	return
}

// GetNodePools calls CCE REST API to get the information of all node pools in a designated cluster.
func GetNodePools(client *huaweicloudsdk.ServiceClient, clusterId string) (r GetCCENodePoolsResult) {
	nodePoolsURL := getNodePoolsURL(client, clusterId)
	_, r.Err = client.Get(nodePoolsURL, &r.Body, nil)
	return
}

// DeleteNode calls CCE REST API to delete a node with UID of nodeId in a designated cluster.
func DeleteNode(client *huaweicloudsdk.ServiceClient, clusterId string, nodeId string) (r DeleteNodeResult) {
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
func UpdateNodePool(client *huaweicloudsdk.ServiceClient, clusterId string, nodepoolId string, opt UpdateOptsBuilder) (r UpdateResult) {
	b, err := opt.ToClustersUpdateMap()
	if err != nil {
		r.Err = err
		return r
	}

	var result *http.Response

	result, r.Err = client.Put(updateNodePoolURL(client, clusterId, nodepoolId), b, &r.Body, &huaweicloudsdk.RequestOpts{
		OkCodes: []int{200, 202},
	})

	if r.Err == nil {
		r.Header = result.Header
	} else {
		fmt.Printf("Failed to Update NodePool Size huaweicloud cloud provider: %v\n", r.Err)
	}

	return
}
