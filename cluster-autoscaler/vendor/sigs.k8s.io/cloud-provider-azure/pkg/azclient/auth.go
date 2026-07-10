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
	"errors"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
)

var (
	ErrNoValidAuthMethodFound = errors.New("no valid authentication method found")
)

type AuthProvider struct {
	ComputeCredential              azcore.TokenCredential
	AdditionalComputeClientOptions []func(option *arm.ClientOptions)
	NetworkCredential              azcore.TokenCredential
	CloudConfig                    cloud.Configuration
}

func NewAuthProvider(
	armConfig *ARMClientConfig,
	config *AzureAuthConfig,
	options ...AuthProviderOption,
) (*AuthProvider, error) {
	opts := defaultAuthProviderOptions()
	for _, opt := range options {
		opt(opts)
	}

	clientOption, _, err := GetAzCoreClientOption(armConfig)
	if err != nil {
		return nil, err
	}
	for _, fn := range opts.ClientOptionsMutFn {
		fn(clientOption)
	}

	aadFederatedTokenFile, federatedTokenEnabled := config.GetAzureFederatedTokenFile()
	switch {
	case federatedTokenEnabled:
		return newAuthProviderWithWorkloadIdentity(aadFederatedTokenFile, armConfig, config, clientOption, opts)
	case config.UseManagedIdentityExtension:
		return newAuthProviderWithManagedIdentity(armConfig, config, clientOption, opts)
	case len(config.GetAADClientSecret()) > 0:
		return newAuthProviderWithServicePrincipalClientSecret(armConfig, config, clientOption, opts)
	case len(config.AADClientCertPath) > 0:
		return newAuthProviderWithServicePrincipalClientCertificate(armConfig, config, clientOption, opts)
	case len(config.AADMSIDataPlaneIdentityPath) > 0:
		return newAuthProviderWithUserAssignedIdentity(config, clientOption, opts)
	default:
		return &AuthProvider{
			CloudConfig: clientOption.Cloud,
		}, nil
	}
}

func (factory *AuthProvider) GetAzIdentity() azcore.TokenCredential {
	return factory.ComputeCredential
}

func (factory *AuthProvider) GetNetworkAzIdentity() azcore.TokenCredential {
	if factory.NetworkCredential != nil {
		return factory.NetworkCredential
	}
	return factory.ComputeCredential
}

func (factory *AuthProvider) DefaultTokenScope() string {
	return DefaultTokenScopeFor(factory.CloudConfig)
}

func DefaultTokenScopeFor(cloudCfg cloud.Configuration) string {
	audience := cloudCfg.Services[cloud.ResourceManager].Audience
	return fmt.Sprintf("%s/.default", strings.TrimRight(audience, "/"))
}
