/*
Copyright 2024 The Kubernetes Authors.

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

package privatelinkservice

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v9"

	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/privatelinkserviceclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/cache"
	"sigs.k8s.io/cloud-provider-azure/pkg/consts"
)

const (
	DefaultCacheTTL = 120 * time.Second
)

func NewCache(
	client privatelinkserviceclient.Interface,
	cacheTTL time.Duration,
	disableAPICallCache bool,
) (cache.Resource, error) {
	getter := func(ctx context.Context, key string) (interface{}, error) {
		resourceGroup, frontendID := parsePLSCacheKey(key)
		plsList, err := client.List(ctx, resourceGroup)
		if err != nil {
			return nil, err
		}

		for i := range plsList {
			pls := plsList[i]
			if pls.Properties == nil {
				continue
			}
			for _, fipConfig := range pls.Properties.LoadBalancerFrontendIPConfigurations {
				if strings.EqualFold(*fipConfig.ID, frontendID) {
					return pls, nil
				}
			}
		}

		notFound := &armnetwork.PrivateLinkService{
			ID: to.Ptr(consts.PrivateLinkServiceNotExistID),
		}
		return notFound, nil
	}

	if cacheTTL == 0 {
		cacheTTL = DefaultCacheTTL
	}
	return cache.NewTimedCache(cacheTTL, getter, disableAPICallCache)
}

func getPLSCacheKey(resourceGroup, plsLBFrontendID string) string {
	return fmt.Sprintf("%s*%s", resourceGroup, plsLBFrontendID)
}

func parsePLSCacheKey(key string) (string, string) {
	splits := strings.Split(key, "*")
	return splits[0], splits[1]
}
