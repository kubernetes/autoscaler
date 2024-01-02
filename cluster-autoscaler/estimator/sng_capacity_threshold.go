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
	"k8s.io/klog/v2"
)

type sngCapacityThreshold struct {
}

// NodeLimit returns maximum number of new nodes that can be added to the cluster
// based on capacity of current node group and total capacity of similar node groups. Possible return values are:
//   - -1 when this node group AND similar node groups have no available capacity
//   - 0 when context is not set. Return value of 0 means that there is no limit.
//   - Any positive number representing maximum possible number of new nodes
func (t *sngCapacityThreshold) NodeLimit(nodeGroup cloudprovider.NodeGroup, context EstimationContext) int {
	if context == nil {
		return 0
	}
	totalAvailableCapacity := t.computeNodeGroupCapacity(nodeGroup)
	for _, sng := range context.SimilarNodeGroups() {
		totalAvailableCapacity += t.computeNodeGroupCapacity(sng)
	}
	if totalAvailableCapacity <= 0 {
		return -1
	}
	return totalAvailableCapacity
}

func (t *sngCapacityThreshold) computeNodeGroupCapacity(nodeGroup cloudprovider.NodeGroup) int {
	nodeGroupTargetSize, err := nodeGroup.TargetSize()
	// Should not ever happen as only valid node groups are passed to estimator
	if err != nil {
		klog.Errorf("Error while computing available capacity of a node group %v: can't get target size of the group: %v", nodeGroup.Id(), err)
		return 0
	}
	groupCapacity := nodeGroup.MaxSize() - nodeGroupTargetSize
	if groupCapacity > 0 {
		return groupCapacity
	}
	return 0
}

// DurationLimit always returns 0 for this threshold, meaning that no limit is set.
func (t *sngCapacityThreshold) DurationLimit(cloudprovider.NodeGroup, EstimationContext) time.Duration {
	return 0
}

// NewSngCapacityThreshold returns a Threshold that can be used to limit binpacking
// by available capacity of similar node groups
func NewSngCapacityThreshold() Threshold {
	return &sngCapacityThreshold{}
}
