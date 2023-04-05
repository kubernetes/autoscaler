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

package util

import (
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
)

const (
	// ResourceNvidiaGPU is the name of the Nvidia GPU resource.
	ResourceNvidiaGPU = "nvidia.com/gpu"
	// ResourceDirectX is the name of the DirectX resource on windows.
	ResourceDirectX = "microsoft.com/directx"
	// DefaultGPUType is the type of GPU used in NAP if the user
	// don't specify what type of GPU his pod wants.
	DefaultGPUType = "nvidia-tesla-k80"
)

// NodeHasGpu returns true if a given node has GPU hardware.
// The result will be true if there is hardware capability. It doesn't matter
// if the drivers are installed and GPU is ready to use.
func NodeHasGpu(GPULabel string, node *apiv1.Node) bool {
	_, hasGpuLabel := node.Labels[GPULabel]
	gpuAllocatable, hasGpuAllocatable := node.Status.Allocatable[ResourceNvidiaGPU]
	return hasGpuLabel || (hasGpuAllocatable && !gpuAllocatable.IsZero())
}

// PodRequestsGpu returns true if a given pod has GPU request.
func PodRequestsGpu(pod *apiv1.Pod) bool {
	for _, container := range pod.Spec.Containers {
		if container.Resources.Requests != nil {
			_, gpuFound := container.Resources.Requests[ResourceNvidiaGPU]
			if gpuFound {
				return true
			}
		}
	}
	return false
}

// GetNodeGPUFromCloudProvider returns the GPU the node has. Returned GPU has the GPU label of the
// passed in cloud provider. If the node doesn't have a GPU, returns nil.
func GetNodeGPUFromCloudProvider(provider cloudprovider.CloudProvider, node *apiv1.Node) *cloudprovider.GpuConfig {
	gpuLabel := provider.GPULabel()
	if NodeHasGpu(gpuLabel, node) {
		return &cloudprovider.GpuConfig{Label: gpuLabel, Type: node.Labels[gpuLabel], ResourceName: ResourceNvidiaGPU}
	}
	return nil
}
