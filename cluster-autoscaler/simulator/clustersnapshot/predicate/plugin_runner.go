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
	"context"
	"fmt"
	"strings"
	"sync"

	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/client-go/util/workqueue"

	apiv1 "k8s.io/api/core/v1"
	schedulerimpl "k8s.io/kubernetes/pkg/scheduler/framework"
)

// SchedulerPluginRunner can be used to run various phases of scheduler plugins through the scheduler framework.
type SchedulerPluginRunner struct {
	fwHandle    *framework.Handle
	snapshot    clustersnapshot.ClusterSnapshot
	lastIndex   int
	parallelism int
}

// NewSchedulerPluginRunner builds a SchedulerPluginRunner.
func NewSchedulerPluginRunner(fwHandle *framework.Handle, snapshot clustersnapshot.ClusterSnapshot, parallelism int) *SchedulerPluginRunner {
	return &SchedulerPluginRunner{fwHandle: fwHandle, snapshot: snapshot, parallelism: parallelism}
}

// RunFiltersUntilPassingNode runs the scheduler framework PreFilter phase once, and then keeps running the Filter phase for all nodes in the cluster that match the provided
// function - until a Node where the Filters pass is found. Filters are only run for matching Nodes. If no matching Node with passing Filters is found, an error is returned.
//
// The node iteration always starts from the next Node from the last Node that was found by this method. TODO: Extract the iteration strategy out of SchedulerPluginRunner.
func (p *SchedulerPluginRunner) RunFiltersUntilPassingNode(pod *apiv1.Pod, nodeMatches func(*framework.NodeInfo) bool) (*apiv1.Node, *schedulerimpl.CycleState, clustersnapshot.SchedulingError) {
	nodeInfosList, err := p.snapshot.ListNodeInfos()
	if err != nil {
		return nil, nil, clustersnapshot.NewSchedulingInternalError(pod, fmt.Sprintf("error listing NodeInfos: %v", err))
	}

	p.fwHandle.DelegatingLister.UpdateDelegate(p.snapshot)
	defer p.fwHandle.DelegatingLister.ResetDelegate()

	state := schedulerimpl.NewCycleState()
	// Run the PreFilter phase of the framework for the Pod. This allows plugins to precompute some things (for all Nodes in the cluster at once) and
	// save them in the CycleState. During the Filter phase, plugins can retrieve the precomputes from the CycleState and use them for answering the Filter
	// for a given Node.
	preFilterResult, preFilterStatus, _ := p.fwHandle.Framework.RunPreFilterPlugins(context.TODO(), state, pod)
	if !preFilterStatus.IsSuccess() {
		// If any of the plugin PreFilter methods isn't successful, the corresponding Filter method can't be run, so the whole scheduling cycle is aborted.
		// Match that behavior here.
		return nil, nil, clustersnapshot.NewFailingPredicateError(pod, preFilterStatus.Plugin(), preFilterStatus.Reasons(), "PreFilter failed", "")
	}

	var (
		foundNode  *apiv1.Node
		foundIndex int
		mu         sync.Mutex
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	checkNode := func(i int) {
		nodeIndex := (p.lastIndex + i) % len(nodeInfosList)
		nodeInfo := nodeInfosList[nodeIndex]

		// Plugins can filter some Nodes out during the PreFilter phase, if they're sure the Nodes won't work for the Pod at that stage.
		// Filters are only run for Nodes that haven't been filtered out during the PreFilter phase. Match that behavior here - skip such Nodes.
		if !preFilterResult.AllNodes() && !preFilterResult.NodeNames.Has(nodeInfo.Node().Name) {
			return
		}

		// Nodes with the Unschedulable bit set will be rejected by one of the plugins during the Filter phase below. We can check that quickly here
		// and short-circuit to avoid running the expensive Filter phase at all in this case.
		if nodeInfo.Node().Spec.Unschedulable {
			return
		}

		// Check if the NodeInfo matches the provided filtering condition. This should be less expensive than running the Filter phase below, so
		// check this first.
		if !nodeMatches(nodeInfo) {
			return
		}

		// Run the Filter phase of the framework. Plugins retrieve the state they saved during PreFilter from CycleState, and answer whether the
		// given Pod can be scheduled on the given Node.
		filterStatus := p.fwHandle.Framework.RunFilterPlugins(context.TODO(), state, pod, nodeInfo.ToScheduler())
		if filterStatus.IsSuccess() {
			// Filter passed for all plugins, so this pod can be scheduled on this Node.
			mu.Lock()
			defer mu.Unlock()
			if foundNode == nil {
				foundNode = nodeInfo.Node()
				foundIndex = nodeIndex
				cancel()
			}
		}
		// Filter didn't pass for some plugin, so this Node won't work - move on to the next one.
	}

	workqueue.ParallelizeUntil(ctx, p.parallelism, len(nodeInfosList), checkNode)

	if foundNode != nil {
		p.lastIndex = (foundIndex + 1) % len(nodeInfosList)
		return foundNode, state, nil
	}

	return nil, nil, clustersnapshot.NewNoNodesPassingPredicatesFoundError(pod)
}

// RunFiltersOnNode runs the scheduler framework PreFilter and Filter phases to check if the given pod can be scheduled on the given node.
func (p *SchedulerPluginRunner) RunFiltersOnNode(pod *apiv1.Pod, nodeName string) (*apiv1.Node, *schedulerimpl.CycleState, clustersnapshot.SchedulingError) {
	nodeInfo, err := p.snapshot.GetNodeInfo(nodeName)
	if err != nil {
		return nil, nil, clustersnapshot.NewSchedulingInternalError(pod, fmt.Sprintf("error obtaining NodeInfo for name %q: %v", nodeName, err))
	}

	p.fwHandle.DelegatingLister.UpdateDelegate(p.snapshot)
	defer p.fwHandle.DelegatingLister.ResetDelegate()

	state := schedulerimpl.NewCycleState()
	// Run the PreFilter phase of the framework for the Pod and check the results. See the corresponding comments in RunFiltersUntilPassingNode() for more info.
	preFilterResult, preFilterStatus, nodeFilteringPlugins := p.fwHandle.Framework.RunPreFilterPlugins(context.TODO(), state, pod)
	if !preFilterStatus.IsSuccess() {
		// nil check on preFilterStatus not required, as IsSuccess returns true for nil
		return nil, nil, clustersnapshot.NewFailingPredicateError(pod, preFilterStatus.Plugin(), preFilterStatus.Reasons(), "PreFilter failed", "")
	}
	if !preFilterResult.AllNodes() && !preFilterResult.NodeNames.Has(nodeInfo.Node().Name) {
		// in this scope, preFilterStatus is most likely nil
		return nil, nil, clustersnapshot.NewFailingPredicateError(pod, strings.Join(nodeFilteringPlugins.UnsortedList(), ", "), nil, "PreFilter filtered the Node out", "")
	}

	// Run the Filter phase of the framework for the Pod and the Node and check the results. See the corresponding comments in RunFiltersUntilPassingNode() for more info.
	filterStatus := p.fwHandle.Framework.RunFilterPlugins(context.TODO(), state, pod, nodeInfo.ToScheduler())
	if !filterStatus.IsSuccess() {
		filterName := filterStatus.Plugin()
		filterReasons := filterStatus.Reasons()
		unexpectedErrMsg := ""
		if !filterStatus.IsRejected() {
			unexpectedErrMsg = fmt.Sprintf("unexpected filter status %q", filterStatus.Code().String())
		}
		return nil, nil, clustersnapshot.NewFailingPredicateError(pod, filterName, filterReasons, unexpectedErrMsg, p.failingFilterDebugInfo(filterName, nodeInfo))
	}

	// PreFilter and Filter phases checked, this Pod can be scheduled on this Node.
	return nodeInfo.Node(), state, nil
}

// RunReserveOnNode runs the scheduler framework Reserve phase to update the scheduler plugins state to reflect the Pod being scheduled on the Node.
func (p *SchedulerPluginRunner) RunReserveOnNode(pod *apiv1.Pod, nodeName string, postFilterState *schedulerimpl.CycleState) error {
	p.fwHandle.DelegatingLister.UpdateDelegate(p.snapshot)
	defer p.fwHandle.DelegatingLister.ResetDelegate()

	status := p.fwHandle.Framework.RunReservePluginsReserve(context.Background(), postFilterState, pod, nodeName)
	if !status.IsSuccess() {
		return fmt.Errorf("couldn't reserve node %s for pod %s/%s: %v", nodeName, pod.Namespace, pod.Name, status.Message())
	}
	return nil
}

func (p *SchedulerPluginRunner) failingFilterDebugInfo(filterName string, nodeInfo *framework.NodeInfo) string {
	infoParts := []string{fmt.Sprintf("nodeName: %q", nodeInfo.Node().Name)}

	switch filterName {
	case "TaintToleration":
		infoParts = append(infoParts, fmt.Sprintf("nodeTaints: %#v", nodeInfo.Node().Spec.Taints))
	}

	return strings.Join(infoParts, ", ")
}
