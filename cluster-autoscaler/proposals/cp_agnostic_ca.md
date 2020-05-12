# Decoupling ClusterAutoScalerCore (separate library) from CloudProvider implementations
##### Author: bhargav-naik

### Introduction
Current ClusterAutoscaler is released as a bundle which includes core and all the CloudProviders included.
When someone intends to write implementation of the CloudProvider they need to fork the intended version ClusterAutoscaler and provide their implementation.
In this proposal, a different design approach is presented which will decouple the ClusterAutoscaler core from CloudProviders and make ClusterAutoScalerCore available as a separate release. 

### Issues
- In the current design the ClusterAutoscaler core has dependency on the CloudProvider implementations.
- For proprietary CloudProvider implementations it results in an development cycle to fork new version from public repo and port the CloudProvider implementation to the latest fork.
- Also the ClusterAutoscaler library comes with a lot of unnecessary CloudProvider specific dependencies which might not be required for other CloudProviders.

### Solution
- Release ClusterAutoscalerCore as a separate module which won't have any Cloudprovider specific code/dependencies.
- Release CloudProvider specific ClusterAutoscaler[PROVIDER] as separate module. This module depends on specific version of ClusterAutoscaler.
- For proprietary CloudProvider implementations can depend on ClusterAutoscalerCore module and wont get any other CloudProvider specific dependencies.
- Also the development cycle for CloudProvider implementation reduces to just a version bump of ClusterAutoscalerCore if the CloudProvider,Nodegroup interfaces are intact and CloudProvider doesn't intend to add any feature.

### Implementation
- Introduce a Registration mechanism in ClusterAutoScalerCore which the CloudProvider can use to register the implementation at init.
```go
type CloudProviderFactory func(opts config.AutoscalingOptions, do NodeGroupDiscoveryOptions, rl *ResourceLimiter) CloudProvider

// All registered cloud providers.
var (
	cloudProvidersMutex           sync.Mutex
	cloudProviders                = make(map[string]CloudProviderFactory)
)

func RegisterCloudProvider(name string, cloud CloudProviderFactory) {
	cloudProvidersMutex.Lock()
	defer cloudProvidersMutex.Unlock()
	if _, found := cloudProviders[name]; found {
		klog.Fatalf("Cloud provider %q was registered twice", name)
	}
	cloudProviders[name] = cloud
	klog.V(1).Infof("Registered cloud provider %q", name)
}

func GetCloudProvider(name string, opts config.AutoscalingOptions, do NodeGroupDiscoveryOptions, rl *ResourceLimiter) CloudProvider {
	cloudProvidersMutex.Lock()
	defer cloudProvidersMutex.Unlock()
	f, found := cloudProviders[name]
	if !found {
		return nil
	}
	return f(opts, do, rl)
}

func GetAvailiableCloudProviderNames() []string {
	cloudProvidersMutex.Lock()
	defer cloudProvidersMutex.Unlock()
	keys := []string{}
	for k := range cloudProviders {
		keys = append(keys, k)
	}
	return keys
}
```
- Move CloudProvider specific code out of clusterAutoScalerCore and let each of them have their own go module (ClusterAutoScale[AWS]) which depends on ClusterAutoScalerCore.

- CloudProvider can provide an init block in which they can register their implementation.
```go
package aws

func init() {
	cloudprovider.RegisterCloudProvider("aws", aws.BuildAWS)
}
