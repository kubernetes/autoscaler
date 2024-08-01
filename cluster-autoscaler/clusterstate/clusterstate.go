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
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate/api"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate/utils"
	"k8s.io/autoscaler/cluster-autoscaler/metrics"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodegroupconfig"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodegroups/asyncnodegroups"
	"k8s.io/autoscaler/cluster-autoscaler/utils/backoff"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"

	klog "k8s.io/klog/v2"
)

const (
	// MaxNodeStartupTime is the maximum time from the moment the node is registered to the time the node is ready.
	MaxNodeStartupTime = 15 * time.Minute
	// maxErrorMessageSize is the maximum size of error messages displayed in config map as the max size of configmap is 1MB.
	maxErrorMessageSize = 500
	// messageTrancated is displayed at the end of a trancated message.
	messageTrancated = "<truncated>"
)

var (
	errMaxNodeProvisionTimeProviderNotSet = errors.New("MaxNodeProvisionTimeProvider was not set in cluster state")
)

// ScaleUpRequest contains information about the requested node group scale up.
type ScaleUpRequest struct {
	// NodeGroup is the node group to be scaled up.
	NodeGroup cloudprovider.NodeGroup
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
	// NodeGroup is the node group of the deleted node.
	NodeGroup cloudprovider.NodeGroup
	// Time is the time when the node deletion was requested.
	Time time.Time
	// ExpectedDeleteTime is the time when the node is expected to be deleted.
	ExpectedDeleteTime time.Time
}

// ClusterStateRegistryConfig contains configuration information for ClusterStateRegistry.
type ClusterStateRegistryConfig struct {
	// Maximum percentage of unready nodes in total, if the number of unready nodes is higher than OkTotalUnreadyCount.
	MaxTotalUnreadyPercentage float64
	// Minimum number of nodes that must be unready for MaxTotalUnreadyPercentage to apply.
	// This is to ensure that in very small clusters (e.g. 2 nodes) a single node's failure doesn't disable autoscaling.
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

// ScaleUpFailure contains information about a failure of a scale-up.
type ScaleUpFailure struct {
	NodeGroup cloudprovider.NodeGroup
	Reason    metrics.FailedScaleUpReason
	Time      time.Time
}

// ClusterStateRegistry is a structure to keep track the current state of the cluster.
type ClusterStateRegistry struct {
	sync.Mutex
	config                             ClusterStateRegistryConfig
	scaleUpRequests                    map[string]*ScaleUpRequest // nodeGroupName -> ScaleUpRequest
	scaleDownRequests                  []*ScaleDownRequest
	nodes                              []*apiv1.Node
	nodeInfosForGroups                 map[string]*schedulerframework.NodeInfo
	cloudProvider                      cloudprovider.CloudProvider
	perNodeGroupReadiness              map[string]Readiness
	totalReadiness                     Readiness
	acceptableRanges                   map[string]AcceptableRange
	incorrectNodeGroupSizes            map[string]IncorrectNodeGroupSize
	unregisteredNodes                  map[string]UnregisteredNode
	deletedNodes                       map[string]struct{}
	candidatesForScaleDown             map[string][]string
	backoff                            backoff.Backoff
	lastStatus                         *api.ClusterAutoscalerStatus
	lastScaleDownUpdateTime            time.Time
	logRecorder                        *utils.LogEventRecorder
	cloudProviderNodeInstances         map[string][]cloudprovider.Instance
	previousCloudProviderNodeInstances map[string][]cloudprovider.Instance
	cloudProviderNodeInstancesCache    *utils.CloudProviderNodeInstancesCache
	interrupt                          chan struct{}
	nodeGroupConfigProcessor           nodegroupconfig.NodeGroupConfigProcessor
	asyncNodeGroupStateChecker         asyncnodegroups.AsyncNodeGroupStateChecker

	// scaleUpFailures contains information about scale-up failures for each node group. It should be
	// cleared periodically to avoid unnecessary accumulation.
	scaleUpFailures map[string][]ScaleUpFailure
}

// NodeGroupScalingSafety contains information about the safety of the node group to scale up/down.
type NodeGroupScalingSafety struct {
	SafeToScale   bool
	Healthy       bool
	BackoffStatus backoff.Status
}

// NewClusterStateRegistry creates new ClusterStateRegistry.
func NewClusterStateRegistry(cloudProvider cloudprovider.CloudProvider, config ClusterStateRegistryConfig, logRecorder *utils.LogEventRecorder, backoff backoff.Backoff, nodeGroupConfigProcessor nodegroupconfig.NodeGroupConfigProcessor, asyncNodeGroupStateChecker asyncnodegroups.AsyncNodeGroupStateChecker) *ClusterStateRegistry {
	return &ClusterStateRegistry{
		scaleUpRequests:                 make(map[string]*ScaleUpRequest),
		scaleDownRequests:               make([]*ScaleDownRequest, 0),
		nodes:                           make([]*apiv1.Node, 0),
		cloudProvider:                   cloudProvider,
		config:                          config,
		perNodeGroupReadiness:           make(map[string]Readiness),
		acceptableRanges:                make(map[string]AcceptableRange),
		incorrectNodeGroupSizes:         make(map[string]IncorrectNodeGroupSize),
		unregisteredNodes:               make(map[string]UnregisteredNode),
		deletedNodes:                    make(map[string]struct{}),
		candidatesForScaleDown:          make(map[string][]string),
		backoff:                         backoff,
		lastStatus:                      utils.EmptyClusterAutoscalerStatus(),
		logRecorder:                     logRecorder,
		cloudProviderNodeInstancesCache: utils.NewCloudProviderNodeInstancesCache(cloudProvider),
		interrupt:                       make(chan struct{}),
		scaleUpFailures:                 make(map[string][]ScaleUpFailure),
		nodeGroupConfigProcessor:        nodeGroupConfigProcessor,
		asyncNodeGroupStateChecker:      asyncNodeGroupStateChecker,
	}
}

// Start starts components running in background.
func (csr *ClusterStateRegistry) Start() {
	csr.cloudProviderNodeInstancesCache.Start(csr.interrupt)
}

// Stop stops components running in background.
func (csr *ClusterStateRegistry) Stop() {
	close(csr.interrupt)
}

// RegisterScaleUp registers scale-up for give node group
func (csr *ClusterStateRegistry) RegisterScaleUp(nodeGroup cloudprovider.NodeGroup, delta int, currentTime time.Time) {
	csr.Lock()
	defer csr.Unlock()
	csr.registerOrUpdateScaleUpNoLock(nodeGroup, delta, currentTime)
}

// MaxNodeProvisionTime returns MaxNodeProvisionTime value that should be used for the given NodeGroup.
// TODO(BigDarkClown): remove this method entirely, it is a redundant wrapper
func (csr *ClusterStateRegistry) MaxNodeProvisionTime(nodeGroup cloudprovider.NodeGroup) (time.Duration, error) {
	return csr.nodeGroupConfigProcessor.GetMaxNodeProvisionTime(nodeGroup)
}

func (csr *ClusterStateRegistry) registerOrUpdateScaleUpNoLock(nodeGroup cloudprovider.NodeGroup, delta int, currentTime time.Time) {
	maxNodeProvisionTime, err := csr.MaxNodeProvisionTime(nodeGroup)
	if err != nil {
		klog.Warningf("Couldn't update scale up request: failed to get maxNodeProvisionTime for node group %s: %v", nodeGroup.Id(), err)
		return
	}

	scaleUpRequest, found := csr.scaleUpRequests[nodeGroup.Id()]
	if !found && delta > 0 {
		scaleUpRequest = &ScaleUpRequest{
			NodeGroup:       nodeGroup,
			Increase:        delta,
			Time:            currentTime,
			ExpectedAddTime: currentTime.Add(maxNodeProvisionTime),
		}
		csr.scaleUpRequests[nodeGroup.Id()] = scaleUpRequest
		return
	}

	if !found {
		// delta <=0
		return
	}

	// update the old request
	if scaleUpRequest.Increase+delta <= 0 {
		// increase <= 0 means that there is no scale-up intent really
		delete(csr.scaleUpRequests, nodeGroup.Id())
		return
	}

	scaleUpRequest.Increase += delta
	if delta > 0 {
		// if we are actually adding new nodes shift Time and ExpectedAddTime
		scaleUpRequest.Time = currentTime
		scaleUpRequest.ExpectedAddTime = currentTime.Add(maxNodeProvisionTime)
	}
}

// RegisterScaleDown registers node scale down.
func (csr *ClusterStateRegistry) RegisterScaleDown(nodeGroup cloudprovider.NodeGroup,
	nodeName string, currentTime time.Time, expectedDeleteTime time.Time) {
	request := &ScaleDownRequest{
		NodeGroup:          nodeGroup,
		NodeName:           nodeName,
		Time:               currentTime,
		ExpectedDeleteTime: expectedDeleteTime,
	}
	csr.Lock()
	defer csr.Unlock()
	csr.scaleDownRequests = append(csr.scaleDownRequests, request)
}

// To be executed under a lock.
func (csr *ClusterStateRegistry) updateScaleRequests(currentTime time.Time) {
	// clean up stale backoff info
	csr.backoff.RemoveStaleBackoffData(currentTime)

	for nodeGroupName, scaleUpRequest := range csr.scaleUpRequests {
		if !csr.areThereUpcomingNodesInNodeGroup(nodeGroupName) {
			// scale up finished successfully, remove request
			delete(csr.scaleUpRequests, nodeGroupName)
			klog.V(4).Infof("Scale up in group %v finished successfully in %v",
				nodeGroupName, currentTime.Sub(scaleUpRequest.Time))
			continue
		}

		if scaleUpRequest.ExpectedAddTime.Before(currentTime) {
			klog.Warningf("Scale-up timed out for node group %v after %v",
				nodeGroupName, currentTime.Sub(scaleUpRequest.Time))
			csr.logRecorder.Eventf(apiv1.EventTypeWarning, "ScaleUpTimedOut",
				"Nodes added to group %s failed to register within %v",
				scaleUpRequest.NodeGroup.Id(), currentTime.Sub(scaleUpRequest.Time))
			availableGPUTypes := csr.cloudProvider.GetAvailableGPUTypes()
			gpuResource, gpuType := "", ""
			nodeInfo, err := scaleUpRequest.NodeGroup.TemplateNodeInfo()
			if err != nil {
				klog.Warningf("Failed to get template node info for a node group: %s", err)
			} else {
				gpuResource, gpuType = gpu.GetGpuInfoForMetrics(csr.cloudProvider.GetNodeGpuConfig(nodeInfo.Node()), availableGPUTypes, nodeInfo.Node(), scaleUpRequest.NodeGroup)
			}
			csr.registerFailedScaleUpNoLock(scaleUpRequest.NodeGroup, metrics.Timeout, cloudprovider.InstanceErrorInfo{
				ErrorClass:   cloudprovider.OtherErrorClass,
				ErrorCode:    "timeout",
				ErrorMessage: fmt.Sprintf("Scale-up timed out for node group %v after %v", nodeGroupName, currentTime.Sub(scaleUpRequest.Time)),
			}, gpuResource, gpuType, currentTime)
			delete(csr.scaleUpRequests, nodeGroupName)
		}
	}

	newScaleDownRequests := make([]*ScaleDownRequest, 0)
	for _, scaleDownRequest := range csr.scaleDownRequests {
		if scaleDownRequest.ExpectedDeleteTime.After(currentTime) {
			newScaleDownRequests = append(newScaleDownRequests, scaleDownRequest)
		}
	}
	csr.scaleDownRequests = newScaleDownRequests
}

// To be executed under a lock.
func (csr *ClusterStateRegistry) backoffNodeGroup(nodeGroup cloudprovider.NodeGroup, errorInfo cloudprovider.InstanceErrorInfo, currentTime time.Time) {
	nodeGroupInfo := csr.nodeInfosForGroups[nodeGroup.Id()]
	backoffUntil := csr.backoff.Backoff(nodeGroup, nodeGroupInfo, errorInfo, currentTime)
	klog.Warningf("Disabling scale-up for node group %v until %v; errorClass=%v; errorCode=%v", nodeGroup.Id(), backoffUntil, errorInfo.ErrorClass, errorInfo.ErrorCode)
}

// RegisterFailedScaleUp should be called after getting error from cloudprovider
// when trying to scale-up node group. It will mark this group as not safe to autoscale
// for some time.
func (csr *ClusterStateRegistry) RegisterFailedScaleUp(nodeGroup cloudprovider.NodeGroup, reason string, errorMessage, gpuResourceName, gpuType string, currentTime time.Time) {
	csr.Lock()
	defer csr.Unlock()
	csr.registerFailedScaleUpNoLock(nodeGroup, metrics.FailedScaleUpReason(reason), cloudprovider.InstanceErrorInfo{
		ErrorClass:   cloudprovider.OtherErrorClass,
		ErrorCode:    string(reason),
		ErrorMessage: errorMessage,
	}, gpuResourceName, gpuType, currentTime)
}

// RegisterFailedScaleDown records failed scale-down for a nodegroup.
// We don't need to implement this function for cluster state registry
func (csr *ClusterStateRegistry) RegisterFailedScaleDown(_ cloudprovider.NodeGroup, _ string, _ time.Time) {
}

func (csr *ClusterStateRegistry) registerFailedScaleUpNoLock(nodeGroup cloudprovider.NodeGroup, reason metrics.FailedScaleUpReason, errorInfo cloudprovider.InstanceErrorInfo, gpuResourceName, gpuType string, currentTime time.Time) {
	csr.scaleUpFailures[nodeGroup.Id()] = append(csr.scaleUpFailures[nodeGroup.Id()], ScaleUpFailure{NodeGroup: nodeGroup, Reason: reason, Time: currentTime})
	metrics.RegisterFailedScaleUp(reason, gpuResourceName, gpuType)
	csr.backoffNodeGroup(nodeGroup, errorInfo, currentTime)
}

// UpdateNodes updates the state of the nodes in the ClusterStateRegistry and recalculates the stats
func (csr *ClusterStateRegistry) UpdateNodes(nodes []*apiv1.Node, nodeInfosForGroups map[string]*schedulerframework.NodeInfo, currentTime time.Time) error {
	csr.updateNodeGroupMetrics()
	targetSizes, err := getTargetSizes(csr.cloudProvider)
	if err != nil {
		return err
	}
	metrics.UpdateNodeGroupTargetSize(targetSizes)

	cloudProviderNodeInstances, err := csr.getCloudProviderNodeInstances()
	if err != nil {
		return err
	}
	cloudProviderNodesRemoved := csr.getCloudProviderDeletedNodes(nodes)
	notRegistered := getNotRegisteredNodes(nodes, cloudProviderNodeInstances, currentTime)

	csr.Lock()
	defer csr.Unlock()

	csr.nodes = nodes
	csr.nodeInfosForGroups = nodeInfosForGroups
	csr.previousCloudProviderNodeInstances = csr.cloudProviderNodeInstances
	csr.cloudProviderNodeInstances = cloudProviderNodeInstances

	csr.updateUnregisteredNodes(notRegistered)
	csr.updateCloudProviderDeletedNodes(cloudProviderNodesRemoved)
	csr.updateReadinessStats(currentTime)

	// update acceptable ranges based on requests from last loop and targetSizes
	// updateScaleRequests relies on acceptableRanges being up to date
	csr.updateAcceptableRanges(targetSizes)
	csr.updateScaleRequests(currentTime)
	csr.handleInstanceCreationErrors(currentTime)
	//  recalculate acceptable ranges after removing timed out requests
	csr.updateAcceptableRanges(targetSizes)
	csr.updateIncorrectNodeGroupSizes(currentTime)
	return nil
}

// Recalculate cluster state after scale-ups or scale-downs were registered.
func (csr *ClusterStateRegistry) Recalculate() {
	targetSizes, err := getTargetSizes(csr.cloudProvider)
	if err != nil {
		klog.Warningf("Failed to get target sizes, when trying to recalculate cluster state: %v", err)
	}

	csr.Lock()
	defer csr.Unlock()
	csr.updateAcceptableRanges(targetSizes)
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

	totalUnready := len(csr.totalReadiness.Unready)

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
		klog.Warningf("Failed to find acceptable ranges for %v", nodeGroupName)
		return false
	}

	readiness, found := csr.perNodeGroupReadiness[nodeGroupName]
	if !found {
		// No nodes but target == 0 or just scaling up.
		if acceptable.CurrentTarget == 0 || (acceptable.MinNodes == 0 && acceptable.CurrentTarget > 0) {
			return true
		}
		klog.Warningf("Failed to find readiness information for %v", nodeGroupName)
		return false
	}

	unjustifiedUnready := 0
	// Too few nodes, something is missing. Below the expected node count.
	if len(readiness.Ready) < acceptable.MinNodes {
		unjustifiedUnready += acceptable.MinNodes - len(readiness.Ready)
	}
	// TODO: verify against max nodes as well.
	if unjustifiedUnready > csr.config.OkTotalUnreadyCount &&
		float64(unjustifiedUnready) > csr.config.MaxTotalUnreadyPercentage/100.0*
			float64(len(readiness.Ready)+len(readiness.Unready)+len(readiness.NotStarted)) {
		return false
	}

	return true
}

// updateNodeGroupMetrics looks at NodeGroups provided by cloudprovider and updates corresponding metrics
func (csr *ClusterStateRegistry) updateNodeGroupMetrics() {
	autoscaled := 0
	autoprovisioned := 0
	for _, nodeGroup := range csr.cloudProvider.NodeGroups() {
		if !nodeGroup.Exist() {
			continue
		}
		if nodeGroup.Autoprovisioned() {
			autoprovisioned++
		} else {
			autoscaled++
		}
	}
	metrics.UpdateNodeGroupsCount(autoscaled, autoprovisioned)
}

// BackoffStatusForNodeGroup queries the backoff status of the node group
func (csr *ClusterStateRegistry) BackoffStatusForNodeGroup(nodeGroup cloudprovider.NodeGroup, now time.Time) backoff.Status {
	return csr.backoff.BackoffStatus(nodeGroup, csr.nodeInfosForGroups[nodeGroup.Id()], now)
}

// NodeGroupScaleUpSafety returns information about node group safety to be scaled up now.
func (csr *ClusterStateRegistry) NodeGroupScaleUpSafety(nodeGroup cloudprovider.NodeGroup, now time.Time) NodeGroupScalingSafety {
	isHealthy := csr.IsNodeGroupHealthy(nodeGroup.Id())
	backoffStatus := csr.backoff.BackoffStatus(nodeGroup, csr.nodeInfosForGroups[nodeGroup.Id()], now)
	return NodeGroupScalingSafety{SafeToScale: isHealthy && !backoffStatus.IsBackedOff, Healthy: isHealthy, BackoffStatus: backoffStatus}
}

func (csr *ClusterStateRegistry) getProvisionedAndTargetSizesForNodeGroup(nodeGroupName string) (provisioned, target int, ok bool) {
	if len(csr.acceptableRanges) == 0 {
		klog.Warningf("AcceptableRanges have not been populated yet. Skip checking")
		return 0, 0, false
	}

	acceptable, found := csr.acceptableRanges[nodeGroupName]
	if !found {
		klog.Warningf("Failed to find acceptable ranges for %v", nodeGroupName)
		return 0, 0, false
	}
	target = acceptable.CurrentTarget

	readiness, found := csr.perNodeGroupReadiness[nodeGroupName]
	if !found {
		// No need to warn if node group has size 0 (was scaled to 0 before).
		if acceptable.MinNodes != 0 {
			klog.Warningf("Failed to find readiness information for %v", nodeGroupName)
		}
		return 0, target, true
	}
	provisioned = len(readiness.Registered) - len(readiness.NotStarted)

	return provisioned, target, true
}

func (csr *ClusterStateRegistry) areThereUpcomingNodesInNodeGroup(nodeGroupName string) bool {
	provisioned, target, ok := csr.getProvisionedAndTargetSizesForNodeGroup(nodeGroupName)
	if !ok {
		return false
	}
	return target > provisioned
}

// IsNodeGroupAtTargetSize returns true if the number of nodes provisioned in the group is equal to the target number of nodes.
func (csr *ClusterStateRegistry) IsNodeGroupAtTargetSize(nodeGroupName string) bool {
	provisioned, target, ok := csr.getProvisionedAndTargetSizesForNodeGroup(nodeGroupName)
	if !ok {
		return false
	}
	return target == provisioned
}

// IsNodeGroupScalingUp returns true if the node group is currently scaling up.
func (csr *ClusterStateRegistry) IsNodeGroupScalingUp(nodeGroupName string) bool {
	if !csr.areThereUpcomingNodesInNodeGroup(nodeGroupName) {
		return false
	}
	_, found := csr.scaleUpRequests[nodeGroupName]
	return found
}

// HasNodeGroupStartedScaleUp returns true if the node group has started scale up regardless
// of whether there are any upcoming nodes. This is useful in the case when the node group's
// size reverts back to its previous size before the next UpdatesCall and we want to know
// if a scale up for node group has started.
func (csr *ClusterStateRegistry) HasNodeGroupStartedScaleUp(nodeGroupName string) bool {
	csr.Lock()
	defer csr.Unlock()
	_, found := csr.scaleUpRequests[nodeGroupName]
	return found
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
		readiness := csr.perNodeGroupReadiness[nodeGroup.Id()]
		result[nodeGroup.Id()] = AcceptableRange{
			MinNodes:      size - len(readiness.LongUnregistered),
			MaxNodes:      size,
			CurrentTarget: size,
		}
	}
	for nodeGroupName, scaleUpRequest := range csr.scaleUpRequests {
		acceptableRange := result[nodeGroupName]
		acceptableRange.MinNodes -= scaleUpRequest.Increase
		result[nodeGroupName] = acceptableRange
	}
	for _, scaleDownRequest := range csr.scaleDownRequests {
		acceptableRange := result[scaleDownRequest.NodeGroup.Id()]
		acceptableRange.MaxNodes++
		result[scaleDownRequest.NodeGroup.Id()] = acceptableRange
	}
	csr.acceptableRanges = result
}

// Readiness contains readiness information about a group of nodes.
type Readiness struct {
	// Names of ready nodes.
	Ready []string
	// Names of unready nodes that broke down after they started.
	Unready []string
	// Names of nodes that are being currently deleted. They exist in K8S but
	// are not included in NodeGroup.TargetSize().
	Deleted []string
	// Names of nodes that are not yet fully started.
	NotStarted []string
	// Names of all registered nodes in the group (ready/unready/deleted/etc).
	Registered []string
	// Names of nodes that failed to register within a reasonable limit.
	LongUnregistered []string
	// Names of nodes that haven't yet registered.
	Unregistered []string
	// Time when the readiness was measured.
	Time time.Time
	// Names of nodes that are Unready due to missing resources.
	// This field is only used for exposing information externally and
	// doesn't influence CA behavior.
	ResourceUnready []string
}

func (csr *ClusterStateRegistry) updateReadinessStats(currentTime time.Time) {
	perNodeGroup := make(map[string]Readiness)
	total := Readiness{Time: currentTime}

	update := func(current Readiness, node *apiv1.Node, nr kube_util.NodeReadiness) Readiness {
		current.Registered = append(current.Registered, node.Name)
		if _, isDeleted := csr.deletedNodes[node.Name]; isDeleted {
			current.Deleted = append(current.Deleted, node.Name)
		} else if nr.Ready {
			current.Ready = append(current.Ready, node.Name)
		} else if node.CreationTimestamp.Time.Add(MaxNodeStartupTime).After(currentTime) {
			current.NotStarted = append(current.NotStarted, node.Name)
		} else {
			current.Unready = append(current.Unready, node.Name)
			if nr.Reason == kube_util.ResourceUnready {
				current.ResourceUnready = append(current.ResourceUnready, node.Name)
			}
		}
		return current
	}

	for _, node := range csr.nodes {
		nodeGroup, errNg := csr.cloudProvider.NodeGroupForNode(node)
		nr, errReady := kube_util.GetNodeReadiness(node)

		// Node is most likely not autoscaled, however check the errors.
		if nodeGroup == nil || reflect.ValueOf(nodeGroup).IsNil() {
			if errNg != nil {
				klog.Warningf("Failed to get nodegroup for %s: %v", node.Name, errNg)
			}
			if errReady != nil {
				klog.Warningf("Failed to get readiness info for %s: %v", node.Name, errReady)
			}
		} else {
			perNodeGroup[nodeGroup.Id()] = update(perNodeGroup[nodeGroup.Id()], node, nr)
		}
		total = update(total, node, nr)
	}

	for _, unregistered := range csr.unregisteredNodes {
		nodeGroup, errNg := csr.cloudProvider.NodeGroupForNode(unregistered.Node)
		if errNg != nil {
			klog.Warningf("Failed to get nodegroup for %s: %v", unregistered.Node.Name, errNg)
			continue
		}
		if nodeGroup == nil || reflect.ValueOf(nodeGroup).IsNil() {
			klog.Warningf("Nodegroup is nil for %s", unregistered.Node.Name)
			continue
		}
		perNgCopy := perNodeGroup[nodeGroup.Id()]
		maxNodeProvisionTime, err := csr.MaxNodeProvisionTime(nodeGroup)
		if err != nil {
			klog.Warningf("Failed to get maxNodeProvisionTime for node %s in node group %s: %v", unregistered.Node.Name, nodeGroup.Id(), err)
			continue
		}
		if unregistered.UnregisteredSince.Add(maxNodeProvisionTime).Before(currentTime) {
			perNgCopy.LongUnregistered = append(perNgCopy.LongUnregistered, unregistered.Node.Name)
			total.LongUnregistered = append(total.LongUnregistered, unregistered.Node.Name)
		} else {
			perNgCopy.Unregistered = append(perNgCopy.Unregistered, unregistered.Node.Name)
			total.Unregistered = append(total.Unregistered, unregistered.Node.Name)
		}
		perNodeGroup[nodeGroup.Id()] = perNgCopy
	}
	if len(total.LongUnregistered) > 0 {
		klog.V(3).Infof("Found longUnregistered Nodes %s", total.LongUnregistered)
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
			klog.Warningf("Acceptable range for node group %s not found", nodeGroup.Id())
			continue
		}
		if csr.asyncNodeGroupStateChecker.IsUpcoming(nodeGroup) {
			// Nodes for upcoming node groups reside in-memory and wait for node group to be fully
			// created. There is no need to mark their sizes incorrect.
			continue
		}
		readiness, found := csr.perNodeGroupReadiness[nodeGroup.Id()]
		if !found {
			// if MinNodes == 0 node group has been scaled to 0 and everything's fine
			if acceptableRange.MinNodes != 0 {
				klog.Warningf("Readiness for node group %s not found", nodeGroup.Id())
			}
			continue
		}
		unregisteredNodes := len(readiness.Unregistered) + len(readiness.LongUnregistered)
		if len(readiness.Registered) > acceptableRange.CurrentTarget ||
			len(readiness.Registered) < acceptableRange.CurrentTarget-unregisteredNodes {
			incorrect := IncorrectNodeGroupSize{
				CurrentSize:   len(readiness.Registered),
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

// GetUnregisteredNodes returns a list of all unregistered nodes.
func (csr *ClusterStateRegistry) GetUnregisteredNodes() []UnregisteredNode {
	csr.Lock()
	defer csr.Unlock()

	result := make([]UnregisteredNode, 0, len(csr.unregisteredNodes))
	for _, unregistered := range csr.unregisteredNodes {
		result = append(result, unregistered)
	}
	return result
}

func (csr *ClusterStateRegistry) updateCloudProviderDeletedNodes(deletedNodes []*apiv1.Node) {
	result := make(map[string]struct{}, len(deletedNodes))
	for _, deleted := range deletedNodes {
		result[deleted.Name] = struct{}{}
	}
	csr.deletedNodes = result
}

// UpdateScaleDownCandidates updates scale down candidates
func (csr *ClusterStateRegistry) UpdateScaleDownCandidates(nodes []*apiv1.Node, now time.Time) {
	result := make(map[string][]string)
	for _, node := range nodes {
		group, err := csr.cloudProvider.NodeGroupForNode(node)
		if err != nil {
			klog.Warningf("Failed to get node group for %s: %v", node.Name, err)
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
		AutoscalerStatus: api.ClusterAutoscalerRunning,
		NodeGroups:       make([]api.NodeGroupStatus, 0),
	}
	nodeGroupsLastStatus := make(map[string]api.NodeGroupStatus)
	for _, nodeGroup := range csr.lastStatus.NodeGroups {
		nodeGroupsLastStatus[nodeGroup.Name] = nodeGroup
	}
	for _, nodeGroup := range csr.cloudProvider.NodeGroups() {
		nodeGroupStatus := api.NodeGroupStatus{
			Name: nodeGroup.Id(),
		}
		readiness := csr.perNodeGroupReadiness[nodeGroup.Id()]
		acceptable := csr.acceptableRanges[nodeGroup.Id()]

		nodeGroupLastStatus := nodeGroupsLastStatus[nodeGroup.Id()]

		// Health.
		nodeGroupStatus.Health = buildHealthStatusNodeGroup(
			csr.IsNodeGroupHealthy(nodeGroup.Id()), readiness, acceptable, nodeGroup.MinSize(), nodeGroup.MaxSize(), nodeGroupLastStatus.Health)

		// Scale up.
		nodeGroupStatus.ScaleUp = csr.buildScaleUpStatusNodeGroup(
			nodeGroup,
			readiness,
			acceptable, now, nodeGroupLastStatus.ScaleUp)

		// Scale down.
		nodeGroupStatus.ScaleDown = buildScaleDownStatusNodeGroup(
			csr.candidatesForScaleDown[nodeGroup.Id()], csr.lastScaleDownUpdateTime, nodeGroupLastStatus.ScaleDown)

		result.NodeGroups = append(result.NodeGroups, nodeGroupStatus)
	}
	result.ClusterWide.Health =
		buildHealthStatusClusterwide(csr.IsClusterHealthy(), csr.totalReadiness, csr.lastStatus.ClusterWide.Health)
	result.ClusterWide.ScaleUp =
		buildScaleUpStatusClusterwide(result.NodeGroups, csr.totalReadiness, csr.lastStatus.ClusterWide.ScaleUp)
	result.ClusterWide.ScaleDown =
		buildScaleDownStatusClusterwide(csr.candidatesForScaleDown, csr.lastScaleDownUpdateTime, csr.lastStatus.ClusterWide.ScaleDown)

	csr.lastStatus = result
	return result
}

// GetClusterReadiness returns current readiness stats of cluster
func (csr *ClusterStateRegistry) GetClusterReadiness() Readiness {
	return csr.totalReadiness
}

func buildNodeCount(readiness Readiness) api.NodeCount {
	return api.NodeCount{
		Registered: api.RegisteredNodeCount{
			Total:        len(readiness.Registered),
			Ready:        len(readiness.Ready),
			NotStarted:   len(readiness.NotStarted),
			BeingDeleted: len(readiness.Deleted),
			Unready: api.RegisteredUnreadyNodeCount{
				Total:           len(readiness.Unready),
				ResourceUnready: len(readiness.ResourceUnready),
			},
		},
		LongUnregistered: len(readiness.LongUnregistered),
		Unregistered:     len(readiness.Unregistered),
	}
}

func buildHealthStatusNodeGroup(isHealthy bool, readiness Readiness, acceptable AcceptableRange, minSize, maxSize int, lastStatus api.NodeGroupHealthCondition) api.NodeGroupHealthCondition {
	condition := api.NodeGroupHealthCondition{
		NodeCounts:          buildNodeCount(readiness),
		CloudProviderTarget: acceptable.CurrentTarget,
		MinSize:             minSize,
		MaxSize:             maxSize,
		LastProbeTime:       metav1.Time{Time: readiness.Time},
	}
	if isHealthy {
		condition.Status = api.ClusterAutoscalerHealthy
	} else {
		condition.Status = api.ClusterAutoscalerUnhealthy
	}
	if condition.Status == lastStatus.Status {
		condition.LastTransitionTime = lastStatus.LastTransitionTime
	} else {
		condition.LastTransitionTime = condition.LastProbeTime
	}
	return condition
}

func (csr *ClusterStateRegistry) buildScaleUpStatusNodeGroup(nodeGroup cloudprovider.NodeGroup, readiness Readiness, acceptable AcceptableRange, now time.Time, lastStatus api.NodeGroupScaleUpCondition) api.NodeGroupScaleUpCondition {
	isScaleUpInProgress := csr.IsNodeGroupScalingUp(nodeGroup.Id())
	scaleUpSafety := csr.NodeGroupScaleUpSafety(nodeGroup, now)
	condition := api.NodeGroupScaleUpCondition{
		LastProbeTime: metav1.Time{Time: readiness.Time},
	}
	if isScaleUpInProgress {
		condition.Status = api.ClusterAutoscalerInProgress
	} else if !scaleUpSafety.Healthy {
		condition.Status = api.ClusterAutoscalerUnhealthy
	} else if !scaleUpSafety.SafeToScale {
		condition.Status = api.ClusterAutoscalerBackoff
		condition.BackoffInfo = api.BackoffInfo{
			ErrorCode:    scaleUpSafety.BackoffStatus.ErrorInfo.ErrorCode,
			ErrorMessage: truncateIfExceedMaxLength(scaleUpSafety.BackoffStatus.ErrorInfo.ErrorMessage, maxErrorMessageSize),
		}
	} else {
		condition.Status = api.ClusterAutoscalerNoActivity
	}
	if condition.Status == lastStatus.Status {
		condition.LastTransitionTime = lastStatus.LastTransitionTime
	} else {
		condition.LastTransitionTime = condition.LastProbeTime
	}
	return condition
}

func buildScaleDownStatusNodeGroup(candidates []string, lastProbed time.Time, lastStatus api.ScaleDownCondition) api.ScaleDownCondition {
	condition := api.ScaleDownCondition{
		Candidates:    len(candidates),
		LastProbeTime: metav1.Time{Time: lastProbed},
	}
	if len(candidates) > 0 {
		condition.Status = api.ClusterAutoscalerCandidatesPresent
	} else {
		condition.Status = api.ClusterAutoscalerNoCandidates
	}
	if condition.Status == lastStatus.Status {
		condition.LastTransitionTime = lastStatus.LastTransitionTime
	} else {
		condition.LastTransitionTime = condition.LastProbeTime
	}
	return condition
}

func buildHealthStatusClusterwide(isHealthy bool, readiness Readiness, lastStatus api.ClusterHealthCondition) api.ClusterHealthCondition {
	condition := api.ClusterHealthCondition{
		NodeCounts:    buildNodeCount(readiness),
		LastProbeTime: metav1.Time{Time: readiness.Time},
	}
	if isHealthy {
		condition.Status = api.ClusterAutoscalerHealthy
	} else {
		condition.Status = api.ClusterAutoscalerUnhealthy
	}
	if condition.Status == lastStatus.Status {
		condition.LastTransitionTime = lastStatus.LastTransitionTime
	} else {
		condition.LastTransitionTime = condition.LastProbeTime
	}
	return condition
}

func buildScaleUpStatusClusterwide(nodeGroupsStatuses []api.NodeGroupStatus, readiness Readiness, lastStatus api.ClusterScaleUpCondition) api.ClusterScaleUpCondition {
	isScaleUpInProgress := false
	for _, nodeGroupStatus := range nodeGroupsStatuses {
		if nodeGroupStatus.ScaleUp.Status == api.ClusterAutoscalerInProgress {
			isScaleUpInProgress = true
			break
		}
	}
	condition := api.ClusterScaleUpCondition{
		LastProbeTime: metav1.Time{Time: readiness.Time},
	}
	if isScaleUpInProgress {
		condition.Status = api.ClusterAutoscalerInProgress
	} else {
		condition.Status = api.ClusterAutoscalerNoActivity
	}
	if condition.Status == lastStatus.Status {
		condition.LastTransitionTime = lastStatus.LastTransitionTime
	} else {
		condition.LastTransitionTime = condition.LastProbeTime
	}
	return condition
}

func buildScaleDownStatusClusterwide(candidates map[string][]string, lastProbed time.Time, lastStatus api.ScaleDownCondition) api.ScaleDownCondition {
	totalCandidates := 0
	for _, val := range candidates {
		totalCandidates += len(val)
	}
	condition := api.ScaleDownCondition{
		Candidates:    totalCandidates,
		LastProbeTime: metav1.Time{Time: lastProbed},
	}
	if totalCandidates > 0 {
		condition.Status = api.ClusterAutoscalerCandidatesPresent
	} else {
		condition.Status = api.ClusterAutoscalerNoCandidates
	}
	if condition.Status == lastStatus.Status {
		condition.LastTransitionTime = lastStatus.LastTransitionTime
	} else {
		condition.LastTransitionTime = condition.LastProbeTime
	}
	return condition
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
// The function may overestimate the number of nodes. The second return value contains the names of upcoming nodes
// that are already registered in the cluster.
func (csr *ClusterStateRegistry) GetUpcomingNodes() (upcomingCounts map[string]int, registeredNodeNames map[string][]string) {
	csr.Lock()
	defer csr.Unlock()

	upcomingCounts = map[string]int{}
	registeredNodeNames = map[string][]string{}
	for _, nodeGroup := range csr.cloudProvider.NodeGroups() {
		id := nodeGroup.Id()
		if csr.asyncNodeGroupStateChecker.IsUpcoming(nodeGroup) {
			size, err := nodeGroup.TargetSize()
			if size >= 0 || err != nil {
				upcomingCounts[id] = size
			}
			continue
		}
		readiness := csr.perNodeGroupReadiness[id]
		ar := csr.acceptableRanges[id]
		// newNodes is the number of nodes that
		newNodes := ar.CurrentTarget - (len(readiness.Ready) + len(readiness.Unready) + len(readiness.LongUnregistered))
		if newNodes <= 0 {
			// Negative value is unlikely but theoretically possible.
			continue
		}
		upcomingCounts[id] = newNodes
		// newNodes should include instances that have registered with k8s but aren't ready yet, instances that came up on the cloud provider side
		// but haven't registered with k8s yet, and instances that haven't even come up on the cloud provider side yet (but are reflected in the target
		// size). The first category is categorized as NotStarted in readiness, the other two aren't registered with k8s, so they shouldn't be
		// included.
		registeredNodeNames[id] = readiness.NotStarted
	}
	return upcomingCounts, registeredNodeNames
}

// getCloudProviderNodeInstances returns map keyed on node group id where value is list of node instances
// as returned by NodeGroup.Nodes().
func (csr *ClusterStateRegistry) getCloudProviderNodeInstances() (map[string][]cloudprovider.Instance, error) {
	for _, nodeGroup := range csr.cloudProvider.NodeGroups() {
		if csr.IsNodeGroupScalingUp(nodeGroup.Id()) {
			csr.cloudProviderNodeInstancesCache.InvalidateCacheEntry(nodeGroup)
		}
	}
	return csr.cloudProviderNodeInstancesCache.GetCloudProviderNodeInstances()
}

// Calculates which of the existing cloud provider nodes are not yet registered in Kubernetes.
// As we are expecting for those instances to be Ready soon (O(~minutes)), to speed up the scaling process,
// we are injecting a temporary, fake nodes to continue scaling based on in-memory cluster state.
func getNotRegisteredNodes(allNodes []*apiv1.Node, cloudProviderNodeInstances map[string][]cloudprovider.Instance, time time.Time) []UnregisteredNode {
	registered := sets.NewString()
	for _, node := range allNodes {
		registered.Insert(node.Spec.ProviderID)
	}
	notRegistered := make([]UnregisteredNode, 0)
	for _, instances := range cloudProviderNodeInstances {
		for _, instance := range instances {
			if !registered.Has(instance.Id) && expectedToRegister(instance) {
				notRegistered = append(notRegistered, UnregisteredNode{
					Node:              FakeNode(instance, cloudprovider.FakeNodeUnregistered),
					UnregisteredSince: time,
				})
			}
		}
	}
	return notRegistered
}

func expectedToRegister(instance cloudprovider.Instance) bool {
	return instance.Status == nil || (instance.Status.State != cloudprovider.InstanceDeleting && instance.Status.ErrorInfo == nil)
}

// Calculates which of the registered nodes in Kubernetes that do not exist in cloud provider.
func (csr *ClusterStateRegistry) getCloudProviderDeletedNodes(allNodes []*apiv1.Node) []*apiv1.Node {
	nodesRemoved := make([]*apiv1.Node, 0)
	for _, node := range allNodes {
		if !csr.hasCloudProviderInstance(node) {
			nodesRemoved = append(nodesRemoved, node)
		}
	}
	return nodesRemoved
}

func (csr *ClusterStateRegistry) hasCloudProviderInstance(node *apiv1.Node) bool {
	exists, err := csr.cloudProvider.HasInstance(node)
	if err == nil {
		return exists
	}
	if !errors.Is(err, cloudprovider.ErrNotImplemented) {
		klog.Warningf("Failed to check cloud provider has instance for %s: %v", node.Name, err)
	}
	return !taints.HasToBeDeletedTaint(node)
}

// GetAutoscaledNodesCount calculates and returns the actual and the target number of nodes
// belonging to autoscaled node groups in the cluster.
func (csr *ClusterStateRegistry) GetAutoscaledNodesCount() (currentSize, targetSize int) {
	csr.Lock()
	defer csr.Unlock()

	for _, accRange := range csr.acceptableRanges {
		targetSize += accRange.CurrentTarget
	}
	for _, readiness := range csr.perNodeGroupReadiness {
		currentSize += len(readiness.Registered) - len(readiness.NotStarted)
	}
	return currentSize, targetSize
}

func (csr *ClusterStateRegistry) handleInstanceCreationErrors(currentTime time.Time) {
	nodeGroups := csr.cloudProvider.NodeGroups()

	for _, nodeGroup := range nodeGroups {
		csr.handleInstanceCreationErrorsForNodeGroup(
			nodeGroup,
			csr.cloudProviderNodeInstances[nodeGroup.Id()],
			csr.previousCloudProviderNodeInstances[nodeGroup.Id()],
			currentTime)
	}
}

func (csr *ClusterStateRegistry) handleInstanceCreationErrorsForNodeGroup(
	nodeGroup cloudprovider.NodeGroup,
	currentInstances []cloudprovider.Instance,
	previousInstances []cloudprovider.Instance,
	currentTime time.Time) {

	_, currentUniqueErrorMessagesForErrorCode, currentErrorCodeToInstance := csr.buildInstanceToErrorCodeMappings(currentInstances)
	previousInstanceToErrorCode, _, _ := csr.buildInstanceToErrorCodeMappings(previousInstances)

	for errorCode, instances := range currentErrorCodeToInstance {
		if len(instances) > 0 {
			klog.V(4).Infof("Found %v instances with errorCode %v in nodeGroup %v", len(instances), errorCode, nodeGroup.Id())
		}
	}

	// If node group is scaling up and there are new node-create requests which cannot be satisfied because of
	// out-of-resources errors we:
	//  - emit event
	//  - alter the scale-up
	//  - increase scale-up failure metric
	//  - backoff the node group
	for errorCode, instances := range currentErrorCodeToInstance {
		unseenInstanceIds := make([]string, 0)
		for _, instance := range instances {
			if _, seen := previousInstanceToErrorCode[instance.Id]; !seen {
				unseenInstanceIds = append(unseenInstanceIds, instance.Id)
			}
		}

		klog.V(1).Infof("Failed adding %v nodes (%v unseen previously) to group %v due to %v; errorMessages=%#v", len(instances), len(unseenInstanceIds), nodeGroup.Id(), errorCode, currentUniqueErrorMessagesForErrorCode[errorCode])
		if len(unseenInstanceIds) > 0 && csr.IsNodeGroupScalingUp(nodeGroup.Id()) {
			csr.logRecorder.Eventf(
				apiv1.EventTypeWarning,
				"ScaleUpFailed",
				"Failed adding %v nodes to group %v due to %v; source errors: %v",
				len(unseenInstanceIds),
				nodeGroup.Id(),
				errorCode,
				csr.buildErrorMessageEventString(currentUniqueErrorMessagesForErrorCode[errorCode]))

			availableGPUTypes := csr.cloudProvider.GetAvailableGPUTypes()
			gpuResource, gpuType := "", ""
			nodeInfo, err := nodeGroup.TemplateNodeInfo()
			if err != nil {
				klog.Warningf("Failed to get template node info for a node group: %s", err)
			} else {
				gpuResource, gpuType = gpu.GetGpuInfoForMetrics(csr.cloudProvider.GetNodeGpuConfig(nodeInfo.Node()), availableGPUTypes, nodeInfo.Node(), nodeGroup)
			}
			// Decrease the scale up request by the number of deleted nodes
			csr.registerOrUpdateScaleUpNoLock(nodeGroup, -len(unseenInstanceIds), currentTime)

			csr.registerFailedScaleUpNoLock(nodeGroup, metrics.FailedScaleUpReason(errorCode.code), cloudprovider.InstanceErrorInfo{
				ErrorClass:   errorCode.class,
				ErrorCode:    errorCode.code,
				ErrorMessage: csr.buildErrorMessageEventString(currentUniqueErrorMessagesForErrorCode[errorCode]),
			}, gpuResource, gpuType, currentTime)
		}
	}
}

func (csr *ClusterStateRegistry) buildErrorMessageEventString(uniqErrorMessages []string) string {
	var sb strings.Builder
	maxErrors := 3
	for i, errorMessage := range uniqErrorMessages {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(errorMessage)

	}
	if len(uniqErrorMessages) > maxErrors {
		sb.WriteString(", ...")
	}
	return sb.String()
}

type errorCode struct {
	code  string
	class cloudprovider.InstanceErrorClass
}

func (c errorCode) String() string {
	return fmt.Sprintf("%v.%v", c.class, c.code)
}

func (csr *ClusterStateRegistry) buildInstanceToErrorCodeMappings(instances []cloudprovider.Instance) (instanceToErrorCode map[string]errorCode, uniqueErrorMessagesForErrorCode map[errorCode][]string, errorCodeToInstance map[errorCode][]cloudprovider.Instance) {
	instanceToErrorCode = make(map[string]errorCode)
	uniqueErrorMessagesForErrorCode = make(map[errorCode][]string)
	errorCodeToInstance = make(map[errorCode][]cloudprovider.Instance)

	uniqErrorMessagesForErrorCodeTmp := make(map[errorCode]map[string]bool)
	for _, instance := range instances {
		if instance.Status != nil && instance.Status.State == cloudprovider.InstanceCreating && instance.Status.ErrorInfo != nil {
			errorInfo := instance.Status.ErrorInfo
			errorCode := errorCode{errorInfo.ErrorCode, errorInfo.ErrorClass}

			if _, found := uniqErrorMessagesForErrorCodeTmp[errorCode]; !found {
				uniqErrorMessagesForErrorCodeTmp[errorCode] = make(map[string]bool)
			}
			instanceToErrorCode[instance.Id] = errorCode
			uniqErrorMessagesForErrorCodeTmp[errorCode][errorInfo.ErrorMessage] = true
			errorCodeToInstance[errorCode] = append(errorCodeToInstance[errorCode], instance)
		}
	}

	for errorCode, uniqueErrorMessages := range uniqErrorMessagesForErrorCodeTmp {
		for errorMessage := range uniqueErrorMessages {
			uniqueErrorMessagesForErrorCode[errorCode] = append(uniqueErrorMessagesForErrorCode[errorCode], errorMessage)
		}
	}

	return
}

// GetCreatedNodesWithErrors returns a map from node group id to list of nodes which reported a create error.
func (csr *ClusterStateRegistry) GetCreatedNodesWithErrors() map[string][]*apiv1.Node {
	csr.Lock()
	defer csr.Unlock()

	nodesWithCreateErrors := make(map[string][]*apiv1.Node)
	for nodeGroupId, nodeGroupInstances := range csr.cloudProviderNodeInstances {
		_, _, instancesByErrorCode := csr.buildInstanceToErrorCodeMappings(nodeGroupInstances)
		for _, instances := range instancesByErrorCode {
			for _, instance := range instances {
				nodesWithCreateErrors[nodeGroupId] = append(nodesWithCreateErrors[nodeGroupId], FakeNode(instance, cloudprovider.FakeNodeCreateError))
			}
		}
	}
	return nodesWithCreateErrors
}

// RefreshCloudProviderNodeInstancesCache refreshes cloud provider node instances cache.
func (csr *ClusterStateRegistry) RefreshCloudProviderNodeInstancesCache() {
	csr.cloudProviderNodeInstancesCache.Refresh()
}

// InvalidateNodeInstancesCacheEntry removes a node group from the cloud provider node instances cache.
func (csr *ClusterStateRegistry) InvalidateNodeInstancesCacheEntry(nodeGroup cloudprovider.NodeGroup) {
	csr.cloudProviderNodeInstancesCache.InvalidateCacheEntry(nodeGroup)
}

// FakeNode creates a fake node with Name field populated and FakeNodeReasonAnnotation added
func FakeNode(instance cloudprovider.Instance, reason string) *apiv1.Node {
	return &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: instance.Id,
			Annotations: map[string]string{
				cloudprovider.FakeNodeReasonAnnotation: reason,
			},
		},
		Spec: apiv1.NodeSpec{
			ProviderID: instance.Id,
		},
	}
}

// PeriodicCleanup performs clean-ups that should be done periodically, e.g.
// each Autoscaler loop.
func (csr *ClusterStateRegistry) PeriodicCleanup() {
	// Clear the scale-up failures info so they don't accumulate.
	csr.clearScaleUpFailures()
}

// clearScaleUpFailures clears the scale-up failures map.
func (csr *ClusterStateRegistry) clearScaleUpFailures() {
	csr.Lock()
	defer csr.Unlock()
	csr.scaleUpFailures = make(map[string][]ScaleUpFailure)
}

// GetScaleUpFailures returns the scale-up failures map.
func (csr *ClusterStateRegistry) GetScaleUpFailures() map[string][]ScaleUpFailure {
	csr.Lock()
	defer csr.Unlock()
	result := make(map[string][]ScaleUpFailure)
	for nodeGroupId, failures := range csr.scaleUpFailures {
		result[nodeGroupId] = failures
	}
	return result
}

func truncateIfExceedMaxLength(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}
	untrancatedLen := maxLength - len(messageTrancated)
	if untrancatedLen < 0 {
		return s[:maxLength]
	}
	return s[:untrancatedLen] + messageTrancated
}
