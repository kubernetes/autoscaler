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

package volc

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/volcengine/volc-sdk-golang/base"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/volcengine/volcengine-go-sdk/volcengine"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/volcengine/volcengine-go-sdk/volcengine/credentials"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/volcengine/volcengine-go-sdk/volcengine/request"
)

var SignRequestHandler = request.NamedHandler{
	Name: "volc.SignRequestHandler", Fn: SignSDKRequest,
}

func SignSDKRequest(req *request.Request) {

	region := req.ClientInfo.SigningRegion

	var (
		dynamicCredentials *credentials.Credentials
		dynamicRegion      *string
		c1                 base.Credentials
		err                error
	)

	if req.Config.DynamicCredentialsIncludeError != nil {
		dynamicCredentials, dynamicRegion, err = req.Config.DynamicCredentialsIncludeError(req.Context())
		if err != nil {
			req.Error = err
			return
		}
	} else if req.Config.DynamicCredentials != nil {
		dynamicCredentials, dynamicRegion = req.Config.DynamicCredentials(req.Context())
	}

	if req.Config.DynamicCredentials != nil || req.Config.DynamicCredentialsIncludeError != nil {
		if volcengine.StringValue(dynamicRegion) == "" {
			req.Error = volcengine.ErrMissingRegion
			return
		}
		region = volcengine.StringValue(dynamicRegion)
	} else if region == "" {
		region = volcengine.StringValue(req.Config.Region)
	}

	name := req.ClientInfo.SigningName
	if name == "" {
		name = req.ClientInfo.ServiceID
	}

	if dynamicCredentials == nil {
		c1, err = req.Config.Credentials.GetBase(region, name)
	} else {
		c1, err = dynamicCredentials.GetBase(region, name)
	}

	if err != nil {
		req.Error = err
		return
	}

	r := c1.Sign(req.HTTPRequest)
	req.HTTPRequest.Header = r.Header
}
