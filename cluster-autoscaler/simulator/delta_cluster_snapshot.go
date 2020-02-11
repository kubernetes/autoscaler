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
	"errors"
	"fmt"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	schedulerlisters "k8s.io/kubernetes/pkg/scheduler/listers"
	schedulernodeinfo "k8s.io/kubernetes/pkg/scheduler/nodeinfo"
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
type deltaSnapshotPodLister DeltaClusterSnapshot

type internalDeltaSnapshotData struct {
	baseData *internalDeltaSnapshotData

	addedNodeInfoMap    map[string]*schedulernodeinfo.NodeInfo
	modifiedNodeInfoMap map[string]*schedulernodeinfo.NodeInfo
	deletedNodeInfos    map[string]bool

	nodeInfoList         []*schedulernodeinfo.NodeInfo
	podList              []*apiv1.Pod
	havePodsWithAffinity []*schedulernodeinfo.NodeInfo
}

var errNodeNotFound = errors.New("node not found")

func newInternalDeltaSnapshotData() *internalDeltaSnapshotData {
	return &internalDeltaSnapshotData{
		addedNodeInfoMap:    make(map[string]*schedulernodeinfo.NodeInfo),
		modifiedNodeInfoMap: make(map[string]*schedulernodeinfo.NodeInfo),
		deletedNodeInfos:    make(map[string]bool),
	}
}

func (data *internalDeltaSnapshotData) getNodeInfo(name string) (*schedulernodeinfo.NodeInfo, bool) {
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

func (data *internalDeltaSnapshotData) getNodeInfoLocal(name string) (*schedulernodeinfo.NodeInfo, bool) {
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

func (data *internalDeltaSnapshotData) getNodeInfoList() []*schedulernodeinfo.NodeInfo {
	if data == nil {
		return nil
	}
	if data.nodeInfoList == nil {
		data.nodeInfoList = data.buildNodeInfoList()
	}
	return data.nodeInfoList
}

// Contains costly copying throughout the struct chain. Use wisely.
func (data *internalDeltaSnapshotData) buildNodeInfoList() []*schedulernodeinfo.NodeInfo {
	baseList := data.baseData.getNodeInfoList()
	totalLen := len(baseList) + len(data.addedNodeInfoMap)
	var nodeInfoList []*schedulernodeinfo.NodeInfo

	if len(data.deletedNodeInfos) > 0 || len(data.modifiedNodeInfoMap) > 0 {
		nodeInfoList = make([]*schedulernodeinfo.NodeInfo, 0, totalLen+100)
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
		nodeInfoList = make([]*schedulernodeinfo.NodeInfo, len(baseList), totalLen+100)
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
	nodeInfo := schedulernodeinfo.NewNodeInfo()
	if err := nodeInfo.SetNode(node); err != nil {
		return fmt.Errorf("cannot set node in NodeInfo: %v", err)
	}
	return data.addNodeInfo(nodeInfo)
}

func (data *internalDeltaSnapshotData) addNodeInfo(nodeInfo *schedulernodeinfo.NodeInfo) error {
	if _, found := data.getNodeInfo(nodeInfo.Node().Name); found {
		return fmt.Errorf("node %s already in snapshot", nodeInfo.Node().Name)
	}
	data.addedNodeInfoMap[nodeInfo.Node().Name] = nodeInfo
	if data.nodeInfoList != nil {
		data.nodeInfoList = append(data.nodeInfoList, nodeInfo)
	}
	return nil
}

func (data *internalDeltaSnapshotData) clearCaches() {
	data.nodeInfoList = nil
	data.clearPodCaches()
}

func (data *internalDeltaSnapshotData) clearPodCaches() {
	data.podList = nil
	data.havePodsWithAffinity = nil
}

func (data *internalDeltaSnapshotData) updateNode(node *schedulernodeinfo.NodeInfo) error {
	if _, found := data.addedNodeInfoMap[node.Node().Name]; found {
		data.removeNode(node.Node().Name)
	}
	if _, found := data.modifiedNodeInfoMap[node.Node().Name]; found {
		data.removeNode(node.Node().Name)
	}

	return data.addNodeInfo(node)
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

func (data *internalDeltaSnapshotData) addPod(pod *apiv1.Pod, nodeName string) error {
	dni, found := data.getNodeInfoLocal(nodeName)
	if !found {
		bni, found := data.baseData.getNodeInfo(nodeName)
		if !found {
			return errNodeNotFound
		}
		dni = bni.Clone()
		data.modifiedNodeInfoMap[nodeName] = dni
		data.clearCaches()
	}

	dni.AddPod(pod)
	if data.podList != nil || data.havePodsWithAffinity != nil {
		data.clearPodCaches()
	}
	return nil
}

func (data *internalDeltaSnapshotData) removePod(namespace, name, nodeName string) error {
	dni, found := data.getNodeInfoLocal(nodeName)
	if !found {
		bni, found := data.baseData.getNodeInfo(nodeName)
		if !found {
			return errNodeNotFound
		}
		dni = bni.Clone()
		data.modifiedNodeInfoMap[nodeName] = dni
		data.clearCaches()
	}

	podFound := false
	for _, pod := range dni.Pods() {
		if pod.Namespace == namespace && pod.Name == name {
			if err := dni.RemovePod(pod); err != nil {
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

func (data *internalDeltaSnapshotData) getPodList() []*apiv1.Pod {
	if data == nil {
		return nil
	}
	if data.podList == nil {
		data.podList = data.buildPodList()
	}
	return data.podList
}

func (data *internalDeltaSnapshotData) buildPodList() []*apiv1.Pod {
	podList := []*apiv1.Pod{}
	if len(data.deletedNodeInfos) > 0 || len(data.modifiedNodeInfoMap) > 0 {
		nodeInfos := data.getNodeInfoList()
		for _, ni := range nodeInfos {
			podList = append(podList, ni.Pods()...)
		}
	} else {
		basePodList := data.baseData.getPodList()
		copy(podList, basePodList)
		for _, ni := range data.addedNodeInfoMap {
			podList = append(podList, ni.Pods()...)
		}
	}
	return podList
}

func (data *internalDeltaSnapshotData) fork() *internalDeltaSnapshotData {
	forkedData := newInternalDeltaSnapshotData()
	forkedData.baseData = data
	return forkedData
}

func (data *internalDeltaSnapshotData) commit() *internalDeltaSnapshotData {
	for _, node := range data.modifiedNodeInfoMap {
		data.baseData.updateNode(node)
	}
	for _, node := range data.addedNodeInfoMap {
		data.baseData.addNodeInfo(node)
	}
	for node := range data.deletedNodeInfos {
		data.baseData.removeNode(node)
	}
	return data.baseData
}

// List returns list of all node infos.
func (snapshot *deltaSnapshotNodeLister) List() ([]*schedulernodeinfo.NodeInfo, error) {
	return snapshot.data.getNodeInfoList(), nil
}

// HavePodsWithAffinityList returns list of all node infos with pods that have affinity constrints.
func (snapshot *deltaSnapshotNodeLister) HavePodsWithAffinityList() ([]*schedulernodeinfo.NodeInfo, error) {
	data := snapshot.data
	if data.havePodsWithAffinity != nil {
		return data.havePodsWithAffinity, nil
	}

	nodeInfoList := snapshot.data.getNodeInfoList()
	havePodsWithAffinityList := make([]*schedulernodeinfo.NodeInfo, 0, len(nodeInfoList))
	for _, node := range nodeInfoList {
		if len(node.PodsWithAffinity()) > 0 {
			havePodsWithAffinityList = append(havePodsWithAffinityList, node)
		}
	}
	data.havePodsWithAffinity = havePodsWithAffinityList
	return data.havePodsWithAffinity, nil
}

// Get returns node info by node name.
func (snapshot *deltaSnapshotNodeLister) Get(nodeName string) (*schedulernodeinfo.NodeInfo, error) {
	return (*DeltaClusterSnapshot)(snapshot).getNodeInfo(nodeName)
}

func (snapshot *DeltaClusterSnapshot) getNodeInfo(nodeName string) (*schedulernodeinfo.NodeInfo, error) {
	data := snapshot.data
	node, found := data.getNodeInfo(nodeName)
	if !found {
		return nil, errNodeNotFound
	}
	return node, nil
}

// List returns all pods matching selector.
func (snapshot *deltaSnapshotPodLister) List(selector labels.Selector) ([]*apiv1.Pod, error) {
	data := snapshot.data
	if data.podList == nil {
		data.podList = data.buildPodList()
	}

	if selector.Empty() {
		// no restrictions, yay
		return data.podList, nil
	}

	selectedPods := make([]*apiv1.Pod, 0, len(data.podList))
	for _, pod := range data.podList {
		if selector.Matches(labels.Set(pod.Labels)) {
			selectedPods = append(selectedPods, pod)
		}
	}
	return selectedPods, nil
}

// FilteredList returns all pods matching selector and filter.
func (snapshot *deltaSnapshotPodLister) FilteredList(podFilter schedulerlisters.PodFilter, selector labels.Selector) ([]*apiv1.Pod, error) {
	data := snapshot.data
	if data.podList == nil {
		data.podList = data.buildPodList()
	}

	selectedPods := make([]*apiv1.Pod, 0, len(data.podList))
	for _, pod := range data.podList {
		if podFilter(pod) && selector.Matches(labels.Set(pod.Labels)) {
			selectedPods = append(selectedPods, pod)
		}
	}
	return selectedPods, nil
}

// Pods returns pod lister.
func (snapshot *DeltaClusterSnapshot) Pods() schedulerlisters.PodLister {
	return (*deltaSnapshotPodLister)(snapshot)
}

// NodeInfos returns node lister.
func (snapshot *DeltaClusterSnapshot) NodeInfos() schedulerlisters.NodeInfoLister {
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
	if snapshot.data.baseData == nil {
		return fmt.Errorf("snapshot not forked")
	}
	snapshot.data = snapshot.data.baseData
	return nil

}

// Commit commits changes done after forking.
// Time: O(n), where n = size of delta (number of nodes added, modified or deleted since forking)
func (snapshot *DeltaClusterSnapshot) Commit() error {
	snapshot.data = snapshot.data.commit()
	return nil
}

// Clear reset cluster snapshot to empty, unforked state
// Time: O(1)
func (snapshot *DeltaClusterSnapshot) Clear() {
	snapshot.data = newInternalDeltaSnapshotData()
}
