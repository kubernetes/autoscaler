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
)

var EnvironmentMapping = map[string]*cloud.Configuration{
	"AZURECHINACLOUD":        &cloud.AzureChina,
	"AZURECLOUD":             &cloud.AzurePublic,
	"AZUREPUBLICCLOUD":       &cloud.AzurePublic,
	"AZUREUSGOVERNMENT":      &cloud.AzureGovernment,
	"AZUREUSGOVERNMENTCLOUD": &cloud.AzureGovernment, //TODO: deprecate
}

const (
	// EnvironmentFilepathName captures the name of the environment variable containing the path to the file
	// to be used while populating the Azure Environment.
	EnvironmentFilepathName = "AZURE_ENVIRONMENT_FILEPATH"
)

func AzureCloudConfigFromName(cloudName string) *cloud.Configuration {
	if cloudName == "" {
		return &cloud.AzurePublic
	}
	cloudName = strings.ToUpper(strings.TrimSpace(cloudName))
	if cloudConfig, ok := EnvironmentMapping[cloudName]; ok {
		return cloudConfig
	}
	return nil
}

// AzureCloudConfigFromURL returns cloud config from url
// track2 sdk will add this one in the near future https://github.com/Azure/azure-sdk-for-go/issues/20959
func AzureCloudConfigFromURL(endpoint string) (*cloud.Configuration, error) {
	managementEndpoint := fmt.Sprintf("%s%s", strings.TrimSuffix(endpoint, "/"), "/metadata/endpoints?api-version=2019-05-01")
	res, err := http.Get(managementEndpoint) //nolint
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	metadata := []struct {
		Authentication struct {
			Audiences     []string
			LoginEndpoint string
		}
		Name, ResourceManager string
	}{}
	err = json.Unmarshal(body, &metadata)
	if err != nil {
		return nil, err
	}

	if len(metadata) > 0 {
		// We use the endpoint to build our config, but on ASH the config returned
		// does not contain the endpoint, and this is not accounted for. This
		// ultimately unsets it for the returned config, causing the bootstrap of
		// the provider to fail. Instead, check if the endpoint is returned, and if
		// it is not then set it.
		if len(metadata[0].ResourceManager) == 0 {
			metadata[0].ResourceManager = endpoint
		}
		return &cloud.Configuration{
			ActiveDirectoryAuthorityHost: metadata[0].Authentication.LoginEndpoint,
			Services: map[cloud.ServiceName]cloud.ServiceConfiguration{
				cloud.ResourceManager: {
					Endpoint: metadata[0].ResourceManager,
					Audience: metadata[0].Authentication.Audiences[0],
				},
			},
		}, nil
	}
	return nil, nil
}

func AzureCloudConfigOverrideFromEnv(config *cloud.Configuration) (*cloud.Configuration, error) {
	if config == nil {
		config = &cloud.AzurePublic
	}
	envFilePath, ok := os.LookupEnv(EnvironmentFilepathName)
	if !ok {
		return config, nil
	}
	content, err := os.ReadFile(envFilePath)
	if err != nil {
		return nil, err
	}
	var envConfig Environment
	if err = json.Unmarshal(content, &envConfig); err != nil {
		return nil, err
	}
	if len(envConfig.ActiveDirectoryEndpoint) > 0 {
		config.ActiveDirectoryAuthorityHost = envConfig.ActiveDirectoryEndpoint
	}
	if len(envConfig.ResourceManagerEndpoint) > 0 && len(envConfig.TokenAudience) > 0 {
		config.Services[cloud.ResourceManager] = cloud.ServiceConfiguration{
			Endpoint: envConfig.ResourceManagerEndpoint,
			Audience: envConfig.TokenAudience,
		}
	}
	return config, nil
}

// GetAzureCloudConfig returns the cloud configuration for the given ARMClientConfig.
func GetAzureCloudConfig(armConfig *ARMClientConfig) (*cloud.Configuration, error) {
	if armConfig == nil {
		return &cloud.AzurePublic, nil
	}
	if armConfig.ResourceManagerEndpoint != "" {
		return AzureCloudConfigFromURL(armConfig.ResourceManagerEndpoint)
	}

	return AzureCloudConfigOverrideFromEnv(AzureCloudConfigFromName(armConfig.Cloud))
}

// Environment represents a set of endpoints for each of Azure's Clouds.
type Environment struct {
	Name                         string             `json:"name"`
	ManagementPortalURL          string             `json:"managementPortalURL"`
	PublishSettingsURL           string             `json:"publishSettingsURL"`
	ServiceManagementEndpoint    string             `json:"serviceManagementEndpoint"`
	ResourceManagerEndpoint      string             `json:"resourceManagerEndpoint"`
	ActiveDirectoryEndpoint      string             `json:"activeDirectoryEndpoint"`
	GalleryEndpoint              string             `json:"galleryEndpoint"`
	KeyVaultEndpoint             string             `json:"keyVaultEndpoint"`
	ManagedHSMEndpoint           string             `json:"managedHSMEndpoint"`
	GraphEndpoint                string             `json:"graphEndpoint"`
	ServiceBusEndpoint           string             `json:"serviceBusEndpoint"`
	BatchManagementEndpoint      string             `json:"batchManagementEndpoint"`
	MicrosoftGraphEndpoint       string             `json:"microsoftGraphEndpoint"`
	StorageEndpointSuffix        string             `json:"storageEndpointSuffix"`
	CosmosDBDNSSuffix            string             `json:"cosmosDBDNSSuffix"`
	MariaDBDNSSuffix             string             `json:"mariaDBDNSSuffix"`
	MySQLDatabaseDNSSuffix       string             `json:"mySqlDatabaseDNSSuffix"`
	PostgresqlDatabaseDNSSuffix  string             `json:"postgresqlDatabaseDNSSuffix"`
	SQLDatabaseDNSSuffix         string             `json:"sqlDatabaseDNSSuffix"`
	TrafficManagerDNSSuffix      string             `json:"trafficManagerDNSSuffix"`
	KeyVaultDNSSuffix            string             `json:"keyVaultDNSSuffix"`
	ManagedHSMDNSSuffix          string             `json:"managedHSMDNSSuffix"`
	ServiceBusEndpointSuffix     string             `json:"serviceBusEndpointSuffix"`
	ServiceManagementVMDNSSuffix string             `json:"serviceManagementVMDNSSuffix"`
	ResourceManagerVMDNSSuffix   string             `json:"resourceManagerVMDNSSuffix"`
	ContainerRegistryDNSSuffix   string             `json:"containerRegistryDNSSuffix"`
	TokenAudience                string             `json:"tokenAudience"`
	APIManagementHostNameSuffix  string             `json:"apiManagementHostNameSuffix"`
	SynapseEndpointSuffix        string             `json:"synapseEndpointSuffix"`
	DatalakeSuffix               string             `json:"datalakeSuffix"`
	ResourceIdentifiers          ResourceIdentifier `json:"resourceIdentifiers"`
}

// ResourceIdentifier contains a set of Azure resource IDs.
type ResourceIdentifier struct {
	Graph               string `json:"graph"`
	KeyVault            string `json:"keyVault"`
	Datalake            string `json:"datalake"`
	Batch               string `json:"batch"`
	OperationalInsights string `json:"operationalInsights"`
}
