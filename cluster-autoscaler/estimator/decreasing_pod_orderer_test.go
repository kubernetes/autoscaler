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
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/utils/test"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
)

func TestPodPriorityProcessor(t *testing.T) {
	p1 := test.BuildTestPod("p1", 1, 1)
	p2 := test.BuildTestPod("p2", 2, 1)
	p3 := test.BuildTestPod("p3", 2, 100)
	node := makeNode(4, 600, "node1", "zone-sun")
	testCases := map[string]struct {
		inputPods    []*apiv1.Pod
		expectedPods []*apiv1.Pod
	}{
		"single pod": {
			inputPods:    []*apiv1.Pod{p1},
			expectedPods: []*apiv1.Pod{p1},
		},
		"sorted list of pods": {
			inputPods:    []*apiv1.Pod{p3, p2, p1},
			expectedPods: []*apiv1.Pod{p3, p2, p1},
		},
		"randomised list of pods": {
			inputPods:    []*apiv1.Pod{p1, p3, p2},
			expectedPods: []*apiv1.Pod{p3, p2, p1},
		},
		"empty pod list": {
			inputPods:    []*apiv1.Pod{},
			expectedPods: []*apiv1.Pod{},
		},
	}

	for tn, tc := range testCases {
		t.Run(tn, func(t *testing.T) {
			tc := tc
			t.Parallel()
			processor := NewDecreasingPodOrderer()
			nodeInfo := schedulerframework.NewNodeInfo()
			nodeInfo.SetNode(node)
			actual := processor.Order(tc.inputPods, nodeInfo, nil)
			assert.Equal(t, tc.expectedPods, actual)
		})
	}
}
