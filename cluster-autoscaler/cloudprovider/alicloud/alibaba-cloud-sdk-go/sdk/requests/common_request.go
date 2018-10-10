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
	"bytes"
	"fmt"
	"io"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/alicloud/alibaba-cloud-sdk-go/sdk/errors"
	"strings"
)

// CommonRequest wrap base request
type CommonRequest struct {
	*baseRequest

	Version string
	ApiName string
	Product string

	// roa params
	PathPattern string
	PathParams  map[string]string

	Ontology AcsRequest
}

// NewCommonRequest returns CommonRequest
func NewCommonRequest() (request *CommonRequest) {
	request = &CommonRequest{
		baseRequest: defaultBaseRequest(),
	}
	request.Headers["x-sdk-invoke-type"] = "common"
	request.PathParams = make(map[string]string)
	return
}

// String returns CommonRequest
func (request *CommonRequest) String() string {
	request.TransToAcsRequest()
	request.BuildQueries()
	request.BuildUrl()

	resultBuilder := bytes.Buffer{}

	mapOutput := func(m map[string]string) {
		if len(m) > 0 {
			for key, value := range m {
				resultBuilder.WriteString(key + ": " + value + "\n")
			}
		}
	}

	// Request Line
	resultBuilder.WriteString("\n")
	resultBuilder.WriteString(fmt.Sprintf("%s %s %s/1.1\n", request.Method, request.GetQueries(), strings.ToUpper(request.Scheme)))

	// Headers
	resultBuilder.WriteString("Host" + ": " + request.Domain + "\n")
	mapOutput(request.Headers)

	resultBuilder.WriteString("\n")
	// Body
	if len(request.Content) > 0 {
		resultBuilder.WriteString(string(request.Content) + "\n")
	} else {
		mapOutput(request.FormParams)
	}

	return resultBuilder.String()
}

// TransToAcsRequest convert common request
func (request *CommonRequest) TransToAcsRequest() {
	if len(request.Version) == 0 {
		errors.NewClientError(errors.MissingParamErrorCode, "Common request [version] is required", nil)
	}
	if len(request.ApiName) == 0 && len(request.PathPattern) == 0 {
		errors.NewClientError(errors.MissingParamErrorCode, "At least one of [ApiName] and [PathPattern] should has a value", nil)
	}
	if len(request.Domain) == 0 && len(request.Product) == 0 {
		errors.NewClientError(errors.MissingParamErrorCode, "At least one of [Domain] and [Product] should has a value", nil)
	}

	if len(request.PathPattern) > 0 {
		roaRequest := &RoaRequest{}
		roaRequest.initWithCommonRequest(request)
		request.Ontology = roaRequest
	} else {
		rpcRequest := &RpcRequest{}
		rpcRequest.baseRequest = request.baseRequest
		rpcRequest.product = request.Product
		rpcRequest.version = request.Version
		rpcRequest.actionName = request.ApiName
		request.Ontology = rpcRequest
	}

}

// BuildUrl returns request url
func (request *CommonRequest) BuildUrl() string {
	if len(request.Port) > 0 {
		return strings.ToLower(request.Scheme) + "://" + request.Domain + ":" + request.Port + request.BuildQueries()
	}

	return strings.ToLower(request.Scheme) + "://" + request.Domain + request.BuildQueries()
}

// BuildQueries returns queries
func (request *CommonRequest) BuildQueries() string {
	return request.Ontology.BuildQueries()
}

// GetUrl returns url
func (request *CommonRequest) GetUrl() string {
	if len(request.Port) > 0 {
		return strings.ToLower(request.Scheme) + "://" + request.Domain + ":" + request.Port + request.GetQueries()
	}

	return strings.ToLower(request.Scheme) + "://" + request.Domain + request.GetQueries()
}

// GetQueries returns queries
func (request *CommonRequest) GetQueries() string {
	return request.Ontology.GetQueries()
}

// GetBodyReader returns body
func (request *CommonRequest) GetBodyReader() io.Reader {
	return request.Ontology.GetBodyReader()
}

// GetStyle returns style
func (request *CommonRequest) GetStyle() string {
	return request.Ontology.GetStyle()
}

func (request *CommonRequest) addPathParam(key, value string) {
	request.PathParams[key] = value
}
