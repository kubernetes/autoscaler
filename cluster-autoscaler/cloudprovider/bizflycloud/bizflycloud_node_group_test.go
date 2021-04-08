/*
Copyright 2021 The Kubernetes Authors.

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

package bizflycloud

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/bizflycloud/gobizfly"
)

func TestNodeGroup_TargetSize(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		numberOfNodes := 3

		client := &bizflyClientMock{}
		ng := testNodeGroup(client, &gobizfly.WorkerPoolWithNodes{
			ExtendedWorkerPool: gobizfly.ExtendedWorkerPool{
				WorkerPool: gobizfly.WorkerPool{
					DesiredSize: numberOfNodes,
				},
			},
		})

		size, err := ng.TargetSize()
		assert.NoError(t, err)
		assert.Equal(t, numberOfNodes, size, "target size is not correct")
	})
}

func TestNodeGroup_IncreaseSize(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		numberOfNodes := 3
		client := &bizflyClientMock{}
		ng := testNodeGroup(client, &gobizfly.WorkerPoolWithNodes{
			ExtendedWorkerPool: gobizfly.ExtendedWorkerPool{
				WorkerPool: gobizfly.WorkerPool{
					DesiredSize: numberOfNodes,
					MaxSize:     10,
				},
			},
		})

		delta := 2

		newCount := numberOfNodes + delta
		client.On("UpdateClusterWorkerPool", ctx, ng.clusterID, ng.id, &gobizfly.UpdateWorkerPoolRequest{
			DesiredSize: newCount,
		},
			nil).Return(nil).Once()

		err := ng.IncreaseSize(delta)
		assert.NoError(t, err)
	})

	t.Run("successful increase to maximum", func(t *testing.T) {
		numberOfNodes := 2
		maxNodes := 3
		client := &bizflyClientMock{}
		ng := testNodeGroup(client, &gobizfly.WorkerPoolWithNodes{
			ExtendedWorkerPool: gobizfly.ExtendedWorkerPool{
				WorkerPool: gobizfly.WorkerPool{
					DesiredSize:       numberOfNodes,
					EnableAutoScaling: true,
					MinSize:           1,
					MaxSize:           maxNodes,
				},
			},
		})

		delta := 1
		newCount := numberOfNodes + delta
		client.On("UpdateClusterWorkerPool", ctx, ng.clusterID, ng.id, &gobizfly.UpdateWorkerPoolRequest{
			DesiredSize: newCount,
		}, nil).Return(nil).Once()

		err := ng.IncreaseSize(delta)
		assert.NoError(t, err)
	})

	t.Run("negative increase", func(t *testing.T) {
		numberOfNodes := 3
		delta := -1
		client := &bizflyClientMock{}
		ng := testNodeGroup(client, &gobizfly.WorkerPoolWithNodes{
			ExtendedWorkerPool: gobizfly.ExtendedWorkerPool{
				WorkerPool: gobizfly.WorkerPool{
					DesiredSize: numberOfNodes,
				},
			},
		})
		err := ng.IncreaseSize(delta)

		exp := fmt.Errorf("delta must be positive, have: %d", delta)
		assert.EqualError(t, err, exp.Error(), "size increase must be positive")
	})

	t.Run("zero increase", func(t *testing.T) {
		numberOfNodes := 3
		delta := 0
		client := &bizflyClientMock{}
		ng := testNodeGroup(client, &gobizfly.WorkerPoolWithNodes{
			ExtendedWorkerPool: gobizfly.ExtendedWorkerPool{
				WorkerPool: gobizfly.WorkerPool{
					DesiredSize: numberOfNodes,
				},
			},
		})

		exp := fmt.Errorf("delta must be positive, have: %d", delta)

		err := ng.IncreaseSize(delta)
		assert.EqualError(t, err, exp.Error(), "size increase must be positive")
	})

	t.Run("large increase above maximum", func(t *testing.T) {
		numberOfNodes := 195
		delta := 10

		client := &bizflyClientMock{}
		ng := testNodeGroup(client, &gobizfly.WorkerPoolWithNodes{
			ExtendedWorkerPool: gobizfly.ExtendedWorkerPool{
				WorkerPool: gobizfly.WorkerPool{
					DesiredSize: numberOfNodes,
				},
			},
		})

		exp := fmt.Errorf("size increase is too large. current: %d desired: %d max: %d",
			numberOfNodes, numberOfNodes+delta, ng.MaxSize())

		err := ng.IncreaseSize(delta)
		assert.EqualError(t, err, exp.Error(), "size increase is too large")
	})
}

func TestNodeGroup_DecreaseTargetSize(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		numberOfNodes := 5
		client := &bizflyClientMock{}
		ng := testNodeGroup(client, &gobizfly.WorkerPoolWithNodes{
			ExtendedWorkerPool: gobizfly.ExtendedWorkerPool{
				WorkerPool: gobizfly.WorkerPool{
					DesiredSize: numberOfNodes,
				},
			},
		})

		delta := -2

		newCount := numberOfNodes + delta
		client.On("UpdateClusterWorkerPool", ctx, ng.clusterID, ng.id, &gobizfly.UpdateWorkerPoolRequest{
			DesiredSize: newCount,
		}, nil,
		).Return(nil).Once()

		err := ng.DecreaseTargetSize(delta)
		assert.NoError(t, err)
	})

	t.Run("successful decrease to minimum", func(t *testing.T) {
		client := &bizflyClientMock{}
		ng := testNodeGroup(client, &gobizfly.WorkerPoolWithNodes{
			ExtendedWorkerPool: gobizfly.ExtendedWorkerPool{
				WorkerPool: gobizfly.WorkerPool{
					DesiredSize:       2,
					EnableAutoScaling: true,
					MinSize:           1,
					MaxSize:           5,
				},
			},
		})

		delta := -1
		newCount := ng.nodePool.DesiredSize + delta
		client.On("UpdateClusterWorkerPool", ctx, ng.clusterID, ng.id, &gobizfly.UpdateWorkerPoolRequest{
			DesiredSize: newCount,
		}, nil,
		).Return(nil).Once()

		err := ng.DecreaseTargetSize(delta)
		assert.NoError(t, err)
	})

	t.Run("positive decrease", func(t *testing.T) {
		numberOfNodes := 5
		client := &bizflyClientMock{}
		ng := testNodeGroup(client, &gobizfly.WorkerPoolWithNodes{
			ExtendedWorkerPool: gobizfly.ExtendedWorkerPool{
				WorkerPool: gobizfly.WorkerPool{
					DesiredSize: numberOfNodes,
				},
			},
		})

		delta := 1
		err := ng.DecreaseTargetSize(delta)

		exp := fmt.Errorf("delta must be negative, have: %d", delta)
		assert.EqualError(t, err, exp.Error(), "size decrease must be negative")
	})

	t.Run("zero decrease", func(t *testing.T) {
		numberOfNodes := 5
		client := &bizflyClientMock{}
		ng := testNodeGroup(client, &gobizfly.WorkerPoolWithNodes{
			ExtendedWorkerPool: gobizfly.ExtendedWorkerPool{
				WorkerPool: gobizfly.WorkerPool{
					DesiredSize: numberOfNodes,
				},
			},
		})

		delta := 0
		exp := fmt.Errorf("delta must be negative, have: %d", delta)

		err := ng.DecreaseTargetSize(delta)
		assert.EqualError(t, err, exp.Error(), "size decrease must be negative")
	})

	t.Run("small decrease below minimum", func(t *testing.T) {
		delta := -2
		numberOfNodes := 3
		client := &bizflyClientMock{}
		ng := testNodeGroup(client, &gobizfly.WorkerPoolWithNodes{
			ExtendedWorkerPool: gobizfly.ExtendedWorkerPool{
				WorkerPool: gobizfly.WorkerPool{
					DesiredSize: numberOfNodes,
					MinSize:     2,
					MaxSize:     5,
				},
			},
		})

		exp := fmt.Errorf("size decrease is too small. current: %d desired: %d min: %d",
			numberOfNodes, numberOfNodes+delta, ng.MinSize())
		err := ng.DecreaseTargetSize(delta)
		assert.EqualError(t, err, exp.Error(), "size decrease is too small")
	})
}

func TestNodeGroup_DeleteNodes(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		client := &bizflyClientMock{}
		ng := testNodeGroup(client, &gobizfly.WorkerPoolWithNodes{
			ExtendedWorkerPool: gobizfly.ExtendedWorkerPool{
				WorkerPool: gobizfly.WorkerPool{
					DesiredSize: 3,
				},
			},
		})

		nodes := []*apiv1.Node{
			{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{nodeIDLabel: "1"}}},
			{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{nodeIDLabel: "2"}}},
			{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{nodeIDLabel: "3"}}},
		}

		// this should be called three times (the number of nodes)
		client.On("DeleteClusterWorkerPoolNode", ctx, ng.clusterID, ng.id, "1", nil).Return(nil).Once()
		client.On("DeleteClusterWorkerPoolNode", ctx, ng.clusterID, ng.id, "2", nil).Return(nil).Once()
		client.On("DeleteClusterWorkerPoolNode", ctx, ng.clusterID, ng.id, "3", nil).Return(nil).Once()

		err := ng.DeleteNodes(nodes)
		assert.NoError(t, err)
	})

	t.Run("client deleting node fails", func(t *testing.T) {
		client := &bizflyClientMock{}
		ng := testNodeGroup(client, &gobizfly.WorkerPoolWithNodes{
			ExtendedWorkerPool: gobizfly.ExtendedWorkerPool{
				WorkerPool: gobizfly.WorkerPool{
					DesiredSize: 3,
				},
			},
		})

		nodes := []*apiv1.Node{
			{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{nodeIDLabel: "1"}}},
			{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{nodeIDLabel: "2"}}},
			{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{nodeIDLabel: "3"}}},
		}

		// client is called twice, first run is successfully but the second one
		// fails with a random error. In this case DeleteNodes() should return
		// immediately.
		client.On("DeleteClusterWorkerPoolNode", ctx, ng.clusterID, ng.id, "1", nil).Return(nil).Once()
		client.On("DeleteClusterWorkerPoolNode", ctx, ng.clusterID, ng.id, "2", nil).Return(errors.New("random error")).Once()

		err := ng.DeleteNodes(nodes)
		assert.Error(t, err)
	})
}

func TestNodeGroup_Nodes(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		client := &bizflyClientMock{}
		ng := testNodeGroup(client, &gobizfly.WorkerPoolWithNodes{
			ExtendedWorkerPool: gobizfly.ExtendedWorkerPool{
				WorkerPool: gobizfly.WorkerPool{
					DesiredSize: 2,
				},
			},
			Nodes: []gobizfly.PoolNode{
				{
					PhysicalID: "xxxxxxxx1",
				},
				{
					PhysicalID: "xxxxxxxx2",
				},
			},
		})

		exp := []cloudprovider.Instance{
			{
				Id: "bizflycloud://xxxxxxxx1",
			},
			{
				Id: "bizflycloud://xxxxxxxx2",
			},
		}

		nodes, err := ng.Nodes()
		assert.NoError(t, err)
		assert.Equal(t, exp, nodes, "nodes do not match")
	})

	t.Run("failure (nil node pool)", func(t *testing.T) {
		client := &bizflyClientMock{}
		ng := testNodeGroup(client, nil)

		_, err := ng.Nodes()
		assert.Error(t, err, "Nodes() should return an error")
	})
}

func TestNodeGroup_Debug(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		client := &bizflyClientMock{}
		ng := testNodeGroup(client, &gobizfly.WorkerPoolWithNodes{
			ExtendedWorkerPool: gobizfly.ExtendedWorkerPool{
				WorkerPool: gobizfly.WorkerPool{
					DesiredSize: 3,
					MinSize:     1,
					MaxSize:     200,
				},
			},
		})

		d := ng.Debug()
		exp := "cluster ID: 1 (min:1 max:200)"
		assert.Equal(t, exp, d, "debug string do not match")
	})
}

func TestNodeGroup_Exist(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		client := &bizflyClientMock{}
		ng := testNodeGroup(client, &gobizfly.WorkerPoolWithNodes{
			ExtendedWorkerPool: gobizfly.ExtendedWorkerPool{
				WorkerPool: gobizfly.WorkerPool{
					DesiredSize: 3,
				},
			},
		})

		exist := ng.Exist()
		assert.Equal(t, true, exist, "node pool should exist")
	})

	t.Run("failure", func(t *testing.T) {
		client := &bizflyClientMock{}
		ng := testNodeGroup(client, nil)

		exist := ng.Exist()
		assert.Equal(t, false, exist, "node pool should not exist")
	})
}

func testNodeGroup(client *bizflyClientMock, np *gobizfly.WorkerPoolWithNodes) *NodeGroup {
	var minNodes, maxNodes int
	if np != nil {
		minNodes = np.MinSize
		maxNodes = np.MaxSize
	}
	return &NodeGroup{
		id:        "1",
		clusterID: "1",
		client:    client,
		nodePool:  np,
		minSize:   minNodes,
		maxSize:   maxNodes,
	}
}

type bizflyClientMock struct {
	mock.Mock
}

func (m *bizflyClientMock) Get(ctx context.Context, id string) (*gobizfly.FullCluster, error) {
	args := m.Called(ctx, id, nil)
	return args.Get(0).(*gobizfly.FullCluster), args.Error(1)
}

func (m *bizflyClientMock) GetClusterWorkerPool(ctx context.Context, clusterUID string, PoolID string) (*gobizfly.WorkerPoolWithNodes, error) {
	args := m.Called(ctx, clusterUID, PoolID)
	return args.Get(0).(*gobizfly.WorkerPoolWithNodes), args.Error(1)
}

func (m *bizflyClientMock) UpdateClusterWorkerPool(ctx context.Context, clusterUID string, PoolID string, uwp *gobizfly.UpdateWorkerPoolRequest) error {
	args := m.Called(ctx, clusterUID, PoolID, uwp, nil)
	return args.Error(0)
}

func (m *bizflyClientMock) DeleteClusterWorkerPoolNode(ctx context.Context, clusterUID string, PoolID string, NodeID string) error {
	args := m.Called(ctx, clusterUID, PoolID, NodeID, nil)
	return args.Error(0)
}
