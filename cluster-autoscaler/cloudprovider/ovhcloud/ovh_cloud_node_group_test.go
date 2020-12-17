/*
Copyright 2020 The Kubernetes Authors.

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

package ovhcloud

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/ovhcloud/sdk"
)

func newTestNodeGroup(t *testing.T) cloudprovider.NodeGroup {
	cfg := `{
		"project_id": "projectID",
		"cluster_id": "clusterID",
		"authentication_type": "consumer",
		"application_endpoint": "ovh-eu",
		"application_key": "key",
		"application_secret": "secret",
		"application_consumer_key": "consumer_key"
	}`

	manager, err := NewManager(bytes.NewBufferString(cfg))
	if err != nil {
		assert.FailNow(t, "failed to create manager", err)
	}

	client := &sdk.ClientMock{}
	ctx := context.Background()

	client.On("ListNodePoolNodes", ctx, "projectID", "clusterID", "id").Return(
		[]sdk.Node{
			{
				ID:     "id-1",
				Name:   "node-1",
				Status: "READY",
			},
			{
				ID:     "id-2",
				Name:   "node-2",
				Status: "INSTALLING",
			},
			{
				ID:     "id-3",
				Name:   "node-3",
				Status: "ERROR",
			},
		}, nil,
	)

	name := "pool-b2-7"
	size := uint32(3)
	min := uint32(1)
	max := uint32(5)

	opts := sdk.CreateNodePoolOpts{
		FlavorName:   "b2-7",
		Name:         &name,
		DesiredNodes: &size,
		MinNodes:     &min,
		MaxNodes:     &max,
		Autoscale:    true,
	}

	client.On("CreateNodePool", ctx, "projectID", "clusterID", &opts).Return(
		&sdk.NodePool{
			ID:           "id",
			Name:         "pool-b2-7",
			Flavor:       "b2-7",
			Autoscale:    true,
			DesiredNodes: 3,
			MinNodes:     1,
			MaxNodes:     5,
		}, nil,
	)

	client.On("DeleteNodePool", ctx, "projectID", "clusterID", "id").Return(&sdk.NodePool{}, nil)

	manager.Client = client
	return &NodeGroup{
		Manager: manager,
		NodePool: sdk.NodePool{
			ID:           "id",
			Name:         "pool-b2-7",
			Flavor:       "b2-7",
			Autoscale:    true,
			DesiredNodes: 3,
			MinNodes:     1,
			MaxNodes:     5,
		},

		CurrentSize: 3,
	}
}

func TestOVHCloudNodeGroup_MaxSize(t *testing.T) {
	group := newTestNodeGroup(t)

	t.Run("check default node group max size", func(t *testing.T) {
		max := group.MaxSize()

		assert.Equal(t, 5, max)
	})
}

func TestOVHCloudNodeGroup_MinSize(t *testing.T) {
	group := newTestNodeGroup(t)

	t.Run("check default node group min size", func(t *testing.T) {
		min := group.MinSize()

		assert.Equal(t, 1, min)
	})
}

func TestOVHCloudNodeGroup_TargetSize(t *testing.T) {
	group := newTestNodeGroup(t)

	t.Run("check default node group target size", func(t *testing.T) {
		size, err := group.TargetSize()
		assert.NoError(t, err)

		assert.Equal(t, 3, size)
	})
}

func TestOVHCloudNodeGroup_IncreaseSize(t *testing.T) {
	group := newTestNodeGroup(t)

	t.Run("check increase size below max size", func(t *testing.T) {
		err := group.IncreaseSize(1)
		assert.NoError(t, err)

		size, err := group.TargetSize()
		assert.NoError(t, err)

		assert.Equal(t, 4, size)
	})

	t.Run("check increase size above max size", func(t *testing.T) {
		err := group.IncreaseSize(5)
		assert.Error(t, err)
	})

	t.Run("check increase size with negative delta", func(t *testing.T) {
		err := group.IncreaseSize(-1)
		assert.Error(t, err)
	})
}

func TestOVHCloudNodeGroup_DeleteNodes(t *testing.T) {
	group := newTestNodeGroup(t)

	t.Run("check delete nodes above min size", func(t *testing.T) {
		err := group.DeleteNodes([]*v1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-1",
					Labels: map[string]string{
						"nodepool": group.Id(),
					},
				},
			},
		})
		assert.NoError(t, err)

		size, err := group.TargetSize()
		assert.NoError(t, err)

		assert.Equal(t, 2, size)
	})

	t.Run("check delete nodes empty nodes array", func(t *testing.T) {
		err := group.DeleteNodes(nil)

		size, err := group.TargetSize()
		assert.NoError(t, err)

		assert.Equal(t, 2, size)
	})

	t.Run("check delete nodes below min size", func(t *testing.T) {
		err := group.DeleteNodes([]*v1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-1",
					Labels: map[string]string{
						"nodepool": group.Id(),
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-2",
					Labels: map[string]string{
						"nodepool": group.Id(),
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-3",
					Labels: map[string]string{
						"nodepool": group.Id(),
					},
				},
			},
		})
		assert.Error(t, err)
	})

	t.Run("check delete nodes with wrong label name", func(t *testing.T) {
		err := group.DeleteNodes([]*v1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-1",
					Labels: map[string]string{
						"nodepool": "unknown",
					},
				},
			},
		})

		assert.Error(t, err)
	})
}

func TestOVHCloudNodeGroup_DecreaseTargetSize(t *testing.T) {
	group := newTestNodeGroup(t)

	t.Run("check decrease size above min size", func(t *testing.T) {
		err := group.DecreaseTargetSize(-1)
		assert.NoError(t, err)

		size, err := group.TargetSize()
		assert.NoError(t, err)

		assert.Equal(t, 2, size)
	})

	t.Run("check decrease size below min size", func(t *testing.T) {
		err := group.DecreaseTargetSize(-5)
		assert.Error(t, err)
	})

	t.Run("check decrease size with positive delta", func(t *testing.T) {
		err := group.DecreaseTargetSize(1)
		assert.Error(t, err)
	})
}

func TestOVHCloudNodeGroup_Id(t *testing.T) {
	group := newTestNodeGroup(t)

	t.Run("check node group id fetch", func(t *testing.T) {
		name := group.Id()

		assert.Equal(t, "pool-b2-7", name)
	})
}

func TestOVHCloudNodeGroup_Debug(t *testing.T) {
	group := newTestNodeGroup(t)

	t.Run("check node group debug print", func(t *testing.T) {
		debug := group.Debug()

		assert.Equal(t, "pool-b2-7 (3:1:5)", debug)
	})
}

func TestOVHCloudNodeGroup_Nodes(t *testing.T) {
	group := newTestNodeGroup(t)

	t.Run("check nodes list in node group", func(t *testing.T) {
		nodes, err := group.Nodes()
		assert.NoError(t, err)

		assert.Equal(t, 3, len(nodes))

		assert.Equal(t, "id-1/node-1", nodes[0].Id)
		assert.Equal(t, cloudprovider.InstanceRunning, nodes[0].Status.State)

		assert.Equal(t, "id-2/node-2", nodes[1].Id)
		assert.Equal(t, cloudprovider.InstanceCreating, nodes[1].Status.State)

		assert.Equal(t, "id-3/node-3", nodes[2].Id)
		assert.Equal(t, cloudprovider.OtherErrorClass, nodes[2].Status.ErrorInfo.ErrorClass)
		assert.Equal(t, "ERROR", nodes[2].Status.ErrorInfo.ErrorCode)
	})
}

func TestOVHCloudNodeGroup_TemplateNodeInfo(t *testing.T) {
	group := newTestNodeGroup(t)

	t.Run("check node template filled with group id", func(t *testing.T) {
		template, err := group.TemplateNodeInfo()
		assert.NoError(t, err)

		node := template.Node()
		assert.NotNil(t, node)

		assert.Contains(t, node.ObjectMeta.Name, fmt.Sprintf("%s-node-", group.Id()))
		assert.Equal(t, group.Id(), node.Labels["nodepool"])
	})
}

func TestOVHCloudNodeGroup_Exist(t *testing.T) {
	group := newTestNodeGroup(t)

	t.Run("check exist is true", func(t *testing.T) {
		exist := group.Exist()

		assert.True(t, exist)
	})
}

func TestOVHCloudNodeGroup_Create(t *testing.T) {
	group := newTestNodeGroup(t)

	t.Run("check create node group", func(t *testing.T) {
		newGroup, err := group.Create()
		assert.NoError(t, err)

		size, err := newGroup.TargetSize()
		assert.NoError(t, err)

		assert.Equal(t, "pool-b2-7", newGroup.Id())
		assert.Equal(t, 3, size)
		assert.Equal(t, 1, newGroup.MinSize())
		assert.Equal(t, 5, newGroup.MaxSize())
	})
}

func TestOVHCloudNodeGroup_Delete(t *testing.T) {
	group := newTestNodeGroup(t)

	t.Run("check delete node group", func(t *testing.T) {
		err := group.Delete()

		assert.NoError(t, err)
	})
}

func TestOVHCloudNodeGroup_Autoprovisioned(t *testing.T) {
	group := newTestNodeGroup(t)

	t.Run("check auto-provisioned is false", func(t *testing.T) {
		provisioned := group.Autoprovisioned()

		assert.False(t, provisioned)
	})
}
