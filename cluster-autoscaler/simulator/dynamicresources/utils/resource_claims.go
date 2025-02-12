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
	"slices"

	apiv1 "k8s.io/api/core/v1"
	resourceapi "k8s.io/api/resource/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/component-helpers/scheduling/corev1"
	"k8s.io/utils/ptr"
)

// ClaimOwningPod returns the name and UID of the Pod owner of the provided claim. If the claim isn't
// owned by a Pod, empty strings are returned.
func ClaimOwningPod(claim *resourceapi.ResourceClaim) (string, types.UID) {
	for _, owner := range claim.OwnerReferences {
		if ptr.Deref(owner.Controller, false) && owner.APIVersion == "v1" && owner.Kind == "Pod" {
			return owner.Name, owner.UID
		}
	}
	return "", ""
}

// ClaimAllocated returns whether the provided claim is allocated.
func ClaimAllocated(claim *resourceapi.ResourceClaim) bool {
	return claim.Status.Allocation != nil
}

// ClaimInUse returns whether the provided claim is currently reserved for any consumer.
func ClaimInUse(claim *resourceapi.ResourceClaim) bool {
	return len(claim.Status.ReservedFor) > 0
}

// ClaimReservedForPod returns whether the provided claim is currently reserved for the provided pod.
func ClaimReservedForPod(claim *resourceapi.ResourceClaim, pod *apiv1.Pod) bool {
	for _, consumerRef := range claim.Status.ReservedFor {
		if claimConsumerReferenceMatchesPod(pod, consumerRef) {
			return true
		}
	}
	return false
}

// ClaimFullyReserved returns whether the provided claim already has the maximum possible reservations
// set, and no more can be added.
func ClaimFullyReserved(claim *resourceapi.ResourceClaim) bool {
	return len(claim.Status.ReservedFor) >= resourceapi.ResourceClaimReservedForMaxSize
}

// ClaimAvailableOnNode returns whether the provided claim is allocated and available on the provided Node.
func ClaimAvailableOnNode(claim *resourceapi.ResourceClaim, node *apiv1.Node) (bool, error) {
	if !ClaimAllocated(claim) {
		// Not allocated so not available anywhere.
		return false, nil
	}
	selector := claim.Status.Allocation.NodeSelector
	if selector == nil {
		// nil means available everywhere.
		return true, nil
	}
	return corev1.MatchNodeSelectorTerms(node, claim.Status.Allocation.NodeSelector)
}

// DeallocateClaimInPlace clears the allocation of the provided claim.
func DeallocateClaimInPlace(claim *resourceapi.ResourceClaim) {
	claim.Status.Allocation = nil
}

// ClearPodReservationInPlace clears the reservation for the provided pod in the provided Claim. It is a no-op
// if the claim isn't reserved for the Pod.
func ClearPodReservationInPlace(claim *resourceapi.ResourceClaim, pod *apiv1.Pod) {
	if len(claim.Status.ReservedFor) == 0 {
		return
	}

	podReservationIndex := -1
	for i, consumerRef := range claim.Status.ReservedFor {
		if claimConsumerReferenceMatchesPod(pod, consumerRef) {
			podReservationIndex = i
			break
		}
	}
	if podReservationIndex != -1 {
		claim.Status.ReservedFor = slices.Delete(claim.Status.ReservedFor, podReservationIndex, podReservationIndex+1)
	}
}

// AddPodReservationInPlace adds a reservation for the provided pod to the provided Claim. It is a no-op
// if the claim is already reserved for the Pod.
func AddPodReservationInPlace(claim *resourceapi.ResourceClaim, pod *apiv1.Pod) {
	if !ClaimReservedForPod(claim, pod) {
		claim.Status.ReservedFor = append(claim.Status.ReservedFor, PodClaimConsumerReference(pod))
	}
}

// PodClaimConsumerReference returns a consumer reference entry for a ResourceClaim status ReservedFor field.
func PodClaimConsumerReference(pod *apiv1.Pod) resourceapi.ResourceClaimConsumerReference {
	return resourceapi.ResourceClaimConsumerReference{
		Name:     pod.Name,
		UID:      pod.UID,
		Resource: "pods",
		APIGroup: "",
	}
}

// PodClaimOwnerReference returns an OwnerReference for a pod-owned ResourceClaim.
func PodClaimOwnerReference(pod *apiv1.Pod) metav1.OwnerReference {
	truePtr := true
	return metav1.OwnerReference{
		APIVersion:         "v1",
		Kind:               "Pod",
		Name:               pod.Name,
		UID:                pod.UID,
		BlockOwnerDeletion: &truePtr,
		Controller:         &truePtr,
	}
}

func claimConsumerReferenceMatchesPod(pod *apiv1.Pod, ref resourceapi.ResourceClaimConsumerReference) bool {
	return ref.APIGroup == "" && ref.Resource == "pods" && ref.Name == pod.Name && ref.UID == pod.UID
}
