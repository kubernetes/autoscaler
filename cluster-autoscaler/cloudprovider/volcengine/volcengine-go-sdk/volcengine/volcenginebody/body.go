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

package volcenginebody

// Copy from https://github.com/aws/aws-sdk-go
// May have been modified by Beijing Volcanoengine Technology Ltd.

import (
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"
	"strings"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/volcengine/volcengine-go-sdk/private/protocol/query/queryutil"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/volcengine/volcengine-go-sdk/volcengine"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/volcengine/volcengine-go-sdk/volcengine/custom"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/volcengine/volcengine-go-sdk/volcengine/request"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/volcengine/volcengine-go-sdk/volcengine/volcengineerr"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/volcengine/volcengine-go-sdk/volcengine/volcengineutil"
)

func BodyParam(body *url.Values, r *request.Request) {
	var (
		isForm bool
	)
	contentType := r.HTTPRequest.Header.Get("Content-Type")
	newBody := body
	if len(contentType) > 0 && strings.Contains(strings.ToLower(contentType), "x-www-form-urlencoded") {
		isForm = true
		newBody = &url.Values{}
	}

	if !isForm && len(contentType) > 0 {
		r.Error = volcengineerr.New("SerializationError", "not support such content-type", nil)
		return
	}

	if reflect.TypeOf(r.Params) == reflect.TypeOf(&map[string]interface{}{}) {
		m := *(r.Params).(*map[string]interface{})
		for k, v := range m {
			if reflect.TypeOf(v).String() == "string" {
				newBody.Add(k, v.(string))
			} else {
				newBody.Add(k, fmt.Sprintf("%v", v))
			}
		}
	} else if err := queryutil.Parse(*newBody, r.Params, false); err != nil {
		r.Error = volcengineerr.New("SerializationError", "failed encoding Query request", err)
		return
	}

	//extra process
	if r.Config.ExtraHttpParameters != nil {
		extra := r.Config.ExtraHttpParameters(r.Context())
		if extra != nil {
			for k, value := range extra {
				newBody.Add(k, value)
			}
		}
	}
	if r.Config.ExtraHttpParametersWithMeta != nil {
		extra := r.Config.ExtraHttpParametersWithMeta(r.Context(), custom.RequestMetadata{
			ServiceName: r.ClientInfo.ServiceName,
			Version:     r.ClientInfo.APIVersion,
			Action:      r.Operation.Name,
			HttpMethod:  r.Operation.HTTPMethod,
			Region:      *r.Config.Region,
			Request:     r.HTTPRequest,
			RawQuery:    body,
		})
		if extra != nil {
			for k, value := range extra {
				newBody.Add(k, value)
			}
		}
	}

	if isForm {
		r.HTTPRequest.URL.RawQuery = body.Encode()
		r.HTTPRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=utf-8")
		r.SetBufferBody([]byte(newBody.Encode()))
		return
	}

	r.Input = volcengineutil.ParameterToMap(body.Encode(), r.Config.LogSensitives,
		r.Config.LogLevel.Matches(volcengine.LogInfoWithInputAndOutput) || r.Config.LogLevel.Matches(volcengine.LogDebugWithInputAndOutput))

	r.HTTPRequest.URL.RawQuery = newBody.Encode()
}

func BodyJson(body *url.Values, r *request.Request) {
	method := strings.ToUpper(r.HTTPRequest.Method)
	if v := r.HTTPRequest.Header.Get("Content-Type"); len(v) == 0 {
		r.HTTPRequest.Header.Set("Content-Type", "application/json; charset=utf-8")
	}

	if v := r.HTTPRequest.Header.Get("Content-Type"); !strings.Contains(strings.ToLower(v), "application/json") || method == "GET" {
		return
	}

	input := make(map[string]interface{})
	b, _ := json.Marshal(r.Params)
	_ = json.Unmarshal(b, &input)
	if r.Config.ExtraHttpJsonBody != nil {
		r.Config.ExtraHttpJsonBody(r.Context(), &input, custom.RequestMetadata{
			ServiceName: r.ClientInfo.ServiceName,
			Version:     r.ClientInfo.APIVersion,
			Action:      r.Operation.Name,
			HttpMethod:  r.Operation.HTTPMethod,
			Region:      *r.Config.Region,
			Request:     r.HTTPRequest,
			RawQuery:    body,
		})
		b, _ = json.Marshal(input)
	}
	r.SetStringBody(string(b))

	r.HTTPRequest.URL.RawQuery = body.Encode()
	r.IsJsonBody = true

	r.Input = volcengineutil.BodyToMap(input, r.Config.LogSensitives,
		r.Config.LogLevel.Matches(volcengine.LogInfoWithInputAndOutput) || r.Config.LogLevel.Matches(volcengine.LogDebugWithInputAndOutput))
	r.Params = nil
	r.HTTPRequest.Header.Set("Accept", "application/json")
}
