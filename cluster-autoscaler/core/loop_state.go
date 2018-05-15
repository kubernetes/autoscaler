/*
Copyright 2018 The Kubernetes Authors.

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

package core

import (
	"fmt"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/clusterstate/utils"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/metrics"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	"k8s.io/autoscaler/cluster-autoscaler/utils/tpu"

	apiv1 "k8s.io/api/core/v1"

	"github.com/golang/glog"
)

// ExecutionResult describes the status of a call to a LoopStateHandler phase method,
// whether the processing can continue further (next phase can be called) or not.
type ExecutionResult int

const (
	// CanContinue means that the processing can continue further
	CanContinue ExecutionResult = iota
	// ShouldStop means that the processing should be stopped and error (or nil) returned
	ShouldStop
)

// LoopStateHandler hides complexity of Autoscaler.RunOnce implementation,
// splitting it into a number of phases.
// Each phase method returns ExecutionResult to show if processing can continue further
// and a result (error or nil) to return from Autoscaler.RunOnce if the processing should stop.
// NOTE: Do not use `err != nil` idiom to decide if the processing can continue.
// LoopStateHandler instance is a very short-lived one, persisting state of a single RunOnce call.
type LoopStateHandler interface {
	// RefreshState obtains new Node-level data from the cloud provider
	// and updates internal ClusterStateRegistry, until metrics.UpdateState.
	// See LoopStateHandler doc for description of return values.
	RefreshState() (ExecutionResult, errors.AutoscalerError)
	// CheckPreconditions verifies Node-level preconditions for continuing
	// to processing Pod-level data.
	// See LoopStateHandler doc for description of return values.
	CheckPreconditions() (ExecutionResult, errors.AutoscalerError)
	// PreparePodData obtains Pod-level data for further processing.
	// See LoopStateHandler doc for description of return values.
	PreparePodData() (ExecutionResult, errors.AutoscalerError)
	// ScaleUp executes scale-up, if it's required.
	// See LoopStateHandler doc for description of return values.
	ScaleUp() (ExecutionResult, errors.AutoscalerError)
	// ScaleDown executes scale-down logic, both dry-run & actual scale-down, if it's required.
	// See LoopStateHandler doc for description of return values.
	ScaleDown() (ExecutionResult, errors.AutoscalerError)
}

// loopState is the default implementation of LoopStateHandler
type loopState struct {
	currentTime        time.Time
	startTime          time.Time
	scaleDownForbidden bool

	autoscaler *StaticAutoscaler
	context    *context.AutoscalingContext

	allNodes                                       []*apiv1.Node
	readyNodes                                     []*apiv1.Node
	allScheduled                                   []*apiv1.Pod
	unschedulablePodsToHelp                        []*apiv1.Pod
	unschedulableWaitingForLowerPriorityPreemption []*apiv1.Pod
}

// NewLoopStateHandler creates new instance of the default LoopStateHandler implementation
func NewLoopStateHandler(autoscaler *StaticAutoscaler, currentTime time.Time) *loopState {
	return &loopState{
		currentTime: currentTime,
		startTime:   time.Now(),
		autoscaler:  autoscaler,
		context:     autoscaler.AutoscalingContext,
	}
}

// See LoopStateHandler interface
func (ls *loopState) RefreshState() (ExecutionResult, errors.AutoscalerError) {
	if result, err := ls.refreshCloudProvider(); result != CanContinue {
		return result, err
	}
	if result, err := ls.obtainNodeLists(); result != CanContinue {
		return result, err
	}
	return ls.updateClusterStateRegistry()
}

// See LoopStateHandler interface
func (ls *loopState) CheckPreconditions() (ExecutionResult, errors.AutoscalerError) {
	if result, err := ls.cleanUnregisteredNodes(); result != CanContinue {
		return result, err
	}
	if result, err := ls.checkClusterHealth(); result != CanContinue {
		return result, err
	}
	return ls.fixNodeGroupSizes()
}

// See LoopStateHandler interface
func (ls *loopState) PreparePodData() (ExecutionResult, errors.AutoscalerError) {
	allUnschedulablePods, err := ls.autoscaler.UnschedulablePodLister().List()
	if err != nil {
		glog.Errorf("Failed to list unscheduled pods: %v", err)
		return ls.onError(errors.ApiCallError, err)
	}
	metrics.UpdateUnschedulablePodsCount(len(allUnschedulablePods))

	ls.allScheduled, err = ls.autoscaler.ScheduledPodLister().List()
	if err != nil {
		glog.Errorf("Failed to list scheduled pods: %v", err)
		return ls.onError(errors.ApiCallError, err)
	}

	allUnschedulablePods, ls.allScheduled, err = ls.autoscaler.podListProcessor.Process(
		ls.context, allUnschedulablePods, ls.allScheduled, ls.allNodes)
	if err != nil {
		glog.Errorf("Failed to process pod list: %v", err)
		return ls.onError(errors.InternalError, err)
	}

	ConfigurePredicateCheckerForLoop(allUnschedulablePods, ls.allScheduled, ls.autoscaler.PredicateChecker)

	// We need to check whether pods marked as unschedulable are actually unschedulable.
	// It's likely we added a new node and the scheduler just haven't managed to put the
	// pod on in yet. In this situation we don't want to trigger another scale-up.
	//
	// It's also important to prevent uncontrollable cluster growth if CA's simulated
	// scheduler differs in opinion with real scheduler. Example of such situation:
	// - CA and Scheduler has slightly different configuration
	// - Scheduler can't schedule a pod and marks it as unschedulable
	// - CA added a node which should help the pod
	// - Scheduler doesn't schedule the pod on the new node
	//   because according to it logic it doesn't fit there
	// - CA see the pod is still unschedulable, so it adds another node to help it
	//
	// With the check enabled the last point won't happen because CA will ignore a pod
	// which is supposed to schedule on an existing node.
	ls.scaleDownForbidden = false

	unschedulablePodsWithoutTPUs := tpu.ClearTPURequests(allUnschedulablePods)

	// Some unschedulable pods can be waiting for lower priority pods preemption so they have nominated node to run.
	// Such pods don't require scale up but should be considered during scale down.
	var unschedulablePods []*apiv1.Pod
	unschedulablePods, ls.unschedulableWaitingForLowerPriorityPreemption = FilterOutExpendableAndSplit(
		unschedulablePodsWithoutTPUs, ls.autoscaler.ExpendablePodsPriorityCutoff)

	glog.V(4).Infof("Filtering out schedulables")
	filterOutSchedulableStart := time.Now()
	ls.unschedulablePodsToHelp = FilterOutSchedulable(unschedulablePods, ls.readyNodes, ls.allScheduled,
		ls.unschedulableWaitingForLowerPriorityPreemption, ls.autoscaler.PredicateChecker, ls.autoscaler.ExpendablePodsPriorityCutoff)
	metrics.UpdateDurationFromStart(metrics.FilterOutSchedulable, filterOutSchedulableStart)

	if len(ls.unschedulablePodsToHelp) != len(unschedulablePods) {
		glog.V(2).Info("Schedulable pods present")
		ls.scaleDownForbidden = true
	} else {
		glog.V(4).Info("No schedulable pods")
	}

	return ls.onSuccess()
}

// See LoopStateHandler interface
func (ls *loopState) ScaleUp() (ExecutionResult, errors.AutoscalerError) {
	if len(ls.unschedulablePodsToHelp) == 0 {
		glog.V(1).Info("No unschedulable pods")
	} else if ls.autoscaler.MaxNodesTotal > 0 && len(ls.readyNodes) >= ls.autoscaler.MaxNodesTotal {
		glog.V(1).Info("Max total nodes in cluster reached")
	} else if allPodsAreNew(ls.unschedulablePodsToHelp, ls.currentTime) {
		// The assumption here is that these pods have been created very recently and probably there
		// is more pods to come. In theory we could check the newest pod time but then if pod were created
		// slowly but at the pace of 1 every 2 seconds then no scale up would be triggered for long time.
		// We also want to skip a real scale down (just like if the pods were handled).
		ls.scaleDownForbidden = true
		glog.V(1).Info("Unschedulable pods are very new, waiting one iteration for more")
	} else {
		daemonsets, err := ls.autoscaler.ListerRegistry.DaemonSetLister().List()
		if err != nil {
			glog.Errorf("Failed to get daemonset list")
			return ls.onError(errors.ApiCallError, err)
		}

		scaleUpStart := time.Now()
		metrics.UpdateLastTime(metrics.ScaleUp, scaleUpStart)

		scaledUp, typedErr := ScaleUp(ls.context, ls.unschedulablePodsToHelp, ls.readyNodes, daemonsets)

		metrics.UpdateDurationFromStart(metrics.ScaleUp, scaleUpStart)

		if typedErr != nil {
			glog.Errorf("Failed to scale up: %v", typedErr)
			ls.onTypedError(typedErr)
		} else if scaledUp {
			ls.autoscaler.lastScaleUpTime = ls.currentTime
			// No scale down in this iteration.
			return ls.onEarlyExit()
		}
	}

	return ls.onSuccess()
}

// See LoopStateHandler interface
func (ls *loopState) ScaleDown() (ExecutionResult, errors.AutoscalerError) {
	if !ls.autoscaler.ScaleDownEnabled {
		return ls.onSuccess()
	}

	pdbs, err := ls.autoscaler.PodDisruptionBudgetLister().List()
	if err != nil {
		glog.Errorf("Failed to list pod disruption budgets: %v", err)
		return ls.onError(errors.ApiCallError, err)
	}

	unneededStart := time.Now()

	glog.V(4).Infof("Calculating unneeded nodes")

	ls.autoscaler.scaleDown.CleanUp(ls.currentTime)
	potentiallyUnneeded := getPotentiallyUnneededNodes(ls.context, ls.allNodes)

	typedErr := ls.autoscaler.scaleDown.UpdateUnneededNodes(
		ls.allNodes, potentiallyUnneeded,
		append(ls.allScheduled, ls.unschedulableWaitingForLowerPriorityPreemption...),
		ls.currentTime, pdbs)
	if typedErr != nil {
		glog.Errorf("Failed to scale down: %v", typedErr)
		ls.onTypedError(typedErr)
	}

	metrics.UpdateDurationFromStart(metrics.FindUnneeded, unneededStart)

	if glog.V(4) {
		for key, val := range ls.autoscaler.scaleDown.unneededNodes {
			glog.V(4).Infof("%s is unneeded since %s duration %s", key, val.String(), ls.currentTime.Sub(val).String())
		}
	}

	// In dry run only utilization is updated
	calculateUnneededOnly := ls.scaleDownForbidden ||
		ls.autoscaler.lastScaleUpTime.Add(ls.autoscaler.ScaleDownDelayAfterAdd).After(ls.currentTime) ||
		ls.autoscaler.lastScaleDownFailTime.Add(ls.autoscaler.ScaleDownDelayAfterFailure).After(ls.currentTime) ||
		ls.autoscaler.lastScaleDownDeleteTime.Add(ls.autoscaler.ScaleDownDelayAfterDelete).After(ls.currentTime) ||
		ls.autoscaler.scaleDown.nodeDeleteStatus.IsDeleteInProgress()

	glog.V(4).Infof("Scale down status: unneededOnly=%v lastScaleUpTime=%s "+
		"lastScaleDownDeleteTime=%v lastScaleDownFailTime=%s scaleDownForbidden=%v isDeleteInProgress=%v",
		calculateUnneededOnly, ls.autoscaler.lastScaleUpTime, ls.autoscaler.lastScaleDownDeleteTime, ls.autoscaler.lastScaleDownFailTime,
		ls.scaleDownForbidden, ls.autoscaler.scaleDown.nodeDeleteStatus.IsDeleteInProgress())

	if calculateUnneededOnly {
		return ls.onSuccess()
	}

	glog.V(4).Infof("Starting scale down")

	// We want to delete unneeded Node Groups only if there was no recent scale up,
	// and there is no current delete in progress and there was no recent errors.
	if ls.context.NodeAutoprovisioningEnabled {
		err := cleanUpNodeAutoprovisionedGroups(ls.context.CloudProvider, ls.context.LogRecorder)
		if err != nil {
			glog.Warningf("Failed to clean up unneeded node groups: %v", err)
		}
	}

	scaleDownStart := time.Now()
	metrics.UpdateLastTime(metrics.ScaleDown, scaleDownStart)
	result, typedErr := ls.autoscaler.scaleDown.TryToScaleDown(ls.allNodes, ls.allScheduled, pdbs, ls.currentTime)
	metrics.UpdateDurationFromStart(metrics.ScaleDown, scaleDownStart)

	// TODO: revisit result handling
	if typedErr != nil {
		glog.Errorf("Failed to scale down: %v", err)
		ls.autoscaler.lastScaleDownFailTime = ls.currentTime
		return ls.onTypedError(typedErr)
	}
	if result == ScaleDownNodeDeleted || result == ScaleDownNodeDeleteStarted {
		ls.autoscaler.lastScaleDownDeleteTime = ls.currentTime
	}

	return ls.onSuccess()
}

// single-use helper methods start here, extracted to limit complexity of any single method
// all these methods return the values used by phase methods of LoopStateHandler interface

func (ls *loopState) refreshCloudProvider() (ExecutionResult, errors.AutoscalerError) {
	err := ls.context.CloudProvider.Refresh()
	if err != nil {
		glog.Errorf("Failed to refresh cloud provider config: %v", err)
		return ls.onError(errors.CloudProviderError, err)
	}
	return ls.onSuccess()
}

func (ls *loopState) obtainNodeLists() (ExecutionResult, errors.AutoscalerError) {
	var err error

	ls.allNodes, err = ls.autoscaler.AllNodeLister().List()
	if err != nil {
		glog.Errorf("Failed to list all nodes: %v", err)
		return ls.onError(errors.ApiCallError, err)
	}
	if len(ls.allNodes) == 0 {
		return ls.onEmptyCluster("nodes", true)
	}

	ls.readyNodes, err = ls.autoscaler.ReadyNodeLister().List()
	if err != nil {
		glog.Errorf("Failed to list ready nodes: %v", err)
		return ls.onError(errors.ApiCallError, err)
	}

	// Handle GPU case - allocatable GPU may be equal to 0 up to 15 minutes after
	// node registers as ready. See https://github.com/kubernetes/kubernetes/issues/54959
	// Treat those nodes as unready until GPU actually becomes available and let
	// our normal handling for booting up nodes deal with this.
	// TODO: Remove this call when we handle dynamically provisioned resources.
	ls.allNodes, ls.readyNodes = gpu.FilterOutNodesWithUnreadyGpus(ls.allNodes, ls.readyNodes)
	if len(ls.readyNodes) == 0 {
		// Cluster Autoscaler may start running before nodes are ready.
		// Timeout ensures no ClusterUnhealthy events are published immediately in this case.
		emit := ls.currentTime.After(ls.autoscaler.startTime.Add(nodesNotReadyAfterStartTimeout))
		return ls.onEmptyCluster("ready nodes", emit)
	}
	return ls.onSuccess()
}

func (ls *loopState) updateClusterStateRegistry() (ExecutionResult, errors.AutoscalerError) {
	err := ls.autoscaler.ClusterStateRegistry.UpdateNodes(ls.allNodes, ls.currentTime)
	if err != nil {
		glog.Errorf("Failed to update node registry: %v", err)
		ls.autoscaler.scaleDown.CleanUpUnneededNodes()
		return ls.onError(errors.CloudProviderError, err)
	}
	UpdateClusterStateMetrics(ls.autoscaler.ClusterStateRegistry)

	metrics.UpdateDurationFromStart(metrics.UpdateState, ls.startTime)
	metrics.UpdateLastTime(metrics.Autoscaling, time.Now())

	return ls.onSuccess()
}

func (ls *loopState) cleanUnregisteredNodes() (ExecutionResult, errors.AutoscalerError) {
	// Check if there are any nodes that failed to register in Kubernetes master.
	unregisteredNodes := ls.autoscaler.ClusterStateRegistry.GetUnregisteredNodes()
	if len(unregisteredNodes) > 0 {
		glog.V(1).Infof("%d unregistered nodes present", len(unregisteredNodes))
		removedAny, err := removeOldUnregisteredNodes(unregisteredNodes, ls.context, ls.currentTime, ls.context.LogRecorder)
		// There was a problem with removing unregistered nodes. Retry in the next loop.
		if err != nil {
			if removedAny {
				glog.Warningf("Some unregistered nodes were removed, but got error: %v", err)
			} else {
				glog.Errorf("Failed to remove unregistered nodes: %v", err)
			}
			return ls.onError(errors.CloudProviderError, err)
		}
		// Some nodes were removed. Let's skip this iteration, the next one should be better.
		if removedAny {
			glog.V(0).Infof("Some unregistered nodes were removed, skipping iteration")
			return ls.onEarlyExit()
		}
	}
	return ls.onSuccess()
}

func (ls *loopState) checkClusterHealth() (ExecutionResult, errors.AutoscalerError) {
	if !ls.autoscaler.ClusterStateRegistry.IsClusterHealthy() {
		glog.Warning("Cluster is not ready for autoscaling")
		ls.autoscaler.scaleDown.CleanUpUnneededNodes()
		ls.context.LogRecorder.Eventf(apiv1.EventTypeWarning, "ClusterUnhealthy", "Cluster is unhealthy")
		return ls.onEarlyExit()
	}
	return ls.onSuccess()
}

func (ls *loopState) fixNodeGroupSizes() (ExecutionResult, errors.AutoscalerError) {
	// Check if there has been a constant difference between the number of nodes in k8s and
	// the number of nodes on the cloud provider side.
	// TODO: andrewskim - add protection for ready AWS nodes.
	fixedSomething, err := fixNodeGroupSize(ls.context, ls.currentTime)
	if err != nil {
		glog.Errorf("Failed to fix node group sizes: %v", err)
		return ls.onError(errors.CloudProviderError, err)
	}
	if fixedSomething {
		glog.V(0).Infof("Some node group target size was fixed, skipping the iteration")
		return ls.onEarlyExit()
	}

	return ls.onSuccess()
}

// general helper methods start here

func (ls *loopState) onTypedError(err errors.AutoscalerError) (ExecutionResult, errors.AutoscalerError) {
	return ShouldStop, err
}

func (ls *loopState) onError(errType errors.AutoscalerErrorType, err error) (ExecutionResult, errors.AutoscalerError) {
	return ls.onTypedError(errors.ToAutoscalerError(errType, err))
}

func (ls *loopState) onEarlyExit() (ExecutionResult, errors.AutoscalerError) {
	return ShouldStop, nil
}

func (ls *loopState) onSuccess() (ExecutionResult, errors.AutoscalerError) {
	return CanContinue, nil
}

func (ls *loopState) onEmptyCluster(kind string, emitEvent bool) (ExecutionResult, errors.AutoscalerError) {
	ls.autoscaler.scaleDown.CleanUpUnneededNodes()
	UpdateEmptyClusterStateMetrics()

	status := fmt.Sprintf("Cluster has no %v.", kind)
	if ls.context.WriteStatusConfigMap {
		utils.WriteStatusConfigMap(ls.context.ClientSet, ls.context.ConfigNamespace, status, ls.context.LogRecorder)
	}
	if emitEvent {
		ls.context.LogRecorder.Eventf(apiv1.EventTypeWarning, "ClusterUnhealthy", status)
	}
	glog.Warningf("No %v in the cluster", kind)
	return ls.onEarlyExit()
}
