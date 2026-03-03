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

package dynamicresources

import (
	"slices"
	"testing"

	v1 "k8s.io/api/resource/v1"
)

func TestGetDriverNamesForMetrics(t *testing.T) {
	tests := map[string]struct {
		resourceSlices       []*v1.ResourceSlice
		wantDrivers          []string
		wantDriversCompacted string
	}{
		"NilSlice": {
			resourceSlices:       nil,
			wantDrivers:          nil,
			wantDriversCompacted: "",
		},
		"EmptySlice": {
			resourceSlices:       []*v1.ResourceSlice{},
			wantDrivers:          nil,
			wantDriversCompacted: "",
		},
		"OneDriver": {
			resourceSlices: []*v1.ResourceSlice{
				{
					Spec: v1.ResourceSliceSpec{
						Driver: "compute-domain.nvidia.com",
					},
				},
			},
			wantDrivers:          []string{"compute-domain.nvidia.com"},
			wantDriversCompacted: "compute-domain.nvidia.com",
		},
		"TwoDrivers": {
			resourceSlices: []*v1.ResourceSlice{
				{
					Spec: v1.ResourceSliceSpec{
						Driver: "compute-domain.nvidia.com",
					},
				},
				{
					Spec: v1.ResourceSliceSpec{
						Driver: "dra.net",
					},
				},
			},
			wantDrivers:          []string{"compute-domain.nvidia.com", "dra.net"},
			wantDriversCompacted: "compute-domain.nvidia.com,dra.net",
		},
		"TwoDriversUnsorted": {
			resourceSlices: []*v1.ResourceSlice{
				{
					Spec: v1.ResourceSliceSpec{
						Driver: "dra.net",
					},
				},
				{
					Spec: v1.ResourceSliceSpec{
						Driver: "compute-domain.nvidia.com",
					},
				},
			},
			wantDrivers:          []string{"compute-domain.nvidia.com", "dra.net"},
			wantDriversCompacted: "compute-domain.nvidia.com,dra.net",
		},
		"TwoDriversWithDuplicates": {
			resourceSlices: []*v1.ResourceSlice{
				{
					Spec: v1.ResourceSliceSpec{
						Driver: "compute-domain.nvidia.com",
					},
				},
				{
					Spec: v1.ResourceSliceSpec{
						Driver: "dra.net",
					},
				},
				{
					Spec: v1.ResourceSliceSpec{
						Driver: "compute-domain.nvidia.com",
					},
				},
			},
			wantDrivers:          []string{"compute-domain.nvidia.com", "dra.net"},
			wantDriversCompacted: "compute-domain.nvidia.com,dra.net",
		},
		"CustomDriver": {
			resourceSlices: []*v1.ResourceSlice{
				{
					Spec: v1.ResourceSliceSpec{
						Driver: "custom-driver",
					},
				},
			},
			wantDrivers:          []string{customDriverName},
			wantDriversCompacted: customDriverName,
		},
		"KnownDriverAndCustomDriver": {
			resourceSlices: []*v1.ResourceSlice{
				{
					Spec: v1.ResourceSliceSpec{
						Driver: "compute-domain.nvidia.com",
					},
				},
				{
					Spec: v1.ResourceSliceSpec{
						Driver: "custom-driver",
					},
				},
			},
			wantDrivers:          []string{"compute-domain.nvidia.com", customDriverName},
			wantDriversCompacted: "compute-domain.nvidia.com," + customDriverName,
		},
		"KnownDriverAndCustomDriverUnsorted": {
			resourceSlices: []*v1.ResourceSlice{
				{
					Spec: v1.ResourceSliceSpec{
						Driver: "custom-driver",
					},
				},
				{
					Spec: v1.ResourceSliceSpec{
						Driver: "compute-domain.nvidia.com",
					},
				},
			},
			wantDrivers:          []string{"compute-domain.nvidia.com", customDriverName},
			wantDriversCompacted: "compute-domain.nvidia.com," + customDriverName,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			drivers := GetDriverNamesForMetrics(test.resourceSlices)
			if !slices.Equal(drivers, test.wantDrivers) {
				t.Errorf("GetDriverNamesForMetrics() = %v, want %v", drivers, test.wantDrivers)
			}
			compacted := GetDriverNamesForMetricsCompacted(test.resourceSlices)
			if compacted != test.wantDriversCompacted {
				t.Errorf("GetDriverNamesForMetricsCompacted() = %v, want %v", compacted, test.wantDriversCompacted)
			}
		})
	}
}

func BenchmarkGetDriverNamesForMetrics(b *testing.B) {
	resourceSlices := []*v1.ResourceSlice{
		{
			Spec: v1.ResourceSliceSpec{
				Driver: "compute-domain.nvidia.com",
			},
		},
		{
			Spec: v1.ResourceSliceSpec{
				Driver: "dra.net",
			},
		},
		{
			Spec: v1.ResourceSliceSpec{
				Driver: "custom-driver",
			},
		},
	}

	for b.Loop() {
		GetDriverNamesForMetrics(resourceSlices)
	}
}

func BenchmarkGetDriverNamesForMetricsCompacted(b *testing.B) {
	resourceSlices := []*v1.ResourceSlice{
		{
			Spec: v1.ResourceSliceSpec{
				Driver: "compute-domain.nvidia.com",
			},
		},
		{
			Spec: v1.ResourceSliceSpec{
				Driver: "dra.net",
			},
		},
		{
			Spec: v1.ResourceSliceSpec{
				Driver: "custom-driver",
			},
		},
	}

	for b.Loop() {
		GetDriverNamesForMetricsCompacted(resourceSlices)
	}
}
