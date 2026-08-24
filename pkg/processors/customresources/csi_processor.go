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

package customresources

import (
	"context"

	apiv1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/cluster-autoscaler/pkg/cloudprovider"
	ca_context "sigs.k8s.io/cluster-autoscaler/pkg/context"
	csisnapshot "sigs.k8s.io/cluster-autoscaler/pkg/simulator/csi/snapshot"
	drasnapshot "sigs.k8s.io/cluster-autoscaler/pkg/simulator/dynamicresources/snapshot"
	"sigs.k8s.io/cluster-autoscaler/pkg/utils/errors"
	"sigs.k8s.io/cluster-autoscaler/pkg/utils/kubernetes"
)

// CSICustomResourcesProcessor is a processor that filters out nodes with unready CSI resources.
type CSICustomResourcesProcessor struct {
}

// FilterOutNodesWithUnreadyResources filters out nodes with unready CSI resources.
func (p *CSICustomResourcesProcessor) FilterOutNodesWithUnreadyResources(ctx context.Context, autoscalingCtx *ca_context.AutoscalingContext, allNodes, readyNodes []*apiv1.Node, _ *drasnapshot.Snapshot, csiSnapshot *csisnapshot.Snapshot) ([]*apiv1.Node, []*apiv1.Node) {
	logger := klog.FromContext(ctx)
	newAllNodes := make([]*apiv1.Node, 0)
	newReadyNodes := make([]*apiv1.Node, 0)
	nodesWithUnreadyCSI := make(map[string]*apiv1.Node)
	if csiSnapshot == nil {
		logger.Info("Cannot filter out nodes with unready CSI resources. The CSI snapshot is nil. Processing will be skipped.")

		return allNodes, readyNodes
	}

	for _, node := range readyNodes {
		ng, err := autoscalingCtx.CloudProvider.NodeGroupForNode(ctx, node)
		if err != nil {
			newReadyNodes = append(newReadyNodes, node)
			logger.Info("Failed to get node group for node, skipping CSI readiness check and keeping node in ready list", "node", klog.KObj(node), "err", err)
			continue
		}
		if ng == nil {
			newReadyNodes = append(newReadyNodes, node)
			continue
		}

		// TODO: Use TemplateNodeInfoRegistry after #8882 is merged
		templateNodeInfo, err := ng.TemplateNodeInfo(ctx)
		if err != nil {
			newReadyNodes = append(newReadyNodes, node)
			logger.Info("Failed to get template node info for node group", "nodeGroupId", ng.Id(), "err", err)
			continue
		}

		// if cloudprovider does not provide CSI related stuff, then we can skip the CSI readiness check
		if templateNodeInfo.CSINode == nil {
			newReadyNodes = append(newReadyNodes, node)
			logger.V(5).Info("No CSI node found for node, skipping CSI readiness check and keeping node in ready list.", "node", klog.KObj(node))
			continue
		}

		csiNode, err := csiSnapshot.Get(node.Name)
		if err != nil {
			newReadyNodes = append(newReadyNodes, node)
			logger.V(5).Info("Failed to get CSI node for node, skipping CSI readiness check and keeping node in ready list", "node", klog.KObj(node), "err", err)
			continue
		}

		if areDriversInstalled(csiNode, templateNodeInfo.CSINode) {
			newReadyNodes = append(newReadyNodes, node)
		} else {
			nodesWithUnreadyCSI[node.Name] = kubernetes.GetUnreadyNodeCopy(node, kubernetes.ResourceUnready)
		}
	}
	for _, node := range allNodes {
		if newNode, found := nodesWithUnreadyCSI[node.Name]; found {
			newAllNodes = append(newAllNodes, newNode)
		} else {
			newAllNodes = append(newAllNodes, node)
		}
	}
	return newAllNodes, newReadyNodes
}

// GetNodeResourceTargets returns mapping of resource names to their targets.
// CSI processor doesn't track resource targets, so it returns an empty list.
func (p *CSICustomResourcesProcessor) GetNodeResourceTargets(ctx context.Context, autoscalingCtx *ca_context.AutoscalingContext, node *apiv1.Node, nodeGroup cloudprovider.NodeGroup) ([]CustomResourceTarget, errors.AutoscalerError) {
	return []CustomResourceTarget{}, nil
}

// CleanUp cleans up processor's internal structures.
func (p *CSICustomResourcesProcessor) CleanUp() {
}

func areDriversInstalled(csiNode *storagev1.CSINode, templateCSINode *storagev1.CSINode) bool {
	defaultDrivers := templateCSINode.Spec.Drivers

	installedDrivers := make(map[string]bool)
	for _, csiDriver := range csiNode.Spec.Drivers {
		installedDrivers[csiDriver.Name] = true
	}
	for _, driver := range defaultDrivers {
		if _, found := installedDrivers[driver.Name]; !found {
			return false
		}
	}
	return true
}
