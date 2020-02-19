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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
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
	"asgCacheTTL": 900,
	"vmssCacheTTL": 60}`

const invalidAzureCfg = `{{}"cloud": "AzurePublicCloud",}`

func TestCreateAzureManagerValidConfig(t *testing.T) {
	manager, err := CreateAzureManager(strings.NewReader(validAzureCfg), cloudprovider.NodeGroupDiscoveryOptions{})

	expectdConfig := &Config{
		Cloud:           "AzurePublicCloud",
		TenantID:        "fakeId",
		SubscriptionID:  "fakeId",
		ResourceGroup:   "fakeId",
		VMType:          "vmss",
		AADClientID:     "fakeId",
		AADClientSecret: "fakeId",
		AsgCacheTTL:     900,
		VmssCacheTTL:    60,
	}

	assert.NoError(t, err)
	assert.Equal(t, expectdConfig, manager.config, "unexpected azure manager configuration")
}

func TestCreateAzureManagerInvalidConfig(t *testing.T) {
	_, err := CreateAzureManager(strings.NewReader(invalidAzureCfg), cloudprovider.NodeGroupDiscoveryOptions{})
	assert.Error(t, err, "failed to unmarshal config body")
}
