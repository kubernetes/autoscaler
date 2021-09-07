/*
Copyright 2016 The Kubernetes Authors.

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

package linode

import (
	"context"
	"net/http"
	"testing"

	"github.com/linode/linodego"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
)

func TestNodeGroup_TargetSize(t *testing.T) {
	client := &linodeClientMock{}
	clusterID := 123
	nodes := makeTestNodePoolNodes(123, 125)
	pool := makeMockNodePool(123, nodes, linodego.LKEClusterPoolAutoscaler{
		Min:     len(nodes) - 1,
		Max:     len(nodes) + 1,
		Enabled: true,
	})
	ng := nodeGroupFromPool(client, clusterID, &pool)

	size, err := ng.TargetSize()
	assert.NoError(t, err)
	assert.Equal(t, size, len(nodes))
}

func TestNodeGroup_IncreaseSize(t *testing.T) {
	client := &linodeClientMock{}
	ctx := context.Background()
	clusterID := 123
	nodes := makeTestNodePoolNodes(123, 125)
	pool := makeMockNodePool(123, nodes, linodego.LKEClusterPoolAutoscaler{
		Enabled: true,
		Min:     1,
		Max:     7,
	})
	ng := nodeGroupFromPool(client, clusterID, &pool)

	t.Run("fails with delta of 0", func(t *testing.T) {
		err := ng.IncreaseSize(0)
		assert.Error(t, err)
	})

	t.Run("fails with negative delta", func(t *testing.T) {
		err := ng.IncreaseSize(-1)
		assert.Error(t, err)
	})

	t.Run("fails when out of bounds", func(t *testing.T) {
		err := ng.IncreaseSize((ng.MaxSize() - ng.pool.Count) + 1)
		assert.Error(t, err)
	})

	t.Run("fails on generic UpdateLKEClusterPool API call error", func(t *testing.T) {
		mockErr := linodego.Error{
			Code:    http.StatusServiceUnavailable,
			Message: http.StatusText(http.StatusServiceUnavailable),
		}
		client.On("UpdateLKEClusterPool", ctx, clusterID, pool.ID,
			linodego.LKEClusterPoolUpdateOptions{Count: pool.Count + 1},
		).Return(&linodego.LKEClusterPool{}, mockErr).Once()

		err := ng.IncreaseSize(1)
		assert.Error(t, err)
	})

	t.Run("successfully add 1 node", func(t *testing.T) {
		newCount := ng.pool.Count + 1
		mockedPool := makeMockNodePool(123,
			append(ng.pool.Linodes, makeTestNodePoolNode(126)),
			pool.Autoscaler,
		)
		client.On("UpdateLKEClusterPool", ctx, clusterID, pool.ID,
			linodego.LKEClusterPoolUpdateOptions{Count: pool.Count + 1},
		).Return(&mockedPool, nil).Once()

		err := ng.IncreaseSize(1)
		assert.NoError(t, err)
		assert.Equal(t, newCount, len(ng.pool.Linodes))
	})

	t.Run("successfully add multiple nodes", func(t *testing.T) {
		newCount := ng.pool.Count + 2
		mockedPool := makeMockNodePool(123,
			append(ng.pool.Linodes, makeTestNodePoolNodes(127, 128)...),
			pool.Autoscaler,
		)
		client.On("UpdateLKEClusterPool", ctx, clusterID, pool.ID,
			linodego.LKEClusterPoolUpdateOptions{Count: 6},
		).Return(&mockedPool, nil).Once()

		err := ng.IncreaseSize(2)
		assert.NoError(t, err)
		assert.Equal(t, newCount, len(ng.pool.Linodes))
	})
}

func TestNodeGroup_DecreaseTargetSize(t *testing.T) {
	ng := &NodeGroup{}
	err := ng.DecreaseTargetSize(-1)
	assert.EqualError(t, err, cloudprovider.ErrNotImplemented.Error())
}

func TestNodeGroup_DeleteNodes(t *testing.T) {
	client := &linodeClientMock{}
	ctx := context.Background()
	pool := makeMockNodePool(123, makeTestNodePoolNodes(123, 127), linodego.LKEClusterPoolAutoscaler{
		Min:     3,
		Max:     10,
		Enabled: true,
	})
	ng := nodeGroupFromPool(client, 123, &pool)

	t.Run("fails with malformed providerID", func(t *testing.T) {
		nodes := []*v1.Node{{Spec: v1.NodeSpec{ProviderID: "linode://aaa"}}}
		err := ng.DeleteNodes(nodes)
		assert.Error(t, err)
	})

	t.Run("fails with unrelated node", func(t *testing.T) {
		nodes := []*v1.Node{{Spec: v1.NodeSpec{ProviderID: "linode://555"}}}
		err := ng.DeleteNodes(nodes)
		assert.Error(t, err)
	})

	t.Run("fails on generic DeleteLKEClusterPoolNode API call error", func(t *testing.T) {
		nodes := []*v1.Node{{Spec: v1.NodeSpec{ProviderID: "linode://123"}}}
		mockErr := linodego.Error{
			Code:    http.StatusServiceUnavailable,
			Message: http.StatusText(http.StatusServiceUnavailable),
		}
		client.On("DeleteLKEClusterPoolNode", ctx, ng.clusterID, "123").Return(mockErr).Once()

		err := ng.DeleteNodes(nodes)
		assert.Error(t, err)
	})

	t.Run("sucessfully deletes 1 node", func(t *testing.T) {
		newSize := len(pool.Linodes) - 1
		nodes := []*v1.Node{{Spec: v1.NodeSpec{ProviderID: "linode://127"}}}
		client.On("DeleteLKEClusterPoolNode", ctx, ng.clusterID, "127").Return(nil).Once()

		err := ng.DeleteNodes(nodes)
		assert.NoError(t, err)
		assert.Equal(t, newSize, len(pool.Linodes))
	})

	t.Run("successfully deletes multiple nodes", func(t *testing.T) {
		newSize := len(pool.Linodes) - 2
		nodes := []*v1.Node{
			{Spec: v1.NodeSpec{ProviderID: "linode://123"}},
			{Spec: v1.NodeSpec{ProviderID: "linode://124"}},
		}
		client.On("DeleteLKEClusterPoolNode", ctx, ng.clusterID, "123").Return(nil).Once()
		client.On("DeleteLKEClusterPoolNode", ctx, ng.clusterID, "124").Return(nil).Once()

		err := ng.DeleteNodes(nodes)
		assert.NoError(t, err)
		assert.Equal(t, newSize, len(pool.Linodes))
	})
}

func TestNodeGroup_Nodes(t *testing.T) {
	ng := NodeGroup{
		pool: &linodego.LKEClusterPool{
			ID: 123,
			Linodes: []linodego.LKEClusterPoolLinode{
				{ID: "123", InstanceID: 123},
				{ID: "124", InstanceID: 124},
				{ID: "125", InstanceID: 125},
			},
		},
	}

	instancesList, err := ng.Nodes()
	assert.NoError(t, err)
	assert.Equal(t, 3, len(instancesList))
	assert.Equal(t, instancesList, []cloudprovider.Instance{
		{Id: "linode://123"}, {Id: "linode://124"}, {Id: "linode://125"},
	})
}

func TestNodeGroup_Others(t *testing.T) {
	client := linodeClientMock{}
	ng := NodeGroup{
		client: &client,
		pool: &linodego.LKEClusterPool{
			ID: 123,
			Linodes: []linodego.LKEClusterPoolLinode{
				{ID: "123", InstanceID: 123},
				{ID: "124", InstanceID: 124},
				{ID: "125", InstanceID: 125},
				{ID: "126", InstanceID: 126},
			},
			Autoscaler: linodego.LKEClusterPoolAutoscaler{
				Min: 1,
				Max: 5,
			},
		},
	}

	assert.Equal(t, 1, ng.MinSize())
	assert.Equal(t, 5, ng.MaxSize())

	ts, err := ng.TargetSize()
	assert.NoError(t, err)
	assert.Equal(t, 4, ts)
	assert.Equal(t, "123", ng.Id())
	assert.Equal(t, "node group ID: 123 (min:1 max:5)", ng.Debug())
	assert.Equal(t, true, ng.Exist())
	assert.Equal(t, false, ng.Autoprovisioned())

	_, err = ng.TemplateNodeInfo()
	assert.EqualError(t, err, cloudprovider.ErrNotImplemented.Error())

	_, err = ng.GetOptions(config.NodeGroupAutoscalingOptions{})
	assert.EqualError(t, err, cloudprovider.ErrNotImplemented.Error())

	_, err = ng.Create()
	assert.EqualError(t, err, cloudprovider.ErrNotImplemented.Error())

	err = ng.Delete()
	assert.EqualError(t, err, cloudprovider.ErrNotImplemented.Error())
}
