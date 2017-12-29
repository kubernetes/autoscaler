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

	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/stretchr/testify/assert"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
)

func newTestAzureManager() *AzureManager {
	return &AzureManager{
		config:           &Config{VMType: vmTypeVMSS},
		env:              azure.PublicCloud,
		interrupt:        make(chan struct{}),
		instanceIDsCache: make(map[string]string),
		nodeGroups:       make([]cloudprovider.NodeGroup, 0),
		nodeGroupsCache:  make(map[AzureRef]cloudprovider.NodeGroup),

		disksClient:                     &DisksClientMock{},
		interfacesClient:                &InterfacesClientMock{},
		storageAccountsClient:           &AccountsClientMock{},
		deploymentsClient:               &DeploymentsClientMock{},
		virtualMachinesClient:           &VirtualMachinesClientMock{},
		virtualMachineScaleSetsClient:   &VirtualMachineScaleSetsClientMock{},
		virtualMachineScaleSetVMsClient: &VirtualMachineScaleSetVMsClientMock{},
	}
}

func newTestProvider() (*AzureCloudProvider, error) {
	manager := newTestAzureManager()
	resourceLimiter := cloudprovider.NewResourceLimiter(
		map[string]int64{cloudprovider.ResourceNameCores: 1, cloudprovider.ResourceNameMemory: 10000000},
		map[string]int64{cloudprovider.ResourceNameCores: 10, cloudprovider.ResourceNameMemory: 100000000})
	return BuildAzureCloudProvider(manager, nil, resourceLimiter)
}

func TestBuildAzureCloudProvider(t *testing.T) {
	resourceLimiter := cloudprovider.NewResourceLimiter(
		map[string]int64{cloudprovider.ResourceNameCores: 1, cloudprovider.ResourceNameMemory: 10000000},
		map[string]int64{cloudprovider.ResourceNameCores: 10, cloudprovider.ResourceNameMemory: 100000000})
	m := newTestAzureManager()
	_, err := BuildAzureCloudProvider(m, []string{"bad spec"}, resourceLimiter)
	assert.Error(t, err)

	_, err = BuildAzureCloudProvider(m, nil, resourceLimiter)
	assert.NoError(t, err)
}

func TestAddNodeGroup(t *testing.T) {
	provider, err := newTestProvider()
	assert.NoError(t, err)

	err = provider.addNodeGroup("bad spec")
	assert.Error(t, err)
	assert.Equal(t, len(provider.nodeGroups), 0)

	err = provider.addNodeGroup("1:5:test-asg")
	assert.NoError(t, err)
	assert.Equal(t, len(provider.nodeGroups), 1)
}

func TestName(t *testing.T) {
	provider, err := newTestProvider()
	assert.NoError(t, err)
	assert.Equal(t, provider.Name(), "azure")
}

func TestNodeGroups(t *testing.T) {
	provider, err := newTestProvider()
	assert.NoError(t, err)
	assert.Equal(t, len(provider.NodeGroups()), 0)

	err = provider.addNodeGroup("1:5:test-asg")
	assert.NoError(t, err)
	assert.Equal(t, len(provider.NodeGroups()), 1)
}

func TestNodeGroupForNode(t *testing.T) {
	provider, err := newTestProvider()
	assert.NoError(t, err)

	err = provider.addNodeGroup("1:5:test-asg")
	assert.NoError(t, err)
	assert.Equal(t, len(provider.nodeGroups), 1)

	node := &apiv1.Node{
		Spec: apiv1.NodeSpec{
			ProviderID: "azure://123E4567-E89B-12D3-A456-426655440000",
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

func TestBuildNodeGroup(t *testing.T) {
	provider, err := newTestProvider()
	assert.NoError(t, err)

	_, err = provider.buildNodeGroup("a")
	assert.Error(t, err)
	_, err = provider.buildNodeGroup("a:b:c")
	assert.Error(t, err)
	_, err = provider.buildNodeGroup("1:")
	assert.Error(t, err)
	_, err = provider.buildNodeGroup("1:2:")
	assert.Error(t, err)

	_, err = provider.buildNodeGroup("-1:2:")
	assert.Error(t, err)

	_, err = provider.buildNodeGroup("5:3:")
	assert.Error(t, err)

	_, err = provider.buildNodeGroup("5:ddd:test-name")
	assert.Error(t, err)

	asg, err := provider.buildNodeGroup("111:222:test-name")
	assert.NoError(t, err)
	assert.Equal(t, 111, asg.MinSize())
	assert.Equal(t, 222, asg.MaxSize())
	assert.Equal(t, "test-name", asg.Id())
}
