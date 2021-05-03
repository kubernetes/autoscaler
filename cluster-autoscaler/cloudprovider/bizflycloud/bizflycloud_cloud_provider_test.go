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
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/bizflycloud/gobizfly"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
)

func testCloudProvider(t *testing.T, client *bizflyClientMock) *bizflycloudCloudProvider {
	cfg := `{"cluster_id": "123456", "token": "123123123", "url": "https://manage.bizflycloud.vn", "version": "test"}`

	manager, err := newManagerTest(bytes.NewBufferString(cfg))
	assert.NoError(t, err)
	rl := &cloudprovider.ResourceLimiter{}

	// fill the test provider with some example
	if client == nil {
		client = &bizflyClientMock{}
		ctx := context.Background()

		client.On("Get", ctx, manager.clusterID, nil).Return(&gobizfly.FullCluster{
			ExtendedCluster: gobizfly.ExtendedCluster{
				Cluster: gobizfly.Cluster{
					UID:              "cluster-1",
					Name:             "test-cluster",
					WorkerPoolsCount: 4,
				},
				WorkerPools: []gobizfly.ExtendedWorkerPool{
					{
						WorkerPool: gobizfly.WorkerPool{},
						UID:        "pool-1",
					},
				},
			},
			Stat: gobizfly.ClusterStat{},
		}, nil).Once()

		client.On("GetClusterWorkerPool", ctx, manager.clusterID, "pool-1", nil).Return(&gobizfly.WorkerPoolWithNodes{
			ExtendedWorkerPool: gobizfly.ExtendedWorkerPool{
				WorkerPool: gobizfly.WorkerPool{
					EnableAutoScaling: true,
				},
				UID: "pool-1",
			},
			Nodes: []gobizfly.PoolNode{
				{
					ID:   "1",
					Name: "node-1",
				},
				{
					ID:   "2",
					Name: "node-2",
				},
			},
		}, nil).Once()
	}

	manager.client = client
	provider, err := newBizflyCloudProvider(manager, rl)
	assert.NoError(t, err)
	return provider
}

func TestNewBizflyCloudProvider(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		_ = testCloudProvider(t, nil)
	})
}

func TestBizflyCloudProvider_Name(t *testing.T) {
	provider := testCloudProvider(t, nil)

	t.Run("success", func(t *testing.T) {
		name := provider.Name()
		assert.Equal(t, cloudprovider.BizflyCloudProviderName, name, "provider name doesn't match")
	})
}

func TestBizflyCloudProvider_NodeGroups(t *testing.T) {
	provider := testCloudProvider(t, nil)

	t.Run("zero groups", func(t *testing.T) {
		provider.manager.nodeGroups = []*NodeGroup{}
		nodes := provider.NodeGroups()
		assert.Equal(t, len(nodes), 0, "number of nodes do not match")
	})
}

func TestBizflyCloudProvider_NodeGroupForNode(t *testing.T) {
	clusterID := "123456"

	t.Run("success", func(t *testing.T) {
		client := &bizflyClientMock{}
		ctx := context.Background()

		client.On("Get", ctx, clusterID, nil).Return(
			&gobizfly.FullCluster{},
			nil,
		).Once()

		provider := testCloudProvider(t, client)

		node := &apiv1.Node{
			Spec: apiv1.NodeSpec{
				ProviderID: toProviderID("droplet-4"),
			},
		}
		nodeGroup, err := provider.NodeGroupForNode(node)
		assert.NoError(t, err)
		assert.Nil(t, nodeGroup)
	})
	t.Run("node does not exist", func(t *testing.T) {
		client := &bizflyClientMock{}
		ctx := context.Background()

		client.On("Get", ctx, clusterID, nil).Return(
			&gobizfly.FullCluster{},
			nil,
		).Once()

		provider := testCloudProvider(t, client)

		node := &apiv1.Node{
			Spec: apiv1.NodeSpec{
				ProviderID: toProviderID("xxxxx-7"),
			},
		}

		nodeGroup, err := provider.NodeGroupForNode(node)
		assert.NoError(t, err)
		assert.Nil(t, nodeGroup)
	})
}
