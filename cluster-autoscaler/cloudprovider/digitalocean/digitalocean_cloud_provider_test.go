/*
Copyright 2019 The Kubernetes Authors.

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

package digitalocean

import (
	"bytes"
	"context"
	"testing"

	"github.com/digitalocean/godo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
)

func testCloudProvider(t *testing.T, client *doClientMock) *digitaloceanCloudProvider {
	cfg := `{"cluster_id": "123456", "token": "123-123-123", "url": "https://api.digitalocean.com/v2", "version": "dev"}`

	manager, err := newManager(bytes.NewBufferString(cfg))
	assert.NoError(t, err)
	rl := &cloudprovider.ResourceLimiter{}

	// fill the test provider with some example
	if client == nil {
		client = &doClientMock{}
		ctx := context.Background()

		client.On("ListNodePools", ctx, manager.clusterID, nil).Return(
			[]*godo.KubernetesNodePool{
				{
					ID: "1",
					Nodes: []*godo.KubernetesNode{
						{ID: "1", Status: &godo.KubernetesNodeStatus{State: "running"}},
						{ID: "2", Status: &godo.KubernetesNodeStatus{State: "running"}},
					},
					AutoScale: true,
				},
				{
					ID: "2",
					Nodes: []*godo.KubernetesNode{
						{ID: "3", Status: &godo.KubernetesNodeStatus{State: "deleting"}},
						{ID: "4", Status: &godo.KubernetesNodeStatus{State: "running"}},
					},
					AutoScale: true,
				},
				{
					ID: "3",
					Nodes: []*godo.KubernetesNode{
						{ID: "5", Status: &godo.KubernetesNodeStatus{State: "provisioning"}},
						{ID: "6", Status: &godo.KubernetesNodeStatus{State: "running"}},
					},
					AutoScale: true,
				},
				{
					ID: "4",
					Nodes: []*godo.KubernetesNode{
						{ID: "7", Status: &godo.KubernetesNodeStatus{State: "draining"}},
						{ID: "8", Status: &godo.KubernetesNodeStatus{State: "running"}},
					},
					AutoScale: false,
				},
			},
			&godo.Response{},
			nil,
		).Once()
	}

	manager.client = client

	return newDigitalOceanCloudProvider(manager, rl)
}

func TestNewDigitalOceanCloudProvider(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		_ = testCloudProvider(t, nil)
	})
}

func TestDigitalOceanCloudProvider_Name(t *testing.T) {
	provider := testCloudProvider(t, nil)

	t.Run("success", func(t *testing.T) {
		name := provider.Name()
		assert.Equal(t, cloudprovider.DigitalOceanProviderName, name, "provider name doesn't match")
	})
}

func TestDigitalOceanCloudProvider_NodeGroups(t *testing.T) {
	provider := testCloudProvider(t, nil)
	err := provider.manager.Refresh()
	assert.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		nodes := provider.NodeGroups()
		assert.Equal(t, len(nodes), 3, "number of nodes do not match")
	})

	t.Run("zero groups", func(t *testing.T) {
		provider.manager.nodeGroups = []*NodeGroup{}
		nodes := provider.NodeGroups()
		assert.Equal(t, len(nodes), 0, "number of nodes do not match")
	})
}

func TestDigitalOceanCloudProvider_NodeGroupForNode(t *testing.T) {
	clusterID := "123456"

	t.Run("success", func(t *testing.T) {
		client := &doClientMock{}
		ctx := context.Background()

		client.On("ListNodePools", ctx, clusterID, nil).Return(
			[]*godo.KubernetesNodePool{
				{
					ID: "1",
					Nodes: []*godo.KubernetesNode{
						{ID: "2", Status: &godo.KubernetesNodeStatus{State: "deleting"}, DropletID: "droplet-2"},
						{ID: "3", Status: &godo.KubernetesNodeStatus{State: "running"}, DropletID: "droplet-3"},
					},
					AutoScale: true,
				},
				{
					ID: "2",
					Nodes: []*godo.KubernetesNode{
						{ID: "4", Status: &godo.KubernetesNodeStatus{State: "provisioning"}, DropletID: "droplet-4"},
						{ID: "5", Status: &godo.KubernetesNodeStatus{State: "draining"}, DropletID: "droplet-5"},
					},
					AutoScale: true,
				},
			},
			&godo.Response{},
			nil,
		).Once()

		provider := testCloudProvider(t, client)
		err := provider.manager.Refresh()
		assert.NoError(t, err)

		// let's get the nodeGroup for the node with ID 4
		node := &apiv1.Node{
			Spec: apiv1.NodeSpec{
				ProviderID: toProviderID("droplet-4"),
			},
		}

		nodeGroup, err := provider.NodeGroupForNode(node)
		require.NoError(t, err)
		require.NotNil(t, nodeGroup)
		require.Equal(t, nodeGroup.Id(), "2", "node group ID does not match")
	})

	t.Run("node does not exist", func(t *testing.T) {
		client := &doClientMock{}
		ctx := context.Background()

		client.On("ListNodePools", ctx, clusterID, nil).Return(
			[]*godo.KubernetesNodePool{
				{
					ID: "1",
					Nodes: []*godo.KubernetesNode{
						{ID: "2", Status: &godo.KubernetesNodeStatus{State: "deleting"}, DropletID: "droplet-2"},
						{ID: "3", Status: &godo.KubernetesNodeStatus{State: "running"}, DropletID: "droplet-3"},
					},
				},
				{
					ID: "2",
					Nodes: []*godo.KubernetesNode{
						{ID: "4", Status: &godo.KubernetesNodeStatus{State: "provisioning"}, DropletID: "droplet-4"},
						{ID: "5", Status: &godo.KubernetesNodeStatus{State: "draining"}, DropletID: "droplet-5"},
					},
				},
			},
			&godo.Response{},
			nil,
		).Once()

		provider := testCloudProvider(t, client)
		err := provider.manager.Refresh()
		assert.NoError(t, err)

		node := &apiv1.Node{
			Spec: apiv1.NodeSpec{
				ProviderID: toProviderID("droplet-7"),
			},
		}

		nodeGroup, err := provider.NodeGroupForNode(node)
		assert.NoError(t, err)
		assert.Nil(t, nodeGroup)
	})
}
