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

		client.On("ListNodePools", ctx, manager.clusterID, nil).Return(
			[]*godo.KubernetesNodePool{
				{
					ID: "1",
					Nodes: []*godo.KubernetesNode{
						{ID: "1", Status: &godo.KubernetesNodeStatus{State: "running"}},
						{ID: "2", Status: &godo.KubernetesNodeStatus{State: "running"}},
					},
					Tags: []string{
						"k8s-cluster-autoscaler-enabled:true",
					},
				},
				{
					ID: "2",
					Nodes: []*godo.KubernetesNode{
						{ID: "3", Status: &godo.KubernetesNodeStatus{State: "deleting"}},
						{ID: "4", Status: &godo.KubernetesNodeStatus{State: "running"}},
					},
					Tags: []string{
						"k8s-cluster-autoscaler-enabled:true",
					},
				},
				{
					ID: "3",
					Nodes: []*godo.KubernetesNode{
						{ID: "5", Status: &godo.KubernetesNodeStatus{State: "provisioning"}},
						{ID: "6", Status: &godo.KubernetesNodeStatus{State: "running"}},
					},
					Tags: []string{
						"k8s-cluster-autoscaler-enabled:true",
					},
				},
				{
					ID: "4",
					Nodes: []*godo.KubernetesNode{
						{ID: "7", Status: &godo.KubernetesNodeStatus{State: "draining"}},
						{ID: "8", Status: &godo.KubernetesNodeStatus{State: "running"}},
					},
					Tags: []string{
						"k8s-cluster-autoscaler-enabled:true",
					},
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
					Tags: []string{
						"k8s-cluster-autoscaler-enabled:true",
						"k8s-cluster-autoscaler-min:3",
						"k8s-cluster-autoscaler-max:10",
					},
				},
				{
					ID: "2",
					Nodes: []*godo.KubernetesNode{
						{ID: "3", Status: &godo.KubernetesNodeStatus{State: "running"}},
						{ID: "4", Status: &godo.KubernetesNodeStatus{State: "running"}},
					},
					Tags: []string{
						"k8s-cluster-autoscaler-enabled:true",
						"k8s-cluster-autoscaler-min:5",
						"k8s-cluster-autoscaler-max:20",
					},
				},
				{
					// this node pool doesn't have any min and max tags,
					// therefore this should get assigned the default minimum
					// and maximum defaults
					ID: "3",
					Nodes: []*godo.KubernetesNode{
						{ID: "5", Status: &godo.KubernetesNodeStatus{State: "running"}},
						{ID: "6", Status: &godo.KubernetesNodeStatus{State: "running"}},
					},
					Tags: []string{
						"k8s-cluster-autoscaler-enabled:true",
					},
				},
			},
			&godo.Response{},
			nil,
		).Once()

		manager.client = client
		err = manager.Refresh()
		assert.NoError(t, err)
		assert.Equal(t, len(manager.nodeGroups), 3, "number of nodes do not match")

		// first node group
		assert.Equal(t, manager.nodeGroups[0].minSize, 3, "minimum node for first group does not match")
		assert.Equal(t, manager.nodeGroups[0].maxSize, 10, "maximum node for first group does not match")

		// second node group
		assert.Equal(t, manager.nodeGroups[1].minSize, 5, "minimum node for second group does not match")
		assert.Equal(t, manager.nodeGroups[1].maxSize, 20, "maximum node for second group does not match")

		// third node group
		assert.Equal(t, manager.nodeGroups[2].minSize, minNodePoolSize, "minimum node for third group should match the default")
		assert.Equal(t, manager.nodeGroups[2].maxSize, maxNodePoolSize, "maximum node for third group should match the default")
	})
}

func Test_parseTags(t *testing.T) {
	cases := []struct {
		name    string
		tags    []string
		want    *nodeSpec
		wantErr bool
	}{
		{
			name: "good config (single)",
			tags: []string{
				"k8s-cluster-autoscaler-enabled:true",
				"k8s-cluster-autoscaler-min:3",
				"k8s-cluster-autoscaler-max:10",
			},
			want: &nodeSpec{
				min:     3,
				max:     10,
				enabled: true,
			},
		},
		{
			name: "good config (disabled)",
			tags: []string{
				"k8s-cluster-autoscaler-min:3",
				"k8s-cluster-autoscaler-max:10",
			},
			want: &nodeSpec{
				min: 3,
				max: 10,
			},
		},
		{
			name: "good config (disabled with no values)",
			tags: []string{},
			want: &nodeSpec{},
		},
		{
			name: "good config - empty tags",
			tags: []string{""},
			want: &nodeSpec{},
		},
		{
			name: "bad tags - malformed",
			tags: []string{
				"k8s-cluster-autoscaler-enabled:true",
				"k8s-cluster-autoscaler-min=3",
				"k8s-cluster-autoscaler-max=10",
			},
			wantErr: true,
		},
		{
			name: "bad tags - no numerical min node size",
			tags: []string{
				"k8s-cluster-autoscaler-enabled:true",
				"k8s-cluster-autoscaler-min:three",
				"k8s-cluster-autoscaler-max:10",
			},
			wantErr: true,
		},
		{
			name: "bad tags - no numerical max node size",
			tags: []string{
				"k8s-cluster-autoscaler-enabled:true",
				"k8s-cluster-autoscaler-min:3",
				"k8s-cluster-autoscaler-max:ten",
			},
			wantErr: true,
		},
		{
			name: "bad tags - min is higher than max",
			tags: []string{
				"k8s-cluster-autoscaler-enabled:true",
				"k8s-cluster-autoscaler-min:5",
				"k8s-cluster-autoscaler-max:4",
			},
			wantErr: true,
		},
		{
			name: "bad tags - max is set to zero",
			tags: []string{
				"k8s-cluster-autoscaler-enabled:true",
				"k8s-cluster-autoscaler-min:5",
				"k8s-cluster-autoscaler-max:0",
			},
			wantErr: true,
		},
		{
			name: "bad tags - max is set to negative, no min",
			tags: []string{
				"k8s-cluster-autoscaler-enabled:true",
				"k8s-cluster-autoscaler-max:-5",
			},
			wantErr: true,
		},
		{
			// TODO(arslan): remove this once we support zero count node pools on our end
			name: "bad tags - min is set to zero",
			tags: []string{
				"k8s-cluster-autoscaler-enabled:true",
				"k8s-cluster-autoscaler-min:0",
				"k8s-cluster-autoscaler-max:5",
			},
			wantErr: true,
		},
	}

	for _, ts := range cases {
		ts := ts

		t.Run(ts.name, func(t *testing.T) {
			got, err := parseTags(ts.tags)
			if ts.wantErr && err == nil {
				assert.Error(t, err)
				return
			}

			if ts.wantErr {
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, ts.want, got, "\ngot: %#v\nwant: %#v", got, ts.want)
		})
	}

}
