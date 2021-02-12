/*
Copyright 2017 The Kubernetes Authors.

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

package mcm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	kubeletapis "k8s.io/kubernetes/pkg/kubelet/apis"
)

func TestBuildGenericLabels(t *testing.T) {
	var (
		instanceTypeC4Large = "c4.large"
		regionUSEast1       = "us-east-1"
		zoneUSEast1a        = "us-east-1a"
		testHostName        = "test-hostname"
	)

	labels := buildGenericLabels(&nodeTemplate{
		InstanceType: &instanceType{
			InstanceType: instanceTypeC4Large,
			VCPU:         2,
			MemoryMb:     3840,
		},
		Region: regionUSEast1,
		Zone:   zoneUSEast1a,
	}, testHostName)

	assert.Equal(t, cloudprovider.DefaultArch, labels[kubeletapis.LabelArch])
	assert.Equal(t, cloudprovider.DefaultArch, labels[apiv1.LabelArchStable])

	assert.Equal(t, cloudprovider.DefaultOS, labels[kubeletapis.LabelOS])
	assert.Equal(t, cloudprovider.DefaultOS, labels[apiv1.LabelOSStable])

	assert.Equal(t, instanceTypeC4Large, labels[apiv1.LabelInstanceType])
	assert.Equal(t, instanceTypeC4Large, labels[apiv1.LabelInstanceTypeStable])

	assert.Equal(t, regionUSEast1, labels[apiv1.LabelZoneRegion])
	assert.Equal(t, regionUSEast1, labels[apiv1.LabelZoneRegionStable])

	assert.Equal(t, zoneUSEast1a, labels[apiv1.LabelZoneFailureDomain])
	assert.Equal(t, zoneUSEast1a, labels[apiv1.LabelZoneFailureDomainStable])

	assert.Equal(t, testHostName, labels[apiv1.LabelHostname])
}

func TestGenerationOfCorrectZoneValueFromMCLabel(t *testing.T) {
	var (
		resultingZone string
		zoneA         = "zone-a"
		zoneB         = "zone-b"
		randomKey     = "random-key"
		randomValue   = "random-value"
	)

	// Basic test to get zone value
	resultingZone = getZoneValueFromMCLabels(map[string]string{
		apiv1.LabelZoneFailureDomainStable: zoneA,
	})
	assert.Equal(t, resultingZone, zoneA)

	// Prefer LabelZoneFailureDomainStable label over LabelZoneFailureDomain
	resultingZone = getZoneValueFromMCLabels(map[string]string{
		apiv1.LabelZoneFailureDomainStable: zoneA,
		apiv1.LabelZoneFailureDomain:       zoneB,
	})
	assert.Equal(t, resultingZone, zoneA)

	// Fallback to LabelZoneFailureDomain when LabelZoneFailureDomainStable is not found
	resultingZone = getZoneValueFromMCLabels(map[string]string{
		randomKey:                    randomValue,
		apiv1.LabelZoneFailureDomain: zoneB,
	})
	assert.Equal(t, resultingZone, zoneB)

	// When neither of the labels are found
	resultingZone = getZoneValueFromMCLabels(map[string]string{
		randomKey: randomValue,
	})
	assert.Equal(t, resultingZone, "")
}
