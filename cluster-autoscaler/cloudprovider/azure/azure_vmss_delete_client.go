/*
Copyright 2025 The Kubernetes Authors.

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

//go:generate sh -c "mockgen -source=azure_vmss_delete_client.go -package=azure VMSSDeleteClient | cat ../../../hack/boilerplate/boilerplate.go.txt - > azure_mock_vmss_delete_client_test.go"

package azure

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v6"

	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/virtualmachinescalesetclient"
)

// VMSSDeleteClient is an interface for async VMSS operations.
// This interface wraps the azclient's VMSS client to access BeginDeleteInstances and BeginCreateOrUpdate
// via the embedded SDK client for async polling.
type VMSSDeleteClient interface {
	BeginDeleteInstances(ctx context.Context, resourceGroupName string, vmScaleSetName string, vmInstanceIDs armcompute.VirtualMachineScaleSetVMInstanceRequiredIDs, options *armcompute.VirtualMachineScaleSetsClientBeginDeleteInstancesOptions) (*runtime.Poller[armcompute.VirtualMachineScaleSetsClientDeleteInstancesResponse], error)
	BeginCreateOrUpdate(ctx context.Context, resourceGroupName string, vmScaleSetName string, parameters armcompute.VirtualMachineScaleSet, options *armcompute.VirtualMachineScaleSetsClientBeginCreateOrUpdateOptions) (*runtime.Poller[armcompute.VirtualMachineScaleSetsClientCreateOrUpdateResponse], error)
}

// vmssDeleteClientWrapper wraps the azclient's VMSS client.
type vmssDeleteClientWrapper struct {
	client *virtualmachinescalesetclient.Client
}

// NewVMSSDeleteClient creates a wrapper around the azclient's VMSS client.
// The azclient's Client struct embeds *armcompute.VirtualMachineScaleSetsClient,
// which provides access to BeginDeleteInstances.
func NewVMSSDeleteClient(vmssClient virtualmachinescalesetclient.Interface) VMSSDeleteClient {
	// Type assert to get the concrete Client which embeds the SDK client
	client, ok := vmssClient.(*virtualmachinescalesetclient.Client)
	if !ok {
		// This should not happen in production, but handle gracefully
		return nil
	}
	return &vmssDeleteClientWrapper{client: client}
}

// BeginDeleteInstances implements the VMSSDeleteClient interface.
func (w *vmssDeleteClientWrapper) BeginDeleteInstances(ctx context.Context, resourceGroupName string, vmScaleSetName string, vmInstanceIDs armcompute.VirtualMachineScaleSetVMInstanceRequiredIDs, options *armcompute.VirtualMachineScaleSetsClientBeginDeleteInstancesOptions) (*runtime.Poller[armcompute.VirtualMachineScaleSetsClientDeleteInstancesResponse], error) {
	return w.client.VirtualMachineScaleSetsClient.BeginDeleteInstances(ctx, resourceGroupName, vmScaleSetName, vmInstanceIDs, options)
}

// BeginCreateOrUpdate implements the VMSSDeleteClient interface.
func (w *vmssDeleteClientWrapper) BeginCreateOrUpdate(ctx context.Context, resourceGroupName string, vmScaleSetName string, parameters armcompute.VirtualMachineScaleSet, options *armcompute.VirtualMachineScaleSetsClientBeginCreateOrUpdateOptions) (*runtime.Poller[armcompute.VirtualMachineScaleSetsClientCreateOrUpdateResponse], error) {
	return w.client.VirtualMachineScaleSetsClient.BeginCreateOrUpdate(ctx, resourceGroupName, vmScaleSetName, parameters, options)
}
