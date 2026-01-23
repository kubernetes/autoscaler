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

package gpu

import (
	"fmt"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	podutils "k8s.io/autoscaler/cluster-autoscaler/utils/pod"
	"k8s.io/klog/v2"
)

const (
	// ResourceIntelGaudi is the name of the Intel Gaudi resource.
	ResourceIntelGaudi = "habana.ai/gaudi"
	// ResourceIntelGPU is the name of the Intel GPU resource (Xe driver).
	ResourceIntelGPU = "gpu.intel.com/xe"
	// ResourceAMDGPU is the name of the AMD GPU resource.
	ResourceAMDGPU = "amd.com/gpu"
	// ResourceNvidiaGPU is the name of the Nvidia GPU resource.
	ResourceNvidiaGPU = "nvidia.com/gpu"
	// ResourceDirectX is the name of the DirectX resource on windows.
	ResourceDirectX = "microsoft.com/directx"
	// DefaultGPUType is the type of GPU used in NAP if the user
	// don't specify what type of GPU his pod wants.
	DefaultGPUType = "nvidia-tesla-k80"
)

// GPUVendorResourceNames centralized list of all known GPU vendor extended resource names.
// Extend this slice if new vendor resource names are added.
var GPUVendorResourceNames = []apiv1.ResourceName{
	ResourceNvidiaGPU,
	ResourceIntelGaudi,
	ResourceIntelGPU,
	ResourceAMDGPU,
	ResourceDirectX,
}

const (
	// MetricsGenericGPU - for when there is no information about GPU type
	MetricsGenericGPU = "generic"
	// MetricsMissingGPU - for when there's a label, but GPU didn't appear
	MetricsMissingGPU = "missing-gpu"
	// MetricsUnexpectedLabelGPU - for when there's a label, but no GPU at all
	MetricsUnexpectedLabelGPU = "unexpected-label"
	// MetricsUnknownGPU - for when GPU type is unknown
	MetricsUnknownGPU = "not-listed"
	// MetricsErrorGPU - for when there was an error obtaining GPU type
	MetricsErrorGPU = "error"
	// MetricsNoGPU - for when there is no GPU and no label all
	MetricsNoGPU = ""
)

// GetGpuInfoForMetrics returns the name of the custom resource and the GPU used on the node or empty string if there's no GPU
// if the GPU type is unknown, "generic" is returned
// NOTE: current implementation is GKE/GCE-specific
func GetGpuInfoForMetrics(gpuConfig *cloudprovider.GpuConfig, availableGPUTypes map[string]struct{}, node *apiv1.Node, nodeGroup cloudprovider.NodeGroup) (gpuResource string, gpuType string) {
	// There is no sign of GPU
	if gpuConfig == nil {
		return "", MetricsNoGPU
	}

	resourceName := gpuConfig.ExtendedResourceName
	capacity, capacityFound := node.Status.Capacity[resourceName]
	// There is no label value, fallback to generic solution
	if gpuConfig.Type == "" && capacityFound && !capacity.IsZero() {
		return resourceName.String(), MetricsGenericGPU
	}

	// GPU is exposed using DRA, capacity won't be present
	if gpuConfig.ExposedViaDra() {
		draResourceName := fmt.Sprintf("dra_%s", gpuConfig.DraDriverName)
		return draResourceName, validateGpuType(availableGPUTypes, gpuConfig.Type)
	}

	// GKE-specific label & capacity are present - consistent state
	if capacityFound {
		return resourceName.String(), validateGpuType(availableGPUTypes, gpuConfig.Type)
	}

	// GKE-specific label present but no capacity (yet?) - check the node template
	if nodeGroup != nil {
		template, err := nodeGroup.TemplateNodeInfo()
		if err != nil {
			klog.Warningf("Failed to build template for getting GPU metrics for node %v: %v", node.Name, err)
			return resourceName.String(), MetricsErrorGPU
		}

		if _, found := template.Node().Status.Capacity[resourceName]; found {
			return resourceName.String(), MetricsMissingGPU
		}

		// if template does not define GPUs we assume node will not have any even if it has gpu label
		klog.Warningf("Template does not define GPUs even though node from its node group does; node=%v", node.Name)
		return resourceName.String(), MetricsUnexpectedLabelGPU
	}

	return resourceName.String(), MetricsUnexpectedLabelGPU
}

func validateGpuType(availableGPUTypes map[string]struct{}, gpu string) string {
	if _, found := availableGPUTypes[gpu]; found {
		return gpu
	}
	return MetricsUnknownGPU
}

// NodeHasGpu returns true if a given node has GPU hardware.
// The result will be true if there is hardware capability. It doesn't matter
// if the drivers are installed and GPU is ready to use.
func NodeHasGpu(GPULabel string, node *apiv1.Node) bool {
	_, hasGpuLabel := node.Labels[GPULabel]
	if hasGpuLabel {
		return true
	}
	// Check for extended resources as well
	_, hasGpuAllocatable := NodeHasGpuAllocatable(node)
	return hasGpuAllocatable
}

// NodeHasGpuAllocatable returns the GPU allocatable value and whether the node has GPU allocatable resources.
// It checks all known GPU vendor resource names and returns the first non-zero allocatable GPU value found.
func NodeHasGpuAllocatable(node *apiv1.Node) (gpuAllocatableValue int64, hasGpuAllocatable bool) {
	for _, gpuVendorResourceName := range GPUVendorResourceNames {
		gpuAllocatable, found := node.Status.Allocatable[gpuVendorResourceName]
		if found && !gpuAllocatable.IsZero() {
			return gpuAllocatable.Value(), true
		}
	}
	return 0, false
}

// PodRequestsGpu returns true if a given pod has GPU request.
func PodRequestsGpu(pod *apiv1.Pod) bool {
	podRequests := podutils.PodRequests(pod)
	for _, gpuVendorResourceName := range GPUVendorResourceNames {
		if _, found := podRequests[gpuVendorResourceName]; found {
			return true
		}
	}
	return false
}

// DetectNodeGPUResourceName inspects the node's allocatable resources and returns the first
// known GPU extended resource name that has non-zero allocatable. Falls back to Nvidia for
// backward compatibility if none are found but a GPU label is present.
func DetectNodeGPUResourceName(node *apiv1.Node) apiv1.ResourceName {
	for _, rn := range GPUVendorResourceNames {
		if qty, ok := node.Status.Allocatable[rn]; ok && !qty.IsZero() {
			return rn
		}
	}
	// Fallback: preserve previous behavior (defaulting to Nvidia) if label existed
	return ResourceNvidiaGPU
}

// GetNodeGPUFromCloudProvider returns the GPU the node has. Returned GPU has the GPU label of the
// passed in cloud provider. If the node doesn't have a GPU, returns nil.
func GetNodeGPUFromCloudProvider(provider cloudprovider.CloudProvider, node *apiv1.Node) *cloudprovider.GpuConfig {
	gpuLabel := provider.GPULabel()
	if NodeHasGpu(gpuLabel, node) {
		return &cloudprovider.GpuConfig{
			Label:                gpuLabel,
			Type:                 node.Labels[gpuLabel],
			ExtendedResourceName: DetectNodeGPUResourceName(node),
		}
	}
	return nil
}
