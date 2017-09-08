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
	glog.Warningf("Scale up: %v %v", id, delta)
	args := m.Called(id, delta)
	return args.Error(0)
}

type onScaleDownMock struct {
	mock.Mock
}

func (m *onScaleDownMock) ScaleDown(id string, name string) error {
	glog.Warningf("Scale down: %v %v", id, name)
	args := m.Called(id, name)
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
		ListerRegistry:           listerRegistry,
		lastScaleUpTime:          time.Now(),
		lastScaleDownFailedTrial: time.Now(),
		scaleDown:                sd}

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

	// Do nothing.
	readyNodeListerMock.On("List").Return([]*apiv1.Node{n1, n2}, nil).Once()
	allNodeListerMock.On("List").Return([]*apiv1.Node{n1, n2}, nil).Once()
	scheduledPodMock.On("List").Return([]*apiv1.Pod{p1}, nil).Once()
	unschedulablePodMock.On("List").Return([]*apiv1.Pod{p2}, nil).Once()
	podDisruptionBudgetListerMock.On("List").Return([]*policyv1.PodDisruptionBudget{}, nil).Once()

	provider.AddNode("ng1", n2)

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
