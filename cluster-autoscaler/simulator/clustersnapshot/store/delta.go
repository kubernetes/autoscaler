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
	"context"
	"fmt"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
	drasnapshot "k8s.io/autoscaler/cluster-autoscaler/simulator/dynamicresources/snapshot"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
	fwk "k8s.io/kube-scheduler/framework"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
)

// DeltaSnapshotStore is an implementation of ClusterSnapshotStore optimized for typical Cluster Autoscaler usage - (fork, add stuff, revert), repeated many times per loop.
//
// Complexity of some notable operations:
//
//	fork - O(1)
//	revert - O(1)
//	commit - O(n)
//	list all pods (no filtering) - O(n), cached
//	list all pods (with filtering) - O(n)
//	list node infos - O(n), cached
//
// Watch out for:
//
// * Node deletions, pod additions & deletions - invalidates cache of current snapshot
// (when forked affects delta, but not base.)
//
// * Pod affinity - causes scheduler framework to list pods with non-empty selector,
// so basic caching doesn't help.
//
// * DRA objects are tracked in the separate snapshot and while they don't exactly share
// memory and time complexities of DeltaSnapshotStore - they are optimized for
// cluster autoscaler operations
type DeltaSnapshotStore struct {
	data        *internalDeltaSnapshotData
	draSnapshot *drasnapshot.Snapshot
	parallelism int
}

type deltaSnapshotStoreNodeLister DeltaSnapshotStore
type deltaSnapshotStoreStorageLister DeltaSnapshotStore

type internalDeltaSnapshotData struct {
	baseData *internalDeltaSnapshotData

	addedNodeInfoMap    map[string]fwk.NodeInfo
	modifiedNodeInfoMap map[string]fwk.NodeInfo
	deletedNodeInfos    map[string]bool

	nodeInfoList                     []fwk.NodeInfo
	havePodsWithAffinity             []fwk.NodeInfo
	havePodsWithRequiredAntiAffinity []fwk.NodeInfo
	pvcNamespaceMap                  map[string]int
}

func newInternalDeltaSnapshotData() *internalDeltaSnapshotData {
	return &internalDeltaSnapshotData{
		addedNodeInfoMap:    make(map[string]fwk.NodeInfo),
		modifiedNodeInfoMap: make(map[string]fwk.NodeInfo),
		deletedNodeInfos:    make(map[string]bool),
	}
}

func (data *internalDeltaSnapshotData) getNodeInfo(name string) (fwk.NodeInfo, bool) {
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

func (data *internalDeltaSnapshotData) getNodeInfoLocal(name string) (fwk.NodeInfo, bool) {
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

func (data *internalDeltaSnapshotData) getNodeInfoList() []fwk.NodeInfo {
	if data == nil {
		return nil
	}
	if data.nodeInfoList == nil {
		data.nodeInfoList = data.buildNodeInfoList()
	}
	return data.nodeInfoList
}

// Contains costly copying throughout the struct chain. Use wisely.
func (data *internalDeltaSnapshotData) buildNodeInfoList() []fwk.NodeInfo {
	baseList := data.baseData.getNodeInfoList()
	totalLen := len(baseList) + len(data.addedNodeInfoMap)
	var nodeInfoList []fwk.NodeInfo

	if len(data.deletedNodeInfos) > 0 || len(data.modifiedNodeInfoMap) > 0 {
		nodeInfoList = make([]fwk.NodeInfo, 0, totalLen)
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
		nodeInfoList = make([]fwk.NodeInfo, len(baseList), totalLen)
		copy(nodeInfoList, baseList)
	}

	for _, ani := range data.addedNodeInfoMap {
		nodeInfoList = append(nodeInfoList, ani)
	}

	return nodeInfoList
}

func (data *internalDeltaSnapshotData) addNode(node *apiv1.Node) (fwk.NodeInfo, error) {
	nodeInfo := schedulerframework.NewNodeInfo()
	nodeInfo.SetNode(node)
	err := data.addNodeInfo(nodeInfo)
	if err != nil {
		return nil, err
	}
	return nodeInfo, nil
}

func (data *internalDeltaSnapshotData) addNodeInfo(nodeInfo fwk.NodeInfo) error {
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

	if len(nodeInfo.GetPods()) > 0 {
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
	// TODO: update the cache when adding/removing pods instead of invalidating the whole cache
	data.pvcNamespaceMap = nil
}

func (data *internalDeltaSnapshotData) removeNodeInfo(nodeName string) error {
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
		return clustersnapshot.ErrNodeNotFound
	}

	_, foundInBase := data.baseData.getNodeInfo(nodeName)
	if foundInBase {
		// If node was found in the underlying data, mark it as deleted in delta.
		data.deletedNodeInfos[nodeName] = true
	}

	if !foundInBase && !foundInDelta {
		// Node not found in the chain.
		return clustersnapshot.ErrNodeNotFound
	}

	// Maybe consider deleting from the lists instead. Maybe not.
	data.clearCaches()
	return nil
}

func (data *internalDeltaSnapshotData) nodeInfoToModify(nodeName string) (fwk.NodeInfo, bool) {
	dni, found := data.getNodeInfoLocal(nodeName)
	if !found {
		if _, found := data.deletedNodeInfos[nodeName]; found {
			return nil, false
		}
		bni, found := data.baseData.getNodeInfo(nodeName)
		if !found {
			return nil, false
		}
		dni = bni.Snapshot()
		data.modifiedNodeInfoMap[nodeName] = dni
		data.clearCaches()
	}
	return dni, true
}

func (data *internalDeltaSnapshotData) addPod(pod *apiv1.Pod, nodeName string) error {
	ni, found := data.nodeInfoToModify(nodeName)
	if !found {
		return clustersnapshot.ErrNodeNotFound
	}

	podInfo, _ := schedulerframework.NewPodInfo(pod)
	ni.AddPodInfo(podInfo)

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
		return clustersnapshot.ErrNodeNotFound
	}

	podFound := false
	logger := klog.Background()
	for _, podInfo := range ni.GetPods() {
		if podInfo.GetPod().Namespace == namespace && podInfo.GetPod().Name == name {
			if err := ni.RemovePod(logger, podInfo.GetPod()); err != nil {
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

func (data *internalDeltaSnapshotData) isPVCUsedByPods(key string) bool {
	if data.pvcNamespaceMap != nil {
		return data.pvcNamespaceMap[key] > 0
	}
	nodeInfos := data.getNodeInfoList()
	pvcNamespaceMap := make(map[string]int)
	for _, v := range nodeInfos {
		for k, i := range v.GetPVCRefCounts() {
			pvcNamespaceMap[k] += i
		}
	}
	data.pvcNamespaceMap = pvcNamespaceMap
	return data.pvcNamespaceMap[key] > 0
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
		if err := data.baseData.removeNodeInfo(node); err != nil {
			return nil, err
		}
	}
	for _, node := range data.modifiedNodeInfoMap {
		if err := data.baseData.removeNodeInfo(node.Node().Name); err != nil {
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
func (snapshot *deltaSnapshotStoreNodeLister) List() ([]fwk.NodeInfo, error) {
	return snapshot.data.getNodeInfoList(), nil
}

// HavePodsWithAffinityList returns list of all node infos with pods that have affinity constrints.
func (snapshot *deltaSnapshotStoreNodeLister) HavePodsWithAffinityList() ([]fwk.NodeInfo, error) {
	data := snapshot.data
	if data.havePodsWithAffinity != nil {
		return data.havePodsWithAffinity, nil
	}

	nodeInfoList := snapshot.data.getNodeInfoList()
	havePodsWithAffinityList := make([]fwk.NodeInfo, 0, len(nodeInfoList))
	for _, node := range nodeInfoList {
		if len(node.GetPodsWithAffinity()) > 0 {
			havePodsWithAffinityList = append(havePodsWithAffinityList, node)
		}
	}
	data.havePodsWithAffinity = havePodsWithAffinityList
	return data.havePodsWithAffinity, nil
}

// HavePodsWithRequiredAntiAffinityList returns the list of NodeInfos of nodes with pods with required anti-affinity terms.
func (snapshot *deltaSnapshotStoreNodeLister) HavePodsWithRequiredAntiAffinityList() ([]fwk.NodeInfo, error) {
	data := snapshot.data
	if data.havePodsWithRequiredAntiAffinity != nil {
		return data.havePodsWithRequiredAntiAffinity, nil
	}

	nodeInfoList := snapshot.data.getNodeInfoList()
	havePodsWithRequiredAntiAffinityList := make([]fwk.NodeInfo, 0, len(nodeInfoList))
	for _, node := range nodeInfoList {
		if len(node.GetPodsWithRequiredAntiAffinity()) > 0 {
			havePodsWithRequiredAntiAffinityList = append(havePodsWithRequiredAntiAffinityList, node)
		}
	}
	data.havePodsWithRequiredAntiAffinity = havePodsWithRequiredAntiAffinityList
	return data.havePodsWithRequiredAntiAffinity, nil
}

// Get returns node info by node name.
func (snapshot *deltaSnapshotStoreNodeLister) Get(nodeName string) (fwk.NodeInfo, error) {
	return (*DeltaSnapshotStore)(snapshot).getNodeInfo(nodeName)
}

// IsPVCUsedByPods returns if PVC is used by pods
func (snapshot *deltaSnapshotStoreStorageLister) IsPVCUsedByPods(key string) bool {
	return (*DeltaSnapshotStore)(snapshot).IsPVCUsedByPods(key)
}

func (snapshot *DeltaSnapshotStore) getNodeInfo(nodeName string) (fwk.NodeInfo, error) {
	data := snapshot.data
	node, found := data.getNodeInfo(nodeName)
	if !found {
		return nil, clustersnapshot.ErrNodeNotFound
	}
	return node, nil
}

// NodeInfos returns node lister.
func (snapshot *DeltaSnapshotStore) NodeInfos() schedulerframework.NodeInfoLister {
	return (*deltaSnapshotStoreNodeLister)(snapshot)
}

// StorageInfos returns storage lister
func (snapshot *DeltaSnapshotStore) StorageInfos() schedulerframework.StorageInfoLister {
	return (*deltaSnapshotStoreStorageLister)(snapshot)
}

// ResourceClaims exposes snapshot as ResourceClaimTracker
func (snapshot *DeltaSnapshotStore) ResourceClaims() schedulerframework.ResourceClaimTracker {
	return snapshot.DraSnapshot().ResourceClaims()
}

// ResourceSlices exposes snapshot as ResourceSliceLister.
func (snapshot *DeltaSnapshotStore) ResourceSlices() schedulerframework.ResourceSliceLister {
	return snapshot.DraSnapshot().ResourceSlices()
}

// DeviceClasses exposes the snapshot as DeviceClassLister.
func (snapshot *DeltaSnapshotStore) DeviceClasses() schedulerframework.DeviceClassLister {
	return snapshot.DraSnapshot().DeviceClasses()
}

// NewDeltaSnapshotStore creates instances of DeltaSnapshotStore.
func NewDeltaSnapshotStore(parallelism int) *DeltaSnapshotStore {
	snapshot := &DeltaSnapshotStore{
		parallelism: parallelism,
	}
	snapshot.clear()
	return snapshot
}

// DraSnapshot returns the DRA snapshot.
func (snapshot *DeltaSnapshotStore) DraSnapshot() *drasnapshot.Snapshot {
	return snapshot.draSnapshot
}

// AddSchedulerNodeInfo adds a NodeInfo.
func (snapshot *DeltaSnapshotStore) AddSchedulerNodeInfo(nodeInfo fwk.NodeInfo) error {
	if _, err := snapshot.data.addNode(nodeInfo.Node()); err != nil {
		return err
	}
	for _, podInfo := range nodeInfo.GetPods() {
		if err := snapshot.data.addPod(podInfo.GetPod(), nodeInfo.Node().Name); err != nil {
			return err
		}
	}
	return nil
}

// setClusterStatePodsSequential sets the pods in cluster state in a sequential way.
func (snapshot *DeltaSnapshotStore) setClusterStatePodsSequential(nodeInfos []fwk.NodeInfo, nodeNameToIdx map[string]int, scheduledPods []*apiv1.Pod) {
	for _, pod := range scheduledPods {
		if nodeIdx, ok := nodeNameToIdx[pod.Spec.NodeName]; ok {
			// Can add pod directly. Cache will be cleared afterwards.
			podInfo, _ := schedulerframework.NewPodInfo(pod)
			nodeInfos[nodeIdx].AddPodInfo(podInfo)
		}
	}
}

// setClusterStatePodsParallelized sets the pods in cluster state in parallel based on snapshot.parallelism value.
func (snapshot *DeltaSnapshotStore) setClusterStatePodsParallelized(nodeInfos []fwk.NodeInfo, nodeNameToIdx map[string]int, scheduledPods []*apiv1.Pod) {
	podsForNode := make([][]*apiv1.Pod, len(nodeInfos))
	for _, pod := range scheduledPods {
		nodeIdx, ok := nodeNameToIdx[pod.Spec.NodeName]
		if !ok {
			continue
		}
		podsForNode[nodeIdx] = append(podsForNode[nodeIdx], pod)
	}

	ctx := context.Background()
	workqueue.ParallelizeUntil(ctx, snapshot.parallelism, len(nodeInfos), func(nodeIdx int) {
		nodeInfo := nodeInfos[nodeIdx]
		for _, pod := range podsForNode[nodeIdx] {
			// Can add pod directly. Cache will be cleared afterwards.
			podInfo, _ := schedulerframework.NewPodInfo(pod)
			nodeInfo.AddPodInfo(podInfo)
		}
	})
}

// SetClusterState sets the cluster state.
func (snapshot *DeltaSnapshotStore) SetClusterState(nodes []*apiv1.Node, scheduledPods []*apiv1.Pod, draSnapshot *drasnapshot.Snapshot) error {
	snapshot.clear()

	nodeNameToIdx := make(map[string]int, len(nodes))
	nodeInfos := make([]fwk.NodeInfo, len(nodes))
	for i, node := range nodes {
		nodeInfo, err := snapshot.data.addNode(node)
		if err != nil {
			return err
		}
		nodeNameToIdx[node.Name] = i
		nodeInfos[i] = nodeInfo
	}

	if snapshot.parallelism > 1 {
		snapshot.setClusterStatePodsParallelized(nodeInfos, nodeNameToIdx, scheduledPods)
	} else {
		// TODO(macsko): Migrate to setClusterStatePodsParallelized for parallelism == 1
		// after making sure the implementation is always correct in CA 1.33.
		snapshot.setClusterStatePodsSequential(nodeInfos, nodeNameToIdx, scheduledPods)
	}

	// Clear caches after adding pods.
	snapshot.data.clearCaches()

	if draSnapshot == nil {
		snapshot.draSnapshot = drasnapshot.NewEmptySnapshot()
	} else {
		snapshot.draSnapshot = draSnapshot
	}

	return nil
}

// RemoveSchedulerNodeInfo removes nodes (and pods scheduled to it) from the snapshot.
func (snapshot *DeltaSnapshotStore) RemoveSchedulerNodeInfo(nodeName string) error {
	return snapshot.data.removeNodeInfo(nodeName)
}

// ForceAddPod adds pod to the snapshot and schedules it to given node.
func (snapshot *DeltaSnapshotStore) ForceAddPod(pod *apiv1.Pod, nodeName string) error {
	return snapshot.data.addPod(pod, nodeName)
}

// ForceRemovePod removes pod from the snapshot.
func (snapshot *DeltaSnapshotStore) ForceRemovePod(namespace, podName, nodeName string) error {
	return snapshot.data.removePod(namespace, podName, nodeName)
}

// IsPVCUsedByPods returns if the pvc is used by any pod
func (snapshot *DeltaSnapshotStore) IsPVCUsedByPods(key string) bool {
	return snapshot.data.isPVCUsedByPods(key)
}

// Fork creates a fork of snapshot state. All modifications can later be reverted to moment of forking via Revert()
// Time: O(1)
func (snapshot *DeltaSnapshotStore) Fork() {
	snapshot.data = snapshot.data.fork()
	snapshot.draSnapshot.Fork()
}

// Revert reverts snapshot state to moment of forking.
// Time: O(1)
func (snapshot *DeltaSnapshotStore) Revert() {
	if snapshot.data.baseData != nil {
		snapshot.data = snapshot.data.baseData
	}
	snapshot.draSnapshot.Revert()
}

// Commit commits changes done after forking.
// Time: O(n), where n = size of delta (number of nodes added, modified or deleted since forking)
func (snapshot *DeltaSnapshotStore) Commit() error {
	newData, err := snapshot.data.commit()
	if err != nil {
		return err
	}
	snapshot.data = newData
	snapshot.draSnapshot.Commit()
	return nil
}

// Clear reset cluster snapshot to empty, unforked state
// Time: O(1)
func (snapshot *DeltaSnapshotStore) clear() {
	snapshot.data = newInternalDeltaSnapshotData()
	snapshot.draSnapshot = drasnapshot.NewEmptySnapshot()
}
