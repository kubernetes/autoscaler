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
						Driver: "driver1",
					},
				},
			},
			wantDrivers:          []string{"driver1"},
			wantDriversCompacted: "driver1",
		},
		"TwoDrivers": {
			resourceSlices: []*v1.ResourceSlice{
				{
					Spec: v1.ResourceSliceSpec{
						Driver: "driver1",
					},
				},
				{
					Spec: v1.ResourceSliceSpec{
						Driver: "driver2",
					},
				},
			},
			wantDrivers:          []string{"driver1", "driver2"},
			wantDriversCompacted: "driver1,driver2",
		},
		"TwoDriversUnsorted": {
			resourceSlices: []*v1.ResourceSlice{
				{
					Spec: v1.ResourceSliceSpec{
						Driver: "driver2",
					},
				},
				{
					Spec: v1.ResourceSliceSpec{
						Driver: "driver1",
					},
				},
			},
			wantDrivers:          []string{"driver1", "driver2"},
			wantDriversCompacted: "driver1,driver2",
		},
		"TwoDriversWithDuplicates": {
			resourceSlices: []*v1.ResourceSlice{
				{
					Spec: v1.ResourceSliceSpec{
						Driver: "driver1",
					},
				},
				{
					Spec: v1.ResourceSliceSpec{
						Driver: "driver2",
					},
				},
				{
					Spec: v1.ResourceSliceSpec{
						Driver: "driver1",
					},
				},
			},
			wantDrivers:          []string{"driver1", "driver2"},
			wantDriversCompacted: "driver1,driver2",
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
				Driver: "driver1",
			},
		},
		{
			Spec: v1.ResourceSliceSpec{
				Driver: "driver2",
			},
		},
		{
			Spec: v1.ResourceSliceSpec{
				Driver: "driver1",
			},
		},
	}

	b.ResetTimer()
	for b.Loop() {
		GetDriverNamesForMetrics(resourceSlices)
	}
}

func BenchmarkGetDriverNamesForMetricsCompacted(b *testing.B) {
	resourceSlices := []*v1.ResourceSlice{
		{
			Spec: v1.ResourceSliceSpec{
				Driver: "driver1",
			},
		},
		{
			Spec: v1.ResourceSliceSpec{
				Driver: "driver2",
			},
		},
		{
			Spec: v1.ResourceSliceSpec{
				Driver: "driver1",
			},
		},
	}

	for b.Loop() {
		GetDriverNamesForMetricsCompacted(resourceSlices)
	}
}
