/*
Copyright 2022 The Kubernetes Authors.

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

package unneeded

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	testprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	ca_context "k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/resource"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/status"
	. "k8s.io/autoscaler/cluster-autoscaler/core/test"
	nodeprocessors "k8s.io/autoscaler/cluster-autoscaler/processors/nodes"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
	"k8s.io/client-go/kubernetes/fake"
)

const (
	testScaleDownTimeout    = 5 * time.Minute
	defaultNodeCpu          = 10
	defaultNodeMem          = 100
	defaultNodeGroupMaxSize = 100
)

func TestUpdate(t *testing.T) {
	initialTimestamp := time.Now()
	finalTimestamp := initialTimestamp.Add(1 * time.Minute)
	testCases := []struct {
		desc           string
		initialNodes   []simulator.NodeToBeRemoved
		finalNodes     []simulator.NodeToBeRemoved
		wantTimestamps map[string]time.Time
		wantVersions   map[string]string
	}{
		{
			desc: "added then deleted",
			initialNodes: []simulator.NodeToBeRemoved{
				makeNode("n1", "v1"),
				makeNode("n2", "v1"),
				makeNode("n3", "v1"),
			},
			finalNodes: []simulator.NodeToBeRemoved{},
		},
		{
			desc:         "added in last call",
			initialNodes: []simulator.NodeToBeRemoved{},
			finalNodes: []simulator.NodeToBeRemoved{
				makeNode("n1", "v1"),
				makeNode("n2", "v1"),
				makeNode("n3", "v1"),
			},
			wantTimestamps: map[string]time.Time{"n1": finalTimestamp, "n2": finalTimestamp, "n3": finalTimestamp},
			wantVersions:   map[string]string{"n1": "v1", "n2": "v1", "n3": "v1"},
		},
		{
			desc: "single one remaining",
			initialNodes: []simulator.NodeToBeRemoved{
				makeNode("n1", "v1"),
				makeNode("n2", "v1"),
				makeNode("n3", "v1"),
			},
			finalNodes: []simulator.NodeToBeRemoved{
				makeNode("n2", "v2"),
			},
			wantTimestamps: map[string]time.Time{"n2": initialTimestamp},
			wantVersions:   map[string]string{"n2": "v2"},
		},
		{
			desc: "single one older",
			initialNodes: []simulator.NodeToBeRemoved{
				makeNode("n2", "v1"),
			},
			finalNodes: []simulator.NodeToBeRemoved{
				makeNode("n1", "v2"),
				makeNode("n2", "v2"),
				makeNode("n3", "v2"),
			},
			wantTimestamps: map[string]time.Time{"n1": finalTimestamp, "n2": initialTimestamp, "n3": finalTimestamp},
			wantVersions:   map[string]string{"n1": "v2", "n2": "v2", "n3": "v2"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()
			nodes := NewNodes(nil, nil)

			provider := testprovider.NewTestCloudProviderBuilder().Build()
			ctx := &ca_context.AutoscalingContext{CloudProvider: provider}

			nodes.Update(ctx, tc.initialNodes, initialTimestamp)
			nodes.Update(ctx, tc.finalNodes, finalTimestamp)

			wantNodes := len(tc.wantTimestamps)
			assert.Equal(t, wantNodes, len(nodes.AsList()))
			assert.Equal(t, wantNodes, len(nodes.byName))
			for _, n := range nodes.AsList() {
				nn, found := nodes.byName[n.Node.Name]
				assert.True(t, found)
				assert.Equal(t, tc.wantTimestamps[n.Node.Name], nn.since)
				assert.Equal(t, tc.wantVersions[n.Node.Name], version(nn.ntbr))
			}
		})
	}
}

const testVersion = "testVersion"

func makeNode(name, version string) simulator.NodeToBeRemoved {
	n := BuildTestNode(name, 1000, 10)
	n.Annotations = map[string]string{testVersion: version}
	return simulator.NodeToBeRemoved{Node: n}
}

func version(n simulator.NodeToBeRemoved) string {
	return n.Node.Annotations[testVersion]
}

func TestRemovableAt(t *testing.T) {
	testCases := []struct {
		name                string
		numEmpty            int
		numDrain            int
		minSize             int
		targetSize          int
		numOngoingDeletions int
		numEmptyToRemove    int
		numDrainToRemove    int
	}{
		{
			name:                "Node group min size is not reached",
			numEmpty:            3,
			numDrain:            2,
			minSize:             1,
			targetSize:          10,
			numOngoingDeletions: 2,
			numEmptyToRemove:    3,
			numDrainToRemove:    2,
		},
		{
			name:                "Node group min size is reached for drain nodes",
			numEmpty:            3,
			numDrain:            5,
			minSize:             1,
			targetSize:          10,
			numOngoingDeletions: 2,
			numEmptyToRemove:    3,
			numDrainToRemove:    4,
		},
		{
			name:                "Node group min size is reached for empty and drain nodes",
			numEmpty:            3,
			numDrain:            5,
			minSize:             1,
			targetSize:          5,
			numOngoingDeletions: 2,
			numEmptyToRemove:    2,
			numDrainToRemove:    0,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ng := testprovider.NewTestNodeGroup("ng", 100, tc.minSize, tc.targetSize, true, false, "", nil, nil)
			empty := []simulator.NodeToBeRemoved{}
			for i := 0; i < tc.numEmpty; i++ {
				empty = append(empty, simulator.NodeToBeRemoved{
					Node: BuildTestNode(fmt.Sprintf("empty-%d", i), 10, 100),
				})
			}
			drain := []simulator.NodeToBeRemoved{}
			for i := 0; i < tc.numDrain; i++ {
				drain = append(drain, simulator.NodeToBeRemoved{
					Node:             BuildTestNode(fmt.Sprintf("drain-%d", i), 10, 100),
					PodsToReschedule: []*apiv1.Pod{BuildTestPod(fmt.Sprintf("pod-%d", i), 1, 1)},
				})
			}

			removableNodes := append(empty, drain...)
			provider := testprovider.NewTestCloudProviderBuilder().Build()
			provider.InsertNodeGroup(ng)
			for _, node := range removableNodes {
				provider.AddNode("ng", node.Node)
			}

			as := &fakeActuationStatus{deletionCount: map[string]int{"ng": tc.numOngoingDeletions}}

			rsLister, err := kube_util.NewTestReplicaSetLister(nil)
			assert.NoError(t, err)
			registry := kube_util.NewListerRegistry(nil, nil, nil, nil, nil, nil, nil, rsLister, nil)
			autoscalingCtx, err := NewScaleTestAutoscalingContext(config.AutoscalingOptions{ScaleDownSimulationTimeout: 5 * time.Minute}, &fake.Clientset{}, registry, provider, nil, nil)
			assert.NoError(t, err)
			expectedThreshold := 5 * time.Minute
			fakeTimeGetter := &fakeScaleDownTimeGetter{
				unneededTime: 0,
				unreadyTime:  expectedThreshold,
			}
			n := NewNodes(fakeTimeGetter, &resource.LimitsFinder{})

			n.Update(&autoscalingCtx, removableNodes, time.Now().Add(-10*time.Minute)) //add -10 min to work correctly with unneeded time threshold

			gotEmptyToRemove, gotDrainToRemove, _ := n.RemovableAt(&autoscalingCtx, nodeprocessors.ScaleDownContext{
				ActuationStatus:     as,
				ResourcesLeft:       resource.Limits{},
				ResourcesWithLimits: []string{},
			}, time.Now())
			if len(gotDrainToRemove) != tc.numDrainToRemove || len(gotEmptyToRemove) != tc.numEmptyToRemove {
				t.Errorf("%s: getNodesToRemove() return %d, %d, want %d, %d", tc.name, len(gotEmptyToRemove), len(gotDrainToRemove), tc.numEmptyToRemove, tc.numDrainToRemove)
			}

			candidates := n.AsList()
			candidateMap := make(map[string]time.Duration)
			for _, c := range candidates {
				candidateMap[c.Node.Name] = c.RemovalThreshold
			}

			for _, node := range gotEmptyToRemove {
				nodeName := node.Node.Name
				got, ok := candidateMap[nodeName]
				if !ok {
					t.Errorf("Node %s not found in AsList", nodeName)
				} else if got != expectedThreshold {
					t.Errorf("Node %s has threshold %v, want %v", nodeName, got, expectedThreshold)
				}
			}
		})
	}
}

type nodeStateConfig struct {
	isReady                        bool
	sinceOffset                    time.Duration
	hasScaleDownDisabledAnnotation bool
}

type thresholdConfig struct {
	unneeded time.Duration
	unready  time.Duration
}

type nodeGroupConfigTest struct {
	minSize      int
	targetSize   int
	isConfigured bool
}

type unremovableTestCase struct {
	name                  string
	nodeConfig            nodeStateConfig
	thresholds            thresholdConfig
	groupConfig           nodeGroupConfigTest
	shouldTimeGetterError bool
	expectedReason        simulator.UnremovableReason
}

func TestRemovableAt_UnremovableReasons(t *testing.T) {
	now := time.Now()

	baseCase := unremovableTestCase{
		nodeConfig: nodeStateConfig{
			isReady:                        true,
			sinceOffset:                    -10 * time.Minute,
			hasScaleDownDisabledAnnotation: false,
		},
		thresholds: thresholdConfig{
			unneeded: 5 * time.Minute,
			unready:  5 * time.Minute,
		},
		groupConfig: nodeGroupConfigTest{
			minSize:      0,
			targetSize:   1,
			isConfigured: true,
		},
	}

	testCases := []unremovableTestCase{
		{
			name:                  "ThresholdRetrievalFails",
			nodeConfig:            baseCase.nodeConfig,
			thresholds:            baseCase.thresholds,
			groupConfig:           baseCase.groupConfig,
			shouldTimeGetterError: true,
			expectedReason:        simulator.UnexpectedError,
		},
		{
			name: "ScaleDownDisabledAnnotation",
			nodeConfig: nodeStateConfig{
				isReady:                        baseCase.nodeConfig.isReady,
				sinceOffset:                    baseCase.nodeConfig.sinceOffset,
				hasScaleDownDisabledAnnotation: true,
			},
			thresholds:     baseCase.thresholds,
			groupConfig:    baseCase.groupConfig,
			expectedReason: simulator.ScaleDownDisabledAnnotation,
		},
		{
			name:       "NotUnneededLongEnough",
			nodeConfig: baseCase.nodeConfig,
			thresholds: thresholdConfig{
				unneeded: 15 * time.Minute,
				unready:  baseCase.thresholds.unready,
			},
			groupConfig:    baseCase.groupConfig,
			expectedReason: simulator.NotUnneededLongEnough,
		},
		{
			name: "NotUnreadyLongEnough",
			nodeConfig: nodeStateConfig{
				isReady:                        false,
				sinceOffset:                    baseCase.nodeConfig.sinceOffset,
				hasScaleDownDisabledAnnotation: baseCase.nodeConfig.hasScaleDownDisabledAnnotation,
			},
			thresholds: thresholdConfig{
				unready:  15 * time.Minute,
				unneeded: baseCase.thresholds.unneeded,
			},
			groupConfig:    baseCase.groupConfig,
			expectedReason: simulator.NotUnreadyLongEnough,
		},
		{
			name:       "NotAutoscaled",
			nodeConfig: baseCase.nodeConfig,
			thresholds: baseCase.thresholds,
			groupConfig: nodeGroupConfigTest{
				minSize:      baseCase.groupConfig.minSize,
				isConfigured: false,
			},
			expectedReason: simulator.NotAutoscaled,
		},
		{
			name:       "NodeGroupMinSizeReached",
			nodeConfig: baseCase.nodeConfig,
			thresholds: baseCase.thresholds,
			groupConfig: nodeGroupConfigTest{
				minSize:      1,
				targetSize:   1,
				isConfigured: baseCase.groupConfig.isConfigured,
			},
			expectedReason: simulator.NodeGroupMinSizeReached,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ng := testprovider.NewTestNodeGroup("ng", defaultNodeGroupMaxSize, tc.groupConfig.minSize, tc.groupConfig.targetSize, true, false, "", nil, nil)
			nodeName := fmt.Sprintf("test-node-%s", tc.name)
			node := buildTestNodeWithConfig(nodeName, tc.nodeConfig, now)
			nodesToProcess := []simulator.NodeToBeRemoved{{Node: node}}

			provider := testprovider.NewTestCloudProviderBuilder().Build()
			if tc.groupConfig.isConfigured {
				provider.InsertNodeGroup(ng)
				provider.AddNode("ng", node)
			}

			autoscalingCtx := ca_context.AutoscalingContext{
				CloudProvider: provider,
				AutoscalingOptions: config.AutoscalingOptions{
					ScaleDownSimulationTimeout: testScaleDownTimeout,
				},
			}

			var timeGetter scaleDownTimeGetter
			if tc.shouldTimeGetterError {
				timeGetter = &fakeScaleDownTimeGetter{returnError: true}
			} else {
				timeGetter = &fakeScaleDownTimeGetter{
					unneededTime: tc.thresholds.unneeded,
					unreadyTime:  tc.thresholds.unready,
				}
			}
			n := NewNodes(timeGetter, &resource.LimitsFinder{})

			n.Update(&autoscalingCtx, nodesToProcess, now.Add(tc.nodeConfig.sinceOffset))

			sdCtx := nodeprocessors.ScaleDownContext{
				ActuationStatus:     &fakeActuationStatus{deletionCount: map[string]int{}},
				ResourcesLeft:       nil,
				ResourcesWithLimits: []string{},
			}

			gotEmptyToRemove, gotDrainToRemove, gotUnremovable := n.RemovableAt(&autoscalingCtx, sdCtx, now)

			assert.Empty(t, gotEmptyToRemove, "Expected no empty nodes to be removable")
			assert.Empty(t, gotDrainToRemove, "Expected no drain nodes to be removable")

			assert.Len(t, gotUnremovable, 1, "Expected number of unremovable nodes mismatch")
			if len(gotUnremovable) > 0 {
				assert.Equal(t, nodeName, gotUnremovable[0].Node.Name, "Unremovable node name mismatch")
				assert.Equal(t, tc.expectedReason, gotUnremovable[0].Reason, "UnremovableReason mismatch for test case")
			}
		})
	}
}

func buildTestNodeWithConfig(name string, config nodeStateConfig, now time.Time) *apiv1.Node {
	node := BuildTestNode(name, defaultNodeCpu, defaultNodeMem)
	SetNodeReadyState(node, config.isReady, now.Add(config.sinceOffset))

	if config.hasScaleDownDisabledAnnotation {
		if node.Annotations == nil {
			node.Annotations = make(map[string]string)
		}
		node.Annotations["cluster-autoscaler.kubernetes.io/scale-down-disabled"] = "true"
	}
	return node
}

func TestNodeLoadFromExistingTaints(t *testing.T) {
	deletionCandidateTaint := taints.DeletionCandidateTaint()
	currentTime := time.Now()

	n1 := BuildTestNode("n1", 1000, 1000)
	SetNodeReadyState(n1, true, currentTime)
	nt1 := deletionCandidateTaint
	ntt1 := currentTime.Add(-time.Minute * 2)
	nt1.Value = fmt.Sprint(ntt1.Unix())
	n1.Spec.Taints = append(n1.Spec.Taints, nt1)

	n2 := BuildTestNode("n2", 1000, 1000)
	SetNodeReadyState(n2, true, currentTime)

	n3 := BuildTestNode("n3", 1000, 1000)
	SetNodeReadyState(n3, true, currentTime)
	nt3 := deletionCandidateTaint
	ntt3 := currentTime.Add(-time.Minute * 20)
	nt3.Value = fmt.Sprint(ntt3.Unix())
	n3.Spec.Taints = append(n3.Spec.Taints, nt3)

	n4 := BuildTestNode("n4", 1000, 1000)
	SetNodeReadyState(n4, true, currentTime)
	nt4 := deletionCandidateTaint
	nt4.Value = "INVALID_VALUE"
	n4.Spec.Taints = append(n4.Spec.Taints, nt4)

	testCases := []struct {
		name                     string
		allNodes                 []*apiv1.Node
		expectedUnneededNodes    []*apiv1.Node
		nodeDeletionCandidateTTL time.Duration
	}{
		{
			name:                     "All deletion candidate nodes with standard TTL",
			allNodes:                 []*apiv1.Node{n1, n2},
			expectedUnneededNodes:    []*apiv1.Node{n1},
			nodeDeletionCandidateTTL: time.Minute * 5,
		},
		{
			name:                     "Nodes with expired deletion candidate taint",
			allNodes:                 []*apiv1.Node{n1, n2, n3},
			expectedUnneededNodes:    []*apiv1.Node{n1},
			nodeDeletionCandidateTTL: time.Minute * 5,
		},
		{
			name:                     "Nodes with invalid deletion candidate taint",
			allNodes:                 []*apiv1.Node{n1, n2, n3, n4},
			expectedUnneededNodes:    []*apiv1.Node{n1},
			nodeDeletionCandidateTTL: time.Minute * 5,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			nodes := NewNodes(nil, nil)

			allNodeLister := kube_util.NewTestNodeLister(nil)
			allNodeLister.SetNodes(tc.allNodes)

			readyNodeLister := kube_util.NewTestNodeLister(nil)
			readyNodeLister.SetNodes(tc.allNodes)

			provider := testprovider.NewTestCloudProviderBuilder().Build()
			ctx := &ca_context.AutoscalingContext{CloudProvider: provider, AutoscalingOptions: config.AutoscalingOptions{NodeDeletionCandidateTTL: tc.nodeDeletionCandidateTTL}}
			ctx.ListerRegistry = kube_util.NewListerRegistry(allNodeLister, readyNodeLister,
				nil, nil, nil, nil, nil, nil, nil)
			nodes.LoadFromExistingTaints(ctx, currentTime)

			unneededNodes := nodes.AsList()

			assert.Equal(t, len(tc.expectedUnneededNodes), len(unneededNodes),
				"Expected %d unneeded nodes but got %d", len(tc.expectedUnneededNodes), len(unneededNodes))

			expectedNodeNames := make(map[string]bool)
			for _, node := range tc.expectedUnneededNodes {
				expectedNodeNames[node.Name] = true
			}
			for _, candidate := range unneededNodes {
				_, found := expectedNodeNames[candidate.Node.Name]
				assert.True(t, found, "Node %s was not expected to be unneeded", candidate.Node.Name)
			}
			for _, expectedNode := range tc.expectedUnneededNodes {
				assert.True(t, nodes.Contains(expectedNode.Name),
					"Expected node %s to be in unneeded nodes but wasn't found", expectedNode.Name)
			}
		})
	}

}

type fakeActuationStatus struct {
	recentEvictions []*apiv1.Pod
	deletionCount   map[string]int
}

func (f *fakeActuationStatus) RecentEvictions() []*apiv1.Pod {
	return f.recentEvictions
}

func (f *fakeActuationStatus) DeletionsInProgress() ([]string, []string) {
	return nil, nil
}

func (f *fakeActuationStatus) DeletionResults() (map[string]status.NodeDeleteResult, time.Time) {
	return nil, time.Time{}
}

func (f *fakeActuationStatus) DeletionsCount(nodeGroup string) int {
	return f.deletionCount[nodeGroup]
}

type fakeScaleDownTimeGetter struct {
	unneededTime time.Duration
	unreadyTime  time.Duration
	returnError  bool
}

func (f *fakeScaleDownTimeGetter) GetScaleDownUnneededTime(cloudprovider.NodeGroup) (time.Duration, error) {
	if f.returnError {
		return 0, fmt.Errorf("simulated error getting unneeded time")
	}
	return f.unneededTime, nil
}

func (f *fakeScaleDownTimeGetter) GetScaleDownUnreadyTime(cloudprovider.NodeGroup) (time.Duration, error) {
	if f.returnError {
		return 0, fmt.Errorf("simulated error getting unready time")
	}
	return f.unreadyTime, nil
}
