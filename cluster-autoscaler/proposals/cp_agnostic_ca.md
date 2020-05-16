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
package cloudprovider

type CloudProviderFactory func(opts config.AutoscalingOptions, do NodeGroupDiscoveryOptions, rl *ResourceLimiter) CloudProvider

// All registered cloud providers.
var (
	mutex                         sync.Mutex
	cloudProviderFactory          CloudProviderFactory
	podListProcessor              pods.PodListProcessor
	nodeGroupListProcessor        nodegroups.NodeGroupListProcessor
	nodeGroupSetProcessor 		  nodegroupset.NodeGroupSetProcessor
	scaleUpStatusProcessor 		  status.ScaleUpStatusProcessor
	scaleDownNodeProcessor 		  nodes.ScaleDownNodeProcessor
	scaleDownStatusProcessor 	  status.ScaleDownStatusProcessor
	autoscalingStatusProcessor    status.AutoscalingStatusProcessor
	nodeGroupManager 			  nodegroups.NodeGroupManager
	nodeInfoProcessor 			  nodeinfos.NodeInfoProcessor
)

func RegisterCloudProviderFactory(cloudProviderFactory CloudProviderFactory) {
	mutex.Lock()
	defer mutex.Unlock()
	if cloudProviderFactory != nil {
		klog.Fatalf("Cloud provider %q was registered twice", name)
	}
	cloudProviderFactory = cloudProviderFactory
}

func GetCloudProvider(name string, opts config.AutoscalingOptions, do NodeGroupDiscoveryOptions, rl *ResourceLimiter) CloudProvider {
	cloudProvidersMutex.Lock()
	defer cloudProvidersMutex.Unlock()
	if cloudProviderFactory != nil {
	    return cloudProviderFactory(opts, do, rl)
	}
	return nil
}

func RegisterPodListProcessor(podListProcessor pods.PodListProcessor) {
	mutex.Lock()
	defer mutex.Unlock()
	podListProcessor = podListProcessor
}

func GetPodListProcessor() pods.PodListProcessor {
	mutex.Lock()
	defer mutex.Unlock()
	return podListProcessor
}

func RegisterNodeGroupListProcessor(nodeGroupListProcessor nodegroups.NodeGroupListProcessor) {
	mutex.Lock()
	defer mutex.Unlock()
	nodeGroupListProcessor = nodeGroupListProcessor
}

func GetNodeGroupListProcessor() nodegroups.NodeGroupListProcessor {
	mutex.Lock()
	defer mutex.Unlock()
	return nodeGroupListProcessor
}

// Similarly add methods to register other processor which can be used by cloudprovider to override at boot time

```
- Move CloudProvider specific code out of clusterAutoScalerCore and let each of them have their own go module (ClusterAutoScale[AWS]) which depends on ClusterAutoScalerCore.

- CloudProvider can provide an init block in which they can register their implementation.
```go
package aws

func init() {
	cloudprovider.RegisterCloudProvider("aws", aws.BuildAWS)
    cloudprovider.RegisterNodeGroupListProcessor(AWSNodeGroupListProcessorImplementation)
}
```
- In the main class after default processor are initialized, have a block of code to check and override processor in case the cloudProvider has registered any implementation.
```main.go

302	opts.Processors = ca_processors.DefaultProcessors()
303     if cloudprovider.GetNodeGroupListProcessor != nil {
304        opts.Processors.NodeGroupListProcessor = cloudprovider.GetNodeGroupListProcessor
305     }

// Similarly have a block to check and override the processor if any proccessor is registered by the cloudProvider in init block
```

- Move cloudProvider specific nodeGroup implementation to their distribution.
