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

// GetNodeLimit returns available capacity of the cluster. Return value of 0 means that no limit is set.
func (l *clusterCapacityThreshold) GetNodeLimit(_ cloudprovider.NodeGroup, context *EstimationContext) int {
	if context == nil || (context.clusterMaxNodeLimit <= 0) || (context.clusterMaxNodeLimit < context.currentNodeCount) {
		return 0
	}
	return context.clusterMaxNodeLimit - context.currentNodeCount
}

// GetDurationLimit always returns 0 for this threshold, i.e. no limit is set.
func (l *clusterCapacityThreshold) GetDurationLimit() time.Duration {
	return 0
}

// NewClusterCapacityThreshold returns a variable Threshold to limit
// maximum nodes in binpacking. Value of the limit depends on maximum number of nodes
// allowed and total number of nodes currently running in the cluster
func NewClusterCapacityThreshold() Threshold {
	return &clusterCapacityThreshold{}
}
