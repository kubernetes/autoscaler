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
	"k8s.io/autoscaler/cluster-autoscaler/simulator/predicatechecker"

	apiv1 "k8s.io/api/core/v1"
	klog "k8s.io/klog/v2"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
)

// HintingSimulator is a helper object for simulating scheduler behavior.
// TODO(x13n): Reuse this in filter_out_unschedulable.
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
// Returns a list of nodes that were chosen or an error if scheduling any of the pods failed.
// Note: this function does not fork clusterSnapshot: this has to be done by the caller.
func (s *HintingSimulator) TrySchedulePods(clusterSnapshot clustersnapshot.ClusterSnapshot, pods []*apiv1.Pod, isNodeAcceptable func(string) bool) ([]string, error) {
	var nodeNames []string
	for _, pod := range pods {
		klog.V(5).Infof("Looking for place for %s/%s", pod.Namespace, pod.Name)
		nodeName, err := s.findNodeWithHints(clusterSnapshot, pod, isNodeAcceptable)
		if err != nil {
			return nil, err
		}
		if nodeName == "" {
			nodeName, err = s.schedulePod(clusterSnapshot, pod, isNodeAcceptable)
			if err != nil {
				return nil, err
			}
		}
		nodeNames = append(nodeNames, nodeName)
	}
	return nodeNames, nil
}

func (s *HintingSimulator) findNodeWithHints(clusterSnapshot clustersnapshot.ClusterSnapshot, pod *apiv1.Pod, isNodeAcceptable func(string) bool) (string, error) {
	hk := HintKeyFromPod(pod)
	if hintedNode, hasHint := s.hints.Get(hk); hasHint && isNodeAcceptable(hintedNode) {
		if err := s.predicateChecker.CheckPredicates(clusterSnapshot, pod, hintedNode); err == nil {
			klog.V(4).Infof("Pod %s/%s can be moved to %s", pod.Namespace, pod.Name, hintedNode)
			if err := clusterSnapshot.AddPod(pod, hintedNode); err != nil {
				return "", fmt.Errorf("Simulating scheduling of %s/%s to %s return error; %v", pod.Namespace, pod.Name, hintedNode, err)
			}
			s.hints.Set(hk, hintedNode)
			return hintedNode, nil
		}
	}
	return "", nil
}

func (s *HintingSimulator) schedulePod(clusterSnapshot clustersnapshot.ClusterSnapshot, pod *apiv1.Pod, isNodeAcceptable func(string) bool) (string, error) {
	newNodeName, err := s.predicateChecker.FitsAnyNodeMatching(clusterSnapshot, pod, func(nodeInfo *schedulerframework.NodeInfo) bool {
		return isNodeAcceptable(nodeInfo.Node().Name)
	})
	if err != nil {
		return "", fmt.Errorf("failed to find place for %s/%s", pod.Namespace, pod.Name)
	}
	klog.V(4).Infof("Pod %s/%s can be moved to %s", pod.Namespace, pod.Name, newNodeName)
	if err := clusterSnapshot.AddPod(pod, newNodeName); err != nil {
		return "", fmt.Errorf("Simulating scheduling of %s/%s to %s return error; %v", pod.Namespace, pod.Name, newNodeName, err)
	}
	s.hints.Set(HintKeyFromPod(pod), newNodeName)
	return newNodeName, nil
}

// DropOldHints drops old scheduling hints.
func (s *HintingSimulator) DropOldHints() {
	s.hints.DropOld()
}

// ScheduleAnywhere can be passed to TrySchedulePods when there are no extra restrictions on nodes to consider.
func ScheduleAnywhere(_ string) bool {
	return true
}
