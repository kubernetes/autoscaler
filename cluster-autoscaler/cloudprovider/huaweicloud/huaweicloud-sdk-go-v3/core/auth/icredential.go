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

package auth

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/auth/global"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/request"
	"os"
)

type ICredential interface {
	ProcessAuthRequest(httpRequest *request.DefaultHttpRequest) (*request.DefaultHttpRequest, error)
}

func LoadCredentialFromEnv(defaultType string) ICredential {
	ak := os.Getenv("HUAWEICLOUD_SDK_AK")
	sk := os.Getenv("HUAWEICLOUD_SDK_SK")

	if defaultType == "basic.Credentials" {
		projectId := os.Getenv("HUAWEICLOUD_SDK_PROJECT_ID")
		return basic.NewCredentialsBuilder().
			WithAk(ak).
			WithSk(sk).
			WithProjectId(projectId).
			Build()
	} else if defaultType == "global.Credentials" {
		domainId := os.Getenv("HUAWEICLOUD_SDK_DOMAIN_ID")
		return global.NewCredentialsBuilder().
			WithAk(ak).
			WithSk(sk).
			WithDomainId(domainId).
			Build()
	} else {
		return nil
	}
}
