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
	"strings"

	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"

	apiv1 "k8s.io/api/core/v1"
	v1listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/klog/v2"
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
func (p *SchedulerBasedPredicateChecker) FitsAnyNode(clusterSnapshot clustersnapshot.ClusterSnapshotStore, pod *apiv1.Pod) (string, clustersnapshot.SchedulingError) {
	return p.FitsAnyNodeMatching(clusterSnapshot, pod, func(*framework.NodeInfo) bool {
		return true
	})
}

// FitsAnyNodeMatching checks if the given pod can be placed on any of the given nodes matching the provided function.
func (p *SchedulerBasedPredicateChecker) FitsAnyNodeMatching(clusterSnapshot clustersnapshot.ClusterSnapshotStore, pod *apiv1.Pod, nodeMatches func(*framework.NodeInfo) bool) (string, clustersnapshot.SchedulingError) {
	if clusterSnapshot == nil {
		return "", clustersnapshot.NewSchedulingInternalError(pod, "ClusterSnapshot not provided")
	}

	nodeInfosList, err := clusterSnapshot.ListNodeInfos()
	if err != nil {
		// This should never happen.
		//
		// Scheduler requires interface returning error, but no implementation
		// of ClusterSnapshot ever does it.
		klog.Errorf("Error obtaining nodeInfos from schedulerLister")
		return "", clustersnapshot.NewSchedulingInternalError(pod, "error obtaining nodeInfos from schedulerLister")
	}

	p.fwHandle.DelegatingLister.UpdateDelegate(clusterSnapshot)
	defer p.fwHandle.DelegatingLister.ResetDelegate()

	state := schedulerframework.NewCycleState()
	preFilterResult, preFilterStatus, _ := p.fwHandle.Framework.RunPreFilterPlugins(context.TODO(), state, pod)
	if !preFilterStatus.IsSuccess() {
		return "", clustersnapshot.NewFailingPredicateError(pod, preFilterStatus.Plugin(), preFilterStatus.Reasons(), "PreFilter failed", "")
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
	return "", clustersnapshot.NewNoNodesPassingPredicatesFoundError(pod)
}

// CheckPredicates checks if the given pod can be placed on the given node.
func (p *SchedulerBasedPredicateChecker) CheckPredicates(clusterSnapshot clustersnapshot.ClusterSnapshotStore, pod *apiv1.Pod, nodeName string) clustersnapshot.SchedulingError {
	if clusterSnapshot == nil {
		return clustersnapshot.NewSchedulingInternalError(pod, "ClusterSnapshot not provided")
	}
	nodeInfo, err := clusterSnapshot.GetNodeInfo(nodeName)
	if err != nil {
		return clustersnapshot.NewSchedulingInternalError(pod, fmt.Sprintf("error obtaining NodeInfo for name %q: %v", nodeName, err))
	}

	p.fwHandle.DelegatingLister.UpdateDelegate(clusterSnapshot)
	defer p.fwHandle.DelegatingLister.ResetDelegate()

	state := schedulerframework.NewCycleState()
	_, preFilterStatus, _ := p.fwHandle.Framework.RunPreFilterPlugins(context.TODO(), state, pod)
	if !preFilterStatus.IsSuccess() {
		return clustersnapshot.NewFailingPredicateError(pod, preFilterStatus.Plugin(), preFilterStatus.Reasons(), "PreFilter failed", "")
	}

	filterStatus := p.fwHandle.Framework.RunFilterPlugins(context.TODO(), state, pod, nodeInfo.ToScheduler())

	if !filterStatus.IsSuccess() {
		filterName := filterStatus.Plugin()
		filterReasons := filterStatus.Reasons()
		unexpectedErrMsg := ""
		if !filterStatus.IsRejected() {
			unexpectedErrMsg = fmt.Sprintf("unexpected filter status %q", filterStatus.Code().String())
		}
		return clustersnapshot.NewFailingPredicateError(pod, filterName, filterReasons, unexpectedErrMsg, p.failingFilterDebugInfo(filterName, nodeInfo))
	}

	return nil
}

func (p *SchedulerBasedPredicateChecker) failingFilterDebugInfo(filterName string, nodeInfo *framework.NodeInfo) string {
	infoParts := []string{fmt.Sprintf("nodeName: %q", nodeInfo.Node().Name)}

	switch filterName {
	case "TaintToleration":
		infoParts = append(infoParts, fmt.Sprintf("nodeTaints: %#v", nodeInfo.Node().Spec.Taints))
	}

	return strings.Join(infoParts, ", ")
}
