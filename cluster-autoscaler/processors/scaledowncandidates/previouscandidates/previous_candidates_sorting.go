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

package previouscandidates

import (
	"time"

	apiv1 "k8s.io/api/core/v1"
)

// PreviousCandidates is a struct that store scale down candidates from previous loop.
type PreviousCandidates struct {
	candidates map[string]bool
}

// NewPreviousCandidates return empty PreviousCandidates struct.
func NewPreviousCandidates() *PreviousCandidates {
	return &PreviousCandidates{}
}

// UpdateScaleDownCandidates updates scale down candidates.
func (p *PreviousCandidates) UpdateScaleDownCandidates(nodes []*apiv1.Node, now time.Time) {
	result := make(map[string]bool)
	for _, node := range nodes {
		result[node.Name] = true
	}
	p.candidates = result
}

// ScaleDownEarlierThan return true if node1 is in candidate list and node2 isn't.
func (p *PreviousCandidates) ScaleDownEarlierThan(node1, node2 *apiv1.Node) bool {
	if p.isPreviousCandidate(node1) && !(p.isPreviousCandidate(node2)) {
		return true
	}
	return false
}

func (p *PreviousCandidates) isPreviousCandidate(node *apiv1.Node) bool {
	return p.candidates[node.Name]
}
