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
	"errors"
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
	providerazure "sigs.k8s.io/cloud-provider-azure/pkg/provider"
	"sigs.k8s.io/cloud-provider-azure/pkg/retry"
)

const (
	// The path of deployment parameters for standard vm.
	deploymentParametersPath = "/var/lib/azure/azuredeploy.parameters.json"

	imdsServerURL = "http://169.254.169.254"

	// backoff
	backoffRetriesDefault  = 6
	backoffExponentDefault = 1.5
	backoffDurationDefault = 5 // in seconds
	backoffJitterDefault   = 1.0

	// rate limit
	rateLimitQPSDefault         float32 = 1.0
	rateLimitBucketDefault              = 5
	rateLimitReadQPSEnvVar              = "RATE_LIMIT_READ_QPS"
	rateLimitReadBucketsEnvVar          = "RATE_LIMIT_READ_BUCKETS"
	rateLimitWriteQPSEnvVar             = "RATE_LIMIT_WRITE_QPS"
	rateLimitWriteBucketsEnvVar         = "RATE_LIMIT_WRITE_BUCKETS"

	// VmssSizeRefreshPeriodDefault in seconds
	VmssSizeRefreshPeriodDefault = 30

	// auth methods
	authMethodPrincipal = "principal"
	authMethodCLI       = "cli"

	// toggle
	dynamicInstanceListDefault = false
	enableVmssFlexDefault      = false
	enableForceDeleteDefault   = false
)

// CloudProviderRateLimitConfig indicates the rate limit config for each clients.
type CloudProviderRateLimitConfig struct {
	// The default rate limit config options.
	azclients.RateLimitConfig

	// Rate limit config for each clients. Values would override default settings above.
	InterfaceRateLimit              *azclients.RateLimitConfig `json:"interfaceRateLimit,omitempty" yaml:"interfaceRateLimit,omitempty"`
	VirtualMachineRateLimit         *azclients.RateLimitConfig `json:"virtualMachineRateLimit,omitempty" yaml:"virtualMachineRateLimit,omitempty"`
	StorageAccountRateLimit         *azclients.RateLimitConfig `json:"storageAccountRateLimit,omitempty" yaml:"storageAccountRateLimit,omitempty"`
	DiskRateLimit                   *azclients.RateLimitConfig `json:"diskRateLimit,omitempty" yaml:"diskRateLimit,omitempty"`
	VirtualMachineScaleSetRateLimit *azclients.RateLimitConfig `json:"virtualMachineScaleSetRateLimit,omitempty" yaml:"virtualMachineScaleSetRateLimit,omitempty"`
	KubernetesServiceRateLimit      *azclients.RateLimitConfig `json:"kubernetesServiceRateLimit,omitempty" yaml:"kubernetesServiceRateLimit,omitempty"`
}

// Config holds the configuration parsed from the --cloud-config flag
type Config struct {
	CloudProviderRateLimitConfig

	Cloud          string `json:"cloud" yaml:"cloud"`
	Location       string `json:"location" yaml:"location"`
	TenantID       string `json:"tenantId" yaml:"tenantId"`
	SubscriptionID string `json:"subscriptionId" yaml:"subscriptionId"`
	ClusterName    string `json:"clusterName" yaml:"clusterName"`
	// ResourceGroup is the MC_ resource group where the nodes are located.
	ResourceGroup string `json:"resourceGroup" yaml:"resourceGroup"`
	// ClusterResourceGroup is the resource group where the cluster is located.
	ClusterResourceGroup string `json:"clusterResourceGroup" yaml:"clusterResourceGroup"`
	VMType               string `json:"vmType" yaml:"vmType"`

	// ARMBaseURLForAPClient is the URL to use for operations for the VMs pool.
	// It can override the default public ARM endpoint for VMs pool scale operations.
	ARMBaseURLForAPClient string `json:"armBaseURLForAPClient" yaml:"armBaseURLForAPClient"`

	// AuthMethod determines how to authorize requests for the Azure
	// cloud. Valid options are "principal" (= the traditional
	// service principle approach) and "cli" (= load az command line
	// config file). The default is "principal".
	AuthMethod string `json:"authMethod" yaml:"authMethod"`

	// Settings for a service principal.

	AADClientID                  string `json:"aadClientId" yaml:"aadClientId"`
	AADClientSecret              string `json:"aadClientSecret" yaml:"aadClientSecret"`
	AADClientCertPath            string `json:"aadClientCertPath" yaml:"aadClientCertPath"`
	AADClientCertPassword        string `json:"aadClientCertPassword" yaml:"aadClientCertPassword"`
	AADFederatedTokenFile        string `json:"aadFederatedTokenFile" yaml:"aadFederatedTokenFile"`
	UseManagedIdentityExtension  bool   `json:"useManagedIdentityExtension" yaml:"useManagedIdentityExtension"`
	UseWorkloadIdentityExtension bool   `json:"useWorkloadIdentityExtension" yaml:"useWorkloadIdentityExtension"`
	UserAssignedIdentityID       string `json:"userAssignedIdentityID" yaml:"userAssignedIdentityID"`

	// Configs only for standard vmType (agent pools).
	Deployment           string                 `json:"deployment" yaml:"deployment"`
	DeploymentParameters map[string]interface{} `json:"deploymentParameters" yaml:"deploymentParameters"`

	// VMSS metadata cache TTL in seconds, only applies for vmss type
	VmssCacheTTL int64 `json:"vmssCacheTTL" yaml:"vmssCacheTTL"`

	// VMSS instances cache TTL in seconds, only applies for vmss type
	VmssVmsCacheTTL int64 `json:"vmssVmsCacheTTL" yaml:"vmssVmsCacheTTL"`

	// Jitter in seconds subtracted from the VMSS cache TTL before the first refresh
	VmssVmsCacheJitter int `json:"vmssVmsCacheJitter" yaml:"vmssVmsCacheJitter"`

	// number of latest deployments that will not be deleted
	MaxDeploymentsCount int64 `json:"maxDeploymentsCount" yaml:"maxDeploymentsCount"`

	// Enable exponential backoff to manage resource request retries
	CloudProviderBackoff         bool    `json:"cloudProviderBackoff,omitempty" yaml:"cloudProviderBackoff,omitempty"`
	CloudProviderBackoffRetries  int     `json:"cloudProviderBackoffRetries,omitempty" yaml:"cloudProviderBackoffRetries,omitempty"`
	CloudProviderBackoffExponent float64 `json:"cloudProviderBackoffExponent,omitempty" yaml:"cloudProviderBackoffExponent,omitempty"`
	CloudProviderBackoffDuration int     `json:"cloudProviderBackoffDuration,omitempty" yaml:"cloudProviderBackoffDuration,omitempty"`
	CloudProviderBackoffJitter   float64 `json:"cloudProviderBackoffJitter,omitempty" yaml:"cloudProviderBackoffJitter,omitempty"`

	// EnableForceDelete defines whether to enable force deletion on the APIs
	EnableForceDelete bool `json:"enableForceDelete,omitempty" yaml:"enableForceDelete,omitempty"`

	// EnableDynamicInstanceList defines whether to enable dynamic instance workflow for instance information check
	EnableDynamicInstanceList bool `json:"enableDynamicInstanceList,omitempty" yaml:"enableDynamicInstanceList,omitempty"`

	// EnableVmssFlex defines whether to enable Vmss Flex support or not
	EnableVmssFlex bool `json:"enableVmssFlex,omitempty" yaml:"enableVmssFlex,omitempty"`

	// (DEPRECATED, DO NOT USE) EnableDetailedCSEMessage defines whether to emit error messages in the CSE error body info
	EnableDetailedCSEMessage bool `json:"enableDetailedCSEMessage,omitempty" yaml:"enableDetailedCSEMessage,omitempty"`

	// (DEPRECATED, DO NOT USE) GetVmssSizeRefreshPeriod (seconds) defines how frequently to call GET VMSS API to fetch VMSS info per nodegroup instance
	GetVmssSizeRefreshPeriod int `json:"getVmssSizeRefreshPeriod,omitempty" yaml:"getVmssSizeRefreshPeriod,omitempty"`
}

// BuildAzureConfig returns a Config object for the Azure clients
func BuildAzureConfig(configReader io.Reader) (*Config, error) {
	var err error
	cfg := &Config{}

	if configReader != nil {
		body, err := ioutil.ReadAll(configReader)
		if err != nil {
			return nil, fmt.Errorf("failed to read config: %v", err)
		}
		err = json.Unmarshal(body, cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal config body: %v", err)
		}
	} else {
		cfg.Cloud = os.Getenv("ARM_CLOUD")
		cfg.Location = os.Getenv("LOCATION")
		cfg.ResourceGroup = os.Getenv("ARM_RESOURCE_GROUP")
		cfg.TenantID = os.Getenv("ARM_TENANT_ID")
		if tenantId := os.Getenv("AZURE_TENANT_ID"); tenantId != "" {
			cfg.TenantID = tenantId
		}
		cfg.AADClientID = os.Getenv("ARM_CLIENT_ID")
		if clientId := os.Getenv("AZURE_CLIENT_ID"); clientId != "" {
			cfg.AADClientID = clientId
		}
		cfg.AADFederatedTokenFile = os.Getenv("AZURE_FEDERATED_TOKEN_FILE")
		cfg.AADClientSecret = os.Getenv("ARM_CLIENT_SECRET")
		cfg.VMType = strings.ToLower(os.Getenv("ARM_VM_TYPE"))
		cfg.AADClientCertPath = os.Getenv("ARM_CLIENT_CERT_PATH")
		cfg.AADClientCertPassword = os.Getenv("ARM_CLIENT_CERT_PASSWORD")
		cfg.Deployment = os.Getenv("ARM_DEPLOYMENT")

		subscriptionID, err := getSubscriptionIdFromInstanceMetadata()
		if err != nil {
			return nil, err
		}
		cfg.SubscriptionID = subscriptionID

		useManagedIdentityExtensionFromEnv := os.Getenv("ARM_USE_MANAGED_IDENTITY_EXTENSION")
		if len(useManagedIdentityExtensionFromEnv) > 0 {
			cfg.UseManagedIdentityExtension, err = strconv.ParseBool(useManagedIdentityExtensionFromEnv)
			if err != nil {
				return nil, err
			}
		}

		useWorkloadIdentityExtensionFromEnv := os.Getenv("ARM_USE_WORKLOAD_IDENTITY_EXTENSION")
		if len(useWorkloadIdentityExtensionFromEnv) > 0 {
			cfg.UseWorkloadIdentityExtension, err = strconv.ParseBool(useWorkloadIdentityExtensionFromEnv)
			if err != nil {
				return nil, err
			}
		}

		if cfg.UseManagedIdentityExtension && cfg.UseWorkloadIdentityExtension {
			return nil, errors.New("you can not combine both managed identity and workload identity as an authentication mechanism")
		}

		userAssignedIdentityIDFromEnv := os.Getenv("ARM_USER_ASSIGNED_IDENTITY_ID")
		if userAssignedIdentityIDFromEnv != "" {
			cfg.UserAssignedIdentityID = userAssignedIdentityIDFromEnv
		}

		if vmssCacheTTL := os.Getenv("AZURE_VMSS_CACHE_TTL"); vmssCacheTTL != "" {
			cfg.VmssCacheTTL, err = strconv.ParseInt(vmssCacheTTL, 10, 0)
			if err != nil {
				return nil, fmt.Errorf("failed to parse AZURE_VMSS_CACHE_TTL %q: %v", vmssCacheTTL, err)
			}
		}

		if vmssVmsCacheTTL := os.Getenv("AZURE_VMSS_VMS_CACHE_TTL"); vmssVmsCacheTTL != "" {
			cfg.VmssVmsCacheTTL, err = strconv.ParseInt(vmssVmsCacheTTL, 10, 0)
			if err != nil {
				return nil, fmt.Errorf("failed to parse AZURE_VMSS_VMS_CACHE_TTL %q: %v", vmssVmsCacheTTL, err)
			}
		}

		if vmssVmsCacheJitter := os.Getenv("AZURE_VMSS_VMS_CACHE_JITTER"); vmssVmsCacheJitter != "" {
			cfg.VmssVmsCacheJitter, err = strconv.Atoi(vmssVmsCacheJitter)
			if err != nil {
				return nil, fmt.Errorf("failed to parse AZURE_VMSS_VMS_CACHE_JITTER %q: %v", vmssVmsCacheJitter, err)
			}
		}

		if threshold := os.Getenv("AZURE_MAX_DEPLOYMENT_COUNT"); threshold != "" {
			cfg.MaxDeploymentsCount, err = strconv.ParseInt(threshold, 10, 0)
			if err != nil {
				return nil, fmt.Errorf("failed to parse AZURE_MAX_DEPLOYMENT_COUNT %q: %v", threshold, err)
			}
		}

		if enableBackoff := os.Getenv("ENABLE_BACKOFF"); enableBackoff != "" {
			cfg.CloudProviderBackoff, err = strconv.ParseBool(enableBackoff)
			if err != nil {
				return nil, fmt.Errorf("failed to parse ENABLE_BACKOFF %q: %v", enableBackoff, err)
			}
		}

		if enableDynamicInstanceList := os.Getenv("AZURE_ENABLE_DYNAMIC_INSTANCE_LIST"); enableDynamicInstanceList != "" {
			cfg.EnableDynamicInstanceList, err = strconv.ParseBool(enableDynamicInstanceList)
			if err != nil {
				return nil, fmt.Errorf("failed to parse AZURE_ENABLE_DYNAMIC_INSTANCE_LIST %q: %v", enableDynamicInstanceList, err)
			}
		} else {
			cfg.EnableDynamicInstanceList = dynamicInstanceListDefault
		}

		if getVmssSizeRefreshPeriod := os.Getenv("AZURE_GET_VMSS_SIZE_REFRESH_PERIOD"); getVmssSizeRefreshPeriod != "" {
			cfg.GetVmssSizeRefreshPeriod, err = strconv.Atoi(getVmssSizeRefreshPeriod)
			if err != nil {
				return nil, fmt.Errorf("failed to parse AZURE_GET_VMSS_SIZE_REFRESH_PERIOD %q: %v", getVmssSizeRefreshPeriod, err)
			}
		} else {
			cfg.GetVmssSizeRefreshPeriod = VmssSizeRefreshPeriodDefault
		}

		if enableVmssFlex := os.Getenv("AZURE_ENABLE_VMSS_FLEX"); enableVmssFlex != "" {
			cfg.EnableVmssFlex, err = strconv.ParseBool(enableVmssFlex)
			if err != nil {
				return nil, fmt.Errorf("failed to parse AZURE_ENABLE_VMSS_FLEX %q: %v", enableVmssFlex, err)
			}
		} else {
			cfg.EnableVmssFlex = enableVmssFlexDefault
		}

		if cfg.CloudProviderBackoff {
			if backoffRetries := os.Getenv("BACKOFF_RETRIES"); backoffRetries != "" {
				retries, err := strconv.ParseInt(backoffRetries, 10, 0)
				if err != nil {
					return nil, fmt.Errorf("failed to parse BACKOFF_RETRIES %q: %v", retries, err)
				}
				cfg.CloudProviderBackoffRetries = int(retries)
			} else {
				cfg.CloudProviderBackoffRetries = backoffRetriesDefault
			}

			if backoffExponent := os.Getenv("BACKOFF_EXPONENT"); backoffExponent != "" {
				cfg.CloudProviderBackoffExponent, err = strconv.ParseFloat(backoffExponent, 64)
				if err != nil {
					return nil, fmt.Errorf("failed to parse BACKOFF_EXPONENT %q: %v", backoffExponent, err)
				}
			} else {
				cfg.CloudProviderBackoffExponent = backoffExponentDefault
			}

			if backoffDuration := os.Getenv("BACKOFF_DURATION"); backoffDuration != "" {
				duration, err := strconv.ParseInt(backoffDuration, 10, 0)
				if err != nil {
					return nil, fmt.Errorf("failed to parse BACKOFF_DURATION %q: %v", backoffDuration, err)
				}
				cfg.CloudProviderBackoffDuration = int(duration)
			} else {
				cfg.CloudProviderBackoffDuration = backoffDurationDefault
			}

			if backoffJitter := os.Getenv("BACKOFF_JITTER"); backoffJitter != "" {
				cfg.CloudProviderBackoffJitter, err = strconv.ParseFloat(backoffJitter, 64)
				if err != nil {
					return nil, fmt.Errorf("failed to parse BACKOFF_JITTER %q: %v", backoffJitter, err)
				}
			} else {
				cfg.CloudProviderBackoffJitter = backoffJitterDefault
			}
		}
	}

	// always read the following from environment variables since azure.json doesn't have these fields
	cfg.ClusterName = os.Getenv("CLUSTER_NAME")
	cfg.ClusterResourceGroup = os.Getenv("ARM_CLUSTER_RESOURCE_GROUP")
	cfg.ARMBaseURLForAPClient = os.Getenv("ARM_BASE_URL_FOR_AP_CLIENT")

	cfg.TrimSpace()

	if cloudProviderRateLimit := os.Getenv("CLOUD_PROVIDER_RATE_LIMIT"); cloudProviderRateLimit != "" {
		cfg.CloudProviderRateLimit, err = strconv.ParseBool(cloudProviderRateLimit)
		if err != nil {
			return nil, fmt.Errorf("failed to parse CLOUD_PROVIDER_RATE_LIMIT: %q, %v", cloudProviderRateLimit, err)
		}
	}

	if enableForceDelete := os.Getenv("AZURE_ENABLE_FORCE_DELETE"); enableForceDelete != "" {
		cfg.EnableForceDelete, err = strconv.ParseBool(enableForceDelete)
		if err != nil {
			return nil, fmt.Errorf("failed to parse AZURE_ENABLE_FORCE_DELETE: %q, %v", enableForceDelete, err)
		}

		cfg.EnableForceDelete = enableForceDeleteDefault
	}

	err = initializeCloudProviderRateLimitConfig(&cfg.CloudProviderRateLimitConfig)
	if err != nil {
		return nil, err
	}

	// Defaulting vmType to vmss.
	if cfg.VMType == "" {
		cfg.VMType = vmTypeVMSS
	}

	// Read parameters from deploymentParametersPath if it is not set.
	if cfg.VMType == vmTypeStandard && len(cfg.DeploymentParameters) == 0 {
		parameters, err := readDeploymentParameters(deploymentParametersPath)
		if err != nil {
			klog.Errorf("readDeploymentParameters failed with error: %v", err)
			return nil, err
		}

		cfg.DeploymentParameters = parameters
	}

	if cfg.MaxDeploymentsCount == 0 {
		cfg.MaxDeploymentsCount = int64(defaultMaxDeploymentsCount)
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// initializeCloudProviderRateLimitConfig initializes rate limit configs.
func initializeCloudProviderRateLimitConfig(config *CloudProviderRateLimitConfig) error {
	if config == nil {
		return nil
	}

	// Assign read rate limit defaults if no configuration was passed in.
	if config.CloudProviderRateLimitQPS == 0 {
		if rateLimitQPSFromEnv := os.Getenv(rateLimitReadQPSEnvVar); rateLimitQPSFromEnv != "" {
			rateLimitQPS, err := strconv.ParseFloat(rateLimitQPSFromEnv, 0)
			if err != nil {
				return fmt.Errorf("failed to parse %s: %q, %v", rateLimitReadQPSEnvVar, rateLimitQPSFromEnv, err)
			}
			config.CloudProviderRateLimitQPS = float32(rateLimitQPS)
		} else {
			config.CloudProviderRateLimitQPS = rateLimitQPSDefault
		}
	}

	if config.CloudProviderRateLimitBucket == 0 {
		if rateLimitBucketFromEnv := os.Getenv(rateLimitReadBucketsEnvVar); rateLimitBucketFromEnv != "" {
			rateLimitBucket, err := strconv.ParseInt(rateLimitBucketFromEnv, 10, 0)
			if err != nil {
				return fmt.Errorf("failed to parse %s: %q, %v", rateLimitReadBucketsEnvVar, rateLimitBucketFromEnv, err)
			}
			config.CloudProviderRateLimitBucket = int(rateLimitBucket)
		} else {
			config.CloudProviderRateLimitBucket = rateLimitBucketDefault
		}
	}

	// Assign write rate limit defaults if no configuration was passed in.
	if config.CloudProviderRateLimitQPSWrite == 0 {
		if rateLimitQPSWriteFromEnv := os.Getenv(rateLimitWriteQPSEnvVar); rateLimitQPSWriteFromEnv != "" {
			rateLimitQPSWrite, err := strconv.ParseFloat(rateLimitQPSWriteFromEnv, 0)
			if err != nil {
				return fmt.Errorf("failed to parse %s: %q, %v", rateLimitWriteQPSEnvVar, rateLimitQPSWriteFromEnv, err)
			}
			config.CloudProviderRateLimitQPSWrite = float32(rateLimitQPSWrite)
		} else {
			config.CloudProviderRateLimitQPSWrite = config.CloudProviderRateLimitQPS
		}
	}

	if config.CloudProviderRateLimitBucketWrite == 0 {
		if rateLimitBucketWriteFromEnv := os.Getenv(rateLimitWriteBucketsEnvVar); rateLimitBucketWriteFromEnv != "" {
			rateLimitBucketWrite, err := strconv.ParseInt(rateLimitBucketWriteFromEnv, 10, 0)
			if err != nil {
				return fmt.Errorf("failed to parse %s: %q, %v", rateLimitWriteBucketsEnvVar, rateLimitBucketWriteFromEnv, err)
			}
			config.CloudProviderRateLimitBucketWrite = int(rateLimitBucketWrite)
		} else {
			config.CloudProviderRateLimitBucketWrite = config.CloudProviderRateLimitBucket
		}
	}

	config.InterfaceRateLimit = overrideDefaultRateLimitConfig(&config.RateLimitConfig, config.InterfaceRateLimit)
	config.VirtualMachineRateLimit = overrideDefaultRateLimitConfig(&config.RateLimitConfig, config.VirtualMachineRateLimit)
	config.StorageAccountRateLimit = overrideDefaultRateLimitConfig(&config.RateLimitConfig, config.StorageAccountRateLimit)
	config.DiskRateLimit = overrideDefaultRateLimitConfig(&config.RateLimitConfig, config.DiskRateLimit)
	config.VirtualMachineScaleSetRateLimit = overrideDefaultRateLimitConfig(&config.RateLimitConfig, config.VirtualMachineScaleSetRateLimit)
	config.KubernetesServiceRateLimit = overrideDefaultRateLimitConfig(&config.RateLimitConfig, config.KubernetesServiceRateLimit)

	return nil
}

// overrideDefaultRateLimitConfig overrides the default CloudProviderRateLimitConfig.
func overrideDefaultRateLimitConfig(defaults, config *azclients.RateLimitConfig) *azclients.RateLimitConfig {
	// If config not set, apply defaults.
	if config == nil {
		return defaults
	}

	// Remain disabled if it's set explicitly.
	if !config.CloudProviderRateLimit {
		return &azclients.RateLimitConfig{CloudProviderRateLimit: false}
	}

	// Apply default values.
	if config.CloudProviderRateLimitQPS == 0 {
		config.CloudProviderRateLimitQPS = defaults.CloudProviderRateLimitQPS
	}
	if config.CloudProviderRateLimitBucket == 0 {
		config.CloudProviderRateLimitBucket = defaults.CloudProviderRateLimitBucket
	}
	if config.CloudProviderRateLimitQPSWrite == 0 {
		config.CloudProviderRateLimitQPSWrite = defaults.CloudProviderRateLimitQPSWrite
	}
	if config.CloudProviderRateLimitBucketWrite == 0 {
		config.CloudProviderRateLimitBucketWrite = defaults.CloudProviderRateLimitBucketWrite
	}

	return config
}

func (cfg *Config) getAzureClientConfig(authorizer autorest.Authorizer, env *azure.Environment) *azclients.ClientConfig {
	pollingDelay := 30 * time.Second
	azClientConfig := &azclients.ClientConfig{
		Location:                cfg.Location,
		SubscriptionID:          cfg.SubscriptionID,
		ResourceManagerEndpoint: env.ResourceManagerEndpoint,
		Authorizer:              authorizer,
		Backoff:                 &retry.Backoff{Steps: 1},
		RestClientConfig: azclients.RestClientConfig{
			PollingDelay: &pollingDelay,
		},
	}

	if cfg.CloudProviderBackoff {
		azClientConfig.Backoff = &retry.Backoff{
			Steps:    cfg.CloudProviderBackoffRetries,
			Factor:   cfg.CloudProviderBackoffExponent,
			Duration: time.Duration(cfg.CloudProviderBackoffDuration) * time.Second,
			Jitter:   cfg.CloudProviderBackoffJitter,
		}
	}

	return azClientConfig
}

// TrimSpace removes all leading and trailing white spaces.
func (cfg *Config) TrimSpace() {
	cfg.Cloud = strings.TrimSpace(cfg.Cloud)
	cfg.Location = strings.TrimSpace(cfg.Location)
	cfg.TenantID = strings.TrimSpace(cfg.TenantID)
	cfg.SubscriptionID = strings.TrimSpace(cfg.SubscriptionID)
	cfg.ClusterName = strings.TrimSpace(cfg.ClusterName)
	cfg.ResourceGroup = strings.TrimSpace(cfg.ResourceGroup)
	cfg.ClusterResourceGroup = strings.TrimSpace(cfg.ClusterResourceGroup)
	cfg.VMType = strings.TrimSpace(cfg.VMType)
	cfg.AADClientID = strings.TrimSpace(cfg.AADClientID)
	cfg.AADClientSecret = strings.TrimSpace(cfg.AADClientSecret)
	cfg.AADClientCertPath = strings.TrimSpace(cfg.AADClientCertPath)
	cfg.AADClientCertPassword = strings.TrimSpace(cfg.AADClientCertPassword)
	cfg.Deployment = strings.TrimSpace(cfg.Deployment)
}

func (cfg *Config) validate() error {
	if cfg.ResourceGroup == "" {
		return fmt.Errorf("resource group not set")
	}

	if cfg.VMType == vmTypeStandard {
		if cfg.Deployment == "" {
			return fmt.Errorf("deployment not set")
		}

		if len(cfg.DeploymentParameters) == 0 {
			return fmt.Errorf("deploymentParameters not set")
		}
	}

	if cfg.SubscriptionID == "" {
		return fmt.Errorf("subscription ID not set")
	}

	if cfg.UseManagedIdentityExtension {
		return nil
	}

	if cfg.TenantID == "" {
		return fmt.Errorf("tenant ID not set")
	}

	switch cfg.AuthMethod {
	case "", authMethodPrincipal:
		if cfg.AADClientID == "" {
			return errors.New("ARM Client ID not set")
		}
	case authMethodCLI:
		// Nothing to check at the moment.
	default:
		return fmt.Errorf("unsupported authorization method: %s", cfg.AuthMethod)
	}

	if cfg.CloudProviderBackoff && cfg.CloudProviderBackoffRetries == 0 {
		return fmt.Errorf("Cloud provider backoff is enabled but retries are not set")
	}

	return nil
}

// getSubscriptionId reads the Subscription ID from the instance metadata.
func getSubscriptionIdFromInstanceMetadata() (string, error) {
	subscriptionID, present := os.LookupEnv("ARM_SUBSCRIPTION_ID")
	if !present {
		metadataService, err := providerazure.NewInstanceMetadataService(imdsServerURL)
		if err != nil {
			return "", err
		}

		metadata, err := metadataService.GetMetadata(0)
		if err != nil {
			return "", err
		}

		return metadata.Compute.SubscriptionID, nil
	}
	return subscriptionID, nil
}
