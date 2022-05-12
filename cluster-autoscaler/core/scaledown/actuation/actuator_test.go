package actuation

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/deletiontracker"
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
			emptyNodes: generateNodes(10),
			wantEmpty:  generateNodes(10),
		},
		"empty nodes exceeding max limit, no deletions in progress": {
			emptyNodes: generateNodes(11),
			wantEmpty:  generateNodes(10),
		},
		"empty nodes with deletions in progress, within budget": {
			emptyDeletionsInProgress: 1,
			drainDeletionsInProgress: 1,
			emptyNodes:               generateNodes(8),
			wantEmpty:                generateNodes(8),
		},
		"empty nodes with deletions in progress, exceeding budget": {
			emptyDeletionsInProgress: 1,
			drainDeletionsInProgress: 1,
			emptyNodes:               generateNodes(10),
			wantEmpty:                generateNodes(8),
		},
		"empty nodes with deletions in progress, 0 budget left": {
			emptyDeletionsInProgress: 5,
			drainDeletionsInProgress: 5,
			emptyNodes:               generateNodes(10),
			wantEmpty:                []*apiv1.Node{},
		},
		"empty nodes with deletions in progress, budget exceeded (shouldn't happen, just a sanity check)": {
			emptyDeletionsInProgress: 50,
			drainDeletionsInProgress: 50,
			emptyNodes:               generateNodes(10),
			wantEmpty:                []*apiv1.Node{},
		},
		// Drain nodes only.
		"drain nodes within max limit, no deletions in progress": {
			drainNodes: generateNodes(5),
			wantDrain:  generateNodes(5),
		},
		"drain nodes exceeding max limit, no deletions in progress": {
			drainNodes: generateNodes(6),
			wantDrain:  generateNodes(5),
		},
		"drain nodes with deletions in progress, within budget": {
			emptyDeletionsInProgress: 1,
			drainDeletionsInProgress: 2,
			drainNodes:               generateNodes(3),
			wantDrain:                generateNodes(3),
		},
		"drain nodes with deletions in progress, exceeding drain budget": {
			emptyDeletionsInProgress: 1,
			drainDeletionsInProgress: 2,
			drainNodes:               generateNodes(5),
			wantDrain:                generateNodes(3),
		},
		"drain nodes with deletions in progress, 0 drain budget left": {
			emptyDeletionsInProgress: 1,
			drainDeletionsInProgress: 5,
			drainNodes:               generateNodes(5),
			wantDrain:                []*apiv1.Node{},
		},
		"drain nodes with deletions in progress, drain budget exceeded (shouldn't happen, just a sanity check)": {
			emptyDeletionsInProgress: 1,
			drainDeletionsInProgress: 50,
			drainNodes:               generateNodes(5),
			wantDrain:                []*apiv1.Node{},
		},
		"drain nodes with deletions in progress, exceeding overall budget": {
			emptyDeletionsInProgress: 7,
			drainDeletionsInProgress: 1,
			drainNodes:               generateNodes(4),
			wantDrain:                generateNodes(2),
		},
		"drain nodes with deletions in progress, 0 overall budget left": {
			emptyDeletionsInProgress: 10,
			drainNodes:               generateNodes(4),
			wantDrain:                []*apiv1.Node{},
		},
		"drain nodes with deletions in progress, overall budget exceeded (shouldn't happen, just a sanity check)": {
			emptyDeletionsInProgress: 50,
			drainNodes:               generateNodes(4),
			wantDrain:                []*apiv1.Node{},
		},
		// Empty and drain nodes together.
		"empty&drain nodes within max limits, no deletions in progress": {
			emptyNodes: generateNodes(5),
			drainNodes: generateNodes(5),
			wantDrain:  generateNodes(5),
			wantEmpty:  generateNodes(5),
		},
		"empty&drain nodes exceeding overall limit, no deletions in progress": {
			emptyNodes: generateNodes(8),
			drainNodes: generateNodes(8),
			wantDrain:  generateNodes(2),
			wantEmpty:  generateNodes(8),
		},
		"empty&drain nodes exceeding drain limit, no deletions in progress": {
			emptyNodes: generateNodes(2),
			drainNodes: generateNodes(8),
			wantDrain:  generateNodes(5),
			wantEmpty:  generateNodes(2),
		},
		"empty&drain nodes with deletions in progress, 0 overall budget left": {
			emptyDeletionsInProgress: 10,
			emptyNodes:               generateNodes(5),
			drainNodes:               generateNodes(5),
			wantEmpty:                []*apiv1.Node{},
			wantDrain:                []*apiv1.Node{},
		},
		"empty&drain nodes with deletions in progress, overall budget exceeded (shouldn't happen, just a sanity check)": {
			emptyDeletionsInProgress: 50,
			emptyNodes:               generateNodes(5),
			drainNodes:               generateNodes(5),
			wantEmpty:                []*apiv1.Node{},
			wantDrain:                []*apiv1.Node{},
		},
		"empty&drain nodes with deletions in progress, 0 drain budget left": {
			drainDeletionsInProgress: 5,
			emptyNodes:               generateNodes(5),
			drainNodes:               generateNodes(5),
			wantEmpty:                generateNodes(5),
			wantDrain:                []*apiv1.Node{},
		},
		"empty&drain nodes with deletions in progress, drain budget exceeded (shouldn't happen, just a sanity check)": {
			drainDeletionsInProgress: 9,
			emptyNodes:               generateNodes(5),
			drainNodes:               generateNodes(5),
			wantEmpty:                generateNodes(1),
			wantDrain:                []*apiv1.Node{},
		},
		"empty&drain nodes with deletions in progress, overall budget exceeded, only empty nodes fit": {
			emptyDeletionsInProgress: 5,
			drainDeletionsInProgress: 3,
			emptyNodes:               generateNodes(5),
			drainNodes:               generateNodes(2),
			wantEmpty:                generateNodes(2),
			wantDrain:                []*apiv1.Node{},
		},
		"empty&drain nodes with deletions in progress, overall budget exceeded, both empty&drain nodes fit": {
			emptyDeletionsInProgress: 5,
			drainDeletionsInProgress: 3,
			emptyNodes:               generateNodes(1),
			drainNodes:               generateNodes(2),
			wantEmpty:                generateNodes(1),
			wantDrain:                generateNodes(1),
		},
		"empty&drain nodes with deletions in progress, drain budget exceeded": {
			emptyDeletionsInProgress: 1,
			drainDeletionsInProgress: 3,
			emptyNodes:               generateNodes(4),
			drainNodes:               generateNodes(5),
			wantEmpty:                generateNodes(4),
			wantDrain:                generateNodes(2),
		},
	} {
		t.Run(tn, func(t *testing.T) {
			ctx := &context.AutoscalingContext{
				AutoscalingOptions: config.AutoscalingOptions{
					MaxScaleDownParallelism: 10,
					MaxDrainParallelism:     5,
				},
			}
			ndr := deletiontracker.NewNodeDeletionTracker(1 * time.Hour)
			for i := 0; i < tc.emptyDeletionsInProgress; i++ {
				ndr.StartDeletion("ng1", fmt.Sprintf("empty-node-%d", i))
			}
			for i := 0; i < tc.drainDeletionsInProgress; i++ {
				ndr.StartDeletionWithDrain("ng2", fmt.Sprintf("drain-node-%d", i))
			}
			actuator := NewActuator(ctx, nil, ndr)
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

func generateNodes(count int) []*apiv1.Node {
	var result []*apiv1.Node
	for i := 0; i < count; i++ {
		result = append(result, &apiv1.Node{ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("node-%d", i)}})
	}
	return result
}
