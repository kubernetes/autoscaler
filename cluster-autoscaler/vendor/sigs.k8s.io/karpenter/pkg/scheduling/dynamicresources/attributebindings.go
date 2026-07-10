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
	"maps"
	"unique"

	resourcev1 "k8s.io/api/resource/v1"
	"k8s.io/apimachinery/pkg/util/sets"

	"sigs.k8s.io/karpenter/pkg/cloudprovider"
)

// AttributeBindings models attributes that share a value across devices on an instance, where
// the concrete value is not known at scheduling simulation time. Bindings are declared by cloud
// providers and used to satisfy MatchAttribute constraints for runtime-only attributes.
//
// Bindings are transitive: if device A is bound to B and B is bound to C under the same
// (attribute, nodePool, instanceType) triple, then A is also bound to C. Transitivity is
// computed during construction via BFS closure.
//
// Example: A ResourceClaim requires a GPU and NIC that share a PCI root. The PCI root ID is
// only known after node launch, but the cloud provider declares that specific GPU-NIC pairs
// on a given instance type will share this attribute.
type AttributeBindings map[resourcev1.QualifiedName]map[string]map[InstanceTypeID]map[cloudprovider.DeviceID]sets.Set[cloudprovider.DeviceID]

// HasBindings returns true if the given device has any attribute binding entries for the specified
// attribute, nodePool, and instanceType. This is used to check whether a device participates in
// any binding group before accepting it into a constraint via the binding fallback path.
func (ab AttributeBindings) HasBindings(nodePool string, instanceType InstanceTypeID, attribute resourcev1.QualifiedName, device DeviceID) bool {
	attributeBindings, ok := ab[attribute]
	if !ok {
		return false
	}
	nodePoolBindings, ok := attributeBindings[nodePool]
	if !ok {
		return false
	}
	itBindings, ok := nodePoolBindings[instanceType]
	if !ok {
		return false
	}
	_, ok = itBindings[device.DeviceID]
	return ok
}

// Bound returns true if deviceA and deviceB are bound under the given attribute for the specified
// nodePool and instanceType. When deviceA and deviceB are the same device, Bound returns true if
// the device participates in any binding for the attribute (i.e. has a non-empty bound set).
// Returns false if any component of the lookup is missing.
func (ab AttributeBindings) Bound(nodePool string, instanceType InstanceTypeID, attribute resourcev1.QualifiedName, deviceA DeviceID, deviceB DeviceID) bool {
	attributeBindings, ok := ab[attribute]
	if !ok {
		return false
	}
	nodePoolBindings, ok := attributeBindings[nodePool]
	if !ok {
		return false
	}
	itBindings, ok := nodePoolBindings[instanceType]
	if !ok {
		return false
	}
	deviceBindings, ok := itBindings[deviceA.DeviceID]
	if !ok {
		return false
	}
	if deviceA.DeviceID == deviceB.DeviceID {
		return deviceBindings.Len() > 0
	}
	return deviceBindings.Has(deviceB.DeviceID)
}

// BuildAttributeBindings constructs the attribute binding graph from cloud provider instance type
// metadata. It creates symmetric pairs for each declared binding and computes the transitive
// closure per (attribute, nodePool, instanceType) triple.
//
//nolint:gocyclo
func BuildAttributeBindings(instanceTypesByNodePool map[string][]*cloudprovider.InstanceType) AttributeBindings {
	bindings := make(map[resourcev1.QualifiedName]map[string]map[InstanceTypeID]map[cloudprovider.DeviceID]sets.Set[cloudprovider.DeviceID])
	for nodePool, instanceTypes := range instanceTypesByNodePool {
		for _, instanceType := range instanceTypes {
			for _, itBinding := range instanceType.DynamicResources.AttributeBindings {
				if len(itBinding.Devices) < 2 {
					continue
				}
				bindingsForAttribute, ok := bindings[itBinding.Attribute]
				if !ok {
					bindingsForAttribute = make(map[string]map[InstanceTypeID]map[cloudprovider.DeviceID]sets.Set[cloudprovider.DeviceID])
					bindings[itBinding.Attribute] = bindingsForAttribute
				}
				bindingsForNodePool, ok := bindingsForAttribute[nodePool]
				if !ok {
					bindingsForNodePool = make(map[InstanceTypeID]map[cloudprovider.DeviceID]sets.Set[cloudprovider.DeviceID])
					bindingsForAttribute[nodePool] = bindingsForNodePool
				}
				bindingsForInstanceType, ok := bindingsForNodePool[unique.Make(instanceType.Name)]
				if !ok {
					bindingsForInstanceType = make(map[cloudprovider.DeviceID]sets.Set[cloudprovider.DeviceID])
					bindingsForNodePool[unique.Make(instanceType.Name)] = bindingsForInstanceType
				}
				for i := range itBinding.Devices {
					bindingsForDevice, ok := bindingsForInstanceType[itBinding.Devices[i]]
					if !ok {
						bindingsForDevice = sets.New[cloudprovider.DeviceID]()
						bindingsForInstanceType[itBinding.Devices[i]] = bindingsForDevice
					}
					for j := range itBinding.Devices {
						if i == j {
							continue
						}
						bindingsForDevice.Insert(itBinding.Devices[j])
					}
				}
			}
		}
	}
	// Compute transitive closure for each (attribute, nodePool, instanceType) triple. If A and B
	// and B and C are declared in separate AttributeBindings, then A and C must also hold. We BFS from
	// each device over the original (direct) binding graph, then replace the device's set with
	// all reachable devices. Closures are computed from the original sets before any replacements
	// to avoid contaminating mid-pass results.
	for _, nodePoolBindings := range bindings {
		for _, instanceTypeBindings := range nodePoolBindings {
			for _, deviceBindings := range instanceTypeBindings {
				closures := make(map[cloudprovider.DeviceID]sets.Set[cloudprovider.DeviceID], len(deviceBindings))
				for device := range deviceBindings {
					visited := sets.New[cloudprovider.DeviceID]()
					queue := []cloudprovider.DeviceID{device}
					for len(queue) > 0 {
						curr := queue[0]
						queue = queue[1:]
						if visited.Has(curr) {
							continue
						}
						visited.Insert(curr)
						for neighbor := range deviceBindings[curr] {
							if !visited.Has(neighbor) {
								queue = append(queue, neighbor)
							}
						}
					}
					visited.Delete(device) // a device is not bound to itself
					closures[device] = visited
				}
				maps.Copy(deviceBindings, closures)
			}
		}
	}
	return bindings
}
