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
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/golang/glog"
)

const (
	// ResourceNvidiaGPU is the name of the Nvidia GPU resource.
	ResourceNvidiaGPU = "nvidia.com/gpu"
	// GPULabel is the label added to nodes with GPU resource on GKE.
	GPULabel = "cloud.google.com/gke-accelerator"
	// DefaultGPUType is the type of GPU used in NAP if the user
	// don't specify what type of GPU his pod wants.
	DefaultGPUType = "nvidia-tesla-k80"
)

// FilterOutNodesWithUnreadyGpus removes nodes that should have GPU, but don't have it in allocatable
// from ready nodes list and updates their status to unready on all nodes list.
// This is a hack/workaround for nodes with GPU coming up without installed drivers, resulting
// in GPU missing from their allocatable and capacity.
func FilterOutNodesWithUnreadyGpus(allNodes, readyNodes []*apiv1.Node) ([]*apiv1.Node, []*apiv1.Node) {
	newAllNodes := make([]*apiv1.Node, 0)
	newReadyNodes := make([]*apiv1.Node, 0)
	nodesWithUnreadyGpu := make(map[string]*apiv1.Node)
	for _, node := range readyNodes {
		isUnready := false
		_, hasGpuLabel := node.Labels[GPULabel]
		gpuAllocatable, hasGpuAllocatable := node.Status.Allocatable[ResourceNvidiaGPU]
		// We expect node to have GPU based on label, but it doesn't show up
		// on node object. Assume the node is still not fully started (installing
		// GPU drivers).
		if hasGpuLabel && (!hasGpuAllocatable || gpuAllocatable.IsZero()) {
			newNode, err := getUnreadyNodeCopy(node)
			if err != nil {
				glog.Errorf("Failed to override status of node %v with unready GPU: %v",
					node.Name, err)
			} else {
				glog.V(3).Infof("Overriding status of node %v, which seems to have unready GPU",
					node.Name)
				nodesWithUnreadyGpu[newNode.Name] = newNode
				isUnready = true
			}
		}
		if !isUnready {
			newReadyNodes = append(newReadyNodes, node)
		}
	}
	// Override any node with unready GPU with its "unready" copy
	for _, node := range allNodes {
		if newNode, found := nodesWithUnreadyGpu[node.Name]; found {
			newAllNodes = append(newAllNodes, newNode)
		} else {
			newAllNodes = append(newAllNodes, node)
		}
	}
	return newAllNodes, newReadyNodes
}

func getUnreadyNodeCopy(node *apiv1.Node) (*apiv1.Node, error) {
	newNode := node.DeepCopy()
	newReadyCondition := apiv1.NodeCondition{
		Type:               apiv1.NodeReady,
		Status:             apiv1.ConditionFalse,
		LastTransitionTime: node.CreationTimestamp,
	}
	newNodeConditions := []apiv1.NodeCondition{newReadyCondition}
	for _, condition := range newNode.Status.Conditions {
		if condition.Type != apiv1.NodeReady {
			newNodeConditions = append(newNodeConditions, condition)
		}
	}
	newNode.Status.Conditions = newNodeConditions
	return newNode, nil
}

// GpuRequestInfo contains an information about a set of pods requesting a GPU.
type GpuRequestInfo struct {
	// MaxRequest is maximum GPU request among pods
	MaxRequest resource.Quantity
	// Pods is a list of pods requesting GPU
	Pods []*apiv1.Pod
	// SystemLabels is a set of system labels corresponding to selected GPU
	// that needs to be passed to cloudprovider
	SystemLabels map[string]string
}

// GetGpuRequests returns a GpuRequestInfo for each type of GPU requested by
// any pod in pods argument. If the pod requests GPU, but doesn't specify what
// type of GPU it wants (via NodeSelector) it assumes it's DefaultGPUType.
func GetGpuRequests(pods []*apiv1.Pod) map[string]GpuRequestInfo {
	result := make(map[string]GpuRequestInfo)
	for _, pod := range pods {
		var podGpu resource.Quantity
		for _, container := range pod.Spec.Containers {
			if container.Resources.Requests != nil {
				containerGpu := container.Resources.Requests[ResourceNvidiaGPU]
				podGpu.Add(containerGpu)
			}
		}
		if podGpu.Value() == 0 {
			continue
		}

		gpuType := DefaultGPUType
		if gpuTypeFromSelector, found := pod.Spec.NodeSelector[GPULabel]; found {
			gpuType = gpuTypeFromSelector
		}

		requestInfo, found := result[gpuType]
		if !found {
			requestInfo = GpuRequestInfo{
				MaxRequest: podGpu,
				Pods:       make([]*apiv1.Pod, 0),
				SystemLabels: map[string]string{
					GPULabel: gpuType,
				},
			}
		}
		if podGpu.Cmp(requestInfo.MaxRequest) > 0 {
			requestInfo.MaxRequest = podGpu
		}
		requestInfo.Pods = append(requestInfo.Pods, pod)
		result[gpuType] = requestInfo
	}
	return result
}
