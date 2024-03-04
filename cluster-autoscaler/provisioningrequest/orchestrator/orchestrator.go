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

package orchestrator

import (
	"errors"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/estimator"
	"k8s.io/autoscaler/cluster-autoscaler/processors/status"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/apis/autoscaling.x-k8s.io/v1beta1"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/checkcapacity"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/conditions"
	provreq_pods "k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/pods"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/provreqclient"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/scheduling"
	ca_errors "k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"

	ca_processors "k8s.io/autoscaler/cluster-autoscaler/processors"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
)

type scaleUpMode interface {
	ScaleUp([]*apiv1.Pod, []*apiv1.Node, []*appsv1.DaemonSet,
		map[string]*schedulerframework.NodeInfo) (*status.ScaleUpStatus, ca_errors.AutoscalerError)
}

// provReqOrchestrator is an orchestrator that contains orchestrators for all supported Provisioning Classes.
type provReqOrchestrator struct {
	initialized  bool
	context      *context.AutoscalingContext
	client       provreqclient.ProvisioningRequestClient
	injector     *scheduling.HintingSimulator
	scaleUpModes []scaleUpMode
}

// New return new orchestrator.
func New(kubeConfig *rest.Config) (*provReqOrchestrator, error) {
	client, err := provreqclient.NewProvisioningRequestClient(kubeConfig)
	if err != nil {
		return nil, err
	}

	return &provReqOrchestrator{client: client}, nil
}

// Initialize initialize orchestrator.
func (o *provReqOrchestrator) Initialize(
	autoscalingContext *context.AutoscalingContext,
	processors *ca_processors.AutoscalingProcessors,
	clusterStateRegistry *clusterstate.ClusterStateRegistry,
	estimatorBuilder estimator.EstimatorBuilder,
	taintConfig taints.TaintConfig,
) {
	o.initialized = true
	checkCapacityMode := checkcapacity.New(autoscalingContext, o.client, o.injector)
	o.scaleUpModes = []scaleUpMode{checkCapacityMode}
}

// ScaleUp run ScaleUp for each Provisionining Class. As of now, CA pick one ProvisioningRequest,
// so only one ProvisioningClass return non empty scaleUp result.
// In case we implement multiple ProvisioningRequest ScaleUp, the funtion should return combined status
func (o *provReqOrchestrator) ScaleUp(
	unschedulablePods []*apiv1.Pod,
	nodes []*apiv1.Node,
	daemonSets []*appsv1.DaemonSet,
	nodeInfos map[string]*schedulerframework.NodeInfo,
) (*status.ScaleUpStatus, ca_errors.AutoscalerError) {
	var status *status.ScaleUpStatus
	var combinedError error
	o.context.ClusterSnapshot.Fork()
	defer o.context.ClusterSnapshot.Revert()
	o.bookCapacity()
	for _, scaleUpMode := range o.scaleUpModes {
		st, err := scaleUpMode.ScaleUp(unschedulablePods, nodes, daemonSets, nodeInfos)
		errors.Join(combinedError, err)
		if st != nil {
			status = st
		}
	}
	return status, ca_errors.ToAutoscalerError(ca_errors.InternalError, combinedError)
}

// ScaleUpToNodeGroupMinSize doesn't have implementation for ProvisioningRequest Orchestrator.
func (o *provReqOrchestrator) ScaleUpToNodeGroupMinSize(
	nodes []*apiv1.Node,
	nodeInfos map[string]*schedulerframework.NodeInfo,
) (*status.ScaleUpStatus, ca_errors.AutoscalerError) {
	return nil, nil
}

func (o *provReqOrchestrator) bookCapacity() error {
	provReqs, err := o.client.ProvisioningRequests()
	if err != nil {
		return fmt.Errorf("couldn't fetch ProvisioningRequests in the cluster: %v", err)
	}
	podsToCreate := []*apiv1.Pod{}
	for _, provReq := range provReqs {
		if conditions.ShouldCapacityBeBooked(provReq) {
			pods, err := provreq_pods.PodsForProvisioningRequest(provReq)
			if err != nil {
				// ClusterAutoscaler was able to create pods before, so we shouldn't have error here.
				// If there is an error, mark PR as invalid, because we won't be able to book capacity
				// for it anyway.
				conditions.AddOrUpdateCondition(provReq, v1beta1.Failed, metav1.ConditionTrue, conditions.FailedToBookCapacityReason, fmt.Sprintf("Couldn't create pods, err: %v", err), metav1.Now())
				continue
			}
			podsToCreate = append(podsToCreate, pods...)
		}
	}
	if len(podsToCreate) == 0 {
		return nil
	}
	// scheduling the pods to reserve capacity for provisioning request with BookCapacity condition
	if _, _, err = o.injector.TrySchedulePods(o.context.ClusterSnapshot, podsToCreate, scheduling.ScheduleAnywhere, false); err != nil {
		klog.Warningf("Error during capacity booking: %v", err)
	}
	return nil
}
