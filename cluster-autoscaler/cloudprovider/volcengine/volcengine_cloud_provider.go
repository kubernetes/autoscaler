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

package volcengine

import (
	"fmt"
	"io"
	"os"
	"strings"

	"gopkg.in/gcfg.v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/config/dynamic"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	"k8s.io/klog/v2"
)

// volcengineCloudProvider implements CloudProvider interface.
type volcengineCloudProvider struct {
	volcengineManager VolcengineManager
	resourceLimiter   *cloudprovider.ResourceLimiter
	scalingGroups     []*AutoScalingGroup
}

// Name returns name of the cloud provider.
func (v *volcengineCloudProvider) Name() string {
	return cloudprovider.VolcengineProviderName
}

// NodeGroups returns all node groups configured for this cloud provider.
func (v *volcengineCloudProvider) NodeGroups() []cloudprovider.NodeGroup {
	result := make([]cloudprovider.NodeGroup, 0, len(v.scalingGroups))
	for _, ng := range v.scalingGroups {
		result = append(result, ng)
	}
	return result
}

// NodeGroupForNode returns the node group for the given node, nil if the node
// should not be processed by cluster autoscaler, or non-nil error if such
// occurred. Must be implemented.
func (v *volcengineCloudProvider) NodeGroupForNode(node *apiv1.Node) (cloudprovider.NodeGroup, error) {
	instanceId, err := ecsInstanceFromProviderId(node.Spec.ProviderID)
	if err != nil {
		return nil, err
	}
	if len(instanceId) == 0 {
		klog.Warningf("Node %v has no providerId", node.Name)
		return nil, fmt.Errorf("provider id missing from node: %s", node.Name)
	}
	return v.volcengineManager.GetAsgForInstance(instanceId)
}

// HasInstance returns whether the node has corresponding instance in cloud provider,
// true if the node has an instance, false if it no longer exists
func (v *volcengineCloudProvider) HasInstance(node *apiv1.Node) (bool, error) {
	return true, cloudprovider.ErrNotImplemented
}

// Pricing returns pricing model for this cloud provider or error if not available.
// Implementation optional.
func (v *volcengineCloudProvider) Pricing() (cloudprovider.PricingModel, errors.AutoscalerError) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetAvailableMachineTypes get all machine types that can be requested from the cloud provider.
// Implementation optional.
func (v *volcengineCloudProvider) GetAvailableMachineTypes() ([]string, error) {
	return []string{}, nil
}

// NewNodeGroup builds a theoretical node group based on the node definition provided. The node group is not automatically
// created on the cloud provider side. The node group is not returned by NodeGroups() until it is created.
// Implementation optional.
func (v *volcengineCloudProvider) NewNodeGroup(machineType string, labels map[string]string, systemLabels map[string]string, taints []apiv1.Taint, extraResources map[string]resource.Quantity) (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetResourceLimiter returns struct containing limits (max, min) for resources (cores, memory etc.).
func (v *volcengineCloudProvider) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	return v.resourceLimiter, nil
}

// GPULabel returns the label added to nodes with GPU resource.
func (v *volcengineCloudProvider) GPULabel() string {
	return ""
}

// GetAvailableGPUTypes return all available GPU types cloud provider supports.
func (v *volcengineCloudProvider) GetAvailableGPUTypes() map[string]struct{} {
	return map[string]struct{}{}
}

// Cleanup cleans up open resources before the cloud provider is destroyed, i.e. go routines etc.
func (v *volcengineCloudProvider) Cleanup() error {
	return nil
}

// Refresh is called before every main loop and can be used to dynamically update cloud provider state.
// In particular the list of node groups returned by NodeGroups can change as a result of CloudProvider.Refresh().
func (v *volcengineCloudProvider) Refresh() error {
	return nil
}

// GetNodeGpuConfig returns the label, type and resource name for the GPU added to node. If node doesn't have
// any GPUs, it returns nil.
func (v *volcengineCloudProvider) GetNodeGpuConfig(node *apiv1.Node) *cloudprovider.GpuConfig {
	return gpu.GetNodeGPUFromCloudProvider(v, node)
}

func (v *volcengineCloudProvider) addNodeGroup(spec string) error {
	group, err := buildScalingGroupFromSpec(v.volcengineManager, spec)
	if err != nil {
		klog.Errorf("Failed to build scaling group from spec: %v", err)
		return err
	}
	v.addAsg(group)
	return nil
}

func (v *volcengineCloudProvider) addAsg(asg *AutoScalingGroup) {
	v.scalingGroups = append(v.scalingGroups, asg)
	v.volcengineManager.RegisterAsg(asg)
}

func buildScalingGroupFromSpec(manager VolcengineManager, spec string) (*AutoScalingGroup, error) {
	nodeGroupSpec, err := dynamic.SpecFromString(spec, true)
	if err != nil {
		return nil, fmt.Errorf("failed to parse node group spec: %v", err)
	}
	group, err := manager.GetAsgById(nodeGroupSpec.Name)
	if err != nil {
		klog.Errorf("scaling group %s not exists", nodeGroupSpec.Name)
		return nil, err
	}
	return &AutoScalingGroup{
		manager:           manager,
		asgId:             nodeGroupSpec.Name,
		minInstanceNumber: group.minInstanceNumber,
		maxInstanceNumber: group.maxInstanceNumber,
	}, nil
}

// BuildVolcengine builds CloudProvider implementation for Volcengine
func BuildVolcengine(opts config.AutoscalingOptions, do cloudprovider.NodeGroupDiscoveryOptions, rl *cloudprovider.ResourceLimiter) cloudprovider.CloudProvider {
	if opts.CloudConfig == "" {
		klog.Fatalf("The path to the cloud provider configuration file must be set via the --cloud-config command line parameter")
	}
	cloudConf, err := readConf(opts.CloudConfig)
	if err != nil {
		klog.Warningf("Failed to read cloud provider configuration: %v", err)
		cloudConf = &cloudConfig{}
	}

	if !cloudConf.validate() {
		klog.Fatalf("Failed to validate cloud provider configuration: %v", err)
	}

	manager, err := CreateVolcengineManager(cloudConf)
	if err != nil {
		klog.Fatalf("Failed to create volcengine manager: %v", err)
	}

	provider, err := buildVolcengineProvider(manager, do, rl)
	if err != nil {
		klog.Fatalf("Failed to create volcengine cloud provider: %v", err)
	}

	return provider
}

func buildVolcengineProvider(manager VolcengineManager, do cloudprovider.NodeGroupDiscoveryOptions, rl *cloudprovider.ResourceLimiter) (cloudprovider.CloudProvider, error) {
	if !do.StaticDiscoverySpecified() {
		return nil, fmt.Errorf("static discovery configuration must be provided for volcengine cloud provider")
	}

	provider := &volcengineCloudProvider{
		volcengineManager: manager,
		resourceLimiter:   rl,
	}

	for _, spec := range do.NodeGroupSpecs {
		if err := provider.addNodeGroup(spec); err != nil {
			klog.Warningf("Failed to add node group from spec %s: %v", spec, err)
			return nil, err
		}
	}

	return provider, nil
}

func readConf(confFile string) (*cloudConfig, error) {
	var conf io.ReadCloser
	conf, err := os.Open(confFile)
	if err != nil {
		return nil, err
	}
	defer conf.Close()

	var cloudConfig cloudConfig
	if err = gcfg.ReadInto(&cloudConfig, conf); err != nil {
		return nil, err
	}

	return &cloudConfig, nil
}

func ecsInstanceFromProviderId(providerId string) (string, error) {
	if !strings.HasPrefix(providerId, "volcengine://") {
		return "", fmt.Errorf("providerId %q doesn't match prefix %q", providerId, "volcengine://")
	}
	return strings.TrimPrefix(providerId, "volcengine://"), nil
}
