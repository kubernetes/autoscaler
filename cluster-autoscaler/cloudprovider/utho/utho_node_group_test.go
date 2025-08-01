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
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uthoplatforms/utho-go/utho"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
)

func TestNodeGroup_Debug(t *testing.T) {
	client := &uthoClientMock{}
	ng := testNodeGroup(client, &utho.NodepoolDetails{
		Count:    3,
		MinNodes: 1,
		MaxNodes: 10,
	})

	d := ng.Debug()
	exp := "node group ID: pool-123 (min:1 max:10)"
	assert.Equal(t, exp, d, "debug string do not match")
}

func TestNodeGroup_TargetSize(t *testing.T) {
	numberOfNodes := 5

	client := &uthoClientMock{}
	ng := testNodeGroup(client, &utho.NodepoolDetails{
		Count: numberOfNodes,
	})

	size, err := ng.TargetSize()
	assert.NoError(t, err)
	assert.Equal(t, numberOfNodes, size, "target size is not correct")
}

func TestNodeGroup_TargetSize_WithZeroCount(t *testing.T) {
	client := &uthoClientMock{}
	ng := testNodeGroup(client, &utho.NodepoolDetails{
		Count: 0,
	})

	size, err := ng.TargetSize()
	assert.NoError(t, err)
	assert.Equal(t, 0, size, "target size should be zero")
}

func TestNodeGroup_IncreaseSize(t *testing.T) {
	client := &uthoClientMock{}

	t.Run("basic increase", func(t *testing.T) {
		numberOfNodes := 2
		delta := 1
		ng := testNodeGroup(client, &utho.NodepoolDetails{
			Count:    numberOfNodes,
			MinNodes: 2,
			MaxNodes: 3,
		})

		newQuant := numberOfNodes + delta
		client.On("UpdateNodePool", context.Background(),
			utho.UpdateKubernetesAutoscaleNodepool{
				ClusterId:  ng.clusterID,
				NodePoolId: ng.id,
				Count:      strconv.Itoa(newQuant),
			}).Return(&utho.UpdateKubernetesAutoscaleNodepoolResponse{Count: newQuant}, nil).Once()

		client.On("ReadNodePool", context.Background(), ng.clusterID, ng.id).
			Return(&utho.NodepoolDetails{
				Count:    newQuant,
				MinNodes: 2,
				MaxNodes: 3,
			}, nil).Once()

		err := ng.IncreaseSize(delta)
		assert.NoError(t, err)
		client.AssertExpectations(t)
	})

	t.Run("negative increase", func(t *testing.T) {
		numberOfNodes := 3
		delta := -1
		client := &uthoClientMock{}
		ng := testNodeGroup(client, &utho.NodepoolDetails{
			Count: numberOfNodes,
		})
		err := ng.IncreaseSize(delta)

		exp := fmt.Errorf("delta must be positive, have: %d", delta)
		assert.EqualError(t, err, exp.Error(), "size increase must be positive")
	})

	t.Run("zero increase", func(t *testing.T) {
		numberOfNodes := 3
		delta := 0
		client := &uthoClientMock{}
		ng := testNodeGroup(client, &utho.NodepoolDetails{
			Count: numberOfNodes,
		})

		exp := fmt.Errorf("delta must be positive, have: %d", delta)

		err := ng.IncreaseSize(delta)
		assert.EqualError(t, err, exp.Error(), "size increase must be positive")
	})

	t.Run("large increase above maximum", func(t *testing.T) {
		nodes := 15
		delta := 10

		client := &uthoClientMock{}
		ng := testNodeGroup(client, &utho.NodepoolDetails{
			Count: nodes,
		})

		exp := fmt.Errorf("size increase is too large. current: %d desired: %d max: %d",
			nodes, nodes+delta, ng.MaxSize())

		err := ng.IncreaseSize(delta)
		assert.EqualError(t, err, exp.Error(), "size increase is too large")
	})
}

func TestNodeGroup_DecreaseTargetSize(t *testing.T) {
	t.Run("basic decrease", func(t *testing.T) {
		client := &uthoClientMock{}

		nodeQuant := 3
		delta := -1
		ng := testNodeGroup(client, &utho.NodepoolDetails{
			Count:    nodeQuant,
			MinNodes: 2,
			MaxNodes: 3,
		})

		newQuant := nodeQuant + delta
		client.On("UpdateNodePool", context.Background(),
			utho.UpdateKubernetesAutoscaleNodepool{
				ClusterId:  ng.clusterID,
				NodePoolId: ng.id,
				Count:      strconv.Itoa(newQuant),
				Label:      uthoLabel,
				Size:       strconv.Itoa(newQuant),
			}).Return(&utho.UpdateKubernetesAutoscaleNodepoolResponse{Count: newQuant}, nil).Once()

		err := ng.DecreaseTargetSize(delta)
		assert.NoError(t, err, "expected no error when decreasing target size")
		assert.Equal(t, newQuant, ng.nodePool.Count, "node pool count should be updated")
		client.AssertExpectations(t)
	})

	t.Run("positive decrease", func(t *testing.T) {
		nodes := 5
		client := &uthoClientMock{}
		ng := testNodeGroup(client, &utho.NodepoolDetails{
			Count: nodes,
		})

		delta := 1
		err := ng.DecreaseTargetSize(delta)

		exp := fmt.Errorf("delta must be negative, have: %d", delta)
		assert.EqualError(t, err, exp.Error(), "size decrease must be negative")
	})

	t.Run("small decrease below minimum", func(t *testing.T) {
		delta := -2
		numberOfNodes := 3
		client := &uthoClientMock{}
		ng := testNodeGroup(client, &utho.NodepoolDetails{
			Count:    numberOfNodes,
			MinNodes: 2,
			MaxNodes: 5,
		})

		exp := fmt.Errorf("node group %s: size decrease is too small. current size: %d, desired size: %d, minimum size: %d",
			"pool-123", numberOfNodes, numberOfNodes+delta, ng.MinSize())
		err := ng.DecreaseTargetSize(delta)
		assert.EqualError(t, err, exp.Error(), "size decrease is too small")
	})
}

func TestNodeGroup_Nodes(t *testing.T) {
	client := &uthoClientMock{}
	ng := testNodeGroup(client, &utho.NodepoolDetails{
		Workers: []utho.WorkerNode{
			{
				ID: 123,
			},
			{
				ID: 456,
			},
		},
	})

	instances := []cloudprovider.Instance{
		{
			Id: "utho://123",
		},
		{
			Id: "utho://456",
		},
	}

	nodes, err := ng.Nodes()
	assert.NoError(t, err)
	assert.ElementsMatch(t, instances, nodes, "nodes do not match")
}

func TestNodeGroup_DeleteNodes(t *testing.T) {
	t.Run("deleting node", func(t *testing.T) {
		ctx := context.Background()
		client := &uthoClientMock{}
		ng := testNodeGroup(client, &utho.NodepoolDetails{
			Count: 2, MinNodes: 2, MaxNodes: 3, Workers: []utho.WorkerNode{{ID: 123}},
		})

		nodes := []*apiv1.Node{
			{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{nodeIDLabel: "123"}}},
		}

		client.On("DeleteNode", ctx, utho.DeleteNodeParams{
			ClusterId: 1111,
			PoolId:    ng.id,
			NodeId:    "123",
		}).Return(&utho.DeleteResponse{}, nil).Once()

		err := ng.DeleteNodes(nodes)
		assert.NoError(t, err)
		client.AssertExpectations(t)
	})

	t.Run("delete failure", func(t *testing.T) {
		ctx := context.Background()
		client := &uthoClientMock{}
		ng := testNodeGroup(client, &utho.NodepoolDetails{
			Count: 2, MinNodes: 2, MaxNodes: 3, Workers: []utho.WorkerNode{{ID: 123}},
		})

		nodes := []*apiv1.Node{
			{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{nodeIDLabel: "123"}}},
		}

		client.On("DeleteNode", ctx, utho.DeleteNodeParams{
			ClusterId: ng.clusterID,
			PoolId:    ng.id,
			NodeId:    "123",
		}).Return(&utho.DeleteResponse{}, errors.New("delete error")).Once()

		err := ng.DeleteNodes(nodes)
		assert.Error(t, err)
		errMsg := err.Error()
		assert.Contains(t, errMsg, "deleting node failed for cluster", "error should indicate node deletion failure")
		assert.Contains(t, errMsg, "node pool \"pool-123\"", "error should contain node pool ID")
		assert.Contains(t, errMsg, "node \"123\"", "error should contain node ID")
		assert.Contains(t, errMsg, "delete error", "error should contain underlying error message")
		client.AssertExpectations(t)
	})
}

func TestNodeGroup_DeleteNodes_WithMissingLabel(t *testing.T) {
	client := &uthoClientMock{}
	ng := testNodeGroup(client, &utho.NodepoolDetails{
		Count: 1,
		Workers: []utho.WorkerNode{
			{ID: 123},
		},
	})

	nodes := []*apiv1.Node{
		{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{}}},
	}
	err := ng.DeleteNodes(nodes)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "node ID label is missing")
}

func TestNodeGroup_Exist(t *testing.T) {
	client := &uthoClientMock{}
	nodeGroup := testNodeGroup(client, &utho.NodepoolDetails{MinNodes: 1, MaxNodes: 2})

	assert.Equal(t, true, nodeGroup.Exist(), "nodegroup should exist")

}

func TestNodeGroup_IncreaseSize_WithUpdateFailure(t *testing.T) {
	t.Run("API update failure", func(t *testing.T) {
		client := &uthoClientMock{}
		nodeGroup := testNodeGroup(client, &utho.NodepoolDetails{
			Count:    2,
			MinNodes: 1,
			MaxNodes: 5,
		})

		expectedErr := errors.New("API update failed")
		client.On("UpdateNodePool", context.Background(),
			utho.UpdateKubernetesAutoscaleNodepool{
				ClusterId:  nodeGroup.clusterID,
				NodePoolId: nodeGroup.id,
				Count:      "3",
			}).Return((*utho.UpdateKubernetesAutoscaleNodepoolResponse)(nil), expectedErr).Once()

		err := nodeGroup.IncreaseSize(1)
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		client.AssertExpectations(t)
	})

	t.Run("unexpected count after update", func(t *testing.T) {
		client := &uthoClientMock{}
		nodeGroup := testNodeGroup(client, &utho.NodepoolDetails{
			Count:    2,
			MinNodes: 1,
			MaxNodes: 5,
		})

		client.On("UpdateNodePool", context.Background(),
			utho.UpdateKubernetesAutoscaleNodepool{
				ClusterId:  nodeGroup.clusterID,
				NodePoolId: nodeGroup.id,
				Count:      "3",
			}).Return(&utho.UpdateKubernetesAutoscaleNodepoolResponse{Count: 2}, nil).Once()

		client.On("ReadNodePool", context.Background(), nodeGroup.clusterID, nodeGroup.id).
			Return(&utho.NodepoolDetails{
				Count:    2, // Still returns old count to simulate update failure
				MinNodes: 1,
				MaxNodes: 5,
			}, nil).Once()

		err := nodeGroup.IncreaseSize(1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "couldn't increase size")
		client.AssertExpectations(t)
	})
}

func TestNodeGroup_DeleteNodes_MultipleNodes(t *testing.T) {
	t.Run("successful multiple node deletion", func(t *testing.T) {
		ctx := context.Background()
		client := &uthoClientMock{}
		nodeGroup := testNodeGroup(client, &utho.NodepoolDetails{
			Count:    3,
			MinNodes: 1,
			MaxNodes: 5,
			Workers: []utho.WorkerNode{
				{ID: 123},
				{ID: 456},
			},
		})

		nodes := []*apiv1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{nodeIDLabel: "123"},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{nodeIDLabel: "456"},
				},
			},
		}

		// Mock both delete calls
		client.On("DeleteNode", ctx, utho.DeleteNodeParams{
			ClusterId: nodeGroup.clusterID,
			PoolId:    nodeGroup.id,
			NodeId:    "123",
		}).Return(&utho.DeleteResponse{}, nil).Once()

		client.On("DeleteNode", ctx, utho.DeleteNodeParams{
			ClusterId: nodeGroup.clusterID,
			PoolId:    nodeGroup.id,
			NodeId:    "456",
		}).Return(&utho.DeleteResponse{}, nil).Once()

		err := nodeGroup.DeleteNodes(nodes)
		assert.NoError(t, err)
		assert.Equal(t, 1, nodeGroup.nodePool.Count) // Should decrease by 2
		client.AssertExpectations(t)
	})
}

func TestNodeGroup_DecreaseTargetSize_WithValidation(t *testing.T) {
	t.Run("decrease beyond min size", func(t *testing.T) {
		client := &uthoClientMock{}
		nodeGroup := testNodeGroup(client, &utho.NodepoolDetails{
			Count:    3,
			MinNodes: 2,
			MaxNodes: 5,
		})

		err := nodeGroup.DecreaseTargetSize(-2)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "size decrease is too small")
	})

	t.Run("successful decrease", func(t *testing.T) {
		client := &uthoClientMock{}
		nodeGroup := testNodeGroup(client, &utho.NodepoolDetails{
			Count:    5,
			MinNodes: 2,
			MaxNodes: 5,
		})

		client.On("UpdateNodePool", context.Background(),
			utho.UpdateKubernetesAutoscaleNodepool{
				ClusterId:  nodeGroup.clusterID,
				NodePoolId: nodeGroup.id,
				Count:      "4",
				Label:      uthoLabel,
				Size:       "4",
			}).Return(&utho.UpdateKubernetesAutoscaleNodepoolResponse{Count: 4}, nil).Once()

		err := nodeGroup.DecreaseTargetSize(-1)
		assert.NoError(t, err)
		assert.Equal(t, 4, nodeGroup.nodePool.Count)
		client.AssertExpectations(t)
	})
}

func TestNodeGroup_Nodes_WithInvalidWorkers(t *testing.T) {
	client := &uthoClientMock{}
	nodeGroup := testNodeGroup(client, &utho.NodepoolDetails{
		Workers: []utho.WorkerNode{
			{ID: 0},  // Invalid ID
			{ID: -1}, // Invalid ID
		},
	})

	nodes, err := nodeGroup.Nodes()
	assert.NoError(t, err)
	for _, instance := range nodes {
		assert.Contains(t, instance.Id, uthoProviderIDPrefix, "provider ID should contain prefix")
	}
}

func TestNodeGroup_MinSize_Validation(t *testing.T) {
	client := &uthoClientMock{}
	ng := testNodeGroup(client, &utho.NodepoolDetails{
		MinNodes: 2,
		MaxNodes: 5,
	})

	assert.Equal(t, 2, ng.MinSize())
	assert.Less(t, ng.MinSize(), ng.MaxSize(), "min size should be less than max size")
}

func TestNodeGroup_DeleteNodes_WorkerStatus(t *testing.T) {
	t.Run("delete node with pending status", func(t *testing.T) {
		ctx := context.Background()
		client := &uthoClientMock{}
		ng := testNodeGroup(client, &utho.NodepoolDetails{
			Count: 3,
			Workers: []utho.WorkerNode{
				{ID: 123, Status: "Pending"},
			},
		})

		nodes := []*apiv1.Node{
			{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{nodeIDLabel: "123"}}},
		}

		client.On("DeleteNode", ctx, utho.DeleteNodeParams{
			ClusterId: ng.clusterID,
			PoolId:    ng.id,
			NodeId:    "123",
		}).Return(&utho.DeleteResponse{}, nil).Once()

		err := ng.DeleteNodes(nodes)
		assert.NoError(t, err)
		client.AssertExpectations(t)
	})
}

func TestNodeGroup_TemplateNodeInfo(t *testing.T) {
	t.Run("valid template node info", func(t *testing.T) {
		client := &uthoClientMock{}
		nodeGroup := testNodeGroup(client, &utho.NodepoolDetails{
			Workers: []utho.WorkerNode{
				{ID: 123, Cpu: 2, Ram: 4096},
			},
		})

		nodeInfo, err := nodeGroup.TemplateNodeInfo()
		assert.NoError(t, err)
		assert.NotNil(t, nodeInfo)

		node := nodeInfo.Node()
		assert.NotNil(t, node)
		assert.Equal(t, "linux", node.Labels["kubernetes.io/os"])
		assert.Equal(t, "amd64", node.Labels["kubernetes.io/arch"])
		assert.Equal(t, "123", node.Labels["node_id"])
	})

	t.Run("no workers in node pool", func(t *testing.T) {
		client := &uthoClientMock{}
		nodeGroup := testNodeGroup(client, &utho.NodepoolDetails{
			Workers: []utho.WorkerNode{},
		})

		nodeInfo, err := nodeGroup.TemplateNodeInfo()
		assert.Error(t, err)
		assert.Nil(t, nodeInfo)
		assert.Contains(t, err.Error(), "node pool pool-123 has no example worker")
	})

	t.Run("nil node pool", func(t *testing.T) {
		client := &uthoClientMock{}
		nodeGroup := testNodeGroup(client, nil)

		nodeInfo, err := nodeGroup.TemplateNodeInfo()
		assert.Error(t, err)
		assert.Nil(t, nodeInfo)
		assert.Contains(t, err.Error(), fmt.Sprintf("node pool %s has no example worker to derive resources", nodeGroup.id))
	})
}

func testNodeGroup(client nodeGroupClient, np *utho.NodepoolDetails) *NodeGroup {
	if np == nil {
		np = &utho.NodepoolDetails{}
	}
	return &NodeGroup{
		id:        "pool-123",
		clusterID: 1111,
		client:    client,
		nodePool:  np,
		minSize:   np.MinNodes,
		maxSize:   np.MaxNodes,
	}
}
