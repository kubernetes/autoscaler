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
	"fmt"
	"net/netip"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v9"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"

	"sigs.k8s.io/cloud-provider-azure/pkg/consts"
	"sigs.k8s.io/cloud-provider-azure/pkg/provider/securitygroup"
	fnutil "sigs.k8s.io/cloud-provider-azure/pkg/util/collectionutil"
	"sigs.k8s.io/cloud-provider-azure/pkg/util/iputil"
)

var (
	ErrSetBothLoadBalancerSourceRangesAndAllowedIPRanges = fmt.Errorf(
		"cannot set both spec.LoadBalancerSourceRanges and service annotation %s", consts.ServiceAnnotationAllowedIPRanges,
	)
)

type AccessControl struct {
	logger   logr.Logger
	svc      *v1.Service
	sgHelper *securitygroup.RuleHelper

	// immutable pre-compute states.
	SourceRanges                           []netip.Prefix
	AllowedIPRanges                        []netip.Prefix
	AllowedServiceTags                     []string
	invalidRanges                          []string
	securityRuleDestinationPortsByProtocol map[armnetwork.SecurityRuleProtocol][]int32
}

type accessControlOptions struct {
	EventEmitter K8sEventEmitter
}

var defaultAccessControlOptions = accessControlOptions{
	EventEmitter: noopEventEmitter,
}

type AccessControlOption func(*accessControlOptions)

func WithEventEmitter(emitter K8sEventEmitter) AccessControlOption {
	return func(o *accessControlOptions) {
		o.EventEmitter = emitter
	}
}

func NewAccessControl(logger logr.Logger, svc *v1.Service, sg *armnetwork.SecurityGroup, opts ...AccessControlOption) (*AccessControl, error) {
	logger = logger.WithName("AccessControl").WithValues("security-group", ptr.To(sg.Name))

	options := defaultAccessControlOptions
	for _, opt := range opts {
		opt(&options)
	}
	eventEmitter := options.EventEmitter

	sgHelper, err := securitygroup.NewSecurityGroupHelper(logger, sg)
	if err != nil {
		logger.Error(err, "Failed to initialize RuleHelper")
		return nil, err
	}
	sourceRanges, invalidSourceRanges, err := SourceRanges(svc)
	if err != nil {
		logger.Error(err, "Failed to parse SourceRange configuration")

		// Backward compatibility: no error but emit a warning event.
		eventEmitter(svc, v1.EventTypeWarning, "InvalidSourceRanges", EventMessageOfInvalidSourceRanges(invalidSourceRanges))
	}
	allowedIPRanges, invalidAllowedIPRanges, err := AllowedIPRanges(svc)
	if err != nil {
		logger.Error(err, "Failed to parse AllowedIPRanges configuration")

		// Backward compatibility: no error but emit a warning event.
		eventEmitter(svc, v1.EventTypeWarning, "InvalidAllowedIPRanges", EventMessageOfInvalidAllowedIPRanges(invalidAllowedIPRanges))
	}
	allowedServiceTags := AllowedServiceTags(svc)
	securityRuleDestinationPortsByProtocol, err := SecurityRuleDestinationPortsByProtocol(svc)
	if err != nil {
		logger.Error(err, "Failed to parse service Spec.Ports")
		return nil, err
	}
	if len(sourceRanges) > 0 && len(allowedIPRanges) > 0 {
		logger.Error(ErrSetBothLoadBalancerSourceRangesAndAllowedIPRanges, "Forbidden configuration")
		return nil, ErrSetBothLoadBalancerSourceRangesAndAllowedIPRanges
	}

	if len(sourceRanges) > 0 && len(allowedServiceTags) > 0 {
		logger.Info(
			"Service is using both of spec.loadBalancerSourceRanges and annotation service.beta.kubernetes.io/azure-allowed-service-tags",
		)
		// Backward compatibility: emit a warning event.
		// It won't work as expected if both are used at the same time.
		eventEmitter(svc, v1.EventTypeWarning, "ConflictConfiguration", EventMessageOfConflictLoadBalancerSourceRangesAndAllowedIPRanges())
	}

	return &AccessControl{
		logger:                                 logger,
		svc:                                    svc,
		sgHelper:                               sgHelper,
		SourceRanges:                           sourceRanges,
		AllowedIPRanges:                        allowedIPRanges,
		AllowedServiceTags:                     allowedServiceTags,
		invalidRanges:                          append(invalidSourceRanges, invalidAllowedIPRanges...),
		securityRuleDestinationPortsByProtocol: securityRuleDestinationPortsByProtocol,
	}, nil
}

// IsAllowFromInternet returns true if the given service is allowed to be accessed from internet.
// To be specific,
// 1. For all types of LB, it returns false if the given service is specified with `service tags` or `not allowed all IP ranges`, including invalid IP ranges.
// 2. For internal LB, it returns true iff the given service is explicitly specified with `allowed all IP ranges`. Refer: https://github.com/kubernetes-sigs/cloud-provider-azure/issues/698
func (ac *AccessControl) IsAllowFromInternet() bool {
	if len(ac.AllowedServiceTags) > 0 {
		return false
	}
	if len(ac.SourceRanges) > 0 && !iputil.IsPrefixesAllowAll(ac.SourceRanges) {
		return false
	}
	if len(ac.AllowedIPRanges) > 0 && !iputil.IsPrefixesAllowAll(ac.AllowedIPRanges) {
		return false
	}
	if len(ac.invalidRanges) > 0 {
		return false
	}
	if !IsInternal(ac.svc) {
		return true
	}
	// Internal LB with explicit allowedAll IP ranges is allowed to be accessed from internet.
	return len(ac.AllowedIPRanges) > 0 || len(ac.SourceRanges) > 0
}

// DenyAllExceptSourceRanges returns true if it needs to block any VNet traffic not on the allow list.
// By default, NSG allow traffic from the VNet.
func (ac *AccessControl) DenyAllExceptSourceRanges() bool {
	var (
		annotationEnabled      = strings.EqualFold(ac.svc.Annotations[consts.ServiceAnnotationDenyAllExceptLoadBalancerSourceRanges], "true")
		sourceRangeSpecified   = len(ac.SourceRanges) > 0 || len(ac.AllowedIPRanges) > 0
		invalidRangesSpecified = len(ac.invalidRanges) > 0
	)
	return (annotationEnabled && sourceRangeSpecified) || invalidRangesSpecified
}

// AllowedIPv4Ranges returns the IPv4 ranges that are allowed to access the LoadBalancer.
func (ac *AccessControl) AllowedIPv4Ranges() []netip.Prefix {
	var rv []netip.Prefix
	for _, cidr := range ac.SourceRanges {
		if cidr.Addr().Is4() {
			rv = append(rv, cidr)
		}
	}
	for _, cidr := range ac.AllowedIPRanges {
		if cidr.Addr().Is4() {
			rv = append(rv, cidr)
		}
	}
	return rv
}

// AllowedIPv6Ranges returns the IPv6 ranges that are allowed to access the LoadBalancer.
func (ac *AccessControl) AllowedIPv6Ranges() []netip.Prefix {
	var rv []netip.Prefix
	for _, cidr := range ac.SourceRanges {
		if cidr.Addr().Is6() {
			rv = append(rv, cidr)
		}
	}
	for _, cidr := range ac.AllowedIPRanges {
		if cidr.Addr().Is6() {
			rv = append(rv, cidr)
		}
	}
	return rv
}

// PatchSecurityGroup checks and adds rules for the given destination IP addresses.
func (ac *AccessControl) PatchSecurityGroup(dstIPv4Addresses, dstIPv6Addresses []netip.Addr) error {
	logger := ac.logger.WithName("PatchSecurityGroup")

	var (
		allowedIPRanges    = append(ac.AllowedIPv4Ranges(), ac.AllowedIPv6Ranges()...)
		allowedServiceTags = ac.AllowedServiceTags
	)
	if ac.IsAllowFromInternet() {
		allowedFromIPRanges := iputil.IsPrefixesAllowAll(ac.AllowedIPRanges)
		allowedFromSourceRanges := iputil.IsPrefixesAllowAll(ac.SourceRanges)
		if !allowedFromIPRanges && !allowedFromSourceRanges {
			// If it's allowed from IP Ranges or Source Ranges, skip adding the internet service tag.
			allowedServiceTags = append(allowedServiceTags, securitygroup.ServiceTagInternet)
		}
	}

	{
		// Aggregate allowed IP ranges.
		ipRanges := iputil.AggregatePrefixes(allowedIPRanges)
		if len(ipRanges) != len(allowedIPRanges) {
			logger.Info("Overlapping IP ranges detected", "allowed-ip-ranges", allowedIPRanges, "aggregated-ip-ranges", ipRanges)
		}
		allowedIPRanges = ipRanges
	}
	var allowedIPv4Ranges, allowedIPv6Ranges = iputil.GroupPrefixesByFamily(allowedIPRanges)

	logger.V(10).Info("Start patching",
		"num-allowed-ipv4-ranges", len(allowedIPv4Ranges),
		"num-allowed-ipv6-ranges", len(allowedIPv6Ranges),
		"num-allowed-service-tags", len(allowedServiceTags),
	)

	protocols := []armnetwork.SecurityRuleProtocol{
		armnetwork.SecurityRuleProtocolTCP,
		armnetwork.SecurityRuleProtocolUDP,
		armnetwork.SecurityRuleProtocolAsterisk,
	}

	for _, protocol := range protocols {
		dstPorts, found := ac.securityRuleDestinationPortsByProtocol[protocol]
		if !found {
			continue
		}
		if len(dstIPv4Addresses) > 0 {
			for _, tag := range allowedServiceTags {
				err := ac.sgHelper.AddRuleForAllowedServiceTag(tag, protocol, dstIPv4Addresses, dstPorts)
				if err != nil {
					return fmt.Errorf("add rule for allowed service tag on IPv4: %w", err)
				}
			}

			if len(allowedIPv4Ranges) > 0 {
				err := ac.sgHelper.AddRuleForAllowedIPRanges(allowedIPv4Ranges, protocol, dstIPv4Addresses, dstPorts)
				if err != nil {
					return fmt.Errorf("add rule for allowed IP ranges on IPv4: %w", err)
				}
			}
		}
		if len(dstIPv6Addresses) > 0 {
			for _, tag := range allowedServiceTags {
				err := ac.sgHelper.AddRuleForAllowedServiceTag(tag, protocol, dstIPv6Addresses, dstPorts)
				if err != nil {
					return fmt.Errorf("add rule for allowed service tag on IPv6: %w", err)
				}
			}

			if len(allowedIPv6Ranges) > 0 {
				err := ac.sgHelper.AddRuleForAllowedIPRanges(allowedIPv6Ranges, protocol, dstIPv6Addresses, dstPorts)
				if err != nil {
					return fmt.Errorf("add rule for allowed IP ranges on IPv6: %w", err)
				}
			}
		}
	}

	if ac.DenyAllExceptSourceRanges() {
		if len(dstIPv4Addresses) > 0 {
			if err := ac.sgHelper.AddRuleForDenyAll(dstIPv4Addresses); err != nil {
				return fmt.Errorf("add rule for deny all on IPv4: %w", err)
			}
		}
		if len(dstIPv6Addresses) > 0 {
			if err := ac.sgHelper.AddRuleForDenyAll(dstIPv6Addresses); err != nil {
				return fmt.Errorf("add rule for deny all on IPv6: %w", err)
			}
		}
	}

	logger.V(10).Info("Completed patching")

	return nil
}

// CleanSecurityGroup removes the given IP addresses from the SecurityGroup.
func (ac *AccessControl) CleanSecurityGroup(
	dstIPv4Addresses, dstIPv6Addresses []netip.Addr,
	retainPortRanges map[armnetwork.SecurityRuleProtocol][]int32,
) error {
	logger := ac.logger.WithName("CleanSecurityGroup").
		WithValues("num-dst-ipv4-addresses", len(dstIPv4Addresses)).
		WithValues("num-dst-ipv6-addresses", len(dstIPv6Addresses))
	logger.V(10).Info("Start cleaning")

	var (
		ipv4Prefixes = fnutil.Map(func(addr netip.Addr) string { return addr.String() }, dstIPv4Addresses)
		ipv6Prefixes = fnutil.Map(func(addr netip.Addr) string { return addr.String() }, dstIPv6Addresses)
	)

	protocols := []armnetwork.SecurityRuleProtocol{
		armnetwork.SecurityRuleProtocolTCP,
		armnetwork.SecurityRuleProtocolUDP,
		armnetwork.SecurityRuleProtocolAsterisk,
	}

	for _, protocol := range protocols {
		retainDstPorts := retainPortRanges[protocol]
		if err := ac.sgHelper.RemoveDestinationFromRules(protocol, ipv4Prefixes, retainDstPorts); err != nil {
			logger.Error(err, "Failed to remove IPv4 destination from rules")
			return err
		}

		if err := ac.sgHelper.RemoveDestinationFromRules(protocol, ipv6Prefixes, retainDstPorts); err != nil {
			logger.Error(err, "Failed to remove IPv6 destination from rules")
			return err
		}
	}

	logger.V(10).Info("Completed cleaning")
	return nil
}

// RetainSecurityGroup retains the given destination IP addresses from the SecurityGroup.
func (ac *AccessControl) RetainSecurityGroup(dstIPv4Addresses, dstIPv6Addresses []netip.Addr) error {
	logger := ac.logger.WithName("RetainSecurityGroup").
		WithValues("num-dst-ipv4-addresses", len(dstIPv4Addresses)).
		WithValues("num-dst-ipv6-addresses", len(dstIPv6Addresses))
	logger.V(10).Info("Start retaining")
	defer logger.V(10).Info("Completed retaining")

	var (
		ipv4Prefixes     = fnutil.Map(func(addr netip.Addr) string { return addr.String() }, dstIPv4Addresses)
		ipv6Prefixes     = fnutil.Map(func(addr netip.Addr) string { return addr.String() }, dstIPv6Addresses)
		retainedPrefixes = append(ipv4Prefixes, ipv6Prefixes...)
	)

	if err := ac.sgHelper.RetainDestinationFromRules(retainedPrefixes); err != nil {
		logger.Error(err, "Failed to retain destination from rules")
		return err
	}

	return nil
}

// SecurityGroup returns the SecurityGroup object with patched rules and indicates if the rules had been changed.
// There are mainly two operations to alter the SecurityGroup:
// 1. `PatchSecurityGroup`: Add rules for the given destination IP addresses.
// 2. `CleanSecurityGroup`: Remove the given destination IP addresses from all rules.
// It would return unchanged SecurityGroup and `false` if the operations undo each other.
func (ac *AccessControl) SecurityGroup() (*armnetwork.SecurityGroup, bool, error) {
	return ac.sgHelper.SecurityGroup()
}

// SecurityRuleDestinationPortsByProtocol returns the service ports grouped by SecurityGroup protocol.
func SecurityRuleDestinationPortsByProtocol(svc *v1.Service) (map[armnetwork.SecurityRuleProtocol][]int32, error) {
	rv := make(map[armnetwork.SecurityRuleProtocol][]int32)
	for _, port := range svc.Spec.Ports {
		protocol, err := securitygroup.ProtocolFromKubernetes(port.Protocol)
		if err != nil {
			return nil, err
		}

		var p int32
		if consts.IsK8sServiceDisableLoadBalancerFloatingIP(svc) {
			p = port.NodePort
		} else {
			p = port.Port
		}

		rv[protocol] = append(rv[protocol], p)
	}
	return rv, nil
}
