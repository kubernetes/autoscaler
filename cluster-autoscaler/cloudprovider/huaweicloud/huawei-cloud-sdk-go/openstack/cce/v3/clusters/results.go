package clusters

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huawei-cloud-sdk-go"
)

// CCE model
type CCECluster struct {
	Kind       string   `json:"kind"`
	APIVersion string   `json:"apiVersion"`
	Metadata   Metadata `json:"metadata"`
	Status     Status   `json:"status"`
}

// Node pool model
type NodePool struct {
	Kind           string         `json:"kind"`
	APIVersion     string         `json:"apiVersion"`
	Metadata       Metadata       `json:"metadata"`
	Spec           Spec           `json:"spec"`
	NodePoolStatus NodePoolStatus `json:"status"`
}

// Node pools model
type NodePools struct {
	Kind       string     `json:"kind"`
	APIVersion string     `json:"apiVersion"`
	Items      []NodePool `json:"items"`
}

// Node model
type CCENode struct {
	Kind       string   `json:"kind"`
	APIVersion string   `json:"apiVersion"`
	Metadata   Metadata `json:"metadata"`
	Spec       Spec     `json:"spec"`
	Status     Status   `json:"status"`
}

// Request body for updating a node pool
type RequestBody struct {
	Metadata Metadata `json:"metadata"`
	Spec     Spec     `json:"spec"`
}

type Status struct {
	Phase     string `json:"phase"`
	JobId     string `json:"jobId"`
	ServerId  string `json:"serverId"`
	PrivateIP string `json:"privateIP"`
	PublicIP  string `json:"publicIP"`
}

// "metadata" section for CCE cluster and node pool
type Metadata struct {
	Name string `json:"name"`
	Uid  string `json:"uid"`
}

// "autoscaling" section in the request body of updating a node pool
type Autoscaling struct {
	Enable       bool `json:"enable"`
	MinNodeCount int  `json:"minNodeCount"`
	MaxNodeCount int  `json:"maxNodeCount"`
}

// "spec" section in the request body of updating a node pool
type Spec struct {
	InitialNodeCount int         `json:"initialNodeCount"`
	Autoscaling      Autoscaling `json:"autoscaling"`
}

// "status" section of a node pool
type NodePoolStatus struct {
	CurrentNode int    `json:"currentNode"`
	Phase       string `json:"phase"`
}

// ActionResult represents the result of server action operations, like reboot.
// Call its ExtractErr method to determine if the action succeeded or failed.
type ActionResult struct {
	huawei_cloud_sdk_go.ErrResult
}

type ChangeResult struct {
	huawei_cloud_sdk_go.Result
}

// Cluster
type GetCCEClusterResult struct {
	huawei_cloud_sdk_go.Result
}

func (r GetCCEClusterResult) Extract() (CCECluster, error) {
	var res CCECluster
	err := r.ExtractInto(&res)
	return res, err
}

// NodePools
type GetCCENodePoolsResult struct {
	huawei_cloud_sdk_go.Result
}

// UpdateResult is the response of a Update operation.
type UpdateResult struct {
	huawei_cloud_sdk_go.Result
}

func (r UpdateResult) Extract() (string, error) {
	var s struct {
		UUID string
	}
	err := r.ExtractInto(&s)
	return s.UUID, err
}

func (r GetCCENodePoolsResult) Extract() (NodePools, error) {
	var res NodePools
	err := r.ExtractInto(&res)
	return res, err
}

// DeleteNodeResult is the error section in the result of deleting a node.
type DeleteNodeResult struct {
	huawei_cloud_sdk_go.ErrResult
}

// ToClusterUpdateMap is a method that has to be implemented by RequestBody for interface UpdateOptsBuilder.
// This method assembles a request body based on the contents of opts(type RequestBody).
func (opts RequestBody) ToClustersUpdateMap() (map[string]interface{}, error) {
	return huawei_cloud_sdk_go.BuildRequestBody(opts, "")
}

// ToClusterUpdateMap is a method that has to be implemented by Metadata for interface UpdateOptsBuilder.
// This method assembles a request body based on the contents of opts(type Metadata).
func (opts Metadata) ToClustersUpdateMap() (map[string]interface{}, error) {
	return huawei_cloud_sdk_go.BuildRequestBody(opts, "")
}

// ToClusterUpdateMap is a method that has to be implemented by Autoscaling for interface UpdateOptsBuilder.
// This method assembles a request body based on the contents of opts(type Autoscaling).
func (opts Autoscaling) ToClustersUpdateMap() (map[string]interface{}, error) {
	return huawei_cloud_sdk_go.BuildRequestBody(opts, "")
}