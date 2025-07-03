//go:build coreweave
// +build coreweave

package builder

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/coreweave"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/client-go/informers"
)

// AvailableCloudProviders supported by the cloud provider builder.
var AvailableCloudProviders = []string{
	cloudprovider.CoreWeaveProviderName,
}

// DefaultCloudProvider for coreweave-only build is coreweave.
const DefaultCloudProvider = cloudprovider.CoreWeaveProviderName


func buildCloudProvider(opts config.AutoscalingOptions, do cloudprovider.NodeGroupDiscoveryOptions, rl *cloudprovider.ResourceLimiter, _ informers.SharedInformerFactory) cloudprovider.CloudProvider {
	return coreweave.BuildCoreWeave(opts, do, rl)
}
