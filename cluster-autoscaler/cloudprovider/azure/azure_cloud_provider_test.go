/*
Copyright 2017 The Kubernetes Authors.

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

	"github.com/Azure/azure-sdk-for-go/arm/compute"
	"github.com/Azure/azure-sdk-for-go/arm/resources/resources"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/stretchr/testify/assert"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
)

func newTestAzureManager(t *testing.T) *AzureManager {
	manager := &AzureManager{
		env:                  azure.PublicCloud,
		explicitlyConfigured: make(map[string]bool),
		config: &Config{
			ResourceGroup: "test",
			VMType:        vmTypeVMSS,
		},

		azClient: &azClient{
			disksClient:           &DisksClientMock{},
			interfacesClient:      &InterfacesClientMock{},
			storageAccountsClient: &AccountsClientMock{},
			deploymentsClient: &DeploymentsClientMock{
				FakeStore: make(map[string]resources.DeploymentExtended),
			},
			virtualMachinesClient: &VirtualMachinesClientMock{
				FakeStore: make(map[string]map[string]compute.VirtualMachine),
			},
			virtualMachineScaleSetsClient: &VirtualMachineScaleSetsClientMock{
				FakeStore: make(map[string]map[string]compute.VirtualMachineScaleSet),
			},
			virtualMachineScaleSetVMsClient: &VirtualMachineScaleSetVMsClientMock{},
		},
	}
	cache, error := newAsgCache()
	assert.NoError(t, error)

	manager.asgCache = cache
	return manager
}

func newTestProvider(t *testing.T) *AzureCloudProvider {
	manager := newTestAzureManager(t)
	resourceLimiter := cloudprovider.NewResourceLimiter(
		map[string]int64{cloudprovider.ResourceNameCores: 1, cloudprovider.ResourceNameMemory: 10000000},
		map[string]int64{cloudprovider.ResourceNameCores: 10, cloudprovider.ResourceNameMemory: 100000000})

	return &AzureCloudProvider{
		azureManager:    manager,
		resourceLimiter: resourceLimiter,
	}
}

func TestBuildAzureCloudProvider(t *testing.T) {
	resourceLimiter := cloudprovider.NewResourceLimiter(
		map[string]int64{cloudprovider.ResourceNameCores: 1, cloudprovider.ResourceNameMemory: 10000000},
		map[string]int64{cloudprovider.ResourceNameCores: 10, cloudprovider.ResourceNameMemory: 100000000})
	m := newTestAzureManager(t)
	_, err := BuildAzureCloudProvider(m, resourceLimiter)
	assert.NoError(t, err)
}

func TestName(t *testing.T) {
	provider := newTestProvider(t)
	assert.Equal(t, provider.Name(), "azure")
}

func TestNodeGroups(t *testing.T) {
	provider := newTestProvider(t)
	assert.Equal(t, len(provider.NodeGroups()), 0)

	registered := provider.azureManager.RegisterAsg(
		newTestScaleSet(provider.azureManager, "test-asg"))
	assert.True(t, registered)
	assert.Equal(t, len(provider.NodeGroups()), 1)
}

func TestNodeGroupForNode(t *testing.T) {
	provider := newTestProvider(t)
	registered := provider.azureManager.RegisterAsg(
		newTestScaleSet(provider.azureManager, "test-asg"))
	assert.True(t, registered)
	assert.Equal(t, len(provider.NodeGroups()), 1)

	node := &apiv1.Node{
		Spec: apiv1.NodeSpec{
			ProviderID: "azure://" + fakeVirtualMachineScaleSetVMID,
		},
	}
	group, err := provider.NodeGroupForNode(node)
	assert.NoError(t, err)
	assert.NotNil(t, group, "Group should not be nil")
	assert.Equal(t, group.Id(), "test-asg")
	assert.Equal(t, group.MinSize(), 1)
	assert.Equal(t, group.MaxSize(), 5)

	// test node in cluster that is not in a group managed by cluster autoscaler
	nodeNotInGroup := &apiv1.Node{
		Spec: apiv1.NodeSpec{
			ProviderID: "azure:///subscriptions/subscripion/resourceGroups/test-resource-group/providers/Microsoft.Compute/virtualMachines/test-instance-id-not-in-group",
		},
	}
	group, err = provider.NodeGroupForNode(nodeNotInGroup)
	assert.NoError(t, err)
	assert.Nil(t, group)
}
