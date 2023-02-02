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
	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2022-03-01/compute"
	"github.com/Azure/go-autorest/autorest/to"
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
	vm      *compute.VirtualMachine
	vmssVM  *compute.VirtualMachineScaleSetVM

	Manage   Manage
	VMSSName string

	// re-export fields
	// common fields
	ID        string
	Name      string
	Location  string
	Tags      map[string]string
	Zones     []string
	Type      string
	Plan      *compute.Plan
	Resources *[]compute.VirtualMachineExtension

	// fields of VirtualMachine
	Identity                 *compute.VirtualMachineIdentity
	VirtualMachineProperties *compute.VirtualMachineProperties

	// fields of VirtualMachineScaleSetVM
	InstanceID                         string
	SKU                                *compute.Sku
	VirtualMachineScaleSetVMProperties *compute.VirtualMachineScaleSetVMProperties
}

func FromVirtualMachine(vm *compute.VirtualMachine, opt ...ManageOption) *VirtualMachine {
	v := &VirtualMachine{
		vm:      vm,
		Variant: VariantVirtualMachine,

		ID:        to.String(vm.ID),
		Name:      to.String(vm.Name),
		Type:      to.String(vm.Type),
		Location:  to.String(vm.Location),
		Tags:      to.StringMap(vm.Tags),
		Zones:     to.StringSlice(vm.Zones),
		Plan:      vm.Plan,
		Resources: vm.Resources,

		Identity:                 vm.Identity,
		VirtualMachineProperties: vm.VirtualMachineProperties,
	}

	for _, opt := range opt {
		opt(v)
	}

	return v
}

func FromVirtualMachineScaleSetVM(vm *compute.VirtualMachineScaleSetVM, opt ManageOption) *VirtualMachine {
	v := &VirtualMachine{
		Variant: VariantVirtualMachineScaleSetVM,
		vmssVM:  vm,

		ID:        to.String(vm.ID),
		Name:      to.String(vm.Name),
		Type:      to.String(vm.Type),
		Location:  to.String(vm.Location),
		Tags:      to.StringMap(vm.Tags),
		Zones:     to.StringSlice(vm.Zones),
		Plan:      vm.Plan,
		Resources: vm.Resources,

		SKU:                                vm.Sku,
		InstanceID:                         to.String(vm.InstanceID),
		VirtualMachineScaleSetVMProperties: vm.VirtualMachineScaleSetVMProperties,
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

func (vm *VirtualMachine) AsVirtualMachine() *compute.VirtualMachine {
	return vm.vm
}

func (vm *VirtualMachine) AsVirtualMachineScaleSetVM() *compute.VirtualMachineScaleSetVM {
	return vm.vmssVM
}
