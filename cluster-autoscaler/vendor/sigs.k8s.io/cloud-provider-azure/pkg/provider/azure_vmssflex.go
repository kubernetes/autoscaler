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
	"sigs.k8s.io/cloud-provider-azure/pkg/util/lockmap"
	vmutil "sigs.k8s.io/cloud-provider-azure/pkg/util/vm"
)

// ErrorVmssIDIsEmpty indicates the vmss id is empty.
var ErrorVmssIDIsEmpty = errors.New("VMSS ID is empty")

// FlexScaleSet implements VMSet interface for Azure Flexible VMSS.
type FlexScaleSet struct {
	*Cloud

	vmssFlexCache            azcache.Resource
	vmssFlexVMNameToVmssID   *sync.Map
	vmssFlexVMNameToNodeName *sync.Map
	vmssFlexVMCache          azcache.Resource

	// lockMap in cache refresh
	lockMap *lockmap.LockMap
}

// RefreshCaches invalidates and renew all related caches.
func (fs *FlexScaleSet) RefreshCaches() error {
	logger := log.Background().WithName("fs.RefreshCaches")
	var err error
	fs.vmssFlexCache, err = fs.newVmssFlexCache()
	if err != nil {
		logger.Error(err, "failed to create or refresh vmssFlexCache")
		return err
	}
	fs.vmssFlexVMCache, err = fs.newVmssFlexVMCache()
	if err != nil {
		logger.Error(err, "failed to create or refresh vmssFlexVMCache")
		return err
	}
	return nil
}

func newFlexScaleSet(az *Cloud) (VMSet, error) {
	fs := &FlexScaleSet{
		Cloud:                    az,
		vmssFlexVMNameToVmssID:   &sync.Map{},
		vmssFlexVMNameToNodeName: &sync.Map{},
		lockMap:                  lockmap.NewLockMap(),
	}

	if err := fs.RefreshCaches(); err != nil {
		return nil, err
	}

	return fs, nil
}

// GetPrimaryVMSetName returns the VM set name depending on the configured vmType.
// It returns config.PrimaryScaleSetName for vmss and config.PrimaryAvailabilitySetName for standard vmType.
func (fs *FlexScaleSet) GetPrimaryVMSetName() string {
	return fs.PrimaryScaleSetName
}

// getNodeVMSetName returns the vmss flex name by the node name.
func (fs *FlexScaleSet) getNodeVmssFlexName(ctx context.Context, nodeName string) (string, error) {
	vmssFlexID, err := fs.getNodeVmssFlexID(ctx, nodeName)
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
func (fs *FlexScaleSet) GetNodeVMSetName(ctx context.Context, node *v1.Node) (string, error) {
	return fs.getNodeVmssFlexName(ctx, node.Name)
}

// GetAgentPoolVMSetNames returns all vmSet names according to the nodes
func (fs *FlexScaleSet) GetAgentPoolVMSetNames(ctx context.Context, nodes []*v1.Node) ([]*string, error) {
	logger := log.FromContextOrBackground(ctx).WithName("GetAgentPoolVMSetNames")
	vmSetNames := make([]*string, 0)
	for _, node := range nodes {
		vmSetName, err := fs.GetNodeVMSetName(ctx, node)
		if err != nil {
			logger.Error(err, "Unable to get the vmss flex name by node name", "node", node.Name)
			continue
		}
		vmSetNames = append(vmSetNames, &vmSetName)
	}
	return vmSetNames, nil
}

// GetVMSetNames selects all possible availability sets or scale sets
// (depending vmType configured) for service load balancer, if the service has
// no loadbalancer mode annotation returns the primary VMSet. If service annotation
// for loadbalancer exists then returns the eligible VMSet. The mode selection
// annotation would be ignored when using one SLB per cluster.
func (fs *FlexScaleSet) GetVMSetNames(ctx context.Context, service *v1.Service, nodes []*v1.Node) ([]*string, error) {
	logger := log.FromContextOrBackground(ctx).WithName("fs.GetVMSetNames")
	hasMode, isAuto, serviceVMSetName := fs.getServiceLoadBalancerMode(service)
	if !hasMode || fs.UseStandardLoadBalancer() {
		// no mode specified in service annotation or use single SLB mode
		// default to PrimaryScaleSetName
		vmssFlexNames := to.SliceOfPtrs(fs.PrimaryScaleSetName)
		return vmssFlexNames, nil
	}

	vmssFlexNames, err := fs.GetAgentPoolVMSetNames(ctx, nodes)
	if err != nil {
		logger.Error(err, "GetAgentPoolVMSetNames failed")
		return nil, err
	}

	if !isAuto {
		found := false
		for asx := range vmssFlexNames {
			if strings.EqualFold(*(vmssFlexNames)[asx], serviceVMSetName) {
				found = true
				serviceVMSetName = *(vmssFlexNames)[asx]
				break
			}
		}
		if !found {
			err = fmt.Errorf("scale set (%s) - not found", serviceVMSetName)
			logger.Error(err, "scale set in service annotation not found", "scaleSet", serviceVMSetName)
			return nil, err
		}
		return to.SliceOfPtrs(serviceVMSetName), nil
	}
	return vmssFlexNames, nil
}

// GetNodeNameByProviderID gets the node name by provider ID.
// providerID example:
// azure:///subscriptions/sub/resourceGroups/rg/providers/Microsoft.Compute/virtualMachines/flexprofile-mp-0_df53ee36
// Different from vmas where vm name is always equal to nodeName, we need to further map vmName to actual nodeName in vmssflex.
// Note: nodeName is always equal ptr.Derefs.ToLower(*vm.Properties.OSProfile.ComputerName, "")
func (fs *FlexScaleSet) GetNodeNameByProviderID(ctx context.Context, providerID string) (types.NodeName, error) {
	// NodeName is part of providerID for standard instances.
	matches := providerIDRE.FindStringSubmatch(providerID)
	if len(matches) != 2 {
		return "", errors.New("error splitting providerID")
	}

	nodeName, err := fs.getNodeNameByVMName(ctx, matches[1])
	if err != nil {
		return "", err
	}
	return types.NodeName(nodeName), nil
}

// GetInstanceIDByNodeName gets the cloud provider ID by node name.
// It must return ("", cloudprovider.InstanceNotFound) if the instance does
// not exist or is no longer running.
func (fs *FlexScaleSet) GetInstanceIDByNodeName(ctx context.Context, name string) (string, error) {
	logger := log.FromContextOrBackground(ctx).WithName("GetInstanceIDByNodeName")
	machine, err := fs.getVmssFlexVM(ctx, name, azcache.CacheReadTypeUnsafe)
	if err != nil {
		return "", err
	}
	if machine.ID == nil {
		return "", fmt.Errorf("ProviderID of node(%s) is nil", name)
	}
	resourceID := *machine.ID
	convertedResourceID, err := ConvertResourceGroupNameToLower(resourceID)
	if err != nil {
		logger.Error(err, "ConvertResourceGroupNameToLower failed")
		return "", err
	}
	return convertedResourceID, nil
}

// GetInstanceTypeByNodeName gets the instance type by node name.
func (fs *FlexScaleSet) GetInstanceTypeByNodeName(ctx context.Context, name string) (string, error) {
	logger := log.FromContextOrBackground(ctx).WithName("GetInstanceTypeByNodeName")
	machine, err := fs.getVmssFlexVM(ctx, name, azcache.CacheReadTypeUnsafe)
	if err != nil {
		logger.Error(err, "fs.getVmssFlexVM failed", "node", name)
		return "", err
	}

	if machine.Properties.HardwareProfile == nil {
		return "", fmt.Errorf("HardwareProfile of node(%s) is nil", name)
	}
	return string(*machine.Properties.HardwareProfile.VMSize), nil
}

// GetZoneByNodeName gets availability zone for the specified node. If the node is not running
// with availability zone, then it returns fault domain.
// for details, refer to https://kubernetes-sigs.github.io/cloud-provider-azure/topics/availability-zones/#node-labels
func (fs *FlexScaleSet) GetZoneByNodeName(ctx context.Context, name string) (cloudprovider.Zone, error) {
	logger := log.FromContextOrBackground(ctx).WithName("GetZoneByNodeName")
	vm, err := fs.getVmssFlexVM(ctx, name, azcache.CacheReadTypeUnsafe)
	if err != nil {
		logger.Error(err, "fs.getVmssFlexVM failed", "node", name)
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

		failureDomain = fs.makeZone(ptr.Deref(vm.Location, ""), zoneID)
	} else if vm.Properties.InstanceView != nil && vm.Properties.InstanceView.PlatformFaultDomain != nil {
		// Availability zone is not used for the node, falling back to fault domain.
		failureDomain = strconv.Itoa(int(ptr.Deref(vm.Properties.InstanceView.PlatformFaultDomain, 0)))
	} else {
		err = fmt.Errorf("failed to get zone info")
		logger.Error(err, "got unexpected error")
		return cloudprovider.Zone{}, err
	}

	zone := cloudprovider.Zone{
		FailureDomain: strings.ToLower(failureDomain),
		Region:        strings.ToLower(ptr.Deref(vm.Location, "")),
	}
	return zone, nil
}

// GetProvisioningStateByNodeName returns the provisioningState for the specified node.
func (fs *FlexScaleSet) GetProvisioningStateByNodeName(ctx context.Context, name string) (provisioningState string, err error) {
	vm, err := fs.getVmssFlexVM(ctx, name, azcache.CacheReadTypeDefault)
	if err != nil {
		return provisioningState, err
	}

	if vm.Properties == nil || vm.Properties.ProvisioningState == nil {
		return provisioningState, nil
	}

	return ptr.Deref(vm.Properties.ProvisioningState, ""), nil
}

// GetPowerStatusByNodeName returns the powerState for the specified node.
func (fs *FlexScaleSet) GetPowerStatusByNodeName(ctx context.Context, name string) (powerState string, err error) {
	logger := log.FromContextOrBackground(ctx).WithName("fs.GetPowerStatusByNodeName")
	vm, err := fs.getVmssFlexVM(ctx, name, azcache.CacheReadTypeDefault)
	if err != nil {
		return powerState, err
	}

	if vm.Properties.InstanceView != nil {
		return vmutil.GetVMPowerState(ptr.Deref(vm.Name, ""), vm.Properties.InstanceView.Statuses), nil
	}

	// vm.Properties.InstanceView or vm.Properties.InstanceView.Statuses are nil when the VM is under deleting.
	logger.V(3).Info("InstanceView for node is nil, assuming it's deleting", "node", name)
	return consts.VMPowerStateUnknown, nil
}

// GetPrimaryInterface gets machine primary network interface by node name.
func (fs *FlexScaleSet) GetPrimaryInterface(ctx context.Context, nodeName string) (*armnetwork.Interface, error) {
	logger := log.FromContextOrBackground(ctx).WithName("GetPrimaryInterface")
	machine, err := fs.getVmssFlexVM(ctx, nodeName, azcache.CacheReadTypeDefault)
	if err != nil {
		logger.Error(err, "fs.getVmssFlexVMWithoutInstanceView failed", "node", nodeName)
		return nil, err
	}

	primaryNicID, err := getPrimaryInterfaceID(machine)
	if err != nil {
		return nil, err
	}
	nicName, err := getLastSegment(primaryNicID, "/")
	if err != nil {
		return nil, err
	}

	nicResourceGroup, err := extractResourceGroupByNicID(primaryNicID)
	if err != nil {
		return nil, err
	}

	nic, rerr := fs.ComputeClientFactory.GetInterfaceClient().Get(ctx, nicResourceGroup, nicName, nil)
	if rerr != nil {
		return nil, rerr
	}

	return nic, nil
}

// GetIPByNodeName gets machine private IP and public IP by node name.
func (fs *FlexScaleSet) GetIPByNodeName(ctx context.Context, name string) (string, string, error) {
	logger := log.FromContextOrBackground(ctx).WithName("GetIPByNodeName")
	nic, err := fs.GetPrimaryInterface(ctx, name)
	if err != nil {
		return "", "", err
	}

	ipConfig, err := getPrimaryIPConfig(nic)
	if err != nil {
		logger.Error(err, "getPrimaryIPConfig failed", "node", name, "nic", nic)
		return "", "", err
	}

	privateIP := *ipConfig.Properties.PrivateIPAddress
	publicIP := ""
	if ipConfig.Properties.PublicIPAddress != nil && ipConfig.Properties.PublicIPAddress.ID != nil {
		pipID := *ipConfig.Properties.PublicIPAddress.ID
		pipName, err := getLastSegment(pipID, "/")
		if err != nil {
			return "", "", fmt.Errorf("failed to publicIP name for node %q with pipID %q", name, pipID)
		}
		pip, existsPip, err := fs.getPublicIPAddress(ctx, fs.ResourceGroup, pipName, azcache.CacheReadTypeDefault)
		if err != nil {
			return "", "", err
		}
		if existsPip {
			publicIP = *pip.Properties.IPAddress
		}
	}

	return privateIP, publicIP, nil
}

// GetPrivateIPsByNodeName returns a slice of all private ips assigned to node (ipv6 and ipv4)
// TODO (khenidak): This should read all nics, not just the primary
// allowing users to split ipv4/v6 on multiple nics
func (fs *FlexScaleSet) GetPrivateIPsByNodeName(ctx context.Context, name string) ([]string, error) {
	ips := make([]string, 0)
	nic, err := fs.GetPrimaryInterface(ctx, name)
	if err != nil {
		return ips, err
	}

	if nic.Properties.IPConfigurations == nil {
		return ips, fmt.Errorf("nic.Properties.IPConfigurations for nic (nicname=%s) is nil", *nic.Name)
	}

	for _, ipConfig := range nic.Properties.IPConfigurations {
		if ipConfig.Properties.PrivateIPAddress != nil {
			ips = append(ips, *ipConfig.Properties.PrivateIPAddress)
		}
	}

	return ips, nil
}

// GetNodeNameByIPConfigurationID gets the nodeName and vmSetName by IP configuration ID.
func (fs *FlexScaleSet) GetNodeNameByIPConfigurationID(ctx context.Context, ipConfigurationID string) (string, string, error) {
	logger := log.FromContextOrBackground(ctx).WithName("GetNodeNameByIPConfigurationID")
	nodeName, vmssFlexName, _, err := fs.getNodeInformationByIPConfigurationID(ctx, ipConfigurationID)
	if err != nil {
		logger.Error(err, "fs.GetNodeNameByIPConfigurationID failed", "ipConfigurationID", ipConfigurationID)
		return "", "", err
	}

	return nodeName, strings.ToLower(vmssFlexName), nil
}

func (fs *FlexScaleSet) getNodeInformationByIPConfigurationID(ctx context.Context, ipConfigurationID string) (string, string, string, error) {
	logger := log.FromContextOrBackground(ctx).WithName("getNodeInformationByIPConfigurationID")
	nicResourceGroup, nicName, err := getResourceGroupAndNameFromNICID(ipConfigurationID)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to get resource group and name from ip config ID %s: %w", ipConfigurationID, err)
	}

	// get vmName by nic name
	vmName, err := fs.GetVMNameByIPConfigurationName(ctx, nicResourceGroup, nicName)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to get vm name of ip config ID %s: %w", ipConfigurationID, err)
	}

	nodeName, err := fs.getNodeNameByVMName(ctx, vmName)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to map VM Name to NodeName: VM Name %s: %w", vmName, err)
	}

	vmssFlexName, err := fs.getNodeVmssFlexName(ctx, nodeName)
	if err != nil {
		logger.Error(err, "Unable to get the vmss flex name by node name", "vmName", vmName)
		return "", "", "", err
	}

	return nodeName, strings.ToLower(vmssFlexName), nicName, nil
}

// GetNodeCIDRMaskByProviderID returns the node CIDR subnet mask by provider ID.
func (fs *FlexScaleSet) GetNodeCIDRMasksByProviderID(ctx context.Context, providerID string) (int, int, error) {
	logger := log.FromContextOrBackground(ctx).WithName("GetNodeCIDRMasksByProviderID")
	nodeNameWrapper, err := fs.GetNodeNameByProviderID(ctx, providerID)
	if err != nil {
		logger.Error(err, "Unable to get the vmss flex vm node name by providerID", "providerID", providerID)
		return 0, 0, err
	}
	nodeName := mapNodeNameToVMName(nodeNameWrapper)

	vmssFlex, err := fs.getVmssFlexByNodeName(ctx, nodeName, azcache.CacheReadTypeDefault)
	if err != nil {
		if errors.Is(err, cloudprovider.InstanceNotFound) {
			return consts.DefaultNodeMaskCIDRIPv4, consts.DefaultNodeMaskCIDRIPv6, nil
		}
		return 0, 0, err
	}

	var ipv4Mask, ipv6Mask int
	if v4, ok := vmssFlex.Tags[consts.VMSetCIDRIPV4TagKey]; ok && v4 != nil {
		ipv4Mask, err = strconv.Atoi(ptr.Deref(v4, ""))
		if err != nil {
			logger.Error(err, "error when parsing the value of the ipv4 mask size", "value", ptr.Deref(v4, ""))
		}
	}
	if v6, ok := vmssFlex.Tags[consts.VMSetCIDRIPV6TagKey]; ok && v6 != nil {
		ipv6Mask, err = strconv.Atoi(ptr.Deref(v6, ""))
		if err != nil {
			logger.Error(err, "error when parsing the value of the ipv6 mask size", "value", ptr.Deref(v6, ""))
		}
	}

	return ipv4Mask, ipv6Mask, nil
}

// EnsureHostInPool ensures the given VM's Primary NIC's Primary IP Configuration is
// participating in the specified LoadBalancer Backend Pool, which returns (resourceGroup, vmasName, instanceID, vmssVM, error).
func (fs *FlexScaleSet) EnsureHostInPool(ctx context.Context, service *v1.Service, nodeName types.NodeName, backendPoolID string, vmSetNameOfLB string) (string, string, string, *armcompute.VirtualMachineScaleSetVM, error) {
	logger := log.FromContextOrBackground(ctx).WithName("EnsureHostInPool")
	serviceName := getServiceName(service)
	name := mapNodeNameToVMName(nodeName)
	vmssFlexName, err := fs.getNodeVmssFlexName(ctx, name)
	if err != nil {
		logger.Error(err, "failed to get VMSS Flex Name", "name", name)
		return "", "", "", nil, nil
	}

	// Check scale set name:
	// - For basic SKU load balancer, return error as VMSS Flex does not support basic load balancer.
	// - For single standard SKU load balancer, backend could belong to multiple VMSS, so we
	//   don't check vmSet for it.
	// - For multiple standard SKU load balancers, return nil if the node's scale set is mismatched with vmSetNameOfLB
	needCheck := false
	if !fs.UseStandardLoadBalancer() {
		return "", "", "", nil, fmt.Errorf("EnsureHostInPool: VMSS Flex does not support Basic Load Balancer")
	}
	if vmSetNameOfLB != "" && needCheck && !strings.EqualFold(vmSetNameOfLB, vmssFlexName) {
		logger.V(3).Info("skips node because it is not in the ScaleSet", "node", name, "ScaleSet", vmSetNameOfLB)
		return "", "", "", nil, errNotInVMSet
	}

	nic, err := fs.GetPrimaryInterface(ctx, name)
	if err != nil {
		logger.Error(err, "fs.GetPrimaryInterface failed", "node", name, "vmSetNameOfLB", vmSetNameOfLB)
		return "", "", "", nil, err
	}

	if *nic.Properties.ProvisioningState == consts.NicFailedState {
		klog.Warningf("EnsureHostInPool skips node %s because its primary nic %s is in Failed state", nodeName, *nic.Name)
		return "", "", "", nil, nil
	}

	var primaryIPConfig *armnetwork.InterfaceIPConfiguration
	ipv6 := isBackendPoolIPv6(backendPoolID)
	if !fs.ipv6DualStackEnabled && !ipv6 {
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
	newBackendPools := []*armnetwork.BackendAddressPool{}
	if primaryIPConfig.Properties.LoadBalancerBackendAddressPools != nil {
		newBackendPools = primaryIPConfig.Properties.LoadBalancerBackendAddressPools
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

	if fs.UseStandardLoadBalancer() && len(newBackendPools) > 0 {
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
			logger.V(4).Info("Node has already been added to LB, omit adding it to a new one", "node", nodeName, "LB", oldLBName)
			return "", "", "", nil, nil
		}
	}

	newBackendPools = append(newBackendPools,
		&armnetwork.BackendAddressPool{
			ID: ptr.To(backendPoolID),
		})

	primaryIPConfig.Properties.LoadBalancerBackendAddressPools = newBackendPools

	nicName := *nic.Name
	logger.V(3).Info("updating", "nicupdate", serviceName, "nic", nicName)
	err = fs.CreateOrUpdateInterface(ctx, service, nic)
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

func (fs *FlexScaleSet) ensureVMSSFlexInPool(ctx context.Context, _ *v1.Service, nodes []*v1.Node, backendPoolID string, vmSetNameOfLB string) error {
	logger := log.FromContextOrBackground(ctx).WithName("ensureVMSSFlexInPool")
	logger.V(2).Info("ensuring VMSS Flex with backendPoolID", "backendPoolID", backendPoolID)
	vmssFlexIDsMap := make(map[string]bool)

	if !fs.UseStandardLoadBalancer() {
		return fmt.Errorf("ensureVMSSFlexInPool: VMSS Flex does not support Basic Load Balancer")
	}

	// the single standard load balancer supports multiple vmss in its backend while
	// multiple standard load balancers doesn't
	if fs.UseStandardLoadBalancer() {
		for _, node := range nodes {
			if fs.ExcludeMasterNodesFromStandardLB() && isControlPlaneNode(node) {
				continue
			}

			shouldExcludeLoadBalancer, err := fs.ShouldNodeExcludedFromLoadBalancer(node.Name)
			if err != nil {
				logger.Error(err, "ShouldNodeExcludedFromLoadBalancer failed", "node", node.Name)
				return err
			}
			if shouldExcludeLoadBalancer {
				logger.V(4).Info("Excluding unmanaged/external-resource-group node", "node", node.Name)
				continue
			}

			// in this scenario the vmSetName is an empty string and the name of vmss should be obtained from the provider IDs of nodes
			vmssFlexID, err := fs.getNodeVmssFlexID(ctx, node.Name)
			if err != nil {
				logger.Error(err, "failed to get VMSS Flex ID of node, will skip checking and continue", "node", node.Name)
				continue
			}
			resourceGroupName, err := fs.GetNodeResourceGroup(node.Name)
			if err != nil {
				logger.Error(err, "failed to get resource group of node, will skip checking and continue", "node", node.Name)
				continue
			}

			// only vmsses in the resource group same as it's in azure config are included
			if strings.EqualFold(resourceGroupName, fs.ResourceGroup) {
				vmssFlexIDsMap[vmssFlexID] = true
			}
		}
	} else {
		vmssFlexID, err := fs.getVmssFlexIDByName(ctx, vmSetNameOfLB)
		if err != nil {
			logger.Error(err, "failed to get VMSS Flex ID of vmSet", "vmSet", vmSetNameOfLB)
			return err
		}
		vmssFlexIDsMap[vmssFlexID] = true
	}

	logger.V(2).Info("begins to update VMSS list with backendPoolID", "VMSS list", vmssFlexIDsMap, "backendPoolID", backendPoolID)
	for vmssFlexID := range vmssFlexIDsMap {
		vmssFlex, err := fs.getVmssFlexByVmssFlexID(ctx, vmssFlexID, azcache.CacheReadTypeDefault)
		if err != nil {
			return err
		}
		vmssFlexName := *vmssFlex.Name

		// When vmss is being deleted, CreateOrUpdate API would report "the vmss is being deleted" error.
		// Since it is being deleted, we shouldn't send more CreateOrUpdate requests for it.
		if vmssFlex.Properties.ProvisioningState != nil && strings.EqualFold(*vmssFlex.Properties.ProvisioningState, consts.ProvisionStateDeleting) {
			logger.V(3).Info("found vmss being deleted, skipping", "vmss", vmssFlexID)
			continue
		}

		if vmssFlex.Properties.VirtualMachineProfile == nil || vmssFlex.Properties.VirtualMachineProfile.NetworkProfile == nil || vmssFlex.Properties.VirtualMachineProfile.NetworkProfile.NetworkInterfaceConfigurations == nil {
			logger.V(4).Info("cannot obtain the primary network interface configuration of vmss, just skip it as it might not have default vm profile", "vmss", vmssFlexID)
			continue
		}
		vmssNIC := vmssFlex.Properties.VirtualMachineProfile.NetworkProfile.NetworkInterfaceConfigurations
		primaryNIC, err := getPrimaryNetworkInterfaceConfiguration(vmssNIC, vmssFlexName)
		if err != nil {
			return err
		}
		primaryIPConfig, err := getPrimaryIPConfigFromVMSSNetworkConfig(primaryNIC, backendPoolID, vmssFlexName)
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

		if fs.UseStandardLoadBalancer() && len(loadBalancerBackendAddressPools) > 0 {
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
				logger.V(4).Info("VMSS has already been added to LB, omit adding it to a new one", "vmss", vmssFlexID, "LB", oldLBName)
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
			Location: vmssFlex.Location,
			Properties: &armcompute.VirtualMachineScaleSetProperties{
				VirtualMachineProfile: &armcompute.VirtualMachineScaleSetVMProfile{
					NetworkProfile: &armcompute.VirtualMachineScaleSetNetworkProfile{
						NetworkInterfaceConfigurations: vmssNIC,
						NetworkAPIVersion:              to.Ptr(armcompute.NetworkAPIVersionTwoThousandTwenty1101),
					},
				},
			},
		}

		defer func() {
			_ = fs.vmssFlexCache.Delete(consts.VmssFlexKey)
		}()

		logger.V(2).Info("begins to add vmss with new backendPoolID", "vmss", vmssFlexName, "backendPoolID", backendPoolID)
		rerr := fs.CreateOrUpdateVMSS(fs.ResourceGroup, vmssFlexName, newVMSS)
		if rerr != nil {
			logger.Error(rerr, "CreateOrUpdateVMSS with new backendPoolID", "vmss", vmssFlexName, "backendPoolID", backendPoolID)
			return rerr
		}
	}
	return nil
}

// EnsureHostsInPool ensures the given Node's primary IP configurations are
// participating in the specified LoadBalancer Backend Pool.
func (fs *FlexScaleSet) EnsureHostsInPool(ctx context.Context, service *v1.Service, nodes []*v1.Node, backendPoolID string, vmSetNameOfLB string) error {
	logger := log.FromContextOrBackground(ctx).WithName("EnsureHostsInPool")
	mc := metrics.NewMetricContext("services", "vmssflex_ensure_hosts_in_pool", fs.ResourceGroup, fs.SubscriptionID, getServiceName(service))
	isOperationSucceeded := false
	defer func() {
		mc.ObserveOperationWithResult(isOperationSucceeded)
	}()

	err := fs.ensureVMSSFlexInPool(ctx, service, nodes, backendPoolID, vmSetNameOfLB)
	if err != nil {
		return err
	}

	hostUpdates := make([]func() error, 0, len(nodes))
	for _, node := range nodes {
		localNodeName := node.Name
		if fs.UseStandardLoadBalancer() && fs.ExcludeMasterNodesFromStandardLB() && isControlPlaneNode(node) {
			logger.V(4).Info("Excluding master node from load balancer backendpool", "node", localNodeName, "backendPoolID", backendPoolID)
			continue
		}

		shouldExcludeLoadBalancer, err := fs.ShouldNodeExcludedFromLoadBalancer(localNodeName)
		if err != nil {
			logger.Error(err, "ShouldNodeExcludedFromLoadBalancer failed", "node", localNodeName)
			return err
		}
		if shouldExcludeLoadBalancer {
			logger.V(4).Info("Excluding unmanaged/external-resource-group node", "node", localNodeName)
			continue
		}

		f := func() error {
			_, _, _, _, err := fs.EnsureHostInPool(ctx, service, types.NodeName(localNodeName), backendPoolID, vmSetNameOfLB)
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

	isOperationSucceeded = true
	return nil
}

func (fs *FlexScaleSet) ensureBackendPoolDeletedFromVmssFlex(ctx context.Context, backendPoolIDs []string, vmSetName string) error {
	logger := log.FromContextOrBackground(ctx).WithName("ensureBackendPoolDeletedFromVmssFlex")
	vmssNamesMap := make(map[string]bool)
	if fs.UseStandardLoadBalancer() {
		cached, err := fs.vmssFlexCache.Get(ctx, consts.VmssFlexKey, azcache.CacheReadTypeDefault)
		if err != nil {
			logger.Error(err, "failed to get vmss flex from cache")
			return err
		}
		vmssFlexes := cached.(*sync.Map)
		vmssFlexes.Range(func(_, value interface{}) bool {
			vmssFlex := value.(*armcompute.VirtualMachineScaleSet)
			vmssNamesMap[ptr.Deref(vmssFlex.Name, "")] = true
			return true
		})
	} else {
		vmssNamesMap[vmSetName] = true
	}
	return fs.EnsureBackendPoolDeletedFromVMSets(ctx, vmssNamesMap, backendPoolIDs)
}

// EnsureBackendPoolDeletedFromVMSets ensures the loadBalancer backendAddressPools deleted from the specified VMSS Flex
func (fs *FlexScaleSet) EnsureBackendPoolDeletedFromVMSets(ctx context.Context, vmssNamesMap map[string]bool, backendPoolIDs []string) error {
	logger := log.FromContextOrBackground(ctx).WithName("fs.EnsureBackendPoolDeletedFromVMSets")
	vmssUpdaters := make([]func() error, 0, len(vmssNamesMap))
	errors := make([]error, 0, len(vmssNamesMap))
	for vmssName := range vmssNamesMap {
		vmssName := vmssName
		vmss, err := fs.getVmssFlexByName(ctx, vmssName)
		if err != nil {
			logger.Error(err, "failed to get VMSS", "vmss", vmssName)
			errors = append(errors, err)
			continue
		}

		// When vmss is being deleted, CreateOrUpdate API would report "the vmss is being deleted" error.
		// Since it is being deleted, we shouldn't send more CreateOrUpdate requests for it.
		if vmss.Properties.ProvisioningState != nil && strings.EqualFold(*vmss.Properties.ProvisioningState, consts.ProvisionStateDeleting) {
			logger.V(3).Info("found vmss being deleted, skipping", "vmss", vmssName)
			continue
		}
		if vmss.Properties.VirtualMachineProfile == nil || vmss.Properties.VirtualMachineProfile.NetworkProfile == nil || vmss.Properties.VirtualMachineProfile.NetworkProfile.NetworkInterfaceConfigurations == nil {
			logger.V(4).Info("cannot obtain the primary network interface configurations, of vmss", "vmss", vmssName)
			continue
		}
		vmssNIC := vmss.Properties.VirtualMachineProfile.NetworkProfile.NetworkInterfaceConfigurations
		primaryNIC, err := getPrimaryNetworkInterfaceConfiguration(vmssNIC, vmssName)
		if err != nil {
			logger.Error(err, "failed to get the primary network interface config of the VMSS", "vmss", vmssName)
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
			newVMSS := armcompute.VirtualMachineScaleSet{
				Location: vmss.Location,
				Properties: &armcompute.VirtualMachineScaleSetProperties{
					VirtualMachineProfile: &armcompute.VirtualMachineScaleSetVMProfile{
						NetworkProfile: &armcompute.VirtualMachineScaleSetNetworkProfile{
							NetworkInterfaceConfigurations: vmssNIC,
							NetworkAPIVersion:              to.Ptr(armcompute.NetworkAPIVersionTwoThousandTwenty1101),
						},
					},
				},
			}

			defer func() {
				_ = fs.vmssFlexCache.Delete(consts.VmssFlexKey)
			}()

			logger.V(2).Info("begins to delete backendPoolIDs from vmss", "backendPoolIDs", backendPoolIDs, "vmss", vmssName)
			rerr := fs.CreateOrUpdateVMSS(fs.ResourceGroup, vmssName, newVMSS)
			if rerr != nil {
				logger.Error(rerr, "CreateOrUpdateVMSS for backendPoolIDs", "vmss", vmssName, "backendPoolIDs", backendPoolIDs)
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

// EnsureBackendPoolDeleted ensures the loadBalancer backendAddressPools deleted from the specified nodes.
func (fs *FlexScaleSet) EnsureBackendPoolDeleted(ctx context.Context, service *v1.Service, backendPoolIDs []string, vmSetName string, backendAddressPools []*armnetwork.BackendAddressPool, deleteFromVMSet bool) (bool, error) {
	logger := log.FromContextOrBackground(ctx).WithName("EnsureBackendPoolDeleted")
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
	for _, backendPool := range backendAddressPools {
		for _, backendPoolID := range backendPoolIDs {
			if strings.EqualFold(ptr.Deref(backendPool.ID, ""), backendPoolID) && backendPool.Properties != nil && backendPool.Properties.BackendIPConfigurations != nil {
				for _, ipConf := range backendPool.Properties.BackendIPConfigurations {
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
		nodeName, vmssFlexName, nicName, err := fs.getNodeInformationByIPConfigurationID(ctx, ipConfigurationID)
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
			if fs.UseStandardLoadBalancer() {
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

	logger.V(2).Info("Ensure backendPoolIDs deleted from the VMSS", "backendPoolIDs", backendPoolIDs)
	if deleteFromVMSet {
		err := fs.ensureBackendPoolDeletedFromVmssFlex(ctx, backendPoolIDs, vmSetName)
		if err != nil {
			allErrs = append(allErrs, err)
		}
	}

	logger.V(2).Info("Ensure backendPoolIDs deleted from the VMSS VMs", "backendPoolIDs", backendPoolIDs)
	logger.V(2).Info("go into fs.ensureBackendPoolDeletedFromNode", "vmssFlexVMNameMap", vmssFlexVMNameMap, "size", len(vmssFlexVMNameMap))
	nicUpdated, err := fs.ensureBackendPoolDeletedFromNode(ctx, vmssFlexVMNameMap, backendPoolIDs)
	logger.V(2).Info("exit from fs.ensureBackendPoolDeletedFromNode")
	if err != nil {
		allErrs = append(allErrs, err)
	}

	if len(allErrs) > 0 {
		return nicUpdated, utilerrors.Flatten(utilerrors.NewAggregate(allErrs))
	}

	isOperationSucceeded = true
	return nicUpdated, nil
}

func (fs *FlexScaleSet) ensureBackendPoolDeletedFromNode(ctx context.Context, vmssFlexVMNameMap map[string]string, backendPoolIDs []string) (bool, error) {
	logger := log.FromContextOrBackground(ctx).WithName("ensureBackendPoolDeletedFromNode")
	nicUpdaters := make([]func() error, 0)
	allErrs := make([]error, 0)
	nics := map[string]*armnetwork.Interface{} // nicName -> nic
	for nodeName, nicName := range vmssFlexVMNameMap {
		if _, ok := nics[nicName]; ok {
			continue
		}

		nic, rerr := fs.ComputeClientFactory.GetInterfaceClient().Get(ctx, fs.ResourceGroup, nicName, nil)
		if rerr != nil {
			return false, fmt.Errorf("ensureBackendPoolDeletedFromNode: failed to get interface of name %s: %w", nicName, rerr)
		}

		if *nic.Properties.ProvisioningState == consts.NicFailedState {
			klog.Warningf("EnsureBackendPoolDeleted skips node %s because its primary nic %s is in Failed state", nodeName, *nic.Name)
			continue
		}

		if nic.Properties != nil && nic.Properties.IPConfigurations != nil {
			nicName := ptr.Deref(nic.Name, "")
			nics[nicName] = nic
		}
	}
	var nicUpdated atomic.Bool
	for _, nic := range nics {
		nic := nic
		newIPConfigs := nic.Properties.IPConfigurations
		for j, ipConf := range newIPConfigs {
			if !ptr.Deref(ipConf.Properties.Primary, false) {
				continue
			}
			// found primary ip configuration
			if ipConf.Properties.LoadBalancerBackendAddressPools != nil {
				newLBAddressPools := ipConf.Properties.LoadBalancerBackendAddressPools
				for k := len(newLBAddressPools) - 1; k >= 0; k-- {
					pool := newLBAddressPools[k]
					for _, backendPoolID := range backendPoolIDs {
						if strings.EqualFold(ptr.Deref(pool.ID, ""), backendPoolID) {
							newLBAddressPools = append(newLBAddressPools[:k], newLBAddressPools[k+1:]...)
						}
					}
				}
				newIPConfigs[j].Properties.LoadBalancerBackendAddressPools = newLBAddressPools
			}
		}
		nic.Properties.IPConfigurations = newIPConfigs

		nicUpdaters = append(nicUpdaters, func() error {
			logger.V(2).Info("begins to CreateOrUpdate for NIC with backendPoolIDs", "resourceGroup", fs.ResourceGroup, "nicName", ptr.Deref(nic.Name, ""), "backendPoolIDs", backendPoolIDs)
			_, rerr := fs.ComputeClientFactory.GetInterfaceClient().CreateOrUpdate(ctx, fs.ResourceGroup, ptr.Deref(nic.Name, ""), *nic)
			if rerr != nil {
				logger.Error(rerr, "CreateOrUpdate NIC failed", "resourceGroup", fs.ResourceGroup, "nicName", ptr.Deref(nic.Name, ""))
				return rerr
			}
			nicUpdated.Store(true)
			logger.V(2).Info("done")
			return nil
		})
	}
	logger.V(2).Info("NIC updaters size", "nicUpdatersSize", len(nicUpdaters))
	errs := utilerrors.AggregateGoroutines(nicUpdaters...)
	if errs != nil {
		allErrs = append(allErrs, utilerrors.Flatten(errs))
	}
	if len(allErrs) > 0 {
		return nicUpdated.Load(), utilerrors.Flatten(utilerrors.NewAggregate(allErrs))
	}
	return nicUpdated.Load(), nil
}
