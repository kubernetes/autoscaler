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
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/estimator"
	"k8s.io/autoscaler/cluster-autoscaler/processors/status"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/provreqclient"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/scheduling"
	ca_errors "k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"

	ca_processors "k8s.io/autoscaler/cluster-autoscaler/processors"
)

// ProvisioningClass is an interface for ProvisioningRequests.
type ProvisioningClass interface {
	Provision([]*apiv1.Pod, []*apiv1.Node, []*appsv1.DaemonSet,
		map[string]*framework.NodeInfo) (*status.ScaleUpStatus, ca_errors.AutoscalerError)
	Initialize(*context.AutoscalingContext, *ca_processors.AutoscalingProcessors, *clusterstate.ClusterStateRegistry,
		estimator.EstimatorBuilder, taints.TaintConfig, *scheduling.HintingSimulator)
}

// provReqOrchestrator is an orchestrator that contains orchestrators for all supported Provisioning Classes.
type provReqOrchestrator struct {
	initialized         bool
	context             *context.AutoscalingContext
	client              *provreqclient.ProvisioningRequestClient
	injector            *scheduling.HintingSimulator
	provisioningClasses []ProvisioningClass
}

// New return new orchestrator.
func New(client *provreqclient.ProvisioningRequestClient, classes []ProvisioningClass) *provReqOrchestrator {
	return &provReqOrchestrator{
		client:              client,
		provisioningClasses: classes,
	}
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
	o.context = autoscalingContext
	o.injector = scheduling.NewHintingSimulator(autoscalingContext.PredicateChecker)
	for _, mode := range o.provisioningClasses {
		mode.Initialize(autoscalingContext, processors, clusterStateRegistry, estimatorBuilder, taintConfig, o.injector)
	}
}

// ScaleUp run ScaleUp for each Provisionining Class. As of now, CA pick one ProvisioningRequest,
// so only one ProvisioningClass return non empty scaleUp result.
// In case we implement multiple ProvisioningRequest ScaleUp, the function should return combined status
func (o *provReqOrchestrator) ScaleUp(
	unschedulablePods []*apiv1.Pod,
	nodes []*apiv1.Node,
	daemonSets []*appsv1.DaemonSet,
	nodeInfos map[string]*framework.NodeInfo,
	_ bool, // Provision() doesn't use this parameter.
) (*status.ScaleUpStatus, ca_errors.AutoscalerError) {
	if !o.initialized {
		return &status.ScaleUpStatus{}, ca_errors.ToAutoscalerError(ca_errors.InternalError, fmt.Errorf("provisioningrequest.Orchestrator is not initialized"))
	}

	o.context.ClusterSnapshot.Fork()
	defer o.context.ClusterSnapshot.Revert()

	// unschedulablePods pods should belong to one ProvisioningClass, so only one provClass should try to ScaleUp.
	for _, provClass := range o.provisioningClasses {
		st, err := provClass.Provision(unschedulablePods, nodes, daemonSets, nodeInfos)
		if err != nil || st != nil && st.Result != status.ScaleUpNotTried {
			return st, err
		}
	}
	return &status.ScaleUpStatus{Result: status.ScaleUpNotTried}, nil
}

// ScaleUpToNodeGroupMinSize doesn't have implementation for ProvisioningRequest Orchestrator.
func (o *provReqOrchestrator) ScaleUpToNodeGroupMinSize(
	nodes []*apiv1.Node,
	nodeInfos map[string]*framework.NodeInfo,
) (*status.ScaleUpStatus, ca_errors.AutoscalerError) {
	return nil, nil
}
