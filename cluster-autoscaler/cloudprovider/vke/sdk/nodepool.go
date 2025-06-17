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
	for i, nodepool := range nodepools {
		if uint32(nodepool.CurrentNodes) < nodepools[i].MinNodes {
			delta := nodepools[i].MinNodes - uint32(nodepool.CurrentNodes)
			for j := 0; j < int(delta); j++ {
				_, err := c.AddNode(ctx, clusterID, nodepool.ID)
				if err != nil {
					klog.Errorf("Failed to add node to nodepool %s", nodepool.ID)
				}
			}
		}
	}
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

// CreateNodePoolOpts defines required fields to create a node pool
// Not using for this project.
type CreateNodePoolOpts struct {
	Name       *string `json:"name,omitempty"`
	FlavorName string  `json:"flavorName"`

	Autoscale     bool `json:"autoscale"`
	MonthlyBilled bool `json:"monthlyBilled"`
	AntiAffinity  bool `json:"antiAffinity"`

	MinNodes *uint32 `json:"minNodes,omitempty"`
	MaxNodes *uint32 `json:"maxNodes,omitempty"`
}

// CreateNodePool allows to creates a node pool in a cluster
// Not using for this project.
func (c *Client) CreateNodePool(ctx context.Context, projectID string, clusterID string, opts *CreateNodePoolOpts) (*NodePool, error) {
	nodepool := &NodePool{}

	return nodepool, c.CallAPIWithContext(
		ctx,
		"POST",
		fmt.Sprintf("/cloud/project/%s/kube/%s/nodepool", projectID, clusterID),
		opts,
		&nodepool,
		nil,
		nil,
		true,
	)
}

// UpdateNodePoolOpts defines required fields to update a node pool
type UpdateNodePoolOpts struct {
	MinNodes *uint32 `json:"minNodes,omitempty"`
	MaxNodes *uint32 `json:"maxNodes,omitempty"`

	Autoscale *bool `json:"autoscale,omitempty"`

	NodesToRemove []string `json:"nodesToRemove,omitempty"`
}

// UpdateNodePool allows to update a specific node pool properties (this call is used for resize)
func (c *Client) UpdateNodePool(ctx context.Context, clusterID string, poolID string, opts *UpdateNodePoolOpts) (*NodePool, error) {
	nodepool := &NodePool{}

	return nodepool, c.CallAPIWithContext(
		ctx,
		"PUT",
		fmt.Sprintf("/cluster/%s/nodegroups/%s", clusterID, poolID),
		opts,
		&nodepool,
		nil,
		nil,
		true,
	)
}

// DeleteNodePool allows to delete a specific node pool
// Not using for this project.
func (c *Client) DeleteNodePool(ctx context.Context, projectID string, clusterID string, poolID string) (*NodePool, error) {
	nodepool := &NodePool{}

	return nodepool, c.CallAPIWithContext(
		ctx,
		"DELETE",
		fmt.Sprintf("/cloud/project/%s/kube/%s/nodepool/%s", projectID, clusterID, poolID),
		nil,
		&nodepool,
		nil,
		nil,
		true,
	)
}
func (c *Client) DeleteNode(ctx context.Context, clusterID, NodeGroupID, Id string) error {
	klog.V(2).Infof("Deleting node %s from cluster %s", Id, clusterID)
	c.CallAPIWithContext(
		ctx,
		"DELETE",
		fmt.Sprintf("/cluster/%s/nodegroups/%s/nodes/%s", clusterID, NodeGroupID, Id),
		nil,
		nil,
		nil,
		nil,
		true,
	)
	return nil
}
func (c *Client) AddNode(ctx context.Context, clusterID, NodeGroupID string) (*Node, error) {
	klog.V(2).Infof("Adding node to cluster %s", clusterID)
	node := &Node{}
	return node, c.CallAPIWithContext(
		ctx,
		"PUT",
		fmt.Sprintf("/cluster/%s/nodegroups/%s/nodes/add", clusterID, NodeGroupID),
		nil,
		node,
		nil,
		nil,
		true,
	)
}
