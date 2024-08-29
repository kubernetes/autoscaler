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

package loadbalancer

import (
	"errors"
	"fmt"
	"net/netip"
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"

	"sigs.k8s.io/cloud-provider-azure/pkg/consts"
)

// IsInternal returns true if the given service is internal load balancer.
func IsInternal(svc *v1.Service) bool {
	value, found := svc.Annotations[consts.ServiceAnnotationLoadBalancerInternal]
	return found && strings.ToLower(value) == "true"
}

// IsExternal returns true if the given service is external load balancer.
func IsExternal(svc *v1.Service) bool {
	return !IsInternal(svc)
}

// AllowedServiceTags returns the allowed service tags configured by user through AKS custom annotation.
func AllowedServiceTags(svc *v1.Service) ([]string, error) {
	const Sep = ","

	value, found := svc.Annotations[consts.ServiceAnnotationAllowedServiceTags]
	if !found {
		return nil, nil
	}

	tags := strings.Split(strings.TrimSpace(value), Sep)
	for i := range tags {
		tags[i] = strings.TrimSpace(tags[i])
	}
	return tags, nil
}

// AllowedIPRanges returns the allowed IP ranges configured by user through AKS custom annotations:
// service.beta.kubernetes.io/azure-allowed-ip-ranges and service.beta.kubernetes.io/load-balancer-source-ranges
func AllowedIPRanges(svc *v1.Service) ([]netip.Prefix, []string, error) {
	const (
		Sep = ","
	)
	var (
		errs          []error
		validRanges   []netip.Prefix
		invalidRanges []string
	)

	for _, key := range []string{consts.ServiceAnnotationAllowedIPRanges, v1.AnnotationLoadBalancerSourceRangesKey} {
		value, found := svc.Annotations[key]
		if !found {
			continue
		}

		var errsByKey []error
		for _, p := range strings.Split(strings.TrimSpace(value), Sep) {
			p = strings.TrimSpace(p)
			prefix, err := ParseCIDR(p)
			if err != nil {
				errsByKey = append(errsByKey, err)
				invalidRanges = append(invalidRanges, p)
			} else {
				validRanges = append(validRanges, prefix)
			}
		}
		if len(errsByKey) > 0 {
			errs = append(errs, fmt.Errorf("invalid service annotation %s:%s: %w", key, value, errors.Join(errsByKey...)))
		}
	}

	if len(errs) > 0 {
		return validRanges, invalidRanges, errors.Join(errs...)
	}
	return validRanges, invalidRanges, nil
}

// SourceRanges returns the allowed IP ranges configured by user through `spec.LoadBalancerSourceRanges`.
func SourceRanges(svc *v1.Service) ([]netip.Prefix, []string, error) {
	var (
		errs          []error
		validRanges   []netip.Prefix
		invalidRanges []string
	)
	// Read from spec
	for _, p := range svc.Spec.LoadBalancerSourceRanges {
		p = strings.TrimSpace(p)
		prefix, err := ParseCIDR(p)
		if err != nil {
			errs = append(errs, err)
			invalidRanges = append(invalidRanges, p)
		} else {
			validRanges = append(validRanges, prefix)
		}
	}
	if len(errs) > 0 {
		return validRanges, invalidRanges, fmt.Errorf("invalid service.Spec.LoadBalancerSourceRanges [%v]: %w", svc.Spec.LoadBalancerSourceRanges, errors.Join(errs...))
	}
	return validRanges, invalidRanges, nil
}

type AccessControl struct {
	svc *v1.Service

	// immutable redundant states.
	sourceRanges       []netip.Prefix
	allowedIPRanges    []netip.Prefix
	allowedServiceTags []string
	invalidRanges      []string
}

func NewAccessControl(svc *v1.Service) (*AccessControl, error) {
	logger := klog.Background().
		WithName("LoadBalancer.AccessControl").
		WithValues("service-name", svc.Name)

	sourceRanges, invalidSourceRanges, err := SourceRanges(svc)
	if err != nil {
		logger.Error(err, "Failed to parse SourceRange configuration")
	}
	allowedIPRanges, invalidAllowedIPRanges, err := AllowedIPRanges(svc)
	if err != nil {
		logger.Error(err, "Failed to parse AllowedIPRanges configuration")
	}
	allowedServiceTags, err := AllowedServiceTags(svc)
	if err != nil {
		return nil, err
	}

	return &AccessControl{
		svc:                svc,
		sourceRanges:       sourceRanges,
		allowedIPRanges:    allowedIPRanges,
		allowedServiceTags: allowedServiceTags,
		invalidRanges:      append(invalidSourceRanges, invalidAllowedIPRanges...),
	}, nil
}

// SourceRanges returns the allowed IP ranges configured by user through `spec.LoadBalancerSourceRanges` and standard annotation.
func (ac *AccessControl) SourceRanges() []netip.Prefix {
	return ac.sourceRanges
}

// AllowedIPRanges returns the allowed IP ranges configured by user through AKS custom annotation.
func (ac *AccessControl) AllowedIPRanges() []netip.Prefix {
	return ac.allowedIPRanges
}

// AllowedServiceTags returns the allowed service tags configured by user through AKS custom annotation.
func (ac *AccessControl) AllowedServiceTags() []string {
	return ac.allowedServiceTags
}

// InvalidRanges returns the invalid IP ranges provided by user in sourceRanges and allowedIPRanges.
func (ac *AccessControl) InvalidRanges() []string {
	return ac.invalidRanges
}

// IsAllowFromInternet returns true if the given service is allowed to be accessed from internet.
// To be specific,
// 1. For all types of LB, it returns false if the given service is specified with `service tags` or `not allowed all IP ranges`, including invalid IP ranges.
// 2. For internal LB, it returns true iff the given service is explicitly specified with `allowed all IP ranges`. Refer: https://github.com/kubernetes-sigs/cloud-provider-azure/issues/698
func (ac *AccessControl) IsAllowFromInternet() bool {
	if len(ac.allowedServiceTags) > 0 {
		return false
	}
	if len(ac.sourceRanges) > 0 && !IsCIDRsAllowAll(ac.sourceRanges) {
		return false
	}
	if len(ac.allowedIPRanges) > 0 && !IsCIDRsAllowAll(ac.allowedIPRanges) {
		return false
	}
	if len(ac.invalidRanges) > 0 {
		return false
	}
	if IsExternal(ac.svc) {
		return true
	}
	// Internal LB with explicit allowedAll IP ranges is allowed to be accessed from internet.
	return len(ac.allowedIPRanges) > 0 || len(ac.sourceRanges) > 0
}

// IPV4Sources returns the allowed sources for IPv4.
func (ac *AccessControl) IPV4Sources() []string {
	var rv []string

	if ac.IsAllowFromInternet() {
		rv = append(rv, "Internet")
	}
	for _, cidr := range ac.sourceRanges {
		if cidr.Addr().Is4() {
			rv = append(rv, cidr.String())
		}
	}
	for _, cidr := range ac.allowedIPRanges {
		if cidr.Addr().Is4() {
			rv = append(rv, cidr.String())
		}
	}
	rv = append(rv, ac.allowedServiceTags...)

	return rv
}

// IPV6Sources returns the allowed sources for IPv6.
func (ac *AccessControl) IPV6Sources() []string {
	var (
		rv []string
	)
	if ac.IsAllowFromInternet() {
		rv = append(rv, "Internet")
	}
	for _, cidr := range ac.sourceRanges {
		if cidr.Addr().Is6() {
			rv = append(rv, cidr.String())
		}
	}
	for _, cidr := range ac.allowedIPRanges {
		if cidr.Addr().Is6() {
			rv = append(rv, cidr.String())
		}
	}
	rv = append(rv, ac.allowedServiceTags...)

	return rv
}
