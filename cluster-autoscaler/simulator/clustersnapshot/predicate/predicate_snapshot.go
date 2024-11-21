/*
Copyright 2024 The Kubernetes Authors.

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

package predicate

import (
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
)

// PredicateSnapshot implements ClusterSnapshot on top of a ClusterSnapshotStore by using
// SchedulerBasedPredicateChecker to check scheduler predicates.
type PredicateSnapshot struct {
	clustersnapshot.ClusterSnapshotStore
	pluginRunner *SchedulerPluginRunner
	draEnabled   bool
}

// NewPredicateSnapshot builds a PredicateSnapshot.
func NewPredicateSnapshot(snapshotStore clustersnapshot.ClusterSnapshotStore, fwHandle *framework.Handle, draEnabled bool) *PredicateSnapshot {
	return &PredicateSnapshot{
		ClusterSnapshotStore: snapshotStore,
		pluginRunner:         NewSchedulerPluginRunner(fwHandle, snapshotStore),
		draEnabled:           draEnabled,
	}
}

// SchedulePod adds pod to the snapshot and schedules it to given node.
func (s *PredicateSnapshot) SchedulePod(pod *apiv1.Pod, nodeName string) clustersnapshot.SchedulingError {
	if schedErr := s.pluginRunner.RunFiltersOnNode(pod, nodeName); schedErr != nil {
		return schedErr
	}
	if err := s.ClusterSnapshotStore.ForceAddPod(pod, nodeName); err != nil {
		return clustersnapshot.NewSchedulingInternalError(pod, err.Error())
	}
	return nil
}

// SchedulePodOnAnyNodeMatching adds pod to the snapshot and schedules it to any node matching the provided function.
func (s *PredicateSnapshot) SchedulePodOnAnyNodeMatching(pod *apiv1.Pod, anyNodeMatching func(*framework.NodeInfo) bool) (string, clustersnapshot.SchedulingError) {
	nodeName, schedErr := s.pluginRunner.RunFiltersUntilPassingNode(pod, anyNodeMatching)
	if schedErr != nil {
		return "", schedErr
	}
	if err := s.ClusterSnapshotStore.ForceAddPod(pod, nodeName); err != nil {
		return "", clustersnapshot.NewSchedulingInternalError(pod, err.Error())
	}
	return nodeName, nil
}

// UnschedulePod removes the given Pod from the given Node inside the snapshot.
func (s *PredicateSnapshot) UnschedulePod(namespace string, podName string, nodeName string) error {
	return s.ClusterSnapshotStore.ForceRemovePod(namespace, podName, nodeName)
}

// CheckPredicates checks whether scheduler predicates pass for the given pod on the given node.
func (s *PredicateSnapshot) CheckPredicates(pod *apiv1.Pod, nodeName string) clustersnapshot.SchedulingError {
	return s.pluginRunner.RunFiltersOnNode(pod, nodeName)
}
