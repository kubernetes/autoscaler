// +build !gce,!aws,!azure,!kubemark

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
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/azure"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/gce"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/gke"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/kubemark"
	"k8s.io/autoscaler/cluster-autoscaler/config"
)

// AvailableCloudProviders supported by the cloud provider builder.
var AvailableCloudProviders = []string{
	aws.ProviderName,
	azure.ProviderName,
	gce.ProviderNameGCE,
	gke.ProviderNameGKE,
	kubemark.ProviderName,
}

// DefaultCloudProvider is GCE.
const DefaultCloudProvider = gce.ProviderNameGCE

func buildCloudProvider(opts config.AutoscalingOptions, do cloudprovider.NodeGroupDiscoveryOptions, rl *cloudprovider.ResourceLimiter) cloudprovider.CloudProvider {
	switch opts.CloudProviderName {
	case gce.ProviderNameGCE:
		return gce.BuildGCE(opts, do, rl)
	case gke.ProviderNameGKE:
		return gke.BuildGKE(opts, do, rl)
	case aws.ProviderName:
		return aws.BuildAWS(opts, do, rl)
	case azure.ProviderName:
		return azure.BuildAzure(opts, do, rl)
	case kubemark.ProviderName:
		return kubemark.BuildKubemark(opts, do, rl)
	}

	return nil
}
