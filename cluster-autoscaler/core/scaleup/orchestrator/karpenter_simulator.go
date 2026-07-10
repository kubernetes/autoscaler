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
	"hash/fnv"
	"math"
	"sort"
	"strings"
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	ca_context "k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaleup/equivalence"
	"k8s.io/autoscaler/cluster-autoscaler/estimator"
	"k8s.io/autoscaler/cluster-autoscaler/expander"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodegroupset"
	"k8s.io/autoscaler/cluster-autoscaler/processors/status"
	"k8s.io/autoscaler/cluster-autoscaler/resourcequotas"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/karpenter"
	"k8s.io/utils/clock"

	karpenterv1 "sigs.k8s.io/karpenter/pkg/apis/v1"
	karpentercloudprovider "sigs.k8s.io/karpenter/pkg/cloudprovider"
	"sigs.k8s.io/karpenter/pkg/controllers/provisioning/scheduling"
	"sigs.k8s.io/karpenter/pkg/controllers/state"
	karpenteroptions "sigs.k8s.io/karpenter/pkg/operator/options"
	karpenterscheduling "sigs.k8s.io/karpenter/pkg/scheduling"
	karpenterresources "sigs.k8s.io/karpenter/pkg/utils/resources"
)

const caIgnoreReservedValue = "ca-ignore-reserved-value"

// DefaultKarpenterConverter implements karpenter.KarpenterConverter.
type DefaultKarpenterConverter struct {
	// ITNameToNodeGroupsMap maps an InstanceType name back to the list of CA NodeGroups it represents.
	ITNameToNodeGroupsMap map[string][]cloudprovider.NodeGroup
	// ITNameToPoolMap maps an InstanceType name back to its parent NodePool name.
	ITNameToPoolMap map[string]string
	// OfferingMap maps a physical NodeGroup.Id() to its constructed Offering.
	OfferingMap     map[string]*karpentercloudprovider.Offering
	// PricingModel is used to calculate offering prices.
	PricingModel    cloudprovider.PricingModel
}

func (c *DefaultKarpenterConverter) ITNameToNodeGroups() map[string][]cloudprovider.NodeGroup {
	return c.ITNameToNodeGroupsMap
}

func (c *DefaultKarpenterConverter) ITNameToPool() map[string]string {
	return c.ITNameToPoolMap
}

// Convert translates CA NodeGroups into Karpenter primitives.
func (c *DefaultKarpenterConverter) Convert(nodeGroups []cloudprovider.NodeGroup, nodeInfos map[string]*framework.NodeInfo) ([]*karpenterv1.NodePool, map[string][]*karpentercloudprovider.InstanceType) {
	c.ITNameToNodeGroupsMap = make(map[string][]cloudprovider.NodeGroup)
	c.ITNameToPoolMap = make(map[string]string)
	c.OfferingMap = make(map[string]*karpentercloudprovider.Offering)

	customLabelKeys := sets.New[string]()
	for _, ng := range nodeGroups {
		ni, found := nodeInfos[ng.Id()]
		if !found {
			continue
		}
		for k := range ni.Node().Labels {
			if k == apiv1.LabelHostname || k == apiv1.LabelTopologyZone || k == apiv1.LabelTopologyRegion || k == karpenterv1.CapacityTypeLabelKey {
				continue
			}
			customLabelKeys.Insert(k)
		}
	}

	type poolGroupKey struct {
		taints string
	}
	poolGroups := make(map[poolGroupKey][]cloudprovider.NodeGroup)

	for _, ng := range nodeGroups {
		ni, found := nodeInfos[ng.Id()]
		if !found {
			continue
		}
		node := ni.Node()
		key := poolGroupKey{
			taints: c.SerializeTaints(node.Spec.Taints),
		}
		poolGroups[key] = append(poolGroups[key], ng)
	}

	var nodePools []*karpenterv1.NodePool
	itMap := make(map[string][]*karpentercloudprovider.InstanceType)

	for key, groups := range poolGroups {
		hasher := fnv.New32a()
		hasher.Write([]byte(key.taints))
		hashStr := fmt.Sprintf("%08x", hasher.Sum32())
		nodePoolName := fmt.Sprintf("pool-%s", hashStr)

		// Group by 'Static' Identity: Physical IT Name + Resources + Static Labels
		type itIdentity struct {
			itName    string
			resources string
			dsKeys    string
		}
		collapsedGroups := make(map[itIdentity][]cloudprovider.NodeGroup)
		for _, ng := range groups {
			ni := nodeInfos[ng.Id()]
			node := ni.Node()
			id := itIdentity{
				itName:    c.GetPhysicalITName(node.Labels, ng.Id()),
				resources: c.SerializeResources(node.Status.Capacity),
				dsKeys:    getCompatibleDSKeys(ng, nodeInfos),
			}
			collapsedGroups[id] = append(collapsedGroups[id], ng)
		}

		var keys []itIdentity
		for k := range collapsedGroups {
			keys = append(keys, k)
		}
		sort.Slice(keys, func(i, j int) bool {
			if keys[i].itName != keys[j].itName {
				return keys[i].itName < keys[j].itName
			}
			if keys[i].resources != keys[j].resources {
				return keys[i].resources < keys[j].resources
			}
			return keys[i].dsKeys < keys[j].dsKeys
		})

		var instanceTypes []*karpentercloudprovider.InstanceType
		for _, id := range keys {
			ngs := collapsedGroups[id]
			// Sort ngs by Id to ensure deterministic representative selection
			sort.Slice(ngs, func(i, j int) bool {
				return ngs[i].Id() < ngs[j].Id()
			})
			itName := id.itName
			// Ensure Karpenter IT names remain unique if different static identities share one physical name
			if _, exists := c.ITNameToNodeGroupsMap[itName]; exists {
				itName = fmt.Sprintf("%s-%d", itName, len(c.ITNameToNodeGroupsMap))
			}

			c.ITNameToNodeGroupsMap[itName] = ngs
			c.ITNameToPoolMap[itName] = nodePoolName
			it := c.CreateInstanceType(itName, ngs, nodeInfos, nodePoolName, customLabelKeys)
			instanceTypes = append(instanceTypes, it)
		}

		// Sort instance types by price (cheapest first) as Karpenter's scheduler expects them to be sorted by price
		sort.Slice(instanceTypes, func(i, j int) bool {
			iPrice := math.MaxFloat64
			for _, off := range instanceTypes[i].Offerings {
				if off.Price < iPrice {
					iPrice = off.Price
				}
			}
			jPrice := math.MaxFloat64
			for _, off := range instanceTypes[j].Offerings {
				if off.Price < jPrice {
					jPrice = off.Price
				}
			}
			if iPrice != jPrice {
				return iPrice < jPrice
			}
			return instanceTypes[i].Name < instanceTypes[j].Name
		})

		nodePool := &karpenterv1.NodePool{
			ObjectMeta: metav1.ObjectMeta{
				Name: nodePoolName,
				UID:  types.UID(nodePoolName + "-uid"),
			},
			Spec: karpenterv1.NodePoolSpec{
				Template: karpenterv1.NodeClaimTemplate{
					Spec: karpenterv1.NodeClaimTemplateSpec{
						NodeClassRef: &karpenterv1.NodeClassReference{
							Group: "ca.prototype",
							Kind:  "PrototypeNodeClass",
							Name:  "default",
						},
						Taints: nodeInfos[groups[0].Id()].Node().Spec.Taints,
					},
				},
			},
		}

		// Collect all labels from all node groups in this pool to build NodePool requirements.
		// This ensures that pods requesting these labels won't be rejected by Karpenter's scheduler
		// if they are present in CA's NodeGroups.
		// We must union values for the same key across different instance types to avoid an empty intersection.
		combinedValues := make(map[string]sets.Set[string])
		for _, it := range instanceTypes {
			for _, req := range it.Requirements.Values() {
				if karpenterv1.RestrictedLabels.Has(req.Key) {
					continue
				}
				if _, ok := combinedValues[req.Key]; !ok {
					combinedValues[req.Key] = sets.New[string]()
				}
				combinedValues[req.Key].Insert(req.Values()...)
			}
			for _, off := range it.Offerings {
				for _, req := range off.Requirements.Values() {
					if karpenterv1.RestrictedLabels.Has(req.Key) {
						continue
					}
					if _, ok := combinedValues[req.Key]; !ok {
						combinedValues[req.Key] = sets.New[string]()
					}
					combinedValues[req.Key].Insert(req.Values()...)
				}
			}
		}
		combinedRequirements := karpenterscheduling.NewRequirements()
		for key, values := range combinedValues {
			if karpenterv1.WellKnownLabels.Has(key) {
				combinedRequirements.Add(karpenterscheduling.NewRequirement(key, apiv1.NodeSelectorOpIn, values.UnsortedList()...))
			} else {
				combinedRequirements.Add(karpenterscheduling.NewRequirement(key, apiv1.NodeSelectorOpNotIn, caIgnoreReservedValue))
			}
		}
		nodePool.Spec.Template.Spec.Requirements = combinedRequirements.NodeSelectorRequirements()

		nodePools = append(nodePools, nodePool)
		itMap[nodePoolName] = instanceTypes
	}

	sort.Slice(nodePools, func(i, j int) bool {
		return nodePools[i].Name < nodePools[j].Name
	})

	return nodePools, itMap
}

func (c *DefaultKarpenterConverter) GetPhysicalITName(labels map[string]string, defaultName string) string {
	if val, ok := labels[apiv1.LabelInstanceTypeStable]; ok {
		return val
	}
	if val, ok := labels[apiv1.LabelInstanceType]; ok {
		return val
	}
	return defaultName
}

func (c *DefaultKarpenterConverter) SerializeTaints(taints []apiv1.Taint) string {
	var parts []string
	for _, t := range taints {
		parts = append(parts, fmt.Sprintf("%s=%s:%s", t.Key, t.Value, t.Effect))
	}
	sort.Strings(parts)
	return strings.Join(parts, ",")
}

func (c *DefaultKarpenterConverter) SerializeResources(res apiv1.ResourceList) string {
	return fmt.Sprintf("c:%d,m:%d", res.Cpu().MilliValue(), res.Memory().Value())
}

func (c *DefaultKarpenterConverter) SerializeStaticLabels(labels map[string]string) string {
	var parts []string
	for k, v := range labels {
		if k == apiv1.LabelHostname || k == apiv1.LabelTopologyZone || k == apiv1.LabelTopologyRegion || k == karpenterv1.CapacityTypeLabelKey {
			continue
		}
		parts = append(parts, k+"="+v)
	}
	sort.Strings(parts)
	return strings.Join(parts, ",")
}

func (c *DefaultKarpenterConverter) CreateInstanceType(name string, ngs []cloudprovider.NodeGroup, nodeInfos map[string]*framework.NodeInfo, nodePoolName string, customLabelKeys sets.Set[string]) *karpentercloudprovider.InstanceType {
	representativeNg := ngs[0]
	representativeNi := nodeInfos[representativeNg.Id()]
	node := representativeNi.Node()

	// Collect homogeneous labels across all node groups in this IT
	homogeneousLabels := make(map[string]string)
	if len(ngs) > 0 {
		firstNi := nodeInfos[ngs[0].Id()]
		if firstNi != nil {
			for k, v := range firstNi.Node().Labels {
				if k == apiv1.LabelHostname || k == apiv1.LabelTopologyZone || k == apiv1.LabelTopologyRegion || k == karpenterv1.CapacityTypeLabelKey {
					continue
				}
				homogeneousLabels[k] = v
			}
		}
		for _, ng := range ngs[1:] {
			ni := nodeInfos[ng.Id()]
			if ni == nil {
				continue
			}
			for k, v := range homogeneousLabels {
				if val, ok := ni.Node().Labels[k]; !ok || val != v {
					delete(homogeneousLabels, k)
				}
			}
		}
	}

	labels := make(map[string]string)
	labels[apiv1.LabelArchStable] = karpenterv1.ArchitectureAmd64
	labels[apiv1.LabelOSStable] = "linux"
	labels[apiv1.LabelInstanceTypeStable] = name
	labels[karpenterv1.NodePoolLabelKey] = nodePoolName
	labels["ca.prototype/prototypenodeclass"] = "default"
	for k, v := range homogeneousLabels {
		labels[k] = v
	}

	requirements := karpenterscheduling.NewLabelRequirements(labels)

	var offerings karpentercloudprovider.Offerings
	for _, ng := range ngs {
		ni := nodeInfos[ng.Id()]
		n := ni.Node()
		zone := n.Labels[apiv1.LabelTopologyZone]
		if zone == "" {
			zone = "default-zone"
		}
		capType := n.Labels[karpenterv1.CapacityTypeLabelKey]
		if capType == "" {
			capType = karpenterv1.CapacityTypeOnDemand
		}

		offeringRequirements := karpenterscheduling.NewRequirements(
			karpenterscheduling.NewRequirement(karpenterv1.CapacityTypeLabelKey, apiv1.NodeSelectorOpIn, capType),
		)
		if zone != "" && zone != "default-zone" {
			offeringRequirements.Add(karpenterscheduling.NewRequirement(apiv1.LabelTopologyZone, apiv1.NodeSelectorOpIn, zone))
		}
		for key := range customLabelKeys {
			if val, ok := n.Labels[key]; ok {
				offeringRequirements.Add(karpenterscheduling.NewRequirement(key, apiv1.NodeSelectorOpIn, val))
			} else {
				offeringRequirements.Add(karpenterscheduling.NewRequirement(key, apiv1.NodeSelectorOpDoesNotExist))
			}
		}

		cpu := float64(n.Status.Capacity.Cpu().MilliValue()) / 1000.0
		memGiB := float64(n.Status.Capacity.Memory().Value()) / (1024.0 * 1024.0 * 1024.0)
		gpu := 0.0
		for k, v := range n.Status.Capacity {
			if strings.Contains(string(k), "gpu") {
				gpu += float64(v.Value())
			}
		}

		price := (1.0 * cpu) + (0.125 * memGiB) + (25.0 * gpu)
		if capType == karpenterv1.CapacityTypeSpot {
			price = price * 0.3
		}
		if c.PricingModel != nil {
			if p, err := c.PricingModel.NodePrice(n, time.Now(), time.Now().Add(time.Hour)); err == nil {
				price = p
			}
		}

		offering := &karpentercloudprovider.Offering{
			Available:    true,
			Price:        price,
			Requirements: offeringRequirements,
		}
		offerings = append(offerings, offering)
		c.OfferingMap[ng.Id()] = offering
	}

	capacity := node.Status.Capacity.DeepCopy()
	if representativeNi.CSINode != nil {
		for _, driver := range representativeNi.CSINode.Spec.Drivers {
			if driver.Allocatable != nil && driver.Allocatable.Count != nil {
				capacity[apiv1.ResourceName(driver.Name)] = *resource.NewQuantity(int64(*driver.Allocatable.Count), resource.DecimalSI)
			}
		}
	}

	return &karpentercloudprovider.InstanceType{
		Name:         name,
		Requirements: requirements,
		Offerings:    offerings,
		Capacity:     capacity,
		Overhead: &karpentercloudprovider.InstanceTypeOverhead{
			KubeReserved:   karpenterresources.Subtract(node.Status.Capacity, node.Status.Allocatable),
			SystemReserved: apiv1.ResourceList{},
		},
	}
}

// KarpenterSimulator implements ScaleUpSimulator using Karpenter's scheduler.
type KarpenterSimulator struct {
	defaultSimulator ScaleUpSimulator
	converter        karpenter.KarpenterConverter
	processor        nodegroupset.NodeGroupSetProcessor
}

// NewKarpenterSimulator creates a new KarpenterSimulator.
func NewKarpenterSimulator(defaultSimulator ScaleUpSimulator, converter karpenter.KarpenterConverter, processor nodegroupset.NodeGroupSetProcessor) *KarpenterSimulator {
	return &KarpenterSimulator{
		defaultSimulator: defaultSimulator,
		converter:        converter,
		processor:        processor,
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
) ([]expander.Option, map[string]status.Reasons, map[string][]estimator.PodEquivalenceGroup, error) {
	var fallbackPods []*apiv1.Pod
	var karpenterPods []*apiv1.Pod
	for _, p := range unschedulablePods {
		if s.isFallbackPod(p) {
			fallbackPods = append(fallbackPods, p)
		} else {
			karpenterPods = append(karpenterPods, p)
		}
	}

	// 1. Mixed all-or-nothing check
	if allOrNothing && len(karpenterPods) > 0 && len(fallbackPods) > 0 {
		return s.defaultSimulator.Simulate(autoscalingCtx, podEquivalenceGroups, unschedulablePods, nodes, nodeGroups, nodeInfos, tracker, now, allOrNothing)
	}

	// 2. Karpenter Path (Priority)
	if len(karpenterPods) > 0 {
		options, err := s.runKarpenterSimulation(autoscalingCtx, karpenterPods, nodes, nodeGroups, nodeInfos, tracker)
		if err != nil {
			return nil, nil, nil, err
		}

		if len(options) > 0 {
			schedulable := make(map[string][]estimator.PodEquivalenceGroup)
			s.populateSchedulablePodGroups(options, podEquivalenceGroups, schedulable)
			return options, nil, schedulable, nil
		}

		if len(fallbackPods) == 0 {
			return nil, nil, nil, nil
		}
	}

	// 3. Fallback Path
	return s.defaultSimulator.Simulate(autoscalingCtx, podEquivalenceGroups, unschedulablePods, nodes, nodeGroups, nodeInfos, tracker, now, allOrNothing)
}

func (s *KarpenterSimulator) targetSizeOrZero(ng cloudprovider.NodeGroup) int {
	sz, err := ng.TargetSize()
	if err != nil {
		return 0
	}
	return sz
}

func (s *KarpenterSimulator) isFallbackPod(p *apiv1.Pod) bool {
	for _, c := range p.Spec.Containers {
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
) ([]expander.Option, error) {
	ctx := karpenteroptions.ToContext(context.Background(), &karpenteroptions.Options{})

	clk := clock.RealClock{}
	nodeInfos, _ := autoscalingCtx.ClusterSnapshot.ListNodeInfos()

	// 1. Fork the snapshot to isolate simulation changes
	autoscalingCtx.ClusterSnapshot.Fork()
	defer autoscalingCtx.ClusterSnapshot.Revert()

	// Convert node groups to Karpenter primitives FIRST to populate the OfferingMap
	nodePools, instanceTypes := s.converter.Convert(nodeGroups, templateNodeInfos)

	// 2. Track current planned sizes to enforce MaxSize limit enforcement
	currentPlannedSizes := make(map[string]int)
	for _, ng := range nodeGroups {
		sz, _ := ng.TargetSize()
		currentPlannedSizes[ng.Id()] = sz
	}

	// 3. Pre-Simulation Pruning: Disable offerings for node groups already at MaxSize
	if defConv, ok := s.converter.(*DefaultKarpenterConverter); ok {
		for _, ng := range nodeGroups {
			if currentPlannedSizes[ng.Id()] >= ng.MaxSize() {
				if off, found := defConv.OfferingMap[ng.Id()]; found {
					off.Available = false
				}
			}
		}
	}

	batchSize := 500
	remainingPods := append([]*apiv1.Pod{}, pods...)
	var allNewNodeClaims []*scheduling.NodeClaim
	allPodErrors := make(map[*apiv1.Pod]error)

	finalPlan := make(map[string]int)

	for len(remainingPods) > 0 {
		limit := batchSize
		if len(remainingPods) < limit {
			limit = len(remainingPods)
		}
		batchPods := remainingPods[:limit]
		remainingPods = remainingPods[limit:]

		// Reload snapshot node infos dynamically to capture simulated nodes/pods committed in prior batches
		nodeInfos, _ = autoscalingCtx.ClusterSnapshot.ListNodeInfos()

		// Hydrate cluster state with pruning for current batch
		relevantPods, relevantNodes, err := karpenter.HydrateClusterState(ctx, autoscalingCtx.ClusterSnapshot, nodeInfos, batchPods)
		if err != nil {
			return nil, err
		}

		// Retrieve all pods currently in the snapshot to pass to DirectClient
		var snapshotPods []*apiv1.Pod
		for _, ni := range nodeInfos {
			for _, pi := range ni.Pods() {
				snapshotPods = append(snapshotPods, pi.Pod)
			}
		}
		// Initialize DirectClient, state.Cluster and stateNodes
		optimizedClient := karpenter.NewDirectClient(autoscalingCtx.ClusterSnapshot, append(snapshotPods, batchPods...), nodePools, nil)
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

		var stateNodes []*state.StateNode
		for sn := range cluster.Nodes() {
			stateNodes = append(stateNodes, sn)
		}

		topology, err := scheduling.NewTopology(ctx, optimizedClient, cluster, stateNodes, nodePools, instanceTypes, batchPods)
		if err != nil {
			return nil, err
		}

		// Initialize VolumeTopology and pre-calculate requirements for batch pods
		volumeTopology := scheduling.NewVolumeTopology(optimizedClient)
		volumeReqsByPod := make(map[types.UID][]karpenterscheduling.Requirements)
		for _, p := range batchPods {
			reqs, err := volumeTopology.GetRequirements(ctx, p)
			if err == nil && len(reqs) > 0 {
				volumeReqsByPod[p.UID] = reqs
			}
		}

		var daemonSetPods []*apiv1.Pod
		seenDS := sets.New[string]()
		for _, ni := range templateNodeInfos {
			for _, pi := range ni.Pods() {
				if dsName, ok := isDaemonSetPod(pi.Pod); ok {
					if !seenDS.Has(dsName) {
						seenDS.Insert(dsName)
						dsPod := pi.Pod.DeepCopy()
						dsPod.Spec.NodeName = ""
						daemonSetPods = append(daemonSetPods, dsPod)
					}
				}
			}
		}

		ks := scheduling.NewScheduler(
			ctx,
			optimizedClient,
			nodePools,
			cluster,
			stateNodes,
			topology,
			instanceTypes,
			daemonSetPods,
			&karpenter.NoopRecorder{},
			clk,
			volumeReqsByPod,
			nil, // dynamicresources.Allocator
			scheduling.NumConcurrentReconciles(autoscalingCtx.AutoscalingOptions.PredicateParallelism),
		)

		results, err := ks.Solve(ctx, batchPods)
		if err != nil {
			return nil, err
		}

		// Resolve and Commit claims using our new Group-Expand-Balance logic
		committed, batchDeltas, recycled := s.resolveAndCommitClaims(autoscalingCtx, results.NewNodeClaims, nodeGroups, currentPlannedSizes, templateNodeInfos, tracker)

		allNewNodeClaims = append(allNewNodeClaims, committed...)
		remainingPods = append(remainingPods, recycled...)

		for ngId, delta := range batchDeltas {
			finalPlan[ngId] += delta
		}

		for p, err := range results.PodErrors {
			allPodErrors[p] = err
		}
	}

	// Build the final CompositeNodeGroup and Option
	fullPlan := []nodegroupset.ScaleUpInfo{}
	totalNodes := 0
	var representativeNg cloudprovider.NodeGroup
	maxDelta := -1

	nodeGroupsMap := make(map[string]cloudprovider.NodeGroup)
	for _, ng := range nodeGroups {
		nodeGroupsMap[ng.Id()] = ng
	}

	for ngId, delta := range finalPlan {
		group, ok := nodeGroupsMap[ngId]
		if !ok {
			continue
		}
		targetSize, _ := group.TargetSize()
		fullPlan = append(fullPlan, nodegroupset.ScaleUpInfo{
			Group:       group,
			CurrentSize: targetSize,
			NewSize:     targetSize + delta,
			MaxSize:     group.MaxSize(),
		})
		totalNodes += delta
		if delta > maxDelta {
			maxDelta = delta
			representativeNg = group
		}
	}

	if len(fullPlan) == 0 {
		return nil, nil
	}

	// Sort fullPlan by Group ID to guarantee absolute determinism
	sort.Slice(fullPlan, func(i, j int) bool {
		return fullPlan[i].Group.Id() < fullPlan[j].Group.Id()
	})

	compositeNg := &CompositeNodeGroup{
		NodeGroup: representativeNg,
		Plan:      fullPlan,
	}

	var allPods []*apiv1.Pod
	for _, nc := range allNewNodeClaims {
		allPods = append(allPods, nc.Pods...)
	}

	return []expander.Option{
		{
			NodeGroup: compositeNg,
			NodeCount: totalNodes,
			Pods:      allPods,
			Debug:     fmt.Sprintf("Karpenter multi-offering plan for %d nodes across %d groups", totalNodes, len(fullPlan)),
		},
	}, nil
}

// TODO(x13n): Implement Resource Clamping & Pod Pruning to further optimize topology initialization.
// By subtracting resources of irrelevant pods from node allocatable and hiding those pods from the topology map,
// we can significantly reduce O(N) complexity of affinity calculations without losing capacity accuracy.
func (s *KarpenterSimulator) resolveAndCommitClaims(
	autoscalingCtx *ca_context.AutoscalingContext,
	newClaims []*scheduling.NodeClaim,
	nodeGroups []cloudprovider.NodeGroup,
	currentPlannedSizes map[string]int,
	templateNodeInfos map[string]*framework.NodeInfo,
	tracker *resourcequotas.Tracker,
) (committed []*scheduling.NodeClaim, batchDeltas map[string]int, recycled []*apiv1.Pod) {
	batchDeltas = make(map[string]int)

	groups := make(map[string][]*scheduling.NodeClaim)
	for _, nc := range newClaims {
		key := s.serializeRequirements(nc.Requirements)
		groups[key] = append(groups[key], nc)
	}

	var groupKeys []string
	for k := range groups {
		groupKeys = append(groupKeys, k)
	}
	sort.Strings(groupKeys)

	for _, key := range groupKeys {
		slice := groups[key]
		count := len(slice)
		if count == 0 {
			continue
		}
		representativeClaim := slice[0]

		var eligibleNgs []cloudprovider.NodeGroup
		var matchingNgs []cloudprovider.NodeGroup

		for _, it := range representativeClaim.InstanceTypeOptions {
			sourceNgs, ok := s.converter.ITNameToNodeGroups()[it.Name]
			if !ok {
				continue
			}
			poolName := s.converter.ITNameToPool()[it.Name]
			var typeMatchingNgs []cloudprovider.NodeGroup
			for _, ng := range sourceNgs {
				if s.ngMatchesRequirements(autoscalingCtx, ng, it.Name, poolName, representativeClaim.Requirements, templateNodeInfos) {
					typeMatchingNgs = append(typeMatchingNgs, ng)
					matchingNgs = append(matchingNgs, ng)
				}
			}

			var typeEligibleNgs []cloudprovider.NodeGroup
			for _, ng := range typeMatchingNgs {
				if currentPlannedSizes[ng.Id()] < ng.MaxSize() {
					typeEligibleNgs = append(typeEligibleNgs, ng)
				}
			}

			if len(typeEligibleNgs) > 0 {
				eligibleNgs = typeEligibleNgs
				break
			}
		}

		if len(eligibleNgs) == 0 {
			if len(matchingNgs) > 0 {
				s.pruneOfferings(matchingNgs)
			}
			for _, nc := range slice {
				recycled = append(recycled, nc.Pods...)
			}
			continue
		}

		options := s.buildExpanderOptions(autoscalingCtx, eligibleNgs, count, slice, templateNodeInfos)
		if len(options) == 0 {
			s.pruneOfferings(matchingNgs)
			for _, nc := range slice {
				recycled = append(recycled, nc.Pods...)
			}
			continue
		}

		bestOption := autoscalingCtx.ExpanderStrategy.BestOption(options, templateNodeInfos)
		if bestOption == nil {
			for _, nc := range slice {
				recycled = append(recycled, nc.Pods...)
			}
			continue
		}

		allowedCount := count
		nodeInfo, found := templateNodeInfos[bestOption.NodeGroup.Id()]
		if found {
			checkResult, err := tracker.CheckQuota(autoscalingCtx, bestOption.NodeGroup, nodeInfo.Node(), allowedCount)
			if err == nil {
				allowedCount = checkResult.AllowedDelta
			}
		}

		similaritySet := append([]cloudprovider.NodeGroup{bestOption.NodeGroup}, bestOption.SimilarNodeGroups...)
		balancePlan, err := s.processor.BalanceScaleUpBetweenGroups(autoscalingCtx, similaritySet, allowedCount)

		finalAllowedCount := 0
		if err != nil {
			ng := bestOption.NodeGroup
			delta := ng.MaxSize() - currentPlannedSizes[ng.Id()]
			if delta > allowedCount {
				delta = allowedCount
			}
			if delta > 0 {
				if found {
					checkResult, err := tracker.CheckQuota(autoscalingCtx, ng, nodeInfo.Node(), delta)
					if err == nil {
						delta = checkResult.AllowedDelta
					}
					if delta > 0 {
						if _, err := tracker.ConsumeQuota(autoscalingCtx, ng, nodeInfo.Node(), delta); err != nil {
							// log error or ignore
						}
					}
				}
				if delta > 0 {
					batchDeltas[ng.Id()] += delta
					currentPlannedSizes[ng.Id()] += delta
					finalAllowedCount = delta
				}
			}
		} else {
			for i := range balancePlan {
				sui := &balancePlan[i]
				delta := sui.NewSize - sui.CurrentSize
				if delta <= 0 {
					continue
				}
				var ngNodeInfo *framework.NodeInfo
				for _, g := range similaritySet {
					if g.Id() == sui.Group.Id() {
						ngNodeInfo = templateNodeInfos[g.Id()]
						break
					}
				}
				if ngNodeInfo == nil {
					continue
				}
				checkResult, err := tracker.CheckQuota(autoscalingCtx, sui.Group, ngNodeInfo.Node(), delta)
				if err != nil {
					continue
				}
				allowedDelta := checkResult.AllowedDelta
				if allowedDelta < delta {
					sui.NewSize = sui.CurrentSize + allowedDelta
				}
				if allowedDelta > 0 {
					if _, err := tracker.ConsumeQuota(autoscalingCtx, sui.Group, ngNodeInfo.Node(), allowedDelta); err != nil {
						// log error
					}
				}
			}

			for _, sui := range balancePlan {
				delta := sui.NewSize - sui.CurrentSize
				if delta > 0 {
					batchDeltas[sui.Group.Id()] += delta
					currentPlannedSizes[sui.Group.Id()] += delta
					finalAllowedCount += delta
				}
			}
		}

		var allowedClaims []*scheduling.NodeClaim
		if finalAllowedCount < count {
			allowedClaims = slice[:finalAllowedCount]
			for _, nc := range slice[finalAllowedCount:] {
				recycled = append(recycled, nc.Pods...)
			}
		} else {
			allowedClaims = slice
		}

		committed = append(committed, allowedClaims...)

		var targetGroups []cloudprovider.NodeGroup
		for suiGroup, delta := range batchDeltas {
			var group cloudprovider.NodeGroup
			for _, g := range similaritySet {
				if g.Id() == suiGroup {
					group = g
					break
				}
			}
			if group != nil {
				for i := 0; i < delta; i++ {
					targetGroups = append(targetGroups, group)
				}
			}
		}

		for i, nc := range allowedClaims {
			if i >= len(targetGroups) {
				break
			}
			targetNg := targetGroups[i]

			var templateNode *apiv1.Node
			if ni, found := templateNodeInfos[targetNg.Id()]; found {
				templateNode = ni.Node()
			}
			if templateNode == nil {
				ni, _ := targetNg.TemplateNodeInfo()
				if ni != nil {
					templateNode = ni.Node()
				}
			}

			if templateNode != nil {
				simNodeName := fmt.Sprintf("simulated-node-%s-%s", targetNg.Id(), uuid.NewUUID())
				simNode := templateNode.DeepCopy()
				simNode.Name = simNodeName
				if simNode.Labels == nil {
					simNode.Labels = make(map[string]string)
				}
				simNode.Labels[apiv1.LabelHostname] = simNodeName
				simNode.Labels["ca.prototype/nodegroup-id"] = targetNg.Id()

				simNodeInfo := framework.NewNodeInfo(simNode, nil)
				if err := autoscalingCtx.ClusterSnapshot.AddNodeInfo(simNodeInfo); err == nil {
					for _, p := range nc.Pods {
						p.Spec.NodeName = simNodeName
						_ = autoscalingCtx.ClusterSnapshot.ForceAddPod(p, simNodeName)
					}
				}
			}
		}
	}

	return committed, batchDeltas, recycled
}

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

func (s *KarpenterSimulator) buildExpanderOptions(
	ctx *ca_context.AutoscalingContext,
	matchingNgs []cloudprovider.NodeGroup,
	count int,
	claims []*scheduling.NodeClaim,
	templateNodeInfos map[string]*framework.NodeInfo,
) []expander.Option {
	var options []expander.Option
	processed := make(map[string]bool)

	matchingMap := make(map[string]cloudprovider.NodeGroup)
	for _, ng := range matchingNgs {
		matchingMap[ng.Id()] = ng
	}

	var allPods []*apiv1.Pod
	for _, nc := range claims {
		allPods = append(allPods, nc.Pods...)
	}

	for _, ng := range matchingNgs {
		if processed[ng.Id()] {
			continue
		}

		similar, err := s.processor.FindSimilarNodeGroups(ctx, ng, templateNodeInfos)
		if err != nil {
			similar = []cloudprovider.NodeGroup{}
		}

		var validSimilar []cloudprovider.NodeGroup
		for _, sng := range similar {
			if _, matches := matchingMap[sng.Id()]; matches {
				validSimilar = append(validSimilar, sng)
			}
		}

		processed[ng.Id()] = true
		for _, sng := range validSimilar {
			processed[sng.Id()] = true
		}

		options = append(options, expander.Option{
			NodeGroup:         ng,
			SimilarNodeGroups: validSimilar,
			NodeCount:         count,
			Pods:              allPods,
		})
	}

	return options
}

func (s *KarpenterSimulator) pruneOfferings(ngs []cloudprovider.NodeGroup) {
	if defConv, ok := s.converter.(*DefaultKarpenterConverter); ok {
		for _, ng := range ngs {
			if off, found := defConv.OfferingMap[ng.Id()]; found {
				off.Available = false
			}
		}
	}
}

func (s *KarpenterSimulator) ngMatchesRequirements(ctx *ca_context.AutoscalingContext, ng cloudprovider.NodeGroup, itName string, nodePoolName string, reqs karpenterscheduling.Requirements, templateNodeInfos map[string]*framework.NodeInfo) bool {
	var node *apiv1.Node
	if ni, found := templateNodeInfos[ng.Id()]; found {
		node = ni.Node()
	}

	if node == nil {
		nodeInfos, _ := ctx.ClusterSnapshot.ListNodeInfos()
		for _, ni := range nodeInfos {
			if ni.Node().Labels["ca.prototype/nodegroup-id"] == ng.Id() || strings.Contains(ni.Node().Name, ng.Id()) {
				node = ni.Node()
				break
			}
		}
	}

	if node == nil {
		ni, _ := ng.TemplateNodeInfo()
		if ni != nil {
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
	labels[apiv1.LabelInstanceTypeStable] = itName
	labels[apiv1.LabelArchStable] = karpenterv1.ArchitectureAmd64
	labels[apiv1.LabelOSStable] = "linux"
	if _, ok := labels[karpenterv1.CapacityTypeLabelKey]; !ok {
		labels[karpenterv1.CapacityTypeLabelKey] = karpenterv1.CapacityTypeOnDemand
	}
	labels[karpenterv1.NodePoolLabelKey] = nodePoolName
	labels["ca.prototype/prototypenodeclass"] = "default"

	nodeReqs := karpenterscheduling.NewLabelRequirements(labels)
	if err := nodeReqs.Compatible(reqs, karpenterscheduling.AllowUndefinedWellKnownLabels); err != nil {
		return false
	}
	return true
}

func (s *KarpenterSimulator) populateSchedulablePodGroups(options []expander.Option, podEquivalenceGroups []*equivalence.PodGroup, schedulablePodGroups map[string][]estimator.PodEquivalenceGroup) {
	if schedulablePodGroups == nil {
		return
	}
	podToGroup := make(map[string]*equivalence.PodGroup)
	for _, eg := range podEquivalenceGroups {
		for _, p := range eg.Pods {
			key := p.Namespace + "/" + p.Name
			podToGroup[key] = eg
		}
	}
	for _, opt := range options {
		ngId := opt.NodeGroup.Id()
		groups := make(map[*equivalence.PodGroup][]*apiv1.Pod)
		for _, p := range opt.Pods {
			key := p.Namespace + "/" + p.Name
			if eg, ok := podToGroup[key]; ok {
				groups[eg] = append(groups[eg], p)
			}
		}
		for eg, pods := range groups {
			schedulablePodGroups[ngId] = append(schedulablePodGroups[ngId], estimator.PodEquivalenceGroup{
				Pods: pods,
			})
			eg.Schedulable = true
			eg.SchedulableGroups = append(eg.SchedulableGroups, ngId)
		}
	}
}



func isDaemonSetPod(pod *apiv1.Pod) (string, bool) {
	for _, ref := range pod.OwnerReferences {
		if ref.Kind == "DaemonSet" {
			return ref.Name, true
		}
	}
	return "", false
}

func getCompatibleDSKeys(ng cloudprovider.NodeGroup, nodeInfos map[string]*framework.NodeInfo) string {
	ni, found := nodeInfos[ng.Id()]
	if !found {
		return ""
	}
	var dsKeys []string
	for _, pi := range ni.Pods() {
		if dsName, ok := isDaemonSetPod(pi.Pod); ok {
			dsKeys = append(dsKeys, dsName)
		}
	}
	sort.Strings(dsKeys)
	return strings.Join(dsKeys, ",")
}
