/*
Copyright The Kubernetes Authors.

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

package sdk

// ResourceMetadata mirrors nebius.common.v1.ResourceMetadata.
type ResourceMetadata struct {
	ID              string
	ParentID        string
	Name            string
	ResourceVersion int64
	Labels          map[string]string
	raw             []byte // preserved for round-trip
}

// NodeGroup mirrors nebius.mk8s.v1.NodeGroup.
type NodeGroup struct {
	Metadata *ResourceMetadata
	Spec     *NodeGroupSpec
	Status   *NodeGroupStatus
}

// NodeGroupSpec mirrors nebius.mk8s.v1.NodeGroupSpec.
type NodeGroupSpec struct {
	Version string

	// Size: exactly one of FixedNodeCount or Autoscaling is set.
	FixedNodeCount *int64
	Autoscaling    *NodeGroupAutoscalingSpec

	// Opaque fields preserved for round-trip during updates.
	templateRaw  []byte
	strategyRaw  []byte
	autoRepairRaw []byte
}

// NodeGroupAutoscalingSpec mirrors nebius.mk8s.v1.NodeGroupAutoscalingSpec.
type NodeGroupAutoscalingSpec struct {
	MinNodeCount int64
	MaxNodeCount int64
}

// NodeGroupStatus mirrors nebius.mk8s.v1.NodeGroupStatus.
type NodeGroupStatus struct {
	TargetNodeCount int64
}

// Instance mirrors nebius.compute.v1.Instance.
type Instance struct {
	Metadata *ResourceMetadata
}

// ListNodeGroupsRequest mirrors nebius.mk8s.v1.ListNodeGroupsRequest.
type ListNodeGroupsRequest struct {
	ParentID  string
	PageToken string
}

// ListNodeGroupsResponse mirrors nebius.mk8s.v1.ListNodeGroupsResponse.
type ListNodeGroupsResponse struct {
	Items         []*NodeGroup
	NextPageToken string
}

// GetNodeGroupRequest mirrors nebius.mk8s.v1.GetNodeGroupRequest.
type GetNodeGroupRequest struct {
	ID string
}

// UpdateNodeGroupRequest mirrors nebius.mk8s.v1.UpdateNodeGroupRequest.
type UpdateNodeGroupRequest struct {
	Metadata *ResourceMetadata
	Spec     *NodeGroupSpec
}

// ListInstancesRequest mirrors nebius.compute.v1.ListInstancesRequest.
type ListInstancesRequest struct {
	ParentID  string
	PageToken string
}

// ListInstancesResponse mirrors nebius.compute.v1.ListInstancesResponse.
type ListInstancesResponse struct {
	Items         []*Instance
	NextPageToken string
}

// DeleteInstanceRequest mirrors nebius.compute.v1.DeleteInstanceRequest.
type DeleteInstanceRequest struct {
	ID string
}
