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

package core

import (
	"fmt"
	"sort"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate/utils"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
	core "k8s.io/client-go/testing"
	apiv1 "k8s.io/kubernetes/pkg/api/v1"
	batchv1 "k8s.io/kubernetes/pkg/apis/batch/v1"
	policyv1 "k8s.io/kubernetes/pkg/apis/policy/v1beta1"
	"k8s.io/kubernetes/pkg/client/clientset_generated/clientset/fake"

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

	SetNodeReadyState(n1, true, time.Time{})
	SetNodeReadyState(n2, true, time.Time{})
	SetNodeReadyState(n3, true, time.Time{})
	SetNodeReadyState(n4, true, time.Time{})

	fakeClient := &fake.Clientset{}
	fakeRecorder := kube_util.CreateEventRecorder(fakeClient)
	fakeLogRecorder, _ := utils.NewStatusMapRecorder(fakeClient, fakeRecorder, false)

	provider := testprovider.NewTestCloudProvider(nil, nil)
	provider.AddNodeGroup("ng1", 1, 10, 2)
	provider.AddNode("ng1", n1)
	provider.AddNode("ng1", n2)
	provider.AddNode("ng1", n3)
	provider.AddNode("ng1", n4)

	context := AutoscalingContext{
		AutoscalingOptions: AutoscalingOptions{
			ScaleDownUtilizationThreshold: 0.35,
		},
		ClusterStateRegistry: clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{}),
		PredicateChecker:     simulator.NewTestPredicateChecker(),
		LogRecorder:          fakeLogRecorder,
	}
	sd := NewScaleDown(&context)
	sd.UpdateUnneededNodes([]*apiv1.Node{n1, n2, n3, n4}, []*apiv1.Pod{p1, p2, p3, p4}, time.Now(), nil)

	assert.Equal(t, 1, len(sd.unneededNodes))
	addTime, found := sd.unneededNodes["n2"]
	assert.True(t, found)
	assert.Contains(t, sd.podLocationHints, p2.Namespace+"/"+p2.Name)
	assert.Equal(t, 4, len(sd.nodeUtilizationMap))

	sd.unneededNodes["n1"] = time.Now()
	sd.UpdateUnneededNodes([]*apiv1.Node{n1, n2, n3, n4}, []*apiv1.Pod{p1, p2, p3, p4}, time.Now(), nil)

	assert.Equal(t, 1, len(sd.unneededNodes))
	addTime2, found := sd.unneededNodes["n2"]
	assert.True(t, found)
	assert.Equal(t, addTime, addTime2)
	assert.Equal(t, 4, len(sd.nodeUtilizationMap))
}

func TestDrainNode(t *testing.T) {
	deletedPods := make(chan string, 10)
	updatedNodes := make(chan string, 10)
	fakeClient := &fake.Clientset{}

	p1 := BuildTestPod("p1", 100, 0)
	p2 := BuildTestPod("p2", 300, 0)
	n1 := BuildTestNode("n1", 1000, 1000)
	SetNodeReadyState(n1, true, time.Time{})

	fakeClient.Fake.AddReactor("get", "pods", func(action core.Action) (bool, runtime.Object, error) {
		return true, nil, errors.NewNotFound(apiv1.Resource("pod"), "whatever")
	})
	fakeClient.Fake.AddReactor("get", "nodes", func(action core.Action) (bool, runtime.Object, error) {
		return true, n1, nil
	})
	fakeClient.Fake.AddReactor("create", "pods", func(action core.Action) (bool, runtime.Object, error) {
		createAction := action.(core.CreateAction)
		if createAction == nil {
			return false, nil, nil
		}
		eviction := createAction.GetObject().(*policyv1.Eviction)
		if eviction == nil {
			return false, nil, nil
		}
		deletedPods <- eviction.Name
		return true, nil, nil
	})
	fakeClient.Fake.AddReactor("update", "nodes", func(action core.Action) (bool, runtime.Object, error) {
		update := action.(core.UpdateAction)
		obj := update.GetObject().(*apiv1.Node)
		updatedNodes <- obj.Name
		return true, obj, nil
	})
	err := drainNode(n1, []*apiv1.Pod{p1, p2}, fakeClient, kube_util.CreateEventRecorder(fakeClient), 20, 5*time.Second, 0*time.Second)
	assert.NoError(t, err)
	deleted := make([]string, 0)
	deleted = append(deleted, getStringFromChan(deletedPods))
	deleted = append(deleted, getStringFromChan(deletedPods))
	sort.Strings(deleted)
	assert.Equal(t, p1.Name, deleted[0])
	assert.Equal(t, p2.Name, deleted[1])
	assert.Equal(t, n1.Name, getStringFromChan(updatedNodes))
}

func TestDrainNodeWithRetries(t *testing.T) {
	deletedPods := make(chan string, 10)
	updatedNodes := make(chan string, 10)
	// Simulate pdb of size 1, by making them goroutine succeed sequentially
	// and fail/retry before they can proceed.
	ticket := make(chan bool, 1)
	fakeClient := &fake.Clientset{}

	p1 := BuildTestPod("p1", 100, 0)
	p2 := BuildTestPod("p2", 300, 0)
	p3 := BuildTestPod("p3", 300, 0)
	n1 := BuildTestNode("n1", 1000, 1000)
	SetNodeReadyState(n1, true, time.Time{})

	fakeClient.Fake.AddReactor("get", "pods", func(action core.Action) (bool, runtime.Object, error) {
		return true, nil, errors.NewNotFound(apiv1.Resource("pod"), "whatever")
	})
	fakeClient.Fake.AddReactor("get", "nodes", func(action core.Action) (bool, runtime.Object, error) {
		return true, n1, nil
	})
	fakeClient.Fake.AddReactor("create", "pods", func(action core.Action) (bool, runtime.Object, error) {
		createAction := action.(core.CreateAction)
		if createAction == nil {
			return false, nil, nil
		}
		eviction := createAction.GetObject().(*policyv1.Eviction)
		if eviction == nil {
			return false, nil, nil
		}
		select {
		case <-ticket:
			deletedPods <- eviction.Name
			return true, nil, nil
		default:
			select {
			case ticket <- true:
			default:
			}
			return true, nil, fmt.Errorf("Too many concurrent evictions")
		}
	})
	fakeClient.Fake.AddReactor("update", "nodes", func(action core.Action) (bool, runtime.Object, error) {
		update := action.(core.UpdateAction)
		obj := update.GetObject().(*apiv1.Node)
		updatedNodes <- obj.Name
		return true, obj, nil
	})
	err := drainNode(n1, []*apiv1.Pod{p1, p2, p3}, fakeClient, kube_util.CreateEventRecorder(fakeClient), 20, 5*time.Second, 0*time.Second)
	assert.NoError(t, err)
	deleted := make([]string, 0)
	deleted = append(deleted, getStringFromChan(deletedPods))
	deleted = append(deleted, getStringFromChan(deletedPods))
	deleted = append(deleted, getStringFromChan(deletedPods))
	sort.Strings(deleted)
	assert.Equal(t, p1.Name, deleted[0])
	assert.Equal(t, p2.Name, deleted[1])
	assert.Equal(t, p3.Name, deleted[2])
	assert.Equal(t, n1.Name, getStringFromChan(updatedNodes))
}

func TestScaleDown(t *testing.T) {
	deletedPods := make(chan string, 10)
	updatedNodes := make(chan string, 10)
	deletedNodes := make(chan string, 10)
	fakeClient := &fake.Clientset{}

	job := batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "job",
			Namespace: "default",
			SelfLink:  "/apivs/extensions/v1beta1/namespaces/default/jobs/job",
		},
	}
	n1 := BuildTestNode("n1", 1000, 1000)
	SetNodeReadyState(n1, true, time.Time{})
	n2 := BuildTestNode("n2", 1000, 1000)
	SetNodeReadyState(n2, true, time.Time{})
	p1 := BuildTestPod("p1", 100, 0)
	p1.Annotations = map[string]string{
		"kubernetes.io/created-by": RefJSON(&job),
	}

	p2 := BuildTestPod("p2", 800, 0)
	p1.Spec.NodeName = "n1"
	p2.Spec.NodeName = "n2"

	fakeClient.Fake.AddReactor("list", "pods", func(action core.Action) (bool, runtime.Object, error) {
		return true, &apiv1.PodList{Items: []apiv1.Pod{*p1, *p2}}, nil
	})
	fakeClient.Fake.AddReactor("get", "pods", func(action core.Action) (bool, runtime.Object, error) {
		return true, nil, errors.NewNotFound(apiv1.Resource("pod"), "whatever")
	})
	fakeClient.Fake.AddReactor("get", "nodes", func(action core.Action) (bool, runtime.Object, error) {
		getAction := action.(core.GetAction)
		switch getAction.GetName() {
		case n1.Name:
			return true, n1, nil
		case n2.Name:
			return true, n2, nil
		}
		return true, nil, fmt.Errorf("Wrong node: %v", getAction.GetName())
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

	provider := testprovider.NewTestCloudProvider(nil, func(nodeGroup string, node string) error {
		deletedNodes <- node
		return nil
	})
	provider.AddNodeGroup("ng1", 1, 10, 2)
	provider.AddNode("ng1", n1)
	provider.AddNode("ng1", n2)
	assert.NotNil(t, provider)

	fakeRecorder := kube_util.CreateEventRecorder(fakeClient)
	fakeLogRecorder, _ := utils.NewStatusMapRecorder(fakeClient, fakeRecorder, false)
	context := &AutoscalingContext{
		AutoscalingOptions: AutoscalingOptions{
			ScaleDownUtilizationThreshold: 0.5,
			ScaleDownUnneededTime:         time.Minute,
			MaxGratefulTerminationSec:     60,
		},
		PredicateChecker:     simulator.NewTestPredicateChecker(),
		CloudProvider:        provider,
		ClientSet:            fakeClient,
		Recorder:             fakeRecorder,
		ClusterStateRegistry: clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{}),
		LogRecorder:          fakeLogRecorder,
	}
	scaleDown := NewScaleDown(context)
	scaleDown.UpdateUnneededNodes([]*apiv1.Node{n1, n2}, []*apiv1.Pod{p1, p2}, time.Now().Add(-5*time.Minute), nil)
	result, err := scaleDown.TryToScaleDown([]*apiv1.Node{n1, n2}, []*apiv1.Pod{p1, p2}, nil)
	assert.NoError(t, err)
	assert.Equal(t, ScaleDownNodeDeleted, result)
	assert.Equal(t, n1.Name, getStringFromChan(deletedNodes))
	assert.Equal(t, n1.Name, getStringFromChan(updatedNodes))
}

func TestNoScaleDownUnready(t *testing.T) {
	fakeClient := &fake.Clientset{}
	n1 := BuildTestNode("n1", 1000, 1000)
	SetNodeReadyState(n1, false, time.Now().Add(-3*time.Minute))
	n2 := BuildTestNode("n2", 1000, 1000)
	SetNodeReadyState(n2, true, time.Time{})
	p2 := BuildTestPod("p2", 800, 0)
	p2.Spec.NodeName = "n2"

	fakeClient.Fake.AddReactor("list", "pods", func(action core.Action) (bool, runtime.Object, error) {
		return true, &apiv1.PodList{Items: []apiv1.Pod{*p2}}, nil
	})
	fakeClient.Fake.AddReactor("get", "pods", func(action core.Action) (bool, runtime.Object, error) {
		return true, nil, errors.NewNotFound(apiv1.Resource("pod"), "whatever")
	})
	fakeClient.Fake.AddReactor("get", "nodes", func(action core.Action) (bool, runtime.Object, error) {
		getAction := action.(core.GetAction)
		switch getAction.GetName() {
		case n1.Name:
			return true, n1, nil
		case n2.Name:
			return true, n2, nil
		}
		return true, nil, fmt.Errorf("Wrong node: %v", getAction.GetName())
	})

	provider := testprovider.NewTestCloudProvider(nil, func(nodeGroup string, node string) error {
		t.Fatalf("Unexpected deletion of %s", node)
		return nil
	})
	provider.AddNodeGroup("ng1", 1, 10, 2)
	provider.AddNode("ng1", n1)
	provider.AddNode("ng1", n2)

	fakeRecorder := kube_util.CreateEventRecorder(fakeClient)
	fakeLogRecorder, _ := utils.NewStatusMapRecorder(fakeClient, fakeRecorder, false)
	context := &AutoscalingContext{
		AutoscalingOptions: AutoscalingOptions{
			ScaleDownUtilizationThreshold: 0.5,
			ScaleDownUnneededTime:         time.Minute,
			ScaleDownUnreadyTime:          time.Hour,
			MaxGratefulTerminationSec:     60,
		},
		PredicateChecker:     simulator.NewTestPredicateChecker(),
		CloudProvider:        provider,
		ClientSet:            fakeClient,
		Recorder:             fakeRecorder,
		ClusterStateRegistry: clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{}),
		LogRecorder:          fakeLogRecorder,
	}

	// N1 is unready so it requires a bigger unneeded time.
	scaleDown := NewScaleDown(context)
	scaleDown.UpdateUnneededNodes([]*apiv1.Node{n1, n2}, []*apiv1.Pod{p2}, time.Now().Add(-5*time.Minute), nil)
	result, err := scaleDown.TryToScaleDown([]*apiv1.Node{n1, n2}, []*apiv1.Pod{p2}, nil)
	assert.NoError(t, err)
	assert.Equal(t, ScaleDownNoUnneeded, result)

	deletedNodes := make(chan string, 10)

	provider = testprovider.NewTestCloudProvider(nil, func(nodeGroup string, node string) error {
		deletedNodes <- node
		return nil
	})
	SetNodeReadyState(n1, false, time.Now().Add(-3*time.Hour))
	provider.AddNodeGroup("ng1", 1, 10, 2)
	provider.AddNode("ng1", n1)
	provider.AddNode("ng1", n2)

	// N1 has been unready for 2 hours, ok to delete.
	context.CloudProvider = provider
	scaleDown = NewScaleDown(context)
	scaleDown.UpdateUnneededNodes([]*apiv1.Node{n1, n2}, []*apiv1.Pod{p2}, time.Now().Add(-2*time.Hour), nil)
	result, err = scaleDown.TryToScaleDown([]*apiv1.Node{n1, n2}, []*apiv1.Pod{p2}, nil)
	assert.NoError(t, err)
	assert.Equal(t, ScaleDownNodeDeleted, result)
	assert.Equal(t, n1.Name, getStringFromChan(deletedNodes))
}

func TestScaleDownNoMove(t *testing.T) {
	fakeClient := &fake.Clientset{}

	job := batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "job",
			Namespace: "default",
			SelfLink:  "/apivs/extensions/v1beta1/namespaces/default/jobs/job",
		},
	}
	n1 := BuildTestNode("n1", 1000, 1000)
	SetNodeReadyState(n1, true, time.Time{})

	// N2 is unready so no pods can be moved there.
	n2 := BuildTestNode("n2", 1000, 1000)
	SetNodeReadyState(n2, false, time.Time{})

	p1 := BuildTestPod("p1", 100, 0)
	p1.Annotations = map[string]string{
		"kubernetes.io/created-by": RefJSON(&job),
	}

	p2 := BuildTestPod("p2", 800, 0)
	p1.Spec.NodeName = "n1"
	p2.Spec.NodeName = "n2"

	fakeClient.Fake.AddReactor("list", "pods", func(action core.Action) (bool, runtime.Object, error) {
		return true, &apiv1.PodList{Items: []apiv1.Pod{*p1, *p2}}, nil
	})
	fakeClient.Fake.AddReactor("get", "pods", func(action core.Action) (bool, runtime.Object, error) {
		return true, nil, errors.NewNotFound(apiv1.Resource("pod"), "whatever")
	})
	fakeClient.Fake.AddReactor("get", "nodes", func(action core.Action) (bool, runtime.Object, error) {
		getAction := action.(core.GetAction)
		switch getAction.GetName() {
		case n1.Name:
			return true, n1, nil
		case n2.Name:
			return true, n2, nil
		}
		return true, nil, fmt.Errorf("Wrong node: %v", getAction.GetName())
	})
	fakeClient.Fake.AddReactor("delete", "pods", func(action core.Action) (bool, runtime.Object, error) {
		t.FailNow()
		return false, nil, nil
	})
	fakeClient.Fake.AddReactor("update", "nodes", func(action core.Action) (bool, runtime.Object, error) {
		t.FailNow()
		return false, nil, nil
	})
	provider := testprovider.NewTestCloudProvider(nil, func(nodeGroup string, node string) error {
		t.FailNow()
		return nil
	})
	provider.AddNodeGroup("ng1", 1, 10, 2)
	provider.AddNode("ng1", n1)
	provider.AddNode("ng1", n2)
	assert.NotNil(t, provider)

	fakeRecorder := kube_util.CreateEventRecorder(fakeClient)
	fakeLogRecorder, _ := utils.NewStatusMapRecorder(fakeClient, fakeRecorder, false)
	context := &AutoscalingContext{
		AutoscalingOptions: AutoscalingOptions{
			ScaleDownUtilizationThreshold: 0.5,
			ScaleDownUnneededTime:         time.Minute,
			ScaleDownUnreadyTime:          time.Hour,
			MaxGratefulTerminationSec:     60,
		},
		PredicateChecker:     simulator.NewTestPredicateChecker(),
		CloudProvider:        provider,
		ClientSet:            fakeClient,
		Recorder:             fakeRecorder,
		ClusterStateRegistry: clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{}),
		LogRecorder:          fakeLogRecorder,
	}
	scaleDown := NewScaleDown(context)
	scaleDown.UpdateUnneededNodes([]*apiv1.Node{n1, n2}, []*apiv1.Pod{p1, p2}, time.Now().Add(5*time.Minute), nil)
	result, err := scaleDown.TryToScaleDown([]*apiv1.Node{n1, n2}, []*apiv1.Pod{p1, p2}, nil)
	assert.NoError(t, err)
	assert.Equal(t, ScaleDownNoUnneeded, result)
}

func getStringFromChan(c chan string) string {
	select {
	case val := <-c:
		return val
	case <-time.After(time.Second * 10):
		return "Nothing returned"
	}
}
