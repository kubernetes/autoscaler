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

package azure

import (
	"context"
	"net/http"
	"testing"

	providerazureconsts "sigs.k8s.io/cloud-provider-azure/pkg/consts"
	"sigs.k8s.io/cluster-autoscaler/pkg/cloudprovider"

	armpolicy "github.com/Azure/azure-sdk-for-go/sdk/azcore/arm/policy"
	azcorepolicy "github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	armcomputev8 "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v8"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"k8s.io/utils/ptr"
)

func TestFetchSKUs(t *testing.T) {
	ctx := context.Background()
	transport := &recordingTransport{
		responseBody: `{
			"value": [{
				"name": "Standard_D2s_v3",
				"resourceType": "virtualMachines",
				"locations": ["eastus"],
				"capabilities": [
					{"name": "vCPUs", "value": "2"},
					{"name": "MemoryGB", "value": "8"}
				]
			}]
		}`,
	}
	skuClient, err := armcomputev8.NewResourceSKUsClient("subscription", staticTokenCredential{}, &armpolicy.ClientOptions{
		ClientOptions: azcorepolicy.ClientOptions{Transport: transport},
	})
	require.NoError(t, err)

	cache := &azureCache{azClient: &azClient{skuClient: skuClient}}
	cache.skus, err = cache.fetchSKUs(ctx, "eastus")
	require.NoError(t, err)
	assert.True(t, cache.HasVMSKUs())

	require.NotNil(t, transport.request)
	assert.Equal(t, http.MethodGet, transport.request.Method)
	assert.Equal(t, "/subscriptions/subscription/providers/Microsoft.Compute/skus", transport.request.URL.Path)
	assert.Equal(t, "2021-07-01", transport.request.URL.Query().Get("api-version"))
	assert.Equal(t, "location eq 'eastus'", transport.request.URL.Query().Get("$filter"))

	sku, err := cache.GetSKU(ctx, "Standard_D2s_v3", "eastus")
	require.NoError(t, err)
	cpus, err := sku.VCPU()
	require.NoError(t, err)
	assert.Equal(t, int64(2), cpus)
	memory, err := sku.Memory()
	require.NoError(t, err)
	assert.Equal(t, float64(8), memory)
}

func TestFetchSKUsRequiresLocation(t *testing.T) {
	cache := &azureCache{}
	skus, err := cache.fetchSKUs(context.Background(), "")
	require.EqualError(t, err, "location not specified")
	assert.Nil(t, skus)
}

func TestFetchVMsPools(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	provider := newTestProvider(t)
	ac := provider.azureManager.azureCache
	mockAgentpoolclient := NewMockAgentPoolsClient(ctrl)
	ac.azClient.agentPoolClient = mockAgentpoolclient

	vmsPool := getTestVMsAgentPool(false)
	vmssPoolType := armcontainerservice.AgentPoolTypeVirtualMachineScaleSets
	vmssPool := armcontainerservice.AgentPool{
		Name: ptr.To("vmsspool1"),
		Properties: &armcontainerservice.ManagedClusterAgentPoolProfileProperties{
			Type: &vmssPoolType,
		},
	}
	invalidPool := armcontainerservice.AgentPool{}
	fakeAPListPager := getFakeAgentpoolListPager(&vmsPool, &vmssPool, &invalidPool)
	mockAgentpoolclient.EXPECT().NewListPager(gomock.Any(), gomock.Any(), nil).
		Return(fakeAPListPager)

	vmsPoolMap, err := ac.fetchVMsPools()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(vmsPoolMap))

	_, ok := vmsPoolMap[ptr.Deref(vmsPool.Name, "")]
	assert.True(t, ok)
}

func TestRegister(t *testing.T) {
	provider := newTestProvider(t)
	ss := newTestScaleSet(provider.azureManager, "ss")

	ac := provider.azureManager.azureCache
	ac.registeredNodeGroups = []cloudprovider.NodeGroup{ss}

	isSuccess := ac.Register(ss)
	assert.False(t, isSuccess)

	ss1 := newTestScaleSet(provider.azureManager, "ss")
	ss1.minSize = 2
	isSuccess = ac.Register(ss1)
	assert.True(t, isSuccess)
}

func TestUnRegister(t *testing.T) {
	provider := newTestProvider(t)
	ss := newTestScaleSet(provider.azureManager, "ss")
	ss1 := newTestScaleSet(provider.azureManager, "ss1")

	ac := provider.azureManager.azureCache
	ac.registeredNodeGroups = []cloudprovider.NodeGroup{ss, ss1}

	isSuccess := ac.Unregister(ss)
	assert.True(t, isSuccess)
	assert.Equal(t, 1, len(ac.registeredNodeGroups))
}

func TestFindForInstance(t *testing.T) {
	provider := newTestProvider(t)
	ac := provider.azureManager.azureCache

	inst := azureRef{Name: "/subscriptions/sub/resourceGroups/rg/providers/foo"}
	ac.unownedInstances = make(map[azureRef]bool)
	ac.unownedInstances[inst] = true
	nodeGroup, err := ac.FindForInstance(&inst, providerazureconsts.VMTypeVMSS)
	assert.Nil(t, nodeGroup)
	assert.NoError(t, err)

	ac.unownedInstances[inst] = false
	nodeGroup, err = ac.FindForInstance(&inst, providerazureconsts.VMTypeStandard)
	assert.Nil(t, nodeGroup)
	assert.NoError(t, err)
	assert.True(t, ac.unownedInstances[inst])
}
