/*
Copyright 2020 The Kubernetes Authors.

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

package huaweicloud

import (
	"fmt"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/klog/v2"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework/v1alpha1"
)

// AutoScalingGroup represents a HuaweiCloud's 'Auto Scaling Group' which also can be treated as a node group.
type AutoScalingGroup struct {
	cloudServiceManager CloudServiceManager

	groupID           string
	minInstanceNumber int
	maxInstanceNumber int
}

// Check if our AutoScalingGroup implements necessary interface.
var _ cloudprovider.NodeGroup = &AutoScalingGroup{}

// MaxSize returns maximum size of the node group.
func (asg *AutoScalingGroup) MaxSize() int {
	return asg.maxInstanceNumber
}

// MinSize returns minimum size of the node group.
func (asg *AutoScalingGroup) MinSize() int {
	return asg.minInstanceNumber
}

// TargetSize returns the current target size of the node group. It is possible that the
// number of nodes in Kubernetes is different at the moment but should be equal
// to Size() once everything stabilizes (new nodes finish startup and registration or
// removed nodes are deleted completely). Implementation required.
//
// Target size is desire instance number of the auto scaling group, and not equal to current instance number if the
// auto scaling group is in increasing or decreasing process.
func (asg *AutoScalingGroup) TargetSize() (int, error) {
	desireNumber, err := asg.cloudServiceManager.GetDesireInstanceNumber(asg.groupID)
	if err != nil {
		klog.Warningf("failed to get group target size. groupID: %s, error: %v", asg.groupID, err)
		return 0, err
	}

	return desireNumber, nil
}

// IncreaseSize increases the size of the node group. To delete a node you need
// to explicitly name it and use DeleteNode. This function should wait until
// node group size is updated. Implementation required.
func (asg *AutoScalingGroup) IncreaseSize(delta int) error {
	err := asg.cloudServiceManager.IncreaseSizeInstance(asg.groupID, delta)
	if err != nil {
		klog.Warningf("failed to increase size for group: %s, error: %v", asg.groupID, err)
		return err
	}

	return nil
}

// DeleteNodes deletes nodes from this node group. Error is returned either on
// failure or if the given node doesn't belong to this node group. This function
// should wait until node group size is updated. Implementation required.
func (asg *AutoScalingGroup) DeleteNodes(nodes []*apiv1.Node) error {
	instances, err := asg.cloudServiceManager.GetInstances(asg.groupID)
	if err != nil {
		klog.Warningf("failed to get nodes from group: %s, error: %v", asg.groupID, err)
		return err
	}

	// TODO(RainbowMango): Check if all node in this group.
	//  If one of the node is not belong to this group, just return error.

	servers := make([]string, 0, len(instances))
	for _, instance := range instances {
		servers = append(servers, *instance.InstanceId)
	}

	err = asg.cloudServiceManager.DeleteServers(servers)
	if err != nil {
		klog.Warningf("failed to delete nodes. error: %v", err)
		return err
	}

	return nil
}

// DecreaseTargetSize decreases the target size of the node group. This function
// doesn't permit to delete any existing node and can be used only to reduce the
// request for new nodes that have not been yet fulfilled. Delta should be negative.
// It is assumed that cloud provider will not delete the existing nodes when there
// is an option to just decrease the target. Implementation required.
func (asg *AutoScalingGroup) DecreaseTargetSize(delta int) error {
	// TODO(RainbowMango): Just remove nodes from group not delete them?
	return cloudprovider.ErrNotImplemented
}

// Id returns an unique identifier of the node group.
func (asg *AutoScalingGroup) Id() string {
	return asg.groupID
}

// Debug returns a string containing all information regarding this node group.
func (asg *AutoScalingGroup) Debug() string {
	return asg.String()
}

// Nodes returns a list of all nodes that belong to this node group.
// It is required that Instance objects returned by this method have Id field set.
// Other fields are optional.
// This list should include also instances that might have not become a kubernetes node yet.
func (asg *AutoScalingGroup) Nodes() ([]cloudprovider.Instance, error) {
	instances, err := asg.cloudServiceManager.GetInstances(asg.groupID)
	if err != nil {
		klog.Warningf("failed to get nodes from group: %s, error: %v", asg.groupID, err)
		return nil, err
	}
	if len(instances) == 0 {
		return nil, nil
	}

	// TODO(RainbowMango) Convert AS instances to cloud provider instances. Especially convert status.
	providerInstances := make([]cloudprovider.Instance, 0, len(instances))
	for i := range instances {
		pInstances := cloudprovider.Instance{
			Id:     *instances[i].InstanceId,
			Status: nil,
		}

		providerInstances = append(providerInstances, pInstances)
	}
	return nil, nil
}

// TemplateNodeInfo returns a schedulerframework.NodeInfo structure of an empty
// (as if just started) node. This will be used in scale-up simulations to
// predict what would a new node look like if a node group was expanded. The returned
// NodeInfo is expected to have a fully populated Node object, with all of the labels,
// capacity and allocatable information as well as all pods that are started on
// the node by default, using manifest (most likely only kube-proxy). Implementation optional.
func (asg *AutoScalingGroup) TemplateNodeInfo() (*schedulerframework.NodeInfo, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// Exist checks if the node group really exists on the cloud provider side. Allows to tell the
// theoretical node group from the real one. Implementation required.
func (asg *AutoScalingGroup) Exist() bool {
	// Since all group synced from remote and we do not support auto provision,
	// so we can assume that the group always exist.
	return true
}

// Create creates the node group on the cloud provider side. Implementation optional.
func (asg *AutoScalingGroup) Create() (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// Delete deletes the node group on the cloud provider side.
// This will be executed only for autoprovisioned node groups, once their size drops to 0.
// Implementation optional.
func (asg *AutoScalingGroup) Delete() error {
	return cloudprovider.ErrNotImplemented
}

// Autoprovisioned returns true if the node group is autoprovisioned. An autoprovisioned group
// was created by CA and can be deleted when scaled to 0.
//
// Always return false because the node group should maintained by user.
func (asg *AutoScalingGroup) Autoprovisioned() bool {
	return false
}

// String dumps current groups meta data.
func (asg *AutoScalingGroup) String() string {
	return fmt.Sprintf("group: %s min=%d max=%d", asg.groupID, asg.minInstanceNumber, asg.maxInstanceNumber)
}
