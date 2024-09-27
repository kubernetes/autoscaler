/*
Copyright 2022 The Kubernetes Authors.

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

package planner

import (
	"fmt"
	"math"
	"time"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/eligibility"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/pdb"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/resource"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/unneeded"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/unremovable"
	"k8s.io/autoscaler/cluster-autoscaler/processors"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodes"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/drainability/rules"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/options"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/scheduling"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/utilization"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	pod_util "k8s.io/autoscaler/cluster-autoscaler/utils/pod"
	klog "k8s.io/klog/v2"
)

type eligibilityChecker interface {
	FilterOutUnremovable(context *context.AutoscalingContext, scaleDownCandidates []*apiv1.Node, timestamp time.Time, unremovableNodes *unremovable.Nodes) ([]string, map[string]utilization.Info, []*simulator.UnremovableNode)
}

type removalSimulator interface {
	DropOldHints()
	SimulateNodeRemoval(node string, podDestinations map[string]bool, timestamp time.Time, remainingPdbTracker pdb.RemainingPdbTracker) (*simulator.NodeToBeRemoved, *simulator.UnremovableNode)
}

// controllerReplicasCalculator calculates a number of target and expected replicas for a given controller.
type controllerReplicasCalculator interface {
	getReplicas(metav1.OwnerReference, string) (*replicasInfo, error)
}

type replicasInfo struct {
	targetReplicas, currentReplicas int32
}

// Planner is responsible for deciding which nodes should be deleted during scale down.
type Planner struct {
	context               *context.AutoscalingContext
	unremovableNodes      *unremovable.Nodes
	unneededNodes         *unneeded.Nodes
	rs                    removalSimulator
	actuationInjector     *scheduling.HintingSimulator
	latestUpdate          time.Time
	minUpdateInterval     time.Duration
	eligibilityChecker    eligibilityChecker
	nodeUtilizationMap    map[string]utilization.Info
	actuationStatus       scaledown.ActuationStatus
	resourceLimitsFinder  *resource.LimitsFinder
	cc                    controllerReplicasCalculator
	scaleDownSetProcessor nodes.ScaleDownSetProcessor
}

// New creates a new Planner object.
func New(context *context.AutoscalingContext, processors *processors.AutoscalingProcessors, deleteOptions options.NodeDeleteOptions, drainabilityRules rules.Rules) *Planner {
	resourceLimitsFinder := resource.NewLimitsFinder(processors.CustomResourcesProcessor)
	minUpdateInterval := context.AutoscalingOptions.NodeGroupDefaults.ScaleDownUnneededTime
	if minUpdateInterval == 0*time.Nanosecond {
		minUpdateInterval = 1 * time.Nanosecond
	}
	return &Planner{
		context:               context,
		unremovableNodes:      unremovable.NewNodes(),
		unneededNodes:         unneeded.NewNodes(processors.NodeGroupConfigProcessor, resourceLimitsFinder),
		rs:                    simulator.NewRemovalSimulator(context.ListerRegistry, context.ClusterSnapshot, context.PredicateChecker, deleteOptions, drainabilityRules, true),
		actuationInjector:     scheduling.NewHintingSimulator(context.PredicateChecker),
		eligibilityChecker:    eligibility.NewChecker(processors.NodeGroupConfigProcessor),
		nodeUtilizationMap:    make(map[string]utilization.Info),
		resourceLimitsFinder:  resourceLimitsFinder,
		cc:                    newControllerReplicasCalculator(context.ListerRegistry),
		scaleDownSetProcessor: processors.ScaleDownSetProcessor,
		minUpdateInterval:     minUpdateInterval,
	}
}

// UpdateClusterState needs to be periodically invoked to provide Planner with
// up-to-date information about the cluster.
// Planner will evaluate scaleDownCandidates in the order provided here.
func (p *Planner) UpdateClusterState(podDestinations, scaleDownCandidates []*apiv1.Node, as scaledown.ActuationStatus, currentTime time.Time) errors.AutoscalerError {
	updateInterval := currentTime.Sub(p.latestUpdate)
	if updateInterval < p.minUpdateInterval {
		p.minUpdateInterval = updateInterval
	}
	p.latestUpdate = currentTime
	p.actuationStatus = as
	// Avoid persisting changes done by the simulation.
	p.context.ClusterSnapshot.Fork()
	defer p.context.ClusterSnapshot.Revert()
	err := p.injectRecentlyEvictedPods()
	if err != nil {
		klog.Warningf("Not all recently evicted pods could be injected")
	}
	deletions := asMap(merged(as.DeletionsInProgress()))
	podDestinations = filterOutOngoingDeletions(podDestinations, deletions)
	scaleDownCandidates = filterOutOngoingDeletions(scaleDownCandidates, deletions)
	p.categorizeNodes(asMap(nodeNames(podDestinations)), scaleDownCandidates)
	p.rs.DropOldHints()
	p.actuationInjector.DropOldHints()
	return nil
}

// CleanUpUnneededNodes forces Planner to forget about all nodes considered
// unneeded so far.
func (p *Planner) CleanUpUnneededNodes() {
	p.unneededNodes.Clear()
}

// NodesToDelete returns all Nodes that could be removed right now, according
// to the Planner.
func (p *Planner) NodesToDelete(_ time.Time) (empty, needDrain []*apiv1.Node) {
	empty, needDrain = []*apiv1.Node{}, []*apiv1.Node{}
	nodes, err := allNodes(p.context.ClusterSnapshot)
	if err != nil {
		klog.Errorf("Nothing will scale down, failed to list nodes from ClusterSnapshot: %v", err)
		return nil, nil
	}
	resourceLimiter, err := p.context.CloudProvider.GetResourceLimiter()
	if err != nil {
		klog.Errorf("Nothing will scale down, failed to create resource limiter: %v", err)
		return nil, nil
	}
	limitsLeft := p.resourceLimitsFinder.LimitsLeft(p.context, nodes, resourceLimiter, p.latestUpdate)
	emptyRemovable, needDrainRemovable, unremovable := p.unneededNodes.RemovableAt(p.context, p.latestUpdate, limitsLeft, resourceLimiter.GetResources(), p.actuationStatus)
	for _, u := range unremovable {
		p.unremovableNodes.Add(u)
	}
	needDrainRemovable = sortByRisk(needDrainRemovable)
	nodesToRemove := p.scaleDownSetProcessor.GetNodesToRemove(
		p.context,
		// We need to pass empty nodes first, as there might be some non-empty scale
		// downs already in progress. If we pass the empty nodes first, they will be first
		// to get deleted, thus we decrease chances of hitting the limit on non-empty scale down.
		append(emptyRemovable, needDrainRemovable...),
		// No need to limit the number of nodes, since it will happen later, in the actuation stage.
		// It will make a more appropriate decision by using additional information about deletions
		// in progress.
		math.MaxInt)
	for _, nodeToRemove := range nodesToRemove {
		if len(nodeToRemove.PodsToReschedule) > 0 {
			needDrain = append(needDrain, nodeToRemove.Node)
		} else {
			empty = append(empty, nodeToRemove.Node)
		}
	}

	return empty, needDrain
}

func allNodes(s clustersnapshot.ClusterSnapshot) ([]*apiv1.Node, error) {
	nodeInfos, err := s.ListNodeInfos()
	if err != nil {
		// This should never happen, List() returns err only because scheduler interface requires it.
		return nil, err
	}
	nodes := make([]*apiv1.Node, len(nodeInfos))
	for i, ni := range nodeInfos {
		nodes[i] = ni.Node()
	}
	return nodes, nil
}

// UnneededNodes returns a list of nodes currently considered as unneeded.
func (p *Planner) UnneededNodes() []*apiv1.Node {
	return p.unneededNodes.AsList()
}

// UnremovableNodes returns a list of nodes currently considered as unremovable.
func (p *Planner) UnremovableNodes() []*simulator.UnremovableNode {
	return p.unremovableNodes.AsList()
}

// NodeUtilizationMap returns a map with utilization of nodes.
func (p *Planner) NodeUtilizationMap() map[string]utilization.Info {
	return p.nodeUtilizationMap
}

// injectRecentlyEvictedPods injects pods into ClusterSnapshot, to allow
// subsequent simulation to anticipate which pods will end up getting replaced
// due to being evicted by previous scale down(s). This function injects pods
// which were recently evicted (it is up to ActuationStatus to decide what
// "recently" means in this case). The existing pods from currently drained
// nodes are already added before scale-up to optimize scale-up latency.
//
// For pods that are controlled by controller known by CA, it will check whether
// they have been recreated and will inject only not yet recreated pods.
func (p *Planner) injectRecentlyEvictedPods() error {
	recentlyEvictedRecreatablePods := pod_util.FilterRecreatablePods(p.actuationStatus.RecentEvictions())
	return p.injectPods(filterOutRecreatedPods(recentlyEvictedRecreatablePods, p.cc))
}

func filterOutRecreatedPods(pods []*apiv1.Pod, cc controllerReplicasCalculator) []*apiv1.Pod {
	var podsToInject []*apiv1.Pod
	addedReplicas := make(map[string]int32)
	for _, pod := range pods {
		ownerRef := getKnownOwnerRef(pod.GetOwnerReferences())
		// in case of unknown ownerRef (i.e. not recognized by CA) we still inject
		// the pod, to be on the safe side in case there is some custom controller
		// that will recreate the pod.
		if ownerRef == nil {
			podsToInject = append(podsToInject, pod)
			continue
		}
		rep, err := cc.getReplicas(*ownerRef, pod.Namespace)
		if err != nil {
			podsToInject = append(podsToInject, pod)
			continue
		}
		ownerUID := string(ownerRef.UID)
		if rep.targetReplicas > rep.currentReplicas && addedReplicas[ownerUID] < rep.targetReplicas-rep.currentReplicas {
			podsToInject = append(podsToInject, pod)
			addedReplicas[ownerUID] += 1
		}
	}
	return podsToInject
}

func (p *Planner) injectPods(pods []*apiv1.Pod) error {
	pods = pod_util.ClearPodNodeNames(pods)
	// Note: We're using ScheduleAnywhere, but the pods won't schedule back
	// on the drained nodes due to taints.
	statuses, _, err := p.actuationInjector.TrySchedulePods(p.context.ClusterSnapshot, pods, scheduling.ScheduleAnywhere, true)
	if err != nil {
		return fmt.Errorf("cannot scale down, an unexpected error occurred: %v", err)
	}
	if len(statuses) != len(pods) {
		return fmt.Errorf("can reschedule only %d out of %d pods from ongoing deletions", len(statuses), len(pods))
	}
	return nil
}

// categorizeNodes determines, for each node, whether it can be eventually
// removed or if there are reasons preventing that.
func (p *Planner) categorizeNodes(podDestinations map[string]bool, scaleDownCandidates []*apiv1.Node) {
	unremovableTimeout := p.latestUpdate.Add(p.context.AutoscalingOptions.UnremovableNodeRecheckTimeout)
	unremovableCount := 0
	var removableList []simulator.NodeToBeRemoved
	atomicScaleDownNodesCount := 0
	p.unremovableNodes.Update(p.context.ClusterSnapshot, p.latestUpdate)
	currentlyUnneededNodeNames, utilizationMap, ineligible := p.eligibilityChecker.FilterOutUnremovable(p.context, scaleDownCandidates, p.latestUpdate, p.unremovableNodes)
	for _, n := range ineligible {
		p.unremovableNodes.Add(n)
	}
	p.nodeUtilizationMap = utilizationMap
	timer := time.NewTimer(p.context.ScaleDownSimulationTimeout)

	for i, node := range currentlyUnneededNodeNames {
		if timedOut(timer) {
			klog.Warningf("%d out of %d nodes skipped in scale down simulation due to timeout.", len(currentlyUnneededNodeNames)-i, len(currentlyUnneededNodeNames))
			break
		}
		if len(removableList)-atomicScaleDownNodesCount >= p.unneededNodesLimit() {
			klog.V(4).Infof("%d out of %d nodes skipped in scale down simulation: there are already %d unneeded nodes so no point in looking for more. Total atomic scale down nodes: %d", len(currentlyUnneededNodeNames)-i, len(currentlyUnneededNodeNames), len(removableList), atomicScaleDownNodesCount)
			break
		}
		removable, unremovable := p.rs.SimulateNodeRemoval(node, podDestinations, p.latestUpdate, p.context.RemainingPdbTracker)
		if removable != nil {
			_, inParallel, _ := p.context.RemainingPdbTracker.CanRemovePods(removable.PodsToReschedule)
			if !inParallel {
				removable.IsRisky = true
			}
			delete(podDestinations, removable.Node.Name)
			p.context.RemainingPdbTracker.RemovePods(removable.PodsToReschedule)
			removableList = append(removableList, *removable)
			if p.atomicScaleDownNode(removable) {
				atomicScaleDownNodesCount++
				klog.V(2).Infof("Considering node %s for atomic scale down. Total atomic scale down nodes count: %d", removable.Node.Name, atomicScaleDownNodesCount)
			}
		}
		if unremovable != nil {
			unremovableCount += 1
			p.unremovableNodes.AddTimeout(unremovable, unremovableTimeout)
		}
	}
	p.unneededNodes.Update(removableList, p.latestUpdate)
	if unremovableCount > 0 {
		klog.V(1).Infof("%v nodes found to be unremovable in simulation, will re-check them at %v", unremovableCount, unremovableTimeout)
	}
}

// atomicScaleDownNode checks if the removable node would be considered for atomic scale down.
func (p *Planner) atomicScaleDownNode(node *simulator.NodeToBeRemoved) bool {
	nodeGroup, err := p.context.CloudProvider.NodeGroupForNode(node.Node)
	if err != nil {
		klog.Errorf("failed to get node info for %v: %s", node.Node.Name, err)
		return false
	}
	autoscalingOptions, err := nodeGroup.GetOptions(p.context.NodeGroupDefaults)
	if err != nil && err != cloudprovider.ErrNotImplemented {
		klog.Errorf("Failed to get autoscaling options for node group %s: %v", nodeGroup.Id(), err)
		return false
	}
	if autoscalingOptions != nil && autoscalingOptions.ZeroOrMaxNodeScaling {
		return true
	}
	return false
}

// unneededNodesLimit returns the number of nodes after which calculating more
// unneeded nodes is a waste of time. The reasoning behind it is essentially as
// follows.
// If the nodes are being removed instantly, then during each iteration we're
// going to delete up to MaxScaleDownParallelism nodes. Therefore, it doesn't
// really make sense to add more unneeded nodes than that.
// Let N = MaxScaleDownParallelism. When there are no unneeded nodes, we only
// need to find N of them in the first iteration. Once the unneeded time
// accumulates for them, only up to N will get deleted in a single iteration.
// When there are >0 unneeded nodes, we only need to add N more: once the first
// N will be deleted, we'll need another iteration for the next N nodes to get
// deleted.
// Of course, a node may stop being unneeded at any given time. To prevent
// slowdown stemming from having too little unneeded nodes, we're adding an
// extra buffer of N nodes. Note that we don't have to be super precise about
// the buffer size - if it is too small, we'll simply remove less than N nodes
// in one iteration.
// Finally, we know that in practice nodes are not removed instantly,
// especially when they require draining, so incrementing the limit by N every
// loop may in practice lead the limit to increase too much after a number of
// loops. To help with that, we can put another, not incremental upper bound on
// the limit: with max unneded time U and loop interval I, we're going to have
// up to U/I loops before a node is removed. This means that the total number
// of unneeded nodes shouldn't really exceed N*U/I - scale down will not be
// able to keep up with removing them anyway.
func (p *Planner) unneededNodesLimit() int {
	n := p.context.AutoscalingOptions.MaxScaleDownParallelism
	extraBuffer := n
	limit := len(p.unneededNodes.AsList()) + n + extraBuffer
	// TODO(x13n): Use moving average instead of min.
	loopInterval := int64(p.minUpdateInterval)
	u := int64(p.context.AutoscalingOptions.NodeGroupDefaults.ScaleDownUnneededTime)
	if u < loopInterval {
		u = loopInterval
	}
	upperBound := n*int(u/loopInterval) + extraBuffer
	if upperBound < limit {
		return upperBound
	}
	return limit
}

// getKnownOwnerRef returns ownerRef that is known by CA and CA knows the logic of how this controller recreates pods.
func getKnownOwnerRef(ownerRefs []metav1.OwnerReference) *metav1.OwnerReference {
	for _, ownerRef := range ownerRefs {
		switch ownerRef.Kind {
		case "StatefulSet", "Job", "ReplicaSet", "ReplicationController":
			return &ownerRef
		}
	}
	return nil
}

func merged(a, b []string) []string {
	return append(append(make([]string, 0, len(a)+len(b)), a...), b...)
}

func asMap(strs []string) map[string]bool {
	m := make(map[string]bool, len(strs))
	for _, s := range strs {
		m[s] = true
	}
	return m
}

func nodeNames(nodes []*apiv1.Node) []string {
	names := make([]string, len(nodes))
	for i, node := range nodes {
		names[i] = node.Name
	}
	return names
}

func filterOutOngoingDeletions(ns []*apiv1.Node, deleted map[string]bool) []*apiv1.Node {
	rv := make([]*apiv1.Node, 0, len(ns))
	for _, n := range ns {
		if deleted[n.Name] {
			continue
		}
		rv = append(rv, n)
	}
	return rv
}

func sortByRisk(nodes []simulator.NodeToBeRemoved) []simulator.NodeToBeRemoved {
	riskyNodes := []simulator.NodeToBeRemoved{}
	okNodes := []simulator.NodeToBeRemoved{}
	for _, nodeToRemove := range nodes {
		if nodeToRemove.IsRisky {
			riskyNodes = append(riskyNodes, nodeToRemove)
		} else {
			okNodes = append(okNodes, nodeToRemove)
		}
	}
	return append(okNodes, riskyNodes...)
}

func timedOut(timer *time.Timer) bool {
	select {
	case <-timer.C:
		return true
	default:
		return false
	}
}
