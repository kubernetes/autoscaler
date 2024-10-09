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
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"sigs.k8s.io/yaml"

	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"

	"k8s.io/klog/v2"

	"sigs.k8s.io/cloud-provider-azure/pkg/azclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/consts"
)

var (
	// ErrorNoAuth indicates that no credentials are provided.
	ErrorNoAuth = fmt.Errorf("no credentials provided for Azure cloud provider")
)

const (
	maxReadLength = 10 * 1 << 20 // 10MB
)

// AzureAuthConfig holds auth related part of cloud config
type AzureAuthConfig struct {
	azclient.ARMClientConfig `json:",inline" yaml:",inline"`
	azclient.AzureAuthConfig `json:",inline" yaml:",inline"`

	// The ID of the Azure Subscription that the cluster is deployed in
	SubscriptionID string `json:"subscriptionId,omitempty" yaml:"subscriptionId,omitempty"`
	// IdentitySystem indicates the identity provider. Relevant only to hybrid clouds (Azure Stack).
	// Allowed values are 'azure_ad' (default), 'adfs'.
	IdentitySystem string `json:"identitySystem,omitempty" yaml:"identitySystem,omitempty"`

	// The ID of the Azure Subscription that the network resources are deployed in
	NetworkResourceSubscriptionID string `json:"networkResourceSubscriptionID,omitempty" yaml:"networkResourceSubscriptionID,omitempty"`
}

// GetServicePrincipalToken creates a new service principal token based on the configuration.
//
// By default, the cluster and its network resources are deployed in the same AAD Tenant and Subscription,
// and all azure clients use this method to fetch Service Principal Token.
//
// If NetworkResourceTenantID and NetworkResourceSubscriptionID are specified to have different values than TenantID and SubscriptionID, network resources are deployed in different AAD Tenant and Subscription than those for the cluster,
// than only azure clients except VM/VMSS and network resource ones use this method to fetch Token.
// For tokens for VM/VMSS and network resource ones, please check GetMultiTenantServicePrincipalToken and GetNetworkResourceServicePrincipalToken.
func GetServicePrincipalToken(config *AzureAuthConfig, env *azure.Environment, resource string) (*adal.ServicePrincipalToken, error) {
	var tenantID string
	if strings.EqualFold(config.IdentitySystem, consts.ADFSIdentitySystem) {
		tenantID = consts.ADFSIdentitySystem
	} else {
		tenantID = config.TenantID
	}

	if resource == "" {
		resource = env.ServiceManagementEndpoint
	}

	if config.UseFederatedWorkloadIdentityExtension {
		klog.V(2).Infoln("azure: using workload identity extension to retrieve access token")
		oauthConfig, err := adal.NewOAuthConfigWithAPIVersion(env.ActiveDirectoryEndpoint, config.TenantID, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create the OAuth config: %w", err)
		}

		jwtCallback := func() (string, error) {
			jwt, err := os.ReadFile(config.AADFederatedTokenFile)
			if err != nil {
				return "", fmt.Errorf("failed to read a file with a federated token: %w", err)
			}
			return string(jwt), nil
		}

		token, err := adal.NewServicePrincipalTokenFromFederatedTokenCallback(*oauthConfig, config.AADClientID, jwtCallback, env.ResourceManagerEndpoint)
		if err != nil {
			return nil, fmt.Errorf("failed to create a workload identity token: %w", err)
		}
		return token, nil
	}

	if config.UseManagedIdentityExtension {
		klog.V(2).Infoln("azure: using managed identity extension to retrieve access token")
		msiEndpoint, err := adal.GetMSIVMEndpoint()
		if err != nil {
			return nil, fmt.Errorf("error getting the managed service identity endpoint: %w", err)
		}
		if len(config.UserAssignedIdentityID) > 0 {
			klog.V(4).Info("azure: using User Assigned MSI ID to retrieve access token")
			resourceID, err := azure.ParseResourceID(config.UserAssignedIdentityID)
			if err == nil &&
				strings.EqualFold(resourceID.Provider, "Microsoft.ManagedIdentity") &&
				strings.EqualFold(resourceID.ResourceType, "userAssignedIdentities") {
				klog.V(4).Info("azure: User Assigned MSI ID is resource ID")
				return adal.NewServicePrincipalTokenFromMSIWithIdentityResourceID(msiEndpoint,
					resource,
					config.UserAssignedIdentityID)
			}

			klog.V(4).Info("azure: User Assigned MSI ID is client ID")
			return adal.NewServicePrincipalTokenFromMSIWithUserAssignedID(msiEndpoint,
				resource,
				config.UserAssignedIdentityID)
		}
		klog.V(4).Info("azure: using System Assigned MSI to retrieve access token")
		return adal.NewServicePrincipalTokenFromMSI(
			msiEndpoint,
			resource)
	}

	oauthConfig, err := adal.NewOAuthConfigWithAPIVersion(env.ActiveDirectoryEndpoint, tenantID, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating the OAuth config: %w", err)
	}

	if len(config.AADClientSecret) > 0 {
		klog.V(2).Infoln("azure: using client_id+client_secret to retrieve access token")
		return adal.NewServicePrincipalToken(
			*oauthConfig,
			config.AADClientID,
			config.AADClientSecret,
			resource)
	}

	if len(config.AADClientCertPath) > 0 {
		klog.V(2).Infoln("azure: using jwt client_assertion (client_cert+client_private_key) to retrieve access token")
		certData, err := os.ReadFile(config.AADClientCertPath)
		if err != nil {
			return nil, fmt.Errorf("reading the client certificate from file %s: %w", config.AADClientCertPath, err)
		}
		certificate, privateKey, err := adal.DecodePfxCertificateData(certData, config.AADClientCertPassword)
		if err != nil {
			return nil, fmt.Errorf("decoding the client certificate: %w", err)
		}
		return adal.NewServicePrincipalTokenFromCertificate(
			*oauthConfig,
			config.AADClientID,
			certificate,
			privateKey,
			resource)
	}

	return nil, ErrorNoAuth
}

// GetMultiTenantServicePrincipalToken is used when (and only when) NetworkResourceTenantID and NetworkResourceSubscriptionID are specified to have different values than TenantID and SubscriptionID.
//
// In that scenario, network resources are deployed in different AAD Tenant and Subscription than those for the cluster,
// and this method creates a new multi-tenant service principal token based on the configuration.
//
// PrimaryToken of the returned multi-tenant token is for the AAD Tenant specified by TenantID, and AuxiliaryToken of the returned multi-tenant token is for the AAD Tenant specified by NetworkResourceTenantID.
//
// Azure VM/VMSS clients use this multi-tenant token, in order to operate those VM/VMSS in AAD Tenant specified by TenantID, and meanwhile in their payload they are referencing network resources (e.g. Load Balancer, Network Security Group, etc.) in AAD Tenant specified by NetworkResourceTenantID.
func GetMultiTenantServicePrincipalToken(config *AzureAuthConfig, env *azure.Environment) (*adal.MultiTenantServicePrincipalToken, error) {
	err := config.checkConfigWhenNetworkResourceInDifferentTenant()
	if err != nil {
		return nil, fmt.Errorf("got error getting multi-tenant service principal token: %w", err)
	}

	multiTenantOAuthConfig, err := adal.NewMultiTenantOAuthConfig(
		env.ActiveDirectoryEndpoint, config.TenantID, []string{config.NetworkResourceTenantID}, adal.OAuthOptions{})
	if err != nil {
		return nil, fmt.Errorf("creating the multi-tenant OAuth config: %w", err)
	}

	if len(config.AADClientSecret) > 0 {
		klog.V(2).Infoln("azure: using client_id+client_secret to retrieve multi-tenant access token")
		return adal.NewMultiTenantServicePrincipalToken(
			multiTenantOAuthConfig,
			config.AADClientID,
			config.AADClientSecret,
			env.ServiceManagementEndpoint)
	}

	if len(config.AADClientCertPath) > 0 {
		klog.V(2).Infoln("azure: using jwt client_assertion (client_cert+client_private_key) to retrieve multi-tenant access token")
		certData, err := os.ReadFile(config.AADClientCertPath)
		if err != nil {
			return nil, fmt.Errorf("reading the client certificate from file %s: %w", config.AADClientCertPath, err)
		}
		certificate, privateKey, err := adal.DecodePfxCertificateData(certData, config.AADClientCertPassword)
		if err != nil {
			return nil, fmt.Errorf("decoding the client certificate: %w", err)
		}
		return adal.NewMultiTenantServicePrincipalTokenFromCertificate(
			multiTenantOAuthConfig,
			config.AADClientID,
			certificate,
			privateKey,
			env.ServiceManagementEndpoint)
	}

	return nil, ErrorNoAuth
}

// GetNetworkResourceServicePrincipalToken is used when (and only when) NetworkResourceTenantID and NetworkResourceSubscriptionID are specified to have different values than TenantID and SubscriptionID.
//
// In that scenario, network resources are deployed in different AAD Tenant and Subscription than those for the cluster,
// and this method creates a new service principal token for network resources tenant based on the configuration.
//
// Azure network resource (Load Balancer, Public IP, Route Table, Network Security Group and their sub level resources) clients use this multi-tenant token, in order to operate resources in AAD Tenant specified by NetworkResourceTenantID.
func GetNetworkResourceServicePrincipalToken(config *AzureAuthConfig, env *azure.Environment) (*adal.ServicePrincipalToken, error) {
	err := config.checkConfigWhenNetworkResourceInDifferentTenant()
	if err != nil {
		return nil, fmt.Errorf("got error(%w) in getting network resources service principal token", err)
	}

	oauthConfig, err := adal.NewOAuthConfigWithAPIVersion(env.ActiveDirectoryEndpoint, config.NetworkResourceTenantID, nil)
	if err != nil {
		return nil, fmt.Errorf("creating the OAuth config for network resources tenant: %w", err)
	}

	if len(config.AADClientSecret) > 0 {
		klog.V(2).Infoln("azure: using client_id+client_secret to retrieve access token for network resources tenant")
		return adal.NewServicePrincipalToken(
			*oauthConfig,
			config.AADClientID,
			config.AADClientSecret,
			env.ServiceManagementEndpoint)
	}

	if len(config.AADClientCertPath) > 0 {
		klog.V(2).Infoln("azure: using jwt client_assertion (client_cert+client_private_key) to retrieve access token for network resources tenant")
		certData, err := os.ReadFile(config.AADClientCertPath)
		if err != nil {
			return nil, fmt.Errorf("reading the client certificate from file %s: %w", config.AADClientCertPath, err)
		}
		certificate, privateKey, err := adal.DecodePfxCertificateData(certData, config.AADClientCertPassword)
		if err != nil {
			return nil, fmt.Errorf("decoding the client certificate: %w", err)
		}
		return adal.NewServicePrincipalTokenFromCertificate(
			*oauthConfig,
			config.AADClientID,
			certificate,
			privateKey,
			env.ServiceManagementEndpoint)
	}

	return nil, ErrorNoAuth
}

// ParseAzureEnvironment returns the azure environment.
// If 'resourceManagerEndpoint' is set, the environment is computed by querying the cloud's resource manager endpoint.
// Otherwise, a pre-defined Environment is looked up by name.
func ParseAzureEnvironment(cloudName, resourceManagerEndpoint, identitySystem string) (*azure.Environment, error) {
	var env azure.Environment
	var err error
	if resourceManagerEndpoint != "" {
		klog.V(4).Infof("Loading environment from resource manager endpoint: %s", resourceManagerEndpoint)
		nameOverride := azure.OverrideProperty{Key: azure.EnvironmentName, Value: cloudName}
		env, err = azure.EnvironmentFromURL(resourceManagerEndpoint, nameOverride)
		if err == nil {
			azureStackOverrides(&env, resourceManagerEndpoint, identitySystem)
		}
	} else if cloudName == "" {
		klog.V(4).Info("Using public cloud environment")
		env = azure.PublicCloud
	} else {
		klog.V(4).Infof("Using %s environment", cloudName)
		env, err = azure.EnvironmentFromName(cloudName)
	}
	return &env, err
}

// ParseAzureAuthConfig returns a parsed configuration for an Azure cloudprovider config file
func ParseAzureAuthConfig(configReader io.Reader) (*AzureAuthConfig, *azure.Environment, error) {
	var config AzureAuthConfig

	if configReader == nil {
		return nil, nil, errors.New("nil config is provided")
	}

	limitedReader := &io.LimitedReader{R: configReader, N: maxReadLength}
	configContents, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, nil, err
	}
	if limitedReader.N <= 0 {
		return nil, nil, errors.New("the read limit is reached")
	}
	err = yaml.Unmarshal(configContents, &config)
	if err != nil {
		return nil, nil, err
	}

	environment, err := ParseAzureEnvironment(config.Cloud, config.ResourceManagerEndpoint, config.IdentitySystem)
	if err != nil {
		return nil, nil, err
	}

	return &config, environment, nil
}

// UsesNetworkResourceInDifferentTenant determines whether the AzureAuthConfig indicates to use network resources in
// different AAD Tenant than those for the cluster. Return true when NetworkResourceTenantID is specified  and not equal
// to one defined in global configs
func (config *AzureAuthConfig) UsesNetworkResourceInDifferentTenant() bool {
	return len(config.NetworkResourceTenantID) > 0 && !strings.EqualFold(config.NetworkResourceTenantID, config.TenantID)
}

// UsesNetworkResourceInDifferentSubscription determines whether the AzureAuthConfig indicates to use network resources
// in different Subscription than those for the cluster. Return true when NetworkResourceSubscriptionID is specified
// and not equal to one defined in global configs
func (config *AzureAuthConfig) UsesNetworkResourceInDifferentSubscription() bool {
	return len(config.NetworkResourceSubscriptionID) > 0 && !strings.EqualFold(config.NetworkResourceSubscriptionID, config.SubscriptionID)
}

// azureStackOverrides ensures that the Environment matches what AKSe currently generates for Azure Stack
func azureStackOverrides(env *azure.Environment, resourceManagerEndpoint, identitySystem string) {
	env.ManagementPortalURL = strings.Replace(resourceManagerEndpoint, "https://management.", "https://portal.", -1)
	env.ServiceManagementEndpoint = env.TokenAudience
	env.ResourceManagerVMDNSSuffix = strings.Replace(resourceManagerEndpoint, "https://management.", "cloudapp.", -1)
	env.ResourceManagerVMDNSSuffix = strings.TrimSuffix(env.ResourceManagerVMDNSSuffix, "/")
	if strings.EqualFold(identitySystem, consts.ADFSIdentitySystem) {
		env.ActiveDirectoryEndpoint = strings.TrimSuffix(env.ActiveDirectoryEndpoint, "/")
		env.ActiveDirectoryEndpoint = strings.TrimSuffix(env.ActiveDirectoryEndpoint, "adfs")
	}
}

// checkConfigWhenNetworkResourceInDifferentTenant checks configuration for the scenario of using network resource in different tenant
func (config *AzureAuthConfig) checkConfigWhenNetworkResourceInDifferentTenant() error {
	if !config.UsesNetworkResourceInDifferentTenant() {
		return fmt.Errorf("NetworkResourceTenantID must be configured")
	}

	if strings.EqualFold(config.IdentitySystem, consts.ADFSIdentitySystem) {
		return fmt.Errorf("ADFS identity system is not supported")
	}

	if config.UseManagedIdentityExtension {
		return fmt.Errorf("managed identity is not supported")
	}

	return nil
}
