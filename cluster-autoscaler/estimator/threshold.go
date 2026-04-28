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
	// NodeLimit return the node limit and optional debug string
	NodeLimit(cloudprovider.NodeGroup, EstimationContext) (int, string)
	// DurationLimit return the estimation duration limit and optional debug string
	DurationLimit(cloudprovider.NodeGroup, EstimationContext) (time.Duration, string)
	// Name returns the name of the Threshold
	Name() string
}
