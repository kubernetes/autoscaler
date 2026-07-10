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

package orchestrator

import (
	"context"
	"fmt"
	"math"
	"slices"
	"sort"
	"strings"
	"time"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	ca_context "k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaleup/equivalence"
	podutils "k8s.io/autoscaler/cluster-autoscaler/utils/pod"
	"k8s.io/autoscaler/cluster-autoscaler/estimator"
	"k8s.io/autoscaler/cluster-autoscaler/expander"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodegroupset"
	"k8s.io/autoscaler/cluster-autoscaler/processors/status"
	"k8s.io/autoscaler/cluster-autoscaler/resourcequotas"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/karpenter"
	"k8s.io/autoscaler/cluster-autoscaler/utils/annotations"
	"k8s.io/utils/clock"

	karpenterv1 "sigs.k8s.io/karpenter/pkg/apis/v1"
	"sigs.k8s.io/karpenter/pkg/controllers/provisioning/scheduling"
	"sigs.k8s.io/karpenter/pkg/controllers/state"
	karpenteroptions "sigs.k8s.io/karpenter/pkg/operator/options"
	karpenterscheduling "sigs.k8s.io/karpenter/pkg/scheduling"
)

// KarpenterSimulator implements ScaleUpSimulator using Karpenter's scheduler.
type KarpenterSimulator struct {
	defaultSimulator ScaleUpSimulator
	converter        karpenter.KarpenterConverter
	processor        nodegroupset.NodeGroupSetProcessor
	batchingEnabled  bool
}

// NewKarpenterSimulator creates a new KarpenterSimulator.
func NewKarpenterSimulator(defaultSimulator ScaleUpSimulator, converter karpenter.KarpenterConverter, processor nodegroupset.NodeGroupSetProcessor, batchingEnabled bool) *KarpenterSimulator {
	return &KarpenterSimulator{
		defaultSimulator: defaultSimulator,
		converter:        converter,
		processor:        processor,
		batchingEnabled:  batchingEnabled,
	}
}

func (s *KarpenterSimulator) Simulate(
	autoscalingCtx *ca_context.AutoscalingContext,
	podEquivalenceGroups []*equivalence.PodGroup,
	unschedulablePods []*apiv1.Pod,
	nodes []*apiv1.Node,
	nodeGroups []cloudprovider.NodeGroup,
	nodeInfos map[string]*framework.NodeInfo,
	tracker *resourcequotas.Tracker,
	now time.Time,
	allOrNothing bool,
) ([][]expander.Option, map[string]status.Reasons, map[string][]estimator.PodEquivalenceGroup, error) {
	var fallbackPods []*apiv1.Pod
	var karpenterPods []*apiv1.Pod
	for _, p := range unschedulablePods {
		if s.isFallbackPod(p) {
			fallbackPods = append(fallbackPods, p)
		} else {
			karpenterPods = append(karpenterPods, p)
		}
	}

	// 1. Mixed all-or-nothing check:
	// WHY: When allOrNothing is requested and unschedulablePods contains both Karpenter-eligible pods and fallback pods,
	// evaluating them via separate simulation paths could cause partial scale-ups or inaccurate capacity decisions.
	// Routing the combined set through defaultSimulator ensures an atomic, all-or-nothing evaluation across all pods.
	if allOrNothing && len(karpenterPods) > 0 && len(fallbackPods) > 0 {
		return s.defaultSimulator.Simulate(autoscalingCtx, podEquivalenceGroups, unschedulablePods, nodes, nodeGroups, nodeInfos, tracker, now, allOrNothing)
	}

	// 2. Karpenter Path (Priority):
	if len(karpenterPods) > 0 {
		// WHY: Under Salvo scale-up mode, large batches of pods are processed in iterative scale-up salvos.
		// Forcing batchingEnabled = true and capping slicedPods to 1000 bounds Karpenter's solver time and memory
		// complexity to O(1000) per iteration. This prevents solver timeouts during massive scale-up spikes while
		// relying on Salvo loop iterations to schedule any remaining pods.
		batchingEnabled := s.batchingEnabled
		if autoscalingCtx != nil && autoscalingCtx.SalvoScaleUp {
			batchingEnabled = true
		}
		limit := len(karpenterPods)
		if batchingEnabled && limit > 1000 {
			limit = 1000
		}
		slicedPods := karpenterPods[:limit]

		decisions, err := s.runKarpenterSimulation(autoscalingCtx, slicedPods, nodes, nodeGroups, nodeInfos, tracker)
		if err != nil {
			return nil, nil, nil, err
		}

		if len(decisions) > 0 {
			schedulable := make(map[string][]estimator.PodEquivalenceGroup)
			var allOptions []expander.Option
			for _, opts := range decisions {
				allOptions = append(allOptions, opts...)
			}
			s.populateSchedulablePodGroups(allOptions, podEquivalenceGroups, schedulable)
			return decisions, nil, schedulable, nil
		}

		if len(fallbackPods) == 0 {
			var skippedNodeGroups map[string]status.Reasons
			if s.defaultSimulator != nil {
				_, skippedNodeGroups, _, err = s.defaultSimulator.Simulate(autoscalingCtx, podEquivalenceGroups, unschedulablePods, nodes, nodeGroups, nodeInfos, tracker, now, allOrNothing)
				if err != nil {
					return nil, nil, nil, err
				}
			}
			return nil, skippedNodeGroups, nil, nil
		}
	}

	// 3. Fallback Path
	return s.defaultSimulator.Simulate(autoscalingCtx, podEquivalenceGroups, unschedulablePods, nodes, nodeGroups, nodeInfos, tracker, now, allOrNothing)
}

// isFallbackPod identifies pods requiring CA's legacy simulation path instead of Karpenter solver.
// WHY: Karpenter's solver does not natively support Dynamic Resource Allocation (DRA) resource claims
// or ProvisioningRequest annotations. Routing these pods to defaultSimulator prevents Karpenter solver
// errors and ensures CA's legacy scheduling predicates handle these specialized workload features.
func (s *KarpenterSimulator) isFallbackPod(p *apiv1.Pod) bool {
	for _, c := range p.Spec.Containers {
		if len(c.Resources.Claims) > 0 {
			return true
		}
	}
	for _, c := range p.Spec.InitContainers {
		if len(c.Resources.Claims) > 0 {
			return true
		}
	}
	if _, ok := p.Annotations["scheduling.k8s.io/provisioning-request"]; ok {
		return true
	}
	return false
}

func (s *KarpenterSimulator) runKarpenterSimulation(
	autoscalingCtx *ca_context.AutoscalingContext,
	pods []*apiv1.Pod,
	nodes []*apiv1.Node,
	nodeGroups []cloudprovider.NodeGroup,
	templateNodeInfos map[string]*framework.NodeInfo,
	tracker *resourcequotas.Tracker,
) ([][]expander.Option, error) {
	// WHY (UID Pod Pointer Reuse): Karpenter's solver, state.Cluster, and VolumeTopology rely heavily on pod.UID as map keys.
	// Speculatively constructed or unschedulable pods from CA may lack a valid UID. Assigning synthetic UIDs
	// ("uid-" + pod.Name or generated UUID) guarantees unique map keying and prevents panics or state lookup misses.
	origPodMap := make(map[*apiv1.Pod]*apiv1.Pod)
	uidSetPods := make([]*apiv1.Pod, len(pods))
	for i, p := range pods {
		if p.UID == "" {
			cloned := p.DeepCopy()
			if cloned.Name != "" {
				cloned.UID = types.UID("uid-" + cloned.Name)
			} else {
				cloned.UID = types.UID(uuid.NewUUID())
			}
			uidSetPods[i] = cloned
			origPodMap[cloned] = p
		} else {
			uidSetPods[i] = p
			origPodMap[p] = p
		}
	}
	pods = uidSetPods

	ctx := karpenteroptions.ToContext(context.Background(), &karpenteroptions.Options{})
	clk := clock.RealClock{}

	// 1. Fork the snapshot to isolate simulation changes
	autoscalingCtx.ClusterSnapshot.Fork()
	defer autoscalingCtx.ClusterSnapshot.Revert()

	// Convert node groups to Karpenter primitives FIRST to populate the ConversionResult
	conv, err := s.converter.Convert(nodeGroups, templateNodeInfos)
	if err != nil {
		return nil, err
	}
	nodePools := conv.NodePools
	instanceTypes := conv.InstanceTypes

	// 2. Track planned sizes and disable offerings for node groups already at MaxSize.
	// WHY: If a NodeGroup has reached MaxSize, allowing Karpenter's solver to consider its corresponding instance type
	// offering could result in NodeClaims for unexpandable node groups. Mutating Available = false on shared instance type
	// offerings prior to running ks.Solve() forces Karpenter to seek alternative valid instance types or node pools.
	currentPlannedSizes := make(map[string]int, len(nodeGroups))
	for _, ng := range nodeGroups {
		sz, _ := ng.TargetSize()
		currentPlannedSizes[ng.Id()] = sz
		if sz >= ng.MaxSize() {
			if off, found := conv.OfferingMap[ng.Id()]; found && off != nil {
				off.Available = false
			}
		}
	}

	nodeInfos, _ := autoscalingCtx.ClusterSnapshot.ListNodeInfos()

	// Hydrate cluster state with pruning for current batch
	relevantPods, relevantNodes, err := karpenter.HydrateClusterState(ctx, autoscalingCtx.ClusterSnapshot, nodeInfos, pods)
	if err != nil {
		return nil, err
	}

	// Initialize DirectClient, state.Cluster and stateNodes
	optimizedClient := karpenter.NewDirectClient(autoscalingCtx.ClusterSnapshot, relevantPods, relevantNodes, nodePools, nil)
	cluster := state.NewCluster(clk, optimizedClient, nil)
	for _, node := range relevantNodes {
		if err := cluster.UpdateNode(ctx, node); err != nil {
			return nil, err
		}
	}
	for _, pod := range relevantPods {
		if err := cluster.UpdatePod(ctx, pod); err != nil {
			return nil, err
		}
	}

	var allStateNodes []*state.StateNode
	var salvoCandidateNodes []*state.StateNode
	for sn := range cluster.Nodes() {
		allStateNodes = append(allStateNodes, sn)
		if annotations.IsSalvoNode(sn.Node) {
			salvoCandidateNodes = append(salvoCandidateNodes, sn)
		}
	}

	topology, err := scheduling.NewTopology(ctx, optimizedClient, cluster, allStateNodes, nodePools, instanceTypes, pods)
	if err != nil {
		return nil, err
	}

	// Initialize VolumeTopology and pre-calculate requirements for batch pods
	volumeTopology := scheduling.NewVolumeTopology(optimizedClient)
	volumeReqsByPod := make(map[types.UID][]karpenterscheduling.Requirements)
	for _, p := range pods {
		reqs, err := volumeTopology.GetRequirements(ctx, p)
		if err == nil && len(reqs) > 0 {
			volumeReqsByPod[p.UID] = reqs
		}
	}

	var daemonSetPods []*apiv1.Pod
	seenDS := sets.New[string]()
	for _, ni := range templateNodeInfos {
		for _, pi := range ni.Pods() {
			if podutils.IsDaemonSetPod(pi.Pod) {
				dsName := pi.Pod.Name
				if controllerRef := metav1.GetControllerOf(pi.Pod); controllerRef != nil {
					dsName = controllerRef.Name
				}
				if !seenDS.Has(dsName) {
					seenDS.Insert(dsName)
					dsPod := pi.Pod.DeepCopy()
					dsPod.Spec.NodeName = ""
					daemonSetPods = append(daemonSetPods, dsPod)
				}
			}
		}
	}

	// WHY (Salvo Node Candidate Placement): filterOutSchedulable has already verified that none of the pending pods
	// can be scheduled on existing non-Salvo nodes in ClusterSnapshot.
	// Therefore, ScaleUp only includes Salvo nodes (upcoming nodes created during earlier Salvo iterations)
	// as candidate existing nodes in salvoCandidateNodes for placing pending pods.
	// Non-Salvo nodes remain in state.Cluster and Topology to preserve full cluster context for pod affinity,
	// anti-affinity, and topology spread constraints.
	ks := scheduling.NewScheduler(
		ctx,
		optimizedClient,
		nodePools,
		cluster,
		salvoCandidateNodes,
		topology,
		instanceTypes,
		daemonSetPods,
		&karpenter.NoopRecorder{},
		clk,
		volumeReqsByPod,
		nil, // dynamicresources.Allocator
		scheduling.NumConcurrentReconciles(autoscalingCtx.AutoscalingOptions.PredicateParallelism),
	)

	results, err := ks.Solve(ctx, pods)
	if err != nil {
		return nil, err
	}

	for _, nc := range results.NewNodeClaims {
		for i, p := range nc.Pods {
			if origPod, ok := origPodMap[p]; ok {
				nc.Pods[i] = origPod
			}
		}
	}

	decisions := s.resolveClaimsToOptions(autoscalingCtx, conv, results.NewNodeClaims, nodeGroups, currentPlannedSizes, templateNodeInfos)
	return decisions, nil
}

func (s *KarpenterSimulator) resolveClaimsToOptions(
	autoscalingCtx *ca_context.AutoscalingContext,
	conv *karpenter.ConversionResult,
	newClaims []*scheduling.NodeClaim,
	nodeGroups []cloudprovider.NodeGroup,
	currentPlannedSizes map[string]int,
	templateNodeInfos map[string]*framework.NodeInfo,
) [][]expander.Option {
	groups := make(map[string][]*scheduling.NodeClaim)
	for _, nc := range newClaims {
		itNames := make([]string, 0, len(nc.InstanceTypeOptions))
		for _, it := range nc.InstanceTypeOptions {
			itNames = append(itNames, it.Name)
		}
		if len(itNames) > 0 {
			nc.Requirements.Add(karpenterscheduling.NewRequirement(apiv1.LabelInstanceTypeStable, apiv1.NodeSelectorOpIn, itNames...))
		}
		key := s.serializeRequirements(nc.Requirements)
		groups[key] = append(groups[key], nc)
	}

	var groupKeys []string
	for k := range groups {
		groupKeys = append(groupKeys, k)
	}
	sort.Strings(groupKeys)

	var decisions [][]expander.Option

	for _, key := range groupKeys {
		slice := groups[key]
		count := len(slice)
		if count == 0 {
			continue
		}
		representativeClaim := slice[0]

		sort.Slice(representativeClaim.InstanceTypeOptions, func(i, j int) bool {
			itI := representativeClaim.InstanceTypeOptions[i]
			itJ := representativeClaim.InstanceTypeOptions[j]
			priceI := math.MaxFloat64
			for _, off := range itI.Offerings {
				if off.Price < priceI {
					priceI = off.Price
				}
			}
			priceJ := math.MaxFloat64
			for _, off := range itJ.Offerings {
				if off.Price < priceJ {
					priceJ = off.Price
				}
			}
			if priceI != priceJ {
				return priceI < priceJ
			}
			return itI.Name < itJ.Name
		})

		var matchingNgs []cloudprovider.NodeGroup
		poolName := representativeClaim.NodePoolName
		for _, it := range representativeClaim.InstanceTypeOptions {
			sourceNgs := conv.NodeGroupsFor(poolName, it.Name)
			for _, ng := range sourceNgs {
				if currentPlannedSizes[ng.Id()] < ng.MaxSize() && s.ngMatchesRequirements(autoscalingCtx, ng, poolName, representativeClaim.Requirements, templateNodeInfos) {
					matchingNgs = append(matchingNgs, ng)
				}
			}
		}

		if len(matchingNgs) == 0 {
			continue
		}

		options := s.buildExpanderOptionsForClaim(autoscalingCtx, conv, matchingNgs, count, slice, representativeClaim, templateNodeInfos)
		if len(options) > 0 {
			decisions = append(decisions, options)
		}
	}
	return decisions
}

func (s *KarpenterSimulator) buildExpanderOptionsForClaim(
	ctx *ca_context.AutoscalingContext,
	conv *karpenter.ConversionResult,
	matchingNgs []cloudprovider.NodeGroup,
	count int,
	claims []*scheduling.NodeClaim,
	representativeClaim *scheduling.NodeClaim,
	templateNodeInfos map[string]*framework.NodeInfo,
) []expander.Option {
	var options []expander.Option
	var allPods []*apiv1.Pod
	for _, nc := range claims {
		allPods = append(allPods, nc.Pods...)
	}

	for _, ng := range matchingNgs {
		isSimilarValid := func(candidate cloudprovider.NodeGroup, candidateNodeInfo *framework.NodeInfo) bool {
			if candidate == nil || candidateNodeInfo == nil || candidateNodeInfo.Node() == nil {
				return false
			}
			candPoolName := conv.PoolForNodeGroup(candidate.Id())
			return s.ngMatchesRequirements(ctx, candidate, candPoolName, representativeClaim.Requirements, templateNodeInfos)
		}

		options = append(options, expander.Option{
			NodeGroup:      ng,
			NodeCount:      count,
			Pods:           allPods,
			Debug:          fmt.Sprintf("Karpenter decision for %d nodes in %s", count, ng.Id()),
			IsSimilarValid: isSimilarValid,
		})
	}

	return options
}

// serializeRequirements creates a deterministic string key for a set of NodeClaim requirements.
// WHY (Requirement Serialization): Karpenter NodeClaims with semantically identical requirements may list label keys or value slices
// in non-deterministic iteration order. Lexicographically sorting requirement keys and allowed value strings guarantees that all
// NodeClaims representing identical node shapes produce the exact same group key string, enabling accurate claim grouping
// and NodeCount aggregation.
func (s *KarpenterSimulator) serializeRequirements(reqs karpenterscheduling.Requirements) string {
	keys := reqs.Keys().UnsortedList()
	sort.Strings(keys)
	var parts []string
	for _, k := range keys {
		req := reqs.Get(k)
		vals := req.Values()
		sort.Strings(vals)
		parts = append(parts, fmt.Sprintf("%s:%s:%s", k, req.Operator(), strings.Join(vals, ",")))
	}
	return strings.Join(parts, ";")
}

func (s *KarpenterSimulator) ngMatchesRequirements(ctx *ca_context.AutoscalingContext, ng cloudprovider.NodeGroup, nodePoolName string, unifiedReqs karpenterscheduling.Requirements, templateNodeInfos map[string]*framework.NodeInfo) bool {
	if ng == nil {
		return false
	}
	var node *apiv1.Node
	if ni, found := templateNodeInfos[ng.Id()]; found && ni != nil {
		node = ni.Node()
	}

	if node == nil {
		if ni, err := ng.TemplateNodeInfo(); err == nil && ni != nil {
			node = ni.Node()
		}
	}

	if node == nil {
		return false
	}

	labels := make(map[string]string)
	for k, v := range node.Labels {
		labels[k] = v
	}
	physicalITName := karpenter.InstanceTypeNameFromLabels(node.Labels, ng.Id())
	if _, ok := labels[apiv1.LabelInstanceTypeStable]; !ok || labels[apiv1.LabelInstanceTypeStable] == "" {
		labels[apiv1.LabelInstanceTypeStable] = physicalITName
	}
	if _, ok := labels[apiv1.LabelArchStable]; !ok {
		if val, betaOk := labels["beta.kubernetes.io/arch"]; betaOk {
			labels[apiv1.LabelArchStable] = val
		} else {
			labels[apiv1.LabelArchStable] = karpenterv1.ArchitectureAmd64
		}
	}
	if _, ok := labels[apiv1.LabelOSStable]; !ok {
		if val, betaOk := labels["beta.kubernetes.io/os"]; betaOk {
			labels[apiv1.LabelOSStable] = val
		} else {
			labels[apiv1.LabelOSStable] = "linux"
		}
	}
	if _, ok := labels[karpenterv1.CapacityTypeLabelKey]; !ok {
		labels[karpenterv1.CapacityTypeLabelKey] = karpenterv1.CapacityTypeOnDemand
	}
	labels[karpenterv1.NodePoolLabelKey] = nodePoolName
	labels["ca.prototype/prototypenodeclass"] = "default"

	nodeReqs := karpenterscheduling.NewLabelRequirements(labels)
	if err := nodeReqs.Compatible(unifiedReqs, karpenterscheduling.AllowUndefinedWellKnownLabels); err != nil {
		return false
	}
	return true
}

// populateSchedulablePodGroups correlates pods scheduled by Karpenter back to CA's PodEquivalenceGroups.
// WHY (Types.NamespacedName Map Keying): Pod pointers may be copied or reconstructed during Karpenter's scheduling and state
// hydration steps, rendering direct pointer identity checks unreliable. Constructing canonical string keys ("Namespace/Name" or
// "Namespace/uid-UID") allows accurate cross-referencing between Karpenter scheduled pod outputs and CA's equivalence groups.
func (s *KarpenterSimulator) populateSchedulablePodGroups(options []expander.Option, podEquivalenceGroups []*equivalence.PodGroup, schedulablePodGroups map[string][]estimator.PodEquivalenceGroup) {
	if schedulablePodGroups == nil {
		return
	}
	podKey := func(p *apiv1.Pod) string {
		if p.Name != "" {
			return p.Namespace + "/" + p.Name
		}
		return p.Namespace + "/uid-" + string(p.UID)
	}
	podToGroup := make(map[string]*equivalence.PodGroup)
	for _, eg := range podEquivalenceGroups {
		for _, p := range eg.Pods {
			key := podKey(p)
			podToGroup[key] = eg
		}
	}
	for _, opt := range options {
		ngId := opt.NodeGroup.Id()
		groups := make(map[*equivalence.PodGroup][]*apiv1.Pod)
		for _, p := range opt.Pods {
			key := podKey(p)
			if eg, ok := podToGroup[key]; ok {
				groups[eg] = append(groups[eg], p)
			}
		}
		for eg, pods := range groups {
			schedulablePodGroups[ngId] = append(schedulablePodGroups[ngId], estimator.PodEquivalenceGroup{
				Pods: pods,
			})
			eg.Schedulable = true
			if !slices.Contains(eg.SchedulableGroups, ngId) {
				eg.SchedulableGroups = append(eg.SchedulableGroups, ngId)
			}
		}
	}
}
