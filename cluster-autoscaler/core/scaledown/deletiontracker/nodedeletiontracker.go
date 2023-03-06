/*
Copyright 2016 The Kubernetes Authors.

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

package deletiontracker

import (
	"sync"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/status"
	"k8s.io/autoscaler/cluster-autoscaler/utils/expiring"

	apiv1 "k8s.io/api/core/v1"
	klog "k8s.io/klog/v2"
	"k8s.io/utils/clock"
)

// NodeDeletionTracker keeps track of node deletions.
type NodeDeletionTracker struct {
	sync.Mutex
	// A map which keeps track of deletions in progress for nodepools.
	// Key is a node group id and value is a number of node deletions in progress.
	deletionsPerNodeGroup map[string]int
	// This mapping contains node names of all empty nodes currently undergoing deletion.
	emptyNodeDeletions map[string]bool
	// This mapping contains node names of all nodes currently undergoing drain and deletion.
	drainedNodeDeletions map[string]bool
	// Clock for checking current time.
	clock clock.PassiveClock
	// Helper struct for tracking pod evictions.
	evictions *expiring.List
	// How long evictions are considered as recent.
	evictionsTTL time.Duration
	// Helper struct for tracking deletion results.
	deletionResults *expiring.List
}

type deletionResult struct {
	nodeName string
	result   status.NodeDeleteResult
}

// NewNodeDeletionTracker creates new NodeDeletionTracker.
func NewNodeDeletionTracker(podEvictionsTTL time.Duration) *NodeDeletionTracker {
	return &NodeDeletionTracker{
		deletionsPerNodeGroup: make(map[string]int),
		emptyNodeDeletions:    make(map[string]bool),
		drainedNodeDeletions:  make(map[string]bool),
		clock:                 clock.RealClock{},
		evictions:             expiring.NewList(),
		evictionsTTL:          podEvictionsTTL,
		deletionResults:       expiring.NewList(),
	}
}

// StartDeletion increments node deletion in progress counter for the given nodegroup.
func (n *NodeDeletionTracker) StartDeletion(nodeGroupId, nodeName string) {
	n.Lock()
	defer n.Unlock()
	n.deletionsPerNodeGroup[nodeGroupId]++
	n.emptyNodeDeletions[nodeName] = true
}

// StartDeletionWithDrain is equivalent to StartDeletion, but for counting nodes that are drained first.
func (n *NodeDeletionTracker) StartDeletionWithDrain(nodeGroupId, nodeName string) {
	n.Lock()
	defer n.Unlock()
	n.deletionsPerNodeGroup[nodeGroupId]++
	n.drainedNodeDeletions[nodeName] = true
}

// EndDeletion decrements node deletion in progress counter for the given nodegroup.
func (n *NodeDeletionTracker) EndDeletion(nodeGroupId, nodeName string, result status.NodeDeleteResult) {
	n.Lock()
	defer n.Unlock()

	n.deletionResults.RegisterElement(&deletionResult{nodeName, result})
	value, found := n.deletionsPerNodeGroup[nodeGroupId]
	if !found {
		klog.Errorf("This should never happen, counter for %s in NodeDeletionTracker wasn't found", nodeGroupId)
		return
	}
	if value <= 0 {
		klog.Errorf("This should never happen, counter for %s in NodeDeletionTracker isn't greater than 0, counter value is %d", nodeGroupId, value)
	}
	n.deletionsPerNodeGroup[nodeGroupId]--
	if n.deletionsPerNodeGroup[nodeGroupId] <= 0 {
		delete(n.deletionsPerNodeGroup, nodeGroupId)
	}
	delete(n.emptyNodeDeletions, nodeName)
	delete(n.drainedNodeDeletions, nodeName)
}

// DeletionsInProgress returns a list of all node names currently undergoing deletion.
func (n *NodeDeletionTracker) DeletionsInProgress() ([]string, []string) {
	n.Lock()
	defer n.Unlock()
	return mapKeysSlice(n.emptyNodeDeletions), mapKeysSlice(n.drainedNodeDeletions)
}

func mapKeysSlice(m map[string]bool) []string {
	s := make([]string, len(m))
	i := 0
	for k := range m {
		s[i] = k
		i++
	}
	return s
}

// RegisterEviction stores information about a pod that was recently evicted.
func (n *NodeDeletionTracker) RegisterEviction(pod *apiv1.Pod) {
	n.Lock()
	defer n.Unlock()
	n.evictions.RegisterElement(pod)
}

// RecentEvictions returns a list of pods that were recently evicted by Cluster Autoscaler.
func (n *NodeDeletionTracker) RecentEvictions() []*apiv1.Pod {
	n.Lock()
	defer n.Unlock()
	n.evictions.DropNotNewerThan(n.clock.Now().Add(-n.evictionsTTL))
	els := n.evictions.ToSlice()
	pods := make([]*apiv1.Pod, 0, len(els))
	for _, el := range els {
		pods = append(pods, el.(*apiv1.Pod))
	}
	return pods
}

// DeletionsCount returns the number of deletions in progress for the given node group.
func (n *NodeDeletionTracker) DeletionsCount(nodeGroupId string) int {
	n.Lock()
	defer n.Unlock()
	return n.deletionsPerNodeGroup[nodeGroupId]
}

// DeletionResults returns deletion results in a map form, along with the timestamp of last result.
func (n *NodeDeletionTracker) DeletionResults() (map[string]status.NodeDeleteResult, time.Time) {
	n.Lock()
	defer n.Unlock()
	els, ts := n.deletionResults.ToSliceWithTimestamp()
	drs := make(map[string]status.NodeDeleteResult)
	for _, el := range els {
		dr := el.(*deletionResult)
		drs[dr.nodeName] = dr.result
	}
	return drs, ts
}

// ClearResultsNotNewerThan iterates over existing deletion results and keeps
// only the ones that are newer than the provided timestamp.
func (n *NodeDeletionTracker) ClearResultsNotNewerThan(t time.Time) {
	n.Lock()
	defer n.Unlock()
	n.deletionResults.DropNotNewerThan(t)
}

// Snapshot return a copy of NodeDeletionTracker.
func (n *NodeDeletionTracker) Snapshot() *NodeDeletionTracker {
	n.Lock()
	defer n.Unlock()
	snapshot := NewNodeDeletionTracker(n.evictionsTTL)
	for k, val := range n.emptyNodeDeletions {
		snapshot.emptyNodeDeletions[k] = val
	}
	for k, val := range n.drainedNodeDeletions {
		snapshot.drainedNodeDeletions[k] = val
	}
	for k, val := range n.deletionsPerNodeGroup {
		snapshot.deletionsPerNodeGroup[k] = val
	}
	for _, eviction := range n.evictions.ToSlice() {
		snapshot.evictions.RegisterElement(eviction)
	}
	for _, result := range n.deletionResults.ToSlice() {
		snapshot.deletionResults.RegisterElement(result)
	}
	return snapshot
}
