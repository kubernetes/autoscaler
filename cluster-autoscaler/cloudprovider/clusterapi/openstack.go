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

package clusterapi

import (
	"encoding/json"
	"fmt"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	kubeletapis "k8s.io/kubernetes/pkg/kubelet/apis"
	"math/rand"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
)

type parsedProviderConfig struct {
	SSHPublicKeys []string `json:"sshPublicKeys"`

	CloudProvider     string               `json:"cloudProvider"`
	CloudProviderSpec runtime.RawExtension `json:"cloudProviderSpec"`

	OperatingSystem     string               `json:"operatingSystem"`
	OperatingSystemSpec runtime.RawExtension `json:"operatingSystemSpec"`
}

type rawConfig struct {
	// TODO can a MD's machine spec be stored in a secret (https://github.com/kubermatic/machine-controller/pull/88)?
	//      If so, we have to check for this case.

	// Auth details
	IdentityEndpoint string `json:"identityEndpoint"`
	Username         string `json:"username"`
	Password         string `json:"password"`
	DomainName       string `json:"domainName"`
	TenantName       string `json:"tenantName"`
	TokenID          string `json:"tokenId"`

	// Machine details
	Image            string            `json:"image"`
	Flavor           string            `json:"flavor"`
	SecurityGroups   []string          `json:"securityGroups"`
	Network          string            `json:"network"`
	Subnet           string            `json:"subnet"`
	FloatingIPPool   string            `json:"floatingIpPool"`
	AvailabilityZone string            `json:"availabilityZone"`
	Region           string            `json:"region"`
	Tags             map[string]string `json:"tags"`
}

// TODO get these dynamically from Openstack
var knownFlavors = map[string]struct{ ram, disk, vcpus int64 }{
	"l1.tiny":    {4096, 25, 1},
	"m1.small":   {8192, 50, 2},
	"m1.tiny":    {4096, 50, 1},
	"l1.xlarge":  {65536, 400, 16},
	"l1.2xlarge": {131072, 800, 32},
	"l1.medium":  {16384, 100, 4},
	"l1.4xlarge": {262144, 1600, 64},
	"m1.large":   {32768, 50, 8},
	"l1.small":   {8192, 50, 2},
	"l1.large":   {32768, 200, 8},
	"m1.micro":   {2048, 50, 1},
	"m1.medium":  {16384, 50, 4},
}

func buildNodeFromOpenstackMachineDeployment(md *v1alpha1.MachineDeployment) (*apiv1.Node, error) {
	providerSpec := md.Spec.Template.Spec.ProviderSpec

	if providerSpec.Value == nil {
		return nil, fmt.Errorf("providerconfig.value is nil")
	}
	pconfig := parsedProviderConfig{}
	err := json.Unmarshal(providerSpec.Value.Raw, &pconfig)
	if err != nil {
		return nil, err
	}
	if "openstack" != pconfig.CloudProvider {
		return nil, fmt.Errorf("non-openstack cloudprovider field (%s) when decoding providerConfig", pconfig.CloudProvider)
	}

	var rawConfig rawConfig
	err = json.Unmarshal(pconfig.CloudProviderSpec.Raw, &rawConfig)
	if err != nil {
		return nil, err
	}

	flavor, ok := knownFlavors[rawConfig.Flavor]
	if !ok {
		return nil, fmt.Errorf("unknown openstack flavor: %s", rawConfig.Flavor)
	}

	node := apiv1.Node{}
	nodeName := fmt.Sprintf("%s-%d", md.Name, rand.Int63())

	node.ObjectMeta = metav1.ObjectMeta{
		Name:     nodeName,
		SelfLink: fmt.Sprintf("/api/v1/nodes/%s", nodeName),
		Labels:   map[string]string{},
	}

	node.Status = apiv1.NodeStatus{
		Capacity: apiv1.ResourceList{},
	}

	// TODO: get a real value.
	node.Status.Capacity[apiv1.ResourcePods] = *resource.NewQuantity(110, resource.DecimalSI)
	node.Status.Capacity[apiv1.ResourceCPU] = *resource.NewQuantity(flavor.vcpus, resource.DecimalSI)
	node.Status.Capacity[apiv1.ResourceMemory] = *resource.NewQuantity(flavor.ram*1024*1024, resource.BinarySI)

	// TODO: use proper allocatable!!
	node.Status.Allocatable = node.Status.Capacity

	// NodeLabels
	//node.Labels = cloudprovider.JoinStringMaps(node.Labels, extractLabelsFromAsg(template.Tags))
	// GenericLabels
	node.Labels = cloudprovider.JoinStringMaps(node.Labels, buildGenericLabels(&rawConfig, nodeName))

	// TODO can a MD specify taints?
	//node.Spec.Taints = extractTaintsFromMD(md)

	node.Status.Conditions = cloudprovider.BuildReadyConditions()
	return &node, nil
}

func buildGenericLabels(rawConfig *rawConfig, nodeName string) map[string]string {
	result := make(map[string]string)
	// TODO: extract it somehow
	result[kubeletapis.LabelArch] = cloudprovider.DefaultArch
	result[kubeletapis.LabelOS] = cloudprovider.DefaultOS

	result[kubeletapis.LabelZoneRegion] = rawConfig.Region
	result[kubeletapis.LabelZoneFailureDomain] = rawConfig.AvailabilityZone
	result[kubeletapis.LabelHostname] = nodeName
	return result
}
