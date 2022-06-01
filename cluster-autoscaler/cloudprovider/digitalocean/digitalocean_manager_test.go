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
	"errors"
	"testing"

	"github.com/digitalocean/godo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewManager(t *testing.T) {
	t.Run("success with literal token", func(t *testing.T) {
		cfg := `{"cluster_id": "123456", "token": "123-123-123", "url": "https://api.digitalocean.com/v2", "version": "dev"}`

		manager, err := newManager(bytes.NewBufferString(cfg))
		require.NoError(t, err)
		assert.Equal(t, manager.clusterID, "123456", "cluster ID does not match")
	})
	t.Run("success with token file", func(t *testing.T) {
		cfg := `{"cluster_id": "123456", "token_file": "testdata/correct_token", "url": "https://api.digitalocean.com/v2", "version": "dev"}`

		manager, err := newManager(bytes.NewBufferString(cfg))
		require.NoError(t, err)
		assert.Equal(t, manager.clusterID, "123456", "cluster ID does not match")
	})

	t.Run("empty token", func(t *testing.T) {
		cfg := `{"cluster_id": "123456", "token": "", "url": "https://api.digitalocean.com/v2", "version": "dev"}`

		_, err := newManager(bytes.NewBufferString(cfg))
		assert.EqualError(t, err, errors.New("access token is not provided").Error())
	})
	t.Run("literal and token file", func(t *testing.T) {
		cfg := `{"cluster_id": "123456", "token": "123-123-123", "token_file": "tokendata/correct_token", "url": "https://api.digitalocean.com/v2", "version": "dev"}`

		_, err := newManager(bytes.NewBufferString(cfg))
		assert.EqualError(t, err, errors.New("access token literal and access token file must not be provided together").Error())
	})
	t.Run("missing token file", func(t *testing.T) {
		cfg := `{"cluster_id": "123456", "token_file": "testdata/missing_token", "url": "https://api.digitalocean.com/v2", "version": "dev"}`

		_, err := newManager(bytes.NewBufferString(cfg))
		require.NotNil(t, err)
		assert.Contains(t, err.Error(), "failed to read token file")
	})
	t.Run("empty token file", func(t *testing.T) {
		cfg := `{"cluster_id": "123456", "token_file": "testdata/empty_token", "url": "https://api.digitalocean.com/v2", "version": "dev"}`

		_, err := newManager(bytes.NewBufferString(cfg))
		assert.EqualError(t, err, errors.New(`token file "testdata/empty_token" is empty`).Error())
	})
	t.Run("all whitespace token file", func(t *testing.T) {
		cfg := `{"cluster_id": "123456", "token_file": "testdata/whitespace_token", "url": "https://api.digitalocean.com/v2", "version": "dev"}`

		_, err := newManager(bytes.NewBufferString(cfg))
		assert.EqualError(t, err, errors.New(`token file "testdata/whitespace_token" is empty`).Error())
	})
	t.Run("empty cluster ID", func(t *testing.T) {
		cfg := `{"cluster_id": "", "token": "123-123-123", "url": "https://api.digitalocean.com/v2", "version": "dev"}`

		_, err := newManager(bytes.NewBufferString(cfg))
		assert.EqualError(t, err, errors.New("cluster ID is not provided").Error())
	})
}

func TestDigitalOceanManager_Refresh(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		cfg := `{"cluster_id": "123456", "token": "123-123-123", "url": "https://api.digitalocean.com/v2", "version": "dev"}`

		manager, err := newManager(bytes.NewBufferString(cfg))
		assert.NoError(t, err)

		client := &doClientMock{}
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
					AutoScale: true,
				},
			},
			&godo.Response{},
			nil,
		).Once()

		manager.client = client
		err = manager.Refresh()
		assert.NoError(t, err)
		assert.Equal(t, len(manager.nodeGroups), 4, "number of nodes do not match")
	})

}

func TestDigitalOceanManager_RefreshWithNodeSpec(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		cfg := `{"cluster_id": "123456", "token": "123-123-123", "url": "https://api.digitalocean.com/v2", "version": "dev"}`

		manager, err := newManager(bytes.NewBufferString(cfg))
		assert.NoError(t, err)

		client := &doClientMock{}
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
					MinNodes:  3,
					MaxNodes:  10,
				},
				{
					ID: "2",
					Nodes: []*godo.KubernetesNode{
						{ID: "3", Status: &godo.KubernetesNodeStatus{State: "running"}},
						{ID: "4", Status: &godo.KubernetesNodeStatus{State: "running"}},
					},
					AutoScale: true,
					MinNodes:  5,
					MaxNodes:  20,
				},
				{
					// this node pool doesn't have autoscale config, therefore
					// this should default to disabled auto-scale.
					ID: "3",
					Nodes: []*godo.KubernetesNode{
						{ID: "5", Status: &godo.KubernetesNodeStatus{State: "running"}},
						{ID: "6", Status: &godo.KubernetesNodeStatus{State: "running"}},
					},
				},
			},
			&godo.Response{},
			nil,
		).Once()

		manager.client = client
		err = manager.Refresh()
		assert.NoError(t, err)
		assert.Equal(t, len(manager.nodeGroups), 2, "number of node groups do not match")

		// first node group
		assert.Equal(t, manager.nodeGroups[0].minSize, 3, "minimum node for first group does not match")
		assert.Equal(t, manager.nodeGroups[0].maxSize, 10, "maximum node for first group does not match")

		// second node group
		assert.Equal(t, manager.nodeGroups[1].minSize, 5, "minimum node for second group does not match")
		assert.Equal(t, manager.nodeGroups[1].maxSize, 20, "maximum node for second group does not match")
	})
}
