/*
Copyright 2020 The Kubernetes Authors.

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

package simulator

import (
	"testing"
	"time"

	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"

	"github.com/stretchr/testify/assert"

	apiv1 "k8s.io/api/core/v1"
)

func TestCheckPredicate(t *testing.T) {
	p450 := BuildTestPod("p450", 450, 500000)
	p600 := BuildTestPod("p600", 600, 500000)
	p8000 := BuildTestPod("p8000", 8000, 0)
	p500 := BuildTestPod("p500", 500, 500000)

	n1000 := BuildTestNode("n1000", 1000, 2000000)
	SetNodeReadyState(n1000, true, time.Time{})

	tests := []struct {
		name          string
		node          *apiv1.Node
		scheduledPods []*apiv1.Pod
		testPod       *apiv1.Pod
		expectError   bool
	}{
		{
			name:          "other pod - insuficient cpu",
			node:          n1000,
			scheduledPods: []*apiv1.Pod{p450},
			testPod:       p600,
			expectError:   true,
		},
		{
			name:          "other pod - ok",
			node:          n1000,
			scheduledPods: []*apiv1.Pod{p450},
			testPod:       p500,
			expectError:   false,
		},
		{
			name:          "empty - insuficient cpu",
			node:          n1000,
			scheduledPods: []*apiv1.Pod{},
			testPod:       p8000,
			expectError:   true,
		},
		{
			name:          "empty - ok",
			node:          n1000,
			scheduledPods: []*apiv1.Pod{},
			testPod:       p600,
			expectError:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			predicateChecker, err := NewTestPredicateChecker()
			clusterSnapshot := NewBasicClusterSnapshot()
			err = clusterSnapshot.AddNodeWithPods(tt.node, tt.scheduledPods)
			assert.NoError(t, err)

			predicateError := predicateChecker.CheckPredicates(clusterSnapshot, tt.testPod, tt.node.Name)
			if tt.expectError {
				assert.NotNil(t, predicateError)
				assert.Equal(t, NotSchedulablePredicateError, predicateError.ErrorType())
				assert.Equal(t, "Insufficient cpu", predicateError.Message())
				assert.Contains(t, predicateError.VerboseMessage(), "Insufficient cpu; predicateName=NodeResourcesFit")
			} else {
				assert.Nil(t, predicateError)
			}
		})
	}
}

func TestFitsAnyNode(t *testing.T) {
	p900 := BuildTestPod("p900", 900, 1000)
	p1900 := BuildTestPod("p1900", 1900, 1000)
	p2100 := BuildTestPod("p2100", 2100, 1000)

	n1000 := BuildTestNode("n1000", 1000, 2000000)
	n2000 := BuildTestNode("n2000", 2000, 2000000)

	var err error

	clusterSnapshot := NewBasicClusterSnapshot()
	err = clusterSnapshot.AddNode(n1000)
	assert.NoError(t, err)
	err = clusterSnapshot.AddNode(n2000)
	assert.NoError(t, err)

	predicateChecker, err := NewTestPredicateChecker()
	assert.NoError(t, err)

	nodeName, err := predicateChecker.FitsAnyNode(clusterSnapshot, p900)
	assert.NoError(t, err)
	assert.True(t, nodeName == "n1000" || nodeName == "n2000")

	nodeName, err = predicateChecker.FitsAnyNode(clusterSnapshot, p1900)
	assert.NoError(t, err)
	assert.Equal(t, "n2000", nodeName)

	nodeName, err = predicateChecker.FitsAnyNode(clusterSnapshot, p2100)
	assert.Error(t, err)
}

func TestDebugInfo(t *testing.T) {
	p1 := BuildTestPod("p1", 0, 0)
	node1 := BuildTestNode("n1", 1000, 2000000)
	node1.Spec.Taints = []apiv1.Taint{
		{
			Key:    "SomeTaint",
			Value:  "WhyNot?",
			Effect: apiv1.TaintEffectNoSchedule,
		},
		{
			Key:    "RandomTaint",
			Value:  "JustBecause",
			Effect: apiv1.TaintEffectNoExecute,
		},
	}
	SetNodeReadyState(node1, true, time.Time{})

	predicateChecker, err := NewTestPredicateChecker()
	assert.NoError(t, err)

	clusterSnapshot := NewBasicClusterSnapshot()

	err = clusterSnapshot.AddNode(node1)
	assert.NoError(t, err)

	predicateErr := predicateChecker.CheckPredicates(clusterSnapshot, p1, "n1")
	assert.NotNil(t, predicateErr)
	assert.Equal(t, "node(s) had taint {SomeTaint: WhyNot?}, that the pod didn't tolerate", predicateErr.Message())
	assert.Contains(t, predicateErr.VerboseMessage(), "RandomTaint")
}
