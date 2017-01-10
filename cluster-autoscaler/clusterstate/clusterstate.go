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

	"k8s.io/contrib/cluster-autoscaler/cloudprovider"
	"k8s.io/contrib/cluster-autoscaler/utils/deletetaint"
	apiv1 "k8s.io/kubernetes/pkg/api/v1"

	"github.com/golang/glog"
)

const (
	// MaxNodeStartupTime is the maximum time from the moment the node is registered to the time the node is ready.
	MaxNodeStartupTime = 5 * time.Minute

	// MaxStatusSettingDelayAfterCreation is the maximum time for node to set its initial status after the
	// node is registered.
	MaxStatusSettingDelayAfterCreation = time.Minute
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
	// Number of nodes that can be unready in total. If the number is higer than that then MaxTotalUnreadyPercentage applies.
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
	// FirstObserved is the time whtn the given difference occurred.
	FirstObserved time.Time
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
}

// NewClusterStateRegistry creates new ClusterStateRegistry.
func NewClusterStateRegistry(cloudProvider cloudprovider.CloudProvider, config ClusterStateRegistryConfig) *ClusterStateRegistry {
	return &ClusterStateRegistry{
		scaleUpRequests:       make([]*ScaleUpRequest, 0),
		scaleDownRequests:     make([]*ScaleDownRequest, 0),
		nodes:                 make([]*apiv1.Node, 0),
		cloudProvider:         cloudProvider,
		config:                config,
		perNodeGroupReadiness: make(map[string]Readiness),
		acceptableRanges:      make(map[string]AcceptableRange),
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

	csr.Lock()
	defer csr.Unlock()

	csr.cleanUp(currentTime)
	csr.nodes = nodes
	csr.updateReadinessStats(currentTime)
	csr.updateAcceptableRanges(targetSizes)
	csr.updateIncorrectNodeGroupSizes(currentTime)
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
func (csr *ClusterStateRegistry) IsClusterHealthy(currentTime time.Time) bool {
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
	readiness, found := csr.perNodeGroupReadiness[nodeGroupName]
	if !found {
		glog.Warningf("Failed to find readiness information for %v", nodeGroupName)
		return false
	}
	acceptable, found := csr.acceptableRanges[nodeGroupName]
	if !found {
		glog.Warningf("Failed to find acceptable ranges for %v", nodeGroupName)
		return false
	}

	unjustifiedUnready := 0
	// Too few nodes, something is missing. Below the expected node count.
	if readiness.Ready < acceptable.MinNodes {
		unjustifiedUnready += acceptable.MinNodes - readiness.Ready
	}
	// TODO: verify against maxnodes as well.

	glog.V(2).Infof("NodeGroupHealth %s: ready=%d, acceptable min=%d max=%d target=%d",
		nodeGroupName,
		readiness.Ready,
		acceptable.MinNodes,
		acceptable.MaxNodes,
		acceptable.CurrentTarget,
	)

	if unjustifiedUnready > csr.config.OkTotalUnreadyCount &&
		float64(unjustifiedUnready) > csr.config.MaxTotalUnreadyPercentage/100.0*
			float64(readiness.Ready+readiness.Unready+readiness.NotStarted+readiness.LongNotStarted) {
		return false
	}

	return true
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

// calculateAcceptableRanges calcualtes how many nodes can be in a cluster.
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
}

func (csr *ClusterStateRegistry) updateReadinessStats(currentTime time.Time) {

	perNodeGroup := make(map[string]Readiness)
	total := Readiness{}

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
		ready, _, errReady := GetReadinessState(node)

		// Node is most likely not autoscaled, however check the errors.
		if reflect.ValueOf(nodeGroup).IsNil() {
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
	csr.perNodeGroupReadiness = perNodeGroup
	csr.totalReadiness = total
}

// Calculates which node groups have incorrect size.
func (csr *ClusterStateRegistry) updateIncorrectNodeGroupSizes(currentTime time.Time) {
	result := make(map[string]IncorrectNodeGroupSize)
	for _, nodeGroup := range csr.cloudProvider.NodeGroups() {
		readiness, found := csr.perNodeGroupReadiness[nodeGroup.Id()]
		if !found {
			glog.Warningf("Readiness for node group %s not found", nodeGroup.Id())
			continue
		}
		acceptableRange, found := csr.acceptableRanges[nodeGroup.Id()]
		if !found {
			glog.Warningf("Acceptable range for node group %s not found", nodeGroup.Id())
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

// GetReadinessState gets readiness state for the node
func GetReadinessState(node *apiv1.Node) (isNodeReady bool, lastTransitionTime time.Time, err error) {
	for _, condition := range node.Status.Conditions {
		if condition.Type == apiv1.NodeReady {
			if condition.Status == apiv1.ConditionTrue {
				return true, condition.LastTransitionTime.Time, nil
			}
			return false, condition.LastTransitionTime.Time, nil
		}
	}
	return false, time.Time{}, fmt.Errorf("NodeReady condition for %s not found", node.Name)
}

func isNodeNotStarted(node *apiv1.Node) bool {
	for _, condition := range node.Status.Conditions {
		if condition.Type == apiv1.NodeReady &&
			condition.Status == apiv1.ConditionFalse &&
			condition.LastTransitionTime.Time.Sub(node.CreationTimestamp.Time) < MaxStatusSettingDelayAfterCreation {
			return true
		}
	}
	return false
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
			// Negative value is unlikely but theroetically possible.
			continue
		}
		result[id] = newNodes
	}
	return result
}
