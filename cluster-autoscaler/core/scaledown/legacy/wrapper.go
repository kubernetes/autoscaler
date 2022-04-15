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
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/deletiontracker"
	"k8s.io/autoscaler/cluster-autoscaler/processors/status"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/utilization"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"

	apiv1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1beta1"
)

// ScaleDownWrapper wraps legacy scaledown logic to satisfy scaledown.Planner &
// scaledown.Actuator interfaces.
type ScaleDownWrapper struct {
	sd   *ScaleDown
	pdbs []*policyv1.PodDisruptionBudget
}

// NewScaleDownWrapper returns a new ScaleDownWrapper
func NewScaleDownWrapper(sd *ScaleDown) *ScaleDownWrapper {
	return &ScaleDownWrapper{
		sd: sd,
	}
}

// UpdateClusterState updates unneeded nodes in the underlying ScaleDown.
func (p *ScaleDownWrapper) UpdateClusterState(podDestinations, scaleDownCandidates []*apiv1.Node, actuationStatus scaledown.ActuationStatus, pdbs []*policyv1.PodDisruptionBudget, currentTime time.Time) errors.AutoscalerError {
	p.sd.CleanUp(currentTime)
	p.pdbs = pdbs
	return p.sd.UpdateUnneededNodes(podDestinations, scaleDownCandidates, currentTime, pdbs)
}

// CleanUpUnneededNodes cleans up unneeded nodes.
func (p *ScaleDownWrapper) CleanUpUnneededNodes() {
	p.sd.CleanUpUnneededNodes()
}

// NodesToDelete lists nodes to delete. Current implementation is a no-op, the
// wrapper leverages shared state instead.
// TODO(x13n): Implement this and get rid of sharing state between planning and
// actuation.
func (p *ScaleDownWrapper) NodesToDelete() (empty, needDrain []*apiv1.Node) {
	return nil, nil
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

// StartDeletion triggers an actual scale down logic.
func (p *ScaleDownWrapper) StartDeletion(empty, needDrain []*apiv1.Node, currentTime time.Time) (*status.ScaleDownStatus, errors.AutoscalerError) {
	return p.sd.TryToScaleDown(currentTime, p.pdbs)
}

// CheckStatus snapshots current deletion status
func (p *ScaleDownWrapper) CheckStatus() scaledown.ActuationStatus {
	// TODO: snapshot information from the tracker instead of keeping live
	// updated object.
	return &actuationStatus{
		ndt: p.sd.nodeDeletionTracker,
	}
}

// ClearResultsNotNewerThan clears old node deletion results kept by the
// Actuator.
func (p *ScaleDownWrapper) ClearResultsNotNewerThan(t time.Time) {
	// TODO: implement this once results are not cleared while being
	// fetched.
}

type actuationStatus struct {
	ndt *deletiontracker.NodeDeletionTracker
}

// DeletionsInProgress returns node names of currently deleted nodes.
// Current implementation is not aware of the actual nodes names, so it returns
// a fake node name instead.
// TODO: Return real node names
func (a *actuationStatus) DeletionsInProgress() []string {
	if a.ndt.IsNonEmptyNodeDeleteInProgress() {
		return []string{"fake-node-name"}
	}
	return nil
}

// DeletionsCount returns total number of ongoing deletions in a given node
// group.
func (a *actuationStatus) DeletionsCount(nodeGroupId string) int {
	return a.ndt.GetDeletionsInProgress(nodeGroupId)
}

// RecentEvictions should return a list of recently evicted pods. Since legacy
// scale down logic only drains at most one node at a time, this safeguard is
// not really needed there, so we can just return an empty list.
func (a *actuationStatus) RecentEvictions() []*apiv1.Pod {
	return nil
}

// DeletionResults returns a map of recent node deletion results.
func (a *actuationStatus) DeletionResults() map[string]status.NodeDeleteResult {
	// TODO: update nodeDeletionTracker so it doesn't get & clear in the
	// same step.
	return a.ndt.GetAndClearNodeDeleteResults()
}
