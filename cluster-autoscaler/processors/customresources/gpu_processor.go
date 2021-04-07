/*
Copyright 2021 The Kubernetes Authors.

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

package customresources

import (
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	"k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	"k8s.io/klog/v2"
)

// GpuCustomResourcesProcessor handles only the GPU custom resource. It assumes,
// that the GPU may not become allocatable immediately after the node creation.
// It uses additional hacks to predict the type/count of GPUs in that case.
type GpuCustomResourcesProcessor struct {
}

// FilterOutNodesWithUnreadyResources removes nodes that should have GPU, but don't have
// it in allocatable from ready nodes list and updates their status to unready on all nodes list.
// This is a hack/workaround for nodes with GPU coming up without installed drivers, resulting
// in GPU missing from their allocatable and capacity.
func (p *GpuCustomResourcesProcessor) FilterOutNodesWithUnreadyResources(context *context.AutoscalingContext, allNodes, readyNodes []*apiv1.Node) ([]*apiv1.Node, []*apiv1.Node) {
	newAllNodes := make([]*apiv1.Node, 0)
	newReadyNodes := make([]*apiv1.Node, 0)
	nodesWithUnreadyGpu := make(map[string]*apiv1.Node)
	for _, node := range readyNodes {
		_, hasGpuLabel := node.Labels[context.CloudProvider.GPULabel()]
		gpuAllocatable, hasGpuAllocatable := node.Status.Allocatable[gpu.ResourceNvidiaGPU]
		// We expect node to have GPU based on label, but it doesn't show up
		// on node object. Assume the node is still not fully started (installing
		// GPU drivers).
		if hasGpuLabel && (!hasGpuAllocatable || gpuAllocatable.IsZero()) {
			klog.V(3).Infof("Overriding status of node %v, which seems to have unready GPU",
				node.Name)
			nodesWithUnreadyGpu[node.Name] = kubernetes.GetUnreadyNodeCopy(node)
		} else {
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

// GetNodeResourceTargets returns mapping of resource names to their targets.
// This includes resources which are not yet ready to use and visible in kubernetes.
func (p *GpuCustomResourcesProcessor) GetNodeResourceTargets(context *context.AutoscalingContext, node *apiv1.Node, nodeGroup cloudprovider.NodeGroup) ([]CustomResourceTarget, errors.AutoscalerError) {
	gpuTarget, err := p.GetNodeGpuTarget(context.CloudProvider.GPULabel(), node, nodeGroup)
	return []CustomResourceTarget{gpuTarget}, err
}

// GetNodeGpuTarget returns the gpu target of a given node. This includes gpus
// that are not ready to use and visible in kubernetes.
func (p *GpuCustomResourcesProcessor) GetNodeGpuTarget(GPULabel string, node *apiv1.Node, nodeGroup cloudprovider.NodeGroup) (CustomResourceTarget, errors.AutoscalerError) {
	gpuLabel, found := node.Labels[GPULabel]
	if !found {
		return CustomResourceTarget{}, nil
	}

	gpuAllocatable, found := node.Status.Allocatable[gpu.ResourceNvidiaGPU]
	if found && gpuAllocatable.Value() > 0 {
		return CustomResourceTarget{gpuLabel, gpuAllocatable.Value()}, nil
	}

	// A node is supposed to have GPUs (based on label), but they're not available yet
	// (driver haven't installed yet?).
	// Unfortunately we can't deduce how many GPUs it will actually have from labels (just
	// that it will have some).
	// Ready for some evil hacks? Well, you won't be disappointed - let's pretend we haven't
	// seen the node and just use the template we use for scale from 0. It'll be our little
	// secret.

	if nodeGroup == nil {
		// We expect this code path to be triggered by situation when we are looking at a node which is expected to have gpus (has gpu label)
		// But those are not yet visible in node's resource (e.g. gpu drivers are still being installed).
		// In case of node coming from autoscaled node group we would look and node group template here.
		// But for nodes coming from non-autoscaled groups we have no such possibility.
		// Let's hope it is a transient error. As long as it exists we will not scale nodes groups with gpus.
		return CustomResourceTarget{}, errors.NewAutoscalerError(errors.InternalError, "node without with gpu label, without capacity not belonging to autoscaled node group")
	}

	template, err := nodeGroup.TemplateNodeInfo()
	if err != nil {
		klog.Errorf("Failed to build template for getting GPU estimation for node %v: %v", node.Name, err)
		return CustomResourceTarget{}, errors.ToAutoscalerError(errors.CloudProviderError, err)
	}
	if gpuCapacity, found := template.Node().Status.Capacity[gpu.ResourceNvidiaGPU]; found {
		return CustomResourceTarget{gpuLabel, gpuCapacity.Value()}, nil
	}

	// if template does not define gpus we assume node will not have any even if ith has gpu label
	klog.Warningf("Template does not define gpus even though node from its node group does; node=%v", node.Name)
	return CustomResourceTarget{}, nil
}

// CleanUp cleans up processor's internal structures.
func (p *GpuCustomResourcesProcessor) CleanUp() {
}
