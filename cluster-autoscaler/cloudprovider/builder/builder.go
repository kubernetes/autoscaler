/*
Copyright 2018 The Kubernetes Authors.

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

package builder

import (
	"sort"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	coreoptions "k8s.io/autoscaler/cluster-autoscaler/core/options"
	"k8s.io/client-go/informers"
	"k8s.io/klog/v2"
)

// CloudProviderBuilder builds a cloud provider from provided parameters.
type CloudProviderBuilder func(opts *coreoptions.AutoscalerOptions, do cloudprovider.NodeGroupDiscoveryOptions, rl *cloudprovider.ResourceLimiter, informerFactory informers.SharedInformerFactory) cloudprovider.CloudProvider

var (
	cloudProviderBuilders = make(map[string]CloudProviderBuilder)
	defaultCloudProvider  string
)

// RegisterCloudProvider registers a cloud provider builder.
func RegisterCloudProvider(name string, builder CloudProviderBuilder) {
	if _, ok := cloudProviderBuilders[name]; ok {
		klog.Fatalf("Cloud provider %s already registered", name)
	}
	cloudProviderBuilders[name] = builder
}

// GetCloudProviderBuilder returns a cloud provider builder by name.
func GetCloudProviderBuilder(name string) (CloudProviderBuilder, bool) {
	builder, ok := cloudProviderBuilders[name]
	return builder, ok
}

// AvailableCloudProviders returns the list of supported cloud providers.
func AvailableCloudProviders() []string {
	providers := make([]string, 0, len(cloudProviderBuilders))
	for name := range cloudProviderBuilders {
		providers = append(providers, name)
	}
	sort.Strings(providers)
	return providers
}

// SetDefaultCloudProvider sets the default cloud provider name.
func SetDefaultCloudProvider(name string) {
	defaultCloudProvider = name
}

// GetDefaultCloudProvider returns the default cloud provider name.
func GetDefaultCloudProvider() string {
	return defaultCloudProvider
}

// DefaultCloudProvider returns the default cloud provider name.
func DefaultCloudProvider() string {
	if def := GetDefaultCloudProvider(); def != "" {
		return def
	}
	return cloudprovider.GceProviderName
}
