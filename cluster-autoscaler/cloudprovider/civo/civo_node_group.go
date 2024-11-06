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

package civo

import (
	"errors"
	"fmt"
	"math/rand"

	"k8s.io/apimachinery/pkg/api/resource"
	civocloud "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/civo/civo-cloud-sdk-go"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	autoscaler "k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	"k8s.io/klog/v2"
)

// NodeGroup implements cloudprovider.NodeGroup interface. NodeGroup contains
// configuration info and functions to control a set of nodes that have the
// same capacity and set of labels.
type NodeGroup struct {
	id           string
	clusterID    string
	client       nodeGroupClient
	nodePool     *civocloud.KubernetesPool
	minSize      int
	maxSize      int
	getOptions   *autoscaler.NodeGroupAutoscalingOptions
	nodeTemplate *CivoNodeTemplate
}

// CivoNodeTemplate reference to implements TemplateNodeInfo
type CivoNodeTemplate struct {
	// Size represents the pool size of civocloud
	Size          string            `json:"size,omitempty"`
	CPUCores      int               `json:"cpu_cores,omitempty"`
	RAMMegabytes  int               `json:"ram_mb,omitempty"`
	DiskGigabytes int               `json:"disk_gb,omitempty"`
	Labels        map[string]string `json:"labels,omitempty"`
	Taints        []apiv1.Taint     `json:"taint,omitempty"`
	GpuCount      int               `json:"gpu_count,omitempty"`
	Region        string            `json:"region,omitempty"`
}

// MaxSize returns maximum size of the node group.
func (n *NodeGroup) MaxSize() int {
	return n.maxSize
}

// MinSize returns minimum size of the node group.
func (n *NodeGroup) MinSize() int {
	return n.minSize
}

// GetOptions returns the options used to create this node group.
func (n *NodeGroup) GetOptions(autoscaler.NodeGroupAutoscalingOptions) (*autoscaler.NodeGroupAutoscalingOptions, error) {
	return n.getOptions, nil
}

// TargetSize returns the current target size of the node group. It is possible
// that the number of nodes in Kubernetes is different at the moment but should
// be equal to Size() once everything stabilizes (new nodes finish startup and
// registration or removed nodes are deleted completely). Implementation
// required.
func (n *NodeGroup) TargetSize() (int, error) {
	return n.nodePool.Count, nil
}

// IncreaseSize increases the size of the node group. To delete a node you need
// to explicitly name it and use DeleteNode. This function should wait until
// node group size is updated. Implementation required.
func (n *NodeGroup) IncreaseSize(delta int) error {
	if delta <= 0 {
		return fmt.Errorf("delta must be positive, have: %d", delta)
	}

	targetSize := n.nodePool.Count + delta

	if targetSize > n.MaxSize() {
		return fmt.Errorf("size increase is too large. current: %d desired: %d max: %d",
			n.nodePool.Count, targetSize, n.MaxSize())
	}

	req := &civocloud.KubernetesClusterPoolUpdateConfig{
		Count:  targetSize,
		Region: Region,
	}
	updatedNodePool, err := n.client.UpdateKubernetesClusterPool(n.clusterID, n.id, req)
	if err != nil {
		return err
	}

	if targetSize > n.MaxSize() {
		return fmt.Errorf("size increase too large. current: %d, desired: %d, max: %d",
			updatedNodePool.Count, targetSize, n.MaxSize())
	}

	// update internal cache
	n.nodePool.Count = targetSize
	return nil
}

// AtomicIncreaseSize is not implemented.
func (n *NodeGroup) AtomicIncreaseSize(delta int) error {
	return cloudprovider.ErrNotImplemented
}

// DeleteNodes deletes nodes from this node group (and also increasing the size
// of the node group with that). Error is returned either on failure or if the
// given node doesn't belong to this node group. This function should wait
// until node group size is updated. Implementation required.
func (n *NodeGroup) DeleteNodes(nodes []*apiv1.Node) error {
	for _, node := range nodes {
		instanceID := toNodeID(node.Spec.ProviderID)
		klog.V(4).Infof("deleteing node: %q", instanceID)
		_, err := n.client.DeleteKubernetesClusterPoolInstance(n.clusterID, n.id, instanceID)
		if err != nil {
			return fmt.Errorf("deleting node failed for cluster: %q node pool: %q node: %q: %s",
				n.clusterID, n.id, node.Name, err)
		}

		// decrement the count by one  after a successful delete
		n.nodePool.Count--
	}

	return nil
}

// DecreaseTargetSize decreases the target size of the node group. This function
// doesn't permit to delete any existing node and can be used only to reduce the
// request for new nodes that have not been yet fulfilled. Delta should be negative.
// It is assumed that cloud provider will not delete the existing nodes when there
// is an option to just decrease the target. Implementation required.
func (n *NodeGroup) DecreaseTargetSize(delta int) error {
	if delta >= 0 {
		return fmt.Errorf("delta must be negative, have: %d", delta)
	}

	targetSize := n.nodePool.Count + delta
	if targetSize < n.MinSize() {
		return fmt.Errorf("size decrease is too small. current: %d desired: %d min: %d",
			n.nodePool.Count, targetSize, n.MinSize())
	}

	req := &civocloud.KubernetesClusterPoolUpdateConfig{
		Count:  targetSize,
		Region: Region,
	}

	updatedNodePool, err := n.client.UpdateKubernetesClusterPool(n.clusterID, n.id, req)
	if err != nil {
		return err
	}

	if updatedNodePool.Count != targetSize {
		return fmt.Errorf("couldn't increase size to %d (delta: %d). Current size is: %d",
			targetSize, delta, updatedNodePool.Count)
	}

	// update internal cache
	n.nodePool.Count = targetSize
	return nil
}

// Id returns an unique identifier of the node group.
func (n *NodeGroup) Id() string {
	return n.id
}

// Debug returns a string containing all information regarding this node group.
func (n *NodeGroup) Debug() string {
	return fmt.Sprintf("id: %s (min:%d max:%d)", n.Id(), n.MinSize(), n.MaxSize())
}

// Nodes returns a list of all nodes that belong to this node group.  It is
// required that Instance objects returned by this method have Id field set.
// Other fields are optional.
func (n *NodeGroup) Nodes() ([]cloudprovider.Instance, error) {
	if n.nodePool == nil {
		return nil, errors.New("node pool instance is not created")
	}

	return toInstances(n.nodePool.Instances), nil
}

// TemplateNodeInfo returns a schedulernodeinfo.NodeInfo structure of an empty
// (as if just started) node. This will be used in scale-up simulations to
// predict what would a new node look like if a node group was expanded. The
// returned NodeInfo is expected to have a fully populated Node object, with
// all of the labels, capacity and allocatable information as well as all pods
// that are started on the node by default, using manifest (most likely only
// kube-proxy). Implementation optional.
func (n *NodeGroup) TemplateNodeInfo() (*framework.NodeInfo, error) {
	node, err := n.buildNodeFromTemplate(n.Id(), n.nodeTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to build node from template")
	}

	nodeInfo := framework.NewNodeInfo(node, nil, &framework.PodInfo{Pod: cloudprovider.BuildKubeProxy(n.Id())})
	return nodeInfo, nil
}

// Exist checks if the node group really exists on the cloud provider side.
// Allows to tell the theoretical node group from the real one. Implementation
// required.
func (n *NodeGroup) Exist() bool {
	return n.nodePool != nil
}

// Create creates the node group on the cloud provider side. Implementation
// optional.
func (n *NodeGroup) Create() (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// Delete deletes the node group on the cloud provider side.  This will be
// executed only for autoprovisioned node groups, once their size drops to 0.
// Implementation optional.
func (n *NodeGroup) Delete() error {
	return cloudprovider.ErrNotImplemented
}

// Autoprovisioned returns true if the node group is autoprovisioned. An
// autoprovisioned group was created by CA and can be deleted when scaled to 0.
func (n *NodeGroup) Autoprovisioned() bool {
	return false
}

// toInstances converts a slice of civogo.KubernetesInstance to
// cloudprovider.Instance
func toInstances(nodes []civocloud.KubernetesInstance) []cloudprovider.Instance {
	instances := make([]cloudprovider.Instance, 0, len(nodes))
	for _, nd := range nodes {
		instances = append(instances, toInstance(nd))
	}
	return instances
}

// toInstance converts the given civogo.KubernetesInstance to a
// cloudprovider.Instance
func toInstance(node civocloud.KubernetesInstance) cloudprovider.Instance {
	return cloudprovider.Instance{
		Id:     toProviderID(node.ID),
		Status: toInstanceStatus(node.Status),
	}
}

// toInstanceStatus converts the given civo instance status to a
// cloudprovider.InstanceStatus
func toInstanceStatus(nodeState string) *cloudprovider.InstanceStatus {
	if nodeState == "" {
		return nil
	}

	st := &cloudprovider.InstanceStatus{}
	switch nodeState {
	case "BUILDING":
		st.State = cloudprovider.InstanceCreating
	case "ACTIVE":
		st.State = cloudprovider.InstanceRunning
	case "DELETING":
		st.State = cloudprovider.InstanceDeleting
	default:
		st.ErrorInfo = &cloudprovider.InstanceErrorInfo{
			ErrorClass:   cloudprovider.OtherErrorClass,
			ErrorCode:    "no-code-civo",
			ErrorMessage: nodeState,
		}
	}

	return st
}

// buildNodeFromTemplate returns a Node object from the given template
func (n *NodeGroup) buildNodeFromTemplate(name string, template *CivoNodeTemplate) (*apiv1.Node, error) {
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
	node.Status.Capacity[gpu.ResourceNvidiaGPU] = *resource.NewQuantity(int64(template.GpuCount), resource.DecimalSI)

	node.Status.Allocatable = node.Status.Capacity

	// GenericLabels and NodeLabels
	node.Labels = cloudprovider.JoinStringMaps(node.Labels, buildLabels(template, nodeName))

	_, ok := node.Labels["kubernetes.civo.com/civo-node-pool"]
	if !ok {
		node.Labels["kubernetes.civo.com/civo-node-pool"] = n.Id()
	}

	node.Spec.Taints = template.Taints

	node.Status.Conditions = cloudprovider.BuildReadyConditions()

	return node, nil
}

func buildLabels(template *CivoNodeTemplate, nodeName string) map[string]string {
	result := make(map[string]string)

	// NodeLabels
	for key, value := range template.Labels {
		result[key] = value
	}

	// GenericLabels
	result[apiv1.LabelOSStable] = cloudprovider.DefaultOS
	result[apiv1.LabelInstanceTypeStable] = template.Size
	result[apiv1.LabelTopologyRegion] = template.Region
	result[apiv1.LabelHostname] = nodeName

	return result
}
