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
	"regexp"
	"strings"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/price"
	"k8s.io/autoscaler/cluster-autoscaler/config/dynamic"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/kubernetes/plugin/pkg/scheduler/schedulercache"
)

// awsCloudProvider implements CloudProvider interface.
type awsCloudProvider struct {
	awsManager      *AwsManager
	asgs            []*Asg
	priceModel      cloudprovider.PricingModel
	resourceLimiter *cloudprovider.ResourceLimiter
}

// BuildAwsCloudProvider builds CloudProvider implementation for AWS.
func BuildAwsCloudProvider(awsManager *AwsManager, discoveryOpts cloudprovider.NodeGroupDiscoveryOptions, resourceLimiter *cloudprovider.ResourceLimiter, descriptor price.ShapeDescriptor) (cloudprovider.CloudProvider, error) {
	if err := discoveryOpts.Validate(); err != nil {
		return nil, fmt.Errorf("Failed to build an aws cloud provider: %v", err)
	}
	if discoveryOpts.StaticDiscoverySpecified() {
		return buildStaticallyDiscoveringProvider(awsManager, discoveryOpts.NodeGroupSpecs, resourceLimiter, descriptor)
	}
	if discoveryOpts.AutoDiscoverySpecified() {
		return buildAutoDiscoveringProvider(awsManager, discoveryOpts.NodeGroupAutoDiscoverySpec, resourceLimiter, descriptor)
	}
	return nil, fmt.Errorf("Failed to build an aws cloud provider: Either node group specs or node group auto discovery spec must be specified")
}

func buildAutoDiscoveringProvider(awsManager *AwsManager, spec string, resourceLimiter *cloudprovider.ResourceLimiter, descriptor price.ShapeDescriptor) (*awsCloudProvider, error) {
	tokens := strings.Split(spec, ":")
	if len(tokens) != 2 {
		return nil, fmt.Errorf("Invalid node group auto discovery spec specified via --node-group-auto-discovery: %s", spec)
	}
	discoverer := tokens[0]
	if discoverer != "asg" {
		return nil, fmt.Errorf("Unsupported discoverer specified: %s", discoverer)
	}
	param := tokens[1]
	paramTokens := strings.Split(param, "=")
	parameterKey := paramTokens[0]
	if parameterKey != "tag" {
		return nil, fmt.Errorf("Unsupported parameter key \"%s\" is specified for discoverer \"%s\". The only supported key is \"tag\"", parameterKey, discoverer)
	}
	tag := paramTokens[1]
	if tag == "" {
		return nil, fmt.Errorf("Invalid ASG tag for auto discovery specified: ASG tag must not be empty")
	}
	// Use the k8s cluster name tag to only discover asgs of the cluster denoted by clusterName
	// See https://github.com/kubernetes/kubernetes/blob/9ef85a7/pkg/cloudprovider/providers/aws/tags.go#L30-L34
	// for more information about the tag
	tags := strings.Split(tag, ",")
	asgs, err := awsManager.getAutoscalingGroupsByTags(tags)
	if err != nil {
		return nil, fmt.Errorf("Failed to get ASGs: %v", err)
	}

	aws := &awsCloudProvider{
		awsManager:      awsManager,
		asgs:            make([]*Asg, 0),
		resourceLimiter: resourceLimiter,
		priceModel:      NewPriceModel(awsManager.asgs, descriptor),
	}
	for _, asg := range asgs {
		aws.addAsg(buildAsg(aws.awsManager, int(*asg.MinSize), int(*asg.MaxSize), *asg.AutoScalingGroupName))
	}
	return aws, nil
}

func buildStaticallyDiscoveringProvider(awsManager *AwsManager, specs []string, resourceLimiter *cloudprovider.ResourceLimiter, descriptor price.ShapeDescriptor) (*awsCloudProvider, error) {
	aws := &awsCloudProvider{
		awsManager:      awsManager,
		asgs:            make([]*Asg, 0),
		resourceLimiter: resourceLimiter,
		priceModel:      NewPriceModel(awsManager.asgs, descriptor),
	}
	for _, spec := range specs {
		if err := aws.addNodeGroup(spec); err != nil {
			return nil, err
		}
	}
	return aws, nil
}

// Cleanup stops the go routine that is handling the current view of the ASGs in the form of a cache
func (aws *awsCloudProvider) Cleanup() error {
	aws.awsManager.Cleanup()
	return nil
}

// addNodeGroup adds node group defined in string spec. Format:
// minNodes:maxNodes:asgName
func (aws *awsCloudProvider) addNodeGroup(spec string) error {
	asg, err := buildAsgFromSpec(spec, aws.awsManager)
	if err != nil {
		return err
	}
	aws.addAsg(asg)
	return nil
}

// addAsg adds and registers an asg to this cloud provider
func (aws *awsCloudProvider) addAsg(asg *Asg) {
	aws.asgs = append(aws.asgs, asg)
	aws.awsManager.RegisterAsg(asg)
}

// Name returns name of the cloud provider.
func (aws *awsCloudProvider) Name() string {
	return "aws"
}

// NodeGroups returns all node groups configured for this cloud provider.
func (aws *awsCloudProvider) NodeGroups() []cloudprovider.NodeGroup {
	result := make([]cloudprovider.NodeGroup, 0, len(aws.asgs))
	for _, asg := range aws.asgs {
		result = append(result, asg)
	}
	return result
}

// NodeGroupForNode returns the node group for the given node.
func (aws *awsCloudProvider) NodeGroupForNode(node *apiv1.Node) (cloudprovider.NodeGroup, error) {
	ref, err := AwsRefFromProviderId(node.Spec.ProviderID)
	if err != nil {
		return nil, err
	}
	asg, err := aws.awsManager.GetAsgForInstance(ref)
	return asg, err
}

// Pricing returns pricing model for this cloud provider or error if not available.
func (aws *awsCloudProvider) Pricing() (cloudprovider.PricingModel, errors.AutoscalerError) {
	// aws.awsManager.asgs.FindForInstance()
	return aws.priceModel, nil
}

// GetAvailableMachineTypes get all machine types that can be requested from the cloud provider.
func (aws *awsCloudProvider) GetAvailableMachineTypes() ([]string, error) {
	return []string{}, nil
}

// NewNodeGroup builds a theoretical node group based on the node definition provided. The node group is not automatically
// created on the cloud provider side. The node group is not returned by NodeGroups() until it is created.
func (aws *awsCloudProvider) NewNodeGroup(machineType string, labels map[string]string, extraResources map[string]resource.Quantity) (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetResourceLimiter returns struct containing limits (max, min) for resources (cores, memory etc.).
func (aws *awsCloudProvider) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	return aws.resourceLimiter, nil
}

// Refresh is called before every main loop and can be used to dynamically update cloud provider state.
// In particular the list of node groups returned by NodeGroups can change as a result of CloudProvider.Refresh().
func (aws *awsCloudProvider) Refresh() error {
	return nil
}

// AwsRef contains a reference to some entity in AWS/GKE world.
type AwsRef struct {
	Name string
}

var validAwsRefIdRegex = regexp.MustCompile(`^aws\:\/\/\/[-0-9a-z]*\/[-0-9a-z]*$`)

// AwsRefFromProviderId creates InstanceConfig object from provider id which
// must be in format: aws:///zone/name
func AwsRefFromProviderId(id string) (*AwsRef, error) {
	if validAwsRefIdRegex.FindStringSubmatch(id) == nil {
		return nil, fmt.Errorf("Wrong id: expected format aws:///<zone>/<name>, got %v", id)
	}
	splitted := strings.Split(id[7:], "/")
	return &AwsRef{
		Name: splitted[1],
	}, nil
}

// Asg implements NodeGroup interface.
type Asg struct {
	AwsRef

	awsManager *AwsManager

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

// TargetSize returns the current TARGET size of the node group. It is possible that the
// number is different from the number of nodes registered in Kubernetes.
func (asg *Asg) TargetSize() (int, error) {
	size, err := asg.awsManager.GetAsgSize(asg)
	return int(size), err
}

// Exist checks if the node group really exists on the cloud provider side. Allows to tell the
// theoretical node group from the real one.
func (asg *Asg) Exist() bool {
	return true
}

// Create creates the node group on the cloud provider side.
func (asg *Asg) Create() error {
	return cloudprovider.ErrAlreadyExist
}

// Autoprovisioned returns true if the node group is autoprovisioned.
func (asg *Asg) Autoprovisioned() bool {
	return false
}

// Delete deletes the node group on the cloud provider side.
// This will be executed only for autoprovisioned node groups, once their size drops to 0.
func (asg *Asg) Delete() error {
	return cloudprovider.ErrNotImplemented
}

// IncreaseSize increases Asg size
func (asg *Asg) IncreaseSize(delta int) error {
	if delta <= 0 {
		return fmt.Errorf("size increase must be positive")
	}
	size, err := asg.awsManager.GetAsgSize(asg)
	if err != nil {
		return err
	}
	if int(size)+delta > asg.MaxSize() {
		return fmt.Errorf("size increase too large - desired:%d max:%d", int(size)+delta, asg.MaxSize())
	}
	return asg.awsManager.SetAsgSize(asg, size+int64(delta))
}

// DecreaseTargetSize decreases the target size of the node group. This function
// doesn't permit to delete any existing node and can be used only to reduce the
// request for new nodes that have not been yet fulfilled. Delta should be negative.
// It is assumed that cloud provider will not delete the existing nodes if the size
// when there is an option to just decrease the target.
func (asg *Asg) DecreaseTargetSize(delta int) error {
	if delta >= 0 {
		return fmt.Errorf("size decrease size must be negative")
	}
	size, err := asg.awsManager.GetAsgSize(asg)
	if err != nil {
		return err
	}
	nodes, err := asg.awsManager.GetAsgNodes(asg)
	if err != nil {
		return err
	}
	if int(size)+delta < len(nodes) {
		return fmt.Errorf("attempt to delete existing nodes targetSize:%d delta:%d existingNodes: %d",
			size, delta, len(nodes))
	}
	return asg.awsManager.SetAsgSize(asg, size+int64(delta))
}

// Belongs returns true if the given node belongs to the NodeGroup.
func (asg *Asg) Belongs(node *apiv1.Node) (bool, error) {
	ref, err := AwsRefFromProviderId(node.Spec.ProviderID)
	if err != nil {
		return false, err
	}
	targetAsg, err := asg.awsManager.GetAsgForInstance(ref)
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

// DeleteNodes deletes the nodes from the group.
func (asg *Asg) DeleteNodes(nodes []*apiv1.Node) error {
	size, err := asg.awsManager.GetAsgSize(asg)
	if err != nil {
		return err
	}
	if int(size) <= asg.MinSize() {
		return fmt.Errorf("min size reached, nodes will not be deleted")
	}
	refs := make([]*AwsRef, 0, len(nodes))
	for _, node := range nodes {
		belongs, err := asg.Belongs(node)
		if err != nil {
			return err
		}
		if belongs != true {
			return fmt.Errorf("%s belongs to a different asg than %s", node.Name, asg.Id())
		}
		awsref, err := AwsRefFromProviderId(node.Spec.ProviderID)
		if err != nil {
			return err
		}
		refs = append(refs, awsref)
	}
	return asg.awsManager.DeleteInstances(refs)
}

// Id returns asg id.
func (asg *Asg) Id() string {
	return asg.Name
}

// Debug returns a debug string for the Asg.
func (asg *Asg) Debug() string {
	return fmt.Sprintf("%s (%d:%d)", asg.Id(), asg.MinSize(), asg.MaxSize())
}

// Nodes returns a list of all nodes that belong to this node group.
func (asg *Asg) Nodes() ([]string, error) {
	return asg.awsManager.GetAsgNodes(asg)
}

// TemplateNodeInfo returns a node template for this node group.
func (asg *Asg) TemplateNodeInfo() (*schedulercache.NodeInfo, error) {
	template, err := asg.awsManager.getAsgTemplate(asg.Name)
	if err != nil {
		return nil, err
	}

	node, err := asg.awsManager.buildNodeFromTemplate(asg, template)
	if err != nil {
		return nil, err
	}

	nodeInfo := schedulercache.NewNodeInfo(cloudprovider.BuildKubeProxy(asg.Name))
	nodeInfo.SetNode(node)
	return nodeInfo, nil
}

func buildAsgFromSpec(value string, awsManager *AwsManager) (*Asg, error) {
	spec, err := dynamic.SpecFromString(value, true)

	if err != nil {
		return nil, fmt.Errorf("failed to parse node group spec: %v", err)
	}

	asg := buildAsg(awsManager, spec.MinSize, spec.MaxSize, spec.Name)

	return asg, nil
}

func buildAsg(awsManager *AwsManager, minSize int, maxSize int, name string) *Asg {
	return &Asg{
		awsManager: awsManager,
		minSize:    minSize,
		maxSize:    maxSize,
		AwsRef: AwsRef{
			Name: name,
		},
	}
}
