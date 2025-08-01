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
	"github.com/uthoplatforms/utho-go/utho"
)

// Test fixtures
var (
	testConfig = `{"token": "123-456", "cluster_id": "1111"}`

	defaultNodePools = []utho.NodepoolDetails{
		{
			ID:        "pool-1",
			AutoScale: true,
			MinNodes:  1,
			MaxNodes:  3,
			Count:     2,
			Workers: []utho.WorkerNode{
				{ID: 123, Status: "Active"},
			},
		},
		{
			ID:        "pool-2",
			AutoScale: true,
			MinNodes:  2,
			MaxNodes:  5,
			Count:     3,
			Workers: []utho.WorkerNode{
				{ID: 456, Status: "Active"},
			},
		},
		{
			ID:        "pool-3",
			AutoScale: false, // Non-autoscaling pool
			MinNodes:  1,
			MaxNodes:  4,
			Count:     2,
		},
	}
)

// Helper function to create a test manager
func createTestManager(t *testing.T) *Manager {
	manager, err := newManager(strings.NewReader(testConfig))
	require.NoError(t, err)
	return manager
}

// Helper function to setup mock client with node pools
func setupMockClient(t *testing.T, manager *Manager, nodePools []utho.NodepoolDetails) *uthoClientMock {
	client := &uthoClientMock{}
	ctx := context.Background()

	client.On("ListNodePools", ctx, manager.clusterID, nil).Return(
		nodePools,
		&utho.Meta{},
		nil,
	).Once()

	manager.client = client
	return client
}

func TestManager_newManager(t *testing.T) {
	t.Run("valid configuration", func(t *testing.T) {
		manager := createTestManager(t)
		assert.Equal(t, "1111", manager.clusterID, "invalid cluster id")
	})

	t.Run("missing token", func(t *testing.T) {
		config := `{"token": "", "cluster_id": "1111"}`
		_, err := newManager(strings.NewReader(config))
		assert.EqualError(t, err, "access token is not provided")
	})

	t.Run("missing cluster id", func(t *testing.T) {
		config := `{"token": "123-345", "cluster_id": ""}`
		_, err := newManager(strings.NewReader(config))
		assert.EqualError(t, err, "cluster ID is not provided and couldn't be retrieved from nodes")
	})
}

func TestManager_Refresh(t *testing.T) {
	t.Run("successful refresh with autoscaling pools", func(t *testing.T) {
		manager := createTestManager(t)
		client := setupMockClient(t, manager, defaultNodePools)

		err := manager.Refresh()
		assert.NoError(t, err)
		assert.Equal(t, 2, len(manager.nodeGroups), "should only include autoscaling pools")

		// Verify first node group
		assert.Equal(t, 1, manager.nodeGroups[0].minSize)
		assert.Equal(t, 3, manager.nodeGroups[0].MaxSize())

		// Verify second node group
		assert.Equal(t, 2, manager.nodeGroups[1].minSize)
		assert.Equal(t, 5, manager.nodeGroups[1].maxSize)

		client.AssertExpectations(t)
	})

	t.Run("refresh with API error", func(t *testing.T) {
		manager := createTestManager(t)
		client := &uthoClientMock{}
		ctx := context.Background()

		expectedError := errors.New("API error")
		client.On("ListNodePools", ctx, manager.clusterID, nil).Return(
			[]utho.NodepoolDetails{},
			&utho.Meta{},
			expectedError,
		).Once()

		manager.client = client
		err := manager.Refresh()
		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		client.AssertExpectations(t)
	})
}

func TestManager_ListNodePools(t *testing.T) {
	t.Run("successfully list multiple node pools", func(t *testing.T) {
		manager := createTestManager(t)
		client := setupMockClient(t, manager, defaultNodePools)

		err := manager.Refresh()
		assert.NoError(t, err)

		assert.Equal(t, 2, len(manager.nodeGroups))
		assert.Equal(t, "pool-1", manager.nodeGroups[0].id)
		assert.Equal(t, "pool-2", manager.nodeGroups[1].id)
		client.AssertExpectations(t)
	})

	t.Run("list pools when API returns error", func(t *testing.T) {
		manager := createTestManager(t)
		client := &uthoClientMock{}
		ctx := context.Background()

		expectedError := errors.New("API error")
		client.On("ListNodePools", ctx, manager.clusterID, nil).Return(
			[]utho.NodepoolDetails{},
			&utho.Meta{},
			expectedError,
		).Once()

		manager.client = client
		err := manager.Refresh()
		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		client.AssertExpectations(t)
	})

	t.Run("filter out non-autoscaling pools", func(t *testing.T) {
		manager := createTestManager(t)
		client := setupMockClient(t, manager, defaultNodePools)

		err := manager.Refresh()
		assert.NoError(t, err)

		assert.Equal(t, 2, len(manager.nodeGroups))
		assert.Equal(t, "pool-1", manager.nodeGroups[0].id)
		assert.Equal(t, "pool-2", manager.nodeGroups[1].id)
		client.AssertExpectations(t)
	})

	t.Run("handle empty node pools list", func(t *testing.T) {
		manager := createTestManager(t)
		client := setupMockClient(t, manager, []utho.NodepoolDetails{})

		err := manager.Refresh()
		assert.NoError(t, err)
		assert.Equal(t, 0, len(manager.nodeGroups))
		client.AssertExpectations(t)
	})
}

func TestNodeGroup_Nodes_EmptyWorkers(t *testing.T) {
	client := &uthoClientMock{}
	nodeGroup := testNodeGroup(client, &utho.NodepoolDetails{
		Workers: []utho.WorkerNode{},
	})

	nodes, err := nodeGroup.Nodes()
	assert.NoError(t, err)
	assert.Empty(t, nodes, "nodes should be empty when no workers exist")
}

func TestNodeGroup_Nodes_NilNodePool(t *testing.T) {
	client := &uthoClientMock{}
	nodeGroup := &NodeGroup{
		id:        "pool-123",
		clusterID: 1111,
		client:    client,
		nodePool:  nil,
	}

	nodes, err := nodeGroup.Nodes()
	assert.Error(t, err)
	assert.Nil(t, nodes)
	assert.Contains(t, err.Error(), "node pool instance is not created")
}
