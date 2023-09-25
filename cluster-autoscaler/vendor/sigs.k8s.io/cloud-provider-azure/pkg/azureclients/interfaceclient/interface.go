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

package interfaceclient

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2022-07-01/network"

	"sigs.k8s.io/cloud-provider-azure/pkg/retry"
)

const (
	// APIVersion is the API version for network.
	APIVersion = "2022-07-01"
	// AzureStackCloudAPIVersion is the API version for Azure Stack
	AzureStackCloudAPIVersion = "2018-11-01"

	// ComputeAPIVersion is the API version for compute. It is required to get VMSS network interface.
	ComputeAPIVersion = "2017-03-30"
	// AzureStackComputeAPIVersion is the API version for compute for Azure Stack. It is required to get VMSS network interface.
	AzureStackComputeAPIVersion = "2018-11-01"

	// AzureStackCloudName is the cloud name of Azure Stack
	AzureStackCloudName = "AZURESTACKCLOUD"
)

// Interface is the client interface for NetworkInterface.
// Don't forget to run "hack/update-mock-clients.sh" command to generate the mock client.
type Interface interface {
	// Get gets a network.Interface.
	Get(ctx context.Context, resourceGroupName string, networkInterfaceName string, expand string) (result network.Interface, rerr *retry.Error)

	// GetVirtualMachineScaleSetNetworkInterface gets a network.Interface of VMSS VM.
	GetVirtualMachineScaleSetNetworkInterface(ctx context.Context, resourceGroupName string, virtualMachineScaleSetName string, virtualmachineIndex string, networkInterfaceName string, expand string) (result network.Interface, rerr *retry.Error)

	// CreateOrUpdate creates or updates a network.Interface.
	CreateOrUpdate(ctx context.Context, resourceGroupName string, networkInterfaceName string, parameters network.Interface) *retry.Error

	// Delete deletes a network interface by name.
	Delete(ctx context.Context, resourceGroupName string, networkInterfaceName string) *retry.Error
}
