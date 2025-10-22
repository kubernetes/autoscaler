package proxmox

import (
	"sync"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
)

type InstanceCache struct {
	mu        sync.RWMutex
	instances []cloudprovider.Instance
}

func NewInstanceCache() *InstanceCache {
	return &InstanceCache{}
}

func (ic *InstanceCache) Len() int {
	ic.mu.RLock()
	defer ic.mu.RUnlock()
	return len(ic.instances)
}

func (ic *InstanceCache) Pop() *cloudprovider.Instance {
	var instance cloudprovider.Instance
	ic.mu.Lock()
	defer ic.mu.Unlock()

	if ic.Len() > 0 {
		ic.instances = append(ic.instances)

		instance, ic.instances = ic.instances[len(ic.instances)-1], ic.instances[:len(ic.instances)-1]
	}

	return &instance
}

func (ic *InstanceCache) Add(instance ...cloudprovider.Instance) {
	ic.mu.Lock()
	defer ic.mu.Unlock()

	ic.instances = append(ic.instances, instance...)
}

func (ic *InstanceCache) Remove(instanceID string) {
	ic.mu.Lock()
	defer ic.mu.Unlock()

	for i, instance := range ic.instances {
		if instance.Id == instanceID {
			ic.instances = append(ic.instances[:i], ic.instances[i+1:]...)
			break
		}
	}
}

func (ic *InstanceCache) Clear() {
	ic.mu.Lock()
	defer ic.mu.Unlock()

	ic.instances = []cloudprovider.Instance{}
}

func (ic *InstanceCache) Items() []cloudprovider.Instance {
	ic.mu.RLock()
	defer ic.mu.RUnlock()
	// Return a copy of the instances slice
	instances := make([]cloudprovider.Instance, len(ic.instances))
	copy(instances, ic.instances)
	return instances
}
