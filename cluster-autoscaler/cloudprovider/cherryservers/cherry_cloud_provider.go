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

package cherryservers

import (
	"io"
	"os"
	"regexp"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/config/dynamic"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	klog "k8s.io/klog/v2"
)

const (
	// ProviderName is the cloud provider name for Cherry Servers
	ProviderName = "cherryservers"
	// GPULabel is the label added to nodes with GPU resource.
	GPULabel = "cherryservers.com/gpu"
	// DefaultControllerNodeLabelKey is the label added to Master/Controller to identify as
	// master/controller node.
	DefaultControllerNodeLabelKey = "node-role.kubernetes.io/master"
	// ControllerNodeIdentifierEnv is the string for the environment variable.
	ControllerNodeIdentifierEnv = "CHERRY_CONTROLLER_NODE_IDENTIFIER_LABEL"
)

var (
	availableGPUTypes = map[string]struct{}{}
)

// cherryCloudProvider implements CloudProvider interface from cluster-autoscaler/cloudprovider module.
type cherryCloudProvider struct {
	cherryManager       cherryManager
	resourceLimiter     *cloudprovider.ResourceLimiter
	nodeGroups          []cherryNodeGroup
	controllerNodeLabel string
}

func buildCherryCloudProvider(cherryManager cherryManager, resourceLimiter *cloudprovider.ResourceLimiter) (*cherryCloudProvider, error) {
	controllerNodeLabel := os.Getenv(ControllerNodeIdentifierEnv)
	if controllerNodeLabel == "" {
		klog.V(3).Infof("env %s not set, using default: %s", ControllerNodeIdentifierEnv, DefaultControllerNodeLabelKey)
		controllerNodeLabel = DefaultControllerNodeLabelKey
	}

	ccp := &cherryCloudProvider{
		cherryManager:       cherryManager,
		resourceLimiter:     resourceLimiter,
		nodeGroups:          []cherryNodeGroup{},
		controllerNodeLabel: controllerNodeLabel,
	}
	return ccp, nil
}

// Name returns the name of the cloud provider.
func (ccp *cherryCloudProvider) Name() string {
	return ProviderName
}

// GPULabel returns the label added to nodes with GPU resource.
func (ccp *cherryCloudProvider) GPULabel() string {
	return GPULabel
}

// GetAvailableGPUTypes return all available GPU types cloud provider supports
func (ccp *cherryCloudProvider) GetAvailableGPUTypes() map[string]struct{} {
	return availableGPUTypes
}

// GetNodeGpuConfig returns the label, type and resource name for the GPU added to node. If node doesn't have
// any GPUs, it returns nil.
func (ccp *cherryCloudProvider) GetNodeGpuConfig(node *apiv1.Node) *cloudprovider.GpuConfig {
	return gpu.GetNodeGPUFromCloudProvider(ccp, node)
}

// NodeGroups returns all node groups managed by this cloud provider.
func (ccp *cherryCloudProvider) NodeGroups() []cloudprovider.NodeGroup {
	groups := make([]cloudprovider.NodeGroup, len(ccp.nodeGroups))
	for i := range ccp.nodeGroups {
		groups[i] = &ccp.nodeGroups[i]
	}
	return groups
}

// AddNodeGroup appends a node group to the list of node groups managed by this cloud provider.
func (ccp *cherryCloudProvider) AddNodeGroup(group cherryNodeGroup) {
	ccp.nodeGroups = append(ccp.nodeGroups, group)
}

// NodeGroupForNode returns the node group that a given node belongs to.
//
// Since only a single node group is currently supported, the first node group is always returned.
func (ccp *cherryCloudProvider) NodeGroupForNode(node *apiv1.Node) (cloudprovider.NodeGroup, error) {
	// ignore control plane nodes
	if _, found := node.ObjectMeta.Labels[ccp.controllerNodeLabel]; found {
		return nil, nil
	}
	nodeGroupId, err := ccp.cherryManager.NodeGroupForNode(node.ObjectMeta.Labels, node.Spec.ProviderID)
	if err != nil {
		return nil, err
	}
	if nodeGroupId == "" {
		return nil, nil
	}
	for i, nodeGroup := range ccp.nodeGroups {
		if nodeGroup.Id() == nodeGroupId {
			return &(ccp.nodeGroups[i]), nil
		}
	}
	return nil, nil
}

// HasInstance returns whether a given node has a corresponding instance in this cloud provider
func (ccp *cherryCloudProvider) HasInstance(node *apiv1.Node) (bool, error) {
	return true, cloudprovider.ErrNotImplemented
}

// Pricing returns pricing model for this cloud provider or error if not available.
func (ccp *cherryCloudProvider) Pricing() (cloudprovider.PricingModel, errors.AutoscalerError) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetAvailableMachineTypes is not implemented.
func (ccp *cherryCloudProvider) GetAvailableMachineTypes() ([]string, error) {
	return []string{}, nil
}

// NewNodeGroup is not implemented.
func (ccp *cherryCloudProvider) NewNodeGroup(machineType string, labels map[string]string, systemLabels map[string]string,
	taints []apiv1.Taint, extraResources map[string]resource.Quantity) (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetResourceLimiter returns resource constraints for the cloud provider
func (ccp *cherryCloudProvider) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	return ccp.resourceLimiter, nil
}

// Refresh is called before every autoscaler main loop.
//
// Currently only prints debug information.
func (ccp *cherryCloudProvider) Refresh() error {
	for _, nodegroup := range ccp.nodeGroups {
		klog.V(3).Info(nodegroup.Debug())
	}
	return nil
}

// Cleanup currently does nothing.
func (ccp *cherryCloudProvider) Cleanup() error {
	return nil
}

// BuildCherry is called by the autoscaler to build a Cherry Servers cloud provider.
//
// The cherryManager is created here, and the node groups are created
// based on the specs provided via the command line parameters.
func BuildCherry(opts config.AutoscalingOptions, do cloudprovider.NodeGroupDiscoveryOptions, rl *cloudprovider.ResourceLimiter) cloudprovider.CloudProvider {
	var config io.ReadCloser

	if opts.CloudConfig != "" {
		var err error
		config, err = os.Open(opts.CloudConfig)
		if err != nil {
			klog.Fatalf("Couldn't open cloud provider configuration %s: %#v", opts.CloudConfig, err)
		}
		defer config.Close()
	}

	manager, err := createCherryManager(config, do, opts)
	if err != nil {
		klog.Fatalf("Failed to create cherry manager: %v", err)
	}

	provider, err := buildCherryCloudProvider(manager, rl)
	if err != nil {
		klog.Fatalf("Failed to create cherry cloud provider: %v", err)
	}

	if len(do.NodeGroupSpecs) == 0 {
		klog.Fatalf("Must specify at least one node group with --nodes=<min>:<max>:<name>,...")
	}

	validNodepoolName := regexp.MustCompile(`^[a-z0-9A-Z]+[a-z0-9A-Z\-\.\_]*[a-z0-9A-Z]+$|^[a-z0-9A-Z]{1}$`)

	for _, nodegroupSpec := range do.NodeGroupSpecs {
		spec, err := dynamic.SpecFromString(nodegroupSpec, scaleToZeroSupported)
		if err != nil {
			klog.Fatalf("Could not parse node group spec %s: %v", nodegroupSpec, err)
		}

		if !validNodepoolName.MatchString(spec.Name) || len(spec.Name) > 63 {
			klog.Fatalf("Invalid nodepool name: %s\nMust be a valid kubernetes label value", spec.Name)
		}

		targetSize, err := manager.nodeGroupSize(spec.Name)
		if err != nil {
			klog.Fatalf("Could not set current nodes in node group: %v", err)
		}
		ng := newCherryNodeGroup(manager, spec.Name, spec.MinSize, spec.MaxSize, targetSize, waitForStatusTimeStep, deleteNodesBatchingDelay)

		provider.AddNodeGroup(ng)
	}

	return provider
}
