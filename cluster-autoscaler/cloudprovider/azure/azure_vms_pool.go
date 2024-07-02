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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v4"
	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2022-08-01/compute"
	"github.com/Azure/go-autorest/autorest/to"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/config/dynamic"
	klog "k8s.io/klog/v2"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
)

// VMsPool is single instance VM pool
type VMsPool struct {
	azureRef
	manager              *AzureManager
	resourceGroup        string // MC_ resource group for nodes
	clusterResourceGroup string // resource group for the cluster itself
	clusterName          string

	minSize int
	maxSize int

	curSize         int64
	sizeMutex       sync.Mutex
	lastSizeRefresh time.Time
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

		manager:              am,
		resourceGroup:        am.config.ResourceGroup,
		clusterResourceGroup: am.config.ClusterResourceGroup,
		clusterName:          am.config.ClusterName,

		curSize: -1,
		minSize: spec.MinSize,
		maxSize: spec.MaxSize,
	}

	return nodepool, nil
}

const (
	getRequestContextTimeout    = 1 * time.Minute
	updateRequestContextTimeout = 5 * time.Minute
)

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

func (agentPool *VMsPool) getAgentpool() (armcontainerservice.AgentPool, error) {
	ctx, cancel := getContextWithTimeout(getRequestContextTimeout)
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
	// TODO(wenxuan): exlude the nodes in deallocated/deallocating states if agentpool is using deallocate scale down policy
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
	versionedAP, err := agentPool.getAgentpool()
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
			versionedAP.Properties.VirtualMachinesProfile.Scale.Manual[0].Count = to.Int32Ptr(int32(count))
			requestBody.Properties.VirtualMachinesProfile = &virtualMachineProfile
			return requestBody, nil
		}
		// aks-managed CAS will be using AutoscaleProfile
		// TODO(wenxuan): support AutoscaleProfile once it become available with the new API release
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

	klog.V(1).Infof("Scaling up vms pool %s to new target count: %d", agentPool.Name, count)
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

	updateCtx, cancel := getContextWithTimeout(updateRequestContextTimeout)
	defer cancel()

	if _, err = poller.PollUntilDone(updateCtx, nil); err == nil {
		// success path
		klog.Infof("agentPoolClient.BeginCreateOrUpdate for aks cluster %s agentpool %s succeeded", agentPool.clusterName, agentPool.Name)
		agentPool.curSize = count
		agentPool.lastSizeRefresh = time.Now()
		klog.V(6).Infof("setVMsPoolNodeCount: invalidating cache, new size %d", count)
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
		klog.V(3).Infof("min size %d reached, nodes will not be deleted", agentPool.MinSize())
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

// scaleDownNodes delete or deallocate the nodes from the agent pool based on the scale down policy.
func (agentPool *VMsPool) scaleDownNodes(providerIDs []string) error {
	agentPool.sizeMutex.Lock()
	defer agentPool.sizeMutex.Unlock()

	if len(providerIDs) == 0 {
		return nil
	}

	klog.V(3).Infof("Deleting nodes from agent pool %s: %v", agentPool.Name, providerIDs)

	deleteCtx, deleteCancel := getContextWithCancel()
	defer deleteCancel()

	machineNames := make([]*string, len(providerIDs))
	for i, providerID := range providerIDs {
		// extract the machine name from the providerID by splitting the providerID by '/' and get the last element
		// The providerID look like this:
		// "azure:///subscriptions/feb5b150-60fe-4441-be73-8c02a524f55a/resourceGroups/mc_wxrg_play-vms_eastus/providers/Microsoft.Compute/virtualMachines/aks-nodes-32301838-vms0"
		providerIDParts := strings.Split(providerID, "/")
		machineNames[i] = &providerIDParts[len(providerIDParts)-1]
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

	updateCtx, cancel := getContextWithTimeout(updateRequestContextTimeout)
	defer cancel()
	if _, err = poller.PollUntilDone(updateCtx, nil); err == nil {
		klog.Infof("agentPoolClient.BeginDeleteMachines for aks cluster %s agentpool %s succeeded, machine names: %s",
			agentPool.clusterName, agentPool.Name, providerIDs)
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
	return nil
}

// Id returns the name of the agentPool
func (agentPool *VMsPool) Id() string {
	return agentPool.azureRef.Name
}

// Debug returns a string with basic details of the agentPool
func (agentPool *VMsPool) Debug() string {
	return fmt.Sprintf("%s (%d:%d)", agentPool.Id(), agentPool.MinSize(), agentPool.MaxSize())
}

// getCurSize returns the current vm count in the vms agent pool
// It use azure cache as the source of truth.
func (agentPool *VMsPool) getCurSize() (int64, error) {
	agentPool.sizeMutex.Lock()
	defer agentPool.sizeMutex.Unlock()

	if agentPool.lastSizeRefresh.Add(15 * time.Second).After(time.Now()) {
		klog.V(3).Infof("VMs Agentpool: %s, returning in-memory size: %d", agentPool.Name, agentPool.curSize)
		return agentPool.curSize, nil
	}

	vms, err := agentPool.getVMsFromCache()
	if err != nil {
		klog.Errorf("Failed to get vms pool %s from cache with error: %v", agentPool.Name, err)
		return 0, err
	}

	realSize := int64(len(vms))
	if agentPool.curSize != realSize {
		// invalidate the instance cache if the size has changed.
		klog.V(5).Infof("VMs Agentpool getCurSize: %s curSize(%d) != real size (%d), invalidating azure cache", agentPool.Name, agentPool.curSize, realSize)
		agentPool.manager.invalidateCache()
	}

	agentPool.curSize = realSize
	agentPool.lastSizeRefresh = time.Now()
	return agentPool.curSize, nil
}

func (agentPool *VMsPool) getVMsFromCache() ([]compute.VirtualMachine, error) {
	// vmsPoolMap is a map of agent pool name to the list of virtual machines
	vmsPoolMap := agentPool.manager.azureCache.getVirtualMachines()
	if _, ok := vmsPoolMap[agentPool.Name]; !ok {
		return []compute.VirtualMachine{}, fmt.Errorf("vms pool %s not found in the cache", agentPool.Name)
	}

	return vmsPoolMap[agentPool.Name], nil
}

// Nodes returns the list of nodes in the vms agentPool.
func (agentPool *VMsPool) Nodes() ([]cloudprovider.Instance, error) {
	vms, err := agentPool.getVMsFromCache()
	if err != nil {
		return nil, err
	}

	nodes := make([]cloudprovider.Instance, 0, len(vms))
	for _, vm := range vms {
		if len(*vm.ID) == 0 {
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
	// TODO(wenxuan): implement this method when vms pool can fully support GPU nodepool
	return nil, cloudprovider.ErrNotImplemented
}

// AtomicIncreaseSize is not implemented.
func (agentPool *VMsPool) AtomicIncreaseSize(delta int) error {
	return cloudprovider.ErrNotImplemented
}
