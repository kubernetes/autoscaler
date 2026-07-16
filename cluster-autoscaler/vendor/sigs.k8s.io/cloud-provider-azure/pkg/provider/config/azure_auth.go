/*
Copyright 2020 The Kubernetes Authors.

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

package config

import (
	"fmt"
	"strings"

	"sigs.k8s.io/cloud-provider-azure/pkg/azclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/policy/ratelimit"
)

var (
	// ErrorNoAuth indicates that no credentials are provided.
	ErrorNoAuth = fmt.Errorf("no credentials provided for Azure cloud provider")
)

// AzureClientConfig holds azure client related part of cloud config
type AzureClientConfig struct {
	azclient.ARMClientConfig               `json:",inline" yaml:",inline"`
	azclient.AzureAuthConfig               `json:",inline" yaml:",inline"`
	ratelimit.CloudProviderRateLimitConfig `json:",inline" yaml:",inline"`
	CloudProviderCacheConfig               `json:",inline" yaml:",inline"`
	// Backoff retry limit
	CloudProviderBackoffRetries int `json:"cloudProviderBackoffRetries,omitempty" yaml:"cloudProviderBackoffRetries,omitempty"`
	// Backoff duration
	CloudProviderBackoffDuration int `json:"cloudProviderBackoffDuration,omitempty" yaml:"cloudProviderBackoffDuration,omitempty"`

	// The ID of the Azure Subscription that the cluster is deployed in
	SubscriptionID string `json:"subscriptionId,omitempty" yaml:"subscriptionId,omitempty"`
	// IdentitySystem indicates the identity provider. Relevant only to hybrid clouds (Azure Stack).
	// Allowed values are 'azure_ad' (default), 'adfs'.
	IdentitySystem string `json:"identitySystem,omitempty" yaml:"identitySystem,omitempty"`

	// The ID of the Azure Subscription that the network resources are deployed in
	NetworkResourceSubscriptionID string `json:"networkResourceSubscriptionID,omitempty" yaml:"networkResourceSubscriptionID,omitempty"`
}

// UsesNetworkResourceInDifferentSubscription determines whether the AzureAuthConfig indicates to use network resources
// in different Subscription than those for the cluster. Return true when NetworkResourceSubscriptionID is specified
// and not equal to one defined in global configs
func (config *AzureClientConfig) UsesNetworkResourceInDifferentSubscription() bool {
	return len(config.NetworkResourceSubscriptionID) > 0 && !strings.EqualFold(config.NetworkResourceSubscriptionID, config.SubscriptionID)
}
