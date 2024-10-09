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

package dynamicresources

import (
	"fmt"

	apiv1 "k8s.io/api/core/v1"
	resourceapi "k8s.io/api/resource/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/component-helpers/scheduling/corev1"
	"k8s.io/utils/ptr"
)

// ClaimOwningPod returns the name and UID of the Pod owner of the provided claim. If the claim isn't
// owned by a Pod, empty strings are returned.
func ClaimOwningPod(claim *resourceapi.ResourceClaim) (string, types.UID) {
	for _, owner := range claim.OwnerReferences {
		if ptr.Deref(owner.Controller, false) &&
			owner.APIVersion == "v1" &&
			owner.Kind == "Pod" {
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

// ClaimReservedForPod returns whether the provided claim is currently reserved for the provided Pod.
func ClaimReservedForPod(claim *resourceapi.ResourceClaim, pod *apiv1.Pod) bool {
	for _, consumerRef := range claim.Status.ReservedFor {
		if claimConsumerReferenceMatchesPod(pod, consumerRef) {
			return true
		}
	}
	return false
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
	newReservedFor := make([]resourceapi.ResourceClaimConsumerReference, 0, len(claim.Status.ReservedFor))
	for _, consumerRef := range claim.Status.ReservedFor {
		if claimConsumerReferenceMatchesPod(pod, consumerRef) {
			continue
		}
		newReservedFor = append(newReservedFor, consumerRef)
	}
	claim.Status.ReservedFor = newReservedFor
}

// AddPodReservationIfNeededInPlace adds a reservation for the provided pod to the provided Claim. It is a no-op
// if the claim is already reserved for the Pod. Error is returned if the claim already has the maximum number of
// reservations.
func AddPodReservationIfNeededInPlace(claim *resourceapi.ResourceClaim, pod *apiv1.Pod) error {
	alreadyReservedForPod := false
	for _, consumerRef := range claim.Status.ReservedFor {
		if claimConsumerReferenceMatchesPod(pod, consumerRef) {
			alreadyReservedForPod = true
			break
		}
	}
	if !alreadyReservedForPod {
		if len(claim.Status.ReservedFor) >= resourceapi.ResourceClaimReservedForMaxSize {
			return fmt.Errorf("claim already reserved for %d consumers, can't add more", len(claim.Status.ReservedFor))
		}
		claim.Status.ReservedFor = append(claim.Status.ReservedFor, podClaimConsumerReference(pod))
	}
	return nil
}

// NodeInfoResourceClaims returns all ResourceClaims contained in the PodInfos in this NodeInfo. Shared claims
// are taken into account, each claim should only be returned once.
func NodeInfoResourceClaims(nodeInfo *framework.NodeInfo) []*resourceapi.ResourceClaim {
	processedClaims := map[resourceClaimRef]bool{}
	var result []*resourceapi.ResourceClaim
	for _, pod := range nodeInfo.Pods {
		for _, claim := range pod.NeededResourceClaims {
			if processedClaims[resourceClaimRef{Namespace: claim.Namespace, Name: claim.Name}] {
				// Shared claim, already grouped.
				continue
			}
			result = append(result, claim)
			processedClaims[resourceClaimRef{Namespace: claim.Namespace, Name: claim.Name}] = true
		}
	}
	return result
}

// PodNeedsResourceClaims returns whether the pod references any ResourceClaims.
func PodNeedsResourceClaims(pod *apiv1.Pod) bool {
	return len(pod.Spec.ResourceClaims) > 0
}

func claimConsumerReferenceMatchesPod(pod *apiv1.Pod, ref resourceapi.ResourceClaimConsumerReference) bool {
	return ref.APIGroup == "" && ref.Resource == "pods" && ref.Name == pod.Name && ref.UID == pod.UID
}

func podClaimConsumerReference(pod *apiv1.Pod) resourceapi.ResourceClaimConsumerReference {
	return resourceapi.ResourceClaimConsumerReference{
		Name:     pod.Name,
		UID:      pod.UID,
		Resource: "pods",
		APIGroup: "",
	}
}

func podClaimOwnerReference(pod *apiv1.Pod) metav1.OwnerReference {
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
