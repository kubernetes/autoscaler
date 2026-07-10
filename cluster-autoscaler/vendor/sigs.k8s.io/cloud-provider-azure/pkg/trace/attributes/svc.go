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

package attributes

import (
	"go.opentelemetry.io/otel/attribute"
	v1 "k8s.io/api/core/v1"

	"sigs.k8s.io/cloud-provider-azure/pkg/consts"
)

func FeatureOfService(svc *v1.Service) []attribute.KeyValue {
	hasAnnotation := func(key string) bool {
		if svc.Annotations == nil {
			return false
		}
		v, ok := svc.Annotations[key]
		return ok && v != ""
	}

	return []attribute.KeyValue{
		attribute.Bool("annotations.allowed_service_tags", hasAnnotation(consts.ServiceAnnotationAllowedServiceTags)),
		attribute.Bool("annotations.allowed_ip_ranges", hasAnnotation(consts.ServiceAnnotationAllowedIPRanges)),
		attribute.Bool("annotations.internal_load_balancer", hasAnnotation(consts.ServiceAnnotationLoadBalancerInternal)),
		attribute.Bool("annotations.load_balancer_source_ranges", hasAnnotation(v1.AnnotationLoadBalancerSourceRangesKey)),
		attribute.Bool("annotations.additional_public_ips", hasAnnotation(consts.ServiceAnnotationAdditionalPublicIPs)),
		attribute.Bool("annotations.pip.name", hasAnnotation(consts.ServiceAnnotationPIPNameDualStack[false])),
		attribute.Bool("annotations.pip.prefix", hasAnnotation(consts.ServiceAnnotationPIPPrefixIDDualStack[false])),
		attribute.Bool("spec.load_balancer_source_ranges", len(svc.Spec.LoadBalancerSourceRanges) > 0),
	}
}
