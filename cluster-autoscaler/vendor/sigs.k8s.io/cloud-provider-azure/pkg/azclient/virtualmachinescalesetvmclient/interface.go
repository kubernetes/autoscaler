/*
Copyright 2023 The Kubernetes Authors.

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

// +azure:enableclientgen:=true
package virtualmachinescalesetvmclient

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	armcompute "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v5"

	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/utils"
)

// +azure:client:verbs=get;delete,resource=VirtualMachineScaleSet,subResource=VirtualMachineScaleSetVM,packageName=github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v5,packageAlias=armcompute,clientName=VirtualMachineScaleSetVMsClient,expand=false
type Interface interface {
	utils.SubResourceGetFunc[armcompute.VirtualMachineScaleSetVM]
	utils.SubResourceDeleteFunc[armcompute.VirtualMachineScaleSetVM]
	utils.SubResourceListFunc[armcompute.VirtualMachineScaleSetVM]
	ListVMInstanceView(ctx context.Context, resourceGroupName string, parentResourceName string) (result []*armcompute.VirtualMachineScaleSetVM, rerr error)

	// Update updates a VirtualMachineScaleSetVM.
	Update(ctx context.Context, resourceGroupName string, VMScaleSetName string, instanceID string, parameters armcompute.VirtualMachineScaleSetVM) (*armcompute.VirtualMachineScaleSetVM, error)
	GetInstanceView(ctx context.Context, resourceGroupName string, vmScaleSetName string, instanceID string) (*armcompute.VirtualMachineScaleSetVMInstanceView, error)
	BeginUpdate(ctx context.Context, resourceGroupName string, vmScaleSetName string, instanceID string, parameters armcompute.VirtualMachineScaleSetVM, options *armcompute.VirtualMachineScaleSetVMsClientBeginUpdateOptions) (*runtime.Poller[armcompute.VirtualMachineScaleSetVMsClientUpdateResponse], error)
}
