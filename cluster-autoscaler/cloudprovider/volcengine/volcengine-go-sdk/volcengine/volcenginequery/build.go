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

package volcenginequery

// Copy from https://github.com/aws/aws-sdk-go
// May have been modified by Beijing Volcanoengine Technology Ltd.

import (
	"net/url"
	"strings"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/volcengine/volcengine-go-sdk/volcengine/request"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/volcengine/volcengine-go-sdk/volcengine/volcenginebody"
)

// BuildHandler is a named request handler for building volcenginequery protocol requests
var BuildHandler = request.NamedHandler{Name: "volcenginesdk.volcenginequery.Build", Fn: Build}

// Build builds a request for a Volcengine Query service.
func Build(r *request.Request) {
	body := url.Values{
		"Action":  {r.Operation.Name},
		"Version": {r.ClientInfo.APIVersion},
	}
	//r.HTTPRequest.Header.Add("Accept", "application/json")
	//method := strings.ToUpper(r.HTTPRequest.Method)

	if r.Config.ExtraUserAgent != nil && *r.Config.ExtraUserAgent != "" {
		if strings.HasPrefix(*r.Config.ExtraUserAgent, "/") {
			request.AddToUserAgent(r, *r.Config.ExtraUserAgent)
		} else {
			request.AddToUserAgent(r, "/"+*r.Config.ExtraUserAgent)
		}

	}
	r.HTTPRequest.Host = r.HTTPRequest.URL.Host
	v := r.HTTPRequest.Header.Get("Content-Type")
	if (strings.ToUpper(r.HTTPRequest.Method) == "PUT" ||
		strings.ToUpper(r.HTTPRequest.Method) == "POST" ||
		strings.ToUpper(r.HTTPRequest.Method) == "DELETE" ||
		strings.ToUpper(r.HTTPRequest.Method) == "PATCH") &&
		strings.Contains(strings.ToLower(v), "application/json") {
		r.HTTPRequest.Header.Set("Content-Type", "application/json; charset=utf-8")
		volcenginebody.BodyJson(&body, r)
	} else {
		volcenginebody.BodyParam(&body, r)
	}
}
