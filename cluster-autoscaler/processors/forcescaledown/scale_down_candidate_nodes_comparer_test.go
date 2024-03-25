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
	"testing"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
)

func TestScaleDownEarlierThan(t *testing.T) {
	taint := apiv1.Taint{
		Key:    taints.ForceScaleDownTaint,
		Effect: apiv1.TaintEffectNoSchedule,
	}
	candidate1 := BuildTestNode("candidate1", 100, 0)
	candidate1.Spec.Taints = []apiv1.Taint{taint}
	candidate2 := BuildTestNode("candidate2", 100, 0)
	candidate2.Spec.Taints = []apiv1.Taint{taint}
	nonCandidate1 := BuildTestNode("non-candidate1", 100, 0)
	nonCandidate2 := BuildTestNode("non-candidate2", 100, 0)

	processor := NewScaleDownCandidateNodesCompare()
	testCases := []struct {
		name  string
		node1 *apiv1.Node
		node2 *apiv1.Node
		want  bool
	}{
		{
			name:  "Compare two candidates",
			node1: candidate1,
			node2: candidate2,
			want:  false,
		},
		{
			name:  "Compare two non-candidates",
			node1: nonCandidate1,
			node2: nonCandidate2,
			want:  false,
		},
		{
			name:  "Compare candidate and non-candidate",
			node1: candidate1,
			node2: nonCandidate2,
			want:  true,
		},
		{
			name:  "Compare non-candidate and candidate",
			node1: nonCandidate1,
			node2: candidate2,
			want:  false,
		},
	}
	for index := range testCases {
		tc := testCases[index]
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := processor.ScaleDownEarlierThan(tc.node1, tc.node2)
			assert.Equal(t, got, tc.want)
		})
	}
}
