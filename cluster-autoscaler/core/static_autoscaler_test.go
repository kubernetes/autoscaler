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
	"bytes"
	stdcontext "context"
	"flag"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	mockprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/mocks"
	testprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate"
	clusterstate_utils "k8s.io/autoscaler/cluster-autoscaler/clusterstate/utils"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/actuation"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/deletiontracker"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/planner"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/status"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaleup/orchestrator"
	. "k8s.io/autoscaler/cluster-autoscaler/core/test"
	core_utils "k8s.io/autoscaler/cluster-autoscaler/core/utils"
	"k8s.io/autoscaler/cluster-autoscaler/estimator"
	"k8s.io/autoscaler/cluster-autoscaler/observers/loopstart"
	ca_processors "k8s.io/autoscaler/cluster-autoscaler/processors"
	"k8s.io/autoscaler/cluster-autoscaler/processors/callbacks"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodegroupconfig"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodegroups/asyncnodegroups"
	"k8s.io/autoscaler/cluster-autoscaler/processors/scaledowncandidates"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/drainability/rules"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/options"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/utilization"
	"k8s.io/autoscaler/cluster-autoscaler/utils/drain"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	"k8s.io/autoscaler/cluster-autoscaler/utils/scheduler"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
	kube_record "k8s.io/client-go/tools/record"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes/fake"
	v1appslister "k8s.io/client-go/listers/apps/v1"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	klog "k8s.io/klog/v2"
)

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

func (l *daemonSetListerMock) List(selector labels.Selector) ([]*appsv1.DaemonSet, error) {
	args := l.Called(selector)
	return args.Get(0).([]*appsv1.DaemonSet), args.Error(1)
}

func (l *daemonSetListerMock) DaemonSets(namespace string) v1appslister.DaemonSetNamespaceLister {
	args := l.Called(namespace)
	return args.Get(0).(v1appslister.DaemonSetNamespaceLister)
}

func (l *daemonSetListerMock) GetPodDaemonSets(pod *apiv1.Pod) ([]*appsv1.DaemonSet, error) {
	args := l.Called()
	return args.Get(0).([]*appsv1.DaemonSet), args.Error(1)
}

func (l *daemonSetListerMock) GetHistoryDaemonSets(history *appsv1.ControllerRevision) ([]*appsv1.DaemonSet, error) {
	args := l.Called()
	return args.Get(0).([]*appsv1.DaemonSet), args.Error(1)
}

type onScaleUpMock struct {
	mock.Mock
}

func (m *onScaleUpMock) ScaleUp(id string, delta int) error {
	klog.Infof("Scale up: %v %v", id, delta)
	args := m.Called(id, delta)
	return args.Error(0)
}

type onScaleDownMock struct {
	mock.Mock
}

func (m *onScaleDownMock) ScaleDown(id string, name string) error {
	klog.Infof("Scale down: %v %v", id, name)
	args := m.Called(id, name)
	return args.Error(0)
}

type onNodeGroupCreateMock struct {
	mock.Mock
}

func (m *onNodeGroupCreateMock) Create(id string) error {
	klog.Infof("Create group: %v", id)
	args := m.Called(id)
	return args.Error(0)
}

type onNodeGroupDeleteMock struct {
	mock.Mock
}

func (m *onNodeGroupDeleteMock) Delete(id string) error {
	klog.Infof("Delete group: %v", id)
	args := m.Called(id)
	return args.Error(0)
}

func setUpScaleDownActuator(ctx *context.AutoscalingContext, autoscalingOptions config.AutoscalingOptions) {
	deleteOptions := options.NewNodeDeleteOptions(autoscalingOptions)
	ctx.ScaleDownActuator = actuation.NewActuator(ctx, nil, deletiontracker.NewNodeDeletionTracker(0*time.Second), deleteOptions, rules.Default(deleteOptions), NewTestProcessors(ctx).NodeGroupConfigProcessor)
}

type nodeGroup struct {
	name  string
	nodes []*apiv1.Node
	min   int
	max   int
}
type scaleCall struct {
	ng    string
	delta int
}

type scaleDownStatusProcessorMock struct {
	called          int
	scaleDownStatus *status.ScaleDownStatus
}

func (p *scaleDownStatusProcessorMock) Process(_ *context.AutoscalingContext, st *status.ScaleDownStatus) {
	p.called += 1
	p.scaleDownStatus = st
}

func (p *scaleDownStatusProcessorMock) CleanUp() {
}

type commonMocks struct {
	readyNodeLister           *kube_util.TestNodeLister
	allNodeLister             *kube_util.TestNodeLister
	allPodLister              *podListerMock
	podDisruptionBudgetLister *podDisruptionBudgetListerMock
	daemonSetLister           *daemonSetListerMock
	nodeDeletionTracker       *deletiontracker.NodeDeletionTracker

	onScaleUp   *onScaleUpMock
	onScaleDown *onScaleDownMock
}

func newCommonMocks() *commonMocks {
	return &commonMocks{
		readyNodeLister:           kubernetes.NewTestNodeLister(nil),
		allNodeLister:             kubernetes.NewTestNodeLister(nil),
		allPodLister:              &podListerMock{},
		podDisruptionBudgetLister: &podDisruptionBudgetListerMock{},
		daemonSetLister:           &daemonSetListerMock{},
		onScaleUp:                 &onScaleUpMock{},
		onScaleDown:               &onScaleDownMock{},
	}
}

type autoscalerSetupConfig struct {
	nodeGroups          []*nodeGroup
	nodeStateUpdateTime time.Time
	autoscalingOptions  config.AutoscalingOptions
	clusterStateConfig  clusterstate.ClusterStateRegistryConfig
	mocks               *commonMocks
	nodesDeleted        chan bool
}

func setupCloudProvider(config *autoscalerSetupConfig) (*testprovider.TestCloudProvider, error) {
	provider := testprovider.NewTestCloudProvider(
		func(id string, delta int) error {
			return config.mocks.onScaleUp.ScaleUp(id, delta)
		}, func(id string, name string) error {
			ret := config.mocks.onScaleDown.ScaleDown(id, name)
			config.nodesDeleted <- true
			return ret
		})

	for _, ng := range config.nodeGroups {
		provider.AddNodeGroup(ng.name, ng.min, ng.max, len(ng.nodes))
		for _, node := range ng.nodes {
			provider.AddNode(ng.name, node)
		}
		reflectedNg := reflect.ValueOf(provider.GetNodeGroup(ng.name)).Interface().(*testprovider.TestNodeGroup)
		if reflectedNg == nil {
			return nil, fmt.Errorf("Nodegroup '%v' found as nil after setting up cloud provider", ng.name)
		}
	}
	return provider, nil
}

func setupAutoscalingContext(opts config.AutoscalingOptions, provider cloudprovider.CloudProvider, processorCallbacks callbacks.ProcessorCallbacks) (context.AutoscalingContext, error) {
	context, err := NewScaleTestAutoscalingContext(opts, &fake.Clientset{}, nil, provider, processorCallbacks, nil)
	if err != nil {
		return context, err
	}
	return context, nil
}

func setupAutoscaler(config *autoscalerSetupConfig) (*StaticAutoscaler, error) {
	provider, err := setupCloudProvider(config)
	if err != nil {
		return nil, err
	}

	allNodes := make([]*apiv1.Node, 0)
	for _, ng := range config.nodeGroups {
		allNodes = append(allNodes, ng.nodes...)
	}

	// Create context with mocked lister registry.
	processorCallbacks := newStaticAutoscalerProcessorCallbacks()

	context, err := setupAutoscalingContext(config.autoscalingOptions, provider, processorCallbacks)

	if err != nil {
		return nil, err
	}

	setUpScaleDownActuator(&context, config.autoscalingOptions)

	listerRegistry := kube_util.NewListerRegistry(config.mocks.allNodeLister, config.mocks.readyNodeLister, config.mocks.allPodLister,
		config.mocks.podDisruptionBudgetLister, config.mocks.daemonSetLister,
		nil, nil, nil, nil)
	context.ListerRegistry = listerRegistry

	ngConfigProcesssor := nodegroupconfig.NewDefaultNodeGroupConfigProcessor(config.autoscalingOptions.NodeGroupDefaults)

	processors := NewTestProcessors(&context)

	clusterState := clusterstate.NewClusterStateRegistry(provider, config.clusterStateConfig, context.LogRecorder, NewBackoff(), ngConfigProcesssor, processors.AsyncNodeGroupStateChecker)

	clusterState.UpdateNodes(allNodes, nil, config.nodeStateUpdateTime)

	sdPlanner, sdActuator := newScaleDownPlannerAndActuator(&context, processors, clusterState, config.mocks.nodeDeletionTracker)
	suOrchestrator := orchestrator.New()

	suOrchestrator.Initialize(&context, processors, clusterState, newEstimatorBuilder(), taints.TaintConfig{})

	autoscaler := &StaticAutoscaler{
		AutoscalingContext:   &context,
		clusterStateRegistry: clusterState,
		scaleDownPlanner:     sdPlanner,
		scaleDownActuator:    sdActuator,
		scaleUpOrchestrator:  suOrchestrator,
		processors:           processors,
		loopStartNotifier:    loopstart.NewObserversList(nil),
		processorCallbacks:   processorCallbacks,
	}

	return autoscaler, nil
}

// TODO: Refactor tests to use setupAutoscaler

func TestStaticAutoscalerRunOnce(t *testing.T) {
	readyNodeLister := kubernetes.NewTestNodeLister(nil)
	allNodeLister := kubernetes.NewTestNodeLister(nil)
	allPodListerMock := &podListerMock{}
	podDisruptionBudgetListerMock := &podDisruptionBudgetListerMock{}
	daemonSetListerMock := &daemonSetListerMock{}
	onScaleUpMock := &onScaleUpMock{}
	onScaleDownMock := &onScaleDownMock{}
	deleteFinished := make(chan bool, 1)

	n1 := BuildTestNode("n1", 1000, 1000)
	SetNodeReadyState(n1, true, time.Now())
	n2 := BuildTestNode("n2", 1000, 1000)
	SetNodeReadyState(n2, true, time.Now())
	n3 := BuildTestNode("n3", 1000, 1000)
	n4 := BuildTestNode("n4", 1000, 1000)

	p1 := BuildTestPod("p1", 600, 100)
	p1.Spec.NodeName = "n1"
	p2 := BuildTestPod("p2", 600, 100, MarkUnschedulable())

	tn := BuildTestNode("tn", 1000, 1000)
	tni := schedulerframework.NewNodeInfo()
	tni.SetNode(tn)

	provider := testprovider.NewTestAutoprovisioningCloudProvider(
		func(id string, delta int) error {
			return onScaleUpMock.ScaleUp(id, delta)
		}, func(id string, name string) error {
			ret := onScaleDownMock.ScaleDown(id, name)
			deleteFinished <- true
			return ret
		},
		nil, nil,
		nil, map[string]*schedulerframework.NodeInfo{"ng1": tni, "ng2": tni, "ng3": tni})
	provider.AddNodeGroup("ng1", 1, 10, 1)
	provider.AddNode("ng1", n1)
	ng1 := reflect.ValueOf(provider.GetNodeGroup("ng1")).Interface().(*testprovider.TestNodeGroup)
	assert.NotNil(t, ng1)
	assert.NotNil(t, provider)

	// Create context with mocked lister registry.
	options := config.AutoscalingOptions{
		NodeGroupDefaults: config.NodeGroupAutoscalingOptions{
			ScaleDownUnneededTime:         time.Minute,
			ScaleDownUnreadyTime:          time.Minute,
			ScaleDownUtilizationThreshold: 0.5,
			MaxNodeProvisionTime:          10 * time.Second,
		},
		EstimatorName:           estimator.BinpackingEstimatorName,
		EnforceNodeGroupMinSize: true,
		ScaleDownEnabled:        true,
		MaxNodesTotal:           1,
		MaxCoresTotal:           10,
		MaxMemoryTotal:          100000,
	}
	processorCallbacks := newStaticAutoscalerProcessorCallbacks()

	context, err := NewScaleTestAutoscalingContext(options, &fake.Clientset{}, nil, provider, processorCallbacks, nil)
	assert.NoError(t, err)

	setUpScaleDownActuator(&context, options)

	listerRegistry := kube_util.NewListerRegistry(allNodeLister, readyNodeLister, allPodListerMock, podDisruptionBudgetListerMock, daemonSetListerMock,
		nil, nil, nil, nil)
	context.ListerRegistry = listerRegistry

	clusterStateConfig := clusterstate.ClusterStateRegistryConfig{
		OkTotalUnreadyCount: 1,
	}
	processors := NewTestProcessors(&context)
	clusterState := clusterstate.NewClusterStateRegistry(provider, clusterStateConfig, context.LogRecorder, NewBackoff(), nodegroupconfig.NewDefaultNodeGroupConfigProcessor(options.NodeGroupDefaults), processors.AsyncNodeGroupStateChecker)
	sdPlanner, sdActuator := newScaleDownPlannerAndActuator(&context, processors, clusterState, nil)
	suOrchestrator := orchestrator.New()
	suOrchestrator.Initialize(&context, processors, clusterState, newEstimatorBuilder(), taints.TaintConfig{})

	autoscaler := &StaticAutoscaler{
		AutoscalingContext:    &context,
		clusterStateRegistry:  clusterState,
		lastScaleUpTime:       time.Now(),
		lastScaleDownFailTime: time.Now(),
		scaleDownPlanner:      sdPlanner,
		scaleDownActuator:     sdActuator,
		scaleUpOrchestrator:   suOrchestrator,
		processors:            processors,
		loopStartNotifier:     loopstart.NewObserversList(nil),
		processorCallbacks:    processorCallbacks,
		initialized:           true,
	}

	// MaxNodesTotal reached.
	readyNodeLister.SetNodes([]*apiv1.Node{n1})
	allNodeLister.SetNodes([]*apiv1.Node{n1})
	allPodListerMock.On("List").Return([]*apiv1.Pod{p1, p2}, nil).Twice()
	daemonSetListerMock.On("List", labels.Everything()).Return([]*appsv1.DaemonSet{}, nil).Once()
	podDisruptionBudgetListerMock.On("List").Return([]*policyv1.PodDisruptionBudget{}, nil).Once()

	err = autoscaler.RunOnce(time.Now())
	assert.NoError(t, err)
	mock.AssertExpectationsForObjects(t, allPodListerMock, podDisruptionBudgetListerMock, daemonSetListerMock, onScaleUpMock, onScaleDownMock)

	// Scale up.
	readyNodeLister.SetNodes([]*apiv1.Node{n1})
	allNodeLister.SetNodes([]*apiv1.Node{n1})
	allPodListerMock.On("List").Return([]*apiv1.Pod{p1, p2}, nil).Twice()
	daemonSetListerMock.On("List", labels.Everything()).Return([]*appsv1.DaemonSet{}, nil).Once()
	podDisruptionBudgetListerMock.On("List").Return([]*policyv1.PodDisruptionBudget{}, nil).Once()
	onScaleUpMock.On("ScaleUp", "ng1", 1).Return(nil).Once()

	context.MaxNodesTotal = 10
	err = autoscaler.RunOnce(time.Now().Add(time.Hour))
	assert.NoError(t, err)
	mock.AssertExpectationsForObjects(t, allPodListerMock,
		podDisruptionBudgetListerMock, daemonSetListerMock, onScaleUpMock, onScaleDownMock)

	// Mark unneeded nodes.
	readyNodeLister.SetNodes([]*apiv1.Node{n1, n2})
	allNodeLister.SetNodes([]*apiv1.Node{n1, n2})
	allPodListerMock.On("List").Return([]*apiv1.Pod{p1}, nil).Twice()
	daemonSetListerMock.On("List", labels.Everything()).Return([]*appsv1.DaemonSet{}, nil).Once()
	podDisruptionBudgetListerMock.On("List").Return([]*policyv1.PodDisruptionBudget{}, nil).Once()

	provider.AddNode("ng1", n2)
	ng1.SetTargetSize(2)

	err = autoscaler.RunOnce(time.Now().Add(2 * time.Hour))
	assert.NoError(t, err)
	mock.AssertExpectationsForObjects(t, allPodListerMock,
		podDisruptionBudgetListerMock, daemonSetListerMock, onScaleUpMock, onScaleDownMock)

	// Scale down.
	readyNodeLister.SetNodes([]*apiv1.Node{n1, n2})
	allNodeLister.SetNodes([]*apiv1.Node{n1, n2})
	allPodListerMock.On("List").Return([]*apiv1.Pod{p1}, nil).Times(3)
	daemonSetListerMock.On("List", labels.Everything()).Return([]*appsv1.DaemonSet{}, nil).Once()
	podDisruptionBudgetListerMock.On("List").Return([]*policyv1.PodDisruptionBudget{}, nil).Once()
	onScaleDownMock.On("ScaleDown", "ng1", "n2").Return(nil).Once()

	err = autoscaler.RunOnce(time.Now().Add(3 * time.Hour))
	waitForDeleteToFinish(t, deleteFinished)
	assert.NoError(t, err)
	mock.AssertExpectationsForObjects(t, allPodListerMock,
		podDisruptionBudgetListerMock, daemonSetListerMock, onScaleUpMock, onScaleDownMock)

	// Mark unregistered nodes.
	readyNodeLister.SetNodes([]*apiv1.Node{n1, n2})
	allNodeLister.SetNodes([]*apiv1.Node{n1, n2})
	allPodListerMock.On("List").Return([]*apiv1.Pod{p1, p2}, nil).Twice()
	daemonSetListerMock.On("List", labels.Everything()).Return([]*appsv1.DaemonSet{}, nil).Once()
	podDisruptionBudgetListerMock.On("List").Return([]*policyv1.PodDisruptionBudget{}, nil).Once()

	provider.AddNodeGroup("ng2", 0, 10, 1)
	provider.AddNode("ng2", n3)

	err = autoscaler.RunOnce(time.Now().Add(4 * time.Hour))
	assert.NoError(t, err)
	mock.AssertExpectationsForObjects(t, allPodListerMock,
		podDisruptionBudgetListerMock, daemonSetListerMock, onScaleUpMock, onScaleDownMock)

	// Remove unregistered nodes.
	readyNodeLister.SetNodes([]*apiv1.Node{n1, n2})
	allNodeLister.SetNodes([]*apiv1.Node{n1, n2})
	allPodListerMock.On("List").Return([]*apiv1.Pod{p1, p2}, nil).Twice()
	daemonSetListerMock.On("List", labels.Everything()).Return([]*appsv1.DaemonSet{}, nil).Once()
	onScaleDownMock.On("ScaleDown", "ng2", "n3").Return(nil).Once()
	podDisruptionBudgetListerMock.On("List").Return([]*policyv1.PodDisruptionBudget{}, nil).Once()

	err = autoscaler.RunOnce(time.Now().Add(5 * time.Hour))
	waitForDeleteToFinish(t, deleteFinished)
	assert.NoError(t, err)
	mock.AssertExpectationsForObjects(t, allPodListerMock,
		podDisruptionBudgetListerMock, daemonSetListerMock, onScaleUpMock, onScaleDownMock)

	// Scale up to node group min size.
	readyNodeLister.SetNodes([]*apiv1.Node{n4})
	allNodeLister.SetNodes([]*apiv1.Node{n4})
	allPodListerMock.On("List").Return([]*apiv1.Pod{}, nil).Twice()
	daemonSetListerMock.On("List", labels.Everything()).Return([]*appsv1.DaemonSet{}, nil)
	podDisruptionBudgetListerMock.On("List").Return([]*policyv1.PodDisruptionBudget{}, nil)
	onScaleUpMock.On("ScaleUp", "ng3", 2).Return(nil).Once() // 2 new nodes are supposed to be scaled up.

	provider.AddNodeGroup("ng3", 3, 10, 1)
	provider.AddNode("ng3", n4)

	err = autoscaler.RunOnce(time.Now().Add(5 * time.Hour))
	assert.NoError(t, err)
	mock.AssertExpectationsForObjects(t, onScaleUpMock)
}

func TestStaticAutoscalerRunOnceWithScaleDownDelayPerNG(t *testing.T) {
	onScaleUpMock := &onScaleUpMock{}
	onScaleDownMock := &onScaleDownMock{}
	deleteFinished := make(chan bool, 1)

	n1 := BuildTestNode("n1", 1000, 1000)
	SetNodeReadyState(n1, true, time.Now())
	n2 := BuildTestNode("n2", 1000, 1000)
	SetNodeReadyState(n2, true, time.Now())

	tn := BuildTestNode("tn", 1000, 1000)
	tni := schedulerframework.NewNodeInfo()
	tni.SetNode(tn)

	provider := testprovider.NewTestAutoprovisioningCloudProvider(
		func(id string, delta int) error {
			return onScaleUpMock.ScaleUp(id, delta)
		}, func(id string, name string) error {
			ret := onScaleDownMock.ScaleDown(id, name)
			deleteFinished <- true
			return ret
		},
		nil, nil,
		nil, map[string]*schedulerframework.NodeInfo{"ng1": tni, "ng2": tni})
	assert.NotNil(t, provider)

	provider.AddNodeGroup("ng1", 0, 10, 1)
	ng1 := reflect.ValueOf(provider.GetNodeGroup("ng1")).Interface().(*testprovider.TestNodeGroup)
	assert.NotNil(t, ng1)

	provider.AddNodeGroup("ng2", 0, 10, 1)
	ng2 := reflect.ValueOf(provider.GetNodeGroup("ng2")).Interface().(*testprovider.TestNodeGroup)
	assert.NotNil(t, ng2)

	p1 := BuildTestPod("p1", 400, 100)
	p1.Annotations[drain.PodSafeToEvictKey] = "true"
	p1.Spec.NodeName = "n1"

	p2 := BuildTestPod("p2", 400, 100)
	p2.Annotations[drain.PodSafeToEvictKey] = "true"
	p2.Spec.NodeName = "n2"

	testCases := []struct {
		description           string
		beforeTest            func(processors *ca_processors.AutoscalingProcessors)
		expectedScaleDownNG   string
		expectedScaleDownNode string
		afterTest             func(processors *ca_processors.AutoscalingProcessors)
	}{
		// Case 1:
		// ng1 scaled up recently
		// both ng1 and ng2 have under-utilized nodes
		// expectation: under-utilized node in ng2 should be scaled down
		{
			description: "ng1 scaled up recently - both ng1 and ng2 have under-utilized nodes",
			beforeTest: func(processors *ca_processors.AutoscalingProcessors) {
				// make CA think ng1 scaled up recently
				processors.ScaleStateNotifier.RegisterScaleUp(ng1, 1, time.Now().Add(-3*time.Minute))
			},
			expectedScaleDownNG:   "ng2",
			expectedScaleDownNode: "n2",
			afterTest: func(processors *ca_processors.AutoscalingProcessors) {
				// reset scale up in ng1 so that it doesn't block scale down in the next test
				// scale down is always recorded relative to time.Now(), no matter
				// what currentTime time is passed to RunOnce()
				processors.ScaleStateNotifier.RegisterScaleUp(ng1, 1, time.Time{})
			},
		},

		// Case 2:
		// ng2 scaled down recently
		// both ng1 and ng2 have under-utilized nodes
		// expectation: under-utilized node in ng1 should be scaled down
		{
			description: "ng2 scaled down recently - both ng1 and ng2 have under-utilized nodes",
			beforeTest: func(processors *ca_processors.AutoscalingProcessors) {
				// make CA think ng2 scaled down recently
				processors.ScaleStateNotifier.RegisterScaleDown(ng2, "n3", time.Now().Add(-3*time.Minute), time.Now())
			},
			expectedScaleDownNG:   "ng1",
			expectedScaleDownNode: "n1",
			afterTest: func(processors *ca_processors.AutoscalingProcessors) {
				// reset scale down in ng1 and ng2 so that it doesn't block scale down in the next test
				// scale down is always recorded relative to time.Now(), no matter
				// what currentTime time is passed to RunOnce()
				processors.ScaleStateNotifier.RegisterScaleDown(ng2, "n3", time.Time{}, time.Time{})
				processors.ScaleStateNotifier.RegisterScaleDown(ng1, "n1", time.Time{}, time.Time{})
			},
		},

		// Case 3:
		// ng1 had a scale down failure
		// both ng1 and ng2 have under-utilized nodes
		// expectation: under-utilized node in ng2 should be scaled down
		{
			description: "ng1 had scale-down failure - both ng1 and ng2 have under-utilized nodes",
			beforeTest: func(processors *ca_processors.AutoscalingProcessors) {
				// Make CA think scale down failed in ng1
				processors.ScaleStateNotifier.RegisterFailedScaleDown(ng1, "scale down failed", time.Now().Add(-3*time.Minute))
			},
			expectedScaleDownNG:   "ng2",
			expectedScaleDownNode: "n2",
			afterTest: func(processors *ca_processors.AutoscalingProcessors) {

			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			readyNodeLister := kubernetes.NewTestNodeLister(nil)
			allNodeLister := kubernetes.NewTestNodeLister(nil)
			allPodListerMock := &podListerMock{}
			podDisruptionBudgetListerMock := &podDisruptionBudgetListerMock{}
			daemonSetListerMock := &daemonSetListerMock{}

			// Create context with mocked lister registry.
			options := config.AutoscalingOptions{
				NodeGroupDefaults: config.NodeGroupAutoscalingOptions{
					ScaleDownUnneededTime:         config.DefaultScaleDownUnneededTime,
					ScaleDownUnreadyTime:          time.Minute,
					ScaleDownUtilizationThreshold: 0.5,
					MaxNodeProvisionTime:          10 * time.Second,
				},
				EstimatorName:              estimator.BinpackingEstimatorName,
				EnforceNodeGroupMinSize:    true,
				ScaleDownEnabled:           true,
				MaxNodesTotal:              1,
				MaxCoresTotal:              10,
				MaxMemoryTotal:             100000,
				ScaleDownDelayTypeLocal:    true,
				ScaleDownDelayAfterAdd:     5 * time.Minute,
				ScaleDownDelayAfterDelete:  5 * time.Minute,
				ScaleDownDelayAfterFailure: 5 * time.Minute,
			}
			processorCallbacks := newStaticAutoscalerProcessorCallbacks()

			context, err := NewScaleTestAutoscalingContext(options, &fake.Clientset{}, nil, provider, processorCallbacks, nil)
			assert.NoError(t, err)

			setUpScaleDownActuator(&context, options)

			listerRegistry := kube_util.NewListerRegistry(allNodeLister, readyNodeLister, allPodListerMock, podDisruptionBudgetListerMock, daemonSetListerMock,
				nil, nil, nil, nil)
			context.ListerRegistry = listerRegistry

			clusterStateConfig := clusterstate.ClusterStateRegistryConfig{
				OkTotalUnreadyCount: 1,
			}
			processors := NewTestProcessors(&context)
			sddProcessor := scaledowncandidates.NewScaleDownCandidatesDelayProcessor()
			processors.ScaleStateNotifier.Register(sddProcessor)
			scaleDownCandidatesComparers := []scaledowncandidates.CandidatesComparer{}
			cp := scaledowncandidates.NewCombinedScaleDownCandidatesProcessor()
			cp.Register(scaledowncandidates.NewScaleDownCandidatesSortingProcessor(scaleDownCandidatesComparers))
			cp.Register(sddProcessor)
			processors.ScaleDownNodeProcessor = cp
			clusterState := clusterstate.NewClusterStateRegistry(provider, clusterStateConfig, context.LogRecorder, NewBackoff(), nodegroupconfig.NewDefaultNodeGroupConfigProcessor(options.NodeGroupDefaults), processors.AsyncNodeGroupStateChecker)
			processors.ScaleStateNotifier.Register(clusterState)

			sdPlanner, sdActuator := newScaleDownPlannerAndActuator(&context, processors, clusterState, nil)
			suOrchestrator := orchestrator.New()
			suOrchestrator.Initialize(&context, processors, clusterState, newEstimatorBuilder(), taints.TaintConfig{})

			autoscaler := &StaticAutoscaler{
				AutoscalingContext:    &context,
				clusterStateRegistry:  clusterState,
				lastScaleUpTime:       time.Now(),
				lastScaleDownFailTime: time.Now(),
				scaleDownPlanner:      sdPlanner,
				scaleDownActuator:     sdActuator,
				scaleUpOrchestrator:   suOrchestrator,
				processors:            processors,
				loopStartNotifier:     loopstart.NewObserversList(nil),
				processorCallbacks:    processorCallbacks,
				initialized:           true,
			}

			tc.beforeTest(processors)

			provider.AddNode("ng1", n1)
			provider.AddNode("ng2", n2)
			ng1.SetTargetSize(1)
			ng2.SetTargetSize(1)

			// Mark unneeded nodes.
			readyNodeLister.SetNodes([]*apiv1.Node{n1, n2})
			allNodeLister.SetNodes([]*apiv1.Node{n1, n2})
			allPodListerMock.On("List").Return([]*apiv1.Pod{p1}, nil).Twice()
			daemonSetListerMock.On("List", labels.Everything()).Return([]*appsv1.DaemonSet{}, nil).Once()
			podDisruptionBudgetListerMock.On("List").Return([]*policyv1.PodDisruptionBudget{}, nil).Once()

			err = autoscaler.RunOnce(time.Now())
			assert.NoError(t, err)
			mock.AssertExpectationsForObjects(t, allPodListerMock,
				podDisruptionBudgetListerMock, daemonSetListerMock, onScaleUpMock, onScaleDownMock)

			// Scale down nodegroup
			readyNodeLister.SetNodes([]*apiv1.Node{n1, n2})
			allNodeLister.SetNodes([]*apiv1.Node{n1, n2})
			allPodListerMock.On("List").Return([]*apiv1.Pod{p1}, nil).Times(3)
			daemonSetListerMock.On("List", labels.Everything()).Return([]*appsv1.DaemonSet{}, nil).Once()
			podDisruptionBudgetListerMock.On("List").Return([]*policyv1.PodDisruptionBudget{}, nil)
			onScaleDownMock.On("ScaleDown", tc.expectedScaleDownNG, tc.expectedScaleDownNode).Return(nil).Once()

			err = autoscaler.RunOnce(time.Now().Add(config.DefaultScaleDownUnneededTime))
			waitForDeleteToFinish(t, deleteFinished)
			assert.NoError(t, err)
			mock.AssertExpectationsForObjects(t, allPodListerMock,
				podDisruptionBudgetListerMock, daemonSetListerMock, onScaleUpMock, onScaleDownMock)

			tc.afterTest(processors)
		})
	}
}

func TestStaticAutoscalerRunOnceWithAutoprovisionedEnabled(t *testing.T) {
	readyNodeLister := kubernetes.NewTestNodeLister(nil)
	allNodeLister := kubernetes.NewTestNodeLister(nil)
	allPodListerMock := &podListerMock{}
	podDisruptionBudgetListerMock := &podDisruptionBudgetListerMock{}
	daemonSetListerMock := &daemonSetListerMock{}
	onScaleUpMock := &onScaleUpMock{}
	onScaleDownMock := &onScaleDownMock{}
	onNodeGroupCreateMock := &onNodeGroupCreateMock{}
	onNodeGroupDeleteMock := &onNodeGroupDeleteMock{}
	nodeGroupManager := &MockAutoprovisioningNodeGroupManager{t, 0}
	nodeGroupListProcessor := &MockAutoprovisioningNodeGroupListProcessor{t}
	deleteFinished := make(chan bool, 1)

	n1 := BuildTestNode("n1", 100, 1000)
	SetNodeReadyState(n1, true, time.Now())
	n2 := BuildTestNode("n2", 1000, 1000)
	SetNodeReadyState(n2, true, time.Now())

	p1 := BuildTestPod("p1", 100, 100)
	p1.Spec.NodeName = "n1"
	p2 := BuildTestPod("p2", 600, 100, MarkUnschedulable())

	tn1 := BuildTestNode("tn1", 100, 1000)
	SetNodeReadyState(tn1, true, time.Now())
	tni1 := schedulerframework.NewNodeInfo()
	tni1.SetNode(tn1)
	tn2 := BuildTestNode("tn2", 1000, 1000)
	SetNodeReadyState(tn2, true, time.Now())
	tni2 := schedulerframework.NewNodeInfo()
	tni2.SetNode(tn2)
	tn3 := BuildTestNode("tn3", 100, 1000)
	SetNodeReadyState(tn2, true, time.Now())
	tni3 := schedulerframework.NewNodeInfo()
	tni3.SetNode(tn3)

	provider := testprovider.NewTestAutoprovisioningCloudProvider(
		func(id string, delta int) error {
			return onScaleUpMock.ScaleUp(id, delta)
		}, func(id string, name string) error {
			ret := onScaleDownMock.ScaleDown(id, name)
			deleteFinished <- true
			return ret
		}, func(id string) error {
			return onNodeGroupCreateMock.Create(id)
		}, func(id string) error {
			return onNodeGroupDeleteMock.Delete(id)
		},
		[]string{"TN1", "TN2"}, map[string]*schedulerframework.NodeInfo{"TN1": tni1, "TN2": tni2, "ng1": tni3})
	provider.AddNodeGroup("ng1", 1, 10, 1)
	provider.AddAutoprovisionedNodeGroup("autoprovisioned-TN1", 0, 10, 0, "TN1")
	autoprovisionedTN1 := reflect.ValueOf(provider.GetNodeGroup("autoprovisioned-TN1")).Interface().(*testprovider.TestNodeGroup)
	assert.NotNil(t, autoprovisionedTN1)
	provider.AddNode("ng1,", n1)
	assert.NotNil(t, provider)

	// Create context with mocked lister registry.
	options := config.AutoscalingOptions{
		NodeGroupDefaults: config.NodeGroupAutoscalingOptions{
			ScaleDownUnneededTime:         time.Minute,
			ScaleDownUnreadyTime:          time.Minute,
			ScaleDownUtilizationThreshold: 0.5,
			MaxNodeProvisionTime:          10 * time.Second,
		},
		EstimatorName:                    estimator.BinpackingEstimatorName,
		ScaleDownEnabled:                 true,
		MaxNodesTotal:                    100,
		MaxCoresTotal:                    100,
		MaxMemoryTotal:                   100000,
		NodeAutoprovisioningEnabled:      true,
		MaxAutoprovisionedNodeGroupCount: 10,
	}
	processorCallbacks := newStaticAutoscalerProcessorCallbacks()

	context, err := NewScaleTestAutoscalingContext(options, &fake.Clientset{}, nil, provider, processorCallbacks, nil)
	assert.NoError(t, err)

	setUpScaleDownActuator(&context, options)

	processors := NewTestProcessors(&context)
	processors.NodeGroupManager = nodeGroupManager
	processors.NodeGroupListProcessor = nodeGroupListProcessor

	listerRegistry := kube_util.NewListerRegistry(allNodeLister, readyNodeLister, allPodListerMock,
		podDisruptionBudgetListerMock, daemonSetListerMock,
		nil, nil, nil, nil)
	context.ListerRegistry = listerRegistry

	clusterStateConfig := clusterstate.ClusterStateRegistryConfig{
		OkTotalUnreadyCount: 0,
	}
	clusterState := clusterstate.NewClusterStateRegistry(provider, clusterStateConfig, context.LogRecorder, NewBackoff(), nodegroupconfig.NewDefaultNodeGroupConfigProcessor(options.NodeGroupDefaults), processors.AsyncNodeGroupStateChecker)

	sdPlanner, sdActuator := newScaleDownPlannerAndActuator(&context, processors, clusterState, nil)
	suOrchestrator := orchestrator.New()
	suOrchestrator.Initialize(&context, processors, clusterState, newEstimatorBuilder(), taints.TaintConfig{})

	autoscaler := &StaticAutoscaler{
		AutoscalingContext:    &context,
		clusterStateRegistry:  clusterState,
		lastScaleUpTime:       time.Now(),
		lastScaleDownFailTime: time.Now(),
		scaleDownPlanner:      sdPlanner,
		scaleDownActuator:     sdActuator,
		scaleUpOrchestrator:   suOrchestrator,
		processors:            processors,
		loopStartNotifier:     loopstart.NewObserversList(nil),
		processorCallbacks:    processorCallbacks,
		initialized:           true,
	}

	// Scale up.
	readyNodeLister.SetNodes([]*apiv1.Node{n1})
	allNodeLister.SetNodes([]*apiv1.Node{n1})
	allPodListerMock.On("List").Return([]*apiv1.Pod{p1, p2}, nil).Twice()
	podDisruptionBudgetListerMock.On("List").Return([]*policyv1.PodDisruptionBudget{}, nil).Once()
	daemonSetListerMock.On("List", labels.Everything()).Return([]*appsv1.DaemonSet{}, nil).Once()
	onNodeGroupCreateMock.On("Create", "autoprovisioned-TN2").Return(nil).Once()
	onScaleUpMock.On("ScaleUp", "autoprovisioned-TN2", 1).Return(nil).Once()

	err = autoscaler.RunOnce(time.Now().Add(time.Hour))
	assert.NoError(t, err)
	mock.AssertExpectationsForObjects(t, allPodListerMock,
		podDisruptionBudgetListerMock, daemonSetListerMock, onScaleUpMock, onScaleDownMock)

	// Fix target size.
	autoprovisionedTN1.SetTargetSize(0)

	// Remove autoprovisioned node group and mark unneeded nodes.
	readyNodeLister.SetNodes([]*apiv1.Node{n1, n2})
	allNodeLister.SetNodes([]*apiv1.Node{n1, n2})
	allPodListerMock.On("List").Return([]*apiv1.Pod{p1}, nil).Twice()
	podDisruptionBudgetListerMock.On("List").Return([]*policyv1.PodDisruptionBudget{}, nil).Once()
	daemonSetListerMock.On("List", labels.Everything()).Return([]*appsv1.DaemonSet{}, nil).Once()
	onNodeGroupDeleteMock.On("Delete", "autoprovisioned-TN1").Return(nil).Once()

	provider.AddAutoprovisionedNodeGroup("autoprovisioned-TN2", 0, 10, 1, "TN1")
	provider.AddNode("autoprovisioned-TN2", n2)

	err = autoscaler.RunOnce(time.Now().Add(1 * time.Hour))
	assert.NoError(t, err)
	mock.AssertExpectationsForObjects(t, allPodListerMock,
		podDisruptionBudgetListerMock, daemonSetListerMock, onScaleUpMock, onScaleDownMock)

	// Scale down.
	readyNodeLister.SetNodes([]*apiv1.Node{n1, n2})
	allNodeLister.SetNodes([]*apiv1.Node{n1, n2})
	allPodListerMock.On("List").Return([]*apiv1.Pod{p1}, nil).Times(3)
	podDisruptionBudgetListerMock.On("List").Return([]*policyv1.PodDisruptionBudget{}, nil).Once()
	daemonSetListerMock.On("List", labels.Everything()).Return([]*appsv1.DaemonSet{}, nil).Once()
	onNodeGroupDeleteMock.On("Delete", "autoprovisioned-"+
		"TN1").Return(nil).Once()
	onScaleDownMock.On("ScaleDown", "autoprovisioned-TN2", "n2").Return(nil).Once()

	err = autoscaler.RunOnce(time.Now().Add(2 * time.Hour))
	waitForDeleteToFinish(t, deleteFinished)
	assert.NoError(t, err)
	mock.AssertExpectationsForObjects(t, allPodListerMock,
		podDisruptionBudgetListerMock, daemonSetListerMock, onScaleUpMock, onScaleDownMock)
}

func TestStaticAutoscalerRunOnceWithALongUnregisteredNode(t *testing.T) {
	readyNodeLister := kubernetes.NewTestNodeLister(nil)
	allNodeLister := kubernetes.NewTestNodeLister(nil)
	allPodListerMock := &podListerMock{}
	podDisruptionBudgetListerMock := &podDisruptionBudgetListerMock{}
	daemonSetListerMock := &daemonSetListerMock{}
	onScaleUpMock := &onScaleUpMock{}
	onScaleDownMock := &onScaleDownMock{}
	deleteFinished := make(chan bool, 1)

	now := time.Now()
	later := now.Add(1 * time.Minute)

	n1 := BuildTestNode("n1", 1000, 1000)
	SetNodeReadyState(n1, true, time.Now())
	n2 := BuildTestNode("n2", 1000, 1000)
	SetNodeReadyState(n2, true, time.Now())

	p1 := BuildTestPod("p1", 600, 100)
	p1.Spec.NodeName = "n1"
	p2 := BuildTestPod("p2", 600, 100, MarkUnschedulable())

	provider := testprovider.NewTestCloudProvider(
		func(id string, delta int) error {
			return onScaleUpMock.ScaleUp(id, delta)
		}, func(id string, name string) error {
			ret := onScaleDownMock.ScaleDown(id, name)
			deleteFinished <- true
			return ret
		})
	provider.AddNodeGroup("ng1", 2, 10, 2)
	provider.AddNode("ng1", n1)

	// broken node, that will be just hanging out there during
	// the test (it can't be removed since that would validate group min size)
	brokenNode := BuildTestNode("broken", 1000, 1000)
	provider.AddNode("ng1", brokenNode)

	ng1 := reflect.ValueOf(provider.GetNodeGroup("ng1")).Interface().(*testprovider.TestNodeGroup)
	assert.NotNil(t, ng1)
	assert.NotNil(t, provider)

	// Create context with mocked lister registry.
	options := config.AutoscalingOptions{
		NodeGroupDefaults: config.NodeGroupAutoscalingOptions{
			ScaleDownUnneededTime:         time.Minute,
			ScaleDownUnreadyTime:          time.Minute,
			ScaleDownUtilizationThreshold: 0.5,
			MaxNodeProvisionTime:          10 * time.Second,
		},
		EstimatorName:    estimator.BinpackingEstimatorName,
		ScaleDownEnabled: true,
		MaxNodesTotal:    10,
		MaxCoresTotal:    10,
		MaxMemoryTotal:   100000,
	}
	processorCallbacks := newStaticAutoscalerProcessorCallbacks()

	context, err := NewScaleTestAutoscalingContext(options, &fake.Clientset{}, nil, provider, processorCallbacks, nil)
	assert.NoError(t, err)

	setUpScaleDownActuator(&context, options)

	listerRegistry := kube_util.NewListerRegistry(allNodeLister, readyNodeLister, allPodListerMock,
		podDisruptionBudgetListerMock, daemonSetListerMock,
		nil, nil, nil, nil)
	context.ListerRegistry = listerRegistry

	clusterStateConfig := clusterstate.ClusterStateRegistryConfig{
		OkTotalUnreadyCount: 1,
	}
	processors := NewTestProcessors(&context)

	clusterState := clusterstate.NewClusterStateRegistry(provider, clusterStateConfig, context.LogRecorder, NewBackoff(), nodegroupconfig.NewDefaultNodeGroupConfigProcessor(options.NodeGroupDefaults), processors.AsyncNodeGroupStateChecker)
	// broken node detected as unregistered

	nodes := []*apiv1.Node{n1}
	// nodeInfos, _ := getNodeInfosForGroups(nodes, provider, listerRegistry, []*appsv1.DaemonSet{}, context.PredicateChecker)
	clusterState.UpdateNodes(nodes, nil, now)

	// broken node failed to register in time
	clusterState.UpdateNodes(nodes, nil, later)

	sdPlanner, sdActuator := newScaleDownPlannerAndActuator(&context, processors, clusterState, nil)
	suOrchestrator := orchestrator.New()
	suOrchestrator.Initialize(&context, processors, clusterState, newEstimatorBuilder(), taints.TaintConfig{})

	autoscaler := &StaticAutoscaler{
		AutoscalingContext:    &context,
		clusterStateRegistry:  clusterState,
		lastScaleUpTime:       time.Now(),
		lastScaleDownFailTime: time.Now(),
		scaleDownPlanner:      sdPlanner,
		scaleDownActuator:     sdActuator,
		scaleUpOrchestrator:   suOrchestrator,
		processors:            processors,
		loopStartNotifier:     loopstart.NewObserversList(nil),
		processorCallbacks:    processorCallbacks,
	}

	// Scale up.
	readyNodeLister.SetNodes([]*apiv1.Node{n1})
	allNodeLister.SetNodes([]*apiv1.Node{n1})
	allPodListerMock.On("List").Return([]*apiv1.Pod{p1, p2}, nil).Twice()
	daemonSetListerMock.On("List", labels.Everything()).Return([]*appsv1.DaemonSet{}, nil).Once()
	podDisruptionBudgetListerMock.On("List").Return([]*policyv1.PodDisruptionBudget{}, nil).Once()
	onScaleUpMock.On("ScaleUp", "ng1", 1).Return(nil).Once()

	err = autoscaler.RunOnce(later.Add(time.Hour))
	assert.NoError(t, err)
	mock.AssertExpectationsForObjects(t, allPodListerMock,
		podDisruptionBudgetListerMock, daemonSetListerMock, onScaleUpMock, onScaleDownMock)

	// Remove broken node after going over min size
	provider.AddNode("ng1", n2)
	ng1.SetTargetSize(3)

	readyNodeLister.SetNodes([]*apiv1.Node{n1, n2})
	allNodeLister.SetNodes([]*apiv1.Node{n1, n2})
	allPodListerMock.On("List").Return([]*apiv1.Pod{p1, p2}, nil).Twice()
	onScaleDownMock.On("ScaleDown", "ng1", "broken").Return(nil).Once()
	daemonSetListerMock.On("List", labels.Everything()).Return([]*appsv1.DaemonSet{}, nil).Once()
	podDisruptionBudgetListerMock.On("List").Return([]*policyv1.PodDisruptionBudget{}, nil).Once()

	err = autoscaler.RunOnce(later.Add(2 * time.Hour))
	waitForDeleteToFinish(t, deleteFinished)
	assert.NoError(t, err)
	mock.AssertExpectationsForObjects(t, allPodListerMock,
		podDisruptionBudgetListerMock, daemonSetListerMock, onScaleUpMock, onScaleDownMock)
}

func TestStaticAutoscalerRunOncePodsWithPriorities(t *testing.T) {
	readyNodeLister := kubernetes.NewTestNodeLister(nil)
	allNodeLister := kubernetes.NewTestNodeLister(nil)
	allPodListerMock := &podListerMock{}
	podDisruptionBudgetListerMock := &podDisruptionBudgetListerMock{}
	daemonSetListerMock := &daemonSetListerMock{}
	onScaleUpMock := &onScaleUpMock{}
	onScaleDownMock := &onScaleDownMock{}
	deleteFinished := make(chan bool, 1)

	n1 := BuildTestNode("n1", 100, 1000)
	SetNodeReadyState(n1, true, time.Now())
	n2 := BuildTestNode("n2", 1000, 1000)
	SetNodeReadyState(n2, true, time.Now())
	n3 := BuildTestNode("n3", 1000, 1000)
	SetNodeReadyState(n3, true, time.Now())

	// shared owner reference
	ownerRef := GenerateOwnerReferences("rs", "ReplicaSet", "extensions/v1beta1", "")
	var priority100 int32 = 100
	var priority1 int32 = 1

	p1 := BuildTestPod("p1", 40, 0)
	p1.OwnerReferences = ownerRef
	p1.Spec.NodeName = "n1"
	p1.Spec.Priority = &priority1

	p2 := BuildTestPod("p2", 400, 0)
	p2.OwnerReferences = ownerRef
	p2.Spec.NodeName = "n2"
	p2.Spec.Priority = &priority1

	p3 := BuildTestPod("p3", 400, 0)
	p3.OwnerReferences = ownerRef
	p3.Spec.NodeName = "n2"
	p3.Spec.Priority = &priority100

	p4 := BuildTestPod("p4", 500, 0, MarkUnschedulable())
	p4.OwnerReferences = ownerRef
	p4.Spec.Priority = &priority100

	p5 := BuildTestPod("p5", 800, 0, MarkUnschedulable())
	p5.OwnerReferences = ownerRef
	p5.Spec.Priority = &priority100
	p5.Status.NominatedNodeName = "n3"

	p6 := BuildTestPod("p6", 1000, 0, MarkUnschedulable())
	p6.OwnerReferences = ownerRef
	p6.Spec.Priority = &priority100

	provider := testprovider.NewTestCloudProvider(
		func(id string, delta int) error {
			return onScaleUpMock.ScaleUp(id, delta)
		}, func(id string, name string) error {
			ret := onScaleDownMock.ScaleDown(id, name)
			deleteFinished <- true
			return ret
		})
	provider.AddNodeGroup("ng1", 0, 10, 1)
	provider.AddNodeGroup("ng2", 0, 10, 2)
	provider.AddNode("ng1", n1)
	provider.AddNode("ng2", n2)
	provider.AddNode("ng2", n3)
	assert.NotNil(t, provider)
	ng2 := reflect.ValueOf(provider.GetNodeGroup("ng2")).Interface().(*testprovider.TestNodeGroup)
	assert.NotNil(t, ng2)

	// Create context with mocked lister registry.
	options := config.AutoscalingOptions{
		NodeGroupDefaults: config.NodeGroupAutoscalingOptions{
			ScaleDownUnneededTime:         time.Minute,
			ScaleDownUtilizationThreshold: 0.5,
			ScaleDownUnreadyTime:          time.Minute,
			MaxNodeProvisionTime:          10 * time.Second,
		},
		EstimatorName:                estimator.BinpackingEstimatorName,
		ScaleDownEnabled:             true,
		MaxNodesTotal:                10,
		MaxCoresTotal:                10,
		MaxMemoryTotal:               100000,
		ExpendablePodsPriorityCutoff: 10,
		NodeDeletionBatcherInterval:  0 * time.Second,
	}
	processorCallbacks := newStaticAutoscalerProcessorCallbacks()

	context, err := NewScaleTestAutoscalingContext(options, &fake.Clientset{}, nil, provider, processorCallbacks, nil)
	assert.NoError(t, err)

	setUpScaleDownActuator(&context, options)

	listerRegistry := kube_util.NewListerRegistry(allNodeLister, readyNodeLister, allPodListerMock,
		podDisruptionBudgetListerMock, daemonSetListerMock,
		nil, nil, nil, nil)
	context.ListerRegistry = listerRegistry

	clusterStateConfig := clusterstate.ClusterStateRegistryConfig{
		OkTotalUnreadyCount: 1,
	}

	processors := NewTestProcessors(&context)
	clusterState := clusterstate.NewClusterStateRegistry(provider, clusterStateConfig, context.LogRecorder, NewBackoff(), nodegroupconfig.NewDefaultNodeGroupConfigProcessor(options.NodeGroupDefaults), processors.AsyncNodeGroupStateChecker)
	sdPlanner, sdActuator := newScaleDownPlannerAndActuator(&context, processors, clusterState, nil)
	suOrchestrator := orchestrator.New()
	suOrchestrator.Initialize(&context, processors, clusterState, newEstimatorBuilder(), taints.TaintConfig{})

	autoscaler := &StaticAutoscaler{
		AutoscalingContext:    &context,
		clusterStateRegistry:  clusterState,
		lastScaleUpTime:       time.Now(),
		lastScaleDownFailTime: time.Now(),
		scaleDownPlanner:      sdPlanner,
		scaleDownActuator:     sdActuator,
		scaleUpOrchestrator:   suOrchestrator,
		processors:            processors,
		loopStartNotifier:     loopstart.NewObserversList(nil),
		processorCallbacks:    processorCallbacks,
	}

	// Scale up
	readyNodeLister.SetNodes([]*apiv1.Node{n1, n2, n3})
	allNodeLister.SetNodes([]*apiv1.Node{n1, n2, n3})
	allPodListerMock.On("List").Return([]*apiv1.Pod{p1, p2, p3, p4, p5, p6}, nil).Twice()
	daemonSetListerMock.On("List", labels.Everything()).Return([]*appsv1.DaemonSet{}, nil).Once()
	podDisruptionBudgetListerMock.On("List").Return([]*policyv1.PodDisruptionBudget{}, nil).Once()
	onScaleUpMock.On("ScaleUp", "ng2", 1).Return(nil).Once()

	err = autoscaler.RunOnce(time.Now())
	assert.NoError(t, err)
	mock.AssertExpectationsForObjects(t, allPodListerMock,
		podDisruptionBudgetListerMock, daemonSetListerMock, onScaleUpMock, onScaleDownMock)

	// Mark unneeded nodes.
	readyNodeLister.SetNodes([]*apiv1.Node{n1, n2, n3})
	allNodeLister.SetNodes([]*apiv1.Node{n1, n2, n3})
	allPodListerMock.On("List").Return([]*apiv1.Pod{p1, p2, p3, p4, p5}, nil).Twice()
	daemonSetListerMock.On("List", labels.Everything()).Return([]*appsv1.DaemonSet{}, nil).Once()
	podDisruptionBudgetListerMock.On("List").Return([]*policyv1.PodDisruptionBudget{}, nil).Once()

	ng2.SetTargetSize(2)

	err = autoscaler.RunOnce(time.Now().Add(2 * time.Hour))
	assert.NoError(t, err)
	mock.AssertExpectationsForObjects(t, allPodListerMock,
		podDisruptionBudgetListerMock, daemonSetListerMock, onScaleUpMock, onScaleDownMock)

	// Scale down.
	readyNodeLister.SetNodes([]*apiv1.Node{n1, n2, n3})
	allNodeLister.SetNodes([]*apiv1.Node{n1, n2, n3})
	allPodListerMock.On("List").Return([]*apiv1.Pod{p1, p2, p3, p4, p5}, nil).Times(3)
	daemonSetListerMock.On("List", labels.Everything()).Return([]*appsv1.DaemonSet{}, nil).Once()
	podDisruptionBudgetListerMock.On("List").Return([]*policyv1.PodDisruptionBudget{}, nil).Once()
	onScaleDownMock.On("ScaleDown", "ng1", "n1").Return(nil).Once()

	p4.Spec.NodeName = "n2"

	err = autoscaler.RunOnce(time.Now().Add(3 * time.Hour))
	waitForDeleteToFinish(t, deleteFinished)
	assert.NoError(t, err)
	mock.AssertExpectationsForObjects(t, allPodListerMock,
		podDisruptionBudgetListerMock, daemonSetListerMock, onScaleUpMock, onScaleDownMock)
}

func TestStaticAutoscalerRunOnceWithFilteringOnBinPackingEstimator(t *testing.T) {
	readyNodeLister := kubernetes.NewTestNodeLister(nil)
	allNodeLister := kubernetes.NewTestNodeLister(nil)
	allPodListerMock := &podListerMock{}
	podDisruptionBudgetListerMock := &podDisruptionBudgetListerMock{}
	daemonSetListerMock := &daemonSetListerMock{}
	onScaleUpMock := &onScaleUpMock{}
	onScaleDownMock := &onScaleDownMock{}

	n1 := BuildTestNode("n1", 2000, 1000)
	SetNodeReadyState(n1, true, time.Now())
	n2 := BuildTestNode("n2", 2000, 1000)
	SetNodeReadyState(n2, true, time.Now())

	// shared owner reference
	ownerRef := GenerateOwnerReferences("rs", "ReplicaSet", "extensions/v1beta1", "")

	p1 := BuildTestPod("p1", 1400, 0)
	p1.OwnerReferences = ownerRef
	p3 := BuildTestPod("p3", 1400, 0)
	p3.Spec.NodeName = "n1"
	p3.OwnerReferences = ownerRef
	p4 := BuildTestPod("p4", 1400, 0)
	p4.Spec.NodeName = "n2"
	p4.OwnerReferences = ownerRef

	provider := testprovider.NewTestCloudProvider(
		func(id string, delta int) error {
			return onScaleUpMock.ScaleUp(id, delta)
		}, func(id string, name string) error {
			return onScaleDownMock.ScaleDown(id, name)
		})
	provider.AddNodeGroup("ng1", 0, 10, 2)
	provider.AddNode("ng1", n1)

	provider.AddNodeGroup("ng2", 0, 10, 1)
	provider.AddNode("ng2", n2)

	assert.NotNil(t, provider)

	// Create context with mocked lister registry.
	options := config.AutoscalingOptions{
		NodeGroupDefaults: config.NodeGroupAutoscalingOptions{
			ScaleDownUtilizationThreshold: 0.5,
			MaxNodeProvisionTime:          10 * time.Second,
		},
		EstimatorName:                estimator.BinpackingEstimatorName,
		ScaleDownEnabled:             false,
		MaxNodesTotal:                10,
		MaxCoresTotal:                10,
		MaxMemoryTotal:               100000,
		ExpendablePodsPriorityCutoff: 10,
	}
	processorCallbacks := newStaticAutoscalerProcessorCallbacks()

	context, err := NewScaleTestAutoscalingContext(options, &fake.Clientset{}, nil, provider, processorCallbacks, nil)
	assert.NoError(t, err)

	setUpScaleDownActuator(&context, options)

	listerRegistry := kube_util.NewListerRegistry(allNodeLister, readyNodeLister, allPodListerMock,
		podDisruptionBudgetListerMock, daemonSetListerMock,
		nil, nil, nil, nil)
	context.ListerRegistry = listerRegistry

	clusterStateConfig := clusterstate.ClusterStateRegistryConfig{
		OkTotalUnreadyCount: 1,
	}

	processors := NewTestProcessors(&context)
	clusterState := clusterstate.NewClusterStateRegistry(provider, clusterStateConfig, context.LogRecorder, NewBackoff(), nodegroupconfig.NewDefaultNodeGroupConfigProcessor(options.NodeGroupDefaults), processors.AsyncNodeGroupStateChecker)
	sdPlanner, sdActuator := newScaleDownPlannerAndActuator(&context, processors, clusterState, nil)

	autoscaler := &StaticAutoscaler{
		AutoscalingContext:    &context,
		clusterStateRegistry:  clusterState,
		lastScaleUpTime:       time.Now(),
		lastScaleDownFailTime: time.Now(),
		scaleDownPlanner:      sdPlanner,
		scaleDownActuator:     sdActuator,
		processors:            processors,
		loopStartNotifier:     loopstart.NewObserversList(nil),
		processorCallbacks:    processorCallbacks,
	}

	// Scale up
	readyNodeLister.SetNodes([]*apiv1.Node{n1, n2})
	allNodeLister.SetNodes([]*apiv1.Node{n1, n2})
	allPodListerMock.On("List").Return([]*apiv1.Pod{p1, p3, p4}, nil).Twice()
	daemonSetListerMock.On("List", labels.Everything()).Return([]*appsv1.DaemonSet{}, nil).Once()
	podDisruptionBudgetListerMock.On("List").Return([]*policyv1.PodDisruptionBudget{}, nil).Once()

	err = autoscaler.RunOnce(time.Now())
	assert.NoError(t, err)

	mock.AssertExpectationsForObjects(t, allPodListerMock,
		podDisruptionBudgetListerMock, daemonSetListerMock, onScaleUpMock, onScaleDownMock)
}

func TestStaticAutoscalerRunOnceWithFilteringOnUpcomingNodesEnabledNoScaleUp(t *testing.T) {
	readyNodeLister := kubernetes.NewTestNodeLister(nil)
	allNodeLister := kubernetes.NewTestNodeLister(nil)
	allPodListerMock := &podListerMock{}
	podDisruptionBudgetListerMock := &podDisruptionBudgetListerMock{}
	daemonSetListerMock := &daemonSetListerMock{}
	onScaleUpMock := &onScaleUpMock{}
	onScaleDownMock := &onScaleDownMock{}

	n2 := BuildTestNode("n2", 2000, 1000)
	SetNodeReadyState(n2, true, time.Now())
	n3 := BuildTestNode("n3", 2000, 1000)
	SetNodeReadyState(n3, true, time.Now())

	// shared owner reference
	ownerRef := GenerateOwnerReferences("rs", "ReplicaSet", "extensions/v1beta1", "")

	p1 := BuildTestPod("p1", 1400, 0)
	p1.OwnerReferences = ownerRef
	p2 := BuildTestPod("p2", 1400, 0)
	p2.Spec.NodeName = "n2"
	p2.OwnerReferences = ownerRef
	p3 := BuildTestPod("p3", 1400, 0)
	p3.Spec.NodeName = "n3"
	p3.OwnerReferences = ownerRef

	provider := testprovider.NewTestCloudProvider(
		func(id string, delta int) error {
			return onScaleUpMock.ScaleUp(id, delta)
		}, func(id string, name string) error {
			return onScaleDownMock.ScaleDown(id, name)
		})
	provider.AddNodeGroup("ng1", 0, 10, 2)
	provider.AddNode("ng1", n2)

	provider.AddNodeGroup("ng2", 0, 10, 1)
	provider.AddNode("ng2", n3)

	assert.NotNil(t, provider)

	// Create context with mocked lister registry.
	options := config.AutoscalingOptions{
		NodeGroupDefaults: config.NodeGroupAutoscalingOptions{
			ScaleDownUtilizationThreshold: 0.5,
			MaxNodeProvisionTime:          10 * time.Second,
		},
		EstimatorName:                estimator.BinpackingEstimatorName,
		ScaleDownEnabled:             false,
		MaxNodesTotal:                10,
		MaxCoresTotal:                10,
		MaxMemoryTotal:               100000,
		ExpendablePodsPriorityCutoff: 10,
	}
	processorCallbacks := newStaticAutoscalerProcessorCallbacks()

	context, err := NewScaleTestAutoscalingContext(options, &fake.Clientset{}, nil, provider, processorCallbacks, nil)
	assert.NoError(t, err)

	setUpScaleDownActuator(&context, options)

	listerRegistry := kube_util.NewListerRegistry(allNodeLister, readyNodeLister, allPodListerMock,
		podDisruptionBudgetListerMock, daemonSetListerMock,
		nil, nil, nil, nil)
	context.ListerRegistry = listerRegistry

	clusterStateConfig := clusterstate.ClusterStateRegistryConfig{
		OkTotalUnreadyCount: 1,
	}

	processors := NewTestProcessors(&context)
	clusterState := clusterstate.NewClusterStateRegistry(provider, clusterStateConfig, context.LogRecorder, NewBackoff(), nodegroupconfig.NewDefaultNodeGroupConfigProcessor(options.NodeGroupDefaults), processors.AsyncNodeGroupStateChecker)
	sdPlanner, sdActuator := newScaleDownPlannerAndActuator(&context, processors, clusterState, nil)

	autoscaler := &StaticAutoscaler{
		AutoscalingContext:    &context,
		clusterStateRegistry:  clusterState,
		lastScaleUpTime:       time.Now(),
		lastScaleDownFailTime: time.Now(),
		scaleDownPlanner:      sdPlanner,
		scaleDownActuator:     sdActuator,
		processors:            processors,
		loopStartNotifier:     loopstart.NewObserversList(nil),
		processorCallbacks:    processorCallbacks,
	}

	// Scale up
	readyNodeLister.SetNodes([]*apiv1.Node{n2, n3})
	allNodeLister.SetNodes([]*apiv1.Node{n2, n3})
	allPodListerMock.On("List").Return([]*apiv1.Pod{p1, p2, p3}, nil).Twice()
	daemonSetListerMock.On("List", labels.Everything()).Return([]*appsv1.DaemonSet{}, nil).Once()
	podDisruptionBudgetListerMock.On("List").Return([]*policyv1.PodDisruptionBudget{}, nil).Once()

	err = autoscaler.RunOnce(time.Now())
	assert.NoError(t, err)

	mock.AssertExpectationsForObjects(t, allPodListerMock,
		podDisruptionBudgetListerMock, daemonSetListerMock, onScaleUpMock, onScaleDownMock)
}

// We should not touch taints from unselected node groups.
func TestStaticAutoscalerRunOnceWithUnselectedNodeGroups(t *testing.T) {
	n1 := BuildTestNode("n1", 1000, 1000)
	n1.Spec.Taints = append(n1.Spec.Taints, apiv1.Taint{
		Key:    taints.DeletionCandidateTaint,
		Value:  fmt.Sprint(time.Now().Unix()),
		Effect: apiv1.TaintEffectPreferNoSchedule,
	})
	SetNodeReadyState(n1, true, time.Now())
	n2 := BuildTestNode("n2", 1000, 1000)
	n2.Spec.Taints = append(n2.Spec.Taints, apiv1.Taint{
		Key:    taints.DeletionCandidateTaint,
		Value:  fmt.Sprint(time.Now().Unix()),
		Effect: apiv1.TaintEffectPreferNoSchedule,
	})
	SetNodeReadyState(n2, true, time.Now())

	p1 := BuildTestPod("p1", 600, 100)
	p1.Spec.NodeName = n1.Name

	// set minimal cloud provider where only ng1 is defined as selected node group
	provider := testprovider.NewTestCloudProvider(nil, nil)
	provider.AddNodeGroup("ng1", 1, 10, 1)
	provider.AddNode("ng1", n1)
	assert.NotNil(t, provider)

	tests := map[string]struct {
		node           *apiv1.Node
		pods           []*apiv1.Pod
		expectedTaints []apiv1.Taint
	}{
		"Node from selected node groups can get their deletion candidate taints removed": {
			node:           n1,
			pods:           []*apiv1.Pod{p1},
			expectedTaints: []apiv1.Taint{},
		},
		"Node from non-selected node groups should keep their deletion candidate taints": {
			node:           n2,
			pods:           nil,
			expectedTaints: n2.Spec.Taints,
		},
	}

	for name, test := range tests {
		// prevent issues with scoping, we should be able to get rid of that with Go 1.22
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			// Create fake listers for the generated nodes, nothing returned by the rest (but the ones used in the tested path have to be defined).
			readyNodeLister := kubernetes.NewTestNodeLister([]*apiv1.Node{test.node})
			allNodeLister := kubernetes.NewTestNodeLister([]*apiv1.Node{test.node})
			allPodListerMock := kubernetes.NewTestPodLister(test.pods)
			daemonSetLister, err := kubernetes.NewTestDaemonSetLister(nil)
			assert.NoError(t, err)
			listerRegistry := kube_util.NewListerRegistry(allNodeLister, readyNodeLister, allPodListerMock,
				kubernetes.NewTestPodDisruptionBudgetLister(nil), daemonSetLister,
				nil, nil, nil, nil)

			// Create context with minimal autoscalingOptions that guarantee we reach the tested logic.
			autoscalingOptions := config.AutoscalingOptions{
				ScaleDownEnabled:      true,
				MaxBulkSoftTaintCount: 10,
				MaxBulkSoftTaintTime:  3 * time.Second,
			}
			processorCallbacks := newStaticAutoscalerProcessorCallbacks()
			clientset := fake.NewSimpleClientset(test.node)
			context, err := NewScaleTestAutoscalingContext(autoscalingOptions, clientset, listerRegistry, provider, processorCallbacks, nil)
			assert.NoError(t, err)

			// Create CSR with unhealthy cluster protection effectively disabled, to guarantee we reach the tested logic.
			clusterStateConfig := clusterstate.ClusterStateRegistryConfig{
				OkTotalUnreadyCount: 1,
			}
			processors := NewTestProcessors(&context)

			clusterState := clusterstate.NewClusterStateRegistry(provider, clusterStateConfig, context.LogRecorder, NewBackoff(), nodegroupconfig.NewDefaultNodeGroupConfigProcessor(autoscalingOptions.NodeGroupDefaults), processors.AsyncNodeGroupStateChecker)

			// Setting the Actuator is necessary for testing any scale-down logic, it shouldn't have anything to do in this test.
			sdActuator := actuation.NewActuator(&context, clusterState, deletiontracker.NewNodeDeletionTracker(0*time.Second), options.NodeDeleteOptions{}, nil, processors.NodeGroupConfigProcessor)
			context.ScaleDownActuator = sdActuator

			// Fake planner that keeps track of the scale-down candidates passed to UpdateClusterState.
			sdPlanner := &candidateTrackingFakePlanner{}

			autoscaler := &StaticAutoscaler{
				AutoscalingContext:   &context,
				clusterStateRegistry: clusterState,
				scaleDownPlanner:     sdPlanner,
				scaleDownActuator:    sdActuator,
				processors:           processors,
				loopStartNotifier:    loopstart.NewObserversList(nil),
				processorCallbacks:   processorCallbacks,
			}

			err = autoscaler.RunOnce(time.Now().Add(5 * time.Hour))
			assert.NoError(t, err)
			newNode, err := clientset.CoreV1().Nodes().Get(stdcontext.TODO(), test.node.Name, metav1.GetOptions{})
			assert.NoError(t, err)
			assert.Equal(t, test.expectedTaints, newNode.Spec.Taints)
		})
	}
}

func TestStaticAutoscalerRunOnceWithBypassedSchedulers(t *testing.T) {
	bypassedScheduler := "bypassed-scheduler"
	nonBypassedScheduler := "non-bypassed-scheduler"
	options := config.AutoscalingOptions{
		NodeGroupDefaults: config.NodeGroupAutoscalingOptions{
			ScaleDownUnneededTime:         time.Minute,
			ScaleDownUnreadyTime:          time.Minute,
			ScaleDownUtilizationThreshold: 0.5,
			MaxNodeProvisionTime:          10 * time.Second,
		},
		EstimatorName:    estimator.BinpackingEstimatorName,
		ScaleDownEnabled: true,
		MaxNodesTotal:    10,
		MaxCoresTotal:    10,
		MaxMemoryTotal:   100000,
		BypassedSchedulers: scheduler.GetBypassedSchedulersMap([]string{
			apiv1.DefaultSchedulerName,
			bypassedScheduler,
		}),
	}
	now := time.Now()

	n1 := BuildTestNode("n1", 1000, 1000)
	SetNodeReadyState(n1, true, now)

	ngs := []*nodeGroup{{
		name:  "ng1",
		min:   1,
		max:   10,
		nodes: []*apiv1.Node{n1},
	}}

	p1 := BuildTestPod("p1", 600, 100)
	p1.Spec.NodeName = "n1"
	p2 := BuildTestPod("p2", 100, 100, AddSchedulerName(bypassedScheduler))
	p3 := BuildTestPod("p3", 600, 100)                                      // Not yet processed by scheduler, default scheduler is ignored
	p4 := BuildTestPod("p4", 600, 100, AddSchedulerName(bypassedScheduler)) // non-default scheduler & ignored, expects a scale-up
	p5 := BuildTestPod("p5", 600, 100, AddSchedulerName(nonBypassedScheduler))

	testSetupConfig := &autoscalerSetupConfig{
		autoscalingOptions:  options,
		nodeGroups:          ngs,
		nodeStateUpdateTime: now,
		mocks:               newCommonMocks(),
		clusterStateConfig: clusterstate.ClusterStateRegistryConfig{
			OkTotalUnreadyCount: 1,
		},
	}

	testCases := map[string]struct {
		setupConfig     *autoscalerSetupConfig
		pods            []*apiv1.Pod
		expectedScaleUp *scaleCall
	}{
		"Unprocessed pod with bypassed scheduler doesn't cause a scale-up when there's capacity": {
			pods:        []*apiv1.Pod{p1, p2},
			setupConfig: testSetupConfig,
		},
		"Unprocessed pod with bypassed scheduler causes a scale-up when there's no capacity - Default Scheduler": {
			pods: []*apiv1.Pod{p1, p3},
			expectedScaleUp: &scaleCall{
				ng:    "ng1",
				delta: 1,
			},
			setupConfig: testSetupConfig,
		},
		"Unprocessed pod with bypassed scheduler causes a scale-up when there's no capacity - Non-default Scheduler": {
			pods:        []*apiv1.Pod{p1, p4},
			setupConfig: testSetupConfig,
			expectedScaleUp: &scaleCall{
				ng:    "ng1",
				delta: 1,
			},
		},
		"Unprocessed pod with non-bypassed scheduler doesn't cause a scale-up when there's no capacity": {
			pods:        []*apiv1.Pod{p1, p5},
			setupConfig: testSetupConfig,
		},
	}

	for tcName, tc := range testCases {
		t.Run(tcName, func(t *testing.T) {
			autoscaler, err := setupAutoscaler(tc.setupConfig)
			assert.NoError(t, err)

			tc.setupConfig.mocks.readyNodeLister.SetNodes([]*apiv1.Node{n1})
			tc.setupConfig.mocks.allNodeLister.SetNodes([]*apiv1.Node{n1})
			tc.setupConfig.mocks.allPodLister.On("List").Return(tc.pods, nil).Twice()
			tc.setupConfig.mocks.daemonSetLister.On("List", labels.Everything()).Return([]*appsv1.DaemonSet{}, nil).Once()
			tc.setupConfig.mocks.podDisruptionBudgetLister.On("List").Return([]*policyv1.PodDisruptionBudget{}, nil).Once()
			if tc.expectedScaleUp != nil {
				tc.setupConfig.mocks.onScaleUp.On("ScaleUp", tc.expectedScaleUp.ng, tc.expectedScaleUp.delta).Return(nil).Once()
			}
			err = autoscaler.RunOnce(now.Add(time.Hour))
			assert.NoError(t, err)
			mock.AssertExpectationsForObjects(t, tc.setupConfig.mocks.allPodLister,
				tc.setupConfig.mocks.podDisruptionBudgetLister, tc.setupConfig.mocks.daemonSetLister, tc.setupConfig.mocks.onScaleUp, tc.setupConfig.mocks.onScaleDown)
		})
	}

}

func TestStaticAutoscalerInstanceCreationErrors(t *testing.T) {
	// setup
	provider := &mockprovider.CloudProvider{}

	// Create context with mocked lister registry.
	options := config.AutoscalingOptions{
		NodeGroupDefaults: config.NodeGroupAutoscalingOptions{
			ScaleDownUnneededTime:         time.Minute,
			ScaleDownUnreadyTime:          time.Minute,
			ScaleDownUtilizationThreshold: 0.5,
			MaxNodeProvisionTime:          10 * time.Second,
		},
		EstimatorName:                estimator.BinpackingEstimatorName,
		ScaleDownEnabled:             true,
		MaxNodesTotal:                10,
		MaxCoresTotal:                10,
		MaxMemoryTotal:               100000,
		ExpendablePodsPriorityCutoff: 10,
	}
	processorCallbacks := newStaticAutoscalerProcessorCallbacks()

	context, err := NewScaleTestAutoscalingContext(options, &fake.Clientset{}, nil, provider, processorCallbacks, nil)
	assert.NoError(t, err)

	clusterStateConfig := clusterstate.ClusterStateRegistryConfig{
		OkTotalUnreadyCount: 1,
	}

	nodeGroupConfigProcessor := nodegroupconfig.NewDefaultNodeGroupConfigProcessor(options.NodeGroupDefaults)
	asyncNodeGroupStateChecker := asyncnodegroups.NewDefaultAsyncNodeGroupStateChecker()
	clusterState := clusterstate.NewClusterStateRegistry(provider, clusterStateConfig, context.LogRecorder, NewBackoff(), nodeGroupConfigProcessor, asyncNodeGroupStateChecker)
	autoscaler := &StaticAutoscaler{
		AutoscalingContext:    &context,
		clusterStateRegistry:  clusterState,
		lastScaleUpTime:       time.Now(),
		lastScaleDownFailTime: time.Now(),
		processorCallbacks:    processorCallbacks,
	}

	nodeGroupA := &mockprovider.NodeGroup{}
	nodeGroupB := &mockprovider.NodeGroup{}

	// Three nodes with out-of-resources errors
	nodeGroupA.On("Exist").Return(true)
	nodeGroupA.On("Autoprovisioned").Return(false)
	nodeGroupA.On("TargetSize").Return(5, nil)
	nodeGroupA.On("Id").Return("A")
	nodeGroupA.On("DeleteNodes", mock.Anything).Return(nil)
	nodeGroupA.On("GetOptions", options.NodeGroupDefaults).Return(&options.NodeGroupDefaults, nil)
	nodeGroupA.On("Nodes").Return([]cloudprovider.Instance{
		{
			Id: "A1",
			Status: &cloudprovider.InstanceStatus{
				State: cloudprovider.InstanceRunning,
			},
		},
		{
			Id: "A2",
			Status: &cloudprovider.InstanceStatus{
				State: cloudprovider.InstanceCreating,
			},
		},
		{
			Id: "A3",
			Status: &cloudprovider.InstanceStatus{
				State: cloudprovider.InstanceCreating,
				ErrorInfo: &cloudprovider.InstanceErrorInfo{
					ErrorClass: cloudprovider.OutOfResourcesErrorClass,
					ErrorCode:  "RESOURCE_POOL_EXHAUSTED",
				},
			},
		},
		{
			Id: "A4",
			Status: &cloudprovider.InstanceStatus{
				State: cloudprovider.InstanceCreating,
				ErrorInfo: &cloudprovider.InstanceErrorInfo{
					ErrorClass: cloudprovider.OutOfResourcesErrorClass,
					ErrorCode:  "RESOURCE_POOL_EXHAUSTED",
				},
			},
		},
		{
			Id: "A5",
			Status: &cloudprovider.InstanceStatus{
				State: cloudprovider.InstanceCreating,
				ErrorInfo: &cloudprovider.InstanceErrorInfo{
					ErrorClass: cloudprovider.OutOfResourcesErrorClass,
					ErrorCode:  "QUOTA",
				},
			},
		},
		{
			Id: "A6",
			Status: &cloudprovider.InstanceStatus{
				State: cloudprovider.InstanceCreating,
				ErrorInfo: &cloudprovider.InstanceErrorInfo{
					ErrorClass: cloudprovider.OtherErrorClass,
					ErrorCode:  "OTHER",
				},
			},
		},
	}, nil).Twice()

	nodeGroupB.On("Exist").Return(true)
	nodeGroupB.On("Autoprovisioned").Return(false)
	nodeGroupB.On("TargetSize").Return(5, nil)
	nodeGroupB.On("Id").Return("B")
	nodeGroupB.On("DeleteNodes", mock.Anything).Return(nil)
	nodeGroupB.On("GetOptions", options.NodeGroupDefaults).Return(&options.NodeGroupDefaults, nil)
	nodeGroupB.On("Nodes").Return([]cloudprovider.Instance{
		{
			Id: "B1",
			Status: &cloudprovider.InstanceStatus{
				State: cloudprovider.InstanceRunning,
			},
		},
	}, nil)

	provider.On("NodeGroups").Return([]cloudprovider.NodeGroup{nodeGroupA})
	provider.On("NodeGroupForNode", mock.Anything).Return(
		func(node *apiv1.Node) cloudprovider.NodeGroup {
			if strings.HasPrefix(node.Spec.ProviderID, "A") {
				return nodeGroupA
			}
			if strings.HasPrefix(node.Spec.ProviderID, "B") {
				return nodeGroupB
			}
			return nil
		}, nil)
	provider.On("HasInstance", mock.Anything).Return(
		func(node *apiv1.Node) bool {
			return false
		}, nil)

	now := time.Now()

	clusterState.RefreshCloudProviderNodeInstancesCache()
	// propagate nodes info in cluster state
	clusterState.UpdateNodes([]*apiv1.Node{}, nil, now)

	// delete nodes with create errors
	autoscaler.deleteCreatedNodesWithErrors()

	// nodes should be deleted
	expectedDeleteCalls := 1
	nodeGroupA.AssertNumberOfCalls(t, "DeleteNodes", expectedDeleteCalls)

	// check delete was called on correct nodes
	nodeGroupA.AssertCalled(t, "DeleteNodes", mock.MatchedBy(
		func(nodes []*apiv1.Node) bool {
			if len(nodes) != 4 {
				return false
			}
			names := make(map[string]bool)
			for _, node := range nodes {
				names[node.Spec.ProviderID] = true
			}
			return names["A3"] && names["A4"] && names["A5"] && names["A6"]
		}))

	// TODO assert that scaleup was failed (separately for QUOTA and RESOURCE_POOL_EXHAUSTED)

	clusterState.RefreshCloudProviderNodeInstancesCache()

	// propagate nodes info in cluster state again
	// no changes in what provider returns
	clusterState.UpdateNodes([]*apiv1.Node{}, nil, now)

	// delete nodes with create errors
	autoscaler.deleteCreatedNodesWithErrors()

	// nodes should be deleted again
	expectedDeleteCalls += 1
	nodeGroupA.AssertNumberOfCalls(t, "DeleteNodes", expectedDeleteCalls)

	nodeGroupA.AssertCalled(t, "DeleteNodes", mock.MatchedBy(
		func(nodes []*apiv1.Node) bool {
			if len(nodes) != 4 {
				return false
			}
			names := make(map[string]bool)
			for _, node := range nodes {
				names[node.Spec.ProviderID] = true
			}
			return names["A3"] && names["A4"] && names["A5"] && names["A6"]
		}))

	// TODO assert that scaleup is not failed again

	// restub node group A so nodes are no longer reporting errors
	nodeGroupA.On("Nodes").Return([]cloudprovider.Instance{
		{
			Id: "A1",
			Status: &cloudprovider.InstanceStatus{
				State: cloudprovider.InstanceRunning,
			},
		},
		{
			Id: "A2",
			Status: &cloudprovider.InstanceStatus{
				State: cloudprovider.InstanceCreating,
			},
		},
		{
			Id: "A3",
			Status: &cloudprovider.InstanceStatus{
				State: cloudprovider.InstanceDeleting,
			},
		},
		{
			Id: "A4",
			Status: &cloudprovider.InstanceStatus{
				State: cloudprovider.InstanceDeleting,
			},
		},
		{
			Id: "A5",
			Status: &cloudprovider.InstanceStatus{
				State: cloudprovider.InstanceDeleting,
			},
		},
		{
			Id: "A6",
			Status: &cloudprovider.InstanceStatus{
				State: cloudprovider.InstanceDeleting,
			},
		},
	}, nil)

	clusterState.RefreshCloudProviderNodeInstancesCache()

	// update cluster state
	clusterState.UpdateNodes([]*apiv1.Node{}, nil, now)

	// delete nodes with create errors
	autoscaler.deleteCreatedNodesWithErrors()

	// we expect no more Delete Nodes, don't increase expectedDeleteCalls
	nodeGroupA.AssertNumberOfCalls(t, "DeleteNodes", expectedDeleteCalls)

	// failed node not included by NodeGroupForNode
	nodeGroupC := &mockprovider.NodeGroup{}
	nodeGroupC.On("Exist").Return(true)
	nodeGroupC.On("Autoprovisioned").Return(false)
	nodeGroupC.On("TargetSize").Return(1, nil)
	nodeGroupC.On("Id").Return("C")
	nodeGroupC.On("DeleteNodes", mock.Anything).Return(nil)
	nodeGroupC.On("GetOptions", options.NodeGroupDefaults).Return(&options.NodeGroupDefaults, nil)
	nodeGroupC.On("Nodes").Return([]cloudprovider.Instance{
		{
			Id: "C1",
			Status: &cloudprovider.InstanceStatus{
				State: cloudprovider.InstanceCreating,
				ErrorInfo: &cloudprovider.InstanceErrorInfo{
					ErrorClass: cloudprovider.OutOfResourcesErrorClass,
					ErrorCode:  "QUOTA",
				},
			},
		},
	}, nil)
	provider = &mockprovider.CloudProvider{}
	provider.On("NodeGroups").Return([]cloudprovider.NodeGroup{nodeGroupC})
	provider.On("NodeGroupForNode", mock.Anything).Return(nil, nil)
	provider.On("HasInstance", mock.Anything).Return(
		func(node *apiv1.Node) bool {
			return false
		}, nil)

	clusterState = clusterstate.NewClusterStateRegistry(provider, clusterStateConfig, context.LogRecorder, NewBackoff(), nodeGroupConfigProcessor, asyncNodeGroupStateChecker)
	clusterState.RefreshCloudProviderNodeInstancesCache()
	autoscaler.clusterStateRegistry = clusterState

	// update cluster state
	clusterState.UpdateNodes([]*apiv1.Node{}, nil, time.Now())

	// No nodes are deleted when failed nodes don't have matching node groups
	autoscaler.deleteCreatedNodesWithErrors()
	nodeGroupC.AssertNumberOfCalls(t, "DeleteNodes", 0)

	nodeGroupAtomic := &mockprovider.NodeGroup{}
	nodeGroupAtomic.On("Exist").Return(true)
	nodeGroupAtomic.On("Autoprovisioned").Return(false)
	nodeGroupAtomic.On("TargetSize").Return(3, nil)
	nodeGroupAtomic.On("Id").Return("D")
	nodeGroupAtomic.On("DeleteNodes", mock.Anything).Return(nil)
	nodeGroupAtomic.On("GetOptions", options.NodeGroupDefaults).Return(
		&config.NodeGroupAutoscalingOptions{
			ZeroOrMaxNodeScaling: true,
		}, nil)
	nodeGroupAtomic.On("Nodes").Return([]cloudprovider.Instance{
		{
			Id: "D1",
			Status: &cloudprovider.InstanceStatus{
				State: cloudprovider.InstanceRunning,
			},
		},
		{
			Id: "D2",
			Status: &cloudprovider.InstanceStatus{
				State: cloudprovider.InstanceRunning,
			},
		},
		{
			Id: "D3",
			Status: &cloudprovider.InstanceStatus{
				State: cloudprovider.InstanceCreating,
				ErrorInfo: &cloudprovider.InstanceErrorInfo{
					ErrorClass: cloudprovider.OtherErrorClass,
					ErrorCode:  "OTHER",
				},
			},
		},
	}, nil).Twice()
	provider = &mockprovider.CloudProvider{}
	provider.On("NodeGroups").Return([]cloudprovider.NodeGroup{nodeGroupAtomic})
	provider.On("NodeGroupForNode", mock.Anything).Return(
		func(node *apiv1.Node) cloudprovider.NodeGroup {
			if strings.HasPrefix(node.Spec.ProviderID, "D") {
				return nodeGroupAtomic
			}
			return nil
		}, nil).Times(3)

	clusterState = clusterstate.NewClusterStateRegistry(provider, clusterStateConfig, context.LogRecorder, NewBackoff(), nodeGroupConfigProcessor, asyncNodeGroupStateChecker)
	clusterState.RefreshCloudProviderNodeInstancesCache()
	autoscaler.CloudProvider = provider
	autoscaler.clusterStateRegistry = clusterState
	// propagate nodes info in cluster state
	clusterState.UpdateNodes([]*apiv1.Node{}, nil, now)

	// delete nodes with create errors
	autoscaler.deleteCreatedNodesWithErrors()

	nodeGroupAtomic.AssertCalled(t, "DeleteNodes", mock.MatchedBy(
		func(nodes []*apiv1.Node) bool {
			if len(nodes) != 3 {
				return false
			}
			names := make(map[string]bool)
			for _, node := range nodes {
				names[node.Spec.ProviderID] = true
			}
			return names["D1"] && names["D2"] && names["D3"]
		}))

	// Node group with getOptions error gets no deletes.
	nodeGroupError := &mockprovider.NodeGroup{}
	nodeGroupError.On("Exist").Return(true)
	nodeGroupError.On("Autoprovisioned").Return(false)
	nodeGroupError.On("TargetSize").Return(1, nil)
	nodeGroupError.On("Id").Return("E")
	nodeGroupError.On("DeleteNodes", mock.Anything).Return(nil)
	nodeGroupError.On("GetOptions", options.NodeGroupDefaults).Return(nil, fmt.Errorf("Failed to get options"))
	nodeGroupError.On("Nodes").Return([]cloudprovider.Instance{
		{
			Id: "E1",
			Status: &cloudprovider.InstanceStatus{
				State: cloudprovider.InstanceRunning,
			},
		},
		{

			Id: "E2",
			Status: &cloudprovider.InstanceStatus{
				State: cloudprovider.InstanceCreating,
				ErrorInfo: &cloudprovider.InstanceErrorInfo{
					ErrorClass: cloudprovider.OutOfResourcesErrorClass,
					ErrorCode:  "QUOTA",
				},
			},
		},
	}, nil)

	provider = &mockprovider.CloudProvider{}
	provider.On("NodeGroups").Return([]cloudprovider.NodeGroup{nodeGroupError})
	provider.On("NodeGroupForNode", mock.Anything).Return(
		func(node *apiv1.Node) cloudprovider.NodeGroup {
			if strings.HasPrefix(node.Spec.ProviderID, "E") {
				return nodeGroupError
			}
			return nil
		}, nil).Times(2)

	clusterState = clusterstate.NewClusterStateRegistry(provider, clusterStateConfig, context.LogRecorder, NewBackoff(), nodeGroupConfigProcessor, asyncNodeGroupStateChecker)
	clusterState.RefreshCloudProviderNodeInstancesCache()
	autoscaler.CloudProvider = provider
	autoscaler.clusterStateRegistry = clusterState
	// propagate nodes info in cluster state
	clusterState.UpdateNodes([]*apiv1.Node{}, nil, now)

	// delete nodes with create errors
	autoscaler.deleteCreatedNodesWithErrors()

	nodeGroupError.AssertNumberOfCalls(t, "DeleteNodes", 0)
}

type candidateTrackingFakePlanner struct {
	lastCandidateNodes map[string]bool
}

func (f *candidateTrackingFakePlanner) UpdateClusterState(podDestinations, scaleDownCandidates []*apiv1.Node, as scaledown.ActuationStatus, currentTime time.Time) errors.AutoscalerError {
	f.lastCandidateNodes = map[string]bool{}
	for _, node := range scaleDownCandidates {
		f.lastCandidateNodes[node.Name] = true
	}
	return nil
}

func (f *candidateTrackingFakePlanner) CleanUpUnneededNodes() {
}

func (f *candidateTrackingFakePlanner) NodesToDelete(currentTime time.Time) (empty, needDrain []*apiv1.Node) {
	return nil, nil
}

func (f *candidateTrackingFakePlanner) UnneededNodes() []*apiv1.Node {
	return nil
}

func (f *candidateTrackingFakePlanner) UnremovableNodes() []*simulator.UnremovableNode {
	return nil
}

func (f *candidateTrackingFakePlanner) NodeUtilizationMap() map[string]utilization.Info {
	return nil
}

func assertSnapshotNodeCount(t *testing.T, snapshot clustersnapshot.ClusterSnapshot, wantCount int) {
	nodeInfos, err := snapshot.NodeInfos().List()
	assert.NoError(t, err)
	assert.Len(t, nodeInfos, wantCount)
}

func assertNodesNotInSnapshot(t *testing.T, snapshot clustersnapshot.ClusterSnapshot, nodeNames map[string]bool) {
	nodeInfos, err := snapshot.NodeInfos().List()
	assert.NoError(t, err)
	for _, nodeInfo := range nodeInfos {
		assert.NotContains(t, nodeNames, nodeInfo.Node().Name)
	}
}

func assertNodesInSnapshot(t *testing.T, snapshot clustersnapshot.ClusterSnapshot, nodeNames map[string]bool) {
	nodeInfos, err := snapshot.NodeInfos().List()
	assert.NoError(t, err)
	snapshotNodeNames := map[string]bool{}
	for _, nodeInfo := range nodeInfos {
		snapshotNodeNames[nodeInfo.Node().Name] = true
	}
	for nodeName := range nodeNames {
		assert.Contains(t, snapshotNodeNames, nodeName)
	}
}

func TestStaticAutoscalerUpcomingScaleDownCandidates(t *testing.T) {
	startTime := time.Time{}

	// Generate a number of ready and unready nodes created at startTime, spread across multiple node groups.
	provider := testprovider.NewTestCloudProvider(nil, nil)
	allNodeNames := map[string]bool{}
	readyNodeNames := map[string]bool{}
	notReadyNodeNames := map[string]bool{}
	var allNodes []*apiv1.Node
	var readyNodes []*apiv1.Node

	readyNodesCount := 4
	unreadyNodesCount := 2
	nodeGroupCount := 2
	for ngNum := 0; ngNum < nodeGroupCount; ngNum++ {
		ngName := fmt.Sprintf("ng-%d", ngNum)
		provider.AddNodeGroup(ngName, 0, 1000, readyNodesCount+unreadyNodesCount)

		for i := 0; i < readyNodesCount; i++ {
			node := BuildTestNode(fmt.Sprintf("%s-ready-node-%d", ngName, i), 2000, 1000)
			node.CreationTimestamp = metav1.NewTime(startTime)
			SetNodeReadyState(node, true, startTime)
			provider.AddNode(ngName, node)

			allNodes = append(allNodes, node)
			allNodeNames[node.Name] = true

			readyNodes = append(readyNodes, node)
			readyNodeNames[node.Name] = true
		}
		for i := 0; i < unreadyNodesCount; i++ {
			node := BuildTestNode(fmt.Sprintf("%s-unready-node-%d", ngName, i), 2000, 1000)
			node.CreationTimestamp = metav1.NewTime(startTime)
			SetNodeReadyState(node, false, startTime)
			provider.AddNode(ngName, node)

			allNodes = append(allNodes, node)
			allNodeNames[node.Name] = true

			notReadyNodeNames[node.Name] = true
		}
	}

	// Create fake listers for the generated nodes, nothing returned by the rest (but the ones used in the tested path have to be defined).
	allNodeLister := kubernetes.NewTestNodeLister(allNodes)
	readyNodeLister := kubernetes.NewTestNodeLister(readyNodes)
	daemonSetLister, err := kubernetes.NewTestDaemonSetLister(nil)
	assert.NoError(t, err)
	listerRegistry := kube_util.NewListerRegistry(allNodeLister, readyNodeLister,
		kubernetes.NewTestPodLister(nil),
		kubernetes.NewTestPodDisruptionBudgetLister(nil), daemonSetLister, nil, nil, nil, nil)

	// Create context with minimal autoscalingOptions that guarantee we reach the tested logic.
	// We're only testing the input to UpdateClusterState which should be called whenever scale-down is enabled, other autoscalingOptions shouldn't matter.
	autoscalingOptions := config.AutoscalingOptions{ScaleDownEnabled: true}
	processorCallbacks := newStaticAutoscalerProcessorCallbacks()
	ctx, err := NewScaleTestAutoscalingContext(autoscalingOptions, &fake.Clientset{}, listerRegistry, provider, processorCallbacks, nil)
	assert.NoError(t, err)

	processors := NewTestProcessors(&ctx)

	// Create CSR with unhealthy cluster protection effectively disabled, to guarantee we reach the tested logic.
	csrConfig := clusterstate.ClusterStateRegistryConfig{OkTotalUnreadyCount: nodeGroupCount * unreadyNodesCount}
	csr := clusterstate.NewClusterStateRegistry(provider, csrConfig, ctx.LogRecorder, NewBackoff(), nodegroupconfig.NewDefaultNodeGroupConfigProcessor(config.NodeGroupAutoscalingOptions{MaxNodeProvisionTime: 15 * time.Minute}), processors.AsyncNodeGroupStateChecker)

	// Setting the Actuator is necessary for testing any scale-down logic, it shouldn't have anything to do in this test.
	actuator := actuation.NewActuator(&ctx, csr, deletiontracker.NewNodeDeletionTracker(0*time.Second), options.NodeDeleteOptions{}, nil, NewTestProcessors(&ctx).NodeGroupConfigProcessor)
	ctx.ScaleDownActuator = actuator

	// Fake planner that keeps track of the scale-down candidates passed to UpdateClusterState.
	planner := &candidateTrackingFakePlanner{}

	autoscaler := &StaticAutoscaler{
		AutoscalingContext:   &ctx,
		clusterStateRegistry: csr,
		scaleDownActuator:    actuator,
		scaleDownPlanner:     planner,
		processors:           processors,
		loopStartNotifier:    loopstart.NewObserversList(nil),
		processorCallbacks:   processorCallbacks,
	}

	// RunOnce run right when the nodes are created. Ready nodes should be passed as scale-down candidates, unready nodes should be classified as
	// NotStarted and not passed as scale-down candidates (or inserted into the cluster snapshot). The fake upcoming nodes also shouldn't be passed,
	// but they should be inserted into the snapshot.
	err = autoscaler.RunOnce(startTime)
	assert.NoError(t, err)
	assert.Equal(t, readyNodeNames, planner.lastCandidateNodes)
	assertNodesInSnapshot(t, autoscaler.ClusterSnapshot, readyNodeNames)
	assertNodesNotInSnapshot(t, autoscaler.ClusterSnapshot, notReadyNodeNames)
	assertSnapshotNodeCount(t, autoscaler.ClusterSnapshot, len(allNodeNames)) // Ready nodes + fake upcoming copies for unready nodes.

	// RunOnce run in the last moment when unready nodes are still classified as NotStarted - assertions the same as above.
	err = autoscaler.RunOnce(startTime.Add(clusterstate.MaxNodeStartupTime).Add(-time.Second))
	assert.NoError(t, err)
	assert.Equal(t, readyNodeNames, planner.lastCandidateNodes)
	assertNodesInSnapshot(t, autoscaler.ClusterSnapshot, readyNodeNames)
	assertNodesNotInSnapshot(t, autoscaler.ClusterSnapshot, notReadyNodeNames)
	assertSnapshotNodeCount(t, autoscaler.ClusterSnapshot, len(allNodeNames)) // Ready nodes + fake upcoming copies for unready nodes.

	// RunOnce run in the first moment when unready nodes exceed the startup threshold, stop being classified as NotStarted, and start being classified
	// Unready instead. The unready nodes should be passed as scale-down candidates at this point, and inserted into the snapshot. Fake upcoming
	// nodes should no longer be inserted.
	err = autoscaler.RunOnce(startTime.Add(clusterstate.MaxNodeStartupTime).Add(time.Second))
	assert.Equal(t, allNodeNames, planner.lastCandidateNodes)
	assertNodesInSnapshot(t, autoscaler.ClusterSnapshot, allNodeNames)
	assertSnapshotNodeCount(t, autoscaler.ClusterSnapshot, len(allNodeNames)) // Ready nodes + actual unready nodes.
}

func TestStaticAutoscalerProcessorCallbacks(t *testing.T) {
	processorCallbacks := newStaticAutoscalerProcessorCallbacks()
	assert.Equal(t, false, processorCallbacks.disableScaleDownForLoop)
	assert.Equal(t, 0, len(processorCallbacks.extraValues))

	processorCallbacks.DisableScaleDownForLoop()
	assert.Equal(t, true, processorCallbacks.disableScaleDownForLoop)
	processorCallbacks.reset()
	assert.Equal(t, false, processorCallbacks.disableScaleDownForLoop)

	_, found := processorCallbacks.GetExtraValue("blah")
	assert.False(t, found)

	processorCallbacks.SetExtraValue("blah", "some value")
	value, found := processorCallbacks.GetExtraValue("blah")
	assert.True(t, found)
	assert.Equal(t, "some value", value)

	processorCallbacks.reset()
	assert.Equal(t, 0, len(processorCallbacks.extraValues))
	_, found = processorCallbacks.GetExtraValue("blah")
	assert.False(t, found)
}

func TestRemoveFixNodeTargetSize(t *testing.T) {
	sizeChanges := make(chan string, 10)
	now := time.Now()

	ng1_1 := BuildTestNode("ng1-1", 1000, 1000)
	ng1_1.Spec.ProviderID = "ng1-1"
	provider := testprovider.NewTestCloudProvider(func(nodegroup string, delta int) error {
		sizeChanges <- fmt.Sprintf("%s/%d", nodegroup, delta)
		return nil
	}, nil)
	provider.AddNodeGroup("ng1", 1, 10, 3)
	provider.AddNode("ng1", ng1_1)

	fakeClient := &fake.Clientset{}
	fakeLogRecorder, _ := clusterstate_utils.NewStatusMapRecorder(fakeClient, "kube-system", kube_record.NewFakeRecorder(5), false, "my-cool-configmap")

	context := &context.AutoscalingContext{
		AutoscalingOptions: config.AutoscalingOptions{
			NodeGroupDefaults: config.NodeGroupAutoscalingOptions{
				MaxNodeProvisionTime: 45 * time.Minute,
			},
		},
		CloudProvider: provider,
	}

	clusterState := clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{
		MaxTotalUnreadyPercentage: 10,
		OkTotalUnreadyCount:       1,
	}, fakeLogRecorder, NewBackoff(), nodegroupconfig.NewDefaultNodeGroupConfigProcessor(context.AutoscalingOptions.NodeGroupDefaults), asyncnodegroups.NewDefaultAsyncNodeGroupStateChecker())
	err := clusterState.UpdateNodes([]*apiv1.Node{ng1_1}, nil, now.Add(-time.Hour))
	assert.NoError(t, err)

	// Nothing should be fixed. The incorrect size state is not old enough.
	removed, err := fixNodeGroupSize(context, clusterState, now.Add(-50*time.Minute))
	assert.NoError(t, err)
	assert.False(t, removed)

	// Node group should be decreased.
	removed, err = fixNodeGroupSize(context, clusterState, now)
	assert.NoError(t, err)
	assert.True(t, removed)
	change := core_utils.GetStringFromChan(sizeChanges)
	assert.Equal(t, "ng1/-2", change)
}

func TestRemoveOldUnregisteredNodes(t *testing.T) {
	deletedNodes := make(chan string, 10)

	now := time.Now()

	ng1_1 := BuildTestNode("ng1-1", 1000, 1000)
	ng1_1.Spec.ProviderID = "ng1-1"
	ng1_2 := BuildTestNode("ng1-2", 1000, 1000)
	ng1_2.Spec.ProviderID = "ng1-2"
	provider := testprovider.NewTestCloudProvider(nil, func(nodegroup string, node string) error {
		deletedNodes <- fmt.Sprintf("%s/%s", nodegroup, node)
		return nil
	})
	provider.AddNodeGroup("ng1", 1, 10, 2)
	provider.AddNode("ng1", ng1_1)
	provider.AddNode("ng1", ng1_2)

	fakeClient := &fake.Clientset{}
	fakeLogRecorder, _ := clusterstate_utils.NewStatusMapRecorder(fakeClient, "kube-system", kube_record.NewFakeRecorder(5), false, "my-cool-configmap")

	context := &context.AutoscalingContext{
		AutoscalingOptions: config.AutoscalingOptions{
			NodeGroupDefaults: config.NodeGroupAutoscalingOptions{
				MaxNodeProvisionTime: 45 * time.Minute,
			},
		},
		CloudProvider: provider,
	}
	clusterState := clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{
		MaxTotalUnreadyPercentage: 10,
		OkTotalUnreadyCount:       1,
	}, fakeLogRecorder, NewBackoff(), nodegroupconfig.NewDefaultNodeGroupConfigProcessor(context.AutoscalingOptions.NodeGroupDefaults), asyncnodegroups.NewDefaultAsyncNodeGroupStateChecker())
	err := clusterState.UpdateNodes([]*apiv1.Node{ng1_1}, nil, now.Add(-time.Hour))
	assert.NoError(t, err)

	unregisteredNodes := clusterState.GetUnregisteredNodes()
	assert.Equal(t, 1, len(unregisteredNodes))

	autoscaler := &StaticAutoscaler{
		AutoscalingContext:   context,
		clusterStateRegistry: clusterState,
	}

	// Nothing should be removed. The unregistered node is not old enough.
	removed, err := autoscaler.removeOldUnregisteredNodes(unregisteredNodes, context, clusterState, now.Add(-50*time.Minute), fakeLogRecorder)
	assert.NoError(t, err)
	assert.False(t, removed)

	// ng1_2 should be removed.
	removed, err = autoscaler.removeOldUnregisteredNodes(unregisteredNodes, context, clusterState, now, fakeLogRecorder)
	assert.NoError(t, err)
	assert.True(t, removed)
	deletedNode := core_utils.GetStringFromChan(deletedNodes)
	assert.Equal(t, "ng1/ng1-2", deletedNode)
}

func TestRemoveOldUnregisteredNodesAtomic(t *testing.T) {
	deletedNodes := make(chan string, 10)

	now := time.Now()
	provider := testprovider.NewTestCloudProvider(nil, func(nodegroup string, node string) error {
		deletedNodes <- fmt.Sprintf("%s/%s", nodegroup, node)
		return nil
	})
	provider.AddNodeGroupWithCustomOptions("atomic-ng", 0, 10, 10, &config.NodeGroupAutoscalingOptions{
		MaxNodeProvisionTime: 45 * time.Minute,
		ZeroOrMaxNodeScaling: true,
	})
	regNode := BuildTestNode("atomic-ng-0", 1000, 1000)
	regNode.Spec.ProviderID = "atomic-ng-0"
	provider.AddNode("atomic-ng", regNode)
	for i := 1; i < 10; i++ {
		node := BuildTestNode(fmt.Sprintf("atomic-ng-%v", i), 1000, 1000)
		node.Spec.ProviderID = fmt.Sprintf("atomic-ng-%v", i)
		provider.AddNode("atomic-ng", node)
	}

	fakeClient := &fake.Clientset{}
	fakeLogRecorder, _ := clusterstate_utils.NewStatusMapRecorder(fakeClient, "kube-system", kube_record.NewFakeRecorder(5), false, "my-cool-configmap")

	context := &context.AutoscalingContext{
		AutoscalingOptions: config.AutoscalingOptions{
			NodeGroupDefaults: config.NodeGroupAutoscalingOptions{
				MaxNodeProvisionTime: time.Hour,
			},
		},
		CloudProvider: provider,
	}
	clusterState := clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{
		MaxTotalUnreadyPercentage: 10,
		OkTotalUnreadyCount:       1,
	}, fakeLogRecorder, NewBackoff(), nodegroupconfig.NewDefaultNodeGroupConfigProcessor(context.AutoscalingOptions.NodeGroupDefaults), asyncnodegroups.NewDefaultAsyncNodeGroupStateChecker())
	err := clusterState.UpdateNodes([]*apiv1.Node{regNode}, nil, now.Add(-time.Hour))
	assert.NoError(t, err)

	unregisteredNodes := clusterState.GetUnregisteredNodes()
	assert.Equal(t, 9, len(unregisteredNodes))

	autoscaler := &StaticAutoscaler{
		AutoscalingContext:   context,
		clusterStateRegistry: clusterState,
	}

	// Nothing should be removed. The unregistered node is not old enough.
	removed, err := autoscaler.removeOldUnregisteredNodes(unregisteredNodes, context, clusterState, now.Add(-50*time.Minute), fakeLogRecorder)
	assert.NoError(t, err)
	assert.False(t, removed)

	// unregNode is long unregistered, so all of the nodes should be removed due to ZeroOrMaxNodeScaling option
	removed, err = autoscaler.removeOldUnregisteredNodes(unregisteredNodes, context, clusterState, now, fakeLogRecorder)
	assert.NoError(t, err)
	assert.True(t, removed)
	wantNames, deletedNames := []string{}, []string{}
	for i := 0; i < 10; i++ {
		deletedNames = append(deletedNames, core_utils.GetStringFromChan(deletedNodes))
		wantNames = append(wantNames, fmt.Sprintf("atomic-ng/atomic-ng-%v", i))
	}

	assert.ElementsMatch(t, wantNames, deletedNames)
}

func TestSubtractNodes(t *testing.T) {
	ns := make([]*apiv1.Node, 5)
	for i := 0; i < len(ns); i++ {
		ns[i] = BuildTestNode(fmt.Sprintf("n%d", i), 1000, 1000)
	}
	testCases := []struct {
		a []*apiv1.Node
		b []*apiv1.Node
		c []*apiv1.Node
	}{
		{
			a: ns,
			b: nil,
			c: ns,
		},
		{
			a: nil,
			b: ns,
			c: nil,
		},
		{
			a: ns,
			b: []*apiv1.Node{ns[3]},
			c: []*apiv1.Node{ns[0], ns[1], ns[2], ns[4]},
		},
		{
			a: ns,
			b: []*apiv1.Node{ns[0], ns[1], ns[2], ns[4]},
			c: []*apiv1.Node{ns[3]},
		},
		{
			a: []*apiv1.Node{ns[3]},
			b: []*apiv1.Node{ns[0], ns[1], ns[2], ns[4]},
			c: []*apiv1.Node{ns[3]},
		},
	}
	for _, tc := range testCases {
		got := subtractNodes(tc.a, tc.b)
		assert.Equal(t, nodeNames(got), nodeNames(tc.c))

		got = subtractNodesByName(tc.a, nodeNames(tc.b))
		assert.Equal(t, nodeNames(got), nodeNames(tc.c))
	}
}

func TestFilterOutYoungPods(t *testing.T) {
	now := time.Now()
	klog.InitFlags(nil)
	flag.CommandLine.Parse([]string{"--logtostderr=false"})

	p1 := BuildTestPod("p1", 500, 1000)
	p1.CreationTimestamp = metav1.NewTime(now.Add(-1 * time.Minute))
	p2 := BuildTestPod("p2", 500, 1000)
	p2.CreationTimestamp = metav1.NewTime(now.Add(-1 * time.Minute))
	p2.Annotations = map[string]string{
		podScaleUpDelayAnnotationKey: "5m",
	}
	p3 := BuildTestPod("p3", 500, 1000)
	p3.CreationTimestamp = metav1.NewTime(now.Add(-1 * time.Minute))
	p3.Annotations = map[string]string{
		podScaleUpDelayAnnotationKey: "2m",
	}
	p4 := BuildTestPod("p4", 500, 1000)
	p4.CreationTimestamp = metav1.NewTime(now.Add(-1 * time.Minute))
	p4.Annotations = map[string]string{
		podScaleUpDelayAnnotationKey: "error",
	}

	tests := []struct {
		name               string
		newPodScaleUpDelay time.Duration
		runTime            time.Time
		pods               []*apiv1.Pod
		expectedPods       []*apiv1.Pod
		expectedError      string
	}{
		{
			name:               "annotation delayed pod checking now",
			newPodScaleUpDelay: 0,
			runTime:            now,
			pods:               []*apiv1.Pod{p1, p2},
			expectedPods:       []*apiv1.Pod{p1},
		},
		{
			name:               "annotation delayed pod checking after delay",
			newPodScaleUpDelay: 0,
			runTime:            now.Add(5 * time.Minute),
			pods:               []*apiv1.Pod{p1, p2},
			expectedPods:       []*apiv1.Pod{p1, p2},
		},
		{
			name:               "globally delayed pods",
			newPodScaleUpDelay: 5 * time.Minute,
			runTime:            now,
			pods:               []*apiv1.Pod{p1, p2},
			expectedPods:       []*apiv1.Pod(nil),
		},
		{
			name:               "annotation delay smaller than global",
			newPodScaleUpDelay: 5 * time.Minute,
			runTime:            now.Add(2 * time.Minute),
			pods:               []*apiv1.Pod{p1, p3},
			expectedPods:       []*apiv1.Pod(nil),
			expectedError:      "Failed to set pod scale up delay for",
		},
		{
			name:               "annotation delay with error",
			newPodScaleUpDelay: 0,
			runTime:            now,
			pods:               []*apiv1.Pod{p1, p4},
			expectedPods:       []*apiv1.Pod{p1, p4},
			expectedError:      "Failed to parse pod",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context := context.AutoscalingContext{
				AutoscalingOptions: config.AutoscalingOptions{
					NewPodScaleUpDelay: tt.newPodScaleUpDelay,
				},
			}
			autoscaler := &StaticAutoscaler{
				AutoscalingContext: &context,
			}

			var buf bytes.Buffer
			klog.SetOutput(&buf)
			defer func() {
				klog.SetOutput(os.Stderr)
			}()

			actual := autoscaler.filterOutYoungPods(tt.pods, tt.runTime)

			assert.Equal(t, tt.expectedPods, actual)
			if tt.expectedError != "" {
				assert.Contains(t, buf.String(), tt.expectedError)
			}
		})
	}
}

func TestStaticAutoscalerRunOnceInvokesScaleDownStatusProcessor(t *testing.T) {
	options := config.AutoscalingOptions{
		NodeGroupDefaults: config.NodeGroupAutoscalingOptions{
			ScaleDownUnneededTime:         -1 * time.Nanosecond, // enforce immediate scaledown/drain for ready
			ScaleDownUnreadyTime:          -1 * time.Nanosecond, // enforce immediate scaledown/drain for unready
			ScaleDownUtilizationThreshold: 0.5,
			MaxNodeProvisionTime:          10 * time.Second,
		},
		EstimatorName:    estimator.BinpackingEstimatorName,
		ScaleDownEnabled: true,
		MaxNodesTotal:    10,
		MaxCoresTotal:    10,
		MaxMemoryTotal:   100000,
	}
	now := time.Now()
	n1 := BuildTestNode("n1", 1000, 1000)
	SetNodeReadyState(n1, true, now)
	n2 := BuildTestNode("n2", 1000, 1000)
	SetNodeReadyState(n2, true, now)

	underUtilizedPod := BuildTestPod("p1", 20, 20, WithNodeName("n1"))
	utilizedPod := BuildTestPod("p1", 800, 800, WithNodeName("n1"))

	testCases := map[string]struct {
		pods                         []*apiv1.Pod
		nodes                        []*apiv1.Node
		fakeDeletionResults          map[string]status.NodeDeleteResult
		fakeDeletionResultsNodeGroup string
		expectedStatus               *status.ScaleDownStatus
	}{
		"no candidates": {
			pods:  []*apiv1.Pod{utilizedPod},
			nodes: []*apiv1.Node{n1},
			expectedStatus: &status.ScaleDownStatus{
				Result:          status.ScaleDownNoNodeDeleted,
				ScaledDownNodes: []*status.ScaleDownNode{},
				UnremovableNodes: []*status.UnremovableNode{
					{
						Node:        n1,
						BlockingPod: nil,
						Reason:      simulator.NotUnderutilized,
					},
				},
				RemovedNodeGroups:     []cloudprovider.NodeGroup{},
				NodeDeleteResults:     map[string]status.NodeDeleteResult{},
				NodeDeleteResultsAsOf: time.Time{},
			},
		},
		"scaledown": {
			pods:  []*apiv1.Pod{underUtilizedPod},
			nodes: []*apiv1.Node{n1, n2},
			expectedStatus: &status.ScaleDownStatus{
				Result: status.ScaleDownNodeDeleteStarted,
				ScaledDownNodes: []*status.ScaleDownNode{
					{
						Node: n2,
					},
				},
				UnremovableNodes: []*status.UnremovableNode{
					{
						Node:   n1,
						Reason: simulator.BlockedByPod,
						BlockingPod: &drain.BlockingPod{
							Pod:    underUtilizedPod,
							Reason: drain.NotReplicated,
						},
					},
				},
				RemovedNodeGroups:     []cloudprovider.NodeGroup{},
				NodeDeleteResults:     map[string]status.NodeDeleteResult{},
				NodeDeleteResultsAsOf: time.Time{},
			},
		},
		"no candidates, node deleted": {
			pods:  []*apiv1.Pod{utilizedPod},
			nodes: []*apiv1.Node{n1},
			fakeDeletionResults: map[string]status.NodeDeleteResult{"n1": {
				Err:        nil,
				ResultType: status.NodeDeleteOk,
			}},
			fakeDeletionResultsNodeGroup: "ng1",
			expectedStatus: &status.ScaleDownStatus{
				Result:          status.ScaleDownNoNodeDeleted,
				ScaledDownNodes: []*status.ScaleDownNode{},
				UnremovableNodes: []*status.UnremovableNode{
					{
						Node:        n1,
						BlockingPod: nil,
						Reason:      simulator.NotUnderutilized,
					},
				},
				RemovedNodeGroups: []cloudprovider.NodeGroup{},
				NodeDeleteResults: map[string]status.NodeDeleteResult{"n1": {
					Err:        nil,
					ResultType: status.NodeDeleteOk,
				}},
				NodeDeleteResultsAsOf: time.Time{},
			},
		},
	}

	for testName, test := range testCases {
		// prevent issues with scoping, we should be able to get rid of that with Go 1.22
		test := test
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			mocks := newCommonMocks()
			if test.fakeDeletionResults != nil {
				tracker := deletiontracker.NewNodeDeletionTracker(time.Second * 0)
				for node, result := range test.fakeDeletionResults {
					tracker.StartDeletion(test.fakeDeletionResultsNodeGroup, node)
					tracker.EndDeletion(test.fakeDeletionResultsNodeGroup, node, result)
				}

				mocks.nodeDeletionTracker = tracker
			}
			setupConfig := &autoscalerSetupConfig{
				autoscalingOptions: options,
				nodeGroups: []*nodeGroup{{
					name:  "ng1",
					min:   0,
					max:   10,
					nodes: test.nodes,
				}},
				nodeStateUpdateTime: now,
				mocks:               mocks,
				clusterStateConfig: clusterstate.ClusterStateRegistryConfig{
					OkTotalUnreadyCount: 1,
				},
			}
			autoscaler, err := setupAutoscaler(setupConfig)
			assert.NoError(t, err)

			statusProcessor := &scaleDownStatusProcessorMock{}
			autoscaler.processors.ScaleDownStatusProcessor = statusProcessor

			setupConfig.mocks.readyNodeLister.SetNodes(test.nodes)
			setupConfig.mocks.allNodeLister.SetNodes(test.nodes)
			setupConfig.mocks.allPodLister.On("List").Return(test.pods, nil)
			setupConfig.mocks.daemonSetLister.On("List", labels.Everything()).Return([]*appsv1.DaemonSet{}, nil)
			setupConfig.mocks.podDisruptionBudgetLister.On("List").Return([]*policyv1.PodDisruptionBudget{}, nil)
			setupConfig.mocks.onScaleDown.On("ScaleDown", "ng1", "n2").Return(nil).Maybe()

			err = autoscaler.RunOnce(now.Add(time.Hour))
			assert.NoError(t, err)

			assert.Equal(t, statusProcessor.called, 1)
			opts := cmp.Options{
				// These fields are not important for this check and may clutter the whole plot
				cmpopts.IgnoreFields(status.UnremovableNode{}, "NodeGroup", "UtilInfo"),
				cmpopts.IgnoreFields(status.ScaleDownNode{}, "NodeGroup", "UtilInfo"),
				cmpopts.IgnoreFields(status.ScaleDownStatus{}, "NodeDeleteResultsAsOf"),
				cmpopts.EquateEmpty(),
			}
			if diff := cmp.Diff(test.expectedStatus, statusProcessor.scaleDownStatus, opts); diff != "" {
				t.Errorf("ScaleDownStatusProcessor.Process(...): err diff (-want +got):\n%s", diff)
			}

			mock.AssertExpectationsForObjects(t,
				setupConfig.mocks.allPodLister,
				setupConfig.mocks.podDisruptionBudgetLister,
				setupConfig.mocks.daemonSetLister,
				setupConfig.mocks.onScaleUp,
				setupConfig.mocks.onScaleDown,
			)
		})
	}

}

func waitForDeleteToFinish(t *testing.T, deleteFinished <-chan bool) {
	t.Helper()
	select {
	case <-deleteFinished:
		return
	case <-time.After(10 * time.Second):
		t.Fatalf("Node delete not finished")
	}
}

func newScaleDownPlannerAndActuator(ctx *context.AutoscalingContext, p *ca_processors.AutoscalingProcessors, cs *clusterstate.ClusterStateRegistry, nodeDeletionTracker *deletiontracker.NodeDeletionTracker) (scaledown.Planner, scaledown.Actuator) {
	ctx.MaxScaleDownParallelism = 10
	ctx.MaxDrainParallelism = 1
	ctx.NodeDeletionBatcherInterval = 0 * time.Second
	ctx.NodeDeleteDelayAfterTaint = 1 * time.Millisecond
	ctx.ScaleDownSimulationTimeout = 10 * time.Second
	deleteOptions := options.NodeDeleteOptions{
		SkipNodesWithSystemPods:           true,
		SkipNodesWithLocalStorage:         true,
		SkipNodesWithCustomControllerPods: true,
	}

	if nodeDeletionTracker == nil {
		nodeDeletionTracker = deletiontracker.NewNodeDeletionTracker(0 * time.Second)
	}
	planner := planner.New(ctx, p, deleteOptions, nil)
	actuator := actuation.NewActuator(ctx, cs, nodeDeletionTracker, deleteOptions, nil, p.NodeGroupConfigProcessor)
	return planner, actuator
}

func newEstimatorBuilder() estimator.EstimatorBuilder {
	estimatorBuilder, _ := estimator.NewEstimatorBuilder(
		estimator.BinpackingEstimatorName,
		estimator.NewThresholdBasedEstimationLimiter(nil),
		estimator.NewDecreasingPodOrderer(),
		nil,
	)

	return estimatorBuilder
}
