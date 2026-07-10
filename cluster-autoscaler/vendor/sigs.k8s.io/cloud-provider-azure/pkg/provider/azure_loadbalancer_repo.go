/*
Copyright 2023 The Kubernetes Authors.

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
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v9"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"

	azcache "sigs.k8s.io/cloud-provider-azure/pkg/cache"
	"sigs.k8s.io/cloud-provider-azure/pkg/consts"
	"sigs.k8s.io/cloud-provider-azure/pkg/log"
	"sigs.k8s.io/cloud-provider-azure/pkg/util/errutils"
	utilsets "sigs.k8s.io/cloud-provider-azure/pkg/util/sets"
)

// DeleteLB invokes az.NetworkClientFactory.GetLoadBalancerClient().Delete with exponential backoff retry
func (az *Cloud) DeleteLB(ctx context.Context, service *v1.Service, lbName string) error {
	logger := log.FromContextOrBackground(ctx).WithName("DeleteLB")
	rgName := az.getLoadBalancerResourceGroup()
	rerr := az.NetworkClientFactory.GetLoadBalancerClient().Delete(ctx, rgName, lbName)
	if rerr == nil {
		// Invalidate the cache right after updating
		_ = az.lbCache.Delete(lbName)
		return nil
	}

	logger.Error(rerr, "LoadbalancerClient.Delete failed", "lbName", lbName)
	az.Event(service, v1.EventTypeWarning, "DeleteLoadBalancer", rerr.Error())
	return rerr
}

// ListLB invokes az.NetworkClientFactory.GetLoadBalancerClient().List with exponential backoff retry
func (az *Cloud) ListLB(ctx context.Context, service *v1.Service) ([]*armnetwork.LoadBalancer, error) {
	logger := log.FromContextOrBackground(ctx).WithName("ListLB")
	rgName := az.getLoadBalancerResourceGroup()
	allLBs, rerr := az.NetworkClientFactory.GetLoadBalancerClient().List(ctx, rgName)
	if rerr != nil {
		if exist, err := errutils.CheckResourceExistsFromAzcoreError(rerr); !exist && err == nil {
			return nil, nil
		}
		az.Event(service, v1.EventTypeWarning, "ListLoadBalancers", rerr.Error())
		logger.Error(rerr, "LoadbalancerClient.List failure", "resourceGroup", rgName)
		return nil, rerr
	}
	logger.V(2).Info("LoadbalancerClient.List success", "resourceGroup", rgName)
	return allLBs, nil
}

// ListManagedLBs invokes az.NetworkClientFactory.GetLoadBalancerClient().List and filter out
// those that are not managed by cloud provider azure or not associated to a managed VMSet.
func (az *Cloud) ListManagedLBs(ctx context.Context, service *v1.Service, nodes []*v1.Node, clusterName string) ([]*armnetwork.LoadBalancer, error) {
	logger := log.FromContextOrBackground(ctx).WithName("ListManagedLBs")
	allLBs, err := az.ListLB(ctx, service)
	if err != nil {
		return nil, err
	}

	if allLBs == nil {
		klog.Warningf("ListManagedLBs: no LBs found")
		return nil, nil
	}

	managedLBNames := utilsets.NewString(clusterName)
	managedLBs := make([]*armnetwork.LoadBalancer, 0)
	if strings.EqualFold(az.LoadBalancerSKU, consts.LoadBalancerSKUBasic) {
		// return early if wantLb=false
		if nodes == nil {
			logger.V(4).Info("return all LBs in the resource group, including unmanaged LBs", "resourceGroup", az.getLoadBalancerResourceGroup())
			return allLBs, nil
		}

		agentPoolVMSetNamesMap := make(map[string]bool)
		agentPoolVMSetNames, err := az.VMSet.GetAgentPoolVMSetNames(ctx, nodes)
		if err != nil {
			return nil, fmt.Errorf("ListManagedLBs: failed to get agent pool vmSet names: %w", err)
		}

		if len(agentPoolVMSetNames) > 0 {
			for _, vmSetName := range agentPoolVMSetNames {
				logger.V(6).Info("found agent pool", "vmSet Name", *vmSetName)
				agentPoolVMSetNamesMap[strings.ToLower(*vmSetName)] = true
			}
		}

		for agentPoolVMSetName := range agentPoolVMSetNamesMap {
			managedLBNames.Insert(az.mapVMSetNameToLoadBalancerName(agentPoolVMSetName, clusterName))
		}
	}

	if az.UseMultipleStandardLoadBalancers() {
		for _, multiSLBConfig := range az.MultipleStandardLoadBalancerConfigurations {
			managedLBNames.Insert(multiSLBConfig.Name, fmt.Sprintf("%s%s", multiSLBConfig.Name, consts.InternalLoadBalancerNameSuffix))
		}
	}

	for _, lb := range allLBs {
		if managedLBNames.Has(trimSuffixIgnoreCase(ptr.Deref(lb.Name, ""), consts.InternalLoadBalancerNameSuffix)) {
			managedLBs = append(managedLBs, lb)
			logger.V(4).Info("found managed LB", "loadBalancerName", ptr.Deref(lb.Name, ""))
		}
	}

	return managedLBs, nil
}

// CreateOrUpdateLB invokes az.NetworkClientFactory.GetLoadBalancerClient().CreateOrUpdate with exponential backoff retry
func (az *Cloud) CreateOrUpdateLB(ctx context.Context, service *v1.Service, lb armnetwork.LoadBalancer) error {
	logger := log.FromContextOrBackground(ctx).WithName("CreateOrUpdateLB")
	lb = cleanupSubnetInFrontendIPConfigurations(&lb)

	rgName := az.getLoadBalancerResourceGroup()
	_, err := az.NetworkClientFactory.GetLoadBalancerClient().CreateOrUpdate(ctx, rgName, ptr.Deref(lb.Name, ""), lb)
	logger.V(10).Info("LoadbalancerClient.CreateOrUpdate: end", "loadBalancerName", *lb.Name)
	if err == nil {
		// Invalidate the cache right after updating
		_ = az.lbCache.Delete(*lb.Name)
		return nil
	}

	lbJSON, _ := json.Marshal(lb)
	klog.Warningf("LoadbalancerClient.CreateOrUpdate(%s) failed: %v, LoadBalancer request: %s", ptr.Deref(lb.Name, ""), err, string(lbJSON))
	var rerr *azcore.ResponseError
	if !errors.As(err, &rerr) {
		return err
	}
	// Invalidate the cache because ETAG precondition mismatch.
	if rerr.StatusCode == http.StatusPreconditionFailed {
		logger.V(3).Info("LoadBalancer cache is cleanup because of http.StatusPreconditionFailed", "loadBalancerName", ptr.Deref(lb.Name, ""))
		_ = az.lbCache.Delete(*lb.Name)
	}

	retryErrorMessage := rerr.Error()
	// Invalidate the cache because another new operation has canceled the current request.
	if strings.Contains(strings.ToLower(retryErrorMessage), consts.OperationCanceledErrorMessage) {
		logger.V(3).Info("LoadBalancer cache is cleanup because CreateOrUpdate is canceled by another operation", "loadBalancerName", ptr.Deref(lb.Name, ""))
		_ = az.lbCache.Delete(*lb.Name)
	}

	// The LB update may fail because the referenced PIP is not in the Succeeded provisioning state
	if strings.Contains(strings.ToLower(retryErrorMessage), strings.ToLower(consts.ReferencedResourceNotProvisionedMessageCode)) {
		matches := pipErrorMessageRE.FindStringSubmatch(retryErrorMessage)
		if len(matches) != 3 {
			logger.Error(nil, "Failed to parse the retry error message", "retryErrorMessage", retryErrorMessage)
			return rerr
		}
		pipRG, pipName := matches[1], matches[2]
		logger.V(3).Info("The public IP referenced by load balancer is not in Succeeded provisioning state, will try to update it", "pipName", pipName, "loadBalancerName", ptr.Deref(lb.Name, ""))
		pip, _, err := az.getPublicIPAddress(ctx, pipRG, pipName, azcache.CacheReadTypeDefault)
		if err != nil {
			logger.Error(err, "Failed to get the public IP", "pip", pipName, "resourceGroup", pipRG)
			return rerr
		}
		// Perform a dummy update to fix the provisioning state
		err = az.CreateOrUpdatePIP(service, pipRG, pip)
		if err != nil {
			logger.Error(err, "Failed to update the public IP", "pip", pipName, "resourceGroup", pipRG)
			return rerr
		}
		// Invalidate the LB cache, return the error, and the controller manager
		// would retry the LB update in the next reconcile loop
		_ = az.lbCache.Delete(*lb.Name)
	}

	return rerr
}

func (az *Cloud) CreateOrUpdateLBBackendPool(ctx context.Context, lbName string, backendPool *armnetwork.BackendAddressPool) error {
	logger := log.FromContextOrBackground(ctx).WithName("CreateOrUpdateLBBackendPool")
	logger.V(4).Info("updating backend pool in LB", "backendPoolName", ptr.Deref(backendPool.Name, ""), "loadBalancerName", lbName)
	_, err := az.NetworkClientFactory.GetBackendAddressPoolClient().CreateOrUpdate(ctx, az.getLoadBalancerResourceGroup(), lbName, ptr.Deref(backendPool.Name, ""), *backendPool)
	if err == nil {
		// Invalidate the cache right after updating
		_ = az.lbCache.Delete(lbName)
		return nil
	}
	var rerr *azcore.ResponseError
	if !errors.As(err, &rerr) {
		return err
	}

	// Invalidate the cache because ETAG precondition mismatch.
	if rerr.StatusCode == http.StatusPreconditionFailed {
		logger.V(3).Info("LoadBalancer cache is cleanup because of http.StatusPreconditionFailed", "loadBalancerName", lbName)
		_ = az.lbCache.Delete(lbName)
	}

	retryErrorMessage := rerr.Error()
	// Invalidate the cache because another new operation has canceled the current request.
	if strings.Contains(strings.ToLower(retryErrorMessage), consts.OperationCanceledErrorMessage) {
		logger.V(3).Info("LoadBalancer cache is cleanup because CreateOrUpdate is canceled by another operation", "loadBalancerName", lbName)
		_ = az.lbCache.Delete(lbName)
	}

	return rerr
}

func (az *Cloud) DeleteLBBackendPool(ctx context.Context, lbName, backendPoolName string) error {
	logger := log.FromContextOrBackground(ctx).WithName("DeleteLBBackendPool")
	logger.V(4).Info("deleting backend pool in LB", "backendPoolName", backendPoolName, "loadBalancerName", lbName)
	err := az.NetworkClientFactory.GetBackendAddressPoolClient().Delete(ctx, az.getLoadBalancerResourceGroup(), lbName, backendPoolName)
	if err == nil {
		// Invalidate the cache right after updating
		_ = az.lbCache.Delete(lbName)
		return nil
	}

	var rerr *azcore.ResponseError
	if !errors.As(err, &rerr) {
		return err
	}
	// Invalidate the cache because ETAG precondition mismatch.
	if rerr.StatusCode == http.StatusPreconditionFailed {
		logger.V(3).Info("LoadBalancer cache is cleanup because of http.StatusPreconditionFailed", "loadBalancerName", lbName)
		_ = az.lbCache.Delete(lbName)
	}

	retryErrorMessage := rerr.Error()
	// Invalidate the cache because another new operation has canceled the current request.
	if strings.Contains(strings.ToLower(retryErrorMessage), consts.OperationCanceledErrorMessage) {
		logger.V(3).Info("LoadBalancer cache is cleanup because CreateOrUpdate is canceled by another operation", "loadBalancerName", lbName)
		_ = az.lbCache.Delete(lbName)
	}

	return rerr
}

func cleanupSubnetInFrontendIPConfigurations(lb *armnetwork.LoadBalancer) armnetwork.LoadBalancer {
	if lb.Properties == nil || lb.Properties.FrontendIPConfigurations == nil {
		return *lb
	}

	frontendIPConfigurations := lb.Properties.FrontendIPConfigurations
	for i := range frontendIPConfigurations {
		config := frontendIPConfigurations[i]
		if config.Properties != nil &&
			config.Properties.Subnet != nil &&
			config.Properties.Subnet.ID != nil {
			subnet := armnetwork.Subnet{
				ID: config.Properties.Subnet.ID,
			}
			if config.Properties.Subnet.Name != nil {
				subnet.Name = config.Properties.Subnet.Name
			}
			config.Properties.Subnet = &subnet
			frontendIPConfigurations[i] = config
			continue
		}
	}

	lb.Properties.FrontendIPConfigurations = frontendIPConfigurations
	return *lb
}

// MigrateToIPBasedBackendPoolAndWaitForCompletion use the migration API to migrate from
// NIC-based to IP-based LB backend pools. It also makes sure the number of IP addresses
// in the backend pools is expected.
func (az *Cloud) MigrateToIPBasedBackendPoolAndWaitForCompletion(
	ctx context.Context,
	lbName string, backendPoolNames []string, nicsCountMap map[string]int,
) error {
	logger := log.FromContextOrBackground(ctx).WithName("MigrateToIPBasedBackendPoolAndWaitForCompletion")
	if _, rerr := az.NetworkClientFactory.GetLoadBalancerClient().MigrateToIPBased(ctx, az.ResourceGroup, lbName, &armnetwork.LoadBalancersClientMigrateToIPBasedOptions{
		Parameters: &armnetwork.MigrateLoadBalancerToIPBasedRequest{
			Pools: to.SliceOfPtrs(backendPoolNames...),
		},
	}); rerr != nil {
		backendPoolNamesStr := strings.Join(backendPoolNames, ",")
		logger.Error(rerr, "Failed to migrate to IP based backend pool", "lb", lbName, "backendPool", backendPoolNamesStr)
		return rerr
	}

	succeeded := make(map[string]bool)
	for bpName := range nicsCountMap {
		succeeded[bpName] = false
	}

	err := wait.PollImmediateWithContext(ctx, 5*time.Second, 10*time.Minute, func(ctx context.Context) (done bool, err error) {
		for bpName, nicsCount := range nicsCountMap {
			if succeeded[bpName] {
				continue
			}

			bp, rerr := az.NetworkClientFactory.GetBackendAddressPoolClient().Get(ctx, az.ResourceGroup, lbName, bpName)
			if rerr != nil {
				logger.Error(rerr, "Failed to get backend pool", "backendPool", bpName, "lb", lbName)
				return false, rerr
			}

			if countIPsOnBackendPool(bp) != nicsCount {
				logger.V(4).Info("Expected IPs vs current IPs, will retry in 5s", "expectedIPs", nicsCount, "currentIPs", countIPsOnBackendPool(bp))
				return false, nil
			}
			succeeded[bpName] = true
		}
		return true, nil
	})
	if err != nil {
		if errors.Is(err, wait.ErrWaitTimeout) {
			klog.Warningf("MigrateToIPBasedBackendPoolAndWaitForCompletion: Timeout waiting for migration to IP based backend pool for lb %s, backend pool %s", lbName, strings.Join(backendPoolNames, ","))
			return nil
		}

		logger.Error(err, "Failed to wait for migration to IP based backend pool", "lb", lbName, "backendPool", strings.Join(backendPoolNames, ","))
		return err
	}

	return nil
}

func (az *Cloud) newLBCache() (azcache.Resource, error) {
	getter := func(ctx context.Context, key string) (interface{}, error) {
		logger := log.FromContextOrBackground(ctx).WithName("newLBCache")
		lb, err := az.NetworkClientFactory.GetLoadBalancerClient().Get(ctx, az.getLoadBalancerResourceGroup(), key, nil)
		exists, rerr := checkResourceExistsFromError(err)
		if rerr != nil {
			return nil, rerr
		}

		if !exists {
			logger.V(2).Info("Load balancer not found", "loadBalancerName", key)
			return nil, nil
		}

		return lb, nil
	}

	if az.LoadBalancerCacheTTLInSeconds == 0 {
		az.LoadBalancerCacheTTLInSeconds = loadBalancerCacheTTLDefaultInSeconds
	}
	return azcache.NewTimedCache(time.Duration(az.LoadBalancerCacheTTLInSeconds)*time.Second, getter, az.DisableAPICallCache)
}

func (az *Cloud) getAzureLoadBalancer(ctx context.Context, name string, crt azcache.AzureCacheReadType) (lb *armnetwork.LoadBalancer, exists bool, err error) {
	cachedLB, err := az.lbCache.GetWithDeepCopy(ctx, name, crt)
	if err != nil {
		return lb, false, err
	}

	if cachedLB == nil {
		return lb, false, nil
	}

	return cachedLB.(*armnetwork.LoadBalancer), true, nil
}

// isBackendPoolOnSameLB checks whether newBackendPoolID is on the same load balancer as existingBackendPools.
// Since both public and internal LBs are supported, lbName and lbName-internal are treated as same.
// If not same, the lbName for existingBackendPools would also be returned.
func isBackendPoolOnSameLB(newBackendPoolID string, existingBackendPools []string) (bool, string, error) {
	matches := backendPoolIDRE.FindStringSubmatch(newBackendPoolID)
	if len(matches) != 3 {
		return false, "", fmt.Errorf("new backendPoolID %q is in wrong format", newBackendPoolID)
	}

	newLBName := matches[1]
	newLBNameTrimmed := trimSuffixIgnoreCase(newLBName, consts.InternalLoadBalancerNameSuffix)
	for _, backendPool := range existingBackendPools {
		matches := backendPoolIDRE.FindStringSubmatch(backendPool)
		if len(matches) != 3 {
			return false, "", fmt.Errorf("existing backendPoolID %q is in wrong format", backendPool)
		}

		lbName := matches[1]
		if !strings.EqualFold(trimSuffixIgnoreCase(lbName, consts.InternalLoadBalancerNameSuffix), newLBNameTrimmed) {
			return false, lbName, nil
		}
	}

	return true, "", nil
}

func (az *Cloud) serviceOwnsRule(service *v1.Service, rule string) bool {
	if !strings.EqualFold(string(service.Spec.ExternalTrafficPolicy), string(v1.ServiceExternalTrafficPolicyTypeLocal)) &&
		rule == consts.SharedProbeName {
		return true
	}
	prefix := az.getRulePrefix(service)
	return strings.HasPrefix(strings.ToUpper(rule), strings.ToUpper(prefix))
}

func isNICPool(bp *armnetwork.BackendAddressPool) bool {
	logger := klog.Background().WithName("isNICPool").WithValues("backendPoolName", ptr.Deref(bp.Name, ""))
	if bp.Properties != nil &&
		bp.Properties.LoadBalancerBackendAddresses != nil {
		for _, addr := range bp.Properties.LoadBalancerBackendAddresses {
			if ptr.Deref(addr.Properties.IPAddress, "") == "" {
				logger.V(4).Info("The load balancer backend address has empty ip address, assuming it is a NIC pool",
					"loadBalancerBackendAddress", ptr.Deref(addr.Name, ""))
				return true
			}
		}
	}
	return false
}

// cleanupBasicLoadBalancer removes outdated basic load balancers
// when the loadBalancerSkus is `Standard`. It ensures the backend pool of the basic
// load balancer is empty before deleting the load balancer.
func (az *Cloud) cleanupBasicLoadBalancer(
	ctx context.Context, clusterName string, service *v1.Service, existingLBs []*armnetwork.LoadBalancer,
) ([]*armnetwork.LoadBalancer, error) {
	logger := log.FromContextOrBackground(ctx).WithName("cleanupBasicLoadBalancer")
	if !az.UseStandardLoadBalancer() {
		return existingLBs, nil
	}

	var elbRemoved bool
	for i := len(existingLBs) - 1; i >= 0; i-- {
		lb := existingLBs[i]
		if lb != nil && lb.SKU != nil && lb.SKU.Name != nil && *lb.SKU.Name == armnetwork.LoadBalancerSKUNameBasic {
			logger.V(2).Info("found basic load balancer, removing it", "loadBalancerName", *lb.Name)
			if err := az.safeDeleteLoadBalancer(ctx, *lb, clusterName, service); err != nil {
				logger.Error(err, "failed to delete outdated basic load balancer", "loadBalancerName", *lb.Name)
				return nil, err
			}
			existingLBs = append(existingLBs[:i], existingLBs[i+1:]...)
			if !strings.Contains(strings.ToLower(ptr.Deref(lb.Name, "")), strings.ToLower(consts.InternalLoadBalancerNameSuffix)) {
				elbRemoved = true
			}
		}
	}
	// The lb refs in pip will be changed after the removal, so we need to
	// reinitialize the cache to prevent etag mismatches.
	if elbRemoved {
		var err error
		az.pipCache, err = az.newPIPCache()
		if err != nil {
			logger.Error(err, "failed to refresh pip cache")
			return nil, err
		}
	}
	return existingLBs, nil
}
