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

package alicloud

import (
	"fmt"
	"os"
	"strings"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/config/dynamic"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	klog "k8s.io/klog/v2"
)

const (
	// GPULabel is the label added to nodes with GPU resource.
	GPULabel = "aliyun.accelerator/nvidia_name"
)

var (
	availableGPUTypes = map[string]struct{}{
		"Tesla-P4": {},
		"M40":      {},
		"P100":     {},
		"V100":     {},
	}
)

type aliCloudProvider struct {
	manager         *AliCloudManager
	asgs            []*Asg
	resourceLimiter *cloudprovider.ResourceLimiter
}

// BuildAliCloudProvider builds CloudProvider implementation for AliCloud.
func BuildAliCloudProvider(manager *AliCloudManager, discoveryOpts cloudprovider.NodeGroupDiscoveryOptions, resourceLimiter *cloudprovider.ResourceLimiter) (cloudprovider.CloudProvider, error) {
	// TODO add discoveryOpts parameters check.
	if discoveryOpts.StaticDiscoverySpecified() {
		return buildStaticallyDiscoveringProvider(manager, discoveryOpts.NodeGroupSpecs, resourceLimiter)
	}
	if discoveryOpts.AutoDiscoverySpecified() {
		return nil, fmt.Errorf("only support static discovery scaling group in alicloud for now")
	}
	return nil, fmt.Errorf("failed to build alicloud provider: node group specs must be specified")
}

func buildStaticallyDiscoveringProvider(manager *AliCloudManager, specs []string, resourceLimiter *cloudprovider.ResourceLimiter) (*aliCloudProvider, error) {
	acp := &aliCloudProvider{
		manager:         manager,
		asgs:            make([]*Asg, 0),
		resourceLimiter: resourceLimiter,
	}
	for _, spec := range specs {
		if err := acp.addNodeGroup(spec); err != nil {
			klog.Warningf("failed to add node group to alicloud provider with spec: %s", spec)
			return nil, err
		}
	}
	return acp, nil
}

// add node group defined in string spec. Format:
// minNodes:maxNodes:asgName
func (ali *aliCloudProvider) addNodeGroup(spec string) error {
	asg, err := buildAsgFromSpec(spec, ali.manager)
	if err != nil {
		klog.Errorf("failed to build ASG from spec,because of %s", err.Error())
		return err
	}
	ali.addAsg(asg)
	return nil
}

// add and register an asg to this cloud provider
func (ali *aliCloudProvider) addAsg(asg *Asg) {
	ali.asgs = append(ali.asgs, asg)
	ali.manager.RegisterAsg(asg)
}

func (ali *aliCloudProvider) Name() string {
	return cloudprovider.AlicloudProviderName
}

// GPULabel returns the label added to nodes with GPU resource.
func (ali *aliCloudProvider) GPULabel() string {
	return GPULabel
}

// GetAvailableGPUTypes return all available GPU types cloud provider supports
func (ali *aliCloudProvider) GetAvailableGPUTypes() map[string]struct{} {
	return availableGPUTypes
}

func (ali *aliCloudProvider) NodeGroups() []cloudprovider.NodeGroup {
	result := make([]cloudprovider.NodeGroup, 0, len(ali.asgs))
	for _, asg := range ali.asgs {
		result = append(result, asg)
	}
	return result
}

// NodeGroupForNode returns the node group for the given node.
func (ali *aliCloudProvider) NodeGroupForNode(node *apiv1.Node) (cloudprovider.NodeGroup, error) {
	instanceId, err := ecsInstanceIdFromProviderId(node.Spec.ProviderID)
	if err != nil {
		klog.Errorf("failed to get instance Id from provider Id:%s,because of %s", node.Spec.ProviderID, err.Error())
		return nil, err
	}
	return ali.manager.GetAsgForInstance(instanceId)
}

// Pricing returns pricing model for this cloud provider or error if not available.
func (ali *aliCloudProvider) Pricing() (cloudprovider.PricingModel, errors.AutoscalerError) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetAvailableMachineTypes get all machine types that can be requested from the cloud provider.
func (ali *aliCloudProvider) GetAvailableMachineTypes() ([]string, error) {
	return []string{}, nil
}

// NewNodeGroup builds a theoretical node group based on the node definition provided. The node group is not automatically
// created on the cloud provider side. The node group is not returned by NodeGroups() until it is created.
func (ali *aliCloudProvider) NewNodeGroup(machineType string, labels map[string]string, systemLabels map[string]string, taints []apiv1.Taint, extraResources map[string]resource.Quantity) (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetResourceLimiter returns struct containing limits (max, min) for resources (cores, memory etc.).
func (ali *aliCloudProvider) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	return ali.resourceLimiter, nil
}

// Refresh is called before every main loop and can be used to dynamically update cloud provider state.
// In particular the list of node groups returned by NodeGroups can change as a result of CloudProvider.Refresh().
func (ali *aliCloudProvider) Refresh() error {
	return nil
}

// Cleanup stops the go routine that is handling the current view of the ASGs in the form of a cache
func (ali *aliCloudProvider) Cleanup() error {
	return nil
}

// AliRef contains a reference to ECS instance or .
type AliRef struct {
	ID     string
	Region string
}

// ECSInstanceIdFromProviderId must be in format: `REGION.INSTANCE_ID`
func ecsInstanceIdFromProviderId(id string) (string, error) {
	parts := strings.Split(id, ".")
	if len(parts) < 2 {
		return "", fmt.Errorf("AliCloud: unexpected ProviderID format, providerID=%s", id)
	}
	return parts[1], nil
}

func buildAsgFromSpec(value string, manager *AliCloudManager) (*Asg, error) {
	spec, err := dynamic.SpecFromString(value, true)

	if err != nil {
		return nil, fmt.Errorf("failed to parse node group spec: %v", err)
	}

	// check auto scaling group is exists or not
	_, err = manager.aService.getScalingGroupByID(spec.Name)
	if err != nil {
		klog.Errorf("your scaling group: %s does not exist", spec.Name)
		return nil, err
	}

	asg := buildAsg(manager, spec.MinSize, spec.MaxSize, spec.Name, manager.cfg.getRegion())

	return asg, nil
}

func buildAsg(manager *AliCloudManager, minSize int, maxSize int, id string, regionId string) *Asg {
	return &Asg{
		manager:  manager,
		minSize:  minSize,
		maxSize:  maxSize,
		regionId: regionId,
		id:       id,
	}
}

// BuildAlicloud returns alicloud provider
func BuildAlicloud(opts config.AutoscalingOptions, do cloudprovider.NodeGroupDiscoveryOptions, rl *cloudprovider.ResourceLimiter) cloudprovider.CloudProvider {
	var aliManager *AliCloudManager
	var aliError error
	if opts.CloudConfig != "" {
		config, fileErr := os.Open(opts.CloudConfig)
		if fileErr != nil {
			klog.Fatalf("Couldn't open cloud provider configuration %s: %#v", opts.CloudConfig, fileErr)
		}
		defer config.Close()
		aliManager, aliError = CreateAliCloudManager(config)
	} else {
		aliManager, aliError = CreateAliCloudManager(nil)
	}
	if aliError != nil {
		klog.Fatalf("Failed to create Alicloud Manager: %v", aliError)
	}
	cloudProvider, err := BuildAliCloudProvider(aliManager, do, rl)
	if err != nil {
		klog.Fatalf("Failed to create Alicloud cloud provider: %v", err)
	}
	return cloudProvider
}
