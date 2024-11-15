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

package gce

import (
	"math"
	"strconv"
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/gce/localssdsize"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	"k8s.io/autoscaler/cluster-autoscaler/utils/units"

	klog "k8s.io/klog/v2"
)

// GcePriceModel implements PriceModel interface for GCE.
type GcePriceModel struct {
	PriceInfo            PriceInfo
	localSSDSizeProvider localssdsize.LocalSSDSizeProvider
}

// NewGcePriceModel gets a new instance of GcePriceModel
func NewGcePriceModel(info PriceInfo, localSSDSizeProvider localssdsize.LocalSSDSizeProvider) *GcePriceModel {
	return &GcePriceModel{
		PriceInfo:            info,
		localSSDSizeProvider: localSSDSizeProvider,
	}
}

const (
	preemptibleLabel              = "cloud.google.com/gke-preemptible"
	spotLabel                     = "cloud.google.com/gke-spot"
	ephemeralStorageLocalSsdLabel = "cloud.google.com/gke-ephemeral-storage-local-ssd"
	bootDiskTypeLabel             = "cloud.google.com/gke-boot-disk"
)

// DefaultBootDiskSizeGB is 100 GB.
const DefaultBootDiskSizeGB = 100

// NodePrice returns a price of running the given node for a given period of time.
// All prices are in USD.
func (model *GcePriceModel) NodePrice(node *apiv1.Node, startTime time.Time, endTime time.Time) (float64, error) {
	price := 0.0
	basePriceFound := false
	machineType := ""
	if node.Labels != nil {
		if _machineType, found := getInstanceTypeFromLabels(node.Labels); found {
			machineType = _machineType
		}
	}
	// Base instance price
	priceMapToUse := model.PriceInfo.InstancePrices()
	if hasPreemptiblePricing(node) {
		priceMapToUse = model.PriceInfo.PreemptibleInstancePrices()
	}
	if basePricePerHour, found := priceMapToUse[machineType]; found {
		price = basePricePerHour * getHours(startTime, endTime)
		basePriceFound = true
	} else {
		klog.Warningf("Pricing information not found for instance type %v; will fallback to default pricing", machineType)
	}
	if !basePriceFound {
		price = model.getBasePrice(node.Status.Capacity, machineType, startTime, endTime)
		price = price * model.getPreemptibleDiscount(node)
	}

	// Ephemeral Storage
	// Local SSD price
	if node.Labels[ephemeralStorageLocalSsdLabel] == "true" || node.Annotations[EphemeralStorageLocalSsdAnnotation] == "true" {
		localSsdCount, _ := strconv.ParseFloat(node.Annotations[LocalSsdCountAnnotation], 64)
		localSsdPrice := model.PriceInfo.LocalSsdPricePerHour()
		if hasPreemptiblePricing(node) {
			localSsdPrice = model.PriceInfo.SpotLocalSsdPricePerHour()
		}
		price += localSsdCount * float64(model.localSSDSizeProvider.SSDSizeInGiB(machineType)) * localSsdPrice * getHours(startTime, endTime)
	}

	// Boot disk price
	bootDiskSize, _ := strconv.ParseInt(node.Annotations[BootDiskSizeAnnotation], 10, 64)
	if bootDiskSize == 0 {
		klog.V(5).Infof("Boot disk size is not found for node %s, using default size %v", node.Name, DefaultBootDiskSizeGB)
		bootDiskSize = DefaultBootDiskSizeGB
	}
	bootDiskType := node.Annotations[BootDiskTypeAnnotation]
	if val, ok := node.Labels[bootDiskTypeLabel]; ok {
		bootDiskType = val
	}
	if bootDiskType == "" {
		klog.V(5).Infof("Boot disk type is not found for node %s, using default type %s", node.Name, DefaultBootDiskType)
		bootDiskType = DefaultBootDiskType
	}
	bootDiskPrice := model.PriceInfo.BootDiskPricePerHour()[bootDiskType]

	price += bootDiskPrice * float64(bootDiskSize) * getHours(startTime, endTime)

	// GPUs
	if gpuRequest, found := node.Status.Capacity[gpu.ResourceNvidiaGPU]; found {
		gpuPrice := model.PriceInfo.BaseGpuPricePerHour()
		if node.Labels != nil {
			priceMapToUse := model.PriceInfo.GpuPrices()
			if hasPreemptiblePricing(node) {
				priceMapToUse = model.PriceInfo.PreemptibleGpuPrices()
			}
			if gpuType, found := node.Labels[GPULabel]; found {
				if _, found := priceMapToUse[gpuType]; found {
					gpuPrice = priceMapToUse[gpuType]
				} else {
					klog.Warningf("Pricing information not found for GPU type %v; will fallback to default pricing", gpuType)
				}
			}
		}
		price += float64(gpuRequest.MilliValue()) / 1000.0 * gpuPrice * getHours(startTime, endTime)
	}

	return price, nil
}

func (model *GcePriceModel) getPreemptibleDiscount(node *apiv1.Node) float64 {
	if !hasPreemptiblePricing(node) {
		return 1.0
	}
	instanceType, found := getInstanceTypeFromLabels(node.Labels)
	if !found {
		return 1.0
	}
	instanceFamily, _ := GetMachineFamily(instanceType)

	discountMap := model.PriceInfo.PredefinedPreemptibleDiscount()
	if IsCustomMachine(instanceType) {
		discountMap = model.PriceInfo.CustomPreemptibleDiscount()
	}

	if _, found := discountMap[instanceFamily]; found {
		return discountMap[instanceFamily]
	}
	return preemptibleDiscount
}

// PodPrice returns a theoretical minimum price of running a pod for a given
// period of time on a perfectly matching machine.
func (model *GcePriceModel) PodPrice(pod *apiv1.Pod, startTime time.Time, endTime time.Time) (float64, error) {
	price := 0.0
	for _, container := range pod.Spec.Containers {
		price += model.getBasePrice(container.Resources.Requests, "", startTime, endTime)
		price += model.getAdditionalPrice(container.Resources.Requests, startTime, endTime)
	}
	return price, nil
}

func (model *GcePriceModel) getBasePrice(resources apiv1.ResourceList, instanceType string, startTime time.Time, endTime time.Time) float64 {
	if len(resources) == 0 {
		return 0
	}
	hours := getHours(startTime, endTime)
	instanceFamily, _ := GetMachineFamily(instanceType)
	isCustom := IsCustomMachine(instanceType)
	price := 0.0

	cpu := resources[apiv1.ResourceCPU]
	cpuPrice := model.PriceInfo.BaseCpuPricePerHour()
	cpuPriceMap := model.PriceInfo.PredefinedCpuPricePerHour()
	if isCustom {
		cpuPriceMap = model.PriceInfo.CustomCpuPricePerHour()
	}
	if _, found := cpuPriceMap[instanceFamily]; found {
		cpuPrice = cpuPriceMap[instanceFamily]
	}
	price += float64(cpu.MilliValue()) / 1000.0 * cpuPrice * hours

	mem := resources[apiv1.ResourceMemory]
	memPrice := model.PriceInfo.BaseMemoryPricePerHourPerGb()
	memPriceMap := model.PriceInfo.PredefinedMemoryPricePerHourPerGb()
	if isCustom {
		memPriceMap = model.PriceInfo.CustomMemoryPricePerHourPerGb()
	}
	if _, found := memPriceMap[instanceFamily]; found {
		memPrice = memPriceMap[instanceFamily]
	}
	price += float64(mem.Value()) / float64(units.GiB) * memPrice * hours

	ephemeralStorage := resources[apiv1.ResourceEphemeralStorage]
	// For simplification using a fixed price for default boot disk.
	ephemeralStoragePrice := model.PriceInfo.BootDiskPricePerHour()[DefaultBootDiskType]
	price += float64(ephemeralStorage.Value()) / float64(units.GiB) * ephemeralStoragePrice * hours

	return price
}

func (model *GcePriceModel) getAdditionalPrice(resources apiv1.ResourceList, startTime time.Time, endTime time.Time) float64 {
	if len(resources) == 0 {
		return 0
	}
	hours := getHours(startTime, endTime)
	price := 0.0
	gpu := resources[gpu.ResourceNvidiaGPU]
	price += float64(gpu.MilliValue()) / 1000.0 * model.PriceInfo.BaseGpuPricePerHour() * hours
	return price
}

func getHours(startTime time.Time, endTime time.Time) float64 {
	minutes := math.Ceil(float64(endTime.Sub(startTime)) / float64(time.Minute))
	hours := minutes / 60.0
	return hours
}

// hasPreemptiblePricing returns whether we should use preemptible pricing for a node, based on labels. Spot VMs have
// dynamic pricing, which is different than the static pricing for Preemptible VMs we use here. However it should be close
// enough in practice and we really only look at prices in comparison with each other. Spot VMs will always be cheaper
// than corresponding non-preemptible VMs. So for the purposes of pricing, Spot VMs are treated the same as
// Preemptible VMs.
func hasPreemptiblePricing(node *apiv1.Node) bool {
	if node.Labels == nil {
		return false
	}
	return node.Labels[preemptibleLabel] == "true" || node.Labels[spotLabel] == "true"
}

func getInstanceTypeFromLabels(labels map[string]string) (string, bool) {
	machineType, found := labels[apiv1.LabelInstanceTypeStable]
	if !found {
		machineType, found = labels[apiv1.LabelInstanceType]
	}
	return machineType, found
}
