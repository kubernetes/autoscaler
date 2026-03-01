/*
Copyright The Kubernetes Authors.

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
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	resourceapi "k8s.io/api/resource/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestDiff(t *testing.T) {
	testCases := map[string]struct {
		subset                   []*resourceapi.ResourceSlice
		superset                 []*resourceapi.ResourceSlice
		wantMissingResourcePools []MissingResourcePool
	}{
		"ExactMatch": {
			subset: []*resourceapi.ResourceSlice{
				makeResourceSlice("pool1", "driver1", "dev1", "dev2"),
			},
			superset: []*resourceapi.ResourceSlice{
				makeResourceSlice("pool1", "driver1", "dev1", "dev2"),
			},
			wantMissingResourcePools: []MissingResourcePool{},
		},
		"SubsetHasMoreDevices": {
			subset: []*resourceapi.ResourceSlice{
				makeResourceSlice("pool1", "driver1", "dev1", "dev2", "dev3"),
			},
			superset: []*resourceapi.ResourceSlice{
				makeResourceSlice("pool1", "driver1", "dev1", "dev2"),
			},
			// Shape mismatch (subset bigger) -> superset pool is unmatched -> missing.
			wantMissingResourcePools: []MissingResourcePool{
				{Driver: "driver1", Pool: "pool1", Devices: []string{"dev1", "dev2"}},
			},
		},
		"SubsetHasLessDevices": {
			subset: []*resourceapi.ResourceSlice{
				makeResourceSlice("pool1", "driver1", "dev1"),
			},
			superset: []*resourceapi.ResourceSlice{
				makeResourceSlice("pool1", "driver1", "dev1", "dev2"),
			},
			// Shape mismatch (subset smaller) -> superset pool is unmatched -> missing.
			wantMissingResourcePools: []MissingResourcePool{
				{Driver: "driver1", Pool: "pool1", Devices: []string{"dev1", "dev2"}},
			},
		},
		"SubsetHasDifferentDevices": {
			subset: []*resourceapi.ResourceSlice{
				makeResourceSlice("pool1", "driver1", "dev3", "dev4"),
			},
			superset: []*resourceapi.ResourceSlice{
				makeResourceSlice("pool1", "driver1", "dev1", "dev2"),
			},
			wantMissingResourcePools: []MissingResourcePool{
				{Driver: "driver1", Pool: "pool1", Devices: []string{"dev1", "dev2"}},
			},
		},
		"IgnoredExtraPoolsInSubset": {
			subset: []*resourceapi.ResourceSlice{
				makeResourceSlice("pool1", "driver1", "dev1"),
				makeResourceSlice("pool2", "driver1", "dev2"),
			},
			superset: []*resourceapi.ResourceSlice{
				makeResourceSlice("pool1", "driver1", "dev1"),
			},
			wantMissingResourcePools: []MissingResourcePool{},
		},
		"MissingPoolsInSubset": {
			subset: []*resourceapi.ResourceSlice{
				makeResourceSlice("pool1", "driver1", "dev1"),
			},
			superset: []*resourceapi.ResourceSlice{
				makeResourceSlice("pool1", "driver1", "dev1"),
				makeResourceSlice("pool2", "driver1", "dev2"),
			},
			wantMissingResourcePools: []MissingResourcePool{
				{Driver: "driver1", Pool: "pool2", Devices: []string{"dev2"}},
			},
		},
		"PoolNameIndependence": {
			subset: []*resourceapi.ResourceSlice{
				makeResourceSlice("pool2", "driver1", "dev1"),
			},
			superset: []*resourceapi.ResourceSlice{
				makeResourceSlice("pool1", "driver1", "dev1"),
			},
			// Should match because logic ignores pool name during matching
			wantMissingResourcePools: []MissingResourcePool{},
		},
		"DuplicatePoolsInSuperset": {
			subset: []*resourceapi.ResourceSlice{
				makeResourceSlice("pool3", "driver1", "dev1"),
				makeResourceSlice("pool4", "driver1", "dev1"),
			},
			superset: []*resourceapi.ResourceSlice{
				makeResourceSlice("pool1", "driver1", "dev1"),
				makeResourceSlice("pool2", "driver1", "dev1"),
			},
			wantMissingResourcePools: []MissingResourcePool{},
		},
		"MultipleSlicesSamePool": {
			subset: []*resourceapi.ResourceSlice{
				makeResourceSlice("pool1", "driver1", "dev1"),
				makeResourceSlice("pool1", "driver1", "dev2"),
			},
			superset: []*resourceapi.ResourceSlice{
				makeResourceSlice("pool1", "driver1", "dev1", "dev2"),
			},
			wantMissingResourcePools: []MissingResourcePool{},
		},
		"InsufficientMatchingPoolsInSubset": {
			subset: []*resourceapi.ResourceSlice{
				makeResourceSlice("pool3", "driver1", "dev1"),
			},
			superset: []*resourceapi.ResourceSlice{
				makeResourceSlice("pool1", "driver1", "dev1"),
				makeResourceSlice("pool2", "driver1", "dev1"),
			},
			wantMissingResourcePools: []MissingResourcePool{
				{Driver: "driver1", Pool: "pool2", Devices: []string{"dev1"}},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			gotMissing := normalizeMissingPools(GetMissingSliceDevices(tc.subset, tc.superset))
			wantMissing := normalizeMissingPools(tc.wantMissingResourcePools)
			wantHas := len(tc.wantMissingResourcePools) == 0
			gotHas := HasSliceDevices(tc.subset, tc.superset)

			// Verify WalkMissingResourcePools matches GetMissingSliceDevices
			var walkedMissing []MissingResourcePool
			WalkMissingResourcePools(tc.subset, tc.superset, func(m MissingResourcePool) bool {
				walkedMissing = append(walkedMissing, m)
				return true
			})
			walkedMissing = normalizeMissingPools(walkedMissing)

			assert.ElementsMatch(t, wantMissing, gotMissing)
			assert.ElementsMatch(t, gotMissing, walkedMissing, "WalkMissingResourcePools result should match GetMissingSliceDevices")
			assert.Equal(t, wantHas, gotHas, "HasSliceDevices mismatch")
		})
	}
}

// normalizeMissingPools sorts the devices in each missing resource pool inplace
// returns the same slice just for convenience.
func normalizeMissingPools(pools []MissingResourcePool) []MissingResourcePool {
	for i := range pools {
		sort.Strings(pools[i].Devices)
	}

	return pools
}

func makeResourceSlice(pool, driver string, devices ...string) *resourceapi.ResourceSlice {
	sliceDevices := make([]resourceapi.Device, len(devices))
	for i, d := range devices {
		sliceDevices[i] = resourceapi.Device{Name: d}
	}
	return &resourceapi.ResourceSlice{
		ObjectMeta: metav1.ObjectMeta{
			Name: "slice-" + pool,
		},
		Spec: resourceapi.ResourceSliceSpec{
			Driver: driver,
			Pool: resourceapi.ResourcePool{
				Name: pool,
			},
			Devices: sliceDevices,
		},
	}
}
