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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v5"
	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2022-08-01/compute"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/stretchr/testify/assert"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestExtractLabelsFromTags(t *testing.T) {
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

	labels := extractLabelsFromTags(tags)
	assert.Len(t, labels, 3)
	assert.Equal(t, expectedNodeLabelValue, labels[expectedNodeLabelKey])
	assert.Equal(t, escapedSlashNodeLabelValue, labels[expectedSlashEscapedNodeLabelKey])
	assert.Equal(t, escapedUnderscoreNodeLabelValue, labels[expectedUnderscoreEscapedNodeLabelKey])
}

func TestExtractTaintsFromTags(t *testing.T) {
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

	taints := extractTaintsFromTags(tags)
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
		{
			Key:    "dedicated", // duplicate key, should be ignored
			Value:  "foo",
			Effect: apiv1.TaintEffectNoSchedule,
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
	template, err := buildNodeTemplateFromVMSS(testVmss, map[string]string{}, "")
	assert.NoError(t, err)
	labels := buildGenericLabels(template, testNodeName)
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
	template, err := buildNodeTemplateFromVMSS(testVmss, map[string]string{}, "")
	assert.NoError(t, err)
	labels := buildGenericLabels(template, testNodeName)

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
func TestBuildNodeTemplateFromVMPool(t *testing.T) {
	agentPoolName := "testpool"
	location := "eastus"
	skuName := "Standard_DS2_v2"
	labelKey := "foo"
	labelVal := "bar"
	taintStr := "dedicated=foo:NoSchedule,boo=fizz:PreferNoSchedule,group=bar:NoExecute"

	osType := armcontainerservice.OSTypeLinux
	osDiskType := armcontainerservice.OSDiskTypeEphemeral
	zone1 := "1"
	zone2 := "2"

	vmpool := armcontainerservice.AgentPool{
		Name: to.StringPtr(agentPoolName),
		Properties: &armcontainerservice.ManagedClusterAgentPoolProfileProperties{
			NodeLabels: map[string]*string{
				"existing":   to.StringPtr("label"),
				"department": to.StringPtr("engineering"),
			},
			NodeTaints:        []*string{to.StringPtr("group=bar:NoExecute")},
			OSType:            &osType,
			OSDiskType:        &osDiskType,
			AvailabilityZones: []*string{&zone1, &zone2},
		},
	}

	labelsFromSpec := map[string]string{labelKey: labelVal}
	taintsFromSpec := taintStr

	template, err := buildNodeTemplateFromVMPool(vmpool, location, skuName, labelsFromSpec, taintsFromSpec)
	assert.NoError(t, err)
	assert.Equal(t, skuName, template.SkuName)
	assert.Equal(t, location, template.Location)
	assert.ElementsMatch(t, []string{zone1, zone2}, template.Zones)
	assert.Equal(t, "linux", template.InstanceOS)
	assert.NotNil(t, template.VMPoolNodeTemplate)
	assert.Equal(t, agentPoolName, template.VMPoolNodeTemplate.AgentPoolName)
	assert.Equal(t, &osDiskType, template.VMPoolNodeTemplate.OSDiskType)
	// Labels: should include both from NodeLabels and labelsFromSpec
	assert.Contains(t, template.VMPoolNodeTemplate.Labels, "existing")
	assert.Equal(t, "label", *template.VMPoolNodeTemplate.Labels["existing"])
	assert.Contains(t, template.VMPoolNodeTemplate.Labels, "department")
	assert.Equal(t, "engineering", *template.VMPoolNodeTemplate.Labels["department"])
	assert.Contains(t, template.VMPoolNodeTemplate.Labels, labelKey)
	assert.Equal(t, labelVal, *template.VMPoolNodeTemplate.Labels[labelKey])
	// Taints: should include both from NodeTaints and taintsFromSpec
	taintSet := makeTaintSet(template.VMPoolNodeTemplate.Taints)
	expectedTaints := []apiv1.Taint{
		{Key: "group", Value: "bar", Effect: apiv1.TaintEffectNoExecute},
		{Key: "dedicated", Value: "foo", Effect: apiv1.TaintEffectNoSchedule},
		{Key: "boo", Value: "fizz", Effect: apiv1.TaintEffectPreferNoSchedule},
	}
	assert.Equal(t, makeTaintSet(expectedTaints), taintSet)
}

func makeTaintSet(taints []apiv1.Taint) map[apiv1.Taint]bool {
	set := make(map[apiv1.Taint]bool)
	for _, taint := range taints {
		set[taint] = true
	}
	return set
}

func TestBuildNodeFromTemplateWithLabelPrediction(t *testing.T) {
	poolName := "testpool"
	testSkuName := "Standard_DS2_v2"
	testNodeName := "test-node"

	vmss := compute.VirtualMachineScaleSet{
		Response: autorest.Response{},
		Sku:      &compute.Sku{Name: &testSkuName},
		Plan:     nil,
		VirtualMachineScaleSetProperties: &compute.VirtualMachineScaleSetProperties{
			VirtualMachineProfile: &compute.VirtualMachineScaleSetVMProfile{
				StorageProfile: &compute.VirtualMachineScaleSetStorageProfile{
					OsDisk: &compute.VirtualMachineScaleSetOSDisk{
						DiffDiskSettings: nil, // This makes it managed
						ManagedDisk: &compute.VirtualMachineScaleSetManagedDiskParameters{
							StorageAccountType: compute.StorageAccountTypesPremiumLRS,
						},
					},
				},
			},
		},
		Tags: map[string]*string{
			"poolName": &poolName,
		},
		Zones:    &[]string{"1", "2"},
		Location: to.StringPtr("westus"),
	}

	template, err := buildNodeTemplateFromVMSS(vmss, map[string]string{}, "")
	assert.NoError(t, err)

	manager := &AzureManager{}
	node, err := buildNodeFromTemplate(testNodeName, template, manager, false, true)
	assert.NoError(t, err)
	assert.NotNil(t, node)

	// Verify label prediction labels are added
	assert.Equal(t, poolName, node.Labels["agentpool"])
	assert.Equal(t, poolName, node.Labels["kubernetes.azure.com/agentpool"])
	assert.Equal(t, "managed", node.Labels["storageprofile"])
	assert.Equal(t, "managed", node.Labels["kubernetes.azure.com/storageprofile"])
}
