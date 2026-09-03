/*
Copyright 2022 The Kubernetes Authors.

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

package externalgrpc

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/externalgrpc/protos"
)

type cloudProviderServerMock struct {
	protos.UnimplementedCloudProviderServer

	mock.Mock
}

func (c *cloudProviderServerMock) NodeGroups(ctx context.Context, req *protos.NodeGroupsRequest) (*protos.NodeGroupsResponse, error) {
	args := c.Called(ctx, req)
	return args.Get(0).(*protos.NodeGroupsResponse), args.Error(1)
}

func (c *cloudProviderServerMock) NodeGroupForNode(ctx context.Context, req *protos.NodeGroupForNodeRequest) (*protos.NodeGroupForNodeResponse, error) {
	args := c.Called(ctx, req)
	return args.Get(0).(*protos.NodeGroupForNodeResponse), args.Error(1)
}

func (c *cloudProviderServerMock) PricingNodePrice(ctx context.Context, req *protos.PricingNodePriceRequest) (*protos.PricingNodePriceResponse, error) {
	args := c.Called(ctx, req)
	return args.Get(0).(*protos.PricingNodePriceResponse), args.Error(1)
}

func (c *cloudProviderServerMock) PricingPodPrice(ctx context.Context, req *protos.PricingPodPriceRequest) (*protos.PricingPodPriceResponse, error) {
	args := c.Called(ctx, req)
	return args.Get(0).(*protos.PricingPodPriceResponse), args.Error(1)
}

func (c *cloudProviderServerMock) GPULabel(ctx context.Context, req *protos.GPULabelRequest) (*protos.GPULabelResponse, error) {
	args := c.Called(ctx, req)
	return args.Get(0).(*protos.GPULabelResponse), args.Error(1)
}

func (c *cloudProviderServerMock) GetAvailableGPUTypes(ctx context.Context, req *protos.GetAvailableGPUTypesRequest) (*protos.GetAvailableGPUTypesResponse, error) {
	args := c.Called(ctx, req)
	return args.Get(0).(*protos.GetAvailableGPUTypesResponse), args.Error(1)
}

func (c *cloudProviderServerMock) Cleanup(ctx context.Context, req *protos.CleanupRequest) (*protos.CleanupResponse, error) {
	args := c.Called(ctx, req)
	return args.Get(0).(*protos.CleanupResponse), args.Error(1)
}

func (c *cloudProviderServerMock) Refresh(ctx context.Context, req *protos.RefreshRequest) (*protos.RefreshResponse, error) {
	args := c.Called(ctx, req)
	return args.Get(0).(*protos.RefreshResponse), args.Error(1)
}

func (c *cloudProviderServerMock) NodeGroupTargetSize(ctx context.Context, req *protos.NodeGroupTargetSizeRequest) (*protos.NodeGroupTargetSizeResponse, error) {
	args := c.Called(ctx, req)
	return args.Get(0).(*protos.NodeGroupTargetSizeResponse), args.Error(1)
}

func (c *cloudProviderServerMock) NodeGroupIncreaseSize(ctx context.Context, req *protos.NodeGroupIncreaseSizeRequest) (*protos.NodeGroupIncreaseSizeResponse, error) {
	args := c.Called(ctx, req)
	return args.Get(0).(*protos.NodeGroupIncreaseSizeResponse), args.Error(1)
}

func (c *cloudProviderServerMock) NodeGroupDeleteNodes(ctx context.Context, req *protos.NodeGroupDeleteNodesRequest) (*protos.NodeGroupDeleteNodesResponse, error) {
	args := c.Called(ctx, req)
	return args.Get(0).(*protos.NodeGroupDeleteNodesResponse), args.Error(1)
}

func (c *cloudProviderServerMock) NodeGroupDecreaseTargetSize(ctx context.Context, req *protos.NodeGroupDecreaseTargetSizeRequest) (*protos.NodeGroupDecreaseTargetSizeResponse, error) {
	args := c.Called(ctx, req)
	return args.Get(0).(*protos.NodeGroupDecreaseTargetSizeResponse), args.Error(1)
}

func (c *cloudProviderServerMock) NodeGroupNodes(ctx context.Context, req *protos.NodeGroupNodesRequest) (*protos.NodeGroupNodesResponse, error) {
	args := c.Called(ctx, req)
	return args.Get(0).(*protos.NodeGroupNodesResponse), args.Error(1)
}

func (c *cloudProviderServerMock) NodeGroupTemplateNodeInfo(ctx context.Context, req *protos.NodeGroupTemplateNodeInfoRequest) (*protos.NodeGroupTemplateNodeInfoResponse, error) {
	args := c.Called(ctx, req)
	return args.Get(0).(*protos.NodeGroupTemplateNodeInfoResponse), args.Error(1)
}

func (c *cloudProviderServerMock) NodeGroupGetOptions(ctx context.Context, req *protos.NodeGroupAutoscalingOptionsRequest) (*protos.NodeGroupAutoscalingOptionsResponse, error) {
	args := c.Called(ctx, req)
	return args.Get(0).(*protos.NodeGroupAutoscalingOptionsResponse), args.Error(1)
}

func setupTest(t *testing.T) (protos.CloudProviderClient, *cloudProviderServerMock, func()) {
	t.Helper()
	lis, err := net.Listen("tcp", ":0")
	require.NoError(t, err)

	conn, err := grpc.Dial(lis.Addr().String(), grpc.WithInsecure())
	require.NoError(t, err)

	server := grpc.NewServer()
	m := &cloudProviderServerMock{}
	protos.RegisterCloudProviderServer(server, m)
	require.NoError(t, err)

	go server.Serve(lis)

	client := protos.NewCloudProviderClient(conn)
	return client, m, func() {
		server.Stop()
		conn.Close()
		lis.Close()
	}
}
