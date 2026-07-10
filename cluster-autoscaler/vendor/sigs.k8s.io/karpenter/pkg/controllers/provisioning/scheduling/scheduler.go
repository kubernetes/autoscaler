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
	"bytes"
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/awslabs/operatorpkg/option"
	"github.com/awslabs/operatorpkg/serrors"
	"github.com/samber/lo"
	"go.uber.org/multierr"
	corev1 "k8s.io/api/core/v1"
	resourcev1 "k8s.io/api/resource/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/klog/v2"
	"k8s.io/utils/clock"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	autoscalingv1alpha1 "sigs.k8s.io/karpenter/pkg/apis/autoscaling/v1alpha1"
	v1 "sigs.k8s.io/karpenter/pkg/apis/v1"
	"sigs.k8s.io/karpenter/pkg/cloudprovider"
	"sigs.k8s.io/karpenter/pkg/controllers/state"
	"sigs.k8s.io/karpenter/pkg/events"
	"sigs.k8s.io/karpenter/pkg/metrics"
	"sigs.k8s.io/karpenter/pkg/operator/injection"
	karpopts "sigs.k8s.io/karpenter/pkg/operator/options"
	"sigs.k8s.io/karpenter/pkg/scheduling"
	"sigs.k8s.io/karpenter/pkg/scheduling/dynamicresources"
	"sigs.k8s.io/karpenter/pkg/utils/disruption"
	"sigs.k8s.io/karpenter/pkg/utils/pod"
	"sigs.k8s.io/karpenter/pkg/utils/resources"
)

type ReservedOfferingMode int

// TODO: Evaluate if another mode should be created for drift. The problem with strict is that it assumes we can run
// multiple scheduling loops to make progress, but if scheduling all pods from the drifted node in a single iteration
// requires fallback, we're at a stalemate. This makes strict a non-starter for drift IMO.
// On the other hand, fallback will result in non-ideal launches when there's constrained capacity. This should be
// rectified by consolidation, but if we can be "right" the at initial launch that would be preferable.
// One potential improvement is a "preferences" type strategy, where we attempt to schedule the pod without fallback
// first. This is an improvement over the current fallback strategy since it ensures all new nodeclaims are attempted,
// before then attempting all nodepools, but it still doesn't address the case when offerings are reserved pessimistically.
// I don't believe there's a solution to this short of the max-flow based instance selection algorithm, which has its own
// drawbacks.
const (
	// ReservedOfferingModeFallbackAlways indicates to the scheduler that the addition of a pod to a nodeclaim which
	// results in all potential reserved offerings being filtered out is allowed (e.g. on-demand / spot fallback).
	ReservedOfferingModeFallback ReservedOfferingMode = iota
	// ReservedOfferingModeStrict indicates that the scheduler should fail to add a pod to a nodeclaim if doing so would
	// prevent it from scheduling to reserved capacity, when it would have otherwise.
	ReservedOfferingModeStrict
)

type PreferencePolicy int

const (
	// PreferencePolicyRespect indicates to the scheduler that it should attempt to respect all preference requirements
	// and topologies. The scheduler will treat all preferences as required at first and then will slowly relax
	// these requirements one at a time until it is able to schedule the pod
	PreferencePolicyRespect PreferencePolicy = iota
	// PreferencePolicyIgnore indicates to the scheduler that it should ignore all preference requirements and
	// topologies. Preferences include preferredDuringSchedulingIgnoredDuringExecution affinities and ScheduleAnyways
	// topologySpreadConstraints
	PreferencePolicyIgnore
)

type options struct {
	reservedOfferingMode    ReservedOfferingMode
	preferencePolicy        PreferencePolicy
	minValuesPolicy         karpopts.MinValuesPolicy
	numConcurrentReconciles int
	enforceConsolidateAfter bool
}

type Options = option.Function[options]

var DisableReservedCapacityFallback = func(opts *options) {
	opts.reservedOfferingMode = ReservedOfferingModeStrict
}

var IgnorePreferences = func(opts *options) {
	opts.preferencePolicy = PreferencePolicyIgnore
}

var NumConcurrentReconciles = func(numConcurrentReconciles int) func(*options) {
	return func(opts *options) {
		opts.numConcurrentReconciles = numConcurrentReconciles
	}
}

var MinValuesPolicy = func(policy karpopts.MinValuesPolicy) func(*options) {
	return func(opts *options) {
		opts.minValuesPolicy = policy
	}
}

var IsConsolidationSimulation = func(opts *options) {
	opts.enforceConsolidateAfter = true
}

func NewScheduler(
	ctx context.Context,
	kubeClient client.Client,
	nodePools []*v1.NodePool,
	cluster *state.Cluster,
	stateNodes []*state.StateNode,
	topology *Topology,
	instanceTypes map[string][]*cloudprovider.InstanceType,
	daemonSetPods []*corev1.Pod,
	recorder events.Recorder,
	clock clock.Clock,
	volumeReqsByPod map[types.UID][]scheduling.Requirements,
	allocator *dynamicresources.Allocator,
	opts ...Options,
) *Scheduler {
	minValuesPolicy := option.Resolve(opts...).minValuesPolicy

	// if any of the nodePools add a taint with a prefer no schedule effect, we add a toleration for the taint
	// during preference relaxation
	toleratePreferNoSchedule := false
	for _, np := range nodePools {
		for _, taint := range np.Spec.Template.Spec.Taints {
			if taint.Effect == corev1.TaintEffectPreferNoSchedule {
				toleratePreferNoSchedule = true
			}
		}
	}
	// Pre-filter instance types eligible for NodePools to reduce work done during scheduling loops for pods
	// if no templates remain, we still want to build the scheduler so that Karpenter can ack pods which can schedule to existing and in-flight capacity
	templates := lo.FilterMap(nodePools, func(np *v1.NodePool, _ int) (*NodeClaimTemplate, bool) {
		var err error
		nct := NewNodeClaimTemplate(np)
		nct.InstanceTypeOptions, _, err = filterInstanceTypesByRequirements(instanceTypes[np.Name], nct.Requirements, &corev1.Pod{}, corev1.ResourceList{}, []DaemonOverheadGroup{{InstanceTypes: instanceTypes[np.Name], HostPortUsage: scheduling.NewHostPortUsage()}}, corev1.ResourceList{}, minValuesPolicy == karpopts.MinValuesPolicyBestEffort)
		if len(nct.InstanceTypeOptions) == 0 {
			if instanceTypeFilterErr, ok := lo.ErrorsAs[InstanceTypeFilterError](err); ok && instanceTypeFilterErr.minValuesIncompatibleErr != nil {
				recorder.Publish(NoCompatibleInstanceTypes(np, true))
				log.FromContext(ctx).WithValues("NodePool", klog.KObj(np)).Info("skipping, nodepool requirements filtered out all instance types", "minValuesIncompatibleErr", instanceTypeFilterErr.minValuesIncompatibleErr)
			} else {
				recorder.Publish(NoCompatibleInstanceTypes(np, false))
				log.FromContext(ctx).WithValues("NodePool", klog.KObj(np)).Info("skipping, nodepool requirements filtered out all instance types")
			}
			return nil, false
		}
		return nct, true
	})
	s := &Scheduler{
		uuid:                 uuid.NewUUID(),
		kubeClient:           kubeClient,
		nodeClaimTemplates:   templates,
		topology:             topology,
		cluster:              cluster,
		daemonOverheadGroups: buildDaemonOverheadGroups(ctx, templates, daemonSetPods),
		cachedPodData:        map[types.UID]*PodData{}, // cache pod data to avoid having to continually recompute it
		volumeReqsByPod:      volumeReqsByPod,          // Volume requirements per pod
		recorder:             recorder,
		preferences:          &Preferences{ToleratePreferNoSchedule: toleratePreferNoSchedule},
		remainingResources: lo.SliceToMap(nodePools, func(np *v1.NodePool) (string, corev1.ResourceList) {
			return np.Name, corev1.ResourceList(np.Spec.Limits)
		}),
		clock:                   clock,
		reservationManager:      NewReservationManager(instanceTypes),
		reservedOfferingMode:    option.Resolve(opts...).reservedOfferingMode,
		preferencePolicy:        option.Resolve(opts...).preferencePolicy,
		minValuesPolicy:         minValuesPolicy,
		numConcurrentReconciles: lo.Ternary(option.Resolve(opts...).numConcurrentReconciles > 0, option.Resolve(opts...).numConcurrentReconciles, 1),
		allocator:               allocator,
		instanceTypes:           instanceTypes,
		cachedResourceClaims:    map[types.NamespacedName]*resourcev1.ResourceClaim{},
	}

	npByName := lo.SliceToMap(nodePools, func(np *v1.NodePool) (string, *v1.NodePool) {
		return np.Name, np
	})

	nodeToNodePool := lo.SliceToMap(stateNodes, func(n *state.StateNode) (string, *v1.NodePool) {
		return n.Name(), npByName[n.Labels()[v1.NodePoolLabelKey]]
	})
	// Build a set of node names that are marked for deletion so we can exempt their pods
	// from the consolidateAfter destination check
	deletingNodeNames := sets.New[string]()
	for n := range cluster.Nodes() {
		if n.MarkedForDeletion() {
			deletingNodeNames.Insert(n.Name())
		}
	}
	s.deletingNodeNames = deletingNodeNames
	s.calculateExistingNodeClaims(ctx, stateNodes, daemonSetPods, nodeToNodePool, option.Resolve(opts...).enforceConsolidateAfter)
	return s
}

type PodData struct {
	Requests                 corev1.ResourceList
	Requirements             scheduling.Requirements
	StrictRequirements       scheduling.Requirements
	HasResourceClaimRequests bool
	VolumeRequirements       []scheduling.Requirements // Volume topology requirement alternatives

	// ResourceClaims are the resolved ResourceClaim objects referenced by the pod, populated for DRA pods when the
	// allocator is enabled. ResourceClaimErr records a resolution failure (e.g. a claim that hasn't been created yet),
	// in which case the pod is deferred to a subsequent scheduling loop.
	ResourceClaims   []*resourcev1.ResourceClaim
	ResourceClaimErr error
}

type Scheduler struct {
	uuid                    types.UID // Unique UUID attached to this scheduling loop
	newNodeClaims           []*NodeClaim
	existingNodes           []*ExistingNode
	nodeClaimTemplates      []*NodeClaimTemplate
	remainingResources      map[string]corev1.ResourceList // (NodePool name) -> remaining resources for that NodePool
	daemonOverheadGroups    map[*NodeClaimTemplate][]DaemonOverheadGroup
	cachedPodData           map[types.UID]*PodData                  // (Pod Namespace/Name) -> pre-computed data for pods to avoid re-computation and memory usage
	volumeReqsByPod         map[types.UID][]scheduling.Requirements // Volume topology requirement alternatives per pod
	preferences             *Preferences
	topology                *Topology
	cluster                 *state.Cluster
	recorder                events.Recorder
	kubeClient              client.Client
	clock                   clock.Clock
	reservationManager      *ReservationManager
	reservedOfferingMode    ReservedOfferingMode
	preferencePolicy        PreferencePolicy
	minValuesPolicy         karpopts.MinValuesPolicy
	numConcurrentReconciles int
	deletingNodeNames       sets.Set[string]

	// allocator simulates DRA device allocation for pods with ResourceClaims. It is nil when DRA support is disabled.
	allocator *dynamicresources.Allocator
	// instanceTypes is the per-NodePool instance type set, used to resolve template devices for existing nodes.
	instanceTypes map[string][]*cloudprovider.InstanceType
	// cachedResourceClaims memoizes ResourceClaim lookups for the duration of a single scheduling loop.
	cachedResourceClaims map[types.NamespacedName]*resourcev1.ResourceClaim
}

// DRAError indicates a pod will not be attempted to be scheduled because it has Dynamic Resource Allocation requirements
// that are not yet supported by Karpenter
type DRAError struct {
	error
}

func NewDRAError(err error) DRAError {
	return DRAError{error: err}
}

func IsDRAError(err error) bool {
	draErr := &DRAError{}
	return errors.As(err, draErr)
}

func (e DRAError) Unwrap() error {
	return e.error
}

// Results contains the results of the scheduling operation
type Results struct {
	NewNodeClaims              []*NodeClaim
	ExistingNodes              []*ExistingNode
	PodErrors                  map[*corev1.Pod]error
	DRAClaimAllocationMetadata map[types.NamespacedName]*dynamicresources.ResourceClaimAllocationMetadata
}

// Record sends eventing and log messages back for the results that were produced from a scheduling run
// It also nominates nodes in the cluster state based on the scheduling run to signal to other components
// leveraging the cluster state that a previous scheduling run that was recorded is relying on these nodes
func (r Results) Record(ctx context.Context, recorder events.Recorder, cluster *state.Cluster) {
	// Report failures and nominations
	for p, err := range r.PodErrors {
		if IsReservedOfferingError(err) {
			continue
		}
		if IsDRAError(err) {
			recorder.Publish(PodFailedToScheduleEvent(p, err))
			log.FromContext(ctx).WithValues("Pod", klog.KObj(p)).Info("skipping pod with Dynamic Resource Allocation requirements, not yet supported by Karpenter")
			continue
		}
		log.FromContext(ctx).WithValues("Pod", klog.KObj(p)).Error(err, "could not schedule pod")
		recorder.Publish(PodFailedToScheduleEvent(p, err))
	}
	// Nominate nodes only for real pods. Virtual buffer pods (injected by
	// GetPendingPods for CapacityBuffer) must NOT trigger nomination because:
	//   1. Nomination blocks ALL disruption (drift, expiry) — we only want to
	//      block emptiness, which is handled by cluster.HasBufferPods instead.
	//   2. Buffer pods are re-injected every pass, so nomination would never
	//      expire, making buffer nodes permanently undisruptable.
	for _, existing := range r.ExistingNodes {
		realPods := lo.Filter(existing.Pods, func(p *corev1.Pod, _ int) bool {
			return !isVirtualBufferPod(p)
		})
		if len(realPods) > 0 {
			cluster.NominateNodeForPod(ctx, existing.ProviderID())
		}
		for _, p := range realPods {
			recorder.Publish(NominatePodEvent(p, existing.Node, existing.NodeClaim))
		}
	}
	// Report new nodes, or exit to avoid log spam
	newCount := 0
	for _, nodeClaim := range r.NewNodeClaims {
		newCount += len(nodeClaim.Pods)
	}
	if newCount == 0 {
		return
	}
	log.FromContext(ctx).WithValues("nodeclaims", len(r.NewNodeClaims), "pods", newCount).Info("computed new nodeclaim(s) to fit pod(s)")
	// Report in flight newNodes, or exit to avoid log spam
	inflightCount := 0
	existingCount := 0
	for _, node := range lo.Filter(r.ExistingNodes, func(node *ExistingNode, _ int) bool { return len(node.Pods) > 0 }) {
		inflightCount++
		existingCount += len(node.Pods)
	}
	if existingCount == 0 {
		return
	}
	log.FromContext(ctx).WithValues("nodes", inflightCount, "pods", existingCount).Info("computed unready node(s) will fit pod(s)")
}

func isVirtualBufferPod(p *corev1.Pod) bool {
	return p.Annotations[autoscalingv1alpha1.FakePodAnnotationKey] == autoscalingv1alpha1.FakePodAnnotationValue
}

func (r Results) ReservedOfferingErrors() map[*corev1.Pod]error {
	return lo.PickBy(r.PodErrors, func(_ *corev1.Pod, err error) bool {
		return IsReservedOfferingError(err)
	})
}

func (r Results) DRAErrors() map[*corev1.Pod]error {
	return lo.PickBy(r.PodErrors, func(_ *corev1.Pod, err error) bool {
		return IsDRAError(err)
	})
}

func (r Results) NodePoolToPodMapping() map[string][]*corev1.Pod {
	result := make(map[string][]*corev1.Pod)

	for _, nc := range r.NewNodeClaims {
		nodePoolName := nc.Labels[v1.NodePoolLabelKey]
		result[nodePoolName] = append(result[nodePoolName], nc.Pods...)
	}

	for _, nc := range r.ExistingNodes {
		nodePoolName := nc.Labels()[v1.NodePoolLabelKey]
		result[nodePoolName] = append(result[nodePoolName], nc.Pods...)
	}

	return result
}

func (r Results) ExistingNodeToPodMapping() map[string][]*corev1.Pod {
	return lo.SliceToMap(lo.Filter(r.ExistingNodes, func(n *ExistingNode, _ int) bool {
		// Filter out nodes that are not managed
		return n.Managed()
	}), func(n *ExistingNode) (string, []*corev1.Pod) {
		return n.NodeClaim.Name, n.Pods
	})
}

// AllNonPendingPodsScheduled returns true if all pods scheduled.
// We don't care if a pod was pending before consolidation and will still be pending after. It may be a pod that we can't
// schedule at all and don't want it to block consolidation.
func (r Results) AllNonPendingPodsScheduled() bool {
	return len(lo.OmitBy(r.PodErrors, func(p *corev1.Pod, err error) bool {
		return pod.IsProvisionable(p)
	})) == 0
}

// NonPendingPodSchedulingErrors creates a string that describes why pods wouldn't schedule that is suitable for presentation
func (r Results) NonPendingPodSchedulingErrors() string {
	errs := lo.OmitBy(r.PodErrors, func(p *corev1.Pod, err error) bool {
		return pod.IsProvisionable(p)
	})
	if len(errs) == 0 {
		return "No Pod Scheduling Errors"
	}
	var msg bytes.Buffer
	fmt.Fprintf(&msg, "not all pods would schedule, ")
	const MaxErrors = 5
	numErrors := 0
	for k, err := range errs {
		fmt.Fprintf(&msg, "%s/%s => %s ", k.Namespace, k.Name, err)
		numErrors++
		if numErrors >= MaxErrors {
			fmt.Fprintf(&msg, " and %d other(s)", len(errs)-MaxErrors)
			break
		}
	}
	return msg.String()
}

// TruncateInstanceTypes filters the result based on the maximum number of instanceTypes that needs
// to be considered. This filters all instance types generated in NewNodeClaims in the Results
func (r Results) TruncateInstanceTypes(ctx context.Context, maxInstanceTypes int) Results {
	var validNewNodeClaims []*NodeClaim
	for _, newNodeClaim := range r.NewNodeClaims {
		// The InstanceTypeOptions are truncated due to limitations in sending the number of instances to launch API.
		var err error
		newNodeClaim.InstanceTypeOptions, err = newNodeClaim.InstanceTypeOptions.Truncate(ctx, newNodeClaim.Requirements, maxInstanceTypes)
		if err != nil {
			// Check if the truncated InstanceTypeOptions in each NewNodeClaim from the results still satisfy the minimum requirements
			// If number of InstanceTypes in the NodeClaim cannot satisfy the minimum requirements, add its Pods to error map with reason.
			for _, pod := range newNodeClaim.Pods {
				r.PodErrors[pod] = serrors.Wrap(fmt.Errorf("pod didn’t schedule because NodePool couldn’t meet minValues requirements, %w", err), "NodePool", klog.KRef("", newNodeClaim.NodePoolName))
			}
		} else {
			validNewNodeClaims = append(validNewNodeClaims, newNodeClaim)
		}
	}
	r.NewNodeClaims = validNewNodeClaims
	return r
}

//nolint:gocyclo
func (s *Scheduler) Solve(ctx context.Context, pods []*corev1.Pod) (Results, error) {
	defer metrics.Measure(DurationSeconds, map[string]string{ControllerLabel: injection.GetControllerName(ctx)})()
	// We loop trying to schedule unschedulable pods as long as we are making progress.  This solves a few
	// issues including pods with affinity to another pod in the batch. We could topo-sort to solve this, but it wouldn't
	// solve the problem of scheduling pods where a particular order is needed to prevent a max-skew violation. E.g. if we
	// had 5xA pods and 5xB pods were they have a zonal topology spread, but A can only go in one zone and B in another.
	// We need to schedule them alternating, A, B, A, B, .... and this solution also solves that as well.
	podErrors := map[*corev1.Pod]error{}
	// Reset the metric for the controller, so we don't keep old ids around
	UnschedulablePodsCount.DeletePartialMatch(map[string]string{ControllerLabel: injection.GetControllerName(ctx)})
	PendingPodsByEffectiveZone.DeletePartialMatch(map[string]string{ControllerLabel: injection.GetControllerName(ctx)})
	QueueDepth.DeletePartialMatch(map[string]string{ControllerLabel: injection.GetControllerName(ctx)})
	podCountByZone := make(map[string]int)
	for _, p := range pods {
		s.updateCachedPodData(ctx, p)
		if p.Status.Phase == corev1.PodPending {
			zone := s.computeEffectiveZoneFromPod(p)
			podCountByZone[zone]++
		}
	}

	q := NewQueue(pods, s.cachedPodData)

	startTime := s.clock.Now()
	for {
		UnfinishedWorkSeconds.Set(s.clock.Since(startTime).Seconds(), map[string]string{ControllerLabel: injection.GetControllerName(ctx), schedulingIDLabel: string(s.uuid)})
		QueueDepth.Set(float64(len(q.pods)), map[string]string{ControllerLabel: injection.GetControllerName(ctx), schedulingIDLabel: string(s.uuid)})

		// Try the next pod
		pod, ok := q.Pop()
		if !ok {
			break
		}
		// We relax the pod all the way the first time we see it
		// If we don't schedule it, we store the original pod (with preferences)
		// in the queue and give ourselves another chance to schedule it later
		if err := s.trySchedule(ctx, pod.DeepCopy()); err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				log.FromContext(ctx).V(1).WithValues("duration", s.clock.Since(startTime).Truncate(time.Second), "scheduling-id", string(s.uuid)).Info("scheduling simulation timed out")
				break
			}
			podErrors[pod] = err
			if e := s.topology.Update(ctx, pod); e != nil && !errors.Is(e, context.DeadlineExceeded) {
				log.FromContext(ctx).Error(e, "failed updating topology")
			}
			// Update the cached podData since the pod was relaxed, and it could have changed its requirement set
			s.updateCachedPodData(ctx, pod)
			q.Push(pod)
		} else {
			delete(podErrors, pod)
		}
	}
	UnfinishedWorkSeconds.Delete(map[string]string{ControllerLabel: injection.GetControllerName(ctx), schedulingIDLabel: string(s.uuid)})
	for _, m := range s.newNodeClaims {
		m.FinalizeScheduling(s.draDriversForNodeClaim(m)...)
	}

	controllerName := injection.GetControllerName(ctx)
	for zone, count := range podCountByZone {
		PendingPodsByEffectiveZone.Set(float64(count), map[string]string{
			ControllerLabel: controllerName,
			"zone":          zone,
		})
	}

	results := Results{
		NewNodeClaims: s.newNodeClaims,
		ExistingNodes: s.existingNodes,
		PodErrors:     podErrors,
	}
	if s.allocator != nil {
		results.DRAClaimAllocationMetadata = lo.MapKeys(
			s.allocator.ResourceClaimAllocationMetadata(),
			func(_ *dynamicresources.ResourceClaimAllocationMetadata, k dynamicresources.ResourceClaimID) types.NamespacedName {
				return k.Value()
			},
		)
	}
	return results, ctx.Err()
}

func (s *Scheduler) trySchedule(ctx context.Context, p *corev1.Pod) error {
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		err := s.add(ctx, p)
		if err == nil {
			return nil
		}
		// We should only relax the pod's requirements when the error is not a reserved offering error because the pod may be
		// able to schedule later without relaxing constraints. This could occur in this scheduling run, if other NodeClaims
		// release the required reservations when constrained, or in subsequent runs. For an example, reference the following
		// test: "shouldn't relax preferences when a pod fails to schedule due to a reserved offering error".
		if IsReservedOfferingError(err) {
			return err
		}
		// DRA errors are permanent while the IgnoreDRARequests flag is enabled, so we shouldn't attempt to relax
		// pod requirements as we don't want to schedule the pod.
		if IsDRAError(err) {
			return err
		}
		// Eventually we won't be able to relax anymore and this while loop will exit
		if relaxed := s.preferences.Relax(ctx, p); !relaxed {
			return err
		}
		if e := s.topology.Update(ctx, p); e != nil && !errors.Is(e, context.DeadlineExceeded) {
			log.FromContext(ctx).Error(e, "failed updating topology")
		}
		// Update the cached podData since the pod was relaxed, and it could have changed its requirement set
		s.updateCachedPodData(ctx, p)
	}
}

func (s *Scheduler) updateCachedPodData(ctx context.Context, p *corev1.Pod) {
	var requirements scheduling.Requirements
	if s.preferencePolicy == PreferencePolicyIgnore {
		requirements = scheduling.NewStrictPodRequirements(p)
	} else {
		requirements = scheduling.NewPodRequirements(p)
	}
	strictRequirements := requirements
	if scheduling.HasPreferredNodeAffinity(p) {
		// strictPodRequirements is important as it ensures we don't inadvertently restrict the possible pod domains by a
		// preferred node affinity.  Only required node affinities can actually reduce pod domains.
		strictRequirements = scheduling.NewStrictPodRequirements(p)
	}
	data := &PodData{
		Requests:                 resources.RequestsForPods(p),
		Requirements:             requirements,
		StrictRequirements:       strictRequirements,
		HasResourceClaimRequests: pod.HasDRARequirements(p),
		VolumeRequirements:       s.volumeReqsByPod[p.UID], // Volume requirements
	}
	// Resolve the pod's ResourceClaims once, in the sequential path, so the parallel candidate evaluation can reuse them
	// without per-candidate API lookups. A resolution failure is recorded and surfaced as a scheduling error in add().
	if data.HasResourceClaimRequests && s.allocator != nil {
		data.ResourceClaims, data.ResourceClaimErr = s.resolvePodClaims(ctx, p)
	}
	s.cachedPodData[p.UID] = data
}

func (s *Scheduler) add(ctx context.Context, pod *corev1.Pod) error {
	// Check if pod has DRA requirements - if so, return DRA error when IgnoreDRARequests is enabled
	if s.cachedPodData[pod.UID].HasResourceClaimRequests && karpopts.FromContext(ctx).IgnoreDRARequests {
		return NewDRAError(fmt.Errorf("pod has Dynamic Resource Allocation requirements that are not yet supported by Karpenter"))
	}
	// If the pod's ResourceClaims couldn't be resolved (e.g. a referenced claim hasn't been created yet), no candidate
	// can satisfy it. Surface the error directly so the pod is deferred and retried once the claim exists.
	if err := s.cachedPodData[pod.UID].ResourceClaimErr; err != nil {
		return err
	}

	// first try to schedule against an in-flight real node
	if err := s.addToExistingNode(ctx, pod); err == nil {
		return nil
	}
	// Consider using https://pkg.go.dev/container/heap
	sort.Slice(s.newNodeClaims, func(a, b int) bool { return len(s.newNodeClaims[a].Pods) < len(s.newNodeClaims[b].Pods) })

	// Pick existing node that we are about to create
	if err := s.addToInflightNode(ctx, pod); err == nil {
		return nil
	}
	if len(s.nodeClaimTemplates) == 0 {
		return fmt.Errorf("nodepool requirements filtered out all available instance types")
	}
	err := s.addToNewNodeClaim(ctx, pod)
	if err == nil {
		return nil
	}
	return err
}

func (s *Scheduler) addToExistingNode(ctx context.Context, p *corev1.Pod) error {
	idx := math.MaxInt
	var mu sync.Mutex

	var existingNode *ExistingNode
	var requirements scheduling.Requirements
	var allocationResult *dynamicresources.AllocationResult

	// determine the volumes that will be mounted if the pod schedules
	volumes, err := scheduling.GetVolumes(ctx, s.kubeClient, p)
	if err != nil {
		return err
	}
	parallelizeUntil(s.numConcurrentReconciles, len(s.existingNodes), func(i int) bool {
		if s.existingNodes[i].isUnderConsolidateAfter && (!pod.IsPending(p) && !s.deletingNodeNames.Has(p.Spec.NodeName)) {
			// We shouldn't try to schedule candidate pods onto nodes that are under consolidate after.
			// Pending pods and pods from deleting nodes are exempt.
			return true
		}
		r, result, err := s.existingNodes[i].CanAdd(ctx, p, s.cachedPodData[p.UID], volumes, s.allocator)
		if err == nil {
			mu.Lock()
			defer mu.Unlock()

			// Ensure that we always take an earlier successful schedule to keep consistent ordering
			if i >= idx {
				return false
			}
			existingNode = s.existingNodes[i]
			requirements = r
			allocationResult = result
			idx = i
			return false
		}
		return true
	})
	// If we set the existingNode to something valid, this means that we successfully scheduled to one of these nodes
	if existingNode != nil {
		existingNode.Add(ctx, p, s.cachedPodData[p.UID], requirements, volumes, allocationResult)
		return nil
	}
	return fmt.Errorf("failed scheduling pod to existing nodes")
}

func (s *Scheduler) addToInflightNode(ctx context.Context, pod *corev1.Pod) error {
	idx := math.MaxInt
	var mu sync.Mutex

	var inflightNodeClaim *NodeClaim
	var updatedRequirements scheduling.Requirements
	var updatedInstanceTypes []*cloudprovider.InstanceType
	var offeringsToReserve []*cloudprovider.Offering
	var allocationResult *dynamicresources.AllocationResult
	parallelizeUntil(s.numConcurrentReconciles, len(s.newNodeClaims), func(i int) bool {
		r, its, ofr, result, err := s.newNodeClaims[i].CanAdd(ctx, pod, s.cachedPodData[pod.UID], false, s.allocator)
		if err == nil {
			mu.Lock()
			defer mu.Unlock()

			// Ensure that we always take an earlier successful schedule to keep consistent ordering
			if i >= idx {
				return false
			}
			inflightNodeClaim = s.newNodeClaims[i]
			updatedRequirements = r
			updatedInstanceTypes = its
			offeringsToReserve = ofr
			allocationResult = result
			idx = i
			return false
		}
		return true
	})
	if inflightNodeClaim != nil {
		inflightNodeClaim.Add(ctx, pod, s.cachedPodData[pod.UID], updatedRequirements, updatedInstanceTypes, offeringsToReserve, allocationResult, s.allocator)
		return nil
	}
	return fmt.Errorf("failed scheduling pod to inflight nodes")
}

//nolint:gocyclo
func (s *Scheduler) addToNewNodeClaim(ctx context.Context, pod *corev1.Pod) error {
	idx := math.MaxInt
	var mu sync.Mutex

	var newNodeClaim *NodeClaim
	var updatedRequirements scheduling.Requirements
	var updatedInstanceTypes []*cloudprovider.InstanceType
	var offeringsToReserve []*cloudprovider.Offering
	var allocationResult *dynamicresources.AllocationResult

	errs := make([]error, len(s.nodeClaimTemplates))
	parallelizeUntil(s.numConcurrentReconciles, len(s.nodeClaimTemplates), func(i int) bool {
		its := s.nodeClaimTemplates[i].InstanceTypeOptions
		// if limits have been applied to the nodepool, ensure we filter instance types to avoid violating those limits
		if remaining, ok := s.remainingResources[s.nodeClaimTemplates[i].NodePoolName]; ok {
			// Node limits can be enforced early, since we know exactly how much capacity in nodes will be consumed by any instance type (1 node).
			nodesRemaining, ok := remaining[resources.Node]
			if ok && nodesRemaining.IsZero() {
				errs[i] = serrors.Wrap(fmt.Errorf("node limits have been exhausted for nodepool"), "NodePool", klog.KRef("", s.nodeClaimTemplates[i].NodePoolName))
				return true
			}
			its = filterByRemainingResources(its, remaining)
			if len(its) == 0 {
				errs[i] = serrors.Wrap(fmt.Errorf("all available instance types exceed limits for nodepool"), "NodePool", klog.KRef("", s.nodeClaimTemplates[i].NodePoolName))
				return true
			} else if len(s.nodeClaimTemplates[i].InstanceTypeOptions) != len(its) {
				log.FromContext(ctx).V(1).WithValues(
					"NodePool", klog.KRef("", s.nodeClaimTemplates[i].NodePoolName),
				).Info("instance types were excluded because they would breach limits",
					"excluded", len(s.nodeClaimTemplates[i].InstanceTypeOptions)-len(its),
					"total", len(s.nodeClaimTemplates[i].InstanceTypeOptions))
			}
		}
		nodeClaim := NewNodeClaim(s.nodeClaimTemplates[i], s.topology, s.daemonOverheadGroups[s.nodeClaimTemplates[i]], its, s.reservationManager, s.reservedOfferingMode)
		r, its, ofs, result, err := nodeClaim.CanAdd(ctx, pod, s.cachedPodData[pod.UID], s.minValuesPolicy == karpopts.MinValuesPolicyBestEffort, s.allocator)
		if err != nil {
			errs[i] = err

			// If the pod is compatible with a NodePool with reserved offerings available, we shouldn't fall back to a NodePool
			// with a lower weight. We could consider allowing "fallback" to NodePools with equal weight if they also have
			// reserved capacity in the future if scheduling latency becomes an issue.
			if IsReservedOfferingError(err) {
				mu.Lock()
				defer mu.Unlock()

				// A reserved offering error means that any subsequent successful after this NodeClaimTemplate isn't valid
				if i >= idx {
					return false
				}
				newNodeClaim = nil
				updatedRequirements = nil
				updatedInstanceTypes = nil
				offeringsToReserve = nil
				allocationResult = nil
				idx = i
				return false
			}
			return true
		}
		mu.Lock()
		defer mu.Unlock()

		// Ensure that we always take an earlier successful schedule to keep consistent ordering
		// We care about this particularly with NewNodeClaims because NodeClaims should be evaluated by weight
		if i >= idx {
			return false
		}

		_, minValuesRelaxed := lo.Find(nodeClaim.Requirements.Keys().UnsortedList(), func(k string) bool {
			updated := r.Get(k).MinValues
			original := nodeClaim.Requirements.Get(k).MinValues
			return original != nil && updated != nil && lo.FromPtr(updated) < lo.FromPtr(original)
		})
		if minValuesRelaxed {
			nodeClaim.Annotations[v1.NodeClaimMinValuesRelaxedAnnotationKey] = "true"
		} else {
			nodeClaim.Annotations[v1.NodeClaimMinValuesRelaxedAnnotationKey] = "false"
		}

		newNodeClaim = nodeClaim
		updatedRequirements = r
		updatedInstanceTypes = its
		offeringsToReserve = ofs
		allocationResult = result
		idx = i
		return false
	})
	if newNodeClaim != nil {
		// we will launch this nodeClaim and need to track its maximum possible resource usage against our remaining resources
		newNodeClaim.Add(ctx, pod, s.cachedPodData[pod.UID], updatedRequirements, updatedInstanceTypes, offeringsToReserve, allocationResult, s.allocator)
		s.newNodeClaims = append(s.newNodeClaims, newNodeClaim)
		s.remainingResources[newNodeClaim.NodePoolName] = subtractMax(s.remainingResources[newNodeClaim.NodePoolName], newNodeClaim.InstanceTypeOptions)
		return nil
	}
	return multierr.Combine(errs...)
}

func (s *Scheduler) calculateExistingNodeClaims(ctx context.Context, stateNodes []*state.StateNode, daemonSetPods []*corev1.Pod, nodePoolMap map[string]*v1.NodePool, enforceConsolidateAfter bool) {
	// create our existing nodes
	for _, node := range stateNodes {
		taints := node.Taints()
		daemons := s.getCompatibleDaemonPods(ctx, node, taints, daemonSetPods)
		isUnderConsolidateAfter := enforceConsolidateAfter && disruption.IsUnderConsolidateAfter(nodePoolMap[node.Name()], node.NodeClaim, s.clock)
		s.existingNodes = append(s.existingNodes, NewExistingNode(node, s.topology, taints, resources.RequestsForPods(daemons...), s.instanceTypeForNode(node), isUnderConsolidateAfter))
		s.updateRemainingResources(node)
	}
	s.sortExistingNodes()
}

// getCompatibleDaemonPods filters daemon pods that can schedule to the given node
func (s *Scheduler) getCompatibleDaemonPods(ctx context.Context, node *state.StateNode, taints []corev1.Taint, daemonSetPods []*corev1.Pod) []*corev1.Pod {
	var daemons []*corev1.Pod
	for _, p := range daemonSetPods {
		if s.shouldSkipDaemonPod(ctx, p) {
			continue
		}
		if s.isDaemonPodCompatibleWithNode(p, taints, node.Labels()) {
			daemons = append(daemons, p)
		}
	}
	return daemons
}

// shouldSkipDaemonPod checks if a daemon pod should be skipped due to DRA requirements
func (s *Scheduler) shouldSkipDaemonPod(ctx context.Context, p *corev1.Pod) bool {
	return pod.HasDRARequirements(p) && karpopts.FromContext(ctx).IgnoreDRARequests
}

// isDaemonPodCompatibleWithNode checks if a daemon pod is compatible with the node
func (s *Scheduler) isDaemonPodCompatibleWithNode(p *corev1.Pod, taints []corev1.Taint, nodeLabels map[string]string) bool {
	if err := scheduling.Taints(taints).ToleratesPod(p); err != nil {
		return false
	}
	if err := scheduling.NewLabelRequirements(nodeLabels).Compatible(scheduling.NewStrictPodRequirements(p)); err != nil {
		return false
	}
	return true
}

// updateRemainingResources updates the remaining resources for the node's nodepool
func (s *Scheduler) updateRemainingResources(node *state.StateNode) {
	// We don't use the status field and instead recompute the remaining resources to ensure we have a consistent view
	// of the cluster during scheduling.  Depending on how node creation falls out, this will also work for cases where
	// we don't create NodeClaim resources.
	if _, ok := s.remainingResources[node.Labels()[v1.NodePoolLabelKey]]; ok {
		s.remainingResources[node.Labels()[v1.NodePoolLabelKey]] = resources.Subtract(s.remainingResources[node.Labels()[v1.NodePoolLabelKey]], node.Capacity())
	}
}

// sortExistingNodes sorts existing nodes with initialized nodes first
func (s *Scheduler) sortExistingNodes() {
	// Order the existing nodes for scheduling with initialized nodes first
	// This is done specifically for consolidation where we want to make sure we schedule to initialized nodes
	// before we attempt to schedule uninitialized ones
	sort.SliceStable(s.existingNodes, func(i, j int) bool {
		if s.existingNodes[i].Initialized() && !s.existingNodes[j].Initialized() {
			return true
		}
		if !s.existingNodes[i].Initialized() && s.existingNodes[j].Initialized() {
			return false
		}
		return s.existingNodes[i].Name() < s.existingNodes[j].Name()
	})
}

// computeEffectiveZoneFromPod calculates the effective zone constraint by intersecting
// pod-level zone signals, PVC volume zones, and TSC valid domains. This can be the
// specific zone name if exactly one zone, "flexible" if multiple zones, "none" if no intersection.
//
//nolint:gocyclo
func (s *Scheduler) computeEffectiveZoneFromPod(pod *corev1.Pod) string {
	podData := s.cachedPodData[pod.UID]
	tscZoneValidDomains, satisfiable := s.topology.GetTopologyZoneConstraints(pod, podData.Requirements)
	if !satisfiable {
		return "none"
	}

	zoneReq := podData.StrictRequirements.Get(corev1.LabelTopologyZone)
	volZoneReq := volumeZoneReq(podData.VolumeRequirements)

	var zonalValues []string
	if zoneReq.Operator() == corev1.NodeSelectorOpIn {
		zonalValues = zoneReq.Values()
	} else if volZoneReq != nil {
		zonalValues = volZoneReq.Values()
	} else if len(tscZoneValidDomains) > 0 {
		zonalValues = sets.List(tscZoneValidDomains)
	} else {
		return "flexible"
	}

	var matchCount int
	var matchedZone string
	for _, zone := range zonalValues {
		if !zoneReq.Has(zone) {
			continue
		}
		if volZoneReq != nil && !volZoneReq.Has(zone) {
			continue
		}
		if len(tscZoneValidDomains) > 0 && !tscZoneValidDomains.Has(zone) {
			continue
		}
		matchCount++
		if matchCount == 1 {
			matchedZone = zone
		} else {
			return "flexible"
		}
	}
	return lo.Ternary(matchCount == 1, matchedZone, "none")
}

// volumeZoneReq returns a single Requirement representing the union of zone constraints
// across all volume alternatives. Returns nil if volumes don't constrain zones.
func volumeZoneReq(volumeReqs []scheduling.Requirements) *scheduling.Requirement {
	if len(volumeReqs) == 0 {
		return nil
	}
	var merged *scheduling.Requirement
	for _, vol := range volumeReqs {
		if vol == nil {
			return nil
		}
		req := vol.Get(corev1.LabelTopologyZone)
		if req.Operator() != corev1.NodeSelectorOpIn {
			return nil
		}
		if len(volumeReqs) == 1 {
			return req
		}
		if merged == nil {
			merged = scheduling.NewRequirement(corev1.LabelTopologyZone, corev1.NodeSelectorOpIn, req.Values()...)
		} else {
			merged.Insert(req.Values()...)
		}
	}
	return merged
}

// parallelizeUntil is an implementation of workqueue.ParallelizeUntil that modifies the
// doWorkPiece so that a worker always finishes its work when it pulls a piece off of pieces
// The function returns a bool that represents whether the worker should continue doing work
// or whether the worker should finish
func parallelizeUntil(workers, pieces int, doWorkPiece func(int) bool) {
	toProcess := make(chan int, pieces)
	for i := range pieces {
		toProcess <- i
	}
	close(toProcess)
	if pieces < workers {
		workers = pieces
	}
	wg := sync.WaitGroup{}
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for work := range toProcess {
				if !doWorkPiece(work) {
					return
				}
			}
		}()
	}
	wg.Wait()
}

type DaemonOverheadGroup struct {
	InstanceTypes  []*cloudprovider.InstanceType
	DaemonOverhead corev1.ResourceList
	HostPortUsage  *scheduling.HostPortUsage
}

// buildDaemonOverheadGroups groups instance types by their compatible daemon pods and computes the following for NodeClaimTemplate and group
// - Overhead required for daemons to schedule for any node provisioned by the NodeClaimTemplate
// - Requested host ports for DaemonSet pods
func buildDaemonOverheadGroups(ctx context.Context, nodeClaimTemplates []*NodeClaimTemplate, daemonSetPods []*corev1.Pod) map[*NodeClaimTemplate][]DaemonOverheadGroup {
	return lo.SliceToMap(nodeClaimTemplates, func(nct *NodeClaimTemplate) (*NodeClaimTemplate, []DaemonOverheadGroup) {
		groups := map[string]*DaemonOverheadGroup{}
		for _, it := range nct.InstanceTypeOptions {
			compatible := lo.Filter(daemonSetPods, func(p *corev1.Pod, _ int) bool {
				if pod.HasDRARequirements(p) && karpopts.FromContext(ctx).IgnoreDRARequests {
					return false
				}
				return isDaemonPodCompatible(nct, it, p)
			})
			key := podSetKey(compatible)
			if g, ok := groups[key]; ok {
				g.InstanceTypes = append(g.InstanceTypes, it)
			} else {
				var overhead corev1.ResourceList
				if len(compatible) > 0 {
					overhead = resources.RequestsForPods(compatible...)
				}
				hostPortUsage := scheduling.NewHostPortUsage()
				for _, p := range compatible {
					hostPortUsage.Add(p, scheduling.GetHostPorts(p))
				}
				groups[key] = &DaemonOverheadGroup{
					InstanceTypes:  []*cloudprovider.InstanceType{it},
					DaemonOverhead: overhead,
					HostPortUsage:  hostPortUsage,
				}
			}
		}
		result := lo.Map(lo.Values(groups), func(g *DaemonOverheadGroup, _ int) DaemonOverheadGroup { return *g })
		return nct, result
	})
}

// podSetKey creates a deterministic key from a list of pods for grouping.
func podSetKey(pods []*corev1.Pod) string {
	if len(pods) == 0 {
		return ""
	}
	keys := make([]string, len(pods))
	for i, p := range pods {
		keys[i] = client.ObjectKeyFromObject(p).String()
	}
	sort.Strings(keys)
	return strings.Join(keys, ",")
}

// isDaemonPodCompatible determines if the daemon pod is compatible with the NodeClaimTemplate for daemon scheduling
func isDaemonPodCompatible(nodeClaimTemplate *NodeClaimTemplate, it *cloudprovider.InstanceType, pod *corev1.Pod) bool {
	preferences := &Preferences{}
	// Add a toleration for PreferNoSchedule since a daemon pod shouldn't respect the preference
	_ = preferences.toleratePreferNoScheduleTaints(pod)
	if err := scheduling.Taints(nodeClaimTemplate.Spec.Taints).ToleratesPod(pod); err != nil {
		return false
	}
	for {
		podRequirements := scheduling.NewStrictPodRequirements(pod)
		// We don't consider pod preferences for scheduling requirements since we know that pod preferences won't matter with Daemonset scheduling
		if nodeClaimTemplate.Requirements.IsCompatible(podRequirements, scheduling.AllowUndefinedWellKnownLabels) &&
			// We use Intersects instead of IsCompatible for instance type requirements since we want to ignore any custom keys on the daemonset pod since they
			// will not be available on the instance type requirements.
			it.Requirements.Intersects(podRequirements) == nil {
			return true
		}
		// If relaxing the Node Affinity term didn't succeed, then this DaemonSet can't schedule to this NodePool
		// We don't consider other forms of relaxation here since we don't consider pod affinities/anti-affinities
		// when considering DaemonSet schedulability
		if preferences.removeRequiredNodeAffinityTerm(pod) == nil {
			return false
		}
	}
}

// subtractMax returns the remaining resources after subtracting the max resource quantity per instance type. To avoid
// overshooting out, we need to pessimistically assume that if e.g. we request a 2, 4 or 8 CPU instance type
// that the 8 CPU instance type is all that will be available.  This could cause a batch of pods to take multiple rounds
// to schedule.
func subtractMax(remaining corev1.ResourceList, instanceTypes []*cloudprovider.InstanceType) corev1.ResourceList {
	// shouldn't occur, but to be safe
	if len(instanceTypes) == 0 {
		return remaining
	}
	var allInstanceResources []corev1.ResourceList
	for _, it := range instanceTypes {
		allInstanceResources = append(allInstanceResources, it.Capacity)
	}
	result := corev1.ResourceList{}
	itResources := resources.MaxResources(allInstanceResources...)
	for k, v := range remaining {
		cp := v.DeepCopy()
		cp.Sub(itResources[k])
		result[k] = cp
	}
	return result
}

// filterByRemainingResources is used to filter out instance types that if launched would exceed the nodepool limits
func filterByRemainingResources(instanceTypes []*cloudprovider.InstanceType, remaining corev1.ResourceList) []*cloudprovider.InstanceType {
	var filtered []*cloudprovider.InstanceType
	for _, it := range instanceTypes {
		itResources := it.Capacity
		viableInstance := true
		for resourceName, remainingQuantity := range remaining {
			// if the instance capacity is greater than the remaining quantity for this resource
			if resources.Cmp(itResources[resourceName], remainingQuantity) > 0 {
				viableInstance = false
			}
		}
		if viableInstance {
			filtered = append(filtered, it)
		}
	}
	return filtered
}
