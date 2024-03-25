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
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
)

func TestGetPodDestinationCandidates(t *testing.T) {
	testCases := []struct {
		name       string
		nodes      []*apiv1.Node
		forcedNods []string
		wantNodes  []*apiv1.Node
	}{
		{
			name: "no nodes and no pod destination candidates",
		},
		{
			name: "force-scale-down nodes should not be returned",
			nodes: []*apiv1.Node{
				BuildTestNode("n1", 1000, 10),
				BuildTestNode("n2", 1000, 10),
			},
			forcedNods: []string{"n1"},
			wantNodes: []*apiv1.Node{
				BuildTestNode("n2", 1000, 10),
			},
		},
	}

	for index := range testCases {
		tc := testCases[index]
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			forcedNodeNames := map[string]bool{}
			for _, name := range tc.forcedNods {
				forcedNodeNames[name] = true
			}
			for index := range tc.nodes {
				if forcedNodeNames[tc.nodes[index].Name] {
					taint := apiv1.Taint{Key: taints.ForceScaleDownTaint, Effect: apiv1.TaintEffectNoSchedule}
					tc.nodes[index].Spec.Taints = append(tc.nodes[index].Spec.Taints, taint)
				}
			}
			ctx := &context.AutoscalingContext{}
			processor := NewScaleDownCandidateNodesProcessor()
			result, err := processor.GetPodDestinationCandidates(ctx, tc.nodes)
			assert.NoError(t, err)
			assert.ElementsMatch(t, tc.wantNodes, result)
		})
	}
}
