/*
Copyright 2023 The Kubernetes Authors.

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

package volcengine

import (
	"fmt"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/klog/v2"
)

// AutoScalingGroup represents a Volcengine 'Auto Scaling Group' which also can be treated as a node group.
type AutoScalingGroup struct {
	manager           VolcengineManager
	asgId             string
	minInstanceNumber int
	maxInstanceNumber int
}

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
func (asg *AutoScalingGroup) TargetSize() (int, error) {
	return asg.manager.GetAsgDesireCapacity(asg.asgId)
}

// IncreaseSize increases the size of the node group. To delete a node you need
// to explicitly name it and use DeleteNode. This function should wait until
// node group size is updated. Implementation required.
func (asg *AutoScalingGroup) IncreaseSize(delta int) error {
	if delta <= 0 {
		return fmt.Errorf("size increase must be positive")
	}
	size, err := asg.manager.GetAsgDesireCapacity(asg.asgId)
	if err != nil {
		return err
	}
	if size+delta > asg.MaxSize() {
		return fmt.Errorf("size increase is too large - desired:%d max:%d", size+delta, asg.MaxSize())
	}
	return asg.manager.SetAsgTargetSize(asg.asgId, size+delta)
}

// AtomicIncreaseSize is not implemented.
func (asg *AutoScalingGroup) AtomicIncreaseSize(delta int) error {
	return cloudprovider.ErrNotImplemented
}

// DeleteNodes deletes nodes from this node group. Error is returned either on
// failure or if the given node doesn't belong to this node group. This function
// should wait until node group size is updated. Implementation required.
func (asg *AutoScalingGroup) DeleteNodes(nodes []*apiv1.Node) error {
	size, err := asg.manager.GetAsgDesireCapacity(asg.asgId)
	if err != nil {
		klog.Errorf("Failed to get desire capacity for %s: %v", asg.asgId, err)
		return err
	}
	if size <= asg.MinSize() {
		klog.Errorf("Failed to delete nodes from %s: min size reached", asg.asgId)
		return fmt.Errorf("asg min size reached")
	}

	instanceIds := make([]string, 0, len(nodes))
	for _, node := range nodes {
		belongs, err := asg.belongs(node)
		if err != nil {
			return err
		}
		if !belongs {
			return fmt.Errorf("node %s doesn't belong to asg %s", node.Name, asg.asgId)
		}
		instanceId, err := ecsInstanceFromProviderId(node.Spec.ProviderID)
		if err != nil {
			return err
		}
		instanceIds = append(instanceIds, instanceId)
	}
	return asg.manager.DeleteScalingInstances(asg.asgId, instanceIds)
}

func (asg *AutoScalingGroup) belongs(node *apiv1.Node) (bool, error) {
	instanceId, err := ecsInstanceFromProviderId(node.Spec.ProviderID)
	if err != nil {
		return false, err
	}
	targetAsg, err := asg.manager.GetAsgForInstance(instanceId)
	if err != nil {
		return false, err
	}
	if targetAsg == nil {
		return false, nil
	}
	return targetAsg.Id() == asg.asgId, nil
}

// DecreaseTargetSize decreases the target size of the node group. This function
// doesn't permit to delete any existing node and can be used only to reduce the
// request for new nodes that have not been yet fulfilled. Delta should be negative.
// It is assumed that cloud provider will not delete the existing nodes when there
// is an option to just decrease the target. Implementation required.
func (asg *AutoScalingGroup) DecreaseTargetSize(delta int) error {
	if delta >= 0 {
		return fmt.Errorf("size decrease size must be negative")
	}
	desireCapacity, err := asg.manager.GetAsgDesireCapacity(asg.asgId)
	if err != nil {
		klog.Errorf("Failed to get desire capacity for %s: %v", asg.asgId, err)
		return err
	}
	allNodes, err := asg.manager.GetAsgNodes(asg.asgId)
	if err != nil {
		klog.Errorf("Failed to get nodes for %s: %v", asg.asgId, err)
		return err
	}
	if desireCapacity+delta < len(allNodes) {
		return fmt.Errorf("size decrease is too large, need to delete existing node - newDesiredCapacity:%d currentNodes:%d", desireCapacity+delta, len(allNodes))
	}

	return asg.manager.SetAsgDesireCapacity(asg.asgId, desireCapacity+delta)
}

// Id returns an unique identifier of the node group.
func (asg *AutoScalingGroup) Id() string {
	return asg.asgId
}

// Debug returns a string containing all information regarding this node group.
func (asg *AutoScalingGroup) Debug() string {
	return fmt.Sprintf("%s (%d:%d)", asg.Id(), asg.MinSize(), asg.MaxSize())
}

// Nodes returns a list of all nodes that belong to this node group.
// It is required that Instance objects returned by this method have Id field set.
// Other fields are optional.
// This list should include also instances that might have not become a kubernetes node yet.
func (asg *AutoScalingGroup) Nodes() ([]cloudprovider.Instance, error) {
	nodes, err := asg.manager.GetAsgNodes(asg.asgId)
	if err != nil {
		return nil, err
	}
	return nodes, nil
}

// TemplateNodeInfo returns a framework.NodeInfo structure of an empty
// (as if just started) node. This will be used in scale-up simulations to
// predict what would a new node look like if a node group was expanded. The returned
// NodeInfo is expected to have a fully populated Node object, with all of the labels,
// capacity and allocatable information as well as all pods that are started on
// the node by default, using manifest (most likely only kube-proxy). Implementation optional.
func (asg *AutoScalingGroup) TemplateNodeInfo() (*framework.NodeInfo, error) {
	template, err := asg.manager.getAsgTemplate(asg.asgId)
	if err != nil {
		return nil, err
	}
	node, err := asg.manager.buildNodeFromTemplateName(asg.asgId, template)
	if err != nil {
		return nil, err
	}
	nodeInfo := framework.NewNodeInfo(node, nil, &framework.PodInfo{Pod: cloudprovider.BuildKubeProxy(asg.asgId)})
	return nodeInfo, nil
}

// Exist checks if the node group really exists on the cloud provider side. Allows to tell the
// theoretical node group from the real one. Implementation required.
func (asg *AutoScalingGroup) Exist() bool {
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
func (asg *AutoScalingGroup) Autoprovisioned() bool {
	return false
}

// GetOptions returns NodeGroupAutoscalingOptions that should be used for this particular
// NodeGroup. Returning a nil will result in using default options.
// Implementation optional.
func (asg *AutoScalingGroup) GetOptions(defaults config.NodeGroupAutoscalingOptions) (*config.NodeGroupAutoscalingOptions, error) {
	return nil, cloudprovider.ErrNotImplemented
}
