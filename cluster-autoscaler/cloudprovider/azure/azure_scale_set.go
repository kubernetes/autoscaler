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
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/config/dynamic"
	klog "k8s.io/klog/v2"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
	"sigs.k8s.io/cloud-provider-azure/pkg/retry"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2022-08-01/compute"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/to"
)

var (
	defaultVmssInstancesRefreshPeriod = 5 * time.Minute
	vmssContextTimeout                = 3 * time.Minute
	asyncContextTimeout               = 30 * time.Minute
	vmssSizeMutex                     sync.Mutex
)

const (
	provisioningStateCreating  string = "Creating"
	provisioningStateDeleting  string = "Deleting"
	provisioningStateFailed    string = "Failed"
	provisioningStateMigrating string = "Migrating"
	provisioningStateSucceeded string = "Succeeded"
	provisioningStateUpdating  string = "Updating"
)

// ScaleSet implements NodeGroup interface.
type ScaleSet struct {
	azureRef
	manager *AzureManager

	minSize int
	maxSize int

	enableForceDelete         bool
	enableDynamicInstanceList bool

	// curSize tracks (and caches) the number of VMs in this ScaleSet.
	// It is periodically updated from vmss.Sku.Capacity, with VMSS itself coming
	// either from azure.Cache (which periodically does VMSS.List)
	// or from direct VMSS.Get (used for Spot).
	curSize int64
	// lastSizeRefresh is the time curSize was last refreshed from vmss.Sku.Capacity.
	// Together with sizeRefreshPeriod, it is used to determine if it is time to refresh curSize.
	lastSizeRefresh time.Time
	// sizeRefreshPeriod is how often curSize is refreshed from vmss.Sku.Capacity.
	// (Set from azureCache.refreshInterval = VmssCacheTTL or [defaultMetadataCache]refreshInterval = 1min)
	sizeRefreshPeriod time.Duration
	// getVmssSizeRefreshPeriod is how often curSize should be refreshed in case VMSS.Get call is used (only spot instances).
	// (Set from GetVmssSizeRefreshPeriod, if specified = get-vmss-size-refresh-period = 30s,
	//   or override from autoscalerProfile.GetVmssSizeRefreshPeriod)
	getVmssSizeRefreshPeriod time.Duration

	instancesRefreshPeriod time.Duration
	instancesRefreshJitter int

	sizeMutex           sync.Mutex
	instanceMutex       sync.Mutex
	instanceCache       []cloudprovider.Instance
	lastInstanceRefresh time.Time

	// uses Azure Dedicated Host
	dedicatedHost bool

	scaleDownPolicy cloudprovider.ScaleDownPolicy
}

// NewScaleSet creates a new NewScaleSet.
func NewScaleSet(spec *dynamic.NodeGroupSpec, az *AzureManager, curSize int64, dedicatedHost bool) (*ScaleSet, error) {
	scaleSet := &ScaleSet{
		azureRef: azureRef{
			Name: spec.Name,
		},
		minSize:                   spec.MinSize,
		maxSize:                   spec.MaxSize,
		manager:                   az,
		curSize:                   curSize,
		sizeRefreshPeriod:         az.azureCache.refreshInterval,
		enableDynamicInstanceList: az.config.EnableDynamicInstanceList,
		instancesRefreshJitter:    az.config.VmssVmsCacheJitter,
		enableForceDelete:         az.config.EnableForceDelete,
		dedicatedHost:             dedicatedHost,
	}

	if az.config.VmssVmsCacheTTL != 0 {
		scaleSet.instancesRefreshPeriod = time.Duration(az.config.VmssVmsCacheTTL) * time.Second
	} else {
		scaleSet.instancesRefreshPeriod = defaultVmssInstancesRefreshPeriod
	}

	if az.config.GetVmssSizeRefreshPeriod != 0 {
		scaleSet.getVmssSizeRefreshPeriod = time.Duration(az.config.GetVmssSizeRefreshPeriod) * time.Second
	} else {
		scaleSet.getVmssSizeRefreshPeriod = time.Duration(VmssSizeRefreshPeriodDefault) * time.Second
	}

	switch az.config.ScaleDownPolicy {
	case "":
		scaleSet.scaleDownPolicy = cloudprovider.Delete
	case string(cloudprovider.Delete), string(cloudprovider.Hibernate):
		scaleSet.scaleDownPolicy = cloudprovider.ScaleDownPolicy(az.config.ScaleDownPolicy)
	default:
		return nil, fmt.Errorf("unsupported Azure ScaleDownPolicy config: %s not recognized", az.config.ScaleDownPolicy)
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

// GetOptions returns NodeGroupAutoscalingOptions that should be used for this particular
// NodeGroup. Returning a nil will result in using default options.
func (scaleSet *ScaleSet) GetOptions(defaults config.NodeGroupAutoscalingOptions) (*config.NodeGroupAutoscalingOptions, error) {
	template, err := scaleSet.getVMSSFromCache()
	if err != nil {
		return nil, err
	}
	return scaleSet.manager.GetScaleSetOptions(*template.Name, defaults), nil
}

// ScaleDownPolicy returns the policy for the node group on scale downs. Whether Delete or Hibernate.
func (scaleSet *ScaleSet) ScaleDownPolicy() cloudprovider.ScaleDownPolicy {
	return scaleSet.scaleDownPolicy
}

// MaxSize returns maximum size of the node group.
func (scaleSet *ScaleSet) MaxSize() int {
	return scaleSet.maxSize
}

func (scaleSet *ScaleSet) getVMSSFromCache() (compute.VirtualMachineScaleSet, error) {
	allVMSS := scaleSet.manager.azureCache.getScaleSets()

	if _, exists := allVMSS[scaleSet.Name]; !exists {
		return compute.VirtualMachineScaleSet{}, fmt.Errorf("could not find vmss: %s", scaleSet.Name)
	}

	return allVMSS[scaleSet.Name], nil
}

func (scaleSet *ScaleSet) getCurSize() (int64, error) {
	scaleSet.sizeMutex.Lock()
	defer scaleSet.sizeMutex.Unlock()

	set, err := scaleSet.getVMSSFromCache()
	if err != nil {
		klog.Errorf("failed to get information for VMSS: %s, error: %v", scaleSet.Name, err)
		return -1, err
	}

	effectiveSizeRefreshPeriod := scaleSet.sizeRefreshPeriod

	// If the scale set is Spot, we want to have a more fresh view of the Sku.Capacity field.
	// This is because evictions can happen at any given point in time,
	// even before VMs are materialized as nodes. We should be able to
	// react to those and have the autoscaler readjust the goal again to force restoration.
	// Taking into account only if orchestrationMode == Uniform because flex mode can have
	// combination of spot and regular vms
	if isSpot(&set) {
		effectiveSizeRefreshPeriod = scaleSet.getVmssSizeRefreshPeriod
	}

	if scaleSet.lastSizeRefresh.Add(effectiveSizeRefreshPeriod).After(time.Now()) {
		klog.V(3).Infof("VMSS: %s, returning in-memory size: %d", scaleSet.Name, scaleSet.curSize)
		return scaleSet.curSize, nil
	}

	// If the scale set is on Spot, make a GET VMSS call to fetch more updated fresh info
	if isSpot(&set) {
		ctx, cancel := getContextWithCancel()
		defer cancel()

		var rerr *retry.Error
		set, rerr = scaleSet.manager.azClient.virtualMachineScaleSetsClient.Get(ctx, scaleSet.manager.config.ResourceGroup,
			scaleSet.Name)
		if rerr != nil {
			klog.Errorf("failed to get information for VMSS: %s, error: %v", scaleSet.Name, rerr)
			return -1, err
		}
	}

	vmssSizeMutex.Lock()
	curSize := *set.Sku.Capacity
	vmssSizeMutex.Unlock()

	if scaleSet.curSize != curSize {
		// Invalidate the instance cache if the capacity has changed.
		klog.V(5).Infof("VMSS %q size changed from: %d to %d, invalidating instance cache", scaleSet.Name, scaleSet.curSize, curSize)
		scaleSet.invalidateInstanceCache()
	}
	klog.V(3).Infof("VMSS: %s, in-memory size: %d, new size: %d", scaleSet.Name, scaleSet.curSize, curSize)

	scaleSet.curSize = curSize
	scaleSet.lastSizeRefresh = time.Now()
	return scaleSet.curSize, nil
}

func isSpot(vmss *compute.VirtualMachineScaleSet) bool {
	return vmss != nil && vmss.VirtualMachineScaleSetProperties != nil &&
		vmss.VirtualMachineScaleSetProperties.VirtualMachineProfile != nil &&
		vmss.VirtualMachineScaleSetProperties.VirtualMachineProfile.Priority == compute.Spot
}

// GetScaleSetSize gets Scale Set size.
func (scaleSet *ScaleSet) GetScaleSetSize() (int64, error) {
	// First, get the size of the ScaleSet reported by API
	// -1 indiciates the ScaleSet hasn't been initialized
	size, err := scaleSet.getCurSize()
	if size == -1 || err != nil {
		klog.V(3).Infof("getScaleSetSize: either size is -1 (actual: %d) or error exists (actual err:%v)", size, err)
		return size, err
	}
	// If the policy for this ScaleSet is Hibernate, the TargetSize is the capacity reported by VMSS minus the nodes
	// in deallocated and deallocating states
	if scaleSet.scaleDownPolicy == cloudprovider.Hibernate {
		totalDeallocationInstances, err := scaleSet.countDeallocatedInstances()
		if err != nil {
			klog.Errorf("getScaleSetSize: error countDeallocatedInstances for scaleSet %s,"+
				"err: %v", scaleSet.Name, err)
			return -1, err
		}
		size -= int64(totalDeallocationInstances)
		klog.V(3).Infof("Found: %d instances in deallocated state, returning target size: %d for scaleSet %s",
			totalDeallocationInstances, size, scaleSet.Name)
	}
	return size, nil
}

func (scaleSet *ScaleSet) waitForDeleteInstances(future *azure.Future, requiredIds *compute.VirtualMachineScaleSetVMInstanceRequiredIDs) {
	ctx, cancel := getContextWithCancel()
	defer cancel()

	klog.V(3).Infof("Calling virtualMachineScaleSetsClient.WaitForDeleteInstancesResult(%v) for %s", requiredIds.InstanceIds, scaleSet.Name)
	httpResponse, err := scaleSet.manager.azClient.virtualMachineScaleSetsClient.WaitForDeleteInstancesResult(ctx, future, scaleSet.manager.config.ResourceGroup)
	isSuccess, err := isSuccessHTTPResponse(httpResponse, err)
	if isSuccess {
		klog.V(3).Infof("virtualMachineScaleSetsClient.WaitForDeleteInstancesResult(%v) for %s success", requiredIds.InstanceIds, scaleSet.Name)
		return
	}
	klog.Errorf("virtualMachineScaleSetsClient.WaitForDeleteInstancesResult - DeleteInstances for instances %v for %s failed with error: %v", requiredIds.InstanceIds, scaleSet.Name, err)
}

// updateVMSSCapacity invokes virtualMachineScaleSetsClient to update the capacity for VMSS.
func (scaleSet *ScaleSet) updateVMSSCapacity(future *azure.Future) {
	var err error

	defer func() {
		if err != nil {
			klog.Errorf("Failed to update the capacity for vmss %s with error %v, invalidate the cache so as to get the real size from API", scaleSet.Name, err)
			// Invalidate the VMSS size cache in order to fetch the size from the API.
			scaleSet.invalidateLastSizeRefreshWithLock()
			scaleSet.manager.invalidateCache()
		}
	}()

	ctx, cancel := getContextWithCancel()
	defer cancel()

	klog.V(3).Infof("Calling virtualMachineScaleSetsClient.WaitForCreateOrUpdateResult(%s)", scaleSet.Name)
	httpResponse, err := scaleSet.manager.azClient.virtualMachineScaleSetsClient.WaitForCreateOrUpdateResult(ctx, future, scaleSet.manager.config.ResourceGroup)

	isSuccess, err := isSuccessHTTPResponse(httpResponse, err)
	if isSuccess {
		klog.V(3).Infof("virtualMachineScaleSetsClient.WaitForCreateOrUpdateResult(%s) success", scaleSet.Name)
		scaleSet.invalidateInstanceCache()
		return
	}

	klog.Errorf("virtualMachineScaleSetsClient.WaitForCreateOrUpdateResult - updateVMSSCapacity for scale set %q failed: %v", scaleSet.Name, err)
}

func (scaleSet *ScaleSet) createOrUpdateInstances(vmssInfo *compute.VirtualMachineScaleSet, newSize int64) error {
	if vmssInfo == nil {
		return fmt.Errorf("vmssInfo cannot be nil while increating scaleSet capacity")
	}

	scaleSet.sizeMutex.Lock()
	defer scaleSet.sizeMutex.Unlock()

	// Update the new capacity to cache.
	vmssSizeMutex.Lock()
	vmssInfo.Sku.Capacity = &newSize
	vmssSizeMutex.Unlock()

	// Compose a new VMSS for updating.
	op := compute.VirtualMachineScaleSet{
		Name:     vmssInfo.Name,
		Sku:      vmssInfo.Sku,
		Location: vmssInfo.Location,
	}

	if vmssInfo.ExtendedLocation != nil {
		op.ExtendedLocation = &compute.ExtendedLocation{
			Name: vmssInfo.ExtendedLocation.Name,
			Type: vmssInfo.ExtendedLocation.Type,
		}

		klog.V(3).Infof("Passing ExtendedLocation information if it is not nil, with Edge Zone name:(%s)", *op.ExtendedLocation.Name)
	}

	ctx, cancel := getContextWithTimeout(vmssContextTimeout)
	defer cancel()
	klog.V(3).Infof("Waiting for virtualMachineScaleSetsClient.CreateOrUpdateAsync(%s)", scaleSet.Name)
	future, rerr := scaleSet.manager.azClient.virtualMachineScaleSetsClient.CreateOrUpdateAsync(ctx,
		scaleSet.manager.config.ResourceGroup, scaleSet.Name, op)
	if rerr != nil {
		klog.Errorf("virtualMachineScaleSetsClient.CreateOrUpdate for scale set %q failed: %+v", scaleSet.Name, rerr)
		return azureToAutoscalerError(rerr)
	}

	// Proactively set the VMSS size so autoscaler makes better decisions.
	scaleSet.curSize = newSize
	scaleSet.lastSizeRefresh = time.Now()

	go scaleSet.waitForCreateOrUpdateInstances(future)
	return nil
}

// waitForCreateOrUpdate waits for the outcome of VMSS capacity update initiated via CreateOrUpdateAsync.
func (scaleSet *ScaleSet) waitForCreateOrUpdateInstances(future *azure.Future) {
	var err error

	defer func() {
		// Invalidate instanceCache on success and failure. Failure might have created a few instances, but it is very rare.
		scaleSet.invalidateInstanceCache()
		if err != nil {
			klog.Errorf("Failed to update the capacity for vmss %s with error %v, invalidate the cache so as to get "+
				"the real size from API", scaleSet.Name, err)
			// Invalidate the VMSS size cache in order to fetch the size from the API.
			scaleSet.invalidateLastSizeRefreshWithLock()
			scaleSet.manager.invalidateCache()
		}
	}()

	ctx, cancel := getContextWithTimeout(asyncContextTimeout)
	defer cancel()

	klog.V(3).Infof("Calling virtualMachineScaleSetsClient.WaitForCreateOrUpdateResult(%s)", scaleSet.Name)
	httpResponse, err := scaleSet.manager.azClient.virtualMachineScaleSetsClient.WaitForCreateOrUpdateResult(ctx,
		future, scaleSet.manager.config.ResourceGroup)
	if httpResponse != nil && httpResponse.Body != nil {
		defer httpResponse.Body.Close()
	}
	isSuccess, err := isSuccessHTTPResponse(httpResponse, err)

	if isSuccess {
		klog.V(3).Infof("waitForCreateOrUpdateInstances(%s) success", scaleSet.Name)
		return
	}

	klog.Errorf("waitForCreateOrUpdateInstances(%s) failed, err: %v", scaleSet.Name, err)
}

// SetScaleSetSize sets ScaleSet size.
func (scaleSet *ScaleSet) SetScaleSetSize(size int64, delta int) error {
	vmssInfo, err := scaleSet.getVMSSFromCache()
	if err != nil {
		klog.Errorf("Failed to get information for VMSS (%q): %v", scaleSet.Name, err)
		return err
	}

	requiredInstances := delta
	// If the policy is hibernate, then attempt to satisfy the request by starting existing instances
	if scaleSet.scaleDownPolicy == cloudprovider.Hibernate {
		deallocatedInstances, err := scaleSet.getDeallocatedInstances()
		if err != nil {
			klog.Errorf("SetScaleSetSize: error getDeallocatedInstances for scaleSet %s, "+
				"err : %v", scaleSet.Name, err)
		} else {
			klog.V(3).Infof("Attempting to start: %d instances from deallocated state", requiredInstances)

			// Go through current instances and attempt to reallocate them
			for _, instance := range deallocatedInstances {
				// we're done
				if requiredInstances <= 0 {
					break
				}
				instancesToStart := []*azureRef{{Name: instance.Id}}
				err := scaleSet.startInstances(instancesToStart)
				if err != nil {
					klog.Errorf("Failed to start instances %v in scale set %q: %v", instancesToStart, scaleSet.Name, err)
					continue
				}
				klog.V(3).Infof("Successfully started instances %v in scale set %q", instancesToStart, scaleSet.Name)
				requiredInstances--
			}
		}
	}

	// If after reallocating instances we still need more instances or we're just in Delete mode
	// send a scale request
	if requiredInstances > 0 {
		klog.V(3).Infof("Remaining unsatisfied count is %d. Attempting to increase scale set %q "+
			"capacity", requiredInstances, scaleSet.Name)
		err := scaleSet.createOrUpdateInstances(&vmssInfo, size)
		if err != nil {
			klog.Errorf("Failed to increase capacity for scale set %q to %d: %v", scaleSet.Name, requiredInstances, err)
			return err
		}
	}
	return nil
}

// startInstances starts the given instances. All instances must be controlled by the same nodegroup.
func (scaleSet *ScaleSet) startInstances(instances []*azureRef) error {
	if len(instances) == 0 {
		return nil
	}

	klog.V(3).Infof("Starting vmss instances %v", instances)

	commonNg, err := scaleSet.manager.GetNodeGroupForInstance(instances[0])
	if err != nil {
		return err
	}

	instancesToStart := []*azureRef{}
	for _, instance := range instances {
		err = scaleSet.verifyNodeGroup(instance, commonNg.Id())
		if err != nil {
			return err
		}

		if cpi, found, err := scaleSet.getInstanceByProviderID(instance.Name); found && err == nil && cpi.Status != nil &&
			cpi.Status.State == cloudprovider.InstanceRunning {
			klog.V(3).Infof("Skipping deleting instance %s as its current state is running", instance.Name)
			continue
		}
		instancesToStart = append(instancesToStart, instance)
	}

	// nothing to delete
	if len(instancesToStart) == 0 {
		klog.V(3).Infof("No new instances eligible for starting, skipping")
		return nil
	}

	instanceIDs := []string{}
	for _, instance := range instancesToStart {
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
	klog.V(3).Infof("Calling virtualMachineScaleSetsClient.StartInstancesAsync(%v) for %s",
		requiredIds.InstanceIds, scaleSet.Name)
	future, rerr := scaleSet.manager.azClient.virtualMachineScaleSetsClient.StartInstancesAsync(ctx, resourceGroup,
		commonNg.Id(), *requiredIds)
	scaleSet.instanceMutex.Unlock()
	if rerr != nil {
		klog.Errorf("virtualMachineScaleSetsClient.StartInstancesAsync for instances %v for %s failed: "+
			"%+v", requiredIds.InstanceIds, scaleSet.Name, rerr)
		return rerr.Error()
	}

	// Proactively set the status of the instances to be running in cache
	for _, instance := range instancesToStart {
		scaleSet.setInstanceStatusByProviderID(instance.Name, cloudprovider.InstanceStatus{State: cloudprovider.InstanceRunning})
	}

	go scaleSet.waitForStartInstances(future, requiredIds)
	return nil
}

func (scaleSet *ScaleSet) waitForStartInstances(future *azure.Future, requiredIds *compute.VirtualMachineScaleSetVMInstanceRequiredIDs) {
	ctx, cancel := getContextWithCancel()
	defer cancel()

	klog.V(3).Infof("Calling virtualMachineScaleSetsClient.WaitForStartInstancesResult(%v) for %s", requiredIds.InstanceIds, scaleSet.Name)
	httpResponse, err := scaleSet.manager.azClient.virtualMachineScaleSetsClient.WaitForStartInstancesResult(ctx, future,
		scaleSet.manager.config.ResourceGroup)
	if httpResponse != nil && httpResponse.Body != nil {
		defer httpResponse.Body.Close()
	}
	isSuccess, err := isSuccessHTTPResponse(httpResponse, err)

	if isSuccess {
		klog.V(3).Infof("WaitForStartInstancesResult(%v) for %s success", requiredIds.InstanceIds, scaleSet.Name)
		// No need to invalidateInstanceCache because the states were proactively set to Running.
		return
	}

	scaleSet.invalidateInstanceCache()
	klog.Errorf("WaitForStartInstancesResult(%v) for %s failed with error: %v",
		requiredIds.InstanceIds, scaleSet.Name, err)
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
	klog.Infof("Scaleset %s has size %d, will increase by %d", scaleSet.Name, size, delta)
	if err != nil {
		return err
	}

	if size == -1 {
		return fmt.Errorf("the scale set %s is under initialization, skipping IncreaseSize", scaleSet.Name)
	}

	if int(size)+delta > scaleSet.MaxSize() {
		return fmt.Errorf("size increase too large - desired:%d max:%d", int(size)+delta, scaleSet.MaxSize())
	}

	return scaleSet.SetScaleSetSize(size+int64(delta), delta)
}

// AtomicIncreaseSize is not implemented.
func (scaleSet *ScaleSet) AtomicIncreaseSize(delta int) error {
	return cloudprovider.ErrNotImplemented
}

// GetScaleSetVms returns list of nodes for the given scale set.
func (scaleSet *ScaleSet) GetScaleSetVms() ([]compute.VirtualMachineScaleSetVM, *retry.Error) {
	klog.V(4).Infof("GetScaleSetVms: starts")
	ctx, cancel := getContextWithTimeout(vmssContextTimeout)
	defer cancel()

	resourceGroup := scaleSet.manager.config.ResourceGroup
	vmList, rerr := scaleSet.manager.azClient.virtualMachineScaleSetVMsClient.List(ctx, resourceGroup, scaleSet.Name, "instanceView")
	klog.V(4).Infof("GetScaleSetVms: scaleSet.Name: %s, vmList: %v", scaleSet.Name, vmList)
	if rerr != nil {
		klog.Errorf("VirtualMachineScaleSetVMsClient.List failed for %s: %v", scaleSet.Name, rerr)
		return nil, rerr
	}

	return vmList, nil
}

// GetFlexibleScaleSetVms returns list of nodes for flexible scale set.
func (scaleSet *ScaleSet) GetFlexibleScaleSetVms() ([]compute.VirtualMachine, *retry.Error) {
	klog.V(4).Infof("GetScaleSetVms: starts")
	ctx, cancel := getContextWithTimeout(vmssContextTimeout)
	defer cancel()

	// get VMSS info from cache to obtain ID currently scaleSet does not store ID info.
	vmssInfo, err := scaleSet.getVMSSFromCache()

	if err != nil {
		klog.Errorf("Failed to get information for VMSS (%q): %v", scaleSet.Name, err)
		var rerr = &retry.Error{
			RawError: err,
		}
		return nil, rerr
	}
	vmList, rerr := scaleSet.manager.azClient.virtualMachinesClient.ListVmssFlexVMsWithoutInstanceView(ctx, *vmssInfo.ID)
	if rerr != nil {
		klog.Errorf("VirtualMachineScaleSetVMsClient.List failed for %s: %v", scaleSet.Name, rerr)
		return nil, rerr
	}
	klog.V(4).Infof("GetFlexibleScaleSetVms: scaleSet.Name: %s, vmList: %v", scaleSet.Name, vmList)
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
func (scaleSet *ScaleSet) DeleteInstances(instances []*azureRef, hasUnregisteredNodes bool) error {
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

		if cpi, found, err := scaleSet.getInstanceByProviderID(instance.Name); found && err == nil && cpi.Status != nil && cpi.Status.State == cloudprovider.InstanceDeleting {
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
	klog.V(3).Infof("Calling virtualMachineScaleSetsClient.DeleteInstancesAsync(%v), force delete set to %v", requiredIds.InstanceIds, scaleSet.enableForceDelete)
	future, rerr := scaleSet.manager.azClient.virtualMachineScaleSetsClient.DeleteInstancesAsync(ctx, resourceGroup, commonAsg.Id(), *requiredIds, scaleSet.enableForceDelete)

	if scaleSet.enableForceDelete && isOperationNotAllowed(rerr) {
		klog.Infof("falling back to normal delete for instances %v for %s", requiredIds.InstanceIds, scaleSet.Name)
		future, rerr = scaleSet.manager.azClient.virtualMachineScaleSetsClient.DeleteInstancesAsync(ctx, resourceGroup,
			commonAsg.Id(), *requiredIds, false)
	}

	scaleSet.instanceMutex.Unlock()
	if rerr != nil {
		klog.Errorf("virtualMachineScaleSetsClient.DeleteInstancesAsync for instances %v failed: %v", requiredIds.InstanceIds, rerr)
		return rerr.Error()
	}

	// Proactively decrement scale set size so that we don't
	// go below minimum node count if cache data is stale
	// only do it for non-unregistered nodes
	if !hasUnregisteredNodes {
		scaleSet.sizeMutex.Lock()
		scaleSet.curSize -= int64(len(instanceIDs))
		scaleSet.lastSizeRefresh = time.Now()
		scaleSet.sizeMutex.Unlock()
	}

	// Proactively set the status of the instances to be deleted in cache
	for _, instance := range instancesToDelete {
		scaleSet.setInstanceStatusByProviderID(instance.Name, cloudprovider.InstanceStatus{State: cloudprovider.InstanceDeleting})
	}

	go scaleSet.waitForDeleteInstances(future, requiredIds)

	return nil
}

// DeleteNodes deletes the nodes from the group.
func (scaleSet *ScaleSet) DeleteNodes(nodes []*apiv1.Node) error {
	klog.V(4).Infof("Delete nodes using scaleDownPolicy=%s requested: %q\n", scaleSet.scaleDownPolicy, nodes)
	size, err := scaleSet.GetScaleSetSize()
	if err != nil {
		return err
	}

	if int(size) <= scaleSet.MinSize() {
		return fmt.Errorf("min size reached, nodes will not be deleted")
	}

	refs := make([]*azureRef, 0, len(nodes))
	hasUnregisteredNodes := false
	for _, node := range nodes {
		belongs, err := scaleSet.Belongs(node)
		if err != nil {
			return err
		}

		if belongs != true {
			return fmt.Errorf("%s belongs to a different asg than %s", node.Name, scaleSet.Id())
		}

		if node.Annotations[cloudprovider.FakeNodeReasonAnnotation] == cloudprovider.FakeNodeUnregistered {
			hasUnregisteredNodes = true
		}
		ref := &azureRef{
			Name: node.Spec.ProviderID,
		}
		refs = append(refs, ref)
	}

	if scaleSet.scaleDownPolicy == cloudprovider.Hibernate {
		return scaleSet.deallocateInstances(refs)
	}
	return scaleSet.DeleteInstances(refs, hasUnregisteredNodes)
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
	template, err := scaleSet.getVMSSFromCache()
	if err != nil {
		return nil, err
	}

	node, err := buildNodeFromTemplate(scaleSet.Name, template, scaleSet.manager, scaleSet.enableDynamicInstanceList)
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

	orchestrationMode, err := scaleSet.getOrchestrationMode()
	if err != nil {
		klog.Errorf("failed to get information for VMSS: %s, error: %v", scaleSet.Name, err)
		return nil, err
	}

	klog.V(4).Infof("VMSS: orchestration Mode %s", orchestrationMode)

	if orchestrationMode == compute.Uniform {
		err := scaleSet.buildScaleSetCache(lastRefresh)
		if err != nil {
			return nil, err
		}

	} else if orchestrationMode == compute.Flexible {
		if scaleSet.manager.config.EnableVmssFlex {
			err := scaleSet.buildScaleSetCacheForFlex(lastRefresh)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, fmt.Errorf("vmss - %q with Flexible orchestration detected but 'enableVmssFlex' feature flag is turned off", scaleSet.Name)
		}

	} else {
		return nil, fmt.Errorf("Failed to determine orchestration mode for vmss %q", scaleSet.Name)
	}

	klog.V(4).Infof("Nodes: returns")
	return scaleSet.instanceCache, nil
}

func (scaleSet *ScaleSet) buildScaleSetCache(lastRefresh time.Time) error {
	klog.V(3).Infof("buildScaleSetCache: resetting instance Cache for scaleSet %s",
		scaleSet.Name)
	vms, rerr := scaleSet.GetScaleSetVms()
	if rerr != nil {
		if isAzureRequestsThrottled(rerr) {
			// Log a warning and update the instance refresh time so that it would retry later.
			// Ensure to retry no sooner than rerr.RetryAfter
			klog.Warningf("buildScaleSetCache: GetScaleSetVms() is throttled with message %v, would return the cached instances", rerr)
			nextRefresh := lastRefresh.Add(scaleSet.instancesRefreshPeriod)
			if nextRefresh.Before(rerr.RetryAfter) {
				delay := rerr.RetryAfter.Sub(nextRefresh)
				lastRefresh = lastRefresh.Add(delay)
			}
			scaleSet.lastInstanceRefresh = lastRefresh
			return nil
		}
		return rerr.Error()
	}

	scaleSet.instanceCache = scaleSet.buildInstanceCache(vms)
	scaleSet.lastInstanceRefresh = lastRefresh

	return nil
}

func (scaleSet *ScaleSet) buildScaleSetCacheForFlex(lastRefresh time.Time) error {
	vms, rerr := scaleSet.GetFlexibleScaleSetVms()
	if rerr != nil {
		if isAzureRequestsThrottled(rerr) {
			// Log a warning and update the instance refresh time so that it would retry after cache expiration
			klog.Warningf("GetFlexibleScaleSetVms() is throttled with message %v, would return the cached instances", rerr)
			scaleSet.lastInstanceRefresh = lastRefresh
			return nil
		}
		return rerr.Error()
	}

	scaleSet.instanceCache = scaleSet.buildInstanceCache(vms)
	scaleSet.lastInstanceRefresh = lastRefresh

	return nil
}

// Note that the GetScaleSetVms() results is not used directly because for the List endpoint,
// their resource ID format is not consistent with Get endpoint
func (scaleSet *ScaleSet) buildInstanceCache(vmList interface{}) []cloudprovider.Instance {
	instances := []cloudprovider.Instance{}

	switch vms := vmList.(type) {
	case []compute.VirtualMachineScaleSetVM:
		for _, vm := range vms {
			powerState := vmPowerStateRunning
			if vm.InstanceView != nil && vm.InstanceView.Statuses != nil {
				powerState = vmPowerStateFromStatuses(*vm.InstanceView.Statuses)
			}
			scaleSet.addInstanceToCache(&instances, vm.ID, vm.ProvisioningState, powerState)
		}
	case []compute.VirtualMachine:
		for _, vm := range vms {
			powerState := vmPowerStateRunning
			if vm.InstanceView != nil && vm.InstanceView.Statuses != nil {
				powerState = vmPowerStateFromStatuses(*vm.InstanceView.Statuses)
			}
			scaleSet.addInstanceToCache(&instances, vm.ID, vm.ProvisioningState, powerState)
		}
	}

	return instances
}

func (scaleSet *ScaleSet) addInstanceToCache(instances *[]cloudprovider.Instance, id *string, provisioningState *string, powerState string) {
	// The resource ID is empty string, which indicates the instance may be in deleting state.
	if len(*id) == 0 {
		return
	}

	resourceID, err := convertResourceGroupNameToLower(*id)
	if err != nil {
		// This shouldn't happen. Log a warning message for tracking.
		klog.Warningf("buildInstanceCache.convertResourceGroupNameToLower failed with error: %v", err)
		return
	}

	*instances = append(*instances, cloudprovider.Instance{
		Id:     "azure://" + resourceID,
		Status: scaleSet.instanceStatusFromProvisioningStateAndPowerState(resourceID, provisioningState, powerState),
	})
}

func (scaleSet *ScaleSet) getInstanceByProviderID(providerID string) (cloudprovider.Instance, bool, error) {
	scaleSet.instanceMutex.Lock()
	defer scaleSet.instanceMutex.Unlock()

	err := scaleSet.validateInstanceCacheWithoutLock()
	if err != nil {
		klog.Errorf("getInstanceByProviderID: error validating instanceCache for providerID %s for scaleSet %s, "+
			"err: %v", providerID, scaleSet.Name, err)
		return cloudprovider.Instance{}, false, err
	}

	for _, instance := range scaleSet.instanceCache {
		if instance.Id == providerID {
			return instance, true, nil
		}
	}
	return cloudprovider.Instance{}, false, nil
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
	scaleSet.lastInstanceRefresh = time.Now()
}

// instanceStatusFromProvisioningStateAndPowerState converts the VM provisioning state and power state to cloudprovider.InstanceStatus
func (scaleSet *ScaleSet) instanceStatusFromProvisioningStateAndPowerState(resourceId string, provisioningState *string, powerState string) *cloudprovider.InstanceStatus {
	// Prefer the proactive cache view of the instance state if we aren't in a terminal state
	// This is because the power state may be taking longer to update and we don't want
	// an unfortunate VM update (TTL 5 min) to reset that state to running.
	if provisioningState == nil || *provisioningState == string(compute.GalleryProvisioningStateUpdating) {
		providerID := azurePrefix + resourceId
		for _, instance := range scaleSet.instanceCache {
			if instance.Id == providerID {
				return instance.Status
			}
		}
		return nil
	}

	klog.V(5).Infof("Getting vm instance provisioning state %s for %s", *provisioningState, resourceId)

	status := &cloudprovider.InstanceStatus{}
	switch *provisioningState {
	case provisioningStateDeleting:
		status.State = cloudprovider.InstanceDeleting
	case provisioningStateCreating:
		status.State = cloudprovider.InstanceCreating
	case provisioningStateFailed:
		// Provisioning can fail both during instance creation or after the instance is running.
		// Per https://learn.microsoft.com/en-us/azure/virtual-machines/states-billing#provisioning-states,
		// ProvisioningState represents the most recent provisioning state, therefore only report
		// InstanceCreating errors when the power state indicates the instance has not yet started running
		if !isRunningVmPowerState(powerState) {
			klog.V(4).Infof("VM %s reports failed provisioning state with non-running power state: %s", resourceId, powerState)
			status.State = cloudprovider.InstanceCreating
			status.ErrorInfo = &cloudprovider.InstanceErrorInfo{
				ErrorClass:   cloudprovider.OutOfResourcesErrorClass,
				ErrorCode:    "provisioning-state-failed",
				ErrorMessage: "Azure failed to provision a node for this node group",
			}
		} else {
			klog.V(5).Infof("VM %s reports a failed provisioning state but is running (%s)", resourceId, powerState)
			status.State = cloudprovider.InstanceRunning
		}
	case provisioningStateSucceeded:
		if powerState == vmPowerStateDeallocated {
			status.State = cloudprovider.InstanceHibernated
		} else if powerState == vmPowerStateDeallocating {
			status.State = cloudprovider.InstanceHibernating
		}
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

func (scaleSet *ScaleSet) invalidateLastSizeRefreshWithLock() {
	scaleSet.sizeMutex.Lock()
	scaleSet.lastSizeRefresh = time.Now().Add(-1 * scaleSet.sizeRefreshPeriod)
	scaleSet.sizeMutex.Unlock()
}

func (scaleSet *ScaleSet) getOrchestrationMode() (compute.OrchestrationMode, error) {
	vmss, err := scaleSet.getVMSSFromCache()
	if err != nil {
		klog.Errorf("failed to get information for VMSS: %s, error: %v", scaleSet.Name, err)
		return "", err
	}
	return vmss.OrchestrationMode, nil
}

func isOperationNotAllowed(rerr *retry.Error) bool {
	return rerr != nil && rerr.ServiceErrorCode() == retry.OperationNotAllowed
}

// deallocateInstances deallocates the given instances. All instances must be controlled by the same nodegroup.
func (scaleSet *ScaleSet) deallocateInstances(instances []*azureRef) error {
	if len(instances) == 0 {
		return nil
	}

	klog.V(3).Infof("Deallocating vmss instances %v", instances)

	commonNg, err := scaleSet.manager.GetNodeGroupForInstance(instances[0])
	if err != nil {
		return err
	}

	instancesToDeallocate := []*azureRef{}
	for _, instance := range instances {
		err = scaleSet.verifyNodeGroup(instance, commonNg.Id())
		if err != nil {
			return err
		}

		// Instances in the "InstanceHibernated" state are the only ones currently being ignored. However, if the
		// deallocation process fails for instances in the "InstanceHibernating" state, we are currently invalidating
		// the cache by calling "waitForDeallocateInstances()" without implementing proper error handling for these cases.
		// Consequently, we do not intend to skip these instances. This approach is simply a conservative measure to
		// ensure that all instances are accounted.
		if cpi, found, err := scaleSet.getInstanceByProviderID(instance.Name); found && err == nil && cpi.Status != nil &&
			cpi.Status.State == cloudprovider.InstanceHibernated {
			klog.V(3).Infof("Skipping deleting instance %s as its current state is deallocated", instance.Name)
			continue
		}
		instancesToDeallocate = append(instancesToDeallocate, instance)
	}

	// nothing to delete
	if len(instancesToDeallocate) == 0 {
		klog.V(3).Infof("No new instances eligible for deallocation, skipping")
		return nil
	}

	instanceIDs := []string{}
	for _, instance := range instancesToDeallocate {
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
	klog.V(3).Infof("Calling virtualMachineScaleSetsClient.DeallocateInstancesAsync(%v) "+
		"for %s", requiredIds.InstanceIds, scaleSet.Name)
	future, rerr := scaleSet.manager.azClient.virtualMachineScaleSetsClient.DeallocateInstancesAsync(ctx, resourceGroup,
		commonNg.Id(), *requiredIds)
	scaleSet.instanceMutex.Unlock()
	if rerr != nil {
		klog.Errorf("virtualMachineScaleSetsClient.DeallocateInstancesAsync for instances %v for %s "+
			"failed: %+v", requiredIds.InstanceIds, scaleSet.Name, rerr)
		return azureToAutoscalerError(rerr)
	}

	// Proactively set the status of the instances to be running in cache as Hibernating. Status will change to Hibernated on success
	for _, instance := range instancesToDeallocate {
		scaleSet.setInstanceStatusByProviderID(instance.Name, cloudprovider.InstanceStatus{State: cloudprovider.InstanceHibernating})
	}

	go scaleSet.waitForDeallocateInstances(future, instancesToDeallocate, requiredIds)
	return nil
}

func (scaleSet *ScaleSet) verifyNodeGroup(instance *azureRef, commonNgID string) error {
	ng, err := scaleSet.manager.GetNodeGroupForInstance(instance)
	if err != nil {
		return err
	}

	if !strings.EqualFold(ng.Id(), commonNgID) {
		return fmt.Errorf("cannot delete instance (%s) which don't belong to the same Scale Set (%q)",
			instance.Name, commonNgID)
	}
	return nil
}

func (scaleSet *ScaleSet) waitForDeallocateInstances(future *azure.Future, instancesToDeallocate []*azureRef,
	requiredIds *compute.VirtualMachineScaleSetVMInstanceRequiredIDs) {
	ctx, cancel := getContextWithTimeout(asyncContextTimeout)
	defer cancel()

	klog.V(3).Infof("Calling virtualMachineScaleSetsClient.WaitForDeallocateInstancesResult(%v) for %s",
		requiredIds.InstanceIds, scaleSet.Name)
	httpResponse, err := scaleSet.manager.azClient.virtualMachineScaleSetsClient.WaitForDeallocateInstancesResult(ctx,
		future, scaleSet.manager.config.ResourceGroup)
	if httpResponse != nil && httpResponse.Body != nil {
		defer httpResponse.Body.Close()
	}
	isSuccess, err := isSuccessHTTPResponse(httpResponse, err)

	if isSuccess {
		klog.V(3).Infof("WaitForDeallocateInstancesResult(%v) for %s success",
			requiredIds.InstanceIds, scaleSet.Name)
		// Set the status of the instances to deallocated only if WaitForDeallocate Call Succeeds
		for _, instance := range instancesToDeallocate {
			scaleSet.setInstanceStatusByProviderID(instance.Name, cloudprovider.InstanceStatus{State: cloudprovider.InstanceHibernated})
		}
		return
	}

	scaleSet.invalidateInstanceCache()
	klog.Errorf("WaitForDeallocateInstancesResult(%v) for %s failed with error: %v", requiredIds.InstanceIds, scaleSet.Name, err)
}

// validateInstanceCacheWithoutLock is used a helper function for validateInstanceCache, get and set methods.
func (scaleSet *ScaleSet) validateInstanceCacheWithoutLock() error {
	if scaleSet.lastInstanceRefresh.Add(scaleSet.instancesRefreshPeriod).After(time.Now()) {
		klog.V(3).Infof("validateInstanceCacheWithoutLock: no need to reset instance Cache for scaleSet %s",
			scaleSet.Name)
		return nil
	}

	splay := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(scaleSet.instancesRefreshJitter + 1)
	lastRefresh := time.Now().Add(-time.Second * time.Duration(splay))
	return scaleSet.buildScaleSetCache(lastRefresh)
}

// getDeallocatedInstances returns list of deallocated instances.
func (scaleSet *ScaleSet) getDeallocatedInstances() ([]cloudprovider.Instance, error) {
	scaleSet.instanceMutex.Lock()
	defer scaleSet.instanceMutex.Unlock()

	err := scaleSet.validateInstanceCacheWithoutLock()
	if err != nil {
		klog.Errorf("getDeallocatedInstances: error validating instanceCache for deallocated instances for scaleSet %s, "+
			"err: %v", scaleSet.Name, err)
		return []cloudprovider.Instance{}, err
	}

	instances := []cloudprovider.Instance{}
	for _, instance := range scaleSet.instanceCache {
		if instance.Status != nil && instance.Status.State == cloudprovider.InstanceHibernated {
			instances = append(instances, instance)
		}
	}
	return instances, nil
}

func (scaleSet *ScaleSet) countDeallocatedInstances() (int, error) {
	scaleSet.instanceMutex.Lock()
	defer scaleSet.instanceMutex.Unlock()

	err := scaleSet.validateInstanceCacheWithoutLock()
	if err != nil {
		klog.Errorf("countDeallocatedInstances: error validating instanceCache for scaleSet %s, "+
			"err: %v", scaleSet.Name, err)
		return 0, err
	}

	var numInstances int
	for _, instance := range scaleSet.instanceCache {
		if instance.Status != nil && instance.Status.State == cloudprovider.InstanceHibernated {
			numInstances++
		}
	}
	return numInstances, nil
}

func (scaleSet *ScaleSet) getSKU() string {
	vmssInfo, err := scaleSet.getVMSSFromCache()
	if err != nil {
		klog.Errorf("Failed to get information for VMSS (%q): %v", scaleSet.Name, err)
		return ""
	}
	return to.String(vmssInfo.Sku.Name)
}

func getCrpOperation(future *azure.Future) string {
	if future != nil {
		resp := future.Response()
		defer resp.Body.Close()
		header := resp.Header["X-Ms-Request-Id"]
		if len(header) > 0 {
			return header[0]
		}
	}
	return ""
}
