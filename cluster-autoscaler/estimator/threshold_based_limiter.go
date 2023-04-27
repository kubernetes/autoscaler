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
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	klog "k8s.io/klog/v2"
)

type thresholdBasedEstimationLimiter struct {
	maxDuration time.Duration
	maxNodes    int
	nodes       int
	start       time.Time
}

func (tbel *thresholdBasedEstimationLimiter) StartEstimation([]*apiv1.Pod, cloudprovider.NodeGroup) {
	tbel.start = time.Now()
	tbel.nodes = 0
}

func (*thresholdBasedEstimationLimiter) EndEstimation() {}

func (tbel *thresholdBasedEstimationLimiter) PermissionToAddNode() bool {
	if tbel.maxNodes > 0 && tbel.nodes >= tbel.maxNodes {
		klog.V(4).Infof("Capping binpacking after exceeding threshold of %d nodes", tbel.maxNodes)
		return false
	}
	timeDefined := tbel.maxDuration > 0 && tbel.start != time.Time{}
	if timeDefined && time.Now().After(tbel.start.Add(tbel.maxDuration)) {
		klog.V(4).Infof("Capping binpacking after exceeding max duration of %v", tbel.maxDuration)
		return false
	}
	tbel.nodes++
	return true
}

// NewThresholdBasedEstimationLimiter returns an EstimationLimiter that will prevent estimation
// after either a node count- of time-based threshold is reached. This is meant to prevent cases
// where binpacking of hundreds or thousands of nodes takes extremely long time rendering CA
// incredibly slow or even completely crashing it.
func NewThresholdBasedEstimationLimiter(maxNodes int, maxDuration time.Duration) EstimationLimiter {
	return &thresholdBasedEstimationLimiter{
		maxNodes:    maxNodes,
		maxDuration: maxDuration,
	}
}
