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

// Threshold provides resources configuration for threshold based estimation limiter.
// Return value of 0 means that no limit is set.
type Threshold interface {
	NodeLimit(cloudprovider.NodeGroup, EstimationContext) NodeLimitResult
	DurationLimit(cloudprovider.NodeGroup, EstimationContext) DurationLimitResult
}

// NodeLimitResult encapsulates the result of a node limit evaluation.
type NodeLimitResult struct {
	// Limit is the maximum number of new nodes. -1 means no capacity, 0 means no limit.
	Limit int
	// Reason provides contextual debugging information if a limit is enforced.
	Reason string
}

// DurationLimitResult encapsulates the result of a duration limit evaluation.
type DurationLimitResult struct {
	// Duration is the estimation duration limit. 0 means no limit.
	Duration time.Duration
	// Reason provides contextual debugging information if a limit is enforced.
	Reason string
}
