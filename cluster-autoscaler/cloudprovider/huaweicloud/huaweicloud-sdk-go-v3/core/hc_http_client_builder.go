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
	"fmt"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/auth"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/config"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/impl"
	"reflect"
	"strings"
)

type HcHttpClientBuilder struct {
	credentialsType []string
	credentials     auth.ICredential
	endpoint        string
	httpConfig      *config.HttpConfig
}

func NewHcHttpClientBuilder() *HcHttpClientBuilder {
	hcHttpClientBuilder := &HcHttpClientBuilder{
		credentialsType: []string{"basic.Credentials"},
	}
	return hcHttpClientBuilder
}

func (builder *HcHttpClientBuilder) WithCredentialsType(credentialsType string) *HcHttpClientBuilder {
	builder.credentialsType = strings.Split(credentialsType, ",")
	return builder
}

func (builder *HcHttpClientBuilder) WithEndpoint(endpoint string) *HcHttpClientBuilder {
	builder.endpoint = endpoint
	return builder
}

func (builder *HcHttpClientBuilder) WithHttpConfig(httpConfig *config.HttpConfig) *HcHttpClientBuilder {
	builder.httpConfig = httpConfig
	return builder
}

func (builder *HcHttpClientBuilder) WithCredential(iCredential auth.ICredential) *HcHttpClientBuilder {
	builder.credentials = iCredential
	return builder
}

func (builder *HcHttpClientBuilder) Build() *HcHttpClient {
	if builder.httpConfig == nil {
		builder.httpConfig = config.DefaultHttpConfig()
	}

	if builder.credentials == nil {
		builder.credentials = auth.LoadCredentialFromEnv(builder.credentialsType[0])
	}

	match := false
	givenCredentialsType := reflect.TypeOf(builder.credentials).String()
	for _, credentialsType := range builder.credentialsType {
		if credentialsType == givenCredentialsType {
			match = true
			break
		}
	}

	if !match {
		panic(fmt.Sprintf("Need credential type is %s, actually is %s", builder.credentialsType, reflect.TypeOf(builder.credentials).String()))
	}

	defaultHttpClient := impl.NewDefaultHttpClient(builder.httpConfig)

	hcHttpClient := NewHcHttpClient(defaultHttpClient).WithEndpoint(builder.endpoint).WithCredential(builder.credentials)
	return hcHttpClient
}
