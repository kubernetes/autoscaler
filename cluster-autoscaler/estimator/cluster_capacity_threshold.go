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

type clusterCapacityThreshold struct {
}

// NodeLimit returns maximum number of new nodes that can be added to the cluster
// based on its capacity. Possible return values are:
//   - -1 when cluster has no available capacity
//   - 0 when context or cluster-wide node limit is not set. Return value of 0 means that there is no limit.
//   - Any positive number representing maximum possible number of new nodes
func (l *clusterCapacityThreshold) NodeLimit(_ cloudprovider.NodeGroup, context EstimationContext) int {
	if context == nil || context.ClusterMaxNodeLimit() == 0 {
		return 0
	}
	if (context.ClusterMaxNodeLimit() < 0) || (context.ClusterMaxNodeLimit() <= context.CurrentNodeCount()) {
		return -1
	}
	return context.ClusterMaxNodeLimit() - context.CurrentNodeCount()
}

// DurationLimit always returns 0 for this threshold, meaning that no limit is set.
func (l *clusterCapacityThreshold) DurationLimit(cloudprovider.NodeGroup, EstimationContext) time.Duration {
	return 0
}

// NewClusterCapacityThreshold returns a Threshold that can be used to limit binpacking
// by available cluster capacity
func NewClusterCapacityThreshold() Threshold {
	return &clusterCapacityThreshold{}
}
