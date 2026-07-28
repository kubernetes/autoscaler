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
	"k8s.io/autoscaler/cluster-autoscaler/config"
)

func newTestProvider(t *testing.T) (*VKEProvider, *sdk.ClientMock) {
	t.Helper()
	manager, client := newTestManager(t)

	provider := &VKEProvider{
		manager:            manager,
		autoscalingOptions: config.AutoscalingOptions{},
		resourceLimiter:    cloudprovider.NewResourceLimiter(nil, nil),
	}
	return provider, client
}

func TestVKEProvider_Name(t *testing.T) {
	provider, _ := newTestProvider(t)
	assert.Equal(t, cloudprovider.VKEProviderName, provider.Name())
}

func TestVKEProvider_NodeGroupsAndRefresh(t *testing.T) {
	provider, client := newTestProvider(t)
	ctx := context.Background()

	assert.Empty(t, provider.NodeGroups())

	pools := []sdk.NodePool{
		{ID: "pool-1", Name: "workers", Flavor: "flavor-1", MinNodes: 1, MaxNodes: 5, CurrentNodes: 2},
		{ID: "pool-2", Name: "gpu", Flavor: "flavor-gpu", MinNodes: 0, MaxNodes: 3, CurrentNodes: 0},
	}
	client.On("ListNodePools", ctx, "clusterID").Return(pools, nil)

	// Skip OpenStack auth in unit tests by injecting client and bypassing ReAuthenticate.
	provider.manager.NodePools = pools
	groups := provider.NodeGroups()
	assert.Len(t, groups, 2)
	assert.Equal(t, "workers", groups[0].Id())
	assert.Equal(t, "gpu", groups[1].Id())
}

func TestVKEProvider_NodeGroupForNode(t *testing.T) {
	provider, client := newTestProvider(t)
	ctx := context.Background()

	provider.manager.NodePools = []sdk.NodePool{
		{ID: "pool-1", Name: "workers", Flavor: "flavor-1", MinNodes: 1, MaxNodes: 5, CurrentNodes: 1},
	}

	t.Run("empty provider id", func(t *testing.T) {
		ng, err := provider.NodeGroupForNode(&apiv1.Node{})
		assert.NoError(t, err)
		assert.Nil(t, ng)
	})

	t.Run("from cache", func(t *testing.T) {
		cached := &NodeGroup{NodePool: sdk.NodePool{ID: "pool-1", Name: "workers"}, Manager: provider.manager}
		provider.manager.setNodeGroupPerProviderID("openstack:///instance-1", cached)

		ng, err := provider.NodeGroupForNode(&apiv1.Node{
			ObjectMeta: metav1.ObjectMeta{Name: "node-1"},
			Spec:       apiv1.NodeSpec{ProviderID: "openstack:///instance-1"},
		})
		assert.NoError(t, err)
		assert.Equal(t, "workers", ng.Id())
	})

	t.Run("by listing nodes", func(t *testing.T) {
		provider.manager.NodeGroupPerProviderID = make(map[string]*NodeGroup)
		client.On("ListNodePoolNodes", ctx, "clusterID", "pool-1").Return(
			[]sdk.Node{{Id: "instance-2", Status: "READY"}}, nil,
		)

		ng, err := provider.NodeGroupForNode(&apiv1.Node{
			ObjectMeta: metav1.ObjectMeta{Name: "node-2"},
			Spec:       apiv1.NodeSpec{ProviderID: "openstack:///instance-2"},
		})
		assert.NoError(t, err)
		assert.NotNil(t, ng)
		assert.Equal(t, "workers", ng.Id())
	})
}

func TestVKEProvider_GPUAndMachineTypes(t *testing.T) {
	provider, client := newTestProvider(t)
	ctx := context.Background()

	client.On("ListClusterFlavors", ctx, "clusterID").Return(
		[]sdk.Flavor{
			{Id: "flavor-1", State: "available", VCPUs: 2, GPUs: 0, RAM: 4},
			{Id: "flavor-gpu", State: "available", VCPUs: 8, GPUs: 1, RAM: 32},
			{Id: "flavor-old", State: "unavailable", VCPUs: 2, GPUs: 0, RAM: 4},
		}, nil,
	)

	types, err := provider.GetAvailableMachineTypes()
	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{"flavor-1", "flavor-gpu"}, types)

	gpus := provider.GetAvailableGPUTypes()
	_, ok := gpus["flavor-gpu"]
	assert.True(t, ok)
	assert.Equal(t, GPULabel, provider.GPULabel())
}

func TestVKEProvider_OptionalMethods(t *testing.T) {
	provider, _ := newTestProvider(t)

	_, pricingErr := provider.Pricing()
	assert.Equal(t, cloudprovider.ErrNotImplemented, pricingErr)

	_, hasErr := provider.HasInstance(&apiv1.Node{})
	assert.Equal(t, cloudprovider.ErrNotImplemented, hasErr)

	assert.NoError(t, provider.Cleanup())

	rl, err := provider.GetResourceLimiter()
	assert.NoError(t, err)
	assert.NotNil(t, rl)
}
