/*
Copyright 2025 The Kubernetes Authors.

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

package utho

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/utho/utho-go"
)

// Helper function to set up mock ListNodePools
func setupMockListNodePools(ctx context.Context, client *uthoClientMock, clusterID string, nodePools []utho.NodepoolDetails, err error) {
	client.On("ListNodePools", ctx, clusterID, nil).Return(nodePools, &utho.Meta{}, err).Once()
}

func TestCloudProvider_NewCloudProvider_Success(t *testing.T) {
	config := `{"token": "123-456", "cluster_id": "111"}`

	manager, err := newManager(strings.NewReader(config))
	require.NoError(t, err)

	client := &uthoClientMock{}
	ctx := context.Background()

	setupMockListNodePools(ctx, client, "123", []utho.NodepoolDetails{
		{ID: "1234", AutoScale: true, MinNodes: 1, MaxNodes: 2},
		{ID: "4567", AutoScale: true, MinNodes: 5, MaxNodes: 8},
		{ID: "9876", AutoScale: false, MinNodes: 5, MaxNodes: 8},
	}, nil)

	manager.client = client

	_ = newUthoCloudProvider(manager, &cloudprovider.ResourceLimiter{})
}

func TestUthoCloudProvider_NewNodeGroup_Success(t *testing.T) {
	config := `{"token": "123-456", "cluster_id": "111"}`

	manager, err := newManager(strings.NewReader(config))
	require.NoError(t, err)

	client := &uthoClientMock{}
	ctx := context.Background()

	setupMockListNodePools(ctx, client, manager.clusterID, []utho.NodepoolDetails{
		{ID: "1234", AutoScale: true, MinNodes: 1, MaxNodes: 2},
		{ID: "4567", AutoScale: true, MinNodes: 5, MaxNodes: 8},
		{ID: "9876", AutoScale: false, MinNodes: 5, MaxNodes: 8},
	}, nil)

	manager.client = client
	provider := newUthoCloudProvider(manager, &cloudprovider.ResourceLimiter{})

	err = provider.Refresh()
	assert.NoError(t, err)

	nodes := provider.NodeGroups()
	assert.Equal(t, 2, len(nodes), "number of nodes do not match")
}

func TestUthoCloudProvider_NodeGroupForNode_Success(t *testing.T) {
	config := `{"token": "123-456", "cluster_id": "111"}`

	manager, err := newManager(strings.NewReader(config))
	require.NoError(t, err)

	client := &uthoClientMock{}
	ctx := context.Background()

	setupMockListNodePools(ctx, client, manager.clusterID, []utho.NodepoolDetails{
		{
			ID:        "pool-123",
			AutoScale: true,
			Workers: []utho.WorkerNode{
				{
					ID:     1234,
					Status: "Active",
				},
			},
			MinNodes: 1,
			MaxNodes: 2,
		},
		{
			ID:        "pool-456",
			AutoScale: true,
			Workers: []utho.WorkerNode{
				{
					ID:     456,
					Status: "Active",
				},
			},
			MinNodes: 5,
			MaxNodes: 8,
		},
	}, nil)

	manager.client = client
	provider := newUthoCloudProvider(manager, &cloudprovider.ResourceLimiter{})

	err = provider.Refresh()
	assert.NoError(t, err)

	node := &apiv1.Node{Spec: apiv1.NodeSpec{ProviderID: toProviderID("1234")}}
	nodeGroup, err := provider.NodeGroupForNode(node)
	require.NoError(t, err)

	require.NotNil(t, nodeGroup)
	require.Equal(t, "pool-123", nodeGroup.Id(), "nodegroup IDs do not match")

}

func TestUthoCloudProvider_Name_ReturnsCorrectName(t *testing.T) {
	config := `{"token": "123-456", "cluster_id": "111"}`

	manager, err := newManager(strings.NewReader(config))
	require.NoError(t, err)

	p := newUthoCloudProvider(manager, &cloudprovider.ResourceLimiter{})
	assert.Equal(t, cloudprovider.UthoProviderName, p.Name(), "provider name doesn't match")
}

func TestUthoCloudProvider_Refresh_EmptyNodePools(t *testing.T) {
	config := `{"token": "123-456", "cluster_id": "111"}`

	manager, err := newManager(strings.NewReader(config))
	require.NoError(t, err)

	client := &uthoClientMock{}
	ctx := context.Background()

	setupMockListNodePools(ctx, client, manager.clusterID, []utho.NodepoolDetails{}, nil)

	manager.client = client
	provider := newUthoCloudProvider(manager, &cloudprovider.ResourceLimiter{})

	err = provider.Refresh()
	assert.NoError(t, err)

	nodes := provider.NodeGroups()
	assert.Equal(t, 0, len(nodes), "expected no node groups")

}

func TestUthoCloudProvider_Refresh_ErrorFromAPI(t *testing.T) {
	config := `{"token": "123-456", "cluster_id": "111"}`

	manager, err := newManager(strings.NewReader(config))
	require.NoError(t, err)

	client := &uthoClientMock{}
	ctx := context.Background()

	setupMockListNodePools(ctx, client, manager.clusterID, nil, errors.New("mock error"))

	manager.client = client
	provider := newUthoCloudProvider(manager, &cloudprovider.ResourceLimiter{})

	err = provider.Refresh()
	assert.Error(t, err, "expected an error from ListNodePools")

}
