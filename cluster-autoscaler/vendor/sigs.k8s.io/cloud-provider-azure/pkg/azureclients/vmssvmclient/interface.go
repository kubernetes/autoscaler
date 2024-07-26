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

package vmssvmclient

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2022-08-01/compute"
	"github.com/Azure/go-autorest/autorest/azure"

	"sigs.k8s.io/cloud-provider-azure/pkg/retry"
)

const (
	// APIVersion is the API version for VMSS.
	APIVersion = "2022-03-01"
	// AzureStackCloudAPIVersion is the API version for Azure Stack
	AzureStackCloudAPIVersion = "2019-07-01"
	// AzureStackCloudName is the cloud name of Azure Stack
	AzureStackCloudName = "AZURESTACKCLOUD"
)

// Interface is the client interface for VirtualMachineScaleSetVM.
// Don't forget to run "hack/update-mock-clients.sh" command to generate the mock client.
type Interface interface {
	// Get gets a VirtualMachineScaleSetVM.
	Get(ctx context.Context, resourceGroupName string, VMScaleSetName string, instanceID string, expand compute.InstanceViewTypes) (compute.VirtualMachineScaleSetVM, *retry.Error)

	// List gets a list of VirtualMachineScaleSetVMs in the virtualMachineScaleSet.
	List(ctx context.Context, resourceGroupName string, virtualMachineScaleSetName string, expand string) ([]compute.VirtualMachineScaleSetVM, *retry.Error)

	// Update updates a VirtualMachineScaleSetVM.
	Update(ctx context.Context, resourceGroupName string, VMScaleSetName string, instanceID string, parameters compute.VirtualMachineScaleSetVM, source string) (*compute.VirtualMachineScaleSetVM, *retry.Error)

	// UpdateAsync updates a VirtualMachineScaleSetVM asynchronously
	UpdateAsync(ctx context.Context, resourceGroupName string, VMScaleSetName string, instanceID string, parameters compute.VirtualMachineScaleSetVM, source string) (*azure.Future, *retry.Error)

	// WaitForUpdateResult waits for the response of the update request
	WaitForUpdateResult(ctx context.Context, future *azure.Future, resourceGroupName, source string) (*compute.VirtualMachineScaleSetVM, *retry.Error)

	// UpdateVMs updates a list of VirtualMachineScaleSetVM from map[instanceID]compute.VirtualMachineScaleSetVM.
	UpdateVMs(ctx context.Context, resourceGroupName string, VMScaleSetName string, instances map[string]compute.VirtualMachineScaleSetVM, source string, batchSize int) *retry.Error
}
