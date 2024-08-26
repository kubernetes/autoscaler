/*
Copyright 2022 The Kubernetes Authors.

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
	vmutil "sigs.k8s.io/cloud-provider-azure/pkg/util/vm"
)

var (
	// ErrorVmssIDIsEmpty indicates the vmss id is empty.
	ErrorVmssIDIsEmpty = errors.New("VMSS ID is empty")
)

// FlexScaleSet implements VMSet interface for Azure Flexible VMSS.
type FlexScaleSet struct {
	*Cloud

	vmssFlexCache            azcache.Resource
	vmssFlexVMNameToVmssID   *sync.Map
	vmssFlexVMNameToNodeName *sync.Map
	vmssFlexVMCache          azcache.Resource

	// lockMap in cache refresh
	lockMap *LockMap
}

func newFlexScaleSet(ctx context.Context, az *Cloud) (VMSet, error) {
	fs := &FlexScaleSet{
		Cloud:                    az,
		vmssFlexVMNameToVmssID:   &sync.Map{},
		vmssFlexVMNameToNodeName: &sync.Map{},
		lockMap:                  newLockMap(),
	}

	var err error
	fs.vmssFlexCache, err = fs.newVmssFlexCache(ctx)
	if err != nil {
		return nil, err
	}
	fs.vmssFlexVMCache, err = fs.newVmssFlexVMCache(ctx)
	if err != nil {
		return nil, err
	}

	return fs, nil
}

// GetPrimaryVMSetName returns the VM set name depending on the configured vmType.
// It returns config.PrimaryScaleSetName for vmss and config.PrimaryAvailabilitySetName for standard vmType.
func (fs *FlexScaleSet) GetPrimaryVMSetName() string {
	return fs.Config.PrimaryScaleSetName
}

// getNodeVMSetName returns the vmss flex name by the node name.
func (fs *FlexScaleSet) getNodeVmssFlexName(nodeName string) (string, error) {
	vmssFlexID, err := fs.getNodeVmssFlexID(nodeName)
	if err != nil {
		return "", err
	}
	vmssFlexName, err := getLastSegment(vmssFlexID, "/")
	if err != nil {
		return "", err
	}
	return vmssFlexName, nil

}

// GetNodeVMSetName returns the availability set or vmss name by the node name.
// It will return empty string when using standalone vms.
func (fs *FlexScaleSet) GetNodeVMSetName(node *v1.Node) (string, error) {
	return fs.getNodeVmssFlexName(node.Name)
}

// GetAgentPoolVMSetNames returns all vmSet names according to the nodes
func (fs *FlexScaleSet) GetAgentPoolVMSetNames(nodes []*v1.Node) (*[]string, error) {
	vmSetNames := make([]string, 0)
	for _, node := range nodes {
		vmSetName, err := fs.GetNodeVMSetName(node)
		if err != nil {
			klog.Errorf("Unable to get the vmss flex name by node name %s: %v", node.Name, err)
			continue
		}
		vmSetNames = append(vmSetNames, vmSetName)
	}
	return &vmSetNames, nil
}

// GetVMSetNames selects all possible availability sets or scale sets
// (depending vmType configured) for service load balancer, if the service has
// no loadbalancer mode annotation returns the primary VMSet. If service annotation
// for loadbalancer exists then returns the eligible VMSet. The mode selection
// annotation would be ignored when using one SLB per cluster.
func (fs *FlexScaleSet) GetVMSetNames(service *v1.Service, nodes []*v1.Node) (*[]string, error) {
	hasMode, isAuto, serviceVMSetName := fs.getServiceLoadBalancerMode(service)
	if !hasMode || fs.useStandardLoadBalancer() {
		// no mode specified in service annotation or use single SLB mode
		// default to PrimaryScaleSetName
		vmssFlexNames := &[]string{fs.Config.PrimaryScaleSetName}
		return vmssFlexNames, nil
	}

	vmssFlexNames, err := fs.GetAgentPoolVMSetNames(nodes)
	if err != nil {
		klog.Errorf("fs.GetVMSetNames - GetAgentPoolVMSetNames failed err=(%v)", err)
		return nil, err
	}

	if !isAuto {
		found := false
		for asx := range *vmssFlexNames {
			if strings.EqualFold((*vmssFlexNames)[asx], serviceVMSetName) {
				found = true
				serviceVMSetName = (*vmssFlexNames)[asx]
				break
			}
		}
		if !found {
			klog.Errorf("fs.GetVMSetNames - scale set (%s) in service annotation not found", serviceVMSetName)
			return nil, fmt.Errorf("scale set (%s) - not found", serviceVMSetName)
		}
		return &[]string{serviceVMSetName}, nil
	}
	return vmssFlexNames, nil
}

// GetNodeNameByProviderID gets the node name by provider ID.
// providerID example:
// azure:///subscriptions/sub/resourceGroups/rg/providers/Microsoft.Compute/virtualMachines/flexprofile-mp-0_df53ee36
// Different from vmas where vm name is always equal to nodeName, we need to further map vmName to actual nodeName in vmssflex.
// Note: nodeName is always equal pointer.StringDerefs.ToLower(*vm.OsProfile.ComputerName, "")
func (fs *FlexScaleSet) GetNodeNameByProviderID(providerID string) (types.NodeName, error) {
	// NodeName is part of providerID for standard instances.
	matches := providerIDRE.FindStringSubmatch(providerID)
	if len(matches) != 2 {
		return "", errors.New("error splitting providerID")
	}

	nodeName, err := fs.getNodeNameByVMName(matches[1])
	if err != nil {
		return "", err
	}
	return types.NodeName(nodeName), nil
}

// GetInstanceIDByNodeName gets the cloud provider ID by node name.
// It must return ("", cloudprovider.InstanceNotFound) if the instance does
// not exist or is no longer running.
func (fs *FlexScaleSet) GetInstanceIDByNodeName(name string) (string, error) {
	machine, err := fs.getVmssFlexVM(name, azcache.CacheReadTypeUnsafe)
	if err != nil {
		return "", err
	}
	if machine.ID == nil {
		return "", fmt.Errorf("ProviderID of node(%s) is nil", name)
	}
	resourceID := *machine.ID
	convertedResourceID, err := ConvertResourceGroupNameToLower(resourceID)
	if err != nil {
		klog.Errorf("ConvertResourceGroupNameToLower failed with error: %v", err)
		return "", err
	}
	return convertedResourceID, nil

}

// GetInstanceTypeByNodeName gets the instance type by node name.
func (fs *FlexScaleSet) GetInstanceTypeByNodeName(name string) (string, error) {
	machine, err := fs.getVmssFlexVM(name, azcache.CacheReadTypeUnsafe)
	if err != nil {
		klog.Errorf("fs.GetInstanceTypeByNodeName(%s) failed: fs.getVmssFlexVMWithoutInstanceView(%s) err=%v", name, name, err)
		return "", err
	}

	if machine.HardwareProfile == nil {
		return "", fmt.Errorf("HardwareProfile of node(%s) is nil", name)
	}
	return string(machine.HardwareProfile.VMSize), nil
}

// GetZoneByNodeName gets availability zone for the specified node. If the node is not running
// with availability zone, then it returns fault domain.
// for details, refer to https://kubernetes-sigs.github.io/cloud-provider-azure/topics/availability-zones/#node-labels
func (fs *FlexScaleSet) GetZoneByNodeName(name string) (cloudprovider.Zone, error) {
	vm, err := fs.getVmssFlexVM(name, azcache.CacheReadTypeUnsafe)
	if err != nil {
		klog.Errorf("fs.GetZoneByNodeName(%s) failed: fs.getVmssFlexVMWithoutInstanceView(%s) err=%v", name, name, err)
		return cloudprovider.Zone{}, err
	}

	var failureDomain string
	if vm.Zones != nil && len(*vm.Zones) > 0 {
		// Get availability zone for the node.
		zones := *vm.Zones
		zoneID, err := strconv.Atoi(zones[0])
		if err != nil {
			return cloudprovider.Zone{}, fmt.Errorf("failed to parse zone %q: %w", zones, err)
		}

		failureDomain = fs.makeZone(pointer.StringDeref(vm.Location, ""), zoneID)
	} else if vm.VirtualMachineProperties.InstanceView != nil && vm.VirtualMachineProperties.InstanceView.PlatformFaultDomain != nil {
		// Availability zone is not used for the node, falling back to fault domain.
		failureDomain = strconv.Itoa(int(pointer.Int32Deref(vm.VirtualMachineProperties.InstanceView.PlatformFaultDomain, 0)))
	} else {
		err = fmt.Errorf("failed to get zone info")
		klog.Errorf("GetZoneByNodeName: got unexpected error %v", err)
		return cloudprovider.Zone{}, err
	}

	zone := cloudprovider.Zone{
		FailureDomain: strings.ToLower(failureDomain),
		Region:        strings.ToLower(pointer.StringDeref(vm.Location, "")),
	}
	return zone, nil
}

// GetProvisioningStateByNodeName returns the provisioningState for the specified node.
func (fs *FlexScaleSet) GetProvisioningStateByNodeName(name string) (provisioningState string, err error) {
	vm, err := fs.getVmssFlexVM(name, azcache.CacheReadTypeDefault)
	if err != nil {
		return provisioningState, err
	}

	if vm.VirtualMachineProperties == nil || vm.VirtualMachineProperties.ProvisioningState == nil {
		return provisioningState, nil
	}

	return pointer.StringDeref(vm.VirtualMachineProperties.ProvisioningState, ""), nil
}

// GetPowerStatusByNodeName returns the powerState for the specified node.
func (fs *FlexScaleSet) GetPowerStatusByNodeName(name string) (powerState string, err error) {
	vm, err := fs.getVmssFlexVM(name, azcache.CacheReadTypeDefault)
	if err != nil {
		return powerState, err
	}

	if vm.InstanceView != nil {
		return vmutil.GetVMPowerState(ptr.Deref(vm.Name, ""), vm.InstanceView.Statuses), nil
	}

	// vm.InstanceView or vm.InstanceView.Statuses are nil when the VM is under deleting.
	klog.V(3).Infof("InstanceView for node %q is nil, assuming it's deleting", name)
	return consts.VMPowerStateUnknown, nil
}

// GetPrimaryInterface gets machine primary network interface by node name.
func (fs *FlexScaleSet) GetPrimaryInterface(nodeName string) (network.Interface, error) {
	machine, err := fs.getVmssFlexVM(nodeName, azcache.CacheReadTypeDefault)
	if err != nil {
		klog.Errorf("fs.GetInstanceTypeByNodeName(%s) failed: fs.getVmssFlexVMWithoutInstanceView(%s) err=%v", nodeName, nodeName, err)
		return network.Interface{}, err
	}

	primaryNicID, err := getPrimaryInterfaceID(machine)
	if err != nil {
		return network.Interface{}, err
	}
	nicName, err := getLastSegment(primaryNicID, "/")
	if err != nil {
		return network.Interface{}, err
	}

	nicResourceGroup, err := extractResourceGroupByNicID(primaryNicID)
	if err != nil {
		return network.Interface{}, err
	}

	ctx, cancel := getContextWithCancel()
	defer cancel()
	nic, rerr := fs.InterfacesClient.Get(ctx, nicResourceGroup, nicName, "")
	if rerr != nil {
		return network.Interface{}, rerr.Error()
	}

	return nic, nil
}

// GetIPByNodeName gets machine private IP and public IP by node name.
func (fs *FlexScaleSet) GetIPByNodeName(name string) (string, string, error) {
	nic, err := fs.GetPrimaryInterface(name)
	if err != nil {
		return "", "", err
	}

	ipConfig, err := getPrimaryIPConfig(nic)
	if err != nil {
		klog.Errorf("fs.GetIPByNodeName(%s) failed: getPrimaryIPConfig(%v), err=%v", name, nic, err)
		return "", "", err
	}

	privateIP := *ipConfig.PrivateIPAddress
	publicIP := ""
	if ipConfig.PublicIPAddress != nil && ipConfig.PublicIPAddress.ID != nil {
		pipID := *ipConfig.PublicIPAddress.ID
		pipName, err := getLastSegment(pipID, "/")
		if err != nil {
			return "", "", fmt.Errorf("failed to publicIP name for node %q with pipID %q", name, pipID)
		}
		pip, existsPip, err := fs.getPublicIPAddress(fs.ResourceGroup, pipName, azcache.CacheReadTypeDefault)
		if err != nil {
			return "", "", err
		}
		if existsPip {
			publicIP = *pip.IPAddress
		}
	}

	return privateIP, publicIP, nil

}

// GetPrivateIPsByNodeName returns a slice of all private ips assigned to node (ipv6 and ipv4)
// TODO (khenidak): This should read all nics, not just the primary
// allowing users to split ipv4/v6 on multiple nics
func (fs *FlexScaleSet) GetPrivateIPsByNodeName(name string) ([]string, error) {
	ips := make([]string, 0)
	nic, err := fs.GetPrimaryInterface(name)
	if err != nil {
		return ips, err
	}

	if nic.IPConfigurations == nil {
		return ips, fmt.Errorf("nic.IPConfigurations for nic (nicname=%s) is nil", *nic.Name)
	}

	for _, ipConfig := range *(nic.IPConfigurations) {
		if ipConfig.PrivateIPAddress != nil {
			ips = append(ips, *(ipConfig.PrivateIPAddress))
		}
	}

	return ips, nil
}

// GetNodeNameByIPConfigurationID gets the nodeName and vmSetName by IP configuration ID.
func (fs *FlexScaleSet) GetNodeNameByIPConfigurationID(ipConfigurationID string) (string, string, error) {
	nodeName, vmssFlexName, _, err := fs.getNodeInformationByIPConfigurationID(ipConfigurationID)
	if err != nil {
		klog.Errorf("fs.GetNodeNameByIPConfigurationID(%s) failed. Error: %v", ipConfigurationID, err)
		return "", "", err
	}

	return nodeName, strings.ToLower(vmssFlexName), nil
}

func (fs *FlexScaleSet) getNodeInformationByIPConfigurationID(ipConfigurationID string) (string, string, string, error) {
	nicResourceGroup, nicName, err := getResourceGroupAndNameFromNICID(ipConfigurationID)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to get resource group and name from ip config ID %s: %w", ipConfigurationID, err)
	}

	// get vmName by nic name
	vmName, err := fs.GetVMNameByIPConfigurationName(nicResourceGroup, nicName)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to get vm name of ip config ID %s: %w", ipConfigurationID, err)
	}

	nodeName, err := fs.getNodeNameByVMName(vmName)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to map VM Name to NodeName: VM Name %s: %w", vmName, err)
	}

	vmssFlexName, err := fs.getNodeVmssFlexName(nodeName)

	if err != nil {
		klog.Errorf("Unable to get the vmss flex name by node name %s: %v", vmName, err)
		return "", "", "", err
	}

	return nodeName, strings.ToLower(vmssFlexName), nicName, nil
}

// GetNodeCIDRMaskByProviderID returns the node CIDR subnet mask by provider ID.
func (fs *FlexScaleSet) GetNodeCIDRMasksByProviderID(providerID string) (int, int, error) {
	nodeNameWrapper, err := fs.GetNodeNameByProviderID(providerID)
	if err != nil {
		klog.Errorf("Unable to get the vmss flex vm node name by providerID %s: %v", providerID, err)
		return 0, 0, err
	}
	nodeName := mapNodeNameToVMName(nodeNameWrapper)

	vmssFlex, err := fs.getVmssFlexByNodeName(nodeName, azcache.CacheReadTypeDefault)
	if err != nil {
		if errors.Is(err, cloudprovider.InstanceNotFound) {
			return consts.DefaultNodeMaskCIDRIPv4, consts.DefaultNodeMaskCIDRIPv6, nil
		}
		return 0, 0, err
	}

	var ipv4Mask, ipv6Mask int
	if v4, ok := vmssFlex.Tags[consts.VMSetCIDRIPV4TagKey]; ok && v4 != nil {
		ipv4Mask, err = strconv.Atoi(pointer.StringDeref(v4, ""))
		if err != nil {
			klog.Errorf("GetNodeCIDRMasksByProviderID: error when paring the value of the ipv4 mask size %s: %v", pointer.StringDeref(v4, ""), err)
		}
	}
	if v6, ok := vmssFlex.Tags[consts.VMSetCIDRIPV6TagKey]; ok && v6 != nil {
		ipv6Mask, err = strconv.Atoi(pointer.StringDeref(v6, ""))
		if err != nil {
			klog.Errorf("GetNodeCIDRMasksByProviderID: error when paring the value of the ipv6 mask size%s: %v", pointer.StringDeref(v6, ""), err)
		}
	}

	return ipv4Mask, ipv6Mask, nil
}

// EnsureHostInPool ensures the given VM's Primary NIC's Primary IP Configuration is
// participating in the specified LoadBalancer Backend Pool, which returns (resourceGroup, vmasName, instanceID, vmssVM, error).
func (fs *FlexScaleSet) EnsureHostInPool(service *v1.Service, nodeName types.NodeName, backendPoolID string, vmSetNameOfLB string) (string, string, string, *compute.VirtualMachineScaleSetVM, error) {
	serviceName := getServiceName(service)
	name := mapNodeNameToVMName(nodeName)
	vmssFlexName, err := fs.getNodeVmssFlexName(name)
	if err != nil {
		klog.Errorf("EnsureHostInPool: failed to get VMSS Flex Name %s: %v", name, err)
		return "", "", "", nil, nil
	}

	// Check scale set name:
	// - For basic SKU load balancer, return error as VMSS Flex does not support basic load balancer.
	// - For single standard SKU load balancer, backend could belong to multiple VMSS, so we
	//   don't check vmSet for it.
	// - For multiple standard SKU load balancers, return nil if the node's scale set is mismatched with vmSetNameOfLB
	needCheck := false
	if !fs.useStandardLoadBalancer() {
		return "", "", "", nil, fmt.Errorf("EnsureHostInPool: VMSS Flex does not support Basic Load Balancer")
	}
	if vmSetNameOfLB != "" && needCheck && !strings.EqualFold(vmSetNameOfLB, vmssFlexName) {
		klog.V(3).Infof("EnsureHostInPool skips node %s because it is not in the ScaleSet %s", name, vmSetNameOfLB)
		return "", "", "", nil, errNotInVMSet
	}

	nic, err := fs.GetPrimaryInterface(name)
	if err != nil {
		klog.Errorf("error: fs.EnsureHostInPool(%s), s.GetPrimaryInterface(%s), vmSetNameOfLB: %s, err=%v", name, name, vmSetNameOfLB, err)
		return "", "", "", nil, err
	}

	if nic.ProvisioningState == consts.NicFailedState {
		klog.Warningf("EnsureHostInPool skips node %s because its primary nic %s is in Failed state", nodeName, *nic.Name)
		return "", "", "", nil, nil
	}

	var primaryIPConfig *network.InterfaceIPConfiguration
	ipv6 := isBackendPoolIPv6(backendPoolID)
	if !fs.Cloud.ipv6DualStackEnabled && !ipv6 {
		primaryIPConfig, err = getPrimaryIPConfig(nic)
		if err != nil {
			return "", "", "", nil, err
		}
	} else {
		primaryIPConfig, err = getIPConfigByIPFamily(nic, ipv6)
		if err != nil {
			return "", "", "", nil, err
		}
	}

	foundPool := false
	newBackendPools := []network.BackendAddressPool{}
	if primaryIPConfig.LoadBalancerBackendAddressPools != nil {
		newBackendPools = *primaryIPConfig.LoadBalancerBackendAddressPools
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

	if fs.useStandardLoadBalancer() && len(newBackendPools) > 0 {
		// Although standard load balancer supports backends from multiple availability
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
			klog.V(4).Infof("Node %q has already been added to LB %q, omit adding it to a new one", nodeName, oldLBName)
			return "", "", "", nil, nil
		}
	}

	newBackendPools = append(newBackendPools,
		network.BackendAddressPool{
			ID: pointer.String(backendPoolID),
		})

	primaryIPConfig.LoadBalancerBackendAddressPools = &newBackendPools

	nicName := *nic.Name
	klog.V(3).Infof("nicupdate(%s): nic(%s) - updating", serviceName, nicName)
	err = fs.CreateOrUpdateInterface(service, nic)
	if err != nil {
		return "", "", "", nil, err
	}

	// Get the node resource group.
	nodeResourceGroup, err := fs.GetNodeResourceGroup(name)
	if err != nil {
		return "", "", "", nil, err
	}

	return nodeResourceGroup, vmssFlexName, name, nil, nil

}

func (fs *FlexScaleSet) ensureVMSSFlexInPool(_ *v1.Service, nodes []*v1.Node, backendPoolID string, vmSetNameOfLB string) error {
	klog.V(2).Infof("ensureVMSSFlexInPool: ensuring VMSS Flex with backendPoolID %s", backendPoolID)
	vmssFlexIDsMap := make(map[string]bool)

	if !fs.useStandardLoadBalancer() {
		return fmt.Errorf("ensureVMSSFlexInPool: VMSS Flex does not support Basic Load Balancer")
	}

	// the single standard load balancer supports multiple vmss in its backend while
	// multiple standard load balancers doesn't
	if fs.useStandardLoadBalancer() {
		for _, node := range nodes {
			if fs.excludeMasterNodesFromStandardLB() && isControlPlaneNode(node) {
				continue
			}

			shouldExcludeLoadBalancer, err := fs.ShouldNodeExcludedFromLoadBalancer(node.Name)
			if err != nil {
				klog.Errorf("ShouldNodeExcludedFromLoadBalancer(%s) failed with error: %v", node.Name, err)
				return err
			}
			if shouldExcludeLoadBalancer {
				klog.V(4).Infof("Excluding unmanaged/external-resource-group node %q", node.Name)
				continue
			}

			// in this scenario the vmSetName is an empty string and the name of vmss should be obtained from the provider IDs of nodes
			vmssFlexID, err := fs.getNodeVmssFlexID(node.Name)
			if err != nil {
				klog.Errorf("ensureVMSSFlexInPool: failed to get VMSS Flex ID of node: %s, will skip checking and continue", node.Name)
				continue
			}
			resourceGroupName, err := fs.GetNodeResourceGroup(node.Name)
			if err != nil {
				klog.Errorf("ensureVMSSFlexInPool: failed to get resource group of node: %s, will skip checking and continue", node.Name)
				continue
			}

			// only vmsses in the resource group same as it's in azure config are included
			if strings.EqualFold(resourceGroupName, fs.ResourceGroup) {
				vmssFlexIDsMap[vmssFlexID] = true
			}
		}
	} else {
		vmssFlexID, err := fs.getVmssFlexIDByName(vmSetNameOfLB)
		if err != nil {
			klog.Errorf("ensureVMSSFlexInPool: failed to get VMSS Flex ID of vmSet: %s", vmSetNameOfLB)
			return err
		}
		vmssFlexIDsMap[vmssFlexID] = true
	}

	klog.V(2).Infof("ensureVMSSFlexInPool begins to update VMSS list %v with backendPoolID %s", vmssFlexIDsMap, backendPoolID)
	for vmssFlexID := range vmssFlexIDsMap {
		vmssFlex, err := fs.getVmssFlexByVmssFlexID(vmssFlexID, azcache.CacheReadTypeDefault)
		if err != nil {
			return err
		}
		vmssFlexName := *vmssFlex.Name

		// When vmss is being deleted, CreateOrUpdate API would report "the vmss is being deleted" error.
		// Since it is being deleted, we shouldn't send more CreateOrUpdate requests for it.
		if vmssFlex.ProvisioningState != nil && strings.EqualFold(*vmssFlex.ProvisioningState, consts.ProvisionStateDeleting) {
			klog.V(3).Infof("ensureVMSSFlexInPool: found vmss %s being deleted, skipping", vmssFlexID)
			continue
		}

		if vmssFlex.VirtualMachineProfile == nil || vmssFlex.VirtualMachineProfile.NetworkProfile == nil || vmssFlex.VirtualMachineProfile.NetworkProfile.NetworkInterfaceConfigurations == nil {
			klog.V(4).Infof("ensureVMSSFlexInPool: cannot obtain the primary network interface configuration of vmss %s, just skip it as it might not have default vm profile", vmssFlexID)
			continue
		}
		vmssNIC := *vmssFlex.VirtualMachineProfile.NetworkProfile.NetworkInterfaceConfigurations
		primaryNIC, err := getPrimaryNetworkInterfaceConfiguration(vmssNIC, vmssFlexName)
		if err != nil {
			return err
		}
		primaryIPConfig, err := getPrimaryIPConfigFromVMSSNetworkConfig(primaryNIC, backendPoolID, vmssFlexName)
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

		if fs.useStandardLoadBalancer() && len(loadBalancerBackendAddressPools) > 0 {
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
				klog.V(4).Infof("VMSS %q has already been added to LB %q, omit adding it to a new one", vmssFlexID, oldLBName)
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
			Location: vmssFlex.Location,
			VirtualMachineScaleSetProperties: &compute.VirtualMachineScaleSetProperties{
				VirtualMachineProfile: &compute.VirtualMachineScaleSetVMProfile{
					NetworkProfile: &compute.VirtualMachineScaleSetNetworkProfile{
						NetworkInterfaceConfigurations: &vmssNIC,
						NetworkAPIVersion:              compute.TwoZeroTwoZeroHyphenMinusOneOneHyphenMinusZeroOne,
					},
				},
			},
		}

		defer func() {
			_ = fs.vmssFlexCache.Delete(consts.VmssFlexKey)
		}()

		klog.V(2).Infof("ensureVMSSFlexInPool begins to add vmss(%s) with new backendPoolID %s", vmssFlexName, backendPoolID)
		rerr := fs.CreateOrUpdateVMSS(fs.ResourceGroup, vmssFlexName, newVMSS)
		if rerr != nil {
			klog.Errorf("ensureVMSSFlexInPool CreateOrUpdateVMSS(%s) with new backendPoolID %s, err: %v", vmssFlexName, backendPoolID, err)
			return rerr.Error()
		}
	}
	return nil
}

// EnsureHostsInPool ensures the given Node's primary IP configurations are
// participating in the specified LoadBalancer Backend Pool.
func (fs *FlexScaleSet) EnsureHostsInPool(service *v1.Service, nodes []*v1.Node, backendPoolID string, vmSetNameOfLB string) error {
	mc := metrics.NewMetricContext("services", "vmssflex_ensure_hosts_in_pool", fs.ResourceGroup, fs.SubscriptionID, getServiceName(service))
	isOperationSucceeded := false
	defer func() {
		mc.ObserveOperationWithResult(isOperationSucceeded)
	}()
	hostUpdates := make([]func() error, 0, len(nodes))

	for _, node := range nodes {
		localNodeName := node.Name
		if fs.useStandardLoadBalancer() && fs.excludeMasterNodesFromStandardLB() && isControlPlaneNode(node) {
			klog.V(4).Infof("Excluding master node %q from load balancer backendpool %q", localNodeName, backendPoolID)
			continue
		}

		shouldExcludeLoadBalancer, err := fs.ShouldNodeExcludedFromLoadBalancer(localNodeName)
		if err != nil {
			klog.Errorf("ShouldNodeExcludedFromLoadBalancer(%s) failed with error: %v", localNodeName, err)
			return err
		}
		if shouldExcludeLoadBalancer {
			klog.V(4).Infof("Excluding unmanaged/external-resource-group node %q", localNodeName)
			continue
		}

		f := func() error {
			_, _, _, _, err := fs.EnsureHostInPool(service, types.NodeName(localNodeName), backendPoolID, vmSetNameOfLB)
			if err != nil {
				return fmt.Errorf("ensure(%s): backendPoolID(%s) - failed to ensure host in pool: %w", getServiceName(service), backendPoolID, err)
			}
			return nil
		}
		hostUpdates = append(hostUpdates, f)
	}

	errs := utilerrors.AggregateGoroutines(hostUpdates...)
	if errs != nil {
		return utilerrors.Flatten(errs)
	}

	err := fs.ensureVMSSFlexInPool(service, nodes, backendPoolID, vmSetNameOfLB)
	if err != nil {
		return err
	}

	isOperationSucceeded = true
	return nil
}

func (fs *FlexScaleSet) ensureBackendPoolDeletedFromVmssFlex(backendPoolIDs []string, vmSetName string) error {
	vmssNamesMap := make(map[string]bool)
	if fs.useStandardLoadBalancer() {
		cached, err := fs.vmssFlexCache.Get(consts.VmssFlexKey, azcache.CacheReadTypeDefault)
		if err != nil {
			klog.Errorf("ensureBackendPoolDeletedFromVmssFlex: failed to get vmss flex from cache: %v", err)
			return err
		}
		vmssFlexes := cached.(*sync.Map)
		vmssFlexes.Range(func(key, value interface{}) bool {
			vmssFlex := value.(*compute.VirtualMachineScaleSet)
			vmssNamesMap[pointer.StringDeref(vmssFlex.Name, "")] = true
			return true
		})
	} else {
		vmssNamesMap[vmSetName] = true
	}
	return fs.EnsureBackendPoolDeletedFromVMSets(vmssNamesMap, backendPoolIDs)
}

// EnsureBackendPoolDeletedFromVMSets ensures the loadBalancer backendAddressPools deleted from the specified VMSS Flex
func (fs *FlexScaleSet) EnsureBackendPoolDeletedFromVMSets(vmssNamesMap map[string]bool, backendPoolIDs []string) error {
	vmssUpdaters := make([]func() error, 0, len(vmssNamesMap))
	errors := make([]error, 0, len(vmssNamesMap))
	for vmssName := range vmssNamesMap {
		vmssName := vmssName
		vmss, err := fs.getVmssFlexByName(vmssName)
		if err != nil {
			klog.Errorf("fs.EnsureBackendPoolDeletedFromVMSets: failed to get VMSS %s: %v", vmssName, err)
			errors = append(errors, err)
			continue
		}

		// When vmss is being deleted, CreateOrUpdate API would report "the vmss is being deleted" error.
		// Since it is being deleted, we shouldn't send more CreateOrUpdate requests for it.
		if vmss.ProvisioningState != nil && strings.EqualFold(*vmss.ProvisioningState, consts.ProvisionStateDeleting) {
			klog.V(3).Infof("fs.EnsureBackendPoolDeletedFromVMSets: found vmss %s being deleted, skipping", vmssName)
			continue
		}
		if vmss.VirtualMachineProfile == nil || vmss.VirtualMachineProfile.NetworkProfile == nil || vmss.VirtualMachineProfile.NetworkProfile.NetworkInterfaceConfigurations == nil {
			klog.V(4).Infof("fs.EnsureBackendPoolDeletedFromVMSets: cannot obtain the primary network interface configurations, of vmss %s", vmssName)
			continue
		}
		vmssNIC := *vmss.VirtualMachineProfile.NetworkProfile.NetworkInterfaceConfigurations
		primaryNIC, err := getPrimaryNetworkInterfaceConfiguration(vmssNIC, vmssName)
		if err != nil {
			klog.Errorf("fs.EnsureBackendPoolDeletedFromVMSets: failed to get the primary network interface config of the VMSS %s: %v", vmssName, err)
			errors = append(errors, err)
			continue
		}
		foundTotal := false
		for _, backendPoolID := range backendPoolIDs {
			found, err := deleteBackendPoolFromIPConfig("FlexSet.EnsureBackendPoolDeletedFromVMSets", backendPoolID, vmssName, primaryNIC)
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
							NetworkAPIVersion:              compute.TwoZeroTwoZeroHyphenMinusOneOneHyphenMinusZeroOne,
						},
					},
				},
			}

			defer func() {
				_ = fs.vmssFlexCache.Delete(consts.VmssFlexKey)
			}()

			klog.V(2).Infof("fs.EnsureBackendPoolDeletedFromVMSets begins to delete backendPoolIDs %q from vmss(%s)", backendPoolIDs, vmssName)
			rerr := fs.CreateOrUpdateVMSS(fs.ResourceGroup, vmssName, newVMSS)
			if rerr != nil {
				klog.Errorf("fs.EnsureBackendPoolDeletedFromVMSets CreateOrUpdateVMSS(%s) for backendPoolIDs %q, err: %v", vmssName, backendPoolIDs, rerr)
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

// EnsureBackendPoolDeleted ensures the loadBalancer backendAddressPools deleted from the specified nodes.
func (fs *FlexScaleSet) EnsureBackendPoolDeleted(service *v1.Service, backendPoolIDs []string, vmSetName string, backendAddressPools *[]network.BackendAddressPool, deleteFromVMSet bool) (bool, error) {
	// Returns nil if backend address pools already deleted.
	if backendAddressPools == nil {
		return false, nil
	}

	mc := metrics.NewMetricContext("services", "vmssflex_ensure_backend_pool_deleted", fs.ResourceGroup, fs.SubscriptionID, getServiceName(service))
	isOperationSucceeded := false
	defer func() {
		mc.ObserveOperationWithResult(isOperationSucceeded)
	}()

	ipConfigurationIDs := []string{}
	for _, backendPool := range *backendAddressPools {
		for _, backendPoolID := range backendPoolIDs {
			if strings.EqualFold(pointer.StringDeref(backendPool.ID, ""), backendPoolID) && backendPool.BackendAddressPoolPropertiesFormat != nil && backendPool.BackendIPConfigurations != nil {
				for _, ipConf := range *backendPool.BackendIPConfigurations {
					if ipConf.ID == nil {
						continue
					}

					ipConfigurationIDs = append(ipConfigurationIDs, *ipConf.ID)
				}
			}
		}
	}

	vmssFlexVMNameMap := make(map[string]string)
	allErrs := make([]error, 0)
	for i := range ipConfigurationIDs {
		ipConfigurationID := ipConfigurationIDs[i]
		nodeName, vmssFlexName, nicName, err := fs.getNodeInformationByIPConfigurationID(ipConfigurationID)
		if err != nil {
			continue
		}
		if nodeName == "" {
			continue
		}
		resourceGroupName, err := fs.GetNodeResourceGroup(nodeName)
		if err != nil {
			continue
		}
		// only vmsses in the resource group same as it's in azure config are included
		if strings.EqualFold(resourceGroupName, fs.ResourceGroup) {
			if fs.useStandardLoadBalancer() {
				vmssFlexVMNameMap[nodeName] = nicName
			} else {
				if strings.EqualFold(vmssFlexName, vmSetName) {
					vmssFlexVMNameMap[nodeName] = nicName
				} else {
					// Only remove nodes belonging to specified vmSet.
					continue
				}
			}

		}
	}

	klog.V(2).Infof("Ensure backendPoolIDs %q deleted from the VMSS.", backendPoolIDs)
	if deleteFromVMSet {
		err := fs.ensureBackendPoolDeletedFromVmssFlex(backendPoolIDs, vmSetName)
		if err != nil {
			allErrs = append(allErrs, err)
		}
	}

	klog.V(2).Infof("Ensure backendPoolIDs %q deleted from the VMSS VMs.", backendPoolIDs)
	klog.V(2).Infof("go into fs.ensureBackendPoolDeletedFromNode, vmssFlexVMNameMap: %s, size: %d", vmssFlexVMNameMap, len(vmssFlexVMNameMap))
	nicUpdated, err := fs.ensureBackendPoolDeletedFromNode(vmssFlexVMNameMap, backendPoolIDs)
	klog.V(2).Infof("exit from fs.ensureBackendPoolDeletedFromNode")
	if err != nil {
		allErrs = append(allErrs, err)
	}

	if len(allErrs) > 0 {
		return nicUpdated, utilerrors.Flatten(utilerrors.NewAggregate(allErrs))
	}

	isOperationSucceeded = true
	return nicUpdated, nil

}

func (fs *FlexScaleSet) ensureBackendPoolDeletedFromNode(vmssFlexVMNameMap map[string]string, backendPoolIDs []string) (bool, error) {
	nicUpdaters := make([]func() error, 0)
	allErrs := make([]error, 0)
	nics := map[string]network.Interface{} // nicName -> nic
	for nodeName, nicName := range vmssFlexVMNameMap {
		if _, ok := nics[nicName]; ok {
			continue
		}

		ctx, cancel := getContextWithCancel()
		defer cancel()
		nic, rerr := fs.InterfacesClient.Get(ctx, fs.ResourceGroup, nicName, "")
		if rerr != nil {
			return false, fmt.Errorf("ensureBackendPoolDeletedFromNode: failed to get interface of name %s: %w", nicName, rerr.Error())
		}

		if nic.ProvisioningState == consts.NicFailedState {
			klog.Warningf("EnsureBackendPoolDeleted skips node %s because its primary nic %s is in Failed state", nodeName, *nic.Name)
			continue
		}

		if nic.InterfacePropertiesFormat != nil && nic.InterfacePropertiesFormat.IPConfigurations != nil {
			nicName := pointer.StringDeref(nic.Name, "")
			nics[nicName] = nic
		}
	}
	var nicUpdated atomic.Bool
	for _, nic := range nics {
		nic := nic
		newIPConfigs := *nic.IPConfigurations
		for j, ipConf := range newIPConfigs {
			if !pointer.BoolDeref(ipConf.Primary, false) {
				continue
			}
			// found primary ip configuration
			if ipConf.LoadBalancerBackendAddressPools != nil {
				newLBAddressPools := *ipConf.LoadBalancerBackendAddressPools
				for k := len(newLBAddressPools) - 1; k >= 0; k-- {
					pool := newLBAddressPools[k]
					for _, backendPoolID := range backendPoolIDs {
						if strings.EqualFold(pointer.StringDeref(pool.ID, ""), backendPoolID) {
							newLBAddressPools = append(newLBAddressPools[:k], newLBAddressPools[k+1:]...)
						}
					}
				}
				newIPConfigs[j].LoadBalancerBackendAddressPools = &newLBAddressPools
			}
		}
		nic.IPConfigurations = &newIPConfigs

		nicUpdaters = append(nicUpdaters, func() error {
			ctx, cancel := getContextWithCancel()
			defer cancel()
			klog.V(2).Infof("EnsureBackendPoolDeleted begins to CreateOrUpdate for NIC(%s, %s) with backendPoolIDs %q", fs.ResourceGroup, pointer.StringDeref(nic.Name, ""), backendPoolIDs)
			rerr := fs.InterfacesClient.CreateOrUpdate(ctx, fs.ResourceGroup, pointer.StringDeref(nic.Name, ""), nic)
			if rerr != nil {
				klog.Errorf("EnsureBackendPoolDeleted CreateOrUpdate for NIC(%s, %s) failed with error %v", fs.ResourceGroup, pointer.StringDeref(nic.Name, ""), rerr.Error())
				return rerr.Error()
			}
			nicUpdated.Store(true)
			klog.V(2).Infof("EnsureBackendPoolDeleted done")
			return nil
		})
	}
	klog.V(2).Infof("nicUpdaters size: %d", len(nicUpdaters))
	errs := utilerrors.AggregateGoroutines(nicUpdaters...)
	if errs != nil {
		allErrs = append(allErrs, utilerrors.Flatten(errs))
	}
	if len(allErrs) > 0 {
		return nicUpdated.Load(), utilerrors.Flatten(utilerrors.NewAggregate(allErrs))
	}
	return nicUpdated.Load(), nil
}
