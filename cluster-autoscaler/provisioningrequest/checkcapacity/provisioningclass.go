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

package checkcapacity

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/apis/provisioningrequest/autoscaling.x-k8s.io/v1"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/estimator"
	"k8s.io/autoscaler/cluster-autoscaler/processors/status"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/conditions"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/provreqclient"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/provreqwrapper"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/scheduling"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"

	ca_processors "k8s.io/autoscaler/cluster-autoscaler/processors"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
)

type checkCapacityProvClass struct {
	context  *context.AutoscalingContext
	client   *provreqclient.ProvisioningRequestClient
	injector *scheduling.HintingSimulator
}

// New create check-capacity scale-up mode.
func New(
	client *provreqclient.ProvisioningRequestClient,
) *checkCapacityProvClass {
	return &checkCapacityProvClass{client: client}
}

func (o *checkCapacityProvClass) Initialize(
	autoscalingContext *context.AutoscalingContext,
	processors *ca_processors.AutoscalingProcessors,
	clusterStateRegistry *clusterstate.ClusterStateRegistry,
	estimatorBuilder estimator.EstimatorBuilder,
	taintConfig taints.TaintConfig,
	injector *scheduling.HintingSimulator,
) {
	o.context = autoscalingContext
	o.injector = injector
}

// Provision return if there is capacity in the cluster for pods from ProvisioningRequest.
func (o *checkCapacityProvClass) Provision(
	unschedulablePods []*apiv1.Pod,
	nodes []*apiv1.Node,
	daemonSets []*appsv1.DaemonSet,
	nodeInfos map[string]*schedulerframework.NodeInfo,
) (*status.ScaleUpStatus, errors.AutoscalerError) {
	if len(unschedulablePods) == 0 {
		return &status.ScaleUpStatus{Result: status.ScaleUpNotTried}, nil
	}

	prs := provreqclient.ProvisioningRequestsForPods(o.client, unschedulablePods)
	prs = provreqclient.FilterOutProvisioningClass(prs, v1.ProvisioningClassCheckCapacity)
	if len(prs) == 0 {
		return &status.ScaleUpStatus{Result: status.ScaleUpNotTried}, nil
	}
	// Pick 1 ProvisioningRequest.
	pr := prs[0]
	o.context.ClusterSnapshot.Fork()
	defer o.context.ClusterSnapshot.Revert()

	scaleUpIsSuccessful, err := o.checkcapacity(unschedulablePods, pr)
	if err != nil {
		return status.UpdateScaleUpError(&status.ScaleUpStatus{}, errors.NewAutoscalerError(errors.InternalError, "error during ScaleUp: %s", err.Error()))
	}
	if scaleUpIsSuccessful {
		return &status.ScaleUpStatus{Result: status.ScaleUpSuccessful}, nil
	}
	return &status.ScaleUpStatus{Result: status.ScaleUpNoOptionsAvailable}, nil
}

// Assuming that all unschedulable pods comes from one ProvisioningRequest.
func (o *checkCapacityProvClass) checkcapacity(unschedulablePods []*apiv1.Pod, provReq *provreqwrapper.ProvisioningRequest) (capacityAvailable bool, err error) {
	capacityAvailable = true
	st, _, err := o.injector.TrySchedulePods(o.context.ClusterSnapshot, unschedulablePods, scheduling.ScheduleAnywhere, true)
	if len(st) < len(unschedulablePods) || err != nil {
		conditions.AddOrUpdateCondition(provReq, v1.Provisioned, metav1.ConditionFalse, conditions.CapacityIsNotFoundReason, "Capacity is not found, CA will try to find it later.", metav1.Now())
		capacityAvailable = false
	} else {
		conditions.AddOrUpdateCondition(provReq, v1.Provisioned, metav1.ConditionTrue, conditions.CapacityIsFoundReason, conditions.CapacityIsFoundMsg, metav1.Now())
	}
	_, updErr := o.client.UpdateProvisioningRequest(provReq.ProvisioningRequest)
	if updErr != nil {
		return false, fmt.Errorf("failed to update Provisioned condition to ProvReq %s/%s, err: %v", provReq.Namespace, provReq.Name, updErr)
	}
	return capacityAvailable, err
}
