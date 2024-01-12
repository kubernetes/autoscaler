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
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/processors/status"
	provreq_pods "k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/pods"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/provreqclient"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/provreqwrapper"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/scheduling"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"
	"k8s.io/client-go/rest"

	ca_processors "k8s.io/autoscaler/cluster-autoscaler/processors"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
)

type provisioningRequestClient interface {
	ProvisioningRequests() ([]*provreqwrapper.ProvisioningRequest, error)
	ProvisioningRequest(namespace, name string) (*provreqwrapper.ProvisioningRequest, error)
}

type provReqOrchestrator struct {
	initialized bool
	context     *context.AutoscalingContext
	client      provisioningRequestClient
	injector    *scheduling.HintingSimulator
}

func New(kubeConfig *rest.Config) (*provReqOrchestrator, error) {
	client, err := provreqclient.NewProvisioningRequestClient(kubeConfig)
	if err != nil {
		return nil, err
	}

	return &provReqOrchestrator{client: client}, nil
}

func (o *provReqOrchestrator) Initialize(
	autoscalingContext *context.AutoscalingContext,
	processors *ca_processors.AutoscalingProcessors,
	clusterStateRegistry *clusterstate.ClusterStateRegistry,
	taintConfig taints.TaintConfig,
) {
	o.initialized = true
	o.context = autoscalingContext
	o.injector = scheduling.NewHintingSimulator(autoscalingContext.PredicateChecker)
}

func (o *provReqOrchestrator) ScaleUp(
	unschedulablePods []*apiv1.Pod,
	nodes []*apiv1.Node,
	daemonSets []*appsv1.DaemonSet,
	nodeInfos map[string]*schedulerframework.NodeInfo) (*status.ScaleUpStatus, errors.AutoscalerError) {
	if !o.initialized {
		return status.UpdateScaleUpError(&status.ScaleUpStatus{}, errors.NewAutoscalerError(errors.InternalError, "ScaleUpOrchestrator is not initialized"))
	}
	if len(unschedulablePods) == 0 {
		return &status.ScaleUpStatus{}, nil
	}
	_, err := o.verifyProvisioningRequestClass(unschedulablePods)
	if err != nil {
		return status.UpdateScaleUpError(&status.ScaleUpStatus{}, errors.NewAutoscalerError(errors.InternalError, err.Error()))
	}

	o.context.ClusterSnapshot.Fork()
	defer o.context.ClusterSnapshot.Revert()
	err = o.bookCapacity()
	if err != nil {
		return nil, errors.NewAutoscalerError(errors.InternalError, err.Error())
	}
	scaleUpIsSuccessful, err := o.scaleUp(unschedulablePods)
	if err != nil {
		return status.UpdateScaleUpError(&status.ScaleUpStatus{}, errors.NewAutoscalerError(errors.InternalError, "error during ScaleUp: %s", err.Error()))
	}
	if scaleUpIsSuccessful {
		return &status.ScaleUpStatus{Result: status.ScaleUpSuccessful}, nil
	}
	return &status.ScaleUpStatus{Result: status.ScaleUpNoOptionsAvailable}, nil
}

func (o *provReqOrchestrator) ScaleUpToNodeGroupMinSize(
	nodes []*apiv1.Node,
	nodeInfos map[string]*schedulerframework.NodeInfo) (*status.ScaleUpStatus, errors.AutoscalerError) {
	return nil, nil
}

func (o *provReqOrchestrator) bookCapacity() error {
	provReqs, err := o.client.ProvisioningRequests()
	if err != nil {
		return fmt.Errorf("Couldn't fetch ProvisioningRequests in the cluster: %v", err)
	}
	podsToCreate := []*apiv1.Pod{}
	for _, provReq := range provReqs {
		if HasBookCapacityCondition(provReq) {
			pods, err := provreq_pods.PodsForProvisioningRequest(provReq)
			if err != nil {
				// ClusterAutoscaler was able to create pods before, so we shouldn't have error here.
				// If there is an error, mark PR as invalid, because we won't be able to book capacity
				// for it anyway.
				SetCondition(provReq, RejectedCondition, "Couldn't book capacity", fmt.Sprintf("couldn't create pods, err: %v", err))
				continue
			}
			podsToCreate = append(podsToCreate, pods...)
		}
	}
	if len(podsToCreate) == 0 {
		return nil
	}
	// scheduling the pods to reserve capacity for provisioning request with BookCapacity condition
	o.injector.TrySchedulePods(o.context.ClusterSnapshot, podsToCreate, scheduling.ScheduleAnywhere, false)
	return nil
}

// Assuming that all unschedulable pods comes from one ProvisioningRequest.
func (o *provReqOrchestrator) scaleUp(unschedulablePods []*apiv1.Pod) (bool, error) {
	provReq, err := o.client.ProvisioningRequest(unschedulablePods[0].Namespace, unschedulablePods[0].OwnerReferences[0].Name)
	if err != nil {
		return false, fmt.Errorf("Failed retrive ProvisioningRequest from unscheduled pods, err: %v", err)
	}
	st, _, _ := o.injector.TrySchedulePods(o.context.ClusterSnapshot, unschedulablePods, scheduling.ScheduleAnywhere, true)
	if len(st) == len(unschedulablePods) {
		SetCondition(provReq, BookCapacityCondition, "Capacity is found", "")
		return true, nil
	}
	SetCondition(provReq, PendingCondition, "Capacity is not found", "")
	return false, nil
}

// verifyPods check that all pods belong to one ProvisioningRequest that belongs to check-capacity ProvisioningRequst class.
func (o *provReqOrchestrator) verifyProvisioningRequestClass(unschedulablePods []*apiv1.Pod) (*provreqwrapper.ProvisioningRequest, error) {
	provReq, err := o.client.ProvisioningRequest(unschedulablePods[0].Namespace, unschedulablePods[0].OwnerReferences[0].Name)
	if err != nil {
		return nil, fmt.Errorf("Failed retrive ProvisioningRequest from unscheduled pods, err: %v", err)
	}
	if provReq.V1Beta1().Spec.ProvisioningClassName != CheckCapacityClass {
		return nil, fmt.Errorf("ProvisioningRequestClass is not %s", CheckCapacityClass)
	}
	for _, pod := range unschedulablePods {
		if pod.Namespace != unschedulablePods[0].Namespace {
			return nil, fmt.Errorf("Pods %s and %s are from different namespaces", pod.Name, unschedulablePods[0].Name)
		}
		if pod.OwnerReferences[0].Name != unschedulablePods[0].OwnerReferences[0].Name {
			return nil, fmt.Errorf("Pods %s and %s have different OwnerReference", pod.Name, unschedulablePods[0].Name)
		}
	}
	return provReq, nil
}
