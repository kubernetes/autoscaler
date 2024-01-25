//go:build !gce && !aws && !azure && !kubemark && !alicloud && !magnum && !digitalocean && !clusterapi && !huaweicloud && !ionoscloud && !linode && !hetzner && !bizflycloud && !brightbox && !equinixmetal && !oci && !vultr && !tencentcloud && !scaleway && !externalgrpc && !civo && !rancher && !volcengine && !baiducloud && !cherry && !cloudstack && !exoscale && !kamatera && !ovhcloud && !kwok
// +build !gce,!aws,!azure,!kubemark,!alicloud,!magnum,!digitalocean,!clusterapi,!huaweicloud,!ionoscloud,!linode,!hetzner,!bizflycloud,!brightbox,!equinixmetal,!oci,!vultr,!tencentcloud,!scaleway,!externalgrpc,!civo,!rancher,!volcengine,!baiducloud,!cherry,!cloudstack,!exoscale,!kamatera,!ovhcloud,!kwok

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
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/bizflycloud"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/brightbox"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/cherryservers"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/civo"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/cloudstack"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/clusterapi"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/digitalocean"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/equinixmetal"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/exoscale"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/externalgrpc"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/gce"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/hetzner"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/ionoscloud"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/kamatera"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/kwok"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/linode"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/magnum"
	oci "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/instancepools"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/ovhcloud"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/rancher"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/scaleway"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/tencentcloud"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/volcengine"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/vultr"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/client-go/informers"
)

// AvailableCloudProviders supported by the cloud provider builder.
var AvailableCloudProviders = []string{
	cloudprovider.AwsProviderName,
	cloudprovider.AzureProviderName,
	cloudprovider.GceProviderName,
	cloudprovider.AlicloudProviderName,
	cloudprovider.CherryServersProviderName,
	cloudprovider.CloudStackProviderName,
	cloudprovider.BaiducloudProviderName,
	cloudprovider.MagnumProviderName,
	cloudprovider.DigitalOceanProviderName,
	cloudprovider.ExoscaleProviderName,
	cloudprovider.ExternalGrpcProviderName,
	cloudprovider.HuaweicloudProviderName,
	cloudprovider.HetznerProviderName,
	cloudprovider.OracleCloudProviderName,
	cloudprovider.OVHcloudProviderName,
	cloudprovider.ClusterAPIProviderName,
	cloudprovider.IonoscloudProviderName,
	cloudprovider.KamateraProviderName,
	cloudprovider.KwokProviderName,
	cloudprovider.LinodeProviderName,
	cloudprovider.BizflyCloudProviderName,
	cloudprovider.BrightboxProviderName,
	cloudprovider.EquinixMetalProviderName,
	cloudprovider.VultrProviderName,
	cloudprovider.TencentcloudProviderName,
	cloudprovider.CivoProviderName,
	cloudprovider.ScalewayProviderName,
	cloudprovider.RancherProviderName,
	cloudprovider.VolcengineProviderName,
}

// DefaultCloudProvider is GCE.
const DefaultCloudProvider = cloudprovider.GceProviderName

func buildCloudProvider(opts config.AutoscalingOptions,
	do cloudprovider.NodeGroupDiscoveryOptions,
	rl *cloudprovider.ResourceLimiter,
	informerFactory informers.SharedInformerFactory) cloudprovider.CloudProvider {
	switch opts.CloudProviderName {
	case cloudprovider.BizflyCloudProviderName:
		return bizflycloud.BuildBizflyCloud(opts, do, rl)
	case cloudprovider.GceProviderName:
		return gce.BuildGCE(opts, do, rl)
	case cloudprovider.AwsProviderName:
		return aws.BuildAWS(opts, do, rl)
	case cloudprovider.AzureProviderName:
		return azure.BuildAzure(opts, do, rl)
	case cloudprovider.AlicloudProviderName:
		return alicloud.BuildAlicloud(opts, do, rl)
	case cloudprovider.CherryServersProviderName:
		return cherryservers.BuildCherry(opts, do, rl)
	case cloudprovider.CloudStackProviderName:
		return cloudstack.BuildCloudStack(opts, do, rl)
	case cloudprovider.BaiducloudProviderName:
		return baiducloud.BuildBaiducloud(opts, do, rl)
	case cloudprovider.BrightboxProviderName:
		return brightbox.BuildBrightbox(opts, do, rl)
	case cloudprovider.DigitalOceanProviderName:
		return digitalocean.BuildDigitalOcean(opts, do, rl)
	case cloudprovider.ExoscaleProviderName:
		return exoscale.BuildExoscale(opts, do, rl)
	case cloudprovider.ExternalGrpcProviderName:
		return externalgrpc.BuildExternalGrpc(opts, do, rl)
	case cloudprovider.MagnumProviderName:
		return magnum.BuildMagnum(opts, do, rl)
	case cloudprovider.HuaweicloudProviderName:
		return huaweicloud.BuildHuaweiCloud(opts, do, rl)
	case cloudprovider.OVHcloudProviderName:
		return ovhcloud.BuildOVHcloud(opts, do, rl)
	case cloudprovider.HetznerProviderName:
		return hetzner.BuildHetzner(opts, do, rl)
	case cloudprovider.PacketProviderName, cloudprovider.EquinixMetalProviderName:
		return equinixmetal.BuildCloudProvider(opts, do, rl)
	case cloudprovider.ClusterAPIProviderName:
		return clusterapi.BuildClusterAPI(opts, do, rl)
	case cloudprovider.IonoscloudProviderName:
		return ionoscloud.BuildIonosCloud(opts, do, rl)
	case cloudprovider.KamateraProviderName:
		return kamatera.BuildKamatera(opts, do, rl)
	case cloudprovider.KwokProviderName:
		return kwok.BuildKwok(opts, do, rl, informerFactory)
	case cloudprovider.LinodeProviderName:
		return linode.BuildLinode(opts, do, rl)
	case cloudprovider.OracleCloudProviderName:
		return oci.BuildOCI(opts, do, rl)
	case cloudprovider.VultrProviderName:
		return vultr.BuildVultr(opts, do, rl)
	case cloudprovider.TencentcloudProviderName:
		return tencentcloud.BuildTencentcloud(opts, do, rl)
	case cloudprovider.CivoProviderName:
		return civo.BuildCivo(opts, do, rl)
	case cloudprovider.ScalewayProviderName:
		return scaleway.BuildScaleway(opts, do, rl)
	case cloudprovider.RancherProviderName:
		return rancher.BuildRancher(opts, do, rl)
	case cloudprovider.VolcengineProviderName:
		return volcengine.BuildVolcengine(opts, do, rl)
	}
	return nil
}
