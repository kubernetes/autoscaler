/*
Copyright 2025 The Kubernetes Authors.

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

package snapshot

import (
	"fmt"

	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/common"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	schedulerinterface "k8s.io/kube-scheduler/framework"
)

// Snapshot represents a snapshot of CSI node information for cluster simulation.
type Snapshot struct {
	csiNodes *common.PatchSet[string, *storagev1.CSINode]
}

// NewSnapshot creates a new Snapshot from a map of CSI nodes.
// csiNodes argument is a map of csiNode name and corresponding CSINode object.
//
//	{"csi-node-1": &storagev1.CSINode{}}
func NewSnapshot(csiNodes map[string]*storagev1.CSINode) *Snapshot {
	csiNdodePatch := common.NewPatchFromMap(csiNodes)
	return &Snapshot{
		csiNodes: common.NewPatchSet(csiNdodePatch),
	}
}

func (s *Snapshot) listCSINodes() []*storagev1.CSINode {
	csiNodes := s.csiNodes.AsMap()
	csiNodesList := make([]*storagev1.CSINode, 0, len(csiNodes))
	for _, csiNode := range csiNodes {
		csiNodesList = append(csiNodesList, csiNode)
	}
	return csiNodesList
}

// CSINodes returns a CSI node lister for the snapshot.
func (s *Snapshot) CSINodes() schedulerinterface.CSINodeLister {
	return SnapshotCSINodeLister{snapshot: s}
}

// AddCSINodes adds a list of CSI nodes to the snapshot.
func (s *Snapshot) AddCSINodes(csiNodes []*storagev1.CSINode) error {
	for _, csiNode := range csiNodes {
		err := s.AddCSINode(csiNode)
		if err != nil {
			return err
		}
	}
	return nil
}

// Get returns a CSI node from the snapshot by name.
func (s *Snapshot) Get(name string) (*storagev1.CSINode, error) {
	csiNode, found := s.csiNodes.FindValue(name)
	if !found {
		return nil, fmt.Errorf("csi nodes %s not found", name)
	}
	return csiNode, nil
}

// AddCSINode adds a CSI node to the snapshot.
func (s *Snapshot) AddCSINode(csiNode *storagev1.CSINode) error {
	if _, alreadyInSnapshot := s.csiNodes.FindValue(csiNode.Name); alreadyInSnapshot {
		return fmt.Errorf("csi node %s already in snapshot", csiNode.Name)
	}

	s.csiNodes.SetCurrent(csiNode.Name, csiNode)
	return nil
}

// AddCSINodeInfoToNodeInfo adds a CSI node to the node info.
func (s *Snapshot) AddCSINodeInfoToNodeInfo(nodeInfo *framework.NodeInfo) (*framework.NodeInfo, error) {
	csiNode, err := s.Get(nodeInfo.Node().Name)
	if err != nil {
		return nil, err
	}
	return nodeInfo.SetCSINode(csiNode), nil
}

// RemoveCSINode removes a CSI node from the snapshot.
func (s *Snapshot) RemoveCSINode(name string) {
	s.csiNodes.DeleteCurrent(name)
}

// Commit commits the snapshot.
func (s *Snapshot) Commit() {
	s.csiNodes.Commit()
}

// Revert reverts the snapshot.
func (s *Snapshot) Revert() {
	s.csiNodes.Revert()
}

// Fork forks the snapshot.
func (s *Snapshot) Fork() {
	s.csiNodes.Fork()
}

// NewEmptySnapshot creates a new empty snapshot.
func NewEmptySnapshot() *Snapshot {
	csiNdodePatch := common.NewPatch[string, *storagev1.CSINode]()
	return &Snapshot{
		csiNodes: common.NewPatchSet(csiNdodePatch),
	}
}
