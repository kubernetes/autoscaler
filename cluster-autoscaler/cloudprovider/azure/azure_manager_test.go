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
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	azclients "k8s.io/legacy-cloud-providers/azure/clients"
	"k8s.io/legacy-cloud-providers/azure/clients/vmssclient/mockvmssclient"
	"k8s.io/legacy-cloud-providers/azure/clients/vmssvmclient/mockvmssvmclient"
	"k8s.io/legacy-cloud-providers/azure/retry"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-12-01/compute"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2017-05-10/resources"
	"github.com/Azure/go-autorest/autorest/date"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
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
	manager, err := CreateAzureManager(strings.NewReader(validAzureCfg), cloudprovider.NodeGroupDiscoveryOptions{})

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
		},
	}

	assert.NoError(t, err)
	assert.Equal(t, true, reflect.DeepEqual(*expectedConfig, *manager.config), "unexpected azure manager configuration")
}

func TestCreateAzureManagerValidConfigForStandardVMType(t *testing.T) {
	manager, err := CreateAzureManager(strings.NewReader(validAzureCfgForStandardVMType), cloudprovider.NodeGroupDiscoveryOptions{})
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
	discoveryOpts := cloudprovider.NodeGroupDiscoveryOptions{NodeGroupAutoDiscoverySpecs: []string{
		"label:cluster-autoscaler-enabled=true,cluster-autoscaler-name=fake-cluster",
		"label:test-tag=test-value,another-test-tag=another-test-value",
	}}
	timeLayout := "2006-01-02 15:04:05"
	timeBenchMark, _ := time.Parse(timeLayout, "2000-01-01 00:00:00")
	fakeDeployments := map[string]resources.DeploymentExtended{
		"cluster-autoscaler-0001": {
			Name: to.StringPtr("cluster-autoscaler-0001"),
			Properties: &resources.DeploymentPropertiesExtended{
				ProvisioningState: to.StringPtr("Succeeded"),
				Parameters: map[string]interface{}{
					"PoolName01VMSize": to.StringPtr("PoolName01"),
				},
				Template: map[string]interface{}{
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
				Timestamp: &date.Time{Time: timeBenchMark},
			},
		},
	}
	manager.azClient.deploymentsClient = &DeploymentsClientMock{
		FakeStore: fakeDeployments,
	}
	specs, err2 := parseLabelAutoDiscoverySpecs(discoveryOpts)
	assert.NoError(t, err2)
	result, err3 := manager.getFilteredAutoscalingGroups(specs)
	expectedNodeGroup := []cloudprovider.NodeGroup{(*AgentPool)(nil)}
	assert.NoError(t, err3)
	assert.Equal(t, expectedNodeGroup, result, "NodeGroup does not match, expected: %v, actual: %v", expectedNodeGroup, result)

	// parseLabelAutoDiscoverySpecs with invalid NodeGroupDiscoveryOptions
	invalidDiscoveryOpts := cloudprovider.NodeGroupDiscoveryOptions{NodeGroupAutoDiscoverySpecs: []string{"label:keywithoutvalue"}}
	specs, err4 := parseLabelAutoDiscoverySpecs(invalidDiscoveryOpts)
	expectedCfg := []labelAutoDiscoveryConfig([]labelAutoDiscoveryConfig(nil))
	expectedErr := fmt.Errorf("invalid key=value pair [keywithoutvalue]")
	assert.Equal(t, expectedCfg, specs, "Return labelAutoDiscoveryConfig does not match, expected: %v, actual: %v", expectedCfg, specs)
	assert.Equal(t, expectedErr, err4, "parseLabelAutoDiscoverySpecs return error does not match, expected: %v, actual: %v", expectedErr, err4)
}

func TestCreateAzureManagerValidConfigForStandardVMTypeWithoutDeploymentParameters(t *testing.T) {
	manager, err := CreateAzureManager(strings.NewReader(validAzureCfgForStandardVMTypeWithoutDeploymentParameters), cloudprovider.NodeGroupDiscoveryOptions{})
	expectedErr := "open /var/lib/azure/azuredeploy.parameters.json: no such file or directory"
	assert.Nil(t, manager)
	assert.Equal(t, expectedErr, err.Error(), "return error does not match, expected: %v, actual: %v", expectedErr, err.Error())
}

func TestCreateAzureManagerWithNilConfig(t *testing.T) {
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
	os.Setenv("AZURE_MAX_DEPLOYMENT_COUNT", "8")
	os.Setenv("ENABLE_BACKOFF", "true")
	os.Setenv("BACKOFF_RETRIES", "1")
	os.Setenv("BACKOFF_EXPONENT", "1")
	os.Setenv("BACKOFF_DURATION", "1")
	os.Setenv("BACKOFF_JITTER", "1")
	os.Setenv("CLOUD_PROVIDER_RATE_LIMIT", "true")

	manager, err := CreateAzureManager(nil, cloudprovider.NodeGroupDiscoveryOptions{})
	assert.NoError(t, err)
	assert.Equal(t, true, reflect.DeepEqual(*expectedConfig, *manager.config), "unexpected azure manager configuration")

	// invalid bool for ARM_USE_MANAGED_IDENTITY_EXTENSION
	os.Setenv("ARM_USE_MANAGED_IDENTITY_EXTENSION", "invalidbool")
	manager, err = CreateAzureManager(nil, cloudprovider.NodeGroupDiscoveryOptions{})
	expectedErr0 := "strconv.ParseBool: parsing \"invalidbool\": invalid syntax"
	assert.Nil(t, manager)
	assert.Equal(t, expectedErr0, err.Error(), "Return err does not match, expected: %v, actual: %v", expectedErr0, err.Error())
	// revert back to good ARM_USE_MANAGED_IDENTITY_EXTENSION
	os.Setenv("ARM_USE_MANAGED_IDENTITY_EXTENSION", "true")

	// invalid int for AZURE_VMSS_CACHE_TTL
	os.Setenv("AZURE_VMSS_CACHE_TTL", "invalidint")
	manager, err = CreateAzureManager(nil, cloudprovider.NodeGroupDiscoveryOptions{})
	expectedErr := fmt.Errorf("failed to parse AZURE_VMSS_CACHE_TTL \"invalidint\": strconv.ParseInt: parsing \"invalidint\": invalid syntax")
	assert.Nil(t, manager)
	assert.Equal(t, expectedErr, err, "Return err does not match, expected: %v, actual: %v", expectedErr, err)
	// revert back to good AZURE_VMSS_CACHE_TTL
	os.Setenv("AZURE_VMSS_CACHE_TTL", "100")

	// invalid int for AZURE_MAX_DEPLOYMENT_COUNT
	os.Setenv("AZURE_MAX_DEPLOYMENT_COUNT", "invalidint")
	manager, err = CreateAzureManager(nil, cloudprovider.NodeGroupDiscoveryOptions{})
	expectedErr = fmt.Errorf("failed to parse AZURE_MAX_DEPLOYMENT_COUNT \"invalidint\": strconv.ParseInt: parsing \"invalidint\": invalid syntax")
	assert.Nil(t, manager)
	assert.Equal(t, expectedErr, err, "Return err does not match, expected: %v, actual: %v", expectedErr, err)
	// revert back to good AZURE_MAX_DEPLOYMENT_COUNT
	os.Setenv("AZURE_MAX_DEPLOYMENT_COUNT", "8")

	// zero AZURE_MAX_DEPLOYMENT_COUNT will use default value
	os.Setenv("AZURE_MAX_DEPLOYMENT_COUNT", "0")
	manager, err = CreateAzureManager(nil, cloudprovider.NodeGroupDiscoveryOptions{})
	assert.NoError(t, err)
	assert.Equal(t, int64(defaultMaxDeploymentsCount), (*manager.config).MaxDeploymentsCount, "MaxDeploymentsCount does not match.")
	// revert back to good AZURE_MAX_DEPLOYMENT_COUNT
	os.Setenv("AZURE_MAX_DEPLOYMENT_COUNT", "8")

	// invalid bool for ENABLE_BACKOFF
	os.Setenv("ENABLE_BACKOFF", "invalidbool")
	manager, err = CreateAzureManager(nil, cloudprovider.NodeGroupDiscoveryOptions{})
	expectedErr = fmt.Errorf("failed to parse ENABLE_BACKOFF \"invalidbool\": strconv.ParseBool: parsing \"invalidbool\": invalid syntax")
	assert.Nil(t, manager)
	assert.Equal(t, expectedErr, err, "Return err does not match, expected: %v, actual: %v", expectedErr, err)
	// revert back to good ENABLE_BACKOFF
	os.Setenv("ENABLE_BACKOFF", "true")

	// invalid int for BACKOFF_RETRIES
	os.Setenv("BACKOFF_RETRIES", "invalidint")
	manager, err = CreateAzureManager(nil, cloudprovider.NodeGroupDiscoveryOptions{})
	expectedErr = fmt.Errorf("failed to parse BACKOFF_RETRIES '\\x00': strconv.ParseInt: parsing \"invalidint\": invalid syntax")
	assert.Nil(t, manager)
	assert.Equal(t, expectedErr, err, "Return err does not match, expected: %v, actual: %v", expectedErr, err)
	// revert back to good BACKOFF_RETRIES
	os.Setenv("BACKOFF_RETRIES", "1")

	// empty BACKOFF_RETRIES will use default value
	os.Setenv("BACKOFF_RETRIES", "")
	manager, err = CreateAzureManager(nil, cloudprovider.NodeGroupDiscoveryOptions{})
	assert.NoError(t, err)
	assert.Equal(t, backoffRetriesDefault, (*manager.config).CloudProviderBackoffRetries, "CloudProviderBackoffRetries does not match.")
	// revert back to good BACKOFF_RETRIES
	os.Setenv("BACKOFF_RETRIES", "1")

	// invalid float for BACKOFF_EXPONENT
	os.Setenv("BACKOFF_EXPONENT", "invalidfloat")
	manager, err = CreateAzureManager(nil, cloudprovider.NodeGroupDiscoveryOptions{})
	expectedErr = fmt.Errorf("failed to parse BACKOFF_EXPONENT \"invalidfloat\": strconv.ParseFloat: parsing \"invalidfloat\": invalid syntax")
	assert.Nil(t, manager)
	assert.Equal(t, expectedErr, err, "Return err does not match, expected: %v, actual: %v", expectedErr, err)
	// revert back to good BACKOFF_EXPONENT
	os.Setenv("BACKOFF_EXPONENT", "1")

	// empty BACKOFF_EXPONENT will use default value
	os.Setenv("BACKOFF_EXPONENT", "")
	manager, err = CreateAzureManager(nil, cloudprovider.NodeGroupDiscoveryOptions{})
	assert.NoError(t, err)
	assert.Equal(t, backoffExponentDefault, (*manager.config).CloudProviderBackoffExponent, "CloudProviderBackoffExponent does not match.")
	// revert back to good BACKOFF_EXPONENT
	os.Setenv("BACKOFF_EXPONENT", "1")

	// invalid int for BACKOFF_DURATION
	os.Setenv("BACKOFF_DURATION", "invalidint")
	manager, err = CreateAzureManager(nil, cloudprovider.NodeGroupDiscoveryOptions{})
	expectedErr = fmt.Errorf("failed to parse BACKOFF_DURATION \"invalidint\": strconv.ParseInt: parsing \"invalidint\": invalid syntax")
	assert.Nil(t, manager)
	assert.Equal(t, expectedErr, err, "Return err does not match, expected: %v, actual: %v", expectedErr, err)
	// revert back to good BACKOFF_DURATION
	os.Setenv("BACKOFF_DURATION", "1")

	// empty BACKOFF_DURATION will use default value
	os.Setenv("BACKOFF_DURATION", "")
	manager, err = CreateAzureManager(nil, cloudprovider.NodeGroupDiscoveryOptions{})
	assert.NoError(t, err)
	assert.Equal(t, backoffDurationDefault, (*manager.config).CloudProviderBackoffDuration, "CloudProviderBackoffDuration does not match.")
	// revert back to good BACKOFF_DURATION
	os.Setenv("BACKOFF_DURATION", "1")

	// invalid float for BACKOFF_JITTER
	os.Setenv("BACKOFF_JITTER", "invalidfloat")
	manager, err = CreateAzureManager(nil, cloudprovider.NodeGroupDiscoveryOptions{})
	expectedErr = fmt.Errorf("failed to parse BACKOFF_JITTER \"invalidfloat\": strconv.ParseFloat: parsing \"invalidfloat\": invalid syntax")
	assert.Nil(t, manager)
	assert.Equal(t, expectedErr, err, "Return err does not match, expected: %v, actual: %v", expectedErr, err)
	// revert back to good BACKOFF_JITTER
	os.Setenv("BACKOFF_JITTER", "1")

	// empty BACKOFF_JITTER will use default value
	os.Setenv("BACKOFF_JITTER", "")
	manager, err = CreateAzureManager(nil, cloudprovider.NodeGroupDiscoveryOptions{})
	assert.NoError(t, err)
	assert.Equal(t, backoffJitterDefault, (*manager.config).CloudProviderBackoffJitter, "CloudProviderBackoffJitter does not match.")
	// revert back to good BACKOFF_JITTER
	os.Setenv("BACKOFF_JITTER", "1")

	// invalid bool for CLOUD_PROVIDER_RATE_LIMIT
	os.Setenv("CLOUD_PROVIDER_RATE_LIMIT", "invalidbool")
	manager, err = CreateAzureManager(nil, cloudprovider.NodeGroupDiscoveryOptions{})
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
	_, err := CreateAzureManager(strings.NewReader(invalidAzureCfg), cloudprovider.NodeGroupDiscoveryOptions{})
	assert.Error(t, err, "failed to unmarshal config body")
}

func TestFetchExplicitAsgs(t *testing.T) {
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
	manager.fetchExplicitAsgs(ngdo.NodeGroupSpecs)

	asgs := manager.asgCache.get()
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
	err := testAS.manager.fetchExplicitAsgs([]string{"1:5:testAS"})
	expectedErr := fmt.Errorf("failed to parse node group spec: deployment not found")
	assert.Equal(t, expectedErr, err, "testAS.manager.fetchExplicitAsgs return error does not match, expected: %v, actual: %v", expectedErr, err)
	err = testAS.manager.fetchExplicitAsgs(nil)
	assert.NoError(t, err)

	// test invalidVMType
	manager.config.VMType = "invalidVMType"
	err = manager.fetchExplicitAsgs(ngdo.NodeGroupSpecs)
	expectedErr = fmt.Errorf("failed to parse node group spec: vmtype invalidVMType not supported")
	assert.Equal(t, expectedErr, err, "manager.fetchExplicitAsgs return error does not match, expected: %v, actual: %v", expectedErr, err)
}

func TestParseLabelAutoDiscoverySpecs(t *testing.T) {
	testCases := []struct {
		name        string
		specs       []string
		expected    []labelAutoDiscoveryConfig
		expectedErr bool
	}{
		{
			name: "ValidSpec",
			specs: []string{
				"label:cluster-autoscaler-enabled=true,cluster-autoscaler-name=fake-cluster",
				"label:test-tag=test-value,another-test-tag=another-test-value",
			},
			expected: []labelAutoDiscoveryConfig{
				{Selector: map[string]string{"cluster-autoscaler-enabled": "true", "cluster-autoscaler-name": "fake-cluster"}},
				{Selector: map[string]string{"test-tag": "test-value", "another-test-tag": "another-test-value"}},
			},
		},
		{
			name:        "MissingAutoDiscoverLabel",
			specs:       []string{"test-tag=test-value,another-test-tag"},
			expectedErr: true,
		},
		{
			name:        "InvalidAutoDiscoerLabel",
			specs:       []string{"invalid:test-tag=test-value,another-test-tag"},
			expectedErr: true,
		},
		{
			name:        "MissingValue",
			specs:       []string{"label:test-tag="},
			expectedErr: true,
		},
		{
			name:        "MissingKey",
			specs:       []string{"label:=test-val"},
			expectedErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ngdo := cloudprovider.NodeGroupDiscoveryOptions{NodeGroupAutoDiscoverySpecs: tc.specs}
			actual, err := parseLabelAutoDiscoverySpecs(ngdo)
			if tc.expectedErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.True(t, assert.ObjectsAreEqualValues(tc.expected, actual), "expected %#v, but found: %#v", tc.expected, actual)
		})
	}
}

func TestListScalesets(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	manager := newTestAzureManager(t)
	vmssTag := "fake-tag"
	vmssTagValue := "fake-value"
	vmssName := "test-vmss"

	ngdo := cloudprovider.NodeGroupDiscoveryOptions{
		NodeGroupAutoDiscoverySpecs: []string{fmt.Sprintf("label:%s=%s", vmssTag, vmssTagValue)},
	}
	specs, err := parseLabelAutoDiscoverySpecs(ngdo)
	assert.NoError(t, err)

	testCases := []struct {
		name              string
		specs             map[string]string
		isListVMSSFail    bool
		expected          []cloudprovider.NodeGroup
		expectedErrString string
	}{
		{
			name:  "ValidMinMax",
			specs: map[string]string{"min": "5", "max": "50"},
			expected: []cloudprovider.NodeGroup{&ScaleSet{
				azureRef: azureRef{
					Name: vmssName,
				},
				minSize:           5,
				maxSize:           50,
				manager:           manager,
				curSize:           -1,
				sizeRefreshPeriod: defaultVmssSizeRefreshPeriod,
			}},
		},
		{
			name:              "InvalidMin",
			specs:             map[string]string{"min": "some-invalid-string"},
			expectedErrString: "invalid minimum size specified for vmss:",
		},
		{
			name:              "NoMin",
			specs:             map[string]string{"max": "50"},
			expectedErrString: fmt.Sprintf("no minimum size specified for vmss: %s", vmssName),
		},
		{
			name:              "InvalidMax",
			specs:             map[string]string{"min": "5", "max": "some-invalid-string"},
			expectedErrString: "invalid maximum size specified for vmss:",
		},
		{
			name:              "NoMax",
			specs:             map[string]string{"min": "5"},
			expectedErrString: fmt.Sprintf("no maximum size specified for vmss: %s", vmssName),
		},
		{
			name:              "MinLessThanZero",
			specs:             map[string]string{"min": "-4", "max": "20"},
			expectedErrString: fmt.Sprintf("minimum size must be a non-negative number of nodes"),
		},
		{
			name:              "MinGreaterThanMax",
			specs:             map[string]string{"min": "50", "max": "5"},
			expectedErrString: "maximum size must be greater than minimum size",
		},
		{
			name:              "ListVMSSFail",
			specs:             map[string]string{"min": "5", "max": "50"},
			isListVMSSFail:    true,
			expectedErrString: "List VMSS failed",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tags := make(map[string]*string)
			tags[vmssTag] = &vmssTagValue
			if val, ok := tc.specs["min"]; ok {
				tags["min"] = &val
			}
			if val, ok := tc.specs["max"]; ok {
				tags["max"] = &val
			}

			expectedScaleSets := []compute.VirtualMachineScaleSet{fakeVMSSWithTags(vmssName, tags)}
			mockVMSSClient := mockvmssclient.NewMockInterface(ctrl)
			if tc.isListVMSSFail {
				mockVMSSClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup).Return(nil, &retry.Error{RawError: fmt.Errorf("List VMSS failed")}).AnyTimes()
			} else {
				mockVMSSClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup).Return(expectedScaleSets, nil).AnyTimes()
			}
			manager.azClient.virtualMachineScaleSetsClient = mockVMSSClient

			asgs, err := manager.listScaleSets(specs)
			if tc.expectedErrString != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErrString)
				return
			}
			assert.NoError(t, err)
			assert.True(t, assert.ObjectsAreEqualValues(tc.expected, asgs), "expected %#v, but found: %#v", tc.expected, asgs)
		})
	}
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

	specs, err := parseLabelAutoDiscoverySpecs(ngdo)
	assert.NoError(t, err)

	asgs, err := manager.getFilteredAutoscalingGroups(specs)
	assert.NoError(t, err)
	expectedAsgs := []cloudprovider.NodeGroup{&ScaleSet{
		azureRef: azureRef{
			Name: vmssName,
		},
		minSize:           minVal,
		maxSize:           maxVal,
		manager:           manager,
		curSize:           -1,
		sizeRefreshPeriod: defaultVmssSizeRefreshPeriod,
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

	specs, err := parseLabelAutoDiscoverySpecs(ngdo)
	assert.NoError(t, err)

	asgs1, err1 := manager.getFilteredAutoscalingGroups(specs)
	assert.Nil(t, asgs1)
	assert.Nil(t, err1)

	manager.config.VMType = "invalidVMType"
	expectedErr := fmt.Errorf("vmType \"invalidVMType\" not supported")
	asgs, err2 := manager.getFilteredAutoscalingGroups(specs)
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

	specs, err := parseLabelAutoDiscoverySpecs(ngdo)
	assert.NoError(t, err)
	manager.asgAutoDiscoverySpecs = specs

	// assert cache is empty before fetching auto asgs
	asgs := manager.asgCache.get()
	assert.Equal(t, 0, len(asgs))

	manager.fetchAutoAsgs()
	asgs = manager.asgCache.get()
	assert.Equal(t, 1, len(asgs))
	assert.Equal(t, vmssName, asgs[0].Id())
	assert.Equal(t, minVal, asgs[0].MinSize())
	assert.Equal(t, maxVal, asgs[0].MaxSize())

	// test explicitlyConfigured
	manager.explicitlyConfigured[vmssName] = true
	manager.fetchAutoAsgs()
	asgs = manager.asgCache.get()
	assert.Equal(t, 1, len(asgs))
}

func TestInitializeCloudProviderRateLimitConfigWithNoConfigReturnsNoError(t *testing.T) {
	err := InitializeCloudProviderRateLimitConfig(nil)
	assert.Nil(t, err, "err should be nil")
}

func TestInitializeCloudProviderRateLimitConfigWithNoRateLimitSettingsReturnsDefaults(t *testing.T) {
	emptyConfig := &CloudProviderRateLimitConfig{}
	err := InitializeCloudProviderRateLimitConfig(emptyConfig)

	assert.NoError(t, err)
	assert.Equal(t, emptyConfig.CloudProviderRateLimitQPS, rateLimitQPSDefault)
	assert.Equal(t, emptyConfig.CloudProviderRateLimitBucket, rateLimitBucketDefault)
	assert.Equal(t, emptyConfig.CloudProviderRateLimitQPSWrite, rateLimitQPSDefault)
	assert.Equal(t, emptyConfig.CloudProviderRateLimitBucketWrite, rateLimitBucketDefault)
}

func TestInitializeCloudProviderRateLimitConfigWithReadRateLimitSettingsFromEnv(t *testing.T) {
	emptyConfig := &CloudProviderRateLimitConfig{}
	var rateLimitReadQPS float32 = 3.0
	rateLimitReadBuckets := 10
	os.Setenv(rateLimitReadQPSEnvVar, fmt.Sprintf("%.1f", rateLimitReadQPS))
	os.Setenv(rateLimitReadBucketsEnvVar, fmt.Sprintf("%d", rateLimitReadBuckets))

	err := InitializeCloudProviderRateLimitConfig(emptyConfig)
	assert.NoError(t, err)
	assert.Equal(t, emptyConfig.CloudProviderRateLimitQPS, rateLimitReadQPS)
	assert.Equal(t, emptyConfig.CloudProviderRateLimitBucket, rateLimitReadBuckets)
	assert.Equal(t, emptyConfig.CloudProviderRateLimitQPSWrite, rateLimitReadQPS)
	assert.Equal(t, emptyConfig.CloudProviderRateLimitBucketWrite, rateLimitReadBuckets)

	os.Unsetenv(rateLimitReadBucketsEnvVar)
	os.Unsetenv(rateLimitReadQPSEnvVar)
}

func TestInitializeCloudProviderRateLimitConfigWithReadAndWriteRateLimitSettingsFromEnv(t *testing.T) {
	emptyConfig := &CloudProviderRateLimitConfig{}
	var rateLimitReadQPS float32 = 3.0
	rateLimitReadBuckets := 10
	var rateLimitWriteQPS float32 = 6.0
	rateLimitWriteBuckets := 20

	os.Setenv(rateLimitReadQPSEnvVar, fmt.Sprintf("%.1f", rateLimitReadQPS))
	os.Setenv(rateLimitReadBucketsEnvVar, fmt.Sprintf("%d", rateLimitReadBuckets))
	os.Setenv(rateLimitWriteQPSEnvVar, fmt.Sprintf("%.1f", rateLimitWriteQPS))
	os.Setenv(rateLimitWriteBucketsEnvVar, fmt.Sprintf("%d", rateLimitWriteBuckets))

	err := InitializeCloudProviderRateLimitConfig(emptyConfig)

	assert.NoError(t, err)
	assert.Equal(t, emptyConfig.CloudProviderRateLimitQPS, rateLimitReadQPS)
	assert.Equal(t, emptyConfig.CloudProviderRateLimitBucket, rateLimitReadBuckets)
	assert.Equal(t, emptyConfig.CloudProviderRateLimitQPSWrite, rateLimitWriteQPS)
	assert.Equal(t, emptyConfig.CloudProviderRateLimitBucketWrite, rateLimitWriteBuckets)

	os.Unsetenv(rateLimitReadQPSEnvVar)
	os.Unsetenv(rateLimitReadBucketsEnvVar)
	os.Unsetenv(rateLimitWriteQPSEnvVar)
	os.Unsetenv(rateLimitWriteBucketsEnvVar)
}

func TestInitializeCloudProviderRateLimitConfigWithReadAndWriteRateLimitAlreadySetInConfig(t *testing.T) {
	var rateLimitReadQPS float32 = 3.0
	rateLimitReadBuckets := 10
	var rateLimitWriteQPS float32 = 6.0
	rateLimitWriteBuckets := 20

	configWithRateLimits := &CloudProviderRateLimitConfig{
		RateLimitConfig: azclients.RateLimitConfig{
			CloudProviderRateLimitBucket:      rateLimitReadBuckets,
			CloudProviderRateLimitBucketWrite: rateLimitWriteBuckets,
			CloudProviderRateLimitQPS:         rateLimitReadQPS,
			CloudProviderRateLimitQPSWrite:    rateLimitWriteQPS,
		},
	}

	os.Setenv(rateLimitReadQPSEnvVar, "99")
	os.Setenv(rateLimitReadBucketsEnvVar, "99")
	os.Setenv(rateLimitWriteQPSEnvVar, "99")
	os.Setenv(rateLimitWriteBucketsEnvVar, "99")

	err := InitializeCloudProviderRateLimitConfig(configWithRateLimits)

	assert.NoError(t, err)
	assert.Equal(t, configWithRateLimits.CloudProviderRateLimitQPS, rateLimitReadQPS)
	assert.Equal(t, configWithRateLimits.CloudProviderRateLimitBucket, rateLimitReadBuckets)
	assert.Equal(t, configWithRateLimits.CloudProviderRateLimitQPSWrite, rateLimitWriteQPS)
	assert.Equal(t, configWithRateLimits.CloudProviderRateLimitBucketWrite, rateLimitWriteBuckets)

	os.Unsetenv(rateLimitReadQPSEnvVar)
	os.Unsetenv(rateLimitReadBucketsEnvVar)
	os.Unsetenv(rateLimitWriteQPSEnvVar)
	os.Unsetenv(rateLimitWriteBucketsEnvVar)
}

func TestInitializeCloudProviderRateLimitConfigWithInvalidReadAndWriteRateLimitSettingsFromEnv(t *testing.T) {
	emptyConfig := &CloudProviderRateLimitConfig{}
	var rateLimitReadQPS float32 = 3.0
	rateLimitReadBuckets := 10
	var rateLimitWriteQPS float32 = 6.0
	rateLimitWriteBuckets := 20

	invalidSetting := "invalid"
	testCases := []struct {
		desc                                 string
		isInvalidRateLimitReadQPSEnvVar      bool
		isInvalidRateLimitReadBucketsEnvVar  bool
		isInvalidRateLimitWriteQPSEnvVar     bool
		isInvalidRateLimitWriteBucketsEnvVar bool
		expectedErr                          bool
		expectedErrMsg                       error
	}{
		{
			desc:                            "an error shall be returned if invalid rateLimitReadQPSEnvVar",
			isInvalidRateLimitReadQPSEnvVar: true,
			expectedErr:                     true,
			expectedErrMsg:                  fmt.Errorf("failed to parse %s: %q, strconv.ParseFloat: parsing \"invalid\": invalid syntax", rateLimitReadQPSEnvVar, invalidSetting),
		},
		{
			desc:                                "an error shall be returned if invalid rateLimitReadBucketsEnvVar",
			isInvalidRateLimitReadBucketsEnvVar: true,
			expectedErr:                         true,
			expectedErrMsg:                      fmt.Errorf("failed to parse %s: %q, strconv.ParseInt: parsing \"invalid\": invalid syntax", rateLimitReadBucketsEnvVar, invalidSetting),
		},
		{
			desc:                             "an error shall be returned if invalid rateLimitWriteQPSEnvVar",
			isInvalidRateLimitWriteQPSEnvVar: true,
			expectedErr:                      true,
			expectedErrMsg:                   fmt.Errorf("failed to parse %s: %q, strconv.ParseFloat: parsing \"invalid\": invalid syntax", rateLimitWriteQPSEnvVar, invalidSetting),
		},
		{
			desc:                                 "an error shall be returned if invalid rateLimitWriteBucketsEnvVar",
			isInvalidRateLimitWriteBucketsEnvVar: true,
			expectedErr:                          true,
			expectedErrMsg:                       fmt.Errorf("failed to parse %s: %q, strconv.ParseInt: parsing \"invalid\": invalid syntax", rateLimitWriteBucketsEnvVar, invalidSetting),
		},
	}

	for i, test := range testCases {
		if test.isInvalidRateLimitReadQPSEnvVar {
			os.Setenv(rateLimitReadQPSEnvVar, invalidSetting)
		} else {
			os.Setenv(rateLimitReadQPSEnvVar, fmt.Sprintf("%.1f", rateLimitReadQPS))
		}
		if test.isInvalidRateLimitReadBucketsEnvVar {
			os.Setenv(rateLimitReadBucketsEnvVar, invalidSetting)
		} else {
			os.Setenv(rateLimitReadBucketsEnvVar, fmt.Sprintf("%d", rateLimitReadBuckets))
		}
		if test.isInvalidRateLimitWriteQPSEnvVar {
			os.Setenv(rateLimitWriteQPSEnvVar, invalidSetting)
		} else {
			os.Setenv(rateLimitWriteQPSEnvVar, fmt.Sprintf("%.1f", rateLimitWriteQPS))
		}
		if test.isInvalidRateLimitWriteBucketsEnvVar {
			os.Setenv(rateLimitWriteBucketsEnvVar, invalidSetting)
		} else {
			os.Setenv(rateLimitWriteBucketsEnvVar, fmt.Sprintf("%d", rateLimitWriteBuckets))
		}

		err := InitializeCloudProviderRateLimitConfig(emptyConfig)

		assert.Equal(t, test.expectedErr, err != nil, "TestCase[%d]: %s, return error: %v", i, test.desc, err)
		assert.Equal(t, test.expectedErrMsg, err, "TestCase[%d]: %s, expected: %v, return: %v", i, test.desc, test.expectedErrMsg, err)

		os.Unsetenv(rateLimitReadQPSEnvVar)
		os.Unsetenv(rateLimitReadBucketsEnvVar)
		os.Unsetenv(rateLimitWriteQPSEnvVar)
		os.Unsetenv(rateLimitWriteBucketsEnvVar)
	}
}

func TestOverrideDefaultRateLimitConfig(t *testing.T) {
	var rateLimitReadQPS float32 = 3.0
	rateLimitReadBuckets := 10
	var rateLimitWriteQPS float32 = 6.0
	rateLimitWriteBuckets := 20

	defaultConfigWithRateLimits := &CloudProviderRateLimitConfig{
		RateLimitConfig: azclients.RateLimitConfig{
			CloudProviderRateLimitBucket:      rateLimitReadBuckets,
			CloudProviderRateLimitBucketWrite: rateLimitWriteBuckets,
			CloudProviderRateLimitQPS:         rateLimitReadQPS,
			CloudProviderRateLimitQPSWrite:    rateLimitWriteQPS,
		},
	}

	configWithRateLimits := &CloudProviderRateLimitConfig{
		RateLimitConfig: azclients.RateLimitConfig{
			CloudProviderRateLimit:            true,
			CloudProviderRateLimitBucket:      0,
			CloudProviderRateLimitBucketWrite: 0,
			CloudProviderRateLimitQPS:         0,
			CloudProviderRateLimitQPSWrite:    0,
		},
	}

	newconfig := overrideDefaultRateLimitConfig(&defaultConfigWithRateLimits.RateLimitConfig, &configWithRateLimits.RateLimitConfig)

	assert.Equal(t, defaultConfigWithRateLimits.CloudProviderRateLimitQPS, newconfig.CloudProviderRateLimitQPS)
	assert.Equal(t, defaultConfigWithRateLimits.CloudProviderRateLimitBucket, newconfig.CloudProviderRateLimitBucket)
	assert.Equal(t, defaultConfigWithRateLimits.CloudProviderRateLimitQPSWrite, newconfig.CloudProviderRateLimitQPSWrite)
	assert.Equal(t, defaultConfigWithRateLimits.CloudProviderRateLimitBucketWrite, newconfig.CloudProviderRateLimitBucketWrite)

	falseCloudProviderRateLimit := &CloudProviderRateLimitConfig{
		RateLimitConfig: azclients.RateLimitConfig{
			CloudProviderRateLimit: false,
		},
	}
	newconfig = overrideDefaultRateLimitConfig(&defaultConfigWithRateLimits.RateLimitConfig, &falseCloudProviderRateLimit.RateLimitConfig)
	assert.Equal(t, &falseCloudProviderRateLimit.RateLimitConfig, newconfig)
}

func TestGetSubscriptionIdFromInstanceMetadata(t *testing.T) {
	// metadataURL in azure_manager.go is not available for our tests, expect fail.
	result, err := getSubscriptionIdFromInstanceMetadata()
	expected := ""
	assert.NotNil(t, err.Error())
	assert.Equal(t, expected, result, "Verify return result failed, expected: %v, actual: %v", expected, result)
}

func TestManagerRefreshAndCleanup(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	manager := newTestAzureManager(t)
	err := manager.Refresh()
	assert.NoError(t, err)
	manager.Cleanup()
}
