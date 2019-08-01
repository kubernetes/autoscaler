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

	"github.com/stretchr/testify/assert"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/digitalocean/godo"
)

func TestNewManager(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		cfg := `{"cluster_id": "123456", "token": "123-123-123", "url": "https://api.digitalocean.com/v2", "version": "dev"}`

		manager, err := newManager(bytes.NewBufferString(cfg))
		assert.NoError(t, err)
		assert.Equal(t, manager.clusterID, "123456", "cluster ID does not match")
	})

	t.Run("empty token", func(t *testing.T) {
		cfg := `{"cluster_id": "123456", "token": "", "url": "https://api.digitalocean.com/v2", "version": "dev"}`

		_, err := newManager(bytes.NewBufferString(cfg))
		assert.EqualError(t, err, errors.New("access token is not provided").Error())
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
				{ID: "1"},
				{ID: "2"},
				{ID: "3"},
				{ID: "4"},
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

		manager.client = client
		err = manager.Refresh()
		assert.NoError(t, err)
		assert.Equal(t, len(manager.nodeGroups), 4, "number of nodes do not match")
	})

}
