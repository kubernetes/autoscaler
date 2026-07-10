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
	"crypto"
	"crypto/x509"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/msi-dataplane/pkg/dataplane"

	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/armauth"
)

type (
	NewWorkloadIdentityCredentialFn func(
		options *azidentity.WorkloadIdentityCredentialOptions,
	) (azcore.TokenCredential, error)

	NewManagedIdentityCredentialFn func(
		options *azidentity.ManagedIdentityCredentialOptions,
	) (azcore.TokenCredential, error)

	NewClientSecretCredentialFn func(
		tenantID string,
		clientID string,
		clientSecret string,
		options *azidentity.ClientSecretCredentialOptions,
	) (azcore.TokenCredential, error)

	NewClientCertificateCredentialFn func(
		tenantID string,
		clientID string,
		certs []*x509.Certificate,
		key crypto.PrivateKey,
		options *azidentity.ClientCertificateCredentialOptions,
	) (azcore.TokenCredential, error)

	NewKeyVaultCredentialFn func(
		credential azcore.TokenCredential,
		secretResourceID armauth.SecretResourceID,
	) (azcore.TokenCredential, error)

	NewUserAssignedIdentityCredentialFn func(
		ctx context.Context,
		credentialPath string,
		opts ...dataplane.Option,
	) (azcore.TokenCredential, error)
)

func DefaultNewWorkloadIdentityCredentialFn() NewWorkloadIdentityCredentialFn {
	return func(options *azidentity.WorkloadIdentityCredentialOptions) (azcore.TokenCredential, error) {
		return azidentity.NewWorkloadIdentityCredential(options)
	}
}

func DefaultNewManagedIdentityCredentialFn() NewManagedIdentityCredentialFn {
	return func(options *azidentity.ManagedIdentityCredentialOptions) (azcore.TokenCredential, error) {
		return azidentity.NewManagedIdentityCredential(options)
	}
}

func DefaultNewClientSecretCredentialFn() NewClientSecretCredentialFn {
	return func(
		tenantID string,
		clientID string,
		clientSecret string,
		options *azidentity.ClientSecretCredentialOptions,
	) (azcore.TokenCredential, error) {
		return azidentity.NewClientSecretCredential(tenantID, clientID, clientSecret, options)
	}
}
func DefaultNewClientCertificateCredentialFn() NewClientCertificateCredentialFn {
	return func(
		tenantID string,
		clientID string,
		certs []*x509.Certificate,
		key crypto.PrivateKey,
		options *azidentity.ClientCertificateCredentialOptions,
	) (azcore.TokenCredential, error) {
		return azidentity.NewClientCertificateCredential(tenantID, clientID, certs, key, options)
	}
}

func DefaultNewKeyVaultCredentialFn() NewKeyVaultCredentialFn {
	return func(
		credential azcore.TokenCredential,
		secretResourceID armauth.SecretResourceID,
	) (azcore.TokenCredential, error) {
		return armauth.NewKeyVaultCredential(credential, secretResourceID)
	}
}

func DefaultNewUserAssignedIdentityCredentialFn() NewUserAssignedIdentityCredentialFn {
	return dataplane.NewUserAssignedIdentityCredential
}

type AuthProviderOption func(option *authProviderOptions)

type authProviderOptions struct {
	ClientOptionsMutFn []func(option *policy.ClientOptions)
	// The following credential factory functions are for testing purposes only
	// and should not be modified by users of this package
	NewWorkloadIdentityCredentialFn     NewWorkloadIdentityCredentialFn
	NewManagedIdentityCredentialFn      NewManagedIdentityCredentialFn
	NewClientSecretCredentialFn         NewClientSecretCredentialFn
	NewClientCertificateCredentialFn    NewClientCertificateCredentialFn
	NewKeyVaultCredentialFn             NewKeyVaultCredentialFn
	NewUserAssignedIdentityCredentialFn NewUserAssignedIdentityCredentialFn
	ReadFileFn                          func(name string) ([]byte, error)
	ParseCertificatesFn                 func(certData []byte, password []byte) ([]*x509.Certificate, crypto.PrivateKey, error)
}

func defaultAuthProviderOptions() *authProviderOptions {
	return &authProviderOptions{
		ClientOptionsMutFn:                  []func(option *policy.ClientOptions){},
		NewWorkloadIdentityCredentialFn:     DefaultNewWorkloadIdentityCredentialFn(),
		NewManagedIdentityCredentialFn:      DefaultNewManagedIdentityCredentialFn(),
		NewClientSecretCredentialFn:         DefaultNewClientSecretCredentialFn(),
		NewClientCertificateCredentialFn:    DefaultNewClientCertificateCredentialFn(),
		NewKeyVaultCredentialFn:             DefaultNewKeyVaultCredentialFn(),
		NewUserAssignedIdentityCredentialFn: DefaultNewUserAssignedIdentityCredentialFn(),
		ReadFileFn:                          os.ReadFile,
		ParseCertificatesFn:                 azidentity.ParseCertificates,
	}
}

func WithClientOptionsMutFn(fn func(option *policy.ClientOptions)) AuthProviderOption {
	return func(option *authProviderOptions) {
		option.ClientOptionsMutFn = append(option.ClientOptionsMutFn, fn)
	}
}
