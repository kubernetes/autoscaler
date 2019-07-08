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
	"net/http"
	"testing"

	"github.com/Azure/go-autorest/autorest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
)

func newTestScaleSet(manager *AzureManager, name string) *ScaleSet {
	return &ScaleSet{
		azureRef: azureRef{
			Name: name,
		},
		manager: manager,
		minSize: 1,
		maxSize: 5,
	}
}

func TestMaxSize(t *testing.T) {
	provider := newTestProvider(t)
	registered := provider.azureManager.RegisterAsg(
		newTestScaleSet(provider.azureManager, "test-asg"))
	assert.True(t, registered)
	assert.Equal(t, len(provider.NodeGroups()), 1)
	assert.Equal(t, provider.NodeGroups()[0].MaxSize(), 5)
}

func TestMinSize(t *testing.T) {
	provider := newTestProvider(t)
	registered := provider.azureManager.RegisterAsg(
		newTestScaleSet(provider.azureManager, "test-asg"))
	assert.True(t, registered)
	assert.Equal(t, len(provider.NodeGroups()), 1)
	assert.Equal(t, provider.NodeGroups()[0].MinSize(), 1)
}

func TestTargetSize(t *testing.T) {
	provider := newTestProvider(t)
	registered := provider.azureManager.RegisterAsg(
		newTestScaleSet(provider.azureManager, "test-asg"))
	assert.True(t, registered)
	assert.Equal(t, len(provider.NodeGroups()), 1)

	targetSize, err := provider.NodeGroups()[0].TargetSize()
	assert.NoError(t, err)
	assert.Equal(t, targetSize, 2)
}

func TestIncreaseSize(t *testing.T) {
	provider := newTestProvider(t)
	registered := provider.azureManager.RegisterAsg(
		newTestScaleSet(provider.azureManager, "test-asg"))
	assert.True(t, registered)
	assert.Equal(t, len(provider.NodeGroups()), 1)

	// current target size is 2.
	targetSize, err := provider.NodeGroups()[0].TargetSize()
	assert.NoError(t, err)
	assert.Equal(t, targetSize, 2)

	// increase 3 nodes.
	err = provider.NodeGroups()[0].IncreaseSize(3)
	assert.NoError(t, err)

	// new target size should be 5.
	targetSize, err = provider.NodeGroups()[0].TargetSize()
	assert.NoError(t, err)
	assert.Equal(t, 5, targetSize)
}

func TestBelongs(t *testing.T) {
	provider := newTestProvider(t)
	registered := provider.azureManager.RegisterAsg(
		newTestScaleSet(provider.azureManager, "test-asg"))
	assert.True(t, registered)

	scaleSet, ok := provider.NodeGroups()[0].(*ScaleSet)
	assert.True(t, ok)

	invalidNode := &apiv1.Node{
		Spec: apiv1.NodeSpec{
			ProviderID: "azure:///subscriptions/test-subscrition-id/resourcegroups/invalid-asg/providers/microsoft.compute/virtualmachinescalesets/agents/virtualmachines/0",
		},
	}
	_, err := scaleSet.Belongs(invalidNode)
	assert.Error(t, err)

	validNode := &apiv1.Node{
		Spec: apiv1.NodeSpec{
			ProviderID: "azure://" + fakeVirtualMachineScaleSetVMID,
		},
	}
	belongs, err := scaleSet.Belongs(validNode)
	assert.Equal(t, true, belongs)
	assert.NoError(t, err)
}

func TestDeleteNodes(t *testing.T) {
	manager := newTestAzureManager(t)
	scaleSetClient := &VirtualMachineScaleSetsClientMock{}
	response := autorest.Response{
		Response: &http.Response{
			Status: "OK",
		},
	}
	scaleSetClient.On("DeleteInstances", mock.Anything, "test-asg", mock.Anything, mock.Anything).Return(response, nil)
	manager.azClient.virtualMachineScaleSetsClient = scaleSetClient

	resourceLimiter := cloudprovider.NewResourceLimiter(
		map[string]int64{cloudprovider.ResourceNameCores: 1, cloudprovider.ResourceNameMemory: 10000000},
		map[string]int64{cloudprovider.ResourceNameCores: 10, cloudprovider.ResourceNameMemory: 100000000})
	provider, err := BuildAzureCloudProvider(manager, resourceLimiter)
	assert.NoError(t, err)

	registered := manager.RegisterAsg(
		newTestScaleSet(manager, "test-asg"))
	assert.True(t, registered)

	node := &apiv1.Node{
		Spec: apiv1.NodeSpec{
			ProviderID: "azure://" + fakeVirtualMachineScaleSetVMID,
		},
	}
	scaleSet, ok := provider.NodeGroups()[0].(*ScaleSet)
	assert.True(t, ok)
	err = scaleSet.DeleteNodes([]*apiv1.Node{node})
	assert.NoError(t, err)
	scaleSetClient.AssertNumberOfCalls(t, "DeleteInstances", 1)
}

func TestId(t *testing.T) {
	provider := newTestProvider(t)
	registered := provider.azureManager.RegisterAsg(
		newTestScaleSet(provider.azureManager, "test-asg"))
	assert.True(t, registered)
	assert.Equal(t, len(provider.NodeGroups()), 1)
	assert.Equal(t, provider.NodeGroups()[0].Id(), "test-asg")
}

func TestDebug(t *testing.T) {
	asg := ScaleSet{
		manager: newTestAzureManager(t),
		minSize: 5,
		maxSize: 55,
	}
	asg.Name = "test-scale-set"
	assert.Equal(t, asg.Debug(), "test-scale-set (5:55)")
}

func TestScaleSetNodes(t *testing.T) {
	provider := newTestProvider(t)
	registered := provider.azureManager.RegisterAsg(
		newTestScaleSet(provider.azureManager, "test-asg"))
	assert.True(t, registered)
	assert.Equal(t, len(provider.NodeGroups()), 1)

	fakeProviderID := "azure://" + fakeVirtualMachineScaleSetVMID
	node := &apiv1.Node{
		Spec: apiv1.NodeSpec{
			ProviderID: fakeProviderID,
		},
	}
	group, err := provider.NodeGroupForNode(node)
	assert.NoError(t, err)
	assert.NotNil(t, group, "Group should not be nil")
	assert.Equal(t, group.Id(), "test-asg")
	assert.Equal(t, group.MinSize(), 1)
	assert.Equal(t, group.MaxSize(), 5)

	ss, ok := group.(*ScaleSet)
	assert.True(t, ok)
	assert.NotNil(t, ss)
	instances, err := group.Nodes()
	assert.NoError(t, err)
	assert.Equal(t, len(instances), 1)
	assert.Equal(t, instances[0], cloudprovider.Instance{Id: fakeProviderID})
}

func TestTemplateNodeInfo(t *testing.T) {
	provider := newTestProvider(t)
	registered := provider.azureManager.RegisterAsg(
		newTestScaleSet(provider.azureManager, "test-asg"))
	assert.True(t, registered)
	assert.Equal(t, len(provider.NodeGroups()), 1)

	asg := ScaleSet{
		manager: newTestAzureManager(t),
		minSize: 1,
		maxSize: 5,
	}
	asg.Name = "test-scale-set"

	nodeInfo, err := asg.TemplateNodeInfo()
	assert.NoError(t, err)
	assert.NotNil(t, nodeInfo)
	assert.NotEmpty(t, nodeInfo.Pods())
}
