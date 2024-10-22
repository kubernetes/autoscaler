/*
Copyright 2016 The Kubernetes Authors.

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

package aws

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	klog "k8s.io/klog/v2"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
)

const (
	// GPULabel is the label added to nodes with GPU resource.
	GPULabel = "k8s.amazonaws.com/accelerator"
	// nodeNotPresentErr indicates no node with the given identifier present in AWS
	nodeNotPresentErr = "node is not present in aws"
)

var (
	availableGPUTypes = map[string]struct{}{
		"nvidia-tesla-k80":  {},
		"nvidia-tesla-p100": {},
		"nvidia-tesla-v100": {},
		"nvidia-tesla-t4":   {},
		"nvidia-tesla-a100": {},
		"nvidia-a10g":       {},
		"nvidia-l4":         {},
		"nvidia-l40s":       {},
	}
)

// awsCloudProvider implements CloudProvider interface.
type awsCloudProvider struct {
	awsManager      *AwsManager
	resourceLimiter *cloudprovider.ResourceLimiter
}

// BuildAwsCloudProvider builds CloudProvider implementation for AWS.
func BuildAwsCloudProvider(awsManager *AwsManager, resourceLimiter *cloudprovider.ResourceLimiter) (cloudprovider.CloudProvider, error) {
	aws := &awsCloudProvider{
		awsManager:      awsManager,
		resourceLimiter: resourceLimiter,
	}
	return aws, nil
}

// Cleanup stops the go routine that is handling the current view of the ASGs in the form of a cache
func (aws *awsCloudProvider) Cleanup() error {
	aws.awsManager.Cleanup()
	return nil
}

// Name returns name of the cloud provider.
func (aws *awsCloudProvider) Name() string {
	return cloudprovider.AwsProviderName
}

// GPULabel returns the label added to nodes with GPU resource.
func (aws *awsCloudProvider) GPULabel() string {
	return GPULabel
}

// GetAvailableGPUTypes return all available GPU types cloud provider supports
func (aws *awsCloudProvider) GetAvailableGPUTypes() map[string]struct{} {
	return availableGPUTypes
}

// GetNodeGpuConfig returns the label, type and resource name for the GPU added to node. If node doesn't have
// any GPUs, it returns nil.
func (aws *awsCloudProvider) GetNodeGpuConfig(node *apiv1.Node) *cloudprovider.GpuConfig {
	return gpu.GetNodeGPUFromCloudProvider(aws, node)
}

// NodeGroups returns all node groups configured for this cloud provider.
func (aws *awsCloudProvider) NodeGroups() []cloudprovider.NodeGroup {
	asgs := aws.awsManager.getAsgs()
	ngs := make([]cloudprovider.NodeGroup, 0, len(asgs))
	for _, asg := range asgs {
		ngs = append(ngs, &AwsNodeGroup{
			asg:        asg,
			awsManager: aws.awsManager,
		})
	}

	return ngs
}

// NodeGroupForNode returns the node group for the given node.
func (aws *awsCloudProvider) NodeGroupForNode(node *apiv1.Node) (cloudprovider.NodeGroup, error) {
	if len(node.Spec.ProviderID) == 0 {
		klog.Warningf("Node %v has no providerId", node.Name)
		return nil, nil
	}
	ref, err := AwsRefFromProviderId(node.Spec.ProviderID)
	if err != nil {
		return nil, err
	}
	asg := aws.awsManager.GetAsgForInstance(*ref)

	if asg == nil {
		return nil, nil
	}

	return &AwsNodeGroup{
		asg:        asg,
		awsManager: aws.awsManager,
	}, nil
}

// HasInstance returns whether a given node has a corresponding instance in this cloud provider
func (aws *awsCloudProvider) HasInstance(node *apiv1.Node) (bool, error) {
	// we haven't implemented a way to check if a fargate instance
	// exists in the cloud provider
	// returning 'true' because we are assuming the node exists in AWS
	// this is the default behavior if the check is unimplemented
	if strings.HasPrefix(node.GetName(), "fargate") {
		return true, cloudprovider.ErrNotImplemented
	}

	// avoid log spam for not autoscaled asgs:
	//   Nodes that belong to an asg that is not autoscaled will not be found in the asgCache below,
	//   so do not trigger warning spam by returning an error from being unable to find them.
	//   Annotation is not automated, but users that see the warning can add the annotation to avoid it.
	if node.Annotations != nil && node.Annotations["k8s.io/cluster-autoscaler-enabled"] == "false" {
		return false, nil
	}

	awsRef, err := AwsRefFromProviderId(node.Spec.ProviderID)
	if err != nil {
		return false, err
	}

	// we don't care about the status
	status, err := aws.awsManager.asgCache.InstanceStatus(*awsRef)
	if status != nil {
		return true, nil
	}

	return false, fmt.Errorf("%s: %v", nodeNotPresentErr, err)
}

// Pricing returns pricing model for this cloud provider or error if not available.
func (aws *awsCloudProvider) Pricing() (cloudprovider.PricingModel, errors.AutoscalerError) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetAvailableMachineTypes get all machine types that can be requested from the cloud provider.
func (aws *awsCloudProvider) GetAvailableMachineTypes() ([]string, error) {
	return []string{}, nil
}

// NewNodeGroup builds a theoretical node group based on the node definition provided. The node group is not automatically
// created on the cloud provider side. The node group is not returned by NodeGroups() until it is created.
func (aws *awsCloudProvider) NewNodeGroup(machineType string, labels map[string]string, systemLabels map[string]string,
	taints []apiv1.Taint, extraResources map[string]resource.Quantity) (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetResourceLimiter returns struct containing limits (max, min) for resources (cores, memory etc.).
func (aws *awsCloudProvider) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	return aws.resourceLimiter, nil
}

// Refresh is called before every main loop and can be used to dynamically update cloud provider state.
// In particular the list of node groups returned by NodeGroups can change as a result of CloudProvider.Refresh().
func (aws *awsCloudProvider) Refresh() error {
	return aws.awsManager.Refresh()
}

// AwsRef contains a reference to some entity in AWS world.
type AwsRef struct {
	Name string
}

// AwsInstanceRef contains a reference to an instance in the AWS world.
type AwsInstanceRef struct {
	ProviderID string
	Name       string
}

var validAwsRefIdRegex = regexp.MustCompile(fmt.Sprintf(`^aws\:\/\/\/[-0-9a-z]*\/[-0-9a-z]*(\/[-0-9a-z\.]*)?$|aws\:\/\/\/[-0-9a-z]*\/%s.*$`, placeholderInstanceNamePrefix))

// AwsRefFromProviderId creates AwsInstanceRef object from provider id which
// must be in format: aws:///zone/name
func AwsRefFromProviderId(id string) (*AwsInstanceRef, error) {
	if validAwsRefIdRegex.FindStringSubmatch(id) == nil {
		return nil, fmt.Errorf("wrong id: expected format aws:///<zone>/<name>, got %v", id)
	}
	splitted := strings.Split(id[7:], "/")
	return &AwsInstanceRef{
		ProviderID: id,
		Name:       splitted[1],
	}, nil
}

// AwsNodeGroup implements NodeGroup interface.
type AwsNodeGroup struct {
	awsManager *AwsManager
	asg        *asg
}

// MaxSize returns maximum size of the node group.
func (ng *AwsNodeGroup) MaxSize() int {
	return ng.asg.maxSize
}

// MinSize returns minimum size of the node group.
func (ng *AwsNodeGroup) MinSize() int {
	return ng.asg.minSize
}

// TargetSize returns the current TARGET size of the node group. It is possible that the
// number is different from the number of nodes registered in Kubernetes.
func (ng *AwsNodeGroup) TargetSize() (int, error) {
	return ng.asg.curSize, nil
}

// Exist checks if the node group really exists on the cloud provider side. Allows to tell the
// theoretical node group from the real one.
func (ng *AwsNodeGroup) Exist() bool {
	return true
}

// Create creates the node group on the cloud provider side.
func (ng *AwsNodeGroup) Create() (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrAlreadyExist
}

// Autoprovisioned returns true if the node group is autoprovisioned.
func (ng *AwsNodeGroup) Autoprovisioned() bool {
	return false
}

// Delete deletes the node group on the cloud provider side.
// This will be executed only for autoprovisioned node groups, once their size drops to 0.
func (ng *AwsNodeGroup) Delete() error {
	return cloudprovider.ErrNotImplemented
}

// GetOptions returns NodeGroupAutoscalingOptions that should be used for this particular
// NodeGroup. Returning a nil will result in using default options.
func (ng *AwsNodeGroup) GetOptions(defaults config.NodeGroupAutoscalingOptions) (*config.NodeGroupAutoscalingOptions, error) {
	if ng.asg == nil || ng.asg.Tags == nil || len(ng.asg.Tags) == 0 {
		return &defaults, nil
	}
	return ng.awsManager.GetAsgOptions(*ng.asg, defaults), nil
}

// IncreaseSize increases Asg size
func (ng *AwsNodeGroup) IncreaseSize(delta int) error {
	if delta <= 0 {
		return fmt.Errorf("size increase must be positive")
	}
	size := ng.asg.curSize
	if size+delta > ng.asg.maxSize {
		return fmt.Errorf("size increase too large - desired:%d max:%d", size+delta, ng.asg.maxSize)
	}
	return ng.awsManager.SetAsgSize(ng.asg, size+delta)
}

// AtomicIncreaseSize is not implemented.
func (ng *AwsNodeGroup) AtomicIncreaseSize(delta int) error {
	return cloudprovider.ErrNotImplemented
}

// DecreaseTargetSize decreases the target size of the node group. This function
// doesn't permit to delete any existing node and can be used only to reduce the
// request for new nodes that have not been yet fulfilled. Delta should be negative.
// It is assumed that cloud provider will not delete the existing nodes if the size
// when there is an option to just decrease the target.
func (ng *AwsNodeGroup) DecreaseTargetSize(delta int) error {
	if delta >= 0 {
		return fmt.Errorf("size decrease size must be negative")
	}

	size := ng.asg.curSize
	nodes, err := ng.awsManager.GetAsgNodes(ng.asg.AwsRef)
	if err != nil {
		return err
	}
	if int(size)+delta < len(nodes) {
		return fmt.Errorf("attempt to delete existing nodes targetSize:%d delta:%d existingNodes: %d",
			size, delta, len(nodes))
	}
	return ng.awsManager.SetAsgSize(ng.asg, size+delta)
}

// Belongs returns true if the given node belongs to the NodeGroup.
func (ng *AwsNodeGroup) Belongs(node *apiv1.Node) (bool, error) {
	ref, err := AwsRefFromProviderId(node.Spec.ProviderID)
	if err != nil {
		return false, err
	}
	targetAsg := ng.awsManager.GetAsgForInstance(*ref)
	if targetAsg == nil {
		return false, fmt.Errorf("%s doesn't belong to a known asg", node.Name)
	}
	if targetAsg.AwsRef != ng.asg.AwsRef {
		return false, nil
	}
	return true, nil
}

// DeleteNodes deletes the nodes from the group.
func (ng *AwsNodeGroup) DeleteNodes(nodes []*apiv1.Node) error {
	size := ng.asg.curSize
	if int(size) <= ng.MinSize() {
		return fmt.Errorf("min size reached, nodes will not be deleted")
	}
	refs := make([]*AwsInstanceRef, 0, len(nodes))
	for _, node := range nodes {
		belongs, err := ng.Belongs(node)
		if err != nil {
			return err
		}
		if !belongs {
			return fmt.Errorf("%s belongs to a different asg than %s", node.Name, ng.Id())
		}
		awsref, err := AwsRefFromProviderId(node.Spec.ProviderID)
		if err != nil {
			return err
		}
		refs = append(refs, awsref)
	}
	return ng.awsManager.DeleteInstances(refs)
}

// Id returns asg id.
func (ng *AwsNodeGroup) Id() string {
	return ng.asg.Name
}

// Debug returns a debug string for the Asg.
func (ng *AwsNodeGroup) Debug() string {
	return fmt.Sprintf("%s (%d:%d)", ng.Id(), ng.MinSize(), ng.MaxSize())
}

// Nodes returns a list of all nodes that belong to this node group.
func (ng *AwsNodeGroup) Nodes() ([]cloudprovider.Instance, error) {
	asgNodes, err := ng.awsManager.GetAsgNodes(ng.asg.AwsRef)
	if err != nil {
		return nil, err
	}

	instances := make([]cloudprovider.Instance, len(asgNodes))

	for i, asgNode := range asgNodes {
		var status *cloudprovider.InstanceStatus
		instanceStatusString, err := ng.awsManager.GetInstanceStatus(asgNode)
		if err != nil {
			klog.V(4).Infof("Could not get instance status, continuing anyways: %v", err)
		} else if instanceStatusString != nil && *instanceStatusString == placeholderUnfulfillableStatus {
			status = &cloudprovider.InstanceStatus{
				State: cloudprovider.InstanceCreating,
				ErrorInfo: &cloudprovider.InstanceErrorInfo{
					ErrorClass:   cloudprovider.OutOfResourcesErrorClass,
					ErrorCode:    placeholderUnfulfillableStatus,
					ErrorMessage: "AWS cannot provision any more instances for this node group",
				},
			}
		}
		instances[i] = cloudprovider.Instance{
			Id:     asgNode.ProviderID,
			Status: status,
		}
	}
	return instances, nil
}

// TemplateNodeInfo returns a node template for this node group.
func (ng *AwsNodeGroup) TemplateNodeInfo() (*schedulerframework.NodeInfo, error) {
	template, err := ng.awsManager.getAsgTemplate(ng.asg)
	if err != nil {
		return nil, err
	}

	node, err := ng.awsManager.buildNodeFromTemplate(ng.asg, template)
	if err != nil {
		return nil, err
	}

	nodeInfo := schedulerframework.NewNodeInfo(cloudprovider.BuildKubeProxy(ng.asg.Name))
	nodeInfo.SetNode(node)
	return nodeInfo, nil
}

// BuildAWS builds AWS cloud provider, manager etc.
func BuildAWS(opts config.AutoscalingOptions, do cloudprovider.NodeGroupDiscoveryOptions, rl *cloudprovider.ResourceLimiter) cloudprovider.CloudProvider {
	var cfg io.ReadCloser
	if opts.CloudConfig != "" {
		var err error
		cfg, err = os.Open(opts.CloudConfig)
		if err != nil {
			klog.Fatalf("Couldn't open cloud provider configuration %s: %#v", opts.CloudConfig, err)
		}
		defer cfg.Close()
	}

	sdkProvider, err := createAWSSDKProvider(cfg)
	if err != nil {
		klog.Fatalf("Failed to create AWS SDK Provider: %v", err)
	}

	// Generate EC2 list
	instanceTypes, lastUpdateTime := GetStaticEC2InstanceTypes()
	if opts.AWSUseStaticInstanceList {
		klog.Warningf("Using static EC2 Instance Types, this list could be outdated. Last update time: %s", lastUpdateTime)
	} else {
		generatedInstanceTypes, err := GenerateEC2InstanceTypes(sdkProvider.session)
		if err != nil {
			klog.Errorf("Failed to generate AWS EC2 Instance Types: %v, falling back to static list with last update time: %s", err, lastUpdateTime)
		}
		if generatedInstanceTypes == nil {
			generatedInstanceTypes = map[string]*InstanceType{}
		}
		// fallback on the static list if we miss any instance types in the generated output
		// credits to: https://github.com/lyft/cni-ipvlan-vpc-k8s/pull/80
		for k, v := range instanceTypes {
			_, ok := generatedInstanceTypes[k]
			if ok {
				continue
			}
			klog.Infof("Using static instance type %s", k)
			generatedInstanceTypes[k] = v
		}
		instanceTypes = generatedInstanceTypes

		keys := make([]string, 0, len(instanceTypes))
		for key := range instanceTypes {
			keys = append(keys, key)
		}

		klog.Infof("Successfully load %d EC2 Instance Types %s", len(keys), keys)
	}

	manager, err := CreateAwsManager(sdkProvider, do, instanceTypes)
	if err != nil {
		klog.Fatalf("Failed to create AWS Manager: %v", err)
	}

	provider, err := BuildAwsCloudProvider(manager, rl)
	if err != nil {
		klog.Fatalf("Failed to create AWS cloud provider: %v", err)
	}
	RegisterMetrics()
	return provider
}
