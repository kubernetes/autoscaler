/*
Copyright 2023 The Kubernetes Authors.

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
	"sync"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2022-08-01/compute"
	"github.com/Azure/go-autorest/autorest/to"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/klog/v2"
)

/*
- "instanceCache" is included in the scaleSet data structures and holds
status information of the instances / vms. This data is used by the CAS
to make scaleUp / scaleDown decisions based on what is the current state
the cluster without making an api call.
- The time for this cache is represented by "instancesRefreshPeriod" which
by default is defaultVmssInstancesRefreshPeriod ~ 5 mins.
- "lastInstanceRefresh" represents the time when the cache was validated
the last time.
- Following methods are defined related to the instanceCache:
    - invalidateInstanceCache()
    - validateInstanceCache()
    - validateInstanceCacheWithoutLock()
    - updateInstanceCache()
    - getInstanceByProviderID()
    - getInstancesByState()
    - getInstanceCacheSize()
    - setInstanceStatusByProviderID()
    - setInstanceStatusByProviderID()
*/

// InstanceCache tracks the VMs in the ScaleSet, in the form of corresponding cloudprovider.Instances.
// This struct also contains related locks and cache interval variables.
type InstanceCache struct {
	// instanceCache tracks the VMs in the ScaleSet, in the form of corresponding cloudprovider.Instances.
	// instanceCache directly backs the efficient response to NodeGroup.Nodes(), implemented by ScaleSet.Nodes().
	// It is periodially updated from VMSS using virtualMachineScaleSetVMsClient.List().
	instanceCache []cloudprovider.Instance
	// instancesRefreshPeriod is how often instance cache is refreshed from VMSS.
	// (Set from VmssVmsCacheTTL or defaultVmssInstancesRefreshPeriod = 5min)
	instancesRefreshPeriod time.Duration
	// lastInstanceRefresh is the time instanceCache was last refreshed from VMSS.
	// Together with instancesRefreshPeriod, it is used to determine if it is time to refresh instanceCache.
	lastInstanceRefresh time.Time
	// instancesRefreshJitter (in seconds) is used to ensure refreshes (which involve expensive List call)
	// don't happen at exactly the same time on all ScaleSets
	instancesRefreshJitter int
	// instanceMutex is used for protecting instance cache from concurrent access
	instanceMutex sync.Mutex
}

// invalidateInstanceCache invalidates the instanceCache by modifying the lastInstanceRefresh.
func (scaleSet *ScaleSet) invalidateInstanceCache() {
	scaleSet.instanceMutex.Lock()
	defer scaleSet.instanceMutex.Unlock()
	// Set the instanceCache as outdated.
	klog.V(3).Infof("invalidating instanceCache for %s", scaleSet.Name)
	scaleSet.lastInstanceRefresh = time.Now().Add(-1 * scaleSet.instancesRefreshPeriod)
}

// validateInstanceCache updates the instanceCache if it has expired. It acquires lock.
func (scaleSet *ScaleSet) validateInstanceCache() error {
	scaleSet.instanceMutex.Lock()
	defer scaleSet.instanceMutex.Unlock()
	return scaleSet.validateInstanceCacheWithoutLock()
}

// validateInstanceCacheWithoutLock is used a helper function for validateInstanceCache, get and set methods.
func (scaleSet *ScaleSet) validateInstanceCacheWithoutLock() error {
	if scaleSet.lastInstanceRefresh.Add(scaleSet.instancesRefreshPeriod).After(time.Now()) {
		klog.V(3).Infof("validateInstanceCacheWithoutLock: no need to reset instance Cache for scaleSet %s",
			scaleSet.Name)
		return nil
	}

	return scaleSet.updateInstanceCache()
}

// updateInstanceCache forcefully updates the cache without checking the timer - lastInstanceRefresh.
// Caller is responsible for acquiring lock on the instanceCache.
func (scaleSet *ScaleSet) updateInstanceCache() error {
	orchestrationMode, err := scaleSet.getOrchestrationMode()
	if err != nil {
		klog.Errorf("failed to get information for VMSS: %s, error: %v", scaleSet.Name, err)
		return err
	}

	if orchestrationMode == compute.Flexible {
		if scaleSet.manager.config.EnableVmssFlexNodes {
			return scaleSet.buildScaleSetCacheForFlex()
		}
		return fmt.Errorf("vmss - %q with Flexible orchestration detected but 'enableVmssFlexNodes' feature flag is turned off", scaleSet.Name)
	} else if orchestrationMode == compute.Uniform {
		return scaleSet.buildScaleSetCacheForUniform()
	}

	return fmt.Errorf("failed to determine orchestration mode for vmss %q", scaleSet.Name)
}

// getInstanceByProviderID returns instance from instanceCache if given providerID exists.
func (scaleSet *ScaleSet) getInstanceByProviderID(providerID string) (cloudprovider.Instance, bool, error) {
	scaleSet.instanceMutex.Lock()
	defer scaleSet.instanceMutex.Unlock()

	err := scaleSet.validateInstanceCacheWithoutLock()
	if err != nil {
		klog.Errorf("getInstanceByProviderID: error validating instanceCache for providerID %s for scaleSet %s, err: %v",
			providerID, scaleSet.Name, err)
		return cloudprovider.Instance{}, false, err
	}

	for _, instance := range scaleSet.instanceCache {
		if instance.Id == providerID {
			return instance, true, nil
		}
	}
	return cloudprovider.Instance{}, false, nil
}

// getInstancesByState returns list of instances with given state.
func (scaleSet *ScaleSet) getInstancesByState(state cloudprovider.InstanceState) ([]cloudprovider.Instance, error) {
	scaleSet.instanceMutex.Lock()
	defer scaleSet.instanceMutex.Unlock()

	err := scaleSet.validateInstanceCacheWithoutLock()
	if err != nil {
		klog.Errorf("getInstancesByState: error validating instanceCache for state %d for scaleSet %s, "+
			"err: %v", state, scaleSet.Name, err)
		return []cloudprovider.Instance{}, err
	}

	instances := []cloudprovider.Instance{}
	for _, instance := range scaleSet.instanceCache {
		if instance.Status != nil && instance.Status.State == state {
			instances = append(instances, instance)
		}
	}
	return instances, nil
}

// getInstanceCacheSize returns the size of the instanceCache.
func (scaleSet *ScaleSet) getInstanceCacheSize() (int64, error) {
	scaleSet.instanceMutex.Lock()
	defer scaleSet.instanceMutex.Unlock()

	err := scaleSet.validateInstanceCacheWithoutLock()
	if err != nil {
		klog.Errorf("getInstanceCacheSize: error validating instanceCache for scaleSet: %s, "+
			"err: %v", scaleSet.Name, err)
		return -1, err
	}

	return int64(len(scaleSet.instanceCache)), nil
}

// setInstanceStatusByProviderID sets the status for an instance with given providerID.
// It reset the cache if stale and sets the status by acquiring a lock.
func (scaleSet *ScaleSet) setInstanceStatusByProviderID(providerID string, status cloudprovider.InstanceStatus) {
	scaleSet.instanceMutex.Lock()
	defer scaleSet.instanceMutex.Unlock()

	err := scaleSet.validateInstanceCacheWithoutLock()
	if err != nil {
		klog.Errorf("setInstanceStatusByProviderID: error validating instanceCache for providerID %s for "+
			"scaleSet: %s, err: %v", providerID, scaleSet.Name, err)
		// return no error because CAS runs with the expectation that future runs will refresh instance Cache
	}

	for k, instance := range scaleSet.instanceCache {
		if instance.Id == providerID {
			klog.V(3).Infof("setInstanceStatusByProviderID: setting instance state for %s for scaleSet "+
				"%s to %d", instance.Id, scaleSet.Name, status.State)
			scaleSet.instanceCache[k].Status = &status
			break
		}
	}
}

// instanceStatusFromVM converts the VM provisioning state to cloudprovider.InstanceStatus.
func (scaleSet *ScaleSet) instanceStatusFromVM(vm *compute.VirtualMachineScaleSetVM) *cloudprovider.InstanceStatus {
	// Prefer the proactive cache view of the instance state if we aren't in a terminal state
	// This is because the power state may be taking longer to update and we don't want
	// an unfortunate VM update (TTL 5 min) to reset that state to running.
	if vm.ProvisioningState == nil || *vm.ProvisioningState == string(compute.GalleryProvisioningStateUpdating) {
		resourceID, _ := convertResourceGroupNameToLower(*vm.ID)
		providerID := azurePrefix + resourceID
		for _, instance := range scaleSet.instanceCache {
			if instance.Id == providerID {
				return instance.Status
			}
		}
		return nil
	}

	status := &cloudprovider.InstanceStatus{}
	switch *vm.ProvisioningState {
	case string(compute.GalleryProvisioningStateDeleting):
		status.State = cloudprovider.InstanceDeleting
	case string(compute.GalleryProvisioningStateCreating):
		status.State = cloudprovider.InstanceCreating
	case string(compute.GalleryProvisioningStateFailed):
		powerState := vmPowerStateRunning
		if vm.InstanceView != nil && vm.InstanceView.Statuses != nil {
			powerState = vmPowerStateFromStatuses(*vm.InstanceView.Statuses)
		}

		// Provisioning can fail both during instance creation or after the instance is running.
		// Per https://learn.microsoft.com/en-us/azure/virtual-machines/states-billing#provisioning-states,
		// ProvisioningState represents the most recent provisioning state, therefore only report
		// InstanceCreating errors when the power state indicates the instance has not yet started running
		if !isRunningVmPowerState(powerState) {
			klog.V(4).Infof("VM %s reports failed provisioning state with non-running power state: %s", *vm.ID, powerState)
			status.State = cloudprovider.InstanceCreating
			status.ErrorInfo = &cloudprovider.InstanceErrorInfo{
				ErrorClass:   cloudprovider.OutOfResourcesErrorClass,
				ErrorCode:    "provisioning-state-failed",
				ErrorMessage: "Azure failed to provision a node for this node group",
			}
		} else {
			klog.V(5).Infof("VM %s reports a failed provisioning state but is running (%s)", *vm.ID, powerState)
			status.State = cloudprovider.InstanceRunning
		}
	default:
		status.State = cloudprovider.InstanceRunning
	}

	// Add vmssCSE Provisioning Failed Message in error info body for vmssCSE Extensions if enableDetailedCSEMessage is true
	if scaleSet.enableDetailedCSEMessage && vm.InstanceView != nil {
		if err, failed := scaleSet.cseErrors(vm.InstanceView.Extensions); failed {
			status.State = cloudprovider.InstanceCreating
			errorInfo := &cloudprovider.InstanceErrorInfo{
				ErrorClass:   cloudprovider.OtherErrorClass,
				ErrorCode:    vmssExtensionProvisioningFailed,
				ErrorMessage: fmt.Sprintf("%s: %v", to.String(vm.Name), err),
			}
			status.ErrorInfo = errorInfo
		}
	}

	return status
}
