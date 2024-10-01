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

import resourceapi "k8s.io/api/resource/v1alpha3"

type snapshotSliceLister Snapshot

func (s snapshotSliceLister) List() ([]*resourceapi.ResourceSlice, error) {
	var result []*resourceapi.ResourceSlice
	for _, slices := range s.resourceSlicesByNodeName {
		for _, slice := range slices {
			result = append(result, slice)
		}
	}
	result = append(result, s.nonNodeLocalResourceSlices...)
	return result, nil
}
