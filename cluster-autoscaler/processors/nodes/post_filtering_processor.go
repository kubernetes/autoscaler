/*
Copyright 2021 The Kubernetes Authors.

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
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"
)

// PostFilteringScaleDownNodeProcessor selects first maxCount nodes (if possible) to be removed
type PostFilteringScaleDownNodeProcessor struct {
}

// GetNodesToRemove selects up to maxCount nodes for deletion, by selecting a first maxCount candidates
func (n *PostFilteringScaleDownNodeProcessor) GetNodesToRemove(ctx *context.AutoscalingContext, candidates []simulator.NodeToBeRemoved, maxCount int) []simulator.NodeToBeRemoved {
	end := len(candidates)
	if len(candidates) > maxCount {
		end = maxCount
	}
	return candidates[:end]
}

// CleanUp is called at CA termination
func (n *PostFilteringScaleDownNodeProcessor) CleanUp() {
}

// NewPostFilteringScaleDownNodeProcessor returns a new PostFilteringScaleDownNodeProcessor
func NewPostFilteringScaleDownNodeProcessor() *PostFilteringScaleDownNodeProcessor {
	return &PostFilteringScaleDownNodeProcessor{}
}
