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
	"fmt"

	apiv1 "k8s.io/api/core/v1"
	resourceapi "k8s.io/api/resource/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// TestClaimWithPodOwnership returns a copy of the provided claim with OwnerReferences set up so that
// the claim is owned by the provided pod.
func TestClaimWithPodOwnership(pod *apiv1.Pod, claim *resourceapi.ResourceClaim) *resourceapi.ResourceClaim {
	result := claim.DeepCopy()
	result.OwnerReferences = []metav1.OwnerReference{PodClaimOwnerReference(pod)}
	return result
}

// TestClaimWithPodReservations returns a copy of the provided claim with reservations for the provided pods added
// to ReservedFor.
func TestClaimWithPodReservations(claim *resourceapi.ResourceClaim, pods ...*apiv1.Pod) *resourceapi.ResourceClaim {
	result := claim.DeepCopy()
	for _, pod := range pods {
		result.Status.ReservedFor = append(result.Status.ReservedFor, PodClaimConsumerReference(pod))
	}
	return result
}

// TestClaimWithAllocation returns a copy of the provided claim with an allocation set.
func TestClaimWithAllocation(claim *resourceapi.ResourceClaim, allocation *resourceapi.AllocationResult) *resourceapi.ResourceClaim {
	result := claim.DeepCopy()
	defaultAlloc := &resourceapi.AllocationResult{
		Devices: resourceapi.DeviceAllocationResult{
			Results: []resourceapi.DeviceRequestAllocationResult{
				{Request: "req1", Driver: "driver.example.com", Pool: "pool1", Device: "device1"},
			},
		},
	}
	if allocation == nil {
		allocation = defaultAlloc
	}
	result.Status.Allocation = allocation
	return result
}

func testClaimReservations(count int) []resourceapi.ResourceClaimConsumerReference {
	var result []resourceapi.ResourceClaimConsumerReference
	for i := range count {
		podName := fmt.Sprintf("pod-%d", i)
		result = append(result, resourceapi.ResourceClaimConsumerReference{Resource: "pods",
			Name: podName, UID: types.UID(podName + "Uid")})
	}
	return result
}
