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
