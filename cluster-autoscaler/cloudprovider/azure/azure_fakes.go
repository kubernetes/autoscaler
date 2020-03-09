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
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-07-01/compute"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2017-05-10/resources"
	"github.com/Azure/azure-sdk-for-go/services/storage/mgmt/2018-07-01/storage"
	"github.com/Azure/go-autorest/autorest"
	"github.com/stretchr/testify/mock"
	"k8s.io/legacy-cloud-providers/azure/retry"
)

const (
	fakeVirtualMachineScaleSetVMID = "/subscriptions/test-subscription-id/resourceGroups/test-asg/providers/Microsoft.Compute/virtualMachineScaleSets/agents/virtualMachines/0"
)

// VirtualMachineScaleSetsClientMock mocks for VirtualMachineScaleSetsClient.
type VirtualMachineScaleSetsClientMock struct {
	mock.Mock
	mutex     sync.Mutex
	FakeStore map[string]map[string]compute.VirtualMachineScaleSet
}

// Get gets the VirtualMachineScaleSet by vmScaleSetName.
func (client *VirtualMachineScaleSetsClientMock) Get(ctx context.Context, resourceGroupName string, vmScaleSetName string) (result compute.VirtualMachineScaleSet, rerr *retry.Error) {
	capacity := int64(2)
	name := "Standard_D8_V3" // typo to test case-insensitive lookup
	location := "switzerlandwest"
	properties := compute.VirtualMachineScaleSetProperties{}
	return compute.VirtualMachineScaleSet{
		Name: &vmScaleSetName,
		Sku: &compute.Sku{
			Capacity: &capacity,
			Name:     &name,
		},
		Location:                         &location,
		VirtualMachineScaleSetProperties: &properties,
	}, nil
}

// CreateOrUpdate creates or updates the VirtualMachineScaleSet.
func (client *VirtualMachineScaleSetsClientMock) CreateOrUpdate(ctx context.Context, resourceGroupName string, VMScaleSetName string, parameters compute.VirtualMachineScaleSet) *retry.Error {
	client.mutex.Lock()
	defer client.mutex.Unlock()

	if _, ok := client.FakeStore[resourceGroupName]; !ok {
		client.FakeStore[resourceGroupName] = make(map[string]compute.VirtualMachineScaleSet)
	}
	client.FakeStore[resourceGroupName][VMScaleSetName] = parameters

	return nil
}

// DeleteInstances deletes a set of instances for specified VirtualMachineScaleSet.
func (client *VirtualMachineScaleSetsClientMock) DeleteInstances(ctx context.Context, resourceGroupName string, vmScaleSetName string, vmInstanceIDs compute.VirtualMachineScaleSetVMInstanceRequiredIDs) *retry.Error {
	args := client.Called(resourceGroupName, vmScaleSetName, vmInstanceIDs)
	if args.Error(1) != nil {
		return &retry.Error{RawError: args.Error(1)}
	}
	return nil
}

// List gets a list of VirtualMachineScaleSets.
func (client *VirtualMachineScaleSetsClientMock) List(ctx context.Context, resourceGroupName string) (result []compute.VirtualMachineScaleSet, rerr *retry.Error) {
	client.mutex.Lock()
	defer client.mutex.Unlock()

	result = []compute.VirtualMachineScaleSet{}
	if _, ok := client.FakeStore[resourceGroupName]; ok {
		for _, v := range client.FakeStore[resourceGroupName] {
			result = append(result, v)
		}
	}

	return result, nil
}

// VirtualMachineScaleSetVMsClientMock mocks for VirtualMachineScaleSetVMsClient.
type VirtualMachineScaleSetVMsClientMock struct {
	mock.Mock
}

// Get gets a VirtualMachineScaleSetVM by VMScaleSetName and instanceID.
func (m *VirtualMachineScaleSetVMsClientMock) Get(ctx context.Context, resourceGroupName string, VMScaleSetName string, instanceID string, expand compute.InstanceViewTypes) (result compute.VirtualMachineScaleSetVM, rerr *retry.Error) {
	ID := fakeVirtualMachineScaleSetVMID
	vmID := "123E4567-E89B-12D3-A456-426655440000"
	properties := compute.VirtualMachineScaleSetVMProperties{
		VMID: &vmID,
	}
	return compute.VirtualMachineScaleSetVM{
		ID:                                 &ID,
		InstanceID:                         &instanceID,
		VirtualMachineScaleSetVMProperties: &properties,
	}, nil
}

// List gets a list of VirtualMachineScaleSetVMs.
func (m *VirtualMachineScaleSetVMsClientMock) List(ctx context.Context, resourceGroupName string, virtualMachineScaleSetName string, expand string) (result []compute.VirtualMachineScaleSetVM, rerr *retry.Error) {
	ID := fakeVirtualMachineScaleSetVMID
	instanceID := "0"
	vmID := "123E4567-E89B-12D3-A456-426655440000"
	properties := compute.VirtualMachineScaleSetVMProperties{
		VMID: &vmID,
	}
	result = append(result, compute.VirtualMachineScaleSetVM{
		ID:                                 &ID,
		InstanceID:                         &instanceID,
		VirtualMachineScaleSetVMProperties: &properties,
	})

	return result, nil
}

// Update updates a  VirtualMachineScaleSetVM
func (m *VirtualMachineScaleSetVMsClientMock) Update(ctx context.Context, resourceGroupName string, VMScaleSetName string, instanceID string, parameters compute.VirtualMachineScaleSetVM, source string) *retry.Error {
	return nil
}

// UpdateVMs updates a list of VirtualMachineScaleSetVM from map[instanceID]compute.VirtualMachineScaleSetVM.
func (m *VirtualMachineScaleSetVMsClientMock) UpdateVMs(ctx context.Context, resourceGroupName string, VMScaleSetName string, instances map[string]compute.VirtualMachineScaleSetVM, source string) *retry.Error {
	return nil
}

// VirtualMachinesClientMock mocks for VirtualMachinesClient.
type VirtualMachinesClientMock struct {
	mock.Mock

	mutex     sync.Mutex
	FakeStore map[string]map[string]compute.VirtualMachine
}

// Get gets the VirtualMachine by VMName.
func (m *VirtualMachinesClientMock) Get(ctx context.Context, resourceGroupName string, VMName string, expand compute.InstanceViewTypes) (result compute.VirtualMachine, err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if _, ok := m.FakeStore[resourceGroupName]; ok {
		if entity, ok := m.FakeStore[resourceGroupName][VMName]; ok {
			return entity, nil
		}
	}
	return result, autorest.DetailedError{
		StatusCode: http.StatusNotFound,
		Message:    "Not such VM",
	}
}

// List gets a lit of VirtualMachine inside the resource group.
func (m *VirtualMachinesClientMock) List(ctx context.Context, resourceGroupName string) (result []compute.VirtualMachine, err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, ok := m.FakeStore[resourceGroupName]; ok {
		for _, v := range m.FakeStore[resourceGroupName] {
			result = append(result, v)
		}
	}

	return result, nil
}

// Delete deletes the VirtualMachine by VMName.
func (m *VirtualMachinesClientMock) Delete(ctx context.Context, resourceGroupName string, VMName string) (resp *http.Response, err error) {
	args := m.Called(resourceGroupName, VMName)
	return nil, args.Error(1)
}

// InterfacesClientMock mocks for InterfacesClient.
type InterfacesClientMock struct {
	mock.Mock
}

// Delete deletes the interface by networkInterfaceName.
func (m *InterfacesClientMock) Delete(ctx context.Context, resourceGroupName string, networkInterfaceName string) (resp *http.Response, err error) {
	args := m.Called(resourceGroupName, networkInterfaceName)
	return nil, args.Error(1)
}

// DisksClientMock mocks for DisksClient.
type DisksClientMock struct {
	mock.Mock
}

// Delete deletes the disk by diskName.
func (m *DisksClientMock) Delete(ctx context.Context, resourceGroupName string, diskName string) (resp *http.Response, err error) {
	args := m.Called(resourceGroupName, diskName)
	return nil, args.Error(1)
}

// AccountsClientMock mocks for AccountsClient.
type AccountsClientMock struct {
	mock.Mock
}

// ListKeys get a list of keys by accountName.
func (m *AccountsClientMock) ListKeys(ctx context.Context, resourceGroupName string, accountName string) (result storage.AccountListKeysResult, err error) {
	args := m.Called(resourceGroupName, accountName)
	return storage.AccountListKeysResult{}, args.Error(1)
}

// DeploymentsClientMock mocks for DeploymentsClient.
type DeploymentsClientMock struct {
	mock.Mock

	mutex     sync.Mutex
	FakeStore map[string]resources.DeploymentExtended
}

// Get gets the DeploymentExtended by deploymentName.
func (m *DeploymentsClientMock) Get(ctx context.Context, resourceGroupName string, deploymentName string) (result resources.DeploymentExtended, err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	deploy, ok := m.FakeStore[deploymentName]
	if !ok {
		return result, fmt.Errorf("deployment not found")
	}

	return deploy, nil
}

// ExportTemplate exports the deployment's template.
func (m *DeploymentsClientMock) ExportTemplate(ctx context.Context, resourceGroupName string, deploymentName string) (result resources.DeploymentExportResult, err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	deploy, ok := m.FakeStore[deploymentName]
	if !ok {
		return result, fmt.Errorf("deployment not found")
	}

	return resources.DeploymentExportResult{
		Template: deploy.Properties.Template,
	}, nil
}

// CreateOrUpdate creates or updates the Deployment.
func (m *DeploymentsClientMock) CreateOrUpdate(ctx context.Context, resourceGroupName string, deploymentName string, parameters resources.Deployment) (resp *http.Response, err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	deploy, ok := m.FakeStore[deploymentName]
	if !ok {
		deploy = resources.DeploymentExtended{
			Properties: &resources.DeploymentPropertiesExtended{},
		}
		m.FakeStore[deploymentName] = deploy
	}

	deploy.Properties.Parameters = parameters.Properties.Parameters
	deploy.Properties.Template = parameters.Properties.Template
	return nil, nil
}

// List gets all the deployments for a resource group.
func (m *DeploymentsClientMock) List(ctx context.Context, resourceGroupName, filter string, top *int32) (result []resources.DeploymentExtended, err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	result = make([]resources.DeploymentExtended, 0)
	for i := range m.FakeStore {
		result = append(result, m.FakeStore[i])
	}

	return result, nil
}

// Delete deletes the given deployment
func (m *DeploymentsClientMock) Delete(ctx context.Context, resourceGroupName, deploymentName string) (resp *http.Response, err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, ok := m.FakeStore[deploymentName]; !ok {
		return nil, fmt.Errorf("there is no such a deployment with name %s", deploymentName)
	}

	delete(m.FakeStore, deploymentName)

	return
}

func fakeVMSSWithTags(vmssName string, tags map[string]*string) compute.VirtualMachineScaleSet {
	skuName := "Standard_D4_v2"
	var vmssCapacity int64 = 3

	return compute.VirtualMachineScaleSet{
		Name: &vmssName,
		Sku: &compute.Sku{
			Capacity: &vmssCapacity,
			Name:     &skuName,
		},
		Tags: tags,
	}

}
