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

package azure

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"strconv"
	"strings"
	"time"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
	"k8s.io/klog/v2"
	azclients "sigs.k8s.io/cloud-provider-azure/pkg/azureclients"
	providerazureconsts "sigs.k8s.io/cloud-provider-azure/pkg/consts"
	providerazure "sigs.k8s.io/cloud-provider-azure/pkg/provider"
	providerazureconfig "sigs.k8s.io/cloud-provider-azure/pkg/provider/config"
	"sigs.k8s.io/cloud-provider-azure/pkg/retry"
)

const (
	// The path of deployment parameters for standard vm.
	deploymentParametersPath = "/var/lib/azure/azuredeploy.parameters.json"

	imdsServerURL = "http://169.254.169.254"

	// auth methods
	authMethodPrincipal = "principal"
	authMethodCLI       = "cli"
)

// Config holds the configuration parsed from the --cloud-config flag or the environment variables.
// Contains both general Azure cloud provider configuration (i.e., in azure.json) and CAS configurations/options specifically for Azure provider.
type Config struct {
	// Azure cloud provider configuration, which is generally shared with other Azure components.
	providerazure.Config `json:",inline" yaml:",inline"`

	// Legacy fields, which are only here for backward compatibility. To be deprecated.
	legacyConfig `json:",inline" yaml:",inline"`

	ClusterName string `json:"clusterName" yaml:"clusterName"`
	// ClusterResourceGroup is the resource group where the cluster is located.
	ClusterResourceGroup string `json:"clusterResourceGroup" yaml:"clusterResourceGroup"`

	// ARMBaseURLForAPClient is the URL to use for operations for the VMs pool.
	// It can override the default public ARM endpoint for VMs pool scale operations.
	ARMBaseURLForAPClient string `json:"armBaseURLForAPClient" yaml:"armBaseURLForAPClient"`

	// AuthMethod determines how to authorize requests for the Azure
	// cloud. Valid options are "principal" (= the traditional
	// service principle approach) and "cli" (= load az command line
	// config file). The default is "principal".
	// 08/16/2024: This field is awkward, given the existence of UseManagedIdentityExtension and UseFederatedWorkloadIdentityExtension.
	// Ideally, either it should be deprecated, or reworked to be on the same "dimension" as the two above, if not reworking those two.
	AuthMethod string `json:"authMethod" yaml:"authMethod"`

	// Configs only for standard vmType (agent pools).
	Deployment           string                 `json:"deployment" yaml:"deployment"`
	DeploymentParameters map[string]interface{} `json:"deploymentParameters" yaml:"deploymentParameters"`

	// Config only for AKS
	NodeResourceGroup string `json:"nodeResourceGroup" yaml:"nodeResourceGroup"`

	MaxDeploymentsCount int64 `json:"maxDeploymentsCount" yaml:"maxDeploymentsCount"`

	// Jitter in seconds subtracted from the VMSS cache TTL before the first refresh
	VmssVmsCacheJitter int `json:"vmssVmsCacheJitter" yaml:"vmssVmsCacheJitter"`

	// number of latest deployments that will not be deleted
	// EnableForceDelete defines whether to enable force deletion on the APIs
	EnableForceDelete bool `json:"enableForceDelete,omitempty" yaml:"enableForceDelete,omitempty"`

	// EnableVMsAgentPool defines whether to support VMs agentpool type in addition to VMSS type
	EnableVMsAgentPool bool `json:"enableVMsAgentPool,omitempty" yaml:"enableVMsAgentPool,omitempty"`

	// (DEPRECATED, DO NOT USE) EnableDynamicInstanceList defines whether to enable dynamic instance workflow for instance information check
	EnableDynamicInstanceList bool `json:"enableDynamicInstanceList,omitempty" yaml:"enableDynamicInstanceList,omitempty"`

	// (DEPRECATED, DO NOT USE) EnableDetailedCSEMessage defines whether to emit error messages in the CSE error body info
	EnableDetailedCSEMessage bool `json:"enableDetailedCSEMessage,omitempty" yaml:"enableDetailedCSEMessage,omitempty"`

	// (DEPRECATED, DO NOT USE) GetVmssSizeRefreshPeriod (seconds) defines how frequently to call GET VMSS API to fetch VMSS info per nodegroup instance
	GetVmssSizeRefreshPeriod int `json:"getVmssSizeRefreshPeriod,omitempty" yaml:"getVmssSizeRefreshPeriod,omitempty"`

	UseFederatedWorkloadIdentityExtension bool   `json:"useFederatedWorkloadIdentityExtension" yaml:"useFederatedWorkloadIdentityExtension"`
	AADFederatedTokenFile                 string `json:"aadFederatedTokenFile" yaml:"aadFederatedTokenFile"`
}

// These are only here for backward compabitility. Their equivalent exists in providerazure.Config with a different name.
type legacyConfig struct {
	// Being renamed to UseFederatedWorkloadIdentityExtension
	UseWorkloadIdentityExtension *bool `json:"useWorkloadIdentityExtension" yaml:"useWorkloadIdentityExtension"`
	// VMSS metadata cache TTL in seconds, only applies for vmss type; being renamed to VmssCacheTTLInSeconds
	VmssCacheTTL *int64 `json:"vmssCacheTTL" yaml:"vmssCacheTTL"`
	// VMSS instances cache TTL in seconds, only applies for vmss type; being renamed to VmssVirtualMachinesCacheTTLInSeconds
	VmssVmsCacheTTL *int64 `json:"vmssVmsCacheTTL" yaml:"vmssVmsCacheTTL"`
	// EnableVmssFlex defines whether to enable Vmss Flex support or not; being renamed to EnableVmssFlexNodes
	EnableVmssFlex *bool `json:"enableVmssFlex,omitempty" yaml:"enableVmssFlex,omitempty"`
}

// BuildAzureConfig returns a Config object for the Azure clients
func BuildAzureConfig(configReader io.Reader) (*Config, error) {
	var err error
	cfg := &Config{}

	// Static defaults
	cfg.EnableDynamicInstanceList = false
	cfg.EnableVmssFlexNodes = false
	cfg.EnableVMsAgentPool = false
	cfg.CloudProviderBackoffRetries = providerazureconsts.BackoffRetriesDefault
	cfg.CloudProviderBackoffExponent = providerazureconsts.BackoffExponentDefault
	cfg.CloudProviderBackoffDuration = providerazureconsts.BackoffDurationDefault
	cfg.CloudProviderBackoffJitter = providerazureconsts.BackoffJitterDefault
	cfg.VMType = providerazureconsts.VMTypeVMSS
	cfg.MaxDeploymentsCount = int64(defaultMaxDeploymentsCount)

	// Config file overrides defaults
	if configReader != nil {
		body, err := ioutil.ReadAll(configReader)
		if err != nil {
			return nil, fmt.Errorf("failed to read config: %v", err)
		}
		err = json.Unmarshal(body, cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal config body: %v", err)
		}
	}

	// Legacy config fields, take precedence if provided.
	if cfg.UseWorkloadIdentityExtension != nil {
		cfg.UseFederatedWorkloadIdentityExtension = *cfg.UseWorkloadIdentityExtension
	}
	if cfg.VmssCacheTTL != nil {
		if *cfg.VmssCacheTTL > int64(^uint32(0)) {
			return nil, fmt.Errorf("VmssCacheTTL value %d is too large", *cfg.VmssCacheTTL)
		}
		cfg.VmssCacheTTLInSeconds = int(*cfg.VmssCacheTTL)
	}
	if cfg.VmssVmsCacheTTL != nil {
		if *cfg.VmssVmsCacheTTL > int64(^uint32(0)) {
			return nil, fmt.Errorf("VmssVmsCacheTTL value %d is too large", *cfg.VmssVmsCacheTTL)
		}
		cfg.VmssVirtualMachinesCacheTTLInSeconds = int(*cfg.VmssVmsCacheTTL)
	}
	if cfg.EnableVmssFlex != nil {
		cfg.EnableVmssFlexNodes = *cfg.EnableVmssFlex
	}

	// Each of these environment variables, if provided, will override what's in the config file.
	// Note that this "retrieval from env" does not exist in cloud-provider-azure library (at the time of this comment).
	if _, err = assignFromEnvIfExists(&cfg.ClusterName, "CLUSTER_NAME"); err != nil {
		return nil, err
	}
	if _, err = assignFromEnvIfExists(&cfg.ClusterResourceGroup, "ARM_CLUSTER_RESOURCE_GROUP"); err != nil {
		return nil, err
	}
	if _, err = assignFromEnvIfExists(&cfg.ARMBaseURLForAPClient, "ARM_BASE_URL_FOR_AP_CLIENT"); err != nil {
		return nil, err
	}
	if _, err = assignFromEnvIfExists(&cfg.Cloud, "ARM_CLOUD"); err != nil {
		return nil, err
	}
	if _, err = assignFromEnvIfExists(&cfg.Location, "LOCATION"); err != nil {
		return nil, err
	}
	if _, err = assignFromEnvIfExists(&cfg.ResourceGroup, "ARM_RESOURCE_GROUP"); err != nil {
		return nil, err
	}
	if _, err = assignFromEnvIfExists(&cfg.TenantID, "ARM_TENANT_ID"); err != nil {
		return nil, err
	}
	if _, err = assignFromEnvIfExists(&cfg.TenantID, "AZURE_TENANT_ID"); err != nil { // taking precedence
		return nil, err
	}
	if _, err = assignFromEnvIfExists(&cfg.AADClientID, "ARM_CLIENT_ID"); err != nil {
		return nil, err
	}
	if _, err = assignFromEnvIfExists(&cfg.AADClientID, "AZURE_CLIENT_ID"); err != nil { // taking precedence
		return nil, err
	}
	if _, err = assignFromEnvIfExists(&cfg.AADFederatedTokenFile, "AZURE_FEDERATED_TOKEN_FILE"); err != nil {
		return nil, err
	}
	if _, err = assignFromEnvIfExists(&cfg.AADClientSecret, "ARM_CLIENT_SECRET"); err != nil {
		return nil, err
	}
	if _, err = assignFromEnvIfExists(&cfg.VMType, "ARM_VM_TYPE"); err != nil {
		return nil, err
	}
	if _, err = assignFromEnvIfExists(&cfg.AADClientCertPath, "ARM_CLIENT_CERT_PATH"); err != nil {
		return nil, err
	}
	if _, err = assignFromEnvIfExists(&cfg.AADClientCertPassword, "ARM_CLIENT_CERT_PASSWORD"); err != nil {
		return nil, err
	}
	if _, err = assignFromEnvIfExists(&cfg.Deployment, "ARM_DEPLOYMENT"); err != nil {
		return nil, err
	}
	if _, err = assignFromEnvIfExists(&cfg.SubscriptionID, "ARM_SUBSCRIPTION_ID"); err != nil {
		return nil, err
	}
	if _, err = assignBoolFromEnvIfExists(&cfg.UseManagedIdentityExtension, "ARM_USE_MANAGED_IDENTITY_EXTENSION"); err != nil {
		return nil, err
	}
	if _, err = assignBoolFromEnvIfExists(&cfg.UseFederatedWorkloadIdentityExtension, "ARM_USE_FEDERATED_WORKLOAD_IDENTITY_EXTENSION"); err != nil {
		return nil, err
	}
	if _, err = assignBoolFromEnvIfExists(&cfg.UseFederatedWorkloadIdentityExtension, "ARM_USE_WORKLOAD_IDENTITY_EXTENSION"); err != nil { // taking precedence
		return nil, err
	}
	if _, err = assignFromEnvIfExists(&cfg.UserAssignedIdentityID, "ARM_USER_ASSIGNED_IDENTITY_ID"); err != nil {
		return nil, err
	}
	if _, err = assignIntFromEnvIfExists(&cfg.VmssCacheTTLInSeconds, "AZURE_VMSS_CACHE_TTL_IN_SECONDS"); err != nil {
		return nil, err
	}
	if _, err = assignIntFromEnvIfExists(&cfg.VmssCacheTTLInSeconds, "AZURE_VMSS_CACHE_TTL"); err != nil { // taking precedence
		return nil, err
	}
	if _, err = assignIntFromEnvIfExists(&cfg.VmssVirtualMachinesCacheTTLInSeconds, "AZURE_VMSS_VMS_CACHE_TTL_IN_SECONDS"); err != nil {
		return nil, err
	}
	if _, err = assignIntFromEnvIfExists(&cfg.VmssVirtualMachinesCacheTTLInSeconds, "AZURE_VMSS_VMS_CACHE_TTL"); err != nil { // taking precedence
		return nil, err
	}
	if _, err = assignIntFromEnvIfExists(&cfg.VmssVmsCacheJitter, "AZURE_VMSS_VMS_CACHE_JITTER"); err != nil {
		return nil, err
	}
	if _, err = assignIntFromEnvIfExists(&cfg.GetVmssSizeRefreshPeriod, "AZURE_GET_VMSS_SIZE_REFRESH_PERIOD"); err != nil {
		return nil, err
	}
	if _, err = assignInt64FromEnvIfExists(&cfg.MaxDeploymentsCount, "AZURE_MAX_DEPLOYMENT_COUNT"); err != nil {
		return nil, err
	}
	if _, err = assignBoolFromEnvIfExists(&cfg.CloudProviderBackoff, "ENABLE_BACKOFF"); err != nil {
		return nil, err
	}
	if _, err = assignBoolFromEnvIfExists(&cfg.EnableForceDelete, "AZURE_ENABLE_FORCE_DELETE"); err != nil {
		return nil, err
	}
	if _, err = assignBoolFromEnvIfExists(&cfg.EnableVMsAgentPool, "AZURE_ENABLE_VMS_AGENT_POOLS"); err != nil {
		return nil, err
	}
	if _, err = assignBoolFromEnvIfExists(&cfg.EnableDynamicInstanceList, "AZURE_ENABLE_DYNAMIC_INSTANCE_LIST"); err != nil {
		return nil, err
	}
	if _, err = assignBoolFromEnvIfExists(&cfg.EnableDetailedCSEMessage, "AZURE_ENABLE_DETAILED_CSE_MESSAGE"); err != nil {
		return nil, err
	}
	if _, err = assignBoolFromEnvIfExists(&cfg.EnableVmssFlexNodes, "AZURE_ENABLE_VMSS_FLEX_NODES"); err != nil {
		return nil, err
	}
	if _, err = assignBoolFromEnvIfExists(&cfg.EnableVmssFlexNodes, "AZURE_ENABLE_VMSS_FLEX"); err != nil { // taking precedence
		return nil, err
	}
	if cfg.CloudProviderBackoff {
		if _, err = assignIntFromEnvIfExists(&cfg.CloudProviderBackoffRetries, "BACKOFF_RETRIES"); err != nil {
			return nil, err
		}
		if _, err = assignFloat64FromEnvIfExists(&cfg.CloudProviderBackoffExponent, "BACKOFF_EXPONENT"); err != nil {
			return nil, err
		}
		if _, err = assignIntFromEnvIfExists(&cfg.CloudProviderBackoffDuration, "BACKOFF_DURATION"); err != nil {
			return nil, err
		}
		if _, err = assignFloat64FromEnvIfExists(&cfg.CloudProviderBackoffJitter, "BACKOFF_JITTER"); err != nil {
			return nil, err
		}
	}
	if _, err = assignBoolFromEnvIfExists(&cfg.CloudProviderRateLimit, "CLOUD_PROVIDER_RATE_LIMIT"); err != nil {
		return nil, err
	}
	if _, err = assignFloat32FromEnvIfExists(&cfg.CloudProviderRateLimitQPS, "RATE_LIMIT_READ_QPS"); err != nil {
		return nil, err
	}
	if _, err = assignIntFromEnvIfExists(&cfg.CloudProviderRateLimitBucket, "RATE_LIMIT_READ_BUCKETS"); err != nil {
		return nil, err
	}
	if _, err = assignFloat32FromEnvIfExists(&cfg.CloudProviderRateLimitQPSWrite, "RATE_LIMIT_WRITE_QPS"); err != nil {
		return nil, err
	}
	if _, err = assignIntFromEnvIfExists(&cfg.CloudProviderRateLimitBucketWrite, "RATE_LIMIT_WRITE_BUCKETS"); err != nil {
		return nil, err
	}
	if _, err = assignFromEnvIfExists(&cfg.NodeResourceGroup, "AZURE_NODE_RESOURCE_GROUP"); err != nil {
		return nil, err
	}

	// Nonstatic defaults
	cfg.VMType = strings.ToLower(cfg.VMType)
	if cfg.MaxDeploymentsCount == 0 {
		// 0 means "use default" in this case.
		// This means, if it is valued by the config file, but explicitly set to 0 in the env, it will retreat to default.
		cfg.MaxDeploymentsCount = int64(defaultMaxDeploymentsCount)
	}
	if cfg.SubscriptionID == "" {
		metadataService, err := providerazure.NewInstanceMetadataService(imdsServerURL)
		if err != nil {
			return nil, err
		}

		metadata, err := metadataService.GetMetadata(0)
		if err != nil {
			return nil, err
		}

		cfg.SubscriptionID = metadata.Compute.SubscriptionID
	}
	if cfg.VMType == providerazureconsts.VMTypeStandard && len(cfg.DeploymentParameters) == 0 {
		// Read parameters from deploymentParametersPath if it is not set.
		parameters, err := readDeploymentParameters(deploymentParametersPath)
		if err != nil {
			klog.Errorf("readDeploymentParameters failed with error: %v", err)
			return nil, err
		}

		cfg.DeploymentParameters = parameters
	}
	providerazureconfig.InitializeCloudProviderRateLimitConfig(&cfg.CloudProviderRateLimitConfig)

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// A "fork" of az.getAzureClientConfig with BYO authorizer (e.g., for CLI auth) and custom polling delay support
func (cfg *Config) getAzureClientConfig(authorizer autorest.Authorizer, env *azure.Environment) *azclients.ClientConfig {
	pollingDelay := 30 * time.Second
	azClientConfig := &azclients.ClientConfig{
		CloudName:               cfg.Cloud,
		Location:                cfg.Location,
		SubscriptionID:          cfg.SubscriptionID,
		ResourceManagerEndpoint: env.ResourceManagerEndpoint,
		Authorizer:              authorizer,
		Backoff:                 &retry.Backoff{Steps: 1},
		RestClientConfig: azclients.RestClientConfig{
			PollingDelay: &pollingDelay,
		},
		DisableAzureStackCloud: cfg.DisableAzureStackCloud,
		UserAgent:              cfg.UserAgent,
	}

	if cfg.CloudProviderBackoff {
		azClientConfig.Backoff = &retry.Backoff{
			Steps:    cfg.CloudProviderBackoffRetries,
			Factor:   cfg.CloudProviderBackoffExponent,
			Duration: time.Duration(cfg.CloudProviderBackoffDuration) * time.Second,
			Jitter:   cfg.CloudProviderBackoffJitter,
		}
	}

	if cfg.HasExtendedLocation() {
		azClientConfig.ExtendedLocation = &azclients.ExtendedLocation{
			Name: cfg.ExtendedLocationName,
			Type: cfg.ExtendedLocationType,
		}
	}

	return azClientConfig
}

func (cfg *Config) validate() error {
	if cfg.ResourceGroup == "" {
		return fmt.Errorf("resource group not set")
	}

	if cfg.VMType == providerazureconsts.VMTypeStandard {
		if cfg.Deployment == "" {
			return fmt.Errorf("deployment not set")
		}

		if len(cfg.DeploymentParameters) == 0 {
			return fmt.Errorf("deploymentParameters not set")
		}
	}

	if cfg.VMType == vmTypeAKS {
		// Cluster name is a mandatory param to proceed.
		if cfg.ClusterName == "" {
			return fmt.Errorf("cluster name not set for type %+v", cfg.VMType)
		}
	}

	if cfg.SubscriptionID == "" {
		return fmt.Errorf("subscription ID not set")
	}

	if cfg.UseManagedIdentityExtension && cfg.UseFederatedWorkloadIdentityExtension {
		return fmt.Errorf("you can not combine both managed identity and workload identity as an authentication mechanism")
	}

	if cfg.VMType != providerazureconsts.VMTypeStandard && cfg.VMType != providerazureconsts.VMTypeVMSS && cfg.VMType != vmTypeAKS {
		return fmt.Errorf("unsupported VM type: %s", cfg.VMType)
	}

	if !cfg.UseManagedIdentityExtension && !cfg.UseFederatedWorkloadIdentityExtension {
		if cfg.TenantID == "" {
			return fmt.Errorf("tenant ID not set")
		}

		switch cfg.AuthMethod {
		case "", authMethodPrincipal:
			if cfg.AADClientID == "" {
				return fmt.Errorf("ARM Client ID not set")
			}
		case authMethodCLI:
			// Nothing to check at the moment.
		default:
			return fmt.Errorf("unsupported authorization method: %s", cfg.AuthMethod)
		}
	}

	if cfg.CloudProviderBackoff && cfg.CloudProviderBackoffRetries == 0 {
		return fmt.Errorf("Cloud provider backoff is enabled but retries are not set")
	}

	return nil
}

func assignFromEnvIfExists(assignee *string, name string) (bool, error) {
	if assignee == nil {
		return false, fmt.Errorf("assignee is nil")
	}
	if val, present := os.LookupEnv(name); present && strings.TrimSpace(val) != "" {
		*assignee = strings.TrimSpace(val)
		return true, nil
	}
	return false, nil
}

func assignBoolFromEnvIfExists(assignee *bool, name string) (bool, error) {
	if assignee == nil {
		return false, fmt.Errorf("assignee is nil")
	}
	var err error
	if val, present := os.LookupEnv(name); present && strings.TrimSpace(val) != "" {
		*assignee, err = strconv.ParseBool(val)
		if err != nil {
			return false, fmt.Errorf("failed to parse %s %q: %v", name, val, err)
		}
		return true, nil
	}
	return false, nil
}

func assignIntFromEnvIfExists(assignee *int, name string) (bool, error) {
	if assignee == nil {
		return false, fmt.Errorf("assignee is nil")
	}
	var err error
	if val, present := os.LookupEnv(name); present && strings.TrimSpace(val) != "" {
		*assignee, err = parseInt32(val, 10)
		if err != nil {
			return false, fmt.Errorf("failed to parse %s %q: %v", name, val, err)
		}
		return true, nil
	}
	return false, nil
}

func assignInt64FromEnvIfExists(assignee *int64, name string) (bool, error) {
	if assignee == nil {
		return false, fmt.Errorf("assignee is nil")
	}
	var err error
	if val, present := os.LookupEnv(name); present && strings.TrimSpace(val) != "" {
		*assignee, err = strconv.ParseInt(val, 10, 0)
		if err != nil {
			return false, fmt.Errorf("failed to parse %s %q: %v", name, val, err)
		}
		return true, nil
	}
	return false, nil
}

func assignFloat32FromEnvIfExists(assignee *float32, name string) (bool, error) {
	if assignee == nil {
		return false, fmt.Errorf("assignee is nil")
	}
	var err error
	if val, present := os.LookupEnv(name); present && strings.TrimSpace(val) != "" {
		*assignee, err = parseFloat32(val)
		if err != nil {
			return false, fmt.Errorf("failed to parse %s %q: %v", name, val, err)
		}
		return true, nil
	}
	return false, nil
}

func assignFloat64FromEnvIfExists(assignee *float64, name string) (bool, error) {
	if assignee == nil {
		return false, fmt.Errorf("assignee is nil")
	}
	var err error
	if val, present := os.LookupEnv(name); present && strings.TrimSpace(val) != "" {
		*assignee, err = strconv.ParseFloat(val, 64)
		if err != nil {
			return false, fmt.Errorf("failed to parse %s %q: %v", name, val, err)
		}
		return true, nil
	}
	return false, nil
}
