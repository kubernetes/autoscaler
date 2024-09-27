/*
Copyright 2021 The Kubernetes Authors.

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

package tencentcloud

import (
	"fmt"
	"regexp"
	"strings"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
)

// TcRef contains a reference to some entity in Tencentcloud/TKE world.
type TcRef struct {
	ID   string
	Zone string
}

func (ref TcRef) String() string {
	if ref.Zone == "" {
		return ref.ID
	}
	return fmt.Sprintf("%s/%s", ref.Zone, ref.ID)
}

// ToProviderID converts tcRef to string in format used as ProviderId in Node object.
func (ref TcRef) ToProviderID() string {
	return fmt.Sprintf("qcloud:///%s/%s", ref.Zone, ref.ID)
}

// TcRefFromProviderID creates InstanceConfig object from provider id which
// must be in format: qcloud:///100003/ins-3ven36lk
func TcRefFromProviderID(id string) (TcRef, error) {
	validIDRegex := regexp.MustCompile(`^qcloud\:\/\/\/[-0-9a-z]*\/[-0-9a-z]*$`)
	if validIDRegex.FindStringSubmatch(id) == nil {
		return TcRef{}, fmt.Errorf("valid provider id: expected format qcloud:///zoneid/ins-<name>, got %v", id)
	}
	splitted := strings.Split(id[10:], "/")
	return TcRef{
		ID: splitted[1],
	}, nil
}

// Asg implements NodeGroup interface.
type Asg interface {
	cloudprovider.NodeGroup

	TencentcloudRef() TcRef
	GetScalingType() string
	SetScalingType(string)
}

type tcAsg struct {
	tencentcloudRef TcRef

	tencentcloudManager TencentcloudManager
	minSize             int
	maxSize             int
	scalingType         string
}

// TencentcloudRef returns ASG's TencentcloudRef
func (asg *tcAsg) TencentcloudRef() TcRef {
	return asg.tencentcloudRef
}

// GetScalingType returns ASG's scaling type.
func (asg *tcAsg) GetScalingType() string {
	return asg.scalingType
}

// SetScalingType set ASG's scaling type.
func (asg *tcAsg) SetScalingType(scalingType string) {
	asg.scalingType = scalingType
}

// MaxSize returns maximum size of the node group.
func (asg *tcAsg) MaxSize() int {
	return asg.maxSize
}

// MinSize returns minimum size of the node group.
func (asg *tcAsg) MinSize() int {
	return asg.minSize
}

// TargetSize returns the current TARGET size of the node group. It is possible that the
// number is different from the number of nodes registered in Kuberentes.
func (asg *tcAsg) TargetSize() (int, error) {
	size, err := asg.tencentcloudManager.GetAsgSize(asg)
	return int(size), err
}

// IncreaseSize increases Asg size
func (asg *tcAsg) IncreaseSize(delta int) error {
	if delta <= 0 {
		return fmt.Errorf("size increase must be positive")
	}
	size, err := asg.tencentcloudManager.GetAsgSize(asg)
	if err != nil {
		return err
	}
	if int(size)+delta > asg.MaxSize() {
		return fmt.Errorf("size increase too large - desired:%d max:%d", int(size)+delta, asg.MaxSize())
	}
	return asg.tencentcloudManager.SetAsgSize(asg, size+int64(delta))
}

// AtomicIncreaseSize is not implemented.
func (asg *tcAsg) AtomicIncreaseSize(delta int) error {
	return cloudprovider.ErrNotImplemented
}

// DecreaseTargetSize decreases the target size of the node group. This function
// doesn't permit to delete any existing node and can be used only to reduce the
// request for new nodes that have not been yet fulfilled. Delta should be negative.
// It is assumed that cloud provider will not delete the existing nodes if the size
// when there is an option to just decrease the target.
func (asg *tcAsg) DecreaseTargetSize(delta int) error {
	if delta >= 0 {
		return fmt.Errorf("size decrease size must be negative")
	}
	size, err := asg.tencentcloudManager.GetAsgSize(asg)
	if err != nil {
		return err
	}
	if int(size)+delta < asg.MinSize() {
		return fmt.Errorf("size increase too small - desired:%d min:%d", int(size)+delta, asg.MaxSize())
	}
	nodes, err := asg.tencentcloudManager.GetAsgNodes(asg)
	if err != nil {
		return err
	}
	if int(size)+delta < len(nodes) {
		return fmt.Errorf("attempt to delete existing nodes targetSize:%d delta:%d existingNodes: %d",
			size, delta, len(nodes))
	}
	return asg.tencentcloudManager.SetAsgSize(asg, size+int64(delta))
}

// Belongs returns true if the given node belongs to the NodeGroup.
func (asg *tcAsg) Belongs(node *apiv1.Node) (bool, error) {
	if node.Spec.ProviderID == "" {
		return false, nil
	}
	ref, err := TcRefFromProviderID(node.Spec.ProviderID)
	if err != nil {
		return false, err
	}
	targetAsg, err := asg.tencentcloudManager.GetAsgForInstance(ref)
	if err != nil {
		return false, err
	}
	if targetAsg == nil {
		return false, fmt.Errorf("%s doesn't belong to a known asg", node.Name)
	}
	if targetAsg.Id() != asg.Id() {
		return false, nil
	}
	return true, nil
}

// Exist checks if the node group really exists on the cloud provider side. Allows to tell the
// theoretical node group from the real one.
func (asg *tcAsg) Exist() bool {
	return true
}

// Create creates the node group on the cloud provider side.
func (asg *tcAsg) Create() (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrAlreadyExist
}

// Delete deletes the node group on the cloud provider side.
// This will be executed only for autoprovisioned node groups, once their size drops to 0.
func (asg *tcAsg) Delete() error {
	return cloudprovider.ErrNotImplemented
}

// Autoprovisioned returns true if the node group is autoprovisioned.
func (asg *tcAsg) Autoprovisioned() bool {
	return false
}

// DeleteNodes deletes the nodes from the group.
func (asg *tcAsg) DeleteNodes(nodes []*apiv1.Node) error {
	size, err := asg.tencentcloudManager.GetAsgSize(asg)
	if err != nil {
		return err
	}

	if int(size) <= asg.MinSize() {
		return fmt.Errorf("min size reached, nodes will not be deleted")
	}

	refs := make([]TcRef, 0, len(nodes))
	for _, node := range nodes {
		belongs, err := asg.Belongs(node)
		if err != nil {
			return err
		}
		if !belongs {
			return fmt.Errorf("%s,%s belongs to a different asg than %s", node.Name, node.Spec.ProviderID, asg.Id())
		}
		tencentcloudref, err := TcRefFromProviderID(node.Spec.ProviderID)
		if err != nil {
			return err
		}
		refs = append(refs, tencentcloudref)
	}

	return asg.tencentcloudManager.DeleteInstances(refs)
}

// Id returns asg id.
func (asg *tcAsg) Id() string {
	return asg.tencentcloudRef.ID
}

// Debug returns a debug string for the Asg.
func (asg *tcAsg) Debug() string {
	return fmt.Sprintf("%s (%d:%d)", asg.Id(), asg.MinSize(), asg.MaxSize())
}

// Nodes returns a list of all nodes that belong to this node group.
func (asg *tcAsg) Nodes() ([]cloudprovider.Instance, error) {
	return asg.tencentcloudManager.GetAsgNodes(asg)
}

// TemplateNodeInfo returns a node template for this node group.
func (asg *tcAsg) TemplateNodeInfo() (*framework.NodeInfo, error) {
	node, err := asg.tencentcloudManager.GetAsgTemplateNode(asg)
	if err != nil {
		return nil, err
	}
	klog.V(4).Infof("Generate tencentcloud template: labels=%v taints=%v allocatable=%v", node.Labels, node.Spec.Taints, node.Status.Allocatable)

	nodeInfo := framework.NewNodeInfo(node, nil)
	return nodeInfo, nil
}

// GetOptions returns NodeGroupAutoscalingOptions that should be used for this particular
// NodeGroup. Returning a nil will result in using default options.
func (asg *tcAsg) GetOptions(defaults config.NodeGroupAutoscalingOptions) (*config.NodeGroupAutoscalingOptions, error) {
	return nil, nil
}
