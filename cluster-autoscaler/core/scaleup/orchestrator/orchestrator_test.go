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

package orchestrator

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/processors/nodegroupconfig"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodegroups/asyncnodegroups"
	kube_record "k8s.io/client-go/tools/record"
	"k8s.io/component-base/metrics/legacyregistry"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	testprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaleup/resource"
	. "k8s.io/autoscaler/cluster-autoscaler/core/test"
	"k8s.io/autoscaler/cluster-autoscaler/core/utils"
	"k8s.io/autoscaler/cluster-autoscaler/estimator"
	"k8s.io/autoscaler/cluster-autoscaler/metrics"
	"k8s.io/autoscaler/cluster-autoscaler/processors"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodegroupset"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodeinfosprovider"
	"k8s.io/autoscaler/cluster-autoscaler/processors/status"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
	"k8s.io/autoscaler/cluster-autoscaler/utils/units"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/fake"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"

	"github.com/stretchr/testify/assert"
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
	config := &ScaleUpTestConfig{
		Nodes: []NodeConfig{
			{Name: "n1", Cpu: 100, Memory: 100, Gpu: 0, Ready: true, Group: "ng1"},
			{Name: "n2", Cpu: 1000, Memory: 1000, Gpu: 0, Ready: true, Group: "ng2"},
		},
		Pods: []PodConfig{
			{Name: "p1", Cpu: 80, Memory: 0, Gpu: 0, Node: "n1", ToleratesGpu: false},
			{Name: "p2", Cpu: 800, Memory: 0, Gpu: 0, Node: "n2", ToleratesGpu: false},
		},
		ExtraPods: []PodConfig{
			{Name: "p-new", Cpu: 500, Memory: 0, Gpu: 0, Node: "", ToleratesGpu: false},
		},
		ExpansionOptionToChoose: &GroupSizeChange{GroupName: "ng2", SizeChange: 1},
	}
	expectedResults := &ScaleTestResults{
		FinalOption: GroupSizeChange{GroupName: "ng2", SizeChange: 1},
		ScaleUpStatus: ScaleUpStatusInfo{
			PodsTriggeredScaleUp: []string{"p-new"},
		},
	}

	simpleScaleUpTest(t, config, expectedResults)
}

// There are triggering, remaining & awaiting pods.
func TestMixedScaleUp(t *testing.T) {
	config := &ScaleUpTestConfig{
		Nodes: []NodeConfig{
			{Name: "n1", Cpu: 100, Memory: 1000, Gpu: 0, Ready: true, Group: "ng1"},
			{Name: "n2", Cpu: 1000, Memory: 100, Gpu: 0, Ready: true, Group: "ng2"},
		},
		Pods: []PodConfig{
			{Name: "p1", Cpu: 80, Memory: 0, Gpu: 0, Node: "n1", ToleratesGpu: false},
			{Name: "p2", Cpu: 800, Memory: 0, Gpu: 0, Node: "n2", ToleratesGpu: false},
		},
		ExtraPods: []PodConfig{
			// can only be scheduled on ng2
			{Name: "triggering", Cpu: 900, Memory: 0, Gpu: 0, Node: "", ToleratesGpu: false},
			// can't be scheduled
			{Name: "remaining", Cpu: 2000, Memory: 0, Gpu: 0, Node: "", ToleratesGpu: false},
			// can only be scheduled on ng1
			{Name: "awaiting", Cpu: 0, Memory: 200, Gpu: 0, Node: "", ToleratesGpu: false},
		},
		ExpansionOptionToChoose: &GroupSizeChange{GroupName: "ng2", SizeChange: 1},
	}
	expectedResults := &ScaleTestResults{
		FinalOption: GroupSizeChange{GroupName: "ng2", SizeChange: 1},
		ScaleUpStatus: ScaleUpStatusInfo{
			PodsTriggeredScaleUp:    []string{"triggering"},
			PodsRemainUnschedulable: []string{"remaining"},
			PodsAwaitEvaluation:     []string{"awaiting"},
		},
	}

	simpleScaleUpTest(t, config, expectedResults)
}

func TestZeroOrMaxNodeScaling(t *testing.T) {
	options := defaultOptions
	options.NodeGroupDefaults.ZeroOrMaxNodeScaling = true

	optionsWithLimitedMaxCores := options
	optionsWithLimitedMaxCores.MaxCoresTotal = 3

	optionsWithLimitedMaxMemory := options
	optionsWithLimitedMaxMemory.MaxMemoryTotal = 3000

	optionsWithLimitedMaxNodes := options
	optionsWithLimitedMaxNodes.MaxNodesTotal = 5

	n := BuildTestNode("n", 1000, 1000)
	SetNodeReadyState(n, true, time.Time{})
	nodeInfo := schedulerframework.NewNodeInfo()
	nodeInfo.SetNode(n)

	cases := map[string]struct {
		testConfig      *ScaleUpTestConfig
		expectedResults *ScaleTestResults
		isScaleUpOk     bool
	}{
		"Atomic ScaleUp OK": {
			testConfig: &ScaleUpTestConfig{
				Groups: []NodeGroupConfig{
					{Name: "ng1", MaxSize: 5},
					{Name: "ng2", MaxSize: 5},
				},
				Nodes: []NodeConfig{
					{Name: "n1", Cpu: 900, Memory: 900, Gpu: 0, Ready: true, Group: "ng1"},
				},
				Pods: []PodConfig{
					{Name: "p1", Cpu: 900, Memory: 900, Gpu: 0, Node: "n1", ToleratesGpu: false},
				},
				ExtraPods: []PodConfig{
					{Name: "p-new", Cpu: 1000, Memory: 1000, Gpu: 0, Node: "", ToleratesGpu: false},
				},
				// ExpansionOptionToChoose: &GroupSizeChange{GroupName: "ng2", SizeChange: 2},
				Options: &options,
				NodeTemplateConfigs: map[string]*NodeTemplateConfig{
					"ng2": {
						NodeGroupName: "ng2",
						MachineType:   "ct4p",
						NodeInfo:      nodeInfo,
					},
				},
			},
			expectedResults: &ScaleTestResults{
				FinalOption: GroupSizeChange{GroupName: "ng2", SizeChange: 5},
				ScaleUpStatus: ScaleUpStatusInfo{
					PodsTriggeredScaleUp: []string{"p-new"},
				},
			},
			isScaleUpOk: true,
		},
		"Atomic ScaleUp with similar node groups": {
			testConfig: &ScaleUpTestConfig{
				Groups: []NodeGroupConfig{
					{Name: "ng1", MaxSize: 5},
					{Name: "ng2", MaxSize: 5},
					{Name: "ng3", MaxSize: 4},
				},
				Nodes: []NodeConfig{
					{Name: "n1", Cpu: 900, Memory: 900, Gpu: 0, Ready: true, Group: "ng1"},
				},
				Pods: []PodConfig{
					{Name: "p1", Cpu: 900, Memory: 900, Gpu: 0, Node: "n1", ToleratesGpu: false},
				},
				ExtraPods: []PodConfig{
					{Name: "p-new-1", Cpu: 1000, Memory: 1000, Gpu: 0, Node: "", ToleratesGpu: false},
					{Name: "p-new-2", Cpu: 1000, Memory: 1000, Gpu: 0, Node: "", ToleratesGpu: false},
					{Name: "p-new-3", Cpu: 1000, Memory: 1000, Gpu: 0, Node: "", ToleratesGpu: false},
					{Name: "p-new-4", Cpu: 1000, Memory: 1000, Gpu: 0, Node: "", ToleratesGpu: false},
				},
				// ExpansionOptionToChoose: &GroupSizeChange{GroupName: "ng2", SizeChange: 2},
				Options: &options,
				NodeTemplateConfigs: map[string]*NodeTemplateConfig{
					"ng2": {
						NodeGroupName: "ng2",
						MachineType:   "ct4p",
						NodeInfo:      nodeInfo,
					},
					"ng3": {
						NodeGroupName: "ng3",
						MachineType:   "ct4p",
						NodeInfo:      nodeInfo,
					},
				},
				ExpansionOptionToChoose: &GroupSizeChange{GroupName: "ng3", SizeChange: 4},
			},
			expectedResults: &ScaleTestResults{
				FinalOption: GroupSizeChange{GroupName: "ng3", SizeChange: 4},
				ScaleUpStatus: ScaleUpStatusInfo{
					PodsTriggeredScaleUp: []string{"p-new-1", "p-new-2", "p-new-3", "p-new-4"},
				},
			},

			isScaleUpOk: true,
		},
		"Atomic ScaleUp Mixed": {
			testConfig: &ScaleUpTestConfig{
				Groups: []NodeGroupConfig{
					{Name: "ng1", MaxSize: 5},
					{Name: "ng2", MaxSize: 5},
					{Name: "ng3", MaxSize: 5},
				},
				Nodes: []NodeConfig{
					{Name: "n1", Cpu: 500, Memory: 2000, Gpu: 0, Ready: true, Group: "ng1"},
					{Name: "n2", Cpu: 2000, Memory: 500, Gpu: 0, Ready: true, Group: "ng2"},
				},
				Pods: []PodConfig{
					{Name: "p1", Cpu: 400, Memory: 1900, Gpu: 0, Node: "n1", ToleratesGpu: false},
					{Name: "p2", Cpu: 1900, Memory: 400, Gpu: 0, Node: "n2", ToleratesGpu: false},
				},
				ExtraPods: []PodConfig{
					{Name: "p-triggering", Cpu: 1000, Memory: 1000, Gpu: 0, Node: "", ToleratesGpu: false},
					{Name: "p-remaining", Cpu: 2000, Memory: 2000, Gpu: 0, Node: "", ToleratesGpu: false},
					{Name: "p-awaiting", Cpu: 100, Memory: 1800, Gpu: 0, Node: "", ToleratesGpu: false},
				},
				ExpansionOptionToChoose: &GroupSizeChange{GroupName: "ng3", SizeChange: 5},
				Options:                 &options,
				NodeTemplateConfigs: map[string]*NodeTemplateConfig{
					"ng3": {
						NodeGroupName: "ng3",
						MachineType:   "ct4p",
						NodeInfo:      nodeInfo,
					},
				},
			},
			expectedResults: &ScaleTestResults{
				FinalOption: GroupSizeChange{GroupName: "ng3", SizeChange: 5},
				ScaleUpStatus: ScaleUpStatusInfo{
					PodsTriggeredScaleUp:    []string{"p-triggering"},
					PodsRemainUnschedulable: []string{"p-remaining"},
					PodsAwaitEvaluation:     []string{"p-awaiting"},
				},
			},
			isScaleUpOk: true,
		},
		"Atomic ScaleUp max cores limit hit": {
			testConfig: &ScaleUpTestConfig{
				Groups: []NodeGroupConfig{
					{Name: "ng1", MaxSize: 5},
					{Name: "ng2", MaxSize: 3},
				},
				Nodes: []NodeConfig{
					{Name: "n1", Cpu: 900, Memory: 900, Gpu: 0, Ready: true, Group: "ng1"},
				},
				Pods: []PodConfig{
					{Name: "p1", Cpu: 900, Memory: 900, Gpu: 0, Node: "n1", ToleratesGpu: false},
				},
				ExtraPods: []PodConfig{
					{Name: "p-new-1", Cpu: 1000, Memory: 1000, Gpu: 0, Node: "", ToleratesGpu: false},
					{Name: "p-new-2", Cpu: 1000, Memory: 1000, Gpu: 0, Node: "", ToleratesGpu: false},
				},
				// ExpansionOptionToChoose: &GroupSizeChange{GroupName: "ng2", SizeChange: 2},
				Options: &optionsWithLimitedMaxCores,
				NodeTemplateConfigs: map[string]*NodeTemplateConfig{
					"ng2": {
						NodeGroupName: "ng2",
						MachineType:   "ct4p",
						NodeInfo:      nodeInfo,
					},
				},
			},
			expectedResults: &ScaleTestResults{
				NoScaleUpReason: "max cluster cpu limit reached",
				ScaleUpStatus: ScaleUpStatusInfo{
					PodsRemainUnschedulable: []string{"p-new-1", "p-new-2"},
				},
			},
			isScaleUpOk: false,
		},
		"Atomic ScaleUp max memory limit hit": {
			testConfig: &ScaleUpTestConfig{
				Groups: []NodeGroupConfig{
					{Name: "ng1", MaxSize: 5},
					{Name: "ng2", MaxSize: 3},
				},
				Nodes: []NodeConfig{
					{Name: "n1", Cpu: 900, Memory: 900, Gpu: 0, Ready: true, Group: "ng1"},
				},
				Pods: []PodConfig{
					{Name: "p1", Cpu: 900, Memory: 900, Gpu: 0, Node: "n1", ToleratesGpu: false},
				},
				ExtraPods: []PodConfig{
					{Name: "p-new-1", Cpu: 1000, Memory: 1000, Gpu: 0, Node: "", ToleratesGpu: false},
					{Name: "p-new-2", Cpu: 1000, Memory: 1000, Gpu: 0, Node: "", ToleratesGpu: false},
				},
				// ExpansionOptionToChoose: &GroupSizeChange{GroupName: "ng2", SizeChange: 2},
				Options: &optionsWithLimitedMaxMemory,
				NodeTemplateConfigs: map[string]*NodeTemplateConfig{
					"ng2": {
						NodeGroupName: "ng2",
						MachineType:   "ct4p",
						NodeInfo:      nodeInfo,
					},
				},
			},
			expectedResults: &ScaleTestResults{
				NoScaleUpReason: "max cluster memory limit reached",
				ScaleUpStatus: ScaleUpStatusInfo{
					PodsRemainUnschedulable: []string{"p-new-1", "p-new-2"},
				},
			},
			isScaleUpOk: false,
		},
		"Atomic ScaleUp max nodes count limit hit": {
			testConfig: &ScaleUpTestConfig{
				Groups: []NodeGroupConfig{
					{Name: "ng1", MaxSize: 2},
					{Name: "ng2", MaxSize: 5},
				},
				Nodes: []NodeConfig{
					{Name: "n1", Cpu: 900, Memory: 900, Gpu: 0, Ready: true, Group: "ng1"},
					{Name: "n2", Cpu: 900, Memory: 900, Gpu: 0, Ready: true, Group: "ng1"},
				},
				Pods: []PodConfig{
					{Name: "p1", Cpu: 900, Memory: 900, Gpu: 0, Node: "n1", ToleratesGpu: false},
					{Name: "p2", Cpu: 900, Memory: 900, Gpu: 0, Node: "n1", ToleratesGpu: false},
				},
				ExtraPods: []PodConfig{
					{Name: "p-new-1", Cpu: 1000, Memory: 1000, Gpu: 0, Node: "", ToleratesGpu: false},
					{Name: "p-new-2", Cpu: 1000, Memory: 1000, Gpu: 0, Node: "", ToleratesGpu: false},
				},
				// ExpansionOptionToChoose: &GroupSizeChange{GroupName: "ng2", SizeChange: 2},
				Options: &optionsWithLimitedMaxNodes,
				NodeTemplateConfigs: map[string]*NodeTemplateConfig{
					"ng2": {
						NodeGroupName: "ng2",
						MachineType:   "ct4p",
						NodeInfo:      nodeInfo,
					},
				},
			},
			expectedResults: &ScaleTestResults{
				NoScaleUpReason: "atomic scale-up exceeds cluster node count limit",
				ScaleUpStatus: ScaleUpStatusInfo{
					PodsRemainUnschedulable: []string{"p-new-1", "p-new-2"},
				},
			},
			isScaleUpOk: false,
		},
	}
	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			if tc.isScaleUpOk {
				simpleScaleUpTest(t, tc.testConfig, tc.expectedResults)
			} else {
				simpleNoScaleUpTest(t, tc.testConfig, tc.expectedResults)
			}
		})
	}
}

func TestScaleUpMaxCoresLimitHit(t *testing.T) {
	options := defaultOptions
	options.MaxCoresTotal = 9
	config := &ScaleUpTestConfig{
		Nodes: []NodeConfig{
			{Name: "n1", Cpu: 2000, Memory: 100, Gpu: 0, Ready: true, Group: "ng1"},
			{Name: "n2", Cpu: 4000, Memory: 1000, Gpu: 0, Ready: true, Group: "ng2"},
		},
		Pods: []PodConfig{
			{Name: "p1", Cpu: 1000, Memory: 0, Gpu: 0, Node: "n1", ToleratesGpu: false},
			{Name: "p2", Cpu: 3000, Memory: 0, Gpu: 0, Node: "n2", ToleratesGpu: false},
		},
		ExtraPods: []PodConfig{
			{Name: "p-new-1", Cpu: 2000, Memory: 0, Gpu: 0, Node: "", ToleratesGpu: false},
			{Name: "p-new-2", Cpu: 2000, Memory: 0, Gpu: 0, Node: "", ToleratesGpu: false},
		},
		ExpansionOptionToChoose: &GroupSizeChange{GroupName: "ng1", SizeChange: 2},
		Options:                 &options,
	}
	results := &ScaleTestResults{
		FinalOption: GroupSizeChange{GroupName: "ng1", SizeChange: 1},
		ScaleUpStatus: ScaleUpStatusInfo{
			PodsTriggeredScaleUp: []string{"p-new-1", "p-new-2"},
		},
	}

	simpleScaleUpTest(t, config, results)
}

func TestScaleUpMaxCoresLimitHitWithNotAutoscaledGroup(t *testing.T) {
	options := defaultOptions
	options.MaxCoresTotal = 9
	config := &ScaleUpTestConfig{
		Nodes: []NodeConfig{
			{Name: "n1", Cpu: 2000, Memory: 100, Gpu: 0, Ready: true, Group: "ng1"},
			{Name: "n2", Cpu: 4000, Memory: 1000, Gpu: 0, Ready: true, Group: ""},
		},
		Pods: []PodConfig{
			{Name: "p1", Cpu: 1000, Memory: 0, Gpu: 0, Node: "n1", ToleratesGpu: false},
			{Name: "p2", Cpu: 3000, Memory: 0, Gpu: 0, Node: "n2", ToleratesGpu: false},
		},
		ExtraPods: []PodConfig{
			{Name: "p-new-1", Cpu: 2000, Memory: 0, Gpu: 0, Node: "", ToleratesGpu: false},
			{Name: "p-new-2", Cpu: 2000, Memory: 0, Gpu: 0, Node: "", ToleratesGpu: false},
		},
		ExpansionOptionToChoose: &GroupSizeChange{GroupName: "ng1", SizeChange: 2},
		Options:                 &options,
	}
	results := &ScaleTestResults{
		FinalOption: GroupSizeChange{GroupName: "ng1", SizeChange: 1},
		ScaleUpStatus: ScaleUpStatusInfo{
			PodsTriggeredScaleUp: []string{"p-new-1", "p-new-2"},
		},
	}

	simpleScaleUpTest(t, config, results)
}

func TestScaleUpMaxMemoryLimitHit(t *testing.T) {
	options := defaultOptions
	options.MaxMemoryTotal = 1300 * utils.MiB
	config := &ScaleUpTestConfig{
		Nodes: []NodeConfig{
			{Name: "n1", Cpu: 2000, Memory: 100 * utils.MiB, Gpu: 0, Ready: true, Group: "ng1"},
			{Name: "n2", Cpu: 4000, Memory: 1000 * utils.MiB, Gpu: 0, Ready: true, Group: "ng2"},
		},
		Pods: []PodConfig{
			{Name: "p1", Cpu: 1000, Memory: 0, Gpu: 0, Node: "n1", ToleratesGpu: false},
			{Name: "p2", Cpu: 3000, Memory: 0, Gpu: 0, Node: "n2", ToleratesGpu: false},
		},
		ExtraPods: []PodConfig{
			{Name: "p-new-1", Cpu: 2000, Memory: 100 * utils.MiB, Gpu: 0, Node: "", ToleratesGpu: false},
			{Name: "p-new-2", Cpu: 2000, Memory: 100 * utils.MiB, Gpu: 0, Node: "", ToleratesGpu: false},
			{Name: "p-new-3", Cpu: 2000, Memory: 100 * utils.MiB, Gpu: 0, Node: "", ToleratesGpu: false},
		},
		ExpansionOptionToChoose: &GroupSizeChange{GroupName: "ng1", SizeChange: 3},
		Options:                 &options,
	}
	results := &ScaleTestResults{
		FinalOption: GroupSizeChange{GroupName: "ng1", SizeChange: 2},
		ScaleUpStatus: ScaleUpStatusInfo{
			PodsTriggeredScaleUp: []string{"p-new-1", "p-new-2", "p-new-3"},
		},
	}

	simpleScaleUpTest(t, config, results)
}

func TestScaleUpMaxMemoryLimitHitWithNotAutoscaledGroup(t *testing.T) {
	options := defaultOptions
	options.MaxMemoryTotal = 1300 * utils.MiB
	config := &ScaleUpTestConfig{
		Nodes: []NodeConfig{
			{Name: "n1", Cpu: 2000, Memory: 100 * utils.MiB, Gpu: 0, Ready: true, Group: "ng1"},
			{Name: "n2", Cpu: 4000, Memory: 1000 * utils.MiB, Gpu: 0, Ready: true, Group: ""},
		},
		Pods: []PodConfig{
			{Name: "p1", Cpu: 1000, Memory: 0, Gpu: 0, Node: "n1", ToleratesGpu: false},
			{Name: "p2", Cpu: 3000, Memory: 0, Gpu: 0, Node: "n2", ToleratesGpu: false},
		},
		ExtraPods: []PodConfig{
			{Name: "p-new-1", Cpu: 2000, Memory: 100 * utils.MiB, Gpu: 0, Node: "", ToleratesGpu: false},
			{Name: "p-new-2", Cpu: 2000, Memory: 100 * utils.MiB, Gpu: 0, Node: "", ToleratesGpu: false},
			{Name: "p-new-3", Cpu: 2000, Memory: 100 * utils.MiB, Gpu: 0, Node: "", ToleratesGpu: false},
		},
		ExpansionOptionToChoose: &GroupSizeChange{GroupName: "ng1", SizeChange: 3},
		Options:                 &options,
	}
	results := &ScaleTestResults{
		FinalOption: GroupSizeChange{GroupName: "ng1", SizeChange: 2},
		ScaleUpStatus: ScaleUpStatusInfo{
			PodsTriggeredScaleUp: []string{"p-new-1", "p-new-2", "p-new-3"},
		},
	}

	simpleScaleUpTest(t, config, results)
}

func TestScaleUpTwoGroups(t *testing.T) {
	options := defaultOptions
	options.BalanceSimilarNodeGroups = true
	options.ParallelScaleUp = true
	config := &ScaleUpTestConfig{
		Nodes: []NodeConfig{
			{Name: "ng1-n1", Cpu: 1500, Memory: 1000 * utils.MiB, Ready: true, Group: "ng1"},
			{Name: "ng2-n1", Cpu: 1500, Memory: 1000 * utils.MiB, Ready: true, Group: "ng2"},
		},
		Pods: []PodConfig{
			{Name: "p1", Cpu: 1400, Node: "ng1-n1"},
			{Name: "p2", Cpu: 1400, Node: "ng2-n1"},
		},
		ExtraPods: []PodConfig{
			{Name: "p3", Cpu: 1400},
			{Name: "p4", Cpu: 1400},
		},
		Options: &options,
	}
	testCases := []struct {
		desc     string
		parallel bool
	}{
		{
			desc:     "synchronous scale up",
			parallel: false,
		},
		{
			desc:     "parallel scale up",
			parallel: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			config.Options.ParallelScaleUp = tc.parallel
			result := runSimpleScaleUpTest(t, config)
			assert.True(t, result.ScaleUpStatus.WasSuccessful())
			assert.Nil(t, result.ScaleUpError)
			assert.Equal(t, result.GroupTargetSizes, map[string]int{
				"ng1": 2,
				"ng2": 2,
			})
			assert.ElementsMatch(t, result.ScaleUpStatus.PodsTriggeredScaleUp, []string{"p3", "p4"})
		})
	}
}

func TestCloudProviderFailingToScaleUpGroups(t *testing.T) {
	options := defaultOptions
	options.BalanceSimilarNodeGroups = true
	config := &ScaleUpTestConfig{
		Groups: []NodeGroupConfig{
			{Name: "ng1", MaxSize: 2},
			{Name: "ng2", MaxSize: 2},
		},
		Nodes: []NodeConfig{
			{Name: "ng1-n1", Cpu: 1500, Memory: 1000 * utils.MiB, Ready: true, Group: "ng1"},
			{Name: "ng2-n1", Cpu: 1500, Memory: 1000 * utils.MiB, Ready: true, Group: "ng2"},
		},
		Pods: []PodConfig{
			{Name: "p1", Cpu: 1400, Node: "ng1-n1"},
			{Name: "p2", Cpu: 1400, Node: "ng2-n1"},
		},
		ExtraPods: []PodConfig{
			{Name: "p3", Cpu: 1400},
			{Name: "p4", Cpu: 1400},
		},
		Options: &options,
	}
	failAlwaysScaleUp := func(group string, i int) error {
		return fmt.Errorf("provider error for: %s", group)
	}
	failOnceScaleUp := func() testprovider.OnScaleUpFunc {
		var executed atomic.Bool
		return func(group string, _ int) error {
			if !executed.Swap(true) {
				return fmt.Errorf("provider error for: %s", group)
			}
			return nil
		}
	}
	testCases := []struct {
		desc                     string
		parallel                 bool
		onScaleUp                testprovider.OnScaleUpFunc
		expectConcurrentErrors   bool
		expectedTotalTargetSizes int
	}{
		{
			desc:                     "synchronous scale up - two failures",
			parallel:                 false,
			onScaleUp:                failAlwaysScaleUp,
			expectConcurrentErrors:   false,
			expectedTotalTargetSizes: 3, // first error stops scale up process
		},
		{
			desc:                     "parallel scale up - two failures",
			parallel:                 true,
			onScaleUp:                failAlwaysScaleUp,
			expectConcurrentErrors:   true,
			expectedTotalTargetSizes: 4,
		},
		{
			desc:                     "synchronous scale up - one failure",
			parallel:                 false,
			onScaleUp:                failOnceScaleUp(),
			expectConcurrentErrors:   false,
			expectedTotalTargetSizes: 3,
		},
		{
			desc:                     "parallel scale up - one failure",
			parallel:                 true,
			onScaleUp:                failOnceScaleUp(),
			expectConcurrentErrors:   false,
			expectedTotalTargetSizes: 4,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			config.Options.ParallelScaleUp = tc.parallel
			config.OnScaleUp = tc.onScaleUp
			result := runSimpleScaleUpTest(t, config)
			assert.False(t, result.ScaleUpStatus.WasSuccessful())
			assert.Equal(t, errors.CloudProviderError, result.ScaleUpError.Type())
			assert.Equal(t, tc.expectedTotalTargetSizes, result.GroupTargetSizes["ng1"]+result.GroupTargetSizes["ng2"])
			assert.Equal(t, tc.expectConcurrentErrors, strings.Contains(result.ScaleUpError.Error(), "...and other concurrent errors"))
		})
	}
}

func TestScaleUpCapToMaxTotalNodesLimit(t *testing.T) {
	options := defaultOptions
	options.MaxNodesTotal = 3
	config := &ScaleUpTestConfig{
		Nodes: []NodeConfig{
			{Name: "n1", Cpu: 2000, Memory: 100 * utils.MiB, Gpu: 0, Ready: true, Group: "ng1"},
			{Name: "n2", Cpu: 4000, Memory: 1000 * utils.MiB, Gpu: 0, Ready: true, Group: "ng2"},
		},
		Pods: []PodConfig{
			{Name: "p1", Cpu: 1000, Memory: 0, Gpu: 0, Node: "n1", ToleratesGpu: false},
			{Name: "p2", Cpu: 3000, Memory: 0, Gpu: 0, Node: "n2", ToleratesGpu: false},
		},
		ExtraPods: []PodConfig{
			{Name: "p-new-1", Cpu: 4000, Memory: 100 * utils.MiB, Gpu: 0, Node: "", ToleratesGpu: false},
			{Name: "p-new-2", Cpu: 4000, Memory: 100 * utils.MiB, Gpu: 0, Node: "", ToleratesGpu: false},
			{Name: "p-new-3", Cpu: 4000, Memory: 100 * utils.MiB, Gpu: 0, Node: "", ToleratesGpu: false},
		},
		ExpansionOptionToChoose: &GroupSizeChange{GroupName: "ng2", SizeChange: 3},
		Options:                 &options,
	}
	results := &ScaleTestResults{
		FinalOption: GroupSizeChange{GroupName: "ng2", SizeChange: 1},
		ScaleUpStatus: ScaleUpStatusInfo{
			PodsTriggeredScaleUp: []string{"p-new-1", "p-new-2", "p-new-3"},
		},
	}

	simpleScaleUpTest(t, config, results)
}

func TestScaleUpCapToMaxTotalNodesLimitWithNotAutoscaledGroup(t *testing.T) {
	options := defaultOptions
	options.MaxNodesTotal = 3
	config := &ScaleUpTestConfig{
		Nodes: []NodeConfig{
			{Name: "n1", Cpu: 2000, Memory: 100 * utils.MiB, Gpu: 0, Ready: true, Group: ""},
			{Name: "n2", Cpu: 4000, Memory: 1000 * utils.MiB, Gpu: 0, Ready: true, Group: "ng2"},
		},
		Pods: []PodConfig{
			{Name: "p1", Cpu: 1000, Memory: 0, Gpu: 0, Node: "n1", ToleratesGpu: false},
			{Name: "p2", Cpu: 3000, Memory: 0, Gpu: 0, Node: "n2", ToleratesGpu: false},
		},
		ExtraPods: []PodConfig{
			{Name: "p-new-1", Cpu: 4000, Memory: 100 * utils.MiB, Gpu: 0, Node: "", ToleratesGpu: false},
			{Name: "p-new-2", Cpu: 4000, Memory: 100 * utils.MiB, Gpu: 0, Node: "", ToleratesGpu: false},
			{Name: "p-new-3", Cpu: 4000, Memory: 100 * utils.MiB, Gpu: 0, Node: "", ToleratesGpu: false},
		},
		ExpansionOptionToChoose: &GroupSizeChange{GroupName: "ng2", SizeChange: 3},
		Options:                 &options,
	}
	results := &ScaleTestResults{
		FinalOption: GroupSizeChange{GroupName: "ng2", SizeChange: 1},
		ScaleUpStatus: ScaleUpStatusInfo{
			PodsTriggeredScaleUp: []string{"p-new-1", "p-new-2", "p-new-3"},
		},
	}

	simpleScaleUpTest(t, config, results)
}

func TestWillConsiderGpuAndStandardPoolForPodWhichDoesNotRequireGpu(t *testing.T) {
	options := defaultOptions
	options.MaxNodesTotal = 100
	config := &ScaleUpTestConfig{
		Nodes: []NodeConfig{
			{Name: "gpu-node-1", Cpu: 2000, Memory: 1000 * utils.MiB, Gpu: 1, Ready: true, Group: "gpu-pool"},
			{Name: "std-node-1", Cpu: 2000, Memory: 1000 * utils.MiB, Gpu: 0, Ready: true, Group: "std-pool"},
		},
		Pods: []PodConfig{
			{Name: "gpu-pod-1", Cpu: 2000, Memory: 1000 * utils.MiB, Gpu: 1, Node: "gpu-node-1", ToleratesGpu: true},
			{Name: "std-pod-1", Cpu: 2000, Memory: 1000 * utils.MiB, Gpu: 0, Node: "std-node-1", ToleratesGpu: false},
		},
		ExtraPods: []PodConfig{
			{Name: "extra-std-pod", Cpu: 2000, Memory: 1000 * utils.MiB, Gpu: 0, Node: "", ToleratesGpu: true},
		},
		ExpansionOptionToChoose: &GroupSizeChange{GroupName: "std-pool", SizeChange: 1},
		Options:                 &options,
	}
	results := &ScaleTestResults{
		FinalOption: GroupSizeChange{GroupName: "std-pool", SizeChange: 1},
		ExpansionOptions: []GroupSizeChange{
			{GroupName: "std-pool", SizeChange: 1},
			{GroupName: "gpu-pool", SizeChange: 1},
		},
		ScaleUpStatus: ScaleUpStatusInfo{
			PodsTriggeredScaleUp: []string{"extra-std-pod"},
		},
	}

	simpleScaleUpTest(t, config, results)
}

func TestWillConsiderOnlyGpuPoolForPodWhichDoesRequiresGpu(t *testing.T) {
	options := defaultOptions
	options.MaxNodesTotal = 100
	config := &ScaleUpTestConfig{
		Nodes: []NodeConfig{
			{Name: "gpu-node-1", Cpu: 2000, Memory: 1000 * utils.MiB, Gpu: 1, Ready: true, Group: "gpu-pool"},
			{Name: "std-node-1", Cpu: 2000, Memory: 1000 * utils.MiB, Gpu: 0, Ready: true, Group: "std-pool"},
		},
		Pods: []PodConfig{
			{Name: "gpu-pod-1", Cpu: 2000, Memory: 1000 * utils.MiB, Gpu: 1, Node: "gpu-node-1", ToleratesGpu: true},
			{Name: "std-pod-1", Cpu: 2000, Memory: 1000 * utils.MiB, Gpu: 0, Node: "std-node-1", ToleratesGpu: false},
		},
		ExtraPods: []PodConfig{
			{Name: "extra-gpu-pod", Cpu: 2000, Memory: 1000 * utils.MiB, Gpu: 1, Node: "", ToleratesGpu: true},
		},
		ExpansionOptionToChoose: &GroupSizeChange{GroupName: "gpu-pool", SizeChange: 1},
		Options:                 &options,
	}
	results := &ScaleTestResults{
		FinalOption: GroupSizeChange{GroupName: "gpu-pool", SizeChange: 1},
		ExpansionOptions: []GroupSizeChange{
			{GroupName: "gpu-pool", SizeChange: 1},
		},
		ScaleUpStatus: ScaleUpStatusInfo{
			PodsTriggeredScaleUp: []string{"extra-gpu-pod"},
		},
	}

	simpleScaleUpTest(t, config, results)
}

func TestWillConsiderAllPoolsWhichFitTwoPodsRequiringGpus(t *testing.T) {
	options := defaultOptions
	options.MaxNodesTotal = 100
	config := &ScaleUpTestConfig{
		Nodes: []NodeConfig{
			{Name: "gpu-1-node-1", Cpu: 2000, Memory: 1000 * utils.MiB, Gpu: 1, Ready: true, Group: "gpu-1-pool"},
			{Name: "gpu-2-node-1", Cpu: 2000, Memory: 1000 * utils.MiB, Gpu: 2, Ready: true, Group: "gpu-2-pool"},
			{Name: "gpu-4-node-1", Cpu: 2000, Memory: 1000 * utils.MiB, Gpu: 4, Ready: true, Group: "gpu-4-pool"},
			{Name: "std-node-1", Cpu: 2000, Memory: 1000 * utils.MiB, Gpu: 0, Ready: true, Group: "std-pool"},
		},
		Pods: []PodConfig{
			{Name: "gpu-pod-1", Cpu: 2000, Memory: 1000 * utils.MiB, Gpu: 1, Node: "gpu-1-node-1", ToleratesGpu: true},
			{Name: "gpu-pod-2", Cpu: 2000, Memory: 1000 * utils.MiB, Gpu: 2, Node: "gpu-2-node-1", ToleratesGpu: true},
			{Name: "gpu-pod-3", Cpu: 2000, Memory: 1000 * utils.MiB, Gpu: 4, Node: "gpu-4-node-1", ToleratesGpu: true},
			{Name: "std-pod-1", Cpu: 2000, Memory: 1000 * utils.MiB, Gpu: 0, Node: "std-node-1", ToleratesGpu: false},
		},
		ExtraPods: []PodConfig{
			{Name: "extra-gpu-pod-1", Cpu: 1, Memory: 1 * utils.MiB, Gpu: 1, Node: "", ToleratesGpu: true}, // CPU and mem negligible
			{Name: "extra-gpu-pod-2", Cpu: 1, Memory: 1 * utils.MiB, Gpu: 1, Node: "", ToleratesGpu: true}, // CPU and mem negligible
			{Name: "extra-gpu-pod-3", Cpu: 1, Memory: 1 * utils.MiB, Gpu: 1, Node: "", ToleratesGpu: true}, // CPU and mem negligible
		},
		ExpansionOptionToChoose: &GroupSizeChange{GroupName: "gpu-1-pool", SizeChange: 3},
		Options:                 &options,
	}
	results := &ScaleTestResults{
		FinalOption: GroupSizeChange{GroupName: "gpu-1-pool", SizeChange: 3},
		ExpansionOptions: []GroupSizeChange{
			{GroupName: "gpu-1-pool", SizeChange: 3},
			{GroupName: "gpu-2-pool", SizeChange: 2},
			{GroupName: "gpu-4-pool", SizeChange: 1},
		},
		ScaleUpStatus: ScaleUpStatusInfo{
			PodsTriggeredScaleUp: []string{"extra-gpu-pod-1", "extra-gpu-pod-2", "extra-gpu-pod-3"},
		},
	}

	simpleScaleUpTest(t, config, results)
}

// No scale up scenarios.
func TestNoScaleUpMaxCoresLimitHit(t *testing.T) {
	options := defaultOptions
	options.MaxCoresTotal = 7
	options.MaxMemoryTotal = 1150
	config := &ScaleUpTestConfig{
		Nodes: []NodeConfig{
			{Name: "n1", Cpu: 2000, Memory: 100, Gpu: 0, Ready: true, Group: "ng1"},
			{Name: "n2", Cpu: 4000, Memory: 1000, Gpu: 0, Ready: true, Group: "ng2"},
		},
		Pods: []PodConfig{
			{Name: "p1", Cpu: 1000, Memory: 0, Gpu: 0, Node: "n1", ToleratesGpu: false},
			{Name: "p2", Cpu: 3000, Memory: 0, Gpu: 0, Node: "n2", ToleratesGpu: false},
		},
		ExtraPods: []PodConfig{
			{Name: "p-new-1", Cpu: 2000, Memory: 0, Gpu: 0, Node: "", ToleratesGpu: false},
			{Name: "p-new-2", Cpu: 2000, Memory: 0, Gpu: 0, Node: "", ToleratesGpu: false},
		},
		Options: &options,
	}
	results := &ScaleTestResults{
		NoScaleUpReason: "max cluster cpu, memory limit reached",
		ScaleUpStatus: ScaleUpStatusInfo{
			PodsRemainUnschedulable: []string{"p-new-1", "p-new-2"},
		},
	}

	simpleNoScaleUpTest(t, config, results)
}

func TestNoCreateNodeGroupMaxCoresLimitHit(t *testing.T) {
	options := defaultOptions
	options.MaxCoresTotal = 7
	options.MaxMemoryTotal = 100000
	options.NodeAutoprovisioningEnabled = true

	largeNode := BuildTestNode("n", 8000, 8000)
	SetNodeReadyState(largeNode, true, time.Time{})
	largeNodeInfo := schedulerframework.NewNodeInfo()
	largeNodeInfo.SetNode(largeNode)

	config := &ScaleUpTestConfig{
		EnableAutoprovisioning: true,
		Nodes: []NodeConfig{
			{Name: "n1", Cpu: 2000, Memory: 1000, Gpu: 0, Ready: true, Group: "ng-small"},
		},
		Pods: []PodConfig{},
		ExtraPods: []PodConfig{
			{Name: "large-pod", Cpu: 8000, Memory: 0, Gpu: 0, Node: "", ToleratesGpu: false},
		},
		Options: &options,
		NodeTemplateConfigs: map[string]*NodeTemplateConfig{
			"n1-standard-8": {
				NodeGroupName: "autoprovisioned-n1-standard-8",
				MachineType:   "n1-standard-8",
				NodeInfo:      largeNodeInfo,
			},
		},
	}
	results := &ScaleTestResults{
		NoScaleUpReason: "Insufficient cpu",
		ScaleUpStatus: ScaleUpStatusInfo{
			PodsRemainUnschedulable: []string{"large-pod"},
		},
	}

	simpleNoScaleUpTest(t, config, results)
}

func TestAllOrNothing(t *testing.T) {
	options := defaultOptions

	extraPods := []PodConfig{}
	extraPodNames := []string{}
	for i := 0; i < 11; i++ {
		podName := fmt.Sprintf("pod-%d", i)
		extraPods = append(extraPods, PodConfig{Name: podName, Cpu: 1000, Memory: 100})
		extraPodNames = append(extraPodNames, podName)
	}

	config := &ScaleUpTestConfig{
		Nodes: []NodeConfig{
			{Name: "n1", Cpu: 1000, Memory: 1000, Gpu: 0, Ready: true, Group: "ng"},
		},
		Pods:         []PodConfig{},
		ExtraPods:    extraPods,
		Options:      &options,
		AllOrNothing: true,
	}

	result := &ScaleTestResults{
		NoScaleUpReason: "all-or-nothing",
		ScaleUpStatus: ScaleUpStatusInfo{
			PodsRemainUnschedulable: extraPodNames,
		},
	}

	simpleNoScaleUpTest(t, config, result)
}

func simpleScaleUpTest(t *testing.T, config *ScaleUpTestConfig, expectedResults *ScaleTestResults) {
	results := runSimpleScaleUpTest(t, config)
	assert.NotNil(t, results.GroupSizeChanges, "Expected scale up event")
	assert.NotNil(t, results.GroupSizeChanges[0], "Expected scale up event")
	assert.Equal(t, expectedResults.FinalOption, results.GroupSizeChanges[0])
	assert.True(t, results.ScaleUpStatus.WasSuccessful())
	nodeEventSeen := false
	for _, event := range results.Events {
		if strings.Contains(event, "TriggeredScaleUp") && strings.Contains(event, expectedResults.FinalOption.GroupName) {
			nodeEventSeen = true
		}
		if len(expectedResults.ScaleUpStatus.PodsRemainUnschedulable) == 0 {
			assert.NotRegexp(t, regexp.MustCompile("NotTriggerScaleUp"), event)
		}
	}
	assert.True(t, nodeEventSeen)
	if len(expectedResults.ExpansionOptions) > 0 {
		// Empty ExpansionOptions means we do not want to do any assertions
		// on contents of actual scaleUp options
		if config.ExpansionOptionToChoose != nil {
			// Check that option to choose is part of expected options.
			assert.Contains(t, expectedResults.ExpansionOptions, *config.ExpansionOptionToChoose, "final expected expansion option must be in expected expansion options")
			assert.Contains(t, results.ExpansionOptions, *config.ExpansionOptionToChoose, "final expected expansion option must be in expected expansion options")
		}
		assert.ElementsMatch(t, results.ExpansionOptions, expectedResults.ExpansionOptions,
			"actual and expected expansion options should be the same")
	}
	if expectedResults.GroupTargetSizes != nil {
		assert.Equal(t, expectedResults.GroupTargetSizes, results.GroupTargetSizes)
	}
	assert.ElementsMatch(t, results.ScaleUpStatus.PodsTriggeredScaleUp, expectedResults.ScaleUpStatus.PodsTriggeredScaleUp,
		"actual and expected triggering pods should be the same")
	assert.ElementsMatch(t, results.ScaleUpStatus.PodsRemainUnschedulable, expectedResults.ScaleUpStatus.PodsRemainUnschedulable,
		"actual and expected remaining pods should be the same")
	assert.ElementsMatch(t, results.ScaleUpStatus.PodsAwaitEvaluation, expectedResults.ScaleUpStatus.PodsAwaitEvaluation,
		"actual and expected awaiting evaluation pods should be the same")
}

func simpleNoScaleUpTest(t *testing.T, config *ScaleUpTestConfig, expectedResults *ScaleTestResults) {
	results := runSimpleScaleUpTest(t, config)
	assert.Nil(t, results.GroupSizeChanges)
	assert.False(t, results.ScaleUpStatus.WasSuccessful())
	noScaleUpEventSeen := false
	for _, event := range results.Events {
		if strings.Contains(event, "NotTriggerScaleUp") {
			if strings.Contains(event, expectedResults.NoScaleUpReason) {
				noScaleUpEventSeen = true
			} else {
				// Surprisingly useful for debugging.
				fmt.Println("Event:", event)
			}
		}
		assert.NotRegexp(t, regexp.MustCompile("TriggeredScaleUp"), event)
	}
	assert.True(t, noScaleUpEventSeen)
	assert.ElementsMatch(t, results.ScaleUpStatus.PodsTriggeredScaleUp, expectedResults.ScaleUpStatus.PodsTriggeredScaleUp,
		"actual and expected triggering pods should be the same")
	assert.ElementsMatch(t, results.ScaleUpStatus.PodsRemainUnschedulable, expectedResults.ScaleUpStatus.PodsRemainUnschedulable,
		"actual and expected remaining pods should be the same")
	assert.ElementsMatch(t, results.ScaleUpStatus.PodsAwaitEvaluation, expectedResults.ScaleUpStatus.PodsAwaitEvaluation,
		"actual and expected awaiting evaluation pods should be the same")
}

func runSimpleScaleUpTest(t *testing.T, config *ScaleUpTestConfig) *ScaleUpTestResult {
	now := time.Now()
	groupSizeChangesChannel := make(chan GroupSizeChange, 20)
	groupNodes := make(map[string][]*apiv1.Node)

	// build nodes
	nodes := make([]*apiv1.Node, 0, len(config.Nodes))
	for _, n := range config.Nodes {
		node := buildTestNode(n, now)
		nodes = append(nodes, node)
		if n.Group != "" {
			groupNodes[n.Group] = append(groupNodes[n.Group], node)
		}
	}

	// build and setup pods
	pods := make([]*apiv1.Pod, len(config.Pods))
	for i, p := range config.Pods {
		pods[i] = buildTestPod(p)
	}
	extraPods := make([]*apiv1.Pod, len(config.ExtraPods))
	for i, p := range config.ExtraPods {
		extraPods[i] = buildTestPod(p)
	}
	podLister := kube_util.NewTestPodLister(pods)
	listers := kube_util.NewListerRegistry(nil, nil, podLister, nil, nil, nil, nil, nil, nil)

	// setup node groups
	var provider *testprovider.TestCloudProvider
	onScaleUpFunc := func(nodeGroup string, increase int) error {
		groupSizeChangesChannel <- GroupSizeChange{GroupName: nodeGroup, SizeChange: increase}
		if config.OnScaleUp != nil {
			return config.OnScaleUp(nodeGroup, increase)
		}
		return nil
	}
	onCreateGroupFunc := func(nodeGroup string) error {
		if config.OnCreateGroup != nil {
			return config.OnCreateGroup(nodeGroup)
		}
		return fmt.Errorf("unexpected node group create: OnCreateGroup not defined")
	}
	if len(config.NodeTemplateConfigs) > 0 {
		machineTypes := []string{}
		machineTemplates := map[string]*schedulerframework.NodeInfo{}
		for _, ntc := range config.NodeTemplateConfigs {
			machineTypes = append(machineTypes, ntc.MachineType)
			machineTemplates[ntc.NodeGroupName] = ntc.NodeInfo
			machineTemplates[ntc.MachineType] = ntc.NodeInfo
		}
		provider = testprovider.NewTestAutoprovisioningCloudProvider(onScaleUpFunc, nil, onCreateGroupFunc, nil, machineTypes, machineTemplates)
	} else {
		provider = testprovider.NewTestCloudProvider(onScaleUpFunc, nil)
	}
	options := defaultOptions
	if config.Options != nil {
		options = *config.Options
	}
	resourceLimiter := cloudprovider.NewResourceLimiter(
		map[string]int64{cloudprovider.ResourceNameCores: options.MinCoresTotal, cloudprovider.ResourceNameMemory: options.MinMemoryTotal},
		map[string]int64{cloudprovider.ResourceNameCores: options.MaxCoresTotal, cloudprovider.ResourceNameMemory: options.MaxMemoryTotal})
	provider.SetResourceLimiter(resourceLimiter)
	groupConfigs := make(map[string]*NodeGroupConfig)
	for _, group := range config.Groups {
		groupConfigs[group.Name] = &group
	}
	for name, nodesInGroup := range groupNodes {
		groupConfig := groupConfigs[name]
		if groupConfig == nil {
			groupConfig = &NodeGroupConfig{
				Name:    name,
				MinSize: 1,
				MaxSize: 10,
			}
		}
		provider.AddNodeGroup(name, groupConfig.MinSize, groupConfig.MaxSize, len(nodesInGroup))
		for _, n := range nodesInGroup {
			provider.AddNode(name, n)
		}
	}

	// Build node groups without any nodes
	for name, ng := range groupConfigs {
		if provider.GetNodeGroup(name) == nil {
			tng := provider.BuildNodeGroup(name, ng.MinSize, ng.MaxSize, 0, true, false, config.NodeTemplateConfigs[name].MachineType, &options.NodeGroupDefaults)
			provider.InsertNodeGroup(tng)
		}
	}
	// build orchestrator
	context, err := NewScaleTestAutoscalingContext(options, &fake.Clientset{}, listers, provider, nil, nil)
	assert.NoError(t, err)
	nodeInfos, err := nodeinfosprovider.NewDefaultTemplateNodeInfoProvider(nil, false).
		Process(&context, nodes, []*appsv1.DaemonSet{}, taints.TaintConfig{}, now)
	assert.NoError(t, err)
	clusterState := clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{}, context.LogRecorder, NewBackoff(), nodegroupconfig.NewDefaultNodeGroupConfigProcessor(options.NodeGroupDefaults), asyncnodegroups.NewDefaultAsyncNodeGroupStateChecker())
	clusterState.UpdateNodes(nodes, nodeInfos, time.Now())
	processors := NewTestProcessors(&context)
	processors.ScaleStateNotifier.Register(clusterState)
	if config.EnableAutoprovisioning {
		processors.NodeGroupListProcessor = &MockAutoprovisioningNodeGroupListProcessor{T: t}
		processors.NodeGroupManager = &MockAutoprovisioningNodeGroupManager{T: t, ExtraGroups: 0}
	}
	orchestrator := New()
	orchestrator.Initialize(&context, processors, clusterState, newEstimatorBuilder(), taints.TaintConfig{})
	expander := NewMockRepotingStrategy(t, config.ExpansionOptionToChoose)
	context.ExpanderStrategy = expander

	// scale up
	scaleUpStatus, scaleUpErr := orchestrator.ScaleUp(extraPods, nodes, []*appsv1.DaemonSet{}, nodeInfos, config.AllOrNothing)
	processors.ScaleUpStatusProcessor.Process(&context, scaleUpStatus)

	// aggregate group size changes
	close(groupSizeChangesChannel)
	var groupSizeChanges []GroupSizeChange
	for change := range groupSizeChangesChannel {
		groupSizeChanges = append(groupSizeChanges, change)
	}

	// aggregate events
	eventsChannel := context.Recorder.(*kube_record.FakeRecorder).Events
	close(eventsChannel)
	events := []string{}
	for event := range eventsChannel {
		events = append(events, event)
	}

	// build target sizes
	targetSizes := make(map[string]int)
	for _, group := range provider.NodeGroups() {
		targetSizes[group.Id()], _ = group.TargetSize()
	}

	return &ScaleUpTestResult{
		ScaleUpError:     scaleUpErr,
		ScaleUpStatus:    simplifyScaleUpStatus(scaleUpStatus),
		GroupSizeChanges: groupSizeChanges,
		Events:           events,
		GroupTargetSizes: targetSizes,
		ExpansionOptions: expander.LastInputOptions(),
	}
}

func buildTestNode(n NodeConfig, now time.Time) *apiv1.Node {
	node := BuildTestNode(n.Name, n.Cpu, n.Memory)
	if n.Gpu > 0 {
		AddGpusToNode(node, n.Gpu)
	}
	SetNodeReadyState(node, n.Ready, now.Add(-2*time.Minute))
	return node
}

func buildTestPod(p PodConfig) *apiv1.Pod {
	pod := BuildTestPod(p.Name, p.Cpu, p.Memory)
	if p.Gpu > 0 {
		RequestGpuForPod(pod, p.Gpu)
	}
	if p.ToleratesGpu {
		TolerateGpuForPod(pod)
	}
	if p.Node != "" {
		pod.Spec.NodeName = p.Node
	}
	return pod
}

func TestScaleUpUnhealthy(t *testing.T) {
	now := time.Now()
	someTimeAgo := now.Add(-2 * time.Minute)
	n1 := BuildTestNode("n1", 100, 1000)
	SetNodeReadyState(n1, true, someTimeAgo)
	n2 := BuildTestNode("n2", 1000, 1000)
	SetNodeReadyState(n2, true, someTimeAgo)

	p1 := BuildTestPod("p1", 80, 0)
	p2 := BuildTestPod("p2", 800, 0)
	p1.Spec.NodeName = "n1"
	p2.Spec.NodeName = "n2"

	podLister := kube_util.NewTestPodLister([]*apiv1.Pod{p1, p2})
	listers := kube_util.NewListerRegistry(nil, nil, podLister, nil, nil, nil, nil, nil, nil)

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
	context, err := NewScaleTestAutoscalingContext(options, &fake.Clientset{}, listers, provider, nil, nil)
	assert.NoError(t, err)

	nodes := []*apiv1.Node{n1, n2}
	nodeInfos, _ := nodeinfosprovider.NewDefaultTemplateNodeInfoProvider(nil, false).Process(&context, nodes, []*appsv1.DaemonSet{}, taints.TaintConfig{}, now)
	clusterState := clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{}, context.LogRecorder, NewBackoff(), nodegroupconfig.NewDefaultNodeGroupConfigProcessor(config.NodeGroupAutoscalingOptions{MaxNodeProvisionTime: 15 * time.Minute}), asyncnodegroups.NewDefaultAsyncNodeGroupStateChecker())
	clusterState.UpdateNodes(nodes, nodeInfos, time.Now())
	p3 := BuildTestPod("p-new", 550, 0)

	processors := NewTestProcessors(&context)
	suOrchestrator := New()
	suOrchestrator.Initialize(&context, processors, clusterState, newEstimatorBuilder(), taints.TaintConfig{})
	scaleUpStatus, err := suOrchestrator.ScaleUp([]*apiv1.Pod{p3}, nodes, []*appsv1.DaemonSet{}, nodeInfos, false)

	assert.NoError(t, err)
	// Node group is unhealthy.
	assert.False(t, scaleUpStatus.WasSuccessful())
}

func TestBinpackingLimiter(t *testing.T) {
	n1 := BuildTestNode("n1", 1000, 1000)
	n2 := BuildTestNode("n2", 100000, 100000)
	now := time.Now()

	SetNodeReadyState(n1, true, now.Add(-2*time.Minute))
	SetNodeReadyState(n2, true, now.Add(-2*time.Minute))

	nodes := []*apiv1.Node{n1, n2}

	podLister := kube_util.NewTestPodLister([]*apiv1.Pod{})
	listers := kube_util.NewListerRegistry(nil, nil, podLister, nil, nil, nil, nil, nil, nil)

	provider := testprovider.NewTestCloudProvider(func(nodeGroup string, increase int) error {
		return nil
	}, nil)

	options := defaultOptions
	provider.AddNodeGroup("ng1", 1, 10, 1)
	provider.AddNode("ng1", n1)
	provider.AddNodeGroup("ng2", 1, 10, 1)
	provider.AddNode("ng2", n2)
	assert.NotNil(t, provider)

	context, err := NewScaleTestAutoscalingContext(options, &fake.Clientset{}, listers, provider, nil, nil)
	assert.NoError(t, err)

	nodeInfos, err := nodeinfosprovider.NewDefaultTemplateNodeInfoProvider(nil, false).
		Process(&context, nodes, []*appsv1.DaemonSet{}, taints.TaintConfig{}, now)
	assert.NoError(t, err)

	clusterState := clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{}, context.LogRecorder, NewBackoff(), nodegroupconfig.NewDefaultNodeGroupConfigProcessor(config.NodeGroupAutoscalingOptions{MaxNodeProvisionTime: 15 * time.Minute}), asyncnodegroups.NewDefaultAsyncNodeGroupStateChecker())
	clusterState.UpdateNodes(nodes, nodeInfos, time.Now())

	extraPod := BuildTestPod("p-new", 500, 0)

	processors := NewTestProcessors(&context)

	// We should stop binpacking after finding expansion option from first node group.
	processors.BinpackingLimiter = &MockBinpackingLimiter{}

	suOrchestrator := New()
	suOrchestrator.Initialize(&context, processors, clusterState, newEstimatorBuilder(), taints.TaintConfig{})

	expander := NewMockRepotingStrategy(t, nil)
	context.ExpanderStrategy = expander

	scaleUpStatus, err := suOrchestrator.ScaleUp([]*apiv1.Pod{extraPod}, nodes, []*appsv1.DaemonSet{}, nodeInfos, false)
	processors.ScaleUpStatusProcessor.Process(&context, scaleUpStatus)
	assert.NoError(t, err)
	assert.True(t, scaleUpStatus.WasSuccessful())

	expansionOptions := expander.LastInputOptions()
	// Only 1 expansion option should be there. Without BinpackingLimiter there will be 2.
	assert.True(t, len(expansionOptions) == 1)
}

func TestScaleUpNoHelp(t *testing.T) {
	n1 := BuildTestNode("n1", 100, 1000)
	now := time.Now()
	SetNodeReadyState(n1, true, now.Add(-2*time.Minute))

	p1 := BuildTestPod("p1", 80, 0)
	p1.Spec.NodeName = "n1"

	podLister := kube_util.NewTestPodLister([]*apiv1.Pod{p1})
	listers := kube_util.NewListerRegistry(nil, nil, podLister, nil, nil, nil, nil, nil, nil)

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
	context, err := NewScaleTestAutoscalingContext(options, &fake.Clientset{}, listers, provider, nil, nil)
	assert.NoError(t, err)

	nodes := []*apiv1.Node{n1}
	nodeInfos, _ := nodeinfosprovider.NewDefaultTemplateNodeInfoProvider(nil, false).Process(&context, nodes, []*appsv1.DaemonSet{}, taints.TaintConfig{}, now)
	clusterState := clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{}, context.LogRecorder, NewBackoff(), nodegroupconfig.NewDefaultNodeGroupConfigProcessor(config.NodeGroupAutoscalingOptions{MaxNodeProvisionTime: 15 * time.Minute}), asyncnodegroups.NewDefaultAsyncNodeGroupStateChecker())
	clusterState.UpdateNodes(nodes, nodeInfos, time.Now())
	p3 := BuildTestPod("p-new", 500, 0)

	processors := NewTestProcessors(&context)
	suOrchestrator := New()
	suOrchestrator.Initialize(&context, processors, clusterState, newEstimatorBuilder(), taints.TaintConfig{})
	scaleUpStatus, err := suOrchestrator.ScaleUp([]*apiv1.Pod{p3}, nodes, []*appsv1.DaemonSet{}, nodeInfos, false)
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

type constNodeGroupSetProcessor struct {
	similarNodeGroups []cloudprovider.NodeGroup
}

func (p *constNodeGroupSetProcessor) FindSimilarNodeGroups(_ *context.AutoscalingContext, _ cloudprovider.NodeGroup, _ map[string]*schedulerframework.NodeInfo) ([]cloudprovider.NodeGroup, errors.AutoscalerError) {
	return p.similarNodeGroups, nil
}

func (p *constNodeGroupSetProcessor) BalanceScaleUpBetweenGroups(_ *context.AutoscalingContext, _ []cloudprovider.NodeGroup, _ int) ([]nodegroupset.ScaleUpInfo, errors.AutoscalerError) {
	return nil, nil
}

func (p *constNodeGroupSetProcessor) CleanUp() {}

func TestComputeSimilarNodeGroups(t *testing.T) {
	podGroup1 := estimator.PodEquivalenceGroup{Pods: []*v1.Pod{BuildTestPod("p1", 100, 1000)}}
	podGroup2 := estimator.PodEquivalenceGroup{Pods: []*v1.Pod{BuildTestPod("p2", 100, 1000)}}
	podGroup3 := estimator.PodEquivalenceGroup{Pods: []*v1.Pod{BuildTestPod("p3", 100, 1000)}}

	testCases := []struct {
		name                  string
		nodeGroup             string
		similarNodeGroups     []string
		otherNodeGroups       []string
		balancingEnabled      bool
		schedulablePodGroups  map[string][]estimator.PodEquivalenceGroup
		wantSimilarNodeGroups []string
	}{
		{
			name:                  "no similar node groups",
			nodeGroup:             "ng1",
			otherNodeGroups:       []string{"pg1", "pg2"},
			balancingEnabled:      true,
			wantSimilarNodeGroups: []string{},
		},
		{
			name:                  "some similar node groups, but no schedulable pods",
			nodeGroup:             "ng1",
			similarNodeGroups:     []string{"ng2", "ng3"},
			otherNodeGroups:       []string{"pg1", "pg2"},
			balancingEnabled:      true,
			wantSimilarNodeGroups: []string{},
		},
		{
			name:              "some similar node groups and same schedulable pods, but balancing disabled",
			nodeGroup:         "ng1",
			similarNodeGroups: []string{"ng2", "ng3"},
			otherNodeGroups:   []string{"pg1", "pg2"},
			balancingEnabled:  false,
			schedulablePodGroups: map[string][]estimator.PodEquivalenceGroup{
				"ng1": {podGroup1},
				"ng2": {podGroup1},
				"ng3": {podGroup1},
				"pg1": {podGroup1},
				"pg2": {podGroup1},
			},
			wantSimilarNodeGroups: []string{},
		},
		{
			name:              "some similar node groups and same schedulable pods",
			nodeGroup:         "ng1",
			similarNodeGroups: []string{"ng2", "ng3"},
			otherNodeGroups:   []string{"pg1", "pg2"},
			balancingEnabled:  true,
			schedulablePodGroups: map[string][]estimator.PodEquivalenceGroup{
				"ng1": {podGroup1},
				"ng2": {podGroup1},
				"ng3": {podGroup1},
				"pg1": {podGroup1},
				"pg2": {podGroup1},
			},
			wantSimilarNodeGroups: []string{"ng2", "ng3"},
		},
		{
			name:              "similar node groups can schedule more pods",
			nodeGroup:         "ng1",
			similarNodeGroups: []string{"ng2", "ng3"},
			otherNodeGroups:   []string{"pg1", "pg2"},
			balancingEnabled:  true,
			schedulablePodGroups: map[string][]estimator.PodEquivalenceGroup{
				"ng1": {podGroup1},
				"ng2": {podGroup1, podGroup2},
				"ng3": {podGroup1, podGroup2, podGroup3},
				"pg1": {podGroup1, podGroup2},
				"pg2": {podGroup1, podGroup2, podGroup3},
			},
			wantSimilarNodeGroups: []string{"ng2", "ng3"},
		},
		{
			name:              "similar node groups can schedule different/no pods",
			nodeGroup:         "ng1",
			similarNodeGroups: []string{"ng2", "ng3"},
			otherNodeGroups:   []string{"pg1", "pg2"},
			balancingEnabled:  true,
			schedulablePodGroups: map[string][]estimator.PodEquivalenceGroup{
				"ng1": {podGroup1, podGroup2},
				"ng2": {podGroup1},
				"pg1": {podGroup1},
			},
			wantSimilarNodeGroups: []string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			provider := testprovider.NewTestCloudProvider(func(string, int) error { return nil }, nil)
			nodeGroupSetProcessor := &constNodeGroupSetProcessor{}
			now := time.Now()

			allNodeGroups := []string{tc.nodeGroup}
			allNodeGroups = append(allNodeGroups, tc.similarNodeGroups...)
			allNodeGroups = append(allNodeGroups, tc.otherNodeGroups...)

			var nodes []*apiv1.Node
			for _, ng := range allNodeGroups {
				nodeName := fmt.Sprintf("%s-node", ng)
				node := BuildTestNode(nodeName, 100, 1000)
				SetNodeReadyState(node, true, now.Add(-2*time.Minute))
				nodes = append(nodes, node)

				provider.AddNodeGroup(ng, 0, 10, 1)
				provider.AddNode(ng, node)
			}

			for _, ng := range tc.similarNodeGroups {
				nodeGroupSetProcessor.similarNodeGroups = append(nodeGroupSetProcessor.similarNodeGroups, provider.GetNodeGroup(ng))
			}

			listers := kube_util.NewListerRegistry(nil, nil, kube_util.NewTestPodLister(nil), nil, nil, nil, nil, nil, nil)
			ctx, err := NewScaleTestAutoscalingContext(config.AutoscalingOptions{BalanceSimilarNodeGroups: tc.balancingEnabled}, &fake.Clientset{}, listers, provider, nil, nil)
			assert.NoError(t, err)

			nodeInfos, _ := nodeinfosprovider.NewDefaultTemplateNodeInfoProvider(nil, false).Process(&ctx, nodes, []*appsv1.DaemonSet{}, taints.TaintConfig{}, now)
			clusterState := clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{}, ctx.LogRecorder, NewBackoff(), nodegroupconfig.NewDefaultNodeGroupConfigProcessor(config.NodeGroupAutoscalingOptions{MaxNodeProvisionTime: 15 * time.Minute}), asyncnodegroups.NewDefaultAsyncNodeGroupStateChecker())
			assert.NoError(t, clusterState.UpdateNodes(nodes, nodeInfos, time.Now()))

			suOrchestrator := &ScaleUpOrchestrator{}
			suOrchestrator.Initialize(&ctx, &processors.AutoscalingProcessors{NodeGroupSetProcessor: nodeGroupSetProcessor}, clusterState, newEstimatorBuilder(), taints.TaintConfig{})
			similarNodeGroups := suOrchestrator.ComputeSimilarNodeGroups(provider.GetNodeGroup(tc.nodeGroup), nodeInfos, tc.schedulablePodGroups, now)

			var gotSimilarNodeGroups []string
			for _, ng := range similarNodeGroups {
				gotSimilarNodeGroups = append(gotSimilarNodeGroups, ng.Id())
			}
			assert.ElementsMatch(t, gotSimilarNodeGroups, tc.wantSimilarNodeGroups)
		})
	}
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

	now := time.Now()

	for gid, gconf := range testCfg {
		provider.AddNodeGroup(gid, gconf.min, gconf.max, gconf.size)
		for i := 0; i < gconf.size; i++ {
			nodeName := fmt.Sprintf("%v-node-%v", gid, i)
			node := BuildTestNode(nodeName, 100, 1000)
			SetNodeReadyState(node, true, now.Add(-2*time.Minute))
			nodes = append(nodes, node)

			pod := BuildTestPod(fmt.Sprintf("%v-pod-%v", gid, i), 80, 0)
			pod.Spec.NodeName = nodeName
			podList = append(podList, pod)

			provider.AddNode(gid, node)
		}
	}

	podLister := kube_util.NewTestPodLister(podList)
	listers := kube_util.NewListerRegistry(nil, nil, podLister, nil, nil, nil, nil, nil, nil)

	options := config.AutoscalingOptions{
		EstimatorName:            estimator.BinpackingEstimatorName,
		BalanceSimilarNodeGroups: true,
		MaxCoresTotal:            config.DefaultMaxClusterCores,
		MaxMemoryTotal:           config.DefaultMaxClusterMemory,
	}
	context, err := NewScaleTestAutoscalingContext(options, &fake.Clientset{}, listers, provider, nil, nil)
	assert.NoError(t, err)

	nodeInfos, _ := nodeinfosprovider.NewDefaultTemplateNodeInfoProvider(nil, false).Process(&context, nodes, []*appsv1.DaemonSet{}, taints.TaintConfig{}, now)
	clusterState := clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{}, context.LogRecorder, NewBackoff(), nodegroupconfig.NewDefaultNodeGroupConfigProcessor(config.NodeGroupAutoscalingOptions{MaxNodeProvisionTime: 15 * time.Minute}), asyncnodegroups.NewDefaultAsyncNodeGroupStateChecker())
	clusterState.UpdateNodes(nodes, nodeInfos, time.Now())

	pods := make([]*apiv1.Pod, 0)
	for i := 0; i < 2; i++ {
		pods = append(pods, BuildTestPod(fmt.Sprintf("test-pod-%v", i), 80, 0))
	}

	processors := NewTestProcessors(&context)
	suOrchestrator := New()
	suOrchestrator.Initialize(&context, processors, clusterState, newEstimatorBuilder(), taints.TaintConfig{})
	scaleUpStatus, typedErr := suOrchestrator.ScaleUp(pods, nodes, []*appsv1.DaemonSet{}, nodeInfos, false)

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
	listers := kube_util.NewListerRegistry(nil, nil, podLister, nil, nil, nil, nil, nil, nil)
	context, err := NewScaleTestAutoscalingContext(options, fakeClient, listers, provider, nil, nil)
	assert.NoError(t, err)

	clusterState := clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{}, context.LogRecorder, NewBackoff(), nodegroupconfig.NewDefaultNodeGroupConfigProcessor(config.NodeGroupAutoscalingOptions{MaxNodeProvisionTime: 15 * time.Minute}), asyncnodegroups.NewDefaultAsyncNodeGroupStateChecker())

	processors := NewTestProcessors(&context)
	processors.NodeGroupListProcessor = &MockAutoprovisioningNodeGroupListProcessor{T: t}
	processors.NodeGroupManager = &MockAutoprovisioningNodeGroupManager{T: t, ExtraGroups: 0}

	nodes := []*apiv1.Node{}
	nodeInfos, _ := nodeinfosprovider.NewDefaultTemplateNodeInfoProvider(nil, false).Process(&context, nodes, []*appsv1.DaemonSet{}, taints.TaintConfig{}, time.Now())

	suOrchestrator := New()
	suOrchestrator.Initialize(&context, processors, clusterState, newEstimatorBuilder(), taints.TaintConfig{})
	scaleUpStatus, err := suOrchestrator.ScaleUp([]*apiv1.Pod{p1}, nodes, []*appsv1.DaemonSet{}, nodeInfos, false)
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
	listers := kube_util.NewListerRegistry(nil, nil, podLister, nil, nil, nil, nil, nil, nil)
	context, err := NewScaleTestAutoscalingContext(options, fakeClient, listers, provider, nil, nil)
	assert.NoError(t, err)

	clusterState := clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{}, context.LogRecorder, NewBackoff(), nodegroupconfig.NewDefaultNodeGroupConfigProcessor(config.NodeGroupAutoscalingOptions{MaxNodeProvisionTime: 15 * time.Minute}), asyncnodegroups.NewDefaultAsyncNodeGroupStateChecker())

	processors := NewTestProcessors(&context)
	processors.NodeGroupListProcessor = &MockAutoprovisioningNodeGroupListProcessor{T: t}
	processors.NodeGroupManager = &MockAutoprovisioningNodeGroupManager{T: t, ExtraGroups: 2}

	nodes := []*apiv1.Node{}
	nodeInfos, _ := nodeinfosprovider.NewDefaultTemplateNodeInfoProvider(nil, false).Process(&context, nodes, []*appsv1.DaemonSet{}, taints.TaintConfig{}, time.Now())

	suOrchestrator := New()
	suOrchestrator.Initialize(&context, processors, clusterState, newEstimatorBuilder(), taints.TaintConfig{})
	scaleUpStatus, err := suOrchestrator.ScaleUp([]*apiv1.Pod{p1, p2, p3}, nodes, []*appsv1.DaemonSet{}, nodeInfos, false)
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

func TestScaleUpToMeetNodeGroupMinSize(t *testing.T) {
	podLister := kube_util.NewTestPodLister([]*apiv1.Pod{})
	listers := kube_util.NewListerRegistry(nil, nil, podLister, nil, nil, nil, nil, nil, nil)
	provider := testprovider.NewTestCloudProvider(func(nodeGroup string, increase int) error {
		assert.Equal(t, "ng1", nodeGroup)
		assert.Equal(t, 1, increase)
		return nil
	}, nil)
	resourceLimiter := cloudprovider.NewResourceLimiter(
		map[string]int64{cloudprovider.ResourceNameCores: 0, cloudprovider.ResourceNameMemory: 0},
		map[string]int64{cloudprovider.ResourceNameCores: 48, cloudprovider.ResourceNameMemory: 1000},
	)
	provider.SetResourceLimiter(resourceLimiter)

	// Test cases:
	// ng1: current size 1, min size 3, cores limit 48, memory limit 1000 => scale up with 1 new node.
	// ng2: current size 1, min size 1, cores limit 48, memory limit 1000 => no scale up.
	n1 := BuildTestNode("n1", 16000, 32)
	SetNodeReadyState(n1, true, time.Now())
	n2 := BuildTestNode("n2", 16000, 32)
	SetNodeReadyState(n2, true, time.Now())
	provider.AddNodeGroup("ng1", 3, 10, 1)
	provider.AddNode("ng1", n1)
	provider.AddNodeGroup("ng2", 1, 10, 1)
	provider.AddNode("ng2", n2)

	options := config.AutoscalingOptions{
		EstimatorName:  estimator.BinpackingEstimatorName,
		MaxCoresTotal:  config.DefaultMaxClusterCores,
		MaxMemoryTotal: config.DefaultMaxClusterMemory,
	}
	context, err := NewScaleTestAutoscalingContext(options, &fake.Clientset{}, listers, provider, nil, nil)
	assert.NoError(t, err)

	nodes := []*apiv1.Node{n1, n2}
	nodeInfos, _ := nodeinfosprovider.NewDefaultTemplateNodeInfoProvider(nil, false).Process(&context, nodes, []*appsv1.DaemonSet{}, taints.TaintConfig{}, time.Now())
	processors := NewTestProcessors(&context)
	clusterState := clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{}, context.LogRecorder, NewBackoff(), nodegroupconfig.NewDefaultNodeGroupConfigProcessor(config.NodeGroupAutoscalingOptions{MaxNodeProvisionTime: 15 * time.Minute}), asyncnodegroups.NewDefaultAsyncNodeGroupStateChecker())
	clusterState.UpdateNodes(nodes, nodeInfos, time.Now())

	suOrchestrator := New()
	suOrchestrator.Initialize(&context, processors, clusterState, newEstimatorBuilder(), taints.TaintConfig{})
	scaleUpStatus, err := suOrchestrator.ScaleUpToNodeGroupMinSize(nodes, nodeInfos)
	assert.NoError(t, err)
	assert.True(t, scaleUpStatus.WasSuccessful())
	assert.Equal(t, 1, len(scaleUpStatus.ScaleUpInfos))
	assert.Equal(t, 2, scaleUpStatus.ScaleUpInfos[0].NewSize)
	assert.Equal(t, "ng1", scaleUpStatus.ScaleUpInfos[0].Group.Id())
}

func TestScaleupAsyncNodeGroupsEnabled(t *testing.T) {
	t1 := BuildTestNode("t1", 100, 0)
	SetNodeReadyState(t1, true, time.Time{})
	ti1 := schedulerframework.NewNodeInfo()
	ti1.SetNode(t1)

	t2 := BuildTestNode("t2", 0, 100)
	SetNodeReadyState(t2, true, time.Time{})
	ti2 := schedulerframework.NewNodeInfo()
	ti2.SetNode(t2)

	testCases := []struct {
		upcomingNodeGroupsNames []string
		podsToAdd               []*v1.Pod
		isUpcomingMockMap       map[string]bool
		machineTypes            []string
		machineTemplates        map[string]*schedulerframework.NodeInfo
		expectedCreatedGroups   map[string]bool
		expectedExpandedGroups  map[string]int
	}{
		{
			upcomingNodeGroupsNames: []string{"T1"},
			podsToAdd:               []*v1.Pod{BuildTestPod("p1", 80, 0), BuildTestPod("p2", 80, 0)},
			isUpcomingMockMap:       map[string]bool{"autoprovisioned-T1": true},
			machineTypes:            []string{"T1"},
			machineTemplates:        map[string]*schedulerframework.NodeInfo{"T1": ti1},
			expectedCreatedGroups:   map[string]bool{},
			expectedExpandedGroups:  map[string]int{"autoprovisioned-T1": 2},
		},
		{
			upcomingNodeGroupsNames: []string{},
			podsToAdd:               []*v1.Pod{BuildTestPod("p1", 80, 0)},
			isUpcomingMockMap:       map[string]bool{},
			machineTypes:            []string{"T1"},
			machineTemplates:        map[string]*schedulerframework.NodeInfo{"T1": ti1},
			expectedCreatedGroups:   map[string]bool{"autoprovisioned-T1": true},
			expectedExpandedGroups:  map[string]int{"autoprovisioned-T1": 1},
		},
		{
			upcomingNodeGroupsNames: []string{"T1"},
			podsToAdd:               []*v1.Pod{BuildTestPod("p3", 0, 100), BuildTestPod("p2", 0, 100)},
			isUpcomingMockMap:       map[string]bool{"autoprovisioned-T1": true},
			machineTypes:            []string{"T1", "T2"},
			machineTemplates:        map[string]*schedulerframework.NodeInfo{"T1": ti1, "T2": ti2},
			expectedCreatedGroups:   map[string]bool{"autoprovisioned-T2": true},
			expectedExpandedGroups:  map[string]int{"autoprovisioned-T2": 2},
		},
	}

	for _, tc := range testCases {
		createdGroups := make(map[string]bool)
		expandedGroups := make(map[string]int)
		fakeClient := &fake.Clientset{}

		provider := testprovider.NewTestAutoprovisioningCloudProvider(
			func(nodeGroup string, increase int) error {
				expandedGroups[nodeGroup] += increase
				return nil
			}, nil, func(nodeGroup string) error {
				createdGroups[nodeGroup] = true
				return nil
			}, nil, tc.machineTypes, tc.machineTemplates)

		for _, upcomingNodeName := range tc.upcomingNodeGroupsNames {
			provider.AddNodeGroup(upcomingNodeName, 0, 10, 0)
		}

		options := config.AutoscalingOptions{
			NodeAutoprovisioningEnabled: true,
			AsyncNodeGroupsEnabled:      true,
		}
		podLister := kube_util.NewTestPodLister([]*apiv1.Pod{})
		listers := kube_util.NewListerRegistry(nil, nil, podLister, nil, nil, nil, nil, nil, nil)
		context, err := NewScaleTestAutoscalingContext(options, fakeClient, listers, provider, nil, nil)
		assert.NoError(t, err)

		clusterState := clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{}, context.LogRecorder, NewBackoff(), nodegroupconfig.NewDefaultNodeGroupConfigProcessor(config.NodeGroupAutoscalingOptions{MaxNodeProvisionTime: 15 * time.Minute}), asyncnodegroups.NewDefaultAsyncNodeGroupStateChecker())

		processors := NewTestProcessors(&context)
		processors.NodeGroupListProcessor = &MockAutoprovisioningNodeGroupListProcessor{T: t}
		processors.NodeGroupManager = &MockAutoprovisioningNodeGroupManager{T: t, ExtraGroups: 0}
		processors.AsyncNodeGroupStateChecker = &asyncnodegroups.MockAsyncNodeGroupStateChecker{IsUpcomingNodeGroup: tc.isUpcomingMockMap}

		nodes := []*apiv1.Node{}
		nodeInfos, _ := nodeinfosprovider.NewDefaultTemplateNodeInfoProvider(nil, false).Process(&context, nodes, []*appsv1.DaemonSet{}, taints.TaintConfig{}, time.Now())

		suOrchestrator := New()
		suOrchestrator.Initialize(&context, processors, clusterState, newEstimatorBuilder(), taints.TaintConfig{})
		scaleUpStatus, err := suOrchestrator.ScaleUp(tc.podsToAdd, nodes, []*appsv1.DaemonSet{}, nodeInfos, false)
		assert.NoError(t, err)
		assert.True(t, scaleUpStatus.WasSuccessful())

		assert.Equal(t, len(tc.expectedCreatedGroups), len(createdGroups))
		assert.Equal(t, len(tc.expectedExpandedGroups), len(expandedGroups))

		for groupName := range tc.expectedCreatedGroups {
			assert.True(t, createdGroups[groupName])
		}
		for groupName, expectedExpandedGroupValue := range tc.expectedExpandedGroups {
			assert.Equal(t, expectedExpandedGroupValue, expandedGroups[groupName])
		}
	}

}

func TestCheckDeltaWithinLimits(t *testing.T) {
	type testcase struct {
		limits            resource.Limits
		delta             resource.Delta
		exceededResources []string
	}
	tests := []testcase{
		{
			limits:            resource.Limits{"a": 10},
			delta:             resource.Delta{"a": 10},
			exceededResources: []string{},
		},
		{
			limits:            resource.Limits{"a": 10},
			delta:             resource.Delta{"a": 11},
			exceededResources: []string{"a"},
		},
		{
			limits:            resource.Limits{"a": 10},
			delta:             resource.Delta{"b": 10},
			exceededResources: []string{},
		},
		{
			limits:            resource.Limits{"a": resource.LimitUnknown},
			delta:             resource.Delta{"a": 0},
			exceededResources: []string{},
		},
		{
			limits:            resource.Limits{"a": resource.LimitUnknown},
			delta:             resource.Delta{"a": 1},
			exceededResources: []string{"a"},
		},
		{
			limits:            resource.Limits{"a": 10, "b": 20, "c": 30},
			delta:             resource.Delta{"a": 11, "b": 20, "c": 31},
			exceededResources: []string{"a", "c"},
		},
	}

	for _, test := range tests {
		checkResult := resource.CheckDeltaWithinLimits(test.limits, test.delta)
		if len(test.exceededResources) == 0 {
			assert.Equal(t, resource.LimitsNotExceeded(), checkResult)
		} else {
			assert.Equal(t, resource.LimitsCheckResult{Exceeded: true, ExceededResources: test.exceededResources}, checkResult)
		}
	}
}

func TestAuthErrorHandling(t *testing.T) {
	metrics.RegisterAll(false)
	config := &ScaleUpTestConfig{
		Groups: []NodeGroupConfig{
			{Name: "ng1", MaxSize: 2},
		},
		Nodes: []NodeConfig{
			{Name: "ng1-n1", Cpu: 1500, Memory: 1000 * utils.MiB, Ready: true, Group: "ng1"},
		},
		ExtraPods: []PodConfig{
			{Name: "p1", Cpu: 1000},
		},
		OnScaleUp: func(group string, i int) error {
			return errors.NewAutoscalerError(errors.AutoscalerErrorType("authError"), "auth error")
		},
		Options: &defaultOptions,
	}
	results := runSimpleScaleUpTest(t, config)
	expected := errors.NewAutoscalerError(
		errors.AutoscalerErrorType("authError"),
		"failed to increase node group size: auth error",
	)
	assert.Equal(t, expected, results.ScaleUpError)
	assertLegacyRegistryEntry(t, "cluster_autoscaler_failed_scale_ups_total{reason=\"authError\"} 1")
}

func assertLegacyRegistryEntry(t *testing.T, entry string) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(legacyregistry.Handler().ServeHTTP)
	handler.ServeHTTP(rr, req)
	// Check that the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
	// Check that the failed scale up reason is set correctly.
	assert.Contains(t, rr.Body.String(), entry)
}

func simplifyScaleUpStatus(scaleUpStatus *status.ScaleUpStatus) ScaleUpStatusInfo {
	remainUnschedulable := []string{}
	for _, nsi := range scaleUpStatus.PodsRemainUnschedulable {
		remainUnschedulable = append(remainUnschedulable, nsi.Pod.Name)
	}
	return ScaleUpStatusInfo{
		Result:                  scaleUpStatus.Result,
		PodsTriggeredScaleUp:    ExtractPodNames(scaleUpStatus.PodsTriggeredScaleUp),
		PodsRemainUnschedulable: remainUnschedulable,
		PodsAwaitEvaluation:     ExtractPodNames(scaleUpStatus.PodsAwaitEvaluation),
	}
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
