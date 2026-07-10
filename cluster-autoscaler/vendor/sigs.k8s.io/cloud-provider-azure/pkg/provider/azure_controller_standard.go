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
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v7"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"

	azcache "sigs.k8s.io/cloud-provider-azure/pkg/cache"
	"sigs.k8s.io/cloud-provider-azure/pkg/log"
	"sigs.k8s.io/cloud-provider-azure/pkg/util/errutils"
)

// AttachDisk attaches a disk to vm
func (as *availabilitySet) AttachDisk(ctx context.Context, nodeName types.NodeName, diskMap map[string]*AttachDiskOptions) error {
	logger := log.FromContextOrBackground(ctx).WithName("AttachDisk")
	vm, err := as.getVirtualMachine(ctx, nodeName, azcache.CacheReadTypeDefault)
	if err != nil {
		return err
	}

	vmName := mapNodeNameToVMName(nodeName)
	nodeResourceGroup, err := as.GetNodeResourceGroup(vmName)
	if err != nil {
		return err
	}

	disks := make([]*armcompute.DataDisk, len(vm.Properties.StorageProfile.DataDisks))
	copy(disks, vm.Properties.StorageProfile.DataDisks)

	for k, v := range diskMap {
		diskURI := k
		opt := v
		attached := false
		for _, disk := range vm.Properties.StorageProfile.DataDisks {
			if disk.ManagedDisk != nil && strings.EqualFold(*disk.ManagedDisk.ID, diskURI) && disk.Lun != nil {
				if *disk.Lun == opt.Lun {
					attached = true
					break
				}
				return fmt.Errorf("disk(%s) already attached to node(%s) on LUN(%d), but target LUN is %d", diskURI, nodeName, *disk.Lun, opt.Lun)
			}
		}
		if attached {
			logger.V(2).Info("azureDisk - disk already attached to node on LUN", "diskURI", diskURI, "nodeName", nodeName, "LUN", opt.Lun)
			continue
		}

		managedDisk := &armcompute.ManagedDiskParameters{ID: &diskURI}
		if opt.DiskEncryptionSetID == "" {
			if vm.Properties.StorageProfile.OSDisk != nil &&
				vm.Properties.StorageProfile.OSDisk.ManagedDisk != nil &&
				vm.Properties.StorageProfile.OSDisk.ManagedDisk.DiskEncryptionSet != nil &&
				vm.Properties.StorageProfile.OSDisk.ManagedDisk.DiskEncryptionSet.ID != nil {
				// set diskEncryptionSet as value of os disk by default
				opt.DiskEncryptionSetID = *vm.Properties.StorageProfile.OSDisk.ManagedDisk.DiskEncryptionSet.ID
			}
		}
		if opt.DiskEncryptionSetID != "" {
			managedDisk.DiskEncryptionSet = &armcompute.DiskEncryptionSetParameters{ID: &opt.DiskEncryptionSetID}
		}
		disks = append(disks,
			&armcompute.DataDisk{
				Name:                    &opt.DiskName,
				Lun:                     &opt.Lun,
				Caching:                 to.Ptr(opt.CachingMode),
				CreateOption:            to.Ptr(armcompute.DiskCreateOptionTypesAttach),
				ManagedDisk:             managedDisk,
				WriteAcceleratorEnabled: ptr.To(opt.WriteAcceleratorEnabled),
			})
	}

	newVM := armcompute.VirtualMachine{
		Properties: &armcompute.VirtualMachineProperties{
			StorageProfile: &armcompute.StorageProfile{
				DataDisks: disks,
			},
		},
		Location: vm.Location,
	}
	logger.V(2).Info("azureDisk - update: vm - attach disk list", "resourceGroup", nodeResourceGroup, "vmName", vmName, "diskMap", diskMap)

	result, rerr := as.ComputeClientFactory.GetVirtualMachineClient().CreateOrUpdate(ctx, nodeResourceGroup, vmName, newVM)
	if rerr != nil {
		logger.Error(rerr, "azureDisk - attach disk list failed", "diskMap", diskMap, "resourceGroup", nodeResourceGroup, "vmName", vmName)
		if exists, err := errutils.CheckResourceExistsFromAzcoreError(rerr); !exists && err == nil {
			logger.Error(err, "azureDisk - begin to filterNonExistingDisks", "diskMap", diskMap, "resourceGroup", nodeResourceGroup, "vmName", vmName)
			disks := FilterNonExistingDisks(ctx, as.ComputeClientFactory, newVM.Properties.StorageProfile.DataDisks)
			newVM.Properties.StorageProfile.DataDisks = disks
			result, rerr = as.ComputeClientFactory.GetVirtualMachineClient().CreateOrUpdate(ctx, nodeResourceGroup, vmName, newVM)
		}
	}

	logger.V(2).Info("azureDisk - update: vm - attach disk list returned", "resourceGroup", nodeResourceGroup, "vmName", vmName, "diskMap", diskMap, "error", err)

	if rerr == nil && result != nil {
		as.updateCache(vmName, result)
	} else {
		_ = as.DeleteCacheForNode(ctx, vmName)
	}
	return rerr
}

func (as *availabilitySet) DeleteCacheForNode(ctx context.Context, nodeName string) error {
	logger := log.FromContextOrBackground(ctx).WithName("DeleteCacheForNode")
	err := as.vmCache.Delete(nodeName)
	if err == nil {
		logger.V(2).Info("Succeed", "nodeName", nodeName)
	} else {
		logger.Error(err, "failed", "nodeName", nodeName)
	}
	return err
}

// DetachDisk detaches a disk from VM
func (as *availabilitySet) DetachDisk(ctx context.Context, nodeName types.NodeName, diskMap map[string]string, forceDetach bool) error {
	logger := log.FromContextOrBackground(ctx).WithName("DetachDisk")
	vm, err := as.getVirtualMachine(ctx, nodeName, azcache.CacheReadTypeDefault)
	if err != nil {
		// if host doesn't exist, no need to detach
		klog.Warningf("azureDisk - cannot find node %s, skip detaching disk list(%s)", nodeName, diskMap)
		return nil
	}

	vmName := mapNodeNameToVMName(nodeName)
	nodeResourceGroup, err := as.GetNodeResourceGroup(vmName)
	if err != nil {
		return err
	}

	disks := make([]*armcompute.DataDisk, len(vm.Properties.StorageProfile.DataDisks))
	copy(disks, vm.Properties.StorageProfile.DataDisks)

	bFoundDisk := false
	for i, disk := range disks {
		for diskURI, diskName := range diskMap {
			if disk.Lun != nil && (disk.Name != nil && diskName != "" && strings.EqualFold(*disk.Name, diskName)) ||
				(disk.Vhd != nil && disk.Vhd.URI != nil && diskURI != "" && strings.EqualFold(*disk.Vhd.URI, diskURI)) ||
				(disk.ManagedDisk != nil && diskURI != "" && strings.EqualFold(*disk.ManagedDisk.ID, diskURI)) {
				// found the disk
				logger.V(2).Info("azureDisk - detach disk", "diskName", diskName, "diskURI", diskURI)
				disks[i].ToBeDetached = ptr.To(true)
				if forceDetach {
					disks[i].DetachOption = to.Ptr(armcompute.DiskDetachOptionTypesForceDetach)
				}
				bFoundDisk = true
			}
		}
	}

	if !bFoundDisk {
		// only log here, next action is to update VM status with original meta data
		klog.Warningf("detach azure disk on node(%s): disk list(%s) not found", nodeName, diskMap)
	} else {
		if as.IsStackCloud() {
			// Azure stack does not support ToBeDetached flag, use original way to detach disk
			newDisks := []*armcompute.DataDisk{}
			for _, disk := range disks {
				if !ptr.Deref(disk.ToBeDetached, false) {
					newDisks = append(newDisks, disk)
				}
			}
			disks = newDisks
		}
	}

	newVM := armcompute.VirtualMachine{
		Properties: &armcompute.VirtualMachineProperties{
			StorageProfile: &armcompute.StorageProfile{
				DataDisks: disks,
			},
		},
		Location: vm.Location,
	}
	logger.V(2).Info("azureDisk - update: vm node - detach disk list", "resourceGroup", nodeResourceGroup, "vmName", vmName, "nodeName", nodeName, "diskMap", diskMap)

	result, err := as.ComputeClientFactory.GetVirtualMachineClient().CreateOrUpdate(ctx, nodeResourceGroup, vmName, newVM)
	if err != nil {
		logger.Error(err, "azureDisk - detach disk list failed", "diskMap", diskMap, "resourceGroup", nodeResourceGroup, "vmName", vmName)
		var exists bool
		if exists, err = errutils.CheckResourceExistsFromAzcoreError(err); !exists && err == nil {
			logger.Error(err, "azureDisk - begin to filterNonExistingDisks", "diskMap", diskMap, "resourceGroup", nodeResourceGroup, "vmName", vmName)
			disks := FilterNonExistingDisks(ctx, as.ComputeClientFactory, vm.Properties.StorageProfile.DataDisks)
			newVM.Properties.StorageProfile.DataDisks = disks
			result, err = as.ComputeClientFactory.GetVirtualMachineClient().CreateOrUpdate(ctx, nodeResourceGroup, vmName, newVM)
		}
	}

	logger.V(2).Info("azureDisk - update: vm - detach disk list returned with error", "resourceGroup", nodeResourceGroup, "vmName", vmName, "diskMap", diskMap, "error", err)

	if err == nil && result != nil {
		as.updateCache(vmName, result)
	} else {
		_ = as.DeleteCacheForNode(ctx, vmName)
	}
	return err
}

// UpdateVM updates a vm
func (as *availabilitySet) UpdateVM(ctx context.Context, nodeName types.NodeName) error {
	vmName := mapNodeNameToVMName(nodeName)
	nodeResourceGroup, err := as.GetNodeResourceGroup(vmName)
	if err != nil {
		return err
	}

	result, rerr := as.ComputeClientFactory.GetVirtualMachineClient().CreateOrUpdate(ctx, nodeResourceGroup, vmName, armcompute.VirtualMachine{})
	if rerr != nil {
		if exists, err := errutils.CheckResourceExistsFromAzcoreError(rerr); !exists && err == nil {
			// if the VM does not exist, we should not update the cache
			return nil
		}
	}

	if rerr == nil && result != nil {
		as.updateCache(vmName, result)
	} else {
		_ = as.DeleteCacheForNode(ctx, vmName)
	}
	return rerr
}

func (as *availabilitySet) updateCache(nodeName string, vm *armcompute.VirtualMachine) {
	logger := log.Background().WithName("updateCache")
	if nodeName == "" {
		logger.Error(nil, "updateCache failed with empty nodeName", "nodeName", nodeName)
		return
	}
	if vm == nil || vm.Properties == nil {
		logger.Error(nil, "updateCache failed with nil vm or vm.Properties", "nodeName", nodeName)
		return
	}
	as.vmCache.Update(nodeName, vm)
	logger.V(2).Info("updateCache successfully", "nodeName", nodeName)
}

// GetDataDisks gets a list of data disks attached to the node.
func (as *availabilitySet) GetDataDisks(ctx context.Context, nodeName types.NodeName, crt azcache.AzureCacheReadType) ([]*armcompute.DataDisk, *string, error) {
	vm, err := as.getVirtualMachine(ctx, nodeName, crt)
	if err != nil {
		return nil, nil, err
	}

	if vm == nil || vm.Properties.StorageProfile.DataDisks == nil {
		return nil, nil, nil
	}
	return vm.Properties.StorageProfile.DataDisks, vm.Properties.ProvisioningState, nil
}
