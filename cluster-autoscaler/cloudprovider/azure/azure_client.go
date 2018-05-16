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

	"github.com/Azure/azure-sdk-for-go/arm/disk"
	"github.com/Azure/azure-sdk-for-go/arm/network"
	"github.com/Azure/azure-sdk-for-go/arm/resources/resources"
	"github.com/Azure/azure-sdk-for-go/arm/storage"
	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2017-12-01/compute"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/golang/glog"
)

// VirtualMachineScaleSetsClient defines needed functions for azure compute.VirtualMachineScaleSetsClient.
type VirtualMachineScaleSetsClient interface {
	Get(ctx context.Context, resourceGroupName string, vmScaleSetName string) (result compute.VirtualMachineScaleSet, err error)
	CreateOrUpdate(ctx context.Context, resourceGroupName string, name string, parameters compute.VirtualMachineScaleSet) (resp *http.Response, err error)
	DeleteInstances(ctx context.Context, resourceGroupName string, vmScaleSetName string, vmInstanceIDs compute.VirtualMachineScaleSetVMInstanceRequiredIDs) (resp *http.Response, err error)
	List(ctx context.Context, resourceGroupName string) (result []compute.VirtualMachineScaleSet, err error)
}

// VirtualMachineScaleSetVMsClient defines needed functions for azure compute.VirtualMachineScaleSetVMsClient.
type VirtualMachineScaleSetVMsClient interface {
	Get(ctx context.Context, resourceGroupName string, VMScaleSetName string, instanceID string) (result compute.VirtualMachineScaleSetVM, err error)
	List(ctx context.Context, resourceGroupName string, virtualMachineScaleSetName string, filter string, selectParameter string, expand string) (result []compute.VirtualMachineScaleSetVM, err error)
}

// VirtualMachinesClient defines needed functions for azure compute.VirtualMachinesClient.
type VirtualMachinesClient interface {
	Get(ctx context.Context, resourceGroupName string, VMName string, expand compute.InstanceViewTypes) (result compute.VirtualMachine, err error)
	Delete(ctx context.Context, resourceGroupName string, VMName string) (resp *http.Response, err error)
	List(ctx context.Context, resourceGroupName string) (result []compute.VirtualMachine, err error)
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

// azVirtualMachineScaleSetsClient implements VirtualMachineScaleSetsClient.
type azVirtualMachineScaleSetsClient struct {
	client compute.VirtualMachineScaleSetsClient
}

func newAzVirtualMachineScaleSetsClient(subscriptionID, endpoint string, servicePrincipalToken *adal.ServicePrincipalToken) *azVirtualMachineScaleSetsClient {
	virtualMachineScaleSetsClient := compute.NewVirtualMachineScaleSetsClient(subscriptionID)
	virtualMachineScaleSetsClient.BaseURI = endpoint
	virtualMachineScaleSetsClient.Authorizer = autorest.NewBearerAuthorizer(servicePrincipalToken)
	virtualMachineScaleSetsClient.PollingDelay = 5 * time.Second
	configureUserAgent(&virtualMachineScaleSetsClient.Client)

	return &azVirtualMachineScaleSetsClient{
		client: virtualMachineScaleSetsClient,
	}
}

func (az *azVirtualMachineScaleSetsClient) CreateOrUpdate(ctx context.Context, resourceGroupName string, VMScaleSetName string, parameters compute.VirtualMachineScaleSet) (resp *http.Response, err error) {
	glog.V(10).Infof("azVirtualMachineScaleSetsClient.CreateOrUpdate(%q,%q): start", resourceGroupName, VMScaleSetName)
	defer func() {
		glog.V(10).Infof("azVirtualMachineScaleSetsClient.CreateOrUpdate(%q,%q): end", resourceGroupName, VMScaleSetName)
	}()

	future, err := az.client.CreateOrUpdate(ctx, resourceGroupName, VMScaleSetName, parameters)
	if err != nil {
		return future.Response(), err
	}

	err = future.WaitForCompletion(ctx, az.client.Client)
	return future.Response(), err
}

func (az *azVirtualMachineScaleSetsClient) Get(ctx context.Context, resourceGroupName string, VMScaleSetName string) (result compute.VirtualMachineScaleSet, err error) {
	glog.V(10).Infof("azVirtualMachineScaleSetsClient.Get(%q,%q): start", resourceGroupName, VMScaleSetName)
	defer func() {
		glog.V(10).Infof("azVirtualMachineScaleSetsClient.Get(%q,%q): end", resourceGroupName, VMScaleSetName)
	}()

	return az.client.Get(ctx, resourceGroupName, VMScaleSetName)
}

func (az *azVirtualMachineScaleSetsClient) List(ctx context.Context, resourceGroupName string) (result []compute.VirtualMachineScaleSet, err error) {
	glog.V(10).Infof("azVirtualMachineScaleSetsClient.List(%q,%q): start", resourceGroupName)
	defer func() {
		glog.V(10).Infof("azVirtualMachineScaleSetsClient.List(%q,%q): end", resourceGroupName)
	}()

	iterator, err := az.client.ListComplete(ctx, resourceGroupName)
	if err != nil {
		return nil, err
	}

	result = make([]compute.VirtualMachineScaleSet, 0)
	for ; iterator.NotDone(); err = iterator.Next() {
		if err != nil {
			return nil, err
		}

		result = append(result, iterator.Value())
	}

	return result, nil
}

func (az *azVirtualMachineScaleSetsClient) DeleteInstances(ctx context.Context, resourceGroupName string, vmScaleSetName string, vmInstanceIDs compute.VirtualMachineScaleSetVMInstanceRequiredIDs) (resp *http.Response, err error) {
	glog.V(10).Infof("azVirtualMachineScaleSetsClient.DeleteInstances(%q,%q,%q): start", resourceGroupName, vmScaleSetName, vmInstanceIDs)
	defer func() {
		glog.V(10).Infof("azVirtualMachineScaleSetsClient.DeleteInstances(%q,%q,%q): end", resourceGroupName, vmScaleSetName, vmInstanceIDs)
	}()

	future, err := az.client.DeleteInstances(ctx, resourceGroupName, vmScaleSetName, vmInstanceIDs)
	if err != nil {
		return future.Response(), err
	}

	err = future.WaitForCompletion(ctx, az.client.Client)
	return future.Response(), err
}

// azVirtualMachineScaleSetVMsClient implements VirtualMachineScaleSetVMsClient.
type azVirtualMachineScaleSetVMsClient struct {
	client compute.VirtualMachineScaleSetVMsClient
}

func newAzVirtualMachineScaleSetVMsClient(subscriptionID, endpoint string, servicePrincipalToken *adal.ServicePrincipalToken) *azVirtualMachineScaleSetVMsClient {
	virtualMachineScaleSetVMsClient := compute.NewVirtualMachineScaleSetVMsClient(subscriptionID)
	virtualMachineScaleSetVMsClient.BaseURI = endpoint
	virtualMachineScaleSetVMsClient.Authorizer = autorest.NewBearerAuthorizer(servicePrincipalToken)
	virtualMachineScaleSetVMsClient.PollingDelay = 5 * time.Second
	configureUserAgent(&virtualMachineScaleSetVMsClient.Client)

	return &azVirtualMachineScaleSetVMsClient{
		client: virtualMachineScaleSetVMsClient,
	}
}

func (az *azVirtualMachineScaleSetVMsClient) Get(ctx context.Context, resourceGroupName string, VMScaleSetName string, instanceID string) (result compute.VirtualMachineScaleSetVM, err error) {
	glog.V(10).Infof("azVirtualMachineScaleSetVMsClient.Get(%q,%q,%q): start", resourceGroupName, VMScaleSetName, instanceID)
	defer func() {
		glog.V(10).Infof("azVirtualMachineScaleSetVMsClient.Get(%q,%q,%q): end", resourceGroupName, VMScaleSetName, instanceID)
	}()

	return az.client.Get(ctx, resourceGroupName, VMScaleSetName, instanceID)
}

func (az *azVirtualMachineScaleSetVMsClient) List(ctx context.Context, resourceGroupName string, virtualMachineScaleSetName string, filter string, selectParameter string, expand string) (result []compute.VirtualMachineScaleSetVM, err error) {
	glog.V(10).Infof("azVirtualMachineScaleSetVMsClient.List(%q,%q,%q): start", resourceGroupName, virtualMachineScaleSetName, filter)
	defer func() {
		glog.V(10).Infof("azVirtualMachineScaleSetVMsClient.List(%q,%q,%q): end", resourceGroupName, virtualMachineScaleSetName, filter)
	}()

	iterator, err := az.client.ListComplete(ctx, resourceGroupName, virtualMachineScaleSetName, filter, selectParameter, expand)
	if err != nil {
		return nil, err
	}

	result = make([]compute.VirtualMachineScaleSetVM, 0)
	for ; iterator.NotDone(); err = iterator.Next() {
		if err != nil {
			return nil, err
		}

		result = append(result, iterator.Value())
	}

	return result, nil
}

// azVirtualMachinesClient implements VirtualMachinesClient.
type azVirtualMachinesClient struct {
	client compute.VirtualMachinesClient
}

func newAzVirtualMachinesClient(subscriptionID, endpoint string, servicePrincipalToken *adal.ServicePrincipalToken) *azVirtualMachinesClient {
	virtualMachinesClient := compute.NewVirtualMachinesClient(subscriptionID)
	virtualMachinesClient.BaseURI = endpoint
	virtualMachinesClient.Authorizer = autorest.NewBearerAuthorizer(servicePrincipalToken)
	virtualMachinesClient.PollingDelay = 5 * time.Second
	configureUserAgent(&virtualMachinesClient.Client)

	return &azVirtualMachinesClient{
		client: virtualMachinesClient,
	}
}

func (az *azVirtualMachinesClient) Get(ctx context.Context, resourceGroupName string, VMName string, expand compute.InstanceViewTypes) (result compute.VirtualMachine, err error) {
	glog.V(10).Infof("azVirtualMachinesClient.Get(%q,%q,%q): start", resourceGroupName, VMName, expand)
	defer func() {
		glog.V(10).Infof("azVirtualMachinesClient.Get(%q,%q,%q): end", resourceGroupName, VMName, expand)
	}()

	return az.client.Get(ctx, resourceGroupName, VMName, expand)
}

func (az *azVirtualMachinesClient) Delete(ctx context.Context, resourceGroupName string, VMName string) (resp *http.Response, err error) {
	glog.V(10).Infof("azVirtualMachinesClient.Delete(%q,%q): start", resourceGroupName, VMName)
	defer func() {
		glog.V(10).Infof("azVirtualMachinesClient.Delete(%q,%q): end", resourceGroupName, VMName)
	}()

	future, err := az.client.Delete(ctx, resourceGroupName, VMName)
	if err != nil {
		return future.Response(), err
	}

	err = future.WaitForCompletion(ctx, az.client.Client)
	return future.Response(), err
}

func (az *azVirtualMachinesClient) List(ctx context.Context, resourceGroupName string) (result []compute.VirtualMachine, err error) {
	glog.V(10).Infof("azVirtualMachinesClient.List(%q): start", resourceGroupName)
	defer func() {
		glog.V(10).Infof("azVirtualMachinesClient.List(%q): end", resourceGroupName)
	}()

	iterator, err := az.client.ListComplete(ctx, resourceGroupName)
	if err != nil {
		return nil, err
	}

	result = make([]compute.VirtualMachine, 0)
	for ; iterator.NotDone(); err = iterator.Next() {
		if err != nil {
			return nil, err
		}

		result = append(result, iterator.Value())
	}

	return result, nil
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

	scaleSetsClient := newAzVirtualMachineScaleSetsClient(cfg.SubscriptionID, env.ResourceManagerEndpoint, spt)
	glog.V(5).Infof("Created scale set client with authorizer: %v", scaleSetsClient)

	scaleSetVMsClient := newAzVirtualMachineScaleSetVMsClient(cfg.SubscriptionID, env.ResourceManagerEndpoint, spt)
	glog.V(5).Infof("Created scale set vm client with authorizer: %v", scaleSetVMsClient)

	virtualMachinesClient := newAzVirtualMachinesClient(cfg.SubscriptionID, env.ResourceManagerEndpoint, spt)
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
