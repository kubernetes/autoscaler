//go:build slicer
// +build slicer

/*
Copyright 2025 The Kubernetes Authors.

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
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/slicer"
	coreoptions "k8s.io/autoscaler/cluster-autoscaler/core/options"
	"k8s.io/client-go/informers"
)

// AvailableCloudProviders supported by the slicer cloud provider builder.
var AvailableCloudProviders = []string{
	cloudprovider.SlicerProviderName,
}

// DefaultCloudProvider for Slicer-only build is Slicer.
const DefaultCloudProvider = cloudprovider.SlicerProviderName

func buildCloudProvider(opts *coreoptions.AutoscalerOptions, do cloudprovider.NodeGroupDiscoveryOptions, rl *cloudprovider.ResourceLimiter, _ informers.SharedInformerFactory) cloudprovider.CloudProvider {
	switch opts.CloudProviderName {
	case cloudprovider.SlicerProviderName:
		return slicer.BuildSlicer(opts, do, rl)
	}

	return nil
}
