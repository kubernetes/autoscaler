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

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2022-07-01/network"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	"k8s.io/utils/pointer"

	azcache "sigs.k8s.io/cloud-provider-azure/pkg/cache"
	"sigs.k8s.io/cloud-provider-azure/pkg/consts"
	"sigs.k8s.io/cloud-provider-azure/pkg/retry"
	utilsets "sigs.k8s.io/cloud-provider-azure/pkg/util/sets"
)

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

// ListManagedLBs invokes az.LoadBalancerClient.List and filter out
// those that are not managed by cloud provider azure or not associated to a managed VMSet.
func (az *Cloud) ListManagedLBs(service *v1.Service, nodes []*v1.Node, clusterName string) (*[]network.LoadBalancer, error) {
	allLBs, err := az.ListLB(service)
	if err != nil {
		return nil, err
	}

	if allLBs == nil {
		klog.Warningf("ListManagedLBs: no LBs found")
		return nil, nil
	}

	managedLBNames := utilsets.NewString(clusterName)
	managedLBs := make([]network.LoadBalancer, 0)
	if strings.EqualFold(az.LoadBalancerSku, consts.LoadBalancerSkuBasic) {
		// return early if wantLb=false
		if nodes == nil {
			klog.V(4).Infof("ListManagedLBs: return all LBs in the resource group %s, including unmanaged LBs", az.getLoadBalancerResourceGroup())
			return &allLBs, nil
		}

		agentPoolVMSetNamesMap := make(map[string]bool)
		agentPoolVMSetNames, err := az.VMSet.GetAgentPoolVMSetNames(nodes)
		if err != nil {
			return nil, fmt.Errorf("ListManagedLBs: failed to get agent pool vmSet names: %w", err)
		}

		if agentPoolVMSetNames != nil && len(*agentPoolVMSetNames) > 0 {
			for _, vmSetName := range *agentPoolVMSetNames {
				klog.V(6).Infof("ListManagedLBs: found agent pool vmSet name %s", vmSetName)
				agentPoolVMSetNamesMap[strings.ToLower(vmSetName)] = true
			}
		}

		for agentPoolVMSetName := range agentPoolVMSetNamesMap {
			managedLBNames.Insert(az.mapVMSetNameToLoadBalancerName(agentPoolVMSetName, clusterName))
		}
	}

	if az.useMultipleStandardLoadBalancers() {
		for _, multiSLBConfig := range az.MultipleStandardLoadBalancerConfigurations {
			managedLBNames.Insert(multiSLBConfig.Name, fmt.Sprintf("%s%s", multiSLBConfig.Name, consts.InternalLoadBalancerNameSuffix))
		}
	}

	for _, lb := range allLBs {
		if managedLBNames.Has(trimSuffixIgnoreCase(pointer.StringDeref(lb.Name, ""), consts.InternalLoadBalancerNameSuffix)) {
			managedLBs = append(managedLBs, lb)
			klog.V(4).Infof("ListManagedLBs: found managed LB %s", pointer.StringDeref(lb.Name, ""))
		}
	}

	return &managedLBs, nil
}

// CreateOrUpdateLB invokes az.LoadBalancerClient.CreateOrUpdate with exponential backoff retry
func (az *Cloud) CreateOrUpdateLB(service *v1.Service, lb network.LoadBalancer) error {
	ctx, cancel := getContextWithCancel()
	defer cancel()

	lb = cleanupSubnetInFrontendIPConfigurations(&lb)

	rgName := az.getLoadBalancerResourceGroup()
	rerr := az.LoadBalancerClient.CreateOrUpdate(ctx, rgName, pointer.StringDeref(lb.Name, ""), lb, pointer.StringDeref(lb.Etag, ""))
	klog.V(10).Infof("LoadBalancerClient.CreateOrUpdate(%s): end", *lb.Name)
	if rerr == nil {
		// Invalidate the cache right after updating
		_ = az.lbCache.Delete(*lb.Name)
		return nil
	}

	lbJSON, _ := json.Marshal(lb)
	klog.Warningf("LoadBalancerClient.CreateOrUpdate(%s) failed: %v, LoadBalancer request: %s", pointer.StringDeref(lb.Name, ""), rerr.Error(), string(lbJSON))

	// Invalidate the cache because ETAG precondition mismatch.
	if rerr.HTTPStatusCode == http.StatusPreconditionFailed {
		klog.V(3).Infof("LoadBalancer cache for %s is cleanup because of http.StatusPreconditionFailed", pointer.StringDeref(lb.Name, ""))
		_ = az.lbCache.Delete(*lb.Name)
	}

	retryErrorMessage := rerr.Error().Error()
	// Invalidate the cache because another new operation has canceled the current request.
	if strings.Contains(strings.ToLower(retryErrorMessage), consts.OperationCanceledErrorMessage) {
		klog.V(3).Infof("LoadBalancer cache for %s is cleanup because CreateOrUpdate is canceled by another operation", pointer.StringDeref(lb.Name, ""))
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
		klog.V(3).Infof("The public IP %s referenced by load balancer %s is not in Succeeded provisioning state, will try to update it", pipName, pointer.StringDeref(lb.Name, ""))
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

	klog.V(4).Infof("CreateOrUpdateLBBackendPool: updating backend pool %s in LB %s", pointer.StringDeref(backendPool.Name, ""), lbName)
	rerr := az.LoadBalancerClient.CreateOrUpdateBackendPools(ctx, az.getLoadBalancerResourceGroup(), lbName, pointer.StringDeref(backendPool.Name, ""), backendPool, pointer.StringDeref(backendPool.Etag, ""))
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

// MigrateToIPBasedBackendPoolAndWaitForCompletion use the migration API to migrate from
// NIC-based to IP-based LB backend pools. It also makes sure the number of IP addresses
// in the backend pools is expected.
func (az *Cloud) MigrateToIPBasedBackendPoolAndWaitForCompletion(
	lbName string, backendPoolNames []string, nicsCountMap map[string]int,
) error {
	if rerr := az.LoadBalancerClient.MigrateToIPBasedBackendPool(context.Background(), az.ResourceGroup, lbName, backendPoolNames); rerr != nil {
		backendPoolNamesStr := strings.Join(backendPoolNames, ",")
		klog.Errorf("MigrateToIPBasedBackendPoolAndWaitForCompletion: Failed to migrate to IP based backend pool for lb %s, backend pool %s: %s", lbName, backendPoolNamesStr, rerr.Error().Error())
		return rerr.Error()
	}

	succeeded := make(map[string]bool)
	for bpName := range nicsCountMap {
		succeeded[bpName] = false
	}

	err := wait.PollImmediate(5*time.Second, 10*time.Minute, func() (done bool, err error) {
		for bpName, nicsCount := range nicsCountMap {
			if succeeded[bpName] {
				continue
			}

			bp, rerr := az.LoadBalancerClient.GetLBBackendPool(context.Background(), az.ResourceGroup, lbName, bpName, "")
			if rerr != nil {
				klog.Errorf("MigrateToIPBasedBackendPoolAndWaitForCompletion: Failed to get backend pool %s for lb %s: %s", bpName, lbName, rerr.Error().Error())
				return false, rerr.Error()
			}

			if countIPsOnBackendPool(bp) != nicsCount {
				klog.V(4).Infof("MigrateToIPBasedBackendPoolAndWaitForCompletion: Expected IPs %d, current IPs %d, will retry in 5s", nicsCount, countIPsOnBackendPool(bp))
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

		klog.Errorf("MigrateToIPBasedBackendPoolAndWaitForCompletion: Failed to wait for migration to IP based backend pool for lb %s, backend pool %s: %s", lbName, strings.Join(backendPoolNames, ","), err.Error())
		return err
	}

	return nil
}

func (az *Cloud) newLBCache() (azcache.Resource, error) {
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
	return azcache.NewTimedCache(time.Duration(az.LoadBalancerCacheTTLInSeconds)*time.Second, getter, az.Config.DisableAPICallCache)
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

// isBackendPoolOnSameLB checks whether newBackendPoolID is on the same load balancer as existingBackendPools.
// Since both public and internal LBs are supported, lbName and lbName-internal are treated as same.
// If not same, the lbName for existingBackendPools would also be returned.
func isBackendPoolOnSameLB(newBackendPoolID string, existingBackendPools []string) (bool, string, error) {
	matches := backendPoolIDRE.FindStringSubmatch(newBackendPoolID)
	if len(matches) != 2 {
		return false, "", fmt.Errorf("new backendPoolID %q is in wrong format", newBackendPoolID)
	}

	newLBName := matches[1]
	newLBNameTrimmed := trimSuffixIgnoreCase(newLBName, consts.InternalLoadBalancerNameSuffix)
	for _, backendPool := range existingBackendPools {
		matches := backendPoolIDRE.FindStringSubmatch(backendPool)
		if len(matches) != 2 {
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
