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
	"errors"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2022-07-01/network"

	v1 "k8s.io/api/core/v1"
	cloudprovider "k8s.io/cloud-provider"
	"k8s.io/klog/v2"
	utilnet "k8s.io/utils/net"
	"k8s.io/utils/pointer"

	"sigs.k8s.io/cloud-provider-azure/pkg/cache"
	"sigs.k8s.io/cloud-provider-azure/pkg/consts"
	"sigs.k8s.io/cloud-provider-azure/pkg/metrics"
	utilsets "sigs.k8s.io/cloud-provider-azure/pkg/util/sets"
)

type BackendPool interface {
	// EnsureHostsInPool ensures the nodes join the backend pool of the load balancer
	EnsureHostsInPool(service *v1.Service, nodes []*v1.Node, backendPoolID, vmSetName, clusterName, lbName string, backendPool network.BackendAddressPool) error

	// CleanupVMSetFromBackendPoolByCondition removes nodes of the unwanted vmSet from the lb backend pool.
	// This is needed in two scenarios:
	// 1. When migrating from single SLB to multiple SLBs, the existing
	// SLB's backend pool contains nodes from different agent pools, while we only want the
	// nodes from the primary agent pool to join the backend pool.
	// 2. When migrating from dedicated SLB to shared SLB (or vice versa), we should move the vmSet from
	// one SLB to another one.
	CleanupVMSetFromBackendPoolByCondition(slb *network.LoadBalancer, service *v1.Service, nodes []*v1.Node, clusterName string, shouldRemoveVMSetFromSLB func(string) bool) (*network.LoadBalancer, error)

	// ReconcileBackendPools creates the inbound backend pool if it is not existed, and removes nodes that are supposed to be
	// excluded from the load balancers.
	ReconcileBackendPools(clusterName string, service *v1.Service, lb *network.LoadBalancer) (bool, bool, *network.LoadBalancer, error)

	// GetBackendPrivateIPs returns the private IPs of LoadBalancer's backend pool
	GetBackendPrivateIPs(clusterName string, service *v1.Service, lb *network.LoadBalancer) ([]string, []string)
}

type backendPoolTypeNodeIPConfig struct {
	*Cloud
}

func newBackendPoolTypeNodeIPConfig(c *Cloud) BackendPool {
	return &backendPoolTypeNodeIPConfig{c}
}

func (bc *backendPoolTypeNodeIPConfig) EnsureHostsInPool(service *v1.Service, nodes []*v1.Node, backendPoolID, vmSetName, _, _ string, _ network.BackendAddressPool) error {
	return bc.VMSet.EnsureHostsInPool(service, nodes, backendPoolID, vmSetName)
}

func isLBBackendPoolsExisting(lbBackendPoolNames map[bool]string, bpName *string) (found, isIPv6 bool) {
	if strings.EqualFold(pointer.StringDeref(bpName, ""), lbBackendPoolNames[consts.IPVersionIPv4]) {
		isIPv6 = false
		found = true
	}
	if strings.EqualFold(pointer.StringDeref(bpName, ""), lbBackendPoolNames[consts.IPVersionIPv6]) {
		isIPv6 = true
		found = true
	}
	return found, isIPv6
}

func (bc *backendPoolTypeNodeIPConfig) CleanupVMSetFromBackendPoolByCondition(slb *network.LoadBalancer, service *v1.Service, _ []*v1.Node, clusterName string, shouldRemoveVMSetFromSLB func(string) bool) (*network.LoadBalancer, error) {
	v4Enabled, v6Enabled := getIPFamiliesEnabled(service)

	lbBackendPoolNames := getBackendPoolNames(clusterName)
	lbBackendPoolIDs := bc.getBackendPoolIDs(clusterName, pointer.StringDeref(slb.Name, ""))
	newBackendPools := make([]network.BackendAddressPool, 0)
	if slb.LoadBalancerPropertiesFormat != nil && slb.BackendAddressPools != nil {
		newBackendPools = *slb.BackendAddressPools
	}
	vmSetNameToBackendIPConfigurationsToBeDeleted := make(map[string][]network.InterfaceIPConfiguration)

	for j, bp := range newBackendPools {
		if found, _ := isLBBackendPoolsExisting(lbBackendPoolNames, bp.Name); found {
			klog.V(2).Infof("bc.CleanupVMSetFromBackendPoolByCondition: checking the backend pool %s from standard load balancer %s", pointer.StringDeref(bp.Name, ""), pointer.StringDeref(slb.Name, ""))
			if bp.BackendAddressPoolPropertiesFormat != nil && bp.BackendIPConfigurations != nil {
				for i := len(*bp.BackendIPConfigurations) - 1; i >= 0; i-- {
					ipConf := (*bp.BackendIPConfigurations)[i]
					ipConfigID := pointer.StringDeref(ipConf.ID, "")
					_, vmSetName, err := bc.VMSet.GetNodeNameByIPConfigurationID(ipConfigID)
					if err != nil && !errors.Is(err, cloudprovider.InstanceNotFound) {
						return nil, err
					}

					if shouldRemoveVMSetFromSLB(vmSetName) {
						klog.V(2).Infof("bc.CleanupVMSetFromBackendPoolByCondition: found unwanted vmSet %s, decouple it from the LB", vmSetName)
						// construct a backendPool that only contains the IP config of the node to be deleted
						interfaceIPConfigToBeDeleted := network.InterfaceIPConfiguration{
							ID: pointer.String(ipConfigID),
						}
						vmSetNameToBackendIPConfigurationsToBeDeleted[vmSetName] = append(vmSetNameToBackendIPConfigurationsToBeDeleted[vmSetName], interfaceIPConfigToBeDeleted)
						*bp.BackendIPConfigurations = append((*bp.BackendIPConfigurations)[:i], (*bp.BackendIPConfigurations)[i+1:]...)
					}
				}
			}

			newBackendPools[j] = bp
		}
	}

	for vmSetName := range vmSetNameToBackendIPConfigurationsToBeDeleted {
		shouldRefreshLB := false
		backendIPConfigurationsToBeDeleted := vmSetNameToBackendIPConfigurationsToBeDeleted[vmSetName]
		backendpoolToBeDeleted := []network.BackendAddressPool{}
		lbBackendPoolIDsSlice := []string{}
		findBackendpoolToBeDeleted := func(isIPv6 bool) {
			lbBackendPoolIDsSlice = append(lbBackendPoolIDsSlice, lbBackendPoolIDs[isIPv6])
			backendpoolToBeDeleted = append(backendpoolToBeDeleted, network.BackendAddressPool{
				ID: pointer.String(lbBackendPoolIDs[isIPv6]),
				BackendAddressPoolPropertiesFormat: &network.BackendAddressPoolPropertiesFormat{
					BackendIPConfigurations: &backendIPConfigurationsToBeDeleted,
				},
			})
		}
		if v4Enabled {
			findBackendpoolToBeDeleted(consts.IPVersionIPv4)
		}
		if v6Enabled {
			findBackendpoolToBeDeleted(consts.IPVersionIPv6)
		}
		// decouple the backendPool from the node
		shouldRefreshLB, err := bc.VMSet.EnsureBackendPoolDeleted(service, lbBackendPoolIDsSlice, vmSetName, &backendpoolToBeDeleted, true)
		if err != nil {
			return nil, err
		}
		if shouldRefreshLB {
			slb, _, err := bc.getAzureLoadBalancer(pointer.StringDeref(slb.Name, ""), cache.CacheReadTypeForceRefresh)
			if err != nil {
				return nil, fmt.Errorf("bc.CleanupVMSetFromBackendPoolByCondition: failed to get load balancer %s, err: %w", pointer.StringDeref(slb.Name, ""), err)
			}
		}
	}

	return slb, nil
}

func (bc *backendPoolTypeNodeIPConfig) ReconcileBackendPools(
	clusterName string,
	service *v1.Service,
	lb *network.LoadBalancer,
) (bool, bool, *network.LoadBalancer, error) {
	var newBackendPools []network.BackendAddressPool
	var err error
	if lb.BackendAddressPools != nil {
		newBackendPools = *lb.BackendAddressPools
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

	backendpoolToBeDeleted := []network.BackendAddressPool{}
	lbBackendPoolIDsSlice := []string{}
	for i := len(newBackendPools) - 1; i >= 0; i-- {
		bp := newBackendPools[i]
		found, isIPv6 := isLBBackendPoolsExisting(lbBackendPoolNames, bp.Name)
		if found {
			klog.V(10).Infof("bc.ReconcileBackendPools for service (%s): lb backendpool - found wanted backendpool. not adding anything", serviceName)
			foundBackendPools[isBackendPoolIPv6(pointer.StringDeref(bp.Name, ""))] = true

			// Don't bother to remove unused nodeIPConfiguration if backend pool is pre configured
			if isBackendPoolPreConfigured {
				break
			}

			// If the LB backend pool type is configured from nodeIP or podIP
			// to nodeIPConfiguration, we need to decouple the VM NICs from the LB
			// before attaching nodeIPs/podIPs to the LB backend pool.
			if bp.BackendAddressPoolPropertiesFormat != nil &&
				bp.LoadBalancerBackendAddresses != nil &&
				len(*bp.LoadBalancerBackendAddresses) > 0 {
				if removeNodeIPAddressesFromBackendPool(bp, []string{}, true, false) {
					isMigration = true
					bp.VirtualNetwork = nil
					if err := bc.CreateOrUpdateLBBackendPool(lbName, bp); err != nil {
						klog.Errorf("bc.ReconcileBackendPools for service (%s): failed to cleanup IP based backend pool %s: %s", serviceName, lbBackendPoolNames[isIPv6], err.Error())
						return false, false, nil, fmt.Errorf("bc.ReconcileBackendPools for service (%s): failed to cleanup IP based backend pool %s: %w", serviceName, lbBackendPoolNames[isIPv6], err)
					}
					newBackendPools[i] = bp
					lb.BackendAddressPools = &newBackendPools
					backendPoolsUpdated = true
				}
			}

			var backendIPConfigurationsToBeDeleted, bipConfigNotFound, bipConfigExclude []network.InterfaceIPConfiguration
			if bp.BackendAddressPoolPropertiesFormat != nil && bp.BackendIPConfigurations != nil {
				for _, ipConf := range *bp.BackendIPConfigurations {
					ipConfID := pointer.StringDeref(ipConf.ID, "")
					nodeName, _, err := bc.VMSet.GetNodeNameByIPConfigurationID(ipConfID)
					if err != nil {
						if errors.Is(err, cloudprovider.InstanceNotFound) {
							klog.V(2).Infof("bc.ReconcileBackendPools for service (%s): vm not found for ipConfID %s", serviceName, ipConfID)
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
						klog.Errorf("bc.ReconcileBackendPools: ShouldNodeExcludedFromLoadBalancer(%s) failed with error: %v", nodeName, err)
						return false, false, nil, err
					}
					if shouldExcludeLoadBalancer {
						klog.V(2).Infof("bc.ReconcileBackendPools for service (%s): lb backendpool - found unwanted node %s, decouple it from the LB %s", serviceName, nodeName, lbName)
						// construct a backendPool that only contains the IP config of the node to be deleted
						bipConfigExclude = append(bipConfigExclude, network.InterfaceIPConfiguration{ID: pointer.String(ipConfID)})
					}
				}
			}
			backendIPConfigurationsToBeDeleted = getBackendIPConfigurationsToBeDeleted(bp, bipConfigNotFound, bipConfigExclude)
			if len(backendIPConfigurationsToBeDeleted) > 0 {
				backendpoolToBeDeleted = append(backendpoolToBeDeleted, network.BackendAddressPool{
					ID: pointer.String(lbBackendPoolIDs[isIPv6]),
					BackendAddressPoolPropertiesFormat: &network.BackendAddressPoolPropertiesFormat{
						BackendIPConfigurations: &backendIPConfigurationsToBeDeleted,
					},
				})
				lbBackendPoolIDsSlice = append(lbBackendPoolIDsSlice, lbBackendPoolIDs[isIPv6])
			}
		} else {
			klog.V(10).Infof("bc.ReconcileBackendPools for service (%s): lb backendpool - found unmanaged backendpool %s", serviceName, pointer.StringDeref(bp.Name, ""))
		}
	}
	if len(backendpoolToBeDeleted) > 0 {
		// decouple the backendPool from the node
		updated, err := bc.VMSet.EnsureBackendPoolDeleted(service, lbBackendPoolIDsSlice, vmSetName, &backendpoolToBeDeleted, false)
		if err != nil {
			return false, false, nil, err
		}
		if updated {
			backendPoolsUpdated = true
		}
	}

	if backendPoolsUpdated {
		klog.V(4).Infof("bc.ReconcileBackendPools for service(%s): refreshing load balancer %s", serviceName, lbName)
		lb, _, err = bc.getAzureLoadBalancer(lbName, cache.CacheReadTypeForceRefresh)
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
	bp network.BackendAddressPool,
	bipConfigNotFound, bipConfigExclude []network.InterfaceIPConfiguration,
) []network.InterfaceIPConfiguration {
	if bp.BackendAddressPoolPropertiesFormat == nil || bp.BackendIPConfigurations == nil {
		return []network.InterfaceIPConfiguration{}
	}

	bipConfigNotFoundIDSet := utilsets.NewString()
	bipConfigExcludeIDSet := utilsets.NewString()
	for _, ipConfig := range bipConfigNotFound {
		bipConfigNotFoundIDSet.Insert(pointer.StringDeref(ipConfig.ID, ""))
	}
	for _, ipConfig := range bipConfigExclude {
		bipConfigExcludeIDSet.Insert(pointer.StringDeref(ipConfig.ID, ""))
	}

	var bipConfigToBeDeleted []network.InterfaceIPConfiguration
	ipConfigs := *bp.BackendIPConfigurations
	for i := len(ipConfigs) - 1; i >= 0; i-- {
		ipConfigID := pointer.StringDeref(ipConfigs[i].ID, "")
		if bipConfigNotFoundIDSet.Has(ipConfigID) {
			bipConfigToBeDeleted = append(bipConfigToBeDeleted, ipConfigs[i])
			ipConfigs = append(ipConfigs[:i], ipConfigs[i+1:]...)
		}
	}

	var unwantedIPConfigs []network.InterfaceIPConfiguration
	for _, ipConfig := range ipConfigs {
		ipConfigID := pointer.StringDeref(ipConfig.ID, "")
		if bipConfigExcludeIDSet.Has(ipConfigID) {
			unwantedIPConfigs = append(unwantedIPConfigs, ipConfig)
		}
	}
	if len(unwantedIPConfigs) == len(ipConfigs) {
		klog.V(2).Info("getBackendIPConfigurationsToBeDeleted: the pool is empty or will be empty after removing the unwanted IP addresses, skipping the removal")
		return bipConfigToBeDeleted
	}
	return append(bipConfigToBeDeleted, unwantedIPConfigs...)
}

func (bc *backendPoolTypeNodeIPConfig) GetBackendPrivateIPs(clusterName string, service *v1.Service, lb *network.LoadBalancer) ([]string, []string) {
	serviceName := getServiceName(service)
	lbBackendPoolNames := getBackendPoolNames(clusterName)
	if lb.LoadBalancerPropertiesFormat == nil || lb.LoadBalancerPropertiesFormat.BackendAddressPools == nil {
		return nil, nil
	}

	backendPrivateIPv4s, backendPrivateIPv6s := utilsets.NewString(), utilsets.NewString()
	for _, bp := range *lb.BackendAddressPools {
		found, _ := isLBBackendPoolsExisting(lbBackendPoolNames, bp.Name)
		if found {
			klog.V(10).Infof("bc.GetBackendPrivateIPs for service (%s): found wanted backendpool %s", serviceName, pointer.StringDeref(bp.Name, ""))
			if bp.BackendAddressPoolPropertiesFormat != nil && bp.BackendIPConfigurations != nil {
				for _, backendIPConfig := range *bp.BackendIPConfigurations {
					ipConfigID := pointer.StringDeref(backendIPConfig.ID, "")
					nodeName, _, err := bc.VMSet.GetNodeNameByIPConfigurationID(ipConfigID)
					if err != nil {
						klog.Errorf("bc.GetBackendPrivateIPs for service (%s): GetNodeNameByIPConfigurationID failed with error: %v", serviceName, err)
						continue
					}
					privateIPsSet, ok := bc.nodePrivateIPs[strings.ToLower(nodeName)]
					if !ok {
						klog.Warningf("bc.GetBackendPrivateIPs for service (%s): failed to get private IPs of node %s", serviceName, nodeName)
						continue
					}
					privateIPs := privateIPsSet.UnsortedList()
					for _, ip := range privateIPs {
						klog.V(2).Infof("bc.GetBackendPrivateIPs for service (%s): lb backendpool - found private IPs %s of node %s", serviceName, ip, nodeName)
						if utilnet.IsIPv4String(ip) {
							backendPrivateIPv4s.Insert(ip)
						} else {
							backendPrivateIPv6s.Insert(ip)
						}
					}
				}
			}
		} else {
			klog.V(10).Infof("bc.GetBackendPrivateIPs for service (%s): found unmanaged backendpool %s", serviceName, pointer.StringDeref(bp.Name, ""))
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

func (bi *backendPoolTypeNodeIP) EnsureHostsInPool(service *v1.Service, nodes []*v1.Node, _, _, clusterName, lbName string, backendPool network.BackendAddressPool) error {
	isIPv6 := isBackendPoolIPv6(pointer.StringDeref(backendPool.Name, ""))
	vnetResourceGroup := bi.ResourceGroup
	if len(bi.VnetResourceGroup) > 0 {
		vnetResourceGroup = bi.VnetResourceGroup
	}
	vnetID := fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Network/virtualNetworks/%s", bi.SubscriptionID, vnetResourceGroup, bi.VnetName)

	var (
		changed               bool
		numOfAdd, numOfDelete int
		activeNodes           *utilsets.IgnoreCaseSet
		err                   error
	)
	if bi.useMultipleStandardLoadBalancers() {
		if !isLocalService(service) {
			activeNodes = bi.getActiveNodesByLoadBalancerName(lbName)
		} else {
			key := strings.ToLower(getServiceName(service))
			si, found := bi.getLocalServiceInfo(key)
			if found && !strings.EqualFold(si.lbName, lbName) {
				klog.V(4).InfoS("EnsureHostsInPool: the service is not on the load balancer",
					"service", key,
					"previous load balancer", lbName,
					"current load balancer", si.lbName)
				return nil
			}
			activeNodes, err = bi.getLocalServiceEndpointsNodeNames(service)
			if err != nil {
				return err
			}
		}
	}

	lbBackendPoolName := bi.getBackendPoolNameForService(service, clusterName, isIPv6)
	if strings.EqualFold(pointer.StringDeref(backendPool.Name, ""), lbBackendPoolName) &&
		backendPool.BackendAddressPoolPropertiesFormat != nil {
		backendPool.VirtualNetwork = &network.SubResource{
			ID: &vnetID,
		}

		if backendPool.LoadBalancerBackendAddresses == nil {
			lbBackendPoolAddresses := make([]network.LoadBalancerBackendAddress, 0)
			backendPool.LoadBalancerBackendAddresses = &lbBackendPoolAddresses
		}

		existingIPs := utilsets.NewString()
		for _, loadBalancerBackendAddress := range *backendPool.LoadBalancerBackendAddresses {
			if loadBalancerBackendAddress.LoadBalancerBackendAddressPropertiesFormat != nil &&
				loadBalancerBackendAddress.IPAddress != nil {
				klog.V(4).Infof("bi.EnsureHostsInPool: found existing IP %s in the backend pool %s", pointer.StringDeref(loadBalancerBackendAddress.IPAddress, ""), lbBackendPoolName)
				existingIPs.Insert(pointer.StringDeref(loadBalancerBackendAddress.IPAddress, ""))
			}
		}

		var nodeIPsToBeAdded []string
		nodePrivateIPsSet := utilsets.NewString()
		for _, node := range nodes {
			if isControlPlaneNode(node) {
				klog.V(4).Infof("bi.EnsureHostsInPool: skipping control plane node %s", node.Name)
				continue
			}

			privateIP := getNodePrivateIPAddress(node, isIPv6)
			nodePrivateIPsSet.Insert(privateIP)

			if bi.useMultipleStandardLoadBalancers() {
				if !activeNodes.Has(node.Name) {
					klog.V(4).Infof("bi.EnsureHostsInPool: node %s should not be in load balancer %q", node.Name, lbName)
					continue
				}
			}

			if !existingIPs.Has(privateIP) {
				name := node.Name
				klog.V(6).Infof("bi.EnsureHostsInPool: adding %s with ip address %s", name, privateIP)
				nodeIPsToBeAdded = append(nodeIPsToBeAdded, privateIP)
				numOfAdd++
			}
		}
		changed = bi.addNodeIPAddressesToBackendPool(&backendPool, nodeIPsToBeAdded)

		var nodeIPsToBeDeleted []string
		for _, loadBalancerBackendAddress := range *backendPool.LoadBalancerBackendAddresses {
			ip := pointer.StringDeref(loadBalancerBackendAddress.IPAddress, "")
			if !nodePrivateIPsSet.Has(ip) {
				klog.V(4).Infof("bi.EnsureHostsInPool: removing IP %s because it is deleted or should be excluded", ip)
				nodeIPsToBeDeleted = append(nodeIPsToBeDeleted, ip)
				changed = true
				numOfDelete++
			} else if bi.useMultipleStandardLoadBalancers() && activeNodes != nil {
				nodeName, ok := bi.nodePrivateIPToNodeNameMap[ip]
				if !ok {
					klog.Warningf("bi.EnsureHostsInPool: cannot find node name for private IP %s", ip)
					continue
				}
				if !activeNodes.Has(nodeName) {
					klog.V(4).Infof("bi.EnsureHostsInPool: removing IP %s because it should not be in this load balancer", ip)
					nodeIPsToBeDeleted = append(nodeIPsToBeDeleted, ip)
					changed = true
					numOfDelete++
				}
			}
		}
		removeNodeIPAddressesFromBackendPool(backendPool, nodeIPsToBeDeleted, false, bi.useMultipleStandardLoadBalancers())
	}
	if changed {
		klog.V(2).Infof("bi.EnsureHostsInPool: updating backend pool %s of load balancer %s to add %d nodes and remove %d nodes", lbBackendPoolName, lbName, numOfAdd, numOfDelete)
		if err := bi.CreateOrUpdateLBBackendPool(lbName, backendPool); err != nil {
			return fmt.Errorf("bi.EnsureHostsInPool: failed to update backend pool %s: %w", lbBackendPoolName, err)
		}
	}

	return nil
}

func (bi *backendPoolTypeNodeIP) CleanupVMSetFromBackendPoolByCondition(slb *network.LoadBalancer, _ *v1.Service, nodes []*v1.Node, clusterName string, shouldRemoveVMSetFromSLB func(string) bool) (*network.LoadBalancer, error) {
	lbBackendPoolNames := getBackendPoolNames(clusterName)
	newBackendPools := make([]network.BackendAddressPool, 0)
	if slb.LoadBalancerPropertiesFormat != nil && slb.BackendAddressPools != nil {
		newBackendPools = *slb.BackendAddressPools
	}

	updatedPrivateIPs := map[bool]bool{}
	for j, bp := range newBackendPools {
		found, isIPv6 := isLBBackendPoolsExisting(lbBackendPoolNames, bp.Name)
		if found {
			klog.V(2).Infof("bi.CleanupVMSetFromBackendPoolByCondition: checking the backend pool %s from standard load balancer %s", pointer.StringDeref(bp.Name, ""), pointer.StringDeref(slb.Name, ""))
			vmIPsToBeDeleted := utilsets.NewString()
			for _, node := range nodes {
				vmSetName, err := bi.VMSet.GetNodeVMSetName(node)
				if err != nil {
					return nil, err
				}

				if shouldRemoveVMSetFromSLB(vmSetName) {
					privateIP := getNodePrivateIPAddress(node, isIPv6)
					klog.V(4).Infof("bi.CleanupVMSetFromBackendPoolByCondition: removing ip %s from the backend pool %s", privateIP, lbBackendPoolNames[isIPv6])
					vmIPsToBeDeleted.Insert(privateIP)
				}
			}

			if bp.BackendAddressPoolPropertiesFormat != nil && bp.LoadBalancerBackendAddresses != nil {
				for i := len(*bp.LoadBalancerBackendAddresses) - 1; i >= 0; i-- {
					if (*bp.LoadBalancerBackendAddresses)[i].LoadBalancerBackendAddressPropertiesFormat != nil &&
						vmIPsToBeDeleted.Has(pointer.StringDeref((*bp.LoadBalancerBackendAddresses)[i].IPAddress, "")) {
						*bp.LoadBalancerBackendAddresses = append((*bp.LoadBalancerBackendAddresses)[:i], (*bp.LoadBalancerBackendAddresses)[i+1:]...)
						updatedPrivateIPs[isIPv6] = true
					}
				}
			}

			newBackendPools[j] = bp
		} else {
			klog.V(10).Infof("bi.CleanupVMSetFromBackendPoolByCondition: found unmanaged backendpool %s from standard load balancer %q", pointer.StringDeref(bp.Name, ""), pointer.StringDeref(slb.Name, ""))
		}

	}
	for isIPv6 := range updatedPrivateIPs {
		klog.V(2).Infof("bi.CleanupVMSetFromBackendPoolByCondition: updating lb %s since there are private IP updates", pointer.StringDeref(slb.Name, ""))
		slb.BackendAddressPools = &newBackendPools

		for _, backendAddressPool := range *slb.BackendAddressPools {
			if strings.EqualFold(lbBackendPoolNames[isIPv6], pointer.StringDeref(backendAddressPool.Name, "")) {
				if err := bi.CreateOrUpdateLBBackendPool(pointer.StringDeref(slb.Name, ""), backendAddressPool); err != nil {
					return nil, fmt.Errorf("bi.CleanupVMSetFromBackendPoolByCondition: "+
						"failed to create or update backend pool %s: %w", lbBackendPoolNames[isIPv6], err)
				}
			}
		}
	}

	return slb, nil
}

func (bi *backendPoolTypeNodeIP) ReconcileBackendPools(clusterName string, service *v1.Service, lb *network.LoadBalancer) (bool, bool, *network.LoadBalancer, error) {
	var newBackendPools []network.BackendAddressPool
	if lb.BackendAddressPools != nil {
		newBackendPools = *lb.BackendAddressPools
	}

	var backendPoolsUpdated, shouldRefreshLB, isOperationSucceeded, isMigration, updated bool
	foundBackendPools := map[bool]bool{}
	lbName := *lb.Name
	serviceName := getServiceName(service)
	lbBackendPoolNames := bi.getBackendPoolNamesForService(service, clusterName)
	vmSetName := bi.mapLoadBalancerNameToVMSet(lbName, clusterName)
	lbBackendPoolIDs := bi.getBackendPoolIDsForService(service, clusterName, pointer.StringDeref(lb.Name, ""))
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
			klog.V(10).Infof("bi.ReconcileBackendPools for service (%s): found wanted backendpool. Not adding anything", serviceName)
			foundBackendPools[isIPv6] = true
			lbBackendPoolIDsSlice = append(lbBackendPoolIDsSlice, lbBackendPoolIDs[isIPv6])

			if nicsCount := countNICsOnBackendPool(bp); nicsCount > 0 {
				nicsCountMap[pointer.StringDeref(bp.Name, "")] = nicsCount
				klog.V(4).Infof(
					"bi.ReconcileBackendPools for service(%s): found NIC-based backendpool %s with %d NICs, will migrate to IP-based",
					serviceName,
					pointer.StringDeref(bp.Name, ""),
					nicsCount,
				)
				isMigration = true
			}
		} else {
			klog.V(10).Infof("bi.ReconcileBackendPools for service (%s): found unmanaged backendpool %s", serviceName, *bp.Name)
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
				name, err := getLBNameFromBackendPoolID(id)
				if err != nil {
					klog.Errorf("bi.ReconcileBackendPools for service (%s): failed to get LB name from backend pool ID: %s", serviceName, err.Error())
					return false, false, nil, err
				}
				backendPoolNames = append(backendPoolNames, name)
			}

			if err := bi.MigrateToIPBasedBackendPoolAndWaitForCompletion(lbName, backendPoolNames, nicsCountMap); err != nil {
				backendPoolNamesStr := strings.Join(backendPoolNames, ",")
				klog.Errorf("Failed to migrate to IP based backend pool for lb %s, backend pool %s: %s", lbName, backendPoolNamesStr, err.Error())
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
		klog.V(2).Infof("bi.ReconcileBackendPools for service (%s) and vmSet (%s): ensuring the LB is decoupled from the VMSet", serviceName, vmSetName)
		shouldRefreshLB, err = bi.VMSet.EnsureBackendPoolDeleted(service, lbBackendPoolIDsSlice, vmSetName, lb.BackendAddressPools, true)
		if err != nil {
			klog.Errorf("bi.ReconcileBackendPools for service (%s): failed to EnsureBackendPoolDeleted: %s", serviceName, err.Error())
			return false, false, nil, err
		}

		for _, i := range bpIdxes {
			bp := newBackendPools[i]
			var nodeIPAddressesToBeDeleted []string
			for _, nodeName := range bi.excludeLoadBalancerNodes.UnsortedList() {
				for _, ip := range bi.nodePrivateIPs[strings.ToLower(nodeName)].UnsortedList() {
					klog.V(2).Infof("bi.ReconcileBackendPools for service (%s): found unwanted node private IP %s, decouple it from the LB %s", serviceName, ip, lbName)
					nodeIPAddressesToBeDeleted = append(nodeIPAddressesToBeDeleted, ip)
				}
			}
			if len(nodeIPAddressesToBeDeleted) > 0 {
				if removeNodeIPAddressesFromBackendPool(bp, nodeIPAddressesToBeDeleted, false, false) {
					updated = true
				}
			}
			// delete the vnet in LoadBalancerBackendAddresses and ensure it is in the backend pool level
			var vnet string
			if bp.BackendAddressPoolPropertiesFormat != nil {
				if bp.VirtualNetwork == nil ||
					pointer.StringDeref(bp.VirtualNetwork.ID, "") == "" {
					if bp.LoadBalancerBackendAddresses != nil {
						for _, a := range *bp.LoadBalancerBackendAddresses {
							if a.LoadBalancerBackendAddressPropertiesFormat != nil &&
								a.VirtualNetwork != nil {
								if vnet == "" {
									vnet = pointer.StringDeref(a.VirtualNetwork.ID, "")
								}
								a.VirtualNetwork = nil
							}
						}
					}
					if vnet != "" {
						bp.VirtualNetwork = &network.SubResource{
							ID: pointer.String(vnet),
						}
						updated = true
					}
				}
			}

			if updated {
				(*lb.BackendAddressPools)[i] = bp
				if err := bi.CreateOrUpdateLBBackendPool(lbName, bp); err != nil {
					return false, false, nil, fmt.Errorf("bi.ReconcileBackendPools for service (%s): lb backendpool - failed to update backend pool %s for load balancer %s: %w", serviceName, pointer.StringDeref(bp.Name, ""), lbName, err)
				}
				shouldRefreshLB = true
			}
		}
	}

	shouldRefreshLB = shouldRefreshLB || isMigration
	if shouldRefreshLB {
		klog.V(4).Infof("bi.ReconcileBackendPools for service(%s): refreshing load balancer %s", serviceName, lbName)
		lb, _, err = bi.getAzureLoadBalancer(lbName, cache.CacheReadTypeForceRefresh)
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

func (bi *backendPoolTypeNodeIP) GetBackendPrivateIPs(clusterName string, service *v1.Service, lb *network.LoadBalancer) ([]string, []string) {
	serviceName := getServiceName(service)
	lbBackendPoolNames := bi.getBackendPoolNamesForService(service, clusterName)
	if lb.LoadBalancerPropertiesFormat == nil || lb.LoadBalancerPropertiesFormat.BackendAddressPools == nil {
		return nil, nil
	}

	backendPrivateIPv4s, backendPrivateIPv6s := utilsets.NewString(), utilsets.NewString()
	for _, bp := range *lb.BackendAddressPools {
		found, _ := isLBBackendPoolsExisting(lbBackendPoolNames, bp.Name)
		if found {
			klog.V(10).Infof("bi.GetBackendPrivateIPs for service (%s): found wanted backendpool %s", serviceName, pointer.StringDeref(bp.Name, ""))
			if bp.BackendAddressPoolPropertiesFormat != nil && bp.LoadBalancerBackendAddresses != nil {
				for _, backendAddress := range *bp.LoadBalancerBackendAddresses {
					ipAddress := backendAddress.IPAddress
					if ipAddress != nil {
						klog.V(2).Infof("bi.GetBackendPrivateIPs for service (%s): lb backendpool - found private IP %q", serviceName, *ipAddress)
						if utilnet.IsIPv4String(*ipAddress) {
							backendPrivateIPv4s.Insert(*ipAddress)
						} else if utilnet.IsIPv6String(*ipAddress) {
							backendPrivateIPv6s.Insert(*ipAddress)
						}
					} else {
						klog.V(4).Infof("bi.GetBackendPrivateIPs for service (%s): lb backendpool - found null private IP", serviceName)
					}
				}
			}
		} else {
			klog.V(10).Infof("bi.GetBackendPrivateIPs for service (%s): found unmanaged backendpool %s", serviceName, pointer.StringDeref(bp.Name, ""))
		}
	}
	return backendPrivateIPv4s.UnsortedList(), backendPrivateIPv6s.UnsortedList()
}

func newBackendPool(lb *network.LoadBalancer, isBackendPoolPreConfigured bool, preConfiguredBackendPoolLoadBalancerTypes, serviceName, lbBackendPoolName string) bool {
	if isBackendPoolPreConfigured {
		klog.V(2).Infof("newBackendPool for service (%s)(true): lb backendpool - PreConfiguredBackendPoolLoadBalancerTypes %s has been set but can not find corresponding backend pool %q, ignoring it",
			serviceName,
			preConfiguredBackendPoolLoadBalancerTypes,
			lbBackendPoolName)
		isBackendPoolPreConfigured = false
	}

	if lb.BackendAddressPools == nil {
		lb.BackendAddressPools = &[]network.BackendAddressPool{}
	}
	*lb.BackendAddressPools = append(*lb.BackendAddressPools, network.BackendAddressPool{
		Name:                               pointer.String(lbBackendPoolName),
		BackendAddressPoolPropertiesFormat: &network.BackendAddressPoolPropertiesFormat{},
	})

	// Always returns false
	return isBackendPoolPreConfigured
}

func (az *Cloud) addNodeIPAddressesToBackendPool(backendPool *network.BackendAddressPool, nodeIPAddresses []string) bool {
	if backendPool.LoadBalancerBackendAddresses == nil {
		lbBackendPoolAddresses := make([]network.LoadBalancerBackendAddress, 0)
		backendPool.LoadBalancerBackendAddresses = &lbBackendPoolAddresses
	}

	var changed bool
	addresses := *backendPool.LoadBalancerBackendAddresses
	for _, ipAddress := range nodeIPAddresses {
		if !hasIPAddressInBackendPool(backendPool, ipAddress) {
			name := az.nodePrivateIPToNodeNameMap[ipAddress]
			klog.V(4).Infof("bi.addNodeIPAddressesToBackendPool: adding %s to the backend pool %s", ipAddress, pointer.StringDeref(backendPool.Name, ""))
			addresses = append(addresses, network.LoadBalancerBackendAddress{
				Name: pointer.String(name),
				LoadBalancerBackendAddressPropertiesFormat: &network.LoadBalancerBackendAddressPropertiesFormat{
					IPAddress: pointer.String(ipAddress),
				},
			})
			changed = true
		}
	}
	backendPool.LoadBalancerBackendAddresses = &addresses
	return changed
}

func hasIPAddressInBackendPool(backendPool *network.BackendAddressPool, ipAddress string) bool {
	if backendPool.LoadBalancerBackendAddresses == nil {
		return false
	}

	addresses := *backendPool.LoadBalancerBackendAddresses
	for _, address := range addresses {
		if address.LoadBalancerBackendAddressPropertiesFormat != nil &&
			pointer.StringDeref(address.IPAddress, "") == ipAddress {
			return true
		}
	}

	return false
}

func removeNodeIPAddressesFromBackendPool(
	backendPool network.BackendAddressPool,
	nodeIPAddresses []string,
	removeAll, useMultipleStandardLoadBalancers bool,
) bool {
	changed := false
	nodeIPsSet := utilsets.NewString(nodeIPAddresses...)

	if backendPool.BackendAddressPoolPropertiesFormat == nil ||
		backendPool.LoadBalancerBackendAddresses == nil {
		return false
	}

	addresses := *backendPool.LoadBalancerBackendAddresses
	for i := len(addresses) - 1; i >= 0; i-- {
		if addresses[i].LoadBalancerBackendAddressPropertiesFormat != nil {
			ipAddress := pointer.StringDeref((*backendPool.LoadBalancerBackendAddresses)[i].IPAddress, "")
			if ipAddress == "" {
				klog.V(4).Infof("removeNodeIPAddressFromBackendPool: LoadBalancerBackendAddress %s is not IP-based, skipping", pointer.StringDeref(addresses[i].Name, ""))
				continue
			}
			if removeAll || nodeIPsSet.Has(ipAddress) {
				klog.V(4).Infof("removeNodeIPAddressFromBackendPool: removing %s from the backend pool %s", ipAddress, pointer.StringDeref(backendPool.Name, ""))
				addresses = append(addresses[:i], addresses[i+1:]...)
				changed = true
			}
		}
	}

	if removeAll {
		backendPool.LoadBalancerBackendAddresses = &addresses
		return changed
	}

	// Allow the pool to be empty when EnsureHostsInPool for multiple standard load balancers clusters,
	// or one node could occur in multiple backend pools.
	if len(addresses) == 0 && !useMultipleStandardLoadBalancers {
		klog.V(2).Info("removeNodeIPAddressFromBackendPool: the pool is empty or will be empty after removing the unwanted IP addresses, skipping the removal")
		changed = false
	} else if changed {
		backendPool.LoadBalancerBackendAddresses = &addresses
	}

	return changed
}
