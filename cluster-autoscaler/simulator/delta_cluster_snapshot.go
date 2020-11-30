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

package simulator

import (
	"fmt"

	apiv1 "k8s.io/api/core/v1"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
)

// DeltaClusterSnapshot is an implementation of ClusterSnapshot optimized for typical Cluster Autoscaler usage - (fork, add stuff, revert), repeated many times per loop.
//
// Complexity of some notable operations:
//	fork - O(1)
//	revert - O(1)
//	commit - O(n)
//	list all pods (no filtering) - O(n), cached
//	list all pods (with filtering) - O(n)
//	list node infos - O(n), cached
//
// Watch out for:
//	node deletions, pod additions & deletions - invalidates cache of current snapshot
//		(when forked affects delta, but not base.)
//	pod affinity - causes scheduler framework to list pods with non-empty selector,
//		so basic caching doesn't help.
//
type DeltaClusterSnapshot struct {
	data *internalDeltaSnapshotData
}

type deltaSnapshotNodeLister DeltaClusterSnapshot

type internalDeltaSnapshotData struct {
	baseData *internalDeltaSnapshotData

	addedNodeInfoMap    map[string]*schedulerframework.NodeInfo
	modifiedNodeInfoMap map[string]*schedulerframework.NodeInfo
	deletedNodeInfos    map[string]bool

	nodeInfoList                     []*schedulerframework.NodeInfo
	havePodsWithAffinity             []*schedulerframework.NodeInfo
	havePodsWithRequiredAntiAffinity []*schedulerframework.NodeInfo
}

func newInternalDeltaSnapshotData() *internalDeltaSnapshotData {
	return &internalDeltaSnapshotData{
		addedNodeInfoMap:    make(map[string]*schedulerframework.NodeInfo),
		modifiedNodeInfoMap: make(map[string]*schedulerframework.NodeInfo),
		deletedNodeInfos:    make(map[string]bool),
	}
}

func (data *internalDeltaSnapshotData) getNodeInfo(name string) (*schedulerframework.NodeInfo, bool) {
	if data == nil {
		return nil, false
	}
	if nodeInfo, found := data.getNodeInfoLocal(name); found {
		return nodeInfo, found
	}
	if data.deletedNodeInfos[name] {
		return nil, false
	}
	return data.baseData.getNodeInfo(name)
}

func (data *internalDeltaSnapshotData) getNodeInfoLocal(name string) (*schedulerframework.NodeInfo, bool) {
	if data == nil {
		return nil, false
	}
	if nodeInfo, found := data.addedNodeInfoMap[name]; found {
		return nodeInfo, true
	}
	if nodeInfo, found := data.modifiedNodeInfoMap[name]; found {
		return nodeInfo, true
	}
	return nil, false
}

func (data *internalDeltaSnapshotData) getNodeInfoList() []*schedulerframework.NodeInfo {
	if data == nil {
		return nil
	}
	if data.nodeInfoList == nil {
		data.nodeInfoList = data.buildNodeInfoList()
	}
	return data.nodeInfoList
}

// Contains costly copying throughout the struct chain. Use wisely.
func (data *internalDeltaSnapshotData) buildNodeInfoList() []*schedulerframework.NodeInfo {
	baseList := data.baseData.getNodeInfoList()
	totalLen := len(baseList) + len(data.addedNodeInfoMap)
	var nodeInfoList []*schedulerframework.NodeInfo

	if len(data.deletedNodeInfos) > 0 || len(data.modifiedNodeInfoMap) > 0 {
		nodeInfoList = make([]*schedulerframework.NodeInfo, 0, totalLen)
		for _, bni := range baseList {
			if data.deletedNodeInfos[bni.Node().Name] {
				continue
			}
			if mni, found := data.modifiedNodeInfoMap[bni.Node().Name]; found {
				nodeInfoList = append(nodeInfoList, mni)
				continue
			}
			nodeInfoList = append(nodeInfoList, bni)
		}
	} else {
		nodeInfoList = make([]*schedulerframework.NodeInfo, len(baseList), totalLen)
		copy(nodeInfoList, baseList)
	}

	for _, ani := range data.addedNodeInfoMap {
		nodeInfoList = append(nodeInfoList, ani)
	}

	return nodeInfoList
}

// Convenience method to avoid writing loop for adding nodes.
func (data *internalDeltaSnapshotData) addNodes(nodes []*apiv1.Node) error {
	for _, node := range nodes {
		if err := data.addNode(node); err != nil {
			return err
		}
	}
	return nil
}

func (data *internalDeltaSnapshotData) addNode(node *apiv1.Node) error {
	nodeInfo := schedulerframework.NewNodeInfo()
	if err := nodeInfo.SetNode(node); err != nil {
		return fmt.Errorf("cannot set node in NodeInfo: %v", err)
	}
	return data.addNodeInfo(nodeInfo)
}

func (data *internalDeltaSnapshotData) addNodeInfo(nodeInfo *schedulerframework.NodeInfo) error {
	if _, found := data.getNodeInfo(nodeInfo.Node().Name); found {
		return fmt.Errorf("node %s already in snapshot", nodeInfo.Node().Name)
	}

	if _, found := data.deletedNodeInfos[nodeInfo.Node().Name]; found {
		delete(data.deletedNodeInfos, nodeInfo.Node().Name)
		data.modifiedNodeInfoMap[nodeInfo.Node().Name] = nodeInfo
	} else {
		data.addedNodeInfoMap[nodeInfo.Node().Name] = nodeInfo
	}

	if data.nodeInfoList != nil {
		data.nodeInfoList = append(data.nodeInfoList, nodeInfo)
	}

	if len(nodeInfo.Pods) > 0 {
		data.clearPodCaches()
	}

	return nil
}

func (data *internalDeltaSnapshotData) clearCaches() {
	data.nodeInfoList = nil
	data.clearPodCaches()
}

func (data *internalDeltaSnapshotData) clearPodCaches() {
	data.havePodsWithAffinity = nil
	data.havePodsWithRequiredAntiAffinity = nil
}

func (data *internalDeltaSnapshotData) removeNode(nodeName string) error {
	_, foundInDelta := data.addedNodeInfoMap[nodeName]
	if foundInDelta {
		// If node was added within this delta, delete this change.
		delete(data.addedNodeInfoMap, nodeName)
	}

	if _, modified := data.modifiedNodeInfoMap[nodeName]; modified {
		// If node was modified within this delta, delete this change.
		delete(data.modifiedNodeInfoMap, nodeName)
	}

	if _, deleted := data.deletedNodeInfos[nodeName]; deleted {
		// If node was deleted within this delta, fail with error.
		return errNodeNotFound
	}

	_, foundInBase := data.baseData.getNodeInfo(nodeName)
	if foundInBase {
		// If node was found in the underlying data, mark it as deleted in delta.
		data.deletedNodeInfos[nodeName] = true
	}

	if !foundInBase && !foundInDelta {
		// Node not found in the chain.
		return errNodeNotFound
	}

	// Maybe consider deleting from the lists instead. Maybe not.
	data.clearCaches()
	return nil
}

func (data *internalDeltaSnapshotData) nodeInfoToModify(nodeName string) (*schedulerframework.NodeInfo, bool) {
	dni, found := data.getNodeInfoLocal(nodeName)
	if !found {
		if _, found := data.deletedNodeInfos[nodeName]; found {
			return nil, false
		}
		bni, found := data.baseData.getNodeInfo(nodeName)
		if !found {
			return nil, false
		}
		dni = bni.Clone()
		data.modifiedNodeInfoMap[nodeName] = dni
		data.clearCaches()
	}
	return dni, true
}

func (data *internalDeltaSnapshotData) addPod(pod *apiv1.Pod, nodeName string) error {
	ni, found := data.nodeInfoToModify(nodeName)
	if !found {
		return errNodeNotFound
	}

	ni.AddPod(pod)

	// Maybe consider deleting from the list in the future. Maybe not.
	data.clearCaches()
	return nil
}

func (data *internalDeltaSnapshotData) removePod(namespace, name, nodeName string) error {
	// This always clones node info, even if the pod is actually missing.
	// Not sure if we mind, since removing non-existent pod
	// probably means things are very bad anyway.
	ni, found := data.nodeInfoToModify(nodeName)
	if !found {
		return errNodeNotFound
	}

	podFound := false
	for _, podInfo := range ni.Pods {
		if podInfo.Pod.Namespace == namespace && podInfo.Pod.Name == name {
			if err := ni.RemovePod(podInfo.Pod); err != nil {
				return fmt.Errorf("cannot remove pod; %v", err)
			}
			podFound = true
			break
		}
	}
	if !podFound {
		return fmt.Errorf("pod %s/%s not in snapshot", namespace, name)
	}

	// Maybe consider deleting from the list in the future. Maybe not.
	data.clearCaches()
	return nil
}

func (data *internalDeltaSnapshotData) fork() *internalDeltaSnapshotData {
	forkedData := newInternalDeltaSnapshotData()
	forkedData.baseData = data
	return forkedData
}

func (data *internalDeltaSnapshotData) commit() (*internalDeltaSnapshotData, error) {
	if data.baseData == nil {
		// do nothing... as in basic snapshot.
		return data, nil
	}
	for node := range data.deletedNodeInfos {
		if err := data.baseData.removeNode(node); err != nil {
			return nil, err
		}
	}
	for _, node := range data.modifiedNodeInfoMap {
		if err := data.baseData.removeNode(node.Node().Name); err != nil {
			return nil, err
		}
		if err := data.baseData.addNodeInfo(node); err != nil {
			return nil, err
		}
	}
	for _, node := range data.addedNodeInfoMap {
		if err := data.baseData.addNodeInfo(node); err != nil {
			return nil, err
		}
	}
	return data.baseData, nil
}

// List returns list of all node infos.
func (snapshot *deltaSnapshotNodeLister) List() ([]*schedulerframework.NodeInfo, error) {
	return snapshot.data.getNodeInfoList(), nil
}

// HavePodsWithAffinityList returns list of all node infos with pods that have affinity constrints.
func (snapshot *deltaSnapshotNodeLister) HavePodsWithAffinityList() ([]*schedulerframework.NodeInfo, error) {
	data := snapshot.data
	if data.havePodsWithAffinity != nil {
		return data.havePodsWithAffinity, nil
	}

	nodeInfoList := snapshot.data.getNodeInfoList()
	havePodsWithAffinityList := make([]*schedulerframework.NodeInfo, 0, len(nodeInfoList))
	for _, node := range nodeInfoList {
		if len(node.PodsWithAffinity) > 0 {
			havePodsWithAffinityList = append(havePodsWithAffinityList, node)
		}
	}
	data.havePodsWithAffinity = havePodsWithAffinityList
	return data.havePodsWithAffinity, nil
}

// HavePodsWithRequiredAntiAffinityList returns the list of NodeInfos of nodes with pods with required anti-affinity terms.
func (snapshot *deltaSnapshotNodeLister) HavePodsWithRequiredAntiAffinityList() ([]*schedulerframework.NodeInfo, error) {
	data := snapshot.data
	if data.havePodsWithRequiredAntiAffinity != nil {
		return data.havePodsWithRequiredAntiAffinity, nil
	}

	nodeInfoList := snapshot.data.getNodeInfoList()
	havePodsWithRequiredAntiAffinityList := make([]*schedulerframework.NodeInfo, 0, len(nodeInfoList))
	for _, node := range nodeInfoList {
		if len(node.PodsWithRequiredAntiAffinity) > 0 {
			havePodsWithRequiredAntiAffinityList = append(havePodsWithRequiredAntiAffinityList, node)
		}
	}
	data.havePodsWithRequiredAntiAffinity = havePodsWithRequiredAntiAffinityList
	return data.havePodsWithRequiredAntiAffinity, nil
}

// Get returns node info by node name.
func (snapshot *deltaSnapshotNodeLister) Get(nodeName string) (*schedulerframework.NodeInfo, error) {
	return (*DeltaClusterSnapshot)(snapshot).getNodeInfo(nodeName)
}

func (snapshot *DeltaClusterSnapshot) getNodeInfo(nodeName string) (*schedulerframework.NodeInfo, error) {
	data := snapshot.data
	node, found := data.getNodeInfo(nodeName)
	if !found {
		return nil, errNodeNotFound
	}
	return node, nil
}

// NodeInfos returns node lister.
func (snapshot *DeltaClusterSnapshot) NodeInfos() schedulerframework.NodeInfoLister {
	return (*deltaSnapshotNodeLister)(snapshot)
}

// NewDeltaClusterSnapshot creates instances of DeltaClusterSnapshot.
func NewDeltaClusterSnapshot() *DeltaClusterSnapshot {
	snapshot := &DeltaClusterSnapshot{}
	snapshot.Clear()
	return snapshot
}

// AddNode adds node to the snapshot.
func (snapshot *DeltaClusterSnapshot) AddNode(node *apiv1.Node) error {
	return snapshot.data.addNode(node)
}

// AddNodes adds nodes in batch to the snapshot.
func (snapshot *DeltaClusterSnapshot) AddNodes(nodes []*apiv1.Node) error {
	return snapshot.data.addNodes(nodes)
}

// AddNodeWithPods adds a node and set of pods to be scheduled to this node to the snapshot.
func (snapshot *DeltaClusterSnapshot) AddNodeWithPods(node *apiv1.Node, pods []*apiv1.Pod) error {
	if err := snapshot.AddNode(node); err != nil {
		return err
	}
	for _, pod := range pods {
		if err := snapshot.AddPod(pod, node.Name); err != nil {
			return err
		}
	}
	return nil
}

// RemoveNode removes nodes (and pods scheduled to it) from the snapshot.
func (snapshot *DeltaClusterSnapshot) RemoveNode(nodeName string) error {
	return snapshot.data.removeNode(nodeName)
}

// AddPod adds pod to the snapshot and schedules it to given node.
func (snapshot *DeltaClusterSnapshot) AddPod(pod *apiv1.Pod, nodeName string) error {
	return snapshot.data.addPod(pod, nodeName)
}

// RemovePod removes pod from the snapshot.
func (snapshot *DeltaClusterSnapshot) RemovePod(namespace, podName, nodeName string) error {
	return snapshot.data.removePod(namespace, podName, nodeName)
}

// Fork creates a fork of snapshot state. All modifications can later be reverted to moment of forking via Revert()
// Forking already forked snapshot is not allowed and will result with an error.
// Time: O(1)
func (snapshot *DeltaClusterSnapshot) Fork() error {
	if snapshot.data.baseData != nil {
		return fmt.Errorf("snapshot already forked")
	}
	snapshot.data = snapshot.data.fork()
	return nil
}

// Revert reverts snapshot state to moment of forking.
// Time: O(1)
func (snapshot *DeltaClusterSnapshot) Revert() error {
	if snapshot.data.baseData != nil {
		snapshot.data = snapshot.data.baseData
	}
	return nil

}

// Commit commits changes done after forking.
// Time: O(n), where n = size of delta (number of nodes added, modified or deleted since forking)
func (snapshot *DeltaClusterSnapshot) Commit() error {
	newData, err := snapshot.data.commit()
	if err != nil {
		return err
	}
	snapshot.data = newData
	return nil
}

// Clear reset cluster snapshot to empty, unforked state
// Time: O(1)
func (snapshot *DeltaClusterSnapshot) Clear() {
	snapshot.data = newInternalDeltaSnapshotData()
}
