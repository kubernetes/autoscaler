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

type combinedScaleDownCandidatesProcessor struct {
	processors []nodes.ScaleDownNodeProcessor
}

// NewCombinedScaleDownCandidatesProcessor returns a default implementation of the scale down candidates
// processor, which wraps and sequentially runs other sub-processors.
func NewCombinedScaleDownCandidatesProcessor() *combinedScaleDownCandidatesProcessor {
	return &combinedScaleDownCandidatesProcessor{}

}

// Register registers a new ScaleDownNodeProcessor
func (p *combinedScaleDownCandidatesProcessor) Register(np nodes.ScaleDownNodeProcessor) {
	p.processors = append(p.processors, np)
}

// GetPodDestinationCandidates returns nodes that potentially could act as destinations for pods
// that would become unscheduled after a scale down.
func (p *combinedScaleDownCandidatesProcessor) GetPodDestinationCandidates(ctx *context.AutoscalingContext, nodes []*apiv1.Node) ([]*apiv1.Node, errors.AutoscalerError) {
	var err errors.AutoscalerError
	for _, processor := range p.processors {
		nodes, err = processor.GetPodDestinationCandidates(ctx, nodes)
		if err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// GetScaleDownCandidates returns nodes that potentially could be scaled down.
func (p *combinedScaleDownCandidatesProcessor) GetScaleDownCandidates(ctx *context.AutoscalingContext, nodes []*apiv1.Node) ([]*apiv1.Node, errors.AutoscalerError) {
	var err errors.AutoscalerError
	for _, processor := range p.processors {
		nodes, err = processor.GetScaleDownCandidates(ctx, nodes)
		if err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// CleanUp is called at CA termination
func (p *combinedScaleDownCandidatesProcessor) CleanUp() {
	for _, processor := range p.processors {
		processor.CleanUp()
	}
}
