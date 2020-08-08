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
	"io/ioutil"
	"net/http"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2017-05-10/resources"
	"github.com/Azure/azure-sdk-for-go/services/storage/mgmt/2018-07-01/storage"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"

	klog "k8s.io/klog/v2"
	"k8s.io/legacy-cloud-providers/azure/clients/containerserviceclient"
	"k8s.io/legacy-cloud-providers/azure/clients/diskclient"
	"k8s.io/legacy-cloud-providers/azure/clients/interfaceclient"
	"k8s.io/legacy-cloud-providers/azure/clients/storageaccountclient"
	"k8s.io/legacy-cloud-providers/azure/clients/vmclient"
	"k8s.io/legacy-cloud-providers/azure/clients/vmssclient"
	"k8s.io/legacy-cloud-providers/azure/clients/vmssvmclient"
)

// DeploymentsClient defines needed functions for azure network.DeploymentsClient.
type DeploymentsClient interface {
	Get(ctx context.Context, resourceGroupName string, deploymentName string) (result resources.DeploymentExtended, err error)
	List(ctx context.Context, resourceGroupName string, filter string, top *int32) (result []resources.DeploymentExtended, err error)
	ExportTemplate(ctx context.Context, resourceGroupName string, deploymentName string) (result resources.DeploymentExportResult, err error)
	CreateOrUpdate(ctx context.Context, resourceGroupName string, deploymentName string, parameters resources.Deployment) (resp *http.Response, err error)
	Delete(ctx context.Context, resourceGroupName string, deploymentName string) (resp *http.Response, err error)
}

type azDeploymentsClient struct {
	client resources.DeploymentsClient
}

func newAzDeploymentsClient(subscriptionID, endpoint string, servicePrincipalToken *adal.ServicePrincipalToken) *azDeploymentsClient {
	deploymentsClient := resources.NewDeploymentsClient(subscriptionID)
	deploymentsClient.BaseURI = endpoint
	deploymentsClient.Authorizer = autorest.NewBearerAuthorizer(servicePrincipalToken)
	deploymentsClient.PollingDelay = 5 * time.Second
	configureUserAgent(&deploymentsClient.Client)

	return &azDeploymentsClient{
		client: deploymentsClient,
	}
}

func (az *azDeploymentsClient) Get(ctx context.Context, resourceGroupName string, deploymentName string) (result resources.DeploymentExtended, err error) {
	klog.V(10).Infof("azDeploymentsClient.Get(%q,%q): start", resourceGroupName, deploymentName)
	defer func() {
		klog.V(10).Infof("azDeploymentsClient.Get(%q,%q): end", resourceGroupName, deploymentName)
	}()

	return az.client.Get(ctx, resourceGroupName, deploymentName)
}

func (az *azDeploymentsClient) ExportTemplate(ctx context.Context, resourceGroupName string, deploymentName string) (result resources.DeploymentExportResult, err error) {
	klog.V(10).Infof("azDeploymentsClient.ExportTemplate(%q,%q): start", resourceGroupName, deploymentName)
	defer func() {
		klog.V(10).Infof("azDeploymentsClient.ExportTemplate(%q,%q): end", resourceGroupName, deploymentName)
	}()

	return az.client.ExportTemplate(ctx, resourceGroupName, deploymentName)
}

func (az *azDeploymentsClient) CreateOrUpdate(ctx context.Context, resourceGroupName string, deploymentName string, parameters resources.Deployment) (resp *http.Response, err error) {
	klog.V(10).Infof("azDeploymentsClient.CreateOrUpdate(%q,%q): start", resourceGroupName, deploymentName)
	defer func() {
		klog.V(10).Infof("azDeploymentsClient.CreateOrUpdate(%q,%q): end", resourceGroupName, deploymentName)
	}()

	future, err := az.client.CreateOrUpdate(ctx, resourceGroupName, deploymentName, parameters)
	if err != nil {
		return future.Response(), err
	}

	err = future.WaitForCompletionRef(ctx, az.client.Client)
	return future.Response(), err
}

func (az *azDeploymentsClient) List(ctx context.Context, resourceGroupName, filter string, top *int32) (result []resources.DeploymentExtended, err error) {
	klog.V(10).Infof("azDeploymentsClient.List(%q): start", resourceGroupName)
	defer func() {
		klog.V(10).Infof("azDeploymentsClient.List(%q): end", resourceGroupName)
	}()

	iterator, err := az.client.ListByResourceGroupComplete(ctx, resourceGroupName, filter, top)
	if err != nil {
		return nil, err
	}

	result = make([]resources.DeploymentExtended, 0)
	for ; iterator.NotDone(); err = iterator.Next() {
		if err != nil {
			return nil, err
		}

		result = append(result, iterator.Value())
	}

	return result, err
}

func (az *azDeploymentsClient) Delete(ctx context.Context, resourceGroupName, deploymentName string) (resp *http.Response, err error) {
	klog.V(10).Infof("azDeploymentsClient.Delete(%q,%q): start", resourceGroupName, deploymentName)
	defer func() {
		klog.V(10).Infof("azDeploymentsClient.Delete(%q,%q): end", resourceGroupName, deploymentName)
	}()

	future, err := az.client.Delete(ctx, resourceGroupName, deploymentName)
	if err != nil {
		return future.Response(), err
	}

	err = future.WaitForCompletionRef(ctx, az.client.Client)
	return future.Response(), err
}

type azAccountsClient struct {
	client storage.AccountsClient
}

type azClient struct {
	virtualMachineScaleSetsClient   vmssclient.Interface
	virtualMachineScaleSetVMsClient vmssvmclient.Interface
	virtualMachinesClient           vmclient.Interface
	deploymentsClient               DeploymentsClient
	interfacesClient                interfaceclient.Interface
	disksClient                     diskclient.Interface
	storageAccountsClient           storageaccountclient.Interface
	managedKubernetesServicesClient containerserviceclient.Interface
}

// newServicePrincipalTokenFromCredentials creates a new ServicePrincipalToken using values of the
// passed credentials map.
func newServicePrincipalTokenFromCredentials(config *Config, env *azure.Environment) (*adal.ServicePrincipalToken, error) {
	oauthConfig, err := adal.NewOAuthConfig(env.ActiveDirectoryEndpoint, config.TenantID)
	if err != nil {
		return nil, fmt.Errorf("creating the OAuth config: %v", err)
	}

	if config.UseManagedIdentityExtension {
		klog.V(2).Infoln("azure: using managed identity extension to retrieve access token")
		msiEndpoint, err := adal.GetMSIVMEndpoint()
		if err != nil {
			return nil, fmt.Errorf("getting the managed service identity endpoint: %v", err)
		}
		if len(config.UserAssignedIdentityID) > 0 {
			klog.V(4).Info("azure: using User Assigned MSI ID to retrieve access token")
			return adal.NewServicePrincipalTokenFromMSIWithUserAssignedID(msiEndpoint,
				env.ServiceManagementEndpoint,
				config.UserAssignedIdentityID)
		}
		klog.V(4).Info("azure: using System Assigned MSI to retrieve access token")
		return adal.NewServicePrincipalTokenFromMSI(
			msiEndpoint,
			env.ServiceManagementEndpoint)
	}

	if len(config.AADClientSecret) > 0 {
		klog.V(2).Infoln("azure: using client_id+client_secret to retrieve access token")
		return adal.NewServicePrincipalToken(
			*oauthConfig,
			config.AADClientID,
			config.AADClientSecret,
			env.ServiceManagementEndpoint)
	}

	if len(config.AADClientCertPath) > 0 && len(config.AADClientCertPassword) > 0 {
		klog.V(2).Infoln("azure: using jwt client_assertion (client_cert+client_private_key) to retrieve access token")
		certData, err := ioutil.ReadFile(config.AADClientCertPath)
		if err != nil {
			return nil, fmt.Errorf("reading the client certificate from file %s: %v", config.AADClientCertPath, err)
		}
		certificate, privateKey, err := decodePkcs12(certData, config.AADClientCertPassword)
		if err != nil {
			return nil, fmt.Errorf("decoding the client certificate: %v", err)
		}
		return adal.NewServicePrincipalTokenFromCertificate(
			*oauthConfig,
			config.AADClientID,
			certificate,
			privateKey,
			env.ServiceManagementEndpoint)
	}

	return nil, fmt.Errorf("no credentials provided for AAD application %s", config.AADClientID)
}

func newAzClient(cfg *Config, env *azure.Environment) (*azClient, error) {
	spt, err := newServicePrincipalTokenFromCredentials(cfg, env)
	if err != nil {
		return nil, err
	}

	azClientConfig := cfg.getAzureClientConfig(spt, env)
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

	deploymentsClient := newAzDeploymentsClient(cfg.SubscriptionID, env.ResourceManagerEndpoint, spt)
	klog.V(5).Infof("Created deployments client with authorizer: %v", deploymentsClient)

	interfaceClientConfig := azClientConfig.WithRateLimiter(cfg.InterfaceRateLimit)
	interfacesClient := interfaceclient.New(interfaceClientConfig)
	klog.V(5).Infof("Created interfaces client with authorizer: %v", interfacesClient)

	accountClientConfig := azClientConfig.WithRateLimiter(cfg.StorageAccountRateLimit)
	storageAccountsClient := storageaccountclient.New(accountClientConfig)
	klog.V(5).Infof("Created storage accounts client with authorizer: %v", storageAccountsClient)

	diskClientConfig := azClientConfig.WithRateLimiter(cfg.DiskRateLimit)
	disksClient := diskclient.New(diskClientConfig)
	klog.V(5).Infof("Created disks client with authorizer: %v", disksClient)

	aksClientConfig := azClientConfig.WithRateLimiter(cfg.KubernetesServiceRateLimit)
	kubernetesServicesClient := containerserviceclient.New(aksClientConfig)
	klog.V(5).Infof("Created kubernetes services client with authorizer: %v", kubernetesServicesClient)

	return &azClient{
		disksClient:                     disksClient,
		interfacesClient:                interfacesClient,
		virtualMachineScaleSetsClient:   scaleSetsClient,
		virtualMachineScaleSetVMsClient: scaleSetVMsClient,
		deploymentsClient:               deploymentsClient,
		virtualMachinesClient:           virtualMachinesClient,
		storageAccountsClient:           storageAccountsClient,
		managedKubernetesServicesClient: kubernetesServicesClient,
	}, nil
}
