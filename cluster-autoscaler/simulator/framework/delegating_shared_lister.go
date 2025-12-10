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

	corev1 "k8s.io/api/core/v1"
	resourceapi "k8s.io/api/resource/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/dynamic-resource-allocation/structured"
	"k8s.io/klog/v2"
	fwk "k8s.io/kube-scheduler/framework"
)

// SharedLister groups all interfaces that Cluster Autoscaler needs to implement for integrating with kube-scheduler.
type SharedLister interface {
	fwk.SharedLister
	fwk.SharedDRAManager
}

// DelegatingSchedulerSharedLister implements k8s.io/kube-scheduler/framework interfaces by passing the logic to a delegate. Delegate can be updated.
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
func (lister *DelegatingSchedulerSharedLister) NodeInfos() fwk.NodeInfoLister {
	return lister.delegate.NodeInfos()
}

// StorageInfos returns a StorageInfoLister
func (lister *DelegatingSchedulerSharedLister) StorageInfos() fwk.StorageInfoLister {
	return lister.delegate.StorageInfos()
}

// ResourceClaims returns a ResourceClaimTracker.
func (lister *DelegatingSchedulerSharedLister) ResourceClaims() fwk.ResourceClaimTracker {
	return lister.delegate.ResourceClaims()
}

// ResourceSlices returns a ResourceSliceLister.
func (lister *DelegatingSchedulerSharedLister) ResourceSlices() fwk.ResourceSliceLister {
	return lister.delegate.ResourceSlices()
}

// DeviceClasses returns a DeviceClassLister.
func (lister *DelegatingSchedulerSharedLister) DeviceClasses() fwk.DeviceClassLister {
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

// DeviceClassResolver returns a DeviceClassResolver.
func (lister *DelegatingSchedulerSharedLister) DeviceClassResolver() fwk.DeviceClassResolver {
	return lister.delegate.DeviceClassResolver()
}

type unsetSharedLister struct{}
type unsetNodeInfoLister unsetSharedLister
type unsetStorageInfoLister unsetSharedLister
type unsetResourceClaimTracker unsetSharedLister
type unsetResourceSliceLister unsetSharedLister
type unsetDeviceClassLister unsetSharedLister
type unsetDeviceClassResolver unsetSharedLister

// List always returns an error
func (lister *unsetNodeInfoLister) List() ([]fwk.NodeInfo, error) {
	return nil, fmt.Errorf("lister not set in delegate")
}

// HavePodsWithAffinityList always returns an error
func (lister *unsetNodeInfoLister) HavePodsWithAffinityList() ([]fwk.NodeInfo, error) {
	return nil, fmt.Errorf("lister not set in delegate")
}

// HavePodsWithRequiredAntiAffinityList always returns an error.
func (lister *unsetNodeInfoLister) HavePodsWithRequiredAntiAffinityList() ([]fwk.NodeInfo, error) {
	return nil, fmt.Errorf("lister not set in delegate")
}

// Get always returns an error
func (lister *unsetNodeInfoLister) Get(nodeName string) (fwk.NodeInfo, error) {
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

func (u unsetResourceClaimTracker) GatherAllocatedState() (*structured.AllocatedState, error) {
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

func (u unsetResourceSliceLister) ListWithDeviceTaintRules() ([]*resourceapi.ResourceSlice, error) {
	return nil, fmt.Errorf("lister not set in delegate")
}

func (u unsetDeviceClassLister) List() ([]*resourceapi.DeviceClass, error) {
	return nil, fmt.Errorf("lister not set in delegate")
}

func (u unsetDeviceClassLister) Get(className string) (*resourceapi.DeviceClass, error) {
	return nil, fmt.Errorf("lister not set in delegate")
}

func (u unsetDeviceClassResolver) GetDeviceClass(resourceName corev1.ResourceName) *resourceapi.DeviceClass {
	klog.Errorf("lister not set in delegate")
	return nil
}

// NodeInfos returns a fake NodeInfoLister which always returns an error
func (lister *unsetSharedLister) NodeInfos() fwk.NodeInfoLister {
	return (*unsetNodeInfoLister)(lister)
}

// StorageInfos returns a fake StorageInfoLister which always returns an error
func (lister *unsetSharedLister) StorageInfos() fwk.StorageInfoLister {
	return (*unsetStorageInfoLister)(lister)
}

func (lister *unsetSharedLister) ResourceClaims() fwk.ResourceClaimTracker {
	return (*unsetResourceClaimTracker)(lister)
}

func (lister *unsetSharedLister) ResourceSlices() fwk.ResourceSliceLister {
	return (*unsetResourceSliceLister)(lister)
}

func (lister *unsetSharedLister) DeviceClasses() fwk.DeviceClassLister {
	return (*unsetDeviceClassLister)(lister)
}

func (lister *unsetSharedLister) DeviceClassResolver() fwk.DeviceClassResolver {
	return (*unsetDeviceClassResolver)(lister)
}

var unsetSharedListerSingleton *unsetSharedLister
