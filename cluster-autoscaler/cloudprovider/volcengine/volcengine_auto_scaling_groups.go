/*
Copyright 2023 The Kubernetes Authors.

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

package volcengine

import (
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/volcengine/volcengine-go-sdk/volcengine"
	"k8s.io/klog/v2"
)

type autoScalingGroupsCache struct {
	registeredAsgs           []*AutoScalingGroup
	instanceToAsg            map[string]*AutoScalingGroup
	cacheMutex               sync.Mutex
	instancesNotInManagedAsg map[string]struct{}
	asgService               AutoScalingService
}

func newAutoScalingGroupsCache(asgService AutoScalingService) *autoScalingGroupsCache {
	registry := &autoScalingGroupsCache{
		registeredAsgs:           make([]*AutoScalingGroup, 0),
		instanceToAsg:            make(map[string]*AutoScalingGroup),
		instancesNotInManagedAsg: make(map[string]struct{}),
		asgService:               asgService,
	}

	go wait.Forever(func() {
		registry.cacheMutex.Lock()
		defer registry.cacheMutex.Unlock()
		if err := registry.regenerateCache(); err != nil {
			klog.Errorf("Error while regenerating ASG cache: %v", err)
		}
	}, time.Hour)

	return registry
}

func (c *autoScalingGroupsCache) Register(asg *AutoScalingGroup) {
	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()
	c.registeredAsgs = append(c.registeredAsgs, asg)
}

func (c *autoScalingGroupsCache) FindForInstance(instanceId string) (*AutoScalingGroup, error) {
	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()
	if asg, found := c.instanceToAsg[instanceId]; found {
		return asg, nil
	}
	if _, found := c.instancesNotInManagedAsg[instanceId]; found {
		return nil, nil
	}
	if err := c.regenerateCache(); err != nil {
		return nil, err
	}
	if asg, found := c.instanceToAsg[instanceId]; found {
		return asg, nil
	}

	// instance does not belong to any configured ASG
	c.instancesNotInManagedAsg[instanceId] = struct{}{}
	return nil, nil
}

func (c *autoScalingGroupsCache) regenerateCache() error {
	newCache := make(map[string]*AutoScalingGroup)
	for _, asg := range c.registeredAsgs {
		instances, err := c.asgService.ListScalingInstancesByGroupId(asg.asgId)
		if err != nil {
			return err
		}
		for _, instance := range instances {
			newCache[volcengine.StringValue(instance.InstanceId)] = asg
		}
	}
	c.instanceToAsg = newCache
	return nil
}
