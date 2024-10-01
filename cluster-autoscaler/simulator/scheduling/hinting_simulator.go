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
	"fmt"

	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/predicatechecker"
	"k8s.io/autoscaler/cluster-autoscaler/utils/klogx"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"

	apiv1 "k8s.io/api/core/v1"
)

// Status contains information about pods scheduled by the HintingSimulator
type Status struct {
	Pod      *apiv1.Pod
	NodeName string
}

// HintingSimulator is a helper object for simulating scheduler behavior.
type HintingSimulator struct {
	predicateChecker predicatechecker.PredicateChecker
	hints            *Hints
}

// NewHintingSimulator returns a new HintingSimulator.
func NewHintingSimulator(predicateChecker predicatechecker.PredicateChecker) *HintingSimulator {
	return &HintingSimulator{
		predicateChecker: predicateChecker,
		hints:            NewHints(),
	}
}

// TrySchedulePods attempts to schedule provided pods on any acceptable nodes.
// Each node is considered acceptable iff isNodeAcceptable() returns true.
// Returns a list of scheduled pods with assigned pods and the count of overflowing
// controllers, or an error if an unexpected error occurs.
// If the breakOnFailure is set to true, the function will stop scheduling attempts
// after the first scheduling attempt that fails. This is useful if all provided
// pods need to be scheduled.
// Note: this function does not fork clusterSnapshot: this has to be done by the caller.
func (s *HintingSimulator) TrySchedulePods(clusterSnapshot clustersnapshot.ClusterSnapshot, pods []*apiv1.Pod, isNodeAcceptable func(*framework.NodeInfo) bool, breakOnFailure bool) ([]Status, int, error) {
	similarPods := NewSimilarPodsScheduling()

	var statuses []Status
	loggingQuota := klogx.PodsLoggingQuota()
	for _, pod := range pods {
		klogx.V(5).UpTo(loggingQuota).Infof("Looking for place for %s/%s", pod.Namespace, pod.Name)
		nodeName, schedulingState, err := s.findNodeWithHints(clusterSnapshot, pod, isNodeAcceptable)
		if err != nil {
			return nil, 0, err
		}

		if nodeName == "" {
			nodeName, schedulingState = s.findNode(similarPods, clusterSnapshot, pod, loggingQuota, isNodeAcceptable)
		}

		if nodeName != "" {
			klogx.V(4).UpTo(loggingQuota).Infof("Pod %s/%s can be moved to %s", pod.Namespace, pod.Name, nodeName)
			if err := clusterSnapshot.SchedulePod(pod, nodeName, schedulingState); err != nil {
				return nil, 0, fmt.Errorf("simulating scheduling of %s/%s to %s return error; %v", pod.Namespace, pod.Name, nodeName, err)
			}
			statuses = append(statuses, Status{Pod: pod, NodeName: nodeName})
		} else if breakOnFailure {
			break
		}
	}
	klogx.V(4).Over(loggingQuota).Infof("There were also %v other logs from HintingSimulator.TrySchedulePods func that were capped.", -loggingQuota.Left())
	return statuses, similarPods.OverflowingControllerCount(), nil
}

func (s *HintingSimulator) findNodeWithHints(clusterSnapshot clustersnapshot.ClusterSnapshot, pod *apiv1.Pod, isNodeAcceptable func(*framework.NodeInfo) bool) (string, *schedulerframework.CycleState, error) {
	hk := HintKeyFromPod(pod)
	if hintedNode, hasHint := s.hints.Get(hk); hasHint {
		if schedulingState, err := s.predicateChecker.CheckPredicates(clusterSnapshot, pod, hintedNode); err == nil {
			s.hints.Set(hk, hintedNode)

			nodeInfo, err := clusterSnapshot.GetNodeInfo(hintedNode)
			if err != nil {
				return "", nil, err
			}

			if isNodeAcceptable(nodeInfo) {
				return hintedNode, schedulingState, nil
			}
		}
	}
	return "", nil, nil
}

func (s *HintingSimulator) findNode(similarPods *SimilarPodsScheduling, clusterSnapshot clustersnapshot.ClusterSnapshot, pod *apiv1.Pod, loggingQuota *klogx.Quota, isNodeAcceptable func(*framework.NodeInfo) bool) (string, *schedulerframework.CycleState) {
	if similarPods.IsSimilarUnschedulable(pod) {
		klogx.V(4).UpTo(loggingQuota).Infof("failed to find place for %s/%s based on similar pods scheduling", pod.Namespace, pod.Name)
		return "", nil
	}

	newNodeName, schedulingState, err := s.predicateChecker.FitsAnyNodeMatching(clusterSnapshot, pod, isNodeAcceptable)
	if err != nil {
		klogx.V(4).UpTo(loggingQuota).Infof("failed to find place for %s/%s: %v", pod.Namespace, pod.Name, err)
		similarPods.SetUnschedulable(pod)
		return "", nil
	}

	s.hints.Set(HintKeyFromPod(pod), newNodeName)
	return newNodeName, schedulingState
}

// DropOldHints drops old scheduling hints.
func (s *HintingSimulator) DropOldHints() {
	s.hints.DropOld()
}

// ScheduleAnywhere can be passed to TrySchedulePods when there are no extra restrictions on nodes to consider.
func ScheduleAnywhere(_ *framework.NodeInfo) bool {
	return true
}
