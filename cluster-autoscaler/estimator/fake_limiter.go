/*
Copyright 2022 The Kubernetes Authors.

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

package estimator

import (
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
)

type fakeEstimationLimiter struct {
	nodes    int
	maxNodes int
}

func (_ *fakeEstimationLimiter) StartEstimation([]*apiv1.Pod, cloudprovider.NodeGroup) {}
func (_ *fakeEstimationLimiter) EndEstimation()                                        {}
func (f *fakeEstimationLimiter) PermissionToAddNodes(nodes int) bool {
	f.nodes += nodes
	return f.maxNodes == 0 || f.nodes <= f.maxNodes
}

func NewFakeEstimationLimiter(maxNodes int) EstimationLimiter {
	return &fakeEstimationLimiter{
		nodes: 0,
		maxNodes: maxNodes,
	}
}
