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

package custom

import (
	"context"
	"net/http"
	"net/url"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/volcengine/volcengine-go-sdk/volcengine/client/metadata"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/volcengine/volcengine-go-sdk/volcengine/response"
)

type SdkInterceptor struct {
	Before BeforeCall
	After  AfterCall
}

type RequestInfo struct {
	Context    context.Context
	Request    *http.Request
	Response   *http.Response
	Name       string
	Method     string
	ClientInfo metadata.ClientInfo
	URI        string
	Header     http.Header
	URL        *url.URL
	Input      interface{}
	Output     interface{}
	Metadata   response.ResponseMetadata
	Error      error
}

type BeforeCall func(RequestInfo) interface{}

type AfterCall func(RequestInfo, interface{})
