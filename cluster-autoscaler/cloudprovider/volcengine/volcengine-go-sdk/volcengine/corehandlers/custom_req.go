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

package corehandlers

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/volcengine/volcengine-go-sdk/volcengine/custom"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/volcengine/volcengine-go-sdk/volcengine/request"
)

var CustomerRequestHandler = request.NamedHandler{
	Name: "core.CustomerRequestHandler",
	Fn: func(r *request.Request) {
		if r.Config.ExtendHttpRequest != nil {
			r.Config.ExtendHttpRequest(r.Context(), r.HTTPRequest)
		}

		if r.Config.ExtendHttpRequestWithMeta != nil {
			r.Config.ExtendHttpRequestWithMeta(r.Context(), r.HTTPRequest, custom.RequestMetadata{
				ServiceName: r.ClientInfo.ServiceName,
				Version:     r.ClientInfo.APIVersion,
				Action:      r.Operation.Name,
				HttpMethod:  r.Operation.HTTPMethod,
				Region:      *r.Config.Region,
			})
		}
	},
}
