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
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	kubeletapis "k8s.io/kubernetes/pkg/kubelet/apis"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
	"testing"
)

func TestBuildNodeFromOpenstackMachineDeploymentMissingProviderConfig(t *testing.T) {
	node, err := buildNodeFromOpenstackMachineDeployment(&v1alpha1.MachineDeployment{})

	assert.Nil(t, node)
	assert.EqualError(t, err, "providerconfig.value is nil")
}

func TestBuildNodeFromOpenstackMachineDeploymentWrongProviderConfig(t *testing.T) {
	providerConfig, _ := json.Marshal(struct {
		CloudProvider string
	}{
		CloudProvider: "invalid",
	})
	node, err := buildNodeFromOpenstackMachineDeployment(&v1alpha1.MachineDeployment{
		Spec: v1alpha1.MachineDeploymentSpec{
			Template: v1alpha1.MachineTemplateSpec{
				Spec: v1alpha1.MachineSpec{
					ProviderSpec: v1alpha1.ProviderSpec{
						Value: &runtime.RawExtension{
							Raw: providerConfig,
						},
					},
				},
			},
		},
	})

	assert.Nil(t, node)
	assert.EqualError(t, err, "non-openstack cloudprovider field (invalid) when decoding providerConfig")
}

func TestBuildNodeFromOpenstackMachineDeploymentUnknownFlavor(t *testing.T) {
	cloudProviderSpec, _ := json.Marshal(rawConfig{
		Flavor: "invalid",
	})
	providerConfig, _ := json.Marshal(parsedProviderConfig{
		CloudProvider: "openstack",
		CloudProviderSpec: runtime.RawExtension{
			Raw: cloudProviderSpec,
		},
	})

	node, err := buildNodeFromOpenstackMachineDeployment(&v1alpha1.MachineDeployment{
		Spec: v1alpha1.MachineDeploymentSpec{
			Template: v1alpha1.MachineTemplateSpec{
				Spec: v1alpha1.MachineSpec{
					ProviderSpec: v1alpha1.ProviderSpec{
						Value: &runtime.RawExtension{
							Raw: providerConfig,
						},
					},
				},
			},
		},
	})

	assert.Nil(t, node)
	assert.EqualError(t, err, "unknown openstack flavor: invalid")
}

func TestBuildGenericLabels(t *testing.T) {
	labels := buildGenericLabels(&rawConfig{
		Region:           "region",
		AvailabilityZone: "zone",
	}, "my-node")

	assert.Equal(t, cloudprovider.DefaultArch, labels[kubeletapis.LabelArch])
	assert.Equal(t, cloudprovider.DefaultOS, labels[kubeletapis.LabelOS])
	assert.Equal(t, "region", labels[kubeletapis.LabelZoneRegion])
	assert.Equal(t, "zone", labels[kubeletapis.LabelZoneFailureDomain])
	assert.Equal(t, "my-node", labels[kubeletapis.LabelHostname])
}
