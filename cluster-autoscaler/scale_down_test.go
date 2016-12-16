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

package main

import (
	"testing"
	"time"

	"k8s.io/contrib/cluster-autoscaler/simulator"
	. "k8s.io/contrib/cluster-autoscaler/utils/test"

	"k8s.io/kubernetes/pkg/api/errors"
	apiv1 "k8s.io/kubernetes/pkg/api/v1"
	"k8s.io/kubernetes/pkg/client/clientset_generated/release_1_5/fake"
	"k8s.io/kubernetes/pkg/client/testing/core"
	"k8s.io/kubernetes/pkg/runtime"

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

	context := AutoscalingContext{
		PredicateChecker:              simulator.NewTestPredicateChecker(),
		ScaleDownUtilizationThreshold: 0.35,
	}

	result, hints, utilization := FindUnneededNodes(context, []*apiv1.Node{n1, n2, n3, n4}, map[string]time.Time{},
		[]*apiv1.Pod{p1, p2, p3, p4}, make(map[string]string),
		simulator.NewUsageTracker(), time.Now())

	assert.Equal(t, 1, len(result))
	addTime, found := result["n2"]
	assert.True(t, found)
	assert.Contains(t, hints, p2.Namespace+"/"+p2.Name)
	assert.Equal(t, 4, len(utilization))

	result["n1"] = time.Now()
	result2, hints, utilization := FindUnneededNodes(context, []*apiv1.Node{n1, n2, n3, n4}, result,
		[]*apiv1.Pod{p1, p2, p3, p4}, hints,
		simulator.NewUsageTracker(), time.Now())

	assert.Equal(t, 1, len(result2))
	addTime2, found := result2["n2"]
	assert.True(t, found)
	assert.Equal(t, addTime, addTime2)
	assert.Equal(t, 4, len(utilization))
}

func TestDrainNode(t *testing.T) {
	deletedPods := make(chan string, 10)
	updatedNodes := make(chan string, 10)
	fakeClient := &fake.Clientset{}

	p1 := BuildTestPod("p1", 100, 0)
	p2 := BuildTestPod("p2", 300, 0)
	n1 := BuildTestNode("n1", 1000, 1000)

	fakeClient.Fake.AddReactor("list", "pods", func(action core.Action) (bool, runtime.Object, error) {
		return true, &apiv1.PodList{Items: []apiv1.Pod{*p1, *p2}}, nil
	})
	fakeClient.Fake.AddReactor("get", "pods", func(action core.Action) (bool, runtime.Object, error) {
		return true, nil, errors.NewNotFound(apiv1.Resource("pod"), "whatever")
	})
	fakeClient.Fake.AddReactor("get", "nodes", func(action core.Action) (bool, runtime.Object, error) {
		return true, n1, nil
	})
	fakeClient.Fake.AddReactor("delete", "pods", func(action core.Action) (bool, runtime.Object, error) {
		deleteAction := action.(core.DeleteAction)
		deletedPods <- deleteAction.GetName()
		return true, nil, nil
	})
	fakeClient.Fake.AddReactor("update", "nodes", func(action core.Action) (bool, runtime.Object, error) {
		update := action.(core.UpdateAction)
		obj := update.GetObject().(*apiv1.Node)
		updatedNodes <- obj.Name
		return true, obj, nil
	})
	err := drainNode(n1, []*apiv1.Pod{p1, p2}, fakeClient, createEventRecorder(fakeClient), 20)
	assert.NoError(t, err)
	assert.Equal(t, p1.Name, getStringFromChan(deletedPods))
	assert.Equal(t, p2.Name, getStringFromChan(deletedPods))
	assert.Equal(t, n1.Name, getStringFromChan(updatedNodes))
}

func TestCleanNodes(t *testing.T) {
	updatedNodes := make(chan string, 10)
	fakeClient := &fake.Clientset{}

	n1 := BuildTestNode("n1", 1000, 1000)
	addToBeDeletedTaint(n1)
	n2 := BuildTestNode("n2", 1000, 1000)

	fakeClient.Fake.AddReactor("get", "nodes", func(action core.Action) (bool, runtime.Object, error) {
		get := action.(core.GetAction)
		if get.GetName() == n1.Name {
			return true, n1, nil
		}
		if get.GetName() == n2.Name {
			return true, n2, nil
		}
		return true, nil, errors.NewNotFound(apiv1.Resource("node"), get.GetName())
	})
	fakeClient.Fake.AddReactor("update", "nodes", func(action core.Action) (bool, runtime.Object, error) {
		update := action.(core.UpdateAction)
		obj := update.GetObject().(*apiv1.Node)
		updatedNodes <- obj.Name
		return true, obj, nil
	})
	err := cleanToBeDeleted([]*apiv1.Node{n1, n2}, fakeClient, createEventRecorder(fakeClient))
	assert.NoError(t, err)
	assert.Equal(t, n1.Name, getStringFromChan(updatedNodes))
}

func getStringFromChan(c chan string) string {
	select {
	case val := <-c:
		return val
	case <-time.After(time.Second * 10):
		return "Nothing returned"
	}
}
