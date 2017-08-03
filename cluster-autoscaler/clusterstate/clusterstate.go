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

package clusterstate

import (
	"fmt"
	"reflect"
	"sync"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate/api"
	"k8s.io/autoscaler/cluster-autoscaler/utils/deletetaint"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/golang/glog"
)

const (
	// MaxNodeStartupTime is the maximum time from the moment the node is registered to the time the node is ready.
	MaxNodeStartupTime = 5 * time.Minute

	// MaxStatusSettingDelayAfterCreation is the maximum time for node to set its initial status after the
	// node is registered.
	MaxStatusSettingDelayAfterCreation = 2 * time.Minute
)

// ScaleUpRequest contains information about the requested node group scale up.
type ScaleUpRequest struct {
	// NodeGroupName is the node group to be scaled up.
	NodeGroupName string
	// Time is the time when the request was submitted.
	Time time.Time
	// ExpectedAddTime is the time at which the request should be fulfilled.
	ExpectedAddTime time.Time
	// How much the node group is increased.
	Increase int
}

// ScaleDownRequest contains information about the requested node deletion.
type ScaleDownRequest struct {
	// NodeName is the name of the node to be deleted.
	NodeName string
	// NodeGroupName is the node group of the deleted node.
	NodeGroupName string
	// Time is the time when the node deletion was requested.
	Time time.Time
	// ExpectedDeleteTime is the time when the node is excpected to be deleted.
	ExpectedDeleteTime time.Time
}

// ClusterStateRegistryConfig contains configuration information for ClusterStateRegistry.
type ClusterStateRegistryConfig struct {
	// Maximum percentage of unready nodes in total in, if the number is higher than OkTotalUnreadyCount
	MaxTotalUnreadyPercentage float64
	// Number of nodes that can be unready in total. If the number is higher than that then MaxTotalUnreadyPercentage applies.
	OkTotalUnreadyCount int
}

// IncorrectNodeGroupSize contains information about how much the current size of the node group
// differs from the expected size. Prolonged, stable mismatch is an indication of quota
// or startup issues.
type IncorrectNodeGroupSize struct {
	// ExpectedSize is the size of the node group measured on the cloud provider side.
	ExpectedSize int
	// CurrentSize is the size of the node group measured on the kubernetes side.
	CurrentSize int
	// FirstObserved is the time when the given difference occurred.
	FirstObserved time.Time
}

// UnregisteredNode contains information about nodes that are present on the cluster provider side
// but failed to register in Kubernetes.
type UnregisteredNode struct {
	// Node is a dummy node that contains only the name of the node.
	Node *apiv1.Node
	// UnregisteredSince is the time when the node was first spotted.
	UnregisteredSince time.Time
}

// ClusterStateRegistry is a structure to keep track the current state of the cluster.
type ClusterStateRegistry struct {
	sync.Mutex
	config                  ClusterStateRegistryConfig
	scaleUpRequests         []*ScaleUpRequest
	scaleDownRequests       []*ScaleDownRequest
	nodes                   []*apiv1.Node
	cloudProvider           cloudprovider.CloudProvider
	perNodeGroupReadiness   map[string]Readiness
	totalReadiness          Readiness
	acceptableRanges        map[string]AcceptableRange
	incorrectNodeGroupSizes map[string]IncorrectNodeGroupSize
	unregisteredNodes       map[string]UnregisteredNode
	candidatesForScaleDown  map[string][]string
	lastStatus              *api.ClusterAutoscalerStatus
	lastScaleDownUpdateTime time.Time
}

// NewClusterStateRegistry creates new ClusterStateRegistry.
func NewClusterStateRegistry(cloudProvider cloudprovider.CloudProvider, config ClusterStateRegistryConfig) *ClusterStateRegistry {
	emptyStatus := &api.ClusterAutoscalerStatus{
		ClusterwideConditions: make([]api.ClusterAutoscalerCondition, 0),
		NodeGroupStatuses:     make([]api.NodeGroupStatus, 0),
	}
	return &ClusterStateRegistry{
		scaleUpRequests:         make([]*ScaleUpRequest, 0),
		scaleDownRequests:       make([]*ScaleDownRequest, 0),
		nodes:                   make([]*apiv1.Node, 0),
		cloudProvider:           cloudProvider,
		config:                  config,
		perNodeGroupReadiness:   make(map[string]Readiness),
		acceptableRanges:        make(map[string]AcceptableRange),
		incorrectNodeGroupSizes: make(map[string]IncorrectNodeGroupSize),
		unregisteredNodes:       make(map[string]UnregisteredNode),
		candidatesForScaleDown:  make(map[string][]string),
		lastStatus:              emptyStatus,
	}
}

// RegisterScaleUp registers scale up.
func (csr *ClusterStateRegistry) RegisterScaleUp(request *ScaleUpRequest) {
	csr.Lock()
	defer csr.Unlock()
	csr.scaleUpRequests = append(csr.scaleUpRequests, request)
}

// RegisterScaleDown registers node scale down.
func (csr *ClusterStateRegistry) RegisterScaleDown(request *ScaleDownRequest) {
	csr.Lock()
	defer csr.Unlock()
	csr.scaleDownRequests = append(csr.scaleDownRequests, request)
}

// To be executed under a lock.
func (csr *ClusterStateRegistry) cleanUp(currentTime time.Time) {
	newSur := make([]*ScaleUpRequest, 0)
	for _, sur := range csr.scaleUpRequests {
		if sur.ExpectedAddTime.After(currentTime) {
			newSur = append(newSur, sur)
		}
	}
	csr.scaleUpRequests = newSur

	newSdr := make([]*ScaleDownRequest, 0)
	for _, sdr := range csr.scaleDownRequests {
		if sdr.ExpectedDeleteTime.After(currentTime) {
			newSdr = append(newSdr, sdr)
		}
	}
	csr.scaleDownRequests = newSdr
}

// UpdateNodes updates the state of the nodes in the ClusterStateRegistry and recalculates the statss
func (csr *ClusterStateRegistry) UpdateNodes(nodes []*apiv1.Node, currentTime time.Time) error {
	targetSizes, err := getTargetSizes(csr.cloudProvider)
	if err != nil {
		return err
	}
	notRegistered, err := getNotRegisteredNodes(nodes, csr.cloudProvider, currentTime)
	if err != nil {
		return err
	}

	csr.Lock()
	defer csr.Unlock()

	csr.cleanUp(currentTime)
	csr.nodes = nodes

	csr.updateReadinessStats(currentTime)
	csr.updateAcceptableRanges(targetSizes)
	csr.updateIncorrectNodeGroupSizes(currentTime)
	csr.updateUnregisteredNodes(notRegistered)
	return nil
}

// getTargetSizes gets target sizes of node groups.
func getTargetSizes(cp cloudprovider.CloudProvider) (map[string]int, error) {
	result := make(map[string]int)
	for _, ng := range cp.NodeGroups() {
		size, err := ng.TargetSize()
		if err != nil {
			return map[string]int{}, err
		}
		result[ng.Id()] = size
	}
	return result, nil
}

// IsClusterHealthy returns true if the cluster health is within the acceptable limits
func (csr *ClusterStateRegistry) IsClusterHealthy() bool {
	csr.Lock()
	defer csr.Unlock()

	totalUnready := csr.totalReadiness.Unready + csr.totalReadiness.LongNotStarted

	if totalUnready > csr.config.OkTotalUnreadyCount &&
		float64(totalUnready) > csr.config.MaxTotalUnreadyPercentage/100.0*float64(len(csr.nodes)) {
		return false
	}

	return true
}

// IsNodeGroupHealthy returns true if the node group health is within the acceptable limits
func (csr *ClusterStateRegistry) IsNodeGroupHealthy(nodeGroupName string) bool {
	acceptable, found := csr.acceptableRanges[nodeGroupName]
	if !found {
		glog.Warningf("Failed to find acceptable ranges for %v", nodeGroupName)
		return false
	}

	readiness, found := csr.perNodeGroupReadiness[nodeGroupName]
	if !found {
		// No nodes but target == 0 or just scaling up.
		if acceptable.CurrentTarget == 0 || (acceptable.MinNodes == 0 && acceptable.CurrentTarget > 0) {
			return true
		}
		glog.Warningf("Failed to find readiness information for %v", nodeGroupName)
		return false
	}

	unjustifiedUnready := 0
	// Too few nodes, something is missing. Below the expected node count.
	if readiness.Ready < acceptable.MinNodes {
		unjustifiedUnready += acceptable.MinNodes - readiness.Ready
	}
	// TODO: verify against maxnodes as well.

	if unjustifiedUnready > csr.config.OkTotalUnreadyCount &&
		float64(unjustifiedUnready) > csr.config.MaxTotalUnreadyPercentage/100.0*
			float64(readiness.Ready+readiness.Unready+readiness.NotStarted+readiness.LongNotStarted) {
		return false
	}

	return true
}

// IsNodeGroupScalingUp returns true if the node group is currently scaling up.
func (csr *ClusterStateRegistry) IsNodeGroupScalingUp(nodeGroupName string) bool {
	acceptable, found := csr.acceptableRanges[nodeGroupName]
	if !found {
		glog.Warningf("Failed to find acceptable ranges for %v", nodeGroupName)
		return false
	}

	readiness, found := csr.perNodeGroupReadiness[nodeGroupName]
	if !found {
		// Node group has size 0 (was scaled to 0 before).
		if acceptable.MinNodes == 0 {
			return acceptable.CurrentTarget > 0
		}
		glog.Warningf("Failed to find readiness information for %v", nodeGroupName)
		return false
	}

	provisioned := readiness.Registered - readiness.NotStarted - readiness.LongNotStarted
	if acceptable.CurrentTarget <= provisioned {
		return false
	}
	// Let's check if there is an active scale up request
	for _, request := range csr.scaleUpRequests {
		if request.NodeGroupName == nodeGroupName {
			return true
		}
	}
	return false
}

// AcceptableRange contains information about acceptable size of a node group.
type AcceptableRange struct {
	// MinNodes is the minimum number of nodes in the group.
	MinNodes int
	// MaxNodes is the maximum number of nodes in the group.
	MaxNodes int
	// CurrentTarget is the current target size of the group.
	CurrentTarget int
}

// updateAcceptableRanges updates cluster state registry with how many nodes can be in a cluster.
// The function assumes that the nodeGroup.TargetSize() is the desired number of nodes.
// So if there has been a recent scale up of size 5 then there should be between targetSize-5 and targetSize
// nodes in ready state. In the same way, if there have been 3 nodes removed recently then
// the expected number of ready nodes is between targetSize and targetSize + 3.
func (csr *ClusterStateRegistry) updateAcceptableRanges(targetSize map[string]int) {
	result := make(map[string]AcceptableRange)
	for _, nodeGroup := range csr.cloudProvider.NodeGroups() {
		size := targetSize[nodeGroup.Id()]
		result[nodeGroup.Id()] = AcceptableRange{
			MinNodes:      size,
			MaxNodes:      size,
			CurrentTarget: size,
		}
	}
	for _, sur := range csr.scaleUpRequests {
		val := result[sur.NodeGroupName]
		val.MinNodes -= sur.Increase
		result[sur.NodeGroupName] = val
	}
	for _, sdr := range csr.scaleDownRequests {
		val := result[sdr.NodeGroupName]
		val.MaxNodes += 1
		result[sdr.NodeGroupName] = val
	}
	csr.acceptableRanges = result
}

// Readiness contains readiness information about a group of nodes.
type Readiness struct {
	// Number of ready nodes.
	Ready int
	// Number of unready nodes that broke down after they started.
	Unready int
	// Number of nodes that are being currently deleted. They exist in K8S but
	// are not included in NodeGroup.TargetSize().
	Deleted int
	// Number of nodes that failed to start within a reasonable limit.
	LongNotStarted int
	// Number of nodes that are not yet fully started.
	NotStarted int
	// Number of all registered nodes in the group (ready/unready/deleted/etc).
	Registered int
	// Time when the readiness was measured.
	Time time.Time
}

func (csr *ClusterStateRegistry) updateReadinessStats(currentTime time.Time) {

	perNodeGroup := make(map[string]Readiness)
	total := Readiness{Time: currentTime}

	update := func(current Readiness, node *apiv1.Node, ready bool) Readiness {
		current.Registered++
		if deletetaint.HasToBeDeletedTaint(node) {
			current.Deleted++
		} else if isNodeNotStarted(node) && node.CreationTimestamp.Time.Add(MaxNodeStartupTime).Before(currentTime) {
			current.LongNotStarted++
		} else if isNodeNotStarted(node) {
			current.NotStarted++
		} else if ready {
			current.Ready++
		} else {
			current.Unready++
		}
		return current
	}

	for _, node := range csr.nodes {
		nodeGroup, errNg := csr.cloudProvider.NodeGroupForNode(node)
		ready, _, errReady := kube_util.GetReadinessState(node)

		// Node is most likely not autoscaled, however check the errors.
		if nodeGroup == nil || reflect.ValueOf(nodeGroup).IsNil() {
			if errNg != nil {
				glog.Warningf("Failed to get nodegroup for %s: %v", node.Name, errNg)
			}
			if errReady != nil {
				glog.Warningf("Failed to get readiness info for %s: %v", node.Name, errReady)
			}
		} else {
			perNodeGroup[nodeGroup.Id()] = update(perNodeGroup[nodeGroup.Id()], node, ready)
		}
		total = update(total, node, ready)
	}
	for ngId, ngReadiness := range perNodeGroup {
		ngReadiness.Time = currentTime
		perNodeGroup[ngId] = ngReadiness
	}
	csr.perNodeGroupReadiness = perNodeGroup
	csr.totalReadiness = total
}

// Calculates which node groups have incorrect size.
func (csr *ClusterStateRegistry) updateIncorrectNodeGroupSizes(currentTime time.Time) {
	result := make(map[string]IncorrectNodeGroupSize)
	for _, nodeGroup := range csr.cloudProvider.NodeGroups() {
		acceptableRange, found := csr.acceptableRanges[nodeGroup.Id()]
		if !found {
			glog.Warningf("Acceptable range for node group %s not found", nodeGroup.Id())
			continue
		}
		readiness, found := csr.perNodeGroupReadiness[nodeGroup.Id()]
		if !found {
			// if MinNodes == 0 node group has been scaled to 0 and everything's fine
			if acceptableRange.MinNodes != 0 {
				glog.Warningf("Readiness for node group %s not found", nodeGroup.Id())
			}
			continue
		}
		if readiness.Registered > acceptableRange.MaxNodes ||
			readiness.Registered < acceptableRange.MinNodes {
			incorrect := IncorrectNodeGroupSize{
				CurrentSize:   readiness.Registered,
				ExpectedSize:  acceptableRange.CurrentTarget,
				FirstObserved: currentTime,
			}
			existing, found := csr.incorrectNodeGroupSizes[nodeGroup.Id()]
			if found {
				if incorrect.CurrentSize == existing.CurrentSize &&
					incorrect.ExpectedSize == existing.ExpectedSize {
					incorrect = existing
				}
			}
			result[nodeGroup.Id()] = incorrect
		}
	}
	csr.incorrectNodeGroupSizes = result
}

func (csr *ClusterStateRegistry) updateUnregisteredNodes(unregisteredNodes []UnregisteredNode) {
	result := make(map[string]UnregisteredNode)
	for _, unregistered := range unregisteredNodes {
		if prev, found := csr.unregisteredNodes[unregistered.Node.Name]; found {
			result[unregistered.Node.Name] = prev
		} else {
			result[unregistered.Node.Name] = unregistered
		}
	}
	csr.unregisteredNodes = result
}

//GetUnregisteredNodes returns a list of all unregistered nodes.
func (csr *ClusterStateRegistry) GetUnregisteredNodes() []UnregisteredNode {
	csr.Lock()
	defer csr.Unlock()

	result := make([]UnregisteredNode, 0, len(csr.unregisteredNodes))
	for _, unregistered := range csr.unregisteredNodes {
		result = append(result, unregistered)
	}
	return result
}

// UpdateScaleDownCandidates updates scale down candidates
func (csr *ClusterStateRegistry) UpdateScaleDownCandidates(nodes []*apiv1.Node, now time.Time) {
	result := make(map[string][]string)
	for _, node := range nodes {
		group, err := csr.cloudProvider.NodeGroupForNode(node)
		if err != nil {
			glog.Warningf("Failed to get node group for %s: %v", node.Name, err)
			continue
		}
		if group == nil || reflect.ValueOf(group).IsNil() {
			continue
		}
		result[group.Id()] = append(result[group.Id()], node.Name)
	}
	csr.candidatesForScaleDown = result
	csr.lastScaleDownUpdateTime = now
}

// GetStatus returns ClusterAutoscalerStatus with the current cluster autoscaler status.
func (csr *ClusterStateRegistry) GetStatus(now time.Time) *api.ClusterAutoscalerStatus {
	result := &api.ClusterAutoscalerStatus{
		ClusterwideConditions: make([]api.ClusterAutoscalerCondition, 0),
		NodeGroupStatuses:     make([]api.NodeGroupStatus, 0),
	}
	for _, nodeGroup := range csr.cloudProvider.NodeGroups() {
		nodeGroupStatus := api.NodeGroupStatus{
			ProviderID: nodeGroup.Id(),
			Conditions: make([]api.ClusterAutoscalerCondition, 0),
		}
		readiness := csr.perNodeGroupReadiness[nodeGroup.Id()]
		acceptable := csr.acceptableRanges[nodeGroup.Id()]

		// Health.
		nodeGroupStatus.Conditions = append(nodeGroupStatus.Conditions, buildHealthStatusNodeGroup(
			csr.IsNodeGroupHealthy(nodeGroup.Id()), readiness, acceptable, nodeGroup.MinSize(), nodeGroup.MaxSize()))

		// Scale up.
		nodeGroupStatus.Conditions = append(nodeGroupStatus.Conditions, buildScaleUpStatusNodeGroup(
			csr.IsNodeGroupScalingUp(nodeGroup.Id()), readiness, acceptable))

		// Scale down.
		nodeGroupStatus.Conditions = append(nodeGroupStatus.Conditions, buildScaleDownStatusNodeGroup(
			csr.candidatesForScaleDown[nodeGroup.Id()], csr.lastScaleDownUpdateTime))

		result.NodeGroupStatuses = append(result.NodeGroupStatuses, nodeGroupStatus)
	}
	result.ClusterwideConditions = append(result.ClusterwideConditions,
		buildHealthStatusClusterwide(csr.IsClusterHealthy(), csr.totalReadiness))
	result.ClusterwideConditions = append(result.ClusterwideConditions,
		buildScaleUpStatusClusterwide(result.NodeGroupStatuses, csr.totalReadiness))
	result.ClusterwideConditions = append(result.ClusterwideConditions,
		buildScaleDownStatusClusterwide(csr.candidatesForScaleDown, csr.lastScaleDownUpdateTime))

	updateLastTransition(csr.lastStatus, result)
	csr.lastStatus = result
	return result
}

// GetClusterReadiness returns current readiness stats of cluster
func (csr *ClusterStateRegistry) GetClusterReadiness() Readiness {
	return csr.totalReadiness
}

func buildHealthStatusNodeGroup(isReady bool, readiness Readiness, acceptable AcceptableRange, minSize, maxSize int) api.ClusterAutoscalerCondition {
	condition := api.ClusterAutoscalerCondition{
		Type: api.ClusterAutoscalerHealth,
		Message: fmt.Sprintf("ready=%d unready=%d notStarted=%d longNotStarted=%d registered=%d cloudProviderTarget=%d (minSize=%d, maxSize=%d)",
			readiness.Ready,
			readiness.Unready,
			readiness.NotStarted,
			readiness.LongNotStarted,
			readiness.Registered,
			acceptable.CurrentTarget,
			minSize,
			maxSize),
		LastProbeTime: metav1.Time{Time: readiness.Time},
	}
	if isReady {
		condition.Status = api.ClusterAutoscalerHealthy
	} else {
		condition.Status = api.ClusterAutoscalerUnhealthy
	}
	return condition
}

func buildScaleUpStatusNodeGroup(isScaleUpInProgress bool, readiness Readiness, acceptable AcceptableRange) api.ClusterAutoscalerCondition {
	condition := api.ClusterAutoscalerCondition{
		Type: api.ClusterAutoscalerScaleUp,
		Message: fmt.Sprintf("ready=%d cloudProviderTarget=%d",
			readiness.Ready,
			acceptable.CurrentTarget),
		LastProbeTime: metav1.Time{Time: readiness.Time},
	}
	if isScaleUpInProgress {
		condition.Status = api.ClusterAutoscalerInProgress
	} else {
		condition.Status = api.ClusterAutoscalerNoActivity
	}
	return condition
}

func buildScaleDownStatusNodeGroup(candidates []string, lastProbed time.Time) api.ClusterAutoscalerCondition {
	condition := api.ClusterAutoscalerCondition{
		Type:          api.ClusterAutoscalerScaleDown,
		Message:       fmt.Sprintf("candidates=%d", len(candidates)),
		LastProbeTime: metav1.Time{Time: lastProbed},
	}
	if len(candidates) > 0 {
		condition.Status = api.ClusterAutoscalerCandidatesPresent
	} else {
		condition.Status = api.ClusterAutoscalerNoCandidates
	}
	return condition
}

func buildHealthStatusClusterwide(isReady bool, readiness Readiness) api.ClusterAutoscalerCondition {
	condition := api.ClusterAutoscalerCondition{
		Type: api.ClusterAutoscalerHealth,
		Message: fmt.Sprintf("ready=%d unready=%d notStarted=%d longNotStarted=%d registered=%d",
			readiness.Ready,
			readiness.Unready,
			readiness.NotStarted,
			readiness.LongNotStarted,
			readiness.Registered,
		),
		LastProbeTime: metav1.Time{Time: readiness.Time},
	}
	if isReady {
		condition.Status = api.ClusterAutoscalerHealthy
	} else {
		condition.Status = api.ClusterAutoscalerUnhealthy
	}
	return condition
}

func buildScaleUpStatusClusterwide(nodeGroupStatuses []api.NodeGroupStatus, readiness Readiness) api.ClusterAutoscalerCondition {
	isScaleUpInProgress := false
	for _, nodeGroupStatuses := range nodeGroupStatuses {
		for _, condition := range nodeGroupStatuses.Conditions {
			if condition.Type == api.ClusterAutoscalerScaleUp &&
				condition.Status == api.ClusterAutoscalerInProgress {
				isScaleUpInProgress = true
			}
		}
	}

	condition := api.ClusterAutoscalerCondition{
		Type: api.ClusterAutoscalerScaleUp,
		Message: fmt.Sprintf("ready=%d registered=%d",
			readiness.Ready,
			readiness.Registered),
		LastProbeTime: metav1.Time{Time: readiness.Time},
	}
	if isScaleUpInProgress {
		condition.Status = api.ClusterAutoscalerInProgress
	} else {
		condition.Status = api.ClusterAutoscalerNoActivity
	}
	return condition
}

func buildScaleDownStatusClusterwide(candidates map[string][]string, lastProbed time.Time) api.ClusterAutoscalerCondition {
	totalCandidates := 0
	for _, val := range candidates {
		totalCandidates += len(val)
	}
	condition := api.ClusterAutoscalerCondition{
		Type:          api.ClusterAutoscalerScaleDown,
		Message:       fmt.Sprintf("candidates=%d", totalCandidates),
		LastProbeTime: metav1.Time{Time: lastProbed},
	}
	if totalCandidates > 0 {
		condition.Status = api.ClusterAutoscalerCandidatesPresent
	} else {
		condition.Status = api.ClusterAutoscalerNoCandidates
	}
	return condition
}

func isNodeNotStarted(node *apiv1.Node) bool {
	for _, condition := range node.Status.Conditions {
		if condition.Type == apiv1.NodeReady &&
			condition.Status == apiv1.ConditionFalse &&
			condition.LastTransitionTime.Time.Sub(node.CreationTimestamp.Time) < MaxStatusSettingDelayAfterCreation {
			return true
		}
		if condition.Type == apiv1.NodeOutOfDisk &&
			condition.Status == apiv1.ConditionTrue &&
			condition.LastTransitionTime.Time.Sub(node.CreationTimestamp.Time) < MaxStatusSettingDelayAfterCreation {
			return true
		}
		if condition.Type == apiv1.NodeNetworkUnavailable &&
			condition.Status == apiv1.ConditionTrue &&
			condition.LastTransitionTime.Time.Sub(node.CreationTimestamp.Time) < MaxStatusSettingDelayAfterCreation {
			return true
		}
	}
	return false
}

func updateLastTransition(oldStatus, newStatus *api.ClusterAutoscalerStatus) {
	newStatus.ClusterwideConditions = updateLastTransitionSingleList(
		oldStatus.ClusterwideConditions, newStatus.ClusterwideConditions)
	updatedNgStatuses := make([]api.NodeGroupStatus, 0)
	for _, ngStatus := range newStatus.NodeGroupStatuses {
		oldConds := make([]api.ClusterAutoscalerCondition, 0)
		for _, oldNgStatus := range oldStatus.NodeGroupStatuses {
			if ngStatus.ProviderID == oldNgStatus.ProviderID {
				oldConds = oldNgStatus.Conditions
				break
			}
		}
		newConds := updateLastTransitionSingleList(oldConds, ngStatus.Conditions)
		updatedNgStatuses = append(
			updatedNgStatuses,
			api.NodeGroupStatus{
				ProviderID: ngStatus.ProviderID,
				Conditions: newConds,
			})
	}
	newStatus.NodeGroupStatuses = updatedNgStatuses
}

func updateLastTransitionSingleList(oldConds, newConds []api.ClusterAutoscalerCondition) []api.ClusterAutoscalerCondition {
	result := make([]api.ClusterAutoscalerCondition, 0)
	// We have ~3 conditions, so O(n**2) is good enough
	for _, condition := range newConds {
		condition.LastTransitionTime = condition.LastProbeTime
		for _, oldCondition := range oldConds {
			if condition.Type == oldCondition.Type {
				if condition.Status == oldCondition.Status {
					condition.LastTransitionTime = oldCondition.LastTransitionTime
				}
				break
			}
		}
		result = append(result, condition)
	}
	return result
}

// GetIncorrectNodeGroupSize gets IncorrectNodeGroupSizeInformation for the given node group.
func (csr *ClusterStateRegistry) GetIncorrectNodeGroupSize(nodeGroupName string) *IncorrectNodeGroupSize {
	result, found := csr.incorrectNodeGroupSizes[nodeGroupName]
	if !found {
		return nil
	}
	return &result
}

// GetUpcomingNodes returns how many new nodes will be added shortly to the node groups or should become ready soon.
// The functiom may overestimate the number of nodes.
func (csr *ClusterStateRegistry) GetUpcomingNodes() map[string]int {
	csr.Lock()
	defer csr.Unlock()

	result := make(map[string]int)
	for _, nodeGroup := range csr.cloudProvider.NodeGroups() {
		id := nodeGroup.Id()
		readiness := csr.perNodeGroupReadiness[id]
		ar := csr.acceptableRanges[id]
		// newNodes is the number of nodes that
		newNodes := ar.CurrentTarget - (readiness.Ready + readiness.Unready + readiness.LongNotStarted)
		if newNodes <= 0 {
			// Negative value is unlikely but theoretically possible.
			continue
		}
		result[id] = newNodes
	}
	return result
}

// Calculates which of the existing cloud provider nodes are not registered in Kubernetes.
func getNotRegisteredNodes(allNodes []*apiv1.Node, cloudProvider cloudprovider.CloudProvider, time time.Time) ([]UnregisteredNode, error) {
	registered := sets.NewString()
	for _, node := range allNodes {
		registered.Insert(node.Spec.ProviderID)
	}
	notRegistered := make([]UnregisteredNode, 0)
	for _, nodeGroup := range cloudProvider.NodeGroups() {
		nodes, err := nodeGroup.Nodes()
		if err != nil {
			return []UnregisteredNode{}, err
		}
		for _, node := range nodes {
			if !registered.Has(node) {
				notRegistered = append(notRegistered, UnregisteredNode{
					Node: &apiv1.Node{
						ObjectMeta: metav1.ObjectMeta{
							Name: node,
						},
						Spec: apiv1.NodeSpec{
							ProviderID: node,
						},
					},
					UnregisteredSince: time,
				})
			}
		}
	}
	return notRegistered, nil
}
