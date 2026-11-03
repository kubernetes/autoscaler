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
	"bytes"
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/vke/sdk"
)

func validConfigJSON() string {
	return `{
		"cluster_id": "clusterID",
		"tenant_id": "tenantID",
		"application_key": "appKey",
		"application_secret": "appSecret",
		"openstack_auth_url": "https://example.com:5000/v3"
	}`
}

func newTestManager(t *testing.T) (*VKEManager, *sdk.ClientMock) {
	t.Helper()

	manager, err := NewManager(bytes.NewBufferString(validConfigJSON()))
	assert.NoError(t, err)

	client := &sdk.ClientMock{}
	manager.Client = client
	return manager, client
}

func TestNewManager(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		manager, err := NewManager(bytes.NewBufferString(validConfigJSON()))
		assert.NoError(t, err)
		assert.Equal(t, "clusterID", manager.ClusterID)
		assert.Equal(t, "tenantID", manager.ProjectID)
		assert.Nil(t, manager.Client)
		assert.Nil(t, manager.OpenStackProvider)
	})

	t.Run("missing cluster_id", func(t *testing.T) {
		cfg := `{
			"tenant_id": "tenantID",
			"application_key": "appKey",
			"application_secret": "appSecret",
			"openstack_auth_url": "https://example.com:5000/v3"
		}`
		_, err := NewManager(bytes.NewBufferString(cfg))
		assert.Error(t, err)
	})

	t.Run("invalid json", func(t *testing.T) {
		_, err := NewManager(bytes.NewBufferString("{"))
		assert.Error(t, err)
	})
}

func TestVKEManager_getFlavorByName(t *testing.T) {
	manager, client := newTestManager(t)
	ctx := context.Background()

	client.On("ListClusterFlavors", ctx, "clusterID").Return(
		[]sdk.Flavor{
			{Id: "flavor-1", State: "available", VCPUs: 2, GPUs: 0, RAM: 4},
			{Id: "flavor-gpu", State: "available", VCPUs: 8, GPUs: 1, RAM: 32},
		}, nil,
	)

	flavor, err := manager.getFlavorByName("flavor-1")
	assert.NoError(t, err)
	assert.Equal(t, 2, flavor.VCPUs)

	_, err = manager.getFlavorByName("missing")
	assert.Error(t, err)
}

func TestVKEManager_NodeGroupPerProviderIDCache(t *testing.T) {
	manager, _ := newTestManager(t)
	ng := &NodeGroup{Manager: manager}

	manager.setNodeGroupPerProviderID("openstack:///instance-1", ng)
	assert.Equal(t, ng, manager.getNodeGroupPerProviderID("openstack:///instance-1"))
	assert.Nil(t, manager.getNodeGroupPerProviderID("openstack:///missing"))
}

func TestVKEManager_NodeGroupPerProviderIDConcurrency(t *testing.T) {
	manager, _ := newTestManager(t)
	ng := &NodeGroup{Manager: manager}

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			manager.setNodeGroupPerProviderID("openstack:///instance", ng)
		}()
		go func() {
			defer wg.Done()
			_ = manager.getNodeGroupPerProviderID("openstack:///instance")
		}()
	}
	wg.Wait()
}

func TestVKEManager_flavorCacheExpiry(t *testing.T) {
	manager, client := newTestManager(t)
	ctx := context.Background()

	client.On("ListClusterFlavors", ctx, "clusterID").Return(
		[]sdk.Flavor{{Id: "flavor-1", State: "available", VCPUs: 2}}, nil,
	).Once()

	_, err := manager.getFlavorByName("flavor-1")
	assert.NoError(t, err)

	manager.FlavorsCacheExpirationTime = time.Now().Add(-time.Minute)
	client.On("ListClusterFlavors", ctx, "clusterID").Return(
		[]sdk.Flavor{{Id: "flavor-1", State: "available", VCPUs: 4}}, nil,
	).Once()

	flavor, err := manager.getFlavorByName("flavor-1")
	assert.NoError(t, err)
	assert.Equal(t, 4, flavor.VCPUs)
	client.AssertExpectations(t)
}
