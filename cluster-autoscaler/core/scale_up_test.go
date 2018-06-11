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
	"regexp"
	"strings"
	"testing"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	testprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate/utils"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/estimator"
	"k8s.io/autoscaler/cluster-autoscaler/expander/random"
	ca_processors "k8s.io/autoscaler/cluster-autoscaler/processors"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodegroups"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
	"k8s.io/autoscaler/cluster-autoscaler/utils/units"

	apiv1 "k8s.io/api/core/v1"
	extensionsv1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	core "k8s.io/client-go/testing"
	kube_record "k8s.io/client-go/tools/record"
	schedulercache "k8s.io/kubernetes/pkg/scheduler/cache"

	"github.com/stretchr/testify/assert"
	"k8s.io/autoscaler/cluster-autoscaler/expander"
)

var defaultOptions = context.AutoscalingOptions{
	EstimatorName:  estimator.BinpackingEstimatorName,
	MaxCoresTotal:  config.DefaultMaxClusterCores,
	MaxMemoryTotal: config.DefaultMaxClusterMemory * units.Gigabyte,
	MinCoresTotal:  0,
	MinMemoryTotal: 0,
}

func TestScaleUpOK(t *testing.T) {
	config := &scaleTestConfig{
		nodes: []nodeConfig{
			{"n1", 100, 100, 0, true, "ng1"},
			{"n2", 1000, 1000, 0, true, "ng2"},
		},
		pods: []podConfig{
			{"p1", 80, 0, 0, "n1"},
			{"p2", 800, 0, 0, "n2"},
		},
		extraPods: []podConfig{
			{"p-new", 500, 0, 0, ""},
		},
		scaleUpOptionToChoose: groupSizeChange{groupName: "ng2", sizeChange: 1},
		expectedFinalScaleUp:  groupSizeChange{groupName: "ng2", sizeChange: 1},
		options:               defaultOptions,
	}

	simpleScaleUpTest(t, config)
}

func TestScaleUpMaxCoresLimitHit(t *testing.T) {
	options := defaultOptions
	options.MaxCoresTotal = 9
	config := &scaleTestConfig{
		nodes: []nodeConfig{
			{"n1", 2000, 100, 0, true, "ng1"},
			{"n2", 4000, 1000, 0, true, "ng2"},
		},
		pods: []podConfig{
			{"p1", 1000, 0, 0, "n1"},
			{"p2", 3000, 0, 0, "n2"},
		},
		extraPods: []podConfig{
			{"p-new-1", 2000, 0, 0, ""},
			{"p-new-2", 2000, 0, 0, ""},
		},
		scaleUpOptionToChoose: groupSizeChange{groupName: "ng1", sizeChange: 2},
		expectedFinalScaleUp:  groupSizeChange{groupName: "ng1", sizeChange: 1},
		options:               options,
	}

	simpleScaleUpTest(t, config)
}

const MB = 1024 * 1024

func TestScaleUpMaxMemoryLimitHit(t *testing.T) {
	options := defaultOptions
	options.MaxMemoryTotal = 1300 * MB
	config := &scaleTestConfig{
		nodes: []nodeConfig{
			{"n1", 2000, 100 * MB, 0, true, "ng1"},
			{"n2", 4000, 1000 * MB, 0, true, "ng2"},
		},
		pods: []podConfig{
			{"p1", 1000, 0, 0, "n1"},
			{"p2", 3000, 0, 0, "n2"},
		},
		extraPods: []podConfig{
			{"p-new-1", 2000, 100 * MB, 0, ""},
			{"p-new-2", 2000, 100 * MB, 0, ""},
			{"p-new-3", 2000, 100 * MB, 0, ""},
		},
		scaleUpOptionToChoose: groupSizeChange{groupName: "ng1", sizeChange: 3},
		expectedFinalScaleUp:  groupSizeChange{groupName: "ng1", sizeChange: 2},
		options:               options,
	}

	simpleScaleUpTest(t, config)
}

func TestScaleUpCapToMaxTotalNodesLimit(t *testing.T) {
	options := defaultOptions
	options.MaxNodesTotal = 3
	config := &scaleTestConfig{
		nodes: []nodeConfig{
			{"n1", 2000, 100 * MB, 0, true, "ng1"},
			{"n2", 4000, 1000 * MB, 0, true, "ng2"},
		},
		pods: []podConfig{
			{"p1", 1000, 0, 0, "n1"},
			{"p2", 3000, 0, 0, "n2"},
		},
		extraPods: []podConfig{
			{"p-new-1", 4000, 100 * MB, 0, ""},
			{"p-new-2", 4000, 100 * MB, 0, ""},
			{"p-new-3", 4000, 100 * MB, 0, ""},
		},
		scaleUpOptionToChoose: groupSizeChange{groupName: "ng2", sizeChange: 3},
		expectedFinalScaleUp:  groupSizeChange{groupName: "ng2", sizeChange: 1},
		options:               options,
	}

	simpleScaleUpTest(t, config)
}

func TestWillConsiderGpuAndStandardPoolForPodWhichDoesNotRequireGpu(t *testing.T) {
	options := defaultOptions
	options.MaxNodesTotal = 100
	config := &scaleTestConfig{
		nodes: []nodeConfig{
			{"gpu-node-1", 2000, 1000 * MB, 1, true, "gpu-pool"},
			{"std-node-1", 2000, 1000 * MB, 0, true, "std-pool"},
		},
		pods: []podConfig{
			{"gpu-pod-1", 2000, 1000 * MB, 1, "gpu-node-1"},
			{"std-pod-1", 2000, 1000 * MB, 0, "std-node-1"},
		},
		extraPods: []podConfig{
			{"extra-std-pod", 2000, 1000 * MB, 0, ""},
		},
		expectedScaleUpOptions: []groupSizeChange{
			{groupName: "std-pool", sizeChange: 1},
			{groupName: "gpu-pool", sizeChange: 1},
		},
		scaleUpOptionToChoose: groupSizeChange{groupName: "std-pool", sizeChange: 1},
		expectedFinalScaleUp:  groupSizeChange{groupName: "std-pool", sizeChange: 1},
		options:               options,
	}

	simpleScaleUpTest(t, config)
}

func TestWillConsiderOnlyGpuPoolForPodWhichDoesRequiresGpu(t *testing.T) {
	options := defaultOptions
	options.MaxNodesTotal = 100
	config := &scaleTestConfig{
		nodes: []nodeConfig{
			{"gpu-node-1", 2000, 1000 * MB, 1, true, "gpu-pool"},
			{"std-node-1", 2000, 1000 * MB, 0, true, "std-pool"},
		},
		pods: []podConfig{
			{"gpu-pod-1", 2000, 1000 * MB, 1, "gpu-node-1"},
			{"std-pod-1", 2000, 1000 * MB, 0, "std-node-1"},
		},
		extraPods: []podConfig{
			{"extra-gpu-pod", 2000, 1000 * MB, 1, ""},
		},
		expectedScaleUpOptions: []groupSizeChange{
			{groupName: "gpu-pool", sizeChange: 1},
		},
		scaleUpOptionToChoose: groupSizeChange{groupName: "gpu-pool", sizeChange: 1},
		expectedFinalScaleUp:  groupSizeChange{groupName: "gpu-pool", sizeChange: 1},
		options:               options,
	}

	simpleScaleUpTest(t, config)
}

func TestWillConsiderAllPoolsWhichFitTwoPodsRequiringGpus(t *testing.T) {
	options := defaultOptions
	options.MaxNodesTotal = 100
	config := &scaleTestConfig{
		nodes: []nodeConfig{
			{"gpu-1-node-1", 2000, 1000 * MB, 1, true, "gpu-1-pool"},
			{"gpu-2-node-1", 2000, 1000 * MB, 2, true, "gpu-2-pool"},
			{"gpu-4-node-1", 2000, 1000 * MB, 4, true, "gpu-4-pool"},
			{"std-node-1", 2000, 1000 * MB, 0, true, "std-pool"},
		},
		pods: []podConfig{
			{"gpu-pod-1", 2000, 1000 * MB, 1, "gpu-1-node-1"},
			{"gpu-pod-2", 2000, 1000 * MB, 2, "gpu-2-node-1"},
			{"gpu-pod-3", 2000, 1000 * MB, 4, "gpu-4-node-1"},
			{"std-pod-1", 2000, 1000 * MB, 0, "std-node-1"},
		},
		extraPods: []podConfig{
			{"extra-gpu-pod-1", 1, 1 * MB, 1, ""}, // CPU and mem negligible
			{"extra-gpu-pod-2", 1, 1 * MB, 1, ""}, // CPU and mem negligible
			{"extra-gpu-pod-3", 1, 1 * MB, 1, ""}, // CPU and mem negligible
		},
		expectedScaleUpOptions: []groupSizeChange{
			{groupName: "gpu-1-pool", sizeChange: 3},
			{groupName: "gpu-2-pool", sizeChange: 2},
			{groupName: "gpu-4-pool", sizeChange: 1},
		},
		scaleUpOptionToChoose: groupSizeChange{groupName: "gpu-1-pool", sizeChange: 3},
		expectedFinalScaleUp:  groupSizeChange{groupName: "gpu-1-pool", sizeChange: 3},
		options:               options,
	}

	simpleScaleUpTest(t, config)
}

type assertingStrategy struct {
	initialNodeConfigs     []nodeConfig
	expectedScaleUpOptions []groupSizeChange
	scaleUpOptionToChoose  groupSizeChange
	t                      *testing.T
}

func (s assertingStrategy) BestOption(options []expander.Option, nodeInfo map[string]*schedulercache.NodeInfo) *expander.Option {
	if len(s.expectedScaleUpOptions) > 0 {
		// empty s.expectedScaleUpOptions means we do not want to do assertion on contents of actual scaleUp options

		// precondition check that option to choose is part of expected options
		assert.Contains(s.t, s.expectedScaleUpOptions, s.scaleUpOptionToChoose, "scaleUpOptionToChoose must be present in expectedScaleUpOptions")

		actualScaleUpOptions := expanderOptionsToGroupSizeChanges(options)
		assert.Subset(s.t, actualScaleUpOptions, s.expectedScaleUpOptions,
			"actual %s and expected %s scaleUp options differ",
			actualScaleUpOptions,
			s.expectedScaleUpOptions)
		assert.Equal(s.t, len(actualScaleUpOptions), len(s.expectedScaleUpOptions),
			"actual %s and expected %s scaleUp options differ",
			actualScaleUpOptions,
			s.expectedScaleUpOptions)
	}

	for _, option := range options {
		scaleUpOption := expanderOptionToGroupSizeChange(option)
		if scaleUpOption == s.scaleUpOptionToChoose {
			return &option
		}
	}
	assert.Fail(s.t, "did not find scaleUpOptionToChoose %s", s.scaleUpOptionToChoose)
	return nil
}

func expanderOptionsToGroupSizeChanges(options []expander.Option) []groupSizeChange {
	groupSizeChanges := make([]groupSizeChange, 0, len(options))
	for _, option := range options {
		groupSizeChange := expanderOptionToGroupSizeChange(option)
		groupSizeChanges = append(groupSizeChanges, groupSizeChange)
	}
	return groupSizeChanges
}

func expanderOptionToGroupSizeChange(option expander.Option) groupSizeChange {
	groupName := option.NodeGroup.Id()
	groupSizeIncrement := option.NodeCount
	scaleUpOption := groupSizeChange{groupName, groupSizeIncrement}
	return scaleUpOption
}

func simpleScaleUpTest(t *testing.T, config *scaleTestConfig) {
	expandedGroups := make(chan groupSizeChange, 10)
	fakeClient := &fake.Clientset{}

	groups := make(map[string][]*apiv1.Node)
	nodes := make([]*apiv1.Node, len(config.nodes))
	for i, n := range config.nodes {
		node := BuildTestNode(n.name, n.cpu, n.memory)
		if n.gpu > 0 {
			AddGpusToNode(node, n.gpu)
		}
		SetNodeReadyState(node, n.ready, time.Now())
		nodes[i] = node
		groups[n.group] = append(groups[n.group], node)
	}

	pods := make(map[string][]apiv1.Pod)
	for _, p := range config.pods {
		pod := buildTestPod(p)
		pods[p.node] = append(pods[p.node], *pod)
	}

	fakeClient.Fake.AddReactor("list", "pods", func(action core.Action) (bool, runtime.Object, error) {
		list := action.(core.ListAction)
		fieldstring := list.GetListRestrictions().Fields.String()
		for _, node := range nodes {
			if strings.Contains(fieldstring, node.Name) {
				return true, &apiv1.PodList{Items: pods[node.Name]}, nil
			}
		}
		return true, nil, fmt.Errorf("Failed to list: %v", list)
	})

	provider := testprovider.NewTestCloudProvider(func(nodeGroup string, increase int) error {
		expandedGroups <- groupSizeChange{groupName: nodeGroup, sizeChange: increase}
		return nil
	}, nil)

	for name, nodesInGroup := range groups {
		provider.AddNodeGroup(name, 1, 10, len(nodesInGroup))
		for _, n := range nodesInGroup {
			provider.AddNode(name, n)
		}
	}

	resourceLimiter := cloudprovider.NewResourceLimiter(
		map[string]int64{cloudprovider.ResourceNameCores: config.options.MinCoresTotal, cloudprovider.ResourceNameMemory: config.options.MinMemoryTotal},
		map[string]int64{cloudprovider.ResourceNameCores: config.options.MaxCoresTotal, cloudprovider.ResourceNameMemory: config.options.MaxMemoryTotal})
	provider.SetResourceLimiter(resourceLimiter)

	assert.NotNil(t, provider)

	fakeRecorder := kube_record.NewFakeRecorder(5)
	fakeLogRecorder, _ := utils.NewStatusMapRecorder(fakeClient, "kube-system", kube_record.NewFakeRecorder(5), false)
	clusterState := clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{}, fakeLogRecorder)

	clusterState.UpdateNodes(nodes, time.Now())

	context := &context.AutoscalingContext{
		AutoscalingOptions: config.options,
		PredicateChecker:   simulator.NewTestPredicateChecker(),
		CloudProvider:      provider,
		ClientSet:          fakeClient,
		Recorder:           fakeRecorder,
		ExpanderStrategy: assertingStrategy{
			initialNodeConfigs:     config.nodes,
			expectedScaleUpOptions: config.expectedScaleUpOptions,
			scaleUpOptionToChoose:  config.scaleUpOptionToChoose,
			t: t},
		LogRecorder: fakeLogRecorder,
	}

	extraPods := make([]*apiv1.Pod, len(config.extraPods))
	for i, p := range config.extraPods {
		pod := buildTestPod(p)
		extraPods[i] = pod
	}

	processors := ca_processors.TestProcessors()

	status, err := ScaleUp(context, processors, clusterState, extraPods, nodes, []*extensionsv1.DaemonSet{})
	processors.ScaleUpStatusProcessor.Process(context, status)
	assert.NoError(t, err)
	assert.True(t, status.ScaledUp)

	expandedGroup := getGroupSizeChangeFromChan(expandedGroups)
	assert.NotNil(t, expandedGroup, "Expected scale up event")
	assert.Equal(t, config.expectedFinalScaleUp, *expandedGroup)

	nodeEventSeen := false
	for eventsLeft := true; eventsLeft; {
		select {
		case event := <-fakeRecorder.Events:
			if strings.Contains(event, "TriggeredScaleUp") && strings.Contains(event, config.expectedFinalScaleUp.groupName) {
				nodeEventSeen = true
			}
			assert.NotRegexp(t, regexp.MustCompile("NotTriggerScaleUp"), event)
		default:
			eventsLeft = false
		}
	}
	assert.True(t, nodeEventSeen)
}

func getGroupSizeChangeFromChan(c chan groupSizeChange) *groupSizeChange {
	select {
	case val := <-c:
		return &val
	case <-time.After(10 * time.Second):
		return nil
	}
}

func buildTestPod(p podConfig) *apiv1.Pod {
	pod := BuildTestPod(p.name, p.cpu, p.memory)
	if p.gpu > 0 {
		RequestGpuForPod(pod, p.gpu)
	}
	if p.node != "" {
		pod.Spec.NodeName = p.node
	}
	return pod
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

	fakeRecorder := kube_util.CreateEventRecorder(fakeClient)
	fakeLogRecorder, _ := utils.NewStatusMapRecorder(fakeClient, "kube-system", fakeRecorder, false)
	clusterState := clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{}, fakeLogRecorder)
	clusterState.RegisterScaleUp(&clusterstate.ScaleUpRequest{
		NodeGroupName:   "ng2",
		Increase:        1,
		Time:            time.Now(),
		ExpectedAddTime: time.Now().Add(5 * time.Minute),
	})
	clusterState.UpdateNodes([]*apiv1.Node{n1, n2}, time.Now())

	context := &context.AutoscalingContext{
		AutoscalingOptions: context.AutoscalingOptions{
			EstimatorName:  estimator.BinpackingEstimatorName,
			MaxCoresTotal:  config.DefaultMaxClusterCores,
			MaxMemoryTotal: config.DefaultMaxClusterMemory,
		},
		PredicateChecker: simulator.NewTestPredicateChecker(),
		CloudProvider:    provider,
		ClientSet:        fakeClient,
		Recorder:         fakeRecorder,
		ExpanderStrategy: random.NewStrategy(),
		LogRecorder:      fakeLogRecorder,
	}
	p3 := BuildTestPod("p-new", 550, 0)

	processors := ca_processors.TestProcessors()

	status, err := ScaleUp(context, processors, clusterState, []*apiv1.Pod{p3}, []*apiv1.Node{n1, n2}, []*extensionsv1.DaemonSet{})
	assert.NoError(t, err)
	// A node is already coming - no need for scale up.
	assert.False(t, status.ScaledUp)
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

	fakeRecorder := kube_util.CreateEventRecorder(fakeClient)
	fakeLogRecorder, _ := utils.NewStatusMapRecorder(fakeClient, "kube-system", fakeRecorder, false)
	clusterState := clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{}, fakeLogRecorder)
	clusterState.RegisterScaleUp(&clusterstate.ScaleUpRequest{
		NodeGroupName:   "ng2",
		Increase:        1,
		Time:            time.Now(),
		ExpectedAddTime: time.Now().Add(5 * time.Minute),
	})
	clusterState.UpdateNodes([]*apiv1.Node{n1, n2}, time.Now())

	context := &context.AutoscalingContext{
		AutoscalingOptions: defaultOptions,
		PredicateChecker:   simulator.NewTestPredicateChecker(),
		CloudProvider:      provider,
		ClientSet:          fakeClient,
		Recorder:           fakeRecorder,
		ExpanderStrategy:   random.NewStrategy(),
		LogRecorder:        fakeLogRecorder,
	}
	p3 := BuildTestPod("p-new", 550, 0)

	processors := ca_processors.TestProcessors()
	status, err := ScaleUp(context, processors, clusterState, []*apiv1.Pod{p3, p3}, []*apiv1.Node{n1, n2}, []*extensionsv1.DaemonSet{})

	assert.NoError(t, err)
	// Two nodes needed but one node is already coming, so it should increase by one.
	assert.True(t, status.ScaledUp)
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

	fakeRecorder := kube_util.CreateEventRecorder(fakeClient)
	fakeLogRecorder, _ := utils.NewStatusMapRecorder(fakeClient, "kube-system", fakeRecorder, false)
	clusterState := clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{}, fakeLogRecorder)
	clusterState.UpdateNodes([]*apiv1.Node{n1, n2}, time.Now())
	context := &context.AutoscalingContext{
		AutoscalingOptions: context.AutoscalingOptions{
			EstimatorName:  estimator.BinpackingEstimatorName,
			MaxCoresTotal:  config.DefaultMaxClusterCores,
			MaxMemoryTotal: config.DefaultMaxClusterMemory,
		},
		PredicateChecker: simulator.NewTestPredicateChecker(),
		CloudProvider:    provider,
		ClientSet:        fakeClient,
		Recorder:         fakeRecorder,
		ExpanderStrategy: random.NewStrategy(),
		LogRecorder:      fakeLogRecorder,
	}
	p3 := BuildTestPod("p-new", 550, 0)

	processors := ca_processors.TestProcessors()
	status, err := ScaleUp(context, processors, clusterState, []*apiv1.Pod{p3}, []*apiv1.Node{n1, n2}, []*extensionsv1.DaemonSet{})

	assert.NoError(t, err)
	// Node group is unhealthy.
	assert.False(t, status.ScaledUp)
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

	fakeRecorder := kube_record.NewFakeRecorder(5)
	fakeLogRecorder, _ := utils.NewStatusMapRecorder(fakeClient, "kube-system", kube_record.NewFakeRecorder(5), false)
	clusterState := clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{}, fakeLogRecorder)
	clusterState.UpdateNodes([]*apiv1.Node{n1}, time.Now())
	context := &context.AutoscalingContext{
		AutoscalingOptions: context.AutoscalingOptions{
			EstimatorName:  estimator.BinpackingEstimatorName,
			MaxCoresTotal:  config.DefaultMaxClusterCores,
			MaxMemoryTotal: config.DefaultMaxClusterMemory,
		},
		PredicateChecker: simulator.NewTestPredicateChecker(),
		CloudProvider:    provider,
		ClientSet:        fakeClient,
		Recorder:         fakeRecorder,
		ExpanderStrategy: random.NewStrategy(),
		LogRecorder:      fakeLogRecorder,
	}
	p3 := BuildTestPod("p-new", 500, 0)

	processors := ca_processors.TestProcessors()
	status, err := ScaleUp(context, processors, clusterState, []*apiv1.Pod{p3}, []*apiv1.Node{n1}, []*extensionsv1.DaemonSet{})
	processors.ScaleUpStatusProcessor.Process(context, status)

	assert.NoError(t, err)
	assert.False(t, status.ScaledUp)
	var event string
	select {
	case event = <-fakeRecorder.Events:
	default:
		t.Fatal("No Event recorded, expected NotTriggerScaleUp event")
	}
	assert.Regexp(t, regexp.MustCompile("NotTriggerScaleUp"), event)
}

func TestScaleUpBalanceGroups(t *testing.T) {
	fakeClient := &fake.Clientset{}
	provider := testprovider.NewTestCloudProvider(func(string, int) error {
		return nil
	}, nil)

	type ngInfo struct {
		min, max, size int
	}
	testCfg := map[string]ngInfo{
		"ng1": {min: 1, max: 1, size: 1},
		"ng2": {min: 1, max: 2, size: 1},
		"ng3": {min: 1, max: 5, size: 1},
		"ng4": {min: 1, max: 5, size: 3},
	}
	podMap := make(map[string]*apiv1.Pod, len(testCfg))
	nodes := make([]*apiv1.Node, 0)

	for gid, gconf := range testCfg {
		provider.AddNodeGroup(gid, gconf.min, gconf.max, gconf.size)
		for i := 0; i < gconf.size; i++ {
			nodeName := fmt.Sprintf("%v-node-%v", gid, i)
			node := BuildTestNode(nodeName, 100, 1000)
			SetNodeReadyState(node, true, time.Now())
			nodes = append(nodes, node)

			pod := BuildTestPod(fmt.Sprintf("%v-pod-%v", gid, i), 80, 0)
			pod.Spec.NodeName = nodeName
			podMap[gid] = pod

			provider.AddNode(gid, node)
		}
	}

	fakeClient.Fake.AddReactor("list", "pods", func(action core.Action) (bool, runtime.Object, error) {
		list := action.(core.ListAction)
		fieldstring := list.GetListRestrictions().Fields.String()
		matcher, err := regexp.Compile("ng[0-9]")
		if err != nil {
			return false, &apiv1.PodList{Items: []apiv1.Pod{}}, err
		}
		matches := matcher.FindStringSubmatch(fieldstring)
		if len(matches) != 1 {
			return false, &apiv1.PodList{Items: []apiv1.Pod{}}, fmt.Errorf("parse error")
		}
		return true, &apiv1.PodList{Items: []apiv1.Pod{*(podMap[matches[0]])}}, nil
	})

	fakeRecorder := kube_record.NewFakeRecorder(5)
	fakeLogRecorder, _ := utils.NewStatusMapRecorder(fakeClient, "kube-system", kube_record.NewFakeRecorder(5), false)
	clusterState := clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{}, fakeLogRecorder)
	clusterState.UpdateNodes(nodes, time.Now())
	context := &context.AutoscalingContext{
		AutoscalingOptions: context.AutoscalingOptions{
			EstimatorName:            estimator.BinpackingEstimatorName,
			BalanceSimilarNodeGroups: true,
			MaxCoresTotal:            config.DefaultMaxClusterCores,
			MaxMemoryTotal:           config.DefaultMaxClusterMemory,
		},
		PredicateChecker: simulator.NewTestPredicateChecker(),
		CloudProvider:    provider,
		ClientSet:        fakeClient,
		Recorder:         fakeRecorder,
		ExpanderStrategy: random.NewStrategy(),
		LogRecorder:      fakeLogRecorder,
	}

	pods := make([]*apiv1.Pod, 0)
	for i := 0; i < 2; i++ {
		pods = append(pods, BuildTestPod(fmt.Sprintf("test-pod-%v", i), 80, 0))
	}

	processors := ca_processors.TestProcessors()
	status, typedErr := ScaleUp(context, processors, clusterState, pods, nodes, []*extensionsv1.DaemonSet{})

	assert.NoError(t, typedErr)
	assert.True(t, status.ScaledUp)
	groupMap := make(map[string]cloudprovider.NodeGroup, 3)
	for _, group := range provider.NodeGroups() {
		groupMap[group.Id()] = group
	}

	ng2size, err := groupMap["ng2"].TargetSize()
	assert.NoError(t, err)
	ng3size, err := groupMap["ng3"].TargetSize()
	assert.NoError(t, err)
	assert.Equal(t, 2, ng2size)
	assert.Equal(t, 2, ng3size)
}

func TestScaleUpAutoprovisionedNodeGroup(t *testing.T) {
	createdGroups := make(chan string, 10)
	expandedGroups := make(chan string, 10)

	p1 := BuildTestPod("p1", 80, 0)

	fakeClient := &fake.Clientset{}

	t1 := BuildTestNode("t1", 4000, 1000000)
	SetNodeReadyState(t1, true, time.Time{})
	ti1 := schedulercache.NewNodeInfo()
	ti1.SetNode(t1)

	provider := testprovider.NewTestAutoprovisioningCloudProvider(
		func(nodeGroup string, increase int) error {
			expandedGroups <- fmt.Sprintf("%s-%d", nodeGroup, increase)
			return nil
		}, nil, func(nodeGroup string) error {
			createdGroups <- nodeGroup
			return nil
		}, nil, []string{"T1"}, map[string]*schedulercache.NodeInfo{"T1": ti1})

	fakeRecorder := kube_util.CreateEventRecorder(fakeClient)
	fakeLogRecorder, _ := utils.NewStatusMapRecorder(fakeClient, "kube-system", fakeRecorder, false)
	clusterState := clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{}, fakeLogRecorder)

	context := &context.AutoscalingContext{
		AutoscalingOptions: context.AutoscalingOptions{
			EstimatorName:                    estimator.BinpackingEstimatorName,
			MaxCoresTotal:                    5000 * 64,
			MaxMemoryTotal:                   5000 * 64 * 20,
			NodeAutoprovisioningEnabled:      true,
			MaxAutoprovisionedNodeGroupCount: 10,
		},
		PredicateChecker: simulator.NewTestPredicateChecker(),
		CloudProvider:    provider,
		ClientSet:        fakeClient,
		Recorder:         fakeRecorder,
		ExpanderStrategy: random.NewStrategy(),
		LogRecorder:      fakeLogRecorder,
	}

	processors := ca_processors.TestProcessors()
	processors.NodeGroupListProcessor = nodegroups.NewAutoprovisioningNodeGroupListProcessor()

	status, err := ScaleUp(context, processors, clusterState, []*apiv1.Pod{p1}, []*apiv1.Node{}, []*extensionsv1.DaemonSet{})
	assert.NoError(t, err)
	assert.True(t, status.ScaledUp)
	assert.Equal(t, "autoprovisioned-T1", getStringFromChan(createdGroups))
	assert.Equal(t, "autoprovisioned-T1-1", getStringFromChan(expandedGroups))
}
