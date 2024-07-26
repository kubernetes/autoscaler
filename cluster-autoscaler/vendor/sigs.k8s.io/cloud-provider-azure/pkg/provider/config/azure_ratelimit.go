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

package config

import (
	azclients "sigs.k8s.io/cloud-provider-azure/pkg/azureclients"
	"sigs.k8s.io/cloud-provider-azure/pkg/consts"
)

const (
	DefaultAtachDetachDiskQPS    = 6.0
	DefaultAtachDetachDiskBucket = 10
)

// CloudProviderRateLimitConfig indicates the rate limit config for each clients.
type CloudProviderRateLimitConfig struct {
	// The default rate limit config options.
	azclients.RateLimitConfig `json:",inline" yaml:",inline"`

	// Rate limit config for each clients. Values would override default settings above.
	RouteRateLimit                  *azclients.RateLimitConfig `json:"routeRateLimit,omitempty" yaml:"routeRateLimit,omitempty"`
	SubnetsRateLimit                *azclients.RateLimitConfig `json:"subnetsRateLimit,omitempty" yaml:"subnetsRateLimit,omitempty"`
	InterfaceRateLimit              *azclients.RateLimitConfig `json:"interfaceRateLimit,omitempty" yaml:"interfaceRateLimit,omitempty"`
	RouteTableRateLimit             *azclients.RateLimitConfig `json:"routeTableRateLimit,omitempty" yaml:"routeTableRateLimit,omitempty"`
	LoadBalancerRateLimit           *azclients.RateLimitConfig `json:"loadBalancerRateLimit,omitempty" yaml:"loadBalancerRateLimit,omitempty"`
	PublicIPAddressRateLimit        *azclients.RateLimitConfig `json:"publicIPAddressRateLimit,omitempty" yaml:"publicIPAddressRateLimit,omitempty"`
	SecurityGroupRateLimit          *azclients.RateLimitConfig `json:"securityGroupRateLimit,omitempty" yaml:"securityGroupRateLimit,omitempty"`
	VirtualMachineRateLimit         *azclients.RateLimitConfig `json:"virtualMachineRateLimit,omitempty" yaml:"virtualMachineRateLimit,omitempty"`
	StorageAccountRateLimit         *azclients.RateLimitConfig `json:"storageAccountRateLimit,omitempty" yaml:"storageAccountRateLimit,omitempty"`
	DiskRateLimit                   *azclients.RateLimitConfig `json:"diskRateLimit,omitempty" yaml:"diskRateLimit,omitempty"`
	SnapshotRateLimit               *azclients.RateLimitConfig `json:"snapshotRateLimit,omitempty" yaml:"snapshotRateLimit,omitempty"`
	VirtualMachineScaleSetRateLimit *azclients.RateLimitConfig `json:"virtualMachineScaleSetRateLimit,omitempty" yaml:"virtualMachineScaleSetRateLimit,omitempty"`
	VirtualMachineSizeRateLimit     *azclients.RateLimitConfig `json:"virtualMachineSizesRateLimit,omitempty" yaml:"virtualMachineSizesRateLimit,omitempty"`
	AvailabilitySetRateLimit        *azclients.RateLimitConfig `json:"availabilitySetRateLimit,omitempty" yaml:"availabilitySetRateLimit,omitempty"`
	AttachDetachDiskRateLimit       *azclients.RateLimitConfig `json:"attachDetachDiskRateLimit,omitempty" yaml:"attachDetachDiskRateLimit,omitempty"`
	ContainerServiceRateLimit       *azclients.RateLimitConfig `json:"containerServiceRateLimit,omitempty" yaml:"containerServiceRateLimit,omitempty"`
	DeploymentRateLimit             *azclients.RateLimitConfig `json:"deploymentRateLimit,omitempty" yaml:"deploymentRateLimit,omitempty"`
	PrivateDNSRateLimit             *azclients.RateLimitConfig `json:"privateDNSRateLimit,omitempty" yaml:"privateDNSRateLimit,omitempty"`
	PrivateDNSZoneGroupRateLimit    *azclients.RateLimitConfig `json:"privateDNSZoneGroupRateLimit,omitempty" yaml:"privateDNSZoneGroupRateLimit,omitempty"`
	PrivateEndpointRateLimit        *azclients.RateLimitConfig `json:"privateEndpointRateLimit,omitempty" yaml:"privateEndpointRateLimit,omitempty"`
	PrivateLinkServiceRateLimit     *azclients.RateLimitConfig `json:"privateLinkServiceRateLimit,omitempty" yaml:"privateLinkServiceRateLimit,omitempty"`
	VirtualNetworkRateLimit         *azclients.RateLimitConfig `json:"virtualNetworkRateLimit,omitempty" yaml:"virtualNetworkRateLimit,omitempty"`
}

// InitializeCloudProviderRateLimitConfig initializes rate limit configs.
func InitializeCloudProviderRateLimitConfig(config *CloudProviderRateLimitConfig) {
	if config == nil {
		return
	}

	// Assign read rate limit defaults if no configuration was passed in.
	if config.CloudProviderRateLimitQPS == 0 {
		config.CloudProviderRateLimitQPS = consts.RateLimitQPSDefault
	}
	if config.CloudProviderRateLimitBucket == 0 {
		config.CloudProviderRateLimitBucket = consts.RateLimitBucketDefault
	}
	// Assign write rate limit defaults if no configuration was passed in.
	if config.CloudProviderRateLimitQPSWrite == 0 {
		config.CloudProviderRateLimitQPSWrite = config.CloudProviderRateLimitQPS
	}
	if config.CloudProviderRateLimitBucketWrite == 0 {
		config.CloudProviderRateLimitBucketWrite = config.CloudProviderRateLimitBucket
	}

	config.RouteRateLimit = overrideDefaultRateLimitConfig(&config.RateLimitConfig, config.RouteRateLimit)
	config.SubnetsRateLimit = overrideDefaultRateLimitConfig(&config.RateLimitConfig, config.SubnetsRateLimit)
	config.InterfaceRateLimit = overrideDefaultRateLimitConfig(&config.RateLimitConfig, config.InterfaceRateLimit)
	config.RouteTableRateLimit = overrideDefaultRateLimitConfig(&config.RateLimitConfig, config.RouteTableRateLimit)
	config.LoadBalancerRateLimit = overrideDefaultRateLimitConfig(&config.RateLimitConfig, config.LoadBalancerRateLimit)
	config.PublicIPAddressRateLimit = overrideDefaultRateLimitConfig(&config.RateLimitConfig, config.PublicIPAddressRateLimit)
	config.SecurityGroupRateLimit = overrideDefaultRateLimitConfig(&config.RateLimitConfig, config.SecurityGroupRateLimit)
	config.VirtualMachineRateLimit = overrideDefaultRateLimitConfig(&config.RateLimitConfig, config.VirtualMachineRateLimit)
	config.StorageAccountRateLimit = overrideDefaultRateLimitConfig(&config.RateLimitConfig, config.StorageAccountRateLimit)
	config.DiskRateLimit = overrideDefaultRateLimitConfig(&config.RateLimitConfig, config.DiskRateLimit)
	config.SnapshotRateLimit = overrideDefaultRateLimitConfig(&config.RateLimitConfig, config.SnapshotRateLimit)
	config.VirtualMachineScaleSetRateLimit = overrideDefaultRateLimitConfig(&config.RateLimitConfig, config.VirtualMachineScaleSetRateLimit)
	config.VirtualMachineSizeRateLimit = overrideDefaultRateLimitConfig(&config.RateLimitConfig, config.VirtualMachineSizeRateLimit)
	config.AvailabilitySetRateLimit = overrideDefaultRateLimitConfig(&config.RateLimitConfig, config.AvailabilitySetRateLimit)

	atachDetachDiskRateLimitConfig := azclients.RateLimitConfig{
		CloudProviderRateLimit:            true,
		CloudProviderRateLimitQPSWrite:    DefaultAtachDetachDiskQPS,
		CloudProviderRateLimitBucketWrite: DefaultAtachDetachDiskBucket,
	}
	config.AttachDetachDiskRateLimit = overrideDefaultRateLimitConfig(&atachDetachDiskRateLimitConfig, config.AttachDetachDiskRateLimit)
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
