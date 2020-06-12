// +build !gce,!aws,!azure,!kubemark,!alicloud,!magnum,!digitalocean,!clusterapi,!huaweicloud

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
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/alicloud"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/azure"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/baiducloud"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/clusterapi"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/digitalocean"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/exoscale"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/gce"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud"
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
	cloudprovider.ExoscaleProviderName,
	cloudprovider.HuaweicloudProviderName,
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
	case cloudprovider.ExoscaleProviderName:
		return exoscale.BuildExoscale(opts, do, rl)
	case cloudprovider.MagnumProviderName:
		return magnum.BuildMagnum(opts, do, rl)
	case cloudprovider.HuaweicloudProviderName:
		return huaweicloud.BuildHuaweiCloud(opts, do, rl)
	case packet.ProviderName:
		return packet.BuildPacket(opts, do, rl)
	case clusterapi.ProviderName:
		return clusterapi.BuildClusterAPI(opts, do, rl)
	}
	return nil
}
