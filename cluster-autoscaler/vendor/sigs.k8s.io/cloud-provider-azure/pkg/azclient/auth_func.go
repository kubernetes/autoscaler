/*
Copyright 2025 The Kubernetes Authors.

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
	"errors"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/msi-dataplane/pkg/dataplane"

	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/armauth"
)

var (
	ErrAuxiliaryTokenProviderNotSet = errors.New("auxiliary token provider is not set when multi-tenant is enabled for MSI")
	ErrNewKeyVaultCredentialFailed  = errors.New("create KeyVaultCredential failed")
)

// newAuthProviderWithWorkloadIdentity creates a new AuthProvider with workload identity.
// The caller is responsible for checking if workload identity is enabled.
// NOTE: it does NOT support multi-tenant scenarios.
func newAuthProviderWithWorkloadIdentity(
	aadFederatedTokenFile string,
	armConfig *ARMClientConfig,
	config *AzureAuthConfig,
	clientOptions *policy.ClientOptions,
	opts *authProviderOptions,
) (*AuthProvider, error) {
	computeCredential, err := opts.NewWorkloadIdentityCredentialFn(&azidentity.WorkloadIdentityCredentialOptions{
		ClientOptions: *clientOptions,
		ClientID:      config.GetAADClientID(),
		TenantID:      armConfig.GetTenantID(),
		TokenFilePath: aadFederatedTokenFile,
	})
	if err != nil {
		return nil, err
	}

	return &AuthProvider{
		ComputeCredential: computeCredential,
		CloudConfig:       clientOptions.Cloud,
	}, nil
}

// newAuthProviderWithManagedIdentity creates a new AuthProvider with managed identity.
// When multi-tenant is enabled, it uses the auxiliary token provider to create a network credential
// for cross-tenant resource access. If multi-tenant is enabled but the auxiliary token provider
// is not configured, it returns an error.
func newAuthProviderWithManagedIdentity(
	armConfig *ARMClientConfig,
	config *AzureAuthConfig,
	clientOptions *policy.ClientOptions,
	opts *authProviderOptions,
) (*AuthProvider, error) {
	credOptions := &azidentity.ManagedIdentityCredentialOptions{
		ClientOptions: *clientOptions,
	}
	if len(config.UserAssignedIdentityID) > 0 {
		if strings.Contains(strings.ToUpper(config.UserAssignedIdentityID), "/SUBSCRIPTIONS/") {
			credOptions.ID = azidentity.ResourceID(config.UserAssignedIdentityID)
		} else {
			credOptions.ID = azidentity.ClientID(config.UserAssignedIdentityID)
		}
	}

	computeCredential, err := opts.NewManagedIdentityCredentialFn(credOptions)
	if err != nil {
		return nil, err
	}

	rv := &AuthProvider{
		ComputeCredential: computeCredential,
		CloudConfig:       clientOptions.Cloud,
	}

	if !IsMultiTenant(armConfig) {
		return rv, nil
	}

	if config.AuxiliaryTokenProvider == nil {
		return nil, ErrAuxiliaryTokenProviderNotSet
	}

	// Use AuxiliaryTokenProvider as the network credential
	networkCredential, err := opts.NewKeyVaultCredentialFn(
		computeCredential,
		config.AuxiliaryTokenProvider.SecretResourceID(),
	)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrNewKeyVaultCredentialFailed, err)
	}

	// Additionally, we need to add the auxiliary token to the HTTP header when making requests to the compute resources
	additionalComputeClientOptions := []func(option *arm.ClientOptions){
		func(option *arm.ClientOptions) {
			option.PerRetryPolicies = append(option.PerRetryPolicies, armauth.NewAuxiliaryAuthPolicy(
				[]azcore.TokenCredential{networkCredential},
				DefaultTokenScopeFor(clientOptions.Cloud),
			))
		},
	}
	rv.NetworkCredential = networkCredential
	rv.AdditionalComputeClientOptions = additionalComputeClientOptions
	return rv, nil
}

// newAuthProviderWithServicePrincipalClientSecret creates a new AuthProvider with service principal client secret.
// When multi-tenant is enabled, it creates a compute credential with additional allowed tenants for cross-tenant access.
func newAuthProviderWithServicePrincipalClientSecret(
	armConfig *ARMClientConfig,
	config *AzureAuthConfig,
	clientOptions *policy.ClientOptions,
	opts *authProviderOptions,
) (*AuthProvider, error) {
	var (
		computeCredential azcore.TokenCredential
		networkCredential azcore.TokenCredential
	)

	if !IsMultiTenant(armConfig) {
		// Single tenant
		credOptions := &azidentity.ClientSecretCredentialOptions{
			ClientOptions: *clientOptions,
		}
		var err error
		computeCredential, err = opts.NewClientSecretCredentialFn(
			armConfig.GetTenantID(),
			config.GetAADClientID(),
			config.GetAADClientSecret(),
			credOptions,
		)
		if err != nil {
			return nil, err
		}

		return &AuthProvider{
			ComputeCredential: computeCredential,
			CloudConfig:       clientOptions.Cloud,
		}, nil
	}

	// Network credential for network resource access
	{
		credOptions := &azidentity.ClientSecretCredentialOptions{
			ClientOptions: *clientOptions,
		}
		var err error
		networkCredential, err = opts.NewClientSecretCredentialFn(
			armConfig.NetworkResourceTenantID,
			config.GetAADClientID(),
			config.GetAADClientSecret(),
			credOptions,
		)
		if err != nil {
			return nil, err
		}
	}

	// Compute credential with additional allowed tenants for cross-tenant access
	{
		credOptions := &azidentity.ClientSecretCredentialOptions{
			ClientOptions:              *clientOptions,
			AdditionallyAllowedTenants: []string{armConfig.NetworkResourceTenantID},
		}
		var err error
		computeCredential, err = opts.NewClientSecretCredentialFn(
			armConfig.GetTenantID(),
			config.GetAADClientID(),
			config.GetAADClientSecret(),
			credOptions,
		)
		if err != nil {
			return nil, err
		}
	}

	return &AuthProvider{
		ComputeCredential: computeCredential,
		NetworkCredential: networkCredential,
		CloudConfig:       clientOptions.Cloud,
	}, nil
}

// newAuthProviderWithServicePrincipalClientCertificate creates a new AuthProvider with service principal client certificate.
// When multi-tenant is enabled, it creates a compute credential with additional allowed tenants for cross-tenant access.
func newAuthProviderWithServicePrincipalClientCertificate(
	armConfig *ARMClientConfig,
	config *AzureAuthConfig,
	clientOptions *policy.ClientOptions,
	opts *authProviderOptions,
) (*AuthProvider, error) {
	certData, err := opts.ReadFileFn(config.AADClientCertPath)
	if err != nil {
		return nil, fmt.Errorf("reading the client certificate from file %s: %w", config.AADClientCertPath, err)
	}
	certificate, privateKey, err := opts.ParseCertificatesFn(certData, []byte(config.AADClientCertPassword))
	if err != nil {
		return nil, fmt.Errorf("decoding the client certificate: %w", err)
	}

	var (
		computeCredential azcore.TokenCredential
		networkCredential azcore.TokenCredential
	)

	if !IsMultiTenant(armConfig) {
		// Single tenant
		credOptions := &azidentity.ClientCertificateCredentialOptions{
			ClientOptions:        *clientOptions,
			SendCertificateChain: true,
		}
		computeCredential, err = opts.NewClientCertificateCredentialFn(
			armConfig.GetTenantID(),
			config.GetAADClientID(),
			certificate,
			privateKey,
			credOptions,
		)
		if err != nil {
			return nil, err
		}
		return &AuthProvider{
			ComputeCredential: computeCredential,
			CloudConfig:       clientOptions.Cloud,
		}, nil
	}

	// Network credential for network resource access
	{
		credOptions := &azidentity.ClientCertificateCredentialOptions{
			ClientOptions:        *clientOptions,
			SendCertificateChain: true,
		}
		networkCredential, err = opts.NewClientCertificateCredentialFn(
			armConfig.NetworkResourceTenantID,
			config.GetAADClientID(),
			certificate,
			privateKey,
			credOptions,
		)
		if err != nil {
			return nil, err
		}
	}

	// Compute credential with additional allowed tenants for cross-tenant access
	{
		credOptions := &azidentity.ClientCertificateCredentialOptions{
			ClientOptions:              *clientOptions,
			AdditionallyAllowedTenants: []string{armConfig.NetworkResourceTenantID},
			SendCertificateChain:       true,
		}
		computeCredential, err = opts.NewClientCertificateCredentialFn(
			armConfig.GetTenantID(),
			config.GetAADClientID(),
			certificate,
			privateKey,
			credOptions,
		)
		if err != nil {
			return nil, err
		}
	}

	return &AuthProvider{
		ComputeCredential: computeCredential,
		NetworkCredential: networkCredential,
		CloudConfig:       clientOptions.Cloud,
	}, nil
}

func newAuthProviderWithUserAssignedIdentity(
	config *AzureAuthConfig,
	clientOptions *policy.ClientOptions,
	opts *authProviderOptions,
) (*AuthProvider, error) {
	computeCredential, err := opts.NewUserAssignedIdentityCredentialFn(
		context.Background(),
		config.AADMSIDataPlaneIdentityPath,
		dataplane.WithClientOpts(azcore.ClientOptions{Cloud: clientOptions.Cloud}),
	)
	if err != nil {
		return nil, err
	}

	return &AuthProvider{
		ComputeCredential: computeCredential,
		CloudConfig:       clientOptions.Cloud,
	}, nil
}
