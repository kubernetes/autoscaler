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

package predicatechecker

import (
	"context"
	"fmt"

	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"

	apiv1 "k8s.io/api/core/v1"
	v1listers "k8s.io/client-go/listers/core/v1"
	klog "k8s.io/klog/v2"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
)

// SchedulerBasedPredicateChecker checks whether all required predicates pass for given Pod and Node.
// The verification is done by calling out to scheduler code.
type SchedulerBasedPredicateChecker struct {
	fwHandle   *framework.Handle
	nodeLister v1listers.NodeLister
	podLister  v1listers.PodLister
	lastIndex  int
}

// NewSchedulerBasedPredicateChecker builds scheduler based PredicateChecker.
func NewSchedulerBasedPredicateChecker(fwHandle *framework.Handle) *SchedulerBasedPredicateChecker {
	return &SchedulerBasedPredicateChecker{fwHandle: fwHandle}
}

// FitsAnyNode checks if the given pod can be placed on any of the given nodes.
func (p *SchedulerBasedPredicateChecker) FitsAnyNode(clusterSnapshot clustersnapshot.ClusterSnapshot, pod *apiv1.Pod) (string, error) {
	return p.FitsAnyNodeMatching(clusterSnapshot, pod, func(*framework.NodeInfo) bool {
		return true
	})
}

// FitsAnyNodeMatching checks if the given pod can be placed on any of the given nodes matching the provided function.
func (p *SchedulerBasedPredicateChecker) FitsAnyNodeMatching(clusterSnapshot clustersnapshot.ClusterSnapshot, pod *apiv1.Pod, nodeMatches func(*framework.NodeInfo) bool) (string, error) {
	if clusterSnapshot == nil {
		return "", fmt.Errorf("ClusterSnapshot not provided")
	}

	nodeInfosList, err := clusterSnapshot.ListNodeInfos()
	if err != nil {
		// This should never happen.
		//
		// Scheduler requires interface returning error, but no implementation
		// of ClusterSnapshot ever does it.
		klog.Errorf("Error obtaining nodeInfos from schedulerLister")
		return "", fmt.Errorf("error obtaining nodeInfos from schedulerLister")
	}

	p.fwHandle.DelegatingLister.UpdateDelegate(clusterSnapshot)
	defer p.fwHandle.DelegatingLister.ResetDelegate()

	state := schedulerframework.NewCycleState()
	preFilterResult, preFilterStatus, _ := p.fwHandle.Framework.RunPreFilterPlugins(context.TODO(), state, pod)
	if !preFilterStatus.IsSuccess() {
		return "", fmt.Errorf("error running pre filter plugins for pod %s; %s", pod.Name, preFilterStatus.Message())
	}

	for i := range nodeInfosList {
		nodeInfo := nodeInfosList[(p.lastIndex+i)%len(nodeInfosList)]
		if !nodeMatches(nodeInfo) {
			continue
		}

		if !preFilterResult.AllNodes() && !preFilterResult.NodeNames.Has(nodeInfo.Node().Name) {
			continue
		}

		// Be sure that the node is schedulable.
		if nodeInfo.Node().Spec.Unschedulable {
			continue
		}

		filterStatus := p.fwHandle.Framework.RunFilterPlugins(context.TODO(), state, pod, nodeInfo.ToScheduler())
		if filterStatus.IsSuccess() {
			p.lastIndex = (p.lastIndex + i + 1) % len(nodeInfosList)
			return nodeInfo.Node().Name, nil
		}
	}
	return "", fmt.Errorf("cannot put pod %s on any node", pod.Name)
}

// CheckPredicates checks if the given pod can be placed on the given node.
func (p *SchedulerBasedPredicateChecker) CheckPredicates(clusterSnapshot clustersnapshot.ClusterSnapshot, pod *apiv1.Pod, nodeName string) *PredicateError {
	if clusterSnapshot == nil {
		return NewPredicateError(InternalPredicateError, "", "ClusterSnapshot not provided", nil, emptyString)
	}
	nodeInfo, err := clusterSnapshot.GetNodeInfo(nodeName)
	if err != nil {
		errorMessage := fmt.Sprintf("Error obtaining NodeInfo for name %s; %v", nodeName, err)
		return NewPredicateError(InternalPredicateError, "", errorMessage, nil, emptyString)
	}

	p.fwHandle.DelegatingLister.UpdateDelegate(clusterSnapshot)
	defer p.fwHandle.DelegatingLister.ResetDelegate()

	state := schedulerframework.NewCycleState()
	_, preFilterStatus, _ := p.fwHandle.Framework.RunPreFilterPlugins(context.TODO(), state, pod)
	if !preFilterStatus.IsSuccess() {
		return NewPredicateError(
			InternalPredicateError,
			"",
			preFilterStatus.Message(),
			preFilterStatus.Reasons(),
			emptyString)
	}

	filterStatus := p.fwHandle.Framework.RunFilterPlugins(context.TODO(), state, pod, nodeInfo.ToScheduler())

	if !filterStatus.IsSuccess() {
		filterName := filterStatus.Plugin()
		filterMessage := filterStatus.Message()
		filterReasons := filterStatus.Reasons()
		if filterStatus.IsRejected() {
			return NewPredicateError(
				NotSchedulablePredicateError,
				filterName,
				filterMessage,
				filterReasons,
				p.buildDebugInfo(filterName, nodeInfo))
		}
		return NewPredicateError(
			InternalPredicateError,
			filterName,
			filterMessage,
			filterReasons,
			p.buildDebugInfo(filterName, nodeInfo))
	}

	return nil
}

func (p *SchedulerBasedPredicateChecker) buildDebugInfo(filterName string, nodeInfo *framework.NodeInfo) func() string {
	switch filterName {
	case "TaintToleration":
		taints := nodeInfo.Node().Spec.Taints
		return func() string {
			return fmt.Sprintf("taints on node: %#v", taints)
		}
	default:
		return emptyString
	}
}
