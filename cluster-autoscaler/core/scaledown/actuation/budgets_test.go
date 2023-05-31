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

	testprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/deletiontracker"
)

func TestCropNodesToBudgets(t *testing.T) {
	testNg := testprovider.NewTestNodeGroup("test-ng", 0, 100, 3, true, false, "n1-standard-2", nil, nil)
	atomic3 := sizedNodeGroup("atomic-3", 3, true)
	atomic4 := sizedNodeGroup("atomic-4", 4, true)
	atomic8 := sizedNodeGroup("atomic-8", 8, true)
	atomic11 := sizedNodeGroup("atomic-11", 11, true)
	for tn, tc := range map[string]struct {
		emptyDeletionsInProgress int
		drainDeletionsInProgress int
		empty                    []*nodeBucket
		drain                    []*nodeBucket
		wantEmpty                []*nodeBucket
		wantDrain                []*nodeBucket
	}{
		"no nodes": {
			empty:     []*nodeBucket{},
			drain:     []*nodeBucket{},
			wantEmpty: []*nodeBucket{},
			wantDrain: []*nodeBucket{},
		},
		// Empty nodes only.
		"empty nodes within max limit, no deletions in progress": {
			empty:     generatenodeBucketList(testNg, 0, 10),
			wantEmpty: generatenodeBucketList(testNg, 0, 10),
		},
		"empty nodes exceeding max limit, no deletions in progress": {
			empty:     generatenodeBucketList(testNg, 0, 11),
			wantEmpty: generatenodeBucketList(testNg, 0, 10),
		},
		"empty atomic node group exceeding max limit": {
			empty:     generatenodeBucketList(atomic11, 0, 11),
			wantEmpty: generatenodeBucketList(atomic11, 0, 11),
		},
		"empty regular and atomic": {
			empty:     append(generatenodeBucketList(testNg, 0, 8), generatenodeBucketList(atomic3, 0, 3)...),
			wantEmpty: append(generatenodeBucketList(atomic3, 0, 3), generatenodeBucketList(testNg, 0, 7)...),
		},
		"multiple empty atomic": {
			empty: append(
				append(
					generatenodeBucketList(testNg, 0, 3),
					generatenodeBucketList(atomic8, 0, 8)...),
				generatenodeBucketList(atomic3, 0, 3)...),
			wantEmpty: append(generatenodeBucketList(atomic8, 0, 8), generatenodeBucketList(testNg, 0, 2)...),
		},
		"empty nodes with deletions in progress, within budget": {
			emptyDeletionsInProgress: 1,
			drainDeletionsInProgress: 1,
			empty:                    generatenodeBucketList(testNg, 0, 8),
			wantEmpty:                generatenodeBucketList(testNg, 0, 8),
		},
		"empty nodes with deletions in progress, exceeding budget": {
			emptyDeletionsInProgress: 1,
			drainDeletionsInProgress: 1,
			empty:                    generatenodeBucketList(testNg, 0, 10),
			wantEmpty:                generatenodeBucketList(testNg, 0, 8),
		},
		"empty atomic nodes with deletions in progress, exceeding budget": {
			emptyDeletionsInProgress: 3,
			drainDeletionsInProgress: 3,
			empty:                    generatenodeBucketList(atomic8, 0, 8),
			wantEmpty:                generatenodeBucketList(atomic8, 0, 8),
		},
		"empty nodes with deletions in progress, 0 budget left": {
			emptyDeletionsInProgress: 5,
			drainDeletionsInProgress: 5,
			empty:                    generatenodeBucketList(testNg, 0, 10),
			wantEmpty:                []*nodeBucket{},
		},
		"empty atomic nodes with deletions in progress, 0 budget left": {
			emptyDeletionsInProgress: 5,
			drainDeletionsInProgress: 5,
			empty:                    generatenodeBucketList(atomic3, 0, 3),
			wantEmpty:                []*nodeBucket{},
		},
		"empty nodes with deletions in progress, budget exceeded": {
			emptyDeletionsInProgress: 50,
			drainDeletionsInProgress: 50,
			empty:                    generatenodeBucketList(testNg, 0, 10),
			wantEmpty:                []*nodeBucket{},
		},
		// Drain nodes only.
		"drain nodes within max limit, no deletions in progress": {
			drain:     generatenodeBucketList(testNg, 0, 5),
			wantDrain: generatenodeBucketList(testNg, 0, 5),
		},
		"drain nodes exceeding max limit, no deletions in progress": {
			drain:     generatenodeBucketList(testNg, 0, 6),
			wantDrain: generatenodeBucketList(testNg, 0, 5),
		},
		"drain atomic exceeding limit": {
			drain:     generatenodeBucketList(atomic8, 0, 8),
			wantDrain: generatenodeBucketList(atomic8, 0, 8),
		},
		"drain regular and atomic exceeding limit": {
			drain:     append(generatenodeBucketList(testNg, 0, 3), generatenodeBucketList(atomic3, 0, 3)...),
			wantDrain: append(generatenodeBucketList(atomic3, 0, 3), generatenodeBucketList(testNg, 0, 2)...),
		},
		"multiple drain atomic": {
			drain: append(
				append(
					generatenodeBucketList(testNg, 0, 3),
					generatenodeBucketList(atomic3, 0, 3)...),
				generatenodeBucketList(atomic4, 0, 4)...),
			wantDrain: append(generatenodeBucketList(atomic3, 0, 3), generatenodeBucketList(testNg, 0, 2)...),
		},
		"drain nodes with deletions in progress, within budget": {
			emptyDeletionsInProgress: 1,
			drainDeletionsInProgress: 2,
			drain:                    generatenodeBucketList(testNg, 0, 3),
			wantDrain:                generatenodeBucketList(testNg, 0, 3),
		},
		"drain nodes with deletions in progress, exceeding drain budget": {
			emptyDeletionsInProgress: 1,
			drainDeletionsInProgress: 2,
			drain:                    generatenodeBucketList(testNg, 0, 5),
			wantDrain:                generatenodeBucketList(testNg, 0, 3),
		},
		"drain atomic nodes with deletions in progress, exceeding drain budget": {
			emptyDeletionsInProgress: 1,
			drainDeletionsInProgress: 2,
			drain:                    generatenodeBucketList(atomic4, 0, 4),
			wantDrain:                generatenodeBucketList(atomic4, 0, 4),
		},
		"drain nodes with deletions in progress, 0 drain budget left": {
			emptyDeletionsInProgress: 1,
			drainDeletionsInProgress: 5,
			drain:                    generatenodeBucketList(testNg, 0, 5),
			wantDrain:                []*nodeBucket{},
		},
		"drain atomic nodes with deletions in progress, 0 drain budget left": {
			emptyDeletionsInProgress: 1,
			drainDeletionsInProgress: 5,
			drain:                    generatenodeBucketList(atomic4, 0, 4),
			wantDrain:                []*nodeBucket{},
		},
		"drain nodes with deletions in progress, drain budget exceeded": {
			emptyDeletionsInProgress: 1,
			drainDeletionsInProgress: 50,
			drain:                    generatenodeBucketList(testNg, 0, 5),
			wantDrain:                []*nodeBucket{},
		},
		"drain nodes with deletions in progress, exceeding overall budget": {
			emptyDeletionsInProgress: 7,
			drainDeletionsInProgress: 1,
			drain:                    generatenodeBucketList(testNg, 0, 4),
			wantDrain:                generatenodeBucketList(testNg, 0, 2),
		},
		"drain nodes with deletions in progress, 0 overall budget left": {
			emptyDeletionsInProgress: 10,
			drain:                    generatenodeBucketList(testNg, 0, 4),
			wantDrain:                []*nodeBucket{},
		},
		"drain nodes with deletions in progress, overall budget exceeded": {
			emptyDeletionsInProgress: 50,
			drain:                    generatenodeBucketList(testNg, 0, 4),
			wantDrain:                []*nodeBucket{},
		},
		// Empty and drain nodes together.
		"empty&drain nodes within max limits, no deletions in progress": {
			empty:     generatenodeBucketList(testNg, 0, 5),
			drain:     generatenodeBucketList(testNg, 0, 5),
			wantDrain: generatenodeBucketList(testNg, 0, 5),
			wantEmpty: generatenodeBucketList(testNg, 0, 5),
		},
		"empty&drain atomic nodes within max limits, no deletions in progress": {
			empty:     generatenodeBucketList(atomic3, 0, 3),
			drain:     generatenodeBucketList(atomic4, 0, 4),
			wantEmpty: generatenodeBucketList(atomic3, 0, 3),
			wantDrain: generatenodeBucketList(atomic4, 0, 4),
		},
		"empty&drain nodes exceeding overall limit, no deletions in progress": {
			empty:     generatenodeBucketList(testNg, 0, 8),
			drain:     generatenodeBucketList(testNg, 0, 8),
			wantDrain: generatenodeBucketList(testNg, 0, 2),
			wantEmpty: generatenodeBucketList(testNg, 0, 8),
		},
		"empty&drain atomic nodes exceeding overall limit, no deletions in progress": {
			empty:     generatenodeBucketList(atomic8, 0, 8),
			drain:     generatenodeBucketList(atomic4, 0, 4),
			wantEmpty: generatenodeBucketList(atomic8, 0, 8),
			wantDrain: []*nodeBucket{},
		},
		"empty&drain atomic nodes exceeding drain limit, no deletions in progress": {
			empty:     generatenodeBucketList(atomic4, 0, 4),
			drain:     generatenodeBucketList(atomic8, 0, 8),
			wantEmpty: generatenodeBucketList(atomic4, 0, 4),
			wantDrain: []*nodeBucket{},
		},
		"empty&drain atomic and regular nodes exceeding drain limit, no deletions in progress": {
			empty:     append(generatenodeBucketList(testNg, 0, 5), generatenodeBucketList(atomic3, 0, 3)...),
			drain:     generatenodeBucketList(atomic8, 0, 8),
			wantEmpty: append(generatenodeBucketList(atomic3, 0, 3), generatenodeBucketList(testNg, 0, 5)...),
			wantDrain: []*nodeBucket{},
		},
		"empty regular and drain atomic nodes exceeding overall limit, no deletions in progress": {
			drain:     generatenodeBucketList(atomic8, 0, 8),
			empty:     generatenodeBucketList(testNg, 0, 5),
			wantDrain: generatenodeBucketList(atomic8, 0, 8),
			wantEmpty: generatenodeBucketList(testNg, 0, 2),
		},
		"empty&drain nodes exceeding drain limit, no deletions in progress": {
			empty:     generatenodeBucketList(testNg, 0, 2),
			drain:     generatenodeBucketList(testNg, 0, 8),
			wantDrain: generatenodeBucketList(testNg, 0, 5),
			wantEmpty: generatenodeBucketList(testNg, 0, 2),
		},
		"empty&drain nodes with deletions in progress, 0 overall budget left": {
			emptyDeletionsInProgress: 10,
			empty:                    generatenodeBucketList(testNg, 0, 5),
			drain:                    generatenodeBucketList(testNg, 0, 5),
			wantEmpty:                []*nodeBucket{},
			wantDrain:                []*nodeBucket{},
		},
		"empty&drain nodes with deletions in progress, overall budget exceeded (shouldn't happen, just a sanity check)": {
			emptyDeletionsInProgress: 50,
			empty:                    generatenodeBucketList(testNg, 0, 5),
			drain:                    generatenodeBucketList(testNg, 0, 5),
			wantEmpty:                []*nodeBucket{},
			wantDrain:                []*nodeBucket{},
		},
		"empty&drain nodes with deletions in progress, 0 drain budget left": {
			drainDeletionsInProgress: 5,
			empty:                    generatenodeBucketList(testNg, 0, 5),
			drain:                    generatenodeBucketList(testNg, 0, 5),
			wantEmpty:                generatenodeBucketList(testNg, 0, 5),
			wantDrain:                []*nodeBucket{},
		},
		"empty&drain nodes with deletions in progress, drain budget exceeded (shouldn't happen, just a sanity check)": {
			drainDeletionsInProgress: 9,
			empty:                    generatenodeBucketList(testNg, 0, 5),
			drain:                    generatenodeBucketList(testNg, 0, 5),
			wantEmpty:                generatenodeBucketList(testNg, 0, 1),
			wantDrain:                []*nodeBucket{},
		},
		"empty&drain nodes with deletions in progress, overall budget exceeded, only empty nodes fit": {
			emptyDeletionsInProgress: 5,
			drainDeletionsInProgress: 3,
			empty:                    generatenodeBucketList(testNg, 0, 5),
			drain:                    generatenodeBucketList(testNg, 0, 2),
			wantEmpty:                generatenodeBucketList(testNg, 0, 2),
			wantDrain:                []*nodeBucket{},
		},
		"empty&drain nodes with deletions in progress, overall budget exceeded, both empty&drain nodes fit": {
			emptyDeletionsInProgress: 5,
			drainDeletionsInProgress: 3,
			empty:                    generatenodeBucketList(testNg, 0, 1),
			drain:                    generatenodeBucketList(testNg, 0, 2),
			wantEmpty:                generatenodeBucketList(testNg, 0, 1),
			wantDrain:                generatenodeBucketList(testNg, 0, 1),
		},
		"empty&drain nodes with deletions in progress, drain budget exceeded": {
			emptyDeletionsInProgress: 1,
			drainDeletionsInProgress: 3,
			empty:                    generatenodeBucketList(testNg, 0, 4),
			drain:                    generatenodeBucketList(testNg, 0, 5),
			wantEmpty:                generatenodeBucketList(testNg, 0, 4),
			wantDrain:                generatenodeBucketList(testNg, 0, 2),
		},
	} {
		t.Run(tn, func(t *testing.T) {
			provider := testprovider.NewTestCloudProvider(nil, func(nodeGroup string, node string) error {
				return nil
			})
			for _, bucket := range append(tc.empty, tc.drain...) {
				bucket.group.(*testprovider.TestNodeGroup).SetCloudProvider(provider)
				provider.InsertNodeGroup(bucket.group)
				for _, node := range bucket.nodes {
					provider.AddNode(bucket.group.Id(), node)
				}
			}

			ctx := &context.AutoscalingContext{
				AutoscalingOptions: config.AutoscalingOptions{
					MaxScaleDownParallelism:     10,
					MaxDrainParallelism:         5,
					NodeDeletionBatcherInterval: 0 * time.Second,
					NodeDeleteDelayAfterTaint:   1 * time.Second,
				},
				CloudProvider: provider,
			}
			ndt := deletiontracker.NewNodeDeletionTracker(1 * time.Hour)
			for i := 0; i < tc.emptyDeletionsInProgress; i++ {
				ndt.StartDeletion("ng1", fmt.Sprintf("empty-node-%d", i))
			}
			for i := 0; i < tc.drainDeletionsInProgress; i++ {
				ndt.StartDeletionWithDrain("ng2", fmt.Sprintf("drain-node-%d", i))
			}
			emptyList, drainList := []*apiv1.Node{}, []*apiv1.Node{}
			for _, bucket := range tc.empty {
				emptyList = append(emptyList, bucket.nodes...)
			}
			for _, bucket := range tc.drain {
				drainList = append(drainList, bucket.nodes...)
			}

			budgeter := NewScaleDownBudgetProcessor(ctx, ndt)
			gotEmpty, gotDrain := budgeter.CropNodes(emptyList, drainList)
			// a
			if diff := cmp.Diff(tc.wantEmpty, gotEmpty, cmpopts.EquateEmpty(), transformnodeBucket); diff != "" {
				t.Errorf("cropNodesToBudgets empty nodes diff (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tc.wantDrain, gotDrain, cmpopts.EquateEmpty(), transformnodeBucket); diff != "" {
				t.Errorf("cropNodesToBudgets drain nodes diff (-want +got):\n%s", diff)
			}
		})
	}
}

// transformnodeBucket transforms a nodeBucket to a structure that can be directly compared with other node bucket.
var transformnodeBucket = cmp.Transformer("transformnodeBucket", func(b nodeBucket) interface{} {
	return struct {
		Group string
		Nodes []*apiv1.Node
	}{
		Group: b.group.Id(),
		Nodes: b.nodes,
	}
})
