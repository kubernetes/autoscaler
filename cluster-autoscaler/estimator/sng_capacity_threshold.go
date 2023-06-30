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

func (l *sngCapacityThreshold) GetNodeLimit(_ cloudprovider.NodeGroup, context *EstimationContext) int {
	var totalCapacity int
	for _, nodeGroup := range context.GetSimilarNodeGroups() {
		nodeGroupTargetSize, err := nodeGroup.TargetSize()
		// Should not ever happen as only valid node groups are passed to estimator
		if err != nil {
			klog.Errorf("Error while computing available capacity of a node group %v: can't get target size of the group", nodeGroup.Id(), err)
			continue
		}
		groupCapacity := nodeGroup.MaxSize() - nodeGroupTargetSize
		if groupCapacity > 0 {
			totalCapacity += groupCapacity
		}
	}
	return totalCapacity
}

func (l *sngCapacityThreshold) GetDurationLimit() time.Duration {
	return 0
}

// NewSngCapacityThreshold returns a Threshold that should be used to limit
// result and duration of binpacking by given static values
func NewSngCapacityThreshold() Threshold {
	return &sngCapacityThreshold{}
}
