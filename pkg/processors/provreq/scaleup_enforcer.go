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

package provreq

import (
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/processors/pods"
)

// ProvisioningRequestScaleUpEnforcer forces scale up if there is any unschedulable pod that belongs to ProvisioningRequest.
type ProvisioningRequestScaleUpEnforcer struct {
}

// NewProvisioningRequestScaleUpEnforcer creates a ProvisioningRequest scale up enforcer.
func NewProvisioningRequestScaleUpEnforcer() pods.ScaleUpEnforcer {
	return &ProvisioningRequestScaleUpEnforcer{}
}

// ShouldForceScaleUp forces scale up if there is any unschedulable pod that belongs to ProvisioningRequest.
func (p *ProvisioningRequestScaleUpEnforcer) ShouldForceScaleUp(unschedulablePods []*apiv1.Pod) bool {
	for _, pod := range unschedulablePods {
		if _, ok := provisioningRequestName(pod); ok {
			return true
		}
	}
	return false
}
