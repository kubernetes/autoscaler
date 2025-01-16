/*
Copyright 2019 The Kubernetes Authors.

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

package magnum

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	klog "k8s.io/klog/v2"
)

// How long to sleep after deleting nodes, to ensure that multiple requests arrive in order.
var postDeleteSleepDuration = 5 * time.Second

// magnumNodeGroup implements NodeGroup interface from cluster-autoscaler/cloudprovider.
//
// Represents a homogeneous collection of nodes within a cluster,
// which can be dynamically resized between a minimum and maximum
// number of nodes.
type magnumNodeGroup struct {
	magnumManager magnumManager

	// Human readable ID for logs and CA status configmap.
	id string
	// Node group UUID in Magnum.
	UUID string

	// To be locked when resizing the cluster, or reading
	// cluster state that could be being modified.
	// Shared between all node groups.
	clusterUpdateLock *sync.Mutex

	minSize    int
	maxSize    int
	targetSize int

	// deletedNodes tracks nodes which have been requested for deletion.
	// Heat can't always delete a node immediately if there is another concurrent update,
	// so reporting a node as being in a failed state multiple times can cause the autoscaler
	// to try to repeatedly delete it.
	// Maps provider ID -> time of deletion request.
	deletedNodes map[string]time.Time
	nodeTemplate *MagnumNodeTemplate
	// getOptions   *autoscaler.NodeGroupAutoscalingOptions
}

// MagnumNodeTemplate reference to implements TemplateNodeInfo
type MagnumNodeTemplate struct {
	CPUCores      int               `json:"cpu_cores,omitempty"`
	RAMMegabytes  int               `json:"ram_mb,omitempty"`
	DiskGigabytes int               `json:"disk_gb,omitempty"`
	Labels        map[string]string `json:"labels,omitempty"`
}

// IncreaseSize increases the number of nodes by replacing the cluster's node_count.
//
// Takes precautions so that the cluster is not modified while in an UPDATE_IN_PROGRESS state.
// Blocks until the cluster has reached UPDATE_COMPLETE.
func (ng *magnumNodeGroup) IncreaseSize(delta int) error {
	ng.clusterUpdateLock.Lock()
	defer ng.clusterUpdateLock.Unlock()

	if delta <= 0 {
		return fmt.Errorf("size increase must be positive")
	}

	size := ng.targetSize
	if size+delta > ng.MaxSize() {
		return fmt.Errorf("size increase too large, desired:%d max:%d", size+delta, ng.MaxSize())
	}

	klog.V(2).Infof("Increasing size by %d, %d->%d", delta, size, size+delta)
	err := ng.magnumManager.updateNodeCount(ng.UUID, size+delta)
	if err != nil {
		return fmt.Errorf("could not increase cluster size: %v", err)
	}

	ng.targetSize += delta
	return nil
}

// AtomicIncreaseSize is not implemented.
func (ng *magnumNodeGroup) AtomicIncreaseSize(delta int) error {
	return cloudprovider.ErrNotImplemented
}

// deleteNodes deletes a set of nodes chosen by the autoscaler.
func (ng *magnumNodeGroup) DeleteNodes(nodes []*apiv1.Node) error {
	ng.clusterUpdateLock.Lock()
	defer ng.clusterUpdateLock.Unlock()

	size := ng.targetSize

	var nodeNames []string
	for _, node := range nodes {
		nodeNames = append(nodeNames, node.Name)
	}
	klog.V(2).Infof("Deleting nodes: %v", nodeNames)

	// Check that the total number of nodes to be deleted will not take the node group below its minimum size
	if size-len(nodes) < ng.MinSize() {
		return fmt.Errorf("size decrease too large, desired:%d min:%d", size-len(nodes), ng.MinSize())
	}

	var nodeRefs []NodeRef
	for _, node := range nodes {
		nodeRefs = append(nodeRefs, NodeRef{
			Name:       node.Name,
			SystemUUID: node.Status.NodeInfo.SystemUUID,
			ProviderID: node.Spec.ProviderID,
			IsFake:     isFakeNode(node),
		})
	}

	err := ng.magnumManager.deleteNodes(ng.UUID, nodeRefs, size-len(nodes))
	if err != nil {
		return fmt.Errorf("manager error deleting nodes: %v", err)
	}

	ng.targetSize = size - len(nodes)

	now := time.Now()
	for _, node := range nodes {
		ng.deletedNodes[node.Spec.ProviderID] = now
	}

	// Sleep for a few seconds to ensure that delete requests are received in order.
	time.Sleep(postDeleteSleepDuration)

	return nil
}

// ForceDeleteNodes deletes nodes from the group regardless of constraints.
func (ng *magnumNodeGroup) ForceDeleteNodes(nodes []*apiv1.Node) error {
	return cloudprovider.ErrNotImplemented
}

// DecreaseTargetSize decreases the cluster node_count in magnum.
func (ng *magnumNodeGroup) DecreaseTargetSize(delta int) error {
	ng.clusterUpdateLock.Lock()
	defer ng.clusterUpdateLock.Unlock()

	if delta >= 0 {
		return fmt.Errorf("size decrease must be negative")
	}
	size := ng.targetSize
	if size+delta < ng.MinSize() {
		return fmt.Errorf("size decrease too large, desired:%d min:%d", size+delta, ng.MinSize())
	}

	klog.V(2).Infof("Decreasing target size by %d, %d->%d", delta, ng.targetSize, ng.targetSize+delta)
	err := ng.magnumManager.updateNodeCount(ng.UUID, ng.targetSize+delta)
	if err != nil {
		return fmt.Errorf("could not decrease target size: %v", err)
	}
	ng.targetSize += delta
	return nil
}

// Id returns the node group ID
func (ng *magnumNodeGroup) Id() string {
	return ng.id
}

// Debug returns a string formatted with the node group's min, max and target sizes.
func (ng *magnumNodeGroup) Debug() string {
	ng.clusterUpdateLock.Lock()
	defer ng.clusterUpdateLock.Unlock()
	return fmt.Sprintf("%s min=%d max=%d target=%d", ng.id, ng.minSize, ng.maxSize, ng.targetSize)
}

// Nodes returns a list of nodes that belong to this node group.
func (ng *magnumNodeGroup) Nodes() ([]cloudprovider.Instance, error) {
	ng.clusterUpdateLock.Lock()
	defer ng.clusterUpdateLock.Unlock()

	instances, err := ng.magnumManager.getNodes(ng.UUID)
	if err != nil {
		return nil, fmt.Errorf("could not get nodes: %v", err)
	}

	for node, deletedTime := range ng.deletedNodes {
		if time.Since(deletedTime) > 10*time.Minute {
			// Remove the node from the list of recently deleted nodes after 10 minutes.
			klog.V(3).Infof("Removing node %s from the deleted nodes cache after 10 minutes", node)
			delete(ng.deletedNodes, node)
		}
	}

	for i, instance := range instances {
		if instance.Status.State == cloudprovider.InstanceDeleting {
			continue
		}
		if deleteTime, ok := ng.deletedNodes[instance.Id]; ok {
			// This node has recently been requested for deletion, report the state as delete in progress.
			klog.V(3).Infof("Node %s has received deletetion request %s ago, reporting it as delete in progress instead of %v", instance.Id, time.Since(deleteTime), instance.Status.State)
			instances[i].Status.State = cloudprovider.InstanceDeleting
		}

	}
	return instances, nil
}

// TemplateNodeInfo returns a node template for this node group.
func (ng *magnumNodeGroup) TemplateNodeInfo() (*framework.NodeInfo, error) {
	node, err := ng.buildNodeFromTemplate(ng.Id(), ng.nodeTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to build node from template")
	}
	klog.V(5).Infof("TemplateNodeInfo: built template for nodegroup: %s", ng.Id())
	nodeInfo := framework.NewNodeInfo(node, nil, &framework.PodInfo{Pod: cloudprovider.BuildKubeProxy(ng.Id())})

	return nodeInfo, nil
}

// Exist returns if this node group exists.
// Currently always returns true.
func (ng *magnumNodeGroup) Exist() bool {
	return true
}

// Create creates the node group on the cloud provider side.
func (ng *magnumNodeGroup) Create() (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrAlreadyExist
}

// Delete deletes the node group on the cloud provider side.
func (ng *magnumNodeGroup) Delete() error {
	return cloudprovider.ErrNotImplemented
}

// Autoprovisioned returns if the nodegroup is autoprovisioned.
func (ng *magnumNodeGroup) Autoprovisioned() bool {
	return false
}

// GetOptions returns NodeGroupAutoscalingOptions that should be used for this particular
// NodeGroup. Returning a nil will result in using default options.
func (ng *magnumNodeGroup) GetOptions(defaults config.NodeGroupAutoscalingOptions) (*config.NodeGroupAutoscalingOptions, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// MaxSize returns the maximum allowed size of the node group.
func (ng *magnumNodeGroup) MaxSize() int {
	return ng.maxSize
}

// MinSize returns the minimum allowed size of the node group.
func (ng *magnumNodeGroup) MinSize() int {
	return ng.minSize
}

// TargetSize returns the target size of the node group.
func (ng *magnumNodeGroup) TargetSize() (int, error) {
	return ng.targetSize, nil
}

// buildNodeFromTemplate returns a Node object from the given template
func (ng *magnumNodeGroup) buildNodeFromTemplate(name string, template *MagnumNodeTemplate) (*apiv1.Node, error) {
	node := &apiv1.Node{}
	nodeName := fmt.Sprintf("%s-nodegroup-%d", name, rand.Int63())

	node.ObjectMeta = metav1.ObjectMeta{
		Name:     nodeName,
		SelfLink: fmt.Sprintf("/api/v1/nodes/%s", nodeName),
		Labels:   map[string]string{},
	}

	node.Status = apiv1.NodeStatus{
		Capacity: apiv1.ResourceList{},
	}
	node.Status.Capacity[apiv1.ResourcePods] = *resource.NewQuantity(110, resource.DecimalSI)
	node.Status.Capacity[apiv1.ResourceCPU] = *resource.NewQuantity(int64(template.CPUCores*1000), resource.DecimalSI)
	node.Status.Capacity[apiv1.ResourceMemory] = *resource.NewQuantity(int64(template.RAMMegabytes*1024*1024), resource.DecimalSI)
	node.Status.Capacity[apiv1.ResourceEphemeralStorage] = *resource.NewQuantity(int64(template.DiskGigabytes*1024*1024*1024), resource.DecimalSI)

	node.Status.Allocatable = node.Status.Capacity

	// GenericLabels and NodeLabels
	node.Labels = cloudprovider.JoinStringMaps(node.Labels, buildLabels(template, nodeName))

	node.Status.Conditions = cloudprovider.BuildReadyConditions()

	return node, nil
}

func buildLabels(template *MagnumNodeTemplate, nodeName string) map[string]string {
	result := make(map[string]string)

	// NodeLabels
	for key, value := range template.Labels {
		result[key] = value
	}

	return result
}
