package clusters

import "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huawei-cloud-sdk-go"

const (
	clusterDenoter  = "clusters"
	nodepoolDenoter = "nodepools"
	nodeDenoter     = "nodes"
)

// getClusterURL returns a URL for getting the information of a cce cluster with designated clusterID.
// REST API:
// 		GET /api/v3/projects/{project_id}/clusters/{cluster_id}
// Example:
// 		https://cce.cn-north-1.myhuaweicloud.com/api/v3/projects/017a290a8242480e82de8db804c1718d/clusters/19d4f935-4c45-11ea-b2e7-0255ac101eee
func getClusterURL(sc *huawei_cloud_sdk_go.ServiceClient, clusterID string) string {
	return sc.ServiceURL(clusterDenoter, clusterID)
}

// getNodePoolsURL returns a URL for getting the information of all nodepools attached to a cce cluster with designated clusterID.
// REST API:
//		GET /api/v3/projects/{project_id}/clusters/{cluster_id}/nodepools
// Example:
// 		https://cce.cn-north-4.myhuaweicloud.com/api/v3/projects/9071a38e7f6a4ba7b7bcbeb7d4ea6efc/clusters/25b884ab-623f-11ea-9981-0255ac101546/nodepools
func getNodePoolsURL(sc *huawei_cloud_sdk_go.ServiceClient, clusterID string) string {
	return sc.ServiceURL(clusterDenoter, clusterID, nodepoolDenoter)
}

// deleteNodeURL returns a URL for deleting a node with uid of nodeId in a cce cluster with clusterID.
// REST API:
// 		DELETE /api/v3/projects/{project_id}/clusters/{cluster_id}/nodes/{node_id}
// Example:
// 		https://cce.cn-north-4.myhuaweicloud.com/api/v3/projects/9071a38e7f6a4ba7b7bcbeb7d4ea6efc/clusters/25b884ab-623f-11ea-9981-0255ac101546/nodes/56a90712-6306-11ea-9981-0255ac101546
func deleteNodeURL(sc *huawei_cloud_sdk_go.ServiceClient, clusterID string, nodeId string) string {
	return sc.ServiceURL(clusterDenoter, clusterID, nodeDenoter, nodeId)
}

// updateNodePoolURL returns a URL for updating the number of nodes in a node pool with id nodepoolID which attached to
// a cce cluster with id clusterID.
// REST API:
// 		PUT /api/v3/projects/{project_id}/clusters/{cluster_id}/nodepools/{nodepool_id}
// Example:
// 		PUT https://cce.cn-north-4.myhuaweicloud.com/api/v3/projects/9071a38e7f6a4ba7b7bcbeb7d4ea6efc/clusters/25b884ab-623f-11ea-9981-0255ac101546/nodepools/8445aeed-6240-11ea-a1c6-0255ac101d44
func updateNodePoolURL(sc *huawei_cloud_sdk_go.ServiceClient, clusterID string, nodepoolID string) string {
	return sc.ServiceURL(clusterDenoter, clusterID, nodepoolDenoter, nodepoolID)
}