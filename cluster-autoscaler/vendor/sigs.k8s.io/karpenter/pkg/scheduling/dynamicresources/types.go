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
	"fmt"
	"unique"

	"github.com/samber/lo"
	corev1 "k8s.io/api/core/v1"
	resourcev1 "k8s.io/api/resource/v1"
	"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/karpenter/pkg/cloudprovider"
	"sigs.k8s.io/karpenter/pkg/scheduling"
)

// ID types used throughout the allocator. These use unique.Handle for efficient comparison.
type (
	DriverID        = unique.Handle[string]
	PoolID          = unique.Handle[string]
	InstanceTypeID  = unique.Handle[string]
	NodeClaimID     = unique.Handle[string]
	NodePoolID      = unique.Handle[string]
	ResourceClaimID = unique.Handle[types.NamespacedName]
)

func resourceClaimID(claim *resourcev1.ResourceClaim) ResourceClaimID {
	return unique.Make(types.NamespacedName{Namespace: claim.Namespace, Name: claim.Name})
}

// DeviceID wraps cloudprovider.DeviceID with scheduling-specific metadata.
// Template indicates whether the device comes from a cloud provider template
// (potential device) rather than an in-cluster published ResourceSlice.
type DeviceID struct {
	cloudprovider.DeviceID
	Template bool
}

func (id *DeviceID) String() string {
	return fmt.Sprintf("%s%s/%s/%s", lo.Ternary(id.Template, "virtual/", ""), id.Driver.Value(), id.Pool.Value(), id.Device.Value())
}

// NodeClaim abstracts over existing nodes, pre-initialized nodes, and in-flight NodeClaims
// for the allocator. This allows the allocator to work uniformly regardless of the NodeClaim's
// lifecycle phase.
type NodeClaim interface {
	// ID returns a unique identifier for this NodeClaim.
	ID() NodeClaimID
	// NodeName returns the name of the concrete node backing this NodeClaim, or "" for in-flight/new NodeClaims that
	// don't yet have a node. Published ResourceSlices pinned to a node via spec.nodeName are only accessible from the
	// existing node with this name.
	NodeName() string
	// NodePoolID returns the NodePool this NodeClaim belongs to.
	NodePoolID() NodePoolID
	// Requirements returns the current scheduling requirements.
	Requirements() scheduling.Requirements
	// InstanceTypes returns the candidate instance types. For existing nodes, this is a single entry.
	InstanceTypes() []InstanceTypeID
	// ResourceSlices returns per-instance-type in-flight ResourceSlice templates.
	// Empty for existing initialized nodes. For pre-initialized nodes, contains outstanding
	// templates under the single known instance type. For in-flight NodeClaims, contains
	// all candidate instance types' templates.
	ResourceSlices() map[InstanceTypeID][]ResourceSlice
}

// ResourceSlice is the allocator's abstraction over both in-cluster (API server) ResourceSlices
// and cloud-provider-supplied ResourceSliceTemplates. This interface avoids copying all in-cluster
// slices into a common struct on every scheduling loop.
type ResourceSlice interface {
	// Driver returns the DRA driver name for this slice.
	Driver() unique.Handle[string]
	// Pool returns the resource pool this slice belongs to.
	Pool() cloudprovider.ResourcePool
	// Devices returns the devices in this slice.
	Devices() []cloudprovider.Device
	// Potential returns true if this slice represents potential (not yet published) devices
	// from a cloud provider template. Returns false for in-cluster slices published to the
	// API server.
	//
	// The allocator uses this distinction for:
	//   - Tracking allocations: potential devices are tracked per-NodeClaim and per-InstanceType;
	//     published devices are tracked globally.
	//   - Requirement narrowing: published (non-potential) non-node-local devices may carry
	//     topology requirements that constrain the NodeClaim. Potential devices are always
	//     node-local and do not constrain topology.
	Potential() bool
	// NodeName returns the name of the node the slice's devices are local to when the slice pins itself to a single
	// node via spec.nodeName, or "" otherwise. A node-name-pinned slice is accessible only from that exact node, so it
	// can never satisfy an in-flight NodeClaim — only an existing node with the same name.
	NodeName() string
	// NodeSelector returns the node selector if the slice uses label-based node affinity, or nil.
	NodeSelector() *corev1.NodeSelector
	// AllNodes returns true if the slice's devices are accessible from all nodes.
	AllNodes() bool
	// Generation returns the pool generation this slice belongs to. Newer generations
	// supersede older ones. Template slices return 0.
	Generation() int64
	// ResourceSliceCount returns the total number of slices expected in this pool.
	// Used to determine pool completeness. Template slices return 1.
	ResourceSliceCount() int64
	// SharedCounters returns the counter set definitions declared by this slice.
	// Returns nil if the slice declares no counter sets.
	SharedCounters() []resourcev1.CounterSet
}

// apiServerSlice adapts a resourcev1.ResourceSlice from the API server to the ResourceSlice interface.
type apiServerSlice struct {
	slice *resourcev1.ResourceSlice
	// devices is lazily populated on first access to avoid repeated conversion.
	devices []cloudprovider.Device
}

// NewAPIServerSlice wraps an API server ResourceSlice to implement the ResourceSlice interface.
func NewAPIServerSlice(slice *resourcev1.ResourceSlice) ResourceSlice {
	return &apiServerSlice{slice: slice}
}

func (s *apiServerSlice) Driver() unique.Handle[string] {
	return unique.Make(s.slice.Spec.Driver)
}

func (s *apiServerSlice) Pool() cloudprovider.ResourcePool {
	return cloudprovider.ResourcePool{
		Name: unique.Make(s.slice.Spec.Pool.Name),
	}
}

func (s *apiServerSlice) Devices() []cloudprovider.Device {
	if s.devices == nil {
		s.devices = make([]cloudprovider.Device, len(s.slice.Spec.Devices))
		for i, d := range s.slice.Spec.Devices {
			attrs := make(map[resourcev1.QualifiedName]resourcev1.DeviceAttribute, len(d.Attributes))
			for k, v := range d.Attributes {
				attrs[k] = v
			}
			capacity := make(map[resourcev1.QualifiedName]resourcev1.DeviceCapacity, len(d.Capacity))
			for k, v := range d.Capacity {
				capacity[k] = v
			}
			s.devices[i] = cloudprovider.Device{
				Name:                     unique.Make(d.Name),
				Attributes:               attrs,
				Capacity:                 capacity,
				AllowMultipleAllocations: lo.FromPtr(d.AllowMultipleAllocations),
				ConsumesCounters:         d.ConsumesCounters,
			}
		}
	}
	return s.devices
}

func (s *apiServerSlice) Potential() bool {
	return false
}

func (s *apiServerSlice) NodeName() string {
	return lo.FromPtr(s.slice.Spec.NodeName)
}

func (s *apiServerSlice) NodeSelector() *corev1.NodeSelector {
	return s.slice.Spec.NodeSelector
}

func (s *apiServerSlice) AllNodes() bool {
	if s.slice.Spec.AllNodes == nil {
		return false
	}
	return *s.slice.Spec.AllNodes
}

func (s *apiServerSlice) Generation() int64 {
	return s.slice.Spec.Pool.Generation
}

func (s *apiServerSlice) ResourceSliceCount() int64 {
	return s.slice.Spec.Pool.ResourceSliceCount
}

func (s *apiServerSlice) SharedCounters() []resourcev1.CounterSet {
	return s.slice.Spec.SharedCounters
}

// templateSlice adapts a cloudprovider.ResourceSliceTemplate to the ResourceSlice interface.
type templateSlice struct {
	template *cloudprovider.ResourceSliceTemplate
}

// NewTemplateSlice wraps a cloud provider ResourceSliceTemplate to implement the ResourceSlice interface.
func NewTemplateSlice(t *cloudprovider.ResourceSliceTemplate) ResourceSlice {
	return &templateSlice{template: t}
}

func (s *templateSlice) Driver() unique.Handle[string] {
	return s.template.Driver
}

func (s *templateSlice) Pool() cloudprovider.ResourcePool {
	return s.template.Pool
}

func (s *templateSlice) Devices() []cloudprovider.Device {
	return s.template.Devices
}

func (s *templateSlice) Potential() bool {
	return true
}

func (s *templateSlice) NodeName() string {
	return ""
}

func (s *templateSlice) NodeSelector() *corev1.NodeSelector {
	return nil
}

func (s *templateSlice) AllNodes() bool {
	return false
}

func (s *templateSlice) Generation() int64 {
	return 0
}

func (s *templateSlice) ResourceSliceCount() int64 {
	return 1
}

func (s *templateSlice) SharedCounters() []resourcev1.CounterSet {
	return s.template.SharedCounters
}

// nodeSelectorsToRequirements extracts scheduling requirements from a NodeSelector.
// Returns nil if the NodeSelector is nil (no topology constraints).
func nodeSelectorsToRequirements(ns *corev1.NodeSelector) *scheduling.Requirements {
	if ns == nil {
		return nil
	}
	reqs := scheduling.NewRequirements()
	for _, term := range ns.NodeSelectorTerms {
		termReqs := scheduling.NewNodeSelectorRequirements(term.MatchExpressions...)
		reqs.Add(termReqs.Values()...)
	}
	return &reqs
}
