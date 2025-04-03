/*
Copyright 2024 The Kubernetes Authors.

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

package store

import (
	"fmt"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
	drasnapshot "k8s.io/autoscaler/cluster-autoscaler/simulator/dynamicresources/snapshot"
	"k8s.io/klog/v2"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
)

// BasicSnapshotStore is simple, reference implementation of ClusterSnapshotStore.
// It is inefficient. But hopefully bug-free and good for initial testing.
type BasicSnapshotStore struct {
	data []*internalBasicSnapshotData
}

type internalBasicSnapshotData struct {
	nodeInfoMap        map[string]*schedulerframework.NodeInfo
	pvcNamespacePodMap map[string]map[string]bool
	draSnapshot        drasnapshot.Snapshot
}

func (data *internalBasicSnapshotData) listNodeInfos() []*schedulerframework.NodeInfo {
	nodeInfoList := make([]*schedulerframework.NodeInfo, 0, len(data.nodeInfoMap))
	for _, v := range data.nodeInfoMap {
		nodeInfoList = append(nodeInfoList, v)
	}
	return nodeInfoList
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
	return nil, clustersnapshot.ErrNodeNotFound
}

func (data *internalBasicSnapshotData) isPVCUsedByPods(key string) bool {
	if v, found := data.pvcNamespacePodMap[key]; found && v != nil && len(v) > 0 {
		return true
	}
	return false
}

func (data *internalBasicSnapshotData) addPvcUsedByPod(pod *apiv1.Pod) {
	if pod == nil {
		return
	}
	nameSpace := pod.GetNamespace()
	for _, volume := range pod.Spec.Volumes {
		if volume.PersistentVolumeClaim == nil {
			continue
		}
		k := schedulerframework.GetNamespacedName(nameSpace, volume.PersistentVolumeClaim.ClaimName)
		_, found := data.pvcNamespacePodMap[k]
		if !found {
			data.pvcNamespacePodMap[k] = make(map[string]bool)
		}
		data.pvcNamespacePodMap[k][pod.GetName()] = true
	}
}

func (data *internalBasicSnapshotData) removePvcUsedByPod(pod *apiv1.Pod) {
	if pod == nil {
		return
	}

	nameSpace := pod.GetNamespace()
	for _, volume := range pod.Spec.Volumes {
		if volume.PersistentVolumeClaim == nil {
			continue
		}
		k := schedulerframework.GetNamespacedName(nameSpace, volume.PersistentVolumeClaim.ClaimName)
		if _, found := data.pvcNamespacePodMap[k]; found {
			delete(data.pvcNamespacePodMap[k], pod.GetName())
			if len(data.pvcNamespacePodMap[k]) == 0 {
				delete(data.pvcNamespacePodMap, k)
			}
		}
	}
}

func newInternalBasicSnapshotData() *internalBasicSnapshotData {
	return &internalBasicSnapshotData{
		nodeInfoMap:        make(map[string]*schedulerframework.NodeInfo),
		pvcNamespacePodMap: make(map[string]map[string]bool),
	}
}

func (data *internalBasicSnapshotData) clone() *internalBasicSnapshotData {
	clonedNodeInfoMap := make(map[string]*schedulerframework.NodeInfo)
	for k, v := range data.nodeInfoMap {
		clonedNodeInfoMap[k] = v.Snapshot()
	}
	clonedPvcNamespaceNodeMap := make(map[string]map[string]bool)
	for k, v := range data.pvcNamespacePodMap {
		clonedPvcNamespaceNodeMap[k] = make(map[string]bool)
		for k1, v1 := range v {
			clonedPvcNamespaceNodeMap[k][k1] = v1
		}
	}
	return &internalBasicSnapshotData{
		nodeInfoMap:        clonedNodeInfoMap,
		pvcNamespacePodMap: clonedPvcNamespaceNodeMap,
		draSnapshot:        data.draSnapshot.Clone(),
	}
}

func (data *internalBasicSnapshotData) addNode(node *apiv1.Node) error {
	if _, found := data.nodeInfoMap[node.Name]; found {
		return fmt.Errorf("node %s already in snapshot", node.Name)
	}
	nodeInfo := schedulerframework.NewNodeInfo()
	nodeInfo.SetNode(node)
	data.nodeInfoMap[node.Name] = nodeInfo
	return nil
}

func (data *internalBasicSnapshotData) removeNodeInfo(nodeName string) error {
	if _, found := data.nodeInfoMap[nodeName]; !found {
		return clustersnapshot.ErrNodeNotFound
	}
	for _, pod := range data.nodeInfoMap[nodeName].Pods {
		data.removePvcUsedByPod(pod.Pod)
	}
	delete(data.nodeInfoMap, nodeName)
	return nil
}

func (data *internalBasicSnapshotData) addPod(pod *apiv1.Pod, nodeName string) error {
	if _, found := data.nodeInfoMap[nodeName]; !found {
		return clustersnapshot.ErrNodeNotFound
	}
	data.nodeInfoMap[nodeName].AddPod(pod)
	data.addPvcUsedByPod(pod)
	return nil
}

func (data *internalBasicSnapshotData) removePod(namespace, podName, nodeName string) error {
	nodeInfo, found := data.nodeInfoMap[nodeName]
	if !found {
		return clustersnapshot.ErrNodeNotFound
	}
	logger := klog.Background()
	for _, podInfo := range nodeInfo.Pods {
		if podInfo.Pod.Namespace == namespace && podInfo.Pod.Name == podName {
			data.removePvcUsedByPod(podInfo.Pod)
			err := nodeInfo.RemovePod(logger, podInfo.Pod)
			if err != nil {
				data.addPvcUsedByPod(podInfo.Pod)
				return fmt.Errorf("cannot remove pod; %v", err)
			}
			return nil
		}
	}
	return fmt.Errorf("pod %s/%s not in snapshot", namespace, podName)
}

// NewBasicSnapshotStore creates instances of BasicSnapshotStore.
func NewBasicSnapshotStore() *BasicSnapshotStore {
	snapshot := &BasicSnapshotStore{}
	snapshot.clear()
	return snapshot
}

func (snapshot *BasicSnapshotStore) getInternalData() *internalBasicSnapshotData {
	return snapshot.data[len(snapshot.data)-1]
}

// DraSnapshot returns the DRA snapshot.
func (snapshot *BasicSnapshotStore) DraSnapshot() drasnapshot.Snapshot {
	return snapshot.getInternalData().draSnapshot
}

// AddSchedulerNodeInfo adds a NodeInfo.
func (snapshot *BasicSnapshotStore) AddSchedulerNodeInfo(nodeInfo *schedulerframework.NodeInfo) error {
	if err := snapshot.getInternalData().addNode(nodeInfo.Node()); err != nil {
		return err
	}
	for _, podInfo := range nodeInfo.Pods {
		if err := snapshot.getInternalData().addPod(podInfo.Pod, nodeInfo.Node().Name); err != nil {
			return err
		}
	}
	return nil
}

// SetClusterState sets the cluster state.
func (snapshot *BasicSnapshotStore) SetClusterState(nodes []*apiv1.Node, scheduledPods []*apiv1.Pod, draSnapshot drasnapshot.Snapshot) error {
	snapshot.clear()

	knownNodes := make(map[string]bool)
	for _, node := range nodes {
		if err := snapshot.getInternalData().addNode(node); err != nil {
			return err
		}
		knownNodes[node.Name] = true
	}
	for _, pod := range scheduledPods {
		if knownNodes[pod.Spec.NodeName] {
			if err := snapshot.getInternalData().addPod(pod, pod.Spec.NodeName); err != nil {
				return err
			}
		}
	}
	snapshot.getInternalData().draSnapshot = draSnapshot
	return nil
}

// RemoveSchedulerNodeInfo removes nodes (and pods scheduled to it) from the snapshot.
func (snapshot *BasicSnapshotStore) RemoveSchedulerNodeInfo(nodeName string) error {
	return snapshot.getInternalData().removeNodeInfo(nodeName)
}

// ForceAddPod adds pod to the snapshot and schedules it to given node.
func (snapshot *BasicSnapshotStore) ForceAddPod(pod *apiv1.Pod, nodeName string) error {
	return snapshot.getInternalData().addPod(pod, nodeName)
}

// ForceRemovePod removes pod from the snapshot.
func (snapshot *BasicSnapshotStore) ForceRemovePod(namespace, podName, nodeName string) error {
	return snapshot.getInternalData().removePod(namespace, podName, nodeName)
}

// IsPVCUsedByPods returns if the pvc is used by any pod
func (snapshot *BasicSnapshotStore) IsPVCUsedByPods(key string) bool {
	return snapshot.getInternalData().isPVCUsedByPods(key)
}

// Fork creates a fork of snapshot state. All modifications can later be reverted to moment of forking via Revert()
func (snapshot *BasicSnapshotStore) Fork() {
	forkData := snapshot.getInternalData().clone()
	snapshot.data = append(snapshot.data, forkData)
}

// Revert reverts snapshot state to moment of forking.
func (snapshot *BasicSnapshotStore) Revert() {
	if len(snapshot.data) == 1 {
		return
	}
	snapshot.data = snapshot.data[:len(snapshot.data)-1]
}

// Commit commits changes done after forking.
func (snapshot *BasicSnapshotStore) Commit() error {
	if len(snapshot.data) <= 1 {
		// do nothing
		return nil
	}
	snapshot.data = append(snapshot.data[:len(snapshot.data)-2], snapshot.data[len(snapshot.data)-1])
	return nil
}

// clear reset cluster snapshot to empty, unforked state
func (snapshot *BasicSnapshotStore) clear() {
	baseData := newInternalBasicSnapshotData()
	snapshot.data = []*internalBasicSnapshotData{baseData}
}

// implementation of SharedLister interface

type basicSnapshotStoreNodeLister BasicSnapshotStore
type basicSnapshotStoreStorageLister BasicSnapshotStore

// NodeInfos exposes snapshot as NodeInfoLister.
func (snapshot *BasicSnapshotStore) NodeInfos() schedulerframework.NodeInfoLister {
	return (*basicSnapshotStoreNodeLister)(snapshot)
}

// StorageInfos exposes snapshot as StorageInfoLister.
func (snapshot *BasicSnapshotStore) StorageInfos() schedulerframework.StorageInfoLister {
	return (*basicSnapshotStoreStorageLister)(snapshot)
}

// ResourceClaims exposes snapshot as ResourceClaimTracker
func (snapshot *BasicSnapshotStore) ResourceClaims() schedulerframework.ResourceClaimTracker {
	return snapshot.DraSnapshot().ResourceClaims()
}

// ResourceSlices exposes snapshot as ResourceSliceLister.
func (snapshot *BasicSnapshotStore) ResourceSlices() schedulerframework.ResourceSliceLister {
	return snapshot.DraSnapshot().ResourceSlices()
}

// DeviceClasses exposes the snapshot as DeviceClassLister.
func (snapshot *BasicSnapshotStore) DeviceClasses() schedulerframework.DeviceClassLister {
	return snapshot.DraSnapshot().DeviceClasses()
}

// List returns the list of nodes in the snapshot.
func (snapshot *basicSnapshotStoreNodeLister) List() ([]*schedulerframework.NodeInfo, error) {
	return (*BasicSnapshotStore)(snapshot).getInternalData().listNodeInfos(), nil
}

// HavePodsWithAffinityList returns the list of nodes with at least one pods with inter-pod affinity
func (snapshot *basicSnapshotStoreNodeLister) HavePodsWithAffinityList() ([]*schedulerframework.NodeInfo, error) {
	return (*BasicSnapshotStore)(snapshot).getInternalData().listNodeInfosThatHavePodsWithAffinityList()
}

// HavePodsWithRequiredAntiAffinityList returns the list of NodeInfos of nodes with pods with required anti-affinity terms.
func (snapshot *basicSnapshotStoreNodeLister) HavePodsWithRequiredAntiAffinityList() ([]*schedulerframework.NodeInfo, error) {
	return (*BasicSnapshotStore)(snapshot).getInternalData().listNodeInfosThatHavePodsWithRequiredAntiAffinityList()
}

// Returns the NodeInfo of the given node name.
func (snapshot *basicSnapshotStoreNodeLister) Get(nodeName string) (*schedulerframework.NodeInfo, error) {
	return (*BasicSnapshotStore)(snapshot).getInternalData().getNodeInfo(nodeName)
}

// Returns the IsPVCUsedByPods in a given key.
func (snapshot *basicSnapshotStoreStorageLister) IsPVCUsedByPods(key string) bool {
	return (*BasicSnapshotStore)(snapshot).getInternalData().isPVCUsedByPods(key)
}
