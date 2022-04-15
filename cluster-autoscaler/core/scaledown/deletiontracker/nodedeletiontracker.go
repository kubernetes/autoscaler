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

	"k8s.io/autoscaler/cluster-autoscaler/processors/status"
	klog "k8s.io/klog/v2"
)

// NodeDeletionTracker keeps track of node deletions.
// TODO: extend to implement ActuationStatus interface
type NodeDeletionTracker struct {
	sync.Mutex
	nonEmptyNodeDeleteInProgress bool
	// A map of node delete results by node name. It's being constantly emptied into ScaleDownStatus
	// objects in order to notify the ScaleDownStatusProcessor that the node drain has ended or that
	// an error occurred during the deletion process.
	nodeDeleteResults map[string]status.NodeDeleteResult
	// A map which keeps track of deletions in progress for nodepools.
	// Key is a node group id and value is a number of node deletions in progress.
	deletionsInProgress map[string]int
}

// NewNodeDeletionTracker creates new NodeDeletionTracker.
func NewNodeDeletionTracker() *NodeDeletionTracker {
	return &NodeDeletionTracker{
		nodeDeleteResults:   make(map[string]status.NodeDeleteResult),
		deletionsInProgress: make(map[string]int),
	}
}

// IsNonEmptyNodeDeleteInProgress returns true if a non empty node is being deleted.
func (n *NodeDeletionTracker) IsNonEmptyNodeDeleteInProgress() bool {
	n.Lock()
	defer n.Unlock()
	return n.nonEmptyNodeDeleteInProgress
}

// SetNonEmptyNodeDeleteInProgress sets non empty node deletion in progress status.
func (n *NodeDeletionTracker) SetNonEmptyNodeDeleteInProgress(status bool) {
	n.Lock()
	defer n.Unlock()
	n.nonEmptyNodeDeleteInProgress = status
}

// StartDeletion increments node deletion in progress counter for the given nodegroup.
func (n *NodeDeletionTracker) StartDeletion(nodeGroupId string) {
	n.Lock()
	defer n.Unlock()
	n.deletionsInProgress[nodeGroupId]++
}

// EndDeletion decrements node deletion in progress counter for the given nodegroup.
func (n *NodeDeletionTracker) EndDeletion(nodeGroupId string) {
	n.Lock()
	defer n.Unlock()

	value, found := n.deletionsInProgress[nodeGroupId]
	if !found {
		klog.Errorf("This should never happen, counter for %s in DelayedNodeDeletionStatus wasn't found", nodeGroupId)
		return
	}
	if value <= 0 {
		klog.Errorf("This should never happen, counter for %s in DelayedNodeDeletionStatus isn't greater than 0, counter value is %d", nodeGroupId, value)
	}
	n.deletionsInProgress[nodeGroupId]--
	if n.deletionsInProgress[nodeGroupId] <= 0 {
		delete(n.deletionsInProgress, nodeGroupId)
	}
}

// GetDeletionsInProgress returns the number of deletions in progress for the given node group.
func (n *NodeDeletionTracker) GetDeletionsInProgress(nodeGroupId string) int {
	n.Lock()
	defer n.Unlock()
	return n.deletionsInProgress[nodeGroupId]
}

// AddNodeDeleteResult adds a node delete result to the result map.
func (n *NodeDeletionTracker) AddNodeDeleteResult(nodeName string, result status.NodeDeleteResult) {
	n.Lock()
	defer n.Unlock()
	n.nodeDeleteResults[nodeName] = result
}

// GetAndClearNodeDeleteResults returns the whole result map and replaces it with a new empty one.
func (n *NodeDeletionTracker) GetAndClearNodeDeleteResults() map[string]status.NodeDeleteResult {
	n.Lock()
	defer n.Unlock()
	results := n.nodeDeleteResults
	n.nodeDeleteResults = make(map[string]status.NodeDeleteResult)
	return results
}
