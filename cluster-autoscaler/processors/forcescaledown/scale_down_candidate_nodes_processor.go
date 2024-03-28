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

package forcescaledown

import (
	apiv1 "k8s.io/api/core/v1"

	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"
)

// ScaleDownCandidateNodesProcessor is a processor to handle pod destination candidates and scale down candidates.
type ScaleDownCandidateNodesProcessor struct{}

// NewScaleDownCandidateNodesProcessor returns a new ScaleDownCandidateNodesProcessor struct.
func NewScaleDownCandidateNodesProcessor() *ScaleDownCandidateNodesProcessor {
	return &ScaleDownCandidateNodesProcessor{}
}

// GetPodDestinationCandidates returns nodes that potentially could act as destinations for pods
// that would become unscheduled after a scale down.
func (p *ScaleDownCandidateNodesProcessor) GetPodDestinationCandidates(ctx *context.AutoscalingContext, nodes []*apiv1.Node) ([]*apiv1.Node, errors.AutoscalerError) {
	result := []*apiv1.Node{}
	for _, node := range nodes {
		if !taints.HasForceScaleDownTaint(node) {
			result = append(result, node)
		}
	}
	return result, nil
}

// GetScaleDownCandidates returns the scale down candidate nodes from the input.
func (p *ScaleDownCandidateNodesProcessor) GetScaleDownCandidates(ctx *context.AutoscalingContext, nodes []*apiv1.Node) ([]*apiv1.Node, errors.AutoscalerError) {
	return nodes, nil
}

// CleanUp is called at CA termination.
func (p *ScaleDownCandidateNodesProcessor) CleanUp() {}
