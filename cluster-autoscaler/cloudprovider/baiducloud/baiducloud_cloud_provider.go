/*
Copyright 2018 The Kubernetes Authors.

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

package baiducloud

import (
	"fmt"
	"io"
	"os"
	"strings"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/config/dynamic"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	klog "k8s.io/klog/v2"
)

const (
	// GPULabel is the label added to nodes with GPU resource.
	GPULabel = "baidu/nvidia_name"
)

var (
	availableGPUTypes = map[string]struct{}{
		"nTeslaV100":    {},
		"nTeslaP40":     {},
		"nTeslaP4":      {},
		"nTeslaV100-16": {},
		"nTeslaV100-32": {},
	}
)

// baiducloudCloudProvider implements CloudProvider interface.
type baiducloudCloudProvider struct {
	baiducloudManager *BaiducloudManager
	asgs              []*Asg
	resourceLimiter   *cloudprovider.ResourceLimiter
}

// BuildBaiducloud builds baiducloud cloud provider, manager etc.
func BuildBaiducloud(opts config.AutoscalingOptions, do cloudprovider.NodeGroupDiscoveryOptions, rl *cloudprovider.ResourceLimiter) cloudprovider.CloudProvider {
	var cfg io.ReadCloser
	if opts.CloudConfig != "" {
		var err error
		cfg, err = os.Open(opts.CloudConfig)
		if err != nil {
			klog.Fatalf("Couldn't open cloud provider configuration %s: %#v", opts.CloudConfig, err)
		}
		defer cfg.Close()
	}

	manager, err := CreateBaiducloudManager(cfg)
	if err != nil {
		klog.Fatalf("Failed to create Baiducloud Manager: %v", err)
	}

	provider, err := BuildBaiducloudCloudProvider(manager, do, rl)
	if err != nil {
		klog.Fatalf("Failed to create Baiducloud cloud provider: %v", err)
	}
	return provider
}

// BuildBaiducloudCloudProvider builds CloudProvider implementation for Baiducloud.
func BuildBaiducloudCloudProvider(manager *BaiducloudManager, discoveryOpts cloudprovider.NodeGroupDiscoveryOptions, resourceLimiter *cloudprovider.ResourceLimiter) (cloudprovider.CloudProvider, error) {
	if discoveryOpts.StaticDiscoverySpecified() {
		return buildStaticallyDiscoveringProvider(manager, discoveryOpts.NodeGroupSpecs, resourceLimiter)
	}
	if discoveryOpts.AutoDiscoverySpecified() {
		return nil, fmt.Errorf("only support static discovery scaling group in baiducloud for now")
	}
	return nil, fmt.Errorf("failed to build baiducloud provider: node group specs must be specified")
}

func buildStaticallyDiscoveringProvider(manager *BaiducloudManager, specs []string, resourceLimiter *cloudprovider.ResourceLimiter) (*baiducloudCloudProvider, error) {
	bcp := &baiducloudCloudProvider{
		baiducloudManager: manager,
		asgs:              make([]*Asg, 0),
		resourceLimiter:   resourceLimiter,
	}
	if len(specs) > 200 {
		return nil, fmt.Errorf("currently, baiducloud cloud provider not support ASG‘s number > 200")
	}
	for _, spec := range specs {
		if err := bcp.addNodeGroup(spec); err != nil {
			return nil, err
		}
	}
	klog.V(4).Infof("create baiducloudCloudProvider success.")
	return bcp, nil
}

// addNodeGroup adds node group defined in string spec. Format:
// minNodes:maxNodes:asgName
func (baiducloud *baiducloudCloudProvider) addNodeGroup(spec string) error {
	asg, err := buildAsgFromSpec(spec, baiducloud.baiducloudManager)
	if err != nil {
		return err
	}
	baiducloud.addAsg(asg)
	return nil
}

func buildAsgFromSpec(value string, baiducloudManager *BaiducloudManager) (*Asg, error) {
	spec, err := dynamic.SpecFromString(value, false)
	if err != nil {
		return nil, fmt.Errorf("failed to parse node group spec: %v", err)
	}
	asg := buildAsg(baiducloudManager, spec.MinSize, spec.MaxSize, spec.Name)
	return asg, nil
}

func buildAsg(baiducloudManager *BaiducloudManager, minSize int, maxSize int, name string) *Asg {
	return &Asg{
		baiducloudManager: baiducloudManager,
		minSize:           minSize,
		maxSize:           maxSize,
		BaiducloudRef: BaiducloudRef{
			Name: name,
		},
	}
}

// addAsg adds and registers an Asg to this cloud provider.
func (baiducloud *baiducloudCloudProvider) addAsg(asg *Asg) {
	baiducloud.asgs = append(baiducloud.asgs, asg)
	baiducloud.baiducloudManager.RegisterAsg(asg)
}

// Name returns name of the cloud provider.
func (baiducloud *baiducloudCloudProvider) Name() string {
	return cloudprovider.BaiducloudProviderName
}

// NodeGroups returns all node groups configured for this cloud provider.
func (baiducloud *baiducloudCloudProvider) NodeGroups() []cloudprovider.NodeGroup {
	result := make([]cloudprovider.NodeGroup, 0, len(baiducloud.asgs))
	for _, asg := range baiducloud.asgs {
		result = append(result, asg)
	}
	return result
}

// GPULabel returns the label added to nodes with GPU resource.
func (baiducloud *baiducloudCloudProvider) GPULabel() string {
	return GPULabel
}

// GetAvailableGPUTypes returns all available GPU types cloud provider supports.
func (baiducloud *baiducloudCloudProvider) GetAvailableGPUTypes() map[string]struct{} {
	return availableGPUTypes
}

// GetNodeGpuConfig returns the label, type and resource name for the GPU added to node. If node doesn't have
// any GPUs, it returns nil.
func (baiducloud *baiducloudCloudProvider) GetNodeGpuConfig(node *apiv1.Node) *cloudprovider.GpuConfig {
	return gpu.GetNodeGPUFromCloudProvider(baiducloud, node)
}

// NodeGroupForNode returns the node group for the given node, nil if the node
// should not be processed by cluster autoscaler, or non-nil error if such
// occurred. Must be implemented.
func (baiducloud *baiducloudCloudProvider) NodeGroupForNode(node *apiv1.Node) (cloudprovider.NodeGroup, error) {
	splitted := strings.Split(node.Spec.ProviderID, "//")
	if len(splitted) != 2 {
		return nil, fmt.Errorf("parse ProviderID failed: %v", node.Spec.ProviderID)
	}
	asg, err := baiducloud.baiducloudManager.GetAsgForInstance(splitted[1])
	return asg, err
}

// HasInstance returns whether a given node has a corresponding instance in this cloud provider
func (baiducloud *baiducloudCloudProvider) HasInstance(*apiv1.Node) (bool, error) {
	return true, cloudprovider.ErrNotImplemented
}

// Pricing returns pricing model for this cloud provider or error if not available.
// Implementation optional.
func (baiducloud *baiducloudCloudProvider) Pricing() (cloudprovider.PricingModel, errors.AutoscalerError) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetAvailableMachineTypes get all machine types that can be requested from the cloud provider.
// Implementation optional.
func (baiducloud *baiducloudCloudProvider) GetAvailableMachineTypes() ([]string, error) {
	return []string{}, cloudprovider.ErrNotImplemented
}

// NewNodeGroup builds a theoretical node group based on the node definition provided. The node group is not automatically
// created on the cloud provider side. The node group is not returned by NodeGroups() until it is created.
// Implementation optional.
func (baiducloud *baiducloudCloudProvider) NewNodeGroup(machineType string, labels map[string]string, systemLabels map[string]string,
	taints []apiv1.Taint, extraResources map[string]resource.Quantity) (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetResourceLimiter returns struct containing limits (max, min) for resources (cores, memory etc.).
func (baiducloud *baiducloudCloudProvider) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	return baiducloud.resourceLimiter, nil
}

// Cleanup cleans up open resources before the cloud provider is destroyed, i.e. go routines etc.
func (baiducloud *baiducloudCloudProvider) Cleanup() error {
	return nil
}

// Refresh is called before every main loop and can be used to dynamically update cloud provider state.
// In particular the list of node groups returned by NodeGroups can change as a result of CloudProvider.Refresh().
func (baiducloud *baiducloudCloudProvider) Refresh() error {
	return nil
}

// BaiducloudRef contains a reference to some entity in baiducloud world.
type BaiducloudRef struct {
	Name string
}

// Asg implements NodeGroup interface.
type Asg struct {
	BaiducloudRef
	baiducloudManager *BaiducloudManager

	minSize int
	maxSize int
}

// MaxSize returns maximum size of the node group.
func (asg *Asg) MaxSize() int {
	return asg.maxSize
}

// MinSize returns minimum size of the node group.
func (asg *Asg) MinSize() int {
	return asg.minSize
}

// TargetSize returns the current target size of the node group. It is possible that the
// number of nodes in Kubernetes is different at the moment but should be equal
// to Size() once everything stabilizes (new nodes finish startup and registration or
// removed nodes are deleted completely). Implementation required.
func (asg *Asg) TargetSize() (int, error) {
	size, err := asg.baiducloudManager.GetAsgSize(asg)
	return int(size), err
}

// IncreaseSize increases the size of the node group. To delete a node you need
// to explicitly name it and use DeleteNode. This function should wait until
// node group size is updated. Implementation required.
func (asg *Asg) IncreaseSize(delta int) error {
	if delta <= 0 {
		return fmt.Errorf("size increase must be positive")
	}
	size, err := asg.baiducloudManager.GetAsgSize(asg)
	if err != nil {
		return err
	}
	if int(size)+delta > asg.MaxSize() {
		return fmt.Errorf("size increase too large - desired:%d max:%d", int(size)+delta, asg.MaxSize())
	}
	return asg.baiducloudManager.ScaleUpCluster(delta, asg.Name)
}

// AtomicIncreaseSize is not implemented.
func (asg *Asg) AtomicIncreaseSize(delta int) error {
	return cloudprovider.ErrNotImplemented
}

// DeleteNodes deletes nodes from this node group. Error is returned either on
// failure or if the given node doesn't belong to this node group. This function
// should wait until node group size is updated. Implementation required.
func (asg *Asg) DeleteNodes(nodes []*apiv1.Node) error {
	size, err := asg.baiducloudManager.GetAsgSize(asg)
	if err != nil {
		return err
	}
	if int(size) <= asg.MinSize() {
		return fmt.Errorf("min size reached, nodes will not be deleted")
	}
	nodeID := make([]string, len(nodes))
	for _, node := range nodes {
		klog.Infof("Delete node : %s", node.Spec.ProviderID)
		splitted := strings.Split(node.Spec.ProviderID, "//")
		if len(splitted) != 2 {
			return fmt.Errorf("Not expected name: %s\n", node.Spec.ProviderID)
		}
		belong, err := asg.Belongs(splitted[1])
		if err != nil {
			klog.Errorf("failed to check whether node:%s is belong to asg:%s", node.GetName(), asg.Id())
			return err
		}
		if !belong {
			return fmt.Errorf("%s belongs to a different asg than %s", node.Name, asg.Id())
		}
		// todo: if the node exists.
		nodeID = append(nodeID, splitted[1])
	}
	return asg.baiducloudManager.ScaleDownCluster(nodeID)
}

// Belongs returns true if the given node belongs to the NodeGroup.
func (asg *Asg) Belongs(instanceID string) (bool, error) {
	targetAsg, err := asg.baiducloudManager.GetAsgForInstance(instanceID)
	if err != nil {
		return false, err
	}
	if targetAsg == nil {
		return false, fmt.Errorf("%s doesn't belong to a known Asg", instanceID)
	}
	if targetAsg.Id() != asg.Id() {
		return false, nil
	}
	return true, nil
}

// DecreaseTargetSize decreases the target size of the node group. This function
// doesn't permit to delete any existing node and can be used only to reduce the
// request for new nodes that have not been yet fulfilled. Delta should be negative.
// It is assumed that cloud provider will not delete the existing nodes when there
// is an option to just decrease the target. Implementation required.
func (asg *Asg) DecreaseTargetSize(delta int) error {
	return cloudprovider.ErrNotImplemented
}

// Id returns an unique identifier of the node group.
func (asg *Asg) Id() string {
	return asg.Name
}

// Debug returns a string containing all information regarding this node group.
func (asg *Asg) Debug() string {
	return fmt.Sprintf("%s (%d:%d)", asg.Id(), asg.MinSize(), asg.MaxSize())
}

// Nodes returns a list of all nodes that belong to this node group.
// It is required that Instance objects returned by this method have Id field set.
// Other fields are optional.
func (asg *Asg) Nodes() ([]cloudprovider.Instance, error) {
	asgNodes, err := asg.baiducloudManager.GetAsgNodes(asg)
	if err != nil {
		return nil, err
	}
	instances := make([]cloudprovider.Instance, len(asgNodes))

	for i, asgNode := range asgNodes {
		instances[i] = cloudprovider.Instance{Id: asgNode}
	}
	return instances, nil
}

// TemplateNodeInfo returns a framework.NodeInfo structure of an empty
// (as if just started) node. This will be used in scale-up simulations to
// predict what would a new node look like if a node group was expanded. The returned
// NodeInfo is expected to have a fully populated Node object, with all of the labels,
// capacity and allocatable information as well as all pods that are started on
// the node by default, using manifest (most likely only kube-proxy). Implementation optional.
func (asg *Asg) TemplateNodeInfo() (*framework.NodeInfo, error) {
	template, err := asg.baiducloudManager.getAsgTemplate(asg.Name)
	if err != nil {
		return nil, err
	}
	node, err := asg.baiducloudManager.buildNodeFromTemplate(asg, template)
	if err != nil {
		return nil, err
	}
	nodeInfo := framework.NewNodeInfo(node, nil, &framework.PodInfo{Pod: cloudprovider.BuildKubeProxy(asg.Name)})
	return nodeInfo, nil
}

// Exist checks if the node group really exists on the cloud provider side. Allows to tell the
// theoretical node group from the real one. Implementation required.
func (asg *Asg) Exist() bool {
	return true
}

// Create creates the node group on the cloud provider side. Implementation optional.
func (asg *Asg) Create() (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrAlreadyExist
}

// Delete deletes the node group on the cloud provider side.
// This will be executed only for autoprovisioned node groups, once their size drops to 0.
// Implementation optional.
func (asg *Asg) Delete() error {
	return cloudprovider.ErrNotImplemented
}

// Autoprovisioned returns true if the node group is autoprovisioned. An autoprovisioned group
// was created by CA and can be deleted when scaled to 0.
func (asg *Asg) Autoprovisioned() bool {
	return false
}

// GetOptions returns NodeGroupAutoscalingOptions that should be used for this particular
// NodeGroup. Returning a nil will result in using default options.
func (asg *Asg) GetOptions(defaults config.NodeGroupAutoscalingOptions) (*config.NodeGroupAutoscalingOptions, error) {
	return nil, cloudprovider.ErrNotImplemented
}
