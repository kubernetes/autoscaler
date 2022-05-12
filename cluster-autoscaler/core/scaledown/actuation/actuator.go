package actuation

import (
	"time"

	apiv1 "k8s.io/api/core/v1"

	"k8s.io/autoscaler/cluster-autoscaler/clusterstate"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/deletiontracker"
	"k8s.io/autoscaler/cluster-autoscaler/processors/status"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
)

// Actuator is responsible for draining and deleting nodes.
type Actuator struct {
	ctx                 *context.AutoscalingContext
	clusterState        *clusterstate.ClusterStateRegistry
	nodeDeletionTracker *deletiontracker.NodeDeletionTracker
}

// NewActuator returns a new instance of Actuator.
func NewActuator(ctx *context.AutoscalingContext, csr *clusterstate.ClusterStateRegistry, ndr *deletiontracker.NodeDeletionTracker) *Actuator {
	return &Actuator{
		ctx:                 ctx,
		clusterState:        csr,
		nodeDeletionTracker: ndr,
	}
}

// CheckStatus should return an immutable snapshot of ongoing deletions. Before the TODO is addressed, a live object
// is returned instead of an immutable snapshot.
func (a *Actuator) CheckStatus() scaledown.ActuationStatus {
	// TODO: snapshot information from the tracker instead of keeping live
	// updated object.
	return a.nodeDeletionTracker
}

// ClearResultsNotNewerThan removes information about deletions finished before or exactly at the provided timestamp.
func (a *Actuator) ClearResultsNotNewerThan(t time.Time) {
	a.nodeDeletionTracker.ClearResultsNotNewerThan(t)
}

// StartDeletion triggers a new deletion process.
func (a *Actuator) StartDeletion(empty, drain []*apiv1.Node, currentTime time.Time) (*status.ScaleDownStatus, errors.AutoscalerError) {
	results, ts := a.nodeDeletionTracker.DeletionResults()
	scaleDownStatus := &status.ScaleDownStatus{NodeDeleteResults: results, NodeDeleteResultsAsOf: ts}

	_, _ = a.cropNodesToBudgets(empty, drain)

	// TODO: Implement deletion.
	return scaleDownStatus, nil
}

// cropNodesToBudgets crops the provided node lists to respect scale-down max parallelism budgets.
func (a *Actuator) cropNodesToBudgets(empty, needDrain []*apiv1.Node) ([]*apiv1.Node, []*apiv1.Node) {
	emptyInProgress, drainInProgress := a.nodeDeletionTracker.DeletionsInProgress()
	parallelismBudget := a.ctx.MaxScaleDownParallelism - len(emptyInProgress) - len(drainInProgress)
	drainBudget := a.ctx.MaxDrainParallelism - len(drainInProgress)

	var emptyToDelete []*apiv1.Node
	for _, node := range empty {
		if len(emptyToDelete) >= parallelismBudget {
			break
		}
		emptyToDelete = append(emptyToDelete, node)
	}

	parallelismBudgetLeft := parallelismBudget - len(emptyToDelete)
	drainBudget = min(parallelismBudgetLeft, drainBudget)

	var drainToDelete []*apiv1.Node
	for _, node := range needDrain {
		if len(drainToDelete) >= drainBudget {
			break
		}
		drainToDelete = append(drainToDelete, node)
	}

	return emptyToDelete, drainToDelete
}

func min(x, y int) int {
	if x <= y {
		return x
	}
	return y
}
