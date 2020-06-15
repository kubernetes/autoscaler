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
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Azure/go-autorest/autorest/azure"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config/dynamic"
	"k8s.io/klog"
)

const (
	vmTypeVMSS     = "vmss"
	vmTypeStandard = "standard"
	vmTypeACS      = "acs"
	vmTypeAKS      = "aks"

	scaleToZeroSupportedStandard = false
	scaleToZeroSupportedVMSS     = true
	refreshInterval              = 1 * time.Minute

	// The path of deployment parameters for standard vm.
	deploymentParametersPath = "/var/lib/azure/azuredeploy.parameters.json"
)

// AzureManager handles Azure communication and data caching.
type AzureManager struct {
	config   *Config
	azClient *azClient
	env      azure.Environment

	asgCache              *asgCache
	lastRefresh           time.Time
	asgAutoDiscoverySpecs []cloudprovider.LabelAutoDiscoveryConfig
	explicitlyConfigured  map[string]bool
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
	Deployment           string                 `json:"deployment" yaml:"deployment"`
	DeploymentParameters map[string]interface{} `json:"deploymentParameters" yaml:"deploymentParameters"`

	//Configs only for ACS/AKS
	ClusterName string `json:"clusterName" yaml:"clusterName"`
	//Config only for AKS
	NodeResourceGroup string `json:"nodeResourceGroup" yaml:"nodeResourceGroup"`

	// ASG cache TTL in seconds
	AsgCacheTTL int64 `json:"asgCacheTTL" yaml:"asgCacheTTL"`

	// VMSS metadata cache TTL in seconds, only applies for vmss type
	VmssCacheTTL int64 `json:"vmssCacheTTL" yaml:"vmssCacheTTL"`
}

// TrimSpace removes all leading and trailing white spaces.
func (c *Config) TrimSpace() {
	c.Cloud = strings.TrimSpace(c.Cloud)
	c.TenantID = strings.TrimSpace(c.TenantID)
	c.SubscriptionID = strings.TrimSpace(c.SubscriptionID)
	c.ResourceGroup = strings.TrimSpace(c.ResourceGroup)
	c.VMType = strings.TrimSpace(c.VMType)
	c.AADClientID = strings.TrimSpace(c.AADClientID)
	c.AADClientSecret = strings.TrimSpace(c.AADClientSecret)
	c.AADClientCertPath = strings.TrimSpace(c.AADClientCertPath)
	c.AADClientCertPassword = strings.TrimSpace(c.AADClientCertPassword)
	c.Deployment = strings.TrimSpace(c.Deployment)
	c.ClusterName = strings.TrimSpace(c.ClusterName)
	c.NodeResourceGroup = strings.TrimSpace(c.NodeResourceGroup)
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
		cfg.SubscriptionID = os.Getenv("ARM_SUBSCRIPTION_ID")
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

		useManagedIdentityExtensionFromEnv := os.Getenv("ARM_USE_MANAGED_IDENTITY_EXTENSION")
		if len(useManagedIdentityExtensionFromEnv) > 0 {
			cfg.UseManagedIdentityExtension, err = strconv.ParseBool(useManagedIdentityExtensionFromEnv)
			if err != nil {
				return nil, err
			}
		}

		if asgCacheTTL := os.Getenv("AZURE_ASG_CACHE_TTL"); asgCacheTTL != "" {
			cfg.AsgCacheTTL, err = strconv.ParseInt(asgCacheTTL, 10, 0)
			if err != nil {
				return nil, fmt.Errorf("failed to parse AZURE_ASG_CACHE_TTL %q: %v", asgCacheTTL, err)
			}
		}

		if vmssCacheTTL := os.Getenv("AZURE_VMSS_CACHE_TTL"); vmssCacheTTL != "" {
			cfg.VmssCacheTTL, err = strconv.ParseInt(vmssCacheTTL, 10, 0)
			if err != nil {
				return nil, fmt.Errorf("failed to parse AZURE_VMSS_CACHE_TTL %q: %v", vmssCacheTTL, err)
			}
		}
	}
	cfg.TrimSpace()

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

	if cfg.AsgCacheTTL == 0 {
		cfg.AsgCacheTTL = int64(defaultAsgCacheTTL)
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

	cache, err := newAsgCache(cfg.AsgCacheTTL)
	if err != nil {
		return nil, err
	}
	manager.asgCache = cache

	specs, err := discoveryOpts.ParseLabelAutoDiscoverySpecs()
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
	case vmTypeACS:
		fallthrough
	case vmTypeAKS:
		return NewContainerServiceAgentPool(s, m)
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

func (m *AzureManager) getFilteredAutoscalingGroups(filter []cloudprovider.LabelAutoDiscoveryConfig) (asgs []cloudprovider.NodeGroup, err error) {
	if len(filter) == 0 {
		return nil, nil
	}

	switch m.config.VMType {
	case vmTypeVMSS:
		asgs, err = m.listScaleSets(filter)
	case vmTypeStandard:
		asgs, err = m.listAgentPools(filter)
	case vmTypeACS:
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
func (m *AzureManager) listScaleSets(filter []cloudprovider.LabelAutoDiscoveryConfig) (asgs []cloudprovider.NodeGroup, err error) {
	ctx, cancel := getContextWithCancel()
	defer cancel()

	result, err := m.azClient.virtualMachineScaleSetsClient.List(ctx, m.config.ResourceGroup)
	if err != nil {
		klog.Errorf("VirtualMachineScaleSetsClient.List for %v failed: %v", m.config.ResourceGroup, err)
		return nil, err
	}

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
		asg, _ := NewScaleSet(spec, m)
		asgs = append(asgs, asg)
	}

	return asgs, nil
}

// listAgentPools gets a list of agent pools and instanceIDs.
// Note: filter won't take effect for agent pools.
func (m *AzureManager) listAgentPools(filter []cloudprovider.LabelAutoDiscoveryConfig) (asgs []cloudprovider.NodeGroup, err error) {
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
