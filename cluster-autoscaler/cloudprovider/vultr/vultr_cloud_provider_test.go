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

package vultr

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/vultr/govultr"
)

func TestVultrCloudProvider_newVultrCloudProvider(t *testing.T) {
	config := `{"token": "123-456", "cluster_id": "abc"}`

	manager, err := newManager(strings.NewReader(config))
	require.NoError(t, err)

	client := &vultrClientMock{}
	ctx := context.Background()

	client.On("ListNodePools", ctx, manager.clusterID, nil).Return(
		[]govultr.NodePool{
			{
				ID:         "1234",
				AutoScaler: true,
				MinNodes:   1,
				MaxNodes:   2,
			},
			{
				ID:         "4567",
				AutoScaler: true,
				MinNodes:   5,
				MaxNodes:   8,
			},
			{
				ID:         "9876",
				AutoScaler: false,
				MinNodes:   5,
				MaxNodes:   8,
			},
		},
		&govultr.Meta{},
		nil,
	).Once()

	manager.client = client
	rl := &cloudprovider.ResourceLimiter{}

	_ = newVultrCloudProvider(manager, rl)

}

func TestVultrCloudProvider_NewNodeGroup(t *testing.T) {
	config := `{"token": "123-456", "cluster_id": "abc"}`

	manager, err := newManager(strings.NewReader(config))
	require.NoError(t, err)

	client := &vultrClientMock{}
	ctx := context.Background()

	client.On("ListNodePools", ctx, manager.clusterID, nil).Return(
		[]govultr.NodePool{
			{
				ID:         "1234",
				AutoScaler: true,
				MinNodes:   1,
				MaxNodes:   2,
			},
			{
				ID:         "4567",
				AutoScaler: true,
				MinNodes:   5,
				MaxNodes:   8,
			},
			{
				ID:         "9876",
				AutoScaler: false,
				MinNodes:   5,
				MaxNodes:   8,
			},
		},
		&govultr.Meta{},
		nil,
	).Once()

	manager.client = client
	rl := &cloudprovider.ResourceLimiter{}

	provider := newVultrCloudProvider(manager, rl)
	err = provider.Refresh()
	assert.NoError(t, err)

	nodes := provider.NodeGroups()
	assert.Equal(t, len(nodes), 2, "number of nodes do not match")

}

func TestVultrCloudProvider_NodeGroupForNode(t *testing.T) {
	config := `{"token": "123-456", "cluster_id": "abc"}`

	manager, err := newManager(strings.NewReader(config))
	require.NoError(t, err)

	client := &vultrClientMock{}
	ctx := context.Background()

	client.On("ListNodePools", ctx, manager.clusterID, nil).Return(
		[]govultr.NodePool{
			{
				ID:         "a",
				AutoScaler: true,
				Nodes: []govultr.Node{
					{
						ID:     "np-1234",
						Status: "Active",
					},
				},
				MinNodes: 1,
				MaxNodes: 2,
			},
			{
				ID:         "b",
				AutoScaler: true,
				Nodes: []govultr.Node{
					{
						ID:     "np-456",
						Status: "Active",
					},
				},
				MinNodes: 5,
				MaxNodes: 8,
			},
			{
				ID:         "c",
				AutoScaler: false,
				MinNodes:   5,
				MaxNodes:   8,
			},
		},
		&govultr.Meta{},
		nil,
	).Once()

	manager.client = client
	rl := &cloudprovider.ResourceLimiter{}

	provider := newVultrCloudProvider(manager, rl)
	err = provider.Refresh()
	assert.NoError(t, err)

	node := &apiv1.Node{Spec: apiv1.NodeSpec{ProviderID: toProviderID("np-1234")}}

	nodeGroup, err := provider.NodeGroupForNode(node)
	require.NoError(t, err)

	require.NotNil(t, nodeGroup)
	require.Equal(t, nodeGroup.Id(), "a", "nodegroup IDs do not match")
}

func TestVultrCloudProvider_Name(t *testing.T) {
	config := `{"token": "123-456", "cluster_id": "abc"}`

	manager, err := newManager(strings.NewReader(config))
	require.NoError(t, err)

	p := newVultrCloudProvider(manager, &cloudprovider.ResourceLimiter{})
	assert.Equal(t, cloudprovider.VultrProviderName, p.Name(), "provider name doesn't match")
}
