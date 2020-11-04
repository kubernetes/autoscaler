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
	"errors"
	"fmt"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/auth"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/converter"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/def"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/impl"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/request"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/response"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/sdkerr"
	jsoniter "github.com/json-iterator/go"
	"io/ioutil"
	"reflect"
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

func (hc *HcHttpClient) Sync(req interface{}, reqDef *def.HttpRequestDef) (interface{}, error) {
	httpRequest, err := hc.buildRequest(req, reqDef)
	if err != nil {
		return nil, err
	}

	resp, err := hc.httpClient.SyncInvokeHttp(httpRequest)
	if err != nil {
		return nil, err
	}

	return hc.extractResponse(resp, reqDef)
}

func (hc *HcHttpClient) buildRequest(req interface{}, reqDef *def.HttpRequestDef) (*request.DefaultHttpRequest, error) {
	builder := request.NewHttpRequestBuilder().
		WithEndpoint(hc.endpoint).
		WithMethod(reqDef.Method).
		WithPath(reqDef.Path)

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
	attrMaps := hc.GetFieldJsonTags(req)

	for _, fieldDef := range reqDef.RequestFields {
		value, err := hc.GetFieldValueByName(fieldDef.Name, attrMaps, req)
		if err != nil {
			return nil, err
		}
		if !value.IsValid() {
			continue
		}
		switch fieldDef.LocationType {
		case def.Header:
			builder.AddHeaderParam(fieldDef.JsonTag, fmt.Sprintf("%v", value))
		case def.Path:
			builder.AddPathParam(fieldDef.JsonTag, fmt.Sprintf("%v", value))
		case def.Query:
			builder.AddQueryParam(fieldDef.JsonTag, value)
		case def.Body:
			builder.WithBody(value.Interface())
		}
	}
	return builder, nil
}

func (hc *HcHttpClient) GetFieldJsonTags(structName interface{}) map[string]string {
	attrMaps := make(map[string]string)
	t := reflect.TypeOf(structName)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	fieldNum := t.NumField()
	for i := 0; i < fieldNum; i++ {
		jsonTag := t.Field(i).Tag.Get("json")
		if jsonTag != "" {
			attrMaps[t.Field(i).Name] = jsonTag
		}
	}
	return attrMaps
}

func (hc *HcHttpClient) GetFieldValueByName(name string, jsonTag map[string]string, structName interface{}) (reflect.Value, error) {
	v := reflect.ValueOf(structName)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	value := v.FieldByName(name)
	if value.Kind() == reflect.Ptr {
		if value.IsNil() {
			if strings.Contains(jsonTag[name], "omitempty") {
				return reflect.ValueOf(nil), nil
			}
			return reflect.ValueOf(nil), errors.New("request field " + name + " read null value")
		}
		return value.Elem(), nil
	}

	if value.Kind() == reflect.Struct {
		v, err := jsoniter.Marshal(value.Interface())
		if strings.HasPrefix(string(v), "\"") {
			return reflect.ValueOf(strings.Trim(string(v), "\"")), err
		} else {
			return reflect.ValueOf(string(v)), err
		}
	}

	return value, nil
}

func (hc *HcHttpClient) extractResponse(resp *response.DefaultHttpResponse, reqDef *def.HttpRequestDef) (interface{}, error) {
	if resp.GetStatusCode() >= 400 {
		return nil, sdkerr.NewServiceResponseError(resp.Response)
	}

	err := hc.deserializeResponse(resp, reqDef)
	if err != nil {
		return nil, err
	}

	return reqDef.Response, nil
}

func (hc *HcHttpClient) deserializeResponse(resp *response.DefaultHttpResponse, reqDef *def.HttpRequestDef) error {
	data, err := ioutil.ReadAll(resp.Response.Body)
	if err != nil {
		if closeErr := resp.Response.Body.Close(); closeErr != nil {
			return err
		}
		return err
	}
	if err := resp.Response.Body.Close(); err != nil {
		return err
	} else {
		resp.Response.Body = ioutil.NopCloser(bytes.NewBuffer(data))
	}

	hasBody := false
	for _, item := range reqDef.ResponseFields {
		if item.LocationType == def.Header {
			err := hc.deserializeResponseHeaders(resp, reqDef, item)
			if err != nil {
				return sdkerr.ServiceResponseError{
					StatusCode:   resp.GetStatusCode(),
					RequestId:    resp.GetHeader("X-Request-Id"),
					ErrorMessage: err.Error(),
				}
			}
		}

		if item.LocationType == def.Body {
			hasBody = true

			dataOfListOrString := "{ \"body\" : " + string(data) + "}"
			err = jsoniter.Unmarshal([]byte(dataOfListOrString), &reqDef.Response)
			if err != nil {
				return sdkerr.ServiceResponseError{
					StatusCode:   resp.GetStatusCode(),
					RequestId:    resp.GetHeader("X-Request-Id"),
					ErrorMessage: err.Error(),
				}
			}
		}
	}

	if len(data) != 0 && !hasBody {
		err = jsoniter.Unmarshal(data, &reqDef.Response)
		if err != nil {
			return sdkerr.ServiceResponseError{
				StatusCode:   resp.GetStatusCode(),
				RequestId:    resp.GetHeader("X-Request-Id"),
				ErrorMessage: err.Error(),
			}
		}
	}

	return nil
}

func (hc *HcHttpClient) deserializeResponseHeaders(resp *response.DefaultHttpResponse, reqDef *def.HttpRequestDef, item *def.FieldDef) error {
	isPtr, fieldKind := GetFieldInfo(reqDef, item)
	v := reflect.ValueOf(reqDef.Response)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	fieldValue := v.FieldByName(item.Name)
	headerValue := resp.GetHeader(item.JsonTag)

	sdkConverter := converter.StringConverterFactory(fieldKind)
	if sdkConverter == nil {
		return errors.New("failed to convert " + item.JsonTag)
	}

	err := sdkConverter.CovertStringToPrimitiveTypeAndSetField(fieldValue, headerValue, isPtr)
	if err != nil {
		return err
	}

	return nil
}

func GetFieldInfo(reqDef *def.HttpRequestDef, item *def.FieldDef) (bool, string) {
	var fieldKind string
	var isPtr = false
	t := reflect.TypeOf(reqDef.Response)
	if t.Kind() == reflect.Ptr {
		isPtr = true
		t = t.Elem()
	}
	field, _ := t.FieldByName(item.Name)
	if field.Type.Kind() == reflect.Ptr {
		fieldKind = field.Type.Elem().Kind().String()
	} else {
		fieldKind = field.Type.Kind().String()
	}
	return isPtr, fieldKind
}
