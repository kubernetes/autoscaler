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

package actuation

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	core "k8s.io/client-go/testing"

	testprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/deletiontracker"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/status"
	. "k8s.io/autoscaler/cluster-autoscaler/core/test"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/utilization"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
)

func TestCropNodesToBudgets(t *testing.T) {
	for tn, tc := range map[string]struct {
		emptyDeletionsInProgress int
		drainDeletionsInProgress int
		emptyNodes               []*apiv1.Node
		drainNodes               []*apiv1.Node
		wantEmpty                []*apiv1.Node
		wantDrain                []*apiv1.Node
	}{
		"no nodes": {
			emptyNodes: []*apiv1.Node{},
			drainNodes: []*apiv1.Node{},
			wantEmpty:  []*apiv1.Node{},
			wantDrain:  []*apiv1.Node{},
		},
		// Empty nodes only.
		"empty nodes within max limit, no deletions in progress": {
			emptyNodes: generateNodes(10, "empty"),
			wantEmpty:  generateNodes(10, "empty"),
		},
		"empty nodes exceeding max limit, no deletions in progress": {
			emptyNodes: generateNodes(11, "empty"),
			wantEmpty:  generateNodes(10, "empty"),
		},
		"empty nodes with deletions in progress, within budget": {
			emptyDeletionsInProgress: 1,
			drainDeletionsInProgress: 1,
			emptyNodes:               generateNodes(8, "empty"),
			wantEmpty:                generateNodes(8, "empty"),
		},
		"empty nodes with deletions in progress, exceeding budget": {
			emptyDeletionsInProgress: 1,
			drainDeletionsInProgress: 1,
			emptyNodes:               generateNodes(10, "empty"),
			wantEmpty:                generateNodes(8, "empty"),
		},
		"empty nodes with deletions in progress, 0 budget left": {
			emptyDeletionsInProgress: 5,
			drainDeletionsInProgress: 5,
			emptyNodes:               generateNodes(10, "empty"),
			wantEmpty:                []*apiv1.Node{},
		},
		"empty nodes with deletions in progress, budget exceeded (shouldn't happen, just a sanity check)": {
			emptyDeletionsInProgress: 50,
			drainDeletionsInProgress: 50,
			emptyNodes:               generateNodes(10, "empty"),
			wantEmpty:                []*apiv1.Node{},
		},
		// Drain nodes only.
		"drain nodes within max limit, no deletions in progress": {
			drainNodes: generateNodes(5, "drain"),
			wantDrain:  generateNodes(5, "drain"),
		},
		"drain nodes exceeding max limit, no deletions in progress": {
			drainNodes: generateNodes(6, "drain"),
			wantDrain:  generateNodes(5, "drain"),
		},
		"drain nodes with deletions in progress, within budget": {
			emptyDeletionsInProgress: 1,
			drainDeletionsInProgress: 2,
			drainNodes:               generateNodes(3, "drain"),
			wantDrain:                generateNodes(3, "drain"),
		},
		"drain nodes with deletions in progress, exceeding drain budget": {
			emptyDeletionsInProgress: 1,
			drainDeletionsInProgress: 2,
			drainNodes:               generateNodes(5, "drain"),
			wantDrain:                generateNodes(3, "drain"),
		},
		"drain nodes with deletions in progress, 0 drain budget left": {
			emptyDeletionsInProgress: 1,
			drainDeletionsInProgress: 5,
			drainNodes:               generateNodes(5, "drain"),
			wantDrain:                []*apiv1.Node{},
		},
		"drain nodes with deletions in progress, drain budget exceeded (shouldn't happen, just a sanity check)": {
			emptyDeletionsInProgress: 1,
			drainDeletionsInProgress: 50,
			drainNodes:               generateNodes(5, "drain"),
			wantDrain:                []*apiv1.Node{},
		},
		"drain nodes with deletions in progress, exceeding overall budget": {
			emptyDeletionsInProgress: 7,
			drainDeletionsInProgress: 1,
			drainNodes:               generateNodes(4, "drain"),
			wantDrain:                generateNodes(2, "drain"),
		},
		"drain nodes with deletions in progress, 0 overall budget left": {
			emptyDeletionsInProgress: 10,
			drainNodes:               generateNodes(4, "drain"),
			wantDrain:                []*apiv1.Node{},
		},
		"drain nodes with deletions in progress, overall budget exceeded (shouldn't happen, just a sanity check)": {
			emptyDeletionsInProgress: 50,
			drainNodes:               generateNodes(4, "drain"),
			wantDrain:                []*apiv1.Node{},
		},
		// Empty and drain nodes together.
		"empty&drain nodes within max limits, no deletions in progress": {
			emptyNodes: generateNodes(5, "empty"),
			drainNodes: generateNodes(5, "drain"),
			wantDrain:  generateNodes(5, "drain"),
			wantEmpty:  generateNodes(5, "empty"),
		},
		"empty&drain nodes exceeding overall limit, no deletions in progress": {
			emptyNodes: generateNodes(8, "empty"),
			drainNodes: generateNodes(8, "drain"),
			wantDrain:  generateNodes(2, "drain"),
			wantEmpty:  generateNodes(8, "empty"),
		},
		"empty&drain nodes exceeding drain limit, no deletions in progress": {
			emptyNodes: generateNodes(2, "empty"),
			drainNodes: generateNodes(8, "drain"),
			wantDrain:  generateNodes(5, "drain"),
			wantEmpty:  generateNodes(2, "empty"),
		},
		"empty&drain nodes with deletions in progress, 0 overall budget left": {
			emptyDeletionsInProgress: 10,
			emptyNodes:               generateNodes(5, "empty"),
			drainNodes:               generateNodes(5, "drain"),
			wantEmpty:                []*apiv1.Node{},
			wantDrain:                []*apiv1.Node{},
		},
		"empty&drain nodes with deletions in progress, overall budget exceeded (shouldn't happen, just a sanity check)": {
			emptyDeletionsInProgress: 50,
			emptyNodes:               generateNodes(5, "empty"),
			drainNodes:               generateNodes(5, "drain"),
			wantEmpty:                []*apiv1.Node{},
			wantDrain:                []*apiv1.Node{},
		},
		"empty&drain nodes with deletions in progress, 0 drain budget left": {
			drainDeletionsInProgress: 5,
			emptyNodes:               generateNodes(5, "empty"),
			drainNodes:               generateNodes(5, "drain"),
			wantEmpty:                generateNodes(5, "empty"),
			wantDrain:                []*apiv1.Node{},
		},
		"empty&drain nodes with deletions in progress, drain budget exceeded (shouldn't happen, just a sanity check)": {
			drainDeletionsInProgress: 9,
			emptyNodes:               generateNodes(5, "empty"),
			drainNodes:               generateNodes(5, "drain"),
			wantEmpty:                generateNodes(1, "empty"),
			wantDrain:                []*apiv1.Node{},
		},
		"empty&drain nodes with deletions in progress, overall budget exceeded, only empty nodes fit": {
			emptyDeletionsInProgress: 5,
			drainDeletionsInProgress: 3,
			emptyNodes:               generateNodes(5, "empty"),
			drainNodes:               generateNodes(2, "drain"),
			wantEmpty:                generateNodes(2, "empty"),
			wantDrain:                []*apiv1.Node{},
		},
		"empty&drain nodes with deletions in progress, overall budget exceeded, both empty&drain nodes fit": {
			emptyDeletionsInProgress: 5,
			drainDeletionsInProgress: 3,
			emptyNodes:               generateNodes(1, "empty"),
			drainNodes:               generateNodes(2, "drain"),
			wantEmpty:                generateNodes(1, "empty"),
			wantDrain:                generateNodes(1, "drain"),
		},
		"empty&drain nodes with deletions in progress, drain budget exceeded": {
			emptyDeletionsInProgress: 1,
			drainDeletionsInProgress: 3,
			emptyNodes:               generateNodes(4, "empty"),
			drainNodes:               generateNodes(5, "drain"),
			wantEmpty:                generateNodes(4, "empty"),
			wantDrain:                generateNodes(2, "drain"),
		},
	} {
		t.Run(tn, func(t *testing.T) {
			ctx := &context.AutoscalingContext{
				AutoscalingOptions: config.AutoscalingOptions{
					MaxScaleDownParallelism:     10,
					MaxDrainParallelism:         5,
					NodeDeletionBatcherInterval: 0 * time.Second,
					NodeDeleteDelayAfterTaint:   1 * time.Second,
				},
			}
			deleteOptions := simulator.NodeDeleteOptions{
				SkipNodesWithSystemPods:   true,
				SkipNodesWithLocalStorage: true,
				MinReplicaCount:           0,
			}
			ndr := deletiontracker.NewNodeDeletionTracker(1 * time.Hour)
			for i := 0; i < tc.emptyDeletionsInProgress; i++ {
				ndr.StartDeletion("ng1", fmt.Sprintf("empty-node-%d", i))
			}
			for i := 0; i < tc.drainDeletionsInProgress; i++ {
				ndr.StartDeletionWithDrain("ng2", fmt.Sprintf("drain-node-%d", i))
			}

			actuator := NewActuator(ctx, nil, ndr, deleteOptions)
			gotEmpty, gotDrain := actuator.cropNodesToBudgets(tc.emptyNodes, tc.drainNodes)
			if diff := cmp.Diff(tc.wantEmpty, gotEmpty, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("cropNodesToBudgets empty nodes diff (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tc.wantDrain, gotDrain, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("cropNodesToBudgets drain nodes diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestStartDeletion(t *testing.T) {
	testNg := testprovider.NewTestNodeGroup("test-ng", 0, 100, 3, true, false, "n1-standard-2", nil, nil)
	toBeDeletedTaint := apiv1.Taint{Key: taints.ToBeDeletedTaint, Effect: apiv1.TaintEffectNoSchedule}

	for tn, tc := range map[string]struct {
		emptyNodes            []*apiv1.Node
		drainNodes            []*apiv1.Node
		pods                  map[string][]*apiv1.Pod
		failedPodDrain        map[string]bool
		failedNodeDeletion    map[string]bool
		failedNodeTaint       map[string]bool
		wantStatus            *status.ScaleDownStatus
		wantErr               error
		wantDeletedPods       []string
		wantDeletedNodes      []string
		wantTaintUpdates      map[string][][]apiv1.Taint
		wantNodeDeleteResults map[string]status.NodeDeleteResult
	}{
		"nothing to delete": {
			emptyNodes: nil,
			drainNodes: nil,
			wantStatus: &status.ScaleDownStatus{
				Result: status.ScaleDownNoNodeDeleted,
			},
		},
		"empty node deletion": {
			emptyNodes: generateNodes(2, "empty"),
			wantStatus: &status.ScaleDownStatus{
				Result: status.ScaleDownNodeDeleteStarted,
				ScaledDownNodes: []*status.ScaleDownNode{
					{
						Node:        generateNode("empty-node-0"),
						NodeGroup:   testNg,
						EvictedPods: nil,
						UtilInfo:    generateUtilInfo(0, 0),
					},
					{
						Node:        generateNode("empty-node-1"),
						NodeGroup:   testNg,
						EvictedPods: nil,
						UtilInfo:    generateUtilInfo(0, 0),
					},
				},
			},
			wantDeletedNodes: []string{"empty-node-0", "empty-node-1"},
			wantTaintUpdates: map[string][][]apiv1.Taint{
				"empty-node-0": {
					{toBeDeletedTaint},
				},
				"empty-node-1": {
					{toBeDeletedTaint},
				},
			},
			wantNodeDeleteResults: map[string]status.NodeDeleteResult{
				"empty-node-0": {ResultType: status.NodeDeleteOk},
				"empty-node-1": {ResultType: status.NodeDeleteOk},
			},
		},
		"deletion with drain": {
			drainNodes: generateNodes(2, "drain"),
			pods: map[string][]*apiv1.Pod{
				"drain-node-0": removablePods(2, "drain-node-0"),
				"drain-node-1": removablePods(2, "drain-node-1"),
			},
			wantStatus: &status.ScaleDownStatus{
				Result: status.ScaleDownNodeDeleteStarted,
				ScaledDownNodes: []*status.ScaleDownNode{
					{
						Node:        generateNode("drain-node-0"),
						NodeGroup:   testNg,
						EvictedPods: removablePods(2, "drain-node-0"),
						UtilInfo:    generateUtilInfo(2./8., 2./8.),
					},
					{
						Node:        generateNode("drain-node-1"),
						NodeGroup:   testNg,
						EvictedPods: removablePods(2, "drain-node-1"),
						UtilInfo:    generateUtilInfo(2./8., 2./8.),
					},
				},
			},
			wantDeletedNodes: []string{"drain-node-0", "drain-node-1"},
			wantDeletedPods:  []string{"drain-node-0-pod-0", "drain-node-0-pod-1", "drain-node-1-pod-0", "drain-node-1-pod-1"},
			wantTaintUpdates: map[string][][]apiv1.Taint{
				"drain-node-0": {
					{toBeDeletedTaint},
				},
				"drain-node-1": {
					{toBeDeletedTaint},
				},
			},
			wantNodeDeleteResults: map[string]status.NodeDeleteResult{
				"drain-node-0": {ResultType: status.NodeDeleteOk},
				"drain-node-1": {ResultType: status.NodeDeleteOk},
			},
		},
		"empty and drain deletion work correctly together": {
			emptyNodes: generateNodes(2, "empty"),
			drainNodes: generateNodes(2, "drain"),
			pods: map[string][]*apiv1.Pod{
				"drain-node-0": removablePods(2, "drain-node-0"),
				"drain-node-1": removablePods(2, "drain-node-1"),
			},
			wantStatus: &status.ScaleDownStatus{
				Result: status.ScaleDownNodeDeleteStarted,
				ScaledDownNodes: []*status.ScaleDownNode{
					{
						Node:        generateNode("empty-node-0"),
						NodeGroup:   testNg,
						EvictedPods: nil,
						UtilInfo:    generateUtilInfo(0, 0),
					},
					{
						Node:        generateNode("empty-node-1"),
						NodeGroup:   testNg,
						EvictedPods: nil,
						UtilInfo:    generateUtilInfo(0, 0),
					},
					{
						Node:        generateNode("drain-node-0"),
						NodeGroup:   testNg,
						EvictedPods: removablePods(2, "drain-node-0"),
						UtilInfo:    generateUtilInfo(2./8., 2./8.),
					},
					{
						Node:        generateNode("drain-node-1"),
						NodeGroup:   testNg,
						EvictedPods: removablePods(2, "drain-node-1"),
						UtilInfo:    generateUtilInfo(2./8., 2./8.),
					},
				},
			},
			wantDeletedNodes: []string{"empty-node-0", "empty-node-1", "drain-node-0", "drain-node-1"},
			wantDeletedPods:  []string{"drain-node-0-pod-0", "drain-node-0-pod-1", "drain-node-1-pod-0", "drain-node-1-pod-1"},
			wantTaintUpdates: map[string][][]apiv1.Taint{
				"empty-node-0": {
					{toBeDeletedTaint},
				},
				"empty-node-1": {
					{toBeDeletedTaint},
				},
				"drain-node-0": {
					{toBeDeletedTaint},
				},
				"drain-node-1": {
					{toBeDeletedTaint},
				},
			},
			wantNodeDeleteResults: map[string]status.NodeDeleteResult{
				"empty-node-0": {ResultType: status.NodeDeleteOk},
				"empty-node-1": {ResultType: status.NodeDeleteOk},
				"drain-node-0": {ResultType: status.NodeDeleteOk},
				"drain-node-1": {ResultType: status.NodeDeleteOk},
			},
		},
		"failure to taint empty node stops deletion and cleans already applied taints": {
			emptyNodes: generateNodes(4, "empty"),
			drainNodes: generateNodes(1, "drain"),
			pods: map[string][]*apiv1.Pod{
				"drain-node-0": removablePods(2, "drain-node-0"),
			},
			failedNodeTaint: map[string]bool{"empty-node-2": true},
			wantStatus: &status.ScaleDownStatus{
				Result:          status.ScaleDownError,
				ScaledDownNodes: nil,
			},
			wantTaintUpdates: map[string][][]apiv1.Taint{
				"empty-node-0": {
					{toBeDeletedTaint},
					{},
				},
				"empty-node-1": {
					{toBeDeletedTaint},
					{},
				},
			},
			wantErr: cmpopts.AnyError,
		},
		"failure to taint drain node stops further deletion and cleans already applied taints": {
			emptyNodes: generateNodes(2, "empty"),
			drainNodes: generateNodes(4, "drain"),
			pods: map[string][]*apiv1.Pod{
				"drain-node-0": removablePods(2, "drain-node-0"),
				"drain-node-1": removablePods(2, "drain-node-1"),
				"drain-node-2": removablePods(2, "drain-node-2"),
				"drain-node-3": removablePods(2, "drain-node-3"),
			},
			failedNodeTaint: map[string]bool{"drain-node-2": true},
			wantStatus: &status.ScaleDownStatus{
				Result: status.ScaleDownError,
				ScaledDownNodes: []*status.ScaleDownNode{
					{
						Node:        generateNode("empty-node-0"),
						NodeGroup:   testNg,
						EvictedPods: nil,
						UtilInfo:    generateUtilInfo(0, 0),
					},
					{
						Node:        generateNode("empty-node-1"),
						NodeGroup:   testNg,
						EvictedPods: nil,
						UtilInfo:    generateUtilInfo(0, 0),
					},
				},
			},
			wantDeletedNodes: []string{"empty-node-0", "empty-node-1"},
			wantTaintUpdates: map[string][][]apiv1.Taint{
				"empty-node-0": {
					{toBeDeletedTaint},
				},
				"empty-node-1": {
					{toBeDeletedTaint},
				},
			},
			wantNodeDeleteResults: map[string]status.NodeDeleteResult{
				"empty-node-0": {ResultType: status.NodeDeleteOk},
				"empty-node-1": {ResultType: status.NodeDeleteOk},
			},
			wantErr: cmpopts.AnyError,
		},
		"nodes that failed drain are correctly reported in results": {
			drainNodes: generateNodes(4, "drain"),
			pods: map[string][]*apiv1.Pod{
				"drain-node-0": removablePods(3, "drain-node-0"),
				"drain-node-1": removablePods(3, "drain-node-1"),
				"drain-node-2": removablePods(3, "drain-node-2"),
				"drain-node-3": removablePods(3, "drain-node-3"),
			},
			failedPodDrain: map[string]bool{
				"drain-node-0-pod-0": true,
				"drain-node-0-pod-1": true,
				"drain-node-2-pod-1": true,
			},
			wantStatus: &status.ScaleDownStatus{
				Result: status.ScaleDownNodeDeleteStarted,
				ScaledDownNodes: []*status.ScaleDownNode{
					{
						Node:        generateNode("drain-node-0"),
						NodeGroup:   testNg,
						EvictedPods: removablePods(3, "drain-node-0"),
						UtilInfo:    generateUtilInfo(3./8., 3./8.),
					},
					{
						Node:        generateNode("drain-node-1"),
						NodeGroup:   testNg,
						EvictedPods: removablePods(3, "drain-node-1"),
						UtilInfo:    generateUtilInfo(3./8., 3./8.),
					},
					{
						Node:        generateNode("drain-node-2"),
						NodeGroup:   testNg,
						EvictedPods: removablePods(3, "drain-node-2"),
						UtilInfo:    generateUtilInfo(3./8., 3./8.),
					},
					{
						Node:        generateNode("drain-node-3"),
						NodeGroup:   testNg,
						EvictedPods: removablePods(3, "drain-node-3"),
						UtilInfo:    generateUtilInfo(3./8., 3./8.),
					},
				},
			},
			wantDeletedNodes: []string{"drain-node-1", "drain-node-3"},
			wantDeletedPods: []string{
				"drain-node-0-pod-2",
				"drain-node-1-pod-0", "drain-node-1-pod-1", "drain-node-1-pod-2",
				"drain-node-2-pod-0", "drain-node-2-pod-2",
				"drain-node-3-pod-0", "drain-node-3-pod-1", "drain-node-3-pod-2",
			},
			wantTaintUpdates: map[string][][]apiv1.Taint{
				"drain-node-0": {
					{toBeDeletedTaint},
					{},
				},
				"drain-node-1": {
					{toBeDeletedTaint},
				},
				"drain-node-2": {
					{toBeDeletedTaint},
					{},
				},
				"drain-node-3": {
					{toBeDeletedTaint},
				},
			},
			wantNodeDeleteResults: map[string]status.NodeDeleteResult{
				"drain-node-0": {
					ResultType: status.NodeDeleteErrorFailedToEvictPods,
					Err:        cmpopts.AnyError,
					PodEvictionResults: map[string]status.PodEvictionResult{
						"drain-node-0-pod-0": {Pod: removablePod("drain-node-0-pod-0", "drain-node-0"), Err: cmpopts.AnyError, TimedOut: true},
						"drain-node-0-pod-1": {Pod: removablePod("drain-node-0-pod-1", "drain-node-0"), Err: cmpopts.AnyError, TimedOut: true},
						"drain-node-0-pod-2": {Pod: removablePod("drain-node-0-pod-2", "drain-node-0")},
					},
				},
				"drain-node-1": {ResultType: status.NodeDeleteOk},
				"drain-node-2": {
					ResultType: status.NodeDeleteErrorFailedToEvictPods,
					Err:        cmpopts.AnyError,
					PodEvictionResults: map[string]status.PodEvictionResult{
						"drain-node-2-pod-0": {Pod: removablePod("drain-node-2-pod-0", "drain-node-2")},
						"drain-node-2-pod-1": {Pod: removablePod("drain-node-2-pod-1", "drain-node-2"), Err: cmpopts.AnyError, TimedOut: true},
						"drain-node-2-pod-2": {Pod: removablePod("drain-node-2-pod-2", "drain-node-2")},
					},
				},
				"drain-node-3": {ResultType: status.NodeDeleteOk},
			},
		},
		"nodes that failed deletion are correctly reported in results": {
			emptyNodes: generateNodes(2, "empty"),
			drainNodes: generateNodes(2, "drain"),
			pods: map[string][]*apiv1.Pod{
				"drain-node-0": removablePods(2, "drain-node-0"),
				"drain-node-1": removablePods(2, "drain-node-1"),
			},
			failedNodeDeletion: map[string]bool{
				"empty-node-1": true,
				"drain-node-1": true,
			},
			wantStatus: &status.ScaleDownStatus{
				Result: status.ScaleDownNodeDeleteStarted,
				ScaledDownNodes: []*status.ScaleDownNode{
					{
						Node:        generateNode("empty-node-0"),
						NodeGroup:   testNg,
						EvictedPods: nil,
						UtilInfo:    generateUtilInfo(0, 0),
					},
					{
						Node:        generateNode("empty-node-1"),
						NodeGroup:   testNg,
						EvictedPods: nil,
						UtilInfo:    generateUtilInfo(0, 0),
					},
					{
						Node:        generateNode("drain-node-0"),
						NodeGroup:   testNg,
						EvictedPods: removablePods(2, "drain-node-0"),
						UtilInfo:    generateUtilInfo(2./8., 2./8.),
					},
					{
						Node:        generateNode("drain-node-1"),
						NodeGroup:   testNg,
						EvictedPods: removablePods(2, "drain-node-1"),
						UtilInfo:    generateUtilInfo(2./8., 2./8.),
					},
				},
			},
			wantDeletedNodes: []string{"empty-node-0", "drain-node-0"},
			wantDeletedPods: []string{
				"drain-node-0-pod-0", "drain-node-0-pod-1",
				"drain-node-1-pod-0", "drain-node-1-pod-1",
			},
			wantTaintUpdates: map[string][][]apiv1.Taint{
				"empty-node-0": {
					{toBeDeletedTaint},
				},
				"empty-node-1": {
					{toBeDeletedTaint},
					{},
				},
				"drain-node-0": {
					{toBeDeletedTaint},
				},
				"drain-node-1": {
					{toBeDeletedTaint},
					{},
				},
			},
			wantNodeDeleteResults: map[string]status.NodeDeleteResult{
				"empty-node-0": {ResultType: status.NodeDeleteOk},
				"empty-node-1": {ResultType: status.NodeDeleteErrorFailedToDelete, Err: cmpopts.AnyError},
				"drain-node-0": {ResultType: status.NodeDeleteOk},
				"drain-node-1": {ResultType: status.NodeDeleteErrorFailedToDelete, Err: cmpopts.AnyError},
			},
		},
		"DS pods are evicted from empty nodes, but don't block deletion on error": {
			emptyNodes: generateNodes(2, "empty"),
			pods: map[string][]*apiv1.Pod{
				"empty-node-0": {generateDsPod("empty-node-0-ds-pod-0", "empty-node-0"), generateDsPod("empty-node-0-ds-pod-1", "empty-node-0")},
				"empty-node-1": {generateDsPod("empty-node-1-ds-pod-0", "empty-node-1"), generateDsPod("empty-node-1-ds-pod-1", "empty-node-1")},
			},
			failedPodDrain: map[string]bool{"empty-node-1-ds-pod-0": true},
			wantStatus: &status.ScaleDownStatus{
				Result: status.ScaleDownNodeDeleteStarted,
				ScaledDownNodes: []*status.ScaleDownNode{
					{
						Node:        generateNode("empty-node-0"),
						NodeGroup:   testNg,
						EvictedPods: nil,
						UtilInfo:    generateUtilInfo(2./8., 2./8.),
					},
					{
						Node:        generateNode("empty-node-1"),
						NodeGroup:   testNg,
						EvictedPods: nil,
						UtilInfo:    generateUtilInfo(2./8., 2./8.),
					},
				},
			},
			wantDeletedNodes: []string{"empty-node-0", "empty-node-1"},
			wantDeletedPods:  []string{"empty-node-0-ds-pod-0", "empty-node-0-ds-pod-1", "empty-node-1-ds-pod-1"},
			wantTaintUpdates: map[string][][]apiv1.Taint{
				"empty-node-0": {
					{toBeDeletedTaint},
				},
				"empty-node-1": {
					{toBeDeletedTaint},
				},
			},
			wantNodeDeleteResults: map[string]status.NodeDeleteResult{
				"empty-node-0": {ResultType: status.NodeDeleteOk},
				"empty-node-1": {ResultType: status.NodeDeleteOk},
			},
		},
		"nodes with pods are not deleted if the node is passed as empty": {
			emptyNodes: generateNodes(2, "empty-but-with-pods"),
			pods: map[string][]*apiv1.Pod{
				"empty-but-with-pods-node-0": removablePods(2, "empty-but-with-pods-node-0"),
				"empty-but-with-pods-node-1": removablePods(2, "empty-but-with-pods-node-1"),
			},
			wantStatus: &status.ScaleDownStatus{
				Result: status.ScaleDownNodeDeleteStarted,
				ScaledDownNodes: []*status.ScaleDownNode{
					{
						Node:        generateNode("empty-but-with-pods-node-0"),
						NodeGroup:   testNg,
						EvictedPods: nil,
						UtilInfo:    generateUtilInfo(2./8., 2./8.),
					},
					{
						Node:        generateNode("empty-but-with-pods-node-1"),
						NodeGroup:   testNg,
						EvictedPods: nil,
						UtilInfo:    generateUtilInfo(2./8., 2./8.),
					},
				},
			},
			wantDeletedNodes: nil,
			wantDeletedPods:  nil,
			wantTaintUpdates: map[string][][]apiv1.Taint{
				"empty-but-with-pods-node-0": {
					{toBeDeletedTaint},
					{},
				},
				"empty-but-with-pods-node-1": {
					{toBeDeletedTaint},
					{},
				},
			},
			wantNodeDeleteResults: map[string]status.NodeDeleteResult{
				"empty-but-with-pods-node-0": {ResultType: status.NodeDeleteErrorInternal, Err: cmpopts.AnyError},
				"empty-but-with-pods-node-1": {ResultType: status.NodeDeleteErrorInternal, Err: cmpopts.AnyError},
			},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			// This is needed because the tested code starts goroutines that can technically live longer than the execution
			// of a single test case, and the goroutines eventually access tc in fakeClient hooks below.
			tc := tc
			// Insert all nodes into a map to support live node updates and GETs.
			nodesByName := make(map[string]*apiv1.Node)
			nodesLock := sync.Mutex{}
			for _, node := range tc.emptyNodes {
				nodesByName[node.Name] = node
			}
			for _, node := range tc.drainNodes {
				nodesByName[node.Name] = node
			}

			// Set up a fake k8s client to hook and verify certain actions.
			fakeClient := &fake.Clientset{}
			type nodeTaints struct {
				nodeName string
				taints   []apiv1.Taint
			}
			taintUpdates := make(chan nodeTaints, 10)
			deletedNodes := make(chan string, 10)
			deletedPods := make(chan string, 10)

			ds := generateDaemonSet()

			// We're faking the whole k8s client, and some of the code needs to get live nodes and pods, so GET on nodes and pods has to be set up.
			fakeClient.Fake.AddReactor("get", "nodes", func(action core.Action) (bool, runtime.Object, error) {
				nodesLock.Lock()
				defer nodesLock.Unlock()
				getAction := action.(core.GetAction)
				node, found := nodesByName[getAction.GetName()]
				if !found {
					return true, nil, fmt.Errorf("node %q not found", getAction.GetName())
				}
				return true, node, nil
			})
			fakeClient.Fake.AddReactor("get", "pods",
				func(action core.Action) (bool, runtime.Object, error) {
					return true, nil, errors.NewNotFound(apiv1.Resource("pod"), "whatever")
				})
			// Hook node update to gather all taint updates, and to fail the update for certain nodes to simulate errors.
			fakeClient.Fake.AddReactor("update", "nodes",
				func(action core.Action) (bool, runtime.Object, error) {
					nodesLock.Lock()
					defer nodesLock.Unlock()
					update := action.(core.UpdateAction)
					obj := update.GetObject().(*apiv1.Node)
					if tc.failedNodeTaint[obj.Name] {
						return true, nil, fmt.Errorf("SIMULATED ERROR: won't taint")
					}
					nt := nodeTaints{
						nodeName: obj.Name,
					}
					for _, taint := range obj.Spec.Taints {
						nt.taints = append(nt.taints, taint)
					}
					taintUpdates <- nt
					nodesByName[obj.Name] = obj.DeepCopy()
					return true, obj, nil
				})
			// Hook eviction creation to gather which pods were evicted, and to fail the eviction for certain pods to simulate errors.
			fakeClient.Fake.AddReactor("create", "pods",
				func(action core.Action) (bool, runtime.Object, error) {
					createAction := action.(core.CreateAction)
					if createAction == nil {
						return false, nil, nil
					}
					eviction := createAction.GetObject().(*policyv1beta1.Eviction)
					if eviction == nil {
						return false, nil, nil
					}
					if tc.failedPodDrain[eviction.Name] {
						return true, nil, fmt.Errorf("SIMULATED ERROR: won't evict")
					}
					deletedPods <- eviction.Name
					return true, nil, nil
				})

			// Hook node deletion at the level of cloud provider, to gather which nodes were deleted, and to fail the deletion for
			// certain nodes to simulate errors.
			provider := testprovider.NewTestCloudProvider(nil, func(nodeGroup string, node string) error {
				if tc.failedNodeDeletion[node] {
					return fmt.Errorf("SIMULATED ERROR: won't remove node")
				}
				deletedNodes <- node
				return nil
			})
			testNg.SetCloudProvider(provider)
			provider.InsertNodeGroup(testNg)
			for _, node := range nodesByName {
				provider.AddNode("test-ng", node)
			}

			// Set up other needed structures and options.
			opts := config.AutoscalingOptions{
				MaxScaleDownParallelism:        10,
				MaxDrainParallelism:            5,
				MaxPodEvictionTime:             0,
				DaemonSetEvictionForEmptyNodes: true,
			}

			allPods := []*apiv1.Pod{}

			for _, pods := range tc.pods {
				allPods = append(allPods, pods...)
			}

			podLister := kube_util.NewTestPodLister(allPods)
			pdbLister := kube_util.NewTestPodDisruptionBudgetLister([]*policyv1.PodDisruptionBudget{})
			dsLister, err := kube_util.NewTestDaemonSetLister([]*appsv1.DaemonSet{ds})
			if err != nil {
				t.Fatalf("Couldn't create daemonset lister")
			}

			registry := kube_util.NewListerRegistry(nil, nil, podLister, nil, pdbLister, dsLister, nil, nil, nil, nil)
			ctx, err := NewScaleTestAutoscalingContext(opts, fakeClient, registry, provider, nil, nil)
			if err != nil {
				t.Fatalf("Couldn't set up autoscaling context: %v", err)
			}
			csr := clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{}, ctx.LogRecorder, NewBackoff())
			for _, node := range tc.emptyNodes {
				err := ctx.ClusterSnapshot.AddNodeWithPods(node, tc.pods[node.Name])
				if err != nil {
					t.Fatalf("Couldn't add node %q to snapshot: %v", node.Name, err)
				}
			}
			for _, node := range tc.drainNodes {
				pods, found := tc.pods[node.Name]
				if !found {
					t.Fatalf("Drain node %q doesn't have pods defined in the test case.", node.Name)
				}
				err := ctx.ClusterSnapshot.AddNodeWithPods(node, pods)
				if err != nil {
					t.Fatalf("Couldn't add node %q to snapshot: %v", node.Name, err)
				}
			}

			// Create Actuator, run StartDeletion, and verify the error.
			ndt := deletiontracker.NewNodeDeletionTracker(0)
			actuator := Actuator{
				ctx: &ctx, clusterState: csr, nodeDeletionTracker: ndt,
				nodeDeletionBatcher: NewNodeDeletionBatcher(&ctx, csr, ndt, 0*time.Second),
				evictor:             Evictor{EvictionRetryTime: 0, DsEvictionRetryTime: 0, DsEvictionEmptyNodeTimeout: 0, PodEvictionHeadroom: DefaultPodEvictionHeadroom},
			}
			gotStatus, gotErr := actuator.StartDeletion(tc.emptyNodes, tc.drainNodes)
			if diff := cmp.Diff(tc.wantErr, gotErr, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("StartDeletion error diff (-want +got):\n%s", diff)
			}

			// Verify ScaleDownStatus looks as expected.
			ignoreSdNodeOrder := cmpopts.SortSlices(func(a, b *status.ScaleDownNode) bool { return a.Node.Name < b.Node.Name })
			ignoreTimestamps := cmpopts.IgnoreFields(status.ScaleDownStatus{}, "NodeDeleteResultsAsOf")
			cmpNg := cmp.Comparer(func(a, b *testprovider.TestNodeGroup) bool { return a.Id() == b.Id() })
			statusCmpOpts := cmp.Options{ignoreSdNodeOrder, ignoreTimestamps, cmpNg, cmpopts.EquateEmpty()}
			if diff := cmp.Diff(tc.wantStatus, gotStatus, statusCmpOpts); diff != "" {
				t.Errorf("StartDeletion status diff (-want +got):\n%s", diff)
			}

			// Verify that all expected nodes were deleted using the cloud provider hook.
			var gotDeletedNodes []string
		nodesLoop:
			for i := 0; i < len(tc.wantDeletedNodes); i++ {
				select {
				case deletedNode := <-deletedNodes:
					gotDeletedNodes = append(gotDeletedNodes, deletedNode)
				case <-time.After(3 * time.Second):
					t.Errorf("Timeout while waiting for deleted nodes.")
					break nodesLoop
				}
			}
			ignoreStrOrder := cmpopts.SortSlices(func(a, b string) bool { return a < b })
			if diff := cmp.Diff(tc.wantDeletedNodes, gotDeletedNodes, ignoreStrOrder); diff != "" {
				t.Errorf("deletedNodes diff (-want +got):\n%s", diff)
			}

			// Verify that all expected pods were deleted using the fake k8s client hook.
			var gotDeletedPods []string
		podsLoop:
			for i := 0; i < len(tc.wantDeletedPods); i++ {
				select {
				case deletedPod := <-deletedPods:
					gotDeletedPods = append(gotDeletedPods, deletedPod)
				case <-time.After(3 * time.Second):
					t.Errorf("Timeout while waiting for deleted pods.")
					break podsLoop
				}
			}
			if diff := cmp.Diff(tc.wantDeletedPods, gotDeletedPods, ignoreStrOrder); diff != "" {
				t.Errorf("deletedPods diff (-want +got):\n%s", diff)
			}

			// Verify that all expected taint updates happened using the fake k8s client hook.
			allUpdatesCount := 0
			for _, updates := range tc.wantTaintUpdates {
				allUpdatesCount += len(updates)
			}
			gotTaintUpdates := make(map[string][][]apiv1.Taint)
		taintsLoop:
			for i := 0; i < allUpdatesCount; i++ {
				select {
				case taintUpdate := <-taintUpdates:
					gotTaintUpdates[taintUpdate.nodeName] = append(gotTaintUpdates[taintUpdate.nodeName], taintUpdate.taints)
				case <-time.After(3 * time.Second):
					t.Errorf("Timeout while waiting for taint updates.")
					break taintsLoop
				}
			}
			ignoreTaintValue := cmpopts.IgnoreFields(apiv1.Taint{}, "Value")
			if diff := cmp.Diff(tc.wantTaintUpdates, gotTaintUpdates, ignoreTaintValue, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("taintUpdates diff (-want +got):\n%s", diff)
			}

			// Wait for all expected deletions to be reported in NodeDeletionTracker. Reporting happens shortly after the deletion
			// in cloud provider we sync to above and so this will usually not wait at all. However, it can still happen
			// that there is a delay between cloud provider deletion and reporting, in which case the results are not there yet
			// and we need to wait for them before asserting.
			err = waitForDeletionResultsCount(actuator.nodeDeletionTracker, len(tc.wantNodeDeleteResults), 3*time.Second, 200*time.Millisecond)
			if err != nil {
				t.Errorf("Timeout while waiting for node deletion results")
			}

			// Run StartDeletion again to gather node deletion results for deletions started in the previous call, and verify
			// that they look as expected.
			gotNextStatus, gotNextErr := actuator.StartDeletion(nil, nil)
			if gotNextErr != nil {
				t.Errorf("StartDeletion unexpected error: %v", gotNextErr)
			}
			if diff := cmp.Diff(tc.wantNodeDeleteResults, gotNextStatus.NodeDeleteResults, cmpopts.EquateEmpty(), cmpopts.EquateErrors()); diff != "" {
				t.Errorf("NodeDeleteResults diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestStartDeletionInBatchBasic(t *testing.T) {
	deleteInterval := 1 * time.Second

	for _, test := range []struct {
		name                   string
		deleteCalls            int
		numNodesToDelete       map[string][]int //per node group and per call
		failedRequests         map[string]bool  //per node group
		wantSuccessfulDeletion map[string]int   //per node group
	}{
		{
			name:        "Succesfull deletion for all node group",
			deleteCalls: 1,
			numNodesToDelete: map[string][]int{
				"test-ng-1": {4},
				"test-ng-2": {5},
				"test-ng-3": {1},
			},
			wantSuccessfulDeletion: map[string]int{
				"test-ng-1": 4,
				"test-ng-2": 5,
				"test-ng-3": 1,
			},
		},
		{
			name:        "Node deletion failed for one group",
			deleteCalls: 1,
			numNodesToDelete: map[string][]int{
				"test-ng-1": {4},
				"test-ng-2": {5},
				"test-ng-3": {1},
			},
			failedRequests: map[string]bool{
				"test-ng-1": true,
			},
			wantSuccessfulDeletion: map[string]int{
				"test-ng-1": 0,
				"test-ng-2": 5,
				"test-ng-3": 1,
			},
		},
		{
			name:        "Node deletion failed for one group two times",
			deleteCalls: 2,
			numNodesToDelete: map[string][]int{
				"test-ng-1": {4, 3},
				"test-ng-2": {5},
				"test-ng-3": {1},
			},
			failedRequests: map[string]bool{
				"test-ng-1": true,
			},
			wantSuccessfulDeletion: map[string]int{
				"test-ng-1": 0,
				"test-ng-2": 5,
				"test-ng-3": 1,
			},
		},
		{
			name:        "Node deletion failed for all groups",
			deleteCalls: 2,
			numNodesToDelete: map[string][]int{
				"test-ng-1": {4, 3},
				"test-ng-2": {5},
				"test-ng-3": {1},
			},
			failedRequests: map[string]bool{
				"test-ng-1": true,
				"test-ng-2": true,
				"test-ng-3": true,
			},
			wantSuccessfulDeletion: map[string]int{
				"test-ng-1": 0,
				"test-ng-2": 0,
				"test-ng-3": 0,
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			test := test
			gotFailedRequest := func(nodeGroupId string) bool {
				val, _ := test.failedRequests[nodeGroupId]
				return val
			}
			deletedResult := make(chan string)
			fakeClient := &fake.Clientset{}
			provider := testprovider.NewTestCloudProvider(nil, func(nodeGroupId string, node string) error {
				if gotFailedRequest(nodeGroupId) {
					return fmt.Errorf("SIMULATED ERROR: won't remove node")
				}
				deletedResult <- nodeGroupId
				return nil
			})
			// 2d array represent the waves of pushing nodes to delete.
			deleteNodes := [][]*apiv1.Node{}

			for i := 0; i < test.deleteCalls; i++ {
				deleteNodes = append(deleteNodes, []*apiv1.Node{})
			}
			testNg1 := testprovider.NewTestNodeGroup("test-ng-1", 0, 100, 3, true, false, "n1-standard-2", nil, nil)
			testNg2 := testprovider.NewTestNodeGroup("test-ng-2", 0, 100, 3, true, false, "n1-standard-2", nil, nil)
			testNg3 := testprovider.NewTestNodeGroup("test-ng-3", 0, 100, 3, true, false, "n1-standard-2", nil, nil)
			testNg := map[string]*testprovider.TestNodeGroup{
				"test-ng-1": testNg1,
				"test-ng-2": testNg2,
				"test-ng-3": testNg3,
			}

			for ngName, numNodes := range test.numNodesToDelete {
				ng := testNg[ngName]
				provider.InsertNodeGroup(ng)
				ng.SetCloudProvider(provider)
				for i, num := range numNodes {
					nodes := generateNodes(num, ng.Id())
					deleteNodes[i] = append(deleteNodes[i], nodes...)
					for _, node := range nodes {
						provider.AddNode(ng.Id(), node)
					}
				}
			}
			opts := config.AutoscalingOptions{
				MaxScaleDownParallelism:        10,
				MaxDrainParallelism:            5,
				MaxPodEvictionTime:             0,
				DaemonSetEvictionForEmptyNodes: true,
			}

			podLister := kube_util.NewTestPodLister([]*apiv1.Pod{})
			pdbLister := kube_util.NewTestPodDisruptionBudgetLister([]*policyv1.PodDisruptionBudget{})
			registry := kube_util.NewListerRegistry(nil, nil, podLister, nil, pdbLister, nil, nil, nil, nil, nil)
			ctx, err := NewScaleTestAutoscalingContext(opts, fakeClient, registry, provider, nil, nil)
			if err != nil {
				t.Fatalf("Couldn't set up autoscaling context: %v", err)
			}
			csr := clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{}, ctx.LogRecorder, NewBackoff())
			ndt := deletiontracker.NewNodeDeletionTracker(0)
			actuator := Actuator{
				ctx: &ctx, clusterState: csr, nodeDeletionTracker: ndt,
				nodeDeletionBatcher: NewNodeDeletionBatcher(&ctx, csr, ndt, deleteInterval),
				evictor:             Evictor{EvictionRetryTime: 0, DsEvictionRetryTime: 0, DsEvictionEmptyNodeTimeout: 0, PodEvictionHeadroom: DefaultPodEvictionHeadroom},
			}

			for _, nodes := range deleteNodes {
				actuator.StartDeletion(nodes, []*apiv1.Node{})
				time.Sleep(deleteInterval)
			}
			wantDeletedNodes := 0
			for _, num := range test.wantSuccessfulDeletion {
				wantDeletedNodes += num
			}
			gotDeletedNodes := map[string]int{
				"test-ng-1": 0,
				"test-ng-2": 0,
				"test-ng-3": 0,
			}
			for i := 0; i < wantDeletedNodes; i++ {
				select {
				case ngId := <-deletedResult:
					gotDeletedNodes[ngId]++
				case <-time.After(1 * time.Second):
					t.Errorf("Timeout while waiting for deleted nodes.")
					break
				}
			}
			if diff := cmp.Diff(test.wantSuccessfulDeletion, gotDeletedNodes); diff != "" {
				t.Errorf("Successful deleteions per node group diff (-want +got):\n%s", diff)
			}
		})
	}
}

func generateNodes(count int, prefix string) []*apiv1.Node {
	var result []*apiv1.Node
	for i := 0; i < count; i++ {
		name := fmt.Sprintf("node-%d", i)
		if prefix != "" {
			name = prefix + "-" + name
		}
		result = append(result, generateNode(name))
	}
	return result
}

func generateNodesAndNodeGroupMap(count int, prefix string) map[string]*testprovider.TestNodeGroup {
	result := make(map[string]*testprovider.TestNodeGroup)
	for i := 0; i < count; i++ {
		name := fmt.Sprintf("node-%d", i)
		ngName := fmt.Sprintf("test-ng-%v", i)
		if prefix != "" {
			name = prefix + "-" + name
			ngName = prefix + "-" + ngName
		}
		result[name] = testprovider.NewTestNodeGroup(ngName, 0, 100, 3, true, false, "n1-standard-2", nil, nil)
	}
	return result
}

func generateNode(name string) *apiv1.Node {
	return &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Status: apiv1.NodeStatus{
			Allocatable: apiv1.ResourceList{
				apiv1.ResourceCPU:    resource.MustParse("8"),
				apiv1.ResourceMemory: resource.MustParse("8G"),
			},
		},
	}
}

func removablePods(count int, prefix string) []*apiv1.Pod {
	var result []*apiv1.Pod
	for i := 0; i < count; i++ {
		name := fmt.Sprintf("pod-%d", i)
		if prefix != "" {
			name = prefix + "-" + name
		}
		result = append(result, removablePod(name, prefix))
	}
	return result
}

func removablePod(name string, node string) *apiv1.Pod {
	return &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
			Annotations: map[string]string{
				"cluster-autoscaler.kubernetes.io/safe-to-evict": "true",
			},
		},
		Spec: apiv1.PodSpec{
			NodeName: node,
			Containers: []apiv1.Container{
				{
					Name: "test-container",
					Resources: apiv1.ResourceRequirements{
						Requests: map[apiv1.ResourceName]resource.Quantity{
							apiv1.ResourceCPU:    resource.MustParse("1"),
							apiv1.ResourceMemory: resource.MustParse("1G"),
						},
					},
				},
			},
		},
	}
}

func generateDsPod(name string, node string) *apiv1.Pod {
	pod := removablePod(name, node)
	pod.OwnerReferences = GenerateOwnerReferences("ds", "DaemonSet", "apps/v1", "some-uid")
	return pod
}

func generateDaemonSet() *appsv1.DaemonSet {
	return &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ds",
			Namespace: "default",
			SelfLink:  "/apiv1s/apps/v1/namespaces/default/daemonsets/ds",
		},
	}
}

func generateUtilInfo(cpuUtil, memUtil float64) utilization.Info {
	var higherUtilName apiv1.ResourceName
	var higherUtilVal float64
	if cpuUtil > memUtil {
		higherUtilName = apiv1.ResourceCPU
		higherUtilVal = cpuUtil
	} else {
		higherUtilName = apiv1.ResourceMemory
		higherUtilVal = memUtil
	}
	return utilization.Info{
		CpuUtil:      cpuUtil,
		MemUtil:      memUtil,
		ResourceName: higherUtilName,
		Utilization:  higherUtilVal,
	}
}

func waitForDeletionResultsCount(ndt *deletiontracker.NodeDeletionTracker, resultsCount int, timeout, retryTime time.Duration) error {
	// This is quite ugly, but shouldn't matter much since in most cases there shouldn't be a need to wait at all, and
	// the function should return quickly after the first if check.
	// An alternative could be to turn NodeDeletionTracker into an interface, and use an implementation which allows
	// synchronizing calls to EndDeletion in the test code.
	for retryUntil := time.Now().Add(timeout); time.Now().Before(retryUntil); time.Sleep(retryTime) {
		if results, _ := ndt.DeletionResults(); len(results) == resultsCount {
			return nil
		}
	}
	return fmt.Errorf("timed out while waiting for node deletion results")
}
