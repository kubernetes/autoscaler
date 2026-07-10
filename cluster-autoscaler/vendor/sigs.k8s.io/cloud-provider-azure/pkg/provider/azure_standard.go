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
	"hash/crc32"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

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
	vmutil "sigs.k8s.io/cloud-provider-azure/pkg/util/vm"
)

var (
	errNotInVMSet      = errors.New("vm is not in the vmset")
	providerIDRE       = regexp.MustCompile(`.*/subscriptions/(?:.*)/Microsoft.Compute/virtualMachines/(.+)$`)
	backendPoolIDRE    = regexp.MustCompile(`^/subscriptions/(?:.*)/resourceGroups/(?:.*)/providers/Microsoft.Network/loadBalancers/(.+)/backendAddressPools/(.+)`)
	nicResourceGroupRE = regexp.MustCompile(`.*/subscriptions/(?:.*)/resourceGroups/(.+)/providers/Microsoft.Network/networkInterfaces/(?:.*)`)
	nicIDRE            = regexp.MustCompile(`(?i)/subscriptions/(?:.*)/resourceGroups/(.+)/providers/Microsoft.Network/networkInterfaces/(.+)/ipConfigurations/(?:.*)`)
	vmIDRE             = regexp.MustCompile(`(?i)/subscriptions/(?:.*)/resourceGroups/(?:.*)/providers/Microsoft.Compute/virtualMachines/(.+)`)
	vmasIDRE           = regexp.MustCompile(`/subscriptions/(?:.*)/resourceGroups/(?:.*)/providers/Microsoft.Compute/availabilitySets/(.+)`)
)

// returns the full identifier of an availabilitySet
func (az *Cloud) getAvailabilitySetID(resourceGroup, availabilitySetName string) string {
	return fmt.Sprintf(
		consts.AvailabilitySetIDTemplate,
		az.SubscriptionID,
		resourceGroup,
		availabilitySetName)
}

// returns the full identifier of a loadbalancer frontendipconfiguration.
func (az *Cloud) getFrontendIPConfigID(lbName, fipConfigName string) string {
	return az.getFrontendIPConfigIDWithRG(lbName, az.getLoadBalancerResourceGroup(), fipConfigName)
}

func (az *Cloud) getFrontendIPConfigIDWithRG(lbName, rgName, fipConfigName string) string {
	return fmt.Sprintf(
		consts.FrontendIPConfigIDTemplate,
		az.getNetworkResourceSubscriptionID(),
		rgName,
		lbName,
		fipConfigName)
}

// returns the full identifier of a loadbalancer backendpool.
func (az *Cloud) getBackendPoolID(lbName, backendPoolName string) string {
	return az.getBackendPoolIDWithRG(lbName, az.getLoadBalancerResourceGroup(), backendPoolName)
}

func (az *Cloud) getBackendPoolIDWithRG(lbName, rgName, backendPoolName string) string {
	return fmt.Sprintf(
		consts.BackendPoolIDTemplate,
		az.getNetworkResourceSubscriptionID(),
		rgName,
		lbName,
		backendPoolName)
}

func (az *Cloud) getBackendPoolIDs(clusterName, lbName string) map[bool]string {
	return map[bool]string{
		consts.IPVersionIPv4: az.getBackendPoolID(lbName, getBackendPoolName(clusterName, consts.IPVersionIPv4)),
		consts.IPVersionIPv6: az.getBackendPoolID(lbName, getBackendPoolName(clusterName, consts.IPVersionIPv6)),
	}
}

// returns the full identifier of a loadbalancer probe.
func (az *Cloud) getLoadBalancerProbeID(lbName, lbRuleName string) string {
	return az.getLoadBalancerProbeIDWithRG(lbName, az.getLoadBalancerResourceGroup(), lbRuleName)
}

func (az *Cloud) getLoadBalancerProbeIDWithRG(lbName, rgName, lbRuleName string) string {
	return fmt.Sprintf(
		consts.LoadBalancerProbeIDTemplate,
		az.getNetworkResourceSubscriptionID(),
		rgName,
		lbName,
		lbRuleName)
}

// getNetworkResourceSubscriptionID returns the subscription id which hosts network resources
func (az *Cloud) getNetworkResourceSubscriptionID() string {
	if az.UsesNetworkResourceInDifferentSubscription() {
		return az.NetworkResourceSubscriptionID
	}
	return az.SubscriptionID
}

func (az *Cloud) mapLoadBalancerNameToVMSet(lbName, clusterName string) (vmSetName string) {
	vmSetName = trimSuffixIgnoreCase(lbName, consts.InternalLoadBalancerNameSuffix)
	if strings.EqualFold(clusterName, vmSetName) {
		vmSetName = az.VMSet.GetPrimaryVMSetName()
	}

	return vmSetName
}

func (az *Cloud) mapVMSetNameToLoadBalancerName(vmSetName, clusterName string) string {
	if vmSetName == az.VMSet.GetPrimaryVMSetName() {
		return clusterName
	}
	return vmSetName
}

// isControlPlaneNode returns true if the node has a control-plane role label.
// The control-plane role is determined by looking for:
// * a node-role.kubernetes.io/control-plane or node-role.kubernetes.io/master="" label
func isControlPlaneNode(node *v1.Node) bool {
	if _, ok := node.Labels[consts.ControlPlaneNodeRoleLabel]; ok {
		return true
	}
	// include master role labels for k8s < 1.19
	if _, ok := node.Labels[consts.MasterNodeRoleLabel]; ok {
		return true
	}
	if val, ok := node.Labels[consts.NodeLabelRole]; ok && val == "master" {
		return true
	}
	return false
}

// returns the deepest child's identifier from a full identifier string.
func getLastSegment(ID, separator string) (string, error) {
	parts := strings.Split(ID, separator)
	name := parts[len(parts)-1]
	if len(name) == 0 {
		return "", fmt.Errorf("resource name was missing from identifier")
	}

	return name, nil
}

// returns the equivalent LoadBalancerRule, SecurityRule and LoadBalancerProbe
// protocol types for the given Kubernetes protocol type.
func getProtocolsFromKubernetesProtocol(protocol v1.Protocol) (*armnetwork.TransportProtocol, *armnetwork.SecurityRuleProtocol, *armnetwork.ProbeProtocol, error) {
	var transportProto *armnetwork.TransportProtocol
	var securityProto *armnetwork.SecurityRuleProtocol
	var probeProto *armnetwork.ProbeProtocol

	switch protocol {
	case v1.ProtocolTCP:
		transportProto = to.Ptr(armnetwork.TransportProtocolTCP)
		securityProto = to.Ptr(armnetwork.SecurityRuleProtocolTCP)
		probeProto = to.Ptr(armnetwork.ProbeProtocolTCP)
		return transportProto, securityProto, probeProto, nil
	case v1.ProtocolUDP:
		transportProto = to.Ptr(armnetwork.TransportProtocolUDP)
		securityProto = to.Ptr(armnetwork.SecurityRuleProtocolUDP)
		return transportProto, securityProto, nil, nil
	case v1.ProtocolSCTP:
		transportProto = to.Ptr(armnetwork.TransportProtocolAll)
		securityProto = to.Ptr(armnetwork.SecurityRuleProtocolAsterisk)
		return transportProto, securityProto, nil, nil
	default:
		return transportProto, securityProto, probeProto, fmt.Errorf("only TCP, UDP and SCTP are supported for Azure LoadBalancers")
	}
}

// This returns the full identifier of the primary NIC for the given VM.
func getPrimaryInterfaceID(machine *armcompute.VirtualMachine) (string, error) {
	if len(machine.Properties.NetworkProfile.NetworkInterfaces) == 1 {
		return *(machine.Properties.NetworkProfile.NetworkInterfaces)[0].ID, nil
	}

	for _, ref := range machine.Properties.NetworkProfile.NetworkInterfaces {
		if ptr.Deref(ref.Properties.Primary, false) {
			return *ref.ID, nil
		}
	}

	return "", fmt.Errorf("failed to find a primary nic for the vm. vmname=%q", *machine.Name)
}

func getPrimaryIPConfig(nic *armnetwork.Interface) (*armnetwork.InterfaceIPConfiguration, error) {
	if nic.Properties.IPConfigurations == nil {
		return nil, fmt.Errorf("nic.Properties.IPConfigurations for nic (nicname=%q) is nil", *nic.Name)
	}

	if len(nic.Properties.IPConfigurations) == 1 {
		return nic.Properties.IPConfigurations[0], nil
	}

	for _, ref := range nic.Properties.IPConfigurations {
		ref := ref
		if *ref.Properties.Primary {
			return ref, nil
		}
	}

	return nil, fmt.Errorf("failed to determine the primary ipconfig. nicname=%q", *nic.Name)
}

// returns first ip configuration on a nic by family
func getIPConfigByIPFamily(nic *armnetwork.Interface, IPv6 bool) (*armnetwork.InterfaceIPConfiguration, error) {
	if nic.Properties.IPConfigurations == nil {
		return nil, fmt.Errorf("nic.Properties.IPConfigurations for nic (nicname=%q) is nil", *nic.Name)
	}

	var ipVersion armnetwork.IPVersion
	if IPv6 {
		ipVersion = armnetwork.IPVersionIPv6
	} else {
		ipVersion = armnetwork.IPVersionIPv4
	}
	for _, ref := range nic.Properties.IPConfigurations {
		ref := ref
		if ref.Properties.PrivateIPAddress != nil && *ref.Properties.PrivateIPAddressVersion == ipVersion {
			return ref, nil
		}
	}
	return nil, fmt.Errorf("failed to determine the ipconfig(IPv6=%v). nicname=%q", IPv6, ptr.Deref(nic.Name, ""))
}

// getBackendPoolName the LB BackendPool name for a service.
// to ensure backward and forward compat:
// SingleStack -v4 (pre v1.16) => BackendPool name == clusterName
// SingleStack -v6 => BackendPool name == <clusterName>-IPv6 (all cluster bootstrap uses this name)
// DualStack
// => IPv4 BackendPool name == clusterName
// => IPv6 BackendPool name == <clusterName>-IPv6
// This means:
// clusters moving from IPv4 to dualstack will require no changes
// clusters moving from IPv6 to dualstack will require no changes as the IPv4 backend pool will created with <clusterName>
func getBackendPoolName(clusterName string, isIPv6 bool) string {
	if isIPv6 {
		return fmt.Sprintf("%s-%s", clusterName, consts.IPVersionIPv6String)
	}

	return clusterName
}

// getBackendPoolNames returns the IPv4 and IPv6 backend pool names.
func getBackendPoolNames(clusterName string) map[bool]string {
	return map[bool]string{
		consts.IPVersionIPv4: getBackendPoolName(clusterName, consts.IPVersionIPv4),
		consts.IPVersionIPv6: getBackendPoolName(clusterName, consts.IPVersionIPv6),
	}
}

// ifBackendPoolIPv6 checks if a backend pool is of IPv6 according to name/ID.
func isBackendPoolIPv6(name string) bool {
	return managedResourceHasIPv6Suffix(name)
}

func managedResourceHasIPv6Suffix(name string) bool {
	return strings.HasSuffix(strings.ToLower(name), fmt.Sprintf("-%s", consts.IPVersionIPv6StringLower))
}

func (az *Cloud) getLoadBalancerRuleName(service *v1.Service, protocol v1.Protocol, port int32, isIPv6 bool) string {
	prefix := az.getRulePrefix(service)
	ruleName := fmt.Sprintf("%s-%s-%d", prefix, protocol, port)
	subnet := getInternalSubnet(service)
	isDualStack := isServiceDualStack(service)
	if subnet == nil {
		return getResourceByIPFamily(ruleName, isDualStack, isIPv6)
	}

	// Load balancer rule name must be less or equal to 80 characters, so excluding the hyphen two segments cannot exceed 79
	subnetSegment := *subnet
	maxLength := consts.LoadBalancerRuleNameMaxLength - consts.IPFamilySuffixLength
	if len(ruleName)+len(subnetSegment)+1 > maxLength {
		subnetSegment = subnetSegment[:maxLength-len(ruleName)-1]
	}

	return getResourceByIPFamily(fmt.Sprintf("%s-%s-%s-%d", prefix, subnetSegment, protocol, port), isDualStack, isIPv6)
}

func (az *Cloud) getloadbalancerHAmodeRuleName(service *v1.Service, isIPv6 bool) string {
	return az.getLoadBalancerRuleName(service, service.Spec.Ports[0].Protocol, service.Spec.Ports[0].Port, isIPv6)
}

// This returns a human-readable version of the Service used to tag some resources.
// This is only used for human-readable convenience, and not to filter.
func getServiceName(service *v1.Service) string {
	return fmt.Sprintf("%s/%s", service.Namespace, service.Name)
}

// This returns a prefix for loadbalancer/security rules.
func (az *Cloud) getRulePrefix(service *v1.Service) string {
	return az.GetLoadBalancerName(context.TODO(), "", service)
}

func (az *Cloud) getPublicIPName(clusterName string, service *v1.Service, isIPv6 bool) (string, error) {
	logger := log.Background().WithName("getPublicIPName")
	isDualStack := isServiceDualStack(service)
	pipName := fmt.Sprintf("%s-%s", clusterName, az.GetLoadBalancerName(context.TODO(), clusterName, service))
	if id := getServicePIPPrefixID(service, isIPv6); id != "" {
		id, err := getLastSegment(id, "/")
		if err == nil {
			pipName = fmt.Sprintf("%s-%s", pipName, id)
		}
	}

	pipNameSegment := pipName
	maxLength := consts.PIPPrefixNameMaxLength - consts.IPFamilySuffixLength
	if len(pipName) > maxLength {
		pipNameSegment = pipNameSegment[:maxLength]
		logger.V(6).Info("original PIP name is lengthy, truncate it", "originalName", pipName, "truncatedName", pipNameSegment)
	}
	return getResourceByIPFamily(pipNameSegment, isDualStack, isIPv6), nil
}

func publicIPOwnsFrontendIP(service *v1.Service, fip *armnetwork.FrontendIPConfiguration, pip *armnetwork.PublicIPAddress) bool {
	logger := log.Background().WithName("publicIPOwnsFrontendIP")
	if pip != nil &&
		pip.ID != nil &&
		pip.Properties != nil &&
		pip.Properties.IPAddress != nil &&
		fip != nil &&
		fip.Properties != nil &&
		fip.Properties.PublicIPAddress != nil {
		if strings.EqualFold(ptr.Deref(pip.ID, ""), ptr.Deref(fip.Properties.PublicIPAddress.ID, "")) {
			logger.V(6).Info("found secondary service of the frontend IP config", "serviceName", service.Name, "fipName", *fip.Name)
			return true
		}
	}
	return false
}

// This returns the next available rule priority level for a given set of security rules.
func getNextAvailablePriority(rules []*armnetwork.SecurityRule) (int32, error) {
	var smallest int32 = consts.LoadBalancerMinimumPriority
	var spread int32 = 1

outer:
	for smallest < consts.LoadBalancerMaximumPriority {
		for _, rule := range rules {
			if *rule.Properties.Priority == smallest {
				smallest += spread
				continue outer
			}
		}
		// no one else had it
		return smallest, nil
	}

	return -1, fmt.Errorf("securityGroup priorities are exhausted")
}

var polyTable = crc32.MakeTable(crc32.Koopman)

// MakeCRC32 : convert string to CRC32 format
func MakeCRC32(str string) string {
	crc := crc32.New(polyTable)
	_, _ = crc.Write([]byte(str))
	hash := crc.Sum32()
	return strconv.FormatUint(uint64(hash), 10)
}

// availabilitySet implements VMSet interface for Azure availability sets.
type availabilitySet struct {
	*Cloud

	vmasCache azcache.Resource
}

type AvailabilitySetEntry struct {
	VMAS          *armcompute.AvailabilitySet
	ResourceGroup string
}

func (as *availabilitySet) newVMASCache() (azcache.Resource, error) {
	getter := func(ctx context.Context, _ string) (interface{}, error) {
		logger := log.FromContextOrBackground(ctx).WithName("newVMASCache")
		localCache := &sync.Map{}

		allResourceGroups, err := as.GetResourceGroups()
		if err != nil {
			return nil, err
		}

		for _, resourceGroup := range allResourceGroups.UnsortedList() {
			allAvailabilitySets, rerr := as.ComputeClientFactory.GetAvailabilitySetClient().List(ctx, resourceGroup)
			if rerr != nil {
				logger.Error(rerr, "AvailabilitySetsClient.List failed")
				return nil, rerr
			}

			for i := range allAvailabilitySets {
				vmas := allAvailabilitySets[i]
				if strings.EqualFold(ptr.Deref(vmas.Name, ""), "") {
					klog.Warning("failed to get the name of the VMAS")
					continue
				}
				localCache.Store(ptr.Deref(vmas.Name, ""), &AvailabilitySetEntry{
					VMAS:          vmas,
					ResourceGroup: resourceGroup,
				})
			}
		}

		return localCache, nil
	}

	if as.AvailabilitySetsCacheTTLInSeconds == 0 {
		as.AvailabilitySetsCacheTTLInSeconds = consts.VMASCacheTTLDefaultInSeconds
	}

	return azcache.NewTimedCache(time.Duration(as.AvailabilitySetsCacheTTLInSeconds)*time.Second, getter, as.DisableAPICallCache)
}

// RefreshCaches invalidates and renew all related caches.
func (as *availabilitySet) RefreshCaches() error {
	logger := log.Background().WithName("as.RefreshCaches")
	var err error
	as.vmasCache, err = as.newVMASCache()
	if err != nil {
		logger.Error(err, "failed to create or refresh VMAS cache")
		return err
	}
	return nil
}

// newStandardSet creates a new availabilitySet.
func newAvailabilitySet(az *Cloud) (VMSet, error) {
	as := &availabilitySet{
		Cloud: az,
	}

	if err := as.RefreshCaches(); err != nil {
		return nil, err
	}

	return as, nil
}

// GetInstanceIDByNodeName gets the cloud provider ID by node name.
// It must return ("", cloudprovider.InstanceNotFound) if the instance does
// not exist or is no longer running.
func (as *availabilitySet) GetInstanceIDByNodeName(ctx context.Context, name string) (string, error) {
	logger := log.FromContextOrBackground(ctx).WithName("GetInstanceIDByNodeName")
	var machine *armcompute.VirtualMachine
	var err error

	machine, err = as.getVirtualMachine(ctx, types.NodeName(name), azcache.CacheReadTypeUnsafe)
	if errors.Is(err, cloudprovider.InstanceNotFound) {
		klog.Warningf("Unable to find node %s: %v", name, cloudprovider.InstanceNotFound)
		return "", cloudprovider.InstanceNotFound
	}
	if err != nil {
		if as.CloudProviderBackoff {
			logger.V(2).Info("backing off", "node", name)
			machine, err = as.GetVirtualMachineWithRetry(ctx, types.NodeName(name), azcache.CacheReadTypeUnsafe)
			if err != nil {
				logger.V(2).Info("abort backoff", "node", name)
				return "", err
			}
		} else {
			return "", err
		}
	}

	resourceID := *machine.ID
	convertedResourceID, err := ConvertResourceGroupNameToLower(resourceID)
	if err != nil {
		logger.Error(err, "ConvertResourceGroupNameToLower failed")
		return "", err
	}
	return convertedResourceID, nil
}

// GetPowerStatusByNodeName returns the power state of the specified node.
func (as *availabilitySet) GetPowerStatusByNodeName(ctx context.Context, name string) (powerState string, err error) {
	logger := log.FromContextOrBackground(ctx).WithName("GetPowerStatusByNodeName")
	vm, err := as.getVirtualMachine(ctx, types.NodeName(name), azcache.CacheReadTypeDefault)
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

// GetProvisioningStateByNodeName returns the provisioningState for the specified node.
func (as *availabilitySet) GetProvisioningStateByNodeName(ctx context.Context, name string) (provisioningState string, err error) {
	vm, err := as.getVirtualMachine(ctx, types.NodeName(name), azcache.CacheReadTypeDefault)
	if err != nil {
		return provisioningState, err
	}

	if vm.Properties == nil || vm.Properties.ProvisioningState == nil {
		return provisioningState, nil
	}

	return ptr.Deref(vm.Properties.ProvisioningState, ""), nil
}

// GetNodeNameByProviderID gets the node name by provider ID.
func (as *availabilitySet) GetNodeNameByProviderID(_ context.Context, providerID string) (types.NodeName, error) {
	// NodeName is part of providerID for standard instances.
	matches := providerIDRE.FindStringSubmatch(providerID)
	if len(matches) != 2 {
		return "", errors.New("error splitting providerID")
	}

	return types.NodeName(matches[1]), nil
}

// GetInstanceTypeByNodeName gets the instance type by node name.
func (as *availabilitySet) GetInstanceTypeByNodeName(ctx context.Context, name string) (string, error) {
	logger := log.FromContextOrBackground(ctx).WithName("GetInstanceTypeByNodeName")
	machine, err := as.getVirtualMachine(ctx, types.NodeName(name), azcache.CacheReadTypeUnsafe)
	if err != nil {
		logger.Error(err, "as.getVirtualMachine failed", "node", name)
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
func (as *availabilitySet) GetZoneByNodeName(ctx context.Context, name string) (cloudprovider.Zone, error) {
	vm, err := as.getVirtualMachine(ctx, types.NodeName(name), azcache.CacheReadTypeUnsafe)
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

		failureDomain = as.makeZone(ptr.Deref(vm.Location, ""), zoneID)
	} else {
		// Availability zone is not used for the node, falling back to fault domain.
		if prop := vm.Properties; prop == nil || prop.InstanceView == nil {
			failureDomain = "0"
		} else {
			failureDomain = strconv.Itoa(int(ptr.Deref(vm.Properties.InstanceView.PlatformFaultDomain, 0)))
		}
	}

	zone := cloudprovider.Zone{
		FailureDomain: strings.ToLower(failureDomain),
		Region:        strings.ToLower(ptr.Deref(vm.Location, "")),
	}
	return zone, nil
}

// GetPrimaryVMSetName returns the VM set name depending on the configured vmType.
// It returns config.PrimaryScaleSetName for vmss and config.PrimaryAvailabilitySetName for standard vmType.
func (as *availabilitySet) GetPrimaryVMSetName() string {
	return as.PrimaryAvailabilitySetName
}

// GetIPByNodeName gets machine private IP and public IP by node name.
func (as *availabilitySet) GetIPByNodeName(ctx context.Context, name string) (string, string, error) {
	logger := log.FromContextOrBackground(ctx).WithName("GetIPByNodeName")
	nic, err := as.GetPrimaryInterface(ctx, name)
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
		pip, existsPip, err := as.getPublicIPAddress(ctx, as.ResourceGroup, pipName, azcache.CacheReadTypeDefault)
		if err != nil {
			return "", "", err
		}
		if existsPip {
			publicIP = *pip.Properties.IPAddress
		}
	}

	return privateIP, publicIP, nil
}

// returns a list of private ips assigned to node
// TODO (khenidak): This should read all nics, not just the primary
// allowing users to split ipv4/v6 on multiple nics
func (as *availabilitySet) GetPrivateIPsByNodeName(ctx context.Context, name string) ([]string, error) {
	ips := make([]string, 0)
	nic, err := as.GetPrimaryInterface(ctx, name)
	if err != nil {
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

// getAgentPoolAvailabilitySets lists the virtual machines for the resource group and then builds
// a list of availability sets that match the nodes available to k8s.
func (as *availabilitySet) getAgentPoolAvailabilitySets(vms []*armcompute.VirtualMachine, nodes []*v1.Node) (agentPoolAvailabilitySets []*string, err error) {
	logger := log.Background().WithName("as.getAgentPoolAvailabilitySets")
	vmNameToAvailabilitySetID := make(map[string]string, len(vms))
	for vmx := range vms {
		vm := vms[vmx]
		if vm.Properties.AvailabilitySet != nil {
			vmNameToAvailabilitySetID[*vm.Name] = *vm.Properties.AvailabilitySet.ID
		}
	}
	agentPoolAvailabilitySets = []*string{}
	for nx := range nodes {
		nodeName := (*nodes[nx]).Name
		if isControlPlaneNode(nodes[nx]) {
			continue
		}
		asID, ok := vmNameToAvailabilitySetID[nodeName]
		if !ok {
			klog.Warningf("as.getNodeAvailabilitySet - Node(%s) has no availability sets", nodeName)
			continue
		}
		asName, err := getLastSegment(asID, "/")
		if err != nil {
			logger.Error(err, "Node getLastSegment failed", "node", nodeName, "asID", asID)
			return nil, err
		}
		// AvailabilitySet ID is currently upper cased in a non-deterministic way
		// We want to keep it lower case, before the ID get fixed
		asName = strings.ToLower(asName)

		agentPoolAvailabilitySets = append(agentPoolAvailabilitySets, &asName)
	}

	return agentPoolAvailabilitySets, nil
}

// GetVMSetNames selects all possible availability sets or scale sets
// (depending vmType configured) for service load balancer, if the service has
// no loadbalancer mode annotation returns the primary VMSet. If service annotation
// for loadbalancer exists then returns the eligible VMSet. The mode selection
// annotation would be ignored when using one SLB per cluster.
func (as *availabilitySet) GetVMSetNames(ctx context.Context, service *v1.Service, nodes []*v1.Node) (availabilitySetNames []*string, err error) {
	logger := log.FromContextOrBackground(ctx).WithName("as.GetVMSetNames")
	hasMode, isAuto, serviceAvailabilitySetName := as.getServiceLoadBalancerMode(service)
	if !hasMode || as.UseStandardLoadBalancer() {
		// no mode specified in service annotation or use single SLB mode
		// default to PrimaryAvailabilitySetName
		availabilitySetNames = []*string{to.Ptr(as.PrimaryAvailabilitySetName)}
		return availabilitySetNames, nil
	}

	vms, err := as.ListVirtualMachines(ctx, as.ResourceGroup)
	if err != nil {
		logger.Error(err, "as.getNodeAvailabilitySet - ListVirtualMachines failed")
		return nil, err
	}
	availabilitySetNames, err = as.getAgentPoolAvailabilitySets(vms, nodes)
	if err != nil {
		logger.Error(err, "as.getAgentPoolAvailabilitySets failed")
		return nil, err
	}
	if len(availabilitySetNames) == 0 {
		err = fmt.Errorf("no availability sets found for nodes, node count(%d)", len(nodes))
		logger.Error(err, "no availability sets found for nodes in the cluster", "nodeCount", len(nodes))
		return nil, err
	}
	if !isAuto {
		found := false
		for asx := range availabilitySetNames {
			if strings.EqualFold(*availabilitySetNames[asx], serviceAvailabilitySetName) {
				found = true
				break
			}
		}
		if !found {
			err = fmt.Errorf("availability set (%s) - not found", serviceAvailabilitySetName)
			logger.Error(err, "availability set in service annotation not found", "availabilitySetName", serviceAvailabilitySetName)
			return nil, err
		}
		return []*string{to.Ptr(serviceAvailabilitySetName)}, nil
	}

	return availabilitySetNames, nil
}

func (as *availabilitySet) GetNodeVMSetName(ctx context.Context, node *v1.Node) (string, error) {
	logger := log.FromContextOrBackground(ctx).WithName("as.GetNodeVMSetName")
	var hostName string
	for _, nodeAddress := range node.Status.Addresses {
		if strings.EqualFold(string(nodeAddress.Type), string(v1.NodeHostName)) {
			hostName = nodeAddress.Address
		}
	}
	if hostName == "" {
		if name, ok := node.Labels[consts.NodeLabelHostName]; ok {
			hostName = name
		}
	}
	if hostName == "" {
		klog.Warningf("as.GetNodeVMSetName: cannot get host name from node %s", node.Name)
		return "", nil
	}

	vms, err := as.ListVirtualMachines(ctx, as.ResourceGroup)
	if err != nil {
		logger.Error(err, "failed: ListVirtualMachines")
		return "", err
	}

	var asName string
	for _, vm := range vms {
		if strings.EqualFold(ptr.Deref(vm.Name, ""), hostName) {
			if vm.Properties.AvailabilitySet != nil && ptr.Deref(vm.Properties.AvailabilitySet.ID, "") != "" {
				logger.V(4).Info("found vm", "vm", hostName)

				asName, err = getLastSegment(ptr.Deref(vm.Properties.AvailabilitySet.ID, ""), "/")
				if err != nil {
					logger.Error(err, "failed to get last segment of ID", "ID", ptr.Deref(vm.Properties.AvailabilitySet.ID, ""))
					return "", err
				}
			}

			break
		}
	}

	logger.V(4).Info("found availability set name from node name", "set name", asName, "node name", node.Name)
	return asName, nil
}

// GetPrimaryInterface gets machine primary network interface by node name.
func (as *availabilitySet) GetPrimaryInterface(ctx context.Context, nodeName string) (*armnetwork.Interface, error) {
	nic, _, err := as.getPrimaryInterfaceWithVMSet(ctx, nodeName, "")
	return nic, err
}

// extractResourceGroupByNicID extracts the resource group name by nicID.
func extractResourceGroupByNicID(nicID string) (string, error) {
	matches := nicResourceGroupRE.FindStringSubmatch(nicID)
	if len(matches) != 2 {
		return "", fmt.Errorf("error of extracting resourceGroup from nicID %q", nicID)
	}

	return matches[1], nil
}

// getPrimaryInterfaceWithVMSet gets machine primary network interface by node name and vmSet.
func (as *availabilitySet) getPrimaryInterfaceWithVMSet(ctx context.Context, nodeName, vmSetName string) (*armnetwork.Interface, string, error) {
	logger := log.FromContextOrBackground(ctx).WithName("getPrimaryInterfaceWithVMSet")
	var machine *armcompute.VirtualMachine

	machine, err := as.GetVirtualMachineWithRetry(ctx, types.NodeName(nodeName), azcache.CacheReadTypeDefault)
	if err != nil {
		logger.V(2).Info("abort backoff", "nodeName", nodeName, "vmSetName", vmSetName)
		return nil, "", err
	}

	primaryNicID, err := getPrimaryInterfaceID(machine)
	if err != nil {
		return nil, "", err
	}
	nicName, err := getLastSegment(primaryNicID, "/")
	if err != nil {
		return nil, "", err
	}
	nodeResourceGroup, err := as.GetNodeResourceGroup(nodeName)
	if err != nil {
		return nil, "", err
	}

	// Check availability set name. Note that vmSetName is empty string when getting
	// the Node's IP address. While vmSetName is not empty, it should be checked with
	// Node's real availability set name:
	// - For basic SKU load balancer, errNotInVMSet should be returned if the node's
	//   availability set is mismatched with vmSetName.
	// - For single standard SKU load balancer, backend could belong to multiple VMAS, so we
	//   don't check vmSet for it.
	// - For multiple standard SKU load balancers, the behavior is similar to the basic LB.
	needCheck := !as.UseStandardLoadBalancer()

	if vmSetName != "" && needCheck {
		expectedAvailabilitySetID := as.getAvailabilitySetID(nodeResourceGroup, vmSetName)
		if machine.Properties.AvailabilitySet == nil || !strings.EqualFold(*machine.Properties.AvailabilitySet.ID, expectedAvailabilitySetID) {
			logger.V(3).Info("nic is not in the availabilitySet", "nic", nicName, "availabilitySet", vmSetName)
			return nil, "", errNotInVMSet
		}
	}

	nicResourceGroup, err := extractResourceGroupByNicID(primaryNicID)
	if err != nil {
		return nil, "", err
	}

	ctx, cancel := getContextWithCancel()
	defer cancel()
	nic, rerr := as.ComputeClientFactory.GetInterfaceClient().Get(ctx, nicResourceGroup, nicName, nil)
	if rerr != nil {
		return nil, "", rerr
	}

	var availabilitySetID string
	if machine.Properties != nil && machine.Properties.AvailabilitySet != nil {
		availabilitySetID = ptr.Deref(machine.Properties.AvailabilitySet.ID, "")
	}
	return nic, availabilitySetID, nil
}

// EnsureHostInPool ensures the given VM's Primary NIC's Primary IP Configuration is
// participating in the specified LoadBalancer Backend Pool.
func (as *availabilitySet) EnsureHostInPool(ctx context.Context, service *v1.Service, nodeName types.NodeName, backendPoolID string, vmSetName string) (string, string, string, *armcompute.VirtualMachineScaleSetVM, error) {
	logger := log.FromContextOrBackground(ctx).WithName("EnsureHostInPool")
	vmName := mapNodeNameToVMName(nodeName)
	serviceName := getServiceName(service)
	nic, _, err := as.getPrimaryInterfaceWithVMSet(ctx, vmName, vmSetName)
	if err != nil {
		if errors.Is(err, errNotInVMSet) {
			logger.V(3).Info("skips node because it is not in the vmSet", "node", nodeName, "vmSet", vmSetName)
			return "", "", "", nil, nil
		}

		logger.Error(err, "az.VMSet.GetPrimaryInterface.Get() failed", "nodeName", nodeName, "vmName", vmName, "vmSetName", vmSetName)
		return "", "", "", nil, err
	}

	if nic != nil && nic.Properties != nil && nic.Properties.ProvisioningState != nil && *nic.Properties.ProvisioningState == armnetwork.ProvisioningStateFailed {
		klog.Warningf("EnsureHostInPool skips node %s because its primary nic %s is in Failed state", nodeName, *nic.Name)
		return "", "", "", nil, nil
	}

	var primaryIPConfig *armnetwork.InterfaceIPConfiguration
	ipv6 := isBackendPoolIPv6(backendPoolID)
	if !as.ipv6DualStackEnabled && !ipv6 {
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
	if !foundPool {
		if as.UseStandardLoadBalancer() && len(newBackendPools) > 0 {
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
				logger.V(4).Info("Node has already been added to LB, omit adding it to a new one", "node", nodeName, "oldLBName", oldLBName)
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
		err := as.CreateOrUpdateInterface(ctx, service, nic)
		if err != nil {
			return "", "", "", nil, err
		}
	}
	return "", "", "", nil, nil
}

// EnsureHostsInPool ensures the given Node's primary IP configurations are
// participating in the specified LoadBalancer Backend Pool.
func (as *availabilitySet) EnsureHostsInPool(ctx context.Context, service *v1.Service, nodes []*v1.Node, backendPoolID string, vmSetName string) error {
	logger := log.FromContextOrBackground(ctx).WithName("EnsureHostsInPool")
	mc := metrics.NewMetricContext("services", "vmas_ensure_hosts_in_pool", as.ResourceGroup, as.SubscriptionID, getServiceName(service))
	isOperationSucceeded := false
	defer func() {
		mc.ObserveOperationWithResult(isOperationSucceeded)
	}()

	hostUpdates := make([]func() error, 0, len(nodes))
	for _, node := range nodes {
		localNodeName := node.Name
		if as.UseStandardLoadBalancer() && as.ExcludeMasterNodesFromStandardLB() && isControlPlaneNode(node) {
			logger.V(4).Info("Excluding master node from load balancer backendpool", "node", localNodeName, "backendpool", backendPoolID)
			continue
		}

		shouldExcludeLoadBalancer, err := as.ShouldNodeExcludedFromLoadBalancer(localNodeName)
		if err != nil {
			logger.Error(err, "ShouldNodeExcludedFromLoadBalancer failed", "node", localNodeName)
			return err
		}
		if shouldExcludeLoadBalancer {
			logger.V(4).Info("Excluding unmanaged/external-resource-group node", "node", localNodeName)
			continue
		}

		f := func() error {
			_, _, _, _, err := as.EnsureHostInPool(ctx, service, types.NodeName(localNodeName), backendPoolID, vmSetName)
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

// EnsureBackendPoolDeleted ensures the loadBalancer backendAddressPools deleted from the specified nodes.
// backendPoolIDs are the IDs of the backendpools to be deleted.
func (as *availabilitySet) EnsureBackendPoolDeleted(ctx context.Context, service *v1.Service, backendPoolIDs []string, vmSetName string, backendAddressPools []*armnetwork.BackendAddressPool, _ bool) (bool, error) {
	logger := log.FromContextOrBackground(ctx).WithName("az.EnsureBackendPoolDeleted")
	// Returns nil if backend address pools already deleted.
	if backendAddressPools == nil {
		return false, nil
	}

	mc := metrics.NewMetricContext("services", "vmas_ensure_backend_pool_deleted", as.ResourceGroup, as.SubscriptionID, getServiceName(service))
	isOperationSucceeded := false
	defer func() {
		mc.ObserveOperationWithResult(isOperationSucceeded)
	}()

	ipConfigurationIDs := []string{}
	for _, backendPool := range backendAddressPools {
		for _, backendPoolID := range backendPoolIDs {
			if strings.EqualFold(ptr.Deref(backendPool.ID, ""), backendPoolID) {
				if backendPool.Properties != nil &&
					backendPool.Properties.BackendIPConfigurations != nil {
					for _, ipConf := range backendPool.Properties.BackendIPConfigurations {
						if ipConf.ID == nil {
							continue
						}

						ipConfigurationIDs = append(ipConfigurationIDs, *ipConf.ID)
					}
				}
			}
		}
	}
	nicUpdaters := make([]func() error, 0)
	allErrs := make([]error, 0)

	ipconfigPrefixToNicMap := map[string]*armnetwork.Interface{} // ipconfig prefix -> nic
	for i := range ipConfigurationIDs {
		ipConfigurationID := ipConfigurationIDs[i]
		ipConfigIDPrefix := getResourceIDPrefix(ipConfigurationID)
		if _, ok := ipconfigPrefixToNicMap[ipConfigIDPrefix]; ok {
			continue
		}
		nodeName, _, err := as.GetNodeNameByIPConfigurationID(ctx, ipConfigurationID)
		if err != nil && !errors.Is(err, cloudprovider.InstanceNotFound) {
			logger.Error(err, "Failed to GetNodeNameByIPConfigurationID", "ipConfigurationID", ipConfigurationID)
			allErrs = append(allErrs, err)
			continue
		}
		if nodeName == "" {
			continue
		}

		vmName := mapNodeNameToVMName(types.NodeName(nodeName))
		nic, vmasID, err := as.getPrimaryInterfaceWithVMSet(ctx, vmName, vmSetName)
		if err != nil {
			if errors.Is(err, errNotInVMSet) {
				logger.V(3).Info("skips node because it is not in the vmSet", "node", nodeName, "vmSet", vmSetName)
				return false, nil
			}

			logger.Error(err, "az.VMSet.GetPrimaryInterface.Get() failed", "nodeName", nodeName, "vmName", vmName, "vmSetName", vmSetName)
			return false, err
		}
		vmasName, err := getAvailabilitySetNameByID(vmasID)
		if err != nil {
			return false, fmt.Errorf("EnsureBackendPoolDeleted: failed to parse the VMAS ID %s: %w", vmasID, err)
		}
		// Only remove nodes belonging to specified vmSet to basic LB backends.
		// If vmasID is empty, then it is standalone VM.
		if vmasID != "" && !strings.EqualFold(vmasName, vmSetName) {
			logger.V(2).Info("skipping the node belonging to another vm set", "node", nodeName, "vmSet", vmasName)
			continue
		}

		if *nic.Properties.ProvisioningState == consts.NicFailedState {
			klog.Warningf("EnsureBackendPoolDeleted skips node %s because its primary nic %s is in Failed state", nodeName, *nic.Name)
			return false, nil
		}

		if nic.Properties != nil && nic.Properties.IPConfigurations != nil {
			ipconfigPrefixToNicMap[ipConfigIDPrefix] = nic
		}
	}
	v4Enabled, v6Enabled := getIPFamiliesEnabled(service)
	isServiceIPv4 := v4Enabled && !v6Enabled
	var nicUpdated atomic.Bool
	for k := range ipconfigPrefixToNicMap {
		nic := ipconfigPrefixToNicMap[k]
		newIPConfigs := nic.Properties.IPConfigurations
		for j, ipConf := range newIPConfigs {
			if isServiceIPv4 && !ptr.Deref(ipConf.Properties.Primary, false) {
				continue
			}
			// To support IPv6 only and dual-stack clusters, all IP configurations
			// should be checked regardless of primary or not because IPv6 IP configurations
			// are not marked as primary.
			if ipConf.Properties.LoadBalancerBackendAddressPools != nil {
				newLBAddressPools := ipConf.Properties.LoadBalancerBackendAddressPools
				for k := len(newLBAddressPools) - 1; k >= 0; k-- {
					pool := newLBAddressPools[k]
					for _, backendPoolID := range backendPoolIDs {
						if strings.EqualFold(ptr.Deref(pool.ID, ""), backendPoolID) {
							newLBAddressPools = append(newLBAddressPools[:k], newLBAddressPools[k+1:]...)
							break
						}
					}
				}
				newIPConfigs[j].Properties.LoadBalancerBackendAddressPools = newLBAddressPools
			}
		}
		nic.Properties.IPConfigurations = newIPConfigs
		nicUpdaters = append(nicUpdaters, func() error {
			logger.V(2).Info("begins to CreateOrUpdate for NIC with backendPoolIDs", "resourceGroup", as.ResourceGroup, "nicName", ptr.Deref(nic.Name, ""), "backendPoolIDs", backendPoolIDs)
			_, rerr := as.ComputeClientFactory.GetInterfaceClient().CreateOrUpdate(ctx, as.ResourceGroup, ptr.Deref(nic.Name, ""), *nic)
			if rerr != nil {
				logger.Error(rerr, "CreateOrUpdate for NIC failed", "resourceGroup", as.ResourceGroup, "nicName", ptr.Deref(nic.Name, ""))
				return rerr
			}
			nicUpdated.Store(true)
			return nil
		})
	}
	errs := utilerrors.AggregateGoroutines(nicUpdaters...)
	if errs != nil {
		return nicUpdated.Load(), utilerrors.Flatten(errs)
	}
	// Fail if there are other errors.
	if len(allErrs) > 0 {
		return nicUpdated.Load(), utilerrors.Flatten(utilerrors.NewAggregate(allErrs))
	}

	isOperationSucceeded = true
	return nicUpdated.Load(), nil
}

func getAvailabilitySetNameByID(asID string) (string, error) {
	// for standalone VM
	if asID == "" {
		return "", nil
	}

	matches := vmasIDRE.FindStringSubmatch(asID)
	if len(matches) != 2 {
		return "", fmt.Errorf("getAvailabilitySetNameByID: failed to parse the VMAS ID %s", asID)
	}
	vmasName := matches[1]
	return vmasName, nil
}

// GetNodeNameByIPConfigurationID gets the node name and the availabilitySet name by IP configuration ID.
func (as *availabilitySet) GetNodeNameByIPConfigurationID(ctx context.Context, ipConfigurationID string) (string, string, error) {
	logger := log.FromContextOrBackground(ctx).WithName("GetNodeNameByIPConfigurationID")
	matches := nicIDRE.FindStringSubmatch(ipConfigurationID)
	if len(matches) != 3 {
		logger.V(4).Info("Can not extract VM name from ipConfigurationID", "ipConfigurationID", ipConfigurationID)
		return "", "", fmt.Errorf("invalid ip config ID %s", ipConfigurationID)
	}

	nicResourceGroup, nicName := matches[1], matches[2]
	if nicResourceGroup == "" || nicName == "" {
		return "", "", fmt.Errorf("invalid ip config ID %s", ipConfigurationID)
	}
	nic, rerr := as.ComputeClientFactory.GetInterfaceClient().Get(ctx, nicResourceGroup, nicName, nil)
	if rerr != nil {
		return "", "", fmt.Errorf("GetNodeNameByIPConfigurationID(%s): failed to get interface of name %s: %w", ipConfigurationID, nicName, rerr)
	}
	vmID := ""
	if nic.Properties != nil && nic.Properties.VirtualMachine != nil {
		vmID = ptr.Deref(nic.Properties.VirtualMachine.ID, "")
	}
	if vmID == "" {
		logger.V(2).Info("empty vmID", "ipConfigurationID", ipConfigurationID)
		return "", "", nil
	}

	matches = vmIDRE.FindStringSubmatch(vmID)
	if len(matches) != 2 {
		return "", "", fmt.Errorf("invalid virtual machine ID %s", vmID)
	}
	vmName := matches[1]

	vm, err := as.getVirtualMachine(ctx, types.NodeName(vmName), azcache.CacheReadTypeDefault)
	if err != nil {
		logger.Error(err, "Unable to get the virtual machine by node name", "name", vmName)
		return "", "", err
	}
	asID := ""
	if vm.Properties != nil && vm.Properties.AvailabilitySet != nil {
		asID = ptr.Deref(vm.Properties.AvailabilitySet.ID, "")
	}
	if asID == "" {
		return vmName, "", nil
	}

	asName, err := getAvailabilitySetNameByID(asID)
	if err != nil {
		return "", "", fmt.Errorf("cannot get the availability set name by the availability set ID %s: %w", asID, err)
	}
	return vmName, strings.ToLower(asName), nil
}

func (as *availabilitySet) getAvailabilitySetByNodeName(ctx context.Context, nodeName string, crt azcache.AzureCacheReadType) (*armcompute.AvailabilitySet, error) {
	cached, err := as.vmasCache.Get(ctx, consts.VMASKey, crt)
	if err != nil {
		return nil, err
	}
	vmasList := cached.(*sync.Map)

	if vmasList == nil {
		klog.Warning("Couldn't get all vmas from cache")
		return nil, nil
	}

	var result *armcompute.AvailabilitySet
	vmasList.Range(func(_, value interface{}) bool {
		vmasEntry := value.(*AvailabilitySetEntry)
		vmas := vmasEntry.VMAS
		if vmas != nil && vmas.Properties != nil && vmas.Properties.VirtualMachines != nil {
			for _, vmIDRef := range vmas.Properties.VirtualMachines {
				if vmIDRef.ID != nil {
					matches := vmIDRE.FindStringSubmatch(ptr.Deref(vmIDRef.ID, ""))
					if len(matches) != 2 {
						err = fmt.Errorf("invalid vm ID %s", ptr.Deref(vmIDRef.ID, ""))
						return false
					}

					vmName := matches[1]
					if strings.EqualFold(vmName, nodeName) {
						result = vmas
						return false
					}
				}
			}
		}

		return true
	})

	if err != nil {
		return nil, err
	}

	if result == nil {
		klog.Warningf("Unable to find node %s: %v", nodeName, cloudprovider.InstanceNotFound)
		return nil, cloudprovider.InstanceNotFound
	}

	return result, nil
}

// GetNodeCIDRMaskByProviderID returns the node CIDR subnet mask by provider ID.
func (as *availabilitySet) GetNodeCIDRMasksByProviderID(ctx context.Context, providerID string) (int, int, error) {
	logger := log.FromContextOrBackground(ctx).WithName("GetNodeCIDRMasksByProviderID")
	nodeName, err := as.GetNodeNameByProviderID(ctx, providerID)
	if err != nil {
		return 0, 0, err
	}

	vmas, err := as.getAvailabilitySetByNodeName(ctx, string(nodeName), azcache.CacheReadTypeDefault)
	if err != nil {
		if errors.Is(err, cloudprovider.InstanceNotFound) {
			return consts.DefaultNodeMaskCIDRIPv4, consts.DefaultNodeMaskCIDRIPv6, nil
		}
		return 0, 0, err
	}

	var ipv4Mask, ipv6Mask int
	if v4, ok := vmas.Tags[consts.VMSetCIDRIPV4TagKey]; ok && v4 != nil {
		ipv4Mask, err = strconv.Atoi(ptr.Deref(v4, ""))
		if err != nil {
			logger.Error(err, "error when parsing the value of the ipv4 mask size", "value", ptr.Deref(v4, ""))
		}
	}
	if v6, ok := vmas.Tags[consts.VMSetCIDRIPV6TagKey]; ok && v6 != nil {
		ipv6Mask, err = strconv.Atoi(ptr.Deref(v6, ""))
		if err != nil {
			logger.Error(err, "error when parsing the value of the ipv6 mask size", "value", ptr.Deref(v6, ""))
		}
	}

	return ipv4Mask, ipv6Mask, nil
}

// EnsureBackendPoolDeletedFromVMSets ensures the loadBalancer backendAddressPools deleted from the specified VMAS
func (as *availabilitySet) EnsureBackendPoolDeletedFromVMSets(_ context.Context, _ map[string]bool, _ []string) error {
	return nil
}

// GetAgentPoolVMSetNames returns all VMAS names according to the nodes
func (as *availabilitySet) GetAgentPoolVMSetNames(ctx context.Context, nodes []*v1.Node) ([]*string, error) {
	logger := log.FromContextOrBackground(ctx).WithName("as.GetAgentPoolVMSetNames")
	vms, err := as.ListVirtualMachines(ctx, as.ResourceGroup)
	if err != nil {
		logger.Error(err, "failed: ListVirtualMachines")
		return nil, err
	}

	return as.getAgentPoolAvailabilitySets(vms, nodes)
}
