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
	"time"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2022-07-01/network"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"k8s.io/utils/pointer"

	azcache "sigs.k8s.io/cloud-provider-azure/pkg/cache"
	"sigs.k8s.io/cloud-provider-azure/pkg/consts"
	"sigs.k8s.io/cloud-provider-azure/pkg/retry"
)

func (az *Cloud) CreateOrUpdatePLS(_ *v1.Service, resourceGroup string, pls network.PrivateLinkService) error {
	ctx, cancel := getContextWithCancel()
	defer cancel()

	rerr := az.PrivateLinkServiceClient.CreateOrUpdate(ctx, resourceGroup, pointer.StringDeref(pls.Name, ""), pls, pointer.StringDeref(pls.Etag, ""))
	if rerr == nil {
		// Invalidate the cache right after updating
		_ = az.plsCache.Delete(getPLSCacheKey(resourceGroup, pointer.StringDeref((*pls.LoadBalancerFrontendIPConfigurations)[0].ID, "")))
		return nil
	}

	rtJSON, _ := json.Marshal(pls)
	klog.Warningf("PrivateLinkServiceClient.CreateOrUpdate(%s) failed: %v, PrivateLinkService request: %s", pointer.StringDeref(pls.Name, ""), rerr.Error(), string(rtJSON))

	// Invalidate the cache because etag mismatch.
	if rerr.HTTPStatusCode == http.StatusPreconditionFailed {
		klog.V(3).Infof("Private link service cache for %s is cleanup because of http.StatusPreconditionFailed", pointer.StringDeref(pls.Name, ""))
		_ = az.plsCache.Delete(getPLSCacheKey(resourceGroup, pointer.StringDeref((*pls.LoadBalancerFrontendIPConfigurations)[0].ID, "")))
	}
	// Invalidate the cache because another new operation has canceled the current request.
	if strings.Contains(strings.ToLower(rerr.Error().Error()), consts.OperationCanceledErrorMessage) {
		klog.V(3).Infof("Private link service for %s is cleanup because CreateOrUpdatePrivateLinkService is canceled by another operation", pointer.StringDeref(pls.Name, ""))
		_ = az.plsCache.Delete(getPLSCacheKey(resourceGroup, pointer.StringDeref((*pls.LoadBalancerFrontendIPConfigurations)[0].ID, "")))
	}
	klog.Errorf("PrivateLinkServiceClient.CreateOrUpdate(%s) failed: %v", pointer.StringDeref(pls.Name, ""), rerr.Error())
	return rerr.Error()
}

// DeletePLS invokes az.PrivateLinkServiceClient.Delete with exponential backoff retry
func (az *Cloud) DeletePLS(service *v1.Service, resourceGroup, plsName, plsLBFrontendID string) *retry.Error {
	ctx, cancel := getContextWithCancel()
	defer cancel()

	rerr := az.PrivateLinkServiceClient.Delete(ctx, resourceGroup, plsName)
	if rerr == nil {
		// Invalidate the cache right after deleting
		_ = az.plsCache.Delete(getPLSCacheKey(resourceGroup, plsLBFrontendID))
		return nil
	}

	klog.Errorf("PrivateLinkServiceClient.DeletePLS(%s) failed: %s", plsName, rerr.Error().Error())
	az.Event(service, v1.EventTypeWarning, "DeletePrivateLinkService", rerr.Error().Error())
	return rerr
}

// DeletePEConn invokes az.PrivateLinkServiceClient.DeletePEConnection with exponential backoff retry
func (az *Cloud) DeletePEConn(service *v1.Service, resourceGroup, plsName, peConnName string) *retry.Error {
	ctx, cancel := getContextWithCancel()
	defer cancel()

	rerr := az.PrivateLinkServiceClient.DeletePEConnection(ctx, resourceGroup, plsName, peConnName)
	if rerr == nil {
		return nil
	}

	klog.Errorf("PrivateLinkServiceClient.DeletePEConnection(%s-%s) failed: %s", plsName, peConnName, rerr.Error().Error())
	az.Event(service, v1.EventTypeWarning, "DeletePrivateEndpointConnection", rerr.Error().Error())
	return rerr
}

func (az *Cloud) newPLSCache() (azcache.Resource, error) {
	// for PLS cache, key is LBFrontendIPConfiguration ID
	getter := func(key string) (interface{}, error) {
		ctx, cancel := getContextWithCancel()
		defer cancel()
		resourceGroup, frontendID := parsePLSCacheKey(key)
		plsList, err := az.PrivateLinkServiceClient.List(ctx, resourceGroup)
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
					if strings.EqualFold(*fipConfig.ID, frontendID) {
						return &pls, nil
					}
				}

			}
		}

		klog.V(2).Infof("No privateLinkService found for frontendIPConfig %q in rg %q", frontendID, resourceGroup)
		plsNotExistID := consts.PrivateLinkServiceNotExistID
		return &network.PrivateLinkService{ID: &plsNotExistID}, nil
	}

	if az.PlsCacheTTLInSeconds == 0 {
		az.PlsCacheTTLInSeconds = plsCacheTTLDefaultInSeconds
	}
	return azcache.NewTimedCache(time.Duration(az.PlsCacheTTLInSeconds)*time.Second, getter, az.Config.DisableAPICallCache)
}

func (az *Cloud) getPrivateLinkService(resourceGroup string, frontendIPConfigID *string, crt azcache.AzureCacheReadType) (pls network.PrivateLinkService, err error) {
	cachedPLS, err := az.plsCache.GetWithDeepCopy(getPLSCacheKey(resourceGroup, *frontendIPConfigID), crt)
	if err != nil {
		return pls, err
	}
	return *(cachedPLS.(*network.PrivateLinkService)), nil
}

func getPLSCacheKey(resourceGroup, plsLBFrontendID string) string {
	return fmt.Sprintf("%s*%s", resourceGroup, plsLBFrontendID)
}

func parsePLSCacheKey(key string) (string, string) {
	splits := strings.Split(key, "*")
	return splits[0], splits[1]
}
