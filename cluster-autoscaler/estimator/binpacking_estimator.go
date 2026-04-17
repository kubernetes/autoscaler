/*
Copyright 2016 The Kubernetes Authors.

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

package estimator

import (
	"fmt"
	"math"
	"strconv"

	"slices"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/metrics"
	core_utils "k8s.io/autoscaler/cluster-autoscaler/simulator"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/scheduler/framework/plugins/podtopologyspread"
)

// BinpackingNodeEstimator estimates the number of needed nodes to handle the given amount of pods.
type BinpackingNodeEstimator struct {
	clusterSnapshot           clustersnapshot.ClusterSnapshot
	limiter                   EstimationLimiter
	podOrderer                EstimationPodOrderer
	context                   EstimationContext
	estimationAnalyserFunc    EstimationAnalyserFunc // optional
	fastpathBinpackingEnabled bool
}

// estimationState contains helper variables to avoid coping them independently in each function.
type estimationState struct {
	scheduledPods    []*apiv1.Pod
	newNodeNameIndex int
	lastNodeName     string
	newNodeNames     map[string]bool
	// map of node name that has at least one pod scheduled on it
	newNodesWithPods map[string]bool
}

func (s *estimationState) trackScheduledPod(pod *apiv1.Pod, nodeName string) {
	s.newNodesWithPods[nodeName] = true
	s.scheduledPods = append(s.scheduledPods, pod)
}

// NewBinpackingNodeEstimator builds a new BinpackingNodeEstimator.
func NewBinpackingNodeEstimator(
	clusterSnapshot clustersnapshot.ClusterSnapshot,
	limiter EstimationLimiter,
	podOrderer EstimationPodOrderer,
	context EstimationContext,
	estimationAnalyserFunc EstimationAnalyserFunc,
	fastpathBinpackingEnabled bool,
) *BinpackingNodeEstimator {
	return &BinpackingNodeEstimator{
		clusterSnapshot:           clusterSnapshot,
		limiter:                   limiter,
		podOrderer:                podOrderer,
		context:                   context,
		estimationAnalyserFunc:    estimationAnalyserFunc,
		fastpathBinpackingEnabled: fastpathBinpackingEnabled,
	}
}

func newEstimationState() *estimationState {
	return &estimationState{
		scheduledPods:    []*apiv1.Pod{},
		newNodeNameIndex: 0,
		lastNodeName:     "",
		newNodeNames:     map[string]bool{},
		newNodesWithPods: map[string]bool{},
	}
}

// Estimate implements First-Fit bin-packing approximation algorithm
// The ordering of the pods depend on the EstimatePodOrderer, the default
// order is DecreasingPodOrderer
// First-Fit Decreasing bin-packing approximation algorithm.
// See https://en.wikipedia.org/wiki/Bin_packing_problem for more details.
// While it is a multi-dimensional bin packing (cpu, mem, ports) in most cases the main dimension
// will be cpu thus the estimated overprovisioning of 11/9 * optimal + 6/9 should be
// still be maintained.
// It is assumed that all pods from the given list can fit to nodeTemplate.
// Returns the number of nodes needed to accommodate all pods from the list.
func (e *BinpackingNodeEstimator) Estimate(
	podsEquivalenceGroups []PodEquivalenceGroup,
	nodeTemplate *framework.NodeInfo,
	nodeGroup cloudprovider.NodeGroup,
) (int, []*apiv1.Pod) {
	observeBinpackingHeterogeneity(podsEquivalenceGroups, nodeTemplate)

	e.limiter.StartEstimation(podsEquivalenceGroups, nodeGroup, e.context)
	defer e.limiter.EndEstimation()

	podsEquivalenceGroups = e.podOrderer.Order(podsEquivalenceGroups, nodeTemplate, nodeGroup)

	useFastpathOnLastPEG := false
	if e.fastpathBinpackingEnabled {
		bestFastpathPEGindex := determineBestPEGToFastpath(podsEquivalenceGroups, nodeTemplate)
		if bestFastpathPEGindex != -1 {
			bestFastpathPEG := podsEquivalenceGroups[bestFastpathPEGindex]
			pegsWithoutBestForFastpath := append(podsEquivalenceGroups[:bestFastpathPEGindex], podsEquivalenceGroups[bestFastpathPEGindex+1:]...)
			// We'll put at the end the PEG which will benefit the most from fastpath binpacking, since it runs only on the last PEG
			podsEquivalenceGroups = append(pegsWithoutBestForFastpath, bestFastpathPEG)
			useFastpathOnLastPEG = true
		}
	}

	e.clusterSnapshot.Fork()
	defer func() {
		e.clusterSnapshot.Revert()
	}()

	estimationState := newEstimationState()
	newNodesAvailable := true
	for i, podsEquivalenceGroup := range podsEquivalenceGroups {
		var err error
		var remainingPods []*apiv1.Pod

		remainingPods, err = e.tryToScheduleOnExistingNodes(estimationState, podsEquivalenceGroup.Pods)
		if err != nil {
			klog.Error(err.Error())
			return 0, nil
		}

		if newNodesAvailable {
			// Since fastpath binpacking adds just one node to the snapshot, it will cause inaccurate simulations on subsequent loops, therefore we only use it on the last group
			if i == len(podsEquivalenceGroups)-1 && useFastpathOnLastPEG {
				newNodesAvailable, err = e.tryFastPath(estimationState, nodeTemplate, remainingPods)
			} else {
				newNodesAvailable, err = e.tryToScheduleOnNewNodes(estimationState, nodeTemplate, remainingPods)
			}
			if err != nil {
				klog.Error(err.Error())
				return 0, nil
			}
		}
	}

	if e.estimationAnalyserFunc != nil {
		e.estimationAnalyserFunc(e.clusterSnapshot, nodeGroup, estimationState.newNodesWithPods)
	}
	return len(estimationState.newNodesWithPods), estimationState.scheduledPods
}

func (e *BinpackingNodeEstimator) tryToScheduleOnExistingNodes(
	estimationState *estimationState,
	pods []*apiv1.Pod,
) ([]*apiv1.Pod, error) {
	var index int
	for index = 0; index < len(pods); index++ {
		pod := pods[index]

		// Try to schedule the pod on all nodes created during simulation
		nodeName, err := e.clusterSnapshot.SchedulePodOnAnyNodeMatching(pod, clustersnapshot.SchedulingOptions{IsNodeAcceptable: func(nodeInfo *framework.NodeInfo) bool {
			return estimationState.newNodeNames[nodeInfo.Node().Name]
		}})
		if err != nil && err.Type() == clustersnapshot.SchedulingInternalError {
			// Unexpected error.
			return nil, err
		} else if err != nil {
			// The pod couldn't be scheduled on any Node because of scheduling predicates.
			break
		}
		// The pod was scheduled on nodeName.
		estimationState.trackScheduledPod(pod, nodeName)
	}
	return pods[index:], nil
}

// Returns whether it is worth retrying adding new nodes and error in unexpected
// situations where whole estimation should be stopped.
func (e *BinpackingNodeEstimator) tryToScheduleOnNewNodes(
	estimationState *estimationState,
	nodeTemplate *framework.NodeInfo,
	pods []*apiv1.Pod,
) (bool, error) {
	for _, pod := range pods {
		found := false

		if estimationState.lastNodeName != "" {
			// Try to schedule the pod on only newly created node.
			err := e.clusterSnapshot.SchedulePod(pod, estimationState.lastNodeName)
			if err == nil {
				// The pod was scheduled on the newly created node.
				found = true
				estimationState.trackScheduledPod(pod, estimationState.lastNodeName)
			} else if err.Type() == clustersnapshot.SchedulingInternalError {
				// Unexpected error.
				return false, err
			}
			// The pod can't be scheduled on the newly created node because of scheduling predicates.

			// Check if node failed because of topology constraints.
			if isPodUsingHostNameTopologyKey(pod) && hasTopologyConstraintError(err) {
				// If the pod can't be scheduled on the last node because of topology constraints, we can stop binpacking.
				// The pod can't be scheduled on any new node either, because it has the same topology constraints.
				nodeName, err := e.clusterSnapshot.SchedulePodOnAnyNodeMatching(pod, clustersnapshot.SchedulingOptions{IsNodeAcceptable: func(nodeInfo *framework.NodeInfo) bool {
					return nodeInfo.Node().Name != estimationState.lastNodeName // only skip the last node that failed scheduling
				}})
				if err != nil && err.Type() == clustersnapshot.SchedulingInternalError {
					// Unexpected error.
					return false, err
				}
				if nodeName != "" {
					// The pod was scheduled on a different node, so we can continue binpacking.
					found = true
					estimationState.trackScheduledPod(pod, nodeName)
				}
			}
		}

		if !found {
			// If the last node we've added is empty and the pod couldn't schedule on it, it wouldn't be able to schedule
			// on a new node either. There is no point adding more nodes to snapshot in such case, especially because of
			// performance cost each extra node adds to future FitsAnyNodeMatching calls.
			if estimationState.lastNodeName != "" && !estimationState.newNodesWithPods[estimationState.lastNodeName] {
				return true, nil
			}

			// Stop binpacking if we reach the limit of nodes we can add.
			// We return the result of the binpacking that we already performed.
			//
			// The thresholdBasedEstimationLimiter implementation assumes that for
			// each call that returns true, one node gets added. Therefore this
			// must be the last check right before really adding a node.
			if !e.limiter.PermissionToAddNode() {
				return false, nil
			}

			// Add new node
			if err := e.addNewNodeToSnapshot(estimationState, nodeTemplate); err != nil {
				return false, fmt.Errorf("Error while adding new node for template to ClusterSnapshot; %w", err)
			}

			// And try to schedule pod to it.
			// Note that this may still fail (ex. if topology spreading with zonal topologyKey is used);
			// in this case we can't help the pending pod. We keep the node in clusterSnapshot to avoid
			// adding and removing node to snapshot for each such pod.
			if err := e.clusterSnapshot.SchedulePod(pod, estimationState.lastNodeName); err != nil && err.Type() == clustersnapshot.SchedulingInternalError {
				// Unexpected error.
				return false, err
			} else if err != nil {
				// The pod can't be scheduled on the new node because of scheduling predicates.
				break
			}
			// The pod got scheduled on the new node.
			estimationState.trackScheduledPod(pod, estimationState.lastNodeName)
		}
	}
	return true, nil
}

// An optimized version of tryToScheduleOnNewNodes
// It attempts to pack as much pods as possible on a single node, and then estimates the total
// amount of nodes by simple arithmetic: EstimatedNodes = ceil(TotalPods / PodsFittingOnASingleNode)
func (e *BinpackingNodeEstimator) tryFastPath(
	estimationState *estimationState,
	nodeTemplate *framework.NodeInfo,
	pods []*apiv1.Pod,
) (bool, error) {
	if len(pods) == 0 {
		return true, nil
	}
	if !e.limiter.PermissionToAddNode() {
		return false, nil
	}
	// Add test node to snapshot.
	if err := e.addNewNodeToSnapshot(estimationState, nodeTemplate); err != nil {
		return false, fmt.Errorf("Error while adding new node for template to ClusterSnapshot; %w", err)
	}

	i := 0
	for ; i < len(pods); i++ {
		if err := e.clusterSnapshot.SchedulePod(pods[i], estimationState.lastNodeName); err != nil && err.Type() == clustersnapshot.SchedulingInternalError {
			// Unexpected error.
			return false, err
		} else if err != nil {
			// The pod can't be scheduled on the new node because of scheduling predicates.
			break
		}
		estimationState.trackScheduledPod(pods[i], estimationState.lastNodeName)
	}
	podsPerNode := i
	if podsPerNode == 0 {
		return true, nil
	}
	scaleUpSize := len(pods) / podsPerNode
	if podsPerNode*scaleUpSize < len(pods) {
		scaleUpSize++
	}

	// We already added 1 node and scheduled on it, now the rest we can only mark without the simulation
	for j := 1; j < scaleUpSize; j++ {
		if !e.limiter.PermissionToAddNode() {
			return false, nil
		}
		fakeNodeName := fmt.Sprintf("%s-fake-%d", estimationState.lastNodeName, j)
		podsToSchedule := min(podsPerNode, len(pods)-i)
		for k := range podsToSchedule {
			estimationState.trackScheduledPod(pods[i+k], fakeNodeName)
		}
		i += podsToSchedule
	}

	return true, nil
}

func (e *BinpackingNodeEstimator) addNewNodeToSnapshot(
	estimationState *estimationState,
	template *framework.NodeInfo,
) error {
	newNodeInfo, err := core_utils.SanitizedNodeInfo(template, fmt.Sprintf("e-%d", estimationState.newNodeNameIndex))
	if err != nil {
		return err
	}

	if err := e.clusterSnapshot.AddNodeInfo(newNodeInfo); err != nil {
		return err
	}
	estimationState.newNodeNameIndex++
	estimationState.lastNodeName = newNodeInfo.Node().Name
	estimationState.newNodeNames[estimationState.lastNodeName] = true
	return nil
}

// isTopologyConstraintError determines if an error is related to pod topology spread constraints
// by checking the predicate name and reasons
func hasTopologyConstraintError(err clustersnapshot.SchedulingError) bool {
	if err == nil {
		return false
	}

	// Check reasons for mentions of topology or constraints
	return slices.Contains(err.FailingPredicateReasons(), podtopologyspread.ErrReasonConstraintsNotMatch)
}

// isPodUsingHostNameTopoKey returns true if the pod has any topology spread
// constraint that uses the kubernetes.io/hostname topology key
func isPodUsingHostNameTopologyKey(pod *apiv1.Pod) bool {
	if pod == nil || pod.Spec.TopologySpreadConstraints == nil {
		return false
	}

	for _, constraint := range pod.Spec.TopologySpreadConstraints {
		if constraint.TopologyKey == apiv1.LabelHostname {
			return true
		}
	}

	return false
}

func observeBinpackingHeterogeneity(podsEquivalenceGroups []PodEquivalenceGroup, nodeTemplate *framework.NodeInfo) {
	node := nodeTemplate.Node()
	var instanceType, cpuCount string
	if node != nil {
		if node.Labels != nil {
			instanceType = node.Labels[apiv1.LabelInstanceTypeStable]
		}
		cpuCount = node.Status.Capacity.Cpu().String()
	}
	namespaces := make(map[string]bool)
	for _, peg := range podsEquivalenceGroups {
		e := peg.Exemplar()
		if e != nil {
			namespaces[e.Namespace] = true
		}
	}
	// Quantize # of namespaces to limit metric cardinality.
	nsCountBucket := ""
	if len(namespaces) <= 5 {
		nsCountBucket = strconv.Itoa(len(namespaces))
	} else if len(namespaces) <= 10 {
		nsCountBucket = "6-10"
	} else {
		nsCountBucket = "11+"
	}
	metrics.ObserveBinpackingHeterogeneity(instanceType, cpuCount, nsCountBucket, len(podsEquivalenceGroups))
}

func hasNonHostnamePodAntiAffinity(pod *apiv1.Pod) bool {
	if pod.Spec.Affinity == nil || pod.Spec.Affinity.PodAntiAffinity == nil || pod.Spec.Affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution == nil {
		return false
	}
	for _, affinityTerm := range pod.Spec.Affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution {
		if affinityTerm.TopologyKey != apiv1.LabelHostname {
			return true
		}
	}
	return false
}

func shouldUseFastPath(podEquivalenceGroup PodEquivalenceGroup) bool {
	pod := podEquivalenceGroup.Exemplar()
	if pod == nil {
		return false
	}

	// Fastpath assumes that once a pod failed to schedule on a node, it will never be possible to schedule any more pods on this node.
	// This assumption is not correct in cases of topology spread constraints.
	// Fastpath also assumes that if it was able to schedule pods on a new node, it will be able to then schedule an equal amount of pods on an additional identical new node,
	// which isn't correct in cases of zonal/regional pod anti affinity.
	if pod.Spec.TopologySpreadConstraints != nil || hasNonHostnamePodAntiAffinity(pod) {
		return false
	}

	return true
}

// determineBestPEGToFastpath returns the index of the pod equivalence group that would benefit the most from fastpath binpacking (= largest expected number of (pods - (pods/nodes))
// Simulation complexity is generally O(nodes * pods), and fastpath fully simulates the
// first node, therefore the number of simulations saved would be approximately:
// PodsPerNode * (nodes - 1) = (pods/nodes) * (nodes - 1) = pods - (pods/nodes)
// returns -1 if none of the groups support fastpath binpacking
func determineBestPEGToFastpath(podsEquivalenceGroups []PodEquivalenceGroup, nodeTemplate *framework.NodeInfo) int {
	maxSimulationsSaved := 0
	bestPEGIndex := -1
	for i, peg := range podsEquivalenceGroups {
		if peg.Exemplar() == nil {
			continue
		}
		numNodesByAntiAffinity := 0
		numNodesByCpu := 0
		numNodesByMemory := 0

		if peg.Exemplar().Spec.Affinity != nil && peg.Exemplar().Spec.Affinity.PodAntiAffinity != nil {
			for _, affinityTerm := range peg.Exemplar().Spec.Affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution {
				if affinityTerm.TopologyKey == apiv1.LabelHostname && labelSelectorMatches(affinityTerm.LabelSelector, peg.Exemplar().Labels) {
					numNodesByAntiAffinity = len(peg.Pods)
				}
			}
		}
		if len(peg.Exemplar().Spec.Containers) > 0 && peg.Exemplar().Spec.Containers[0].Resources.Requests != nil {
			resourcesRequests := peg.Exemplar().Spec.Containers[0].Resources.Requests
			if resourcesRequests.Cpu() != nil {
				numNodesByCpu = int(math.Ceil(float64(len(peg.Pods)) * resourcesRequests.Cpu().AsApproximateFloat64() / nodeTemplate.Node().Status.Capacity.Cpu().AsApproximateFloat64()))
			}
			if resourcesRequests.Memory() != nil {
				numNodesByMemory = int(math.Ceil(float64(len(peg.Pods)) * resourcesRequests.Memory().AsApproximateFloat64() / nodeTemplate.Node().Status.Capacity.Memory().AsApproximateFloat64()))
			}
		}
		numNodes := max(numNodesByAntiAffinity, numNodesByCpu, numNodesByMemory)

		simulationsSaved := 0
		if numNodes > 0 {
			simulationsSaved = len(peg.Pods) - (len(peg.Pods) / numNodes)
		}
		// In case the score is equal, we still want to take the latest pod group to remain as close to the original pod ordering as possible
		if simulationsSaved >= maxSimulationsSaved && shouldUseFastPath(peg) {
			bestPEGIndex = i
			maxSimulationsSaved = simulationsSaved
		}
	}
	return bestPEGIndex
}

// labelSelectorMatches checks if the given LabelSelector matches the provided podLabels.
func labelSelectorMatches(selector *metav1.LabelSelector, podLabels map[string]string) bool {
	if selector == nil {
		return false
	}
	ls, err := metav1.LabelSelectorAsSelector(selector)
	if err != nil {
		return false
	}
	return ls.Matches(labels.Set(podLabels))
}
