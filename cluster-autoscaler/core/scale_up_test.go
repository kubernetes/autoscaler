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

	"k8s.io/autoscaler/cluster-autoscaler/core/utils"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	testprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/estimator"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
	"k8s.io/autoscaler/cluster-autoscaler/utils/units"
	kube_record "k8s.io/client-go/tools/record"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/fake"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework/v1alpha1"

	"github.com/stretchr/testify/assert"
	"k8s.io/autoscaler/cluster-autoscaler/expander"
)

var defaultOptions = config.AutoscalingOptions{
	EstimatorName:  estimator.BinpackingEstimatorName,
	MaxCoresTotal:  config.DefaultMaxClusterCores,
	MaxMemoryTotal: config.DefaultMaxClusterMemory * units.GiB,
	MinCoresTotal:  0,
	MinMemoryTotal: 0,
}

// Scale up scenarios.
func TestScaleUpOK(t *testing.T) {
	config := &scaleTestConfig{
		nodes: []nodeConfig{
			{"n1", 100, 100, 0, true, "ng1"},
			{"n2", 1000, 1000, 0, true, "ng2"},
		},
		pods: []podConfig{
			{"p1", 80, 0, 0, "n1", false},
			{"p2", 800, 0, 0, "n2", false},
		},
		extraPods: []podConfig{
			{"p-new", 500, 0, 0, "", false},
		},
		options:                 defaultOptions,
		expansionOptionToChoose: groupSizeChange{groupName: "ng2", sizeChange: 1},
	}
	expectedResults := &scaleTestResults{
		finalOption: groupSizeChange{groupName: "ng2", sizeChange: 1},
		scaleUpStatus: scaleUpStatusInfo{
			podsTriggeredScaleUp: []string{"p-new"},
		},
	}

	simpleScaleUpTest(t, config, expectedResults)
}

// There are triggering, remaining & awaiting pods.
func TestMixedScaleUp(t *testing.T) {
	config := &scaleTestConfig{
		nodes: []nodeConfig{
			{"n1", 100, 1000, 0, true, "ng1"},
			{"n2", 1000, 100, 0, true, "ng2"},
		},
		pods: []podConfig{
			{"p1", 80, 0, 0, "n1", false},
			{"p2", 800, 0, 0, "n2", false},
		},
		extraPods: []podConfig{
			// can only be scheduled on ng2
			{"triggering", 900, 0, 0, "", false},
			// can't be scheduled
			{"remaining", 2000, 0, 0, "", false},
			// can only be scheduled on ng1
			{"awaiting", 0, 200, 0, "", false},
		},
		options:                 defaultOptions,
		expansionOptionToChoose: groupSizeChange{groupName: "ng2", sizeChange: 1},
	}
	expectedResults := &scaleTestResults{
		finalOption: groupSizeChange{groupName: "ng2", sizeChange: 1},
		scaleUpStatus: scaleUpStatusInfo{
			podsTriggeredScaleUp:    []string{"triggering"},
			podsRemainUnschedulable: []string{"remaining"},
			podsAwaitEvaluation:     []string{"awaiting"},
		},
	}

	simpleScaleUpTest(t, config, expectedResults)
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
			{"p1", 1000, 0, 0, "n1", false},
			{"p2", 3000, 0, 0, "n2", false},
		},
		extraPods: []podConfig{
			{"p-new-1", 2000, 0, 0, "", false},
			{"p-new-2", 2000, 0, 0, "", false},
		},
		expansionOptionToChoose: groupSizeChange{groupName: "ng1", sizeChange: 2},
		options:                 options,
	}
	results := &scaleTestResults{
		finalOption: groupSizeChange{groupName: "ng1", sizeChange: 1},
		scaleUpStatus: scaleUpStatusInfo{
			podsTriggeredScaleUp: []string{"p-new-1", "p-new-2"},
		},
	}

	simpleScaleUpTest(t, config, results)
}

func TestScaleUpMaxCoresLimitHitWithNotAutoscaledGroup(t *testing.T) {
	options := defaultOptions
	options.MaxCoresTotal = 9
	config := &scaleTestConfig{
		nodes: []nodeConfig{
			{"n1", 2000, 100, 0, true, "ng1"},
			{"n2", 4000, 1000, 0, true, ""},
		},
		pods: []podConfig{
			{"p1", 1000, 0, 0, "n1", false},
			{"p2", 3000, 0, 0, "n2", false},
		},
		extraPods: []podConfig{
			{"p-new-1", 2000, 0, 0, "", false},
			{"p-new-2", 2000, 0, 0, "", false},
		},
		expansionOptionToChoose: groupSizeChange{groupName: "ng1", sizeChange: 2},
		options:                 options,
	}
	results := &scaleTestResults{
		finalOption: groupSizeChange{groupName: "ng1", sizeChange: 1},
		scaleUpStatus: scaleUpStatusInfo{
			podsTriggeredScaleUp: []string{"p-new-1", "p-new-2"},
		},
	}

	simpleScaleUpTest(t, config, results)
}

func TestScaleUpMaxMemoryLimitHit(t *testing.T) {
	options := defaultOptions
	options.MaxMemoryTotal = 1300 * utils.MiB
	config := &scaleTestConfig{
		nodes: []nodeConfig{
			{"n1", 2000, 100 * utils.MiB, 0, true, "ng1"},
			{"n2", 4000, 1000 * utils.MiB, 0, true, "ng2"},
		},
		pods: []podConfig{
			{"p1", 1000, 0, 0, "n1", false},
			{"p2", 3000, 0, 0, "n2", false},
		},
		extraPods: []podConfig{
			{"p-new-1", 2000, 100 * utils.MiB, 0, "", false},
			{"p-new-2", 2000, 100 * utils.MiB, 0, "", false},
			{"p-new-3", 2000, 100 * utils.MiB, 0, "", false},
		},
		expansionOptionToChoose: groupSizeChange{groupName: "ng1", sizeChange: 3},
		options:                 options,
	}
	results := &scaleTestResults{
		finalOption: groupSizeChange{groupName: "ng1", sizeChange: 2},
		scaleUpStatus: scaleUpStatusInfo{
			podsTriggeredScaleUp: []string{"p-new-1", "p-new-2", "p-new-3"},
		},
	}

	simpleScaleUpTest(t, config, results)
}

func TestScaleUpMaxMemoryLimitHitWithNotAutoscaledGroup(t *testing.T) {
	options := defaultOptions
	options.MaxMemoryTotal = 1300 * utils.MiB
	config := &scaleTestConfig{
		nodes: []nodeConfig{
			{"n1", 2000, 100 * utils.MiB, 0, true, "ng1"},
			{"n2", 4000, 1000 * utils.MiB, 0, true, ""},
		},
		pods: []podConfig{
			{"p1", 1000, 0, 0, "n1", false},
			{"p2", 3000, 0, 0, "n2", false},
		},
		extraPods: []podConfig{
			{"p-new-1", 2000, 100 * utils.MiB, 0, "", false},
			{"p-new-2", 2000, 100 * utils.MiB, 0, "", false},
			{"p-new-3", 2000, 100 * utils.MiB, 0, "", false},
		},
		expansionOptionToChoose: groupSizeChange{groupName: "ng1", sizeChange: 3},
		options:                 options,
	}
	results := &scaleTestResults{
		finalOption: groupSizeChange{groupName: "ng1", sizeChange: 2},
		scaleUpStatus: scaleUpStatusInfo{
			podsTriggeredScaleUp: []string{"p-new-1", "p-new-2", "p-new-3"},
		},
	}

	simpleScaleUpTest(t, config, results)
}

func TestScaleUpCapToMaxTotalNodesLimit(t *testing.T) {
	options := defaultOptions
	options.MaxNodesTotal = 3
	config := &scaleTestConfig{
		nodes: []nodeConfig{
			{"n1", 2000, 100 * utils.MiB, 0, true, "ng1"},
			{"n2", 4000, 1000 * utils.MiB, 0, true, "ng2"},
		},
		pods: []podConfig{
			{"p1", 1000, 0, 0, "n1", false},
			{"p2", 3000, 0, 0, "n2", false},
		},
		extraPods: []podConfig{
			{"p-new-1", 4000, 100 * utils.MiB, 0, "", false},
			{"p-new-2", 4000, 100 * utils.MiB, 0, "", false},
			{"p-new-3", 4000, 100 * utils.MiB, 0, "", false},
		},
		expansionOptionToChoose: groupSizeChange{groupName: "ng2", sizeChange: 3},
		options:                 options,
	}
	results := &scaleTestResults{
		finalOption: groupSizeChange{groupName: "ng2", sizeChange: 1},
		scaleUpStatus: scaleUpStatusInfo{
			podsTriggeredScaleUp: []string{"p-new-1", "p-new-2", "p-new-3"},
		},
	}

	simpleScaleUpTest(t, config, results)
}

func TestScaleUpCapToMaxTotalNodesLimitWithNotAutoscaledGroup(t *testing.T) {
	options := defaultOptions
	options.MaxNodesTotal = 3
	config := &scaleTestConfig{
		nodes: []nodeConfig{
			{"n1", 2000, 100 * utils.MiB, 0, true, ""},
			{"n2", 4000, 1000 * utils.MiB, 0, true, "ng2"},
		},
		pods: []podConfig{
			{"p1", 1000, 0, 0, "n1", false},
			{"p2", 3000, 0, 0, "n2", false},
		},
		extraPods: []podConfig{
			{"p-new-1", 4000, 100 * utils.MiB, 0, "", false},
			{"p-new-2", 4000, 100 * utils.MiB, 0, "", false},
			{"p-new-3", 4000, 100 * utils.MiB, 0, "", false},
		},
		expansionOptionToChoose: groupSizeChange{groupName: "ng2", sizeChange: 3},
		options:                 options,
	}
	results := &scaleTestResults{
		finalOption: groupSizeChange{groupName: "ng2", sizeChange: 1},
		scaleUpStatus: scaleUpStatusInfo{
			podsTriggeredScaleUp: []string{"p-new-1", "p-new-2", "p-new-3"},
		},
	}

	simpleScaleUpTest(t, config, results)
}

func TestWillConsiderGpuAndStandardPoolForPodWhichDoesNotRequireGpu(t *testing.T) {
	options := defaultOptions
	options.MaxNodesTotal = 100
	config := &scaleTestConfig{
		nodes: []nodeConfig{
			{"gpu-node-1", 2000, 1000 * utils.MiB, 1, true, "gpu-pool"},
			{"std-node-1", 2000, 1000 * utils.MiB, 0, true, "std-pool"},
		},
		pods: []podConfig{
			{"gpu-pod-1", 2000, 1000 * utils.MiB, 1, "gpu-node-1", true},
			{"std-pod-1", 2000, 1000 * utils.MiB, 0, "std-node-1", false},
		},
		extraPods: []podConfig{
			{"extra-std-pod", 2000, 1000 * utils.MiB, 0, "", true},
		},
		expansionOptionToChoose: groupSizeChange{groupName: "std-pool", sizeChange: 1},
		options:                 options,
	}
	results := &scaleTestResults{
		finalOption: groupSizeChange{groupName: "std-pool", sizeChange: 1},
		expansionOptions: []groupSizeChange{
			{groupName: "std-pool", sizeChange: 1},
			{groupName: "gpu-pool", sizeChange: 1},
		},
		scaleUpStatus: scaleUpStatusInfo{
			podsTriggeredScaleUp: []string{"extra-std-pod"},
		},
	}

	simpleScaleUpTest(t, config, results)
}

func TestWillConsiderOnlyGpuPoolForPodWhichDoesRequiresGpu(t *testing.T) {
	options := defaultOptions
	options.MaxNodesTotal = 100
	config := &scaleTestConfig{
		nodes: []nodeConfig{
			{"gpu-node-1", 2000, 1000 * utils.MiB, 1, true, "gpu-pool"},
			{"std-node-1", 2000, 1000 * utils.MiB, 0, true, "std-pool"},
		},
		pods: []podConfig{
			{"gpu-pod-1", 2000, 1000 * utils.MiB, 1, "gpu-node-1", true},
			{"std-pod-1", 2000, 1000 * utils.MiB, 0, "std-node-1", false},
		},
		extraPods: []podConfig{
			{"extra-gpu-pod", 2000, 1000 * utils.MiB, 1, "", true},
		},
		expansionOptionToChoose: groupSizeChange{groupName: "gpu-pool", sizeChange: 1},
		options:                 options,
	}
	results := &scaleTestResults{
		finalOption: groupSizeChange{groupName: "gpu-pool", sizeChange: 1},
		expansionOptions: []groupSizeChange{
			{groupName: "gpu-pool", sizeChange: 1},
		},
		scaleUpStatus: scaleUpStatusInfo{
			podsTriggeredScaleUp: []string{"extra-gpu-pod"},
		},
	}

	simpleScaleUpTest(t, config, results)
}

func TestWillConsiderAllPoolsWhichFitTwoPodsRequiringGpus(t *testing.T) {
	options := defaultOptions
	options.MaxNodesTotal = 100
	config := &scaleTestConfig{
		nodes: []nodeConfig{
			{"gpu-1-node-1", 2000, 1000 * utils.MiB, 1, true, "gpu-1-pool"},
			{"gpu-2-node-1", 2000, 1000 * utils.MiB, 2, true, "gpu-2-pool"},
			{"gpu-4-node-1", 2000, 1000 * utils.MiB, 4, true, "gpu-4-pool"},
			{"std-node-1", 2000, 1000 * utils.MiB, 0, true, "std-pool"},
		},
		pods: []podConfig{
			{"gpu-pod-1", 2000, 1000 * utils.MiB, 1, "gpu-1-node-1", true},
			{"gpu-pod-2", 2000, 1000 * utils.MiB, 2, "gpu-2-node-1", true},
			{"gpu-pod-3", 2000, 1000 * utils.MiB, 4, "gpu-4-node-1", true},
			{"std-pod-1", 2000, 1000 * utils.MiB, 0, "std-node-1", false},
		},
		extraPods: []podConfig{
			{"extra-gpu-pod-1", 1, 1 * utils.MiB, 1, "", true}, // CPU and mem negligible
			{"extra-gpu-pod-2", 1, 1 * utils.MiB, 1, "", true}, // CPU and mem negligible
			{"extra-gpu-pod-3", 1, 1 * utils.MiB, 1, "", true}, // CPU and mem negligible
		},
		expansionOptionToChoose: groupSizeChange{groupName: "gpu-1-pool", sizeChange: 3},
		options:                 options,
	}
	results := &scaleTestResults{
		finalOption: groupSizeChange{groupName: "gpu-1-pool", sizeChange: 3},
		expansionOptions: []groupSizeChange{
			{groupName: "gpu-1-pool", sizeChange: 3},
			{groupName: "gpu-2-pool", sizeChange: 2},
			{groupName: "gpu-4-pool", sizeChange: 1},
		},
		scaleUpStatus: scaleUpStatusInfo{
			podsTriggeredScaleUp: []string{"extra-gpu-pod-1", "extra-gpu-pod-2", "extra-gpu-pod-3"},
		},
	}

	simpleScaleUpTest(t, config, results)
}

// No scale up scenarios.
func TestNoScaleUpMaxCoresLimitHit(t *testing.T) {
	options := defaultOptions
	options.MaxCoresTotal = 7
	options.MaxMemoryTotal = 1150
	config := &scaleTestConfig{
		nodes: []nodeConfig{
			{"n1", 2000, 100, 0, true, "ng1"},
			{"n2", 4000, 1000, 0, true, "ng2"},
		},
		pods: []podConfig{
			{"p1", 1000, 0, 0, "n1", false},
			{"p2", 3000, 0, 0, "n2", false},
		},
		extraPods: []podConfig{
			{"p-new-1", 2000, 0, 0, "", false},
			{"p-new-2", 2000, 0, 0, "", false},
		},
		options: options,
	}
	results := &scaleTestResults{
		noScaleUpReason: "max cluster cpu, memory limit reached",
		scaleUpStatus: scaleUpStatusInfo{
			podsRemainUnschedulable: []string{"p-new-1", "p-new-2"},
		},
	}

	simpleNoScaleUpTest(t, config, results)
}

// To implement expander.Strategy, BestOption method must have a struct receiver.
// This prevents it from modifying fields of reportingStrategy, so we need a thin
// pointer wrapper for mutable parts.
type expanderResults struct {
	inputOptions []groupSizeChange
}

type reportingStrategy struct {
	initialNodeConfigs []nodeConfig
	optionToChoose     groupSizeChange
	results            *expanderResults
	t                  *testing.T
}

func (r reportingStrategy) BestOption(options []expander.Option, nodeInfo map[string]*schedulerframework.NodeInfo) *expander.Option {
	r.results.inputOptions = expanderOptionsToGroupSizeChanges(options)
	for _, option := range options {
		groupSizeChange := expanderOptionToGroupSizeChange(option)
		if groupSizeChange == r.optionToChoose {
			return &option
		}
	}
	assert.Fail(r.t, "did not find expansionOptionToChoose %s", r.optionToChoose)
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

func runSimpleScaleUpTest(t *testing.T, config *scaleTestConfig) *scaleTestResults {
	expandedGroups := make(chan groupSizeChange, 10)

	groups := make(map[string][]*apiv1.Node)
	nodes := make([]*apiv1.Node, len(config.nodes))
	for i, n := range config.nodes {
		node := BuildTestNode(n.name, n.cpu, n.memory)
		if n.gpu > 0 {
			AddGpusToNode(node, n.gpu)
		}
		SetNodeReadyState(node, n.ready, time.Now())
		nodes[i] = node
		if n.group != "" {
			groups[n.group] = append(groups[n.group], node)
		}
	}

	pods := make([]*apiv1.Pod, 0)
	for _, p := range config.pods {
		pod := buildTestPod(p)
		pods = append(pods, pod)
	}

	podLister := kube_util.NewTestPodLister(pods)
	listers := kube_util.NewListerRegistry(nil, nil, podLister, nil, nil, nil, nil, nil, nil, nil)

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

	// Create context with non-random expander strategy.
	context, err := NewScaleTestAutoscalingContext(config.options, &fake.Clientset{}, listers, provider, nil)
	assert.NoError(t, err)

	expander := reportingStrategy{
		initialNodeConfigs: config.nodes,
		optionToChoose:     config.expansionOptionToChoose,
		results:            &expanderResults{},
		t:                  t,
	}
	context.ExpanderStrategy = expander

	nodeInfos, _ := utils.GetNodeInfosForGroups(nodes, nil, provider, listers, []*appsv1.DaemonSet{}, context.PredicateChecker, nil)
	clusterState := clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{}, context.LogRecorder, newBackoff())
	clusterState.UpdateNodes(nodes, nodeInfos, time.Now())

	extraPods := make([]*apiv1.Pod, len(config.extraPods))
	for i, p := range config.extraPods {
		pod := buildTestPod(p)
		extraPods[i] = pod
	}

	processors := NewTestProcessors()

	scaleUpStatus, err := ScaleUp(&context, processors, clusterState, extraPods, nodes, []*appsv1.DaemonSet{}, nodeInfos, nil)
	processors.ScaleUpStatusProcessor.Process(&context, scaleUpStatus)

	assert.NoError(t, err)

	expandedGroup := getGroupSizeChangeFromChan(expandedGroups)
	var expandedGroupStruct groupSizeChange
	if expandedGroup != nil {
		expandedGroupStruct = *expandedGroup
	}

	events := []string{}
	for eventsLeft := true; eventsLeft; {
		select {
		case event := <-context.Recorder.(*kube_record.FakeRecorder).Events:
			events = append(events, event)
		default:
			eventsLeft = false
		}
	}

	return &scaleTestResults{
		expansionOptions: expander.results.inputOptions,
		finalOption:      expandedGroupStruct,
		scaleUpStatus:    simplifyScaleUpStatus(scaleUpStatus),
		events:           events,
	}
}

func simpleNoScaleUpTest(t *testing.T, config *scaleTestConfig, expectedResults *scaleTestResults) {
	results := runSimpleScaleUpTest(t, config)

	assert.Equal(t, groupSizeChange{}, results.finalOption)
	assert.False(t, results.scaleUpStatus.WasSuccessful())
	noScaleUpEventSeen := false
	for _, event := range results.events {
		if strings.Contains(event, "NotTriggerScaleUp") {
			if strings.Contains(event, expectedResults.noScaleUpReason) {
				noScaleUpEventSeen = true
			} else {
				// Surprisingly useful for debugging.
				fmt.Println("Event:", event)
			}
		}
		assert.NotRegexp(t, regexp.MustCompile("TriggeredScaleUp"), event)
	}
	assert.True(t, noScaleUpEventSeen)
	assert.ElementsMatch(t, results.scaleUpStatus.podsTriggeredScaleUp, expectedResults.scaleUpStatus.podsTriggeredScaleUp,
		"actual and expected triggering pods should be the same")
	assert.ElementsMatch(t, results.scaleUpStatus.podsRemainUnschedulable, expectedResults.scaleUpStatus.podsRemainUnschedulable,
		"actual and expected remaining pods should be the same")
	assert.ElementsMatch(t, results.scaleUpStatus.podsAwaitEvaluation, expectedResults.scaleUpStatus.podsAwaitEvaluation,
		"actual and expected awaiting evaluation pods should be the same")
}

func simpleScaleUpTest(t *testing.T, config *scaleTestConfig, expectedResults *scaleTestResults) {
	results := runSimpleScaleUpTest(t, config)

	assert.NotNil(t, results.finalOption, "Expected scale up event")
	assert.Equal(t, expectedResults.finalOption, results.finalOption)
	assert.True(t, results.scaleUpStatus.WasSuccessful())
	nodeEventSeen := false
	for _, event := range results.events {
		if strings.Contains(event, "TriggeredScaleUp") && strings.Contains(event, expectedResults.finalOption.groupName) {
			nodeEventSeen = true
		}
		if len(expectedResults.scaleUpStatus.podsRemainUnschedulable) == 0 {
			assert.NotRegexp(t, regexp.MustCompile("NotTriggerScaleUp"), event)
		}
	}
	assert.True(t, nodeEventSeen)

	if len(expectedResults.expansionOptions) > 0 {
		// Empty expansionOptions means we do not want to do any assertions
		// on contents of actual scaleUp options

		// Check that option to choose is part of expected options.
		assert.Contains(t, expectedResults.expansionOptions, config.expansionOptionToChoose, "final expected expansion option must be in expected expansion options")
		assert.Contains(t, results.expansionOptions, config.expansionOptionToChoose, "final expected expansion option must be in expected expansion options")

		assert.ElementsMatch(t, results.expansionOptions, expectedResults.expansionOptions,
			"actual and expected expansion options should be the same")
	}

	assert.ElementsMatch(t, results.scaleUpStatus.podsTriggeredScaleUp, expectedResults.scaleUpStatus.podsTriggeredScaleUp,
		"actual and expected triggering pods should be the same")
	assert.ElementsMatch(t, results.scaleUpStatus.podsRemainUnschedulable, expectedResults.scaleUpStatus.podsRemainUnschedulable,
		"actual and expected remaining pods should be the same")
	assert.ElementsMatch(t, results.scaleUpStatus.podsAwaitEvaluation, expectedResults.scaleUpStatus.podsAwaitEvaluation,
		"actual and expected awaiting evaluation pods should be the same")
}

func getGroupSizeChangeFromChan(c chan groupSizeChange) *groupSizeChange {
	select {
	case val := <-c:
		return &val
	case <-time.After(100 * time.Millisecond):
		return nil
	}
}

func buildTestPod(p podConfig) *apiv1.Pod {
	pod := BuildTestPod(p.name, p.cpu, p.memory)
	if p.gpu > 0 {
		RequestGpuForPod(pod, p.gpu)
	}
	if p.toleratesGpu {
		TolerateGpuForPod(pod)
	}
	if p.node != "" {
		pod.Spec.NodeName = p.node
	}
	return pod
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

	podLister := kube_util.NewTestPodLister([]*apiv1.Pod{p1, p2})
	listers := kube_util.NewListerRegistry(nil, nil, podLister, nil, nil, nil, nil, nil, nil, nil)

	provider := testprovider.NewTestCloudProvider(func(nodeGroup string, increase int) error {
		t.Fatalf("No expansion is expected, but increased %s by %d", nodeGroup, increase)
		return nil
	}, nil)
	provider.AddNodeGroup("ng1", 1, 10, 1)
	provider.AddNodeGroup("ng2", 1, 10, 5)
	provider.AddNode("ng1", n1)
	provider.AddNode("ng2", n2)

	options := config.AutoscalingOptions{
		EstimatorName:  estimator.BinpackingEstimatorName,
		MaxCoresTotal:  config.DefaultMaxClusterCores,
		MaxMemoryTotal: config.DefaultMaxClusterMemory,
	}
	context, err := NewScaleTestAutoscalingContext(options, &fake.Clientset{}, listers, provider, nil)
	assert.NoError(t, err)

	nodes := []*apiv1.Node{n1, n2}
	nodeInfos, _ := utils.GetNodeInfosForGroups(nodes, nil, provider, listers, []*appsv1.DaemonSet{}, context.PredicateChecker, nil)
	clusterState := clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{}, context.LogRecorder, newBackoff())
	clusterState.UpdateNodes(nodes, nodeInfos, time.Now())
	p3 := BuildTestPod("p-new", 550, 0)

	processors := NewTestProcessors()
	scaleUpStatus, err := ScaleUp(&context, processors, clusterState, []*apiv1.Pod{p3}, nodes, []*appsv1.DaemonSet{}, nodeInfos, nil)

	assert.NoError(t, err)
	// Node group is unhealthy.
	assert.False(t, scaleUpStatus.WasSuccessful())
}

func TestScaleUpNoHelp(t *testing.T) {
	n1 := BuildTestNode("n1", 100, 1000)
	SetNodeReadyState(n1, true, time.Now())

	p1 := BuildTestPod("p1", 80, 0)
	p1.Spec.NodeName = "n1"

	podLister := kube_util.NewTestPodLister([]*apiv1.Pod{p1})
	listers := kube_util.NewListerRegistry(nil, nil, podLister, nil, nil, nil, nil, nil, nil, nil)

	provider := testprovider.NewTestCloudProvider(func(nodeGroup string, increase int) error {
		t.Fatalf("No expansion is expected")
		return nil
	}, nil)
	provider.AddNodeGroup("ng1", 1, 10, 1)
	provider.AddNode("ng1", n1)
	assert.NotNil(t, provider)

	options := config.AutoscalingOptions{
		EstimatorName:  estimator.BinpackingEstimatorName,
		MaxCoresTotal:  config.DefaultMaxClusterCores,
		MaxMemoryTotal: config.DefaultMaxClusterMemory,
	}
	context, err := NewScaleTestAutoscalingContext(options, &fake.Clientset{}, listers, provider, nil)
	assert.NoError(t, err)

	nodes := []*apiv1.Node{n1}
	nodeInfos, _ := utils.GetNodeInfosForGroups(nodes, nil, provider, listers, []*appsv1.DaemonSet{}, context.PredicateChecker, nil)
	clusterState := clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{}, context.LogRecorder, newBackoff())
	clusterState.UpdateNodes(nodes, nodeInfos, time.Now())
	p3 := BuildTestPod("p-new", 500, 0)

	processors := NewTestProcessors()
	scaleUpStatus, err := ScaleUp(&context, processors, clusterState, []*apiv1.Pod{p3}, nodes, []*appsv1.DaemonSet{}, nodeInfos, nil)
	processors.ScaleUpStatusProcessor.Process(&context, scaleUpStatus)

	assert.NoError(t, err)
	assert.False(t, scaleUpStatus.WasSuccessful())
	var event string
	select {
	case event = <-context.Recorder.(*kube_record.FakeRecorder).Events:
	default:
		t.Fatal("No Event recorded, expected NotTriggerScaleUp event")
	}
	assert.Regexp(t, regexp.MustCompile("NotTriggerScaleUp"), event)
}

func TestScaleUpBalanceGroups(t *testing.T) {
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
	podList := make([]*apiv1.Pod, 0, len(testCfg))
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
			podList = append(podList, pod)

			provider.AddNode(gid, node)
		}
	}

	podLister := kube_util.NewTestPodLister(podList)
	listers := kube_util.NewListerRegistry(nil, nil, podLister, nil, nil, nil, nil, nil, nil, nil)

	options := config.AutoscalingOptions{
		EstimatorName:            estimator.BinpackingEstimatorName,
		BalanceSimilarNodeGroups: true,
		MaxCoresTotal:            config.DefaultMaxClusterCores,
		MaxMemoryTotal:           config.DefaultMaxClusterMemory,
	}
	context, err := NewScaleTestAutoscalingContext(options, &fake.Clientset{}, listers, provider, nil)
	assert.NoError(t, err)

	nodeInfos, _ := utils.GetNodeInfosForGroups(nodes, nil, provider, listers, []*appsv1.DaemonSet{}, context.PredicateChecker, nil)
	clusterState := clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{}, context.LogRecorder, newBackoff())
	clusterState.UpdateNodes(nodes, nodeInfos, time.Now())

	pods := make([]*apiv1.Pod, 0)
	for i := 0; i < 2; i++ {
		pods = append(pods, BuildTestPod(fmt.Sprintf("test-pod-%v", i), 80, 0))
	}

	processors := NewTestProcessors()
	scaleUpStatus, typedErr := ScaleUp(&context, processors, clusterState, pods, nodes, []*appsv1.DaemonSet{}, nodeInfos, nil)

	assert.NoError(t, typedErr)
	assert.True(t, scaleUpStatus.WasSuccessful())
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
	ti1 := schedulerframework.NewNodeInfo()
	ti1.SetNode(t1)

	provider := testprovider.NewTestAutoprovisioningCloudProvider(
		func(nodeGroup string, increase int) error {
			expandedGroups <- fmt.Sprintf("%s-%d", nodeGroup, increase)
			return nil
		}, nil, func(nodeGroup string) error {
			createdGroups <- nodeGroup
			return nil
		}, nil, []string{"T1"}, map[string]*schedulerframework.NodeInfo{"T1": ti1})

	options := config.AutoscalingOptions{
		EstimatorName:                    estimator.BinpackingEstimatorName,
		MaxCoresTotal:                    5000 * 64,
		MaxMemoryTotal:                   5000 * 64 * 20,
		NodeAutoprovisioningEnabled:      true,
		MaxAutoprovisionedNodeGroupCount: 10,
	}
	podLister := kube_util.NewTestPodLister([]*apiv1.Pod{})
	listers := kube_util.NewListerRegistry(nil, nil, podLister, nil, nil, nil, nil, nil, nil, nil)
	context, err := NewScaleTestAutoscalingContext(options, fakeClient, listers, provider, nil)
	assert.NoError(t, err)

	clusterState := clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{}, context.LogRecorder, newBackoff())

	processors := NewTestProcessors()
	processors.NodeGroupListProcessor = &mockAutoprovisioningNodeGroupListProcessor{t}
	processors.NodeGroupManager = &mockAutoprovisioningNodeGroupManager{t, 0}

	nodes := []*apiv1.Node{}
	nodeInfos, _ := utils.GetNodeInfosForGroups(nodes, nil, provider, context.ListerRegistry, []*appsv1.DaemonSet{}, context.PredicateChecker, nil)

	scaleUpStatus, err := ScaleUp(&context, processors, clusterState, []*apiv1.Pod{p1}, nodes, []*appsv1.DaemonSet{}, nodeInfos, nil)
	assert.NoError(t, err)
	assert.True(t, scaleUpStatus.WasSuccessful())
	assert.Equal(t, "autoprovisioned-T1", utils.GetStringFromChan(createdGroups))
	assert.Equal(t, "autoprovisioned-T1-1", utils.GetStringFromChan(expandedGroups))
}

func TestScaleUpBalanceAutoprovisionedNodeGroups(t *testing.T) {
	createdGroups := make(chan string, 10)
	expandedGroups := make(chan string, 10)

	p1 := BuildTestPod("p1", 80, 0)
	p2 := BuildTestPod("p2", 80, 0)
	p3 := BuildTestPod("p3", 80, 0)

	fakeClient := &fake.Clientset{}

	t1 := BuildTestNode("t1", 100, 1000000)
	SetNodeReadyState(t1, true, time.Time{})
	ti1 := schedulerframework.NewNodeInfo()
	ti1.SetNode(t1)

	provider := testprovider.NewTestAutoprovisioningCloudProvider(
		func(nodeGroup string, increase int) error {
			expandedGroups <- fmt.Sprintf("%s-%d", nodeGroup, increase)
			return nil
		}, nil, func(nodeGroup string) error {
			createdGroups <- nodeGroup
			return nil
		}, nil, []string{"T1"}, map[string]*schedulerframework.NodeInfo{"T1": ti1})

	options := config.AutoscalingOptions{
		BalanceSimilarNodeGroups:         true,
		EstimatorName:                    estimator.BinpackingEstimatorName,
		MaxCoresTotal:                    5000 * 64,
		MaxMemoryTotal:                   5000 * 64 * 20,
		NodeAutoprovisioningEnabled:      true,
		MaxAutoprovisionedNodeGroupCount: 10,
	}
	podLister := kube_util.NewTestPodLister([]*apiv1.Pod{})
	listers := kube_util.NewListerRegistry(nil, nil, podLister, nil, nil, nil, nil, nil, nil, nil)
	context, err := NewScaleTestAutoscalingContext(options, fakeClient, listers, provider, nil)
	assert.NoError(t, err)

	clusterState := clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{}, context.LogRecorder, newBackoff())

	processors := NewTestProcessors()
	processors.NodeGroupListProcessor = &mockAutoprovisioningNodeGroupListProcessor{t}
	processors.NodeGroupManager = &mockAutoprovisioningNodeGroupManager{t, 2}

	nodes := []*apiv1.Node{}
	nodeInfos, _ := utils.GetNodeInfosForGroups(nodes, nil, provider, context.ListerRegistry, []*appsv1.DaemonSet{}, context.PredicateChecker, nil)

	scaleUpStatus, err := ScaleUp(&context, processors, clusterState, []*apiv1.Pod{p1, p2, p3}, nodes, []*appsv1.DaemonSet{}, nodeInfos, nil)
	assert.NoError(t, err)
	assert.True(t, scaleUpStatus.WasSuccessful())
	assert.Equal(t, "autoprovisioned-T1", utils.GetStringFromChan(createdGroups))
	expandedGroupMap := map[string]bool{}
	for i := 0; i < 3; i++ {
		expandedGroupMap[utils.GetStringFromChan(expandedGroups)] = true
	}
	assert.True(t, expandedGroupMap["autoprovisioned-T1-1"])
	assert.True(t, expandedGroupMap["autoprovisioned-T1-1-1"])
	assert.True(t, expandedGroupMap["autoprovisioned-T1-2-1"])
}

func TestCheckScaleUpDeltaWithinLimits(t *testing.T) {
	type testcase struct {
		limits            scaleUpResourcesLimits
		delta             scaleUpResourcesDelta
		exceededResources []string
	}
	tests := []testcase{
		{
			limits:            scaleUpResourcesLimits{"a": 10},
			delta:             scaleUpResourcesDelta{"a": 10},
			exceededResources: []string{},
		},
		{
			limits:            scaleUpResourcesLimits{"a": 10},
			delta:             scaleUpResourcesDelta{"a": 11},
			exceededResources: []string{"a"},
		},
		{
			limits:            scaleUpResourcesLimits{"a": 10},
			delta:             scaleUpResourcesDelta{"b": 10},
			exceededResources: []string{},
		},
		{
			limits:            scaleUpResourcesLimits{"a": scaleUpLimitUnknown},
			delta:             scaleUpResourcesDelta{"a": 0},
			exceededResources: []string{},
		},
		{
			limits:            scaleUpResourcesLimits{"a": scaleUpLimitUnknown},
			delta:             scaleUpResourcesDelta{"a": 1},
			exceededResources: []string{"a"},
		},
		{
			limits:            scaleUpResourcesLimits{"a": 10, "b": 20, "c": 30},
			delta:             scaleUpResourcesDelta{"a": 11, "b": 20, "c": 31},
			exceededResources: []string{"a", "c"},
		},
	}

	for _, test := range tests {
		checkResult := test.limits.checkScaleUpDeltaWithinLimits(test.delta)
		if len(test.exceededResources) == 0 {
			assert.Equal(t, scaleUpLimitsNotExceeded(), checkResult)
		} else {
			assert.Equal(t, scaleUpLimitsCheckResult{true, test.exceededResources}, checkResult)
		}
	}
}
