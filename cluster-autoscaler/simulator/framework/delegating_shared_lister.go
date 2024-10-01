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

package framework

import (
	"fmt"

	resourceapi "k8s.io/api/resource/v1alpha3"
	"k8s.io/apimachinery/pkg/types"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
)

type SharedLister interface {
	schedulerframework.SharedLister
	schedulerframework.SharedDraManager
}

// DelegatingSchedulerSharedLister is an implementation of scheduler.SharedLister which
// passes logic to delegate. Delegate can be updated.
type DelegatingSchedulerSharedLister struct {
	delegate SharedLister
}

// NewDelegatingSchedulerSharedLister creates new NewDelegatingSchedulerSharedLister
func NewDelegatingSchedulerSharedLister() *DelegatingSchedulerSharedLister {
	return &DelegatingSchedulerSharedLister{
		delegate: unsetSharedListerSingleton,
	}
}

// NodeInfos returns a NodeInfoLister.
func (lister *DelegatingSchedulerSharedLister) NodeInfos() schedulerframework.NodeInfoLister {
	return lister.delegate.NodeInfos()
}

// StorageInfos returns a StorageInfoLister
func (lister *DelegatingSchedulerSharedLister) StorageInfos() schedulerframework.StorageInfoLister {
	return lister.delegate.StorageInfos()
}

func (lister *DelegatingSchedulerSharedLister) ResourceClaims() schedulerframework.ResourceClaimTracker {
	return lister.delegate.ResourceClaims()
}

func (lister *DelegatingSchedulerSharedLister) ResourceSlices() schedulerframework.ResourceSliceLister {
	return lister.delegate.ResourceSlices()
}

func (lister *DelegatingSchedulerSharedLister) DeviceClasses() schedulerframework.DeviceClassLister {
	return lister.delegate.DeviceClasses()
}

// UpdateDelegate updates the delegate
func (lister *DelegatingSchedulerSharedLister) UpdateDelegate(delegate SharedLister) {
	lister.delegate = delegate
}

// ResetDelegate resets delegate to
func (lister *DelegatingSchedulerSharedLister) ResetDelegate() {
	lister.delegate = unsetSharedListerSingleton
}

type unsetSharedLister struct{}
type unsetNodeInfoLister unsetSharedLister
type unsetStorageInfoLister unsetSharedLister
type unsetDeviceClassLister unsetSharedLister
type unsetResourceSliceLister unsetSharedLister
type unsetResourceClaimsTracker unsetSharedLister

// List always returns an error
func (lister *unsetNodeInfoLister) List() ([]*schedulerframework.NodeInfo, error) {
	return nil, fmt.Errorf("lister not set in delegate")
}

// HavePodsWithAffinityList always returns an error
func (lister *unsetNodeInfoLister) HavePodsWithAffinityList() ([]*schedulerframework.NodeInfo, error) {
	return nil, fmt.Errorf("lister not set in delegate")
}

// HavePodsWithRequiredAntiAffinityList always returns an error.
func (lister *unsetNodeInfoLister) HavePodsWithRequiredAntiAffinityList() ([]*schedulerframework.NodeInfo, error) {
	return nil, fmt.Errorf("lister not set in delegate")
}

// Get always returns an error
func (lister *unsetNodeInfoLister) Get(nodeName string) (*schedulerframework.NodeInfo, error) {
	return nil, fmt.Errorf("lister not set in delegate")
}

func (lister *unsetStorageInfoLister) IsPVCUsedByPods(key string) bool {
	return false
}
func (u unsetDeviceClassLister) List() ([]*resourceapi.DeviceClass, error) {
	return nil, fmt.Errorf("lister not set in delegate")
}

func (u unsetDeviceClassLister) Get(className string) (*resourceapi.DeviceClass, error) {
	return nil, fmt.Errorf("lister not set in delegate")
}

func (u unsetResourceSliceLister) List() ([]*resourceapi.ResourceSlice, error) {
	return nil, fmt.Errorf("lister not set in delegate")
}

func (u unsetResourceClaimsTracker) Get(namespace, claimName string) (*resourceapi.ResourceClaim, error) {
	return nil, fmt.Errorf("lister not set in delegate")
}

func (u unsetResourceClaimsTracker) List() ([]*resourceapi.ResourceClaim, error) {
	return nil, fmt.Errorf("lister not set in delegate")
}

func (u unsetResourceClaimsTracker) ListAllAllocated() ([]*resourceapi.ResourceClaim, error) {
	return nil, fmt.Errorf("lister not set in delegate")
}

func (u unsetResourceClaimsTracker) SignalClaimPendingAllocation(claimUid types.UID, allocatedClaim *resourceapi.ResourceClaim) error {
	return fmt.Errorf("lister not set in delegate")
}

func (u unsetResourceClaimsTracker) ClaimHasPendingAllocation(claimUid types.UID) bool {
	return false
}

func (u unsetResourceClaimsTracker) RemoveClaimPendingAllocation(claimUid types.UID) (deleted bool) {
	return false
}

func (u unsetResourceClaimsTracker) AssumeClaimAfterAPICall(claim *resourceapi.ResourceClaim) error {
	return fmt.Errorf("lister not set in delegate")
}

func (u unsetResourceClaimsTracker) AssumedClaimRestore(namespace, claimName string) {
}

// NodeInfos: Pods returns a fake NodeInfoLister which always returns an error
func (lister *unsetSharedLister) NodeInfos() schedulerframework.NodeInfoLister {
	return (*unsetNodeInfoLister)(lister)
}

// StorageInfos: Pods returns a fake StorageInfoLister which always returns an error
func (lister *unsetSharedLister) StorageInfos() schedulerframework.StorageInfoLister {
	return (*unsetStorageInfoLister)(lister)
}

func (lister *unsetSharedLister) ResourceClaims() schedulerframework.ResourceClaimTracker {
	return (*unsetResourceClaimsTracker)(lister)
}

func (lister *unsetSharedLister) ResourceSlices() schedulerframework.ResourceSliceLister {
	return (*unsetResourceSliceLister)(lister)
}

func (lister *unsetSharedLister) DeviceClasses() schedulerframework.DeviceClassLister {
	return (*unsetDeviceClassLister)(lister)

}

var unsetSharedListerSingleton *unsetSharedLister
