/*
Copyright 2024 The Kubernetes Authors.

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

	resourceapi "k8s.io/api/resource/v1beta1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/dynamic-resource-allocation/structured"
	"k8s.io/klog/v2"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
)

// SharedLister groups all interfaces that Cluster Autoscaler needs to implement for integrating with kube-scheduler.
type SharedLister interface {
	schedulerframework.SharedLister
	schedulerframework.SharedDRAManager
}

// DelegatingSchedulerSharedLister implements schedulerframework interfaces by passing the logic to a delegate. Delegate can be updated.
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

// ResourceClaims returns a ResourceClaimTracker.
func (lister *DelegatingSchedulerSharedLister) ResourceClaims() schedulerframework.ResourceClaimTracker {
	return lister.delegate.ResourceClaims()
}

// ResourceSlices returns a ResourceSliceLister.
func (lister *DelegatingSchedulerSharedLister) ResourceSlices() schedulerframework.ResourceSliceLister {
	return lister.delegate.ResourceSlices()
}

// DeviceClasses returns a DeviceClassLister.
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
type unsetResourceClaimTracker unsetSharedLister
type unsetResourceSliceLister unsetSharedLister
type unsetDeviceClassLister unsetSharedLister

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

func (u unsetResourceClaimTracker) List() ([]*resourceapi.ResourceClaim, error) {
	return nil, fmt.Errorf("lister not set in delegate")
}

func (u unsetResourceClaimTracker) Get(namespace, claimName string) (*resourceapi.ResourceClaim, error) {
	return nil, fmt.Errorf("lister not set in delegate")
}

func (u unsetResourceClaimTracker) ListAllAllocatedDevices() (sets.Set[structured.DeviceID], error) {
	return nil, fmt.Errorf("lister not set in delegate")
}

func (u unsetResourceClaimTracker) SignalClaimPendingAllocation(claimUID types.UID, allocatedClaim *resourceapi.ResourceClaim) error {
	return fmt.Errorf("lister not set in delegate")
}

func (u unsetResourceClaimTracker) ClaimHasPendingAllocation(claimUID types.UID) bool {
	klog.Errorf("lister not set in delegate")
	return false
}

func (u unsetResourceClaimTracker) RemoveClaimPendingAllocation(claimUID types.UID) (deleted bool) {
	klog.Errorf("lister not set in delegate")
	return false
}

func (u unsetResourceClaimTracker) AssumeClaimAfterAPICall(claim *resourceapi.ResourceClaim) error {
	return fmt.Errorf("lister not set in delegate")
}

func (u unsetResourceClaimTracker) AssumedClaimRestore(namespace, claimName string) {
	klog.Errorf("lister not set in delegate")
}

func (u unsetResourceSliceLister) List() ([]*resourceapi.ResourceSlice, error) {
	return nil, fmt.Errorf("lister not set in delegate")
}

func (u unsetDeviceClassLister) List() ([]*resourceapi.DeviceClass, error) {
	return nil, fmt.Errorf("lister not set in delegate")
}

func (u unsetDeviceClassLister) Get(className string) (*resourceapi.DeviceClass, error) {
	return nil, fmt.Errorf("lister not set in delegate")
}

// NodeInfos returns a fake NodeInfoLister which always returns an error
func (lister *unsetSharedLister) NodeInfos() schedulerframework.NodeInfoLister {
	return (*unsetNodeInfoLister)(lister)
}

// StorageInfos returns a fake StorageInfoLister which always returns an error
func (lister *unsetSharedLister) StorageInfos() schedulerframework.StorageInfoLister {
	return (*unsetStorageInfoLister)(lister)
}

func (lister *unsetSharedLister) ResourceClaims() schedulerframework.ResourceClaimTracker {
	return (*unsetResourceClaimTracker)(lister)
}

func (lister *unsetSharedLister) ResourceSlices() schedulerframework.ResourceSliceLister {
	return (*unsetResourceSliceLister)(lister)
}

func (lister *unsetSharedLister) DeviceClasses() schedulerframework.DeviceClassLister {
	return (*unsetDeviceClassLister)(lister)
}

var unsetSharedListerSingleton *unsetSharedLister
