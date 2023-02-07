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

//go:generate go run ec2_instance_types/gen.go -region $AWS_REGION

package aws

import (
	"errors"
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go/aws"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go/service/autoscaling"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go/service/ec2"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go/service/eks"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
)

const (
	operationWaitTimeout    = 5 * time.Second
	operationPollInterval   = 100 * time.Millisecond
	maxRecordsReturnedByAPI = 100
	maxAsgNamesPerDescribe  = 100
	refreshInterval         = 1 * time.Minute
	autoDiscovererTypeASG   = "asg"
	asgAutoDiscovererKeyTag = "tag"
	optionsTagsPrefix       = "k8s.io/cluster-autoscaler/node-template/autoscaling-options/"
)

// AwsManager is handles aws communication and data caching.
type AwsManager struct {
	awsService            awsWrapper
	asgCache              *asgCache
	lastRefresh           time.Time
	instanceTypes         map[string]*InstanceType
	managedNodegroupCache *managedNodegroupCache
}

type asgTemplate struct {
	InstanceType *InstanceType
	Region       string
	Zone         string
	Tags         []*autoscaling.TagDescription
}

// createAwsManagerInternal allows for custom objects to be passed in by tests
func createAWSManagerInternal(
	awsSDKProvider *awsSDKProvider,
	discoveryOpts cloudprovider.NodeGroupDiscoveryOptions,
	awsService *awsWrapper,
	instanceTypes map[string]*InstanceType,
) (*AwsManager, error) {
	if awsService == nil {
		sess := awsSDKProvider.session
		awsService = &awsWrapper{autoscaling.New(sess), ec2.New(sess), eks.New(sess)}
	}

	specs, err := parseASGAutoDiscoverySpecs(discoveryOpts)
	if err != nil {
		return nil, err
	}

	cache, err := newASGCache(awsService, discoveryOpts.NodeGroupSpecs, specs)
	if err != nil {
		return nil, err
	}

	mngCache := newManagedNodeGroupCache(awsService)

	manager := &AwsManager{
		awsService:            *awsService,
		asgCache:              cache,
		instanceTypes:         instanceTypes,
		managedNodegroupCache: mngCache,
	}

	if err := manager.forceRefresh(); err != nil {
		return nil, err
	}

	return manager, nil
}

// CreateAwsManager constructs awsManager object.
func CreateAwsManager(awsSDKProvider *awsSDKProvider, discoveryOpts cloudprovider.NodeGroupDiscoveryOptions, instanceTypes map[string]*InstanceType) (*AwsManager, error) {
	return createAWSManagerInternal(awsSDKProvider, discoveryOpts, nil, instanceTypes)
}

// Refresh is called before every main loop and can be used to dynamically update cloud provider state.
// In particular the list of node groups returned by NodeGroups can change as a result of CloudProvider.Refresh().
func (m *AwsManager) Refresh() error {
	if m.lastRefresh.Add(refreshInterval).After(time.Now()) {
		return nil
	}
	return m.forceRefresh()
}

func (m *AwsManager) forceRefresh() error {
	if err := m.asgCache.regenerate(); err != nil {
		klog.Errorf("Failed to regenerate ASG cache: %v", err)
		return err
	}
	m.lastRefresh = time.Now()
	klog.V(2).Infof("Refreshed ASG list, next refresh after %v", m.lastRefresh.Add(refreshInterval))
	return nil
}

// GetAsgForInstance returns AsgConfig of the given Instance
func (m *AwsManager) GetAsgForInstance(instance AwsInstanceRef) *asg {
	return m.asgCache.FindForInstance(instance)
}

// Cleanup the ASG cache.
func (m *AwsManager) Cleanup() {
	m.asgCache.Cleanup()
}

func (m *AwsManager) getAsgs() map[AwsRef]*asg {
	return m.asgCache.Get()
}

func (m *AwsManager) getAutoscalingOptions(ref AwsRef) map[string]string {
	return m.asgCache.GetAutoscalingOptions(ref)
}

// SetAsgSize sets ASG size.
func (m *AwsManager) SetAsgSize(asg *asg, size int) error {
	return m.asgCache.SetAsgSize(asg, size)
}

// DeleteInstances deletes the given instances. All instances must be controlled by the same ASG.
func (m *AwsManager) DeleteInstances(instances []*AwsInstanceRef) error {
	if err := m.asgCache.DeleteInstances(instances); err != nil {
		return err
	}
	klog.V(2).Infof("DeleteInstances was called: scheduling an ASG list refresh for next main loop evaluation")
	m.lastRefresh = time.Now().Add(-refreshInterval)
	return nil
}

// GetAsgNodes returns Asg nodes.
func (m *AwsManager) GetAsgNodes(ref AwsRef) ([]AwsInstanceRef, error) {
	return m.asgCache.InstancesByAsg(ref)
}

// GetInstanceStatus returns the status of ASG nodes
func (m *AwsManager) GetInstanceStatus(ref AwsInstanceRef) (*string, error) {
	return m.asgCache.InstanceStatus(ref)
}

func (m *AwsManager) getAsgTemplate(asg *asg) (*asgTemplate, error) {
	if len(asg.AvailabilityZones) < 1 {
		return nil, fmt.Errorf("unable to get first AvailabilityZone for ASG %q", asg.Name)
	}

	az := asg.AvailabilityZones[0]
	region := az[0 : len(az)-1]

	if len(asg.AvailabilityZones) > 1 {
		klog.V(4).Infof("Found multiple availability zones for ASG %q; using %s for %s label\n", asg.Name, az, apiv1.LabelZoneFailureDomain)
	}

	instanceTypeName, err := getInstanceTypeForAsg(m.asgCache, asg)
	if err != nil {
		return nil, err
	}

	if t, ok := m.instanceTypes[instanceTypeName]; ok {
		return &asgTemplate{
			InstanceType: t,
			Region:       region,
			Zone:         az,
			Tags:         asg.Tags,
		}, nil
	}

	return nil, fmt.Errorf("ASG %q uses the unknown EC2 instance type %q", asg.Name, instanceTypeName)
}

// GetAsgOptions parse options extracted from ASG tags and merges them with provided defaults
func (m *AwsManager) GetAsgOptions(asg asg, defaults config.NodeGroupAutoscalingOptions) *config.NodeGroupAutoscalingOptions {
	options := m.getAutoscalingOptions(asg.AwsRef)
	if options == nil || len(options) == 0 {
		return &defaults
	}

	if stringOpt, found := options[config.DefaultScaleDownUtilizationThresholdKey]; found {
		if opt, err := strconv.ParseFloat(stringOpt, 64); err != nil {
			klog.Warningf("failed to convert asg %s %s tag to float: %v",
				asg.Name, config.DefaultScaleDownUtilizationThresholdKey, err)
		} else {
			defaults.ScaleDownUtilizationThreshold = opt
		}
	}

	if stringOpt, found := options[config.DefaultScaleDownGpuUtilizationThresholdKey]; found {
		if opt, err := strconv.ParseFloat(stringOpt, 64); err != nil {
			klog.Warningf("failed to convert asg %s %s tag to float: %v",
				asg.Name, config.DefaultScaleDownGpuUtilizationThresholdKey, err)
		} else {
			defaults.ScaleDownGpuUtilizationThreshold = opt
		}
	}

	if stringOpt, found := options[config.DefaultScaleDownUnneededTimeKey]; found {
		if opt, err := time.ParseDuration(stringOpt); err != nil {
			klog.Warningf("failed to convert asg %s %s tag to duration: %v",
				asg.Name, config.DefaultScaleDownUnneededTimeKey, err)
		} else {
			defaults.ScaleDownUnneededTime = opt
		}
	}

	if stringOpt, found := options[config.DefaultScaleDownUnreadyTimeKey]; found {
		if opt, err := time.ParseDuration(stringOpt); err != nil {
			klog.Warningf("failed to convert asg %s %s tag to duration: %v",
				asg.Name, config.DefaultScaleDownUnreadyTimeKey, err)
		} else {
			defaults.ScaleDownUnreadyTime = opt
		}
	}

	return &defaults
}

func (m *AwsManager) buildNodeFromTemplate(asg *asg, template *asgTemplate) (*apiv1.Node, error) {
	node := apiv1.Node{}
	nodeName := fmt.Sprintf("%s-asg-%d", asg.Name, rand.Int63())

	node.ObjectMeta = metav1.ObjectMeta{
		Name:     nodeName,
		SelfLink: fmt.Sprintf("/api/v1/nodes/%s", nodeName),
		Labels:   map[string]string{},
	}

	node.Status = apiv1.NodeStatus{
		Capacity: apiv1.ResourceList{},
	}

	// TODO: get a real value.
	node.Status.Capacity[apiv1.ResourcePods] = *resource.NewQuantity(110, resource.DecimalSI)
	node.Status.Capacity[apiv1.ResourceCPU] = *resource.NewQuantity(template.InstanceType.VCPU, resource.DecimalSI)
	node.Status.Capacity[gpu.ResourceNvidiaGPU] = *resource.NewQuantity(template.InstanceType.GPU, resource.DecimalSI)
	node.Status.Capacity[apiv1.ResourceMemory] = *resource.NewQuantity(template.InstanceType.MemoryMb*1024*1024, resource.DecimalSI)

	if err := m.updateCapacityWithRequirementsOverrides(&node.Status.Capacity, asg.MixedInstancesPolicy); err != nil {
		return nil, err
	}

	resourcesFromTags := extractAllocatableResourcesFromAsg(template.Tags)
	klog.V(5).Infof("Extracted resources from ASG tags %v", resourcesFromTags)
	for resourceName, val := range resourcesFromTags {
		node.Status.Capacity[apiv1.ResourceName(resourceName)] = *val
	}

	// TODO: use proper allocatable!!
	node.Status.Allocatable = node.Status.Capacity

	// GenericLabels
	node.Labels = cloudprovider.JoinStringMaps(node.Labels, buildGenericLabels(template, nodeName))

	// NodeLabels
	node.Labels = cloudprovider.JoinStringMaps(node.Labels, extractLabelsFromAsg(template.Tags))

	node.Spec.Taints = extractTaintsFromAsg(template.Tags)

	if nodegroupName, clusterName := node.Labels["nodegroup-name"], node.Labels["cluster-name"]; nodegroupName != "" && clusterName != "" {
		klog.V(5).Infof("Nodegroup %s in cluster %s is an EKS managed nodegroup.", nodegroupName, clusterName)

		// Call AWS EKS DescribeNodegroup API, check if keys already exist in Labels and do NOT overwrite
		mngLabels, err := m.managedNodegroupCache.getManagedNodegroupLabels(nodegroupName, clusterName)
		if err != nil {
			klog.Errorf("Failed to get labels from EKS DescribeNodegroup API for nodegroup %s in cluster %s because %s.", nodegroupName, clusterName, err)
		} else if mngLabels != nil && len(mngLabels) > 0 {
			node.Labels = joinNodeLabelsChoosingUserValuesOverAPIValues(node.Labels, mngLabels)
			klog.V(5).Infof("node.Labels : %+v\n", node.Labels)
		}

		mngTaints, err := m.managedNodegroupCache.getManagedNodegroupTaints(nodegroupName, clusterName)
		if err != nil {
			klog.Errorf("Failed to get taints from EKS DescribeNodegroup API for nodegroup %s in cluster %s because %s.", nodegroupName, clusterName, err)
		} else if mngTaints != nil && len(mngTaints) > 0 {
			node.Spec.Taints = append(node.Spec.Taints, mngTaints...)
			klog.V(5).Infof("node.Spec.Taints : %+v\n", node.Spec.Taints)
		}
	}

	node.Status.Conditions = cloudprovider.BuildReadyConditions()
	return &node, nil
}

func joinNodeLabelsChoosingUserValuesOverAPIValues(extractedLabels map[string]string, mngLabels map[string]string) map[string]string {
	result := make(map[string]string)

	// Copy Generic Labels and Labels from ASG
	for k, v := range extractedLabels {
		result[k] = v
	}

	// Copy Labels from EKS DescribeNodegroup API call
	// If the there is a duplicate key, this will overwrite the ASG Tag specified values with the EKS DescribeNodegroup API values
	// We are overwriting them because it seems like EKS isn't sending the ASG Tags to Kubernetes itself
	//     so scale ups based on the ASG Tag aren't working
	for k, v := range mngLabels {
		result[k] = v
	}

	return result
}

func (m *AwsManager) updateCapacityWithRequirementsOverrides(capacity *apiv1.ResourceList, policy *mixedInstancesPolicy) error {
	if policy == nil {
		return nil
	}

	instanceRequirements, err := m.getInstanceRequirementsFromMixedInstancesPolicy(policy)
	if err != nil {
		return fmt.Errorf("error while building node template using instance requirements: (%s)", err)
	}

	if instanceRequirements.VCpuCount != nil && instanceRequirements.VCpuCount.Min != nil {
		(*capacity)[apiv1.ResourceCPU] = *resource.NewQuantity(*instanceRequirements.VCpuCount.Min, resource.DecimalSI)
	}

	if instanceRequirements.MemoryMiB != nil && instanceRequirements.MemoryMiB.Min != nil {
		(*capacity)[apiv1.ResourceMemory] = *resource.NewQuantity(*instanceRequirements.MemoryMiB.Min*1024*1024, resource.DecimalSI)
	}

	for _, manufacturer := range instanceRequirements.AcceleratorManufacturers {
		if *manufacturer == autoscaling.AcceleratorManufacturerNvidia {
			for _, acceleratorType := range instanceRequirements.AcceleratorTypes {
				if *acceleratorType == autoscaling.AcceleratorTypeGpu {
					(*capacity)[gpu.ResourceNvidiaGPU] = *resource.NewQuantity(*instanceRequirements.AcceleratorCount.Min, resource.DecimalSI)
				}
			}
		}
	}

	return nil
}

func (m *AwsManager) getInstanceRequirementsFromMixedInstancesPolicy(policy *mixedInstancesPolicy) (*ec2.InstanceRequirements, error) {
	instanceRequirements := &ec2.InstanceRequirements{}
	if policy.instanceRequirementsOverrides != nil {
		var err error
		instanceRequirements, err = m.awsService.getEC2RequirementsFromAutoscaling(policy.instanceRequirementsOverrides)
		if err != nil {
			return nil, err
		}
	} else if policy.launchTemplate != nil {
		templateData, err := m.awsService.getLaunchTemplateData(policy.launchTemplate.name, policy.launchTemplate.version)
		if err != nil {
			return nil, err
		}

		if templateData.InstanceRequirements != nil {
			instanceRequirements = templateData.InstanceRequirements
		}
	}
	return instanceRequirements, nil
}

func buildGenericLabels(template *asgTemplate, nodeName string) map[string]string {
	result := make(map[string]string)

	result[apiv1.LabelArchStable] = template.InstanceType.Architecture
	result[apiv1.LabelOSStable] = cloudprovider.DefaultOS

	result[apiv1.LabelInstanceTypeStable] = template.InstanceType.InstanceType

	result[apiv1.LabelTopologyRegion] = template.Region
	result[apiv1.LabelTopologyZone] = template.Zone
	result[apiv1.LabelHostname] = nodeName
	return result
}

func extractLabelsFromAsg(tags []*autoscaling.TagDescription) map[string]string {
	result := make(map[string]string)

	for _, tag := range tags {
		k := *tag.Key
		v := *tag.Value
		splits := strings.Split(k, "k8s.io/cluster-autoscaler/node-template/label/")
		// Extract EKS labels from ASG
		if len(splits) <= 1 {
			splits = strings.Split(k, "eks:")
		}
		if len(splits) > 1 {
			label := splits[1]
			if label != "" {
				result[label] = v
			}
		}
	}

	return result
}

func extractAutoscalingOptionsFromTags(tags []*autoscaling.TagDescription) map[string]string {
	options := make(map[string]string)
	for _, tag := range tags {
		if !strings.HasPrefix(aws.StringValue(tag.Key), optionsTagsPrefix) {
			continue
		}
		splits := strings.Split(aws.StringValue(tag.Key), optionsTagsPrefix)
		if len(splits) != 2 || splits[1] == "" {
			continue
		}
		options[splits[1]] = aws.StringValue(tag.Value)
	}
	return options
}

func extractAllocatableResourcesFromAsg(tags []*autoscaling.TagDescription) map[string]*resource.Quantity {
	result := make(map[string]*resource.Quantity)

	for _, tag := range tags {
		k := *tag.Key
		v := *tag.Value
		splits := strings.Split(k, "k8s.io/cluster-autoscaler/node-template/resources/")
		if len(splits) > 1 {
			label := splits[1]
			if label != "" {
				quantity, err := resource.ParseQuantity(v)
				if err != nil {
					continue
				}
				result[label] = &quantity
			}
		}
	}

	return result
}

func extractTaintsFromAsg(tags []*autoscaling.TagDescription) []apiv1.Taint {
	taints := make([]apiv1.Taint, 0)

	for _, tag := range tags {
		k := *tag.Key
		v := *tag.Value
		// The tag value must be in the format <tag>:NoSchedule
		r, _ := regexp.Compile("(.*):(?:NoSchedule|NoExecute|PreferNoSchedule)")
		if r.MatchString(v) {
			splits := strings.Split(k, "k8s.io/cluster-autoscaler/node-template/taint/")
			if len(splits) > 1 {
				values := strings.SplitN(v, ":", 2)
				if len(values) > 1 {
					taints = append(taints, apiv1.Taint{
						Key:    splits[1],
						Value:  values[0],
						Effect: apiv1.TaintEffect(values[1]),
					})
				}
			}
		}
	}
	return taints
}

// An asgAutoDiscoveryConfig specifies how to autodiscover AWS ASGs.
type asgAutoDiscoveryConfig struct {
	// Tags to match on.
	// Any ASG with all of the provided tag keys will be autoscaled.
	Tags map[string]string
}

// ParseASGAutoDiscoverySpecs returns any provided NodeGroupAutoDiscoverySpecs
// parsed into configuration appropriate for ASG autodiscovery.
func parseASGAutoDiscoverySpecs(o cloudprovider.NodeGroupDiscoveryOptions) ([]asgAutoDiscoveryConfig, error) {
	cfgs := make([]asgAutoDiscoveryConfig, len(o.NodeGroupAutoDiscoverySpecs))
	var err error
	for i, spec := range o.NodeGroupAutoDiscoverySpecs {
		cfgs[i], err = parseASGAutoDiscoverySpec(spec)
		if err != nil {
			return nil, err
		}
	}
	return cfgs, nil
}

func parseASGAutoDiscoverySpec(spec string) (asgAutoDiscoveryConfig, error) {
	cfg := asgAutoDiscoveryConfig{}

	tokens := strings.SplitN(spec, ":", 2)
	if len(tokens) != 2 {
		return cfg, fmt.Errorf("invalid node group auto discovery spec specified via --node-group-auto-discovery: %s", spec)
	}
	discoverer := tokens[0]
	if discoverer != autoDiscovererTypeASG {
		return cfg, fmt.Errorf("unsupported discoverer specified: %s", discoverer)
	}
	param := tokens[1]
	kv := strings.SplitN(param, "=", 2)
	if len(kv) != 2 {
		return cfg, fmt.Errorf("invalid key=value pair %s", kv)
	}
	k, v := kv[0], kv[1]
	if k != asgAutoDiscovererKeyTag {
		return cfg, fmt.Errorf("unsupported parameter key \"%s\" is specified for discoverer \"%s\". The only supported key is \"%s\"", k, discoverer, asgAutoDiscovererKeyTag)
	}
	if v == "" {
		return cfg, errors.New("tag value not supplied")
	}
	p := strings.Split(v, ",")
	if len(p) == 0 {
		return cfg, fmt.Errorf("invalid ASG tag for auto discovery specified: ASG tag must not be empty")
	}
	cfg.Tags = make(map[string]string, len(p))
	for _, label := range p {
		lp := strings.SplitN(label, "=", 2)
		if len(lp) > 1 {
			cfg.Tags[lp[0]] = lp[1]
			continue
		}
		cfg.Tags[lp[0]] = ""
	}
	return cfg, nil
}
