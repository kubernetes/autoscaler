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
	"k8s.io/apimachinery/pkg/types"
)

func TestClaimOwningPod(t *testing.T) {
	truePtr := true
	for _, tc := range []struct {
		testName string
		claim    *resourceapi.ResourceClaim
		wantName string
		wantUid  types.UID
	}{
		{
			testName: "claim with no owners",
			claim: &resourceapi.ResourceClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name: "claim", UID: "claim", Namespace: "default",
				},
			},
			wantName: "",
			wantUid:  "",
		},
		{
			testName: "claim with non-Pod owners",
			claim: &resourceapi.ResourceClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name: "claim", UID: "claim", Namespace: "default",
					OwnerReferences: []metav1.OwnerReference{
						{Name: "owner1", UID: "owner1uid", APIVersion: "v1", Kind: "ReplicationController", Controller: &truePtr},
						{Name: "owner2", UID: "owner2uid", APIVersion: "v1", Kind: "ConfigMap"},
					},
				},
			},
			wantName: "",
			wantUid:  "",
		},
		{
			testName: "claim with a Pod non-controller owner",
			claim: &resourceapi.ResourceClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name: "claim", UID: "claim", Namespace: "default",
					OwnerReferences: []metav1.OwnerReference{
						{Name: "owner1", UID: "owner1uid", APIVersion: "v1", Kind: "ReplicationController"},
						{Name: "owner2", UID: "owner2uid", APIVersion: "v1", Kind: "ConfigMap"},
						{Name: "owner3", UID: "owner3uid", APIVersion: "v1", Kind: "Pod"},
					},
				},
			},
			wantName: "",
			wantUid:  "",
		},
		{
			testName: "claim with a Pod controller owner",
			claim: &resourceapi.ResourceClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name: "claim", UID: "claim", Namespace: "default",
					OwnerReferences: []metav1.OwnerReference{
						{Name: "owner1", UID: "owner1uid", APIVersion: "v1", Kind: "ReplicationController"},
						{Name: "owner2", UID: "owner2uid", APIVersion: "v1", Kind: "ConfigMap"},
						{Name: "owner3", UID: "owner3uid", APIVersion: "v1", Kind: "Pod", Controller: &truePtr},
					},
				},
			},
			wantName: "owner3",
			wantUid:  "owner3uid",
		},
	} {
		t.Run(tc.testName, func(t *testing.T) {
			name, uid := ClaimOwningPod(tc.claim)
			if tc.wantName != name {
				t.Errorf("ClaimOwningPod(): unexpected output name: want %s, got %s", tc.wantName, name)
			}
			if tc.wantUid != uid {
				t.Errorf("ClaimOwningPod(): unexpected output UID: want %v, got %v", tc.wantUid, uid)
			}
		})
	}
}

func TestClaimAllocated(t *testing.T) {
	for _, tc := range []struct {
		testName      string
		claim         *resourceapi.ResourceClaim
		wantAllocated bool
	}{
		{
			testName: "claim with empty status",
			claim: &resourceapi.ResourceClaim{
				ObjectMeta: metav1.ObjectMeta{Name: "claim", UID: "claimUid", Namespace: "default"},
				Status:     resourceapi.ResourceClaimStatus{},
			},
			wantAllocated: false,
		},
		{
			testName: "claim with some devices reported in the status, but nil allocation",
			claim: &resourceapi.ResourceClaim{
				ObjectMeta: metav1.ObjectMeta{Name: "claim", UID: "claimUid", Namespace: "default"},
				Status: resourceapi.ResourceClaimStatus{
					Allocation: nil,
					Devices: []resourceapi.AllocatedDeviceStatus{
						{Driver: "driver.foo.com", Pool: "pool", Device: "dev"},
					},
				},
			},
			wantAllocated: false,
		},
		{
			testName: "claim with consumer reservations in the status, but nil allocation",
			claim: &resourceapi.ResourceClaim{
				ObjectMeta: metav1.ObjectMeta{Name: "claim", UID: "claimUid", Namespace: "default"},
				Status: resourceapi.ResourceClaimStatus{
					Allocation: nil,
					ReservedFor: []resourceapi.ResourceClaimConsumerReference{
						{Resource: "pods", Name: "pod", UID: "podUid"},
					},
				},
			},
			wantAllocated: false,
		},
		{
			testName: "claim with non-nil (but empty for some reason) allocation",
			claim: &resourceapi.ResourceClaim{
				ObjectMeta: metav1.ObjectMeta{Name: "claim", UID: "claimUid", Namespace: "default"},
				Status: resourceapi.ResourceClaimStatus{
					Allocation: &resourceapi.AllocationResult{},
				},
			},
			wantAllocated: true,
		},
		{
			testName: "claim with non-nil allocation",
			claim: &resourceapi.ResourceClaim{
				ObjectMeta: metav1.ObjectMeta{Name: "claim", UID: "claimUid", Namespace: "default"},
				Status: resourceapi.ResourceClaimStatus{
					Allocation: &resourceapi.AllocationResult{
						Devices: resourceapi.DeviceAllocationResult{
							Results: []resourceapi.DeviceRequestAllocationResult{
								{Driver: "driver.foo.com", Pool: "pool", Device: "dev"},
							},
						},
					},
				},
			},
			wantAllocated: true,
		},
	} {
		t.Run(tc.testName, func(t *testing.T) {
			allocated := ClaimAllocated(tc.claim)
			if tc.wantAllocated != allocated {
				t.Errorf("ClaimAllocated(): unexpected result: want %v, got %v", tc.wantAllocated, allocated)
			}
		})
	}
}

func TestClaimInUse(t *testing.T) {
	for _, tc := range []struct {
		testName  string
		claim     *resourceapi.ResourceClaim
		wantInUse bool
	}{
		{
			testName: "claim with no reservations",
			claim: &resourceapi.ResourceClaim{
				ObjectMeta: metav1.ObjectMeta{Name: "claim", UID: "claimUid", Namespace: "default"},
				Status:     resourceapi.ResourceClaimStatus{},
			},
			wantInUse: false,
		},
		{
			testName: "claim with a single pod reservation",
			claim: &resourceapi.ResourceClaim{
				ObjectMeta: metav1.ObjectMeta{Name: "claim", UID: "claimUid", Namespace: "default"},
				Status: resourceapi.ResourceClaimStatus{
					ReservedFor: []resourceapi.ResourceClaimConsumerReference{
						{Resource: "pods", Name: "pod", UID: "podUid"},
					},
				},
			},
			wantInUse: true,
		},
		{
			testName: "claim with a single non-pod reservation",
			claim: &resourceapi.ResourceClaim{
				ObjectMeta: metav1.ObjectMeta{Name: "claim", UID: "claimUid", Namespace: "default"},
				Status: resourceapi.ResourceClaimStatus{
					ReservedFor: []resourceapi.ResourceClaimConsumerReference{
						{Resource: "somethingelses", Name: "somethingelse", UID: "somethingelseUid"},
					},
				},
			},
			wantInUse: true,
		},
		{
			testName: "claim with multiple reservations",
			claim: &resourceapi.ResourceClaim{
				ObjectMeta: metav1.ObjectMeta{Name: "claim", UID: "claimUid", Namespace: "default"},
				Status: resourceapi.ResourceClaimStatus{
					ReservedFor: []resourceapi.ResourceClaimConsumerReference{
						{Resource: "pods", Name: "pod1", UID: "pod1Uid"},
						{Resource: "pods", Name: "pod2", UID: "pod2Uid"},
						{Resource: "somethingelses", Name: "somethingelse", UID: "somethingelseUid"},
					},
				},
			},
			wantInUse: true,
		},
	} {
		t.Run(tc.testName, func(t *testing.T) {
			inUse := ClaimInUse(tc.claim)
			if tc.wantInUse != inUse {
				t.Errorf("ClaimInUse(): unexpected result: want %v, got %v", tc.wantInUse, inUse)
			}
		})
	}
}

func TestClaimReservedForPod(t *testing.T) {
	pod := &apiv1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "chosenPod", UID: "chosenPodUid"}}

	for _, tc := range []struct {
		testName     string
		claim        *resourceapi.ResourceClaim
		wantReserved bool
	}{
		{
			testName: "claim with no reservations",
			claim: &resourceapi.ResourceClaim{
				ObjectMeta: metav1.ObjectMeta{Name: "claim", UID: "claimUid", Namespace: "default"},
				Status:     resourceapi.ResourceClaimStatus{},
			},
			wantReserved: false,
		},
		{
			testName: "claim with some reservations, but none match the pod",
			claim: &resourceapi.ResourceClaim{
				ObjectMeta: metav1.ObjectMeta{Name: "claim", UID: "claimUid", Namespace: "default"},
				Status: resourceapi.ResourceClaimStatus{
					ReservedFor: []resourceapi.ResourceClaimConsumerReference{
						{Resource: "pods", Name: "pod1", UID: "pod1Uid"},
						{Resource: "pods", Name: "pod2", UID: "pod2Uid"},
						{Resource: "somethingelses", Name: "somethingelse", UID: "somethingelseUid"},
						{Resource: "pods", Name: "chosenPod", UID: "badUid"},
						{Resource: "pods", Name: "badName", UID: "chosenPodUid"},
						{Resource: "badResource", Name: "chosenPod", UID: "chosenPodUid"},
					},
				},
			},
			wantReserved: false,
		},
		{
			testName: "claim with some reservations, one matches the pod",
			claim: &resourceapi.ResourceClaim{
				ObjectMeta: metav1.ObjectMeta{Name: "claim", UID: "claimUid", Namespace: "default"},
				Status: resourceapi.ResourceClaimStatus{
					ReservedFor: []resourceapi.ResourceClaimConsumerReference{
						{Resource: "pods", Name: "pod1", UID: "pod1Uid"},
						{Resource: "pods", Name: "pod2", UID: "pod2Uid"},
						{Resource: "somethingelses", Name: "somethingelse", UID: "somethingelseUid"},
						{Resource: "pods", Name: "chosenPod", UID: "badUid"},
						{Resource: "pods", Name: "badName", UID: "chosenPodUid"},
						{Resource: "badResource", Name: "chosenPod", UID: "chosenPodUid"},
						{Resource: "pods", Name: "chosenPod", UID: "chosenPodUid"},
					},
				},
			},
			wantReserved: true,
		},
	} {
		t.Run(tc.testName, func(t *testing.T) {
			reserved := ClaimReservedForPod(tc.claim, pod)
			if tc.wantReserved != reserved {
				t.Errorf("ClaimReservedForPod(): unexpected result: want %v, got %v", tc.wantReserved, reserved)
			}
		})
	}
}

func TestClaimFullyReserved(t *testing.T) {
	for _, tc := range []struct {
		testName          string
		claim             *resourceapi.ResourceClaim
		wantFullyReserved bool
	}{
		{
			testName: "claim with no reservations",
			claim: &resourceapi.ResourceClaim{
				ObjectMeta: metav1.ObjectMeta{Name: "claim", UID: "claimUid", Namespace: "default"},
				Status:     resourceapi.ResourceClaimStatus{},
			},
			wantFullyReserved: false,
		},
		{
			testName: "claim with fewer than max reservations",
			claim: &resourceapi.ResourceClaim{
				ObjectMeta: metav1.ObjectMeta{Name: "claim", UID: "claimUid", Namespace: "default"},
				Status: resourceapi.ResourceClaimStatus{
					ReservedFor: testClaimReservations(resourceapi.ResourceClaimReservedForMaxSize - 1),
				},
			},
			wantFullyReserved: false,
		},
		{
			testName: "claim with exactly max reservations",
			claim: &resourceapi.ResourceClaim{
				ObjectMeta: metav1.ObjectMeta{Name: "claim", UID: "claimUid", Namespace: "default"},
				Status: resourceapi.ResourceClaimStatus{
					ReservedFor: testClaimReservations(resourceapi.ResourceClaimReservedForMaxSize),
				},
			},
			wantFullyReserved: true,
		},
		{
			testName: "claim with more than max reservations",
			claim: &resourceapi.ResourceClaim{
				ObjectMeta: metav1.ObjectMeta{Name: "claim", UID: "claimUid", Namespace: "default"},
				Status: resourceapi.ResourceClaimStatus{
					ReservedFor: testClaimReservations(resourceapi.ResourceClaimReservedForMaxSize + 1),
				},
			},
			wantFullyReserved: true,
		},
	} {
		t.Run(tc.testName, func(t *testing.T) {
			fullyReserved := ClaimFullyReserved(tc.claim)
			if tc.wantFullyReserved != fullyReserved {
				t.Errorf("ClaimFullyReserved(): unexpected result: want %v, got %v", tc.wantFullyReserved, fullyReserved)
			}
		})
	}
}

func TestClaimAvailableOnNode(t *testing.T) {
	for _, tc := range []struct {
		testName      string
		claim         *resourceapi.ResourceClaim
		node          *apiv1.Node
		wantAvailable bool
		wantErr       error
	}{
		{
			testName: "unallocated claim",
			claim: &resourceapi.ResourceClaim{
				ObjectMeta: metav1.ObjectMeta{Name: "claim", UID: "claimUid", Namespace: "default"},
				Status:     resourceapi.ResourceClaimStatus{},
			},
			node:          &apiv1.Node{ObjectMeta: metav1.ObjectMeta{Name: "node"}},
			wantAvailable: false,
		},
		{
			testName: "allocated claim, allocation available on all nodes",
			claim: &resourceapi.ResourceClaim{
				ObjectMeta: metav1.ObjectMeta{Name: "claim", UID: "claimUid", Namespace: "default"},
				Status: resourceapi.ResourceClaimStatus{
					Allocation: &resourceapi.AllocationResult{
						NodeSelector: nil,
					},
				},
			},
			node: &apiv1.Node{ObjectMeta: metav1.ObjectMeta{
				Name: "node",
			}},
			wantAvailable: true,
		},
		{
			testName: "allocated claim, allocation available on a subset of nodes, node doesn't match",
			claim: &resourceapi.ResourceClaim{
				ObjectMeta: metav1.ObjectMeta{Name: "claim", UID: "claimUid", Namespace: "default"},
				Status: resourceapi.ResourceClaimStatus{
					Allocation: &resourceapi.AllocationResult{
						NodeSelector: &apiv1.NodeSelector{NodeSelectorTerms: []apiv1.NodeSelectorTerm{{
							MatchExpressions: []apiv1.NodeSelectorRequirement{
								{Key: "someLabel", Operator: apiv1.NodeSelectorOpIn, Values: []string{"val1", "val2"}},
							}},
						}},
					},
				},
			},
			node:          &apiv1.Node{ObjectMeta: metav1.ObjectMeta{Name: "node", Labels: map[string]string{"someLabel": "val3"}}},
			wantAvailable: false,
		},
		{
			testName: "allocated claim, allocation available on a subset of nodes, node matches",
			claim: &resourceapi.ResourceClaim{
				ObjectMeta: metav1.ObjectMeta{Name: "claim", UID: "claimUid", Namespace: "default"},
				Status: resourceapi.ResourceClaimStatus{
					Allocation: &resourceapi.AllocationResult{
						NodeSelector: &apiv1.NodeSelector{NodeSelectorTerms: []apiv1.NodeSelectorTerm{{
							MatchExpressions: []apiv1.NodeSelectorRequirement{
								{Key: "someLabel", Operator: apiv1.NodeSelectorOpIn, Values: []string{"val1", "val2"}},
							}},
						}},
					},
				},
			},
			node:          &apiv1.Node{ObjectMeta: metav1.ObjectMeta{Name: "node", Labels: map[string]string{"someLabel": "val2"}}},
			wantAvailable: true,
		},
		{
			testName: "allocated claim, allocation available on a single node, node doesn't match",
			claim: &resourceapi.ResourceClaim{
				ObjectMeta: metav1.ObjectMeta{Name: "claim", UID: "claimUid", Namespace: "default"},
				Status: resourceapi.ResourceClaimStatus{
					Allocation: &resourceapi.AllocationResult{
						NodeSelector: &apiv1.NodeSelector{NodeSelectorTerms: []apiv1.NodeSelectorTerm{{
							MatchFields: []apiv1.NodeSelectorRequirement{
								{Key: "metadata.name", Operator: apiv1.NodeSelectorOpIn, Values: []string{"otherNode"}},
							}},
						}},
					},
				},
			},
			node:          &apiv1.Node{ObjectMeta: metav1.ObjectMeta{Name: "node"}},
			wantAvailable: false,
		},
		{
			testName: "allocated claim, allocation available on a single node, node matches",
			claim: &resourceapi.ResourceClaim{
				ObjectMeta: metav1.ObjectMeta{Name: "claim", UID: "claimUid", Namespace: "default"},
				Status: resourceapi.ResourceClaimStatus{
					Allocation: &resourceapi.AllocationResult{
						NodeSelector: &apiv1.NodeSelector{NodeSelectorTerms: []apiv1.NodeSelectorTerm{{
							MatchFields: []apiv1.NodeSelectorRequirement{
								{Key: "metadata.name", Operator: apiv1.NodeSelectorOpIn, Values: []string{"node"}},
							}},
						}},
					},
				},
			},
			node:          &apiv1.Node{ObjectMeta: metav1.ObjectMeta{Name: "node"}},
			wantAvailable: true,
		},
		{
			testName: "allocated claim, invalid node selector",
			claim: &resourceapi.ResourceClaim{
				ObjectMeta: metav1.ObjectMeta{Name: "claim", UID: "claimUid", Namespace: "default"},
				Status: resourceapi.ResourceClaimStatus{
					Allocation: &resourceapi.AllocationResult{
						NodeSelector: &apiv1.NodeSelector{NodeSelectorTerms: []apiv1.NodeSelectorTerm{{
							MatchFields: []apiv1.NodeSelectorRequirement{
								{Key: "metadata.name", Operator: "bad-operator", Values: []string{"node"}},
							}},
						}},
					},
				},
			},
			node:          &apiv1.Node{ObjectMeta: metav1.ObjectMeta{Name: "node"}},
			wantAvailable: false,
			wantErr:       cmpopts.AnyError,
		},
	} {
		t.Run(tc.testName, func(t *testing.T) {
			available, err := ClaimAvailableOnNode(tc.claim, tc.node)
			if diff := cmp.Diff(tc.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("ClaimAvailableOnNode(): unexpected error (-want +got): %s", diff)
			}
			if tc.wantAvailable != available {
				t.Errorf("ClaimAvailableOnNode(): unexpected result: want %v, got %v", tc.wantAvailable, available)
			}
		})
	}
}

func TestDeallocateClaimInPlace(t *testing.T) {
	for _, tc := range []struct {
		testName  string
		claim     *resourceapi.ResourceClaim
		wantClaim *resourceapi.ResourceClaim
	}{
		{
			testName: "unallocated claim doesn't change",
			claim: &resourceapi.ResourceClaim{
				ObjectMeta: metav1.ObjectMeta{Name: "claim", UID: "claimUid", Namespace: "default"},
				Spec:       resourceapi.ResourceClaimSpec{Devices: resourceapi.DeviceClaim{Requests: []resourceapi.DeviceRequest{{Name: "request"}}}},
				Status:     resourceapi.ResourceClaimStatus{},
			},
			wantClaim: &resourceapi.ResourceClaim{
				ObjectMeta: metav1.ObjectMeta{Name: "claim", UID: "claimUid", Namespace: "default"},
				Spec:       resourceapi.ResourceClaimSpec{Devices: resourceapi.DeviceClaim{Requests: []resourceapi.DeviceRequest{{Name: "request"}}}},
				Status:     resourceapi.ResourceClaimStatus{},
			},
		},
		{
			testName: "allocated claim gets allocation cleared",
			claim: &resourceapi.ResourceClaim{
				ObjectMeta: metav1.ObjectMeta{Name: "claim", UID: "claimUid", Namespace: "default"},
				Spec:       resourceapi.ResourceClaimSpec{Devices: resourceapi.DeviceClaim{Requests: []resourceapi.DeviceRequest{{Name: "request"}}}},
				Status: resourceapi.ResourceClaimStatus{
					Allocation: &resourceapi.AllocationResult{
						Devices: resourceapi.DeviceAllocationResult{
							Results: []resourceapi.DeviceRequestAllocationResult{
								{Driver: "driver.foo.com", Pool: "pool", Device: "dev"},
							},
						},
					},
				},
			},
			wantClaim: &resourceapi.ResourceClaim{
				ObjectMeta: metav1.ObjectMeta{Name: "claim", UID: "claimUid", Namespace: "default"},
				Spec:       resourceapi.ResourceClaimSpec{Devices: resourceapi.DeviceClaim{Requests: []resourceapi.DeviceRequest{{Name: "request"}}}},
				Status: resourceapi.ResourceClaimStatus{
					Allocation: nil,
				},
			},
		},
	} {
		t.Run(tc.testName, func(t *testing.T) {
			claim := tc.claim.DeepCopy()
			DeallocateClaimInPlace(claim)
			if diff := cmp.Diff(tc.wantClaim, claim); diff != "" {
				t.Errorf("DeallocateClaimInPlace(): unexpected claim state after call (-want +got): %s", diff)
			}
		})
	}
}

func TestClearPodReservationInPlace(t *testing.T) {
	pod := &apiv1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "chosenPod", UID: "chosenPodUid"}}

	for _, tc := range []struct {
		testName  string
		claim     *resourceapi.ResourceClaim
		wantClaim *resourceapi.ResourceClaim
	}{
		{
			testName: "no reservations in claim - nothing to clear",
			claim: &resourceapi.ResourceClaim{
				ObjectMeta: metav1.ObjectMeta{Name: "claim", UID: "claimUid", Namespace: "default"},
				Spec:       resourceapi.ResourceClaimSpec{Devices: resourceapi.DeviceClaim{Requests: []resourceapi.DeviceRequest{{Name: "request"}}}},
				Status: resourceapi.ResourceClaimStatus{
					ReservedFor: nil,
				},
			},
			wantClaim: &resourceapi.ResourceClaim{
				ObjectMeta: metav1.ObjectMeta{Name: "claim", UID: "claimUid", Namespace: "default"},
				Spec:       resourceapi.ResourceClaimSpec{Devices: resourceapi.DeviceClaim{Requests: []resourceapi.DeviceRequest{{Name: "request"}}}},
				Status: resourceapi.ResourceClaimStatus{
					ReservedFor: nil,
				},
			},
		},

		{
			testName: "only reservation for the pod in the claim - removed",
			claim: &resourceapi.ResourceClaim{
				ObjectMeta: metav1.ObjectMeta{Name: "claim", UID: "claimUid", Namespace: "default"},
				Spec:       resourceapi.ResourceClaimSpec{Devices: resourceapi.DeviceClaim{Requests: []resourceapi.DeviceRequest{{Name: "request"}}}},
				Status: resourceapi.ResourceClaimStatus{
					ReservedFor: []resourceapi.ResourceClaimConsumerReference{
						{Resource: "pods", Name: "chosenPod", UID: "chosenPodUid"},
					},
				},
			},
			wantClaim: &resourceapi.ResourceClaim{
				ObjectMeta: metav1.ObjectMeta{Name: "claim", UID: "claimUid", Namespace: "default"},
				Spec:       resourceapi.ResourceClaimSpec{Devices: resourceapi.DeviceClaim{Requests: []resourceapi.DeviceRequest{{Name: "request"}}}},
				Status: resourceapi.ResourceClaimStatus{
					ReservedFor: []resourceapi.ResourceClaimConsumerReference{},
				},
			},
		},
		{
			testName: "some reservations in claim, one for the pod - removed",
			claim: &resourceapi.ResourceClaim{
				ObjectMeta: metav1.ObjectMeta{Name: "claim", UID: "claimUid", Namespace: "default"},
				Spec:       resourceapi.ResourceClaimSpec{Devices: resourceapi.DeviceClaim{Requests: []resourceapi.DeviceRequest{{Name: "request"}}}},
				Status: resourceapi.ResourceClaimStatus{
					ReservedFor: []resourceapi.ResourceClaimConsumerReference{
						{Resource: "pods", Name: "pod1", UID: "pod1Uid"},
						{Resource: "pods", Name: "pod2", UID: "pod2Uid"},
						{Resource: "pods", Name: "chosenPod", UID: "chosenPodUid"},
						{Resource: "somethingelses", Name: "somethingelse", UID: "somethingelseUid"},
						{Resource: "pods", Name: "chosenPod", UID: "badUid"},
						{Resource: "pods", Name: "badName", UID: "chosenPodUid"},
						{Resource: "badResource", Name: "chosenPod", UID: "chosenPodUid"},
					},
				},
			},
			wantClaim: &resourceapi.ResourceClaim{
				ObjectMeta: metav1.ObjectMeta{Name: "claim", UID: "claimUid", Namespace: "default"},
				Spec:       resourceapi.ResourceClaimSpec{Devices: resourceapi.DeviceClaim{Requests: []resourceapi.DeviceRequest{{Name: "request"}}}},
				Status: resourceapi.ResourceClaimStatus{
					ReservedFor: []resourceapi.ResourceClaimConsumerReference{
						{Resource: "pods", Name: "pod1", UID: "pod1Uid"},
						{Resource: "pods", Name: "pod2", UID: "pod2Uid"},
						{Resource: "somethingelses", Name: "somethingelse", UID: "somethingelseUid"},
						{Resource: "pods", Name: "chosenPod", UID: "badUid"},
						{Resource: "pods", Name: "badName", UID: "chosenPodUid"},
						{Resource: "badResource", Name: "chosenPod", UID: "chosenPodUid"},
					},
				},
			},
		},
		{
			testName: "some reservations in claim but not for the pod - nothing to clear",
			claim: &resourceapi.ResourceClaim{
				ObjectMeta: metav1.ObjectMeta{Name: "claim", UID: "claimUid", Namespace: "default"},
				Spec:       resourceapi.ResourceClaimSpec{Devices: resourceapi.DeviceClaim{Requests: []resourceapi.DeviceRequest{{Name: "request"}}}},
				Status: resourceapi.ResourceClaimStatus{
					ReservedFor: []resourceapi.ResourceClaimConsumerReference{
						{Resource: "pods", Name: "pod1", UID: "pod1Uid"},
						{Resource: "pods", Name: "pod2", UID: "pod2Uid"},
						{Resource: "somethingelses", Name: "somethingelse", UID: "somethingelseUid"},
						{Resource: "pods", Name: "chosenPod", UID: "badUid"},
						{Resource: "pods", Name: "badName", UID: "chosenPodUid"},
						{Resource: "badResource", Name: "chosenPod", UID: "chosenPodUid"},
					},
				},
			},
			wantClaim: &resourceapi.ResourceClaim{
				ObjectMeta: metav1.ObjectMeta{Name: "claim", UID: "claimUid", Namespace: "default"},
				Spec:       resourceapi.ResourceClaimSpec{Devices: resourceapi.DeviceClaim{Requests: []resourceapi.DeviceRequest{{Name: "request"}}}},
				Status: resourceapi.ResourceClaimStatus{
					ReservedFor: []resourceapi.ResourceClaimConsumerReference{
						{Resource: "pods", Name: "pod1", UID: "pod1Uid"},
						{Resource: "pods", Name: "pod2", UID: "pod2Uid"},
						{Resource: "somethingelses", Name: "somethingelse", UID: "somethingelseUid"},
						{Resource: "pods", Name: "chosenPod", UID: "badUid"},
						{Resource: "pods", Name: "badName", UID: "chosenPodUid"},
						{Resource: "badResource", Name: "chosenPod", UID: "chosenPodUid"},
					},
				},
			},
		},
	} {
		t.Run(tc.testName, func(t *testing.T) {
			claim := tc.claim.DeepCopy()
			ClearPodReservationInPlace(claim, pod)
			if diff := cmp.Diff(tc.wantClaim, claim); diff != "" {
				t.Errorf("ClearPodReservationInPlace(): unexpected claim state after call (-want +got): %s", diff)
			}
		})
	}
}

func TestAddPodReservationInPlace(t *testing.T) {
	pod := &apiv1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "chosenPod", UID: "chosenPodUid"}}

	for _, tc := range []struct {
		testName  string
		claim     *resourceapi.ResourceClaim
		wantClaim *resourceapi.ResourceClaim
	}{
		{
			testName: "reservation added to empty list",
			claim: &resourceapi.ResourceClaim{
				ObjectMeta: metav1.ObjectMeta{Name: "claim", UID: "claimUid", Namespace: "default"},
				Spec:       resourceapi.ResourceClaimSpec{Devices: resourceapi.DeviceClaim{Requests: []resourceapi.DeviceRequest{{Name: "request"}}}},
				Status: resourceapi.ResourceClaimStatus{
					ReservedFor: nil,
				},
			},
			wantClaim: &resourceapi.ResourceClaim{
				ObjectMeta: metav1.ObjectMeta{Name: "claim", UID: "claimUid", Namespace: "default"},
				Spec:       resourceapi.ResourceClaimSpec{Devices: resourceapi.DeviceClaim{Requests: []resourceapi.DeviceRequest{{Name: "request"}}}},
				Status: resourceapi.ResourceClaimStatus{
					ReservedFor: []resourceapi.ResourceClaimConsumerReference{
						{Resource: "pods", Name: "chosenPod", UID: "chosenPodUid"},
					},
				},
			},
		},
		{
			testName: "reservation added to other reservations",
			claim: &resourceapi.ResourceClaim{
				ObjectMeta: metav1.ObjectMeta{Name: "claim", UID: "claimUid", Namespace: "default"},
				Spec:       resourceapi.ResourceClaimSpec{Devices: resourceapi.DeviceClaim{Requests: []resourceapi.DeviceRequest{{Name: "request"}}}},
				Status: resourceapi.ResourceClaimStatus{
					ReservedFor: []resourceapi.ResourceClaimConsumerReference{
						{Resource: "pods", Name: "pod1", UID: "pod1Uid"},
						{Resource: "pods", Name: "pod2", UID: "pod2Uid"},
						{Resource: "somethingelses", Name: "somethingelse", UID: "somethingelseUid"},
					},
				},
			},
			wantClaim: &resourceapi.ResourceClaim{
				ObjectMeta: metav1.ObjectMeta{Name: "claim", UID: "claimUid", Namespace: "default"},
				Spec:       resourceapi.ResourceClaimSpec{Devices: resourceapi.DeviceClaim{Requests: []resourceapi.DeviceRequest{{Name: "request"}}}},
				Status: resourceapi.ResourceClaimStatus{
					ReservedFor: []resourceapi.ResourceClaimConsumerReference{
						{Resource: "pods", Name: "pod1", UID: "pod1Uid"},
						{Resource: "pods", Name: "pod2", UID: "pod2Uid"},
						{Resource: "somethingelses", Name: "somethingelse", UID: "somethingelseUid"},
						{Resource: "pods", Name: "chosenPod", UID: "chosenPodUid"},
					},
				},
			},
		},
	} {
		t.Run(tc.testName, func(t *testing.T) {
			claim := tc.claim.DeepCopy()
			AddPodReservationInPlace(claim, pod)
			if diff := cmp.Diff(tc.wantClaim, claim); diff != "" {
				t.Errorf("AddPodReservationInPlace(): unexpected claim state after call (-want +got): %s", diff)
			}
		})
	}
}
