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
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v7"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v9"
	"github.com/samber/lo"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	cloudprovider "k8s.io/cloud-provider"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"

	azcache "sigs.k8s.io/cloud-provider-azure/pkg/cache"
	"sigs.k8s.io/cloud-provider-azure/pkg/consts"
	"sigs.k8s.io/cloud-provider-azure/pkg/log"
	"sigs.k8s.io/cloud-provider-azure/pkg/metrics"
	"sigs.k8s.io/cloud-provider-azure/pkg/provider/virtualmachine"
	"sigs.k8s.io/cloud-provider-azure/pkg/util/errutils"
	"sigs.k8s.io/cloud-provider-azure/pkg/util/lockmap"
	vmutil "sigs.k8s.io/cloud-provider-azure/pkg/util/vm"
)

var (
	// ErrorNotVmssInstance indicates an instance is not belonging to any vmss.
	ErrorNotVmssInstance = errors.New("not a vmss instance")
	ErrScaleSetNotFound  = errors.New("scale set not found")

	scaleSetNameRE           = regexp.MustCompile(`.*/subscriptions/(?:.*)/Microsoft.Compute/virtualMachineScaleSets/(.+)/virtualMachines(?:.*)`)
	resourceGroupRE          = regexp.MustCompile(`.*/subscriptions/(?:.*)/resourceGroups/(.+)/providers/Microsoft.Compute/virtualMachineScaleSets/(?:.*)/virtualMachines(?:.*)`)
	vmssIPConfigurationRE    = regexp.MustCompile(`.*/subscriptions/(?:.*)/resourceGroups/(.+)/providers/Microsoft.Compute/virtualMachineScaleSets/(.+)/virtualMachines/(.+)/networkInterfaces(?:.*)`)
	vmssPIPConfigurationRE   = regexp.MustCompile(`.*/subscriptions/(?:.*)/resourceGroups/(.+)/providers/Microsoft.Compute/virtualMachineScaleSets/(.+)/virtualMachines/(.+)/networkInterfaces/(.+)/ipConfigurations/(.+)/publicIPAddresses/(.+)`)
	vmssVMResourceIDTemplate = `/subscriptions/(?:.*)/resourceGroups/(.+)/providers/Microsoft.Compute/virtualMachineScaleSets/(.+)/virtualMachines/(?:\d+)`
	vmssVMResourceIDRE       = regexp.MustCompile(vmssVMResourceIDTemplate)
	vmssVMProviderIDRE       = regexp.MustCompile(fmt.Sprintf("%s%s", "azure://", vmssVMResourceIDTemplate))
)

// vmssMetaInfo contains the metadata for a VMSS.
type vmssMetaInfo struct {
	vmssName      string
	resourceGroup string
}

// nodeIdentity identifies a node within a subscription.
type nodeIdentity struct {
	resourceGroup string
	vmssName      string
	nodeName      string
}

// ScaleSet implements VMSet interface for Azure scale set.
type ScaleSet struct {
	*Cloud

	// availabilitySet is also required for scaleSet because some instances
	// (e.g. control plane nodes) may not belong to any scale sets.
	// this also allows for clusters with both VM and VMSS nodes.
	availabilitySet VMSet

	// flexScaleSet is required for self hosted K8s cluster (for example, capz)
	// It is also used when there are vmssflex node and other types of node in
	// the same cluster.
	flexScaleSet VMSet

	// vmssCache is timed cache where the Store in the cache is a map of
	// Key: consts.VMSSKey
	// Value: sync.Map of [vmssName]*VMSSEntry
	vmssCache azcache.Resource

	// vmssVMCache is timed cache where the Store in the cache is a map of
	// Key: [resourcegroup/vmssName]
	// Value: sync.Map of [vmName]*VMSSVirtualMachineEntry
	vmssVMCache azcache.Resource

	// nonVmssUniformNodesCache is used to store node names from non uniform vm.
	// Currently, the nodes can from avset or vmss flex or individual vm.
	// This cache contains an entry called nonVmssUniformNodesEntry.
	// nonVmssUniformNodesEntry contains avSetVMNodeNames list, clusterNodeNames list
	// and current clusterNodeNames.
	nonVmssUniformNodesCache azcache.Resource

	// lockMap in cache refresh
	lockMap *lockmap.LockMap
}

// RefreshCaches invalidates and renew all related caches.
func (ss *ScaleSet) RefreshCaches() error {
	logger := klog.Background().WithName("ss.RefreshCaches")

	var err error
	ss.vmssCache, err = ss.newVMSSCache()
	if err != nil {
		logger.Error(err, "failed to create or refresh vmss cache")
		return err
	}

	if !ss.DisableAvailabilitySetNodes || ss.EnableVmssFlexNodes {
		ss.nonVmssUniformNodesCache, err = ss.newNonVmssUniformNodesCache()
		if err != nil {
			logger.Error(err, "failed to create or refresh nonVmssUniformNodes cache")
			return err
		}
	}

	ss.vmssVMCache, err = ss.newVMSSVirtualMachinesCache()
	if err != nil {
		logger.Error(err, "failed to create or refresh vmssVM cache")
		return err
	}

	return nil
}

// newScaleSet creates a new ScaleSet.
func newScaleSet(az *Cloud) (VMSet, error) {
	if az.VmssVirtualMachinesCacheTTLInSeconds == 0 {
		az.VmssVirtualMachinesCacheTTLInSeconds = consts.VMSSVirtualMachinesCacheTTLDefaultInSeconds
	}

	var err error
	as, err := newAvailabilitySet(az)
	if err != nil {
		return nil, err
	}
	fs, err := newFlexScaleSet(az)
	if err != nil {
		return nil, err
	}

	ss := &ScaleSet{
		Cloud:           az,
		availabilitySet: as,
		flexScaleSet:    fs,
		lockMap:         lockmap.NewLockMap(),
	}

	if err := ss.RefreshCaches(); err != nil {
		return nil, err
	}

	ss.lockMap = lockmap.NewLockMap()
	return ss, nil
}

func (ss *ScaleSet) getVMSS(ctx context.Context, vmssName string, crt azcache.AzureCacheReadType) (*armcompute.VirtualMachineScaleSet, error) {
	logger := log.FromContextOrBackground(ctx).WithName("getVMSS")
	getter := func(vmssName string) (*armcompute.VirtualMachineScaleSet, error) {
		cached, err := ss.vmssCache.Get(ctx, consts.VMSSKey, crt)
		if err != nil {
			return nil, err
		}
		vmsses := cached.(*sync.Map)
		if vmss, ok := vmsses.Load(vmssName); ok {
			result := vmss.(*VMSSEntry)
			return result.VMSS, nil
		}

		return nil, nil
	}

	vmss, err := getter(vmssName)
	if err != nil {
		return nil, err
	}
	if vmss != nil {
		logger.V(6).Info("Fetched vmss from cache", "vmssName", vmssName, "etag", ptr.Deref(vmss.Etag, ""))
		return vmss, nil
	}

	logger.V(2).Info("Couldn't find VMSS, refreshing the cache", "vmssName", vmssName)
	_ = ss.vmssCache.Delete(consts.VMSSKey)
	vmss, err = getter(vmssName)
	if err != nil {
		return nil, err
	}

	if vmss == nil {
		return nil, cloudprovider.InstanceNotFound
	}
	logger.V(2).Info("Fetched vmss after cache refresh", "vmssName", vmssName, "etag", ptr.Deref(vmss.Etag, ""))
	return vmss, nil
}

// getVmssVMByNodeIdentity find virtualMachineScaleSetVM by nodeIdentity, using node's parent VMSS cache.
// Returns cloudprovider.InstanceNotFound if the node does not belong to the scale set named in nodeIdentity.
func (ss *ScaleSet) getVmssVMByNodeIdentity(ctx context.Context, node *nodeIdentity, crt azcache.AzureCacheReadType) (*virtualmachine.VirtualMachine, error) {
	logger := log.FromContextOrBackground(ctx).WithName("getVmssVMByNodeIdentity")
	// FIXME(ccc): check only if vmss is uniform.
	_, err := getScaleSetVMInstanceID(node.nodeName)
	if err != nil {
		return nil, err
	}

	getter := func(ctx context.Context, crt azcache.AzureCacheReadType) (*virtualmachine.VirtualMachine, bool, error) {
		var found bool
		virtualMachines, err := ss.getVMSSVMsFromCache(ctx, node.resourceGroup, node.vmssName, crt)
		if err != nil {
			return nil, found, err
		}

		if entry, ok := virtualMachines.Load(node.nodeName); ok {
			result := entry.(*VMSSVirtualMachineEntry)
			if result.VirtualMachine == nil {
				klog.Warningf("VM is nil on Node %q, VM is in deleting state", node.nodeName)
				return nil, true, nil
			}
			found = true
			return virtualmachine.FromVirtualMachineScaleSetVM(result.VirtualMachine, virtualmachine.ByVMSS(result.VMSSName)), found, nil
		}

		return nil, found, nil
	}

	vm, found, err := getter(ctx, crt)
	if err != nil {
		return nil, err
	}

	if !found {
		cacheKey := getVMSSVMCacheKey(node.resourceGroup, node.vmssName)
		// lock and try find nodeName from cache again, refresh cache if still not found
		ss.lockMap.LockEntry(cacheKey)
		defer ss.lockMap.UnlockEntry(cacheKey)
		vm, found, err = getter(ctx, crt)
		if err == nil && found && vm != nil {
			logger.V(2).Info("found VMSS VM with nodeName after retry", "nodeName", node.nodeName)
			return vm, nil
		}

		logger.V(2).Info("Couldn't find VMSS VM with nodeName, refreshing the cache", "nodeName", node.nodeName, "vmss", node.vmssName, "resourceGroup", node.resourceGroup)
		vm, found, err = getter(ctx, azcache.CacheReadTypeForceRefresh)
		if err != nil {
			return nil, err
		}
	}

	if found && vm != nil {
		return vm, nil
	}

	if !found || vm == nil {
		klog.Warningf("Unable to find node %s: %v", node.nodeName, cloudprovider.InstanceNotFound)
		return nil, cloudprovider.InstanceNotFound
	}
	return vm, nil
}

// getVmssVM gets virtualMachineScaleSetVM by nodeName from cache.
// Returns cloudprovider.InstanceNotFound if nodeName does not belong to any scale set.
func (ss *ScaleSet) getVmssVM(ctx context.Context, nodeName string, crt azcache.AzureCacheReadType) (*virtualmachine.VirtualMachine, error) {
	node, err := ss.getNodeIdentityByNodeName(ctx, nodeName, crt)
	if err != nil {
		return nil, err
	}

	return ss.getVmssVMByNodeIdentity(ctx, node, crt)
}

// GetPowerStatusByNodeName returns the power state of the specified node.
func (ss *ScaleSet) GetPowerStatusByNodeName(ctx context.Context, name string) (powerState string, err error) {
	logger := log.FromContextOrBackground(ctx).WithName("GetPowerStatusByNodeName")
	vmManagementType, err := ss.getVMManagementTypeByNodeName(ctx, name, azcache.CacheReadTypeUnsafe)
	if err != nil {
		logger.Error(err, "Failed to check VM management type")
		return "", err
	}

	if vmManagementType == ManagedByAvSet {
		// vm is managed by availability set.
		return ss.availabilitySet.GetPowerStatusByNodeName(ctx, name)
	}
	if vmManagementType == ManagedByVmssFlex {
		// vm is managed by vmss flex.
		return ss.flexScaleSet.GetPowerStatusByNodeName(ctx, name)
	}
	// VM is managed by vmss
	vm, err := ss.getVmssVM(ctx, name, azcache.CacheReadTypeDefault)
	if err != nil {
		return powerState, err
	}

	if vm.IsVirtualMachineScaleSetVM() {
		v := vm.AsVirtualMachineScaleSetVM()
		if v.Properties.InstanceView != nil {
			return vmutil.GetVMPowerState(ptr.Deref(v.Name, ""), v.Properties.InstanceView.Statuses), nil
		}
	}

	// vm.Properties.InstanceView or vm.Properties.InstanceView.Statuses are nil when the VM is under deleting.
	logger.V(3).Info("InstanceView for node is nil, assuming it's deleting", "node", name)
	return consts.VMPowerStateUnknown, nil
}

// GetProvisioningStateByNodeName returns the provisioningState for the specified node.
func (ss *ScaleSet) GetProvisioningStateByNodeName(ctx context.Context, name string) (provisioningState string, err error) {
	logger := log.FromContextOrBackground(ctx).WithName("GetProvisioningStateByNodeName")
	vmManagementType, err := ss.getVMManagementTypeByNodeName(ctx, name, azcache.CacheReadTypeUnsafe)
	if err != nil {
		logger.Error(err, "Failed to check VM management type")
		return "", err
	}

	if vmManagementType == ManagedByAvSet {
		// vm is managed by availability set.
		return ss.availabilitySet.GetProvisioningStateByNodeName(ctx, name)
	}
	if vmManagementType == ManagedByVmssFlex {
		// vm is managed by vmss flex.
		return ss.flexScaleSet.GetProvisioningStateByNodeName(ctx, name)
	}

	vm, err := ss.getVmssVM(ctx, name, azcache.CacheReadTypeDefault)
	if err != nil {
		return provisioningState, err
	}

	if vm.VirtualMachineScaleSetVMProperties == nil || vm.VirtualMachineScaleSetVMProperties.ProvisioningState == nil {
		return provisioningState, nil
	}

	return ptr.Deref(vm.VirtualMachineScaleSetVMProperties.ProvisioningState, ""), nil
}

// getCachedVirtualMachineByInstanceID gets scaleSetVMInfo from cache.
// The node must belong to one of scale sets.
func (ss *ScaleSet) getVmssVMByInstanceID(ctx context.Context, resourceGroup, scaleSetName, instanceID string, crt azcache.AzureCacheReadType) (*armcompute.VirtualMachineScaleSetVM, error) {
	logger := log.FromContextOrBackground(ctx).WithName("getVmssVMByInstanceID")
	getter := func(ctx context.Context, crt azcache.AzureCacheReadType) (vm *armcompute.VirtualMachineScaleSetVM, found bool, err error) {
		virtualMachines, err := ss.getVMSSVMsFromCache(ctx, resourceGroup, scaleSetName, crt)
		if err != nil {
			return nil, false, err
		}

		virtualMachines.Range(func(_, value interface{}) bool {
			vmEntry := value.(*VMSSVirtualMachineEntry)
			if strings.EqualFold(vmEntry.ResourceGroup, resourceGroup) &&
				strings.EqualFold(vmEntry.VMSSName, scaleSetName) &&
				strings.EqualFold(vmEntry.InstanceID, instanceID) {
				vm = vmEntry.VirtualMachine
				found = true
				return false
			}

			return true
		})

		return vm, found, nil
	}

	vm, found, err := getter(ctx, crt)
	if err != nil {
		return nil, err
	}
	if !found {
		logger.V(2).Info("Couldn't find VMSS VM with scaleSetName and instanceID, refreshing the cache", "scaleSetName", scaleSetName, "instanceID", instanceID)
		vm, found, err = getter(ctx, azcache.CacheReadTypeForceRefresh)
		if err != nil {
			return nil, err
		}
	}
	if found && vm != nil {
		return vm, nil
	}
	if found && vm == nil {
		logger.V(2).Info("Couldn't find VMSS VM with scaleSetName and instanceID, refreshing the cache if it is expired", "scaleSetName", scaleSetName, "instanceID", instanceID)
		vm, found, err = getter(ctx, azcache.CacheReadTypeDefault)
		if err != nil {
			return nil, err
		}
	}
	if !found || vm == nil {
		// There is a corner case that the VM is deleted but the ip configuration is not deleted from the load balancer.
		// In this case the cloud provider will keep refreshing the cache to search the VM, and will introduce
		// a lot of unnecessary requests to ARM.
		return nil, cloudprovider.InstanceNotFound
	}

	return vm, nil
}

// GetInstanceIDByNodeName gets the cloud provider ID by node name.
// It must return ("", cloudprovider.InstanceNotFound) if the instance does
// not exist or is no longer running.
func (ss *ScaleSet) GetInstanceIDByNodeName(ctx context.Context, name string) (string, error) {
	logger := log.FromContextOrBackground(ctx).WithName("GetInstanceIDByNodeName")
	vmManagementType, err := ss.getVMManagementTypeByNodeName(ctx, name, azcache.CacheReadTypeUnsafe)
	if err != nil {
		logger.Error(err, "Failed to check VM management type")
		return "", err
	}

	if vmManagementType == ManagedByAvSet {
		// vm is managed by availability set.
		return ss.availabilitySet.GetInstanceIDByNodeName(ctx, name)
	}
	if vmManagementType == ManagedByVmssFlex {
		// vm is managed by vmss flex.
		return ss.flexScaleSet.GetInstanceIDByNodeName(ctx, name)
	}

	vm, err := ss.getVmssVM(ctx, name, azcache.CacheReadTypeUnsafe)
	if err != nil {
		// special case: during scaling in, if the vm is deleted and nonVmssUniformNodesCache is refreshed,
		// then getVMManagementTypeByNodeName will return ManagedByVmssUniform no matter what the actual managementType is.
		// In this case, if it is actually a non vmss uniform node, return InstanceNotFound
		if errors.Is(err, ErrorNotVmssInstance) {
			return "", cloudprovider.InstanceNotFound
		}
		logger.Error(err, "Unable to find node", "node", name)
		return "", err
	}

	resourceID := vm.ID
	convertedResourceID, err := ConvertResourceGroupNameToLower(resourceID)
	if err != nil {
		logger.Error(err, "ConvertResourceGroupNameToLower failed")
		return "", err
	}
	return convertedResourceID, nil
}

// GetNodeNameByProviderID gets the node name by provider ID.
// providerID example:
// 1. vmas providerID: azure:///subscriptions/subsid/resourceGroups/rg/providers/Microsoft.Compute/virtualMachines/aks-nodepool1-27053986-0
// 2. vmss providerID:
// azure:///subscriptions/subsid/resourceGroups/rg/providers/Microsoft.Compute/virtualMachineScaleSets/aks-agentpool-22126781-vmss/virtualMachines/1
// /subscriptions/subsid/resourceGroups/rg/providers/Microsoft.Compute/virtualMachineScaleSets/aks-agentpool-22126781-vmss/virtualMachines/k8s-agentpool-36841236-vmss_1
func (ss *ScaleSet) GetNodeNameByProviderID(ctx context.Context, providerID string) (types.NodeName, error) {
	logger := log.FromContextOrBackground(ctx).WithName("GetNodeNameByProviderID")
	vmManagementType, err := ss.getVMManagementTypeByProviderID(ctx, providerID, azcache.CacheReadTypeUnsafe)
	if err != nil {
		logger.Error(err, "Failed to check VM management type")
		return "", err
	}

	if vmManagementType == ManagedByAvSet {
		// vm is managed by availability set.
		return ss.availabilitySet.GetNodeNameByProviderID(ctx, providerID)
	}
	if vmManagementType == ManagedByVmssFlex {
		// vm is managed by vmss flex.
		return ss.flexScaleSet.GetNodeNameByProviderID(ctx, providerID)
	}

	// NodeName is not part of providerID for vmss instances.
	scaleSetName, err := extractScaleSetNameByProviderID(providerID)
	if err != nil {
		return "", fmt.Errorf("error of extracting vmss name for node %q", providerID)
	}
	resourceGroup, err := extractResourceGroupByProviderID(providerID)
	if err != nil {
		return "", fmt.Errorf("error of extracting resource group for node %q", providerID)
	}

	instanceID, err := getLastSegment(providerID, "/")
	if err != nil {
		logger.V(4).Error(err, "Can not extract instanceID from providerID, assuming it is managed by availability set", "providerID", providerID)
		return ss.availabilitySet.GetNodeNameByProviderID(ctx, providerID)
	}

	// instanceID contains scaleSetName (returned by disk.ManagedBy), e.g. k8s-agentpool-36841236-vmss_1
	if strings.HasPrefix(strings.ToLower(instanceID), strings.ToLower(scaleSetName)) {
		instanceID, err = getLastSegment(instanceID, "_")
		if err != nil {
			return "", err
		}
	}

	vm, err := ss.getVmssVMByInstanceID(ctx, resourceGroup, scaleSetName, instanceID, azcache.CacheReadTypeUnsafe)
	if err != nil {
		logger.Error(err, "Unable to find node by providerID", "providerID", providerID)
		return "", err
	}

	if vm.Properties.OSProfile != nil && vm.Properties.OSProfile.ComputerName != nil {
		nodeName := strings.ToLower(*vm.Properties.OSProfile.ComputerName)
		return types.NodeName(nodeName), nil
	}

	return "", nil
}

// GetInstanceTypeByNodeName gets the instance type by node name.
func (ss *ScaleSet) GetInstanceTypeByNodeName(ctx context.Context, name string) (string, error) {
	logger := log.FromContextOrBackground(ctx).WithName("GetInstanceTypeByNodeName")
	vmManagementType, err := ss.getVMManagementTypeByNodeName(ctx, name, azcache.CacheReadTypeUnsafe)
	if err != nil {
		logger.Error(err, "Failed to check VM management type")
		return "", err
	}

	if vmManagementType == ManagedByAvSet {
		// vm is managed by availability set.
		return ss.availabilitySet.GetInstanceTypeByNodeName(ctx, name)
	}
	if vmManagementType == ManagedByVmssFlex {
		// vm is managed by vmss flex.
		return ss.flexScaleSet.GetInstanceTypeByNodeName(ctx, name)
	}

	vm, err := ss.getVmssVM(ctx, name, azcache.CacheReadTypeUnsafe)
	if err != nil {
		return "", err
	}

	if vm.IsVirtualMachineScaleSetVM() {
		v := vm.AsVirtualMachineScaleSetVM()
		if v.SKU != nil && v.SKU.Name != nil {
			return *v.SKU.Name, nil
		}
	}

	return "", nil
}

// GetZoneByNodeName gets availability zone for the specified node. If the node is not running
// with availability zone, then it returns fault domain.
func (ss *ScaleSet) GetZoneByNodeName(ctx context.Context, name string) (cloudprovider.Zone, error) {
	logger := log.FromContextOrBackground(ctx).WithName("GetZoneByNodeName")
	vmManagementType, err := ss.getVMManagementTypeByNodeName(ctx, name, azcache.CacheReadTypeUnsafe)
	if err != nil {
		logger.Error(err, "Failed to check VM management type")
		return cloudprovider.Zone{}, err
	}

	if vmManagementType == ManagedByAvSet {
		// vm is managed by availability set.
		return ss.availabilitySet.GetZoneByNodeName(ctx, name)
	}
	if vmManagementType == ManagedByVmssFlex {
		// vm is managed by vmss flex.
		return ss.flexScaleSet.GetZoneByNodeName(ctx, name)
	}

	vm, err := ss.getVmssVM(ctx, name, azcache.CacheReadTypeUnsafe)
	if err != nil {
		return cloudprovider.Zone{}, err
	}

	var failureDomain string
	if len(vm.Zones) > 0 {
		// Get availability zone for the node.
		zones := vm.Zones
		zoneID, err := strconv.Atoi(*zones[0])
		if err != nil {
			return cloudprovider.Zone{}, fmt.Errorf("failed to parse zone %q: %w", lo.FromSlicePtr(zones), err)
		}

		failureDomain = ss.makeZone(vm.Location, zoneID)
	} else if vm.IsVirtualMachineScaleSetVM() &&
		vm.AsVirtualMachineScaleSetVM().Properties.InstanceView != nil &&
		vm.AsVirtualMachineScaleSetVM().Properties.InstanceView.PlatformFaultDomain != nil {
		// Availability zone is not used for the node, falling back to fault domain.
		failureDomain = strconv.Itoa(int(*vm.AsVirtualMachineScaleSetVM().Properties.InstanceView.PlatformFaultDomain))
	} else {
		err = fmt.Errorf("failed to get zone info")
		logger.Error(err, "got unexpected error")
		_ = ss.DeleteCacheForNode(ctx, name)
		return cloudprovider.Zone{}, err
	}

	return cloudprovider.Zone{
		FailureDomain: strings.ToLower(failureDomain),
		Region:        strings.ToLower(vm.Location),
	}, nil
}

// GetPrimaryVMSetName returns the VM set name depending on the configured vmType.
// It returns config.PrimaryScaleSetName for vmss and config.PrimaryAvailabilitySetName for standard vmType.
func (ss *ScaleSet) GetPrimaryVMSetName() string {
	return ss.PrimaryScaleSetName
}

// GetIPByNodeName gets machine private IP and public IP by node name.
func (ss *ScaleSet) GetIPByNodeName(ctx context.Context, nodeName string) (string, string, error) {
	logger := log.FromContextOrBackground(ctx).WithName("GetIPByNodeName")
	vmManagementType, err := ss.getVMManagementTypeByNodeName(ctx, nodeName, azcache.CacheReadTypeUnsafe)
	if err != nil {
		logger.Error(err, "Failed to check VM management type")
		return "", "", err
	}

	if vmManagementType == ManagedByAvSet {
		// vm is managed by availability set.
		return ss.availabilitySet.GetIPByNodeName(ctx, nodeName)
	}
	if vmManagementType == ManagedByVmssFlex {
		// vm is managed by vmss flex.
		return ss.flexScaleSet.GetIPByNodeName(ctx, nodeName)
	}

	nic, err := ss.GetPrimaryInterface(ctx, nodeName)
	if err != nil {
		logger.Error(err, "GetPrimaryInterface() failed", "nodeName", nodeName)
		return "", "", err
	}

	ipConfig, err := getPrimaryIPConfig(nic)
	if err != nil {
		logger.Error(err, "getPrimaryIPConfig() failed", "nodeName", nodeName, "nic", nic)
		return "", "", err
	}

	internalIP := *ipConfig.Properties.PrivateIPAddress
	publicIP := ""
	if ipConfig.Properties.PublicIPAddress != nil && ipConfig.Properties.PublicIPAddress.ID != nil {
		pipID := *ipConfig.Properties.PublicIPAddress.ID
		matches := vmssPIPConfigurationRE.FindStringSubmatch(pipID)
		if len(matches) == 7 {
			resourceGroupName := matches[1]
			virtualMachineScaleSetName := matches[2]
			virtualMachineIndex := matches[3]
			networkInterfaceName := matches[4]
			IPConfigurationName := matches[5]
			publicIPAddressName := matches[6]
			pip, existsPip, err := ss.getVMSSPublicIPAddress(resourceGroupName, virtualMachineScaleSetName, virtualMachineIndex, networkInterfaceName, IPConfigurationName, publicIPAddressName)
			if err != nil {
				logger.Error(err, "ss.getVMSSPublicIPAddress() failed")
				return "", "", err
			}
			if existsPip && pip.Properties.IPAddress != nil {
				publicIP = *pip.Properties.IPAddress
			}
		} else {
			klog.Warningf("Failed to get VMSS Public IP with ID %s", pipID)
		}
	}

	return internalIP, publicIP, nil
}

func (ss *ScaleSet) getVMSSPublicIPAddress(resourceGroupName string, virtualMachineScaleSetName string, virtualMachineIndex string, networkInterfaceName string, IPConfigurationName string, publicIPAddressName string) (*armnetwork.PublicIPAddress, bool, error) {
	ctx, cancel := getContextWithCancel()
	defer cancel()
	logger := log.FromContextOrBackground(ctx).WithName("getVMSSPublicIPAddress")

	pip, err := ss.NetworkClientFactory.GetPublicIPAddressClient().GetVirtualMachineScaleSetPublicIPAddress(ctx, resourceGroupName, virtualMachineScaleSetName, virtualMachineIndex, networkInterfaceName, IPConfigurationName, publicIPAddressName, nil)
	exists, rerr := checkResourceExistsFromError(err)
	if rerr != nil {
		return nil, false, err
	}

	if !exists {
		logger.V(2).Info("Public IP not found", "publicIPAddressName", publicIPAddressName)
		return nil, false, nil
	}

	return &pip.PublicIPAddress, exists, nil
}

// returns a list of private ips assigned to node
// TODO (khenidak): This should read all nics, not just the primary
// allowing users to split ipv4/v6 on multiple nics
func (ss *ScaleSet) GetPrivateIPsByNodeName(ctx context.Context, nodeName string) ([]string, error) {
	logger := log.FromContextOrBackground(ctx).WithName("GetPrivateIPsByNodeName")
	ips := make([]string, 0)
	vmManagementType, err := ss.getVMManagementTypeByNodeName(ctx, nodeName, azcache.CacheReadTypeUnsafe)
	if err != nil {
		logger.Error(err, "Failed to check VM management type by node name")
		return ips, err
	}

	if vmManagementType == ManagedByAvSet {
		// vm is managed by availability set.
		return ss.availabilitySet.GetPrivateIPsByNodeName(ctx, nodeName)
	}
	if vmManagementType == ManagedByVmssFlex {
		// vm is managed by vmss flex.
		return ss.flexScaleSet.GetPrivateIPsByNodeName(ctx, nodeName)
	}

	nic, err := ss.GetPrimaryInterface(ctx, nodeName)
	if err != nil {
		logger.Error(err, "GetPrimaryInterface() failed", "nodeName", nodeName)
		return ips, err
	}

	if nic.Properties.IPConfigurations == nil {
		return ips, fmt.Errorf("nic.Properties.IPConfigurations for nic (nicname=%q) is nil", *nic.Name)
	}

	for _, ipConfig := range nic.Properties.IPConfigurations {
		if ipConfig.Properties.PrivateIPAddress != nil {
			ips = append(ips, *(ipConfig.Properties.PrivateIPAddress))
		}
	}

	return ips, nil
}

// This returns the full identifier of the primary NIC for the given VM.
func (ss *ScaleSet) getPrimaryInterfaceID(vm *virtualmachine.VirtualMachine) (string, error) {
	machine := vm.AsVirtualMachineScaleSetVM()
	if machine.Properties.NetworkProfile == nil || machine.Properties.NetworkProfile.NetworkInterfaces == nil {
		return "", fmt.Errorf("failed to find the network interfaces for vm %s", ptr.Deref(machine.Name, ""))
	}

	if len(machine.Properties.NetworkProfile.NetworkInterfaces) == 1 {
		return *machine.Properties.NetworkProfile.NetworkInterfaces[0].ID, nil
	}

	for _, ref := range machine.Properties.NetworkProfile.NetworkInterfaces {
		if ptr.Deref(ref.Properties.Primary, false) {
			return *ref.ID, nil
		}
	}

	return "", fmt.Errorf("failed to find a primary nic for the vm. vmname=%q", ptr.Deref(machine.Name, ""))
}

// machineName is composed of computerNamePrefix and 36-based instanceID.
// And instanceID part if in fixed length of 6 characters.
// Refer https://msftstack.wordpress.com/2017/05/10/figuring-out-azure-vm-scale-set-machine-names/.
func getScaleSetVMInstanceID(machineName string) (string, error) {
	nameLength := len(machineName)
	if nameLength < 6 {
		return "", ErrorNotVmssInstance
	}

	instanceID, err := strconv.ParseUint(machineName[nameLength-6:], 36, 64)
	if err != nil {
		return "", ErrorNotVmssInstance
	}

	return fmt.Sprintf("%d", instanceID), nil
}

// extractScaleSetNameByProviderID extracts the scaleset name by vmss node's ProviderID.
func extractScaleSetNameByProviderID(providerID string) (string, error) {
	matches := scaleSetNameRE.FindStringSubmatch(providerID)
	if len(matches) != 2 {
		return "", ErrorNotVmssInstance
	}

	return matches[1], nil
}

// extractResourceGroupByProviderID extracts the resource group name by vmss node's ProviderID.
func extractResourceGroupByProviderID(providerID string) (string, error) {
	matches := resourceGroupRE.FindStringSubmatch(providerID)
	if len(matches) != 2 {
		return "", ErrorNotVmssInstance
	}

	return matches[1], nil
}

// getNodeIdentityByNodeName use the VMSS cache to find a node's resourcegroup and vmss, returned in a nodeIdentity.
func (ss *ScaleSet) getNodeIdentityByNodeName(ctx context.Context, nodeName string, crt azcache.AzureCacheReadType) (*nodeIdentity, error) {
	logger := log.FromContextOrBackground(ctx).WithName("getNodeIdentityByNodeName")
	getter := func(nodeName string, crt azcache.AzureCacheReadType) (*nodeIdentity, error) {
		node := &nodeIdentity{
			nodeName: nodeName,
		}

		cached, err := ss.vmssCache.Get(ctx, consts.VMSSKey, crt)
		if err != nil {
			return nil, err
		}

		vmsses := cached.(*sync.Map)
		vmsses.Range(func(_, value interface{}) bool {
			v := value.(*VMSSEntry)
			if v.VMSS.Name == nil {
				return true
			}

			vmssPrefix := *v.VMSS.Name
			if v.VMSS.Properties.VirtualMachineProfile != nil &&
				v.VMSS.Properties.VirtualMachineProfile.OSProfile != nil &&
				v.VMSS.Properties.VirtualMachineProfile.OSProfile.ComputerNamePrefix != nil {
				vmssPrefix = *v.VMSS.Properties.VirtualMachineProfile.OSProfile.ComputerNamePrefix
			}

			if strings.EqualFold(vmssPrefix, nodeName[:len(nodeName)-6]) {
				node.vmssName = *v.VMSS.Name
				node.resourceGroup = v.ResourceGroup
				return false
			}

			return true
		})
		return node, nil
	}

	// FIXME(ccc): check only if vmss is uniform.
	if _, err := getScaleSetVMInstanceID(nodeName); err != nil {
		return nil, err
	}

	node, err := getter(nodeName, crt)
	if err != nil {
		return nil, err
	}
	if node.vmssName != "" {
		return node, nil
	}

	logger.V(2).Info("Couldn't find VMSS for node, refreshing the cache", "node", nodeName)
	node, err = getter(nodeName, azcache.CacheReadTypeForceRefresh)
	if err != nil {
		return nil, err
	}
	if node.vmssName == "" {
		klog.Warningf("Unable to find node %s: %v", nodeName, cloudprovider.InstanceNotFound)
		return nil, cloudprovider.InstanceNotFound
	}
	return node, nil
}

// listScaleSetVMs lists VMs belonging to the specified scale set.
func (ss *ScaleSet) listScaleSetVMs(scaleSetName, resourceGroup string) ([]*armcompute.VirtualMachineScaleSetVM, error) {
	ctx, cancel := getContextWithCancel()
	defer cancel()

	logger := log.FromContextOrBackground(ctx).WithName("listScaleSetVMs")

	var allVMs []*armcompute.VirtualMachineScaleSetVM
	var rerr error
	if ss.ListVmssVirtualMachinesWithoutInstanceView {
		logger.V(6).Info("listScaleSetVMs called for scaleSetName", "scaleSetName", scaleSetName, "resourceGroup", resourceGroup)
		allVMs, rerr = ss.ComputeClientFactory.GetVirtualMachineScaleSetVMClient().List(ctx, resourceGroup, scaleSetName)
	} else {
		logger.V(6).Info("listScaleSetVMs called for scaleSetName with instanceView", "scaleSetName", scaleSetName, "resourceGroup", resourceGroup)
		allVMs, rerr = ss.ComputeClientFactory.GetVirtualMachineScaleSetVMClient().ListVMInstanceView(ctx, resourceGroup, scaleSetName)
	}
	if rerr != nil {
		logger.Error(rerr, "ComputeClientFactory.GetVirtualMachineScaleSetVMClient().List() failed", "resourceGroup", resourceGroup, "scaleSetName", scaleSetName)
		if exists, err := errutils.CheckResourceExistsFromAzcoreError(rerr); !exists && err == nil {
			return nil, cloudprovider.InstanceNotFound
		}
		return nil, rerr
	}

	return allVMs, nil
}

// getAgentPoolScaleSets lists the virtual machines for the resource group and then builds
// a list of scale sets that match the nodes available to k8s.
func (ss *ScaleSet) getAgentPoolScaleSets(ctx context.Context, nodes []*v1.Node) ([]string, error) {
	logger := log.FromContextOrBackground(ctx).WithName("getAgentPoolScaleSets")
	agentPoolScaleSets := []string{}
	for nx := range nodes {
		if isControlPlaneNode(nodes[nx]) {
			continue
		}

		nodeName := nodes[nx].Name
		shouldExcludeLoadBalancer, err := ss.ShouldNodeExcludedFromLoadBalancer(nodeName)
		if err != nil {
			logger.Error(err, "ShouldNodeExcludedFromLoadBalancer() failed", "nodeName", nodeName)
			return nil, err
		}
		if shouldExcludeLoadBalancer {
			continue
		}

		vm, err := ss.getVmssVM(ctx, nodeName, azcache.CacheReadTypeDefault)
		if err != nil {
			return nil, err
		}

		if vm.VMSSName == "" {
			logger.V(3).Info("Node is not belonging to any known scale sets", "node", nodeName)
			continue
		}

		agentPoolScaleSets = append(agentPoolScaleSets, vm.VMSSName)
	}

	return agentPoolScaleSets, nil
}

// GetVMSetNames selects all possible scale sets for service load balancer. If the service has
// no loadbalancer mode annotation returns the primary VMSet. If service annotation
// for loadbalancer exists then return the eligible VMSet.
func (ss *ScaleSet) GetVMSetNames(ctx context.Context, service *v1.Service, nodes []*v1.Node) ([]*string, error) {
	logger := log.FromContextOrBackground(ctx).WithName("ss.GetVMSetNames")
	hasMode, isAuto, serviceVMSetName := ss.getServiceLoadBalancerMode(service)
	if !hasMode || ss.UseStandardLoadBalancer() {
		// no mode specified in service annotation or use single SLB mode
		// default to PrimaryScaleSetName
		return to.SliceOfPtrs(ss.PrimaryScaleSetName), nil
	}

	scaleSetNames, err := ss.GetAgentPoolVMSetNames(ctx, nodes)
	if err != nil {
		logger.Error(err, "GetAgentPoolVMSetNames() failed")
		return nil, err
	}
	if len(scaleSetNames) == 0 {
		logger.Error(nil, "No scale sets found for nodes in the cluster", "nodeCount", len(nodes))
		return nil, fmt.Errorf("no scale sets found for nodes, node count(%d)", len(nodes))
	}

	if !isAuto {
		found := false
		for asx := range scaleSetNames {
			if strings.EqualFold(*(scaleSetNames)[asx], serviceVMSetName) {
				found = true
				serviceVMSetName = *(scaleSetNames)[asx]
				break
			}
		}
		if !found {
			logger.Error(nil, "scale set in service annotation not found", "scaleSetName", serviceVMSetName)
			return nil, ErrScaleSetNotFound
		}
		return to.SliceOfPtrs(serviceVMSetName), nil
	}

	return scaleSetNames, nil
}

// extractResourceGroupByVMSSNicID extracts the resource group name by vmss nicID.
func extractResourceGroupByVMSSNicID(nicID string) (string, error) {
	matches := vmssIPConfigurationRE.FindStringSubmatch(nicID)
	if len(matches) != 4 {
		return "", fmt.Errorf("error of extracting resourceGroup from nicID %q", nicID)
	}

	return matches[1], nil
}

// GetPrimaryInterface gets machine primary network interface by node name and vmSet.
func (ss *ScaleSet) GetPrimaryInterface(ctx context.Context, nodeName string) (*armnetwork.Interface, error) {
	logger := log.FromContextOrBackground(ctx).WithName("GetPrimaryInterface")
	vmManagementType, err := ss.getVMManagementTypeByNodeName(ctx, nodeName, azcache.CacheReadTypeUnsafe)
	if err != nil {
		logger.Error(err, "Failed to check VM management type by node name")
		return nil, err
	}

	if vmManagementType == ManagedByAvSet {
		// vm is managed by availability set.
		return ss.availabilitySet.GetPrimaryInterface(ctx, nodeName)
	}
	if vmManagementType == ManagedByVmssFlex {
		// vm is managed by vmss flex.
		return ss.flexScaleSet.GetPrimaryInterface(ctx, nodeName)
	}

	vm, err := ss.getVmssVM(ctx, nodeName, azcache.CacheReadTypeDefault)
	if err != nil {
		// VM is availability set, but not cached yet in availabilitySetNodesCache.
		if errors.Is(err, ErrorNotVmssInstance) {
			return ss.availabilitySet.GetPrimaryInterface(ctx, nodeName)
		}

		logger.Error(err, "ss.GetVmssVM() failed", "nodeName", nodeName)
		return nil, err
	}

	primaryInterfaceID, err := ss.getPrimaryInterfaceID(vm)
	if err != nil {
		logger.Error(err, "ss.getPrimaryInterfaceID() failed", "nodeName", nodeName)
		return nil, err
	}

	nicName, err := getLastSegment(primaryInterfaceID, "/")
	if err != nil {
		logger.Error(err, "getLastSegment() failed", "nodeName", nodeName, "primaryInterfaceID", primaryInterfaceID)
		return nil, err
	}
	resourceGroup, err := extractResourceGroupByVMSSNicID(primaryInterfaceID)
	if err != nil {
		return nil, err
	}

	ctx, cancel := getContextWithCancel()
	defer cancel()
	nic, rerr := ss.ComputeClientFactory.GetInterfaceClient().GetVirtualMachineScaleSetNetworkInterface(ctx, resourceGroup, vm.VMSSName,
		vm.InstanceID,
		nicName)
	if rerr != nil {
		exists, realErr := checkResourceExistsFromError(rerr)
		if realErr != nil {
			logger.Error(realErr, "ss.GetVirtualMachineScaleSetNetworkInterface failed", "nodeName", nodeName, "resourceGroup", resourceGroup, "vmssName", vm.VMSSName, "nicName", nicName)
			return nil, realErr
		}

		if !exists {
			return nil, cloudprovider.InstanceNotFound
		}
	}

	// Fix interface's location, which is required when updating the interface.
	// TODO: is this a bug of azure SDK?
	if nic.Location == nil || *nic.Location == "" {
		nic.Location = &vm.Location
	}

	return nic, nil
}

// getPrimaryNetworkInterfaceConfiguration gets primary network interface configuration for VMSS VM or VMSS.
func getPrimaryNetworkInterfaceConfiguration(networkConfigurations []*armcompute.VirtualMachineScaleSetNetworkConfiguration, resource string) (*armcompute.VirtualMachineScaleSetNetworkConfiguration, error) {
	if len(networkConfigurations) == 1 {
		return networkConfigurations[0], nil
	}

	for idx := range networkConfigurations {
		networkConfig := networkConfigurations[idx]
		if networkConfig.Properties.Primary != nil && *networkConfig.Properties.Primary {
			return networkConfig, nil
		}
	}

	return nil, fmt.Errorf("failed to find a primary network configuration for the VMSS VM or VMSS %q", resource)
}

func getPrimaryIPConfigFromVMSSNetworkConfig(config *armcompute.VirtualMachineScaleSetNetworkConfiguration, backendPoolID, resource string) (*armcompute.VirtualMachineScaleSetIPConfiguration, error) {
	ipConfigurations := config.Properties.IPConfigurations
	isIPv6 := isBackendPoolIPv6(backendPoolID)

	if !isIPv6 {
		// There should be exactly one primary IP config.
		// https://learn.microsoft.com/en-us/azure/virtual-network/ip-services/virtual-network-network-interface-addresses?tabs=nic-address-portal#ip-configurations
		if len(ipConfigurations) == 1 {
			return ipConfigurations[0], nil
		}
		for idx := range ipConfigurations {
			ipConfig := ipConfigurations[idx]
			if ipConfig.Properties.Primary != nil && *ipConfig.Properties.Primary {
				return ipConfig, nil
			}
		}
	} else {
		// For IPv6 or dualstack service, we need to pick the right IP configuration based on the cluster ip family
		// IPv6 configuration is only supported as non-primary, so we need to fetch the ip configuration where the
		// privateIPAddressVersion matches the clusterIP family
		for idx := range ipConfigurations {
			ipConfig := ipConfigurations[idx]
			if ipConfig.Properties != nil && ipConfig.Properties.PrivateIPAddressVersion != nil && *ipConfig.Properties.PrivateIPAddressVersion == armcompute.IPVersionIPv6 {
				return ipConfig, nil
			}
		}
	}

	return nil, fmt.Errorf("failed to find a primary IP configuration (IPv6=%t) for the VMSS VM or VMSS %q", isIPv6, resource)
}

// EnsureHostInPool ensures the given VM's Primary NIC's Primary IP Configuration is
// participating in the specified LoadBalancer Backend Pool, which returns (resourceGroup, vmasName, instanceID, vmssVM, error).
func (ss *ScaleSet) EnsureHostInPool(ctx context.Context, _ *v1.Service, nodeName types.NodeName, backendPoolID string, vmSetNameOfLB string) (string, string, string, *armcompute.VirtualMachineScaleSetVM, error) {
	logger := log.FromContextOrBackground(ctx).WithName("EnsureHostInPool").
		WithValues("nodeName", nodeName, "backendPoolID", backendPoolID, "vmSetNameOfLB", vmSetNameOfLB)
	vmName := mapNodeNameToVMName(nodeName)
	vm, err := ss.getVmssVM(ctx, vmName, azcache.CacheReadTypeDefault)
	if err != nil {
		if errors.Is(err, cloudprovider.InstanceNotFound) {
			logger.Info("Skipping node because it is not found", "vmName", vmName)
			return "", "", "", nil, nil
		}

		logger.Error(err, "failed to get vmss vm", "vmName", vmName)
		if !errors.Is(err, ErrorNotVmssInstance) {
			return "", "", "", nil, err
		}
	}
	// In some cases (e.g., BYO nodes), we may get an ErrorNotVmssInstance error,
	// but it has been ignored above, so a nil check is needed here to prevent panic.
	if vm == nil {
		logger.Info("vmss vm not found, skip adding to backend pool", "vmName", vmName)
		return "", "", "", nil, nil
	}
	statuses := vm.GetInstanceViewStatus()
	vmPowerState := vmutil.GetVMPowerState(vm.Name, statuses)
	provisioningState := vm.GetProvisioningState()
	if vmutil.IsNotActiveVMState(provisioningState, vmPowerState) {
		logger.V(2).Info("skip updating the node because it is not in an active state", "vmName", vmName, "provisioningState", provisioningState, "vmPowerState", vmPowerState)
		return "", "", "", nil, nil
	}

	logger.V(2).Info("ensuring the vmss node in LB backendpool", "vmss name", vm.VMSSName, "vmName", vmName, "etag", ptr.Deref(vm.Etag, ""))

	// Check scale set name:
	// - For basic SKU load balancer, return nil if the node's scale set is mismatched with vmSetNameOfLB.
	// - For single standard SKU load balancer, backend could belong to multiple VMSS, so we
	//   don't check vmSet for it.
	// - For multiple standard SKU load balancers, the behavior is similar to the basic load balancer
	needCheck := !ss.UseStandardLoadBalancer()

	if vmSetNameOfLB != "" && needCheck && !strings.EqualFold(vmSetNameOfLB, vm.VMSSName) {
		logger.V(3).Info("skips the node because it is not in the ScaleSet", "vmName", vmName, "vmSetNameOfLB", vmSetNameOfLB)
		return "", "", "", nil, nil
	}

	// Find primary network interface configuration.
	if vm.VirtualMachineScaleSetVMProperties.NetworkProfileConfiguration.NetworkInterfaceConfigurations == nil {
		logger.V(4).Info("cannot obtain the primary network interface configuration, of the vm, probably because the vm's being deleted", "vmName", vmName)
		return "", "", "", nil, nil
	}

	networkInterfaceConfigurations := vm.VirtualMachineScaleSetVMProperties.NetworkProfileConfiguration.NetworkInterfaceConfigurations
	primaryNetworkInterfaceConfiguration, err := getPrimaryNetworkInterfaceConfiguration(networkInterfaceConfigurations, vmName)
	if err != nil {
		return "", "", "", nil, err
	}

	// Find primary network interface configuration.
	primaryIPConfiguration, err := getPrimaryIPConfigFromVMSSNetworkConfig(primaryNetworkInterfaceConfiguration, backendPoolID, vmName)
	if err != nil {
		return "", "", "", nil, err
	}

	// Update primary IP configuration's LoadBalancerBackendAddressPools.
	foundPool := false
	newBackendPools := []*armcompute.SubResource{}
	if primaryIPConfiguration.Properties.LoadBalancerBackendAddressPools != nil {
		newBackendPools = primaryIPConfiguration.Properties.LoadBalancerBackendAddressPools
	}
	for _, existingPool := range newBackendPools {
		if strings.EqualFold(backendPoolID, *existingPool.ID) {
			foundPool = true
			break
		}
	}

	// The backendPoolID has already been found from existing LoadBalancerBackendAddressPools.
	if foundPool {
		return "", "", "", nil, nil
	}

	if ss.UseStandardLoadBalancer() && len(newBackendPools) > 0 {
		// Although standard load balancer supports backends from multiple scale
		// sets, the same network interface couldn't be added to more than one load balancer of
		// the same type. Omit those nodes (e.g. masters) so Azure ARM won't complain
		// about this.
		newBackendPoolsIDs := make([]string, 0, len(newBackendPools))
		for _, pool := range newBackendPools {
			if pool.ID != nil {
				newBackendPoolsIDs = append(newBackendPoolsIDs, *pool.ID)
			}
		}
		isSameLB, oldLBName, err := isBackendPoolOnSameLB(backendPoolID, newBackendPoolsIDs)
		if err != nil {
			return "", "", "", nil, err
		}
		if !isSameLB {
			logger.V(4).Info("The node has already been added to an LB, omit adding it to a new one", "lbName", oldLBName)
			return "", "", "", nil, nil
		}
	}

	// Compose a new vmssVM with added backendPoolID.
	newBackendPools = append(newBackendPools,
		&armcompute.SubResource{
			ID: ptr.To(backendPoolID),
		})
	primaryIPConfiguration.Properties.LoadBalancerBackendAddressPools = newBackendPools
	newVM := &armcompute.VirtualMachineScaleSetVM{
		Location: &vm.Location,
		Properties: &armcompute.VirtualMachineScaleSetVMProperties{
			HardwareProfile: vm.VirtualMachineScaleSetVMProperties.HardwareProfile,
			NetworkProfileConfiguration: &armcompute.VirtualMachineScaleSetVMNetworkProfileConfiguration{
				NetworkInterfaceConfigurations: networkInterfaceConfigurations,
			},
		},
		Etag: vm.Etag,
	}

	// Get the node resource group.
	nodeResourceGroup, err := ss.GetNodeResourceGroup(vmName)
	if err != nil {
		return "", "", "", nil, err
	}

	return nodeResourceGroup, vm.VMSSName, vm.InstanceID, newVM, nil
}

func getVmssAndResourceGroupNameByVMProviderID(providerID string) (string, string, error) {
	matches := vmssVMProviderIDRE.FindStringSubmatch(providerID)
	if len(matches) != 3 {
		return "", "", ErrorNotVmssInstance
	}
	return matches[1], matches[2], nil
}

func getVmssAndResourceGroupNameByVMID(id string) (string, string, error) {
	matches := vmssVMResourceIDRE.FindStringSubmatch(id)
	if len(matches) != 3 {
		return "", "", ErrorNotVmssInstance
	}
	return matches[1], matches[2], nil
}

func (ss *ScaleSet) ensureVMSSInPool(ctx context.Context, _ *v1.Service, nodes []*v1.Node, backendPoolID string, vmSetNameOfLB string) error {
	logger := log.FromContextOrBackground(ctx).WithName("ensureVMSSInPool")
	logger.V(2).Info("ensuring VMSS with backendPoolID", "backendPoolID", backendPoolID)
	vmssNamesMap := make(map[string]bool)

	// the single standard load balancer supports multiple vmss in its backend while
	// multiple standard load balancers and the basic load balancer doesn't
	if ss.UseStandardLoadBalancer() {
		for _, node := range nodes {
			if ss.ExcludeMasterNodesFromStandardLB() && isControlPlaneNode(node) {
				continue
			}

			shouldExcludeLoadBalancer, err := ss.ShouldNodeExcludedFromLoadBalancer(node.Name)
			if err != nil {
				logger.Error(err, "ShouldNodeExcludedFromLoadBalancer failed", "node", node.Name)
				return err
			}
			if shouldExcludeLoadBalancer {
				logger.V(4).Info("Excluding unmanaged/external-resource-group node", "node", node.Name)
				continue
			}

			// in this scenario the vmSetName is an empty string and the name of vmss should be obtained from the provider IDs of nodes
			var resourceGroupName, vmssName string
			if node.Spec.ProviderID != "" {
				resourceGroupName, vmssName, err = getVmssAndResourceGroupNameByVMProviderID(node.Spec.ProviderID)
				if err != nil {
					logger.V(4).Info("the provider ID of node is not the format of VMSS VM, will skip checking and continue", "providerID", node.Spec.ProviderID, "node", node.Name)
					continue
				}
			} else {
				logger.V(4).Info("the provider ID of node is empty, will check the VM ID", "node", node.Name)
				instanceID, err := ss.InstanceID(ctx, types.NodeName(node.Name))
				if err != nil {
					logger.Error(err, "Failed to get instance ID for node", "node", node.Name)
					return err
				}
				resourceGroupName, vmssName, err = getVmssAndResourceGroupNameByVMID(instanceID)
				if err != nil {
					logger.V(4).Info("the instance ID of node is not the format of VMSS VM, will skip checking and continue", "instanceID", instanceID, "node", node.Name)
					continue
				}
			}
			// only vmsses in the resource group same as it's in azure config are included
			if strings.EqualFold(resourceGroupName, ss.ResourceGroup) {
				vmssNamesMap[vmssName] = true
			}
		}
	} else {
		vmssNamesMap[vmSetNameOfLB] = true
	}

	logger.V(2).Info("begins to update VMSS with backendPoolID", "VMSS", vmssNamesMap, "backendPoolID", backendPoolID)
	for vmssName := range vmssNamesMap {
		vmss, err := ss.getVMSS(ctx, vmssName, azcache.CacheReadTypeDefault)
		if err != nil {
			return err
		}

		// When vmss is being deleted, CreateOrUpdate API would report "the vmss is being deleted" error.
		// Since it is being deleted, we shouldn't send more CreateOrUpdate requests for it.
		if vmss.Properties.ProvisioningState != nil && strings.EqualFold(*vmss.Properties.ProvisioningState, consts.ProvisionStateDeleting) {
			logger.V(3).Info("found vmss being deleted, skipping", "vmss", vmssName)
			continue
		}

		if vmss.Properties.VirtualMachineProfile.NetworkProfile.NetworkInterfaceConfigurations == nil {
			logger.V(4).Info("cannot obtain the primary network interface configuration of vmss", "vmss", vmssName)
			continue
		}

		// It is possible to run Windows 2019 nodes in IPv4-only mode in a dual-stack cluster. IPv6 is not supported on
		// Windows 2019 nodes and therefore does not need to be added to the IPv6 backend pool.
		if isWindows2019(vmss) && isBackendPoolIPv6(backendPoolID) {
			logger.V(3).Info("vmss is Windows 2019, skipping adding to IPv6 backend pool", "vmss", vmssName)
			continue
		}

		vmssNIC := vmss.Properties.VirtualMachineProfile.NetworkProfile.NetworkInterfaceConfigurations
		primaryNIC, err := getPrimaryNetworkInterfaceConfiguration(vmssNIC, vmssName)
		if err != nil {
			return err
		}
		// Find primary network interface configuration.
		primaryIPConfig, err := getPrimaryIPConfigFromVMSSNetworkConfig(primaryNIC, backendPoolID, vmssName)
		if err != nil {
			return err
		}

		loadBalancerBackendAddressPools := []*armcompute.SubResource{}
		if primaryIPConfig.Properties.LoadBalancerBackendAddressPools != nil {
			loadBalancerBackendAddressPools = primaryIPConfig.Properties.LoadBalancerBackendAddressPools
		}

		var found bool
		for _, loadBalancerBackendAddressPool := range loadBalancerBackendAddressPools {
			if strings.EqualFold(*loadBalancerBackendAddressPool.ID, backendPoolID) {
				found = true
				break
			}
		}
		if found {
			continue
		}

		if ss.UseStandardLoadBalancer() && len(loadBalancerBackendAddressPools) > 0 {
			// Although standard load balancer supports backends from multiple scale
			// sets, the same network interface couldn't be added to more than one load balancer of
			// the same type. Omit those nodes (e.g. masters) so Azure ARM won't complain
			// about this.
			newBackendPoolsIDs := make([]string, 0, len(loadBalancerBackendAddressPools))
			for _, pool := range loadBalancerBackendAddressPools {
				if pool.ID != nil {
					newBackendPoolsIDs = append(newBackendPoolsIDs, *pool.ID)
				}
			}
			isSameLB, oldLBName, err := isBackendPoolOnSameLB(backendPoolID, newBackendPoolsIDs)
			if err != nil {
				return err
			}
			if !isSameLB {
				logger.V(4).Info("VMSS has already been added to LB, omit adding it to a new one", "vmss", vmssName, "LB", oldLBName)
				return nil
			}
		}

		// Compose a new vmss with added backendPoolID.
		loadBalancerBackendAddressPools = append(loadBalancerBackendAddressPools,
			&armcompute.SubResource{
				ID: ptr.To(backendPoolID),
			})
		primaryIPConfig.Properties.LoadBalancerBackendAddressPools = loadBalancerBackendAddressPools
		newVMSS := armcompute.VirtualMachineScaleSet{
			Location: vmss.Location,
			Properties: &armcompute.VirtualMachineScaleSetProperties{
				VirtualMachineProfile: &armcompute.VirtualMachineScaleSetVMProfile{
					NetworkProfile: &armcompute.VirtualMachineScaleSetNetworkProfile{
						NetworkInterfaceConfigurations: vmssNIC,
					},
				},
			},
			Etag: vmss.Etag,
		}

		// NOTE(mainred): invalidate vmss cache for the vmss is updated.
		// we invalidate the vmss cache anyway, because
		//    - when the vmss is updated, the vmss cache invalid
		//    - when the vmss update failed, we want to get fresh-new vmss in the next round of update, especially for EtagMismatch error
		defer func() {
			logger.V(2).Info("invalidating vmss cache after update", "vmss", vmssName, "reason", "vmss pool update")
			_ = ss.vmssCache.Delete(consts.VMSSKey)
		}()

		logger.V(2).Info("begins to update vmss with new backendPoolID", "vmss", vmssName, "backendPoolID", backendPoolID, "etag", ptr.Deref(vmss.Etag, ""))
		rerr := ss.CreateOrUpdateVMSS(ss.ResourceGroup, vmssName, newVMSS)
		if rerr != nil {
			logger.Error(rerr, "CreateOrUpdateVMSS failed", "vmss", vmssName, "backendPoolID", backendPoolID)
			return rerr
		}
	}
	return nil
}

// isWindows2019 checks if the ImageReference on the VMSS matches a Windows Server 2019 image.
func isWindows2019(vmss *armcompute.VirtualMachineScaleSet) bool {
	if vmss == nil {
		return false
	}

	if vmss.Properties.VirtualMachineProfile == nil || vmss.Properties.VirtualMachineProfile.StorageProfile == nil {
		return false
	}

	storageProfile := vmss.Properties.VirtualMachineProfile.StorageProfile

	if storageProfile.OSDisk == nil || *storageProfile.OSDisk.OSType != armcompute.OperatingSystemTypesWindows {
		return false
	}

	if storageProfile.ImageReference == nil || storageProfile.ImageReference.ID == nil {
		return false
	}
	// example: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/AKS-Windows/providers/Microsoft.Compute/galleries/AKSWindows/images/windows-2019-containerd/versions/17763.5820.240516
	imageRef := *storageProfile.ImageReference.ID
	parts := strings.Split(imageRef, "/")
	if len(parts) < 4 {
		return false
	}

	imageName := parts[len(parts)-3]
	if !strings.EqualFold(imageName, consts.VmssWindows2019ImageGalleryName) {
		return false
	}

	osVersion := strings.Split(parts[len(parts)-1], ".")
	if len(osVersion) != 3 {
		return false
	}
	// Windows Server 2019 is build number 17763
	// https://learn.microsoft.com/en-us/windows-server/get-started/windows-server-release-info
	return osVersion[0] == consts.Windows2019OSBuildVersion
}

func (ss *ScaleSet) ensureHostsInPool(ctx context.Context, service *v1.Service, nodes []*v1.Node, backendPoolID string, vmSetNameOfLB string) error {
	logger := log.FromContextOrBackground(ctx).WithName("ensureHostsInPool")
	mc := metrics.NewMetricContext("services", "vmss_ensure_hosts_in_pool", ss.ResourceGroup, ss.SubscriptionID, getServiceName(service))
	isOperationSucceeded := false
	defer func() {
		mc.ObserveOperationWithResult(isOperationSucceeded)
	}()

	// Track if any VM updates happened to invalidate vmss cache after all defers run
	// The cache invalidation must happen after all per-node defers complete,
	// since those defers call DeleteCacheForNode which repopulates the cache.
	hasVMUpdates := false
	defer func() {
		if hasVMUpdates {
			logger.V(2).Info("invalidating vmss cache after updating vms", "reason", "vmss vms updated")
			if err := ss.vmssCache.Delete(consts.VMSSKey); err != nil {
				logger.Info("Failed to invalidate vmss cache", "err", err)
			}
		}
	}()

	// Ensure the backendPoolID is also added on VMSS itself.
	// Refer to issue kubernetes/kubernetes#80365 for detailed information
	err := ss.ensureVMSSInPool(ctx, service, nodes, backendPoolID, vmSetNameOfLB)
	if err != nil {
		return err
	}

	hostUpdates := make([]func() error, 0, len(nodes))
	nodeUpdates := make(map[vmssMetaInfo]map[string]armcompute.VirtualMachineScaleSetVM)
	errors := make([]error, 0)
	for _, node := range nodes {
		localNodeName := node.Name

		if ss.UseStandardLoadBalancer() && ss.ExcludeMasterNodesFromStandardLB() && isControlPlaneNode(node) {
			logger.V(4).Info("Excluding master node from load balancer backendpool", "node", localNodeName, "backendPoolID", backendPoolID)
			continue
		}

		shouldExcludeLoadBalancer, err := ss.ShouldNodeExcludedFromLoadBalancer(localNodeName)
		if err != nil {
			logger.Error(err, "ShouldNodeExcludedFromLoadBalancer failed", "node", localNodeName)
			return err
		}
		if shouldExcludeLoadBalancer {
			logger.V(4).Info("Excluding unmanaged/external-resource-group node", "node", localNodeName)
			continue
		}

		nodeResourceGroup, nodeVMSS, nodeInstanceID, nodeVMSSVM, err := ss.EnsureHostInPool(ctx, service, types.NodeName(localNodeName), backendPoolID, vmSetNameOfLB)
		if err != nil {
			logger.Error(err, "EnsureHostInPool failed", "service", getServiceName(service), "backendPoolID", backendPoolID)
			errors = append(errors, err)
			continue
		}

		// No need to update if nodeVMSSVM is nil.
		if nodeVMSSVM == nil {
			continue
		}

		nodeVMSSMetaInfo := vmssMetaInfo{vmssName: nodeVMSS, resourceGroup: nodeResourceGroup}
		if v, ok := nodeUpdates[nodeVMSSMetaInfo]; ok {
			v[nodeInstanceID] = *nodeVMSSVM
		} else {
			nodeUpdates[nodeVMSSMetaInfo] = map[string]armcompute.VirtualMachineScaleSetVM{
				nodeInstanceID: *nodeVMSSVM,
			}
		}

		// Invalidate the cache since the VMSS VM would be updated.
		defer func() {
			logger.V(2).Info("invalidating vm cache after pool update", "nodeName", localNodeName, "reason", "vm added to backend pool")
			_ = ss.DeleteCacheForNode(ctx, localNodeName)
		}()
	}

	// Update VMs with best effort that have already been added to nodeUpdates.
	for meta, update := range nodeUpdates {
		// create new instance of meta and update for passing to anonymous function
		meta := meta
		update := update
		hostUpdates = append(hostUpdates, func() error {
			logFields := []interface{}{
				"operation", "EnsureHostsInPool UpdateVMSSVMs",
				"vmssName", meta.vmssName,
				"resourceGroup", meta.resourceGroup,
				"backendPoolID", backendPoolID,
			}
			logger := klog.LoggerWithValues(klog.FromContext(ctx).WithName("ensureHostsInPool.hostUpdates"), logFields...)
			batchSize, err := ss.VMSSBatchSize(ctx, meta.vmssName)
			if err != nil {
				logger.Error(err, "Failed to get vmss batch size")
				return err
			}
			ctx = klog.NewContext(ctx, logger)

			errChan := ss.UpdateVMSSVMsInBatch(ctx, meta, update, batchSize)

			errs := make([]error, 0)
			for err := range errChan {
				if err != nil {
					errs = append(errs, err)
				}
			}
			// Mark that VM updates happened so the defer will invalidate the vmss cache.
			hasVMUpdates = true
			return utilerrors.NewAggregate(errs)
		})
	}
	errs := utilerrors.AggregateGoroutines(hostUpdates...)
	if errs != nil {
		return utilerrors.Flatten(errs)
	}

	// Fail if there are other errors.
	if len(errors) > 0 {
		return utilerrors.Flatten(utilerrors.NewAggregate(errors))
	}

	isOperationSucceeded = true
	return nil
}

// EnsureHostsInPool ensures the given Node's primary IP configurations are
// participating in the specified LoadBalancer Backend Pool.
func (ss *ScaleSet) EnsureHostsInPool(ctx context.Context, service *v1.Service, nodes []*v1.Node, backendPoolID string, vmSetNameOfLB string) error {
	logger := log.FromContextOrBackground(ctx).WithName("EnsureHostsInPool")
	if ss.DisableAvailabilitySetNodes && !ss.EnableVmssFlexNodes {
		return ss.ensureHostsInPool(ctx, service, nodes, backendPoolID, vmSetNameOfLB)
	}
	vmssUniformNodes := make([]*v1.Node, 0)
	vmssFlexNodes := make([]*v1.Node, 0)
	vmasNodes := make([]*v1.Node, 0)
	errors := make([]error, 0)
	for _, node := range nodes {
		localNodeName := node.Name

		if ss.UseStandardLoadBalancer() && ss.ExcludeMasterNodesFromStandardLB() && isControlPlaneNode(node) {
			logger.V(4).Info("Excluding master node from load balancer backendpool", "node", localNodeName, "backendPoolID", backendPoolID)
			continue
		}

		shouldExcludeLoadBalancer, err := ss.ShouldNodeExcludedFromLoadBalancer(localNodeName)
		if err != nil {
			logger.Error(err, "ShouldNodeExcludedFromLoadBalancer failed", "node", localNodeName)
			return err
		}
		if shouldExcludeLoadBalancer {
			logger.V(4).Info("Excluding unmanaged/external-resource-group node", "node", localNodeName)
			continue
		}

		vmManagementType, err := ss.getVMManagementTypeByNodeName(ctx, localNodeName, azcache.CacheReadTypeDefault)
		if err != nil {
			logger.Error(err, "Failed to check vmManagementType by node name", "node", localNodeName)
			errors = append(errors, err)
			continue
		}

		if vmManagementType == ManagedByAvSet {
			// vm is managed by availability set.
			// VMAS nodes should also be added to the SLB backends.
			if ss.UseStandardLoadBalancer() {
				vmasNodes = append(vmasNodes, node)
				continue
			}
			logger.V(3).Info("Skips node because VMAS nodes couldn't be added to basic LB with VMSS backends", "node", localNodeName)
			continue
		}
		if vmManagementType == ManagedByVmssFlex {
			// vm is managed by vmss flex.
			if ss.UseStandardLoadBalancer() {
				vmssFlexNodes = append(vmssFlexNodes, node)
				continue
			}
			logger.V(3).Info("Skips node because VMSS Flex nodes do not support Basic Load Balancer", "node", localNodeName)
			continue
		}
		vmssUniformNodes = append(vmssUniformNodes, node)
	}

	if len(vmssFlexNodes) > 0 {
		vmssFlexError := ss.flexScaleSet.EnsureHostsInPool(ctx, service, vmssFlexNodes, backendPoolID, vmSetNameOfLB)
		errors = append(errors, vmssFlexError)
	}

	if len(vmasNodes) > 0 {
		vmasError := ss.availabilitySet.EnsureHostsInPool(ctx, service, vmasNodes, backendPoolID, vmSetNameOfLB)
		errors = append(errors, vmasError)
	}

	if len(vmssUniformNodes) > 0 {
		vmssUniformError := ss.ensureHostsInPool(ctx, service, vmssUniformNodes, backendPoolID, vmSetNameOfLB)
		errors = append(errors, vmssUniformError)
	}

	allErrors := utilerrors.Flatten(utilerrors.NewAggregate(errors))

	return allErrors
}

// ensureBackendPoolDeletedFromNode ensures the loadBalancer backendAddressPools deleted
// from the specified node, which returns (resourceGroup, vmasName, instanceID, vmssVM, error).
func (ss *ScaleSet) ensureBackendPoolDeletedFromNode(ctx context.Context, nodeName string, backendPoolIDs []string) (string, string, string, *armcompute.VirtualMachineScaleSetVM, error) {
	logger := log.FromContextOrBackground(ctx).WithName("ensureBackendPoolDeletedFromNode").WithValues("nodeName", nodeName, "backendPoolIDs", backendPoolIDs)
	vm, err := ss.getVmssVM(ctx, nodeName, azcache.CacheReadTypeDefault)
	if err != nil {
		if errors.Is(err, cloudprovider.InstanceNotFound) {
			logger.Info("Skipping node because it is not found", "nodeName", nodeName)
			return "", "", "", nil, nil
		}

		return "", "", "", nil, err
	}

	statuses := vm.GetInstanceViewStatus()
	vmPowerState := vmutil.GetVMPowerState(vm.Name, statuses)
	provisioningState := vm.GetProvisioningState()
	if vmutil.IsNotActiveVMState(provisioningState, vmPowerState) {
		logger.V(2).Info("skip updating the node because it is not in an active state", "provisioningState", provisioningState, "vmPowerState", vmPowerState)
		return "", "", "", nil, nil
	}

	// Find primary network interface configuration.
	if vm.VirtualMachineScaleSetVMProperties.NetworkProfileConfiguration.NetworkInterfaceConfigurations == nil {
		logger.V(4).Info("Cannot obtain the primary network interface configuration, of vm, "+
			"probably because the vm's being deleted", "vm", nodeName)
		return "", "", "", nil, nil
	}
	networkInterfaceConfigurations := vm.VirtualMachineScaleSetVMProperties.NetworkProfileConfiguration.NetworkInterfaceConfigurations
	primaryNetworkInterfaceConfiguration, err := getPrimaryNetworkInterfaceConfiguration(networkInterfaceConfigurations, nodeName)
	if err != nil {
		return "", "", "", nil, err
	}

	foundTotal := false
	for _, backendPoolID := range backendPoolIDs {
		found, err := deleteBackendPoolFromIPConfig("ensureBackendPoolDeletedFromNode", backendPoolID, nodeName, primaryNetworkInterfaceConfiguration)
		if err != nil {
			return "", "", "", nil, err
		}
		if found {
			foundTotal = true
		}
	}
	if !foundTotal {
		return "", "", "", nil, nil
	}

	// Compose a new vmssVM with added backendPoolID.
	newVM := &armcompute.VirtualMachineScaleSetVM{
		Location: &vm.Location,
		Properties: &armcompute.VirtualMachineScaleSetVMProperties{
			HardwareProfile: vm.VirtualMachineScaleSetVMProperties.HardwareProfile,
			NetworkProfileConfiguration: &armcompute.VirtualMachineScaleSetVMNetworkProfileConfiguration{
				NetworkInterfaceConfigurations: networkInterfaceConfigurations,
			},
		},
		Etag: vm.Etag,
	}

	// Get the node resource group.
	nodeResourceGroup, err := ss.GetNodeResourceGroup(nodeName)
	if err != nil {
		return "", "", "", nil, err
	}

	return nodeResourceGroup, vm.VMSSName, vm.InstanceID, newVM, nil
}

// GetNodeNameByIPConfigurationID gets the node name and the VMSS name by IP configuration ID.
func (ss *ScaleSet) GetNodeNameByIPConfigurationID(ctx context.Context, ipConfigurationID string) (string, string, error) {
	logger := log.FromContextOrBackground(ctx).WithName("GetNodeNameByIPConfigurationID")
	vmManagementType, err := ss.getVMManagementTypeByIPConfigurationID(ctx, ipConfigurationID, azcache.CacheReadTypeUnsafe)
	if err != nil {
		logger.Error(err, "Failed to check VM management type by IP configuration ID")
		return "", "", err
	}

	if vmManagementType == ManagedByAvSet {
		// vm is managed by availability set.
		return ss.availabilitySet.GetNodeNameByIPConfigurationID(ctx, ipConfigurationID)
	}
	if vmManagementType == ManagedByVmssFlex {
		// vm is managed by vmss flex.
		return ss.flexScaleSet.GetNodeNameByIPConfigurationID(ctx, ipConfigurationID)
	}

	matches := vmssIPConfigurationRE.FindStringSubmatch(ipConfigurationID)
	if len(matches) != 4 {
		return "", "", fmt.Errorf("can not extract scale set name from ipConfigurationID (%s)", ipConfigurationID)
	}

	resourceGroup := matches[1]
	scaleSetName := matches[2]
	instanceID := matches[3]
	vm, err := ss.getVmssVMByInstanceID(ctx, resourceGroup, scaleSetName, instanceID, azcache.CacheReadTypeUnsafe)
	if err != nil {
		logger.Error(err, "Unable to find node by ipConfigurationID", "ipConfigurationID", ipConfigurationID)
		return "", "", err
	}

	if vm.Properties.OSProfile != nil && vm.Properties.OSProfile.ComputerName != nil {
		return strings.ToLower(*vm.Properties.OSProfile.ComputerName), scaleSetName, nil
	}

	return "", "", nil
}

func getScaleSetAndResourceGroupNameByIPConfigurationID(ipConfigurationID string) (string, string, error) {
	logger := log.Background().WithName("getScaleSetAndResourceGroupNameByIPConfigurationID")
	matches := vmssIPConfigurationRE.FindStringSubmatch(ipConfigurationID)
	if len(matches) != 4 {
		logger.V(4).Info("Can not extract scale set name from ipConfigurationID, assuming it is managed by availability set or vmss flex", "ipConfigurationID", ipConfigurationID)
		return "", "", ErrorNotVmssInstance
	}

	resourceGroup := matches[1]
	scaleSetName := matches[2]
	return scaleSetName, resourceGroup, nil
}

func (ss *ScaleSet) ensureBackendPoolDeletedFromVMSS(ctx context.Context, backendPoolIDs []string, vmSetName string) error {
	logger := log.FromContextOrBackground(ctx).WithName("ensureBackendPoolDeletedFromVMSS")
	if !ss.UseStandardLoadBalancer() {
		found := false

		cachedUniform, err := ss.vmssCache.Get(ctx, consts.VMSSKey, azcache.CacheReadTypeDefault)
		if err != nil {
			logger.Error(err, "Failed to get vmss uniform from cache")
			return err
		}
		vmssUniformMap := cachedUniform.(*sync.Map)

		vmssUniformMap.Range(func(_, value interface{}) bool {
			vmssEntry := value.(*VMSSEntry)
			if ptr.Deref(vmssEntry.VMSS.Name, "") == vmSetName {
				found = true
				return false
			}
			return true
		})
		if found {
			return ss.ensureBackendPoolDeletedFromVmssUniform(ctx, backendPoolIDs, vmSetName)
		}

		flexScaleSet := ss.flexScaleSet.(*FlexScaleSet)
		cachedFlex, err := flexScaleSet.vmssFlexCache.Get(ctx, consts.VmssFlexKey, azcache.CacheReadTypeDefault)
		if err != nil {
			logger.Error(err, "Failed to get vmss flex from cache")
			return err
		}
		vmssFlexMap := cachedFlex.(*sync.Map)
		vmssFlexMap.Range(func(_, value interface{}) bool {
			vmssFlex := value.(*armcompute.VirtualMachineScaleSet)
			if ptr.Deref(vmssFlex.Name, "") == vmSetName {
				found = true
				return false
			}
			return true
		})

		if found {
			return flexScaleSet.ensureBackendPoolDeletedFromVmssFlex(ctx, backendPoolIDs, vmSetName)
		}

		return cloudprovider.InstanceNotFound
	}

	err := ss.ensureBackendPoolDeletedFromVmssUniform(ctx, backendPoolIDs, vmSetName)
	if err != nil {
		return err
	}
	if ss.EnableVmssFlexNodes {
		flexScaleSet := ss.flexScaleSet.(*FlexScaleSet)
		err = flexScaleSet.ensureBackendPoolDeletedFromVmssFlex(ctx, backendPoolIDs, vmSetName)
	}
	return err
}

func (ss *ScaleSet) ensureBackendPoolDeletedFromVmssUniform(ctx context.Context, backendPoolIDs []string, vmSetName string) error {
	logger := log.FromContextOrBackground(ctx).WithName("ensureBackendPoolDeletedFromVmssUniform")
	vmssNamesMap := make(map[string]bool)
	// the standard load balancer supports multiple vmss in its backend while the basic SKU doesn't
	if ss.UseStandardLoadBalancer() {
		cachedUniform, err := ss.vmssCache.Get(ctx, consts.VMSSKey, azcache.CacheReadTypeDefault)
		if err != nil {
			logger.Error(err, "Failed to get vmss uniform from cache")
			return err
		}

		vmssUniformMap := cachedUniform.(*sync.Map)
		var errorList []error
		walk := func(_, value interface{}) bool {
			var vmss *armcompute.VirtualMachineScaleSet
			if vmssEntry, ok := value.(*VMSSEntry); ok {
				vmss = vmssEntry.VMSS
			} else if v, ok := value.(*armcompute.VirtualMachineScaleSet); ok {
				vmss = v
			}
			logger.V(2).Info("Ensure backend pools are deleted from vmss uniform", "vmssName", ptr.Deref(vmss.Name, ""), "backendPoolIDs", backendPoolIDs)

			// When vmss is being deleted, CreateOrUpdate API would report "the vmss is being deleted" error.
			// Since it is being deleted, we shouldn't send more CreateOrUpdate requests for it.
			if vmss.Properties.ProvisioningState != nil && strings.EqualFold(*vmss.Properties.ProvisioningState, consts.ProvisionStateDeleting) {
				logger.V(3).Info("found vmss being deleted, skipping", "vmss", ptr.Deref(vmss.Name, ""))
				return true
			}
			if vmss.Properties.VirtualMachineProfile == nil {
				logger.V(4).Info("vmss has no VirtualMachineProfile, skipping", "vmss", ptr.Deref(vmss.Name, ""))
				return true
			}
			if vmss.Properties.VirtualMachineProfile.NetworkProfile.NetworkInterfaceConfigurations == nil {
				logger.V(4).Info("cannot obtain the primary network interface configuration, of vmss", "vmss", ptr.Deref(vmss.Name, ""))
				return true
			}
			vmssNIC := vmss.Properties.VirtualMachineProfile.NetworkProfile.NetworkInterfaceConfigurations
			primaryNIC, err := getPrimaryNetworkInterfaceConfiguration(vmssNIC, ptr.Deref(vmss.Name, ""))
			if err != nil {
				logger.Error(err, "Failed to get the primary network interface config of the VMSS", "vmss", ptr.Deref(vmss.Name, ""))
				errorList = append(errorList, err)
				return true
			}

			handleBackendPool := func(backendPoolID string) bool {
				primaryIPConfig, err := getPrimaryIPConfigFromVMSSNetworkConfig(primaryNIC, backendPoolID, ptr.Deref(vmss.Name, ""))
				if err != nil {
					logger.Error(err, "Failed to find the primary IP config from the VMSS's network config", "vmss", ptr.Deref(vmss.Name, ""))
					errorList = append(errorList, err)
					return true
				}
				loadBalancerBackendAddressPools := make([]*armcompute.SubResource, 0)
				if primaryIPConfig.Properties.LoadBalancerBackendAddressPools != nil {
					loadBalancerBackendAddressPools = primaryIPConfig.Properties.LoadBalancerBackendAddressPools
				}
				for _, loadBalancerBackendAddressPool := range loadBalancerBackendAddressPools {
					logger.V(4).Info("loadBalancerBackendAddressPool on vmss", "backendAddressPool", ptr.Deref(loadBalancerBackendAddressPool.ID, ""), "vmss", ptr.Deref(vmss.Name, ""))
					if strings.EqualFold(ptr.Deref(loadBalancerBackendAddressPool.ID, ""), backendPoolID) {
						logger.V(4).Info("found vmss with backend pool, removing it", "vmss", ptr.Deref(vmss.Name, ""), "backendPool", backendPoolID)
						vmssNamesMap[ptr.Deref(vmss.Name, "")] = true
					}
				}
				return true
			}
			for _, backendPoolID := range backendPoolIDs {
				if !handleBackendPool(backendPoolID) {
					return false
				}
			}

			return true
		}

		// Walk through all cached vmss, and find the vmss that contains the backendPoolID.
		vmssUniformMap.Range(walk)
		if len(errorList) > 0 {
			return utilerrors.Flatten(utilerrors.NewAggregate(errorList))
		}
	} else {
		logger.V(2).Info("Ensure backend pools are deleted from vmss uniform", "vmss", vmSetName, "backendPoolIDs", backendPoolIDs)
		vmssNamesMap[vmSetName] = true
	}

	return ss.EnsureBackendPoolDeletedFromVMSets(ctx, vmssNamesMap, backendPoolIDs)
}

// ensureBackendPoolDeleted ensures the loadBalancer backendAddressPools deleted from the specified nodes.
func (ss *ScaleSet) ensureBackendPoolDeleted(ctx context.Context, service *v1.Service, backendPoolIDs []string, vmSetName string, backendAddressPools []*armnetwork.BackendAddressPool) (bool, error) {
	logger := log.FromContextOrBackground(ctx).WithName("ensureBackendPoolDeleted")
	// Returns nil if backend address pools already deleted.
	if backendAddressPools == nil {
		return false, nil
	}

	mc := metrics.NewMetricContext("services", "vmss_ensure_backend_pool_deleted", ss.ResourceGroup, ss.SubscriptionID, getServiceName(service))
	isOperationSucceeded := false
	defer func() {
		mc.ObserveOperationWithResult(isOperationSucceeded)
	}()

	// Track if any VM updates happened to invalidate vmss cache after all defers run
	// The cache invalidation must happen after all per-node defers complete,
	// since those defers call DeleteCacheForNode which repopulates the cache.
	hasVMUpdates := false
	defer func() {
		if hasVMUpdates {
			logger.V(2).Info("invalidating vmss cache after updating vms", "reason", "vmss vms updated")
			if err := ss.vmssCache.Delete(consts.VMSSKey); err != nil {
				logger.Info("Failed to invalidate vmss cache", "err", err)
			}
		}
	}()

	ipConfigurationIDs := []string{}
	for _, backendPool := range backendAddressPools {
		for _, backendPoolID := range backendPoolIDs {
			if strings.EqualFold(*backendPool.ID, backendPoolID) && backendPool.Properties.BackendIPConfigurations != nil {
				for _, ipConf := range backendPool.Properties.BackendIPConfigurations {
					if ipConf.ID == nil {
						continue
					}

					ipConfigurationIDs = append(ipConfigurationIDs, *ipConf.ID)
				}
			}
		}
	}

	// Ensure the backendPoolID is deleted from the VMSS VMs.
	hostUpdates := make([]func() error, 0, len(ipConfigurationIDs))
	nodeUpdates := make(map[vmssMetaInfo]map[string]armcompute.VirtualMachineScaleSetVM)
	allErrs := make([]error, 0)
	visitedIPConfigIDPrefix := map[string]bool{}
	for i := range ipConfigurationIDs {
		ipConfigurationID := ipConfigurationIDs[i]
		ipConfigurationIDPrefix := getResourceIDPrefix(ipConfigurationID)
		if _, ok := visitedIPConfigIDPrefix[ipConfigurationIDPrefix]; ok {
			continue
		}
		visitedIPConfigIDPrefix[ipConfigurationIDPrefix] = true

		var scaleSetName string
		var err error
		if scaleSetName, err = extractScaleSetNameByProviderID(ipConfigurationID); err == nil {
			// Only remove nodes belonging to specified vmSet to basic LB backends.
			if !ss.UseStandardLoadBalancer() && !strings.EqualFold(scaleSetName, vmSetName) {
				continue
			}
		}

		nodeName, _, err := ss.GetNodeNameByIPConfigurationID(ctx, ipConfigurationID)
		if err != nil {
			if errors.Is(err, ErrorNotVmssInstance) { // Do nothing for the VMAS nodes.
				continue
			}

			if errors.Is(err, cloudprovider.InstanceNotFound) {
				logger.Info("skipping ip config because the corresponding vmss vm is not found", "service", getServiceName(service), "ipConfigurationID", ipConfigurationID)
				continue
			}

			logger.Error(err, "Failed to GetNodeNameByIPConfigurationID", "ipConfigurationID", ipConfigurationID)
			allErrs = append(allErrs, err)
			continue
		}

		nodeResourceGroup, nodeVMSS, nodeInstanceID, nodeVMSSVM, err := ss.ensureBackendPoolDeletedFromNode(ctx, nodeName, backendPoolIDs)
		if err != nil {
			if !errors.Is(err, ErrorNotVmssInstance) { // Do nothing for the VMAS nodes.
				logger.Error(err, "ensureBackendPoolDeletedFromNode failed", "service", getServiceName(service), "backendPoolIDs", backendPoolIDs)
				allErrs = append(allErrs, err)
			}
			continue
		}

		// No need to update if nodeVMSSVM is nil.
		if nodeVMSSVM == nil {
			continue
		}

		nodeVMSSMetaInfo := vmssMetaInfo{vmssName: nodeVMSS, resourceGroup: nodeResourceGroup}
		if v, ok := nodeUpdates[nodeVMSSMetaInfo]; ok {
			v[nodeInstanceID] = *nodeVMSSVM
		} else {
			nodeUpdates[nodeVMSSMetaInfo] = map[string]armcompute.VirtualMachineScaleSetVM{
				nodeInstanceID: *nodeVMSSVM,
			}
		}

		// Invalidate the cache since the VMSS VM would be updated.
		defer func() {
			logger.V(2).Info("invalidating vm cache after pool deletion", "nodeName", nodeName, "reason", "vm removed from backend pool")
			_ = ss.DeleteCacheForNode(ctx, nodeName)
		}()
	}

	// Update VMs with best effort that have already been added to nodeUpdates.
	var updatedVM atomic.Bool
	for meta, update := range nodeUpdates {
		// create new instance of meta and update for passing to anonymous function
		meta := meta
		update := update
		hostUpdates = append(hostUpdates, func() error {
			logFields := []interface{}{
				"operation", "EnsureBackendPoolDeleted UpdateVMSSVMs",
				"vmssName", meta.vmssName,
				"resourceGroup", meta.resourceGroup,
				"backendPoolIDs", backendPoolIDs,
			}

			batchSize, err := ss.VMSSBatchSize(ctx, meta.vmssName)
			if err != nil {
				logger.Error(err, "Failed to get vmss batch size", logFields...)
				return err
			}

			errChan := ss.UpdateVMSSVMsInBatch(ctx, meta, update, batchSize)
			errs := make([]error, 0)
			for err := range errChan {
				if err != nil {
					errs = append(errs, err)
				}
			}
			// Mark that VM updates happened so the defer will invalidate the vmss cache.
			hasVMUpdates = true

			err = utilerrors.NewAggregate(errs)
			if err != nil {
				logger.Error(err, "Failed to update VMs for VMSS", logFields...)
				return err
			}
			updatedVM.Store(true)
			return nil
		})
	}
	errs := utilerrors.AggregateGoroutines(hostUpdates...)
	if errs != nil {
		return updatedVM.Load(), utilerrors.Flatten(errs)
	}

	// Fail if there are other errors.
	if len(allErrs) > 0 {
		return updatedVM.Load(), utilerrors.Flatten(utilerrors.NewAggregate(allErrs))
	}

	isOperationSucceeded = true
	return updatedVM.Load(), nil
}

// EnsureBackendPoolDeleted ensures the loadBalancer backendAddressPools deleted from the specified nodes.
func (ss *ScaleSet) EnsureBackendPoolDeleted(ctx context.Context, service *v1.Service, backendPoolIDs []string, vmSetName string, backendAddressPools []*armnetwork.BackendAddressPool, deleteFromVMSet bool) (bool, error) {
	if backendAddressPools == nil {
		return false, nil
	}
	vmssUniformBackendIPConfigurationsMap := map[string][]*armnetwork.InterfaceIPConfiguration{}
	vmssFlexBackendIPConfigurationsMap := map[string][]*armnetwork.InterfaceIPConfiguration{}
	avSetBackendIPConfigurationsMap := map[string][]*armnetwork.InterfaceIPConfiguration{}

	for _, backendPool := range backendAddressPools {
		for _, backendPoolID := range backendPoolIDs {
			if strings.EqualFold(*backendPool.ID, backendPoolID) &&
				backendPool.Properties != nil &&
				backendPool.Properties.BackendIPConfigurations != nil {
				for _, ipConf := range backendPool.Properties.BackendIPConfigurations {
					if ipConf.ID == nil {
						continue
					}

					vmManagementType, err := ss.getVMManagementTypeByIPConfigurationID(ctx, *ipConf.ID, azcache.CacheReadTypeUnsafe)
					if err != nil {
						klog.Warningf("Failed to check VM management type by ipConfigurationID %s: %v, skip it", *ipConf.ID, err)
					}

					if vmManagementType == ManagedByAvSet {
						// vm is managed by availability set.
						avSetBackendIPConfigurationsMap[backendPoolID] = append(avSetBackendIPConfigurationsMap[backendPoolID], ipConf)
					}
					if vmManagementType == ManagedByVmssFlex {
						// vm is managed by vmss flex.
						vmssFlexBackendIPConfigurationsMap[backendPoolID] = append(vmssFlexBackendIPConfigurationsMap[backendPoolID], ipConf)
					}
					if vmManagementType == ManagedByVmssUniform {
						// vm is managed by vmss uniform.
						vmssUniformBackendIPConfigurationsMap[backendPoolID] = append(vmssUniformBackendIPConfigurationsMap[backendPoolID], ipConf)
					}
				}
			}
		}
	}

	// make sure all vmss including uniform and flex are decoupled from
	// the lb backend pool even if there is no ipConfigs in the backend pool.
	if deleteFromVMSet {
		err := ss.ensureBackendPoolDeletedFromVMSS(ctx, backendPoolIDs, vmSetName)
		if err != nil {
			return false, err
		}
	}

	var updated bool
	vmssUniformBackendPools := []*armnetwork.BackendAddressPool{}
	for backendPoolID, vmssUniformBackendIPConfigurations := range vmssUniformBackendIPConfigurationsMap {
		vmssUniformBackendIPConfigurations := vmssUniformBackendIPConfigurations
		vmssUniformBackendPools = append(vmssUniformBackendPools, &armnetwork.BackendAddressPool{
			ID: ptr.To(backendPoolID),
			Properties: &armnetwork.BackendAddressPoolPropertiesFormat{
				BackendIPConfigurations: vmssUniformBackendIPConfigurations,
			},
		})
	}
	if len(vmssUniformBackendPools) > 0 {
		updatedVM, err := ss.ensureBackendPoolDeleted(ctx, service, backendPoolIDs, vmSetName, vmssUniformBackendPools)
		if err != nil {
			return false, err
		}
		if updatedVM {
			updated = true
		}
	}

	vmssFlexBackendPools := []*armnetwork.BackendAddressPool{}
	for backendPoolID, vmssFlexBackendIPConfigurations := range vmssFlexBackendIPConfigurationsMap {
		vmssFlexBackendIPConfigurations := vmssFlexBackendIPConfigurations
		vmssFlexBackendPools = append(vmssFlexBackendPools, &armnetwork.BackendAddressPool{
			ID: ptr.To(backendPoolID),
			Properties: &armnetwork.BackendAddressPoolPropertiesFormat{
				BackendIPConfigurations: vmssFlexBackendIPConfigurations,
			},
		})
	}
	if len(vmssFlexBackendPools) > 0 {
		updatedNIC, err := ss.flexScaleSet.EnsureBackendPoolDeleted(ctx, service, backendPoolIDs, vmSetName, vmssFlexBackendPools, false)
		if err != nil {
			return false, err
		}
		if updatedNIC {
			updated = true
		}
	}

	avSetBackendPools := []*armnetwork.BackendAddressPool{}
	for backendPoolID, avSetBackendIPConfigurations := range avSetBackendIPConfigurationsMap {
		avSetBackendIPConfigurations := avSetBackendIPConfigurations
		avSetBackendPools = append(avSetBackendPools, &armnetwork.BackendAddressPool{
			ID: ptr.To(backendPoolID),
			Properties: &armnetwork.BackendAddressPoolPropertiesFormat{
				BackendIPConfigurations: avSetBackendIPConfigurations,
			},
		})
	}
	if len(avSetBackendPools) > 0 {
		updatedNIC, err := ss.availabilitySet.EnsureBackendPoolDeleted(ctx, service, backendPoolIDs, vmSetName, avSetBackendPools, false)
		if err != nil {
			return false, err
		}
		if updatedNIC {
			updated = true
		}
	}

	return updated, nil
}

// GetNodeCIDRMaskByProviderID returns the node CIDR subnet mask by provider ID.
func (ss *ScaleSet) GetNodeCIDRMasksByProviderID(ctx context.Context, providerID string) (int, int, error) {
	logger := log.FromContextOrBackground(ctx).WithName("GetNodeCIDRMasksByProviderID")
	vmManagementType, err := ss.getVMManagementTypeByProviderID(ctx, providerID, azcache.CacheReadTypeUnsafe)
	if err != nil {
		logger.Error(err, "Failed to check VM management type")
		return 0, 0, err
	}

	if vmManagementType == ManagedByAvSet {
		// vm is managed by availability set.
		return ss.availabilitySet.GetNodeCIDRMasksByProviderID(ctx, providerID)
	}
	if vmManagementType == ManagedByVmssFlex {
		// vm is managed by vmss flex.
		return ss.flexScaleSet.GetNodeCIDRMasksByProviderID(ctx, providerID)
	}

	_, vmssName, err := getVmssAndResourceGroupNameByVMProviderID(providerID)
	if err != nil {
		return 0, 0, err
	}

	vmss, err := ss.getVMSS(ctx, vmssName, azcache.CacheReadTypeDefault)
	if err != nil {
		return 0, 0, err
	}

	var ipv4Mask, ipv6Mask int
	if v4, ok := vmss.Tags[consts.VMSetCIDRIPV4TagKey]; ok && v4 != nil {
		ipv4Mask, err = strconv.Atoi(ptr.Deref(v4, ""))
		if err != nil {
			logger.Error(err, "Failed to parse the value of the ipv4 mask size", "value", ptr.Deref(v4, ""))
		}
	}
	if v6, ok := vmss.Tags[consts.VMSetCIDRIPV6TagKey]; ok && v6 != nil {
		ipv6Mask, err = strconv.Atoi(ptr.Deref(v6, ""))
		if err != nil {
			logger.Error(err, "Failed to parse the value of the ipv6 mask size", "value", ptr.Deref(v6, ""))
		}
	}

	return ipv4Mask, ipv6Mask, nil
}

// deleteBackendPoolFromIPConfig deletes the backend pool from the IP config.
func deleteBackendPoolFromIPConfig(msg, backendPoolID, resource string, primaryNIC *armcompute.VirtualMachineScaleSetNetworkConfiguration) (bool, error) {
	logger := log.Background().WithName("deleteBackendPoolFromIPConfig")
	primaryIPConfig, err := getPrimaryIPConfigFromVMSSNetworkConfig(primaryNIC, backendPoolID, resource)
	if err != nil {
		logger.Error(err, "Failed to get the primary IP config from the VMSS's network config", "msg", msg, "resource", resource)
		return false, err
	}
	loadBalancerBackendAddressPools := []*armcompute.SubResource{}
	if primaryIPConfig.Properties.LoadBalancerBackendAddressPools != nil {
		loadBalancerBackendAddressPools = primaryIPConfig.Properties.LoadBalancerBackendAddressPools
	}

	var found bool
	var newBackendPools []*armcompute.SubResource
	for i := len(loadBalancerBackendAddressPools) - 1; i >= 0; i-- {
		curPool := loadBalancerBackendAddressPools[i]
		if strings.EqualFold(backendPoolID, *curPool.ID) {
			logger.V(10).Info("gets unwanted backend pool for VMSS OR VMSS VM", "msg", msg, "backendPoolID", backendPoolID, "resource", resource)
			found = true
			newBackendPools = append(loadBalancerBackendAddressPools[:i], loadBalancerBackendAddressPools[i+1:]...)
		}
	}
	if !found {
		return false, nil
	}
	primaryIPConfig.Properties.LoadBalancerBackendAddressPools = newBackendPools
	return true, nil
}

// EnsureBackendPoolDeletedFromVMSets ensures the loadBalancer backendAddressPools deleted from the specified VMSS
func (ss *ScaleSet) EnsureBackendPoolDeletedFromVMSets(ctx context.Context, vmssNamesMap map[string]bool, backendPoolIDs []string) error {
	logger := log.FromContextOrBackground(ctx).WithName("EnsureBackendPoolDeletedFromVMSets")
	vmssUpdaters := make([]func() error, 0, len(vmssNamesMap))
	errors := make([]error, 0, len(vmssNamesMap))
	for vmssName := range vmssNamesMap {
		vmssName := vmssName
		vmss, err := ss.getVMSS(ctx, vmssName, azcache.CacheReadTypeDefault)
		if err != nil {
			logger.Error(err, "Failed to get VMSS", "vmss", vmssName)
			errors = append(errors, err)
			continue
		}

		// When vmss is being deleted, CreateOrUpdate API would report "the vmss is being deleted" error.
		// Since it is being deleted, we shouldn't send more CreateOrUpdate requests for it.
		if vmss.Properties.ProvisioningState != nil && strings.EqualFold(*vmss.Properties.ProvisioningState, consts.ProvisionStateDeleting) {
			logger.V(3).Info("found vmss being deleted, skipping", "vmss", vmssName)
			continue
		}
		if vmss.Properties.VirtualMachineProfile.NetworkProfile.NetworkInterfaceConfigurations == nil {
			logger.V(4).Info("cannot obtain the primary network interface configuration, of vmss", "vmss", vmssName)
			continue
		}
		vmssNIC := vmss.Properties.VirtualMachineProfile.NetworkProfile.NetworkInterfaceConfigurations
		primaryNIC, err := getPrimaryNetworkInterfaceConfiguration(vmssNIC, vmssName)
		if err != nil {
			logger.Error(err, "Failed to get the primary network interface config of the VMSS", "vmss", vmssName)
			errors = append(errors, err)
			continue
		}
		foundTotal := false
		for _, backendPoolID := range backendPoolIDs {
			found, err := deleteBackendPoolFromIPConfig("EnsureBackendPoolDeletedFromVMSets", backendPoolID, vmssName, primaryNIC)
			if err != nil {
				errors = append(errors, err)
				continue
			}
			if found {
				foundTotal = true
			}
		}
		if !foundTotal {
			continue
		}

		vmssUpdaters = append(vmssUpdaters, func() error {
			// Compose a new vmss with added backendPoolID.
			newVMSS := armcompute.VirtualMachineScaleSet{
				Location: vmss.Location,
				Properties: &armcompute.VirtualMachineScaleSetProperties{
					VirtualMachineProfile: &armcompute.VirtualMachineScaleSetVMProfile{
						NetworkProfile: &armcompute.VirtualMachineScaleSetNetworkProfile{
							NetworkInterfaceConfigurations: vmssNIC,
						},
					},
				},
				Etag: vmss.Etag,
			}

			defer func() {
				logger.V(2).Info("invalidating vmss cache after update", "vmss", vmssName, "reason", "backend pool deletion")
				_ = ss.vmssCache.Delete(consts.VMSSKey)
			}()

			logger.V(2).Info("begins to update vmss with backendPoolIDs", "vmss", vmssName, "backendPoolIDs", backendPoolIDs, "etag", ptr.Deref(vmss.Etag, ""))
			rerr := ss.CreateOrUpdateVMSS(ss.ResourceGroup, vmssName, newVMSS)
			if rerr != nil {
				logger.Error(rerr, "CreateOrUpdateVMSS failed with new backendPoolIDs", "vmss", vmssName, "backendPoolIDs", backendPoolIDs)
				return rerr
			}

			return nil
		})
	}

	errs := utilerrors.AggregateGoroutines(vmssUpdaters...)
	if errs != nil {
		return utilerrors.Flatten(errs)
	}
	// Fail if there are other errors.
	if len(errors) > 0 {
		return utilerrors.Flatten(utilerrors.NewAggregate(errors))
	}

	return nil
}

// GetAgentPoolVMSetNames returns all VMSS/VMAS names according to the nodes.
// We need to include the VMAS here because some of the cluster provisioning tools
// like capz allows mixed instance type.
func (ss *ScaleSet) GetAgentPoolVMSetNames(ctx context.Context, nodes []*v1.Node) ([]*string, error) {
	vmSetNames := make([]*string, 0)

	vmssFlexVMNodes := make([]*v1.Node, 0)
	avSetVMNodes := make([]*v1.Node, 0)

	for _, node := range nodes {
		var names []string

		vmManagementType, err := ss.getVMManagementTypeByNodeName(ctx, node.Name, azcache.CacheReadTypeDefault)
		if err != nil {
			return nil, fmt.Errorf("GetAgentPoolVMSetNames: failed to check the node %s management type: %w", node.Name, err)
		}

		if vmManagementType == ManagedByAvSet {
			// vm is managed by vmss flex.
			avSetVMNodes = append(avSetVMNodes, node)
			continue
		}
		if vmManagementType == ManagedByVmssFlex {
			// vm is managed by vmss flex.
			vmssFlexVMNodes = append(vmssFlexVMNodes, node)
			continue
		}

		names, err = ss.getAgentPoolScaleSets(ctx, []*v1.Node{node})
		if err != nil {
			return nil, fmt.Errorf("GetAgentPoolVMSetNames: failed to execute getAgentPoolScaleSets: %w", err)
		}
		vmSetNames = append(vmSetNames, to.SliceOfPtrs(names...)...)
	}

	if len(vmssFlexVMNodes) > 0 {
		vmssFlexVMnames, err := ss.flexScaleSet.GetAgentPoolVMSetNames(ctx, vmssFlexVMNodes)
		if err != nil {
			return nil, fmt.Errorf("ss.flexScaleSet.GetAgentPoolVMSetNames: failed to execute : %w", err)
		}
		vmSetNames = append(vmSetNames, vmssFlexVMnames...)
	}

	if len(avSetVMNodes) > 0 {
		avSetVMnames, err := ss.availabilitySet.GetAgentPoolVMSetNames(ctx, avSetVMNodes)
		if err != nil {
			return nil, fmt.Errorf("ss.availabilitySet.GetAgentPoolVMSetNames: failed to execute : %w", err)
		}
		vmSetNames = append(vmSetNames, avSetVMnames...)
	}

	return vmSetNames, nil
}

func (ss *ScaleSet) GetNodeVMSetName(ctx context.Context, node *v1.Node) (string, error) {
	logger := log.FromContextOrBackground(ctx).WithName("GetNodeVMSetName")
	vmManagementType, err := ss.getVMManagementTypeByNodeName(ctx, node.Name, azcache.CacheReadTypeUnsafe)
	if err != nil {
		logger.Error(err, "Failed to check VM management type")
		return "", err
	}

	if vmManagementType == ManagedByAvSet {
		// vm is managed by availability set.
		return ss.availabilitySet.GetNodeVMSetName(ctx, node)
	}
	if vmManagementType == ManagedByVmssFlex {
		// vm is managed by vmss flex.
		return ss.flexScaleSet.GetNodeVMSetName(ctx, node)
	}

	providerID := node.Spec.ProviderID
	_, vmssName, err := getVmssAndResourceGroupNameByVMProviderID(providerID)
	if err != nil {
		logger.Error(err, "getVmssAndResourceGroupNameByVMProviderID failed")
		return "", err
	}

	logger.V(4).Info("found vmss name from node name", "vmssName", vmssName, "nodeName", node.Name)
	return vmssName, nil
}

// VMSSBatchSize returns the batch size for VMSS operations.
func (ss *ScaleSet) VMSSBatchSize(ctx context.Context, vmssName string) (int, error) {
	logger := log.FromContextOrBackground(ctx).WithName("VMSSBatchSize")
	batchSize := 1
	vmss, err := ss.getVMSS(ctx, vmssName, azcache.CacheReadTypeDefault)
	if err != nil {
		return 0, fmt.Errorf("get vmss batch size: %w", err)
	}
	if _, ok := vmss.Tags[consts.VMSSTagForBatchOperation]; ok {
		batchSize = ss.GetPutVMSSVMBatchSize()
	}
	if batchSize < 1 {
		batchSize = 1
	}
	logger.V(2).Info("Fetch VMSS batch size", "vmss", vmssName, "size", batchSize)
	return batchSize, nil
}

func (ss *ScaleSet) UpdateVMSSVMsInBatch(ctx context.Context, meta vmssMetaInfo, update map[string]armcompute.VirtualMachineScaleSetVM, batchSize int) <-chan error {
	logger := klog.FromContext(ctx).WithName("UpdateVMSSVMsInBatch")
	patchVMFn := func(ctx context.Context, instanceID string, vm *armcompute.VirtualMachineScaleSetVM) (*runtime.Poller[armcompute.VirtualMachineScaleSetVMsClientUpdateResponse], error) {
		logger.V(4).Info("Updating vm", "vmss", meta.vmssName, "instanceID", instanceID, "requestEtag", ptr.Deref(vm.Etag, ""))
		return ss.ComputeClientFactory.GetVirtualMachineScaleSetVMClient().BeginUpdate(ctx, meta.resourceGroup, meta.vmssName, instanceID, *vm, &armcompute.VirtualMachineScaleSetVMsClientBeginUpdateOptions{
			IfMatch: vm.Etag,
		})
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	errChan := make(chan error, len(update))
	pollerChannel := make(chan *runtime.Poller[armcompute.VirtualMachineScaleSetVMsClientUpdateResponse], len(update))
	var pollerGroup sync.WaitGroup
	pollerGroup.Add(1)
	go func() {
		defer pollerGroup.Done()
		for {
			select {
			case poller, ok := <-pollerChannel:
				if !ok {
					// pollerChannel is closed
					return
				}
				if poller == nil {
					continue
				}
				pollerGroup.Add(1)
				go func() {
					defer pollerGroup.Done()
					resp, err := poller.PollUntilDone(ctx, &runtime.PollUntilDoneOptions{
						Frequency: 10 * time.Second,
					})
					if err != nil {
						logger.Error(err, "Failed to update VMs for VMSS with new vm config")
						errChan <- err
					} else {
						logger.V(6).Info("Successfully updated vm", "responseEtag", ptr.Deref(resp.Etag, ""))
					}
				}()
			case <-ctx.Done():
				return
			}
		}
	}()
	// logger.Info("Begin to update VMs for VMSS with new vm config")
	if batchSize > 1 {
		concurrentLock := make(chan struct{}, batchSize)
		var requestGroup sync.WaitGroup
		for instanceID, vm := range update {
			instanceID := instanceID
			vm := vm
			concurrentLock <- struct{}{}
			requestGroup.Add(1)
			go func() {
				defer func() {
					requestGroup.Done()
					<-concurrentLock
				}()
				poller, rerr := patchVMFn(ctx, instanceID, &vm)
				if rerr != nil {
					errChan <- rerr
					return
				}
				if poller != nil {
					pollerChannel <- poller
				}
			}()
		}
		requestGroup.Wait()
		close(concurrentLock)
	} else {
		for instanceID, vm := range update {
			instanceID := instanceID
			vm := vm
			poller, rerr := patchVMFn(ctx, instanceID, &vm)
			if rerr != nil {
				errChan <- rerr
				continue
			}
			if poller != nil {
				pollerChannel <- poller
			}
		}
	}
	close(pollerChannel)
	pollerGroup.Wait()
	close(errChan)
	return errChan
}
