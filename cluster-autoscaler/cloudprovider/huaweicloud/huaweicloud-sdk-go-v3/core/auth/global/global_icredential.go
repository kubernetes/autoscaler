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

package global

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/auth/signer"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/request"
	"strings"
)

type Credentials struct {
	AK            string
	SK            string
	DomainId      string
	SecurityToken string
}

func (s Credentials) ProcessAuthRequest(req *request.DefaultHttpRequest) (*request.DefaultHttpRequest, error) {
	reqBuilder := req.Builder()

	if s.DomainId != "" {
		reqBuilder.AddAutoFilledPathParam("domain_id", s.DomainId)
		reqBuilder.AddHeaderParam("X-Domain-Id", s.DomainId)
	}

	if s.SecurityToken != "" {
		reqBuilder.AddHeaderParam("X-Security-Token", s.SecurityToken)
	}

	if _, ok := req.GetHeaderParams()["Content-Type"]; ok {
		if !strings.Contains(req.GetHeaderParams()["Content-Type"], "application/json") {
			reqBuilder.AddHeaderParam("X-Sdk-Content-Sha256", "UNSIGNED-PAYLOAD")
		}
	}

	r, err := reqBuilder.Build().ConvertRequest()
	if err != nil {
		return nil, err
	}
	headerParams, err := signer.Sign(r, s.AK, s.SK)
	if err != nil {
		return nil, err
	}
	for key, value := range headerParams {
		req.AddHeaderParam(key, value)
	}
	return req, nil
}

type CredentialsBuilder struct {
	Credentials Credentials
}

func NewCredentialsBuilder() *CredentialsBuilder {
	return &CredentialsBuilder{Credentials: Credentials{}}
}

func (builder *CredentialsBuilder) WithAk(ak string) *CredentialsBuilder {
	builder.Credentials.AK = ak
	return builder
}

func (builder *CredentialsBuilder) WithSk(sk string) *CredentialsBuilder {
	builder.Credentials.SK = sk
	return builder
}

func (builder *CredentialsBuilder) WithDomainId(domainId string) *CredentialsBuilder {
	builder.Credentials.DomainId = domainId
	return builder
}

func (builder *CredentialsBuilder) WithSecurityToken(token string) *CredentialsBuilder {
	builder.Credentials.SecurityToken = token
	return builder
}

func (builder *CredentialsBuilder) Build() Credentials {
	return builder.Credentials
}
