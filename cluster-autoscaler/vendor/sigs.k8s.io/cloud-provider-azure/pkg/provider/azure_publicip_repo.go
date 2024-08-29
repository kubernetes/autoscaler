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
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2022-07-01/network"

	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"k8s.io/utils/pointer"

	azcache "sigs.k8s.io/cloud-provider-azure/pkg/cache"
	"sigs.k8s.io/cloud-provider-azure/pkg/consts"
	"sigs.k8s.io/cloud-provider-azure/pkg/util/deepcopy"
)

// CreateOrUpdatePIP invokes az.PublicIPAddressesClient.CreateOrUpdate with exponential backoff retry
func (az *Cloud) CreateOrUpdatePIP(service *v1.Service, pipResourceGroup string, pip network.PublicIPAddress) error {
	ctx, cancel := getContextWithCancel()
	defer cancel()

	rerr := az.PublicIPAddressesClient.CreateOrUpdate(ctx, pipResourceGroup, pointer.StringDeref(pip.Name, ""), pip)
	klog.V(10).Infof("PublicIPAddressesClient.CreateOrUpdate(%s, %s): end", pipResourceGroup, pointer.StringDeref(pip.Name, ""))
	if rerr == nil {
		// Invalidate the cache right after updating
		_ = az.pipCache.Delete(pipResourceGroup)
		return nil
	}

	pipJSON, _ := json.Marshal(pip)
	klog.Warningf("PublicIPAddressesClient.CreateOrUpdate(%s, %s) failed: %s, PublicIP request: %s", pipResourceGroup, pointer.StringDeref(pip.Name, ""), rerr.Error().Error(), string(pipJSON))
	az.Event(service, v1.EventTypeWarning, "CreateOrUpdatePublicIPAddress", rerr.Error().Error())

	// Invalidate the cache because ETAG precondition mismatch.
	if rerr.HTTPStatusCode == http.StatusPreconditionFailed {
		klog.V(3).Infof("PublicIP cache for (%s, %s) is cleanup because of http.StatusPreconditionFailed", pipResourceGroup, pointer.StringDeref(pip.Name, ""))
		_ = az.pipCache.Delete(pipResourceGroup)
	}

	retryErrorMessage := rerr.Error().Error()
	// Invalidate the cache because another new operation has canceled the current request.
	if strings.Contains(strings.ToLower(retryErrorMessage), consts.OperationCanceledErrorMessage) {
		klog.V(3).Infof("PublicIP cache for (%s, %s) is cleanup because CreateOrUpdate is canceled by another operation", pipResourceGroup, pointer.StringDeref(pip.Name, ""))
		_ = az.pipCache.Delete(pipResourceGroup)
	}

	return rerr.Error()
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
	_ = az.pipCache.Delete(pipResourceGroup)
	return nil
}

func (az *Cloud) newPIPCache() (azcache.Resource, error) {
	getter := func(key string) (interface{}, error) {
		ctx, cancel := getContextWithCancel()
		defer cancel()

		pipResourceGroup := key
		pipList, rerr := az.PublicIPAddressesClient.List(ctx, pipResourceGroup)
		if rerr != nil {
			return nil, rerr.Error()
		}

		pipMap := &sync.Map{}
		for _, pip := range pipList {
			pip := pip
			pipMap.Store(strings.ToLower(pointer.StringDeref(pip.Name, "")), &pip)
		}
		return pipMap, nil
	}

	if az.PublicIPCacheTTLInSeconds == 0 {
		az.PublicIPCacheTTLInSeconds = publicIPCacheTTLDefaultInSeconds
	}
	return azcache.NewTimedCache(time.Duration(az.PublicIPCacheTTLInSeconds)*time.Second, getter, az.Config.DisableAPICallCache)
}

func (az *Cloud) getPublicIPAddress(pipResourceGroup string, pipName string, crt azcache.AzureCacheReadType) (network.PublicIPAddress, bool, error) {
	cached, err := az.pipCache.Get(pipResourceGroup, crt)
	if err != nil {
		return network.PublicIPAddress{}, false, err
	}

	pips := cached.(*sync.Map)
	pip, ok := pips.Load(strings.ToLower(pipName))
	if !ok {
		// pip not found, refresh cache and retry
		cached, err = az.pipCache.Get(pipResourceGroup, azcache.CacheReadTypeForceRefresh)
		if err != nil {
			return network.PublicIPAddress{}, false, err
		}
		pips = cached.(*sync.Map)
		pip, ok = pips.Load(strings.ToLower(pipName))
		if !ok {
			return network.PublicIPAddress{}, false, nil
		}
	}

	pip = pip.(*network.PublicIPAddress)
	return *(deepcopy.Copy(pip).(*network.PublicIPAddress)), true, nil
}

func (az *Cloud) listPIP(pipResourceGroup string, crt azcache.AzureCacheReadType) ([]network.PublicIPAddress, error) {
	cached, err := az.pipCache.Get(pipResourceGroup, crt)
	if err != nil {
		return nil, err
	}
	pips := cached.(*sync.Map)
	var ret []network.PublicIPAddress
	pips.Range(func(key, value interface{}) bool {
		pip := value.(*network.PublicIPAddress)
		ret = append(ret, *pip)
		return true
	})
	return ret, nil
}

func (az *Cloud) findMatchedPIP(loadBalancerIP, pipName, pipResourceGroup string) (pip *network.PublicIPAddress, err error) {
	pips, err := az.listPIP(pipResourceGroup, azcache.CacheReadTypeDefault)
	if err != nil {
		return nil, fmt.Errorf("findMatchedPIPByLoadBalancerIP: failed to listPIP: %w", err)
	}

	if loadBalancerIP != "" {
		pip, err = az.findMatchedPIPByLoadBalancerIP(&pips, loadBalancerIP, pipResourceGroup)
		if err != nil {
			return nil, err
		}
		return pip, nil
	}

	if pipResourceGroup != "" {
		pip, err = az.findMatchedPIPByName(&pips, pipName, pipResourceGroup)
		if err != nil {
			return nil, err
		}
	}
	return pip, nil
}

func (az *Cloud) findMatchedPIPByName(pips *[]network.PublicIPAddress, pipName, pipResourceGroup string) (*network.PublicIPAddress, error) {
	for _, pip := range *pips {
		if strings.EqualFold(pointer.StringDeref(pip.Name, ""), pipName) {
			return &pip, nil
		}
	}

	pipList, err := az.listPIP(pipResourceGroup, azcache.CacheReadTypeForceRefresh)
	if err != nil {
		return nil, fmt.Errorf("findMatchedPIPByName: failed to listPIP force refresh: %w", err)
	}
	for _, pip := range pipList {
		if strings.EqualFold(pointer.StringDeref(pip.Name, ""), pipName) {
			return &pip, nil
		}
	}

	return nil, fmt.Errorf("findMatchedPIPByName: failed to find PIP %s in resource group %s", pipName, pipResourceGroup)
}

func (az *Cloud) findMatchedPIPByLoadBalancerIP(pips *[]network.PublicIPAddress, loadBalancerIP, pipResourceGroup string) (*network.PublicIPAddress, error) {
	pip, err := getExpectedPIPFromListByIPAddress(*pips, loadBalancerIP)
	if err != nil {
		pipList, err := az.listPIP(pipResourceGroup, azcache.CacheReadTypeForceRefresh)
		if err != nil {
			return nil, fmt.Errorf("findMatchedPIPByLoadBalancerIP: failed to listPIP force refresh: %w", err)
		}

		pip, err = getExpectedPIPFromListByIPAddress(pipList, loadBalancerIP)
		if err != nil {
			return nil, fmt.Errorf("findMatchedPIPByLoadBalancerIP: cannot find public IP with IP address %s in resource group %s", loadBalancerIP, pipResourceGroup)
		}
	}

	return pip, nil
}

func getExpectedPIPFromListByIPAddress(pips []network.PublicIPAddress, ip string) (*network.PublicIPAddress, error) {
	for _, pip := range pips {
		if pip.PublicIPAddressPropertiesFormat.IPAddress != nil &&
			*pip.PublicIPAddressPropertiesFormat.IPAddress == ip {
			return &pip, nil
		}
	}

	return nil, fmt.Errorf("getExpectedPIPFromListByIPAddress: cannot find public IP with IP address %s", ip)
}
