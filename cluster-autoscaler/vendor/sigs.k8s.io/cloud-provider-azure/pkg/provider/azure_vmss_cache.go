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
	"sync"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2022-08-01/compute"

	"k8s.io/klog/v2"
	"k8s.io/utils/pointer"

	azcache "sigs.k8s.io/cloud-provider-azure/pkg/cache"
	"sigs.k8s.io/cloud-provider-azure/pkg/consts"
	utilsets "sigs.k8s.io/cloud-provider-azure/pkg/util/sets"
)

type VMSSVirtualMachineEntry struct {
	ResourceGroup  string
	VMSSName       string
	InstanceID     string
	VirtualMachine *compute.VirtualMachineScaleSetVM
	LastUpdate     time.Time
}

type VMSSEntry struct {
	VMSS          *compute.VirtualMachineScaleSet
	ResourceGroup string
	LastUpdate    time.Time
}

type NonVmssUniformNodesEntry struct {
	VMSSFlexVMNodeNames   *utilsets.IgnoreCaseSet
	VMSSFlexVMProviderIDs *utilsets.IgnoreCaseSet
	AvSetVMNodeNames      *utilsets.IgnoreCaseSet
	AvSetVMProviderIDs    *utilsets.IgnoreCaseSet
	ClusterNodeNames      *utilsets.IgnoreCaseSet
}

type VMManagementType string

const (
	ManagedByVmssUniform  VMManagementType = "ManagedByVmssUniform"
	ManagedByVmssFlex     VMManagementType = "ManagedByVmssFlex"
	ManagedByAvSet        VMManagementType = "ManagedByAvSet"
	ManagedByUnknownVMSet VMManagementType = "ManagedByUnknownVMSet"
)

func (ss *ScaleSet) newVMSSCache(ctx context.Context) (azcache.Resource, error) {
	getter := func(key string) (interface{}, error) {
		localCache := &sync.Map{} // [vmssName]*vmssEntry

		allResourceGroups, err := ss.GetResourceGroups()
		if err != nil {
			return nil, err
		}

		resourceGroupNotFound := false
		for _, resourceGroup := range allResourceGroups.UnsortedList() {
			allScaleSets, rerr := ss.VirtualMachineScaleSetsClient.List(ctx, resourceGroup)
			if rerr != nil {
				if rerr.IsNotFound() {
					klog.Warningf("Skip caching vmss for resource group %s due to error: %v", resourceGroup, rerr.Error())
					resourceGroupNotFound = true
					continue
				}
				klog.Errorf("VirtualMachineScaleSetsClient.List failed: %v", rerr)
				return nil, rerr.Error()
			}

			for i := range allScaleSets {
				scaleSet := allScaleSets[i]
				if scaleSet.Name == nil || *scaleSet.Name == "" {
					klog.Warning("failed to get the name of VMSS")
					continue
				}
				if scaleSet.OrchestrationMode == "" || scaleSet.OrchestrationMode == compute.Uniform {
					localCache.Store(*scaleSet.Name, &VMSSEntry{
						VMSS:          &scaleSet,
						ResourceGroup: resourceGroup,
						LastUpdate:    time.Now().UTC(),
					})
				}
			}
		}

		if !ss.Cloud.Config.DisableAPICallCache {
			if resourceGroupNotFound {
				// gc vmss vm cache when there is resource group not found
				vmssVMKeys := ss.vmssVMCache.GetStore().ListKeys()
				for _, cacheKey := range vmssVMKeys {
					vmssName := cacheKey[strings.LastIndex(cacheKey, "/")+1:]
					if _, ok := localCache.Load(vmssName); !ok {
						klog.V(2).Infof("remove vmss %s from vmssVMCache due to rg not found", cacheKey)
						_ = ss.vmssVMCache.Delete(cacheKey)
					}
				}
			}
		}
		return localCache, nil
	}

	if ss.Config.VmssCacheTTLInSeconds == 0 {
		ss.Config.VmssCacheTTLInSeconds = consts.VMSSCacheTTLDefaultInSeconds
	}
	return azcache.NewTimedCache(time.Duration(ss.Config.VmssCacheTTLInSeconds)*time.Second, getter, ss.Config.DisableAPICallCache)
}

func (ss *ScaleSet) getVMSSVMsFromCache(resourceGroup, vmssName string, crt azcache.AzureCacheReadType) (*sync.Map, error) {
	cacheKey := getVMSSVMCacheKey(resourceGroup, vmssName)
	entry, err := ss.vmssVMCache.Get(cacheKey, crt)
	if err != nil {
		return nil, err
	}

	if entry == nil {
		err = fmt.Errorf("vmssVMCache entry for resourceGroup (%s), vmssName (%s) returned nil data", resourceGroup, vmssName)
		return nil, err
	}

	virtualMachines := entry.(*sync.Map)
	return virtualMachines, nil
}

// newVMSSVirtualMachinesCache instantiates a new VMs cache for VMs belonging to the provided VMSS.
func (ss *ScaleSet) newVMSSVirtualMachinesCache() (azcache.Resource, error) {
	vmssVirtualMachinesCacheTTL := time.Duration(ss.Config.VmssVirtualMachinesCacheTTLInSeconds) * time.Second

	getter := func(cacheKey string) (interface{}, error) {
		localCache := &sync.Map{} // [nodeName]*VMSSVirtualMachineEntry
		oldCache := make(map[string]*VMSSVirtualMachineEntry)

		if !ss.Cloud.Config.DisableAPICallCache {
			entry, exists, err := ss.vmssVMCache.GetStore().GetByKey(cacheKey)
			if err != nil {
				return nil, err
			}
			if exists {
				cached := entry.(*azcache.AzureCacheEntry).Data
				if cached != nil {
					virtualMachines := cached.(*sync.Map)
					virtualMachines.Range(func(key, value interface{}) bool {
						oldCache[key.(string)] = value.(*VMSSVirtualMachineEntry)
						return true
					})
				}
			}
		}

		result := strings.Split(cacheKey, "/")
		if len(result) < 2 {
			err := fmt.Errorf("invalid cacheKey (%s)", cacheKey)
			return nil, err
		}

		resourceGroupName, vmssName := result[0], result[1]

		vms, err := ss.listScaleSetVMs(vmssName, resourceGroupName)
		if err != nil {
			return nil, err
		}

		for i := range vms {
			vm := vms[i]
			if vm.OsProfile == nil || vm.OsProfile.ComputerName == nil {
				klog.Warningf("failed to get computerName for vmssVM (%q)", vmssName)
				continue
			}

			computerName := strings.ToLower(*vm.OsProfile.ComputerName)
			if vm.NetworkProfile == nil || vm.NetworkProfile.NetworkInterfaces == nil {
				klog.Warningf("skip caching vmssVM %s since its network profile hasn't initialized yet (probably still under creating)", computerName)
				continue
			}

			vmssVMCacheEntry := &VMSSVirtualMachineEntry{
				ResourceGroup:  resourceGroupName,
				VMSSName:       vmssName,
				InstanceID:     pointer.StringDeref(vm.InstanceID, ""),
				VirtualMachine: &vm,
				LastUpdate:     time.Now().UTC(),
			}
			// set cache entry to nil when the VM is under deleting.
			if vm.VirtualMachineScaleSetVMProperties != nil &&
				strings.EqualFold(pointer.StringDeref(vm.VirtualMachineScaleSetVMProperties.ProvisioningState, ""), string(consts.ProvisioningStateDeleting)) {
				klog.V(4).Infof("VMSS virtualMachine %q is under deleting, setting its cache to nil", computerName)
				vmssVMCacheEntry.VirtualMachine = nil
			}
			localCache.Store(computerName, vmssVMCacheEntry)

			if !ss.Cloud.Config.DisableAPICallCache {
				delete(oldCache, computerName)
			}
		}

		if !ss.Cloud.Config.DisableAPICallCache {
			// add old missing cache data with nil entries to prevent aggressive
			// ARM calls during cache invalidation
			for name, vmEntry := range oldCache {
				// if the nil cache entry has existed for vmssVirtualMachinesCacheTTL in the cache
				// then it should not be added back to the cache
				if vmEntry.VirtualMachine == nil && time.Since(vmEntry.LastUpdate) > vmssVirtualMachinesCacheTTL {
					klog.V(5).Infof("ignoring expired entries from old cache for %s", name)
					continue
				}
				LastUpdate := time.Now().UTC()
				if vmEntry.VirtualMachine == nil {
					// if this is already a nil entry then keep the time the nil
					// entry was first created, so we can cleanup unwanted entries
					LastUpdate = vmEntry.LastUpdate
				}

				klog.V(5).Infof("adding old entries to new cache for %s", name)
				localCache.Store(name, &VMSSVirtualMachineEntry{
					ResourceGroup:  vmEntry.ResourceGroup,
					VMSSName:       vmEntry.VMSSName,
					InstanceID:     vmEntry.InstanceID,
					VirtualMachine: nil,
					LastUpdate:     LastUpdate,
				})
			}
		}

		return localCache, nil
	}

	return azcache.NewTimedCache(vmssVirtualMachinesCacheTTL, getter, ss.Cloud.Config.DisableAPICallCache)
}

// DeleteCacheForNode deletes Node from VMSS VM and VM caches.
func (ss *ScaleSet) DeleteCacheForNode(nodeName string) error {
	if ss.Config.DisableAPICallCache {
		return nil
	}
	vmManagementType, err := ss.getVMManagementTypeByNodeName(nodeName, azcache.CacheReadTypeUnsafe)
	if err != nil {
		klog.Errorf("getVMManagementTypeByNodeName(%s) failed with %v", nodeName, err)
		return err
	}

	if vmManagementType == ManagedByAvSet {
		// vm is managed by availability set.
		return ss.availabilitySet.DeleteCacheForNode(nodeName)
	}
	if vmManagementType == ManagedByVmssFlex {
		// vm is managed by vmss flex.
		return ss.flexScaleSet.DeleteCacheForNode(nodeName)
	}

	node, err := ss.getNodeIdentityByNodeName(nodeName, azcache.CacheReadTypeUnsafe)
	if err != nil {
		klog.Errorf("getNodeIdentityByNodeName(%s) failed with %v", nodeName, err)
		return err
	}
	// get sync.Map cache and remove the node from the cache
	cacheKey := getVMSSVMCacheKey(node.resourceGroup, node.vmssName)
	ss.lockMap.LockEntry(cacheKey)
	defer ss.lockMap.UnlockEntry(cacheKey)

	virtualMachines, err := ss.getVMSSVMsFromCache(node.resourceGroup, node.vmssName, azcache.CacheReadTypeUnsafe)
	if err != nil {
		klog.Errorf("getVMSSVMsFromCache(%s, %s) failed with %v", node.resourceGroup, node.vmssName, err)
		return err
	}

	virtualMachines.Delete(nodeName)
	ss.vmssVMCache.Update(cacheKey, virtualMachines)
	klog.V(2).Infof("DeleteCacheForNode(%s, %s, %s) successfully", node.resourceGroup, node.vmssName, nodeName)
	return nil
}

func (ss *ScaleSet) updateCache(nodeName, resourceGroupName, vmssName, instanceID string, updatedVM *compute.VirtualMachineScaleSetVM) error {
	// lock the VMSS entry to ensure a consistent view of the VM map when there are concurrent updates.
	cacheKey := getVMSSVMCacheKey(resourceGroupName, vmssName)
	ss.lockMap.LockEntry(cacheKey)
	defer ss.lockMap.UnlockEntry(cacheKey)

	virtualMachines, err := ss.getVMSSVMsFromCache(resourceGroupName, vmssName, azcache.CacheReadTypeUnsafe)
	if err != nil {
		return fmt.Errorf("updateCache(%s, %s, %s) failed getting vmCache with error: %w", vmssName, resourceGroupName, nodeName, err)
	}

	vmssVMCacheEntry := &VMSSVirtualMachineEntry{
		ResourceGroup:  resourceGroupName,
		VMSSName:       vmssName,
		InstanceID:     instanceID,
		VirtualMachine: updatedVM,
		LastUpdate:     time.Now().UTC(),
	}

	localCache := &sync.Map{}
	localCache.Store(nodeName, vmssVMCacheEntry)

	// copy all elements except current VM to localCache
	virtualMachines.Range(func(key, value interface{}) bool {
		if key.(string) != nodeName {
			localCache.Store(key.(string), value.(*VMSSVirtualMachineEntry))
		}
		return true
	})

	ss.vmssVMCache.Update(cacheKey, localCache)
	klog.V(2).Infof("updateCache(%s, %s, %s) for cacheKey(%s) updated successfully", vmssName, resourceGroupName, nodeName, cacheKey)
	return nil
}

func (ss *ScaleSet) newNonVmssUniformNodesCache() (azcache.Resource, error) {
	getter := func(key string) (interface{}, error) {
		vmssFlexVMNodeNames := utilsets.NewString()
		vmssFlexVMProviderIDs := utilsets.NewString()
		avSetVMNodeNames := utilsets.NewString()
		avSetVMProviderIDs := utilsets.NewString()
		resourceGroups, err := ss.GetResourceGroups()
		if err != nil {
			return nil, err
		}
		klog.V(2).Infof("refresh the cache of NonVmssUniformNodesCache in rg %v", resourceGroups)

		for _, resourceGroup := range resourceGroups.UnsortedList() {
			vms, err := ss.Cloud.ListVirtualMachines(resourceGroup)
			if err != nil {
				return nil, fmt.Errorf("getter function of nonVmssUniformNodesCache: failed to list vms in the resource group %s: %w", resourceGroup, err)
			}
			for _, vm := range vms {
				if vm.OsProfile != nil && vm.OsProfile.ComputerName != nil {
					if vm.VirtualMachineScaleSet != nil {
						vmssFlexVMNodeNames.Insert(strings.ToLower(pointer.StringDeref(vm.OsProfile.ComputerName, "")))
						if vm.ID != nil {
							vmssFlexVMProviderIDs.Insert(ss.ProviderName() + "://" + pointer.StringDeref(vm.ID, ""))
						}
					} else {
						avSetVMNodeNames.Insert(strings.ToLower(pointer.StringDeref(vm.OsProfile.ComputerName, "")))
						if vm.ID != nil {
							avSetVMProviderIDs.Insert(ss.ProviderName() + "://" + pointer.StringDeref(vm.ID, ""))
						}
					}
				}
			}
		}

		// store all the node names in the cluster when the cache data was created.
		nodeNames, err := ss.GetNodeNames()
		if err != nil {
			return nil, err
		}

		localCache := NonVmssUniformNodesEntry{
			VMSSFlexVMNodeNames:   vmssFlexVMNodeNames,
			VMSSFlexVMProviderIDs: vmssFlexVMProviderIDs,
			AvSetVMNodeNames:      avSetVMNodeNames,
			AvSetVMProviderIDs:    avSetVMProviderIDs,
			ClusterNodeNames:      nodeNames,
		}

		return localCache, nil
	}

	if ss.Config.NonVmssUniformNodesCacheTTLInSeconds == 0 {
		ss.Config.NonVmssUniformNodesCacheTTLInSeconds = consts.NonVmssUniformNodesCacheTTLDefaultInSeconds
	}
	return azcache.NewTimedCache(time.Duration(ss.Config.NonVmssUniformNodesCacheTTLInSeconds)*time.Second, getter, ss.Cloud.Config.DisableAPICallCache)
}

func (ss *ScaleSet) getVMManagementTypeByNodeName(nodeName string, crt azcache.AzureCacheReadType) (VMManagementType, error) {
	if ss.DisableAvailabilitySetNodes && !ss.EnableVmssFlexNodes {
		return ManagedByVmssUniform, nil
	}
	ss.lockMap.LockEntry(consts.VMManagementTypeLockKey)
	defer ss.lockMap.UnlockEntry(consts.VMManagementTypeLockKey)
	cached, err := ss.nonVmssUniformNodesCache.Get(consts.NonVmssUniformNodesKey, crt)
	if err != nil {
		return ManagedByUnknownVMSet, err
	}

	if ss.Cloud.Config.DisableAPICallCache {
		if cached.(NonVmssUniformNodesEntry).AvSetVMNodeNames.Has(nodeName) {
			return ManagedByAvSet, nil
		}
		if cached.(NonVmssUniformNodesEntry).VMSSFlexVMNodeNames.Has(nodeName) {
			return ManagedByVmssFlex, nil
		}
		return ManagedByVmssUniform, nil
	}

	cachedNodes := cached.(NonVmssUniformNodesEntry).ClusterNodeNames
	// if the node is not in the cache, assume the node has joined after the last cache refresh and attempt to refresh the cache.
	if !cachedNodes.Has(nodeName) {
		if cached.(NonVmssUniformNodesEntry).AvSetVMNodeNames.Has(nodeName) {
			return ManagedByAvSet, nil
		}

		if cached.(NonVmssUniformNodesEntry).VMSSFlexVMNodeNames.Has(nodeName) {
			return ManagedByVmssFlex, nil
		}

		if isNodeInVMSSVMCache(nodeName, ss.vmssVMCache) {
			return ManagedByVmssUniform, nil
		}

		klog.V(2).Infof("Node %s has joined the cluster since the last VM cache refresh in NonVmssUniformNodesEntry, refreshing the cache", nodeName)
		cached, err = ss.nonVmssUniformNodesCache.Get(consts.NonVmssUniformNodesKey, azcache.CacheReadTypeForceRefresh)
		if err != nil {
			return ManagedByUnknownVMSet, err
		}
	}

	cachedAvSetVMs := cached.(NonVmssUniformNodesEntry).AvSetVMNodeNames
	cachedVmssFlexVMs := cached.(NonVmssUniformNodesEntry).VMSSFlexVMNodeNames

	if cachedAvSetVMs.Has(nodeName) {
		return ManagedByAvSet, nil
	}
	if cachedVmssFlexVMs.Has(nodeName) {
		return ManagedByVmssFlex, nil
	}

	return ManagedByVmssUniform, nil
}

func (ss *ScaleSet) getVMManagementTypeByProviderID(providerID string, crt azcache.AzureCacheReadType) (VMManagementType, error) {
	if ss.DisableAvailabilitySetNodes && !ss.EnableVmssFlexNodes {
		return ManagedByVmssUniform, nil
	}
	_, err := extractScaleSetNameByProviderID(providerID)
	if err == nil {
		return ManagedByVmssUniform, nil
	}

	ss.lockMap.LockEntry(consts.VMManagementTypeLockKey)
	defer ss.lockMap.UnlockEntry(consts.VMManagementTypeLockKey)
	cached, err := ss.nonVmssUniformNodesCache.Get(consts.NonVmssUniformNodesKey, crt)
	if err != nil {
		return ManagedByUnknownVMSet, err
	}

	cachedVmssFlexVMProviderIDs := cached.(NonVmssUniformNodesEntry).VMSSFlexVMProviderIDs
	cachedAvSetVMProviderIDs := cached.(NonVmssUniformNodesEntry).AvSetVMProviderIDs

	if cachedAvSetVMProviderIDs.Has(providerID) {
		return ManagedByAvSet, nil
	}
	if cachedVmssFlexVMProviderIDs.Has(providerID) {
		return ManagedByVmssFlex, nil
	}
	return ManagedByUnknownVMSet, fmt.Errorf("getVMManagementTypeByProviderID : failed to check the providerID %s management type", providerID)

}

// getVMManagementTypeByIPConfigurationID determines the VM type by the following steps:
//  1. If the ipConfigurationID is in the format of vmssIPConfigurationRE, returns vmss uniform.
//  2. If the name of the VM can be obtained by trimming the `-nic` suffix from the nic name, and the VM name is in the
//     VMAS cache, returns availability set.
//  3. If the VM name obtained from step 2 is not in the VMAS cache, try to get the VM name from NIC.VirtualMachine.ID.
//  4. If the VM name obtained from step 3 is in the VMAS cache, returns availability set. Or, returns vmss flex.
func (ss *ScaleSet) getVMManagementTypeByIPConfigurationID(ipConfigurationID string, crt azcache.AzureCacheReadType) (VMManagementType, error) {
	if ss.DisableAvailabilitySetNodes && !ss.EnableVmssFlexNodes {
		return ManagedByVmssUniform, nil
	}

	_, _, err := getScaleSetAndResourceGroupNameByIPConfigurationID(ipConfigurationID)
	if err == nil {
		return ManagedByVmssUniform, nil
	}

	ss.lockMap.LockEntry(consts.VMManagementTypeLockKey)
	defer ss.lockMap.UnlockEntry(consts.VMManagementTypeLockKey)
	cached, err := ss.nonVmssUniformNodesCache.Get(consts.NonVmssUniformNodesKey, crt)
	if err != nil {
		return ManagedByUnknownVMSet, err
	}

	nicResourceGroup, nicName, err := getResourceGroupAndNameFromNICID(ipConfigurationID)
	if err != nil {
		return ManagedByUnknownVMSet, fmt.Errorf("can not extract nic name from ipConfigurationID (%s)", ipConfigurationID)
	}

	vmName := strings.Replace(nicName, "-nic", "", 1)

	cachedAvSetVMs := cached.(NonVmssUniformNodesEntry).AvSetVMNodeNames
	if cachedAvSetVMs.Has(vmName) {
		return ManagedByAvSet, nil
	}

	// If the node is not in the cache, assume the node has joined after the last cache refresh and attempt to refresh the cache
	cached, err = ss.nonVmssUniformNodesCache.Get(consts.NonVmssUniformNodesKey, azcache.CacheReadTypeForceRefresh)
	if err != nil {
		return ManagedByUnknownVMSet, err
	}

	cachedAvSetVMs = cached.(NonVmssUniformNodesEntry).AvSetVMNodeNames
	if cachedAvSetVMs.Has(vmName) {
		return ManagedByAvSet, nil
	}

	// Get the vmName by nic.VirtualMachine.ID if the vmName is not in the format
	// of `vmName-nic`. This introduces an extra ARM call.
	vmName, err = ss.GetVMNameByIPConfigurationName(nicResourceGroup, nicName)
	if err != nil {
		return ManagedByUnknownVMSet, fmt.Errorf("failed to get vm name by ip config ID %s: %w", ipConfigurationID, err)
	}
	if cachedAvSetVMs.Has(vmName) {
		return ManagedByAvSet, nil
	}

	return ManagedByVmssFlex, nil
}

func (az *Cloud) GetVMNameByIPConfigurationName(nicResourceGroup, nicName string) (string, error) {
	ctx, cancel := getContextWithCancel()
	defer cancel()
	nic, rerr := az.InterfacesClient.Get(ctx, nicResourceGroup, nicName, "")
	if rerr != nil {
		return "", fmt.Errorf("failed to get interface of name %s: %w", nicName, rerr.Error())
	}
	if nic.InterfacePropertiesFormat == nil || nic.InterfacePropertiesFormat.VirtualMachine == nil || nic.InterfacePropertiesFormat.VirtualMachine.ID == nil {
		return "", fmt.Errorf("failed to get vm ID of nic %s", pointer.StringDeref(nic.Name, ""))
	}
	vmID := pointer.StringDeref(nic.InterfacePropertiesFormat.VirtualMachine.ID, "")
	matches := vmIDRE.FindStringSubmatch(vmID)
	if len(matches) != 2 {
		return "", fmt.Errorf("invalid virtual machine ID %s", vmID)
	}
	vmName := matches[1]
	return vmName, nil
}
