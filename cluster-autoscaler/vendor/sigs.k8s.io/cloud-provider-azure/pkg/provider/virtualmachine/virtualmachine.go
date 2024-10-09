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
	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2022-08-01/compute"

	"k8s.io/utils/pointer"

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

		ID:        pointer.StringDeref(vm.ID, ""),
		Name:      pointer.StringDeref(vm.Name, ""),
		Type:      pointer.StringDeref(vm.Type, ""),
		Location:  pointer.StringDeref(vm.Location, ""),
		Tags:      stringMap(vm.Tags),
		Zones:     stringSlice(vm.Zones),
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

		ID:        pointer.StringDeref(vm.ID, ""),
		Name:      pointer.StringDeref(vm.Name, ""),
		Type:      pointer.StringDeref(vm.Type, ""),
		Location:  pointer.StringDeref(vm.Location, ""),
		Tags:      stringMap(vm.Tags),
		Zones:     stringSlice(vm.Zones),
		Plan:      vm.Plan,
		Resources: vm.Resources,

		SKU:                                vm.Sku,
		InstanceID:                         pointer.StringDeref(vm.InstanceID, ""),
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

func (vm *VirtualMachine) GetInstanceViewStatus() *[]compute.InstanceViewStatus {
	if vm.IsVirtualMachine() && vm.vm != nil &&
		vm.vm.VirtualMachineProperties != nil &&
		vm.vm.VirtualMachineProperties.InstanceView != nil {
		return vm.vm.VirtualMachineProperties.InstanceView.Statuses
	}
	if vm.IsVirtualMachineScaleSetVM() &&
		vm.vmssVM != nil &&
		vm.vmssVM.VirtualMachineScaleSetVMProperties != nil &&
		vm.vmssVM.VirtualMachineScaleSetVMProperties.InstanceView != nil {
		return vm.vmssVM.VirtualMachineScaleSetVMProperties.InstanceView.Statuses
	}
	return nil
}

func (vm *VirtualMachine) GetProvisioningState() string {
	if vm.IsVirtualMachine() && vm.vm != nil &&
		vm.vm.VirtualMachineProperties != nil &&
		vm.vm.VirtualMachineProperties.ProvisioningState != nil {
		return *vm.vm.VirtualMachineProperties.ProvisioningState
	}
	if vm.IsVirtualMachineScaleSetVM() &&
		vm.vmssVM != nil &&
		vm.vmssVM.VirtualMachineScaleSetVMProperties != nil &&
		vm.vmssVM.VirtualMachineScaleSetVMProperties.ProvisioningState != nil {
		return *vm.vmssVM.VirtualMachineScaleSetVMProperties.ProvisioningState
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

// stringSlice returns a string slice value for the passed string slice pointer. It returns a nil
// slice if the pointer is nil.
func stringSlice(s *[]string) []string {
	if s != nil {
		return *s
	}
	return nil
}
