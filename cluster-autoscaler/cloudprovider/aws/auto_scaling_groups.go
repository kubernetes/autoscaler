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
	"reflect"
	"strings"
	"sync"

	"k8s.io/autoscaler/cluster-autoscaler/config/dynamic"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	klog "k8s.io/klog/v2"
)

const (
	scaleToZeroSupported          = true
	placeholderInstanceNamePrefix = "i-placeholder"
)

type asgCache struct {
	registeredAsgs []*asg
	asgToInstances map[AwsRef][]AwsInstanceRef
	instanceToAsg  map[AwsInstanceRef]*asg
	mutex          sync.Mutex
	service        autoScalingWrapper
	interrupt      chan struct{}

	asgAutoDiscoverySpecs []asgAutoDiscoveryConfig
	explicitlyConfigured  map[AwsRef]bool
}

type launchTemplate struct {
	name    string
	version string
}

type mixedInstancesPolicy struct {
	launchTemplate         *launchTemplate
	instanceTypesOverrides []string
}

type asg struct {
	AwsRef

	minSize int
	maxSize int
	curSize int

	AvailabilityZones       []string
	LaunchConfigurationName string
	LaunchTemplate          *launchTemplate
	MixedInstancesPolicy    *mixedInstancesPolicy
	Tags                    []*autoscaling.TagDescription
}

func newASGCache(service autoScalingWrapper, explicitSpecs []string, autoDiscoverySpecs []asgAutoDiscoveryConfig) (*asgCache, error) {
	registry := &asgCache{
		registeredAsgs:        make([]*asg, 0),
		service:               service,
		asgToInstances:        make(map[AwsRef][]AwsInstanceRef),
		instanceToAsg:         make(map[AwsInstanceRef]*asg),
		interrupt:             make(chan struct{}),
		asgAutoDiscoverySpecs: autoDiscoverySpecs,
		explicitlyConfigured:  make(map[AwsRef]bool),
	}

	if err := registry.parseExplicitAsgs(explicitSpecs); err != nil {
		return nil, err
	}

	return registry, nil
}

// Fetch explicitly configured ASGs. These ASGs should never be unregistered
// during refreshes, even if they no longer exist in AWS.
func (m *asgCache) parseExplicitAsgs(specs []string) error {
	for _, spec := range specs {
		asg, err := m.buildAsgFromSpec(spec)
		if err != nil {
			return fmt.Errorf("failed to parse node group spec: %v", err)
		}
		m.explicitlyConfigured[asg.AwsRef] = true
		m.register(asg)
	}

	return nil
}

// Register ASG. Returns the registered ASG.
func (m *asgCache) register(asg *asg) *asg {
	for i := range m.registeredAsgs {
		if existing := m.registeredAsgs[i]; existing.AwsRef == asg.AwsRef {
			if reflect.DeepEqual(existing, asg) {
				return existing
			}

			klog.V(4).Infof("Updating ASG %s", asg.AwsRef.Name)

			// Explicit registered groups should always use the manually provided min/max
			// values and the not the ones returned by the API
			if !m.explicitlyConfigured[asg.AwsRef] {
				existing.minSize = asg.minSize
				existing.maxSize = asg.maxSize
			}

			existing.curSize = asg.curSize

			// Those information are mainly required to create templates when scaling
			// from zero
			existing.AvailabilityZones = asg.AvailabilityZones
			existing.LaunchConfigurationName = asg.LaunchConfigurationName
			existing.LaunchTemplate = asg.LaunchTemplate
			existing.MixedInstancesPolicy = asg.MixedInstancesPolicy
			existing.Tags = asg.Tags

			return existing
		}
	}
	klog.V(1).Infof("Registering ASG %s", asg.AwsRef.Name)
	m.registeredAsgs = append(m.registeredAsgs, asg)
	return asg
}

// Unregister ASG. Returns the unregistered ASG.
func (m *asgCache) unregister(a *asg) *asg {
	updated := make([]*asg, 0, len(m.registeredAsgs))
	var changed *asg
	for _, existing := range m.registeredAsgs {
		if existing.AwsRef == a.AwsRef {
			klog.V(1).Infof("Unregistered ASG %s", a.AwsRef.Name)
			changed = a
			continue
		}
		updated = append(updated, existing)
	}
	m.registeredAsgs = updated
	return changed
}

func (m *asgCache) buildAsgFromSpec(spec string) (*asg, error) {
	s, err := dynamic.SpecFromString(spec, scaleToZeroSupported)
	if err != nil {
		return nil, fmt.Errorf("failed to parse node group spec: %v", err)
	}
	asg := &asg{
		AwsRef:  AwsRef{Name: s.Name},
		minSize: s.MinSize,
		maxSize: s.MaxSize,
	}
	return asg, nil
}

// Get returns the currently registered ASGs
func (m *asgCache) Get() []*asg {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	return m.registeredAsgs
}

// FindForInstance returns AsgConfig of the given Instance
func (m *asgCache) FindForInstance(instance AwsInstanceRef) *asg {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	return m.findForInstance(instance)
}

func (m *asgCache) findForInstance(instance AwsInstanceRef) *asg {
	if asg, found := m.instanceToAsg[instance]; found {
		return asg
	}

	return nil
}

// InstancesByAsg returns the nodes of an ASG
func (m *asgCache) InstancesByAsg(ref AwsRef) ([]AwsInstanceRef, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if instances, found := m.asgToInstances[ref]; found {
		return instances, nil
	}

	return nil, fmt.Errorf("error while looking for instances of ASG: %s", ref)
}

func (m *asgCache) SetAsgSize(asg *asg, size int) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	return m.setAsgSizeNoLock(asg, size)
}

func (m *asgCache) setAsgSizeNoLock(asg *asg, size int) error {
	params := &autoscaling.SetDesiredCapacityInput{
		AutoScalingGroupName: aws.String(asg.Name),
		DesiredCapacity:      aws.Int64(int64(size)),
		HonorCooldown:        aws.Bool(false),
	}
	klog.V(0).Infof("Setting asg %s size to %d", asg.Name, size)
	_, err := m.service.SetDesiredCapacity(params)
	if err != nil {
		return err
	}

	// Proactively set the ASG size so autoscaler makes better decisions
	asg.curSize = size

	return nil
}

func (m *asgCache) decreaseAsgSizeByOneNoLock(asg *asg) error {
	return m.setAsgSizeNoLock(asg, asg.curSize-1)
}

// DeleteInstances deletes the given instances. All instances must be controlled by the same ASG.
func (m *asgCache) DeleteInstances(instances []*AwsInstanceRef) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if len(instances) == 0 {
		return nil
	}
	commonAsg := m.findForInstance(*instances[0])
	if commonAsg == nil {
		return fmt.Errorf("can't delete instance %s, which is not part of an ASG", instances[0].Name)
	}

	for _, instance := range instances {
		asg := m.findForInstance(*instance)

		if asg != commonAsg {
			instanceIds := make([]string, len(instances))
			for i, instance := range instances {
				instanceIds[i] = instance.Name
			}

			return fmt.Errorf("can't delete instances %s as they belong to at least two different ASGs (%s and %s)", strings.Join(instanceIds, ","), commonAsg.Name, asg.Name)
		}
	}

	for _, instance := range instances {
		// check if the instance is a placeholder - a requested instance that was never created by the node group
		// if it is, just decrease the size of the node group, as there's no specific instance we can remove
		if m.isPlaceholderInstance(instance) {
			klog.V(4).Infof("instance %s is detected as a placeholder, decreasing ASG requested size instead "+
				"of deleting instance", instance.Name)
			m.decreaseAsgSizeByOneNoLock(commonAsg)
		} else {
			params := &autoscaling.TerminateInstanceInAutoScalingGroupInput{
				InstanceId:                     aws.String(instance.Name),
				ShouldDecrementDesiredCapacity: aws.Bool(true),
			}
			resp, err := m.service.TerminateInstanceInAutoScalingGroup(params)
			if err != nil {
				return err
			}
			klog.V(4).Infof(*resp.Activity.Description)

			// Proactively decrement the size so autoscaler makes better decisions
			commonAsg.curSize--
		}
	}
	return nil
}

// isPlaceholderInstance checks if the given instance is only a placeholder
func (m *asgCache) isPlaceholderInstance(instance *AwsInstanceRef) bool {
	return strings.HasPrefix(instance.Name, placeholderInstanceNamePrefix)
}

// Fetch automatically discovered ASGs. These ASGs should be unregistered if
// they no longer exist in AWS.
func (m *asgCache) fetchAutoAsgNames() ([]string, error) {
	groupNames := make([]string, 0)

	for _, spec := range m.asgAutoDiscoverySpecs {
		names, err := m.service.getAutoscalingGroupNamesByTags(spec.Tags)
		if err != nil {
			return nil, fmt.Errorf("cannot autodiscover ASGs: %s", err)
		}

		groupNames = append(groupNames, names...)
	}

	return groupNames, nil
}

func (m *asgCache) buildAsgNames() ([]string, error) {
	// Collect explicitly specified names
	refreshNames := make([]string, len(m.explicitlyConfigured))
	i := 0
	for k := range m.explicitlyConfigured {
		refreshNames[i] = k.Name
		i++
	}

	// Append auto-discovered names
	autoDiscoveredNames, err := m.fetchAutoAsgNames()
	if err != nil {
		return nil, err
	}
	for _, name := range autoDiscoveredNames {
		autoRef := AwsRef{Name: name}

		if m.explicitlyConfigured[autoRef] {
			// This ASG was already explicitly configured, we only need to fetch it once
			continue
		}

		refreshNames = append(refreshNames, name)
	}

	return refreshNames, nil
}

// regenerate the cached view of explicitly configured and auto-discovered ASGs
func (m *asgCache) regenerate() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	newInstanceToAsgCache := make(map[AwsInstanceRef]*asg)
	newAsgToInstancesCache := make(map[AwsRef][]AwsInstanceRef)

	// Build list of knowns ASG names
	refreshNames, err := m.buildAsgNames()
	if err != nil {
		return err
	}

	// Fetch details of all ASGs
	klog.V(4).Infof("Regenerating instance to ASG map for ASGs: %v", refreshNames)
	groups, err := m.service.getAutoscalingGroupsByNames(refreshNames)
	if err != nil {
		return err
	}

	err = m.service.populateLaunchConfigurationInstanceTypeCache(groups)
	if err != nil {
		klog.Warningf("Failed to fully populate all launchConfigurations: %v", err)
	}

	// If currently any ASG has more Desired than running Instances, introduce placeholders
	// for the instances to come up. This is required to track Desired instances that
	// will never come up, like with Spot Request that can't be fulfilled
	groups = m.createPlaceholdersForDesiredNonStartedInstances(groups)

	// Register or update ASGs
	exists := make(map[AwsRef]bool)
	for _, group := range groups {
		asg, err := m.buildAsgFromAWS(group)
		if err != nil {
			return err
		}
		exists[asg.AwsRef] = true

		asg = m.register(asg)

		newAsgToInstancesCache[asg.AwsRef] = make([]AwsInstanceRef, len(group.Instances))

		for i, instance := range group.Instances {
			ref := m.buildInstanceRefFromAWS(instance)
			newInstanceToAsgCache[ref] = asg
			newAsgToInstancesCache[asg.AwsRef][i] = ref
		}
	}

	// Unregister no longer existing auto-discovered ASGs
	for _, asg := range m.registeredAsgs {
		if !exists[asg.AwsRef] && !m.explicitlyConfigured[asg.AwsRef] {
			m.unregister(asg)
		}
	}

	m.asgToInstances = newAsgToInstancesCache
	m.instanceToAsg = newInstanceToAsgCache
	return nil
}

func (m *asgCache) createPlaceholdersForDesiredNonStartedInstances(groups []*autoscaling.Group) []*autoscaling.Group {
	for _, g := range groups {
		desired := *g.DesiredCapacity
		real := int64(len(g.Instances))
		if desired <= real {
			continue
		}

		for i := real; i < desired; i++ {
			id := fmt.Sprintf("%s-%s-%d", placeholderInstanceNamePrefix, *g.AutoScalingGroupName, i)
			klog.V(4).Infof("Instance group %s has only %d instances created while requested count is %d. "+
				"Creating placeholder instance with ID %s.", *g.AutoScalingGroupName, real, desired, id)
			g.Instances = append(g.Instances, &autoscaling.Instance{
				InstanceId:       &id,
				AvailabilityZone: g.AvailabilityZones[0],
			})
		}
	}
	return groups
}

func (m *asgCache) buildAsgFromAWS(g *autoscaling.Group) (*asg, error) {
	spec := dynamic.NodeGroupSpec{
		Name:               aws.StringValue(g.AutoScalingGroupName),
		MinSize:            int(aws.Int64Value(g.MinSize)),
		MaxSize:            int(aws.Int64Value(g.MaxSize)),
		SupportScaleToZero: scaleToZeroSupported,
	}

	if verr := spec.Validate(); verr != nil {
		return nil, fmt.Errorf("failed to create node group spec: %v", verr)
	}

	asg := &asg{
		AwsRef:  AwsRef{Name: spec.Name},
		minSize: spec.MinSize,
		maxSize: spec.MaxSize,

		curSize:                 int(aws.Int64Value(g.DesiredCapacity)),
		AvailabilityZones:       aws.StringValueSlice(g.AvailabilityZones),
		LaunchConfigurationName: aws.StringValue(g.LaunchConfigurationName),
		Tags:                    g.Tags,
	}

	if g.LaunchTemplate != nil {
		asg.LaunchTemplate = m.buildLaunchTemplateFromSpec(g.LaunchTemplate)
	}

	if g.MixedInstancesPolicy != nil {
		getInstanceTypes := func(data []*autoscaling.LaunchTemplateOverrides) []string {
			res := make([]string, len(data))
			for i := 0; i < len(data); i++ {
				res[i] = aws.StringValue(data[i].InstanceType)
			}
			return res
		}

		asg.MixedInstancesPolicy = &mixedInstancesPolicy{
			launchTemplate:         m.buildLaunchTemplateFromSpec(g.MixedInstancesPolicy.LaunchTemplate.LaunchTemplateSpecification),
			instanceTypesOverrides: getInstanceTypes(g.MixedInstancesPolicy.LaunchTemplate.Overrides),
		}
	}

	return asg, nil
}

func (m *asgCache) buildLaunchTemplateFromSpec(ltSpec *autoscaling.LaunchTemplateSpecification) *launchTemplate {
	// NOTE(jaypipes): The LaunchTemplateSpecification.Version is a pointer to
	// string. When the pointer is nil, EC2 AutoScaling API considers the value
	// to be "$Default", however aws.StringValue(ltSpec.Version) will return an
	// empty string (which is not considered the same as "$Default" or a nil
	// string pointer. So, in order to not pass an empty string as the version
	// for the launch template when we communicate with the EC2 AutoScaling API
	// using the information in the launchTemplate, we store the string
	// "$Default" here when the ltSpec.Version is a nil pointer.
	//
	// See:
	//
	// https://github.com/kubernetes/autoscaler/issues/1728
	// https://github.com/aws/aws-sdk-go/blob/81fad3b797f4a9bd1b452a5733dd465eefef1060/service/autoscaling/api.go#L10666-L10671
	//
	// A cleaner alternative might be to make launchTemplate.version a string
	// pointer instead of a string, or even store the aws-sdk-go's
	// LaunchTemplateSpecification structs directly.
	var version string
	if ltSpec.Version == nil {
		version = "$Default"
	} else {
		version = aws.StringValue(ltSpec.Version)
	}
	return &launchTemplate{
		name:    aws.StringValue(ltSpec.LaunchTemplateName),
		version: version,
	}
}

func (m *asgCache) buildInstanceRefFromAWS(instance *autoscaling.Instance) AwsInstanceRef {
	providerID := fmt.Sprintf("aws:///%s/%s", aws.StringValue(instance.AvailabilityZone), aws.StringValue(instance.InstanceId))
	return AwsInstanceRef{
		ProviderID: providerID,
		Name:       aws.StringValue(instance.InstanceId),
	}
}

// Cleanup closes the channel to signal the go routine to stop that is handling the cache
func (m *asgCache) Cleanup() {
	close(m.interrupt)
}
