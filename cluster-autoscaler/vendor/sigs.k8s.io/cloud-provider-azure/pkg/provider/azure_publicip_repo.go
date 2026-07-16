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
	"net/netip"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v9"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"

	azcache "sigs.k8s.io/cloud-provider-azure/pkg/cache"
	"sigs.k8s.io/cloud-provider-azure/pkg/consts"
	"sigs.k8s.io/cloud-provider-azure/pkg/log"
	providererrors "sigs.k8s.io/cloud-provider-azure/pkg/provider/errors"
	"sigs.k8s.io/cloud-provider-azure/pkg/util/deepcopy"
)

var (
	azureReservedIPPrefixes = []netip.Prefix{
		netip.MustParsePrefix("100.64.0.0/10"),
		netip.MustParsePrefix("168.63.129.16/32"),
	}
)

// CreateOrUpdatePIP invokes az.NetworkClientFactory.GetPublicIPAddressClient().CreateOrUpdate with exponential backoff retry
func (az *Cloud) CreateOrUpdatePIP(service *v1.Service, pipResourceGroup string, pip *armnetwork.PublicIPAddress) error {
	ctx, cancel := getContextWithCancel()
	defer cancel()

	logger := log.FromContextOrBackground(ctx).WithName("CreateOrUpdatePIP")

	_, rerr := az.NetworkClientFactory.GetPublicIPAddressClient().CreateOrUpdate(ctx, pipResourceGroup, ptr.Deref(pip.Name, ""), *pip)
	logger.V(10).Info("NetworkClientFactory.GetPublicIPAddressClient().CreateOrUpdate end", "pipResourceGroup", pipResourceGroup, "pipName", ptr.Deref(pip.Name, ""))
	if rerr == nil {
		// Invalidate the cache right after updating
		_ = az.pipCache.Delete(pipResourceGroup)
		return nil
	}

	pipJSON, _ := json.Marshal(pip)
	klog.Warningf("NetworkClientFactory.GetPublicIPAddressClient().CreateOrUpdate(%s, %s) failed: %s, PublicIP request: %s", pipResourceGroup, ptr.Deref(pip.Name, ""), rerr.Error(), string(pipJSON))
	az.Event(service, v1.EventTypeWarning, "CreateOrUpdatePublicIPAddress", rerr.Error())

	// Invalidate the cache because ETAG precondition mismatch.
	var respError *azcore.ResponseError
	if errors.As(rerr, &respError) && respError != nil {
		if respError.StatusCode == http.StatusPreconditionFailed {
			logger.V(3).Info("PublicIP cache is cleanup because of http.StatusPreconditionFailed", "pipResourceGroup", pipResourceGroup, "pipName", ptr.Deref(pip.Name, ""))
			_ = az.pipCache.Delete(pipResourceGroup)
		}
	}

	retryErrorMessage := rerr.Error()
	// Invalidate the cache because another new operation has canceled the current request.
	if strings.Contains(strings.ToLower(retryErrorMessage), consts.OperationCanceledErrorMessage) {
		logger.V(3).Info("PublicIP cache is cleanup because CreateOrUpdate is canceled by another operation", "pipResourceGroup", pipResourceGroup, "pipName", ptr.Deref(pip.Name, ""))
		_ = az.pipCache.Delete(pipResourceGroup)
	}

	return rerr
}

// DeletePublicIP invokes az.NetworkClientFactory.GetPublicIPAddressClient().Delete with exponential backoff retry
func (az *Cloud) DeletePublicIP(service *v1.Service, pipResourceGroup string, pipName string) error {
	ctx, cancel := getContextWithCancel()
	defer cancel()
	logger := log.FromContextOrBackground(ctx).WithName("DeletePublicIP")

	rerr := az.NetworkClientFactory.GetPublicIPAddressClient().Delete(ctx, pipResourceGroup, pipName)
	if rerr != nil {
		logger.Error(rerr, "NetworkClientFactory.GetPublicIPAddressClient().Delete failed", "pipName", pipName)
		az.Event(service, v1.EventTypeWarning, "DeletePublicIPAddress", rerr.Error())

		if strings.Contains(rerr.Error(), consts.CannotDeletePublicIPErrorMessageCode) {
			klog.Warningf("DeletePublicIP for public IP %s failed with error %v, this is because other resources are referencing the public IP. The deletion of the service will continue.", pipName, rerr)
			return nil
		}
		return rerr
	}

	// Invalidate the cache right after deleting
	_ = az.pipCache.Delete(pipResourceGroup)
	return nil
}

func (az *Cloud) newPIPCache() (azcache.Resource, error) {
	getter := func(ctx context.Context, key string) (interface{}, error) {
		pipResourceGroup := key
		pipList, rerr := az.NetworkClientFactory.GetPublicIPAddressClient().List(ctx, pipResourceGroup)
		if rerr != nil {
			return nil, rerr
		}

		pipMap := &sync.Map{}
		for _, pip := range pipList {
			pip := pip
			pipMap.Store(strings.ToLower(ptr.Deref(pip.Name, "")), pip)
		}
		return pipMap, nil
	}

	if az.PublicIPCacheTTLInSeconds == 0 {
		az.PublicIPCacheTTLInSeconds = publicIPCacheTTLDefaultInSeconds
	}
	return azcache.NewTimedCache(time.Duration(az.PublicIPCacheTTLInSeconds)*time.Second, getter, az.DisableAPICallCache)
}

func (az *Cloud) getPublicIPAddress(ctx context.Context, pipResourceGroup string, pipName string, crt azcache.AzureCacheReadType) (*armnetwork.PublicIPAddress, bool, error) {
	logger := klog.FromContext(ctx).WithName("getPublicIPAddress").
		WithValues("pipResourceGroup", pipResourceGroup, "pipName", pipName)
	cached, err := az.pipCache.Get(ctx, pipResourceGroup, crt)
	if err != nil {
		return nil, false, err
	}

	pips := cached.(*sync.Map)
	pip, ok := pips.Load(strings.ToLower(pipName))
	if !ok {
		// pip not found, refresh cache and retry
		cached, err = az.pipCache.Get(ctx, pipResourceGroup, azcache.CacheReadTypeForceRefresh)
		if err != nil {
			return nil, false, err
		}
		pips = cached.(*sync.Map)
		pip, ok = pips.Load(strings.ToLower(pipName))
		if !ok {
			logger.V(4).Info("pip not found")
			return nil, false, nil
		}
	}

	pip = pip.(*armnetwork.PublicIPAddress)
	return (deepcopy.Copy(pip).(*armnetwork.PublicIPAddress)), true, nil
}

func (az *Cloud) listPIP(ctx context.Context, pipResourceGroup string, crt azcache.AzureCacheReadType) ([]*armnetwork.PublicIPAddress, error) {
	cached, err := az.pipCache.Get(ctx, pipResourceGroup, crt)
	if err != nil {
		return nil, err
	}
	pips := cached.(*sync.Map)
	var ret []*armnetwork.PublicIPAddress
	pips.Range(func(_, value interface{}) bool {
		pip := value.(*armnetwork.PublicIPAddress)
		// Deep-copy so callers cannot mutate the cache via the returned slice.
		// This mirrors getPublicIPAddress and keeps failed PUTs from poisoning the cache.
		ret = append(ret, deepcopy.Copy(pip).(*armnetwork.PublicIPAddress))
		return true
	})
	return ret, nil
}

func (az *Cloud) findMatchedPIP(ctx context.Context, loadBalancerIP, pipName, pipResourceGroup string) (pip *armnetwork.PublicIPAddress, err error) {
	if loadBalancerIP != "" {
		if err := validatePublicIPCandidate(loadBalancerIP); err != nil {
			return nil, err
		}
	}

	pips, err := az.listPIP(ctx, pipResourceGroup, azcache.CacheReadTypeDefault)
	if err != nil {
		return nil, fmt.Errorf("findMatchedPIP: failed to listPIP: %w", err)
	}

	if loadBalancerIP != "" {
		pip, err = az.findMatchedPIPByLoadBalancerIP(ctx, pips, loadBalancerIP, pipResourceGroup)
		if err != nil {
			return nil, err
		}
		return pip, nil
	}

	if pipResourceGroup != "" {
		pip, err = az.findMatchedPIPByName(ctx, pips, pipName, pipResourceGroup)
		if err != nil {
			return nil, err
		}
	}
	return pip, nil
}

func (az *Cloud) findMatchedPIPByName(ctx context.Context, pips []*armnetwork.PublicIPAddress, pipName, pipResourceGroup string) (*armnetwork.PublicIPAddress, error) {
	for _, pip := range pips {
		if strings.EqualFold(ptr.Deref(pip.Name, ""), pipName) {
			return pip, nil
		}
	}

	pipList, err := az.listPIP(ctx, pipResourceGroup, azcache.CacheReadTypeForceRefresh)
	if err != nil {
		return nil, fmt.Errorf("findMatchedPIPByName: failed to listPIP force refresh: %w", err)
	}
	for _, pip := range pipList {
		if strings.EqualFold(ptr.Deref(pip.Name, ""), pipName) {
			return pip, nil
		}
	}

	return nil, fmt.Errorf("findMatchedPIPByName: failed to find PIP %s in resource group %s", pipName, pipResourceGroup)
}

func (az *Cloud) findMatchedPIPByLoadBalancerIP(ctx context.Context, pips []*armnetwork.PublicIPAddress, loadBalancerIP, pipResourceGroup string) (*armnetwork.PublicIPAddress, error) {
	pip, err := getExpectedPIPFromListByIPAddress(pips, loadBalancerIP)
	if err != nil {
		pipList, err := az.listPIP(ctx, pipResourceGroup, azcache.CacheReadTypeForceRefresh)
		if err != nil {
			return nil, fmt.Errorf("findMatchedPIPByLoadBalancerIP: failed to listPIP force refresh: %w", err)
		}

		pip, err = getExpectedPIPFromListByIPAddress(pipList, loadBalancerIP)
		if err != nil {
			return nil, fmt.Errorf("findMatchedPIPByLoadBalancerIP: cannot find public IP with IP address %s in resource group %s: %w", loadBalancerIP, pipResourceGroup, err)
		}
	}

	return pip, nil
}

func validatePublicIPCandidate(ip string) error {
	addr, err := netip.ParseAddr(ip)
	if err != nil {
		return providererrors.NewInvalidLoadBalancerIPError(ip)
	}
	addr = addr.Unmap()
	if !addr.IsGlobalUnicast() || addr.IsPrivate() || isAzureReservedIP(addr) {
		return providererrors.NewNonPublicLoadBalancerIPError(ip)
	}
	return nil
}

// isAzureReservedIP reports whether addr belongs to an Azure reserved Virtual
// Network address range not covered by netip.Addr.IsPrivate:
// https://learn.microsoft.com/azure/virtual-network/virtual-networks-faq#what-address-ranges-can-i-use-in-my-virtual-networks
func isAzureReservedIP(addr netip.Addr) bool {
	for _, prefix := range azureReservedIPPrefixes {
		if prefix.Contains(addr) {
			return true
		}
	}
	return false
}

func getExpectedPIPFromListByIPAddress(pips []*armnetwork.PublicIPAddress, ip string) (*armnetwork.PublicIPAddress, error) {
	for _, pip := range pips {
		if pip.Properties.IPAddress != nil &&
			*pip.Properties.IPAddress == ip {
			return pip, nil
		}
	}

	return nil, fmt.Errorf("getExpectedPIPFromListByIPAddress: cannot find public IP with IP address %s", ip)
}

func getPIPRGFromID(pipIDLower string) (string, error) {
	re := regexp.MustCompile(strings.ToLower(`/subscriptions/(?:.*)/resourceGroups/(.+)/providers/Microsoft.Network/publicIPAddresses/(?:.*)`))
	matches := re.FindStringSubmatch(pipIDLower)
	if len(matches) != 2 {
		return "", fmt.Errorf("failed to extract resource group name from public IP ID %s", pipIDLower)
	}
	return matches[1], nil
}
