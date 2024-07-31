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
package virtualmachineclient

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	armcompute "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v5"

	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/utils"
)

// +azure:client:verbs=createorupdate;delete;list,resource=VirtualMachine,packageName=github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v5,packageAlias=armcompute,clientName=VirtualMachinesClient,expand=true,rateLimitKey=virtualMachineRateLimit
type Interface interface {
	utils.GetWithExpandFunc[armcompute.VirtualMachine]
	utils.CreateOrUpdateFunc[armcompute.VirtualMachine]
	utils.DeleteFunc[armcompute.VirtualMachine]
	utils.ListFunc[armcompute.VirtualMachine]
	InstanceView(ctx context.Context, resourceGroupName string, vmName string) (*armcompute.VirtualMachineInstanceView, error)
	ListVMInstanceView(ctx context.Context, resourceGroupName string) (result []*armcompute.VirtualMachine, rerr error)
	BeginAttachDetachDataDisks(ctx context.Context, resourceGroupName string, vmName string, parameters armcompute.AttachDetachDataDisksRequest, options *armcompute.VirtualMachinesClientBeginAttachDetachDataDisksOptions) (*runtime.Poller[armcompute.VirtualMachinesClientAttachDetachDataDisksResponse], error)
	BeginUpdate(ctx context.Context, resourceGroupName string, vmName string, parameters armcompute.VirtualMachineUpdate, options *armcompute.VirtualMachinesClientBeginUpdateOptions) (*runtime.Poller[armcompute.VirtualMachinesClientUpdateResponse], error)
}
