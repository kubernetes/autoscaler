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
	"time"

	"k8s.io/contrib/cluster-autoscaler/simulator"
	. "k8s.io/contrib/cluster-autoscaler/utils/test"

	kube_api "k8s.io/kubernetes/pkg/api"

	"github.com/stretchr/testify/assert"
)

func TestFindUnneededNodes(t *testing.T) {
	p1 := BuildTestPod("p1", 100, 0)
	p1.Spec.NodeName = "n1"

	p2 := BuildTestPod("p2", 300, 0)
	p2.Spec.NodeName = "n2"
	p2.Annotations = map[string]string{
		"kubernetes.io/created-by": "{\"kind\":\"SerializedReference\",\"apiVersion\":\"v1\",\"reference\":{\"kind\":\"ReplicaSet\"}}",
	}

	p3 := BuildTestPod("p3", 400, 0)
	p3.Annotations = map[string]string{
		"kubernetes.io/created-by": "{\"kind\":\"SerializedReference\",\"apiVersion\":\"v1\",\"reference\":{\"kind\":\"ReplicaSet\"}}",
	}
	p3.Spec.NodeName = "n3"

	p4 := BuildTestPod("p4", 2000, 0)
	p4.Annotations = map[string]string{
		"kubernetes.io/created-by": "{\"kind\":\"SerializedReference\",\"apiVersion\":\"v1\",\"reference\":{\"kind\":\"ReplicaSet\"}}",
	}
	p4.Spec.NodeName = "n4"

	n1 := BuildTestNode("n1", 1000, 10)
	n2 := BuildTestNode("n2", 1000, 10)
	n3 := BuildTestNode("n3", 1000, 10)
	n4 := BuildTestNode("n4", 10000, 10)

	result := FindUnneededNodes([]*kube_api.Node{n1, n2, n3, n4}, map[string]time.Time{}, 0.35,
		[]*kube_api.Pod{p1, p2, p3, p4}, simulator.NewTestPredicateChecker())

	assert.Equal(t, 1, len(result))
	addTime, found := result["n2"]
	assert.True(t, found)

	result["n1"] = time.Now()
	result2 := FindUnneededNodes([]*kube_api.Node{n1, n2, n3, n4}, result, 0.35,
		[]*kube_api.Pod{p1, p2, p3, p4}, simulator.NewTestPredicateChecker())

	assert.Equal(t, 1, len(result2))
	addTime2, found := result2["n2"]
	assert.True(t, found)
	assert.Equal(t, addTime, addTime2)
}
