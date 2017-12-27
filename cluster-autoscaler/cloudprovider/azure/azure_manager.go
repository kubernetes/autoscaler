/*
Copyright 2017 The Kubernetes Authors.

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
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"sync"
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

	"gopkg.in/gcfg.v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
)

const (
	vmTypeVMSS     = "vmss"
	vmTypeStandard = "standard"
)

// VirtualMachineScaleSetsClient defines needed functions for azure compute.VirtualMachineScaleSetsClient.
type VirtualMachineScaleSetsClient interface {
	Get(resourceGroupName string, vmScaleSetName string) (result compute.VirtualMachineScaleSet, err error)
	CreateOrUpdate(resourceGroupName string, name string, parameters compute.VirtualMachineScaleSet, cancel <-chan struct{}) (<-chan compute.VirtualMachineScaleSet, <-chan error)
	DeleteInstances(resourceGroupName string, vmScaleSetName string, vmInstanceIDs compute.VirtualMachineScaleSetVMInstanceRequiredIDs, cancel <-chan struct{}) (<-chan compute.OperationStatusResponse, <-chan error)
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

// AzureManager handles Azure communication and data caching.
type AzureManager struct {
	config *Config
	env    azure.Environment

	virtualMachineScaleSetsClient   VirtualMachineScaleSetsClient
	virtualMachineScaleSetVMsClient VirtualMachineScaleSetVMsClient
	virtualMachinesClient           VirtualMachinesClient
	deploymentsClient               DeploymentsClient
	interfacesClient                InterfacesClient
	disksClient                     DisksClient
	storageAccountsClient           AccountsClient

	nodeGroups []cloudprovider.NodeGroup
	// cache of mapping from instance name to nodeGroup.
	nodeGroupsCache map[AzureRef]cloudprovider.NodeGroup
	// cache of mapping from instance name to instanceID.
	instanceIDsCache map[string]string

	cacheMutex sync.Mutex
	interrupt  chan struct{}
}

// Config holds the configuration parsed from the --cloud-config flag
type Config struct {
	Cloud          string `json:"cloud" yaml:"cloud"`
	TenantID       string `json:"tenantId" yaml:"tenantId"`
	SubscriptionID string `json:"subscriptionId" yaml:"subscriptionId"`
	ResourceGroup  string `json:"resourceGroup" yaml:"resourceGroup"`
	VMType         string `json:"vmType" yaml:"vmType"`

	AADClientID                 string `json:"aadClientId" yaml:"aadClientId"`
	AADClientSecret             string `json:"aadClientSecret" yaml:"aadClientSecret"`
	AADClientCertPath           string `json:"aadClientCertPath" yaml:"aadClientCertPath"`
	AADClientCertPassword       string `json:"aadClientCertPassword" yaml:"aadClientCertPassword"`
	UseManagedIdentityExtension bool   `json:"useManagedIdentityExtension" yaml:"useManagedIdentityExtension"`

	// Configs only for standard vmType (agent pools).
	Deployment           string `json:"deployment" yaml:"deployment"`
	APIServerPrivateKey  string `json:"apiServerPrivateKey" yaml:"apiServerPrivateKey"`
	CAPrivateKey         string `json:"caPrivateKey" yaml:"caPrivateKey"`
	ClientPrivateKey     string `json:"clientPrivateKey" yaml:"clientPrivateKey"`
	KubeConfigPrivateKey string `json:"kubeConfigPrivateKey" yaml:"kubeConfigPrivateKey"`
	WindowsAdminPassword string `json:"windowsAdminPassword" yaml:"windowsAdminPassword"`
}

// CreateAzureManager creates Azure Manager object to work with Azure.
func CreateAzureManager(configReader io.Reader) (*AzureManager, error) {
	var err error
	var cfg Config

	if configReader != nil {
		if err := gcfg.ReadInto(&cfg, configReader); err != nil {
			glog.Errorf("Couldn't read config: %v", err)
			return nil, err
		}
	} else {
		cfg.Cloud = os.Getenv("ARM_CLOUD")
		cfg.SubscriptionID = os.Getenv("ARM_SUBSCRIPTION_ID")
		cfg.ResourceGroup = os.Getenv("ARM_RESOURCE_GROUP")
		cfg.TenantID = os.Getenv("ARM_TENANT_ID")
		cfg.AADClientID = os.Getenv("ARM_CLIENT_ID")
		cfg.AADClientSecret = os.Getenv("ARM_CLIENT_SECRET")
		cfg.VMType = strings.ToLower(os.Getenv("ARM_VM_TYPE"))
		cfg.AADClientCertPath = os.Getenv("ARM_CLIENT_CERT_PATH")
		cfg.AADClientCertPassword = os.Getenv("ARM_CLIENT_CERT_PASSWORD")
		cfg.Deployment = os.Getenv("ARM_DEPLOYMENT")
		cfg.APIServerPrivateKey = os.Getenv("ARM_APISEVER_PRIVATE_KEY")
		cfg.CAPrivateKey = os.Getenv("ARM_CA_PRIVATE_KEY")
		cfg.ClientPrivateKey = os.Getenv("ARM_CLIENT_PRIVATE_KEY")
		cfg.KubeConfigPrivateKey = os.Getenv("ARM_KUBECONFIG_PRIVATE_KEY")
		cfg.WindowsAdminPassword = os.Getenv("ARM_WINDOWS_ADMIN_PASSWORD")

		useManagedIdentityExtensionFromEnv := os.Getenv("ARM_USE_MANAGED_IDENTITY_EXTENSION")
		if len(useManagedIdentityExtensionFromEnv) > 0 {
			cfg.UseManagedIdentityExtension, err = strconv.ParseBool(useManagedIdentityExtensionFromEnv)
			if err != nil {
				return nil, err
			}
		}
	}

	// Defaulting vmType to vmss.
	if cfg.VMType == "" {
		cfg.VMType = vmTypeVMSS
	}

	env := azure.PublicCloud
	if cfg.Cloud != "" {
		env, err = azure.EnvironmentFromName(cfg.Cloud)
		if err != nil {
			return nil, err
		}
	}

	if cfg.ResourceGroup == "" {
		return nil, fmt.Errorf("resource group not set")
	}

	if cfg.SubscriptionID == "" {
		return nil, fmt.Errorf("subscription ID not set")
	}

	if cfg.TenantID == "" {
		return nil, fmt.Errorf("tenant ID not set")
	}

	if cfg.AADClientID == "" {
		return nil, fmt.Errorf("ARM Client ID not set")
	}

	if cfg.VMType == vmTypeStandard {
		if cfg.Deployment == "" {
			return nil, fmt.Errorf("deployment not set")
		}

		if cfg.APIServerPrivateKey == "" {
			return nil, fmt.Errorf("apiServerPrivateKey not set")
		}

		if cfg.CAPrivateKey == "" {
			return nil, fmt.Errorf("caPrivateKey not set")
		}

		if cfg.ClientPrivateKey == "" {
			return nil, fmt.Errorf("clientPrivateKey not set")
		}

		if cfg.KubeConfigPrivateKey == "" {
			return nil, fmt.Errorf("kubeConfigPrivateKey not set")
		}
	}

	glog.Infof("Starting azure manager with subscription ID %q", cfg.SubscriptionID)

	spt, err := NewServicePrincipalTokenFromCredentials(&cfg, &env)
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

	// Create azure manager.
	manager := &AzureManager{
		config:                          &cfg,
		env:                             env,
		disksClient:                     disksClient,
		interfacesClient:                interfacesClient,
		virtualMachineScaleSetsClient:   scaleSetsClient,
		virtualMachineScaleSetVMsClient: scaleSetVMsClient,
		deploymentsClient:               deploymentsClient,
		virtualMachinesClient:           virtualMachinesClient,
		storageAccountsClient:           storageAccountsClient,

		interrupt:        make(chan struct{}),
		instanceIDsCache: make(map[string]string),
		nodeGroups:       make([]cloudprovider.NodeGroup, 0),
		nodeGroupsCache:  make(map[AzureRef]cloudprovider.NodeGroup),
	}

	go wait.Until(func() {
		manager.cacheMutex.Lock()
		defer manager.cacheMutex.Unlock()
		if err := manager.regenerateCache(); err != nil {
			glog.Errorf("Error while regenerating AS cache: %v", err)
		}
	}, 5*time.Minute, manager.interrupt)

	return manager, nil
}

// NewServicePrincipalTokenFromCredentials creates a new ServicePrincipalToken using values of the
// passed credentials map.
func NewServicePrincipalTokenFromCredentials(config *Config, env *azure.Environment) (*adal.ServicePrincipalToken, error) {
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

// RegisterNodeGroup registers node group in Azure Manager.
func (m *AzureManager) RegisterNodeGroup(nodeGroup cloudprovider.NodeGroup) {
	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()

	m.nodeGroups = append(m.nodeGroups, nodeGroup)
}

func (m *AzureManager) nodeGroupRegisted(nodeGroup string) bool {
	for _, ng := range m.nodeGroups {
		if nodeGroup == ng.Id() {
			return true
		}
	}

	return false
}

// GetNodeGroupForInstance returns nodeGroup of the given Instance
func (m *AzureManager) GetNodeGroupForInstance(instance *AzureRef) (cloudprovider.NodeGroup, error) {
	glog.V(5).Infof("Looking for node group for instance: %q", instance)

	glog.V(8).Infof("Cache BEFORE: %v\n", m.nodeGroupsCache)

	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()
	if nodeGroup, found := m.nodeGroupsCache[*instance]; found {
		return nodeGroup, nil
	}

	if err := m.regenerateCache(); err != nil {
		return nil, fmt.Errorf("Error while looking for nodeGroup for instance %+v, error: %v", *instance, err)
	}

	glog.V(8).Infof("Cache AFTER: %v\n", m.nodeGroupsCache)

	if nodeGroup, found := m.nodeGroupsCache[*instance]; found {
		return nodeGroup, nil
	}

	// instance does not belong to any configured nodeGroup.
	return nil, nil
}

// GetInstanceIDs gets instanceIDs for specified instances.
func (m *AzureManager) GetInstanceIDs(instances []*AzureRef) []string {
	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()

	instanceIds := make([]string, len(instances))
	for i, instance := range instances {
		instanceIds[i] = m.instanceIDsCache[instance.Name]
	}

	return instanceIds
}

func (m *AzureManager) regenerateCache() (err error) {
	var newCache map[AzureRef]cloudprovider.NodeGroup
	var newInstanceIDsCache map[string]string

	switch m.config.VMType {
	case vmTypeVMSS:
		newCache, newInstanceIDsCache, err = m.listScaleSets()
	case vmTypeStandard:
		// TODO: FIXME
	default:
		err = fmt.Errorf("vmType %q not supported", m.config.VMType)
	}
	if err != nil {
		return err
	}

	m.nodeGroupsCache = newCache
	m.instanceIDsCache = newInstanceIDsCache
	return nil
}

func (m *AzureManager) getNodeGroupByID(id string) cloudprovider.NodeGroup {
	for _, ng := range m.nodeGroups {
		if id == ng.Id() {
			return ng
		}
	}

	return nil
}

// listScaleSets gets a list of scale sets and instanceIDs.
func (m *AzureManager) listScaleSets() (map[AzureRef]cloudprovider.NodeGroup, map[string]string, error) {
	var err error
	scaleSets := make(map[AzureRef]cloudprovider.NodeGroup)
	instanceIDs := make(map[string]string)

	for _, sset := range m.nodeGroups {
		glog.V(4).Infof("Listing Scale Set information for %s", sset.Id())

		resourceGroup := m.config.ResourceGroup
		ssInfo, err := m.virtualMachineScaleSetsClient.Get(resourceGroup, sset.Id())
		if err != nil {
			glog.Errorf("Failed to get scaleSet with name %s: %v", sset.Id(), err)
			return nil, nil, err
		}

		result, err := m.virtualMachineScaleSetVMsClient.List(resourceGroup, *ssInfo.Name, "", "", "")
		if err != nil {
			glog.Errorf("Failed to list vm for scaleSet %s: %v", *ssInfo.Name, err)
			return nil, nil, err
		}

		moreResult := (result.Value != nil && len(*result.Value) > 0)
		for moreResult {
			for _, instance := range *result.Value {
				// Convert to lower because instance.ID is in different in different API calls (e.g. GET and LIST).
				name := "azure://" + strings.ToLower(*instance.ID)
				vmID := "azure://" + strings.ToLower(*instance.VMID)
				ref := AzureRef{
					Name: name,
				}
				vmIDRef := AzureRef{
					Name: vmID,
				}
				scaleSets[ref] = sset
				scaleSets[vmIDRef] = sset
				instanceIDs[name] = *instance.InstanceID
			}

			moreResult = false
			if result.NextLink != nil {
				result, err = m.virtualMachineScaleSetVMsClient.ListNextResults(result)
				if err != nil {
					glog.Errorf("virtualMachineScaleSetVMsClient.ListNextResults failed: %v", err)
					return nil, nil, err
				}

				moreResult = (result.Value != nil && len(*result.Value) > 0)
			}
		}
	}

	return scaleSets, instanceIDs, err
}

// Cleanup closes the channel to signal the go routine to stop that is handling the cache
func (m *AzureManager) Cleanup() {
	close(m.interrupt)
}
