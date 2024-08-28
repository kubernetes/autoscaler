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

package azclient

import (
	"context"
	"net/http"
	"strings"
	"sync"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm/policy"

	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/policy/ratelimit"
	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/utils"
	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/utils/armbalancer"
)

var DefaultResourceClientTransport *http.Client
var once sync.Once

func init() {
	once.Do(func() {
		DefaultResourceClientTransport = &http.Client{
			Transport: armbalancer.New(context.Background(), armbalancer.Options{
				Transport: utils.DefaultTransport,
				PoolSize:  100,
			}),
		}
	})
}

type ClientFactoryConfig struct {
	ratelimit.CloudProviderRateLimitConfig

	// Enable exponential backoff to manage resource request retries
	CloudProviderBackoff bool `json:"cloudProviderBackoff,omitempty" yaml:"cloudProviderBackoff,omitempty"`

	// The ID of the Azure Subscription that the cluster is deployed in
	SubscriptionID string `json:"subscriptionId,omitempty" yaml:"subscriptionId,omitempty"`
}

func GetDefaultResourceClientOption(armConfig *ARMClientConfig, factoryConfig *ClientFactoryConfig) (*policy.ClientOptions, error) {
	armClientOption := policy.ClientOptions{}
	options, err := GetAzCoreClientOption(armConfig)
	if err != nil {
		return nil, err
	}
	armClientOption.ClientOptions = *options

	if armConfig != nil && armConfig.NetworkResourceTenantID != "" && !strings.EqualFold(armConfig.NetworkResourceTenantID, armConfig.GetTenantID()) {
		armClientOption.AuxiliaryTenants = []string{armConfig.NetworkResourceTenantID}
	}

	if factoryConfig != nil {
		//Set retry
		if !factoryConfig.CloudProviderBackoff {
			options.Retry.MaxRetries = 0
		}
	}
	armClientOption.ClientOptions.Transport = DefaultResourceClientTransport
	return &armClientOption, err
}
