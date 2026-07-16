/*
Copyright 2022 The Kubernetes Authors.

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

package virtualmachine

import (
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v7"
	"k8s.io/utils/ptr"

	"sigs.k8s.io/cloud-provider-azure/pkg/consts"
)

type Variant string

const (
	VariantVirtualMachine           Variant = "VirtualMachine"
	VariantVirtualMachineScaleSetVM Variant = "VirtualMachineScaleSetVM"
)

type Manage string

const (
	VMSS Manage = "vmss"
	VMAS Manage = "vmas"
)

type ManageOption = func(*VirtualMachine)

// ByVMSS specifies that the virtual machine is managed by a virtual machine scale set.
func ByVMSS(vmssName string) ManageOption {
	return func(vm *VirtualMachine) {
		vm.Manage = VMSS
		vm.VMSSName = vmssName
	}
}

type VirtualMachine struct {
	Variant Variant
	vm      *armcompute.VirtualMachine
	vmssVM  *armcompute.VirtualMachineScaleSetVM

	Manage   Manage
	VMSSName string

	// re-export fields
	// common fields
	ID        string
	Name      string
	Location  string
	Tags      map[string]string
	Zones     []*string
	Type      string
	Plan      *armcompute.Plan
	Resources []*armcompute.VirtualMachineExtension
	Etag      *string

	// fields of VirtualMachine
	Identity                 *armcompute.VirtualMachineIdentity
	VirtualMachineProperties *armcompute.VirtualMachineProperties

	// fields of VirtualMachineScaleSetVM
	InstanceID                         string
	SKU                                *armcompute.SKU
	VirtualMachineScaleSetVMProperties *armcompute.VirtualMachineScaleSetVMProperties
}

func FromVirtualMachine(vm *armcompute.VirtualMachine, opt ...ManageOption) *VirtualMachine {
	v := &VirtualMachine{
		vm:      vm,
		Variant: VariantVirtualMachine,

		ID:        ptr.Deref(vm.ID, ""),
		Name:      ptr.Deref(vm.Name, ""),
		Type:      ptr.Deref(vm.Type, ""),
		Location:  ptr.Deref(vm.Location, ""),
		Tags:      stringMap(vm.Tags),
		Zones:     vm.Zones,
		Plan:      vm.Plan,
		Resources: vm.Resources,

		Identity:                 vm.Identity,
		VirtualMachineProperties: vm.Properties,
	}

	for _, opt := range opt {
		opt(v)
	}

	return v
}

func FromVirtualMachineScaleSetVM(vm *armcompute.VirtualMachineScaleSetVM, opt ManageOption) *VirtualMachine {
	v := &VirtualMachine{
		Variant: VariantVirtualMachineScaleSetVM,
		vmssVM:  vm,

		ID:        ptr.Deref(vm.ID, ""),
		Name:      ptr.Deref(vm.Name, ""),
		Type:      ptr.Deref(vm.Type, ""),
		Location:  ptr.Deref(vm.Location, ""),
		Tags:      stringMap(vm.Tags),
		Zones:     vm.Zones,
		Plan:      vm.Plan,
		Resources: vm.Resources,
		Etag:      vm.Etag,

		SKU:                                vm.SKU,
		InstanceID:                         ptr.Deref(vm.InstanceID, ""),
		VirtualMachineScaleSetVMProperties: vm.Properties,
	}

	// TODO: should validate manage option
	// VirtualMachineScaleSetVM should always be managed by VMSS
	opt(v)

	return v
}

func (vm *VirtualMachine) IsVirtualMachine() bool {
	return vm.Variant == VariantVirtualMachine
}

func (vm *VirtualMachine) IsVirtualMachineScaleSetVM() bool {
	return vm.Variant == VariantVirtualMachineScaleSetVM
}

func (vm *VirtualMachine) ManagedByVMSS() bool {
	return vm.Manage == VMSS
}

func (vm *VirtualMachine) AsVirtualMachine() *armcompute.VirtualMachine {
	return vm.vm
}

func (vm *VirtualMachine) AsVirtualMachineScaleSetVM() *armcompute.VirtualMachineScaleSetVM {
	return vm.vmssVM
}

func (vm *VirtualMachine) GetInstanceViewStatus() []*armcompute.InstanceViewStatus {
	if vm.IsVirtualMachine() && vm.vm != nil &&
		vm.vm.Properties != nil &&
		vm.vm.Properties.InstanceView != nil {
		return vm.vm.Properties.InstanceView.Statuses
	}
	if vm.IsVirtualMachineScaleSetVM() &&
		vm.vmssVM != nil &&
		vm.vmssVM.Properties != nil &&
		vm.vmssVM.Properties.InstanceView != nil {
		return vm.vmssVM.Properties.InstanceView.Statuses
	}
	return nil
}

func (vm *VirtualMachine) GetProvisioningState() string {
	if vm.IsVirtualMachine() && vm.vm != nil &&
		vm.vm.Properties != nil &&
		vm.vm.Properties.ProvisioningState != nil {
		return *vm.vm.Properties.ProvisioningState
	}
	if vm.IsVirtualMachineScaleSetVM() &&
		vm.vmssVM != nil &&
		vm.vmssVM.Properties != nil &&
		vm.vmssVM.Properties.ProvisioningState != nil {
		return *vm.vmssVM.Properties.ProvisioningState
	}
	return consts.ProvisioningStateUnknown
}

// StringMap returns a map of strings built from the map of string pointers. The empty string is
// used for nil pointers.
func stringMap(msp map[string]*string) map[string]string {
	ms := make(map[string]string, len(msp))
	for k, sp := range msp {
		if sp != nil {
			ms[k] = *sp
		} else {
			ms[k] = ""
		}
	}
	return ms
}
