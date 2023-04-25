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

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2022-03-01/compute"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	cloudvolume "k8s.io/cloud-provider/volume"
	"k8s.io/klog/v2"
)

const (
	azureDiskTopologyKey string = "topology.disk.csi.azure.com/zone"
)

func buildInstanceOS(template compute.VirtualMachineScaleSet) string {
	instanceOS := cloudprovider.DefaultOS
	if template.VirtualMachineProfile != nil && template.VirtualMachineProfile.OsProfile != nil && template.VirtualMachineProfile.OsProfile.WindowsConfiguration != nil {
		instanceOS = "windows"
	}

	return instanceOS
}

func buildGenericLabels(template compute.VirtualMachineScaleSet, nodeName string) map[string]string {
	result := make(map[string]string)

	result[apiv1.LabelArchStable] = cloudprovider.DefaultArch
	result[apiv1.LabelOSStable] = buildInstanceOS(template)

	result[apiv1.LabelInstanceTypeStable] = *template.Sku.Name
	result[apiv1.LabelTopologyRegion] = strings.ToLower(*template.Location)

	if template.Zones != nil && len(*template.Zones) > 0 {
		failureDomains := make([]string, len(*template.Zones))
		for k, v := range *template.Zones {
			failureDomains[k] = strings.ToLower(*template.Location) + "-" + v
		}

		result[apiv1.LabelTopologyZone] = strings.Join(failureDomains[:], cloudvolume.LabelMultiZoneDelimiter)
		result[azureDiskTopologyKey] = strings.Join(failureDomains[:], cloudvolume.LabelMultiZoneDelimiter)
	} else {
		result[apiv1.LabelTopologyZone] = "0"
		result[azureDiskTopologyKey] = ""
	}

	result[apiv1.LabelHostname] = nodeName
	return result
}

func buildNodeFromTemplate(scaleSetName string, template compute.VirtualMachineScaleSet, manager *AzureManager) (*apiv1.Node, error) {
	node := apiv1.Node{}
	nodeName := fmt.Sprintf("%s-asg-%d", scaleSetName, rand.Int63())

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
	if manager.config.EnableDynamicInstanceList {
		var vmssTypeDynamic InstanceType
		klog.V(1).Infof("Fetching instance information for SKU: %s from SKU API", *template.Sku.Name)
		vmssTypeDynamic, dynamicErr = GetVMSSTypeDynamically(template, manager.azureCache)
		if dynamicErr == nil {
			vcpu = vmssTypeDynamic.VCPU
			gpuCount = vmssTypeDynamic.GPU
			memoryMb = vmssTypeDynamic.MemoryMb
		} else {
			klog.Errorf("Dynamically fetching of instance information from SKU api failed with error: %v", dynamicErr)
		}
	}
	if !manager.config.EnableDynamicInstanceList || dynamicErr != nil {
		klog.V(1).Infof("Falling back to static SKU list for SKU: %s", *template.Sku.Name)
		// fall-back on static list of vmss if dynamic workflow fails.
		vmssTypeStatic, staticErr := GetVMSSTypeStatically(template)
		if staticErr == nil {
			vcpu = vmssTypeStatic.VCPU
			gpuCount = vmssTypeStatic.GPU
			memoryMb = vmssTypeStatic.MemoryMb
		} else {
			// return error if neither of the workflows results with vmss data.
			klog.V(1).Infof("Instance type %q not supported, err: %v", *template.Sku.Name, staticErr)
			return nil, staticErr
		}
	}

	node.Status.Capacity[apiv1.ResourcePods] = *resource.NewQuantity(110, resource.DecimalSI)
	node.Status.Capacity[apiv1.ResourceCPU] = *resource.NewQuantity(vcpu, resource.DecimalSI)
	// isNPSeries returns if a SKU is an NP-series SKU
	// SKU API reports GPUs for NP-series but it's actually FPGAs
	if !isNPSeries(*template.Sku.Name) {
		node.Status.Capacity[gpu.ResourceNvidiaGPU] = *resource.NewQuantity(gpuCount, resource.DecimalSI)
	}

	node.Status.Capacity[apiv1.ResourceMemory] = *resource.NewQuantity(memoryMb*1024*1024, resource.DecimalSI)

	resourcesFromTags := extractAllocatableResourcesFromScaleSet(template.Tags)
	for resourceName, val := range resourcesFromTags {
		node.Status.Capacity[apiv1.ResourceName(resourceName)] = *val
	}

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

	// Taints from the Scale Set's Tags
	node.Spec.Taints = extractTaintsFromScaleSet(template.Tags)

	node.Status.Conditions = cloudprovider.BuildReadyConditions()
	return &node, nil
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
