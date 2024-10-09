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
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	azclients "sigs.k8s.io/cloud-provider-azure/pkg/azureclients"
	"sigs.k8s.io/cloud-provider-azure/pkg/azureclients/blobclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azureclients/containerserviceclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azureclients/deploymentclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azureclients/diskclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azureclients/fileclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azureclients/interfaceclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azureclients/loadbalancerclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azureclients/privatednszonegroupclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azureclients/privateendpointclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azureclients/privatelinkserviceclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azureclients/publicipclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azureclients/routeclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azureclients/routetableclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azureclients/securitygroupclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azureclients/snapshotclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azureclients/storageaccountclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azureclients/subnetclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azureclients/vmasclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azureclients/vmclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azureclients/vmsizeclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azureclients/vmssclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azureclients/vmssvmclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azureclients/zoneclient"

	"sigs.k8s.io/yaml"

	azcache "sigs.k8s.io/cloud-provider-azure/pkg/cache"
	"sigs.k8s.io/cloud-provider-azure/pkg/consts"
	ratelimitconfig "sigs.k8s.io/cloud-provider-azure/pkg/provider/config"
	"sigs.k8s.io/cloud-provider-azure/pkg/retry"
	utilsets "sigs.k8s.io/cloud-provider-azure/pkg/util/sets"
	"sigs.k8s.io/cloud-provider-azure/pkg/util/taints"
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

// Config holds the configuration parsed from the --cloud-config flag
// All fields are required unless otherwise specified
// NOTE: Cloud config files should follow the same Kubernetes deprecation policy as
// flags or CLIs. Config fields should not change behavior in incompatible ways and
// should be deprecated for at least 2 release prior to removing.
// See https://kubernetes.io/docs/reference/using-api/deprecation-policy/#deprecating-a-flag-or-cli
// for more details.
type Config struct {
	ratelimitconfig.AzureAuthConfig              `json:",inline" yaml:",inline"`
	ratelimitconfig.CloudProviderRateLimitConfig `json:",inline" yaml:",inline"`

	// The cloud configure type for Azure cloud provider. Supported values are file, secret and merge.
	CloudConfigType configloader.CloudConfigType `json:"cloudConfigType,omitempty" yaml:"cloudConfigType,omitempty"`

	// The name of the resource group that the cluster is deployed in
	ResourceGroup string `json:"resourceGroup,omitempty" yaml:"resourceGroup,omitempty"`
	// The location of the resource group that the cluster is deployed in
	Location string `json:"location,omitempty" yaml:"location,omitempty"`
	// The name of site where the cluster will be deployed to that is more granular than the region specified by the "location" field.
	// Currently only public ip, load balancer and managed disks support this.
	ExtendedLocationName string `json:"extendedLocationName,omitempty" yaml:"extendedLocationName,omitempty"`
	// The type of site that is being targeted.
	// Currently only public ip, load balancer and managed disks support this.
	ExtendedLocationType string `json:"extendedLocationType,omitempty" yaml:"extendedLocationType,omitempty"`
	// The name of the VNet that the cluster is deployed in
	VnetName string `json:"vnetName,omitempty" yaml:"vnetName,omitempty"`
	// The name of the resource group that the Vnet is deployed in
	VnetResourceGroup string `json:"vnetResourceGroup,omitempty" yaml:"vnetResourceGroup,omitempty"`
	// The name of the subnet that the cluster is deployed in
	SubnetName string `json:"subnetName,omitempty" yaml:"subnetName,omitempty"`
	// The name of the security group attached to the cluster's subnet
	SecurityGroupName string `json:"securityGroupName,omitempty" yaml:"securityGroupName,omitempty"`
	// The name of the resource group that the security group is deployed in
	SecurityGroupResourceGroup string `json:"securityGroupResourceGroup,omitempty" yaml:"securityGroupResourceGroup,omitempty"`
	// (Optional in 1.6) The name of the route table attached to the subnet that the cluster is deployed in
	RouteTableName string `json:"routeTableName,omitempty" yaml:"routeTableName,omitempty"`
	// The name of the resource group that the RouteTable is deployed in
	RouteTableResourceGroup string `json:"routeTableResourceGroup,omitempty" yaml:"routeTableResourceGroup,omitempty"`
	// (Optional) The name of the availability set that should be used as the load balancer backend
	// If this is set, the Azure cloudprovider will only add nodes from that availability set to the load
	// balancer backend pool. If this is not set, and multiple agent pools (availability sets) are used, then
	// the cloudprovider will try to add all nodes to a single backend pool which is forbidden.
	// In other words, if you use multiple agent pools (availability sets), you MUST set this field.
	PrimaryAvailabilitySetName string `json:"primaryAvailabilitySetName,omitempty" yaml:"primaryAvailabilitySetName,omitempty"`
	// The type of azure nodes. Candidate values are: vmss, standard and vmssflex.
	// If not set, it will be default to vmss.
	VMType string `json:"vmType,omitempty" yaml:"vmType,omitempty"`
	// The name of the scale set that should be used as the load balancer backend.
	// If this is set, the Azure cloudprovider will only add nodes from that scale set to the load
	// balancer backend pool. If this is not set, and multiple agent pools (scale sets) are used, then
	// the cloudprovider will try to add all nodes to a single backend pool which is forbidden in the basic sku.
	// In other words, if you use multiple agent pools (scale sets), and loadBalancerSku is set to basic, you MUST set this field.
	PrimaryScaleSetName string `json:"primaryScaleSetName,omitempty" yaml:"primaryScaleSetName,omitempty"`
	// Tags determines what tags shall be applied to the shared resources managed by controller manager, which
	// includes load balancer, security group and route table. The supported format is `a=b,c=d,...`. After updated
	// this config, the old tags would be replaced by the new ones.
	// Because special characters are not supported in "tags" configuration, "tags" support would be removed in a future release,
	// please consider migrating the config to "tagsMap".
	Tags string `json:"tags,omitempty" yaml:"tags,omitempty"`
	// TagsMap is similar to Tags but holds tags with special characters such as `=` and `,`.
	TagsMap map[string]string `json:"tagsMap,omitempty" yaml:"tagsMap,omitempty"`
	// SystemTags determines the tag keys managed by cloud provider. If it is not set, no tags would be deleted if
	// the `Tags` is changed. However, the old tags would be deleted if they are neither included in `Tags` nor
	// in `SystemTags` after the update of `Tags`.
	SystemTags string `json:"systemTags,omitempty" yaml:"systemTags,omitempty"`
	// Sku of Load Balancer and Public IP. Candidate values are: basic and standard.
	// If not set, it will be default to basic.
	LoadBalancerSku string `json:"loadBalancerSku,omitempty" yaml:"loadBalancerSku,omitempty"`
	// LoadBalancerName determines the specific name of the load balancer user want to use, working with
	// LoadBalancerResourceGroup
	LoadBalancerName string `json:"loadBalancerName,omitempty" yaml:"loadBalancerName,omitempty"`
	// LoadBalancerResourceGroup determines the specific resource group of the load balancer user want to use, working
	// with LoadBalancerName
	LoadBalancerResourceGroup string `json:"loadBalancerResourceGroup,omitempty" yaml:"loadBalancerResourceGroup,omitempty"`
	// PreConfiguredBackendPoolLoadBalancerTypes determines whether the LoadBalancer BackendPool has been preconfigured.
	// Candidate values are:
	//   "": exactly with today (not pre-configured for any LBs)
	//   "internal": for internal LoadBalancer
	//   "external": for external LoadBalancer
	//   "all": for both internal and external LoadBalancer
	PreConfiguredBackendPoolLoadBalancerTypes string `json:"preConfiguredBackendPoolLoadBalancerTypes,omitempty" yaml:"preConfiguredBackendPoolLoadBalancerTypes,omitempty"`

	// DisableAvailabilitySetNodes disables VMAS nodes support when "VMType" is set to "vmss".
	DisableAvailabilitySetNodes bool `json:"disableAvailabilitySetNodes,omitempty" yaml:"disableAvailabilitySetNodes,omitempty"`
	// EnableVmssFlexNodes enables vmss flex nodes support when "VMType" is set to "vmss".
	EnableVmssFlexNodes bool `json:"enableVmssFlexNodes,omitempty" yaml:"enableVmssFlexNodes,omitempty"`
	// DisableAzureStackCloud disables AzureStackCloud support. It should be used
	// when setting AzureAuthConfig.Cloud with "AZURESTACKCLOUD" to customize ARM endpoints
	// while the cluster is not running on AzureStack.
	DisableAzureStackCloud bool `json:"disableAzureStackCloud,omitempty" yaml:"disableAzureStackCloud,omitempty"`
	// Enable exponential backoff to manage resource request retries
	CloudProviderBackoff bool `json:"cloudProviderBackoff,omitempty" yaml:"cloudProviderBackoff,omitempty"`
	// Use instance metadata service where possible
	UseInstanceMetadata bool `json:"useInstanceMetadata,omitempty" yaml:"useInstanceMetadata,omitempty"`

	// Backoff exponent
	CloudProviderBackoffExponent float64 `json:"cloudProviderBackoffExponent,omitempty" yaml:"cloudProviderBackoffExponent,omitempty"`
	// Backoff jitter
	CloudProviderBackoffJitter float64 `json:"cloudProviderBackoffJitter,omitempty" yaml:"cloudProviderBackoffJitter,omitempty"`

	// ExcludeMasterFromStandardLB excludes master nodes from standard load balancer.
	// If not set, it will be default to true.
	ExcludeMasterFromStandardLB *bool `json:"excludeMasterFromStandardLB,omitempty" yaml:"excludeMasterFromStandardLB,omitempty"`
	// DisableOutboundSNAT disables the outbound SNAT for public load balancer rules.
	// It should only be set when loadBalancerSku is standard. If not set, it will be default to false.
	DisableOutboundSNAT *bool `json:"disableOutboundSNAT,omitempty" yaml:"disableOutboundSNAT,omitempty"`

	// Maximum allowed LoadBalancer Rule Count is the limit enforced by Azure Load balancer
	MaximumLoadBalancerRuleCount int `json:"maximumLoadBalancerRuleCount,omitempty" yaml:"maximumLoadBalancerRuleCount,omitempty"`
	// Backoff retry limit
	CloudProviderBackoffRetries int `json:"cloudProviderBackoffRetries,omitempty" yaml:"cloudProviderBackoffRetries,omitempty"`
	// Backoff duration
	CloudProviderBackoffDuration int `json:"cloudProviderBackoffDuration,omitempty" yaml:"cloudProviderBackoffDuration,omitempty"`
	// NonVmssUniformNodesCacheTTLInSeconds sets the Cache TTL for NonVmssUniformNodesCacheTTLInSeconds
	// if not set, will use default value
	NonVmssUniformNodesCacheTTLInSeconds int `json:"nonVmssUniformNodesCacheTTLInSeconds,omitempty" yaml:"nonVmssUniformNodesCacheTTLInSeconds,omitempty"`
	// VmssCacheTTLInSeconds sets the cache TTL for VMSS
	VmssCacheTTLInSeconds int `json:"vmssCacheTTLInSeconds,omitempty" yaml:"vmssCacheTTLInSeconds,omitempty"`
	// VmssVirtualMachinesCacheTTLInSeconds sets the cache TTL for vmssVirtualMachines
	VmssVirtualMachinesCacheTTLInSeconds int `json:"vmssVirtualMachinesCacheTTLInSeconds,omitempty" yaml:"vmssVirtualMachinesCacheTTLInSeconds,omitempty"`

	// VmssFlexCacheTTLInSeconds sets the cache TTL for VMSS Flex
	VmssFlexCacheTTLInSeconds int `json:"vmssFlexCacheTTLInSeconds,omitempty" yaml:"vmssFlexCacheTTLInSeconds,omitempty"`
	// VmssFlexVMCacheTTLInSeconds sets the cache TTL for vmss flex vms
	VmssFlexVMCacheTTLInSeconds int `json:"vmssFlexVMCacheTTLInSeconds,omitempty" yaml:"vmssFlexVMCacheTTLInSeconds,omitempty"`

	// VmCacheTTLInSeconds sets the cache TTL for vm
	VMCacheTTLInSeconds int `json:"vmCacheTTLInSeconds,omitempty" yaml:"vmCacheTTLInSeconds,omitempty"`
	// LoadBalancerCacheTTLInSeconds sets the cache TTL for load balancer
	LoadBalancerCacheTTLInSeconds int `json:"loadBalancerCacheTTLInSeconds,omitempty" yaml:"loadBalancerCacheTTLInSeconds,omitempty"`
	// NsgCacheTTLInSeconds sets the cache TTL for network security group
	NsgCacheTTLInSeconds int `json:"nsgCacheTTLInSeconds,omitempty" yaml:"nsgCacheTTLInSeconds,omitempty"`
	// RouteTableCacheTTLInSeconds sets the cache TTL for route table
	RouteTableCacheTTLInSeconds int `json:"routeTableCacheTTLInSeconds,omitempty" yaml:"routeTableCacheTTLInSeconds,omitempty"`
	// PlsCacheTTLInSeconds sets the cache TTL for private link service resource
	PlsCacheTTLInSeconds int `json:"plsCacheTTLInSeconds,omitempty" yaml:"plsCacheTTLInSeconds,omitempty"`
	// AvailabilitySetsCacheTTLInSeconds sets the cache TTL for VMAS
	AvailabilitySetsCacheTTLInSeconds int `json:"availabilitySetsCacheTTLInSeconds,omitempty" yaml:"availabilitySetsCacheTTLInSeconds,omitempty"`
	// PublicIPCacheTTLInSeconds sets the cache TTL for public ip
	PublicIPCacheTTLInSeconds int `json:"publicIPCacheTTLInSeconds,omitempty" yaml:"publicIPCacheTTLInSeconds,omitempty"`
	// RouteUpdateWaitingInSeconds is the delay time for waiting route updates to take effect. This waiting delay is added
	// because the routes are not taken effect when the async route updating operation returns success. Default is 30 seconds.
	RouteUpdateWaitingInSeconds int `json:"routeUpdateWaitingInSeconds,omitempty" yaml:"routeUpdateWaitingInSeconds,omitempty"`
	// The user agent for Azure customer usage attribution
	UserAgent string `json:"userAgent,omitempty" yaml:"userAgent,omitempty"`
	// LoadBalancerBackendPoolConfigurationType defines how vms join the load balancer backend pools. Supported values
	// are `nodeIPConfiguration`, `nodeIP` and `podIP`.
	// `nodeIPConfiguration`: vm network interfaces will be attached to the inbound backend pool of the load balancer (default);
	// `nodeIP`: vm private IPs will be attached to the inbound backend pool of the load balancer;
	// `podIP`: pod IPs will be attached to the inbound backend pool of the load balancer (not supported yet).
	LoadBalancerBackendPoolConfigurationType string `json:"loadBalancerBackendPoolConfigurationType,omitempty" yaml:"loadBalancerBackendPoolConfigurationType,omitempty"`
	// PutVMSSVMBatchSize defines how many requests the client send concurrently when putting the VMSS VMs.
	// If it is smaller than or equal to zero, the request will be sent one by one in sequence (default).
	PutVMSSVMBatchSize int `json:"putVMSSVMBatchSize" yaml:"putVMSSVMBatchSize"`
	// PrivateLinkServiceResourceGroup determines the specific resource group of the private link services user want to use
	PrivateLinkServiceResourceGroup string `json:"privateLinkServiceResourceGroup,omitempty" yaml:"privateLinkServiceResourceGroup,omitempty"`

	// EnableMigrateToIPBasedBackendPoolAPI uses the migration API to migrate from NIC-based to IP-based backend pool.
	// The migration API can provide a migration from NIC-based to IP-based backend pool without service downtime.
	// If the API is not used, the migration will be done by decoupling all nodes on the backend pool and then re-attaching
	// node IPs, which will introduce service downtime. The downtime increases with the number of nodes in the backend pool.
	EnableMigrateToIPBasedBackendPoolAPI bool `json:"enableMigrateToIPBasedBackendPoolAPI" yaml:"enableMigrateToIPBasedBackendPoolAPI"`

	// MultipleStandardLoadBalancerConfigurations stores the properties regarding multiple standard load balancers.
	// It will be ignored if LoadBalancerBackendPoolConfigurationType is nodeIPConfiguration.
	// If the length is not 0, it is assumed the multiple standard load balancers mode is on. In this case,
	// there must be one configuration named "<clustername>" or an error will be reported.
	MultipleStandardLoadBalancerConfigurations []MultipleStandardLoadBalancerConfiguration `json:"multipleStandardLoadBalancerConfigurations,omitempty" yaml:"multipleStandardLoadBalancerConfigurations,omitempty"`

	// DisableAPICallCache disables the cache for Azure API calls. It is for ARG support and not all resources will be disabled.
	DisableAPICallCache bool `json:"disableAPICallCache,omitempty" yaml:"disableAPICallCache,omitempty"`

	// RouteUpdateIntervalInSeconds is the interval for updating routes. Default is 30 seconds.
	RouteUpdateIntervalInSeconds int `json:"routeUpdateIntervalInSeconds,omitempty" yaml:"routeUpdateIntervalInSeconds,omitempty"`
	// LoadBalancerBackendPoolUpdateIntervalInSeconds is the interval for updating load balancer backend pool of local services. Default is 30 seconds.
	LoadBalancerBackendPoolUpdateIntervalInSeconds int `json:"loadBalancerBackendPoolUpdateIntervalInSeconds,omitempty" yaml:"loadBalancerBackendPoolUpdateIntervalInSeconds,omitempty"`

	// ClusterServiceLoadBalancerHealthProbeMode determines the health probe mode for cluster service load balancer.
	// Supported values are `shared` and `servicenodeport`.
	// `servicenodeport`: the health probe will be created against each port of each service by watching the backend application (default).
	// `shared`: all cluster services shares one HTTP probe targeting the kube-proxy on the node (<nodeIP>/healthz:10256).
	ClusterServiceLoadBalancerHealthProbeMode string `json:"clusterServiceLoadBalancerHealthProbeMode,omitempty" yaml:"clusterServiceLoadBalancerHealthProbeMode,omitempty"`
	// ClusterServiceSharedLoadBalancerHealthProbePort defines the target port of the shared health probe. Default to 10256.
	ClusterServiceSharedLoadBalancerHealthProbePort int32 `json:"clusterServiceSharedLoadBalancerHealthProbePort,omitempty" yaml:"clusterServiceSharedLoadBalancerHealthProbePort,omitempty"`
	// ClusterServiceSharedLoadBalancerHealthProbePath defines the target path of the shared health probe. Default to `/healthz`.
	ClusterServiceSharedLoadBalancerHealthProbePath string `json:"clusterServiceSharedLoadBalancerHealthProbePath,omitempty" yaml:"clusterServiceSharedLoadBalancerHealthProbePath,omitempty"`
}

// MultipleStandardLoadBalancerConfiguration stores the properties regarding multiple standard load balancers.
type MultipleStandardLoadBalancerConfiguration struct {
	// Name of the public load balancer. There will be an internal load balancer
	// created if needed, and the name will be `<name>-internal`. The internal lb
	// shares the same configurations as the external one. The internal lbs
	// are not needed to be included in `MultipleStandardLoadBalancerConfigurations`.
	// There must be a name of "<clustername>" in the load balancer configuration list.
	Name string `json:"name" yaml:"name"`

	MultipleStandardLoadBalancerConfigurationSpec

	MultipleStandardLoadBalancerConfigurationStatus
}

// MultipleStandardLoadBalancerConfigurationSpec stores the properties regarding multiple standard load balancers.
type MultipleStandardLoadBalancerConfigurationSpec struct {
	// This load balancer can have services placed on it. Defaults to true,
	// can be set to false to drain and eventually remove a load balancer.
	// This only affects services that will be using the LB. For services
	// that is currently using the LB, they will not be affected.
	AllowServicePlacement *bool `json:"allowServicePlacement" yaml:"allowServicePlacement"`

	// A string value that must specify the name of an existing vmSet.
	// All nodes in the given vmSet will always be added to this load balancer.
	// A vmSet can only be the primary vmSet for a single load balancer.
	PrimaryVMSet string `json:"primaryVMSet" yaml:"primaryVMSet"`

	// Services that must match this selector can be placed on this load balancer. If not supplied,
	// services with any labels can be created on the load balancer.
	ServiceLabelSelector *metav1.LabelSelector `json:"serviceLabelSelector" yaml:"serviceLabelSelector"`

	// Services created in namespaces with the supplied label will be allowed to select that load balancer.
	// If not supplied, services created in any namespaces can be created on that load balancer.
	ServiceNamespaceSelector *metav1.LabelSelector `json:"serviceNamespaceSelector" yaml:"serviceNamespaceSelector"`

	// Nodes matching this selector will be preferentially added to the load balancers that
	// they match selectors for. NodeSelector does not override primaryAgentPool for node allocation.
	NodeSelector *metav1.LabelSelector `json:"nodeSelector" yaml:"nodeSelector"`
}

// MultipleStandardLoadBalancerConfigurationStatus stores the properties regarding multiple standard load balancers.
type MultipleStandardLoadBalancerConfigurationStatus struct {
	// ActiveServices stores the services that are supposed to use the load balancer.
	ActiveServices *utilsets.IgnoreCaseSet `json:"activeServices" yaml:"activeServices"`

	// ActiveNodes stores the nodes that are supposed to be in the load balancer.
	// It will be used in EnsureHostsInPool to make sure the given ones are in the backend pool.
	ActiveNodes *utilsets.IgnoreCaseSet `json:"activeNodes" yaml:"activeNodes"`
}

// HasExtendedLocation returns true if extendedlocation prop are specified.
func (config *Config) HasExtendedLocation() bool {
	return config.ExtendedLocationName != "" && config.ExtendedLocationType != ""
}

var (
	_ cloudprovider.Interface    = (*Cloud)(nil)
	_ cloudprovider.Instances    = (*Cloud)(nil)
	_ cloudprovider.LoadBalancer = (*Cloud)(nil)
	_ cloudprovider.Routes       = (*Cloud)(nil)
	_ cloudprovider.Zones        = (*Cloud)(nil)
)

// Cloud holds the config and clients
type Cloud struct {
	Config
	Environment azure.Environment

	RoutesClient                    routeclient.Interface
	SubnetsClient                   subnetclient.Interface
	InterfacesClient                interfaceclient.Interface
	RouteTablesClient               routetableclient.Interface
	LoadBalancerClient              loadbalancerclient.Interface
	PublicIPAddressesClient         publicipclient.Interface
	SecurityGroupsClient            securitygroupclient.Interface
	VirtualMachinesClient           vmclient.Interface
	StorageAccountClient            storageaccountclient.Interface
	DisksClient                     diskclient.Interface
	SnapshotsClient                 snapshotclient.Interface
	FileClient                      fileclient.Interface
	BlobClient                      blobclient.Interface
	VirtualMachineScaleSetsClient   vmssclient.Interface
	VirtualMachineScaleSetVMsClient vmssvmclient.Interface
	VirtualMachineSizesClient       vmsizeclient.Interface
	AvailabilitySetsClient          vmasclient.Interface
	ZoneClient                      zoneclient.Interface
	privateendpointclient           privateendpointclient.Interface
	privatednszonegroupclient       privatednszonegroupclient.Interface
	PrivateLinkServiceClient        privatelinkserviceclient.Interface
	containerServiceClient          containerserviceclient.Interface
	deploymentClient                deploymentclient.Interface
	ComputeClientFactory            azclient.ClientFactory
	NetworkClientFactory            azclient.ClientFactory

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

	vmCache  azcache.Resource
	lbCache  azcache.Resource
	nsgCache azcache.Resource
	rtCache  azcache.Resource
	// public ip cache
	// key: [resourceGroupName]
	// Value: sync.Map of [pipName]*PublicIPAddress
	pipCache azcache.Resource
	// use [resourceGroupName*LBFrontEndIpConfigurationID] as the key and search for PLS attached to the frontEnd
	plsCache azcache.Resource
	// a timed cache storing storage account properties to avoid querying storage account frequently
	storageAccountCache azcache.Resource

	// Add service lister to always get latest service
	serviceLister corelisters.ServiceLister
	// node-sync-loop routine and service-reconcile routine should not update LoadBalancer at the same time
	serviceReconcileLock sync.Mutex

	lockMap *LockMap
	// multipleStandardLoadBalancerConfigurationsSynced make sure the `reconcileMultipleStandardLoadBalancerConfigurations`
	// runs only once every time the cloud provide restarts.
	multipleStandardLoadBalancerConfigurationsSynced bool
	// nodesWithCorrectLoadBalancerByPrimaryVMSet marks nodes that are matched with load balancers by primary vmSet.
	nodesWithCorrectLoadBalancerByPrimaryVMSet      sync.Map
	multipleStandardLoadBalancersActiveServicesLock sync.Mutex
	multipleStandardLoadBalancersActiveNodesLock    sync.Mutex
	localServiceNameToServiceInfoMap                sync.Map
	endpointSlicesCache                             sync.Map
}

// NewCloud returns a Cloud with initialized clients
func NewCloud(ctx context.Context, config *Config, callFromCCM bool) (cloudprovider.Interface, error) {
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
	az.lockMap = newLockMap()
	return az, nil
}

func NewCloudFromConfigFile(ctx context.Context, configFilePath string, calFromCCM bool) (cloudprovider.Interface, error) {
	var (
		cloud cloudprovider.Interface
		err   error
	)

	var configValue *Config
	if configFilePath != "" {
		var config *os.File
		config, err = os.Open(configFilePath)
		if err != nil {
			klog.Fatalf("Couldn't open cloud provider configuration %s: %#v",
				configFilePath, err)
		}

		defer config.Close()
		configValue, err = ParseConfig(config)
		if err != nil {
			klog.Fatalf("Failed to parse Azure cloud provider config: %v", err)
		}
	}
	cloud, err = NewCloud(ctx, configValue, calFromCCM && configFilePath != "")

	if err != nil {
		return nil, fmt.Errorf("could not init cloud provider azure: %w", err)
	}
	if cloud == nil {
		return nil, fmt.Errorf("nil cloud")
	}

	return cloud, nil
}

func NewCloudFromSecret(ctx context.Context, clientBuilder cloudprovider.ControllerClientBuilder, secretName, secretNamespace, cloudConfigKey string) (cloudprovider.Interface, error) {
	config, err := configloader.Load[Config](ctx, &configloader.K8sSecretLoaderConfig{
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
	az, err := NewCloud(ctx, config, true)
	if err != nil {
		return nil, fmt.Errorf("NewCloudFromSecret: failed to initialize cloud from secret %s/%s: %w", secretNamespace, secretName, err)
	}
	az.Initialize(clientBuilder, wait.NeverStop)

	return az, nil
}

// InitializeCloudFromConfig initializes the Cloud from config.
func (az *Cloud) InitializeCloudFromConfig(ctx context.Context, config *Config, fromSecret, callFromCCM bool) error {
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

	env, err := ratelimitconfig.ParseAzureEnvironment(config.Cloud, config.ResourceManagerEndpoint, config.IdentitySystem)
	if err != nil {
		return err
	}

	// Initialize rate limiting config options.
	ratelimitconfig.InitializeCloudProviderRateLimitConfig(&config.CloudProviderRateLimitConfig)

	resourceRequestBackoff := az.setCloudProviderBackoffDefaults(config)

	err = az.setLBDefaults(config)
	if err != nil {
		return err
	}

	az.Config = *config
	az.Environment = *env
	az.ResourceRequestBackoff = resourceRequestBackoff
	az.Metadata, err = NewInstanceMetadataService(consts.ImdsServer)
	if err != nil {
		return err
	}

	if az.MaximumLoadBalancerRuleCount == 0 {
		az.MaximumLoadBalancerRuleCount = consts.MaximumLoadBalancerRuleCount
	}

	if strings.EqualFold(consts.VMTypeVMSS, az.Config.VMType) {
		az.VMSet, err = newScaleSet(ctx, az)
		if err != nil {
			return err
		}
	} else if strings.EqualFold(consts.VMTypeVmssFlex, az.Config.VMType) {
		az.VMSet, err = newFlexScaleSet(ctx, az)
		if err != nil {
			return err
		}
	} else {
		az.VMSet, err = newAvailabilitySet(az)
		if err != nil {
			return err
		}
	}

	if az.isLBBackendPoolTypeNodeIPConfig() {
		az.LoadBalancerBackendPool = newBackendPoolTypeNodeIPConfig(az)
	} else if az.isLBBackendPoolTypeNodeIP() {
		az.LoadBalancerBackendPool = newBackendPoolTypeNodeIP(az)
	}

	if az.useMultipleStandardLoadBalancers() {
		if err := az.checkEnableMultipleStandardLoadBalancers(); err != nil {
			return err
		}
	}
	servicePrincipalToken, err := ratelimitconfig.GetServicePrincipalToken(&config.AzureAuthConfig, env, env.ServiceManagementEndpoint)
	if errors.Is(err, ratelimitconfig.ErrorNoAuth) {
		// Only controller-manager would lazy-initialize from secret, and credentials are required for such case.
		if fromSecret {
			err := fmt.Errorf("no credentials provided for Azure cloud provider")
			klog.Fatal(err)
			return err
		}

		// No credentials provided, useInstanceMetadata should be enabled for Kubelet.
		// TODO(feiskyer): print different error message for Kubelet and controller-manager, as they're
		// requiring different credential settings.
		if !config.UseInstanceMetadata && config.CloudConfigType == configloader.CloudConfigTypeFile {
			return fmt.Errorf("useInstanceMetadata must be enabled without Azure credentials")
		}

		klog.V(2).Infof("Azure cloud provider is starting without credentials")
	} else if err != nil {
		return err
	}
	// No credentials provided, InstanceMetadataService would be used for getting Azure resources.
	// Note that this only applies to Kubelet, controller-manager should configure credentials for managing Azure resources.
	if servicePrincipalToken == nil {
		return nil
	}

	// If uses network resources in different AAD Tenant, then prepare corresponding Service Principal Token for VM/VMSS client and network resources client
	multiTenantServicePrincipalToken, networkResourceServicePrincipalToken, err := az.getAuthTokenInMultiTenantEnv(servicePrincipalToken)
	if err != nil {
		return err
	}
	az.configAzureClients(servicePrincipalToken, multiTenantServicePrincipalToken, networkResourceServicePrincipalToken)

	if az.ComputeClientFactory == nil {
		authProvider, err := azclient.NewAuthProvider(&az.ARMClientConfig, &az.AzureAuthConfig.AzureAuthConfig)
		if err != nil {
			return err
		}
		var cred azcore.TokenCredential
		if authProvider.IsMultiTenantModeEnabled() {
			multiTenantCred := authProvider.GetMultiTenantIdentity()
			networkTenantCred := authProvider.GetNetworkAzIdentity()
			az.NetworkClientFactory, err = azclient.NewClientFactory(&azclient.ClientFactoryConfig{
				SubscriptionID: az.NetworkResourceSubscriptionID,
			}, &az.ARMClientConfig, networkTenantCred)
			if err != nil {
				return err
			}
			cred = multiTenantCred
		} else {
			cred = authProvider.GetAzIdentity()
		}
		az.ComputeClientFactory, err = azclient.NewClientFactory(&azclient.ClientFactoryConfig{
			SubscriptionID: az.SubscriptionID,
		}, &az.ARMClientConfig, cred)
		if err != nil {
			return err
		}
	}

	err = az.initCaches()
	if err != nil {
		return err
	}

	// Common controller contains the function
	// needed by both blob disk and managed disk controllers
	qps := float32(ratelimitconfig.DefaultAtachDetachDiskQPS)
	bucket := ratelimitconfig.DefaultAtachDetachDiskBucket
	if az.Config.AttachDetachDiskRateLimit != nil {
		qps = az.Config.AttachDetachDiskRateLimit.CloudProviderRateLimitQPSWrite
		bucket = az.Config.AttachDetachDiskRateLimit.CloudProviderRateLimitBucketWrite
	}
	klog.V(2).Infof("attach/detach disk operation rate limit QPS: %f, Bucket: %d", qps, bucket)

	// updating routes and syncing zones only in CCM
	if callFromCCM {
		// start delayed route updater.
		if az.RouteUpdateIntervalInSeconds == 0 {
			az.RouteUpdateIntervalInSeconds = consts.DefaultRouteUpdateIntervalInSeconds
		}
		az.routeUpdater = newDelayedRouteUpdater(az, time.Duration(az.RouteUpdateIntervalInSeconds)*time.Second)
		go az.routeUpdater.run(ctx)

		// start backend pool updater.
		if az.useMultipleStandardLoadBalancers() {
			az.backendPoolUpdater = newLoadBalancerBackendPoolUpdater(az, time.Duration(az.LoadBalancerBackendPoolUpdateIntervalInSeconds)*time.Second)
			go az.backendPoolUpdater.run(ctx)
		}

		// Azure Stack does not support zone at the moment
		// https://docs.microsoft.com/en-us/azure-stack/user/azure-stack-network-differences?view=azs-2102
		if !az.isStackCloud() {
			// wait for the success first time of syncing zones
			err = az.syncRegionZonesMap()
			if err != nil {
				klog.Errorf("InitializeCloudFromConfig: failed to sync regional zones map for the first time: %s", err.Error())
				return err
			}

			go az.refreshZones(ctx, az.syncRegionZonesMap)
		}
	}

	return nil
}

func (az *Cloud) useMultipleStandardLoadBalancers() bool {
	return az.useStandardLoadBalancer() && len(az.MultipleStandardLoadBalancerConfigurations) > 0
}

func (az *Cloud) useSingleStandardLoadBalancer() bool {
	return az.useStandardLoadBalancer() && len(az.MultipleStandardLoadBalancerConfigurations) == 0
}

// Multiple standard load balancer mode only supports IP-based load balancers.
func (az *Cloud) checkEnableMultipleStandardLoadBalancers() error {
	if az.isLBBackendPoolTypeNodeIPConfig() {
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

func (az *Cloud) isLBBackendPoolTypeNodeIPConfig() bool {
	return strings.EqualFold(az.LoadBalancerBackendPoolConfigurationType, consts.LoadBalancerBackendPoolConfigurationTypeNodeIPConfiguration)
}

func (az *Cloud) isLBBackendPoolTypeNodeIP() bool {
	return strings.EqualFold(az.LoadBalancerBackendPoolConfigurationType, consts.LoadBalancerBackendPoolConfigurationTypeNodeIP)
}

func (az *Cloud) getPutVMSSVMBatchSize() int {
	return az.PutVMSSVMBatchSize
}

func (az *Cloud) initCaches() (err error) {
	if az.Config.DisableAPICallCache {
		klog.Infof("API call cache is disabled, ignore logs about cache operations")
	}

	az.vmCache, err = az.newVMCache()
	if err != nil {
		return err
	}

	az.lbCache, err = az.newLBCache()
	if err != nil {
		return err
	}

	az.nsgCache, err = az.newNSGCache()
	if err != nil {
		return err
	}

	az.rtCache, err = az.newRouteTableCache()
	if err != nil {
		return err
	}

	az.pipCache, err = az.newPIPCache()
	if err != nil {
		return err
	}

	az.plsCache, err = az.newPLSCache()
	if err != nil {
		return err
	}

	if az.storageAccountCache, err = az.newStorageAccountCache(); err != nil {
		return err
	}
	return nil
}

func (az *Cloud) setLBDefaults(config *Config) error {
	if config.LoadBalancerSku == "" {
		config.LoadBalancerSku = consts.LoadBalancerSkuStandard
	}

	if strings.EqualFold(config.LoadBalancerSku, consts.LoadBalancerSkuStandard) {
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
			return fmt.Errorf("disableOutboundSNAT should only set when loadBalancerSku is standard")
		}
	}
	return nil
}

func (az *Cloud) getAuthTokenInMultiTenantEnv(_ *adal.ServicePrincipalToken) (*adal.MultiTenantServicePrincipalToken, *adal.ServicePrincipalToken, error) {
	var err error
	var multiTenantServicePrincipalToken *adal.MultiTenantServicePrincipalToken
	var networkResourceServicePrincipalToken *adal.ServicePrincipalToken
	if az.Config.UsesNetworkResourceInDifferentTenant() {
		multiTenantServicePrincipalToken, err = ratelimitconfig.GetMultiTenantServicePrincipalToken(&az.Config.AzureAuthConfig, &az.Environment)
		if err != nil {
			return nil, nil, err
		}
		networkResourceServicePrincipalToken, err = ratelimitconfig.GetNetworkResourceServicePrincipalToken(&az.Config.AzureAuthConfig, &az.Environment)
		if err != nil {
			return nil, nil, err
		}
	}
	return multiTenantServicePrincipalToken, networkResourceServicePrincipalToken, nil
}

func (az *Cloud) setCloudProviderBackoffDefaults(config *Config) wait.Backoff {
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
		klog.V(2).Infof("Azure cloudprovider using try backoff: retries=%d, exponent=%f, duration=%d, jitter=%f",
			config.CloudProviderBackoffRetries,
			config.CloudProviderBackoffExponent,
			config.CloudProviderBackoffDuration,
			config.CloudProviderBackoffJitter)
	} else {
		// CloudProviderBackoffRetries will be set to 1 by default as the requirements of Azure SDK.
		config.CloudProviderBackoffRetries = 1
		config.CloudProviderBackoffDuration = consts.BackoffDurationDefault
	}
	return resourceRequestBackoff
}

func (az *Cloud) configAzureClients(
	servicePrincipalToken *adal.ServicePrincipalToken,
	multiTenantServicePrincipalToken *adal.MultiTenantServicePrincipalToken,
	networkResourceServicePrincipalToken *adal.ServicePrincipalToken) {
	azClientConfig := az.getAzureClientConfig(servicePrincipalToken)

	// Prepare AzureClientConfig for all azure clients
	interfaceClientConfig := azClientConfig.WithRateLimiter(az.Config.InterfaceRateLimit)
	vmSizeClientConfig := azClientConfig.WithRateLimiter(az.Config.VirtualMachineSizeRateLimit)
	snapshotClientConfig := azClientConfig.WithRateLimiter(az.Config.SnapshotRateLimit)
	storageAccountClientConfig := azClientConfig.WithRateLimiter(az.Config.StorageAccountRateLimit)
	diskClientConfig := azClientConfig.WithRateLimiter(az.Config.DiskRateLimit)
	vmClientConfig := azClientConfig.WithRateLimiter(az.Config.VirtualMachineRateLimit)
	vmssClientConfig := azClientConfig.WithRateLimiter(az.Config.VirtualMachineScaleSetRateLimit)
	// Error "not an active Virtual Machine Scale Set VM" is not retriable for VMSS VM.
	// But http.StatusNotFound is retriable because of ARM replication latency.
	vmssVMClientConfig := azClientConfig.WithRateLimiter(az.Config.VirtualMachineScaleSetRateLimit)
	vmssVMClientConfig.Backoff = vmssVMClientConfig.Backoff.WithNonRetriableErrors([]string{consts.VmssVMNotActiveErrorMessage}).WithRetriableHTTPStatusCodes([]int{http.StatusNotFound})
	routeClientConfig := azClientConfig.WithRateLimiter(az.Config.RouteRateLimit)
	subnetClientConfig := azClientConfig.WithRateLimiter(az.Config.SubnetsRateLimit)
	routeTableClientConfig := azClientConfig.WithRateLimiter(az.Config.RouteTableRateLimit)
	loadBalancerClientConfig := azClientConfig.WithRateLimiter(az.Config.LoadBalancerRateLimit)
	securityGroupClientConfig := azClientConfig.WithRateLimiter(az.Config.SecurityGroupRateLimit)
	publicIPClientConfig := azClientConfig.WithRateLimiter(az.Config.PublicIPAddressRateLimit)
	containerServiceConfig := azClientConfig.WithRateLimiter(az.Config.ContainerServiceRateLimit)
	deploymentConfig := azClientConfig.WithRateLimiter(az.Config.DeploymentRateLimit)
	privateDNSZoenGroupConfig := azClientConfig.WithRateLimiter(az.Config.PrivateDNSZoneGroupRateLimit)
	privateEndpointConfig := azClientConfig.WithRateLimiter(az.Config.PrivateEndpointRateLimit)
	privateLinkServiceConfig := azClientConfig.WithRateLimiter(az.Config.PrivateLinkServiceRateLimit)
	// TODO(ZeroMagic): add azurefileRateLimit
	fileClientConfig := azClientConfig.WithRateLimiter(nil)
	blobClientConfig := azClientConfig.WithRateLimiter(nil)
	vmasClientConfig := azClientConfig.WithRateLimiter(az.Config.AvailabilitySetRateLimit)
	zoneClientConfig := azClientConfig.WithRateLimiter(nil)

	// If uses network resources in different AAD Tenant, update Authorizer for VM/VMSS/VMAS client config
	if multiTenantServicePrincipalToken != nil {
		multiTenantServicePrincipalTokenAuthorizer := autorest.NewMultiTenantServicePrincipalTokenAuthorizer(multiTenantServicePrincipalToken)
		vmClientConfig.Authorizer = multiTenantServicePrincipalTokenAuthorizer
		vmssClientConfig.Authorizer = multiTenantServicePrincipalTokenAuthorizer
		vmssVMClientConfig.Authorizer = multiTenantServicePrincipalTokenAuthorizer
		vmasClientConfig.Authorizer = multiTenantServicePrincipalTokenAuthorizer
	}

	// If uses network resources in different AAD Tenant, update SubscriptionID and Authorizer for network resources client config
	if networkResourceServicePrincipalToken != nil {
		networkResourceServicePrincipalTokenAuthorizer := autorest.NewBearerAuthorizer(networkResourceServicePrincipalToken)
		routeClientConfig.Authorizer = networkResourceServicePrincipalTokenAuthorizer
		subnetClientConfig.Authorizer = networkResourceServicePrincipalTokenAuthorizer
		routeTableClientConfig.Authorizer = networkResourceServicePrincipalTokenAuthorizer
		loadBalancerClientConfig.Authorizer = networkResourceServicePrincipalTokenAuthorizer
		securityGroupClientConfig.Authorizer = networkResourceServicePrincipalTokenAuthorizer
		publicIPClientConfig.Authorizer = networkResourceServicePrincipalTokenAuthorizer
	}

	if az.UsesNetworkResourceInDifferentSubscription() {
		routeClientConfig.SubscriptionID = az.Config.NetworkResourceSubscriptionID
		subnetClientConfig.SubscriptionID = az.Config.NetworkResourceSubscriptionID
		routeTableClientConfig.SubscriptionID = az.Config.NetworkResourceSubscriptionID
		loadBalancerClientConfig.SubscriptionID = az.Config.NetworkResourceSubscriptionID
		securityGroupClientConfig.SubscriptionID = az.Config.NetworkResourceSubscriptionID
		publicIPClientConfig.SubscriptionID = az.Config.NetworkResourceSubscriptionID
	}

	// Initialize all azure clients based on client config
	az.InterfacesClient = interfaceclient.New(interfaceClientConfig)
	az.VirtualMachineSizesClient = vmsizeclient.New(vmSizeClientConfig)
	az.SnapshotsClient = snapshotclient.New(snapshotClientConfig)
	az.StorageAccountClient = storageaccountclient.New(storageAccountClientConfig)
	az.DisksClient = diskclient.New(diskClientConfig)
	az.VirtualMachinesClient = vmclient.New(vmClientConfig)
	az.VirtualMachineScaleSetsClient = vmssclient.New(vmssClientConfig)
	az.VirtualMachineScaleSetVMsClient = vmssvmclient.New(vmssVMClientConfig)
	az.RoutesClient = routeclient.New(routeClientConfig)
	az.SubnetsClient = subnetclient.New(subnetClientConfig)
	az.RouteTablesClient = routetableclient.New(routeTableClientConfig)
	az.LoadBalancerClient = loadbalancerclient.New(loadBalancerClientConfig)
	az.SecurityGroupsClient = securitygroupclient.New(securityGroupClientConfig)
	az.PublicIPAddressesClient = publicipclient.New(publicIPClientConfig)
	az.FileClient = fileclient.New(fileClientConfig)
	az.BlobClient = blobclient.New(blobClientConfig)
	az.AvailabilitySetsClient = vmasclient.New(vmasClientConfig)
	az.privateendpointclient = privateendpointclient.New(privateEndpointConfig)
	az.privatednszonegroupclient = privatednszonegroupclient.New(privateDNSZoenGroupConfig)
	az.PrivateLinkServiceClient = privatelinkserviceclient.New(privateLinkServiceConfig)
	az.containerServiceClient = containerserviceclient.New(containerServiceConfig)
	az.deploymentClient = deploymentclient.New(deploymentConfig)

	if az.ZoneClient == nil {
		az.ZoneClient = zoneclient.New(zoneClientConfig)
	}
}

func (az *Cloud) getAzureClientConfig(servicePrincipalToken *adal.ServicePrincipalToken) *azclients.ClientConfig {
	azClientConfig := &azclients.ClientConfig{
		CloudName:               az.Config.Cloud,
		Location:                az.Config.Location,
		SubscriptionID:          az.Config.SubscriptionID,
		ResourceManagerEndpoint: az.Environment.ResourceManagerEndpoint,
		Authorizer:              autorest.NewBearerAuthorizer(servicePrincipalToken),
		Backoff:                 &retry.Backoff{Steps: 1},
		DisableAzureStackCloud:  az.Config.DisableAzureStackCloud,
		UserAgent:               az.Config.UserAgent,
	}

	if az.Config.CloudProviderBackoff {
		azClientConfig.Backoff = &retry.Backoff{
			Steps:    az.Config.CloudProviderBackoffRetries,
			Factor:   az.Config.CloudProviderBackoffExponent,
			Duration: time.Duration(az.Config.CloudProviderBackoffDuration) * time.Second,
			Jitter:   az.Config.CloudProviderBackoffJitter,
		}
	}

	if az.Config.HasExtendedLocation() {
		azClientConfig.ExtendedLocation = &azclients.ExtendedLocation{
			Name: az.Config.ExtendedLocationName,
			Type: az.Config.ExtendedLocationType,
		}
	}

	return azClientConfig
}

// ParseConfig returns a parsed configuration for an Azure cloudprovider config file
func ParseConfig(configReader io.Reader) (*Config, error) {
	var config Config
	if configReader == nil {
		return nil, nil
	}

	configContents, err := io.ReadAll(configReader)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(configContents, &config)
	if err != nil {
		return nil, err
	}

	// The resource group name may be in different cases from different Azure APIs, hence it is converted to lower here.
	// See more context at https://github.com/kubernetes/kubernetes/issues/71994.
	config.ResourceGroup = strings.ToLower(config.ResourceGroup)

	// these environment variables are injected by workload identity webhook
	if tenantID := os.Getenv("AZURE_TENANT_ID"); tenantID != "" {
		config.TenantID = tenantID
	}
	if clientID := os.Getenv("AZURE_CLIENT_ID"); clientID != "" {
		config.AADClientID = clientID
	}
	if federatedTokenFile := os.Getenv("AZURE_FEDERATED_TOKEN_FILE"); federatedTokenFile != "" {
		config.AADFederatedTokenFile = federatedTokenFile
		config.UseFederatedWorkloadIdentityExtension = true
	}
	return &config, nil
}

func (az *Cloud) isStackCloud() bool {
	return strings.EqualFold(az.Config.Cloud, consts.AzureStackCloudName) && !az.Config.DisableAzureStackCloud
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
	if az.isStackCloud() {
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
	klog.Infof("Setting up informers for Azure cloud provider")
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
					klog.Errorf("Received unexpected object: %v", obj)
					return
				}
				node, ok = deletedState.Obj.(*v1.Node)
				if !ok {
					klog.Errorf("DeletedFinalStateUnknown contained non-Node object: %v", deletedState.Obj)
					return
				}
			}
			az.updateNodeCaches(node, nil)

			klog.V(4).Infof("Removing node %s from VMSet cache.", node.Name)
			_ = az.VMSet.DeleteCacheForNode(node.Name)
		},
	})
	az.nodeInformerSynced = nodeInformer.HasSynced

	az.serviceLister = informerFactory.Core().V1().Services().Lister()

	az.setUpEndpointSlicesInformer(informerFactory)
}

// updateNodeCaches updates local cache for node's zones and external resource groups.
func (az *Cloud) updateNodeCaches(prevNode, newNode *v1.Node) {
	az.nodeCachesLock.Lock()
	defer az.nodeCachesLock.Unlock()

	if prevNode != nil {
		// Remove from nodeNames cache.
		az.nodeNames.Delete(prevNode.ObjectMeta.Name)

		// Remove from nodeZones cache.
		prevZone, ok := prevNode.ObjectMeta.Labels[v1.LabelTopologyZone]
		if ok && az.isAvailabilityZone(prevZone) {
			az.nodeZones[prevZone].Delete(prevNode.ObjectMeta.Name)
			if az.nodeZones[prevZone].Len() == 0 {
				az.nodeZones[prevZone] = nil
			}
		}

		// Remove from nodeResourceGroups cache.
		_, ok = prevNode.ObjectMeta.Labels[consts.ExternalResourceGroupLabel]
		if ok {
			delete(az.nodeResourceGroups, prevNode.ObjectMeta.Name)
		}

		managed, ok := prevNode.ObjectMeta.Labels[consts.ManagedByAzureLabel]
		isNodeManagedByCloudProvider := !ok || !strings.EqualFold(managed, consts.NotManagedByAzureLabelValue)

		klog.Infof("managed=%v, ok=%v, isNodeManagedByCloudProvider=%v",
			managed, ok, isNodeManagedByCloudProvider)

		// Remove from unmanagedNodes cache
		if !isNodeManagedByCloudProvider {
			az.unmanagedNodes.Delete(prevNode.ObjectMeta.Name)
		}

		// Remove from nodePrivateIPs cache.
		for _, address := range getNodePrivateIPAddresses(prevNode) {
			klog.V(6).Infof("removing IP address %s of the node %s", address, prevNode.Name)
			az.nodePrivateIPs[prevNode.Name].Delete(address)
			delete(az.nodePrivateIPToNodeNameMap, address)
		}

		// if the node is being deleted from the cluster, exclude it from load balancers
		if newNode == nil {
			az.excludeLoadBalancerNodes.Insert(prevNode.ObjectMeta.Name)
			az.nodesWithCorrectLoadBalancerByPrimaryVMSet.Delete(strings.ToLower(prevNode.ObjectMeta.Name))
			delete(az.nodePrivateIPs, strings.ToLower(prevNode.Name))
		}
	}

	if newNode != nil {
		// Add to nodeNames cache.
		az.nodeNames = utilsets.SafeInsert(az.nodeNames, newNode.ObjectMeta.Name)

		// Add to nodeZones cache.
		newZone, ok := newNode.ObjectMeta.Labels[v1.LabelTopologyZone]
		if ok && az.isAvailabilityZone(newZone) {
			az.nodeZones[newZone] = utilsets.SafeInsert(az.nodeZones[newZone], newNode.ObjectMeta.Name)
		}

		// Add to nodeResourceGroups cache.
		newRG, ok := newNode.ObjectMeta.Labels[consts.ExternalResourceGroupLabel]
		if ok && len(newRG) > 0 {
			az.nodeResourceGroups[newNode.ObjectMeta.Name] = strings.ToLower(newRG)
		}

		_, hasExcludeBalancerLabel := newNode.ObjectMeta.Labels[v1.LabelNodeExcludeBalancers]
		managed, ok := newNode.ObjectMeta.Labels[consts.ManagedByAzureLabel]
		isNodeManagedByCloudProvider := !ok || !strings.EqualFold(managed, consts.NotManagedByAzureLabelValue)

		// Update unmanagedNodes cache
		if !isNodeManagedByCloudProvider {
			az.unmanagedNodes.Insert(newNode.ObjectMeta.Name)
		}

		// Update excludeLoadBalancerNodes cache
		switch {
		case !isNodeManagedByCloudProvider:
			az.excludeLoadBalancerNodes.Insert(newNode.ObjectMeta.Name)
			klog.V(6).Infof("excluding Node %q from LoadBalancer because it is not managed by cloud provider", newNode.ObjectMeta.Name)

		case hasExcludeBalancerLabel:
			az.excludeLoadBalancerNodes.Insert(newNode.ObjectMeta.Name)
			klog.V(6).Infof("excluding Node %q from LoadBalancer because it has exclude-from-external-load-balancers label", newNode.ObjectMeta.Name)

		default:
			// Nodes not falling into the three cases above are valid backends and
			// should not appear in excludeLoadBalancerNodes cache.
			az.excludeLoadBalancerNodes.Delete(newNode.ObjectMeta.Name)
		}

		// Add to nodePrivateIPs cache
		for _, address := range getNodePrivateIPAddresses(newNode) {
			if az.nodePrivateIPToNodeNameMap == nil {
				az.nodePrivateIPToNodeNameMap = make(map[string]string)
			}

			klog.V(6).Infof("adding IP address %s of the node %s", address, newNode.Name)
			az.nodePrivateIPs[strings.ToLower(newNode.Name)] = utilsets.SafeInsert(az.nodePrivateIPs[strings.ToLower(newNode.Name)], address)
			az.nodePrivateIPToNodeNameMap[address] = newNode.Name
		}
	}
}

// updateNodeTaint updates node out-of-service taint
func (az *Cloud) updateNodeTaint(node *v1.Node) {
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
			klog.Errorf("failed to remove taint %s from the node %s", v1.TaintNodeOutOfService, node.Name)
		}
	} else {
		// node shutdown taint is added when cloud provider determines instance is shutdown
		if !taints.TaintExists(node.Spec.Taints, nodeOutOfServiceTaint) &&
			taints.TaintExists(node.Spec.Taints, nodeShutdownTaint) {
			klog.V(2).Infof("adding %s taint to node %s", v1.TaintNodeOutOfService, node.Name)
			if err := cloudnodeutil.AddOrUpdateTaintOnNode(az.KubeClient, node.Name, nodeOutOfServiceTaint); err != nil {
				klog.Errorf("failed to add taint %s to the node %s", v1.TaintNodeOutOfService, node.Name)
			}
		} else {
			klog.V(2).Infof("node %s is not ready but either shutdown taint is missing or out-of-service taint is already added, skip adding node out-of-service taint", node.Name)
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
func (az *Cloud) GetNodeVMSet(nodeName types.NodeName, crt azcache.AzureCacheReadType) (VMSet, error) {
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

	vmManagementType, err := ss.getVMManagementTypeByNodeName(string(nodeName), crt)
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
