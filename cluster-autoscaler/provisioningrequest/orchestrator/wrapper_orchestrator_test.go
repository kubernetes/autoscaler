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
	"testing"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/apis/provisioningrequest/autoscaling.x-k8s.io/v1"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/estimator"
	ca_processors "k8s.io/autoscaler/cluster-autoscaler/processors"
	"k8s.io/autoscaler/cluster-autoscaler/processors/status"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
)

const (
	provisioningRequestErrorMsg = "provisioningRequestError"
	regularPodsErrorMsg         = "regularPodsError"
)

func TestWrapperScaleUp(t *testing.T) {
	o := WrapperOrchestrator{
		provReqOrchestrator: &fakeScaleUp{provisioningRequestErrorMsg},
		podsOrchestrator:    &fakeScaleUp{regularPodsErrorMsg},
	}
	regularPods := []*apiv1.Pod{
		BuildTestPod("pod-1", 1, 100),
		BuildTestPod("pod-2", 1, 100),
	}
	provReqPods := []*apiv1.Pod{
		BuildTestPod("pr-pod-1", 1, 100),
		BuildTestPod("pr-pod-2", 1, 100),
	}
	for _, pod := range provReqPods {
		pod.Annotations[v1.ProvisioningRequestPodAnnotationKey] = "true"
	}
	unschedulablePods := append(regularPods, provReqPods...)
	_, err := o.ScaleUp(unschedulablePods, nil, nil, nil, false)
	assert.Equal(t, err.Error(), provisioningRequestErrorMsg)
	_, err = o.ScaleUp(unschedulablePods, nil, nil, nil, false)
	assert.Equal(t, err.Error(), regularPodsErrorMsg)
}

type fakeScaleUp struct {
	errorMsg string
}

func (f *fakeScaleUp) ScaleUp(
	unschedulablePods []*apiv1.Pod,
	nodes []*apiv1.Node,
	daemonSets []*appsv1.DaemonSet,
	nodeInfos map[string]*framework.NodeInfo,
	allOrNothing bool,
) (*status.ScaleUpStatus, errors.AutoscalerError) {
	return nil, errors.NewAutoscalerError(errors.InternalError, f.errorMsg)
}

func (f *fakeScaleUp) Initialize(
	autoscalingContext *context.AutoscalingContext,
	processors *ca_processors.AutoscalingProcessors,
	clusterStateRegistry *clusterstate.ClusterStateRegistry,
	estimatorBuilder estimator.EstimatorBuilder,
	taintConfig taints.TaintConfig,
) {
}

func (f *fakeScaleUp) ScaleUpToNodeGroupMinSize(
	nodes []*apiv1.Node,
	nodeInfos map[string]*framework.NodeInfo,
) (*status.ScaleUpStatus, errors.AutoscalerError) {
	return nil, nil
}
