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

//go:generate go run ec2_instance_types/gen.go

package aws

import (
	"fmt"
	"io"
	"math/rand"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/golang/glog"
	"gopkg.in/gcfg.v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config/dynamic"
	provider_aws "k8s.io/kubernetes/pkg/cloudprovider/providers/aws"
	kubeletapis "k8s.io/kubernetes/pkg/kubelet/apis"
)

const (
	operationWaitTimeout    = 5 * time.Second
	operationPollInterval   = 100 * time.Millisecond
	maxRecordsReturnedByAPI = 100
	refreshInterval         = 1 * time.Minute
)

type asgInformation struct {
	config   *Asg
	basename string
}

// AwsManager is handles aws communication and data caching.
type AwsManager struct {
	service               autoScalingWrapper
	asgCache              *asgCache
	lastRefresh           time.Time
	asgAutoDiscoverySpecs []cloudprovider.ASGAutoDiscoveryConfig
	explicitlyConfigured  map[AwsRef]bool
}

type asgTemplate struct {
	InstanceType *instanceType
	Region       string
	Zone         string
	Tags         []*autoscaling.TagDescription
}

// createAwsManagerInternal allows for a customer autoScalingWrapper to be passed in by tests
func createAWSManagerInternal(
	configReader io.Reader,
	discoveryOpts cloudprovider.NodeGroupDiscoveryOptions,
	service *autoScalingWrapper,
) (*AwsManager, error) {
	if configReader != nil {
		var cfg provider_aws.CloudConfig
		if err := gcfg.ReadInto(&cfg, configReader); err != nil {
			glog.Errorf("Couldn't read config: %v", err)
			return nil, err
		}
	}

	if service == nil {
		service = &autoScalingWrapper{
			autoscaling.New(session.New()),
		}
	}

	cache, err := newASGCache(*service)
	if err != nil {
		return nil, err
	}

	specs, err := discoveryOpts.ParseASGAutoDiscoverySpecs()
	if err != nil {
		return nil, err
	}

	manager := &AwsManager{
		service:               *service,
		asgCache:              cache,
		asgAutoDiscoverySpecs: specs,
		explicitlyConfigured:  make(map[AwsRef]bool),
	}

	if err := manager.fetchExplicitAsgs(discoveryOpts.NodeGroupSpecs); err != nil {
		return nil, err
	}

	if err := manager.forceRefresh(); err != nil {
		return nil, err
	}

	return manager, nil
}

// CreateAwsManager constructs awsManager object.
func CreateAwsManager(configReader io.Reader, discoveryOpts cloudprovider.NodeGroupDiscoveryOptions) (*AwsManager, error) {
	return createAWSManagerInternal(configReader, discoveryOpts, nil)
}

// Fetch explicitly configured ASGs. These ASGs should never be unregistered
// during refreshes, even if they no longer exist in AWS.
func (m *AwsManager) fetchExplicitAsgs(specs []string) error {
	changed := false
	for _, spec := range specs {
		asg, err := m.buildAsgFromSpec(spec)
		if err != nil {
			return fmt.Errorf("failed to parse node group spec: %v", err)
		}
		if m.RegisterAsg(asg) {
			changed = true
		}
		m.explicitlyConfigured[asg.AwsRef] = true
	}

	if changed {
		if err := m.regenerateCache(); err != nil {
			return err
		}
	}
	return nil
}

func (m *AwsManager) buildAsgFromSpec(spec string) (*Asg, error) {
	s, err := dynamic.SpecFromString(spec, scaleToZeroSupported)
	if err != nil {
		return nil, fmt.Errorf("failed to parse node group spec: %v", err)
	}
	asg := &Asg{
		awsManager: m,
		AwsRef:     AwsRef{Name: s.Name},
		minSize:    s.MinSize,
		maxSize:    s.MaxSize,
	}
	return asg, nil
}

// Fetch automatically discovered ASGs. These ASGs should be unregistered if
// they no longer exist in AWS.
func (m *AwsManager) fetchAutoAsgs() error {
	exists := make(map[AwsRef]bool)
	changed := false
	for _, spec := range m.asgAutoDiscoverySpecs {
		groups, err := m.getAutoscalingGroupsByTags(spec.TagKeys)
		if err != nil {
			return fmt.Errorf("cannot autodiscover ASGs: %s", err)
		}
		for _, g := range groups {
			asg, err := m.buildAsgFromAWS(g)
			if err != nil {
				return err
			}
			exists[asg.AwsRef] = true
			if m.explicitlyConfigured[asg.AwsRef] {
				// This ASG was explicitly configured, but would also be
				// autodiscovered. We want the explicitly configured min and max
				// nodes to take precedence.
				glog.V(3).Infof("Ignoring explicitly configured ASG %s for autodiscovery.", asg.AwsRef.Name)
				continue
			}
			if m.RegisterAsg(asg) {
				glog.V(3).Infof("Autodiscovered ASG %s using tags %v", asg.AwsRef.Name, spec.TagKeys)
				changed = true
			}
		}
	}

	for _, asg := range m.getAsgs() {
		if !exists[asg.config.AwsRef] && !m.explicitlyConfigured[asg.config.AwsRef] {
			m.UnregisterAsg(asg.config)
			changed = true
		}
	}

	if changed {
		if err := m.regenerateCache(); err != nil {
			return err
		}
	}

	return nil
}

func (m *AwsManager) buildAsgFromAWS(g *autoscaling.Group) (*Asg, error) {
	spec := dynamic.NodeGroupSpec{
		Name:               aws.StringValue(g.AutoScalingGroupName),
		MinSize:            int(aws.Int64Value(g.MinSize)),
		MaxSize:            int(aws.Int64Value(g.MaxSize)),
		SupportScaleToZero: scaleToZeroSupported,
	}
	if verr := spec.Validate(); verr != nil {
		return nil, fmt.Errorf("failed to create node group spec: %v", verr)
	}
	asg := &Asg{
		awsManager: m,
		AwsRef:     AwsRef{Name: spec.Name},
		minSize:    spec.MinSize,
		maxSize:    spec.MaxSize,
	}
	return asg, nil
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
	if err := m.fetchAutoAsgs(); err != nil {
		glog.Errorf("Failed to fetch ASGs: %v", err)
		return err
	}
	m.lastRefresh = time.Now()
	glog.V(2).Infof("Refreshed ASG list, next refresh after %v", m.lastRefresh.Add(refreshInterval))
	return nil
}

func (m *AwsManager) getAsgs() []*asgInformation {
	return m.asgCache.get()
}

// RegisterAsg registers an ASG.
func (m *AwsManager) RegisterAsg(asg *Asg) bool {
	return m.asgCache.Register(asg)
}

// UnregisterAsg unregisters an ASG.
func (m *AwsManager) UnregisterAsg(asg *Asg) bool {
	return m.asgCache.Unregister(asg)
}

// GetAsgForInstance returns AsgConfig of the given Instance
func (m *AwsManager) GetAsgForInstance(instance *AwsRef) (*Asg, error) {
	return m.asgCache.FindForInstance(instance)
}

func (m *AwsManager) regenerateCache() error {
	m.asgCache.mutex.Lock()
	defer m.asgCache.mutex.Unlock()
	return m.asgCache.regenerate()
}

// Cleanup the ASG cache.
func (m *AwsManager) Cleanup() {
	m.asgCache.Cleanup()
}

func (m *AwsManager) getAutoscalingGroupsByTags(keys []string) ([]*autoscaling.Group, error) {
	return m.service.getAutoscalingGroupsByTags(keys)
}

// GetAsgSize gets ASG size.
func (m *AwsManager) GetAsgSize(asgConfig *Asg) (int64, error) {
	params := &autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{aws.String(asgConfig.Name)},
		MaxRecords:            aws.Int64(1),
	}
	groups, err := m.service.DescribeAutoScalingGroups(params)

	if err != nil {
		return -1, err
	}

	if len(groups.AutoScalingGroups) < 1 {
		return -1, fmt.Errorf("Unable to get first autoscaling.Group for %s", asgConfig.Name)
	}
	asg := *groups.AutoScalingGroups[0]
	return *asg.DesiredCapacity, nil
}

// SetAsgSize sets ASG size.
func (m *AwsManager) SetAsgSize(asg *Asg, size int64) error {
	params := &autoscaling.SetDesiredCapacityInput{
		AutoScalingGroupName: aws.String(asg.Name),
		DesiredCapacity:      aws.Int64(size),
		HonorCooldown:        aws.Bool(false),
	}
	glog.V(0).Infof("Setting asg %s size to %d", asg.Id(), size)
	_, err := m.service.SetDesiredCapacity(params)
	if err != nil {
		return err
	}
	return nil
}

// DeleteInstances deletes the given instances. All instances must be controlled by the same ASG.
func (m *AwsManager) DeleteInstances(instances []*AwsRef) error {
	if len(instances) == 0 {
		return nil
	}
	commonAsg, err := m.asgCache.FindForInstance(instances[0])
	if err != nil {
		return err
	}
	for _, instance := range instances {
		asg, err := m.asgCache.FindForInstance(instance)
		if err != nil {
			return err
		}
		if asg != commonAsg {
			return fmt.Errorf("Connot delete instances which don't belong to the same ASG.")
		}
	}

	for _, instance := range instances {
		params := &autoscaling.TerminateInstanceInAutoScalingGroupInput{
			InstanceId:                     aws.String(instance.Name),
			ShouldDecrementDesiredCapacity: aws.Bool(true),
		}
		resp, err := m.service.TerminateInstanceInAutoScalingGroup(params)
		if err != nil {
			return err
		}
		glog.V(4).Infof(*resp.Activity.Description)
	}

	return nil
}

// GetAsgNodes returns Asg nodes.
func (m *AwsManager) GetAsgNodes(asg *Asg) ([]string, error) {
	result := make([]string, 0)
	group, err := m.service.getAutoscalingGroupByName(asg.Name)
	if err != nil {
		return []string{}, err
	}
	for _, instance := range group.Instances {
		result = append(result,
			fmt.Sprintf("aws:///%s/%s", *instance.AvailabilityZone, *instance.InstanceId))
	}
	return result, nil
}

func (m *AwsManager) getAsgTemplate(name string) (*asgTemplate, error) {
	asg, err := m.service.getAutoscalingGroupByName(name)
	if err != nil {
		return nil, err
	}

	instanceTypeName, err := m.service.getInstanceTypeByLCName(*asg.LaunchConfigurationName)
	if err != nil {
		return nil, err
	}

	if len(asg.AvailabilityZones) < 1 {
		return nil, fmt.Errorf("Unable to get first AvailabilityZone for %s", name)
	}

	az := *asg.AvailabilityZones[0]
	region := az[0 : len(az)-1]

	if len(asg.AvailabilityZones) > 1 {
		glog.Warningf("Found multiple availability zones, using %s\n", az)
	}

	return &asgTemplate{
		InstanceType: InstanceTypes[instanceTypeName],
		Region:       region,
		Zone:         az,
		Tags:         asg.Tags,
	}, nil
}

func (m *AwsManager) buildNodeFromTemplate(asg *Asg, template *asgTemplate) (*apiv1.Node, error) {
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
	node.Status.Capacity[apiv1.ResourceNvidiaGPU] = *resource.NewQuantity(template.InstanceType.GPU, resource.DecimalSI)
	node.Status.Capacity[apiv1.ResourceMemory] = *resource.NewQuantity(template.InstanceType.MemoryMb*1024*1024, resource.DecimalSI)

	// TODO: use proper allocatable!!
	node.Status.Allocatable = node.Status.Capacity

	// NodeLabels
	node.Labels = cloudprovider.JoinStringMaps(node.Labels, extractLabelsFromAsg(template.Tags))
	// GenericLabels
	node.Labels = cloudprovider.JoinStringMaps(node.Labels, buildGenericLabels(template, nodeName))

	node.Spec.Taints = extractTaintsFromAsg(template.Tags)

	node.Status.Conditions = cloudprovider.BuildReadyConditions()
	return &node, nil
}

func buildGenericLabels(template *asgTemplate, nodeName string) map[string]string {
	result := make(map[string]string)
	// TODO: extract it somehow
	result[kubeletapis.LabelArch] = cloudprovider.DefaultArch
	result[kubeletapis.LabelOS] = cloudprovider.DefaultOS

	result[kubeletapis.LabelInstanceType] = template.InstanceType.InstanceType

	result[kubeletapis.LabelZoneRegion] = template.Region
	result[kubeletapis.LabelZoneFailureDomain] = template.Zone
	result[kubeletapis.LabelHostname] = nodeName
	return result
}

func extractLabelsFromAsg(tags []*autoscaling.TagDescription) map[string]string {
	result := make(map[string]string)

	for _, tag := range tags {
		k := *tag.Key
		v := *tag.Value
		splits := strings.Split(k, "k8s.io/cluster-autoscaler/node-template/label/")
		if len(splits) > 1 {
			label := splits[1]
			if label != "" {
				result[label] = v
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
		splits := strings.Split(k, "k8s.io/cluster-autoscaler/node-template/taint/")
		if len(splits) > 1 {
			values := strings.SplitN(v, ":", 2)
			taints = append(taints, apiv1.Taint{
				Key:    splits[1],
				Value:  values[0],
				Effect: apiv1.TaintEffect(values[1]),
			})
		}
	}
	return taints
}
