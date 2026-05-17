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

package v1

import "sort"

// GetUpdateModes returns all UpdateModes
func GetUpdateModes() map[UpdateMode]any {
	return map[UpdateMode]any{
		UpdateModeOff:               nil,
		UpdateModeInitial:           nil,
		UpdateModeRecreate:          nil,
		UpdateModeAuto:              nil,
		UpdateModeInPlaceOrRecreate: nil,
		UpdateModeInPlace:           nil,
	}
}

// GetUpdateModesList returns all supported UpdateModes as a slice of strings
// Note: This list will not return deprecated modes.
func GetUpdateModesList() []string {
	modes := GetUpdateModes()
	result := make([]string, 0, len(modes))
	for mode := range modes {
		if mode != UpdateModeAuto { // Skip the deprecated one
			result = append(result, string(mode))
		}
	}
	sort.Strings(result)
	return result
}

// GetScalingModes returns all supported ScalingModes
func GetScalingModes() map[ContainerScalingMode]any {
	return map[ContainerScalingMode]any{
		ContainerScalingModeAuto: nil,
		ContainerScalingModeOff:  nil,
	}
}

// GetPossibleScalingModes returns all supported ScalingModes as a slice of strings
func GetPossibleScalingModes() []string {
	modes := GetScalingModes()
	result := make([]string, 0, len(modes))
	for mode := range modes {
		result = append(result, string(mode))
	}
	sort.Strings(result)
	return result
}
