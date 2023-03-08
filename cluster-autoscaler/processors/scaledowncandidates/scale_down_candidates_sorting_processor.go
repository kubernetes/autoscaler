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
	apiv1 "k8s.io/api/core/v1"

	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodes"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
)

// ScaleDownCandidatesSortingProcessor is a wrapper for preFilteringProcessor that takes into account previous
// scale down candidates. This is necessary for efficient parallel scale down.
type ScaleDownCandidatesSortingProcessor struct {
	preFilter *nodes.PreFilteringScaleDownNodeProcessor
	sorting   []CandidatesComparer
}

// GetPodDestinationCandidates returns nodes that potentially could act as destinations for pods
// that would become unscheduled after a scale down.
func (p *ScaleDownCandidatesSortingProcessor) GetPodDestinationCandidates(ctx *context.AutoscalingContext,
	nodes []*apiv1.Node) ([]*apiv1.Node, errors.AutoscalerError) {
	return p.preFilter.GetPodDestinationCandidates(ctx, nodes)
}

// GetScaleDownCandidates returns filter nodes and move previous scale down candidates to the beginning of the list.
func (p *ScaleDownCandidatesSortingProcessor) GetScaleDownCandidates(ctx *context.AutoscalingContext,
	nodes []*apiv1.Node) ([]*apiv1.Node, errors.AutoscalerError) {
	candidates, err := p.preFilter.GetScaleDownCandidates(ctx, nodes)
	if err != nil {
		return candidates, err
	}
	n := NodeSorter{nodes: candidates, processors: p.sorting}
	return n.Sort(), err
}

// CleanUp is called at CA termination.
func (p *ScaleDownCandidatesSortingProcessor) CleanUp() {
}

// NewScaleDownCandidatesSortingProcessor returns a new PreFilteringScaleDownNodeProcessor.
func NewScaleDownCandidatesSortingProcessor(sorting []CandidatesComparer) *ScaleDownCandidatesSortingProcessor {
	return &ScaleDownCandidatesSortingProcessor{preFilter: nodes.NewPreFilteringScaleDownNodeProcessor(), sorting: sorting}
}
