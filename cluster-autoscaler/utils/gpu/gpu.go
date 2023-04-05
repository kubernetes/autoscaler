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
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/klog/v2"
)

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
	resourceName := gpuConfig.ResourceName
	capacity, capacityFound := node.Status.Capacity[resourceName]
	// There is no label value, fallback to generic solution
	if gpuConfig.Type == "" && capacityFound && !capacity.IsZero() {
		return resourceName.String(), MetricsGenericGPU
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
