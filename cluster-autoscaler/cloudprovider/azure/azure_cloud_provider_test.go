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

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2022-03-01/compute"     //nolint SA1019 - deprecated package
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2017-05-10/resources" //nolint SA1019 - deprecated package
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"sigs.k8s.io/cloud-provider-azure/pkg/azureclients/vmclient/mockvmclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azureclients/vmssclient/mockvmssclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azureclients/vmssvmclient/mockvmssvmclient"
)

func newTestAzureManager(t *testing.T) *AzureManager {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedScaleSets := newTestVMSSList(3, testASG, "eastus", compute.Uniform)

	expectedVMSSVMs := newTestVMSSVMList(3)
	mockVMSSClient := mockvmssclient.NewMockInterface(ctrl)
	mockVMSSClient.EXPECT().List(gomock.Any(), "rg").Return(expectedScaleSets, nil).AnyTimes()
	mockVMSSVMClient := mockvmssvmclient.NewMockInterface(ctrl)
	mockVMSSVMClient.EXPECT().List(gomock.Any(), "rg", testASG, gomock.Any()).Return(expectedVMSSVMs, nil).AnyTimes()

	manager := &AzureManager{
		env:                  &azure.PublicCloud,
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

	cache := newAzureCache(manager.azClient, refreshInterval, manager.config)

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
		newTestScaleSet(provider.azureManager, testASG))
	assert.True(t, registered)
	assert.Equal(t, len(provider.NodeGroups()), 1)
}

func TestNodeGroupForNode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	orchestrationModes := [2]compute.OrchestrationMode{compute.Uniform, compute.Flexible}

	expectedVMSSVMs := newTestVMSSVMList(3)
	expectedVMs := newTestVMList(3)

	for _, orchMode := range orchestrationModes {
		expectedScaleSets := newTestVMSSList(3, testASG, "eastus", orchMode)
		provider := newTestProvider(t)
		mockVMSSClient := mockvmssclient.NewMockInterface(ctrl)
		mockVMSSClient.EXPECT().List(gomock.Any(), provider.azureManager.config.ResourceGroup).Return(expectedScaleSets, nil)
		provider.azureManager.azClient.virtualMachineScaleSetsClient = mockVMSSClient

		mockVMSSVMClient := mockvmssvmclient.NewMockInterface(ctrl)
		provider.azureManager.azClient.virtualMachineScaleSetVMsClient = mockVMSSVMClient

		mockVMClient := mockvmclient.NewMockInterface(ctrl)
		provider.azureManager.azClient.virtualMachinesClient = mockVMClient

		if orchMode == compute.Uniform {
			mockVMSSVMClient.EXPECT().List(gomock.Any(), provider.azureManager.config.ResourceGroup,
				testASG, gomock.Any()).Return(expectedVMSSVMs, nil).AnyTimes()
		} else {
			provider.azureManager.config.EnableVmssFlex = true
			mockVMClient.EXPECT().ListVmssFlexVMsWithoutInstanceView(gomock.Any(), testASG).Return(expectedVMs, nil).AnyTimes()
		}

		registered := provider.azureManager.RegisterNodeGroup(
			newTestScaleSet(provider.azureManager, testASG))
		provider.azureManager.explicitlyConfigured[testASG] = true
		assert.True(t, registered)
		assert.Equal(t, len(provider.NodeGroups()), 1)

		node := newApiNode(orchMode, 0)
		// refresh cache
		err := provider.azureManager.forceRefresh()
		assert.NoError(t, err)
		group, err := provider.NodeGroupForNode(node)
		assert.NoError(t, err)
		assert.NotNil(t, group, "Group should not be nil")
		assert.Equal(t, group.Id(), testASG)
		assert.Equal(t, group.MinSize(), 1)
		assert.Equal(t, group.MaxSize(), 5)

		// test node in cluster that is not in a group managed by cluster autoscaler
		nodeNotInGroup := &apiv1.Node{
			Spec: apiv1.NodeSpec{
				ProviderID: azurePrefix + "/subscriptions/subscripion/resourceGroups/test-resource-group/providers/" +
					"Microsoft.Compute/virtualMachines/test-instance-id-not-in-group",
			},
		}
		group, err = provider.NodeGroupForNode(nodeNotInGroup)
		assert.NoError(t, err)
		assert.Nil(t, group)
	}
}

func TestNodeGroupForNodeWithNoProviderId(t *testing.T) {
	provider := newTestProvider(t)
	registered := provider.azureManager.RegisterNodeGroup(
		newTestScaleSet(provider.azureManager, testASG))
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
