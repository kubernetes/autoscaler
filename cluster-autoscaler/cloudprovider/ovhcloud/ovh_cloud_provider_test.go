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
	"testing"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/stretchr/testify/assert"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/ovhcloud/sdk"
)

func newTestProvider(t *testing.T) *OVHCloudProvider {
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

	// fill the test provider with some example
	client := &sdk.ClientMock{}
	ctx := context.Background()

	client.On("ListNodePools", ctx, "projectID", "clusterID").Return(
		[]sdk.NodePool{
			{
				ID:           "1",
				Name:         "pool-1",
				Flavor:       "b2-7",
				DesiredNodes: 2,
				MinNodes:     1,
				MaxNodes:     5,
				Autoscale:    true,
			},
			{
				ID:           "2",
				Name:         "pool-2",
				Flavor:       "b2-7",
				DesiredNodes: 1,
				MinNodes:     0,
				MaxNodes:     3,
				Autoscale:    false,
			},
		}, nil,
	)

	client.On("ListFlavors", ctx, "projectID", "clusterID").Return(
		[]sdk.Flavor{
			{
				Name:     "b2-7",
				Category: "b",
				State:    "available",
			},
			{
				Name:     "t1-45",
				Category: "t",
				State:    "available",
			},
			{
				Name:     "unknown",
				Category: "",
				State:    "unavailable",
			},
		}, nil,
	)

	manager.Client = client

	minLimits := map[string]int64{cloudprovider.ResourceNameCores: 1, cloudprovider.ResourceNameMemory: 10000000}
	maxLimits := map[string]int64{cloudprovider.ResourceNameCores: 10, cloudprovider.ResourceNameMemory: 100000000}

	provider := &OVHCloudProvider{
		manager:         manager,
		resourceLimiter: cloudprovider.NewResourceLimiter(minLimits, maxLimits),
	}

	err = provider.Refresh()
	assert.NoError(t, err)

	return provider
}

func TestOVHCloudProvider_BuildOVHcloud(t *testing.T) {
	t.Run("create new OVHcloud provider", func(t *testing.T) {
		_ = newTestProvider(t)
	})
}

func TestOVHCloudProvider_Name(t *testing.T) {
	provider := newTestProvider(t)

	t.Run("check OVHcloud provider name", func(t *testing.T) {
		name := provider.Name()

		assert.Equal(t, cloudprovider.OVHcloudProviderName, name)
	})
}

func TestOVHCloudProvider_NodeGroups(t *testing.T) {
	provider := newTestProvider(t)

	t.Run("check default node groups length", func(t *testing.T) {
		groups := provider.NodeGroups()

		assert.Equal(t, 1, len(groups))
	})

	t.Run("check empty node groups length after reset", func(t *testing.T) {
		provider.manager.NodePools = []sdk.NodePool{}
		groups := provider.NodeGroups()

		assert.Equal(t, 0, len(groups))
	})
}

func TestOVHCloudProvider_NodeGroupForNode(t *testing.T) {
	provider := newTestProvider(t)

	t.Run("check node group with correct label on node", func(t *testing.T) {
		node := &apiv1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "node-1",
				Labels: map[string]string{
					"nodepool": "pool-1",
				},
			},
		}

		group, err := provider.NodeGroupForNode(node)
		assert.NoError(t, err)
		assert.NotNil(t, group)

		assert.Equal(t, "pool-1", group.Id())
		assert.Equal(t, 1, group.MinSize())
		assert.Equal(t, 5, group.MaxSize())
	})

	t.Run("check node group empty if label not exists", func(t *testing.T) {
		node := &apiv1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "node-1",
				Labels: map[string]string{
					"nodepool": "pool-unknown",
				},
			},
		}

		group, err := provider.NodeGroupForNode(node)
		assert.NoError(t, err)
		assert.Nil(t, group)
	})

	t.Run("check node group empty if label not present", func(t *testing.T) {
		node := &apiv1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "node-1",
				Labels: map[string]string{},
			},
		}

		group, err := provider.NodeGroupForNode(node)
		assert.NoError(t, err)
		assert.Nil(t, group)
	})
}

func TestOVHCloudProvider_Pricing(t *testing.T) {
	provider := newTestProvider(t)

	t.Run("not implemented", func(t *testing.T) {
		_, err := provider.Pricing()
		assert.Error(t, err)
	})
}

func TestOVHCloudProvider_GetAvailableMachineTypes(t *testing.T) {
	provider := newTestProvider(t)

	t.Run("check available machine types", func(t *testing.T) {
		flavors, err := provider.GetAvailableMachineTypes()
		assert.NoError(t, err)

		assert.Equal(t, 2, len(flavors))
	})
}

func TestOVHCloudProvider_NewNodeGroup(t *testing.T) {
	provider := newTestProvider(t)

	t.Run("check new node group default values", func(t *testing.T) {
		group, err := provider.NewNodeGroup("b2-7", nil, nil, nil, nil)
		assert.NoError(t, err)

		assert.Contains(t, group.Id(), "b2-7")
		assert.Equal(t, 0, group.MinSize())
		assert.Equal(t, 100, group.MaxSize())
	})
}

func TestOVHCloudProvider_GetResourceLimiter(t *testing.T) {
	provider := newTestProvider(t)

	t.Run("check default resource limiter values", func(t *testing.T) {
		rl, err := provider.GetResourceLimiter()
		assert.NoError(t, err)

		minCpu := rl.GetMin(cloudprovider.ResourceNameCores)
		minMem := rl.GetMin(cloudprovider.ResourceNameMemory)

		maxCpu := rl.GetMax(cloudprovider.ResourceNameCores)
		maxMem := rl.GetMax(cloudprovider.ResourceNameMemory)

		assert.Equal(t, int64(1), minCpu)
		assert.Equal(t, int64(10000000), minMem)
		assert.Equal(t, int64(10), maxCpu)
		assert.Equal(t, int64(100000000), maxMem)
	})
}

func TestOVHCloudProvider_GPULabel(t *testing.T) {
	provider := newTestProvider(t)

	t.Run("check gpu label annotation", func(t *testing.T) {
		label := provider.GPULabel()

		assert.Equal(t, GPULabel, label)
	})
}

func TestOVHCloudProvider_GetAvailableGPUTypes(t *testing.T) {
	provider := newTestProvider(t)

	t.Run("check available gpu machine types", func(t *testing.T) {
		flavors := provider.GetAvailableGPUTypes()

		assert.Equal(t, 1, len(flavors))
		assert.Equal(t, struct{}{}, flavors["t1-45"])
	})
}

func TestOVHCloudProvider_Cleanup(t *testing.T) {
	provider := newTestProvider(t)

	t.Run("check return nil", func(t *testing.T) {
		err := provider.Cleanup()
		assert.NoError(t, err)
	})
}

func TestOVHCloudProvider_Refresh(t *testing.T) {
	provider := newTestProvider(t)

	t.Run("check refresh reset node groups correctly", func(t *testing.T) {
		provider.manager.NodePools = []sdk.NodePool{}
		groups := provider.NodeGroups()

		assert.Equal(t, 0, len(groups))

		err := provider.Refresh()
		assert.NoError(t, err)

		groups = provider.NodeGroups()
		assert.Equal(t, 1, len(groups))
	})
}
