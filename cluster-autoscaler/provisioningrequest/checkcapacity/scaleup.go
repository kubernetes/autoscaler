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
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/processors/status"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/apis/autoscaling.x-k8s.io/v1beta1"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/conditions"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/provreqclient"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/provreqwrapper"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/scheduling"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"

	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
)

type scaleUpMode struct {
	context  *context.AutoscalingContext
	client   provreqclient.ProvisioningRequestClient
	injector *scheduling.HintingSimulator
}

// New create check-capacity scale-up mode.
func New(
	autoscalingContext *context.AutoscalingContext,
	client provreqclient.ProvisioningRequestClient,
	injector *scheduling.HintingSimulator,
) *scaleUpMode {
	return &scaleUpMode{autoscalingContext, client, injector}
}

// ScaleUp return if there is capacity in the cluster for pods from ProvisioningRequest.
func (o *scaleUpMode) ScaleUp(
	unschedulablePods []*apiv1.Pod,
	nodes []*apiv1.Node,
	daemonSets []*appsv1.DaemonSet,
	nodeInfos map[string]*schedulerframework.NodeInfo,
) (*status.ScaleUpStatus, errors.AutoscalerError) {
	if len(unschedulablePods) == 0 {
		return &status.ScaleUpStatus{}, nil
	}
	if _, err := o.verifyProvisioningRequestClass(unschedulablePods); err != nil {
		return status.UpdateScaleUpError(&status.ScaleUpStatus{}, errors.NewAutoscalerError(errors.InternalError, err.Error()))
	}

	o.context.ClusterSnapshot.Fork()
	defer o.context.ClusterSnapshot.Revert()

	scaleUpIsSuccessful, err := o.scaleUp(unschedulablePods)
	if err != nil {
		return status.UpdateScaleUpError(&status.ScaleUpStatus{}, errors.NewAutoscalerError(errors.InternalError, "error during ScaleUp: %s", err.Error()))
	}
	if scaleUpIsSuccessful {
		return &status.ScaleUpStatus{Result: status.ScaleUpSuccessful}, nil
	}
	return &status.ScaleUpStatus{Result: status.ScaleUpNoOptionsAvailable}, nil
}

// Assuming that all unschedulable pods comes from one ProvisioningRequest.
func (o *scaleUpMode) scaleUp(unschedulablePods []*apiv1.Pod) (bool, error) {
	provReq, err := o.client.ProvisioningRequest(unschedulablePods[0].Namespace, unschedulablePods[0].OwnerReferences[0].Name)
	if err != nil {
		return false, fmt.Errorf("failed retrive ProvisioningRequest from unscheduled pods, err: %v", err)
	}
	st, _, err := o.injector.TrySchedulePods(o.context.ClusterSnapshot, unschedulablePods, scheduling.ScheduleAnywhere, true)
	if len(st) < len(unschedulablePods) || err != nil {
		conditions.AddOrUpdateCondition(provReq, v1beta1.Provisioned, metav1.ConditionFalse, conditions.CapacityIsNotFoundReason, "Capacity is not found, CA will try to find it later.", metav1.Now())
		return false, err
	}
	conditions.AddOrUpdateCondition(provReq, v1beta1.Provisioned, metav1.ConditionTrue, conditions.CapacityIsFoundReason, conditions.CapacityIsFoundMsg, metav1.Now())
	return true, nil
}

// verifyPods check that all pods belong to one ProvisioningRequest that belongs to check-capacity ProvisioningRequst class.
func (o *scaleUpMode) verifyProvisioningRequestClass(unschedulablePods []*apiv1.Pod) (*provreqwrapper.ProvisioningRequest, error) {
	provReq, err := o.client.ProvisioningRequest(unschedulablePods[0].Namespace, unschedulablePods[0].OwnerReferences[0].Name)
	if err != nil {
		return nil, fmt.Errorf("failed retrive ProvisioningRequest from unscheduled pods, err: %v", err)
	}
	if provReq.V1Beta1().Spec.ProvisioningClassName != v1beta1.ProvisioningClassCheckCapacity {
		return nil, fmt.Errorf("provisioningRequestClass is not %s", v1beta1.ProvisioningClassCheckCapacity)
	}
	for _, pod := range unschedulablePods {
		if pod.Namespace != unschedulablePods[0].Namespace {
			return nil, fmt.Errorf("pods %s and %s are from different namespaces", pod.Name, unschedulablePods[0].Name)
		}
		if pod.OwnerReferences[0].Name != unschedulablePods[0].OwnerReferences[0].Name {
			return nil, fmt.Errorf("pods %s and %s have different OwnerReference", pod.Name, unschedulablePods[0].Name)
		}
	}
	return provReq, nil
}
