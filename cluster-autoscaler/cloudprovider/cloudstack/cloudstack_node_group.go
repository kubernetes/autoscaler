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

package cloudstack

import (
	"encoding/json"
	"fmt"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/cloudstack/service"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	klog "k8s.io/klog/v2"
)

// asg implements NodeGroup interface.
type asg struct {
	cluster *service.Cluster
	manager *manager
}

// MaxSize returns maximum size of the node group.
func (asg *asg) MaxSize() int {
	return asg.cluster.Maxsize
}

// MinSize returns minimum size of the node group.
func (asg *asg) MinSize() int {
	return asg.cluster.Minsize
}

// TargetSize returns the current TARGET size of the node group. It is possible that the
// number is different from the number of nodes registered in Kubernetes.
func (asg *asg) TargetSize() (int, error) {
	return asg.cluster.WorkerCount, nil
}

// IncreaseSize increases cluster size
func (asg *asg) IncreaseSize(delta int) error {
	klog.Infof("Increase Cluster : %s by %d", asg.cluster.ID, delta)
	if delta <= 0 {
		return fmt.Errorf("Delta must be positive")
	}
	newSize := asg.cluster.WorkerCount + delta
	if newSize > asg.MaxSize() {
		return fmt.Errorf("Delta too large - Wanted : %d Max : %d Have : %d", newSize, asg.MaxSize(), asg.cluster.WorkerCount)
	}

	cluster, err := asg.manager.scaleCluster(asg.cluster.ID, asg.cluster.WorkerCount+delta)
	if err != nil {
		return err
	}
	asg.Copy(cluster)
	return nil
}

// AtomicIncreaseSize is not implemented.
func (asg *asg) AtomicIncreaseSize(delta int) error {
	return cloudprovider.ErrNotImplemented
}

// DecreaseTargetSize decreases the target size of the node group. This function
// doesn't permit to delete any existing node and can be used only to reduce the
// request for new nodes that have not been yet fulfilled. Delta should be negative.
// It is assumed that cloud provider will not delete the existing nodes if the size
// when there is an option to just decrease the target.
func (asg *asg) DecreaseTargetSize(delta int) error {
	return errors.NewAutoscalerError(errors.CloudProviderError, "CloudProvider does not support DecreaseTargetSize")
}

// Belongs returns true if the given node belongs to the NodeGroup.
func (asg *asg) Belongs(node *apiv1.Node) (bool, error) {
	for _, vm := range asg.cluster.VirtualMachines {
		if vm.Name != "" && node.Name != "" && vm.Name == node.Name {
			return true, nil
		}
		if vm.ID == node.Status.NodeInfo.SystemUUID {
			return true, nil
		}
	}
	return false, fmt.Errorf("Unable to find node %s in cluster", node.Name)
}

// DeleteNodes deletes the nodes from the group.
func (asg *asg) DeleteNodes(nodes []*apiv1.Node) error {
	if asg.cluster.WorkerCount-len(nodes) < asg.MinSize() {
		return fmt.Errorf("Goes below minsize. Can not delete %v nodes", len(nodes))
	}

	nodeIDs := make([]string, len(nodes))
	for i, node := range nodes {
		if vm, ok := asg.cluster.VirtualMachineMap[node.Name]; ok {
			nodeIDs[i] = vm.ID
		} else {
			nodeIDs[i] = node.Status.NodeInfo.SystemUUID
		}
	}
	if len(nodeIDs) == 0 {
		return fmt.Errorf("Unable to fetch nodeids from %v", nodes)
	}
	cluster, err := asg.manager.removeNodesFromCluster(asg.cluster.ID, nodeIDs...)
	if err != nil {
		return err
	}
	asg.Copy(cluster)
	return nil
}

// Id returns cluster id.
func (asg *asg) Id() string {
	return asg.cluster.ID
}

// Debug returns cluster id.
func (asg *asg) Debug() string {
	js, _ := json.Marshal(asg.cluster)
	return fmt.Sprintf("Debug : %s", js)
}

// Nodes returns a list of all nodes that belong to this node group.
func (asg *asg) Nodes() ([]cloudprovider.Instance, error) {
	instances := make([]cloudprovider.Instance, len(asg.cluster.VirtualMachines))
	for i := 0; i < len(asg.cluster.VirtualMachines); i++ {
		instances[i] = cloudprovider.Instance{
			Id: asg.cluster.VirtualMachines[i].ID,
		}
	}
	return instances, nil
}

// Exist checks if the node group really exists on the cloud provider side. Allows to tell the
// theoretical node group from the real one.
func (asg *asg) Exist() bool {
	return true
}

// Autoprovisioned returns true if the node group is autoprovisioned.
func (asg *asg) Autoprovisioned() bool {
	return false
}

// Create creates the node group on the cloud provider side.
func (asg *asg) Create() (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// Delete deletes the node group on the cloud provider side.
// This will be executed only for autoprovisioned node groups, once their size drops to 0.
func (asg *asg) Delete() error {
	return cloudprovider.ErrNotImplemented
}

// TemplateNodeInfo returns a node template for this node group.
func (asg *asg) TemplateNodeInfo() (*framework.NodeInfo, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetOptions returns NodeGroupAutoscalingOptions that should be used for this particular
// NodeGroup. Returning a nil will result in using default options.
func (asg *asg) GetOptions(defaults config.NodeGroupAutoscalingOptions) (*config.NodeGroupAutoscalingOptions, error) {
	return nil, cloudprovider.ErrNotImplemented
}

func (asg *asg) Copy(cluster *service.Cluster) {
	asg.cluster = cluster
}
