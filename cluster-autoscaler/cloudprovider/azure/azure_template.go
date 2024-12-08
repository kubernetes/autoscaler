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
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v5"
	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2022-03-01/compute"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	"k8s.io/klog/v2"
	kubeletapis "k8s.io/kubelet/pkg/apis"
)

const (
	// AKSLabelPrefixValue represents the constant prefix for AKSLabelKeyPrefixValue
	AKSLabelPrefixValue = "kubernetes.azure.com"
	// AKSLabelKeyPrefixValue represents prefix for AKS Labels
	AKSLabelKeyPrefixValue = AKSLabelPrefixValue + "/"

	azureDiskTopologyKey = "topology.disk.csi.azure.com/zone"
	// For NP-series SKU, the xilinx device plugin uses that resource name
	// https://github.com/Xilinx/FPGA_as_a_Service/tree/master/k8s-fpga-device-plugin
	xilinxFpgaResourceName = "xilinx.com/fpga-xilinx_u250_gen3x16_xdma_shell_2_1-0"

	// legacyPoolNameTag is the legacy tag that AKS adds to the VMSS with its value
	// being the agentpool name
	legacyPoolNameTag = "poolName"
	// poolNameTag is the new tag that replaces the above one
	// Newly created pools and clusters will have this one on the VMSS
	// instead of the legacy one. We'll have to live with both tags for a while.
	poolNameTag = "aks-managed-poolName"

	// This is the legacy label is added by agentbaker, agentpool={poolName} and we want to predict that
	// a node added to this agentpool will have this as a node label. The value is fetched
	// from the VMSS tag with key poolNameTag/legacyPoolNameTag
	legacyAgentPoolNodeLabelKey = "agentpool"
	// New label that replaces the above
	agentPoolNodeLabelKey = AKSLabelKeyPrefixValue + "agentpool"

	// Storage profile node labels
	legacyStorageProfileNodeLabelKey = "storageprofile"
	storageProfileNodeLabelKey       = AKSLabelKeyPrefixValue + "storageprofile"

	// Storage tier node labels
	legacyStorageTierNodeLabelKey = "storagetier"
	storageTierNodeLabelKey       = AKSLabelKeyPrefixValue + "storagetier"

	// Fips node label
	fipsNodeLabelKey = AKSLabelKeyPrefixValue + "fips_enabled"

	// OS Sku node Label
	osSkuLabelKey = AKSLabelKeyPrefixValue + "os-sku"

	// Security node label
	securityTypeLabelKey = AKSLabelKeyPrefixValue + "security-type"

	customCATrustEnabledLabelKey = AKSLabelKeyPrefixValue + "custom-ca-trust-enabled"
	kataMshvVMIsolationLabelKey  = AKSLabelKeyPrefixValue + "kata-mshv-vm-isolation"

	// Cluster node label
	clusterLabelKey = AKSLabelKeyPrefixValue + "cluster"
)

// NodeTemplate represents a template for a Azure nodepool
type NodeTemplate struct {
	SkuName    string
	InstanceOS string
	Location   string
	Zones      *[]string
	Tags       map[string]*string
	// for vms pool only
	Taints []string
	// for vmss pool only
	InputLabels map[string]string
	InputTaints string
	OSDisk      *compute.VirtualMachineScaleSetOSDisk
}

func buildNodeTemplateFromVMSS(vmss compute.VirtualMachineScaleSet, inputLabels map[string]string, inputTaints string) NodeTemplate {
	instanceOS := cloudprovider.DefaultOS
	if vmss.VirtualMachineProfile != nil &&
		vmss.VirtualMachineProfile.OsProfile != nil &&
		vmss.VirtualMachineProfile.OsProfile.WindowsConfiguration != nil {
		instanceOS = "windows"
	}

	var osDisk *compute.VirtualMachineScaleSetOSDisk
	if vmss.VirtualMachineProfile != nil &&
		vmss.VirtualMachineProfile.StorageProfile != nil &&
		vmss.VirtualMachineProfile.StorageProfile.OsDisk != nil {
		osDisk = vmss.VirtualMachineProfile.StorageProfile.OsDisk
	}
	return NodeTemplate{
		SkuName:     *vmss.Sku.Name,
		Tags:        vmss.Tags,
		Location:    *vmss.Location,
		Zones:       vmss.Zones,
		InstanceOS:  instanceOS,
		InputLabels: inputLabels,
		InputTaints: inputTaints,
		OSDisk:      osDisk,
	}
}

func buildNodeTemplateFromVMsPool(vmsPool armcontainerservice.AgentPool, location string) NodeTemplate {
	var skuName string
	if vmsPool.Properties != nil &&
		vmsPool.Properties.VirtualMachinesProfile != nil &&
		vmsPool.Properties.VirtualMachinesProfile.Scale != nil {
		if len(vmsPool.Properties.VirtualMachinesProfile.Scale.Manual) > 0 &&
			len(vmsPool.Properties.VirtualMachinesProfile.Scale.Manual[0].Sizes) > 0 &&
			vmsPool.Properties.VirtualMachinesProfile.Scale.Manual[0].Sizes[0] != nil {
			skuName = *vmsPool.Properties.VirtualMachinesProfile.Scale.Manual[0].Sizes[0]
		}
		if len(vmsPool.Properties.VirtualMachinesProfile.Scale.Autoscale) > 0 &&
			len(vmsPool.Properties.VirtualMachinesProfile.Scale.Autoscale[0].Sizes) > 0 &&
			vmsPool.Properties.VirtualMachinesProfile.Scale.Autoscale[0].Sizes[0] != nil {
			skuName = *vmsPool.Properties.VirtualMachinesProfile.Scale.Autoscale[0].Sizes[0]
		}
	}

	var labels map[string]*string
	if vmsPool.Properties != nil && vmsPool.Properties.NodeLabels != nil {
		labels = vmsPool.Properties.NodeLabels
	}

	var taints []string
	if vmsPool.Properties != nil && vmsPool.Properties.NodeTaints != nil {
		for _, taint := range vmsPool.Properties.NodeTaints {
			if taint != nil {
				taints = append(taints, *taint)
			}
		}
	}

	var zones []string
	if vmsPool.Properties != nil && vmsPool.Properties.AvailabilityZones != nil {
		for _, zone := range vmsPool.Properties.AvailabilityZones {
			if zone != nil {
				zones = append(zones, *zone)
			}
		}
	}

	var instanceOS string
	if vmsPool.Properties != nil && vmsPool.Properties.OSType != nil {
		instanceOS = strings.ToLower(string(*vmsPool.Properties.OSType))
	}

	return NodeTemplate{
		SkuName:    skuName,
		Tags:       labels,
		Taints:     taints,
		Zones:      &zones,
		InstanceOS: instanceOS,
		Location:   location,
	}
}

func buildNodeFromTemplate(nodeGroupName string, template NodeTemplate, manager *AzureManager, enableDynamicInstanceList bool) (*apiv1.Node, error) {
	node := apiv1.Node{}
	nodeName := fmt.Sprintf("%s-asg-%d", nodeGroupName, rand.Int63())

	node.ObjectMeta = metav1.ObjectMeta{
		Name:     nodeName,
		SelfLink: fmt.Sprintf("/api/v1/nodes/%s", nodeName),
		Labels:   map[string]string{},
	}

	node.Status = apiv1.NodeStatus{
		Capacity: apiv1.ResourceList{},
	}

	var vcpu, gpuCount, memoryMb int64

	// Fetching SKU information from SKU API if enableDynamicInstanceList is true.
	var dynamicErr error
	if enableDynamicInstanceList {
		var vmssTypeDynamic InstanceType
		klog.V(1).Infof("Fetching instance information for SKU: %s from SKU API", template.SkuName)
		vmssTypeDynamic, dynamicErr = GetVMSSTypeDynamically(template, manager.azureCache)
		if dynamicErr == nil {
			vcpu = vmssTypeDynamic.VCPU
			gpuCount = vmssTypeDynamic.GPU
			memoryMb = vmssTypeDynamic.MemoryMb
		} else {
			klog.Errorf("Dynamically fetching of instance information from SKU api failed with error: %v", dynamicErr)
		}
	}
	if !enableDynamicInstanceList || dynamicErr != nil {
		klog.V(1).Infof("Falling back to static SKU list for SKU: %s", template.SkuName)
		// fall-back on static list of vmss if dynamic workflow fails.
		vmssTypeStatic, staticErr := GetVMSSTypeStatically(template)
		if staticErr == nil {
			vcpu = vmssTypeStatic.VCPU
			gpuCount = vmssTypeStatic.GPU
			memoryMb = vmssTypeStatic.MemoryMb
		} else {
			// return error if neither of the workflows results with vmss data.
			klog.V(1).Infof("Instance type %q not supported, err: %v", template.SkuName, staticErr)
			return nil, staticErr
		}
	}

	node.Status.Capacity[apiv1.ResourcePods] = *resource.NewQuantity(110, resource.DecimalSI)
	node.Status.Capacity[apiv1.ResourceCPU] = *resource.NewQuantity(vcpu, resource.DecimalSI)
	// isNPSeries returns if a SKU is an NP-series SKU
	// SKU API reports GPUs for NP-series but it's actually FPGAs
	if isNPSeries(template.SkuName) {
		node.Status.Capacity[xilinxFpgaResourceName] = *resource.NewQuantity(gpuCount, resource.DecimalSI)
	} else {
		node.Status.Capacity[gpu.ResourceNvidiaGPU] = *resource.NewQuantity(gpuCount, resource.DecimalSI)
	}

	node.Status.Capacity[apiv1.ResourceMemory] = *resource.NewQuantity(memoryMb*1024*1024, resource.DecimalSI)

	// TODO: set real allocatable.
	node.Status.Allocatable = node.Status.Capacity

	// NodeLabels
	if template.Tags != nil {
		for k, v := range template.Tags {
			if v != nil {
				node.Labels[k] = *v
			} else {
				node.Labels[k] = ""
			}

		}
	}

	// GenericLabels
	node.Labels = cloudprovider.JoinStringMaps(node.Labels, buildGenericLabels(template, nodeName))

	// Labels from the Scale Set's Tags
	labels := make(map[string]string)

	// Prefer the explicit labels in spec coming from RP over the VMSS template
	if len(template.InputLabels) > 0 {
		labels = template.InputLabels
	} else {
		labels = extractLabelsFromTags(template.Tags)
	}

	// Add the agentpool label, its value should come from the VMSS poolName tag
	// NOTE: The plan is for agentpool label to be deprecated in favor of the aks-prefixed one
	// We will have to live with both labels for a while
	if node.Labels[legacyPoolNameTag] != "" {
		labels[legacyAgentPoolNodeLabelKey] = node.Labels[legacyPoolNameTag]
		labels[agentPoolNodeLabelKey] = node.Labels[legacyPoolNameTag]
	}
	if node.Labels[poolNameTag] != "" {
		labels[legacyAgentPoolNodeLabelKey] = node.Labels[poolNameTag]
		labels[agentPoolNodeLabelKey] = node.Labels[poolNameTag]
	}

	// Add the storage profile and storage tier labels
	if template.OSDisk != nil {
		// ephemeral
		if template.OSDisk.DiffDiskSettings != nil && template.OSDisk.DiffDiskSettings.Option == compute.Local {
			labels[legacyStorageProfileNodeLabelKey] = "ephemeral"
			labels[storageProfileNodeLabelKey] = "ephemeral"
		} else {
			labels[legacyStorageProfileNodeLabelKey] = "managed"
			labels[storageProfileNodeLabelKey] = "managed"
		}
		if template.OSDisk.ManagedDisk != nil {
			labels[legacyStorageTierNodeLabelKey] = string(template.OSDisk.ManagedDisk.StorageAccountType)
			labels[storageTierNodeLabelKey] = string(template.OSDisk.ManagedDisk.StorageAccountType)
		}
		// Add ephemeral-storage value
		if template.OSDisk.DiskSizeGB != nil {
			node.Status.Capacity[apiv1.ResourceEphemeralStorage] = *resource.NewQuantity(int64(int(*template.OSDisk.DiskSizeGB)*1024*1024*1024), resource.DecimalSI)
			klog.V(4).Infof("OS Disk Size from template is: %d", *template.OSDisk.DiskSizeGB)
			klog.V(4).Infof("Setting ephemeral storage to: %v", node.Status.Capacity[apiv1.ResourceEphemeralStorage])
		}
	}

	// If we are on GPU-enabled SKUs, append the accelerator
	// label so that CA makes better decision when scaling from zero for GPU pools
	if isNvidiaEnabledSKU(template.SkuName) {
		labels[GPULabel] = "nvidia"
		labels[legacyGPULabel] = "nvidia"
	}

	// Extract allocatables from tags
	resourcesFromTags := extractAllocatableResourcesFromScaleSet(template.Tags)
	for resourceName, val := range resourcesFromTags {
		node.Status.Capacity[apiv1.ResourceName(resourceName)] = *val
	}

	node.Labels = cloudprovider.JoinStringMaps(node.Labels, labels)
	klog.V(4).Infof("Setting node %s labels to: %s", nodeName, node.Labels)

	var taints []apiv1.Taint
	// Prefer the explicit taints in spec over the tags from vmss or vm
	if template.InputTaints != "" {
		taints = extractTaintsFromSpecString(template.InputTaints)
	} else {
		taints = extractTaintsFromTags(template.Tags)
	}

	// Taints from the Scale Set's Tags
	node.Spec.Taints = taints
	klog.V(4).Infof("Setting node %s taints to: %s", nodeName, node.Spec.Taints)

	node.Status.Conditions = cloudprovider.BuildReadyConditions()
	return &node, nil
}

func buildGenericLabels(template NodeTemplate, nodeName string) map[string]string {
	result := make(map[string]string)

	result[kubeletapis.LabelArch] = cloudprovider.DefaultArch
	result[apiv1.LabelArchStable] = cloudprovider.DefaultArch

	result[kubeletapis.LabelOS] = template.InstanceOS
	result[apiv1.LabelOSStable] = template.InstanceOS

	result[apiv1.LabelInstanceType] = template.SkuName
	result[apiv1.LabelInstanceTypeStable] = template.SkuName
	result[apiv1.LabelZoneRegion] = strings.ToLower(template.Location)
	result[apiv1.LabelTopologyRegion] = strings.ToLower(template.Location)

	if template.Zones != nil && len(*template.Zones) > 0 {
		failureDomains := make([]string, len(*template.Zones))
		for k, v := range *template.Zones {
			failureDomains[k] = strings.ToLower(template.Location) + "-" + v
		}
		//Picks random zones for Multi-zone nodepool when scaling from zero.
		//This random zone will not be the same as the zone of the VMSS that is being created, the purpose of creating
		//the node template with random zone is to initiate scaling from zero on the multi-zone nodepool.
		//Note that the if the customer is to have some pod affinity picking exact zone, this logic won't work.
		//For now, discourage the customers from using podAffinity to pick the availability zones.
		randomZone := failureDomains[rand.Intn(len(failureDomains))]
		result[apiv1.LabelZoneFailureDomain] = randomZone
		result[apiv1.LabelTopologyZone] = randomZone
		result[azureDiskTopologyKey] = randomZone
	} else {
		result[apiv1.LabelZoneFailureDomain] = "0"
		result[apiv1.LabelTopologyZone] = "0"
		result[azureDiskTopologyKey] = ""
	}

	result[apiv1.LabelHostname] = nodeName
	return result
}

func extractLabelsFromTags(tags map[string]*string) map[string]string {
	result := make(map[string]string)

	for tagName, tagValue := range tags {
		splits := strings.Split(tagName, nodeLabelTagName)
		if len(splits) > 1 {
			label := strings.Replace(splits[1], "_", "/", -1)
			label = strings.Replace(label, "~2", "_", -1)
			if label != "" {
				result[label] = *tagValue
			}
		}
	}

	return result
}

func extractTaintsFromTags(tags map[string]*string) []apiv1.Taint {
	taints := make([]apiv1.Taint, 0)

	for tagName, tagValue := range tags {
		// The tag value must be in the format <tag>:NoSchedule
		r, _ := regexp.Compile("(.*):(?:NoSchedule|NoExecute|PreferNoSchedule)")

		if r.MatchString(*tagValue) {
			splits := strings.Split(tagName, nodeTaintTagName)
			if len(splits) > 1 {
				values := strings.SplitN(*tagValue, ":", 2)
				if len(values) > 1 {
					taintKey := strings.Replace(splits[1], "_", "/", -1)
					taintKey = strings.Replace(taintKey, "~2", "_", -1)
					taints = append(taints, apiv1.Taint{
						Key:    taintKey,
						Value:  values[0],
						Effect: apiv1.TaintEffect(values[1]),
					})
				}
			}
		}
	}

	return taints
}

// Example of a valid taints string, is the same argument to kubelet's `--register-with-taints`
// "dedicated=foo:NoSchedule,group=bar:NoExecute,app=fizz:PreferNoSchedule"
func extractTaintsFromSpecString(taintsString string) []apiv1.Taint {
	taints := make([]apiv1.Taint, 0)
	// First split the taints at the separator
	splits := strings.Split(taintsString, ",")
	for _, split := range splits {
		taintSplit := strings.Split(split, "=")
		if len(taintSplit) != 2 {
			continue
		}

		taintKey := taintSplit[0]
		taintValue := taintSplit[1]

		r, _ := regexp.Compile("(.*):(?:NoSchedule|NoExecute|PreferNoSchedule)")
		if !r.MatchString(taintValue) {
			continue
		}

		values := strings.SplitN(taintValue, ":", 2)
		taints = append(taints, apiv1.Taint{
			Key:    taintKey,
			Value:  values[0],
			Effect: apiv1.TaintEffect(values[1]),
		})
	}

	return taints
}

func extractAutoscalingOptionsFromScaleSetTags(tags map[string]*string) map[string]string {
	options := make(map[string]string)
	for tagName, tagValue := range tags {
		if !strings.HasPrefix(tagName, nodeOptionsTagName) {
			continue
		}
		resourceName := strings.Split(tagName, nodeOptionsTagName)
		if len(resourceName) < 2 || resourceName[1] == "" || tagValue == nil {
			continue
		}
		options[resourceName[1]] = strings.ToLower(*tagValue)
	}
	return options
}

func getFloat64Option(options map[string]string, vmssName, name string) (float64, bool) {
	raw, ok := options[strings.ToLower(name)]
	if !ok {
		return 0, false
	}

	option, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		klog.Warningf("failed to convert VMSS %q tag %s_%s value %q to float: %v",
			vmssName, nodeOptionsTagName, name, raw, err)
		return 0, false
	}

	return option, true
}

func getDurationOption(options map[string]string, vmssName, name string) (time.Duration, bool) {
	raw, ok := options[strings.ToLower(name)]
	if !ok {
		return 0, false
	}

	option, err := time.ParseDuration(raw)
	if err != nil {
		klog.Warningf("failed to convert VMSS %q tag %s_%s value %q to duration: %v",
			vmssName, nodeOptionsTagName, name, raw, err)
		return 0, false
	}

	return option, true
}

func extractAllocatableResourcesFromScaleSet(tags map[string]*string) map[string]*resource.Quantity {
	resources := make(map[string]*resource.Quantity)

	for tagName, tagValue := range tags {
		resourceName := strings.Split(tagName, nodeResourcesTagName)
		if len(resourceName) < 2 || resourceName[1] == "" {
			continue
		}

		normalizedResourceName := strings.Replace(resourceName[1], "_", "/", -1)
		normalizedResourceName = strings.Replace(normalizedResourceName, "~2", "/", -1)
		quantity, err := resource.ParseQuantity(*tagValue)
		if err != nil {
			continue
		}
		resources[normalizedResourceName] = &quantity
	}

	return resources
}

// isNPSeries returns if a SKU is an NP-series SKU
// SKU API reports GPUs for NP-series but it's actually FPGAs
func isNPSeries(name string) bool {
	return strings.HasPrefix(strings.ToLower(name), "standard_np")
}
