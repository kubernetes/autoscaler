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

package genericscaleup

import (
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaleup/orchestrator"
	"k8s.io/autoscaler/cluster-autoscaler/estimator"
	ca_processors "k8s.io/autoscaler/cluster-autoscaler/processors"
	"k8s.io/autoscaler/cluster-autoscaler/processors/status"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/apis/autoscaling.x-k8s.io/v1beta1"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/provreqclient"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/scheduling"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
)

type scaleUpMode struct {
	context     *context.AutoscalingContext
	injector    *scheduling.HintingSimulator
	client      provreqclient.ProvisioningRequestClient
	podsScaleUp *orchestrator.ScaleUpOrchestrator
}

// New returns new generic-scale-up mode.
func New(client provreqclient.ProvisioningRequestClient) *scaleUpMode {
	return &scaleUpMode{client: client, podsScaleUp: orchestrator.New()}
}

func (s *scaleUpMode) Initialize(
	autoscalingContext *context.AutoscalingContext,
	processors *ca_processors.AutoscalingProcessors,
	clusterStateRegistry *clusterstate.ClusterStateRegistry,
	estimatorBuilder estimator.EstimatorBuilder,
	taintConfig taints.TaintConfig,
	injector *scheduling.HintingSimulator) {
	s.context = autoscalingContext
	s.injector = injector
	s.podsScaleUp.Initialize(autoscalingContext, processors, clusterStateRegistry, estimatorBuilder, taintConfig)
}

func (s *scaleUpMode) ScaleUp(
	unschedulablePods []*apiv1.Pod,
	nodes []*apiv1.Node,
	daemonSets []*appsv1.DaemonSet,
	nodeInfos map[string]*schedulerframework.NodeInfo,
) (*status.ScaleUpStatus, errors.AutoscalerError) {
	if pr, err := provreqclient.VerifyProvisioningRequestClass(s.client, unschedulablePods, v1beta1.ProvisioningClassGenericScaleUp); err != nil {
		return status.UpdateScaleUpError(&status.ScaleUpStatus{}, errors.NewAutoscalerError(errors.InternalError, "error during ScaleUp: %s", err.Error()))
	} else if pr == nil {
		return &status.ScaleUpStatus{}, nil
	}

	s.context.ClusterSnapshot.Fork()
	defer s.context.ClusterSnapshot.Revert()

	unschedulablePods, err := s.filterSchedulable(unschedulablePods)
	if err != nil {
		return status.UpdateScaleUpError(&status.ScaleUpStatus{}, errors.NewAutoscalerError(errors.InternalError, "error during ScaleUp: %s", err.Error()))
	}

	return s.podsScaleUp.ScaleUp(unschedulablePods, nodes, daemonSets, nodeInfos, true)
}

func (s *scaleUpMode) filterSchedulable(unschedulableCandidates []*apiv1.Pod) ([]*apiv1.Pod, error) {
	statuses, _, err := s.injector.TrySchedulePods(s.context.ClusterSnapshot, unschedulableCandidates, scheduling.ScheduleAnywhere, false)
	if err != nil {
		return nil, err
	}

	scheduledPods := make(map[types.UID]bool)
	for _, status := range statuses {
		scheduledPods[status.Pod.UID] = true
	}

	var unschedulablePods []*apiv1.Pod
	for _, pod := range unschedulableCandidates {
		if !scheduledPods[pod.UID] {
			unschedulablePods = append(unschedulablePods, pod)
		}
	}
	return unschedulablePods, nil
}
