//go:build bizflycloud
// +build bizflycloud

/*
Copyright 2021 The Kubernetes Authors.

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
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/bizflycloud"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/client-go/informers"
)

// AvailableCloudProviders supported by the Bizflycloud provider builder.
var AvailableCloudProviders = []string{
	cloudprovider.BizflyCloudProviderName,
}

// DefaultCloudProvider for Bizflycloud-only build is Bizflycloud.
const DefaultCloudProvider = cloudprovider.BizflyCloudProviderName

func buildCloudProvider(opts config.AutoscalingOptions, do cloudprovider.NodeGroupDiscoveryOptions, rl *cloudprovider.ResourceLimiter, _ informers.SharedInformerFactory) cloudprovider.CloudProvider {
	switch opts.CloudProviderName {
	case cloudprovider.BizflyCloudProviderName:
		return bizflycloud.BuildBizflyCloud(opts, do, rl)
	}

	return nil
}
