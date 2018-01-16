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

package azure

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
)

// Asg is a wrapper over NodeGroup interface.
type Asg interface {
	cloudprovider.NodeGroup

	getAzureRef() azureRef
}

type asgCache struct {
	registeredAsgs     []Asg
	instanceToAsg      map[azureRef]Asg
	notInRegisteredAsg map[azureRef]bool
	mutex              sync.Mutex
	interrupt          chan struct{}
}

func newAsgCache() (*asgCache, error) {
	cache := &asgCache{
		registeredAsgs:     make([]Asg, 0),
		instanceToAsg:      make(map[azureRef]Asg),
		notInRegisteredAsg: make(map[azureRef]bool),
		interrupt:          make(chan struct{}),
	}

	go wait.Until(func() {
		cache.mutex.Lock()
		defer cache.mutex.Unlock()
		if err := cache.regenerate(); err != nil {
			glog.Errorf("Error while regenerating Asg cache: %v", err)
		}
	}, time.Hour, cache.interrupt)

	return cache, nil
}

// Register registers a node group if it hasn't been registered.
func (m *asgCache) Register(asg Asg) bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for i := range m.registeredAsgs {
		if existing := m.registeredAsgs[i]; existing.Id() == asg.Id() {
			if reflect.DeepEqual(existing, asg) {
				return false
			}

			m.registeredAsgs[i] = asg
			glog.V(4).Infof("ASG %q updated", asg.Id())
			m.invalidateUnownedInstanceCache()
			return true
		}
	}

	glog.V(4).Infof("Registering ASG %q", asg.Id())
	m.registeredAsgs = append(m.registeredAsgs, asg)
	m.invalidateUnownedInstanceCache()
	return true
}

func (m *asgCache) invalidateUnownedInstanceCache() {
	glog.V(4).Info("Invalidating unowned instance cache")
	m.notInRegisteredAsg = make(map[azureRef]bool)
}

// Unregister ASG. Returns true if the ASG was unregistered.
func (m *asgCache) Unregister(asg Asg) bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	updated := make([]Asg, 0, len(m.registeredAsgs))
	changed := false
	for _, existing := range m.registeredAsgs {
		if existing.Id() == asg.Id() {
			glog.V(1).Infof("Unregistered ASG %s", asg.Id())
			changed = true
			continue
		}
		updated = append(updated, existing)
	}
	m.registeredAsgs = updated
	return changed
}

func (m *asgCache) get() []Asg {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	return m.registeredAsgs
}

// FindForInstance returns Asg of the given Instance
func (m *asgCache) FindForInstance(instance *azureRef) (Asg, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.notInRegisteredAsg[*instance] {
		// We already know we don't own this instance. Return early and avoid
		// additional calls.
		return nil, nil
	}

	if asg, found := m.instanceToAsg[*instance]; found {
		return asg, nil
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

// Cleanup closes the channel to signal the go routine to stop that is handling the cache
func (m *asgCache) Cleanup() {
	close(m.interrupt)
}

func (m *asgCache) regenerate() error {
	newCache := make(map[azureRef]Asg)

	for _, nsg := range m.registeredAsgs {
		instances, err := nsg.Nodes()
		if err != nil {
			return err
		}

		for _, instance := range instances {
			// Convert to lower because instance.ID is in different in different API calls (e.g. GET and LIST).
			ref := azureRef{Name: strings.ToLower(instance)}
			newCache[ref] = nsg
		}
	}

	m.instanceToAsg = newCache
	return nil
}
