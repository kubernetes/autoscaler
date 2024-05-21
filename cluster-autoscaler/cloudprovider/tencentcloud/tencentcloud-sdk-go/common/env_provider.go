/*
Copyright 2021 The Kubernetes Authors.

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

package common

import (
	"os"

	tcerr "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/tencentcloud/tencentcloud-sdk-go/common/errors"
)

type EnvProvider struct {
	secretIdENV  string
	secretKeyENV string
}

// DefaultEnvProvider return a default provider
// The default environment variable name are TENCENTCLOUD_SECRET_ID and TENCENTCLOUD_SECRET_KEY
func DefaultEnvProvider() *EnvProvider {
	return &EnvProvider{
		secretIdENV:  "TENCENTCLOUD_SECRET_ID",
		secretKeyENV: "TENCENTCLOUD_SECRET_KEY",
	}
}

// NewEnvProvider uses the name of the environment variable you specified to get the credentials
func NewEnvProvider(secretIdEnvName, secretKeyEnvName string) *EnvProvider {
	return &EnvProvider{
		secretIdENV:  secretIdEnvName,
		secretKeyENV: secretKeyEnvName,
	}
}

func (p *EnvProvider) GetCredential() (CredentialIface, error) {
	secretId, ok1 := os.LookupEnv(p.secretIdENV)
	secretKey, ok2 := os.LookupEnv(p.secretKeyENV)
	if !ok1 || !ok2 {
		return nil, envNotSet
	}
	if secretId == "" || secretKey == "" {
		return nil, tcerr.NewTencentCloudSDKError(creErr, "Environmental variable ("+p.secretIdENV+" or "+p.secretKeyENV+") is empty", "")
	}
	return NewCredential(secretId, secretKey), nil
}
