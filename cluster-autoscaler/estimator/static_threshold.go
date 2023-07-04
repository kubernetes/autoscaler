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

package estimator

import (
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
)

type staticThreshold struct {
	maxNodes    int
	maxDuration time.Duration
}

func (l *staticThreshold) NodeLimit(cloudprovider.NodeGroup, EstimationContext) int {
	return l.maxNodes
}

func (l *staticThreshold) DurationLimit(cloudprovider.NodeGroup, EstimationContext) time.Duration {
	return l.maxDuration
}

// NewStaticThreshold returns a Threshold that should be used to limit
// result and duration of binpacking by given static values
func NewStaticThreshold(maxNodes int, maxDuration time.Duration) Threshold {
	return &staticThreshold{
		maxNodes:    maxNodes,
		maxDuration: maxDuration,
	}
}
