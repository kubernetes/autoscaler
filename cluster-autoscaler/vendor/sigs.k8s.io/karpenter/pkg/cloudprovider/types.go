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
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unique"

	"github.com/awslabs/operatorpkg/serrors"
	"github.com/awslabs/operatorpkg/status"
	"github.com/samber/lo"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"

	v1 "sigs.k8s.io/karpenter/pkg/apis/v1"
	"sigs.k8s.io/karpenter/pkg/operator/options"
	"sigs.k8s.io/karpenter/pkg/scheduling"
	"sigs.k8s.io/karpenter/pkg/utils/resources"
)

//go:generate go tool -modfile=../../go.tools.mod controller-gen object:headerFile="../../hack/boilerplate.go.txt" paths="."

var (
	SpotRequirement     = scheduling.NewRequirements(scheduling.NewRequirement(v1.CapacityTypeLabelKey, corev1.NodeSelectorOpIn, v1.CapacityTypeSpot))
	OnDemandRequirement = scheduling.NewRequirements(scheduling.NewRequirement(v1.CapacityTypeLabelKey, corev1.NodeSelectorOpIn, v1.CapacityTypeOnDemand))
	ReservedRequirement = scheduling.NewRequirements(scheduling.NewRequirement(v1.CapacityTypeLabelKey, corev1.NodeSelectorOpIn, v1.CapacityTypeReserved))

	// ReservationIDLabel is a label injected into a reserved offering's requirements which is used to uniquely identify a
	// reservation. For example, a reservation could be shared across multiple NodePools, and the value encoded in this
	// requirement is used to inform the scheduler that a reservation for one should affect the other.
	ReservationIDLabel string

	// ReservedCapacityLabels is the set of additional labels that are associated with reserved offerings. Each reserved
	// offering should define a requirement for these labels, and all other offerings should define a DoesNotExist requirement.
	ReservedCapacityLabels = sets.New[string]()
)

type DriftReason string

type RepairPolicy struct {
	// ConditionType of unhealthy state that is found on the node
	ConditionType corev1.NodeConditionType
	// ConditionStatus condition when a node is unhealthy
	ConditionStatus corev1.ConditionStatus
	// TolerationDuration is the duration the controller will wait
	// before force terminating nodes that are unhealthy.
	TolerationDuration time.Duration
}

// CloudProvider interface is implemented by cloud providers to support provisioning.
type CloudProvider interface {
	// Create launches a NodeClaim with the given resource requests and requirements and returns a hydrated
	// NodeClaim back with resolved NodeClaim labels for the launched NodeClaim
	Create(context.Context, *v1.NodeClaim) (*v1.NodeClaim, error)
	// Delete removes a NodeClaim from the cloudprovider by its provider id. Delete should return
	// NodeClaimNotFoundError if the cloudProvider instance is already terminated and nil if deletion was triggered.
	// Karpenter will keep retrying until Delete returns a NodeClaimNotFound error.
	Delete(context.Context, *v1.NodeClaim) error
	// Get retrieves a NodeClaim from the cloudprovider by its provider id
	Get(context.Context, string) (*v1.NodeClaim, error)
	// List retrieves all NodeClaims from the cloudprovider
	List(context.Context) ([]*v1.NodeClaim, error)
	// GetInstanceTypes returns instance types supported by the cloudprovider.
	// Availability of types or zone may vary by nodepool or over time.  Regardless of
	// availability, the GetInstanceTypes method should always return all instance types,
	// even those with no offerings available.
	GetInstanceTypes(context.Context, *v1.NodePool) ([]*InstanceType, error)
	// IsDrifted returns whether a NodeClaim has drifted from the provisioning requirements
	// it is tied to.
	IsDrifted(context.Context, *v1.NodeClaim) (DriftReason, error)
	// RepairPolicy is for CloudProviders to define a set Unhealthy condition for Karpenter
	// to monitor on the node.
	RepairPolicies() []RepairPolicy
	// Name returns the CloudProvider implementation name.
	Name() string
	// GetSupportedNodeClasses returns CloudProvider NodeClass that implements status.Object
	// NOTE: It returns a list where the first element should be the default NodeClass
	GetSupportedNodeClasses() []status.Object
}

type NodeLifecycleHookResult struct {
	// Requeue indicates the lifecycle controller should requeue with exponential backoff.
	Requeue bool
	// RequeueAfter indicates the lifecycle controller should requeue after the provided duration.
	// If multiple hooks are provided, the shortest interval is respected.
	RequeueAfter time.Duration
}

// NodeLifecycleHook is implemented by cloud providers to gate node registration.
// All registered hooks must return an empty result before node registration completes.
type NodeLifecycleHook interface {
	// Name for the hook.
	Name() string
	// Registered returns true if this hook's preconditions are satisfied and can proceed.
	Registered(context.Context, *v1.NodeClaim) (NodeLifecycleHookResult, error)
}

// InstanceType describes the properties of a potential node (either concrete attributes of an instance of this type
// or supported options in the case of arrays)
// +k8s:deepcopy-gen=true
type InstanceType struct {
	// Name of the instance type, must correspond to corev1.LabelInstanceTypeStable
	Name string
	// Requirements returns a flexible set of properties that may be selected
	// for scheduling. Must be defined for every well known label, even if empty.
	Requirements scheduling.Requirements
	// Note that though this is an array it is expected that all the Offerings are unique from one another
	Offerings Offerings
	// Resources are the full resource capacities for this instance type
	Capacity corev1.ResourceList
	// DynamicResources contains DRA device metadata for this instance type.
	// Cloud providers that do not support DRA may leave this as the zero value.
	DynamicResources DynamicResources
	// Overhead is the amount of resource overhead expected to be used by kubelet and any other system daemons outside
	// of Kubernetes.
	Overhead               *InstanceTypeOverhead
	once                   sync.Once
	allocatableOfferings   []AllocatableOfferings
	capacityOverlayApplied bool
}

type InstanceTypes []*InstanceType

// Since we have a no copy the sync.Once field, we need to maintain a custom
// DeepCopyInto function.
//
//nolint:gocyclo
func (in *InstanceType) DeepCopyInto(out *InstanceType) {
	out.Name = in.Name
	if in.Requirements != nil {
		in, out := &in.Requirements, &out.Requirements
		*out = make(scheduling.Requirements, len(*in))
		for key, val := range *in {
			var outVal *scheduling.Requirement
			if val == nil {
				(*out)[key] = nil
			} else {
				inVal := (*in)[key]
				in, out := &inVal, &outVal
				*out = new(scheduling.Requirement)
				(*in).DeepCopyInto(*out)
			}
			(*out)[key] = outVal
		}
	}
	if in.Offerings != nil {
		in, out := &in.Offerings, &out.Offerings
		*out = make(Offerings, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(Offering)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	if in.Capacity != nil {
		in, out := &in.Capacity, &out.Capacity
		*out = make(corev1.ResourceList, len(*in))
		for key, val := range *in {
			(*out)[key] = val.DeepCopy()
		}
	}
	in.DynamicResources.DeepCopyInto(&out.DynamicResources)
	if in.Overhead != nil {
		in, out := &in.Overhead, &out.Overhead
		*out = new(InstanceTypeOverhead)
		(*in).DeepCopyInto(*out)
	}
}

// AllocatableOfferings pairs an allocatable resource set with the offerings that produce it.
type AllocatableOfferings struct {
	Allocatable corev1.ResourceList
	Offerings   Offerings
}

// precompute is used to ensure we only compute the allocatable resources once as it's called many times
// and the operation is fairly expensive.
func (i *InstanceType) precompute() {
	// Fast path: most instance types have no override offerings.
	// Skip map/order/fmt.Sprintf machinery when not needed.
	hasOverrides := false
	for _, o := range i.Offerings {
		if len(o.CapacityOverride) > 0 || o.OverheadOverride != nil {
			hasOverrides = true
			break
		}
	}
	if !hasOverrides {
		i.allocatableOfferings = []AllocatableOfferings{
			{Allocatable: i.computeAllocatable(nil, nil), Offerings: i.Offerings.Available()},
		}
		return
	}

	i.allocatableOfferings = i.groupOfferingsByOverride()
}

// groupOfferingsByOverride groups available offerings by their (CapacityOverride, OverheadOverride) tuple,
// computing an allocatable for each group. The first group is always the base (no overrides).
func (i *InstanceType) groupOfferingsByOverride() []AllocatableOfferings {
	type overrideKey struct {
		capacity string
		overhead string
	}
	groups := map[overrideKey]*AllocatableOfferings{}
	baseKey := overrideKey{}
	groups[baseKey] = &AllocatableOfferings{}
	order := []overrideKey{baseKey}

	for _, o := range i.Offerings {
		if !o.Available {
			continue
		}
		if len(o.CapacityOverride) == 0 && o.OverheadOverride == nil {
			groups[baseKey].Offerings = append(groups[baseKey].Offerings, o)
			continue
		}
		key := overrideKey{
			capacity: fmt.Sprintf("%v", o.CapacityOverride),
			overhead: fmt.Sprintf("%v", o.OverheadOverride),
		}
		if _, exists := groups[key]; !exists {
			groups[key] = &AllocatableOfferings{}
			order = append(order, key)
		}
		groups[key].Offerings = append(groups[key].Offerings, o)
	}

	// Build allocatable for each group
	result := make([]AllocatableOfferings, 0, len(order))
	for idx, key := range order {
		group := groups[key]
		if idx == 0 {
			group.Allocatable = i.computeAllocatable(nil, nil)
		} else {
			// Use the first offering in the group to get the override values
			o := group.Offerings[0]
			group.Allocatable = i.computeAllocatable(o.CapacityOverride, o.OverheadOverride)
		}
		result = append(result, *group)
	}
	return result
}

// computeAllocatable computes the allocatable resources for a given capacity/overhead override.
// If both are nil, it computes the base allocatable.
func (i *InstanceType) computeAllocatable(capacityOverride corev1.ResourceList, overheadOverride *InstanceTypeOverhead) corev1.ResourceList {
	capacity := i.Capacity
	if len(capacityOverride) > 0 {
		capacity = resources.Merge(i.Capacity, capacityOverride)
	}
	overhead := i.Overhead.Total()
	if overheadOverride != nil {
		overhead = resources.Merge(overhead, overheadOverride.Total())
	}
	allocatable := resources.Subtract(capacity, overhead)

	// Adjust allocatable memory to account for hugepage reservations.
	for name, quantity := range capacity {
		if strings.HasPrefix(string(name), corev1.ResourceHugePagesPrefix) {
			current := allocatable.Memory()
			current.Sub(quantity)
			if current.Sign() == -1 {
				current.Set(0)
			}
			allocatable[corev1.ResourceMemory] = lo.FromPtr(current)
		}
	}
	return allocatable
}

// OfferingPrice returns the price for the offering matching the given zone and
// capacity type. Returns 0, false if no matching offering exists.
func (i *InstanceType) OfferingPrice(zone, capacityType string) (float64, bool) {
	for _, o := range i.Offerings {
		if o.Zone() == zone && o.CapacityType() == capacityType {
			return o.Price, true
		}
	}
	return 0, false
}

func (i *InstanceType) IsPricingOverlayApplied() bool {
	return lo.ContainsBy(i.Offerings, func(of *Offering) bool {
		return of.IsPriceOverlaid()
	})
}

func (i *InstanceType) ApplyCapacityOverlay(updatedCapacity corev1.ResourceList) {
	i.Capacity = lo.Assign(i.Capacity, updatedCapacity)
	i.capacityOverlayApplied = true
}

func (i *InstanceType) IsCapacityOverlayApplied() bool {
	return i.capacityOverlayApplied
}

// AllocatableOfferingsList returns all allocatable groups for this instance type.
// Each group pairs an allocatable resource set with the offerings that produce it.
// The first entry is always the base allocatable (no overrides).
func (i *InstanceType) AllocatableOfferingsList() []AllocatableOfferings {
	i.once.Do(i.precompute)
	return i.allocatableOfferings
}

// Allocatable returns the base allocatable resources (no offering overrides applied).
func (i *InstanceType) Allocatable() corev1.ResourceList {
	i.once.Do(i.precompute)
	return i.allocatableOfferings[0].Allocatable
}

func (its InstanceTypes) OrderByPrice(reqs scheduling.Requirements) InstanceTypes {
	// Order instance types so that we get the cheapest instance types of the available offerings
	sort.Slice(its, func(i, j int) bool {
		iPrice := math.MaxFloat64
		jPrice := math.MaxFloat64

		for _, of := range its[i].Offerings {
			if of.Available && reqs.IsCompatible(of.Requirements, scheduling.AllowUndefinedWellKnownLabels) && of.Price < iPrice {
				iPrice = of.Price
			}
		}
		for _, of := range its[j].Offerings {
			if of.Available && reqs.IsCompatible(of.Requirements, scheduling.AllowUndefinedWellKnownLabels) && of.Price < jPrice {
				jPrice = of.Price
			}
		}
		return iPrice < jPrice
	})
	return its
}

// Compatible returns the list of instanceTypes based on the supported capacityType and zones in the requirements
func (its InstanceTypes) Compatible(requirements scheduling.Requirements) InstanceTypes {
	var filteredInstanceTypes []*InstanceType
	for _, instanceType := range its {
		if instanceType.Offerings.Available().HasCompatible(requirements) {
			filteredInstanceTypes = append(filteredInstanceTypes, instanceType)
		}
	}
	return filteredInstanceTypes
}

// SatisfiesMinValues validates whether the InstanceTypes satisfies the minValues requirements
// It returns the minimum number of needed instance types to satisfy the minValues requirement, and if min values isn't satisfied, a map containing the keys which don't satisfy min values and an error
// that indicates whether the InstanceTypes satisfy the passed-in requirements
// This minNeededInstanceTypes value is dependent on the ordering of instance types, so relying on this value in a
// deterministic way implies that the instance types are sorted ahead of using this method
// For example:
// Requirements:
//   - key: node.kubernetes.io/instance-type
//     operator: In
//     values: ["c4.large","c4.xlarge","c5.large","c5.xlarge","m4.large","m4.xlarge"]
//     minValues: 3
//   - key: karpenter.kwok.sh/instance-family
//     operator: In
//     values: ["c4","c5","m4"]
//     minValues: 3
//
// InstanceTypes: ["c4.large","c5.xlarge","m4.2xlarge"], it PASSES the requirements
//
//		we get the map as : {
//			node.kubernetes.io/instance-type:  ["c4.large","c5.xlarge","m4.2xlarge"],
//			karpenter.k8s.aws/instance-family: ["c4","c5","m4"]
//		}
//	 so it returns 3 and a nil error to indicate a minimum of 3 instance types were required to fulfill the minValues requirements
//
// And if InstanceTypes: ["c4.large","c4.xlarge","c5.2xlarge"], it FAILS the requirements
//
//		we get the map as : {
//			node.kubernetes.io/instance-type:  ["c4.large","c4.xlarge","c5.2xlarge"],
//			karpenter.k8s.aws/instance-family: ["c4","c5"] // minimum requirement failed for this.
//		}
//	  so it returns 3 and a non-nil error to indicate that the instance types weren't able to fulfill the minValues requirements
func (its InstanceTypes) SatisfiesMinValues(requirements scheduling.Requirements) (minNeededInstanceTypes int, unsatisfiableMinValues map[string]int, err error) {
	if !requirements.HasMinValues() {
		return 0, nil, nil
	}
	incompatibleKeys := map[string]int{}
	valuesForKey := map[string]sets.Set[string]{}
	// We validate if sorting by price and truncating the number of instance types to minItems breaks the minValue requirement.
	// If minValue requirement fails, we return an error that indicates the first requirement key that couldn't be satisfied.
	for i, it := range its {
		for _, req := range requirements {
			if req.MinValues != nil {
				if _, ok := valuesForKey[req.Key]; !ok {
					valuesForKey[req.Key] = sets.New[string]()
				}
				valuesForKey[req.Key] = valuesForKey[req.Key].Insert(it.Requirements.Get(req.Key).Values()...)
			}
		}
		for k, v := range valuesForKey {
			// Collect all the min values that are violated
			if len(v) < lo.FromPtr(requirements.Get(k).MinValues) {
				incompatibleKeys[k] = len(v)
			} else {
				// If the key now satisfies min values, remove it from the map.
				delete(incompatibleKeys, k)
			}
		}
		if len(incompatibleKeys) == 0 {
			return i + 1, nil, nil
		}
	}
	if len(incompatibleKeys) != 0 {
		return len(its), incompatibleKeys, serrors.Wrap(fmt.Errorf("minValues requirement is not met for label(s)"), "label(s)", lo.Keys(incompatibleKeys))
	}
	return len(its), nil, nil
}

// Truncate truncates the InstanceTypes based on the passed-in requirements
// It returns an error if it isn't possible to truncate the instance types on maxItems without violating minValues
func (its InstanceTypes) Truncate(ctx context.Context, requirements scheduling.Requirements, maxItems int) (InstanceTypes, error) {
	truncatedInstanceTypes := lo.Slice(its.OrderByPrice(requirements), 0, maxItems)
	// Only check for a validity of NodeClaim if its requirement has minValues in it.
	if requirements.HasMinValues() {
		// If minValues is NOT met for any of the requirement across InstanceTypes, then only allow it if min values policy is set to BestEffort.
		if options.FromContext(ctx).MinValuesPolicy != options.MinValuesPolicyBestEffort {
			if _, _, err := truncatedInstanceTypes.SatisfiesMinValues(requirements); err != nil {
				return its, fmt.Errorf("validating minValues, %w", err)
			}
		}
	}
	return truncatedInstanceTypes, nil
}

// +k8s:deepcopy-gen=true
type InstanceTypeOverhead struct {
	// KubeReserved returns the default resources allocated to kubernetes system daemons by default
	KubeReserved corev1.ResourceList
	// SystemReserved returns the default resources allocated to the OS system daemons by default
	SystemReserved corev1.ResourceList
	// EvictionThreshold returns the resources used to maintain a hard eviction threshold
	EvictionThreshold corev1.ResourceList
}

func (i InstanceTypeOverhead) Total() corev1.ResourceList {
	return resources.Merge(i.KubeReserved, i.SystemReserved, i.EvictionThreshold)
}

// An Offering describes where an InstanceType is available to be used, with the expectation that its properties
// may be tightly coupled (e.g. the availability of an instance type in some zone is scoped to a capacity type) and
// these properties are captured with labels in Requirements.
// Requirements are required to contain the keys v1.CapacityTypeLabelKey and corev1.LabelTopologyZone.
// +k8s:deepcopy-gen=true
type Offering struct {
	Requirements        scheduling.Requirements
	Price               float64
	Available           bool
	ReservationCapacity int

	// CapacityOverride specifies resource overrides for this offering's capacity.
	// Values are merged with the instance type's base capacity — new keys are added,
	// existing keys are replaced. If nil, the offering uses the base capacity as-is.
	CapacityOverride corev1.ResourceList
	// OverheadOverride specifies overhead overrides for this offering.
	// Values are merged with the instance type's base overhead — new keys are added,
	// existing keys are replaced. If nil, the offering uses the base overhead as-is.
	OverheadOverride *InstanceTypeOverhead

	priceOverlayApplied bool
}

func (o *Offering) ApplyPriceOverlay(UpdatedPrice string) {
	o.Price = AdjustedPrice(o.Price, UpdatedPrice)
	o.priceOverlayApplied = true
}

func AdjustedPrice(instanceTypePrice float64, change string) float64 {
	// if price or price adjustment is not defined, then we will return the same price
	if change == "" {
		return instanceTypePrice
	}

	// if price is defined, then we will return the value given in the overlay
	if !strings.HasPrefix(change, "+") && !strings.HasPrefix(change, "-") {
		return lo.Must(strconv.ParseFloat(change, 64))
	}

	// Check if adjustment is a percentage
	isPercentage := strings.HasSuffix(change, "%")
	adjustment := change

	var adjustedPrice float64
	if isPercentage {
		adjustment = strings.TrimSuffix(change, "%")
		// Parse the adjustment value
		// Due to the CEL validation we can assume that
		// there will always be a valid float provided into the spec
		adjustedPrice = instanceTypePrice * (1 + (lo.Must(strconv.ParseFloat(adjustment, 64)) / 100))
	} else {
		adjustedPrice = instanceTypePrice + lo.Must(strconv.ParseFloat(adjustment, 64))
	}

	// Parse the adjustment value
	// Due to the CEL validation we can assume that
	// there will always be a valid float provided into the spec

	// Apply the adjustment
	return lo.Ternary(adjustedPrice >= 0, adjustedPrice, 0)
}

func (o *Offering) IsPriceOverlaid() bool {
	return o.priceOverlayApplied
}

func (o *Offering) CapacityType() string {
	return o.Requirements.Get(v1.CapacityTypeLabelKey).Any()
}

func (o *Offering) Zone() string {
	return o.Requirements.Get(corev1.LabelTopologyZone).Any()
}

func (o *Offering) ReservationID() string {
	return o.Requirements.Get(ReservationIDLabel).Any()
}

// +k8s:deepcopy-gen=true
type Offerings []*Offering

// Available filters the available offerings from the returned offerings
func (ofs Offerings) Available() Offerings {
	return lo.Filter(ofs, func(o *Offering, _ int) bool {
		return o.Available
	})
}

// Compatible returns the offerings based on the passed requirements
func (ofs Offerings) Compatible(reqs scheduling.Requirements) Offerings {
	return lo.Filter(ofs, func(offering *Offering, _ int) bool {
		return reqs.IsCompatible(offering.Requirements, scheduling.AllowUndefinedWellKnownLabels)
	})
}

// HasCompatible returns whether there is a compatible offering based on the passed requirements
func (ofs Offerings) HasCompatible(reqs scheduling.Requirements) bool {
	for _, of := range ofs {
		if reqs.IsCompatible(of.Requirements, scheduling.AllowUndefinedWellKnownLabels) {
			return true
		}
	}
	return false
}

// Cheapest returns the cheapest offering from the returned offerings
func (ofs Offerings) Cheapest() *Offering {
	return lo.MinBy(ofs, func(a, b *Offering) bool {
		return a.Price < b.Price
	})
}

// MostExpensive returns the most expensive offering from the return offerings
func (ofs Offerings) MostExpensive() *Offering {
	return lo.MaxBy(ofs, func(a, b *Offering) bool {
		return a.Price > b.Price
	})
}

// WorstLaunchPrice gets the worst-case launch price from the offerings that are offered on an instance type. Only
// offerings for the capacity type we will launch with are considered. The following precedence order is used to
// determine which capacity type is used: reserved, spot, on-demand.
func (ofs Offerings) WorstLaunchPrice(reqs scheduling.Requirements) float64 {
	for _, ctReqs := range []scheduling.Requirements{
		ReservedRequirement,
		SpotRequirement,
		OnDemandRequirement,
	} {
		if compatOfs := ofs.Compatible(reqs).Compatible(ctReqs); len(compatOfs) != 0 {
			return compatOfs.MostExpensive().Price
		}
	}
	return math.MaxFloat64
}

// NodeClaimNotFoundError is an error type returned by CloudProviders when the reason for failure is NotFound
type NodeClaimNotFoundError struct {
	error
}

func NewNodeClaimNotFoundError(err error) *NodeClaimNotFoundError {
	return &NodeClaimNotFoundError{
		error: err,
	}
}

func (e *NodeClaimNotFoundError) Error() string {
	return fmt.Sprintf("nodeclaim not found, %s", e.error)
}

func (e *NodeClaimNotFoundError) Unwrap() error {
	return e.error
}

func IsNodeClaimNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	var ncnfErr *NodeClaimNotFoundError
	return errors.As(err, &ncnfErr)
}

func IgnoreNodeClaimNotFoundError(err error) error {
	if IsNodeClaimNotFoundError(err) {
		return nil
	}
	return err
}

// InsufficientCapacityError is an error type returned by CloudProviders when a launch fails due to a lack of capacity from NodeClaim requirements
type InsufficientCapacityError struct {
	error
}

func NewInsufficientCapacityError(err error) *InsufficientCapacityError {
	return &InsufficientCapacityError{
		error: err,
	}
}

func (e *InsufficientCapacityError) Error() string {
	return fmt.Sprintf("insufficient capacity, %s", e.error)
}

func (e *InsufficientCapacityError) Unwrap() error {
	return e.error
}

func IsInsufficientCapacityError(err error) bool {
	if err == nil {
		return false
	}
	var icErr *InsufficientCapacityError
	return errors.As(err, &icErr)
}

// NodeClassNotReadyError is an error type returned by CloudProviders when a NodeClass that is used by the launch process doesn't have all its resolved fields
type NodeClassNotReadyError struct {
	error
}

func NewNodeClassNotReadyError(err error) *NodeClassNotReadyError {
	return &NodeClassNotReadyError{
		error: err,
	}
}

func (e *NodeClassNotReadyError) Error() string {
	return fmt.Sprintf("NodeClassRef not ready, %s", e.error)
}

func (e *NodeClassNotReadyError) Unwrap() error {
	return e.error
}

func IsNodeClassNotReadyError(err error) bool {
	if err == nil {
		return false
	}
	var nrError *NodeClassNotReadyError
	return errors.As(err, &nrError)
}

// CreateError is an error type returned by CloudProviders when instance creation fails
type CreateError struct {
	error
	ConditionReason  string
	ConditionMessage string
}

func NewCreateError(err error, reason, message string) *CreateError {
	return &CreateError{
		error:            err,
		ConditionReason:  reason,
		ConditionMessage: message,
	}
}

func (e *CreateError) Error() string {
	return fmt.Sprintf("creating nodeclaim, %s", e.error)
}

func (e *CreateError) Unwrap() error {
	return e.error
}

// UnevaluatedNodePoolError is an error when the NodePool isn't ready for evaluation
type UnevaluatedNodePoolError struct {
	nodePoolName string
}

func NewUnevaluatedNodePoolError(nodePoolName string) *UnevaluatedNodePoolError {
	return &UnevaluatedNodePoolError{
		nodePoolName: nodePoolName,
	}
}

func (e *UnevaluatedNodePoolError) Error() string {
	return fmt.Sprintf("nodepool %q is awaiting evaluation", e.nodePoolName)
}

func IsUnevaluatedNodePoolError(err error) bool {
	if err == nil {
		return false
	}
	var onatnpErr *UnevaluatedNodePoolError
	return errors.As(err, &onatnpErr)
}

// DeviceID is a hashable, unique ID for a device. This ID is absolute - depending on the driver, pool, and device
// names - as opposed to relative - depending on in-memory indexes.
type DeviceID struct {
	Driver unique.Handle[string]
	Pool   unique.Handle[string]
	Device unique.Handle[string]
}

func (id DeviceID) String() string {
	return fmt.Sprintf("%s/%s/%s", id.Driver.Value(), id.Pool.Value(), id.Device.Value())
}
