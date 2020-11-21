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
	"k8s.io/legacy-cloud-providers/azure/clients/vmclient/mockvmclient"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-12-01/compute"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2017-05-10/resources"
	"github.com/Azure/go-autorest/autorest/date"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	azclients "k8s.io/legacy-cloud-providers/azure/clients"
	"k8s.io/legacy-cloud-providers/azure/clients/vmssclient/mockvmssclient"
	"k8s.io/legacy-cloud-providers/azure/clients/vmssvmclient/mockvmssvmclient"
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
	"vmssCacheTTL": 60,
	"vmssVmsCacheTTL": 240,
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
        "vmssCacheTTL": 60,
	"vmssVmsCacheTTL": 240,
	"vmssVmsCacheJitter": 120,
        "maxDeploymentsCount": 8,
        "cloudProviderRateLimit": false,
        "routeRateLimit": {
                "cloudProviderRateLimit": true,
                "cloudProviderRateLimitQPS": 3
        },
        "deployment":"cluster-autoscaler-0001"
}`

const invalidAzureCfg = `{{}"cloud": "AzurePublicCloud",}`

func TestCreateAzureManagerValidConfig(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockVMClient := mockvmclient.NewMockInterface(ctrl)
	mockVMSSClient := mockvmssclient.NewMockInterface(ctrl)
	mockVMSSClient.EXPECT().List(gomock.Any(), "fakeId").Return([]compute.VirtualMachineScaleSet{}, nil).Times(2)
	mockAzClient := &azClient{
		virtualMachinesClient:         mockVMClient,
		virtualMachineScaleSetsClient: mockVMSSClient,
	}
	manager, err := createAzureManagerInternal(strings.NewReader(validAzureCfg), cloudprovider.NodeGroupDiscoveryOptions{}, mockAzClient)

	expectedConfig := &Config{
		Cloud:               "AzurePublicCloud",
		Location:            "southeastasia",
		TenantID:            "fakeId",
		SubscriptionID:      "fakeId",
		ResourceGroup:       "fakeId",
		VMType:              "vmss",
		AADClientID:         "fakeId",
		AADClientSecret:     "fakeId",
		VmssCacheTTL:        60,
		VmssVmsCacheTTL:     240,
		VmssVmsCacheJitter:  120,
		MaxDeploymentsCount: 8,
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

	assert.NoError(t, err)
	assert.Equal(t, true, reflect.DeepEqual(*expectedConfig, *manager.config), "unexpected azure manager configuration")
}

func TestCreateAzureManagerValidConfigForStandardVMType(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockVMClient := mockvmclient.NewMockInterface(ctrl)
	mockVMClient.EXPECT().List(gomock.Any(), "fakeId").Return([]compute.VirtualMachine{}, nil).Times(2)
	mockVMSSClient := mockvmssclient.NewMockInterface(ctrl)
	mockAzClient := &azClient{
		virtualMachinesClient:         mockVMClient,
		virtualMachineScaleSetsClient: mockVMSSClient,
	}
	manager, err := createAzureManagerInternal(strings.NewReader(validAzureCfgForStandardVMType), cloudprovider.NodeGroupDiscoveryOptions{}, mockAzClient)

	expectedConfig := &Config{
		Cloud:               "AzurePublicCloud",
		Location:            "southeastasia",
		TenantID:            "fakeId",
		SubscriptionID:      "fakeId",
		ResourceGroup:       "fakeId",
		VMType:              "standard",
		AADClientID:         "fakeId",
		AADClientSecret:     "fakeId",
		VmssCacheTTL:        60,
		VmssVmsCacheTTL:     240,
		VmssVmsCacheJitter:  120,
		MaxDeploymentsCount: 8,
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
		Deployment: "cluster-autoscaler-0001",
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
	assert.Equal(t, *expectedConfig, *manager.config, "unexpected azure manager configuration, expected: %v, actual: %v", *expectedConfig, *manager.config)
}

func TestCreateAzureManagerValidConfigForStandardVMTypeWithoutDeploymentParameters(t *testing.T) {
	manager, err := createAzureManagerInternal(strings.NewReader(validAzureCfgForStandardVMTypeWithoutDeploymentParameters), cloudprovider.NodeGroupDiscoveryOptions{}, &azClient{})
	expectedErr := "open /var/lib/azure/azuredeploy.parameters.json: no such file or directory"
	assert.Nil(t, manager)
	assert.Equal(t, expectedErr, err.Error(), "return error does not match, expected: %v, actual: %v", expectedErr, err.Error())
}

func TestCreateAzureManagerWithNilConfig(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockVMClient := mockvmclient.NewMockInterface(ctrl)
	mockVMSSClient := mockvmssclient.NewMockInterface(ctrl)
	mockVMSSClient.EXPECT().List(gomock.Any(), "resourceGroup").Return([]compute.VirtualMachineScaleSet{}, nil).AnyTimes()
	mockAzClient := &azClient{
		virtualMachinesClient:         mockVMClient,
		virtualMachineScaleSetsClient: mockVMSSClient,
	}

	expectedConfig := &Config{
		Cloud:                        "AzurePublicCloud",
		Location:                     "southeastasia",
		TenantID:                     "tenantId",
		SubscriptionID:               "subscriptionId",
		ResourceGroup:                "resourceGroup",
		VMType:                       "vmss",
		AADClientID:                  "aadClientId",
		AADClientSecret:              "aadClientSecret",
		AADClientCertPath:            "aadClientCertPath",
		AADClientCertPassword:        "aadClientCertPassword",
		Deployment:                   "deployment",
		ClusterName:                  "clusterName",
		NodeResourceGroup:            "resourcegroup",
		UseManagedIdentityExtension:  true,
		UserAssignedIdentityID:       "UserAssignedIdentityID",
		VmssCacheTTL:                 100,
		VmssVmsCacheTTL:              110,
		VmssVmsCacheJitter:           90,
		MaxDeploymentsCount:          8,
		CloudProviderBackoff:         true,
		CloudProviderBackoffRetries:  1,
		CloudProviderBackoffExponent: 1,
		CloudProviderBackoffDuration: 1,
		CloudProviderBackoffJitter:   1,
		CloudProviderRateLimitConfig: CloudProviderRateLimitConfig{
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
			KubernetesServiceRateLimit: &azclients.RateLimitConfig{
				CloudProviderRateLimit:            true,
				CloudProviderRateLimitBucket:      5,
				CloudProviderRateLimitBucketWrite: 5,
				CloudProviderRateLimitQPS:         1,
				CloudProviderRateLimitQPSWrite:    1,
			},
		},
	}

	os.Setenv("ARM_CLOUD", "AzurePublicCloud")
	os.Setenv("LOCATION", "southeastasia")
	os.Setenv("ARM_SUBSCRIPTION_ID", "subscriptionId")
	os.Setenv("ARM_RESOURCE_GROUP", "resourceGroup")
	os.Setenv("ARM_TENANT_ID", "tenantId")
	os.Setenv("ARM_CLIENT_ID", "aadClientId")
	os.Setenv("ARM_CLIENT_SECRET", "aadClientSecret")
	os.Setenv("ARM_VM_TYPE", "vmss")
	os.Setenv("ARM_CLIENT_CERT_PATH", "aadClientCertPath")
	os.Setenv("ARM_CLIENT_CERT_PASSWORD", "aadClientCertPassword")
	os.Setenv("ARM_DEPLOYMENT", "deployment")
	os.Setenv("AZURE_CLUSTER_NAME", "clusterName")
	os.Setenv("AZURE_NODE_RESOURCE_GROUP", "resourcegroup")
	os.Setenv("ARM_USE_MANAGED_IDENTITY_EXTENSION", "true")
	os.Setenv("ARM_USER_ASSIGNED_IDENTITY_ID", "UserAssignedIdentityID")
	os.Setenv("AZURE_VMSS_CACHE_TTL", "100")
	os.Setenv("AZURE_VMSS_VMS_CACHE_TTL", "110")
	os.Setenv("AZURE_VMSS_VMS_CACHE_JITTER", "90")
	os.Setenv("AZURE_MAX_DEPLOYMENT_COUNT", "8")
	os.Setenv("ENABLE_BACKOFF", "true")
	os.Setenv("BACKOFF_RETRIES", "1")
	os.Setenv("BACKOFF_EXPONENT", "1")
	os.Setenv("BACKOFF_DURATION", "1")
	os.Setenv("BACKOFF_JITTER", "1")
	os.Setenv("CLOUD_PROVIDER_RATE_LIMIT", "true")

	manager, err := createAzureManagerInternal(nil, cloudprovider.NodeGroupDiscoveryOptions{}, mockAzClient)
	assert.NoError(t, err)
	assert.Equal(t, true, reflect.DeepEqual(*expectedConfig, *manager.config), "unexpected azure manager configuration")

	// invalid bool for ARM_USE_MANAGED_IDENTITY_EXTENSION
	os.Setenv("ARM_USE_MANAGED_IDENTITY_EXTENSION", "invalidbool")
	manager, err = createAzureManagerInternal(nil, cloudprovider.NodeGroupDiscoveryOptions{}, mockAzClient)
	expectedErr0 := "strconv.ParseBool: parsing \"invalidbool\": invalid syntax"
	assert.Nil(t, manager)
	assert.Equal(t, expectedErr0, err.Error(), "Return err does not match, expected: %v, actual: %v", expectedErr0, err.Error())
	// revert back to good ARM_USE_MANAGED_IDENTITY_EXTENSION
	os.Setenv("ARM_USE_MANAGED_IDENTITY_EXTENSION", "true")

	// invalid int for AZURE_VMSS_CACHE_TTL
	os.Setenv("AZURE_VMSS_CACHE_TTL", "invalidint")
	manager, err = createAzureManagerInternal(nil, cloudprovider.NodeGroupDiscoveryOptions{}, mockAzClient)
	expectedErr := fmt.Errorf("failed to parse AZURE_VMSS_CACHE_TTL \"invalidint\": strconv.ParseInt: parsing \"invalidint\": invalid syntax")
	assert.Nil(t, manager)
	assert.Equal(t, expectedErr, err, "Return err does not match, expected: %v, actual: %v", expectedErr, err)
	// revert back to good AZURE_VMSS_CACHE_TTL
	os.Setenv("AZURE_VMSS_CACHE_TTL", "100")

	// invalid int for AZURE_MAX_DEPLOYMENT_COUNT
	os.Setenv("AZURE_MAX_DEPLOYMENT_COUNT", "invalidint")
	manager, err = createAzureManagerInternal(nil, cloudprovider.NodeGroupDiscoveryOptions{}, mockAzClient)
	expectedErr = fmt.Errorf("failed to parse AZURE_MAX_DEPLOYMENT_COUNT \"invalidint\": strconv.ParseInt: parsing \"invalidint\": invalid syntax")
	assert.Nil(t, manager)
	assert.Equal(t, expectedErr, err, "Return err does not match, expected: %v, actual: %v", expectedErr, err)
	// revert back to good AZURE_MAX_DEPLOYMENT_COUNT
	os.Setenv("AZURE_MAX_DEPLOYMENT_COUNT", "8")

	// zero AZURE_MAX_DEPLOYMENT_COUNT will use default value
	os.Setenv("AZURE_MAX_DEPLOYMENT_COUNT", "0")
	manager, err = createAzureManagerInternal(nil, cloudprovider.NodeGroupDiscoveryOptions{}, mockAzClient)
	assert.NoError(t, err)
	assert.Equal(t, int64(defaultMaxDeploymentsCount), (*manager.config).MaxDeploymentsCount, "MaxDeploymentsCount does not match.")
	// revert back to good AZURE_MAX_DEPLOYMENT_COUNT
	os.Setenv("AZURE_MAX_DEPLOYMENT_COUNT", "8")

	// invalid bool for ENABLE_BACKOFF
	os.Setenv("ENABLE_BACKOFF", "invalidbool")
	manager, err = createAzureManagerInternal(nil, cloudprovider.NodeGroupDiscoveryOptions{}, mockAzClient)
	expectedErr = fmt.Errorf("failed to parse ENABLE_BACKOFF \"invalidbool\": strconv.ParseBool: parsing \"invalidbool\": invalid syntax")
	assert.Nil(t, manager)
	assert.Equal(t, expectedErr, err, "Return err does not match, expected: %v, actual: %v", expectedErr, err)
	// revert back to good ENABLE_BACKOFF
	os.Setenv("ENABLE_BACKOFF", "true")

	// invalid int for BACKOFF_RETRIES
	os.Setenv("BACKOFF_RETRIES", "invalidint")
	manager, err = createAzureManagerInternal(nil, cloudprovider.NodeGroupDiscoveryOptions{}, mockAzClient)
	expectedErr = fmt.Errorf("failed to parse BACKOFF_RETRIES '\\x00': strconv.ParseInt: parsing \"invalidint\": invalid syntax")
	assert.Nil(t, manager)
	assert.Equal(t, expectedErr, err, "Return err does not match, expected: %v, actual: %v", expectedErr, err)
	// revert back to good BACKOFF_RETRIES
	os.Setenv("BACKOFF_RETRIES", "1")

	// empty BACKOFF_RETRIES will use default value
	os.Setenv("BACKOFF_RETRIES", "")
	manager, err = createAzureManagerInternal(nil, cloudprovider.NodeGroupDiscoveryOptions{}, mockAzClient)
	assert.NoError(t, err)
	assert.Equal(t, backoffRetriesDefault, (*manager.config).CloudProviderBackoffRetries, "CloudProviderBackoffRetries does not match.")
	// revert back to good BACKOFF_RETRIES
	os.Setenv("BACKOFF_RETRIES", "1")

	// invalid float for BACKOFF_EXPONENT
	os.Setenv("BACKOFF_EXPONENT", "invalidfloat")
	manager, err = createAzureManagerInternal(nil, cloudprovider.NodeGroupDiscoveryOptions{}, mockAzClient)
	expectedErr = fmt.Errorf("failed to parse BACKOFF_EXPONENT \"invalidfloat\": strconv.ParseFloat: parsing \"invalidfloat\": invalid syntax")
	assert.Nil(t, manager)
	assert.Equal(t, expectedErr, err, "Return err does not match, expected: %v, actual: %v", expectedErr, err)
	// revert back to good BACKOFF_EXPONENT
	os.Setenv("BACKOFF_EXPONENT", "1")

	// empty BACKOFF_EXPONENT will use default value
	os.Setenv("BACKOFF_EXPONENT", "")
	manager, err = createAzureManagerInternal(nil, cloudprovider.NodeGroupDiscoveryOptions{}, mockAzClient)
	assert.NoError(t, err)
	assert.Equal(t, backoffExponentDefault, (*manager.config).CloudProviderBackoffExponent, "CloudProviderBackoffExponent does not match.")
	// revert back to good BACKOFF_EXPONENT
	os.Setenv("BACKOFF_EXPONENT", "1")

	// invalid int for BACKOFF_DURATION
	os.Setenv("BACKOFF_DURATION", "invalidint")
	manager, err = createAzureManagerInternal(nil, cloudprovider.NodeGroupDiscoveryOptions{}, mockAzClient)
	expectedErr = fmt.Errorf("failed to parse BACKOFF_DURATION \"invalidint\": strconv.ParseInt: parsing \"invalidint\": invalid syntax")
	assert.Nil(t, manager)
	assert.Equal(t, expectedErr, err, "Return err does not match, expected: %v, actual: %v", expectedErr, err)
	// revert back to good BACKOFF_DURATION
	os.Setenv("BACKOFF_DURATION", "1")

	// empty BACKOFF_DURATION will use default value
	os.Setenv("BACKOFF_DURATION", "")
	manager, err = createAzureManagerInternal(nil, cloudprovider.NodeGroupDiscoveryOptions{}, mockAzClient)
	assert.NoError(t, err)
	assert.Equal(t, backoffDurationDefault, (*manager.config).CloudProviderBackoffDuration, "CloudProviderBackoffDuration does not match.")
	// revert back to good BACKOFF_DURATION
	os.Setenv("BACKOFF_DURATION", "1")

	// invalid float for BACKOFF_JITTER
	os.Setenv("BACKOFF_JITTER", "invalidfloat")
	manager, err = createAzureManagerInternal(nil, cloudprovider.NodeGroupDiscoveryOptions{}, mockAzClient)
	expectedErr = fmt.Errorf("failed to parse BACKOFF_JITTER \"invalidfloat\": strconv.ParseFloat: parsing \"invalidfloat\": invalid syntax")
	assert.Nil(t, manager)
	assert.Equal(t, expectedErr, err, "Return err does not match, expected: %v, actual: %v", expectedErr, err)
	// revert back to good BACKOFF_JITTER
	os.Setenv("BACKOFF_JITTER", "1")

	// empty BACKOFF_JITTER will use default value
	os.Setenv("BACKOFF_JITTER", "")
	manager, err = createAzureManagerInternal(nil, cloudprovider.NodeGroupDiscoveryOptions{}, mockAzClient)
	assert.NoError(t, err)
	assert.Equal(t, backoffJitterDefault, (*manager.config).CloudProviderBackoffJitter, "CloudProviderBackoffJitter does not match.")
	// revert back to good BACKOFF_JITTER
	os.Setenv("BACKOFF_JITTER", "1")

	// invalid bool for CLOUD_PROVIDER_RATE_LIMIT
	os.Setenv("CLOUD_PROVIDER_RATE_LIMIT", "invalidbool")
	manager, err = createAzureManagerInternal(nil, cloudprovider.NodeGroupDiscoveryOptions{}, mockAzClient)
	expectedErr = fmt.Errorf("failed to parse CLOUD_PROVIDER_RATE_LIMIT: \"invalidbool\", strconv.ParseBool: parsing \"invalidbool\": invalid syntax")
	assert.Nil(t, manager)
	assert.Equal(t, expectedErr, err, "Return err does not match, expected: %v, actual: %v", expectedErr, err)
	// revert back to good CLOUD_PROVIDER_RATE_LIMIT
	os.Setenv("CLOUD_PROVIDER_RATE_LIMIT", "1")

	os.Unsetenv("ARM_CLOUD")
	os.Unsetenv("ARM_SUBSCRIPTION_ID")
	os.Unsetenv("LOCATION")
	os.Unsetenv("ARM_RESOURCE_GROUP")
	os.Unsetenv("ARM_TENANT_ID")
	os.Unsetenv("ARM_CLIENT_ID")
	os.Unsetenv("ARM_CLIENT_SECRET")
	os.Unsetenv("ARM_VM_TYPE")
	os.Unsetenv("ARM_CLIENT_CERT_PATH")
	os.Unsetenv("ARM_CLIENT_CERT_PASSWORD")
	os.Unsetenv("ARM_DEPLOYMENT")
	os.Unsetenv("AZURE_CLUSTER_NAME")
	os.Unsetenv("AZURE_NODE_RESOURCE_GROUP")
	os.Unsetenv("ARM_USE_MANAGED_IDENTITY_EXTENSION")
	os.Unsetenv("ARM_USER_ASSIGNED_IDENTITY_ID")
	os.Unsetenv("AZURE_VMSS_CACHE_TTL")
	os.Unsetenv("AZURE_MAX_DEPLOYMENT_COUNT")
	os.Unsetenv("ENABLE_BACKOFF")
	os.Unsetenv("BACKOFF_RETRIES")
	os.Unsetenv("BACKOFF_EXPONENT")
	os.Unsetenv("BACKOFF_DURATION")
	os.Unsetenv("BACKOFF_JITTER")
	os.Unsetenv("CLOUD_PROVIDER_RATE_LIMIT")
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

	manager := newTestAzureManager(t)
	expectedVMSSVMs := newTestVMSSVMList(3)
	expectedScaleSets := newTestVMSSList(3, "test-asg", "eastus")

	mockVMSSClient := mockvmssclient.NewMockInterface(ctrl)
	mockVMSSClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup).Return(expectedScaleSets, nil).AnyTimes()
	manager.azClient.virtualMachineScaleSetsClient = mockVMSSClient
	mockVMSSVMClient := mockvmssvmclient.NewMockInterface(ctrl)
	mockVMSSVMClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup, "test-asg", gomock.Any()).Return(expectedVMSSVMs, nil).AnyTimes()
	manager.azClient.virtualMachineScaleSetVMsClient = mockVMSSVMClient
	manager.fetchExplicitNodeGroups(ngdo.NodeGroupSpecs)

	asgs := manager.azureCache.getRegisteredNodeGroups()
	assert.Equal(t, 1, len(asgs))
	assert.Equal(t, name, asgs[0].Id())
	assert.Equal(t, min, asgs[0].MinSize())
	assert.Equal(t, max, asgs[0].MaxSize())

	// test vmTypeStandard
	testAS := newTestAgentPool(newTestAzureManager(t), "testAS")
	timeLayout := "2006-01-02 15:04:05"
	timeBenchMark, _ := time.Parse(timeLayout, "2000-01-01 00:00:00")
	testAS.manager.azClient.deploymentsClient = &DeploymentsClientMock{
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
	testAS.manager.config.VMType = vmTypeStandard
	err := testAS.manager.fetchExplicitNodeGroups([]string{"1:5:testAS"})
	expectedErr := fmt.Errorf("failed to parse node group spec: deployment not found")
	assert.Equal(t, expectedErr, err, "testAS.manager.fetchExplicitNodeGroups return error does not match, expected: %v, actual: %v", expectedErr, err)
	err = testAS.manager.fetchExplicitNodeGroups(nil)
	assert.NoError(t, err)

	// test invalidVMType
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
		minSize:                minVal,
		maxSize:                maxVal,
		manager:                manager,
		curSize:                3,
		instancesRefreshPeriod: defaultVmssInstancesRefreshPeriod,
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
