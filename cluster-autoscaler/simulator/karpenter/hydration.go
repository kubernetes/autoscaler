/*
Copyright 2026 The Kubernetes Authors.

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

package karpenter

import (
	"context"
	"slices"
	"strings"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/utils/annotations"
	podutils "k8s.io/autoscaler/cluster-autoscaler/utils/pod"
)

var zeroQuantity = resource.Quantity{}

// HydrateClusterState filters relevant nodes and pods from the ClusterSnapshot.
// It performs Topology Pruning and Resource Clamping to reduce the number of pods and nodes Karpenter needs to evaluate.
func HydrateClusterState(ctx context.Context, snapshot clustersnapshot.ClusterSnapshot, nodeInfos []*framework.NodeInfo, podsToSchedule []*apiv1.Pod) ([]*apiv1.Pod, []*apiv1.Node, error) {
	pendingSelectors := []labels.Selector{}
	pendingPodLabelSets := []labels.Set{}
	tscTopologyKeys := sets.New[string]()

	// 1. Identify relevant selectors and label sets from pods to schedule
	for _, p := range podsToSchedule {
		if p == nil {
			continue
		}
		pendingPodLabelSets = append(pendingPodLabelSets, labels.Set(p.Labels))
		if p.Spec.Affinity != nil {
			if p.Spec.Affinity.PodAffinity != nil {
				for _, term := range p.Spec.Affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution {
					sel, err := metav1.LabelSelectorAsSelector(term.LabelSelector)
					if err == nil {
						pendingSelectors = append(pendingSelectors, sel)
					}
				}
			}
			if p.Spec.Affinity.PodAntiAffinity != nil {
				for _, term := range p.Spec.Affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution {
					sel, err := metav1.LabelSelectorAsSelector(term.LabelSelector)
					if err == nil {
						pendingSelectors = append(pendingSelectors, sel)
					}
				}
			}
		}
		for _, tsc := range p.Spec.TopologySpreadConstraints {
			sel, err := metav1.LabelSelectorAsSelector(tsc.LabelSelector)
			if err == nil {
				pendingSelectors = append(pendingSelectors, sel)
			}
			if tsc.TopologyKey != "" {
				tscTopologyKeys.Insert(tsc.TopologyKey)
			}
		}
	}

	// 2. Identify TSC domain representative nodes.
	// WHY: TopologySpreadConstraints split candidate nodes into topological domains (e.g., topology.kubernetes.io/zone).
	// Karpenter's topology algorithm requires at least one node per active topology domain to be present in state.Cluster
	// in order to calculate existing domain counts correctly. Preserving one representative node per domain ensures
	// domain counts remain accurate even if all other nodes in that domain are empty or pruned.
	domainRepNodes := make(map[string]bool)
	isDomainRep := make(map[string]bool)
	for _, ni := range nodeInfos {
		if ni == nil {
			continue
		}
		node := ni.Node()
		if node == nil {
			continue
		}
		for topKey := range tscTopologyKeys {
			val := node.Labels[topKey]
			domId := topKey + "=" + val
			if !domainRepNodes[domId] {
				domainRepNodes[domId] = true
				isDomainRep[node.Name] = true
			}
		}
	}

	scheduledPodKeys := make(map[string]bool)
	scheduledPodUIDs := make(map[types.UID]bool)
	scheduledPodPtrs := make(map[*apiv1.Pod]bool)
	for _, p := range podsToSchedule {
		if p == nil {
			continue
		}
		if p.UID != "" {
			scheduledPodUIDs[p.UID] = true
		}
		if p.Name != "" {
			scheduledPodKeys[p.Namespace+"/"+p.Name] = true
		}
		scheduledPodPtrs[p] = true
	}

	isPodToSchedule := func(p *apiv1.Pod) bool {
		if p == nil {
			return false
		}
		if scheduledPodPtrs[p] {
			return true
		}
		if p.UID != "" && scheduledPodUIDs[p.UID] {
			return true
		}
		if p.Name != "" && scheduledPodKeys[p.Namespace+"/"+p.Name] {
			return true
		}
		return false
	}

	uniquePendingReqs := getUniquePendingPodRequests(podsToSchedule)

	var allRelevantPods []*apiv1.Pod
	var allRelevantNodes []*apiv1.Node

	// 3. Filter relevant pods & clamp allocatable resources for pruned pods
	for _, ni := range nodeInfos {
		if ni == nil {
			continue
		}
		node := ni.Node()
		if node == nil {
			continue
		}

		var keptPodsOnNode []*apiv1.Pod
		prunedPodRequests := make(apiv1.ResourceList)

		for _, pi := range ni.Pods() {
			pod := pi.Pod
			if pod == nil {
				continue
			}

			// WHY (Exclude Pending Batch Pods from Background State):
			// If a pod on a snapshot node is one of the pods in podsToSchedule (e.g. from an earlier speculative or
			// iterative simulation step), counting it as an existing background pod in state.Cluster would double-count
			// its resource consumption and occupy its topology spread / anti-affinity slot, preventing Karpenter's solver
			// from placing the pod during the active scheduling simulation.
			if isPodToSchedule(pod) {
				continue
			}

			isRelevant := false

			// Exemption check: HostPorts or CSI Volumes are never pruned.
			// WHY: HostPort pods enforce node-level network port binding constraints, and CSI volume pods enforce
			// node-level PV/PVC attachments. Omitting these pods would cause Karpenter to ignore port conflicts or
			// storage attachments, resulting in invalid pod scheduling decisions.
			if hasHostPort(pod) || hasCSIVolume(pod) {
				isRelevant = true
			}

			// Pending Selector Match: running pod's labels match any pending pod selector
			if !isRelevant {
				podLabelSet := labels.Set(pod.Labels)
				for _, sel := range pendingSelectors {
					if sel.Matches(podLabelSet) {
						isRelevant = true
						break
					}
				}
			}

			// Running Selector Match: running pod's AntiAffinity/TSC selectors match any pending pod label set
			if !isRelevant {
				runningSelectors := getRunningPodSelectors(pod)
				for _, sel := range runningSelectors {
					for _, pendingLabelSet := range pendingPodLabelSets {
						if sel.Matches(pendingLabelSet) {
							isRelevant = true
							break
						}
					}
					if isRelevant {
						break
					}
				}
			}

			if isRelevant {
				keptPodsOnNode = append(keptPodsOnNode, pod)
			} else if !podutils.IsDaemonSetPod(pod) {
				// Sum requests of pruned non-DaemonSet pod for resource clamping.
				// DaemonSet pods are tracked separately by Karpenter's scheduler via daemonSetPods
				// in calculateExistingNodeClaims. Clamping DaemonSet requests here would double-subtract overhead.
				addResourceRequests(prunedPodRequests, podutils.PodRequests(pod))
			}
		}

		clampedNode := normalizeNode(node, prunedPodRequests)

		// Determine if the node should be retained
		hasSpace := hasRemainingCapacity(clampedNode, keptPodsOnNode, uniquePendingReqs)
		if len(keptPodsOnNode) > 0 || isDomainRep[node.Name] || hasSpace {
			allRelevantNodes = append(allRelevantNodes, clampedNode)
			allRelevantPods = append(allRelevantPods, keptPodsOnNode...)
		}
	}

	return allRelevantPods, allRelevantNodes, nil
}

func getRunningPodSelectors(pod *apiv1.Pod) []labels.Selector {
	var selectors []labels.Selector
	if pod == nil {
		return selectors
	}
	if pod.Spec.Affinity != nil && pod.Spec.Affinity.PodAntiAffinity != nil {
		for _, term := range pod.Spec.Affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution {
			sel, err := metav1.LabelSelectorAsSelector(term.LabelSelector)
			if err == nil {
				selectors = append(selectors, sel)
			}
		}
	}
	for _, tsc := range pod.Spec.TopologySpreadConstraints {
		sel, err := metav1.LabelSelectorAsSelector(tsc.LabelSelector)
		if err == nil {
			selectors = append(selectors, sel)
		}
	}
	return selectors
}

func hasHostPort(pod *apiv1.Pod) bool {
	if pod == nil {
		return false
	}
	for _, c := range pod.Spec.Containers {
		for _, p := range c.Ports {
			if p.HostPort > 0 {
				return true
			}
		}
	}
	for _, c := range pod.Spec.InitContainers {
		for _, p := range c.Ports {
			if p.HostPort > 0 {
				return true
			}
		}
	}
	return false
}

func hasCSIVolume(pod *apiv1.Pod) bool {
	if pod == nil {
		return false
	}
	for _, v := range pod.Spec.Volumes {
		if v.PersistentVolumeClaim != nil || v.CSI != nil {
			return true
		}
	}
	return false
}

func addResourceRequests(dst apiv1.ResourceList, reqs apiv1.ResourceList) {
	for rName, rQty := range reqs {
		cur := dst[rName]
		cur.Add(rQty)
		dst[rName] = cur
	}
}

// clampAllocatable subtracts pruned pod resource requests from node Allocatable capacity.
// WHY: Deleting pruned pods from Karpenter's simulation state without adjusting node allocatable capacity
// would artificially free up resources. Clamping Allocatable capacity down by the exact sum of pruned pod requests
// allows us to omit non-interfering pods from Karpenter's snapshot pod list (saving memory and scheduling overhead)
// while mathematically guaranteeing that residual node allocatable capacity remains identical to reality.
func clampAllocatable(allocatable apiv1.ResourceList, prunedReqs apiv1.ResourceList) {
	for rName, rQty := range prunedReqs {
		allocQty, ok := allocatable[rName]
		if !ok {
			continue
		}
		if allocQty.Cmp(rQty) <= 0 {
			allocatable[rName] = zeroQuantity.DeepCopy()
		} else {
			allocQty.Sub(rQty)
			allocatable[rName] = allocQty
		}
	}
}

func getUniquePendingPodRequests(pendingPods []*apiv1.Pod) []apiv1.ResourceList {
	seen := make(map[string]apiv1.ResourceList)
	for _, p := range pendingPods {
		if p == nil {
			continue
		}
		req := podutils.PodRequests(p)
		if _, ok := req[apiv1.ResourcePods]; !ok {
			req[apiv1.ResourcePods] = *resource.NewQuantity(1, resource.DecimalSI)
		}
		key := canonicalResourceListKey(req)
		if _, exists := seen[key]; !exists {
			seen[key] = req
		}
	}
	raw := make([]apiv1.ResourceList, 0, len(seen))
	for _, req := range seen {
		raw = append(raw, req)
	}
	return filterMinimalPodShapes(raw)
}

// maxMinimalPodShapes bounds the size of the Pareto frontier retained for node retention checks.
// WHY: Under pathological workloads with many mutually incomparable resource vectors (e.g. CPU vs Memory vs GPU trade-offs),
// the Pareto frontier |S_min| can theoretically grow up to the number of unique pending pod shapes M.
// Capping |S_min| at 10 bounds the per-node remaining capacity evaluation to O(10 * N) = O(N) comparisons.
const maxMinimalPodShapes = 10

// filterMinimalPodShapes reduces candidate pending pod request vectors to a minimal Pareto frontier S_min.
// WHY: A pod shape V_A (larger) is dominated by shape V_B (smaller) if V_B <= V_A across all resource dimensions.
// Dominated shapes are mathematically redundant for node capacity retention checks: if a node cannot fit V_B, it
// cannot fit V_A; conversely, if a node fits V_A, it already fits V_B. Retaining only non-dominated shapes minimizes
// the set of capacity checks per node while maintaining complete coverage.
//
// BOUNDED FALLBACK RATIONALE: If the number of Pareto minimal shapes exceeds maxMinimalPodShapes (10), S_min collapses
// to a single Global Minimum vector (GlobalMin[R] = min_p req(p, R)). This guarantees strict worst-case linear time
// complexity O(N) under high-cardinality multi-resource workloads, sacrificing slight pruning precision to prevent CA latency spikes.
func filterMinimalPodShapes(shapes []apiv1.ResourceList) []apiv1.ResourceList {
	minimal := make([]apiv1.ResourceList, 0, len(shapes))
	for i, s1 := range shapes {
		dominated := false
		for j, s2 := range shapes {
			if i != j && isPodShapeDominated(s1, s2) {
				dominated = true
				break
			}
		}
		if !dominated {
			minimal = append(minimal, s1)
		}
	}
	// Pathological workload protection: collapse to single Global Minimum vector if minimal shapes exceed limit
	if len(minimal) > maxMinimalPodShapes {
		globalMin := make(apiv1.ResourceList)
		for _, s := range minimal {
			for rName, qty := range s {
				cur, ok := globalMin[rName]
				if !ok || qty.Cmp(cur) < 0 {
					globalMin[rName] = qty.DeepCopy()
				}
			}
		}
		return []apiv1.ResourceList{globalMin}
	}
	return minimal
}

// isPodShapeDominated checks whether s1 is dominated by s2 (i.e. s2 requires <= s1 for all resource dimensions).
// WHY: If s2 <= s1 across CPU, memory, pods, and extended resources, any node that can fit s1 is guaranteed to fit s2.
// Therefore, s1 is redundant and can be omitted from candidate pod shape checks.
func isPodShapeDominated(s1, s2 apiv1.ResourceList) bool {
	// s1 is dominated by s2 if s2 <= s1 across ALL resource dimensions.
	// If s2 requires strictly more of ANY resource than s1, then s2 does not dominate s1.
	for rName, qty2 := range s2 {
		qty1 := s1[rName]
		if qty2.Cmp(qty1) > 0 {
			return false
		}
	}
	return true
}

func canonicalResourceListKey(req apiv1.ResourceList) string {
	keys := make([]string, 0, len(req))
	for k := range req {
		keys = append(keys, string(k))
	}
	slices.Sort(keys)
	var sb strings.Builder
	for _, k := range keys {
		qty := req[apiv1.ResourceName(k)]
		sb.WriteString(k)
		sb.WriteString(":")
		sb.WriteString(qty.String())
		sb.WriteString(";")
	}
	return sb.String()
}

func hasRemainingCapacity(node *apiv1.Node, keptPods []*apiv1.Pod, uniquePendingReqs []apiv1.ResourceList) bool {
	if node == nil || node.Status.Allocatable == nil {
		return false
	}
	remCap := node.Status.Allocatable.DeepCopy()
	for _, p := range keptPods {
		if p != nil {
			clampAllocatable(remCap, podutils.PodRequests(p))
		}
	}

	for _, req := range uniquePendingReqs {
		fits := true
		for rName, minQty := range req {
			remQty, ok := remCap[rName]
			if !ok {
				continue // Node allocatable does not specify or limit this resource
			}
			if remQty.Cmp(minQty) < 0 {
				fits = false
				break
			}
		}
		if fits {
			return true
		}
	}
	return false
}

// normalizeNode ensures required node metadata and allocatable capacities are populated.
//
// WHY Allocatable Fallback:
// In live Kubernetes clusters, Kubelet always populates both Status.Capacity and Status.Allocatable,
// where Allocatable = Capacity - KubeReserved - SystemReserved - EvictionThreshold.
// In minimal unit test fixtures and synthetic mock nodes, test authors frequently specify only Capacity
// and leave Allocatable as nil. To prevent Karpenter's scheduler from panicking or treating mock nodes
// as having 0 schedulable resources in test suites, missing Allocatable is defensively initialized from Capacity
// (assuming 0 reserved overhead for test fixtures).
func normalizeNode(node *apiv1.Node, prunedReqs apiv1.ResourceList) *apiv1.Node {
	needsClamping := len(prunedReqs) > 0
	needsDefaults := node.Labels == nil ||
		node.Labels[apiv1.LabelHostname] == "" ||
		node.Labels[apiv1.LabelHostname] != node.Name ||
		node.Labels[apiv1.LabelArchStable] == "" ||
		node.Labels[apiv1.LabelOSStable] == "" ||
		node.Spec.ProviderID == "" ||
		annotations.IsUpcomingNode(node) ||
		annotations.IsSalvoNode(node) ||
		node.Status.Allocatable == nil ||
		len(node.Status.Capacity) == 0

	if !needsClamping && !needsDefaults {
		return node
	}

	cloned := node.DeepCopy()

	if cloned.Labels == nil {
		cloned.Labels = make(map[string]string)
	}
	if cloned.Labels[apiv1.LabelHostname] == "" || annotations.IsUpcomingNode(node) || annotations.IsSalvoNode(node) || cloned.Labels[apiv1.LabelHostname] != cloned.Name {
		cloned.Labels[apiv1.LabelHostname] = cloned.Name
	}
	if cloned.Labels[apiv1.LabelArchStable] == "" {
		if val := cloned.Labels["beta.kubernetes.io/arch"]; val != "" {
			cloned.Labels[apiv1.LabelArchStable] = val
		} else {
			cloned.Labels[apiv1.LabelArchStable] = "amd64"
		}
	}
	if cloned.Labels[apiv1.LabelOSStable] == "" {
		if val := cloned.Labels["beta.kubernetes.io/os"]; val != "" {
			cloned.Labels[apiv1.LabelOSStable] = val
		} else {
			cloned.Labels[apiv1.LabelOSStable] = "linux"
		}
	}
	if cloned.Spec.ProviderID == "" || annotations.IsUpcomingNode(node) || annotations.IsSalvoNode(node) {
		cloned.Spec.ProviderID = cloned.Name
	}

	// In production clusters, Status.Allocatable is always populated by Kubelet.
	// In minimal unit test fixtures, Allocatable is occasionally left uninitialized (nil).
	// As a defensive fallback for unit tests, default missing Allocatable from Capacity.
	if cloned.Status.Allocatable == nil {
		if len(cloned.Status.Capacity) > 0 {
			cloned.Status.Allocatable = cloned.Status.Capacity.DeepCopy()
		} else {
			cloned.Status.Allocatable = make(apiv1.ResourceList)
		}
	}
	if len(cloned.Status.Capacity) == 0 {
		cloned.Status.Capacity = cloned.Status.Allocatable.DeepCopy()
	}

	if needsClamping {
		clampAllocatable(cloned.Status.Allocatable, prunedReqs)
	}

	return cloned
}



