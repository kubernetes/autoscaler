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

package requests

import (
	"io"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/alicloud/alibaba-cloud-sdk-go/sdk/utils"
	"strings"
)

// RpcRequest wrap base request
type RpcRequest struct {
	*baseRequest
}

func (request *RpcRequest) init() {
	request.baseRequest = defaultBaseRequest()
	request.Method = POST
}

// GetStyle returns RPC
func (*RpcRequest) GetStyle() string {
	return RPC
}

// GetBodyReader return body
func (request *RpcRequest) GetBodyReader() io.Reader {
	if request.FormParams != nil && len(request.FormParams) > 0 {
		formString := utils.GetUrlFormedMap(request.FormParams)
		return strings.NewReader(formString)
	}
	return strings.NewReader("")
}

// BuildQueries builds queries
func (request *RpcRequest) BuildQueries() string {
	request.queries = "/?" + utils.GetUrlFormedMap(request.QueryParams)
	return request.queries
}

// GetQueries returns queries
func (request *RpcRequest) GetQueries() string {
	return request.queries
}

// BuildUrl creates url
func (request *RpcRequest) BuildUrl() string {
	return strings.ToLower(request.Scheme) + "://" + request.Domain + ":" + request.Port + request.BuildQueries()
}

// GetUrl returns url
func (request *RpcRequest) GetUrl() string {
	return strings.ToLower(request.Scheme) + "://" + request.Domain + request.GetQueries()
}

// GetVersion returns version
func (request *RpcRequest) GetVersion() string {
	return request.version
}

// GetActionName returns action name
func (request *RpcRequest) GetActionName() string {
	return request.actionName
}

func (request *RpcRequest) addPathParam(key, value string) {
	panic("not support")
}

// InitWithApiInfo init api info
func (request *RpcRequest) InitWithApiInfo(product, version, action, serviceCode, endpointType string) {
	request.init()
	request.product = product
	request.version = version
	request.actionName = action
	request.locationServiceCode = serviceCode
	request.locationEndpointType = endpointType
}
