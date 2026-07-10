/*
Copyright The Kubernetes Authors.

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

package state

import (
	"sync"
	"sync/atomic"

	"github.com/samber/lo"
	"k8s.io/apimachinery/pkg/util/sets"

	v1 "sigs.k8s.io/karpenter/pkg/apis/v1"
)

// Currently NodeClaims be in one of these states
type NodeClaimState struct {
	// NodeClaims that have been launched
	Active sets.Set[string]
	// NodeClaims that are pending disruption
	PendingDisruption sets.Set[string]
	// NodeClaims are marked for Deletion
	Deleting sets.Set[string]
}

// StateNodePool is a cached version of a NodePool in the cluster that maintains state which is expensive to compute every time it's needed.
type NodePoolState struct {
	mu sync.RWMutex

	nodePoolNameToNodeClaimState map[string]NodeClaimState // node pool name -> node claim state (Active and Deleting node claim names)
	nodeClaimNameToNodePoolName  map[string]string         // node claim name -> node pool name
	nodePoolNameToNodePoolLimit  map[string]*atomic.Int64  // node pool -> nodepool limit
}

func NewNodePoolState() *NodePoolState {
	return &NodePoolState{
		nodePoolNameToNodeClaimState: map[string]NodeClaimState{},
		nodeClaimNameToNodePoolName:  map[string]string{},
		nodePoolNameToNodePoolLimit:  map[string]*atomic.Int64{},
	}
}

// Helper methods that hold the lock

// Sets up the NodePoolState to track NodeClaim
func (n *NodePoolState) SetNodeClaimMapping(npName, ncName string) {
	if npName == "" || ncName == "" {
		return
	}
	n.mu.Lock()
	defer n.mu.Unlock()

	n.ensureNodePoolEntry(npName)
	n.nodeClaimNameToNodePoolName[ncName] = npName
}

// Marks the given NodeClaim as active in NodePoolState
func (n *NodePoolState) MarkNodeClaimActive(npName, ncName string) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.ensureNodePoolEntry(npName)

	n.nodePoolNameToNodeClaimState[npName].PendingDisruption.Delete(ncName)
	n.nodePoolNameToNodeClaimState[npName].Deleting.Delete(ncName)
	n.nodePoolNameToNodeClaimState[npName].Active.Insert(ncName)
}

// Marks the given NodeClaim as Deleting in NodePoolState
func (n *NodePoolState) MarkNodeClaimDeleting(npName, ncName string) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.ensureNodePoolEntry(npName)

	n.nodePoolNameToNodeClaimState[npName].PendingDisruption.Delete(ncName)
	n.nodePoolNameToNodeClaimState[npName].Deleting.Insert(ncName)
	n.nodePoolNameToNodeClaimState[npName].Active.Delete(ncName)
}

// Marks the given NodeClaim as Deleting in NodePoolState
func (n *NodePoolState) MarkNodeClaimPendingDisruption(npName, ncName string) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.ensureNodePoolEntry(npName)

	n.nodePoolNameToNodeClaimState[npName].Active.Delete(ncName)
	n.nodePoolNameToNodeClaimState[npName].Deleting.Delete(ncName)
	n.nodePoolNameToNodeClaimState[npName].PendingDisruption.Insert(ncName)
}

// Cleans up the NodeClaim in NodePoolState and NodePool keys if NodePool is deleted or its sized down to 0
func (n *NodePoolState) Cleanup(ncName string) {
	n.mu.Lock()
	defer n.mu.Unlock()

	npName := n.nodeClaimNameToNodePoolName[ncName]

	if npState, exists := n.nodePoolNameToNodeClaimState[npName]; exists {
		npState.Deleting.Delete(ncName)
		npState.Active.Delete(ncName)
		npState.PendingDisruption.Delete(ncName)

		if npState.Active.Len() == 0 && npState.Deleting.Len() == 0 {
			delete(n.nodePoolNameToNodeClaimState, npName)
			delete(n.nodePoolNameToNodePoolLimit, npName)
		}
	}

	delete(n.nodeClaimNameToNodePoolName, ncName)
}

// Returns the current NodeClaims for a NodePool by its state (active, deleting, pendingdisruption)
func (n *NodePoolState) GetNodeCount(npName string) (active, deleting, pendingdisruption int) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	return n.nodeCounts(npName)
}

// ReserveNodeCount attempts to reserve nodes against a NodePool's limit.
// It ensures that the total of active nodes + deleting nodes + reserved nodes doesn't exceed the limit.
func (n *NodePoolState) ReserveNodeCount(np string, limit int64, wantedLimit int64) int64 {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.ensureNodePoolEntry(np)
	active, deleting, pendingdisruption := n.nodeCounts(np)

	// We retry until CompareAndSwap is successful
	for {
		currentlyReserved := n.nodePoolNameToNodePoolLimit[np].Load()
		remainingLimit := limit - int64(active+deleting+pendingdisruption) - currentlyReserved
		if remainingLimit < 0 {
			return 0
		}
		grantedLimit := lo.Ternary(wantedLimit > remainingLimit,
			remainingLimit,
			wantedLimit,
		)

		if n.nodePoolNameToNodePoolLimit[np].CompareAndSwap(currentlyReserved, currentlyReserved+grantedLimit) {
			return grantedLimit
		}
	}
}

// ReleaseNodeCount releases the NodePoolTracker ReservedNodeLimit
func (n *NodePoolState) ReleaseNodeCount(npName string, count int64) {
	n.mu.Lock()
	defer n.mu.Unlock()

	// We retry until CompareAndSwap is successful
	for {
		currentlyReserved := n.nodePoolNameToNodePoolLimit[npName].Load()
		if n.nodePoolNameToNodePoolLimit[npName].CompareAndSwap(
			currentlyReserved,
			lo.Ternary(currentlyReserved-count < 0, 0, currentlyReserved-count)) {
			return
		}
	}
}

// Methods that expect the caller to hold the lock

func (n *NodePoolState) nodeCounts(npName string) (active, deleting, pendingdisruption int) {
	if st, ok := n.nodePoolNameToNodeClaimState[npName]; ok {
		return len(st.Active), len(st.Deleting), len(st.PendingDisruption)
	}
	return 0, 0, 0
}

// Updates the NodeClaim state and releases the Limit if NodeClaim transitions from not Active to Active
func (n *NodePoolState) UpdateNodeClaim(nodeClaim *v1.NodeClaim, markedForDeletion bool) {
	// If we are launching/deleting a NodeClaim we need to track the state of the NodeClaim and its limits
	npName := nodeClaim.Labels[v1.NodePoolLabelKey]
	if npName == "" {
		return
	}
	n.SetNodeClaimMapping(npName, nodeClaim.Name)

	// If our node/nodeclaim is marked for deletion, we need to make sure that we delete it
	if markedForDeletion {
		n.MarkNodeClaimDeleting(npName, nodeClaim.Name)
	} else {
		n.MarkNodeClaimActive(npName, nodeClaim.Name)
	}
}

func (n *NodePoolState) ensureNodePoolEntry(np string) {
	if _, ok := n.nodePoolNameToNodeClaimState[np]; !ok {
		n.nodePoolNameToNodeClaimState[np] = NodeClaimState{
			Active:            sets.New[string](),
			Deleting:          sets.New[string](),
			PendingDisruption: sets.New[string](),
		}
	}
	if _, ok := n.nodePoolNameToNodePoolLimit[np]; !ok {
		n.nodePoolNameToNodePoolLimit[np] = &atomic.Int64{}
	}
}

func (n *NodePoolState) Reset() {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.nodePoolNameToNodeClaimState = make(map[string]NodeClaimState)
	n.nodeClaimNameToNodePoolName = make(map[string]string)
	n.nodePoolNameToNodePoolLimit = make(map[string]*atomic.Int64)
}
