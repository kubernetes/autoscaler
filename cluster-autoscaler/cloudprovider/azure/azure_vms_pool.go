/*
Copyright 2017 The Kubernetes Authors.

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

package azure

import (
	"fmt"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2022-08-01/compute"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/config/dynamic"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
)

// VMsPool is single instance VM pool
// this is a placeholder for now, no real implementation
type VMsPool struct {
	azureRef
	manager       *AzureManager
	resourceGroup string

	minSize int
	maxSize int

	curSize int64
	// sizeMutex       sync.Mutex
	// lastSizeRefresh time.Time
}

// NewVMsPool creates a new VMsPool
func NewVMsPool(spec *dynamic.NodeGroupSpec, am *AzureManager) *VMsPool {
	nodepool := &VMsPool{
		azureRef: azureRef{
			Name: spec.Name,
		},

		manager:       am,
		resourceGroup: am.config.ResourceGroup,

		curSize: -1,
		minSize: spec.MinSize,
		maxSize: spec.MaxSize,
	}

	return nodepool
}

// MinSize returns the minimum size the cluster is allowed to scaled down
// to as provided by the node spec in --node parameter.
func (agentPool *VMsPool) MinSize() int {
	return agentPool.minSize
}

// Exist is always true since we are initialized with an existing agentpool
func (agentPool *VMsPool) Exist() bool {
	return true
}

// Create creates the node group on the cloud provider side.
func (agentPool *VMsPool) Create() (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrAlreadyExist
}

// Delete deletes the node group on the cloud provider side.
func (agentPool *VMsPool) Delete() error {
	return cloudprovider.ErrNotImplemented
}

// Autoprovisioned is always false since we are initialized with an existing agentpool
func (agentPool *VMsPool) Autoprovisioned() bool {
	return false
}

// GetOptions returns NodeGroupAutoscalingOptions that should be used for this particular
// NodeGroup. Returning a nil will result in using default options.
func (agentPool *VMsPool) GetOptions(defaults config.NodeGroupAutoscalingOptions) (*config.NodeGroupAutoscalingOptions, error) {
	// TODO(wenxuan): Implement this method
	return nil, cloudprovider.ErrNotImplemented
}

// MaxSize returns the maximum size scale limit provided by --node
// parameter to the autoscaler main
func (agentPool *VMsPool) MaxSize() int {
	return agentPool.maxSize
}

// TargetSize returns the current TARGET size of the node group. It is possible that the
// number is different from the number of nodes registered in Kubernetes.
func (agentPool *VMsPool) TargetSize() (int, error) {
	// TODO(wenxuan): Implement this method
	return -1, cloudprovider.ErrNotImplemented
}

// IncreaseSize increase the size through a PUT AP call. It calculates the expected size
// based on a delta provided as parameter
func (agentPool *VMsPool) IncreaseSize(delta int) error {
	// TODO(wenxuan): Implement this method
	return cloudprovider.ErrNotImplemented
}

// DeleteNodes extracts the providerIDs from the node spec and
// delete or deallocate the nodes from the agent pool based on the scale down policy.
func (agentPool *VMsPool) DeleteNodes(nodes []*apiv1.Node) error {
	// TODO(wenxuan): Implement this method
	return cloudprovider.ErrNotImplemented
}

// ForceDeleteNodes deletes nodes from the group regardless of constraints.
func (agentPool *VMsPool) ForceDeleteNodes(nodes []*apiv1.Node) error {
	return cloudprovider.ErrNotImplemented
}

// DecreaseTargetSize decreases the target size of the node group.
func (agentPool *VMsPool) DecreaseTargetSize(delta int) error {
	// TODO(wenxuan): Implement this method
	return cloudprovider.ErrNotImplemented
}

// Id returns the name of the agentPool
func (agentPool *VMsPool) Id() string {
	return agentPool.azureRef.Name
}

// Debug returns a string with basic details of the agentPool
func (agentPool *VMsPool) Debug() string {
	return fmt.Sprintf("%s (%d:%d)", agentPool.Id(), agentPool.MinSize(), agentPool.MaxSize())
}

func (agentPool *VMsPool) getVMsFromCache() ([]compute.VirtualMachine, error) {
	// vmsPoolMap is a map of agent pool name to the list of virtual machines
	vmsPoolMap := agentPool.manager.azureCache.getVirtualMachines()
	if _, ok := vmsPoolMap[agentPool.Name]; !ok {
		return []compute.VirtualMachine{}, fmt.Errorf("vms pool %s not found in the cache", agentPool.Name)
	}

	return vmsPoolMap[agentPool.Name], nil
}

// Nodes returns the list of nodes in the vms agentPool.
func (agentPool *VMsPool) Nodes() ([]cloudprovider.Instance, error) {
	vms, err := agentPool.getVMsFromCache()
	if err != nil {
		return nil, err
	}

	nodes := make([]cloudprovider.Instance, 0, len(vms))
	for _, vm := range vms {
		if vm.ID == nil || len(*vm.ID) == 0 {
			continue
		}
		resourceID, err := convertResourceGroupNameToLower("azure://" + *vm.ID)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, cloudprovider.Instance{Id: resourceID})
	}

	return nodes, nil
}

// TemplateNodeInfo is not implemented.
func (agentPool *VMsPool) TemplateNodeInfo() (*framework.NodeInfo, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// AtomicIncreaseSize is not implemented.
func (agentPool *VMsPool) AtomicIncreaseSize(delta int) error {
	return cloudprovider.ErrNotImplemented
}
