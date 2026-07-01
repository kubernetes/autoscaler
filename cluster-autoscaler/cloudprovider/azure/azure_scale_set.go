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
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v6"

	azerrors "github.com/Azure/azure-sdk-for-go-extensions/pkg/errors"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/config/dynamic"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	klog "k8s.io/klog/v2"
	"k8s.io/utils/ptr"
)

var (
	defaultVmssInstancesRefreshPeriod = 5 * time.Minute
	vmssContextTimeout                = 3 * time.Minute
	asyncContextTimeout               = 30 * time.Minute
	vmssSizeMutex                     sync.Mutex
)

const (
	// VMProvisioningStateCreating indicates the VM is being created.
	VMProvisioningStateCreating string = "Creating"
	// VMProvisioningStateDeleting indicates the VM is being deleted.
	VMProvisioningStateDeleting string = "Deleting"
	// VMProvisioningStateFailed indicates the VM provisioning has failed.
	VMProvisioningStateFailed  string = "Failed"
	provisioningStateMigrating string = "Migrating"
	provisioningStateSucceeded string = "Succeeded"
	provisioningStateUpdating  string = "Updating"
	enableFastDeleteOnFailure  bool   = true
	disableFastDeleteOnFailure bool   = false
)

// ScaleSet implements NodeGroup interface.
type ScaleSet struct {
	azureRef
	manager *AzureManager

	minSize int
	maxSize int

	enableForceDelete         bool
	enableDynamicInstanceList bool
	enableDetailedCSEMessage  bool

	// Current Size (Number of VMs)

	// curSize tracks (and caches) the number of VMs in this ScaleSet.
	// It is periodically updated from vmss.Sku.Capacity, with VMSS itself coming
	// either from azure.Cache (which periodically does VMSS.List)
	// or from direct VMSS.Get (always used for Spot).
	curSize int64
	// sizeRefreshPeriod is how often curSize is refreshed from vmss.Sku.Capacity.
	// (Set from azureCache.refreshInterval = VmssCacheTTL or [defaultMetadataCache]refreshInterval = 1min)
	sizeRefreshPeriod time.Duration
	// lastSizeRefresh is the time curSize was last refreshed from vmss.Sku.Capacity.
	// Together with sizeRefreshPeriod, it is used to determine if it is time to refresh curSize.
	lastSizeRefresh time.Time
	// getVmssSizeRefreshPeriod is how often curSize should be refreshed in case VMSS.Get call is used.
	// (Set from GetVmssSizeRefreshPeriod, if specified = get-vmss-size-refresh-period = 30s
	getVmssSizeRefreshPeriod time.Duration
	// sizeMutex protects curSize (the number of VMs in the ScaleSet) from concurrent access
	sizeMutex sync.Mutex

	InstanceCache

	// uses Azure Dedicated Host
	dedicatedHost bool

	enableFastDeleteOnFailedProvisioning bool

	enableLabelPredictionsOnTemplate bool
}

// NewScaleSet creates a new NewScaleSet.
func NewScaleSet(spec *dynamic.NodeGroupSpec, az *AzureManager, curSize int64, dedicatedHost bool) (*ScaleSet, error) {
	scaleSet := &ScaleSet{
		azureRef: azureRef{
			Name: spec.Name,
		},

		minSize: spec.MinSize,
		maxSize: spec.MaxSize,

		manager:           az,
		curSize:           curSize,
		sizeRefreshPeriod: az.azureCache.refreshInterval,
		InstanceCache: InstanceCache{
			instancesRefreshJitter: az.config.VmssVmsCacheJitter,
		},

		enableForceDelete:                az.config.EnableForceDelete,
		enableDynamicInstanceList:        az.config.EnableDynamicInstanceList,
		enableDetailedCSEMessage:         az.config.EnableDetailedCSEMessage,
		enableLabelPredictionsOnTemplate: az.config.EnableLabelPredictionsOnTemplate,
		dedicatedHost:                    dedicatedHost,
	}

	if az.config.VmssVirtualMachinesCacheTTLInSeconds != 0 {
		scaleSet.instancesRefreshPeriod = time.Duration(az.config.VmssVirtualMachinesCacheTTLInSeconds) * time.Second
	} else {
		scaleSet.instancesRefreshPeriod = defaultVmssInstancesRefreshPeriod
	}

	if az.config.GetVmssSizeRefreshPeriod != 0 {
		scaleSet.getVmssSizeRefreshPeriod = time.Duration(az.config.GetVmssSizeRefreshPeriod) * time.Second
	} else {
		scaleSet.getVmssSizeRefreshPeriod = az.azureCache.refreshInterval
	}

	if az.config.EnableDetailedCSEMessage {
		klog.V(2).Infof("enableDetailedCSEMessage: %t", scaleSet.enableDetailedCSEMessage)
	}

	scaleSet.enableFastDeleteOnFailedProvisioning = az.config.EnableFastDeleteOnFailedProvisioning

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
		klog.Errorf("failed to get information for VMSS: %s", scaleSet.Name)
		// Note: We don't return an error here and instead accept defaults.
		// Every invocation of GetOptions() returns an error if this condition is met:
		// `if err != nil && err != cloudprovider.ErrNotImplemented`
		// The error return value is intended to only capture unimplemented.
		return nil, nil
	}
	return scaleSet.manager.GetScaleSetOptions(*template.Name, defaults), nil
}

// MaxSize returns maximum size of the node group.
func (scaleSet *ScaleSet) MaxSize() int {
	return scaleSet.maxSize
}

// getVMSSFromCache returns the live cached VMSS object.
// Callers that read or write mutable fields shared with resize paths,
// especially SKU.Capacity and Etag, must hold vmssSizeMutex.
func (scaleSet *ScaleSet) getVMSSFromCache() (*armcompute.VirtualMachineScaleSet, error) {
	allVMSS := scaleSet.manager.azureCache.getScaleSets()

	if _, exists := allVMSS[scaleSet.Name]; !exists {
		return nil, fmt.Errorf("could not find vmss: %s", scaleSet.Name)
	}

	return allVMSS[scaleSet.Name], nil
}

func (scaleSet *ScaleSet) getCurSize() (int64, *GetVMSSFailedError) {
	scaleSet.sizeMutex.Lock()
	defer scaleSet.sizeMutex.Unlock()

	set, err := scaleSet.getVMSSFromCache()
	if err != nil {
		klog.Errorf("failed to get information for VMSS: %s, error: %v", scaleSet.Name, err)
		return -1, newGetVMSSFailedError(err, true)
	}

	// // Remove check for returning in-memory size when VMSS is in updating state
	// // If VMSS state is updating, return the currentSize which would've been proactively incremented or decremented by CA
	// // unless it's -1. In that case, its better to initialize it.
	// if scaleSet.curSize != -1 && set.VirtualMachineScaleSetProperties != nil &&
	// 	strings.EqualFold(ptr.Deref(set.VirtualMachineScaleSetProperties.ProvisioningState, ""), string(armcompute.GalleryProvisioningStateUpdating)) {
	// 	klog.V(3).Infof("VMSS %q is in updating state, returning cached size: %d", scaleSet.Name, scaleSet.curSize)
	// 	return scaleSet.curSize, nil
	// }

	effectiveSizeRefreshPeriod := scaleSet.sizeRefreshPeriod

	// If the scale set is Spot, we want to have a more fresh view of the Sku.Capacity field.
	// This is because evictions can happen
	// at any given point in time, even before VMs are materialized as
	// nodes. We should be able to react to those and have the autoscaler
	// readjust the goal again to force restoration.
	if isSpot(set) {
		effectiveSizeRefreshPeriod = scaleSet.getVmssSizeRefreshPeriod
	}

	if scaleSet.lastSizeRefresh.Add(effectiveSizeRefreshPeriod).After(time.Now()) {
		klog.V(3).Infof("VMSS: %s, returning in-memory size: %d", scaleSet.Name, scaleSet.curSize)
		return scaleSet.curSize, nil
	}

	// If the scale set is on Spot, make a GET VMSS call to fetch more updated fresh info
	if isSpot(set) {
		ctx, cancel := getContextWithCancel()
		defer cancel()

		var err error
		set, err = scaleSet.manager.azClient.virtualMachineScaleSetsClient.Get(ctx, scaleSet.manager.config.ResourceGroup, scaleSet.Name, nil)
		if err != nil {
			klog.Errorf("failed to get information for VMSS: %s, error: %v", scaleSet.Name, err)
			return -1, newGetVMSSFailedError(err, azerrors.IsNotFoundErr(err))
		}
		// Persist the freshly-fetched VMSS (including its ETag) so subsequent
		// capacity updates send an up-to-date If-Match.
		scaleSet.manager.azureCache.setScaleSet(scaleSet.Name, set)
	}

	vmssSizeMutex.Lock()
	curSize := *set.SKU.Capacity
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

// getScaleSetSize gets Scale Set size.
func (scaleSet *ScaleSet) getScaleSetSize() (int64, error) {
	// First, get the current size of the ScaleSet
	size, getVMSSError := scaleSet.getCurSize()
	if getVMSSError != nil {
		klog.V(3).Infof("getScaleSetSize: error exists (actual err:%v)", getVMSSError.error)
		return size, getVMSSError.error
	}
	if size == -1 {
		err := fmt.Errorf("failed to get scale set size for %s: cached size is -1 without provider error", scaleSet.Name)
		klog.V(3).Infof("getScaleSetSize: size is -1 (actual err:%v)", err)
		return size, err
	}
	return size, nil
}

// setScaleSetSize sets ScaleSet size.
func (scaleSet *ScaleSet) setScaleSetSize(size int64, delta int) error {
	vmssInfo, err := scaleSet.getVMSSFromCache()
	if err != nil {
		klog.Errorf("Failed to get information for VMSS (%q): %v", scaleSet.Name, err)
		return err
	}

	requiredInstances := delta

	// If after reallocating instances we still need more instances or we're just in Delete mode
	// send a scale request
	if requiredInstances > 0 {
		klog.V(3).Infof("Remaining unsatisfied count is %d. Attempting to increase scale set %q "+
			"capacity", requiredInstances, scaleSet.Name)
		err := scaleSet.createOrUpdateInstances(vmssInfo, size)
		if err != nil {
			klog.Errorf("Failed to increase capacity for scale set %q to %d: %v", scaleSet.Name, requiredInstances, err)
			return err
		}
	}
	return nil
}

// TargetSize returns the current TARGET size of the node group. It is possible that the
// number is different from the number of nodes registered in Kubernetes.
func (scaleSet *ScaleSet) TargetSize() (int, error) {
	size, err := scaleSet.getScaleSetSize()
	return int(size), err
}

// canIncreaseSize checks if the size increase is possible.
// It returns the current size of the scale set if the increase is possible, otherwise returns an error.
func (scaleSet *ScaleSet) canIncreaseSize(delta int) (int64, error) {
	if delta <= 0 {
		return -1, fmt.Errorf("size increase must be positive")
	}

	size, err := scaleSet.getScaleSetSize()
	if err != nil {
		return size, err
	}

	// Defensive: getScaleSetSize already errors on size == -1, so this is unreachable today.
	if size == -1 {
		return size, fmt.Errorf("the scale set %s is under initialization, skipping IncreaseSize", scaleSet.Name)
	}

	if int(size)+delta > scaleSet.MaxSize() {
		return size, fmt.Errorf("size increase too large - desired:%d max:%d", int(size)+delta, scaleSet.MaxSize())
	}

	return size, nil
}

// IncreaseSize increases Scale Set size
func (scaleSet *ScaleSet) IncreaseSize(delta int) error {
	size, err := scaleSet.canIncreaseSize(delta)
	if err != nil {
		return err
	}

	return scaleSet.setScaleSetSize(size+int64(delta), delta)
}

// AtomicIncreaseSize increases the VMSS capacity by delta and blocks until the
// VMSS update operation completes. This ensures that the capacity change has been
// fully applied by Azure before returning.
//
// On success, the VMSS capacity has been updated and new VM instances have been
// created (though they may still be booting/joining the cluster). On failure, all
// caches are invalidated so CA fetches fresh state from Azure.
//
// Note: this intentionally deviates from the NodeGroup interface comment that says
// "doesn't wait until the new instances appear" — the blocking behavior is required
// for atomic-scale-up ProvisioningRequest support to provide a capacity guarantee
// before workloads are admitted.
func (scaleSet *ScaleSet) AtomicIncreaseSize(delta int) error {
	size, err := scaleSet.canIncreaseSize(delta)
	if err != nil {
		return err
	}

	newSize := size + int64(delta)

	vmssInfo, err := scaleSet.getVMSSFromCache()
	if err != nil {
		klog.Errorf("Failed to get information for VMSS (%q): %v", scaleSet.Name, err)
		return err
	}

	klog.V(3).Infof("AtomicIncreaseSize: requesting atomic scale-up of %d instances for scale set %q (current size: %d, new size: %d)",
		delta, scaleSet.Name, size, newSize)

	ctx, cancel := getContextWithTimeout(asyncContextTimeout)
	defer cancel()

	effectiveVMSS, poller, err := scaleSet.initCreateOrUpdate(ctx, vmssInfo, newSize)
	if err != nil {
		klog.Errorf("AtomicIncreaseSize: BeginCreateOrUpdate for scale set %q failed: %v", scaleSet.Name, err)
		return err
	}

	var resp armcompute.VirtualMachineScaleSetsClientCreateOrUpdateResponse
	if poller != nil {
		klog.V(3).Infof("AtomicIncreaseSize: waiting for VMSS %q capacity update to complete", scaleSet.Name)
		resp, err = poller.PollUntilDone(ctx, nil)
		scaleSet.invalidateInstanceCache()
		if err != nil {
			klog.Errorf("AtomicIncreaseSize: VMSS %q capacity update failed during polling: %v", scaleSet.Name, err)
			scaleSet.invalidateLastSizeRefreshWithLock()
			scaleSet.manager.invalidateCache()
			return fmt.Errorf("AtomicIncreaseSize: VMSS %q capacity update failed: %w", scaleSet.Name, err)
		}
	}

	// Only update the size cache after the operation has successfully completed.
	// initCreateOrUpdate may have refreshed the cache with a different VMSS object (an
	// ETag retry), so update that object rather than the pre-retry copy.
	scaleSet.sizeMutex.Lock()
	vmssSizeMutex.Lock()
	if poller == nil && effectiveVMSS != vmssInfo {
		// An ETag retry skipped the PUT because the VMSS already met or exceeded the
		// target; publish its actual capacity rather than a stale (possibly lower) target.
		scaleSet.curSize = *effectiveVMSS.SKU.Capacity
	} else {
		effectiveVMSS.SKU.Capacity = &newSize
		// A successful PUT changes the server-side ETag. Adopt the new one returned by
		// the operation so any follow-up PUT before the next cache refresh still carries a
		// valid If-Match rather than overwriting concurrent changes or hitting a 412.
		if scaleSet.manager.config.EnableVMSSEtag && resp.Etag != nil {
			effectiveVMSS.Etag = resp.Etag
		}
		scaleSet.curSize = newSize
	}
	vmssSizeMutex.Unlock()
	scaleSet.lastSizeRefresh = time.Now()
	scaleSet.sizeMutex.Unlock()

	klog.V(3).Infof("AtomicIncreaseSize: VMSS %q capacity update completed successfully (new size: %d)", scaleSet.Name, newSize)
	return nil
}

// GetScaleSetVms returns list of nodes for the given scale set (includes InstanceView for power state).
func (scaleSet *ScaleSet) GetScaleSetVms() ([]*armcompute.VirtualMachineScaleSetVM, error) {
	ctx, cancel := getContextWithTimeout(vmssContextTimeout)
	defer cancel()

	vmList, err := scaleSet.manager.azClient.virtualMachineScaleSetVMsClient.ListVMInstanceView(ctx, scaleSet.manager.config.ResourceGroup,
		scaleSet.Name)

	klog.V(4).Infof("GetScaleSetVms: scaleSet.Name: %s, vmList: %v", scaleSet.Name, vmList)

	if err != nil {
		klog.Errorf("virtualMachineScaleSetVMsClient.ListVMInstanceView failed for %s: %v", scaleSet.Name, err)
		return nil, err
	}

	return vmList, nil
}

// GetFlexibleScaleSetVms returns list of nodes for flexible scale set.
func (scaleSet *ScaleSet) GetFlexibleScaleSetVms() ([]*armcompute.VirtualMachine, error) {
	klog.V(4).Infof("GetFlexibleScaleSetVms: starts")
	ctx, cancel := getContextWithTimeout(vmssContextTimeout)
	defer cancel()

	// Get VMSS info from cache to obtain ID - scaleSet does not store ID directly
	vmssInfo, err := scaleSet.getVMSSFromCache()
	if err != nil {
		klog.Errorf("Failed to get information for VMSS (%q): %v", scaleSet.Name, err)
		return nil, err
	}

	if vmssInfo.ID == nil {
		return nil, fmt.Errorf("VMSS %s has no ID", scaleSet.Name)
	}

	vmList, err := scaleSet.manager.azClient.virtualMachinesClient.ListVmssFlexVMsWithOutInstanceView(ctx, scaleSet.manager.config.ResourceGroup, *vmssInfo.ID)
	if err != nil {
		klog.Errorf("VirtualMachinesClient.ListVmssFlexVMsWithOutInstanceView failed for %s: %v", scaleSet.Name, err)
		return nil, err
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
	_, err := scaleSet.getScaleSetSize()
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

// isPreconditionFailedError reports whether err is an Azure response error with
// HTTP 412 status, i.e. an ETag If-Match precondition failure.
func isPreconditionFailedError(err error) bool {
	r := azerrors.IsResponseError(err)
	return r != nil && r.StatusCode == http.StatusPreconditionFailed
}

// buildScaleSetCapacityUpdate composes the sparse VMSS object sent on a capacity PUT.
// The SKU is copied (not shared) so the cached vmssInfo.SKU.Capacity is not mutated
// before the API call completes.
func buildScaleSetCapacityUpdate(vmssInfo *armcompute.VirtualMachineScaleSet, newSize int64) armcompute.VirtualMachineScaleSet {
	op := armcompute.VirtualMachineScaleSet{
		Name:     vmssInfo.Name,
		Location: vmssInfo.Location,
		SKU: &armcompute.SKU{
			Name:     vmssInfo.SKU.Name,
			Tier:     vmssInfo.SKU.Tier,
			Capacity: ptr.To[int64](newSize),
		},
	}

	if vmssInfo.ExtendedLocation != nil {
		op.ExtendedLocation = &armcompute.ExtendedLocation{
			Name: vmssInfo.ExtendedLocation.Name,
			Type: vmssInfo.ExtendedLocation.Type,
		}

		klog.V(3).Infof("Passing ExtendedLocation information if it is not nil, with Edge Zone name:(%s)", *op.ExtendedLocation.Name)
	}

	return op
}

// initCreateOrUpdate issues the capacity PUT and returns the VMSS object the caller
// should track going forward. On a normal update that is the same vmssInfo; on an ETag
// retry it is the freshly-fetched object that now lives in the cache, so the caller's
// post-completion ETag/capacity updates land on the cached object rather than a stale
// copy. A nil poller with no error means the update was satisfied without a PUT (the
// VMSS already met the target); the caller should adopt the returned object's capacity.
func (scaleSet *ScaleSet) initCreateOrUpdate(ctx context.Context, vmssInfo *armcompute.VirtualMachineScaleSet, newSize int64) (*armcompute.VirtualMachineScaleSet, *runtime.Poller[armcompute.VirtualMachineScaleSetsClientCreateOrUpdateResponse], error) {
	if vmssInfo == nil {
		return nil, nil, fmt.Errorf("vmssInfo cannot be nil while increasing scaleSet capacity")
	}

	scaleSet.sizeMutex.Lock()
	defer scaleSet.sizeMutex.Unlock()

	op := buildScaleSetCapacityUpdate(vmssInfo, newSize)

	klog.V(3).Infof("Calling virtualMachineScaleSetsClient.BeginCreateOrUpdate(%s)", scaleSet.Name)

	opts := &armcompute.VirtualMachineScaleSetsClientBeginCreateOrUpdateOptions{}
	if scaleSet.manager.config.EnableVMSSEtag {
		// Read the cached ETag under vmssSizeMutex, the same lock that guards
		// ETag writes on operation completion, so the read/write pair is
		// race-free even though initCreateOrUpdate holds only sizeMutex.
		vmssSizeMutex.Lock()
		etag := vmssInfo.Etag
		vmssSizeMutex.Unlock()
		opts.IfMatch = etag
	}
	poller, err := scaleSet.manager.azClient.vmssClientForDelete.BeginCreateOrUpdate(ctx, scaleSet.manager.config.ResourceGroup, scaleSet.Name, op, opts)
	if err != nil {
		klog.Errorf("virtualMachineScaleSetsClient.BeginCreateOrUpdate for scale set %q failed: %+v", scaleSet.Name, err)
		if scaleSet.manager.config.EnableVMSSEtag && isPreconditionFailedError(err) {
			// An ETag precondition failure is an optimistic-concurrency conflict, not a
			// real scale-up failure. Refresh the ETag from a fresh GET and retry once so
			// a lost race does not surface as an error that would back off the node group.
			return scaleSet.retryCreateOrUpdateWithFreshETag(ctx, newSize)
		}
		return nil, nil, err
	}
	return vmssInfo, poller, nil
}

// retryCreateOrUpdateWithFreshETag handles an ETag precondition failure by fetching
// the current VMSS, adopting its ETag, and re-issuing the capacity update once. This
// keeps a lost optimistic-concurrency race from surfacing as a scale-up failure, which
// would otherwise back off the node group. It returns the freshly-fetched VMSS (now in
// the cache) so the caller tracks that object instead of the stale pre-retry copy.
// Callers must hold sizeMutex.
func (scaleSet *ScaleSet) retryCreateOrUpdateWithFreshETag(ctx context.Context, newSize int64) (*armcompute.VirtualMachineScaleSet, *runtime.Poller[armcompute.VirtualMachineScaleSetsClientCreateOrUpdateResponse], error) {
	klog.V(2).Infof("VMSS %s update hit ETag precondition; refreshing ETag and retrying once", scaleSet.Name)

	fresh, err := scaleSet.manager.azClient.virtualMachineScaleSetsClient.Get(ctx, scaleSet.manager.config.ResourceGroup, scaleSet.Name, nil)
	if err != nil {
		klog.Errorf("VMSS %s ETag retry: failed to GET current scale set: %v", scaleSet.Name, err)
		scaleSet.invalidateForPreconditionFailure()
		return nil, nil, err
	}

	// Persist the freshly-fetched VMSS (ETag and capacity) for subsequent operations.
	scaleSet.manager.azureCache.setScaleSet(scaleSet.Name, fresh)

	// If another writer already grew the VMSS to at least our target, the desired floor
	// is already met; don't issue a PUT that would shrink it back down. Return the fresh
	// object (nil poller) so the caller publishes its actual capacity, not a stale target.
	if fresh.SKU != nil && fresh.SKU.Capacity != nil && *fresh.SKU.Capacity >= newSize {
		klog.V(2).Infof("VMSS %s already at capacity %d (>= desired %d) after refresh; skipping retry", scaleSet.Name, *fresh.SKU.Capacity, newSize)
		return fresh, nil, nil
	}

	// Rebuild the update request from the freshly-fetched VMSS so the retried PUT carries
	// current SKU/location details rather than the pre-conflict snapshot.
	op := buildScaleSetCapacityUpdate(fresh, newSize)
	opts := &armcompute.VirtualMachineScaleSetsClientBeginCreateOrUpdateOptions{IfMatch: fresh.Etag}
	poller, err := scaleSet.manager.azClient.vmssClientForDelete.BeginCreateOrUpdate(ctx, scaleSet.manager.config.ResourceGroup, scaleSet.Name, op, opts)
	if err != nil {
		klog.Errorf("VMSS %s ETag retry: BeginCreateOrUpdate failed: %+v", scaleSet.Name, err)
		scaleSet.invalidateForPreconditionFailure()
		return nil, nil, err
	}
	return fresh, poller, nil
}

// invalidateForPreconditionFailure invalidates the instance, size, and manager caches
// so the next loop re-plans from fresh Azure state. Callers must hold sizeMutex.
func (scaleSet *ScaleSet) invalidateForPreconditionFailure() {
	scaleSet.invalidateInstanceCache()
	// Already holding sizeMutex here, so force the size refresh without re-locking.
	scaleSet.invalidateLastSizeRefresh()
	scaleSet.manager.invalidateCache()
}

func (scaleSet *ScaleSet) createOrUpdateInstances(vmssInfo *armcompute.VirtualMachineScaleSet, newSize int64) error {
	ctx, cancel := getContextWithTimeout(vmssContextTimeout)
	defer cancel()
	// For non-atomic scale up, we eagerly update the VMSS size in the cache
	// to avoid overshooting the max size if multiple scale up requests are made concurrently.
	// This preserves the existing behavior (before atomic scale up was added).
	vmssSizeMutex.Lock()
	previousSize := vmssInfo.SKU.Capacity
	vmssInfo.SKU.Capacity = &newSize
	vmssSizeMutex.Unlock()
	effectiveVMSS, poller, err := scaleSet.initCreateOrUpdate(ctx, vmssInfo, newSize)
	if err != nil {
		// The update was not accepted (e.g. an ETag precondition failure), so roll
		// back the eager capacity mutation. Otherwise the rejected desired size would
		// remain on the cached VMSS object and become visible via getCurSize before
		// the next full cache refresh.
		vmssSizeMutex.Lock()
		vmssInfo.SKU.Capacity = previousSize
		vmssSizeMutex.Unlock()
		return err
	}

	// initCreateOrUpdate may have refreshed the cache with a different VMSS object (an
	// ETag retry). Track that object so the async completion below updates the cached
	// copy, and decide the size to publish from it.
	publishedSize := newSize
	vmssSizeMutex.Lock()
	if poller != nil {
		// A PUT is in flight; carry the eager capacity mutation onto whatever object
		// the async completion will update so the cache stays consistent meanwhile.
		effectiveVMSS.SKU.Capacity = &newSize
	} else if effectiveVMSS != vmssInfo {
		// An ETag retry skipped the PUT because the VMSS already met or exceeded the
		// target; publish its actual capacity rather than a stale (possibly lower) target.
		publishedSize = *effectiveVMSS.SKU.Capacity
	}
	vmssSizeMutex.Unlock()

	// Proactively set the VMSS size so autoscaler makes better decisions.
	// initCreateOrUpdate releases sizeMutex before returning, so reacquire it
	// here to keep curSize / lastSizeRefresh updates serialized with readers
	// (e.g. getCurSize).
	scaleSet.sizeMutex.Lock()
	scaleSet.curSize = publishedSize
	scaleSet.lastSizeRefresh = time.Now()
	scaleSet.sizeMutex.Unlock()

	// Poll for completion asynchronously to avoid blocking the autoscaler
	if poller != nil {
		go scaleSet.waitForCreateOrUpdateInstances(poller, effectiveVMSS)
	}
	return nil
}

// waitForCreateOrUpdateInstances waits for the outcome of VMSS capacity update initiated via BeginCreateOrUpdate.
func (scaleSet *ScaleSet) waitForCreateOrUpdateInstances(poller *runtime.Poller[armcompute.VirtualMachineScaleSetsClientCreateOrUpdateResponse], vmssInfo *armcompute.VirtualMachineScaleSet) {
	ctx, cancel := getContextWithTimeout(asyncContextTimeout)
	defer cancel()

	klog.V(3).Infof("Calling PollUntilDone for CreateOrUpdate(%s)", scaleSet.Name)
	resp, err := poller.PollUntilDone(ctx, nil)

	// Invalidate instanceCache on success and failure. Failure might have created a few instances, but it is very rare.
	scaleSet.invalidateInstanceCache()

	if err != nil {
		klog.Errorf("Failed to update the capacity for vmss %s with error %v, invalidate the cache so as to get the real size from API", scaleSet.Name, err)
		// Invalidate the VMSS size cache in order to fetch the size from the API.
		scaleSet.invalidateLastSizeRefreshWithLock()
		scaleSet.manager.invalidateCache()
		return
	}

	// A successful PUT changes the server-side ETag. Adopt the new one returned by
	// the operation so any follow-up PUT before the next cache refresh still carries a
	// valid If-Match rather than overwriting concurrent changes or hitting a 412.
	if scaleSet.manager.config.EnableVMSSEtag {
		if resp.Etag != nil {
			vmssSizeMutex.Lock()
			vmssInfo.Etag = resp.Etag
			vmssSizeMutex.Unlock()
		} else {
			// The PUT succeeded but the response carried no ETag, so the cached one
			// is now stale. Mark the cache for refresh so the next operation fetches
			// a current ETag via GET instead of sending a stale If-Match (which would
			// otherwise be rejected with a 412 and force an extra re-plan).
			scaleSet.invalidateLastSizeRefreshWithLock()
			scaleSet.manager.invalidateCache()
		}
	}

	klog.V(3).Infof("PollUntilDone for CreateOrUpdate(%s) success", scaleSet.Name)
}

// DeleteInstances deletes the given instances. All instances must be controlled by the same nodegroup.
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
		err = scaleSet.verifyNodeGroup(instance, commonAsg.Id())
		if err != nil {
			return err
		}

		if cpi, found, err := scaleSet.getInstanceByProviderID(instance.Name); found && err == nil && cpi.Status != nil &&
			cpi.Status.State == cloudprovider.InstanceDeleting {
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

	// Convert []string to []*string
	instanceIDPtrs := make([]*string, len(instanceIDs))
	for i := range instanceIDs {
		instanceIDPtrs[i] = &instanceIDs[i]
	}

	requiredIds := &armcompute.VirtualMachineScaleSetVMInstanceRequiredIDs{
		InstanceIDs: instanceIDPtrs,
	}

	ctx, cancel := getContextWithTimeout(vmssContextTimeout)
	defer cancel()

	poller, err := scaleSet.deleteInstances(ctx, requiredIds, commonAsg.Id())
	if err != nil {
		klog.Errorf("virtualMachineScaleSetsClient.DeleteInstancesAsync for instances %v for %s failed: %+v", requiredIds.InstanceIDs, scaleSet.Name, err)
		return err
	}

	if !scaleSet.manager.config.StrictCacheUpdates {
		// Proactively decrement scale set size so that we don't
		// go below minimum node count if cache data is stale.
		// If the cached size can't represent the delete batch,
		// mark it stale instead of publishing a negative size.
		// Only do it for non-unregistered nodes.

		if !hasUnregisteredNodes {
			deleteCount := int64(len(instanceIDs))
			scaleSet.sizeMutex.Lock()
			if scaleSet.curSize < deleteCount {
				klog.Warningf("VMSS: %s, cached size %d is smaller than instances to delete %d, invalidating size cache instead of decrementing", scaleSet.Name, scaleSet.curSize, deleteCount)
				scaleSet.lastSizeRefresh = time.Time{}
			} else {
				scaleSet.curSize -= deleteCount
				scaleSet.lastSizeRefresh = time.Now()
			}
			scaleSet.sizeMutex.Unlock()
		}

		// Proactively set the status of the instances to be deleted in cache
		for _, instance := range instancesToDelete {
			scaleSet.setInstanceStatusByProviderID(instance.Name, cloudprovider.InstanceStatus{State: cloudprovider.InstanceDeleting})
			scaleSet.manager.azureCache.setInstanceStateByProviderID(instance.Name, cloudprovider.InstanceDeleting)
		}
	}

	if poller != nil {
		go scaleSet.waitForDeleteInstances(poller, requiredIds)
	}
	return nil
}

func (scaleSet *ScaleSet) waitForDeleteInstances(poller *runtime.Poller[armcompute.VirtualMachineScaleSetsClientDeleteInstancesResponse], requiredIds *armcompute.VirtualMachineScaleSetVMInstanceRequiredIDs) {
	ctx, cancel := getContextWithTimeout(asyncContextTimeout)
	defer cancel()

	klog.V(3).Infof("Calling PollUntilDone for DeleteInstances(%v) for %s", requiredIds.InstanceIDs, scaleSet.Name)
	_, err := poller.PollUntilDone(ctx, nil)
	if err == nil {
		klog.V(3).Infof("PollUntilDone for DeleteInstances(%v) for %s success", requiredIds.InstanceIDs, scaleSet.Name)
		if scaleSet.manager.config.StrictCacheUpdates {
			if refreshErr := scaleSet.manager.forceRefresh(); refreshErr != nil {
				klog.Errorf("forceRefresh failed after successful DeleteInstances(%v) for %s: %v", requiredIds.InstanceIDs, scaleSet.Name, refreshErr)
				scaleSet.manager.invalidateCache()
			}
			scaleSet.invalidateInstanceCache()
		}
		return
	}

	// Retry once on OperationPreempted: this CRP error means a concurrent VMSS
	// mutation (e.g., scale-up, update, another delete) superseded our delete.
	// A single retry is statistically sufficient — two consecutive preemptions
	// would require three operations racing on the same VMSS, which is unlikely
	// in CAS's scale-down flow. This mirrors the pattern from the legacy track1
	// vmssvmclient.updateVMSSVMs().
	if isOperationPreempted(err) {
		klog.V(2).Infof("PollUntilDone for DeleteInstances(%v) for %s was preempted, retrying once", requiredIds.InstanceIDs, scaleSet.Name)
		retryCtx, retryCancel := getContextWithTimeout(vmssContextTimeout)
		retryPoller, retryErr := scaleSet.deleteInstances(retryCtx, requiredIds, scaleSet.Name)
		retryCancel()
		if retryErr == nil && retryPoller != nil {
			_, retryErr = retryPoller.PollUntilDone(ctx, nil)
		}
		if retryErr == nil {
			klog.V(3).Infof("PollUntilDone for DeleteInstances(%v) for %s retry success", requiredIds.InstanceIDs, scaleSet.Name)
			scaleSet.invalidateInstanceCache()
			return
		}
		klog.Errorf("PollUntilDone for DeleteInstances(%v) for %s retry failed: %v", requiredIds.InstanceIDs, scaleSet.Name, retryErr)
	}

	scaleSet.invalidateInstanceCache()
	scaleSet.invalidateLastSizeRefreshWithLock()
	if refreshErr := scaleSet.manager.forceRefresh(); refreshErr != nil {
		klog.Errorf("forceRefresh failed after DeleteInstances(%v) for %s returned error: %v", requiredIds.InstanceIDs, scaleSet.Name, refreshErr)
		scaleSet.manager.invalidateCache()
	}
	klog.Errorf("PollUntilDone for DeleteInstances(%v) for %s failed with error: %v", requiredIds.InstanceIDs, scaleSet.Name, err)
}

// DeleteNodes deletes the nodes from the group.
func (scaleSet *ScaleSet) DeleteNodes(nodes []*apiv1.Node) error {
	klog.V(3).Infof("Delete nodes requested: %q\n", nodes)
	size, err := scaleSet.getScaleSetSize()
	if err != nil {
		return err
	}

	// This only catches callers already at min size. A future change should also reject
	// batches that would go below min size; create-error cleanup fallback is the only
	// currently known path that can reach this check without regular scale-down filtering.
	if int(size) <= scaleSet.MinSize() {
		return fmt.Errorf("min size reached, nodes will not be deleted")
	}
	return scaleSet.ForceDeleteNodes(nodes)
}

// ForceDeleteNodes deletes nodes from the group regardless of constraints.
func (scaleSet *ScaleSet) ForceDeleteNodes(nodes []*apiv1.Node) error {
	klog.V(3).Infof("Delete nodes requested: %q\n", nodes)
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
func (scaleSet *ScaleSet) TemplateNodeInfo() (*framework.NodeInfo, error) {
	vmss, err := scaleSet.getVMSSFromCache()
	if err != nil {
		return nil, err
	}

	inputLabels := map[string]string{}
	inputTaints := ""
	template, err := buildNodeTemplateFromVMSS(vmss, inputLabels, inputTaints)
	if err != nil {
		return nil, err
	}
	node, err := buildNodeFromTemplate(scaleSet.Name, template, scaleSet.manager, scaleSet.enableDynamicInstanceList, scaleSet.enableLabelPredictionsOnTemplate)
	if err != nil {
		return nil, err
	}

	nodeInfo := framework.NewNodeInfo(node, nil, framework.NewPodInfo(cloudprovider.BuildKubeProxy(scaleSet.Name), nil))
	return nodeInfo, nil
}

// Nodes returns a list of all nodes that belong to this node group.
func (scaleSet *ScaleSet) Nodes() ([]cloudprovider.Instance, error) {
	curSize, getVMSSError := scaleSet.getCurSize()
	if getVMSSError != nil {
		klog.Errorf("Failed to get current size for vmss %q: %v", scaleSet.Name, getVMSSError.error)
		if getVMSSError.notFound {
			return []cloudprovider.Instance{}, nil // Don't return error if VMSS not found
		}
		return nil, getVMSSError.error // We want to return error if other errors occur.
	}

	scaleSet.instanceMutex.Lock()
	defer scaleSet.instanceMutex.Unlock()

	if int64(len(scaleSet.instanceCache)) == curSize &&
		scaleSet.lastInstanceRefresh.Add(scaleSet.instancesRefreshPeriod).After(time.Now()) {
		klog.V(4).Infof("Nodes: returns with curSize %d", curSize)
		return scaleSet.instanceCache, nil
	}

	// Forcefully updating the instanceCache as the instanceCacheSize didn't match curSize or cache is invalid.
	err := scaleSet.updateInstanceCache()
	if err != nil {
		return nil, err
	}

	klog.V(4).Infof("Nodes: returns")
	return scaleSet.instanceCache, nil
}

// buildScaleSetCacheForFlex is used by orchestrationMode == armcompute.OrchestrationModeFlexible
func (scaleSet *ScaleSet) buildScaleSetCacheForFlex() error {
	klog.V(3).Infof("buildScaleSetCacheForFlex: resetting instance Cache for scaleSet %s",
		scaleSet.Name)
	splay := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(scaleSet.instancesRefreshJitter + 1)
	lastRefresh := time.Now().Add(-time.Second * time.Duration(splay))

	vms, err := scaleSet.GetFlexibleScaleSetVms()
	if err != nil {
		if throttled, retryAfter := isAzureRequestsThrottled(err); throttled {
			// Log a warning and update the instance refresh time so that it would retry after cache expiration
			// Ensure to retry no sooner than retryAfter
			klog.Warningf("buildScaleSetCacheForFlex: GetFlexibleScaleSetVms() is throttled, would return the cached instances")
			if retryAfter > 0 {
				nextRefresh := lastRefresh.Add(scaleSet.instancesRefreshPeriod)
				retryTime := time.Now().Add(retryAfter)
				if nextRefresh.Before(retryTime) {
					delay := retryTime.Sub(nextRefresh)
					lastRefresh = lastRefresh.Add(delay)
				}
			}
			scaleSet.lastInstanceRefresh = lastRefresh
			return nil
		}
		klog.Errorf("buildScaleSetCacheForFlex: GetFlexibleScaleSetVms() failed with error: %v", err)
		return err
	}

	scaleSet.instanceCache = buildInstanceCacheForFlex(vms, scaleSet.enableFastDeleteOnFailedProvisioning)
	scaleSet.lastInstanceRefresh = lastRefresh

	return nil
}

func (scaleSet *ScaleSet) buildScaleSetCacheForUniform() error {
	klog.V(3).Infof("updateInstanceCache: resetting instance Cache for scaleSet %s",
		scaleSet.Name)
	splay := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(scaleSet.instancesRefreshJitter + 1)
	lastRefresh := time.Now().Add(-time.Second * time.Duration(splay))
	vms, err := scaleSet.GetScaleSetVms()
	if err != nil {
		if throttled, retryAfter := isAzureRequestsThrottled(err); throttled {
			// Log a warning and update the instance refresh time so that it would retry later.
			// Ensure to retry no sooner than retryAfter
			klog.Warningf("updateInstanceCache: GetScaleSetVms() is throttled, would return the cached instances")
			if retryAfter > 0 {
				nextRefresh := lastRefresh.Add(scaleSet.instancesRefreshPeriod)
				retryTime := time.Now().Add(retryAfter)
				if nextRefresh.Before(retryTime) {
					delay := retryTime.Sub(nextRefresh)
					lastRefresh = lastRefresh.Add(delay)
				}
			}
			scaleSet.lastInstanceRefresh = lastRefresh
			return nil
		}
		klog.Errorf("updateInstanceCache: GetScaleSetVms() failed with error: %v", err)
		return err
	}

	instances := []cloudprovider.Instance{}
	// Note that the GetScaleSetVms() results is not used directly because for the List endpoint,
	// their resource ID format is not consistent with Get endpoint
	for i := range vms {
		// The resource ID is empty string, which indicates the instance may be in deleting state.
		if *vms[i].ID == "" {
			continue
		}
		resourceID, err := convertResourceGroupNameToLower(*vms[i].ID)
		if err != nil {
			// This shouldn't happen. Log a warning message for tracking.
			klog.Warningf("updateInstanceCache: buildInstanceCache.convertResourceGroupNameToLower failed with error: %v", err)
			continue
		}

		instances = append(instances, cloudprovider.Instance{
			Id:     azurePrefix + resourceID,
			Status: scaleSet.instanceStatusFromVM(vms[i]),
		})
	}

	scaleSet.instanceCache = instances
	scaleSet.lastInstanceRefresh = lastRefresh

	return nil
}

// Note that the GetScaleSetVms() results is not used directly because for the List endpoint,
// their resource ID format is not consistent with Get endpoint
// buildInstanceCacheForFlex used by orchestrationMode == armcompute.OrchestrationModeFlexible
func buildInstanceCacheForFlex(vms []*armcompute.VirtualMachine, enableFastDeleteOnFailedProvisioning bool) []cloudprovider.Instance {
	var instances []cloudprovider.Instance
	for _, vm := range vms {
		powerState := vmPowerStateRunning
		if vm.Properties != nil && vm.Properties.InstanceView != nil && vm.Properties.InstanceView.Statuses != nil {
			powerState = vmPowerStateFromStatuses(vm.Properties.InstanceView.Statuses)
		}
		var provisioningState *string
		if vm.Properties != nil && vm.Properties.ProvisioningState != nil {
			provisioningState = vm.Properties.ProvisioningState
		}
		addVMToCache(&instances, vm.ID, provisioningState, powerState, enableFastDeleteOnFailedProvisioning)
	}

	return instances
}

// addVMToCache used by orchestrationMode == armcompute.OrchestrationModeFlexible
func addVMToCache(instances *[]cloudprovider.Instance, id, provisioningState *string, powerState string, enableFastDeleteOnFailedProvisioning bool) {
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
		Id:     azurePrefix + resourceID,
		Status: instanceStatusFromProvisioningStateAndPowerState(resourceID, provisioningState, powerState, enableFastDeleteOnFailedProvisioning),
	})
}

// instanceStatusFromProvisioningStateAndPowerState converts the VM provisioning state to cloudprovider.InstanceStatus
// instanceStatusFromProvisioningStateAndPowerState used by orchestrationMode == compute.Flexible
// Suggestion: reunify this with scaleSet.instanceStatusFromVM()
func instanceStatusFromProvisioningStateAndPowerState(resourceID string, provisioningState *string, powerState string, enableFastDeleteOnFailedProvisioning bool) *cloudprovider.InstanceStatus {
	if provisioningState == nil {
		return nil
	}

	klog.V(5).Infof("Getting vm instance provisioning state %s for %s", *provisioningState, resourceID)

	status := &cloudprovider.InstanceStatus{}
	switch *provisioningState {
	case VMProvisioningStateDeleting:
		status.State = cloudprovider.InstanceDeleting
	case VMProvisioningStateCreating:
		status.State = cloudprovider.InstanceCreating
	case VMProvisioningStateFailed:
		status.State = cloudprovider.InstanceRunning

		if enableFastDeleteOnFailedProvisioning {
			// Provisioning can fail both during instance creation or after the instance is running.
			// Per https://learn.microsoft.com/en-us/azure/virtual-machines/states-billing#provisioning-states,
			// ProvisioningState represents the most recent provisioning state, therefore only report
			// InstanceCreating errors when the power state indicates the instance has not yet started running
			if !isRunningVmPowerState(powerState) {
				klog.V(4).Infof("VM %s reports failed provisioning state with non-running power state: %s", resourceID, powerState)
				status.State = cloudprovider.InstanceCreating
				status.ErrorInfo = &cloudprovider.InstanceErrorInfo{
					ErrorClass:   cloudprovider.OutOfResourcesErrorClass,
					ErrorCode:    "provisioning-state-failed",
					ErrorMessage: "Azure failed to provision a node for this node group",
				}
			} else {
				klog.V(5).Infof("VM %s reports a failed provisioning state but is running (%s)", resourceID, powerState)
				status.State = cloudprovider.InstanceRunning
			}
		}
	default:
		status.State = cloudprovider.InstanceRunning
	}

	return status
}

func isSpot(vmss *armcompute.VirtualMachineScaleSet) bool {
	return vmss != nil && vmss.Properties != nil &&
		vmss.Properties.VirtualMachineProfile != nil &&
		vmss.Properties.VirtualMachineProfile.Priority != nil &&
		*vmss.Properties.VirtualMachineProfile.Priority == armcompute.VirtualMachinePriorityTypesSpot
}

func (scaleSet *ScaleSet) invalidateLastSizeRefreshWithLock() {
	scaleSet.sizeMutex.Lock()
	scaleSet.invalidateLastSizeRefresh()
	scaleSet.sizeMutex.Unlock()
}

// invalidateLastSizeRefresh forces the next getCurSize call to refresh curSize
// from the VMSS. Callers must already hold sizeMutex.
func (scaleSet *ScaleSet) invalidateLastSizeRefresh() {
	scaleSet.lastSizeRefresh = time.Now().Add(-1 * scaleSet.sizeRefreshPeriod)
}

func (scaleSet *ScaleSet) getOrchestrationMode() (*armcompute.OrchestrationMode, error) {
	vmss, err := scaleSet.getVMSSFromCache()
	if err != nil {
		klog.Errorf("failed to get information for VMSS: %s, error: %v", scaleSet.Name, err)
		return nil, err
	}
	if vmss.Properties == nil {
		return nil, nil
	}
	return vmss.Properties.OrchestrationMode, nil
}

func (scaleSet *ScaleSet) cseErrors(extensions []*armcompute.VirtualMachineExtensionInstanceView) ([]string, bool) {
	var errs []string
	failed := false
	if extensions != nil {
		for _, extension := range extensions {
			if strings.EqualFold(ptr.Deref(extension.Name, ""), vmssCSEExtensionName) && extension.Statuses != nil {
				for _, status := range extension.Statuses {
					if status.Level != nil && *status.Level == armcompute.StatusLevelTypesError {
						errs = append(errs, ptr.Deref(status.Message, ""))
						failed = true
					}
				}
			}
		}
	}
	return errs, failed
}

func (scaleSet *ScaleSet) getSKU() string {
	vmssInfo, err := scaleSet.getVMSSFromCache()
	if err != nil {
		klog.Errorf("Failed to get information for VMSS (%q): %v", scaleSet.Name, err)
		return ""
	}
	return ptr.Deref(vmssInfo.SKU.Name, "")
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

// GetVMSSFailedError is used to differentiate between
// NotFound and other errors
type GetVMSSFailedError struct {
	notFound bool
	error    error
}

func newGetVMSSFailedError(error error, notFound bool) *GetVMSSFailedError {
	return &GetVMSSFailedError{
		error:    error,
		notFound: notFound,
	}
}

func (v *GetVMSSFailedError) Error() string {
	return v.error.Error()
}
