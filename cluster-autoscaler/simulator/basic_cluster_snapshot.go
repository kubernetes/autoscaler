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
	"k8s.io/apimachinery/pkg/labels"
	schedulerlisters "k8s.io/kubernetes/pkg/scheduler/listers"
	schedulernodeinfo "k8s.io/kubernetes/pkg/scheduler/nodeinfo"
)

// BasicClusterSnapshot is simple, reference implementation of ClusterSnapshot.
// It is inefficient. But hopefully bug-free and good for initial testing.
type BasicClusterSnapshot struct {
	baseData   *internalBasicSnapshotData
	forkedData *internalBasicSnapshotData
}

type internalBasicSnapshotData struct {
	nodeInfoMap map[string]*schedulernodeinfo.NodeInfo
}

func (data *internalBasicSnapshotData) listNodeInfos() ([]*schedulernodeinfo.NodeInfo, error) {
	nodeInfoList := make([]*schedulernodeinfo.NodeInfo, 0, len(data.nodeInfoMap))
	for _, v := range data.nodeInfoMap {
		nodeInfoList = append(nodeInfoList, v)
	}
	return nodeInfoList, nil
}

func (data *internalBasicSnapshotData) listNodeInfosThatHavePodsWithAffinityList() ([]*schedulernodeinfo.NodeInfo, error) {
	havePodsWithAffinityList := make([]*schedulernodeinfo.NodeInfo, 0, len(data.nodeInfoMap))
	for _, v := range data.nodeInfoMap {
		if len(v.PodsWithAffinity()) > 0 {
			havePodsWithAffinityList = append(havePodsWithAffinityList, v)
		}
	}

	return havePodsWithAffinityList, nil
}

func (data *internalBasicSnapshotData) getNodeInfo(nodeName string) (*schedulernodeinfo.NodeInfo, error) {
	if v, ok := data.nodeInfoMap[nodeName]; ok {
		return v, nil
	}
	return nil, fmt.Errorf("node %s not in snapshot", nodeName)
}

func (data *internalBasicSnapshotData) listPods(selector labels.Selector) ([]*apiv1.Pod, error) {
	alwaysTrue := func(p *apiv1.Pod) bool { return true }
	return data.filteredListPods(alwaysTrue, selector)
}

func (data *internalBasicSnapshotData) filteredListPods(podFilter schedulerlisters.PodFilter, selector labels.Selector) ([]*apiv1.Pod, error) {
	pods := make([]*apiv1.Pod, 0)
	for _, n := range data.nodeInfoMap {
		for _, pod := range n.Pods() {
			if podFilter(pod) && selector.Matches(labels.Set(pod.Labels)) {
				pods = append(pods, pod)
			}
		}
	}
	return pods, nil
}

func newInternalBasicSnapshotData() *internalBasicSnapshotData {
	return &internalBasicSnapshotData{
		nodeInfoMap: make(map[string]*schedulernodeinfo.NodeInfo),
	}
}

func (data *internalBasicSnapshotData) clone() *internalBasicSnapshotData {
	clonedNodeInfoMap := make(map[string]*schedulernodeinfo.NodeInfo)
	for k, v := range data.nodeInfoMap {
		clonedNodeInfoMap[k] = v.Clone()
	}
	return &internalBasicSnapshotData{
		nodeInfoMap: clonedNodeInfoMap,
	}
}

func (data *internalBasicSnapshotData) addNode(node *apiv1.Node) error {
	if _, found := data.nodeInfoMap[node.Name]; found {
		return fmt.Errorf("node %s already in snapshot", node.Name)
	}
	nodeInfo := schedulernodeinfo.NewNodeInfo()
	err := nodeInfo.SetNode(node)
	if err != nil {
		return fmt.Errorf("cannot set node in NodeInfo; %v", err)
	}
	data.nodeInfoMap[node.Name] = nodeInfo
	return nil
}

func (data *internalBasicSnapshotData) addNodes(nodes []*apiv1.Node) error {
	for _, node := range nodes {
		if err := data.addNode(node); err != nil {
			return err
		}
	}
	return nil
}

func (data *internalBasicSnapshotData) removeNode(nodeName string) error {
	if _, found := data.nodeInfoMap[nodeName]; !found {
		return fmt.Errorf("node %s not in snapshot", nodeName)
	}
	delete(data.nodeInfoMap, nodeName)
	return nil
}

func (data *internalBasicSnapshotData) addPod(pod *apiv1.Pod, nodeName string) error {
	if _, found := data.nodeInfoMap[nodeName]; !found {
		return fmt.Errorf("node %s not in snapshot", nodeName)
	}
	data.nodeInfoMap[nodeName].AddPod(pod)
	return nil
}

func (data *internalBasicSnapshotData) removePod(namespace string, podName string) error {
	for _, nodeInfo := range data.nodeInfoMap {
		for _, pod := range nodeInfo.Pods() {
			if pod.Namespace == namespace && pod.Name == podName {
				err := nodeInfo.RemovePod(pod)
				if err != nil {
					return fmt.Errorf("cannot remove pod; %v", err)
				}
				return nil
			}
		}
	}
	return fmt.Errorf("pod %s/%s not in snapshot", namespace, podName)
}

// NewBasicClusterSnapshot creates instances of BasicClusterSnapshot.
func NewBasicClusterSnapshot() *BasicClusterSnapshot {
	snapshot := &BasicClusterSnapshot{}
	snapshot.Clear()
	return snapshot
}

func (snapshot *BasicClusterSnapshot) getInternalData() *internalBasicSnapshotData {
	if snapshot.forkedData != nil {
		return snapshot.forkedData
	}
	return snapshot.baseData
}

// AddNode adds node to the snapshot.
func (snapshot *BasicClusterSnapshot) AddNode(node *apiv1.Node) error {
	return snapshot.getInternalData().addNode(node)
}

// AddNodes adds nodes in batch to the snapshot.
func (snapshot *BasicClusterSnapshot) AddNodes(nodes []*apiv1.Node) error {
	return snapshot.getInternalData().addNodes(nodes)
}

// AddNodeWithPods adds a node and set of pods to be scheduled to this node to the snapshot.
func (snapshot *BasicClusterSnapshot) AddNodeWithPods(node *apiv1.Node, pods []*apiv1.Pod) error {
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
func (snapshot *BasicClusterSnapshot) RemoveNode(nodeName string) error {
	return snapshot.getInternalData().removeNode(nodeName)
}

// AddPod adds pod to the snapshot and schedules it to given node.
func (snapshot *BasicClusterSnapshot) AddPod(pod *apiv1.Pod, nodeName string) error {
	return snapshot.getInternalData().addPod(pod, nodeName)
}

// RemovePod removes pod from the snapshot.
func (snapshot *BasicClusterSnapshot) RemovePod(namespace string, podName string) error {
	return snapshot.getInternalData().removePod(namespace, podName)
}

// Fork creates a fork of snapshot state. All modifications can later be reverted to moment of forking via Revert()
// Forking already forked snapshot is not allowed and will result with an error.
func (snapshot *BasicClusterSnapshot) Fork() error {
	if snapshot.forkedData != nil {
		return fmt.Errorf("snapshot already forked")
	}
	snapshot.forkedData = snapshot.baseData.clone()
	return nil
}

// Revert reverts snapshot state to moment of forking.
func (snapshot *BasicClusterSnapshot) Revert() error {
	snapshot.forkedData = nil
	return nil
}

// Commit commits changes done after forking.
func (snapshot *BasicClusterSnapshot) Commit() error {
	if snapshot.forkedData == nil {
		// do nothing
		return nil
	}
	snapshot.baseData = snapshot.forkedData
	snapshot.forkedData = nil
	return nil
}

// Clear reset cluster snapshot to empty, unforked state
func (snapshot *BasicClusterSnapshot) Clear() {
	snapshot.baseData = newInternalBasicSnapshotData()
	snapshot.forkedData = nil
}

// implementation of SharedLister interface

type basicClusterSnapshotNodeLister BasicClusterSnapshot
type basicClusterSnapshotPodLister BasicClusterSnapshot

// Pods exposes snapshot as PodLister
func (snapshot *BasicClusterSnapshot) Pods() schedulerlisters.PodLister {
	return (*basicClusterSnapshotPodLister)(snapshot)
}

// List returns the list of pods in the snapshot.
func (snapshot *basicClusterSnapshotPodLister) List(selector labels.Selector) ([]*apiv1.Pod, error) {
	return (*BasicClusterSnapshot)(snapshot).getInternalData().listPods(selector)
}

// FilteredList returns a filtered list of pods in the snapshot.
func (snapshot *basicClusterSnapshotPodLister) FilteredList(podFilter schedulerlisters.PodFilter, selector labels.Selector) ([]*apiv1.Pod, error) {
	return (*BasicClusterSnapshot)(snapshot).getInternalData().filteredListPods(podFilter, selector)
}

// NodeInfos exposes snapshot as NodeInfoLister.
func (snapshot *BasicClusterSnapshot) NodeInfos() schedulerlisters.NodeInfoLister {
	return (*basicClusterSnapshotNodeLister)(snapshot)
}

// List returns the list of nodes in the snapshot.
func (snapshot *basicClusterSnapshotNodeLister) List() ([]*schedulernodeinfo.NodeInfo, error) {
	return (*BasicClusterSnapshot)(snapshot).getInternalData().listNodeInfos()
}

// HavePodsWithAffinityList returns the list of nodes with at least one pods with inter-pod affinity
func (snapshot *basicClusterSnapshotNodeLister) HavePodsWithAffinityList() ([]*schedulernodeinfo.NodeInfo, error) {
	return (*BasicClusterSnapshot)(snapshot).getInternalData().listNodeInfosThatHavePodsWithAffinityList()
}

// Returns the NodeInfo of the given node name.
func (snapshot *basicClusterSnapshotNodeLister) Get(nodeName string) (*schedulernodeinfo.NodeInfo, error) {
	return (*BasicClusterSnapshot)(snapshot).getInternalData().getNodeInfo(nodeName)
}
