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

	v1 "k8s.io/api/core/v1"
	resourceapi "k8s.io/api/resource/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
)

// SanitizeNodeResourceSlices can be used to duplicate node-local ResourceSlices attached to a Node, when duplicating the Node. The new slices
// are pointed to newNodeName, and nameSuffix is appended to all pool names (pool names have to be unique within a driver, so we can't
// leave them as-is when duplicating).
func SanitizeNodeResourceSlices(nodeLocalSlices []*resourceapi.ResourceSlice, newNodeName, nameSuffix string) (newSlices []*resourceapi.ResourceSlice, oldPoolNames map[string]bool) {
	oldPoolNames = map[string]bool{}
	for _, slice := range nodeLocalSlices {
		sliceCopy := slice.DeepCopy()
		sliceCopy.UID = uuid.NewUUID()
		sliceCopy.Name = fmt.Sprintf("%s-%s", slice.Name, nameSuffix)
		sliceCopy.Spec.Pool.Name = fmt.Sprintf("%s-%s", slice.Spec.Pool.Name, nameSuffix)
		sliceCopy.Spec.NodeName = newNodeName

		oldPoolNames[slice.Spec.Pool.Name] = true
		newSlices = append(newSlices, sliceCopy)
	}
	return newSlices, oldPoolNames
}

// SanitizePodResourceClaims can be used to duplicate ResourceClaims needed by a Pod, when duplicating the Pod.
//   - ResourceClaims owned by oldOwner are duplicated and sanitized, to be owned by a duplicate pod - newOwner.
//   - ResourceClaims not owned by oldOwner are skipped from the result. They are shared claims not bound to the
//     lifecycle of the duplicated pod, so we don't want to duplicate them.
//   - Works for unallocated claims (e.g. if the pod being duplicated isn't scheduled).
//   - Works for claims allocated on a single node that is being duplicated (e.g. if the pod being duplicated is a scheduled DS pod).
//     The name of the old node and its pools have to be provided in this case. Such allocated claims are be pointed to newNodeName,
//     and nameSuffix is appended to all pool names in allocation results, to match the pool names of the new, duplicated node.
//   - Returns an error if any of the allocated claims is not node-local on oldNodeName. Such allocations can't be sanitized, the only
//     option is to clear the allocation and run scheduler filters&reserve to get a new allocation when duplicating a pod.
func SanitizePodResourceClaims(newOwner, oldOwner *v1.Pod, claims []*resourceapi.ResourceClaim, clearAllocations bool, nameSuffix, oldNodeName, newNodeName string, oldNodePoolNames map[string]bool) ([]*resourceapi.ResourceClaim, error) {
	var newResourceClaims []*resourceapi.ResourceClaim
	for _, claim := range claims {
		if ownerName, ownerUid := ClaimOwningPod(claim); ownerName != oldOwner.Name || ownerUid != oldOwner.UID {
			// Only claims owned by the pod are bound to its lifecycle. The lifecycle of other claims is independent, and they're most likely shared
			// by multiple pods. They shouldn't be duplicated when we're duplicating a pod - skip.
			continue
		}

		claimCopy := claim.DeepCopy()
		claimCopy.UID = uuid.NewUUID()
		claimCopy.Name = fmt.Sprintf("%s-%s", claim.Name, nameSuffix)
		claimCopy.OwnerReferences = []metav1.OwnerReference{podClaimOwnerReference(newOwner)}

		if clearAllocations {
			claimCopy.Status.Allocation = nil
		}

		if claimCopy.Status.Allocation == nil {
			// Unallocated claim - just clear the consumer reservations to be sure, and we're done.
			claimCopy.Status.ReservedFor = []resourceapi.ResourceClaimConsumerReference{}
			newResourceClaims = append(newResourceClaims, claimCopy)
			continue
		}

		if singleNodeSelector := nodeSelectorSingleNode(claimCopy.Status.Allocation.NodeSelector); singleNodeSelector != oldNodeName {
			// This claim has an allocation available on more than a single Node. We can't sanitize it, and it shouldn't be duplicated as we'd
			// have multiple claims allocating the same devices.
			return nil, fmt.Errorf("claim %s/%s is allocated, but not node-local on %s - can't be sanitized", claim.Namespace, claim.Name, oldNodeName)
		}

		var sanitizedAllocations []resourceapi.DeviceRequestAllocationResult
		for _, devAlloc := range claim.Status.Allocation.Devices.Results {
			// It's possible to have both node-local and global allocations in a single resource claim. Make sure that all allocations were node-local on the old node.
			if !oldNodePoolNames[devAlloc.Pool] {
				return nil, fmt.Errorf("claim %s/%s has an allocation %s, from a pool that isn't node-local on %s - can't be sanitized", claim.Namespace, claim.Name, devAlloc.Request, oldNodeName)
			}
			devAlloc.Pool = fmt.Sprintf("%s-%s", devAlloc.Pool, nameSuffix)
			sanitizedAllocations = append(sanitizedAllocations, devAlloc)
		}

		claimCopy.Status.Allocation.Devices.Results = sanitizedAllocations
		claimCopy.Status.Allocation.NodeSelector = createNodeSelectorSingleNode(newNodeName)
		claimCopy.Status.ReservedFor = []resourceapi.ResourceClaimConsumerReference{podClaimConsumerReference(newOwner)}

		newResourceClaims = append(newResourceClaims, claimCopy)
	}

	return newResourceClaims, nil
}

// SanitizeResourceClaimRefs returns a duplicate of the provided pod, with nameSuffix appended to all pod-owned ResourceClaim names
// referenced in the Pod object. Names of ResourceClaims not owned by the pod are not changed.
func SanitizeResourceClaimRefs(pod *v1.Pod, nameSuffix string) *v1.Pod {
	podCopy := pod.DeepCopy()
	var sanitizedClaimRefs []v1.PodResourceClaim
	for _, claimRef := range podCopy.Spec.ResourceClaims {
		if claimRef.ResourceClaimTemplateName != nil {
			newTemplateName := fmt.Sprintf("%s-%s", *claimRef.ResourceClaimTemplateName, nameSuffix)
			claimRef.ResourceClaimTemplateName = &newTemplateName
		}
		sanitizedClaimRefs = append(sanitizedClaimRefs, claimRef)
	}
	podCopy.Spec.ResourceClaims = sanitizedClaimRefs

	var sanitizedClaimStatuses []v1.PodResourceClaimStatus
	for _, claimStatus := range podCopy.Status.ResourceClaimStatuses {
		if claimStatus.ResourceClaimName != nil {
			newClaimName := fmt.Sprintf("%s-%s", *claimStatus.ResourceClaimName, nameSuffix)
			claimStatus.ResourceClaimName = &newClaimName
		}
		sanitizedClaimStatuses = append(sanitizedClaimStatuses, claimStatus)
	}
	podCopy.Status.ResourceClaimStatuses = sanitizedClaimStatuses

	return podCopy
}
