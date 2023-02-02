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

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2021-08-01/network"
	"github.com/Azure/go-autorest/autorest/to"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"
	utilnet "k8s.io/utils/net"

	azcache "sigs.k8s.io/cloud-provider-azure/pkg/cache"
	"sigs.k8s.io/cloud-provider-azure/pkg/consts"
)

var strToExtendedLocationType = map[string]network.ExtendedLocationTypes{
	"edgezone": network.ExtendedLocationTypesEdgeZone,
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
	if _, exists := lm.mutexMap[entry]; !exists {
		lm.addEntry(entry)
	}

	lm.Unlock()
	lm.lockEntry(entry)
}

// UnlockEntry release the lock associated with the specific entry
func (lm *lockMap) UnlockEntry(entry string) {
	lm.Lock()
	defer lm.Unlock()

	if _, exists := lm.mutexMap[entry]; !exists {
		return
	}
	lm.unlockEntry(entry)
}

func (lm *lockMap) addEntry(entry string) {
	lm.mutexMap[entry] = &sync.Mutex{}
}

func (lm *lockMap) lockEntry(entry string) {
	lm.mutexMap[entry].Lock()
}

func (lm *lockMap) unlockEntry(entry string) {
	lm.mutexMap[entry].Unlock()
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
			formatted[k] = to.StringPtr(v)
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
			formatted[key] = to.StringPtr(value)
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
			systemTagsMap[systemTag] = to.StringPtr("")
		}
	}

	// if the systemTags is not set, just add/update new currentTagsOnResource and not delete old currentTagsOnResource
	for k, v := range newTags {
		found, key := findKeyInMapCaseInsensitive(currentTagsOnResource, k)

		if !found {
			currentTagsOnResource[k] = v
			changed = true
		} else if !strings.EqualFold(to.String(v), to.String(currentTagsOnResource[key])) {
			currentTagsOnResource[key] = v
			changed = true
		}
	}

	// if the systemTags is set, delete the old currentTagsOnResource
	if len(systemTagsMap) > 0 {
		for k := range currentTagsOnResource {
			if _, ok := newTags[k]; !ok {
				if found, _ := findKeyInMapCaseInsensitive(systemTagsMap, k); !found {
					delete(currentTagsOnResource, k)
					changed = true
				}
			}
		}
	}

	return currentTagsOnResource, changed
}

func (az *Cloud) getVMSetNamesSharingPrimarySLB() sets.String {
	vmSetNames := make([]string, 0)
	if az.NodePoolsWithoutDedicatedSLB != "" {
		vmSetNames = strings.Split(az.Config.NodePoolsWithoutDedicatedSLB, consts.VMSetNamesSharingPrimarySLBDelimiter)
		for i := 0; i < len(vmSetNames); i++ {
			vmSetNames[i] = strings.ToLower(strings.TrimSpace(vmSetNames[i]))
		}
	}

	return sets.NewString(vmSetNames...)
}

func getExtendedLocationTypeFromString(extendedLocationType string) network.ExtendedLocationTypes {
	extendedLocationType = strings.ToLower(extendedLocationType)
	if val, ok := strToExtendedLocationType[extendedLocationType]; ok {
		return val
	}
	return network.ExtendedLocationTypesEdgeZone
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

func getNodePrivateIPAddress(service *v1.Service, node *v1.Node) string {
	isIPV6SVC := utilnet.IsIPv6String(service.Spec.ClusterIP)
	for _, nodeAddress := range node.Status.Addresses {
		if strings.EqualFold(string(nodeAddress.Type), string(v1.NodeInternalIP)) &&
			utilnet.IsIPv6String(nodeAddress.Address) == isIPV6SVC {
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
		if _, ok := ruleNames[to.String(rules[i].Name)]; ok {
			klog.Warningf("Found duplicated rule %s, will be removed.", to.String(rules[i].Name))
			rules = append(rules[:i], rules[i+1:]...)
		}
		ruleNames[to.String(rules[i].Name)] = true
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
