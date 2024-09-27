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

package besteffortatomic

import (
	"reflect"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/klog/v2"

	"k8s.io/autoscaler/cluster-autoscaler/apis/provisioningrequest/autoscaling.x-k8s.io/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate"
	"k8s.io/autoscaler/cluster-autoscaler/context"

	"k8s.io/autoscaler/cluster-autoscaler/core/scaleup/equivalence"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaleup/orchestrator"
	"k8s.io/autoscaler/cluster-autoscaler/estimator"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodegroups"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodegroupset"
	"k8s.io/autoscaler/cluster-autoscaler/processors/status"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/conditions"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/provreqclient"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/provreqwrapper"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/scheduling"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"

	ca_processors "k8s.io/autoscaler/cluster-autoscaler/processors"
)

// Best effort atomic provisionig class requests scale-up only if it's possible
// to atomically request enough resources for all pods specified in a
// ProvisioningRequest. It's "best effort" as it admits workload immediately
// after successful request, without waiting to verify that resources started.
type bestEffortAtomicProvClass struct {
	context             *context.AutoscalingContext
	client              *provreqclient.ProvisioningRequestClient
	schedulingSimulator *scheduling.HintingSimulator
	scaleUpOrchestrator *orchestrator.ScaleUpOrchestrator
}

// New creates best effort atomic provisioning class supporting create capacity scale-up mode.
func New(
	client *provreqclient.ProvisioningRequestClient,
) *bestEffortAtomicProvClass {
	return &bestEffortAtomicProvClass{client: client, scaleUpOrchestrator: orchestrator.New()}
}

func (o *bestEffortAtomicProvClass) Initialize(
	autoscalingContext *context.AutoscalingContext,
	processors *ca_processors.AutoscalingProcessors,
	clusterStateRegistry *clusterstate.ClusterStateRegistry,
	estimatorBuilder estimator.EstimatorBuilder,
	taintConfig taints.TaintConfig,
	schedulingSimulator *scheduling.HintingSimulator,
) {
	o.context = autoscalingContext
	o.schedulingSimulator = schedulingSimulator
	o.scaleUpOrchestrator.Initialize(autoscalingContext, processors, clusterStateRegistry, estimatorBuilder, taintConfig)
}

// Provision returns success if there is, or has just been requested, sufficient capacity in the cluster for pods from ProvisioningRequest.
func (o *bestEffortAtomicProvClass) Provision(
	unschedulablePods []*apiv1.Pod,
	nodes []*apiv1.Node,
	daemonSets []*appsv1.DaemonSet,
	nodeInfos map[string]*framework.NodeInfo,
) (*status.ScaleUpStatus, errors.AutoscalerError) {
	if len(unschedulablePods) == 0 {
		return &status.ScaleUpStatus{Result: status.ScaleUpNotTried}, nil
	}

	combinedStatus := status.NewCombinedStatusSet()
	simulationResults := make(map[*provreqwrapper.ProvisioningRequest]*orchestrator.ScaleUpSimulationResult)

	bestOptionPods := []*apiv1.Pod{}
	skippedNodeGroupsMap := make(map[string]status.Reasons)
	consideredNodeGroupsMap := make(map[string]cloudprovider.NodeGroup)
	aggregatedPodEquivalenceGroups := []*equivalence.PodGroup{}
	bestOptionPodsAwaitEvaluation := []*apiv1.Pod{}

	scaleUpInfosMap := make(map[string]nodegroupset.ScaleUpInfo)
	createNodeGroupResultsMap := make(map[string]nodegroups.CreateNodeGroupResult)

	finalPrs := []*provreqwrapper.ProvisioningRequest{}

	prMap := provreqclient.SegregatePodsByProvisioningRequest(o.client, unschedulablePods)

	var scaleUpStatus *status.ScaleUpStatus
	startTime := time.Now()

	o.context.ClusterSnapshot.Fork()
	defer o.context.ClusterSnapshot.Revert()

	for _, prPods := range prMap {
		pr, err := o.client.ProvisioningRequest(prPods[0].Namespace, prPods[0].OwnerReferences[0].Name)
		if err != nil {
			klog.Errorf("failed to retrieve ProvisioningRequest from unschedulable pod, err: %v", err)
			scaleUpStatus, _ = status.UpdateScaleUpError(&status.ScaleUpStatus{}, errors.NewAutoscalerError(errors.InternalError, "error during ScaleUp: %s", err.Error()))
			combinedStatus.Add(scaleUpStatus)
			continue
		}

		// Skip ProvisioningRequests with non-best-effort provisioning class
		if pr.Spec.ProvisioningClassName != v1.ProvisioningClassBestEffortAtomicScaleUp {
			combinedStatus.Add(&status.ScaleUpStatus{Result: status.ScaleUpNotTried})
			continue
		}

		// Try to schedule pods which can be scheduled without scale-up.
		unschedulablePrPods, err := o.filterOutSchedulable(prPods)
		if err != nil {
			conditions.AddOrUpdateCondition(pr, v1.Provisioned, metav1.ConditionFalse, conditions.FailedToCheckCapacityReason, conditions.FailedToCheckCapacityMsg, metav1.Now())
			if _, updateErr := o.client.UpdateProvisioningRequest(pr.ProvisioningRequest); updateErr != nil {
				klog.Errorf("failed to add Provisioned=false condition to ProvReq %s/%s, err: %v", pr.Namespace, pr.Name, updateErr)
			}
			scaleUpStatus, _ = status.UpdateScaleUpError(&status.ScaleUpStatus{}, errors.NewAutoscalerError(errors.InternalError, "error during ScaleUp: %s", err.Error()))
			combinedStatus.Add(scaleUpStatus)
			continue
		}

		if len(unschedulablePrPods) == 0 {
			// Nothing to do here - everything fits without scale-up.
			conditions.AddOrUpdateCondition(pr, v1.Provisioned, metav1.ConditionTrue, conditions.CapacityIsFoundReason, conditions.CapacityIsFoundMsg, metav1.Now())
			if _, updateErr := o.client.UpdateProvisioningRequest(pr.ProvisioningRequest); updateErr != nil {
				klog.Errorf("failed to add Provisioned=true condition to ProvReq %s/%s, err: %v", pr.Namespace, pr.Name, updateErr)
				scaleUpStatus, _ = status.UpdateScaleUpError(&status.ScaleUpStatus{}, errors.NewAutoscalerError(errors.InternalError, "capacity available, but failed to admit workload: %s", updateErr.Error()))
				combinedStatus.Add(scaleUpStatus)
				continue
			}
			combinedStatus.Add(&status.ScaleUpStatus{Result: status.ScaleUpNotNeeded})
			continue
		}

		if !o.scaleUpOrchestrator.IsInitialized() {
			scaleUpStatus, _ = status.UpdateScaleUpError(&status.ScaleUpStatus{}, errors.NewAutoscalerError(errors.InternalError, "scale up orchestrator is not initialized"))
			combinedStatus.Add(scaleUpStatus)
			continue
		}

		klog.V(1).Infof("ProvisioningRequest %s/%s has %d unschedulable pods", pr.Namespace, pr.Name, len(unschedulablePrPods))

		// Build equivalence groups for unschedulable pods
		podEquivalenceGroups := orchestrator.BuildPodEquivalenceGroups(unschedulablePrPods)

		// Simulation of scale-up and preparation of scale-up plan
		simulationResult, simulationScaleUpStatus, aErr := o.scaleUpOrchestrator.SimulateScaleUp(podEquivalenceGroups, nodes, nodeInfos, true)
		if aErr != nil || (simulationScaleUpStatus != nil && simulationResult == nil) {
			conditions.AddOrUpdateCondition(pr, v1.Provisioned, metav1.ConditionFalse, conditions.CapacityIsNotFoundReason, "Capacity is not found, CA will try to find it later.", metav1.Now())
			if _, updateErr := o.client.UpdateProvisioningRequest(pr.ProvisioningRequest); updateErr != nil {
				klog.Errorf("failed to add Provisioned=false condition to ProvReq %s/%s, err: %v", pr.Namespace, pr.Name, updateErr)
			}

			if aErr != nil {
				scaleUpStatus, _ = status.UpdateScaleUpError(&status.ScaleUpStatus{}, errors.NewAutoscalerError(errors.InternalError, "error during ScaleUp: %s", aErr.Error()))
				combinedStatus.Add(scaleUpStatus)
				continue
			}

			combinedStatus.Add(simulationScaleUpStatus)
			continue
		}

		// Update snapshot with the exported snapshot from the best option
		simulationResult.BestOption.SnapshotExport.Rebase(o.context.ClusterSnapshot)
		if o.context.ClusterSnapshot == nil {
			klog.Errorf("Cluster snapshot is nil")
			combinedStatus.Add(&status.ScaleUpStatus{Result: status.ScaleUpNotTried})
			continue
		}
		reflect.ValueOf(o.context.ClusterSnapshot).Elem().Set(reflect.ValueOf(simulationResult.BestOption.SnapshotExport).Elem())

		// Add simulation result to simulation result array
		simulationResults[pr] = simulationResult

		if time.Since(startTime) > o.context.BestEffortAtomicProvisioningRequestShardedSimulationTimebox {
			klog.Info("Simulation timebox exceeded, stopping further simulations")
			break
		}
	}

	for _, simRes := range simulationResults {
		klog.Warning("simRes.BestOption: ", simRes.BestOption)
	}

	// Prepare node groups for scale-up
	for pr, simulationResult := range simulationResults {
		preparationResult, preparationScaleUpStatus, aErr := o.scaleUpOrchestrator.PrepareNodeGroupsForScaleUp(
			simulationResult.BestOption,
			simulationResult.NewNodes,
			simulationResult.SkippedNodeGroups,
			simulationResult.NodeGroups,
			simulationResult.NodeInfos,
			simulationResult.SchedulablePodGroups,
			simulationResult.PodEquivalenceGroups,
			daemonSets,
			true,
		)
		if aErr != nil || (preparationScaleUpStatus != nil && preparationResult == nil) {
			conditions.AddOrUpdateCondition(pr, v1.Provisioned, metav1.ConditionFalse, conditions.CapacityIsNotFoundReason, "Capacity is not found, CA will try to find it later.", metav1.Now())
			if _, updateErr := o.client.UpdateProvisioningRequest(pr.ProvisioningRequest); updateErr != nil {
				klog.Errorf("failed to add Provisioned=false condition to ProvReq %s/%s, err: %v", pr.Namespace, pr.Name, updateErr)
			}

			if aErr != nil {
				scaleUpStatus, _ = status.UpdateScaleUpError(&status.ScaleUpStatus{}, errors.NewAutoscalerError(errors.InternalError, "error during ScaleUp: %s", aErr.Error()))
				combinedStatus.Add(scaleUpStatus)
				continue
			}

			combinedStatus.Add(preparationScaleUpStatus)
			continue
		}

		// Add simulation results to execution values
		bestOptionPods = append(bestOptionPods, simulationResult.BestOption.Pods...)
		for nodeGroupName, skippedReasons := range simulationResult.SkippedNodeGroups {
			skippedNodeGroupsMap[nodeGroupName] = orchestrator.NewSkippedReasons(append(skippedNodeGroupsMap[nodeGroupName].Reasons(), skippedReasons.Reasons()...)...)
		}
		for _, nodeGroup := range simulationResult.NodeGroups {
			consideredNodeGroupsMap[nodeGroup.Id()] = nodeGroup
		}
		bestOptionPodsAwaitEvaluation = append(bestOptionPodsAwaitEvaluation, orchestrator.GetPodsAwaitingEvaluation(simulationResult.PodEquivalenceGroups, simulationResult.BestOption.NodeGroup.Id())...)
		aggregatedPodEquivalenceGroups = append(aggregatedPodEquivalenceGroups, simulationResult.PodEquivalenceGroups...)
		finalPrs = append(finalPrs, pr)

		// Add preparation result to execution values
		for _, scaleUpInfo := range preparationResult.ScaleUpInfos {
			prevScaleUpInfo, found := scaleUpInfosMap[scaleUpInfo.Group.Id()]
			if found {
				scaleUpInfo.CurrentSize = prevScaleUpInfo.CurrentSize
			}
			scaleUpInfosMap[scaleUpInfo.Group.Id()] = scaleUpInfo
		}
		for _, createNodeGroupResult := range preparationResult.CreateNodeGroupResults {
			createNodeGroupResultsMap[createNodeGroupResult.MainCreatedNodeGroup.Id()] = createNodeGroupResult
		}
	}

	consideredNodeGroups := make([]cloudprovider.NodeGroup, 0, len(consideredNodeGroupsMap))
	for _, nodeGroup := range consideredNodeGroupsMap {
		consideredNodeGroups = append(consideredNodeGroups, nodeGroup)
	}

	scaleUpInfos := make([]nodegroupset.ScaleUpInfo, 0, len(scaleUpInfosMap))
	for _, scaleUpInfo := range scaleUpInfosMap {
		scaleUpInfos = append(scaleUpInfos, scaleUpInfo)
	}

	createNodeGroupResults := make([]nodegroups.CreateNodeGroupResult, 0, len(createNodeGroupResultsMap))
	for _, createNodeGroupResult := range createNodeGroupResultsMap {
		createNodeGroupResults = append(createNodeGroupResults, createNodeGroupResult)
	}

	if len(finalPrs) > 0 && len(bestOptionPods) > 0 {
		// Execute scale-up
		combinedExecutionScaleUpStatus, err := o.scaleUpOrchestrator.ExecuteScaleUp(
			bestOptionPods,
			skippedNodeGroupsMap,
			consideredNodeGroups,
			nodeInfos,
			aggregatedPodEquivalenceGroups,
			scaleUpInfos,
			createNodeGroupResults,
			bestOptionPodsAwaitEvaluation,
			true,
		)

		if err != nil {
			scaleUpStatus, _ = status.UpdateScaleUpError(&status.ScaleUpStatus{}, errors.NewAutoscalerError(errors.InternalError, "error during ScaleUp: %s", err.Error()))

			for _, pr := range finalPrs {
				conditions.AddOrUpdateCondition(pr, v1.Provisioned, metav1.ConditionFalse, conditions.CapacityIsNotFoundReason, "Capacity is not found, CA will try to find it later.", metav1.Now())
				if _, updateErr := o.client.UpdateProvisioningRequest(pr.ProvisioningRequest); updateErr != nil {
					klog.Errorf("failed to add Provisioned=false condition to ProvReq %s/%s, err: %v", pr.Namespace, pr.Name, updateErr)
				}
			}

			combinedStatus.Add(scaleUpStatus)
		} else {
			for _, pr := range finalPrs {
				conditions.AddOrUpdateCondition(pr, v1.Provisioned, metav1.ConditionTrue, conditions.CapacityIsFoundReason, conditions.CapacityIsFoundMsg, metav1.Now())
				if _, updateErr := o.client.UpdateProvisioningRequest(pr.ProvisioningRequest); updateErr != nil {
					klog.Errorf("failed to add Provisioned=true condition to ProvReq %s/%s, err: %v", pr.Namespace, pr.Name, updateErr)
					scaleUpStatus, _ = status.UpdateScaleUpError(&status.ScaleUpStatus{}, errors.NewAutoscalerError(errors.InternalError, "scale up requested, but failed to admit workload: %s", updateErr.Error()))
					combinedStatus.Add(scaleUpStatus)
				}
			}

			combinedStatus.Add(combinedExecutionScaleUpStatus)
		}
	}

	return combinedStatus.Export()
}

func (o *bestEffortAtomicProvClass) filterOutSchedulable(pods []*apiv1.Pod) ([]*apiv1.Pod, error) {
	o.context.ClusterSnapshot.Fork()

	statuses, _, err := o.schedulingSimulator.TrySchedulePods(o.context.ClusterSnapshot, pods, scheduling.ScheduleAnywhere, false)
	if err != nil {
		o.context.ClusterSnapshot.Revert()
		return nil, err
	}

	scheduledPods := make(map[types.UID]bool)
	for _, status := range statuses {
		scheduledPods[status.Pod.UID] = true
	}

	var unschedulablePods []*apiv1.Pod
	for _, pod := range pods {
		if !scheduledPods[pod.UID] {
			unschedulablePods = append(unschedulablePods, pod)
		}
	}
	o.context.ClusterSnapshot.Commit()
	return unschedulablePods, nil
}
