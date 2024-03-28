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

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	klog "k8s.io/klog/v2"
)

type thresholdBasedEstimationLimiter struct {
	maxDuration time.Duration
	maxNodes    int
	nodes       int
	start       time.Time
	thresholds  []Threshold
}

func (tbel *thresholdBasedEstimationLimiter) StartEstimation(_ []PodEquivalenceGroup, nodeGroup cloudprovider.NodeGroup, context EstimationContext) {
	tbel.start = time.Now()
	tbel.nodes = 0
	tbel.maxNodes = 0
	tbel.maxDuration = time.Duration(0)
	for _, threshold := range tbel.thresholds {
		tbel.maxNodes = getMinLimit(tbel.maxNodes, threshold.NodeLimit(nodeGroup, context))
		tbel.maxDuration = getMinLimit(tbel.maxDuration, threshold.DurationLimit(nodeGroup, context))
	}
}

func getMinLimit[V int | time.Duration](baseLimit V, targetLimit V) V {
	if baseLimit < 0 || targetLimit < 0 {
		return -1
	}
	if (baseLimit == 0 || baseLimit > targetLimit) && targetLimit > 0 {
		return targetLimit
	}
	return baseLimit
}

func (*thresholdBasedEstimationLimiter) EndEstimation() {}

func (tbel *thresholdBasedEstimationLimiter) PermissionToAddNode() bool {
	if tbel.maxNodes < 0 || (tbel.maxNodes > 0 && tbel.nodes >= tbel.maxNodes) {
		klog.V(4).Infof("Capping binpacking after exceeding threshold of %d nodes", tbel.maxNodes)
		return false
	}
	timeDefined := tbel.maxDuration > 0 && tbel.start != time.Time{}
	if tbel.maxDuration < 0 || (timeDefined && time.Now().After(tbel.start.Add(tbel.maxDuration))) {
		klog.V(4).Infof("Capping binpacking after exceeding max duration of %v", tbel.maxDuration)
		return false
	}
	tbel.nodes++
	return true
}

// NewThresholdBasedEstimationLimiter returns an EstimationLimiter that will prevent estimation
// after either a node count of time-based threshold is reached. This is meant to prevent cases
// where binpacking of hundreds or thousands of nodes takes extremely long time rendering CA
// incredibly slow or even completely crashing it.
// Thresholds may return:
//   - negative value: no new nodes are allowed to be added if at least one threshold returns negative limit
//   - 0: no limit, thresholds with no limits will be ignored in favor of thresholds with positive or negative limits
//   - positive value: new nodes can be added and this value represents the limit
func NewThresholdBasedEstimationLimiter(thresholds []Threshold) EstimationLimiter {
	return &thresholdBasedEstimationLimiter{thresholds: thresholds}
}
