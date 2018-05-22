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
	"sync"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config/dynamic"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/golang/glog"
)

const scaleToZeroSupported = true

type asgCache struct {
	registeredAsgs []*asg
	asgToInstances map[AwsRef][]AwsInstanceRef
	instanceToAsg  map[AwsInstanceRef]*asg
	mutex          sync.Mutex
	service        autoScalingWrapper
	interrupt      chan struct{}

	asgAutoDiscoverySpecs []cloudprovider.ASGAutoDiscoveryConfig
	explicitlyConfigured  map[AwsRef]bool
}

type asg struct {
	AwsRef

	minSize int
	maxSize int
	curSize int

	AvailabilityZones       []string
	LaunchConfigurationName string
	Tags                    []*autoscaling.TagDescription
}

func newASGCache(service autoScalingWrapper, explicitSpecs []string, autoDiscoverySpecs []cloudprovider.ASGAutoDiscoveryConfig) (*asgCache, error) {
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

// Register ASG. Returns true if the ASG was registered.
func (m *asgCache) register(asg *asg) bool {
	for i := range m.registeredAsgs {
		if existing := m.registeredAsgs[i]; existing.AwsRef == asg.AwsRef {
			if reflect.DeepEqual(existing, asg) {
				return false
			}

			glog.V(4).Infof("Updating ASG %s", asg.AwsRef.Name)

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
			existing.Tags = asg.Tags

			return true
		}
	}
	glog.V(1).Infof("Registering ASG %s", asg.AwsRef.Name)
	m.registeredAsgs = append(m.registeredAsgs, asg)
	return true
}

// Unregister ASG. Returns true if the ASG was unregistered.
func (m *asgCache) unregister(a *asg) bool {
	updated := make([]*asg, 0, len(m.registeredAsgs))
	changed := false
	for _, existing := range m.registeredAsgs {
		if existing.AwsRef == a.AwsRef {
			glog.V(1).Infof("Unregistered ASG %s", a.AwsRef.Name)
			changed = true
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

	return nil, fmt.Errorf("Error while looking for instances of ASG: %s", ref)
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
	glog.V(4).Infof("Regenerating instance to ASG map for ASGs: %v", refreshNames)
	groups, err := m.service.getAutoscalingGroupsByNames(refreshNames)
	if err != nil {
		return err
	}

	// Register or update ASGs
	exists := make(map[AwsRef]bool)
	for _, group := range groups {
		asg, err := m.buildAsgFromAWS(group)
		if err != nil {
			return err
		}
		exists[asg.AwsRef] = true

		m.register(asg)

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
		Tags: g.Tags,
	}
	return asg, nil
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
