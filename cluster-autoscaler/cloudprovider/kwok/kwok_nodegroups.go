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

package kwok

import (
	"context"
	"fmt"

	apiv1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	klog "k8s.io/klog/v2"
)

var (
	sizeIncreaseMustBePositiveErr   = "size increase must be positive"
	maxSizeReachedErr               = "size increase too large"
	minSizeReachedErr               = "min size reached, nodes will not be deleted"
	belowMinSizeErr                 = "can't delete nodes because nodegroup size would go below min size"
	notManagedByKwokErr             = "can't delete node '%v' because it is not managed by kwok"
	sizeDecreaseMustBeNegativeErr   = "size decrease must be negative"
	attemptToDeleteExistingNodesErr = "attempt to delete existing nodes"
)

// MaxSize returns maximum size of the node group.
func (nodeGroup *NodeGroup) MaxSize() int {
	return nodeGroup.maxSize
}

// MinSize returns minimum size of the node group.
func (nodeGroup *NodeGroup) MinSize() int {
	return nodeGroup.minSize
}

// TargetSize returns the current TARGET size of the node group. It is possible that the
// number is different from the number of nodes registered in Kubernetes.
func (nodeGroup *NodeGroup) TargetSize() (int, error) {
	return nodeGroup.targetSize, nil
}

// IncreaseSize increases NodeGroup size.
func (nodeGroup *NodeGroup) IncreaseSize(delta int) error {
	if delta <= 0 {
		return fmt.Errorf(sizeIncreaseMustBePositiveErr)
	}
	size := nodeGroup.targetSize
	newSize := int(size) + delta
	if newSize > nodeGroup.MaxSize() {
		return fmt.Errorf("%s, desired: %d max: %d", maxSizeReachedErr, newSize, nodeGroup.MaxSize())
	}

	klog.V(5).Infof("increasing size of nodegroup '%s' to %v (old size: %v, delta: %v)", nodeGroup.name, newSize, size, delta)

	schedNode, err := nodeGroup.TemplateNodeInfo()
	if err != nil {
		return fmt.Errorf("couldn't create a template node for nodegroup %s", nodeGroup.name)
	}

	for i := 0; i < delta; i++ {
		node := schedNode.Node()
		node.Name = fmt.Sprintf("%s-%s", nodeGroup.name, rand.String(5))
		node.Spec.ProviderID = getProviderID(node.Name)
		_, err := nodeGroup.kubeClient.CoreV1().Nodes().Create(context.Background(), node, v1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("couldn't create new node '%s': %v", node.Name, err)
		}
		nodeGroup.targetSize += 1
	}

	return nil
}

// AtomicIncreaseSize is not implemented.
func (nodeGroup *NodeGroup) AtomicIncreaseSize(delta int) error {
	return cloudprovider.ErrNotImplemented
}

// DeleteNodes deletes the specified nodes from the node group.
func (nodeGroup *NodeGroup) DeleteNodes(nodes []*apiv1.Node) error {
	size := nodeGroup.targetSize
	if size <= nodeGroup.MinSize() {
		return fmt.Errorf(minSizeReachedErr)
	}

	if size-len(nodes) < nodeGroup.MinSize() {
		return fmt.Errorf(belowMinSizeErr)
	}

	for _, node := range nodes {
		// TODO(vadasambar): check if there's a better way than returning an error here
		if node.GetAnnotations()[KwokManagedAnnotation] != "fake" {
			return fmt.Errorf(notManagedByKwokErr, node.GetName())
		}

		// TODO(vadasambar): proceed to delete the next node if the current node deletion errors
		// TODO(vadasambar): collect all the errors and return them after attempting to delete all the nodes to be deleted
		err := nodeGroup.kubeClient.CoreV1().Nodes().Delete(context.Background(), node.GetName(), v1.DeleteOptions{})
		if err != nil {
			return err
		}
		nodeGroup.targetSize -= 1
	}
	return nil
}

// DecreaseTargetSize decreases the target size of the node group. This function
// doesn't permit to delete any existing node and can be used only to reduce the
// request for new nodes that have not been yet fulfilled. Delta should be negative.
func (nodeGroup *NodeGroup) DecreaseTargetSize(delta int) error {
	if delta >= 0 {
		return fmt.Errorf(sizeDecreaseMustBeNegativeErr)
	}
	size := nodeGroup.targetSize
	nodes, err := nodeGroup.getNodeNamesForNodeGroup()
	if err != nil {
		return err
	}
	newSize := int(size) + delta
	if newSize < len(nodes) {
		return fmt.Errorf("%s, targetSize: %d delta: %d existingNodes: %d",
			attemptToDeleteExistingNodesErr, size, delta, len(nodes))
	}

	nodeGroup.targetSize = newSize

	return nil
}

// getNodeNamesForNodeGroup returns list of nodes belonging to the nodegroup
func (nodeGroup *NodeGroup) getNodeNamesForNodeGroup() ([]string, error) {
	names := []string{}

	nodeList, err := nodeGroup.lister.List()
	if err != nil {
		return names, err
	}

	for _, no := range nodeList {
		names = append(names, no.GetName())
	}

	return names, nil
}

// Id returns nodegroup name.
func (nodeGroup *NodeGroup) Id() string {
	return nodeGroup.name
}

// Debug returns a debug string for the nodegroup.
func (nodeGroup *NodeGroup) Debug() string {
	return fmt.Sprintf("%s (%d:%d)", nodeGroup.Id(), nodeGroup.MinSize(), nodeGroup.MaxSize())
}

// Nodes returns a list of all nodes that belong to this node group.
func (nodeGroup *NodeGroup) Nodes() ([]cloudprovider.Instance, error) {
	instances := make([]cloudprovider.Instance, 0)
	nodeNames, err := nodeGroup.getNodeNamesForNodeGroup()
	if err != nil {
		return instances, err
	}
	for _, nodeName := range nodeNames {
		instances = append(instances, cloudprovider.Instance{Id: getProviderID(nodeName), Status: &cloudprovider.InstanceStatus{
			State:     cloudprovider.InstanceRunning,
			ErrorInfo: nil,
		}})
	}
	return instances, nil
}

// TemplateNodeInfo returns a node template for this node group.
func (nodeGroup *NodeGroup) TemplateNodeInfo() (*framework.NodeInfo, error) {
	nodeInfo := framework.NewNodeInfo(nodeGroup.nodeTemplate, nil, &framework.PodInfo{Pod: cloudprovider.BuildKubeProxy(nodeGroup.Id())})
	return nodeInfo, nil
}

// Exist checks if the node group really exists on the cloud provider side.
// Since kwok nodegroup is not backed by anything on cloud provider side
// We can safely return `true` here
func (nodeGroup *NodeGroup) Exist() bool {
	return true
}

// Create creates the node group on the cloud provider side.
// Left unimplemented because Create is not used anywhere
// in the core autoscaler as of writing this
func (nodeGroup *NodeGroup) Create() (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// Delete deletes the node group on the cloud provider side.
// Left unimplemented because Delete is not used anywhere
// in the core autoscaler as of writing this
func (nodeGroup *NodeGroup) Delete() error {
	return cloudprovider.ErrNotImplemented
}

// Autoprovisioned returns true if the node group is autoprovisioned.
func (nodeGroup *NodeGroup) Autoprovisioned() bool {
	return false
}

// GetOptions returns NodeGroupAutoscalingOptions that should be used for this particular
// NodeGroup. Returning a nil will result in using default options.
func (nodeGroup *NodeGroup) GetOptions(defaults config.NodeGroupAutoscalingOptions) (*config.NodeGroupAutoscalingOptions, error) {
	return &defaults, nil
}
