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

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v6"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v8"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/config/dynamic"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	klog "k8s.io/klog/v2"
	"k8s.io/utils/ptr"
)

// VMPool represents a group of standalone virtual machines (VMs) with a single SKU.
// It is part of a mixed-SKU agent pool (an agent pool with type `VirtualMachines`).
// Terminology:
// - Agent pool: A node pool in an AKS cluster.
// - VMs pool: An agent pool of type `VirtualMachines`, which can contain mixed SKUs.
// - VMPool: A subset of VMs within a VMs pool that share the same SKU.
type VMPool struct {
	azureRef
	manager       *AzureManager
	agentPoolName string // the virtual machines agentpool that this VMPool belongs to
	sku           string // sku of the VM in the pool

	minSize int
	maxSize int
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
		manager:       am,
		sku:           sku,
		agentPoolName: agentPoolName,
		minSize:       spec.MinSize,
		maxSize:       spec.MaxSize,
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

// ForceDeleteNodes deletes nodes from the group regardless of constraints.
func (vmPool *VMPool) ForceDeleteNodes(nodes []*apiv1.Node) error {
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

// TargetSize returns the current target size of the node group. This value represents
// the desired number of nodes in the VMPool, which may differ from the actual number
// of nodes currently present.
func (vmPool *VMPool) TargetSize() (int, error) {
	// VMs in the "Deleting" state are not counted towards the target size.
	size, err := vmPool.getCurSize(skipOption{skipDeleting: true, skipFailed: false})
	return int(size), err
}

// IncreaseSize increases the size of the VMPool by sending a PUT request to update the agent pool.
// This method waits until the asynchronous PUT operation completes or the client-side timeout is reached.
func (vmPool *VMPool) IncreaseSize(delta int) error {
	if delta <= 0 {
		return fmt.Errorf("size increase must be positive, current delta: %d", delta)
	}

	// Skip VMs in the failed state so that a PUT AP will be triggered to fix the failed VMs.
	currentSize, err := vmPool.getCurSize(skipOption{skipDeleting: true, skipFailed: true})
	if err != nil {
		return err
	}

	if int(currentSize)+delta > vmPool.MaxSize() {
		return fmt.Errorf("size-increasing request of %d is bigger than max size %d", int(currentSize)+delta, vmPool.MaxSize())
	}

	updateCtx, cancel := getContextWithTimeout(vmsAsyncContextTimeout)
	defer cancel()

	versionedAP, err := vmPool.getAgentpoolFromCache()
	if err != nil {
		klog.Errorf("Failed to get vmPool %s, error: %s", vmPool.agentPoolName, err)
		return err
	}

	count := currentSize + int32(delta)
	requestBody := armcontainerservice.AgentPool{}
	// self-hosted CAS will be using Manual scale profile
	if len(versionedAP.Properties.VirtualMachinesProfile.Scale.Manual) > 0 {
		requestBody = buildRequestBodyForScaleUp(versionedAP, count, vmPool.sku)

	}
	// hosted CAS will be using Autoscale scale profile
	// HostedSystem will be using manual scale profile
	// Both of them need to set the Target-Count and SKU headers
	if versionedAP.Properties.VirtualMachinesProfile.Scale.Autoscale != nil ||
		(versionedAP.Properties.Mode != nil &&
			strings.EqualFold(string(*versionedAP.Properties.Mode), "HostedSystem")) {
		header := make(http.Header)
		header.Set("Target-Count", fmt.Sprintf("%d", count))
		header.Set("SKU", fmt.Sprintf("%s", vmPool.sku))
		updateCtx = policy.WithHTTPHeader(updateCtx, header)
	}

	defer vmPool.manager.invalidateCache()
	poller, err := vmPool.manager.azClient.agentPoolClient.BeginCreateOrUpdate(
		updateCtx,
		vmPool.manager.config.ClusterResourceGroup,
		vmPool.manager.config.ClusterName,
		vmPool.agentPoolName,
		requestBody, nil)

	if err != nil {
		klog.Errorf("Failed to scale up agentpool %s in cluster %s for vmPool %s with error: %v",
			vmPool.agentPoolName, vmPool.manager.config.ClusterName, vmPool.Name, err)
		return err
	}

	if _, err := poller.PollUntilDone(updateCtx, nil /*default polling interval is 30s*/); err != nil {
		klog.Errorf("agentPoolClient.BeginCreateOrUpdate for aks cluster %s agentpool %s for scaling up vmPool %s failed with error %s",
			vmPool.manager.config.ClusterName, vmPool.agentPoolName, vmPool.Name, err)
		return err
	}

	klog.Infof("Successfully scaled up agentpool %s in cluster %s for vmPool %s to size %d",
		vmPool.agentPoolName, vmPool.manager.config.ClusterName, vmPool.Name, count)
	return nil
}

// buildRequestBodyForScaleUp builds the request body for scale up for self-hosted CAS
func buildRequestBodyForScaleUp(agentpool armcontainerservice.AgentPool, count int32, vmSku string) armcontainerservice.AgentPool {
	requestBody := armcontainerservice.AgentPool{
		Properties: &armcontainerservice.ManagedClusterAgentPoolProfileProperties{
			Type: agentpool.Properties.Type,
		},
	}

	// the request body must have the same mode as the original agentpool
	// otherwise the PUT request will fail
	if agentpool.Properties.Mode != nil &&
		*agentpool.Properties.Mode == armcontainerservice.AgentPoolModeSystem {
		systemMode := armcontainerservice.AgentPoolModeSystem
		requestBody.Properties.Mode = &systemMode
	}

	// set the count of the matching manual scale profile to the new target value
	for _, manualProfile := range agentpool.Properties.VirtualMachinesProfile.Scale.Manual {
		if manualProfile != nil && manualProfile.Size != nil &&
			strings.EqualFold(ptr.Deref(manualProfile.Size, ""), vmSku) {
			klog.V(5).Infof("Found matching manual profile for VM SKU: %s, updating count to: %d", vmSku, count)
			manualProfile.Count = ptr.To(count)
			requestBody.Properties.VirtualMachinesProfile = agentpool.Properties.VirtualMachinesProfile
			break
		}
	}
	return requestBody
}

// DeleteNodes removes the specified nodes from the VMPool by extracting their providerIDs
// and performing the appropriate delete or deallocate operation based on the agent pool's
// scale-down policy. This method waits for the asynchronous delete operation to complete,
// with a client-side timeout.
func (vmPool *VMPool) DeleteNodes(nodes []*apiv1.Node) error {
	// Ensure we don't scale below the minimum size by excluding VMs in the "Deleting" state.
	currentSize, err := vmPool.getCurSize(skipOption{skipDeleting: true, skipFailed: false})
	if err != nil {
		return fmt.Errorf("unable to retrieve current size: %w", err)
	}

	if int(currentSize) <= vmPool.MinSize() {
		return fmt.Errorf("cannot delete nodes as minimum size of %d has been reached", vmPool.MinSize())
	}

	providerIDs, err := vmPool.getProviderIDsForNodes(nodes)
	if err != nil {
		return fmt.Errorf("failed to retrieve provider IDs for nodes: %w", err)
	}

	if len(providerIDs) == 0 {
		return nil
	}

	klog.V(3).Infof("Deleting nodes from vmPool %s: %v", vmPool.Name, providerIDs)

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

	deleteCtx, cancel := getContextWithTimeout(vmsAsyncContextTimeout)
	defer cancel()
	defer vmPool.manager.invalidateCache()

	poller, err := vmPool.manager.azClient.agentPoolClient.BeginDeleteMachines(
		deleteCtx,
		vmPool.manager.config.ClusterResourceGroup,
		vmPool.manager.config.ClusterName,
		vmPool.agentPoolName,
		requestBody, nil)
	if err != nil {
		klog.Errorf("Failed to delete nodes from agentpool %s in cluster %s with error: %v",
			vmPool.agentPoolName, vmPool.manager.config.ClusterName, err)
		return err
	}

	if _, err := poller.PollUntilDone(deleteCtx, nil); err != nil {
		klog.Errorf("agentPoolClient.BeginDeleteMachines for aks cluster %s for scaling down vmPool %s failed with error %s",
			vmPool.manager.config.ClusterName, vmPool.agentPoolName, err)
		return err
	}
	klog.Infof("Successfully deleted %d nodes from vmPool %s", len(providerIDs), vmPool.Name)
	return nil
}

func (vmPool *VMPool) getProviderIDsForNodes(nodes []*apiv1.Node) ([]string, error) {
	var providerIDs []string
	for _, node := range nodes {
		belongs, err := vmPool.Belongs(node)
		if err != nil {
			return nil, fmt.Errorf("failed to check if node %s belongs to vmPool %s: %w", node.Name, vmPool.Name, err)
		}
		if !belongs {
			return nil, fmt.Errorf("node %s does not belong to vmPool %s", node.Name, vmPool.Name)
		}
		providerIDs = append(providerIDs, node.Spec.ProviderID)
	}
	return providerIDs, nil
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
	// The TargetSize of a VMPool is automatically adjusted after node deletions.
	// This method is invoked in scenarios such as (see details in clusterstate.go):
	// - len(readiness.Registered) > acceptableRange.CurrentTarget
	// - len(readiness.Registered) < acceptableRange.CurrentTarget - unregisteredNodes

	// For VMPool, this method should not be called because:
	// CurrentTarget = len(readiness.Registered) + unregisteredNodes - len(nodesInDeletingState)
	// Here, nodesInDeletingState is a subset of unregisteredNodes,
	// ensuring len(readiness.Registered) is always within the acceptable range.

	// here we just invalidate the cache to avoid any potential bugs
	vmPool.manager.invalidateCache()
	klog.Warningf("DecreaseTargetSize called for VMPool %s, but it should not be used, invalidating cache", vmPool.Name)
	return nil
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

// skipOption is used to determine whether to skip VMs in certain states when calculating the current size of the vmPool.
type skipOption struct {
	// skipDeleting indicates whether to skip VMs in the "Deleting" state.
	skipDeleting bool
	// skipFailed indicates whether to skip VMs in the "Failed" state.
	skipFailed bool
}

// getCurSize determines the current count of VMs in the vmPool, including unregistered ones.
// The source of truth depends on the pool type (spot or non-spot).
func (vmPool *VMPool) getCurSize(op skipOption) (int32, error) {
	agentPool, err := vmPool.getAgentpoolFromCache()
	if err != nil {
		klog.Errorf("Failed to retrieve agent pool %s from cache: %v", vmPool.agentPoolName, err)
		return -1, err
	}

	// spot pool size is retrieved directly from Azure instead of the cache
	if isSpotAgentPool(agentPool) {
		return vmPool.getSpotPoolSize()
	}

	// non-spot pool size is retrieved from the cache
	vms, err := vmPool.getVMsFromCache(op)
	if err != nil {
		klog.Errorf("Failed to get VMs from cache for agentpool %s with error: %v", vmPool.agentPoolName, err)
		return -1, err
	}
	return int32(len(vms)), nil
}

// getSpotPoolSize retrieves the current size of a spot agent pool directly from Azure.
func (vmPool *VMPool) getSpotPoolSize() (int32, error) {
	ap, err := vmPool.getAgentpoolFromAzure()
	if err != nil {
		klog.Errorf("Failed to get agentpool %s from Azure with error: %v", vmPool.agentPoolName, err)
		return -1, err
	}

	if ap.Properties != nil {
		// the VirtualMachineNodesStatus returned by AKS-RP is constructed from the vm list returned from CRP.
		// it only contains VMs in the running state.
		for _, status := range ap.Properties.VirtualMachineNodesStatus {
			if status != nil {
				if strings.EqualFold(ptr.Deref(status.Size, ""), vmPool.sku) {
					return ptr.Deref(status.Count, 0), nil
				}
			}
		}
	}
	return -1, fmt.Errorf("failed to get the size of spot agentpool %s", vmPool.agentPoolName)
}

// getVMsFromCache retrieves the list of virtual machines in this VMPool.
// If excludeDeleting is true, it skips VMs in the "Deleting" state.
// https://learn.microsoft.com/en-us/azure/virtual-machines/states-billing#provisioning-states
func (vmPool *VMPool) getVMsFromCache(op skipOption) ([]*armcompute.VirtualMachine, error) {
	vmsMap := vmPool.manager.azureCache.getVirtualMachines()
	var filteredVMs []*armcompute.VirtualMachine

	for _, vm := range vmsMap[vmPool.agentPoolName] {
		if vm.Properties == nil ||
			vm.Properties.HardwareProfile == nil ||
			vm.Properties.HardwareProfile.VMSize == nil ||
			!strings.EqualFold(string(*vm.Properties.HardwareProfile.VMSize), vmPool.sku) {
			continue
		}

		if op.skipDeleting && strings.Contains(ptr.Deref(vm.Properties.ProvisioningState, ""), "Deleting") {
			klog.V(4).Infof("Skipping VM %s in deleting state", ptr.Deref(vm.ID, ""))
			continue
		}

		if op.skipFailed && strings.Contains(ptr.Deref(vm.Properties.ProvisioningState, ""), "Failed") {
			klog.V(4).Infof("Skipping VM %s in failed state", ptr.Deref(vm.ID, ""))
			continue
		}

		filteredVMs = append(filteredVMs, vm)
	}
	return filteredVMs, nil
}

// Nodes returns the list of nodes in the vms agentPool.
func (vmPool *VMPool) Nodes() ([]cloudprovider.Instance, error) {
	vms, err := vmPool.getVMsFromCache(skipOption{}) // no skip option, get all VMs
	if err != nil {
		return nil, err
	}

	nodes := make([]cloudprovider.Instance, 0, len(vms))
	for _, vm := range vms {
		if vm.ID == nil || len(*vm.ID) == 0 {
			continue
		}
		resourceID, err := convertResourceGroupNameToLower("azure://" + ptr.Deref(vm.ID, ""))
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, cloudprovider.Instance{Id: resourceID})
	}

	return nodes, nil
}

// TemplateNodeInfo returns a NodeInfo object that can be used to create a new node in the vmPool.
func (vmPool *VMPool) TemplateNodeInfo() (*framework.NodeInfo, error) {
	ap, err := vmPool.getAgentpoolFromCache()
	if err != nil {
		return nil, err
	}
	inputLabels := map[string]string{}
	inputTaints := ""
	template, err := buildNodeTemplateFromVMPool(ap, vmPool.manager.config.Location, vmPool.sku, inputLabels, inputTaints)
	if err != nil {
		return nil, err
	}
	node, err := buildNodeFromTemplate(vmPool.agentPoolName, template, vmPool.manager, vmPool.manager.config.EnableDynamicInstanceList, false)
	if err != nil {
		return nil, err
	}

	nodeInfo := framework.NewNodeInfo(node, nil, framework.NewPodInfo(cloudprovider.BuildKubeProxy(vmPool.agentPoolName), nil))

	return nodeInfo, nil
}

func (vmPool *VMPool) getAgentpoolFromCache() (armcontainerservice.AgentPool, error) {
	vmsPoolMap := vmPool.manager.azureCache.getVMsPoolMap()
	if _, exists := vmsPoolMap[vmPool.agentPoolName]; !exists {
		return armcontainerservice.AgentPool{}, fmt.Errorf("VMs agent pool %s not found in cache", vmPool.agentPoolName)
	}
	return vmsPoolMap[vmPool.agentPoolName], nil
}

// getAgentpoolFromAzure returns the AKS agentpool from Azure
func (vmPool *VMPool) getAgentpoolFromAzure() (armcontainerservice.AgentPool, error) {
	ctx, cancel := getContextWithTimeout(vmsContextTimeout)
	defer cancel()
	resp, err := vmPool.manager.azClient.agentPoolClient.Get(
		ctx,
		vmPool.manager.config.ClusterResourceGroup,
		vmPool.manager.config.ClusterName,
		vmPool.agentPoolName, nil)
	if err != nil {
		return resp.AgentPool, fmt.Errorf("failed to get agentpool %s in cluster %s with error: %v",
			vmPool.agentPoolName, vmPool.manager.config.ClusterName, err)
	}
	return resp.AgentPool, nil
}

// AtomicIncreaseSize is not implemented.
func (vmPool *VMPool) AtomicIncreaseSize(delta int) error {
	return cloudprovider.ErrNotImplemented
}
