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
		manager := newTestManager(t)
		flavorsByName, err := manager.getFlavorsByName()

		manager.Client.(*sdk.ClientMock).AssertCalled(t, "ListClusterFlavors", context.Background(), "projectID", "clusterID")
		assert.NoError(t, err)
		assert.Equal(t, expectedFlavorsByNameFromAPICall, flavorsByName)
		assert.Equal(t, expectedFlavorsByNameFromAPICall, manager.FlavorsCache)
	})

	t.Run("flavors cache expired: renew and list from api", func(t *testing.T) {
		initialFlavorsCache := map[string]sdk.Flavor{
			"custom": {
				Name: "custom",
			},
		}

		manager := newTestManager(t)
		manager.FlavorsCache = initialFlavorsCache
		manager.FlavorsCacheExpirationTime = time.Now()

		flavorsByName, err := manager.getFlavorsByName()

		manager.Client.(*sdk.ClientMock).AssertCalled(t, "ListClusterFlavors", context.Background(), "projectID", "clusterID")
		assert.NoError(t, err)
		assert.Equal(t, expectedFlavorsByNameFromAPICall, flavorsByName)
		assert.Equal(t, expectedFlavorsByNameFromAPICall, manager.FlavorsCache)
	})

	t.Run("flavors cache still valid: list from cache", func(t *testing.T) {
		initialFlavorsCache := map[string]sdk.Flavor{
			"custom": {
				Name: "custom",
			},
		}

		manager := newTestManager(t)
		manager.FlavorsCache = initialFlavorsCache
		manager.FlavorsCacheExpirationTime = time.Now().Add(time.Minute)

		flavorsByName, err := manager.getFlavorsByName()

		manager.Client.(*sdk.ClientMock).AssertNotCalled(t, "ListClusterFlavors", context.Background(), "projectID", "clusterID")
		assert.NoError(t, err)
		assert.Equal(t, initialFlavorsCache, flavorsByName)
		assert.Equal(t, initialFlavorsCache, manager.FlavorsCache)
	})
}

func TestOvhCloudManager_getFlavorByName(t *testing.T) {
	manager := newTestManager(t)

	t.Run("check default node group max size", func(t *testing.T) {
		flavor, err := manager.getFlavorByName("b2-7")
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

func TestOvhCloudManager_setNodeGroupPerProviderID(t *testing.T) {
	manager := newTestManager(t)
	ng1 := NodeGroup{
		CurrentSize: 1,
	}

	type fields struct {
		NodeGroupPerProviderID map[string]*NodeGroup
	}
	type args struct {
		providerID string
		nodeGroup  *NodeGroup
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantCache map[string]*NodeGroup
	}{
		{
			name: "New entry",
			fields: fields{
				NodeGroupPerProviderID: map[string]*NodeGroup{},
			},
			args: args{
				providerID: "providerID1",
				nodeGroup:  &ng1,
			},
			wantCache: map[string]*NodeGroup{
				"providerID1": &ng1,
			},
		}, {
			name: "Replace entry",
			fields: fields{
				NodeGroupPerProviderID: map[string]*NodeGroup{
					"providerID1": {},
				},
			},
			args: args{
				providerID: "providerID1",
				nodeGroup:  &ng1,
			},
			wantCache: map[string]*NodeGroup{
				"providerID1": &ng1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager.NodeGroupPerProviderID = tt.fields.NodeGroupPerProviderID

			manager.setNodeGroupPerProviderID(tt.args.providerID, tt.args.nodeGroup)

			assert.Equal(t, tt.wantCache, manager.NodeGroupPerProviderID)
		})
	}
}

func TestOvhCloudManager_getNodeGroupPerProviderID(t *testing.T) {
	manager := newTestManager(t)
	ng1 := NodeGroup{
		CurrentSize: 1,
	}

	type fields struct {
		NodeGroupPerProviderID map[string]*NodeGroup
	}
	type args struct {
		providerID string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *NodeGroup
	}{
		{
			name: "Node group found",
			fields: fields{
				NodeGroupPerProviderID: map[string]*NodeGroup{
					"providerID1": &ng1,
				},
			},
			args: args{
				providerID: "providerID1",
			},
			want: &ng1,
		},
		{
			name: "Node group not found",
			fields: fields{
				NodeGroupPerProviderID: map[string]*NodeGroup{},
			},
			args: args{
				providerID: "providerID1",
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager.NodeGroupPerProviderID = tt.fields.NodeGroupPerProviderID

			assert.Equalf(t, tt.want, manager.getNodeGroupPerProviderID(tt.args.providerID), "getNodeGroupPerProviderID(%v)", tt.args.providerID)
		})
	}
}

func TestOvhCloudManager_cacheConcurrency(t *testing.T) {
	manager := newTestManager(t)

	t.Run("Check NodeGroupPerProviderID cache is safe for concurrency (needs to be run with -race)", func(t *testing.T) {
		go func() {
			manager.setNodeGroupPerProviderID("", &NodeGroup{})
		}()
		manager.getNodeGroupPerProviderID("")
	})
}

func TestOvhCloudManager_setNodePoolsState(t *testing.T) {
	manager := newTestManager(t)
	np1 := sdk.NodePool{ID: "np1", DesiredNodes: 1}
	np2 := sdk.NodePool{ID: "np2", DesiredNodes: 2}
	np3 := sdk.NodePool{ID: "np3", DesiredNodes: 3}

	type fields struct {
		NodePoolsPerID         map[string]*sdk.NodePool
		NodeGroupPerProviderID map[string]*NodeGroup
	}
	type args struct {
		poolsList []sdk.NodePool

		nodePoolsPerID         map[string]*sdk.NodePool
		nodeGroupPerProviderID map[string]*NodeGroup
	}
	tests := []struct {
		name                       string
		fields                     fields
		args                       args
		wantNodePoolsPerID         map[string]uint32 // ID => desired nodes
		wantNodeGroupPerProviderID map[string]uint32 // ID => desired nodes
	}{
		{
			name: "NodePoolsPerID and NodeGroupPerProviderID empty",
			fields: fields{
				NodePoolsPerID:         map[string]*sdk.NodePool{},
				NodeGroupPerProviderID: map[string]*NodeGroup{},
			},
			args: args{
				poolsList: []sdk.NodePool{
					np1,
				},
			},
			wantNodePoolsPerID:         map[string]uint32{"np1": 1},
			wantNodeGroupPerProviderID: map[string]uint32{},
		},
		{
			name: "NodePoolsPerID and NodeGroupPerProviderID empty",
			fields: fields{
				NodePoolsPerID: map[string]*sdk.NodePool{
					"np2": &np2,
					"np3": &np3,
				},
				NodeGroupPerProviderID: map[string]*NodeGroup{
					"np2-node-id": {NodePool: &np2},
					"np3-node-id": {NodePool: &np3},
				},
			},
			args: args{
				poolsList: []sdk.NodePool{
					{
						ID:           "np1",
						DesiredNodes: 1,
					},
					{
						ID:           "np2",
						DesiredNodes: 20,
					},
				},
			},
			wantNodePoolsPerID: map[string]uint32{
				"np1": 1,  // np1 added
				"np2": 20, // np2 updated
				// np3 removed
			},
			wantNodeGroupPerProviderID: map[string]uint32{
				"np2-node-id": 20,
				"np3-node-id": 3, // Node reference that eventually stays in cache must not crash
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager.NodePoolsPerID = tt.fields.NodePoolsPerID
			manager.NodeGroupPerProviderID = tt.fields.NodeGroupPerProviderID

			manager.setNodePoolsState(tt.args.poolsList)

			assert.Len(t, manager.NodePoolsPerID, len(tt.wantNodePoolsPerID))
			for id, desiredNodes := range tt.wantNodePoolsPerID {
				assert.Equal(t, desiredNodes, manager.NodePoolsPerID[id].DesiredNodes)
			}

			assert.Len(t, manager.NodeGroupPerProviderID, len(tt.wantNodeGroupPerProviderID))
			for nodeID, desiredNodes := range tt.wantNodeGroupPerProviderID {
				assert.Equal(t, desiredNodes, manager.NodeGroupPerProviderID[nodeID].DesiredNodes)
			}
		})
	}
}
