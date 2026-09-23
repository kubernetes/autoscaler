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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/scaleway/scalewaygo"
	"sigs.k8s.io/cluster-autoscaler/pkg/cloudprovider"
	"sigs.k8s.io/cluster-autoscaler/pkg/config"
)

func TestNodeGroup_MaxSize(t *testing.T) {
	ng := &NodeGroup{
		pool: scalewaygo.Pool{
			MaxSize: 10,
		},
	}

	assert.Equal(t, 10, ng.MaxSize(context.Background()))
}

func TestNodeGroup_MinSize(t *testing.T) {
	ng := &NodeGroup{
		pool: scalewaygo.Pool{
			MinSize: 2,
		},
	}

	assert.Equal(t, 2, ng.MinSize(context.Background()))
}

func TestNodeGroup_TargetSize(t *testing.T) {
	ng := &NodeGroup{
		pool: scalewaygo.Pool{
			Size: 5,
		},
	}

	size, err := ng.TargetSize(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 5, size)
}

func TestNodeGroup_IncreaseSize(t *testing.T) {
	t.Run("successful increase", func(t *testing.T) {
		client := new(mockClient)
		pool := scalewaygo.Pool{
			ID:        "pool-1",
			ClusterID: "test-cluster",
			Size:      3,
			MinSize:   1,
			MaxSize:   10,
			Status:    scalewaygo.PoolStatusReady,
		}

		updatedPool := pool
		updatedPool.Size = 5
		updatedPool.Status = scalewaygo.PoolStatusScaling

		client.On("UpdatePool", mock.Anything, "pool-1", 5).Return(updatedPool, nil)

		ng := &NodeGroup{
			Client: client,
			pool:   pool,
			nodes:  make(map[string]*scalewaygo.Node),
		}

		err := ng.IncreaseSize(context.Background(), 2)
		require.NoError(t, err)
		assert.Equal(t, 5, ng.pool.Size)
		assert.Equal(t, scalewaygo.PoolStatusScaling, ng.pool.Status)

		client.AssertExpectations(t)
	})

	t.Run("delta is zero", func(t *testing.T) {
		ng := &NodeGroup{
			pool: scalewaygo.Pool{
				Size:    3,
				MinSize: 1,
				MaxSize: 10,
			},
		}

		err := ng.IncreaseSize(context.Background(), 0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "delta must be strictly positive")
	})

	t.Run("delta is negative", func(t *testing.T) {
		ng := &NodeGroup{
			pool: scalewaygo.Pool{
				Size:    3,
				MinSize: 1,
				MaxSize: 10,
			},
		}

		err := ng.IncreaseSize(context.Background(), -1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "delta must be strictly positive")
	})

	t.Run("exceeds max size", func(t *testing.T) {
		ng := &NodeGroup{
			pool: scalewaygo.Pool{
				Size:    8,
				MinSize: 1,
				MaxSize: 10,
			},
		}

		err := ng.IncreaseSize(context.Background(), 5)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "size increase is too large")
	})

	t.Run("negative target size from overflow", func(t *testing.T) {
		ng := &NodeGroup{
			pool: scalewaygo.Pool{
				Size:    -5, // Simulating corrupted state
				MinSize: 0,
				MaxSize: 10,
			},
		}

		// Starting from negative size (corrupted state)
		err := ng.IncreaseSize(context.Background(), 2)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "size cannot be negative")
	})

	t.Run("API error", func(t *testing.T) {
		client := new(mockClient)
		pool := scalewaygo.Pool{
			ID:        "pool-1",
			ClusterID: "test-cluster",
			Size:      3,
			MinSize:   1,
			MaxSize:   10,
		}

		client.On("UpdatePool", mock.Anything, "pool-1", 5).Return(
			scalewaygo.Pool{},
			fmt.Errorf("API error"),
		)

		ng := &NodeGroup{
			Client: client,
			pool:   pool,
			nodes:  make(map[string]*scalewaygo.Node),
		}

		err := ng.IncreaseSize(context.Background(), 2)
		assert.Error(t, err)

		client.AssertExpectations(t)
	})
}

func TestNodeGroup_DecreaseTargetSize(t *testing.T) {
	t.Run("successful decrease", func(t *testing.T) {
		client := new(mockClient)
		pool := scalewaygo.Pool{
			ID:        "pool-1",
			ClusterID: "test-cluster",
			Size:      5,
			MinSize:   1,
			MaxSize:   10,
			Status:    scalewaygo.PoolStatusReady,
		}

		updatedPool := pool
		updatedPool.Size = 3
		updatedPool.Status = scalewaygo.PoolStatusScaling

		client.On("UpdatePool", mock.Anything, "pool-1", 3).Return(updatedPool, nil)

		ng := &NodeGroup{
			Client: client,
			pool:   pool,
			nodes:  make(map[string]*scalewaygo.Node),
		}

		err := ng.DecreaseTargetSize(context.Background(), -2)
		require.NoError(t, err)
		assert.Equal(t, 3, ng.pool.Size)
		assert.Equal(t, scalewaygo.PoolStatusScaling, ng.pool.Status)

		client.AssertExpectations(t)
	})

	t.Run("delta is zero", func(t *testing.T) {
		ng := &NodeGroup{
			pool: scalewaygo.Pool{
				Size:    3,
				MinSize: 1,
				MaxSize: 10,
			},
		}

		err := ng.DecreaseTargetSize(context.Background(), 0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "delta must be strictly negative")
	})

	t.Run("delta is positive", func(t *testing.T) {
		ng := &NodeGroup{
			pool: scalewaygo.Pool{
				Size:    3,
				MinSize: 1,
				MaxSize: 10,
			},
		}

		err := ng.DecreaseTargetSize(context.Background(), 1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "delta must be strictly negative")
	})

	t.Run("below min size", func(t *testing.T) {
		ng := &NodeGroup{
			pool: scalewaygo.Pool{
				Size:    5,
				MinSize: 3,
				MaxSize: 10,
			},
		}

		// 5 + (-3) = 2, which is below min (3) but not negative
		err := ng.DecreaseTargetSize(context.Background(), -3)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "size decrease is too large")
	})

	t.Run("negative target size", func(t *testing.T) {
		ng := &NodeGroup{
			pool: scalewaygo.Pool{
				Size:    2,
				MinSize: 0,
				MaxSize: 10,
			},
		}

		// Attempting to decrease by more than current size would result in negative
		// 2 + (-5) = -3
		err := ng.DecreaseTargetSize(context.Background(), -5)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "size cannot be negative")
	})

	t.Run("API error", func(t *testing.T) {
		client := new(mockClient)
		pool := scalewaygo.Pool{
			ID:        "pool-1",
			ClusterID: "test-cluster",
			Size:      5,
			MinSize:   1,
			MaxSize:   10,
		}

		client.On("UpdatePool", mock.Anything, "pool-1", 3).Return(
			scalewaygo.Pool{},
			fmt.Errorf("API error"),
		)

		ng := &NodeGroup{
			Client: client,
			pool:   pool,
			nodes:  make(map[string]*scalewaygo.Node),
		}

		err := ng.DecreaseTargetSize(context.Background(), -2)
		assert.Error(t, err)

		client.AssertExpectations(t)
	})
}

func TestNodeGroup_DeleteNodes(t *testing.T) {
	t.Run("successful delete", func(t *testing.T) {
		client := new(mockClient)
		node1 := createTestNode("node-1", "pool-1", "scaleway://fr-par-1/instance-1", scalewaygo.NodeStatusReady)
		node2 := createTestNode("node-2", "pool-1", "scaleway://fr-par-1/instance-2", scalewaygo.NodeStatusReady)

		deletedNode1 := node1
		deletedNode1.Status = scalewaygo.NodeStatusDeleting
		deletedNode2 := node2
		deletedNode2.Status = scalewaygo.NodeStatusDeleting

		client.On("DeleteNode", mock.Anything, "node-1").Return(deletedNode1, nil)
		client.On("DeleteNode", mock.Anything, "node-2").Return(deletedNode2, nil)

		ng := &NodeGroup{
			Client: client,
			pool: scalewaygo.Pool{
				ID:   "pool-1",
				Size: 5,
			},
			nodes: map[string]*scalewaygo.Node{
				"scaleway://fr-par-1/instance-1": &node1,
				"scaleway://fr-par-1/instance-2": &node2,
			},
		}

		k8sNodes := []*apiv1.Node{
			{Spec: apiv1.NodeSpec{ProviderID: "scaleway://fr-par-1/instance-1"}},
			{Spec: apiv1.NodeSpec{ProviderID: "scaleway://fr-par-1/instance-2"}},
		}

		err := ng.DeleteNodes(context.Background(), k8sNodes)
		require.NoError(t, err)
		assert.Equal(t, 3, ng.pool.Size)
		assert.Equal(t, scalewaygo.NodeStatusDeleting, ng.nodes["scaleway://fr-par-1/instance-1"].Status)
		assert.Equal(t, scalewaygo.NodeStatusDeleting, ng.nodes["scaleway://fr-par-1/instance-2"].Status)

		client.AssertExpectations(t)
	})

	t.Run("node not found in pool", func(t *testing.T) {
		client := new(mockClient)

		ng := &NodeGroup{
			Client: client,
			pool: scalewaygo.Pool{
				ID:   "pool-1",
				Size: 3,
			},
			nodes: make(map[string]*scalewaygo.Node),
		}

		k8sNodes := []*apiv1.Node{
			{Spec: apiv1.NodeSpec{ProviderID: "scaleway://fr-par-1/instance-nonexistent"}},
		}

		err := ng.DeleteNodes(context.Background(), k8sNodes)
		// Should not error, just log and continue
		require.NoError(t, err)
		assert.Equal(t, 3, ng.pool.Size) // Size unchanged

		client.AssertExpectations(t)
	})

	t.Run("API error", func(t *testing.T) {
		client := new(mockClient)
		node1 := createTestNode("node-1", "pool-1", "scaleway://fr-par-1/instance-1", scalewaygo.NodeStatusReady)

		client.On("DeleteNode", mock.Anything, "node-1").Return(
			scalewaygo.Node{},
			fmt.Errorf("API error"),
		)

		ng := &NodeGroup{
			Client: client,
			pool: scalewaygo.Pool{
				ID:   "pool-1",
				Size: 3,
			},
			nodes: map[string]*scalewaygo.Node{
				"scaleway://fr-par-1/instance-1": &node1,
			},
		}

		k8sNodes := []*apiv1.Node{
			{Spec: apiv1.NodeSpec{ProviderID: "scaleway://fr-par-1/instance-1"}},
		}

		err := ng.DeleteNodes(context.Background(), k8sNodes)
		assert.Error(t, err)

		client.AssertExpectations(t)
	})
}

func TestNodeGroup_ForceDeleteNodes(t *testing.T) {
	ng := &NodeGroup{}

	err := ng.ForceDeleteNodes(context.Background(), []*apiv1.Node{})
	assert.Equal(t, cloudprovider.ErrNotImplemented, err)
}

func TestNodeGroup_AtomicIncreaseSize(t *testing.T) {
	ng := &NodeGroup{}

	err := ng.AtomicIncreaseSize(context.Background(), 1)
	assert.Equal(t, cloudprovider.ErrNotImplemented, err)
}

func TestNodeGroup_Id(t *testing.T) {
	ng := &NodeGroup{
		pool: scalewaygo.Pool{
			ID: "pool-123",
		},
	}

	assert.Equal(t, "pool-123", ng.Id())
}

func TestNodeGroup_Debug(t *testing.T) {
	ng := &NodeGroup{
		pool: scalewaygo.Pool{
			ID:          "pool-123",
			Status:      scalewaygo.PoolStatusReady,
			Version:     "1.27.0",
			Autoscaling: true,
			Size:        5,
			MinSize:     1,
			MaxSize:     10,
		},
	}

	debug := ng.Debug(context.Background())
	assert.Contains(t, debug, "pool-123")
	assert.Contains(t, debug, "ready")
	assert.Contains(t, debug, "1.27.0")
	assert.Contains(t, debug, "true")
	assert.Contains(t, debug, "size:5")
	assert.Contains(t, debug, "min_size:1")
	assert.Contains(t, debug, "max_size:10")
}

func TestNodeGroup_Nodes(t *testing.T) {
	t.Run("returns all nodes", func(t *testing.T) {
		node1 := createTestNode("node-1", "pool-1", "scaleway://fr-par-1/instance-1", scalewaygo.NodeStatusReady)
		node2 := createTestNode("node-2", "pool-1", "scaleway://fr-par-1/instance-2", scalewaygo.NodeStatusCreating)
		node3 := createTestNode("node-3", "pool-1", "scaleway://fr-par-1/instance-3", scalewaygo.NodeStatusDeleting)

		ng := &NodeGroup{
			pool: scalewaygo.Pool{
				ID: "pool-1",
			},
			nodes: map[string]*scalewaygo.Node{
				"scaleway://fr-par-1/instance-1": &node1,
				"scaleway://fr-par-1/instance-2": &node2,
				"scaleway://fr-par-1/instance-3": &node3,
			},
		}

		instances, err := ng.Nodes(context.Background())
		require.NoError(t, err)
		assert.Len(t, instances, 3)

		// Verify instances have correct data
		providerIDs := make(map[string]bool)
		for _, inst := range instances {
			providerIDs[inst.Id] = true
			assert.NotNil(t, inst.Status)
		}

		assert.True(t, providerIDs["scaleway://fr-par-1/instance-1"])
		assert.True(t, providerIDs["scaleway://fr-par-1/instance-2"])
		assert.True(t, providerIDs["scaleway://fr-par-1/instance-3"])
	})

	t.Run("empty node group", func(t *testing.T) {
		ng := &NodeGroup{
			pool: scalewaygo.Pool{
				ID: "pool-1",
			},
			nodes: make(map[string]*scalewaygo.Node),
		}

		instances, err := ng.Nodes(context.Background())
		require.NoError(t, err)
		assert.Empty(t, instances)
	})
}

func TestNodeGroup_TemplateNodeInfo(t *testing.T) {
	ng := &NodeGroup{
		pool: scalewaygo.Pool{
			ID:   "pool-1",
			Name: "test-pool",
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
				"kubernetes.io/hostname":           "test-node",
				"node.kubernetes.io/instance-type": "DEV1-M",
			},
			NodeTaints: []scalewaygo.Taint{
				{Key: "key1", Value: "value1", Effect: "NoSchedule"},
				{Key: "key2", Effect: "NoExecute"},
			},
		},
	}

	nodeInfo, err := ng.TemplateNodeInfo(context.Background())
	require.NoError(t, err)
	require.NotNil(t, nodeInfo)

	node := nodeInfo.Node()
	require.NotNil(t, node)

	// Verify capacity
	cpuCapacity := node.Status.Capacity[apiv1.ResourceCPU]
	assert.Equal(t, int64(2000), cpuCapacity.Value())
	memCapacity := node.Status.Capacity[apiv1.ResourceMemory]
	assert.Equal(t, int64(4096000000), memCapacity.Value())
	storageCapacity := node.Status.Capacity[apiv1.ResourceEphemeralStorage]
	assert.Equal(t, int64(20000000000), storageCapacity.Value())
	podsCapacity := node.Status.Capacity[apiv1.ResourcePods]
	assert.Equal(t, int64(110), podsCapacity.Value())

	// Verify allocatable
	cpuAllocatable := node.Status.Allocatable[apiv1.ResourceCPU]
	assert.Equal(t, int64(1800), cpuAllocatable.Value())
	memAllocatable := node.Status.Allocatable[apiv1.ResourceMemory]
	assert.Equal(t, int64(3500000000), memAllocatable.Value())
	storageAllocatable := node.Status.Allocatable[apiv1.ResourceEphemeralStorage]
	assert.Equal(t, int64(18000000000), storageAllocatable.Value())
	podsAllocatable := node.Status.Allocatable[apiv1.ResourcePods]
	assert.Equal(t, int64(110), podsAllocatable.Value())

	// Verify labels
	assert.Equal(t, "test-node", node.ObjectMeta.Labels["kubernetes.io/hostname"])
	assert.Equal(t, "DEV1-M", node.ObjectMeta.Labels["node.kubernetes.io/instance-type"])

	// Verify taints
	assert.Len(t, node.Spec.Taints, 2)
	taintMap := make(map[string]apiv1.Taint)
	for _, taint := range node.Spec.Taints {
		taintMap[taint.Key] = taint
	}
	assert.Equal(t, apiv1.TaintEffectNoSchedule, taintMap["key1"].Effect)
	assert.Equal(t, "value1", taintMap["key1"].Value)
	assert.Equal(t, apiv1.TaintEffectNoExecute, taintMap["key2"].Effect)
	assert.Equal(t, "", taintMap["key2"].Value)

	// Verify conditions
	assert.NotEmpty(t, node.Status.Conditions)
}

func TestNodeGroup_Exist(t *testing.T) {
	ng := &NodeGroup{
		pool: scalewaygo.Pool{
			ID: "pool-123",
		},
	}

	// Always returns true in current implementation
	assert.True(t, ng.Exist(context.Background()))
}

func TestNodeGroup_Create(t *testing.T) {
	ng := &NodeGroup{}

	newNg, err := ng.Create(context.Background())
	assert.Equal(t, cloudprovider.ErrNotImplemented, err)
	assert.Nil(t, newNg)
}

func TestNodeGroup_Delete(t *testing.T) {
	ng := &NodeGroup{}

	err := ng.Delete(context.Background())
	assert.Equal(t, cloudprovider.ErrNotImplemented, err)
}

func TestNodeGroup_Autoprovisioned(t *testing.T) {
	ng := &NodeGroup{}

	assert.False(t, ng.Autoprovisioned(context.Background()))
}

func TestNodeGroup_GetOptions(t *testing.T) {
	ng := &NodeGroup{}

	defaults := config.NodeGroupAutoscalingOptions{}
	opts, err := ng.GetOptions(context.Background(), defaults)
	assert.Equal(t, cloudprovider.ErrNotImplemented, err)
	assert.Nil(t, opts)
}

func TestFromScwStatus(t *testing.T) {
	tests := []struct {
		name          string
		status        scalewaygo.NodeStatus
		expectedState cloudprovider.InstanceState
		hasErrorInfo  bool
		errorCode     string
	}{
		{
			name:          "ready",
			status:        scalewaygo.NodeStatusReady,
			expectedState: cloudprovider.InstanceRunning,
			hasErrorInfo:  false,
		},
		{
			name:          "creating",
			status:        scalewaygo.NodeStatusCreating,
			expectedState: cloudprovider.InstanceCreating,
			hasErrorInfo:  false,
		},
		{
			name:          "starting",
			status:        scalewaygo.NodeStatusStarting,
			expectedState: cloudprovider.InstanceCreating,
			hasErrorInfo:  false,
		},
		{
			name:          "registering",
			status:        scalewaygo.NodeStatusRegistering,
			expectedState: cloudprovider.InstanceCreating,
			hasErrorInfo:  false,
		},
		{
			name:          "not_ready",
			status:        scalewaygo.NodeStatusNotReady,
			expectedState: cloudprovider.InstanceCreating,
			hasErrorInfo:  false,
		},
		{
			name:          "deleting",
			status:        scalewaygo.NodeStatusDeleting,
			expectedState: cloudprovider.InstanceDeleting,
			hasErrorInfo:  false,
		},
		{
			name:         "deleted",
			status:       scalewaygo.NodeStatusDeleted,
			hasErrorInfo: true,
			errorCode:    "deleted",
		},
		{
			name:          "creation_error",
			status:        scalewaygo.NodeStatusCreationError,
			expectedState: cloudprovider.InstanceCreating,
			hasErrorInfo:  true,
			errorCode:     "creation_error",
		},
		{
			name:          "upgrading",
			status:        scalewaygo.NodeStatusUpgrading,
			expectedState: cloudprovider.InstanceCreating,
			hasErrorInfo:  false,
		},
		{
			name:         "locked",
			status:       scalewaygo.NodeStatusLocked,
			hasErrorInfo: true,
			errorCode:    "locked",
		},
		{
			name:          "rebooting",
			status:        scalewaygo.NodeStatusRebooting,
			expectedState: cloudprovider.InstanceCreating,
			hasErrorInfo:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := fromScwStatus(tt.status)
			require.NotNil(t, status)
			assert.Equal(t, tt.expectedState, status.State)
			if tt.hasErrorInfo {
				require.NotNil(t, status.ErrorInfo)
				assert.Equal(t, tt.errorCode, status.ErrorInfo.ErrorCode)
			}
		})
	}
}

func TestConvertTaints(t *testing.T) {
	t.Run("convert various taint effects", func(t *testing.T) {
		taints := []scalewaygo.Taint{
			{Key: "key1", Value: "value1", Effect: "NoSchedule"},
			{Key: "key2", Value: "value2", Effect: "NoExecute"},
			{Key: "key3", Value: "value3", Effect: "PreferNoSchedule"},
			{Key: "key4", Effect: "NoSchedule"},
			{Key: "key5", Value: "invalid", Effect: "InvalidEffect"},
		}

		k8sTaints := convertTaints(taints)
		assert.Len(t, k8sTaints, 4) // key5 should be skipped

		taintMap := make(map[string]apiv1.Taint)
		for _, taint := range k8sTaints {
			taintMap[taint.Key] = taint
		}

		// Verify key1
		assert.Equal(t, apiv1.TaintEffectNoSchedule, taintMap["key1"].Effect)
		assert.Equal(t, "value1", taintMap["key1"].Value)

		// Verify key2
		assert.Equal(t, apiv1.TaintEffectNoExecute, taintMap["key2"].Effect)
		assert.Equal(t, "value2", taintMap["key2"].Value)

		// Verify key3
		assert.Equal(t, apiv1.TaintEffectPreferNoSchedule, taintMap["key3"].Effect)
		assert.Equal(t, "value3", taintMap["key3"].Value)

		// Verify key4 (no value)
		assert.Equal(t, apiv1.TaintEffectNoSchedule, taintMap["key4"].Effect)
		assert.Equal(t, "", taintMap["key4"].Value)

		// Verify key5 is not present
		_, exists := taintMap["key5"]
		assert.False(t, exists)
	})

	t.Run("empty taints", func(t *testing.T) {
		k8sTaints := convertTaints(nil)
		assert.Empty(t, k8sTaints)

		k8sTaints = convertTaints([]scalewaygo.Taint{})
		assert.Empty(t, k8sTaints)
	})

	t.Run("valid keys and values are kept as-is", func(t *testing.T) {
		taints := []scalewaygo.Taint{
			{Key: "example.com/key1", Value: "value.with-separators_1", Effect: "NoSchedule"},
		}

		k8sTaints := convertTaints(taints)
		require.Len(t, k8sTaints, 1)
		assert.Equal(t, "example.com/key1", k8sTaints[0].Key)
		assert.Equal(t, "value.with-separators_1", k8sTaints[0].Value)
		assert.Equal(t, apiv1.TaintEffectNoSchedule, k8sTaints[0].Effect)
	})

	t.Run("taints with invalid key or value are skipped", func(t *testing.T) {
		taints := []scalewaygo.Taint{
			{Key: "key1", Value: "value:with:colons", Effect: "NoSchedule"},
			{Key: "key with spaces", Value: "value1", Effect: "NoSchedule"},
			{Key: "not/a/valid/key", Value: "value1", Effect: "NoSchedule"},
			{Key: "", Value: "value1", Effect: "NoSchedule"},
			{Key: strings.Repeat("k", 64), Value: "value1", Effect: "NoSchedule"},
			{Key: "key2", Value: strings.Repeat("v", 64), Effect: "NoSchedule"},
			{Key: "key3", Value: "value3", Effect: "NoSchedule"},
		}

		k8sTaints := convertTaints(taints)
		require.Len(t, k8sTaints, 1)
		assert.Equal(t, "key3", k8sTaints[0].Key)
		assert.Equal(t, "value3", k8sTaints[0].Value)
		assert.Equal(t, apiv1.TaintEffectNoSchedule, k8sTaints[0].Effect)
	})

	t.Run("duplicated keys are all converted", func(t *testing.T) {
		taints := []scalewaygo.Taint{
			{Key: "key1", Value: "value1", Effect: "NoSchedule"},
			{Key: "key1", Value: "value1", Effect: "NoExecute"},
		}

		k8sTaints := convertTaints(taints)
		require.Len(t, k8sTaints, 2)
		assert.Equal(t, apiv1.TaintEffectNoSchedule, k8sTaints[0].Effect)
		assert.Equal(t, apiv1.TaintEffectNoExecute, k8sTaints[1].Effect)
	})
}
