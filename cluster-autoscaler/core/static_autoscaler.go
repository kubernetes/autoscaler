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
	"errors"
	"fmt"
	"reflect"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate/utils"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/actuation"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/deletiontracker"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/pdb"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/planner"
	scaledownstatus "k8s.io/autoscaler/cluster-autoscaler/core/scaledown/status"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaleup"
	orchestrator "k8s.io/autoscaler/cluster-autoscaler/core/scaleup/orchestrator"
	core_utils "k8s.io/autoscaler/cluster-autoscaler/core/utils"
	"k8s.io/autoscaler/cluster-autoscaler/debuggingsnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/estimator"
	"k8s.io/autoscaler/cluster-autoscaler/expander"
	"k8s.io/autoscaler/cluster-autoscaler/metrics"
	"k8s.io/autoscaler/cluster-autoscaler/observers/loopstart"
	ca_processors "k8s.io/autoscaler/cluster-autoscaler/processors"
	"k8s.io/autoscaler/cluster-autoscaler/processors/status"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/drainability/rules"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/options"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/predicatechecker"
	"k8s.io/autoscaler/cluster-autoscaler/utils/backoff"
	caerrors "k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	scheduler_utils "k8s.io/autoscaler/cluster-autoscaler/utils/scheduler"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"
	"k8s.io/utils/integer"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	klog "k8s.io/klog/v2"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
)

const (
	// How old the oldest unschedulable pod should be before starting scale up.
	unschedulablePodTimeBuffer = 2 * time.Second
	// How old the oldest unschedulable pod with GPU should be before starting scale up.
	// The idea is that nodes with GPU are very expensive and we're ready to sacrifice
	// a bit more latency to wait for more pods and make a more informed scale-up decision.
	unschedulablePodWithGpuTimeBuffer = 30 * time.Second

	// NodeUpcomingAnnotation is an annotation CA adds to nodes which are upcoming.
	NodeUpcomingAnnotation = "cluster-autoscaler.k8s.io/upcoming-node"

	// podScaleUpDelayAnnotationKey is an annotation how long pod can wait to be scaled up.
	podScaleUpDelayAnnotationKey = "cluster-autoscaler.kubernetes.io/pod-scale-up-delay"
)

// StaticAutoscaler is an autoscaler which has all the core functionality of a CA but without the reconfiguration feature
type StaticAutoscaler struct {
	// AutoscalingContext consists of validated settings and options for this autoscaler
	*context.AutoscalingContext
	// ClusterState for maintaining the state of cluster nodes.
	clusterStateRegistry    *clusterstate.ClusterStateRegistry
	lastScaleUpTime         time.Time
	lastScaleDownDeleteTime time.Time
	lastScaleDownFailTime   time.Time
	scaleDownPlanner        scaledown.Planner
	scaleDownActuator       scaledown.Actuator
	scaleUpOrchestrator     scaleup.Orchestrator
	processors              *ca_processors.AutoscalingProcessors
	loopStartNotifier       *loopstart.ObserversList
	processorCallbacks      *staticAutoscalerProcessorCallbacks
	initialized             bool
	taintConfig             taints.TaintConfig
}

type staticAutoscalerProcessorCallbacks struct {
	disableScaleDownForLoop bool
	extraValues             map[string]interface{}
	scaleDownPlanner        scaledown.Planner
}

func (callbacks *staticAutoscalerProcessorCallbacks) ResetUnneededNodes() {
	callbacks.scaleDownPlanner.CleanUpUnneededNodes()
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
	predicateChecker predicatechecker.PredicateChecker,
	clusterSnapshot clustersnapshot.ClusterSnapshot,
	autoscalingKubeClients *context.AutoscalingKubeClients,
	processors *ca_processors.AutoscalingProcessors,
	loopStartNotifier *loopstart.ObserversList,
	cloudProvider cloudprovider.CloudProvider,
	expanderStrategy expander.Strategy,
	estimatorBuilder estimator.EstimatorBuilder,
	backoff backoff.Backoff,
	debuggingSnapshotter debuggingsnapshot.DebuggingSnapshotter,
	remainingPdbTracker pdb.RemainingPdbTracker,
	scaleUpOrchestrator scaleup.Orchestrator,
	deleteOptions options.NodeDeleteOptions,
	drainabilityRules rules.Rules) *StaticAutoscaler {

	clusterStateConfig := clusterstate.ClusterStateRegistryConfig{
		MaxTotalUnreadyPercentage: opts.MaxTotalUnreadyPercentage,
		OkTotalUnreadyCount:       opts.OkTotalUnreadyCount,
	}
	clusterStateRegistry := clusterstate.NewClusterStateRegistry(cloudProvider, clusterStateConfig, autoscalingKubeClients.LogRecorder, backoff, processors.NodeGroupConfigProcessor, processors.AsyncNodeGroupStateChecker)
	processorCallbacks := newStaticAutoscalerProcessorCallbacks()
	autoscalingContext := context.NewAutoscalingContext(
		opts,
		predicateChecker,
		clusterSnapshot,
		autoscalingKubeClients,
		cloudProvider,
		expanderStrategy,
		processorCallbacks,
		debuggingSnapshotter,
		remainingPdbTracker,
		clusterStateRegistry)

	taintConfig := taints.NewTaintConfig(opts)
	processors.ScaleDownCandidatesNotifier.Register(clusterStateRegistry)
	processors.ScaleStateNotifier.Register(clusterStateRegistry)

	// TODO: Populate the ScaleDownActuator/Planner fields in AutoscalingContext
	// during the struct creation rather than here.
	scaleDownPlanner := planner.New(autoscalingContext, processors, deleteOptions, drainabilityRules)
	processorCallbacks.scaleDownPlanner = scaleDownPlanner

	ndt := deletiontracker.NewNodeDeletionTracker(0 * time.Second)
	scaleDownActuator := actuation.NewActuator(autoscalingContext, processors.ScaleStateNotifier, ndt, deleteOptions, drainabilityRules, processors.NodeGroupConfigProcessor)
	autoscalingContext.ScaleDownActuator = scaleDownActuator

	if scaleUpOrchestrator == nil {
		scaleUpOrchestrator = orchestrator.New()
	}
	scaleUpOrchestrator.Initialize(autoscalingContext, processors, clusterStateRegistry, estimatorBuilder, taintConfig)

	// Set the initial scale times to be less than the start time so as to
	// not start in cooldown mode.
	initialScaleTime := time.Now().Add(-time.Hour)
	return &StaticAutoscaler{
		AutoscalingContext:      autoscalingContext,
		lastScaleUpTime:         initialScaleTime,
		lastScaleDownDeleteTime: initialScaleTime,
		lastScaleDownFailTime:   initialScaleTime,
		scaleDownPlanner:        scaleDownPlanner,
		scaleDownActuator:       scaleDownActuator,
		scaleUpOrchestrator:     scaleUpOrchestrator,
		processors:              processors,
		loopStartNotifier:       loopStartNotifier,
		processorCallbacks:      processorCallbacks,
		clusterStateRegistry:    clusterStateRegistry,
		taintConfig:             taintConfig,
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
func (a *StaticAutoscaler) cleanUpIfRequired() {
	if a.initialized {
		return
	}

	// CA can die at any time. Removing taints that might have been left from the previous run.
	if allNodes, err := a.AllNodeLister().List(); err != nil {
		klog.Errorf("Failed to list ready nodes, not cleaning up taints: %v", err)
	} else {
		// Make sure we are only cleaning taints from selected node groups.
		selectedNodes := filterNodesFromSelectedGroups(a.CloudProvider, allNodes...)
		taints.CleanAllToBeDeleted(selectedNodes,
			a.AutoscalingContext.ClientSet, a.Recorder, a.CordonNodeBeforeTerminate)
		if a.AutoscalingContext.AutoscalingOptions.MaxBulkSoftTaintCount == 0 {
			// Clean old taints if soft taints handling is disabled
			taints.CleanAllDeletionCandidates(allNodes,
				a.AutoscalingContext.ClientSet, a.Recorder)
		}
	}
	a.initialized = true
}

func (a *StaticAutoscaler) initializeClusterSnapshot(nodes []*apiv1.Node, scheduledPods []*apiv1.Pod) caerrors.AutoscalerError {
	a.ClusterSnapshot.Clear()

	knownNodes := make(map[string]bool)
	for _, node := range nodes {
		if err := a.ClusterSnapshot.AddNode(node); err != nil {
			klog.Errorf("Failed to add node %s to cluster snapshot: %v", node.Name, err)
			return caerrors.ToAutoscalerError(caerrors.InternalError, err)
		}
		knownNodes[node.Name] = true
	}
	for _, pod := range scheduledPods {
		if knownNodes[pod.Spec.NodeName] {
			if err := a.ClusterSnapshot.AddPod(pod, pod.Spec.NodeName); err != nil {
				klog.Errorf("Failed to add pod %s scheduled to node %s to cluster snapshot: %v", pod.Name, pod.Spec.NodeName, err)
				return caerrors.ToAutoscalerError(caerrors.InternalError, err)
			}
		}
	}
	return nil
}

func (a *StaticAutoscaler) initializeRemainingPdbTracker() caerrors.AutoscalerError {
	a.RemainingPdbTracker.Clear()

	pdbs, err := a.PodDisruptionBudgetLister().List()
	if err != nil {
		klog.Errorf("Failed to list pod disruption budgets: %v", err)
		return caerrors.NewAutoscalerError(caerrors.ApiCallError, err.Error())
	}
	err = a.RemainingPdbTracker.SetPdbs(pdbs)
	if err != nil {
		return caerrors.NewAutoscalerError(caerrors.InternalError, err.Error())
	}
	return nil
}

// RunOnce iterates over node groups and scales them up/down if necessary
func (a *StaticAutoscaler) RunOnce(currentTime time.Time) caerrors.AutoscalerError {
	a.cleanUpIfRequired()
	a.processorCallbacks.reset()
	a.clusterStateRegistry.PeriodicCleanup()
	a.DebuggingSnapshotter.StartDataCollection()
	defer a.DebuggingSnapshotter.Flush()

	podLister := a.AllPodLister()
	autoscalingContext := a.AutoscalingContext

	klog.V(4).Info("Starting main loop")

	stateUpdateStart := time.Now()

	// Get nodes and pods currently living on cluster
	allNodes, readyNodes, typedErr := a.obtainNodeLists()
	if typedErr != nil {
		klog.Errorf("Failed to get node list: %v", typedErr)
		return typedErr
	}

	if abortLoop, err := a.processors.ActionableClusterProcessor.ShouldAbort(
		a.AutoscalingContext, allNodes, readyNodes, currentTime); abortLoop {
		return err
	}

	pods, err := podLister.List()
	if err != nil {
		klog.Errorf("Failed to list pods: %v", err)
		return caerrors.ToAutoscalerError(caerrors.ApiCallError, err)
	}
	originalScheduledPods, unschedulablePods := kube_util.ScheduledPods(pods), kube_util.UnschedulablePods(pods)
	schedulerUnprocessed := make([]*apiv1.Pod, 0, 0)
	isSchedulerProcessingIgnored := len(a.BypassedSchedulers) > 0
	if isSchedulerProcessingIgnored {
		schedulerUnprocessed = kube_util.SchedulerUnprocessedPods(pods, a.BypassedSchedulers)
	}

	// Update cluster resource usage metrics
	coresTotal, memoryTotal := calculateCoresMemoryTotal(allNodes, currentTime)
	metrics.UpdateClusterCPUCurrentCores(coresTotal)
	metrics.UpdateClusterMemoryCurrentBytes(memoryTotal)

	daemonsets, err := a.ListerRegistry.DaemonSetLister().List(labels.Everything())
	if err != nil {
		klog.Errorf("Failed to get daemonset list: %v", err)
		return caerrors.ToAutoscalerError(caerrors.ApiCallError, err)
	}
	// Snapshot scale-down actuation status before cache refresh.
	scaleDownActuationStatus := a.scaleDownActuator.CheckStatus()
	// Call CloudProvider.Refresh before any other calls to cloud provider.
	refreshStart := time.Now()
	err = a.AutoscalingContext.CloudProvider.Refresh()
	if a.AutoscalingOptions.AsyncNodeGroupsEnabled {
		// Some node groups might have been created asynchronously, without registering in CSR.
		a.clusterStateRegistry.Recalculate()
	}
	metrics.UpdateDurationFromStart(metrics.CloudProviderRefresh, refreshStart)
	if err != nil {
		klog.Errorf("Failed to refresh cloud provider config: %v", err)
		return caerrors.ToAutoscalerError(caerrors.CloudProviderError, err)
	}
	a.loopStartNotifier.Refresh()

	// Update node groups min/max and maximum number of nodes being set for all node groups after cloud provider refresh
	maxNodesCount := 0
	for _, nodeGroup := range a.AutoscalingContext.CloudProvider.NodeGroups() {
		// Don't report non-existing or upcoming node groups
		if nodeGroup.Exist() {
			metrics.UpdateNodeGroupMin(nodeGroup.Id(), nodeGroup.MinSize())
			metrics.UpdateNodeGroupMax(nodeGroup.Id(), nodeGroup.MaxSize())
			maxNodesCount += nodeGroup.MaxSize()
		}
	}
	if a.MaxNodesTotal > 0 {
		metrics.UpdateMaxNodesCount(integer.IntMin(a.MaxNodesTotal, maxNodesCount))
	} else {
		metrics.UpdateMaxNodesCount(maxNodesCount)
	}
	nonExpendableScheduledPods := core_utils.FilterOutExpendablePods(originalScheduledPods, a.ExpendablePodsPriorityCutoff)
	// Initialize cluster state to ClusterSnapshot
	if typedErr := a.initializeClusterSnapshot(allNodes, nonExpendableScheduledPods); typedErr != nil {
		return typedErr.AddPrefix("failed to initialize ClusterSnapshot: ")
	}
	// Initialize Pod Disruption Budget tracking
	if typedErr := a.initializeRemainingPdbTracker(); typedErr != nil {
		return typedErr.AddPrefix("failed to initialize RemainingPdbTracker: ")
	}

	nodeInfosForGroups, autoscalerError := a.processors.TemplateNodeInfoProvider.Process(autoscalingContext, readyNodes, daemonsets, a.taintConfig, currentTime)
	if autoscalerError != nil {
		klog.Errorf("Failed to get node infos for groups: %v", autoscalerError)
		return autoscalerError.AddPrefix("failed to build node infos for node groups: ")
	}

	a.DebuggingSnapshotter.SetTemplateNodes(nodeInfosForGroups)

	if typedErr := a.updateClusterState(allNodes, nodeInfosForGroups, currentTime); typedErr != nil {
		klog.Errorf("Failed to update cluster state: %v", typedErr)
		return typedErr
	}
	metrics.UpdateDurationFromStart(metrics.UpdateState, stateUpdateStart)

	scaleUpStatus := &status.ScaleUpStatus{Result: status.ScaleUpNotTried}
	scaleUpStatusProcessorAlreadyCalled := false
	scaleDownStatus := &scaledownstatus.ScaleDownStatus{Result: scaledownstatus.ScaleDownNotTried}

	defer func() {
		// Update status information when the loop is done (regardless of reason)
		if autoscalingContext.WriteStatusConfigMap {
			status := a.clusterStateRegistry.GetStatus(currentTime)
			utils.WriteStatusConfigMap(autoscalingContext.ClientSet, autoscalingContext.ConfigNamespace,
				*status, a.AutoscalingContext.LogRecorder, a.AutoscalingContext.StatusConfigMapName, currentTime)
		}

		// This deferred processor execution allows the processors to handle a situation when a scale-(up|down)
		// wasn't even attempted because e.g. the iteration exited earlier.
		if !scaleUpStatusProcessorAlreadyCalled && a.processors != nil && a.processors.ScaleUpStatusProcessor != nil {
			a.processors.ScaleUpStatusProcessor.Process(a.AutoscalingContext, scaleUpStatus)
		}
		if a.processors != nil && a.processors.ScaleDownStatusProcessor != nil {
			// Gather status before scaledown status processor invocation
			nodeDeletionResults, nodeDeletionResultsAsOf := a.scaleDownActuator.DeletionResults()
			scaleDownStatus.NodeDeleteResults = nodeDeletionResults
			scaleDownStatus.NodeDeleteResultsAsOf = nodeDeletionResultsAsOf
			a.scaleDownActuator.ClearResultsNotNewerThan(scaleDownStatus.NodeDeleteResultsAsOf)
			scaleDownStatus.SetUnremovableNodesInfo(a.scaleDownPlanner.UnremovableNodes(), a.scaleDownPlanner.NodeUtilizationMap(), a.CloudProvider)

			a.processors.ScaleDownStatusProcessor.Process(a.AutoscalingContext, scaleDownStatus)
		}

		if a.processors != nil && a.processors.AutoscalingStatusProcessor != nil {
			err := a.processors.AutoscalingStatusProcessor.Process(a.AutoscalingContext, a.clusterStateRegistry, currentTime)
			if err != nil {
				klog.Errorf("AutoscalingStatusProcessor error: %v.", err)
			}
		}
	}()

	// Check if there are any nodes that failed to register in Kubernetes
	// master.
	unregisteredNodes := a.clusterStateRegistry.GetUnregisteredNodes()
	if len(unregisteredNodes) > 0 {
		klog.V(1).Infof("%d unregistered nodes present", len(unregisteredNodes))
		removedAny, err := a.removeOldUnregisteredNodes(unregisteredNodes, autoscalingContext,
			a.clusterStateRegistry, currentTime, autoscalingContext.LogRecorder)
		// There was a problem with removing unregistered nodes. Retry in the next loop.
		if err != nil {
			klog.Warningf("Failed to remove unregistered nodes: %v", err)
		}
		if removedAny {
			klog.V(0).Infof("Some unregistered nodes were removed")
		}
	}

	if !a.clusterStateRegistry.IsClusterHealthy() {
		klog.Warning("Cluster is not ready for autoscaling")
		a.scaleDownPlanner.CleanUpUnneededNodes()
		autoscalingContext.LogRecorder.Eventf(apiv1.EventTypeWarning, "ClusterUnhealthy", "Cluster is unhealthy")
		return nil
	}

	a.deleteCreatedNodesWithErrors()

	// Check if there has been a constant difference between the number of nodes in k8s and
	// the number of nodes on the cloud provider side.
	// TODO: andrewskim - add protection for ready AWS nodes.
	fixedSomething, err := fixNodeGroupSize(autoscalingContext, a.clusterStateRegistry, currentTime)
	if err != nil {
		klog.Errorf("Failed to fix node group sizes: %v", err)
		return caerrors.ToAutoscalerError(caerrors.CloudProviderError, err)
	}
	if fixedSomething {
		klog.V(0).Infof("Some node group target size was fixed, skipping the iteration")
		return nil
	}

	metrics.UpdateLastTime(metrics.Autoscaling, time.Now())

	// SchedulerUnprocessed might be zero here if it was disabled
	metrics.UpdateUnschedulablePodsCount(len(unschedulablePods), len(schedulerUnprocessed))
	if isSchedulerProcessingIgnored {
		// Treat unknown pods as unschedulable, pod list processor will remove schedulable pods
		unschedulablePods = append(unschedulablePods, schedulerUnprocessed...)
	}
	// Upcoming nodes are recently created nodes that haven't registered in the cluster yet, or haven't become ready yet.
	upcomingCounts, registeredUpcoming := a.clusterStateRegistry.GetUpcomingNodes()
	// For each upcoming node we inject a placeholder node faked to appear ready into the cluster snapshot, so that we can pack unschedulable pods on
	// them and not trigger another scale-up.
	// The fake nodes are intentionally not added to the all nodes list, so that they are not considered as candidates for scale-down (which
	// doesn't make sense as they're not real).
	err = a.addUpcomingNodesToClusterSnapshot(upcomingCounts, nodeInfosForGroups)
	if err != nil {
		klog.Errorf("Failed adding upcoming nodes to cluster snapshot: %v", err)
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
		err := a.ClusterSnapshot.RemoveNode(notStartedNodeName)
		if err != nil {
			klog.Errorf("Failed to remove NotStarted node %s from cluster snapshot: %v", notStartedNodeName, err)
			// ErrNodeNotFound shouldn't happen (so it needs to be logged above if it does), but what we care about here is that the
			// node is not in the snapshot - so we don't have to error out in that case.
			if !errors.Is(err, clustersnapshot.ErrNodeNotFound) {
				return caerrors.ToAutoscalerError(caerrors.InternalError, err)
			}
		}
	}

	l, err := a.ClusterSnapshot.NodeInfos().List()
	if err != nil {
		klog.Errorf("Unable to fetch ClusterNode List for Debugging Snapshot, %v", err)
	} else {
		a.AutoscalingContext.DebuggingSnapshotter.SetClusterNodes(l)
	}

	unschedulablePodsToHelp, err := a.processors.PodListProcessor.Process(a.AutoscalingContext, unschedulablePods)

	if err != nil {
		klog.Warningf("Failed to process unschedulable pods: %v", err)
	}

	// finally, filter out pods that are too "young" to safely be considered for a scale-up (delay is configurable)
	unschedulablePodsToHelp = a.filterOutYoungPods(unschedulablePodsToHelp, currentTime)
	preScaleUp := func() time.Time {
		scaleUpStart := time.Now()
		metrics.UpdateLastTime(metrics.ScaleUp, scaleUpStart)
		return scaleUpStart
	}

	postScaleUp := func(scaleUpStart time.Time) (bool, caerrors.AutoscalerError) {
		metrics.UpdateDurationFromStart(metrics.ScaleUp, scaleUpStart)

		if a.processors != nil && a.processors.ScaleUpStatusProcessor != nil {
			a.processors.ScaleUpStatusProcessor.Process(autoscalingContext, scaleUpStatus)
			scaleUpStatusProcessorAlreadyCalled = true
		}

		if typedErr != nil {
			klog.Errorf("Failed to scale up: %v", typedErr)
			return true, typedErr
		}
		if scaleUpStatus.Result == status.ScaleUpSuccessful {
			a.lastScaleUpTime = currentTime
			// No scale down in this iteration.
			scaleDownStatus.Result = scaledownstatus.ScaleDownInCooldown
			return true, nil
		}
		return false, nil
	}

	if len(unschedulablePodsToHelp) == 0 {
		scaleUpStatus.Result = status.ScaleUpNotNeeded
		klog.V(1).Info("No unschedulable pods")
	} else if a.MaxNodesTotal > 0 && len(readyNodes) >= a.MaxNodesTotal {
		scaleUpStatus.Result = status.ScaleUpNoOptionsAvailable
		klog.V(1).Info("Max total nodes in cluster reached")
	} else if !isSchedulerProcessingIgnored && allPodsAreNew(unschedulablePodsToHelp, currentTime) {
		// The assumption here is that these pods have been created very recently and probably there
		// is more pods to come. In theory we could check the newest pod time but then if pod were created
		// slowly but at the pace of 1 every 2 seconds then no scale up would be triggered for long time.
		// We also want to skip a real scale down (just like if the pods were handled).
		a.processorCallbacks.DisableScaleDownForLoop()
		scaleUpStatus.Result = status.ScaleUpInCooldown
		klog.V(1).Info("Unschedulable pods are very new, waiting one iteration for more")
	} else {
		scaleUpStart := preScaleUp()
		scaleUpStatus, typedErr = a.scaleUpOrchestrator.ScaleUp(unschedulablePodsToHelp, readyNodes, daemonsets, nodeInfosForGroups, false)
		if exit, err := postScaleUp(scaleUpStart); exit {
			return err
		}
	}

	if a.ScaleDownEnabled {
		unneededStart := time.Now()

		klog.V(4).Infof("Calculating unneeded nodes")

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
			scaleDownCandidates, err = a.processors.ScaleDownNodeProcessor.GetScaleDownCandidates(
				autoscalingContext, allNodes)
			if err != nil {
				klog.Error(err)
				return err
			}
			podDestinations, err = a.processors.ScaleDownNodeProcessor.GetPodDestinationCandidates(autoscalingContext, allNodes)
			if err != nil {
				klog.Error(err)
				return err
			}
		}

		typedErr := a.scaleDownPlanner.UpdateClusterState(podDestinations, scaleDownCandidates, scaleDownActuationStatus, currentTime)
		// Update clusterStateRegistry and metrics regardless of whether ScaleDown was successful or not.
		unneededNodes := a.scaleDownPlanner.UnneededNodes()
		a.processors.ScaleDownCandidatesNotifier.Update(unneededNodes, currentTime)
		metrics.UpdateUnneededNodesCount(len(unneededNodes))
		if typedErr != nil {
			scaleDownStatus.Result = scaledownstatus.ScaleDownError
			klog.Errorf("Failed to scale down: %v", typedErr)
			return typedErr
		}

		metrics.UpdateDurationFromStart(metrics.FindUnneeded, unneededStart)

		scaleDownInCooldown := a.isScaleDownInCooldown(currentTime, scaleDownCandidates)
		klog.V(4).Infof("Scale down status: lastScaleUpTime=%s lastScaleDownDeleteTime=%v "+
			"lastScaleDownFailTime=%s scaleDownForbidden=%v scaleDownInCooldown=%v",
			a.lastScaleUpTime, a.lastScaleDownDeleteTime, a.lastScaleDownFailTime,
			a.processorCallbacks.disableScaleDownForLoop, scaleDownInCooldown)
		metrics.UpdateScaleDownInCooldown(scaleDownInCooldown)
		// We want to delete unneeded Node Groups only if here is no current delete
		// in progress.
		_, drained := scaleDownActuationStatus.DeletionsInProgress()
		var removedNodeGroups []cloudprovider.NodeGroup
		if len(drained) == 0 {
			var err error
			removedNodeGroups, err = a.processors.NodeGroupManager.RemoveUnneededNodeGroups(autoscalingContext)
			if err != nil {
				klog.Errorf("Error while removing unneeded node groups: %v", err)
			}
			scaleDownStatus.RemovedNodeGroups = removedNodeGroups
		}

		if scaleDownInCooldown {
			scaleDownStatus.Result = scaledownstatus.ScaleDownInCooldown
		} else {
			klog.V(4).Infof("Starting scale down")

			scaleDownStart := time.Now()
			metrics.UpdateLastTime(metrics.ScaleDown, scaleDownStart)
			empty, needDrain := a.scaleDownPlanner.NodesToDelete(currentTime)
			scaleDownResult, scaledDownNodes, typedErr := a.scaleDownActuator.StartDeletion(empty, needDrain)
			scaleDownStatus.Result = scaleDownResult
			scaleDownStatus.ScaledDownNodes = scaledDownNodes
			metrics.UpdateDurationFromStart(metrics.ScaleDown, scaleDownStart)
			metrics.UpdateUnremovableNodesCount(countsByReason(a.scaleDownPlanner.UnremovableNodes()))

			scaleDownStatus.RemovedNodeGroups = removedNodeGroups

			if scaleDownStatus.Result == scaledownstatus.ScaleDownNodeDeleteStarted {
				a.lastScaleDownDeleteTime = currentTime
				a.clusterStateRegistry.Recalculate()
			}

			if scaleDownStatus.Result == scaledownstatus.ScaleDownNoNodeDeleted &&
				a.AutoscalingContext.AutoscalingOptions.MaxBulkSoftTaintCount != 0 {
				taintableNodes := a.scaleDownPlanner.UnneededNodes()

				// Make sure we are only cleaning taints from selected node groups.
				selectedNodes := filterNodesFromSelectedGroups(a.CloudProvider, allNodes...)

				// This is a sanity check to make sure `taintableNodes` only includes
				// nodes from selected nodes.
				taintableNodes = intersectNodes(selectedNodes, taintableNodes)
				untaintableNodes := subtractNodes(selectedNodes, taintableNodes)
				actuation.UpdateSoftDeletionTaints(a.AutoscalingContext, taintableNodes, untaintableNodes)
			}

			if typedErr != nil {
				klog.Errorf("Failed to scale down: %v", typedErr)
				a.lastScaleDownFailTime = currentTime
				return typedErr
			}
		}
	}

	if a.EnforceNodeGroupMinSize {
		scaleUpStart := preScaleUp()
		scaleUpStatus, typedErr = a.scaleUpOrchestrator.ScaleUpToNodeGroupMinSize(readyNodes, nodeInfosForGroups)
		if exit, err := postScaleUp(scaleUpStart); exit {
			return err
		}
	}

	return nil
}

func (a *StaticAutoscaler) addUpcomingNodesToClusterSnapshot(upcomingCounts map[string]int, nodeInfosForGroups map[string]*schedulerframework.NodeInfo) error {
	nodeGroups := a.nodeGroupsById()
	upcomingNodeGroups := make(map[string]int)
	upcomingNodesFromUpcomingNodeGroups := 0
	for nodeGroupName, upcomingNodes := range getUpcomingNodeInfos(upcomingCounts, nodeInfosForGroups) {
		nodeGroup := nodeGroups[nodeGroupName]
		if nodeGroup == nil {
			return fmt.Errorf("failed to find node group: %s", nodeGroupName)
		}
		isUpcomingNodeGroup := a.processors.AsyncNodeGroupStateChecker.IsUpcoming(nodeGroup)
		for _, upcomingNode := range upcomingNodes {
			var pods []*apiv1.Pod
			for _, podInfo := range upcomingNode.Pods {
				pods = append(pods, podInfo.Pod)
			}
			err := a.ClusterSnapshot.AddNodeWithPods(upcomingNode.Node(), pods)
			if err != nil {
				return fmt.Errorf("Failed to add upcoming node %s to cluster snapshot: %w", upcomingNode.Node().Name, err)
			}
			if isUpcomingNodeGroup {
				upcomingNodesFromUpcomingNodeGroups++
				upcomingNodeGroups[nodeGroup.Id()] += 1
			}
		}
	}
	if len(upcomingNodeGroups) > 0 {
		klog.Infof("Injecting %d upcoming node groups with %d upcoming nodes: %v", len(upcomingNodeGroups), upcomingNodesFromUpcomingNodeGroups, upcomingNodeGroups)
	}
	return nil
}

func (a *StaticAutoscaler) isScaleDownInCooldown(currentTime time.Time, scaleDownCandidates []*apiv1.Node) bool {
	scaleDownInCooldown := a.processorCallbacks.disableScaleDownForLoop || len(scaleDownCandidates) == 0

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
func fixNodeGroupSize(context *context.AutoscalingContext, clusterStateRegistry *clusterstate.ClusterStateRegistry, currentTime time.Time) (bool, error) {
	fixed := false
	for _, nodeGroup := range context.CloudProvider.NodeGroups() {
		incorrectSize := clusterStateRegistry.GetIncorrectNodeGroupSize(nodeGroup.Id())
		if incorrectSize == nil {
			continue
		}
		maxNodeProvisionTime, err := clusterStateRegistry.MaxNodeProvisionTime(nodeGroup)
		if err != nil {
			return false, fmt.Errorf("failed to retrieve maxNodeProvisionTime for nodeGroup %s", nodeGroup.Id())
		}
		if incorrectSize.FirstObserved.Add(maxNodeProvisionTime).Before(currentTime) {
			delta := incorrectSize.CurrentSize - incorrectSize.ExpectedSize
			if delta < 0 {
				klog.V(0).Infof("Decreasing size of %s, expected=%d current=%d delta=%d", nodeGroup.Id(),
					incorrectSize.ExpectedSize,
					incorrectSize.CurrentSize,
					delta)
				if err := nodeGroup.DecreaseTargetSize(delta); err != nil {
					return fixed, fmt.Errorf("failed to decrease %s: %v", nodeGroup.Id(), err)
				}
				fixed = true
			}
		}
	}
	return fixed, nil
}

// Removes unregistered nodes if needed. Returns true if anything was removed and error if such occurred.
func (a *StaticAutoscaler) removeOldUnregisteredNodes(allUnregisteredNodes []clusterstate.UnregisteredNode, context *context.AutoscalingContext,
	csr *clusterstate.ClusterStateRegistry, currentTime time.Time, logRecorder *utils.LogEventRecorder) (bool, error) {

	nodeGroups := a.nodeGroupsById()
	nodesToDeleteByNodeGroupId := make(map[string][]clusterstate.UnregisteredNode)
	for _, unregisteredNode := range allUnregisteredNodes {
		nodeGroup, err := a.CloudProvider.NodeGroupForNode(unregisteredNode.Node)
		if err != nil {
			klog.Warningf("Failed to get node group for %s: %v", unregisteredNode.Node.Name, err)
			continue
		}
		if nodeGroup == nil || reflect.ValueOf(nodeGroup).IsNil() {
			klog.Warningf("No node group for node %s, skipping", unregisteredNode.Node.Name)
			continue
		}

		maxNodeProvisionTime, err := csr.MaxNodeProvisionTime(nodeGroup)
		if err != nil {
			return false, fmt.Errorf("failed to retrieve maxNodeProvisionTime for node %s in nodeGroup %s", unregisteredNode.Node.Name, nodeGroup.Id())
		}

		if unregisteredNode.UnregisteredSince.Add(maxNodeProvisionTime).Before(currentTime) {
			klog.V(0).Infof("Marking unregistered node %v for removal", unregisteredNode.Node.Name)
			nodesToDeleteByNodeGroupId[nodeGroup.Id()] = append(nodesToDeleteByNodeGroupId[nodeGroup.Id()], unregisteredNode)
		}
	}

	removedAny := false
	for nodeGroupId, unregisteredNodesToDelete := range nodesToDeleteByNodeGroupId {
		nodeGroup := nodeGroups[nodeGroupId]

		klog.V(0).Infof("Removing %v unregistered nodes for node group %v", len(unregisteredNodesToDelete), nodeGroupId)
		size, err := nodeGroup.TargetSize()
		if err != nil {
			klog.Warningf("Failed to get node group size; nodeGroup=%v; err=%v", nodeGroup.Id(), err)
			continue
		}
		possibleToDelete := size - nodeGroup.MinSize()
		if possibleToDelete <= 0 {
			klog.Warningf("Node group %s min size reached, skipping removal of %v unregistered nodes", nodeGroupId, len(unregisteredNodesToDelete))
			continue
		}
		if len(unregisteredNodesToDelete) > possibleToDelete {
			klog.Warningf("Capping node group %s unregistered node removal to %d nodes, removing all %d would exceed min size constaint", nodeGroupId, possibleToDelete, len(unregisteredNodesToDelete))
			unregisteredNodesToDelete = unregisteredNodesToDelete[:possibleToDelete]
		}
		nodesToDelete := toNodes(unregisteredNodesToDelete)

		nodesToDelete, err = overrideNodesToDeleteForZeroOrMax(a.NodeGroupDefaults, nodeGroup, nodesToDelete)
		if err != nil {
			klog.Warningf("Failed to remove unregistered nodes from node group %s: %v", nodeGroupId, err)
			continue
		}

		err = nodeGroup.DeleteNodes(nodesToDelete)
		csr.InvalidateNodeInstancesCacheEntry(nodeGroup)
		if err != nil {
			klog.Warningf("Failed to remove %v unregistered nodes from node group %s: %v", len(nodesToDelete), nodeGroupId, err)
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

func toNodes(unregisteredNodes []clusterstate.UnregisteredNode) []*apiv1.Node {
	nodes := []*apiv1.Node{}
	for _, n := range unregisteredNodes {
		nodes = append(nodes, n.Node)
	}
	return nodes
}

func (a *StaticAutoscaler) deleteCreatedNodesWithErrors() {
	// We always schedule deleting of incoming errornous nodes
	// TODO[lukaszos] Consider adding logic to not retry delete every loop iteration
	nodeGroups := a.nodeGroupsById()
	nodesToDeleteByNodeGroupId := a.clusterStateRegistry.GetCreatedNodesWithErrors()

	deletedAny := false

	for nodeGroupId, nodesToDelete := range nodesToDeleteByNodeGroupId {
		var err error
		klog.V(1).Infof("Deleting %v from %v node group because of create errors", len(nodesToDelete), nodeGroupId)

		nodeGroup := nodeGroups[nodeGroupId]
		if nodeGroup == nil {
			err = fmt.Errorf("node group %s not found", nodeGroupId)
		} else if nodesToDelete, err = overrideNodesToDeleteForZeroOrMax(a.NodeGroupDefaults, nodeGroup, nodesToDelete); err == nil {
			err = nodeGroup.DeleteNodes(nodesToDelete)
		}

		if err != nil {
			klog.Warningf("Error while trying to delete nodes from %v: %v", nodeGroupId, err)
		} else {
			deletedAny = true
			a.clusterStateRegistry.InvalidateNodeInstancesCacheEntry(nodeGroup)
		}
	}

	if deletedAny {
		klog.V(0).Infof("Some nodes that failed to create were removed, recalculating cluster state.")
		a.clusterStateRegistry.Recalculate()
	}
}

// overrideNodesToDeleteForZeroOrMax returns a list of nodes to delete, taking into account that
// node deletion for a "ZeroOrMaxNodeScaling" node group is atomic and should delete all nodes.
// For a non-"ZeroOrMaxNodeScaling" node group it returns the unchanged list of nodes to delete.
func overrideNodesToDeleteForZeroOrMax(defaults config.NodeGroupAutoscalingOptions, nodeGroup cloudprovider.NodeGroup, nodesToDelete []*apiv1.Node) ([]*apiv1.Node, error) {
	opts, err := nodeGroup.GetOptions(defaults)
	if err != nil && err != cloudprovider.ErrNotImplemented {
		return []*apiv1.Node{}, fmt.Errorf("Failed to get node group options for %s: %s", nodeGroup.Id(), err)
	}
	// If a scale-up of "ZeroOrMaxNodeScaling" node group failed, the cleanup
	// should stick to the all-or-nothing principle. Deleting all nodes.
	if opts != nil && opts.ZeroOrMaxNodeScaling {
		instances, err := nodeGroup.Nodes()
		if err != nil {
			return []*apiv1.Node{}, fmt.Errorf("Failed to fill in nodes to delete from group %s based on ZeroOrMaxNodeScaling option: %s", nodeGroup.Id(), err)
		}
		return instancesToFakeNodes(instances), nil
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

func (a *StaticAutoscaler) nodeGroupsById() map[string]cloudprovider.NodeGroup {
	nodeGroups := make(map[string]cloudprovider.NodeGroup)
	for _, nodeGroup := range a.CloudProvider.NodeGroups() {
		nodeGroups[nodeGroup.Id()] = nodeGroup
	}
	return nodeGroups
}

// Don't consider pods newer than newPodScaleUpDelay or annotated podScaleUpDelay
// seconds old as unschedulable.
func (a *StaticAutoscaler) filterOutYoungPods(allUnschedulablePods []*apiv1.Pod, currentTime time.Time) []*apiv1.Pod {
	var oldUnschedulablePods []*apiv1.Pod
	newPodScaleUpDelay := a.AutoscalingOptions.NewPodScaleUpDelay
	for _, pod := range allUnschedulablePods {
		podAge := currentTime.Sub(pod.CreationTimestamp.Time)
		podScaleUpDelay := newPodScaleUpDelay

		if podScaleUpDelayAnnotationStr, ok := pod.Annotations[podScaleUpDelayAnnotationKey]; ok {
			podScaleUpDelayAnnotation, err := time.ParseDuration(podScaleUpDelayAnnotationStr)
			if err != nil {
				klog.Errorf("Failed to parse pod %q annotation %s: %v", pod.Name, podScaleUpDelayAnnotationKey, err)
			} else {
				if podScaleUpDelayAnnotation < podScaleUpDelay {
					klog.Errorf("Failed to set pod scale up delay for %q through annotation %s: %d is less then %d", pod.Name, podScaleUpDelayAnnotationKey, podScaleUpDelayAnnotation, newPodScaleUpDelay)
				} else {
					podScaleUpDelay = podScaleUpDelayAnnotation
				}
			}
		}

		if podAge > podScaleUpDelay {
			oldUnschedulablePods = append(oldUnschedulablePods, pod)
		} else {
			klog.V(3).Infof("Pod %s is %.3f seconds old, too new to consider unschedulable", pod.Name, podAge.Seconds())
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

	a.CloudProvider.Cleanup()

	a.clusterStateRegistry.Stop()
}

func (a *StaticAutoscaler) obtainNodeLists() ([]*apiv1.Node, []*apiv1.Node, caerrors.AutoscalerError) {
	allNodes, err := a.AllNodeLister().List()
	if err != nil {
		klog.Errorf("Failed to list all nodes: %v", err)
		return nil, nil, caerrors.ToAutoscalerError(caerrors.ApiCallError, err)
	}
	readyNodes, err := a.ReadyNodeLister().List()
	if err != nil {
		klog.Errorf("Failed to list ready nodes: %v", err)
		return nil, nil, caerrors.ToAutoscalerError(caerrors.ApiCallError, err)
	}
	a.reportTaintsCount(allNodes)

	// Handle GPU case - allocatable GPU may be equal to 0 up to 15 minutes after
	// node registers as ready. See https://github.com/kubernetes/kubernetes/issues/54959
	// Treat those nodes as unready until GPU actually becomes available and let
	// our normal handling for booting up nodes deal with this.
	// TODO: Remove this call when we handle dynamically provisioned resources.
	allNodes, readyNodes = a.processors.CustomResourcesProcessor.FilterOutNodesWithUnreadyResources(a.AutoscalingContext, allNodes, readyNodes)
	allNodes, readyNodes = taints.FilterOutNodesWithStartupTaints(a.taintConfig, allNodes, readyNodes)
	return allNodes, readyNodes, nil
}

func filterNodesFromSelectedGroups(cp cloudprovider.CloudProvider, nodes ...*apiv1.Node) []*apiv1.Node {
	filtered := make([]*apiv1.Node, 0, len(nodes))
	for _, n := range nodes {
		if ng, err := cp.NodeGroupForNode(n); err != nil {
			klog.Errorf("Failed to get a node group node node: %v", err)
		} else if ng != nil {
			filtered = append(filtered, n)
		}
	}
	return filtered
}

func (a *StaticAutoscaler) updateClusterState(allNodes []*apiv1.Node, nodeInfosForGroups map[string]*schedulerframework.NodeInfo, currentTime time.Time) caerrors.AutoscalerError {
	err := a.clusterStateRegistry.UpdateNodes(allNodes, nodeInfosForGroups, currentTime)
	if err != nil {
		klog.Errorf("Failed to update node registry: %v", err)
		a.scaleDownPlanner.CleanUpUnneededNodes()
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

func getUpcomingNodeInfos(upcomingCounts map[string]int, nodeInfos map[string]*schedulerframework.NodeInfo) map[string][]*schedulerframework.NodeInfo {
	upcomingNodes := make(map[string][]*schedulerframework.NodeInfo)
	for nodeGroup, numberOfNodes := range upcomingCounts {
		nodeTemplate, found := nodeInfos[nodeGroup]
		if !found {
			klog.Warningf("Couldn't find template for node group %s", nodeGroup)
			continue
		}

		if nodeTemplate.Node().Annotations == nil {
			nodeTemplate.Node().Annotations = make(map[string]string)
		}
		nodeTemplate.Node().Annotations[NodeUpcomingAnnotation] = "true"

		var nodes []*schedulerframework.NodeInfo
		for i := 0; i < numberOfNodes; i++ {
			// Ensure new nodes have different names because nodeName
			// will be used as a map key. Also deep copy pods (daemonsets &
			// any pods added by cloud provider on template).
			nodes = append(nodes, scheduler_utils.DeepCopyTemplateNode(nodeTemplate, fmt.Sprintf("upcoming-%d", i)))
		}
		upcomingNodes[nodeGroup] = nodes
	}
	return upcomingNodes
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
