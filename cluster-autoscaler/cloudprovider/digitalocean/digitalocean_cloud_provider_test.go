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

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/digitalocean/godo"
)

func testCloudProvider(t *testing.T, client *doClientMock) *digitaloceanCloudProvider {
	cfg := `{"cluster_id": "123456", "token": "123-123-123", "url": "https://api.digitalocean.com/v2", "version": "dev"}`

	manager, err := newManager(bytes.NewBufferString(cfg), nil)
	assert.NoError(t, err)
	rl := &cloudprovider.ResourceLimiter{}

	// fill the test provider with some example
	if client == nil {
		client = &doClientMock{}
		ctx := context.Background()

		nodes1 := []*godo.KubernetesNode{
			{ID: "1", Status: &godo.KubernetesNodeStatus{State: "running"}},
			{ID: "2", Status: &godo.KubernetesNodeStatus{State: "running"}},
		}
		nodes2 := []*godo.KubernetesNode{
			{ID: "3", Status: &godo.KubernetesNodeStatus{State: "deleting"}},
			{ID: "4", Status: &godo.KubernetesNodeStatus{State: "running"}},
		}
		nodes3 := []*godo.KubernetesNode{
			{ID: "5", Status: &godo.KubernetesNodeStatus{State: "provisioning"}},
			{ID: "6", Status: &godo.KubernetesNodeStatus{State: "running"}},
		}
		nodes4 := []*godo.KubernetesNode{
			{ID: "7", Status: &godo.KubernetesNodeStatus{State: "draining"}},
			{ID: "8", Status: &godo.KubernetesNodeStatus{State: "running"}},
		}

		client.On("ListNodePools", ctx, manager.clusterID, nil).Return(
			[]*godo.KubernetesNodePool{
				{ID: "1", Nodes: nodes1},
				{ID: "2", Nodes: nodes2},
				{ID: "3", Nodes: nodes3},
				{ID: "4", Nodes: nodes4},
			},
			&godo.Response{},
			nil,
		).Once()

		client.On("GetNodePool", ctx, manager.clusterID, "1").Return(
			&godo.KubernetesNodePool{Nodes: nodes1},
			&godo.Response{},
			nil,
		).Once()
		client.On("GetNodePool", ctx, manager.clusterID, "2").Return(
			&godo.KubernetesNodePool{Nodes: nodes2},
			&godo.Response{},
			nil,
		).Once()

		client.On("GetNodePool", ctx, manager.clusterID, "3").Return(
			&godo.KubernetesNodePool{Nodes: nodes3},
			&godo.Response{},
			nil,
		).Once()
		client.On("GetNodePool", ctx, manager.clusterID, "4").Return(
			&godo.KubernetesNodePool{Nodes: nodes4},
			&godo.Response{},
			nil,
		).Once()
	}

	manager.client = client

	provider, err := newDigitalOceanCloudProvider(manager, rl)
	assert.NoError(t, err)
	return provider

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

	t.Run("success", func(t *testing.T) {
		nodes := provider.NodeGroups()
		assert.Equal(t, len(nodes), 4, "number of nodes do not match")
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

		nodes1 := []*godo.KubernetesNode{
			{ID: "2", Status: &godo.KubernetesNodeStatus{State: "deleting"}},
			{ID: "3", Status: &godo.KubernetesNodeStatus{State: "running"}},
		}
		nodes2 := []*godo.KubernetesNode{
			{ID: "4", Status: &godo.KubernetesNodeStatus{State: "provisioning"}},
			{ID: "5", Status: &godo.KubernetesNodeStatus{State: "draining"}},
		}

		client.On("ListNodePools", ctx, clusterID, nil).Return(
			[]*godo.KubernetesNodePool{
				{ID: "1", Nodes: nodes1},
				{ID: "2", Nodes: nodes2},
			},
			&godo.Response{},
			nil,
		).Once()

		client.On("GetNodePool", ctx, clusterID, "1").Return(
			&godo.KubernetesNodePool{Nodes: nodes1},
			&godo.Response{},
			nil,
		).Once()
		client.On("GetNodePool", ctx, clusterID, "2").Return(
			&godo.KubernetesNodePool{Nodes: nodes2},
			&godo.Response{},
			nil,
		).Once()

		provider := testCloudProvider(t, client)

		// let's get the nodeGroup for the node with ID 4
		node := &apiv1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					nodeIDLabel: "4",
				},
			},
		}

		nodeGroup, err := provider.NodeGroupForNode(node)
		assert.NoError(t, err)
		assert.NotNil(t, nodeGroup)
		assert.Equal(t, nodeGroup.Id(), "2", "node group ID does not match")
	})

	t.Run("node does not exist", func(t *testing.T) {
		client := &doClientMock{}
		ctx := context.Background()

		nodes1 := []*godo.KubernetesNode{
			{ID: "2", Status: &godo.KubernetesNodeStatus{State: "deleting"}},
			{ID: "3", Status: &godo.KubernetesNodeStatus{State: "running"}},
		}
		nodes2 := []*godo.KubernetesNode{
			{ID: "4", Status: &godo.KubernetesNodeStatus{State: "provisioning"}},
			{ID: "5", Status: &godo.KubernetesNodeStatus{State: "draining"}},
		}

		client.On("ListNodePools", ctx, clusterID, nil).Return(
			[]*godo.KubernetesNodePool{
				{ID: "1", Nodes: nodes1},
				{ID: "2", Nodes: nodes2},
			},
			&godo.Response{},
			nil,
		).Once()

		client.On("GetNodePool", ctx, clusterID, "1").Return(
			&godo.KubernetesNodePool{Nodes: nodes1},
			&godo.Response{},
			nil,
		).Once()
		client.On("GetNodePool", ctx, clusterID, "2").Return(
			&godo.KubernetesNodePool{Nodes: nodes2},
			&godo.Response{},
			nil,
		).Once()

		provider := testCloudProvider(t, client)

		node := &apiv1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					nodeIDLabel: "7",
				},
			},
		}

		nodeGroup, err := provider.NodeGroupForNode(node)
		assert.NoError(t, err)
		assert.Nil(t, nodeGroup)
	})
}
