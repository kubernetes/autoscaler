// Copyright 2020 Huawei Technologies Co.,Ltd.
//
// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package request

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type DefaultHttpRequest struct {
	endpoint             string
	path                 string
	method               string
	queryParams          map[string]string
	pathParams           map[string]string
	autoFilledPathParams map[string]string
	headerParams         map[string]string
	body                 interface{}
}

func (httpRequest *DefaultHttpRequest) fillParamsInPath() *DefaultHttpRequest {
	for key, value := range httpRequest.pathParams {
		httpRequest.path = strings.ReplaceAll(httpRequest.path, "{"+key+"}", value)
	}
	for key, value := range httpRequest.autoFilledPathParams {
		httpRequest.path = strings.ReplaceAll(httpRequest.path, "{"+key+"}", value)
	}
	return httpRequest
}

func (httpRequest *DefaultHttpRequest) Builder() *HttpRequestBuilder {
	httpRequestBuilder := HttpRequestBuilder{httpRequest: httpRequest}
	return &httpRequestBuilder
}

func (httpRequest *DefaultHttpRequest) GetEndpoint() string {
	return httpRequest.endpoint
}

func (httpRequest *DefaultHttpRequest) GetPath() string {
	return httpRequest.path
}

func (httpRequest *DefaultHttpRequest) GetMethod() string {
	return httpRequest.method
}

func (httpRequest *DefaultHttpRequest) GetQueryParams() map[string]string {
	return httpRequest.queryParams
}

func (httpRequest *DefaultHttpRequest) GetHeaderParams() map[string]string {
	return httpRequest.headerParams
}

func (httpRequest *DefaultHttpRequest) GetPathPrams() map[string]string {
	return httpRequest.pathParams
}

func (httpRequest *DefaultHttpRequest) GetBody() interface{} {
	return httpRequest.body
}

func (httpRequest *DefaultHttpRequest) GetBodyToBytes() (*bytes.Buffer, error) {
	buf := &bytes.Buffer{}
	if httpRequest.body != nil {
		err := json.NewEncoder(buf).Encode(httpRequest.body)
		if err != nil {
			return nil, err
		}
	}
	return buf, nil
}

func (httpRequest *DefaultHttpRequest) AddQueryParam(key string, value string) {
	httpRequest.queryParams[key] = value
}

func (httpRequest *DefaultHttpRequest) AddPathParam(key string, value string) {
	httpRequest.pathParams[key] = value
}

func (httpRequest *DefaultHttpRequest) AddHeaderParam(key string, value string) {
	httpRequest.headerParams[key] = value
}

func (httpRequest *DefaultHttpRequest) ConvertRequest() (*http.Request, error) {
	buf, err := httpRequest.GetBodyToBytes()
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(httpRequest.GetMethod(), httpRequest.GetEndpoint(), buf)
	if err != nil {
		return nil, err
	}
	httpRequest.fillPath(req)
	httpRequest.fillQueryParams(req)
	httpRequest.fillHeaderParams(req)
	return req, nil
}

func (httpRequest *DefaultHttpRequest) fillHeaderParams(req *http.Request) {
	if len(httpRequest.GetHeaderParams()) == 0 {
		return
	}
	for key, value := range httpRequest.GetHeaderParams() {
		req.Header.Add(key, value)
	}
}

func (httpRequest *DefaultHttpRequest) fillQueryParams(req *http.Request) {
	if len(httpRequest.GetQueryParams()) == 0 {
		return
	}
	q := req.URL.Query()
	for key, value := range httpRequest.GetQueryParams() {
		if strings.HasPrefix(value, "[") {
			var valueList []interface{}
			err := json.Unmarshal([]byte(value), &valueList)
			if err == nil {
				for item := range valueList {
					q.Add(key, fmt.Sprintf("%v", valueList[item]))
				}
			} else {
				q.Add(key, value)
			}
		} else {
			q.Add(key, value)
		}
	}
	req.URL.RawQuery = q.Encode()
}

func (httpRequest *DefaultHttpRequest) fillPath(req *http.Request) {
	if "" != httpRequest.GetPath() {
		req.URL.Path = httpRequest.GetPath()
	}
}

type HttpRequestBuilder struct {
	httpRequest *DefaultHttpRequest
}

func NewHttpRequestBuilder() *HttpRequestBuilder {
	httpRequest := &DefaultHttpRequest{
		queryParams:          make(map[string]string),
		headerParams:         make(map[string]string),
		pathParams:           make(map[string]string),
		autoFilledPathParams: make(map[string]string),
	}
	httpRequestBuilder := &HttpRequestBuilder{
		httpRequest: httpRequest,
	}
	return httpRequestBuilder
}

func (builder *HttpRequestBuilder) WithEndpoint(endpoint string) *HttpRequestBuilder {
	builder.httpRequest.endpoint = endpoint
	return builder
}

func (builder *HttpRequestBuilder) WithPath(path string) *HttpRequestBuilder {
	builder.httpRequest.path = path
	return builder
}

func (builder *HttpRequestBuilder) WithMethod(method string) *HttpRequestBuilder {
	builder.httpRequest.method = method
	return builder
}

func (builder *HttpRequestBuilder) AddQueryParam(key string, value string) *HttpRequestBuilder {
	builder.httpRequest.queryParams[key] = value
	return builder
}

func (builder *HttpRequestBuilder) AddPathParam(key string, value string) *HttpRequestBuilder {
	builder.httpRequest.pathParams[key] = value
	return builder
}

func (builder *HttpRequestBuilder) AddAutoFilledPathParam(key string, value string) *HttpRequestBuilder {
	builder.httpRequest.autoFilledPathParams[key] = value
	return builder
}

func (builder *HttpRequestBuilder) AddHeaderParam(key string, value string) *HttpRequestBuilder {
	builder.httpRequest.headerParams[key] = value
	return builder
}

func (builder *HttpRequestBuilder) WithBody(body interface{}) *HttpRequestBuilder {
	builder.httpRequest.body = body
	return builder
}

func (builder *HttpRequestBuilder) Build() *DefaultHttpRequest {
	return builder.httpRequest.fillParamsInPath()
}
