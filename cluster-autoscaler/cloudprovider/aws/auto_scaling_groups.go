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
	"time"

	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/golang/glog"
)

const scaleToZeroSupported = false

type asgCache struct {
	registeredAsgs     []*asgInformation
	instanceToAsg      map[AwsRef]*Asg
	notInRegisteredAsg map[AwsRef]bool
	mutex              sync.Mutex
	service            autoScalingWrapper
	interrupt          chan struct{}
}

func newASGCache(service autoScalingWrapper) (*asgCache, error) {
	registry := &asgCache{
		registeredAsgs:     make([]*asgInformation, 0),
		service:            service,
		instanceToAsg:      make(map[AwsRef]*Asg),
		notInRegisteredAsg: make(map[AwsRef]bool),
		interrupt:          make(chan struct{}),
	}
	go wait.Until(func() {
		registry.mutex.Lock()
		defer registry.mutex.Unlock()
		if err := registry.regenerate(); err != nil {
			glog.Errorf("Error while regenerating Asg cache: %v", err)
		}
	}, time.Hour, registry.interrupt)

	return registry, nil
}

// Register ASG. Returns true if the ASG was registered.
func (m *asgCache) Register(asg *Asg) bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for i := range m.registeredAsgs {
		if existing := m.registeredAsgs[i].config; existing.AwsRef == asg.AwsRef {
			if reflect.DeepEqual(existing, asg) {
				return false
			}
			m.registeredAsgs[i].config = asg
			glog.V(4).Infof("Updated ASG %s", asg.AwsRef.Name)
			m.invalidateUnownedInstanceCache()
			return true
		}
	}

	glog.V(1).Infof("Registering ASG %s", asg.AwsRef.Name)
	m.registeredAsgs = append(m.registeredAsgs, &asgInformation{
		config: asg,
	})
	m.invalidateUnownedInstanceCache()
	return true
}

// Unregister ASG. Returns true if the ASG was unregistered.
func (m *asgCache) Unregister(asg *Asg) bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	updated := make([]*asgInformation, 0, len(m.registeredAsgs))
	changed := false
	for _, existing := range m.registeredAsgs {
		if existing.config.AwsRef == asg.AwsRef {
			glog.V(1).Infof("Unregistered ASG %s", asg.AwsRef.Name)
			changed = true
			continue
		}
		updated = append(updated, existing)
	}
	m.registeredAsgs = updated
	return changed
}

func (m *asgCache) get() []*asgInformation {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	return m.registeredAsgs
}

// FindForInstance returns AsgConfig of the given Instance
func (m *asgCache) FindForInstance(instance *AwsRef) (*Asg, error) {
	// TODO(negz): Prevent this calling describe ASGs too often.
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.notInRegisteredAsg[*instance] {
		// We already know we don't own this instance. Return early and avoid
		// additional calls to describe ASGs.
		return nil, nil
	}

	if config, found := m.instanceToAsg[*instance]; found {
		return config, nil
	}
	if err := m.regenerate(); err != nil {
		return nil, fmt.Errorf("Error while looking for ASG for instance %+v, error: %v", *instance, err)
	}
	if config, found := m.instanceToAsg[*instance]; found {
		return config, nil
	}

	m.notInRegisteredAsg[*instance] = true
	return nil, nil
}

func (m *asgCache) invalidateUnownedInstanceCache() {
	glog.V(4).Info("Invalidating unowned instance cache")
	m.notInRegisteredAsg = make(map[AwsRef]bool)
}

func (m *asgCache) regenerate() error {
	newCache := make(map[AwsRef]*Asg)

	names := make([]string, len(m.registeredAsgs))
	configs := make(map[string]*Asg)
	for i, asg := range m.registeredAsgs {
		names[i] = asg.config.Name
		configs[asg.config.Name] = asg.config
	}

	glog.V(4).Infof("Regenerating instance to ASG map for ASGs: %v", names)
	groups, err := m.service.getAutoscalingGroupsByNames(names)
	if err != nil {
		return err
	}
	for _, group := range groups {
		for _, instance := range group.Instances {
			ref := AwsRef{Name: aws.StringValue(instance.InstanceId)}
			newCache[ref] = configs[aws.StringValue(group.AutoScalingGroupName)]
		}
	}

	m.instanceToAsg = newCache
	return nil
}

// Cleanup closes the channel to signal the go routine to stop that is handling the cache
func (m *asgCache) Cleanup() {
	close(m.interrupt)
}
