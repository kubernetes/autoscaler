/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

package main

import (
	"testing"

	"k8s.io/contrib/cluster-autoscaler/simulator"
	. "k8s.io/contrib/cluster-autoscaler/utils/test"

	kube_api "k8s.io/kubernetes/pkg/api"

	"github.com/stretchr/testify/assert"
)

func TestFilterOutSchedulable(t *testing.T) {
	p1 := BuildTestPod("p1", 1500, 200000)
	p2 := BuildTestPod("p2", 3000, 200000)
	p3 := BuildTestPod("p3", 100, 200000)
	unschedulablePods := []*kube_api.Pod{p1, p2, p3}

	scheduledPod1 := BuildTestPod("s1", 100, 200000)
	scheduledPod2 := BuildTestPod("s2", 1500, 200000)
	scheduledPod1.Spec.NodeName = "node1"
	scheduledPod2.Spec.NodeName = "node1"

	node := BuildTestNode("node1", 2000, 2000000)

	predicateChecker := simulator.NewTestPredicateChecker()

	res := FilterOutSchedulable(unschedulablePods, []*kube_api.Node{node}, []*kube_api.Pod{scheduledPod1}, predicateChecker)
	assert.Equal(t, 1, len(res))
	assert.Equal(t, p2, res[0])

	res2 := FilterOutSchedulable(unschedulablePods, []*kube_api.Node{node}, []*kube_api.Pod{scheduledPod1, scheduledPod2}, predicateChecker)
	assert.Equal(t, 2, len(res2))
	assert.Equal(t, p1, res2[0])
	assert.Equal(t, p2, res2[1])
}
