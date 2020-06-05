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

package baiducloud

import (
	"fmt"
	"sync"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/baiducloud/baiducloud-sdk-go/cce"
	klog "k8s.io/klog/v2"
)

type autoScalingGroups struct {
	cloudConfig              *CloudConfig
	cceClient                *cce.Client
	registeredAsgs           []*asgInformation
	instanceToAsg            map[string]*Asg
	cacheMutex               sync.Mutex
	instancesNotInManagedAsg map[string]struct{}
}

func newAutoScalingGroups(cfg *CloudConfig, cceClient *cce.Client) *autoScalingGroups {
	registry := &autoScalingGroups{
		cloudConfig:              cfg,
		cceClient:                cceClient,
		registeredAsgs:           make([]*asgInformation, 0),
		instanceToAsg:            make(map[string]*Asg),
		instancesNotInManagedAsg: make(map[string]struct{}),
	}
	return registry
}

// Register registers Asg in Manager.
func (m *autoScalingGroups) Register(asg *Asg) {
	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()

	m.registeredAsgs = append(m.registeredAsgs, &asgInformation{
		config: asg,
	})
}

// FindForInstance returns AsgConfig of the given Instance.
func (m *autoScalingGroups) FindForInstance(instanceID string) (*Asg, error) {
	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()
	if config, found := m.instanceToAsg[instanceID]; found {
		return config, nil
	}
	if _, found := m.instancesNotInManagedAsg[instanceID]; found {
		// The instance is already known to not belong to any configured ASG
		// Skip regenerateCache so that we won't unnecessarily call DescribeAutoScalingGroups
		// See https://github.com/kubernetes/contrib/issues/2541
		klog.V(4).Infof("instancesNotInManagedAsg")
		return nil, nil
	}
	if err := m.regenerateCache(); err != nil {
		return nil, fmt.Errorf("error while looking for ASG for instance %+v, error: %v", instanceID, err)
	}
	if config, found := m.instanceToAsg[instanceID]; found {
		return config, nil
	}
	// instance does not belong to any configured ASG.
	klog.V(4).Infof("Instance %+v is not in any ASG managed by CA."+
		" CA is now memorizing the fact not to unnecessarily call BCE API afterwards trying to find the "+
		"unexistent managed ASG for the instance", instanceID)
	m.instancesNotInManagedAsg[instanceID] = struct{}{}
	return nil, nil
}

func (m *autoScalingGroups) regenerateCache() error {
	newCache := make(map[string]*Asg)
	// TODO: Currently, baiducloud cloudprovider does not support Multiple ASG.
	for _, asg := range m.registeredAsgs {
		klog.V(4).Infof("regenerating ASG information for %s", asg.config.Name)
		instanceList, err := m.cceClient.GetAsgNodes(asg.config.Name, m.cloudConfig.ClusterID)
		if err != nil {
			return err
		}
		for _, instance := range instanceList {
			newCache[instance.InstanceId] = asg.config
		}
	}
	m.instanceToAsg = newCache
	return nil
}
