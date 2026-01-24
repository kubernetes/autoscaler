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

package kamatera

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

// NodeGroup implements cloudprovider.NodeGroup interface. NodeGroup contains
// configuration info and functions to control a set of nodes that have the
// same capacity and set of labels.
type NodeGroup struct {
	id             string
	manager        *manager
	minSize        int
	maxSize        int
	instances      map[string]*Instance // key is the cloud provider ID
	serverConfig   ServerConfig
	templateLabels []string
}

var _ cloudprovider.NodeGroup = (*NodeGroup)(nil)

// MaxSize returns maximum size of the node group.
func (n *NodeGroup) MaxSize() int {
	return n.maxSize
}

// MinSize returns minimum size of the node group.
func (n *NodeGroup) MinSize() int {
	return n.minSize
}

// TargetSize returns the current target size of the node group. It is possible that the
// number of nodes in Kubernetes is different at the moment but should be equal
// to Size() once everything stabilizes (new nodes finish startup and registration or
// removed nodes are deleted completely). Implementation required.
func (n *NodeGroup) TargetSize() (int, error) {
	numInstances := 0
	for _, instance := range n.instances {
		if instance.Status != nil && instance.Status.State != cloudprovider.InstanceDeleting {
			numInstances += 1
		}
	}
	return numInstances, nil
}

// IncreaseSize increases the size of the node group. To delete a node you need
// to explicitly name it and use DeleteNode. This function should wait until
// node group size is updated. Implementation required.
func (n *NodeGroup) IncreaseSize(delta int) error {
	if delta <= 0 {
		return fmt.Errorf("delta must be positive, have: %d", delta)
	}
	currentSize, err := n.TargetSize()
	if err != nil {
		return err
	}
	targetSize := currentSize + delta
	klog.V(2).Infof("Increasing size of node group %s from %d to %d", n.id, currentSize, targetSize)
	if targetSize > n.MaxSize() {
		return fmt.Errorf("size increase is too large. current: %d desired: %d max: %d",
			currentSize, targetSize, n.MaxSize())
	}
	err = n.createInstances(delta)
	if err != nil {
		klog.Errorf("Failed to increase size of node group %s from %d to %d: %v", n.id, currentSize, targetSize, err)
		return err
	}
	return nil
}

// AtomicIncreaseSize is not implemented.
func (n *NodeGroup) AtomicIncreaseSize(delta int) error {
	return cloudprovider.ErrNotImplemented
}

// DeleteNodes deletes nodes from this node group. Error is returned either on
// failure or if the given node doesn't belong to this node group. This function
// should wait until node group size is updated. Implementation required.
func (n *NodeGroup) DeleteNodes(nodes []*apiv1.Node) error {
	klog.V(2).Infof("Deleting %d nodes from node group %s", len(nodes), n.id)
	for _, node := range nodes {
		instance, err := n.findInstanceForNode(node)
		if err != nil {
			return err
		}
		if instance == nil {
			return fmt.Errorf("Failed to delete node %q with provider ID %q: cannot find this node in the node group",
				node.Name, node.Spec.ProviderID)
		}
		err = instance.delete(n.manager.client, n.manager.config.providerIDPrefix, n.manager.config.PoweroffOnScaleDown)
		if err != nil {
			return fmt.Errorf("Failed to delete node %q with provider ID %q: %v",
				node.Name, node.Spec.ProviderID, err)
		}
	}
	return nil
}

// ForceDeleteNodes deletes nodes from the group regardless of constraints.
func (n *NodeGroup) ForceDeleteNodes(nodes []*apiv1.Node) error {
	return cloudprovider.ErrNotImplemented
}

// DecreaseTargetSize decreases the target size of the node group. This function
// doesn't permit to delete any existing node and can be used only to reduce the
// request for new nodes that have not been yet fulfilled. Delta should be negative.
// It is assumed that cloud provider will not delete the existing nodes when there
// is an option to just decrease the target. Implementation required.
func (n *NodeGroup) DecreaseTargetSize(delta int) error {
	// requests for new nodes are always fulfilled so we cannot
	// decrease the size without actually deleting nodes
	return cloudprovider.ErrNotImplemented
}

// Id returns an unique identifier of the node group.
func (n *NodeGroup) Id() string {
	return n.id
}

// Debug returns a string containing all information regarding this node group.
func (n *NodeGroup) Debug() string {
	return fmt.Sprintf("node group ID: %s (min:%d max:%d)", n.Id(), n.MinSize(), n.MaxSize())
}

// Nodes returns a list of all nodes that belong to this node group.
// It is required that Instance objects returned by this method have Id field set.
// Other fields are optional.
// This list should include also instances that might have not become a kubernetes node yet.
func (n *NodeGroup) Nodes() ([]cloudprovider.Instance, error) {
	var instances []cloudprovider.Instance
	for _, instance := range n.instances {
		if instance.Status != nil {
			instances = append(instances, cloudprovider.Instance{
				Id:     instance.Id,
				Status: instance.Status,
			})
		}
	}
	return instances, nil
}

// TemplateNodeInfo returns a framework.NodeInfo structure of an empty
// (as if just started) node. This will be used in scale-up simulations to
// predict what would a new node look like if a node group was expanded. The returned
// NodeInfo is expected to have a fully populated Node object, with all of the labels,
// capacity and allocatable information as well as all pods that are started on
// the node by default, using manifest (most likely only kube-proxy). Implementation optional.
func (n *NodeGroup) TemplateNodeInfo() (*framework.NodeInfo, error) {
	resourceList, err := n.getResourceList()
	if err != nil {
		return nil, fmt.Errorf("failed to create resource list for node group %s error: %v", n.id, err)
	}
	labels := make(map[string]string)
	for _, templateLabel := range n.templateLabels {
		parts := strings.SplitN(templateLabel, "=", 2)
		if len(parts) == 2 {
			labels[parts[0]] = parts[1]
		}
	}
	node := apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:   kamateraServerName(""),
			Labels: labels,
		},
		Status: apiv1.NodeStatus{
			Capacity:   resourceList,
			Conditions: cloudprovider.BuildReadyConditions(),
		},
	}
	node.Status.Allocatable = node.Status.Capacity
	node.Status.Conditions = cloudprovider.BuildReadyConditions()

	nodeInfo := framework.NewNodeInfo(&node, nil, &framework.PodInfo{Pod: cloudprovider.BuildKubeProxy(n.id)})
	return nodeInfo, nil
}

// Exist checks if the node group really exists on the cloud provider side. Allows to tell the
// theoretical node group from the real one. Implementation required.
func (n *NodeGroup) Exist() bool {
	return true
}

// Create creates the node group on the cloud provider side. Implementation optional.
func (n *NodeGroup) Create() (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// Delete deletes the node group on the cloud provider side.
// This will be executed only for autoprovisioned node groups, once their size drops to 0.
// Implementation optional.
func (n *NodeGroup) Delete() error {
	return cloudprovider.ErrNotImplemented
}

// Autoprovisioned returns true if the node group is autoprovisioned. An autoprovisioned group
// was created by CA and can be deleted when scaled to 0.
func (n *NodeGroup) Autoprovisioned() bool {
	return false
}

// GetOptions returns NodeGroupAutoscalingOptions that should be used for this particular
// NodeGroup. Returning a nil will result in using default options.
// Implementation optional.
func (n *NodeGroup) GetOptions(defaults config.NodeGroupAutoscalingOptions) (*config.NodeGroupAutoscalingOptions, error) {
	return &config.NodeGroupAutoscalingOptions{
		ScaleDownUtilizationThreshold:    defaults.ScaleDownUtilizationThreshold,
		ScaleDownGpuUtilizationThreshold: defaults.ScaleDownGpuUtilizationThreshold,
		ScaleDownUnneededTime:            defaults.ScaleDownUnneededTime,
		ScaleDownUnreadyTime:             defaults.ScaleDownUnreadyTime,
		MaxNodeProvisionTime:             time.Hour, // we can't cancel creation in progress so must give it enough time to complete
		ZeroOrMaxNodeScaling:             defaults.ZeroOrMaxNodeScaling,
		IgnoreDaemonSetsUtilization:      defaults.IgnoreDaemonSetsUtilization,
	}, nil
}

func (n *NodeGroup) findInstanceForNode(node *apiv1.Node) (*Instance, error) {
	var instance *Instance
	logPrefix := fmt.Sprintf("ng %s findInstanceForNode(%s)", n.id, node.Name)
	for _, i := range n.instances {
		if i.Id == node.Spec.ProviderID {
			instance = i
		} else if i.Id == formatKamateraProviderID(n.manager.config.providerIDPrefix, node.Name) {
			klog.V(2).Infof("%s: found based on node Name, setting ProviderID", logPrefix)
			node.Spec.ProviderID = formatKamateraProviderID(n.manager.config.providerIDPrefix, node.Name)
			err := setNodeProviderID(n.manager.kubeClient, node.Name, node.Spec.ProviderID)
			if err != nil {
				// this is not a critical error, the autoscaler can continue functioning in this condition
				// as the same node object is used in later code the ProviderID change will be picked up
				klog.Warningf("%s: failed to set node ProviderID: %v", logPrefix, err)
			}
			instance = i
		}
	}
	if instance == nil || instance.Status == nil {
		return nil, nil
	}
	return instance, nil
}

func (n *NodeGroup) createInstances(count int) error {
	for _, instance := range n.instances {
		if instance.Status != nil && instance.Status.State == cloudprovider.InstanceCreating {
			klog.V(4).Infof("createInstances: instance %s is still creating", instance.Id)
			count--
		}
	}
	if count <= 0 {
		klog.V(4).Infof("createInstances: skipping because instance count (%d) is 0 or less due to still creaing instances", count)
		return fmt.Errorf("instances are still being created")
	}
	if n.manager.config.PoweronOnScaleUp {
		var poweronCandidateInstances []*Instance
		for _, instance := range n.manager.snapshotInstances() {
			if sets.New(n.serverConfig.Tags...).Equal(sets.New(instance.Tags...)) && instance.PowerOn == false && instance.Status == nil {
				poweronCandidateInstances = append(poweronCandidateInstances, instance)
			}
		}
		klog.V(2).Infof("createInstances found %d powered off instances matching node group %s", len(poweronCandidateInstances), n.id)
		icount := count
		for i := 0; i < icount && i < len(poweronCandidateInstances); i++ {
			instance := poweronCandidateInstances[i]
			klog.V(4).Infof("createInstances: creating instance %s", instance.Id)
			err := instance.createPoweron(n.manager.client, n.manager.config.providerIDPrefix)
			if err == nil {
				n.instances[instance.Id] = instance
				count--
			}
		}
	}
	if count > 0 {
		serverCommandIds, err := n.manager.client.StartCreateServers(context.Background(), count, n.serverConfig)
		if err != nil {
			return err
		}
		for serverName, commandId := range serverCommandIds {
			instance := n.manager.addCreatingInstance(serverName, commandId, n.serverConfig.Tags)
			n.instances[instance.Id] = instance
			klog.V(4).Infof("%v", n.extendedDebug())
		}
	}
	return nil
}

func (n *NodeGroup) extendedDebug() string {
	// TODO: provide extended debug information regarding this node group
	msgs := []string{n.Debug()}
	for _, instance := range n.instances {
		msgs = append(msgs, instance.extendedDebug())
	}
	return strings.Join(msgs, "\n")
}

func (n *NodeGroup) getResourceList() (apiv1.ResourceList, error) {
	ramMb, err := strconv.Atoi(n.serverConfig.Ram)
	if err != nil {
		return nil, fmt.Errorf("failed to parse server config ram %s: %v", n.serverConfig.Ram, err)
	}
	// TODO: handle CPU types
	if len(n.serverConfig.Cpu) < 2 {
		return nil, fmt.Errorf("failed to parse server config cpu %s", n.serverConfig.Cpu)
	}
	cpuCores, err := strconv.Atoi(n.serverConfig.Cpu[0 : len(n.serverConfig.Cpu)-1])
	if err != nil {
		return nil, fmt.Errorf("failed to parse server config cpu %s: %v", n.serverConfig.Cpu, err)
	}
	// TODO: handle additional disks
	firstDiskSizeGb := 0
	if len(n.serverConfig.Disks) > 0 {
		firstDiskSpec := n.serverConfig.Disks[0]
		for _, attr := range strings.Split(firstDiskSpec, ",") {
			if strings.HasPrefix(attr, "size=") {
				firstDiskSizeGb, err = strconv.Atoi(strings.Split(attr, "=")[1])
				if err != nil {
					return nil, err
				}
			}
		}
	}
	return apiv1.ResourceList{
		// TODO somehow determine the actual pods that will be running
		apiv1.ResourcePods:    *resource.NewQuantity(110, resource.DecimalSI),
		apiv1.ResourceCPU:     *resource.NewQuantity(int64(cpuCores), resource.DecimalSI),
		apiv1.ResourceMemory:  *resource.NewQuantity(int64(ramMb*1024*1024), resource.DecimalSI),
		apiv1.ResourceStorage: *resource.NewQuantity(int64(firstDiskSizeGb*1024*1024*1024), resource.DecimalSI),
	}, nil
}

func setNodeProviderID(kubeClient kubernetes.Interface, nodeName string, value string) error {
	node, err := kubeClient.CoreV1().Nodes().Get(context.Background(), nodeName, metav1.GetOptions{})
	if err != nil {
		klog.Errorf("failed to get node to update provider ID %s %+v", nodeName, err)
		return err
	}
	node.Spec.ProviderID = value
	_, err = kubeClient.CoreV1().Nodes().Update(context.Background(), node, metav1.UpdateOptions{})
	if err != nil {
		klog.Errorf("failed to update node's provider ID %s %+v", nodeName, err)
		return err
	}
	klog.V(2).Infof("updated provider ID on node: %s", nodeName)
	return nil
}
