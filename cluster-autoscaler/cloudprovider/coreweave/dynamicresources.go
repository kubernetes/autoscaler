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

package coreweave

import apiv1 "k8s.io/api/core/v1"

const (
	// DraGPUDriver is the driver used to expose NVIDIA GPU resources via DRA.
	DraGPUDriver = "gpu.nvidia.com"
	// DraGPULabel identifies nodes exposing NVIDIA GPUs through DRA instead
	// of the device plugin. Set it to "true" in NodePool.spec.nodeLabels and
	// ensure it is present when nodes first register.
	DraGPULabel = "node.coreweave.cloud/dra-gpu-enabled"
)

// NodeGpuDraDriverEnabled checks whether the GPU DRA driver is enabled on the node.
func NodeGpuDraDriverEnabled(node *apiv1.Node) bool {
	return node.Labels[DraGPULabel] == "true"
}

// NodePoolGpuDraDriverEnabled checks whether the GPU DRA driver is enabled on the nodepool.
func NodePoolGpuDraDriverEnabled(nodepool *CoreWeaveNodePool) bool {
	return nodepool.GetNodeLabels()[DraGPULabel] == "true"
}
