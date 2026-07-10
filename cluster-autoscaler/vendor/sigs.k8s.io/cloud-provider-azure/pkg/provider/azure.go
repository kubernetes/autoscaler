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

package provider

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	cloudprovider "k8s.io/cloud-provider"
	cloudproviderapi "k8s.io/cloud-provider/api"
	cloudnodeutil "k8s.io/cloud-provider/node/helpers"
	nodeutil "k8s.io/component-helpers/node/util"
	"k8s.io/klog/v2"

	"sigs.k8s.io/cloud-provider-azure/pkg/azclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/configloader"
	azcache "sigs.k8s.io/cloud-provider-azure/pkg/cache"
	"sigs.k8s.io/cloud-provider-azure/pkg/consts"
	"sigs.k8s.io/cloud-provider-azure/pkg/log"
	azureconfig "sigs.k8s.io/cloud-provider-azure/pkg/provider/config"
	"sigs.k8s.io/cloud-provider-azure/pkg/provider/privatelinkservice"
	"sigs.k8s.io/cloud-provider-azure/pkg/provider/routetable"
	"sigs.k8s.io/cloud-provider-azure/pkg/provider/securitygroup"
	"sigs.k8s.io/cloud-provider-azure/pkg/provider/subnet"
	"sigs.k8s.io/cloud-provider-azure/pkg/provider/zone"
	utilsets "sigs.k8s.io/cloud-provider-azure/pkg/util/sets"
	"sigs.k8s.io/cloud-provider-azure/pkg/util/taints"
	"sigs.k8s.io/cloud-provider-azure/pkg/version"
)

var (
	// Master nodes are not added to standard load balancer by default.
	defaultExcludeMasterFromStandardLB = true
	// Outbound SNAT is enabled by default.
	defaultDisableOutboundSNAT = false
	// RouteUpdateWaitingInSeconds is 30 seconds by default.
	defaultRouteUpdateWaitingInSeconds = 30
	nodeOutOfServiceTaint              = &v1.Taint{
		Key:    v1.TaintNodeOutOfService,
		Effect: v1.TaintEffectNoExecute,
	}
	nodeShutdownTaint = &v1.Taint{
		Key:    cloudproviderapi.TaintNodeShutdown,
		Effect: v1.TaintEffectNoSchedule,
	}
)

var (
	_ cloudprovider.Interface    = (*Cloud)(nil)
	_ cloudprovider.Instances    = (*Cloud)(nil)
	_ cloudprovider.LoadBalancer = (*Cloud)(nil)
	_ cloudprovider.Routes       = (*Cloud)(nil)
	_ cloudprovider.Zones        = (*Cloud)(nil)
)

// Cloud holds the config and clients
type Cloud struct {
	azureconfig.Config
	Environment *azclient.Environment

	ComputeClientFactory    azclient.ClientFactory
	NetworkClientFactory    azclient.ClientFactory
	AuthProvider            *azclient.AuthProvider
	ResourceRequestBackoff  wait.Backoff
	Metadata                *InstanceMetadataService
	VMSet                   VMSet
	LoadBalancerBackendPool BackendPool

	// ipv6DualStack allows overriding for unit testing.  It's normally initialized from featuregates
	ipv6DualStackEnabled bool
	// Lock for access to node caches, includes nodeZones, nodeResourceGroups, and unmanagedNodes.
	nodeCachesLock sync.RWMutex
	// nodeNames holds current nodes for tracking added nodes in VM caches.
	nodeNames *utilsets.IgnoreCaseSet
	// nodeZones is a mapping from Zone to a sets.Set[string] of Node's names in the Zone
	// it is updated by the nodeInformer
	nodeZones map[string]*utilsets.IgnoreCaseSet
	// nodeResourceGroups holds nodes external resource groups
	nodeResourceGroups map[string]string
	// unmanagedNodes holds a list of nodes not managed by Azure cloud provider.
	unmanagedNodes *utilsets.IgnoreCaseSet
	// excludeLoadBalancerNodes holds a list of nodes that should be excluded from LoadBalancer.
	excludeLoadBalancerNodes   *utilsets.IgnoreCaseSet
	nodePrivateIPs             map[string]*utilsets.IgnoreCaseSet
	nodePrivateIPToNodeNameMap map[string]string
	// nodeInformerSynced is for determining if the informer has synced.
	nodeInformerSynced cache.InformerSynced

	// routeCIDRsLock holds lock for routeCIDRs cache.
	routeCIDRsLock sync.Mutex
	// routeCIDRs holds cache for route CIDRs.
	routeCIDRs map[string]string

	// regionZonesMap stores all available zones for the subscription by region
	regionZonesMap   map[string][]string
	refreshZonesLock sync.RWMutex

	KubeClient         clientset.Interface
	eventBroadcaster   record.EventBroadcaster
	eventRecorder      record.EventRecorder
	routeUpdater       batchProcessor
	backendPoolUpdater batchProcessor

	vmCache        azcache.Resource
	lbCache        azcache.Resource
	nsgRepo        securitygroup.Repository
	zoneRepo       zone.Repository
	plsRepo        privatelinkservice.Repository
	subnetRepo     subnet.Repository
	routeTableRepo routetable.Repository
	// public ip cache
	// key: [resourceGroupName]
	// Value: sync.Map of [pipName]*PublicIPAddress
	pipCache azcache.Resource
	// Add service lister to always get latest service
	serviceLister corelisters.ServiceLister
	nodeLister    corelisters.NodeLister
	// node-sync-loop routine and service-reconcile routine should not update LoadBalancer at the same time
	serviceReconcileLock sync.Mutex

	// multipleStandardLoadBalancerConfigurationsSynced make sure the `reconcileMultipleStandardLoadBalancerConfigurations`
	// runs only once every time the cloud provide restarts.
	multipleStandardLoadBalancerConfigurationsSynced bool
	// nodesWithCorrectLoadBalancerByPrimaryVMSet marks nodes that are matched with load balancers by primary vmSet.
	nodesWithCorrectLoadBalancerByPrimaryVMSet      sync.Map
	multipleStandardLoadBalancersActiveServicesLock sync.Mutex
	multipleStandardLoadBalancersActiveNodesLock    sync.Mutex
	localServiceNameToServiceInfoMap                sync.Map
	endpointSlicesCache                             sync.Map

	azureResourceLocker *AzureResourceLocker
}

// NewCloud returns a Cloud with initialized clients
func NewCloud(ctx context.Context, clientBuilder cloudprovider.ControllerClientBuilder, config *azureconfig.Config, callFromCCM bool) (cloudprovider.Interface, error) {
	az := &Cloud{
		nodeNames:                  utilsets.NewString(),
		nodeZones:                  map[string]*utilsets.IgnoreCaseSet{},
		nodeResourceGroups:         map[string]string{},
		unmanagedNodes:             utilsets.NewString(),
		routeCIDRs:                 map[string]string{},
		excludeLoadBalancerNodes:   utilsets.NewString(),
		nodePrivateIPs:             map[string]*utilsets.IgnoreCaseSet{},
		nodePrivateIPToNodeNameMap: map[string]string{},
	}

	err := az.InitializeCloudFromConfig(ctx, config, false, callFromCCM)
	if err != nil {
		return nil, err
	}

	az.ipv6DualStackEnabled = true

	if clientBuilder != nil {
		az.KubeClient = clientBuilder.ClientOrDie("azure-cloud-provider")
	}
	az.azureResourceLocker = NewAzureResourceLocker(
		az,
		consts.AzureResourceLockHolderNameCloudControllerManager,
		consts.AzureResourceLockLeaseName,
		consts.AzureResourceLockLeaseNamespace,
		consts.AzureResourceLockLeaseDuration,
	)

	return az, nil
}

func NewCloudFromConfigFile(ctx context.Context, clientBuilder cloudprovider.ControllerClientBuilder, configFilePath string, calFromCCM bool) (cloudprovider.Interface, error) {
	logger := log.FromContextOrBackground(ctx).WithName("NewCloudFromConfigFile")
	var (
		cloud cloudprovider.Interface
		err   error
	)

	var configValue *azureconfig.Config
	if configFilePath != "" {
		var configFile *os.File
		configFile, err = os.Open(configFilePath)
		if err != nil {
			logger.Error(err, "Couldn't open cloud provider configuration", "configFilePath", configFilePath)
			klog.FlushAndExit(klog.ExitFlushTimeout, 1)
		}

		defer configFile.Close()
		configValue, err = azureconfig.ParseConfig(configFile)
		if err != nil {
			logger.Error(err, "Failed to parse Azure cloud provider config")
			klog.FlushAndExit(klog.ExitFlushTimeout, 1)
		}
	}
	cloud, err = NewCloud(ctx, clientBuilder, configValue, calFromCCM && configFilePath != "")
	if err != nil {
		return nil, fmt.Errorf("could not init cloud provider azure: %w", err)
	}
	if cloud == nil {
		return nil, fmt.Errorf("nil cloud")
	}

	return cloud, nil
}

func NewCloudFromSecret(ctx context.Context, clientBuilder cloudprovider.ControllerClientBuilder, secretName, secretNamespace, cloudConfigKey string) (cloudprovider.Interface, error) {
	config, err := configloader.Load[azureconfig.Config](ctx, &configloader.K8sSecretLoaderConfig{
		K8sSecretConfig: configloader.K8sSecretConfig{
			SecretName:      secretName,
			SecretNamespace: secretNamespace,
			CloudConfigKey:  cloudConfigKey,
		},
		KubeClient: clientBuilder.ClientOrDie("cloud-provider-azure"),
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("NewCloudFromSecret: failed to get config from secret %s/%s: %w", secretNamespace, secretName, err)
	}
	az, err := NewCloud(ctx, clientBuilder, config, true)
	if err != nil {
		return nil, fmt.Errorf("NewCloudFromSecret: failed to initialize cloud from secret %s/%s: %w", secretNamespace, secretName, err)
	}
	az.Initialize(clientBuilder, wait.NeverStop)

	return az, nil
}

var (
	// newARMClientFactory is a function that returns a new ARM client factory.
	// It is used to mock the ARM client factory for testing.
	// TODO: use fake options for testing
	newARMClientFactory = azclient.NewClientFactory
)

// InitializeCloudFromConfig initializes the Cloud from config.
func (az *Cloud) InitializeCloudFromConfig(ctx context.Context, config *azureconfig.Config, _, callFromCCM bool) error {
	logger := log.FromContextOrBackground(ctx).WithName("InitializeCloudFromConfig")
	if config == nil {
		// should not reach here
		return fmt.Errorf("InitializeCloudFromConfig: cannot initialize from nil config")
	}

	if config.RouteTableResourceGroup == "" {
		config.RouteTableResourceGroup = config.ResourceGroup
	}

	if config.SecurityGroupResourceGroup == "" {
		config.SecurityGroupResourceGroup = config.ResourceGroup
	}

	if config.PrivateLinkServiceResourceGroup == "" {
		config.PrivateLinkServiceResourceGroup = config.ResourceGroup
	}

	if config.VMType == "" {
		// default to vmss vmType if not set.
		config.VMType = consts.VMTypeVMSS
	}

	if config.RouteUpdateWaitingInSeconds <= 0 {
		config.RouteUpdateWaitingInSeconds = defaultRouteUpdateWaitingInSeconds
	}

	if config.DisableAvailabilitySetNodes && config.VMType != consts.VMTypeVMSS {
		return fmt.Errorf("disableAvailabilitySetNodes %v is only supported when vmType is 'vmss'", config.DisableAvailabilitySetNodes)
	}

	if config.CloudConfigType == "" {
		// The default cloud config type is cloudConfigTypeMerge.
		config.CloudConfigType = configloader.CloudConfigTypeMerge
	} else {
		supportedCloudConfigTypes := utilsets.NewString(
			string(configloader.CloudConfigTypeMerge),
			string(configloader.CloudConfigTypeFile),
			string(configloader.CloudConfigTypeSecret))
		if !supportedCloudConfigTypes.Has(string(config.CloudConfigType)) {
			return fmt.Errorf("cloudConfigType %v is not supported, supported values are %v", config.CloudConfigType, supportedCloudConfigTypes.UnsortedList())
		}
	}

	if config.LoadBalancerBackendPoolConfigurationType == "" ||
		// TODO(nilo19): support pod IP mode in the future
		strings.EqualFold(config.LoadBalancerBackendPoolConfigurationType, consts.LoadBalancerBackendPoolConfigurationTypePODIP) {
		config.LoadBalancerBackendPoolConfigurationType = consts.LoadBalancerBackendPoolConfigurationTypeNodeIPConfiguration
	} else {
		supportedLoadBalancerBackendPoolConfigurationTypes := utilsets.NewString(
			strings.ToLower(consts.LoadBalancerBackendPoolConfigurationTypeNodeIPConfiguration),
			strings.ToLower(consts.LoadBalancerBackendPoolConfigurationTypeNodeIP),
			strings.ToLower(consts.LoadBalancerBackendPoolConfigurationTypePODIP))
		if !supportedLoadBalancerBackendPoolConfigurationTypes.Has(strings.ToLower(config.LoadBalancerBackendPoolConfigurationType)) {
			return fmt.Errorf("loadBalancerBackendPoolConfigurationType %s is not supported, supported values are %v", config.LoadBalancerBackendPoolConfigurationType, supportedLoadBalancerBackendPoolConfigurationTypes.UnsortedList())
		}
	}

	if config.ClusterServiceLoadBalancerHealthProbeMode == "" {
		config.ClusterServiceLoadBalancerHealthProbeMode = consts.ClusterServiceLoadBalancerHealthProbeModeServiceNodePort
	} else {
		supportedClusterServiceLoadBalancerHealthProbeModes := utilsets.NewString(
			strings.ToLower(consts.ClusterServiceLoadBalancerHealthProbeModeServiceNodePort),
			strings.ToLower(consts.ClusterServiceLoadBalancerHealthProbeModeShared),
		)
		if !supportedClusterServiceLoadBalancerHealthProbeModes.Has(strings.ToLower(config.ClusterServiceLoadBalancerHealthProbeMode)) {
			return fmt.Errorf("clusterServiceLoadBalancerHealthProbeMode %s is not supported, supported values are %v", config.ClusterServiceLoadBalancerHealthProbeMode, supportedClusterServiceLoadBalancerHealthProbeModes.UnsortedList())
		}
	}
	if config.ClusterServiceSharedLoadBalancerHealthProbePort == 0 {
		config.ClusterServiceSharedLoadBalancerHealthProbePort = consts.ClusterServiceLoadBalancerHealthProbeDefaultPort
	}
	if config.ClusterServiceSharedLoadBalancerHealthProbePath == "" {
		config.ClusterServiceSharedLoadBalancerHealthProbePath = consts.ClusterServiceLoadBalancerHealthProbeDefaultPath
	}

	clientOps, env, err := azclient.GetAzCoreClientOption(&config.ARMClientConfig)
	if err != nil {
		return err
	}
	resourceRequestBackoff := az.setCloudProviderBackoffDefaults(config)

	err = az.setLBDefaults(config)
	if err != nil {
		return err
	}

	az.Config = *config
	az.Environment = env
	az.ResourceRequestBackoff = resourceRequestBackoff
	az.Metadata, err = NewInstanceMetadataService(consts.ImdsServer)
	if err != nil {
		return err
	}

	if az.MaximumLoadBalancerRuleCount == 0 {
		az.MaximumLoadBalancerRuleCount = consts.MaximumLoadBalancerRuleCount
	}

	if strings.EqualFold(consts.VMTypeVMSS, az.VMType) {
		az.VMSet, err = newScaleSet(az)
		if err != nil {
			return err
		}
	} else if strings.EqualFold(consts.VMTypeVmssFlex, az.VMType) {
		az.VMSet, err = newFlexScaleSet(az)
		if err != nil {
			return err
		}
	} else {
		az.VMSet, err = newAvailabilitySet(az)
		if err != nil {
			return err
		}
	}

	if az.IsLBBackendPoolTypeNodeIPConfig() {
		az.LoadBalancerBackendPool = newBackendPoolTypeNodeIPConfig(az)
	} else if az.IsLBBackendPoolTypeNodeIP() {
		az.LoadBalancerBackendPool = newBackendPoolTypeNodeIP(az)
	}

	if az.UseMultipleStandardLoadBalancers() {
		if err := az.checkEnableMultipleStandardLoadBalancers(); err != nil {
			return err
		}
	}

	if az.AuthProvider == nil {
		var authProvider *azclient.AuthProvider
		authProvider, err = azclient.NewAuthProvider(&az.ARMClientConfig, &az.AzureAuthConfig)
		if err != nil {
			return err
		}
		az.AuthProvider = authProvider
	}
	if az.AuthProvider.GetAzIdentity() == nil {
		// No credentials provided, useInstanceMetadata should be enabled for Kubelet.
		// TODO(feiskyer): print different error message for Kubelet and controller-manager, as they're
		// requiring different credential settings.
		if !config.UseInstanceMetadata && config.CloudConfigType == configloader.CloudConfigTypeFile {
			return fmt.Errorf("useInstanceMetadata must be enabled without Azure credentials")
		}

		logger.V(2).Info("Azure cloud provider is starting without credentials")
	}

	if az.UserAgent == "" {
		k8sVersion := version.Get().GitVersion
		az.UserAgent = fmt.Sprintf("kubernetes-cloudprovider/%s", k8sVersion)
	}

	if az.ComputeClientFactory == nil && az.AuthProvider != nil {
		var (
			computeCred = az.AuthProvider.GetAzIdentity()
			networkCred = az.AuthProvider.GetNetworkAzIdentity() // It would fallback to compute credential if network credential is not set
		)

		networkSubscriptionID := az.getNetworkResourceSubscriptionID() // It would also fallback to compute subscription ID if network subscription ID is not set
		az.NetworkClientFactory, err = newARMClientFactory(&azclient.ClientFactoryConfig{
			SubscriptionID: networkSubscriptionID,
		}, &az.ARMClientConfig, clientOps.Cloud, networkCred)
		if err != nil {
			return err
		}
		logger.Info("Setting up ARM client factory for network resources", "subscriptionID", networkSubscriptionID)

		az.ComputeClientFactory, err = newARMClientFactory(&azclient.ClientFactoryConfig{
			SubscriptionID: az.SubscriptionID,
		}, &az.ARMClientConfig, clientOps.Cloud, computeCred, az.AuthProvider.AdditionalComputeClientOptions...)
		if err != nil {
			return err
		}
		logger.Info("Setting up ARM client factory for compute resources", "subscriptionID", az.SubscriptionID)
	}

	networkClientFactory := az.NetworkClientFactory

	if az.nsgRepo == nil {
		az.nsgRepo, err = securitygroup.NewSecurityGroupRepo(
			az.SecurityGroupResourceGroup,
			az.SecurityGroupName,
			az.NsgCacheTTLInSeconds,
			az.DisableAPICallCache,
			networkClientFactory.GetSecurityGroupClient(),
		)
		if err != nil {
			return err
		}
	}

	if az.zoneRepo == nil {
		az.zoneRepo, err = zone.NewRepo(az.NetworkClientFactory.GetProviderClient())
		if err != nil {
			return err
		}
	}
	if az.plsRepo == nil {
		az.plsRepo, err = privatelinkservice.NewRepo(az.ComputeClientFactory.GetPrivateLinkServiceClient(), time.Duration(az.PlsCacheTTLInSeconds)*time.Second, az.DisableAPICallCache)
		if err != nil {
			return err
		}
	}

	if az.subnetRepo == nil {
		az.subnetRepo, err = subnet.NewRepo(networkClientFactory.GetSubnetClient())
		if err != nil {
			return err
		}
	}

	if az.routeTableRepo == nil {
		az.routeTableRepo, err = routetable.NewRepo(networkClientFactory.GetRouteTableClient(), az.RouteTableResourceGroup, time.Duration(az.RouteTableCacheTTLInSeconds)*time.Second, az.DisableAPICallCache)
		if err != nil {
			return err
		}
	}

	err = az.initCaches()
	if err != nil {
		return err
	}
	// updating routes and syncing zones only in CCM
	if callFromCCM {
		// start delayed route updater.
		if az.RouteUpdateIntervalInSeconds == 0 {
			az.RouteUpdateIntervalInSeconds = consts.DefaultRouteUpdateIntervalInSeconds
		}
		az.routeUpdater = newDelayedRouteUpdater(az, time.Duration(az.RouteUpdateIntervalInSeconds)*time.Second)
		go az.routeUpdater.run(ctx)

		// start backend pool updater.
		if az.UseMultipleStandardLoadBalancers() {
			az.backendPoolUpdater = newLoadBalancerBackendPoolUpdater(az, time.Duration(az.LoadBalancerBackendPoolUpdateIntervalInSeconds)*time.Second)
			go az.backendPoolUpdater.run(ctx)
		}

		// Azure Stack does not support zone at the moment
		// https://docs.microsoft.com/en-us/azure-stack/user/azure-stack-network-differences?view=azs-2102
		if !az.IsStackCloud() {
			// wait for the success first time of syncing zones
			err = az.syncRegionZonesMap(ctx)
			if err != nil {
				logger.Error(err, "Failed to sync regional zones map for the first time")
				return err
			}

			go az.refreshZones(ctx, az.syncRegionZonesMap)
		}
	}

	return nil
}

// Multiple standard load balancer mode only supports IP-based load balancers.
func (az *Cloud) checkEnableMultipleStandardLoadBalancers() error {
	if az.IsLBBackendPoolTypeNodeIPConfig() {
		return fmt.Errorf("multiple standard load balancers cannot be used with backend pool type %s", consts.LoadBalancerBackendPoolConfigurationTypeNodeIPConfiguration)
	}

	names := utilsets.NewString()
	primaryVMSets := utilsets.NewString()
	for _, multiSLBConfig := range az.MultipleStandardLoadBalancerConfigurations {
		if names.Has(multiSLBConfig.Name) {
			return fmt.Errorf("duplicated multiple standard load balancer configuration name %s", multiSLBConfig.Name)
		}
		names.Insert(multiSLBConfig.Name)

		if multiSLBConfig.PrimaryVMSet == "" {
			return fmt.Errorf("multiple standard load balancer configuration %s must have primary VMSet", multiSLBConfig.Name)
		}
		if primaryVMSets.Has(multiSLBConfig.PrimaryVMSet) {
			return fmt.Errorf("duplicated primary VMSet %s in multiple standard load balancer configurations %s", multiSLBConfig.PrimaryVMSet, multiSLBConfig.Name)
		}
		primaryVMSets.Insert(multiSLBConfig.PrimaryVMSet)
	}

	if az.LoadBalancerBackendPoolUpdateIntervalInSeconds == 0 {
		az.LoadBalancerBackendPoolUpdateIntervalInSeconds = consts.DefaultLoadBalancerBackendPoolUpdateIntervalInSeconds
	}

	return nil
}

func (az *Cloud) initCaches() (err error) {
	logger := log.Background().WithName("initCaches")
	if az.DisableAPICallCache {
		logger.Info("API call cache is disabled, ignore logs about cache operations")
	}

	az.vmCache, err = az.newVMCache()
	if err != nil {
		return err
	}

	az.lbCache, err = az.newLBCache()
	if err != nil {
		return err
	}

	az.pipCache, err = az.newPIPCache()
	if err != nil {
		return err
	}

	return nil
}

func (az *Cloud) setLBDefaults(config *azureconfig.Config) error {
	if config.LoadBalancerSKU == "" {
		config.LoadBalancerSKU = consts.LoadBalancerSKUStandard
	}

	if strings.EqualFold(config.LoadBalancerSKU, consts.LoadBalancerSKUStandard) {
		// Do not add master nodes to standard LB by default.
		if config.ExcludeMasterFromStandardLB == nil {
			config.ExcludeMasterFromStandardLB = &defaultExcludeMasterFromStandardLB
		}

		// Enable outbound SNAT by default.
		if config.DisableOutboundSNAT == nil {
			config.DisableOutboundSNAT = &defaultDisableOutboundSNAT
		}
	} else {
		if config.DisableOutboundSNAT != nil && *config.DisableOutboundSNAT {
			return fmt.Errorf("disableOutboundSNAT should only set when loadBalancerSKU is standard")
		}
	}
	return nil
}

func (az *Cloud) setCloudProviderBackoffDefaults(config *azureconfig.Config) wait.Backoff {
	logger := log.Background().WithName("setCloudProviderBackoffDefaults")
	// Conditionally configure resource request backoff
	resourceRequestBackoff := wait.Backoff{
		Steps: 1,
	}
	if config.CloudProviderBackoff {
		// Assign backoff defaults if no configuration was passed in
		if config.CloudProviderBackoffRetries == 0 {
			config.CloudProviderBackoffRetries = consts.BackoffRetriesDefault
		}
		if config.CloudProviderBackoffDuration == 0 {
			config.CloudProviderBackoffDuration = consts.BackoffDurationDefault
		}
		if config.CloudProviderBackoffExponent == 0 {
			config.CloudProviderBackoffExponent = consts.BackoffExponentDefault
		}

		if config.CloudProviderBackoffJitter == 0 {
			config.CloudProviderBackoffJitter = consts.BackoffJitterDefault
		}

		resourceRequestBackoff = wait.Backoff{
			Steps:    config.CloudProviderBackoffRetries,
			Factor:   config.CloudProviderBackoffExponent,
			Duration: time.Duration(config.CloudProviderBackoffDuration) * time.Second,
			Jitter:   config.CloudProviderBackoffJitter,
		}
		logger.V(2).Info("Azure cloudprovider using try backoff",
			"retries", config.CloudProviderBackoffRetries,
			"exponent", config.CloudProviderBackoffExponent,
			"duration", config.CloudProviderBackoffDuration,
			"jitter", config.CloudProviderBackoffJitter)
	} else {
		// CloudProviderBackoffRetries will be set to 1 by default as the requirements of Azure SDK.
		config.CloudProviderBackoffRetries = 1
		config.CloudProviderBackoffDuration = consts.BackoffDurationDefault
	}
	return resourceRequestBackoff
}

// Initialize passes a Kubernetes clientBuilder interface to the cloud provider
func (az *Cloud) Initialize(clientBuilder cloudprovider.ControllerClientBuilder, _ <-chan struct{}) {
	az.KubeClient = clientBuilder.ClientOrDie("azure-cloud-provider")
	az.eventBroadcaster = record.NewBroadcaster()
	az.eventBroadcaster.StartRecordingToSink(&v1core.EventSinkImpl{Interface: az.KubeClient.CoreV1().Events("")})
	az.eventRecorder = az.eventBroadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: "azure-cloud-provider"})
}

// LoadBalancer returns a balancer interface. Also returns true if the interface is supported, false otherwise.
func (az *Cloud) LoadBalancer() (cloudprovider.LoadBalancer, bool) {
	return az, true
}

// Instances returns an instances interface. Also returns true if the interface is supported, false otherwise.
func (az *Cloud) Instances() (cloudprovider.Instances, bool) {
	return az, true
}

// InstancesV2 is an implementation for instances and should only be implemented by external cloud providers.
// Implementing InstancesV2 is behaviorally identical to Instances but is optimized to significantly reduce
// API calls to the cloud provider when registering and syncing nodes. Implementation of this interface will
// disable calls to the Zones interface. Also returns true if the interface is supported, false otherwise.
func (az *Cloud) InstancesV2() (cloudprovider.InstancesV2, bool) {
	return az, true
}

// Zones returns a zones interface. Also returns true if the interface is supported, false otherwise.
// DEPRECATED: Zones is deprecated in favor of retrieving zone/region information from InstancesV2.
// This interface will not be called if InstancesV2 is enabled.
func (az *Cloud) Zones() (cloudprovider.Zones, bool) {
	if az.IsStackCloud() {
		// Azure stack does not support zones at this point
		// https://docs.microsoft.com/en-us/azure-stack/user/azure-stack-network-differences?view=azs-2102
		return nil, false
	}
	return az, true
}

// Clusters returns a clusters interface.  Also returns true if the interface is supported, false otherwise.
func (az *Cloud) Clusters() (cloudprovider.Clusters, bool) {
	return nil, false
}

// Routes returns a routes interface along with whether the interface is supported.
func (az *Cloud) Routes() (cloudprovider.Routes, bool) {
	return az, true
}

// HasClusterID returns true if the cluster has a clusterID
func (az *Cloud) HasClusterID() bool {
	return true
}

// ProviderName returns the cloud provider ID.
func (az *Cloud) ProviderName() string {
	return consts.CloudProviderName
}

// SetInformers sets informers for Azure cloud provider.
func (az *Cloud) SetInformers(informerFactory informers.SharedInformerFactory) {
	logger := log.Background().WithName("SetInformers")
	logger.Info("Setting up informers for Azure cloud provider")
	nodeInformer := informerFactory.Core().V1().Nodes().Informer()
	_, _ = nodeInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			node := obj.(*v1.Node)
			az.updateNodeCaches(nil, node)
			az.updateNodeTaint(node)
		},
		UpdateFunc: func(prev, obj interface{}) {
			prevNode := prev.(*v1.Node)
			newNode := obj.(*v1.Node)
			az.updateNodeCaches(prevNode, newNode)
			az.updateNodeTaint(newNode)
		},
		DeleteFunc: func(obj interface{}) {
			node, isNode := obj.(*v1.Node)
			// We can get DeletedFinalStateUnknown instead of *v1.Node here
			// and we need to handle that correctly.
			if !isNode {
				deletedState, ok := obj.(cache.DeletedFinalStateUnknown)
				if !ok {
					logger.Error(nil, "Received unexpected object", "obj", obj)
					return
				}
				node, ok = deletedState.Obj.(*v1.Node)
				if !ok {
					logger.Error(nil, "DeletedFinalStateUnknown contained non-Node object", "obj", deletedState.Obj)
					return
				}
			}
			az.updateNodeCaches(node, nil)

			logger.V(4).Info("Removing node from VMSet cache", "node", node.Name)
			_ = az.VMSet.DeleteCacheForNode(context.Background(), node.Name)
		},
	})
	az.nodeInformerSynced = nodeInformer.HasSynced

	az.serviceLister = informerFactory.Core().V1().Services().Lister()
	az.nodeLister = informerFactory.Core().V1().Nodes().Lister()

	az.setUpEndpointSlicesInformer(informerFactory)
}

// updateNodeCaches updates local cache for node's zones and external resource groups.
func (az *Cloud) updateNodeCaches(prevNode, newNode *v1.Node) {
	logger := log.Background().WithName("updateNodeCaches")
	az.nodeCachesLock.Lock()
	defer az.nodeCachesLock.Unlock()

	if prevNode != nil {
		// Remove from nodeNames cache.
		az.nodeNames.Delete(prevNode.Name)

		// Remove from nodeZones cache.
		prevZone, ok := prevNode.Labels[v1.LabelTopologyZone]
		if ok && az.isAvailabilityZone(prevZone) {
			az.nodeZones[prevZone].Delete(prevNode.Name)
			if az.nodeZones[prevZone].Len() == 0 {
				az.nodeZones[prevZone] = nil
			}
		}

		// Remove from nodeResourceGroups cache.
		_, ok = prevNode.Labels[consts.ExternalResourceGroupLabel]
		if ok {
			delete(az.nodeResourceGroups, prevNode.Name)
		}

		managed, ok := prevNode.Labels[consts.ManagedByAzureLabel]
		isNodeManagedByCloudProvider := !ok || !strings.EqualFold(managed, consts.NotManagedByAzureLabelValue)

		logger.Info("node management status", "managed", managed, "ok", ok, "isNodeManagedByCloudProvider", isNodeManagedByCloudProvider)

		// Remove from unmanagedNodes cache
		if !isNodeManagedByCloudProvider {
			az.unmanagedNodes.Delete(prevNode.Name)
		}

		// Remove from nodePrivateIPs cache.
		for _, address := range getNodePrivateIPAddresses(prevNode) {
			logger.V(6).Info("removing IP address of the node", "address", address, "node", prevNode.Name)
			az.nodePrivateIPs[prevNode.Name].Delete(address)
			delete(az.nodePrivateIPToNodeNameMap, address)
		}

		// if the node is being deleted from the cluster, exclude it from load balancers
		if newNode == nil {
			az.excludeLoadBalancerNodes.Insert(prevNode.Name)
			az.nodesWithCorrectLoadBalancerByPrimaryVMSet.Delete(strings.ToLower(prevNode.Name))
			delete(az.nodePrivateIPs, strings.ToLower(prevNode.Name))
		}
	}

	if newNode != nil {
		// Add to nodeNames cache.
		az.nodeNames = utilsets.SafeInsert(az.nodeNames, newNode.Name)

		// Add to nodeZones cache.
		newZone, ok := newNode.Labels[v1.LabelTopologyZone]
		if ok && az.isAvailabilityZone(newZone) {
			az.nodeZones[newZone] = utilsets.SafeInsert(az.nodeZones[newZone], newNode.Name)
		}

		// Add to nodeResourceGroups cache.
		newRG, ok := newNode.Labels[consts.ExternalResourceGroupLabel]
		if ok && len(newRG) > 0 {
			az.nodeResourceGroups[newNode.Name] = strings.ToLower(newRG)
		}

		_, hasExcludeBalancerLabel := newNode.Labels[v1.LabelNodeExcludeBalancers]
		managed, ok := newNode.Labels[consts.ManagedByAzureLabel]
		isNodeManagedByCloudProvider := !ok || !strings.EqualFold(managed, consts.NotManagedByAzureLabelValue)

		// Update unmanagedNodes cache
		if !isNodeManagedByCloudProvider {
			az.unmanagedNodes.Insert(newNode.Name)
		}

		// Update excludeLoadBalancerNodes cache
		switch {
		case !isNodeManagedByCloudProvider:
			az.excludeLoadBalancerNodes.Insert(newNode.Name)
			logger.V(6).Info("excluding Node from LoadBalancer because it is not managed by cloud provider", "node", newNode.Name)

		case hasExcludeBalancerLabel:
			az.excludeLoadBalancerNodes.Insert(newNode.Name)
			logger.V(6).Info("excluding Node from LoadBalancer because it has exclude-from-external-load-balancers label", "node", newNode.Name)

		default:
			// Nodes not falling into the three cases above are valid backends and
			// should not appear in excludeLoadBalancerNodes cache.
			az.excludeLoadBalancerNodes.Delete(newNode.Name)
		}

		// Add to nodePrivateIPs cache
		for _, address := range getNodePrivateIPAddresses(newNode) {
			if az.nodePrivateIPToNodeNameMap == nil {
				az.nodePrivateIPToNodeNameMap = make(map[string]string)
			}

			logger.V(6).Info("adding IP address of the node", "address", address, "node", newNode.Name)
			az.nodePrivateIPs[strings.ToLower(newNode.Name)] = utilsets.SafeInsert(az.nodePrivateIPs[strings.ToLower(newNode.Name)], address)
			az.nodePrivateIPToNodeNameMap[address] = newNode.Name
		}
	}
}

// updateNodeTaint updates node out-of-service taint
func (az *Cloud) updateNodeTaint(node *v1.Node) {
	logger := log.Background().WithName("updateNodeTaint")
	if node == nil {
		klog.Warningf("node is nil, skip updating node out-of-service taint (should not happen)")
		return
	}
	if az.KubeClient == nil {
		klog.Warningf("az.KubeClient is nil, skip updating node out-of-service taint")
		return
	}

	if isNodeReady(node) {
		if err := cloudnodeutil.RemoveTaintOffNode(az.KubeClient, node.Name, node, nodeOutOfServiceTaint); err != nil {
			logger.Error(err, "failed to remove taint from the node", "taint", v1.TaintNodeOutOfService, "node", node.Name)
		}
	} else {
		// node shutdown taint is added when cloud provider determines instance is shutdown
		if !taints.TaintExists(node.Spec.Taints, nodeOutOfServiceTaint) &&
			taints.TaintExists(node.Spec.Taints, nodeShutdownTaint) {
			logger.V(2).Info("adding taint to node", "taint", v1.TaintNodeOutOfService, "node", node.Name)
			if err := cloudnodeutil.AddOrUpdateTaintOnNode(az.KubeClient, node.Name, nodeOutOfServiceTaint); err != nil {
				logger.Error(err, "failed to add taint to the node", "taint", v1.TaintNodeOutOfService, "node", node.Name)
			}
		} else {
			logger.V(2).Info("node is not ready but either shutdown taint is missing or out-of-service taint is already added, skip adding node out-of-service taint", "node", node.Name)
		}
	}
}

// GetActiveZones returns all the zones in which k8s nodes are currently running.
func (az *Cloud) GetActiveZones() (*utilsets.IgnoreCaseSet, error) {
	if az.nodeInformerSynced == nil {
		return nil, fmt.Errorf("azure cloud provider doesn't have informers set")
	}

	az.nodeCachesLock.RLock()
	defer az.nodeCachesLock.RUnlock()
	if !az.nodeInformerSynced() {
		return nil, fmt.Errorf("node informer is not synced when trying to GetActiveZones")
	}

	zones := utilsets.NewString()
	for zone, nodes := range az.nodeZones {
		if nodes.Len() > 0 {
			zones.Insert(zone)
		}
	}
	return zones, nil
}

// GetLocation returns the location in which k8s cluster is currently running.
func (az *Cloud) GetLocation() string {
	return az.Location
}

// GetNodeResourceGroup gets resource group for given node.
func (az *Cloud) GetNodeResourceGroup(nodeName string) (string, error) {
	// Kubelet won't set az.nodeInformerSynced, always return configured resourceGroup.
	if az.nodeInformerSynced == nil {
		return az.ResourceGroup, nil
	}

	az.nodeCachesLock.RLock()
	defer az.nodeCachesLock.RUnlock()
	if !az.nodeInformerSynced() {
		return "", fmt.Errorf("node informer is not synced when trying to GetNodeResourceGroup")
	}

	// Return external resource group if it has been cached.
	if cachedRG, ok := az.nodeResourceGroups[nodeName]; ok {
		return cachedRG, nil
	}

	// Return resource group from cloud provider options.
	return az.ResourceGroup, nil
}

// GetNodeNames returns a set of all node names in the k8s cluster.
func (az *Cloud) GetNodeNames() (*utilsets.IgnoreCaseSet, error) {
	// Kubelet won't set az.nodeInformerSynced, return nil.
	if az.nodeInformerSynced == nil {
		return nil, nil
	}

	az.nodeCachesLock.RLock()
	defer az.nodeCachesLock.RUnlock()
	if !az.nodeInformerSynced() {
		return nil, fmt.Errorf("node informer is not synced when trying to GetNodeNames")
	}

	return utilsets.NewString(az.nodeNames.UnsortedList()...), nil
}

// GetResourceGroups returns a set of resource groups that all nodes are running on.
func (az *Cloud) GetResourceGroups() (*utilsets.IgnoreCaseSet, error) {
	// Kubelet won't set az.nodeInformerSynced, always return configured resourceGroup.
	if az.nodeInformerSynced == nil {
		return utilsets.NewString(az.ResourceGroup), nil
	}

	az.nodeCachesLock.RLock()
	defer az.nodeCachesLock.RUnlock()
	if !az.nodeInformerSynced() {
		return nil, fmt.Errorf("node informer is not synced when trying to GetResourceGroups")
	}

	resourceGroups := utilsets.NewString(az.ResourceGroup)
	for _, rg := range az.nodeResourceGroups {
		resourceGroups.Insert(rg)
	}

	return resourceGroups, nil
}

// GetUnmanagedNodes returns a list of nodes not managed by Azure cloud provider (e.g. on-prem nodes).
func (az *Cloud) GetUnmanagedNodes() (*utilsets.IgnoreCaseSet, error) {
	// Kubelet won't set az.nodeInformerSynced, always return nil.
	if az.nodeInformerSynced == nil {
		return nil, nil
	}

	az.nodeCachesLock.RLock()
	defer az.nodeCachesLock.RUnlock()
	if !az.nodeInformerSynced() {
		return nil, fmt.Errorf("node informer is not synced when trying to GetUnmanagedNodes")
	}

	return utilsets.NewString(az.unmanagedNodes.UnsortedList()...), nil
}

// ShouldNodeExcludedFromLoadBalancer returns true if node is unmanaged, in external resource group or labeled with "node.kubernetes.io/exclude-from-external-load-balancers".
func (az *Cloud) ShouldNodeExcludedFromLoadBalancer(nodeName string) (bool, error) {
	// Kubelet won't set az.nodeInformerSynced, always return nil.
	if az.nodeInformerSynced == nil {
		return false, nil
	}

	az.nodeCachesLock.RLock()
	defer az.nodeCachesLock.RUnlock()
	if !az.nodeInformerSynced() {
		return false, fmt.Errorf("node informer is not synced when trying to fetch node caches")
	}

	// Return true if the node is in external resource group.
	if cachedRG, ok := az.nodeResourceGroups[nodeName]; ok && !strings.EqualFold(cachedRG, az.ResourceGroup) {
		return true, nil
	}

	return az.excludeLoadBalancerNodes.Has(nodeName), nil
}

func (az *Cloud) getActiveNodesByLoadBalancerName(lbName string) *utilsets.IgnoreCaseSet {
	az.multipleStandardLoadBalancersActiveNodesLock.Lock()
	defer az.multipleStandardLoadBalancersActiveNodesLock.Unlock()

	for _, multiSLBConfig := range az.MultipleStandardLoadBalancerConfigurations {
		if strings.EqualFold(trimSuffixIgnoreCase(lbName, consts.InternalLoadBalancerNameSuffix), multiSLBConfig.Name) {
			return multiSLBConfig.ActiveNodes
		}
	}

	return utilsets.NewString()
}

func isNodeReady(node *v1.Node) bool {
	if node == nil {
		return false
	}
	if _, c := nodeutil.GetNodeCondition(&node.Status, v1.NodeReady); c != nil {
		return c.Status == v1.ConditionTrue
	}
	return false
}

// getNodeVMSet gets the VMSet interface based on config.VMType and the real virtual machine type.
func (az *Cloud) GetNodeVMSet(ctx context.Context, nodeName types.NodeName, crt azcache.AzureCacheReadType) (VMSet, error) {
	// 1. vmType is standard or vmssflex, return cloud.VMSet directly.
	// 1.1 all the nodes in the cluster are avset nodes.
	// 1.2 all the nodes in the cluster are vmssflex nodes.
	if az.VMType == consts.VMTypeStandard || az.VMType == consts.VMTypeVmssFlex {
		return az.VMSet, nil
	}

	// 2. vmType is Virtual Machine Scale Set (vmss), convert vmSet to ScaleSet.
	// 2.1 all the nodes in the cluster are vmss uniform nodes.
	// 2.2 mix node: the nodes in the cluster can be any of avset nodes, vmss uniform nodes and vmssflex nodes.
	ss, ok := az.VMSet.(*ScaleSet)
	if !ok {
		return nil, fmt.Errorf("error of converting vmSet (%q) to ScaleSet with vmType %q", az.VMSet, az.VMType)
	}

	vmManagementType, err := ss.getVMManagementTypeByNodeName(ctx, string(nodeName), crt)
	if err != nil {
		return nil, fmt.Errorf("getNodeVMSet: failed to check the node %s management type: %w", string(nodeName), err)
	}
	// 3. If the node is managed by availability set, then return ss.availabilitySet.
	if vmManagementType == ManagedByAvSet {
		// vm is managed by availability set.
		return ss.availabilitySet, nil
	}
	if vmManagementType == ManagedByVmssFlex {
		// 4. If the node is managed by vmss flex, then return ss.flexScaleSet.
		// vm is managed by vmss flex.
		return ss.flexScaleSet, nil
	}

	// 5. Node is managed by vmss
	return ss, nil
}
