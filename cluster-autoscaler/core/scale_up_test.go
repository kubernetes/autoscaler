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
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"
	core "k8s.io/client-go/testing"
	"k8s.io/contrib/cluster-autoscaler/cloudprovider/test"
	"k8s.io/contrib/cluster-autoscaler/clusterstate"
	"k8s.io/contrib/cluster-autoscaler/clusterstate/utils"
	"k8s.io/contrib/cluster-autoscaler/estimator"
	"k8s.io/contrib/cluster-autoscaler/expander/random"
	"k8s.io/contrib/cluster-autoscaler/simulator"
	kube_util "k8s.io/contrib/cluster-autoscaler/utils/kubernetes"
	. "k8s.io/contrib/cluster-autoscaler/utils/test"
	apiv1 "k8s.io/kubernetes/pkg/api/v1"
	"k8s.io/kubernetes/pkg/client/clientset_generated/clientset/fake"
)

func TestScaleUpOK(t *testing.T) {
	expandedGroups := make(chan string, 10)
	fakeClient := &fake.Clientset{}

	n1 := BuildTestNode("n1", 100, 1000)
	SetNodeReadyState(n1, true, time.Now())
	n2 := BuildTestNode("n2", 1000, 1000)
	SetNodeReadyState(n2, true, time.Now())

	p1 := BuildTestPod("p1", 80, 0)
	p2 := BuildTestPod("p2", 800, 0)
	p1.Spec.NodeName = "n1"
	p2.Spec.NodeName = "n2"

	fakeClient.Fake.AddReactor("list", "pods", func(action core.Action) (bool, runtime.Object, error) {
		list := action.(core.ListAction)
		fieldstring := list.GetListRestrictions().Fields.String()
		if strings.Contains(fieldstring, "n1") {
			return true, &apiv1.PodList{Items: []apiv1.Pod{*p1}}, nil
		}
		if strings.Contains(fieldstring, "n2") {
			return true, &apiv1.PodList{Items: []apiv1.Pod{*p2}}, nil
		}
		return true, nil, fmt.Errorf("Failed to list: %v", list)
	})

	provider := testprovider.NewTestCloudProvider(func(nodeGroup string, increase int) error {
		expandedGroups <- fmt.Sprintf("%s-%d", nodeGroup, increase)
		return nil
	}, nil)
	provider.AddNodeGroup("ng1", 1, 10, 1)
	provider.AddNodeGroup("ng2", 1, 10, 1)
	provider.AddNode("ng1", n1)
	provider.AddNode("ng2", n2)
	assert.NotNil(t, provider)

	clusterState := clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{})
	clusterState.UpdateNodes([]*apiv1.Node{n1, n2}, time.Now())
	fakeRecorder := kube_util.CreateEventRecorder(fakeClient)
	fakeLogRecorder, _ := utils.NewStatusMapRecorder(fakeClient, fakeRecorder, false)
	context := &AutoscalingContext{
		AutoscalingOptions: AutoscalingOptions{
			EstimatorName: estimator.BinpackingEstimatorName,
		},
		PredicateChecker:     simulator.NewTestPredicateChecker(),
		CloudProvider:        provider,
		ClientSet:            fakeClient,
		Recorder:             fakeRecorder,
		ExpanderStrategy:     random.NewStrategy(),
		ClusterStateRegistry: clusterState,
		LogRecorder:          fakeLogRecorder,
	}
	p3 := BuildTestPod("p-new", 500, 0)

	result, err := ScaleUp(context, []*apiv1.Pod{p3}, []*apiv1.Node{n1, n2})
	assert.NoError(t, err)
	assert.True(t, result)
	assert.Equal(t, "ng2-1", getStringFromChan(expandedGroups))
}

func TestScaleUpNodeComingNoScale(t *testing.T) {
	n1 := BuildTestNode("n1", 100, 1000)
	SetNodeReadyState(n1, true, time.Now())
	n2 := BuildTestNode("n2", 1000, 1000)
	SetNodeReadyState(n2, true, time.Now())

	p1 := BuildTestPod("p1", 80, 0)
	p2 := BuildTestPod("p2", 800, 0)
	p1.Spec.NodeName = "n1"
	p2.Spec.NodeName = "n2"

	fakeClient := &fake.Clientset{}
	fakeClient.Fake.AddReactor("list", "pods", func(action core.Action) (bool, runtime.Object, error) {
		list := action.(core.ListAction)
		fieldstring := list.GetListRestrictions().Fields.String()
		if strings.Contains(fieldstring, "n1") {
			return true, &apiv1.PodList{Items: []apiv1.Pod{*p1}}, nil
		}
		if strings.Contains(fieldstring, "n2") {
			return true, &apiv1.PodList{Items: []apiv1.Pod{*p2}}, nil
		}
		return true, nil, fmt.Errorf("Failed to list: %v", list)
	})

	provider := testprovider.NewTestCloudProvider(func(nodeGroup string, increase int) error {
		t.Fatalf("No expansion is expected, but increased %s by %d", nodeGroup, increase)
		return nil
	}, nil)
	provider.AddNodeGroup("ng1", 1, 10, 1)
	provider.AddNodeGroup("ng2", 1, 10, 2)
	provider.AddNode("ng1", n1)
	provider.AddNode("ng2", n2)

	clusterState := clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{})
	clusterState.RegisterScaleUp(&clusterstate.ScaleUpRequest{
		NodeGroupName:   "ng2",
		Increase:        1,
		Time:            time.Now(),
		ExpectedAddTime: time.Now().Add(5 * time.Minute),
	})
	clusterState.UpdateNodes([]*apiv1.Node{n1, n2}, time.Now())

	fakeRecorder := kube_util.CreateEventRecorder(fakeClient)
	fakeLogRecorder, _ := utils.NewStatusMapRecorder(fakeClient, fakeRecorder, false)
	context := &AutoscalingContext{
		AutoscalingOptions: AutoscalingOptions{
			EstimatorName: estimator.BinpackingEstimatorName,
		},
		PredicateChecker:     simulator.NewTestPredicateChecker(),
		CloudProvider:        provider,
		ClientSet:            fakeClient,
		Recorder:             fakeRecorder,
		ExpanderStrategy:     random.NewStrategy(),
		ClusterStateRegistry: clusterState,
		LogRecorder:          fakeLogRecorder,
	}
	p3 := BuildTestPod("p-new", 550, 0)

	result, err := ScaleUp(context, []*apiv1.Pod{p3}, []*apiv1.Node{n1, n2})
	assert.NoError(t, err)
	// A node is already coming - no need for scale up.
	assert.False(t, result)
}

func TestScaleUpNodeComingHasScale(t *testing.T) {
	n1 := BuildTestNode("n1", 100, 1000)
	SetNodeReadyState(n1, true, time.Now())
	n2 := BuildTestNode("n2", 1000, 1000)
	SetNodeReadyState(n2, true, time.Now())

	p1 := BuildTestPod("p1", 80, 0)
	p2 := BuildTestPod("p2", 800, 0)
	p1.Spec.NodeName = "n1"
	p2.Spec.NodeName = "n2"

	fakeClient := &fake.Clientset{}
	fakeClient.Fake.AddReactor("list", "pods", func(action core.Action) (bool, runtime.Object, error) {
		list := action.(core.ListAction)
		fieldstring := list.GetListRestrictions().Fields.String()
		if strings.Contains(fieldstring, "n1") {
			return true, &apiv1.PodList{Items: []apiv1.Pod{*p1}}, nil
		}
		if strings.Contains(fieldstring, "n2") {
			return true, &apiv1.PodList{Items: []apiv1.Pod{*p2}}, nil
		}
		return true, nil, fmt.Errorf("Failed to list: %v", list)
	})

	expandedGroups := make(chan string, 10)
	provider := testprovider.NewTestCloudProvider(func(nodeGroup string, increase int) error {
		expandedGroups <- fmt.Sprintf("%s-%d", nodeGroup, increase)
		return nil
	}, nil)
	provider.AddNodeGroup("ng1", 1, 10, 1)
	provider.AddNodeGroup("ng2", 1, 10, 2)
	provider.AddNode("ng1", n1)
	provider.AddNode("ng2", n2)

	clusterState := clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{})
	clusterState.RegisterScaleUp(&clusterstate.ScaleUpRequest{
		NodeGroupName:   "ng2",
		Increase:        1,
		Time:            time.Now(),
		ExpectedAddTime: time.Now().Add(5 * time.Minute),
	})
	clusterState.UpdateNodes([]*apiv1.Node{n1, n2}, time.Now())

	fakeRecorder := kube_util.CreateEventRecorder(fakeClient)
	fakeLogRecorder, _ := utils.NewStatusMapRecorder(fakeClient, fakeRecorder, false)
	context := &AutoscalingContext{
		AutoscalingOptions: AutoscalingOptions{
			EstimatorName: estimator.BinpackingEstimatorName,
		},
		PredicateChecker:     simulator.NewTestPredicateChecker(),
		CloudProvider:        provider,
		ClientSet:            fakeClient,
		Recorder:             fakeRecorder,
		ExpanderStrategy:     random.NewStrategy(),
		ClusterStateRegistry: clusterState,
		LogRecorder:          fakeLogRecorder,
	}
	p3 := BuildTestPod("p-new", 550, 0)

	result, err := ScaleUp(context, []*apiv1.Pod{p3, p3}, []*apiv1.Node{n1, n2})
	assert.NoError(t, err)
	// Twho nodes needed but one node is already coming, so it should increase by one.
	assert.True(t, result)
	assert.Equal(t, "ng2-1", getStringFromChan(expandedGroups))
}

func TestScaleUpUnhealthy(t *testing.T) {
	n1 := BuildTestNode("n1", 100, 1000)
	SetNodeReadyState(n1, true, time.Now())
	n2 := BuildTestNode("n2", 1000, 1000)
	SetNodeReadyState(n2, true, time.Now())

	p1 := BuildTestPod("p1", 80, 0)
	p2 := BuildTestPod("p2", 800, 0)
	p1.Spec.NodeName = "n1"
	p2.Spec.NodeName = "n2"

	fakeClient := &fake.Clientset{}
	fakeClient.Fake.AddReactor("list", "pods", func(action core.Action) (bool, runtime.Object, error) {
		list := action.(core.ListAction)
		fieldstring := list.GetListRestrictions().Fields.String()
		if strings.Contains(fieldstring, "n1") {
			return true, &apiv1.PodList{Items: []apiv1.Pod{*p1}}, nil
		}
		if strings.Contains(fieldstring, "n2") {
			return true, &apiv1.PodList{Items: []apiv1.Pod{*p2}}, nil
		}
		return true, nil, fmt.Errorf("Failed to list: %v", list)
	})

	provider := testprovider.NewTestCloudProvider(func(nodeGroup string, increase int) error {
		t.Fatalf("No expansion is expected, but increased %s by %d", nodeGroup, increase)
		return nil
	}, nil)
	provider.AddNodeGroup("ng1", 1, 10, 1)
	provider.AddNodeGroup("ng2", 1, 10, 5)
	provider.AddNode("ng1", n1)
	provider.AddNode("ng2", n2)

	clusterState := clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{})
	clusterState.UpdateNodes([]*apiv1.Node{n1, n2}, time.Now())
	fakeRecorder := kube_util.CreateEventRecorder(fakeClient)
	fakeLogRecorder, _ := utils.NewStatusMapRecorder(fakeClient, fakeRecorder, false)
	context := &AutoscalingContext{
		AutoscalingOptions: AutoscalingOptions{
			EstimatorName: estimator.BinpackingEstimatorName,
		},
		PredicateChecker:     simulator.NewTestPredicateChecker(),
		CloudProvider:        provider,
		ClientSet:            fakeClient,
		Recorder:             fakeRecorder,
		ExpanderStrategy:     random.NewStrategy(),
		ClusterStateRegistry: clusterState,
		LogRecorder:          fakeLogRecorder,
	}
	p3 := BuildTestPod("p-new", 550, 0)

	result, err := ScaleUp(context, []*apiv1.Pod{p3}, []*apiv1.Node{n1, n2})
	assert.NoError(t, err)
	// Node group is unhealthy.
	assert.False(t, result)
}

func TestScaleUpNoHelp(t *testing.T) {
	fakeClient := &fake.Clientset{}
	n1 := BuildTestNode("n1", 100, 1000)
	SetNodeReadyState(n1, true, time.Now())

	p1 := BuildTestPod("p1", 80, 0)
	p1.Spec.NodeName = "n1"

	fakeClient.Fake.AddReactor("list", "pods", func(action core.Action) (bool, runtime.Object, error) {
		list := action.(core.ListAction)
		fieldstring := list.GetListRestrictions().Fields.String()
		if strings.Contains(fieldstring, "n1") {
			return true, &apiv1.PodList{Items: []apiv1.Pod{*p1}}, nil
		}
		return true, nil, fmt.Errorf("Failed to list: %v", list)
	})

	provider := testprovider.NewTestCloudProvider(func(nodeGroup string, increase int) error {
		t.Fatalf("No expansion is expected")
		return nil
	}, nil)
	provider.AddNodeGroup("ng1", 1, 10, 1)
	provider.AddNode("ng1", n1)
	assert.NotNil(t, provider)

	clusterState := clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{})
	clusterState.UpdateNodes([]*apiv1.Node{n1}, time.Now())
	fakeRecorder := kube_util.CreateEventRecorder(fakeClient)
	fakeLogRecorder, _ := utils.NewStatusMapRecorder(fakeClient, fakeRecorder, false)
	context := &AutoscalingContext{
		AutoscalingOptions: AutoscalingOptions{
			EstimatorName: estimator.BinpackingEstimatorName,
		},
		PredicateChecker:     simulator.NewTestPredicateChecker(),
		CloudProvider:        provider,
		ClientSet:            fakeClient,
		Recorder:             fakeRecorder,
		ExpanderStrategy:     random.NewStrategy(),
		ClusterStateRegistry: clusterState,
		LogRecorder:          fakeLogRecorder,
	}
	p3 := BuildTestPod("p-new", 500, 0)

	result, err := ScaleUp(context, []*apiv1.Pod{p3}, []*apiv1.Node{n1})
	assert.NoError(t, err)
	assert.False(t, result)
}
