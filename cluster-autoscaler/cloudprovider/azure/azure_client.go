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

	_ "go.uber.org/mock/mockgen/model" // for go:generate

	azextensions "github.com/Azure/azure-sdk-for-go-extensions/pkg/middleware"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	azurecore_policy "github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v4"
	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-07-01/compute"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/azure/auth"

	klog "k8s.io/klog/v2"

	"sigs.k8s.io/cloud-provider-azure/pkg/azureclients/deploymentclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azureclients/diskclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azureclients/interfaceclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azureclients/storageaccountclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azureclients/vmclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azureclients/vmssclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azureclients/vmssvmclient"
	providerazureconfig "sigs.k8s.io/cloud-provider-azure/pkg/provider/config"
)

//go:generate sh -c "mockgen k8s.io/autoscaler/cluster-autoscaler/cloudprovider/azure AgentPoolsClient >./agentpool_client.go"

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
}

func getAgentpoolClientCredentials(cfg *Config) (azcore.TokenCredential, error) {
	var cred azcore.TokenCredential
	var err error
	if cfg.AuthMethod == authMethodCLI {
		cred, err = azidentity.NewAzureCLICredential(&azidentity.AzureCLICredentialOptions{
			TenantID: cfg.TenantID})
		if err != nil {
			klog.Errorf("NewAzureCLICredential failed: %v", err)
			return nil, err
		}
	} else if cfg.AuthMethod == "" || cfg.AuthMethod == authMethodPrincipal {
		cred, err = azidentity.NewClientSecretCredential(cfg.TenantID, cfg.AADClientID, cfg.AADClientSecret, nil)
		if err != nil {
			klog.Errorf("NewClientSecretCredential failed: %v", err)
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("unsupported authorization method: %s", cfg.AuthMethod)
	}
	return cred, nil
}

func getAgentpoolClientRetryOptions(cfg *Config) azurecore_policy.RetryOptions {
	if cfg.AuthMethod == authMethodCLI {
		return azurecore_policy.RetryOptions{
			MaxRetries: -1, // no retry when using CLI auth for UT
		}
	}
	return azextensions.DefaultRetryOpts()
}

func newAgentpoolClient(cfg *Config) (AgentPoolsClient, error) {
	retryOptions := getAgentpoolClientRetryOptions(cfg)

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
	skuClient                       compute.ResourceSkusClient
	agentPoolClient                 AgentPoolsClient
}

func newAuthorizer(config *Config, env *azure.Environment) (autorest.Authorizer, error) {
	switch config.AuthMethod {
	case authMethodCLI:
		return auth.NewAuthorizerFromCLI()
	case "", authMethodPrincipal:
		token, err := providerazureconfig.GetServicePrincipalToken(&config.AzureAuthConfig, env, "")
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

	// Reference on why selecting ResourceManagerEndpoint as baseURI -
	// https://github.com/Azure/go-autorest/blob/main/autorest/azure/environments.go
	skuClient := compute.NewResourceSkusClientWithBaseURI(azClientConfig.ResourceManagerEndpoint, cfg.SubscriptionID)
	skuClient.Authorizer = azClientConfig.Authorizer
	skuClient.UserAgent = azClientConfig.UserAgent
	klog.V(5).Infof("Created sku client with authorizer: %v", skuClient)

	agentPoolClient, err := newAgentpoolClient(cfg)
	if err != nil {
		// we don't want to fail the whole process so we don't break any existing functionality
		// since this may not be fatal - it is only used by vms pool which is still under development.
		klog.Warningf("newAgentpoolClient failed with error: %s", err)
	}

	return &azClient{
		disksClient:                     disksClient,
		interfacesClient:                interfacesClient,
		virtualMachineScaleSetsClient:   scaleSetsClient,
		virtualMachineScaleSetVMsClient: scaleSetVMsClient,
		deploymentClient:                deploymentClient,
		virtualMachinesClient:           virtualMachinesClient,
		storageAccountsClient:           storageAccountsClient,
		skuClient:                       skuClient,
		agentPoolClient:                 agentPoolClient,
	}, nil
}
