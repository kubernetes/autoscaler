/*
Copyright 2021 The Kubernetes Authors.

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
	"context"
	"fmt"
	"net/url"
	"path"
	"strings"
	"sync"
	"time"

	gce "google.golang.org/api/compute/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/metrics"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

// MigInfoProvider allows obtaining information about MIGs
type MigInfoProvider interface {
	// GetMigInstances returns instances for a given MIG ref
	GetMigInstances(migRef GceRef) ([]GceInstance, error)
	// GetMigForInstance returns MIG ref for a given instance
	GetMigForInstance(instanceRef GceRef) (Mig, error)
	// RegenerateMigInstancesCache regenerates MIGs to instances mapping cache
	RegenerateMigInstancesCache() error
	// GetMigTargetSize returns target size for given MIG ref
	GetMigTargetSize(migRef GceRef) (int64, error)
	// GetMigBasename returns basename for given MIG ref
	GetMigBasename(migRef GceRef) (string, error)
	// GetMigInstanceTemplateName returns instance template name for given MIG ref
	GetMigInstanceTemplateName(migRef GceRef) (InstanceTemplateName, error)
	// GetMigInstanceTemplate returns instance template for given MIG ref
	GetMigInstanceTemplate(migRef GceRef) (*gce.InstanceTemplate, error)
	// GetMigKubeEnv returns kube-env for given MIG ref
	GetMigKubeEnv(migRef GceRef) (KubeEnv, error)
	// GetMigMachineType returns machine type used by a MIG.
	// For custom machines cpu and memory information is based on parsing
	// machine name. For standard types it's retrieved from GCE API.
	GetMigMachineType(migRef GceRef) (MachineType, error)
	// Returns the pagination behavior of the listManagedInstances API method for a given MIG ref
	GetListManagedInstancesResults(migRef GceRef) (string, error)
}

type timeProvider interface {
	Now() time.Time
}

type cachingMigInfoProvider struct {
	migInfoMutex                      sync.Mutex
	cache                             *GceCache
	migLister                         MigLister
	gceClient                         AutoscalingGceClient
	projectId                         string
	concurrentGceRefreshes            int
	migInstanceMutex                  sync.Mutex
	migInstancesMinRefreshWaitTime    time.Duration
	timeProvider                      timeProvider
	bulkGceMigInstancesListingEnabled bool
}

type realTime struct{}

func (r *realTime) Now() time.Time {
	return time.Now()
}

// NewCachingMigInfoProvider creates an instance of caching MigInfoProvider
func NewCachingMigInfoProvider(cache *GceCache, migLister MigLister, gceClient AutoscalingGceClient, projectId string, concurrentGceRefreshes int, migInstancesMinRefreshWaitTime time.Duration, bulkGceMigInstancesListingEnabled bool) MigInfoProvider {
	return &cachingMigInfoProvider{
		cache:                             cache,
		migLister:                         migLister,
		gceClient:                         gceClient,
		projectId:                         projectId,
		concurrentGceRefreshes:            concurrentGceRefreshes,
		migInstancesMinRefreshWaitTime:    migInstancesMinRefreshWaitTime,
		timeProvider:                      &realTime{},
		bulkGceMigInstancesListingEnabled: bulkGceMigInstancesListingEnabled,
	}
}

// GetMigInstances returns instances for a given MIG ref
func (c *cachingMigInfoProvider) GetMigInstances(migRef GceRef) ([]GceInstance, error) {
	instances, found := c.cache.GetMigInstances(migRef)
	if found {
		return instances, nil
	}

	err := c.fillMigInstances(migRef)
	if err != nil {
		return nil, err
	}
	instances, _ = c.cache.GetMigInstances(migRef)
	return instances, nil
}

// GetMigForInstance returns MIG ref for a given instance
func (c *cachingMigInfoProvider) GetMigForInstance(instanceRef GceRef) (Mig, error) {
	c.migInstanceMutex.Lock()
	defer c.migInstanceMutex.Unlock()

	mig, found, err := c.getCachedMigForInstance(instanceRef)
	if found {
		return mig, err
	}

	mig = c.findMigWithMatchingBasename(instanceRef)
	if mig == nil {
		return nil, nil
	}

	err = c.fillMigInstances(mig.GceRef())
	if err != nil {
		return nil, err
	}

	mig, found, err = c.getCachedMigForInstance(instanceRef)
	if !found {
		c.cache.MarkInstanceMigUnknown(instanceRef)
	}
	return mig, err
}

func (c *cachingMigInfoProvider) getCachedMigForInstance(instanceRef GceRef) (Mig, bool, error) {
	if migRef, found := c.cache.GetMigForInstance(instanceRef); found {
		mig, found := c.cache.GetMig(migRef)
		if !found {
			return nil, true, fmt.Errorf("instance %v belongs to unregistered mig %v", instanceRef, migRef)
		}
		return mig, true, nil
	} else if c.cache.IsMigUnknownForInstance(instanceRef) {
		return nil, true, nil
	}
	return nil, false, nil
}

// RegenerateMigInstancesCache regenerates MIGs to instances mapping cache
func (c *cachingMigInfoProvider) RegenerateMigInstancesCache() error {
	c.cache.InvalidateAllMigInstances()
	c.cache.InvalidateAllInstancesToMig()

	if c.bulkGceMigInstancesListingEnabled {
		return c.bulkListMigInstances()
	}

	migs := c.migLister.GetMigs()
	errors := make([]error, len(migs))
	workqueue.ParallelizeUntil(context.Background(), c.concurrentGceRefreshes, len(migs), func(piece int) {
		errors[piece] = c.fillMigInstances(migs[piece].GceRef())
	}, workqueue.WithChunkSize(c.concurrentGceRefreshes))

	for _, err := range errors {
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *cachingMigInfoProvider) bulkListMigInstances() error {
	c.cache.InvalidateMigInstancesStateCount()
	err := c.fillMigInfoCache()
	if err != nil {
		return err
	}
	instances, listErr := c.listInstancesInAllZonesWithMigs()
	migToInstances := groupInstancesToMigs(instances)
	updateErr := c.updateMigInstancesCache(migToInstances)

	if listErr != nil {
		return listErr
	}
	return updateErr
}

func (c *cachingMigInfoProvider) listInstancesInAllZonesWithMigs() ([]GceInstance, error) {
	var zones []string
	for zone := range c.listAllZonesWithMigs() {
		zones = append(zones, zone)
	}
	var allInstances []GceInstance
	errors := make([]error, len(zones))
	zoneInstances := make([][]GceInstance, len(zones))
	defer metrics.UpdateDurationFromStart(metrics.BulkListAllGceInstances, time.Now())

	workqueue.ParallelizeUntil(context.Background(), len(zones), len(zones), func(piece int) {
		zoneInstances[piece], errors[piece] = c.gceClient.FetchAllInstances(c.projectId, zones[piece], "")
	})

	for _, instances := range zoneInstances {
		allInstances = append(allInstances, instances...)
	}
	for _, err := range errors {
		if err != nil {
			return allInstances, err
		}
	}
	return allInstances, nil
}

func groupInstancesToMigs(instances []GceInstance) map[GceRef][]GceInstance {
	migToInstances := map[GceRef][]GceInstance{}
	for _, instance := range instances {
		migToInstances[instance.Igm] = append(migToInstances[instance.Igm], instance)
	}
	return migToInstances
}

func (c *cachingMigInfoProvider) isMigInstancesConsistent(mig Mig, migToInstances map[GceRef][]GceInstance) bool {
	migRef := mig.GceRef()
	instancesStateCount, found := c.cache.GetMigInstancesStateCount(migRef)
	if !found {
		return false
	}
	instanceCount := instancesStateCount[cloudprovider.InstanceRunning] + instancesStateCount[cloudprovider.InstanceCreating] + instancesStateCount[cloudprovider.InstanceDeleting]

	instances, found := migToInstances[migRef]
	if !found && instanceCount > 0 {
		return false
	}
	return instanceCount == int64(len(instances))
}

func (c *cachingMigInfoProvider) isMigCreatingOrDeletingInstances(mig Mig) bool {
	migRef := mig.GceRef()
	instancesStateCount, found := c.cache.GetMigInstancesStateCount(migRef)
	if !found {
		return false
	}
	return instancesStateCount[cloudprovider.InstanceCreating] > 0 || instancesStateCount[cloudprovider.InstanceDeleting] > 0
}

// updateMigInstancesCache updates the mig instances for each mig
func (c *cachingMigInfoProvider) updateMigInstancesCache(migToInstances map[GceRef][]GceInstance) error {
	defer metrics.UpdateDurationFromStart(metrics.BulkListMigInstances, time.Now())
	inconsistentInstancesMigsCount := 0
	defer func() {
		if inconsistentInstancesMigsCount > 0 {
			klog.Warningf("Inconsistent instances migs count: %v", inconsistentInstancesMigsCount)
		}
		metrics.UpdateInconsistentInstancesMigsCount(inconsistentInstancesMigsCount)
	}()
	var errors []error
	for _, mig := range c.migLister.GetMigs() {
		migRef := mig.GceRef()
		// If there is an inconsistency between number of instances according to instances.List
		// and number of instances according to migInstancesStateCount for the given mig, which can be due to
		// - abandoned instance
		// - missing/malformed "created-by" reference
		// we use an igm.ListInstances call as the authoritative source of instance information
		if !c.isMigInstancesConsistent(mig, migToInstances) {
			if err := c.fillMigInstances(migRef); err != nil {
				errors = append(errors, err)
			}
			inconsistentInstancesMigsCount += 1
			continue
		}

		// mig instances are re-fetched along with instance.Status.ErrorInfo for migs with
		// instances in creating or deleting state
		if c.isMigCreatingOrDeletingInstances(mig) {
			if err := c.fillMigInstances(migRef); err != nil {
				errors = append(errors, err)
			}
			continue
		}

		// for all other cases, mig instances cache is updated with the instances obtained from instance.List api
		instances := migToInstances[migRef]
		err := c.cache.SetMigInstances(migRef, instances, c.timeProvider.Now())
		if err != nil {
			errors = append(errors, err)
		}
	}
	if len(errors) > 0 {
		return errors[0]
	}
	return nil
}

func (c *cachingMigInfoProvider) findMigWithMatchingBasename(instanceRef GceRef) Mig {
	for _, mig := range c.migLister.GetMigs() {
		migRef := mig.GceRef()
		basename, err := c.GetMigBasename(migRef)
		if err == nil && migRef.Project == instanceRef.Project && migRef.Zone == instanceRef.Zone && strings.HasPrefix(instanceRef.Name, basename) {
			return mig
		}
	}
	return nil
}

func (c *cachingMigInfoProvider) fillMigInstances(migRef GceRef) error {
	if val, ok := c.cache.GetMigInstancesUpdateTime(migRef); ok {
		// do not regenerate MIG instances cache if last refresh happened recently.
		if c.timeProvider.Now().Sub(val) < c.migInstancesMinRefreshWaitTime {
			klog.V(4).Infof("Not regenerating MIG instances cache for %s, as it was refreshed in last MinRefreshWaitTime (%s).", migRef.String(), c.migInstancesMinRefreshWaitTime)
			return nil
		}
	}
	klog.V(4).Infof("Regenerating MIG instances cache for %s", migRef.String())
	instances, err := c.gceClient.FetchMigInstances(migRef)
	if err != nil {
		c.migLister.HandleMigIssue(migRef, err)
		return err
	}
	// only save information for successful calls, given the errors above may be transient.
	return c.cache.SetMigInstances(migRef, instances, c.timeProvider.Now())
}

func (c *cachingMigInfoProvider) GetMigTargetSize(migRef GceRef) (int64, error) {
	c.migInfoMutex.Lock()
	defer c.migInfoMutex.Unlock()

	targetSize, found := c.cache.GetMigTargetSize(migRef)
	if found {
		return targetSize, nil
	}

	err := c.fillMigInfoCache()
	targetSize, found = c.cache.GetMigTargetSize(migRef)
	if err == nil && found {
		return targetSize, nil
	}

	// fallback to querying for single mig
	targetSize, err = c.gceClient.FetchMigTargetSize(migRef)
	if err != nil {
		c.migLister.HandleMigIssue(migRef, err)
		return 0, err
	}
	c.cache.SetMigTargetSize(migRef, targetSize)
	return targetSize, nil
}

func (c *cachingMigInfoProvider) GetMigBasename(migRef GceRef) (string, error) {
	c.migInfoMutex.Lock()
	defer c.migInfoMutex.Unlock()

	basename, found := c.cache.GetMigBasename(migRef)
	if found {
		return basename, nil
	}

	err := c.fillMigInfoCache()
	basename, found = c.cache.GetMigBasename(migRef)
	if err == nil && found {
		return basename, nil
	}

	// fallback to querying for single mig
	basename, err = c.gceClient.FetchMigBasename(migRef)
	if err != nil {
		c.migLister.HandleMigIssue(migRef, err)
		return "", err
	}
	c.cache.SetMigBasename(migRef, basename)
	return basename, nil
}

func (c *cachingMigInfoProvider) GetMigInstanceTemplateName(migRef GceRef) (InstanceTemplateName, error) {
	c.migInfoMutex.Lock()
	defer c.migInfoMutex.Unlock()

	instanceTemplateName, found := c.cache.GetMigInstanceTemplateName(migRef)
	if found {
		return instanceTemplateName, nil
	}

	err := c.fillMigInfoCache()
	instanceTemplateName, found = c.cache.GetMigInstanceTemplateName(migRef)
	if err == nil && found {
		return instanceTemplateName, nil
	}

	// fallback to querying for single mig
	instanceTemplateName, err = c.gceClient.FetchMigTemplateName(migRef)
	if err != nil {
		c.migLister.HandleMigIssue(migRef, err)
		return InstanceTemplateName{}, err
	}
	c.cache.SetMigInstanceTemplateName(migRef, instanceTemplateName)
	return instanceTemplateName, nil
}

func (c *cachingMigInfoProvider) GetMigInstanceTemplate(migRef GceRef) (*gce.InstanceTemplate, error) {
	instanceTemplateName, err := c.GetMigInstanceTemplateName(migRef)
	if err != nil {
		return nil, err
	}

	template, found := c.cache.GetMigInstanceTemplate(migRef)
	if found && template.Name == instanceTemplateName.Name {
		return template, nil
	}

	klog.V(2).Infof("Instance template of mig %v changed to %v", migRef.Name, instanceTemplateName.Name)
	template, err = c.gceClient.FetchMigTemplate(migRef, instanceTemplateName.Name, instanceTemplateName.Regional)
	if err != nil {
		return nil, err
	}
	c.cache.SetMigInstanceTemplate(migRef, template)
	return template, nil
}

func (c *cachingMigInfoProvider) GetMigKubeEnv(migRef GceRef) (KubeEnv, error) {
	instanceTemplateName, err := c.GetMigInstanceTemplateName(migRef)
	if err != nil {
		return KubeEnv{}, err
	}

	kubeEnv, kubeEnvFound := c.cache.GetMigKubeEnv(migRef)
	if kubeEnvFound && kubeEnv.templateName == instanceTemplateName.Name {
		return kubeEnv, nil
	}

	template, err := c.GetMigInstanceTemplate(migRef)
	if err != nil {
		return KubeEnv{}, err
	}
	kubeEnv, err = ExtractKubeEnv(template)
	if err != nil {
		return KubeEnv{}, err
	}
	c.cache.SetMigKubeEnv(migRef, kubeEnv)
	return kubeEnv, nil
}

// filMigInfoCache needs to be called with migInfoMutex locked
func (c *cachingMigInfoProvider) fillMigInfoCache() error {
	var zones []string
	for zone := range c.listAllZonesWithMigs() {
		zones = append(zones, zone)
	}

	migs := make([][]*gce.InstanceGroupManager, len(zones))
	errors := make([]error, len(zones))
	workqueue.ParallelizeUntil(context.Background(), len(zones), len(zones), func(piece int) {
		migs[piece], errors[piece] = c.gceClient.FetchAllMigs(zones[piece])
	})

	failedZones := map[string]error{}
	failedZoneCount := 0
	for idx, err := range errors {
		if err != nil {
			klog.Errorf("Error listing migs from zone %v; err=%v", zones[idx], err)
			failedZones[zones[idx]] = err
			failedZoneCount++
		}
	}

	if failedZoneCount > 0 && failedZoneCount == len(zones) {
		return fmt.Errorf("%v", errors)
	}

	registeredMigRefs := c.getRegisteredMigRefs()

	for migRef := range registeredMigRefs {
		err, ok := failedZones[migRef.Zone]
		if ok {
			c.migLister.HandleMigIssue(migRef, err)
		}
	}

	for idx, zone := range zones {
		for _, zoneMig := range migs[idx] {
			zoneMigRef := GceRef{
				c.projectId,
				zone,
				zoneMig.Name,
			}

			if registeredMigRefs[zoneMigRef] {
				c.cache.SetMigTargetSize(zoneMigRef, zoneMig.TargetSize)
				c.cache.SetMigBasename(zoneMigRef, zoneMig.BaseInstanceName)
				c.cache.SetListManagedInstancesResults(zoneMigRef, zoneMig.ListManagedInstancesResults)
				c.cache.SetMigInstancesStateCount(zoneMigRef, createInstancesStateCount(zoneMig.TargetSize, zoneMig.CurrentActions))

				templateUrl, err := url.Parse(zoneMig.InstanceTemplate)
				if err == nil {
					_, templateName := path.Split(templateUrl.EscapedPath())
					regional, err := IsInstanceTemplateRegional(templateUrl.String())
					if err != nil {
						klog.Errorf("Error parsing instance template url: %v; err=%v ", templateUrl.String(), err)
					} else {
						c.cache.SetMigInstanceTemplateName(zoneMigRef, InstanceTemplateName{templateName, regional})
					}
				}
			}
		}
	}

	return nil
}

func (c *cachingMigInfoProvider) getRegisteredMigRefs() map[GceRef]bool {
	migRefs := make(map[GceRef]bool)
	for _, mig := range c.migLister.GetMigs() {
		migRefs[mig.GceRef()] = true
	}
	return migRefs
}

func (c *cachingMigInfoProvider) listAllZonesWithMigs() map[string]bool {
	zones := map[string]bool{}
	for _, mig := range c.migLister.GetMigs() {
		zones[mig.GceRef().Zone] = true
	}
	return zones
}

func (c *cachingMigInfoProvider) GetMigMachineType(migRef GceRef) (MachineType, error) {
	template, err := c.GetMigInstanceTemplate(migRef)
	if err != nil {
		return MachineType{}, err
	}
	machineName := template.Properties.MachineType
	if IsCustomMachine(machineName) {
		return NewCustomMachineType(machineName)
	}
	zone := migRef.Zone
	machine, found := c.cache.GetMachine(machineName, zone)
	if !found {
		rawMachine, err := c.gceClient.FetchMachineType(zone, machineName)
		if err != nil {
			c.migLister.HandleMigIssue(migRef, err)
			return MachineType{}, err
		}
		machine, err = NewMachineTypeFromAPI(machineName, rawMachine)
		if err != nil {
			c.migLister.HandleMigIssue(migRef, err)
			return MachineType{}, err
		}
		c.cache.AddMachine(machine, zone)
	}
	return machine, nil
}

func (c *cachingMigInfoProvider) GetListManagedInstancesResults(migRef GceRef) (string, error) {
	c.migInfoMutex.Lock()
	defer c.migInfoMutex.Unlock()

	listManagedInstancesResults, found := c.cache.GetListManagedInstancesResults(migRef)
	if found {
		return listManagedInstancesResults, nil
	}

	err := c.fillMigInfoCache()
	listManagedInstancesResults, found = c.cache.GetListManagedInstancesResults(migRef)
	if err == nil && found {
		return listManagedInstancesResults, nil
	}

	// fallback to querying for a single mig
	listManagedInstancesResults, err = c.gceClient.FetchListManagedInstancesResults(migRef)
	if err != nil {
		c.migLister.HandleMigIssue(migRef, err)
		return "", err
	}
	c.cache.SetListManagedInstancesResults(migRef, listManagedInstancesResults)
	return listManagedInstancesResults, nil
}

func createInstancesStateCount(targetSize int64, actionsSummary *gce.InstanceGroupManagerActionsSummary) map[cloudprovider.InstanceState]int64 {
	if actionsSummary == nil {
		return nil
	}
	stateCount := map[cloudprovider.InstanceState]int64{
		cloudprovider.InstanceCreating: 0,
		cloudprovider.InstanceDeleting: 0,
		cloudprovider.InstanceRunning:  0,
	}
	stateCount[getInstanceState("ABANDONING")] += actionsSummary.Abandoning
	stateCount[getInstanceState("CREATING")] += actionsSummary.Creating
	stateCount[getInstanceState("CREATING_WITHOUT_RETRIES")] += actionsSummary.CreatingWithoutRetries
	stateCount[getInstanceState("DELETING")] += actionsSummary.Deleting
	stateCount[getInstanceState("RECREATING")] += actionsSummary.Recreating
	stateCount[cloudprovider.InstanceRunning] = targetSize - stateCount[cloudprovider.InstanceCreating]
	return stateCount
}
