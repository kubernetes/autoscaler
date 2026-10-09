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

package sdk

import (
	"context"
	"fmt"

	"k8s.io/klog/v2"
)

// NodePool represents a VKE node group.
type NodePool struct {
	ID string `json:"node_group_uuid"`

	Name   string `json:"node_group_name"`
	Flavor string `json:"node_flavor_uuid"`
	Status string `json:"node_groups_status"`

	MinNodes     uint32 `json:"node_group_min_size"`
	MaxNodes     uint32 `json:"node_group_max_size"`
	CurrentNodes int    `json:"current_nodes"`
}

// ListNodePools allows to list all node pools available in a cluster
func (c *Client) ListNodePools(ctx context.Context, clusterID string) ([]NodePool, error) {
	klog.V(2).Infof("Listing node pools for cluster %s", clusterID)
	nodepools := make([]NodePool, 0)
	err := c.CallAPIWithContext(
		ctx,
		"GET",
		fmt.Sprintf("/cluster/%s/nodegroups", clusterID),
		nil,
		&nodepools,
		nil,
		nil,
		true,
	)
	return nodepools, err
}

// GetNodePool allows to display information for a specific node pool
func (c *Client) GetNodePool(ctx context.Context, clusterID string, poolID string) (*NodePool, error) {
	nodepool := &NodePool{}

	return nodepool, c.CallAPIWithContext(
		ctx,
		"GET",
		fmt.Sprintf("/cluster/%s/nodepool/%s", clusterID, poolID),
		nil,
		&nodepool,
		nil,
		nil,
		true,
	)
}

// ListNodePoolNodes allows to display nodes contained in a parent node pool
func (c *Client) ListNodePoolNodes(ctx context.Context, clusterID string, poolID string) ([]Node, error) {
	nodes := make([]Node, 0)

	return nodes, c.CallAPIWithContext(
		ctx,
		"GET",
		fmt.Sprintf("/cluster/%s/nodegroups/%s/nodes", clusterID, poolID),
		nil,
		&nodes,
		nil,
		nil,
		true,
	)
}

// DeleteNode deletes a node from a node group.
func (c *Client) DeleteNode(ctx context.Context, clusterID, nodeGroupID, id string) error {
	klog.V(2).Infof("Deleting node %s from cluster %s", id, clusterID)
	return c.CallAPIWithContext(
		ctx,
		"DELETE",
		fmt.Sprintf("/cluster/%s/nodegroups/%s/nodes/%s", clusterID, nodeGroupID, id),
		nil,
		nil,
		nil,
		nil,
		true,
	)
}

// AddNode adds a node to a node group.
func (c *Client) AddNode(ctx context.Context, clusterID, nodeGroupID string) (*Node, error) {
	klog.V(2).Infof("Adding node to cluster %s", clusterID)
	node := &Node{}
	return node, c.CallAPIWithContext(
		ctx,
		"PUT",
		fmt.Sprintf("/cluster/%s/nodegroups/%s/nodes/add", clusterID, nodeGroupID),
		nil,
		node,
		nil,
		nil,
		true,
	)
}
