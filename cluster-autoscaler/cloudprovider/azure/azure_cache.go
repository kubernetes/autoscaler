/*
Copyright 2018 The Kubernetes Authors.

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

package azure

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v5"
	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2022-08-01/compute"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/Azure/skewer"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"

	"k8s.io/klog/v2"
)

var (
	virtualMachineRE = regexp.MustCompile(`^azure://(?:.*)/providers/Microsoft.Compute/virtualMachines/(.+)$`)
)

// azureCache is used for caching cluster resources state.
//
// It is needed to:
//   - keep track of node groups (VM and VMSS types) in the cluster,
//   - keep track of instances and which node group they belong to,
//     (for VMSS it only keeps track of instanceid-to-nodegroup mapping)
//   - limit repetitive Azure API calls.
//
// It backs efficient responds to
//   - cloudprovider.NodeGroups() (= registeredNodeGroups)
//   - cloudprovider.NodeGroupForNode (via azureManager.GetNodeGroupForInstance => FindForInstance,
//     using instanceToNodeGroup and unownedInstances)
//
// CloudProvider.Refresh, called before every autoscaler loop (every 10s by defaul),
// is implemented by AzureManager.Refresh which makes the cache refresh decision,
// based on AzureManager.lastRefresh and azureCache.refreshInterval.
type azureCache struct {
	mutex     sync.Mutex
	interrupt chan struct{}
	azClient  *azClient

	// refreshInterval specifies how often azureCache needs to be refreshed.
	// The value comes from AZURE_VMSS_CACHE_TTL env var (or 1min if not specified),
	// and is also used by some other caches. Together with AzureManager.lastRefresh,
	// it is uses to decide whether a refresh is needed.
	refreshInterval time.Duration

	// Cache content.

	// resourceGroup specifies the name of the resource group that this cache tracks
	resourceGroup        string
	clusterResourceGroup string
	clusterName          string

	// enableVMsAgentPool specifies whether VMs agent pool type is supported.
	enableVMsAgentPool bool

	// vmType can be one of vmTypeVMSS (default), vmTypeStandard
	vmType string

	vmsPoolMap map[string]armcontainerservice.AgentPool // track the nodepools that're vms pool

	// scaleSets keeps the set of all known scalesets in the resource group, populated/refreshed via VMSS.List() call.
	// It is only used/populated if vmType is vmTypeVMSS (default).
	scaleSets map[string]compute.VirtualMachineScaleSet
	// virtualMachines keeps the set of all VMs in the resource group.
	// It is only used/populated if vmType is vmTypeStandard.
	virtualMachines map[string][]compute.VirtualMachine

	// registeredNodeGroups represents all known NodeGroups.
	registeredNodeGroups []cloudprovider.NodeGroup

	// instanceToNodeGroup maintains a mapping from instance Ids to nodegroups.
	// It is populated from the results of calling Nodes() on each nodegroup.
	// It is used (together with unownedInstances) when looking up the nodegroup
	// for a given instance id (see FindForInstance).
	instanceToNodeGroup map[azureRef]cloudprovider.NodeGroup

	// unownedInstance maintains a set of instance ids not belonging to any nodegroup.
	// It is used (together with instanceToNodeGroup) when looking up the nodegroup for a given instance id.
	// It is reset by invalidateUnownedInstanceCache().
	unownedInstances map[azureRef]bool

	autoscalingOptions map[azureRef]map[string]string
	skus               map[string]*skewer.Cache
}

func newAzureCache(client *azClient, cacheTTL time.Duration, config Config) (*azureCache, error) {
	cache := &azureCache{
		interrupt:            make(chan struct{}),
		azClient:             client,
		refreshInterval:      cacheTTL,
		resourceGroup:        config.ResourceGroup,
		clusterResourceGroup: config.ClusterResourceGroup,
		clusterName:          config.ClusterName,
		enableVMsAgentPool:   config.EnableVMsAgentPool,
		vmType:               config.VMType,
		vmsPoolMap:           make(map[string]armcontainerservice.AgentPool),
		scaleSets:            make(map[string]compute.VirtualMachineScaleSet),
		virtualMachines:      make(map[string][]compute.VirtualMachine),
		registeredNodeGroups: make([]cloudprovider.NodeGroup, 0),
		instanceToNodeGroup:  make(map[azureRef]cloudprovider.NodeGroup),
		unownedInstances:     make(map[azureRef]bool),
		autoscalingOptions:   make(map[azureRef]map[string]string),
		skus:                 make(map[string]*skewer.Cache),
	}

	if config.EnableDynamicInstanceList {
		cache.skus[config.Location] = &skewer.Cache{}
	}

	if err := cache.regenerate(); err != nil {
		klog.Errorf("Error while regenerating Azure cache: %v", err)
	}

	return cache, nil
}

func (m *azureCache) getVMsPoolMap() map[string]armcontainerservice.AgentPool {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	return m.vmsPoolMap
}

func (m *azureCache) getVirtualMachines() map[string][]compute.VirtualMachine {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	return m.virtualMachines
}

func (m *azureCache) getScaleSets() map[string]compute.VirtualMachineScaleSet {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	return m.scaleSets
}

// Cleanup closes the channel to signal the go routine to stop that is handling the cache
func (m *azureCache) Cleanup() {
	close(m.interrupt)
}

func (m *azureCache) regenerate() error {
	err := m.fetchAzureResources()
	if err != nil {
		return err
	}

	// Regenerate instance to node groups mapping.
	newInstanceToNodeGroupCache := make(map[azureRef]cloudprovider.NodeGroup)
	for _, ng := range m.registeredNodeGroups {
		klog.V(4).Infof("regenerate: finding nodes for node group %s", ng.Id())
		instances, err := ng.Nodes()
		if err != nil {
			return err
		}
		klog.V(4).Infof("regenerate: found nodes for node group %s: %+v", ng.Id(), instances)

		for _, instance := range instances {
			ref := azureRef{Name: instance.Id}
			newInstanceToNodeGroupCache[ref] = ng
		}
	}

	// Regenerate VMSS to autoscaling options mapping.
	newAutoscalingOptions := make(map[azureRef]map[string]string)
	for _, vmss := range m.scaleSets {
		ref := azureRef{Name: *vmss.Name}
		options := extractAutoscalingOptionsFromScaleSetTags(vmss.Tags)
		if !reflect.DeepEqual(m.getAutoscalingOptions(ref), options) {
			klog.V(4).Infof("Extracted autoscaling options from %q ScaleSet tags: %v", *vmss.Name, options)
		}
		newAutoscalingOptions[ref] = options
	}

	newSkuCache := make(map[string]*skewer.Cache)
	for location := range m.skus {
		cache, err := m.fetchSKUs(context.Background(), location)
		if err != nil {
			return err
		}
		newSkuCache[location] = cache
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.instanceToNodeGroup = newInstanceToNodeGroupCache
	m.autoscalingOptions = newAutoscalingOptions
	m.skus = newSkuCache

	// Reset unowned instances cache.
	m.unownedInstances = make(map[azureRef]bool)

	return nil
}

// fetchAzureResources retrieves and updates the cached Azure resources.
//
// This function performs the following:
// - Fetches and updates the list of Virtual Machine Scale Sets (VMSS) in the specified resource group.
// - Fetches and updates the list of Virtual Machines (VMs) and identifies the node pools they belong to.
// - Maintains a set of VMs pools and VMSS resources which helps the Cluster Autoscaler (CAS) operate on mixed node pools.
//
// Returns an error if any of the Azure API calls fail.
func (m *azureCache) fetchAzureResources() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// NOTE: this lists virtual machine scale sets, not virtual machine
	// scale set instances
	vmssResult, err := m.fetchScaleSets()
	if err != nil {
		return err
	}
	m.scaleSets = vmssResult
	// we fetch both sets of resources since CAS may operate on mixed nodepools
	vmResult, err := m.fetchVirtualMachines()
	if err != nil {
		return err
	}
	m.virtualMachines = vmResult
	if m.enableVMsAgentPool {
		vmsPoolMap, err := m.fetchVMsPools()
		if err != nil {
			return err
		}
		m.vmsPoolMap = vmsPoolMap
	}

	return nil
}

const (
	legacyAgentpoolNameTag = "poolName"
	agentpoolNameTag       = "aks-managed-poolName"
	agentpoolTypeTag       = "aks-managed-agentpool-type"
	vmsPoolType            = "VirtualMachines"
)

// fetchVirtualMachines returns the updated list of virtual machines in the config resource group using the Azure API.
func (m *azureCache) fetchVirtualMachines() (map[string][]compute.VirtualMachine, error) {
	ctx, cancel := getContextWithCancel()
	defer cancel()

	result, err := m.azClient.virtualMachinesClient.List(ctx, m.resourceGroup)
	if err != nil {
		klog.Errorf("VirtualMachinesClient.List in resource group %q failed: %v", m.resourceGroup, err)
		return nil, err.Error()
	}

	instances := make(map[string][]compute.VirtualMachine)
	for _, instance := range result {
		if instance.Tags == nil {
			continue
		}

		tags := instance.Tags
		vmPoolName := tags[agentpoolNameTag]
		// fall back to legacy tag name if not found
		if vmPoolName == nil {
			vmPoolName = tags[legacyAgentpoolNameTag]
		}
		if vmPoolName == nil {
			continue
		}

		instances[to.String(vmPoolName)] = append(instances[to.String(vmPoolName)], instance)
	}
	return instances, nil
}

// fetchScaleSets returns the updated list of scale sets in the config resource group using the Azure API.
func (m *azureCache) fetchScaleSets() (map[string]compute.VirtualMachineScaleSet, error) {
	ctx, cancel := getContextWithTimeout(vmssContextTimeout)
	defer cancel()

	result, err := m.azClient.virtualMachineScaleSetsClient.List(ctx, m.resourceGroup)
	if err != nil {
		klog.Errorf("VirtualMachineScaleSetsClient.List in resource group %q failed: %v", m.resourceGroup, err)
		return nil, err.Error()
	}

	sets := make(map[string]compute.VirtualMachineScaleSet)
	for _, vmss := range result {
		sets[*vmss.Name] = vmss
	}
	return sets, nil
}

// fetchVMsPools returns the a set of VMs pools in the cluster
func (m *azureCache) fetchVMsPools() (map[string]armcontainerservice.AgentPool, error) {
	ctx, cancel := getContextWithTimeout(vmsListRequestContextTimeout)
	defer cancel()

	if m.azClient.agentPoolClient == nil {
		return nil, fmt.Errorf("agentPoolClient is nil")
	}

	vmsPoolMap := make(map[string]armcontainerservice.AgentPool)

	pager := m.azClient.agentPoolClient.NewListPager(m.clusterResourceGroup, m.clusterName, nil)
	var aps []*armcontainerservice.AgentPool
	for pager.More() {
		resp, err := pager.NextPage(ctx)
		if err != nil {
			klog.Errorf("agentPoolClient.pager.NextPage in cluster %s resource group %s failed: %v",
				m.clusterName, m.clusterResourceGroup, err)
			return nil, err
		}
		aps = append(aps, resp.Value...)
	}

	for _, ap := range aps {
		if ap != nil && ap.Name != nil && ap.Properties != nil && ap.Properties.Type != nil &&
			*ap.Properties.Type == armcontainerservice.AgentPoolTypeVirtualMachines {
			vmsPoolMap[*ap.Name] = *ap
		}
	}

	return vmsPoolMap, nil
}

// Register registers a node group if it hasn't been registered.
func (m *azureCache) Register(nodeGroup cloudprovider.NodeGroup) bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for i := range m.registeredNodeGroups {
		if existing := m.registeredNodeGroups[i]; strings.EqualFold(existing.Id(), nodeGroup.Id()) {
			if existing.MinSize() == nodeGroup.MinSize() && existing.MaxSize() == nodeGroup.MaxSize() {
				// Node group is already registered and min/max size haven't changed, no action required.
				return false
			}
			m.registeredNodeGroups[i] = nodeGroup
			klog.V(4).Infof("Node group %q updated", nodeGroup.Id())
			m.invalidateUnownedInstanceCache()
			return true
		}
	}

	klog.V(4).Infof("Registering Node Group %q", nodeGroup.Id())

	m.registeredNodeGroups = append(m.registeredNodeGroups, nodeGroup)
	m.invalidateUnownedInstanceCache()
	return true
}

func (m *azureCache) invalidateUnownedInstanceCache() {
	klog.V(4).Info("Invalidating unowned instance cache")
	m.unownedInstances = make(map[azureRef]bool)
}

// Unregister node group. Returns true if the node group was unregistered.
func (m *azureCache) Unregister(nodeGroup cloudprovider.NodeGroup) bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	updated := make([]cloudprovider.NodeGroup, 0, len(m.registeredNodeGroups))
	changed := false
	for _, existing := range m.registeredNodeGroups {
		if strings.EqualFold(existing.Id(), nodeGroup.Id()) {
			klog.V(1).Infof("Unregistered node group %s", nodeGroup.Id())
			changed = true
			continue
		}
		updated = append(updated, existing)
	}
	m.registeredNodeGroups = updated
	return changed
}

func (m *azureCache) fetchSKUs(ctx context.Context, location string) (*skewer.Cache, error) {
	return skewer.NewCache(ctx,
		skewer.WithLocation(location),
		skewer.WithResourceClient(m.azClient.skuClient),
	)
}

func (m *azureCache) GetSKU(ctx context.Context, skuName, location string) (skewer.SKU, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	cache, ok := m.skus[location]
	if !ok {
		var err error
		cache, err = m.fetchSKUs(ctx, location)
		if err != nil {
			klog.V(1).Infof("Failed to instantiate cache, err: %v", err)
			return skewer.SKU{}, err
		}
		m.skus[location] = cache
	}

	return cache.Get(ctx, skuName, skewer.VirtualMachines, location)
}

func (m *azureCache) getRegisteredNodeGroups() []cloudprovider.NodeGroup {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	return m.registeredNodeGroups
}

func (m *azureCache) getAutoscalingOptions(ref azureRef) map[string]string {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	return m.autoscalingOptions[ref]
}

// HasInstance returns if a given instance exists in the azure cache
func (m *azureCache) HasInstance(providerID string) (bool, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	resourceID, err := convertResourceGroupNameToLower(providerID)
	if err != nil {
		// Most likely an invalid resource id, we should return an error
		// most of these shouldn't make it here do to higher level
		// validation in the HasInstance azure.cloudprovider function
		return false, err
	}

	if m.getInstanceFromCache(resourceID) != nil {
		return true, nil
	}
	// couldn't find instance in the cache, assume it's deleted
	return false, cloudprovider.ErrNotImplemented
}

// FindForInstance returns node group of the given Instance
func (m *azureCache) FindForInstance(instance *azureRef, vmType string) (cloudprovider.NodeGroup, error) {
	vmsPoolMap := m.getVMsPoolMap()
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.V(4).Infof("FindForInstance: starts, ref: %s", instance.Name)
	resourceID, err := convertResourceGroupNameToLower(instance.Name)
	klog.V(4).Infof("FindForInstance: resourceID: %s", resourceID)
	if err != nil {
		return nil, err
	}
	inst := azureRef{Name: resourceID}
	if m.unownedInstances[inst] {
		// We already know we don't own this instance. Return early and avoid
		// additional calls.
		klog.V(4).Infof("FindForInstance: Couldn't find NodeGroup of instance %q", inst)
		return nil, nil
	}

	// cluster with vmss pool only
	if vmType == vmTypeVMSS && len(vmsPoolMap) == 0 {
		if m.areAllScaleSetsUniform() {
			// Omit virtual machines not managed by vmss only in case of uniform scale set.
			if ok := virtualMachineRE.Match([]byte(inst.Name)); ok {
				klog.V(3).Infof("Instance %q is not managed by vmss, omit it in autoscaler", instance.Name)
				m.unownedInstances[inst] = true
				return nil, nil
			}
		}
	}

	if vmType == vmTypeStandard {
		// Omit virtual machines with providerID not in Azure resource ID format.
		if ok := virtualMachineRE.Match([]byte(inst.Name)); !ok {
			klog.V(3).Infof("Instance %q is not in Azure resource ID format, omit it in autoscaler", instance.Name)
			m.unownedInstances[inst] = true
			return nil, nil
		}
	}

	// Look up caches for the instance.
	klog.V(6).Infof("FindForInstance: attempting to retrieve instance %v from cache", m.instanceToNodeGroup)
	if nodeGroup := m.getInstanceFromCache(inst.Name); nodeGroup != nil {
		klog.V(4).Infof("FindForInstance: found node group %q in cache", nodeGroup.Id())
		return nodeGroup, nil
	}
	klog.V(4).Infof("FindForInstance: Couldn't find node group of instance %q", inst)
	return nil, nil
}

// isAllScaleSetsAreUniform determines if all the scale set autoscaler is monitoring are Uniform or not.
func (m *azureCache) areAllScaleSetsUniform() bool {
	for _, scaleSet := range m.scaleSets {
		if scaleSet.VirtualMachineScaleSetProperties.OrchestrationMode == compute.Flexible {
			return false
		}
	}
	return true
}

// getInstanceFromCache gets the node group from cache. Returns nil if not found.
// Should be called with lock.
func (m *azureCache) getInstanceFromCache(providerID string) cloudprovider.NodeGroup {
	for instanceID, nodeGroup := range m.instanceToNodeGroup {
		if strings.EqualFold(instanceID.GetKey(), providerID) {
			return nodeGroup
		}
	}

	return nil
}
