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

package scaleway

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/scaleway/scalewaygo"
	"testing"
)

func TestNodeGroup_TargetSize(t *testing.T) {
	var nodesNb uint32 = 3

	ng := &NodeGroup{
		p: &scalewaygo.Pool{
			Size: nodesNb,
		},
	}
	size, err := ng.TargetSize()
	assert.NoError(t, err)
	assert.Equal(t, int(nodesNb), size, "target size is wrong")
}

func TestNodeGroup_IncreaseSize(t *testing.T) {
	ctx := context.Background()
	nodesNb := 3
	delta := 2
	client := &clientMock{}
	ng := &NodeGroup{
		Client: client,
		p: &scalewaygo.Pool{
			Size:        uint32(nodesNb),
			MinSize:     1,
			MaxSize:     10,
			Autoscaling: true,
		},
	}

	newSize := uint32(nodesNb + delta)
	client.On("UpdatePool",
		ctx,
		&scalewaygo.UpdatePoolRequest{
			PoolID: ng.p.ID,
			Size:   &newSize,
		}).Return(
		&scalewaygo.Pool{
			Size: newSize,
		}, nil,
	).Once()
	err := ng.IncreaseSize(delta)
	assert.NoError(t, err)
}

func TestNodeGroup_IncreaseNegativeDelta(t *testing.T) {
	nodesNb := 3
	delta := -2
	client := &clientMock{}
	ng := &NodeGroup{
		Client: client,
		p: &scalewaygo.Pool{
			Size: uint32(nodesNb),
		},
	}

	err := ng.IncreaseSize(delta)
	assert.Error(t, err)
}

func TestNodeGroup_IncreaseAboveMaximum(t *testing.T) {
	nodesNb := 3
	delta := 10
	client := &clientMock{}
	ng := &NodeGroup{
		Client: client,
		p: &scalewaygo.Pool{
			Size:    uint32(nodesNb),
			MaxSize: 10,
		},
	}

	err := ng.IncreaseSize(delta)
	assert.Error(t, err)
}

func TestNodeGroup_DecreaseTargetSize(t *testing.T) {
	ctx := context.Background()
	nodesNb := 5
	delta := -4
	client := &clientMock{}
	ng := &NodeGroup{
		Client: client,
		p: &scalewaygo.Pool{
			Size:        uint32(nodesNb),
			MinSize:     1,
			MaxSize:     10,
			Autoscaling: true,
		},
	}

	newSize := uint32(nodesNb + delta)
	client.On("UpdatePool",
		ctx,
		&scalewaygo.UpdatePoolRequest{
			PoolID: ng.p.ID,
			Size:   &newSize,
		}).Return(
		&scalewaygo.Pool{
			Size: newSize,
		}, nil,
	).Once()
	err := ng.DecreaseTargetSize(delta)
	assert.NoError(t, err)
}

func TestNodeGroup_DecreaseTargetSizePositiveDelta(t *testing.T) {
	nodesNb := 3
	delta := 2
	client := &clientMock{}
	ng := &NodeGroup{
		Client: client,
		p: &scalewaygo.Pool{
			Size: uint32(nodesNb),
		},
	}

	err := ng.DecreaseTargetSize(delta)
	assert.Error(t, err)
}

func TestNodeGroup_DecreaseBelowMinimum(t *testing.T) {
	nodesNb := 3
	delta := -3
	client := &clientMock{}
	ng := &NodeGroup{
		Client: client,
		p: &scalewaygo.Pool{
			Size:    uint32(nodesNb),
			MinSize: 1,
		},
	}

	err := ng.DecreaseTargetSize(delta)
	assert.Error(t, err)
}

func TestNodeGroup_DeleteNodes(t *testing.T) {
	ctx := context.Background()
	client := &clientMock{}
	ng := &NodeGroup{
		Client: client,
		nodes: map[string]*scalewaygo.Node{
			"scaleway://instance/fr-par-1/f80ce5b1-7c77-4177-bd5f-0d803f5b7c15": {ID: "6852824b-e409-4c77-94df-819629d135b9", ProviderID: "scaleway://instance/fr-par-1/f80ce5b1-7c77-4177-bd5f-0d803f5b7c15"},
			"scaleway://instance/fr-srr-1/6c22c989-ddce-41d8-98cb-2aea83c72066": {ID: "84acb1a6-0e14-4j36-8b32-71bf7b328c22", ProviderID: "scaleway://instance/fr-srr-1/6c22c989-ddce-41d8-98cb-2aea83c72066"},
			"scaleway://instance/fr-srr-1/fcc3abe0-3a72-4178-8182-2a93fdc72529": {ID: "5c4d832a-d964-4c64-9d53-b9295c206cdd", ProviderID: "scaleway://instance/fr-srr-1/fcc3abe0-3a72-4178-8182-2a93fdc72529"},
		},
		p: &scalewaygo.Pool{
			Size: 3,
		},
	}

	nodes := []*apiv1.Node{
		{Spec: apiv1.NodeSpec{ProviderID: "scaleway://instance/fr-par-1/f80ce5b1-7c77-4177-bd5f-0d803f5b7c15"}},
		{Spec: apiv1.NodeSpec{ProviderID: "scaleway://instance/fr-srr-1/6c22c989-ddce-41d8-98cb-2aea83c72066"}},
		{Spec: apiv1.NodeSpec{ProviderID: "scaleway://instance/fr-srr-1/fcc3abe0-3a72-4178-8182-2a93fdc72529"}},
	}
	client.On("DeleteNode", ctx, &scalewaygo.DeleteNodeRequest{NodeID: ng.nodes["scaleway://instance/fr-par-1/f80ce5b1-7c77-4177-bd5f-0d803f5b7c15"].ID}).Return(&scalewaygo.Node{Status: scalewaygo.NodeStatusDeleting}, nil).Once()
	client.On("DeleteNode", ctx, &scalewaygo.DeleteNodeRequest{NodeID: ng.nodes["scaleway://instance/fr-srr-1/6c22c989-ddce-41d8-98cb-2aea83c72066"].ID}).Return(&scalewaygo.Node{Status: scalewaygo.NodeStatusDeleting}, nil).Once()
	client.On("DeleteNode", ctx, &scalewaygo.DeleteNodeRequest{NodeID: ng.nodes["scaleway://instance/fr-srr-1/fcc3abe0-3a72-4178-8182-2a93fdc72529"].ID}).Return(&scalewaygo.Node{Status: scalewaygo.NodeStatusDeleting}, nil).Once()

	err := ng.DeleteNodes(nodes)
	assert.NoError(t, err)
	assert.Equal(t, uint32(0), ng.p.Size)
}

func TestNodeGroup_DeleteNodesErr(t *testing.T) {
	ctx := context.Background()
	client := &clientMock{}
	ng := &NodeGroup{
		Client: client,
		nodes: map[string]*scalewaygo.Node{
			"nonexistent-on-provider-side": {ID: "unknown"},
		},
	}
	nodes := []*apiv1.Node{
		{Spec: apiv1.NodeSpec{ProviderID: "nonexistent-on-provider-side"}},
	}
	client.On("DeleteNode", ctx, &scalewaygo.DeleteNodeRequest{NodeID: "unknown"}).Return(&scalewaygo.Node{}, errors.New("nonexistent")).Once()

	err := ng.DeleteNodes(nodes)
	assert.Error(t, err)
}

type clientMock struct {
	mock.Mock
}

func (m *clientMock) GetPool(ctx context.Context, req *scalewaygo.GetPoolRequest) (*scalewaygo.Pool, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*scalewaygo.Pool), args.Error(1)
}

func (m *clientMock) ListPools(ctx context.Context, req *scalewaygo.ListPoolsRequest) (*scalewaygo.ListPoolsResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*scalewaygo.ListPoolsResponse), args.Error(1)
}

func (m *clientMock) UpdatePool(ctx context.Context, req *scalewaygo.UpdatePoolRequest) (*scalewaygo.Pool, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*scalewaygo.Pool), args.Error(1)
}

func (m *clientMock) ListNodes(ctx context.Context, req *scalewaygo.ListNodesRequest) (*scalewaygo.ListNodesResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*scalewaygo.ListNodesResponse), args.Error(1)
}

func (m *clientMock) DeleteNode(ctx context.Context, req *scalewaygo.DeleteNodeRequest) (*scalewaygo.Node, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*scalewaygo.Node), args.Error(1)
}
