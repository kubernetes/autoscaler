package snapshot

import (
	"fmt"

	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/common"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
)

type Snapshot struct {
	csiNodes *common.PatchSet[string, *storagev1.CSINode]
}

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

func (s *Snapshot) CSINodes() SnapshotCSINodeLister {
	return SnapshotCSINodeLister{snapshot: s}
}

func (s *Snapshot) AddCSINodes(csiNodes []*storagev1.CSINode) error {
	for _, csiNode := range csiNodes {
		if _, alreadyInSnapshot := s.csiNodes.FindValue(csiNode.Name); alreadyInSnapshot {
			return fmt.Errorf("csi node %s already in snapshot", csiNode.Name)
		}
	}

	for _, csiNode := range csiNodes {
		s.csiNodes.SetCurrent(csiNode.Name, csiNode)
	}

	return nil
}

func (s *Snapshot) Get(name string) (*storagev1.CSINode, error) {
	csiNode, found := s.csiNodes.FindValue(name)
	if !found {
		return nil, fmt.Errorf("csi nodes %s not found", name)
	}
	return csiNode, nil
}

func (s *Snapshot) WrapSchedulerNodeInfo(schedNodeInfo *schedulerframework.NodeInfo) (*framework.NodeInfo, error) {
	csiNode, err := s.Get(schedNodeInfo.Node().Name)
	if err != nil {
		return nil, err
	}
	nodeInfo := framework.WrapSchedulerNodeInfo(schedNodeInfo, nil, nil)
	nodeInfo.CSINode = csiNode
	return nodeInfo, nil
}

func (s *Snapshot) AddCSINode(csiNode *storagev1.CSINode) error {
	if _, alreadyInSnapshot := s.csiNodes.FindValue(csiNode.Name); alreadyInSnapshot {
		return fmt.Errorf("csi node %s already in snapshot", csiNode.Name)
	}

	s.csiNodes.SetCurrent(csiNode.Name, csiNode)
	return nil
}

func (s *Snapshot) AddCSINodeInfoToNodeInfo(nodeInfo *framework.NodeInfo) (*framework.NodeInfo, error) {
	csiNode, err := s.Get(nodeInfo.Node().Name)
	if err != nil {
		return nil, err
	}
	return nodeInfo.AddCSINode(csiNode), nil
}

func (s *Snapshot) RemoveCSINode(name string) {
	s.csiNodes.DeleteCurrent(name)
}

func (s *Snapshot) Commit() {
	s.csiNodes.Commit()
}

func (s *Snapshot) Revert() {
	s.csiNodes.Revert()
}

func (s *Snapshot) Fork() {
	s.csiNodes.Fork()
}

func NewEmptySnapshot() *Snapshot {
	csiNdodePatch := common.NewPatch[string, *storagev1.CSINode]()
	return &Snapshot{
		csiNodes: common.NewPatchSet(csiNdodePatch),
	}
}
