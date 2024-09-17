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
	"testing"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	providerazureconsts "sigs.k8s.io/cloud-provider-azure/pkg/consts"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v5"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestFetchVMsPools(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	provider := newTestProvider(t)
	ac := provider.azureManager.azureCache
	mockAgentpoolclient := NewMockAgentPoolsClient(ctrl)
	ac.azClient.agentPoolClient = mockAgentpoolclient

	vmsPoolName := "vmspool1"
	vmsPool := getTestVMsAgentPool(vmsPoolName, false)
	vmssPoolName := "vmsspool1"
	vmssPoolType := armcontainerservice.AgentPoolTypeVirtualMachineScaleSets
	vmssPool := armcontainerservice.AgentPool{
		Name: &vmssPoolName,
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

	_, ok := vmsPoolMap[vmsPoolName]
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
