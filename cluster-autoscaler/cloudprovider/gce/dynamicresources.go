/*
Copyright 2025 The Kubernetes Authors.

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

import apiv1 "k8s.io/api/core/v1"

const (
	// DraGPUDriver name of the driver used to expose NVIDIA GPU resources
	DraGPUDriver = "gpu.nvidia.com"
	// DraGPULabel is the label added to nodes with GPU resource exposed via DRA.
	DraGPULabel = "cloud.google.com/gke-gpu-dra-driver"
)

// GpuDraDriverEnabled checks whether GPU driver is enabled on the node
func GpuDraDriverEnabled(node *apiv1.Node) bool {
	return node.Labels[DraGPULabel] == "true"
}
