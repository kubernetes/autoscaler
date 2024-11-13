/*
Copyright 2020 The Kubernetes Authors.

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

package predicatechecker

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	testconfig "k8s.io/autoscaler/cluster-autoscaler/config/test"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/utils/scheduler"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"

	apiv1 "k8s.io/api/core/v1"
)

func TestCheckPredicate(t *testing.T) {
	p450 := BuildTestPod("p450", 450, 500000)
	p600 := BuildTestPod("p600", 600, 500000)
	p8000 := BuildTestPod("p8000", 8000, 0)
	p500 := BuildTestPod("p500", 500, 500000)

	n1000 := BuildTestNode("n1000", 1000, 2000000)
	SetNodeReadyState(n1000, true, time.Time{})
	n1000Unschedulable := BuildTestNode("n1000", 1000, 2000000)
	SetNodeReadyState(n1000Unschedulable, true, time.Time{})

	defaultPredicateChecker, err := NewTestPredicateChecker()
	assert.NoError(t, err)

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
	customPredicateChecker, err := NewTestPredicateCheckerWithCustomConfig(customConfig)
	assert.NoError(t, err)

	tests := []struct {
		name             string
		node             *apiv1.Node
		scheduledPods    []*apiv1.Pod
		testPod          *apiv1.Pod
		predicateChecker PredicateChecker
		expectError      bool
	}{
		// default predicate checker test cases
		{
			name:             "default - other pod - insuficient cpu",
			node:             n1000,
			scheduledPods:    []*apiv1.Pod{p450},
			testPod:          p600,
			expectError:      true,
			predicateChecker: defaultPredicateChecker,
		},
		{
			name:             "default - other pod - ok",
			node:             n1000,
			scheduledPods:    []*apiv1.Pod{p450},
			testPod:          p500,
			expectError:      false,
			predicateChecker: defaultPredicateChecker,
		},
		{
			name:             "default - empty - insuficient cpu",
			node:             n1000,
			scheduledPods:    []*apiv1.Pod{},
			testPod:          p8000,
			expectError:      true,
			predicateChecker: defaultPredicateChecker,
		},
		{
			name:             "default - empty - ok",
			node:             n1000,
			scheduledPods:    []*apiv1.Pod{},
			testPod:          p600,
			expectError:      false,
			predicateChecker: defaultPredicateChecker,
		},
		// custom predicate checker test cases
		{
			name:             "custom - other pod - ok",
			node:             n1000,
			scheduledPods:    []*apiv1.Pod{p450},
			testPod:          p600,
			expectError:      false,
			predicateChecker: customPredicateChecker,
		},
		{
			name:             "custom -other pod - ok",
			node:             n1000,
			scheduledPods:    []*apiv1.Pod{p450},
			testPod:          p500,
			expectError:      false,
			predicateChecker: customPredicateChecker,
		},
		{
			name:             "custom -empty - ok",
			node:             n1000,
			scheduledPods:    []*apiv1.Pod{},
			testPod:          p8000,
			expectError:      false,
			predicateChecker: customPredicateChecker,
		},
		{
			name:             "custom -empty - ok",
			node:             n1000,
			scheduledPods:    []*apiv1.Pod{},
			testPod:          p600,
			expectError:      false,
			predicateChecker: customPredicateChecker,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			clusterSnapshot := clustersnapshot.NewBasicClusterSnapshot()
			err = clusterSnapshot.AddNodeInfo(framework.NewTestNodeInfo(tt.node, tt.scheduledPods...))
			assert.NoError(t, err)

			predicateError := tt.predicateChecker.CheckPredicates(clusterSnapshot, tt.testPod, tt.node.Name)
			if tt.expectError {
				assert.NotNil(t, predicateError)
				assert.Equal(t, clustersnapshot.FailingPredicateError, predicateError.Type())
				assert.Equal(t, "NodeResourcesFit", predicateError.FailingPredicateName())
				assert.Equal(t, []string{"Insufficient cpu"}, predicateError.FailingPredicateReasons())
				assert.Contains(t, predicateError.Error(), "NodeResourcesFit")
				assert.Contains(t, predicateError.Error(), "Insufficient cpu")
			} else {
				assert.Nil(t, predicateError)
			}
		})
	}
}

func TestFitsAnyNode(t *testing.T) {
	p900 := BuildTestPod("p900", 900, 1000)
	p1900 := BuildTestPod("p1900", 1900, 1000)
	p2100 := BuildTestPod("p2100", 2100, 1000)

	n1000 := BuildTestNode("n1000", 1000, 2000000)
	n2000 := BuildTestNode("n2000", 2000, 2000000)

	defaultPredicateChecker, err := NewTestPredicateChecker()
	assert.NoError(t, err)

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
	customPredicateChecker, err := NewTestPredicateCheckerWithCustomConfig(customConfig)
	assert.NoError(t, err)

	testCases := []struct {
		name             string
		predicateChecker PredicateChecker
		pod              *apiv1.Pod
		expectedNodes    []string
		expectError      bool
	}{
		// default predicate checker test cases
		{
			name:             "default - small pod - no error",
			predicateChecker: defaultPredicateChecker,
			pod:              p900,
			expectedNodes:    []string{"n1000", "n2000"},
			expectError:      false,
		},
		{
			name:             "default - medium pod - no error",
			predicateChecker: defaultPredicateChecker,
			pod:              p1900,
			expectedNodes:    []string{"n2000"},
			expectError:      false,
		},
		{
			name:             "default - large pod - insufficient cpu",
			predicateChecker: defaultPredicateChecker,
			pod:              p2100,
			expectError:      true,
		},

		// custom predicate checker test cases
		{
			name:             "custom - small pod - no error",
			predicateChecker: customPredicateChecker,
			pod:              p900,
			expectedNodes:    []string{"n1000", "n2000"},
			expectError:      false,
		},
		{
			name:             "custom - medium pod - no error",
			predicateChecker: customPredicateChecker,
			pod:              p1900,
			expectedNodes:    []string{"n1000", "n2000"},
			expectError:      false,
		},
		{
			name:             "custom - large pod - insufficient cpu",
			predicateChecker: customPredicateChecker,
			pod:              p2100,
			expectedNodes:    []string{"n1000", "n2000"},
			expectError:      false,
		},
	}

	clusterSnapshot := clustersnapshot.NewBasicClusterSnapshot()
	err = clusterSnapshot.AddNodeInfo(framework.NewTestNodeInfo(n1000))
	assert.NoError(t, err)
	err = clusterSnapshot.AddNodeInfo(framework.NewTestNodeInfo(n2000))
	assert.NoError(t, err)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			nodeName, err := tc.predicateChecker.FitsAnyNode(clusterSnapshot, tc.pod)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Contains(t, tc.expectedNodes, nodeName)
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

	clusterSnapshot := clustersnapshot.NewBasicClusterSnapshot()

	err := clusterSnapshot.AddNodeInfo(framework.NewTestNodeInfo(node1))
	assert.NoError(t, err)

	// with default predicate checker
	defaultPredicateChecker, err := NewTestPredicateChecker()
	assert.NoError(t, err)
	predicateErr := defaultPredicateChecker.CheckPredicates(clusterSnapshot, p1, "n1")
	assert.NotNil(t, predicateErr)
	assert.Contains(t, predicateErr.FailingPredicateReasons(), "node(s) had untolerated taint {SomeTaint: WhyNot?}")
	assert.Contains(t, predicateErr.Error(), "node(s) had untolerated taint {SomeTaint: WhyNot?}")
	assert.Contains(t, predicateErr.Error(), "RandomTaint")

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
	customPredicateChecker, err := NewTestPredicateCheckerWithCustomConfig(customConfig)
	assert.NoError(t, err)
	predicateErr = customPredicateChecker.CheckPredicates(clusterSnapshot, p1, "n1")
	assert.Nil(t, predicateErr)
}
