/*
Copyright 2024 The Kubernetes Authors.

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

package orchestrator

import (
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	ca_context "k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/estimator"
	"k8s.io/autoscaler/cluster-autoscaler/expander"
	"k8s.io/autoscaler/cluster-autoscaler/processors/status"
	"k8s.io/autoscaler/cluster-autoscaler/resourcequotas"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/klog/v2"

	"k8s.io/autoscaler/cluster-autoscaler/core/scaleup/equivalence"
)

// DefaultSimulator implements ScaleUpSimulator using CA's original binpacking logic.
type DefaultSimulator struct {
	orchestrator *ScaleUpOrchestrator
}

// NewDefaultSimulator creates a new DefaultSimulator.
func NewDefaultSimulator(orchestrator *ScaleUpOrchestrator) *DefaultSimulator {
	return &DefaultSimulator{
		orchestrator: orchestrator,
	}
}

// Simulate runs the CA simulation loop.
func (s *DefaultSimulator) Simulate(
	autoscalingCtx *ca_context.AutoscalingContext,
	podEquivalenceGroups []*equivalence.PodGroup,
	unschedulablePods []*apiv1.Pod,
	nodes []*apiv1.Node,
	nodeGroups []cloudprovider.NodeGroup,
	nodeInfos map[string]*framework.NodeInfo,
	tracker *resourcequotas.Tracker,
	now time.Time,
	allOrNothing bool,
) ([][]expander.Option, map[string]status.Reasons, map[string][]estimator.PodEquivalenceGroup, error) {
	// Filter out invalid node groups
	validNodeGroups, skippedNodeGroups := s.orchestrator.filterValidScaleUpNodeGroups(nodeGroups, nodeInfos, tracker, len(nodes), now, nil)

	// Mark skipped node groups as processed.
	for nodegroupID := range skippedNodeGroups {
		s.orchestrator.processors.BinpackingLimiter.MarkProcessed(autoscalingCtx, nodegroupID)
	}

	// Calculate expansion options
	schedulablePodGroups := map[string][]estimator.PodEquivalenceGroup{}
	var options []expander.Option

	for _, nodeGroup := range validNodeGroups {
		schedulablePodGroups[nodeGroup.Id()] = s.orchestrator.SchedulablePodGroups(podEquivalenceGroups, nodeGroup, nodeInfos[nodeGroup.Id()])
	}

	for _, nodeGroup := range validNodeGroups {
		option := s.orchestrator.ComputeExpansionOption(nodeGroup, schedulablePodGroups, nodeInfos, len(nodes), now, allOrNothing)
		option.IsSimilarValid = func(group cloudprovider.NodeGroup, nodeInfo *framework.NodeInfo) bool {
			if group == nil {
				return false
			}
			return len(schedulablePodGroups[group.Id()]) > 0
		}
		s.orchestrator.processors.BinpackingLimiter.MarkProcessed(autoscalingCtx, nodeGroup.Id())

		if len(option.Pods) == 0 || option.NodeCount == 0 {
			klog.V(4).Infof("No pod can fit to %s", nodeGroup.Id())
		} else if allOrNothing && len(option.Pods) < len(unschedulablePods) {
			klog.V(4).Infof("Some pods can't fit to %s, giving up due to all-or-nothing scale-up strategy", nodeGroup.Id())
		} else {
			options = append(options, option)
		}

		if s.orchestrator.processors.BinpackingLimiter.StopBinpacking(autoscalingCtx, options) {
			break
		}
	}

	if len(options) == 0 {
		return nil, skippedNodeGroups, schedulablePodGroups, nil
	}

	return [][]expander.Option{options}, skippedNodeGroups, schedulablePodGroups, nil
}
