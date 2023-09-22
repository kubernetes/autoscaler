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
	"context"

	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"

	apiv1 "k8s.io/api/core/v1"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
)

// PredicateChecker checks whether all required predicates pass for given Pod and Node.
type PredicateChecker interface {
	// Snapshot gets called at the beginning of a scale up simulation,
	// before any of the other methods. In contrast to the check methods
	// it is allowed to modify the snapshot.
	Snapshot(ctx context.Context, clusterSnapshot clustersnapshot.ClusterSnapshot) error

	FitsAnyNode(clusterSnapshot clustersnapshot.ClusterSnapshot, pod *apiv1.Pod) (string, error)
	FitsAnyNodeMatching(clusterSnapshot clustersnapshot.ClusterSnapshot, pod *apiv1.Pod, nodeMatches func(*schedulerframework.NodeInfo) bool) (string, error)
	CheckPredicates(clusterSnapshot clustersnapshot.ClusterSnapshot, pod *apiv1.Pod, nodeName string) *PredicateError

	// BindPod gets called to record in the cluster snapshot that a pod was scheduled
	// to a node as part of the simulation.
	BindPod(ctx context.Context, clusterSnapshot clustersnapshot.ClusterSnapshot, pod *apiv1.Pod, nodeName string) error
}
