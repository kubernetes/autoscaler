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
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"

	"k8s.io/autoscaler/cluster-autoscaler/apis/provisioningrequest/autoscaling.x-k8s.io/v1"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaleup"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaleup/orchestrator"
	"k8s.io/autoscaler/cluster-autoscaler/estimator"
	"k8s.io/autoscaler/cluster-autoscaler/processors/status"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/conditions"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/provreqclient"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/scheduling"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"

	ca_processors "k8s.io/autoscaler/cluster-autoscaler/processors"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
)

// Best effort atomic provisionig class requests scale-up only if it's possible
// to atomically request enough resources for all pods specified in a
// ProvisioningRequest. It's "best effort" as it admits workload immediately
// after successful request, without waiting to verify that resources started.
type bestEffortAtomicProvClass struct {
	context             *context.AutoscalingContext
	client              *provreqclient.ProvisioningRequestClient
	injector            *scheduling.HintingSimulator
	scaleUpOrchestrator scaleup.Orchestrator
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
	injector *scheduling.HintingSimulator,
) {
	o.context = autoscalingContext
	o.injector = injector
	o.scaleUpOrchestrator.Initialize(autoscalingContext, processors, clusterStateRegistry, estimatorBuilder, taintConfig)
}

// Provision returns success if there is, or has just been requested, sufficient capacity in the cluster for pods from ProvisioningRequest.
func (o *bestEffortAtomicProvClass) Provision(
	unschedulablePods []*apiv1.Pod,
	nodes []*apiv1.Node,
	daemonSets []*appsv1.DaemonSet,
	nodeInfos map[string]*schedulerframework.NodeInfo,
) (*status.ScaleUpStatus, errors.AutoscalerError) {
	if len(unschedulablePods) == 0 {
		return &status.ScaleUpStatus{Result: status.ScaleUpNotTried}, nil
	}
	prs := provreqclient.ProvisioningRequestsForPods(o.client, unschedulablePods)
	prs = provreqclient.FilterOutProvisioningClass(prs, v1.ProvisioningClassBestEffortAtomicScaleUp)
	if len(prs) == 0 {
		return &status.ScaleUpStatus{Result: status.ScaleUpNotTried}, nil
	}
	// Pick 1 ProvisioningRequest.
	pr := prs[0]

	o.context.ClusterSnapshot.Fork()
	defer o.context.ClusterSnapshot.Revert()

	// For provisioning requests, unschedulablePods are actually all injected pods. Some may even be schedulable!
	actuallyUnschedulablePods, err := o.filterOutSchedulable(unschedulablePods)
	if err != nil {
		conditions.AddOrUpdateCondition(pr, v1.Provisioned, metav1.ConditionFalse, conditions.FailedToCheckCapacityReason, conditions.FailedToCheckCapacityMsg, metav1.Now())
		if _, updateErr := o.client.UpdateProvisioningRequest(pr.ProvisioningRequest); updateErr != nil {
			klog.Errorf("failed to add Provisioned=false condition to ProvReq %s/%s, err: %v", pr.Namespace, pr.Name, updateErr)
		}
		return status.UpdateScaleUpError(&status.ScaleUpStatus{}, errors.NewAutoscalerError(errors.InternalError, "error during ScaleUp: %s", err.Error()))
	}

	if len(actuallyUnschedulablePods) == 0 {
		// Nothing to do here - everything fits without scale-up.
		conditions.AddOrUpdateCondition(pr, v1.Provisioned, metav1.ConditionTrue, conditions.CapacityIsFoundReason, conditions.CapacityIsFoundMsg, metav1.Now())
		if _, updateErr := o.client.UpdateProvisioningRequest(pr.ProvisioningRequest); updateErr != nil {
			klog.Errorf("failed to add Provisioned=true condition to ProvReq %s/%s, err: %v", pr.Namespace, pr.Name, updateErr)
			return status.UpdateScaleUpError(&status.ScaleUpStatus{}, errors.NewAutoscalerError(errors.InternalError, "capacity available, but failed to admit workload: %s", updateErr.Error()))
		}
		return &status.ScaleUpStatus{Result: status.ScaleUpNotNeeded}, nil
	}

	st, err := o.scaleUpOrchestrator.ScaleUp(actuallyUnschedulablePods, nodes, daemonSets, nodeInfos, true)
	if err == nil && st.Result == status.ScaleUpSuccessful {
		// Happy path - all is well.
		conditions.AddOrUpdateCondition(pr, v1.Provisioned, metav1.ConditionTrue, conditions.CapacityIsProvisionedReason, conditions.CapacityIsProvisionedMsg, metav1.Now())
		if _, updateErr := o.client.UpdateProvisioningRequest(pr.ProvisioningRequest); updateErr != nil {
			klog.Errorf("failed to add Provisioned=true condition to ProvReq %s/%s, err: %v", pr.Namespace, pr.Name, updateErr)
			return st, errors.NewAutoscalerError(errors.InternalError, "scale up requested, but failed to admit workload: %s", updateErr.Error())
		}
		return st, nil
	}

	// We are not happy with the results.
	conditions.AddOrUpdateCondition(pr, v1.Provisioned, metav1.ConditionFalse, conditions.CapacityIsNotFoundReason, "Capacity is not found, CA will try to find it later.", metav1.Now())
	if _, updateErr := o.client.UpdateProvisioningRequest(pr.ProvisioningRequest); updateErr != nil {
		klog.Errorf("failed to add Provisioned=false condition to ProvReq %s/%s, err: %v", pr.Namespace, pr.Name, updateErr)
	}
	if err != nil {
		return status.UpdateScaleUpError(&status.ScaleUpStatus{}, errors.NewAutoscalerError(errors.InternalError, "error during ScaleUp: %s", err.Error()))
	}
	return st, nil
}

func (o *bestEffortAtomicProvClass) filterOutSchedulable(pods []*apiv1.Pod) ([]*apiv1.Pod, error) {
	statuses, _, err := o.injector.TrySchedulePods(o.context.ClusterSnapshot, pods, scheduling.ScheduleAnywhere, false)
	if err != nil {
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
	return unschedulablePods, nil

}
