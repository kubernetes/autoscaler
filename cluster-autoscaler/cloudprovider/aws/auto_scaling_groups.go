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
	"sync"
	"time"

	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/util/wait"
)

type autoScalingGroups struct {
	registeredAsgs           []*asgInformation
	instanceToAsg            map[AwsRef]*Asg
	cacheMutex               sync.Mutex
	instancesNotInManagedAsg map[AwsRef]struct{}
	service                  autoScalingWrapper
}

func newAutoScalingGroups(service autoScalingWrapper) *autoScalingGroups {
	registry := &autoScalingGroups{
		registeredAsgs:           make([]*asgInformation, 0),
		service:                  service,
		instanceToAsg:            make(map[AwsRef]*Asg),
		instancesNotInManagedAsg: make(map[AwsRef]struct{}),
	}

	go wait.Forever(func() {
		registry.cacheMutex.Lock()
		defer registry.cacheMutex.Unlock()
		if err := registry.regenerateCache(); err != nil {
			glog.Errorf("Error while regenerating Asg cache: %v", err)
		}
	}, time.Hour)

	return registry
}

// Register registers asg in Aws Manager.
func (m *autoScalingGroups) Register(asg *Asg) {
	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()

	m.registeredAsgs = append(m.registeredAsgs, &asgInformation{
		config: asg,
	})
}

// FindForInstance returns AsgConfig of the given Instance
func (m *autoScalingGroups) FindForInstance(instance *AwsRef) (*Asg, error) {
	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()
	if config, found := m.instanceToAsg[*instance]; found {
		return config, nil
	}
	if _, found := m.instancesNotInManagedAsg[*instance]; found {
		// The instance is already known to not belong to any configured ASG
		// Skip regenerateCache so that we won't unnecessarily call DescribeAutoScalingGroups
		// See https://github.com/kubernetes/contrib/issues/2541
		return nil, nil
	}
	if err := m.regenerateCache(); err != nil {
		return nil, fmt.Errorf("Error while looking for ASG for instance %+v, error: %v", *instance, err)
	}
	if config, found := m.instanceToAsg[*instance]; found {
		return config, nil
	}
	// instance does not belong to any configured ASG
	glog.V(6).Infof("Instance %+v is not in any ASG managed by CA. CA is now memorizing the fact not to unnecessarily call AWS API afterwards trying to find the unexistent managed ASG for the instance", *instance)
	m.instancesNotInManagedAsg[*instance] = struct{}{}
	return nil, nil
}

func (m *autoScalingGroups) regenerateCache() error {
	newCache := make(map[AwsRef]*Asg)

	for _, asg := range m.registeredAsgs {
		glog.V(4).Infof("Regenerating ASG information for %s", asg.config.Name)

		group, err := m.service.getAutoscalingGroupByName(asg.config.Name)
		if err != nil {
			return err
		}
		for _, instance := range group.Instances {
			ref := AwsRef{Name: *instance.InstanceId}
			newCache[ref] = asg.config
		}
	}

	m.instanceToAsg = newCache
	return nil
}
