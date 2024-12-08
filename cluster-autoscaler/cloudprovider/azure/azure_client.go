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
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v5"
	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2022-03-01/compute"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/azure/auth"

	klog "k8s.io/klog/v2"

	"sigs.k8s.io/cloud-provider-azure/pkg/azureclients/containerserviceclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azureclients/deploymentclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azureclients/diskclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azureclients/interfaceclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azureclients/storageaccountclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azureclients/vmclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azureclients/vmssclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azureclients/vmssvmclient"
	providerazureconfig "sigs.k8s.io/cloud-provider-azure/pkg/provider/config"
)

//go:generate sh -c "mockgen -source=azure_client.go -destination azure_mock_agentpool_client.go -package azure -exclude_interfaces DeploymentsClient"

const (
	vmsGetRequestContextTimeout    = 1 * time.Minute
	vmsListRequestContextTimeout   = 5 * time.Minute
	vmsPutRequestContextTimeout    = 10 * time.Minute
	vmsDeleteRequestContextTimeout = 10 * time.Minute
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
			if len(cfg.UserAssignedIdentityID) == 0 {
				klog.V(4).Info("Agentpool client: using System Assigned MSI to retrieve access token")
				return azidentity.NewManagedIdentityCredential(nil)
			}
			// Use User Assigned MSI
			klog.V(4).Info("Agentpool client: using User Assigned MSI to retrieve access token")
			return azidentity.NewManagedIdentityCredential(&azidentity.ManagedIdentityCredentialOptions{
				ID: azidentity.ClientID(cfg.UserAssignedIdentityID),
			})
		}

		// Use Service Principal
		if len(cfg.AADClientID) > 0 && len(cfg.AADClientSecret) > 0 {
			klog.V(2).Infoln("Agentpool client: using client_id+client_secret to retrieve access token")
			return azidentity.NewClientSecretCredential(cfg.TenantID, cfg.AADClientID, cfg.AADClientSecret, nil)
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

	if cfg.ARMBaseURLForAPClient != "" {
		klog.V(10).Infof("Using ARMBaseURLForAPClient to create agent pool client")
		return newAgentpoolClientWithConfig(cfg.SubscriptionID, nil, cfg.ARMBaseURLForAPClient, "UNKNOWN", retryOptions)
	}

	return newAgentpoolClientWithPublicEndpoint(cfg, retryOptions)
}

func newAgentpoolClientWithConfig(subscriptionID string, cred azcore.TokenCredential,
	cloudCfgEndpoint, cloudCfgAudience string, retryOptions azurecore_policy.RetryOptions) (AgentPoolsClient, error) {
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
				Telemetry: azextensions.DefaultTelemetryOpts(getUserAgentExtension()),
				Transport: azextensions.DefaultHTTPClient(),
				Retry:     retryOptions,
			},
		})

	if err != nil {
		return nil, fmt.Errorf("failed to init cluster agent pools client: %w", err)
	}

	klog.V(10).Infof("Successfully created agent pool client with ARMBaseURL")
	return agentPoolsClient, nil
}

func newAgentpoolClientWithPublicEndpoint(cfg *Config, retryOptions azurecore_policy.RetryOptions) (AgentPoolsClient, error) {
	cred, err := getAgentpoolClientCredentials(cfg)
	if err != nil {
		klog.Errorf("failed to get agent pool client credentials: %v", err)
		return nil, err
	}

	// default to public cloud
	env := azure.PublicCloud
	if cfg.Cloud != "" {
		env, err = azure.EnvironmentFromName(cfg.Cloud)
		if err != nil {
			klog.Errorf("failed to get environment from name %s: with error: %v", cfg.Cloud, err)
			return nil, err
		}
	}

	return newAgentpoolClientWithConfig(cfg.SubscriptionID, cred, env.ResourceManagerEndpoint, env.TokenAudience, retryOptions)
}

type azClient struct {
	virtualMachineScaleSetsClient   vmssclient.Interface
	virtualMachineScaleSetVMsClient vmssvmclient.Interface
	virtualMachinesClient           vmclient.Interface
	deploymentClient                deploymentclient.Interface
	interfacesClient                interfaceclient.Interface
	disksClient                     diskclient.Interface
	storageAccountsClient           storageaccountclient.Interface
	managedKubernetesServicesClient containerserviceclient.Interface
	skuClient                       compute.ResourceSkusClient
	agentPoolClient                 AgentPoolsClient
}

// Only exist due to cloud-provider-azure not yet supporting workload identity in 1.27
func newServicePrincipalTokenForWorkloadIdentity(config *Config, env *azure.Environment) (*adal.ServicePrincipalToken, error) {
	oauthConfig, err := adal.NewOAuthConfig(env.ActiveDirectoryEndpoint, config.TenantID)
	if err != nil {
		return nil, fmt.Errorf("creating the OAuth config: %v", err)
	}

	klog.V(2).Infoln("azure: using workload identity extension to retrieve access token")
	jwt, err := os.ReadFile(config.AADFederatedTokenFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read a file with a federated token: %v", err)
	}
	//nolint SA1019 - deprecated package
	token, err := adal.NewServicePrincipalTokenFromFederatedToken(*oauthConfig, config.AADClientID, string(jwt), env.ResourceManagerEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to create a workload identity token: %v", err)
	}
	return token, nil
}

func newAuthorizer(config *Config, env *azure.Environment) (autorest.Authorizer, error) {
	switch config.AuthMethod {
	case authMethodCLI:
		return auth.NewAuthorizerFromCLI()
	case "", authMethodPrincipal:
		var token *adal.ServicePrincipalToken
		var err error
		if config.UseFederatedWorkloadIdentityExtension {
			token, err = newServicePrincipalTokenForWorkloadIdentity(config, env)
		} else {
			token, err = providerazureconfig.GetServicePrincipalToken(&config.AzureAuthConfig, env, "")
		}
		if err != nil {
			return nil, fmt.Errorf("retrieve service principal token: %v", err)
		}
		return autorest.NewBearerAuthorizer(token), nil
	default:
		return nil, fmt.Errorf("unsupported authorization method: %s", config.AuthMethod)
	}
}

func newAzClient(cfg *Config, env *azure.Environment) (*azClient, error) {
	authorizer, err := newAuthorizer(cfg, env)
	if err != nil {
		return nil, err
	}

	azClientConfig := cfg.getAzureClientConfig(authorizer, env)
	azClientConfig.UserAgent = getUserAgentExtension()

	vmssClientConfig := azClientConfig.WithRateLimiter(cfg.VirtualMachineScaleSetRateLimit)
	scaleSetsClient := vmssclient.New(vmssClientConfig)
	klog.V(5).Infof("Created scale set client with authorizer: %v", scaleSetsClient)

	vmssVMClientConfig := azClientConfig.WithRateLimiter(cfg.VirtualMachineScaleSetRateLimit)
	scaleSetVMsClient := vmssvmclient.New(vmssVMClientConfig)
	klog.V(5).Infof("Created scale set vm client with authorizer: %v", scaleSetVMsClient)

	vmClientConfig := azClientConfig.WithRateLimiter(cfg.VirtualMachineRateLimit)
	virtualMachinesClient := vmclient.New(vmClientConfig)
	klog.V(5).Infof("Created vm client with authorizer: %v", virtualMachinesClient)

	deploymentConfig := azClientConfig.WithRateLimiter(cfg.DeploymentRateLimit)
	deploymentClient := deploymentclient.New(deploymentConfig)
	klog.V(5).Infof("Created deployments client with authorizer: %v", deploymentClient)

	interfaceClientConfig := azClientConfig.WithRateLimiter(cfg.InterfaceRateLimit)
	interfacesClient := interfaceclient.New(interfaceClientConfig)
	klog.V(5).Infof("Created interfaces client with authorizer: %v", interfacesClient)

	accountClientConfig := azClientConfig.WithRateLimiter(cfg.StorageAccountRateLimit)
	storageAccountsClient := storageaccountclient.New(accountClientConfig)
	klog.V(5).Infof("Created storage accounts client with authorizer: %v", storageAccountsClient)

	diskClientConfig := azClientConfig.WithRateLimiter(cfg.DiskRateLimit)
	disksClient := diskclient.New(diskClientConfig)
	klog.V(5).Infof("Created disks client with authorizer: %v", disksClient)

	aksClientConfig := azClientConfig.WithRateLimiter(&cfg.RateLimitConfig)
	kubernetesServicesClient := containerserviceclient.New(aksClientConfig)
	klog.V(5).Infof("Created kubernetes services client with authorizer: %v", kubernetesServicesClient)

	// Reference on why selecting ResourceManagerEndpoint as baseURI -
	// https://github.com/Azure/go-autorest/blob/main/autorest/azure/environments.go
	skuClient := compute.NewResourceSkusClientWithBaseURI(azClientConfig.ResourceManagerEndpoint, cfg.SubscriptionID)
	skuClient.Authorizer = azClientConfig.Authorizer
	skuClient.UserAgent = azClientConfig.UserAgent
	klog.V(5).Infof("Created sku client with authorizer: %v", skuClient)

	agentPoolClient, err := newAgentpoolClient(cfg)
	if err != nil {
		klog.Errorf("newAgentpoolClient failed with error: %s", err)
		if cfg.EnableVMsAgentPool {
			// only return error if VMs agent pool is supported which is controlled by toggle
			return nil, err
		}
	}

	return &azClient{
		disksClient:                     disksClient,
		interfacesClient:                interfacesClient,
		virtualMachineScaleSetsClient:   scaleSetsClient,
		virtualMachineScaleSetVMsClient: scaleSetVMsClient,
		deploymentClient:                deploymentClient,
		virtualMachinesClient:           virtualMachinesClient,
		storageAccountsClient:           storageAccountsClient,
		managedKubernetesServicesClient: kubernetesServicesClient,
		skuClient:                       skuClient,
		agentPoolClient:                 agentPoolClient,
	}, nil
}
