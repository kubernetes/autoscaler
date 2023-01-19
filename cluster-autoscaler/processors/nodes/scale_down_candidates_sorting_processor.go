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

package nodes

import (
	"sort"
	"time"

	apiv1 "k8s.io/api/core/v1"

	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
)

// ScaleDownCandidatesSortingProcessor is a wrapper for preFilteringProcessor that takes into account previous
// scale down candidates. This is necessary for efficient parallel scale down.
type ScaleDownCandidatesSortingProcessor struct {
	preFilter          *PreFilteringScaleDownNodeProcessor
	previousCandidates *PreviousCandidates
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
	sort.Slice(candidates, func(i, j int) bool {
		return p.previousCandidates.ScaleDownEarlierThan(candidates[i], candidates[j])
	})
	return candidates, nil
}

// CleanUp is called at CA termination.
func (p *ScaleDownCandidatesSortingProcessor) CleanUp() {
}

// NewScaleDownCandidatesSortingProcessor returns a new PreFilteringScaleDownNodeProcessor.
func NewScaleDownCandidatesSortingProcessor() *ScaleDownCandidatesSortingProcessor {
	return &ScaleDownCandidatesSortingProcessor{preFilter: NewPreFilteringScaleDownNodeProcessor(), previousCandidates: NewPreviousCandidates()}
}

// UpdateScaleDownCandidates updates scale down candidates.
func (p *ScaleDownCandidatesSortingProcessor) UpdateScaleDownCandidates(nodes []*apiv1.Node, now time.Time) {
	p.previousCandidates.UpdateScaleDownCandidates(nodes, now)
}
