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
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework/v1alpha1"
)

// BasicClusterSnapshot is simple, reference implementation of ClusterSnapshot.
// It is inefficient. But hopefully bug-free and good for initial testing.
type BasicClusterSnapshot struct {
	baseData   *internalBasicSnapshotData
	forkedData *internalBasicSnapshotData
}

type internalBasicSnapshotData struct {
	nodeInfoMap map[string]*schedulerframework.NodeInfo
}

func (data *internalBasicSnapshotData) listNodeInfos() ([]*schedulerframework.NodeInfo, error) {
	nodeInfoList := make([]*schedulerframework.NodeInfo, 0, len(data.nodeInfoMap))
	for _, v := range data.nodeInfoMap {
		nodeInfoList = append(nodeInfoList, v)
	}
	return nodeInfoList, nil
}

func (data *internalBasicSnapshotData) listNodeInfosThatHavePodsWithAffinityList() ([]*schedulerframework.NodeInfo, error) {
	havePodsWithAffinityList := make([]*schedulerframework.NodeInfo, 0, len(data.nodeInfoMap))
	for _, v := range data.nodeInfoMap {
		if len(v.PodsWithAffinity) > 0 {
			havePodsWithAffinityList = append(havePodsWithAffinityList, v)
		}
	}

	return havePodsWithAffinityList, nil
}

func (data *internalBasicSnapshotData) listNodeInfosThatHavePodsWithRequiredAntiAffinityList() ([]*schedulerframework.NodeInfo, error) {
	havePodsWithRequiredAntiAffinityList := make([]*schedulerframework.NodeInfo, 0, len(data.nodeInfoMap))
	for _, v := range data.nodeInfoMap {
		if len(v.PodsWithRequiredAntiAffinity) > 0 {
			havePodsWithRequiredAntiAffinityList = append(havePodsWithRequiredAntiAffinityList, v)
		}
	}

	return havePodsWithRequiredAntiAffinityList, nil
}

func (data *internalBasicSnapshotData) getNodeInfo(nodeName string) (*schedulerframework.NodeInfo, error) {
	if v, ok := data.nodeInfoMap[nodeName]; ok {
		return v, nil
	}
	return nil, errNodeNotFound
}

func newInternalBasicSnapshotData() *internalBasicSnapshotData {
	return &internalBasicSnapshotData{
		nodeInfoMap: make(map[string]*schedulerframework.NodeInfo),
	}
}

func (data *internalBasicSnapshotData) clone() *internalBasicSnapshotData {
	clonedNodeInfoMap := make(map[string]*schedulerframework.NodeInfo)
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
	nodeInfo := schedulerframework.NewNodeInfo()
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
		return errNodeNotFound
	}
	delete(data.nodeInfoMap, nodeName)
	return nil
}

func (data *internalBasicSnapshotData) addPod(pod *apiv1.Pod, nodeName string) error {
	if _, found := data.nodeInfoMap[nodeName]; !found {
		return errNodeNotFound
	}
	data.nodeInfoMap[nodeName].AddPod(pod)
	return nil
}

func (data *internalBasicSnapshotData) removePod(namespace, podName, nodeName string) error {
	nodeInfo, found := data.nodeInfoMap[nodeName]
	if !found {
		return errNodeNotFound
	}
	for _, podInfo := range nodeInfo.Pods {
		if podInfo.Pod.Namespace == namespace && podInfo.Pod.Name == podName {
			err := nodeInfo.RemovePod(podInfo.Pod)
			if err != nil {
				return fmt.Errorf("cannot remove pod; %v", err)
			}
			return nil
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
func (snapshot *BasicClusterSnapshot) RemovePod(namespace, podName, nodeName string) error {
	return snapshot.getInternalData().removePod(namespace, podName, nodeName)
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

// NodeInfos exposes snapshot as NodeInfoLister.
func (snapshot *BasicClusterSnapshot) NodeInfos() schedulerframework.NodeInfoLister {
	return (*basicClusterSnapshotNodeLister)(snapshot)
}

// List returns the list of nodes in the snapshot.
func (snapshot *basicClusterSnapshotNodeLister) List() ([]*schedulerframework.NodeInfo, error) {
	return (*BasicClusterSnapshot)(snapshot).getInternalData().listNodeInfos()
}

// HavePodsWithAffinityList returns the list of nodes with at least one pods with inter-pod affinity
func (snapshot *basicClusterSnapshotNodeLister) HavePodsWithAffinityList() ([]*schedulerframework.NodeInfo, error) {
	return (*BasicClusterSnapshot)(snapshot).getInternalData().listNodeInfosThatHavePodsWithAffinityList()
}

// HavePodsWithRequiredAntiAffinityList returns the list of NodeInfos of nodes with pods with required anti-affinity terms.
func (snapshot *basicClusterSnapshotNodeLister) HavePodsWithRequiredAntiAffinityList() ([]*schedulerframework.NodeInfo, error) {
	return (*BasicClusterSnapshot)(snapshot).getInternalData().listNodeInfosThatHavePodsWithRequiredAntiAffinityList()
}

// Returns the NodeInfo of the given node name.
func (snapshot *basicClusterSnapshotNodeLister) Get(nodeName string) (*schedulerframework.NodeInfo, error) {
	return (*BasicClusterSnapshot)(snapshot).getInternalData().getNodeInfo(nodeName)
}
