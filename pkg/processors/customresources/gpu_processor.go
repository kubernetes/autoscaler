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
	"context"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/cluster-autoscaler/pkg/cloudprovider"
	ca_context "sigs.k8s.io/cluster-autoscaler/pkg/context"
	csisnapshot "sigs.k8s.io/cluster-autoscaler/pkg/simulator/csi/snapshot"
	drasnapshot "sigs.k8s.io/cluster-autoscaler/pkg/simulator/dynamicresources/snapshot"
	"sigs.k8s.io/cluster-autoscaler/pkg/utils/errors"
	"sigs.k8s.io/cluster-autoscaler/pkg/utils/gpu"
	"sigs.k8s.io/cluster-autoscaler/pkg/utils/kubernetes"
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
func (p *GpuCustomResourcesProcessor) FilterOutNodesWithUnreadyResources(ctx context.Context, autoscalingCtx *ca_context.AutoscalingContext, allNodes, readyNodes []*apiv1.Node, _ *drasnapshot.Snapshot, _ *csisnapshot.Snapshot) ([]*apiv1.Node, []*apiv1.Node) {
	logger := klog.FromContext(ctx)
	newAllNodes := make([]*apiv1.Node, 0)
	newReadyNodes := make([]*apiv1.Node, 0)
	nodesWithUnreadyGpu := make(map[string]*apiv1.Node)
	for _, node := range readyNodes {
		if gpuExposedViaDra(ctx, autoscalingCtx, node) {
			newReadyNodes = append(newReadyNodes, node)
			continue
		}

		_, hasGpuLabel := node.Labels[autoscalingCtx.CloudProvider.GPULabel(ctx)]
		_, hasAnyGpuAllocatable := gpu.NodeHasGpuAllocatable(node)
		if hasGpuLabel && !hasAnyGpuAllocatable {
			logger.V(3).Info("Overriding status of node that seems to have unready GPU", "node", klog.KObj(node))
			nodesWithUnreadyGpu[node.Name] = kubernetes.GetUnreadyNodeCopy(node, kubernetes.ResourceUnready)
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
func (p *GpuCustomResourcesProcessor) GetNodeResourceTargets(ctx context.Context, autoscalingCtx *ca_context.AutoscalingContext, node *apiv1.Node, nodeGroup cloudprovider.NodeGroup) ([]CustomResourceTarget, errors.AutoscalerError) {
	gpuTarget, err := p.GetNodeGpuTarget(ctx, autoscalingCtx, node, nodeGroup)
	return []CustomResourceTarget{gpuTarget}, err
}

// GetNodeGpuTarget returns the gpu target of a given node. This includes gpus
// that are not ready to use and visible in kubernetes.
func (p *GpuCustomResourcesProcessor) GetNodeGpuTarget(ctx context.Context, autoscalingCtx *ca_context.AutoscalingContext, node *apiv1.Node, nodeGroup cloudprovider.NodeGroup) (CustomResourceTarget, errors.AutoscalerError) {
	logger := klog.FromContext(ctx)
	gpuLabel, found := node.Labels[autoscalingCtx.CloudProvider.GPULabel(ctx)]
	if !found {
		return CustomResourceTarget{}, nil
	}

	if gpuExposedViaDra(ctx, autoscalingCtx, node) {
		return CustomResourceTarget{}, nil
	}

	if gpuAllocatableValue, hasGpuAllocatable := gpu.NodeHasGpuAllocatable(node); hasGpuAllocatable {
		return CustomResourceTarget{gpuLabel, gpuAllocatableValue}, nil
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

	template, err := nodeGroup.TemplateNodeInfo(ctx)
	if err != nil {
		logger.Error(err, "Failed to build template for getting GPU estimation for node", "node", klog.KObj(node))
		return CustomResourceTarget{}, errors.ToAutoscalerError(errors.CloudProviderError, err)
	}
	for _, gpuVendorResourceName := range gpu.GPUVendorResourceNames {
		if gpuCapacity, found := template.Node().Status.Capacity[gpuVendorResourceName]; found {
			return CustomResourceTarget{gpuLabel, gpuCapacity.Value()}, nil
		}
	}
	// if template does not define gpus we assume node will not have any even if ith has gpu label
	logger.Info("Template does not define gpus even though node from its node group does", "node", klog.KObj(node))
	return CustomResourceTarget{}, nil
}

// CleanUp cleans up processor's internal structures.
func (p *GpuCustomResourcesProcessor) CleanUp() {
}

func gpuExposedViaDra(ctx context.Context, autoscalingCtx *ca_context.AutoscalingContext, node *apiv1.Node) bool {
	gpuConfig := autoscalingCtx.CloudProvider.GetNodeGpuConfig(ctx, node)
	if gpuConfig == nil {
		return false
	}

	// Devices attached through DRA are not using node allocatable
	// to confirm their attachment, assume that node is ready
	// and will be checked in the separate processor
	return gpuConfig.ExposedViaDra()
}
