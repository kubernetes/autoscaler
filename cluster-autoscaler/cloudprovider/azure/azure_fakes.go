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
	"sync"

	"github.com/Azure/azure-sdk-for-go/arm/compute"
	"github.com/Azure/azure-sdk-for-go/arm/disk"
	"github.com/Azure/azure-sdk-for-go/arm/resources/resources"
	"github.com/Azure/azure-sdk-for-go/arm/storage"
	"github.com/Azure/go-autorest/autorest"
	"github.com/stretchr/testify/mock"
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
func (client *VirtualMachineScaleSetsClientMock) Get(resourceGroupName string,
	vmScaleSetName string) (result compute.VirtualMachineScaleSet, err error) {
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

// CreateOrUpdate creates or updates the VirtualMachineScaleSet.
func (client *VirtualMachineScaleSetsClientMock) CreateOrUpdate(resourceGroupName string, VMScaleSetName string, parameters compute.VirtualMachineScaleSet, cancel <-chan struct{}) (<-chan compute.VirtualMachineScaleSet, <-chan error) {
	client.mutex.Lock()
	defer client.mutex.Unlock()

	resultChan := make(chan compute.VirtualMachineScaleSet, 1)
	errChan := make(chan error, 1)
	var result compute.VirtualMachineScaleSet
	var err error
	defer func() {
		resultChan <- result
		errChan <- err
		close(resultChan)
		close(errChan)
	}()

	if _, ok := client.FakeStore[resourceGroupName]; !ok {
		client.FakeStore[resourceGroupName] = make(map[string]compute.VirtualMachineScaleSet)
	}
	client.FakeStore[resourceGroupName][VMScaleSetName] = parameters
	result = client.FakeStore[resourceGroupName][VMScaleSetName]
	result.Response.Response = &http.Response{
		StatusCode: http.StatusOK,
	}
	err = nil
	return resultChan, errChan
}

// DeleteInstances deletes a set of instances for specified VirtualMachineScaleSet.
func (client *VirtualMachineScaleSetsClientMock) DeleteInstances(resourceGroupName string, vmScaleSetName string,
	vmInstanceIDs compute.VirtualMachineScaleSetVMInstanceRequiredIDs, cancel <-chan struct{}) (<-chan compute.OperationStatusResponse, <-chan error) {
	args := client.Called(resourceGroupName, vmScaleSetName, vmInstanceIDs, cancel)
	errChan := make(chan error)
	go func() {
		errChan <- args.Error(1)
	}()
	return nil, errChan
}

// List get a list of VirtualMachineScaleSets.
func (client *VirtualMachineScaleSetsClientMock) List(resourceGroupName string) (result compute.VirtualMachineScaleSetListResult, err error) {
	client.mutex.Lock()
	defer client.mutex.Unlock()

	value := []compute.VirtualMachineScaleSet{}
	if _, ok := client.FakeStore[resourceGroupName]; ok {
		for _, v := range client.FakeStore[resourceGroupName] {
			value = append(value, v)
		}
	}

	result.Response.Response = &http.Response{
		StatusCode: http.StatusOK,
	}
	result.NextLink = nil
	result.Value = &value
	return result, nil
}

// ListNextResults gets more results of VirtualMachineScaleSets.
func (client *VirtualMachineScaleSetsClientMock) ListNextResults(lastResults compute.VirtualMachineScaleSetListResult) (result compute.VirtualMachineScaleSetListResult, err error) {
	return result, nil
}

// VirtualMachineScaleSetVMsClientMock mocks for VirtualMachineScaleSetVMsClient.
type VirtualMachineScaleSetVMsClientMock struct {
	mock.Mock
}

// Get gets a VirtualMachineScaleSetVM by VMScaleSetName and instanceID.
func (m *VirtualMachineScaleSetVMsClientMock) Get(resourceGroupName string, VMScaleSetName string, instanceID string) (result compute.VirtualMachineScaleSetVM, err error) {
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
func (m *VirtualMachineScaleSetVMsClientMock) List(resourceGroupName string, virtualMachineScaleSetName string, filter string, selectParameter string, expand string) (result compute.VirtualMachineScaleSetVMListResult, err error) {
	value := make([]compute.VirtualMachineScaleSetVM, 1)
	ID := fakeVirtualMachineScaleSetVMID
	instanceID := "0"
	vmID := "123E4567-E89B-12D3-A456-426655440000"
	properties := compute.VirtualMachineScaleSetVMProperties{
		VMID: &vmID,
	}
	value[0] = compute.VirtualMachineScaleSetVM{
		ID:                                 &ID,
		InstanceID:                         &instanceID,
		VirtualMachineScaleSetVMProperties: &properties,
	}

	return compute.VirtualMachineScaleSetVMListResult{
		Value: &value,
	}, nil
}

// ListNextResults gets more results from previous VirtualMachineScaleSetVMListResult.
func (m *VirtualMachineScaleSetVMsClientMock) ListNextResults(lastResults compute.VirtualMachineScaleSetVMListResult) (result compute.VirtualMachineScaleSetVMListResult, err error) {
	return compute.VirtualMachineScaleSetVMListResult{Value: nil}, nil
}

// VirtualMachinesClientMock mocks for VirtualMachinesClient.
type VirtualMachinesClientMock struct {
	mock.Mock

	mutex     sync.Mutex
	FakeStore map[string]map[string]compute.VirtualMachine
}

// CreateOrUpdate creates or updates the VirtualMachine.
func (m *VirtualMachinesClientMock) CreateOrUpdate(resourceGroupName string, VMName string, parameters compute.VirtualMachine, cancel <-chan struct{}) (<-chan compute.VirtualMachine, <-chan error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	resultChan := make(chan compute.VirtualMachine, 1)
	errChan := make(chan error, 1)
	var result compute.VirtualMachine
	var err error
	defer func() {
		resultChan <- result
		errChan <- err
		close(resultChan)
		close(errChan)
	}()
	if _, ok := m.FakeStore[resourceGroupName]; !ok {
		m.FakeStore[resourceGroupName] = make(map[string]compute.VirtualMachine)
	}
	m.FakeStore[resourceGroupName][VMName] = parameters
	result = m.FakeStore[resourceGroupName][VMName]
	result.Response.Response = &http.Response{
		StatusCode: http.StatusOK,
	}
	err = nil
	return resultChan, errChan
}

// Get gets the VirtualMachine by VMName.
func (m *VirtualMachinesClientMock) Get(resourceGroupName string, VMName string, expand compute.InstanceViewTypes) (result compute.VirtualMachine, err error) {
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
func (m *VirtualMachinesClientMock) List(resourceGroupName string) (result compute.VirtualMachineListResult, err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	var value []compute.VirtualMachine
	if _, ok := m.FakeStore[resourceGroupName]; ok {
		for _, v := range m.FakeStore[resourceGroupName] {
			value = append(value, v)
		}
	}
	result.Response.Response = &http.Response{
		StatusCode: http.StatusOK,
	}
	result.NextLink = nil
	result.Value = &value

	return result, nil
}

// ListNextResults gets more results from previous VirtualMachineListResult.
func (m *VirtualMachinesClientMock) ListNextResults(lastResults compute.VirtualMachineListResult) (result compute.VirtualMachineListResult, err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return compute.VirtualMachineListResult{}, nil
}

// Delete deletes the VirtualMachine by VMName.
func (m *VirtualMachinesClientMock) Delete(resourceGroupName string, VMName string, cancel <-chan struct{}) (<-chan compute.OperationStatusResponse, <-chan error) {
	args := m.Called(resourceGroupName, VMName, cancel)
	errChan := make(chan error)
	go func() {
		errChan <- args.Error(1)
	}()
	return nil, errChan
}

// InterfacesClientMock mocks for InterfacesClient.
type InterfacesClientMock struct {
	mock.Mock
}

// Delete deletes the interface by networkInterfaceName.
func (m *InterfacesClientMock) Delete(resourceGroupName string, networkInterfaceName string, cancel <-chan struct{}) (<-chan autorest.Response, <-chan error) {
	args := m.Called(resourceGroupName, networkInterfaceName, cancel)
	errChan := make(chan error)
	go func() {
		errChan <- args.Error(1)
	}()
	return nil, errChan
}

// DisksClientMock mocks for DisksClient.
type DisksClientMock struct {
	mock.Mock
}

// Delete deletes the disk by diskName.
func (m *DisksClientMock) Delete(resourceGroupName string, diskName string, cancel <-chan struct{}) (<-chan disk.OperationStatusResponse, <-chan error) {
	args := m.Called(resourceGroupName, diskName, cancel)
	errChan := make(chan error)
	go func() {
		errChan <- args.Error(1)
	}()
	return nil, errChan
}

// AccountsClientMock mocks for AccountsClient.
type AccountsClientMock struct {
	mock.Mock
}

// ListKeys get a list of keys by accountName.
func (m *AccountsClientMock) ListKeys(resourceGroupName string, accountName string) (result storage.AccountListKeysResult, err error) {
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
func (m *DeploymentsClientMock) Get(resourceGroupName string, deploymentName string) (result resources.DeploymentExtended, err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	deploy, ok := m.FakeStore[deploymentName]
	if !ok {
		return result, fmt.Errorf("deployment not found")
	}

	return deploy, nil
}

// ExportTemplate exports the deployment's template.
func (m *DeploymentsClientMock) ExportTemplate(resourceGroupName string, deploymentName string) (result resources.DeploymentExportResult, err error) {
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
func (m *DeploymentsClientMock) CreateOrUpdate(resourceGroupName string, deploymentName string, parameters resources.Deployment, cancel <-chan struct{}) (<-chan resources.DeploymentExtended, <-chan error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	errChan := make(chan error)
	go func() {
		errChan <- nil
	}()

	deploy, ok := m.FakeStore[deploymentName]
	if !ok {
		deploy = resources.DeploymentExtended{
			Properties: &resources.DeploymentPropertiesExtended{},
		}
		m.FakeStore[deploymentName] = deploy
	}

	deploy.Properties.Parameters = parameters.Properties.Parameters
	deploy.Properties.Template = parameters.Properties.Template
	return nil, errChan
}
