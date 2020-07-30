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
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huawei-cloud-sdk-go"
)

// CCECluster CCE model
type CCECluster struct {
	Kind       string   `json:"kind"`
	APIVersion string   `json:"apiVersion"`
	Metadata   Metadata `json:"metadata"`
	Status     Status   `json:"status"`
}

// NodePool Node pool model
type NodePool struct {
	Kind           string         `json:"kind"`
	APIVersion     string         `json:"apiVersion"`
	Metadata       Metadata       `json:"metadata"`
	Spec           Spec           `json:"spec"`
	NodePoolStatus NodePoolStatus `json:"status"`
}

// NodePools Node pools model
type NodePools struct {
	Kind       string     `json:"kind"`
	APIVersion string     `json:"apiVersion"`
	Items      []NodePool `json:"items"`
}

// CCENode Node model
type CCENode struct {
	Kind       string   `json:"kind"`
	APIVersion string   `json:"apiVersion"`
	Metadata   Metadata `json:"metadata"`
	Spec       Spec     `json:"spec"`
	Status     Status   `json:"status"`
}

// RequestBody Request body for updating a node pool
type RequestBody struct {
	Metadata Metadata `json:"metadata"`
	Spec     Spec     `json:"spec"`
}

// Status "status" section for CCE Node
type Status struct {
	Phase     string `json:"phase"`
	JobId     string `json:"jobId"`
	ServerId  string `json:"serverId"`
	PrivateIP string `json:"privateIP"`
	PublicIP  string `json:"publicIP"`
}

// Metadata "metadata" section for CCE cluster and node pool
type Metadata struct {
	Name string `json:"name"`
	Uid  string `json:"uid"`
}

// Autoscaling "autoscaling" section in the request body of updating a node pool
type Autoscaling struct {
	Enable       bool `json:"enable"`
	MinNodeCount int  `json:"minNodeCount"`
	MaxNodeCount int  `json:"maxNodeCount"`
}

// Spec "spec" section in the request body of updating a node pool
type Spec struct {
	InitialNodeCount int         `json:"initialNodeCount"`
	Autoscaling      Autoscaling `json:"autoscaling"`
}

// NodePoolStatus "status" section of a node pool
type NodePoolStatus struct {
	CurrentNode int    `json:"currentNode"`
	Phase       string `json:"phase"`
}

// ActionResult represents the result of server action operations, like reboot.
// Call its ExtractErr method to determine if the action succeeded or failed.
type ActionResult struct {
	huaweicloudsdk.ErrResult
}

// ChangeResult ...
type ChangeResult struct {
	huaweicloudsdk.Result
}

// GetCCEClusterResult for Cluster
type GetCCEClusterResult struct {
	huaweicloudsdk.Result
}

// Extract cluster info and error
func (r GetCCEClusterResult) Extract() (CCECluster, error) {
	var res CCECluster
	err := r.ExtractInto(&res)
	return res, err
}

// GetCCENodePoolsResult for NodePools
type GetCCENodePoolsResult struct {
	huaweicloudsdk.Result
}

// UpdateResult is the response of a Update operation.
type UpdateResult struct {
	huaweicloudsdk.Result
}

// Extract UUID and error
func (r UpdateResult) Extract() (string, error) {
	var s struct {
		UUID string
	}
	err := r.ExtractInto(&s)
	return s.UUID, err
}

// Extract node pools and error
func (r GetCCENodePoolsResult) Extract() (NodePools, error) {
	var res NodePools
	err := r.ExtractInto(&res)
	return res, err
}

// DeleteNodeResult is the error section in the result of deleting a node.
type DeleteNodeResult struct {
	huaweicloudsdk.ErrResult
}

// ToClustersUpdateMap is a method that has to be implemented by RequestBody for interface UpdateOptsBuilder.
// This method assembles a request body based on the contents of opts(type RequestBody).
func (opts RequestBody) ToClustersUpdateMap() (map[string]interface{}, error) {
	return huaweicloudsdk.BuildRequestBody(opts, "")
}

// ToClustersUpdateMap is a method that has to be implemented by Metadata for interface UpdateOptsBuilder.
// This method assembles a request body based on the contents of opts(type Metadata).
func (opts Metadata) ToClustersUpdateMap() (map[string]interface{}, error) {
	return huaweicloudsdk.BuildRequestBody(opts, "")
}

// ToClustersUpdateMap is a method that has to be implemented by Autoscaling for interface UpdateOptsBuilder.
// This method assembles a request body based on the contents of opts(type Autoscaling).
func (opts Autoscaling) ToClustersUpdateMap() (map[string]interface{}, error) {
	return huaweicloudsdk.BuildRequestBody(opts, "")
}
