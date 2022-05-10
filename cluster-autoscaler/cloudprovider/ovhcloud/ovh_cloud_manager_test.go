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
	"time"

	"github.com/stretchr/testify/assert"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/ovhcloud/sdk"
)

func newTestManager(t *testing.T) *OvhCloudManager {
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

	client.On("ListClusterFlavors", ctx, "projectID", "clusterID").Return(
		[]sdk.Flavor{
			{
				Name:     "b2-7",
				Category: "b",
				State:    "available",
				VCPUs:    2,
				GPUs:     0,
				RAM:      7,
			},
			{
				Name:     "t1-45",
				Category: "t",
				State:    "available",
				VCPUs:    8,
				GPUs:     1,
				RAM:      45,
			},
			{
				Name:     "unknown",
				Category: "",
				State:    "unavailable",
				VCPUs:    2,
				GPUs:     0,
				RAM:      7,
			},
		}, nil,
	)
	manager.Client = client

	return manager
}

func TestOvhCloudManager_getFlavorsByName(t *testing.T) {
	expectedFlavorsByNameFromAPICall := map[string]sdk.Flavor{
		"b2-7": {
			Name:     "b2-7",
			Category: "b",
			State:    "available",
			VCPUs:    2,
			GPUs:     0,
			RAM:      7,
		},
		"t1-45": {
			Name:     "t1-45",
			Category: "t",
			State:    "available",
			VCPUs:    8,
			GPUs:     1,
			RAM:      45,
		},
		"unknown": {
			Name:     "unknown",
			Category: "",
			State:    "unavailable",
			VCPUs:    2,
			GPUs:     0,
			RAM:      7,
		},
	}

	t.Run("brand new manager: list from api", func(t *testing.T) {
		ng := newTestManager(t)
		flavorsByName, err := ng.getFlavorsByName()

		ng.Client.(*sdk.ClientMock).AssertCalled(t, "ListClusterFlavors", context.Background(), "projectID", "clusterID")
		assert.NoError(t, err)
		assert.Equal(t, expectedFlavorsByNameFromAPICall, flavorsByName)
		assert.Equal(t, expectedFlavorsByNameFromAPICall, ng.FlavorsCache)
	})

	t.Run("flavors cache expired: renew and list from api", func(t *testing.T) {
		initialFlavorsCache := map[string]sdk.Flavor{
			"custom": {
				Name: "custom",
			},
		}

		ng := newTestManager(t)
		ng.FlavorsCache = initialFlavorsCache
		ng.FlavorsCacheExpirationTime = time.Now()

		flavorsByName, err := ng.getFlavorsByName()

		ng.Client.(*sdk.ClientMock).AssertCalled(t, "ListClusterFlavors", context.Background(), "projectID", "clusterID")
		assert.NoError(t, err)
		assert.Equal(t, expectedFlavorsByNameFromAPICall, flavorsByName)
		assert.Equal(t, expectedFlavorsByNameFromAPICall, ng.FlavorsCache)
	})

	t.Run("flavors cache still valid: list from cache", func(t *testing.T) {
		initialFlavorsCache := map[string]sdk.Flavor{
			"custom": {
				Name: "custom",
			},
		}

		ng := newTestManager(t)
		ng.FlavorsCache = initialFlavorsCache
		ng.FlavorsCacheExpirationTime = time.Now().Add(time.Minute)

		flavorsByName, err := ng.getFlavorsByName()

		ng.Client.(*sdk.ClientMock).AssertNotCalled(t, "ListClusterFlavors", context.Background(), "projectID", "clusterID")
		assert.NoError(t, err)
		assert.Equal(t, initialFlavorsCache, flavorsByName)
		assert.Equal(t, initialFlavorsCache, ng.FlavorsCache)
	})
}

func TestOvhCloudManager_getFlavorByName(t *testing.T) {
	ng := newTestManager(t)

	t.Run("check default node group max size", func(t *testing.T) {
		flavor, err := ng.getFlavorByName("b2-7")
		assert.NoError(t, err)
		assert.Equal(t, sdk.Flavor{
			Name:     "b2-7",
			Category: "b",
			State:    "available",
			VCPUs:    2,
			GPUs:     0,
			RAM:      7,
		}, flavor)
	})
}
