/*
Copyright 2020 The Kubernetes Authors.

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

package provider

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v9"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	utilnet "k8s.io/utils/net"
	"k8s.io/utils/ptr"

	azcache "sigs.k8s.io/cloud-provider-azure/pkg/cache"
	"sigs.k8s.io/cloud-provider-azure/pkg/consts"
	"sigs.k8s.io/cloud-provider-azure/pkg/log"
)

const (
	IPVersionIPv6 bool = true
	IPVersionIPv4 bool = false
)

var strToExtendedLocationType = map[string]armnetwork.ExtendedLocationTypes{
	"edgezone": armnetwork.ExtendedLocationTypesEdgeZone,
}

func getContextWithCancel() (context.Context, context.CancelFunc) {
	return context.WithCancel(context.Background())
}

// parseTags processes and combines tags from a string and a map into a single map of string pointers.
// It handles tag parsing, trimming, and case-insensitive key conflicts.
//
// Parameters:
//   - tags: A string containing tags in the format "key1=value1,key2=value2"
//   - tagsMap: A map of string key-value pairs representing additional tags
//
// Returns:
//   - A map[string]*string where keys are tag names and values are pointers to tag values
//
// The function prioritizes tags from tagsMap over those from the tags string in case of conflicts.
// It logs warnings for parsing errors and empty keys, and info messages for case-insensitive key replacements.
//
// XXX: return error instead of logging; decouple tag parsing and tag application
func parseTags(tags string, tagsMap map[string]string) map[string]*string {
	logger := log.Background().WithName("parseTags")
	formatted := make(map[string]*string)

	if tags != "" {
		kvs := strings.Split(tags, consts.TagsDelimiter)
		for _, kv := range kvs {
			res := strings.Split(kv, consts.TagKeyValueDelimiter)
			if len(res) != 2 {
				klog.Warningf("parseTags: error when parsing key-value pair %s, would ignore this one", kv)
				continue
			}
			// Avoid generate `Null` string after TrimSpace operation, (e.g. " null", " Null " -> "null"/"Null")
			// `Null` is a reserved tag value by ARM, so the leading/trailing spaces must be preserved.
			// Refer to https://github.com/kubernetes-sigs/cloud-provider-azure/issues/7048.
			k, v := strings.TrimSpace(res[0]), strings.TrimSpace(res[1])
			if strings.EqualFold(v, "null") {
				v = res[1]
			}
			if k == "" {
				klog.Warning("parseTags: empty key, ignoring this key-value pair")
				continue
			}
			formatted[k] = ptr.To(v)
		}
	}

	if len(tagsMap) > 0 {
		for k, v := range tagsMap {
			key, value := strings.TrimSpace(k), strings.TrimSpace(v)
			if strings.EqualFold(value, "null") {
				value = v
			}
			if key == "" {
				klog.Warningf("parseTags: empty key, ignoring this key-value pair")
				continue
			}

			if found, k := findKeyInMapCaseInsensitive(formatted, key); found && k != key {
				logger.V(4).Info("found identical keys from tags and tagsMap (case-insensitive), will replace", "identical keys", k, "keyFromTagsMap", key)
				delete(formatted, k)
			}
			formatted[key] = ptr.To(value)
		}
	}

	return formatted
}

func findKeyInMapCaseInsensitive(targetMap map[string]*string, key string) (bool, string) {
	for k := range targetMap {
		if strings.EqualFold(k, key) {
			return true, k
		}
	}

	return false, ""
}

// This function extends the functionality of findKeyInMapCaseInsensitive by supporting both
// exact case-insensitive key matching and prefix-based key matching in the given map.
// 1. If the key is found in the map (case-insensitively), the function returns true and the matching key in the map.
// 2. If the key's prefix is found in the map (case-insensitively), the function also returns true and the matching key in the map.
// This function is designed to enable systemTags to support prefix-based tag keys,
// allowing more flexible and efficient tag key matching.
func findKeyInMapWithPrefix(targetMap map[string]*string, key string) (bool, string) {
	for k := range targetMap {
		// use prefix-based key matching
		// use case-insensitive comparison
		if strings.HasPrefix(strings.ToLower(key), strings.ToLower(k)) {
			return true, k
		}
	}
	return false, ""
}

func (az *Cloud) reconcileTags(currentTagsOnResource, newTags map[string]*string) (reconciledTags map[string]*string, changed bool) {
	logger := log.Background().WithName("reconcileTags")
	var systemTags []string
	systemTagsMap := make(map[string]*string)

	if az.SystemTags != "" {
		systemTags = strings.Split(az.SystemTags, consts.TagsDelimiter)
		for i := 0; i < len(systemTags); i++ {
			systemTags[i] = strings.TrimSpace(systemTags[i])
		}

		for _, systemTag := range systemTags {
			systemTagsMap[systemTag] = ptr.To("")
		}
	}

	// if the systemTags is not set, just add/update new currentTagsOnResource and not delete old currentTagsOnResource
	for k, v := range newTags {
		found, key := findKeyInMapCaseInsensitive(currentTagsOnResource, k)

		if !found {
			currentTagsOnResource[k] = v
			changed = true
		} else if !strings.EqualFold(ptr.Deref(v, ""), ptr.Deref(currentTagsOnResource[key], "")) {
			currentTagsOnResource[key] = v
			changed = true
		}
	}

	// if the systemTags is set, delete the old currentTagsOnResource
	if len(systemTagsMap) > 0 {
		for k := range currentTagsOnResource {
			if _, ok := newTags[k]; !ok {
				if found, _ := findKeyInMapWithPrefix(systemTagsMap, k); !found {
					logger.V(2).Info("delete tag", "key", k, "value", ptr.Deref(currentTagsOnResource[k], ""))
					delete(currentTagsOnResource, k)
					changed = true
				}
			}
		}
	}

	return currentTagsOnResource, changed
}

func getExtendedLocationTypeFromString(extendedLocationType string) armnetwork.ExtendedLocationTypes {
	extendedLocationType = strings.ToLower(extendedLocationType)
	if val, ok := strToExtendedLocationType[extendedLocationType]; ok {
		return val
	}
	return armnetwork.ExtendedLocationTypesEdgeZone
}

func getNodePrivateIPAddress(node *v1.Node, isIPv6 bool) string {
	logger := log.Background().WithName("getNodePrivateIPAddress")
	for _, nodeAddress := range node.Status.Addresses {
		if strings.EqualFold(string(nodeAddress.Type), string(v1.NodeInternalIP)) &&
			utilnet.IsIPv6String(nodeAddress.Address) == isIPv6 {
			logger.V(6).Info("Get node private IP address", "node", node.Name, "ip", nodeAddress.Address)
			return nodeAddress.Address
		}
	}

	klog.Warningf("getNodePrivateIPAddress: empty ip found for node %s", node.Name)
	return ""
}

func getNodePrivateIPAddresses(node *v1.Node) []string {
	addresses := make([]string, 0)
	for _, nodeAddress := range node.Status.Addresses {
		if strings.EqualFold(string(nodeAddress.Type), string(v1.NodeInternalIP)) {
			addresses = append(addresses, nodeAddress.Address)
		}
	}

	return addresses
}

func getBoolValueFromServiceAnnotations(service *v1.Service, key string) bool {
	if l, found := service.Annotations[key]; found {
		return strings.EqualFold(strings.TrimSpace(l), consts.TrueAnnotationValue)
	}
	return false
}

func sameContentInSlices(s1 []string, s2 []string) bool {
	if len(s1) != len(s2) {
		return false
	}
	map1 := make(map[string]int)
	for _, s := range s1 {
		map1[s]++
	}
	for _, s := range s2 {
		if v, ok := map1[s]; !ok || v <= 0 {
			return false
		}
		map1[s]--
	}
	return true
}

func removeDuplicatedSecurityRules(rules []*armnetwork.SecurityRule) []*armnetwork.SecurityRule {
	ruleNames := make(map[string]bool)
	for i := len(rules) - 1; i >= 0; i-- {
		if _, ok := ruleNames[ptr.Deref(rules[i].Name, "")]; ok {
			klog.Warningf("Found duplicated rule %s, will be removed.", ptr.Deref(rules[i].Name, ""))
			rules = append(rules[:i], rules[i+1:]...)
		}
		ruleNames[ptr.Deref(rules[i].Name, "")] = true
	}
	return rules
}

func getVMSSVMCacheKey(resourceGroup, vmssName string) string {
	cacheKey := strings.ToLower(fmt.Sprintf("%s/%s", resourceGroup, vmssName))
	return cacheKey
}

// isNodeInVMSSVMCache check whether nodeName is in vmssVMCache
func isNodeInVMSSVMCache(nodeName string, vmssVMCache azcache.Resource) bool {
	if vmssVMCache == nil {
		return false
	}

	var isInCache bool

	vmssVMCache.Lock()
	defer vmssVMCache.Unlock()

	for _, entry := range vmssVMCache.GetStore().List() {
		if entry != nil {
			e := entry.(*azcache.AzureCacheEntry)
			e.Lock.Lock()
			data := e.Data
			if data != nil {
				data.(*sync.Map).Range(func(vmName, _ interface{}) bool {
					if vmName != nil && vmName.(string) == nodeName {
						isInCache = true
						return false
					}
					return true
				})
			}
			e.Lock.Unlock()
		}

		if isInCache {
			break
		}
	}

	return isInCache
}

// isServiceDualStack checks if a Service is dual-stack or not.
func isServiceDualStack(svc *v1.Service) bool {
	return len(svc.Spec.IPFamilies) == 2
}

// getIPFamiliesEnabled checks if IPv4, IPv6 are enabled according to svc.Spec.IPFamilies.
func getIPFamiliesEnabled(svc *v1.Service) (v4Enabled bool, v6Enabled bool) {
	for _, ipFamily := range svc.Spec.IPFamilies {
		switch ipFamily {
		case v1.IPv4Protocol:
			v4Enabled = true
		case v1.IPv6Protocol:
			v6Enabled = true
		}
	}
	if !v4Enabled && !v6Enabled {
		v4Enabled = true
	}
	return
}

// getServiceLoadBalancerIP retrieves LB IP from IPv4 annotation, then IPv6 annotation, then service.Spec.LoadBalancerIP.
func getServiceLoadBalancerIP(service *v1.Service, isIPv6 bool) string {
	if service == nil {
		return ""
	}

	if ip, ok := service.Annotations[consts.ServiceAnnotationLoadBalancerIPDualStack[isIPv6]]; ok && ip != "" {
		return ip
	}

	// Retrieve LB IP from service.Spec.LoadBalancerIP (will be deprecated)
	svcLBIP := service.Spec.LoadBalancerIP
	if (net.ParseIP(svcLBIP).To4() != nil && !isIPv6) ||
		(net.ParseIP(svcLBIP).To4() == nil && isIPv6) {
		return svcLBIP
	}
	return ""
}

func getServiceLoadBalancerIPs(service *v1.Service) []string {
	if service == nil {
		return []string{}
	}

	ips := []string{}
	if ip, ok := service.Annotations[consts.ServiceAnnotationLoadBalancerIPDualStack[false]]; ok && ip != "" {
		ips = append(ips, ip)
	}
	if ip, ok := service.Annotations[consts.ServiceAnnotationLoadBalancerIPDualStack[true]]; ok && ip != "" {
		ips = append(ips, ip)
	}
	if len(ips) != 0 {
		return ips
	}

	lbIP := service.Spec.LoadBalancerIP
	if lbIP != "" {
		ips = append(ips, lbIP)
	}

	return ips
}

// setServiceLoadBalancerIP sets LB IP to a Service
func setServiceLoadBalancerIP(service *v1.Service, ip string) {
	if service == nil {
		klog.Warning("setServiceLoadBalancerIP: Service is nil")
		return
	}
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		klog.Warningf("setServiceLoadBalancerIP: IP %q is not valid for Service %q", ip, service.Name)
		return
	}

	isIPv6 := parsedIP.To4() == nil
	if service.Annotations == nil {
		service.Annotations = map[string]string{}
	}
	service.Annotations[consts.ServiceAnnotationLoadBalancerIPDualStack[isIPv6]] = ip
}

func getServicePIPName(service *v1.Service, isIPv6 bool) string {
	if service == nil {
		return ""
	}

	if !isServiceDualStack(service) {
		v4Enabled, v6Enabled := getIPFamiliesEnabled(service)
		if isIPv6 && v6Enabled && !v4Enabled {
			if name := service.Annotations[consts.ServiceAnnotationPIPNameDualStack[true]]; name != "" {
				return name
			}
		}
		return service.Annotations[consts.ServiceAnnotationPIPNameDualStack[false]]
	}

	return service.Annotations[consts.ServiceAnnotationPIPNameDualStack[isIPv6]]
}

func getServicePIPNames(service *v1.Service) []string {
	var ips []string
	for _, ipVersion := range []bool{IPVersionIPv4, IPVersionIPv6} {
		if name := getServicePIPName(service, ipVersion); name != "" {
			ips = append(ips, name)
		}
	}
	return ips
}

func getServicePIPPrefixID(service *v1.Service, isIPv6 bool) string {
	if service == nil {
		return ""
	}

	if !isServiceDualStack(service) {
		v4Enabled, v6Enabled := getIPFamiliesEnabled(service)
		if isIPv6 && v6Enabled && !v4Enabled {
			if id := service.Annotations[consts.ServiceAnnotationPIPPrefixIDDualStack[true]]; id != "" {
				return id
			}
		}
		return service.Annotations[consts.ServiceAnnotationPIPPrefixIDDualStack[false]]
	}

	return service.Annotations[consts.ServiceAnnotationPIPPrefixIDDualStack[isIPv6]]
}

// getResourceByIPFamily returns the resource name of with IPv6 suffix when
// it is a dual-stack Service and the resource is of IPv6.
// NOTICE: For PIPs of IPv6 Services created with CCM v1.27.1, after the CCM is upgraded,
// the old PIPs will be recreated.
func getResourceByIPFamily(resource string, isDualStack, isIPv6 bool) string {
	if isDualStack && isIPv6 {
		return fmt.Sprintf("%s-%s", resource, consts.IPVersionIPv6String)
	}
	return resource
}

// isFIPIPv6 checks if the frontend IP configuration is of IPv6.
// NOTICE: isFIPIPv6 assumes the FIP is owned by the Service and it is the primary Service.
func (az *Cloud) isFIPIPv6(service *v1.Service, fip *armnetwork.FrontendIPConfiguration) (bool, error) {
	isDualStack := isServiceDualStack(service)
	if !isDualStack {
		if len(service.Spec.IPFamilies) == 0 {
			return false, nil
		}
		return service.Spec.IPFamilies[0] == v1.IPv6Protocol, nil
	}
	return managedResourceHasIPv6Suffix(ptr.Deref(fip.Name, "")), nil
}

// getResourceIDPrefix returns a substring from the provided one between beginning and the last "/".
func getResourceIDPrefix(id string) string {
	idx := strings.LastIndexByte(id, '/')
	if idx == -1 {
		return id // Should not happen
	}
	return id[:idx]
}

func getBackendPoolNameFromBackendPoolID(backendPoolID string) (string, error) {
	matches := backendPoolIDRE.FindStringSubmatch(backendPoolID)
	if len(matches) != 3 {
		return "", fmt.Errorf("backendPoolID %q is in wrong format", backendPoolID)
	}

	return matches[2], nil
}

func countNICsOnBackendPool(backendPool *armnetwork.BackendAddressPool) int {
	if backendPool.Properties == nil ||
		backendPool.Properties.BackendIPConfigurations == nil {
		return 0
	}

	return len(backendPool.Properties.BackendIPConfigurations)
}

func countIPsOnBackendPool(backendPool *armnetwork.BackendAddressPool) int {
	if backendPool.Properties == nil ||
		backendPool.Properties.LoadBalancerBackendAddresses == nil {
		return 0
	}

	var ipsCount int
	for _, loadBalancerBackendAddress := range backendPool.Properties.LoadBalancerBackendAddresses {
		if loadBalancerBackendAddress.Properties != nil &&
			ptr.Deref(loadBalancerBackendAddress.Properties.IPAddress, "") != "" {
			ipsCount++
		}
	}

	return ipsCount
}

// StringInSlice check if string in a list
func StringInSlice(s string, list []string) bool {
	for _, item := range list {
		if item == s {
			return true
		}
	}
	return false
}

// StringInSliceIgnoreCase checks if a string is in a list, ignoring case.
func StringInSliceIgnoreCase(s string, list []string) bool {
	for _, item := range list {
		if strings.EqualFold(item, s) {
			return true
		}
	}
	return false
}

// IntInSlice checks if an int is in a list
func IntInSlice(i int, list []int) bool {
	for _, item := range list {
		if item == i {
			return true
		}
	}
	return false
}

func isLocalService(service *v1.Service) bool {
	return service.Spec.ExternalTrafficPolicy == v1.ServiceExternalTrafficPolicyLocal
}

func getServiceIPFamily(service *v1.Service) string {
	if len(service.Spec.IPFamilies) > 1 {
		return consts.IPVersionDualStackString
	}
	for _, ipFamily := range service.Spec.IPFamilies {
		if ipFamily == v1.IPv6Protocol {
			return consts.IPVersionIPv6String
		}
	}
	return consts.IPVersionIPv4String
}

// getResourceGroupAndNameFromNICID parses the ip configuration ID to get the resource group and nic name.
func getResourceGroupAndNameFromNICID(ipConfigurationID string) (string, string, error) {
	logger := log.Background().WithName("getResourceGroupAndNameFromNICID")
	matches := nicIDRE.FindStringSubmatch(ipConfigurationID)
	if len(matches) != 3 {
		logger.V(4).Info("Can not extract nic name from ipConfigurationID", "ipConfigurationID", ipConfigurationID)
		return "", "", fmt.Errorf("invalid ip config ID %s", ipConfigurationID)
	}

	nicResourceGroup, nicName := matches[1], matches[2]
	if nicResourceGroup == "" || nicName == "" {
		return "", "", fmt.Errorf("invalid ip config ID %s", ipConfigurationID)
	}
	return nicResourceGroup, nicName, nil
}

func isInternalLoadBalancer(lb *armnetwork.LoadBalancer) bool {
	return strings.HasSuffix(strings.ToLower(*lb.Name), consts.InternalLoadBalancerNameSuffix)
}

// trimSuffixIgnoreCase trims the suffix from the string, case-insensitive.
// It returns the original string if the suffix is not found.
// The returning string is in lower case.
func trimSuffixIgnoreCase(str, suf string) string {
	str = strings.ToLower(str)
	suf = strings.ToLower(suf)
	if strings.HasSuffix(str, suf) {
		return strings.TrimSuffix(str, suf)
	}
	return str
}

func isEmptyLabelSelector(selector *metav1.LabelSelector) bool {
	if selector == nil {
		return true
	}
	return len(selector.MatchLabels) == 0 && len(selector.MatchExpressions) == 0
}
