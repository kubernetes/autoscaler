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
	"reflect"
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
)

// ScaleDownCandidatesDelayProcessor is a processor to filter out
// nodes according to scale down delay per nodegroup
type ScaleDownCandidatesDelayProcessor struct {
	scaleUps          map[string]time.Time
	scaleDowns        map[string]time.Time
	scaleDownFailures map[string]time.Time
}

// GetPodDestinationCandidates returns nodes as is no processing is required here
func (p *ScaleDownCandidatesDelayProcessor) GetPodDestinationCandidates(ctx *context.AutoscalingContext,
	nodes []*apiv1.Node) ([]*apiv1.Node, errors.AutoscalerError) {
	return nodes, nil
}

// GetScaleDownCandidates returns filter nodes based on if scale down is enabled or disabled per nodegroup.
func (p *ScaleDownCandidatesDelayProcessor) GetScaleDownCandidates(ctx *context.AutoscalingContext,
	nodes []*apiv1.Node) ([]*apiv1.Node, errors.AutoscalerError) {
	result := []*apiv1.Node{}

	for _, node := range nodes {
		nodeGroup, err := ctx.CloudProvider.NodeGroupForNode(node)
		if err != nil {
			klog.Warningf("Error while checking node group for %s: %v", node.Name, err)
			continue
		}
		if nodeGroup == nil || reflect.ValueOf(nodeGroup).IsNil() {
			klog.V(4).Infof("Node %s should not be processed by cluster autoscaler (no node group config)", node.Name)
			continue
		}

		currentTime := time.Now()

		recent := func(m map[string]time.Time, d time.Duration, msg string) bool {
			if !m[nodeGroup.Id()].IsZero() && m[nodeGroup.Id()].Add(d).After(currentTime) {
				klog.V(4).Infof("Skipping scale down on node group %s because it %s recently at %v",
					nodeGroup.Id(), msg, m[nodeGroup.Id()])
				return true
			}

			return false
		}

		if recent(p.scaleUps, ctx.ScaleDownDelayAfterAdd, "scaled up") {
			continue
		}

		if recent(p.scaleDowns, ctx.ScaleDownDelayAfterDelete, "scaled down") {
			continue
		}

		if recent(p.scaleDownFailures, ctx.ScaleDownDelayAfterFailure, "failed to scale down") {
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
func (p *ScaleDownCandidatesDelayProcessor) RegisterScaleUp(nodeGroup cloudprovider.NodeGroup,
	_ int, currentTime time.Time) {
	p.scaleUps[nodeGroup.Id()] = currentTime
}

// RegisterScaleDown records when the last scale down happened for a nodegroup.
func (p *ScaleDownCandidatesDelayProcessor) RegisterScaleDown(nodeGroup cloudprovider.NodeGroup,
	nodeName string, currentTime time.Time, _ time.Time) {
	p.scaleDowns[nodeGroup.Id()] = currentTime
}

// RegisterFailedScaleUp records when the last scale up failed for a nodegroup.
func (p *ScaleDownCandidatesDelayProcessor) RegisterFailedScaleUp(_ cloudprovider.NodeGroup,
	_ string, _ string, _ string, _ string, _ time.Time) {
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
