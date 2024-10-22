/*
Copyright 2020 The Kubernetes Authors.

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

package azure

import (
	"fmt"
	"strings"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2022-08-01/compute"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/stretchr/testify/assert"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestExtractLabelsFromScaleSet(t *testing.T) {
	expectedNodeLabelKey := "zip"
	expectedNodeLabelValue := "zap"
	extraNodeLabelValue := "buzz"
	blankString := ""

	escapedSlashNodeLabelKey := "spam_egg"
	escapedSlashNodeLabelValue := "foo"
	expectedSlashEscapedNodeLabelKey := "spam/egg"

	escapedUnderscoreNodeLabelKey := "foo~2bar"
	escapedUnderscoreNodeLabelValue := "egg"
	expectedUnderscoreEscapedNodeLabelKey := "foo_bar"

	tags := map[string]*string{
		fmt.Sprintf("%s%s", nodeLabelTagName, expectedNodeLabelKey): &expectedNodeLabelValue,
		"fizz": &extraNodeLabelValue,
		"bip":  &blankString,
		fmt.Sprintf("%s%s", nodeLabelTagName, escapedSlashNodeLabelKey):      &escapedSlashNodeLabelValue,
		fmt.Sprintf("%s%s", nodeLabelTagName, escapedUnderscoreNodeLabelKey): &escapedUnderscoreNodeLabelValue,
	}

	labels := extractLabelsFromScaleSet(tags)
	assert.Len(t, labels, 3)
	assert.Equal(t, expectedNodeLabelValue, labels[expectedNodeLabelKey])
	assert.Equal(t, escapedSlashNodeLabelValue, labels[expectedSlashEscapedNodeLabelKey])
	assert.Equal(t, escapedUnderscoreNodeLabelValue, labels[expectedUnderscoreEscapedNodeLabelKey])
}

func TestExtractTaintsFromScaleSet(t *testing.T) {
	noScheduleTaintValue := "foo:NoSchedule"
	noExecuteTaintValue := "bar:NoExecute"
	preferNoScheduleTaintValue := "fizz:PreferNoSchedule"
	noSplitTaintValue := "some_value"
	blankTaintValue := ""
	regularTagValue := "baz"

	tags := map[string]*string{
		fmt.Sprintf("%s%s", nodeTaintTagName, "dedicated"):                          &noScheduleTaintValue,
		fmt.Sprintf("%s%s", nodeTaintTagName, "group"):                              &noExecuteTaintValue,
		fmt.Sprintf("%s%s", nodeTaintTagName, "app"):                                &preferNoScheduleTaintValue,
		fmt.Sprintf("%s%s", nodeTaintTagName, "k8s.io_testing_underscore_to_slash"): &preferNoScheduleTaintValue,
		"bar": &regularTagValue,
		fmt.Sprintf("%s%s", nodeTaintTagName, "blank"):   &blankTaintValue,
		fmt.Sprintf("%s%s", nodeTaintTagName, "nosplit"): &noSplitTaintValue,
	}

	expectedTaints := []apiv1.Taint{
		{
			Key:    "dedicated",
			Value:  "foo",
			Effect: apiv1.TaintEffectNoSchedule,
		},
		{
			Key:    "group",
			Value:  "bar",
			Effect: apiv1.TaintEffectNoExecute,
		},
		{
			Key:    "app",
			Value:  "fizz",
			Effect: apiv1.TaintEffectPreferNoSchedule,
		},
		{
			Key:    "k8s.io/testing/underscore/to/slash",
			Value:  "fizz",
			Effect: apiv1.TaintEffectPreferNoSchedule,
		},
	}

	taints := extractTaintsFromScaleSet(tags)
	assert.Len(t, taints, 4)
	assert.Equal(t, makeTaintSet(expectedTaints), makeTaintSet(taints))
}

func TestExtractTaintsFromSpecString(t *testing.T) {
	taintsString := []string{
		"dedicated=foo:NoSchedule",
		"group=bar:NoExecute",
		"app=fizz:PreferNoSchedule",
		"k8s.io/testing/underscore/to/slash=fizz:PreferNoSchedule",
		"bar=baz",
		"blank=",
		"nosplit=some_value",
	}

	expectedTaints := []apiv1.Taint{
		{
			Key:    "dedicated",
			Value:  "foo",
			Effect: apiv1.TaintEffectNoSchedule,
		},
		{
			Key:    "group",
			Value:  "bar",
			Effect: apiv1.TaintEffectNoExecute,
		},
		{
			Key:    "app",
			Value:  "fizz",
			Effect: apiv1.TaintEffectPreferNoSchedule,
		},
		{
			Key:    "k8s.io/testing/underscore/to/slash",
			Value:  "fizz",
			Effect: apiv1.TaintEffectPreferNoSchedule,
		},
	}

	taints := extractTaintsFromSpecString(strings.Join(taintsString, ","))
	assert.Len(t, taints, 4)
	assert.Equal(t, makeTaintSet(expectedTaints), makeTaintSet(taints))
}

func TestExtractAllocatableResourcesFromScaleSet(t *testing.T) {
	tags := map[string]*string{
		fmt.Sprintf("%s%s", nodeResourcesTagName, "cpu"):                        to.StringPtr("100m"),
		fmt.Sprintf("%s%s", nodeResourcesTagName, "memory"):                     to.StringPtr("100M"),
		fmt.Sprintf("%s%s", nodeResourcesTagName, "ephemeral-storage"):          to.StringPtr("20G"),
		fmt.Sprintf("%s%s", nodeResourcesTagName, "nvidia.com_Tesla-P100-PCIE"): to.StringPtr("4"),
	}

	labels := extractAllocatableResourcesFromScaleSet(tags)

	assert.Equal(t, resource.NewMilliQuantity(100, resource.DecimalSI).String(), labels["cpu"].String())
	expectedMemory := resource.MustParse("100M")
	assert.Equal(t, (&expectedMemory).String(), labels["memory"].String())
	expectedEphemeralStorage := resource.MustParse("20G")
	assert.Equal(t, (&expectedEphemeralStorage).String(), labels["ephemeral-storage"].String())
	exepectedCustomAllocatable := resource.MustParse("4")
	assert.Equal(t, (&exepectedCustomAllocatable).String(), labels["nvidia.com/Tesla-P100-PCIE"].String())
}

func TestTopologyFromScaleSet(t *testing.T) {
	testNodeName := "test-node"
	testSkuName := "test-sku"
	testVmss := compute.VirtualMachineScaleSet{
		Response: autorest.Response{},
		Sku:      &compute.Sku{Name: &testSkuName},
		Plan:     nil,
		VirtualMachineScaleSetProperties: &compute.VirtualMachineScaleSetProperties{
			VirtualMachineProfile: &compute.VirtualMachineScaleSetVMProfile{OsProfile: nil}},
		Zones:    &[]string{"1", "2", "3"},
		Location: to.StringPtr("westus"),
	}
	expectedZoneValues := []string{"westus-1", "westus-2", "westus-3"}

	labels := buildGenericLabels(testVmss, testNodeName)
	failureDomain, ok := labels[apiv1.LabelZoneFailureDomain]
	assert.True(t, ok)
	topologyZone, ok := labels[apiv1.LabelTopologyZone]
	assert.True(t, ok)
	azureDiskTopology, ok := labels[azureDiskTopologyKey]
	assert.True(t, ok)

	assert.Contains(t, expectedZoneValues, failureDomain)
	assert.Contains(t, expectedZoneValues, topologyZone)
	assert.Contains(t, expectedZoneValues, azureDiskTopology)
}

func TestEmptyTopologyFromScaleSet(t *testing.T) {
	testNodeName := "test-node"
	testSkuName := "test-sku"
	testVmss := compute.VirtualMachineScaleSet{
		Response: autorest.Response{},
		Sku:      &compute.Sku{Name: &testSkuName},
		Plan:     nil,
		VirtualMachineScaleSetProperties: &compute.VirtualMachineScaleSetProperties{
			VirtualMachineProfile: &compute.VirtualMachineScaleSetVMProfile{OsProfile: nil}},
		Location: to.StringPtr("westus"),
	}

	expectedFailureDomain := "0"
	expectedTopologyZone := "0"
	expectedAzureDiskTopology := ""
	labels := buildGenericLabels(testVmss, testNodeName)

	failureDomain, ok := labels[apiv1.LabelZoneFailureDomain]
	assert.True(t, ok)
	assert.Equal(t, expectedFailureDomain, failureDomain)

	topologyZone, ok := labels[apiv1.LabelTopologyZone]
	assert.True(t, ok)
	assert.Equal(t, expectedTopologyZone, topologyZone)

	azureDiskTopology, ok := labels[azureDiskTopologyKey]
	assert.True(t, ok)
	assert.Equal(t, expectedAzureDiskTopology, azureDiskTopology)
}

func makeTaintSet(taints []apiv1.Taint) map[apiv1.Taint]bool {
	set := make(map[apiv1.Taint]bool)
	for _, taint := range taints {
		set[taint] = true
	}
	return set
}
