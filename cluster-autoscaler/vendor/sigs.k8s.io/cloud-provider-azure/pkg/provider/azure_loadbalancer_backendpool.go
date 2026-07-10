/*
Copyright 2021 The Kubernetes Authors.

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

//go:generate sh -c "mockgen -destination=$GOPATH/src/sigs.k8s.io/cloud-provider-azure/pkg/provider/azure_mock_loadbalancer_backendpool.go -source=$GOPATH/src/sigs.k8s.io/cloud-provider-azure/pkg/provider/azure_loadbalancer_backendpool.go -package=provider BackendPool"

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v9"
	v1 "k8s.io/api/core/v1"
	cloudprovider "k8s.io/cloud-provider"
	"k8s.io/klog/v2"
	utilnet "k8s.io/utils/net"
	"k8s.io/utils/ptr"

	"sigs.k8s.io/cloud-provider-azure/pkg/cache"
	"sigs.k8s.io/cloud-provider-azure/pkg/consts"
	"sigs.k8s.io/cloud-provider-azure/pkg/log"
	"sigs.k8s.io/cloud-provider-azure/pkg/metrics"
	utilsets "sigs.k8s.io/cloud-provider-azure/pkg/util/sets"
)

type BackendPool interface {
	// EnsureHostsInPool ensures the nodes join the backend pool of the load balancer
	EnsureHostsInPool(ctx context.Context, service *v1.Service, nodes []*v1.Node, backendPoolID, vmSetName, clusterName, lbName string, backendPool *armnetwork.BackendAddressPool) error

	// ReconcileBackendPools creates the inbound backend pool if it is not existed, and removes nodes that are supposed to be
	// excluded from the load balancers.
	ReconcileBackendPools(ctx context.Context, clusterName string, service *v1.Service, lb *armnetwork.LoadBalancer) (bool, bool, *armnetwork.LoadBalancer, error)

	// GetBackendPrivateIPs returns the private IPs of LoadBalancer's backend pool
	GetBackendPrivateIPs(ctx context.Context, clusterName string, service *v1.Service, lb *armnetwork.LoadBalancer) ([]string, []string)
}

type backendPoolTypeNodeIPConfig struct {
	*Cloud
}

func newBackendPoolTypeNodeIPConfig(c *Cloud) BackendPool {
	return &backendPoolTypeNodeIPConfig{c}
}

func (bc *backendPoolTypeNodeIPConfig) EnsureHostsInPool(ctx context.Context, service *v1.Service, nodes []*v1.Node, backendPoolID, vmSetName, _, _ string, _ *armnetwork.BackendAddressPool) error {
	return bc.VMSet.EnsureHostsInPool(ctx, service, nodes, backendPoolID, vmSetName)
}

func isLBBackendPoolsExisting(lbBackendPoolNames map[bool]string, bpName *string) (found, isIPv6 bool) {
	if strings.EqualFold(ptr.Deref(bpName, ""), lbBackendPoolNames[consts.IPVersionIPv4]) {
		isIPv6 = false
		found = true
	}
	if strings.EqualFold(ptr.Deref(bpName, ""), lbBackendPoolNames[consts.IPVersionIPv6]) {
		isIPv6 = true
		found = true
	}
	return found, isIPv6
}

func (bc *backendPoolTypeNodeIPConfig) ReconcileBackendPools(
	ctx context.Context,
	clusterName string,
	service *v1.Service,
	lb *armnetwork.LoadBalancer,
) (bool, bool, *armnetwork.LoadBalancer, error) {
	logger := log.FromContextOrBackground(ctx).WithName("bc.ReconcileBackendPools")
	var newBackendPools []*armnetwork.BackendAddressPool
	var err error
	if lb.Properties.BackendAddressPools != nil {
		newBackendPools = lb.Properties.BackendAddressPools
	}

	var backendPoolsCreated, backendPoolsUpdated, isOperationSucceeded, isMigration bool
	foundBackendPools := map[bool]bool{}
	lbName := *lb.Name

	serviceName := getServiceName(service)
	lbBackendPoolNames := getBackendPoolNames(clusterName)
	lbBackendPoolIDs := bc.getBackendPoolIDs(clusterName, lbName)
	vmSetName := bc.mapLoadBalancerNameToVMSet(lbName, clusterName)
	isBackendPoolPreConfigured := bc.isBackendPoolPreConfigured(service)

	mc := metrics.NewMetricContext("services", "migrate_to_nic_based_backend_pool", bc.ResourceGroup, bc.getNetworkResourceSubscriptionID(), serviceName)

	backendpoolToBeDeleted := []*armnetwork.BackendAddressPool{}
	lbBackendPoolIDsSlice := []string{}
	for i := len(newBackendPools) - 1; i >= 0; i-- {
		bp := newBackendPools[i]
		found, isIPv6 := isLBBackendPoolsExisting(lbBackendPoolNames, bp.Name)
		if found {
			logger.V(10).Info("lb backendpool - found wanted backendpool. not adding anything", "serviceName", serviceName)
			foundBackendPools[isBackendPoolIPv6(ptr.Deref(bp.Name, ""))] = true

			// Don't bother to remove unused nodeIPConfiguration if backend pool is pre configured
			if isBackendPoolPreConfigured {
				break
			}

			// If the LB backend pool type is configured from nodeIP or podIP
			// to nodeIPConfiguration, we need to decouple the VM NICs from the LB
			// before attaching nodeIPs/podIPs to the LB backend pool.
			if bp.Properties != nil &&
				bp.Properties.LoadBalancerBackendAddresses != nil &&
				len(bp.Properties.LoadBalancerBackendAddresses) > 0 {
				if removeNodeIPAddressesFromBackendPool(bp, []string{}, true, false, false) {
					isMigration = true
					bp.Properties.VirtualNetwork = nil
					if err := bc.CreateOrUpdateLBBackendPool(ctx, lbName, bp); err != nil {
						logger.Error(err, "failed to cleanup IP based backend pool", "service", serviceName, "backendPoolName", lbBackendPoolNames[isIPv6])
						return false, false, nil, fmt.Errorf("bc.ReconcileBackendPools for service (%s): failed to cleanup IP based backend pool %s: %w", serviceName, lbBackendPoolNames[isIPv6], err)
					}
					newBackendPools[i] = bp
					lb.Properties.BackendAddressPools = newBackendPools
					backendPoolsUpdated = true
				}
			}

			var backendIPConfigurationsToBeDeleted, bipConfigNotFound, bipConfigExclude []*armnetwork.InterfaceIPConfiguration
			if bp.Properties != nil && bp.Properties.BackendIPConfigurations != nil {
				for _, ipConf := range bp.Properties.BackendIPConfigurations {
					ipConfID := ptr.Deref(ipConf.ID, "")
					nodeName, _, err := bc.VMSet.GetNodeNameByIPConfigurationID(ctx, ipConfID)
					if err != nil {
						if errors.Is(err, cloudprovider.InstanceNotFound) {
							logger.V(2).Info("vm not found for ipConfID", "serviceName", serviceName, "ipConfID", ipConfID)
							bipConfigNotFound = append(bipConfigNotFound, ipConf)
						} else {
							return false, false, nil, err
						}
					}

					// If a node is not supposed to be included in the LB, it
					// would not be in the `nodes` slice. We need to check the nodes that
					// have been added to the LB's backendpool, find the unwanted ones and
					// delete them from the pool.
					shouldExcludeLoadBalancer, err := bc.ShouldNodeExcludedFromLoadBalancer(nodeName)
					if err != nil {
						logger.Error(err, "bc.ShouldNodeExcludedFromLoadBalancer failed", "node", nodeName)
						return false, false, nil, err
					}
					if shouldExcludeLoadBalancer {
						logger.V(2).Info("lb backendpool - found unwanted node, decouple it from the LB", "serviceName", serviceName, "nodeName", nodeName, "lbName", lbName)
						// construct a backendPool that only contains the IP config of the node to be deleted
						bipConfigExclude = append(bipConfigExclude, &armnetwork.InterfaceIPConfiguration{ID: ptr.To(ipConfID)})
					}
				}
			}
			backendIPConfigurationsToBeDeleted = getBackendIPConfigurationsToBeDeleted(*bp, bipConfigNotFound, bipConfigExclude)
			if len(backendIPConfigurationsToBeDeleted) > 0 {
				backendpoolToBeDeleted = append(backendpoolToBeDeleted, &armnetwork.BackendAddressPool{
					ID: ptr.To(lbBackendPoolIDs[isIPv6]),
					Properties: &armnetwork.BackendAddressPoolPropertiesFormat{
						BackendIPConfigurations: backendIPConfigurationsToBeDeleted,
					},
				})
				lbBackendPoolIDsSlice = append(lbBackendPoolIDsSlice, lbBackendPoolIDs[isIPv6])
			}
		} else {
			logger.V(10).Info("lb backendpool - found unmanaged backendpool", "serviceName", serviceName, "backendPoolName", ptr.Deref(bp.Name, ""))
		}
	}
	if len(backendpoolToBeDeleted) > 0 {
		// decouple the backendPool from the node
		updated, err := bc.VMSet.EnsureBackendPoolDeleted(ctx, service, lbBackendPoolIDsSlice, vmSetName, backendpoolToBeDeleted, false)
		if err != nil {
			return false, false, nil, err
		}
		if updated {
			backendPoolsUpdated = true
		}
	}

	if backendPoolsUpdated {
		logger.V(4).Info("refreshing load balancer", "serviceName", serviceName, "lbName", lbName)
		lb, _, err = bc.getAzureLoadBalancer(ctx, lbName, cache.CacheReadTypeForceRefresh)
		if err != nil {
			return false, false, nil, fmt.Errorf("bc.ReconcileBackendPools for service (%s): failed to get loadbalancer %s: %w", serviceName, lbName, err)
		}
	}

	for _, ipFamily := range service.Spec.IPFamilies {
		if foundBackendPools[ipFamily == v1.IPv6Protocol] {
			continue
		}
		isBackendPoolPreConfigured = newBackendPool(lb, isBackendPoolPreConfigured,
			bc.PreConfiguredBackendPoolLoadBalancerTypes, serviceName,
			lbBackendPoolNames[ipFamily == v1.IPv6Protocol])
		backendPoolsCreated = true
	}

	if isMigration {
		defer func() {
			mc.ObserveOperationWithResult(isOperationSucceeded)
		}()
	}

	isOperationSucceeded = true
	return isBackendPoolPreConfigured, backendPoolsCreated, lb, err
}

func getBackendIPConfigurationsToBeDeleted(
	bp armnetwork.BackendAddressPool,
	bipConfigNotFound, bipConfigExclude []*armnetwork.InterfaceIPConfiguration,
) []*armnetwork.InterfaceIPConfiguration {
	logger := log.Background().WithName("getBackendIPConfigurationsToBeDeleted")
	if bp.Properties == nil || bp.Properties.BackendIPConfigurations == nil {
		return []*armnetwork.InterfaceIPConfiguration{}
	}

	bipConfigNotFoundIDSet := utilsets.NewString()
	bipConfigExcludeIDSet := utilsets.NewString()
	for _, ipConfig := range bipConfigNotFound {
		bipConfigNotFoundIDSet.Insert(ptr.Deref(ipConfig.ID, ""))
	}
	for _, ipConfig := range bipConfigExclude {
		bipConfigExcludeIDSet.Insert(ptr.Deref(ipConfig.ID, ""))
	}

	var bipConfigToBeDeleted []*armnetwork.InterfaceIPConfiguration
	ipConfigs := bp.Properties.BackendIPConfigurations
	for i := len(ipConfigs) - 1; i >= 0; i-- {
		ipConfigID := ptr.Deref(ipConfigs[i].ID, "")
		if bipConfigNotFoundIDSet.Has(ipConfigID) {
			bipConfigToBeDeleted = append(bipConfigToBeDeleted, ipConfigs[i])
			ipConfigs = append(ipConfigs[:i], ipConfigs[i+1:]...)
		}
	}

	var unwantedIPConfigs []*armnetwork.InterfaceIPConfiguration
	for _, ipConfig := range ipConfigs {
		ipConfigID := ptr.Deref(ipConfig.ID, "")
		if bipConfigExcludeIDSet.Has(ipConfigID) {
			unwantedIPConfigs = append(unwantedIPConfigs, ipConfig)
		}
	}
	if len(unwantedIPConfigs) == len(ipConfigs) {
		logger.V(2).Info("the pool is empty or will be empty after removing the unwanted IP addresses, skipping the removal")
		return bipConfigToBeDeleted
	}
	return append(bipConfigToBeDeleted, unwantedIPConfigs...)
}

func (bc *backendPoolTypeNodeIPConfig) GetBackendPrivateIPs(ctx context.Context, clusterName string, service *v1.Service, lb *armnetwork.LoadBalancer) ([]string, []string) {
	logger := log.FromContextOrBackground(ctx).WithName("bc.GetBackendPrivateIPs")
	serviceName := getServiceName(service)
	lbBackendPoolNames := getBackendPoolNames(clusterName)
	if lb.Properties == nil || lb.Properties.BackendAddressPools == nil {
		return nil, nil
	}

	backendPrivateIPv4s, backendPrivateIPv6s := utilsets.NewString(), utilsets.NewString()
	for _, bp := range lb.Properties.BackendAddressPools {
		found, _ := isLBBackendPoolsExisting(lbBackendPoolNames, bp.Name)
		if found {
			logger.V(10).Info("found wanted backendpool", "serviceName", serviceName, "backendPoolName", ptr.Deref(bp.Name, ""))
			if bp.Properties != nil && bp.Properties.BackendIPConfigurations != nil {
				for _, backendIPConfig := range bp.Properties.BackendIPConfigurations {
					ipConfigID := ptr.Deref(backendIPConfig.ID, "")
					nodeName, _, err := bc.VMSet.GetNodeNameByIPConfigurationID(ctx, ipConfigID)
					if err != nil {
						logger.Error(err, "failed: GetNodeNameByIPConfigurationID", "service", serviceName)
						continue
					}
					privateIPsSet, ok := bc.nodePrivateIPs[strings.ToLower(nodeName)]
					if !ok {
						klog.Warningf("bc.GetBackendPrivateIPs for service (%s): failed to get private IPs of node %s", serviceName, nodeName)
						continue
					}
					privateIPs := privateIPsSet.UnsortedList()
					for _, ip := range privateIPs {
						logger.V(2).Info("lb backendpool - found private IPs of node", "serviceName", serviceName, "ip", ip, "nodeName", nodeName)
						if utilnet.IsIPv4String(ip) {
							backendPrivateIPv4s.Insert(ip)
						} else {
							backendPrivateIPv6s.Insert(ip)
						}
					}
				}
			}
		} else {
			logger.V(10).Info("found unmanaged backendpool", "serviceName", serviceName, "backendPoolName", ptr.Deref(bp.Name, ""))
		}
	}
	return backendPrivateIPv4s.UnsortedList(), backendPrivateIPv6s.UnsortedList()
}

type backendPoolTypeNodeIP struct {
	*Cloud
}

func newBackendPoolTypeNodeIP(c *Cloud) BackendPool {
	return &backendPoolTypeNodeIP{c}
}

func (az *Cloud) getVnetResourceID() string {
	rg := az.ResourceGroup
	if len(az.VnetResourceGroup) > 0 {
		rg = az.VnetResourceGroup
	}
	return fmt.Sprintf(
		"/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Network/virtualNetworks/%s",
		az.SubscriptionID,
		rg,
		az.VnetName,
	)
}

func (bi *backendPoolTypeNodeIP) EnsureHostsInPool(ctx context.Context, service *v1.Service, nodes []*v1.Node, _, _, clusterName, lbName string, backendPool *armnetwork.BackendAddressPool) error {
	logger := log.FromContextOrBackground(ctx).WithName("bi.EnsureHostsInPool")
	if backendPool == nil {
		backendPool = &armnetwork.BackendAddressPool{}
	}
	isIPv6 := isBackendPoolIPv6(ptr.Deref(backendPool.Name, ""))

	var (
		changed               bool
		numOfAdd, numOfDelete int
		activeNodes           *utilsets.IgnoreCaseSet
	)
	if bi.UseMultipleStandardLoadBalancers() {
		if !isLocalService(service) {
			activeNodes = bi.getActiveNodesByLoadBalancerName(lbName)
		} else {
			key := strings.ToLower(getServiceName(service))
			si, found := bi.getLocalServiceInfo(key)
			if found && !strings.EqualFold(si.lbName, lbName) {
				logger.V(4).Info("the service is not on the load balancer",
					"service", key,
					"previous load balancer", lbName,
					"current load balancer", si.lbName)
				return nil
			}
			activeNodes = bi.getLocalServiceEndpointsNodeNames(service)
		}

		if isNICPool(backendPool) {
			logger.V(4).Info("skipping NIC-based backend pool", "backendPoolName", ptr.Deref(backendPool.Name, ""))
			return nil
		}
	}

	lbBackendPoolName := bi.getBackendPoolNameForService(service, clusterName, isIPv6)
	if strings.EqualFold(ptr.Deref(backendPool.Name, ""), lbBackendPoolName) &&
		backendPool.Properties != nil {
		if backendPool.Properties.LoadBalancerBackendAddresses == nil {
			lbBackendPoolAddresses := make([]*armnetwork.LoadBalancerBackendAddress, 0)
			backendPool.Properties.LoadBalancerBackendAddresses = lbBackendPoolAddresses
		}

		existingIPs := utilsets.NewString()
		for _, loadBalancerBackendAddress := range backendPool.Properties.LoadBalancerBackendAddresses {
			if loadBalancerBackendAddress.Properties != nil &&
				loadBalancerBackendAddress.Properties.IPAddress != nil {
				logger.V(4).Info("found existing IP in the backend pool", "ip", ptr.Deref(loadBalancerBackendAddress.Properties.IPAddress, ""), "backendPoolName", lbBackendPoolName)
				existingIPs.Insert(ptr.Deref(loadBalancerBackendAddress.Properties.IPAddress, ""))
			}
		}

		var nodeIPsToBeAdded []string
		nodePrivateIPsSet := utilsets.NewString()
		for _, node := range nodes {
			if isControlPlaneNode(node) {
				logger.V(4).Info("skipping control plane node", "nodeName", node.Name)
				continue
			}

			privateIP := getNodePrivateIPAddress(node, isIPv6)
			if privateIP != "" {
				nodePrivateIPsSet.Insert(privateIP)
			}

			if bi.UseMultipleStandardLoadBalancers() {
				if activeNodes != nil && !activeNodes.Has(node.Name) {
					logger.V(4).Info("node should not be in load balancer", "nodeName", node.Name, "lbName", lbName)
					continue
				}
			}

			if !existingIPs.Has(privateIP) {
				name := node.Name
				logger.V(6).Info("adding node with ip address", "nodeName", name, "ip", privateIP)
				nodeIPsToBeAdded = append(nodeIPsToBeAdded, privateIP)
				numOfAdd++
			}
		}
		changed = bi.addNodeIPAddressesToBackendPool(backendPool, nodeIPsToBeAdded)

		var nodeIPsToBeDeleted []string
		for _, loadBalancerBackendAddress := range backendPool.Properties.LoadBalancerBackendAddresses {
			ip := ptr.Deref(loadBalancerBackendAddress.Properties.IPAddress, "")
			if !nodePrivateIPsSet.Has(ip) {
				logger.V(4).Info("removing IP because it is deleted or should be excluded", "ip", ip)
				nodeIPsToBeDeleted = append(nodeIPsToBeDeleted, ip)
				changed = true
				numOfDelete++
			} else if bi.UseMultipleStandardLoadBalancers() && activeNodes != nil {
				nodeName, ok := bi.nodePrivateIPToNodeNameMap[ip]
				if !ok {
					klog.Warningf("bi.EnsureHostsInPool: cannot find node name for private IP %s", ip)
					continue
				}
				if !activeNodes.Has(nodeName) {
					logger.V(4).Info("removing IP because it should not be in this load balancer", "ip", ip)
					nodeIPsToBeDeleted = append(nodeIPsToBeDeleted, ip)
					changed = true
					numOfDelete++
				}
			}
		}
		removeNodeIPAddressesFromBackendPool(backendPool, nodeIPsToBeDeleted, false, bi.UseMultipleStandardLoadBalancers(), true)
	}
	if changed {
		logger.V(2).Info("updating backend pool of load balancer to add and remove nodes", "backendPoolName", lbBackendPoolName, "lbName", lbName, "numOfAdd", numOfAdd, "numOfDelete", numOfDelete)
		if err := bi.CreateOrUpdateLBBackendPool(ctx, lbName, backendPool); err != nil {
			return fmt.Errorf("bi.EnsureHostsInPool: failed to update backend pool %s: %w", lbBackendPoolName, err)
		}
	}

	return nil
}

func (bi *backendPoolTypeNodeIP) ReconcileBackendPools(ctx context.Context, clusterName string, service *v1.Service, lb *armnetwork.LoadBalancer) (bool, bool, *armnetwork.LoadBalancer, error) {
	logger := log.FromContextOrBackground(ctx).WithName("bi.ReconcileBackendPools")
	var newBackendPools []*armnetwork.BackendAddressPool
	if lb.Properties.BackendAddressPools != nil {
		newBackendPools = lb.Properties.BackendAddressPools
	}

	var backendPoolsUpdated, shouldRefreshLB, isOperationSucceeded, isMigration, updated bool
	foundBackendPools := map[bool]bool{}
	lbName := *lb.Name
	serviceName := getServiceName(service)
	lbBackendPoolNames := bi.getBackendPoolNamesForService(service, clusterName)
	vmSetName := bi.mapLoadBalancerNameToVMSet(lbName, clusterName)
	lbBackendPoolIDs := bi.getBackendPoolIDsForService(service, clusterName, ptr.Deref(lb.Name, ""))
	isBackendPoolPreConfigured := bi.isBackendPoolPreConfigured(service)

	mc := metrics.NewMetricContext("services", "migrate_to_ip_based_backend_pool", bi.ResourceGroup, bi.getNetworkResourceSubscriptionID(), serviceName)

	var (
		err                   error
		bpIdxes               []int
		lbBackendPoolIDsSlice []string
	)
	nicsCountMap := make(map[string]int)
	for i := len(newBackendPools) - 1; i >= 0; i-- {
		bp := newBackendPools[i]
		found, isIPv6 := isLBBackendPoolsExisting(lbBackendPoolNames, bp.Name)
		if found {
			bpIdxes = append(bpIdxes, i)
			logger.V(10).Info("found wanted backendpool. Not adding anything", "serviceName", serviceName)
			foundBackendPools[isIPv6] = true
			lbBackendPoolIDsSlice = append(lbBackendPoolIDsSlice, lbBackendPoolIDs[isIPv6])

			if nicsCount := countNICsOnBackendPool(bp); nicsCount > 0 {
				nicsCountMap[ptr.Deref(bp.Name, "")] = nicsCount
				logger.V(4).Info(
					"found NIC-based backendpool with NICs, will migrate to IP-based",
					"serviceName", serviceName,
					"backendPoolName", ptr.Deref(bp.Name, ""),
					"nicsCount", nicsCount,
				)
				isMigration = true
			}
		} else {
			logger.V(10).Info("found unmanaged backendpool", "serviceName", serviceName, "backendPoolName", *bp.Name)
		}
	}

	// Don't bother to remove unused nodeIP if backend pool is pre configured
	if !isBackendPoolPreConfigured {
		// If the LB backend pool type is configured from nodeIPConfiguration
		// to nodeIP, we need to decouple the VM NICs from the LB
		// before attaching nodeIPs/podIPs to the LB backend pool.
		// If the migration API is enabled, we use the migration API to decouple
		// the VM NICs from the LB. Then we manually decouple the VMSS
		// and its VMs from the LB by EnsureBackendPoolDeleted. These manual operations
		// cannot be omitted because we use the VMSS manual upgrade policy.
		// If the migration API is not enabled, we manually decouple the VM NICs and
		// the VMSS from the LB by EnsureBackendPoolDeleted. If no NIC-based backend
		// pool is found (it is not a migration scenario), EnsureBackendPoolDeleted would be a no-op.
		if isMigration && bi.EnableMigrateToIPBasedBackendPoolAPI {
			var backendPoolNames []string
			for _, id := range lbBackendPoolIDsSlice {
				name, err := getBackendPoolNameFromBackendPoolID(id)
				if err != nil {
					logger.Error(err, "failed to get LB name from backend pool ID", "service", serviceName)
					return false, false, nil, err
				}
				backendPoolNames = append(backendPoolNames, name)
			}

			if err := bi.MigrateToIPBasedBackendPoolAndWaitForCompletion(ctx, lbName, backendPoolNames, nicsCountMap); err != nil {
				backendPoolNamesStr := strings.Join(backendPoolNames, ",")
				logger.Error(err, "Failed to migrate to IP based backend pool", "lbName", lbName, "backendPoolNames", backendPoolNamesStr)
				return false, false, nil, err
			}
		}

		// EnsureBackendPoolDeleted is useful in the following scenarios:
		// 1. Migrate from NIC-based to IP-based backend pool if the migration
		// API is not enabled.
		// 2. Migrate from NIC-based to IP-based backend pool when the migration
		// API is enabled. This is needed because since we use the manual upgrade
		// policy on VMSS so the migration API will not change the VMSS and VMSS
		// VMs during the migration.
		// 3. Decouple vmss from the lb if the backend pool is empty when using
		// ip-based LB. Ref: https://github.com/kubernetes-sigs/cloud-provider-azure/pull/2829.
		logger.V(2).Info("ensuring the LB is decoupled from the VMSet", "serviceName", serviceName, "vmSetName", vmSetName)
		shouldRefreshLB, err = bi.VMSet.EnsureBackendPoolDeleted(ctx, service, lbBackendPoolIDsSlice, vmSetName, lb.Properties.BackendAddressPools, true)
		if err != nil {
			logger.Error(err, "failed to EnsureBackendPoolDeleted", "service", serviceName)
			return false, false, nil, err
		}

		for _, i := range bpIdxes {
			bp := newBackendPools[i]
			var nodeIPAddressesToBeDeleted []string
			for _, nodeName := range bi.excludeLoadBalancerNodes.UnsortedList() {
				for _, ip := range bi.nodePrivateIPs[strings.ToLower(nodeName)].UnsortedList() {
					logger.V(2).Info("found unwanted node private IP, decouple it from the LB", "serviceName", serviceName, "ip", ip, "lbName", lbName)
					nodeIPAddressesToBeDeleted = append(nodeIPAddressesToBeDeleted, ip)
				}
			}
			if len(nodeIPAddressesToBeDeleted) > 0 {
				if removeNodeIPAddressesFromBackendPool(bp, nodeIPAddressesToBeDeleted, false, false, true) {
					updated = true
				}
			}
			// delete the vnet in LoadBalancerBackendAddresses and ensure it is in the backend pool level
			var vnet string
			if bp.Properties != nil {
				if bp.Properties.VirtualNetwork == nil ||
					ptr.Deref(bp.Properties.VirtualNetwork.ID, "") == "" {
					if bp.Properties.LoadBalancerBackendAddresses != nil {
						for _, a := range bp.Properties.LoadBalancerBackendAddresses {
							if a.Properties != nil &&
								a.Properties.VirtualNetwork != nil {
								if vnet == "" {
									vnet = ptr.Deref(a.Properties.VirtualNetwork.ID, "")
								}
								a.Properties.VirtualNetwork = nil
							}
						}
					}
					if vnet != "" {
						bp.Properties.VirtualNetwork = &armnetwork.SubResource{
							ID: ptr.To(vnet),
						}
						updated = true
					}
				}
			}

			if updated {
				(lb.Properties.BackendAddressPools)[i] = bp
				if err := bi.CreateOrUpdateLBBackendPool(ctx, lbName, bp); err != nil {
					return false, false, nil, fmt.Errorf("bi.ReconcileBackendPools for service (%s): lb backendpool - failed to update backend pool %s for load balancer %s: %w", serviceName, ptr.Deref(bp.Name, ""), lbName, err)
				}
				shouldRefreshLB = true
			}
		}
	}

	shouldRefreshLB = shouldRefreshLB || isMigration
	if shouldRefreshLB {
		logger.V(4).Info("refreshing load balancer", "serviceName", serviceName, "lbName", lbName)
		lb, _, err = bi.getAzureLoadBalancer(ctx, lbName, cache.CacheReadTypeForceRefresh)
		if err != nil {
			return false, false, nil, fmt.Errorf("bi.ReconcileBackendPools for service (%s): failed to get loadbalancer %s: %w", serviceName, lbName, err)
		}
	}

	for _, ipFamily := range service.Spec.IPFamilies {
		if foundBackendPools[ipFamily == v1.IPv6Protocol] {
			continue
		}
		isBackendPoolPreConfigured = newBackendPool(lb, isBackendPoolPreConfigured,
			bi.PreConfiguredBackendPoolLoadBalancerTypes, serviceName,
			lbBackendPoolNames[ipFamily == v1.IPv6Protocol])
		backendPoolsUpdated = true
	}

	if isMigration {
		defer func() {
			mc.ObserveOperationWithResult(isOperationSucceeded)
		}()
	}

	isOperationSucceeded = true
	return isBackendPoolPreConfigured, backendPoolsUpdated, lb, nil
}

func (bi *backendPoolTypeNodeIP) GetBackendPrivateIPs(ctx context.Context, clusterName string, service *v1.Service, lb *armnetwork.LoadBalancer) ([]string, []string) {
	logger := log.FromContextOrBackground(ctx).WithName("GetBackendPrivateIPs")
	serviceName := getServiceName(service)
	lbBackendPoolNames := bi.getBackendPoolNamesForService(service, clusterName)
	if lb.Properties == nil || lb.Properties.BackendAddressPools == nil {
		return nil, nil
	}

	backendPrivateIPv4s, backendPrivateIPv6s := utilsets.NewString(), utilsets.NewString()
	for _, bp := range lb.Properties.BackendAddressPools {
		found, _ := isLBBackendPoolsExisting(lbBackendPoolNames, bp.Name)
		if found {
			logger.V(10).Info("found wanted backendpool", "serviceName", serviceName, "backendPoolName", ptr.Deref(bp.Name, ""))
			if bp.Properties != nil && bp.Properties.LoadBalancerBackendAddresses != nil {
				for _, backendAddress := range bp.Properties.LoadBalancerBackendAddresses {
					ipAddress := backendAddress.Properties.IPAddress
					if ipAddress != nil {
						logger.V(2).Info("lb backendpool - found private IP", "serviceName", serviceName, "ip", *ipAddress)
						if utilnet.IsIPv4String(*ipAddress) {
							backendPrivateIPv4s.Insert(*ipAddress)
						} else if utilnet.IsIPv6String(*ipAddress) {
							backendPrivateIPv6s.Insert(*ipAddress)
						}
					} else {
						logger.V(4).Info("lb backendpool - found null private IP", "serviceName", serviceName)
					}
				}
			}
		} else {
			logger.V(10).Info("found unmanaged backendpool", "serviceName", serviceName, "backendPoolName", ptr.Deref(bp.Name, ""))
		}
	}
	return backendPrivateIPv4s.UnsortedList(), backendPrivateIPv6s.UnsortedList()
}

// getBackendPoolNameForService returns all node names in the backend pool.
func (bi *backendPoolTypeNodeIP) getBackendPoolNodeNames(bp *armnetwork.BackendAddressPool) []string {
	nodeNames := utilsets.NewString()
	if bp.Properties != nil && bp.Properties.LoadBalancerBackendAddresses != nil {
		for _, backendAddress := range bp.Properties.LoadBalancerBackendAddresses {
			if backendAddress.Properties != nil {
				ip := ptr.Deref(backendAddress.Properties.IPAddress, "")
				nodeNames.Insert(bi.nodePrivateIPToNodeNameMap[ip])
			}
		}
	}
	return nodeNames.UnsortedList()
}

func newBackendPool(lb *armnetwork.LoadBalancer, isBackendPoolPreConfigured bool, preConfiguredBackendPoolLoadBalancerTypes, serviceName, lbBackendPoolName string) bool {
	logger := log.Background().WithName("newBackendPool")
	if isBackendPoolPreConfigured {
		logger.V(2).Info("lb backendpool - PreConfiguredBackendPoolLoadBalancerTypes has been set but can not find corresponding backend pool, ignoring it",
			"serviceName", serviceName,
			"preConfiguredBackendPoolLoadBalancerTypes", preConfiguredBackendPoolLoadBalancerTypes,
			"backendPoolName", lbBackendPoolName)
		isBackendPoolPreConfigured = false
	}

	if lb.Properties.BackendAddressPools == nil {
		lb.Properties.BackendAddressPools = []*armnetwork.BackendAddressPool{}
	}
	lb.Properties.BackendAddressPools = append(lb.Properties.BackendAddressPools, &armnetwork.BackendAddressPool{
		Name:       ptr.To(lbBackendPoolName),
		Properties: &armnetwork.BackendAddressPoolPropertiesFormat{},
	})

	// Always returns false
	return isBackendPoolPreConfigured
}

func (az *Cloud) addNodeIPAddressesToBackendPool(backendPool *armnetwork.BackendAddressPool, nodeIPAddresses []string) bool {
	logger := log.Background().WithName("bi.addNodeIPAddressesToBackendPool")
	vnetID := az.getVnetResourceID()
	if backendPool.Properties != nil {
		if backendPool.Properties.VirtualNetwork == nil ||
			backendPool.Properties.VirtualNetwork.ID == nil {
			backendPool.Properties.VirtualNetwork = &armnetwork.SubResource{
				ID: &vnetID,
			}
		}
	} else {
		backendPool.Properties = &armnetwork.BackendAddressPoolPropertiesFormat{
			VirtualNetwork: &armnetwork.SubResource{
				ID: &vnetID,
			},
		}
	}

	if backendPool.Properties.LoadBalancerBackendAddresses == nil {
		lbBackendPoolAddresses := make([]*armnetwork.LoadBalancerBackendAddress, 0)
		backendPool.Properties.LoadBalancerBackendAddresses = lbBackendPoolAddresses
	}

	var changed bool
	addresses := backendPool.Properties.LoadBalancerBackendAddresses
	for _, ipAddress := range nodeIPAddresses {
		if !hasIPAddressInBackendPool(backendPool, ipAddress) {
			name := az.nodePrivateIPToNodeNameMap[ipAddress]
			logger.V(4).Info("adding node to the backend pool", "ip", ipAddress, "backendPoolName", ptr.Deref(backendPool.Name, ""))
			addresses = append(addresses, &armnetwork.LoadBalancerBackendAddress{
				Name: ptr.To(name),
				Properties: &armnetwork.LoadBalancerBackendAddressPropertiesFormat{
					IPAddress: ptr.To(ipAddress),
				},
			})
			backendPool.Properties.LoadBalancerBackendAddresses = addresses
			changed = true
		}
	}
	return changed
}

func hasIPAddressInBackendPool(backendPool *armnetwork.BackendAddressPool, ipAddress string) bool {
	if backendPool.Properties.LoadBalancerBackendAddresses == nil {
		return false
	}

	addresses := backendPool.Properties.LoadBalancerBackendAddresses
	for _, address := range addresses {
		if address.Properties != nil &&
			ptr.Deref(address.Properties.IPAddress, "") == ipAddress {
			return true
		}
	}

	return false
}

func removeNodeIPAddressesFromBackendPool(
	backendPool *armnetwork.BackendAddressPool,
	nodeIPAddresses []string,
	removeAll, UseMultipleStandardLoadBalancers, isNodeIP bool,
) bool {
	changed := false
	nodeIPsSet := utilsets.NewString(nodeIPAddresses...)

	logger := log.Background().WithName("removeNodeIPAddressFromBackendPool")

	if backendPool.Properties == nil ||
		backendPool.Properties.LoadBalancerBackendAddresses == nil {
		return false
	}

	addresses := backendPool.Properties.LoadBalancerBackendAddresses
	for i := len(addresses) - 1; i >= 0; i-- {
		if addresses[i].Properties != nil {
			ipAddress := ptr.Deref((backendPool.Properties.LoadBalancerBackendAddresses)[i].Properties.IPAddress, "")
			if ipAddress == "" {
				if isNodeIP {
					logger.V(4).Info("LoadBalancerBackendAddress is not IP-based, removing", "LoadBalancerBackendAddress", ptr.Deref(addresses[i].Name, ""))
					addresses = append(addresses[:i], addresses[i+1:]...)
					changed = true
				} else {
					logger.V(4).Info("LoadBalancerBackendAddress is not IP-based, skipping", "LoadBalancerBackendAddress", ptr.Deref(addresses[i].Name, ""))
				}
				continue
			}
			if removeAll || nodeIPsSet.Has(ipAddress) {
				logger.V(4).Info("removing IP from the backend pool", "ip", ipAddress, "backendPoolName", ptr.Deref(backendPool.Name, ""))
				addresses = append(addresses[:i], addresses[i+1:]...)
				changed = true
			}
		}
	}

	if removeAll {
		backendPool.Properties.LoadBalancerBackendAddresses = addresses
		return changed
	}

	// Allow the pool to be empty when EnsureHostsInPool for multiple standard load balancers clusters,
	// or one node could occur in multiple backend pools.
	if len(addresses) == 0 && !UseMultipleStandardLoadBalancers {
		logger.V(2).Info("the pool is empty or will be empty after removing the unwanted IP addresses, skipping the removal")
		changed = false
	} else if changed {
		backendPool.Properties.LoadBalancerBackendAddresses = addresses
	}

	return changed
}
