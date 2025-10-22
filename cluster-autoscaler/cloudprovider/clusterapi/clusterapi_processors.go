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

package clusterapi

import (
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	klog "k8s.io/klog/v2"
)

// ScaleDownNodeUpgradeProcessor is a processor to filter out
// nodes that are undergoing an upgrade through a MachineDeployment.
type ScaleDownNodeUpgradeProcessor struct {
	controller *machineController
}

// NewScaleDownNodeUpgradeProcessor returns a new ScaleDownNodeUpgradeProcessor for use when
// registering a new upgrade scale down processor.
func NewScaleDownNodeUpgradeProcessor(c *machineController) *ScaleDownNodeUpgradeProcessor {
	return &ScaleDownNodeUpgradeProcessor{controller: c}
}

// GetPodDestinationCandidates returns nodes as is no processing is required here
func (p *ScaleDownNodeUpgradeProcessor) GetPodDestinationCandidates(ctx *context.AutoscalingContext,
	nodes []*apiv1.Node) ([]*apiv1.Node, errors.AutoscalerError) {
	return nodes, nil
}

// GetScaleDownCandidates returns filter nodes based on if scale down is enabled or disabled per nodegroup.
func (p *ScaleDownNodeUpgradeProcessor) GetScaleDownCandidates(ctx *context.AutoscalingContext,
	nodes []*apiv1.Node) ([]*apiv1.Node, errors.AutoscalerError) {
	result := []*apiv1.Node{}

	for _, node := range nodes {
		// check scale down, continue if not good
		ng, err := p.controller.nodeGroupForNode(node)
		if err != nil {
			klog.Warningf("Error while checking node group for node %s: %v", node.Name, err)
			continue
		}
		if ng == nil {
			// this is at level 6 because the core scale down processor will already log this info
			klog.V(6).Infof("Node %s will be skipped as it does not belong to a node group", node.Name)
			continue
		}

		rollingout, err := ng.IsMachineDeploymentAndRollingOut()
		if err != nil {
			klog.Warningf("Failed to determine rolling out status for MachineDeployment %s: %v", ng.scalableResource.ID(), err)
			continue
		}

		// A node is a good candidate for scale down if it is not currently part of a MachineDeployment that is rolling out.
		if rollingout {
			klog.V(4).Infof("Node %s will be skipped as it is currently under rollout", node.Name)
			continue
		}
		result = append(result, node)
	}
	return result, nil
}

// CleanUp is called at CA termination.
func (p *ScaleDownNodeUpgradeProcessor) CleanUp() {
}
