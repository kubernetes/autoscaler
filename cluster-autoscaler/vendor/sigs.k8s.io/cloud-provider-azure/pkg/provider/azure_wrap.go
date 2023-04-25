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
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2022-03-01/compute"
	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2021-08-01/network"
	"github.com/Azure/go-autorest/autorest/to"

	"k8s.io/apimachinery/pkg/types"
	cloudprovider "k8s.io/cloud-provider"
	"k8s.io/klog/v2"

	azcache "sigs.k8s.io/cloud-provider-azure/pkg/cache"
	"sigs.k8s.io/cloud-provider-azure/pkg/consts"
	"sigs.k8s.io/cloud-provider-azure/pkg/retry"
)

var (
	vmCacheTTLDefaultInSeconds           = 60
	loadBalancerCacheTTLDefaultInSeconds = 120
	nsgCacheTTLDefaultInSeconds          = 120
	routeTableCacheTTLDefaultInSeconds   = 120
	publicIPCacheTTLDefaultInSeconds     = 120
	plsCacheTTLDefaultInSeconds          = 120

	azureNodeProviderIDRE    = regexp.MustCompile(`^azure:///subscriptions/(?:.*)/resourceGroups/(?:.*)/providers/Microsoft.Compute/(?:.*)`)
	azureResourceGroupNameRE = regexp.MustCompile(`.*/subscriptions/(?:.*)/resourceGroups/(.+)/providers/(?:.*)`)
)

// checkExistsFromError inspects an error and returns a true if err is nil,
// false if error is an autorest.Error with StatusCode=404 and will return the
// error back if error is another status code or another type of error.
func checkResourceExistsFromError(err *retry.Error) (bool, *retry.Error) {
	if err == nil {
		return true, nil
	}

	if err.HTTPStatusCode == http.StatusNotFound {
		return false, nil
	}

	return false, err
}

// getVirtualMachine calls 'VirtualMachinesClient.Get' with a timed cache
// The service side has throttling control that delays responses if there are multiple requests onto certain vm
// resource request in short period.
func (az *Cloud) getVirtualMachine(nodeName types.NodeName, crt azcache.AzureCacheReadType) (vm compute.VirtualMachine, err error) {
	vmName := string(nodeName)
	cachedVM, err := az.vmCache.Get(vmName, crt)
	if err != nil {
		return vm, err
	}

	if cachedVM == nil {
		klog.Warningf("Unable to find node %s: %v", nodeName, cloudprovider.InstanceNotFound)
		return vm, cloudprovider.InstanceNotFound
	}

	return *(cachedVM.(*compute.VirtualMachine)), nil
}

func (az *Cloud) getRouteTable(crt azcache.AzureCacheReadType) (routeTable network.RouteTable, exists bool, err error) {
	if len(az.RouteTableName) == 0 {
		return routeTable, false, fmt.Errorf("Route table name is not configured")
	}

	cachedRt, err := az.rtCache.GetWithDeepCopy(az.RouteTableName, crt)
	if err != nil {
		return routeTable, false, err
	}

	if cachedRt == nil {
		return routeTable, false, nil
	}

	return *(cachedRt.(*network.RouteTable)), true, nil
}

func (az *Cloud) getPIPCacheKey(pipResourceGroup string, pipName string) string {
	resourceGroup := az.ResourceGroup
	if pipResourceGroup != "" {
		resourceGroup = pipResourceGroup
	}
	return fmt.Sprintf("%s%s%s", resourceGroup, consts.PIPCacheKeySeparator, pipName)
}

func (az *Cloud) getPublicIPAddress(pipResourceGroup string, pipName string, crt azcache.AzureCacheReadType) (network.PublicIPAddress, bool, error) {
	pip := network.PublicIPAddress{}
	cacheKey := az.getPIPCacheKey(pipResourceGroup, pipName)
	cachedPIP, err := az.pipCache.GetWithDeepCopy(cacheKey, crt)
	if err != nil {
		return pip, false, err
	}

	if cachedPIP == nil {
		return pip, false, nil
	}

	return *(cachedPIP.(*network.PublicIPAddress)), true, nil
}

func (az *Cloud) getSubnet(virtualNetworkName string, subnetName string) (network.Subnet, bool, error) {
	var rg string
	if len(az.VnetResourceGroup) > 0 {
		rg = az.VnetResourceGroup
	} else {
		rg = az.ResourceGroup
	}

	ctx, cancel := getContextWithCancel()
	defer cancel()
	subnet, err := az.SubnetsClient.Get(ctx, rg, virtualNetworkName, subnetName, "")
	exists, rerr := checkResourceExistsFromError(err)
	if rerr != nil {
		return subnet, false, rerr.Error()
	}

	if !exists {
		klog.V(2).Infof("Subnet %q not found", subnetName)
		return subnet, false, nil
	}

	return subnet, exists, nil
}

func (az *Cloud) getAzureLoadBalancer(name string, crt azcache.AzureCacheReadType) (lb *network.LoadBalancer, exists bool, err error) {
	cachedLB, err := az.lbCache.GetWithDeepCopy(name, crt)
	if err != nil {
		return lb, false, err
	}

	if cachedLB == nil {
		return lb, false, nil
	}

	return cachedLB.(*network.LoadBalancer), true, nil
}

func (az *Cloud) getSecurityGroup(crt azcache.AzureCacheReadType) (network.SecurityGroup, error) {
	nsg := network.SecurityGroup{}
	if az.SecurityGroupName == "" {
		return nsg, fmt.Errorf("securityGroupName is not configured")
	}

	securityGroup, err := az.nsgCache.GetWithDeepCopy(az.SecurityGroupName, crt)
	if err != nil {
		return nsg, err
	}

	if securityGroup == nil {
		return nsg, fmt.Errorf("nsg %q not found", az.SecurityGroupName)
	}

	return *(securityGroup.(*network.SecurityGroup)), nil
}

func (az *Cloud) getPrivateLinkService(frontendIPConfigID *string, crt azcache.AzureCacheReadType) (pls network.PrivateLinkService, err error) {
	cachedPLS, err := az.plsCache.GetWithDeepCopy(*frontendIPConfigID, crt)
	if err != nil {
		return pls, err
	}
	return *(cachedPLS.(*network.PrivateLinkService)), nil
}

func (az *Cloud) newVMCache() (*azcache.TimedCache, error) {
	getter := func(key string) (interface{}, error) {
		// Currently InstanceView request are used by azure_zones, while the calls come after non-InstanceView
		// request. If we first send an InstanceView request and then a non InstanceView request, the second
		// request will still hit throttling. This is what happens now for cloud controller manager: In this
		// case we do get instance view every time to fulfill the azure_zones requirement without hitting
		// throttling.
		// Consider adding separate parameter for controlling 'InstanceView' once node update issue #56276 is fixed
		ctx, cancel := getContextWithCancel()
		defer cancel()

		resourceGroup, err := az.GetNodeResourceGroup(key)
		if err != nil {
			return nil, err
		}

		vm, verr := az.VirtualMachinesClient.Get(ctx, resourceGroup, key, compute.InstanceViewTypesInstanceView)
		exists, rerr := checkResourceExistsFromError(verr)
		if rerr != nil {
			return nil, rerr.Error()
		}

		if !exists {
			klog.V(2).Infof("Virtual machine %q not found", key)
			return nil, nil
		}

		if vm.VirtualMachineProperties != nil &&
			strings.EqualFold(to.String(vm.VirtualMachineProperties.ProvisioningState), string(compute.ProvisioningStateDeleting)) {
			klog.V(2).Infof("Virtual machine %q is under deleting", key)
			return nil, nil
		}

		return &vm, nil
	}

	if az.VMCacheTTLInSeconds == 0 {
		az.VMCacheTTLInSeconds = vmCacheTTLDefaultInSeconds
	}
	return azcache.NewTimedcache(time.Duration(az.VMCacheTTLInSeconds)*time.Second, getter)
}

func (az *Cloud) newLBCache() (*azcache.TimedCache, error) {
	getter := func(key string) (interface{}, error) {
		ctx, cancel := getContextWithCancel()
		defer cancel()

		lb, err := az.LoadBalancerClient.Get(ctx, az.getLoadBalancerResourceGroup(), key, "")
		exists, rerr := checkResourceExistsFromError(err)
		if rerr != nil {
			return nil, rerr.Error()
		}

		if !exists {
			klog.V(2).Infof("Load balancer %q not found", key)
			return nil, nil
		}

		return &lb, nil
	}

	if az.LoadBalancerCacheTTLInSeconds == 0 {
		az.LoadBalancerCacheTTLInSeconds = loadBalancerCacheTTLDefaultInSeconds
	}
	return azcache.NewTimedcache(time.Duration(az.LoadBalancerCacheTTLInSeconds)*time.Second, getter)
}

func (az *Cloud) newNSGCache() (*azcache.TimedCache, error) {
	getter := func(key string) (interface{}, error) {
		ctx, cancel := getContextWithCancel()
		defer cancel()
		nsg, err := az.SecurityGroupsClient.Get(ctx, az.SecurityGroupResourceGroup, key, "")
		exists, rerr := checkResourceExistsFromError(err)
		if rerr != nil {
			return nil, rerr.Error()
		}

		if !exists {
			klog.V(2).Infof("Security group %q not found", key)
			return nil, nil
		}

		return &nsg, nil
	}

	if az.NsgCacheTTLInSeconds == 0 {
		az.NsgCacheTTLInSeconds = nsgCacheTTLDefaultInSeconds
	}
	return azcache.NewTimedcache(time.Duration(az.NsgCacheTTLInSeconds)*time.Second, getter)
}

func (az *Cloud) newRouteTableCache() (*azcache.TimedCache, error) {
	getter := func(key string) (interface{}, error) {
		ctx, cancel := getContextWithCancel()
		defer cancel()
		rt, err := az.RouteTablesClient.Get(ctx, az.RouteTableResourceGroup, key, "")
		exists, rerr := checkResourceExistsFromError(err)
		if rerr != nil {
			return nil, rerr.Error()
		}

		if !exists {
			klog.V(2).Infof("Route table %q not found", key)
			return nil, nil
		}

		return &rt, nil
	}

	if az.RouteTableCacheTTLInSeconds == 0 {
		az.RouteTableCacheTTLInSeconds = routeTableCacheTTLDefaultInSeconds
	}
	return azcache.NewTimedcache(time.Duration(az.RouteTableCacheTTLInSeconds)*time.Second, getter)
}

func (az *Cloud) newPIPCache() (*azcache.TimedCache, error) {
	getter := func(key string) (interface{}, error) {
		ctx, cancel := getContextWithCancel()
		defer cancel()

		parsedKey := strings.Split(strings.TrimSpace(key), consts.PIPCacheKeySeparator)
		if len(parsedKey) != 2 {
			return nil, fmt.Errorf("failed to parse public ip rg and name from cache key %q", key)
		}
		pipResourceGroup, pipName := strings.TrimSpace(parsedKey[0]), strings.TrimSpace(parsedKey[1])

		pip, err := az.PublicIPAddressesClient.Get(ctx, pipResourceGroup, pipName, "")
		exists, rerr := checkResourceExistsFromError(err)
		if rerr != nil {
			return nil, rerr.Error()
		}

		if !exists {
			klog.V(2).Infof("Public IP %q in rg %q not found", pipName, pipResourceGroup)
			return nil, nil
		}

		return &pip, nil
	}

	if az.PublicIPCacheTTLInSeconds == 0 {
		az.PublicIPCacheTTLInSeconds = publicIPCacheTTLDefaultInSeconds
	}
	return azcache.NewTimedcache(time.Duration(az.PublicIPCacheTTLInSeconds)*time.Second, getter)
}

func (az *Cloud) newPLSCache() (*azcache.TimedCache, error) {
	// for PLS cache, key is LBFrontendIPConfiguration ID
	getter := func(key string) (interface{}, error) {
		ctx, cancel := getContextWithCancel()
		defer cancel()
		plsList, err := az.PrivateLinkServiceClient.List(ctx, az.PrivateLinkServiceResourceGroup)
		exists, rerr := checkResourceExistsFromError(err)
		if rerr != nil {
			return nil, rerr.Error()
		}

		if exists {
			for i := range plsList {
				pls := plsList[i]
				if pls.PrivateLinkServiceProperties == nil {
					continue
				}
				fipConfigs := pls.PrivateLinkServiceProperties.LoadBalancerFrontendIPConfigurations
				if fipConfigs == nil {
					continue
				}
				for _, fipConfig := range *fipConfigs {
					if strings.EqualFold(*fipConfig.ID, key) {
						return &pls, nil
					}
				}

			}
		}

		klog.V(2).Infof("No privateLinkService found for frontendIPConfig %q", key)
		plsNotExistID := consts.PrivateLinkServiceNotExistID
		return &network.PrivateLinkService{ID: &plsNotExistID}, nil
	}

	if az.PlsCacheTTLInSeconds == 0 {
		az.PlsCacheTTLInSeconds = plsCacheTTLDefaultInSeconds
	}
	return azcache.NewTimedcache(time.Duration(az.PlsCacheTTLInSeconds)*time.Second, getter)
}

func (az *Cloud) useStandardLoadBalancer() bool {
	return strings.EqualFold(az.LoadBalancerSku, consts.LoadBalancerSkuStandard)
}

func (az *Cloud) excludeMasterNodesFromStandardLB() bool {
	return az.ExcludeMasterFromStandardLB != nil && *az.ExcludeMasterFromStandardLB
}

func (az *Cloud) disableLoadBalancerOutboundSNAT() bool {
	if !az.useStandardLoadBalancer() || az.DisableOutboundSNAT == nil {
		return false
	}

	return *az.DisableOutboundSNAT
}

// IsNodeUnmanaged returns true if the node is not managed by Azure cloud provider.
// Those nodes includes on-prem or VMs from other clouds. They will not be added to load balancer
// backends. Azure routes and managed disks are also not supported for them.
func (az *Cloud) IsNodeUnmanaged(nodeName string) (bool, error) {
	unmanagedNodes, err := az.GetUnmanagedNodes()
	if err != nil {
		return false, err
	}

	return unmanagedNodes.Has(nodeName), nil
}

// IsNodeUnmanagedByProviderID returns true if the node is not managed by Azure cloud provider.
// All managed node's providerIDs are in format 'azure:///subscriptions/<id>/resourceGroups/<rg>/providers/Microsoft.Compute/.*'
func (az *Cloud) IsNodeUnmanagedByProviderID(providerID string) bool {
	return !azureNodeProviderIDRE.Match([]byte(providerID))
}

// ConvertResourceGroupNameToLower converts the resource group name in the resource ID to be lowered.
func ConvertResourceGroupNameToLower(resourceID string) (string, error) {
	matches := azureResourceGroupNameRE.FindStringSubmatch(resourceID)
	if len(matches) != 2 {
		return "", fmt.Errorf("%q isn't in Azure resource ID format %q", resourceID, azureResourceGroupNameRE.String())
	}

	resourceGroup := matches[1]
	return strings.Replace(resourceID, resourceGroup, strings.ToLower(resourceGroup), 1), nil
}

// isBackendPoolOnSameLB checks whether newBackendPoolID is on the same load balancer as existingBackendPools.
// Since both public and internal LBs are supported, lbName and lbName-internal are treated as same.
// If not same, the lbName for existingBackendPools would also be returned.
func isBackendPoolOnSameLB(newBackendPoolID string, existingBackendPools []string) (bool, string, error) {
	matches := backendPoolIDRE.FindStringSubmatch(newBackendPoolID)
	if len(matches) != 2 {
		return false, "", fmt.Errorf("new backendPoolID %q is in wrong format", newBackendPoolID)
	}

	newLBName := matches[1]
	newLBNameTrimmed := strings.TrimSuffix(newLBName, consts.InternalLoadBalancerNameSuffix)
	for _, backendPool := range existingBackendPools {
		matches := backendPoolIDRE.FindStringSubmatch(backendPool)
		if len(matches) != 2 {
			return false, "", fmt.Errorf("existing backendPoolID %q is in wrong format", backendPool)
		}

		lbName := matches[1]
		if !strings.EqualFold(strings.TrimSuffix(lbName, consts.InternalLoadBalancerNameSuffix), newLBNameTrimmed) {
			return false, lbName, nil
		}
	}

	return true, "", nil
}
