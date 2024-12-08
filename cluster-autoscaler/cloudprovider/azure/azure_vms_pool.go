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
	"strings"
	"sync"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v5"
	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2022-03-01/compute"
	"github.com/Azure/go-autorest/autorest/to"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/config/dynamic"
	klog "k8s.io/klog/v2"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
)

// VMsPool consists of standalone VMs, it's not backed by VMSS
type VMsPool struct {
	azureRef
	manager              *AzureManager
	resourceGroup        string // MC_ resource group for nodes
	clusterResourceGroup string // resource group for the cluster itself
	clusterName          string
	location             string

	minSize int
	maxSize int

	curSize           int64
	sizeMutex         sync.Mutex
	lastSizeRefresh   time.Time
	sizeRefreshPeriod time.Duration

	enableDynamicInstanceList bool
}

// NewVMsPool creates a new VMsPool
func NewVMsPool(spec *dynamic.NodeGroupSpec, am *AzureManager) (*VMsPool, error) {
	if am.azClient.agentPoolClient == nil {
		return nil, fmt.Errorf("agentPoolClient is nil")
	}
	nodepool := &VMsPool{
		azureRef: azureRef{
			Name: spec.Name,
		},

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

// MinSize returns the minimum size the cluster is allowed to scaled down
// to as provided by the node spec in --node parameter.
func (agentPool *VMsPool) MinSize() int {
	return agentPool.minSize
}

// Exist is always true since we are initialized with an existing agentpool
func (agentPool *VMsPool) Exist() bool {
	return true
}

// Create creates the node group on the cloud provider side.
func (agentPool *VMsPool) Create() (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrAlreadyExist
}

// Delete deletes the node group on the cloud provider side.
func (agentPool *VMsPool) Delete() error {
	return cloudprovider.ErrNotImplemented
}

// Autoprovisioned is always false since we are initialized with an existing agentpool
func (agentPool *VMsPool) Autoprovisioned() bool {
	return false
}

// GetOptions returns NodeGroupAutoscalingOptions that should be used for this particular
// NodeGroup. Returning a nil will result in using default options.
func (agentPool *VMsPool) GetOptions(defaults config.NodeGroupAutoscalingOptions) (*config.NodeGroupAutoscalingOptions, error) {
	// TODO(wenxuan): implement this method when vms pool can fully support GPU nodepool
	return nil, nil
}

// MaxSize returns the maximum size scale limit provided by --node
// parameter to the autoscaler main
func (agentPool *VMsPool) MaxSize() int {
	return agentPool.maxSize
}

// TargetSize returns the current TARGET size of the node group. It is possible that the
// number is different from the number of nodes registered in Kubernetes.
func (agentPool *VMsPool) TargetSize() (int, error) {
	size, err := agentPool.getVMsPoolSize()
	return int(size), err
}

func (agentPool *VMsPool) getAgentpoolFromAzure() (armcontainerservice.AgentPool, error) {
	ctx, cancel := getContextWithTimeout(vmsGetRequestContextTimeout)
	defer cancel()
	resp, err := agentPool.manager.azClient.agentPoolClient.Get(
		ctx,
		agentPool.clusterResourceGroup,
		agentPool.clusterName,
		agentPool.Name, nil)
	if err != nil {
		return resp.AgentPool, fmt.Errorf("failed to get agentpool %s in cluster %s with error: %v",
			agentPool.Name, agentPool.clusterName, err)
	}
	return resp.AgentPool, nil
}

// getVMsPoolSize returns the current size of the vms pool, which is the vm count in the vms pool
func (agentPool *VMsPool) getVMsPoolSize() (int64, error) {
	size, err := agentPool.getCurSize()
	if err != nil {
		klog.Errorf("Failed to get vms pool %s node count with error: %s", agentPool.Name, err)
		return size, err
	}
	if size == -1 {
		klog.Errorf("vms pool %s size is -1, it is still being initialized", agentPool.Name)
		return size, fmt.Errorf("getVMsPoolSize: size is -1 for vms pool %s", agentPool.Name)
	}
	return size, nil
}

// IncreaseSize increase the size through a PUT AP call.
func (agentPool *VMsPool) IncreaseSize(delta int) error {
	if delta <= 0 {
		return fmt.Errorf("size increase must be positive, current delta: %d", delta)
	}

	currentSize, err := agentPool.getVMsPoolSize()
	if err != nil {
		return err
	}

	if currentSize == -1 {
		return fmt.Errorf("the vms pool %s is still under initialization, skip the size increase", agentPool.Name)
	}

	if int(currentSize)+delta > agentPool.MaxSize() {
		return fmt.Errorf("size-increasing request of %d is bigger than max size %d", int(currentSize)+delta, agentPool.MaxSize())
	}

	return agentPool.scaleUpToCount(currentSize + int64(delta))
}

func (agentPool *VMsPool) buildRequestBodyForScaleUp(count int64) (armcontainerservice.AgentPool, error) {
	versionedAP, err := agentPool.getAgentpoolFromCache()
	if err != nil {
		klog.Errorf("Failed to get vms pool %s, error: %s", agentPool.Name, err)
		return armcontainerservice.AgentPool{}, err
	}

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

		// self-hosted CAS will be using Manual scale profile
		if len(versionedAP.Properties.VirtualMachinesProfile.Scale.Manual) > 0 {
			// set the count for first manual scale profile to the new target value
			virtualMachineProfile := *versionedAP.Properties.VirtualMachinesProfile
			virtualMachineProfile.Scale.Manual[0].Count = to.Int32Ptr(int32(count))
			requestBody.Properties.VirtualMachinesProfile = &virtualMachineProfile
			return requestBody, nil
		}

		// aks-managed CAS will be using Auto scale Profile
		if len(versionedAP.Properties.VirtualMachinesProfile.Scale.Autoscale) > 0 {
			// set the MinCount and MaxCount for first AutoscaleProfile to the new target value
			virtualMachineProfile := *versionedAP.Properties.VirtualMachinesProfile
			virtualMachineProfile.Scale.Autoscale[0].MinCount = to.Int32Ptr(int32(count))
			virtualMachineProfile.Scale.Autoscale[0].MaxCount = to.Int32Ptr(int32(count))
			requestBody.Properties.VirtualMachinesProfile = &virtualMachineProfile
			return requestBody, nil
		}
	}
	return armcontainerservice.AgentPool{}, fmt.Errorf("failed to build request body for scale up, agentpool doesn't have valid virtualMachinesProfile")
}

// scaleUpToCount sets node count for vms agent pool to target value through PUT AP call.
func (agentPool *VMsPool) scaleUpToCount(count int64) error {
	agentPool.sizeMutex.Lock()
	defer agentPool.sizeMutex.Unlock()

	updateCtx, updateCancel := getContextWithCancel()
	defer updateCancel()

	requestBody, err := agentPool.buildRequestBodyForScaleUp(count)
	if err != nil {
		klog.Errorf("Failed to build request body for scale up, error: %s", err)
		return err
	}

	poller, err := agentPool.manager.azClient.agentPoolClient.BeginCreateOrUpdate(
		updateCtx,
		agentPool.clusterResourceGroup,
		agentPool.clusterName,
		agentPool.Name,
		requestBody, nil)

	if err != nil {
		klog.Errorf("Failed to update agentpool %s in cluster %s with error: %v",
			agentPool.azureRef.Name, agentPool.clusterName, err)
		return err
	}

	updateCtx, cancel := getContextWithTimeout(vmsPutRequestContextTimeout)
	defer cancel()

	if _, err = poller.PollUntilDone(updateCtx, nil); err == nil {
		// success path
		agentPool.curSize = count
		agentPool.lastSizeRefresh = time.Now()
		agentPool.manager.invalidateCache()
		return nil
	}

	klog.Errorf("agentPoolClient.BeginCreateOrUpdate for aks cluster %s agentpool %s failed with error %s",
		agentPool.clusterName, agentPool.Name, err)
	return nil
}

// DeleteNodes extracts the providerIDs from the node spec and
// delete or deallocate the nodes based on the scale down policy of agentpool.
func (agentPool *VMsPool) DeleteNodes(nodes []*apiv1.Node) error {
	currentSize, err := agentPool.getVMsPoolSize()
	if err != nil {
		return err
	}

	// if the target size is smaller than the min size, return an error
	if int(currentSize) <= agentPool.MinSize() {
		return fmt.Errorf("min size %d reached, nodes will not be deleted", agentPool.MinSize())
	}

	var providerIDs []string
	for _, node := range nodes {
		belongs, err := agentPool.Belongs(node)
		if err != nil {
			return err
		}
		if !belongs {
			return fmt.Errorf("node %s does not belong to agent pool %s", node.Name, agentPool.Name)
		}

		providerIDs = append(providerIDs, node.Spec.ProviderID)
	}

	return agentPool.scaleDownNodes(providerIDs)
}

// scaleDownNodes delete nodes from the agent pool
func (agentPool *VMsPool) scaleDownNodes(providerIDs []string) error {
	agentPool.sizeMutex.Lock()
	defer agentPool.sizeMutex.Unlock()

	if len(providerIDs) == 0 {
		return nil
	}

	klog.V(3).Infof("Deleting nodes from agent pool %s: %v", agentPool.Name, providerIDs)

	deleteCtx, deleteCancel := getContextWithTimeout(vmsDeleteRequestContextTimeout)
	defer deleteCancel()

	machineNames := make([]*string, len(providerIDs))
	for i, providerID := range providerIDs {
		// extract the machine name from the providerID by splitting the providerID by '/' and get the last element
		// The providerID look like this:
		// "azure:///subscriptions/0000000-0000-0000-0000-00000000000/resourceGroups/mc_wxrg_play-vms_eastus/providers/Microsoft.Compute/virtualMachines/aks-nodes-32301838-vms0"
		machineName, err := resourceName(providerID)
		if err != nil {
			return err
		}
		machineNames[i] = &machineName
	}

	requestBody := armcontainerservice.AgentPoolDeleteMachinesParameter{
		MachineNames: machineNames,
	}

	poller, err := agentPool.manager.azClient.agentPoolClient.BeginDeleteMachines(
		deleteCtx,
		agentPool.clusterResourceGroup,
		agentPool.clusterName,
		agentPool.Name,
		requestBody, nil)
	if err != nil {
		klog.Errorf("Failed to update agentpool %s in cluster %s with error: %v",
			agentPool.azureRef.Name, agentPool.clusterName, err)
		return err
	}

	defer agentPool.manager.invalidateCache()

	updateCtx, cancel := getContextWithTimeout(vmsPutRequestContextTimeout)
	defer cancel()
	if _, err = poller.PollUntilDone(updateCtx, nil); err == nil {
		return nil
	}

	klog.Errorf("agentPoolClient.BeginDeleteMachines for aks cluster %s agentpool %s failed with error %s",
		agentPool.clusterName, agentPool.Name, err)

	return nil
}

// Belongs returns true if the given k8s node belongs to this vms nodepool.
func (agentPool *VMsPool) Belongs(node *apiv1.Node) (bool, error) {
	klog.V(6).Infof("Check if node belongs to this vms pool:%s, node:%v\n", agentPool, node)

	ref := &azureRef{
		Name: node.Spec.ProviderID,
	}

	nodeGroup, err := agentPool.manager.GetNodeGroupForInstance(ref)
	if err != nil {
		return false, err
	}
	if nodeGroup == nil {
		return false, fmt.Errorf("%s doesn't belong to a known node group", node.Name)
	}
	if !strings.EqualFold(nodeGroup.Id(), agentPool.Id()) {
		return false, nil
	}
	return true, nil
}

// DecreaseTargetSize decreases the target size of the node group.
func (agentPool *VMsPool) DecreaseTargetSize(delta int) error {
	agentPool.manager.invalidateCache()
	_, err := agentPool.getVMsPoolSize()
	if err != nil {
		klog.Warningf("DecreaseTargetSize: failed with error: %v", err)
	}
	return err
}

// Id returns the name of the agentPool
func (agentPool *VMsPool) Id() string {
	return agentPool.azureRef.Name
}

// Debug returns a string with basic details of the agentPool
func (agentPool *VMsPool) Debug() string {
	return fmt.Sprintf("%s (%d:%d)", agentPool.Id(), agentPool.MinSize(), agentPool.MaxSize())
}

func isSpotVMsPool(ap armcontainerservice.AgentPool) bool {
	if ap.Properties != nil && ap.Properties.ScaleSetPriority != nil {
		return strings.EqualFold(string(*ap.Properties.ScaleSetPriority), "Spot")
	}
	return false
}

func getNodeCountFromAgentPool(ap armcontainerservice.AgentPool) int32 {
	size := int32(0)
	if ap.Properties != nil {
		// the VirtualMachineNodesStatus returned by AKS-RP is constructed from the vm list returned from CRP.
		// it only contains VMs in the running state.
		for _, status := range ap.Properties.VirtualMachineNodesStatus {
			if status.Count != nil {
				size += to.Int32(status.Count)
			}
		}
	}
	return size
}

const (
	spotPoolCacheTTL = 15 * time.Second
)

// getCurSize returns the current vm count in the vms agent pool
// It uses azure cache as the source of truth for non-spot pools, and resorts to AKS-RP for spot pools
func (agentPool *VMsPool) getCurSize() (int64, error) {
	agentPool.sizeMutex.Lock()
	defer agentPool.sizeMutex.Unlock()

	vmsPool, err := agentPool.getAgentpoolFromCache()
	if err != nil {
		klog.Errorf("Failed to get vms pool %s from cache with error: %v", agentPool.Name, err)
		return -1, err
	}

	// spot pool has a shorter cache TTL
	cacheTTL := agentPool.sizeRefreshPeriod
	if isSpotVMsPool(vmsPool) {
		cacheTTL = spotPoolCacheTTL
	}

	if agentPool.lastSizeRefresh.Add(cacheTTL).After(time.Now()) {
		klog.V(3).Infof("VMs Agentpool: %s, returning in-memory size: %d", agentPool.Name, agentPool.curSize)
		return agentPool.curSize, nil
	}

	if isSpotVMsPool(vmsPool) {
		vmsPool, err = agentPool.getAgentpoolFromAzure()
		if err != nil {
			klog.Errorf("Failed to get vms pool %s from azure with error: %v", agentPool.Name, err)
			return -1, err
		}
	}

	realSize := int64(getNodeCountFromAgentPool(vmsPool))

	if agentPool.curSize != realSize {
		// invalidate the instance cache if the size has changed.
		klog.V(5).Infof("VMs Agentpool %s getCurSize: curSize(%d) != real size (%d), invalidating azure cache", agentPool.Name, agentPool.curSize, realSize)
		agentPool.manager.invalidateCache()
	}

	agentPool.curSize = realSize
	agentPool.lastSizeRefresh = time.Now()
	return agentPool.curSize, nil
}

func (agentPool *VMsPool) getVMsFromCache() ([]compute.VirtualMachine, error) {
	curSize, err := agentPool.getCurSize()
	if err != nil {
		klog.Errorf("Failed to get current size for VMs pool %q: %v", agentPool.Name, err)
		return nil, err
	}

	// vmsMap is a map of agent pool name to the list of virtual machines belongs to the agent pool
	// this method may return an empty list if the agentpool has no nodes inside, i.e. minCount is set to 0
	vmsMap := agentPool.manager.azureCache.getVirtualMachines()
	if int64(len(vmsMap[agentPool.Name])) != curSize {
		klog.V(5).Infof("VMs Agentpool %s vm list size (%d) != curSize (%d), invalidating azure cache",
			agentPool.Name, len(vmsMap[agentPool.Name]), curSize)
		agentPool.manager.invalidateCache()
	}

	return vmsMap[agentPool.Name], nil
}

// Nodes returns the list of nodes in the vms agentPool.
func (agentPool *VMsPool) Nodes() ([]cloudprovider.Instance, error) {
	vms, err := agentPool.getVMsFromCache()
	if err != nil {
		return nil, err
	}

	nodes := make([]cloudprovider.Instance, 0, len(vms))
	for _, vm := range vms {
		if len(to.String(vm.ID)) == 0 {
			continue
		}
		resourceID, err := convertResourceGroupNameToLower("azure://" + *vm.ID)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, cloudprovider.Instance{Id: resourceID})
	}

	return nodes, nil
}

// TemplateNodeInfo is not implemented.
func (agentPool *VMsPool) TemplateNodeInfo() (*schedulerframework.NodeInfo, error) {
	vmsPool, err := agentPool.getAgentpoolFromCache()
	if err != nil {
		return nil, err
	}

	template := buildNodeTemplateFromVMsPool(vmsPool, agentPool.location)
	node, err := buildNodeFromTemplate(agentPool.Name, template, agentPool.manager, agentPool.enableDynamicInstanceList)

	if err != nil {
		return nil, err
	}

	nodeInfo := schedulerframework.NewNodeInfo(cloudprovider.BuildKubeProxy(agentPool.Name))
	nodeInfo.SetNode(node)
	return nodeInfo, nil
}

func (agentPool *VMsPool) getAgentpoolFromCache() (armcontainerservice.AgentPool, error) {
	vmsPoolMap := agentPool.manager.azureCache.getVMsPoolMap()
	if _, exists := vmsPoolMap[agentPool.Name]; !exists {
		return armcontainerservice.AgentPool{}, fmt.Errorf("VMs agent pool %s not found in cache", agentPool.Name)
	}
	return vmsPoolMap[agentPool.Name], nil
}

// AtomicIncreaseSize is not implemented.
func (agentPool *VMsPool) AtomicIncreaseSize(delta int) error {
	return cloudprovider.ErrNotImplemented
}
