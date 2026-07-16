/*
Copyright 2024 The Kubernetes Authors.

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

package config

import (
	"strings"

	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/configloader"
	"sigs.k8s.io/cloud-provider-azure/pkg/consts"
)

// Config holds the configuration parsed from the --cloud-config flag
// All fields are required unless otherwise specified
// NOTE: Cloud config files should follow the same Kubernetes deprecation policy as
// flags or CLIs. Config fields should not change behavior in incompatible ways and
// should be deprecated for at least 2 release prior to removing.
// See https://kubernetes.io/docs/reference/using-api/deprecation-policy/#deprecating-a-flag-or-cli
// for more details.
type Config struct {
	AzureClientConfig `json:",inline" yaml:",inline"`

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
	// SystemTags now support prefix match, which means that if a key in `SystemTags` is a prefix of a key in `Tags`, that tag will not be deleted
	SystemTags string `json:"systemTags,omitempty" yaml:"systemTags,omitempty"`
	// Sku of Load Balancer and Public IP. Candidate values are: basic and standard.
	// If not set, it will be default to basic.
	LoadBalancerSKU string `json:"loadBalancerSku,omitempty" yaml:"loadBalancerSku,omitempty"`
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

	// LoadBalancerBackendPoolConfigurationType defines how vms join the load balancer backend pools. Supported values
	// are `nodeIPConfiguration`, `nodeIP` and `podIP`.
	// `nodeIPConfiguration`: vm network interfaces will be attached to the inbound backend pool of the load balancer (default);
	// `nodeIP`: vm private IPs will be attached to the inbound backend pool of the load balancer;
	// `podIP`: pod IPs will be attached to the inbound backend pool of the load balancer (not supported yet).
	LoadBalancerBackendPoolConfigurationType string `json:"loadBalancerBackendPoolConfigurationType,omitempty" yaml:"loadBalancerBackendPoolConfigurationType,omitempty"`
	// PutVMSSVMBatchSize defines how many requests the client send concurrently when putting the VMSS VMs.
	// If it is smaller than or equal to one, the request will be sent one by one in sequence (default).
	PutVMSSVMBatchSize int `json:"putVMSSVMBatchSize" yaml:"putVMSSVMBatchSize"`
	// PrivateLinkServiceResourceGroup determines the specific resource group of the private link services user want to use
	PrivateLinkServiceResourceGroup string `json:"privateLinkServiceResourceGroup,omitempty" yaml:"privateLinkServiceResourceGroup,omitempty"`

	// EnableMigrateToIPBasedBackendPoolAPI uses the migration API to migrate from NIC-based to IP-based backend pool.
	// The migration API can provide a migration from NIC-based to IP-based backend pool without service downtime.
	// If the API is not used, the migration will be done by decoupling all nodes on the backend pool and then re-attaching
	// node IPs, which will introduce service downtime. The downtime increases with the number of nodes in the backend pool.
	EnableMigrateToIPBasedBackendPoolAPI bool `json:"enableMigrateToIPBasedBackendPoolAPI" yaml:"enableMigrateToIPBasedBackendPoolAPI"`

	// EnableIPTagMutationForExistingPublicIP enables in-place mutation of
	// FirstPartyUsage IP tags on existing public IPs. When enabled, only
	// FirstPartyUsage-type IP tags can be updated in-place; attempting to
	// change other IP tag types (e.g. RoutingPreference) on an existing PIP
	// produces a reconciliation error.
	EnableIPTagMutationForExistingPublicIP bool `json:"enableIPTagMutationForExistingPublicIP,omitempty" yaml:"enableIPTagMutationForExistingPublicIP,omitempty"`

	// MultipleStandardLoadBalancerConfigurations stores the properties regarding multiple standard load balancers.
	// It will be ignored if LoadBalancerBackendPoolConfigurationType is nodeIPConfiguration.
	// If the length is not 0, it is assumed the multiple standard load balancers mode is on. In this case,
	// there must be one configuration named "<clustername>" or an error will be reported.
	MultipleStandardLoadBalancerConfigurations []MultipleStandardLoadBalancerConfiguration `json:"multipleStandardLoadBalancerConfigurations,omitempty" yaml:"multipleStandardLoadBalancerConfigurations,omitempty"`

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

// HasExtendedLocation returns true if extendedlocation prop are specified.
func (az *Config) HasExtendedLocation() bool {
	return az.ExtendedLocationName != "" && az.ExtendedLocationType != ""
}

func (az *Config) IsLBBackendPoolTypeNodeIPConfig() bool {
	return strings.EqualFold(az.LoadBalancerBackendPoolConfigurationType, consts.LoadBalancerBackendPoolConfigurationTypeNodeIPConfiguration)
}

func (az *Config) IsLBBackendPoolTypeNodeIP() bool {
	return strings.EqualFold(az.LoadBalancerBackendPoolConfigurationType, consts.LoadBalancerBackendPoolConfigurationTypeNodeIP)
}

func (az *Config) GetPutVMSSVMBatchSize() int {
	return az.PutVMSSVMBatchSize
}

func (az *Config) UseStandardLoadBalancer() bool {
	return strings.EqualFold(az.LoadBalancerSKU, consts.LoadBalancerSKUStandard)
}

func (az *Config) ExcludeMasterNodesFromStandardLB() bool {
	return az.ExcludeMasterFromStandardLB != nil && *az.ExcludeMasterFromStandardLB
}

func (az *Config) DisableLoadBalancerOutboundSNAT() bool {
	if !az.UseStandardLoadBalancer() || az.DisableOutboundSNAT == nil {
		return false
	}

	return *az.DisableOutboundSNAT
}

func (az *Config) UseMultipleStandardLoadBalancers() bool {
	return az.UseStandardLoadBalancer() && len(az.MultipleStandardLoadBalancerConfigurations) > 0
}

func (az *Config) UseSingleStandardLoadBalancer() bool {
	return az.UseStandardLoadBalancer() && len(az.MultipleStandardLoadBalancerConfigurations) == 0
}

func (az *Config) IsStackCloud() bool {
	return strings.EqualFold(az.Cloud, consts.AzureStackCloudName) && !az.DisableAzureStackCloud
}
