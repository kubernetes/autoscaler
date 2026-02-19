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

package snapshot

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	apiv1 "k8s.io/api/core/v1"
	resourceapi "k8s.io/api/resource/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	drautils "k8s.io/autoscaler/cluster-autoscaler/simulator/dynamicresources/utils"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/utils/test"
	schedulerimpl "k8s.io/kubernetes/pkg/scheduler/framework"
)

var (
	node1Name = "node1"
	node2Name = "node2"
	node3Name = "node3"
	trueValue = true

	node1Slice1  = &resourceapi.ResourceSlice{ObjectMeta: metav1.ObjectMeta{Name: "local-slice-1", UID: "local-slice-1"}, Spec: resourceapi.ResourceSliceSpec{NodeName: &node1Name}}
	node1Slice2  = &resourceapi.ResourceSlice{ObjectMeta: metav1.ObjectMeta{Name: "local-slice-2", UID: "local-slice-2"}, Spec: resourceapi.ResourceSliceSpec{NodeName: &node1Name}}
	node2Slice1  = &resourceapi.ResourceSlice{ObjectMeta: metav1.ObjectMeta{Name: "local-slice-3", UID: "local-slice-3"}, Spec: resourceapi.ResourceSliceSpec{NodeName: &node2Name}}
	node2Slice2  = &resourceapi.ResourceSlice{ObjectMeta: metav1.ObjectMeta{Name: "local-slice-4", UID: "local-slice-4"}, Spec: resourceapi.ResourceSliceSpec{NodeName: &node2Name}}
	node3Slice1  = &resourceapi.ResourceSlice{ObjectMeta: metav1.ObjectMeta{Name: "local-slice-5", UID: "local-slice-5"}, Spec: resourceapi.ResourceSliceSpec{NodeName: &node3Name}}
	node3Slice2  = &resourceapi.ResourceSlice{ObjectMeta: metav1.ObjectMeta{Name: "local-slice-6", UID: "local-slice-6"}, Spec: resourceapi.ResourceSliceSpec{NodeName: &node3Name}}
	globalSlice1 = &resourceapi.ResourceSlice{ObjectMeta: metav1.ObjectMeta{Name: "global-slice-1", UID: "global-slice-1"}, Spec: resourceapi.ResourceSliceSpec{AllNodes: &trueValue}}
	globalSlice2 = &resourceapi.ResourceSlice{ObjectMeta: metav1.ObjectMeta{Name: "global-slice-2", UID: "global-slice-2"}, Spec: resourceapi.ResourceSliceSpec{AllNodes: &trueValue}}

	node1 = test.BuildTestNode("node1", 1000, 1000)
	pod1  = test.BuildTestPod("pod1", 1, 1,
		test.WithResourceClaim("ownClaim1", "pod1-ownClaim1-abc", "pod1-ownClaim1"),
		test.WithResourceClaim("ownClaim2", "pod1-ownClaim2-abc", "pod1-ownClaim2"),
		test.WithResourceClaim("sharedClaim1", "sharedClaim1", ""),
		test.WithResourceClaim("sharedClaim2", "sharedClaim2", ""),
	)
	pod2 = test.BuildTestPod("pod2", 1, 1,
		test.WithResourceClaim("ownClaim1", "pod2-ownClaim1-abc", "pod1-ownClaim1"),
		test.WithResourceClaim("sharedClaim1", "sharedClaim1", ""),
		test.WithResourceClaim("sharedClaim3", "sharedClaim3", ""),
	)

	sharedClaim1  = &resourceapi.ResourceClaim{ObjectMeta: metav1.ObjectMeta{Name: "sharedClaim1", UID: "sharedClaim1", Namespace: "default"}}
	sharedClaim2  = &resourceapi.ResourceClaim{ObjectMeta: metav1.ObjectMeta{Name: "sharedClaim2", UID: "sharedClaim2", Namespace: "default"}}
	sharedClaim3  = &resourceapi.ResourceClaim{ObjectMeta: metav1.ObjectMeta{Name: "sharedClaim3", UID: "sharedClaim3", Namespace: "default"}}
	pod1OwnClaim1 = drautils.TestClaimWithPodOwnership(pod1,
		&resourceapi.ResourceClaim{ObjectMeta: metav1.ObjectMeta{Name: "pod1-ownClaim1-abc", UID: "pod1-ownClaim1-abc", Namespace: "default"}},
	)
	pod1OwnClaim2 = drautils.TestClaimWithPodOwnership(pod1,
		&resourceapi.ResourceClaim{ObjectMeta: metav1.ObjectMeta{Name: "pod1-ownClaim2-abc", UID: "pod1-ownClaim2-abc", Namespace: "default"}},
	)
	pod2OwnClaim1 = drautils.TestClaimWithPodOwnership(pod2,
		&resourceapi.ResourceClaim{ObjectMeta: metav1.ObjectMeta{Name: "pod2-ownClaim1-abc", UID: "pod2-ownClaim1-abc", Namespace: "default"}},
	)

	deviceClass1 = &resourceapi.DeviceClass{ObjectMeta: metav1.ObjectMeta{Name: "class1", UID: "class1-uid"}}
	deviceClass2 = &resourceapi.DeviceClass{ObjectMeta: metav1.ObjectMeta{Name: "class2", UID: "class2-uid"}}
)

func TestSnapshotResourceClaims(t *testing.T) {
	pod1NoClaimsInStatus := pod1.DeepCopy()
	pod1NoClaimsInStatus.Status.ResourceClaimStatuses = nil

	for _, tc := range []struct {
		testName string

		claims map[ResourceClaimId]*resourceapi.ResourceClaim

		claimsModFun        func(snapshot *Snapshot) error
		wantClaimsModFunErr error

		pod              *apiv1.Pod
		wantPodClaims    []*resourceapi.ResourceClaim
		wantPodClaimsErr error

		wantAllClaims []*resourceapi.ResourceClaim
	}{
		{
			testName: "PodClaims(): missing pod-owned claim referenced by pod is an error",
			claims: map[ResourceClaimId]*resourceapi.ResourceClaim{
				GetClaimId(sharedClaim1):  sharedClaim1.DeepCopy(),
				GetClaimId(sharedClaim2):  sharedClaim2.DeepCopy(),
				GetClaimId(pod2OwnClaim1): pod2OwnClaim1.DeepCopy(),
				GetClaimId(pod1OwnClaim1): pod1OwnClaim1.DeepCopy(),
			},
			pod:              pod1,
			wantPodClaimsErr: cmpopts.AnyError,
		},
		{
			testName: "PodClaims(): missing shared claim referenced by pod is an error",
			claims: map[ResourceClaimId]*resourceapi.ResourceClaim{
				GetClaimId(sharedClaim1):  sharedClaim1.DeepCopy(),
				GetClaimId(pod2OwnClaim1): pod2OwnClaim1.DeepCopy(),
				GetClaimId(pod1OwnClaim1): pod1OwnClaim1.DeepCopy(),
				GetClaimId(pod1OwnClaim2): pod1OwnClaim2.DeepCopy(),
			},
			pod:              pod1,
			wantPodClaimsErr: cmpopts.AnyError,
		},
		{
			testName: "PodClaims(): claim template set but no claim name in status is an error",
			claims: map[ResourceClaimId]*resourceapi.ResourceClaim{
				GetClaimId(sharedClaim1):  sharedClaim1.DeepCopy(),
				GetClaimId(sharedClaim2):  sharedClaim2.DeepCopy(),
				GetClaimId(pod2OwnClaim1): pod2OwnClaim1.DeepCopy(),
				GetClaimId(pod1OwnClaim1): pod1OwnClaim1.DeepCopy(),
				GetClaimId(pod1OwnClaim2): pod1OwnClaim2.DeepCopy(),
			},
			pod:              pod1NoClaimsInStatus,
			wantPodClaimsErr: cmpopts.AnyError,
		},
		{
			testName: "PodClaims(): all shared and pod-owned claims are returned for a pod",
			claims: map[ResourceClaimId]*resourceapi.ResourceClaim{
				GetClaimId(sharedClaim1):  sharedClaim1.DeepCopy(),
				GetClaimId(sharedClaim2):  sharedClaim2.DeepCopy(),
				GetClaimId(sharedClaim3):  sharedClaim3.DeepCopy(),
				GetClaimId(pod2OwnClaim1): pod2OwnClaim1.DeepCopy(),
				GetClaimId(pod1OwnClaim1): pod1OwnClaim1.DeepCopy(),
				GetClaimId(pod1OwnClaim2): pod1OwnClaim2.DeepCopy(),
			},
			pod:           pod1,
			wantPodClaims: []*resourceapi.ResourceClaim{sharedClaim1, sharedClaim2, pod1OwnClaim1, pod1OwnClaim2},
		},
		{
			testName: "AddClaims(): trying to add a duplicate claim is an error",
			claims: map[ResourceClaimId]*resourceapi.ResourceClaim{
				GetClaimId(sharedClaim1):  sharedClaim1.DeepCopy(),
				GetClaimId(pod1OwnClaim1): pod1OwnClaim1.DeepCopy(),
			},
			claimsModFun: func(snapshot *Snapshot) error {
				return snapshot.AddClaims([]*resourceapi.ResourceClaim{sharedClaim2.DeepCopy(), sharedClaim1.DeepCopy()})
			},
			wantClaimsModFunErr: cmpopts.AnyError,
			wantAllClaims:       []*resourceapi.ResourceClaim{sharedClaim1, pod1OwnClaim1}, // unchanged on error
		},
		{
			testName: "AddClaims(): new claims are correctly added",
			claims: map[ResourceClaimId]*resourceapi.ResourceClaim{
				GetClaimId(sharedClaim1):  sharedClaim1.DeepCopy(),
				GetClaimId(pod1OwnClaim1): pod1OwnClaim1.DeepCopy(),
			},
			claimsModFun: func(snapshot *Snapshot) error {
				if err := snapshot.AddClaims([]*resourceapi.ResourceClaim{sharedClaim2.DeepCopy(), pod2OwnClaim1.DeepCopy()}); err != nil {
					return err
				}
				return snapshot.AddClaims([]*resourceapi.ResourceClaim{sharedClaim3.DeepCopy(), pod1OwnClaim2.DeepCopy()})
			},
			wantAllClaims: []*resourceapi.ResourceClaim{sharedClaim1, sharedClaim2, sharedClaim3, pod1OwnClaim1, pod1OwnClaim2, pod2OwnClaim1}, // 4 new claims added
		},
		{
			testName: "RemovePodOwnedClaims(): pod-owned claims are correctly removed",
			claims: map[ResourceClaimId]*resourceapi.ResourceClaim{
				GetClaimId(sharedClaim1):  sharedClaim1.DeepCopy(),
				GetClaimId(sharedClaim2):  sharedClaim2.DeepCopy(),
				GetClaimId(sharedClaim3):  sharedClaim3.DeepCopy(),
				GetClaimId(pod2OwnClaim1): pod2OwnClaim1.DeepCopy(),
				GetClaimId(pod1OwnClaim1): pod1OwnClaim1.DeepCopy(),
				GetClaimId(pod1OwnClaim2): pod1OwnClaim2.DeepCopy(),
			},
			claimsModFun: func(snapshot *Snapshot) error {
				snapshot.RemovePodOwnedClaims(pod1)
				return nil
			},
			pod:              pod1,
			wantPodClaimsErr: cmpopts.AnyError,
			wantAllClaims:    []*resourceapi.ResourceClaim{sharedClaim1, sharedClaim2, sharedClaim3, pod2OwnClaim1}, // pod1OwnClaim1 and pod1OwnClaim2 removed
		},
		{
			testName: "RemovePodOwnedClaims(): pod reservations in shared claims are correctly removed",
			claims: map[ResourceClaimId]*resourceapi.ResourceClaim{
				GetClaimId(sharedClaim1):  drautils.TestClaimWithPodReservations(sharedClaim1, pod1, pod2),
				GetClaimId(sharedClaim2):  drautils.TestClaimWithPodReservations(sharedClaim2, pod1),
				GetClaimId(sharedClaim3):  drautils.TestClaimWithPodReservations(sharedClaim3, pod2),
				GetClaimId(pod2OwnClaim1): pod2OwnClaim1.DeepCopy(),
				GetClaimId(pod1OwnClaim1): pod1OwnClaim1.DeepCopy(),
				GetClaimId(pod1OwnClaim2): pod1OwnClaim2.DeepCopy(),
			},
			claimsModFun: func(snapshot *Snapshot) error {
				snapshot.RemovePodOwnedClaims(pod1)
				return nil
			},
			pod:              pod1,
			wantPodClaimsErr: cmpopts.AnyError,
			wantAllClaims: []*resourceapi.ResourceClaim{
				drautils.TestClaimWithPodReservations(sharedClaim1, pod2), // pod1 reservation removed
				sharedClaim2, // pod1 reservation removed
				drautils.TestClaimWithPodReservations(sharedClaim3, pod2), // unchanged
				pod2OwnClaim1, // unchanged
			},
		},
		{
			testName: "ReservePodClaims(): missing claims are an error",
			claims: map[ResourceClaimId]*resourceapi.ResourceClaim{
				GetClaimId(sharedClaim1):  drautils.TestClaimWithPodReservations(sharedClaim1, pod2),
				GetClaimId(sharedClaim3):  drautils.TestClaimWithPodReservations(sharedClaim3, pod2),
				GetClaimId(pod2OwnClaim1): pod2OwnClaim1.DeepCopy(),
				GetClaimId(pod1OwnClaim1): drautils.TestClaimWithPodReservations(pod1OwnClaim1, pod1),
				GetClaimId(pod1OwnClaim2): pod1OwnClaim2.DeepCopy(),
			},
			claimsModFun: func(snapshot *Snapshot) error {
				// sharedClaim2 is missing, so this should be an error.
				return snapshot.ReservePodClaims(pod1)
			},
			wantClaimsModFunErr: cmpopts.AnyError,
			wantAllClaims: []*resourceapi.ResourceClaim{ // unchanged on error
				drautils.TestClaimWithPodReservations(sharedClaim1, pod2),
				drautils.TestClaimWithPodReservations(sharedClaim3, pod2),
				pod2OwnClaim1,
				drautils.TestClaimWithPodReservations(pod1OwnClaim1, pod1),
				pod1OwnClaim2,
			},
		},
		{
			testName: "ReservePodClaims(): trying to exceed max reservation limit is an error",
			claims: map[ResourceClaimId]*resourceapi.ResourceClaim{
				GetClaimId(sharedClaim1):  drautils.TestClaimWithPodReservations(sharedClaim1, pod2),
				GetClaimId(sharedClaim2):  drautils.TestClaimWithPodReservations(sharedClaim2, testPods(resourceapi.ResourceClaimReservedForMaxSize)...),
				GetClaimId(sharedClaim3):  drautils.TestClaimWithPodReservations(sharedClaim3, pod2),
				GetClaimId(pod2OwnClaim1): pod2OwnClaim1.DeepCopy(),
				GetClaimId(pod1OwnClaim1): drautils.TestClaimWithPodReservations(pod1OwnClaim1, pod1),
				GetClaimId(pod1OwnClaim2): pod1OwnClaim2.DeepCopy(),
			},
			claimsModFun: func(snapshot *Snapshot) error {
				// sharedClaim2 is missing in claims above, so this should be an error.
				return snapshot.ReservePodClaims(pod1)
			},
			wantClaimsModFunErr: cmpopts.AnyError,
			wantAllClaims: []*resourceapi.ResourceClaim{ // unchanged on error
				drautils.TestClaimWithPodReservations(sharedClaim1, pod2),
				drautils.TestClaimWithPodReservations(sharedClaim2, testPods(resourceapi.ResourceClaimReservedForMaxSize)...),
				drautils.TestClaimWithPodReservations(sharedClaim3, pod2),
				pod2OwnClaim1,
				drautils.TestClaimWithPodReservations(pod1OwnClaim1, pod1),
				pod1OwnClaim2,
			},
		},
		{
			testName: "ReservePodClaims(): pod reservations are correctly added",
			claims: map[ResourceClaimId]*resourceapi.ResourceClaim{
				GetClaimId(sharedClaim1):  drautils.TestClaimWithPodReservations(sharedClaim1, pod2),
				GetClaimId(sharedClaim2):  sharedClaim2.DeepCopy(),
				GetClaimId(sharedClaim3):  drautils.TestClaimWithPodReservations(sharedClaim3, pod2),
				GetClaimId(pod2OwnClaim1): pod2OwnClaim1.DeepCopy(),
				GetClaimId(pod1OwnClaim1): drautils.TestClaimWithPodReservations(pod1OwnClaim1, pod1),
				GetClaimId(pod1OwnClaim2): pod1OwnClaim2.DeepCopy(),
			},
			claimsModFun: func(snapshot *Snapshot) error {
				return snapshot.ReservePodClaims(pod1)
			},
			pod: pod1,
			wantPodClaims: []*resourceapi.ResourceClaim{
				drautils.TestClaimWithPodReservations(sharedClaim1, pod2, pod1),
				drautils.TestClaimWithPodReservations(sharedClaim2, pod1),
				drautils.TestClaimWithPodReservations(pod1OwnClaim1, pod1),
				drautils.TestClaimWithPodReservations(pod1OwnClaim2, pod1),
			},
			wantAllClaims: []*resourceapi.ResourceClaim{
				drautils.TestClaimWithPodReservations(sharedClaim1, pod2, pod1), // pod1 reservation added to another reservation in a shared claim
				drautils.TestClaimWithPodReservations(sharedClaim2, pod1),       // pod1 reservation added to a shared claim
				drautils.TestClaimWithPodReservations(sharedClaim3, pod2),       // unchanged
				pod2OwnClaim1, // unchanged
				drautils.TestClaimWithPodReservations(pod1OwnClaim1, pod1), // unchanged, pod1 reservation already present
				drautils.TestClaimWithPodReservations(pod1OwnClaim2, pod1), // pod1 reservation added to its own claim
			},
		},
		{
			testName: "UnreservePodClaims(): missing claim is an error",
			claims: map[ResourceClaimId]*resourceapi.ResourceClaim{
				GetClaimId(sharedClaim1):  drautils.TestClaimWithPodReservations(sharedClaim1, pod1, pod2),
				GetClaimId(sharedClaim3):  drautils.TestClaimWithPodReservations(sharedClaim3, pod2),
				GetClaimId(pod2OwnClaim1): pod2OwnClaim1.DeepCopy(),
				GetClaimId(pod1OwnClaim1): drautils.TestClaimWithPodReservations(pod1OwnClaim1, pod1),
				GetClaimId(pod1OwnClaim2): pod1OwnClaim2.DeepCopy(),
			},
			claimsModFun: func(snapshot *Snapshot) error {
				// sharedClaim2 is missing in claims above, so this should be an error.
				return snapshot.UnreservePodClaims(pod1)
			},
			wantClaimsModFunErr: cmpopts.AnyError,
			wantAllClaims: []*resourceapi.ResourceClaim{ // unchanged on error
				drautils.TestClaimWithPodReservations(sharedClaim1, pod1, pod2),
				drautils.TestClaimWithPodReservations(sharedClaim3, pod2),
				pod2OwnClaim1,
				drautils.TestClaimWithPodReservations(pod1OwnClaim1, pod1),
				pod1OwnClaim2,
			},
		},
		{
			testName: "UnreservePodClaims(): correctly removes reservations from pod-owned and shared claims",
			claims: map[ResourceClaimId]*resourceapi.ResourceClaim{
				GetClaimId(sharedClaim1):  drautils.TestClaimWithPodReservations(sharedClaim1, pod1, pod2),
				GetClaimId(sharedClaim2):  drautils.TestClaimWithPodReservations(sharedClaim2, pod1),
				GetClaimId(sharedClaim3):  drautils.TestClaimWithPodReservations(sharedClaim3, pod2),
				GetClaimId(pod2OwnClaim1): pod2OwnClaim1.DeepCopy(),
				GetClaimId(pod1OwnClaim1): drautils.TestClaimWithPodReservations(pod1OwnClaim1, pod1),
				GetClaimId(pod1OwnClaim2): pod1OwnClaim2.DeepCopy(),
			},
			claimsModFun: func(snapshot *Snapshot) error {
				return snapshot.UnreservePodClaims(pod1)
			},
			pod: pod1,
			wantPodClaims: []*resourceapi.ResourceClaim{
				drautils.TestClaimWithPodReservations(sharedClaim1, pod2),
				sharedClaim2,
				pod1OwnClaim1,
				pod1OwnClaim2,
			},
			wantAllClaims: []*resourceapi.ResourceClaim{
				drautils.TestClaimWithPodReservations(sharedClaim1, pod2), // pod1 reservation removed, pod2 left
				sharedClaim2, // pod1 reservation removed, none left
				drautils.TestClaimWithPodReservations(sharedClaim3, pod2), // unchanged
				pod2OwnClaim1, // unchanged
				pod1OwnClaim1, // pod1 reservation removed
				pod1OwnClaim2, // unchanged
			},
		},
		{
			testName: "UnreservePodClaims(): correctly clears allocations from pod-owned and unused shared claims",
			claims: map[ResourceClaimId]*resourceapi.ResourceClaim{
				GetClaimId(sharedClaim1):  drautils.TestClaimWithAllocation(drautils.TestClaimWithPodReservations(sharedClaim1, pod1, pod2), nil),
				GetClaimId(sharedClaim2):  drautils.TestClaimWithAllocation(drautils.TestClaimWithPodReservations(sharedClaim2, pod1), nil),
				GetClaimId(sharedClaim3):  drautils.TestClaimWithAllocation(drautils.TestClaimWithPodReservations(sharedClaim3, pod2), nil),
				GetClaimId(pod2OwnClaim1): drautils.TestClaimWithAllocation(pod2OwnClaim1.DeepCopy(), nil),
				GetClaimId(pod1OwnClaim1): drautils.TestClaimWithAllocation(drautils.TestClaimWithPodReservations(pod1OwnClaim1, pod1), nil),
				GetClaimId(pod1OwnClaim2): drautils.TestClaimWithAllocation(pod1OwnClaim2.DeepCopy(), nil),
			},
			claimsModFun: func(snapshot *Snapshot) error {
				return snapshot.UnreservePodClaims(pod1)
			},
			pod: pod1,
			wantPodClaims: []*resourceapi.ResourceClaim{
				drautils.TestClaimWithAllocation(drautils.TestClaimWithPodReservations(sharedClaim1, pod2), nil),
				sharedClaim2,
				pod1OwnClaim1,
				pod1OwnClaim2,
			},
			wantAllClaims: []*resourceapi.ResourceClaim{
				drautils.TestClaimWithAllocation(drautils.TestClaimWithPodReservations(sharedClaim1, pod2), nil), // sharedClaim1 still in use by pod2, so allocation kept
				sharedClaim2, // pod1 reservation removed, none left so allocation cleared
				drautils.TestClaimWithAllocation(drautils.TestClaimWithPodReservations(sharedClaim3, pod2), nil), // unchanged
				drautils.TestClaimWithAllocation(pod2OwnClaim1, nil),                                             // unchanged
				pod1OwnClaim1, // pod1 reservation removed, allocation cleared
				pod1OwnClaim2, // allocation cleared despite lack of reservation
			},
		},
	} {
		t.Run(tc.testName, func(t *testing.T) {
			snapshot := NewSnapshot(tc.claims, nil, nil, nil)

			if tc.claimsModFun != nil {
				err := tc.claimsModFun(snapshot)
				if diff := cmp.Diff(tc.wantClaimsModFunErr, err, cmpopts.EquateErrors()); diff != "" {
					t.Fatalf("Snapshot modification: unexpected error (-want +got): %s", diff)
				}
			}

			if tc.pod != nil {
				podClaims, err := snapshot.PodClaims(tc.pod)
				if diff := cmp.Diff(tc.wantPodClaimsErr, err, cmpopts.EquateErrors()); diff != "" {
					t.Fatalf("Snapshot.PodClaims(): unexpected error (-want +got): %s", diff)
				}
				if diff := cmp.Diff(tc.wantPodClaims, podClaims, cmpopts.EquateEmpty(), test.IgnoreObjectOrder[*resourceapi.ResourceClaim]()); diff != "" {
					t.Errorf("Snapshot.PodClaims(): unexpected output (-want +got): %s", diff)
				}
			}

			if tc.wantAllClaims != nil {
				allClaims, err := snapshot.ResourceClaims().List()
				if err != nil {
					t.Fatalf("ResourceClaims().List(): unexpected error: %v", err)
				}
				if diff := cmp.Diff(tc.wantAllClaims, allClaims, cmpopts.EquateEmpty(), test.IgnoreObjectOrder[*resourceapi.ResourceClaim]()); diff != "" {
					t.Errorf("Snapshot: unexpected ResourceClaim state (-want +got): %s", diff)
				}
			}
		})
	}
}

func TestSnapshotResourceSlices(t *testing.T) {
	localSlices := map[string][]*resourceapi.ResourceSlice{
		"node1": {node1Slice1, node1Slice2},
		"node2": {node2Slice1, node2Slice2},
	}
	globalSlices := []*resourceapi.ResourceSlice{globalSlice1, globalSlice2}
	allSlices := append(globalSlices, node1Slice1, node1Slice2, node2Slice1, node2Slice2)
	extraNode3Slice1 := node3Slice1
	extraNode3Slice2 := node3Slice2

	for _, tc := range []struct {
		testName string

		slicesModFun        func(snapshot *Snapshot) error
		wantSlicesModFunErr error

		nodeName            string
		wantNodeSlices      []*resourceapi.ResourceSlice
		wantNodeSlicesFound bool

		wantAllSlices []*resourceapi.ResourceSlice
	}{
		{
			testName:            "NodeResourceSlices(): unknown nodeName results in found=false",
			nodeName:            "node3",
			wantNodeSlicesFound: false,
		},
		{
			testName:            "NodeResourceSlices(): all node-local slices are correctly returned",
			nodeName:            "node2",
			wantNodeSlicesFound: true,
			wantNodeSlices:      []*resourceapi.ResourceSlice{node2Slice1, node2Slice2},
		},
		{
			testName: "AddNodeResourceSlices(): adding slices for a Node that already has slices tracked is an error",
			slicesModFun: func(snapshot *Snapshot) error {
				return snapshot.AddNodeResourceSlices("node1", []*resourceapi.ResourceSlice{node1Slice1})
			},
			wantSlicesModFunErr: cmpopts.AnyError,
			wantAllSlices:       allSlices,
		},
		{
			testName: "AddNodeResourceSlices(): adding slices for a new Node works correctly",
			slicesModFun: func(snapshot *Snapshot) error {
				return snapshot.AddNodeResourceSlices("node3", []*resourceapi.ResourceSlice{extraNode3Slice1, extraNode3Slice2})
			},
			nodeName:            "node3",
			wantNodeSlicesFound: true,
			wantNodeSlices:      []*resourceapi.ResourceSlice{extraNode3Slice1, extraNode3Slice2},
			wantAllSlices:       append(allSlices, extraNode3Slice1, extraNode3Slice2),
		},
		{
			testName: "RemoveNodeResourceSlices(): removing slices for a non-existing Node is a no-op",
			slicesModFun: func(snapshot *Snapshot) error {
				snapshot.RemoveNodeResourceSlices("node3")
				return nil
			},
			wantAllSlices: allSlices,
		},
		{
			testName: "RemoveNodeResourceSlices(): removing slices for an existing Node works correctly",
			slicesModFun: func(snapshot *Snapshot) error {
				snapshot.RemoveNodeResourceSlices("node2")
				return nil
			},
			wantAllSlices: []*resourceapi.ResourceSlice{node1Slice1, node1Slice2, globalSlice1, globalSlice2},
		},
	} {
		t.Run(tc.testName, func(t *testing.T) {
			snapshot := NewSnapshot(nil, localSlices, globalSlices, nil)

			if tc.slicesModFun != nil {
				err := tc.slicesModFun(snapshot)
				if diff := cmp.Diff(tc.wantSlicesModFunErr, err, cmpopts.EquateErrors()); diff != "" {
					t.Fatalf("Snapshot modification: unexpected error (-want +got): %s", diff)
				}
			}

			if tc.nodeName != "" {
				nodeSlices, found := snapshot.NodeResourceSlices(tc.nodeName)
				if tc.wantNodeSlicesFound != found {
					t.Fatalf("Snapshot.NodeResourceSlices(): unexpected found value: want %v, got %v", tc.wantNodeSlicesFound, found)
				}
				if diff := cmp.Diff(tc.wantNodeSlices, nodeSlices, cmpopts.EquateEmpty(), test.IgnoreObjectOrder[*resourceapi.ResourceSlice]()); diff != "" {
					t.Errorf("Snapshot.NodeResourceSlices(): unexpected output (-want +got): %s", diff)
				}
			}

			if tc.wantAllSlices != nil {
				allSlices, err := snapshot.ResourceSlices().ListWithDeviceTaintRules()
				if err != nil {
					t.Fatalf("ResourceSlices().List(): unexpected error: %v", err)
				}
				if diff := cmp.Diff(tc.wantAllSlices, allSlices, cmpopts.EquateEmpty(), test.IgnoreObjectOrder[*resourceapi.ResourceSlice]()); diff != "" {
					t.Errorf("Snapshot: unexpected ResourceSlice state (-want +got): %s", diff)
				}
			}
		})
	}
}

func TestSnapshotWrapSchedulerNodeInfo(t *testing.T) {
	noClaimsPod1 := test.BuildTestPod("noClaimsPod1", 1, 1)
	noClaimsPod2 := test.BuildTestPod("noClaimsPod2", 1, 1)
	missingClaimPod := test.BuildTestPod("missingClaimPod", 1, 1, test.WithResourceClaim("ref1", "missing-claim-abc", "missing-claim"))
	noSlicesNode := test.BuildTestNode("noSlicesNode", 1000, 1000)

	noDraNodeInfo := schedulerimpl.NewNodeInfo(noClaimsPod1, noClaimsPod2)
	noDraNodeInfo.SetNode(noSlicesNode)

	resourceSlicesNodeInfo := schedulerimpl.NewNodeInfo(noClaimsPod1, noClaimsPod2)
	resourceSlicesNodeInfo.SetNode(node1)

	resourceClaimsNodeInfo := schedulerimpl.NewNodeInfo(pod1, pod2, noClaimsPod1, noClaimsPod2)
	resourceClaimsNodeInfo.SetNode(noSlicesNode)

	fullDraNodeInfo := schedulerimpl.NewNodeInfo(pod1, pod2, noClaimsPod1, noClaimsPod2)
	fullDraNodeInfo.SetNode(node1)

	missingClaimNodeInfo := schedulerimpl.NewNodeInfo(pod1, pod2, noClaimsPod1, noClaimsPod2, missingClaimPod)
	missingClaimNodeInfo.SetNode(node1)

	for _, tc := range []struct {
		testName      string
		schedNodeInfo *schedulerimpl.NodeInfo
		wantNodeInfo  *framework.NodeInfo
		wantErr       error
	}{
		{
			testName:      "no data to add to the wrapper",
			schedNodeInfo: noDraNodeInfo,
			wantNodeInfo:  framework.WrapSchedulerNodeInfo(noDraNodeInfo, nil, nil),
		},
		{
			testName:      "ResourceSlices added to the wrapper",
			schedNodeInfo: resourceSlicesNodeInfo,
			wantNodeInfo:  framework.WrapSchedulerNodeInfo(resourceSlicesNodeInfo, []*resourceapi.ResourceSlice{node1Slice1, node1Slice2}, nil),
		},
		{
			testName:      "ResourceClaims added to the wrapper",
			schedNodeInfo: resourceClaimsNodeInfo,
			wantNodeInfo: framework.WrapSchedulerNodeInfo(resourceClaimsNodeInfo, nil, map[types.UID]framework.PodExtraInfo{
				"pod1": {NeededResourceClaims: []*resourceapi.ResourceClaim{pod1OwnClaim1, pod1OwnClaim2, sharedClaim1, sharedClaim2}},
				"pod2": {NeededResourceClaims: []*resourceapi.ResourceClaim{pod2OwnClaim1, sharedClaim1, sharedClaim3}},
			}),
		},
		{
			testName:      "ResourceSlices and ResourceClaims added to the wrapper",
			schedNodeInfo: fullDraNodeInfo,
			wantNodeInfo: framework.WrapSchedulerNodeInfo(fullDraNodeInfo, []*resourceapi.ResourceSlice{node1Slice1, node1Slice2}, map[types.UID]framework.PodExtraInfo{
				"pod1": {NeededResourceClaims: []*resourceapi.ResourceClaim{pod1OwnClaim1, pod1OwnClaim2, sharedClaim1, sharedClaim2}},
				"pod2": {NeededResourceClaims: []*resourceapi.ResourceClaim{pod2OwnClaim1, sharedClaim1, sharedClaim3}},
			}),
		},
		{
			testName:      "pod in NodeInfo with a missing claim is an error",
			schedNodeInfo: missingClaimNodeInfo,
			wantNodeInfo:  nil,
			wantErr:       cmpopts.AnyError,
		},
	} {
		t.Run(tc.testName, func(t *testing.T) {
			claims := map[ResourceClaimId]*resourceapi.ResourceClaim{
				GetClaimId(sharedClaim1):  sharedClaim1,
				GetClaimId(sharedClaim2):  sharedClaim2,
				GetClaimId(sharedClaim3):  sharedClaim3,
				GetClaimId(pod2OwnClaim1): pod2OwnClaim1,
				GetClaimId(pod1OwnClaim1): pod1OwnClaim1,
				GetClaimId(pod1OwnClaim2): pod1OwnClaim2,
			}
			localSlices := map[string][]*resourceapi.ResourceSlice{
				"node1": {node1Slice1, node1Slice2},
				"node2": {node2Slice1, node2Slice2},
			}
			globalSlices := []*resourceapi.ResourceSlice{globalSlice1, globalSlice2}
			snapshot := NewSnapshot(claims, localSlices, globalSlices, nil)
			nodeInfo, err := snapshot.WrapSchedulerNodeInfo(tc.schedNodeInfo)
			if diff := cmp.Diff(tc.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("Snapshot.WrapSchedulerNodeInfo(): unexpected error (-want +got): %s", diff)
			}
			cmpOpts := []cmp.Option{cmpopts.EquateEmpty(), cmp.AllowUnexported(framework.NodeInfo{}, schedulerimpl.NodeInfo{}),
				cmpopts.IgnoreUnexported(schedulerimpl.PodInfo{}),
				test.IgnoreObjectOrder[*resourceapi.ResourceClaim](), test.IgnoreObjectOrder[*resourceapi.ResourceSlice]()}

			if diff := cmp.Diff(tc.wantNodeInfo, nodeInfo, cmpOpts...); diff != "" {
				t.Errorf("Snapshot.WrapSchedulerNodeInfo(): unexpected output (-want +got): %s", diff)
			}
		})
	}
}

func testPods(count int) []*apiv1.Pod {
	var result []*apiv1.Pod
	for i := range count {
		result = append(result, test.BuildTestPod(fmt.Sprintf("test-pod-%d", i), 1, 1))
	}
	return result
}

func TestSnapshotForkCommitRevert(t *testing.T) {
	initialClaims := map[ResourceClaimId]*resourceapi.ResourceClaim{
		GetClaimId(sharedClaim1):  sharedClaim1.DeepCopy(),
		GetClaimId(pod1OwnClaim1): pod1OwnClaim1.DeepCopy(),
		GetClaimId(pod1OwnClaim2): pod1OwnClaim2.DeepCopy(),
	}
	initialDeviceClasses := map[string]*resourceapi.DeviceClass{deviceClass1.Name: deviceClass1.DeepCopy(), deviceClass2.Name: deviceClass2.DeepCopy()}
	initialLocalSlices := map[string][]*resourceapi.ResourceSlice{*node1Slice1.Spec.NodeName: {node1Slice1.DeepCopy()}}
	initialGlobalSlices := []*resourceapi.ResourceSlice{globalSlice1.DeepCopy(), globalSlice2.DeepCopy()}
	initialState := NewSnapshot(initialClaims, initialLocalSlices, initialGlobalSlices, initialDeviceClasses)

	addedClaim := sharedClaim2.DeepCopy()
	addedNodeSlice := node2Slice1.DeepCopy()
	podToReserve := pod1.DeepCopy()

	modifiedClaims := map[ResourceClaimId]*resourceapi.ResourceClaim{
		GetClaimId(sharedClaim1):  drautils.TestClaimWithPodReservations(sharedClaim1, podToReserve),
		GetClaimId(sharedClaim2):  drautils.TestClaimWithPodReservations(addedClaim, podToReserve),
		GetClaimId(pod1OwnClaim1): drautils.TestClaimWithPodReservations(pod1OwnClaim1, podToReserve),
		GetClaimId(pod1OwnClaim2): drautils.TestClaimWithPodReservations(pod1OwnClaim2, podToReserve),
	}
	modifiedLocalSlices := map[string][]*resourceapi.ResourceSlice{*addedNodeSlice.Spec.NodeName: {addedNodeSlice.DeepCopy()}}
	// Expected state after modifications are applied
	modifiedState := NewSnapshot(
		modifiedClaims,
		modifiedLocalSlices,
		initialGlobalSlices,
		initialDeviceClasses,
	)

	applyModifications := func(t *testing.T, s *Snapshot) {
		t.Helper()

		addedSlices := []*resourceapi.ResourceSlice{addedNodeSlice.DeepCopy()}
		if err := s.AddNodeResourceSlices(*addedNodeSlice.Spec.NodeName, addedSlices); err != nil {
			t.Fatalf("failed to add %s resource slices: %v", *addedNodeSlice.Spec.NodeName, err)
		}
		if err := s.AddClaims([]*resourceapi.ResourceClaim{addedClaim}); err != nil {
			t.Fatalf("failed to add %s claim: %v", addedClaim.Name, err)
		}
		if err := s.ReservePodClaims(podToReserve); err != nil {
			t.Fatalf("failed to reserve claim %s for pod %s: %v", sharedClaim1.Name, podToReserve.Name, err)
		}

		s.RemoveNodeResourceSlices(*node1Slice1.Spec.NodeName)
	}

	compareSnapshots := func(t *testing.T, want, got *Snapshot, msg string) {
		t.Helper()
		if diff := cmp.Diff(want, got, SnapshotFlattenedComparer(), cmpopts.EquateEmpty()); diff != "" {
			t.Errorf("%s: Snapshot state mismatch (-want +got):\n%s", msg, diff)
		}
	}

	t.Run("Fork", func(t *testing.T) {
		snapshot := CloneTestSnapshot(initialState)
		snapshot.Fork()
		applyModifications(t, snapshot)
		compareSnapshots(t, modifiedState, snapshot, "After Fork and Modify")
	})

	t.Run("ForkRevert", func(t *testing.T) {
		snapshot := CloneTestSnapshot(initialState)
		snapshot.Fork()
		applyModifications(t, snapshot)
		snapshot.Revert()
		compareSnapshots(t, initialState, snapshot, "After Fork, Modify, Revert")
	})

	t.Run("ForkCommit", func(t *testing.T) {
		snapshot := CloneTestSnapshot(initialState)
		snapshot.Fork()
		applyModifications(t, snapshot)
		snapshot.Commit()
		compareSnapshots(t, modifiedState, snapshot, "After Fork, Modify, Commit")
	})

	t.Run("ForkForkRevertRevert", func(t *testing.T) {
		snapshot := CloneTestSnapshot(initialState)
		snapshot.Fork()
		applyModifications(t, snapshot)
		snapshot.Fork()

		// Apply further modifications in second fork (add claim3, slice3)
		furtherModifiedClaim3 := sharedClaim3.DeepCopy()
		furtherModifiedSlice3 := node3Slice1.DeepCopy()
		if err := snapshot.AddClaims([]*resourceapi.ResourceClaim{furtherModifiedClaim3}); err != nil {
			t.Fatalf("AddClaims failed in second fork: %v", err)
		}
		if err := snapshot.AddNodeResourceSlices("node3", []*resourceapi.ResourceSlice{furtherModifiedSlice3}); err != nil {
			t.Fatalf("AddNodeResourceSlices failed in second fork: %v", err)
		}

		snapshot.Revert() // Revert second fork
		compareSnapshots(t, modifiedState, snapshot, "After Fork, Modify, Fork, Modify, Revert")

		snapshot.Revert() // Revert first fork
		compareSnapshots(t, initialState, snapshot, "After Fork, Modify, Fork, Modify, Revert, Revert")
	})

	t.Run("ForkForkCommitRevert", func(t *testing.T) {
		snapshot := CloneTestSnapshot(initialState)
		snapshot.Fork()
		snapshot.Fork()
		applyModifications(t, snapshot)
		snapshot.Commit() // Commit second fork into first fork
		compareSnapshots(t, modifiedState, snapshot, "After Fork, Fork, Modify, Commit")

		snapshot.Revert() // Revert first fork (which now contains committed changes)
		compareSnapshots(t, initialState, snapshot, "After Fork, Fork, Modify, Commit, Revert")
	})

	t.Run("ForkForkRevertCommit", func(t *testing.T) {
		snapshot := CloneTestSnapshot(initialState)
		snapshot.Fork()
		applyModifications(t, snapshot)
		snapshot.Fork()
		// Apply further mofications in second fork (add claim3, slice3)
		furtherModifiedClaim3 := sharedClaim3.DeepCopy()
		furtherModifiedSlice3 := node3Slice1.DeepCopy()
		if err := snapshot.AddClaims([]*resourceapi.ResourceClaim{furtherModifiedClaim3}); err != nil {
			t.Fatalf("AddClaims failed in second fork: %v", err)
		}
		if err := snapshot.AddNodeResourceSlices("node3", []*resourceapi.ResourceSlice{furtherModifiedSlice3}); err != nil {
			t.Fatalf("AddNodeResourceSlices failed in second fork: %v", err)
		}

		snapshot.Revert() // Revert second fork
		compareSnapshots(t, modifiedState, snapshot, "After Fork, Modify, Fork, Modify, Revert")

		snapshot.Commit() // Commit first fork (with original modifications)
		compareSnapshots(t, modifiedState, snapshot, "After Fork, Modify, Fork, Modify, Revert, Commit")
	})

	t.Run("CommitNoFork", func(t *testing.T) {
		snapshot := CloneTestSnapshot(initialState)
		snapshot.Commit() // Should be a no-op
		compareSnapshots(t, initialState, snapshot, "After Commit with no Fork")
	})

	t.Run("RevertNoFork", func(t *testing.T) {
		snapshot := CloneTestSnapshot(initialState)
		snapshot.Revert() // Should be a no-op
		compareSnapshots(t, initialState, snapshot, "After Revert with no Fork")
	})

	t.Run("ForkCommitRevert", func(t *testing.T) {
		snapshot := CloneTestSnapshot(initialState)
		snapshot.Fork()
		applyModifications(t, snapshot)
		snapshot.Commit()
		// Now try to revert the committed changes (should be no-op as only base layer exists)
		snapshot.Revert()
		compareSnapshots(t, modifiedState, snapshot, "After Fork, Modify, Commit, Revert")
	})

	t.Run("ForkRevertFork", func(t *testing.T) {
		snapshot := CloneTestSnapshot(initialState)
		snapshot.Fork()
		applyModifications(t, snapshot)
		snapshot.Revert()
		compareSnapshots(t, initialState, snapshot, "After Fork, Modify, Revert")

		snapshot.Fork() // Fork again from the reverted (initial) state
		differentClaim := sharedClaim3.DeepCopy()
		if err := snapshot.AddClaims([]*resourceapi.ResourceClaim{differentClaim}); err != nil {
			t.Fatalf("AddClaims failed in second fork: %v", err)
		}

		expectedState := CloneTestSnapshot(initialState)
		expectedState.AddClaims([]*resourceapi.ResourceClaim{differentClaim.DeepCopy()}) // Apply same change to expected state

		compareSnapshots(t, expectedState, snapshot, "After Fork, Modify, Revert, Fork, Modify")
	})
}
