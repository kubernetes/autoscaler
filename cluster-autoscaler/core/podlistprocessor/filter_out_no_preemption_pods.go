/*
Copyright 2023 The Kubernetes Authors.

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

package podlistprocessor

import (
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	core_utils "k8s.io/autoscaler/cluster-autoscaler/core/utils"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
)

type filterOutNoPreemptionPodsListProcessor struct {
	clusterSnapshot clustersnapshot.ClusterSnapshot
}

func NewFilterOutNoPreemptionPodsListProcessor(scheduledPods []*apiv1.Pod, allNodes []*apiv1.Node) (*filterOutNoPreemptionPodsListProcessor, error) {
	f := filterOutNoPreemptionPodsListProcessor{
		clusterSnapshot: clustersnapshot.NewDefaultClusterSnapshot(),
	}

	if err := core_utils.InitializeClusterSnapshot(f.clusterSnapshot, allNodes, scheduledPods); err != nil {
		return nil, err
	}
	return &f, nil
}

func (p *filterOutNoPreemptionPodsListProcessor) Update(_ []*apiv1.Pod, _ []*apiv1.Node) error {
	return nil
}

func (p *filterOutNoPreemptionPodsListProcessor) Process(
	context *context.AutoscalingContext,
	unschedulablePods []*apiv1.Pod) ([]*apiv1.Pod, error) {
	return []*apiv1.Pod{}, nil
}

func (p *filterOutNoPreemptionPodsListProcessor) CleanUp() {

}
