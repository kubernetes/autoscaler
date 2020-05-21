# Decoupling ClusterAutoScalerCore (separate library) from CloudProvider implementations
##### Author: bhargav-naik

### Introduction
Current ClusterAutoscaler is released as a bundle which includes core and all the CloudProviders included.
When someone intends to write implementation of the CloudProvider they need to fork the intended version ClusterAutoscaler and provide their implementation.
In this proposal, a different design approach is presented which will decouple the ClusterAutoscaler core from CloudProviders and make ClusterAutoScalerCore available as a separate release. 

### Issues
- In the current design the ClusterAutoscaler core has dependency on the CloudProvider implementations.
- For proprietary CloudProvider implementations it results in an development cycle to fork new version from public repo and port the CloudProvider implementation to the latest fork.
- Also the ClusterAutoscaler library comes with a lot of unnecessary CloudProvider specific dependencies which might not be required for other CloudProviders. (We dont want to address this)

### Solution
Version 1:  (Not acceptable, as we don't want to disrupt the existing dev/build cycle)
- Release ClusterAutoscalerCore as a separate module which won't have any Cloudprovider specific code/dependencies.
- Release CloudProvider specific ClusterAutoscaler[PROVIDER] as separate module. This module depends on specific version of ClusterAutoscaler.
- For proprietary CloudProvider implementations can depend on ClusterAutoscalerCore module and wont get any other CloudProvider specific dependencies.
- Also the development cycle for CloudProvider implementation reduces to just a version bump of ClusterAutoscalerCore if the CloudProvider,Nodegroup interfaces are intact and CloudProvider doesn't intend to add any feature.

Version 2:
- Let the existing cloudProvider code remain in the current hierarchy.
- Introduce external_cloud_provider.go under cloudprovider.external package: A shim layer which will give external_cloud_provider mechanism to register their implementation.

### Implementation
```go
package external

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/klog"
	"sync"
)

type CloudProviderFactory func(opts config.AutoscalingOptions, do cloudprovider.NodeGroupDiscoveryOptions, rl *cloudprovider.ResourceLimiter) cloudprovider.CloudProvider

var (
	mutex                         sync.Mutex
	cloudProviderFactory          CloudProviderFactory
)

func RegisterExternalCloudProvider(cpf CloudProviderFactory) {
	mutex.Lock()
	defer mutex.Unlock()
	if cloudProviderFactory != nil {
		klog.Fatalf("Cloud provider %q was registered twice")
	}
	cloudProviderFactory = cpf
}

func BuildExternalCloudProvider(opts config.AutoscalingOptions, do cloudprovider.NodeGroupDiscoveryOptions, rl *cloudprovider.ResourceLimiter) cloudprovider.CloudProvider {
	mutex.Lock()
	defer mutex.Unlock()
	if cloudProviderFactory != nil {
		klog.Fatalf("external cloudProvider not registered")
	}
	return cloudProviderFactory(opts, do, rl)
}
```

- Modify buidler_all.go
```go
// +build  !external !gce,!aws,!azure,!kubemark,!alicloud,!magnum,!digitalocean,!clusterapi
package builder

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/alicloud"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/azure"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/baiducloud"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/clusterapi"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/digitalocean"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/external"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/gce"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/magnum"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/packet"
	"k8s.io/autoscaler/cluster-autoscaler/config"
)

// AvailableCloudProviders supported by the cloud provider builder.
var AvailableCloudProviders = []string{
	cloudprovider.AwsProviderName,
	cloudprovider.AzureProviderName,
	cloudprovider.GceProviderName,
	cloudprovider.AlicloudProviderName,
	cloudprovider.BaiducloudProviderName,
	cloudprovider.MagnumProviderName,
	cloudprovider.DigitalOceanProviderName,
	cloudprovider.ExternalCloudProviderName,
	clusterapi.ProviderName,
}

// DefaultCloudProvider is GCE.
const DefaultCloudProvider = cloudprovider.GceProviderName

func buildCloudProvider(opts config.AutoscalingOptions, do cloudprovider.NodeGroupDiscoveryOptions, rl *cloudprovider.ResourceLimiter) cloudprovider.CloudProvider {
	switch opts.CloudProviderName {
	case cloudprovider.GceProviderName:
		return gce.BuildGCE(opts, do, rl)
	case cloudprovider.AwsProviderName:
		return aws.BuildAWS(opts, do, rl)
	case cloudprovider.AzureProviderName:
		return azure.BuildAzure(opts, do, rl)
	case cloudprovider.AlicloudProviderName:
		return alicloud.BuildAlicloud(opts, do, rl)
	case cloudprovider.BaiducloudProviderName:
		return baiducloud.BuildBaiducloud(opts, do, rl)
	case cloudprovider.DigitalOceanProviderName:
		return digitalocean.BuildDigitalOcean(opts, do, rl)
	case cloudprovider.MagnumProviderName:
		return magnum.BuildMagnum(opts, do, rl)
	case packet.ProviderName:
		return packet.BuildPacket(opts, do, rl)
	case clusterapi.ProviderName:
		return clusterapi.BuildClusterAPI(opts, do, rl)
	case cloudprovider.ExternalCloudProviderName:
		return external.BuildExternalCloudProvider(opts, do, rl)
	}
	return nil
}
```

- Introduce builder_external.go
```go
// +build external

package builder

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/external"
	"k8s.io/autoscaler/cluster-autoscaler/config"
)

// AvailableCloudProviders supported by the cloud provider builder.
var AvailableCloudProviders = []string{
	cloudprovider.ExternalCloudProviderName,
}

const DefaultCloudProvider = cloudprovider.ExternalCloudProviderName

func buildCloudProvider(opts config.AutoscalingOptions, do cloudprovider.NodeGroupDiscoveryOptions, rl *cloudprovider.ResourceLimiter) cloudprovider.CloudProvider {
		return external.BuildExternalCloudProvider(ops,do,rl)
}
```

- External cloudprovider will register their implementation using init block. (in their own repo)
```go
package ecp
//ecp: ExternalCloudProvider
import "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/external"

func init() {
	extenral.RegisterExternalCloudProvider("aws", ecp.BuildCloudProvider)
}
```
