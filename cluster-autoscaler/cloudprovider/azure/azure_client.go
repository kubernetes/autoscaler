/*
Copyright 2018 The Kubernetes Authors.

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

package azure

import (
	"context"
	"fmt"
	"os"
	"time"

	_ "go.uber.org/mock/mockgen/model" // for go:generate

	azextensions "github.com/Azure/azure-sdk-for-go-extensions/pkg/middleware"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	azurecore_policy "github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	armcomputev7 "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v7"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v8"
	"github.com/Azure/go-autorest/autorest/azure"

	klog "k8s.io/klog/v2"

	"sigs.k8s.io/cloud-provider-azure/pkg/azclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/accountclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/diskclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/interfaceclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/virtualmachineclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/virtualmachinescalesetclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/virtualmachinescalesetvmclient"
)

//go:generate sh -c "mockgen -source=azure_client.go -package azure -exclude_interfaces DeploymentsClient | cat ../../../hack/boilerplate/boilerplate.go.txt - > azure_mock_agentpool_client.go"
//go:generate sh -c "mockgen -package=azure sigs.k8s.io/cloud-provider-azure/pkg/azclient/virtualmachineclient Interface | cat ../../../hack/boilerplate/boilerplate.go.txt - > azure_mock_virtualmachine_client_test.go"

const (
	vmsContextTimeout      = 5 * time.Minute
	vmsAsyncContextTimeout = 30 * time.Minute
)

// AgentPoolsClient interface defines the methods needed for scaling vms pool.
// it is implemented by track2 sdk armcontainerservice.AgentPoolsClient
type AgentPoolsClient interface {
	Get(ctx context.Context,
		resourceGroupName, resourceName, agentPoolName string,
		options *armcontainerservice.AgentPoolsClientGetOptions) (
		armcontainerservice.AgentPoolsClientGetResponse, error)
	BeginCreateOrUpdate(
		ctx context.Context,
		resourceGroupName, resourceName, agentPoolName string,
		parameters armcontainerservice.AgentPool,
		options *armcontainerservice.AgentPoolsClientBeginCreateOrUpdateOptions) (
		*runtime.Poller[armcontainerservice.AgentPoolsClientCreateOrUpdateResponse], error)
	BeginDeleteMachines(
		ctx context.Context,
		resourceGroupName, resourceName, agentPoolName string,
		machines armcontainerservice.AgentPoolDeleteMachinesParameter,
		options *armcontainerservice.AgentPoolsClientBeginDeleteMachinesOptions) (
		*runtime.Poller[armcontainerservice.AgentPoolsClientDeleteMachinesResponse], error)
	NewListPager(
		resourceGroupName, resourceName string,
		options *armcontainerservice.AgentPoolsClientListOptions,
	) *runtime.Pager[armcontainerservice.AgentPoolsClientListResponse]
}

func getAgentpoolClientCredentials(cfg *Config) (azcore.TokenCredential, error) {
	if cfg.AuthMethod == "" || cfg.AuthMethod == authMethodPrincipal {
		// Use MSI
		if cfg.UseManagedIdentityExtension {
			// Use System Assigned MSI
			if cfg.UserAssignedIdentityID == "" {
				klog.V(4).Info("Agentpool client: using System Assigned MSI to retrieve access token")
				return azidentity.NewManagedIdentityCredential(nil)
			}
			// Use User Assigned MSI
			klog.V(4).Info("Agentpool client: using User Assigned MSI to retrieve access token")
			return azidentity.NewManagedIdentityCredential(&azidentity.ManagedIdentityCredentialOptions{
				ID: azidentity.ClientID(cfg.UserAssignedIdentityID),
			})
		}

		// Use Service Principal with ClientID and ClientSecret
		if cfg.AADClientID != "" && cfg.AADClientSecret != "" {
			klog.V(2).Infoln("Agentpool client: using client_id+client_secret to retrieve access token")
			return azidentity.NewClientSecretCredential(cfg.TenantID, cfg.AADClientID, cfg.AADClientSecret, nil)
		}

		// Use Service Principal with ClientCert and AADClientCertPassword
		if cfg.AADClientID != "" && cfg.AADClientCertPath != "" {
			klog.V(2).Infoln("Agentpool client: using client_cert+client_private_key to retrieve access token")
			certData, err := os.ReadFile(cfg.AADClientCertPath)
			if err != nil {
				return nil, fmt.Errorf("reading the client certificate from file %s failed with error: %w", cfg.AADClientCertPath, err)
			}
			certs, privateKey, err := azidentity.ParseCertificates(certData, []byte(cfg.AADClientCertPassword))
			if err != nil {
				return nil, fmt.Errorf("parsing service principal certificate data failed with error: %w", err)
			}
			return azidentity.NewClientCertificateCredential(cfg.TenantID, cfg.AADClientID, certs, privateKey, &azidentity.ClientCertificateCredentialOptions{
				SendCertificateChain: true,
			})
		}
	}

	if cfg.UseFederatedWorkloadIdentityExtension {
		klog.V(4).Info("Agentpool client: using workload identity for access token")
		return azidentity.NewWorkloadIdentityCredential(&azidentity.WorkloadIdentityCredentialOptions{
			TokenFilePath: cfg.AADFederatedTokenFile,
		})
	}

	return nil, fmt.Errorf("unsupported authorization method: %s", cfg.AuthMethod)
}

func newAgentpoolClient(cfg *Config) (AgentPoolsClient, error) {
	retryOptions := azextensions.DefaultRetryOpts()
	cred, err := getAgentpoolClientCredentials(cfg)
	if err != nil {
		klog.Errorf("failed to get agent pool client credentials: %v", err)
		return nil, err
	}

	env := azure.PublicCloud // default to public cloud
	if cfg.Cloud != "" {
		var err error
		env, err = azure.EnvironmentFromName(cfg.Cloud)
		if err != nil {
			klog.Errorf("failed to get environment from name %s: with error: %v", cfg.Cloud, err)
			return nil, err
		}
	}

	if cfg.ARMBaseURLForAPClient != "" {
		klog.V(10).Infof("Using ARMBaseURLForAPClient to create agent pool client")
		return newAgentpoolClientWithConfig(cfg.SubscriptionID, cred, cfg.ARMBaseURLForAPClient, env.TokenAudience, retryOptions, true /*insecureAllowCredentialWithHTTP*/)
	}

	return newAgentpoolClientWithConfig(cfg.SubscriptionID, cred, env.ResourceManagerEndpoint, env.TokenAudience, retryOptions, false /*insecureAllowCredentialWithHTTP*/)
}

func newAgentpoolClientWithConfig(subscriptionID string, cred azcore.TokenCredential,
	cloudCfgEndpoint, cloudCfgAudience string, retryOptions azurecore_policy.RetryOptions, insecureAllowCredentialWithHTTP bool) (AgentPoolsClient, error) {
	agentPoolsClient, err := armcontainerservice.NewAgentPoolsClient(subscriptionID, cred,
		&policy.ClientOptions{
			ClientOptions: azurecore_policy.ClientOptions{
				Cloud: cloud.Configuration{
					Services: map[cloud.ServiceName]cloud.ServiceConfiguration{
						cloud.ResourceManager: {
							Endpoint: cloudCfgEndpoint,
							Audience: cloudCfgAudience,
						},
					},
				},
				InsecureAllowCredentialWithHTTP: insecureAllowCredentialWithHTTP,
				Telemetry:                       azextensions.DefaultTelemetryOpts(getUserAgentExtension()),
				Transport:                       azextensions.DefaultHTTPClient(),
				Retry:                           retryOptions,
			},
		})

	if err != nil {
		return nil, fmt.Errorf("failed to init cluster agent pools client: %w", err)
	}

	klog.V(10).Infof("Successfully created agent pool client with ARMBaseURL")
	return agentPoolsClient, nil
}

type azClient struct {
	clientFactory                   azclient.ClientFactory
	virtualMachineScaleSetsClient   virtualmachinescalesetclient.Interface
	virtualMachineScaleSetVMsClient virtualmachinescalesetvmclient.Interface
	virtualMachinesClient           virtualmachineclient.Interface
	deploymentClient                DeploymentClient
	interfacesClient                interfaceclient.Interface
	disksClient                     diskclient.Interface
	storageAccountsClient           accountclient.Interface
	skuClient                       *armcomputev7.ResourceSKUsClient
	agentPoolClient                 AgentPoolsClient
	// Wrapper for delete operations
	vmssClientForDelete VMSSDeleteClient
}

func newAzClient(cfg *Config, env *azure.Environment) (*azClient, error) {
	// Create ARMClientConfig for azclient factory
	armConfig := &azclient.ARMClientConfig{
		Cloud:                   cfg.Cloud,
		TenantID:                cfg.TenantID,
		UserAgent:               getUserAgentExtension(),
		ResourceManagerEndpoint: env.ResourceManagerEndpoint,
	}

	// Apply proxy URL or hosted subscription overrides
	if cfg.HostedResourceProxyURL != "" {
		armConfig.ResourceManagerEndpoint = cfg.HostedResourceProxyURL
	}

	// Create AzureAuthConfig for auth provider
	authConfig := &azclient.AzureAuthConfig{
		AADClientID:                           cfg.AADClientID,
		AADClientSecret:                       cfg.AADClientSecret,
		AADClientCertPath:                     cfg.AADClientCertPath,
		AADClientCertPassword:                 cfg.AADClientCertPassword,
		UseManagedIdentityExtension:           cfg.UseManagedIdentityExtension,
		UserAssignedIdentityID:                cfg.UserAssignedIdentityID,
		AADFederatedTokenFile:                 cfg.AADFederatedTokenFile,
		UseFederatedWorkloadIdentityExtension: cfg.UseFederatedWorkloadIdentityExtension,
	}

	// Create auth provider
	authProvider, err := azclient.NewAuthProvider(armConfig, authConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create auth provider: %w", err)
	}

	// Get credentials from auth provider
	cred := authProvider.GetAzIdentity()
	if cred == nil {
		return nil, fmt.Errorf("failed to get Azure credentials from auth provider")
	}

	// Create ClientFactoryConfig with subscription ID and rate limit settings
	subscriptionID := cfg.SubscriptionID
	if cfg.HostedSubscriptionID != "" {
		subscriptionID = cfg.HostedSubscriptionID
	}

	factoryConfig := &azclient.ClientFactoryConfig{
		SubscriptionID:               subscriptionID,
		CloudProviderRateLimitConfig: cfg.CloudProviderRateLimitConfig,
	}

	// Create cloud configuration for NewClientFactory
	cloudConfig := cloud.Configuration{
		Services: map[cloud.ServiceName]cloud.ServiceConfiguration{
			cloud.ResourceManager: {
				Endpoint: armConfig.ResourceManagerEndpoint,
				Audience: env.TokenAudience,
			},
		},
	}

	// Create client factory
	clientFactory, err := azclient.NewClientFactory(factoryConfig, armConfig, cloudConfig, cred)
	if err != nil {
		return nil, fmt.Errorf("failed to create azclient factory: %w", err)
	}
	klog.V(5).Infof("Created Azure client factory")

	// Create SKU client separately using v7 (it's not part of the factory, and skewer v2 requires v7)
	skuClient, err := armcomputev7.NewResourceSKUsClient(subscriptionID, cred, &policy.ClientOptions{
		ClientOptions: azurecore_policy.ClientOptions{
			Cloud: cloud.Configuration{
				Services: map[cloud.ServiceName]cloud.ServiceConfiguration{
					cloud.ResourceManager: {
						Endpoint: armConfig.ResourceManagerEndpoint,
						Audience: env.TokenAudience,
					},
				},
			},
			Telemetry: azextensions.DefaultTelemetryOpts(getUserAgentExtension()),
			Transport: azextensions.DefaultHTTPClient(),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create SKU client: %w", err)
	}
	klog.V(5).Infof("Created sku client")

	// Create agent pool client
	agentPoolClient, err := newAgentpoolClient(cfg)
	if err != nil {
		klog.Errorf("newAgentpoolClient failed with error: %s", err)
		if cfg.EnableVMsAgentPool {
			// only return error if VMs agent pool is supported which is controlled by toggle
			return nil, err
		}
	}

	// Get VMSS client from ClientFactory - the azclient's Client embeds the SDK client
	// which provides access to BeginDeleteInstances
	vmssClient := clientFactory.GetVirtualMachineScaleSetClient()

	vmssClientForDelete := NewVMSSDeleteClient(vmssClient)
	if vmssClientForDelete == nil {
		return nil, fmt.Errorf("failed to create VMSS delete client wrapper: unexpected client type")
	}

	deploymentClient := NewDeploymentClient(clientFactory.GetDeploymentClient())
	if deploymentClient == nil {
		return nil, fmt.Errorf("failed to create deployment client wrapper: unexpected client type")
	}

	return &azClient{
		clientFactory:                   clientFactory,
		virtualMachineScaleSetsClient:   vmssClient,
		virtualMachineScaleSetVMsClient: clientFactory.GetVirtualMachineScaleSetVMClient(),
		virtualMachinesClient:           clientFactory.GetVirtualMachineClient(),
		deploymentClient:                NewDeploymentClient(clientFactory.GetDeploymentClient()),
		interfacesClient:                clientFactory.GetInterfaceClient(),
		disksClient:                     clientFactory.GetDiskClient(),
		storageAccountsClient:           clientFactory.GetAccountClient(),
		skuClient:                       skuClient,
		agentPoolClient:                 agentPoolClient,
		vmssClientForDelete:             vmssClientForDelete,
	}, nil
}
