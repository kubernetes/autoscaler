/*
Copyright 2017 The Kubernetes Authors.

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
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v5"
	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2022-08-01/compute"
	"github.com/Azure/go-autorest/autorest/to"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/config/dynamic"
	klog "k8s.io/klog/v2"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
)

// VMPool represents a pool of standalone VMs with a single SKU.
// It is part of a mixed-SKU agent pool (agent pool with type `VirtualMachines`).
type VMPool struct {
	azureRef
	manager              *AzureManager
	resourceGroup        string // MC_ resource group for nodes
	clusterResourceGroup string // resource group for the cluster itself
	clusterName          string
	agentPoolName        string // the virtual machines agentpool that this VMPool belongs to
	location             string
	sku                  string // sku of the VM in the pool

	minSize           int
	maxSize           int
	curSize           int64
	sizeMutex         sync.Mutex
	lastSizeRefresh   time.Time
	sizeRefreshPeriod time.Duration

	enableDynamicInstanceList bool
}

// NewVMPool creates a new VMPool - a pool of standalone VMs of a single size.
func NewVMPool(spec *dynamic.NodeGroupSpec, am *AzureManager, agentPoolName string, sku string) (*VMPool, error) {
	if am.azClient.agentPoolClient == nil {
		return nil, fmt.Errorf("agentPoolClient is nil")
	}

	nodepool := &VMPool{
		azureRef: azureRef{
			Name: spec.Name, // in format "<agentPoolName>/<sku>"
		},

		sku:           sku,
		agentPoolName: agentPoolName,

		manager:                   am,
		resourceGroup:             am.config.ResourceGroup,
		clusterResourceGroup:      am.config.ClusterResourceGroup,
		clusterName:               am.config.ClusterName,
		location:                  am.config.Location,
		sizeRefreshPeriod:         am.azureCache.refreshInterval,
		enableDynamicInstanceList: am.config.EnableDynamicInstanceList,

		curSize: -1,
		minSize: spec.MinSize,
		maxSize: spec.MaxSize,
	}

	return nodepool, nil
}

// MinSize returns the minimum size the vmPool is allowed to scaled down
// to as provided by the node spec in --node parameter.
func (vmPool *VMPool) MinSize() int {
	return vmPool.minSize
}

// Exist is always true since we are initialized with an existing vmPool
func (vmPool *VMPool) Exist() bool {
	return true
}

// Create creates the node group on the cloud provider side.
func (vmPool *VMPool) Create() (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrAlreadyExist
}

// Delete deletes the node group on the cloud provider side.
func (vmPool *VMPool) Delete() error {
	return cloudprovider.ErrNotImplemented
}

// Autoprovisioned is always false since we are initialized with an existing agentpool
func (vmPool *VMPool) Autoprovisioned() bool {
	return false
}

// GetOptions returns NodeGroupAutoscalingOptions that should be used for this particular
// NodeGroup. Returning a nil will result in using default options.
func (vmPool *VMPool) GetOptions(defaults config.NodeGroupAutoscalingOptions) (*config.NodeGroupAutoscalingOptions, error) {
	// TODO(wenxuan): implement this method when vmPool can fully support GPU nodepool
	return nil, nil
}

// MaxSize returns the maximum size scale limit provided by --node
// parameter to the autoscaler main
func (vmPool *VMPool) MaxSize() int {
	return vmPool.maxSize
}

// TargetSize returns the current TARGET size of the node group. It is possible that the
// number is different from the number of nodes registered in Kubernetes.
func (vmPool *VMPool) TargetSize() (int, error) {
	size, err := vmPool.getVMPoolSize()
	return int(size), err
}

func (vmPool *VMPool) getAgentpoolFromAzure() (armcontainerservice.AgentPool, error) {
	ctx, cancel := getContextWithTimeout(vmsContextTimeout)
	defer cancel()
	resp, err := vmPool.manager.azClient.agentPoolClient.Get(
		ctx,
		vmPool.clusterResourceGroup,
		vmPool.clusterName,
		vmPool.agentPoolName, nil)
	if err != nil {
		return resp.AgentPool, fmt.Errorf("failed to get agentpool %s in cluster %s with error: %v",
			vmPool.agentPoolName, vmPool.clusterName, err)
	}
	return resp.AgentPool, nil
}

// getVMPoolSize returns the current size of the vmPool.
func (vmPool *VMPool) getVMPoolSize() (int64, error) {
	size, err := vmPool.getCurSize()
	if err != nil {
		klog.Errorf("Failed to get vmPool %s size with error: %s", vmPool.Name, err)
		return size, err
	}
	if size == -1 {
		klog.Errorf("vmPool %s size is -1, it is still being initialized", vmPool.Name)
		return size, fmt.Errorf("getVMPoolSize: size is -1 for vmPool %s", vmPool.Name)
	}
	return size, nil
}

// IncreaseSize increase the size through a PUT AP call.
func (vmPool *VMPool) IncreaseSize(delta int) error {
	if delta <= 0 {
		return fmt.Errorf("size increase must be positive, current delta: %d", delta)
	}

	currentSize, err := vmPool.getVMPoolSize()
	if err != nil {
		return err
	}

	if currentSize == -1 {
		return fmt.Errorf("the vmPool %s is still under initialization, skip the size increase", vmPool.Name)
	}

	if int(currentSize)+delta > vmPool.MaxSize() {
		return fmt.Errorf("size-increasing request of %d is bigger than max size %d", int(currentSize)+delta, vmPool.MaxSize())
	}

	return vmPool.scaleUpToCount(currentSize + int64(delta))
}

// buildRequestBodyForScaleUp builds the request body for scale up for self-hosted CAS
func (vmPool *VMPool) buildRequestBodyForScaleUp(versionedAP armcontainerservice.AgentPool, count int64) (armcontainerservice.AgentPool, error) {
	if versionedAP.Properties != nil &&
		versionedAP.Properties.VirtualMachinesProfile != nil &&
		versionedAP.Properties.VirtualMachinesProfile.Scale != nil {
		requestBody := armcontainerservice.AgentPool{
			Properties: &armcontainerservice.ManagedClusterAgentPoolProfileProperties{
				Type: versionedAP.Properties.Type,
			},
		}

		// the request body must have the same mode as the original agentpool
		// otherwise the PUT request will fail
		if versionedAP.Properties.Mode != nil &&
			*versionedAP.Properties.Mode == armcontainerservice.AgentPoolModeSystem {
			systemMode := armcontainerservice.AgentPoolModeSystem
			requestBody.Properties.Mode = &systemMode
		}

		// set the count of the matching manual scale profile to the new target value
		for _, manualProfile := range versionedAP.Properties.VirtualMachinesProfile.Scale.Manual {
			if manualProfile != nil && len(manualProfile.Sizes) == 1 &&
				strings.EqualFold(to.String(manualProfile.Sizes[0]), vmPool.sku) {
				manualProfile.Count = to.Int32Ptr(int32(count))
				requestBody.Properties.VirtualMachinesProfile = versionedAP.Properties.VirtualMachinesProfile
				return requestBody, nil
			}
		}
	}
	return armcontainerservice.AgentPool{}, fmt.Errorf("failed to build request body for scale up, agentpool doesn't have valid virtualMachinesProfile")
}

// scaleUpToCount sets node count for vmPool to target value through PUT AP call.
func (vmPool *VMPool) scaleUpToCount(count int64) error {
	vmPool.sizeMutex.Lock()
	defer vmPool.sizeMutex.Unlock()

	updateCtx, updateCancel := getContextWithCancel()
	defer updateCancel()

	versionedAP, err := vmPool.getAgentpoolFromCache()
	if err != nil {
		klog.Errorf("Failed to get vmPool %s, error: %s", vmPool.agentPoolName, err)
		return err
	}

	requestBody := armcontainerservice.AgentPool{}
	// self-hosted CAS will be using Manual scale profile
	if len(versionedAP.Properties.VirtualMachinesProfile.Scale.Manual) > 0 {
		requestBody, err = vmPool.buildRequestBodyForScaleUp(versionedAP, count)
		if err != nil {
			klog.Errorf("Failed to build request body for scale up VMPool %s, error: %s", vmPool.Name, err)
			return err
		}
	} else { // AKS-managed CAS will use custom header for setting the target count
		header := make(http.Header)
		header.Set("Target-Count", fmt.Sprintf("%d", count))
		updateCtx = policy.WithHTTPHeader(updateCtx, header)
	}

	defer vmPool.manager.invalidateCache()
	poller, err := vmPool.manager.azClient.agentPoolClient.BeginCreateOrUpdate(
		updateCtx,
		vmPool.clusterResourceGroup,
		vmPool.clusterName,
		vmPool.agentPoolName,
		requestBody, nil)

	if err != nil {
		klog.Errorf("Failed to scale up agentpool %s in cluster %s for vmPool %s with error: %v",
			vmPool.agentPoolName, vmPool.clusterName, vmPool.Name, err)
		return err
	}

	vmPool.curSize = count
	vmPool.lastSizeRefresh = time.Now()

	go func() {
		updateCtx, cancel := getContextWithTimeout(vmsAsyncContextTimeout)
		defer cancel()
		if _, err := poller.PollUntilDone(updateCtx, nil); err != nil {
			klog.Errorf("agentPoolClient.BeginCreateOrUpdate for aks cluster %s agentpool %s for scaling up vmPool %s failed with error %s",
				vmPool.clusterName, vmPool.agentPoolName, vmPool.Name, err)
		}
	}()

	return err
}

// DeleteNodes extracts the providerIDs from the node spec and
// delete or deallocate the nodes based on the scale down policy of agentpool.
func (vmPool *VMPool) DeleteNodes(nodes []*apiv1.Node) error {
	currentSize, err := vmPool.getVMPoolSize()
	if err != nil {
		return err
	}

	// if the target size is smaller than the min size, return an error
	if int(currentSize) <= vmPool.MinSize() {
		return fmt.Errorf("min size %d reached, nodes will not be deleted", vmPool.MinSize())
	}

	var providerIDs []string
	for _, node := range nodes {
		belongs, err := vmPool.Belongs(node)
		if err != nil {
			return err
		}
		if !belongs {
			return fmt.Errorf("node %s does not belong to vmPool %s", node.Name, vmPool.Name)
		}

		providerIDs = append(providerIDs, node.Spec.ProviderID)
	}

	return vmPool.scaleDownNodes(providerIDs)
}

// scaleDownNodes delete nodes from the vmPool
func (vmPool *VMPool) scaleDownNodes(providerIDs []string) error {
	vmPool.sizeMutex.Lock()
	defer vmPool.sizeMutex.Unlock()

	if len(providerIDs) == 0 {
		return nil
	}

	klog.V(3).Infof("Deleting nodes from vmPool %s: %v", vmPool.Name, providerIDs)

	deleteCtx, deleteCancel := getContextWithTimeout(vmsContextTimeout)
	defer deleteCancel()

	machineNames := make([]*string, len(providerIDs))
	for i, providerID := range providerIDs {
		// extract the machine name from the providerID by splitting the providerID by '/' and get the last element
		// The providerID look like this:
		// "azure:///subscriptions/0000000-0000-0000-0000-00000000000/resourceGroups/mc_myrg_mycluster_eastus/providers/Microsoft.Compute/virtualMachines/aks-mypool-12345678-vms0"
		machineName, err := resourceName(providerID)
		if err != nil {
			return err
		}
		machineNames[i] = &machineName
	}

	requestBody := armcontainerservice.AgentPoolDeleteMachinesParameter{
		MachineNames: machineNames,
	}

	defer vmPool.manager.invalidateCache()

	poller, err := vmPool.manager.azClient.agentPoolClient.BeginDeleteMachines(
		deleteCtx,
		vmPool.clusterResourceGroup,
		vmPool.clusterName,
		vmPool.agentPoolName,
		requestBody, nil)
	if err != nil {
		klog.Errorf("Failed to delete nodes from agentpool %s in cluster %s with error: %v",
			vmPool.agentPoolName, vmPool.clusterName, err)
		return err
	}

	vmPool.curSize -= int64(len(providerIDs))
	vmPool.lastSizeRefresh = time.Now()
	klog.Infof("Decreased vmPool %s size to %d", vmPool.Name, vmPool.curSize)

	go func() {
		updateCtx, cancel := getContextWithTimeout(vmsAsyncContextTimeout)
		defer cancel()
		if _, err := poller.PollUntilDone(updateCtx, nil); err != nil {
			klog.Errorf("agentPoolClient.BeginDeleteMachines for aks cluster %s for scaling down vmPool %s failed with error %s",
				vmPool.clusterName, vmPool.agentPoolName, err)
		}
		klog.Infof("Successfully deleted nodes from vmPool %s", vmPool.Name)
	}()

	return nil
}

// Belongs returns true if the given k8s node belongs to this vms nodepool.
func (vmPool *VMPool) Belongs(node *apiv1.Node) (bool, error) {
	klog.V(6).Infof("Check if node belongs to this vmPool:%s, node:%v\n", vmPool, node)

	ref := &azureRef{
		Name: node.Spec.ProviderID,
	}

	nodeGroup, err := vmPool.manager.GetNodeGroupForInstance(ref)
	if err != nil {
		return false, err
	}
	if nodeGroup == nil {
		return false, fmt.Errorf("%s doesn't belong to a known node group", node.Name)
	}
	if !strings.EqualFold(nodeGroup.Id(), vmPool.Id()) {
		return false, nil
	}
	return true, nil
}

// DecreaseTargetSize decreases the target size of the node group.
func (vmPool *VMPool) DecreaseTargetSize(delta int) error {
	vmPool.manager.invalidateCache()
	_, err := vmPool.getVMPoolSize()
	if err != nil {
		klog.Warningf("DecreaseTargetSize: failed with error: %v", err)
	}
	return err
}

// Id returns the name of the agentPool, it is in the format of <agentpoolname>/<sku>
// e.g. mypool1/Standard_D2s_v3
func (vmPool *VMPool) Id() string {
	return vmPool.azureRef.Name
}

// Debug returns a string with basic details of the agentPool
func (vmPool *VMPool) Debug() string {
	return fmt.Sprintf("%s (%d:%d)", vmPool.Id(), vmPool.MinSize(), vmPool.MaxSize())
}

func isSpotAgentPool(ap armcontainerservice.AgentPool) bool {
	if ap.Properties != nil && ap.Properties.ScaleSetPriority != nil {
		return strings.EqualFold(string(*ap.Properties.ScaleSetPriority), "Spot")
	}
	return false
}

// getNodeCountForSkuFromAgentPool returns the count of nodes with the given sku in the agent pool
func getNodeCountForSkuFromAgentPool(ap armcontainerservice.AgentPool, sku string) int32 {
	if ap.Properties != nil {
		// the VirtualMachineNodesStatus returned by AKS-RP is constructed from the vm list returned from CRP.
		// it only contains VMs in the running state.
		for _, status := range ap.Properties.VirtualMachineNodesStatus {
			if status != nil {
				if strings.EqualFold(to.String(status.Size), sku) {
					return to.Int32(status.Count)
				}
			}
		}
	}
	return 0
}

const (
	spotPoolCacheTTL = 15 * time.Second
)

// getCurSize returns the current vm count in the vmPool
// It uses azure cache as the source of truth for non-spot pools, and resorts to AKS-RP for spot pools
func (vmPool *VMPool) getCurSize() (int64, error) {
	vmPool.sizeMutex.Lock()
	defer vmPool.sizeMutex.Unlock()

	ap, err := vmPool.getAgentpoolFromCache()
	if err != nil {
		klog.Errorf("Failed to get agentpool %s from cache with error: %v", vmPool.agentPoolName, err)
		return -1, err
	}

	// spot pool has a shorter cache TTL
	cacheTTL := vmPool.sizeRefreshPeriod
	if isSpotAgentPool(ap) {
		cacheTTL = spotPoolCacheTTL
	}

	if vmPool.lastSizeRefresh.Add(cacheTTL).After(time.Now()) {
		klog.V(3).Infof("VMPool: %s, returning in-memory size: %d", vmPool.Name, vmPool.curSize)
		return vmPool.curSize, nil
	}

	if isSpotAgentPool(ap) {
		ap, err = vmPool.getAgentpoolFromAzure()
		if err != nil {
			klog.Errorf("Failed to get agentpool %s from azure with error: %v", vmPool.agentPoolName, err)
			return -1, err
		}
	}

	realSize := int64(getNodeCountForSkuFromAgentPool(ap, vmPool.sku))

	if vmPool.curSize != realSize {
		// invalidate the instance cache if the size has changed.
		klog.V(5).Infof("VMpool %s getCurSize: curSize(%d) != real size (%d), invalidating azure cache", vmPool.Name, vmPool.curSize, realSize)
		vmPool.manager.invalidateCache()
	}

	vmPool.curSize = realSize
	vmPool.lastSizeRefresh = time.Now()
	return vmPool.curSize, nil
}

// getVMsFromCache returns the list of virtual machines in this vmPool
func (vmPool *VMPool) getVMsFromCache() ([]compute.VirtualMachine, error) {
	curSize, err := vmPool.getCurSize()
	if err != nil {
		klog.Errorf("Failed to get current size for VMPool %q: %v", vmPool.Name, err)
		return nil, err
	}

	// vmsMap is a map of agent pool name to the list of virtual machines belongs to the agent pool
	// this method may return an empty list if the agentpool has no nodes inside, i.e. minCount is set to 0
	vmsMap := vmPool.manager.azureCache.getVirtualMachines()
	var vmsWithsku []compute.VirtualMachine
	for _, vm := range vmsMap[vmPool.agentPoolName] {
		if vm.HardwareProfile == nil {
			continue
		}
		if strings.EqualFold(string(vm.HardwareProfile.VMSize), vmPool.sku) {
			vmsWithsku = append(vmsWithsku, vm)
		}
	}

	if int64(len(vmsWithsku)) != curSize {
		klog.V(5).Infof("VMPool %s has vm list size (%d) != curSize (%d), invalidating azure cache",
			vmPool.Name, len(vmsWithsku), curSize)
		vmPool.manager.invalidateCache()
	}

	return vmsWithsku, nil
}

// Nodes returns the list of nodes in the vms agentPool.
func (vmPool *VMPool) Nodes() ([]cloudprovider.Instance, error) {
	vms, err := vmPool.getVMsFromCache()
	if err != nil {
		return nil, err
	}

	nodes := make([]cloudprovider.Instance, 0, len(vms))
	for _, vm := range vms {
		if vm.ID == nil || len(*vm.ID) == 0 {
			continue
		}
		resourceID, err := convertResourceGroupNameToLower("azure://" + to.String(vm.ID))
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, cloudprovider.Instance{Id: resourceID})
	}

	return nodes, nil
}

// TemplateNodeInfo returns a NodeInfo object that can be used to create a new node in the vmPool.
func (vmPool *VMPool) TemplateNodeInfo() (*schedulerframework.NodeInfo, error) {
	ap, err := vmPool.getAgentpoolFromCache()
	if err != nil {
		return nil, err
	}

	template := buildNodeTemplateFromVMPool(ap, vmPool.location, vmPool.sku)
	node, err := buildNodeFromTemplate(vmPool.agentPoolName, template, vmPool.manager, vmPool.enableDynamicInstanceList)
	if err != nil {
		return nil, err
	}

	nodeInfo := schedulerframework.NewNodeInfo(cloudprovider.BuildKubeProxy(vmPool.agentPoolName))
	nodeInfo.SetNode(node)
	return nodeInfo, nil
}

func (vmPool *VMPool) getAgentpoolFromCache() (armcontainerservice.AgentPool, error) {
	vmsPoolMap := vmPool.manager.azureCache.getVMsPoolMap()
	if _, exists := vmsPoolMap[vmPool.agentPoolName]; !exists {
		return armcontainerservice.AgentPool{}, fmt.Errorf("VMs agent pool %s not found in cache", vmPool.agentPoolName)
	}
	return vmsPoolMap[vmPool.agentPoolName], nil
}

// AtomicIncreaseSize is not implemented.
func (vmPool *VMPool) AtomicIncreaseSize(delta int) error {
	return cloudprovider.ErrNotImplemented
}
