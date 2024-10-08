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
	"sync"
)

type autoScalingGroups struct {
	registeredAsgs []*asgInformation
	instanceToAsg  map[string]*Asg
	cacheMutex     sync.Mutex
	service        *autoScalingWrapper
}

func newAutoScalingGroups(service *autoScalingWrapper) *autoScalingGroups {
	registry := &autoScalingGroups{
		registeredAsgs: make([]*asgInformation, 0),
		service:        service,
		instanceToAsg:  make(map[string]*Asg),
	}

	return registry
}

// Register registers asg in AliCloud Manager.
func (m *autoScalingGroups) Register(asg *Asg) {
	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()

	m.registeredAsgs = append(m.registeredAsgs, &asgInformation{
		config: asg,
	})
}

// FindForInstance returns AsgConfig of the given Instance
func (m *autoScalingGroups) FindForInstance(instanceId string) (*Asg, error) {
	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()
	if config, found := m.instanceToAsg[instanceId]; found {
		return config, nil
	}
	return nil, nil
}

func (m *autoScalingGroups) RegenerateCache() error {
	newCache := make(map[string]*Asg)

	for _, asg := range m.registeredAsgs {
		instances, err := m.service.getScalingInstancesByGroup(asg.config.id)
		if err != nil {
			return err
		}
		for _, instance := range instances {
			newCache[instance.InstanceId] = asg.config
		}
	}

	m.instanceToAsg = newCache
	return nil
}
