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

package utils

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	apiv1 "k8s.io/api/core/v1"
	resourceapi "k8s.io/api/resource/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/utils/test"
	"k8s.io/utils/set"
)

func TestSanitizedNodeResourceSlices(t *testing.T) {
	newNodeName := "newNode"
	nameSuffix := "abc"

	device1 := resourceapi.Device{Name: "device1"}
	device2 := resourceapi.Device{Name: "device2"}
	devices := []resourceapi.Device{device1, device2}

	allNodesSlice := &resourceapi.ResourceSlice{
		ObjectMeta: metav1.ObjectMeta{Name: "allNodesSlice", UID: "allNodesSlice"},
		Spec: resourceapi.ResourceSliceSpec{
			AllNodes: true,
			Driver:   "driver.example.com",
			Pool:     resourceapi.ResourcePool{Name: "all-nodes-pool1", Generation: 13, ResourceSliceCount: 37},
			Devices:  devices,
		},
	}
	multipleNodesSlice := &resourceapi.ResourceSlice{
		ObjectMeta: metav1.ObjectMeta{Name: "multipleNodesSlice", UID: "multipleNodesSlice"},
		Spec: resourceapi.ResourceSliceSpec{
			NodeSelector: &apiv1.NodeSelector{},
			Driver:       "driver.example.com",
			Pool:         resourceapi.ResourcePool{Name: "multiple-nodes-pool1", Generation: 13, ResourceSliceCount: 37},
			Devices:      devices,
		},
	}
	pool1Slice1 := &resourceapi.ResourceSlice{
		ObjectMeta: metav1.ObjectMeta{Name: "pool1Slice1", UID: "pool1Slice1"},
		Spec: resourceapi.ResourceSliceSpec{
			NodeName: "oldNode",
			Driver:   "driver.example.com",
			Pool:     resourceapi.ResourcePool{Name: "oldNode-pool1", Generation: 13, ResourceSliceCount: 37},
			Devices:  devices,
		},
	}
	pool1Slice2 := &resourceapi.ResourceSlice{
		ObjectMeta: metav1.ObjectMeta{Name: "pool1Slice2", UID: "pool1Slice2"},
		Spec: resourceapi.ResourceSliceSpec{
			NodeName: "oldNode",
			Driver:   "driver.example.com",
			Pool:     resourceapi.ResourcePool{Name: "oldNode-pool1", Generation: 13, ResourceSliceCount: 37},
			Devices:  devices,
		},
	}
	pool2Slice1 := &resourceapi.ResourceSlice{
		ObjectMeta: metav1.ObjectMeta{Name: "pool2Slice1", UID: "pool2Slice1"},
		Spec: resourceapi.ResourceSliceSpec{
			NodeName: "oldNode",
			Driver:   "driver.example.com",
			Pool:     resourceapi.ResourcePool{Name: "oldNode-pool2", Generation: 13, ResourceSliceCount: 37},
			Devices:  devices,
		},
	}
	pool2Slice2 := &resourceapi.ResourceSlice{
		ObjectMeta: metav1.ObjectMeta{Name: "pool2Slice2", UID: "pool2Slice2"},
		Spec: resourceapi.ResourceSliceSpec{
			NodeName: "oldNode",
			Driver:   "driver.example.com",
			Pool:     resourceapi.ResourcePool{Name: "oldNode-pool2", Generation: 13, ResourceSliceCount: 37},
			Devices:  devices,
		},
	}
	pool1Slice1Sanitized := &resourceapi.ResourceSlice{
		ObjectMeta: metav1.ObjectMeta{Name: "pool1Slice1-abc"},
		Spec: resourceapi.ResourceSliceSpec{
			NodeName: "newNode",
			Driver:   "driver.example.com",
			Pool:     resourceapi.ResourcePool{Name: "oldNode-pool1-abc", Generation: 13, ResourceSliceCount: 37},
			Devices:  devices,
		},
	}
	pool1Slice2Sanitized := &resourceapi.ResourceSlice{
		ObjectMeta: metav1.ObjectMeta{Name: "pool1Slice2-abc"},
		Spec: resourceapi.ResourceSliceSpec{
			NodeName: "newNode",
			Driver:   "driver.example.com",
			Pool:     resourceapi.ResourcePool{Name: "oldNode-pool1-abc", Generation: 13, ResourceSliceCount: 37},
			Devices:  devices,
		},
	}
	pool2Slice1Sanitized := &resourceapi.ResourceSlice{
		ObjectMeta: metav1.ObjectMeta{Name: "pool2Slice1-abc"},
		Spec: resourceapi.ResourceSliceSpec{
			NodeName: "newNode",
			Driver:   "driver.example.com",
			Pool:     resourceapi.ResourcePool{Name: "oldNode-pool2-abc", Generation: 13, ResourceSliceCount: 37},
			Devices:  devices,
		},
	}
	pool2Slice2Sanitized := &resourceapi.ResourceSlice{
		ObjectMeta: metav1.ObjectMeta{Name: "pool2Slice2-abc"},
		Spec: resourceapi.ResourceSliceSpec{
			NodeName: "newNode",
			Driver:   "driver.example.com",
			Pool:     resourceapi.ResourcePool{Name: "oldNode-pool2-abc", Generation: 13, ResourceSliceCount: 37},
			Devices:  devices,
		},
	}

	for _, tc := range []struct {
		testName      string
		slices        []*resourceapi.ResourceSlice
		wantSlices    []*resourceapi.ResourceSlice
		wantPoolNames set.Set[string]
		wantErr       error
	}{
		{
			testName:      "single node-local slice",
			slices:        []*resourceapi.ResourceSlice{pool1Slice1},
			wantSlices:    []*resourceapi.ResourceSlice{pool1Slice1Sanitized},
			wantPoolNames: set.New("oldNode-pool1"),
		},
		{
			testName:      "multiple node-local slices from single pool",
			slices:        []*resourceapi.ResourceSlice{pool1Slice1, pool1Slice2},
			wantSlices:    []*resourceapi.ResourceSlice{pool1Slice1Sanitized, pool1Slice2Sanitized},
			wantPoolNames: set.New("oldNode-pool1"),
		},
		{
			testName:      "multiple node-local slices from multiple pools",
			slices:        []*resourceapi.ResourceSlice{pool1Slice1, pool1Slice2, pool2Slice1, pool2Slice2},
			wantSlices:    []*resourceapi.ResourceSlice{pool1Slice1Sanitized, pool1Slice2Sanitized, pool2Slice1Sanitized, pool2Slice2Sanitized},
			wantPoolNames: set.New("oldNode-pool1", "oldNode-pool2"),
		},
		{
			testName: "global slices are an error",
			slices:   []*resourceapi.ResourceSlice{pool1Slice1, pool1Slice2, pool2Slice1, pool2Slice2, allNodesSlice},
			wantErr:  cmpopts.AnyError,
		},
		{
			testName: "multi-node slices are an error",
			slices:   []*resourceapi.ResourceSlice{pool1Slice1, pool1Slice2, pool2Slice1, pool2Slice2, multipleNodesSlice},
			wantErr:  cmpopts.AnyError,
		},
	} {
		t.Run(tc.testName, func(t *testing.T) {
			slices, poolNames, err := SanitizedNodeResourceSlices(tc.slices, newNodeName, nameSuffix)
			if diff := cmp.Diff(tc.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("SanitizeNodeResourceSlices(): unexpected error (-want +got): %s", diff)
			}
			for i, slice := range slices {
				origSlice := tc.slices[i]
				if slice.UID == "" || slice.UID == origSlice.UID {
					t.Errorf("SanitizeNodeResourceSlices(): slice %q UID wasn't properly sanitized: old UID %q, sanitized UID %q", origSlice.Name, origSlice.UID, slice.UID)
				}
			}
			if diff := cmp.Diff(tc.wantSlices, slices, cmpopts.IgnoreFields(metav1.ObjectMeta{}, "UID")); diff != "" {
				t.Errorf("SanitizeNodeResourceSlices(): unexpected slices (-want +got): %s", diff)
			}
			if diff := cmp.Diff(tc.wantPoolNames, poolNames); diff != "" {
				t.Errorf("SanitizeNodeResourceSlices(): unexpected pool names (-want +got): %s", diff)
			}
		})
	}
}

func TestSanitizedResourceClaimRefs(t *testing.T) {
	nameSuffix := "abc"

	noClaimRefsPod := test.BuildTestPod("noClaimRefsPod", 1, 1)
	sharedClaimsOnlyPod := test.BuildTestPod("sharedClaimsOnlyPod", 1, 1,
		test.WithResourceClaim("claim1", "sharedClaim1", ""),
		test.WithResourceClaim("claim2", "sharedClaim2", ""))
	ownAndSharedClaimsPod := test.BuildTestPod("ownAndSharedClaimsPod", 1, 1,
		test.WithResourceClaim("claim1", "sharedClaim1", ""),
		test.WithResourceClaim("claim2", "sharedClaim2", ""),
		test.WithResourceClaim("claim3", "own-claim-1-template-xxx", "own-claim-1-template"),
		test.WithResourceClaim("claim4", "own-claim-2-template-xxx", "own-claim-2-template"))

	for _, tc := range []struct {
		testName string
		pod      *apiv1.Pod
		wantPod  *apiv1.Pod
	}{
		{
			testName: "pod not referencing any ResourceClaims isn't changed",
			pod:      noClaimRefsPod,
			wantPod:  noClaimRefsPod,
		},
		{
			testName: "pod referencing only shared claims isn't changed",
			pod:      sharedClaimsOnlyPod,
			wantPod:  sharedClaimsOnlyPod,
		},
		{
			testName: "references to pod-owned claims get the suffix appended",
			pod:      ownAndSharedClaimsPod,
			wantPod: test.BuildTestPod("ownAndSharedClaimsPod", 1, 1,
				test.WithResourceClaim("claim1", "sharedClaim1", ""),
				test.WithResourceClaim("claim2", "sharedClaim2", ""),
				test.WithResourceClaim("claim3", "own-claim-1-template-xxx-abc", "own-claim-1-template"),
				test.WithResourceClaim("claim4", "own-claim-2-template-xxx-abc", "own-claim-2-template")),
		},
	} {
		t.Run(tc.testName, func(t *testing.T) {
			pod := SanitizedResourceClaimRefs(tc.pod, nameSuffix)
			if diff := cmp.Diff(tc.wantPod, pod); diff != "" {
				t.Errorf("SanitizeResourceClaimRefs(): unexpected output (-want +got): %s", diff)
			}
		})
	}
}

func TestSanitizedPodResourceClaims(t *testing.T) {
	nameSuffix := "abc"
	pod1 := &apiv1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod1", UID: "pod1Uid", Namespace: "default"}}
	pod2 := &apiv1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod2", UID: "pod2Uid", Namespace: "default"}}
	pod3 := &apiv1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod3", UID: "pod3Uid", Namespace: "default"}}
	oldOwner := &apiv1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "oldOwner", UID: "oldOwnerUid", Namespace: "default"}}
	newOwner := &apiv1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "newOwner", UID: "newOwnerUid", Namespace: "default"}}
	oldOwnerClaim1 := TestClaimWithPodOwnership(oldOwner,
		&resourceapi.ResourceClaim{ObjectMeta: metav1.ObjectMeta{Name: "oldOwnerClaim1", UID: "oldOwnerClaim1Uid", Namespace: "default"}},
	)
	oldOwnerClaim2 := TestClaimWithPodOwnership(oldOwner,
		&resourceapi.ResourceClaim{ObjectMeta: metav1.ObjectMeta{Name: "oldOwnerClaim2", UID: "oldOwnerClaim2Uid", Namespace: "default"}},
	)
	newOwnerClaim1 := TestClaimWithPodOwnership(newOwner,
		&resourceapi.ResourceClaim{ObjectMeta: metav1.ObjectMeta{Name: "oldOwnerClaim1-abc" /* UID should be randomized*/, Namespace: "default"}},
	)
	newOwnerClaim2 := TestClaimWithPodOwnership(newOwner,
		&resourceapi.ResourceClaim{ObjectMeta: metav1.ObjectMeta{Name: "oldOwnerClaim2-abc" /* UID should be randomized*/, Namespace: "default"}},
	)
	sharedClaim1 := &resourceapi.ResourceClaim{
		ObjectMeta: metav1.ObjectMeta{Name: "sharedClaim1", UID: "sharedClaim1Uid", Namespace: "default", OwnerReferences: []metav1.OwnerReference{{Kind: "NotAPod", Name: "someName"}}},
	}
	sharedClaim2 := &resourceapi.ResourceClaim{
		ObjectMeta: metav1.ObjectMeta{Name: "sharedClaim2", UID: "sharedClaim2Uid", Namespace: "default", OwnerReferences: []metav1.OwnerReference{{Kind: "NotAPod", Name: "someName"}}},
	}
	multipleNodesAllocation := &resourceapi.AllocationResult{
		NodeSelector: &apiv1.NodeSelector{NodeSelectorTerms: []apiv1.NodeSelectorTerm{{
			MatchExpressions: []apiv1.NodeSelectorRequirement{
				{Key: "someLabel", Operator: apiv1.NodeSelectorOpIn, Values: []string{"val1", "val2"}},
			}},
		}},
		Devices: resourceapi.DeviceAllocationResult{
			Results: []resourceapi.DeviceRequestAllocationResult{
				{Request: "request1", Driver: "driver.foo.com", Pool: "sharedPool1", Device: "device1"},
			},
		},
	}
	allNodesAllocation := &resourceapi.AllocationResult{
		NodeSelector: nil,
		Devices: resourceapi.DeviceAllocationResult{
			Results: []resourceapi.DeviceRequestAllocationResult{
				{Request: "request1", Driver: "driver.foo.com", Pool: "globalPool1", Device: "device1"},
			},
		},
	}
	singleNodeAllocationWithGlobalDevice := &resourceapi.AllocationResult{
		NodeSelector: &apiv1.NodeSelector{NodeSelectorTerms: []apiv1.NodeSelectorTerm{{
			MatchFields: []apiv1.NodeSelectorRequirement{
				{Key: "metadata.name", Operator: apiv1.NodeSelectorOpIn, Values: []string{"testNode"}},
			}},
		}},
		Devices: resourceapi.DeviceAllocationResult{
			Results: []resourceapi.DeviceRequestAllocationResult{
				{Request: "request1", Driver: "driver.foo.com", Pool: "testNodePool1", Device: "device1"},
				{Request: "request2", Driver: "driver.foo.com", Pool: "testNodePool1", Device: "device2"},
				{Request: "request3", Driver: "driver.foo.com", Pool: "testNodePool2", Device: "device1"},
				{Request: "request4", Driver: "driver.foo.com", Pool: "globalPool1", Device: "device1"},
			},
		},
	}
	singleNodeAllocation := &resourceapi.AllocationResult{
		NodeSelector: &apiv1.NodeSelector{NodeSelectorTerms: []apiv1.NodeSelectorTerm{{
			MatchFields: []apiv1.NodeSelectorRequirement{
				{Key: "metadata.name", Operator: apiv1.NodeSelectorOpIn, Values: []string{"testNode"}},
			}},
		}},
		Devices: resourceapi.DeviceAllocationResult{
			Results: []resourceapi.DeviceRequestAllocationResult{
				{Request: "request1", Driver: "driver.foo.com", Pool: "testNodePool1", Device: "device1"},
				{Request: "request2", Driver: "driver.foo.com", Pool: "testNodePool1", Device: "device2"},
				{Request: "request3", Driver: "driver.foo.com", Pool: "testNodePool2", Device: "device1"},
			},
		},
	}
	singleNodeAllocationSanitized := &resourceapi.AllocationResult{
		NodeSelector: &apiv1.NodeSelector{NodeSelectorTerms: []apiv1.NodeSelectorTerm{{
			MatchFields: []apiv1.NodeSelectorRequirement{
				{Key: "metadata.name", Operator: apiv1.NodeSelectorOpIn, Values: []string{"newNode"}},
			}},
		}},
		Devices: resourceapi.DeviceAllocationResult{
			Results: []resourceapi.DeviceRequestAllocationResult{
				{Request: "request1", Driver: "driver.foo.com", Pool: "testNodePool1-abc", Device: "device1"},
				{Request: "request2", Driver: "driver.foo.com", Pool: "testNodePool1-abc", Device: "device2"},
				{Request: "request3", Driver: "driver.foo.com", Pool: "testNodePool2-abc", Device: "device1"},
			},
		},
	}

	for _, tc := range []struct {
		testName         string
		claims           []*resourceapi.ResourceClaim
		oldNodeName      string
		newNodeName      string
		oldNodePoolNames set.Set[string]
		wantClaims       []*resourceapi.ResourceClaim
		wantErr          error
	}{
		{
			testName:   "no claims to sanitize",
			claims:     []*resourceapi.ResourceClaim{},
			wantClaims: []*resourceapi.ResourceClaim{},
		},
		{
			testName:   "unallocated shared claims are returned unchanged",
			claims:     []*resourceapi.ResourceClaim{sharedClaim1, sharedClaim2},
			wantClaims: []*resourceapi.ResourceClaim{sharedClaim1, sharedClaim2},
		},
		{
			testName: "allocated and reserved shared claims are returned unchanged",
			claims: []*resourceapi.ResourceClaim{
				TestClaimWithPodReservations(TestClaimWithAllocation(sharedClaim1, nil), pod1, pod2, pod3, oldOwner),
				TestClaimWithPodReservations(TestClaimWithAllocation(sharedClaim2, nil), pod1, pod2, pod3, oldOwner),
			},
			wantClaims: []*resourceapi.ResourceClaim{
				TestClaimWithPodReservations(TestClaimWithAllocation(sharedClaim1, nil), pod1, pod2, pod3, oldOwner),
				TestClaimWithPodReservations(TestClaimWithAllocation(sharedClaim2, nil), pod1, pod2, pod3, oldOwner),
			},
		},
		{
			testName:   "unallocated pod-owned claims are sanitized",
			claims:     []*resourceapi.ResourceClaim{oldOwnerClaim1, oldOwnerClaim2},
			wantClaims: []*resourceapi.ResourceClaim{newOwnerClaim1, newOwnerClaim2},
		},
		{
			testName:   "unallocated pod-owned claims have reservations cleared if needed", // This shouldn't normally happen, just a sanity check.
			claims:     []*resourceapi.ResourceClaim{TestClaimWithPodReservations(oldOwnerClaim1, oldOwner), TestClaimWithPodReservations(oldOwnerClaim2, oldOwner)},
			wantClaims: []*resourceapi.ResourceClaim{newOwnerClaim1, newOwnerClaim2},
		},
		{
			testName: "pod-owned claims available on multiple Nodes can't be sanitized",
			claims:   []*resourceapi.ResourceClaim{oldOwnerClaim1, TestClaimWithAllocation(oldOwnerClaim2, multipleNodesAllocation)},
			wantErr:  cmpopts.AnyError,
		},
		{
			testName: "pod-owned claims available on all Nodes can't be sanitized",
			claims:   []*resourceapi.ResourceClaim{oldOwnerClaim1, TestClaimWithAllocation(oldOwnerClaim2, allNodesAllocation)},
			wantErr:  cmpopts.AnyError,
		},
		{
			testName:    "pod-owned claims available on a different Node than the one being sanitized can't be sanitized", // This shouldn't normally happen, just a sanity check.
			claims:      []*resourceapi.ResourceClaim{oldOwnerClaim1, TestClaimWithAllocation(oldOwnerClaim2, singleNodeAllocation)},
			oldNodeName: "differentNode",
			newNodeName: "newNode",
			wantErr:     cmpopts.AnyError,
		},
		{
			testName:         "pod-owned claims available on a single Node but with non-local Devices allocated in the result", // This can happen if both a Node-local and a global device is allocated in the same claim.
			claims:           []*resourceapi.ResourceClaim{oldOwnerClaim1, TestClaimWithAllocation(oldOwnerClaim2, singleNodeAllocationWithGlobalDevice)},
			oldNodeName:      "testNode",
			newNodeName:      "newNode",
			oldNodePoolNames: set.New("testNodePool1", "testNodePool2"),
			wantErr:          cmpopts.AnyError,
		},
		{
			testName: "allocated and reserved pod-owned claims are sanitized",
			claims: []*resourceapi.ResourceClaim{
				TestClaimWithPodReservations(TestClaimWithAllocation(oldOwnerClaim1, singleNodeAllocation), oldOwner),
				TestClaimWithPodReservations(TestClaimWithAllocation(oldOwnerClaim2, singleNodeAllocation), oldOwner),
			},
			oldNodeName:      "testNode",
			newNodeName:      "newNode",
			oldNodePoolNames: set.New("testNodePool1", "testNodePool2"),
			wantClaims: []*resourceapi.ResourceClaim{
				TestClaimWithPodReservations(TestClaimWithAllocation(newOwnerClaim1, singleNodeAllocationSanitized), newOwner),
				TestClaimWithPodReservations(TestClaimWithAllocation(newOwnerClaim2, singleNodeAllocationSanitized), newOwner),
			},
		},
		{
			testName: "allocated pod-owned claims are reserved for the pod during sanitization if needed", // This shouldn't normally happen, just a sanity check.
			claims: []*resourceapi.ResourceClaim{
				TestClaimWithAllocation(oldOwnerClaim1, singleNodeAllocation),
				TestClaimWithAllocation(oldOwnerClaim2, singleNodeAllocation),
			},
			oldNodeName:      "testNode",
			newNodeName:      "newNode",
			oldNodePoolNames: set.New("testNodePool1", "testNodePool2"),
			wantClaims: []*resourceapi.ResourceClaim{
				TestClaimWithPodReservations(TestClaimWithAllocation(newOwnerClaim1, singleNodeAllocationSanitized), newOwner),
				TestClaimWithPodReservations(TestClaimWithAllocation(newOwnerClaim2, singleNodeAllocationSanitized), newOwner),
			},
		},
	} {
		t.Run(tc.testName, func(t *testing.T) {
			claims, err := SanitizedPodResourceClaims(newOwner, oldOwner, tc.claims, nameSuffix, tc.newNodeName, tc.oldNodeName, tc.oldNodePoolNames)
			if diff := cmp.Diff(tc.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("SanitizePodResourceClaims(): unexpected error (-want +got): %s", diff)
			}
			if tc.wantErr != nil {
				return
			}

			// Verify ResourceClaim UIDs.
			if len(tc.claims) != len(claims) {
				t.Fatalf("SanitizePodResourceClaims(): unexpected output length: want %v, got %v", len(tc.claims), len(claims))
			}
			for i, newClaim := range claims {
				oldClaim := tc.claims[i]
				if oldClaim.Name == "sharedClaim1" || oldClaim.Name == "sharedClaim2" {
					// For shared claims, verify that the UID is not changed.
					if gotUid, oldUid := newClaim.UID, oldClaim.UID; gotUid != oldUid {
						t.Errorf("shared ResourceClaim UID changed - got %q, old UID was %q", gotUid, oldUid)
					}
				} else {
					// For pod-owned claims, verify that the UID is randomized.
					if gotUid, oldUid := newClaim.UID, oldClaim.UID; gotUid == "" || gotUid == oldUid {
						t.Errorf("sanitized ResourceClaim UID wasn't randomized - got %q, old UID was %q", gotUid, oldUid)
					}
				}
			}
			// Verify the rest of ResourceClaims.
			if diff := cmp.Diff(tc.wantClaims, claims, cmpopts.EquateEmpty(), cmpopts.IgnoreFields(metav1.ObjectMeta{}, "UID")); diff != "" {
				t.Errorf("SanitizePodResourceClaims(): unexpected output (-want +got): %s", diff)
			}
		})
	}
}
