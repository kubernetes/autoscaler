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
)

// AttachDisk attaches a disk to vm
func (as *availabilitySet) AttachDisk(ctx context.Context, nodeName types.NodeName, diskMap map[string]*AttachDiskOptions) (*azure.Future, error) {
	vm, err := as.getVirtualMachine(nodeName, azcache.CacheReadTypeDefault)
	if err != nil {
		return nil, err
	}

	vmName := mapNodeNameToVMName(nodeName)
	nodeResourceGroup, err := as.GetNodeResourceGroup(vmName)
	if err != nil {
		return nil, err
	}

	disks := make([]compute.DataDisk, len(*vm.StorageProfile.DataDisks))
	copy(disks, *vm.StorageProfile.DataDisks)

	for k, v := range diskMap {
		diskURI := k
		opt := v
		attached := false
		for _, disk := range *vm.StorageProfile.DataDisks {
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
			if vm.StorageProfile.OsDisk != nil &&
				vm.StorageProfile.OsDisk.ManagedDisk != nil &&
				vm.StorageProfile.OsDisk.ManagedDisk.DiskEncryptionSet != nil &&
				vm.StorageProfile.OsDisk.ManagedDisk.DiskEncryptionSet.ID != nil {
				// set diskEncryptionSet as value of os disk by default
				opt.diskEncryptionSetID = *vm.StorageProfile.OsDisk.ManagedDisk.DiskEncryptionSet.ID
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

	newVM := compute.VirtualMachineUpdate{
		VirtualMachineProperties: &compute.VirtualMachineProperties{
			StorageProfile: &compute.StorageProfile{
				DataDisks: &disks,
			},
		},
	}
	klog.V(2).Infof("azureDisk - update(%s): vm(%s) - attach disk list(%s)", nodeResourceGroup, vmName, diskMap)
	// Invalidate the cache right after updating
	defer func() {
		_ = as.DeleteCacheForNode(vmName)
	}()

	future, rerr := as.VirtualMachinesClient.UpdateAsync(ctx, nodeResourceGroup, vmName, newVM, "attach_disk")
	if rerr != nil {
		klog.Errorf("azureDisk - attach disk list(%s) on rg(%s) vm(%s) failed, err: %v", diskMap, nodeResourceGroup, vmName, rerr)
		if rerr.HTTPStatusCode == http.StatusNotFound {
			klog.Errorf("azureDisk - begin to filterNonExistingDisks(%v) on rg(%s) vm(%s)", diskMap, nodeResourceGroup, vmName)
			disks := as.filterNonExistingDisks(ctx, *newVM.VirtualMachineProperties.StorageProfile.DataDisks)
			newVM.VirtualMachineProperties.StorageProfile.DataDisks = &disks
			future, rerr = as.VirtualMachinesClient.UpdateAsync(ctx, nodeResourceGroup, vmName, newVM, "attach_disk")
		}
	}

	klog.V(2).Infof("azureDisk - update(%s): vm(%s) - attach disk list(%s) returned with %v", nodeResourceGroup, vmName, diskMap, rerr)
	if rerr != nil {
		return future, rerr.Error()
	}
	return future, nil
}

func (as *availabilitySet) DeleteCacheForNode(nodeName string) error {
	_ = as.cloud.vmCache.Delete(nodeName)
	return nil
}

// WaitForUpdateResult waits for the response of the update request
func (as *availabilitySet) WaitForUpdateResult(ctx context.Context, future *azure.Future, nodeName types.NodeName, source string) error {
	vmName := mapNodeNameToVMName(nodeName)
	nodeResourceGroup, err := as.GetNodeResourceGroup(vmName)
	if err != nil {
		return err
	}

	if rerr := as.VirtualMachinesClient.WaitForUpdateResult(ctx, future, nodeResourceGroup, source); rerr != nil {
		return rerr.Error()
	}
	return nil
}

// DetachDisk detaches a disk from VM
func (as *availabilitySet) DetachDisk(ctx context.Context, nodeName types.NodeName, diskMap map[string]string) error {
	vm, err := as.getVirtualMachine(nodeName, azcache.CacheReadTypeDefault)
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

	disks := make([]compute.DataDisk, len(*vm.StorageProfile.DataDisks))
	copy(disks, *vm.StorageProfile.DataDisks)

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
		if strings.EqualFold(as.cloud.Environment.Name, consts.AzureStackCloudName) && !as.Config.DisableAzureStackCloud {
			// Azure stack does not support ToBeDetached flag, use original way to detach disk
			newDisks := []compute.DataDisk{}
			for _, disk := range disks {
				if !to.Bool(disk.ToBeDetached) {
					newDisks = append(newDisks, disk)
				}
			}
			disks = newDisks
		}
	}

	newVM := compute.VirtualMachineUpdate{
		VirtualMachineProperties: &compute.VirtualMachineProperties{
			StorageProfile: &compute.StorageProfile{
				DataDisks: &disks,
			},
		},
	}
	klog.V(2).Infof("azureDisk - update(%s): vm(%s) - detach disk list(%s)", nodeResourceGroup, vmName, nodeName, diskMap)
	// Invalidate the cache right after updating
	defer func() {
		_ = as.DeleteCacheForNode(vmName)
	}()

	rerr := as.VirtualMachinesClient.Update(ctx, nodeResourceGroup, vmName, newVM, "detach_disk")
	if rerr != nil {
		klog.Errorf("azureDisk - detach disk list(%s) on rg(%s) vm(%s) failed, err: %v", diskMap, nodeResourceGroup, vmName, rerr)
		if rerr.HTTPStatusCode == http.StatusNotFound {
			klog.Errorf("azureDisk - begin to filterNonExistingDisks(%v) on rg(%s) vm(%s)", diskMap, nodeResourceGroup, vmName)
			disks := as.filterNonExistingDisks(ctx, *vm.StorageProfile.DataDisks)
			newVM.VirtualMachineProperties.StorageProfile.DataDisks = &disks
			rerr = as.VirtualMachinesClient.Update(ctx, nodeResourceGroup, vmName, newVM, "detach_disk")
		}
	}

	klog.V(2).Infof("azureDisk - update(%s): vm(%s) - detach disk list(%s) returned with %v", nodeResourceGroup, vmName, diskMap, rerr)
	if rerr != nil {
		return rerr.Error()
	}
	return nil
}

// UpdateVM updates a vm
func (as *availabilitySet) UpdateVM(ctx context.Context, nodeName types.NodeName) error {
	vmName := mapNodeNameToVMName(nodeName)
	nodeResourceGroup, err := as.GetNodeResourceGroup(vmName)
	if err != nil {
		return err
	}
	klog.V(2).Infof("azureDisk - update(%s): vm(%s)", nodeResourceGroup, vmName)
	// Invalidate the cache right after updating
	defer func() {
		_ = as.DeleteCacheForNode(vmName)
	}()

	rerr := as.VirtualMachinesClient.Update(ctx, nodeResourceGroup, vmName, compute.VirtualMachineUpdate{}, "update_vm")
	klog.V(2).Infof("azureDisk - update(%s): vm(%s) - returned with %v", nodeResourceGroup, vmName, rerr)
	if rerr != nil {
		return rerr.Error()
	}
	return nil
}

// GetDataDisks gets a list of data disks attached to the node.
func (as *availabilitySet) GetDataDisks(nodeName types.NodeName, crt azcache.AzureCacheReadType) ([]compute.DataDisk, *string, error) {
	vm, err := as.getVirtualMachine(nodeName, crt)
	if err != nil {
		return nil, nil, err
	}

	if vm.StorageProfile.DataDisks == nil {
		return nil, nil, nil
	}

	return *vm.StorageProfile.DataDisks, vm.ProvisioningState, nil
}
