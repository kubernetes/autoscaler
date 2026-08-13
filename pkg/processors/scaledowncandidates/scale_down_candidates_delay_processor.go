/*
Copyright 2023 The Kubernetes Authors.

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

package scaledowncandidates

import (
	"context"
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"

	"sigs.k8s.io/cluster-autoscaler/pkg/cloudprovider"
	ca_context "sigs.k8s.io/cluster-autoscaler/pkg/context"
	"sigs.k8s.io/cluster-autoscaler/pkg/utils/errors"
)

// ScaleDownCandidatesDelayProcessor is a processor to filter out
// nodes according to scale down delay per nodegroup
type ScaleDownCandidatesDelayProcessor struct {
	scaleUps          map[string]time.Time
	scaleDowns        map[string]time.Time
	scaleDownFailures map[string]time.Time
}

// GetPodDestinationCandidates returns nodes as is no processing is required here
func (p *ScaleDownCandidatesDelayProcessor) GetPodDestinationCandidates(autoscalingCtx *ca_context.AutoscalingContext,
	nodes []*apiv1.Node) ([]*apiv1.Node, errors.AutoscalerError) {
	return nodes, nil
}

// GetScaleDownCandidates returns filter nodes based on if scale down is enabled or disabled per nodegroup.
func (p *ScaleDownCandidatesDelayProcessor) GetScaleDownCandidates(ctx context.Context, autoscalingCtx *ca_context.AutoscalingContext,
	nodes []*apiv1.Node) ([]*apiv1.Node, errors.AutoscalerError) {
	logger := klog.FromContext(ctx)
	result := []*apiv1.Node{}
	alreadyLoggedGroups := make(map[string]bool)

	for _, node := range nodes {
		nodeGroup, err := autoscalingCtx.CloudProvider.NodeGroupForNode(ctx, node)
		if err != nil {
			logger.Info("Error while checking node group for node", "node", klog.KObj(node), "err", err)
			continue
		}
		if nodeGroup == nil {
			logger.V(4).Info("Node should not be processed by cluster autoscaler (no node group config)", "node", klog.KObj(node))
			continue
		}

		currentTime := time.Now()

		recent := func(m map[string]time.Time, d time.Duration, msg string) bool {
			if !m[nodeGroup.Id()].IsZero() && m[nodeGroup.Id()].Add(d).After(currentTime) {
				if !alreadyLoggedGroups[nodeGroup.Id()] {
					logger.V(4).Info("Skipping scale down on node group due to recent activity", "nodeGroupId", nodeGroup.Id(), "activity", msg, "time", m[nodeGroup.Id()])
					alreadyLoggedGroups[nodeGroup.Id()] = true
				}
				return true
			}

			return false
		}

		if recent(p.scaleUps, autoscalingCtx.ScaleDownDelayAfterAdd, "scaled up") {
			continue
		}

		if recent(p.scaleDowns, autoscalingCtx.ScaleDownDelayAfterDelete, "scaled down") {
			continue
		}

		if recent(p.scaleDownFailures, autoscalingCtx.ScaleDownDelayAfterFailure, "failed to scale down") {
			continue
		}

		result = append(result, node)
	}
	return result, nil
}

// CleanUp is called at CA termination.
func (p *ScaleDownCandidatesDelayProcessor) CleanUp() {
}

// RegisterScaleUp records when the last scale up happened for a nodegroup.
func (p *ScaleDownCandidatesDelayProcessor) RegisterScaleUp(ctx context.Context, nodeGroup cloudprovider.NodeGroup,
	_ int, currentTime time.Time) {
	p.scaleUps[nodeGroup.Id()] = currentTime
}

// RegisterScaleDown records when the last scale down happened for a nodegroup.
func (p *ScaleDownCandidatesDelayProcessor) RegisterScaleDown(nodeGroup cloudprovider.NodeGroup,
	nodeName string, currentTime time.Time, _ time.Time) {
	p.scaleDowns[nodeGroup.Id()] = currentTime
}

// RegisterFailedScaleUp records when the last scale up failed for a nodegroup.
func (p *ScaleDownCandidatesDelayProcessor) RegisterFailedScaleUp(ctx context.Context, _ cloudprovider.NodeGroup, _ int, _ cloudprovider.InstanceErrorInfo, currentTime time.Time) {
}

// RegisterFailedScaleDown records failed scale-down for a nodegroup.
func (p *ScaleDownCandidatesDelayProcessor) RegisterFailedScaleDown(nodeGroup cloudprovider.NodeGroup,
	reason string, currentTime time.Time) {
	p.scaleDownFailures[nodeGroup.Id()] = currentTime
}

// NewScaleDownCandidatesDelayProcessor returns a new ScaleDownCandidatesDelayProcessor.
func NewScaleDownCandidatesDelayProcessor() *ScaleDownCandidatesDelayProcessor {
	return &ScaleDownCandidatesDelayProcessor{
		scaleUps:          make(map[string]time.Time),
		scaleDowns:        make(map[string]time.Time),
		scaleDownFailures: make(map[string]time.Time),
	}
}
