/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

package gce

import (
	"fmt"
	"sync"
	"time"

	"k8s.io/contrib/cluster-autoscaler/config"
	gceurl "k8s.io/contrib/cluster-autoscaler/utils/gce_url"

	"github.com/golang/glog"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	gce "google.golang.org/api/compute/v1"
	"k8s.io/kubernetes/pkg/util/wait"
)

type gceManager struct {
	migs       []*config.MigConfig
	service    *gce.Service
	migCache   map[config.InstanceConfig]*config.MigConfig
	cacheMutex sync.Mutex
}

// CreateGceManager constructs gceManager object.
func CreateGceManager(migs []*config.MigConfig) (*gceManager, error) {
	// Create Google Compute Engine service.
	client := oauth2.NewClient(oauth2.NoContext, google.ComputeTokenSource(""))
	gceService, err := gce.New(client)
	if err != nil {
		return nil, err
	}

	manager := &gceManager{
		migs:     migs,
		service:  gceService,
		migCache: map[config.InstanceConfig]*config.MigConfig{},
	}

	go wait.Forever(func() { manager.regenerateCacheIgnoreError() }, time.Hour)

	return manager, nil
}

func (m *gceManager) GetMigSize(migConf *config.MigConfig) (int64, error) {
	mig, err := m.service.InstanceGroupManagers.Get(migConf.Project, migConf.Zone, migConf.Name).Do()
	if err != nil {
		return -1, err
	}
	return mig.TargetSize, nil
}

func (m *gceManager) SetMigSize(migConf *config.MigConfig, size int64) error {
	op, err := m.service.InstanceGroupManagers.Resize(migConf.Project, migConf.Zone, migConf.Name, size).Do()
	if err != nil {
		return err
	}
	if err := m.waitForOp(op); err != nil {
		return err
	}
	return nil
}

func (m *gceManager) waitForOp(op *gce.Operation) error {
	return nil
}

// All instances must be controlled by the same MIG.
func (m *gceManager) DeleteInstances(instances []*config.InstanceConfig) error {
	if len(instances) == 0 {
		return nil
	}
	commonMig, err := m.GetMigForInstance(instances[0])
	if err != nil {
		return err
	}
	for _, instance := range instances {
		mig, err := m.GetMigForInstance(instance)
		if err != nil {
			return err
		}
		if mig != commonMig {
			return fmt.Errorf("Connot delete instances which don't belong to the same MIG.")
		}
	}

	req := gce.InstanceGroupManagersDeleteInstancesRequest{
		Instances: []string{},
	}
	for _, instance := range instances {
		req.Instances = append(req.Instances, gceurl.GenerateInstanceUrl(instance.Project, instance.Zone, instance.Name))
	}

	op, err := m.service.InstanceGroupManagers.DeleteInstances(commonMig.Project, commonMig.Zone, commonMig.Name, &req).Do()
	if err != nil {
		return err
	}
	if err := m.waitForOp(op); err != nil {
		return err
	}
	return nil
}

func (m *gceManager) GetMigForInstance(instance *config.InstanceConfig) (*config.MigConfig, error) {
	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()
	if mig, found := m.migCache[*instance]; found {
		return mig, nil
	}
	if err := m.regenerateCache(); err != nil {
		return nil, fmt.Errorf("Error while looking for MIG for instance %+v, error: %v", *instance, err)
	}
	if mig, found := m.migCache[*instance]; found {
		return mig, nil
	}
	return nil, fmt.Errorf("Instance %+v does not belong to any known MIG", *instance)
}

func (m *gceManager) regenerateCacheIgnoreError() {
	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()
	if err := m.regenerateCache(); err != nil {
		glog.Errorf("Error while regenerating Mig cache: %v", err)
	}
}

func (m *gceManager) regenerateCache() error {
	newMigCache := map[config.InstanceConfig]*config.MigConfig{}

	for _, mig := range m.migs {
		instances, err := m.service.InstanceGroupManagers.ListManagedInstances(mig.Project, mig.Zone, mig.Name).Do()
		if err != nil {
			return err
		}
		for _, instance := range instances.ManagedInstances {
			project, zone, name, err := gceurl.ParseInstanceUrl(instance.Instance)
			if err != nil {
				return err
			}
			newMigCache[config.InstanceConfig{Project: project, Zone: zone, Name: name}] = mig
		}
	}

	m.migCache = newMigCache
	return nil
}
