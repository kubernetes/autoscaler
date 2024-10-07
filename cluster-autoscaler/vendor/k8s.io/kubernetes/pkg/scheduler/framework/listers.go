/*
Copyright 2019 The Kubernetes Authors.

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
	resourceapi "k8s.io/api/resource/v1alpha3"
	"k8s.io/apimachinery/pkg/types"
)

// NodeInfoLister interface represents anything that can list/get NodeInfo objects from node name.
type NodeInfoLister interface {
	// List returns the list of NodeInfos.
	List() ([]*NodeInfo, error)
	// HavePodsWithAffinityList returns the list of NodeInfos of nodes with pods with affinity terms.
	HavePodsWithAffinityList() ([]*NodeInfo, error)
	// HavePodsWithRequiredAntiAffinityList returns the list of NodeInfos of nodes with pods with required anti-affinity terms.
	HavePodsWithRequiredAntiAffinityList() ([]*NodeInfo, error)
	// Get returns the NodeInfo of the given node name.
	Get(nodeName string) (*NodeInfo, error)
}

// StorageInfoLister interface represents anything that handles storage-related operations and resources.
type StorageInfoLister interface {
	// IsPVCUsedByPods returns true/false on whether the PVC is used by one or more scheduled pods,
	// keyed in the format "namespace/name".
	IsPVCUsedByPods(key string) bool
}

// SharedLister groups scheduler-specific listers.
type SharedLister interface {
	NodeInfos() NodeInfoLister
	StorageInfos() StorageInfoLister
}

// ResourceSliceLister is used by the DRA plugin to obtain ResourceSlices.
type ResourceSliceLister interface {
	// List returns a list of all ResourceSlices.
	List() ([]*resourceapi.ResourceSlice, error)
}

// DeviceClassLister is used by the DRA plugin to obtain DeviceClasses.
type DeviceClassLister interface {
	// List returns a list of all DeviceClasses.
	List() ([]*resourceapi.DeviceClass, error)
	// Get returns the DeviceClass with the given className.
	Get(className string) (*resourceapi.DeviceClass, error)
}

// ResourceClaimTracker is used by the DRA plugin to obtain ResourceClaims, and track changes to ResourceClaims in-memory.
//
// If the claims are meant to be allocated in the API during the binding phase (when used by scheduler), the tracker helps avoid
// race conditions between scheduling and binding phases (as well as between the binding phase and the informer cache update).
//
// If the binding phase is not used (e.g. when used by Cluster Autoscaler which only runs the scheduling phase, and simulates binding in-memory),
// the tracker allows the framework user to obtain the claim allocations produced by the DRA plugin, and persist them outside of the API (e.g. in-memory).
type ResourceClaimTracker interface {
	// List lists ResourceClaims. The claims can either be obtained from a real informer, or simulated in-memory.
	//
	// If a real informer is used, changes made via AssumeClaimAfterAPICall() should be applied to the obtained claims until
	// they're reflected in the informer cache.
	//
	// If claims are meant to be allocated during the binding phase, changes made via SignalClaimPendingAllocation() should
	// be applied to the obtained claims until RemoveClaimPendingAllocation() is called. If the binding phase is not run,
	// this is not a requirement - the user can choose whether to persist the SignalClaimPendingAllocation() changes.
	List() ([]*resourceapi.ResourceClaim, error)
	// Get works like List(), but for a single claim.
	Get(namespace, claimName string) (*resourceapi.ResourceClaim, error)
	// ListAllAllocated works like List(), but only lists claims that are allocated.
	ListAllAllocated() ([]*resourceapi.ResourceClaim, error)

	// SignalClaimPendingAllocation is called when a Node is picked for a Pod, and the claim allocation is computed at
	// the end of the scheduling phase (in Reserve).
	//
	// If the allocation is meant to be persisted during the binding phase, pods referencing the claim should be blocked
	// from scheduling (ClaimHasPendingAllocation() should answer true) until that happens (RemoveClaimPendingAllocation() is called).
	// The changes should then also be immediately reflected in List/Get/ListAllAllocated calls.
	//
	// If the binding phase is not run, this method allows the framework user to obtain the allocation computed by the DRA
	// plugin (and e.g. persist it in-memory) with no additional requirements.
	SignalClaimPendingAllocation(claimUID types.UID, allocatedClaim *resourceapi.ResourceClaim) error
	// ClaimHasPendingAllocation answers whether a given claim has a pending allocation during the scheduling phase. Claims with a pending allocation
	// block other pods referencing them from scheduling until the allocation is persisted.
	ClaimHasPendingAllocation(claimUID types.UID) bool
	// RemoveClaimPendingAllocation is called in the binding phase when the claim's pending allocation has already been persisted, or
	// has been aborted and won't be persisted. ClaimHasPendingAllocation() should stop answering true after this.
	RemoveClaimPendingAllocation(claimUID types.UID) (deleted bool)

	// AssumeClaimAfterAPICall is called in the binding phase, after the API call which persists the claim allocation. The claim update should
	// be immediately reflected in Get/List/ListAllAllocated calls. This is done to avoid a race condition between the API call, and the informer
	// cache being updated to reflect it.
	AssumeClaimAfterAPICall(claim *resourceapi.ResourceClaim) error
	// AssumedClaimRestore is called in Unreserve in case of errors during Reserve or later phases. It should stop reflecting the changes made via
	// AssumeClaimAfterAPICall and revert back to the informer object.
	AssumedClaimRestore(namespace, claimName string)
}

// SharedDraManager is used by the DRA plugin to obtain DRA objects, and track modifications to them in-memory.
// The plugin's default implementation obtains the objects from the API. A different implementation can be
// plugged into the framework in order to simulate the state of DRA objects. For example, Cluster Autoscaler
// can use this to provide the correct DRA object state to the DRA plugin when simulating scheduling changes in-memory.
type SharedDraManager interface {
	ResourceClaims() ResourceClaimTracker
	ResourceSlices() ResourceSliceLister
	DeviceClasses() DeviceClassLister
}
