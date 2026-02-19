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

package predicate

import (
	"fmt"
	"maps"
	"math/rand"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert" // TODO: Migrate the rest of the assertions to cmp for consistency.

	apiv1 "k8s.io/api/core/v1"
	resourceapi "k8s.io/api/resource/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apiserver/pkg/util/feature"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot/store"
	csisnapshot "k8s.io/autoscaler/cluster-autoscaler/simulator/csi/snapshot"
	drasnapshot "k8s.io/autoscaler/cluster-autoscaler/simulator/dynamicresources/snapshot"
	drautils "k8s.io/autoscaler/cluster-autoscaler/simulator/dynamicresources/utils"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
	featuretesting "k8s.io/component-base/featuregate/testing"
	"k8s.io/kubernetes/pkg/features"
	schedulerimpl "k8s.io/kubernetes/pkg/scheduler/framework"
)

var snapshots = map[string]func() (clustersnapshot.ClusterSnapshot, error){
	"basic": func() (clustersnapshot.ClusterSnapshot, error) {
		fwHandle, err := framework.NewTestFrameworkHandle()
		if err != nil {
			return nil, err
		}
		return NewPredicateSnapshot(store.NewBasicSnapshotStore(), fwHandle, true, 1, true), nil
	},
	"delta": func() (clustersnapshot.ClusterSnapshot, error) {
		fwHandle, err := framework.NewTestFrameworkHandle()
		if err != nil {
			return nil, err
		}
		return NewPredicateSnapshot(store.NewDeltaSnapshotStore(16), fwHandle, true, 1, true), nil
	},
}

func extractNodes(nodeInfos []*framework.NodeInfo) []*apiv1.Node {
	nodes := []*apiv1.Node{}
	for _, ni := range nodeInfos {
		nodes = append(nodes, ni.Node())
	}
	return nodes
}

type snapshotState struct {
	nodes       []*apiv1.Node
	podsByNode  map[string][]*apiv1.Pod
	draSnapshot *drasnapshot.Snapshot
	csiSnapshot *csisnapshot.Snapshot
}

func compareStates(t *testing.T, a, b snapshotState) {
	if diff := cmp.Diff(a.nodes, b.nodes, cmpopts.EquateEmpty(), IgnoreObjectOrder[*apiv1.Node]()); diff != "" {
		t.Errorf("Nodes: unexpected diff (-want +got): %s", diff)
	}
	if diff := cmp.Diff(a.podsByNode, b.podsByNode, cmpopts.EquateEmpty(), IgnoreObjectOrder[*apiv1.Pod]()); diff != "" {
		t.Errorf("Pods: unexpected diff (-want +got): %s", diff)
	}

	aClaims, err := a.draSnapshot.ResourceClaims().List()
	if err != nil {
		t.Errorf("ResourceClaims().List(): unexpected error: %v", err)
	}
	bClaims, err := b.draSnapshot.ResourceClaims().List()
	if err != nil {
		t.Errorf("ResourceClaims().List(): unexpected error: %v", err)
	}
	claimDiffOpts := []cmp.Option{
		cmpopts.EquateEmpty(),
		IgnoreObjectOrder[*resourceapi.ResourceClaim](),
		cmpopts.SortSlices(func(ref1, ref2 resourceapi.ResourceClaimConsumerReference) bool { return ref1.Name < ref2.Name }),
		// The DRA plugin adds a finalizer to the claim during the allocation. The finalizer shouldn't have any effect on scheduling,
		// so we don't care about it in CA. Ignore the Finalizers field to simplify assertions.
		cmpopts.IgnoreFields(metav1.ObjectMeta{}, "Finalizers"),
	}
	if diff := cmp.Diff(aClaims, bClaims, claimDiffOpts...); diff != "" {
		t.Errorf("ResourceClaims().List(): unexpected diff (-want +got): %s", diff)
	}

	aSlices, err := a.draSnapshot.ResourceSlices().ListWithDeviceTaintRules()
	if err != nil {
		t.Errorf("ResourceSlices().List(): unexpected error: %v", err)
	}
	bSlices, err := b.draSnapshot.ResourceSlices().ListWithDeviceTaintRules()
	if err != nil {
		t.Errorf("ResourceSlices().List(): unexpected error: %v", err)
	}
	if diff := cmp.Diff(aSlices, bSlices, cmpopts.EquateEmpty(), IgnoreObjectOrder[*resourceapi.ResourceSlice]()); diff != "" {
		t.Errorf("ResourceSlices().List(): unexpected diff (-want +got): %s", diff)
	}

	aClasses, err := a.draSnapshot.DeviceClasses().List()
	if err != nil {
		t.Errorf("DeviceClasses().List(): unexpected error: %v", err)
	}
	bClasses, err := b.draSnapshot.DeviceClasses().List()
	if err != nil {
		t.Errorf("DeviceClasses().List(): unexpected error: %v", err)
	}
	if diff := cmp.Diff(aClasses, bClasses, cmpopts.EquateEmpty(), IgnoreObjectOrder[*resourceapi.DeviceClass]()); diff != "" {
		t.Errorf("DeviceClasses().List(): unexpected diff (-want +got): %s", diff)
	}

	aCSINodes, err := a.csiSnapshot.CSINodes().List()
	if err != nil {
		t.Errorf("CSINodes().List(): unexpected error: %v", err)
	}

	bCSINodes, err := b.csiSnapshot.CSINodes().List()
	if err != nil {
		t.Errorf("CSINodes().List(): unexpected error: %v", err)
	}

	if diff := cmp.Diff(aCSINodes, bCSINodes, cmpopts.EquateEmpty(), IgnoreObjectOrder[*storagev1.CSINode]()); diff != "" {
		t.Errorf("CSINodes().List(): unexpected diff (-want +got): %s", diff)
	}
}

func getSnapshotState(t *testing.T, snapshot clustersnapshot.ClusterSnapshot) snapshotState {
	nodes, err := snapshot.ListNodeInfos()
	assert.NoError(t, err)
	pods := map[string][]*apiv1.Pod{}
	for _, nodeInfo := range nodes {
		for _, podInfo := range nodeInfo.Pods() {
			pods[nodeInfo.Node().Name] = append(pods[nodeInfo.Node().Name], podInfo.Pod)
		}
	}
	return snapshotState{nodes: extractNodes(nodes), podsByNode: pods, draSnapshot: snapshot.DraSnapshot(), csiSnapshot: snapshot.CsiSnapshot()}
}

func startSnapshot(t *testing.T, snapshotFactory func() (clustersnapshot.ClusterSnapshot, error), state snapshotState) clustersnapshot.ClusterSnapshot {
	snapshot, err := snapshotFactory()
	assert.NoError(t, err)
	var pods []*apiv1.Pod
	for _, nodePods := range state.podsByNode {
		for _, pod := range nodePods {
			pods = append(pods, pod)
		}
	}

	draSnapshot := drasnapshot.CloneTestSnapshot(state.draSnapshot)
	csiSnapshot := csisnapshot.CloneTestSnapshot(state.csiSnapshot)
	err = snapshot.SetClusterState(state.nodes, pods, draSnapshot, csiSnapshot)
	assert.NoError(t, err)
	return snapshot
}

type modificationTestCase struct {
	name          string
	op            func(clustersnapshot.ClusterSnapshot) error
	wantErr       error
	state         snapshotState
	modifiedState snapshotState
}

func (tc modificationTestCase) runAndValidateOp(t *testing.T, snapshot clustersnapshot.ClusterSnapshot) {
	err := tc.op(snapshot)
	if diff := cmp.Diff(tc.wantErr, err, cmpopts.EquateErrors()); diff != "" {
		t.Errorf("unexpected error when modifying the snapshot (-want +got): %s", diff)
	}
}

func validTestCases(t *testing.T, snapshotName string) []modificationTestCase {
	node := BuildTestNode("specialNode", 10, 100)
	otherNode := BuildTestNode("otherNode", 10, 100)
	largeNode := BuildTestNode("largeNode", 9999, 9999)

	csiNode := BuildCSINode(node)
	otherCSINode := BuildCSINode(otherNode)
	largeCSINode := BuildCSINode(largeNode)

	nodeSelector := &apiv1.NodeSelector{
		NodeSelectorTerms: []apiv1.NodeSelectorTerm{
			{
				MatchFields: []apiv1.NodeSelectorRequirement{
					{
						Key:      "metadata.name",
						Operator: apiv1.NodeSelectorOpIn,
						Values:   []string{node.Name},
					},
				},
			},
		},
	}
	wrongNodeSelector := &apiv1.NodeSelector{
		NodeSelectorTerms: []apiv1.NodeSelectorTerm{
			{
				MatchFields: []apiv1.NodeSelectorRequirement{
					{
						Key:      "metadata.name",
						Operator: apiv1.NodeSelectorOpIn,
						Values:   []string{"wrongNode"},
					},
				},
			},
		},
	}
	resourceSlices := []*resourceapi.ResourceSlice{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "slice1", UID: "slice1Uid"},
			Spec: resourceapi.ResourceSliceSpec{
				NodeName: &node.Name,
				Driver:   "driver.foo.com",
				Pool: resourceapi.ResourcePool{
					Name:               "pool1",
					ResourceSliceCount: 1,
				},
				Devices: []resourceapi.Device{
					{Name: "dev1"},
					{Name: "dev2"},
					{Name: "dev3"},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "slice2", UID: "slice2Uid"},
			Spec: resourceapi.ResourceSliceSpec{
				NodeName: &node.Name,
				Driver:   "driver.bar.com",
				Pool: resourceapi.ResourcePool{
					Name:               "pool2",
					ResourceSliceCount: 1,
				},
				Devices: []resourceapi.Device{
					{Name: "dev1"},
					{Name: "dev2"},
					{Name: "dev3"},
				},
			},
		},
	}

	pod := BuildTestPod("specialPod", 1, 1)
	largePod := BuildTestPod("largePod", 999, 999)
	podWithClaims := BuildTestPod("podWithClaims", 1, 1,
		WithResourceClaim("claim1", "sharedClaim", ""), WithResourceClaim("claim2", "podOwnedClaim", "podOwnedClaimTemplate"))

	podOwnedClaim := drautils.TestClaimWithPodOwnership(podWithClaims,
		&resourceapi.ResourceClaim{
			ObjectMeta: metav1.ObjectMeta{Name: "podOwnedClaim", UID: "podOwnedClaimUid", Namespace: "default"},
			Spec: resourceapi.ResourceClaimSpec{
				Devices: resourceapi.DeviceClaim{
					Requests: []resourceapi.DeviceRequest{
						{
							Name: "req1",
							Exactly: &resourceapi.ExactDeviceRequest{
								DeviceClassName: "defaultClass",
								Selectors:       []resourceapi.DeviceSelector{{CEL: &resourceapi.CELDeviceSelector{Expression: `device.driver == "driver.foo.com"`}}},
								AllocationMode:  resourceapi.DeviceAllocationModeExactCount,
								Count:           3,
							},
						},
					},
				},
			},
		},
	)
	podOwnedClaimAlloc := &resourceapi.AllocationResult{
		NodeSelector: nodeSelector,
		Devices: resourceapi.DeviceAllocationResult{
			Results: []resourceapi.DeviceRequestAllocationResult{
				{Request: "req1", Driver: "driver.foo.com", Pool: "pool1", Device: "dev1"},
				{Request: "req1", Driver: "driver.foo.com", Pool: "pool1", Device: "dev2"},
				{Request: "req1", Driver: "driver.foo.com", Pool: "pool1", Device: "dev3"},
			},
		},
	}
	podOwnedClaimAllocWrongNode := &resourceapi.AllocationResult{
		NodeSelector: wrongNodeSelector,
		Devices: resourceapi.DeviceAllocationResult{
			Results: []resourceapi.DeviceRequestAllocationResult{
				{Request: "req1", Driver: "driver.foo.com", Pool: "pool1", Device: "dev1"},
				{Request: "req1", Driver: "driver.foo.com", Pool: "pool1", Device: "dev2"},
				{Request: "req1", Driver: "driver.foo.com", Pool: "pool1", Device: "dev3"},
			},
		},
	}
	sharedClaim := &resourceapi.ResourceClaim{
		ObjectMeta: metav1.ObjectMeta{Name: "sharedClaim", UID: "sharedClaimUid", Namespace: "default"},
		Spec: resourceapi.ResourceClaimSpec{
			Devices: resourceapi.DeviceClaim{
				Requests: []resourceapi.DeviceRequest{
					{
						Name: "req1",
						Exactly: &resourceapi.ExactDeviceRequest{
							DeviceClassName: "defaultClass",
							Selectors:       []resourceapi.DeviceSelector{{CEL: &resourceapi.CELDeviceSelector{Expression: `device.driver == "driver.bar.com"`}}},
							AllocationMode:  resourceapi.DeviceAllocationModeAll,
						},
					},
				},
			},
		},
	}
	sharedClaimAlloc := &resourceapi.AllocationResult{
		NodeSelector: nodeSelector,
		Devices: resourceapi.DeviceAllocationResult{
			Results: []resourceapi.DeviceRequestAllocationResult{
				{Request: "req1", Driver: "driver.bar.com", Pool: "pool2", Device: "dev1"},
				{Request: "req1", Driver: "driver.bar.com", Pool: "pool2", Device: "dev2"},
				{Request: "req1", Driver: "driver.bar.com", Pool: "pool2", Device: "dev3"},
			},
		},
	}

	deviceClasses := map[string]*resourceapi.DeviceClass{
		"defaultClass": {ObjectMeta: metav1.ObjectMeta{Name: "defaultClass", UID: "defaultClassUid"}},
	}

	testCases := []modificationTestCase{
		{
			name: "add empty nodeInfo",
			op: func(snapshot clustersnapshot.ClusterSnapshot) error {
				return snapshot.AddNodeInfo(framework.NewTestNodeInfoWithCSI(node, csiNode))
			},
			modifiedState: snapshotState{
				nodes:       []*apiv1.Node{node},
				csiSnapshot: createCSISnapshot(csiNode),
			},
		},
		{
			name: "add nodeInfo",
			op: func(snapshot clustersnapshot.ClusterSnapshot) error {
				return snapshot.AddNodeInfo(framework.NewTestNodeInfoWithCSI(node, csiNode, pod))
			},
			modifiedState: snapshotState{
				nodes:       []*apiv1.Node{node},
				podsByNode:  map[string][]*apiv1.Pod{node.Name: {pod}},
				csiSnapshot: createCSISnapshot(csiNode),
			},
		},
		{
			name: "remove nodeInfo",
			state: snapshotState{
				nodes:       []*apiv1.Node{node},
				csiSnapshot: createCSISnapshot(csiNode),
			},
			op: func(snapshot clustersnapshot.ClusterSnapshot) error {
				return snapshot.RemoveNodeInfo(node.Name)
			},
		},
		{
			name: "remove nodeInfo, then add it back",
			state: snapshotState{
				nodes:       []*apiv1.Node{node},
				csiSnapshot: createCSISnapshot(csiNode),
			},
			op: func(snapshot clustersnapshot.ClusterSnapshot) error {
				if err := snapshot.RemoveNodeInfo(node.Name); err != nil {
					return err
				}
				return snapshot.AddNodeInfo(framework.NewTestNodeInfoWithCSI(node, csiNode))
			},
			modifiedState: snapshotState{
				nodes:       []*apiv1.Node{node},
				csiSnapshot: createCSISnapshot(csiNode),
			},
		},
		{
			name: "add pod, then remove nodeInfo",
			state: snapshotState{
				nodes:       []*apiv1.Node{node},
				csiSnapshot: createCSISnapshot(csiNode),
			},
			op: func(snapshot clustersnapshot.ClusterSnapshot) error {
				if schedErr := snapshot.ForceAddPod(pod, node.Name); schedErr != nil {
					return schedErr
				}
				return snapshot.RemoveNodeInfo(node.Name)
			},
		},
		{
			name: "schedule pod",
			state: snapshotState{
				nodes:       []*apiv1.Node{node},
				csiSnapshot: createCSISnapshot(csiNode),
			},
			op: func(snapshot clustersnapshot.ClusterSnapshot) error {
				return snapshot.SchedulePod(pod, node.Name)
			},
			modifiedState: snapshotState{
				nodes:       []*apiv1.Node{node},
				podsByNode:  map[string][]*apiv1.Pod{node.Name: {pod}},
				csiSnapshot: createCSISnapshot(csiNode),
			},
		},
		{
			name: "schedule pod on any Node (scheduling predicates only work for one)",
			state: snapshotState{
				nodes:       []*apiv1.Node{node, largeNode},
				csiSnapshot: createCSISnapshot(csiNode, largeCSINode),
			},
			op: func(snapshot clustersnapshot.ClusterSnapshot) error {
				foundNodeName, err := snapshot.SchedulePodOnAnyNodeMatching(largePod, func(_ *framework.NodeInfo) bool { return true })
				if diff := cmp.Diff(largeNode.Name, foundNodeName); diff != "" {
					t.Errorf("SchedulePodOnAnyNodeMatching(): unexpected output (-want +got): %s", diff)
				}
				return err
			},
			modifiedState: snapshotState{
				nodes:       []*apiv1.Node{node, largeNode},
				podsByNode:  map[string][]*apiv1.Pod{largeNode.Name: {largePod}},
				csiSnapshot: createCSISnapshot(csiNode, largeCSINode),
			},
		},
		{
			name: "schedule pod on any Node matching (matching only works for one)",
			state: snapshotState{
				nodes:       []*apiv1.Node{node, largeNode},
				csiSnapshot: createCSISnapshot(csiNode, largeCSINode),
			},
			op: func(snapshot clustersnapshot.ClusterSnapshot) error {
				foundNodeName, err := snapshot.SchedulePodOnAnyNodeMatching(pod, func(info *framework.NodeInfo) bool { return info.Node().Name == node.Name })
				if diff := cmp.Diff(node.Name, foundNodeName); diff != "" {
					t.Errorf("SchedulePodOnAnyNodeMatching(): unexpected output (-want +got): %s", diff)
				}
				return err
			},
			modifiedState: snapshotState{
				nodes:       []*apiv1.Node{node, largeNode},
				podsByNode:  map[string][]*apiv1.Pod{node.Name: {pod}},
				csiSnapshot: createCSISnapshot(csiNode, largeCSINode),
			},
		},
		{
			name: "scheduling pod with failing predicates is an error",
			state: snapshotState{
				nodes:       []*apiv1.Node{node},
				csiSnapshot: createCSISnapshot(csiNode),
			},
			op: func(snapshot clustersnapshot.ClusterSnapshot) error {
				return snapshot.SchedulePod(largePod, node.Name)
			},
			// SchedulePod should fail on scheduling predicates because the pod is too big for the node.
			wantErr: clustersnapshot.NewFailingPredicateError(nil, "", nil, "", ""), // Only the type of the error is asserted (via cmp.EquateErrors() and errors.Is()), so the parameters don't matter here.
			// The state shouldn't change on error.
			modifiedState: snapshotState{
				nodes:       []*apiv1.Node{node},
				csiSnapshot: createCSISnapshot(csiNode),
			},
		},
		{
			name: "scheduling pod on any Node with failing predicates on all Nodes is an error",
			state: snapshotState{
				nodes:       []*apiv1.Node{node, otherNode},
				csiSnapshot: createCSISnapshot(csiNode, otherCSINode),
			},
			op: func(snapshot clustersnapshot.ClusterSnapshot) error {
				foundNodeName, err := snapshot.SchedulePodOnAnyNodeMatching(largePod, func(_ *framework.NodeInfo) bool { return true })
				if foundNodeName != "" {
					t.Errorf("SchedulePodOnAnyNodeMatching(): unexpected output: want empty string, got %q", foundNodeName)
				}
				return err
			},
			// SchedulePodOnAnyNodeMatching should fail on scheduling predicates because the pod is too big for either Node.
			wantErr: clustersnapshot.NewNoNodesPassingPredicatesFoundError(nil), // Only the type of the error is asserted (via cmp.EquateErrors() and errors.Is()), so the parameters don't matter here.
			// The state shouldn't change on error.
			modifiedState: snapshotState{
				nodes:       []*apiv1.Node{node, otherNode},
				csiSnapshot: createCSISnapshot(csiNode, otherCSINode),
			},
		},
		{
			name: "scheduling pod on any matching Node with no Nodes matching is an error",
			state: snapshotState{
				nodes:       []*apiv1.Node{node, otherNode},
				csiSnapshot: createCSISnapshot(csiNode, otherCSINode),
			},
			op: func(snapshot clustersnapshot.ClusterSnapshot) error {
				foundNodeName, err := snapshot.SchedulePodOnAnyNodeMatching(pod, func(_ *framework.NodeInfo) bool { return false })
				if foundNodeName != "" {
					t.Errorf("SchedulePodOnAnyNodeMatching(): unexpected output: want empty string, got %q", foundNodeName)
				}
				return err
			},
			// SchedulePodOnAnyNodeMatching should fail on scheduling predicates because the pod is too big for either Node.
			wantErr: clustersnapshot.NewNoNodesPassingPredicatesFoundError(nil), // Only the type of the error is asserted (via cmp.EquateErrors() and errors.Is()), so the parameters don't matter here.
			// The state shouldn't change on error.
			modifiedState: snapshotState{
				nodes:       []*apiv1.Node{node, otherNode},
				csiSnapshot: createCSISnapshot(csiNode, otherCSINode),
			},
		},
		{
			name: "unschedule pod",
			state: snapshotState{
				nodes:       []*apiv1.Node{largeNode},
				csiSnapshot: createCSISnapshot(largeCSINode),
				podsByNode:  map[string][]*apiv1.Pod{largeNode.Name: {withNodeName(pod, largeNode.Name), withNodeName(largePod, largeNode.Name)}},
			},
			op: func(snapshot clustersnapshot.ClusterSnapshot) error {
				return snapshot.UnschedulePod(pod.Namespace, pod.Name, largeNode.Name)
			},
			modifiedState: snapshotState{
				nodes:       []*apiv1.Node{largeNode},
				csiSnapshot: createCSISnapshot(largeCSINode),
				podsByNode:  map[string][]*apiv1.Pod{largeNode.Name: {withNodeName(largePod, largeNode.Name)}},
			},
		},
		{
			name: "add empty nodeInfo with LocalResourceSlices",
			state: snapshotState{
				draSnapshot: drasnapshot.NewSnapshot(nil, nil, nil, deviceClasses),
			},
			op: func(snapshot clustersnapshot.ClusterSnapshot) error {
				nodeInfo := framework.NewNodeInfo(node, resourceSlices)
				nodeInfo.CSINode = csiNode
				return snapshot.AddNodeInfo(nodeInfo)
			},
			// LocalResourceSlices from the NodeInfo should get added to the DRA snapshot.
			modifiedState: snapshotState{
				nodes:       []*apiv1.Node{node},
				draSnapshot: drasnapshot.NewSnapshot(nil, map[string][]*resourceapi.ResourceSlice{node.Name: resourceSlices}, nil, deviceClasses),
				csiSnapshot: createCSISnapshot(csiNode),
			},
		},
		{
			name: "add nodeInfo with LocalResourceSlices and NeededResourceClaims",
			state: snapshotState{
				draSnapshot: drasnapshot.NewSnapshot(
					map[drasnapshot.ResourceClaimId]*resourceapi.ResourceClaim{
						drasnapshot.GetClaimId(sharedClaim): drautils.TestClaimWithAllocation(sharedClaim, sharedClaimAlloc),
					}, nil, nil, deviceClasses),
			},
			// The pod in the added NodeInfo references the shared claim already in the DRA snapshot, and a new pod-owned allocated claim that
			// needs to be added to the DRA snapshot.
			op: func(snapshot clustersnapshot.ClusterSnapshot) error {
				podInfo := framework.NewPodInfo(podWithClaims, []*resourceapi.ResourceClaim{
					drautils.TestClaimWithAllocation(podOwnedClaim, podOwnedClaimAlloc),
					drautils.TestClaimWithAllocation(sharedClaim, sharedClaimAlloc),
				})
				nodeInfo := framework.NewNodeInfo(node, resourceSlices, podInfo)
				nodeInfo.CSINode = csiNode
				return snapshot.AddNodeInfo(nodeInfo)
			},
			// The shared claim should just get a reservation for the pod added in the DRA snapshot.
			// The pod-owned claim should get added to the DRA snapshot, with a reservation for the pod.
			// LocalResourceSlices from the NodeInfo should get added to the DRA snapshot.
			modifiedState: snapshotState{
				nodes:      []*apiv1.Node{node},
				podsByNode: map[string][]*apiv1.Pod{node.Name: {podWithClaims}},
				draSnapshot: drasnapshot.NewSnapshot(
					map[drasnapshot.ResourceClaimId]*resourceapi.ResourceClaim{
						drasnapshot.GetClaimId(podOwnedClaim): drautils.TestClaimWithPodReservations(drautils.TestClaimWithAllocation(podOwnedClaim, podOwnedClaimAlloc), podWithClaims),
						drasnapshot.GetClaimId(sharedClaim):   drautils.TestClaimWithPodReservations(drautils.TestClaimWithAllocation(sharedClaim, sharedClaimAlloc), podWithClaims),
					},
					map[string][]*resourceapi.ResourceSlice{node.Name: resourceSlices}, nil, deviceClasses),
				csiSnapshot: createCSISnapshot(csiNode),
			},
		},
		{
			name: "adding LocalResourceSlices for an already tracked Node is an error",
			state: snapshotState{
				draSnapshot: drasnapshot.NewSnapshot(nil, map[string][]*resourceapi.ResourceSlice{node.Name: resourceSlices}, nil, deviceClasses),
			},
			op: func(snapshot clustersnapshot.ClusterSnapshot) error {
				nodeInfo := framework.NewNodeInfo(node, resourceSlices)
				nodeInfo.CSINode = csiNode
				return snapshot.AddNodeInfo(nodeInfo)
			},
			// LocalResourceSlices for the Node already exist in the DRA snapshot, so trying to add them again should be an error.
			wantErr: cmpopts.AnyError,
			// The state shouldn't change on error.
			modifiedState: snapshotState{
				draSnapshot: drasnapshot.NewSnapshot(nil, map[string][]*resourceapi.ResourceSlice{node.Name: resourceSlices}, nil, deviceClasses),
			},
		},
		{
			name: "adding already tracked pod-owned ResourceClaims is an error",
			state: snapshotState{
				draSnapshot: drasnapshot.NewSnapshot(
					map[drasnapshot.ResourceClaimId]*resourceapi.ResourceClaim{
						drasnapshot.GetClaimId(sharedClaim):   drautils.TestClaimWithAllocation(sharedClaim, sharedClaimAlloc),
						drasnapshot.GetClaimId(podOwnedClaim): podOwnedClaim.DeepCopy(),
					}, nil, nil, deviceClasses),
			},
			op: func(snapshot clustersnapshot.ClusterSnapshot) error {
				podInfo := framework.NewPodInfo(podWithClaims, []*resourceapi.ResourceClaim{
					drautils.TestClaimWithAllocation(podOwnedClaim, podOwnedClaimAlloc),
					drautils.TestClaimWithAllocation(sharedClaim, sharedClaimAlloc),
				})
				return snapshot.AddNodeInfo(framework.NewNodeInfo(node, resourceSlices, podInfo))
			},
			// The pod-owned claim already exists in the DRA snapshot, so trying to add it again should be an error.
			wantErr: cmpopts.AnyError,
			// The state shouldn't change on error.
			// TODO(DRA): Until transaction-like clean-up is implemented in AddNodeInfo, the state is not cleaned up on error. Make modifiedState identical to initial state after the clean-up is implemented.
			modifiedState: snapshotState{
				draSnapshot: drasnapshot.NewSnapshot(
					map[drasnapshot.ResourceClaimId]*resourceapi.ResourceClaim{
						drasnapshot.GetClaimId(sharedClaim):   drautils.TestClaimWithAllocation(sharedClaim, sharedClaimAlloc),
						drasnapshot.GetClaimId(podOwnedClaim): podOwnedClaim,
					}, map[string][]*resourceapi.ResourceSlice{node.Name: resourceSlices}, nil, deviceClasses),
			},
		},
		{
			name: "adding unallocated pod-owned ResourceClaims is an error",
			state: snapshotState{
				draSnapshot: drasnapshot.NewSnapshot(
					map[drasnapshot.ResourceClaimId]*resourceapi.ResourceClaim{
						drasnapshot.GetClaimId(sharedClaim): drautils.TestClaimWithAllocation(sharedClaim, sharedClaimAlloc),
					}, nil, nil, deviceClasses),
			},
			op: func(snapshot clustersnapshot.ClusterSnapshot) error {
				podInfo := framework.NewPodInfo(podWithClaims, []*resourceapi.ResourceClaim{
					podOwnedClaim.DeepCopy(),
					drautils.TestClaimWithAllocation(sharedClaim, sharedClaimAlloc),
				})
				return snapshot.AddNodeInfo(framework.NewNodeInfo(node, resourceSlices, podInfo))
			},
			// The added pod-owned claim isn't allocated, so AddNodeInfo should fail.
			wantErr: cmpopts.AnyError,
			// The state shouldn't change on error.
			// TODO(DRA): Until transaction-like clean-up is implemented in AddNodeInfo, the state is not cleaned up on error. Make modifiedState identical to initial state after the clean-up is implemented.
			modifiedState: snapshotState{
				draSnapshot: drasnapshot.NewSnapshot(
					map[drasnapshot.ResourceClaimId]*resourceapi.ResourceClaim{
						drasnapshot.GetClaimId(sharedClaim):   drautils.TestClaimWithPodReservations(drautils.TestClaimWithAllocation(sharedClaim, sharedClaimAlloc), podWithClaims),
						drasnapshot.GetClaimId(podOwnedClaim): drautils.TestClaimWithPodReservations(podOwnedClaim, podWithClaims),
					}, map[string][]*resourceapi.ResourceSlice{node.Name: resourceSlices}, nil, deviceClasses),
			},
		},
		{
			name: "adding pod-owned ResourceClaims allocated to the wrong Node is an error",
			state: snapshotState{
				draSnapshot: drasnapshot.NewSnapshot(
					map[drasnapshot.ResourceClaimId]*resourceapi.ResourceClaim{
						drasnapshot.GetClaimId(sharedClaim): drautils.TestClaimWithAllocation(sharedClaim, sharedClaimAlloc),
					}, nil, nil, deviceClasses),
			},
			op: func(snapshot clustersnapshot.ClusterSnapshot) error {
				podInfo := framework.NewPodInfo(podWithClaims, []*resourceapi.ResourceClaim{
					drautils.TestClaimWithAllocation(podOwnedClaim, podOwnedClaimAllocWrongNode),
					drautils.TestClaimWithAllocation(sharedClaim, sharedClaimAlloc),
				})
				return snapshot.AddNodeInfo(framework.NewNodeInfo(node, resourceSlices, podInfo))
			},
			// The added pod-owned claim is allocated to a different Node than the one being added, so AddNodeInfo should fail.
			wantErr: cmpopts.AnyError,
			// The state shouldn't change on error.
			// TODO(DRA): Until transaction-like clean-up is implemented in AddNodeInfo, the state is not cleaned up on error. Make modifiedState identical to initial state after the clean-up is implemented.
			modifiedState: snapshotState{
				draSnapshot: drasnapshot.NewSnapshot(
					map[drasnapshot.ResourceClaimId]*resourceapi.ResourceClaim{
						drasnapshot.GetClaimId(sharedClaim):   drautils.TestClaimWithPodReservations(drautils.TestClaimWithAllocation(sharedClaim, sharedClaimAlloc), podWithClaims),
						drasnapshot.GetClaimId(podOwnedClaim): drautils.TestClaimWithPodReservations(drautils.TestClaimWithAllocation(podOwnedClaim, podOwnedClaimAllocWrongNode), podWithClaims),
					}, map[string][]*resourceapi.ResourceSlice{node.Name: resourceSlices}, nil, deviceClasses),
			},
		},
		{
			name: "adding a pod referencing a shared claim already at max reservations is an error",
			state: snapshotState{
				draSnapshot: drasnapshot.NewSnapshot(
					map[drasnapshot.ResourceClaimId]*resourceapi.ResourceClaim{
						drasnapshot.GetClaimId(sharedClaim): fullyReservedClaim(drautils.TestClaimWithAllocation(sharedClaim, sharedClaimAlloc)),
					}, nil, nil, deviceClasses),
			},
			op: func(snapshot clustersnapshot.ClusterSnapshot) error {
				podInfo := framework.NewPodInfo(podWithClaims, []*resourceapi.ResourceClaim{
					drautils.TestClaimWithAllocation(podOwnedClaim, podOwnedClaimAlloc),
					fullyReservedClaim(drautils.TestClaimWithAllocation(sharedClaim, sharedClaimAlloc)),
				})
				return snapshot.AddNodeInfo(framework.NewNodeInfo(node, resourceSlices, podInfo))
			},
			// The shared claim referenced by the pod is already at the max reservation count, and no more reservations can be added - this should be an error to match scheduler behavior.
			wantErr: cmpopts.AnyError,
			// The state shouldn't change on error.
			// TODO(DRA): Until transaction-like clean-up is implemented in AddNodeInfo, the state is not cleaned up on error. Make modifiedState identical to initial state after the clean-up is implemented.
			modifiedState: snapshotState{
				draSnapshot: drasnapshot.NewSnapshot(
					map[drasnapshot.ResourceClaimId]*resourceapi.ResourceClaim{
						drasnapshot.GetClaimId(sharedClaim):   fullyReservedClaim(drautils.TestClaimWithAllocation(sharedClaim, sharedClaimAlloc)),
						drasnapshot.GetClaimId(podOwnedClaim): drautils.TestClaimWithAllocation(podOwnedClaim, podOwnedClaimAlloc),
					}, map[string][]*resourceapi.ResourceSlice{node.Name: resourceSlices}, nil, deviceClasses),
			},
		},
		{
			name: "adding a pod referencing its own claim without adding the claim is an error",
			state: snapshotState{
				draSnapshot: drasnapshot.NewSnapshot(
					map[drasnapshot.ResourceClaimId]*resourceapi.ResourceClaim{
						drasnapshot.GetClaimId(sharedClaim): drautils.TestClaimWithAllocation(sharedClaim, sharedClaimAlloc),
					}, nil, nil, deviceClasses),
			},
			op: func(snapshot clustersnapshot.ClusterSnapshot) error {
				podInfo := framework.NewPodInfo(podWithClaims, []*resourceapi.ResourceClaim{
					drautils.TestClaimWithAllocation(sharedClaim, sharedClaimAlloc),
				})
				return snapshot.AddNodeInfo(framework.NewNodeInfo(node, resourceSlices, podInfo))
			},
			// The added pod references a pod-owned claim that isn't present in the PodInfo - this should be an error.
			wantErr: cmpopts.AnyError,
			// The state shouldn't change on error.
			// TODO(DRA): Until transaction-like clean-up is implemented in AddNodeInfo, the state is not cleaned up on error. Make modifiedState identical to initial state after the clean-up is implemented.
			modifiedState: snapshotState{
				draSnapshot: drasnapshot.NewSnapshot(
					map[drasnapshot.ResourceClaimId]*resourceapi.ResourceClaim{
						drasnapshot.GetClaimId(sharedClaim): drautils.TestClaimWithAllocation(sharedClaim, sharedClaimAlloc),
					}, map[string][]*resourceapi.ResourceSlice{node.Name: resourceSlices}, nil, deviceClasses),
			},
		},
		{
			name: "remove nodeInfo with LocalResourceSlices and NeededResourceClaims",
			// Start with a NodeInfo with LocalResourceSlices and pods with NeededResourceClaims in the DRA snapshot.
			// One claim is shared, one is pod-owned.
			state: snapshotState{
				nodes:       []*apiv1.Node{node},
				csiSnapshot: createCSISnapshot(csiNode),
				podsByNode:  map[string][]*apiv1.Pod{node.Name: {withNodeName(podWithClaims, node.Name)}},
				draSnapshot: drasnapshot.NewSnapshot(
					map[drasnapshot.ResourceClaimId]*resourceapi.ResourceClaim{
						drasnapshot.GetClaimId(podOwnedClaim): drautils.TestClaimWithPodReservations(drautils.TestClaimWithAllocation(podOwnedClaim, podOwnedClaimAlloc), podWithClaims),
						drasnapshot.GetClaimId(sharedClaim):   drautils.TestClaimWithPodReservations(drautils.TestClaimWithAllocation(sharedClaim, sharedClaimAlloc), podWithClaims),
					},
					map[string][]*resourceapi.ResourceSlice{node.Name: resourceSlices},
					nil, deviceClasses),
			},
			op: func(snapshot clustersnapshot.ClusterSnapshot) error {
				return snapshot.RemoveNodeInfo(node.Name)
			},
			// LocalResourceSlices for the removed Node should get removed from the DRA snapshot.
			// The pod-owned claim referenced by a pod from the removed Node should get removed from the DRA snapshot.
			// The shared claim referenced by a pod from the removed Node should stay in the DRA snapshot, but the pod's reservation should be removed.
			modifiedState: snapshotState{
				draSnapshot: drasnapshot.NewSnapshot(
					map[drasnapshot.ResourceClaimId]*resourceapi.ResourceClaim{
						drasnapshot.GetClaimId(sharedClaim): drautils.TestClaimWithAllocation(sharedClaim, sharedClaimAlloc),
					}, nil, nil, deviceClasses),
			},
		},
		{
			name: "remove nodeInfo with LocalResourceSlices and NeededResourceClaims, then add it back",
			// Start with a NodeInfo with LocalResourceSlices and pods with NeededResourceClaims in the DRA snapshot.
			state: snapshotState{
				nodes:       []*apiv1.Node{node},
				podsByNode:  map[string][]*apiv1.Pod{node.Name: {withNodeName(podWithClaims, node.Name)}},
				csiSnapshot: createCSISnapshot(csiNode),
				draSnapshot: drasnapshot.NewSnapshot(
					map[drasnapshot.ResourceClaimId]*resourceapi.ResourceClaim{
						drasnapshot.GetClaimId(podOwnedClaim): drautils.TestClaimWithPodReservations(drautils.TestClaimWithAllocation(podOwnedClaim, podOwnedClaimAlloc), podWithClaims),
						drasnapshot.GetClaimId(sharedClaim):   drautils.TestClaimWithPodReservations(drautils.TestClaimWithAllocation(sharedClaim, sharedClaimAlloc), podWithClaims),
					},
					map[string][]*resourceapi.ResourceSlice{node.Name: resourceSlices},
					nil, deviceClasses),
			},
			// Remove the NodeInfo and then add it back to the snapshot.
			op: func(snapshot clustersnapshot.ClusterSnapshot) error {
				if err := snapshot.RemoveNodeInfo(node.Name); err != nil {
					return err
				}
				podInfo := framework.NewPodInfo(podWithClaims, []*resourceapi.ResourceClaim{
					drautils.TestClaimWithAllocation(podOwnedClaim, podOwnedClaimAlloc),
					drautils.TestClaimWithAllocation(sharedClaim, sharedClaimAlloc),
				})
				nodeInfo := framework.NewNodeInfo(node, resourceSlices, podInfo)
				nodeInfo.CSINode = csiNode
				return snapshot.AddNodeInfo(nodeInfo)
			},
			// The state should be identical to the initial one after the modifications.
			modifiedState: snapshotState{
				nodes:      []*apiv1.Node{node},
				podsByNode: map[string][]*apiv1.Pod{node.Name: {podWithClaims}},
				draSnapshot: drasnapshot.NewSnapshot(
					map[drasnapshot.ResourceClaimId]*resourceapi.ResourceClaim{
						drasnapshot.GetClaimId(podOwnedClaim): drautils.TestClaimWithPodReservations(drautils.TestClaimWithAllocation(podOwnedClaim, podOwnedClaimAlloc), podWithClaims),
						drasnapshot.GetClaimId(sharedClaim):   drautils.TestClaimWithPodReservations(drautils.TestClaimWithAllocation(sharedClaim, sharedClaimAlloc), podWithClaims),
					},
					map[string][]*resourceapi.ResourceSlice{node.Name: resourceSlices}, nil, deviceClasses),
				csiSnapshot: createCSISnapshot(csiNode),
			},
		},
		{
			name: "removing LocalResourceSlices for a non-existing Node is an error",
			state: snapshotState{
				draSnapshot: drasnapshot.NewSnapshot(nil, map[string][]*resourceapi.ResourceSlice{node.Name: resourceSlices}, nil, deviceClasses),
			},
			op: func(snapshot clustersnapshot.ClusterSnapshot) error {
				return snapshot.RemoveNodeInfo("wrong-name")
			},
			// The removed Node isn't in the snapshot, so this should be an error.
			wantErr: cmpopts.AnyError,
			// The state shouldn't change on error.
			modifiedState: snapshotState{
				draSnapshot: drasnapshot.NewSnapshot(nil, map[string][]*resourceapi.ResourceSlice{node.Name: resourceSlices}, nil, deviceClasses),
			},
		},
		{
			name: "schedule pod with NeededResourceClaims to an existing nodeInfo",
			// Start with a NodeInfo with LocalResourceSlices but no Pods. The DRA snapshot already tracks all the claims
			// that the pod references, but they aren't allocated yet.
			state: snapshotState{
				nodes:       []*apiv1.Node{node},
				csiSnapshot: createCSISnapshot(csiNode),
				draSnapshot: drasnapshot.NewSnapshot(
					map[drasnapshot.ResourceClaimId]*resourceapi.ResourceClaim{
						drasnapshot.GetClaimId(podOwnedClaim): podOwnedClaim.DeepCopy(),
						drasnapshot.GetClaimId(sharedClaim):   sharedClaim.DeepCopy(),
					},
					map[string][]*resourceapi.ResourceSlice{node.Name: resourceSlices}, nil, deviceClasses),
			},
			// Run SchedulePod, which should allocate the claims in the DRA snapshot via the DRA scheduler plugin.
			op: func(snapshot clustersnapshot.ClusterSnapshot) error {
				return snapshot.SchedulePod(podWithClaims, node.Name)
			},
			// The pod should get added to the Node.
			// The claims referenced by the Pod should get allocated and reserved for the Pod.
			modifiedState: snapshotState{
				nodes:       []*apiv1.Node{node},
				csiSnapshot: createCSISnapshot(csiNode),
				podsByNode:  map[string][]*apiv1.Pod{node.Name: {podWithClaims}},
				draSnapshot: drasnapshot.NewSnapshot(
					map[drasnapshot.ResourceClaimId]*resourceapi.ResourceClaim{
						drasnapshot.GetClaimId(podOwnedClaim): drautils.TestClaimWithPodReservations(drautils.TestClaimWithAllocation(podOwnedClaim, podOwnedClaimAlloc), podWithClaims),
						drasnapshot.GetClaimId(sharedClaim):   drautils.TestClaimWithPodReservations(drautils.TestClaimWithAllocation(sharedClaim, sharedClaimAlloc), podWithClaims),
					},
					map[string][]*resourceapi.ResourceSlice{node.Name: resourceSlices}, nil, deviceClasses),
			},
		},
		{
			name: "schedule pod with NeededResourceClaims (some of them shared and already allocated) to an existing nodeInfo",
			// Start with a NodeInfo with LocalResourceSlices but no Pods. The DRA snapshot already tracks all the claims
			// that the pod references. The shared claim is already allocated, the pod-owned one isn't yet.
			state: snapshotState{
				nodes:       []*apiv1.Node{node},
				csiSnapshot: createCSISnapshot(csiNode),
				draSnapshot: drasnapshot.NewSnapshot(
					map[drasnapshot.ResourceClaimId]*resourceapi.ResourceClaim{
						drasnapshot.GetClaimId(podOwnedClaim): podOwnedClaim.DeepCopy(),
						drasnapshot.GetClaimId(sharedClaim):   drautils.TestClaimWithAllocation(sharedClaim, sharedClaimAlloc),
					},
					map[string][]*resourceapi.ResourceSlice{node.Name: resourceSlices}, nil, deviceClasses),
			},
			// Run SchedulePod, which should allocate the pod-owned claim in the DRA snapshot via the DRA scheduler plugin.
			op: func(snapshot clustersnapshot.ClusterSnapshot) error {
				return snapshot.SchedulePod(podWithClaims, node.Name)
			},
			// The pod should get added to the Node.
			// The pod-owned claim referenced by the Pod should get allocated. Both claims referenced by the Pod should get reserved for the Pod.
			modifiedState: snapshotState{
				nodes:       []*apiv1.Node{node},
				csiSnapshot: createCSISnapshot(csiNode),
				podsByNode:  map[string][]*apiv1.Pod{node.Name: {podWithClaims}},
				draSnapshot: drasnapshot.NewSnapshot(
					map[drasnapshot.ResourceClaimId]*resourceapi.ResourceClaim{
						drasnapshot.GetClaimId(podOwnedClaim): drautils.TestClaimWithPodReservations(drautils.TestClaimWithAllocation(podOwnedClaim, podOwnedClaimAlloc), podWithClaims),
						drasnapshot.GetClaimId(sharedClaim):   drautils.TestClaimWithPodReservations(drautils.TestClaimWithAllocation(sharedClaim, sharedClaimAlloc), podWithClaims),
					},
					map[string][]*resourceapi.ResourceSlice{node.Name: resourceSlices}, nil, deviceClasses),
			},
		},
		{
			name: "scheduling pod with failing DRA predicates is an error",
			// Start with a NodeInfo with LocalResourceSlices but no Pods. The DRA snapshot doesn't track one of the claims
			// referenced by the Pod we're trying to schedule.
			state: snapshotState{
				nodes:       []*apiv1.Node{node},
				csiSnapshot: createCSISnapshot(csiNode),
				draSnapshot: drasnapshot.NewSnapshot(
					map[drasnapshot.ResourceClaimId]*resourceapi.ResourceClaim{
						drasnapshot.GetClaimId(sharedClaim): sharedClaim.DeepCopy(),
					},
					map[string][]*resourceapi.ResourceSlice{node.Name: resourceSlices}, nil, deviceClasses),
			},
			op: func(snapshot clustersnapshot.ClusterSnapshot) error {
				return snapshot.SchedulePod(podWithClaims, node.Name)
			},
			// SchedulePod should fail at checking scheduling predicates, because the DRA plugin can't find one of the claims.
			wantErr: clustersnapshot.NewFailingPredicateError(nil, "", nil, "", ""), // Only the type of the error is asserted (via cmp.EquateErrors() and errors.Is()), so the parameters don't matter here.
			// The state shouldn't change on error.
			modifiedState: snapshotState{
				nodes:       []*apiv1.Node{node},
				csiSnapshot: createCSISnapshot(csiNode),
				draSnapshot: drasnapshot.NewSnapshot(
					map[drasnapshot.ResourceClaimId]*resourceapi.ResourceClaim{
						drasnapshot.GetClaimId(sharedClaim): sharedClaim,
					},
					map[string][]*resourceapi.ResourceSlice{node.Name: resourceSlices}, nil, deviceClasses),
			},
		},
		{
			name: "scheduling pod referencing a shared claim already at max reservations is an error",
			// Start with a NodeInfo with LocalResourceSlices but no Pods. The DRA snapshot already tracks all the claims
			// that the pod references. The shared claim is already allocated and at max reservations.
			state: snapshotState{
				nodes:       []*apiv1.Node{node},
				csiSnapshot: createCSISnapshot(csiNode),
				draSnapshot: drasnapshot.NewSnapshot(
					map[drasnapshot.ResourceClaimId]*resourceapi.ResourceClaim{
						drasnapshot.GetClaimId(podOwnedClaim): podOwnedClaim.DeepCopy(),
						drasnapshot.GetClaimId(sharedClaim):   fullyReservedClaim(drautils.TestClaimWithAllocation(sharedClaim, sharedClaimAlloc)),
					},
					map[string][]*resourceapi.ResourceSlice{node.Name: resourceSlices}, nil, deviceClasses),
			},
			op: func(snapshot clustersnapshot.ClusterSnapshot) error {
				return snapshot.SchedulePod(podWithClaims, node.Name)
			},
			// The shared claim referenced by the pod is already at the max reservation count, and no more reservations can be added - this should be an error to match scheduler behavior.
			wantErr: clustersnapshot.NewSchedulingInternalError(nil, ""), // Only the type of the error is asserted (via cmp.EquateErrors() and errors.Is()), so the parameters don't matter here.
			// The state shouldn't change on error.
			// TODO(DRA): Until transaction-like clean-up is implemented in SchedulePod, the state is not cleaned up on error. Make modifiedState identical to initial state after the clean-up is implemented.
			modifiedState: snapshotState{
				nodes:       []*apiv1.Node{node},
				csiSnapshot: createCSISnapshot(csiNode),
				draSnapshot: drasnapshot.NewSnapshot(
					map[drasnapshot.ResourceClaimId]*resourceapi.ResourceClaim{
						drasnapshot.GetClaimId(podOwnedClaim): drautils.TestClaimWithAllocation(podOwnedClaim, podOwnedClaimAlloc),
						drasnapshot.GetClaimId(sharedClaim):   fullyReservedClaim(drautils.TestClaimWithAllocation(sharedClaim, sharedClaimAlloc)),
					},
					map[string][]*resourceapi.ResourceSlice{node.Name: resourceSlices}, nil, deviceClasses),
			},
		},
		{
			name: "schedule pod with NeededResourceClaims to any Node (only one Node has ResourceSlices)",
			// Start with a NodeInfo with LocalResourceSlices but no Pods, plus some other Nodes that don't have any slices. The DRA snapshot already tracks all the claims
			// that the pod references, but they aren't allocated yet.
			state: snapshotState{
				nodes:       []*apiv1.Node{otherNode, node, largeNode},
				csiSnapshot: createCSISnapshot(csiNode, otherCSINode, largeCSINode),
				draSnapshot: drasnapshot.NewSnapshot(
					map[drasnapshot.ResourceClaimId]*resourceapi.ResourceClaim{
						drasnapshot.GetClaimId(podOwnedClaim): podOwnedClaim.DeepCopy(),
						drasnapshot.GetClaimId(sharedClaim):   sharedClaim.DeepCopy(),
					},
					map[string][]*resourceapi.ResourceSlice{node.Name: resourceSlices}, nil, deviceClasses),
			},
			// Run SchedulePod, which should allocate the claims in the DRA snapshot via the DRA scheduler plugin.
			op: func(snapshot clustersnapshot.ClusterSnapshot) error {
				foundNodeName, err := snapshot.SchedulePodOnAnyNodeMatching(podWithClaims, func(_ *framework.NodeInfo) bool { return true })
				if diff := cmp.Diff(node.Name, foundNodeName); diff != "" {
					t.Errorf("SchedulePodOnAnyNodeMatching(): unexpected output (-want +got): %s", diff)
				}
				return err
			},
			// The pod should get added to the Node.
			// The claims referenced by the Pod should get allocated and reserved for the Pod.
			modifiedState: snapshotState{
				nodes:       []*apiv1.Node{otherNode, node, largeNode},
				csiSnapshot: createCSISnapshot(csiNode, otherCSINode, largeCSINode),
				podsByNode:  map[string][]*apiv1.Pod{node.Name: {podWithClaims}},
				draSnapshot: drasnapshot.NewSnapshot(
					map[drasnapshot.ResourceClaimId]*resourceapi.ResourceClaim{
						drasnapshot.GetClaimId(podOwnedClaim): drautils.TestClaimWithPodReservations(drautils.TestClaimWithAllocation(podOwnedClaim, podOwnedClaimAlloc), podWithClaims),
						drasnapshot.GetClaimId(sharedClaim):   drautils.TestClaimWithPodReservations(drautils.TestClaimWithAllocation(sharedClaim, sharedClaimAlloc), podWithClaims),
					},
					map[string][]*resourceapi.ResourceSlice{node.Name: resourceSlices}, nil, deviceClasses),
			},
		},
		{
			name: "scheduling pod on any Node with failing DRA predicates is an error",
			// Start with a NodeInfo with LocalResourceSlices but no Pods, plus some other Nodes that don't have any slices. The DRA snapshot doesn't track one of the claims
			// referenced by the Pod we're trying to schedule.
			state: snapshotState{
				nodes:       []*apiv1.Node{otherNode, node, largeNode},
				csiSnapshot: createCSISnapshot(csiNode, otherCSINode, largeCSINode),
				draSnapshot: drasnapshot.NewSnapshot(
					map[drasnapshot.ResourceClaimId]*resourceapi.ResourceClaim{
						drasnapshot.GetClaimId(sharedClaim): sharedClaim.DeepCopy(),
					},
					map[string][]*resourceapi.ResourceSlice{node.Name: resourceSlices}, nil, deviceClasses),
			},
			op: func(snapshot clustersnapshot.ClusterSnapshot) error {
				foundNodeName, err := snapshot.SchedulePodOnAnyNodeMatching(podWithClaims, func(_ *framework.NodeInfo) bool { return true })
				if foundNodeName != "" {
					t.Errorf("SchedulePodOnAnyNodeMatching(): unexpected output: want empty string, got %q", foundNodeName)
				}
				return err
			},
			// SchedulePodOnAnyNodeMatching should fail at checking scheduling predicates for every Node, because the DRA plugin can't find one of the claims.
			wantErr: clustersnapshot.NewFailingPredicateError(nil, "", nil, "", ""), // Only the type of the error is asserted (via cmp.EquateErrors() and errors.Is()), so the parameters don't matter here.
			// The state shouldn't change on error.
			modifiedState: snapshotState{
				nodes:       []*apiv1.Node{otherNode, node, largeNode},
				csiSnapshot: createCSISnapshot(csiNode, otherCSINode, largeCSINode),
				draSnapshot: drasnapshot.NewSnapshot(
					map[drasnapshot.ResourceClaimId]*resourceapi.ResourceClaim{
						drasnapshot.GetClaimId(sharedClaim): sharedClaim,
					},
					map[string][]*resourceapi.ResourceSlice{node.Name: resourceSlices}, nil, deviceClasses),
			},
		},
		{
			name: "scheduling pod referencing a shared claim already at max reservations on any Node is an error",
			// Start with a NodeInfo with LocalResourceSlices but no Pods, plus some other Nodes that don't have any slices. The DRA snapshot already tracks all the claims
			// that the pod references. The shared claim is already allocated and at max reservations.
			state: snapshotState{
				nodes:       []*apiv1.Node{otherNode, node, largeNode},
				csiSnapshot: createCSISnapshot(csiNode, otherCSINode, largeCSINode),
				draSnapshot: drasnapshot.NewSnapshot(
					map[drasnapshot.ResourceClaimId]*resourceapi.ResourceClaim{
						drasnapshot.GetClaimId(podOwnedClaim): podOwnedClaim.DeepCopy(),
						drasnapshot.GetClaimId(sharedClaim):   fullyReservedClaim(drautils.TestClaimWithAllocation(sharedClaim, sharedClaimAlloc)),
					},
					map[string][]*resourceapi.ResourceSlice{node.Name: resourceSlices}, nil, deviceClasses),
			},
			op: func(snapshot clustersnapshot.ClusterSnapshot) error {
				foundNodeName, err := snapshot.SchedulePodOnAnyNodeMatching(podWithClaims, func(_ *framework.NodeInfo) bool { return true })
				if foundNodeName != "" {
					t.Errorf("SchedulePodOnAnyNodeMatching(): unexpected output: want empty string, got %q", foundNodeName)
				}
				return err
			},
			// SchedulePodOnAnyNodeMatching should fail at trying to add a reservation to the shared claim for every Node.
			wantErr: clustersnapshot.NewSchedulingInternalError(nil, ""), // Only the type of the error is asserted (via cmp.EquateErrors() and errors.Is()), so the parameters don't matter here.
			// The state shouldn't change on error.
			// TODO(DRA): Until transaction-like clean-up is implemented in SchedulePodOnAnyNodeMatching, the state is not cleaned up on error. Make modifiedState identical to initial state after the clean-up is implemented.
			modifiedState: snapshotState{
				nodes:       []*apiv1.Node{otherNode, node, largeNode},
				csiSnapshot: createCSISnapshot(csiNode, otherCSINode, largeCSINode),
				draSnapshot: drasnapshot.NewSnapshot(
					map[drasnapshot.ResourceClaimId]*resourceapi.ResourceClaim{
						drasnapshot.GetClaimId(podOwnedClaim): drautils.TestClaimWithAllocation(podOwnedClaim, podOwnedClaimAlloc),
						drasnapshot.GetClaimId(sharedClaim):   fullyReservedClaim(drautils.TestClaimWithAllocation(sharedClaim, sharedClaimAlloc)),
					},
					map[string][]*resourceapi.ResourceSlice{node.Name: resourceSlices}, nil, deviceClasses),
			},
		},
		{
			name: "unschedule Pod with NeededResourceClaims",
			// Start with a Pod already scheduled on a Node. The pod references a pod-owned and a shared claim, both used only by the pod.
			state: snapshotState{
				nodes:       []*apiv1.Node{node},
				csiSnapshot: createCSISnapshot(csiNode),
				podsByNode:  map[string][]*apiv1.Pod{node.Name: {withNodeName(pod, node.Name), withNodeName(podWithClaims, node.Name)}},
				draSnapshot: drasnapshot.NewSnapshot(
					map[drasnapshot.ResourceClaimId]*resourceapi.ResourceClaim{
						drasnapshot.GetClaimId(podOwnedClaim): drautils.TestClaimWithPodReservations(drautils.TestClaimWithAllocation(podOwnedClaim, podOwnedClaimAlloc), podWithClaims),
						drasnapshot.GetClaimId(sharedClaim):   drautils.TestClaimWithPodReservations(drautils.TestClaimWithAllocation(sharedClaim, sharedClaimAlloc), podWithClaims),
					},
					map[string][]*resourceapi.ResourceSlice{node.Name: resourceSlices}, nil, deviceClasses),
			},
			op: func(snapshot clustersnapshot.ClusterSnapshot) error {
				return snapshot.UnschedulePod(podWithClaims.Namespace, podWithClaims.Name, node.Name)
			},
			// The unscheduled pod should be removed from the Node.
			// The claims referenced by the pod should stay in the DRA snapshot, but the pod's reservations should get removed, and the claims should be deallocated.
			modifiedState: snapshotState{
				nodes:       []*apiv1.Node{node},
				csiSnapshot: createCSISnapshot(csiNode),
				podsByNode:  map[string][]*apiv1.Pod{node.Name: {withNodeName(pod, node.Name)}},
				draSnapshot: drasnapshot.NewSnapshot(
					map[drasnapshot.ResourceClaimId]*resourceapi.ResourceClaim{
						drasnapshot.GetClaimId(podOwnedClaim): podOwnedClaim,
						drasnapshot.GetClaimId(sharedClaim):   sharedClaim,
					},
					map[string][]*resourceapi.ResourceSlice{node.Name: resourceSlices}, nil, deviceClasses),
			},
		},
		{
			name: "unschedule Pod with NeededResourceClaims and schedule it back",
			// Start with a Pod already scheduled on a Node. The pod references a pod-owned and a shared claim, both used only by the pod.
			state: snapshotState{
				nodes:       []*apiv1.Node{node},
				csiSnapshot: createCSISnapshot(csiNode),
				podsByNode:  map[string][]*apiv1.Pod{node.Name: {withNodeName(pod, node.Name), withNodeName(podWithClaims, node.Name)}},
				draSnapshot: drasnapshot.NewSnapshot(
					map[drasnapshot.ResourceClaimId]*resourceapi.ResourceClaim{
						drasnapshot.GetClaimId(podOwnedClaim): drautils.TestClaimWithPodReservations(drautils.TestClaimWithAllocation(podOwnedClaim, podOwnedClaimAlloc), podWithClaims),
						drasnapshot.GetClaimId(sharedClaim):   drautils.TestClaimWithPodReservations(drautils.TestClaimWithAllocation(sharedClaim, sharedClaimAlloc), podWithClaims),
					},
					map[string][]*resourceapi.ResourceSlice{node.Name: resourceSlices}, nil, deviceClasses),
			},
			op: func(snapshot clustersnapshot.ClusterSnapshot) error {
				if err := snapshot.UnschedulePod(podWithClaims.Namespace, podWithClaims.Name, node.Name); err != nil {
					return err
				}
				return snapshot.SchedulePod(withNodeName(podWithClaims, node.Name), node.Name)
			},
			// The state shouldn't change.
			modifiedState: snapshotState{
				nodes:       []*apiv1.Node{node},
				csiSnapshot: createCSISnapshot(csiNode),
				podsByNode:  map[string][]*apiv1.Pod{node.Name: {withNodeName(pod, node.Name), withNodeName(podWithClaims, node.Name)}},
				draSnapshot: drasnapshot.NewSnapshot(
					map[drasnapshot.ResourceClaimId]*resourceapi.ResourceClaim{
						drasnapshot.GetClaimId(podOwnedClaim): drautils.TestClaimWithPodReservations(drautils.TestClaimWithAllocation(podOwnedClaim, podOwnedClaimAlloc), podWithClaims),
						drasnapshot.GetClaimId(sharedClaim):   drautils.TestClaimWithPodReservations(drautils.TestClaimWithAllocation(sharedClaim, sharedClaimAlloc), podWithClaims),
					},
					map[string][]*resourceapi.ResourceSlice{node.Name: resourceSlices}, nil, deviceClasses),
			},
		},
		{
			name: "unschedule Pod with NeededResourceClaims (some are shared and still used by other pods)",
			// Start with a Pod already scheduled on a Node. The pod references a pod-owned and a shared claim used by other pods.
			state: snapshotState{
				nodes:       []*apiv1.Node{node},
				csiSnapshot: createCSISnapshot(csiNode),
				podsByNode:  map[string][]*apiv1.Pod{node.Name: {withNodeName(pod, node.Name), withNodeName(podWithClaims, node.Name)}},
				draSnapshot: drasnapshot.NewSnapshot(
					map[drasnapshot.ResourceClaimId]*resourceapi.ResourceClaim{
						drasnapshot.GetClaimId(podOwnedClaim): drautils.TestClaimWithPodReservations(drautils.TestClaimWithAllocation(podOwnedClaim, podOwnedClaimAlloc), podWithClaims),
						drasnapshot.GetClaimId(sharedClaim):   drautils.TestClaimWithPodReservations(drautils.TestClaimWithAllocation(sharedClaim, sharedClaimAlloc), podWithClaims, pod),
					},
					map[string][]*resourceapi.ResourceSlice{node.Name: resourceSlices}, nil, deviceClasses),
			},
			op: func(snapshot clustersnapshot.ClusterSnapshot) error {
				return snapshot.UnschedulePod(podWithClaims.Namespace, podWithClaims.Name, node.Name)
			},
			// The unscheduled pod should be removed from the Node.
			// The claims referenced by the pod should stay in the DRA snapshot, but the pod's reservations should get removed.
			// The pod-owned claim should get deallocated, but the shared one shouldn't.
			modifiedState: snapshotState{
				nodes:       []*apiv1.Node{node},
				csiSnapshot: createCSISnapshot(csiNode),
				podsByNode:  map[string][]*apiv1.Pod{node.Name: {withNodeName(pod, node.Name)}},
				draSnapshot: drasnapshot.NewSnapshot(
					map[drasnapshot.ResourceClaimId]*resourceapi.ResourceClaim{
						drasnapshot.GetClaimId(podOwnedClaim): podOwnedClaim,
						drasnapshot.GetClaimId(sharedClaim):   drautils.TestClaimWithPodReservations(drautils.TestClaimWithAllocation(sharedClaim, sharedClaimAlloc), pod),
					},
					map[string][]*resourceapi.ResourceSlice{node.Name: resourceSlices}, nil, deviceClasses),
			},
		},
		{
			name: "unschedule Pod with NeededResourceClaims (some are shared and still used by other pods) and schedule it back",
			// Start with a Pod with NeededResourceClaims already scheduled on a Node. The pod references a pod-owned and a shared claim used by other pods.
			state: snapshotState{
				nodes:       []*apiv1.Node{node},
				csiSnapshot: createCSISnapshot(csiNode),
				podsByNode:  map[string][]*apiv1.Pod{node.Name: {withNodeName(pod, node.Name), withNodeName(podWithClaims, node.Name)}},
				draSnapshot: drasnapshot.NewSnapshot(
					map[drasnapshot.ResourceClaimId]*resourceapi.ResourceClaim{
						drasnapshot.GetClaimId(podOwnedClaim): drautils.TestClaimWithPodReservations(drautils.TestClaimWithAllocation(podOwnedClaim, podOwnedClaimAlloc), podWithClaims),
						drasnapshot.GetClaimId(sharedClaim):   drautils.TestClaimWithPodReservations(drautils.TestClaimWithAllocation(sharedClaim, sharedClaimAlloc), podWithClaims, pod),
					},
					map[string][]*resourceapi.ResourceSlice{node.Name: resourceSlices}, nil, deviceClasses),
			},
			op: func(snapshot clustersnapshot.ClusterSnapshot) error {
				if err := snapshot.UnschedulePod(podWithClaims.Namespace, podWithClaims.Name, node.Name); err != nil {
					return err
				}
				return snapshot.SchedulePod(withNodeName(podWithClaims, node.Name), node.Name)
			},
			// The state shouldn't change.
			modifiedState: snapshotState{
				nodes:       []*apiv1.Node{node},
				csiSnapshot: createCSISnapshot(csiNode),
				podsByNode:  map[string][]*apiv1.Pod{node.Name: {withNodeName(pod, node.Name), withNodeName(podWithClaims, node.Name)}},
				draSnapshot: drasnapshot.NewSnapshot(
					map[drasnapshot.ResourceClaimId]*resourceapi.ResourceClaim{
						drasnapshot.GetClaimId(podOwnedClaim): drautils.TestClaimWithPodReservations(drautils.TestClaimWithAllocation(podOwnedClaim, podOwnedClaimAlloc), podWithClaims),
						drasnapshot.GetClaimId(sharedClaim):   drautils.TestClaimWithPodReservations(drautils.TestClaimWithAllocation(sharedClaim, sharedClaimAlloc), podWithClaims, pod),
					},
					map[string][]*resourceapi.ResourceSlice{node.Name: resourceSlices}, nil, deviceClasses),
			},
		},
		{
			name: "get/list NodeInfo with DRA objects",
			// Start with a Pod with NeededResourceClaims already scheduled on a Node. The pod references a pod-owned and a shared claim used by other pods. There are other Nodes
			// and pods in the cluster.
			state: snapshotState{
				nodes:       []*apiv1.Node{node, otherNode},
				csiSnapshot: createCSISnapshot(csiNode, otherCSINode),
				podsByNode:  map[string][]*apiv1.Pod{node.Name: {withNodeName(pod, node.Name), withNodeName(podWithClaims, node.Name)}},
				draSnapshot: drasnapshot.NewSnapshot(
					map[drasnapshot.ResourceClaimId]*resourceapi.ResourceClaim{
						drasnapshot.GetClaimId(podOwnedClaim): drautils.TestClaimWithPodReservations(drautils.TestClaimWithAllocation(podOwnedClaim, podOwnedClaimAlloc), podWithClaims),
						drasnapshot.GetClaimId(sharedClaim):   drautils.TestClaimWithPodReservations(drautils.TestClaimWithAllocation(sharedClaim, sharedClaimAlloc), podWithClaims, pod),
					},
					map[string][]*resourceapi.ResourceSlice{node.Name: resourceSlices}, nil, deviceClasses),
			},
			op: func(snapshot clustersnapshot.ClusterSnapshot) error {
				nodeInfoDiffOpts := []cmp.Option{
					// We don't care about this field staying the same, and it differs because it's a global counter bumped on every AddPod.
					cmpopts.IgnoreFields(schedulerimpl.NodeInfo{}, "Generation"),
					cmp.AllowUnexported(framework.NodeInfo{}, schedulerimpl.NodeInfo{}),
					cmpopts.IgnoreUnexported(schedulerimpl.PodInfo{}),
					cmpopts.SortSlices(func(i1, i2 *framework.NodeInfo) bool { return i1.Node().Name < i2.Node().Name }),
					IgnoreObjectOrder[*resourceapi.ResourceClaim](),
					IgnoreObjectOrder[*resourceapi.ResourceSlice](),
				}

				// Verify that GetNodeInfo works as expected.
				nodeInfo, err := snapshot.GetNodeInfo(node.Name)
				if err != nil {
					return err
				}
				wantNodeInfo := framework.NewNodeInfo(node, resourceSlices,
					framework.NewPodInfo(withNodeName(pod, node.Name), nil),
					framework.NewPodInfo(withNodeName(podWithClaims, node.Name), []*resourceapi.ResourceClaim{
						drautils.TestClaimWithPodReservations(drautils.TestClaimWithAllocation(podOwnedClaim, podOwnedClaimAlloc), podWithClaims),
						drautils.TestClaimWithPodReservations(drautils.TestClaimWithAllocation(sharedClaim, sharedClaimAlloc), podWithClaims, pod),
					}),
				)
				wantNodeInfo.CSINode = csiNode

				if diff := cmp.Diff(wantNodeInfo, nodeInfo, nodeInfoDiffOpts...); diff != "" {
					t.Errorf("GetNodeInfo(): unexpected output (-want +got): %s", diff)
				}

				// Verify that ListNodeInfo works as expected.
				nodeInfos, err := snapshot.ListNodeInfos()
				if err != nil {
					return err
				}
				wantNodeInfos := []*framework.NodeInfo{wantNodeInfo, framework.NewTestNodeInfoWithCSI(otherNode, otherCSINode)}
				if diff := cmp.Diff(wantNodeInfos, nodeInfos, nodeInfoDiffOpts...); diff != "" {
					t.Errorf("ListNodeInfos(): unexpected output (-want +got): %s", diff)
				}

				return nil
			},
			// The state shouldn't change.
			modifiedState: snapshotState{
				nodes:       []*apiv1.Node{node, otherNode},
				csiSnapshot: createCSISnapshot(csiNode, otherCSINode),
				podsByNode:  map[string][]*apiv1.Pod{node.Name: {withNodeName(pod, node.Name), withNodeName(podWithClaims, node.Name)}},
				draSnapshot: drasnapshot.NewSnapshot(
					map[drasnapshot.ResourceClaimId]*resourceapi.ResourceClaim{
						drasnapshot.GetClaimId(podOwnedClaim): drautils.TestClaimWithPodReservations(drautils.TestClaimWithAllocation(podOwnedClaim, podOwnedClaimAlloc), podWithClaims),
						drasnapshot.GetClaimId(sharedClaim):   drautils.TestClaimWithPodReservations(drautils.TestClaimWithAllocation(sharedClaim, sharedClaimAlloc), podWithClaims, pod),
					},
					map[string][]*resourceapi.ResourceSlice{node.Name: resourceSlices}, nil, deviceClasses),
			},
		},
		{
			name: "get/list NodeInfo with CSI objects",
			state: snapshotState{
				nodes:       []*apiv1.Node{node, otherNode},
				podsByNode:  map[string][]*apiv1.Pod{node.Name: {withNodeName(pod, node.Name)}},
				csiSnapshot: createCSISnapshot(csiNode, otherCSINode),
			},
			op: func(snapshot clustersnapshot.ClusterSnapshot) error {
				// Verify GetNodeInfo wraps CSINode correctly.
				gotNodeInfo, err := snapshot.GetNodeInfo(node.Name)
				if err != nil {
					return err
				}
				if gotNodeInfo.CSINode == nil {
					t.Errorf("GetNodeInfo(): expected CSINode to be set for node %q, got nil", node.Name)
				} else {
					if diff := cmp.Diff(csiNode, gotNodeInfo.CSINode, cmpopts.IgnoreFields(metav1.ObjectMeta{}, "ResourceVersion")); diff != "" {
						t.Errorf("GetNodeInfo(): unexpected CSINode (-want +got): %s", diff)
					}
				}

				// Verify ListNodeInfos wraps CSINode correctly for all nodes.
				gotNodeInfos, err := snapshot.ListNodeInfos()
				if err != nil {
					return err
				}
				var gotCSINodes []*storagev1.CSINode
				for _, ni := range gotNodeInfos {
					gotCSINodes = append(gotCSINodes, ni.CSINode)
				}

				wantCSINodes := []*storagev1.CSINode{csiNode, otherCSINode}
				ignoreCsiNodeOrderOpt := cmpopts.SortSlices(func(n1, n2 *storagev1.CSINode) bool { return n1.Name < n2.Name })
				if diff := cmp.Diff(wantCSINodes, gotCSINodes, ignoreCsiNodeOrderOpt); diff != "" {
					t.Errorf("ListNodeInfos(): unexpected CSInodes (-want +got): %s", diff)
				}
				return nil
			},
			modifiedState: snapshotState{
				nodes:       []*apiv1.Node{node, otherNode},
				podsByNode:  map[string][]*apiv1.Pod{node.Name: {withNodeName(pod, node.Name)}},
				csiSnapshot: createCSISnapshot(csiNode, otherCSINode),
			},
		},
	}

	for i := range testCases {
		if testCases[i].modifiedState.draSnapshot == nil {
			testCases[i].modifiedState.draSnapshot = drasnapshot.NewEmptySnapshot()
		}

		if testCases[i].state.draSnapshot == nil {
			testCases[i].state.draSnapshot = drasnapshot.NewEmptySnapshot()
		}

		if testCases[i].modifiedState.csiSnapshot == nil {
			testCases[i].modifiedState.csiSnapshot = csisnapshot.NewEmptySnapshot()
		}

		if testCases[i].state.csiSnapshot == nil {
			testCases[i].state.csiSnapshot = csisnapshot.NewEmptySnapshot()
		}
	}

	return testCases
}

func TestForking(t *testing.T) {
	// Uncomment to get logs from the DRA plugin.
	// var fs flag.FlagSet
	// klog.InitFlags(&fs)
	//if err := fs.Set("v", "10"); err != nil {
	//	t.Fatalf("Error while setting higher klog verbosity: %v", err)
	//}
	featuretesting.SetFeatureGateDuringTest(t, feature.DefaultFeatureGate, features.DynamicResourceAllocation, true)

	node := BuildTestNode("specialNode-2", 10, 100)
	for snapshotName, snapshotFactory := range snapshots {
		for _, tc := range validTestCases(t, snapshotName) {
			t.Run(fmt.Sprintf("%s: %s base", snapshotName, tc.name), func(t *testing.T) {
				snapshot := startSnapshot(t, snapshotFactory, tc.state)
				tc.runAndValidateOp(t, snapshot)

				// Modifications should be applied.
				compareStates(t, tc.modifiedState, getSnapshotState(t, snapshot))
			})
			t.Run(fmt.Sprintf("%s: %s fork", snapshotName, tc.name), func(t *testing.T) {
				snapshot := startSnapshot(t, snapshotFactory, tc.state)

				snapshot.Fork()

				tc.runAndValidateOp(t, snapshot)

				// Modifications should be applied.
				compareStates(t, tc.modifiedState, getSnapshotState(t, snapshot))
			})
			t.Run(fmt.Sprintf("%s: %s fork & revert", snapshotName, tc.name), func(t *testing.T) {
				snapshot := startSnapshot(t, snapshotFactory, tc.state)

				snapshot.Fork()

				tc.runAndValidateOp(t, snapshot)

				snapshot.Revert()

				// Modifications should no longer be applied.
				compareStates(t, tc.state, getSnapshotState(t, snapshot))
			})
			t.Run(fmt.Sprintf("%s: %s fork & fork & revert & revert", snapshotName, tc.name), func(t *testing.T) {
				snapshot := startSnapshot(t, snapshotFactory, tc.state)

				snapshot.Fork()
				tc.runAndValidateOp(t, snapshot)
				snapshot.Fork()

				csiNode := BuildCSINode(node)
				snapshot.AddNodeInfo(framework.NewTestNodeInfoWithCSI(node, csiNode))

				snapshot.Revert()
				snapshot.Revert()

				// Modifications should no longer be applied.
				compareStates(t, tc.state, getSnapshotState(t, snapshot))
			})
			t.Run(fmt.Sprintf("%s: %s fork & commit", snapshotName, tc.name), func(t *testing.T) {
				snapshot := startSnapshot(t, snapshotFactory, tc.state)

				snapshot.Fork()
				tc.runAndValidateOp(t, snapshot)

				err := snapshot.Commit()
				assert.NoError(t, err)

				// Modifications should be applied.
				compareStates(t, tc.modifiedState, getSnapshotState(t, snapshot))
			})
			t.Run(fmt.Sprintf("%s: %s fork & fork & commit & revert", snapshotName, tc.name), func(t *testing.T) {
				snapshot := startSnapshot(t, snapshotFactory, tc.state)

				snapshot.Fork()
				snapshot.Fork()
				tc.runAndValidateOp(t, snapshot)

				err := snapshot.Commit()
				assert.NoError(t, err)

				// Modifications should be applied.
				compareStates(t, tc.modifiedState, getSnapshotState(t, snapshot))

				snapshot.Revert()
				// Modifications should no longer be applied.
				compareStates(t, tc.state, getSnapshotState(t, snapshot))

			})
			t.Run(fmt.Sprintf("%s: %s fork & fork & revert & commit", snapshotName, tc.name), func(t *testing.T) {
				snapshot := startSnapshot(t, snapshotFactory, tc.state)

				snapshot.Fork()
				tc.runAndValidateOp(t, snapshot)
				snapshot.Fork()
				csiNode := BuildCSINode(node)
				snapshot.AddNodeInfo(framework.NewTestNodeInfoWithCSI(node, csiNode))
				snapshot.Revert()
				err := snapshot.Commit()
				assert.NoError(t, err)

				// Modifications should be applied.
				compareStates(t, tc.modifiedState, getSnapshotState(t, snapshot))
			})
			t.Run(fmt.Sprintf("%s: %s cache, fork & commit", snapshotName, tc.name), func(t *testing.T) {
				snapshot := startSnapshot(t, snapshotFactory, tc.state)

				// Allow caches to be build.
				_, err := snapshot.NodeInfos().List()
				assert.NoError(t, err)
				_, err = snapshot.NodeInfos().HavePodsWithAffinityList()
				assert.NoError(t, err)

				snapshot.Fork()
				assert.NoError(t, err)

				tc.runAndValidateOp(t, snapshot)

				err = snapshot.Commit()
				assert.NoError(t, err)

				// Modifications should be applied.
				compareStates(t, tc.modifiedState, getSnapshotState(t, snapshot))
			})
		}
	}
}

func TestSetClusterState(t *testing.T) {
	// Run with -count=1 to avoid caching.
	localRand := rand.New(rand.NewSource(time.Now().Unix()))

	nodeCount := localRand.Intn(99) + 1
	podCount := localRand.Intn(999) + 1
	extraNodeCount := localRand.Intn(100)
	extraPodCount := localRand.Intn(1000)

	nodes := clustersnapshot.CreateTestNodes(nodeCount)
	pods := clustersnapshot.CreateTestPods(podCount)
	podsByNode := clustersnapshot.AssignTestPodsToNodes(pods, nodes)

	// Create CSI nodes for all nodes
	csiNodeMap := map[string]*storagev1.CSINode{}
	for _, node := range nodes {
		csiNodeMap[node.Name] = BuildCSINode(node)
	}

	state := snapshotState{nodes: nodes, podsByNode: podsByNode, draSnapshot: drasnapshot.NewEmptySnapshot(), csiSnapshot: csisnapshot.NewSnapshot(csiNodeMap)}

	extraNodes := clustersnapshot.CreateTestNodesWithPrefix("extra", extraNodeCount)

	allNodes := make([]*apiv1.Node, len(nodes)+len(extraNodes), len(nodes)+len(extraNodes))
	copy(allNodes, nodes)
	copy(allNodes[len(nodes):], extraNodes)

	extraPods := clustersnapshot.CreateTestPodsWithPrefix("extra", extraPodCount)
	extraPodsByNode := clustersnapshot.AssignTestPodsToNodes(extraPods, allNodes)

	allPodsByNode := map[string][]*apiv1.Pod{}
	maps.Copy(allPodsByNode, podsByNode)
	for nodeName, extraNodePods := range extraPodsByNode {
		allPodsByNode[nodeName] = append(allPodsByNode[nodeName], extraNodePods...)
	}

	for name, snapshotFactory := range snapshots {
		t.Run(fmt.Sprintf("%s: clear base %d nodes %d pods", name, nodeCount, podCount),
			func(t *testing.T) {
				snapshot := startSnapshot(t, snapshotFactory, state)
				compareStates(t, state, getSnapshotState(t, snapshot))

				assert.NoError(t, snapshot.SetClusterState(nil, nil, nil, nil /*csiSnapshot*/))

				compareStates(t, snapshotState{draSnapshot: drasnapshot.NewEmptySnapshot(), csiSnapshot: csisnapshot.NewEmptySnapshot()}, getSnapshotState(t, snapshot))
			})
		t.Run(fmt.Sprintf("%s: clear base %d nodes %d pods and set a new state", name, nodeCount, podCount),
			func(t *testing.T) {
				snapshot := startSnapshot(t, snapshotFactory, state)
				compareStates(t, state, getSnapshotState(t, snapshot))

				newNodes, newPods := clustersnapshot.CreateTestNodes(13), clustersnapshot.CreateTestPods(37)
				newPodsByNode := clustersnapshot.AssignTestPodsToNodes(newPods, newNodes)

				// Create CSI nodes for new nodes
				newCSINodeMap := map[string]*storagev1.CSINode{}
				for _, node := range newNodes {
					newCSINodeMap[node.Name] = BuildCSINode(node)
				}

				assert.NoError(t, snapshot.SetClusterState(newNodes, newPods, nil, csisnapshot.NewSnapshot(newCSINodeMap)))

				compareStates(t, snapshotState{nodes: newNodes, podsByNode: newPodsByNode, draSnapshot: drasnapshot.NewEmptySnapshot(), csiSnapshot: csisnapshot.NewSnapshot(newCSINodeMap)}, getSnapshotState(t, snapshot))
			})
		t.Run(fmt.Sprintf("%s: clear fork %d nodes %d pods %d extra nodes %d extra pods", name, nodeCount, podCount, extraNodeCount, extraPodCount),
			func(t *testing.T) {
				snapshot := startSnapshot(t, snapshotFactory, state)
				compareStates(t, state, getSnapshotState(t, snapshot))

				snapshot.Fork()

				for _, node := range extraNodes {
					csiNode := BuildCSINode(node)
					err := snapshot.AddNodeInfo(framework.NewTestNodeInfoWithCSI(node, csiNode))
					assert.NoError(t, err)
				}

				for _, pod := range extraPods {
					err := snapshot.ForceAddPod(pod, pod.Spec.NodeName)
					assert.NoError(t, err)
				}

				// Create CSI nodes for all nodes (original + extra)
				allCSINodeMap := map[string]*storagev1.CSINode{}
				for _, node := range allNodes {
					allCSINodeMap[node.Name] = BuildCSINode(node)
				}

				compareStates(t, snapshotState{nodes: allNodes, podsByNode: allPodsByNode, draSnapshot: drasnapshot.NewEmptySnapshot(), csiSnapshot: csisnapshot.NewSnapshot(allCSINodeMap)}, getSnapshotState(t, snapshot))

				assert.NoError(t, snapshot.SetClusterState(nil, nil, nil, nil /*csiSnapshot*/))

				compareStates(t, snapshotState{draSnapshot: drasnapshot.NewEmptySnapshot(), csiSnapshot: csisnapshot.NewEmptySnapshot()}, getSnapshotState(t, snapshot))

				// SetClusterState() should break out of forked state.
				snapshot.Fork()
			})
	}
}

func TestNode404(t *testing.T) {
	// Anything and everything that returns errNodeNotFound should be tested here.
	ops := []struct {
		name string
		op   func(clustersnapshot.ClusterSnapshot) error
	}{
		{"force add pod", func(snapshot clustersnapshot.ClusterSnapshot) error {
			return snapshot.ForceAddPod(BuildTestPod("p1", 0, 0), "node")
		}},
		{"force remove pod", func(snapshot clustersnapshot.ClusterSnapshot) error {
			return snapshot.ForceRemovePod("default", "p1", "node")
		}},
		{"schedule pod", func(snapshot clustersnapshot.ClusterSnapshot) error {
			return snapshot.SchedulePod(BuildTestPod("p1", 0, 0), "node")
		}},
		{"unschedule pod", func(snapshot clustersnapshot.ClusterSnapshot) error {
			return snapshot.UnschedulePod("default", "p1", "node")
		}},
		{"get scheduler NodeInfo", func(snapshot clustersnapshot.ClusterSnapshot) error {
			_, err := snapshot.NodeInfos().Get("node")
			return err
		}},
		{"get internal NodeInfo", func(snapshot clustersnapshot.ClusterSnapshot) error {
			_, err := snapshot.GetNodeInfo("node")
			return err
		}},
		{"remove NodeInfo", func(snapshot clustersnapshot.ClusterSnapshot) error {
			return snapshot.RemoveNodeInfo("node")
		}},
	}

	for name, snapshotFactory := range snapshots {
		for _, op := range ops {
			t.Run(fmt.Sprintf("%s: %s empty", name, op.name),
				func(t *testing.T) {
					snapshot, err := snapshotFactory()
					assert.NoError(t, err)

					// Empty snapshot - shouldn't be able to operate on nodes that are not here.
					err = op.op(snapshot)
					assert.Error(t, err)
				})

			t.Run(fmt.Sprintf("%s: %s fork", name, op.name),
				func(t *testing.T) {
					snapshot, err := snapshotFactory()
					assert.NoError(t, err)

					node := BuildTestNode("node", 10, 100)
					csiNode := BuildCSINode(node)
					err = snapshot.AddNodeInfo(framework.NewTestNodeInfoWithCSI(node, csiNode))
					assert.NoError(t, err)

					snapshot.Fork()
					assert.NoError(t, err)

					err = snapshot.RemoveNodeInfo("node")
					assert.NoError(t, err)

					// Node deleted after fork - shouldn't be able to operate on it.
					err = op.op(snapshot)
					assert.Error(t, err)

					err = snapshot.Commit()
					assert.NoError(t, err)

					// Node deleted before commit - shouldn't be able to operate on it.
					err = op.op(snapshot)
					assert.Error(t, err)
				})

			t.Run(fmt.Sprintf("%s: %s base", name, op.name),
				func(t *testing.T) {
					snapshot, err := snapshotFactory()
					assert.NoError(t, err)

					node := BuildTestNode("node", 10, 100)
					csiNode := BuildCSINode(node)
					err = snapshot.AddNodeInfo(framework.NewTestNodeInfoWithCSI(node, csiNode))
					assert.NoError(t, err)

					err = snapshot.RemoveNodeInfo("node")
					assert.NoError(t, err)

					// Node deleted from base - shouldn't be able to operate on it.
					err = op.op(snapshot)
					assert.Error(t, err)
				})
		}
	}
}

func TestNodeAlreadyExists(t *testing.T) {
	node := BuildTestNode("node", 10, 100)
	csiNode := BuildCSINode(node)
	pod := BuildTestPod("pod", 1, 1)
	pod.Spec.NodeName = node.Name

	ops := []struct {
		name string
		op   func(clustersnapshot.ClusterSnapshot) error
	}{
		{"add scheduler nodeInfo", func(snapshot clustersnapshot.ClusterSnapshot) error {
			nodeInfo := schedulerimpl.NewNodeInfo()
			nodeInfo.SetNode(node)
			return snapshot.AddSchedulerNodeInfo(nodeInfo)
		}},
		{"add internal NodeInfo", func(snapshot clustersnapshot.ClusterSnapshot) error {
			return snapshot.AddNodeInfo(framework.NewTestNodeInfoWithCSI(node, csiNode, pod))
		}},
	}

	for name, snapshotFactory := range snapshots {
		for _, op := range ops {
			t.Run(fmt.Sprintf("%s: %s base", name, op.name),
				func(t *testing.T) {
					snapshot, err := snapshotFactory()
					assert.NoError(t, err)

					err = snapshot.AddNodeInfo(framework.NewTestNodeInfoWithCSI(node, csiNode))
					assert.NoError(t, err)

					// Node already in base.
					err = op.op(snapshot)
					assert.Error(t, err)
				})

			t.Run(fmt.Sprintf("%s: %s base, forked", name, op.name),
				func(t *testing.T) {
					snapshot, err := snapshotFactory()
					assert.NoError(t, err)

					err = snapshot.AddNodeInfo(framework.NewTestNodeInfoWithCSI(node, csiNode))
					assert.NoError(t, err)

					snapshot.Fork()
					assert.NoError(t, err)

					// Node already in base, shouldn't be able to add in fork.
					err = op.op(snapshot)
					assert.Error(t, err)
				})

			t.Run(fmt.Sprintf("%s: %s fork", name, op.name),
				func(t *testing.T) {
					snapshot, err := snapshotFactory()
					assert.NoError(t, err)

					snapshot.Fork()

					err = snapshot.AddNodeInfo(framework.NewTestNodeInfoWithCSI(node, csiNode))
					assert.NoError(t, err)

					// Node already in fork.
					err = op.op(snapshot)
					assert.Error(t, err)
				})
			t.Run(fmt.Sprintf("%s: %s committed", name, op.name),
				func(t *testing.T) {
					snapshot, err := snapshotFactory()
					assert.NoError(t, err)

					snapshot.Fork()

					err = snapshot.AddNodeInfo(framework.NewTestNodeInfoWithCSI(node, csiNode))
					assert.NoError(t, err)

					err = snapshot.Commit()
					assert.NoError(t, err)

					// Node already in new base.
					err = op.op(snapshot)
					assert.Error(t, err)
				})
		}
	}
}

func TestPVCUsedByPods(t *testing.T) {
	node := BuildTestNode("node", 10, 10)
	pod1 := BuildTestPod("pod1", 10, 10)
	pod1.Spec.NodeName = node.Name
	pod1.Spec.Volumes = []apiv1.Volume{
		{
			Name: "v1",
			VolumeSource: apiv1.VolumeSource{
				PersistentVolumeClaim: &apiv1.PersistentVolumeClaimVolumeSource{
					ClaimName: "claim1",
				},
			},
		},
	}
	pod2 := BuildTestPod("pod2", 10, 10)
	pod2.Spec.NodeName = node.Name
	pod2.Spec.Volumes = []apiv1.Volume{
		{
			Name: "v1",
			VolumeSource: apiv1.VolumeSource{
				PersistentVolumeClaim: &apiv1.PersistentVolumeClaimVolumeSource{
					ClaimName: "claim1",
				},
			},
		},
		{
			Name: "v1",
			VolumeSource: apiv1.VolumeSource{
				PersistentVolumeClaim: &apiv1.PersistentVolumeClaimVolumeSource{
					ClaimName: "claim2",
				},
			},
		},
	}
	nonPvcPod := BuildTestPod("pod3", 10, 10)
	nonPvcPod.Spec.NodeName = node.Name
	nonPvcPod.Spec.Volumes = []apiv1.Volume{
		{
			Name: "v1",
			VolumeSource: apiv1.VolumeSource{
				NFS: &apiv1.NFSVolumeSource{
					Server:   "",
					Path:     "",
					ReadOnly: false,
				},
			},
		},
	}
	testcase := []struct {
		desc              string
		node              *apiv1.Node
		pods              []*apiv1.Pod
		claimName         string
		exists            bool
		removePod         string
		existsAfterRemove bool
	}{
		{
			desc:      "pvc new pod with volume fetch",
			node:      node,
			pods:      []*apiv1.Pod{pod1},
			claimName: "claim1",
			exists:    true,
			removePod: "",
		},
		{
			desc:      "pvc new pod with incorrect volume fetch",
			node:      node,
			pods:      []*apiv1.Pod{pod1},
			claimName: "incorrect-claim",
			exists:    false,
			removePod: "",
		},
		{
			desc:      "new pod with non-pvc volume fetch",
			node:      node,
			pods:      []*apiv1.Pod{nonPvcPod},
			claimName: "incorrect-claim",
			exists:    false,
			removePod: "",
		},
		{
			desc:              "pvc new pod with delete volume fetch",
			node:              node,
			pods:              []*apiv1.Pod{pod1},
			claimName:         "claim1",
			exists:            true,
			removePod:         "pod1",
			existsAfterRemove: false,
		},
		{
			desc:              "pvc two pods with duplicated volume, delete one pod, fetch",
			node:              node,
			pods:              []*apiv1.Pod{pod1, pod2},
			claimName:         "claim1",
			exists:            true,
			removePod:         "pod1",
			existsAfterRemove: true,
		},
		{
			desc:              "pvc and non-pvc pods, fetch and delete non-pvc pod",
			node:              node,
			pods:              []*apiv1.Pod{pod1, nonPvcPod},
			claimName:         "claim1",
			exists:            true,
			removePod:         "pod3",
			existsAfterRemove: true,
		},
		{
			desc:              "pvc and non-pvc pods, delete pvc pod and fetch",
			node:              node,
			pods:              []*apiv1.Pod{pod1, nonPvcPod},
			claimName:         "claim1",
			exists:            true,
			removePod:         "pod1",
			existsAfterRemove: false,
		},
	}

	for snapshotName, snapshotFactory := range snapshots {
		for _, tc := range testcase {
			t.Run(fmt.Sprintf("%s with snapshot (%s)", tc.desc, snapshotName), func(t *testing.T) {
				snapshot, err := snapshotFactory()
				assert.NoError(t, err)
				csiNode := BuildCSINode(tc.node)
				err = snapshot.AddNodeInfo(framework.NewTestNodeInfoWithCSI(tc.node, csiNode, tc.pods...))
				assert.NoError(t, err)

				volumeExists := snapshot.StorageInfos().IsPVCUsedByPods(schedulerimpl.GetNamespacedName("default", tc.claimName))
				assert.Equal(t, tc.exists, volumeExists)

				if tc.removePod != "" {
					err = snapshot.ForceRemovePod("default", tc.removePod, "node")
					assert.NoError(t, err)

					volumeExists = snapshot.StorageInfos().IsPVCUsedByPods(schedulerimpl.GetNamespacedName("default", tc.claimName))
					assert.Equal(t, tc.existsAfterRemove, volumeExists)
				}
			})
		}
	}
}

func TestPVCClearAndFork(t *testing.T) {
	node := BuildTestNode("node", 10, 10)
	pod1 := BuildTestPod("pod1", 10, 10)
	pod1.Spec.NodeName = node.Name
	pod1.Spec.Volumes = []apiv1.Volume{
		{
			Name: "v1",
			VolumeSource: apiv1.VolumeSource{
				PersistentVolumeClaim: &apiv1.PersistentVolumeClaimVolumeSource{
					ClaimName: "claim1",
				},
			},
		},
	}
	pod2 := BuildTestPod("pod2", 10, 10)
	pod2.Spec.NodeName = node.Name
	pod2.Spec.Volumes = []apiv1.Volume{
		{
			Name: "v1",
			VolumeSource: apiv1.VolumeSource{
				PersistentVolumeClaim: &apiv1.PersistentVolumeClaimVolumeSource{
					ClaimName: "claim1",
				},
			},
		},
		{
			Name: "v1",
			VolumeSource: apiv1.VolumeSource{
				PersistentVolumeClaim: &apiv1.PersistentVolumeClaimVolumeSource{
					ClaimName: "claim2",
				},
			},
		},
	}
	nonPvcPod := BuildTestPod("pod3", 10, 10)
	nonPvcPod.Spec.NodeName = node.Name
	nonPvcPod.Spec.Volumes = []apiv1.Volume{
		{
			Name: "v1",
			VolumeSource: apiv1.VolumeSource{
				NFS: &apiv1.NFSVolumeSource{
					Server:   "",
					Path:     "",
					ReadOnly: false,
				},
			},
		},
	}

	for snapshotName, snapshotFactory := range snapshots {
		t.Run(fmt.Sprintf("fork and revert snapshot with pvc pods with snapshot: %s", snapshotName), func(t *testing.T) {
			snapshot, err := snapshotFactory()
			assert.NoError(t, err)
			csiNode := BuildCSINode(node)
			err = snapshot.AddNodeInfo(framework.NewTestNodeInfoWithCSI(node, csiNode, pod1))
			assert.NoError(t, err)
			volumeExists := snapshot.StorageInfos().IsPVCUsedByPods(schedulerimpl.GetNamespacedName("default", "claim1"))
			assert.Equal(t, true, volumeExists)

			snapshot.Fork()
			assert.NoError(t, err)
			volumeExists = snapshot.StorageInfos().IsPVCUsedByPods(schedulerimpl.GetNamespacedName("default", "claim1"))
			assert.Equal(t, true, volumeExists)

			err = snapshot.ForceAddPod(pod2, "node")
			assert.NoError(t, err)

			volumeExists = snapshot.StorageInfos().IsPVCUsedByPods(schedulerimpl.GetNamespacedName("default", "claim2"))
			assert.Equal(t, true, volumeExists)

			snapshot.Revert()

			volumeExists = snapshot.StorageInfos().IsPVCUsedByPods(schedulerimpl.GetNamespacedName("default", "claim2"))
			assert.Equal(t, false, volumeExists)

		})

		t.Run(fmt.Sprintf("clear snapshot with pvc pods with snapshot: %s", snapshotName), func(t *testing.T) {
			snapshot, err := snapshotFactory()
			assert.NoError(t, err)
			csiNode := BuildCSINode(node)
			err = snapshot.AddNodeInfo(framework.NewTestNodeInfoWithCSI(node, csiNode, pod1))
			assert.NoError(t, err)
			volumeExists := snapshot.StorageInfos().IsPVCUsedByPods(schedulerimpl.GetNamespacedName("default", "claim1"))
			assert.Equal(t, true, volumeExists)

			assert.NoError(t, snapshot.SetClusterState(nil, nil, nil, nil /*csiSnapshot*/))
			volumeExists = snapshot.StorageInfos().IsPVCUsedByPods(schedulerimpl.GetNamespacedName("default", "claim1"))
			assert.Equal(t, false, volumeExists)

		})
	}
}

func TestWithForkedSnapshot(t *testing.T) {
	// Uncomment to get logs from the DRA plugin.
	// var fs flag.FlagSet
	// klog.InitFlags(&fs)
	//if err := fs.Set("v", "10"); err != nil {
	//	t.Fatalf("Error while setting higher klog verbosity: %v", err)
	//}
	featuretesting.SetFeatureGateDuringTest(t, feature.DefaultFeatureGate, features.DynamicResourceAllocation, true)

	err := fmt.Errorf("some error")
	for snapshotName, snapshotFactory := range snapshots {
		for _, tc := range validTestCases(t, snapshotName) {
			snapshot := startSnapshot(t, snapshotFactory, tc.state)
			successFunc := func() (bool, error) {
				tc.runAndValidateOp(t, snapshot)
				return true, err
			}
			failedFunc := func() (bool, error) {
				tc.runAndValidateOp(t, snapshot)
				return false, err
			}
			t.Run(fmt.Sprintf("%s: %s WithForkedSnapshot for failed function", snapshotName, tc.name), func(t *testing.T) {
				err1, err2 := clustersnapshot.WithForkedSnapshot(snapshot, failedFunc)
				assert.Error(t, err1)
				assert.NoError(t, err2)

				// Modifications should not be applied.
				compareStates(t, tc.state, getSnapshotState(t, snapshot))
			})
			t.Run(fmt.Sprintf("%s: %s WithForkedSnapshot for success function", snapshotName, tc.name), func(t *testing.T) {
				err1, err2 := clustersnapshot.WithForkedSnapshot(snapshot, successFunc)
				assert.Error(t, err1)
				assert.NoError(t, err2)

				// Modifications should be applied.
				compareStates(t, tc.modifiedState, getSnapshotState(t, snapshot))
			})
		}
	}
}

func withNodeName(pod *apiv1.Pod, nodeName string) *apiv1.Pod {
	result := pod.DeepCopy()
	result.Spec.NodeName = nodeName
	return result
}

func createCSISnapshot(csiNodes ...*storagev1.CSINode) *csisnapshot.Snapshot {
	csiNodeMap := map[string]*storagev1.CSINode{}
	for _, csiNode := range csiNodes {
		csiNodeMap[csiNode.Name] = csiNode
	}
	return csisnapshot.NewSnapshot(csiNodeMap)
}

func fullyReservedClaim(claim *resourceapi.ResourceClaim) *resourceapi.ResourceClaim {
	result := claim.DeepCopy()
	for i := range resourceapi.ResourceClaimReservedForMaxSize {
		name := fmt.Sprintf("reserving-pod-%d", i)
		reservingPod := &apiv1.Pod{ObjectMeta: metav1.ObjectMeta{Name: name, UID: types.UID(name + "-uid"), Namespace: "default"}}
		result.Status.ReservedFor = append(result.Status.ReservedFor, drautils.PodClaimConsumerReference(reservingPod))
	}
	return result
}
