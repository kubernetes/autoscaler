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

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2022-07-01/network"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"
	utilnet "k8s.io/utils/net"
	"k8s.io/utils/pointer"

	azcache "sigs.k8s.io/cloud-provider-azure/pkg/cache"
	"sigs.k8s.io/cloud-provider-azure/pkg/consts"
)

var strToExtendedLocationType = map[string]network.ExtendedLocationTypes{
	"edgezone": network.EdgeZone,
}

// lockMap used to lock on entries
type lockMap struct {
	sync.Mutex
	mutexMap map[string]*sync.Mutex
}

// NewLockMap returns a new lock map
func newLockMap() *lockMap {
	return &lockMap{
		mutexMap: make(map[string]*sync.Mutex),
	}
}

// LockEntry acquires a lock associated with the specific entry
func (lm *lockMap) LockEntry(entry string) {
	lm.Lock()
	// check if entry does not exists, then add entry
	mutex, exists := lm.mutexMap[entry]
	if !exists {
		mutex = &sync.Mutex{}
		lm.mutexMap[entry] = mutex
	}
	lm.Unlock()
	mutex.Lock()
}

// UnlockEntry release the lock associated with the specific entry
func (lm *lockMap) UnlockEntry(entry string) {
	lm.Lock()
	defer lm.Unlock()

	mutex, exists := lm.mutexMap[entry]
	if !exists {
		return
	}
	mutex.Unlock()
}

func getContextWithCancel() (context.Context, context.CancelFunc) {
	return context.WithCancel(context.Background())
}

func convertMapToMapPointer(origin map[string]string) map[string]*string {
	newly := make(map[string]*string)
	for k, v := range origin {
		value := v
		newly[k] = &value
	}
	return newly
}

func parseTags(tags string, tagsMap map[string]string) map[string]*string {
	formatted := make(map[string]*string)

	if tags != "" {
		kvs := strings.Split(tags, consts.TagsDelimiter)
		for _, kv := range kvs {
			res := strings.Split(kv, consts.TagKeyValueDelimiter)
			if len(res) != 2 {
				klog.Warningf("parseTags: error when parsing key-value pair %s, would ignore this one", kv)
				continue
			}
			k, v := strings.TrimSpace(res[0]), strings.TrimSpace(res[1])
			if k == "" {
				klog.Warning("parseTags: empty key, ignoring this key-value pair")
				continue
			}
			formatted[k] = pointer.String(v)
		}
	}

	if len(tagsMap) > 0 {
		for key, value := range tagsMap {
			key, value := strings.TrimSpace(key), strings.TrimSpace(value)
			if key == "" {
				klog.Warningf("parseTags: empty key, ignoring this key-value pair")
				continue
			}

			if found, k := findKeyInMapCaseInsensitive(formatted, key); found && k != key {
				klog.V(4).Infof("parseTags: found identical keys: %s from tags and %s from tagsMap (case-insensitive), %s will replace %s", k, key, key, k)
				delete(formatted, k)
			}
			formatted[key] = pointer.String(value)
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

func (az *Cloud) reconcileTags(currentTagsOnResource, newTags map[string]*string) (reconciledTags map[string]*string, changed bool) {
	var systemTags []string
	systemTagsMap := make(map[string]*string)

	if az.SystemTags != "" {
		systemTags = strings.Split(az.SystemTags, consts.TagsDelimiter)
		for i := 0; i < len(systemTags); i++ {
			systemTags[i] = strings.TrimSpace(systemTags[i])
		}

		for _, systemTag := range systemTags {
			systemTagsMap[systemTag] = pointer.String("")
		}
	}

	// if the systemTags is not set, just add/update new currentTagsOnResource and not delete old currentTagsOnResource
	for k, v := range newTags {
		found, key := findKeyInMapCaseInsensitive(currentTagsOnResource, k)

		if !found {
			currentTagsOnResource[k] = v
			changed = true
		} else if !strings.EqualFold(pointer.StringDeref(v, ""), pointer.StringDeref(currentTagsOnResource[key], "")) {
			currentTagsOnResource[key] = v
			changed = true
		}
	}

	// if the systemTags is set, delete the old currentTagsOnResource
	if len(systemTagsMap) > 0 {
		for k := range currentTagsOnResource {
			if _, ok := newTags[k]; !ok {
				if found, _ := findKeyInMapCaseInsensitive(systemTagsMap, k); !found {
					klog.V(2).Infof("reconcileTags: delete tag %s: %s", k, pointer.StringDeref(currentTagsOnResource[k], ""))
					delete(currentTagsOnResource, k)
					changed = true
				}
			}
		}
	}

	return currentTagsOnResource, changed
}

func (az *Cloud) getVMSetNamesSharingPrimarySLB() sets.Set[string] {
	vmSetNames := make([]string, 0)
	if az.NodePoolsWithoutDedicatedSLB != "" {
		vmSetNames = strings.Split(az.Config.NodePoolsWithoutDedicatedSLB, consts.VMSetNamesSharingPrimarySLBDelimiter)
		for i := 0; i < len(vmSetNames); i++ {
			vmSetNames[i] = strings.ToLower(strings.TrimSpace(vmSetNames[i]))
		}
	}

	return sets.New(vmSetNames...)
}

func getExtendedLocationTypeFromString(extendedLocationType string) network.ExtendedLocationTypes {
	extendedLocationType = strings.ToLower(extendedLocationType)
	if val, ok := strToExtendedLocationType[extendedLocationType]; ok {
		return val
	}
	return network.EdgeZone
}

func getServiceAdditionalPublicIPs(service *v1.Service) ([]string, error) {
	if service == nil {
		return nil, nil
	}

	result := []string{}
	if val, ok := service.Annotations[consts.ServiceAnnotationAdditionalPublicIPs]; ok {
		pips := strings.Split(strings.TrimSpace(val), ",")
		for _, pip := range pips {
			ip := strings.TrimSpace(pip)
			if ip == "" {
				continue // skip empty string
			}

			if net.ParseIP(ip) == nil {
				return nil, fmt.Errorf("%s is not a valid IP address", ip)
			}

			result = append(result, ip)
		}
	}

	return result, nil
}

func getNodePrivateIPAddress(node *v1.Node, isIPv6 bool) string {
	for _, nodeAddress := range node.Status.Addresses {
		if strings.EqualFold(string(nodeAddress.Type), string(v1.NodeInternalIP)) &&
			utilnet.IsIPv6String(nodeAddress.Address) == isIPv6 {
			klog.V(6).Infof("getNodePrivateIPAddress: node %s, ip %s", node.Name, nodeAddress.Address)
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

func removeDuplicatedSecurityRules(rules []network.SecurityRule) []network.SecurityRule {
	ruleNames := make(map[string]bool)
	for i := len(rules) - 1; i >= 0; i-- {
		if _, ok := ruleNames[pointer.StringDeref(rules[i].Name, "")]; ok {
			klog.Warningf("Found duplicated rule %s, will be removed.", pointer.StringDeref(rules[i].Name, ""))
			rules = append(rules[:i], rules[i+1:]...)
		}
		ruleNames[pointer.StringDeref(rules[i].Name, "")] = true
	}
	return rules
}

func getVMSSVMCacheKey(resourceGroup, vmssName string) string {
	cacheKey := strings.ToLower(fmt.Sprintf("%s/%s", resourceGroup, vmssName))
	return cacheKey
}

// isNodeInVMSSVMCache check whether nodeName is in vmssVMCache
func isNodeInVMSSVMCache(nodeName string, vmssVMCache *azcache.TimedCache) bool {
	if vmssVMCache == nil {
		return false
	}

	var isInCache bool

	vmssVMCache.Lock.Lock()
	defer vmssVMCache.Lock.Unlock()

	for _, entry := range vmssVMCache.Store.List() {
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

func extractVmssVMName(name string) (string, string, error) {
	split := strings.SplitAfter(name, consts.VMSSNameSeparator)
	if len(split) < 2 {
		klog.V(3).Infof("Failed to extract vmssVMName %q", name)
		return "", "", ErrorNotVmssInstance
	}

	ssName := strings.Join(split[0:len(split)-1], "")
	// removing the trailing `vmssNameSeparator` since we used SplitAfter
	ssName = ssName[:len(ssName)-1]
	instanceID := split[len(split)-1]
	return ssName, instanceID, nil
}

// isServiceDualStack checks if a Service is dual-stack or not.
func isServiceDualStack(svc *v1.Service) bool {
	return len(svc.Spec.IPFamilies) == 2
}

// getIPFamiliesEnabled checks if IPv4, IPv6 are enabled according to svc.Spec.IPFamilies.
func getIPFamiliesEnabled(svc *v1.Service) (v4Enabled bool, v6Enabled bool) {
	for _, ipFamily := range svc.Spec.IPFamilies {
		if ipFamily == v1.IPv4Protocol {
			v4Enabled = true
		} else if ipFamily == v1.IPv6Protocol {
			v6Enabled = true
		}
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
	if service.Annotations == nil {
		service.Annotations = map[string]string{}
	}
	isIPv6 := net.ParseIP(ip).To4() == nil
	service.Annotations[consts.ServiceAnnotationLoadBalancerIPDualStack[isIPv6]] = ip
}

func getServicePIPName(service *v1.Service, isIPv6 bool) string {
	if service == nil {
		return ""
	}

	if !isServiceDualStack(service) {
		return service.Annotations[consts.ServiceAnnotationPIPNameDualStack[false]]
	}

	return service.Annotations[consts.ServiceAnnotationPIPNameDualStack[isIPv6]]
}

func getServicePIPPrefixID(service *v1.Service, isIPv6 bool) string {
	if service == nil {
		return ""
	}

	if !isServiceDualStack(service) {
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
		return fmt.Sprintf("%s-%s", resource, v6Suffix)
	}
	return resource
}

// isFIPIPv6 checks if the frontend IP configuration is of IPv6.
func (az *Cloud) isFIPIPv6(fip *network.FrontendIPConfiguration, pipResourceGroup string, isInternal bool) (isIPv6 bool, err error) {
	pips, err := az.listPIP(pipResourceGroup, azcache.CacheReadTypeDefault)
	if err != nil {
		return false, fmt.Errorf("isFIPIPv6: failed to list pip: %w", err)
	}
	if isInternal {
		if fip.FrontendIPConfigurationPropertiesFormat != nil {
			if fip.FrontendIPConfigurationPropertiesFormat.PrivateIPAddressVersion != "" {
				return fip.FrontendIPConfigurationPropertiesFormat.PrivateIPAddressVersion == network.IPv6, nil
			}
			return net.ParseIP(pointer.StringDeref(fip.FrontendIPConfigurationPropertiesFormat.PrivateIPAddress, "")).To4() == nil, nil
		}
		klog.Errorf("Checking IP Family of frontend IP configuration %q of internal Service but its"+
			" FrontendIPConfigurationPropertiesFormat is nil. It's considered to be IPv4",
			pointer.StringDeref(fip.Name, ""))
		return
	}
	var fipPIPID string
	if fip.FrontendIPConfigurationPropertiesFormat != nil && fip.FrontendIPConfigurationPropertiesFormat.PublicIPAddress != nil {
		fipPIPID = pointer.StringDeref(fip.FrontendIPConfigurationPropertiesFormat.PublicIPAddress.ID, "")
	}
	for _, pip := range pips {
		id := pointer.StringDeref(pip.ID, "")
		if !strings.EqualFold(fipPIPID, id) {
			continue
		}
		if pip.PublicIPAddressPropertiesFormat != nil {
			// First check PublicIPAddressVersion, then IPAddress
			if pip.PublicIPAddressPropertiesFormat.PublicIPAddressVersion == network.IPv6 ||
				net.ParseIP(pointer.StringDeref(pip.PublicIPAddressPropertiesFormat.IPAddress, "")).To4() == nil {
				isIPv6 = true
				break
			}
		}
		break
	}
	return isIPv6, nil
}

// getResourceIDPrefix returns a substring from the provided one between beginning and the last "/".
func getResourceIDPrefix(id string) string {
	idx := strings.LastIndexByte(id, '/')
	if idx == -1 {
		return id // Should not happen
	}
	return id[:idx]
}

func getLBNameFromBackendPoolID(backendPoolID string) (string, error) {
	matches := backendPoolIDRE.FindStringSubmatch(backendPoolID)
	if len(matches) != 2 {
		return "", fmt.Errorf("backendPoolID %q is in wrong format", backendPoolID)
	}

	return matches[1], nil
}

func countNICsOnBackendPool(backendPool network.BackendAddressPool) int {
	if backendPool.BackendAddressPoolPropertiesFormat == nil ||
		backendPool.BackendIPConfigurations == nil {
		return 0
	}

	return len(*backendPool.BackendIPConfigurations)
}

func countIPsOnBackendPool(backendPool network.BackendAddressPool) int {
	if backendPool.BackendAddressPoolPropertiesFormat == nil ||
		backendPool.LoadBalancerBackendAddresses == nil {
		return 0
	}

	var ipsCount int
	for _, loadBalancerBackendAddress := range *backendPool.LoadBalancerBackendAddresses {
		if loadBalancerBackendAddress.LoadBalancerBackendAddressPropertiesFormat != nil &&
			pointer.StringDeref(loadBalancerBackendAddress.IPAddress, "") != "" {
			ipsCount++
		}
	}

	return ipsCount
}
