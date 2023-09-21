/*
Copyright 2020 The Kubernetes Authors.

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

package vmclient

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2022-08-01/compute"
	"github.com/Azure/go-autorest/autorest/azure"

	"sigs.k8s.io/cloud-provider-azure/pkg/retry"
)

const (
	// APIVersion is the API version for VirtualMachine.
	APIVersion = "2022-03-01"
	// AzureStackCloudAPIVersion is the API version for Azure Stack
	AzureStackCloudAPIVersion = "2017-12-01"
	// AzureStackCloudName is the cloud name of Azure Stack
	AzureStackCloudName = "AZURESTACKCLOUD"
)

// Interface is the client interface for VirtualMachines.
// Don't forget to run "hack/update-mock-clients.sh" command to generate the mock client.
type Interface interface {
	// Get gets a VirtualMachine.
	Get(ctx context.Context, resourceGroupName string, VMName string, expand compute.InstanceViewTypes) (compute.VirtualMachine, *retry.Error)

	// List gets a list of VirtualMachines in the resourceGroupName.
	List(ctx context.Context, resourceGroupName string) ([]compute.VirtualMachine, *retry.Error)

	// ListVmssFlexVMsWithoutInstanceView gets a list of VirtualMachine in the VMSS Flex without InstanceView.
	ListVmssFlexVMsWithoutInstanceView(ctx context.Context, vmssFlexID string) ([]compute.VirtualMachine, *retry.Error)

	// ListVmssFlexVMsWithOnlyInstanceView gets a list of VirtualMachine in the VMSS Flex with only InstanceView.
	ListVmssFlexVMsWithOnlyInstanceView(ctx context.Context, vmssFlexID string) ([]compute.VirtualMachine, *retry.Error)

	// CreateOrUpdate creates or updates a VirtualMachine.
	CreateOrUpdate(ctx context.Context, resourceGroupName string, VMName string, parameters compute.VirtualMachine, source string) *retry.Error

	// Update updates a VirtualMachine.
	Update(ctx context.Context, resourceGroupName string, VMName string, parameters compute.VirtualMachineUpdate, source string) (*compute.VirtualMachine, *retry.Error)

	// UpdateAsync updates a VirtualMachine asynchronously
	UpdateAsync(ctx context.Context, resourceGroupName string, VMName string, parameters compute.VirtualMachineUpdate, source string) (*azure.Future, *retry.Error)

	// WaitForUpdateResult waits for the response of the update request
	WaitForUpdateResult(ctx context.Context, future *azure.Future, resourceGroupName, source string) (*compute.VirtualMachine, *retry.Error)

	// Delete deletes a VirtualMachine.
	Delete(ctx context.Context, resourceGroupName string, VMName string) *retry.Error
}
