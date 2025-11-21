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
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v7"
)

// getFakeVMListPager creates a fake pager for listing VMs.
func getFakeVMListPager(vms []armcompute.VirtualMachine) *runtime.Pager[armcompute.VirtualMachinesClientListResponse] {
	// Convert to pointers
	vmPointers := make([]*armcompute.VirtualMachine, len(vms))
	for i := range vms {
		vmPointers[i] = &vms[i]
	}

	return runtime.NewPager(runtime.PagingHandler[armcompute.VirtualMachinesClientListResponse]{
		More: func(page armcompute.VirtualMachinesClientListResponse) bool {
			return false
		},
		Fetcher: func(ctx context.Context, page *armcompute.VirtualMachinesClientListResponse) (armcompute.VirtualMachinesClientListResponse, error) {
			return armcompute.VirtualMachinesClientListResponse{
				VirtualMachineListResult: armcompute.VirtualMachineListResult{
					Value: vmPointers,
				},
			}, nil
		},
	})
}

// getFakeVMSSListPager creates a fake pager for listing VMSSs.
func getFakeVMSSListPager(vmsss []armcompute.VirtualMachineScaleSet) *runtime.Pager[armcompute.VirtualMachineScaleSetsClientListResponse] {
	// Convert to pointers
	vmssPointers := make([]*armcompute.VirtualMachineScaleSet, len(vmsss))
	for i := range vmsss {
		vmssPointers[i] = &vmsss[i]
	}

	return runtime.NewPager(runtime.PagingHandler[armcompute.VirtualMachineScaleSetsClientListResponse]{
		More: func(page armcompute.VirtualMachineScaleSetsClientListResponse) bool {
			return false
		},
		Fetcher: func(ctx context.Context, page *armcompute.VirtualMachineScaleSetsClientListResponse) (armcompute.VirtualMachineScaleSetsClientListResponse, error) {
			return armcompute.VirtualMachineScaleSetsClientListResponse{
				VirtualMachineScaleSetListResult: armcompute.VirtualMachineScaleSetListResult{
					Value: vmssPointers,
				},
			}, nil
		},
	})
}

// getFakeVMSSVMListPager creates a fake pager for listing VMSS VMs.
func getFakeVMSSVMListPager(vms []armcompute.VirtualMachineScaleSetVM) *runtime.Pager[armcompute.VirtualMachineScaleSetVMsClientListResponse] {
	// Convert to pointers
	vmPointers := make([]*armcompute.VirtualMachineScaleSetVM, len(vms))
	for i := range vms {
		vmPointers[i] = &vms[i]
	}

	return runtime.NewPager(runtime.PagingHandler[armcompute.VirtualMachineScaleSetVMsClientListResponse]{
		More: func(page armcompute.VirtualMachineScaleSetVMsClientListResponse) bool {
			return false
		},
		Fetcher: func(ctx context.Context, page *armcompute.VirtualMachineScaleSetVMsClientListResponse) (armcompute.VirtualMachineScaleSetVMsClientListResponse, error) {
			return armcompute.VirtualMachineScaleSetVMsClientListResponse{
				VirtualMachineScaleSetVMListResult: armcompute.VirtualMachineScaleSetVMListResult{
					Value: vmPointers,
				},
			}, nil
		},
	})
}
