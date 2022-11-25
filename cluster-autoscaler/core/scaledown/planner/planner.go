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
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/eligibility"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/resource"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/unneeded"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/unremovable"
	"k8s.io/autoscaler/cluster-autoscaler/processors"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/scheduling"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/utilization"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	pod_util "k8s.io/autoscaler/cluster-autoscaler/utils/pod"

	apiv1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	klog "k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/scheduler/framework"
)

type eligibilityChecker interface {
	FilterOutUnremovable(context *context.AutoscalingContext, scaleDownCandidates []*apiv1.Node, timestamp time.Time, unremovableNodes *unremovable.Nodes) ([]string, map[string]utilization.Info, []*simulator.UnremovableNode)
}

type removalSimulator interface {
	DropOldHints()
	SimulateNodeRemoval(node string, podDestinations map[string]bool, timestamp time.Time, pdbs []*policyv1.PodDisruptionBudget) (*simulator.NodeToBeRemoved, *simulator.UnremovableNode)
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
	context              *context.AutoscalingContext
	unremovableNodes     *unremovable.Nodes
	unneededNodes        *unneeded.Nodes
	rs                   removalSimulator
	actuationInjector    *scheduling.HintingSimulator
	latestUpdate         time.Time
	eligibilityChecker   eligibilityChecker
	nodeUtilizationMap   map[string]utilization.Info
	actuationStatus      scaledown.ActuationStatus
	resourceLimitsFinder *resource.LimitsFinder
	cc                   controllerReplicasCalculator
}

// New creates a new Planner object.
func New(context *context.AutoscalingContext, processors *processors.AutoscalingProcessors, deleteOptions simulator.NodeDeleteOptions) *Planner {
	resourceLimitsFinder := resource.NewLimitsFinder(processors.CustomResourcesProcessor)
	return &Planner{
		context:              context,
		unremovableNodes:     unremovable.NewNodes(),
		unneededNodes:        unneeded.NewNodes(processors.NodeGroupConfigProcessor, resourceLimitsFinder),
		rs:                   simulator.NewRemovalSimulator(context.ListerRegistry, context.ClusterSnapshot, context.PredicateChecker, simulator.NewUsageTracker(), deleteOptions, true),
		actuationInjector:    scheduling.NewHintingSimulator(context.PredicateChecker),
		eligibilityChecker:   eligibility.NewChecker(processors.NodeGroupConfigProcessor),
		nodeUtilizationMap:   make(map[string]utilization.Info),
		resourceLimitsFinder: resourceLimitsFinder,
		cc:                   newControllerReplicasCalculator(context.ListerRegistry),
	}
}

// UpdateClusterState needs to be periodically invoked to provide Planner with
// up-to-date information about the cluster.
// Planner will evaluate scaleDownCandidates in the order provided here.
func (p *Planner) UpdateClusterState(podDestinations, scaleDownCandidates []*apiv1.Node, as scaledown.ActuationStatus, pdb []*policyv1.PodDisruptionBudget, currentTime time.Time) errors.AutoscalerError {
	p.latestUpdate = currentTime
	p.actuationStatus = as
	// Avoid persisting changes done by the simulation.
	p.context.ClusterSnapshot.Fork()
	defer p.context.ClusterSnapshot.Revert()
	err := p.injectOngoingActuation()
	if err != nil {
		p.CleanUpUnneededNodes()
		return errors.ToAutoscalerError(errors.UnexpectedScaleDownStateError, err)
	}
	deletions := asMap(merged(as.DeletionsInProgress()))
	podDestinations = filterOutOngoingDeletions(podDestinations, deletions)
	scaleDownCandidates = filterOutOngoingDeletions(scaleDownCandidates, deletions)
	p.categorizeNodes(asMap(nodeNames(podDestinations)), scaleDownCandidates, pdb)
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
func (p *Planner) NodesToDelete() (empty, needDrain []*apiv1.Node) {
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
	empty, needDrain, unremovable := p.unneededNodes.RemovableAt(p.context, p.latestUpdate, limitsLeft, resourceLimiter.GetResources(), p.actuationStatus)
	for _, u := range unremovable {
		p.unremovableNodes.Add(u)
	}
	// TODO: filter results with ScaleDownSetProcessor.GetNodesToRemove
	return empty, needDrain
}

func allNodes(s clustersnapshot.ClusterSnapshot) ([]*apiv1.Node, error) {
	nodeInfos, err := s.NodeInfos().List()
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

// injectOngoingActuation injects pods into ClusterSnapshot, to allow
// subsequent simulation to anticipate which pods will end up getting replaced
// due to being evicted by previous scale down(s). There are two sets of such
// pods:
//   - existing pods from currently drained nodes
//   - pods which were recently evicted (it is up to ActuationStatus to decide
//     what "recently" means in this case).
//
// For pods that are controlled by controller known by CA, it will check whether
// they have been recreated and will inject only not yet recreated pods.
func (p *Planner) injectOngoingActuation() error {
	currentlyDrainedRecreatablePods := filterRecreatable(currentlyDrainedPods(p.context.ClusterSnapshot.NodeInfos(), p.actuationStatus))
	recentlyEvictedRecreatablePods := filterRecreatable(p.actuationStatus.RecentEvictions())
	err := p.injectPods(currentlyDrainedRecreatablePods)
	if err != nil {
		return err
	}
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

func currentlyDrainedPods(niLister framework.NodeInfoLister, as scaledown.ActuationStatus) []*apiv1.Pod {
	var pods []*apiv1.Pod
	_, ds := as.DeletionsInProgress()
	for _, d := range ds {
		ni, err := niLister.Get(d)
		if err != nil {
			klog.Warningf("Couldn't get node %v info, assuming the node got deleted already: %v", d, err)
			continue
		}
		for _, pi := range ni.Pods {
			pods = append(pods, pi.Pod)
		}
	}
	return pods
}

func filterRecreatable(pods []*apiv1.Pod) []*apiv1.Pod {
	filtered := make([]*apiv1.Pod, 0, len(pods))
	for _, p := range pods {
		if pod_util.IsStaticPod(p) || pod_util.IsMirrorPod(p) || pod_util.IsDaemonSetPod(p) {
			continue
		}
		filtered = append(filtered, p)
	}
	return filtered
}

func (p *Planner) injectPods(pods []*apiv1.Pod) error {
	pods = clearNodeName(pods)
	// Note: We're using ScheduleAnywhere, but the pods won't schedule back
	// on the drained nodes due to taints.
	statuses, _, err := p.actuationInjector.TrySchedulePods(p.context.ClusterSnapshot, pods, scheduling.ScheduleAnywhere, true)
	if err != nil {
		return fmt.Errorf("cannot scale down, an unexpected error occurred: %v", err)
	}
	if len(statuses) != len(pods) {
		return fmt.Errorf("cannot scale down, can reschedule only %d out of %d pods from ongoing deletions", len(statuses), len(pods))
	}
	return nil
}

// categorizeNodes determines, for each node, whether it can be eventually
// removed or if there are reasons preventing that.
// TODO: Track remaining PDB budget.
func (p *Planner) categorizeNodes(podDestinations map[string]bool, scaleDownCandidates []*apiv1.Node, pdbs []*policyv1.PodDisruptionBudget) {
	unremovableTimeout := p.latestUpdate.Add(p.context.AutoscalingOptions.UnremovableNodeRecheckTimeout)
	unremovableCount := 0
	var removableList []simulator.NodeToBeRemoved
	p.unremovableNodes.Update(p.context.ClusterSnapshot.NodeInfos(), p.latestUpdate)
	currentlyUnneededNodeNames, utilizationMap, ineligible := p.eligibilityChecker.FilterOutUnremovable(p.context, scaleDownCandidates, p.latestUpdate, p.unremovableNodes)
	for _, n := range ineligible {
		p.unremovableNodes.Add(n)
	}
	p.nodeUtilizationMap = utilizationMap
	for _, node := range currentlyUnneededNodeNames {
		// TODO(x13n): break on timeout. Figure out how to handle nodes
		// identified as unneeded in previous iteration, but now
		// skipped due to timeout.
		removable, unremovable := p.rs.SimulateNodeRemoval(node, podDestinations, p.latestUpdate, pdbs)
		if unremovable != nil {
			unremovableCount += 1
			p.unremovableNodes.AddTimeout(unremovable, unremovableTimeout)
		}
		if removable != nil {
			delete(podDestinations, removable.Node.Name)
			removableList = append(removableList, *removable)
		}
	}
	p.unneededNodes.Update(removableList, p.latestUpdate)
	if unremovableCount > 0 {
		klog.V(1).Infof("%v nodes found to be unremovable in simulation, will re-check them at %v", unremovableCount, unremovableTimeout)
	}
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

func clearNodeName(pods []*apiv1.Pod) []*apiv1.Pod {
	newpods := make([]*apiv1.Pod, 0, len(pods))
	for _, podptr := range pods {
		newpod := *podptr
		newpod.Spec.NodeName = ""
		newpods = append(newpods, &newpod)
	}
	return newpods
}
