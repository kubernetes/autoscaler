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

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-12-01/compute"
	"github.com/stretchr/testify/assert"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	azclients "k8s.io/legacy-cloud-providers/azure/clients"
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

func TestCreateAzureManagerInvalidConfig(t *testing.T) {
	_, err := CreateAzureManager(strings.NewReader(invalidAzureCfg), cloudprovider.NodeGroupDiscoveryOptions{})
	assert.Error(t, err, "failed to unmarshal config body")
}

func TestFetchExplicitAsgs(t *testing.T) {
	min, max, name := 1, 15, "test-asg"
	ngdo := cloudprovider.NodeGroupDiscoveryOptions{
		NodeGroupSpecs: []string{
			fmt.Sprintf("%d:%d:%s", min, max, name),
		},
	}

	manager := newTestAzureManager(t)
	manager.fetchExplicitAsgs(ngdo.NodeGroupSpecs)

	asgs := manager.asgCache.get()
	assert.Equal(t, 1, len(asgs))
	assert.Equal(t, name, asgs[0].Id())
	assert.Equal(t, min, asgs[0].MinSize())
	assert.Equal(t, max, asgs[0].MaxSize())
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

			scaleSetClient := &VirtualMachineScaleSetsClientMock{
				FakeStore: map[string]map[string]compute.VirtualMachineScaleSet{
					"test": {
						vmssName: fakeVMSSWithTags(vmssName, tags),
					},
				},
			}

			manager.azClient.virtualMachineScaleSetsClient = scaleSetClient
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
	scaleSetClient := &VirtualMachineScaleSetsClientMock{
		FakeStore: map[string]map[string]compute.VirtualMachineScaleSet{
			"test": {
				vmssName: fakeVMSSWithTags(vmssName, map[string]*string{vmssTag: &vmssTagValue, "min": &min, "max": &max}),
			},
		},
	}

	manager := newTestAzureManager(t)
	manager.azClient.virtualMachineScaleSetsClient = scaleSetClient
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

func TestFetchAutoAsgsVmss(t *testing.T) {
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
	scaleSetClient := &VirtualMachineScaleSetsClientMock{
		FakeStore: map[string]map[string]compute.VirtualMachineScaleSet{
			"test": {
				vmssName: fakeVMSSWithTags(vmssName, map[string]*string{vmssTag: &vmssTagValue, "min": &minString, "max": &maxString}),
			},
		},
	}

	manager := newTestAzureManager(t)
	manager.azClient.virtualMachineScaleSetsClient = scaleSetClient
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
}
