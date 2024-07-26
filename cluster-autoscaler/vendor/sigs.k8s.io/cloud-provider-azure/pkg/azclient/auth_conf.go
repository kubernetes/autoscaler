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
	"os"

	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/utils"
)

// AzureAuthConfig holds auth related part of cloud config
type AzureAuthConfig struct {
	// The ClientID for an AAD application with RBAC access to talk to Azure RM APIs
	AADClientID string `json:"aadClientId,omitempty" yaml:"aadClientId,omitempty"`
	// The ClientSecret for an AAD application with RBAC access to talk to Azure RM APIs
	AADClientSecret string `json:"aadClientSecret,omitempty" yaml:"aadClientSecret,omitempty" datapolicy:"token"`
	// The path of a client certificate for an AAD application with RBAC access to talk to Azure RM APIs
	AADClientCertPath string `json:"aadClientCertPath,omitempty" yaml:"aadClientCertPath,omitempty"`
	// The password of the client certificate for an AAD application with RBAC access to talk to Azure RM APIs
	AADClientCertPassword string `json:"aadClientCertPassword,omitempty" yaml:"aadClientCertPassword,omitempty" datapolicy:"password"`
	// Use managed service identity for the virtual machine to access Azure ARM APIs
	UseManagedIdentityExtension bool `json:"useManagedIdentityExtension,omitempty" yaml:"useManagedIdentityExtension,omitempty"`
	// UserAssignedIdentityID contains the Client ID of the user assigned MSI which is assigned to the underlying VMs. If empty the user assigned identity is not used.
	// More details of the user assigned identity can be found at: https://docs.microsoft.com/en-us/azure/active-directory/managed-service-identity/overview
	// For the user assigned identity specified here to be used, the UseManagedIdentityExtension has to be set to true.
	UserAssignedIdentityID string `json:"userAssignedIdentityID,omitempty" yaml:"userAssignedIdentityID,omitempty"`
	// The AAD federated token file
	AADFederatedTokenFile string `json:"aadFederatedTokenFile,omitempty" yaml:"aadFederatedTokenFile,omitempty"`
	// Use workload identity federation for the virtual machine to access Azure ARM APIs
	UseFederatedWorkloadIdentityExtension bool `json:"useFederatedWorkloadIdentityExtension,omitempty" yaml:"useFederatedWorkloadIdentityExtension,omitempty"`
	// Auxiliary token provider for accessing resources from network tenant
	// Require MSI to be enabled and have permission to access the KeyVault
	AuxiliaryTokenProvider *AzureAuthAuxiliaryTokenProvider `json:"auxiliaryTokenProvider,omitempty" yaml:"auxiliaryTokenProvider,omitempty"`
}

type AzureAuthAuxiliaryTokenProvider struct {
	KeyVaultURL string `json:"keyVaultURL,omitempty" yaml:"keyVaultURL,omitempty"`
	SecretName  string `json:"secretName" yaml:"secretName"`
}

func (config *AzureAuthConfig) GetAADClientID() string {
	// these environment variables are injected by workload identity webhook
	if clientID := os.Getenv(utils.AzureClientID); clientID != "" {
		return clientID
	}
	return config.AADClientID
}

func (config *AzureAuthConfig) GetAADClientSecret() string {
	// these environment variables are injected by workload identity webhook
	if clientSecret := os.Getenv(utils.AzureClientSecret); clientSecret != "" {
		return clientSecret
	}
	return config.AADClientSecret
}

func (config *AzureAuthConfig) GetAzureFederatedTokenFile() (string, bool) {
	// these environment variables are injected by workload identity webhook
	if clientCertPath := os.Getenv(utils.AzureFederatedTokenFile); clientCertPath != "" {
		return clientCertPath, true
	}
	return config.AADFederatedTokenFile, config.UseFederatedWorkloadIdentityExtension
}
