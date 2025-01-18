/*
Copyright 2025 The Kubernetes Authors.

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

package pods

import apiv1 "k8s.io/api/core/v1"

// ScaleUpEnforcer can force scale up even if all pods are new or MaxNodesTotal was achieved.
type ScaleUpEnforcer interface {
	ShouldForceScaleUp(unschedulablePods []*apiv1.Pod) bool
}

// NoOpScaleUpEnforcer returns false by default in case of ProvisioningRequests disabled.
type NoOpScaleUpEnforcer struct {
}

// NewDefaultScaleUpEnforcer creates an instance of ScaleUpEnforcer.
func NewDefaultScaleUpEnforcer() ScaleUpEnforcer {
	return &NoOpScaleUpEnforcer{}
}

// ShouldForceScaleUp returns false by default.
func (p *NoOpScaleUpEnforcer) ShouldForceScaleUp(unschedulablePods []*apiv1.Pod) bool {
	return false
}
