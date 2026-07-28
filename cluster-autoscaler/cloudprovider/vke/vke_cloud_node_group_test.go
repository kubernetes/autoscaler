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

package vke

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/vke/sdk"
)

func newTestNodeGroup(t *testing.T) *NodeGroup {
	t.Helper()
	manager, client := newTestManager(t)
	ctx := context.Background()

	client.On("ListClusterFlavors", ctx, "clusterID").Return(
		[]sdk.Flavor{
			{Id: "flavor-1", State: "available", VCPUs: 2, GPUs: 0, RAM: 4},
			{Id: "flavor-gpu", State: "available", VCPUs: 8, GPUs: 1, RAM: 32},
		}, nil,
	)

	return &NodeGroup{
		Manager: manager,
		NodePool: sdk.NodePool{
			ID:           "pool-1",
			Name:         "workers",
			Flavor:       "flavor-1",
			MinNodes:     1,
			MaxNodes:     5,
			CurrentNodes: 2,
		},
		CurrentSize: -1,
	}
}

func TestNodeGroup_MaxMinSize(t *testing.T) {
	ng := newTestNodeGroup(t)
	assert.Equal(t, 5, ng.MaxSize())
	assert.Equal(t, 1, ng.MinSize())
	assert.Equal(t, "workers", ng.Id())
}

func TestNodeGroup_TargetSize(t *testing.T) {
	ng := newTestNodeGroup(t)

	size, err := ng.TargetSize()
	assert.NoError(t, err)
	assert.Equal(t, 2, size)

	ng.CurrentSize = 4
	size, err = ng.TargetSize()
	assert.NoError(t, err)
	assert.Equal(t, 4, size)
}

func TestNodeGroup_IncreaseSize(t *testing.T) {
	ng := newTestNodeGroup(t)
	client := ng.Manager.Client.(*sdk.ClientMock)
	ctx := context.Background()

	t.Run("rejects non-positive delta", func(t *testing.T) {
		assert.Error(t, ng.IncreaseSize(0))
	})

	t.Run("rejects above max", func(t *testing.T) {
		assert.Error(t, ng.IncreaseSize(10))
	})

	t.Run("adds nodes and updates current size", func(t *testing.T) {
		client.On("AddNode", ctx, "clusterID", "pool-1").Return(&sdk.Node{Id: "n1"}, nil).Once()
		client.On("AddNode", ctx, "clusterID", "pool-1").Return(&sdk.Node{Id: "n2"}, nil).Once()

		err := ng.IncreaseSize(2)
		assert.NoError(t, err)
		assert.Equal(t, 4, ng.CurrentSize)
	})
}

func TestNodeGroup_DeleteNodes(t *testing.T) {
	ng := newTestNodeGroup(t)
	ng.CurrentSize = 3
	client := ng.Manager.Client.(*sdk.ClientMock)
	ctx := context.Background()

	nodes := []*apiv1.Node{
		{ObjectMeta: metav1.ObjectMeta{Name: "node-1"}, Spec: apiv1.NodeSpec{ProviderID: "openstack:///instance-1"}},
	}

	t.Run("rejects below min", func(t *testing.T) {
		ng.CurrentSize = 1
		err := ng.DeleteNodes([]*apiv1.Node{
			{Spec: apiv1.NodeSpec{ProviderID: "openstack:///instance-1"}},
		})
		assert.Error(t, err)
		ng.CurrentSize = 3
	})

	t.Run("deletes node and updates current size", func(t *testing.T) {
		client.On("DeleteNode", ctx, "clusterID", "pool-1", "instance-1").Return(nil).Once()
		err := ng.DeleteNodes(nodes)
		assert.NoError(t, err)
		assert.Equal(t, 2, ng.CurrentSize)
	})
}

func TestNodeGroup_Nodes(t *testing.T) {
	ng := newTestNodeGroup(t)
	client := ng.Manager.Client.(*sdk.ClientMock)
	ctx := context.Background()

	client.On("ListNodePoolNodes", ctx, "clusterID", "pool-1").Return(
		[]sdk.Node{
			{Id: "instance-1", Status: "READY"},
			{Id: "openstack:///instance-2", Status: "INSTALLING"},
		}, nil,
	)

	instances, err := ng.Nodes()
	assert.NoError(t, err)
	assert.Len(t, instances, 2)
	assert.Equal(t, "openstack:///instance-1", instances[0].Id)
	assert.Equal(t, cloudprovider.InstanceRunning, instances[0].Status.State)
	assert.Equal(t, "openstack:///instance-2", instances[1].Id)
	assert.Equal(t, cloudprovider.InstanceCreating, instances[1].Status.State)
	assert.Equal(t, ng, ng.Manager.getNodeGroupPerProviderID("openstack:///instance-1"))
}

func TestNodeGroup_TemplateNodeInfo(t *testing.T) {
	ng := newTestNodeGroup(t)
	info, err := ng.TemplateNodeInfo()
	assert.NoError(t, err)
	assert.NotNil(t, info.Node())
	assert.Equal(t, "workers", info.Node().Labels[NodePoolLabel])
}

func TestNodeGroup_OptionalMethods(t *testing.T) {
	ng := newTestNodeGroup(t)
	assert.True(t, ng.Exist())
	assert.False(t, ng.Autoprovisioned())
	assert.Equal(t, cloudprovider.ErrNotImplemented, ng.DecreaseTargetSize(-1))

	_, err := ng.Create()
	assert.Equal(t, cloudprovider.ErrNotImplemented, err)
	assert.Equal(t, cloudprovider.ErrNotImplemented, ng.Delete())
}

func TestToInstanceStatus(t *testing.T) {
	assert.Equal(t, cloudprovider.InstanceCreating, toInstanceStatus("INSTALLING").State)
	assert.Equal(t, cloudprovider.InstanceDeleting, toInstanceStatus("DELETING").State)
	assert.Equal(t, cloudprovider.InstanceRunning, toInstanceStatus("READY").State)
	assert.NotNil(t, toInstanceStatus("UNKNOWN").ErrorInfo)
}
