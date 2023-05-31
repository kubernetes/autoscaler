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
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	apiv1 "k8s.io/api/core/v1"

	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/deletiontracker"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"
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
				SkipNodesWithSystemPods:           true,
				SkipNodesWithLocalStorage:         true,
				MinReplicaCount:                   0,
				SkipNodesWithCustomControllerPods: true,
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
