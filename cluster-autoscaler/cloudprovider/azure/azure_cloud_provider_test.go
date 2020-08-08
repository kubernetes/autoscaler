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
	"io/ioutil"
	"os"
	"testing"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	azclients "k8s.io/legacy-cloud-providers/azure/clients"
	"k8s.io/legacy-cloud-providers/azure/clients/vmssclient/mockvmssclient"
	"k8s.io/legacy-cloud-providers/azure/clients/vmssvmclient/mockvmssvmclient"

	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2017-05-10/resources"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func newTestAzureManager(t *testing.T) *AzureManager {
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

	registered := provider.azureManager.RegisterAsg(
		newTestScaleSet(provider.azureManager, "test-asg"))
	assert.True(t, registered)
	assert.Equal(t, len(provider.NodeGroups()), 1)

	node := &apiv1.Node{
		Spec: apiv1.NodeSpec{
			ProviderID: "azure://" + fmt.Sprintf(fakeVirtualMachineScaleSetVMID, 0),
		},
	}
	// refresh cache
	provider.azureManager.regenerateCache()
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
	registered := provider.azureManager.RegisterAsg(
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

func TestBuildAzure(t *testing.T) {
	expectedConfig := &Config{
		Cloud:               "AzurePublicCloud",
		Location:            "southeastasia",
		TenantID:            "tenantId",
		SubscriptionID:      "subId",
		ResourceGroup:       "rg",
		VMType:              "vmss",
		AADClientID:         "clientId",
		AADClientSecret:     "clientsecret",
		MaxDeploymentsCount: 10,
		CloudProviderRateLimitConfig: CloudProviderRateLimitConfig{
			RateLimitConfig: azclients.RateLimitConfig{
				CloudProviderRateLimit:            false,
				CloudProviderRateLimitBucket:      5,
				CloudProviderRateLimitBucketWrite: 5,
				CloudProviderRateLimitQPS:         1,
				CloudProviderRateLimitQPSWrite:    1,
			},
			InterfaceRateLimit: &azclients.RateLimitConfig{
				CloudProviderRateLimit:            false,
				CloudProviderRateLimitBucket:      5,
				CloudProviderRateLimitBucketWrite: 5,
				CloudProviderRateLimitQPS:         1,
				CloudProviderRateLimitQPSWrite:    1,
			},
			VirtualMachineRateLimit: &azclients.RateLimitConfig{
				CloudProviderRateLimit:            false,
				CloudProviderRateLimitBucket:      5,
				CloudProviderRateLimitBucketWrite: 5,
				CloudProviderRateLimitQPS:         1,
				CloudProviderRateLimitQPSWrite:    1,
			},
			StorageAccountRateLimit: &azclients.RateLimitConfig{
				CloudProviderRateLimit:            false,
				CloudProviderRateLimitBucket:      5,
				CloudProviderRateLimitBucketWrite: 5,
				CloudProviderRateLimitQPS:         1,
				CloudProviderRateLimitQPSWrite:    1,
			},
			DiskRateLimit: &azclients.RateLimitConfig{
				CloudProviderRateLimit:            false,
				CloudProviderRateLimitBucket:      5,
				CloudProviderRateLimitBucketWrite: 5,
				CloudProviderRateLimitQPS:         1,
				CloudProviderRateLimitQPSWrite:    1,
			},
			VirtualMachineScaleSetRateLimit: &azclients.RateLimitConfig{
				CloudProviderRateLimit:            false,
				CloudProviderRateLimitBucket:      5,
				CloudProviderRateLimitBucketWrite: 5,
				CloudProviderRateLimitQPS:         1,
				CloudProviderRateLimitQPSWrite:    1,
			},
			KubernetesServiceRateLimit: &azclients.RateLimitConfig{
				CloudProviderRateLimit:            false,
				CloudProviderRateLimitBucket:      5,
				CloudProviderRateLimitBucketWrite: 5,
				CloudProviderRateLimitQPS:         1,
				CloudProviderRateLimitQPSWrite:    1,
			},
		},
	}

	cloudConfig := "cfg.json"
	azureCfg := `{
		"cloud": "AzurePublicCloud",
		"tenantId": "tenantId",
		"subscriptionId": "subId",
		"aadClientId": "clientId",
		"aadClientSecret": "clientsecret",
		"resourceGroup": "rg",
		"location": "southeastasia"
	}`
	err := ioutil.WriteFile(cloudConfig, []byte(azureCfg), 0644)
	assert.Nil(t, err)
	defer os.Remove(cloudConfig)

	os.Setenv("ARM_CLOUD", "AzurePublicCloud")
	os.Setenv("LOCATION", "southeastasia")
	os.Setenv("ARM_RESOURCE_GROUP", "rg")
	os.Setenv("ARM_SUBSCRIPTION_ID", "subId")
	os.Setenv("ARM_TENANT_ID", "tenantId")
	os.Setenv("ARM_CLIENT_ID", "clientId")
	os.Setenv("ARM_CLIENT_SECRET", "clientsecret")

	resourceLimiter := &cloudprovider.ResourceLimiter{}
	discoveryOptions := cloudprovider.NodeGroupDiscoveryOptions{}

	testCases := []struct {
		name        string
		cloudConfig string
	}{
		{
			name:        "BuildAzure should create Azure Manager using cloud-config file",
			cloudConfig: cloudConfig,
		},
		{
			name: "BuildAzure should create Azure Manager using environment variable",
		},
	}

	for _, test := range testCases {
		opts := config.AutoscalingOptions{
			CloudConfig: test.cloudConfig,
		}
		cloudProvider := BuildAzure(opts, discoveryOptions, resourceLimiter)
		assert.Equal(t, expectedConfig, cloudProvider.(*AzureCloudProvider).azureManager.config, test.name)
	}

	os.Unsetenv("ARM_CLOUD")
	os.Unsetenv("LOCATION")
	os.Unsetenv("ARM_RESOURCE_GROUP")
	os.Unsetenv("ARM_SUBSCRIPTION_ID")
	os.Unsetenv("ARM_TENANT_ID")
	os.Unsetenv("ARM_CLIENT_ID")
	os.Unsetenv("ARM_CLIENT_SECRET")
}
