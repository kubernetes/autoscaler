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

type CloudProviderCacheConfig struct {
	// DisableAPICallCache disables the cache for Azure API calls. It is for ARG support and not all resources will be disabled.
	DisableAPICallCache bool `json:"disableAPICallCache,omitempty" yaml:"disableAPICallCache,omitempty"`
	// NonVmssUniformNodesCacheTTLInSeconds sets the Cache TTL for NonVmssUniformNodesCacheTTLInSeconds
	// if not set, will use default value
	NonVmssUniformNodesCacheTTLInSeconds int `json:"nonVmssUniformNodesCacheTTLInSeconds,omitempty" yaml:"nonVmssUniformNodesCacheTTLInSeconds,omitempty"`
	// VmssCacheTTLInSeconds sets the cache TTL for VMSS
	VmssCacheTTLInSeconds int `json:"vmssCacheTTLInSeconds,omitempty" yaml:"vmssCacheTTLInSeconds,omitempty"`
	// VmssVirtualMachinesCacheTTLInSeconds sets the cache TTL for vmssVirtualMachines
	VmssVirtualMachinesCacheTTLInSeconds int `json:"vmssVirtualMachinesCacheTTLInSeconds,omitempty" yaml:"vmssVirtualMachinesCacheTTLInSeconds,omitempty"`
	// ListVmssVirtualMachinesWithoutInstanceView makes the provider list VMSS virtual machines without instance view.
	ListVmssVirtualMachinesWithoutInstanceView bool `json:"listVmssVirtualMachinesWithoutInstanceView,omitempty" yaml:"listVmssVirtualMachinesWithoutInstanceView,omitempty"`

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
}
