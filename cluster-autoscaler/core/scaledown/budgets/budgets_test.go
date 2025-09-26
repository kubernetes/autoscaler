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

package budgets

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	testprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/deletiontracker"
	"k8s.io/autoscaler/cluster-autoscaler/core/test"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
	"k8s.io/client-go/kubernetes/fake"
)

func TestCropNodesToBudgets(t *testing.T) {
	testNg := testprovider.NewTestNodeGroup("test-ng", 0, 100, 3, true, false, "n1-standard-2", nil, nil)
	testNg2 := testprovider.NewTestNodeGroup("test-ng2", 0, 100, 3, true, false, "n1-standard-2", nil, nil)
	atomic3 := sizedNodeGroup("atomic-3", 3, true)
	atomic4 := sizedNodeGroup("atomic-4", 4, true)
	atomic5 := sizedNodeGroup("atomic-5", 5, true)
	atomic8 := sizedNodeGroup("atomic-8", 8, true)
	atomic11 := sizedNodeGroup("atomic-11", 11, true)
	for tn, tc := range map[string]struct {
		emptyDeletionsInProgress int
		drainDeletionsInProgress int
		empty                    []*NodeGroupView
		drain                    []*NodeGroupView
		wantEmpty                []*NodeGroupView
		wantDrain                []*NodeGroupView
		otherNodes               []*NodeGroupView
	}{
		"no nodes": {
			empty:     []*NodeGroupView{},
			drain:     []*NodeGroupView{},
			wantEmpty: []*NodeGroupView{},
			wantDrain: []*NodeGroupView{},
		},
		// Empty nodes only.
		"empty nodes within max limit, no deletions in progress": {
			empty:     generateNodeGroupViewList(testNg, 0, 10),
			wantEmpty: generateNodeGroupViewList(testNg, 0, 10),
		},
		"empty nodes exceeding max limit, no deletions in progress": {
			empty:     generateNodeGroupViewList(testNg, 0, 11),
			wantEmpty: generateNodeGroupViewList(testNg, 0, 10),
		},
		"empty atomic node group exceeding max limit": {
			empty:     generateNodeGroupViewList(atomic11, 0, 11),
			wantEmpty: generateNodeGroupViewList(atomic11, 0, 11),
		},
		"empty regular and atomic": {
			empty:     append(generateNodeGroupViewList(testNg, 0, 8), generateNodeGroupViewList(atomic3, 0, 3)...),
			wantEmpty: append(generateNodeGroupViewList(atomic3, 0, 3), generateNodeGroupViewList(testNg, 0, 7)...),
		},
		"multiple empty atomic": {
			empty: append(
				append(
					generateNodeGroupViewList(testNg, 0, 3),
					generateNodeGroupViewList(atomic8, 0, 8)...),
				generateNodeGroupViewList(atomic3, 0, 3)...),
			wantEmpty: append(generateNodeGroupViewList(atomic8, 0, 8), generateNodeGroupViewList(testNg, 0, 2)...),
		},
		"empty nodes with deletions in progress, within budget": {
			emptyDeletionsInProgress: 1,
			drainDeletionsInProgress: 1,
			empty:                    generateNodeGroupViewList(testNg, 0, 8),
			wantEmpty:                generateNodeGroupViewList(testNg, 0, 8),
		},
		"empty nodes with deletions in progress, exceeding budget": {
			emptyDeletionsInProgress: 1,
			drainDeletionsInProgress: 1,
			empty:                    generateNodeGroupViewList(testNg, 0, 10),
			wantEmpty:                generateNodeGroupViewList(testNg, 0, 8),
		},
		"empty atomic nodes with deletions in progress, exceeding budget": {
			emptyDeletionsInProgress: 3,
			drainDeletionsInProgress: 3,
			empty:                    generateNodeGroupViewList(atomic8, 0, 8),
			wantEmpty:                generateNodeGroupViewList(atomic8, 0, 8),
		},
		"empty nodes with deletions in progress, 0 budget left": {
			emptyDeletionsInProgress: 5,
			drainDeletionsInProgress: 5,
			empty:                    generateNodeGroupViewList(testNg, 0, 10),
			wantEmpty:                []*NodeGroupView{},
		},
		"empty atomic nodes with deletions in progress, 0 budget left": {
			emptyDeletionsInProgress: 5,
			drainDeletionsInProgress: 5,
			empty:                    generateNodeGroupViewList(atomic3, 0, 3),
			wantEmpty:                []*NodeGroupView{},
		},
		"empty nodes with deletions in progress, budget exceeded": {
			emptyDeletionsInProgress: 50,
			drainDeletionsInProgress: 50,
			empty:                    generateNodeGroupViewList(testNg, 0, 10),
			wantEmpty:                []*NodeGroupView{},
		},
		// Drain nodes only.
		"drain nodes within max limit, no deletions in progress": {
			drain:     generateNodeGroupViewList(testNg, 0, 5),
			wantDrain: generateNodeGroupViewList(testNg, 0, 5),
		},
		"multiple drain node groups": {
			drain:     append(generateNodeGroupViewList(testNg, 0, 5), generateNodeGroupViewList(testNg2, 0, 5)...),
			wantDrain: generateNodeGroupViewList(testNg, 0, 5),
		},
		"drain nodes exceeding max limit, no deletions in progress": {
			drain:     generateNodeGroupViewList(testNg, 0, 6),
			wantDrain: generateNodeGroupViewList(testNg, 0, 5),
		},
		"drain atomic exceeding limit": {
			drain:     generateNodeGroupViewList(atomic8, 0, 8),
			wantDrain: generateNodeGroupViewList(atomic8, 0, 8),
		},
		"drain regular and atomic exceeding limit": {
			drain:     append(generateNodeGroupViewList(testNg, 0, 3), generateNodeGroupViewList(atomic3, 0, 3)...),
			wantDrain: append(generateNodeGroupViewList(atomic3, 0, 3), generateNodeGroupViewList(testNg, 0, 2)...),
		},
		"multiple drain atomic": {
			drain: append(
				append(
					generateNodeGroupViewList(testNg, 0, 3),
					generateNodeGroupViewList(atomic3, 0, 3)...),
				generateNodeGroupViewList(atomic4, 0, 4)...),
			wantDrain: append(generateNodeGroupViewList(atomic3, 0, 3), generateNodeGroupViewList(testNg, 0, 2)...),
		},
		"drain nodes with deletions in progress, within budget": {
			emptyDeletionsInProgress: 1,
			drainDeletionsInProgress: 2,
			drain:                    generateNodeGroupViewList(testNg, 0, 3),
			wantDrain:                generateNodeGroupViewList(testNg, 0, 3),
		},
		"drain nodes with deletions in progress, exceeding drain budget": {
			emptyDeletionsInProgress: 1,
			drainDeletionsInProgress: 2,
			drain:                    generateNodeGroupViewList(testNg, 0, 5),
			wantDrain:                generateNodeGroupViewList(testNg, 0, 3),
		},
		"drain atomic nodes with deletions in progress, exceeding drain budget": {
			emptyDeletionsInProgress: 1,
			drainDeletionsInProgress: 2,
			drain:                    generateNodeGroupViewList(atomic4, 0, 4),
			wantDrain:                generateNodeGroupViewList(atomic4, 0, 4),
		},
		"drain nodes with deletions in progress, 0 drain budget left": {
			emptyDeletionsInProgress: 1,
			drainDeletionsInProgress: 5,
			drain:                    generateNodeGroupViewList(testNg, 0, 5),
			wantDrain:                []*NodeGroupView{},
		},
		"drain atomic nodes with deletions in progress, 0 drain budget left": {
			emptyDeletionsInProgress: 1,
			drainDeletionsInProgress: 5,
			drain:                    generateNodeGroupViewList(atomic4, 0, 4),
			wantDrain:                []*NodeGroupView{},
		},
		"drain nodes with deletions in progress, drain budget exceeded": {
			emptyDeletionsInProgress: 1,
			drainDeletionsInProgress: 50,
			drain:                    generateNodeGroupViewList(testNg, 0, 5),
			wantDrain:                []*NodeGroupView{},
		},
		"drain nodes with deletions in progress, exceeding overall budget": {
			emptyDeletionsInProgress: 7,
			drainDeletionsInProgress: 1,
			drain:                    generateNodeGroupViewList(testNg, 0, 4),
			wantDrain:                generateNodeGroupViewList(testNg, 0, 2),
		},
		"drain nodes with deletions in progress, 0 overall budget left": {
			emptyDeletionsInProgress: 10,
			drain:                    generateNodeGroupViewList(testNg, 0, 4),
			wantDrain:                []*NodeGroupView{},
		},
		"drain nodes with deletions in progress, overall budget exceeded": {
			emptyDeletionsInProgress: 50,
			drain:                    generateNodeGroupViewList(testNg, 0, 4),
			wantDrain:                []*NodeGroupView{},
		},
		// Empty and drain nodes together.
		"empty&drain nodes within max limits, no deletions in progress": {
			empty:     generateNodeGroupViewList(testNg, 0, 5),
			drain:     generateNodeGroupViewList(testNg2, 0, 5),
			wantDrain: generateNodeGroupViewList(testNg2, 0, 5),
			wantEmpty: generateNodeGroupViewList(testNg, 0, 5),
		},
		"empty&drain atomic nodes within max limits, no deletions in progress": {
			empty:     generateNodeGroupViewList(atomic3, 0, 3),
			drain:     generateNodeGroupViewList(atomic4, 0, 4),
			wantEmpty: generateNodeGroupViewList(atomic3, 0, 3),
			wantDrain: generateNodeGroupViewList(atomic4, 0, 4),
		},
		"empty&drain nodes exceeding overall limit, no deletions in progress": {
			empty:     generateNodeGroupViewList(testNg, 0, 8),
			drain:     generateNodeGroupViewList(testNg2, 0, 8),
			wantDrain: generateNodeGroupViewList(testNg2, 0, 2),
			wantEmpty: generateNodeGroupViewList(testNg, 0, 8),
		},
		"empty&drain atomic nodes exceeding overall limit, no deletions in progress": {
			empty:     generateNodeGroupViewList(atomic8, 0, 8),
			drain:     generateNodeGroupViewList(atomic4, 0, 4),
			wantEmpty: generateNodeGroupViewList(atomic8, 0, 8),
			wantDrain: []*NodeGroupView{},
		},
		"empty&drain atomic nodes in same group within overall limit, no deletions in progress": {
			empty:     generateNodeGroupViewList(atomic8, 0, 5),
			drain:     generateNodeGroupViewList(atomic8, 5, 8),
			wantEmpty: generateNodeGroupViewList(atomic8, 0, 5),
			wantDrain: generateNodeGroupViewList(atomic8, 5, 8),
		},
		"empty&drain atomic nodes in same group exceeding overall limit, no deletions in progress": {
			empty:     generateNodeGroupViewList(atomic11, 0, 8),
			drain:     generateNodeGroupViewList(atomic11, 8, 11),
			wantEmpty: generateNodeGroupViewList(atomic11, 0, 8),
			wantDrain: generateNodeGroupViewList(atomic11, 8, 11),
		},
		"empty&drain atomic nodes in same group exceeding drain limit, no deletions in progress": {
			empty:     generateNodeGroupViewList(atomic8, 0, 2),
			drain:     generateNodeGroupViewList(atomic8, 2, 8),
			wantEmpty: generateNodeGroupViewList(atomic8, 0, 2),
			wantDrain: generateNodeGroupViewList(atomic8, 2, 8),
		},
		"empty&drain atomic nodes in same group not matching registered node group size, no deletions in progress": {
			empty:      generateNodeGroupViewList(atomic8, 0, 2),
			drain:      generateNodeGroupViewList(atomic8, 2, 4),
			wantEmpty:  []*NodeGroupView{},
			wantDrain:  []*NodeGroupView{},
			otherNodes: generateNodeGroupViewList(atomic8, 4, 8),
		},
		"empty&drain atomic nodes in same group matching registered node group size, no deletions in progress": {
			empty:     generateNodeGroupViewList(atomic8, 0, 2),
			drain:     generateNodeGroupViewList(atomic8, 2, 4),
			wantEmpty: generateNodeGroupViewList(atomic8, 0, 2),
			wantDrain: generateNodeGroupViewList(atomic8, 2, 4),
		},
		"empty&drain atomic nodes exceeding drain limit, no deletions in progress": {
			empty:     generateNodeGroupViewList(atomic4, 0, 4),
			drain:     generateNodeGroupViewList(atomic8, 0, 8),
			wantEmpty: generateNodeGroupViewList(atomic4, 0, 4),
			wantDrain: []*NodeGroupView{},
		},
		"empty&drain atomic and regular nodes exceeding drain limit, no deletions in progress": {
			empty:     append(generateNodeGroupViewList(testNg, 0, 5), generateNodeGroupViewList(atomic3, 0, 3)...),
			drain:     generateNodeGroupViewList(atomic8, 0, 8),
			wantEmpty: append(generateNodeGroupViewList(atomic3, 0, 3), generateNodeGroupViewList(testNg, 0, 5)...),
			wantDrain: []*NodeGroupView{},
		},
		"empty&drain regular and atomic nodes in same group exceeding overall limit, no deletions in progress": {
			empty:     append(generateNodeGroupViewList(testNg, 0, 5), generateNodeGroupViewList(atomic11, 0, 8)...),
			drain:     generateNodeGroupViewList(atomic11, 8, 11),
			wantEmpty: generateNodeGroupViewList(atomic11, 0, 8),
			wantDrain: generateNodeGroupViewList(atomic11, 8, 11),
		},
		"empty&drain regular and multiple atomic nodes in same group exceeding drain limit, no deletions in progress": {
			empty:     append(append(generateNodeGroupViewList(testNg, 0, 5), generateNodeGroupViewList(atomic8, 0, 5)...), generateNodeGroupViewList(atomic11, 0, 8)...),
			drain:     append(generateNodeGroupViewList(atomic11, 8, 11), generateNodeGroupViewList(atomic8, 5, 8)...),
			wantEmpty: append(generateNodeGroupViewList(atomic8, 0, 5), generateNodeGroupViewList(testNg, 0, 2)...),
			wantDrain: generateNodeGroupViewList(atomic8, 5, 8),
		},
		"empty&drain multiple atomic nodes in same group exceeding overall limit, no deletions in progress": {
			empty:     append(append(generateNodeGroupViewList(atomic3, 0, 3), generateNodeGroupViewList(atomic4, 0, 2)...), generateNodeGroupViewList(atomic11, 0, 11)...),
			drain:     generateNodeGroupViewList(atomic4, 2, 4),
			wantEmpty: append(generateNodeGroupViewList(atomic3, 0, 3), generateNodeGroupViewList(atomic4, 0, 2)...),
			wantDrain: generateNodeGroupViewList(atomic4, 2, 4),
		},
		"empty regular and drain atomic nodes exceeding overall limit, no deletions in progress": {
			drain:     generateNodeGroupViewList(atomic8, 0, 8),
			empty:     generateNodeGroupViewList(testNg, 0, 5),
			wantDrain: generateNodeGroupViewList(atomic8, 0, 8),
			wantEmpty: generateNodeGroupViewList(testNg, 0, 2),
		},
		"empty&drain nodes exceeding drain limit, no deletions in progress": {
			empty:     generateNodeGroupViewList(testNg, 0, 2),
			drain:     generateNodeGroupViewList(testNg2, 0, 8),
			wantDrain: generateNodeGroupViewList(testNg2, 0, 5),
			wantEmpty: generateNodeGroupViewList(testNg, 0, 2),
		},
		"empty&drain nodes with deletions in progress, 0 overall budget left": {
			emptyDeletionsInProgress: 10,
			empty:                    generateNodeGroupViewList(testNg, 0, 5),
			drain:                    generateNodeGroupViewList(testNg2, 0, 5),
			wantEmpty:                []*NodeGroupView{},
			wantDrain:                []*NodeGroupView{},
		},
		"empty&drain atomic nodes with deletions in progress, 0 overall budget left": {
			emptyDeletionsInProgress: 10,
			empty:                    generateNodeGroupViewList(atomic4, 0, 4),
			drain:                    generateNodeGroupViewList(atomic3, 0, 3),
			wantEmpty:                []*NodeGroupView{},
			wantDrain:                []*NodeGroupView{},
		},
		"empty&drain atomic nodes in same group with deletions in progress, 0 overall budget left": {
			emptyDeletionsInProgress: 10,
			empty:                    generateNodeGroupViewList(atomic8, 0, 4),
			drain:                    generateNodeGroupViewList(atomic8, 4, 8),
			wantEmpty:                []*NodeGroupView{},
			wantDrain:                []*NodeGroupView{},
		},
		"empty&drain nodes with deletions in progress, overall budget exceeded (shouldn't happen, just a sanity check)": {
			emptyDeletionsInProgress: 50,
			empty:                    generateNodeGroupViewList(testNg, 0, 5),
			drain:                    generateNodeGroupViewList(testNg2, 0, 5),
			wantEmpty:                []*NodeGroupView{},
			wantDrain:                []*NodeGroupView{},
		},
		"empty&drain nodes with deletions in progress, 0 drain budget left": {
			drainDeletionsInProgress: 5,
			empty:                    generateNodeGroupViewList(testNg, 0, 5),
			drain:                    generateNodeGroupViewList(testNg2, 0, 5),
			wantEmpty:                generateNodeGroupViewList(testNg, 0, 5),
			wantDrain:                []*NodeGroupView{},
		},
		"empty&drain atomic nodes with deletions in progress, 0 drain budget left": {
			drainDeletionsInProgress: 5,
			empty:                    generateNodeGroupViewList(atomic4, 0, 4),
			drain:                    generateNodeGroupViewList(atomic3, 0, 3),
			wantEmpty:                generateNodeGroupViewList(atomic4, 0, 4),
			wantDrain:                []*NodeGroupView{},
		},
		"empty&drain nodes with deletions in progress, drain budget exceeded (shouldn't happen, just a sanity check)": {
			drainDeletionsInProgress: 9,
			empty:                    generateNodeGroupViewList(testNg, 0, 5),
			drain:                    generateNodeGroupViewList(testNg2, 0, 5),
			wantEmpty:                generateNodeGroupViewList(testNg, 0, 1),
			wantDrain:                []*NodeGroupView{},
		},
		"empty&drain nodes with deletions in progress, overall budget exceeded, only empty nodes fit": {
			emptyDeletionsInProgress: 5,
			drainDeletionsInProgress: 3,
			empty:                    generateNodeGroupViewList(testNg, 0, 5),
			drain:                    generateNodeGroupViewList(testNg2, 0, 2),
			wantEmpty:                generateNodeGroupViewList(testNg, 0, 2),
			wantDrain:                []*NodeGroupView{},
		},
		"empty&drain nodes with deletions in progress, overall budget exceeded, both empty&drain nodes fit": {
			emptyDeletionsInProgress: 5,
			drainDeletionsInProgress: 3,
			empty:                    generateNodeGroupViewList(testNg, 0, 1),
			drain:                    generateNodeGroupViewList(testNg2, 0, 2),
			wantEmpty:                generateNodeGroupViewList(testNg, 0, 1),
			wantDrain:                generateNodeGroupViewList(testNg2, 0, 1),
		},
		"empty&drain nodes with deletions in progress, drain budget exceeded": {
			emptyDeletionsInProgress: 1,
			drainDeletionsInProgress: 3,
			empty:                    generateNodeGroupViewList(testNg, 0, 4),
			drain:                    generateNodeGroupViewList(testNg2, 0, 5),
			wantEmpty:                generateNodeGroupViewList(testNg, 0, 4),
			wantDrain:                generateNodeGroupViewList(testNg2, 0, 2),
		},
		"empty&drain atomic nodes with deletions in progress, overall budget exceeded, only empty nodes fit": {
			emptyDeletionsInProgress: 5,
			drainDeletionsInProgress: 2,
			empty:                    generateNodeGroupViewList(atomic4, 0, 4),
			drain:                    generateNodeGroupViewList(atomic3, 0, 3),
			wantEmpty:                generateNodeGroupViewList(atomic4, 0, 4),
			wantDrain:                []*NodeGroupView{},
		},
		"empty&drain atomic nodes in same group with deletions in progress, both empty&drain nodes fit": {
			emptyDeletionsInProgress: 5,
			drainDeletionsInProgress: 2,
			empty:                    generateNodeGroupViewList(atomic3, 0, 2),
			drain:                    generateNodeGroupViewList(atomic3, 2, 3),
			wantEmpty:                generateNodeGroupViewList(atomic3, 0, 2),
			wantDrain:                generateNodeGroupViewList(atomic3, 2, 3),
		},
		"empty&drain atomic nodes in same group with deletions in progress, overall budget exceeded, both empty&drain nodes fit": {
			emptyDeletionsInProgress: 5,
			drainDeletionsInProgress: 2,
			empty:                    generateNodeGroupViewList(atomic8, 0, 6),
			drain:                    generateNodeGroupViewList(atomic8, 6, 8),
			wantEmpty:                generateNodeGroupViewList(atomic8, 0, 6),
			wantDrain:                generateNodeGroupViewList(atomic8, 6, 8),
		},
		"empty&drain atomic nodes in same group with deletions in progress, drain budget exceeded, both empty&drain nodes fit": {
			emptyDeletionsInProgress: 2,
			drainDeletionsInProgress: 2,
			empty:                    generateNodeGroupViewList(atomic5, 0, 1),
			drain:                    generateNodeGroupViewList(atomic5, 1, 5),
			wantEmpty:                generateNodeGroupViewList(atomic5, 0, 1),
			wantDrain:                generateNodeGroupViewList(atomic5, 1, 5),
		},
		"empty&drain regular and atomic nodes with deletions in progress, overall budget exceeded, only empty atomic is deleted": {
			emptyDeletionsInProgress: 5,
			drainDeletionsInProgress: 2,
			empty:                    append(generateNodeGroupViewList(testNg, 0, 4), generateNodeGroupViewList(atomic4, 0, 4)...),
			drain:                    append(generateNodeGroupViewList(testNg2, 0, 4), generateNodeGroupViewList(atomic3, 0, 3)...),
			wantEmpty:                generateNodeGroupViewList(atomic4, 0, 4),
			wantDrain:                []*NodeGroupView{},
		},
		"empty&drain regular and atomic nodes in same group with deletions in progress, overall budget exceeded, both empty&drain atomic nodes fit": {
			emptyDeletionsInProgress: 5,
			drainDeletionsInProgress: 2,
			empty:                    append(generateNodeGroupViewList(testNg, 0, 4), generateNodeGroupViewList(atomic4, 0, 2)...),
			drain:                    append(generateNodeGroupViewList(testNg2, 0, 4), generateNodeGroupViewList(atomic4, 2, 4)...),
			wantEmpty:                generateNodeGroupViewList(atomic4, 0, 2),
			wantDrain:                generateNodeGroupViewList(atomic4, 2, 4),
		},
		"empty&drain regular and atomic nodes in same group with deletions in progress, both empty&drain nodes fit": {
			emptyDeletionsInProgress: 2,
			drainDeletionsInProgress: 2,
			empty:                    append(generateNodeGroupViewList(testNg, 0, 4), generateNodeGroupViewList(atomic4, 0, 2)...),
			drain:                    append(generateNodeGroupViewList(testNg2, 0, 4), generateNodeGroupViewList(atomic4, 2, 4)...),
			wantEmpty:                append(generateNodeGroupViewList(atomic4, 0, 2), generateNodeGroupViewList(testNg, 0, 2)...),
			wantDrain:                generateNodeGroupViewList(atomic4, 2, 4),
		},
	} {
		t.Run(tn, func(t *testing.T) {
			provider := testprovider.NewTestCloudProviderBuilder().WithOnScaleDown(func(nodeGroup string, node string) error {
				return nil
			}).Build()
			allNodes := []*apiv1.Node{}
			for _, bucket := range append(append(tc.empty, tc.drain...), tc.otherNodes...) {
				bucket.Group.(*testprovider.TestNodeGroup).SetCloudProvider(provider)
				provider.InsertNodeGroup(bucket.Group)
				for _, node := range bucket.Nodes {
					provider.AddNode(bucket.Group.Id(), node)
				}
				allNodes = append(allNodes, bucket.Nodes...)
			}

			options := config.AutoscalingOptions{
				MaxScaleDownParallelism:     10,
				MaxDrainParallelism:         5,
				NodeDeletionBatcherInterval: 0 * time.Second,
				NodeDeleteDelayAfterTaint:   1 * time.Second,
			}

			ctx, err := test.NewScaleTestAutoscalingContext(options, &fake.Clientset{}, nil, provider, nil, nil)
			assert.NoError(t, err)
			ndt := deletiontracker.NewNodeDeletionTracker(1 * time.Hour)
			for i := 0; i < tc.emptyDeletionsInProgress; i++ {
				ndt.StartDeletion("ng1", fmt.Sprintf("empty-node-%d", i))
			}
			for i := 0; i < tc.drainDeletionsInProgress; i++ {
				ndt.StartDeletionWithDrain("ng2", fmt.Sprintf("drain-node-%d", i))
			}
			emptyList, drainList := []*apiv1.Node{}, []*apiv1.Node{}
			for _, bucket := range tc.empty {
				emptyList = append(emptyList, bucket.Nodes...)
			}
			for _, bucket := range tc.drain {
				drainList = append(drainList, bucket.Nodes...)
			}

			clustersnapshot.InitializeClusterSnapshotOrDie(t, ctx.ClusterSnapshot, allNodes, nil)
			budgeter := NewScaleDownBudgetProcessor(&ctx)
			gotEmpty, gotDrain := budgeter.CropNodes(ndt, emptyList, drainList)
			if diff := cmp.Diff(tc.wantEmpty, gotEmpty, cmpopts.EquateEmpty(), transformNodeGroupView); diff != "" {
				t.Errorf("cropNodesToBudgets empty nodes diff (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tc.wantDrain, gotDrain, cmpopts.EquateEmpty(), transformNodeGroupView); diff != "" {
				t.Errorf("cropNodesToBudgets drain nodes diff (-want +got):\n%s", diff)
			}
		})
	}
}

// transformNodeGroupView transforms a NodeGroupView to a structure that can be directly compared with other node bucket.
var transformNodeGroupView = cmp.Transformer("transformNodeGroupView", func(b NodeGroupView) interface{} {
	return struct {
		Group string
		Nodes []*apiv1.Node
	}{
		Group: b.Group.Id(),
		Nodes: b.Nodes,
	}
})

func sizedNodeGroup(id string, size int, atomic bool) cloudprovider.NodeGroup {
	ng := testprovider.NewTestNodeGroup(id, 10000, 0, size, true, false, "n1-standard-2", nil, nil)
	ng.SetOptions(&config.NodeGroupAutoscalingOptions{
		ZeroOrMaxNodeScaling: atomic,
	})
	return ng
}

func generateNodes(from, to int, prefix string) []*apiv1.Node {
	var result []*apiv1.Node
	for i := from; i < to; i++ {
		name := fmt.Sprintf("node-%d", i)
		if prefix != "" {
			name = prefix + "-" + name
		}
		result = append(result, generateNode(name))
	}
	return result
}

func generateNodeGroupViewList(ng cloudprovider.NodeGroup, from, to int) []*NodeGroupView {
	return []*NodeGroupView{
		{
			Group: ng,
			Nodes: generateNodes(from, to, ng.Id()),
		},
	}
}

func generateNode(name string) *apiv1.Node {
	return &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: apiv1.NodeSpec{
			ProviderID: name,
		},
		Status: apiv1.NodeStatus{
			Allocatable: apiv1.ResourceList{
				apiv1.ResourceCPU:    resource.MustParse("8"),
				apiv1.ResourceMemory: resource.MustParse("8G"),
			},
		},
	}
}
