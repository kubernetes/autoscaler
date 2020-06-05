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

//go:generate go run azure_instance_types/gen.go

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

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config/dynamic"
	klog "k8s.io/klog/v2"
	providerazure "k8s.io/legacy-cloud-providers/azure"
	azclients "k8s.io/legacy-cloud-providers/azure/clients"
	"k8s.io/legacy-cloud-providers/azure/retry"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
)

const (
	vmTypeVMSS     = "vmss"
	vmTypeStandard = "standard"
	vmTypeAKS      = "aks"

	scaleToZeroSupportedStandard = false
	scaleToZeroSupportedVMSS     = true
	refreshInterval              = 1 * time.Minute

	// The path of deployment parameters for standard vm.
	deploymentParametersPath = "/var/lib/azure/azuredeploy.parameters.json"

	vmssTagMin                     = "min"
	vmssTagMax                     = "max"
	autoDiscovererTypeLabel        = "label"
	labelAutoDiscovererKeyMinNodes = "min"
	labelAutoDiscovererKeyMaxNodes = "max"
	metadataURL                    = "http://169.254.169.254/metadata/instance"

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
)

var validLabelAutoDiscovererKeys = strings.Join([]string{
	labelAutoDiscovererKeyMinNodes,
	labelAutoDiscovererKeyMaxNodes,
}, ", ")

// A labelAutoDiscoveryConfig specifies how to autodiscover Azure scale sets.
type labelAutoDiscoveryConfig struct {
	// Key-values to match on.
	Selector map[string]string
}

// AzureManager handles Azure communication and data caching.
type AzureManager struct {
	config   *Config
	azClient *azClient
	env      azure.Environment

	asgCache              *asgCache
	lastRefresh           time.Time
	asgAutoDiscoverySpecs []labelAutoDiscoveryConfig
	explicitlyConfigured  map[string]bool
}

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
}

// Config holds the configuration parsed from the --cloud-config flag
type Config struct {
	CloudProviderRateLimitConfig

	Cloud          string `json:"cloud" yaml:"cloud"`
	Location       string `json:"location" yaml:"location"`
	TenantID       string `json:"tenantId" yaml:"tenantId"`
	SubscriptionID string `json:"subscriptionId" yaml:"subscriptionId"`
	ResourceGroup  string `json:"resourceGroup" yaml:"resourceGroup"`
	VMType         string `json:"vmType" yaml:"vmType"`

	AADClientID                 string `json:"aadClientId" yaml:"aadClientId"`
	AADClientSecret             string `json:"aadClientSecret" yaml:"aadClientSecret"`
	AADClientCertPath           string `json:"aadClientCertPath" yaml:"aadClientCertPath"`
	AADClientCertPassword       string `json:"aadClientCertPassword" yaml:"aadClientCertPassword"`
	UseManagedIdentityExtension bool   `json:"useManagedIdentityExtension" yaml:"useManagedIdentityExtension"`
	UserAssignedIdentityID      string `json:"userAssignedIdentityID" yaml:"userAssignedIdentityID"`

	// Configs only for standard vmType (agent pools).
	Deployment           string                 `json:"deployment" yaml:"deployment"`
	DeploymentParameters map[string]interface{} `json:"deploymentParameters" yaml:"deploymentParameters"`

	//Configs only for AKS
	ClusterName string `json:"clusterName" yaml:"clusterName"`
	//Config only for AKS
	NodeResourceGroup string `json:"nodeResourceGroup" yaml:"nodeResourceGroup"`

	// VMSS metadata cache TTL in seconds, only applies for vmss type
	VmssCacheTTL int64 `json:"vmssCacheTTL" yaml:"vmssCacheTTL"`

	// number of latest deployments that will not be deleted
	MaxDeploymentsCount int64 `json:"maxDeploymentsCount" yaml:"maxDeploymentsCount"`

	// Enable exponential backoff to manage resource request retries
	CloudProviderBackoff         bool    `json:"cloudProviderBackoff,omitempty" yaml:"cloudProviderBackoff,omitempty"`
	CloudProviderBackoffRetries  int     `json:"cloudProviderBackoffRetries,omitempty" yaml:"cloudProviderBackoffRetries,omitempty"`
	CloudProviderBackoffExponent float64 `json:"cloudProviderBackoffExponent,omitempty" yaml:"cloudProviderBackoffExponent,omitempty"`
	CloudProviderBackoffDuration int     `json:"cloudProviderBackoffDuration,omitempty" yaml:"cloudProviderBackoffDuration,omitempty"`
	CloudProviderBackoffJitter   float64 `json:"cloudProviderBackoffJitter,omitempty" yaml:"cloudProviderBackoffJitter,omitempty"`
}

// InitializeCloudProviderRateLimitConfig initializes rate limit configs.
func InitializeCloudProviderRateLimitConfig(config *CloudProviderRateLimitConfig) error {
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

func (cfg *Config) getAzureClientConfig(servicePrincipalToken *adal.ServicePrincipalToken, env *azure.Environment) *azclients.ClientConfig {
	azClientConfig := &azclients.ClientConfig{
		Location:                cfg.Location,
		SubscriptionID:          cfg.SubscriptionID,
		ResourceManagerEndpoint: env.ResourceManagerEndpoint,
		Authorizer:              autorest.NewBearerAuthorizer(servicePrincipalToken),
		Backoff:                 &retry.Backoff{Steps: 1},
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
	cfg.ResourceGroup = strings.TrimSpace(cfg.ResourceGroup)
	cfg.VMType = strings.TrimSpace(cfg.VMType)
	cfg.AADClientID = strings.TrimSpace(cfg.AADClientID)
	cfg.AADClientSecret = strings.TrimSpace(cfg.AADClientSecret)
	cfg.AADClientCertPath = strings.TrimSpace(cfg.AADClientCertPath)
	cfg.AADClientCertPassword = strings.TrimSpace(cfg.AADClientCertPassword)
	cfg.Deployment = strings.TrimSpace(cfg.Deployment)
	cfg.ClusterName = strings.TrimSpace(cfg.ClusterName)
	cfg.NodeResourceGroup = strings.TrimSpace(cfg.NodeResourceGroup)
}

// CreateAzureManager creates Azure Manager object to work with Azure.
func CreateAzureManager(configReader io.Reader, discoveryOpts cloudprovider.NodeGroupDiscoveryOptions) (*AzureManager, error) {
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
		cfg.AADClientID = os.Getenv("ARM_CLIENT_ID")
		cfg.AADClientSecret = os.Getenv("ARM_CLIENT_SECRET")
		cfg.VMType = strings.ToLower(os.Getenv("ARM_VM_TYPE"))
		cfg.AADClientCertPath = os.Getenv("ARM_CLIENT_CERT_PATH")
		cfg.AADClientCertPassword = os.Getenv("ARM_CLIENT_CERT_PASSWORD")
		cfg.Deployment = os.Getenv("ARM_DEPLOYMENT")
		cfg.ClusterName = os.Getenv("AZURE_CLUSTER_NAME")
		cfg.NodeResourceGroup = os.Getenv("AZURE_NODE_RESOURCE_GROUP")

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
	cfg.TrimSpace()

	if cloudProviderRateLimit := os.Getenv("CLOUD_PROVIDER_RATE_LIMIT"); cloudProviderRateLimit != "" {
		cfg.CloudProviderRateLimit, err = strconv.ParseBool(cloudProviderRateLimit)
		if err != nil {
			return nil, fmt.Errorf("failed to parse CLOUD_PROVIDER_RATE_LIMIT: %q, %v", cloudProviderRateLimit, err)
		}
	}

	err = InitializeCloudProviderRateLimitConfig(&cfg.CloudProviderRateLimitConfig)
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

	// Defaulting env to Azure Public Cloud.
	env := azure.PublicCloud
	if cfg.Cloud != "" {
		env, err = azure.EnvironmentFromName(cfg.Cloud)
		if err != nil {
			return nil, err
		}
	}

	if err := validateConfig(cfg); err != nil {
		return nil, err
	}

	klog.Infof("Starting azure manager with subscription ID %q", cfg.SubscriptionID)

	azClient, err := newAzClient(cfg, &env)
	if err != nil {
		return nil, err
	}

	// Create azure manager.
	manager := &AzureManager{
		config:               cfg,
		env:                  env,
		azClient:             azClient,
		explicitlyConfigured: make(map[string]bool),
	}

	cache, err := newAsgCache()
	if err != nil {
		return nil, err
	}
	manager.asgCache = cache

	specs, err := parseLabelAutoDiscoverySpecs(discoveryOpts)
	if err != nil {
		return nil, err
	}
	manager.asgAutoDiscoverySpecs = specs

	if err := manager.fetchExplicitAsgs(discoveryOpts.NodeGroupSpecs); err != nil {
		return nil, err
	}

	if err := manager.forceRefresh(); err != nil {
		return nil, err
	}

	return manager, nil
}

func (m *AzureManager) fetchExplicitAsgs(specs []string) error {
	changed := false
	for _, spec := range specs {
		asg, err := m.buildAsgFromSpec(spec)
		if err != nil {
			return fmt.Errorf("failed to parse node group spec: %v", err)
		}
		if m.RegisterAsg(asg) {
			changed = true
		}
		m.explicitlyConfigured[asg.Id()] = true
	}

	if changed {
		if err := m.regenerateCache(); err != nil {
			return err
		}
	}
	return nil
}

func (m *AzureManager) buildAsgFromSpec(spec string) (cloudprovider.NodeGroup, error) {
	scaleToZeroSupported := scaleToZeroSupportedStandard
	if strings.EqualFold(m.config.VMType, vmTypeVMSS) {
		scaleToZeroSupported = scaleToZeroSupportedVMSS
	}
	s, err := dynamic.SpecFromString(spec, scaleToZeroSupported)
	if err != nil {
		return nil, fmt.Errorf("failed to parse node group spec: %v", err)
	}

	switch m.config.VMType {
	case vmTypeStandard:
		return NewAgentPool(s, m)
	case vmTypeVMSS:
		return NewScaleSet(s, m)
	case vmTypeAKS:
		return NewAKSAgentPool(s, m)
	default:
		return nil, fmt.Errorf("vmtype %s not supported", m.config.VMType)
	}
}

// Refresh is called before every main loop and can be used to dynamically update cloud provider state.
// In particular the list of node groups returned by NodeGroups can change as a result of CloudProvider.Refresh().
func (m *AzureManager) Refresh() error {
	if m.lastRefresh.Add(refreshInterval).After(time.Now()) {
		return nil
	}
	return m.forceRefresh()
}

func (m *AzureManager) forceRefresh() error {
	// TODO: Refactor some of this logic out of forceRefresh and
	// consider merging the list call with the Nodes() call
	if err := m.fetchAutoAsgs(); err != nil {
		klog.Errorf("Failed to fetch ASGs: %v", err)
	}
	if err := m.regenerateCache(); err != nil {
		klog.Errorf("Failed to regenerate ASG cache: %v", err)
		return err
	}
	m.lastRefresh = time.Now()
	klog.V(2).Infof("Refreshed ASG list, next refresh after %v", m.lastRefresh.Add(refreshInterval))
	return nil
}

// Fetch automatically discovered ASGs. These ASGs should be unregistered if
// they no longer exist in Azure.
func (m *AzureManager) fetchAutoAsgs() error {
	groups, err := m.getFilteredAutoscalingGroups(m.asgAutoDiscoverySpecs)
	if err != nil {
		return fmt.Errorf("cannot autodiscover ASGs: %s", err)
	}

	changed := false
	exists := make(map[string]bool)
	for _, asg := range groups {
		asgID := asg.Id()
		exists[asgID] = true
		if m.explicitlyConfigured[asgID] {
			// This ASG was explicitly configured, but would also be
			// autodiscovered. We want the explicitly configured min and max
			// nodes to take precedence.
			klog.V(3).Infof("Ignoring explicitly configured ASG %s for autodiscovery.", asg.Id())
			continue
		}
		if m.RegisterAsg(asg) {
			klog.V(3).Infof("Autodiscovered ASG %s using tags %v", asg.Id(), m.asgAutoDiscoverySpecs)
			changed = true
		}
	}

	for _, asg := range m.getAsgs() {
		asgID := asg.Id()
		if !exists[asgID] && !m.explicitlyConfigured[asgID] {
			m.UnregisterAsg(asg)
			changed = true
		}
	}

	if changed {
		if err := m.regenerateCache(); err != nil {
			return err
		}
	}

	return nil
}

func (m *AzureManager) getAsgs() []cloudprovider.NodeGroup {
	return m.asgCache.get()
}

// RegisterAsg registers an ASG.
func (m *AzureManager) RegisterAsg(asg cloudprovider.NodeGroup) bool {
	return m.asgCache.Register(asg)
}

// UnregisterAsg unregisters an ASG.
func (m *AzureManager) UnregisterAsg(asg cloudprovider.NodeGroup) bool {
	return m.asgCache.Unregister(asg)
}

// GetAsgForInstance returns AsgConfig of the given Instance
func (m *AzureManager) GetAsgForInstance(instance *azureRef) (cloudprovider.NodeGroup, error) {
	return m.asgCache.FindForInstance(instance, m.config.VMType)
}

func (m *AzureManager) regenerateCache() error {
	m.asgCache.mutex.Lock()
	defer m.asgCache.mutex.Unlock()
	return m.asgCache.regenerate()
}

// Cleanup the ASG cache.
func (m *AzureManager) Cleanup() {
	m.asgCache.Cleanup()
}

func (m *AzureManager) getFilteredAutoscalingGroups(filter []labelAutoDiscoveryConfig) (asgs []cloudprovider.NodeGroup, err error) {
	if len(filter) == 0 {
		return nil, nil
	}

	switch m.config.VMType {
	case vmTypeVMSS:
		asgs, err = m.listScaleSets(filter)
	case vmTypeStandard:
		asgs, err = m.listAgentPools(filter)
	case vmTypeAKS:
		return nil, nil
	default:
		err = fmt.Errorf("vmType %q not supported", m.config.VMType)
	}
	if err != nil {
		return nil, err
	}

	return asgs, nil
}

// listScaleSets gets a list of scale sets and instanceIDs.
func (m *AzureManager) listScaleSets(filter []labelAutoDiscoveryConfig) ([]cloudprovider.NodeGroup, error) {
	ctx, cancel := getContextWithCancel()
	defer cancel()

	result, rerr := m.azClient.virtualMachineScaleSetsClient.List(ctx, m.config.ResourceGroup)
	if rerr != nil {
		klog.Errorf("VirtualMachineScaleSetsClient.List for %v failed: %v", m.config.ResourceGroup, rerr)
		return nil, rerr.Error()
	}

	var asgs []cloudprovider.NodeGroup
	for _, scaleSet := range result {
		if len(filter) > 0 {
			if scaleSet.Tags == nil || len(scaleSet.Tags) == 0 {
				continue
			}

			if !matchDiscoveryConfig(scaleSet.Tags, filter) {
				continue
			}
		}
		spec := &dynamic.NodeGroupSpec{
			Name:               *scaleSet.Name,
			MinSize:            1,
			MaxSize:            -1,
			SupportScaleToZero: scaleToZeroSupportedVMSS,
		}

		if val, ok := scaleSet.Tags["min"]; ok {
			if minSize, err := strconv.Atoi(*val); err == nil {
				spec.MinSize = minSize
			} else {
				return asgs, fmt.Errorf("invalid minimum size specified for vmss: %s", err)
			}
		} else {
			return asgs, fmt.Errorf("no minimum size specified for vmss: %s", *scaleSet.Name)
		}
		if spec.MinSize < 0 {
			return asgs, fmt.Errorf("minimum size must be a non-negative number of nodes")
		}
		if val, ok := scaleSet.Tags["max"]; ok {
			if maxSize, err := strconv.Atoi(*val); err == nil {
				spec.MaxSize = maxSize
			} else {
				return asgs, fmt.Errorf("invalid maximum size specified for vmss: %s", err)
			}
		} else {
			return asgs, fmt.Errorf("no maximum size specified for vmss: %s", *scaleSet.Name)
		}
		if spec.MaxSize < spec.MinSize {
			return asgs, fmt.Errorf("maximum size must be greater than minimum size")
		}

		asg, _ := NewScaleSet(spec, m)
		asgs = append(asgs, asg)
	}

	return asgs, nil
}

// listAgentPools gets a list of agent pools and instanceIDs.
// Note: filter won't take effect for agent pools.
func (m *AzureManager) listAgentPools(filter []labelAutoDiscoveryConfig) (asgs []cloudprovider.NodeGroup, err error) {
	ctx, cancel := getContextWithCancel()
	defer cancel()
	deploy, err := m.azClient.deploymentsClient.Get(ctx, m.config.ResourceGroup, m.config.Deployment)
	if err != nil {
		klog.Errorf("deploymentsClient.Get(%s, %s) failed: %v", m.config.ResourceGroup, m.config.Deployment, err)
		return nil, err
	}

	parameters := deploy.Properties.Parameters.(map[string]interface{})
	for k := range parameters {
		if k == "masterVMSize" || !strings.HasSuffix(k, "VMSize") {
			continue
		}

		poolName := strings.TrimRight(k, "VMSize")
		spec := &dynamic.NodeGroupSpec{
			Name:               poolName,
			MinSize:            1,
			MaxSize:            -1,
			SupportScaleToZero: scaleToZeroSupportedStandard,
		}
		asg, _ := NewAgentPool(spec, m)
		asgs = append(asgs, asg)
	}

	return asgs, nil
}

// ParseLabelAutoDiscoverySpecs returns any provided NodeGroupAutoDiscoverySpecs
// parsed into configuration appropriate for ASG autodiscovery.
func parseLabelAutoDiscoverySpecs(o cloudprovider.NodeGroupDiscoveryOptions) ([]labelAutoDiscoveryConfig, error) {
	cfgs := make([]labelAutoDiscoveryConfig, len(o.NodeGroupAutoDiscoverySpecs))
	var err error
	for i, spec := range o.NodeGroupAutoDiscoverySpecs {
		cfgs[i], err = parseLabelAutoDiscoverySpec(spec)
		if err != nil {
			return nil, err
		}
	}
	return cfgs, nil
}

// parseLabelAutoDiscoverySpec parses a single spec and returns the corredponding node group spec.
func parseLabelAutoDiscoverySpec(spec string) (labelAutoDiscoveryConfig, error) {
	cfg := labelAutoDiscoveryConfig{
		Selector: make(map[string]string),
	}

	tokens := strings.Split(spec, ":")
	if len(tokens) != 2 {
		return cfg, fmt.Errorf("spec \"%s\" should be discoverer:key=value,key=value", spec)
	}
	discoverer := tokens[0]
	if discoverer != autoDiscovererTypeLabel {
		return cfg, fmt.Errorf("unsupported discoverer specified: %s", discoverer)
	}

	for _, arg := range strings.Split(tokens[1], ",") {
		kv := strings.Split(arg, "=")
		if len(kv) != 2 {
			return cfg, fmt.Errorf("invalid key=value pair %s", kv)
		}
		k, v := kv[0], kv[1]
		if k == "" || v == "" {
			return cfg, fmt.Errorf("empty value not allowed in key=value tag pairs")
		}
		cfg.Selector[k] = v
	}
	return cfg, nil
}

// getSubscriptionId reads the Subscription ID from the instance metadata.
func getSubscriptionIdFromInstanceMetadata() (string, error) {
	subscriptionID, present := os.LookupEnv("ARM_SUBSCRIPTION_ID")
	if !present {
		metadataService, err := providerazure.NewInstanceMetadataService(metadataURL)
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
