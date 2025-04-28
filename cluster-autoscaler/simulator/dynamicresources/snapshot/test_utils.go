/*
Copyright 2025 The Kubernetes Authors.

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
	"maps"

	"github.com/google/go-cmp/cmp"
	resourceapi "k8s.io/api/resource/v1beta1"
)

// CloneTestSnapshot creates a deep copy of the provided Snapshot.
// This function is intended for testing purposes only.
func CloneTestSnapshot(snapshot *Snapshot) *Snapshot {
	cloned := &Snapshot{
		deviceClasses:  newPatchSet[string, *resourceapi.DeviceClass](),
		resourceSlices: newPatchSet[string, []*resourceapi.ResourceSlice](),
		resourceClaims: newPatchSet[ResourceClaimId, *resourceapi.ResourceClaim](),
	}

	for i := 0; i < len(snapshot.deviceClasses.patches); i++ {
		devicesPatch := snapshot.deviceClasses.patches[i]
		clonedPatch := newPatch[string, *resourceapi.DeviceClass]()
		for key, deviceClass := range devicesPatch.Modified {
			clonedPatch.Modified[key] = deviceClass.DeepCopy()
		}
		maps.Copy(clonedPatch.Deleted, devicesPatch.Deleted)
		cloned.deviceClasses.patches = append(cloned.deviceClasses.patches, clonedPatch)
	}

	for i := 0; i < len(snapshot.resourceSlices.patches); i++ {
		slicesPatch := snapshot.resourceSlices.patches[i]
		clonedPatch := newPatch[string, []*resourceapi.ResourceSlice]()
		for key, slices := range slicesPatch.Modified {
			deepCopySlices := make([]*resourceapi.ResourceSlice, len(slices))
			for i, slice := range slices {
				deepCopySlices[i] = slice.DeepCopy()
			}

			clonedPatch.Modified[key] = deepCopySlices
		}
		maps.Copy(clonedPatch.Deleted, slicesPatch.Deleted)
		cloned.resourceSlices.patches = append(cloned.resourceSlices.patches, clonedPatch)
	}

	for i := 0; i < len(snapshot.resourceClaims.patches); i++ {
		claimsPatch := snapshot.resourceClaims.patches[i]
		clonedPatch := newPatch[ResourceClaimId, *resourceapi.ResourceClaim]()
		for claimId, claim := range claimsPatch.Modified {
			clonedPatch.Modified[claimId] = claim.DeepCopy()
		}
		maps.Copy(clonedPatch.Deleted, claimsPatch.Deleted)
		cloned.resourceClaims.patches = append(cloned.resourceClaims.patches, clonedPatch)
	}

	return cloned
}

// SnapshotFlattenedComparer returns a cmp.Option that provides a custom comparer function
// for comparing two *Snapshot objects based on their underlying data maps, it doesn't
// compare the underlying patchsets, instead flattened objects are compared.
// This function is intended for testing purposes only.
func SnapshotFlattenedComparer() cmp.Option {
	return cmp.Comparer(func(a, b *Snapshot) bool {
		if a == nil || b == nil {
			return a == b
		}

		devicesEqual := cmp.Equal(a.deviceClasses.AsMap(), b.deviceClasses.AsMap())
		slicesEqual := cmp.Equal(a.resourceSlices.AsMap(), b.resourceSlices.AsMap())
		claimsEqual := cmp.Equal(a.resourceClaims.AsMap(), b.resourceClaims.AsMap())

		return devicesEqual && slicesEqual && claimsEqual
	})
}
