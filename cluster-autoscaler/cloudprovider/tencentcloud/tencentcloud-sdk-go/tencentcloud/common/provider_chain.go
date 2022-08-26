/*
Copyright 2016 The Kubernetes Authors.

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
	tcerr "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
)

type ProviderChain struct {
	Providers []Provider
}

// NewProviderChain returns a provider chain in your custom order
func NewProviderChain(providers []Provider) Provider {
	return &ProviderChain{
		Providers: providers,
	}
}

// DefaultProviderChain returns a default provider chain and try to get credentials in the following order:
//  1. Environment variable
//  2. Profile
//  3. CvmRole
//
// If you want to customize the search order, please use the function NewProviderChain
func DefaultProviderChain() Provider {
	return NewProviderChain([]Provider{DefaultEnvProvider(), DefaultProfileProvider(), DefaultCvmRoleProvider()})
}

func (c *ProviderChain) GetCredential() (CredentialIface, error) {
	for _, provider := range c.Providers {
		cred, err := provider.GetCredential()
		if err != nil {
			if err == envNotSet || err == fileDoseNotExist || err == noCvmRole {
				continue
			} else {
				return nil, err
			}
		}
		return cred, err
	}
	return nil, tcerr.NewTencentCloudSDKError(creErr, "no credential found in every providers", "")

}
