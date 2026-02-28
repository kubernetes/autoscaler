/*
Copyright 2024 The Kubernetes Authors.

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

package predicate

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	clientsetfake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/kubernetes/pkg/scheduler/apis/config"
	scheduler_config_latest "k8s.io/kubernetes/pkg/scheduler/apis/config/latest"

	testconfig "k8s.io/autoscaler/cluster-autoscaler/config/test"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot/store"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/utils/scheduler"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
)

func TestRunFiltersOnNode(t *testing.T) {
	p450 := BuildTestPod("p450", 450, 500000)
	p600 := BuildTestPod("p600", 600, 500000)
	p8000 := BuildTestPod("p8000", 8000, 0)
	p500 := BuildTestPod("p500", 500, 500000)
	podWithAffinity := BuildTestPod("pod_with_affinity", 500, 500, WithNodeNamesAffinity("n1000"))
	podWithInvalidAffinity := BuildTestPod("pod_with_affinity", 500, 500, WithNodeNamesAffinity("non-existing-node"))

	n1000 := BuildTestNode("n1000", 1000, 2000000)
	SetNodeReadyState(n1000, true, time.Time{})
	n1000Unschedulable := BuildTestNode("n1000", 1000, 2000000)
	SetNodeReadyState(n1000Unschedulable, true, time.Time{})

	// temp dir
	tmpDir, err := os.MkdirTemp("", "scheduler-configs")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	customConfigFile := filepath.Join(tmpDir, "custom_config.yaml")
	if err := os.WriteFile(customConfigFile,
		[]byte(testconfig.SchedulerConfigNodeResourcesFitDisabled),
		os.FileMode(0600)); err != nil {
		t.Fatal(err)
	}
	customConfig, err := scheduler.ConfigFromPath(customConfigFile)
	assert.NoError(t, err)

	tests := []struct {
		name                        string
		customConfig                *config.KubeSchedulerConfiguration
		node                        *apiv1.Node
		scheduledPods               []*apiv1.Pod
		testPod                     *apiv1.Pod
		expectError                 bool
		wantFailingPredicateName    string
		wantFailingPredicateReasons []string
		wantErrorSubstrings         []string
	}{
		// default predicate checker test cases
		{
			name:                        "default - other pod - insuficient cpu",
			node:                        n1000,
			scheduledPods:               []*apiv1.Pod{p450},
			testPod:                     p600,
			expectError:                 true,
			wantFailingPredicateName:    "NodeResourcesFit",
			wantFailingPredicateReasons: []string{"Insufficient cpu"},
			wantErrorSubstrings:         []string{"NodeResourcesFit", "Insufficient cpu"},
		},
		{
			name:          "default - other pod - ok",
			node:          n1000,
			scheduledPods: []*apiv1.Pod{p450},
			testPod:       p500,
			expectError:   false,
		},
		{
			name:                        "default - empty - insuficient cpu",
			node:                        n1000,
			scheduledPods:               []*apiv1.Pod{},
			testPod:                     p8000,
			expectError:                 true,
			wantFailingPredicateName:    "NodeResourcesFit",
			wantFailingPredicateReasons: []string{"Insufficient cpu"},
			wantErrorSubstrings:         []string{"NodeResourcesFit", "Insufficient cpu"},
		},
		{
			name:                        "default - empty - ok",
			node:                        n1000,
			scheduledPods:               []*apiv1.Pod{},
			testPod:                     p600,
			expectError:                 false,
			wantFailingPredicateName:    "NodeResourcesFit",
			wantFailingPredicateReasons: []string{"Insufficient cpu"},
		},
		{
			name:          "default - affinity on existing node - ok",
			node:          n1000,
			scheduledPods: []*apiv1.Pod{},
			testPod:       podWithAffinity,
			expectError:   false,
		},
		{
			name:                     "default - affinity on non-existing node - error",
			node:                     n1000,
			scheduledPods:            []*apiv1.Pod{},
			testPod:                  podWithInvalidAffinity,
			expectError:              true,
			customConfig:             customConfig,
			wantFailingPredicateName: "NodeAffinity",
			wantErrorSubstrings:      []string{"PreFilter filtered the Node out"},
		},
		// custom predicate checker test cases
		{
			name:          "custom - other pod - ok",
			node:          n1000,
			scheduledPods: []*apiv1.Pod{p450},
			testPod:       p600,
			expectError:   false,
			customConfig:  customConfig,
		},
		{
			name:          "custom -other pod - ok",
			node:          n1000,
			scheduledPods: []*apiv1.Pod{p450},
			testPod:       p500,
			expectError:   false,
			customConfig:  customConfig,
		},
		{
			name:          "custom -empty - ok",
			node:          n1000,
			scheduledPods: []*apiv1.Pod{},
			testPod:       p8000,
			expectError:   false,
			customConfig:  customConfig,
		},
		{
			name:          "custom -empty - ok",
			node:          n1000,
			scheduledPods: []*apiv1.Pod{},
			testPod:       p600,
			expectError:   false,
			customConfig:  customConfig,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pluginRunner, snapshot, err := newTestPluginRunnerAndSnapshot(tt.customConfig)
			assert.NoError(t, err)
			err = snapshot.AddNodeInfo(framework.NewTestNodeInfo(tt.node, tt.scheduledPods...))
			assert.NoError(t, err)

			node, state, predicateError := pluginRunner.RunFiltersOnNode(tt.testPod, tt.node.Name)
			if tt.expectError {
				assert.Nil(t, node)
				assert.Nil(t, state)
				assert.NotNil(t, predicateError)
				assert.Equal(t, clustersnapshot.FailingPredicateError, predicateError.Type())
				assert.Equal(t, tt.wantFailingPredicateName, predicateError.FailingPredicateName())
				assert.Equal(t, tt.wantFailingPredicateReasons, predicateError.FailingPredicateReasons())
				for _, wantErrorSubstring := range tt.wantErrorSubstrings {
					assert.Contains(t, predicateError.Error(), wantErrorSubstring)
				}
			} else {
				assert.Nil(t, predicateError)
				assert.NotNil(t, state)
				assert.Equal(t, tt.node, node)
			}
		})
	}
}

func TestRunFilterUntilPassingNode(t *testing.T) {
	p900 := BuildTestPod("p900", 900, 1000)
	p1900 := BuildTestPod("p1900", 1900, 1000)
	p2100 := BuildTestPod("p2100", 2100, 1000)

	n1000 := BuildTestNode("n1000", 1000, 2000000)
	n2000 := BuildTestNode("n2000", 2000, 2000000)

	// temp dir
	tmpDir, err := os.MkdirTemp("", "scheduler-configs")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	customConfigFile := filepath.Join(tmpDir, "custom_config.yaml")
	if err := os.WriteFile(customConfigFile,
		[]byte(testconfig.SchedulerConfigNodeResourcesFitDisabled),
		os.FileMode(0600)); err != nil {
		t.Fatal(err)
	}
	customConfig, err := scheduler.ConfigFromPath(customConfigFile)
	assert.NoError(t, err)

	testCases := []struct {
		name          string
		customConfig  *config.KubeSchedulerConfiguration
		pod           *apiv1.Pod
		expectedNodes []string
		expectError   bool
	}{
		// default predicate checker test cases
		{
			name:          "default - small pod - no error",
			pod:           p900,
			expectedNodes: []string{"n1000", "n2000"},
			expectError:   false,
		},
		{
			name:          "default - medium pod - no error",
			pod:           p1900,
			expectedNodes: []string{"n2000"},
			expectError:   false,
		},
		{
			name:        "default - large pod - insufficient cpu",
			pod:         p2100,
			expectError: true,
		},

		// custom predicate checker test cases
		{
			name:          "custom - small pod - no error",
			customConfig:  customConfig,
			pod:           p900,
			expectedNodes: []string{"n1000", "n2000"},
			expectError:   false,
		},
		{
			name:          "custom - medium pod - no error",
			customConfig:  customConfig,
			pod:           p1900,
			expectedNodes: []string{"n1000", "n2000"},
			expectError:   false,
		},
		{
			name:          "custom - large pod - insufficient cpu",
			customConfig:  customConfig,
			pod:           p2100,
			expectedNodes: []string{"n1000", "n2000"},
			expectError:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pluginRunner, snapshot, err := newTestPluginRunnerAndSnapshot(tc.customConfig)
			assert.NoError(t, err)

			err = snapshot.AddNodeInfo(framework.NewTestNodeInfo(n1000))
			assert.NoError(t, err)
			err = snapshot.AddNodeInfo(framework.NewTestNodeInfo(n2000))
			assert.NoError(t, err)

			node, state, err := pluginRunner.RunFiltersUntilPassingNode(tc.pod, func(info *framework.NodeInfo) bool { return true })
			if tc.expectError {
				assert.Nil(t, node)
				assert.Nil(t, state)
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, state)
				assert.Contains(t, tc.expectedNodes, node.Name)
			}
		})
	}
}

func TestDebugInfo(t *testing.T) {
	p1 := BuildTestPod("p1", 0, 0)
	node1 := BuildTestNode("n1", 1000, 2000000)
	node1.Spec.Taints = []apiv1.Taint{
		{
			Key:    "SomeTaint",
			Value:  "WhyNot?",
			Effect: apiv1.TaintEffectNoSchedule,
		},
		{
			Key:    "RandomTaint",
			Value:  "JustBecause",
			Effect: apiv1.TaintEffectNoExecute,
		},
	}
	SetNodeReadyState(node1, true, time.Time{})

	// with default predicate checker
	defaultPluginRunner, clusterSnapshot, err := newTestPluginRunnerAndSnapshot(nil)
	assert.NoError(t, err)

	err = clusterSnapshot.AddNodeInfo(framework.NewTestNodeInfo(node1))
	assert.NoError(t, err)

	_, _, predicateErr := defaultPluginRunner.RunFiltersOnNode(p1, "n1")
	assert.NotNil(t, predicateErr)
	assert.Contains(t, predicateErr.FailingPredicateReasons(), "node(s) had untolerated taint(s)")
	assert.Contains(t, predicateErr.Error(), "node(s) had untolerated taint(s)")
	assert.Contains(t, predicateErr.Error(), "RandomTaint")
	assert.Contains(t, predicateErr.Error(), "SomeTaint")

	// with custom predicate checker

	// temp dir
	tmpDir, err := os.MkdirTemp("", "scheduler-configs")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	customConfigFile := filepath.Join(tmpDir, "custom_config.yaml")
	if err := os.WriteFile(customConfigFile,
		[]byte(testconfig.SchedulerConfigTaintTolerationDisabled),
		os.FileMode(0600)); err != nil {
		t.Fatal(err)
	}

	customConfig, err := scheduler.ConfigFromPath(customConfigFile)
	assert.NoError(t, err)
	customPluginRunner, clusterSnapshot, err := newTestPluginRunnerAndSnapshot(customConfig)
	assert.NoError(t, err)

	err = clusterSnapshot.AddNodeInfo(framework.NewTestNodeInfo(node1))
	assert.NoError(t, err)

	_, _, predicateErr = customPluginRunner.RunFiltersOnNode(p1, "n1")
	assert.Nil(t, predicateErr)
}

func newTestPluginRunnerAndSnapshot(schedConfig *config.KubeSchedulerConfiguration) (*SchedulerPluginRunner, clustersnapshot.ClusterSnapshot, error) {
	if schedConfig == nil {
		defaultConfig, err := scheduler_config_latest.Default()
		if err != nil {
			return nil, nil, err
		}
		schedConfig = defaultConfig
	}

	fwHandle, err := framework.NewHandle(context.Background(), informers.NewSharedInformerFactory(clientsetfake.NewSimpleClientset(), 0), schedConfig, true, false)
	if err != nil {
		return nil, nil, err
	}
	snapshot := NewPredicateSnapshot(store.NewBasicSnapshotStore(), fwHandle, true, 1, false)
	return NewSchedulerPluginRunner(fwHandle, snapshot, 1), snapshot, nil
}

func BenchmarkRunFiltersUntilPassingNode(b *testing.B) {
	pod := BuildTestPod("p", 100, 1000, WithPodHostnameAntiAffinity(map[string]string{"app": "p"}), WithLabels(map[string]string{"app": "p"}))
	nodes := make([]*apiv1.Node, 0, 5001)
	podsOnNodes := make(map[string][]*apiv1.Pod)

	for i := 0; i < 5000; i++ {
		nodeName := fmt.Sprintf("n-%d", i)
		node := BuildTestNode(nodeName, 10, 1000)
		nodes = append(nodes, node)
		// Add 10 small pods to each node
		pods := make([]*apiv1.Pod, 0, 10)
		for j := 0; j < 10; j++ {
			pods = append(pods, BuildTestPod(fmt.Sprintf("p-%d-%d", i, j), 1, 1, WithLabels(map[string]string{"app": "p"})))
		}
		podsOnNodes[nodeName] = pods
	}
	// Last node is the only one that can fit the pod.
	lastNodeName := fmt.Sprintf("n-%d", len(nodes))
	lastNode := BuildTestNode(lastNodeName, 1000, 1000)
	nodes = append(nodes, lastNode)

	pluginRunner, snapshot, err := newTestPluginRunnerAndSnapshot(nil)
	assert.NoError(b, err)

	for _, node := range nodes {
		err := snapshot.AddNodeInfo(framework.NewTestNodeInfo(node, podsOnNodes[node.Name]...))
		assert.NoError(b, err)
	}

	testCases := []struct {
		parallelism int
	}{
		{parallelism: 1},
		{parallelism: 2},
		{parallelism: 4},
		{parallelism: 8},
		{parallelism: 16},
	}

	for _, tc := range testCases {
		b.Run(fmt.Sprintf("parallelism-%d", tc.parallelism), func(b *testing.B) {
			pluginRunner.parallelism = tc.parallelism
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				pluginRunner.lastIndex = 0 // Reset state for each run
				_, _, err := pluginRunner.RunFiltersUntilPassingNode(pod, func(info *framework.NodeInfo) bool { return true })
				assert.NoError(b, err)
			}
		})
	}
}
