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

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2022-03-01/compute"
	"github.com/Azure/go-autorest/autorest/to"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"

	azcache "sigs.k8s.io/cloud-provider-azure/pkg/cache"
	"sigs.k8s.io/cloud-provider-azure/pkg/consts"
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
	VMSSFlexVMNodeNames   sets.String
	VMSSFlexVMProviderIDs sets.String
	AvSetVMNodeNames      sets.String
	AvSetVMProviderIDs    sets.String
	ClusterNodeNames      sets.String
}

type VMManagementType string

const (
	ManagedByVmssUniform  VMManagementType = "ManagedByVmssUniform"
	ManagedByVmssFlex     VMManagementType = "ManagedByVmssFlex"
	ManagedByAvSet        VMManagementType = "ManagedByAvSet"
	ManagedByUnknownVMSet VMManagementType = "ManagedByUnknownVMSet"
)

func (ss *ScaleSet) newVMSSCache(ctx context.Context) (*azcache.TimedCache, error) {
	getter := func(key string) (interface{}, error) {
		localCache := &sync.Map{} // [vmssName]*vmssEntry

		allResourceGroups, err := ss.GetResourceGroups()
		if err != nil {
			return nil, err
		}

		resourceGroupNotFound := false
		for _, resourceGroup := range allResourceGroups.List() {
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

		if resourceGroupNotFound {
			// gc vmss vm cache when there is resource group not found
			vmssVMKeys := ss.vmssVMCache.Store.ListKeys()
			for _, cacheKey := range vmssVMKeys {
				vmssName := cacheKey[strings.LastIndex(cacheKey, "/")+1:]
				if _, ok := localCache.Load(vmssName); !ok {
					klog.V(2).Infof("remove vmss %s from vmssVMCache due to rg not found", cacheKey)
					_ = ss.vmssVMCache.Delete(cacheKey)
				}
			}
		}
		return localCache, nil
	}

	if ss.Config.VmssCacheTTLInSeconds == 0 {
		ss.Config.VmssCacheTTLInSeconds = consts.VMSSCacheTTLDefaultInSeconds
	}
	return azcache.NewTimedcache(time.Duration(ss.Config.VmssCacheTTLInSeconds)*time.Second, getter)
}

func extractVmssVMName(name string) (string, string, error) {
	split := strings.SplitAfter(name, consts.VMSSNameSeparator)
	if len(split) < 2 {
		klog.V(3).Infof("Failed to extract vmssVMName %q", name)
		return "", "", ErrorNotVmssInstance
	}

	ssName := strings.Join(split[0:len(split)-1], "")
	// removing the trailing `vmssNameSeparator` since we used SplitAfter
	ssName = ssName[:len(ssName)-1]
	instanceID := split[len(split)-1]
	return ssName, instanceID, nil
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

// gcVMSSVMCache delete stale VMSS VMs caches from deleted VMSSes.
func (ss *ScaleSet) gcVMSSVMCache() error {
	return ss.vmssCache.Delete(consts.VMSSKey)
}

// newVMSSVirtualMachinesCache instantiates a new VMs cache for VMs belonging to the provided VMSS.
func (ss *ScaleSet) newVMSSVirtualMachinesCache() (*azcache.TimedCache, error) {
	vmssVirtualMachinesCacheTTL := time.Duration(ss.Config.VmssVirtualMachinesCacheTTLInSeconds) * time.Second

	getter := func(cacheKey string) (interface{}, error) {
		localCache := &sync.Map{} // [nodeName]*VMSSVirtualMachineEntry

		oldCache := make(map[string]*VMSSVirtualMachineEntry)

		entry, exists, err := ss.vmssVMCache.Store.GetByKey(cacheKey)
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

		result := strings.Split(cacheKey, "/")
		if len(result) < 2 {
			err = fmt.Errorf("Invalid cacheKey (%s)", cacheKey)
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
				InstanceID:     to.String(vm.InstanceID),
				VirtualMachine: &vm,
				LastUpdate:     time.Now().UTC(),
			}
			// set cache entry to nil when the VM is under deleting.
			if vm.VirtualMachineScaleSetVMProperties != nil &&
				strings.EqualFold(to.String(vm.VirtualMachineScaleSetVMProperties.ProvisioningState), string(compute.ProvisioningStateDeleting)) {
				klog.V(4).Infof("VMSS virtualMachine %q is under deleting, setting its cache to nil", computerName)
				vmssVMCacheEntry.VirtualMachine = nil
			}
			localCache.Store(computerName, vmssVMCacheEntry)

			delete(oldCache, computerName)
		}

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

		return localCache, nil
	}

	return azcache.NewTimedcache(vmssVirtualMachinesCacheTTL, getter)
}

// DeleteCacheForNode deletes Node from VMSS VM and VM caches.
func (ss *ScaleSet) DeleteCacheForNode(nodeName string) error {
	vmManagementType, err := ss.getVMManagementTypeByNodeName(nodeName, azcache.CacheReadTypeUnsafe)
	if err != nil {
		klog.Errorf("Failed to check VM management type: %v", err)
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
		klog.Errorf("DeleteCacheForNode(%s) failed with error: %v", nodeName, err)
		return err
	}

	err = ss.vmssVMCache.Delete(getVMSSVMCacheKey(node.resourceGroup, node.vmssName))
	if err != nil {
		klog.Errorf("DeleteCacheForNode(%s) failed to remove from vmssVMCache with error: %v", nodeName, err)
		return err
	}

	if err := ss.gcVMSSVMCache(); err != nil {
		klog.Errorf("DeleteCacheForNode(%s) failed to gc stale vmss caches: %v", nodeName, err)
	}

	return nil
}

func (ss *ScaleSet) updateCache(nodeName, resourceGroupName, vmssName, instanceID string, updatedVM *compute.VirtualMachineScaleSetVM) error {
	// lock the VMSS entry to ensure a consistent view of the VM map when there are concurrent updates.
	cacheKey := getVMSSVMCacheKey(resourceGroupName, vmssName)
	ss.lockMap.LockEntry(cacheKey)
	defer ss.lockMap.UnlockEntry(cacheKey)

	virtualMachines, err := ss.getVMSSVMsFromCache(resourceGroupName, vmssName, azcache.CacheReadTypeUnsafe)
	if err != nil {
		err = fmt.Errorf("updateCache(%s, %s, %s) failed getting vmCache with error: %w", vmssName, resourceGroupName, nodeName, err)
		return err
	}

	localCache := &sync.Map{}

	vmssVMCacheEntry := &VMSSVirtualMachineEntry{
		ResourceGroup:  resourceGroupName,
		VMSSName:       vmssName,
		InstanceID:     instanceID,
		VirtualMachine: updatedVM,
		LastUpdate:     time.Now().UTC(),
	}

	localCache.Store(nodeName, vmssVMCacheEntry)

	virtualMachines.Range(func(key, value interface{}) bool {
		if key.(string) != nodeName {
			localCache.Store(key.(string), value.(*VMSSVirtualMachineEntry))
		}
		return true
	})

	ss.vmssVMCache.Update(cacheKey, localCache)
	klog.V(4).Infof("updateCache(%s, %s, %s) for cacheKey(%s) updated successfully", vmssName, resourceGroupName, nodeName, cacheKey)
	return nil
}

func (ss *ScaleSet) newNonVmssUniformNodesCache() (*azcache.TimedCache, error) {
	getter := func(key string) (interface{}, error) {
		vmssFlexVMNodeNames := sets.NewString()
		vmssFlexVMProviderIDs := sets.NewString()
		avSetVMNodeNames := sets.NewString()
		avSetVMProviderIDs := sets.NewString()
		resourceGroups, err := ss.GetResourceGroups()
		if err != nil {
			return nil, err
		}
		klog.V(2).Infof("refresh the cache of NonVmssUniformNodesCache in rg %v", resourceGroups)

		for _, resourceGroup := range resourceGroups.List() {
			vms, err := ss.Cloud.ListVirtualMachines(resourceGroup)
			if err != nil {
				return nil, fmt.Errorf("getter function of nonVmssUniformNodesCache: failed to list vms in the resource group %s: %w", resourceGroup, err)
			}
			for _, vm := range vms {
				if vm.OsProfile != nil && vm.OsProfile.ComputerName != nil {
					if vm.VirtualMachineScaleSet != nil {
						vmssFlexVMNodeNames.Insert(strings.ToLower(to.String(vm.OsProfile.ComputerName)))
						if vm.ID != nil {
							vmssFlexVMProviderIDs.Insert(ss.ProviderName() + "://" + to.String(vm.ID))
						}
					} else {
						avSetVMNodeNames.Insert(strings.ToLower(to.String(vm.OsProfile.ComputerName)))
						if vm.ID != nil {
							avSetVMProviderIDs.Insert(ss.ProviderName() + "://" + to.String(vm.ID))
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
	return azcache.NewTimedcache(time.Duration(ss.Config.NonVmssUniformNodesCacheTTLInSeconds)*time.Second, getter)
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

	matches := nicIDRE.FindStringSubmatch(ipConfigurationID)
	if len(matches) != 3 {
		return ManagedByUnknownVMSet, fmt.Errorf("can not extract nic name from ipConfigurationID (%s)", ipConfigurationID)
	}

	nicResourceGroup, nicName := matches[1], matches[2]
	if nicResourceGroup == "" || nicName == "" {
		return ManagedByUnknownVMSet, fmt.Errorf("invalid ip config ID %s", ipConfigurationID)
	}

	vmName := strings.Replace(nicName, "-nic", "", 1)

	cachedAvSetVMs := cached.(NonVmssUniformNodesEntry).AvSetVMNodeNames

	if cachedAvSetVMs.Has(vmName) {
		return ManagedByAvSet, nil
	}
	return ManagedByVmssFlex, nil
}
