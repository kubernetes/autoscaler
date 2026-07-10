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

package cloudprovider

import (
	"unique"

	resourcev1 "k8s.io/api/resource/v1"
)

// DynamicResources contains DRA device metadata for an instance type. Cloud providers populate this
// to enable DRA-aware scheduling simulation.
type DynamicResources struct {
	// ResourceSliceTemplates describes the DRA devices expected on this instance type.
	// Each entry is an in-memory template for a ResourceSlice that the corresponding
	// DRA driver is expected to publish once the node is running.
	ResourceSliceTemplates []*ResourceSliceTemplate

	// AttributeBindings declares sets of devices that will share a common attribute
	// value at runtime, even though the concrete value is not known at simulation time.
	AttributeBindings []*AttributeBinding
}

// ResourceSliceTemplate is the cloud provider's in-memory representation of a ResourceSlice.
// It only describes node-local devices — node-affinity fields are omitted since all
// cloud-provider-supplied templates are implicitly node-local.
type ResourceSliceTemplate struct {
	// Driver is the DRA driver name (e.g., "gpu.nvidia.com").
	Driver unique.Handle[string]
	// Pool identifies the device pool this slice belongs to.
	Pool ResourcePool
	// Devices is the list of expected devices in this slice.
	Devices []Device
	// SharedCounters is the list of expected counter sets in this slice.
	// A ResourceSliceTemplate with SharedCounters must not also contain Devices —
	// these are mutually exclusive. If both are set, the devices will be silently
	// ignored during pool gathering.
	SharedCounters []resourcev1.CounterSet
}

// ResourcePool identifies a pool of devices within a driver.
type ResourcePool struct {
	Name unique.Handle[string]
}

// Device describes a single device within a ResourceSlice, with attributes for CEL
// selector evaluation and MatchAttribute constraint checking.
type Device struct {
	// Name uniquely identifies this device within its pool.
	Name unique.Handle[string]
	// Attributes are the device's properties used for CEL selector evaluation and
	// MatchAttribute constraint checking.
	Attributes map[resourcev1.QualifiedName]resourcev1.DeviceAttribute
	// Capacity defines the set of capacities for this device.
	Capacity map[resourcev1.QualifiedName]resourcev1.DeviceCapacity
	// AllowMultipleAllocations marks whether the device is allowed to be allocated
	// to multiple DeviceRequests.
	AllowMultipleAllocations bool
	// ConsumesCounters declares which sharedCounters this device consumes from
	// on allocation.
	ConsumesCounters []resourcev1.DeviceCounterConsumption
}

// AttributeBinding declares that a set of devices on an instance type will share a common
// value for a given attribute at runtime. This enables the allocator to satisfy MatchAttribute
// constraints for runtime-only attributes that cannot be included in device templates.
type AttributeBinding struct {
	// Devices is the set of devices that share the attribute value. Must contain
	// at least 2 devices; bindings with fewer are ignored.
	Devices []DeviceID
	// Attribute is the fully qualified attribute name that the devices share.
	Attribute resourcev1.QualifiedName
}
