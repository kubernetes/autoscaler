/*
Copyright 2016 The Kubernetes Authors.

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
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/context"

	klog "k8s.io/klog/v2"
)

// NewCloudProvider builds a cloud provider from provided parameters.
func NewCloudProvider(opts config.AutoscalingOptions) cloudprovider.CloudProvider {
	klog.V(1).Infof("Building %s cloud provider.", opts.CloudProviderName)

	do := cloudprovider.NodeGroupDiscoveryOptions{
		NodeGroupSpecs:              opts.NodeGroups,
		NodeGroupAutoDiscoverySpecs: opts.NodeGroupAutoDiscovery,
	}

	rl := context.NewResourceLimiterFromAutoscalingOptions(opts)

	if opts.CloudProviderName == "" {
		// Ideally this would be an error, but several unit tests of the
		// StaticAutoscaler depend on this behaviour.
		klog.Warning("Returning a nil cloud provider")
		return nil
	}

	provider := buildCloudProvider(opts, do, rl)
	if provider != nil {
		return provider
	}

	klog.Fatalf("Unknown cloud provider: %s", opts.CloudProviderName)
	return nil // This will never happen because the Fatalf will os.Exit
}
