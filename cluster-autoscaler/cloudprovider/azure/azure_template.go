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
	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2022-08-01/compute"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	"k8s.io/klog/v2"
)

const (
	azureDiskTopologyKey string = "topology.disk.csi.azure.com/zone"
	// AKSLabelPrefixValue represents the constant prefix for AKSLabelKeyPrefixValue
	AKSLabelPrefixValue = "kubernetes.azure.com"
	// AKSLabelKeyPrefixValue represents prefix for AKS Labels
	AKSLabelKeyPrefixValue = AKSLabelPrefixValue + "/"
)

// NodeTemplate represents a template for a Azure nodepool
type NodeTemplate struct {
	SkuName    string
	InstanceOS string
	Location   string
	Zones      *[]string
	Tags       map[string]*string
	Taints     []string
}

func buildNodeTemplateFromVMSS(vmss compute.VirtualMachineScaleSet) NodeTemplate {
	instanceOS := cloudprovider.DefaultOS
	if vmss.VirtualMachineProfile != nil &&
		vmss.VirtualMachineProfile.OsProfile != nil &&
		vmss.VirtualMachineProfile.OsProfile.WindowsConfiguration != nil {
		instanceOS = "windows"
	}
	return NodeTemplate{
		SkuName:    *vmss.Sku.Name,
		Tags:       vmss.Tags,
		Location:   *vmss.Location,
		Zones:      vmss.Zones,
		InstanceOS: instanceOS,
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
	if !isNPSeries(template.SkuName) {
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
	node.Labels = cloudprovider.JoinStringMaps(node.Labels, extractLabelsFromScaleSet(template.Tags))

	resourcesFromTags := extractAllocatableResourcesFromScaleSet(template.Tags)
	for resourceName, val := range resourcesFromTags {
		node.Status.Capacity[apiv1.ResourceName(resourceName)] = *val
	}

	if len(template.Taints) > 0 {
		node.Spec.Taints = extractTaintsFromTemplate(template.Taints) // Taints from the VMs Node Template
	} else {
		node.Spec.Taints = extractTaintsFromScaleSet(template.Tags) // Taints from the Scale Set's Tags
	}

	node.Status.Conditions = cloudprovider.BuildReadyConditions()
	return &node, nil
}

func buildGenericLabels(template NodeTemplate, nodeName string) map[string]string {
	result := make(map[string]string)

	result[apiv1.LabelArchStable] = cloudprovider.DefaultArch
	result[apiv1.LabelOSStable] = template.InstanceOS

	result[apiv1.LabelInstanceTypeStable] = template.SkuName
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
		result[apiv1.LabelTopologyZone] = randomZone
		result[azureDiskTopologyKey] = randomZone
	} else {
		result[apiv1.LabelTopologyZone] = "0"
		result[azureDiskTopologyKey] = ""
	}

	result[apiv1.LabelHostname] = nodeName
	return result
}

func extractLabelsFromScaleSet(tags map[string]*string) map[string]string {
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

func extractTaintsFromScaleSet(tags map[string]*string) []apiv1.Taint {
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

func extractTaintsFromTemplate(taints []string) []apiv1.Taint {
	result := make([]apiv1.Taint, 0)
	for _, taint := range taints {
		parsedTaint, err := parseTaint(taint)
		if err != nil {
			klog.Warningf("failed to parse taint %q: %v", taint, err)
			continue
		}
		result = append(result, parsedTaint)
	}

	return result
}

// parseTaint parses a taint string, whose format must be either
// '<key>=<value>:<effect>', '<key>:<effect>', or '<key>'.
func parseTaint(taintStr string) (apiv1.Taint, error) {
	var taint apiv1.Taint
	var key string
	var value string
	var effect apiv1.TaintEffect

	parts := strings.Split(taintStr, ":")
	switch len(parts) {
	case 1:
		key = parts[0]
	case 2:
		effect = apiv1.TaintEffect(parts[1])

		partsKV := strings.Split(parts[0], "=")
		if len(partsKV) > 2 {
			return taint, fmt.Errorf("invalid taint spec: %v", taintStr)
		}
		key = partsKV[0]
		if len(partsKV) == 2 {
			value = partsKV[1]
		}
	default:
		return taint, fmt.Errorf("invalid taint spec: %v", taintStr)
	}

	taint.Key = key
	taint.Value = value
	taint.Effect = effect

	return taint, nil
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
