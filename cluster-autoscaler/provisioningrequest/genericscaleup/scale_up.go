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
	"k8s.io/autoscaler/cluster-autoscaler/processors/status"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/apis/autoscaling.x-k8s.io/v1beta1"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/provreqclient"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
)

type scaleUpMode struct {
	client provreqclient.ProvisioningRequestClient
}

func New(client provreqclient.ProvisioningRequestClient) *scaleUpMode {
	return &scaleUpMode{client}
}

func (s *scaleUpMode) ScaleUp(
	unschedulablePods []*apiv1.Pod,
	nodes []*apiv1.Node,
	daemonSets []*appsv1.DaemonSet,
	nodeInfos map[string]*schedulerframework.NodeInfo,
) (*status.ScaleUpStatus, errors.AutoscalerError) {
	if len(unschedulablePods) == 0 {
		return &status.ScaleUpStatus{}, nil
	}
	if _, err := provreqclient.VerifyProvisioningRequestClass(s.client, unschedulablePods, v1beta1.ProvisioningClassGenericScaleUp); err != nil {
		return status.UpdateScaleUpError(&status.ScaleUpStatus{}, errors.NewAutoscalerError(errors.InternalError, err.Error()))
	}
	//TODO(yaroslava): implement
	return nil, nil
}
