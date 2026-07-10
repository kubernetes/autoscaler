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
	"strings"

	resourcev1 "k8s.io/api/resource/v1"
	"k8s.io/apimachinery/pkg/util/sets"

	"sigs.k8s.io/karpenter/pkg/cloudprovider"
)

// Constraint is the interface for inter-device constraints evaluated during the DFS.
// Constraints are stateful — Add() modifies internal state (e.g., pinning an attribute value)
// and Remove() reverses exactly one successful Add(). This enables backtracking.
type Constraint interface {
	// Add is called when allocating a device. Returns false if the constraint is violated.
	// If false is returned, the constraint's state must not be modified.
	Add(requestName string, device cloudprovider.Device, deviceID DeviceID) bool
	// Remove is called during backtracking to reverse one successful Add().
	Remove(requestName string, device cloudprovider.Device, deviceID DeviceID)
	// Reset clears all mutable state, returning the constraint to its initial (unpinned) condition.
	Reset()
}

// MatchAttributeConstraint enforces that all devices allocated for the constrained requests
// share a common attribute value. The first device pins the value; subsequent devices must match.
//
// For in-flight devices where the attribute is absent (runtime-only), the constraint falls
// back to consulting AttributeBindings — see the AttributeBindingFallback field.
type MatchAttributeConstraint struct {
	// RequestNames is the set of request names this constraint applies to.
	// If empty, the constraint applies to all requests in the claim.
	RequestNames sets.Set[string]
	// AttributeName is the fully qualified attribute name to match across devices.
	AttributeName resourcev1.QualifiedName
	// AttributeValue holds the pinned attribute value from the first allocated device.
	// Nil when the first device used a binding fallback (runtime-only attribute).
	AttributeValue *resourcev1.DeviceAttribute

	// AttributeBindingFallback, if non-nil, is consulted when a device does not have
	// the attribute in its template. This enables MatchAttribute constraints for
	// runtime-only attributes that are declared via cloud provider AttributeBindings.
	AttributeBindingFallback *AttributeBindingFallback
	// UsedBinding is true when the constraint was established via attribute binding
	// fallback (first device lacked a concrete value). Once set, devices with concrete
	// attribute values are rejected to prevent mixing evaluation paths.
	UsedBinding bool
	// AllocatedDeviceIDs tracks the DeviceIDs of devices that have been added, in order.
	// Used for attribute binding lookups and to determine constraint state (empty = no pin).
	AllocatedDeviceIDs []DeviceID
}

// AttributeBindingFallback holds the context needed to check attribute bindings when
// a device's template is missing the constrained attribute.
type AttributeBindingFallback struct {
	Bindings       AttributeBindings
	NodePool       string
	InstanceTypeID InstanceTypeID
}

//nolint:gocyclo
func (m *MatchAttributeConstraint) Add(requestName string, device cloudprovider.Device, deviceID DeviceID) bool {
	if !m.appliesTo(requestName) {
		return true
	}

	attribute := LookupAttribute(device, deviceID, m.AttributeName)

	// Concrete and binding paths are mutually exclusive: once a constraint is established
	// via one path, devices evaluated via the other path are rejected. This prevents
	// mixing concrete attribute comparison with binding fallback within the same constraint.
	if attribute != nil {
		// Reject concrete values if the constraint was established via binding fallback.
		if m.UsedBinding {
			return false
		}
		if len(m.AllocatedDeviceIDs) == 0 {
			m.AttributeValue = attribute
			m.AllocatedDeviceIDs = append(m.AllocatedDeviceIDs, deviceID)
			return true
		}
		if !AttributeValuesEqual(m.AttributeValue, attribute) {
			return false
		}
		m.AllocatedDeviceIDs = append(m.AllocatedDeviceIDs, deviceID)
		return true
	}
	// Reject binding fallback if the constraint was established via concrete values.
	if len(m.AllocatedDeviceIDs) > 0 && !m.UsedBinding {
		return false
	}

	fb := m.AttributeBindingFallback
	// Attribute is absent — try attribute binding fallback.
	if fb == nil {
		return false
	}
	// Reject devices that don't participate in any binding group for this attribute.
	if !fb.Bindings.HasBindings(fb.NodePool, fb.InstanceTypeID, m.AttributeName, deviceID) {
		return false
	}
	if len(m.AllocatedDeviceIDs) == 0 {
		m.UsedBinding = true
		m.AllocatedDeviceIDs = append(m.AllocatedDeviceIDs, deviceID)
		return true
	}

	// Bindings are transitive, so checking against any one previously allocated device
	// is sufficient — if deviceID is bound to one, it's bound to all.
	if !fb.Bindings.Bound(fb.NodePool, fb.InstanceTypeID, m.AttributeName, m.AllocatedDeviceIDs[0], deviceID) {
		return false
	}
	m.AllocatedDeviceIDs = append(m.AllocatedDeviceIDs, deviceID)
	return true
}

func (m *MatchAttributeConstraint) Remove(requestName string, device cloudprovider.Device, deviceID DeviceID) {
	if !m.appliesTo(requestName) {
		return
	}
	if len(m.AllocatedDeviceIDs) > 0 {
		m.AllocatedDeviceIDs = m.AllocatedDeviceIDs[:len(m.AllocatedDeviceIDs)-1]
	}
	if len(m.AllocatedDeviceIDs) == 0 {
		m.AttributeValue = nil
		m.UsedBinding = false
	}
}

func (m *MatchAttributeConstraint) Reset() {
	m.AttributeValue = nil
	m.UsedBinding = false
	m.AllocatedDeviceIDs = nil
}

func (m *MatchAttributeConstraint) appliesTo(requestName string) bool {
	if m.RequestNames.Len() == 0 {
		return true
	}
	return m.RequestNames.Has(requestName)
}

// LookupAttribute finds a device attribute by its fully qualified name. If not found directly,
// it tries a driver-qualified fallback: if the attribute name has a domain prefix matching the
// device's driver, the lookup is retried with just the ID part.
func LookupAttribute(device cloudprovider.Device, deviceID DeviceID, attributeName resourcev1.QualifiedName) *resourcev1.DeviceAttribute {
	if attr, ok := device.Attributes[attributeName]; ok {
		return &attr
	}
	// Try driver-qualified fallback.
	domain, id, ok := strings.Cut(string(attributeName), "/")
	if ok && domain == deviceID.Driver.Value() {
		if attr, ok := device.Attributes[resourcev1.QualifiedName(id)]; ok {
			return &attr
		}
	}
	return nil
}

// AttributeValuesEqual compares two DeviceAttribute values for equality.
func AttributeValuesEqual(a, b *resourcev1.DeviceAttribute) bool {
	if a == nil || b == nil {
		return false
	}
	// Compare by type — exactly one field should be set per attribute.
	if a.StringValue != nil && b.StringValue != nil {
		return *a.StringValue == *b.StringValue
	}
	if a.IntValue != nil && b.IntValue != nil {
		return *a.IntValue == *b.IntValue
	}
	if a.BoolValue != nil && b.BoolValue != nil {
		return *a.BoolValue == *b.BoolValue
	}
	if a.VersionValue != nil && b.VersionValue != nil {
		return *a.VersionValue == *b.VersionValue
	}
	return false
}
