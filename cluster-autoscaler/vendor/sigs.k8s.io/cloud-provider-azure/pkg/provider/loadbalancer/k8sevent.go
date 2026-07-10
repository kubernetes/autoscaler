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

package loadbalancer

import (
	"fmt"

	"sigs.k8s.io/cloud-provider-azure/pkg/consts"

	"k8s.io/apimachinery/pkg/runtime"
)

type K8sEventEmitter func(obj runtime.Object, eventType, reason, message string)

func noopEventEmitter(_ runtime.Object, _, _, _ string) {}

func EventMessageOfInvalidSourceRanges(sourceRanges []string) string {
	return fmt.Sprintf(
		"Found invalid spec.LoadBalancerSourceRanges %q, ignoring and adding a default DenyAll rule in security group.",
		sourceRanges,
	)
}

func EventMessageOfInvalidAllowedIPRanges(allowedIPRanges []string) string {
	return fmt.Sprintf("Found invalid %s %q, ignoring and adding a default DenyAll rule in security group.",
		consts.ServiceAnnotationAllowedIPRanges,
		allowedIPRanges,
	)
}

func EventMessageOfConflictLoadBalancerSourceRangesAndAllowedIPRanges() string {
	return fmt.Sprintf(
		"Please use annotation %s instead of spec.loadBalancerSourceRanges while using %s annotation at the same time.",
		consts.ServiceAnnotationAllowedIPRanges,
		consts.ServiceAnnotationAllowedServiceTags,
	)
}
