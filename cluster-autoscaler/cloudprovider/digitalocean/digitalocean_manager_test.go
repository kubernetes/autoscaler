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
	"fmt"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/digitalocean/godo"
)

func TestNewManager(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		cfg := `{"cluster_id": "123456", "token": "123-123-123", "url": "https://api.digitalocean.com/v2", "version": "dev"}`

		manager, err := newManager(bytes.NewBufferString(cfg), nil)
		assert.NoError(t, err)
		assert.Equal(t, manager.clusterID, "123456", "cluster ID does not match")
	})

	t.Run("empty token", func(t *testing.T) {
		cfg := `{"cluster_id": "123456", "token": "", "url": "https://api.digitalocean.com/v2", "version": "dev"}`

		_, err := newManager(bytes.NewBufferString(cfg), nil)
		assert.EqualError(t, err, errors.New("access token is not provided").Error())
	})

	t.Run("empty cluster ID", func(t *testing.T) {
		cfg := `{"cluster_id": "", "token": "123-123-123", "url": "https://api.digitalocean.com/v2", "version": "dev"}`

		_, err := newManager(bytes.NewBufferString(cfg), nil)
		assert.EqualError(t, err, errors.New("cluster ID is not provided").Error())
	})
}

func TestDigitalOceanManager_Refresh(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		cfg := `{"cluster_id": "123456", "token": "123-123-123", "url": "https://api.digitalocean.com/v2", "version": "dev"}`

		manager, err := newManager(bytes.NewBufferString(cfg), nil)
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

func TestDigitalOceanManager_RefreshWithNodeSpec(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		cfg := `{"cluster_id": "123456", "token": "123-123-123", "url": "https://api.digitalocean.com/v2", "version": "dev"}`

		specs := []string{
			"3,10,bar", // first group
			"5,20,foo", // second group
		}

		manager, err := newManager(bytes.NewBufferString(cfg), specs)
		assert.NoError(t, err)

		client := &doClientMock{}
		ctx := context.Background()

		nodes1 := []*godo.KubernetesNode{
			{ID: "1", Status: &godo.KubernetesNodeStatus{State: "running"}},
			{ID: "2", Status: &godo.KubernetesNodeStatus{State: "running"}},
		}
		nodes2 := []*godo.KubernetesNode{
			{ID: "3", Status: &godo.KubernetesNodeStatus{State: "running"}},
			{ID: "4", Status: &godo.KubernetesNodeStatus{State: "running"}},
		}

		nodes3 := []*godo.KubernetesNode{
			{ID: "5", Status: &godo.KubernetesNodeStatus{State: "running"}},
			{ID: "6", Status: &godo.KubernetesNodeStatus{State: "running"}},
		}

		client.On("ListNodePools", ctx, manager.clusterID, nil).Return(
			[]*godo.KubernetesNodePool{
				{ID: "1"},
				{ID: "2"},
				{ID: "3"},
			},
			&godo.Response{},
			nil,
		).Once()

		client.On("GetNodePool", ctx, manager.clusterID, "1").Return(
			&godo.KubernetesNodePool{
				Nodes: nodes1,
				Tags:  []string{"k8s-cluster-autoscaler:bar"},
			},
			&godo.Response{},
			nil,
		).Once()
		client.On("GetNodePool", ctx, manager.clusterID, "2").Return(
			&godo.KubernetesNodePool{
				Nodes: nodes2,
				Tags:  []string{"k8s-cluster-autoscaler:foo"},
			},
			&godo.Response{},
			nil,
		).Once()

		// this node pool doesn't have any tags, therefore this should get
		// assigned the default minimum and maximum defaults
		client.On("GetNodePool", ctx, manager.clusterID, "3").Return(
			&godo.KubernetesNodePool{
				Nodes: nodes3,
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

func Test_parseNodeSpec(t *testing.T) {
	cases := []struct {
		name    string
		specs   []string
		want    []*nodeSpec
		wantErr bool
	}{
		{
			name: "good spec (single)",
			specs: []string{
				"3,10,foo",
			},
			want: []*nodeSpec{
				{minSize: 3, maxSize: 10, tagValue: "foo"},
			},
		},
		{
			name: "good specs (multiple)",
			specs: []string{
				"5,20,bar",
				"3,10,foo",
			},
			want: []*nodeSpec{
				{minSize: 5, maxSize: 20, tagValue: "bar"},
				{minSize: 3, maxSize: 10, tagValue: "foo"},
			},
		},
		{
			name: "bad specs - no tag value",
			specs: []string{
				"5,20",
				"3,10,foo",
			},
			wantErr: true,
		},
		{
			name: "bad specs - no numerical min node size",
			specs: []string{
				"five,20,bar",
				"3,10,foo",
			},
			wantErr: true,
		},
		{
			name: "bad specs - no numerical max node size",
			specs: []string{
				"5,twenty,bar",
				"3,10,foo",
			},
			wantErr: true,
		},
		{
			name: "bad specs - empty",
			specs: []string{
				"",
			},
			wantErr: true,
		},
		{
			name: "bad specs - min is higher than max",
			specs: []string{
				"5,4,bar",
				"3,10,foo",
			},
			wantErr: true,
		},
		{
			name: "bad specs - max is set to zero",
			specs: []string{
				"5,20,bar",
				"3,0,foo",
			},
			wantErr: true,
		},
	}

	for _, ts := range cases {
		ts := ts

		t.Run(ts.name, func(t *testing.T) {
			nodeSpecs, err := parseNodeSpec(ts.specs)
			fmt.Printf("err = %+v\n", err)
			if ts.wantErr && err == nil {
				assert.Error(t, err)
				return
			}

			if ts.wantErr {
				return
			}

			assert.NoError(t, err)

			var got []*nodeSpec
			for _, spec := range nodeSpecs {
				got = append(got, spec)
			}

			sort.Slice(got, func(i, j int) bool {
				return got[i].tagValue < got[j].tagValue
			})

			assert.Equal(t, ts.want, got, "\ngot: %#v\nwant: %#v", got, ts.want)
		})
	}
}
