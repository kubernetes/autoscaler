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
	"net/http"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2022-07-01/network"
	"k8s.io/klog/v2"
	"k8s.io/utils/pointer"

	azcache "sigs.k8s.io/cloud-provider-azure/pkg/cache"
	"sigs.k8s.io/cloud-provider-azure/pkg/consts"
)

// CreateOrUpdateRouteTable invokes az.RouteTablesClient.CreateOrUpdate with exponential backoff retry
func (az *Cloud) CreateOrUpdateRouteTable(routeTable network.RouteTable) error {
	ctx, cancel := getContextWithCancel()
	defer cancel()

	rerr := az.RouteTablesClient.CreateOrUpdate(ctx, az.RouteTableResourceGroup, az.RouteTableName, routeTable, pointer.StringDeref(routeTable.Etag, ""))
	if rerr == nil {
		// Invalidate the cache right after updating
		_ = az.rtCache.Delete(*routeTable.Name)
		return nil
	}

	rtJSON, _ := json.Marshal(routeTable)
	klog.Warningf("RouteTablesClient.CreateOrUpdate(%s) failed: %v, RouteTable request: %s", pointer.StringDeref(routeTable.Name, ""), rerr.Error(), string(rtJSON))

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

func (az *Cloud) newRouteTableCache() (azcache.Resource, error) {
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
	return azcache.NewTimedCache(time.Duration(az.RouteTableCacheTTLInSeconds)*time.Second, getter, az.Config.DisableAPICallCache)
}
