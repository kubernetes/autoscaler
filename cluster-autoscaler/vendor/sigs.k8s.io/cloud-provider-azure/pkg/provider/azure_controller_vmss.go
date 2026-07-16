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
	"sigs.k8s.io/cloud-provider-azure/pkg/consts"
	"sigs.k8s.io/cloud-provider-azure/pkg/log"
	"sigs.k8s.io/cloud-provider-azure/pkg/util/errutils"
)

// AttachDisk attaches a disk to vm
func (ss *ScaleSet) AttachDisk(ctx context.Context, nodeName types.NodeName, diskMap map[string]*AttachDiskOptions) error {
	logger := log.FromContextOrBackground(ctx).WithName("AttachDisk")
	vmName := mapNodeNameToVMName(nodeName)
	vm, err := ss.getVmssVM(ctx, vmName, azcache.CacheReadTypeDefault)
	if err != nil {
		return err
	}

	nodeResourceGroup, err := ss.GetNodeResourceGroup(vmName)
	if err != nil {
		return err
	}

	var disks []*armcompute.DataDisk

	storageProfile := vm.AsVirtualMachineScaleSetVM().Properties.StorageProfile
	if storageProfile != nil && storageProfile.DataDisks != nil {
		disks = make([]*armcompute.DataDisk, len(storageProfile.DataDisks))
		copy(disks, storageProfile.DataDisks)
	}

	for k, v := range diskMap {
		diskURI := k
		opt := v
		attached := false
		for _, disk := range storageProfile.DataDisks {
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
			if storageProfile.OSDisk != nil &&
				storageProfile.OSDisk.ManagedDisk != nil &&
				storageProfile.OSDisk.ManagedDisk.DiskEncryptionSet != nil &&
				storageProfile.OSDisk.ManagedDisk.DiskEncryptionSet.ID != nil {
				// set diskEncryptionSet as value of os disk by default
				opt.DiskEncryptionSetID = *storageProfile.OSDisk.ManagedDisk.DiskEncryptionSet.ID
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

	newVM := &armcompute.VirtualMachineScaleSetVM{
		Properties: &armcompute.VirtualMachineScaleSetVMProperties{
			StorageProfile: &armcompute.StorageProfile{
				DataDisks: disks,
			},
		},
	}

	logger.V(2).Info("azureDisk - update: rg vm - attach disk list", "resourceGroup", nodeResourceGroup, "nodeName", nodeName, "diskMap", diskMap)
	result, rerr := ss.ComputeClientFactory.GetVirtualMachineScaleSetVMClient().Update(ctx, nodeResourceGroup, vm.VMSSName, vm.InstanceID, *newVM)
	if rerr != nil {
		logger.Error(rerr, "azureDisk - attach disk list failed", "diskMap", diskMap, "resourceGroup", nodeResourceGroup, "nodeName", nodeName)
		if exists, err := errutils.CheckResourceExistsFromAzcoreError(rerr); !exists && !strings.Contains(rerr.Error(), consts.ParentResourceNotFoundMessageCode) && err == nil {
			logger.Error(err, "azureDisk - begin to filterNonExistingDisks", "diskMap", diskMap, "resourceGroup", nodeResourceGroup, "nodeName", nodeName)
			disks := FilterNonExistingDisks(ctx, ss.ComputeClientFactory, newVM.Properties.StorageProfile.DataDisks)
			newVM.Properties.StorageProfile.DataDisks = disks
			result, rerr = ss.ComputeClientFactory.GetVirtualMachineScaleSetVMClient().Update(ctx, nodeResourceGroup, vm.VMSSName, vm.InstanceID, *newVM)
		}
	}

	logger.V(2).Info("azureDisk - update: rg vm - attach disk list returned with error", "resourceGroup", nodeResourceGroup, "nodeName", nodeName, "diskMap", diskMap, "error", rerr)

	if rerr == nil && result != nil && result.Properties != nil {
		if err := ss.updateCache(ctx, vmName, nodeResourceGroup, vm.VMSSName, vm.InstanceID, result); err != nil {
			logger.Error(err, "updateCache failed", "vmName", vmName, "resourceGroup", nodeResourceGroup, "vmssName", vm.VMSSName, "instanceID", vm.InstanceID)
		}
	} else {
		_ = ss.DeleteCacheForNode(ctx, vmName)
	}
	return rerr
}

// DetachDisk detaches a disk from VM
func (ss *ScaleSet) DetachDisk(ctx context.Context, nodeName types.NodeName, diskMap map[string]string, forceDetach bool) error {
	logger := log.FromContextOrBackground(ctx).WithName("DetachDisk")
	vmName := mapNodeNameToVMName(nodeName)
	vm, err := ss.getVmssVM(ctx, vmName, azcache.CacheReadTypeDefault)
	if err != nil {
		return err
	}

	nodeResourceGroup, err := ss.GetNodeResourceGroup(vmName)
	if err != nil {
		return err
	}

	var disks []*armcompute.DataDisk

	if vm != nil && vm.VirtualMachineScaleSetVMProperties != nil {
		storageProfile := vm.VirtualMachineScaleSetVMProperties.StorageProfile
		if storageProfile != nil && storageProfile.DataDisks != nil {
			disks = make([]*armcompute.DataDisk, len(storageProfile.DataDisks))
			copy(disks, storageProfile.DataDisks)
		}
	}
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
		// No matching disks found to detach - skip the VM update.
		// This can occur when:
		// 1. The VM has no data disks attached
		// 2. The requested disk(s) to detach are not present on the VM
		// Skipping the update avoids unnecessary API calls and prevents generic updates to the VM
		// that can stay pending during critical failures (e.g., zone outages).
		// This update can block the client from seeing the successful detach of the disk(s).
		klog.Warningf("azureDisk - detach disk: VM update skipped as no disks to detach on node(%s) with diskMap(%v)", nodeName, diskMap)
		return nil
	}

	if ss.IsStackCloud() {
		// Azure stack does not support ToBeDetached flag, use original way to detach disk
		var newDisks []*armcompute.DataDisk
		for _, disk := range disks {
			if !ptr.Deref(disk.ToBeDetached, false) {
				newDisks = append(newDisks, disk)
			}
		}
		disks = newDisks
	}

	newVM := &armcompute.VirtualMachineScaleSetVM{
		Properties: &armcompute.VirtualMachineScaleSetVMProperties{
			StorageProfile: &armcompute.StorageProfile{
				DataDisks: disks,
			},
		},
	}

	logger.V(2).Info("azureDisk - update: vm - detach disk list", "resourceGroup", nodeResourceGroup, "nodeName", nodeName, "diskMap", diskMap)
	result, rerr := ss.ComputeClientFactory.GetVirtualMachineScaleSetVMClient().Update(ctx, nodeResourceGroup, vm.VMSSName, vm.InstanceID, *newVM)
	if rerr != nil {
		logger.Error(rerr, "azureDisk - detach disk list failed", "diskMap", diskMap, "resourceGroup", nodeResourceGroup, "nodeName", nodeName)
		if exists, err := errutils.CheckResourceExistsFromAzcoreError(rerr); !exists && !strings.Contains(rerr.Error(), consts.ParentResourceNotFoundMessageCode) && err == nil {
			logger.Error(err, "azureDisk - begin to filterNonExistingDisks", "diskMap", diskMap, "resourceGroup", nodeResourceGroup, "nodeName", nodeName)
			disks := FilterNonExistingDisks(ctx, ss.ComputeClientFactory, newVM.Properties.StorageProfile.DataDisks)
			newVM.Properties.StorageProfile.DataDisks = disks
			result, rerr = ss.ComputeClientFactory.GetVirtualMachineScaleSetVMClient().Update(ctx, nodeResourceGroup, vm.VMSSName, vm.InstanceID, *newVM)
		}
	}

	if rerr == nil && result != nil && result.Properties != nil {
		logger.V(2).Info("azureDisk - update: vm - detach disk returned successfully", "resourceGroup", nodeResourceGroup, "nodeName", nodeName, "diskMap", diskMap)
		if err := ss.updateCache(ctx, vmName, nodeResourceGroup, vm.VMSSName, vm.InstanceID, result); err != nil {
			logger.Error(err, "updateCache failed", "vmName", vmName, "resourceGroup", nodeResourceGroup, "vmssName", vm.VMSSName, "instanceID", vm.InstanceID)
		}
	} else {
		logger.V(2).Info("azureDisk - update: vm - detach disk returned with error", "resourceGroup", nodeResourceGroup, "nodeName", nodeName, "diskMap", diskMap, "error", rerr)
		_ = ss.DeleteCacheForNode(ctx, vmName)
	}
	return rerr
}

// UpdateVM updates a vm
func (ss *ScaleSet) UpdateVM(ctx context.Context, nodeName types.NodeName) error {
	vmName := mapNodeNameToVMName(nodeName)
	vm, err := ss.getVmssVM(ctx, vmName, azcache.CacheReadTypeDefault)
	if err != nil {
		return err
	}

	nodeResourceGroup, err := ss.GetNodeResourceGroup(vmName)
	if err != nil {
		return err
	}

	_, err = ss.ComputeClientFactory.GetVirtualMachineScaleSetVMClient().Update(ctx, nodeResourceGroup, vm.VMSSName, vm.InstanceID, armcompute.VirtualMachineScaleSetVM{})
	return err
}

// GetDataDisks gets a list of data disks attached to the node.
func (ss *ScaleSet) GetDataDisks(ctx context.Context, nodeName types.NodeName, crt azcache.AzureCacheReadType) ([]*armcompute.DataDisk, *string, error) {
	vm, err := ss.getVmssVM(ctx, string(nodeName), crt)
	if err != nil {
		return nil, nil, err
	}

	if vm != nil && vm.AsVirtualMachineScaleSetVM() != nil && vm.AsVirtualMachineScaleSetVM().Properties != nil {
		storageProfile := vm.AsVirtualMachineScaleSetVM().Properties.StorageProfile

		if storageProfile == nil || storageProfile.DataDisks == nil {
			return nil, nil, nil
		}
		return storageProfile.DataDisks, vm.AsVirtualMachineScaleSetVM().Properties.ProvisioningState, nil
	}

	return nil, nil, nil
}
