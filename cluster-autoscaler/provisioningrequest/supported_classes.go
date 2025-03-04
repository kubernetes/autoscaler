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

package provisioningrequest

import (
	"strings"

	v1 "k8s.io/autoscaler/cluster-autoscaler/apis/provisioningrequest/autoscaling.x-k8s.io/v1"
	"k8s.io/klog/v2"
)

const (
	// CheckCapacityProcessorInstanceKey is a a key for ProvReq's Parameters.
	// Value for this key defines the processor instance name used to filter ProvReqs
	// and if not empty, it should match CheckCapacityProcessorInstance defined in CA's options.
	// Unrecommended: Until CA 1.35, ProvReqs with this value as prefix in their class will be also processed.
	CheckCapacityProcessorInstanceKey = "processorInstance"
)

// SupportedProvisioningClass verifies if the ProvisioningRequest with the given checkCapacityProcessorInstance is supported.
func SupportedProvisioningClass(pr *v1.ProvisioningRequest, checkCapacityProcessorInstance string) bool {
	if pr.Spec.ProvisioningClassName == v1.ProvisioningClassBestEffortAtomicScaleUp {
		if checkCapacityProcessorInstance != "" {
			// If processor instance is set, BestEffortAtomicScaleUp should not be processed.
			return false
		}
		return true
	}

	return SupportedCheckCapacityClass(pr, checkCapacityProcessorInstance)
}

// SupportedCheckCapacityClass verifies if the check capacity ProvisioningRequest with the given checkCapacityProcessorInstance is supported.
func SupportedCheckCapacityClass(pr *v1.ProvisioningRequest, checkCapacityProcessorInstance string) bool {
	provisioningClassName := pr.Spec.ProvisioningClassName
	processorInstance := string(pr.Spec.Parameters[CheckCapacityProcessorInstanceKey])

	if checkCapacityProcessorInstance == "" {
		if processorInstance != "" {
			// Processor instance should match
			return false
		}
		// If instance setting not set, just check the name
		return provisioningClassName == v1.ProvisioningClassCheckCapacity
	}

	if processorInstance != "" {
		// If both instances exist, check if they match and the provisioningClassName
		return checkCapacityProcessorInstance == processorInstance && provisioningClassName == v1.ProvisioningClassCheckCapacity
	}

	// If instances not set, check the prefix of provisioningClassName
	if !strings.HasPrefix(provisioningClassName, checkCapacityProcessorInstance) {
		return false
	}
	klog.Warningf("ProvReq %s/%s has prefixed provisioningClassName %q that is not recommended and will be removed in CA 1.35. Parameters should be used instead", pr.Namespace, pr.Name, provisioningClassName)

	return provisioningClassName[len(checkCapacityProcessorInstance):] == v1.ProvisioningClassCheckCapacity
}
