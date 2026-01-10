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
	apiv1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	ca_context "k8s.io/autoscaler/cluster-autoscaler/context"
	csisnapshot "k8s.io/autoscaler/cluster-autoscaler/simulator/csi/snapshot"
	drasnapshot "k8s.io/autoscaler/cluster-autoscaler/simulator/dynamicresources/snapshot"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	"k8s.io/klog/v2"
)

// CSICustomResourcesProcessor is a processor that filters out nodes with unready CSI resources.
type CSICustomResourcesProcessor struct {
}

// FilterOutNodesWithUnreadyResources filters out nodes with unready CSI resources.
func (p *CSICustomResourcesProcessor) FilterOutNodesWithUnreadyResources(autoscalingCtx *ca_context.AutoscalingContext, allNodes, readyNodes []*apiv1.Node, _ *drasnapshot.Snapshot, csiSnapshot *csisnapshot.Snapshot) ([]*apiv1.Node, []*apiv1.Node) {
	newAllNodes := make([]*apiv1.Node, 0)
	newReadyNodes := make([]*apiv1.Node, 0)
	nodesWithUnreadyCSI := make(map[string]*apiv1.Node)
	if csiSnapshot == nil {
		klog.Warningf("Cannot filter out nodes with unready CSI resources. The CSI snapshot is nil. Processing will be skipped.")
		return allNodes, readyNodes
	}

	for _, node := range readyNodes {
		ng, err := autoscalingCtx.CloudProvider.NodeGroupForNode(node)
		if err != nil {
			newReadyNodes = append(newReadyNodes, node)
			klog.Warningf("Failed to get node group for node %s, Skipping CSI readiness check and keeping node in ready list. Error: %v", node.Name, err)
			continue
		}
		if ng == nil {
			newReadyNodes = append(newReadyNodes, node)
			continue
		}

		// TODO: Use TemplateNodeInfoRegistry after #8882 is merged
		templateNodeInfo, err := ng.TemplateNodeInfo()
		if err != nil {
			newReadyNodes = append(newReadyNodes, node)
			klog.Warningf("Failed to get template node info for node group %s with error: %v", ng.Id(), err)
			continue
		}

		// if cloudprovider does not provide CSI related stuff, then we can skip the CSI readiness check
		if templateNodeInfo.CSINode == nil {
			newReadyNodes = append(newReadyNodes, node)
			klog.V(5).Infof("No CSI node found for node %s, Skipping CSI readiness check and keeping node in ready list.", node.Name)
			continue
		}

		csiNode, err := csiSnapshot.Get(node.Name)
		if err != nil {
			newReadyNodes = append(newReadyNodes, node)
			klog.V(5).Infof("Failed to get CSI node for node %s, Skipping CSI readiness check and keeping node in ready list. Error: %v", node.Name, err)
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
func (p *CSICustomResourcesProcessor) GetNodeResourceTargets(autoscalingCtx *ca_context.AutoscalingContext, node *apiv1.Node, nodeGroup cloudprovider.NodeGroup) ([]CustomResourceTarget, errors.AutoscalerError) {
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
