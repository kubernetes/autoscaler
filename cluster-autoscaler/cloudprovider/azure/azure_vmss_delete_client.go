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

package azure

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v5"
)

// VMSSDeleteClient is an interface for deleting VMSS instances.
// This interface wraps the raw SDK client to make it mockable for testing.
type VMSSDeleteClient interface {
	BeginDeleteInstances(ctx context.Context, resourceGroupName string, vmScaleSetName string, vmInstanceIDs armcompute.VirtualMachineScaleSetVMInstanceRequiredIDs, options *armcompute.VirtualMachineScaleSetsClientBeginDeleteInstancesOptions) (*runtime.Poller[armcompute.VirtualMachineScaleSetsClientDeleteInstancesResponse], error)
}

// vmssDeleteClientWrapper wraps the raw SDK client.
type vmssDeleteClientWrapper struct {
	client *armcompute.VirtualMachineScaleSetsClient
}

// NewVMSSDeleteClient creates a new wrapper around the raw SDK client.
func NewVMSSDeleteClient(client *armcompute.VirtualMachineScaleSetsClient) VMSSDeleteClient {
	return &vmssDeleteClientWrapper{client: client}
}

// BeginDeleteInstances implements the VMSSDeleteClient interface.
func (w *vmssDeleteClientWrapper) BeginDeleteInstances(ctx context.Context, resourceGroupName string, vmScaleSetName string, vmInstanceIDs armcompute.VirtualMachineScaleSetVMInstanceRequiredIDs, options *armcompute.VirtualMachineScaleSetsClientBeginDeleteInstancesOptions) (*runtime.Poller[armcompute.VirtualMachineScaleSetsClientDeleteInstancesResponse], error) {
	return w.client.BeginDeleteInstances(ctx, resourceGroupName, vmScaleSetName, vmInstanceIDs, options)
}
