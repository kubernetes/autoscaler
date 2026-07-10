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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"

	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/utils"
)

const (
	// EnvironmentFilepathName captures the name of the environment variable containing the path to the file
	// to be used while populating the Azure Environment.
	EnvironmentFilepathName = "AZURE_ENVIRONMENT_FILEPATH"
)

// OverrideAzureCloudConfigAndEnvConfigFromMetadataService returns cloud config and environment config from url
// track2 sdk will add this one in the near future https://github.com/Azure/azure-sdk-for-go/issues/20959
// cloud and env should not be empty
// it should never return an empty config
func OverrideAzureCloudConfigAndEnvConfigFromMetadataService(endpoint, cloudName string, cloudConfig *cloud.Configuration, env *Environment) error {
	// If the ResourceManagerEndpoint is not set, we should not query the metadata service
	if endpoint == "" {
		return nil
	}

	managementEndpoint := fmt.Sprintf("%s%s", strings.TrimSuffix(endpoint, "/"), "/metadata/endpoints?api-version=2019-05-01")
	res, err := http.Get(managementEndpoint) //nolint
	if err != nil {
		return err
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	metadata := []struct {
		Name            string `json:"name"`
		ResourceManager string `json:"resourceManager,omitempty"`
		Authentication  struct {
			Audiences     []string `json:"audiences"`
			LoginEndpoint string   `json:"loginEndpoint,omitempty"`
		} `json:"authentication"`
		Suffixes struct {
			AcrLoginServer *string `json:"acrLoginServer,omitempty"`
			Storage        *string `json:"storage,omitempty"`
		} `json:"suffixes,omitempty"`
	}{}
	err = json.Unmarshal(body, &metadata)
	if err != nil {
		return err
	}

	for _, item := range metadata {
		if cloudName == "" || strings.EqualFold(item.Name, cloudName) {
			// We use the endpoint to build our config, but on ASH the config returned
			// does not contain the endpoint, and this is not accounted for. This
			// ultimately unsets it for the returned config, causing the bootstrap of
			// the provider to fail. Instead, check if the endpoint is returned, and if
			// It is not then set it.
			if item.ResourceManager == "" {
				item.ResourceManager = endpoint
			}
			cloudConfig.Services[cloud.ResourceManager] = cloud.ServiceConfiguration{
				Endpoint: item.ResourceManager,
				Audience: item.Authentication.Audiences[0],
			}
			env.ResourceManagerEndpoint = item.ResourceManager
			env.TokenAudience = item.Authentication.Audiences[0]
			if item.Authentication.LoginEndpoint != "" {
				cloudConfig.ActiveDirectoryAuthorityHost = item.Authentication.LoginEndpoint
				env.ActiveDirectoryEndpoint = item.Authentication.LoginEndpoint
			}
			if item.Suffixes.Storage != nil {
				env.StorageEndpointSuffix = *item.Suffixes.Storage
			}
			if item.Suffixes.AcrLoginServer != nil {
				env.ContainerRegistryDNSSuffix = *item.Suffixes.AcrLoginServer
			}
			return nil
		}
	}
	return nil
}

func OverrideAzureCloudConfigFromEnv(cloudName string, config *cloud.Configuration, env *Environment) error {
	if !strings.EqualFold(cloudName, utils.AzureStackCloudName) {
		return nil
	}
	envFilePath, ok := os.LookupEnv(EnvironmentFilepathName)
	if !ok {
		return nil
	}
	content, err := os.ReadFile(envFilePath)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(content, env); err != nil {
		return err
	}
	if len(env.ResourceManagerEndpoint) > 0 && len(env.TokenAudience) > 0 {
		config.Services[cloud.ResourceManager] = cloud.ServiceConfiguration{
			Endpoint: env.ResourceManagerEndpoint,
			Audience: env.TokenAudience,
		}
	}
	if len(env.ActiveDirectoryEndpoint) > 0 {
		config.ActiveDirectoryAuthorityHost = env.ActiveDirectoryEndpoint
	}
	return nil
}

func GetAzureCloudConfigAndEnvConfig(armConfig *ARMClientConfig) (cloud.Configuration, *Environment, error) {
	var cloudName string
	if armConfig != nil {
		cloudName = armConfig.Cloud
	}
	config := utils.AzureCloudConfigFromName(cloudName)
	if armConfig == nil {
		return *config, nil, nil
	}
	env := EnvironmentFromName(cloudName)

	err := OverrideAzureCloudConfigAndEnvConfigFromMetadataService(armConfig.ResourceManagerEndpoint, cloudName, config, env)
	if err != nil {
		return *config, nil, err
	}
	err = OverrideAzureCloudConfigFromEnv(cloudName, config, env)
	return *config, env, err
}

// Environment represents a set of endpoints for each of Azure's Clouds.
type Environment struct {
	Name                         string              `json:"name"`
	ManagementPortalURL          string              `json:"managementPortalURL,omitempty"`
	PublishSettingsURL           string              `json:"publishSettingsURL,omitempty"`
	ServiceManagementEndpoint    string              `json:"serviceManagementEndpoint,omitempty"`
	ResourceManagerEndpoint      string              `json:"resourceManagerEndpoint,omitempty"`
	ActiveDirectoryEndpoint      string              `json:"activeDirectoryEndpoint,omitempty"`
	GalleryEndpoint              string              `json:"galleryEndpoint,omitempty"`
	KeyVaultEndpoint             string              `json:"keyVaultEndpoint,omitempty"`
	ManagedHSMEndpoint           string              `json:"managedHSMEndpoint,omitempty"`
	GraphEndpoint                string              `json:"graphEndpoint,omitempty"`
	ServiceBusEndpoint           string              `json:"serviceBusEndpoint,omitempty"`
	BatchManagementEndpoint      string              `json:"batchManagementEndpoint,omitempty"`
	MicrosoftGraphEndpoint       string              `json:"microsoftGraphEndpoint,omitempty"`
	StorageEndpointSuffix        string              `json:"storageEndpointSuffix,omitempty"`
	CosmosDBDNSSuffix            string              `json:"cosmosDBDNSSuffix,omitempty"`
	MariaDBDNSSuffix             string              `json:"mariaDBDNSSuffix,omitempty"`
	MySQLDatabaseDNSSuffix       string              `json:"mySqlDatabaseDNSSuffix,omitempty"`
	PostgresqlDatabaseDNSSuffix  string              `json:"postgresqlDatabaseDNSSuffix,omitempty"`
	SQLDatabaseDNSSuffix         string              `json:"sqlDatabaseDNSSuffix,omitempty"`
	TrafficManagerDNSSuffix      string              `json:"trafficManagerDNSSuffix,omitempty"`
	KeyVaultDNSSuffix            string              `json:"keyVaultDNSSuffix,omitempty"`
	ManagedHSMDNSSuffix          string              `json:"managedHSMDNSSuffix,omitempty"`
	ServiceBusEndpointSuffix     string              `json:"serviceBusEndpointSuffix,omitempty"`
	ServiceManagementVMDNSSuffix string              `json:"serviceManagementVMDNSSuffix,omitempty"`
	ResourceManagerVMDNSSuffix   string              `json:"resourceManagerVMDNSSuffix,omitempty"`
	ContainerRegistryDNSSuffix   string              `json:"containerRegistryDNSSuffix,omitempty"`
	TokenAudience                string              `json:"tokenAudience,omitempty"`
	APIManagementHostNameSuffix  string              `json:"apiManagementHostNameSuffix,omitempty"`
	SynapseEndpointSuffix        string              `json:"synapseEndpointSuffix,omitempty"`
	DatalakeSuffix               string              `json:"datalakeSuffix,omitempty"`
	ResourceIdentifiers          *ResourceIdentifier `json:"resourceIdentifiers,omitempty"`
}
type ResourceIdentifier struct {
	Graph               string `json:"graph,omitempty"`
	KeyVault            string `json:"keyVault,omitempty"`
	Datalake            string `json:"datalake,omitempty"`
	Batch               string `json:"batch,omitempty"`
	OperationalInsights string `json:"operationalInsights,omitempty"`
	OSSRDBMS            string `json:"ossRDBMS,omitempty"`
	Storage             string `json:"storage,omitempty"`
	Synapse             string `json:"synapse,omitempty"`
	ServiceBus          string `json:"serviceBus,omitempty"`
	SQLDatabase         string `json:"sqlDatabase,omitempty"`
	CosmosDB            string `json:"cosmosDB,omitempty"`
	ManagedHSM          string `json:"managedHSM,omitempty"`
	MicrosoftGraph      string `json:"microsoftGraph,omitempty"`
}

var EnvironmentMapping = map[string]*Environment{
	"AZURECHINACLOUD":        ChinaCloud,
	"AZURECLOUD":             PublicCloud,
	"AZUREPUBLICCLOUD":       PublicCloud,
	"AZUREUSGOVERNMENT":      USGovernmentCloud,
	"AZUREUSGOVERNMENTCLOUD": USGovernmentCloud, //TODO: deprecate
}

const NotAvailable = "N/A" // NotAvailable is used for endpoints and resource IDs that are not available for a given cloud.

var (
	// PublicCloud is the default public Azure cloud environment
	PublicCloud = &Environment{
		Name:                         "AzurePublicCloud",
		ManagementPortalURL:          "https://manage.windowsazure.com/",
		PublishSettingsURL:           "https://manage.windowsazure.com/publishsettings/index",
		ServiceManagementEndpoint:    "https://management.core.windows.net/",
		ResourceManagerEndpoint:      "https://management.azure.com/",
		ActiveDirectoryEndpoint:      "https://login.microsoftonline.com/",
		GalleryEndpoint:              "https://gallery.azure.com/",
		KeyVaultEndpoint:             "https://vault.azure.net/",
		ManagedHSMEndpoint:           "https://managedhsm.azure.net/",
		GraphEndpoint:                "https://graph.windows.net/",
		ServiceBusEndpoint:           "https://servicebus.windows.net/",
		BatchManagementEndpoint:      "https://batch.core.windows.net/",
		MicrosoftGraphEndpoint:       "https://graph.microsoft.com/",
		StorageEndpointSuffix:        "core.windows.net",
		CosmosDBDNSSuffix:            "documents.azure.com",
		MariaDBDNSSuffix:             "mariadb.database.azure.com",
		MySQLDatabaseDNSSuffix:       "mysql.database.azure.com",
		PostgresqlDatabaseDNSSuffix:  "postgres.database.azure.com",
		SQLDatabaseDNSSuffix:         "database.windows.net",
		TrafficManagerDNSSuffix:      "trafficmanager.net",
		KeyVaultDNSSuffix:            "vault.azure.net",
		ManagedHSMDNSSuffix:          "managedhsm.azure.net",
		ServiceBusEndpointSuffix:     "servicebus.windows.net",
		ServiceManagementVMDNSSuffix: "cloudapp.net",
		ResourceManagerVMDNSSuffix:   "cloudapp.azure.com",
		ContainerRegistryDNSSuffix:   "azurecr.io",
		TokenAudience:                "https://management.azure.com/",
		APIManagementHostNameSuffix:  "azure-api.net",
		SynapseEndpointSuffix:        "dev.azuresynapse.net",
		DatalakeSuffix:               "azuredatalakestore.net",
		ResourceIdentifiers: &ResourceIdentifier{
			Graph:               "https://graph.windows.net/",
			KeyVault:            "https://vault.azure.net",
			Datalake:            "https://datalake.azure.net/",
			Batch:               "https://batch.core.windows.net/",
			OperationalInsights: "https://api.loganalytics.io",
			OSSRDBMS:            "https://ossrdbms-aad.database.windows.net",
			Storage:             "https://storage.azure.com/",
			Synapse:             "https://dev.azuresynapse.net",
			ServiceBus:          "https://servicebus.azure.net/",
			SQLDatabase:         "https://database.windows.net/",
			CosmosDB:            "https://cosmos.azure.com",
			ManagedHSM:          "https://managedhsm.azure.net",
			MicrosoftGraph:      "https://graph.microsoft.com/",
		},
	}

	// USGovernmentCloud is the cloud environment for the US Government
	USGovernmentCloud = &Environment{
		Name:                         "AzureUSGovernmentCloud",
		ManagementPortalURL:          "https://manage.windowsazure.us/",
		PublishSettingsURL:           "https://manage.windowsazure.us/publishsettings/index",
		ServiceManagementEndpoint:    "https://management.core.usgovcloudapi.net/",
		ResourceManagerEndpoint:      "https://management.usgovcloudapi.net/",
		ActiveDirectoryEndpoint:      "https://login.microsoftonline.us/",
		GalleryEndpoint:              "https://gallery.usgovcloudapi.net/",
		KeyVaultEndpoint:             "https://vault.usgovcloudapi.net/",
		ManagedHSMEndpoint:           NotAvailable,
		GraphEndpoint:                "https://graph.windows.net/",
		ServiceBusEndpoint:           "https://servicebus.usgovcloudapi.net/",
		BatchManagementEndpoint:      "https://batch.core.usgovcloudapi.net/",
		MicrosoftGraphEndpoint:       "https://graph.microsoft.us/",
		StorageEndpointSuffix:        "core.usgovcloudapi.net",
		CosmosDBDNSSuffix:            "documents.azure.us",
		MariaDBDNSSuffix:             "mariadb.database.usgovcloudapi.net",
		MySQLDatabaseDNSSuffix:       "mysql.database.usgovcloudapi.net",
		PostgresqlDatabaseDNSSuffix:  "postgres.database.usgovcloudapi.net",
		SQLDatabaseDNSSuffix:         "database.usgovcloudapi.net",
		TrafficManagerDNSSuffix:      "usgovtrafficmanager.net",
		KeyVaultDNSSuffix:            "vault.usgovcloudapi.net",
		ManagedHSMDNSSuffix:          NotAvailable,
		ServiceBusEndpointSuffix:     "servicebus.usgovcloudapi.net",
		ServiceManagementVMDNSSuffix: "usgovcloudapp.net",
		ResourceManagerVMDNSSuffix:   "cloudapp.usgovcloudapi.net",
		ContainerRegistryDNSSuffix:   "azurecr.us",
		TokenAudience:                "https://management.usgovcloudapi.net/",
		APIManagementHostNameSuffix:  "azure-api.us",
		SynapseEndpointSuffix:        "dev.azuresynapse.usgovcloudapi.net",
		DatalakeSuffix:               NotAvailable,
		ResourceIdentifiers: &ResourceIdentifier{
			Graph:               "https://graph.windows.net/",
			KeyVault:            "https://vault.usgovcloudapi.net",
			Datalake:            NotAvailable,
			Batch:               "https://batch.core.usgovcloudapi.net/",
			OperationalInsights: "https://api.loganalytics.us",
			OSSRDBMS:            "https://ossrdbms-aad.database.usgovcloudapi.net",
			Storage:             "https://storage.azure.com/",
			Synapse:             "https://dev.azuresynapse.usgovcloudapi.net",
			ServiceBus:          "https://servicebus.azure.net/",
			SQLDatabase:         "https://database.usgovcloudapi.net/",
			CosmosDB:            "https://cosmos.azure.com",
			ManagedHSM:          NotAvailable,
			MicrosoftGraph:      "https://graph.microsoft.us/",
		},
	}

	// ChinaCloud is the cloud environment operated in China
	ChinaCloud = &Environment{
		Name:                         "AzureChinaCloud",
		ManagementPortalURL:          "https://manage.chinacloudapi.com/",
		PublishSettingsURL:           "https://manage.chinacloudapi.com/publishsettings/index",
		ServiceManagementEndpoint:    "https://management.core.chinacloudapi.cn/",
		ResourceManagerEndpoint:      "https://management.chinacloudapi.cn/",
		ActiveDirectoryEndpoint:      "https://login.chinacloudapi.cn/",
		GalleryEndpoint:              "https://gallery.chinacloudapi.cn/",
		KeyVaultEndpoint:             "https://vault.azure.cn/",
		ManagedHSMEndpoint:           NotAvailable,
		GraphEndpoint:                "https://graph.chinacloudapi.cn/",
		ServiceBusEndpoint:           "https://servicebus.chinacloudapi.cn/",
		BatchManagementEndpoint:      "https://batch.chinacloudapi.cn/",
		MicrosoftGraphEndpoint:       "https://microsoftgraph.chinacloudapi.cn/",
		StorageEndpointSuffix:        "core.chinacloudapi.cn",
		CosmosDBDNSSuffix:            "documents.azure.cn",
		MariaDBDNSSuffix:             "mariadb.database.chinacloudapi.cn",
		MySQLDatabaseDNSSuffix:       "mysql.database.chinacloudapi.cn",
		PostgresqlDatabaseDNSSuffix:  "postgres.database.chinacloudapi.cn",
		SQLDatabaseDNSSuffix:         "database.chinacloudapi.cn",
		TrafficManagerDNSSuffix:      "trafficmanager.cn",
		KeyVaultDNSSuffix:            "vault.azure.cn",
		ManagedHSMDNSSuffix:          NotAvailable,
		ServiceBusEndpointSuffix:     "servicebus.chinacloudapi.cn",
		ServiceManagementVMDNSSuffix: "chinacloudapp.cn",
		ResourceManagerVMDNSSuffix:   "cloudapp.chinacloudapi.cn",
		ContainerRegistryDNSSuffix:   "azurecr.cn",
		TokenAudience:                "https://management.chinacloudapi.cn/",
		APIManagementHostNameSuffix:  "azure-api.cn",
		SynapseEndpointSuffix:        "dev.azuresynapse.azure.cn",
		DatalakeSuffix:               NotAvailable,
		ResourceIdentifiers: &ResourceIdentifier{
			Graph:               "https://graph.chinacloudapi.cn/",
			KeyVault:            "https://vault.azure.cn",
			Datalake:            NotAvailable,
			Batch:               "https://batch.chinacloudapi.cn/",
			OperationalInsights: NotAvailable,
			OSSRDBMS:            "https://ossrdbms-aad.database.chinacloudapi.cn",
			Storage:             "https://storage.azure.com/",
			Synapse:             "https://dev.azuresynapse.net",
			ServiceBus:          "https://servicebus.azure.net/",
			SQLDatabase:         "https://database.chinacloudapi.cn/",
			CosmosDB:            "https://cosmos.azure.com",
			ManagedHSM:          NotAvailable,
			MicrosoftGraph:      "https://microsoftgraph.chinacloudapi.cn",
		},
	}
)

func EnvironmentFromName(cloudName string) *Environment {
	cloudName = strings.ToUpper(strings.TrimSpace(cloudName))
	if cloudConfig, ok := EnvironmentMapping[cloudName]; ok {
		return cloudConfig
	}
	return PublicCloud
}
