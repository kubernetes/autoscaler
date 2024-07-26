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

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2022-08-01/compute"
	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2022-07-01/network"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	cloudprovider "k8s.io/cloud-provider"
	"k8s.io/klog/v2"
	"k8s.io/utils/pointer"
	"k8s.io/utils/ptr"

	azcache "sigs.k8s.io/cloud-provider-azure/pkg/cache"
	"sigs.k8s.io/cloud-provider-azure/pkg/consts"
	"sigs.k8s.io/cloud-provider-azure/pkg/metrics"
	"sigs.k8s.io/cloud-provider-azure/pkg/provider/virtualmachine"
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
	lockMap *LockMap
}

// newScaleSet creates a new ScaleSet.
func newScaleSet(ctx context.Context, az *Cloud) (VMSet, error) {
	if az.Config.VmssVirtualMachinesCacheTTLInSeconds == 0 {
		az.Config.VmssVirtualMachinesCacheTTLInSeconds = consts.VMSSVirtualMachinesCacheTTLDefaultInSeconds
	}

	var err error
	as, err := newAvailabilitySet(az)
	if err != nil {
		return nil, err
	}
	fs, err := newFlexScaleSet(ctx, az)
	if err != nil {
		return nil, err
	}

	ss := &ScaleSet{
		Cloud:           az,
		availabilitySet: as,
		flexScaleSet:    fs,
		lockMap:         newLockMap(),
	}

	if !ss.DisableAvailabilitySetNodes || ss.EnableVmssFlexNodes {
		ss.nonVmssUniformNodesCache, err = ss.newNonVmssUniformNodesCache()
		if err != nil {
			return nil, err
		}
	}

	ss.vmssCache, err = ss.newVMSSCache(ctx)
	if err != nil {
		return nil, err
	}

	ss.vmssVMCache, err = ss.newVMSSVirtualMachinesCache()
	if err != nil {
		return nil, err
	}

	ss.lockMap = newLockMap()
	return ss, nil
}

func (ss *ScaleSet) getVMSS(vmssName string, crt azcache.AzureCacheReadType) (*compute.VirtualMachineScaleSet, error) {
	getter := func(vmssName string) (*compute.VirtualMachineScaleSet, error) {
		cached, err := ss.vmssCache.Get(consts.VMSSKey, crt)
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
		return vmss, nil
	}

	klog.V(2).Infof("Couldn't find VMSS with name %s, refreshing the cache", vmssName)
	_ = ss.vmssCache.Delete(consts.VMSSKey)
	vmss, err = getter(vmssName)
	if err != nil {
		return nil, err
	}

	if vmss == nil {
		return nil, cloudprovider.InstanceNotFound
	}
	return vmss, nil
}

// getVmssVMByNodeIdentity find virtualMachineScaleSetVM by nodeIdentity, using node's parent VMSS cache.
// Returns cloudprovider.InstanceNotFound if the node does not belong to the scale set named in nodeIdentity.
func (ss *ScaleSet) getVmssVMByNodeIdentity(node *nodeIdentity, crt azcache.AzureCacheReadType) (*virtualmachine.VirtualMachine, error) {
	// FIXME(ccc): check only if vmss is uniform.
	_, err := getScaleSetVMInstanceID(node.nodeName)
	if err != nil {
		return nil, err
	}

	getter := func(crt azcache.AzureCacheReadType) (*virtualmachine.VirtualMachine, bool, error) {
		var found bool
		virtualMachines, err := ss.getVMSSVMsFromCache(node.resourceGroup, node.vmssName, crt)
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

	vm, found, err := getter(crt)
	if err != nil {
		return nil, err
	}

	if !found {
		cacheKey := getVMSSVMCacheKey(node.resourceGroup, node.vmssName)
		// lock and try find nodeName from cache again, refresh cache if still not found
		ss.lockMap.LockEntry(cacheKey)
		defer ss.lockMap.UnlockEntry(cacheKey)
		vm, found, err = getter(crt)
		if err == nil && found && vm != nil {
			klog.V(2).Infof("found VMSS VM with nodeName %s after retry", node.nodeName)
			return vm, nil
		}

		klog.V(2).Infof("Couldn't find VMSS VM with nodeName %s, refreshing the cache(vmss: %s, rg: %s)", node.nodeName, node.vmssName, node.resourceGroup)
		vm, found, err = getter(azcache.CacheReadTypeForceRefresh)
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
func (ss *ScaleSet) getVmssVM(nodeName string, crt azcache.AzureCacheReadType) (*virtualmachine.VirtualMachine, error) {
	node, err := ss.getNodeIdentityByNodeName(nodeName, crt)
	if err != nil {
		return nil, err
	}

	return ss.getVmssVMByNodeIdentity(node, crt)
}

// GetPowerStatusByNodeName returns the power state of the specified node.
func (ss *ScaleSet) GetPowerStatusByNodeName(name string) (powerState string, err error) {
	vmManagementType, err := ss.getVMManagementTypeByNodeName(name, azcache.CacheReadTypeUnsafe)
	if err != nil {
		klog.Errorf("Failed to check VM management type: %v", err)
		return "", err
	}

	if vmManagementType == ManagedByAvSet {
		// vm is managed by availability set.
		return ss.availabilitySet.GetPowerStatusByNodeName(name)
	}
	if vmManagementType == ManagedByVmssFlex {
		// vm is managed by vmss flex.
		return ss.flexScaleSet.GetPowerStatusByNodeName(name)
	}
	// VM is managed by vmss
	vm, err := ss.getVmssVM(name, azcache.CacheReadTypeDefault)
	if err != nil {
		return powerState, err
	}

	if vm.IsVirtualMachineScaleSetVM() {
		v := vm.AsVirtualMachineScaleSetVM()
		if v.InstanceView != nil {
			return vmutil.GetVMPowerState(ptr.Deref(v.Name, ""), v.InstanceView.Statuses), nil
		}
	}

	// vm.InstanceView or vm.InstanceView.Statuses are nil when the VM is under deleting.
	klog.V(3).Infof("InstanceView for node %q is nil, assuming it's deleting", name)
	return consts.VMPowerStateUnknown, nil
}

// GetProvisioningStateByNodeName returns the provisioningState for the specified node.
func (ss *ScaleSet) GetProvisioningStateByNodeName(name string) (provisioningState string, err error) {
	vmManagementType, err := ss.getVMManagementTypeByNodeName(name, azcache.CacheReadTypeUnsafe)
	if err != nil {
		klog.Errorf("Failed to check VM management type: %v", err)
		return "", err
	}

	if vmManagementType == ManagedByAvSet {
		// vm is managed by availability set.
		return ss.availabilitySet.GetProvisioningStateByNodeName(name)
	}
	if vmManagementType == ManagedByVmssFlex {
		// vm is managed by vmss flex.
		return ss.flexScaleSet.GetProvisioningStateByNodeName(name)
	}

	vm, err := ss.getVmssVM(name, azcache.CacheReadTypeDefault)
	if err != nil {
		return provisioningState, err
	}

	if vm.VirtualMachineScaleSetVMProperties == nil || vm.VirtualMachineScaleSetVMProperties.ProvisioningState == nil {
		return provisioningState, nil
	}

	return pointer.StringDeref(vm.VirtualMachineScaleSetVMProperties.ProvisioningState, ""), nil
}

// getCachedVirtualMachineByInstanceID gets scaleSetVMInfo from cache.
// The node must belong to one of scale sets.
func (ss *ScaleSet) getVmssVMByInstanceID(resourceGroup, scaleSetName, instanceID string, crt azcache.AzureCacheReadType) (*compute.VirtualMachineScaleSetVM, error) {
	getter := func(crt azcache.AzureCacheReadType) (vm *compute.VirtualMachineScaleSetVM, found bool, err error) {
		virtualMachines, err := ss.getVMSSVMsFromCache(resourceGroup, scaleSetName, crt)
		if err != nil {
			return nil, false, err
		}

		virtualMachines.Range(func(key, value interface{}) bool {
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

	vm, found, err := getter(crt)
	if err != nil {
		return nil, err
	}
	if !found {
		klog.V(2).Infof("Couldn't find VMSS VM with scaleSetName %q and instanceID %q, refreshing the cache", scaleSetName, instanceID)
		vm, found, err = getter(azcache.CacheReadTypeForceRefresh)
		if err != nil {
			return nil, err
		}
	}
	if found && vm != nil {
		return vm, nil
	}
	if found && vm == nil {
		klog.V(2).Infof("Couldn't find VMSS VM with scaleSetName %q and instanceID %q, refreshing the cache if it is expired", scaleSetName, instanceID)
		vm, found, err = getter(azcache.CacheReadTypeDefault)
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
func (ss *ScaleSet) GetInstanceIDByNodeName(name string) (string, error) {
	vmManagementType, err := ss.getVMManagementTypeByNodeName(name, azcache.CacheReadTypeUnsafe)
	if err != nil {
		klog.Errorf("Failed to check VM management type: %v", err)
		return "", err
	}

	if vmManagementType == ManagedByAvSet {
		// vm is managed by availability set.
		return ss.availabilitySet.GetInstanceIDByNodeName(name)
	}
	if vmManagementType == ManagedByVmssFlex {
		// vm is managed by vmss flex.
		return ss.flexScaleSet.GetInstanceIDByNodeName(name)
	}

	vm, err := ss.getVmssVM(name, azcache.CacheReadTypeUnsafe)
	if err != nil {
		// special case: during scaling in, if the vm is deleted and nonVmssUniformNodesCache is refreshed,
		// then getVMManagementTypeByNodeName will return ManagedByVmssUniform no matter what the actual managementType is.
		// In this case, if it is actually a non vmss uniform node, return InstanceNotFound
		if errors.Is(err, ErrorNotVmssInstance) {
			return "", cloudprovider.InstanceNotFound
		}
		klog.Errorf("Unable to find node %s: %v", name, err)
		return "", err
	}

	resourceID := vm.ID
	convertedResourceID, err := ConvertResourceGroupNameToLower(resourceID)
	if err != nil {
		klog.Errorf("ConvertResourceGroupNameToLower failed with error: %v", err)
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
func (ss *ScaleSet) GetNodeNameByProviderID(providerID string) (types.NodeName, error) {

	vmManagementType, err := ss.getVMManagementTypeByProviderID(providerID, azcache.CacheReadTypeUnsafe)
	if err != nil {
		klog.Errorf("Failed to check VM management type: %v", err)
		return "", err
	}

	if vmManagementType == ManagedByAvSet {
		// vm is managed by availability set.
		return ss.availabilitySet.GetNodeNameByProviderID(providerID)
	}
	if vmManagementType == ManagedByVmssFlex {
		// vm is managed by vmss flex.
		return ss.flexScaleSet.GetNodeNameByProviderID(providerID)
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
		klog.V(4).Infof("Can not extract instanceID from providerID (%s), assuming it is managed by availability set: %v", providerID, err)
		return ss.availabilitySet.GetNodeNameByProviderID(providerID)
	}

	// instanceID contains scaleSetName (returned by disk.ManagedBy), e.g. k8s-agentpool-36841236-vmss_1
	if strings.HasPrefix(strings.ToLower(instanceID), strings.ToLower(scaleSetName)) {
		instanceID, err = getLastSegment(instanceID, "_")
		if err != nil {
			return "", err
		}
	}

	vm, err := ss.getVmssVMByInstanceID(resourceGroup, scaleSetName, instanceID, azcache.CacheReadTypeUnsafe)
	if err != nil {
		klog.Errorf("Unable to find node by providerID %s: %v", providerID, err)
		return "", err
	}

	if vm.OsProfile != nil && vm.OsProfile.ComputerName != nil {
		nodeName := strings.ToLower(*vm.OsProfile.ComputerName)
		return types.NodeName(nodeName), nil
	}

	return "", nil
}

// GetInstanceTypeByNodeName gets the instance type by node name.
func (ss *ScaleSet) GetInstanceTypeByNodeName(name string) (string, error) {
	vmManagementType, err := ss.getVMManagementTypeByNodeName(name, azcache.CacheReadTypeUnsafe)
	if err != nil {
		klog.Errorf("Failed to check VM management type: %v", err)
		return "", err
	}

	if vmManagementType == ManagedByAvSet {
		// vm is managed by availability set.
		return ss.availabilitySet.GetInstanceTypeByNodeName(name)
	}
	if vmManagementType == ManagedByVmssFlex {
		// vm is managed by vmss flex.
		return ss.flexScaleSet.GetInstanceTypeByNodeName(name)
	}

	vm, err := ss.getVmssVM(name, azcache.CacheReadTypeUnsafe)
	if err != nil {
		return "", err
	}

	if vm.IsVirtualMachineScaleSetVM() {
		v := vm.AsVirtualMachineScaleSetVM()
		if v.Sku != nil && v.Sku.Name != nil {
			return *v.Sku.Name, nil
		}
	}

	return "", nil
}

// GetZoneByNodeName gets availability zone for the specified node. If the node is not running
// with availability zone, then it returns fault domain.
func (ss *ScaleSet) GetZoneByNodeName(name string) (cloudprovider.Zone, error) {
	vmManagementType, err := ss.getVMManagementTypeByNodeName(name, azcache.CacheReadTypeUnsafe)
	if err != nil {
		klog.Errorf("Failed to check VM management type: %v", err)
		return cloudprovider.Zone{}, err
	}

	if vmManagementType == ManagedByAvSet {
		// vm is managed by availability set.
		return ss.availabilitySet.GetZoneByNodeName(name)
	}
	if vmManagementType == ManagedByVmssFlex {
		// vm is managed by vmss flex.
		return ss.flexScaleSet.GetZoneByNodeName(name)
	}

	vm, err := ss.getVmssVM(name, azcache.CacheReadTypeUnsafe)
	if err != nil {
		return cloudprovider.Zone{}, err
	}

	var failureDomain string
	if vm.Zones != nil && len(vm.Zones) > 0 {
		// Get availability zone for the node.
		zones := vm.Zones
		zoneID, err := strconv.Atoi(zones[0])
		if err != nil {
			return cloudprovider.Zone{}, fmt.Errorf("failed to parse zone %q: %w", zones, err)
		}

		failureDomain = ss.makeZone(vm.Location, zoneID)
	} else if vm.IsVirtualMachineScaleSetVM() &&
		vm.AsVirtualMachineScaleSetVM().InstanceView != nil &&
		vm.AsVirtualMachineScaleSetVM().InstanceView.PlatformFaultDomain != nil {
		// Availability zone is not used for the node, falling back to fault domain.
		failureDomain = strconv.Itoa(int(*vm.AsVirtualMachineScaleSetVM().InstanceView.PlatformFaultDomain))
	} else {
		err = fmt.Errorf("failed to get zone info")
		klog.Errorf("GetZoneByNodeName: got unexpected error %v", err)
		_ = ss.DeleteCacheForNode(name)
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
	return ss.Config.PrimaryScaleSetName
}

// GetIPByNodeName gets machine private IP and public IP by node name.
func (ss *ScaleSet) GetIPByNodeName(nodeName string) (string, string, error) {
	vmManagementType, err := ss.getVMManagementTypeByNodeName(nodeName, azcache.CacheReadTypeUnsafe)
	if err != nil {
		klog.Errorf("Failed to check VM management type: %v", err)
		return "", "", err
	}

	if vmManagementType == ManagedByAvSet {
		// vm is managed by availability set.
		return ss.availabilitySet.GetIPByNodeName(nodeName)
	}
	if vmManagementType == ManagedByVmssFlex {
		// vm is managed by vmss flex.
		return ss.flexScaleSet.GetIPByNodeName(nodeName)
	}

	nic, err := ss.GetPrimaryInterface(nodeName)
	if err != nil {
		klog.Errorf("error: ss.GetIPByNodeName(%s), GetPrimaryInterface(%q), err=%v", nodeName, nodeName, err)
		return "", "", err
	}

	ipConfig, err := getPrimaryIPConfig(nic)
	if err != nil {
		klog.Errorf("error: ss.GetIPByNodeName(%s), getPrimaryIPConfig(%v), err=%v", nodeName, nic, err)
		return "", "", err
	}

	internalIP := *ipConfig.PrivateIPAddress
	publicIP := ""
	if ipConfig.PublicIPAddress != nil && ipConfig.PublicIPAddress.ID != nil {
		pipID := *ipConfig.PublicIPAddress.ID
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
				klog.Errorf("ss.getVMSSPublicIPAddress() failed with error: %v", err)
				return "", "", err
			}
			if existsPip && pip.IPAddress != nil {
				publicIP = *pip.IPAddress
			}
		} else {
			klog.Warningf("Failed to get VMSS Public IP with ID %s", pipID)
		}
	}

	return internalIP, publicIP, nil
}

func (ss *ScaleSet) getVMSSPublicIPAddress(resourceGroupName string, virtualMachineScaleSetName string, virtualMachineIndex string, networkInterfaceName string, IPConfigurationName string, publicIPAddressName string) (network.PublicIPAddress, bool, error) {
	ctx, cancel := getContextWithCancel()
	defer cancel()

	pip, err := ss.PublicIPAddressesClient.GetVirtualMachineScaleSetPublicIPAddress(ctx, resourceGroupName, virtualMachineScaleSetName, virtualMachineIndex, networkInterfaceName, IPConfigurationName, publicIPAddressName, "")
	exists, rerr := checkResourceExistsFromError(err)
	if rerr != nil {
		return pip, false, rerr.Error()
	}

	if !exists {
		klog.V(2).Infof("Public IP %q not found", publicIPAddressName)
		return pip, false, nil
	}

	return pip, exists, nil
}

// returns a list of private ips assigned to node
// TODO (khenidak): This should read all nics, not just the primary
// allowing users to split ipv4/v6 on multiple nics
func (ss *ScaleSet) GetPrivateIPsByNodeName(nodeName string) ([]string, error) {
	ips := make([]string, 0)
	vmManagementType, err := ss.getVMManagementTypeByNodeName(nodeName, azcache.CacheReadTypeUnsafe)
	if err != nil {
		klog.Errorf("Failed to check VM management type: %v", err)
		return ips, err
	}

	if vmManagementType == ManagedByAvSet {
		// vm is managed by availability set.
		return ss.availabilitySet.GetPrivateIPsByNodeName(nodeName)
	}
	if vmManagementType == ManagedByVmssFlex {
		// vm is managed by vmss flex.
		return ss.flexScaleSet.GetPrivateIPsByNodeName(nodeName)
	}

	nic, err := ss.GetPrimaryInterface(nodeName)
	if err != nil {
		klog.Errorf("error: ss.GetIPByNodeName(%s), GetPrimaryInterface(%q), err=%v", nodeName, nodeName, err)
		return ips, err
	}

	if nic.IPConfigurations == nil {
		return ips, fmt.Errorf("nic.IPConfigurations for nic (nicname=%q) is nil", *nic.Name)
	}

	for _, ipConfig := range *(nic.IPConfigurations) {
		if ipConfig.PrivateIPAddress != nil {
			ips = append(ips, *(ipConfig.PrivateIPAddress))
		}
	}

	return ips, nil
}

// This returns the full identifier of the primary NIC for the given VM.
func (ss *ScaleSet) getPrimaryInterfaceID(vm *virtualmachine.VirtualMachine) (string, error) {
	machine := vm.AsVirtualMachineScaleSetVM()
	if machine.NetworkProfile == nil || machine.NetworkProfile.NetworkInterfaces == nil {
		return "", fmt.Errorf("failed to find the network interfaces for vm %s", pointer.StringDeref(machine.Name, ""))
	}

	if len(*machine.NetworkProfile.NetworkInterfaces) == 1 {
		return *(*machine.NetworkProfile.NetworkInterfaces)[0].ID, nil
	}

	for _, ref := range *machine.NetworkProfile.NetworkInterfaces {
		if pointer.BoolDeref(ref.Primary, false) {
			return *ref.ID, nil
		}
	}

	return "", fmt.Errorf("failed to find a primary nic for the vm. vmname=%q", pointer.StringDeref(machine.Name, ""))
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
func (ss *ScaleSet) getNodeIdentityByNodeName(nodeName string, crt azcache.AzureCacheReadType) (*nodeIdentity, error) {
	getter := func(nodeName string, crt azcache.AzureCacheReadType) (*nodeIdentity, error) {
		node := &nodeIdentity{
			nodeName: nodeName,
		}

		cached, err := ss.vmssCache.Get(consts.VMSSKey, crt)
		if err != nil {
			return nil, err
		}

		vmsses := cached.(*sync.Map)
		vmsses.Range(func(key, value interface{}) bool {
			v := value.(*VMSSEntry)
			if v.VMSS.Name == nil {
				return true
			}

			vmssPrefix := *v.VMSS.Name
			if v.VMSS.VirtualMachineProfile != nil &&
				v.VMSS.VirtualMachineProfile.OsProfile != nil &&
				v.VMSS.VirtualMachineProfile.OsProfile.ComputerNamePrefix != nil {
				vmssPrefix = *v.VMSS.VirtualMachineProfile.OsProfile.ComputerNamePrefix
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

	klog.V(2).Infof("Couldn't find VMSS for node %s, refreshing the cache", nodeName)
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
func (ss *ScaleSet) listScaleSetVMs(scaleSetName, resourceGroup string) ([]compute.VirtualMachineScaleSetVM, error) {
	ctx, cancel := getContextWithCancel()
	defer cancel()

	allVMs, rerr := ss.VirtualMachineScaleSetVMsClient.List(ctx, resourceGroup, scaleSetName, string(compute.InstanceViewTypesInstanceView))
	if rerr != nil {
		klog.Errorf("VirtualMachineScaleSetVMsClient.List(%s, %s) failed: %v", resourceGroup, scaleSetName, rerr)
		if rerr.IsNotFound() {
			return nil, cloudprovider.InstanceNotFound
		}
		return nil, rerr.Error()
	}

	return allVMs, nil
}

// getAgentPoolScaleSets lists the virtual machines for the resource group and then builds
// a list of scale sets that match the nodes available to k8s.
func (ss *ScaleSet) getAgentPoolScaleSets(nodes []*v1.Node) (*[]string, error) {
	agentPoolScaleSets := &[]string{}
	for nx := range nodes {
		if isControlPlaneNode(nodes[nx]) {
			continue
		}

		nodeName := nodes[nx].Name
		shouldExcludeLoadBalancer, err := ss.ShouldNodeExcludedFromLoadBalancer(nodeName)
		if err != nil {
			klog.Errorf("ShouldNodeExcludedFromLoadBalancer(%s) failed with error: %v", nodeName, err)
			return nil, err
		}
		if shouldExcludeLoadBalancer {
			continue
		}

		vm, err := ss.getVmssVM(nodeName, azcache.CacheReadTypeDefault)
		if err != nil {
			return nil, err
		}

		if vm.VMSSName == "" {
			klog.V(3).Infof("Node %q is not belonging to any known scale sets", nodeName)
			continue
		}

		*agentPoolScaleSets = append(*agentPoolScaleSets, vm.VMSSName)
	}

	return agentPoolScaleSets, nil
}

// GetVMSetNames selects all possible scale sets for service load balancer. If the service has
// no loadbalancer mode annotation returns the primary VMSet. If service annotation
// for loadbalancer exists then return the eligible VMSet.
func (ss *ScaleSet) GetVMSetNames(service *v1.Service, nodes []*v1.Node) (*[]string, error) {
	hasMode, isAuto, serviceVMSetName := ss.getServiceLoadBalancerMode(service)
	if !hasMode || ss.useStandardLoadBalancer() {
		// no mode specified in service annotation or use single SLB mode
		// default to PrimaryScaleSetName
		scaleSetNames := &[]string{ss.Config.PrimaryScaleSetName}
		return scaleSetNames, nil
	}

	scaleSetNames, err := ss.GetAgentPoolVMSetNames(nodes)
	if err != nil {
		klog.Errorf("ss.GetVMSetNames - GetAgentPoolVMSetNames failed err=(%v)", err)
		return nil, err
	}
	if len(*scaleSetNames) == 0 {
		klog.Errorf("ss.GetVMSetNames - No scale sets found for nodes in the cluster, node count(%d)", len(nodes))
		return nil, fmt.Errorf("no scale sets found for nodes, node count(%d)", len(nodes))
	}

	if !isAuto {
		found := false
		for asx := range *scaleSetNames {
			if strings.EqualFold((*scaleSetNames)[asx], serviceVMSetName) {
				found = true
				serviceVMSetName = (*scaleSetNames)[asx]
				break
			}
		}
		if !found {
			klog.Errorf("ss.GetVMSetNames - scale set (%s) in service annotation not found", serviceVMSetName)
			return nil, ErrScaleSetNotFound
		}
		return &[]string{serviceVMSetName}, nil
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
func (ss *ScaleSet) GetPrimaryInterface(nodeName string) (network.Interface, error) {
	vmManagementType, err := ss.getVMManagementTypeByNodeName(nodeName, azcache.CacheReadTypeUnsafe)
	if err != nil {
		klog.Errorf("Failed to check VM management type: %v", err)
		return network.Interface{}, err
	}

	if vmManagementType == ManagedByAvSet {
		// vm is managed by availability set.
		return ss.availabilitySet.GetPrimaryInterface(nodeName)
	}
	if vmManagementType == ManagedByVmssFlex {
		// vm is managed by vmss flex.
		return ss.flexScaleSet.GetPrimaryInterface(nodeName)
	}

	vm, err := ss.getVmssVM(nodeName, azcache.CacheReadTypeDefault)
	if err != nil {
		// VM is availability set, but not cached yet in availabilitySetNodesCache.
		if errors.Is(err, ErrorNotVmssInstance) {
			return ss.availabilitySet.GetPrimaryInterface(nodeName)
		}

		klog.Errorf("error: ss.GetPrimaryInterface(%s), ss.getVmssVM(%s), err=%v", nodeName, nodeName, err)
		return network.Interface{}, err
	}

	primaryInterfaceID, err := ss.getPrimaryInterfaceID(vm)
	if err != nil {
		klog.Errorf("error: ss.GetPrimaryInterface(%s), ss.getPrimaryInterfaceID(), err=%v", nodeName, err)
		return network.Interface{}, err
	}

	nicName, err := getLastSegment(primaryInterfaceID, "/")
	if err != nil {
		klog.Errorf("error: ss.GetPrimaryInterface(%s), getLastSegment(%s), err=%v", nodeName, primaryInterfaceID, err)
		return network.Interface{}, err
	}
	resourceGroup, err := extractResourceGroupByVMSSNicID(primaryInterfaceID)
	if err != nil {
		return network.Interface{}, err
	}

	ctx, cancel := getContextWithCancel()
	defer cancel()
	nic, rerr := ss.InterfacesClient.GetVirtualMachineScaleSetNetworkInterface(ctx, resourceGroup, vm.VMSSName,
		vm.InstanceID,
		nicName, "")
	if rerr != nil {
		exists, realErr := checkResourceExistsFromError(rerr)
		if realErr != nil {
			klog.Errorf("error: ss.GetPrimaryInterface(%s), ss.GetVirtualMachineScaleSetNetworkInterface.Get(%s, %s, %s), err=%v", nodeName, resourceGroup, vm.VMSSName, nicName, realErr)
			return network.Interface{}, realErr.Error()
		}

		if !exists {
			return network.Interface{}, cloudprovider.InstanceNotFound
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
func getPrimaryNetworkInterfaceConfiguration(networkConfigurations []compute.VirtualMachineScaleSetNetworkConfiguration, resource string) (*compute.VirtualMachineScaleSetNetworkConfiguration, error) {
	if len(networkConfigurations) == 1 {
		return &networkConfigurations[0], nil
	}

	for idx := range networkConfigurations {
		networkConfig := &networkConfigurations[idx]
		if networkConfig.Primary != nil && *networkConfig.Primary {
			return networkConfig, nil
		}
	}

	return nil, fmt.Errorf("failed to find a primary network configuration for the VMSS VM or VMSS %q", resource)
}

func getPrimaryIPConfigFromVMSSNetworkConfig(config *compute.VirtualMachineScaleSetNetworkConfiguration, backendPoolID, resource string) (*compute.VirtualMachineScaleSetIPConfiguration, error) {
	ipConfigurations := *config.IPConfigurations
	isIPv6 := isBackendPoolIPv6(backendPoolID)

	if !isIPv6 {
		// There should be exactly one primary IP config.
		// https://learn.microsoft.com/en-us/azure/virtual-network/ip-services/virtual-network-network-interface-addresses?tabs=nic-address-portal#ip-configurations
		if len(ipConfigurations) == 1 {
			return &ipConfigurations[0], nil
		}
		for idx := range ipConfigurations {
			ipConfig := &ipConfigurations[idx]
			if ipConfig.Primary != nil && *ipConfig.Primary {
				return ipConfig, nil
			}
		}
	} else {
		// For IPv6 or dualstack service, we need to pick the right IP configuration based on the cluster ip family
		// IPv6 configuration is only supported as non-primary, so we need to fetch the ip configuration where the
		// privateIPAddressVersion matches the clusterIP family
		for idx := range ipConfigurations {
			ipConfig := &ipConfigurations[idx]
			if ipConfig.PrivateIPAddressVersion == compute.IPv6 {
				return ipConfig, nil
			}
		}
	}

	return nil, fmt.Errorf("failed to find a primary IP configuration (IPv6=%t) for the VMSS VM or VMSS %q", isIPv6, resource)
}

// EnsureHostInPool ensures the given VM's Primary NIC's Primary IP Configuration is
// participating in the specified LoadBalancer Backend Pool, which returns (resourceGroup, vmasName, instanceID, vmssVM, error).
func (ss *ScaleSet) EnsureHostInPool(_ *v1.Service, nodeName types.NodeName, backendPoolID string, vmSetNameOfLB string) (string, string, string, *compute.VirtualMachineScaleSetVM, error) {
	logger := klog.Background().WithName("EnsureHostInPool").
		WithValues("nodeName", nodeName, "backendPoolID", backendPoolID, "vmSetNameOfLB", vmSetNameOfLB)
	vmName := mapNodeNameToVMName(nodeName)
	vm, err := ss.getVmssVM(vmName, azcache.CacheReadTypeDefault)
	if err != nil {
		if errors.Is(err, cloudprovider.InstanceNotFound) {
			klog.Infof("EnsureHostInPool: skipping node %s because it is not found", vmName)
			return "", "", "", nil, nil
		}

		logger.Error(err, "failed to get vmss vm", "vmName", vmName)
		if !errors.Is(err, ErrorNotVmssInstance) {
			return "", "", "", nil, err
		}
	}
	statuses := vm.GetInstanceViewStatus()
	vmPowerState := vmutil.GetVMPowerState(vm.Name, statuses)
	provisioningState := vm.GetProvisioningState()
	if vmutil.IsNotActiveVMState(provisioningState, vmPowerState) {
		logger.V(2).Info("skip updating the node because it is not in an active state", "vmName", vmName, "provisioningState", provisioningState, "vmPowerState", vmPowerState)
		return "", "", "", nil, nil
	}

	logger.V(2).Info("ensuring the vmss node in LB backendpool", "vmss name", vm.VMSSName)

	// Check scale set name:
	// - For basic SKU load balancer, return nil if the node's scale set is mismatched with vmSetNameOfLB.
	// - For single standard SKU load balancer, backend could belong to multiple VMSS, so we
	//   don't check vmSet for it.
	// - For multiple standard SKU load balancers, the behavior is similar to the basic load balancer
	needCheck := false
	if !ss.useStandardLoadBalancer() {
		// need to check the vmSet name when using the basic LB
		needCheck = true
	}

	if vmSetNameOfLB != "" && needCheck && !strings.EqualFold(vmSetNameOfLB, vm.VMSSName) {
		logger.V(3).Info("skips the node %s because it is not in the ScaleSet %s", vmName, vmSetNameOfLB)
		return "", "", "", nil, nil
	}

	// Find primary network interface configuration.
	if vm.VirtualMachineScaleSetVMProperties.NetworkProfileConfiguration.NetworkInterfaceConfigurations == nil {
		logger.V(4).Info("cannot obtain the primary network interface configuration, of the vm, probably because the vm's being deleted", "vmName", vmName)
		return "", "", "", nil, nil
	}

	networkInterfaceConfigurations := *vm.VirtualMachineScaleSetVMProperties.NetworkProfileConfiguration.NetworkInterfaceConfigurations
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
	newBackendPools := []compute.SubResource{}
	if primaryIPConfiguration.LoadBalancerBackendAddressPools != nil {
		newBackendPools = *primaryIPConfiguration.LoadBalancerBackendAddressPools
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

	if ss.useStandardLoadBalancer() && len(newBackendPools) > 0 {
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
		compute.SubResource{
			ID: pointer.String(backendPoolID),
		})
	primaryIPConfiguration.LoadBalancerBackendAddressPools = &newBackendPools
	newVM := &compute.VirtualMachineScaleSetVM{
		Location: &vm.Location,
		VirtualMachineScaleSetVMProperties: &compute.VirtualMachineScaleSetVMProperties{
			HardwareProfile: vm.VirtualMachineScaleSetVMProperties.HardwareProfile,
			NetworkProfileConfiguration: &compute.VirtualMachineScaleSetVMNetworkProfileConfiguration{
				NetworkInterfaceConfigurations: &networkInterfaceConfigurations,
			},
		},
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

func (ss *ScaleSet) ensureVMSSInPool(_ *v1.Service, nodes []*v1.Node, backendPoolID string, vmSetNameOfLB string) error {
	klog.V(2).Infof("ensureVMSSInPool: ensuring VMSS with backendPoolID %s", backendPoolID)
	vmssNamesMap := make(map[string]bool)

	// the single standard load balancer supports multiple vmss in its backend while
	// multiple standard load balancers and the basic load balancer doesn't
	if ss.useStandardLoadBalancer() {
		for _, node := range nodes {
			if ss.excludeMasterNodesFromStandardLB() && isControlPlaneNode(node) {
				continue
			}

			shouldExcludeLoadBalancer, err := ss.ShouldNodeExcludedFromLoadBalancer(node.Name)
			if err != nil {
				klog.Errorf("ShouldNodeExcludedFromLoadBalancer(%s) failed with error: %v", node.Name, err)
				return err
			}
			if shouldExcludeLoadBalancer {
				klog.V(4).Infof("Excluding unmanaged/external-resource-group node %q", node.Name)
				continue
			}

			// in this scenario the vmSetName is an empty string and the name of vmss should be obtained from the provider IDs of nodes
			var resourceGroupName, vmssName string
			if node.Spec.ProviderID != "" {
				resourceGroupName, vmssName, err = getVmssAndResourceGroupNameByVMProviderID(node.Spec.ProviderID)
				if err != nil {
					klog.V(4).Infof("ensureVMSSInPool: the provider ID %s of node %s is not the format of VMSS VM, will skip checking and continue", node.Spec.ProviderID, node.Name)
					continue
				}
			} else {
				klog.V(4).Infof("ensureVMSSInPool: the provider ID of node %s is empty, will check the VM ID", node.Name)
				instanceID, err := ss.InstanceID(context.TODO(), types.NodeName(node.Name))
				if err != nil {
					klog.Errorf("ensureVMSSInPool: Failed to get instance ID for node %q: %v", node.Name, err)
					return err
				}
				resourceGroupName, vmssName, err = getVmssAndResourceGroupNameByVMID(instanceID)
				if err != nil {
					klog.V(4).Infof("ensureVMSSInPool: the instance ID %s of node %s is not the format of VMSS VM, will skip checking and continue", node.Spec.ProviderID, node.Name)
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

	klog.V(2).Infof("ensureVMSSInPool begins to update VMSS %v with backendPoolID %s", vmssNamesMap, backendPoolID)
	for vmssName := range vmssNamesMap {
		vmss, err := ss.getVMSS(vmssName, azcache.CacheReadTypeDefault)
		if err != nil {
			return err
		}

		// When vmss is being deleted, CreateOrUpdate API would report "the vmss is being deleted" error.
		// Since it is being deleted, we shouldn't send more CreateOrUpdate requests for it.
		if vmss.ProvisioningState != nil && strings.EqualFold(*vmss.ProvisioningState, consts.ProvisionStateDeleting) {
			klog.V(3).Infof("ensureVMSSInPool: found vmss %s being deleted, skipping", vmssName)
			continue
		}

		if vmss.VirtualMachineProfile.NetworkProfile.NetworkInterfaceConfigurations == nil {
			klog.V(4).Infof("EnsureHostInPool: cannot obtain the primary network interface configuration of vmss %s", vmssName)
			continue
		}
		vmssNIC := *vmss.VirtualMachineProfile.NetworkProfile.NetworkInterfaceConfigurations
		primaryNIC, err := getPrimaryNetworkInterfaceConfiguration(vmssNIC, vmssName)
		if err != nil {
			return err
		}
		// Find primary network interface configuration.
		primaryIPConfig, err := getPrimaryIPConfigFromVMSSNetworkConfig(primaryNIC, backendPoolID, vmssName)
		if err != nil {
			return err
		}

		loadBalancerBackendAddressPools := []compute.SubResource{}
		if primaryIPConfig.LoadBalancerBackendAddressPools != nil {
			loadBalancerBackendAddressPools = *primaryIPConfig.LoadBalancerBackendAddressPools
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

		if ss.useStandardLoadBalancer() && len(loadBalancerBackendAddressPools) > 0 {
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
				klog.V(4).Infof("VMSS %q has already been added to LB %q, omit adding it to a new one", vmssName, oldLBName)
				return nil
			}
		}

		// Compose a new vmss with added backendPoolID.
		loadBalancerBackendAddressPools = append(loadBalancerBackendAddressPools,
			compute.SubResource{
				ID: pointer.String(backendPoolID),
			})
		primaryIPConfig.LoadBalancerBackendAddressPools = &loadBalancerBackendAddressPools
		newVMSS := compute.VirtualMachineScaleSet{
			Location: vmss.Location,
			VirtualMachineScaleSetProperties: &compute.VirtualMachineScaleSetProperties{
				VirtualMachineProfile: &compute.VirtualMachineScaleSetVMProfile{
					NetworkProfile: &compute.VirtualMachineScaleSetNetworkProfile{
						NetworkInterfaceConfigurations: &vmssNIC,
					},
				},
			},
		}

		klog.V(2).Infof("ensureVMSSInPool begins to update vmss(%s) with new backendPoolID %s", vmssName, backendPoolID)
		rerr := ss.CreateOrUpdateVMSS(ss.ResourceGroup, vmssName, newVMSS)
		if rerr != nil {
			klog.Errorf("ensureVMSSInPool CreateOrUpdateVMSS(%s) with new backendPoolID %s, err: %v", vmssName, backendPoolID, err)
			return rerr.Error()
		}
	}
	return nil
}

func (ss *ScaleSet) ensureHostsInPool(service *v1.Service, nodes []*v1.Node, backendPoolID string, vmSetNameOfLB string) error {
	mc := metrics.NewMetricContext("services", "vmss_ensure_hosts_in_pool", ss.ResourceGroup, ss.SubscriptionID, getServiceName(service))
	isOperationSucceeded := false
	defer func() {
		mc.ObserveOperationWithResult(isOperationSucceeded)
	}()

	hostUpdates := make([]func() error, 0, len(nodes))
	nodeUpdates := make(map[vmssMetaInfo]map[string]compute.VirtualMachineScaleSetVM)
	errors := make([]error, 0)
	for _, node := range nodes {
		localNodeName := node.Name

		if ss.useStandardLoadBalancer() && ss.excludeMasterNodesFromStandardLB() && isControlPlaneNode(node) {
			klog.V(4).Infof("Excluding master node %q from load balancer backendpool %q", localNodeName, backendPoolID)
			continue
		}

		shouldExcludeLoadBalancer, err := ss.ShouldNodeExcludedFromLoadBalancer(localNodeName)
		if err != nil {
			klog.Errorf("ShouldNodeExcludedFromLoadBalancer(%s) failed with error: %v", localNodeName, err)
			return err
		}
		if shouldExcludeLoadBalancer {
			klog.V(4).Infof("Excluding unmanaged/external-resource-group node %q", localNodeName)
			continue
		}

		nodeResourceGroup, nodeVMSS, nodeInstanceID, nodeVMSSVM, err := ss.EnsureHostInPool(service, types.NodeName(localNodeName), backendPoolID, vmSetNameOfLB)
		if err != nil {
			klog.Errorf("EnsureHostInPool(%s): backendPoolID(%s) - failed to ensure host in pool: %q", getServiceName(service), backendPoolID, err)
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
			nodeUpdates[nodeVMSSMetaInfo] = map[string]compute.VirtualMachineScaleSetVM{
				nodeInstanceID: *nodeVMSSVM,
			}
		}

		// Invalidate the cache since the VMSS VM would be updated.
		defer func() {
			_ = ss.DeleteCacheForNode(localNodeName)
		}()
	}

	// Update VMs with best effort that have already been added to nodeUpdates.
	for meta, update := range nodeUpdates {
		// create new instance of meta and update for passing to anonymous function
		meta := meta
		update := update
		hostUpdates = append(hostUpdates, func() error {
			ctx, cancel := getContextWithCancel()
			defer cancel()

			logFields := []interface{}{
				"operation", "EnsureHostsInPool UpdateVMSSVMs",
				"vmssName", meta.vmssName,
				"resourceGroup", meta.resourceGroup,
				"backendPoolID", backendPoolID,
			}

			batchSize, err := ss.VMSSBatchSize(meta.vmssName)
			if err != nil {
				klog.ErrorS(err, "Failed to get vmss batch size", logFields...)
				return err
			}

			klog.V(2).InfoS("Begin to update VMs for VMSS with new backendPoolID", logFields...)
			rerr := ss.VirtualMachineScaleSetVMsClient.UpdateVMs(ctx, meta.resourceGroup, meta.vmssName, update, "network_update", batchSize)
			if rerr != nil {
				klog.ErrorS(err, "Failed to update VMs for VMSS", logFields...)
				return rerr.Error()
			}

			return nil
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

	// Ensure the backendPoolID is also added on VMSS itself.
	// Refer to issue kubernetes/kubernetes#80365 for detailed information
	err := ss.ensureVMSSInPool(service, nodes, backendPoolID, vmSetNameOfLB)
	if err != nil {
		return err
	}

	isOperationSucceeded = true
	return nil
}

// EnsureHostsInPool ensures the given Node's primary IP configurations are
// participating in the specified LoadBalancer Backend Pool.
func (ss *ScaleSet) EnsureHostsInPool(service *v1.Service, nodes []*v1.Node, backendPoolID string, vmSetNameOfLB string) error {
	if ss.DisableAvailabilitySetNodes && !ss.EnableVmssFlexNodes {
		return ss.ensureHostsInPool(service, nodes, backendPoolID, vmSetNameOfLB)
	}
	vmssUniformNodes := make([]*v1.Node, 0)
	vmssFlexNodes := make([]*v1.Node, 0)
	vmasNodes := make([]*v1.Node, 0)
	errors := make([]error, 0)
	for _, node := range nodes {
		localNodeName := node.Name

		if ss.useStandardLoadBalancer() && ss.excludeMasterNodesFromStandardLB() && isControlPlaneNode(node) {
			klog.V(4).Infof("Excluding master node %q from load balancer backendpool %q", localNodeName, backendPoolID)
			continue
		}

		shouldExcludeLoadBalancer, err := ss.ShouldNodeExcludedFromLoadBalancer(localNodeName)
		if err != nil {
			klog.Errorf("ShouldNodeExcludedFromLoadBalancer(%s) failed with error: %v", localNodeName, err)
			return err
		}
		if shouldExcludeLoadBalancer {
			klog.V(4).Infof("Excluding unmanaged/external-resource-group node %q", localNodeName)
			continue
		}

		vmManagementType, err := ss.getVMManagementTypeByNodeName(localNodeName, azcache.CacheReadTypeDefault)
		if err != nil {
			klog.Errorf("Failed to check vmManagementType(%s): %v", localNodeName, err)
			errors = append(errors, err)
			continue
		}

		if vmManagementType == ManagedByAvSet {
			// vm is managed by availability set.
			// VMAS nodes should also be added to the SLB backends.
			if ss.useStandardLoadBalancer() {
				vmasNodes = append(vmasNodes, node)
				continue
			}
			klog.V(3).Infof("EnsureHostsInPool skips node %s because VMAS nodes couldn't be added to basic LB with VMSS backends", localNodeName)
			continue
		}
		if vmManagementType == ManagedByVmssFlex {
			// vm is managed by vmss flex.
			if ss.useStandardLoadBalancer() {
				vmssFlexNodes = append(vmssFlexNodes, node)
				continue
			}
			klog.V(3).Infof("EnsureHostsInPool skips node %s because VMSS Flex nodes deos not support Basic Load Balancer", localNodeName)
			continue
		}
		vmssUniformNodes = append(vmssUniformNodes, node)
	}

	if len(vmssFlexNodes) > 0 {
		vmssFlexError := ss.flexScaleSet.EnsureHostsInPool(service, vmssFlexNodes, backendPoolID, vmSetNameOfLB)
		errors = append(errors, vmssFlexError)
	}

	if len(vmasNodes) > 0 {
		vmasError := ss.availabilitySet.EnsureHostsInPool(service, vmasNodes, backendPoolID, vmSetNameOfLB)
		errors = append(errors, vmasError)
	}

	if len(vmssUniformNodes) > 0 {
		vmssUniformError := ss.ensureHostsInPool(service, vmssUniformNodes, backendPoolID, vmSetNameOfLB)
		errors = append(errors, vmssUniformError)
	}

	allErrors := utilerrors.Flatten(utilerrors.NewAggregate(errors))

	return allErrors
}

// ensureBackendPoolDeletedFromNode ensures the loadBalancer backendAddressPools deleted
// from the specified node, which returns (resourceGroup, vmasName, instanceID, vmssVM, error).
func (ss *ScaleSet) ensureBackendPoolDeletedFromNode(nodeName string, backendPoolIDs []string) (string, string, string, *compute.VirtualMachineScaleSetVM, error) {
	logger := klog.Background().WithName("ensureBackendPoolDeletedFromNode").WithValues("nodeName", nodeName, "backendPoolIDs", backendPoolIDs)
	vm, err := ss.getVmssVM(nodeName, azcache.CacheReadTypeDefault)
	if err != nil {
		if errors.Is(err, cloudprovider.InstanceNotFound) {
			klog.Infof("ensureBackendPoolDeletedFromNode: skipping node %s because it is not found", nodeName)
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
		klog.V(4).Infof("ensureBackendPoolDeletedFromNode: cannot obtain the primary network interface configuration, of vm %s, "+
			"probably because the vm's being deleted", nodeName)
		return "", "", "", nil, nil
	}
	networkInterfaceConfigurations := *vm.VirtualMachineScaleSetVMProperties.NetworkProfileConfiguration.NetworkInterfaceConfigurations
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
	newVM := &compute.VirtualMachineScaleSetVM{
		Location: &vm.Location,
		VirtualMachineScaleSetVMProperties: &compute.VirtualMachineScaleSetVMProperties{
			HardwareProfile: vm.VirtualMachineScaleSetVMProperties.HardwareProfile,
			NetworkProfileConfiguration: &compute.VirtualMachineScaleSetVMNetworkProfileConfiguration{
				NetworkInterfaceConfigurations: &networkInterfaceConfigurations,
			},
		},
	}

	// Get the node resource group.
	nodeResourceGroup, err := ss.GetNodeResourceGroup(nodeName)
	if err != nil {
		return "", "", "", nil, err
	}

	return nodeResourceGroup, vm.VMSSName, vm.InstanceID, newVM, nil
}

// GetNodeNameByIPConfigurationID gets the node name and the VMSS name by IP configuration ID.
func (ss *ScaleSet) GetNodeNameByIPConfigurationID(ipConfigurationID string) (string, string, error) {
	vmManagementType, err := ss.getVMManagementTypeByIPConfigurationID(ipConfigurationID, azcache.CacheReadTypeUnsafe)
	if err != nil {
		klog.Errorf("Failed to check VM management type: %v", err)
		return "", "", err
	}

	if vmManagementType == ManagedByAvSet {
		// vm is managed by availability set.
		return ss.availabilitySet.GetNodeNameByIPConfigurationID(ipConfigurationID)
	}
	if vmManagementType == ManagedByVmssFlex {
		// vm is managed by vmss flex.
		return ss.flexScaleSet.GetNodeNameByIPConfigurationID(ipConfigurationID)
	}

	matches := vmssIPConfigurationRE.FindStringSubmatch(ipConfigurationID)
	if len(matches) != 4 {
		return "", "", fmt.Errorf("can not extract scale set name from ipConfigurationID (%s)", ipConfigurationID)
	}

	resourceGroup := matches[1]
	scaleSetName := matches[2]
	instanceID := matches[3]
	vm, err := ss.getVmssVMByInstanceID(resourceGroup, scaleSetName, instanceID, azcache.CacheReadTypeUnsafe)
	if err != nil {
		klog.Errorf("Unable to find node by ipConfigurationID %s: %v", ipConfigurationID, err)
		return "", "", err
	}

	if vm.OsProfile != nil && vm.OsProfile.ComputerName != nil {
		return strings.ToLower(*vm.OsProfile.ComputerName), scaleSetName, nil
	}

	return "", "", nil
}

func getScaleSetAndResourceGroupNameByIPConfigurationID(ipConfigurationID string) (string, string, error) {
	matches := vmssIPConfigurationRE.FindStringSubmatch(ipConfigurationID)
	if len(matches) != 4 {
		klog.V(4).Infof("Can not extract scale set name from ipConfigurationID (%s), assuming it is managed by availability set or vmss flex", ipConfigurationID)
		return "", "", ErrorNotVmssInstance
	}

	resourceGroup := matches[1]
	scaleSetName := matches[2]
	return scaleSetName, resourceGroup, nil
}

func (ss *ScaleSet) ensureBackendPoolDeletedFromVMSS(backendPoolIDs []string, vmSetName string) error {
	if !ss.useStandardLoadBalancer() {
		found := false

		cachedUniform, err := ss.vmssCache.Get(consts.VMSSKey, azcache.CacheReadTypeDefault)
		if err != nil {
			klog.Errorf("ensureBackendPoolDeletedFromVMSS: failed to get vmss uniform from cache: %v", err)
			return err
		}
		vmssUniformMap := cachedUniform.(*sync.Map)

		vmssUniformMap.Range(func(key, value interface{}) bool {
			vmssEntry := value.(*VMSSEntry)
			if pointer.StringDeref(vmssEntry.VMSS.Name, "") == vmSetName {
				found = true
				return false
			}
			return true
		})
		if found {
			return ss.ensureBackendPoolDeletedFromVmssUniform(backendPoolIDs, vmSetName)
		}

		flexScaleSet := ss.flexScaleSet.(*FlexScaleSet)
		cachedFlex, err := flexScaleSet.vmssFlexCache.Get(consts.VmssFlexKey, azcache.CacheReadTypeDefault)
		if err != nil {
			klog.Errorf("ensureBackendPoolDeletedFromVMSS: failed to get vmss flex from cache: %v", err)
			return err
		}
		vmssFlexMap := cachedFlex.(*sync.Map)
		vmssFlexMap.Range(func(key, value interface{}) bool {
			vmssFlex := value.(*compute.VirtualMachineScaleSet)
			if pointer.StringDeref(vmssFlex.Name, "") == vmSetName {
				found = true
				return false
			}
			return true
		})

		if found {
			return flexScaleSet.ensureBackendPoolDeletedFromVmssFlex(backendPoolIDs, vmSetName)
		}

		return cloudprovider.InstanceNotFound
	}

	err := ss.ensureBackendPoolDeletedFromVmssUniform(backendPoolIDs, vmSetName)
	if err != nil {
		return err
	}
	if ss.EnableVmssFlexNodes {
		flexScaleSet := ss.flexScaleSet.(*FlexScaleSet)
		err = flexScaleSet.ensureBackendPoolDeletedFromVmssFlex(backendPoolIDs, vmSetName)
	}
	return err
}

func (ss *ScaleSet) ensureBackendPoolDeletedFromVmssUniform(backendPoolIDs []string, vmSetName string) error {
	vmssNamesMap := make(map[string]bool)
	// the standard load balancer supports multiple vmss in its backend while the basic sku doesn't
	if ss.useStandardLoadBalancer() {
		cachedUniform, err := ss.vmssCache.Get(consts.VMSSKey, azcache.CacheReadTypeDefault)
		if err != nil {
			klog.Errorf("ensureBackendPoolDeletedFromVMSS: failed to get vmss uniform from cache: %v", err)
			return err
		}

		vmssUniformMap := cachedUniform.(*sync.Map)
		var errorList []error
		walk := func(key, value interface{}) bool {
			var vmss *compute.VirtualMachineScaleSet
			if vmssEntry, ok := value.(*VMSSEntry); ok {
				vmss = vmssEntry.VMSS
			} else if v, ok := value.(*compute.VirtualMachineScaleSet); ok {
				vmss = v
			}
			klog.V(2).Infof("ensureBackendPoolDeletedFromVmssUniform: vmss %q, backendPoolIDs %q", pointer.StringDeref(vmss.Name, ""), backendPoolIDs)

			// When vmss is being deleted, CreateOrUpdate API would report "the vmss is being deleted" error.
			// Since it is being deleted, we shouldn't send more CreateOrUpdate requests for it.
			if vmss.ProvisioningState != nil && strings.EqualFold(*vmss.ProvisioningState, consts.ProvisionStateDeleting) {
				klog.V(3).Infof("ensureBackendPoolDeletedFromVMSS: found vmss %s being deleted, skipping", pointer.StringDeref(vmss.Name, ""))
				return true
			}
			if vmss.VirtualMachineProfile == nil {
				klog.V(4).Infof("ensureBackendPoolDeletedFromVMSS: vmss %s has no VirtualMachineProfile, skipping", pointer.StringDeref(vmss.Name, ""))
				return true
			}
			if vmss.VirtualMachineProfile.NetworkProfile.NetworkInterfaceConfigurations == nil {
				klog.V(4).Infof("ensureBackendPoolDeletedFromVMSS: cannot obtain the primary network interface configuration, of vmss %s", pointer.StringDeref(vmss.Name, ""))
				return true
			}
			vmssNIC := *vmss.VirtualMachineProfile.NetworkProfile.NetworkInterfaceConfigurations
			primaryNIC, err := getPrimaryNetworkInterfaceConfiguration(vmssNIC, pointer.StringDeref(vmss.Name, ""))
			if err != nil {
				klog.Errorf("ensureBackendPoolDeletedFromVMSS: failed to get the primary network interface config of the VMSS %s: %v", pointer.StringDeref(vmss.Name, ""), err)
				errorList = append(errorList, err)
				return true
			}

			handleBackendPool := func(backendPoolID string) bool {
				primaryIPConfig, err := getPrimaryIPConfigFromVMSSNetworkConfig(primaryNIC, backendPoolID, pointer.StringDeref(vmss.Name, ""))
				if err != nil {
					klog.Errorf("ensureBackendPoolDeletedFromVMSS: failed to find the primary IP config from the VMSS %s's network config : %v", pointer.StringDeref(vmss.Name, ""), err)
					errorList = append(errorList, err)
					return true
				}
				loadBalancerBackendAddressPools := make([]compute.SubResource, 0)
				if primaryIPConfig.LoadBalancerBackendAddressPools != nil {
					loadBalancerBackendAddressPools = *primaryIPConfig.LoadBalancerBackendAddressPools
				}
				for _, loadBalancerBackendAddressPool := range loadBalancerBackendAddressPools {
					klog.V(4).Infof("ensureBackendPoolDeletedFromVMSS: loadBalancerBackendAddressPool (%s) on vmss (%s)", pointer.StringDeref(loadBalancerBackendAddressPool.ID, ""), pointer.StringDeref(vmss.Name, ""))
					if strings.EqualFold(pointer.StringDeref(loadBalancerBackendAddressPool.ID, ""), backendPoolID) {
						klog.V(4).Infof("ensureBackendPoolDeletedFromVMSS: found vmss %s with backend pool %s, removing it", pointer.StringDeref(vmss.Name, ""), backendPoolID)
						vmssNamesMap[pointer.StringDeref(vmss.Name, "")] = true
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
		klog.V(2).Infof("ensureBackendPoolDeletedFromVmssUniform: vmss %q, backendPoolIDs %q", vmSetName, backendPoolIDs)
		vmssNamesMap[vmSetName] = true
	}

	return ss.EnsureBackendPoolDeletedFromVMSets(vmssNamesMap, backendPoolIDs)
}

// ensureBackendPoolDeleted ensures the loadBalancer backendAddressPools deleted from the specified nodes.
func (ss *ScaleSet) ensureBackendPoolDeleted(service *v1.Service, backendPoolIDs []string, vmSetName string, backendAddressPools *[]network.BackendAddressPool) (bool, error) {
	// Returns nil if backend address pools already deleted.
	if backendAddressPools == nil {
		return false, nil
	}

	mc := metrics.NewMetricContext("services", "vmss_ensure_backend_pool_deleted", ss.ResourceGroup, ss.SubscriptionID, getServiceName(service))
	isOperationSucceeded := false
	defer func() {
		mc.ObserveOperationWithResult(isOperationSucceeded)
	}()

	ipConfigurationIDs := []string{}
	for _, backendPool := range *backendAddressPools {
		for _, backendPoolID := range backendPoolIDs {
			if strings.EqualFold(*backendPool.ID, backendPoolID) && backendPool.BackendIPConfigurations != nil {
				for _, ipConf := range *backendPool.BackendIPConfigurations {
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
	nodeUpdates := make(map[vmssMetaInfo]map[string]compute.VirtualMachineScaleSetVM)
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
			if !ss.useStandardLoadBalancer() && !strings.EqualFold(scaleSetName, vmSetName) {
				continue
			}
		}

		nodeName, _, err := ss.GetNodeNameByIPConfigurationID(ipConfigurationID)
		if err != nil {
			if errors.Is(err, ErrorNotVmssInstance) { // Do nothing for the VMAS nodes.
				continue
			}

			if errors.Is(err, cloudprovider.InstanceNotFound) {
				klog.Infof("ensureBackendPoolDeleted(%s): skipping ip config %s because the corresponding vmss vm is not"+
					" found", getServiceName(service), ipConfigurationID)
				continue
			}

			klog.Errorf("Failed to GetNodeNameByIPConfigurationID(%s): %v", ipConfigurationID, err)
			allErrs = append(allErrs, err)
			continue
		}

		nodeResourceGroup, nodeVMSS, nodeInstanceID, nodeVMSSVM, err := ss.ensureBackendPoolDeletedFromNode(nodeName, backendPoolIDs)
		if err != nil {
			if !errors.Is(err, ErrorNotVmssInstance) { // Do nothing for the VMAS nodes.
				klog.Errorf("ensureBackendPoolDeleted(%s): backendPoolIDs(%q) - failed with error %v", getServiceName(service), backendPoolIDs, err)
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
			nodeUpdates[nodeVMSSMetaInfo] = map[string]compute.VirtualMachineScaleSetVM{
				nodeInstanceID: *nodeVMSSVM,
			}
		}

		// Invalidate the cache since the VMSS VM would be updated.
		defer func() {
			_ = ss.DeleteCacheForNode(nodeName)
		}()
	}

	// Update VMs with best effort that have already been added to nodeUpdates.
	var updatedVM atomic.Bool
	for meta, update := range nodeUpdates {
		// create new instance of meta and update for passing to anonymous function
		meta := meta
		update := update
		hostUpdates = append(hostUpdates, func() error {
			ctx, cancel := getContextWithCancel()
			defer cancel()

			logFields := []interface{}{
				"operation", "EnsureBackendPoolDeleted UpdateVMSSVMs",
				"vmssName", meta.vmssName,
				"resourceGroup", meta.resourceGroup,
				"backendPoolIDs", backendPoolIDs,
			}

			batchSize, err := ss.VMSSBatchSize(meta.vmssName)
			if err != nil {
				klog.ErrorS(err, "Failed to get vmss batch size", logFields...)
				return err
			}

			klog.V(2).InfoS("Begin to update VMs for VMSS with new backendPoolID", logFields...)
			rerr := ss.VirtualMachineScaleSetVMsClient.UpdateVMs(ctx, meta.resourceGroup, meta.vmssName, update, "network_update", batchSize)
			if rerr != nil {
				klog.ErrorS(err, "Failed to update VMs for VMSS", logFields...)
				return rerr.Error()
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
func (ss *ScaleSet) EnsureBackendPoolDeleted(service *v1.Service, backendPoolIDs []string, vmSetName string, backendAddressPools *[]network.BackendAddressPool, deleteFromVMSet bool) (bool, error) {
	if backendAddressPools == nil {
		return false, nil
	}
	vmssUniformBackendIPConfigurationsMap := map[string][]network.InterfaceIPConfiguration{}
	vmssFlexBackendIPConfigurationsMap := map[string][]network.InterfaceIPConfiguration{}
	avSetBackendIPConfigurationsMap := map[string][]network.InterfaceIPConfiguration{}

	for _, backendPool := range *backendAddressPools {
		for _, backendPoolID := range backendPoolIDs {
			if strings.EqualFold(*backendPool.ID, backendPoolID) && backendPool.BackendIPConfigurations != nil {
				for _, ipConf := range *backendPool.BackendIPConfigurations {
					if ipConf.ID == nil {
						continue
					}

					vmManagementType, err := ss.getVMManagementTypeByIPConfigurationID(*ipConf.ID, azcache.CacheReadTypeUnsafe)
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
		err := ss.ensureBackendPoolDeletedFromVMSS(backendPoolIDs, vmSetName)
		if err != nil {
			return false, err
		}
	}

	var updated bool
	vmssUniformBackendPools := []network.BackendAddressPool{}
	for backendPoolID, vmssUniformBackendIPConfigurations := range vmssUniformBackendIPConfigurationsMap {
		vmssUniformBackendIPConfigurations := vmssUniformBackendIPConfigurations
		vmssUniformBackendPools = append(vmssUniformBackendPools, network.BackendAddressPool{
			ID: pointer.String(backendPoolID),
			BackendAddressPoolPropertiesFormat: &network.BackendAddressPoolPropertiesFormat{
				BackendIPConfigurations: &vmssUniformBackendIPConfigurations,
			},
		})
	}
	if len(vmssUniformBackendPools) > 0 {
		updatedVM, err := ss.ensureBackendPoolDeleted(service, backendPoolIDs, vmSetName, &vmssUniformBackendPools)
		if err != nil {
			return false, err
		}
		if updatedVM {
			updated = true
		}
	}

	vmssFlexBackendPools := []network.BackendAddressPool{}
	for backendPoolID, vmssFlexBackendIPConfigurations := range vmssFlexBackendIPConfigurationsMap {
		vmssFlexBackendIPConfigurations := vmssFlexBackendIPConfigurations
		vmssFlexBackendPools = append(vmssFlexBackendPools, network.BackendAddressPool{
			ID: pointer.String(backendPoolID),
			BackendAddressPoolPropertiesFormat: &network.BackendAddressPoolPropertiesFormat{
				BackendIPConfigurations: &vmssFlexBackendIPConfigurations,
			},
		})
	}
	if len(vmssFlexBackendPools) > 0 {
		updatedNIC, err := ss.flexScaleSet.EnsureBackendPoolDeleted(service, backendPoolIDs, vmSetName, &vmssFlexBackendPools, false)
		if err != nil {
			return false, err
		}
		if updatedNIC {
			updated = true
		}
	}

	avSetBackendPools := []network.BackendAddressPool{}
	for backendPoolID, avSetBackendIPConfigurations := range avSetBackendIPConfigurationsMap {
		avSetBackendIPConfigurations := avSetBackendIPConfigurations
		avSetBackendPools = append(avSetBackendPools, network.BackendAddressPool{
			ID: pointer.String(backendPoolID),
			BackendAddressPoolPropertiesFormat: &network.BackendAddressPoolPropertiesFormat{
				BackendIPConfigurations: &avSetBackendIPConfigurations,
			},
		})
	}
	if len(avSetBackendPools) > 0 {
		updatedNIC, err := ss.availabilitySet.EnsureBackendPoolDeleted(service, backendPoolIDs, vmSetName, &avSetBackendPools, false)
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
func (ss *ScaleSet) GetNodeCIDRMasksByProviderID(providerID string) (int, int, error) {
	vmManagementType, err := ss.getVMManagementTypeByProviderID(providerID, azcache.CacheReadTypeUnsafe)
	if err != nil {
		klog.Errorf("Failed to check VM management type: %v", err)
		return 0, 0, err
	}

	if vmManagementType == ManagedByAvSet {
		// vm is managed by availability set.
		return ss.availabilitySet.GetNodeCIDRMasksByProviderID(providerID)
	}
	if vmManagementType == ManagedByVmssFlex {
		// vm is managed by vmss flex.
		return ss.flexScaleSet.GetNodeCIDRMasksByProviderID(providerID)
	}

	_, vmssName, err := getVmssAndResourceGroupNameByVMProviderID(providerID)
	if err != nil {
		return 0, 0, err
	}

	vmss, err := ss.getVMSS(vmssName, azcache.CacheReadTypeDefault)
	if err != nil {
		return 0, 0, err
	}

	var ipv4Mask, ipv6Mask int
	if v4, ok := vmss.Tags[consts.VMSetCIDRIPV4TagKey]; ok && v4 != nil {
		ipv4Mask, err = strconv.Atoi(pointer.StringDeref(v4, ""))
		if err != nil {
			klog.Errorf("GetNodeCIDRMasksByProviderID: error when paring the value of the ipv4 mask size %s: %v", pointer.StringDeref(v4, ""), err)
		}
	}
	if v6, ok := vmss.Tags[consts.VMSetCIDRIPV6TagKey]; ok && v6 != nil {
		ipv6Mask, err = strconv.Atoi(pointer.StringDeref(v6, ""))
		if err != nil {
			klog.Errorf("GetNodeCIDRMasksByProviderID: error when paring the value of the ipv6 mask size%s: %v", pointer.StringDeref(v6, ""), err)
		}
	}

	return ipv4Mask, ipv6Mask, nil
}

// deleteBackendPoolFromIPConfig deletes the backend pool from the IP config.
func deleteBackendPoolFromIPConfig(msg, backendPoolID, resource string, primaryNIC *compute.VirtualMachineScaleSetNetworkConfiguration) (bool, error) {
	primaryIPConfig, err := getPrimaryIPConfigFromVMSSNetworkConfig(primaryNIC, backendPoolID, resource)
	if err != nil {
		klog.Errorf("%s: failed to get the primary IP config from the VMSS %q's network config: %v", msg, resource, err)
		return false, err
	}
	loadBalancerBackendAddressPools := []compute.SubResource{}
	if primaryIPConfig.LoadBalancerBackendAddressPools != nil {
		loadBalancerBackendAddressPools = *primaryIPConfig.LoadBalancerBackendAddressPools
	}

	var found bool
	var newBackendPools []compute.SubResource
	for i := len(loadBalancerBackendAddressPools) - 1; i >= 0; i-- {
		curPool := loadBalancerBackendAddressPools[i]
		if strings.EqualFold(backendPoolID, *curPool.ID) {
			klog.V(10).Infof("%s gets unwanted backend pool %q for VMSS OR VMSS VM %q", msg, backendPoolID, resource)
			found = true
			newBackendPools = append(loadBalancerBackendAddressPools[:i], loadBalancerBackendAddressPools[i+1:]...)
		}
	}
	if !found {
		return false, nil
	}
	primaryIPConfig.LoadBalancerBackendAddressPools = &newBackendPools
	return true, nil
}

// EnsureBackendPoolDeletedFromVMSets ensures the loadBalancer backendAddressPools deleted from the specified VMSS
func (ss *ScaleSet) EnsureBackendPoolDeletedFromVMSets(vmssNamesMap map[string]bool, backendPoolIDs []string) error {
	vmssUpdaters := make([]func() error, 0, len(vmssNamesMap))
	errors := make([]error, 0, len(vmssNamesMap))
	for vmssName := range vmssNamesMap {
		vmssName := vmssName
		vmss, err := ss.getVMSS(vmssName, azcache.CacheReadTypeDefault)
		if err != nil {
			klog.Errorf("ensureBackendPoolDeletedFromVMSS: failed to get VMSS %s: %v", vmssName, err)
			errors = append(errors, err)
			continue
		}

		// When vmss is being deleted, CreateOrUpdate API would report "the vmss is being deleted" error.
		// Since it is being deleted, we shouldn't send more CreateOrUpdate requests for it.
		if vmss.ProvisioningState != nil && strings.EqualFold(*vmss.ProvisioningState, consts.ProvisionStateDeleting) {
			klog.V(3).Infof("EnsureBackendPoolDeletedFromVMSets: found vmss %s being deleted, skipping", vmssName)
			continue
		}
		if vmss.VirtualMachineProfile.NetworkProfile.NetworkInterfaceConfigurations == nil {
			klog.V(4).Infof("EnsureBackendPoolDeletedFromVMSets: cannot obtain the primary network interface configuration, of vmss %s", vmssName)
			continue
		}
		vmssNIC := *vmss.VirtualMachineProfile.NetworkProfile.NetworkInterfaceConfigurations
		primaryNIC, err := getPrimaryNetworkInterfaceConfiguration(vmssNIC, vmssName)
		if err != nil {
			klog.Errorf("EnsureBackendPoolDeletedFromVMSets: failed to get the primary network interface config of the VMSS %s: %v", vmssName, err)
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
			newVMSS := compute.VirtualMachineScaleSet{
				Location: vmss.Location,
				VirtualMachineScaleSetProperties: &compute.VirtualMachineScaleSetProperties{
					VirtualMachineProfile: &compute.VirtualMachineScaleSetVMProfile{
						NetworkProfile: &compute.VirtualMachineScaleSetNetworkProfile{
							NetworkInterfaceConfigurations: &vmssNIC,
						},
					},
				},
			}

			klog.V(2).Infof("EnsureBackendPoolDeletedFromVMSets begins to update vmss(%s) with backendPoolIDs %q", vmssName, backendPoolIDs)
			rerr := ss.CreateOrUpdateVMSS(ss.ResourceGroup, vmssName, newVMSS)
			if rerr != nil {
				klog.Errorf("EnsureBackendPoolDeletedFromVMSets CreateOrUpdateVMSS(%s) with new backendPoolIDs %q, err: %v", vmssName, backendPoolIDs, rerr)
				return rerr.Error()
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
func (ss *ScaleSet) GetAgentPoolVMSetNames(nodes []*v1.Node) (*[]string, error) {
	vmSetNames := make([]string, 0)

	vmssFlexVMNodes := make([]*v1.Node, 0)
	avSetVMNodes := make([]*v1.Node, 0)

	for _, node := range nodes {
		var names *[]string

		vmManagementType, err := ss.getVMManagementTypeByNodeName(node.Name, azcache.CacheReadTypeDefault)
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

		names, err = ss.getAgentPoolScaleSets([]*v1.Node{node})
		if err != nil {
			return nil, fmt.Errorf("GetAgentPoolVMSetNames: failed to execute getAgentPoolScaleSets: %w", err)
		}
		vmSetNames = append(vmSetNames, *names...)
	}

	if len(vmssFlexVMNodes) > 0 {
		vmssFlexVMnames, err := ss.flexScaleSet.GetAgentPoolVMSetNames(vmssFlexVMNodes)
		if err != nil {
			return nil, fmt.Errorf("ss.flexScaleSet.GetAgentPoolVMSetNames: failed to execute : %w", err)
		}
		vmSetNames = append(vmSetNames, *vmssFlexVMnames...)
	}

	if len(avSetVMNodes) > 0 {
		avSetVMnames, err := ss.availabilitySet.GetAgentPoolVMSetNames(avSetVMNodes)
		if err != nil {
			return nil, fmt.Errorf("ss.availabilitySet.GetAgentPoolVMSetNames: failed to execute : %w", err)
		}
		vmSetNames = append(vmSetNames, *avSetVMnames...)
	}

	return &vmSetNames, nil
}

func (ss *ScaleSet) GetNodeVMSetName(node *v1.Node) (string, error) {
	vmManagementType, err := ss.getVMManagementTypeByNodeName(node.Name, azcache.CacheReadTypeUnsafe)
	if err != nil {
		klog.Errorf("Failed to check VM management type: %v", err)
		return "", err
	}

	if vmManagementType == ManagedByAvSet {
		// vm is managed by availability set.
		return ss.availabilitySet.GetNodeVMSetName(node)
	}
	if vmManagementType == ManagedByVmssFlex {
		// vm is managed by vmss flex.
		return ss.flexScaleSet.GetNodeVMSetName(node)
	}

	providerID := node.Spec.ProviderID
	_, vmssName, err := getVmssAndResourceGroupNameByVMProviderID(providerID)
	if err != nil {
		klog.Errorf("getVmssAndResourceGroupNameByVMProviderID failed: %v", err)
		return "", err
	}

	klog.V(4).Infof("ss.GetNodeVMSetName: found vmss name %s from node name %s", vmssName, node.Name)
	return vmssName, nil
}

// VMSSBatchSize returns the batch size for VMSS operations.
func (ss *ScaleSet) VMSSBatchSize(vmssName string) (int, error) {
	batchSize := 0
	vmss, err := ss.getVMSS(vmssName, azcache.CacheReadTypeDefault)
	if err != nil {
		return 0, fmt.Errorf("get vmss batch size: %w", err)
	}
	if _, ok := vmss.Tags[consts.VMSSTagForBatchOperation]; ok {
		batchSize = ss.getPutVMSSVMBatchSize()
	}
	klog.V(2).InfoS("Fetch VMSS batch size", "vmss", vmssName, "size", batchSize)
	return batchSize, nil
}
