/*
Copyright 2022 The Kubernetes Authors.

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

package unremovable

import (
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"

	apiv1 "k8s.io/api/core/v1"
	klog "k8s.io/klog/v2"
)

// Nodes tracks the state of cluster nodes that cannot be removed.
type Nodes struct {
	ttls    map[string]time.Time
	reasons map[string]*simulator.UnremovableNode
}

// NewNodes returns a new initialized Nodes object.
func NewNodes() *Nodes {
	return &Nodes{
		ttls:    make(map[string]time.Time),
		reasons: make(map[string]*simulator.UnremovableNode),
	}
}

// NodeInfoGetter is anything that can return NodeInfo object by name.
type NodeInfoGetter interface {
	GetNodeInfo(name string) (*framework.NodeInfo, error)
}

// Update updates the internal structure according to current state of the
// cluster. Removes the nodes that are no longer in the nodes list.
func (n *Nodes) Update(nodeInfos NodeInfoGetter, timestamp time.Time) {
	n.reasons = make(map[string]*simulator.UnremovableNode)
	if len(n.ttls) <= 0 {
		return
	}
	newTTLs := make(map[string]time.Time, len(n.ttls))
	for name, ttl := range n.ttls {
		if _, err := nodeInfos.GetNodeInfo(name); err != nil {
			// Not logging on error level as most likely cause is that node is no longer in the cluster.
			klog.Infof("Can't retrieve node %s from snapshot, removing from unremovable nodes, err: %v", name, err)
			continue
		}
		if ttl.After(timestamp) {
			// Keep nodes that are still in the cluster and haven't expired yet.
			newTTLs[name] = ttl
		}
	}
	n.ttls = newTTLs
}

// Contains returns true iff a given node is unremovable.
func (n *Nodes) Contains(nodeName string) bool {
	_, found := n.reasons[nodeName]
	return found
}

// Add adds an unremovable node.
func (n *Nodes) Add(node *simulator.UnremovableNode) {
	n.reasons[node.Node.Name] = node
}

// AddTimeout adds a new unremovable node with a timeout until which the node
// should be considered unremovable.
func (n *Nodes) AddTimeout(node *simulator.UnremovableNode, timeout time.Time) {
	n.ttls[node.Node.Name] = timeout
	n.Add(node)
}

// AddReason adds an unremovable node due to the specified reason.
func (n *Nodes) AddReason(node *apiv1.Node, reason simulator.UnremovableReason) {
	n.Add(&simulator.UnremovableNode{Node: node, Reason: reason, BlockingPod: nil})
}

// AsList returns a list of unremovable nodes.
func (n *Nodes) AsList() []*simulator.UnremovableNode {
	ns := make([]*simulator.UnremovableNode, 0, len(n.reasons))
	for _, node := range n.reasons {
		ns = append(ns, node)
	}
	return ns
}

// HasReason returns true iff a given node has a reason to be unremovable.
func (n *Nodes) HasReason(nodeName string) bool {
	_, found := n.reasons[nodeName]
	return found
}

// IsRecent returns true iff a given node was recently added
func (n *Nodes) IsRecent(nodeName string) bool {
	_, found := n.ttls[nodeName]
	return found
}
