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
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v5"
	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2022-08-01/compute"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2017-05-10/resources"
	"github.com/Azure/go-autorest/autorest/to"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"sigs.k8s.io/cloud-provider-azure/pkg/azureclients/vmclient/mockvmclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azureclients/vmssclient/mockvmssclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azureclients/vmssvmclient/mockvmssvmclient"

	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func newTestAzureManager(t *testing.T) *AzureManager {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedScaleSets := newTestVMSSList(3, "test-asg", "eastus", compute.Uniform)
	expectedVMSSVMs := newTestVMSSVMList(3)
	mockVMSSClient := mockvmssclient.NewMockInterface(ctrl)
	mockVMSSClient.EXPECT().List(gomock.Any(), "rg").Return(expectedScaleSets, nil).AnyTimes()
	mockVMSSVMClient := mockvmssvmclient.NewMockInterface(ctrl)
	mockVMSSVMClient.EXPECT().List(gomock.Any(), "rg", "test-asg", gomock.Any()).Return(expectedVMSSVMs, nil).AnyTimes()
	mockVMClient := mockvmclient.NewMockInterface(ctrl)
	expectedVMs := newTestVMList(3)
	mockVMClient.EXPECT().List(gomock.Any(), "rg").Return(expectedVMs, nil).AnyTimes()

	manager := &AzureManager{
		env:                  azure.PublicCloud,
		explicitlyConfigured: make(map[string]bool),
		config: &Config{
			ResourceGroup:       "rg",
			VMType:              vmTypeVMSS,
			MaxDeploymentsCount: 2,
			Deployment:          "deployment",
			EnableForceDelete:   true,
			Location:            "eastus",
		},
		azClient: &azClient{
			virtualMachineScaleSetsClient:   mockVMSSClient,
			virtualMachineScaleSetVMsClient: mockVMSSVMClient,
			virtualMachinesClient:           mockVMClient,
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

	cache, error := newAzureCache(manager.azClient, refreshInterval, *manager.config)
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
		newTestScaleSet(provider.azureManager, "test-asg"),
	)
	assert.True(t, registered)
	registered = provider.azureManager.RegisterNodeGroup(
		newTestVMsPool(provider.azureManager, "test-vms-pool"),
	)
	assert.True(t, registered)
	assert.Equal(t, len(provider.NodeGroups()), 2)
}

func TestHasInstance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	provider := newTestProvider(t)
	mockVMSSClient := mockvmssclient.NewMockInterface(ctrl)
	mockVMClient := mockvmclient.NewMockInterface(ctrl)
	mockVMSSVMClient := mockvmssvmclient.NewMockInterface(ctrl)
	mockAgentpoolclient := NewMockAgentPoolsClient(ctrl)
	provider.azureManager.azClient.virtualMachinesClient = mockVMClient
	provider.azureManager.azClient.virtualMachineScaleSetsClient = mockVMSSClient
	provider.azureManager.azClient.virtualMachineScaleSetVMsClient = mockVMSSVMClient
	provider.azureManager.azClient.agentPoolClient = mockAgentpoolclient
	provider.azureManager.azureCache.clusterName = "test-cluster"
	provider.azureManager.azureCache.clusterResourceGroup = "test-rg"
	provider.azureManager.azureCache.enableVMsAgentPool = true // enable VMs agent pool to support mixed node group types

	// Simulate node groups and instances
	expectedScaleSets := newTestVMSSList(3, "test-asg", "eastus", compute.Uniform)
	expectedVMsPoolVMs := newTestVMsPoolVMList(3)
	expectedVMSSVMs := newTestVMSSVMList(3)

	mockVMSSClient.EXPECT().List(gomock.Any(), provider.azureManager.config.ResourceGroup).Return(expectedScaleSets, nil).AnyTimes()
	mockVMClient.EXPECT().List(gomock.Any(), provider.azureManager.config.ResourceGroup).Return(expectedVMsPoolVMs, nil).AnyTimes()
	mockVMSSVMClient.EXPECT().List(gomock.Any(), provider.azureManager.config.ResourceGroup, "test-asg", gomock.Any()).Return(expectedVMSSVMs, nil).AnyTimes()
	vmssType := armcontainerservice.AgentPoolTypeVirtualMachines
	vmssPool := armcontainerservice.AgentPool{
		Name: to.StringPtr("test-asg"),
		Properties: &armcontainerservice.ManagedClusterAgentPoolProfileProperties{
			Type: &vmssType,
		},
	}

	vmsPool := getTestVMsAgentPool("test-vms-pool", false)
	fakeAPListPager := getFakeAgentpoolListPager(&vmssPool, &vmsPool)
	mockAgentpoolclient.EXPECT().NewListPager(provider.azureManager.azureCache.clusterResourceGroup, provider.azureManager.azureCache.clusterName, nil).
		Return(fakeAPListPager).AnyTimes()

	// Register node groups
	assert.Equal(t, len(provider.NodeGroups()), 0)
	registered := provider.azureManager.RegisterNodeGroup(
		newTestScaleSet(provider.azureManager, "test-asg"),
	)
	provider.azureManager.explicitlyConfigured["test-asg"] = true
	assert.True(t, registered)

	registered = provider.azureManager.RegisterNodeGroup(
		newTestVMsPool(provider.azureManager, "test-vms-pool"),
	)
	provider.azureManager.explicitlyConfigured["test-vms-pool"] = true
	assert.True(t, registered)
	assert.Equal(t, len(provider.NodeGroups()), 2)

	// Refresh cache
	provider.azureManager.forceRefresh()

	// Test HasInstance for a node from the VMSS pool
	node := newApiNode(compute.Uniform, 0)
	hasInstance, err := provider.azureManager.azureCache.HasInstance(node.Spec.ProviderID)
	assert.True(t, hasInstance)
	assert.NoError(t, err)

	// Test HasInstance for a node from the VMs pool
	vmsPoolNode := newVMsNode(0)
	hasInstance, err = provider.azureManager.azureCache.HasInstance(vmsPoolNode.Spec.ProviderID)
	assert.True(t, hasInstance)
	assert.NoError(t, err)
}

func TestUnownedInstancesFallbackToDeletionTaint(t *testing.T) {
	// VMSS Instances that belong to a VMSS on the cluster but do not belong to a registered ASG
	// should return err unimplemented for HasInstance
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	provider := newTestProvider(t)
	mockVMSSClient := mockvmssclient.NewMockInterface(ctrl)
	mockVMClient := mockvmclient.NewMockInterface(ctrl)
	mockVMSSVMClient := mockvmssvmclient.NewMockInterface(ctrl)
	provider.azureManager.azClient.virtualMachinesClient = mockVMClient
	provider.azureManager.azClient.virtualMachineScaleSetsClient = mockVMSSClient
	provider.azureManager.azClient.virtualMachineScaleSetVMsClient = mockVMSSVMClient

	// 	// Simulate VMSS instances
	unregisteredVMSSInstance := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "unregistered-vmss-node",
		},
		Spec: apiv1.NodeSpec{
			ProviderID: "azure:///subscriptions/sub/resourceGroups/rg/providers/Microsoft.Compute/virtualMachineScaleSets/unregistered-vmss-instance-id/virtualMachines/0",
		},
	}
	// Mock responses to simulate that the instance belongs to a VMSS not in any registered ASG
	expectedVMSSVMs := newTestVMSSVMList(1)
	mockVMSSVMClient.EXPECT().List(gomock.Any(), provider.azureManager.config.ResourceGroup, "unregistered-vmss-instance-id", gomock.Any()).Return(expectedVMSSVMs, nil).AnyTimes()

	// Call HasInstance and check the result
	hasInstance, err := provider.azureManager.azureCache.HasInstance(unregisteredVMSSInstance.Spec.ProviderID)
	assert.False(t, hasInstance)
	assert.Equal(t, cloudprovider.ErrNotImplemented, err)
}

func TestHasInstanceProviderIDErrorValidation(t *testing.T) {
	provider := newTestProvider(t)
	// Test case: Node with an empty ProviderID
	nodeWithoutValidProviderID := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-node",
		},
		Spec: apiv1.NodeSpec{
			ProviderID: "",
		},
	}
	_, err := provider.HasInstance(nodeWithoutValidProviderID)
	assert.Equal(t, "ProviderID for node: test-node is empty, skipped", err.Error())

	// Test cases: Nodes with invalid ProviderID prefixes
	invalidProviderIDs := []string{
		"aazure://",
		"kubemark://",
		"kwok://",
		"incorrect!",
	}

	for _, providerID := range invalidProviderIDs {
		invalidProviderIDNode := &apiv1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-node",
			},
			Spec: apiv1.NodeSpec{
				ProviderID: providerID,
			},
		}
		_, err := provider.HasInstance(invalidProviderIDNode)
		assert.Equal(t, "invalid azure ProviderID prefix for node: test-node, skipped", err.Error())
	}
}

func TestMixedNodeGroups(t *testing.T) {
	ctrl := gomock.NewController(t)
	provider := newTestProvider(t)
	provider.azureManager.azureCache.clusterName = "test-cluster"
	provider.azureManager.azureCache.clusterResourceGroup = "test-rg"
	provider.azureManager.azureCache.enableVMsAgentPool = true // enable VMs agent pool to support mixed node group types
	mockVMSSClient := mockvmssclient.NewMockInterface(ctrl)
	mockVMClient := mockvmclient.NewMockInterface(ctrl)
	mockVMSSVMClient := mockvmssvmclient.NewMockInterface(ctrl)
	mockAgentpoolclient := NewMockAgentPoolsClient(ctrl)
	provider.azureManager.azClient.virtualMachinesClient = mockVMClient
	provider.azureManager.azClient.virtualMachineScaleSetsClient = mockVMSSClient
	provider.azureManager.azClient.virtualMachineScaleSetVMsClient = mockVMSSVMClient
	provider.azureManager.azClient.agentPoolClient = mockAgentpoolclient

	expectedScaleSets := newTestVMSSList(3, "test-asg", "eastus", compute.Uniform)
	expectedVMsPoolVMs := newTestVMsPoolVMList(3)
	expectedVMSSVMs := newTestVMSSVMList(3)

	mockVMSSClient.EXPECT().List(gomock.Any(), provider.azureManager.config.ResourceGroup).Return(expectedScaleSets, nil).AnyTimes()
	mockVMClient.EXPECT().List(gomock.Any(), provider.azureManager.config.ResourceGroup).Return(expectedVMsPoolVMs, nil).AnyTimes()
	mockVMSSVMClient.EXPECT().List(gomock.Any(), provider.azureManager.config.ResourceGroup, "test-asg", gomock.Any()).Return(expectedVMSSVMs, nil).AnyTimes()

	vmssType := armcontainerservice.AgentPoolTypeVirtualMachines
	vmssPool := armcontainerservice.AgentPool{
		Name: to.StringPtr("test-asg"),
		Properties: &armcontainerservice.ManagedClusterAgentPoolProfileProperties{
			Type: &vmssType,
		},
	}

	vmsPool := getTestVMsAgentPool("test-vms-pool", false)
	fakeAPListPager := getFakeAgentpoolListPager(&vmssPool, &vmsPool)
	mockAgentpoolclient.EXPECT().NewListPager(provider.azureManager.azureCache.clusterResourceGroup, provider.azureManager.azureCache.clusterName, nil).
		Return(fakeAPListPager).AnyTimes()

	assert.Equal(t, len(provider.NodeGroups()), 0)
	registered := provider.azureManager.RegisterNodeGroup(
		newTestScaleSet(provider.azureManager, "test-asg"),
	)
	provider.azureManager.explicitlyConfigured["test-asg"] = true
	assert.True(t, registered)

	registered = provider.azureManager.RegisterNodeGroup(
		newTestVMsPool(provider.azureManager, "test-vms-pool"),
	)
	provider.azureManager.explicitlyConfigured["test-vms-pool"] = true
	assert.True(t, registered)
	assert.Equal(t, len(provider.NodeGroups()), 2)

	// refresh cache
	provider.azureManager.forceRefresh()

	// node from vmss pool
	node := newApiNode(compute.Uniform, 0)
	group, err := provider.NodeGroupForNode(node)
	assert.NoError(t, err)
	assert.NotNil(t, group, "Group should not be nil")
	assert.Equal(t, group.Id(), "test-asg")
	assert.Equal(t, group.MinSize(), 1)
	assert.Equal(t, group.MaxSize(), 5)

	// node from vms pool
	vmsPoolNode := newVMsNode(0)
	group, err = provider.NodeGroupForNode(vmsPoolNode)
	assert.NoError(t, err)
	assert.NotNil(t, group, "Group should not be nil")
	assert.Equal(t, group.Id(), "test-vms-pool")
	assert.Equal(t, group.MinSize(), 3)
	assert.Equal(t, group.MaxSize(), 10)
}

func TestNodeGroupForNode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	orchestrationModes := []compute.OrchestrationMode{compute.Uniform, compute.Flexible}

	expectedVMSSVMs := newTestVMSSVMList(3)
	expectedVMs := newTestVMList(3)

	for _, orchMode := range orchestrationModes {
		t.Run(fmt.Sprintf("OrchestrationMode_%v", orchMode), func(t *testing.T) {
			expectedScaleSets := newTestVMSSList(3, "test-asg", "eastus", orchMode)
			provider := newTestProvider(t)
			mockVMSSClient := mockvmssclient.NewMockInterface(ctrl)
			mockVMSSClient.EXPECT().List(gomock.Any(), provider.azureManager.config.ResourceGroup).Return(expectedScaleSets, nil)
			provider.azureManager.azClient.virtualMachineScaleSetsClient = mockVMSSClient
			mockVMClient := mockvmclient.NewMockInterface(ctrl)
			provider.azureManager.azClient.virtualMachinesClient = mockVMClient
			mockVMClient.EXPECT().List(gomock.Any(), provider.azureManager.config.ResourceGroup).Return(expectedVMs, nil).AnyTimes()

			if orchMode == compute.Uniform {
				mockVMSSVMClient := mockvmssvmclient.NewMockInterface(ctrl)
				mockVMSSVMClient.EXPECT().List(gomock.Any(), provider.azureManager.config.ResourceGroup, "test-asg", gomock.Any()).Return(expectedVMSSVMs, nil).AnyTimes()
				provider.azureManager.azClient.virtualMachineScaleSetVMsClient = mockVMSSVMClient
			} else {
				provider.azureManager.config.EnableVmssFlex = true
				mockVMClient.EXPECT().ListVmssFlexVMsWithoutInstanceView(gomock.Any(), "test-asg").Return(expectedVMs, nil).AnyTimes()
			}

			registered := provider.azureManager.RegisterNodeGroup(
				newTestScaleSet(provider.azureManager, "test-asg"))
			provider.azureManager.explicitlyConfigured["test-asg"] = true
			assert.True(t, registered)
			assert.Equal(t, len(provider.NodeGroups()), 1)

			node := newApiNode(orchMode, 0)
			// refresh cache
			provider.azureManager.forceRefresh()
			group, err := provider.NodeGroupForNode(node)
			assert.NoError(t, err)
			assert.NotNil(t, group, "Group should not be nil")
			assert.Equal(t, group.Id(), "test-asg")
			assert.Equal(t, group.MinSize(), 1)
			assert.Equal(t, group.MaxSize(), 5)

			hasInstance, err := provider.HasInstance(node)
			assert.True(t, hasInstance)
			assert.NoError(t, err)

			// test node in cluster that is not in a group managed by cluster autoscaler
			nodeNotInGroup := &apiv1.Node{
				Spec: apiv1.NodeSpec{
					ProviderID: "azure:///subscriptions/subscription/resourceGroups/test-resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/test/virtualMachines/test-instance-id-not-in-group",
				},
			}
			group, err = provider.NodeGroupForNode(nodeNotInGroup)
			assert.NoError(t, err)
			assert.Nil(t, group)

			hasInstance, err = provider.HasInstance(nodeNotInGroup)
			assert.False(t, hasInstance)
			assert.Error(t, err)
			assert.Equal(t, err, cloudprovider.ErrNotImplemented)
		})
	}
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
