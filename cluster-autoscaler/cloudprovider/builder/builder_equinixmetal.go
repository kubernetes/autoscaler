//go:build equinixmetal
// +build equinixmetal

/*
Copyright 2019 The Kubernetes Authors.

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
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/equinixmetal"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/client-go/informers"
)

// AvailableCloudProviders supported by the cloud provider builder.
var AvailableCloudProviders = []string{
	cloudprovider.PacketProviderName,
	cloudprovider.EquinixMetalProviderName,
}

// DefaultCloudProvider for Packet or Equinix Metal-only build is Equinix Metal.
const DefaultCloudProvider = cloudprovider.EquinixMetalProviderName

func buildCloudProvider(opts config.AutoscalingOptions, do cloudprovider.NodeGroupDiscoveryOptions, rl *cloudprovider.ResourceLimiter, _ informers.SharedInformerFactory) cloudprovider.CloudProvider {
	switch opts.CloudProviderName {
	case cloudprovider.PacketProviderName, cloudprovider.EquinixMetalProviderName:
		return equinixmetal.BuildCloudProvider(opts, do, rl)
	}

	return nil
}
