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

package scheduling

import (
	"context"
	"errors"
	"sort"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/utils/klogx"
	corev1helpers "k8s.io/component-helpers/scheduling/corev1"

	apiv1 "k8s.io/api/core/v1"
)

const (
	// SimulationRequestorName is the key for simulation requestor.
	SimulationRequestorName = "requestor"
)

// Status contains information about pods scheduled by the HintingSimulator
type Status struct {
	Pod      *apiv1.Pod
	NodeName string
}

// HintingSimulator is a helper object for simulating scheduler behavior.
type HintingSimulator struct {
	hints *Hints
}

// NewHintingSimulator returns a new HintingSimulator.
func NewHintingSimulator() *HintingSimulator {
	return &HintingSimulator{
		hints: NewHints(),
	}
}

// Result contains the results after scheduling simulations.
type Result struct {
	Statuses                   []Status
	UnprocessedPods            []*apiv1.Pod
	OverflowingControllerCount int
}

// TrySchedulePods attempts to schedule provided pods on any acceptable nodes.
// Each node is considered acceptable iff isNodeAcceptable() returns true.
// Returns a Result object consisting list of scheduled pods with assigned pods,
// a list of unprocessed pods from simulations due to exceeding the context deadline
// and the count of overflowing controllers, or an error if an unexpected error occurs.
// If the breakOnFailure is set to true, the function will stop scheduling attempts
// after the first scheduling attempt that fails. This is useful if all provided
// pods need to be scheduled.
// Note: this function does not fork clusterSnapshot: this has to be done by the caller.
// Simulations may stop early if the context deadline has passed.
func (s *HintingSimulator) TrySchedulePods(ctx context.Context, clusterSnapshot clustersnapshot.ClusterSnapshot, pods []*apiv1.Pod, breakOnFailure bool, opts clustersnapshot.SchedulingOptions) (Result, error) {
	similarPods := NewSimilarPodsScheduling()

	var statuses []Status
	loggingQuota := klogx.PodsLoggingQuota()
	unschedulablePods := make(map[types.UID]bool)

	for _, podGroup := range groupByPriority(pods) {
		podsWithoutHints, groupStatuses, err := s.tryScheduleWithHints(ctx, clusterSnapshot, podGroup, opts, loggingQuota)
		statuses = append(statuses, groupStatuses...)
		if err != nil {
			return handleContextError(ctx, pods, unschedulablePods, statuses, similarPods, loggingQuota, err)
		}

		unschedulablePods, groupStatuses, err = s.tryScheduleWithoutHints(ctx, clusterSnapshot, podsWithoutHints, unschedulablePods, similarPods, breakOnFailure, opts, loggingQuota)
		statuses = append(statuses, groupStatuses...)
		if err != nil {
			return handleContextError(ctx, pods, unschedulablePods, statuses, similarPods, loggingQuota, err)
		}
	}

	klogx.V(4).Over(loggingQuota).Infof("There were also %v other logs from HintingSimulator.TrySchedulePods func that were capped.", -loggingQuota.Left())
	return newResult(pods, unschedulablePods, statuses, similarPods), nil
}

func (s *HintingSimulator) tryScheduleWithHints(ctx context.Context, clusterSnapshot clustersnapshot.ClusterSnapshot, pods []*apiv1.Pod, opts clustersnapshot.SchedulingOptions, loggingQuota *klogx.Quota) ([]*apiv1.Pod, []Status, error) {
	var podsWithoutHints []*apiv1.Pod
	var statuses []Status
	for _, pod := range pods {
		if err := ctx.Err(); err != nil {
			return podsWithoutHints, statuses, err
		}

		klogx.V(5).UpTo(loggingQuota).Infof("Looking for place for %s/%s", pod.Namespace, pod.Name)
		nodeName, err := s.tryScheduleUsingHints(clusterSnapshot, pod, opts.IsNodeAcceptable)
		if err != nil {
			return podsWithoutHints, statuses, err
		}

		if nodeName == "" {
			podsWithoutHints = append(podsWithoutHints, pod)
			continue
		}

		klogx.V(4).UpTo(loggingQuota).Infof("Pod %s/%s can be moved to %s", pod.Namespace, pod.Name, nodeName)
		statuses = append(statuses, Status{Pod: pod, NodeName: nodeName})
	}
	return podsWithoutHints, statuses, nil
}

func (s *HintingSimulator) tryScheduleWithoutHints(ctx context.Context, clusterSnapshot clustersnapshot.ClusterSnapshot, pods []*apiv1.Pod, unschedulablePods map[types.UID]bool, similarPods *SimilarPodsScheduling, breakOnFailure bool, opts clustersnapshot.SchedulingOptions, loggingQuota *klogx.Quota) (map[types.UID]bool, []Status, error) {
	var statuses []Status
	for _, pod := range pods {
		if err := ctx.Err(); err != nil {
			return unschedulablePods, statuses, err
		}

		nodeName, err := s.trySchedule(similarPods, clusterSnapshot, pod, loggingQuota, opts)
		if err != nil {
			return unschedulablePods, statuses, err
		}

		if nodeName != "" {
			klogx.V(4).UpTo(loggingQuota).Infof("Pod %s/%s can be moved to %s", pod.Namespace, pod.Name, nodeName)
			statuses = append(statuses, Status{Pod: pod, NodeName: nodeName})
			continue
		}

		if breakOnFailure {
			unschedulablePods[pod.UID] = true
			klogx.V(4).Over(loggingQuota).Infof("There were also %v other logs from HintingSimulator.TrySchedulePods func that were capped.", -loggingQuota.Left())
			return unschedulablePods, statuses, nil
		}

		unschedulablePods[pod.UID] = true
	}
	return unschedulablePods, statuses, nil
}

func handleContextError(ctx context.Context, pods []*apiv1.Pod, unschedulablePods map[types.UID]bool, statuses []Status, similarPods *SimilarPodsScheduling, loggingQuota *klogx.Quota, err error) (Result, error) {
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		requestor, found := ctx.Value(SimulationRequestorName).(string)
		if !found {
			klogx.V(4).Infof("Scheduling simulation requestor not found from context.")
			requestor = "NOT FOUND"
		}

		unprocessedPodsCount := len(pods) - len(statuses) - len(unschedulablePods)
		klogx.V(4).Infof("Scheduling simulation aborted early due to %v. Requestor: %s, Simulated %d pods. %d pods were unprocessed", err, requestor, len(statuses), unprocessedPodsCount)
		klogx.V(4).Over(loggingQuota).Infof("There were also %v other logs from HintingSimulator.TrySchedulePods func that were capped.", -loggingQuota.Left())

		return newResult(pods, unschedulablePods, statuses, similarPods), nil
	}
	return Result{}, err
}

// tryScheduleUsingHints tries to schedule the provided Pod in the provided clusterSnapshot using hints. If the pod is scheduled, the name of its Node is returned. If the
// pod couldn't be scheduled using hints, an empty string and nil error is returned. Error is only returned for unexpected errors.
func (s *HintingSimulator) tryScheduleUsingHints(clusterSnapshot clustersnapshot.ClusterSnapshot, pod *apiv1.Pod, isNodeAcceptable func(*framework.NodeInfo) bool) (string, error) {
	hk := HintKeyFromPod(pod)
	hintedNode, hasHint := s.hints.Get(hk)
	if !hasHint {
		return "", nil
	}

	nodeInfo, err := clusterSnapshot.GetNodeInfo(hintedNode)
	if err != nil {
		// The hinted Node is no longer in the cluster. No need to error out, we can just look for another one.
		return "", nil
	}
	if isNodeAcceptable != nil && !isNodeAcceptable(nodeInfo) {
		return "", nil
	}

	if err := clusterSnapshot.SchedulePod(pod, hintedNode); err != nil && err.Type() == clustersnapshot.SchedulingInternalError {
		// Unexpected error.
		return "", err
	} else if err != nil {
		// The pod can't be scheduled on the hintedNode because of scheduling predicates.
		return "", nil
	}
	// The pod was scheduled on hintedNode.
	s.hints.Set(hk, hintedNode)
	return hintedNode, nil
}

// trySchedule tries to schedule the provided Pod in the provided clusterSnapshot on any Node passing isNodeAcceptable. If the pod is scheduled, the name of its Node is returned. If no Node
// with passing scheduling predicates could be found, an empty string and nil error is returned. Error is only returned for unexpected errors.
func (s *HintingSimulator) trySchedule(similarPods *SimilarPodsScheduling, clusterSnapshot clustersnapshot.ClusterSnapshot, pod *apiv1.Pod, loggingQuota *klogx.Quota, opts clustersnapshot.SchedulingOptions) (string, error) {
	if similarPods.IsSimilarUnschedulable(pod) {
		klogx.V(4).UpTo(loggingQuota).Infof("failed to find place for %s/%s based on similar pods scheduling", pod.Namespace, pod.Name)
		return "", nil
	}

	newNodeName, err := clusterSnapshot.SchedulePodOnAnyNodeMatching(pod, opts)
	if err != nil && err.Type() == clustersnapshot.SchedulingInternalError {
		// Unexpected error.
		return "", err
	} else if err != nil {
		// The pod couldn't be scheduled on any Node because of scheduling predicates.
		klogx.V(4).UpTo(loggingQuota).Infof("failed to find place for %s/%s: %v", pod.Namespace, pod.Name, err)
		similarPods.SetUnschedulable(pod)
		return "", nil
	}
	// The pod was scheduled on newNodeName.
	s.hints.Set(HintKeyFromPod(pod), newNodeName)
	return newNodeName, nil
}

// DropOldHints drops old scheduling hints.
func (s *HintingSimulator) DropOldHints() {
	s.hints.DropOld()
}

// ScheduleAnywhere can be passed to TrySchedulePods when there are no extra restrictions on nodes to consider.
func ScheduleAnywhere(_ *framework.NodeInfo) bool {
	return true
}

// groupByPriority groups pods by their priority and returns a slice of pod slices, ordered from highest to lowest priority.
func groupByPriority(pods []*apiv1.Pod) [][]*apiv1.Pod {
	podGroups := make(map[int32][]*apiv1.Pod)
	for _, pod := range pods {
		priority := corev1helpers.PodPriority(pod)
		podGroups[priority] = append(podGroups[priority], pod)
	}

	var priorities []int
	for p := range podGroups {
		priorities = append(priorities, int(p))
	}

	sort.Slice(priorities, func(i, j int) bool {
		return priorities[i] > priorities[j]
	})

	var result [][]*apiv1.Pod
	for _, p := range priorities {
		result = append(result, podGroups[int32(p)])
	}
	return result
}

func newResult(allPods []*apiv1.Pod, unschedulablePodMap map[types.UID]bool, statuses []Status, similarPods *SimilarPodsScheduling) Result {
	processedPodMap := make(map[types.UID]bool)
	var unprocessedPods []*apiv1.Pod
	for _, status := range statuses {
		processedPodMap[status.Pod.UID] = true
	}

	for _, pod := range allPods {
		if !processedPodMap[pod.UID] && !unschedulablePodMap[pod.UID] {
			unprocessedPods = append(unprocessedPods, pod)
		}
	}
	return Result{statuses, unprocessedPods, similarPods.OverflowingControllerCount()}
}
