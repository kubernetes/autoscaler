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
)

// GetAllDevices aggregates all Devices from the provided ResourceSlices into one list.
func GetAllDevices(slices []*resourceapi.ResourceSlice) []resourceapi.Device {
	var devices []resourceapi.Device
	for _, slice := range slices {
		devices = append(devices, slice.Spec.Devices...)
	}
	return devices
}

// GroupSlices groups the provided slices by their driver and pool name.
func GroupSlices(slices []*resourceapi.ResourceSlice) map[string]map[string][]*resourceapi.ResourceSlice {
	result := map[string]map[string][]*resourceapi.ResourceSlice{}
	for _, slice := range slices {
		driver := slice.Spec.Driver
		pool := slice.Spec.Pool.Name
		if result[driver] == nil {
			result[driver] = map[string][]*resourceapi.ResourceSlice{}
		}
		result[driver][pool] = append(result[driver][pool], slice)
	}
	return result
}

// AllCurrentGenSlices filters out slices that aren't from the newest pool generation, and returns an error
// if not all slices from the newest generations are provided.
func AllCurrentGenSlices(slices []*resourceapi.ResourceSlice) ([]*resourceapi.ResourceSlice, error) {
	var maxGenSlices []*resourceapi.ResourceSlice
	maxGen := int64(0)
	for _, slice := range slices {
		gen := slice.Spec.Pool.Generation
		if gen > maxGen {
			maxGen = gen
			maxGenSlices = []*resourceapi.ResourceSlice{slice}
			continue
		}
		if gen == maxGen {
			maxGenSlices = append(maxGenSlices, slice)
		}
	}

	foundCurrentSlices := len(maxGenSlices)
	if foundCurrentSlices == 0 {
		return nil, nil
	}

	if wantCurrentSlices := maxGenSlices[0].Spec.Pool.ResourceSliceCount; int64(foundCurrentSlices) != wantCurrentSlices {
		return nil, fmt.Errorf("newest generation: %d, slice count: %d - found only %d slices", maxGen, wantCurrentSlices, foundCurrentSlices)
	}

	return maxGenSlices, nil
}

func nodeSelectorSingleNode(selector *v1.NodeSelector) string {
	if selector == nil {
		// Nil selector means all nodes, so not a single node.
		return ""
	}
	if len(selector.NodeSelectorTerms) != 1 {
		// Selector for a single node doesn't need multiple ORed terms.
		return ""
	}
	term := selector.NodeSelectorTerms[0]
	if len(term.MatchExpressions) > 0 {
		// Selector for a single node doesn't need expression matching.
		return ""
	}
	if len(term.MatchFields) != 1 {
		// Selector for a single node should have just 1 matchFields entry for its nodeName.
		return ""
	}
	matchField := term.MatchFields[0]
	if matchField.Key != "metadata.name" || matchField.Operator != v1.NodeSelectorOpIn || len(matchField.Values) != 1 {
		// Selector for a single node should have operator In with 1 value - the node name.
		return ""
	}
	return matchField.Values[0]
}

func createNodeSelectorSingleNode(nodeName string) *v1.NodeSelector {
	return &v1.NodeSelector{
		NodeSelectorTerms: []v1.NodeSelectorTerm{
			{
				MatchFields: []v1.NodeSelectorRequirement{
					{
						Key:      "metadata.name",
						Operator: v1.NodeSelectorOpIn,
						Values:   []string{nodeName},
					},
				},
			},
		},
	}
}
