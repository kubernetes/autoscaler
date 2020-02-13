/*
Copyright 2016 The Kubernetes Authors.

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
	"strings"
	"testing"
	"time"

	apiv1 "k8s.io/api/core/v1"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
	"k8s.io/kubernetes/pkg/scheduler/algorithm/predicates"
	schedulernodeinfo "k8s.io/kubernetes/pkg/scheduler/nodeinfo"

	"github.com/stretchr/testify/assert"
)

func TestPredicates(t *testing.T) {
	p1 := BuildTestPod("p1", 450, 500000)
	p2 := BuildTestPod("p2", 600, 500000)
	p3 := BuildTestPod("p3", 8000, 0)
	p4 := BuildTestPod("p4", 500, 500000)

	ni1 := schedulernodeinfo.NewNodeInfo(p1)
	ni2 := schedulernodeinfo.NewNodeInfo()
	nodeInfos := map[string]*schedulernodeinfo.NodeInfo{
		"n1": ni1,
		"n2": ni2,
	}
	node1 := BuildTestNode("n1", 1000, 2000000)
	node2 := BuildTestNode("n2", 1000, 2000000)
	SetNodeReadyState(node1, true, time.Time{})
	SetNodeReadyState(node2, true, time.Time{})

	ni1.SetNode(node1)
	ni2.SetNode(node2)

	predicateChecker := NewTestPredicateChecker()

	r1, err := predicateChecker.FitsAny(p2, nodeInfos)
	assert.NoError(t, err)
	assert.Equal(t, "n2", r1)

	_, err = predicateChecker.FitsAny(p3, nodeInfos)
	assert.Error(t, err)

	predicateErr := predicateChecker.CheckPredicates(p2, nil, ni1)
	assert.NotNil(t, predicateErr)
	assert.True(t, strings.Contains(predicateErr.Error(), "Predicates failed"))
	assert.True(t, strings.Contains(predicateErr.VerboseError(), "Insufficient cpu"))

	assert.NotNil(t, predicateChecker.CheckPredicates(p2, nil, ni1))
	assert.Nil(t, predicateChecker.CheckPredicates(p4, nil, ni1))
	assert.Nil(t, predicateChecker.CheckPredicates(p2, nil, ni2))
	assert.Nil(t, predicateChecker.CheckPredicates(p4, nil, ni2))
	assert.NotNil(t, predicateChecker.CheckPredicates(p3, nil, ni2))
}

func TestDebugInfo(t *testing.T) {
	p1 := BuildTestPod("p1", 0, 0)

	ni1 := schedulernodeinfo.NewNodeInfo()
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
	ni1.SetNode(node1)

	predicateChecker := NewTestPredicateChecker()
	predicateChecker.predicates = append(predicateChecker.predicates, PredicateInfo{
		Name:      "PodToleratesNodeTaints",
		Predicate: predicates.PodToleratesNodeTaints,
	})

	predicateErr := predicateChecker.CheckPredicates(p1, nil, ni1)
	assert.NotNil(t, predicateErr)
	assert.True(t, strings.Contains(predicateErr.Error(), "Predicates failed"))
	assert.True(t, strings.Contains(predicateErr.VerboseError(), "RandomTaint"), "got: %v, want: %v", predicateErr.VerboseError(), "RandomTaint")
}
