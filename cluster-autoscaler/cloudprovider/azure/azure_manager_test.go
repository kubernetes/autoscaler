/*
Copyright 2019 The Kubernetes Authors.

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
	"reflect"
	"strings"
	"testing"
	"time"

	"sigs.k8s.io/cloud-provider-azure/pkg/azureclients/vmclient/mockvmclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/retry"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2022-03-01/compute"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2017-05-10/resources"
	"github.com/Azure/go-autorest/autorest/date"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	uber_gomock "go.uber.org/mock/gomock"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	azclients "sigs.k8s.io/cloud-provider-azure/pkg/azureclients"
	"sigs.k8s.io/cloud-provider-azure/pkg/azureclients/vmssclient/mockvmssclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azureclients/vmssvmclient/mockvmssvmclient"
	providerazureconsts "sigs.k8s.io/cloud-provider-azure/pkg/consts"
	providerazure "sigs.k8s.io/cloud-provider-azure/pkg/provider"
	providerazureconfig "sigs.k8s.io/cloud-provider-azure/pkg/provider/config"
)

const validAzureCfg = `{
	"cloud": "AzurePublicCloud",
	"tenantId": "fakeId",
	"subscriptionId": "fakeId",
	"aadClientId": "fakeId",
	"aadClientSecret": "fakeId",
	"resourceGroup": "fakeId",
	"location": "southeastasia",
	"subnetName": "fakeName",
	"securityGroupName": "fakeName",
	"vnetName": "fakeName",
	"routeTableName": "fakeName",
	"primaryAvailabilitySetName": "fakeName",
	"vmssCacheTTLInSeconds": 60,
	"vmssVirtualMachinesCacheTTLInSeconds": 240,
	"vmssVmsCacheJitter": 120,
	"maxDeploymentsCount": 8,
	"cloudProviderRateLimit": false,
	"routeRateLimit": {
		"cloudProviderRateLimit": true,
		"cloudProviderRateLimitQPS": 3
	}
}`

const validAzureCfgLegacy = `{
	"cloud": "AzurePublicCloud",
	"tenantId": "fakeId",
	"subscriptionId": "fakeId",
	"resourceGroup": "fakeId",
	"location": "southeastasia",
	"useWorkloadIdentityExtension": true,
	"subnetName": "fakeName",
	"securityGroupName": "fakeName",
	"vnetName": "fakeName",
	"routeTableName": "fakeName",
	"primaryAvailabilitySetName": "fakeName",
	"vmssCacheTTL": 60,
	"vmssVmsCacheTTL": 240,
	"vmssVmsCacheJitter": 120,
	"maxDeploymentsCount": 8,
	"cloudProviderRateLimit": false,
	"routeRateLimit": {
		"cloudProviderRateLimit": true,
		"cloudProviderRateLimitQPS": 3
	}
}`

const validAzureCfgForStandardVMType = `{
	"cloud": "AzurePublicCloud",
	"tenantId": "fakeId",
	"subscriptionId": "fakeId",
	"aadClientId": "fakeId",
	"aadClientSecret": "fakeId",
	"resourceGroup": "fakeId",
	"vmType":"standard",
	"location": "southeastasia",
	"subnetName": "fakeName",
	"securityGroupName": "fakeName",
	"vnetName": "fakeName",
	"routeTableName": "fakeName",
	"primaryAvailabilitySetName": "fakeName",
	"vmssCacheTTLInSeconds": 60,
	"vmssVirtualMachinesCacheTTLInSeconds": 240,
	"vmssVmsCacheJitter": 120,
	"maxDeploymentsCount": 8,
	"cloudProviderRateLimit": false,
	"routeRateLimit": {
		"cloudProviderRateLimit": true,
		"cloudProviderRateLimitQPS": 3
	},
	"deployment":"cluster-autoscaler-0001",
	"deploymentParameters":{
		"Name": "cluster-autoscaler-0001",
		"Properties":{
			"ProvisioningState": "Succeeded",
			"Parameters": {
				"PoolName01VMSize":"PoolName01"
			},
			"Template": {
				"resources": [
					{
						"type":"Microsoft.Compute/virtualMachines/extensions",
						"name":"cluster-autoscaler-0001-resourceName",
						"properties": {
							"hardwareProfile":{
								"VMSize":"10G"
							}
						}
					}
				]
			}
		}
	}
}`

const validAzureCfgForStandardVMTypeWithoutDeploymentParameters = `{
	"cloud": "AzurePublicCloud",
	"tenantId": "fakeId",
	"subscriptionId": "fakeId",
	"aadClientId": "fakeId",
	"aadClientSecret": "fakeId",
	"resourceGroup": "fakeId",
	"vmType":"standard",
	"location": "southeastasia",
	"subnetName": "fakeName",
	"securityGroupName": "fakeName",
	"vnetName": "fakeName",
	"routeTableName": "fakeName",
	"primaryAvailabilitySetName": "fakeName",
	"vmssCacheTTLInSeconds": 60,
	"vmssVirtualMachinesCacheTTLInSeconds": 240,
	"vmssVmsCacheJitter": 120,
	"maxDeploymentsCount": 8,
	"cloudProviderRateLimit": false,
	"routeRateLimit": {
		"cloudProviderRateLimit": true,
		"cloudProviderRateLimitQPS": 3
	},
	"deployment":"cluster-autoscaler-0001"
}`

const validAzureCfgForVMsPool = `{
	"cloud": "AzurePublicCloud",
	"tenantId": "fakeId",
	"subscriptionId": "fakeId",
	"aadClientId": "fakeId",
	"aadClientSecret": "fakeId",
	"resourceGroup": "fakeId",
	"location": "southeastasia",
	"subnetName": "fakeName",
	"securityGroupName": "fakeName",
	"vnetName": "fakeName",
	"routeTableName": "fakeName",
	"primaryAvailabilitySetName": "fakeName",
	"vmssCacheTTLInSeconds": 60,
	"vmssVirtualMachinesCacheTTLInSeconds": 240,
	"vmssVmsCacheJitter": 120,
	"maxDeploymentsCount": 8,
	"cloudProviderRateLimit": false,
	"routeRateLimit": {
		"cloudProviderRateLimit": true,
		"cloudProviderRateLimitQPS": 3
	},

	"clusterName": "mycluster",
	"clusterResourceGroup": "myrg",
	"armBaseURLForAPClient": "nodeprovisioner-svc.nodeprovisioner.svc.cluster.local"
}`

const (
	invalidAzureCfg = `{{}"cloud": "AzurePublicCloud",}`
	testASG         = "test-asg"
)

func TestCreateAzureManagerValidConfig(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockVMClient := mockvmclient.NewMockInterface(ctrl)
	mockVMSSClient := mockvmssclient.NewMockInterface(ctrl)
	mockVMSSClient.EXPECT().List(gomock.Any(), "fakeId").Return([]compute.VirtualMachineScaleSet{}, nil).Times(2)
	mockVMClient.EXPECT().List(gomock.Any(), "fakeId").Return([]compute.VirtualMachine{}, nil).Times(2)
	mockAzClient := &azClient{
		virtualMachinesClient:         mockVMClient,
		virtualMachineScaleSetsClient: mockVMSSClient,
	}
	manager, err := createAzureManagerInternal(strings.NewReader(validAzureCfg), cloudprovider.NodeGroupDiscoveryOptions{}, mockAzClient)

	expectedConfig := &Config{
		Config: providerazure.Config{
			AzureAuthConfig: providerazureconfig.AzureAuthConfig{
				Cloud:           "AzurePublicCloud",
				TenantID:        "fakeId",
				AADClientID:     "fakeId",
				AADClientSecret: "fakeId",
				SubscriptionID:  "fakeId",
			},
			Location:                             "southeastasia",
			ResourceGroup:                        "fakeId",
			VMType:                               "vmss",
			VmssCacheTTLInSeconds:                60,
			VmssVirtualMachinesCacheTTLInSeconds: 240,
			CloudProviderRateLimitConfig: providerazureconfig.CloudProviderRateLimitConfig{
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
			},
		},
		VmssVmsCacheJitter:  120,
		MaxDeploymentsCount: 8,
		EnableVMsAgentPool:  false,
	}

	assert.NoError(t, err)
	assertStructsMinimallyEqual(t, *expectedConfig, *manager.config)
}

func TestCreateAzureManagerLegacyConfig(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockVMClient := mockvmclient.NewMockInterface(ctrl)
	mockVMSSClient := mockvmssclient.NewMockInterface(ctrl)
	mockVMSSClient.EXPECT().List(gomock.Any(), "fakeId").Return([]compute.VirtualMachineScaleSet{}, nil).Times(2)
	mockVMClient.EXPECT().List(gomock.Any(), "fakeId").Return([]compute.VirtualMachine{}, nil).Times(2)
	mockAzClient := &azClient{
		virtualMachinesClient:         mockVMClient,
		virtualMachineScaleSetsClient: mockVMSSClient,
	}
	manager, err := createAzureManagerInternal(strings.NewReader(validAzureCfgLegacy), cloudprovider.NodeGroupDiscoveryOptions{}, mockAzClient)

	expectedConfig := &Config{
		Config: providerazure.Config{
			AzureAuthConfig: providerazureconfig.AzureAuthConfig{
				Cloud:          "AzurePublicCloud",
				TenantID:       "fakeId",
				SubscriptionID: "fakeId",
			},
			Location:                             "southeastasia",
			ResourceGroup:                        "fakeId",
			VMType:                               "vmss",
			VmssCacheTTLInSeconds:                60,
			VmssVirtualMachinesCacheTTLInSeconds: 240,
			CloudProviderRateLimitConfig: providerazureconfig.CloudProviderRateLimitConfig{
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
			},
		},
		VmssVmsCacheJitter:                    120,
		MaxDeploymentsCount:                   8,
		UseFederatedWorkloadIdentityExtension: true,
	}

	assert.NoError(t, err)
	assertStructsMinimallyEqual(t, *expectedConfig, *manager.config)
}

func TestCreateAzureManagerValidConfigForStandardVMType(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockVMClient := mockvmclient.NewMockInterface(ctrl)
	mockVMClient.EXPECT().List(gomock.Any(), "fakeId").Return([]compute.VirtualMachine{}, nil).Times(2)
	mockVMSSClient := mockvmssclient.NewMockInterface(ctrl)
	mockVMSSClient.EXPECT().List(gomock.Any(), "fakeId").Return([]compute.VirtualMachineScaleSet{}, nil).Times(2)
	mockAzClient := &azClient{
		virtualMachinesClient:         mockVMClient,
		virtualMachineScaleSetsClient: mockVMSSClient,
	}
	manager, err := createAzureManagerInternal(strings.NewReader(validAzureCfgForStandardVMType), cloudprovider.NodeGroupDiscoveryOptions{}, mockAzClient)

	expectedConfig := &Config{
		Config: providerazure.Config{
			AzureAuthConfig: providerazureconfig.AzureAuthConfig{
				Cloud:           "AzurePublicCloud",
				TenantID:        "fakeId",
				AADClientID:     "fakeId",
				AADClientSecret: "fakeId",
				SubscriptionID:  "fakeId",
			},
			Location:                             "southeastasia",
			ResourceGroup:                        "fakeId",
			VMType:                               "standard",
			VmssCacheTTLInSeconds:                60,
			VmssVirtualMachinesCacheTTLInSeconds: 240,
			CloudProviderRateLimitConfig: providerazureconfig.CloudProviderRateLimitConfig{
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
			},
		},
		VmssVmsCacheJitter:  120,
		MaxDeploymentsCount: 8,
		Deployment:          "cluster-autoscaler-0001",
		DeploymentParameters: map[string]interface{}{
			"Name": "cluster-autoscaler-0001",
			"Properties": map[string]interface{}{
				"ProvisioningState": "Succeeded",
				"Parameters": map[string]interface{}{
					"PoolName01VMSize": "PoolName01",
				},
				"Template": map[string]interface{}{
					"resources": []interface{}{
						map[string]interface{}{
							"type": "Microsoft.Compute/virtualMachines/extensions",
							"name": "cluster-autoscaler-0001-resourceName",
							"properties": map[string]interface{}{
								"hardwareProfile": map[string]interface{}{
									"VMSize": "10G",
								},
							},
						},
					},
				},
			},
		},
	}

	assert.NoError(t, err)
	assertStructsMinimallyEqual(t, *expectedConfig, *manager.config)
}

func TestCreateAzureManagerValidConfigForStandardVMTypeWithoutDeploymentParameters(t *testing.T) {
	manager, err := createAzureManagerInternal(strings.NewReader(validAzureCfgForStandardVMTypeWithoutDeploymentParameters), cloudprovider.NodeGroupDiscoveryOptions{}, &azClient{})
	expectedErr := "open /var/lib/azure/azuredeploy.parameters.json: no such file or directory"
	assert.Nil(t, manager)
	assert.Equal(t, expectedErr, err.Error(), "return error does not match, expected: %v, actual: %v", expectedErr, err.Error())
}
func TestCreateAzureManagerValidConfigForVMsPool(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockVMClient := mockvmclient.NewMockInterface(ctrl)
	mockVMSSClient := mockvmssclient.NewMockInterface(ctrl)
	mockVMSSClient.EXPECT().List(gomock.Any(), "fakeId").Return([]compute.VirtualMachineScaleSet{}, nil).Times(2)
	mockVMClient.EXPECT().List(gomock.Any(), "fakeId").Return([]compute.VirtualMachine{}, nil).Times(2)
	mockAzClient := &azClient{
		virtualMachinesClient:         mockVMClient,
		virtualMachineScaleSetsClient: mockVMSSClient,
	}
	manager, err := createAzureManagerInternal(strings.NewReader(validAzureCfgForVMsPool), cloudprovider.NodeGroupDiscoveryOptions{}, mockAzClient)

	expectedConfig := &Config{
		Config: providerazure.Config{
			AzureAuthConfig: providerazureconfig.AzureAuthConfig{
				Cloud:           "AzurePublicCloud",
				TenantID:        "fakeId",
				AADClientID:     "fakeId",
				AADClientSecret: "fakeId",
				SubscriptionID:  "fakeId",
			},
			Location:                             "southeastasia",
			ResourceGroup:                        "fakeId",
			VMType:                               "vmss",
			VmssCacheTTLInSeconds:                60,
			VmssVirtualMachinesCacheTTLInSeconds: 240,
			CloudProviderRateLimitConfig: providerazureconfig.CloudProviderRateLimitConfig{
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
			},
		},
		VmssVmsCacheJitter:    120,
		MaxDeploymentsCount:   8,
		ClusterName:           "mycluster",
		ClusterResourceGroup:  "myrg",
		ARMBaseURLForAPClient: "nodeprovisioner-svc.nodeprovisioner.svc.cluster.local",
	}

	assert.NoError(t, err)
	assertStructsMinimallyEqual(t, *expectedConfig, *manager.config)
}

func TestCreateAzureManagerWithNilConfig(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockVMClient := mockvmclient.NewMockInterface(ctrl)
	mockVMSSClient := mockvmssclient.NewMockInterface(ctrl)
	mockVMSSClient.EXPECT().List(gomock.Any(), "resourceGroup").Return([]compute.VirtualMachineScaleSet{}, nil).AnyTimes()
	mockVMClient.EXPECT().List(gomock.Any(), "resourceGroup").Return([]compute.VirtualMachine{}, nil).AnyTimes()
	unberCtl := uber_gomock.NewController(t)
	mockAgentpoolclient := NewMockAgentPoolsClient(unberCtl)
	vmspool := getTestVMsAgentPool("vmspool", false)
	fakeAPListPager := getFakeAgentpoolListPager(&vmspool)
	mockAgentpoolclient.EXPECT().NewListPager(gomock.Any(), gomock.Any(), nil).Return(fakeAPListPager).AnyTimes()
	mockAzClient := &azClient{
		virtualMachinesClient:         mockVMClient,
		virtualMachineScaleSetsClient: mockVMSSClient,
		agentPoolClient:               mockAgentpoolclient,
	}

	expectedConfig := &Config{
		Config: providerazure.Config{
			AzureAuthConfig: providerazureconfig.AzureAuthConfig{
				Cloud:                       "AzurePublicCloud",
				TenantID:                    "tenantId",
				AADClientID:                 "aadClientId",
				AADClientSecret:             "aadClientSecret",
				AADClientCertPath:           "aadClientCertPath",
				AADClientCertPassword:       "aadClientCertPassword",
				UseManagedIdentityExtension: true,
				UserAssignedIdentityID:      "UserAssignedIdentityID",
				SubscriptionID:              "subscriptionId",
			},
			Location:                             "southeastasia",
			ResourceGroup:                        "resourceGroup",
			VMType:                               "vmss",
			VmssCacheTTLInSeconds:                100,
			VmssVirtualMachinesCacheTTLInSeconds: 110,
			CloudProviderBackoff:                 true,
			CloudProviderBackoffRetries:          1,
			CloudProviderBackoffExponent:         1,
			CloudProviderBackoffDuration:         1,
			CloudProviderBackoffJitter:           1,
			CloudProviderRateLimitConfig: providerazureconfig.CloudProviderRateLimitConfig{
				RateLimitConfig: azclients.RateLimitConfig{
					CloudProviderRateLimit:            true,
					CloudProviderRateLimitBucket:      5,
					CloudProviderRateLimitBucketWrite: 5,
					CloudProviderRateLimitQPS:         1,
					CloudProviderRateLimitQPSWrite:    1,
				},
				InterfaceRateLimit: &azclients.RateLimitConfig{
					CloudProviderRateLimit:            true,
					CloudProviderRateLimitBucket:      5,
					CloudProviderRateLimitBucketWrite: 5,
					CloudProviderRateLimitQPS:         1,
					CloudProviderRateLimitQPSWrite:    1,
				},
				VirtualMachineRateLimit: &azclients.RateLimitConfig{
					CloudProviderRateLimit:            true,
					CloudProviderRateLimitBucket:      5,
					CloudProviderRateLimitBucketWrite: 5,
					CloudProviderRateLimitQPS:         1,
					CloudProviderRateLimitQPSWrite:    1,
				},
				StorageAccountRateLimit: &azclients.RateLimitConfig{
					CloudProviderRateLimit:            true,
					CloudProviderRateLimitBucket:      5,
					CloudProviderRateLimitBucketWrite: 5,
					CloudProviderRateLimitQPS:         1,
					CloudProviderRateLimitQPSWrite:    1,
				},
				DiskRateLimit: &azclients.RateLimitConfig{
					CloudProviderRateLimit:            true,
					CloudProviderRateLimitBucket:      5,
					CloudProviderRateLimitBucketWrite: 5,
					CloudProviderRateLimitQPS:         1,
					CloudProviderRateLimitQPSWrite:    1,
				},
				VirtualMachineScaleSetRateLimit: &azclients.RateLimitConfig{
					CloudProviderRateLimit:            true,
					CloudProviderRateLimitBucket:      5,
					CloudProviderRateLimitBucketWrite: 5,
					CloudProviderRateLimitQPS:         1,
					CloudProviderRateLimitQPSWrite:    1,
				},
			},
		},
		ClusterName:           "mycluster",
		ClusterResourceGroup:  "myrg",
		ARMBaseURLForAPClient: "nodeprovisioner-svc.nodeprovisioner.svc.cluster.local",
		Deployment:            "deployment",
		VmssVmsCacheJitter:    90,
		MaxDeploymentsCount:   8,
		EnableVMsAgentPool:    true,
	}

	t.Setenv("ARM_CLOUD", "AzurePublicCloud")
	t.Setenv("LOCATION", "southeastasia")
	t.Setenv("AZURE_TENANT_ID", "tenantId")
	t.Setenv("AZURE_CLIENT_ID", "aadClientId")
	t.Setenv("ARM_SUBSCRIPTION_ID", "subscriptionId")
	t.Setenv("ARM_RESOURCE_GROUP", "resourceGroup")
	t.Setenv("AZURE_TENANT_ID", "tenantId")
	t.Setenv("ARM_TENANT_ID", "tenantId")
	t.Setenv("AZURE_CLIENT_ID", "aadClientId")
	t.Setenv("ARM_CLIENT_ID", "aadClientId")
	t.Setenv("ARM_CLIENT_SECRET", "aadClientSecret")
	t.Setenv("ARM_VM_TYPE", "vmss")
	t.Setenv("ARM_CLIENT_CERT_PATH", "aadClientCertPath")
	t.Setenv("ARM_CLIENT_CERT_PASSWORD", "aadClientCertPassword")
	t.Setenv("ARM_DEPLOYMENT", "deployment")
	t.Setenv("ARM_USE_MANAGED_IDENTITY_EXTENSION", "true")
	t.Setenv("ARM_USER_ASSIGNED_IDENTITY_ID", "UserAssignedIdentityID")
	t.Setenv("AZURE_VMSS_CACHE_TTL", "100")
	t.Setenv("AZURE_VMSS_VMS_CACHE_TTL", "110")
	t.Setenv("AZURE_VMSS_VMS_CACHE_JITTER", "90")
	t.Setenv("AZURE_MAX_DEPLOYMENT_COUNT", "8")
	t.Setenv("ENABLE_BACKOFF", "true")
	t.Setenv("BACKOFF_RETRIES", "1")
	t.Setenv("BACKOFF_EXPONENT", "1")
	t.Setenv("BACKOFF_DURATION", "1")
	t.Setenv("BACKOFF_JITTER", "1")
	t.Setenv("CLOUD_PROVIDER_RATE_LIMIT", "true")
	t.Setenv("CLUSTER_NAME", "mycluster")
	t.Setenv("ARM_CLUSTER_RESOURCE_GROUP", "myrg")
	t.Setenv("ARM_BASE_URL_FOR_AP_CLIENT", "nodeprovisioner-svc.nodeprovisioner.svc.cluster.local")
	t.Setenv("AZURE_ENABLE_VMS_AGENT_POOLS", "true")

	t.Run("environment variables correctly set", func(t *testing.T) {
		manager, err := createAzureManagerInternal(nil, cloudprovider.NodeGroupDiscoveryOptions{}, mockAzClient)
		assert.NoError(t, err)
		assertStructsMinimallyEqual(t, *expectedConfig, *manager.config)
	})

	t.Run("invalid bool for ARM_USE_MANAGED_IDENTITY_EXTENSION", func(t *testing.T) {
		t.Setenv("ARM_USE_MANAGED_IDENTITY_EXTENSION", "invalidbool")
		manager, err := createAzureManagerInternal(nil, cloudprovider.NodeGroupDiscoveryOptions{}, mockAzClient)
		expectedErr0 := "failed to parse ARM_USE_MANAGED_IDENTITY_EXTENSION \"invalidbool\": strconv.ParseBool: parsing \"invalidbool\": invalid syntax"
		assert.Nil(t, manager)
		assert.Equal(t, expectedErr0, err.Error(), "Return err does not match, expected: %v, actual: %v", expectedErr0, err.Error())
	})

	t.Run("invalid int for AZURE_VMSS_CACHE_TTL", func(t *testing.T) {
		t.Setenv("AZURE_VMSS_CACHE_TTL", "invalidint")
		manager, err := createAzureManagerInternal(nil, cloudprovider.NodeGroupDiscoveryOptions{}, mockAzClient)
		expectedErr := fmt.Errorf("failed to parse AZURE_VMSS_CACHE_TTL \"invalidint\": strconv.ParseInt: parsing \"invalidint\": invalid syntax")
		assert.Nil(t, manager)
		assert.Equal(t, expectedErr, err, "Return err does not match, expected: %v, actual: %v", expectedErr, err)
	})

	t.Run("invalid int for AZURE_GET_VMSS_SIZE_REFRESH_PERIOD", func(t *testing.T) {
		t.Setenv("AZURE_GET_VMSS_SIZE_REFRESH_PERIOD", "invalidint")
		manager, err := createAzureManagerInternal(nil, cloudprovider.NodeGroupDiscoveryOptions{}, mockAzClient)
		expectedErr := fmt.Errorf("failed to parse AZURE_GET_VMSS_SIZE_REFRESH_PERIOD \"invalidint\": strconv.ParseInt: parsing \"invalidint\": invalid syntax")
		assert.Nil(t, manager)
		assert.Equal(t, expectedErr, err, "Return err does not match, expected: %v, actual: %v", expectedErr, err)
	})

	t.Run("invalid int for AZURE_MAX_DEPLOYMENT_COUNT", func(t *testing.T) {
		t.Setenv("AZURE_MAX_DEPLOYMENT_COUNT", "invalidint")
		manager, err := createAzureManagerInternal(nil, cloudprovider.NodeGroupDiscoveryOptions{}, mockAzClient)
		expectedErr := fmt.Errorf("failed to parse AZURE_MAX_DEPLOYMENT_COUNT \"invalidint\": strconv.ParseInt: parsing \"invalidint\": invalid syntax")
		assert.Nil(t, manager)
		assert.Equal(t, expectedErr, err, "Return err does not match, expected: %v, actual: %v", expectedErr, err)
	})

	t.Run("zero AZURE_MAX_DEPLOYMENT_COUNT will use default value", func(t *testing.T) {
		t.Setenv("AZURE_MAX_DEPLOYMENT_COUNT", "0")
		manager, err := createAzureManagerInternal(nil, cloudprovider.NodeGroupDiscoveryOptions{}, mockAzClient)
		assert.NoError(t, err)
		assert.Equal(t, int64(defaultMaxDeploymentsCount), (*manager.config).MaxDeploymentsCount, "MaxDeploymentsCount does not match.")
	})

	t.Run("invalid bool for ENABLE_BACKOFF", func(t *testing.T) {
		t.Setenv("ENABLE_BACKOFF", "invalidbool")
		manager, err := createAzureManagerInternal(nil, cloudprovider.NodeGroupDiscoveryOptions{}, mockAzClient)
		expectedErr := fmt.Errorf("failed to parse ENABLE_BACKOFF \"invalidbool\": strconv.ParseBool: parsing \"invalidbool\": invalid syntax")
		assert.Nil(t, manager)
		assert.Equal(t, expectedErr, err, "Return err does not match, expected: %v, actual: %v", expectedErr, err)
	})

	t.Run("invalid int for BACKOFF_RETRIES", func(t *testing.T) {
		t.Setenv("BACKOFF_RETRIES", "invalidint")
		manager, err := createAzureManagerInternal(nil, cloudprovider.NodeGroupDiscoveryOptions{}, mockAzClient)
		expectedErr := fmt.Errorf("failed to parse BACKOFF_RETRIES \"invalidint\": strconv.ParseInt: parsing \"invalidint\": invalid syntax")
		assert.Nil(t, manager)
		assert.Equal(t, expectedErr, err, "Return err does not match, expected: %v, actual: %v", expectedErr, err)
	})

	t.Run("empty BACKOFF_RETRIES will use default value", func(t *testing.T) {
		t.Setenv("BACKOFF_RETRIES", "")
		manager, err := createAzureManagerInternal(nil, cloudprovider.NodeGroupDiscoveryOptions{}, mockAzClient)
		assert.NoError(t, err)
		assert.Equal(t, providerazureconsts.BackoffRetriesDefault, (*manager.config).CloudProviderBackoffRetries, "CloudProviderBackoffRetries does not match.")
	})

	t.Run("invalid float for BACKOFF_EXPONENT", func(t *testing.T) {
		t.Setenv("BACKOFF_EXPONENT", "invalidfloat")
		manager, err := createAzureManagerInternal(nil, cloudprovider.NodeGroupDiscoveryOptions{}, mockAzClient)
		expectedErr := fmt.Errorf("failed to parse BACKOFF_EXPONENT \"invalidfloat\": strconv.ParseFloat: parsing \"invalidfloat\": invalid syntax")
		assert.Nil(t, manager)
		assert.Equal(t, expectedErr, err, "Return err does not match, expected: %v, actual: %v", expectedErr, err)
	})

	t.Run("empty BACKOFF_EXPONENT will use default value", func(t *testing.T) {
		t.Setenv("BACKOFF_EXPONENT", "")
		manager, err := createAzureManagerInternal(nil, cloudprovider.NodeGroupDiscoveryOptions{}, mockAzClient)
		assert.NoError(t, err)
		assert.Equal(t, providerazureconsts.BackoffExponentDefault, (*manager.config).CloudProviderBackoffExponent, "CloudProviderBackoffExponent does not match.")
	})

	t.Run("invalid int for BACKOFF_DURATION", func(t *testing.T) {
		t.Setenv("BACKOFF_DURATION", "invalidint")
		manager, err := createAzureManagerInternal(nil, cloudprovider.NodeGroupDiscoveryOptions{}, mockAzClient)
		expectedErr := fmt.Errorf("failed to parse BACKOFF_DURATION \"invalidint\": strconv.ParseInt: parsing \"invalidint\": invalid syntax")
		assert.Nil(t, manager)
		assert.Equal(t, expectedErr, err, "Return err does not match, expected: %v, actual: %v", expectedErr, err)
	})

	t.Run("empty BACKOFF_DURATION will use default value", func(t *testing.T) {
		t.Setenv("BACKOFF_DURATION", "")
		manager, err := createAzureManagerInternal(nil, cloudprovider.NodeGroupDiscoveryOptions{}, mockAzClient)
		assert.NoError(t, err)
		assert.Equal(t, providerazureconsts.BackoffDurationDefault, (*manager.config).CloudProviderBackoffDuration, "CloudProviderBackoffDuration does not match.")
	})

	t.Run("invalid float for BACKOFF_JITTER", func(t *testing.T) {
		t.Setenv("BACKOFF_JITTER", "invalidfloat")
		manager, err := createAzureManagerInternal(nil, cloudprovider.NodeGroupDiscoveryOptions{}, mockAzClient)
		expectedErr := fmt.Errorf("failed to parse BACKOFF_JITTER \"invalidfloat\": strconv.ParseFloat: parsing \"invalidfloat\": invalid syntax")
		assert.Nil(t, manager)
		assert.Equal(t, expectedErr, err, "Return err does not match, expected: %v, actual: %v", expectedErr, err)
	})

	t.Run("empty BACKOFF_JITTER will use default value", func(t *testing.T) {
		t.Setenv("BACKOFF_JITTER", "")
		manager, err := createAzureManagerInternal(nil, cloudprovider.NodeGroupDiscoveryOptions{}, mockAzClient)
		assert.NoError(t, err)
		assert.Equal(t, providerazureconsts.BackoffJitterDefault, (*manager.config).CloudProviderBackoffJitter, "CloudProviderBackoffJitter does not match.")
	})

	t.Run("invalid bool for CLOUD_PROVIDER_RATE_LIMIT", func(t *testing.T) {
		t.Setenv("CLOUD_PROVIDER_RATE_LIMIT", "invalidbool")
		manager, err := createAzureManagerInternal(nil, cloudprovider.NodeGroupDiscoveryOptions{}, mockAzClient)
		expectedErr := fmt.Errorf("failed to parse CLOUD_PROVIDER_RATE_LIMIT \"invalidbool\": strconv.ParseBool: parsing \"invalidbool\": invalid syntax")
		assert.Nil(t, manager)
		assert.Equal(t, expectedErr, err, "Return err does not match, expected: %v, actual: %v", expectedErr, err)
	})
}

func TestCreateAzureManagerWithEnvOverridingConfig(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockVMClient := mockvmclient.NewMockInterface(ctrl)
	mockVMSSClient := mockvmssclient.NewMockInterface(ctrl)
	mockVMSSClient.EXPECT().List(gomock.Any(), "resourceGroup").Return([]compute.VirtualMachineScaleSet{}, nil).AnyTimes()
	mockVMClient.EXPECT().List(gomock.Any(), "resourceGroup").Return([]compute.VirtualMachine{}, nil).AnyTimes()
	mockAzClient := &azClient{
		virtualMachinesClient:         mockVMClient,
		virtualMachineScaleSetsClient: mockVMSSClient,
	}

	expectedConfig := &Config{
		Config: providerazure.Config{
			AzureAuthConfig: providerazureconfig.AzureAuthConfig{
				Cloud:                       "AzurePublicCloud",
				TenantID:                    "tenantId",
				AADClientID:                 "aadClientId",
				AADClientSecret:             "aadClientSecret",
				AADClientCertPath:           "aadClientCertPath",
				AADClientCertPassword:       "aadClientCertPassword",
				UseManagedIdentityExtension: true,
				UserAssignedIdentityID:      "UserAssignedIdentityID",
				SubscriptionID:              "subscriptionId",
			},
			Location:                             "southeastasia",
			ResourceGroup:                        "resourceGroup",
			VMType:                               "vmss",
			VmssCacheTTLInSeconds:                100,
			VmssVirtualMachinesCacheTTLInSeconds: 110,
			CloudProviderBackoff:                 true,
			CloudProviderBackoffRetries:          1,
			CloudProviderBackoffExponent:         1,
			CloudProviderBackoffDuration:         1,
			CloudProviderBackoffJitter:           1,
			CloudProviderRateLimitConfig: providerazureconfig.CloudProviderRateLimitConfig{
				RateLimitConfig: azclients.RateLimitConfig{
					CloudProviderRateLimit:            true,
					CloudProviderRateLimitBucket:      5,
					CloudProviderRateLimitBucketWrite: 5,
					CloudProviderRateLimitQPS:         1,
					CloudProviderRateLimitQPSWrite:    1,
				},
				InterfaceRateLimit: &azclients.RateLimitConfig{
					CloudProviderRateLimit:            true,
					CloudProviderRateLimitBucket:      5,
					CloudProviderRateLimitBucketWrite: 5,
					CloudProviderRateLimitQPS:         1,
					CloudProviderRateLimitQPSWrite:    1,
				},
				VirtualMachineRateLimit: &azclients.RateLimitConfig{
					CloudProviderRateLimit:            true,
					CloudProviderRateLimitBucket:      5,
					CloudProviderRateLimitBucketWrite: 5,
					CloudProviderRateLimitQPS:         1,
					CloudProviderRateLimitQPSWrite:    1,
				},
				StorageAccountRateLimit: &azclients.RateLimitConfig{
					CloudProviderRateLimit:            true,
					CloudProviderRateLimitBucket:      5,
					CloudProviderRateLimitBucketWrite: 5,
					CloudProviderRateLimitQPS:         1,
					CloudProviderRateLimitQPSWrite:    1,
				},
				DiskRateLimit: &azclients.RateLimitConfig{
					CloudProviderRateLimit:            true,
					CloudProviderRateLimitBucket:      5,
					CloudProviderRateLimitBucketWrite: 5,
					CloudProviderRateLimitQPS:         1,
					CloudProviderRateLimitQPSWrite:    1,
				},
				VirtualMachineScaleSetRateLimit: &azclients.RateLimitConfig{
					CloudProviderRateLimit:            true,
					CloudProviderRateLimitBucket:      5,
					CloudProviderRateLimitBucketWrite: 5,
					CloudProviderRateLimitQPS:         1,
					CloudProviderRateLimitQPSWrite:    1,
				},
			},
		},
		ClusterName:           "mycluster",
		ClusterResourceGroup:  "myrg",
		ARMBaseURLForAPClient: "nodeprovisioner-svc.nodeprovisioner.svc.cluster.local",
		Deployment:            "deployment",
		VmssVmsCacheJitter:    90,
		MaxDeploymentsCount:   8,
	}

	t.Setenv("ARM_CLOUD", "AzurePublicCloud")
	// LOCATION is not set from env to test getting it from config file
	t.Setenv("AZURE_TENANT_ID", "tenantId")
	t.Setenv("AZURE_CLIENT_ID", "aadClientId")
	t.Setenv("ARM_SUBSCRIPTION_ID", "subscriptionId")
	t.Setenv("ARM_RESOURCE_GROUP", "resourceGroup")
	t.Setenv("AZURE_TENANT_ID", "tenantId")
	t.Setenv("ARM_TENANT_ID", "tenantId")
	t.Setenv("AZURE_CLIENT_ID", "aadClientId")
	t.Setenv("ARM_CLIENT_ID", "aadClientId")
	t.Setenv("ARM_CLIENT_SECRET", "aadClientSecret")
	t.Setenv("ARM_VM_TYPE", "vmss") // this is one of the differences with the config file, expect this to take precedence
	t.Setenv("ARM_CLIENT_CERT_PATH", "aadClientCertPath")
	t.Setenv("ARM_CLIENT_CERT_PASSWORD", "aadClientCertPassword")
	t.Setenv("ARM_DEPLOYMENT", "deployment")
	t.Setenv("ARM_USE_MANAGED_IDENTITY_EXTENSION", "true")
	t.Setenv("ARM_USER_ASSIGNED_IDENTITY_ID", "UserAssignedIdentityID")
	t.Setenv("AZURE_VMSS_CACHE_TTL", "100")
	t.Setenv("AZURE_VMSS_VMS_CACHE_TTL", "110")
	t.Setenv("AZURE_VMSS_VMS_CACHE_JITTER", "90")
	t.Setenv("AZURE_MAX_DEPLOYMENT_COUNT", "8")
	t.Setenv("ENABLE_BACKOFF", "true")
	t.Setenv("BACKOFF_RETRIES", "1")
	t.Setenv("BACKOFF_EXPONENT", "1")
	t.Setenv("BACKOFF_DURATION", "1")
	t.Setenv("BACKOFF_JITTER", "1")
	t.Setenv("CLOUD_PROVIDER_RATE_LIMIT", "true")
	t.Setenv("CLUSTER_NAME", "mycluster")
	t.Setenv("ARM_CLUSTER_RESOURCE_GROUP", "myrg")
	t.Setenv("ARM_BASE_URL_FOR_AP_CLIENT", "nodeprovisioner-svc.nodeprovisioner.svc.cluster.local")

	t.Run("environment variables correctly set", func(t *testing.T) {
		manager, err := createAzureManagerInternal(strings.NewReader(validAzureCfgForStandardVMType), cloudprovider.NodeGroupDiscoveryOptions{}, mockAzClient)
		assert.NoError(t, err)
		assertStructsMinimallyEqual(t, *expectedConfig, *manager.config)
	})
}

func TestCreateAzureManagerInvalidConfig(t *testing.T) {
	_, err := createAzureManagerInternal(strings.NewReader(invalidAzureCfg), cloudprovider.NodeGroupDiscoveryOptions{}, &azClient{})
	assert.Error(t, err, "failed to unmarshal config body")
}

func TestFetchExplicitNodeGroups(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	min, max, name := 1, 15, "test-asg"
	ngdo := cloudprovider.NodeGroupDiscoveryOptions{
		NodeGroupSpecs: []string{
			fmt.Sprintf("%d:%d:%s", min, max, name),
		},
	}

	orchestrationModes := [2]compute.OrchestrationMode{compute.Uniform, compute.Flexible}
	expectedVMSSVMs := newTestVMSSVMList(3)
	expectedVMs := newTestVMList(3)

	for _, orchMode := range orchestrationModes {
		manager := newTestAzureManager(t)
		expectedScaleSets := newTestVMSSList(3, "test-asg", "eastus", compute.Uniform)

		mockVMSSClient := mockvmssclient.NewMockInterface(ctrl)
		mockVMSSClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup).Return(expectedScaleSets, nil).AnyTimes()
		manager.azClient.virtualMachineScaleSetsClient = mockVMSSClient

		if orchMode == compute.Uniform {

			mockVMSSVMClient := mockvmssvmclient.NewMockInterface(ctrl)
			mockVMSSVMClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup, "test-asg", gomock.Any()).Return(expectedVMSSVMs, nil).AnyTimes()
			manager.azClient.virtualMachineScaleSetVMsClient = mockVMSSVMClient
		} else {

			mockVMClient := mockvmclient.NewMockInterface(ctrl)
			manager.config.EnableVmssFlexNodes = true
			mockVMClient.EXPECT().ListVmssFlexVMsWithoutInstanceView(gomock.Any(), "test-asg").Return(expectedVMs, nil).AnyTimes()
			manager.azClient.virtualMachinesClient = mockVMClient
		}

		manager.fetchExplicitNodeGroups(ngdo.NodeGroupSpecs)

		asgs := manager.azureCache.getRegisteredNodeGroups()
		assert.Equal(t, 1, len(asgs))
		assert.Equal(t, name, asgs[0].Id())
		assert.Equal(t, min, asgs[0].MinSize())
		assert.Equal(t, max, asgs[0].MaxSize())
	}

	// test vmTypeStandard
	testAS := newTestAgentPool(newTestAzureManager(t), "testAS")
	timeLayout := "2006-01-02 15:04:05"
	timeBenchMark, _ := time.Parse(timeLayout, "2000-01-01 00:00:00")
	testAS.manager.azClient.deploymentClient = &DeploymentClientMock{
		FakeStore: map[string]resources.DeploymentExtended{
			"cluster-autoscaler-0001": {
				Name: to.StringPtr("cluster-autoscaler-0001"),
				Properties: &resources.DeploymentPropertiesExtended{
					ProvisioningState: to.StringPtr("Succeeded"),
					Timestamp:         &date.Time{Time: timeBenchMark.Add(2 * time.Minute)},
				},
			},
		},
	}
	testAS.manager.config.VMType = providerazureconsts.VMTypeStandard
	err := testAS.manager.fetchExplicitNodeGroups([]string{"1:5:testAS"})
	expectedErr := fmt.Errorf("failed to parse node group spec: %v", retry.NewError(false, fmt.Errorf("deployment not found")).Error())
	assert.Equal(t, expectedErr, err, "testAS.manager.fetchExplicitNodeGroups return error does not match, expected: %v, actual: %v", expectedErr, err)
	err = testAS.manager.fetchExplicitNodeGroups(nil)
	assert.NoError(t, err)

	// test invalidVMType
	manager := newTestAzureManager(t)
	manager.config.VMType = "invalidVMType"
	err = manager.fetchExplicitNodeGroups(ngdo.NodeGroupSpecs)
	expectedErr = fmt.Errorf("failed to parse node group spec: vmtype invalidVMType not supported")
	assert.Equal(t, expectedErr, err, "manager.fetchExplicitNodeGroups return error does not match, expected: %v, actual: %v", expectedErr, err)
}

func TestGetFilteredAutoscalingGroupsVmss(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	vmssName := "test-vmss"
	vmssTag := "fake-tag"
	vmssTagValue := "fake-value"
	min := "1"
	minVal := 1
	max := "5"
	maxVal := 5

	ngdo := cloudprovider.NodeGroupDiscoveryOptions{
		NodeGroupAutoDiscoverySpecs: []string{fmt.Sprintf("label:%s=%s", vmssTag, vmssTagValue)},
	}

	manager := newTestAzureManager(t)
	expectedScaleSets := []compute.VirtualMachineScaleSet{fakeVMSSWithTags(vmssName, map[string]*string{vmssTag: &vmssTagValue, "min": &min, "max": &max})}
	mockVMSSClient := mockvmssclient.NewMockInterface(ctrl)
	mockVMSSClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup).Return(expectedScaleSets, nil).AnyTimes()
	manager.azClient.virtualMachineScaleSetsClient = mockVMSSClient
	err := manager.forceRefresh()
	assert.NoError(t, err)

	specs, err := ParseLabelAutoDiscoverySpecs(ngdo)
	assert.NoError(t, err)

	asgs, err := manager.getFilteredNodeGroups(specs)
	assert.NoError(t, err)
	expectedAsgs := []cloudprovider.NodeGroup{&ScaleSet{
		azureRef: azureRef{
			Name: vmssName,
		},
		minSize:                  minVal,
		maxSize:                  maxVal,
		manager:                  manager,
		enableForceDelete:        manager.config.EnableForceDelete,
		curSize:                  3,
		sizeRefreshPeriod:        manager.azureCache.refreshInterval,
		getVmssSizeRefreshPeriod: time.Duration(manager.azureCache.refreshInterval) * time.Second,
		InstanceCache:            InstanceCache{instancesRefreshPeriod: defaultVmssInstancesRefreshPeriod},
	}}
	assert.True(t, assert.ObjectsAreEqualValues(expectedAsgs, asgs), "expected %#v, but found: %#v", expectedAsgs, asgs)
}

func TestGetFilteredAutoscalingGroupsWithInvalidVMType(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ngdo := cloudprovider.NodeGroupDiscoveryOptions{
		NodeGroupAutoDiscoverySpecs: []string{"label:fake-tag=fake-value"},
	}

	manager := newTestAzureManager(t)
	expectedScaleSets := []compute.VirtualMachineScaleSet{}
	mockVMSSClient := mockvmssclient.NewMockInterface(ctrl)
	mockVMSSClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup).Return(expectedScaleSets, nil).AnyTimes()
	manager.azClient.virtualMachineScaleSetsClient = mockVMSSClient
	manager.config.VMType = vmTypeAKS

	specs, err := ParseLabelAutoDiscoverySpecs(ngdo)
	assert.NoError(t, err)

	expectedErr := fmt.Errorf("vmType \"aks\" does not support autodiscovery")
	asgs, err2 := manager.getFilteredNodeGroups(specs)
	assert.Nil(t, asgs)
	assert.Equal(t, expectedErr, err2, "Not match, expected: %v, actual: %v", expectedErr, err2)

	manager.config.VMType = "invalidVMType"
	expectedErr = fmt.Errorf("vmType \"invalidVMType\" does not support autodiscovery")
	asgs, err2 = manager.getFilteredNodeGroups(specs)
	assert.Nil(t, asgs)
	assert.Equal(t, expectedErr, err2, "Not match, expected: %v, actual: %v", expectedErr, err2)
}

func TestFetchAutoAsgsVmss(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	vmssName := "test-vmss"
	vmssTag := "fake-tag"
	vmssTagValue := "fake-value"
	minString := "1"
	minVal := 1
	maxString := "5"
	maxVal := 5

	ngdo := cloudprovider.NodeGroupDiscoveryOptions{
		NodeGroupAutoDiscoverySpecs: []string{fmt.Sprintf("label:%s=%s", vmssTag, vmssTagValue)},
	}

	expectedScaleSets := []compute.VirtualMachineScaleSet{fakeVMSSWithTags(vmssName, map[string]*string{vmssTag: &vmssTagValue, "min": &minString, "max": &maxString})}
	expectedVMSSVMs := newTestVMSSVMList(1)

	manager := newTestAzureManager(t)
	mockVMSSClient := mockvmssclient.NewMockInterface(ctrl)
	mockVMSSClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup).Return(expectedScaleSets, nil).AnyTimes()
	manager.azClient.virtualMachineScaleSetsClient = mockVMSSClient
	mockVMSSVMClient := mockvmssvmclient.NewMockInterface(ctrl)
	mockVMSSVMClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup, vmssName, gomock.Any()).Return(expectedVMSSVMs, nil).AnyTimes()
	manager.azClient.virtualMachineScaleSetVMsClient = mockVMSSVMClient
	mockVMClient := mockvmclient.NewMockInterface(ctrl)
	manager.azClient.virtualMachinesClient = mockVMClient
	mockVMClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup).Return([]compute.VirtualMachine{}, nil).AnyTimes()
	err := manager.forceRefresh()
	assert.NoError(t, err)

	specs, err := ParseLabelAutoDiscoverySpecs(ngdo)
	assert.NoError(t, err)
	manager.autoDiscoverySpecs = specs

	// assert cache is empty before fetching auto asgs
	asgs := manager.azureCache.getRegisteredNodeGroups()
	assert.Equal(t, 0, len(asgs))

	manager.fetchAutoNodeGroups()
	asgs = manager.azureCache.getRegisteredNodeGroups()
	assert.Equal(t, 1, len(asgs))
	assert.Equal(t, vmssName, asgs[0].Id())
	assert.Equal(t, minVal, asgs[0].MinSize())
	assert.Equal(t, maxVal, asgs[0].MaxSize())

	// test explicitlyConfigured
	manager.explicitlyConfigured[vmssName] = true
	manager.fetchAutoNodeGroups()
	asgs = manager.azureCache.getRegisteredNodeGroups()
	assert.Equal(t, 1, len(asgs))
}

func TestManagerRefreshAndCleanup(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	manager := newTestAzureManager(t)
	err := manager.Refresh()
	assert.NoError(t, err)
	manager.Cleanup()
}

func TestGetScaleSetOptions(t *testing.T) {
	manager := &AzureManager{
		azureCache: &azureCache{
			autoscalingOptions: make(map[azureRef]map[string]string),
		},
	}
	defaultOptions := config.NodeGroupAutoscalingOptions{
		ScaleDownUtilizationThreshold:    0.1,
		ScaleDownGpuUtilizationThreshold: 0.2,
		ScaleDownUnneededTime:            time.Second,
		ScaleDownUnreadyTime:             time.Minute,
	}

	tags := map[string]string{
		config.DefaultScaleDownUtilizationThresholdKey:    "0.2",
		config.DefaultScaleDownGpuUtilizationThresholdKey: "0.3",
		config.DefaultScaleDownUnneededTimeKey:            "30m",
		config.DefaultScaleDownUnreadyTimeKey:             "1h",
	}
	manager.azureCache.autoscalingOptions[azureRef{Name: "test1"}] = tags
	opts := manager.GetScaleSetOptions("test1", defaultOptions)
	assert.Equal(t, opts.ScaleDownUtilizationThreshold, 0.2)
	assert.Equal(t, opts.ScaleDownGpuUtilizationThreshold, 0.3)
	assert.Equal(t, opts.ScaleDownUnneededTime, 30*time.Minute)
	assert.Equal(t, opts.ScaleDownUnreadyTime, time.Hour)

	tags = map[string]string{
		//config.DefaultScaleDownUtilizationThresholdKey: ... // not specified (-> default)
		config.DefaultScaleDownGpuUtilizationThresholdKey: "not-a-float",
		config.DefaultScaleDownUnneededTimeKey:            "1m",
		config.DefaultScaleDownUnreadyTimeKey:             "not-a-duration",
	}
	manager.azureCache.autoscalingOptions[azureRef{Name: "test2"}] = tags
	opts = manager.GetScaleSetOptions("test2", defaultOptions)
	assert.Equal(t, opts.ScaleDownUtilizationThreshold, defaultOptions.ScaleDownUtilizationThreshold)
	assert.Equal(t, opts.ScaleDownGpuUtilizationThreshold, defaultOptions.ScaleDownGpuUtilizationThreshold)
	assert.Equal(t, opts.ScaleDownUnneededTime, time.Minute)
	assert.Equal(t, opts.ScaleDownUnreadyTime, defaultOptions.ScaleDownUnreadyTime)

	manager.azureCache.autoscalingOptions[azureRef{Name: "test3"}] = map[string]string{}
	opts = manager.GetScaleSetOptions("test3", defaultOptions)
	assert.Equal(t, *opts, defaultOptions)
}

func assertStructsMinimallyEqual(t *testing.T, struct1, struct2 interface{}) bool {
	return compareStructFields(t, reflect.ValueOf(struct1), reflect.ValueOf(struct2))
}

func compareStructFields(t *testing.T, v1, v2 reflect.Value) bool {
	if v1.Type() != v2.Type() {
		return assert.Fail(t, "different types", "v1 type: %v, v2 type: %v", v1.Type(), v2.Type())
	}

	for i := 0; i < v1.NumField(); i++ {
		field1 := v1.Field(i)
		field2 := v2.Field(i)
		fieldType := v1.Type().Field(i)

		if field1.IsZero() || reflect.DeepEqual(field1.Interface(), reflect.Zero(field1.Type()).Interface()) {
			continue // Skip zero value fields in struct1
		}

		if field1.Kind() == reflect.Struct {
			// Recursively compare nested structs
			if !compareStructFields(t, field1, field2) {
				return false
			}
		} else {
			if !assert.Equal(t, field1.Interface(), field2.Interface(), "field %s", fieldType.Name) {
				return false
			}
		}
	}

	return true
}
