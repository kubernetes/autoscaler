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

func TestMaxSize(t *testing.T) {
	provider, err := newTestProvider()
	assert.NoError(t, err)

	err = provider.addNodeGroup("1:5:test-asg")
	assert.NoError(t, err)
	assert.Equal(t, len(provider.nodeGroups), 1)
	assert.Equal(t, provider.nodeGroups[0].MaxSize(), 5)
}

func TestMinSize(t *testing.T) {
	provider, err := newTestProvider()
	assert.NoError(t, err)

	err = provider.addNodeGroup("1:5:test-asg")
	assert.NoError(t, err)
	assert.Equal(t, len(provider.nodeGroups), 1)
	assert.Equal(t, provider.nodeGroups[0].MinSize(), 1)
}

func TestTargetSize(t *testing.T) {
	provider, err := newTestProvider()
	assert.NoError(t, err)

	err = provider.addNodeGroup("1:5:test-asg")
	assert.NoError(t, err)
	targetSize, err := provider.nodeGroups[0].TargetSize()
	assert.Equal(t, targetSize, 2)
	assert.NoError(t, err)
}

func TestIncreaseSize(t *testing.T) {
	provider, err := newTestProvider()
	assert.NoError(t, err)

	err = provider.addNodeGroup("1:5:test-asg")
	assert.NoError(t, err)
	assert.Equal(t, len(provider.nodeGroups), 1)

	err = provider.nodeGroups[0].IncreaseSize(1)
	assert.NoError(t, err)
}

func TestBelongs(t *testing.T) {
	provider, err := newTestProvider()
	assert.NoError(t, err)

	err = provider.addNodeGroup("1:5:test-asg")
	assert.NoError(t, err)

	scaleSet, ok := provider.nodeGroups[0].(*ScaleSet)
	assert.True(t, ok)

	invalidNode := &apiv1.Node{
		Spec: apiv1.NodeSpec{
			ProviderID: "azure:///subscriptions/subscriptionId/resourceGroups/kubernetes/providers/Microsoft.Compute/virtualMachines/invalid-instance-id",
		},
	}
	_, err = scaleSet.Belongs(invalidNode)
	assert.Error(t, err)

	validNode := &apiv1.Node{
		Spec: apiv1.NodeSpec{
			ProviderID: "azure://123E4567-E89B-12D3-A456-426655440000",
		},
	}

	belongs, err := scaleSet.Belongs(validNode)
	assert.Equal(t, true, belongs)
	assert.NoError(t, err)
}

func TestDeleteNodes(t *testing.T) {
	manager := newTestAzureManager()
	scaleSetClient := &VirtualMachineScaleSetsClientMock{}
	instanceIds := make([]string, 1)
	instanceIds[0] = "test-instance-id"
	response := autorest.Response{
		Response: &http.Response{
			Status: "OK",
		},
	}
	scaleSetClient.On("DeleteInstances", mock.Anything, "test-asg", mock.Anything, mock.Anything).Return(response, nil)
	manager.virtualMachineScaleSetsClient = scaleSetClient

	resourceLimiter := cloudprovider.NewResourceLimiter(
		map[string]int64{cloudprovider.ResourceNameCores: 1, cloudprovider.ResourceNameMemory: 10000000},
		map[string]int64{cloudprovider.ResourceNameCores: 10, cloudprovider.ResourceNameMemory: 100000000})
	provider, err := BuildAzureCloudProvider(manager, nil, resourceLimiter)
	assert.NoError(t, err)
	err = provider.addNodeGroup("1:5:test-asg")
	assert.NoError(t, err)

	node := &apiv1.Node{
		Spec: apiv1.NodeSpec{
			ProviderID: "azure://123E4567-E89B-12D3-A456-426655440000",
		},
	}
	scaleSet, ok := provider.nodeGroups[0].(*ScaleSet)
	assert.True(t, ok)
	err = scaleSet.DeleteNodes([]*apiv1.Node{node})
	assert.NoError(t, err)
	scaleSetClient.AssertNumberOfCalls(t, "DeleteInstances", 1)
}

func TestId(t *testing.T) {
	provider, err := newTestProvider()
	assert.NoError(t, err)
	err = provider.addNodeGroup("1:5:test-asg")
	assert.NoError(t, err)
	assert.Equal(t, len(provider.nodeGroups), 1)
	assert.Equal(t, provider.nodeGroups[0].Id(), "test-asg")
}

func TestDebug(t *testing.T) {
	asg := ScaleSet{
		AzureManager: newTestAzureManager(),
		minSize:      5,
		maxSize:      55,
	}
	asg.Name = "test-scale-set"
	assert.Equal(t, asg.Debug(), "test-scale-set (5:55)")
}
