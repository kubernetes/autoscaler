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

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"
	"k8s.io/utils/pointer"

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

type AvailabilitySetNodeEntry struct {
	VMNames   sets.String
	NodeNames sets.String
	VMs       []compute.VirtualMachine
}

type VMManagementType string

const (
	ManagedByVmssUniform  VMManagementType = "ManagedByVmssUniform"
	ManagedByAvSet        VMManagementType = "ManagedByAvSet"
	ManagedByUnknownVMSet VMManagementType = "ManagedByUnknownVMSet"
)

func (ss *ScaleSet) newVMSSCache() (*azcache.TimedCache, error) {
	getter := func(key string) (interface{}, error) {
		localCache := &sync.Map{} // [vmssName]*vmssEntry

		allResourceGroups, err := ss.GetResourceGroups()
		if err != nil {
			return nil, err
		}

		resourceGroupNotFound := false
		for _, resourceGroup := range allResourceGroups.List() {
			allScaleSets, rerr := ss.VirtualMachineScaleSetsClient.List(context.Background(), resourceGroup)
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
				localCache.Store(*scaleSet.Name, &VMSSEntry{
					VMSS:          &scaleSet,
					ResourceGroup: resourceGroup,
					LastUpdate:    time.Now().UTC(),
				})
			}
		}

		if resourceGroupNotFound {
			// gc vmss vm cache when there is resource group not found
			removed := map[string]bool{}
			ss.vmssVMCache.Range(func(key, value interface{}) bool {
				cacheKey := key.(string)
				vlistIdx := cacheKey[strings.LastIndex(cacheKey, "/")+1:]
				if _, ok := localCache.Load(vlistIdx); !ok {
					klog.V(2).Infof("remove vmss %s from cache due to rg not found", cacheKey)
					removed[cacheKey] = true
				}
				return true
			})

			for key := range removed {
				ss.vmssVMCache.Delete(key)
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

// getVMSSVMCache returns an *azcache.TimedCache and cache key for a VMSS (creating that cache if new).
func (ss *ScaleSet) getVMSSVMCache(resourceGroup, vmssName string) (string, *azcache.TimedCache, error) {
	cacheKey := strings.ToLower(fmt.Sprintf("%s/%s", resourceGroup, vmssName))
	if entry, ok := ss.vmssVMCache.Load(cacheKey); ok {
		cache := entry.(*azcache.TimedCache)
		return cacheKey, cache, nil
	}

	cache, err := ss.newVMSSVirtualMachinesCache(resourceGroup, vmssName, cacheKey)
	if err != nil {
		return "", nil, err
	}
	ss.vmssVMCache.Store(cacheKey, cache)
	return cacheKey, cache, nil
}

// gcVMSSVMCache delete stale VMSS VMs caches from deleted VMSSes.
func (ss *ScaleSet) gcVMSSVMCache() error {
	cached, err := ss.vmssCache.Get(consts.VMSSKey, azcache.CacheReadTypeUnsafe)
	if err != nil {
		return err
	}

	vmsses := cached.(*sync.Map)
	removed := map[string]bool{}
	ss.vmssVMCache.Range(func(key, value interface{}) bool {
		cacheKey := key.(string)
		vlistIdx := cacheKey[strings.LastIndex(cacheKey, "/")+1:]
		if _, ok := vmsses.Load(vlistIdx); !ok {
			removed[cacheKey] = true
		}
		return true
	})

	for key := range removed {
		ss.vmssVMCache.Delete(key)
	}

	return nil
}

// newVMSSVirtualMachinesCache instantiates a new VMs cache for VMs belonging to the provided VMSS.
func (ss *ScaleSet) newVMSSVirtualMachinesCache(resourceGroupName, vmssName, cacheKey string) (*azcache.TimedCache, error) {
	vmssVirtualMachinesCacheTTL := time.Duration(ss.Config.VmssVirtualMachinesCacheTTLInSeconds) * time.Second

	getter := func(key string) (interface{}, error) {
		localCache := &sync.Map{} // [nodeName]*vmssVirtualMachinesEntry

		oldCache := make(map[string]VMSSVirtualMachineEntry)

		if vmssCache, ok := ss.vmssVMCache.Load(cacheKey); ok {
			// get old cache before refreshing the cache
			cache := vmssCache.(*azcache.TimedCache)
			entry, exists, err := cache.Store.GetByKey(cacheKey)
			if err != nil {
				return nil, err
			}
			if exists {
				cached := entry.(*azcache.AzureCacheEntry).Data
				if cached != nil {
					virtualMachines := cached.(*sync.Map)
					virtualMachines.Range(func(key, value interface{}) bool {
						oldCache[key.(string)] = *value.(*VMSSVirtualMachineEntry)
						return true
					})
				}
			}
		}

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
				strings.EqualFold(pointer.StringDeref(vm.VirtualMachineScaleSetVMProperties.ProvisioningState, ""), string(compute.ProvisioningStateDeleting)) {
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
	managedByAS, err := ss.isNodeManagedByAvailabilitySet(nodeName, azcache.CacheReadTypeUnsafe)
	if err != nil {
		klog.Warningf("DeleteCacheForNode(%s): failed to check if the node is managed by AvailabilitySet", nodeName)
		return nil
	}
	if managedByAS {
		ss.lockMap.LockEntry(consts.AvailabilitySetNodesKey)
		defer ss.lockMap.UnlockEntry(consts.AvailabilitySetNodesKey)

		_ = ss.availabilitySetNodesCache.Delete(consts.AvailabilitySetNodesKey)
		_ = ss.vmCache.Delete(nodeName)
		return nil
	}

	node, err := ss.getNodeIdentityByNodeName(nodeName, azcache.CacheReadTypeUnsafe)
	if err != nil {
		klog.Errorf("DeleteCacheForNode(%s) failed with error: %v", nodeName, err)
		return err
	}

	cacheKey, timedcache, err := ss.getVMSSVMCache(node.resourceGroup, node.vmssName)
	if err != nil {
		klog.Errorf("DeleteCacheForNode(%s) failed with error: %v", nodeName, err)
		return err
	}

	entry, exists, err := timedcache.Store.GetByKey(cacheKey)
	if err != nil {
		return err
	}
	if exists {
		cached := entry.(*azcache.AzureCacheEntry).Data
		if cached != nil {
			virtualMachines := cached.(*sync.Map)
			virtualMachines.Delete(nodeName)
			entry.(*azcache.AzureCacheEntry).Data = virtualMachines
		}
	}

	if err := ss.gcVMSSVMCache(); err != nil {
		klog.Errorf("DeleteCacheForNode(%s) failed to gc stale vmss caches: %v", nodeName, err)
	}

	return nil
}

func (ss *ScaleSet) newAvailabilitySetNodesCache() (*azcache.TimedCache, error) {
	getter := func(key string) (interface{}, error) {
		vmNames := sets.NewString()
		resourceGroups, err := ss.GetResourceGroups()
		if err != nil {
			return nil, err
		}

		vmList := make([]compute.VirtualMachine, 0)
		for _, resourceGroup := range resourceGroups.List() {
			vms, err := ss.Cloud.ListVirtualMachines(resourceGroup)
			if err != nil {
				return nil, fmt.Errorf("newAvailabilitySetNodesCache: failed to list vms in the resource group %s: %w", resourceGroup, err)
			}
			for _, vm := range vms {
				if vm.Name != nil {
					vmNames.Insert(pointer.StringDeref(vm.Name, ""))
					vmList = append(vmList, vm)
				}
			}
		}

		// store all the node names in the cluster when the cache data was created.
		nodeNames, err := ss.GetNodeNames()
		if err != nil {
			return nil, err
		}

		localCache := &AvailabilitySetNodeEntry{
			VMNames:   vmNames,
			NodeNames: nodeNames,
			VMs:       vmList,
		}

		return localCache, nil
	}

	if ss.Config.AvailabilitySetNodesCacheTTLInSeconds == 0 {
		ss.Config.AvailabilitySetNodesCacheTTLInSeconds = consts.AvailabilitySetNodesCacheTTLDefaultInSeconds
	}
	return azcache.NewTimedcache(time.Duration(ss.Config.AvailabilitySetNodesCacheTTLInSeconds)*time.Second, getter)
}

func (ss *ScaleSet) isNodeManagedByAvailabilitySet(nodeName string, crt azcache.AzureCacheReadType) (bool, error) {
	// Assume all nodes are managed by VMSS when DisableAvailabilitySetNodes is enabled.
	if ss.DisableAvailabilitySetNodes {
		klog.V(6).Infof("Assuming node %q is managed by VMSS since DisableAvailabilitySetNodes is set to true", nodeName)
		return false, nil
	}

	cached, err := ss.availabilitySetNodesCache.Get(consts.AvailabilitySetNodesKey, crt)
	if err != nil {
		return false, err
	}

	cachedNodes := cached.(*AvailabilitySetNodeEntry).NodeNames
	// if the node is not in the cache, assume the node has joined after the last cache refresh and attempt to refresh the cache.
	if !cachedNodes.Has(nodeName) {
		klog.V(2).Infof("Node %s has joined the cluster since the last VM cache refresh, refreshing the cache", nodeName)
		cached, err = ss.availabilitySetNodesCache.Get(consts.AvailabilitySetNodesKey, azcache.CacheReadTypeForceRefresh)
		if err != nil {
			return false, err
		}
	}

	cachedVMs := cached.(*AvailabilitySetNodeEntry).VMNames
	return cachedVMs.Has(nodeName), nil
}

func (ss *ScaleSet) getVMManagementTypeByIPConfigurationID(ipConfigurationID string, crt azcache.AzureCacheReadType) (VMManagementType, error) {
	if ss.DisableAvailabilitySetNodes {
		return ManagedByVmssUniform, nil
	}

	_, _, err := getScaleSetAndResourceGroupNameByIPConfigurationID(ipConfigurationID)
	if err == nil {
		return ManagedByVmssUniform, nil
	}

	ss.lockMap.LockEntry(consts.VMManagementTypeLockKey)
	defer ss.lockMap.UnlockEntry(consts.VMManagementTypeLockKey)
	cached, err := ss.availabilitySetNodesCache.Get(consts.AvailabilitySetNodesKey, crt)
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

	cachedAvSetVMs := cached.(*AvailabilitySetNodeEntry).VMNames

	if cachedAvSetVMs.Has(vmName) {
		return ManagedByAvSet, nil
	}
	klog.Warningf("Cannot determine the management type by IP configuration ID %s", ipConfigurationID)
	return ManagedByUnknownVMSet, nil
}
