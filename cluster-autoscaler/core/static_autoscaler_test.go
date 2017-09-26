/*
Copyright 2017 The Kubernetes Authors.

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
	"reflect"
	"testing"
	"time"

	testprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate/utils"
	"k8s.io/autoscaler/cluster-autoscaler/estimator"
	"k8s.io/autoscaler/cluster-autoscaler/expander/random"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"

	apiv1 "k8s.io/api/core/v1"
	extensionsv1 "k8s.io/api/extensions/v1beta1"
	policyv1 "k8s.io/api/policy/v1beta1"
	"k8s.io/client-go/kubernetes/fake"
	kube_record "k8s.io/client-go/tools/record"
	"k8s.io/kubernetes/plugin/pkg/scheduler/schedulercache"

	"github.com/golang/glog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type nodeListerMock struct {
	mock.Mock
}

func (l *nodeListerMock) List() ([]*apiv1.Node, error) {
	args := l.Called()
	return args.Get(0).([]*apiv1.Node), args.Error(1)
}

type podListerMock struct {
	mock.Mock
}

func (l *podListerMock) List() ([]*apiv1.Pod, error) {
	args := l.Called()
	return args.Get(0).([]*apiv1.Pod), args.Error(1)
}

type podDisruptionBudgetListerMock struct {
	mock.Mock
}

func (l *podDisruptionBudgetListerMock) List() ([]*policyv1.PodDisruptionBudget, error) {
	args := l.Called()
	return args.Get(0).([]*policyv1.PodDisruptionBudget), args.Error(1)
}

type daemonSetListerMock struct {
	mock.Mock
}

func (l *daemonSetListerMock) List() ([]*extensionsv1.DaemonSet, error) {
	args := l.Called()
	return args.Get(0).([]*extensionsv1.DaemonSet), args.Error(1)
}

type onScaleUpMock struct {
	mock.Mock
}

func (m *onScaleUpMock) ScaleUp(id string, delta int) error {
	glog.Infof("Scale up: %v %v", id, delta)
	args := m.Called(id, delta)
	return args.Error(0)
}

type onScaleDownMock struct {
	mock.Mock
}

func (m *onScaleDownMock) ScaleDown(id string, name string) error {
	glog.Infof("Scale down: %v %v", id, name)
	args := m.Called(id, name)
	return args.Error(0)
}

type onNodeGroupCreateMock struct {
	mock.Mock
}

func (m *onNodeGroupCreateMock) Create(id string) error {
	glog.Infof("Create group: %v", id)
	args := m.Called(id)
	return args.Error(0)
}

type onNodeGroupDeleteMock struct {
	mock.Mock
}

func (m *onNodeGroupDeleteMock) Delete(id string) error {
	glog.Infof("Delete group: %v", id)
	args := m.Called(id)
	return args.Error(0)
}

func TestStaticAutoscalerRunOnce(t *testing.T) {
	readyNodeListerMock := &nodeListerMock{}
	allNodeListerMock := &nodeListerMock{}
	scheduledPodMock := &podListerMock{}
	unschedulablePodMock := &podListerMock{}
	podDisruptionBudgetListerMock := &podDisruptionBudgetListerMock{}
	daemonSetListerMock := &daemonSetListerMock{}
	onScaleUpMock := &onScaleUpMock{}
	onScaleDownMock := &onScaleDownMock{}

	n1 := BuildTestNode("n1", 1000, 1000)
	n1.Spec.ProviderID = "n1"
	SetNodeReadyState(n1, true, time.Now())
	n2 := BuildTestNode("n2", 1000, 1000)
	n2.Spec.ProviderID = "n2"
	SetNodeReadyState(n2, true, time.Now())
	n3 := BuildTestNode("n3", 1000, 1000)
	n3.Spec.ProviderID = "n3"

	p1 := BuildTestPod("p1", 600, 100)
	p1.Spec.NodeName = "n1"
	p2 := BuildTestPod("p2", 600, 100)

	tn := BuildTestNode("tn", 1000, 1000)
	tni := schedulercache.NewNodeInfo()
	tni.SetNode(tn)

	provider := testprovider.NewTestAutoprovisioningCloudProvider(
		func(id string, delta int) error {
			return onScaleUpMock.ScaleUp(id, delta)
		}, func(id string, name string) error {
			return onScaleDownMock.ScaleDown(id, name)
		},
		nil, nil,
		nil, map[string]*schedulercache.NodeInfo{"ng1": tni})
	provider.AddNodeGroup("ng1", 1, 10, 1)
	provider.AddNode("ng1", n1)
	ng1 := reflect.ValueOf(provider.NodeGroups()[0]).Interface().(*testprovider.TestNodeGroup)
	assert.NotNil(t, provider)

	fakeClient := &fake.Clientset{}
	fakeRecorder := kube_record.NewFakeRecorder(5)
	fakeLogRecorder, _ := utils.NewStatusMapRecorder(fakeClient, "kube-system", kube_record.NewFakeRecorder(5), false)
	clusterState := clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{}, fakeLogRecorder)
	clusterState.UpdateNodes([]*apiv1.Node{n1, n2}, time.Now())

	context := &AutoscalingContext{
		AutoscalingOptions: AutoscalingOptions{
			EstimatorName:                 estimator.BinpackingEstimatorName,
			ScaleDownEnabled:              true,
			ScaleDownUtilizationThreshold: 0.5,
			MaxNodesTotal:                 1,
			MaxCoresTotal:                 10,
			MaxMemoryTotal:                100000,
			ScaleDownUnreadyTime:          time.Minute,
			ScaleDownUnneededTime:         time.Minute,
		},
		PredicateChecker:     simulator.NewTestPredicateChecker(),
		CloudProvider:        provider,
		ClientSet:            fakeClient,
		Recorder:             fakeRecorder,
		ExpanderStrategy:     random.NewStrategy(),
		ClusterStateRegistry: clusterState,
		LogRecorder:          fakeLogRecorder,
	}

	listerRegistry := kube_util.NewListerRegistry(allNodeListerMock, readyNodeListerMock, scheduledPodMock,
		unschedulablePodMock, podDisruptionBudgetListerMock, daemonSetListerMock)

	sd := NewScaleDown(context)

	autoscaler := &StaticAutoscaler{AutoscalingContext: context,
		ListerRegistry:        listerRegistry,
		lastScaleUpTime:       time.Now(),
		lastScaleDownFailTime: time.Now(),
		scaleDown:             sd}

	// MaxNodesTotal reached.
	readyNodeListerMock.On("List").Return([]*apiv1.Node{n1}, nil).Once()
	allNodeListerMock.On("List").Return([]*apiv1.Node{n1}, nil).Once()
	scheduledPodMock.On("List").Return([]*apiv1.Pod{p1}, nil).Once()
	unschedulablePodMock.On("List").Return([]*apiv1.Pod{p2}, nil).Once()
	podDisruptionBudgetListerMock.On("List").Return([]*policyv1.PodDisruptionBudget{}, nil).Once()

	err := autoscaler.RunOnce(time.Now())
	assert.NoError(t, err)
	mock.AssertExpectationsForObjects(t, readyNodeListerMock, allNodeListerMock, scheduledPodMock, unschedulablePodMock,
		podDisruptionBudgetListerMock, daemonSetListerMock, onScaleUpMock, onScaleDownMock)

	// Scale up.
	readyNodeListerMock.On("List").Return([]*apiv1.Node{n1}, nil).Once()
	allNodeListerMock.On("List").Return([]*apiv1.Node{n1}, nil).Once()
	scheduledPodMock.On("List").Return([]*apiv1.Pod{p1}, nil).Once()
	unschedulablePodMock.On("List").Return([]*apiv1.Pod{p2}, nil).Once()
	daemonSetListerMock.On("List").Return([]*extensionsv1.DaemonSet{}, nil).Once()
	onScaleUpMock.On("ScaleUp", "ng1", 1).Return(nil).Once()

	context.MaxNodesTotal = 10
	err = autoscaler.RunOnce(time.Now().Add(time.Hour))
	assert.NoError(t, err)
	mock.AssertExpectationsForObjects(t, readyNodeListerMock, allNodeListerMock, scheduledPodMock, unschedulablePodMock,
		podDisruptionBudgetListerMock, daemonSetListerMock, onScaleUpMock, onScaleDownMock)

	// Mark unneeded nodes.
	readyNodeListerMock.On("List").Return([]*apiv1.Node{n1, n2}, nil).Once()
	allNodeListerMock.On("List").Return([]*apiv1.Node{n1, n2}, nil).Once()
	scheduledPodMock.On("List").Return([]*apiv1.Pod{p1}, nil).Once()
	unschedulablePodMock.On("List").Return([]*apiv1.Pod{}, nil).Once()
	podDisruptionBudgetListerMock.On("List").Return([]*policyv1.PodDisruptionBudget{}, nil).Once()

	provider.AddNode("ng1", n2)
	ng1.SetTargetSize(2)

	err = autoscaler.RunOnce(time.Now().Add(2 * time.Hour))
	assert.NoError(t, err)
	mock.AssertExpectationsForObjects(t, readyNodeListerMock, allNodeListerMock, scheduledPodMock, unschedulablePodMock,
		podDisruptionBudgetListerMock, daemonSetListerMock, onScaleUpMock, onScaleDownMock)

	// Scale down.
	readyNodeListerMock.On("List").Return([]*apiv1.Node{n1, n2}, nil).Once()
	allNodeListerMock.On("List").Return([]*apiv1.Node{n1, n2}, nil).Once()
	scheduledPodMock.On("List").Return([]*apiv1.Pod{p1}, nil).Once()
	unschedulablePodMock.On("List").Return([]*apiv1.Pod{}, nil).Once()
	podDisruptionBudgetListerMock.On("List").Return([]*policyv1.PodDisruptionBudget{}, nil).Once()
	onScaleDownMock.On("ScaleDown", "ng1", "n2").Return(nil).Once()

	err = autoscaler.RunOnce(time.Now().Add(3 * time.Hour))
	waitForDeleteToFinish(t, autoscaler.scaleDown)
	assert.NoError(t, err)
	mock.AssertExpectationsForObjects(t, readyNodeListerMock, allNodeListerMock, scheduledPodMock, unschedulablePodMock,
		podDisruptionBudgetListerMock, daemonSetListerMock, onScaleUpMock, onScaleDownMock)

	// Mark unregistered nodes.
	readyNodeListerMock.On("List").Return([]*apiv1.Node{n1, n2}, nil).Once()
	allNodeListerMock.On("List").Return([]*apiv1.Node{n1, n2}, nil).Once()
	scheduledPodMock.On("List").Return([]*apiv1.Pod{p1}, nil).Once()
	unschedulablePodMock.On("List").Return([]*apiv1.Pod{p2}, nil).Once()
	podDisruptionBudgetListerMock.On("List").Return([]*policyv1.PodDisruptionBudget{}, nil).Once()

	provider.AddNode("ng1", n3)
	ng1.SetTargetSize(3)

	err = autoscaler.RunOnce(time.Now().Add(4 * time.Hour))
	assert.NoError(t, err)
	mock.AssertExpectationsForObjects(t, readyNodeListerMock, allNodeListerMock, scheduledPodMock, unschedulablePodMock,
		podDisruptionBudgetListerMock, daemonSetListerMock, onScaleUpMock, onScaleDownMock)

	// Remove unregistered nodes.
	readyNodeListerMock.On("List").Return([]*apiv1.Node{n1, n2}, nil).Once()
	allNodeListerMock.On("List").Return([]*apiv1.Node{n1, n2}, nil).Once()
	onScaleDownMock.On("ScaleDown", "ng1", "n3").Return(nil).Once()

	err = autoscaler.RunOnce(time.Now().Add(5 * time.Hour))
	waitForDeleteToFinish(t, autoscaler.scaleDown)
	assert.NoError(t, err)
	mock.AssertExpectationsForObjects(t, readyNodeListerMock, allNodeListerMock, scheduledPodMock, unschedulablePodMock,
		podDisruptionBudgetListerMock, daemonSetListerMock, onScaleUpMock, onScaleDownMock)

}

func TestStaticAutoscalerRunOnceWithAutoprovisionedEnabled(t *testing.T) {
	readyNodeListerMock := &nodeListerMock{}
	allNodeListerMock := &nodeListerMock{}
	scheduledPodMock := &podListerMock{}
	unschedulablePodMock := &podListerMock{}
	podDisruptionBudgetListerMock := &podDisruptionBudgetListerMock{}
	daemonSetListerMock := &daemonSetListerMock{}
	onScaleUpMock := &onScaleUpMock{}
	onScaleDownMock := &onScaleDownMock{}
	onNodeGroupCreateMock := &onNodeGroupCreateMock{}
	onNodeGroupDeleteMock := &onNodeGroupDeleteMock{}

	n1 := BuildTestNode("n1", 100, 1000)
	n1.Spec.ProviderID = "n1"
	SetNodeReadyState(n1, true, time.Now())
	n2 := BuildTestNode("n2", 1000, 1000)
	n2.Spec.ProviderID = "n2"
	SetNodeReadyState(n2, true, time.Now())

	p1 := BuildTestPod("p1", 100, 100)
	p1.Spec.NodeName = "n1"
	p2 := BuildTestPod("p2", 600, 100)

	tn1 := BuildTestNode("tn1", 100, 1000)
	SetNodeReadyState(tn1, true, time.Now())
	tni1 := schedulercache.NewNodeInfo()
	tni1.SetNode(tn1)
	tn2 := BuildTestNode("tn2", 1000, 1000)
	SetNodeReadyState(tn2, true, time.Now())
	tni2 := schedulercache.NewNodeInfo()
	tni2.SetNode(tn2)
	tn3 := BuildTestNode("tn3", 100, 1000)
	SetNodeReadyState(tn2, true, time.Now())
	tni3 := schedulercache.NewNodeInfo()
	tni3.SetNode(tn3)

	provider := testprovider.NewTestAutoprovisioningCloudProvider(
		func(id string, delta int) error {
			return onScaleUpMock.ScaleUp(id, delta)
		}, func(id string, name string) error {
			return onScaleDownMock.ScaleDown(id, name)
		}, func(id string) error {
			return onNodeGroupCreateMock.Create(id)
		}, func(id string) error {
			return onNodeGroupDeleteMock.Delete(id)
		},
		[]string{"TN1", "TN2"}, map[string]*schedulercache.NodeInfo{"TN1": tni1, "TN2": tni2, "ng1": tni3})
	provider.AddNodeGroup("ng1", 1, 10, 1)
	provider.AddAutoprovisionedNodeGroup("autoprovisioned-TN1", 0, 10, 0, "TN1")
	autoprovisionedTN1 := reflect.ValueOf(provider.NodeGroups()[1]).Interface().(*testprovider.TestNodeGroup)
	provider.AddNode("ng1,", n1)
	assert.NotNil(t, provider)

	fakeClient := &fake.Clientset{}
	fakeRecorder := kube_record.NewFakeRecorder(5)
	fakeLogRecorder, _ := utils.NewStatusMapRecorder(fakeClient, "kube-system", kube_record.NewFakeRecorder(5), false)
	clusterState := clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{}, fakeLogRecorder)
	clusterState.UpdateNodes([]*apiv1.Node{n1}, time.Now())

	context := &AutoscalingContext{
		AutoscalingOptions: AutoscalingOptions{
			EstimatorName:                    estimator.BinpackingEstimatorName,
			ScaleDownEnabled:                 true,
			ScaleDownUtilizationThreshold:    0.5,
			MaxNodesTotal:                    100,
			MaxCoresTotal:                    100,
			MaxMemoryTotal:                   100000,
			ScaleDownUnreadyTime:             time.Minute,
			ScaleDownUnneededTime:            time.Minute,
			NodeAutoprovisioningEnabled:      true,
			MaxAutoprovisionedNodeGroupCount: 10,
		},
		PredicateChecker:     simulator.NewTestPredicateChecker(),
		CloudProvider:        provider,
		ClientSet:            fakeClient,
		Recorder:             fakeRecorder,
		ExpanderStrategy:     random.NewStrategy(),
		ClusterStateRegistry: clusterState,
		LogRecorder:          fakeLogRecorder,
	}

	listerRegistry := kube_util.NewListerRegistry(allNodeListerMock, readyNodeListerMock, scheduledPodMock,
		unschedulablePodMock, podDisruptionBudgetListerMock, daemonSetListerMock)

	sd := NewScaleDown(context)

	autoscaler := &StaticAutoscaler{AutoscalingContext: context,
		ListerRegistry:        listerRegistry,
		lastScaleUpTime:       time.Now(),
		lastScaleDownFailTime: time.Now(),
		scaleDown:             sd}

	// Scale up.
	readyNodeListerMock.On("List").Return([]*apiv1.Node{n1}, nil).Once()
	allNodeListerMock.On("List").Return([]*apiv1.Node{n1}, nil).Once()
	scheduledPodMock.On("List").Return([]*apiv1.Pod{p1}, nil).Once()
	unschedulablePodMock.On("List").Return([]*apiv1.Pod{p2}, nil).Once()
	daemonSetListerMock.On("List").Return([]*extensionsv1.DaemonSet{}, nil).Once()
	onNodeGroupCreateMock.On("Create", "autoprovisioned-TN2").Return(nil).Once()
	onScaleUpMock.On("ScaleUp", "autoprovisioned-TN2", 1).Return(nil).Once()

	err := autoscaler.RunOnce(time.Now().Add(time.Hour))
	assert.NoError(t, err)
	mock.AssertExpectationsForObjects(t, readyNodeListerMock, allNodeListerMock, scheduledPodMock, unschedulablePodMock,
		podDisruptionBudgetListerMock, daemonSetListerMock, onScaleUpMock, onScaleDownMock)

	// Fix target size.
	autoprovisionedTN1.SetTargetSize(0)

	// Remove autoprovisioned node group and mark unneeded nodes.
	readyNodeListerMock.On("List").Return([]*apiv1.Node{n1, n2}, nil).Once()
	allNodeListerMock.On("List").Return([]*apiv1.Node{n1, n2}, nil).Once()
	scheduledPodMock.On("List").Return([]*apiv1.Pod{p1}, nil).Once()
	unschedulablePodMock.On("List").Return([]*apiv1.Pod{}, nil).Once()
	podDisruptionBudgetListerMock.On("List").Return([]*policyv1.PodDisruptionBudget{}, nil).Once()
	onNodeGroupDeleteMock.On("Delete", "autoprovisioned-TN1").Return(nil).Once()

	provider.AddAutoprovisionedNodeGroup("autoprovisioned-TN2", 0, 10, 1, "TN1")
	provider.AddNode("autoprovisioned-TN2", n2)

	err = autoscaler.RunOnce(time.Now().Add(1 * time.Hour))
	assert.NoError(t, err)
	mock.AssertExpectationsForObjects(t, readyNodeListerMock, allNodeListerMock, scheduledPodMock, unschedulablePodMock,
		podDisruptionBudgetListerMock, daemonSetListerMock, onScaleUpMock, onScaleDownMock)

	// Scale down.
	readyNodeListerMock.On("List").Return([]*apiv1.Node{n1, n2}, nil).Once()
	allNodeListerMock.On("List").Return([]*apiv1.Node{n1, n2}, nil).Once()
	scheduledPodMock.On("List").Return([]*apiv1.Pod{p1}, nil).Once()
	unschedulablePodMock.On("List").Return([]*apiv1.Pod{}, nil).Once()
	podDisruptionBudgetListerMock.On("List").Return([]*policyv1.PodDisruptionBudget{}, nil).Once()
	onNodeGroupDeleteMock.On("Delete", "autoprovisioned-"+
		"TN1").Return(nil).Once()
	onScaleDownMock.On("ScaleDown", "autoprovisioned-TN2", "n2").Return(nil).Once()

	err = autoscaler.RunOnce(time.Now().Add(2 * time.Hour))
	waitForDeleteToFinish(t, autoscaler.scaleDown)
	assert.NoError(t, err)
	mock.AssertExpectationsForObjects(t, readyNodeListerMock, allNodeListerMock, scheduledPodMock, unschedulablePodMock,
		podDisruptionBudgetListerMock, daemonSetListerMock, onScaleUpMock, onScaleDownMock)
}
