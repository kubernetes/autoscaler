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

package scaledown

import (
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/status"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/utilization"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"

	apiv1 "k8s.io/api/core/v1"
)

// Planner is responsible for selecting nodes that should be removed.
type Planner interface {
	// UpdateClusterState provides the Planner with information about the cluster.
	UpdateClusterState(podDestinations, scaleDownCandidates []*apiv1.Node, as ActuationStatus, currentTime time.Time) errors.AutoscalerError
	// CleanUpUnneededNodes resets internal state of the Planner.
	CleanUpUnneededNodes()
	// NodesToDelete returns a list of nodes that can be deleted right now,
	// according to the Planner.
	NodesToDelete(currentTime time.Time) (empty, needDrain []*apiv1.Node)
	// UnneededNodes returns a list of nodes that either can be deleted
	// right now or in a near future, assuming nothing will change in the
	// cluster.
	UnneededNodes() []*apiv1.Node
	// UnremovableNodes returns a list of nodes that cannot be removed.
	// TODO(x13n): Add a guarantee that each node is either unneeded or
	// unremovable. This is not guaranteed by the current implementation.
	UnremovableNodes() []*simulator.UnremovableNode
	// NodeUtilizationMap returns information about utilization of
	// individual cluster nodes.
	NodeUtilizationMap() map[string]utilization.Info
}

// Actuator is responsible for making changes in the cluster: draining and
// deleting nodes.
type Actuator interface {
	// StartDeletion triggers a new deletion process. Nodes passed to this
	// function are not guaranteed to be deleted, it is possible for the
	// Actuator to ignore some of them e.g. if max configured level of
	// parallelism is reached.
	StartDeletion(empty, needDrain []*apiv1.Node) (*status.ScaleDownStatus, errors.AutoscalerError)
	// CheckStatus returns an immutable snapshot of ongoing deletions.
	CheckStatus() ActuationStatus
	// ClearResultsNotNewerThan removes information about deletions finished
	// before or exactly at the provided timestamp.
	ClearResultsNotNewerThan(time.Time)
}

// ActuationStatus is used for feeding Actuator status back into Planner
// TODO: Replace ActuationStatus with simple struct with getter methods.
type ActuationStatus interface {
	// DeletionsInProgress returns two lists of node names that are
	// currently undergoing deletion, for empty and non-empty (i.e. drained)
	// nodes separately.
	DeletionsInProgress() (empty, drained []string)
	// DeletionsCount returns total number of ongoing deletions in a given
	// node group.
	DeletionsCount(nodeGroupId string) int
	// RecentEvictions returns a list of pods that were recently removed by
	// the Actuator and hence are likely to get recreated elsewhere in the
	// cluster.
	RecentEvictions() (pods []*apiv1.Pod)
}
