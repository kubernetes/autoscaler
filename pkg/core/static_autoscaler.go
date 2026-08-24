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

package core

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"
	"time"

	"k8s.io/utils/integer"
	"sigs.k8s.io/cluster-autoscaler/pkg/capacitybuffer/fakepods"
	"sigs.k8s.io/cluster-autoscaler/pkg/cloudprovider"
	"sigs.k8s.io/cluster-autoscaler/pkg/clusterstate"
	"sigs.k8s.io/cluster-autoscaler/pkg/clusterstate/scaleupfailures"
	"sigs.k8s.io/cluster-autoscaler/pkg/clusterstate/utils"
	"sigs.k8s.io/cluster-autoscaler/pkg/config"
	ca_context "sigs.k8s.io/cluster-autoscaler/pkg/context"
	"sigs.k8s.io/cluster-autoscaler/pkg/core/scaledown"
	"sigs.k8s.io/cluster-autoscaler/pkg/core/scaledown/actuation"
	"sigs.k8s.io/cluster-autoscaler/pkg/core/scaledown/deletiontracker"
	"sigs.k8s.io/cluster-autoscaler/pkg/core/scaledown/latencytracker"
	"sigs.k8s.io/cluster-autoscaler/pkg/core/scaledown/pdb"
	"sigs.k8s.io/cluster-autoscaler/pkg/core/scaledown/planner"
	scaledownstatus "sigs.k8s.io/cluster-autoscaler/pkg/core/scaledown/status"
	"sigs.k8s.io/cluster-autoscaler/pkg/core/scaleup"
	"sigs.k8s.io/cluster-autoscaler/pkg/core/scaleup/orchestrator"
	core_utils "sigs.k8s.io/cluster-autoscaler/pkg/core/utils"
	"sigs.k8s.io/cluster-autoscaler/pkg/debuggingsnapshot"
	"sigs.k8s.io/cluster-autoscaler/pkg/estimator"
	"sigs.k8s.io/cluster-autoscaler/pkg/expander"
	"sigs.k8s.io/cluster-autoscaler/pkg/metrics"
	"sigs.k8s.io/cluster-autoscaler/pkg/observers/loopstart"
	"sigs.k8s.io/cluster-autoscaler/pkg/observers/nodegroupchange"
	ca_processors "sigs.k8s.io/cluster-autoscaler/pkg/processors"
	"sigs.k8s.io/cluster-autoscaler/pkg/processors/nodeinfosprovider"
	"sigs.k8s.io/cluster-autoscaler/pkg/processors/status"
	"sigs.k8s.io/cluster-autoscaler/pkg/resourcequotas"
	"sigs.k8s.io/cluster-autoscaler/pkg/simulator"
	"sigs.k8s.io/cluster-autoscaler/pkg/simulator/clustersnapshot"
	csinodeprovider "sigs.k8s.io/cluster-autoscaler/pkg/simulator/csi/provider"
	csisnapshot "sigs.k8s.io/cluster-autoscaler/pkg/simulator/csi/snapshot"
	"sigs.k8s.io/cluster-autoscaler/pkg/simulator/drainability/rules"
	draprovider "sigs.k8s.io/cluster-autoscaler/pkg/simulator/dynamicresources/provider"
	drasnapshot "sigs.k8s.io/cluster-autoscaler/pkg/simulator/dynamicresources/snapshot"
	"sigs.k8s.io/cluster-autoscaler/pkg/simulator/framework"
	"sigs.k8s.io/cluster-autoscaler/pkg/simulator/options"
	"sigs.k8s.io/cluster-autoscaler/pkg/utils/annotations"
	"sigs.k8s.io/cluster-autoscaler/pkg/utils/backoff"
	caerrors "sigs.k8s.io/cluster-autoscaler/pkg/utils/errors"
	kube_util "sigs.k8s.io/cluster-autoscaler/pkg/utils/kubernetes"
	"sigs.k8s.io/cluster-autoscaler/pkg/utils/taints"

	v1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
)

const (
	// How old the oldest unschedulable pod should be before starting scale up.
	unschedulablePodTimeBuffer = 2 * time.Second
	// How old the oldest unschedulable pod with GPU should be before starting scale up.
	// The idea is that nodes with GPU are very expensive and we're ready to sacrifice
	// a bit more latency to wait for more pods and make a more informed scale-up decision.
	unschedulablePodWithGpuTimeBuffer = 30 * time.Second
	// Number of autoscaler loops in a graceful degradation cycle.
	// When autoscaler is unable to handle a large number of pending pods, when attempting to
	// run scheduling simulations, autoscaler enters to a graceful degradation mode.
	// When graceful degradation mode is active, autoscaler does scale ups and scale downs in
	// alternative loops so scale down is not starved because of scale ups. Sometimes
	// scale up could be blocked due to scale down as well. When graceful degradation
	// mode is active every 4th loop is dedicated for scale down and the other loops
	// are dedicated for scale up.
	gracefulDegradationCycle = 4
)

// StaticAutoscaler is an autoscaler which has all the core functionality of a CA but without the reconfiguration feature
type StaticAutoscaler struct {
	// AutoscalingContext consists of validated settings and options for this autoscaler
	*ca_context.AutoscalingContext
	// ClusterState for maintaining the state of cluster nodes.
	clusterStateRegistry       *clusterstate.ClusterStateRegistry
	lastScaleUpTime            time.Time
	lastScaleDownDeleteTime    time.Time
	lastScaleDownFailTime      time.Time
	scaleDownPlanner           scaledown.Planner
	scaleDownActuator          scaledown.Actuator
	scaleUpOrchestrator        scaleup.Orchestrator
	processors                 *ca_processors.AutoscalingProcessors
	loopStartNotifier          *loopstart.ObserversList
	processorCallbacks         *staticAutoscalerProcessorCallbacks
	initialized                bool
	taintConfig                taints.TaintConfig
	capacityBufferPodsRegistry *fakepods.Registry
}

type staticAutoscalerProcessorCallbacks struct {
	disableScaleDownForLoop bool
	extraValues             map[string]interface{}
	scaleDownPlanner        scaledown.Planner
}

func (callbacks *staticAutoscalerProcessorCallbacks) ResetUnneededNodes(ctx context.Context) {
	callbacks.scaleDownPlanner.CleanUpUnneededNodes(ctx)
}

func newStaticAutoscalerProcessorCallbacks() *staticAutoscalerProcessorCallbacks {
	callbacks := &staticAutoscalerProcessorCallbacks{}
	callbacks.reset()
	return callbacks
}

func (callbacks *staticAutoscalerProcessorCallbacks) DisableScaleDownForLoop() {
	callbacks.disableScaleDownForLoop = true
}

func (callbacks *staticAutoscalerProcessorCallbacks) SetExtraValue(key string, value interface{}) {
	callbacks.extraValues[key] = value
}

func (callbacks *staticAutoscalerProcessorCallbacks) GetExtraValue(key string) (value interface{}, found bool) {
	value, found = callbacks.extraValues[key]
	return
}

func (callbacks *staticAutoscalerProcessorCallbacks) reset() {
	callbacks.disableScaleDownForLoop = false
	callbacks.extraValues = make(map[string]interface{})
}

// NewStaticAutoscaler creates an instance of Autoscaler filled with provided parameters
func NewStaticAutoscaler(
	opts config.AutoscalingOptions,
	fwHandle *framework.Handle,
	clusterSnapshot clustersnapshot.ClusterSnapshot,
	autoscalingKubeClients *ca_context.AutoscalingKubeClients,
	processors *ca_processors.AutoscalingProcessors,
	loopStartObservers []loopstart.Observer,
	cloudProvider cloudprovider.CloudProvider,
	expanderStrategy expander.Strategy,
	estimatorBuilder estimator.EstimatorBuilder,
	backoff backoff.Backoff,
	scaleUpFailuresRegistry *scaleupfailures.Registry,
	debuggingSnapshotter debuggingsnapshot.DebuggingSnapshotter,
	remainingPdbTracker pdb.RemainingPdbTracker,
	scaleUpOrchestrator scaleup.Orchestrator,
	deleteOptions options.NodeDeleteOptions,
	drainabilityRules rules.Rules,
	draProvider *draprovider.Provider,
	quotasTrackerOptions resourcequotas.TrackerOptions,
	minQuotasTrackerOptions resourcequotas.TrackerOptions,
	csiProvider *csinodeprovider.Provider,
	capacityBufferPodsRegistry *fakepods.Registry) *StaticAutoscaler {

	klog.V(4).Infof("Creating new static autoscaler with opts: %v", opts)

	templateNodeInfoRegistry := nodeinfosprovider.NewTemplateNodeInfoRegistry(processors.TemplateNodeInfoProvider)

	clusterStateConfig := clusterstate.ClusterStateRegistryConfig{
		MaxTotalUnreadyPercentage: opts.MaxTotalUnreadyPercentage,
		OkTotalUnreadyCount:       opts.OkTotalUnreadyCount,
	}
	// Register the scale up failures registry before CSR so that it is updated before the node group is sent to the backoff.
	processors.ScaleStateNotifier.Register(scaleUpFailuresRegistry)
	clusterStateRegistry := clusterstate.NewNotifiedClusterStateRegistry(cloudProvider, autoscalingKubeClients.LogRecorder, backoff, processors.NodeGroupConfigProcessor, templateNodeInfoRegistry, clusterstate.WithScaleUpFailuresRegistry(scaleUpFailuresRegistry), clusterstate.WithConfig(clusterStateConfig), clusterstate.WithAsyncNodeGroupStateChecker(processors.AsyncNodeGroupStateChecker), clusterstate.WithScaleStateNotifier(processors.ScaleStateNotifier))
	processorCallbacks := newStaticAutoscalerProcessorCallbacks()

	autoscalingCtx := ca_context.NewAutoscalingContext(
		opts,
		fwHandle,
		clusterSnapshot,
		autoscalingKubeClients,
		cloudProvider,
		expanderStrategy,
		processorCallbacks,
		debuggingSnapshotter,
		remainingPdbTracker,
		clusterStateRegistry,
		draProvider,
		templateNodeInfoRegistry,
		csiProvider)

	taintConfig := taints.NewTaintConfig(opts)

	processors.ScaleDownCandidatesNotifier.Register(clusterStateRegistry)
	processors.ScaleStateNotifier.Register(nodegroupchange.NewNodeGroupChangeMetricsProducer(cloudProvider, metrics.DefaultMetrics, templateNodeInfoRegistry))

	// TODO: Populate the ScaleDownActuator/Planner fields in AutoscalingContext
	// during the struct creation rather than here.
	var ndlt *latencytracker.NodeLatencyTracker
	if autoscalingCtx.AutoscalingOptions.NodeRemovalLatencyTrackingEnabled {
		ndlt = latencytracker.NewNodeLatencyTracker(processors.ScaleDownStatusProcessor)
		processors.ScaleDownCandidatesNotifier.Register(ndlt)
		processors.ScaleDownStatusProcessor = ndlt
	}
	quotasTrackerFactory := resourcequotas.NewTrackerFactory(quotasTrackerOptions)
	minQuotasTrackerFactory := resourcequotas.NewTrackerFactory(minQuotasTrackerOptions)

	scaleDownPlanner := planner.New(autoscalingCtx, processors, deleteOptions, drainabilityRules, minQuotasTrackerFactory)
	processorCallbacks.scaleDownPlanner = scaleDownPlanner

	ndt := deletiontracker.NewNodeDeletionTracker(0 * time.Second)
	scaleDownActuator := actuation.NewActuator(autoscalingCtx, processors.ScaleStateNotifier, ndt, deleteOptions, drainabilityRules, processors.NodeGroupConfigProcessor)
	autoscalingCtx.ScaleDownActuator = scaleDownActuator
	if scaleUpOrchestrator == nil {
		scaleUpOrchestrator = orchestrator.New()
	}
	scaleUpOrchestrator.Initialize(autoscalingCtx, processors, clusterStateRegistry, estimatorBuilder, taintConfig, quotasTrackerFactory)

	// Set the initial scale times to be less than the start time so as to
	// not start in cooldown mode.
	initialScaleTime := time.Now().Add(-time.Hour)
	return &StaticAutoscaler{
		AutoscalingContext:         autoscalingCtx,
		lastScaleUpTime:            initialScaleTime,
		lastScaleDownDeleteTime:    initialScaleTime,
		lastScaleDownFailTime:      initialScaleTime,
		scaleDownPlanner:           scaleDownPlanner,
		scaleDownActuator:          scaleDownActuator,
		scaleUpOrchestrator:        scaleUpOrchestrator,
		processors:                 processors,
		loopStartNotifier:          loopstart.NewObserversList(loopStartObservers),
		processorCallbacks:         processorCallbacks,
		clusterStateRegistry:       clusterStateRegistry,
		taintConfig:                taintConfig,
		capacityBufferPodsRegistry: capacityBufferPodsRegistry,
	}
}

// LastScaleUpTime returns last scale up time
func (a *StaticAutoscaler) LastScaleUpTime() time.Time {
	return a.lastScaleUpTime
}

// LastScaleDownDeleteTime returns the last successful scale down time
func (a *StaticAutoscaler) LastScaleDownDeleteTime() time.Time {
	return a.lastScaleDownDeleteTime
}

// Start starts components running in background.
func (a *StaticAutoscaler) Start() error {
	a.clusterStateRegistry.Start()
	return nil
}

// cleanUpIfRequired removes ToBeDeleted taints added by a previous run of CA
// the taints are removed only once per runtime
func (a *StaticAutoscaler) cleanUpIfRequired(ctx context.Context) {
	logger := klog.FromContext(ctx)
	if a.initialized {
		return
	}

	// CA can die at any time. Removing taints that might have been left from the previous run.
	if allNodes, err := a.AllNodeLister().List(); err != nil {
		logger.Error(err, "Failed to list ready nodes, not cleaning up taints")
	} else {
		// Make sure we are only cleaning taints from selected node groups.
		selectedNodes := filterNodesFromSelectedGroups(ctx, a.CloudProvider, allNodes...)
		taints.CleanAllToBeDeleted(ctx, selectedNodes,
			a.AutoscalingContext.ClientSet, a.Recorder, a.CordonNodeBeforeTerminate)
		if a.AutoscalingContext.AutoscalingOptions.MaxBulkSoftTaintCount == 0 {
			// Clean old taints if soft taints handling is disabled
			taints.CleanStaleDeletionCandidates(ctx, allNodes,
				a.AutoscalingContext.ClientSet, a.Recorder, a.NodeDeletionCandidateTTL)
		}
	}
	a.initialized = true
}

func (a *StaticAutoscaler) initializeRemainingPdbTracker(ctx context.Context) caerrors.AutoscalerError {
	logger := klog.FromContext(ctx)
	a.RemainingPdbTracker.Clear()

	pdbs, err := a.PodDisruptionBudgetLister().List()
	if err != nil {
		logger.Error(err, "Failed to list pod disruption budgets")
		return caerrors.NewAutoscalerError(caerrors.ApiCallError, err.Error())
	}
	err = a.RemainingPdbTracker.SetPdbs(pdbs)
	if err != nil {
		return caerrors.NewAutoscalerError(caerrors.InternalError, err.Error())
	}
	return nil
}

// RunOnce iterates over node groups and scales them up/down if necessary
func (a *StaticAutoscaler) RunOnce(ctx context.Context, currentTime time.Time) caerrors.AutoscalerError {
	logger := klog.FromContext(ctx)
	a.cleanUpIfRequired(ctx)
	a.processorCallbacks.reset()
	a.DebuggingSnapshotter.StartDataCollection(ctx)
	defer a.DebuggingSnapshotter.Flush(ctx)
	if a.capacityBufferPodsRegistry != nil {
		a.capacityBufferPodsRegistry.Clear()
	}

	podLister := a.AllPodLister()
	autoscalingCtx := a.AutoscalingContext
	logger.V(4).Info("Starting main loop")

	stateUpdateStart := time.Now()

	var draSnapshot *drasnapshot.Snapshot
	if a.AutoscalingContext.DynamicResourceAllocationEnabled && a.AutoscalingContext.DraProvider != nil {
		var err error
		draSnapshot, err = a.AutoscalingContext.DraProvider.Snapshot()
		if err != nil {
			return caerrors.ToAutoscalerError(caerrors.ApiCallError, err)
		}
	}

	var csiSnapshot *csisnapshot.Snapshot
	if a.AutoscalingContext.CsiProvider != nil {
		var err error
		csiSnapshot, err = a.AutoscalingContext.CsiProvider.Snapshot()
		if err != nil {
			return caerrors.ToAutoscalerError(caerrors.ApiCallError, err)
		}
	}

	// Get nodes and pods currently living on cluster
	allNodes, readyNodes, typedErr := a.obtainNodeLists(ctx, draSnapshot, csiSnapshot)
	if typedErr != nil {
		logger.Error(typedErr, "Failed to get node list")
		return typedErr
	}

	if abortLoop, err := a.processors.ActionableClusterProcessor.ShouldAbort(ctx, a.AutoscalingContext, allNodes, readyNodes, currentTime); abortLoop {
		return err
	}

	podsBySchedulability, err := listPods(ctx, podLister, a.BypassedSchedulers, a.AllowedSchedulers)
	if err != nil {
		return caerrors.ToAutoscalerError(caerrors.ApiCallError, err)
	}

	coresTotal, memoryTotal := calculateCoresMemoryTotal(allNodes, currentTime)
	metrics.UpdateClusterCPUCurrentCores(coresTotal)
	metrics.UpdateClusterMemoryCurrentBytes(memoryTotal)

	daemonsets, err := a.ListerRegistry.DaemonSetLister().List(labels.Everything())
	if err != nil {
		logger.Error(err, "Failed to get daemonset list")
		return caerrors.ToAutoscalerError(caerrors.ApiCallError, err)
	}

	// Snapshot scale-down actuation status before cache refresh.
	scaleDownActuationStatus := a.scaleDownActuator.CheckStatus()
	// Call CloudProvider.Refresh before any other calls to cloud provider.
	refreshStart := time.Now()
	err = a.AutoscalingContext.CloudProvider.Refresh(ctx)
	if a.AutoscalingOptions.AsyncNodeGroupsEnabled {
		// Some node groups might have been created asynchronously, without registering in CSR.
		a.clusterStateRegistry.Recalculate(ctx)
	}
	metrics.UpdateDurationFromStart(ctx, metrics.CloudProviderRefresh, refreshStart)
	if err != nil {
		logger.Error(err, "Failed to refresh cloud provider config")
		return caerrors.ToAutoscalerError(caerrors.CloudProviderError, err)
	}
	a.loopStartNotifier.Refresh(ctx)

	// Update node groups min/max and maximum number of nodes being set for all node groups after cloud provider refresh
	maxNodesCount := 0
	for _, nodeGroup := range a.AutoscalingContext.CloudProvider.NodeGroups(ctx) {
		// Don't report non-existing or upcoming node groups
		if nodeGroup.Exist(ctx) {
			metrics.UpdateNodeGroupMin(nodeGroup.Id(), nodeGroup.MinSize(ctx))
			metrics.UpdateNodeGroupMax(nodeGroup.Id(), nodeGroup.MaxSize(ctx))
			maxNodesCount += nodeGroup.MaxSize(ctx)
		}
	}
	if a.MaxNodesTotal > 0 {
		metrics.UpdateMaxNodesCount(integer.IntMin(a.MaxNodesTotal, maxNodesCount))
	} else {
		metrics.UpdateMaxNodesCount(maxNodesCount)
	}
	allocatedPods := slices.Concat(podsBySchedulability.Scheduled, podsBySchedulability.NominatedNode)
	nonExpendableAllocatedPods := core_utils.FilterOutExpendablePods(allocatedPods, a.ExpendablePodsPriorityCutoff)

	if err := a.ClusterSnapshot.SetClusterState(ctx, allNodes, nonExpendableAllocatedPods, draSnapshot, csiSnapshot); err != nil {
		return caerrors.ToAutoscalerError(caerrors.InternalError, err).AddPrefix("failed to initialize ClusterSnapshot: ")
	}
	// Initialize Pod Disruption Budget tracking
	if typedErr := a.initializeRemainingPdbTracker(ctx); typedErr != nil {
		return typedErr.AddPrefix("failed to initialize RemainingPdbTracker: ")
	}

	if autoscalerError := a.AutoscalingContext.TemplateNodeInfoRegistry.Recompute(ctx, a.AutoscalingContext, allNodes, daemonsets, a.taintConfig, currentTime); autoscalerError != nil {
		logger.Error(autoscalerError, "Failed to recompute template node infos")
		return autoscalerError.AddPrefix("failed to recompute template node infos: ")
	}

	a.DebuggingSnapshotter.SetTemplateNodes(ctx, autoscalingCtx.TemplateNodeInfoRegistry.GetNodeInfos())

	if typedErr := a.updateClusterState(ctx, allNodes, currentTime); typedErr != nil {
		logger.Error(typedErr, "Failed to update cluster state")
		return typedErr
	}
	metrics.UpdateDurationFromStart(ctx, metrics.UpdateState, stateUpdateStart)

	scaleUpStatus := &status.ScaleUpStatus{Result: status.ScaleUpNotTried}
	scaleUpTriggered := false
	scaleDownStatus := &scaledownstatus.ScaleDownStatus{Result: scaledownstatus.ScaleDownNotTried}

	defer func() {
		// Update status information when the loop is done (regardless of reason)
		if autoscalingCtx.WriteStatusConfigMap {
			status := a.clusterStateRegistry.GetStatus(ctx, currentTime)
			utils.WriteStatusConfigMap(ctx, autoscalingCtx.ClientSet, autoscalingCtx.ConfigNamespace,
				*status, a.AutoscalingContext.LogRecorder, a.AutoscalingContext.StatusConfigMapName, currentTime)
		}

		// This deferred processor execution allows the processors to handle a situation when a scale-(up|down)
		// wasn't even attempted because e.g. the iteration exited earlier.
		if !scaleUpTriggered && a.processors.ScaleUpStatusProcessor != nil {
			a.processors.ScaleUpStatusProcessor.Process(ctx, a.AutoscalingContext, scaleUpStatus)
		}
		if a.processors.ScaleDownStatusProcessor != nil {
			// Gather status before scaledown status processor invocation
			nodeDeletionResults, nodeDeletionResultsAsOf := a.scaleDownActuator.DeletionResults()
			scaleDownStatus.NodeDeleteResults = nodeDeletionResults
			scaleDownStatus.NodeDeleteResultsAsOf = nodeDeletionResultsAsOf
			a.scaleDownActuator.ClearResultsNotNewerThan(scaleDownStatus.NodeDeleteResultsAsOf)
			scaleDownStatus.SetUnremovableNodesInfo(ctx, a.scaleDownPlanner.UnremovableNodes(), a.scaleDownPlanner.NodeUtilizationMap(), a.CloudProvider)

			a.processors.ScaleDownStatusProcessor.Process(ctx, a.AutoscalingContext, scaleDownStatus)
		}

		if a.processors.AutoscalingStatusProcessor != nil {
			err := a.processors.AutoscalingStatusProcessor.Process(ctx, a.AutoscalingContext, a.clusterStateRegistry, currentTime)
			if err != nil {
				logger.Error(err, "AutoscalingStatusProcessor error")
			}
		}
	}()

	// Check if there are any nodes that failed to register in Kubernetes
	// master.
	unregisteredNodes := a.clusterStateRegistry.GetUnregisteredNodes()
	if len(unregisteredNodes) > 0 {
		logger.V(1).Info("unregistered nodes present", "unregisteredNodesCount", len(unregisteredNodes))
		removedAny, err := a.removeOldUnregisteredNodes(ctx, unregisteredNodes,
			a.clusterStateRegistry, currentTime, autoscalingCtx.LogRecorder)
		// There was a problem with removing unregistered nodes. Retry in the next loop.
		if err != nil {
			logger.Info("Failed to remove unregistered nodes", "err", err)
		}
		if removedAny {
			logger.V(0).Info("Some unregistered nodes were removed")
		}
	}

	if !a.clusterStateRegistry.IsClusterHealthy() {
		logger.Info("Cluster is not ready for autoscaling")
		a.scaleDownPlanner.CleanUpUnneededNodes(ctx)
		autoscalingCtx.LogRecorder.Eventf(apiv1.EventTypeWarning, "ClusterUnhealthy", "Cluster is unhealthy")
		return nil
	}

	a.deleteCreatedNodesWithErrors(ctx)

	// Check if there has been a constant difference between the number of nodes in k8s and
	// the number of nodes on the cloud provider side.
	// TODO: andrewskim - add protection for ready AWS nodes.
	fixedSomething, err := fixNodeGroupSize(ctx, autoscalingCtx, a.clusterStateRegistry, currentTime)
	if err != nil {
		logger.Error(err, "Failed to fix node group sizes")
		return caerrors.ToAutoscalerError(caerrors.CloudProviderError, err)
	}
	if fixedSomething {
		logger.V(0).Info("Some node group target size was fixed, skipping the iteration")
		return nil
	}

	metrics.UpdateLastTime(metrics.Autoscaling, time.Now())

	// SchedulerUnprocessed might be zero here if it was disabled
	metrics.UpdateUnschedulablePodsCount(len(podsBySchedulability.Unschedulable), len(podsBySchedulability.Unprocessed))
	// Treat unknown pods as unschedulable, pod list processor will remove schedulable pods
	podsBySchedulability.Unschedulable = append(podsBySchedulability.Unschedulable, podsBySchedulability.Unprocessed...)
	// Upcoming nodes are recently created nodes that haven't registered in the cluster yet, or haven't become ready yet.
	upcomingCounts, registeredUpcoming := a.clusterStateRegistry.GetUpcomingNodes(ctx)
	// For each upcoming node we inject a placeholder node faked to appear ready into the cluster snapshot, so that we can pack unschedulable pods on
	// them and not trigger another scale-up.
	// The fake nodes are intentionally not added to the all nodes list, so that they are not considered as candidates for scale-down (which
	// doesn't make sense as they're not real).
	templateNodeInfos := a.AutoscalingContext.TemplateNodeInfoRegistry.GetNodeInfos()
	if _, err := a.addUpcomingNodesToClusterSnapshot(ctx, upcomingCounts, templateNodeInfos, "upcoming-%d"); err != nil {
		logger.Error(err, "Failed adding upcoming nodes to cluster snapshot")
		return caerrors.ToAutoscalerError(caerrors.InternalError, err)
	}
	// Some upcoming nodes can already be registered in the cluster, but not yet ready - we still inject replacements for them above. The actual registered nodes
	// have to be filtered out of the all nodes list so that scale-down can't consider them as candidates. Otherwise, with aggressive scale-down settings, we
	// could be removing the nodes before they have a chance to first become ready (the duration of which should be unrelated to the scale-down settings).
	var allRegisteredUpcoming []string
	for _, ngRegisteredUpcoming := range registeredUpcoming {
		allRegisteredUpcoming = append(allRegisteredUpcoming, ngRegisteredUpcoming...)
	}
	allNodes = subtractNodesByName(allNodes, allRegisteredUpcoming)
	// Remove the nodes from the snapshot as well so that the state is consistent.
	for _, notStartedNodeName := range allRegisteredUpcoming {
		err := a.ClusterSnapshot.RemoveNodeInfo(ctx, notStartedNodeName)
		if err != nil {
			logger.Error(err, "Failed to remove NotStarted node from cluster snapshot", "notStartedNode", notStartedNodeName)

			// ErrNodeNotFound shouldn't happen (so it needs to be logged above if it does), but what we care about here is that the
			// node is not in the snapshot - so we don't have to error out in that case.
			if !errors.Is(err, clustersnapshot.ErrNodeNotFound) {
				return caerrors.ToAutoscalerError(caerrors.InternalError, err)
			}
		}
	}
	allNodeInfos, err := a.ClusterSnapshot.ListNodeInfos()
	if err != nil {
		logger.Error(err, "Unable to fetch ClusterNode List for Debugging Snapshot")
	} else {
		a.AutoscalingContext.DebuggingSnapshotter.SetClusterNodes(ctx, allNodeInfos)
	}

	unschedulablePodsToHelp, err := a.processors.PodListProcessor.Process(ctx, a.AutoscalingContext, podsBySchedulability.Unschedulable)

	if err != nil {
		logger.Info("Failed to process unschedulable pods", "err", err)
	}

	// finally, filter out pods that are too "young" to safely be considered for a scale-up (delay is configurable)
	unschedulablePodsToHelp = a.filterOutYoungPods(ctx, unschedulablePodsToHelp, currentTime)

	shouldScaleUp, scaleUpStatus := a.shouldScaleUp(ctx, unschedulablePodsToHelp, scaleUpStatus, readyNodes, currentTime)

	if err := ctx.Err(); err != nil {
		logger.V(0).Info("Skipping scale-up scale-down, context cancelled", "err", err)
		return nil
	}

	if shouldScaleUp {
		scaleUpTriggered = true
		nodes := make([]*apiv1.Node, len(allNodeInfos))
		for i, nodeInfo := range allNodeInfos {
			nodes[i] = nodeInfo.Node()
		}

		if a.AutoscalingContext.AutoscalingOptions.SalvoScaleUp {
			scaleUpStatus, typedErr = a.runScaleUpSalvo(ctx, currentTime,
				unschedulablePodsToHelp,
				daemonsets,
				nodes,
				templateNodeInfos,
			)
		} else {
			_, scaleUpStatus, typedErr = a.runSingleScaleUp(ctx, currentTime,
				unschedulablePodsToHelp,
				daemonsets,
				nodes,
				templateNodeInfos,
			)
		}

		if scaleUpStatus.Result == status.ScaleUpSuccessful {
			// No scale down in this iteration.
			scaleDownStatus.Result = scaledownstatus.ScaleDownInCooldown
		}
	}

	if a.shouldScaleDown() {
		if typedErr = a.scaleDown(ctx, currentTime, allNodes, scaleDownActuationStatus, scaleDownStatus); typedErr != nil {
			return typedErr
		}
	}

	if a.EnforceNodeGroupMinSize {
		scaleUpTriggered = true
		nodes := make([]*apiv1.Node, len(allNodeInfos))
		for i, nodeInfo := range allNodeInfos {
			nodes[i] = nodeInfo.Node()
		}

		scaleUpFn := func() (*status.ScaleUpStatus, caerrors.AutoscalerError) {
			return a.scaleUpOrchestrator.ScaleUpToNodeGroupMinSize(ctx, nodes, templateNodeInfos)
		}
		_, scaleUpStatus, typedErr = a.instrumentedScaleUp(ctx, currentTime, scaleUpFn)
	}

	return nil
}

func (a *StaticAutoscaler) shouldScaleUp(ctx context.Context, unschedulablePodsToHelp []*apiv1.Pod, scaleUpStatus *status.ScaleUpStatus, readyNodes []*apiv1.Node, currentTime time.Time) (bool, *status.ScaleUpStatus) {
	logger := klog.FromContext(ctx)
	shouldScaleUp := true
	if len(unschedulablePodsToHelp) == 0 {
		scaleUpStatus.Result = status.ScaleUpNotNeeded
		logger.V(1).Info("No unschedulable pods")
		shouldScaleUp = false
	} else if a.MaxNodesTotal > 0 && len(readyNodes) >= a.MaxNodesTotal {
		scaleUpStatus.Result = status.ScaleUpLimitedByMaxNodesTotal
		logger.Info("Max total nodes in cluster reached", "maxNodesTotal", a.MaxNodesTotal, "readyNodesCount", len(readyNodes))
		a.LogRecorder.Eventf(apiv1.EventTypeWarning, "MaxNodesTotalReached",
			"Max total nodes in cluster reached: %v", a.MaxNodesTotal)
		shouldScaleUp = false

		noScaleUpInfoForPods := []status.NoScaleUpInfo{}
		for _, pod := range unschedulablePodsToHelp {
			noScaleUpInfo := status.NoScaleUpInfo{
				Pod: pod,
			}
			noScaleUpInfoForPods = append(noScaleUpInfoForPods, noScaleUpInfo)
		}
		scaleUpStatus.PodsRemainUnschedulable = noScaleUpInfoForPods
	} else if len(a.BypassedSchedulers) == 0 && allPodsAreNew(unschedulablePodsToHelp, currentTime) {
		// The assumption here is that these pods have been created very recently and probably there
		// is more pods to come. In theory, we could check the newest pod time but then if pod were created
		// slowly but at the pace of 1 every 2 seconds then no scale up would be triggered for long time.
		// We also want to skip a real scale down (just like if the pods were handled).
		// This logic only makes sense if CA is not trying to react quickly
		// by bypassing scheduler marking pods as unschedulable.
		a.processorCallbacks.DisableScaleDownForLoop()
		scaleUpStatus.Result = status.ScaleUpInCooldown
		logger.V(1).Info("Unschedulable pods are very new, waiting one iteration for more")
		shouldScaleUp = false
	}

	shouldScaleUp = shouldScaleUp || a.processors.ScaleUpEnforcer.ShouldForceScaleUp(unschedulablePodsToHelp)

	if a.GracefulDegradationEnabled && a.GracefulDegradationLoopCount > 0 && a.GracefulDegradationLoopCount%gracefulDegradationCycle == 0 {
		shouldScaleUp = false
		scaleUpStatus.Result = status.ScaleUpInCooldown
	}

	return shouldScaleUp, scaleUpStatus
}

// instrumentedScaleUp handles a single ScaleUp orchestrator call with metrics and status reporting.
// It accepts a generic scaleUpFn closure, allowing it to handle regular scale-ups or scaling up to node group min size.
func (a *StaticAutoscaler) instrumentedScaleUp(
	ctx context.Context,
	currentTime time.Time,
	scaleUpFn func() (*status.ScaleUpStatus, caerrors.AutoscalerError),
) ([]*apiv1.Pod, *status.ScaleUpStatus, caerrors.AutoscalerError) {
	logger := klog.FromContext(ctx)
	scaleUpStart := time.Now()
	metrics.UpdateLastTime(metrics.ScaleUp, scaleUpStart)

	scaleUpStatus, typedErr := scaleUpFn()
	// Reference copy is sufficient since processors are not expected to modify the slice elements.
	unfilteredPodsTriggeredScaleUp := scaleUpStatus.PodsTriggeredScaleUp

	metrics.UpdateDurationFromStart(ctx, metrics.ScaleUp, scaleUpStart)

	if a.processors.ScaleUpStatusProcessor != nil {
		a.processors.ScaleUpStatusProcessor.Process(ctx, a.AutoscalingContext, scaleUpStatus)
	}

	if typedErr != nil {
		logger.Error(typedErr, "Failed to scale up")
		return unfilteredPodsTriggeredScaleUp, scaleUpStatus, typedErr
	}
	if scaleUpStatus.Result == status.ScaleUpSuccessful {
		a.lastScaleUpTime = currentTime
	}

	return unfilteredPodsTriggeredScaleUp, scaleUpStatus, typedErr
}

func (a *StaticAutoscaler) runSingleScaleUp(
	ctx context.Context,
	currentTime time.Time,
	unschedulablePodsToHelp []*apiv1.Pod,
	daemonsets []*v1.DaemonSet,
	nodes []*apiv1.Node,
	templateNodeInfos map[string]*framework.NodeInfo,
) ([]*apiv1.Pod, *status.ScaleUpStatus, caerrors.AutoscalerError) {
	scaleUpFn := func() (*status.ScaleUpStatus, caerrors.AutoscalerError) {
		return a.scaleUpOrchestrator.ScaleUp(ctx, unschedulablePodsToHelp, nodes, daemonsets, templateNodeInfos, false)
	}
	return a.instrumentedScaleUp(ctx, currentTime, scaleUpFn)
}

func (a *StaticAutoscaler) runScaleUpSalvo(
	ctx context.Context,
	currentTime time.Time,
	unschedulablePodsToHelp []*apiv1.Pod,
	daemonsets []*v1.DaemonSet,
	nodes []*apiv1.Node,
	templateNodeInfos map[string]*framework.NodeInfo,
) (*status.ScaleUpStatus, caerrors.AutoscalerError) {
	logger := klog.FromContext(ctx)
	var scaleUpStatus *status.ScaleUpStatus
	var typedErr caerrors.AutoscalerError
	var handledPods []*apiv1.Pod

	podsMap := make(map[types.UID]*apiv1.Pod)
	for _, pod := range unschedulablePodsToHelp {
		podsMap[pod.UID] = pod
	}

	budget := a.AutoscalingContext.AutoscalingOptions.SalvoScaleUpBudget
	salvoCtx, cancel := context.WithTimeout(context.Background(), budget)
	defer cancel()
	logger.Info("Starting scale up salvo", "podsCount", len(podsMap), "budget", budget)
	i := 0
	for ; ; i++ {
		logger.V(4).Info("Scale up salvo: starting iteration", "iteration", i, "podsLeftCount", len(podsMap))
		unschedulablePods := slices.Collect(maps.Values(podsMap))

		handledPods, scaleUpStatus, typedErr = a.runSingleScaleUp(ctx, currentTime, unschedulablePods, daemonsets, nodes, templateNodeInfos)
		if typedErr != nil {
			logger.Info("Scale up failed, finishing the scale up salvo", "err", typedErr)
			break
		}
		if !scaleUpStatus.WasSuccessful() {
			logger.Info("Scale up not successful, finishing the scale up salvo", "scaleUpResult", scaleUpStatus.Result)
			break
		}
		if len(handledPods) == 0 {
			logger.Info("Empty unfilteredPodsTriggeredScaleUp list - cannot update cluster snapshot, finishing the scale up salvo")

			break
		}

		for _, pod := range handledPods {
			delete(podsMap, pod.UID)
		}

		if len(podsMap) == 0 {
			logger.Info("All unschedulable pods have been helped, finishing the scale up salvo")

			break
		}

		if err := salvoCtx.Err(); err != nil {
			logger.Info("Scale up budget exhausted, finishing the scale up salvo", "budget", budget, "err", err)

			break
		}

		newNodes, err := a.addLatestScaleUpResultsToClusterSnapshot(ctx, i, scaleUpStatus, handledPods, templateNodeInfos)
		if err != nil {
			logger.Info("Failed to update cluster snapshot after scale up, finishing the scale up salvo", "err", err)
			break
		}
		nodes = append(nodes, newNodes...)
	}
	logger.Info("Finished scale up salvo", "iterationCount", i, "unschedulablePodsCount", len(podsMap))
	return scaleUpStatus, typedErr
}

func (a *StaticAutoscaler) updateSoftDeletionTaints(ctx context.Context, allNodes []*apiv1.Node) {
	if a.AutoscalingContext.AutoscalingOptions.MaxBulkSoftTaintCount != 0 {
		taintableNodes := retrieveNodes(a.scaleDownPlanner.UnneededNodes())

		// Make sure we are only cleaning taints from selected node groups.
		selectedNodes := filterNodesFromSelectedGroups(ctx, a.CloudProvider, allNodes...)

		// This is a sanity check to make sure `taintableNodes` only includes
		// nodes from selected nodes.
		taintableNodes = intersectNodes(selectedNodes, taintableNodes)
		untaintableNodes := subtractNodes(selectedNodes, taintableNodes)
		actuation.UpdateSoftDeletionTaints(ctx, a.AutoscalingContext, taintableNodes, untaintableNodes)
	}
}

func (a *StaticAutoscaler) scaleDown(ctx context.Context, currentTime time.Time, allNodes []*apiv1.Node, scaleDownActuationStatus scaledown.ActuationStatus, scaleDownStatus *scaledownstatus.ScaleDownStatus) caerrors.AutoscalerError {

	logger := klog.FromContext(ctx)
	unneededStart := time.Now()
	logger.V(4).Info("Calculating unneeded nodes")

	var scaleDownCandidates []*apiv1.Node
	var podDestinations []*apiv1.Node

	// podDestinations and scaleDownCandidates are initialized based on allNodes variable, which contains only
	// registered nodes in cluster.
	// It does not include any upcoming nodes which can be part of clusterSnapshot. As an alternative to using
	// allNodes here, we could use nodes from clusterSnapshot and explicitly filter out upcoming nodes here but it
	// is of little (if any) benefit.

	if a.processors == nil || a.processors.ScaleDownNodeProcessor == nil {
		scaleDownCandidates = allNodes
		podDestinations = allNodes
	} else {
		var err caerrors.AutoscalerError
		scaleDownCandidates, err = a.processors.ScaleDownNodeProcessor.GetScaleDownCandidates(ctx, a.AutoscalingContext, allNodes)
		if err != nil {
			logger.Error(err, "Failed to get scale down candidates")
			return err
		}
		podDestinations, err = a.processors.ScaleDownNodeProcessor.GetPodDestinationCandidates(a.AutoscalingContext, allNodes)
		if err != nil {
			logger.Error(err, "Failed to get pod destination candidates")
			return err
		}
	}

	typedErr := a.scaleDownPlanner.UpdateClusterState(ctx, podDestinations, scaleDownCandidates, scaleDownActuationStatus, currentTime)
	// Update clusterStateRegistry and metrics regardless of whether ScaleDown was successful or not.
	unneededNodes := a.scaleDownPlanner.UnneededNodes()
	a.processors.ScaleDownCandidatesNotifier.Update(ctx, unneededNodes, currentTime)
	metrics.UpdateUnneededNodesCount(len(unneededNodes))
	if typedErr != nil {
		scaleDownStatus.Result = scaledownstatus.ScaleDownError
		logger.Error(typedErr, "Failed to scale down")
		return typedErr
	}

	metrics.UpdateDurationFromStart(ctx, metrics.FindUnneeded, unneededStart)

	scaleDownInCooldown := a.isScaleDownInCooldown(currentTime)
	logger.V(4).Info("Recording scale down status", "lastScaleUpTime", a.lastScaleUpTime, "lastScaleDownDeleteTime", a.lastScaleDownDeleteTime, "lastScaleDownFailTime", a.lastScaleDownFailTime, "scaleDownForbidden", a.processorCallbacks.disableScaleDownForLoop, "scaleDownInCooldown", scaleDownInCooldown)
	metrics.UpdateScaleDownInCooldown(scaleDownInCooldown)
	// We want to delete unneeded Node Groups only if here is no current delete
	// in progress.
	_, drained := scaleDownActuationStatus.DeletionsInProgress()
	var removedNodeGroups []cloudprovider.NodeGroup
	if len(drained) == 0 {
		var err error
		removedNodeGroups, err = a.processors.NodeGroupManager.RemoveUnneededNodeGroups(ctx, a.AutoscalingContext)
		if err != nil {
			logger.Error(err, "Error while removing unneeded node groups")
		}
		scaleDownStatus.RemovedNodeGroups = removedNodeGroups
	}

	if scaleDownInCooldown {
		scaleDownStatus.Result = scaledownstatus.ScaleDownInCooldown
		a.updateSoftDeletionTaints(ctx, allNodes)
	} else if len(scaleDownCandidates) == 0 {
		logger.V(4).Info("Starting scale down: no scale down candidates. skipping...")
		scaleDownStatus.Result = scaledownstatus.ScaleDownNoCandidates
		metrics.UpdateLastTime(metrics.ScaleDown, time.Now())
		a.updateSoftDeletionTaints(ctx, allNodes)
	} else {
		logger.V(4).Info("Starting scale down")

		scaleDownStart := time.Now()
		metrics.UpdateLastTime(metrics.ScaleDown, scaleDownStart)
		empty, needDrain := a.scaleDownPlanner.NodesToDelete(ctx, currentTime)
		scaleDownResult, scaledDownNodes, typedErr := a.scaleDownActuator.StartDeletion(ctx, empty, needDrain)
		scaleDownStatus.Result = scaleDownResult
		scaleDownStatus.ScaledDownNodes = scaledDownNodes
		metrics.UpdateDurationFromStart(ctx, metrics.ScaleDown, scaleDownStart)
		metrics.UpdateUnremovableNodesCount(countsByReason(a.scaleDownPlanner.UnremovableNodes()))

		scaleDownStatus.RemovedNodeGroups = removedNodeGroups

		if scaleDownStatus.Result == scaledownstatus.ScaleDownNodeDeleteStarted {
			a.lastScaleDownDeleteTime = currentTime
			a.clusterStateRegistry.Recalculate(ctx)
		}
		a.updateSoftDeletionTaints(ctx, allNodes)
		if typedErr != nil {
			logger.Error(typedErr, "Failed to scale down")
			a.lastScaleDownFailTime = currentTime
			return typedErr
		}
	}
	return nil
}

// addUpcomingNodesToClusterSnapshot generates upcoming node infos based on upcomingCounts and adds them to the ClusterSnapshot.
func (a *StaticAutoscaler) addUpcomingNodesToClusterSnapshot(
	ctx context.Context,
	upcomingCounts map[string]int,
	templateNodeInfos map[string]*framework.NodeInfo,
	suffixFmt string,
) ([]*apiv1.Node, error) {
	logger := klog.FromContext(ctx)
	upcomingNodeInfosPerNg, err := getUpcomingNodeInfos(ctx, upcomingCounts, templateNodeInfos, suffixFmt)
	if err != nil {
		return nil, err
	}

	nodeGroups := a.nodeGroupsById(ctx)
	upcomingNodeGroups := make(map[string]int)
	upcomingNodesFromUpcomingNodeGroups := 0
	var newNodes []*apiv1.Node

	for nodeGroupName, upcomingNodeInfos := range upcomingNodeInfosPerNg {
		nodeGroup := nodeGroups[nodeGroupName]
		if nodeGroup == nil {
			return nil, fmt.Errorf("failed to find node group: %s", nodeGroupName)
		}
		for i, upcomingNodeInfo := range upcomingNodeInfos {
			if err := a.ClusterSnapshot.AddNodeInfo(upcomingNodeInfo); err != nil {
				return nil, fmt.Errorf("Failed to add upcoming %d/%d node %q from node group %q to cluster snapshot: %w", i+1, len(upcomingNodeInfos), upcomingNodeInfo.Node().Name, nodeGroupName, err)
			}
			newNodes = append(newNodes, upcomingNodeInfo.Node())
		}
		if a.processors.AsyncNodeGroupStateChecker.IsUpcoming(nodeGroup) {
			upcomingNodesFromUpcomingNodeGroups += len(upcomingNodeInfos)
			upcomingNodeGroups[nodeGroup.Id()] += len(upcomingNodeInfos)
		}
	}
	if len(upcomingNodeGroups) > 0 {
		logger.Info("Injecting upcoming node groups upcoming nodes", "nodeGroupsCount", len(upcomingNodeGroups), "nodes", upcomingNodesFromUpcomingNodeGroups, "nodeGroups", upcomingNodeGroups)
	}
	return newNodes, nil
}

// addLatestScaleUpResultsToClusterSnapshot updates the ClusterSnapshot with upcoming nodes created in the latest scale up to prepare the state for the next scale up:
//   - adds upcoming nodeInfos from the latest scale up to ClusterSnapshot
//   - schedules the Pods that triggered the scale up on the latest nodeInfos
//   - returns the new nodes
func (a *StaticAutoscaler) addLatestScaleUpResultsToClusterSnapshot(ctx context.Context, idx int, scaleUpStatus *status.ScaleUpStatus, handledPods []*apiv1.Pod, templateNodeInfos map[string]*framework.NodeInfo) ([]*apiv1.Node, error) {
	logger := klog.FromContext(ctx)
	salvoSuffix := fmt.Sprintf("salvo-%d", idx)
	upcomingCounts := make(map[string]int)

	for _, suInfo := range scaleUpStatus.ScaleUpInfos {
		nodesToInject := suInfo.NewSize - suInfo.CurrentSize
		if nodesToInject <= 0 {
			return nil, fmt.Errorf("Scale up salvo %d contains scale up info for node group %q with non-positive number of nodes to inject: %d. Current size: %d, new size: %d", idx, suInfo.Group.Id(), nodesToInject, suInfo.CurrentSize, suInfo.NewSize)
		}

		if _, ok := templateNodeInfos[suInfo.Group.Id()]; !ok {
			return nil, fmt.Errorf("Failed to find template node info for node group %q", suInfo.Group.Id())
		}
		upcomingCounts[suInfo.Group.Id()] += nodesToInject
	}

	newNodes, err := a.addUpcomingNodesToClusterSnapshot(ctx, upcomingCounts, templateNodeInfos, "%d-"+salvoSuffix)
	if err != nil {
		return nil, fmt.Errorf("Failed to get upcoming node infos for salvo %d: %v", idx, err)
	}

	for _, pod := range handledPods {
		nodeName, err := a.ClusterSnapshot.SchedulePodOnAnyNodeMatching(pod, clustersnapshot.SchedulingOptions{
			IsNodeAcceptable: func(nodeInfo *framework.NodeInfo) bool {
				return strings.HasSuffix(nodeInfo.Node().Name, salvoSuffix)
			},
		})
		if err != nil {
			nodeGroupIds := []string{}
			for _, suInfo := range scaleUpStatus.ScaleUpInfos {
				nodeGroupIds = append(nodeGroupIds, suInfo.Group.Id())
			}
			return nil, fmt.Errorf("Failed cluster snapshot update: couldn't schedule triggering pod %s on any of the nodes from scaled up node group(s) %v: %v", pod.Name, nodeGroupIds, err)
		}
		logger.V(5).Info("Updated cluster snapshot: scheduled pod on node", "pod", klog.KObj(pod), "nodeName", nodeName)
	}

	return newNodes, nil
}

func (a *StaticAutoscaler) isScaleDownInCooldown(currentTime time.Time) bool {
	scaleDownInCooldown := a.processorCallbacks.disableScaleDownForLoop

	if a.ScaleDownDelayTypeLocal {
		return scaleDownInCooldown
	}
	return scaleDownInCooldown ||
		a.lastScaleUpTime.Add(a.ScaleDownDelayAfterAdd).After(currentTime) ||
		a.lastScaleDownFailTime.Add(a.ScaleDownDelayAfterFailure).After(currentTime) ||
		a.lastScaleDownDeleteTime.Add(a.ScaleDownDelayAfterDelete).After(currentTime)
}

// Sets the target size of node groups to the current number of nodes in them
// if the difference was constant for a prolonged time. Returns true if managed
// to fix something.
func fixNodeGroupSize(ctx context.Context, autoscalingCtx *ca_context.AutoscalingContext, clusterStateRegistry *clusterstate.ClusterStateRegistry, currentTime time.Time) (bool, error) {
	logger := klog.FromContext(ctx)
	fixed := false
	for _, nodeGroup := range autoscalingCtx.CloudProvider.NodeGroups(ctx) {
		incorrectSize := clusterStateRegistry.GetIncorrectNodeGroupSize(nodeGroup.Id())
		if incorrectSize == nil {
			continue
		}
		maxNodeProvisionTime, err := clusterStateRegistry.MaxNodeProvisionTime(ctx, nodeGroup)
		if err != nil {
			return false, fmt.Errorf("failed to retrieve maxNodeProvisionTime for nodeGroup %s", nodeGroup.Id())
		}
		if incorrectSize.FirstObserved.Add(maxNodeProvisionTime).Before(currentTime) {
			delta := incorrectSize.CurrentSize - incorrectSize.ExpectedSize
			if delta < 0 {
				logger.V(0).Info("Decreasing size", "nodeGroupId", nodeGroup.Id(), "targetSize", incorrectSize.ExpectedSize, "size", incorrectSize.CurrentSize, "delta", delta)
				if err := nodeGroup.DecreaseTargetSize(ctx, delta); err != nil {
					return fixed, fmt.Errorf("failed to decrease %s: %v", nodeGroup.Id(), err)
				}
				fixed = true
			}
		}
	}
	return fixed, nil
}

// removeOldUnregisteredNodes removes unregistered nodes if needed. Returns true
// if anything was removed and error if such occurred.
func (a *StaticAutoscaler) removeOldUnregisteredNodes(ctx context.Context, allUnregisteredNodes []clusterstate.UnregisteredNode,
	csr *clusterstate.ClusterStateRegistry, currentTime time.Time, logRecorder *utils.LogEventRecorder) (bool, error) {

	logger := klog.FromContext(ctx)
	unregisteredNodesToRemove, err := a.oldUnregisteredNodes(ctx, allUnregisteredNodes, csr, currentTime)
	if err != nil {
		return false, err
	}

	nodeGroups := a.nodeGroupsById(ctx)
	removedAny := false
	for nodeGroupId, unregisteredNodesToDelete := range unregisteredNodesToRemove {
		nodeGroup := nodeGroups[nodeGroupId]
		logger.V(0).Info("Removing unregistered nodes for node group", "nodesCount", len(unregisteredNodesToDelete), "nodeGroupId", nodeGroupId)
		if !a.ForceDeleteLongUnregisteredNodes {
			size, err := nodeGroup.TargetSize(ctx)
			if err != nil {
				logger.Info("Failed to get node group size", "nodeGroupId", nodeGroup.Id(), "err", err)
				continue
			}
			possibleToDelete := size - nodeGroup.MinSize(ctx)
			if possibleToDelete <= 0 {
				logger.Info("Node group min size reached, skipping removal of unregistered nodes", "nodeGroupId", nodeGroupId, "nodesCount", len(unregisteredNodesToDelete))
				continue
			}
			if len(unregisteredNodesToDelete) > possibleToDelete {
				logger.Info("Capping number of unregistered nodes to be removed from node group, removing all nodes would exceed min size constaint", "nodeGroupId", nodeGroupId, "nodesToDeleteCount", possibleToDelete, "allNodesCount", len(unregisteredNodesToDelete))
				unregisteredNodesToDelete = unregisteredNodesToDelete[:possibleToDelete]
			}
		}

		nodesToDelete := toNodes(unregisteredNodesToDelete)
		nodesToDelete, err := overrideNodesToDeleteForZeroOrMax(ctx, a.NodeGroupDefaults, nodeGroup, nodesToDelete)
		if err != nil {
			logger.Info("Failed to remove unregistered nodes from node group", "nodeGroupId", nodeGroupId, "err", err)
			continue
		}

		if len(nodesToDelete) == 0 {
			continue
		}

		if a.ForceDeleteLongUnregisteredNodes {
			err = nodeGroup.ForceDeleteNodes(ctx, nodesToDelete)
			if err == cloudprovider.ErrNotImplemented {
				err = nodeGroup.DeleteNodes(ctx, nodesToDelete)
			}
		} else {
			err = nodeGroup.DeleteNodes(ctx, nodesToDelete)
		}
		csr.InvalidateNodeInstancesCacheEntry(ctx, nodeGroup)
		if err != nil {
			logger.Info("Failed to remove unregistered nodes from node group", "nodesToDeleteCount", len(nodesToDelete), "nodeGroupId", nodeGroupId, "err", err)
			for _, node := range nodesToDelete {
				logRecorder.Eventf(apiv1.EventTypeWarning, "DeleteUnregisteredFailed",
					"Failed to remove node %s: %v", node.Name, err)
			}
			return removedAny, err
		}
		for _, node := range nodesToDelete {
			logRecorder.Eventf(apiv1.EventTypeNormal, "DeleteUnregistered",
				"Removed unregistered node %v", node.Name)
		}
		metrics.RegisterOldUnregisteredNodesRemoved(len(nodesToDelete))
		removedAny = true
	}
	return removedAny, nil
}

// oldUnregisteredNodes returns old unregistered nodes grouped by their node group id.
func (a *StaticAutoscaler) oldUnregisteredNodes(ctx context.Context, allUnregisteredNodes []clusterstate.UnregisteredNode, csr *clusterstate.ClusterStateRegistry, currentTime time.Time) (map[string][]clusterstate.UnregisteredNode, error) {
	logger := klog.FromContext(ctx)
	nodesByNodeGroupId := make(map[string][]clusterstate.UnregisteredNode)
	for _, unregisteredNode := range allUnregisteredNodes {
		nodeGroup, err := a.CloudProvider.NodeGroupForNode(ctx, unregisteredNode.Node)
		if err != nil {
			logger.Info("Failed to get node group", "node", klog.KObj(unregisteredNode.Node), "err", err)
			continue
		}
		if nodeGroup == nil {
			logger.Info("No node group for node, skipping", "node", klog.KObj(unregisteredNode.Node))
			continue
		}

		maxNodeProvisionTime, err := csr.MaxNodeProvisionTime(ctx, nodeGroup)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve maxNodeProvisionTime for node %s in nodeGroup %s", unregisteredNode.Node.Name, nodeGroup.Id())
		}

		if unregisteredNode.UnregisteredSince.Add(maxNodeProvisionTime).Before(currentTime) {
			logger.V(0).Info("Marking unregistered node for removal", "node", klog.KObj(unregisteredNode.Node))
			nodesByNodeGroupId[nodeGroup.Id()] = append(nodesByNodeGroupId[nodeGroup.Id()], unregisteredNode)
		}
	}

	return nodesByNodeGroupId, nil
}

func toNodes(unregisteredNodes []clusterstate.UnregisteredNode) []*apiv1.Node {
	nodes := []*apiv1.Node{}
	for _, n := range unregisteredNodes {
		nodes = append(nodes, n.Node)
	}
	return nodes
}

func (a *StaticAutoscaler) deleteCreatedNodesWithErrors(ctx context.Context) {
	// We always schedule deleting of incoming errornous nodes
	// TODO[lukaszos] Consider adding logic to not retry delete every loop iteration
	logger := klog.FromContext(ctx)
	nodeGroups := a.nodeGroupsById(ctx)
	nodesToDeleteByNodeGroupId := a.clusterStateRegistry.GetCreatedNodesWithErrors()

	deletedAny := false

	for nodeGroupId, nodesToDelete := range nodesToDeleteByNodeGroupId {
		var err error
		logger.V(1).Info("Deleting node group because of create errors", "nodesToDeleteCount", len(nodesToDelete), "nodeGroupId", nodeGroupId)

		nodeGroup := nodeGroups[nodeGroupId]
		if nodeGroup == nil {
			err = fmt.Errorf("node group %s not found", nodeGroupId)
		} else if nodesToDelete, err = overrideNodesToDeleteForZeroOrMax(ctx, a.NodeGroupDefaults, nodeGroup, nodesToDelete); err == nil && len(nodesToDelete) > 0 {
			if a.ForceDeleteFailedNodes {
				err = nodeGroup.ForceDeleteNodes(ctx, nodesToDelete)
				if errors.Is(err, cloudprovider.ErrNotImplemented) {
					err = nodeGroup.DeleteNodes(ctx, nodesToDelete)
				}
			} else {
				err = nodeGroup.DeleteNodes(ctx, nodesToDelete)
			}
		}

		if err != nil {
			logger.Info("Error while trying to delete nodes", "nodeGroupId", nodeGroupId, "err", err)
		} else if len(nodesToDelete) > 0 {
			deletedAny = true
			a.clusterStateRegistry.InvalidateNodeInstancesCacheEntry(ctx, nodeGroup)
		}
	}

	if deletedAny {
		logger.V(0).Info("Some nodes that failed to create were removed, recalculating cluster state.")
		a.clusterStateRegistry.Recalculate(ctx)
	}
}

// overrideNodesToDeleteForZeroOrMax returns a list of nodes to delete, taking into account that
// node deletion for a "ZeroOrMaxNodeScaling" should either keep or remove all the nodes.
// For a non-"ZeroOrMaxNodeScaling" node group it returns the unchanged list of nodes to delete.
func overrideNodesToDeleteForZeroOrMax(ctx context.Context, defaults config.NodeGroupAutoscalingOptions, nodeGroup cloudprovider.NodeGroup, nodesToDelete []*apiv1.Node) ([]*apiv1.Node, error) {
	opts, err := nodeGroup.GetOptions(ctx, defaults)
	if err != nil && err != cloudprovider.ErrNotImplemented {
		return []*apiv1.Node{}, fmt.Errorf("Failed to get node group options for %s: %s", nodeGroup.Id(), err)
	}
	// If a scale-up of "ZeroOrMaxNodeScaling" node group failed, the cleanup
	// node deletion for a "ZeroOrMaxNodeScaling" node group is atomic and should delete all nodes or none.
	if opts != nil && opts.ZeroOrMaxNodeScaling {
		instances, err := nodeGroup.Nodes(ctx)
		if err != nil {
			return []*apiv1.Node{}, fmt.Errorf("Failed to fill in nodes to delete from group %s based on ZeroOrMaxNodeScaling option: %s", nodeGroup.Id(), err)
		}

		// Remove all nodes in case when either:
		// 1. All nodes are failing
		// 2. AllowNonAtomicScaleUpToMax is false which means we want to atomically remove partially failed node groups
		if len(instances) == len(nodesToDelete) || !opts.AllowNonAtomicScaleUpToMax {
			// Remove all nodes
			return instancesToFakeNodes(instances), nil
		}
		return []*apiv1.Node{}, nil
	}
	// No override needed.
	return nodesToDelete, nil
}

// instancesToNodes returns a list of fake nodes with just names populated,
// so that they can be passed as nodes to delete
func instancesToFakeNodes(instances []cloudprovider.Instance) []*apiv1.Node {
	nodes := []*apiv1.Node{}
	for _, i := range instances {
		nodes = append(nodes, clusterstate.FakeNode(i, ""))
	}
	return nodes
}

func (a *StaticAutoscaler) nodeGroupsById(ctx context.Context) map[string]cloudprovider.NodeGroup {
	nodeGroups := make(map[string]cloudprovider.NodeGroup)
	for _, nodeGroup := range a.CloudProvider.NodeGroups(ctx) {
		nodeGroups[nodeGroup.Id()] = nodeGroup
	}
	return nodeGroups
}

// Don't consider pods newer than newPodScaleUpDelay or annotated podScaleUpDelay
// seconds old as unschedulable.
func (a *StaticAutoscaler) filterOutYoungPods(ctx context.Context, allUnschedulablePods []*apiv1.Pod, currentTime time.Time) []*apiv1.Pod {
	logger := klog.FromContext(ctx)
	var oldUnschedulablePods []*apiv1.Pod
	newPodScaleUpDelay := a.AutoscalingOptions.NewPodScaleUpDelay
	for _, pod := range allUnschedulablePods {
		podAge := currentTime.Sub(pod.CreationTimestamp.Time)
		podScaleUpDelay := newPodScaleUpDelay

		if podScaleUpDelayAnnotationStr, ok := pod.Annotations[annotations.PodScaleUpDelayAnnotationKey]; ok {
			podScaleUpDelayAnnotation, err := time.ParseDuration(podScaleUpDelayAnnotationStr)
			if err != nil {
				logger.Error(err, "Failed to parse pod annotation", "pod", klog.KObj(pod), "annotationKey", annotations.PodScaleUpDelayAnnotationKey)
			} else {
				if podScaleUpDelayAnnotation < podScaleUpDelay {
					logger.Error(nil, "Failed to set pod scale up delay for pod through annotation: podScaleUpDelayAnnotation is less than newPodScaleUpDelay", "pod", klog.KObj(pod), "annotationKey", annotations.PodScaleUpDelayAnnotationKey, "podScaleUpDelayAnnotation", podScaleUpDelayAnnotation, "delay", newPodScaleUpDelay)
				} else {
					podScaleUpDelay = podScaleUpDelayAnnotation
				}
			}
		}

		if podAge > podScaleUpDelay {
			oldUnschedulablePods = append(oldUnschedulablePods, pod)
		} else {
			logger.V(3).Info("Pod is too new to consider unschedulable", "pod", klog.KObj(pod), "podAgeSeconds", podAge.Seconds())
		}
	}
	return oldUnschedulablePods
}

// ExitCleanUp performs all necessary clean-ups when the autoscaler's exiting.
func (a *StaticAutoscaler) ExitCleanUp() {
	a.processors.CleanUp()
	a.DebuggingSnapshotter.Cleanup()

	if !a.AutoscalingContext.WriteStatusConfigMap {
		return
	}
	utils.DeleteStatusConfigMap(a.AutoscalingContext.ClientSet, a.AutoscalingContext.ConfigNamespace, a.AutoscalingContext.StatusConfigMapName)

	a.CloudProvider.Cleanup(context.TODO())

	a.clusterStateRegistry.Stop()
}

func (a *StaticAutoscaler) obtainNodeLists(ctx context.Context, draSnapshot *drasnapshot.Snapshot, csiSnapshot *csisnapshot.Snapshot) ([]*apiv1.Node, []*apiv1.Node, caerrors.AutoscalerError) {
	logger := klog.FromContext(ctx)
	allNodes, err := a.AllNodeLister().List()
	if err != nil {
		logger.Error(err, "Failed to list all nodes")
		return nil, nil, caerrors.ToAutoscalerError(caerrors.ApiCallError, err)
	}
	readyNodes, err := a.ReadyNodeLister().List()
	if err != nil {
		logger.Error(err, "Failed to list ready nodes")
		return nil, nil, caerrors.ToAutoscalerError(caerrors.ApiCallError, err)
	}
	a.reportTaintsCount(allNodes)

	// Handle GPU case - allocatable GPU may be equal to 0 up to 15 minutes after
	// node registers as ready. See https://github.com/kubernetes/kubernetes/issues/54959
	// Treat those nodes as unready until GPU actually becomes available and let
	// our normal handling for booting up nodes deal with this.
	// TODO: Remove this call when we handle dynamically provisioned resources.
	allNodes, readyNodes = a.processors.CustomResourcesProcessor.FilterOutNodesWithUnreadyResources(ctx, a.AutoscalingContext, allNodes, readyNodes, draSnapshot, csiSnapshot)
	allNodes, readyNodes = taints.FilterOutNodesWithStartupTaints(ctx, a.taintConfig, allNodes, readyNodes)
	return allNodes, readyNodes, nil
}

func filterNodesFromSelectedGroups(ctx context.Context, cp cloudprovider.CloudProvider, nodes ...*apiv1.Node) []*apiv1.Node {
	logger := klog.FromContext(ctx)
	filtered := make([]*apiv1.Node, 0, len(nodes))
	for _, n := range nodes {
		if ng, err := cp.NodeGroupForNode(ctx, n); err != nil {
			logger.Error(err, "Failed to get a nodegroup for node", "node", klog.KObj(n))
		} else if ng != nil {
			filtered = append(filtered, n)
		}
	}
	return filtered
}

func (a *StaticAutoscaler) updateClusterState(ctx context.Context, allNodes []*apiv1.Node, currentTime time.Time) caerrors.AutoscalerError {
	logger := klog.FromContext(ctx)
	err := a.clusterStateRegistry.UpdateNodes(ctx, allNodes, currentTime)
	if err != nil {
		logger.Error(err, "Failed to update node registry")
		a.scaleDownPlanner.CleanUpUnneededNodes(ctx)
		return caerrors.ToAutoscalerError(caerrors.CloudProviderError, err)
	}
	core_utils.UpdateClusterStateMetrics(a.clusterStateRegistry)

	return nil
}

func (a *StaticAutoscaler) reportTaintsCount(nodes []*apiv1.Node) {
	foundTaints := taints.CountNodeTaints(nodes, a.taintConfig)
	for taintType, count := range foundTaints {
		metrics.ObserveNodeTaintsCount(taintType, float64(count))
	}
}

func allPodsAreNew(pods []*apiv1.Pod, currentTime time.Time) bool {
	if core_utils.GetOldestCreateTime(pods).Add(unschedulablePodTimeBuffer).After(currentTime) {
		return true
	}
	found, oldest := core_utils.GetOldestCreateTimeWithGpu(pods)
	return found && oldest.Add(unschedulablePodWithGpuTimeBuffer).After(currentTime)
}

func getUpcomingNodeInfos(ctx context.Context, upcomingCounts map[string]int, nodeInfos map[string]*framework.NodeInfo, suffixFmt string) (map[string][]*framework.NodeInfo, error) {
	logger := klog.FromContext(ctx)
	upcomingNodes := make(map[string][]*framework.NodeInfo)
	for nodeGroup, numberOfNodes := range upcomingCounts {
		nodeTemplate, found := nodeInfos[nodeGroup]
		if !found {
			logger.Info("Couldn't find template for node group", "nodeGroupId", nodeGroup)
			continue
		}

		var nodes []*framework.NodeInfo
		for i := 0; i < numberOfNodes; i++ {
			// Ensure new nodes have different names because nodeName
			// will be used as a map key. Also deep copy pods (daemonsets &
			// any pods added by cloud provider on template).
			freshNodeInfo, err := simulator.SanitizedNodeInfo(ctx, nodeTemplate, fmt.Sprintf(suffixFmt, i))
			if err != nil {
				return nil, err
			}
			if freshNodeInfo.Node().Annotations == nil {
				freshNodeInfo.Node().Annotations = make(map[string]string)
			}
			freshNodeInfo.Node().Annotations[annotations.NodeUpcomingAnnotation] = "true"

			nodes = append(nodes, freshNodeInfo)
		}
		upcomingNodes[nodeGroup] = nodes
	}
	return upcomingNodes, nil
}

func calculateCoresMemoryTotal(nodes []*apiv1.Node, timestamp time.Time) (int64, int64) {
	// this function is essentially similar to the calculateScaleDownCoresMemoryTotal
	// we want to check all nodes, aside from those deleting, to sum the cluster resource usage.
	var coresTotal, memoryTotal int64
	for _, node := range nodes {
		if actuation.IsNodeBeingDeleted(node, timestamp) {
			// Nodes being deleted do not count towards total cluster resources
			continue
		}
		cores, memory := core_utils.GetNodeCoresAndMemory(node)

		coresTotal += cores
		memoryTotal += memory
	}

	return coresTotal, memoryTotal
}

func countsByReason(nodes []*simulator.UnremovableNode) map[simulator.UnremovableReason]int {
	counts := make(map[simulator.UnremovableReason]int)

	for _, node := range nodes {
		counts[node.Reason]++
	}

	return counts
}

func subtractNodesByName(nodes []*apiv1.Node, namesToRemove []string) []*apiv1.Node {
	var c []*apiv1.Node
	removeSet := make(map[string]bool)
	for _, name := range namesToRemove {
		removeSet[name] = true
	}
	for _, n := range nodes {
		if removeSet[n.Name] {
			continue
		}
		c = append(c, n)
	}
	return c
}

func subtractNodes(a []*apiv1.Node, b []*apiv1.Node) []*apiv1.Node {
	return subtractNodesByName(a, nodeNames(b))
}

func filterNodesByName(nodes []*apiv1.Node, names []string) []*apiv1.Node {
	c := make([]*apiv1.Node, 0, len(names))
	filterSet := make(map[string]bool, len(names))
	for _, name := range names {
		filterSet[name] = true
	}
	for _, n := range nodes {
		if filterSet[n.Name] {
			c = append(c, n)
		}
	}
	return c
}

// intersectNodes gives intersection of 2 node lists
func intersectNodes(a []*apiv1.Node, b []*apiv1.Node) []*apiv1.Node {
	return filterNodesByName(a, nodeNames(b))
}

func nodeNames(ns []*apiv1.Node) []string {
	names := make([]string, len(ns))
	for i, node := range ns {
		names[i] = node.Name
	}
	return names
}

func retrieveNodes(candidates []*scaledown.UnneededNode) []*apiv1.Node {
	nodes := make([]*apiv1.Node, 0, len(candidates))
	for _, c := range candidates {
		nodes = append(nodes, c.Node)
	}
	return nodes
}

func listPods(ctx context.Context, podLister kube_util.PodLister, bypassedSchedulers, allowedSchedulers map[string]bool) (podsBySchedulability kube_util.PodsBySchedulability, err error) {
	logger := klog.FromContext(ctx)
	pods, err := podLister.List()
	if err != nil {
		logger.Error(err, "Failed to list pods")
		return podsBySchedulability, err
	}
	initialPodCount := len(pods)
	if len(allowedSchedulers) > 0 {
		pods = kube_util.FilterOutPodsByScheduler(pods, allowedSchedulers)
	}
	podsBySchedulability = kube_util.ArrangePodsBySchedulability(pods, bypassedSchedulers)
	// Skip logging in case of the boring scenario, when all pods are scheduled.
	if len(pods) != len(podsBySchedulability.Scheduled) {
		ignoredDueToDisallowed := initialPodCount - len(pods)
		ignored := len(pods) - len(podsBySchedulability.Scheduled) - len(podsBySchedulability.NominatedNode) - len(podsBySchedulability.Unschedulable) - len(podsBySchedulability.Unprocessed)
		logger.Info("Found pods in the cluster", "podsCount", initialPodCount, "scheduledPodsCount", len(podsBySchedulability.Scheduled), "nominatedPodsCount", len(podsBySchedulability.NominatedNode), "unschedulablePodsCount", len(podsBySchedulability.Unschedulable), "unprocessedPodsCount", len(podsBySchedulability.Unprocessed), "ignoredByAllowedSchedulersPodsCount", ignored, "ignoredDueToDisallowedSchedulersPodsCount", ignoredDueToDisallowed)
	}
	return
}

func (a *StaticAutoscaler) shouldScaleDown() bool {
	if !a.ScaleDownEnabled {
		return false
	}
	// When GracefulDegradationMode is active, autoscaler does scale down only on every 4th(gracefulDegradationCycle th) loop.
	return !(a.AutoscalingContext.GracefulDegradationEnabled && a.AutoscalingContext.GracefulDegradationLoopCount > 0 && a.AutoscalingContext.GracefulDegradationLoopCount%gracefulDegradationCycle != 0)
}
