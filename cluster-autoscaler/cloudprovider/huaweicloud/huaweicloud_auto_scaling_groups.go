/*
Copyright 2020 The Kubernetes Authors.

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

package huaweicloud

import (
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
)

type autoScalingGroupCache struct {
	registeredAsgs           []*AutoScalingGroup
	instanceToAsg            map[string]*AutoScalingGroup
	cacheMutex               sync.Mutex
	instancesNotInManagedAsg map[string]struct{}
}

func newAutoScalingGroupCache() *autoScalingGroupCache {
	registry := &autoScalingGroupCache{
		instanceToAsg:            make(map[string]*AutoScalingGroup),
		instancesNotInManagedAsg: make(map[string]struct{}),
	}

	return registry
}

// Register asg in HuaweiCloud Manager.
func (asg *autoScalingGroupCache) Register(autoScalingGroup *AutoScalingGroup) {
	asg.cacheMutex.Lock()
	defer asg.cacheMutex.Unlock()

	asg.registeredAsgs = append(asg.registeredAsgs, autoScalingGroup)
}

// FindForInstance returns AsgConfig of the given Instance
func (asg *autoScalingGroupCache) FindForInstance(instanceId string, csm *cloudServiceManager) (*AutoScalingGroup, error) {
	asg.cacheMutex.Lock()
	defer asg.cacheMutex.Unlock()
	if config, found := asg.instanceToAsg[instanceId]; found {
		return config, nil
	}
	if _, found := asg.instancesNotInManagedAsg[instanceId]; found {
		return nil, nil
	}
	if err := asg.regenerateCache(csm); err != nil {
		return nil, err
	}
	if config, found := asg.instanceToAsg[instanceId]; found {
		return config, nil
	}
	// instance does not belong to any configured ASG
	asg.instancesNotInManagedAsg[instanceId] = struct{}{}
	return nil, nil
}

func (asg *autoScalingGroupCache) generateCache(csm *cloudServiceManager) {
	go wait.Forever(func() {
		asg.cacheMutex.Lock()
		defer asg.cacheMutex.Unlock()
		if err := asg.regenerateCache(csm); err != nil {
			klog.Errorf("failed to do regenerating ASG cache,because of %s", err.Error())
		}
	}, time.Hour)
}

func (asg *autoScalingGroupCache) regenerateCache(csm *cloudServiceManager) error {
	newCache := make(map[string]*AutoScalingGroup)

	for _, registeredAsg := range asg.registeredAsgs {
		instances, err := csm.GetInstances(registeredAsg.groupID)
		if err != nil {
			return err
		}
		for _, instance := range instances {
			newCache[instance.Id] = registeredAsg
		}
	}

	asg.instanceToAsg = newCache
	return nil
}
