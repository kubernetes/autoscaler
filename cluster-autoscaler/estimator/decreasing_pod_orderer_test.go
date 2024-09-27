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

package estimator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/utils/test"
)

func TestPodPriorityProcessor(t *testing.T) {
	pg1 := PodEquivalenceGroup{Pods: []*v1.Pod{test.BuildTestPod("p1", 1, 1)}}
	pg2 := PodEquivalenceGroup{Pods: []*v1.Pod{test.BuildTestPod("p2", 2, 1)}}
	pg3 := PodEquivalenceGroup{Pods: []*v1.Pod{test.BuildTestPod("p3", 2, 100)}}
	node := makeNode(4, 600, 10, "node1", "zone-sun")
	testCases := map[string]struct {
		inputPodsEquivalentGroup    []PodEquivalenceGroup
		expectedPodsEquivalentGroup []PodEquivalenceGroup
	}{
		"single pod": {
			inputPodsEquivalentGroup:    []PodEquivalenceGroup{pg1},
			expectedPodsEquivalentGroup: []PodEquivalenceGroup{pg1},
		},
		"sorted list of pods": {
			inputPodsEquivalentGroup:    []PodEquivalenceGroup{pg3, pg2, pg1},
			expectedPodsEquivalentGroup: []PodEquivalenceGroup{pg3, pg2, pg1},
		},
		"randomised list of pods": {
			inputPodsEquivalentGroup:    []PodEquivalenceGroup{pg1, pg3, pg2},
			expectedPodsEquivalentGroup: []PodEquivalenceGroup{pg3, pg2, pg1},
		},
		"empty pod list": {
			inputPodsEquivalentGroup:    []PodEquivalenceGroup{},
			expectedPodsEquivalentGroup: []PodEquivalenceGroup{},
		},
	}

	for tn, tc := range testCases {
		t.Run(tn, func(t *testing.T) {
			tc := tc
			t.Parallel()
			processor := NewDecreasingPodOrderer()
			nodeInfo := framework.NewNodeInfo(node, nil)
			actual := processor.Order(tc.inputPodsEquivalentGroup, nodeInfo, nil)
			assert.Equal(t, tc.expectedPodsEquivalentGroup, actual)
		})
	}
}
