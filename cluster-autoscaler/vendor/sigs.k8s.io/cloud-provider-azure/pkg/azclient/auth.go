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
	"fmt"
	"os"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"

	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/armauth"
)

type AuthProvider struct {
	FederatedIdentityCredential azcore.TokenCredential

	ManagedIdentityCredential   azcore.TokenCredential
	ClientSecretCredential      azcore.TokenCredential
	ClientCertificateCredential azcore.TokenCredential

	NetworkTokenCredential        azcore.TokenCredential
	NetworkClientSecretCredential azcore.TokenCredential

	MultiTenantCredential azcore.TokenCredential
}

func NewAuthProvider(armConfig *ARMClientConfig, config *AzureAuthConfig, clientOptionsMutFn ...func(option *policy.ClientOptions)) (*AuthProvider, error) {
	clientOption, err := GetAzCoreClientOption(armConfig)
	if err != nil {
		return nil, err
	}
	for _, fn := range clientOptionsMutFn {
		fn(clientOption)
	}
	// federatedIdentityCredential is used for workload identity federation
	var federatedIdentityCredential azcore.TokenCredential
	if aadFederatedTokenFile, enabled := config.GetAzureFederatedTokenFile(); enabled {
		federatedIdentityCredential, err = azidentity.NewWorkloadIdentityCredential(&azidentity.WorkloadIdentityCredentialOptions{
			ClientOptions: *clientOption,
			ClientID:      config.GetAADClientID(),
			TenantID:      armConfig.GetTenantID(),
			TokenFilePath: aadFederatedTokenFile,
		})
		if err != nil {
			return nil, err
		}
	}

	// managedIdentityCredential is used for managed identity extension
	var managedIdentityCredential azcore.TokenCredential
	if config.UseManagedIdentityExtension {
		credOptions := &azidentity.ManagedIdentityCredentialOptions{
			ClientOptions: *clientOption,
		}
		if len(config.UserAssignedIdentityID) > 0 {
			if strings.Contains(strings.ToUpper(config.UserAssignedIdentityID), "/SUBSCRIPTIONS/") {
				credOptions.ID = azidentity.ResourceID(config.UserAssignedIdentityID)
			} else {
				credOptions.ID = azidentity.ClientID(config.UserAssignedIdentityID)
			}
		}
		managedIdentityCredential, err = azidentity.NewManagedIdentityCredential(credOptions)
		if err != nil {
			return nil, err
		}
	}

	var (
		networkTokenCredential azcore.TokenCredential
	)
	if config.UseManagedIdentityExtension && config.AuxiliaryTokenProvider != nil && IsMultiTenant(armConfig) {
		networkTokenCredential, err = armauth.NewKeyVaultCredential(
			managedIdentityCredential,
			config.AuxiliaryTokenProvider.KeyVaultURL,
			config.AuxiliaryTokenProvider.SecretName,
		)
		if err != nil {
			return nil, fmt.Errorf("create KeyVaultCredential for auxiliary token provider: %w", err)
		}
	}

	// ClientSecretCredential is used for client secret
	var clientSecretCredential azcore.TokenCredential
	var networkClientSecretCredential azcore.TokenCredential
	var multiTenantCredential azcore.TokenCredential
	if len(config.GetAADClientSecret()) > 0 {
		credOptions := &azidentity.ClientSecretCredentialOptions{
			ClientOptions: *clientOption,
		}
		clientSecretCredential, err = azidentity.NewClientSecretCredential(armConfig.GetTenantID(), config.GetAADClientID(), config.GetAADClientSecret(), credOptions)
		if err != nil {
			return nil, err
		}
		if IsMultiTenant(armConfig) {
			credOptions := &azidentity.ClientSecretCredentialOptions{
				ClientOptions: *clientOption,
			}
			networkClientSecretCredential, err = azidentity.NewClientSecretCredential(armConfig.NetworkResourceTenantID, config.GetAADClientID(), config.AADClientSecret, credOptions)
			if err != nil {
				return nil, err
			}

			credOptions = &azidentity.ClientSecretCredentialOptions{
				ClientOptions:              *clientOption,
				AdditionallyAllowedTenants: []string{armConfig.NetworkResourceTenantID},
			}
			multiTenantCredential, err = azidentity.NewClientSecretCredential(armConfig.GetTenantID(), config.GetAADClientID(), config.GetAADClientSecret(), credOptions)
			if err != nil {
				return nil, err
			}

		}
	}

	// ClientCertificateCredential is used for client certificate
	var clientCertificateCredential azcore.TokenCredential
	if len(config.AADClientCertPath) > 0 {
		credOptions := &azidentity.ClientCertificateCredentialOptions{
			ClientOptions:        *clientOption,
			SendCertificateChain: true,
		}
		certData, err := os.ReadFile(config.AADClientCertPath)
		if err != nil {
			return nil, fmt.Errorf("reading the client certificate from file %s: %w", config.AADClientCertPath, err)
		}
		certificate, privateKey, err := azidentity.ParseCertificates(certData, []byte(config.AADClientCertPassword))
		if err != nil {
			return nil, fmt.Errorf("decoding the client certificate: %w", err)
		}
		clientCertificateCredential, err = azidentity.NewClientCertificateCredential(armConfig.GetTenantID(), config.GetAADClientID(), certificate, privateKey, credOptions)
		if err != nil {
			return nil, err
		}
		if IsMultiTenant(armConfig) {
			networkClientSecretCredential, err = azidentity.NewClientCertificateCredential(armConfig.NetworkResourceTenantID, config.GetAADClientID(), certificate, privateKey, credOptions)
			if err != nil {
				return nil, err
			}
			credOptions = &azidentity.ClientCertificateCredentialOptions{
				ClientOptions:              *clientOption,
				AdditionallyAllowedTenants: []string{armConfig.NetworkResourceTenantID},
			}
			multiTenantCredential, err = azidentity.NewClientCertificateCredential(armConfig.GetTenantID(), config.GetAADClientID(), certificate, privateKey, credOptions)
			if err != nil {
				return nil, err
			}
		}
	}

	return &AuthProvider{
		FederatedIdentityCredential:   federatedIdentityCredential,
		ManagedIdentityCredential:     managedIdentityCredential,
		ClientSecretCredential:        clientSecretCredential,
		ClientCertificateCredential:   clientCertificateCredential,
		NetworkClientSecretCredential: networkClientSecretCredential,
		NetworkTokenCredential:        networkTokenCredential,
		MultiTenantCredential:         multiTenantCredential,
	}, nil
}

func (factory *AuthProvider) GetAzIdentity() azcore.TokenCredential {
	switch true {
	case factory.FederatedIdentityCredential != nil:
		return factory.FederatedIdentityCredential
	case factory.ManagedIdentityCredential != nil:
		return factory.ManagedIdentityCredential
	case factory.ClientSecretCredential != nil:
		return factory.ClientSecretCredential
	case factory.ClientCertificateCredential != nil:
		return factory.ClientCertificateCredential
	default:
		return nil
	}
}

func (factory *AuthProvider) GetNetworkAzIdentity() azcore.TokenCredential {
	if factory.NetworkClientSecretCredential != nil {
		return factory.NetworkClientSecretCredential
	}
	if factory.NetworkTokenCredential != nil {
		return factory.NetworkTokenCredential
	}
	return nil
}

func (factory *AuthProvider) GetMultiTenantIdentity() azcore.TokenCredential {
	if factory.MultiTenantCredential != nil {
		return factory.MultiTenantCredential
	}
	return nil
}

func (factory *AuthProvider) IsMultiTenantModeEnabled() bool {
	return factory.MultiTenantCredential != nil
}
