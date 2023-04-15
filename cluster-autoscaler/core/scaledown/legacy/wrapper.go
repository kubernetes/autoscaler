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

package legacy

import (
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/actuation"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/status"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/utilization"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"

	apiv1 "k8s.io/api/core/v1"
)

// ScaleDownWrapper wraps legacy scaledown logic to satisfy scaledown.Planner &
// scaledown.Actuator interfaces.
type ScaleDownWrapper struct {
	sd                      *ScaleDown
	actuator                *actuation.Actuator
	lastNodesToDeleteResult status.ScaleDownResult
	lastNodesToDeleteErr    errors.AutoscalerError
}

// NewScaleDownWrapper returns a new ScaleDownWrapper
func NewScaleDownWrapper(sd *ScaleDown, actuator *actuation.Actuator) *ScaleDownWrapper {
	return &ScaleDownWrapper{
		sd:       sd,
		actuator: actuator,
	}
}

// UpdateClusterState updates unneeded nodes in the underlying ScaleDown.
func (p *ScaleDownWrapper) UpdateClusterState(podDestinations, scaleDownCandidates []*apiv1.Node, actuationStatus scaledown.ActuationStatus, currentTime time.Time) errors.AutoscalerError {
	p.sd.CleanUp(currentTime)
	return p.sd.UpdateUnneededNodes(podDestinations, scaleDownCandidates, currentTime)
}

// CleanUpUnneededNodes cleans up unneeded nodes.
func (p *ScaleDownWrapper) CleanUpUnneededNodes() {
	p.sd.CleanUpUnneededNodes()
}

// UnneededNodes returns a list of unneeded nodes.
func (p *ScaleDownWrapper) UnneededNodes() []*apiv1.Node {
	return p.sd.UnneededNodes()
}

// UnremovableNodes returns a list of nodes that cannot be removed.
func (p *ScaleDownWrapper) UnremovableNodes() []*simulator.UnremovableNode {
	return p.sd.UnremovableNodes()
}

// NodeUtilizationMap returns information about utilization of individual
// cluster nodes.
func (p *ScaleDownWrapper) NodeUtilizationMap() map[string]utilization.Info {
	return p.sd.NodeUtilizationMap()
}

// NodesToDelete lists nodes to delete.
//
// The legacy implementation had one status for getting nodes to delete and actually deleting them, so some of
// status.Result values are specific to NodesToDelete. In order not to break the processors that might be depending
// on these values, the Result is still passed between NodesToDelete and StartDeletion. The legacy implementation would
// also short-circuit in case of any errors, while current NodesToDelete doesn't return an error. To preserve that behavior,
// the error returned by legacy TryToScaleDown (now called NodesToDelete) is also passed to StartDeletion.
// TODO: Evaluate if we can get rid of the last bits of shared state.
func (p *ScaleDownWrapper) NodesToDelete(currentTime time.Time) (empty, needDrain []*apiv1.Node) {
	empty, drain, result, err := p.sd.NodesToDelete(currentTime)
	p.lastNodesToDeleteResult = result
	p.lastNodesToDeleteErr = err
	return empty, drain
}

// StartDeletion triggers an actual scale down logic.
func (p *ScaleDownWrapper) StartDeletion(empty, needDrain []*apiv1.Node) (*status.ScaleDownStatus, errors.AutoscalerError) {
	// Done to preserve legacy behavior, see comment on NodesToDelete.
	if p.lastNodesToDeleteErr != nil || p.lastNodesToDeleteResult != status.ScaleDownNodeDeleteStarted {
		// When there is no need for scale-down, p.lastNodesToDeleteResult is set to ScaleDownNoUnneeded. We have to still report node delete
		// results in this case, otherwise they wouldn't get reported until the next call to actuator.StartDeletion (i.e. until the next scale-down
		// attempt).
		// Run actuator.StartDeletion with no nodes just to grab the delete results.
		origStatus, _ := p.actuator.StartDeletion(nil, nil)
		return &status.ScaleDownStatus{
			Result:                p.lastNodesToDeleteResult,
			NodeDeleteResults:     origStatus.NodeDeleteResults,
			NodeDeleteResultsAsOf: origStatus.NodeDeleteResultsAsOf,
		}, p.lastNodesToDeleteErr
	}
	return p.actuator.StartDeletion(empty, needDrain)
}

// CheckStatus snapshots current deletion status
func (p *ScaleDownWrapper) CheckStatus() scaledown.ActuationStatus {
	return p.actuator.CheckStatus()
}

// ClearResultsNotNewerThan clears old node deletion results kept by the
// Actuator.
func (p *ScaleDownWrapper) ClearResultsNotNewerThan(t time.Time) {
	p.actuator.ClearResultsNotNewerThan(t)
}
