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

package securitygroup

import (
	"crypto/md5" //nolint:gosec
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v9"
	v1 "k8s.io/api/core/v1"

	"sigs.k8s.io/cloud-provider-azure/pkg/consts"
	fnutil "sigs.k8s.io/cloud-provider-azure/pkg/util/collectionutil"
	"sigs.k8s.io/cloud-provider-azure/pkg/util/iputil"
)

// IsManagedSecurityRule returns true if the security rule is managed by the cloud provider.
func IsManagedSecurityRule(r *armnetwork.SecurityRule) bool {
	if r == nil || r.Name == nil || r.Properties == nil || r.Properties.Priority == nil {
		return false
	}
	priority := *r.Properties.Priority
	return strings.HasPrefix(*r.Name, SecurityRuleNamePrefix) && consts.LoadBalancerMinimumPriority <= priority && priority <= consts.LoadBalancerMaximumPriority
}

// GenerateAllowSecurityRuleName returns the AllowInbound rule name based on the given rule properties.
func GenerateAllowSecurityRuleName(
	protocol armnetwork.SecurityRuleProtocol,
	ipFamily iputil.Family,
	srcPrefixes []string,
	dstPorts []int32,
) string {
	var ruleID string
	{
		dstPortRanges := fnutil.Map(func(p int32) string { return strconv.FormatInt(int64(p), 10) }, dstPorts)
		// Generate rule ID from protocol, source prefixes and destination port ranges.
		sort.Strings(srcPrefixes)
		sort.Strings(dstPortRanges)

		v := strings.Join([]string{
			string(protocol),
			strings.Join(srcPrefixes, ","),
			strings.Join(dstPortRanges, ","),
		}, "_")

		h := md5.New() //nolint:gosec
		h.Write([]byte(v))

		ruleID = fmt.Sprintf("%x", h.Sum(nil))
	}

	return strings.Join([]string{SecurityRuleNamePrefix, "allow", string(ipFamily), ruleID}, SecurityRuleNameSep)
}

// GenerateDenyAllSecurityRuleName returns the DenyInbound rule name based on the given rule properties.
func GenerateDenyAllSecurityRuleName(ipFamily iputil.Family) string {
	return strings.Join([]string{SecurityRuleNamePrefix, "deny-all", string(ipFamily)}, SecurityRuleNameSep)
}

// NormalizeSecurityRuleAddressPrefixes normalizes the given rule address prefixes.
func NormalizeSecurityRuleAddressPrefixes(vs []string) []string {
	// Remove redundant addresses.
	indexes := make(map[string]bool, len(vs))
	for _, v := range vs {
		indexes[v] = true
	}
	rv := make([]string, 0, len(indexes))
	for k := range indexes {
		rv = append(rv, k)
	}
	sort.Strings(rv)
	return rv
}

// NormalizeDestinationPortRanges normalizes the given destination port ranges.
func NormalizeDestinationPortRanges(dstPorts []int32) []string {
	rv := fnutil.Map(func(p int32) string { return strconv.FormatInt(int64(p), 10) }, dstPorts)
	sort.Strings(rv)
	return rv
}

// ListSourcePrefixes lists the source prefixes for the security rule.
func ListSourcePrefixes(r *armnetwork.SecurityRule) []string {
	var rv []string
	if r.Properties.SourceAddressPrefix != nil {
		rv = append(rv, *r.Properties.SourceAddressPrefix)
	}
	if r.Properties.SourceAddressPrefixes != nil {
		rv = append(rv, fnutil.Map(func(data *string) string { return *data }, r.Properties.SourceAddressPrefixes)...)
	}
	return rv
}

// ListDestinationPrefixes lists the destination prefixes for the security rule.
func ListDestinationPrefixes(r *armnetwork.SecurityRule) []string {
	var rv []string
	if r.Properties.DestinationAddressPrefix != nil && *r.Properties.DestinationAddressPrefix != "" {
		rv = append(rv, *r.Properties.DestinationAddressPrefix)
	}
	if r.Properties.DestinationAddressPrefixes != nil {
		rv = append(rv, fnutil.Map(func(key *string) string { return *key }, r.Properties.DestinationAddressPrefixes)...)
	}
	return rv
}

// SetDestinationPrefixes sets the destination prefixes for the security rule.
func SetDestinationPrefixes(r *armnetwork.SecurityRule, prefixes []string) {
	ps := NormalizeSecurityRuleAddressPrefixes(prefixes)
	if len(ps) == 1 {
		r.Properties.DestinationAddressPrefix = to.Ptr(ps[0])
		r.Properties.DestinationAddressPrefixes = nil
	} else {
		r.Properties.DestinationAddressPrefix = nil
		r.Properties.DestinationAddressPrefixes = to.SliceOfPtrs(ps...)
	}
}

// ListDestinationPortRanges lists the destination port ranges for the security rule.
func ListDestinationPortRanges(r *armnetwork.SecurityRule) ([]int32, error) {
	var values []*string
	if r.Properties.DestinationPortRange != nil {
		values = append(values, r.Properties.DestinationPortRange)
	}
	if r.Properties.DestinationPortRanges != nil {
		values = append(values, r.Properties.DestinationPortRanges...)
	}

	rv := make([]int32, 0, len(values))
	for _, v := range values {
		p, err := strconv.ParseInt(*v, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("parse port range %q: %w", *v, err)
		}
		rv = append(rv, int32(p))
	}

	return rv, nil
}

// SetDestinationPortRanges sets the destination port ranges for the security rule.
func SetDestinationPortRanges(r *armnetwork.SecurityRule, ports []int32) {
	ps := NormalizeDestinationPortRanges(ports)
	r.Properties.DestinationPortRange = nil
	r.Properties.DestinationPortRanges = to.SliceOfPtrs(ps...)
}

// SetAsteriskDestinationPortRange sets the destination port range to * for the security rule.
func SetAsteriskDestinationPortRange(r *armnetwork.SecurityRule) {
	r.Properties.DestinationPortRange = to.Ptr("*")
	r.Properties.DestinationPortRanges = nil
}

// ProtocolFromKubernetes converts the Kubernetes protocol to the Azure security rule protocol.
func ProtocolFromKubernetes(p v1.Protocol) (armnetwork.SecurityRuleProtocol, error) {
	switch p {
	case v1.ProtocolTCP:
		return armnetwork.SecurityRuleProtocolTCP, nil
	case v1.ProtocolUDP:
		return armnetwork.SecurityRuleProtocolUDP, nil
	case v1.ProtocolSCTP:
		return armnetwork.SecurityRuleProtocolAsterisk, nil
	}
	return "", fmt.Errorf("unsupported protocol %s", p)
}
