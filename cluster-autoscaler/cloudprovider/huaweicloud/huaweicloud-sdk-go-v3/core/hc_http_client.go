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

package core

import (
	"bytes"
	"encoding/json"
	jsoniter "github.com/json-iterator/go"
	"io/ioutil"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/auth"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/def"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/impl"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/request"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/response"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/sdkerr"
	"strings"
)

type HcHttpClient struct {
	endpoint   string
	credential auth.ICredential
	httpClient *impl.DefaultHttpClient
}

func NewHcHttpClient(httpClient *impl.DefaultHttpClient) *HcHttpClient {
	return &HcHttpClient{httpClient: httpClient}
}

func (hc *HcHttpClient) WithEndpoint(endpoint string) *HcHttpClient {
	hc.endpoint = endpoint
	return hc
}

func (hc *HcHttpClient) WithCredential(credential auth.ICredential) *HcHttpClient {
	hc.credential = credential
	return hc
}

func (hc *HcHttpClient) Sync(req interface{}, reqDef *def.HttpRequestDef, responseDef *def.HttpResponseDef) (*response.DefaultHttpResponse, error) {
	httpRequest, err := hc.buildRequest(req, reqDef)
	if err != nil {
		return nil, err
	}

	resp, err := hc.httpClient.SyncInvokeHttp(httpRequest)
	if err != nil {
		return nil, err
	}

	return hc.extractResponse(resp, responseDef)
}

func (hc *HcHttpClient) buildRequest(req interface{}, reqDef *def.HttpRequestDef) (*request.DefaultHttpRequest, error) {
	builder := request.NewHttpRequestBuilder().
		WithEndpoint(hc.endpoint).
		WithMethod(reqDef.Method).
		WithPath(reqDef.Path).
		WithBody(reqDef.BodyJson)

	if reqDef.ContentType != "" {
		builder.AddHeaderParam("Content-Type", reqDef.ContentType)
	}
	builder.AddHeaderParam("User-Agent", "huaweicloud-usdk-go/3.0")

	builder, err := hc.fillParamsFromReq(req, reqDef, builder)
	if err != nil {
		return nil, err
	}

	httpRequest, err := hc.credential.ProcessAuthRequest(builder.Build())
	if err != nil {
		return nil, err
	}
	return httpRequest, err
}

func (hc *HcHttpClient) fillParamsFromReq(req interface{}, reqDef *def.HttpRequestDef, builder *request.HttpRequestBuilder) (*request.HttpRequestBuilder, error) {
	toBytes, err := hc.convertToBytes(req)
	if err != nil {
		return nil, err
	}

	for _, fieldDef := range reqDef.RequestFields {
		value := jsoniter.Get(toBytes.Bytes(), fieldDef.Name).ToString()
		if value == "" {
			continue
		}
		switch fieldDef.LocationType {
		case def.Header:
			builder.AddHeaderParam(fieldDef.Name, value)
		case def.Path:
			builder.AddPathParam(fieldDef.Name, value)
		case def.Query:
			builder.AddQueryParam(fieldDef.Name, value)
		}
	}
	return builder, err
}

func (hc *HcHttpClient) convertToBytes(req interface{}) (*bytes.Buffer, error) {
	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).Encode(req)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

func (hc *HcHttpClient) extractResponse(resp *response.DefaultHttpResponse, responseDef *def.HttpResponseDef) (*response.DefaultHttpResponse, error) {
	if resp.GetStatusCode() >= 400 {
		return resp, sdkerr.NewServiceResponseError(resp.Response)
	}

	data, err := ioutil.ReadAll(resp.Response.Body)
	if err != nil {
		if closeErr := resp.Response.Body.Close(); closeErr != nil {
			return nil, err
		}
		return nil, err
	}

	if err := resp.Response.Body.Close(); err != nil {
		return nil, err
	} else {
		resp.Response.Body = ioutil.NopCloser(bytes.NewBuffer(data))
	}

	if len(data) == 0 {
		return resp, nil
	}

	err = jsoniter.Unmarshal(data, responseDef.BodyJson)
	if err != nil {
		if strings.HasPrefix(string(data), "{") {
			return nil, sdkerr.ServiceResponseError{
				StatusCode:   resp.GetStatusCode(),
				RequestId:    resp.GetHeader("X-Request-Id"),
				ErrorMessage: err.Error(),
			}
		}

		dataOfListOrString := "{ \"body\" : " + string(data) + "}"
		err = jsoniter.Unmarshal([]byte(dataOfListOrString), responseDef.BodyJson)
		if err != nil {
			return nil, sdkerr.ServiceResponseError{
				StatusCode:   resp.GetStatusCode(),
				RequestId:    resp.GetHeader("X-Request-Id"),
				ErrorMessage: err.Error(),
			}
		}
	}

	resp.BodyJson = responseDef.BodyJson
	return resp, nil
}
