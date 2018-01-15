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
	"fmt"
	"io/ioutil"
	"time"

	"github.com/Azure/azure-sdk-for-go/arm/compute"
	"github.com/Azure/azure-sdk-for-go/arm/disk"
	"github.com/Azure/azure-sdk-for-go/arm/network"
	"github.com/Azure/azure-sdk-for-go/arm/resources/resources"
	"github.com/Azure/azure-sdk-for-go/arm/storage"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/golang/glog"
)

// VirtualMachineScaleSetsClient defines needed functions for azure compute.VirtualMachineScaleSetsClient.
type VirtualMachineScaleSetsClient interface {
	Get(resourceGroupName string, vmScaleSetName string) (result compute.VirtualMachineScaleSet, err error)
	CreateOrUpdate(resourceGroupName string, name string, parameters compute.VirtualMachineScaleSet, cancel <-chan struct{}) (<-chan compute.VirtualMachineScaleSet, <-chan error)
	DeleteInstances(resourceGroupName string, vmScaleSetName string, vmInstanceIDs compute.VirtualMachineScaleSetVMInstanceRequiredIDs, cancel <-chan struct{}) (<-chan compute.OperationStatusResponse, <-chan error)
	List(resourceGroupName string) (result compute.VirtualMachineScaleSetListResult, err error)
	ListNextResults(lastResults compute.VirtualMachineScaleSetListResult) (result compute.VirtualMachineScaleSetListResult, err error)
}

// VirtualMachineScaleSetVMsClient defines needed functions for azure compute.VirtualMachineScaleSetVMsClient.
type VirtualMachineScaleSetVMsClient interface {
	List(resourceGroupName string, virtualMachineScaleSetName string, filter string, selectParameter string, expand string) (result compute.VirtualMachineScaleSetVMListResult, err error)
	ListNextResults(lastResults compute.VirtualMachineScaleSetVMListResult) (result compute.VirtualMachineScaleSetVMListResult, err error)
}

// VirtualMachinesClient defines needed functions for azure compute.VirtualMachinesClient.
type VirtualMachinesClient interface {
	Get(resourceGroupName string, VMName string, expand compute.InstanceViewTypes) (result compute.VirtualMachine, err error)
	Delete(resourceGroupName string, VMName string, cancel <-chan struct{}) (<-chan compute.OperationStatusResponse, <-chan error)
	List(resourceGroupName string) (result compute.VirtualMachineListResult, err error)
	ListNextResults(lastResults compute.VirtualMachineListResult) (result compute.VirtualMachineListResult, err error)
}

// InterfacesClient defines needed functions for azure network.InterfacesClient.
type InterfacesClient interface {
	Delete(resourceGroupName string, networkInterfaceName string, cancel <-chan struct{}) (<-chan autorest.Response, <-chan error)
}

// DeploymentsClient defines needed functions for azure network.DeploymentsClient.
type DeploymentsClient interface {
	Get(resourceGroupName string, deploymentName string) (result resources.DeploymentExtended, err error)
	ExportTemplate(resourceGroupName string, deploymentName string) (result resources.DeploymentExportResult, err error)
	CreateOrUpdate(resourceGroupName string, deploymentName string, parameters resources.Deployment, cancel <-chan struct{}) (<-chan resources.DeploymentExtended, <-chan error)
}

// DisksClient defines needed functions for azure disk.DisksClient.
type DisksClient interface {
	Delete(resourceGroupName string, diskName string, cancel <-chan struct{}) (<-chan disk.OperationStatusResponse, <-chan error)
}

// AccountsClient defines needed functions for azure storage.AccountsClient.
type AccountsClient interface {
	ListKeys(resourceGroupName string, accountName string) (result storage.AccountListKeysResult, err error)
}

type azClient struct {
	virtualMachineScaleSetsClient   VirtualMachineScaleSetsClient
	virtualMachineScaleSetVMsClient VirtualMachineScaleSetVMsClient
	virtualMachinesClient           VirtualMachinesClient
	deploymentsClient               DeploymentsClient
	interfacesClient                InterfacesClient
	disksClient                     DisksClient
	storageAccountsClient           AccountsClient
}

// newServicePrincipalTokenFromCredentials creates a new ServicePrincipalToken using values of the
// passed credentials map.
func newServicePrincipalTokenFromCredentials(config *Config, env *azure.Environment) (*adal.ServicePrincipalToken, error) {
	oauthConfig, err := adal.NewOAuthConfig(env.ActiveDirectoryEndpoint, config.TenantID)
	if err != nil {
		return nil, fmt.Errorf("creating the OAuth config: %v", err)
	}

	if config.UseManagedIdentityExtension {
		glog.V(2).Infoln("azure: using managed identity extension to retrieve access token")
		msiEndpoint, err := adal.GetMSIVMEndpoint()
		if err != nil {
			return nil, fmt.Errorf("Getting the managed service identity endpoint: %v", err)
		}
		return adal.NewServicePrincipalTokenFromMSI(
			msiEndpoint,
			env.ServiceManagementEndpoint)
	}

	if len(config.AADClientSecret) > 0 {
		glog.V(2).Infoln("azure: using client_id+client_secret to retrieve access token")
		return adal.NewServicePrincipalToken(
			*oauthConfig,
			config.AADClientID,
			config.AADClientSecret,
			env.ServiceManagementEndpoint)
	}

	if len(config.AADClientCertPath) > 0 && len(config.AADClientCertPassword) > 0 {
		glog.V(2).Infoln("azure: using jwt client_assertion (client_cert+client_private_key) to retrieve access token")
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

	return nil, fmt.Errorf("No credentials provided for AAD application %s", config.AADClientID)
}

func newAzClient(cfg *Config, env *azure.Environment) (*azClient, error) {
	spt, err := newServicePrincipalTokenFromCredentials(cfg, env)
	if err != nil {
		return nil, err
	}

	scaleSetsClient := compute.NewVirtualMachineScaleSetsClient(cfg.SubscriptionID)
	scaleSetsClient.BaseURI = env.ResourceManagerEndpoint
	scaleSetsClient.Authorizer = autorest.NewBearerAuthorizer(spt)
	scaleSetsClient.PollingDelay = 5 * time.Second
	configureUserAgent(&scaleSetsClient.Client)
	glog.V(5).Infof("Created scale set client with authorizer: %v", scaleSetsClient)

	scaleSetVMsClient := compute.NewVirtualMachineScaleSetVMsClient(cfg.SubscriptionID)
	scaleSetVMsClient.BaseURI = env.ResourceManagerEndpoint
	scaleSetVMsClient.Authorizer = autorest.NewBearerAuthorizer(spt)
	scaleSetVMsClient.PollingDelay = 5 * time.Second
	configureUserAgent(&scaleSetVMsClient.Client)
	glog.V(5).Infof("Created scale set vm client with authorizer: %v", scaleSetVMsClient)

	virtualMachinesClient := compute.NewVirtualMachinesClient(cfg.SubscriptionID)
	virtualMachinesClient.BaseURI = env.ResourceManagerEndpoint
	virtualMachinesClient.Authorizer = autorest.NewBearerAuthorizer(spt)
	virtualMachinesClient.PollingDelay = 5 * time.Second
	configureUserAgent(&virtualMachinesClient.Client)
	glog.V(5).Infof("Created vm client with authorizer: %v", virtualMachinesClient)

	deploymentsClient := resources.NewDeploymentsClient(cfg.SubscriptionID)
	deploymentsClient.BaseURI = env.ResourceManagerEndpoint
	deploymentsClient.Authorizer = autorest.NewBearerAuthorizer(spt)
	deploymentsClient.PollingDelay = 5 * time.Second
	configureUserAgent(&deploymentsClient.Client)
	glog.V(5).Infof("Created deployments client with authorizer: %v", deploymentsClient)

	interfacesClient := network.NewInterfacesClient(cfg.SubscriptionID)
	interfacesClient.BaseURI = env.ResourceManagerEndpoint
	interfacesClient.Authorizer = autorest.NewBearerAuthorizer(spt)
	interfacesClient.PollingDelay = 5 * time.Second
	glog.V(5).Infof("Created interfaces client with authorizer: %v", interfacesClient)

	storageAccountsClient := storage.NewAccountsClient(cfg.SubscriptionID)
	storageAccountsClient.BaseURI = env.ResourceManagerEndpoint
	storageAccountsClient.Authorizer = autorest.NewBearerAuthorizer(spt)
	storageAccountsClient.PollingDelay = 5 * time.Second
	glog.V(5).Infof("Created storage accounts client with authorizer: %v", storageAccountsClient)

	disksClient := disk.NewDisksClient(cfg.SubscriptionID)
	disksClient.BaseURI = env.ResourceManagerEndpoint
	disksClient.Authorizer = autorest.NewBearerAuthorizer(spt)
	disksClient.PollingDelay = 5 * time.Second
	glog.V(5).Infof("Created disks client with authorizer: %v", disksClient)

	return &azClient{
		disksClient:                     disksClient,
		interfacesClient:                interfacesClient,
		virtualMachineScaleSetsClient:   scaleSetsClient,
		virtualMachineScaleSetVMsClient: scaleSetVMsClient,
		deploymentsClient:               deploymentsClient,
		virtualMachinesClient:           virtualMachinesClient,
		storageAccountsClient:           storageAccountsClient,
	}, nil
}
