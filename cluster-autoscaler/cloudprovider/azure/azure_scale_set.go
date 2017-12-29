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
	"strings"

	"github.com/Azure/azure-sdk-for-go/arm/compute"
	"github.com/golang/glog"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/kubernetes/plugin/pkg/scheduler/schedulercache"
)

// ScaleSet implements NodeGroup interface.
type ScaleSet struct {
	AzureRef
	*AzureManager

	minSize int
	maxSize int
}

// NewScaleSet creates a new NewScaleSet.
func NewScaleSet(name string, minSize, maxSize int, az *AzureManager) (*ScaleSet, error) {
	scaleSet := &ScaleSet{
		AzureRef: AzureRef{
			Name: name,
		},
		minSize:      minSize,
		maxSize:      maxSize,
		AzureManager: az,
	}

	return scaleSet, nil
}

// MinSize returns minimum size of the node group.
func (scaleSet *ScaleSet) MinSize() int {
	return scaleSet.minSize
}

// Exist checks if the node group really exists on the cloud provider side. Allows to tell the
// theoretical node group from the real one.
func (scaleSet *ScaleSet) Exist() bool {
	return true
}

// Create creates the node group on the cloud provider side.
func (scaleSet *ScaleSet) Create() error {
	return cloudprovider.ErrAlreadyExist
}

// Delete deletes the node group on the cloud provider side.
// This will be executed only for autoprovisioned node groups, once their size drops to 0.
func (scaleSet *ScaleSet) Delete() error {
	return cloudprovider.ErrNotImplemented
}

// Autoprovisioned returns true if the node group is autoprovisioned.
func (scaleSet *ScaleSet) Autoprovisioned() bool {
	return false
}

// MaxSize returns maximum size of the node group.
func (scaleSet *ScaleSet) MaxSize() int {
	return scaleSet.maxSize
}

// GetScaleSetSize gets Scale Set size.
func (scaleSet *ScaleSet) GetScaleSetSize() (int64, error) {
	glog.V(5).Infof("Get scale set size for %q", scaleSet.Name)
	resourceGroup := scaleSet.config.ResourceGroup
	set, err := scaleSet.virtualMachineScaleSetsClient.Get(resourceGroup, scaleSet.Name)
	if err != nil {
		return -1, err
	}
	glog.V(5).Infof("Returning scale set (%q) capacity: %d\n", scaleSet.Name, *set.Sku.Capacity)
	return *set.Sku.Capacity, nil
}

// SetScaleSetSize sets ScaleSet size.
func (scaleSet *ScaleSet) SetScaleSetSize(size int64) error {
	resourceGroup := scaleSet.config.ResourceGroup
	op, err := scaleSet.virtualMachineScaleSetsClient.Get(resourceGroup, scaleSet.Name)
	if err != nil {
		return err
	}

	op.Sku.Capacity = &size
	op.VirtualMachineScaleSetProperties.ProvisioningState = nil
	cancel := make(chan struct{})

	_, errChan := scaleSet.virtualMachineScaleSetsClient.CreateOrUpdate(resourceGroup, scaleSet.Name, op, cancel)
	return <-errChan
}

// TargetSize returns the current TARGET size of the node group. It is possible that the
// number is different from the number of nodes registered in Kubernetes.
func (scaleSet *ScaleSet) TargetSize() (int, error) {
	size, err := scaleSet.GetScaleSetSize()
	return int(size), err
}

// IncreaseSize increases Scale Set size
func (scaleSet *ScaleSet) IncreaseSize(delta int) error {
	if delta <= 0 {
		return fmt.Errorf("size increase must be positive")
	}

	size, err := scaleSet.GetScaleSetSize()
	if err != nil {
		return err
	}

	if int(size)+delta > scaleSet.MaxSize() {
		return fmt.Errorf("size increase too large - desired:%d max:%d", int(size)+delta, scaleSet.MaxSize())
	}

	return scaleSet.SetScaleSetSize(size + int64(delta))
}

// GetScaleSetVms returns list of nodes for the given scale set.
func (scaleSet *ScaleSet) GetScaleSetVms() ([]string, error) {
	resourceGroup := scaleSet.config.ResourceGroup
	instances, err := scaleSet.virtualMachineScaleSetVMsClient.List(resourceGroup, scaleSet.Name, "", "", "")

	if err != nil {
		glog.V(4).Infof("VirtualMachineScaleSetVMsClient.List failed for %s: %v", scaleSet.Name, err)
		return []string{}, err
	}

	result := make([]string, 0)
	for _, instance := range *instances.Value {
		// Convert to lower because instance.ID is in different in different API calls (e.g. GET and LIST).
		name := "azure://" + strings.ToLower(*instance.ID)
		result = append(result, name)
	}
	return result, nil

}

// DecreaseTargetSize decreases the target size of the node group. This function
// doesn't permit to delete any existing node and can be used only to reduce the
// request for new nodes that have not been yet fulfilled. Delta should be negative.
// It is assumed that cloud provider will not delete the existing nodes if the size
// when there is an option to just decrease the target.
func (scaleSet *ScaleSet) DecreaseTargetSize(delta int) error {
	if delta >= 0 {
		return fmt.Errorf("size decrease size must be negative")
	}

	size, err := scaleSet.GetScaleSetSize()
	if err != nil {
		return err
	}

	nodes, err := scaleSet.GetScaleSetVms()
	if err != nil {
		return err
	}

	if int(size)+delta < len(nodes) {
		return fmt.Errorf("attempt to delete existing nodes targetSize:%d delta:%d existingNodes: %d",
			size, delta, len(nodes))
	}

	return scaleSet.SetScaleSetSize(size + int64(delta))
}

// Belongs returns true if the given node belongs to the NodeGroup.
func (scaleSet *ScaleSet) Belongs(node *apiv1.Node) (bool, error) {
	glog.V(6).Infof("Check if node belongs to this scale set: scaleset:%v, node:%v\n", scaleSet, node)

	ref := &AzureRef{
		Name: strings.ToLower(node.Spec.ProviderID),
	}

	targetAsg, err := scaleSet.GetNodeGroupForInstance(ref)
	if err != nil {
		return false, err
	}
	if targetAsg == nil {
		return false, fmt.Errorf("%s doesn't belong to a known scale set", node.Name)
	}
	if targetAsg.Id() != scaleSet.Id() {
		return false, nil
	}
	return true, nil
}

// DeleteInstances deletes the given instances. All instances must be controlled by the same ASG.
func (scaleSet *ScaleSet) DeleteInstances(instances []*AzureRef) error {
	if len(instances) == 0 {
		return nil
	}

	commonAsg, err := scaleSet.GetNodeGroupForInstance(instances[0])
	if err != nil {
		return err
	}

	for _, instance := range instances {
		asg, err := scaleSet.GetNodeGroupForInstance(instance)
		if err != nil {
			return err
		}

		if asg != commonAsg {
			return fmt.Errorf("cannot delete instance (%s) which don't belong to the same Scale Set (%q)", instance.GetKey(), commonAsg)
		}
	}

	instanceIds := scaleSet.GetInstanceIDs(instances)
	requiredIds := &compute.VirtualMachineScaleSetVMInstanceRequiredIDs{
		InstanceIds: &instanceIds,
	}
	cancel := make(chan struct{})
	resourceGroup := scaleSet.config.ResourceGroup
	_, errChan := scaleSet.virtualMachineScaleSetsClient.DeleteInstances(resourceGroup, commonAsg.Id(), *requiredIds, cancel)
	return <-errChan
}

// DeleteNodes deletes the nodes from the group.
func (scaleSet *ScaleSet) DeleteNodes(nodes []*apiv1.Node) error {
	glog.V(8).Infof("Delete nodes requested: %v\n", nodes)
	size, err := scaleSet.GetScaleSetSize()
	if err != nil {
		return err
	}

	if int(size) <= scaleSet.MinSize() {
		return fmt.Errorf("min size reached, nodes will not be deleted")
	}

	refs := make([]*AzureRef, 0, len(nodes))
	for _, node := range nodes {
		belongs, err := scaleSet.Belongs(node)
		if err != nil {
			return err
		}

		if belongs != true {
			return fmt.Errorf("%s belongs to a different asg than %s", node.Name, scaleSet.Id())
		}

		azureRef := &AzureRef{
			Name: strings.ToLower(node.Spec.ProviderID),
		}
		refs = append(refs, azureRef)
	}

	return scaleSet.DeleteInstances(refs)
}

// Id returns ScaleSet id.
func (scaleSet *ScaleSet) Id() string {
	return scaleSet.Name
}

// Debug returns a debug string for the Scale Set.
func (scaleSet *ScaleSet) Debug() string {
	return fmt.Sprintf("%s (%d:%d)", scaleSet.Id(), scaleSet.MinSize(), scaleSet.MaxSize())
}

// TemplateNodeInfo returns a node template for this scale set.
func (scaleSet *ScaleSet) TemplateNodeInfo() (*schedulercache.NodeInfo, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// Nodes returns a list of all nodes that belong to this node group.
func (scaleSet *ScaleSet) Nodes() ([]string, error) {
	return scaleSet.GetScaleSetVms()
}
