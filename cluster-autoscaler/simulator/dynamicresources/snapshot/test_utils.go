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
	"github.com/google/go-cmp/cmp"
	resourceapi "k8s.io/api/resource/v1"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/common"
)

// CloneTestSnapshot creates a deep copy of the provided Snapshot.
// This function is intended for testing purposes only.
func CloneTestSnapshot(snapshot *Snapshot) *Snapshot {
	cloneString := func(s string) string { return s }
	cloneResourceClaimId := func(rc ResourceClaimId) ResourceClaimId { return rc }
	cloneDeviceClass := func(dc *resourceapi.DeviceClass) *resourceapi.DeviceClass { return dc.DeepCopy() }
	cloneResourceClaim := func(rc *resourceapi.ResourceClaim) *resourceapi.ResourceClaim { return rc.DeepCopy() }
	cloneResourceSlices := func(rcs []*resourceapi.ResourceSlice) []*resourceapi.ResourceSlice {
		clone := make([]*resourceapi.ResourceSlice, len(rcs))
		for i := range rcs {
			clone[i] = rcs[i].DeepCopy()
		}
		return clone
	}

	deviceClasses := common.ClonePatchSet(snapshot.deviceClasses, cloneString, cloneDeviceClass)
	resourceSlices := common.ClonePatchSet(snapshot.resourceSlices, cloneString, cloneResourceSlices)
	resourceClaims := common.ClonePatchSet(snapshot.resourceClaims, cloneResourceClaimId, cloneResourceClaim)

	return &Snapshot{
		deviceClasses:  deviceClasses,
		resourceSlices: resourceSlices,
		resourceClaims: resourceClaims,
	}
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
