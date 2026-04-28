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

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

const (
	defaultDomain = "api.eu.nebius.com"
)

// gRPC fully-qualified method names for the Nebius API.
const (
	mk8sNodeGroupList   = "/nebius.mk8s.v1.NodeGroupService/List"
	mk8sNodeGroupGet    = "/nebius.mk8s.v1.NodeGroupService/Get"
	mk8sNodeGroupUpdate = "/nebius.mk8s.v1.NodeGroupService/Update"
	computeInstanceList  = "/nebius.compute.v1.InstanceService/List"
	computeInstanceDel   = "/nebius.compute.v1.InstanceService/Delete"
)

// Client provides access to the Nebius MK8S and Compute gRPC APIs.
type Client struct {
	mk8sConn    *grpc.ClientConn
	computeConn *grpc.ClientConn
	token       string
	codec       nebiusCodec
}

// NewClient creates a new Nebius API client. The domain defaults to
// "api.eu.nebius.com" if empty. Service-specific gRPC endpoints are
// derived as "<service>.<domain>:443" (e.g., mk8s.api.eu.nebius.com:443).
func NewClient(ctx context.Context, token, domain string) (*Client, error) {
	if domain == "" {
		domain = defaultDomain
	}

	tlsCreds := credentials.NewTLS(&tls.Config{MinVersion: tls.VersionTLS12})

	mk8sAddr := fmt.Sprintf("mk8s.%s:443", domain)
	mk8sConn, err := grpc.NewClient(mk8sAddr, grpc.WithTransportCredentials(tlsCreds))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MK8S API at %s: %w", mk8sAddr, err)
	}

	computeAddr := fmt.Sprintf("compute.%s:443", domain)
	computeConn, err := grpc.NewClient(computeAddr, grpc.WithTransportCredentials(tlsCreds))
	if err != nil {
		mk8sConn.Close()
		return nil, fmt.Errorf("failed to connect to Compute API at %s: %w", computeAddr, err)
	}

	return &Client{
		mk8sConn:    mk8sConn,
		computeConn: computeConn,
		token:       token,
	}, nil
}

// Close releases the underlying gRPC connections.
func (c *Client) Close() error {
	var firstErr error
	if err := c.mk8sConn.Close(); err != nil && firstErr == nil {
		firstErr = err
	}
	if err := c.computeConn.Close(); err != nil && firstErr == nil {
		firstErr = err
	}
	return firstErr
}

// ListNodeGroups lists node groups for the given cluster, handling one page.
func (c *Client) ListNodeGroups(ctx context.Context, req *ListNodeGroupsRequest) (*ListNodeGroupsResponse, error) {
	resp := &ListNodeGroupsResponse{}
	if err := c.invoke(ctx, c.mk8sConn, mk8sNodeGroupList, req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// GetNodeGroup retrieves a single node group by ID.
func (c *Client) GetNodeGroup(ctx context.Context, req *GetNodeGroupRequest) (*NodeGroup, error) {
	resp := &NodeGroup{}
	if err := c.invoke(ctx, c.mk8sConn, mk8sNodeGroupGet, req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// UpdateNodeGroup updates a node group. The Nebius API returns an Operation,
// which we ignore since the autoscaler polls for state changes.
func (c *Client) UpdateNodeGroup(ctx context.Context, req *UpdateNodeGroupRequest) error {
	// The response is a common.v1.Operation. We don't need to parse it;
	// use a throwaway bytes receiver via the codec.
	var ignoreResp rawBytes
	return c.invoke(ctx, c.mk8sConn, mk8sNodeGroupUpdate, req, &ignoreResp)
}

// ListInstances lists compute instances in a parent folder, handling one page.
func (c *Client) ListInstances(ctx context.Context, req *ListInstancesRequest) (*ListInstancesResponse, error) {
	resp := &ListInstancesResponse{}
	if err := c.invoke(ctx, c.computeConn, computeInstanceList, req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// DeleteInstance deletes a compute instance by ID. The response is an
// Operation which we ignore.
func (c *Client) DeleteInstance(ctx context.Context, req *DeleteInstanceRequest) error {
	var ignoreResp rawBytes
	return c.invoke(ctx, c.computeConn, computeInstanceDel, req, &ignoreResp)
}

func (c *Client) invoke(ctx context.Context, conn *grpc.ClientConn, method string, req, resp interface{}) error {
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+c.token)
	return conn.Invoke(ctx, method, req, resp, grpc.ForceCodec(c.codec))
}

// rawBytes is a no-op response type for RPCs whose response we want to ignore
// (e.g., Operation responses from Update/Delete).
type rawBytes struct {
	data []byte
}

// Ensure the codec handles rawBytes as a passthrough.
func init() {
	// Verified at compile time in wire.go codec methods.
}

// Implement io.Closer so Client can be used as Manager.closer.
var _ io.Closer = (*Client)(nil)
