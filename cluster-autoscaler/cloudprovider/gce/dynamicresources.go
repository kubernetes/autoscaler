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
