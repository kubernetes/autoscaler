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
	"math/rand"
	"strings"
	"sync"
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config/dynamic"
	klog "k8s.io/klog/v2"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
	"k8s.io/legacy-cloud-providers/azure/retry"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-12-01/compute"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/to"
)

var (
	defaultVmssInstancesRefreshPeriod = 5 * time.Minute
	vmssContextTimeout                = 3 * time.Minute
	vmssSizeMutex                     sync.Mutex
)

// ScaleSet implements NodeGroup interface.
type ScaleSet struct {
	azureRef
	manager *AzureManager

	minSize int
	maxSize int

	sizeMutex         sync.Mutex
	curSize           int64
	lastSizeRefresh   time.Time
	sizeRefreshPeriod time.Duration

	instancesRefreshPeriod time.Duration
	instancesRefreshJitter int

	instanceMutex       sync.Mutex
	instanceCache       []cloudprovider.Instance
	lastInstanceRefresh time.Time
}

// NewScaleSet creates a new NewScaleSet.
func NewScaleSet(spec *dynamic.NodeGroupSpec, az *AzureManager, curSize int64) (*ScaleSet, error) {
	scaleSet := &ScaleSet{
		azureRef: azureRef{
			Name: spec.Name,
		},
		minSize: spec.MinSize,
		maxSize: spec.MaxSize,
		manager: az,
		curSize: curSize,

		instancesRefreshJitter: az.config.VmssVmsCacheJitter,
	}

	if az.config.VmssVmsCacheTTL != 0 {
		scaleSet.instancesRefreshPeriod = time.Duration(az.config.VmssVmsCacheTTL) * time.Second
	} else {
		scaleSet.instancesRefreshPeriod = defaultVmssInstancesRefreshPeriod
	}

	return scaleSet, nil
}

// MinSize returns minimum size of the node group.
func (scaleSet *ScaleSet) MinSize() int {
	return scaleSet.minSize
}

// Exist checks if the node group really exists on the cloud provider side. Allows to tell the
// theoretical node group from the real one.
func (scaleSet *ScaleSet) Exist() bool {
	return true
}

// Create creates the node group on the cloud provider side.
func (scaleSet *ScaleSet) Create() (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrAlreadyExist
}

// Delete deletes the node group on the cloud provider side.
// This will be executed only for autoprovisioned node groups, once their size drops to 0.
func (scaleSet *ScaleSet) Delete() error {
	return cloudprovider.ErrNotImplemented
}

// Autoprovisioned returns true if the node group is autoprovisioned.
func (scaleSet *ScaleSet) Autoprovisioned() bool {
	return false
}

// MaxSize returns maximum size of the node group.
func (scaleSet *ScaleSet) MaxSize() int {
	return scaleSet.maxSize
}

func (scaleSet *ScaleSet) getVMSSFromCache() (compute.VirtualMachineScaleSet, *retry.Error) {
	allVMSS := scaleSet.manager.azureCache.getScaleSets()

	if _, exists := allVMSS[scaleSet.Name]; !exists {
		return compute.VirtualMachineScaleSet{}, &retry.Error{RawError: fmt.Errorf("could not find vmss: %s", scaleSet.Name)}
	}

	return allVMSS[scaleSet.Name], nil
}

func (scaleSet *ScaleSet) getCurSize() (int64, error) {
	scaleSet.sizeMutex.Lock()
	defer scaleSet.sizeMutex.Unlock()

	if scaleSet.lastSizeRefresh.Add(scaleSet.sizeRefreshPeriod).After(time.Now()) {
		return scaleSet.curSize, nil
	}

	klog.V(5).Infof("Get scale set size for %q", scaleSet.Name)
	set, rerr := scaleSet.getVMSSFromCache()
	if rerr != nil {
		return -1, rerr.Error()
	}

	// If VMSS state is updating, return the currentSize which would've been proactively incremented or decremented by CA
	if set.VirtualMachineScaleSetProperties != nil && strings.EqualFold(to.String(set.VirtualMachineScaleSetProperties.ProvisioningState), string(compute.ProvisioningStateUpdating)) {
		klog.V(5).Infof("VMSS %q is in updating state, returning cached size: %d", scaleSet.Name, scaleSet.curSize)
		return scaleSet.curSize, nil
	}

	vmssSizeMutex.Lock()
	curSize := *set.Sku.Capacity
	vmssSizeMutex.Unlock()

	klog.V(5).Infof("Getting scale set (%q) capacity: %d\n", scaleSet.Name, curSize)

	if scaleSet.curSize != curSize {
		// Invalidate the instance cache if the capacity has changed.
		scaleSet.invalidateInstanceCache()
	}

	scaleSet.curSize = curSize
	scaleSet.lastSizeRefresh = time.Now()
	return scaleSet.curSize, nil
}

// GetScaleSetSize gets Scale Set size.
func (scaleSet *ScaleSet) GetScaleSetSize() (int64, error) {
	return scaleSet.getCurSize()
}

func (scaleSet *ScaleSet) waitForDeleteInstances(future *azure.Future, requiredIds *compute.VirtualMachineScaleSetVMInstanceRequiredIDs) {
	ctx, cancel := getContextWithCancel()
	defer cancel()

	klog.V(3).Infof("Calling virtualMachineScaleSetsClient.WaitForAsyncOperationResult - DeleteInstances(%v)", requiredIds.InstanceIds)
	httpResponse, err := scaleSet.manager.azClient.virtualMachineScaleSetsClient.WaitForAsyncOperationResult(ctx, future)
	isSuccess, err := isSuccessHTTPResponse(httpResponse, err)
	if isSuccess {
		klog.V(3).Infof("virtualMachineScaleSetsClient.WaitForAsyncOperationResult - DeleteInstances(%v) success", requiredIds.InstanceIds)
		return
	}
	klog.Errorf("virtualMachineScaleSetsClient.WaitForAsyncOperationResult - DeleteInstances for instances %v failed with error: %v", requiredIds.InstanceIds, err)
}

// updateVMSSCapacity invokes virtualMachineScaleSetsClient to update the capacity for VMSS.
func (scaleSet *ScaleSet) updateVMSSCapacity(future *azure.Future) {
	var err error

	defer func() {
		if err != nil {
			klog.Errorf("Failed to update the capacity for vmss %s with error %v, invalidate the cache so as to get the real size from API", scaleSet.Name, err)
			// Invalidate the VMSS size cache in order to fetch the size from the API.
			scaleSet.manager.invalidateCache()
		}
	}()

	ctx, cancel := getContextWithCancel()
	defer cancel()

	klog.V(3).Infof("Calling virtualMachineScaleSetsClient.WaitForAsyncOperationResult - updateVMSSCapacity(%s)", scaleSet.Name)
	httpResponse, err := scaleSet.manager.azClient.virtualMachineScaleSetsClient.WaitForAsyncOperationResult(ctx, future)

	isSuccess, err := isSuccessHTTPResponse(httpResponse, err)
	if isSuccess {
		klog.V(3).Infof("virtualMachineScaleSetsClient.WaitForAsyncOperationResult - updateVMSSCapacity(%s) success", scaleSet.Name)
		scaleSet.invalidateInstanceCache()
		return
	}

	klog.Errorf("virtualMachineScaleSetsClient.WaitForAsyncOperationResult - updateVMSSCapacity for scale set %q failed: %v", scaleSet.Name, err)
}

// SetScaleSetSize sets ScaleSet size.
func (scaleSet *ScaleSet) SetScaleSetSize(size int64) error {
	scaleSet.sizeMutex.Lock()
	defer scaleSet.sizeMutex.Unlock()

	vmssInfo, rerr := scaleSet.getVMSSFromCache()
	if rerr != nil {
		klog.Errorf("Failed to get information for VMSS (%q): %v", scaleSet.Name, rerr)
		return rerr.Error()
	}

	// Update the new capacity to cache.
	vmssSizeMutex.Lock()
	vmssInfo.Sku.Capacity = &size
	vmssSizeMutex.Unlock()

	// Compose a new VMSS for updating.
	op := compute.VirtualMachineScaleSet{
		Name:     vmssInfo.Name,
		Sku:      vmssInfo.Sku,
		Location: vmssInfo.Location,
	}
	ctx, cancel := getContextWithTimeout(vmssContextTimeout)
	defer cancel()
	klog.V(3).Infof("Waiting for virtualMachineScaleSetsClient.CreateOrUpdateAsync(%s)", scaleSet.Name)
	future, rerr := scaleSet.manager.azClient.virtualMachineScaleSetsClient.CreateOrUpdateAsync(ctx, scaleSet.manager.config.ResourceGroup, scaleSet.Name, op)
	if rerr != nil {
		klog.Errorf("virtualMachineScaleSetsClient.CreateOrUpdate for scale set %q failed: %v", scaleSet.Name, rerr)
		return rerr.Error()
	}

	// Proactively set the VMSS size so autoscaler makes better decisions.
	scaleSet.curSize = size
	scaleSet.lastSizeRefresh = time.Now()

	go scaleSet.updateVMSSCapacity(future)
	return nil
}

// TargetSize returns the current TARGET size of the node group. It is possible that the
// number is different from the number of nodes registered in Kubernetes.
func (scaleSet *ScaleSet) TargetSize() (int, error) {
	size, err := scaleSet.GetScaleSetSize()
	return int(size), err
}

// IncreaseSize increases Scale Set size
func (scaleSet *ScaleSet) IncreaseSize(delta int) error {
	if delta <= 0 {
		return fmt.Errorf("size increase must be positive")
	}

	size, err := scaleSet.GetScaleSetSize()
	if err != nil {
		return err
	}

	if size == -1 {
		return fmt.Errorf("the scale set %s is under initialization, skipping IncreaseSize", scaleSet.Name)
	}

	if int(size)+delta > scaleSet.MaxSize() {
		return fmt.Errorf("size increase too large - desired:%d max:%d", int(size)+delta, scaleSet.MaxSize())
	}

	return scaleSet.SetScaleSetSize(size + int64(delta))
}

// GetScaleSetVms returns list of nodes for the given scale set.
func (scaleSet *ScaleSet) GetScaleSetVms() ([]compute.VirtualMachineScaleSetVM, *retry.Error) {
	klog.V(4).Infof("GetScaleSetVms: starts")
	ctx, cancel := getContextWithTimeout(vmssContextTimeout)
	defer cancel()

	resourceGroup := scaleSet.manager.config.ResourceGroup
	vmList, rerr := scaleSet.manager.azClient.virtualMachineScaleSetVMsClient.List(ctx, resourceGroup, scaleSet.Name, "")
	klog.V(4).Infof("GetScaleSetVms: scaleSet.Name: %s, vmList: %v", scaleSet.Name, vmList)
	if rerr != nil {
		klog.Errorf("VirtualMachineScaleSetVMsClient.List failed for %s: %v", scaleSet.Name, rerr)
		return nil, rerr
	}

	return vmList, nil
}

// DecreaseTargetSize decreases the target size of the node group. This function
// doesn't permit to delete any existing node and can be used only to reduce the
// request for new nodes that have not been yet fulfilled. Delta should be negative.
// It is assumed that cloud provider will not delete the existing nodes if the size
// when there is an option to just decrease the target.
func (scaleSet *ScaleSet) DecreaseTargetSize(delta int) error {
	// VMSS size should be changed automatically after the Node deletion, hence this operation is not required.
	// To prevent some unreproducible bugs, an extra refresh of cache is needed.
	scaleSet.invalidateInstanceCache()
	_, err := scaleSet.GetScaleSetSize()
	if err != nil {
		klog.Warningf("DecreaseTargetSize: failed with error: %v", err)
	}
	return err
}

// Belongs returns true if the given node belongs to the NodeGroup.
func (scaleSet *ScaleSet) Belongs(node *apiv1.Node) (bool, error) {
	klog.V(6).Infof("Check if node belongs to this scale set: scaleset:%v, node:%v\n", scaleSet, node)

	ref := &azureRef{
		Name: node.Spec.ProviderID,
	}

	targetAsg, err := scaleSet.manager.GetNodeGroupForInstance(ref)
	if err != nil {
		return false, err
	}
	if targetAsg == nil {
		return false, fmt.Errorf("%s doesn't belong to a known scale set", node.Name)
	}
	if !strings.EqualFold(targetAsg.Id(), scaleSet.Id()) {
		return false, nil
	}
	return true, nil
}

// DeleteInstances deletes the given instances. All instances must be controlled by the same ASG.
func (scaleSet *ScaleSet) DeleteInstances(instances []*azureRef) error {
	if len(instances) == 0 {
		return nil
	}

	klog.V(3).Infof("Deleting vmss instances %v", instances)

	commonAsg, err := scaleSet.manager.GetNodeGroupForInstance(instances[0])
	if err != nil {
		return err
	}

	instancesToDelete := []*azureRef{}
	for _, instance := range instances {
		asg, err := scaleSet.manager.GetNodeGroupForInstance(instance)
		if err != nil {
			return err
		}

		if !strings.EqualFold(asg.Id(), commonAsg.Id()) {
			return fmt.Errorf("cannot delete instance (%s) which don't belong to the same Scale Set (%q)", instance.Name, commonAsg)
		}

		if cpi, found := scaleSet.getInstanceByProviderID(instance.Name); found && cpi.Status != nil && cpi.Status.State == cloudprovider.InstanceDeleting {
			klog.V(3).Infof("Skipping deleting instance %s as its current state is deleting", instance.Name)
			continue
		}
		instancesToDelete = append(instancesToDelete, instance)
	}

	// nothing to delete
	if len(instancesToDelete) == 0 {
		klog.V(3).Infof("No new instances eligible for deletion, skipping")
		return nil
	}

	instanceIDs := []string{}
	for _, instance := range instancesToDelete {
		instanceID, err := getLastSegment(instance.Name)
		if err != nil {
			klog.Errorf("getLastSegment failed with error: %v", err)
			return err
		}
		instanceIDs = append(instanceIDs, instanceID)
	}

	requiredIds := &compute.VirtualMachineScaleSetVMInstanceRequiredIDs{
		InstanceIds: &instanceIDs,
	}

	ctx, cancel := getContextWithTimeout(vmssContextTimeout)
	defer cancel()
	resourceGroup := scaleSet.manager.config.ResourceGroup

	scaleSet.instanceMutex.Lock()
	klog.V(3).Infof("Calling virtualMachineScaleSetsClient.DeleteInstancesAsync(%v)", requiredIds.InstanceIds)
	future, rerr := scaleSet.manager.azClient.virtualMachineScaleSetsClient.DeleteInstancesAsync(ctx, resourceGroup, commonAsg.Id(), *requiredIds)
	scaleSet.instanceMutex.Unlock()
	if rerr != nil {
		klog.Errorf("virtualMachineScaleSetsClient.DeleteInstancesAsync for instances %v failed: %v", requiredIds.InstanceIds, rerr)
		return rerr.Error()
	}

	// Proactively decrement scale set size so that we don't
	// go below minimum node count if cache data is stale
	scaleSet.sizeMutex.Lock()
	scaleSet.curSize -= int64(len(instanceIDs))
	scaleSet.sizeMutex.Unlock()

	// Proactively set the status of the instances to be deleted in cache
	for _, instance := range instancesToDelete {
		scaleSet.setInstanceStatusByProviderID(instance.Name, cloudprovider.InstanceStatus{State: cloudprovider.InstanceDeleting})
	}

	go scaleSet.waitForDeleteInstances(future, requiredIds)

	return nil
}

// DeleteNodes deletes the nodes from the group.
func (scaleSet *ScaleSet) DeleteNodes(nodes []*apiv1.Node) error {
	klog.V(8).Infof("Delete nodes requested: %q\n", nodes)
	size, err := scaleSet.GetScaleSetSize()
	if err != nil {
		return err
	}

	if int(size) <= scaleSet.MinSize() {
		return fmt.Errorf("min size reached, nodes will not be deleted")
	}

	refs := make([]*azureRef, 0, len(nodes))
	for _, node := range nodes {
		belongs, err := scaleSet.Belongs(node)
		if err != nil {
			return err
		}

		if belongs != true {
			return fmt.Errorf("%s belongs to a different asg than %s", node.Name, scaleSet.Id())
		}

		ref := &azureRef{
			Name: node.Spec.ProviderID,
		}
		refs = append(refs, ref)
	}

	return scaleSet.DeleteInstances(refs)
}

// Id returns ScaleSet id.
func (scaleSet *ScaleSet) Id() string {
	return scaleSet.Name
}

// Debug returns a debug string for the Scale Set.
func (scaleSet *ScaleSet) Debug() string {
	return fmt.Sprintf("%s (%d:%d)", scaleSet.Id(), scaleSet.MinSize(), scaleSet.MaxSize())
}

// TemplateNodeInfo returns a node template for this scale set.
func (scaleSet *ScaleSet) TemplateNodeInfo() (*schedulerframework.NodeInfo, error) {
	template, rerr := scaleSet.getVMSSFromCache()
	if rerr != nil {
		return nil, rerr.Error()
	}

	node, err := buildNodeFromTemplate(scaleSet.Name, template)
	if err != nil {
		return nil, err
	}

	nodeInfo := schedulerframework.NewNodeInfo(cloudprovider.BuildKubeProxy(scaleSet.Name))
	nodeInfo.SetNode(node)
	return nodeInfo, nil
}

// Nodes returns a list of all nodes that belong to this node group.
func (scaleSet *ScaleSet) Nodes() ([]cloudprovider.Instance, error) {
	klog.V(4).Infof("Nodes: starts, scaleSet.Name: %s", scaleSet.Name)
	curSize, err := scaleSet.getCurSize()
	if err != nil {
		klog.Errorf("Failed to get current size for vmss %q: %v", scaleSet.Name, err)
		return nil, err
	}

	scaleSet.instanceMutex.Lock()
	defer scaleSet.instanceMutex.Unlock()

	if int64(len(scaleSet.instanceCache)) == curSize &&
		scaleSet.lastInstanceRefresh.Add(scaleSet.instancesRefreshPeriod).After(time.Now()) {
		klog.V(4).Infof("Nodes: returns with curSize %d", curSize)
		return scaleSet.instanceCache, nil
	}

	klog.V(4).Infof("Nodes: starts to get VMSS VMs")
	splay := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(scaleSet.instancesRefreshJitter + 1)
	lastRefresh := time.Now().Add(-time.Second * time.Duration(splay))

	vms, rerr := scaleSet.GetScaleSetVms()
	if rerr != nil {
		if isAzureRequestsThrottled(rerr) {
			// Log a warning and update the instance refresh time so that it would retry after cache expiration
			klog.Warningf("GetScaleSetVms() is throttled with message %v, would return the cached instances", rerr)
			scaleSet.lastInstanceRefresh = lastRefresh
			return scaleSet.instanceCache, nil
		}
		return nil, rerr.Error()
	}

	scaleSet.instanceCache = buildInstanceCache(vms)
	scaleSet.lastInstanceRefresh = lastRefresh
	klog.V(4).Infof("Nodes: returns")
	return scaleSet.instanceCache, nil
}

// Note that the GetScaleSetVms() results is not used directly because for the List endpoint,
// their resource ID format is not consistent with Get endpoint
func buildInstanceCache(vms []compute.VirtualMachineScaleSetVM) []cloudprovider.Instance {
	instances := []cloudprovider.Instance{}

	for _, vm := range vms {
		// The resource ID is empty string, which indicates the instance may be in deleting state.
		if len(*vm.ID) == 0 {
			continue
		}

		resourceID, err := convertResourceGroupNameToLower(*vm.ID)
		if err != nil {
			// This shouldn't happen. Log a warning message for tracking.
			klog.Warningf("buildInstanceCache.convertResourceGroupNameToLower failed with error: %v", err)
			continue
		}

		instances = append(instances, cloudprovider.Instance{
			Id:     "azure://" + resourceID,
			Status: instanceStatusFromVM(vm),
		})
	}

	return instances
}

func (scaleSet *ScaleSet) getInstanceByProviderID(providerID string) (cloudprovider.Instance, bool) {
	scaleSet.instanceMutex.Lock()
	defer scaleSet.instanceMutex.Unlock()
	for _, instance := range scaleSet.instanceCache {
		if instance.Id == providerID {
			return instance, true
		}
	}
	return cloudprovider.Instance{}, false
}

func (scaleSet *ScaleSet) setInstanceStatusByProviderID(providerID string, status cloudprovider.InstanceStatus) {
	scaleSet.instanceMutex.Lock()
	defer scaleSet.instanceMutex.Unlock()
	for k, instance := range scaleSet.instanceCache {
		if instance.Id == providerID {
			klog.V(5).Infof("Setting instance %s status to %v", instance.Id, status)
			scaleSet.instanceCache[k].Status = &status
		}
	}
}

// instanceStatusFromVM converts the VM provisioning state to cloudprovider.InstanceStatus
func instanceStatusFromVM(vm compute.VirtualMachineScaleSetVM) *cloudprovider.InstanceStatus {
	if vm.ProvisioningState == nil {
		return nil
	}

	status := &cloudprovider.InstanceStatus{}
	switch *vm.ProvisioningState {
	case string(compute.ProvisioningStateDeleting):
		status.State = cloudprovider.InstanceDeleting
	case string(compute.ProvisioningStateCreating):
		status.State = cloudprovider.InstanceCreating
	default:
		status.State = cloudprovider.InstanceRunning
	}

	return status
}

func (scaleSet *ScaleSet) invalidateInstanceCache() {
	scaleSet.instanceMutex.Lock()
	// Set the instanceCache as outdated.
	scaleSet.lastInstanceRefresh = time.Now().Add(-1 * scaleSet.instancesRefreshPeriod)
	scaleSet.instanceMutex.Unlock()
}
