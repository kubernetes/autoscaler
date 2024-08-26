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

package provider

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v5"
	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2022-08-01/compute"
	"github.com/Azure/go-autorest/autorest/azure"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"k8s.io/utils/pointer"

	azcache "sigs.k8s.io/cloud-provider-azure/pkg/cache"
	"sigs.k8s.io/cloud-provider-azure/pkg/consts"
	"sigs.k8s.io/cloud-provider-azure/pkg/retry"
)

// AttachDisk attaches a disk to vm
func (ss *ScaleSet) AttachDisk(ctx context.Context, nodeName types.NodeName, diskMap map[string]*AttachDiskOptions) (*azure.Future, error) {
	vmName := mapNodeNameToVMName(nodeName)
	vm, err := ss.getVmssVM(vmName, azcache.CacheReadTypeDefault)
	if err != nil {
		return nil, err
	}

	nodeResourceGroup, err := ss.GetNodeResourceGroup(vmName)
	if err != nil {
		return nil, err
	}

	var disks []compute.DataDisk

	storageProfile := vm.AsVirtualMachineScaleSetVM().StorageProfile

	if storageProfile != nil && storageProfile.DataDisks != nil {
		disks = make([]compute.DataDisk, len(*storageProfile.DataDisks))
		copy(disks, *storageProfile.DataDisks)
	}

	for k, v := range diskMap {
		diskURI := k
		opt := v
		attached := false
		for _, disk := range *storageProfile.DataDisks {
			if disk.ManagedDisk != nil && strings.EqualFold(*disk.ManagedDisk.ID, diskURI) && disk.Lun != nil {
				if *disk.Lun == opt.Lun {
					attached = true
					break
				}
				return nil, fmt.Errorf("disk(%s) already attached to node(%s) on LUN(%d), but target LUN is %d", diskURI, nodeName, *disk.Lun, opt.Lun)

			}
		}
		if attached {
			klog.V(2).Infof("azureDisk - disk(%s) already attached to node(%s) on LUN(%d)", diskURI, nodeName, opt.Lun)
			continue
		}

		managedDisk := &compute.ManagedDiskParameters{ID: &diskURI}
		if opt.DiskEncryptionSetID == "" {
			if storageProfile.OsDisk != nil &&
				storageProfile.OsDisk.ManagedDisk != nil &&
				storageProfile.OsDisk.ManagedDisk.DiskEncryptionSet != nil &&
				storageProfile.OsDisk.ManagedDisk.DiskEncryptionSet.ID != nil {
				// set diskEncryptionSet as value of os disk by default
				opt.DiskEncryptionSetID = *storageProfile.OsDisk.ManagedDisk.DiskEncryptionSet.ID
			}
		}
		if opt.DiskEncryptionSetID != "" {
			managedDisk.DiskEncryptionSet = &compute.DiskEncryptionSetParameters{ID: &opt.DiskEncryptionSetID}
		}
		disks = append(disks,
			compute.DataDisk{
				Name:                    &opt.DiskName,
				Lun:                     &opt.Lun,
				Caching:                 opt.CachingMode,
				CreateOption:            "attach",
				ManagedDisk:             managedDisk,
				WriteAcceleratorEnabled: pointer.Bool(opt.WriteAcceleratorEnabled),
			})
	}

	newVM := compute.VirtualMachineScaleSetVM{
		VirtualMachineScaleSetVMProperties: &compute.VirtualMachineScaleSetVMProperties{
			StorageProfile: &compute.StorageProfile{
				DataDisks: &disks,
			},
		},
	}

	klog.V(2).Infof("azureDisk - update: rg(%s) vm(%s) - attach disk list(%+v)", nodeResourceGroup, nodeName, diskMap)
	future, rerr := ss.VirtualMachineScaleSetVMsClient.UpdateAsync(ctx, nodeResourceGroup, vm.VMSSName, vm.InstanceID, newVM, "attach_disk")
	if rerr != nil {
		klog.Errorf("azureDisk - attach disk list(%+v) on rg(%s) vm(%s) failed, err: %v", diskMap, nodeResourceGroup, nodeName, rerr)
		if rerr.HTTPStatusCode == http.StatusNotFound {
			klog.Errorf("azureDisk - begin to filterNonExistingDisks(%v) on rg(%s) vm(%s)", diskMap, nodeResourceGroup, nodeName)
			disks := FilterNonExistingDisks(ctx, ss.DisksClient, *newVM.VirtualMachineScaleSetVMProperties.StorageProfile.DataDisks)
			newVM.VirtualMachineScaleSetVMProperties.StorageProfile.DataDisks = &disks
			future, rerr = ss.VirtualMachineScaleSetVMsClient.UpdateAsync(ctx, nodeResourceGroup, vm.VMSSName, vm.InstanceID, newVM, "attach_disk")
		}
	}

	klog.V(2).Infof("azureDisk - update: rg(%s) vm(%s) - attach disk list(%+v) returned with %v", nodeResourceGroup, nodeName, diskMap, rerr)
	if rerr != nil {
		return future, rerr.Error()
	}
	return future, nil
}

// WaitForUpdateResult waits for the response of the update request
func (ss *ScaleSet) WaitForUpdateResult(ctx context.Context, future *azure.Future, nodeName types.NodeName, source string) error {
	vmName := mapNodeNameToVMName(nodeName)
	nodeResourceGroup, err := ss.GetNodeResourceGroup(vmName)
	if err != nil {
		return err
	}

	result, rerr := ss.VirtualMachineScaleSetVMsClient.WaitForUpdateResult(ctx, future, nodeResourceGroup, source)
	if rerr != nil {
		return rerr.Error()
	}

	var vmssName, instanceID string
	if result != nil && result.VirtualMachineScaleSetVMProperties != nil {
		// get vmssName, instanceID from cache first
		vm, err := ss.getVmssVM(vmName, azcache.CacheReadTypeDefault)
		if err == nil && vm != nil {
			vmssName = vm.VMSSName
			instanceID = vm.InstanceID
		} else {
			klog.Errorf("getVmssVM failed with error(%v) or nil pointer", err)
		}
	}

	// clean node cache first and then update cache
	_ = ss.DeleteCacheForNode(vmName)
	if vmssName != "" && instanceID != "" {
		if err := ss.updateCache(vmName, nodeResourceGroup, vmssName, instanceID, result); err != nil {
			klog.Errorf("updateCache(%s, %s, %s, %s) failed with error: %v", vmName, nodeResourceGroup, vmssName, instanceID, err)
		}
	}
	return nil
}

// DetachDisk detaches a disk from VM
func (ss *ScaleSet) DetachDisk(ctx context.Context, nodeName types.NodeName, diskMap map[string]string, forceDetach bool) error {
	vmName := mapNodeNameToVMName(nodeName)
	vm, err := ss.getVmssVM(vmName, azcache.CacheReadTypeDefault)
	if err != nil {
		return err
	}

	nodeResourceGroup, err := ss.GetNodeResourceGroup(vmName)
	if err != nil {
		return err
	}

	var disks []compute.DataDisk

	if vm != nil && vm.VirtualMachineScaleSetVMProperties != nil {
		storageProfile := vm.VirtualMachineScaleSetVMProperties.StorageProfile
		if storageProfile != nil && storageProfile.DataDisks != nil {
			disks = make([]compute.DataDisk, len(*storageProfile.DataDisks))
			copy(disks, *storageProfile.DataDisks)
		}
	}
	bFoundDisk := false
	for i, disk := range disks {
		for diskURI, diskName := range diskMap {
			if disk.Lun != nil && (disk.Name != nil && diskName != "" && strings.EqualFold(*disk.Name, diskName)) ||
				(disk.Vhd != nil && disk.Vhd.URI != nil && diskURI != "" && strings.EqualFold(*disk.Vhd.URI, diskURI)) ||
				(disk.ManagedDisk != nil && diskURI != "" && strings.EqualFold(*disk.ManagedDisk.ID, diskURI)) {
				// found the disk
				klog.V(2).Infof("azureDisk - detach disk: name %s uri %s", diskName, diskURI)
				disks[i].ToBeDetached = pointer.Bool(true)
				if forceDetach {
					disks[i].DetachOption = compute.ForceDetach
				}
				bFoundDisk = true
			}
		}
	}

	if !bFoundDisk {
		// only log here, next action is to update VM status with original meta data
		klog.Warningf("detach azure disk on node(%s): disk list(%s) not found", nodeName, diskMap)
	} else {
		if strings.EqualFold(ss.Environment.Name, consts.AzureStackCloudName) && !ss.Config.DisableAzureStackCloud {
			// Azure stack does not support ToBeDetached flag, use original way to detach disk
			var newDisks []compute.DataDisk
			for _, disk := range disks {
				if !pointer.BoolDeref(disk.ToBeDetached, false) {
					newDisks = append(newDisks, disk)
				}
			}
			disks = newDisks
		}
	}

	newVM := compute.VirtualMachineScaleSetVM{
		VirtualMachineScaleSetVMProperties: &compute.VirtualMachineScaleSetVMProperties{
			StorageProfile: &compute.StorageProfile{
				DataDisks: &disks,
			},
		},
	}

	var result *compute.VirtualMachineScaleSetVM
	var rerr *retry.Error

	defer func() {
		_ = ss.DeleteCacheForNode(vmName)

		// Update the cache with the updated result only if its not nil
		// and contains the VirtualMachineScaleSetVMProperties
		if rerr == nil && result != nil && result.VirtualMachineScaleSetVMProperties != nil {
			if err := ss.updateCache(vmName, nodeResourceGroup, vm.VMSSName, vm.InstanceID, result); err != nil {
				klog.Errorf("updateCache(%s, %s, %s, %s) failed with error: %v", vmName, nodeResourceGroup, vm.VMSSName, vm.InstanceID, err)
			}
		}
	}()

	klog.V(2).Infof("azureDisk - update(%s): vm(%s) - detach disk list(%s)", nodeResourceGroup, nodeName, diskMap)
	result, rerr = ss.VirtualMachineScaleSetVMsClient.Update(ctx, nodeResourceGroup, vm.VMSSName, vm.InstanceID, newVM,
		"detach_disk")
	if rerr != nil {
		klog.Errorf("azureDisk - detach disk list(%s) on rg(%s) vm(%s) failed, err: %v", diskMap, nodeResourceGroup, nodeName, rerr)
		if rerr.HTTPStatusCode == http.StatusNotFound {
			klog.Errorf("azureDisk - begin to filterNonExistingDisks(%v) on rg(%s) vm(%s)", diskMap, nodeResourceGroup, nodeName)
			disks := FilterNonExistingDisks(ctx, ss.DisksClient, *newVM.VirtualMachineScaleSetVMProperties.StorageProfile.DataDisks)
			newVM.VirtualMachineScaleSetVMProperties.StorageProfile.DataDisks = &disks
			result, rerr = ss.VirtualMachineScaleSetVMsClient.Update(ctx, nodeResourceGroup, vm.VMSSName, vm.InstanceID, newVM, "detach_disk")
		}
	}

	klog.V(2).Infof("azureDisk - update(%s): vm(%s) - detach disk(%v) returned with %v", nodeResourceGroup, nodeName, diskMap, rerr)
	if rerr != nil {
		return rerr.Error()
	}
	return nil
}

// UpdateVM updates a vm
func (ss *ScaleSet) UpdateVM(ctx context.Context, nodeName types.NodeName) error {
	future, err := ss.UpdateVMAsync(ctx, nodeName)
	if err != nil {
		return err
	}
	return ss.WaitForUpdateResult(ctx, future, nodeName, "update_vm")
}

// UpdateVMAsync updates a vm asynchronously
func (ss *ScaleSet) UpdateVMAsync(ctx context.Context, nodeName types.NodeName) (*azure.Future, error) {
	vmName := mapNodeNameToVMName(nodeName)
	vm, err := ss.getVmssVM(vmName, azcache.CacheReadTypeDefault)
	if err != nil {
		return nil, err
	}

	nodeResourceGroup, err := ss.GetNodeResourceGroup(vmName)
	if err != nil {
		return nil, err
	}

	future, rerr := ss.VirtualMachineScaleSetVMsClient.UpdateAsync(ctx, nodeResourceGroup, vm.VMSSName, vm.InstanceID, compute.VirtualMachineScaleSetVM{}, "update_vmss_instance")
	if rerr != nil {
		return future, rerr.Error()
	}
	return future, nil
}

// GetDataDisks gets a list of data disks attached to the node.
func (ss *ScaleSet) GetDataDisks(nodeName types.NodeName, crt azcache.AzureCacheReadType) ([]*armcompute.DataDisk, *string, error) {
	vm, err := ss.getVmssVM(string(nodeName), crt)
	if err != nil {
		return nil, nil, err
	}

	if vm != nil && vm.AsVirtualMachineScaleSetVM() != nil && vm.AsVirtualMachineScaleSetVM().VirtualMachineScaleSetVMProperties != nil {
		storageProfile := vm.AsVirtualMachineScaleSetVM().StorageProfile

		if storageProfile == nil || storageProfile.DataDisks == nil {
			return nil, nil, nil
		}
		result, err := ToArmcomputeDisk(*storageProfile.DataDisks)
		if err != nil {
			return nil, nil, err
		}
		return result, vm.AsVirtualMachineScaleSetVM().ProvisioningState, nil
	}

	return nil, nil, nil
}
