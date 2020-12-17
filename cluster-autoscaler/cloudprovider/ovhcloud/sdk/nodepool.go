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
	"time"
)

// NodePool defines the nodes group deployed on OVHcloud
type NodePool struct {
	ID        string `json:"id"`
	ProjectID string `json:"projectId"`

	Name       string `json:"name"`
	Flavor     string `json:"flavor"`
	Status     string `json:"status"`
	SizeStatus string `json:"sizeStatus"`

	Autoscale     bool `json:"autoscale"`
	MonthlyBilled bool `json:"monthlyBilled"`
	AntiAffinity  bool `json:"antiAffinity"`

	DesiredNodes   uint32 `json:"desiredNodes"`
	MinNodes       uint32 `json:"minNodes"`
	MaxNodes       uint32 `json:"maxNodes"`
	CurrentNodes   uint32 `json:"currentNodes"`
	AvailableNodes uint32 `json:"availableNodes"`
	UpToDateNodes  uint32 `json:"upToDateNodes"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// ListNodePools allows to list all node pools available in a cluster
func (c *Client) ListNodePools(ctx context.Context, projectID, clusterID string) ([]NodePool, error) {
	nodepools := make([]NodePool, 0)

	return nodepools, c.CallAPIWithContext(
		ctx,
		"GET",
		fmt.Sprintf("/cloud/project/%s/kube/%s/nodepool", projectID, clusterID),
		nil,
		&nodepools,
		nil,
		true,
	)
}

// GetNodePool allows to display information for a specific node pool
func (c *Client) GetNodePool(ctx context.Context, projectID string, clusterID string, poolID string) (*NodePool, error) {
	nodepool := &NodePool{}

	return nodepool, c.CallAPIWithContext(
		ctx,
		"GET",
		fmt.Sprintf("/cloud/project/%s/kube/%s/nodepool/%s", projectID, clusterID, poolID),
		nil,
		&nodepool,
		nil,
		true,
	)
}

// ListNodePoolNodes allows to display nodes contained in a parent node pool
func (c *Client) ListNodePoolNodes(ctx context.Context, projectID string, clusterID string, poolID string) ([]Node, error) {
	nodes := make([]Node, 0)

	return nodes, c.CallAPIWithContext(
		ctx,
		"GET",
		fmt.Sprintf("/cloud/project/%s/kube/%s/nodepool/%s/nodes", projectID, clusterID, poolID),
		nil,
		&nodes,
		nil,
		true,
	)
}

// CreateNodePoolOpts defines required fields to create a node pool
type CreateNodePoolOpts struct {
	Name       *string `json:"name,omitempty"`
	FlavorName string  `json:"flavorName"`

	Autoscale     bool `json:"autoscale"`
	MonthlyBilled bool `json:"monthlyBilled"`
	AntiAffinity  bool `json:"antiAffinity"`

	DesiredNodes *uint32 `json:"desiredNodes,omitempty"`
	MinNodes     *uint32 `json:"minNodes,omitempty"`
	MaxNodes     *uint32 `json:"maxNodes,omitempty"`
}

// CreateNodePool allows to creates a node pool in a cluster
func (c *Client) CreateNodePool(ctx context.Context, projectID string, clusterID string, opts *CreateNodePoolOpts) (*NodePool, error) {
	nodepool := &NodePool{}

	return nodepool, c.CallAPIWithContext(
		ctx,
		"POST",
		fmt.Sprintf("/cloud/project/%s/kube/%s/nodepool", projectID, clusterID),
		opts,
		&nodepool,
		nil,
		true,
	)
}

// UpdateNodePoolOpts defines required fields to update a node pool
type UpdateNodePoolOpts struct {
	DesiredNodes *uint32 `json:"desiredNodes,omitempty"`
	MinNodes     *uint32 `json:"minNodes,omitempty"`
	MaxNodes     *uint32 `json:"maxNodes,omitempty"`

	Autoscale *bool `json:"autoscale"`

	NodesToRemove []string `json:"nodesToRemove,omitempty"`
}

// UpdateNodePool allows to update a specific node pool properties (this call is used for resize)
func (c *Client) UpdateNodePool(ctx context.Context, projectID string, clusterID string, poolID string, opts *UpdateNodePoolOpts) (*NodePool, error) {
	nodepool := &NodePool{}

	return nodepool, c.CallAPIWithContext(
		ctx,
		"PUT",
		fmt.Sprintf("/cloud/project/%s/kube/%s/nodepool/%s", projectID, clusterID, poolID),
		opts,
		&nodepool,
		nil,
		true,
	)
}

// DeleteNodePool allows to delete a specific node pool
func (c *Client) DeleteNodePool(ctx context.Context, projectID string, clusterID string, poolID string) (*NodePool, error) {
	nodepool := &NodePool{}

	return nodepool, c.CallAPIWithContext(
		ctx,
		"DELETE",
		fmt.Sprintf("/cloud/project/%s/kube/%s/nodepool/%s", projectID, clusterID, poolID),
		nil,
		&nodepool,
		nil,
		true,
	)
}
