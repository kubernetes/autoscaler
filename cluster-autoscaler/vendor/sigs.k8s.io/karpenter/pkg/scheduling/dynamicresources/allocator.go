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
	"context"
	"fmt"
	"time"
	"unique"

	"github.com/awslabs/operatorpkg/serrors"
	"github.com/samber/lo"
	corev1 "k8s.io/api/core/v1"
	resourcev1 "k8s.io/api/resource/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	dracel "k8s.io/dynamic-resource-allocation/cel"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"sigs.k8s.io/karpenter/pkg/cloudprovider"
	"sigs.k8s.io/karpenter/pkg/scheduling"
)

// DRA device allocation involves a DFS traversal of a decision tree that can be arbitrarily deep. We set an upper
// limit on the time spent per pod allocation to ensure the scheduler makes progress.
const allocateTimeout = 5 * time.Second

// Allocator manages DRA device allocation across a single scheduling loop. It is shared across
// all per-pod allocation requests and is read-only during Allocate() calls. Mutation occurs only
// during initialization and Commit().
type Allocator struct {
	allocationTracker *AllocationTracker
	attributeBindings AttributeBindings
	kubeClient        client.Client

	// inClusterSlices represents the set of ResourceSlices that are already present on the API server.
	inClusterSlices []ResourceSlice
	// poolCache contains the filtered set of available pools for each NodeClaim. Each time a NodeClaim is constrained,
	// the set of pools available to that NodeClaim is also constrained. This cache tracks the last set of constrained
	// pools to use as a baseline for subsequent filtering operations.
	poolCache map[NodeClaimID][]*Pool
	// claimAllocationMetadata contains metadata for in-memory ResourceClaim allocations, including the allocated devices
	// and the scheduling requirements.
	claimAllocationMetadata map[ResourceClaimID]*ResourceClaimAllocationMetadata
	// deletingPodUIDs is the set of pods that are migrating off their current nodes (pods on deleting nodes, disruption
	// candidates, and individually-deleting pods). A ResourceClaim whose reservedFor is composed entirely of these pods
	// is treated as unallocated and re-run through the DFS, so its device — already freed from the allocated-device seed
	// set for the same reason — is re-allocated rather than treated as committed in place.
	deletingPodUIDs sets.Set[types.UID]
}

// ResourceClaimAllocationMetadataForClaim returns a copy of the allocator's internal ResourceClaim allocation metadata.
// A nil result will be returned if the claim hasn't been allocated by the allocator, i.e. both unallocated claims and
// claims allocated in-cluster will return nil.
func (a *Allocator) ResourceClaimAllocationMetadataForClaim(key types.NamespacedName) *ResourceClaimAllocationMetadata {
	meta, ok := a.claimAllocationMetadata[unique.Make(key)]
	if !ok {
		return nil
	}
	return meta
}

// ResourceClaimAllocationMetadataForClaim returns a copy of the allocator's internal ResourceClaim allocation metadata.
// A nil result will be returned if the claim hasn't been allocated by the allocator, i.e. both unallocated claims and
// claims allocated in-cluster will return nil.
func (a *Allocator) ResourceClaimAllocationMetadata() map[ResourceClaimID]*ResourceClaimAllocationMetadata {
	return a.claimAllocationMetadata
}

// ResourceClaimAllocationMetadata represents thte current allocation state for a given ResourceClaim. This includes
// the NodeClaim the ResourceClaim was transitively allocated for, the topology requirements that will be associated
// with the ResourceClaim, and the set of devices allocated to the ResourceClaim on a per-instance type basis.
type ResourceClaimAllocationMetadata struct {
	// NodeClaimID is the NodeClaim that is transitively associated with the ResourceClaim's allocation, via the pod it
	// was allocated for. If the ResourceClaim was satisfied using any template devices, pods referencing this
	// ResourceClaim may not be bound to any other NodeClaim.
	NodeClaimID NodeClaimID

	// ContributedRequirements represents the requirements that are contributed by each instance type. Each ResourceClaim
	// is transitively associated with a single NodeClaim, via the pod it was allocated for. For each instance type the
	// NodeClaim is superposed across, a different device may be selected to satisfy the claim. In this scenario,
	// depending on the instance type the NodeClaim collapses to, the topology requirements associated with the
	// ResourceClaim may differ. To account for this, we pessimistically treat the ResourceClaim's topology requirements
	// as the intersection of contributed requirements.
	//
	// Consider the following example: a device available in zones A and B is allocated when simulating instance type A
	// and a device available in zone B is allocated for instance type B. While both instance types are candidates, we
	// consider the ResourceClaim to only be available in zone B. If instance type B is released, we'll consider it
	// available in both A and B.
	//
	// It is the responsibility of the allocator to prune instance types that would result in an empty requirement set.
	// For example, let's say we have two instance types A and B. Instance type A is evaluated first and contributes a
	// zone IN A requirement. Instance type B is evaluated second and would contribute a zone IN B requirement. Since this
	// would result in an empty set when intersected with A's requirement, instance type B should be pruned by the
	// allocator. This accounts for an existing limitation in NodeClaim modeling - we can't say "instance type A in zone A
	// OR instance type B in zone B". We may evaluate options for removing this restriction in the future.
	//
	// TODO: Currently contributed requirements aren't reflected in the NodeClaims themselves. When an instance type is
	// released, a NodeClaim's requirements aren't relaxed in the same way as the ResourceClaim's. This is a future
	// optimization we could make to increase the flexibility of generated NodeClaims.
	ContributedRequirements map[InstanceTypeID]scheduling.Requirements

	// totalRequirements represents the intersection of the contributed requirements. These requirements are updated each
	// time instance types are released. If a pod referencing this ResourceClaim is to schedule to a different NodeClaim,
	// that NodeClaim must be compatible with these requirements. We take the set intersection because we don't know which
	// instance type will be selected for the "source" NodeClaim (the NodeClaim represented by NodeClaimID).
	TotalRequirements scheduling.Requirements

	// UsedTemplateDevices is true if the allocation used any template (potential) devices,
	// making the claim node-local to the original NodeClaim.
	UsedTemplateDevices bool

	// Devices represents the devices that will be allocated if the original allocating NodeClaim collapses to a given
	// InstanceType. Each entry includes the device ID and any consumed capacity (for multi-allocatable devices).
	// It isn't strictly necessary to retain this information, but it's used in integration testing to form bindings.
	Devices map[InstanceTypeID][]DeviceAllocationResult
}

// DeviceAllocationResult pairs a device ID with the capacity consumed by this specific allocation.
// ConsumedCapacity is nil for exclusive (non-multi-allocatable) devices.
type DeviceAllocationResult struct {
	DeviceID         DeviceID
	ConsumedCapacity map[resourcev1.QualifiedName]resource.Quantity
}

type AllocatedDeviceState struct {
	// ExclusiveDevices contains devices that are exclusively allocated (one claim owns them).
	ExclusiveDevices sets.Set[cloudprovider.DeviceID]
	// ConsumedCapacity maps multi-allocatable devices to their aggregated consumed capacity.
	ConsumedCapacity map[cloudprovider.DeviceID]map[resourcev1.QualifiedName]resource.Quantity
}

// NewAllocator constructs an Allocator for a single scheduling loop.
// allocatedState contains the set of in-cluster devices that are already allocated
// (exclusive devices and multi-allocatable device consumed capacity).
func NewAllocator(
	inClusterSlices []ResourceSlice,
	allocatedState AllocatedDeviceState,
	attributeBindings AttributeBindings,
	kubeClient client.Client,
	deletingPodUIDs sets.Set[types.UID],
) *Allocator {
	if deletingPodUIDs == nil {
		deletingPodUIDs = sets.New[types.UID]()
	}
	a := &Allocator{
		allocationTracker:       NewAllocationTracker(allocatedState),
		attributeBindings:       attributeBindings,
		kubeClient:              kubeClient,
		inClusterSlices:         inClusterSlices,
		poolCache:               make(map[NodeClaimID][]*Pool),
		claimAllocationMetadata: make(map[ResourceClaimID]*ResourceClaimAllocationMetadata),
		deletingPodUIDs:         deletingPodUIDs,
	}
	// Pre-initialize shared counter budgets for all pools. This ensures InitRemainingCounters is
	// never called during parallel Allocate() calls, keeping the tracker read-only under concurrency.
	for _, pool := range GatherPools(inClusterSlices, scheduling.NewRequirements(), "") {
		a.allocationTracker.InitRemainingCounters(pool)
	}
	return a
}

// AllocationResult contains the output of a successful Allocate() call.
type AllocationResult struct {
	// InstanceTypes is the set of instance types whose allocation succeeded.
	InstanceTypes []InstanceTypeID
	// Requirements contains the accumulated topology requirements from all sources:
	// already-allocated claims (in-cluster and in-memory) and newly allocated devices.
	Requirements scheduling.Requirements
	// Allocation is the opaque handle used to commit the allocation to the Allocator's state.
	Allocation Allocation
}

// Allocation is the interface for committing a successful allocation to the Allocator's shared state.
type Allocation interface {
	Commit(context.Context)
}

// allocation commits per-instance-type device allocations (both in-cluster and template).
type allocation struct {
	// allocator is a reference to the top-level allocator that will be mutated when this allocation is committed.
	allocator *Allocator

	// nodeClaimID represents the source NodeClaim the devices were transitvely allocated for
	nodeClaimID NodeClaimID
	// deviceIDsByIT represents the devices that will be allocated for each instance type. Both in-cluster and template
	// devices are included.
	deviceIDsByIT map[InstanceTypeID][]DeviceID
	// claimMetadata represents the allocation metadata for each ResourceClaim, keyed by the ResourceClaim name. This
	// tracks both devices (for observability / testing) and topology requirements (for non-node local binding).
	claimMetadata map[ResourceClaimID]*ResourceClaimAllocationMetadata
	// filteredPools represents the set of pools that will be available from the NodeClaim if this allocation is committed.
	// This reduces the number of pools we need to filter during subsequent allocations for the NodeClaim.
	filteredPools []*Pool
	// counterConsumptionByIT holds per-IT in-cluster counter deductions from this allocation.
	// Stored in the tracker on Commit() to enable precise release when instance types are pruned.
	counterConsumptionByIT map[InstanceTypeID]map[PoolKey]map[string]map[string]resourcev1.Counter
	// templateCounterConsumptionByIT holds per-IT template counter deductions from this allocation.
	// Subtracted from the tracker's remaining budget on Commit(); the entry is deleted on Release.
	templateCounterConsumptionByIT map[InstanceTypeID]map[PoolKey]map[string]map[string]resourcev1.Counter
	// capacityConsumptionByIT holds per-IT consumed capacity for in-cluster multi-allocatable devices.
	// Stored in the tracker on Commit() using the same pessimistic-max pattern as counters.
	capacityConsumptionByIT map[InstanceTypeID]map[DeviceID]map[resourcev1.QualifiedName]resource.Quantity
	// templateCapacityConsumptionByIT holds per-IT consumed capacity for template multi-allocatable
	// devices. Subtracted from the tracker's remaining budget on Commit().
	templateCapacityConsumptionByIT map[InstanceTypeID]map[DeviceID]map[resourcev1.QualifiedName]resource.Quantity
	// templateCounterTotalsByIT holds the total counter budgets for template pools that were computed
	// locally during Allocate(). Written to the tracker on Commit() to keep Allocate() read-only.
	templateCounterTotalsByIT map[InstanceTypeID]map[PoolKey]map[string]map[string]resourcev1.Counter
}

func (a *allocation) Commit(ctx context.Context) {
	if log.FromContext(ctx).V(1).Enabled() {
		log.FromContext(ctx).V(1).Info(
			"allocated devices",
			"nodeClaimID", a.nodeClaimID.Value(),
			"devicesByResourceClaim", lo.MapEntries(a.claimMetadata, func(claimID ResourceClaimID, meta *ResourceClaimAllocationMetadata) (string, map[string][]string) {
				return claimID.Value().String(), lo.MapEntries(meta.Devices, func(it InstanceTypeID, results []DeviceAllocationResult) (string, []string) {
					return it.Value(), lo.Map(results, func(r DeviceAllocationResult, _ int) string { return r.DeviceID.String() })
				})
			}),
		)
	}
	a.allocator.allocationTracker.Commit(a)
	a.allocator.poolCache[a.nodeClaimID] = a.filteredPools
	for claimID, meta := range a.claimMetadata {
		if _, ok := a.allocator.claimAllocationMetadata[claimID]; ok {
			panic("attempted to commit claim which was already allocated")
		}
		a.allocator.claimAllocationMetadata[claimID] = meta
	}
}

// ReleaseInstanceType removes all device allocations for a specific instance type on a NodeClaim.
// Called by the scheduler when an instance type is pruned from a NodeClaim's candidate set.
// Once all instance types referencing a device are released, the device becomes available
// to other NodeClaims.
func (a *Allocator) ReleaseInstanceType(ctx context.Context, nodeClaimID NodeClaimID, instanceTypeIDs ...InstanceTypeID) {
	a.allocationTracker.ReleaseInstanceTypes(ctx, nodeClaimID, instanceTypeIDs...)

	for _, meta := range a.claimAllocationMetadata {
		if meta.NodeClaimID != nodeClaimID {
			continue
		}

		needsRecomputation := false
		for _, it := range instanceTypeIDs {
			if len(meta.ContributedRequirements[it]) != 0 {
				needsRecomputation = true
			}
			delete(meta.ContributedRequirements, it)
			delete(meta.Devices, it)
		}

		// If any of the pruned instance types contributed to the total requirements for the ResourceClaim, we should
		// recompute the requirements. Although this is not strictly necessary, it may unblock placement of subsequent pods
		// in the simulation which were previously blocked due to these constraints. Doing this in a single simulation
		// increases the upper-bound for binpacking efficiency.
		if needsRecomputation {
			updatedReqs := scheduling.NewRequirements()
			for _, itReqs := range meta.ContributedRequirements {
				for _, req := range itReqs {
					updatedReqs.Add(req)
				}
			}
			meta.TotalRequirements = updatedReqs
		}
	}
}

// Allocate attempts to satisfy the given ResourceClaims for the specified NodeClaim. It returns
// an AllocationResult on success or an error if allocation is not possible.
//
// All claims are passed regardless of allocation state. The allocator classifies each claim:
//   - Allocated (in-cluster): status.allocation is set. Topology requirements are extracted
//     and merged into the baseline. No DFS needed.
//   - Allocated (in-memory): a prior pod in this scheduling loop allocated this claim.
//     Metadata is used to validate compatibility and merge topology.
//   - Unallocated: proceeds through validation and DFS.
func (a *Allocator) Allocate(
	ctx context.Context,
	nodeClaim NodeClaim,
	claims []*resourcev1.ResourceClaim,
) (*AllocationResult, error) {
	if len(claims) == 0 {
		return &AllocationResult{
			InstanceTypes: nodeClaim.InstanceTypes(),
			Requirements:  scheduling.NewRequirements(),
		}, nil
	}

	ctx, cancel := context.WithTimeout(ctx, allocateTimeout)
	defer cancel()

	// Phase 1: Classify claims and build effective requirements from already-allocated claims.
	// Effective requirements start from the NodeClaim's base requirements and are progressively
	// tightened as each already-allocated claim contributes topology. Each claim is checked
	// against the effective requirements at the time it is processed, so mutually incompatible
	// claims (e.g., one pinned to zone A and another to zone B) are detected immediately.
	classifyRes, err := a.ClassifyClaims(nodeClaim, claims)
	if err != nil {
		return nil, err
	}
	// If there are no unallocated claims, return early with the requirements from the preallocated claims
	if len(classifyRes.unallocatedClaims) == 0 {
		return &AllocationResult{
			InstanceTypes: nodeClaim.InstanceTypes(),
			Requirements:  classifyRes.requirements,
		}, nil
	}

	// Phase 2: Pool gathering with cache, using the tightened effective requirements.
	var pools []*Pool
	if cached, ok := a.poolCache[nodeClaim.ID()]; ok {
		pools = FilterPools(cached, classifyRes.requirements, nodeClaim.NodeName())
	} else {
		pools = GatherPools(a.inClusterSlices, classifyRes.requirements, nodeClaim.NodeName())
	}

	// Build template devices by instance type.
	// TODO: Ideally, these devices should be tracked globally across the allocator. They currently aren't because we're
	// modeling a subset of the total possible slices - we exclude those from drivers that have already had pools
	// published. Let's consider a new interface to model this (e.g. outstanding drivers / pools?)
	resourceSlices := nodeClaim.ResourceSlices()
	templateDevicesByIT := make(map[InstanceTypeID][]DeviceWithID, len(resourceSlices))
	for itID, slices := range resourceSlices {
		for _, s := range slices {
			for _, d := range s.Devices() {
				templateDevicesByIT[itID] = append(templateDevicesByIT[itID], DeviceWithID{
					Device: d,
					ID: DeviceID{
						DeviceID: cloudprovider.DeviceID{
							Driver: s.Driver(),
							Pool:   s.Pool().Name,
							Device: d.Name,
						},
						Template: true,
					},
				})
			}
		}
	}

	// Create child allocator.
	child := &allocator{
		Allocator:                  a,
		ctx:                        ctx,
		nodeClaim:                  nodeClaim,
		pools:                      pools,
		templateDevicesByIT:        templateDevicesByIT,
		celCache:                   dracel.NewCache(0, dracel.Features{EnableConsumableCapacity: true}),
		allocatedDevices:           sets.New[DeviceID](),
		allocatingCounters:         make(map[PoolKey]map[string]map[string]resourcev1.Counter),
		templateAllocatingCounters: make(map[PoolKey]map[string]map[string]resourcev1.Counter),
		allocatingCapacity:         make(map[DeviceID]map[resourcev1.QualifiedName]resource.Quantity),
		templateAllocatingCapacity: make(map[DeviceID]map[resourcev1.QualifiedName]resource.Quantity),
		deviceMatchesRequest:       make(map[matchKey]bool),
		requirements:               classifyRes.requirements,
	}

	// Validate unallocated claims and build ClaimData. Binding fallback is nil here — it is
	// set per-IT before each DFS run since it depends on the instance type.
	child.claimData = make([]*ClaimData, len(classifyRes.unallocatedClaims))
	for i, claim := range classifyRes.unallocatedClaims {
		cd, err := ValidateClaimRequest(ctx, a.kubeClient, claim, pools, templateDevicesByIT, child.celCache, nil)
		if err != nil {
			return nil, fmt.Errorf("validating claim %q: %w", claim.Name, err)
		}
		child.claimData[i] = cd
	}

	result, err := child.allocate(nodeClaim.InstanceTypes())
	if err != nil {
		return nil, err
	}
	return result, nil
}

type classificationResult struct {
	// unallocatedClaims is the set of claims that haven't already been allocated, either in-cluster or via previous
	// invocations of the allocator.
	unallocatedClaims []*resourcev1.ResourceClaim
	// requirements is the intersection of the NodeClaim's requiremnts and the requirements derived from allocated claims.
	requirements scheduling.Requirements
}

// ClassifyClaims evaluates the set of claims for the NodeClaims. It checks allocation status and ensures compatibility.
// If any of claim is already allocated and incompatible with the NodeClaim, it returns an error. The result is the set
// of unallocated claims and the cumalative requirements derived from the base NodeClaim and the allocated claims.
//
//nolint:gocyclo
func (a *Allocator) ClassifyClaims(nodeClaim NodeClaim, claims []*resourcev1.ResourceClaim) (classificationResult, error) {
	result := classificationResult{
		requirements: copyRequirements(nodeClaim.Requirements()),
	}

	for _, claim := range claims {
		// A claim reserved entirely by pods that are migrating off their nodes is re-allocated rather than treated as
		// committed in place: the device it currently holds has already been freed from the allocated-device seed set
		// (see gatherAllocatedDevices), so the DFS must re-allocate the claim onto the replacement capacity. A claim
		// reserved by a mix of deleting and live pods — or by any non-pod consumer — stays committed below, preserving
		// the live consumers' device and topology.
		if claim.Status.Allocation != nil && a.claimReservedEntirelyByDeletingPods(claim) {
			result.unallocatedClaims = append(result.unallocatedClaims, claim)
			continue
		}
		// In-cluster allocated: status.allocation is set.
		if claim.Status.Allocation != nil {
			reqs := nodeSelectorsToRequirements(claim.Status.Allocation.NodeSelector)
			if reqs != nil {
				if !result.requirements.IsCompatible(*reqs, scheduling.AllowUndefinedWellKnownLabels) {
					return classificationResult{}, serrors.Wrap(fmt.Errorf("in-cluster allocation topology is incompatible with NodeClaim requirements"), "ResourceClaim", klog.KObj(claim))
				}
				result.requirements.Add(reqs.Values()...)
			}
			continue
		}

		// In-memory allocated: a prior pod already allocated this claim in this loop.
		if meta, ok := a.claimAllocationMetadata[resourceClaimID(claim)]; ok {
			// Template-allocated claims are node-local to the original NodeClaim.
			if meta.UsedTemplateDevices {
				if meta.NodeClaimID != nodeClaim.ID() {
					return classificationResult{}, serrors.Wrap(fmt.Errorf("claim is bound to a different in-flight NodeClaim"), "ResourceClaim", klog.KObj(claim))
				}
				// Same NodeClaim — claim is already satisfied, no requirements to add.
			} else {
				// In-cluster only — check topology compatibility.
				if len(meta.TotalRequirements) != 0 {
					if !result.requirements.IsCompatible(meta.TotalRequirements, scheduling.AllowUndefinedWellKnownLabels) {
						return classificationResult{}, serrors.Wrap(fmt.Errorf("in-memory allocation topology is incompatible with NodeClaim requirements"), "ResourceClaim", klog.KObj(claim))
					}
					result.requirements.Add(meta.TotalRequirements.Values()...)
				}

			}
			continue
		}

		// Unallocated — will proceed through DFS.
		result.unallocatedClaims = append(result.unallocatedClaims, claim)
	}
	return result, nil
}

// claimReservedEntirelyByDeletingPods reports whether a ResourceClaim is reserved by a non-empty set of pod consumers
// that are all in the allocator's deletingPodUIDs set. Such a claim's pods are all migrating off their nodes, so the
// claim should be re-allocated rather than treated as committed in place. A claim reserved by a non-pod consumer, by a
// live pod, or by nothing at all returns false (it stays committed). This mirrors the pod-consumer accounting in the
// deviceallocation controller and the allConsumersDeleting predicate used to free devices.
func (a *Allocator) claimReservedEntirelyByDeletingPods(claim *resourcev1.ResourceClaim) bool {
	if len(claim.Status.ReservedFor) == 0 {
		return false
	}
	for i := range claim.Status.ReservedFor {
		ref := &claim.Status.ReservedFor[i]
		if ref.Resource != string(corev1.ResourcePods) || ref.APIGroup != "" {
			return false
		}
		if !a.deletingPodUIDs.Has(ref.UID) {
			return false
		}
	}
	return true
}

// allocator is the per-Allocate() child struct that holds mutable state for the current DFS.
type allocator struct {
	*Allocator
	ctx context.Context
	// celCache caches compiled CEL expressions for device match evaluation. This isn't stored in the top-level allocator
	// due to write-contention on cache misses. The performance tradeoff of an RWMutex should be evaluated.
	celCache *dracel.Cache

	// nodeClaim is the NodeClaim being evaluated during this allocator call
	nodeClaim NodeClaim
	// itID is the current instance type being evaluated in the DFS.
	itID                 InstanceTypeID
	templateDevicesByIT  map[InstanceTypeID][]DeviceWithID
	claimData            []*ClaimData
	deviceMatchesRequest map[matchKey]bool

	// allocatedDevices represents the set of devices that have currently been allocated in the decision tree. Devices
	// are added and removed from this set as we traverse the tree. This contains minimal allocation metadata and is used
	// for quick lookups.
	allocatedDevices sets.Set[DeviceID]
	// allocatedDevicesMetadata contains additional allocation metadata about the device. Currently this only consists of
	// the claim index.
	allocatedDevicesMetadata []deviceAllocationMetadata

	// allocatingCounters tracks counter deductions for in-cluster devices tentatively allocated in the
	// current DFS. Reset per-IT via restoreState(). Map: poolKey → counterSetName → counterName → consumed counter.
	allocatingCounters map[PoolKey]map[string]map[string]resourcev1.Counter
	// templateAllocatingCounters tracks counter deductions for template devices tentatively allocated
	// in the current DFS. Separated from allocatingCounters to prevent cross-contamination when a
	// template pool and an in-cluster pool share the same PoolKey (same driver+pool name).
	templateAllocatingCounters map[PoolKey]map[string]map[string]resourcev1.Counter
	// templateRemainingCounters holds the remaining counter budgets for template pools.
	// Initialized per-IT from ResourceSliceTemplate.SharedCounters. Template counters are local to the
	// current IT (not shared across NodeClaims), so they live on the child allocator rather than the tracker.
	templateRemainingCounters map[PoolKey]map[string]map[string]resourcev1.Counter

	// allocatingCapacity tracks consumed capacity for in-cluster multi-allocatable devices tentatively
	// allocated in the current DFS. Reset per-IT via restoreState(). Map: deviceID → dimensionName → consumed quantity.
	allocatingCapacity map[DeviceID]map[resourcev1.QualifiedName]resource.Quantity
	// templateAllocatingCapacity tracks consumed capacity for template multi-allocatable devices
	// tentatively allocated in the current DFS. Separated from allocatingCapacity to prevent
	// cross-contamination when template and in-cluster pools share the same device names.
	templateAllocatingCapacity map[DeviceID]map[resourcev1.QualifiedName]resource.Quantity

	// requirements are the topology requirements that are incrementally built up by the DFS. Each time an in-cluster
	// device with topology requirements is allocated, those requirements are added to these requirements. These are
	// restored during backtracking from the snapshots.
	requirements scheduling.Requirements
	// pools represents the set of pools that we're currently evaluating. This can be constrained and relaxed as we
	// traverse the decision tree based on allocated device topology requirements.
	pools      []*Pool
	poolsByKey map[PoolKey]*Pool
	// TODO(jmdeal@): Evaluate using the call stack as the stack rather than an explicit stack, I can't recall why I didn't
	snapshots []backtrackSnapshot
}

// matchKey is used to cache CEL selector evaluation results per (device, claim, request) tuple.
type matchKey struct {
	DeviceID     DeviceID
	ClaimIndex   int
	RequestIndex int
}

// backtrackSnapshot captures the incremental requirements and pool set at a point during the DFS,
// enabling restoration on backtrack when a non-node-local device tightens requirements.
type backtrackSnapshot struct {
	reqs  scheduling.Requirements
	pools []*Pool
}

// deviceAllocationMetadata records a single device allocation during the DFS.
type deviceAllocationMetadata struct {
	claimIndex       int
	deviceWithID     DeviceWithID
	consumedCapacity map[resourcev1.QualifiedName]resource.Quantity
}

// allocate runs a per-instance-type DFS over in-cluster and template devices.
// In-cluster devices are iterated first so the DFS naturally prefers them, minimizing
// variance across instance types. Each IT gets a full DFS; ITs whose DFS fails are pruned.
//
//nolint:gocyclo
func (a *allocator) allocate(instanceTypes []InstanceTypeID) (*AllocationResult, error) {
	var survivingITs []InstanceTypeID
	deviceIDsByIT := make(map[InstanceTypeID][]DeviceID)
	// counterConsumptionByIT tracks in-cluster counter deductions per-IT for pessimistic commit.
	counterConsumptionByIT := make(map[InstanceTypeID]map[PoolKey]map[string]map[string]resourcev1.Counter)
	// templateCounterConsumptionByIT tracks template counter deductions per-IT for cross-pod tracking.
	templateCounterConsumptionByIT := make(map[InstanceTypeID]map[PoolKey]map[string]map[string]resourcev1.Counter)
	// capacityConsumptionByIT tracks per-IT consumed capacity for in-cluster multi-allocatable devices.
	capacityConsumptionByIT := make(map[InstanceTypeID]map[DeviceID]map[resourcev1.QualifiedName]resource.Quantity)
	// templateCapacityConsumptionByIT tracks per-IT consumed capacity for template multi-allocatable devices.
	templateCapacityConsumptionByIT := make(map[InstanceTypeID]map[DeviceID]map[resourcev1.QualifiedName]resource.Quantity)
	// templateCounterTotalsByIT tracks template counter budgets computed locally, for deferred init in Commit().
	templateCounterTotalsByIT := make(map[InstanceTypeID]map[PoolKey]map[string]map[string]resourcev1.Counter)

	// Snapshot initial state for restoration between IT attempts.
	initialPools := a.pools

	claimAllocMeta := make([]*ResourceClaimAllocationMetadata, len(a.claimData))
	for i := range claimAllocMeta {
		meta := &ResourceClaimAllocationMetadata{
			NodeClaimID:             a.nodeClaim.ID(),
			ContributedRequirements: make(map[InstanceTypeID]scheduling.Requirements),
			TotalRequirements:       scheduling.NewRequirements(),
			Devices:                 make(map[InstanceTypeID][]DeviceAllocationResult),
		}
		claimAllocMeta[i] = meta
	}

	for _, itID := range instanceTypes {
		select {
		case <-a.ctx.Done():
			return nil, a.ctx.Err()
		default:
		}

		// Restore to initial state.
		a.itID = itID
		a.restoreState(initialPools)

		// Set binding fallback for this IT on all constraints.
		a.setBindingFallback(&AttributeBindingFallback{
			Bindings:       a.attributeBindings,
			NodePool:       a.nodeClaim.NodePoolID().Value(),
			InstanceTypeID: itID,
		})

		// Pre-DFS feasibility: skip this IT if counter budgets can't possibly satisfy demand.
		if !a.countersFeasible() {
			continue
		}

		if a.dfs(0, 0, 0) {
			survivingITs = append(survivingITs, itID)
			counterConsumptionByIT[itID] = a.allocatingCounters
			templateCounterConsumptionByIT[itID] = a.templateAllocatingCounters
			capacityConsumptionByIT[itID] = a.allocatingCapacity
			templateCapacityConsumptionByIT[itID] = a.templateAllocatingCapacity
			if a.templateRemainingCounters != nil && a.allocationTracker.TemplateRemainingForIT(a.nodeClaim.ID(), itID) == nil {
				templateCounterTotalsByIT[itID] = a.templateRemainingCounters
			}
			// Definsive nils to prevents future regressions where allocatingCounters / templateAllocatingCounters is re-used
			a.allocatingCounters = nil
			a.templateAllocatingCounters = nil
			a.allocatingCapacity = nil
			a.templateAllocatingCapacity = nil

			deviceIDsByIT[itID] = make([]DeviceID, len(a.allocatedDevicesMetadata))
			itReqs := scheduling.NewRequirements()
			for di, da := range a.allocatedDevicesMetadata {
				deviceIDsByIT[itID][di] = da.deviceWithID.ID
				meta := claimAllocMeta[da.claimIndex]
				// Update the contributed requirements for the device, each devices contributed requirements are intersected to
				// find the contributed requirements for the instance type.
				if reqs := da.deviceWithID.TopologyRequirements; reqs != nil {
					claimITReqs, ok := meta.ContributedRequirements[itID]
					if !ok {
						claimITReqs = scheduling.NewRequirements()
						meta.ContributedRequirements[itID] = claimITReqs
					}
					for _, req := range *reqs {
						claimITReqs.Add(req)
						itReqs.Add(req)
					}
				}
				if da.deviceWithID.ID.Template {
					meta.UsedTemplateDevices = true
				}
				meta.Devices[itID] = append(meta.Devices[itID], DeviceAllocationResult{
					DeviceID:         da.deviceWithID.ID,
					ConsumedCapacity: da.consumedCapacity,
				})
			}
			// Update the baseline requirements for subsequent instance type simulations based on the contributed requirements
			// from this instance type. This ensures that instance types don't require disjoint requirements to satisfy the same
			// set of claims. This works around a current limitation in the NodeClaim representation - we can't express
			// "instance type A in zone foo OR instance type B in zone bar"
			for _, req := range itReqs {
				a.requirements.Add(req)
			}
		}
	}

	if len(survivingITs) == 0 {
		return nil, fmt.Errorf("no instance type can satisfy the allocation")
	}

	// Compute the total requirements based on the contributed requirements for each instance type
	claimAllocMetaByRC := make(map[ResourceClaimID]*ResourceClaimAllocationMetadata, len(claimAllocMeta))
	nodeClaimRequirements := scheduling.NewRequirements()
	for claimIdx, meta := range claimAllocMeta {
		meta.TotalRequirements = scheduling.NewRequirements()
		for _, itReqs := range meta.ContributedRequirements {
			for _, req := range itReqs {
				meta.TotalRequirements.Add(req)
				// The requirements injected into the NodeClaim should be the set intersection of requirements across all instance
				// types and resource claims
				nodeClaimRequirements.Add(req)
			}
		}
		claimAllocMetaByRC[a.claimData[claimIdx].ID] = meta
	}

	return &AllocationResult{
		InstanceTypes: survivingITs,
		Requirements:  nodeClaimRequirements,
		Allocation: &allocation{
			allocator:                       a.Allocator,
			nodeClaimID:                     a.nodeClaim.ID(),
			deviceIDsByIT:                   deviceIDsByIT,
			filteredPools:                   FilterPools(initialPools, a.requirements, a.nodeClaim.NodeName()),
			claimMetadata:                   claimAllocMetaByRC,
			counterConsumptionByIT:          counterConsumptionByIT,
			templateCounterConsumptionByIT:  templateCounterConsumptionByIT,
			capacityConsumptionByIT:         capacityConsumptionByIT,
			templateCapacityConsumptionByIT: templateCapacityConsumptionByIT,
			templateCounterTotalsByIT:       templateCounterTotalsByIT,
		},
	}, nil
}

// dfs runs the depth-first search over claims, requests, and device slots. Devices are
// iterated lazily from the current pools and template devices rather than from a prebuilt
// candidate list, so pool re-filtering during requirement tightening is automatically
// reflected in subsequent iterations.
func (a *allocator) dfs(claimIdx, reqIdx, slotIdx int) bool {
	select {
	case <-a.ctx.Done():
		return false
	default:
	}

	// Base case: all claims processed.
	if claimIdx >= len(a.claimData) {
		return true
	}

	cd := a.claimData[claimIdx]

	// Advance past completed requests/claims.
	if reqIdx >= len(cd.Requests) {
		return a.dfs(claimIdx+1, 0, 0)
	}
	rd := &cd.Requests[reqIdx]
	numSlots := a.numSlots(rd)
	if slotIdx >= numSlots {
		return a.dfs(claimIdx, reqIdx+1, 0)
	}

	if rd.AllocationMode == resourcev1.DeviceAllocationModeAll {
		return a.dfsAllMode(claimIdx, reqIdx, slotIdx, cd, rd)
	}
	return a.dfsExactCount(claimIdx, reqIdx, slotIdx, cd, rd)
}

// numSlots returns the number of device slots to fill for a request.
func (a *allocator) numSlots(rd *RequestData) int {
	if rd.AllocationMode == resourcev1.DeviceAllocationModeAll {
		return len(rd.AllDevices) + len(rd.AllTemplateDevicesByIT[a.itID])
	}
	return rd.NumDevices
}

// dfsExactCount handles a single slot for an ExactCount request by iterating devices from
// the current pools (in-cluster) and, if enabled, template devices.
func (a *allocator) dfsExactCount(claimIdx, reqIdx, slotIdx int, cd *ClaimData, rd *RequestData) bool {
	// In-cluster devices from pools (reflects current pool state after any requirement tightening).
	for _, pool := range a.pools {
		if pool.Incomplete {
			continue
		}
		exhausted := a.poolCountersExhausted(pool)
		for _, d := range pool.Devices {
			if exhausted && len(d.ConsumesCounters) > 0 {
				continue
			}
			if a.tryDevice(claimIdx, reqIdx, slotIdx, cd, rd, d) {
				return true
			}
		}
	}
	// Template devices for the current instance type.
	for _, d := range a.templateDevicesByIT[a.itID] {
		if a.tryDevice(claimIdx, reqIdx, slotIdx, cd, rd, d) {
			return true
		}
	}
	return false
}

// dfsAllMode handles a single slot for an All-mode request. Each slot maps to a specific
// predetermined device: in-cluster devices first, then template devices.
func (a *allocator) dfsAllMode(claimIdx, reqIdx, slotIdx int, cd *ClaimData, rd *RequestData) bool {
	inClusterCount := len(rd.AllDevices)
	if slotIdx < inClusterCount {
		d := rd.AllDevices[slotIdx]
		return a.tryDevice(claimIdx, reqIdx, slotIdx, cd, rd, d)
	}
	// Template device slot.
	templateIdx := slotIdx - inClusterCount
	templateDevices := rd.AllTemplateDevicesByIT[a.itID]
	if templateIdx < len(templateDevices) {
		d := templateDevices[templateIdx]
		return a.tryDevice(claimIdx, reqIdx, slotIdx, cd, rd, d)
	}
	return false
}

// tryDevice attempts to allocate a single device at the given position in the DFS tree.
// Returns true if the subtree rooted at this device leads to a complete solution.
//
//nolint:gocyclo
func (a *allocator) tryDevice(
	claimIdx, reqIdx, slotIdx int,
	cd *ClaimData,
	rd *RequestData,
	dw DeviceWithID,
) bool {
	deviceID := dw.ID

	// 1. Availability check — multi-alloc devices use capacity as the gatekeeper;
	//    exclusive devices use binary allocation tracking.
	var consumed map[resourcev1.QualifiedName]resource.Quantity
	if dw.AllowMultipleAllocations {
		var ok bool
		consumed, ok = a.checkCapacity(dw.Device, deviceID, rd)
		if !ok {
			return false
		}
	} else {
		if a.allocationTracker.IsAllocated(deviceID, a.nodeClaim, a.itID) {
			return false
		}
		if a.allocatedDevices.Has(deviceID) {
			return false
		}
	}

	// 2. Counter verification — check shared counter budgets.
	if len(dw.ConsumesCounters) > 0 {
		poolKey := PoolKey{Driver: deviceID.Driver, Pool: deviceID.Pool}
		var remainingCounterSets map[string]map[string]resourcev1.Counter
		if deviceID.Template {
			if a.templateRemainingCounters != nil {
				remainingCounterSets = a.templateRemainingCounters[poolKey]
			}
		} else {
			if a.poolsByKey[poolKey] == nil {
				return false
			}
			remainingCounterSets = a.allocationTracker.RemainingCounters[poolKey]
		}
		if !a.checkCounters(dw.Device, poolKey, remainingCounterSets, deviceID.Template) {
			return false
		}
	}

	// 2. Selector match?
	mk := matchKey{DeviceID: deviceID, ClaimIndex: claimIdx, RequestIndex: reqIdx}
	matched, cached := a.deviceMatchesRequest[mk]
	if !cached {
		var err error
		matched, err = DeviceMatchesSelectors(a.ctx, dw.Device, deviceID, rd.Selectors, a.celCache)
		if err != nil {
			return false
		}
		a.deviceMatchesRequest[mk] = matched
	}
	if !matched {
		return false
	}

	// 3. Constraint satisfaction.
	constraintsAdded := 0
	for _, con := range cd.Constraints {
		if !con.Add(rd.Name, dw.Device, deviceID) {
			for j := constraintsAdded - 1; j >= 0; j-- {
				cd.Constraints[j].Remove(rd.Name, dw.Device, deviceID)
			}
			return false
		}
		constraintsAdded++
	}

	// 4. Requirement compatibility (devices with topology requirements only).
	pushedSnapshot := false
	if dw.TopologyRequirements != nil {
		if !a.requirements.IsCompatible(*dw.TopologyRequirements, scheduling.AllowUndefinedWellKnownLabels) {
			for j := constraintsAdded - 1; j >= 0; j-- {
				cd.Constraints[j].Remove(rd.Name, dw.Device, deviceID)
			}
			return false
		}
		// Push snapshot and update.
		a.snapshots = append(a.snapshots, backtrackSnapshot{
			reqs:  copyRequirements(a.requirements),
			pools: a.pools,
		})
		a.requirements.Add(dw.TopologyRequirements.Values()...)
		a.pools = FilterPools(a.pools, a.requirements, a.nodeClaim.NodeName())
		a.buildPoolIndex()
		pushedSnapshot = true
	}

	// Record allocation.
	a.allocatedDevices.Insert(deviceID)
	a.allocatedDevicesMetadata = append(a.allocatedDevicesMetadata, deviceAllocationMetadata{
		claimIndex:       claimIdx,
		deviceWithID:     dw,
		consumedCapacity: consumed,
	})
	if dw.AllowMultipleAllocations {
		// Ensures a multi-allocatable device has a allocating capacity map, even if it has no capacity dimensions.
		// This is needed so that Commit() can identify multi-alloc devices via capacityConsumptionByIT presence.
		allocatingCapacityMap := lo.Ternary(deviceID.Template, a.templateAllocatingCapacity, a.allocatingCapacity)
		if allocatingCapacityMap[deviceID] == nil {
			allocatingCapacityMap[deviceID] = make(map[resourcev1.QualifiedName]resource.Quantity)
		}
	}
	a.deductAllocatingCapacity(consumed, deviceID, deviceID.Template)
	a.deductAllocatingCounters(dw.Device, PoolKey{Driver: deviceID.Driver, Pool: deviceID.Pool}, deviceID.Template)

	// Recurse.
	if a.dfs(claimIdx, reqIdx, slotIdx+1) {
		return true
	}

	// Backtrack — undo in reverse order of application: capacity, counters, allocation, then
	// requirements/pools, then constraints.
	a.restoreAllocatingCapacity(consumed, deviceID, deviceID.Template)
	a.restoreAllocatingCounters(dw.Device, PoolKey{Driver: deviceID.Driver, Pool: deviceID.Pool}, deviceID.Template)
	a.allocatedDevicesMetadata = a.allocatedDevicesMetadata[:len(a.allocatedDevicesMetadata)-1]
	a.allocatedDevices.Delete(deviceID)

	if pushedSnapshot {
		snapshot := a.snapshots[len(a.snapshots)-1]
		a.snapshots = a.snapshots[:len(a.snapshots)-1]
		a.requirements = snapshot.reqs
		a.pools = snapshot.pools
		a.buildPoolIndex()
	}

	for j := constraintsAdded - 1; j >= 0; j-- {
		cd.Constraints[j].Remove(rd.Name, dw.Device, deviceID)
	}

	return false
}

// restoreState resets the child allocator's mutable DFS state for a new IT attempt.
func (a *allocator) restoreState(pools []*Pool) {
	a.allocatedDevicesMetadata = nil
	a.pools = pools
	a.buildPoolIndex()
	a.allocatedDevices = sets.New[DeviceID]()
	a.allocatingCounters = make(map[PoolKey]map[string]map[string]resourcev1.Counter)
	a.templateAllocatingCounters = make(map[PoolKey]map[string]map[string]resourcev1.Counter)
	a.templateRemainingCounters = a.buildTemplateCounters()
	a.allocatingCapacity = make(map[DeviceID]map[resourcev1.QualifiedName]resource.Quantity)
	a.templateAllocatingCapacity = make(map[DeviceID]map[resourcev1.QualifiedName]resource.Quantity)
	a.snapshots = nil
	for _, cd := range a.claimData {
		for _, c := range cd.Constraints {
			c.Reset()
		}
	}
	// NOTE: Requirements are not reset since instance type requirements are accumulated to ensure the result is
	// representable by a NodeClaim.
}

func (a *allocator) buildPoolIndex() {
	a.poolsByKey = make(map[PoolKey]*Pool, len(a.pools))
	for _, p := range a.pools {
		a.poolsByKey[p.Key] = p
	}
}

// buildTemplateCounters returns the remaining counter budgets for the current IT's template pool.
// If a prior pod's Commit() has already initialized the tracker entry, the tracker's reference is
// returned (with prior deductions reflected). Otherwise, totals are computed locally — the tracker
// write is deferred to Commit() to keep Allocate() read-only on the shared tracker.
func (a *allocator) buildTemplateCounters() map[PoolKey]map[string]map[string]resourcev1.Counter {
	if remaining := a.allocationTracker.TemplateRemainingForIT(a.nodeClaim.ID(), a.itID); remaining != nil {
		return remaining
	}
	slices, ok := a.nodeClaim.ResourceSlices()[a.itID]
	if !ok {
		return nil
	}
	totals := computeTemplateTotals(slices)
	if totals == nil {
		return nil
	}
	return totals
}

// computeTemplateTotals extracts the total SharedCounters budget from template slices.
func computeTemplateTotals(slices []ResourceSlice) map[PoolKey]map[string]map[string]resourcev1.Counter {
	var totalsByPool map[PoolKey]map[string]map[string]resourcev1.Counter
	for _, s := range slices {
		sharedCounters := s.SharedCounters()
		if len(sharedCounters) == 0 {
			continue
		}
		poolKey := PoolKey{Driver: s.Driver(), Pool: s.Pool().Name}
		if totalsByPool == nil {
			totalsByPool = make(map[PoolKey]map[string]map[string]resourcev1.Counter)
		}
		counterSets, ok := totalsByPool[poolKey]
		if !ok {
			counterSets = make(map[string]map[string]resourcev1.Counter)
			totalsByPool[poolKey] = counterSets
		}
		for _, cs := range sharedCounters {
			counterSet, ok := counterSets[cs.Name]
			if !ok {
				counterSet = make(map[string]resourcev1.Counter, len(cs.Counters))
				counterSets[cs.Name] = counterSet
			}
			for counterName, counter := range cs.Counters {
				counterSet[counterName] = resourcev1.Counter{Value: counter.Value.DeepCopy()}
			}
		}
	}
	return totalsByPool
}

// copyRequirements creates a shallow copy of a Requirements map.
func copyRequirements(reqs scheduling.Requirements) scheduling.Requirements {
	cp := scheduling.NewRequirements()
	cp.Add(reqs.Values()...)
	return cp
}

// setBindingFallback sets or clears the AttributeBindingFallback on all MatchAttributeConstraints.
func (a *allocator) setBindingFallback(fallback *AttributeBindingFallback) {
	for _, cd := range a.claimData {
		for _, c := range cd.Constraints {
			if mac, ok := c.(*MatchAttributeConstraint); ok {
				mac.AttributeBindingFallback = fallback
			}
		}
	}
}
