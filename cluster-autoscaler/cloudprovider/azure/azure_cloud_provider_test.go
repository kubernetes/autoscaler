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
	"fmt"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2017-05-10/resources"
	"github.com/Azure/go-autorest/autorest/to"
	"testing"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/legacy-cloud-providers/azure/clients/vmssclient/mockvmssclient"
	"k8s.io/legacy-cloud-providers/azure/clients/vmssvmclient/mockvmssvmclient"

	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func newTestAzureManager(t *testing.T) *AzureManager {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedScaleSets := newTestVMSSList(3, "test-vmss", "eastus")
	expectedVMSSVMs := newTestVMSSVMList(3)
	mockVMSSClient := mockvmssclient.NewMockInterface(ctrl)
	mockVMSSClient.EXPECT().List(gomock.Any(), "rg").Return(expectedScaleSets, nil).AnyTimes()
	mockVMSSVMClient := mockvmssvmclient.NewMockInterface(ctrl)
	mockVMSSVMClient.EXPECT().List(gomock.Any(), "rg", "test-vmss", gomock.Any()).Return(expectedVMSSVMs, nil).AnyTimes()

	manager := &AzureManager{
		env:                  azure.PublicCloud,
		explicitlyConfigured: make(map[string]bool),
		config: &Config{
			ResourceGroup:       "rg",
			VMType:              vmTypeVMSS,
			MaxDeploymentsCount: 2,
			Deployment:          "deployment",
		},
		azClient: &azClient{
			virtualMachineScaleSetsClient:   mockVMSSClient,
			virtualMachineScaleSetVMsClient: mockVMSSVMClient,
			deploymentsClient: &DeploymentsClientMock{
				FakeStore: map[string]resources.DeploymentExtended{
					"deployment": {
						Name: to.StringPtr("deployment"),
						Properties: &resources.DeploymentPropertiesExtended{Template: map[string]interface{}{
							resourcesFieldName: []interface{}{
								map[string]interface{}{
									typeFieldName: nsgResourceType,
								},
								map[string]interface{}{
									typeFieldName: rtResourceType,
								},
							},
						}},
					},
				},
			},
		},
	}

	cache, error := newAzureCache(manager.azClient, refreshInterval, manager.config.ResourceGroup, vmTypeVMSS)
	assert.NoError(t, error)

	manager.azureCache = cache
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

	registered := provider.azureManager.RegisterNodeGroup(
		newTestScaleSet(provider.azureManager, "test-asg"))
	assert.True(t, registered)
	assert.Equal(t, len(provider.NodeGroups()), 1)
}

func TestNodeGroupForNode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedVMSSVMs := newTestVMSSVMList(3)
	expectedScaleSets := newTestVMSSList(3, "test-asg", "eastus")

	provider := newTestProvider(t)
	mockVMSSClient := mockvmssclient.NewMockInterface(ctrl)
	mockVMSSClient.EXPECT().List(gomock.Any(), provider.azureManager.config.ResourceGroup).Return(expectedScaleSets, nil)
	provider.azureManager.azClient.virtualMachineScaleSetsClient = mockVMSSClient
	mockVMSSVMClient := mockvmssvmclient.NewMockInterface(ctrl)
	mockVMSSVMClient.EXPECT().List(gomock.Any(), provider.azureManager.config.ResourceGroup, "test-asg", gomock.Any()).Return(expectedVMSSVMs, nil).AnyTimes()
	provider.azureManager.azClient.virtualMachineScaleSetVMsClient = mockVMSSVMClient

	registered := provider.azureManager.RegisterNodeGroup(
		newTestScaleSet(provider.azureManager, "test-asg"))
	assert.True(t, registered)
	assert.Equal(t, len(provider.NodeGroups()), 1)

	node := &apiv1.Node{
		Spec: apiv1.NodeSpec{
			ProviderID: "azure://" + fmt.Sprintf(fakeVirtualMachineScaleSetVMID, 0),
		},
	}
	// refresh cache
	provider.azureManager.forceRefresh()
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

func TestNodeGroupForNodeWithNoProviderId(t *testing.T) {
	provider := newTestProvider(t)
	registered := provider.azureManager.RegisterNodeGroup(
		newTestScaleSet(provider.azureManager, "test-asg"))
	assert.True(t, registered)
	assert.Equal(t, len(provider.NodeGroups()), 1)

	node := &apiv1.Node{
		Spec: apiv1.NodeSpec{
			ProviderID: "",
		},
	}
	group, err := provider.NodeGroupForNode(node)

	assert.NoError(t, err)
	assert.Equal(t, group, nil)
}
