/*
Copyright 2024 The Kubernetes Authors.

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

package nodes

import (
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/resource"
)

// ScaleDownContext holds the context for scaling down process
type ScaleDownContext struct {
	ActuationStatus         scaledown.ActuationStatus
	MaxNodeCountToBeRemoved int
	ResourcesLeft           resource.Limits
	ResourcesWithLimits     []string
}

// NewDefaultScaleDownSetContext returns ScaleDownContext with passed MaxNodeCountToBeRemoved
func NewDefaultScaleDownSetContext(MaxNodeCountToBeRemoved int) *ScaleDownContext {
	return &ScaleDownContext{
		MaxNodeCountToBeRemoved: MaxNodeCountToBeRemoved,
	}
}
