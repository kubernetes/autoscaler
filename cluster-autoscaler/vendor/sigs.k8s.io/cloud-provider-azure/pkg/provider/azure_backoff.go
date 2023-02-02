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
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2022-03-01/compute"
	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2021-08-01/network"
	"github.com/Azure/go-autorest/autorest/to"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	cloudprovider "k8s.io/cloud-provider"
	"k8s.io/klog/v2"

	azcache "sigs.k8s.io/cloud-provider-azure/pkg/cache"
	"sigs.k8s.io/cloud-provider-azure/pkg/consts"
	"sigs.k8s.io/cloud-provider-azure/pkg/retry"
)

var (
	pipErrorMessageRE = regexp.MustCompile(`(?:.*)/subscriptions/(?:.*)/resourceGroups/(.*)/providers/Microsoft.Network/publicIPAddresses/([^\s]+)(?:.*)`)
)

// RequestBackoff if backoff is disabled in cloud provider it
// returns a new Backoff object steps = 1
// This is to make sure that the requested command executes
// at least once
func (az *Cloud) RequestBackoff() (resourceRequestBackoff wait.Backoff) {
	if az.CloudProviderBackoff {
		return az.ResourceRequestBackoff
	}
	resourceRequestBackoff = wait.Backoff{
		Steps: 1,
	}
	return resourceRequestBackoff
}

// Event creates a event for the specified object.
func (az *Cloud) Event(obj runtime.Object, eventType, reason, message string) {
	if obj != nil && reason != "" {
		az.eventRecorder.Event(obj, eventType, reason, message)
	}
}

// GetVirtualMachineWithRetry invokes az.getVirtualMachine with exponential backoff retry
func (az *Cloud) GetVirtualMachineWithRetry(name types.NodeName, crt azcache.AzureCacheReadType) (compute.VirtualMachine, error) {
	var machine compute.VirtualMachine
	var retryErr error
	err := wait.ExponentialBackoff(az.RequestBackoff(), func() (bool, error) {
		machine, retryErr = az.getVirtualMachine(name, crt)
		if errors.Is(retryErr, cloudprovider.InstanceNotFound) {
			return true, cloudprovider.InstanceNotFound
		}
		if retryErr != nil {
			klog.Errorf("GetVirtualMachineWithRetry(%s): backoff failure, will retry, err=%v", name, retryErr)
			return false, nil
		}
		klog.V(2).Infof("GetVirtualMachineWithRetry(%s): backoff success", name)
		return true, nil
	})
	if errors.Is(err, wait.ErrWaitTimeout) {
		err = retryErr
	}
	return machine, err
}

// ListVirtualMachines invokes az.VirtualMachinesClient.List with exponential backoff retry
func (az *Cloud) ListVirtualMachines(resourceGroup string) ([]compute.VirtualMachine, error) {
	ctx, cancel := getContextWithCancel()
	defer cancel()

	allNodes, rerr := az.VirtualMachinesClient.List(ctx, resourceGroup)
	if rerr != nil {
		klog.Errorf("VirtualMachinesClient.List(%v) failure with err=%v", resourceGroup, rerr)
		return nil, rerr.Error()
	}
	klog.V(6).Infof("VirtualMachinesClient.List(%v) success", resourceGroup)
	return allNodes, nil
}

// getPrivateIPsForMachine is wrapper for optional backoff getting private ips
// list of a node by name
func (az *Cloud) getPrivateIPsForMachine(nodeName types.NodeName) ([]string, error) {
	return az.getPrivateIPsForMachineWithRetry(nodeName)
}

func (az *Cloud) getPrivateIPsForMachineWithRetry(nodeName types.NodeName) ([]string, error) {
	var privateIPs []string
	err := wait.ExponentialBackoff(az.RequestBackoff(), func() (bool, error) {
		var retryErr error
		privateIPs, retryErr = az.VMSet.GetPrivateIPsByNodeName(string(nodeName))
		if retryErr != nil {
			// won't retry since the instance doesn't exist on Azure.
			if errors.Is(retryErr, cloudprovider.InstanceNotFound) {
				return true, retryErr
			}
			klog.Errorf("GetPrivateIPsByNodeName(%s): backoff failure, will retry,err=%v", nodeName, retryErr)
			return false, nil
		}
		klog.V(3).Infof("GetPrivateIPsByNodeName(%s): backoff success", nodeName)
		return true, nil
	})
	return privateIPs, err
}

func (az *Cloud) getIPForMachine(nodeName types.NodeName) (string, string, error) {
	return az.GetIPForMachineWithRetry(nodeName)
}

// GetIPForMachineWithRetry invokes az.getIPForMachine with exponential backoff retry
func (az *Cloud) GetIPForMachineWithRetry(name types.NodeName) (string, string, error) {
	var ip, publicIP string
	err := wait.ExponentialBackoff(az.RequestBackoff(), func() (bool, error) {
		var retryErr error
		ip, publicIP, retryErr = az.VMSet.GetIPByNodeName(string(name))
		if retryErr != nil {
			klog.Errorf("GetIPForMachineWithRetry(%s): backoff failure, will retry,err=%v", name, retryErr)
			return false, nil
		}
		klog.V(3).Infof("GetIPForMachineWithRetry(%s): backoff success", name)
		return true, nil
	})
	return ip, publicIP, err
}

// CreateOrUpdateSecurityGroup invokes az.SecurityGroupsClient.CreateOrUpdate with exponential backoff retry
func (az *Cloud) CreateOrUpdateSecurityGroup(sg network.SecurityGroup) error {
	ctx, cancel := getContextWithCancel()
	defer cancel()

	rerr := az.SecurityGroupsClient.CreateOrUpdate(ctx, az.SecurityGroupResourceGroup, *sg.Name, sg, to.String(sg.Etag))
	klog.V(10).Infof("SecurityGroupsClient.CreateOrUpdate(%s): end", *sg.Name)
	if rerr == nil {
		// Invalidate the cache right after updating
		_ = az.nsgCache.Delete(*sg.Name)
		return nil
	}

	nsgJSON, _ := json.Marshal(sg)
	klog.Warningf("CreateOrUpdateSecurityGroup(%s) failed: %v, NSG request: %s", to.String(sg.Name), rerr.Error(), string(nsgJSON))

	// Invalidate the cache because ETAG precondition mismatch.
	if rerr.HTTPStatusCode == http.StatusPreconditionFailed {
		klog.V(3).Infof("SecurityGroup cache for %s is cleanup because of http.StatusPreconditionFailed", *sg.Name)
		_ = az.nsgCache.Delete(*sg.Name)
	}

	// Invalidate the cache because another new operation has canceled the current request.
	if strings.Contains(strings.ToLower(rerr.Error().Error()), consts.OperationCanceledErrorMessage) {
		klog.V(3).Infof("SecurityGroup cache for %s is cleanup because CreateOrUpdateSecurityGroup is canceled by another operation", *sg.Name)
		_ = az.nsgCache.Delete(*sg.Name)
	}

	return rerr.Error()
}

func cleanupSubnetInFrontendIPConfigurations(lb *network.LoadBalancer) network.LoadBalancer {
	if lb.LoadBalancerPropertiesFormat == nil || lb.FrontendIPConfigurations == nil {
		return *lb
	}

	frontendIPConfigurations := *lb.FrontendIPConfigurations
	for i := range frontendIPConfigurations {
		config := frontendIPConfigurations[i]
		if config.FrontendIPConfigurationPropertiesFormat != nil &&
			config.Subnet != nil &&
			config.Subnet.ID != nil {
			subnet := network.Subnet{
				ID: config.Subnet.ID,
			}
			if config.Subnet.Name != nil {
				subnet.Name = config.FrontendIPConfigurationPropertiesFormat.Subnet.Name
			}
			config.FrontendIPConfigurationPropertiesFormat.Subnet = &subnet
			frontendIPConfigurations[i] = config
			continue
		}
	}

	lb.FrontendIPConfigurations = &frontendIPConfigurations
	return *lb
}

// CreateOrUpdateLB invokes az.LoadBalancerClient.CreateOrUpdate with exponential backoff retry
func (az *Cloud) CreateOrUpdateLB(service *v1.Service, lb network.LoadBalancer) error {
	ctx, cancel := getContextWithCancel()
	defer cancel()

	lb = cleanupSubnetInFrontendIPConfigurations(&lb)

	rgName := az.getLoadBalancerResourceGroup()
	rerr := az.LoadBalancerClient.CreateOrUpdate(ctx, rgName, to.String(lb.Name), lb, to.String(lb.Etag))
	klog.V(10).Infof("LoadBalancerClient.CreateOrUpdate(%s): end", *lb.Name)
	if rerr == nil {
		// Invalidate the cache right after updating
		_ = az.lbCache.Delete(*lb.Name)
		return nil
	}

	lbJSON, _ := json.Marshal(lb)
	klog.Warningf("LoadBalancerClient.CreateOrUpdate(%s) failed: %v, LoadBalancer request: %s", to.String(lb.Name), rerr.Error(), string(lbJSON))

	// Invalidate the cache because ETAG precondition mismatch.
	if rerr.HTTPStatusCode == http.StatusPreconditionFailed {
		klog.V(3).Infof("LoadBalancer cache for %s is cleanup because of http.StatusPreconditionFailed", to.String(lb.Name))
		_ = az.lbCache.Delete(*lb.Name)
	}

	retryErrorMessage := rerr.Error().Error()
	// Invalidate the cache because another new operation has canceled the current request.
	if strings.Contains(strings.ToLower(retryErrorMessage), consts.OperationCanceledErrorMessage) {
		klog.V(3).Infof("LoadBalancer cache for %s is cleanup because CreateOrUpdate is canceled by another operation", to.String(lb.Name))
		_ = az.lbCache.Delete(*lb.Name)
	}

	// The LB update may fail because the referenced PIP is not in the Succeeded provisioning state
	if strings.Contains(strings.ToLower(retryErrorMessage), strings.ToLower(consts.ReferencedResourceNotProvisionedMessageCode)) {
		matches := pipErrorMessageRE.FindStringSubmatch(retryErrorMessage)
		if len(matches) != 3 {
			klog.Errorf("Failed to parse the retry error message %s", retryErrorMessage)
			return rerr.Error()
		}
		pipRG, pipName := matches[1], matches[2]
		klog.V(3).Infof("The public IP %s referenced by load balancer %s is not in Succeeded provisioning state, will try to update it", pipName, to.String(lb.Name))
		pip, _, err := az.getPublicIPAddress(pipRG, pipName, azcache.CacheReadTypeDefault)
		if err != nil {
			klog.Errorf("Failed to get the public IP %s in resource group %s: %v", pipName, pipRG, err)
			return rerr.Error()
		}
		// Perform a dummy update to fix the provisioning state
		err = az.CreateOrUpdatePIP(service, pipRG, pip)
		if err != nil {
			klog.Errorf("Failed to update the public IP %s in resource group %s: %v", pipName, pipRG, err)
			return rerr.Error()
		}
		// Invalidate the LB cache, return the error, and the controller manager
		// would retry the LB update in the next reconcile loop
		_ = az.lbCache.Delete(*lb.Name)
	}

	return rerr.Error()
}

func (az *Cloud) CreateOrUpdateLBBackendPool(lbName string, backendPool network.BackendAddressPool) error {
	ctx, cancel := getContextWithCancel()
	defer cancel()

	klog.V(4).Infof("CreateOrUpdateLBBackendPool: updating backend pool %s in LB %s", to.String(backendPool.Name), lbName)
	rerr := az.LoadBalancerClient.CreateOrUpdateBackendPools(ctx, az.getLoadBalancerResourceGroup(), lbName, to.String(backendPool.Name), backendPool, to.String(backendPool.Etag))
	if rerr == nil {
		// Invalidate the cache right after updating
		_ = az.lbCache.Delete(lbName)
		return nil
	}

	// Invalidate the cache because ETAG precondition mismatch.
	if rerr.HTTPStatusCode == http.StatusPreconditionFailed {
		klog.V(3).Infof("LoadBalancer cache for %s is cleanup because of http.StatusPreconditionFailed", lbName)
		_ = az.lbCache.Delete(lbName)
	}

	retryErrorMessage := rerr.Error().Error()
	// Invalidate the cache because another new operation has canceled the current request.
	if strings.Contains(strings.ToLower(retryErrorMessage), consts.OperationCanceledErrorMessage) {
		klog.V(3).Infof("LoadBalancer cache for %s is cleanup because CreateOrUpdate is canceled by another operation", lbName)
		_ = az.lbCache.Delete(lbName)
	}

	return rerr.Error()
}

func (az *Cloud) DeleteLBBackendPool(lbName, backendPoolName string) error {
	ctx, cancel := getContextWithCancel()
	defer cancel()

	klog.V(4).Infof("DeleteLBBackendPool: deleting backend pool %s in LB %s", backendPoolName, lbName)
	rerr := az.LoadBalancerClient.DeleteLBBackendPool(ctx, az.getLoadBalancerResourceGroup(), lbName, backendPoolName)
	if rerr == nil {
		// Invalidate the cache right after updating
		_ = az.lbCache.Delete(lbName)
		return nil
	}

	// Invalidate the cache because ETAG precondition mismatch.
	if rerr.HTTPStatusCode == http.StatusPreconditionFailed {
		klog.V(3).Infof("LoadBalancer cache for %s is cleanup because of http.StatusPreconditionFailed", lbName)
		_ = az.lbCache.Delete(lbName)
	}

	retryErrorMessage := rerr.Error().Error()
	// Invalidate the cache because another new operation has canceled the current request.
	if strings.Contains(strings.ToLower(retryErrorMessage), consts.OperationCanceledErrorMessage) {
		klog.V(3).Infof("LoadBalancer cache for %s is cleanup because CreateOrUpdate is canceled by another operation", lbName)
		_ = az.lbCache.Delete(lbName)
	}

	return rerr.Error()
}

// ListManagedLBs invokes az.LoadBalancerClient.List and filter out
// those that are not managed by cloud provider azure or not associated to a managed VMSet.
func (az *Cloud) ListManagedLBs(service *v1.Service, nodes []*v1.Node, clusterName string) ([]network.LoadBalancer, error) {
	allLBs, err := az.ListLB(service)
	if err != nil {
		return nil, err
	}

	if allLBs == nil {
		klog.Warningf("ListManagedLBs: no LBs found")
		return nil, nil
	}

	// return early if wantLb=false
	if nodes == nil {
		klog.V(4).Infof("ListManagedLBs: return all LBs in the resource group %s, including unmanaged LBs", az.getLoadBalancerResourceGroup())
		return allLBs, nil
	}

	agentPoolLBs := make([]network.LoadBalancer, 0)
	agentPoolVMSetNames, err := az.VMSet.GetAgentPoolVMSetNames(nodes)
	if err != nil {
		return nil, fmt.Errorf("ListManagedLBs: failed to get agent pool vmSet names: %w", err)
	}

	agentPoolVMSetNamesSet := sets.NewString()
	if agentPoolVMSetNames != nil && len(*agentPoolVMSetNames) > 0 {
		for _, vmSetName := range *agentPoolVMSetNames {
			klog.V(6).Infof("ListManagedLBs: found agent pool vmSet name %s", vmSetName)
			agentPoolVMSetNamesSet.Insert(strings.ToLower(vmSetName))
		}
	}

	for _, lb := range allLBs {
		vmSetNameFromLBName := az.mapLoadBalancerNameToVMSet(to.String(lb.Name), clusterName)
		if strings.EqualFold(strings.TrimSuffix(to.String(lb.Name), consts.InternalLoadBalancerNameSuffix), clusterName) ||
			agentPoolVMSetNamesSet.Has(strings.ToLower(vmSetNameFromLBName)) {
			agentPoolLBs = append(agentPoolLBs, lb)
			klog.V(4).Infof("ListManagedLBs: found agent pool LB %s", to.String(lb.Name))
		}
	}

	return agentPoolLBs, nil
}

// ListLB invokes az.LoadBalancerClient.List with exponential backoff retry
func (az *Cloud) ListLB(service *v1.Service) ([]network.LoadBalancer, error) {
	ctx, cancel := getContextWithCancel()
	defer cancel()

	rgName := az.getLoadBalancerResourceGroup()
	allLBs, rerr := az.LoadBalancerClient.List(ctx, rgName)
	if rerr != nil {
		if rerr.IsNotFound() {
			return nil, nil
		}
		az.Event(service, v1.EventTypeWarning, "ListLoadBalancers", rerr.Error().Error())
		klog.Errorf("LoadBalancerClient.List(%v) failure with err=%v", rgName, rerr)
		return nil, rerr.Error()
	}
	klog.V(2).Infof("LoadBalancerClient.List(%v) success", rgName)
	return allLBs, nil
}

// ListPIP list the PIP resources in the given resource group
func (az *Cloud) ListPIP(service *v1.Service, pipResourceGroup string) ([]network.PublicIPAddress, error) {
	ctx, cancel := getContextWithCancel()
	defer cancel()

	allPIPs, rerr := az.PublicIPAddressesClient.List(ctx, pipResourceGroup)
	if rerr != nil {
		if rerr.IsNotFound() {
			return nil, nil
		}
		az.Event(service, v1.EventTypeWarning, "ListPublicIPs", rerr.Error().Error())
		klog.Errorf("PublicIPAddressesClient.List(%v) failure with err=%v", pipResourceGroup, rerr)
		return nil, rerr.Error()
	}

	klog.V(2).Infof("PublicIPAddressesClient.List(%v) success", pipResourceGroup)
	return allPIPs, nil
}

// CreateOrUpdatePIP invokes az.PublicIPAddressesClient.CreateOrUpdate with exponential backoff retry
func (az *Cloud) CreateOrUpdatePIP(service *v1.Service, pipResourceGroup string, pip network.PublicIPAddress) error {
	ctx, cancel := getContextWithCancel()
	defer cancel()

	rerr := az.PublicIPAddressesClient.CreateOrUpdate(ctx, pipResourceGroup, to.String(pip.Name), pip)
	klog.V(10).Infof("PublicIPAddressesClient.CreateOrUpdate(%s, %s): end", pipResourceGroup, to.String(pip.Name))
	if rerr == nil {
		// Invalidate the cache right after updating
		_ = az.pipCache.Delete(az.getPIPCacheKey(pipResourceGroup, to.String(pip.Name)))
		return nil
	}

	pipJSON, _ := json.Marshal(pip)
	klog.Warningf("PublicIPAddressesClient.CreateOrUpdate(%s, %s) failed: %s, PublicIP request: %s", pipResourceGroup, to.String(pip.Name), rerr.Error().Error(), string(pipJSON))
	az.Event(service, v1.EventTypeWarning, "CreateOrUpdatePublicIPAddress", rerr.Error().Error())

	// Invalidate the cache because ETAG precondition mismatch.
	if rerr.HTTPStatusCode == http.StatusPreconditionFailed {
		klog.V(3).Infof("PublicIP cache for (%s, %s) is cleanup because of http.StatusPreconditionFailed", pipResourceGroup, to.String(pip.Name))
		_ = az.pipCache.Delete(az.getPIPCacheKey(pipResourceGroup, to.String(pip.Name)))
	}

	retryErrorMessage := rerr.Error().Error()
	// Invalidate the cache because another new operation has canceled the current request.
	if strings.Contains(strings.ToLower(retryErrorMessage), consts.OperationCanceledErrorMessage) {
		klog.V(3).Infof("PublicIP cache for (%s, %s) is cleanup because CreateOrUpdate is canceled by another operation", pipResourceGroup, to.String(pip.Name))
		_ = az.pipCache.Delete(az.getPIPCacheKey(pipResourceGroup, to.String(pip.Name)))
	}

	return rerr.Error()
}

// CreateOrUpdateInterface invokes az.InterfacesClient.CreateOrUpdate with exponential backoff retry
func (az *Cloud) CreateOrUpdateInterface(service *v1.Service, nic network.Interface) error {
	ctx, cancel := getContextWithCancel()
	defer cancel()

	rerr := az.InterfacesClient.CreateOrUpdate(ctx, az.ResourceGroup, *nic.Name, nic)
	klog.V(10).Infof("InterfacesClient.CreateOrUpdate(%s): end", *nic.Name)
	if rerr != nil {
		klog.Errorf("InterfacesClient.CreateOrUpdate(%s) failed: %s", *nic.Name, rerr.Error().Error())
		az.Event(service, v1.EventTypeWarning, "CreateOrUpdateInterface", rerr.Error().Error())
		return rerr.Error()
	}

	return nil
}

// DeletePublicIP invokes az.PublicIPAddressesClient.Delete with exponential backoff retry
func (az *Cloud) DeletePublicIP(service *v1.Service, pipResourceGroup string, pipName string) error {
	ctx, cancel := getContextWithCancel()
	defer cancel()

	rerr := az.PublicIPAddressesClient.Delete(ctx, pipResourceGroup, pipName)
	if rerr != nil {
		klog.Errorf("PublicIPAddressesClient.Delete(%s) failed: %s", pipName, rerr.Error().Error())
		az.Event(service, v1.EventTypeWarning, "DeletePublicIPAddress", rerr.Error().Error())

		if strings.Contains(rerr.Error().Error(), consts.CannotDeletePublicIPErrorMessageCode) {
			klog.Warningf("DeletePublicIP for public IP %s failed with error %v, this is because other resources are referencing the public IP. The deletion of the service will continue.", pipName, rerr.Error())
			return nil
		}
		return rerr.Error()
	}

	// Invalidate the cache right after deleting
	_ = az.pipCache.Delete(az.getPIPCacheKey(pipResourceGroup, pipName))
	return nil
}

// DeleteLB invokes az.LoadBalancerClient.Delete with exponential backoff retry
func (az *Cloud) DeleteLB(service *v1.Service, lbName string) *retry.Error {
	ctx, cancel := getContextWithCancel()
	defer cancel()

	rgName := az.getLoadBalancerResourceGroup()
	rerr := az.LoadBalancerClient.Delete(ctx, rgName, lbName)
	if rerr == nil {
		// Invalidate the cache right after updating
		_ = az.lbCache.Delete(lbName)
		return nil
	}

	klog.Errorf("LoadBalancerClient.Delete(%s) failed: %s", lbName, rerr.Error().Error())
	az.Event(service, v1.EventTypeWarning, "DeleteLoadBalancer", rerr.Error().Error())
	return rerr
}

// CreateOrUpdateRouteTable invokes az.RouteTablesClient.CreateOrUpdate with exponential backoff retry
func (az *Cloud) CreateOrUpdateRouteTable(routeTable network.RouteTable) error {
	ctx, cancel := getContextWithCancel()
	defer cancel()

	rerr := az.RouteTablesClient.CreateOrUpdate(ctx, az.RouteTableResourceGroup, az.RouteTableName, routeTable, to.String(routeTable.Etag))
	if rerr == nil {
		// Invalidate the cache right after updating
		_ = az.rtCache.Delete(*routeTable.Name)
		return nil
	}

	rtJSON, _ := json.Marshal(routeTable)
	klog.Warningf("RouteTablesClient.CreateOrUpdate(%s) failed: %v, RouteTable request: %s", to.String(routeTable.Name), rerr.Error(), string(rtJSON))

	// Invalidate the cache because etag mismatch.
	if rerr.HTTPStatusCode == http.StatusPreconditionFailed {
		klog.V(3).Infof("Route table cache for %s is cleanup because of http.StatusPreconditionFailed", *routeTable.Name)
		_ = az.rtCache.Delete(*routeTable.Name)
	}
	// Invalidate the cache because another new operation has canceled the current request.
	if strings.Contains(strings.ToLower(rerr.Error().Error()), consts.OperationCanceledErrorMessage) {
		klog.V(3).Infof("Route table cache for %s is cleanup because CreateOrUpdateRouteTable is canceled by another operation", *routeTable.Name)
		_ = az.rtCache.Delete(*routeTable.Name)
	}
	klog.Errorf("RouteTablesClient.CreateOrUpdate(%s) failed: %v", az.RouteTableName, rerr.Error())
	return rerr.Error()
}

// CreateOrUpdateRoute invokes az.RoutesClient.CreateOrUpdate with exponential backoff retry
func (az *Cloud) CreateOrUpdateRoute(route network.Route) error {
	ctx, cancel := getContextWithCancel()
	defer cancel()

	rerr := az.RoutesClient.CreateOrUpdate(ctx, az.RouteTableResourceGroup, az.RouteTableName, *route.Name, route, to.String(route.Etag))
	klog.V(10).Infof("RoutesClient.CreateOrUpdate(%s): end", *route.Name)
	if rerr == nil {
		_ = az.rtCache.Delete(az.RouteTableName)
		return nil
	}

	if rerr.HTTPStatusCode == http.StatusPreconditionFailed {
		klog.V(3).Infof("Route cache for %s is cleanup because of http.StatusPreconditionFailed", *route.Name)
		_ = az.rtCache.Delete(az.RouteTableName)
	}
	// Invalidate the cache because another new operation has canceled the current request.
	if strings.Contains(strings.ToLower(rerr.Error().Error()), consts.OperationCanceledErrorMessage) {
		klog.V(3).Infof("Route cache for %s is cleanup because CreateOrUpdateRouteTable is canceled by another operation", *route.Name)
		_ = az.rtCache.Delete(az.RouteTableName)
	}
	return rerr.Error()
}

// DeleteRouteWithName invokes az.RoutesClient.CreateOrUpdate with exponential backoff retry
func (az *Cloud) DeleteRouteWithName(routeName string) error {
	ctx, cancel := getContextWithCancel()
	defer cancel()

	rerr := az.RoutesClient.Delete(ctx, az.RouteTableResourceGroup, az.RouteTableName, routeName)
	klog.V(10).Infof("RoutesClient.Delete(%s,%s): end", az.RouteTableName, routeName)
	if rerr == nil {
		return nil
	}

	klog.Errorf("RoutesClient.Delete(%s, %s) failed: %v", az.RouteTableName, routeName, rerr.Error())
	return rerr.Error()
}

// CreateOrUpdateVMSS invokes az.VirtualMachineScaleSetsClient.Update().
func (az *Cloud) CreateOrUpdateVMSS(resourceGroupName string, VMScaleSetName string, parameters compute.VirtualMachineScaleSet) *retry.Error {
	ctx, cancel := getContextWithCancel()
	defer cancel()

	// When vmss is being deleted, CreateOrUpdate API would report "the vmss is being deleted" error.
	// Since it is being deleted, we shouldn't send more CreateOrUpdate requests for it.
	klog.V(3).Infof("CreateOrUpdateVMSS: verify the status of the vmss being created or updated")
	vmss, rerr := az.VirtualMachineScaleSetsClient.Get(ctx, resourceGroupName, VMScaleSetName)
	if rerr != nil {
		klog.Errorf("CreateOrUpdateVMSS: error getting vmss(%s): %v", VMScaleSetName, rerr)
		return rerr
	}
	if vmss.ProvisioningState != nil && strings.EqualFold(*vmss.ProvisioningState, consts.VirtualMachineScaleSetsDeallocating) {
		klog.V(3).Infof("CreateOrUpdateVMSS: found vmss %s being deleted, skipping", VMScaleSetName)
		return nil
	}

	rerr = az.VirtualMachineScaleSetsClient.CreateOrUpdate(ctx, resourceGroupName, VMScaleSetName, parameters)
	klog.V(10).Infof("UpdateVmssVMWithRetry: VirtualMachineScaleSetsClient.CreateOrUpdate(%s): end", VMScaleSetName)
	if rerr != nil {
		klog.Errorf("CreateOrUpdateVMSS: error CreateOrUpdate vmss(%s): %v", VMScaleSetName, rerr)
		return rerr
	}

	return nil
}

func (az *Cloud) CreateOrUpdatePLS(service *v1.Service, pls network.PrivateLinkService) error {
	ctx, cancel := getContextWithCancel()
	defer cancel()

	rerr := az.PrivateLinkServiceClient.CreateOrUpdate(ctx, az.PrivateLinkServiceResourceGroup, to.String(pls.Name), pls, to.String(pls.Etag))
	if rerr == nil {
		// Invalidate the cache right after updating
		_ = az.plsCache.Delete(to.String((*pls.LoadBalancerFrontendIPConfigurations)[0].ID))
		return nil
	}

	rtJSON, _ := json.Marshal(pls)
	klog.Warningf("PrivateLinkServiceClient.CreateOrUpdate(%s) failed: %v, PrivateLinkService request: %s", to.String(pls.Name), rerr.Error(), string(rtJSON))

	// Invalidate the cache because etag mismatch.
	if rerr.HTTPStatusCode == http.StatusPreconditionFailed {
		klog.V(3).Infof("Private link service cache for %s is cleanup because of http.StatusPreconditionFailed", to.String(pls.Name))
		_ = az.plsCache.Delete(to.String((*pls.LoadBalancerFrontendIPConfigurations)[0].ID))
	}
	// Invalidate the cache because another new operation has canceled the current request.
	if strings.Contains(strings.ToLower(rerr.Error().Error()), consts.OperationCanceledErrorMessage) {
		klog.V(3).Infof("Private link service for %s is cleanup because CreateOrUpdatePrivateLinkService is canceled by another operation", to.String(pls.Name))
		_ = az.plsCache.Delete(to.String((*pls.LoadBalancerFrontendIPConfigurations)[0].ID))
	}
	klog.Errorf("PrivateLinkServiceClient.CreateOrUpdate(%s) failed: %v", to.String(pls.Name), rerr.Error())
	return rerr.Error()
}

// DeletePLS invokes az.PrivateLinkServiceClient.Delete with exponential backoff retry
func (az *Cloud) DeletePLS(service *v1.Service, plsName string, plsLBFrontendID string) *retry.Error {
	ctx, cancel := getContextWithCancel()
	defer cancel()

	rerr := az.PrivateLinkServiceClient.Delete(ctx, az.PrivateLinkServiceResourceGroup, plsName)
	if rerr == nil {
		// Invalidate the cache right after deleting
		_ = az.plsCache.Delete(plsLBFrontendID)
		return nil
	}

	klog.Errorf("PrivateLinkServiceClient.DeletePLS(%s) failed: %s", plsName, rerr.Error().Error())
	az.Event(service, v1.EventTypeWarning, "DeletePrivateLinkService", rerr.Error().Error())
	return rerr
}

// DeletePEConn invokes az.PrivateLinkServiceClient.DeletePEConnection with exponential backoff retry
func (az *Cloud) DeletePEConn(service *v1.Service, plsName string, peConnName string) *retry.Error {
	ctx, cancel := getContextWithCancel()
	defer cancel()

	rerr := az.PrivateLinkServiceClient.DeletePEConnection(ctx, az.PrivateLinkServiceResourceGroup, plsName, peConnName)
	if rerr == nil {
		return nil
	}

	klog.Errorf("PrivateLinkServiceClient.DeletePEConnection(%s-%s) failed: %s", plsName, peConnName, rerr.Error().Error())
	az.Event(service, v1.EventTypeWarning, "DeletePrivateEndpointConnection", rerr.Error().Error())
	return rerr
}

// CreateOrUpdateSubnet invokes az.SubnetClient.CreateOrUpdate with exponential backoff retry
func (az *Cloud) CreateOrUpdateSubnet(service *v1.Service, subnet network.Subnet) error {
	ctx, cancel := getContextWithCancel()
	defer cancel()

	var rg string
	if len(az.VnetResourceGroup) > 0 {
		rg = az.VnetResourceGroup
	} else {
		rg = az.ResourceGroup
	}

	rerr := az.SubnetsClient.CreateOrUpdate(ctx, rg, az.VnetName, *subnet.Name, subnet)
	klog.V(10).Infof("SubnetClient.CreateOrUpdate(%s): end", *subnet.Name)
	if rerr != nil {
		klog.Errorf("SubnetClient.CreateOrUpdate(%s) failed: %s", *subnet.Name, rerr.Error().Error())
		az.Event(service, v1.EventTypeWarning, "CreateOrUpdateSubnet", rerr.Error().Error())
		return rerr.Error()
	}

	return nil
}
