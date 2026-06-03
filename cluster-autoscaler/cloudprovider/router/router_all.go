//go:build !gce && !aws && !azure && !kubemark && !alicloud && !magnum && !digitalocean && !clusterapi && !huaweicloud && !ionoscloud && !linode && !hetzner && !bizflycloud && !brightbox && !equinixmetal && !oci && !vultr && !tencentcloud && !scaleway && !externalgrpc && !civo && !rancher && !volcengine && !baiducloud && !cherry && !cloudstack && !exoscale && !kamatera && !ovhcloud && !kwok && !utho && !coreweave
// +build !gce,!aws,!azure,!kubemark,!alicloud,!magnum,!digitalocean,!clusterapi,!huaweicloud,!ionoscloud,!linode,!hetzner,!bizflycloud,!brightbox,!equinixmetal,!oci,!vultr,!tencentcloud,!scaleway,!externalgrpc,!civo,!rancher,!volcengine,!baiducloud,!cherry,!cloudstack,!exoscale,!kamatera,!ovhcloud,!kwok,!utho,!coreweave

/*
Copyright The Kubernetes Authors.

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

package router

import (
	// Blank import to register a cloudprovider outside main or test package.
	// This is by design.
	_ "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/alicloud"
	_ "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws"
	_ "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/azure"
	_ "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/baiducloud"
	_ "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/bizflycloud"
	_ "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/brightbox"
	_ "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/cherryservers"
	_ "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/civo"
	_ "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/cloudstack"
	_ "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/clusterapi"
	_ "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/coreweave"
	_ "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/digitalocean"
	_ "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/equinixmetal"
	_ "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/exoscale"
	_ "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/externalgrpc"
	_ "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/gce"
	_ "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/hetzner"
	_ "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud"
	_ "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/ionoscloud"
	_ "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/kamatera"
	_ "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/kubemark"
	_ "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/kwok"
	_ "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/linode"
	_ "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/magnum"
	_ "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/instancepools"
	_ "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/ovhcloud"
	_ "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/rancher"
	_ "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/scaleway"
	_ "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/tencentcloud"
	_ "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/utho"
	_ "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/volcengine"
	_ "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/vultr"
)

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/builder"
)

func init() {
	builder.SetDefaultCloudProvider(cloudprovider.GceProviderName)
}
