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
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/scaleway/scalewaygo"
)

// mockClient is a mock implementation of scalewaygo.Client
type mockClient struct {
	mock.Mock
}

func (m *mockClient) ListPools(ctx context.Context, clusterID string) (time.Duration, []scalewaygo.Pool, error) {
	args := m.Called(ctx, clusterID)
	return args.Get(0).(time.Duration), args.Get(1).([]scalewaygo.Pool), args.Error(2)
}

func (m *mockClient) UpdatePool(ctx context.Context, poolID string, size int) (scalewaygo.Pool, error) {
	args := m.Called(ctx, poolID, size)
	return args.Get(0).(scalewaygo.Pool), args.Error(1)
}

func (m *mockClient) ListNodes(ctx context.Context, clusterID string) (time.Duration, []scalewaygo.Node, error) {
	args := m.Called(ctx, clusterID)
	return args.Get(0).(time.Duration), args.Get(1).([]scalewaygo.Node), args.Error(2)
}

func (m *mockClient) DeleteNode(ctx context.Context, nodeID string) (scalewaygo.Node, error) {
	args := m.Called(ctx, nodeID)
	return args.Get(0).(scalewaygo.Node), args.Error(1)
}

func createTestPool(id string, autoscaling bool, size, minSize, maxSize int) scalewaygo.Pool {
	now := time.Now()
	return scalewaygo.Pool{
		ID:          id,
		ClusterID:   "test-cluster",
		Name:        fmt.Sprintf("pool-%s", id),
		Status:      scalewaygo.PoolStatusReady,
		Version:     "1.27.0",
		NodeType:    "DEV1-M",
		Autoscaling: autoscaling,
		Size:        size,
		MinSize:     minSize,
		MaxSize:     maxSize,
		Zone:        "fr-par-1",
		Capacity: map[string]int64{
			"cpu":               2000,
			"memory":            4096000000,
			"ephemeral-storage": 20000000000,
			"pods":              110,
		},
		Allocatable: map[string]int64{
			"cpu":               1800,
			"memory":            3500000000,
			"ephemeral-storage": 18000000000,
			"pods":              110,
		},
		Labels: map[string]string{
			"kubernetes.io/hostname": "node-1",
		},
		Taints:           map[string]string{},
		NodePricePerHour: 0.05,
		CreatedAt:        &now,
		UpdatedAt:        &now,
	}
}

func createTestNode(nodeID, poolID, providerID string, status scalewaygo.NodeStatus) scalewaygo.Node {
	now := time.Now()
	return scalewaygo.Node{
		ID:         nodeID,
		PoolID:     poolID,
		ClusterID:  "test-cluster",
		ProviderID: providerID,
		Name:       fmt.Sprintf("node-%s", nodeID),
		Status:     status,
		CreatedAt:  &now,
		UpdatedAt:  &now,
	}
}

func TestScalewayCloudProvider_Name(t *testing.T) {
	client := new(mockClient)
	provider := &scalewayCloudProvider{
		client:    client,
		clusterID: "test-cluster",
	}

	assert.Equal(t, cloudprovider.ScalewayProviderName, provider.Name())
}

func TestScalewayCloudProvider_Refresh(t *testing.T) {
	t.Run("successful refresh", func(t *testing.T) {
		client := new(mockClient)
		pool1 := createTestPool("pool-1", true, 3, 1, 10)
		pool2 := createTestPool("pool-2", true, 2, 1, 5)
		pool3 := createTestPool("pool-3", false, 1, 1, 1) // autoscaling disabled

		node1 := createTestNode("node-1", "pool-1", "scaleway://fr-par-1/instance-1", scalewaygo.NodeStatusReady)
		node2 := createTestNode("node-2", "pool-1", "scaleway://fr-par-1/instance-2", scalewaygo.NodeStatusReady)
		node3 := createTestNode("node-3", "pool-2", "scaleway://fr-par-1/instance-3", scalewaygo.NodeStatusReady)

		client.On("ListPools", mock.Anything, "test-cluster").Return(
			30*time.Second,
			[]scalewaygo.Pool{pool1, pool2, pool3},
			nil,
		)
		client.On("ListNodes", mock.Anything, "test-cluster").Return(
			30*time.Second,
			[]scalewaygo.Node{node1, node2, node3},
			nil,
		)

		provider := &scalewayCloudProvider{
			client:          client,
			clusterID:       "test-cluster",
			refreshInterval: DefaultRefreshInterval,
			nodeGroups:      make(map[string]*NodeGroup),
		}

		err := provider.Refresh()
		require.NoError(t, err)

		// Verify node groups (only autoscaling pools)
		assert.Len(t, provider.nodeGroups, 2)
		assert.Contains(t, provider.nodeGroups, "pool-1")
		assert.Contains(t, provider.nodeGroups, "pool-2")
		assert.NotContains(t, provider.nodeGroups, "pool-3")

		// Verify nodes are assigned correctly
		assert.Len(t, provider.nodeGroups["pool-1"].nodes, 2)
		assert.Len(t, provider.nodeGroups["pool-2"].nodes, 1)

		// Verify refresh interval is updated
		assert.Equal(t, 30*time.Second, provider.refreshInterval)

		client.AssertExpectations(t)
	})

	t.Run("caching prevents refresh", func(t *testing.T) {
		client := new(mockClient)
		pool := createTestPool("pool-1", true, 3, 1, 10)
		node := createTestNode("node-1", "pool-1", "scaleway://fr-par-1/instance-1", scalewaygo.NodeStatusReady)

		// First refresh
		client.On("ListPools", mock.Anything, "test-cluster").Return(
			5*time.Second,
			[]scalewaygo.Pool{pool},
			nil,
		).Once()
		client.On("ListNodes", mock.Anything, "test-cluster").Return(
			5*time.Second,
			[]scalewaygo.Node{node},
			nil,
		).Once()

		provider := &scalewayCloudProvider{
			client:          client,
			clusterID:       "test-cluster",
			refreshInterval: 5 * time.Second,
			nodeGroups:      make(map[string]*NodeGroup),
		}

		err := provider.Refresh()
		require.NoError(t, err)
		assert.NotZero(t, provider.lastRefresh)

		// Second refresh immediately - should be skipped
		err = provider.Refresh()
		require.NoError(t, err)

		// Only one call to each method
		client.AssertExpectations(t)
	})

	t.Run("refresh after interval elapsed", func(t *testing.T) {
		client := new(mockClient)
		pool := createTestPool("pool-1", true, 3, 1, 10)
		node := createTestNode("node-1", "pool-1", "scaleway://fr-par-1/instance-1", scalewaygo.NodeStatusReady)

		// First refresh
		client.On("ListPools", mock.Anything, "test-cluster").Return(
			1*time.Millisecond,
			[]scalewaygo.Pool{pool},
			nil,
		).Twice()
		client.On("ListNodes", mock.Anything, "test-cluster").Return(
			1*time.Millisecond,
			[]scalewaygo.Node{node},
			nil,
		).Twice()

		provider := &scalewayCloudProvider{
			client:          client,
			clusterID:       "test-cluster",
			refreshInterval: 1 * time.Millisecond,
			nodeGroups:      make(map[string]*NodeGroup),
		}

		err := provider.Refresh()
		require.NoError(t, err)

		// Wait for interval to elapse
		time.Sleep(2 * time.Millisecond)

		// Second refresh should execute
		err = provider.Refresh()
		require.NoError(t, err)

		client.AssertExpectations(t)
	})

	t.Run("error on ListPools", func(t *testing.T) {
		client := new(mockClient)

		client.On("ListPools", mock.Anything, "test-cluster").Return(
			time.Duration(0),
			[]scalewaygo.Pool{},
			fmt.Errorf("API error"),
		)

		provider := &scalewayCloudProvider{
			client:          client,
			clusterID:       "test-cluster",
			refreshInterval: DefaultRefreshInterval,
			nodeGroups:      make(map[string]*NodeGroup),
		}

		err := provider.Refresh()
		assert.Error(t, err)
		assert.Equal(t, err, provider.lastRefreshError)

		client.AssertExpectations(t)
	})

	t.Run("error on ListNodes", func(t *testing.T) {
		client := new(mockClient)
		pool := createTestPool("pool-1", true, 3, 1, 10)

		client.On("ListPools", mock.Anything, "test-cluster").Return(
			30*time.Second,
			[]scalewaygo.Pool{pool},
			nil,
		)
		client.On("ListNodes", mock.Anything, "test-cluster").Return(
			time.Duration(0),
			[]scalewaygo.Node{},
			fmt.Errorf("API error"),
		)

		provider := &scalewayCloudProvider{
			client:          client,
			clusterID:       "test-cluster",
			refreshInterval: DefaultRefreshInterval,
			nodeGroups:      make(map[string]*NodeGroup),
		}

		err := provider.Refresh()
		assert.Error(t, err)
		assert.Equal(t, err, provider.lastRefreshError)

		client.AssertExpectations(t)
	})

	t.Run("nodes for non-existent pool are skipped", func(t *testing.T) {
		client := new(mockClient)
		pool1 := createTestPool("pool-1", true, 1, 1, 10)
		node1 := createTestNode("node-1", "pool-1", "scaleway://fr-par-1/instance-1", scalewaygo.NodeStatusReady)
		node2 := createTestNode("node-2", "pool-nonexistent", "scaleway://fr-par-1/instance-2", scalewaygo.NodeStatusReady)

		client.On("ListPools", mock.Anything, "test-cluster").Return(
			30*time.Second,
			[]scalewaygo.Pool{pool1},
			nil,
		)
		client.On("ListNodes", mock.Anything, "test-cluster").Return(
			30*time.Second,
			[]scalewaygo.Node{node1, node2},
			nil,
		)

		provider := &scalewayCloudProvider{
			client:          client,
			clusterID:       "test-cluster",
			refreshInterval: DefaultRefreshInterval,
			nodeGroups:      make(map[string]*NodeGroup),
		}

		err := provider.Refresh()
		require.NoError(t, err)

		// Only pool-1 should have nodes
		assert.Len(t, provider.nodeGroups["pool-1"].nodes, 1)

		client.AssertExpectations(t)
	})
}

func TestScalewayCloudProvider_NodeGroups(t *testing.T) {
	client := new(mockClient)
	pool1 := createTestPool("pool-1", true, 3, 1, 10)
	pool2 := createTestPool("pool-2", true, 2, 1, 5)

	provider := &scalewayCloudProvider{
		client:    client,
		clusterID: "test-cluster",
		nodeGroups: map[string]*NodeGroup{
			"pool-1": {
				Client: client,
				pool:   pool1,
				nodes:  make(map[string]*scalewaygo.Node),
			},
			"pool-2": {
				Client: client,
				pool:   pool2,
				nodes:  make(map[string]*scalewaygo.Node),
			},
		},
	}

	// Pre-convert for testing
	provider.providerNodeGroups = []cloudprovider.NodeGroup{
		provider.nodeGroups["pool-1"],
		provider.nodeGroups["pool-2"],
	}

	nodeGroups := provider.NodeGroups()
	assert.Len(t, nodeGroups, 2)
}

func TestScalewayCloudProvider_NodeGroupForNode(t *testing.T) {
	t.Run("node found in pool", func(t *testing.T) {
		client := new(mockClient)
		pool := createTestPool("pool-1", true, 3, 1, 10)
		node := createTestNode("node-1", "pool-1", "scaleway://fr-par-1/instance-1", scalewaygo.NodeStatusReady)

		provider := &scalewayCloudProvider{
			client:    client,
			clusterID: "test-cluster",
			nodeGroups: map[string]*NodeGroup{
				"pool-1": {
					Client: client,
					pool:   pool,
					nodes: map[string]*scalewaygo.Node{
						"scaleway://fr-par-1/instance-1": &node,
					},
				},
			},
		}

		k8sNode := &apiv1.Node{
			Spec: apiv1.NodeSpec{
				ProviderID: "scaleway://fr-par-1/instance-1",
			},
		}

		ng, err := provider.NodeGroupForNode(k8sNode)
		require.NoError(t, err)
		require.NotNil(t, ng)
		assert.Equal(t, "pool-1", ng.Id())
	})

	t.Run("node not found", func(t *testing.T) {
		client := new(mockClient)
		pool := createTestPool("pool-1", true, 3, 1, 10)

		provider := &scalewayCloudProvider{
			client:    client,
			clusterID: "test-cluster",
			nodeGroups: map[string]*NodeGroup{
				"pool-1": {
					Client: client,
					pool:   pool,
					nodes:  make(map[string]*scalewaygo.Node),
				},
			},
		}

		k8sNode := &apiv1.Node{
			Spec: apiv1.NodeSpec{
				ProviderID: "scaleway://fr-par-1/instance-nonexistent",
			},
		}

		ng, err := provider.NodeGroupForNode(k8sNode)
		require.NoError(t, err)
		assert.Nil(t, ng)
	})
}

func TestScalewayCloudProvider_HasInstance(t *testing.T) {
	provider := &scalewayCloudProvider{}

	t.Run("node with provider ID", func(t *testing.T) {
		node := &apiv1.Node{
			Spec: apiv1.NodeSpec{
				ProviderID: "scaleway://fr-par-1/instance-1",
			},
		}

		hasInstance, err := provider.HasInstance(node)
		require.NoError(t, err)
		assert.True(t, hasInstance)
	})

	t.Run("node without provider ID", func(t *testing.T) {
		node := &apiv1.Node{
			Spec: apiv1.NodeSpec{
				ProviderID: "",
			},
		}

		hasInstance, err := provider.HasInstance(node)
		require.NoError(t, err)
		assert.False(t, hasInstance)
	})
}

func TestScalewayCloudProvider_Pricing(t *testing.T) {
	provider := &scalewayCloudProvider{}

	pricingModel, err := provider.Pricing()
	require.NoError(t, err)
	assert.NotNil(t, pricingModel)
	assert.Equal(t, provider, pricingModel)
}

func TestScalewayCloudProvider_GetAvailableMachineTypes(t *testing.T) {
	provider := &scalewayCloudProvider{}

	machineTypes, err := provider.GetAvailableMachineTypes()
	require.NoError(t, err)
	assert.Empty(t, machineTypes)
}

func TestScalewayCloudProvider_NewNodeGroup(t *testing.T) {
	provider := &scalewayCloudProvider{}

	ng, err := provider.NewNodeGroup("DEV1-M", nil, nil, nil, nil)
	assert.Error(t, err)
	assert.Equal(t, cloudprovider.ErrNotImplemented, err)
	assert.Nil(t, ng)
}

func TestScalewayCloudProvider_GetResourceLimiter(t *testing.T) {
	limiter := &cloudprovider.ResourceLimiter{}
	provider := &scalewayCloudProvider{
		resourceLimiter: limiter,
	}

	result, err := provider.GetResourceLimiter()
	require.NoError(t, err)
	assert.Equal(t, limiter, result)
}

func TestScalewayCloudProvider_GPULabel(t *testing.T) {
	provider := &scalewayCloudProvider{}

	label := provider.GPULabel()
	assert.Equal(t, "k8s.scw.cloud/gpu", label)
}

func TestScalewayCloudProvider_GetAvailableGPUTypes(t *testing.T) {
	provider := &scalewayCloudProvider{}

	gpuTypes := provider.GetAvailableGPUTypes()
	assert.Nil(t, gpuTypes)
}

func TestScalewayCloudProvider_NodePrice(t *testing.T) {
	t.Run("node price calculation", func(t *testing.T) {
		client := new(mockClient)
		pool := createTestPool("pool-1", true, 3, 1, 10)
		pool.NodePricePerHour = 0.10
		node := createTestNode("node-1", "pool-1", "scaleway://fr-par-1/instance-1", scalewaygo.NodeStatusReady)

		provider := &scalewayCloudProvider{
			client:    client,
			clusterID: "test-cluster",
			nodeGroups: map[string]*NodeGroup{
				"pool-1": {
					Client: client,
					pool:   pool,
					nodes: map[string]*scalewaygo.Node{
						"scaleway://fr-par-1/instance-1": &node,
					},
				},
			},
		}

		k8sNode := &apiv1.Node{
			Spec: apiv1.NodeSpec{
				ProviderID: "scaleway://fr-par-1/instance-1",
			},
		}

		startTime := time.Now()
		endTime := startTime.Add(2*time.Hour + 30*time.Minute)

		price, err := provider.NodePrice(k8sNode, startTime, endTime)
		require.NoError(t, err)
		// 2.5 hours rounds up to 3 hours at $0.10/hour = $0.30
		assert.InDelta(t, 0.30, price, 0.001)
	})

	t.Run("node not found", func(t *testing.T) {
		client := new(mockClient)
		provider := &scalewayCloudProvider{
			client:     client,
			clusterID:  "test-cluster",
			nodeGroups: make(map[string]*NodeGroup),
		}

		k8sNode := &apiv1.Node{
			Spec: apiv1.NodeSpec{
				ProviderID: "scaleway://fr-par-1/instance-nonexistent",
			},
		}

		startTime := time.Now()
		endTime := startTime.Add(1 * time.Hour)

		price, err := provider.NodePrice(k8sNode, startTime, endTime)
		require.NoError(t, err)
		assert.Equal(t, 0.0, price)
	})
}

func TestScalewayCloudProvider_PodPrice(t *testing.T) {
	provider := &scalewayCloudProvider{}

	pod := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-pod",
		},
	}

	startTime := time.Now()
	endTime := startTime.Add(1 * time.Hour)

	price, err := provider.PodPrice(pod, startTime, endTime)
	require.NoError(t, err)
	assert.Equal(t, 0.0, price)
}

func TestScalewayCloudProvider_Cleanup(t *testing.T) {
	provider := &scalewayCloudProvider{}

	err := provider.Cleanup()
	assert.NoError(t, err)
}
