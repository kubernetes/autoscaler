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

package clustersnapshot

import (
	"context"
	"fmt"

	apiv1 "k8s.io/api/core/v1"
	resourceapi "k8s.io/api/resource/v1alpha3"
	"k8s.io/autoscaler/cluster-autoscaler/dynamicresources"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/klog/v2"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
)

// BasicClusterSnapshot is simple, reference implementation of ClusterSnapshot.
// It is inefficient. But hopefully bug-free and good for initial testing.
type BasicClusterSnapshot struct {
	data       []*internalBasicSnapshotData
	draEnabled bool
	fwHandle   *framework.Handle
}

type internalBasicSnapshotData struct {
	owningSnapshot     *BasicClusterSnapshot
	nodeInfoMap        map[string]*schedulerframework.NodeInfo
	pvcNamespacePodMap map[string]map[string]bool
	draSnapshot        dynamicresources.Snapshot
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
	return nil, ErrNodeNotFound
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

func newInternalBasicSnapshotData(owningSnapshot *BasicClusterSnapshot) *internalBasicSnapshotData {
	return &internalBasicSnapshotData{
		owningSnapshot:     owningSnapshot,
		nodeInfoMap:        make(map[string]*schedulerframework.NodeInfo),
		pvcNamespacePodMap: make(map[string]map[string]bool),
		draSnapshot:        dynamicresources.Snapshot{},
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
		owningSnapshot:     data.owningSnapshot,
		nodeInfoMap:        clonedNodeInfoMap,
		pvcNamespacePodMap: clonedPvcNamespaceNodeMap,
		draSnapshot:        data.draSnapshot.Clone(),
	}
}

func (data *internalBasicSnapshotData) addNode(node *apiv1.Node, extraSlices []*resourceapi.ResourceSlice) error {
	if _, found := data.nodeInfoMap[node.Name]; found {
		return fmt.Errorf("node %s already in snapshot", node.Name)
	}

	if data.owningSnapshot.draEnabled && len(extraSlices) > 0 {
		// We need to add extra ResourceSlices to the DRA snapshot. The DRA snapshot should contain all real slices after Initialize(),
		// so these should be fake node-local slices for a fake duplicated Node.
		err := data.draSnapshot.AddNodeResourceSlices(node.Name, extraSlices)
		if err != nil {
			return fmt.Errorf("couldn't add ResourceSlices to DRA snapshot: %v", err)
		}
	}

	nodeInfo := schedulerframework.NewNodeInfo()
	nodeInfo.SetNode(node)
	data.nodeInfoMap[node.Name] = nodeInfo
	return nil
}

func (data *internalBasicSnapshotData) removeNodeInfo(nodeName string) error {
	nodeInfo, found := data.nodeInfoMap[nodeName]
	if !found {
		return ErrNodeNotFound
	}
	for _, pod := range nodeInfo.Pods {
		data.removePvcUsedByPod(pod.Pod)
		if data.owningSnapshot.draEnabled {
			data.draSnapshot.RemovePodClaims(pod.Pod)
		}
	}
	delete(data.nodeInfoMap, nodeName)
	if data.owningSnapshot.draEnabled {
		data.draSnapshot.RemoveNodeResourceSlices(nodeName)
	}
	return nil
}

func (data *internalBasicSnapshotData) schedulePod(pod *apiv1.Pod, nodeName string, reserveState *schedulerframework.CycleState, extraClaims []*resourceapi.ResourceClaim) error {
	nodeInfo, found := data.nodeInfoMap[nodeName]
	if !found {
		return ErrNodeNotFound
	}

	if dynamicresources.PodNeedsResourceClaims(pod) && data.owningSnapshot.draEnabled {
		err := data.handlePodClaimsScheduling(pod, nodeInfo.Node(), reserveState, extraClaims)
		if err != nil {
			return err
		}
	}

	nodeInfo.AddPod(pod)
	data.addPvcUsedByPod(pod)
	return nil
}

func (data *internalBasicSnapshotData) handlePodClaimsScheduling(pod *apiv1.Pod, node *apiv1.Node, reserveState *schedulerframework.CycleState, extraClaims []*resourceapi.ResourceClaim) error {
	if len(extraClaims) > 0 {
		// We need to add some extra ResourceClaims to the DRA snapshot. The DRA snapshot should contain all real claims after Initialize(),
		// so these should be fake pod-owned claims for a fake duplicated pod.
		err := data.draSnapshot.AddClaims(extraClaims)
		if err != nil {
			return fmt.Errorf("error while adding allocated ResosurceClaims for pod %s/%s: %v", pod.Namespace, pod.Name, err)
		}
	}
	if reserveState != nil {
		// We need to run the scheduler Reserve phase to allocate the appropriate ResourceClaims in the DRA snapshot. The allocations are
		// actually computed and cached in the Filter phase, and Reserve only grabs them from the cycle state. So this should be quick, but
		// it needs the cycle state from after running the Filter phase.
		err := data.owningSnapshot.runReserveNodeForPod(pod, reserveState, node.Name)
		if err != nil {
			return fmt.Errorf("error while trying to run Reserve node %s for pod %s/%s: %v", node.Name, pod.Namespace, pod.Name, err)
		}
	}

	// The pod isn't added to the ReservedFor field of the claim during the Reserve phase (it happens later, in PreBind). We can just do it
	// manually here. It shouldn't fail, it only fails if ReservedFor is at max length already, but that is checked during the Filter phase.
	err := data.draSnapshot.ReservePodClaims(pod)
	if err != nil {
		return fmt.Errorf("couldnn't add pod reservations to claims, this shouldn't happen: %v", err)
	}

	// Verify that all needed claims are tracked in the DRA snapshot, allocated, and available on the Node.
	claims, err := data.draSnapshot.PodClaims(pod)
	if err != nil {
		return fmt.Errorf("couldn't obtain pod %s/%s claims: %v", pod.Namespace, pod.Name, err)
	}
	for _, claim := range claims {
		if available, err := dynamicresources.ClaimAvailableOnNode(claim, node); err != nil || !available {
			return fmt.Errorf("pod %s/%s needs claim %s to schedule, but it isn't available on node %s (allocated: %v, available: %v, err: %v)", pod.Namespace, pod.Name, claim.Name, node.Name, dynamicresources.ClaimAllocated(claim), available, err)
		}
	}
	return nil
}

func (data *internalBasicSnapshotData) unschedulePod(namespace, podName, nodeName string) error {
	nodeInfo, found := data.nodeInfoMap[nodeName]
	if !found {
		return ErrNodeNotFound
	}
	var foundPod *apiv1.Pod
	for _, podInfo := range nodeInfo.Pods {
		if podInfo.Pod.Namespace == namespace && podInfo.Pod.Name == podName {
			foundPod = podInfo.Pod
			break
		}
	}

	if foundPod == nil {
		return fmt.Errorf("pod %s/%s not in snapshot", namespace, podName)
	}

	data.removePvcUsedByPod(foundPod)

	logger := klog.Background()
	err := nodeInfo.RemovePod(logger, foundPod)
	if err != nil {
		data.addPvcUsedByPod(foundPod)
		return fmt.Errorf("cannot remove pod; %v", err)
	}

	if len(foundPod.Spec.ResourceClaims) == 0 || !data.owningSnapshot.draEnabled {
		return nil
	}

	err = data.draSnapshot.UnreservePodClaims(foundPod)
	if err != nil {
		nodeInfo.AddPod(foundPod)
		data.addPvcUsedByPod(foundPod)
		return fmt.Errorf("cannot unreserve pod's dynamic requests: %v", err)
	}
	return nil
}

// NewBasicClusterSnapshot creates instances of BasicClusterSnapshot.
func NewBasicClusterSnapshot(fwHandle *framework.Handle, draEnabled bool) *BasicClusterSnapshot {
	snapshot := &BasicClusterSnapshot{fwHandle: fwHandle, draEnabled: draEnabled}
	snapshot.Clear()
	return snapshot
}

func (snapshot *BasicClusterSnapshot) getInternalData() *internalBasicSnapshotData {
	return snapshot.data[len(snapshot.data)-1]
}

func (snapshot *BasicClusterSnapshot) runReserveNodeForPod(pod *apiv1.Pod, preReserveCycleState *schedulerframework.CycleState, nodeName string) error {
	snapshot.fwHandle.DelegatingLister.UpdateDelegate(snapshot)
	defer snapshot.fwHandle.DelegatingLister.ResetDelegate()

	status := snapshot.fwHandle.Framework.RunReservePluginsReserve(context.Background(), preReserveCycleState, pod, nodeName)
	if !status.IsSuccess() {
		return fmt.Errorf("couldn't reserve node %s for pod %s/%s: %v", nodeName, pod.Namespace, pod.Name, status.Message())
	}
	return nil
}

func (snapshot *BasicClusterSnapshot) GetNodeInfo(nodeName string) (*framework.NodeInfo, error) {
	data := snapshot.getInternalData()
	schedNodeInfo, err := data.getNodeInfo(nodeName)
	if err != nil {
		return nil, err
	}
	if snapshot.draEnabled {
		return data.draSnapshot.WrapSchedulerNodeInfo(schedNodeInfo)
	}
	return framework.WrapSchedulerNodeInfo(schedNodeInfo), nil
}

func (snapshot *BasicClusterSnapshot) ListNodeInfos() ([]*framework.NodeInfo, error) {
	data := snapshot.getInternalData()
	schedNodeInfos := data.listNodeInfos()
	if snapshot.draEnabled {
		return data.draSnapshot.WrapSchedulerNodeInfos(schedNodeInfos)
	}
	return framework.WrapSchedulerNodeInfos(schedNodeInfos), nil
}

func (snapshot *BasicClusterSnapshot) AddNodeInfo(nodeInfo *framework.NodeInfo) error {
	if err := snapshot.getInternalData().addNode(nodeInfo.Node(), nodeInfo.LocalResourceSlices); err != nil {
		return err
	}
	for _, podInfo := range nodeInfo.Pods {
		if err := snapshot.getInternalData().schedulePod(podInfo.Pod, nodeInfo.Node().Name, nil, podInfo.NeededResourceClaims); err != nil {
			return err
		}
	}
	return nil
}

func (snapshot *BasicClusterSnapshot) Initialize(nodes []*apiv1.Node, scheduledPods []*apiv1.Pod, draSnapshot dynamicresources.Snapshot) error {
	snapshot.Clear()
	baseData := snapshot.getInternalData()

	if snapshot.draEnabled {
		baseData.draSnapshot = draSnapshot
	}

	knownNodes := make(map[string]bool)
	for _, node := range nodes {
		if err := baseData.addNode(node, nil); err != nil {
			return err
		}
		knownNodes[node.Name] = true
	}
	for _, pod := range scheduledPods {
		if knownNodes[pod.Spec.NodeName] {
			if err := baseData.schedulePod(pod, pod.Spec.NodeName, nil, nil); err != nil {
				return err
			}
		}
	}
	return nil
}

func (snapshot *BasicClusterSnapshot) AddResourceClaims(extraClaims []*resourceapi.ResourceClaim) error {
	return snapshot.getInternalData().draSnapshot.AddClaims(extraClaims)
}

func (snapshot *BasicClusterSnapshot) GetPodResourceClaims(pod *apiv1.Pod) ([]*resourceapi.ResourceClaim, error) {
	return snapshot.getInternalData().draSnapshot.PodClaims(pod)
}

// RemoveNodeInfo removes nodes (and pods scheduled to it) from the snapshot.
func (snapshot *BasicClusterSnapshot) RemoveNodeInfo(nodeName string) error {
	return snapshot.getInternalData().removeNodeInfo(nodeName)
}

// SchedulePod adds pod to the snapshot and schedules it to given node.
func (snapshot *BasicClusterSnapshot) SchedulePod(pod *apiv1.Pod, nodeName string, reserveState *schedulerframework.CycleState) error {
	return snapshot.getInternalData().schedulePod(pod, nodeName, reserveState, nil)
}

// UnschedulePod removes pod from the snapshot.
func (snapshot *BasicClusterSnapshot) UnschedulePod(namespace, podName, nodeName string) error {
	return snapshot.getInternalData().unschedulePod(namespace, podName, nodeName)
}

// IsPVCUsedByPods returns if the pvc is used by any pod
func (snapshot *BasicClusterSnapshot) IsPVCUsedByPods(key string) bool {
	return snapshot.getInternalData().isPVCUsedByPods(key)
}

// Fork creates a fork of snapshot state. All modifications can later be reverted to moment of forking via Revert()
func (snapshot *BasicClusterSnapshot) Fork() {
	forkData := snapshot.getInternalData().clone()
	snapshot.data = append(snapshot.data, forkData)
}

// Revert reverts snapshot state to moment of forking.
func (snapshot *BasicClusterSnapshot) Revert() {
	if len(snapshot.data) == 1 {
		return
	}
	snapshot.data = snapshot.data[:len(snapshot.data)-1]
}

// Commit commits changes done after forking.
func (snapshot *BasicClusterSnapshot) Commit() error {
	if len(snapshot.data) <= 1 {
		// do nothing
		return nil
	}
	snapshot.data = append(snapshot.data[:len(snapshot.data)-2], snapshot.data[len(snapshot.data)-1])
	return nil
}

// Clear reset cluster snapshot to empty, unforked state
func (snapshot *BasicClusterSnapshot) Clear() {
	baseData := newInternalBasicSnapshotData(snapshot)
	snapshot.data = []*internalBasicSnapshotData{baseData}
}

// implementation of SharedLister interface

type basicClusterSnapshotNodeLister BasicClusterSnapshot
type basicClusterSnapshotStorageLister BasicClusterSnapshot

// NodeInfos exposes snapshot as NodeInfoLister.
func (snapshot *BasicClusterSnapshot) NodeInfos() schedulerframework.NodeInfoLister {
	return (*basicClusterSnapshotNodeLister)(snapshot)
}

// StorageInfos exposes snapshot as StorageInfoLister.
func (snapshot *BasicClusterSnapshot) StorageInfos() schedulerframework.StorageInfoLister {
	return (*basicClusterSnapshotStorageLister)(snapshot)
}

func (snapshot *BasicClusterSnapshot) ResourceClaims() schedulerframework.ResourceClaimTracker {
	return snapshot.getInternalData().draSnapshot.ResourceClaims()
}

func (snapshot *BasicClusterSnapshot) ResourceSlices() schedulerframework.ResourceSliceLister {
	return snapshot.getInternalData().draSnapshot.ResourceSlices()
}

func (snapshot *BasicClusterSnapshot) DeviceClasses() schedulerframework.DeviceClassLister {
	return snapshot.getInternalData().draSnapshot.DeviceClasses()
}

// List returns the list of nodes in the snapshot.
func (snapshot *basicClusterSnapshotNodeLister) List() ([]*schedulerframework.NodeInfo, error) {
	return (*BasicClusterSnapshot)(snapshot).getInternalData().listNodeInfos(), nil
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

// Returns the IsPVCUsedByPods in a given key.
func (snapshot *basicClusterSnapshotStorageLister) IsPVCUsedByPods(key string) bool {
	return (*BasicClusterSnapshot)(snapshot).getInternalData().isPVCUsedByPods(key)
}
