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

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2022-03-01/compute"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/to"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"

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
				if *disk.Lun == opt.lun {
					attached = true
					break
				} else {
					return nil, fmt.Errorf("disk(%s) already attached to node(%s) on LUN(%d), but target LUN is %d", diskURI, nodeName, *disk.Lun, opt.lun)
				}
			}
		}
		if attached {
			klog.V(2).Infof("azureDisk - disk(%s) already attached to node(%s) on LUN(%d)", diskURI, nodeName, opt.lun)
			continue
		}

		managedDisk := &compute.ManagedDiskParameters{ID: &diskURI}
		if opt.diskEncryptionSetID == "" {
			if storageProfile.OsDisk != nil &&
				storageProfile.OsDisk.ManagedDisk != nil &&
				storageProfile.OsDisk.ManagedDisk.DiskEncryptionSet != nil &&
				storageProfile.OsDisk.ManagedDisk.DiskEncryptionSet.ID != nil {
				// set diskEncryptionSet as value of os disk by default
				opt.diskEncryptionSetID = *storageProfile.OsDisk.ManagedDisk.DiskEncryptionSet.ID
			}
		}
		if opt.diskEncryptionSetID != "" {
			managedDisk.DiskEncryptionSet = &compute.DiskEncryptionSetParameters{ID: &opt.diskEncryptionSetID}
		}
		disks = append(disks,
			compute.DataDisk{
				Name:                    &opt.diskName,
				Lun:                     &opt.lun,
				Caching:                 opt.cachingMode,
				CreateOption:            "attach",
				ManagedDisk:             managedDisk,
				WriteAcceleratorEnabled: to.BoolPtr(opt.writeAcceleratorEnabled),
			})
	}

	newVM := compute.VirtualMachineScaleSetVM{
		VirtualMachineScaleSetVMProperties: &compute.VirtualMachineScaleSetVMProperties{
			StorageProfile: &compute.StorageProfile{
				DataDisks: &disks,
			},
		},
	}

	// Invalidate the cache right after updating
	defer func() {
		_ = ss.DeleteCacheForNode(vmName)
	}()

	klog.V(2).Infof("azureDisk - update(%s): vm(%s) - attach disk list(%s)", nodeResourceGroup, nodeName, diskMap)
	future, rerr := ss.VirtualMachineScaleSetVMsClient.UpdateAsync(ctx, nodeResourceGroup, vm.VMSSName, vm.InstanceID, newVM, "attach_disk")
	if rerr != nil {
		klog.Errorf("azureDisk - attach disk list(%s) on rg(%s) vm(%s) failed, err: %v", diskMap, nodeResourceGroup, nodeName, rerr)
		if rerr.HTTPStatusCode == http.StatusNotFound {
			klog.Errorf("azureDisk - begin to filterNonExistingDisks(%v) on rg(%s) vm(%s)", diskMap, nodeResourceGroup, nodeName)
			disks := ss.filterNonExistingDisks(ctx, *newVM.VirtualMachineScaleSetVMProperties.StorageProfile.DataDisks)
			newVM.VirtualMachineScaleSetVMProperties.StorageProfile.DataDisks = &disks
			future, rerr = ss.VirtualMachineScaleSetVMsClient.UpdateAsync(ctx, nodeResourceGroup, vm.VMSSName, vm.InstanceID, newVM, "attach_disk")
		}
	}

	klog.V(2).Infof("azureDisk - update(%s): vm(%s) - attach disk list(%s, %s) returned with %v", nodeResourceGroup, nodeName, diskMap, rerr)
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

	var result *compute.VirtualMachineScaleSetVM
	var rerr *retry.Error
	defer func() {
		if rerr == nil && result != nil && result.VirtualMachineScaleSetVMProperties != nil {
			// If we have an updated result, we update the vmss vm cache
			vm, err := ss.getVmssVM(vmName, azcache.CacheReadTypeDefault)
			if err != nil {
				return
			}
			_ = ss.updateCache(vmName, nodeResourceGroup, vm.VMSSName, vm.InstanceID, result)
		}
	}()

	result, rerr = ss.VirtualMachineScaleSetVMsClient.WaitForUpdateResult(ctx, future, nodeResourceGroup, source)
	if rerr != nil {
		return rerr.Error()
	}
	return nil
}

// DetachDisk detaches a disk from VM
func (ss *ScaleSet) DetachDisk(ctx context.Context, nodeName types.NodeName, diskMap map[string]string) error {
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
				disks[i].ToBeDetached = to.BoolPtr(true)
				bFoundDisk = true
			}
		}
	}

	if !bFoundDisk {
		// only log here, next action is to update VM status with original meta data
		klog.Errorf("detach azure disk on node(%s): disk list(%s) not found", nodeName, diskMap)
	} else {
		if strings.EqualFold(ss.cloud.Environment.Name, consts.AzureStackCloudName) && !ss.Config.DisableAzureStackCloud {
			// Azure stack does not support ToBeDetached flag, use original way to detach disk
			var newDisks []compute.DataDisk
			for _, disk := range disks {
				if !to.Bool(disk.ToBeDetached) {
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

	var updateResult *compute.VirtualMachineScaleSetVM
	var rerr *retry.Error

	defer func() {
		// If there is an error with Update operation,
		// invalidate the cache
		if rerr != nil {
			_ = ss.DeleteCacheForNode(vmName)
			return
		}

		// Update the cache with the updated result only if its not nil
		// and contains the VirtualMachineScaleSetVMProperties
		if updateResult != nil && updateResult.VirtualMachineScaleSetVMProperties != nil {
			if err := ss.updateCache(vmName, nodeResourceGroup, vm.VMSSName, vm.InstanceID, updateResult); err != nil {
				klog.Errorf("updateCache(%s, %s, %s) failed with error: %v", vmName, nodeResourceGroup, vm.VMSSName, vm.InstanceID, err)
				// if err faced during updating cache, invalidate the cache
				_ = ss.DeleteCacheForNode(vmName)
			}
		}
	}()

	klog.V(2).Infof("azureDisk - update(%s): vm(%s) - detach disk list(%s)", nodeResourceGroup, nodeName, diskMap)
	updateResult, rerr = ss.VirtualMachineScaleSetVMsClient.Update(ctx, nodeResourceGroup, vm.VMSSName, vm.InstanceID, newVM,
		"detach_disk")
	if rerr != nil {
		klog.Errorf("azureDisk - detach disk list(%s) on rg(%s) vm(%s) failed, err: %v", diskMap, nodeResourceGroup, nodeName, rerr)
		if rerr.HTTPStatusCode == http.StatusNotFound {
			klog.Errorf("azureDisk - begin to filterNonExistingDisks(%v) on rg(%s) vm(%s)", diskMap, nodeResourceGroup, nodeName)
			disks := ss.filterNonExistingDisks(ctx, *newVM.VirtualMachineScaleSetVMProperties.StorageProfile.DataDisks)
			newVM.VirtualMachineScaleSetVMProperties.StorageProfile.DataDisks = &disks
			updateResult, rerr = ss.VirtualMachineScaleSetVMsClient.Update(ctx, nodeResourceGroup, vm.VMSSName, vm.InstanceID, newVM, "detach_disk")
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
	vmName := mapNodeNameToVMName(nodeName)
	vm, err := ss.getVmssVM(vmName, azcache.CacheReadTypeDefault)
	if err != nil {
		return err
	}

	nodeResourceGroup, err := ss.GetNodeResourceGroup(vmName)
	if err != nil {
		return err
	}

	var updateResult *compute.VirtualMachineScaleSetVM
	var rerr *retry.Error

	defer func() {
		// If there is an error with Update operation,
		// invalidate the cache
		if rerr != nil {
			_ = ss.DeleteCacheForNode(vmName)
			return
		}

		// Update the cache with the updated result only if its not nil
		// and contains the VirtualMachineScaleSetVMProperties
		if updateResult != nil && updateResult.VirtualMachineScaleSetVMProperties != nil {
			if err := ss.updateCache(vmName, nodeResourceGroup, vm.VMSSName, vm.InstanceID, updateResult); err != nil {
				klog.Errorf("updateCache(%s, %s, %s) failed with error: %v", vmName, nodeResourceGroup, vm.VMSSName, vm.InstanceID, err)
				// if err faced during updating cache, invalidate the cache
				_ = ss.DeleteCacheForNode(vmName)
			}
		}
	}()

	klog.V(2).Infof("azureDisk - update(%s): vm(%s)", nodeResourceGroup, nodeName)
	updateResult, rerr = ss.VirtualMachineScaleSetVMsClient.Update(ctx, nodeResourceGroup, vm.VMSSName, vm.InstanceID, compute.VirtualMachineScaleSetVM{}, "update_vmss_instance")

	klog.V(2).Infof("azureDisk - update(%s): vm(%s) - returned with %v", nodeResourceGroup, nodeName, rerr)
	if rerr != nil {
		return rerr.Error()
	}
	return nil
}

// GetDataDisks gets a list of data disks attached to the node.
func (ss *ScaleSet) GetDataDisks(nodeName types.NodeName, crt azcache.AzureCacheReadType) ([]compute.DataDisk, *string, error) {
	vm, err := ss.getVmssVM(string(nodeName), crt)
	if err != nil {
		return nil, nil, err
	}

	if vm != nil && vm.AsVirtualMachineScaleSetVM() != nil && vm.AsVirtualMachineScaleSetVM().VirtualMachineScaleSetVMProperties != nil {
		storageProfile := vm.AsVirtualMachineScaleSetVM().StorageProfile

		if storageProfile == nil || storageProfile.DataDisks == nil {
			return nil, nil, nil
		}

		return *storageProfile.DataDisks, vm.AsVirtualMachineScaleSetVM().ProvisioningState, nil
	}

	return nil, nil, nil
}
