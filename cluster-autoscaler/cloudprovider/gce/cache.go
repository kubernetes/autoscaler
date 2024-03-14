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

package gce

import (
	"reflect"
	"sync"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"

	gce "google.golang.org/api/compute/v1"
	klog "k8s.io/klog/v2"
)

// MachineTypeKey is used to identify MachineType.
type MachineTypeKey struct {
	Zone            string
	MachineTypeName string
}

// InstanceTemplateName is used to store the name, and
// whether or not the instance template is regional
type InstanceTemplateName struct {
	Name     string
	Regional bool
}

// GceCache is used for caching cluster resources state.
//
// It is needed to:
// - keep track of MIGs in the cluster,
// - keep track of instances in the cluster,
// - keep track of MIGs to instances mapping,
// - keep track of MIGs configuration such as target size and basename,
// - keep track of resource limiters and machine types,
// - limit repetitive GCE API calls.
//
// Cache keeps these values and gives access to getters, setters and
// invalidators all guarded with mutex. Cache does not refresh the data by
// itself - it just provides an interface enabling access to this data.
//
// The caches maintained here differ in terms of expected lifetime. Mig instance,
// basename, target size and instance template name caches need to be refreshed
// each loop to guarantee their freshness. Other values like Migs and instance
// templates are cached for longer periods of time and are refreshed either
// periodically or in response to detecting cluster state changes.
type GceCache struct {
	cacheMutex sync.Mutex

	// Cache content.
	migs                             map[GceRef]Mig
	instances                        map[GceRef][]GceInstance
	instancesUpdateTime              map[GceRef]time.Time
	instancesToMig                   map[GceRef]GceRef
	instancesFromUnknownMig          map[GceRef]bool
	resourceLimiter                  *cloudprovider.ResourceLimiter
	autoscalingOptionsCache          map[GceRef]map[string]string
	machinesCache                    map[MachineTypeKey]MachineType
	migTargetSizeCache               map[GceRef]int64
	migBaseNameCache                 map[GceRef]string
	listManagedInstancesResultsCache map[GceRef]string
	instanceTemplateNameCache        map[GceRef]InstanceTemplateName
	instanceTemplatesCache           map[GceRef]*gce.InstanceTemplate
	kubeEnvCache                     map[GceRef]KubeEnv
}

// NewGceCache creates empty GceCache.
func NewGceCache() *GceCache {
	return &GceCache{
		migs:                             map[GceRef]Mig{},
		instances:                        map[GceRef][]GceInstance{},
		instancesUpdateTime:              map[GceRef]time.Time{},
		instancesToMig:                   map[GceRef]GceRef{},
		instancesFromUnknownMig:          map[GceRef]bool{},
		autoscalingOptionsCache:          map[GceRef]map[string]string{},
		machinesCache:                    map[MachineTypeKey]MachineType{},
		migTargetSizeCache:               map[GceRef]int64{},
		migBaseNameCache:                 map[GceRef]string{},
		listManagedInstancesResultsCache: map[GceRef]string{},
		instanceTemplateNameCache:        map[GceRef]InstanceTemplateName{},
		instanceTemplatesCache:           map[GceRef]*gce.InstanceTemplate{},
		kubeEnvCache:                     map[GceRef]KubeEnv{},
	}
}

// RegisterMig returns true if the node group wasn't in cache before, or its config was updated.
func (gc *GceCache) RegisterMig(newMig Mig) bool {
	gc.cacheMutex.Lock()
	defer gc.cacheMutex.Unlock()

	oldMig, found := gc.migs[newMig.GceRef()]
	if found {
		if !reflect.DeepEqual(oldMig, newMig) {
			gc.migs[newMig.GceRef()] = newMig
			klog.V(4).Infof("Updated Mig %s", newMig.GceRef().String())
			return true
		}
		return false
	}

	klog.V(1).Infof("Registering %s", newMig.GceRef().String())
	gc.migs[newMig.GceRef()] = newMig
	return true
}

// UnregisterMig returns true if the node group has been removed, and false if it was already missing from cache.
func (gc *GceCache) UnregisterMig(toBeRemoved Mig) bool {
	gc.cacheMutex.Lock()
	defer gc.cacheMutex.Unlock()

	_, found := gc.migs[toBeRemoved.GceRef()]
	if found {
		klog.V(1).Infof("Unregistered Mig %s", toBeRemoved.GceRef().String())
		delete(gc.migs, toBeRemoved.GceRef())
		gc.removeMigInstances(toBeRemoved.GceRef())
		return true
	}
	return false
}

// GetMig returns a MIG for a given GceRef.
func (gc *GceCache) GetMig(migRef GceRef) (Mig, bool) {
	gc.cacheMutex.Lock()
	defer gc.cacheMutex.Unlock()

	mig, found := gc.migs[migRef]
	return mig, found
}

// GetMigs returns a copy of migs list.
func (gc *GceCache) GetMigs() []Mig {
	gc.cacheMutex.Lock()
	defer gc.cacheMutex.Unlock()

	migs := make([]Mig, 0, len(gc.migs))
	for _, mig := range gc.migs {
		migs = append(migs, mig)
	}
	return migs
}

// GetMigInstances returns the cached instances for a given MIG GceRef
func (gc *GceCache) GetMigInstances(migRef GceRef) ([]GceInstance, bool) {
	gc.cacheMutex.Lock()
	defer gc.cacheMutex.Unlock()

	instances, found := gc.instances[migRef]
	if found {
		klog.V(5).Infof("Instances cache hit for %s", migRef)
	}
	return append([]GceInstance{}, instances...), found
}

// GetMigInstancesUpdateTime returns the timestamp when the cached instances
// were updated for a given MIG GceRef
func (gc *GceCache) GetMigInstancesUpdateTime(migRef GceRef) (time.Time, bool) {
	gc.cacheMutex.Lock()
	defer gc.cacheMutex.Unlock()

	timestamp, found := gc.instancesUpdateTime[migRef]
	if found {
		klog.V(5).Infof("Instances update time cache hit for %s", migRef)
	}
	return timestamp, found
}

// GetMigForInstance returns the cached MIG for instance GceRef
func (gc *GceCache) GetMigForInstance(instanceRef GceRef) (GceRef, bool) {
	gc.cacheMutex.Lock()
	defer gc.cacheMutex.Unlock()

	migRef, found := gc.instancesToMig[instanceRef]
	if found {
		klog.V(5).Infof("MIG cache hit for %s", instanceRef)
	}
	return migRef, found
}

// IsMigUnknownForInstance checks if MIG was marked as unknown for instance, meaning that
// a Mig to which this instance should belong does not list it as one of its instances.
func (gc *GceCache) IsMigUnknownForInstance(instanceRef GceRef) bool {
	gc.cacheMutex.Lock()
	defer gc.cacheMutex.Unlock()

	unknown, _ := gc.instancesFromUnknownMig[instanceRef]
	if unknown {
		klog.V(5).Infof("Unknown MIG cache hit for %s", instanceRef)
	}
	return unknown
}

// SetMigInstances sets instances for a given Mig ref
func (gc *GceCache) SetMigInstances(migRef GceRef, instances []GceInstance, timeNow time.Time) error {
	gc.cacheMutex.Lock()
	defer gc.cacheMutex.Unlock()

	gc.removeMigInstances(migRef)
	gc.instances[migRef] = append([]GceInstance{}, instances...)
	gc.instancesUpdateTime[migRef] = timeNow
	for _, instance := range instances {
		instanceRef, err := GceRefFromProviderId(instance.Id)
		if err != nil {
			return err
		}
		delete(gc.instancesFromUnknownMig, instanceRef)
		gc.instancesToMig[instanceRef] = migRef
	}
	return nil
}

// MarkInstanceMigUnknown sets instance MIG to unknown, meaning that a Mig to which
// this instance should belong does not list it as one of its instances.
func (gc *GceCache) MarkInstanceMigUnknown(instanceRef GceRef) {
	gc.cacheMutex.Lock()
	defer gc.cacheMutex.Unlock()

	gc.instancesFromUnknownMig[instanceRef] = true
}

// InvalidateAllMigInstances clears the mig instances cache
func (gc *GceCache) InvalidateAllMigInstances() {
	gc.cacheMutex.Lock()
	defer gc.cacheMutex.Unlock()

	klog.V(5).Infof("Mig instances cache invalidated")
	gc.instances = make(map[GceRef][]GceInstance)
	gc.instancesUpdateTime = make(map[GceRef]time.Time)
}

// InvalidateInstancesToMig clears the instance to mig mapping for a GceRef
func (gc *GceCache) InvalidateInstancesToMig(migRef GceRef) {
	gc.cacheMutex.Lock()
	defer gc.cacheMutex.Unlock()

	klog.V(5).Infof("Instances to mig cache invalidated for %s", migRef)
	gc.removeMigInstances(migRef)
}

// InvalidateAllInstancesToMig clears the instance to mig cache
func (gc *GceCache) InvalidateAllInstancesToMig() {
	gc.cacheMutex.Lock()
	defer gc.cacheMutex.Unlock()

	klog.V(5).Infof("Instances to migs cache invalidated")
	gc.instancesToMig = make(map[GceRef]GceRef)
	gc.instancesFromUnknownMig = make(map[GceRef]bool)
}

func (gc *GceCache) removeMigInstances(migRef GceRef) {
	for instanceRef, instanceMigRef := range gc.instancesToMig {
		if migRef == instanceMigRef {
			delete(gc.instancesToMig, instanceRef)
			delete(gc.instancesFromUnknownMig, instanceRef)
		}
	}
}

// SetAutoscalingOptions stores autoscaling options strings obtained from IT.
func (gc *GceCache) SetAutoscalingOptions(ref GceRef, options map[string]string) {
	gc.cacheMutex.Lock()
	defer gc.cacheMutex.Unlock()
	gc.autoscalingOptionsCache[ref] = options
}

// GetAutoscalingOptions return autoscaling options strings obtained from IT.
func (gc *GceCache) GetAutoscalingOptions(ref GceRef) map[string]string {
	gc.cacheMutex.Lock()
	defer gc.cacheMutex.Unlock()
	return gc.autoscalingOptionsCache[ref]
}

// SetResourceLimiter sets resource limiter.
func (gc *GceCache) SetResourceLimiter(resourceLimiter *cloudprovider.ResourceLimiter) {
	gc.cacheMutex.Lock()
	defer gc.cacheMutex.Unlock()

	gc.resourceLimiter = resourceLimiter
}

// GetResourceLimiter returns resource limiter.
func (gc *GceCache) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	gc.cacheMutex.Lock()
	defer gc.cacheMutex.Unlock()

	return gc.resourceLimiter, nil
}

// GetMigTargetSize returns the cached targetSize for a GceRef
func (gc *GceCache) GetMigTargetSize(ref GceRef) (int64, bool) {
	gc.cacheMutex.Lock()
	defer gc.cacheMutex.Unlock()

	size, found := gc.migTargetSizeCache[ref]
	if found {
		klog.V(5).Infof("Target size cache hit for %s", ref)
	}
	return size, found
}

// SetMigTargetSize sets targetSize for a GceRef
func (gc *GceCache) SetMigTargetSize(ref GceRef, size int64) {
	gc.cacheMutex.Lock()
	defer gc.cacheMutex.Unlock()

	gc.migTargetSizeCache[ref] = size
}

// InvalidateMigTargetSize clears the target size cache
func (gc *GceCache) InvalidateMigTargetSize(ref GceRef) {
	gc.cacheMutex.Lock()
	defer gc.cacheMutex.Unlock()

	if _, found := gc.migTargetSizeCache[ref]; found {
		klog.V(5).Infof("Target size cache invalidated for %s", ref)
		delete(gc.migTargetSizeCache, ref)
	}
}

// InvalidateAllMigTargetSizes clears the target size cache
func (gc *GceCache) InvalidateAllMigTargetSizes() {
	gc.cacheMutex.Lock()
	defer gc.cacheMutex.Unlock()

	klog.V(5).Infof("Target size cache invalidated")
	gc.migTargetSizeCache = map[GceRef]int64{}
}

// GetMigInstanceTemplateName returns the cached instance template ref for a mig GceRef
func (gc *GceCache) GetMigInstanceTemplateName(ref GceRef) (InstanceTemplateName, bool) {
	gc.cacheMutex.Lock()
	defer gc.cacheMutex.Unlock()

	instanceTemplateName, found := gc.instanceTemplateNameCache[ref]
	if found {
		klog.V(5).Infof("Instance template names cache hit for %s", ref)
	}
	return instanceTemplateName, found
}

// SetMigInstanceTemplateName sets instance template ref for a mig GceRef
func (gc *GceCache) SetMigInstanceTemplateName(ref GceRef, instanceTemplateName InstanceTemplateName) {
	gc.cacheMutex.Lock()
	defer gc.cacheMutex.Unlock()

	gc.instanceTemplateNameCache[ref] = instanceTemplateName
}

// InvalidateMigInstanceTemplateName clears the instance template ref cache for a mig GceRef
func (gc *GceCache) InvalidateMigInstanceTemplateName(ref GceRef) {
	gc.cacheMutex.Lock()
	defer gc.cacheMutex.Unlock()

	if _, found := gc.instanceTemplateNameCache[ref]; found {
		klog.V(5).Infof("Instance template names cache invalidated for %s", ref)
		delete(gc.instanceTemplateNameCache, ref)
	}
}

// InvalidateAllMigInstanceTemplateNames clears the instance template ref cache
func (gc *GceCache) InvalidateAllMigInstanceTemplateNames() {
	gc.cacheMutex.Lock()
	defer gc.cacheMutex.Unlock()

	klog.V(5).Infof("Instance template names cache invalidated")
	gc.instanceTemplateNameCache = map[GceRef]InstanceTemplateName{}
}

// GetMigInstanceTemplate returns the cached gce.InstanceTemplate for a mig GceRef
func (gc *GceCache) GetMigInstanceTemplate(ref GceRef) (*gce.InstanceTemplate, bool) {
	gc.cacheMutex.Lock()
	defer gc.cacheMutex.Unlock()

	instanceTemplate, found := gc.instanceTemplatesCache[ref]
	if found {
		klog.V(5).Infof("Instance template cache hit for %s", ref)
	}
	return instanceTemplate, found
}

// SetMigInstanceTemplate sets gce.InstanceTemplate for a mig GceRef
func (gc *GceCache) SetMigInstanceTemplate(ref GceRef, instanceTemplate *gce.InstanceTemplate) {
	gc.cacheMutex.Lock()
	defer gc.cacheMutex.Unlock()

	gc.instanceTemplatesCache[ref] = instanceTemplate
}

// InvalidateMigInstanceTemplate clears the instance template cache for a mig GceRef
func (gc *GceCache) InvalidateMigInstanceTemplate(ref GceRef) {
	gc.cacheMutex.Lock()
	defer gc.cacheMutex.Unlock()

	if _, found := gc.instanceTemplatesCache[ref]; found {
		klog.V(5).Infof("Instance template cache invalidated for %s", ref)
		delete(gc.instanceTemplatesCache, ref)
	}
}

// InvalidateAllMigInstanceTemplates clears the instance template cache
func (gc *GceCache) InvalidateAllMigInstanceTemplates() {
	gc.cacheMutex.Lock()
	defer gc.cacheMutex.Unlock()

	klog.V(5).Infof("Instance template cache invalidated")
	gc.instanceTemplatesCache = map[GceRef]*gce.InstanceTemplate{}
}

// GetMigKubeEnv returns the cached KubeEnv for a mig GceRef
func (gc *GceCache) GetMigKubeEnv(ref GceRef) (KubeEnv, bool) {
	gc.cacheMutex.Lock()
	defer gc.cacheMutex.Unlock()

	kubeEnv, found := gc.kubeEnvCache[ref]
	if found {
		klog.V(5).Infof("Kube-env cache hit for %s", ref)
	}
	return kubeEnv, found
}

// SetMigKubeEnv sets KubeEnv for a mig GceRef
func (gc *GceCache) SetMigKubeEnv(ref GceRef, kubeEnv KubeEnv) {
	gc.cacheMutex.Lock()
	defer gc.cacheMutex.Unlock()

	gc.kubeEnvCache[ref] = kubeEnv
}

// InvalidateMigKubeEnv clears the kube-env cache for a mig GceRef
func (gc *GceCache) InvalidateMigKubeEnv(ref GceRef) {
	gc.cacheMutex.Lock()
	defer gc.cacheMutex.Unlock()

	if _, found := gc.kubeEnvCache[ref]; found {
		klog.V(5).Infof("Kube-env cache invalidated for %s", ref)
		delete(gc.kubeEnvCache, ref)
	}
}

// InvalidateAllMigKubeEnvs clears the kube-env cache
func (gc *GceCache) InvalidateAllMigKubeEnvs() {
	gc.cacheMutex.Lock()
	defer gc.cacheMutex.Unlock()

	klog.V(5).Infof("Kube-env cache invalidated")
	gc.kubeEnvCache = map[GceRef]KubeEnv{}
}

// GetMachine retrieves machine type from cache under lock.
func (gc *GceCache) GetMachine(machineTypeName string, zone string) (MachineType, bool) {
	gc.cacheMutex.Lock()
	defer gc.cacheMutex.Unlock()

	m, found := gc.machinesCache[MachineTypeKey{zone, machineTypeName}]
	return m, found
}

// AddMachine adds machine to cache under lock.
func (gc *GceCache) AddMachine(machineType MachineType, zone string) {
	gc.cacheMutex.Lock()
	defer gc.cacheMutex.Unlock()

	gc.machinesCache[MachineTypeKey{zone, machineType.Name}] = machineType
}

// SetMachines sets the machines cache under lock.
func (gc *GceCache) SetMachines(machinesCache map[MachineTypeKey]MachineType) {
	gc.cacheMutex.Lock()
	defer gc.cacheMutex.Unlock()

	gc.machinesCache = map[MachineTypeKey]MachineType{}
	for k, v := range machinesCache {
		gc.machinesCache[k] = v
	}
}

// InvalidateAllMachines invalidates the machines cache under lock.
func (gc *GceCache) InvalidateAllMachines() {
	gc.cacheMutex.Lock()
	defer gc.cacheMutex.Unlock()

	gc.machinesCache = map[MachineTypeKey]MachineType{}
}

// SetMigBasename sets basename for given mig in cache
func (gc *GceCache) SetMigBasename(migRef GceRef, basename string) {
	gc.cacheMutex.Lock()
	defer gc.cacheMutex.Unlock()
	gc.migBaseNameCache[migRef] = basename
}

// GetMigBasename get basename for given mig from cache.
func (gc *GceCache) GetMigBasename(migRef GceRef) (basename string, found bool) {
	gc.cacheMutex.Lock()
	defer gc.cacheMutex.Unlock()
	basename, found = gc.migBaseNameCache[migRef]
	return
}

// InvalidateMigBasename invalidates basename entry for given mig.
func (gc *GceCache) InvalidateMigBasename(migRef GceRef) {
	gc.cacheMutex.Lock()
	defer gc.cacheMutex.Unlock()
	delete(gc.migBaseNameCache, migRef)
}

// InvalidateAllMigBasenames invalidates all basename entries.
func (gc *GceCache) InvalidateAllMigBasenames() {
	gc.cacheMutex.Lock()
	defer gc.cacheMutex.Unlock()
	gc.migBaseNameCache = make(map[GceRef]string)
}

// SetListManagedInstancesResults sets listManagedInstancesResults for a given mig in cache
func (gc *GceCache) SetListManagedInstancesResults(migRef GceRef, listManagedInstancesResults string) {
	gc.cacheMutex.Lock()
	defer gc.cacheMutex.Unlock()
	gc.listManagedInstancesResultsCache[migRef] = listManagedInstancesResults
}

// GetListManagedInstancesResults gets listManagedInstancesResults for a given mig from cache.
func (gc *GceCache) GetListManagedInstancesResults(migRef GceRef) (string, bool) {
	gc.cacheMutex.Lock()
	defer gc.cacheMutex.Unlock()
	listManagedInstancesResults, found := gc.listManagedInstancesResultsCache[migRef]
	return listManagedInstancesResults, found
}

// InvalidateAllListManagedInstancesResults invalidates all listManagedInstancesResults entries.
func (gc *GceCache) InvalidateAllListManagedInstancesResults() {
	gc.cacheMutex.Lock()
	defer gc.cacheMutex.Unlock()
	gc.listManagedInstancesResultsCache = make(map[GceRef]string)
}
