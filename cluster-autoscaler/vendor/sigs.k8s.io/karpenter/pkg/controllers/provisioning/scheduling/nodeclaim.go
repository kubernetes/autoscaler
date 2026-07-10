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

package scheduling

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync/atomic"
	"unique"

	"github.com/samber/lo"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/sets"

	v1 "sigs.k8s.io/karpenter/pkg/apis/v1"
	"sigs.k8s.io/karpenter/pkg/cloudprovider"
	opts "sigs.k8s.io/karpenter/pkg/operator/options"
	"sigs.k8s.io/karpenter/pkg/scheduling"
	"sigs.k8s.io/karpenter/pkg/scheduling/dynamicresources"
	"sigs.k8s.io/karpenter/pkg/utils/resources"
)

// NodeClaim is a set of constraints, compatible pods, and possible instance types that could fulfill these constraints. This
// will be turned into one or more actual node instances within the cluster after bin packing.
type NodeClaim struct {
	NodeClaimTemplate

	Pods                 []*corev1.Pod
	reservationManager   *ReservationManager
	topology             *Topology
	daemonOverheadGroups []DaemonOverheadGroup
	hostname             string

	// We store the reserved offerings rather than appending reservation ID labels for two reasons:
	// - We need to release any reservations that were made in previous iterations and are no longer compatible with the
	//   NodeClaim.
	// - Since other NodeClaims may have released reservations which are compatible with this NodeClaim since the last
	//   time a pod was scheduled, it's possible for the set of reserved offerings to expand as well as contract over
	//   multiple iterations. This has the benefit of maximizing the flexibility of an in-flight NodeClaim, maximizing
	//   the scheduler's binpacking efficiency. Tightening the NodeClaim's requirements before finalization would prevent
	//   this expansion.
	reservedOfferings    cloudprovider.Offerings
	reservedOfferingMode ReservedOfferingMode
}

// ReservedOfferingError indicates a NodeClaim couldn't be created or a pod couldn't be added to an exxisting NodeClaim
// due to
type ReservedOfferingError struct {
	error
}

func NewReservedOfferingError(err error) ReservedOfferingError {
	return ReservedOfferingError{error: err}
}

func IsReservedOfferingError(err error) bool {
	roe := &ReservedOfferingError{}
	return errors.As(err, roe)
}

func (e ReservedOfferingError) Unwrap() error {
	return e.error
}

var nodeID int64

func NewNodeClaim(
	nodeClaimTemplate *NodeClaimTemplate,
	topology *Topology,
	daemonOverheadGroups []DaemonOverheadGroup,
	instanceTypes []*cloudprovider.InstanceType,
	reservationManager *ReservationManager,
	reservedOfferingMode ReservedOfferingMode,
) *NodeClaim {
	hostname := fmt.Sprintf("hostname-placeholder-%04d", atomic.AddInt64(&nodeID, 1))
	template := *nodeClaimTemplate
	template.Requirements = scheduling.NewRequirements()
	template.Requirements.Add(nodeClaimTemplate.Requirements.Values()...)
	template.Requirements.Add(scheduling.NewRequirement(corev1.LabelHostname, corev1.NodeSelectorOpIn, hostname))
	template.InstanceTypeOptions = instanceTypes
	template.Spec.Resources.Requests = corev1.ResourceList{}
	// Deep copy host port usage so each NodeClaim can independently track port usage
	groupsForNodeClaim := make([]DaemonOverheadGroup, len(daemonOverheadGroups))
	for i, g := range daemonOverheadGroups {
		groupsForNodeClaim[i] = DaemonOverheadGroup{
			InstanceTypes:  g.InstanceTypes,
			DaemonOverhead: g.DaemonOverhead,
			HostPortUsage:  g.HostPortUsage.DeepCopy(),
		}
	}

	return &NodeClaim{
		NodeClaimTemplate:    template,
		topology:             topology,
		daemonOverheadGroups: groupsForNodeClaim,
		hostname:             hostname,
		reservedOfferings:    cloudprovider.Offerings{},
		reservationManager:   reservationManager,
		reservedOfferingMode: reservedOfferingMode,
	}
}

// CanAdd returns whether the pod can be added to the NodeClaim
// based on the taints/tolerations, host port compatibility,
// requirements, resources, reserved capacity reservations, and topology requirements
func (n *NodeClaim) CanAdd(ctx context.Context, pod *corev1.Pod, podData *PodData, relaxMinValues bool, allocator *dynamicresources.Allocator) (updatedRequirements scheduling.Requirements, updatedInstanceTypes []*cloudprovider.InstanceType, offeringsToReserve []*cloudprovider.Offering, allocationResult *dynamicresources.AllocationResult, err error) {
	// Check Taints
	if err := scheduling.Taints(n.Spec.Taints).ToleratesPod(pod); err != nil {
		return nil, nil, nil, nil, err
	}

	baseRequirements := scheduling.NewRequirements(n.Requirements.Values()...)

	// Check NodeClaim Affinity Requirements
	if err := baseRequirements.Compatible(podData.Requirements, scheduling.AllowUndefinedWellKnownLabels); err != nil {
		return nil, nil, nil, nil, fmt.Errorf("incompatible requirements, %w", err)
	}
	baseRequirements.Add(podData.Requirements.Values()...)

	// Build the list of volume requirement alternatives to try.
	// Each alternative represents one valid combination of topology requirements for the pod's volumes.
	// If there are no volume requirements, use a single nil entry with no additional topology constraint.
	volumeAlternatives := podData.VolumeRequirements
	if len(volumeAlternatives) == 0 {
		volumeAlternatives = []scheduling.Requirements{nil}
	}

	// Try each volume topology alternative. We need to iterate here because the selected
	// volume topology constraints affect downstream topology checks (e.g., pod anti-affinity).
	var lastErr error
	for _, volReqs := range volumeAlternatives {
		reqs, its, ofs, result, err := n.tryVolumeAlternative(ctx, pod, podData, baseRequirements, volReqs, relaxMinValues, allocator)
		if err != nil {
			lastErr = err
			continue
		}
		return reqs, its, ofs, result, nil
	}
	return nil, nil, nil, nil, lastErr
}

// tryVolumeAlternative attempts to add a pod with a specific set of volume requirements,
// checking topology, instance types, and offerings compatibility.
//
//nolint:gocyclo
func (n *NodeClaim) tryVolumeAlternative(ctx context.Context, pod *corev1.Pod, podData *PodData, baseRequirements scheduling.Requirements, volReqs scheduling.Requirements, relaxMinValues bool, allocator *dynamicresources.Allocator) (scheduling.Requirements, []*cloudprovider.InstanceType, []*cloudprovider.Offering, *dynamicresources.AllocationResult, error) {
	nodeClaimRequirements := scheduling.NewRequirements(baseRequirements.Values()...)

	// Add volume requirements to nodeClaimRequirements ONLY (not to pod's affinity).
	// This ensures the NodeClaim satisfies the selected volume topology constraints,
	// while TSC counting uses pod's original affinity.
	if volReqs != nil {
		if err := nodeClaimRequirements.Compatible(volReqs, scheduling.AllowUndefinedWellKnownLabels); err != nil {
			return nil, nil, nil, nil, fmt.Errorf("incompatible volume requirements, %w", err)
		}
		nodeClaimRequirements.Add(volReqs.Values()...)
	}

	// Simulate DRA device allocation before instance-type filtering so the topology requirements contributed by the
	// allocated devices tighten the NodeClaim's requirements and feed the full filtering pipeline.
	var allocationResult *dynamicresources.AllocationResult
	if podData.HasResourceClaimRequests && allocator != nil {
		if podData.ResourceClaimErr != nil {
			return nil, nil, nil, nil, podData.ResourceClaimErr
		}
		result, err := allocator.Allocate(ctx, &draNodeClaim{nc: n}, podData.ResourceClaims)
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf("allocating dynamic resources, %w", err)
		}
		// Merge the allocation's topology requirements into the NodeClaim's requirements (intersection-based narrowing).
		if err := nodeClaimRequirements.Compatible(result.Requirements, scheduling.AllowUndefinedWellKnownLabels); err != nil {
			return nil, nil, nil, nil, fmt.Errorf("incompatible dynamic resource requirements, %w", err)
		}
		nodeClaimRequirements.Add(result.Requirements.Values()...)
		allocationResult = result
	}

	// Check Topology Requirements
	// NOTE: podData.StrictRequirements does NOT include volume requirements, ensuring TSC counting uses pod's original
	// affinity.
	// NOTE: Topology requirements should come last since they can result in a single domain from a set of compatible
	// domains. This can result in unnecessary failures from subsequent checks that narrow requirements.
	topologyRequirements, err := n.topology.AddRequirements(pod, n.Spec.Taints, podData.StrictRequirements, nodeClaimRequirements, scheduling.AllowUndefinedWellKnownLabels)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	if err = nodeClaimRequirements.Compatible(topologyRequirements, scheduling.AllowUndefinedWellKnownLabels); err != nil {
		return nil, nil, nil, nil, err
	}
	nodeClaimRequirements.Add(topologyRequirements.Values()...)

	// Check instance type combinations
	requests := resources.Merge(n.Spec.Resources.Requests, podData.Requests)

	remaining, unsatisfiableKeys, err := filterInstanceTypesByRequirements(n.InstanceTypeOptions, nodeClaimRequirements, pod, podData.Requests, n.daemonOverheadGroups, requests, relaxMinValues)
	if relaxMinValues {
		// Update min values on the requirements if they are relaxed
		for key, minValues := range unsatisfiableKeys {
			nodeClaimRequirements.Get(key).MinValues = new(minValues)
		}
	}
	if err != nil {
		// We avoid wrapping this err because calling String() on InstanceTypeFilterError is an expensive operation
		// due to calls to resources.Merge and stringifying the nodeClaimRequirements
		return nil, nil, nil, nil, err
	}
	// Apply the DRA-specific instance type filter: only instance types whose device allocation succeeded survive.
	if allocationResult != nil {
		supported := sets.New(lo.Map(allocationResult.InstanceTypes, func(it dynamicresources.InstanceTypeID, _ int) string {
			return it.Value()
		})...)
		remaining = lo.Filter(remaining, func(it *cloudprovider.InstanceType, _ int) bool {
			return supported.Has(it.Name)
		})
		if len(remaining) == 0 {
			return nil, nil, nil, nil, fmt.Errorf("no instance type satisfies both scheduling and dynamic resource requirements")
		}
	}
	ofs, err := n.offeringsToReserve(ctx, remaining, nodeClaimRequirements)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	return nodeClaimRequirements, remaining, ofs, allocationResult, nil
}

// Add updates the NodeClaim to schedule the pod to this NodeClaim, updating
// the NodeClaim with new requirements, instance types, and offerings to reserve
// based on the pod scheduling
func (n *NodeClaim) Add(ctx context.Context, pod *corev1.Pod, podData *PodData, nodeClaimRequirements scheduling.Requirements, instanceTypes []*cloudprovider.InstanceType, offeringsToReserve []*cloudprovider.Offering, allocationResult *dynamicresources.AllocationResult, allocator *dynamicresources.Allocator) {
	// Update node
	n.Pods = append(n.Pods, pod)
	n.InstanceTypeOptions = instanceTypes
	// Daemon overhead is excluded here to avoid double-counting
	n.Spec.Resources.Requests = resources.Merge(n.Spec.Resources.Requests, podData.Requests)
	n.Requirements = nodeClaimRequirements
	n.topology.Register(corev1.LabelHostname, n.hostname)
	n.topology.Record(pod, n.Spec.Taints, nodeClaimRequirements, scheduling.AllowUndefinedWellKnownLabels)
	hostPorts := scheduling.GetHostPorts(pod)
	for _, group := range n.daemonOverheadGroups {
		group.HostPortUsage.Add(pod, hostPorts)
	}
	n.reservationManager.Reserve(n.hostname, offeringsToReserve...)
	n.releaseReservedOfferings(n.reservedOfferings, offeringsToReserve)
	n.reservedOfferings = offeringsToReserve

	// Commit the DRA device allocation now that the placement decision is finalized, then release the allocator's
	// reservations for any instance types that the allocator simulated but were pruned from the NodeClaim's final
	// candidate set (e.g. by offering/fit filters). This frees those instance types' devices for other NodeClaims.
	// The Allocation handle is nil when there were no new device allocations to commit (e.g. every claim was already
	// allocated in-cluster), in which case there is nothing to commit or release.
	if allocationResult != nil && allocationResult.Allocation != nil {
		allocationResult.Allocation.Commit(ctx)
		committed := sets.New(lo.Map(instanceTypes, func(it *cloudprovider.InstanceType, _ int) string { return it.Name })...)
		var pruned []dynamicresources.InstanceTypeID
		for _, it := range allocationResult.InstanceTypes {
			if !committed.Has(it.Value()) {
				pruned = append(pruned, it)
			}
		}
		if len(pruned) > 0 {
			allocator.ReleaseInstanceType(ctx, unique.Make(n.hostname), pruned...)
		}
	}
}

// releaseReservedOfferings releases all offerings which are present in the current reserved offerings, but are not
// present in the updated reserved offerings.
func (n *NodeClaim) releaseReservedOfferings(current, updated cloudprovider.Offerings) {
	updatedIDs := sets.New[string]()
	for _, o := range updated {
		updatedIDs.Insert(o.ReservationID())
	}
	for _, o := range current {
		if !updatedIDs.Has(o.ReservationID()) {
			n.reservationManager.Release(n.hostname, o)
		}
	}
}

// reserveOfferings handles the reservation of `karpenter.sh/capacity-type: reserved` offerings, returning the set of
// reserved offerings. If the ReservedOfferingMode is set to strict, this function may also return an error if it failed
// to reserve compatible offerings when some were available.
//
//nolint:gocyclo
func (n *NodeClaim) offeringsToReserve(
	ctx context.Context,
	instanceTypes []*cloudprovider.InstanceType,
	nodeClaimRequirements scheduling.Requirements,
) (cloudprovider.Offerings, error) {
	if !opts.FromContext(ctx).FeatureGates.ReservedCapacity {
		return nil, nil
	}

	hasCompatibleOffering := false
	var reservedOfferings cloudprovider.Offerings
	for _, it := range instanceTypes {
		for _, o := range it.Offerings {
			if o.CapacityType() != v1.CapacityTypeReserved || !o.Available {
				continue
			}
			// Track every incompatible reserved offering for release. Since releasing a reservation is a no-op when there is no
			// reservation for the given host, there's no need to check that a reservation actually exists for the offering.
			if !nodeClaimRequirements.IsCompatible(o.Requirements, scheduling.AllowUndefinedWellKnownLabels) {
				continue
			}
			hasCompatibleOffering = true
			// Note that reservation is an idempotent operation - if we have previously successfully reserved an offering for
			// this host, this operation is guaranteed to succeed. We may also succeed to make reservations for offerings which
			// failed in previous iterations if other NodeClaims have released them since the last attempt.
			if n.reservationManager.CanReserve(n.hostname, o) {
				reservedOfferings = append(reservedOfferings, o)
			}
		}
	}

	if n.reservedOfferingMode == ReservedOfferingModeStrict {
		// If an instance type with a compatible reserved offering exists, but we failed to make any reservations, we should
		// fail. This could occur when all of the capacity for compatible instances has been reserved by previously created
		// nodeclaims. Since we reserve offering pessimistically, i.e. we will reserve any offering that the instance could
		// be launched with, we should fall back and attempt to schedule this pod in a subsequent scheduling simulation once
		// reservation capacity is available again.
		if hasCompatibleOffering && len(reservedOfferings) == 0 {
			return nil, NewReservedOfferingError(fmt.Errorf("one or more instance types with compatible reserved offerings are available, but could not be reserved"))
		}
		// If the nodeclaim previously had compatible reserved offerings, but the additional requirements filtered those out,
		// we should fail to add the pod to this nodeclaim.
		if len(n.reservedOfferings) != 0 && len(reservedOfferings) == 0 {
			return nil, NewReservedOfferingError(fmt.Errorf("satisfying updated nodeclaim constraints would remove all compatible reserved offering options"))
		}
	}
	return reservedOfferings, nil
}

// Add min daemon resource requests (min across instance types)
func (n *NodeClaim) addDaemonRequests() {
	remaining := sets.New(n.InstanceTypeOptions...)
	var minDaemonOverhead corev1.ResourceList
	for _, g := range n.daemonOverheadGroups {
		// Only consider groups that have at least one instance type still remaining
		hasRemaining := false
		for _, it := range g.InstanceTypes {
			if remaining.Has(it) {
				hasRemaining = true
				break
			}
		}
		if !hasRemaining {
			continue
		}
		if len(minDaemonOverhead) == 0 {
			minDaemonOverhead = g.DaemonOverhead
		} else {
			minDaemonOverhead = resources.MinResources(minDaemonOverhead, g.DaemonOverhead)
		}
	}
	if len(minDaemonOverhead) > 0 {
		n.Spec.Resources.Requests = resources.Merge(n.Spec.Resources.Requests, minDaemonOverhead)
	}
}

// FinalizeScheduling is called once all scheduling has completed and allows the node to perform any cleanup
// necessary before its requirements are used for instance launching. drivers is the set of DRA driver names whose
// devices were allocated to pods scheduled to this NodeClaim; when non-empty it is recorded as an annotation for the
// initialization controller to gate on.
func (n *NodeClaim) FinalizeScheduling(drivers ...string) {
	// We need nodes to have hostnames for topology purposes, but we don't want to pass that node name on to consumers
	// of the node as it will be displayed in error messages
	delete(n.Requirements, corev1.LabelHostname)
	if len(drivers) != 0 {
		slices.Sort(drivers)
		n.Annotations = lo.Assign(n.Annotations, map[string]string{
			v1.DRADriversAnnotationKey: strings.Join(drivers, ","),
		})
	}
	// If there are any reserved offerings tracked, inject those requirements onto the NodeClaim. This ensures that if
	// there are multiple reserved offerings for an instance type, we don't attempt to overlaunch into a single offering.
	if len(n.reservedOfferings) != 0 {
		// Tightening constraint to reserved ensures that we get automatic drift handling when the Node / NodeClaim's capacity
		// type label is dynamically updated by the cloudprovider.
		n.Requirements[v1.CapacityTypeLabelKey] = scheduling.NewRequirement(v1.CapacityTypeLabelKey, corev1.NodeSelectorOpIn, v1.CapacityTypeReserved)
		n.Requirements.Add(scheduling.NewRequirement(
			cloudprovider.ReservationIDLabel,
			corev1.NodeSelectorOpIn,
			lo.Map(n.reservedOfferings, func(o *cloudprovider.Offering, _ int) string { return o.ReservationID() })...,
		))
	}
	// Add min Daemon resource requests
	// Daemon resource requests are excluded during bin packing to avoid double-counting
	// Adding during bin-packing would add it every time a pod is added
	n.addDaemonRequests()
}

func (n *NodeClaim) RemoveInstanceTypeOptionsByPriceAndMinValues(reqs scheduling.Requirements, maxPrice float64) (*NodeClaim, error) {
	n.InstanceTypeOptions = lo.Filter(n.InstanceTypeOptions, func(it *cloudprovider.InstanceType, _ int) bool {
		launchPrice := it.Offerings.Available().WorstLaunchPrice(reqs)
		return launchPrice < maxPrice
	})
	if _, _, err := n.InstanceTypeOptions.SatisfiesMinValues(reqs); err != nil {
		return nil, err
	}
	return n, nil
}

func InstanceTypeList(instanceTypeOptions []*cloudprovider.InstanceType) string {
	var itSb strings.Builder
	for i, it := range instanceTypeOptions {
		// print the first 5 instance types only (indices 0-4)
		if i > 4 {
			fmt.Fprintf(&itSb, " and %d other(s)", len(instanceTypeOptions)-i)
			break
		} else if i > 0 {
			fmt.Fprint(&itSb, ", ")
		}
		fmt.Fprint(&itSb, it.Name)
	}
	return itSb.String()
}

type InstanceTypeFilterError struct {
	// Each of these three flags indicates if that particular criteria was met by at least one instance type
	requirementsMet bool
	fits            bool
	hasOffering     bool

	// requirementsAndFits indicates if a single instance type met the scheduling requirements and had enough resources
	requirementsAndFits bool
	// requirementsAndOffering indicates if a single instance type met the scheduling requirements and was a required offering
	requirementsAndOffering bool
	// fitsAndOffering indicates if a single instance type had enough resources and was a required offering
	fitsAndOffering          bool
	minValuesIncompatibleErr error

	// We capture requirements so that we can know what the requirements were when evaluating instance type compatibility
	requirements scheduling.Requirements
	// We capture podRequests here since when a pod can't schedule due to requests, it's because the pod
	// was on its own on the simulated Node and exceeded the available resources for any instance type for this NodePool
	podRequests corev1.ResourceList
	// We capture the min and max daemonRequests since this contributes to the resources that are required to schedule to this NodePool. The min/max DaemonRequests are the min / max across all instance types.
	minDaemonRequests corev1.ResourceList
	maxDaemonRequests corev1.ResourceList
}

// Updates the min/max daemon overhead bounds for error reporting.
func (e *InstanceTypeFilterError) trackDaemonOverhead(dr corev1.ResourceList) {
	if len(e.maxDaemonRequests) == 0 {
		e.minDaemonRequests = dr
		e.maxDaemonRequests = dr
		return
	}
	e.minDaemonRequests = resources.MinResources(e.minDaemonRequests, dr)
	e.maxDaemonRequests = resources.MaxResources(e.maxDaemonRequests, dr)
}

// Returns total resource requirements as a formatted string, showing min/max bounds when they differ.
func (e InstanceTypeFilterError) resourcesString() string {
	if len(e.minDaemonRequests) == 0 {
		return resources.String(e.podRequests)
	}
	minTotal := resources.Merge(e.minDaemonRequests, e.podRequests)
	maxTotal := resources.Merge(e.maxDaemonRequests, e.podRequests)
	minStr := resources.String(minTotal)
	maxStr := resources.String(maxTotal)
	if minStr == maxStr {
		return minStr
	}
	return fmt.Sprintf("min=%s, max=%s", minStr, maxStr)
}

//nolint:gocyclo
func (e InstanceTypeFilterError) Error() string {
	// minValues is specified in the requirements and is not met
	if e.minValuesIncompatibleErr != nil {
		return fmt.Sprintf("%s, requirements=%s, resources=%s", e.minValuesIncompatibleErr.Error(), e.requirements, e.resourcesString())
	}
	// no instance type met any of the three criteria, meaning each criteria was enough to completely prevent
	// this pod from scheduling
	if !e.requirementsMet && !e.fits && !e.hasOffering {
		return fmt.Sprintf("no instance type met the scheduling requirements or had enough resources or had a required offering, requirements=%s, resources=%s", e.requirements, e.resourcesString())
	}
	// check the other pairwise criteria
	if !e.requirementsMet && !e.fits {
		return fmt.Sprintf("no instance type met the scheduling requirements or had enough resources, requirements=%s, resources=%s", e.requirements, e.resourcesString())
	}
	if !e.requirementsMet && !e.hasOffering {
		return fmt.Sprintf("no instance type met the scheduling requirements or had a required offering, requirements=%s, resources=%s", e.requirements, e.resourcesString())
	}
	if !e.fits && !e.hasOffering {
		return fmt.Sprintf("no instance type had enough resources or had a required offering, requirements=%s, resources=%s", e.requirements, e.resourcesString())
	}
	// and then each individual criteria. These are sort of the same as above in that each one indicates that no
	// instance type matched that criteria at all, so it was enough to exclude all instance types.  I think it's
	// helpful to have these separate, since we can report the multiple excluding criteria above.
	if !e.requirementsMet {
		return fmt.Sprintf("no instance type met all requirements, requirements=%s, resources=%s", e.requirements, e.resourcesString())
	}
	if !e.fits {
		msg := fmt.Sprintf("no instance type has enough resources, requirements=%s, resources=%s", e.requirements, e.resourcesString())
		// special case for a user typo I saw reported once
		if e.podRequests.Cpu().Cmp(resource.MustParse("1M")) >= 0 {
			msg += " (CPU request >= 1 Million, m vs M typo?)"
		}
		return msg
	}
	if !e.hasOffering {
		return fmt.Sprintf("no instance type has the required offering, requirements=%s, resources=%s", e.requirements, e.resourcesString())
	}
	// see if any pair of criteria was enough to exclude all instances
	if e.requirementsAndFits {
		return fmt.Sprintf("no instance type which met the scheduling requirements and had enough resources, had a required offering, requirements=%s, resources=%s", e.requirements, e.resourcesString())
	}
	if e.fitsAndOffering {
		return fmt.Sprintf("no instance type which had enough resources and the required offering met the scheduling requirements, requirements=%s, resources=%s", e.requirements, e.resourcesString())
	}
	if e.requirementsAndOffering {
		return fmt.Sprintf("no instance type which met the scheduling requirements and the required offering had the required resources, requirements=%s, resources=%s", e.requirements, e.resourcesString())
	}
	// finally all instances were filtered out, but we had at least one instance that met each criteria, and met each
	// pairwise set of criteria, so the only thing that remains is no instance which met all three criteria simultaneously
	return fmt.Sprintf("no instance type met the requirements/resources/offering tuple, requirements=%s, resources=%s", e.requirements, e.resourcesString())
}

//nolint:gocyclo
func filterInstanceTypesByRequirements(instanceTypes []*cloudprovider.InstanceType, requirements scheduling.Requirements, pod *corev1.Pod, podRequests corev1.ResourceList, daemonOverheadGroups []DaemonOverheadGroup, totalRequests corev1.ResourceList, relaxMinValues bool) (cloudprovider.InstanceTypes, map[string]int, error) {
	unsatisfiableKeys := map[string]int{}
	// We hold the results of our scheduling simulation inside of this InstanceTypeFilterError struct
	// to reduce the CPU load of having to generate the error string for a failed scheduling simulation
	err := InstanceTypeFilterError{
		requirementsMet: false,
		fits:            false,
		hasOffering:     false,

		requirementsAndFits:     false,
		requirementsAndOffering: false,
		fitsAndOffering:         false,

		requirements: requirements,
		podRequests:  podRequests,
	}
	remaining := cloudprovider.InstanceTypes{}
	// exposed host ports on the pod
	hostPorts := scheduling.GetHostPorts(pod)
	eligibleInstanceTypes := sets.New(instanceTypes...)

	for _, group := range daemonOverheadGroups {
		if portUsageErr := group.HostPortUsage.Conflicts(pod, hostPorts); portUsageErr != nil {
			continue
		}
		var totalRequestsForInstanceType corev1.ResourceList
		if len(group.DaemonOverhead) != 0 {
			err.trackDaemonOverhead(group.DaemonOverhead)
			totalRequestsForInstanceType = resources.MergeInto(totalRequestsForInstanceType, totalRequests)
			totalRequestsForInstanceType = resources.MergeInto(totalRequestsForInstanceType, group.DaemonOverhead)
		} else {
			totalRequestsForInstanceType = totalRequests
		}

		for _, it := range group.InstanceTypes {
			if !eligibleInstanceTypes.Has(it) {
				continue
			}
			// the tradeoff to not short-circuiting on the filtering is that we can report much better error messages
			// about why scheduling failed
			itCompat := compatible(it, requirements)
			itFits, itHasOffering := fits(it, totalRequestsForInstanceType, requirements)

			// track if any single instance type met a single criteria
			err.requirementsMet = err.requirementsMet || itCompat
			err.fits = err.fits || itFits
			err.hasOffering = err.hasOffering || itHasOffering

			// track if any single instance type met the three pairs of criteria
			err.requirementsAndFits = err.requirementsAndFits || (itCompat && itFits && !itHasOffering)
			err.requirementsAndOffering = err.requirementsAndOffering || (itCompat && itHasOffering && !itFits)
			err.fitsAndOffering = err.fitsAndOffering || (itFits && itHasOffering && !itCompat)

			// and if it met all criteria, we keep the instance type and continue filtering.  We now won't be reporting
			// any errors.
			if itCompat && itFits && itHasOffering {
				remaining = append(remaining, it)
			}
		}
	}

	if requirements.HasMinValues() {
		// We don't care about the minimum number of instance types that meet our requirements here, we only care if they meet our requirements.
		_, unsatisfiableKeys, err.minValuesIncompatibleErr = remaining.SatisfiesMinValues(requirements)
		if err.minValuesIncompatibleErr != nil {
			if !relaxMinValues {
				// If MinValuesPolicy is set to Strict, return empty InstanceTypeOptions as we cannot launch with the remaining InstanceTypes when min values is violated.
				remaining = nil
			} else {
				err.minValuesIncompatibleErr = nil
			}
		}
	}
	if len(remaining) == 0 {
		return nil, unsatisfiableKeys, err
	}
	return remaining, unsatisfiableKeys, nil
}

func compatible(instanceType *cloudprovider.InstanceType, requirements scheduling.Requirements) bool {
	return instanceType.Requirements.Intersects(requirements) == nil
}

func fits(instanceType *cloudprovider.InstanceType, requests corev1.ResourceList, requirements scheduling.Requirements) (itFits bool, hasOffering bool) {
	for _, group := range instanceType.AllocatableOfferingsList() {
		resourceFit := resources.Fits(requests, group.Allocatable)
		for _, of := range group.Offerings {
			if requirements.IsCompatible(of.Requirements, scheduling.AllowUndefinedWellKnownLabels) {
				hasOffering = true
				if resourceFit {
					return true, true
				}
				break
			}
		}
	}
	return false, hasOffering
}
