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
	"net/http"
	"testing"

	"github.com/Azure/azure-sdk-for-go/arm/compute"
	"github.com/Azure/go-autorest/autorest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
)

// Mock for VirtualMachineScaleSetsClient
type VirtualMachineScaleSetsClientMock struct {
	mock.Mock
}

func (client *VirtualMachineScaleSetsClientMock) Get(resourceGroupName string,
	vmScaleSetName string) (result compute.VirtualMachineScaleSet, err error) {
	fmt.Printf("Called VirtualMachineScaleSetsClientMock.Get(%s,%s)\n", resourceGroupName, vmScaleSetName)
	capacity := int64(2)
	properties := compute.VirtualMachineScaleSetProperties{}
	return compute.VirtualMachineScaleSet{

		Name: &vmScaleSetName,
		Sku: &compute.Sku{
			Capacity: &capacity,
		},
		VirtualMachineScaleSetProperties: &properties,
	}, nil
}

func (client *VirtualMachineScaleSetsClientMock) CreateOrUpdate(
	resourceGroupName string, name string, parameters compute.VirtualMachineScaleSet, cancel <-chan struct{}) (<-chan compute.VirtualMachineScaleSet, <-chan error) {
	fmt.Printf("Called VirtualMachineScaleSetsClientMock.CreateOrUpdate(%s,%s)\n", resourceGroupName, name)

	errChan := make(chan error)
	go func() {
		errChan <- nil
	}()
	return nil, errChan
}

func (client *VirtualMachineScaleSetsClientMock) DeleteInstances(resourceGroupName string, vmScaleSetName string,
	vmInstanceIDs compute.VirtualMachineScaleSetVMInstanceRequiredIDs, cancel <-chan struct{}) (<-chan compute.OperationStatusResponse, <-chan error) {

	args := client.Called(resourceGroupName, vmScaleSetName, vmInstanceIDs, cancel)
	errChan := make(chan error)
	go func() {
		errChan <- args.Error(1)
	}()
	return nil, errChan
}

// Mock for VirtualMachineScaleSetVMsClient
type VirtualMachineScaleSetVMsClientMock struct {
	mock.Mock
}

func (m *VirtualMachineScaleSetVMsClientMock) List(resourceGroupName string, virtualMachineScaleSetName string, filter string, selectParameter string, expand string) (result compute.VirtualMachineScaleSetVMListResult, err error) {

	value := make([]compute.VirtualMachineScaleSetVM, 1)
	vmInstanceId := "test-instance-id"
	properties := compute.VirtualMachineScaleSetVMProperties{}
	vmId := "67453E12-9BE8-D312-A456-426655440000"
	properties.VMID = &vmId
	value[0] = compute.VirtualMachineScaleSetVM{
		InstanceID:                         &vmInstanceId,
		VirtualMachineScaleSetVMProperties: &properties,
	}

	return compute.VirtualMachineScaleSetVMListResult{
		Value: &value,
	}, nil

}

var testAzureManager = &AzureManager{
	scaleSets:        make([]*scaleSetInformation, 0),
	scaleSetClient:   &VirtualMachineScaleSetsClientMock{},
	scaleSetVmClient: &VirtualMachineScaleSetVMsClientMock{},
	scaleSetCache:    make(map[AzureRef]*ScaleSet),
	interrupt:        make(chan struct{}),
}

func testProvider(t *testing.T, m *AzureManager) *AzureCloudProvider {
	resourceLimiter := cloudprovider.NewResourceLimiter(
		map[string]int64{cloudprovider.ResourceNameCores: 1, cloudprovider.ResourceNameMemory: 10000000},
		map[string]int64{cloudprovider.ResourceNameCores: 10, cloudprovider.ResourceNameMemory: 100000000})
	provider, err := BuildAzureCloudProvider(m, nil, resourceLimiter)
	assert.NoError(t, err)
	return provider
}

func TestBuildAwsCloudProvider(t *testing.T) {
	resourceLimiter := cloudprovider.NewResourceLimiter(
		map[string]int64{cloudprovider.ResourceNameCores: 1, cloudprovider.ResourceNameMemory: 10000000},
		map[string]int64{cloudprovider.ResourceNameCores: 10, cloudprovider.ResourceNameMemory: 100000000})
	m := testAzureManager
	_, err := BuildAzureCloudProvider(m, []string{"bad spec"}, resourceLimiter)
	assert.Error(t, err)

	_, err = BuildAzureCloudProvider(m, nil, resourceLimiter)
	assert.NoError(t, err)
}

func TestAddNodeGroup(t *testing.T) {
	provider := testProvider(t, testAzureManager)
	err := provider.addNodeGroup("bad spec")
	assert.Error(t, err)
	assert.Equal(t, len(provider.scaleSets), 0)

	err = provider.addNodeGroup("1:5:test-asg")
	assert.NoError(t, err)
	assert.Equal(t, len(provider.scaleSets), 1)
}

func TestName(t *testing.T) {
	provider := testProvider(t, testAzureManager)
	assert.Equal(t, provider.Name(), "azure")
}

func TestNodeGroups(t *testing.T) {
	provider := testProvider(t, testAzureManager)
	assert.Equal(t, len(provider.NodeGroups()), 0)
	err := provider.addNodeGroup("1:5:test-asg")
	assert.NoError(t, err)
	assert.Equal(t, len(provider.NodeGroups()), 1)
}

func TestNodeGroupForNode(t *testing.T) {
	node := &apiv1.Node{
		Spec: apiv1.NodeSpec{
			ProviderID: "azure:////123E4567-E89B-12D3-A456-426655440000",
		},
	}

	scaleSetVmClient := VirtualMachineScaleSetVMsClientMock{}

	var testAzureManager = &AzureManager{
		scaleSets:        make([]*scaleSetInformation, 0),
		scaleSetClient:   &VirtualMachineScaleSetsClientMock{},
		scaleSetVmClient: &scaleSetVmClient,
		scaleSetCache:    make(map[AzureRef]*ScaleSet),
	}

	provider := testProvider(t, testAzureManager)
	err := provider.addNodeGroup("1:5:test-asg")
	assert.NoError(t, err)

	assert.Equal(t, len(provider.scaleSets), 1)

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

func TestAzureRefFromProviderId(t *testing.T) {
	_, err := AzureRefFromProviderId("azure:///123")
	assert.Error(t, err)
	_, err = AzureRefFromProviderId("azure://test/rg/test-instance-id")
	assert.Error(t, err)

	// Example id: "azure:///subscriptions/subscriptionId/resourceGroups/kubernetes/providers/Microsoft.Compute/virtualMachines/kubernetes-master"
	azureRef, err := AzureRefFromProviderId("azure:////kubernetes-master")
	assert.NoError(t, err)
	assert.Equal(t, &AzureRef{
		Name: "kubernetes-master",
	}, azureRef)
}

func TestMaxSize(t *testing.T) {
	provider := testProvider(t, testAzureManager)
	err := provider.addNodeGroup("1:5:test-asg")
	assert.NoError(t, err)
	assert.Equal(t, len(provider.scaleSets), 1)
	assert.Equal(t, provider.scaleSets[0].MaxSize(), 5)
}

func TestMinSize(t *testing.T) {
	provider := testProvider(t, testAzureManager)
	err := provider.addNodeGroup("1:5:test-asg")
	assert.NoError(t, err)
	assert.Equal(t, len(provider.scaleSets), 1)
	assert.Equal(t, provider.scaleSets[0].MinSize(), 1)
}

func TestTargetSize(t *testing.T) {
	provider := testProvider(t, testAzureManager)
	err := provider.addNodeGroup("1:5:test-asg")
	assert.NoError(t, err)
	targetSize, err := provider.scaleSets[0].TargetSize()
	assert.Equal(t, targetSize, 2)
	assert.NoError(t, err)
}

func TestIncreaseSize(t *testing.T) {

	var testAzureManager = &AzureManager{

		scaleSets:        make([]*scaleSetInformation, 0),
		scaleSetClient:   &VirtualMachineScaleSetsClientMock{},
		scaleSetVmClient: &VirtualMachineScaleSetVMsClientMock{},
		scaleSetCache:    make(map[AzureRef]*ScaleSet),
	}

	provider := testProvider(t, testAzureManager)

	err := provider.addNodeGroup("1:5:test-asg")
	assert.NoError(t, err)
	assert.Equal(t, len(provider.scaleSets), 1)

	err = provider.scaleSets[0].IncreaseSize(1)
	assert.NoError(t, err)

}

func TestBelongs(t *testing.T) {

	var testAzureManager = &AzureManager{

		scaleSets:        make([]*scaleSetInformation, 0),
		scaleSetClient:   &VirtualMachineScaleSetsClientMock{},
		scaleSetVmClient: &VirtualMachineScaleSetVMsClientMock{},
		scaleSetCache:    make(map[AzureRef]*ScaleSet),
	}

	provider := testProvider(t, testAzureManager)
	err := provider.addNodeGroup("1:5:test-asg")
	assert.NoError(t, err)

	invalidNode := &apiv1.Node{
		Spec: apiv1.NodeSpec{
			ProviderID: "azure:///subscriptions/subscriptionId/resourceGroups/kubernetes/providers/Microsoft.Compute/virtualMachines/invalid-instance-id",
		},
	}

	_, err = provider.scaleSets[0].Belongs(invalidNode)
	assert.Error(t, err)

	validNode := &apiv1.Node{
		Spec: apiv1.NodeSpec{
			ProviderID: "azure:////123E4567-E89B-12D3-A456-426655440000",
		},
	}
	belongs, err := provider.scaleSets[0].Belongs(validNode)
	assert.Equal(t, belongs, true)
	assert.NoError(t, err)

}

func TestDeleteNodes(t *testing.T) {
	scaleSetClient := &VirtualMachineScaleSetsClientMock{}
	m := &AzureManager{
		scaleSets:        make([]*scaleSetInformation, 0),
		scaleSetClient:   scaleSetClient,
		scaleSetVmClient: &VirtualMachineScaleSetVMsClientMock{},
		scaleSetCache:    make(map[AzureRef]*ScaleSet),
	}

	//(resourceGroupName string, vmScaleSetName string,
	// vmInstanceIDs compute.VirtualMachineScaleSetVMInstanceRequiredIDs, cancel <-chan struct{})
	// (result autorest.Response, err error)
	//cancel := make(<-chan struct{})
	instanceIds := make([]string, 1)
	instanceIds[0] = "test-instance-id"

	//requiredIds := compute.VirtualMachineScaleSetVMInstanceRequiredIDs{
	//	InstanceIds: &instanceIds,
	//}
	response := autorest.Response{
		Response: &http.Response{
			Status: "OK",
		},
	}
	scaleSetClient.On("DeleteInstances", mock.Anything, "test-asg", mock.Anything, mock.Anything).Return(response, nil)

	provider := testProvider(t, m)
	err := provider.addNodeGroup("1:5:test-asg")
	assert.NoError(t, err)

	node := &apiv1.Node{
		Spec: apiv1.NodeSpec{
			ProviderID: "azure:////123E4567-E89B-12D3-A456-426655440000",
		},
	}
	err = provider.scaleSets[0].DeleteNodes([]*apiv1.Node{node})
	assert.NoError(t, err)
	scaleSetClient.AssertNumberOfCalls(t, "DeleteInstances", 1)
}

func TestId(t *testing.T) {
	provider := testProvider(t, testAzureManager)
	err := provider.addNodeGroup("1:5:test-asg")
	assert.NoError(t, err)
	assert.Equal(t, len(provider.scaleSets), 1)
	assert.Equal(t, provider.scaleSets[0].Id(), "test-asg")
}

func TestDebug(t *testing.T) {
	asg := ScaleSet{
		azureManager: testAzureManager,
		minSize:      5,
		maxSize:      55,
	}
	asg.Name = "test-scale-set"
	assert.Equal(t, asg.Debug(), "test-scale-set (5:55)")
}

func TestBuildAsg(t *testing.T) {
	_, err := buildScaleSet("a", nil)
	assert.Error(t, err)
	_, err = buildScaleSet("a:b:c", nil)
	assert.Error(t, err)
	_, err = buildScaleSet("1:", nil)
	assert.Error(t, err)
	_, err = buildScaleSet("1:2:", nil)
	assert.Error(t, err)

	_, err = buildScaleSet("-1:2:", nil)
	assert.Error(t, err)

	_, err = buildScaleSet("5:3:", nil)
	assert.Error(t, err)

	_, err = buildScaleSet("5:ddd:test-name", nil)
	assert.Error(t, err)

	asg, err := buildScaleSet("111:222:test-name", nil)
	assert.NoError(t, err)
	assert.Equal(t, 111, asg.MinSize())
	assert.Equal(t, 222, asg.MaxSize())
	assert.Equal(t, "test-name", asg.Name)
}
