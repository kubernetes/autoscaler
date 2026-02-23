package dynamicresources

import (
	"slices"
	"strings"

	v1 "k8s.io/api/resource/v1"
)

// GetDriverNamesForMetrics returns a slice of sorted unique driver names for the given node info.
// For metrics we want to ensure that we always get the same order of driver names
// to reduce cardinality.
func GetDriverNamesForMetrics(resourceSlices []*v1.ResourceSlice) []string {
	if len(resourceSlices) == 0 {
		return nil
	}

	driverNames := make([]string, len(resourceSlices))
	for i, resourceSlice := range resourceSlices {
		driverNames[i] = resourceSlice.Spec.Driver
	}

	slices.Sort(driverNames)
	return slices.Compact(driverNames)
}

// GetDriverNamesForMetricsCompacted returns a comma separated string of sorted unique driver names.
func GetDriverNamesForMetricsCompacted(resourceSlices []*v1.ResourceSlice) string {
	drivers := GetDriverNamesForMetrics(resourceSlices)
	return strings.Join(drivers, ",")
}
