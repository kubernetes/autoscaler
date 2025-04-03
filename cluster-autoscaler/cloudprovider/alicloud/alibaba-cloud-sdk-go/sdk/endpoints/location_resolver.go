/*
Copyright 2018 The Kubernetes Authors.

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

package endpoints

import (
	"encoding/json"
	"sync"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/alicloud/alibaba-cloud-sdk-go/sdk/requests"
	"k8s.io/klog/v2"
)

const (
	// EndpointCacheExpireTime in seconds
	EndpointCacheExpireTime = 3600
)

var lastClearTimePerProduct = struct {
	sync.RWMutex
	cache map[string]int64
}{cache: make(map[string]int64)}

var endpointCache = struct {
	sync.RWMutex
	cache map[string]string
}{cache: make(map[string]string)}

// LocationResolver is kind of resolver
type LocationResolver struct{}

// TryResolve return endpoint
func (resolver *LocationResolver) TryResolve(param *ResolveParam) (endpoint string, support bool, err error) {
	if len(param.LocationProduct) <= 0 {
		support = false
		return
	}

	//get from cache
	cacheKey := param.Product + "#" + param.RegionId
	if endpointCache.cache != nil && len(endpointCache.cache[cacheKey]) > 0 && !CheckCacheIsExpire(cacheKey) {
		endpoint = endpointCache.cache[cacheKey]
		support = true
		return
	}

	//get from remote
	getEndpointRequest := requests.NewCommonRequest()

	getEndpointRequest.Product = "Location"
	getEndpointRequest.Version = "2015-06-12"
	getEndpointRequest.ApiName = "DescribeEndpoints"
	getEndpointRequest.Domain = "location.aliyuncs.com"
	getEndpointRequest.Method = "GET"
	getEndpointRequest.Scheme = requests.HTTPS

	getEndpointRequest.QueryParams["Id"] = param.RegionId
	getEndpointRequest.QueryParams["ServiceCode"] = param.LocationProduct
	if len(param.LocationEndpointType) > 0 {
		getEndpointRequest.QueryParams["Type"] = param.LocationEndpointType
	} else {
		getEndpointRequest.QueryParams["Type"] = "openAPI"
	}

	response, err := param.CommonApi(getEndpointRequest)
	if err != nil {
		klog.Errorf("failed to resolve endpoint, error: %v", err)
		support = false
		return
	}

	var getEndpointResponse GetEndpointResponse
	if !response.IsSuccess() {
		support = false
		return
	}

	err = json.Unmarshal([]byte(response.GetHttpContentString()), &getEndpointResponse)
	if err != nil {
		klog.Errorf("failed to unmarshal endpoint response, error: %v", err)
		support = false
		return
	}

	if !getEndpointResponse.Success || getEndpointResponse.Endpoints == nil {
		support = false
		return
	}
	if len(getEndpointResponse.Endpoints.Endpoint) <= 0 {
		support = false
		return
	}
	if len(getEndpointResponse.Endpoints.Endpoint[0].Endpoint) > 0 {
		endpoint = getEndpointResponse.Endpoints.Endpoint[0].Endpoint
		endpointCache.Lock()
		endpointCache.cache[cacheKey] = endpoint
		endpointCache.Unlock()
		lastClearTimePerProduct.Lock()
		lastClearTimePerProduct.cache[cacheKey] = time.Now().Unix()
		lastClearTimePerProduct.Unlock()
		support = true
		return
	}

	support = false
	return
}

// CheckCacheIsExpire valid the cacheKey
func CheckCacheIsExpire(cacheKey string) bool {
	lastClearTime := lastClearTimePerProduct.cache[cacheKey]
	if lastClearTime <= 0 {
		lastClearTime = time.Now().Unix()
		lastClearTimePerProduct.Lock()
		lastClearTimePerProduct.cache[cacheKey] = lastClearTime
		lastClearTimePerProduct.Unlock()
	}

	now := time.Now().Unix()
	elapsedTime := now - lastClearTime
	if elapsedTime > EndpointCacheExpireTime {
		return true
	}

	return false
}

// GetEndpointResponse returns Endpoints
type GetEndpointResponse struct {
	Endpoints *EndpointsObj
	RequestId string
	Success   bool
}

// EndpointsObj wrapper Endpoint array
type EndpointsObj struct {
	Endpoint []EndpointObj
}

// EndpointObj wrapper endpoint
type EndpointObj struct {
	Protocols   map[string]string
	Type        string
	Namespace   string
	Id          string
	ServiceCode string
	Endpoint    string
}
