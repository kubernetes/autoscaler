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
	"path"
	"regexp"
	"strings"
	"sync"
	"time"

	gce "google.golang.org/api/compute/v1"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
	"sigs.k8s.io/cluster-autoscaler/pkg/cloudprovider"
	"sigs.k8s.io/cluster-autoscaler/pkg/metrics"
)

// MigInfoProvider allows obtaining information about MIGs
type MigInfoProvider interface {
	// GetMigInstances returns instances for a given MIG ref
	GetMigInstances(ctx context.Context, migRef GceRef) ([]GceInstance, error)
	// GetMigForInstance returns MIG ref for a given instance
	GetMigForInstance(ctx context.Context, instanceRef GceRef) (Mig, error)
	// RegenerateMigInstancesCache regenerates MIGs to instances mapping cache
	RegenerateMigInstancesCache(context.Context) error
	// GetMigTargetSize returns target size for given MIG ref
	GetMigTargetSize(ctx context.Context, migRef GceRef) (int64, error)
	// GetMigBasename returns basename for given MIG ref
	GetMigBasename(ctx context.Context, migRef GceRef) (string, error)
	// GetMigInstanceTemplateName returns instance template name for given MIG ref
	GetMigInstanceTemplateName(ctx context.Context, migRef GceRef) (InstanceTemplateName, error)
	// GetMigInstanceTemplate returns instance template for given MIG ref
	GetMigInstanceTemplate(ctx context.Context, migRef GceRef) (*gce.InstanceTemplate, error)
	// GetMigKubeEnv returns kube-env for given MIG ref
	GetMigKubeEnv(ctx context.Context, migRef GceRef) (KubeEnv, error)
	// GetMigMachineType returns machine type used by a MIG.
	// For custom machines cpu and memory information is based on parsing
	// machine name. For standard types it's retrieved from GCE API.
	GetMigMachineType(ctx context.Context, migRef GceRef) (MachineType, error)
	// Returns the pagination behavior of the listManagedInstances Results API method for a given MIG ref
	GetListManagedInstancesResults(ctx context.Context, migRef GceRef) (string, error)
	// GetMigIsStable returns whether given MIG is stable. A stable state means that: none of the instances in the managed instance group is currently undergoing any type of change (for example, creation, restart, or deletion); no future changes are scheduled for instances in the managed instance group; and the managed instance group itself is not being modified.
	GetMigIsStable(ctx context.Context, migRef GceRef) (bool, error)
	// RefreshMigInfo updates the cached information for a specific MIG without rebuilding the full zone cache
	RefreshMigInfo(ctx context.Context, migRef GceRef) error
}

type timeProvider interface {
	Now() time.Time
}

var (
	// Compile a regular expression to find the text between "projects/" and the next "/".
	migProjectSelfLinkRe = regexp.MustCompile(`projects/([^/]+)`)
)

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
	multiProjectCachingEnabled        bool
}

type realTime struct{}

func (r *realTime) Now() time.Time {
	return time.Now()
}

// NewCachingMigInfoProvider creates an instance of caching MigInfoProvider
func NewCachingMigInfoProvider(cache *GceCache, migLister MigLister, gceClient AutoscalingGceClient, projectId string, concurrentGceRefreshes int, migInstancesMinRefreshWaitTime time.Duration, bulkGceMigInstancesListingEnabled bool, multiProjectCachingEnabled bool) MigInfoProvider {
	return &cachingMigInfoProvider{
		cache:                             cache,
		migLister:                         migLister,
		gceClient:                         gceClient,
		projectId:                         projectId,
		concurrentGceRefreshes:            concurrentGceRefreshes,
		migInstancesMinRefreshWaitTime:    migInstancesMinRefreshWaitTime,
		timeProvider:                      &realTime{},
		bulkGceMigInstancesListingEnabled: bulkGceMigInstancesListingEnabled,
		multiProjectCachingEnabled:        multiProjectCachingEnabled,
	}
}

// GetMigInstances returns instances for a given MIG ref
func (c *cachingMigInfoProvider) GetMigInstances(ctx context.Context, migRef GceRef) ([]GceInstance, error) {
	instances, found := c.cache.GetMigInstances(ctx, migRef)
	if found {
		return instances, nil
	}

	// MIG is not in the cache.
	err := c.fillMigInstances(ctx, migRef)
	if err != nil {
		return nil, err
	}
	instances, _ = c.cache.GetMigInstances(ctx, migRef)
	return instances, nil
}

// GetMigForInstance returns MIG ref for a given instance
func (c *cachingMigInfoProvider) GetMigForInstance(ctx context.Context, instanceRef GceRef) (Mig, error) {
	c.migInstanceMutex.Lock()
	defer c.migInstanceMutex.Unlock()

	mig, found, err := c.getCachedMigForInstance(ctx, instanceRef)
	if found {
		return mig, err
	}

	mig = c.findMigWithMatchingBasename(ctx, instanceRef)
	if mig == nil {
		return nil, nil
	}

	// Cache is cleared every loop.
	// If it's not empty, it's been refreshed this loop, and we don't want to refresh it again.
	if !c.cache.IsMigInstancesCacheEmpty(mig.GceRef()) {
		c.cache.MarkInstanceMigUnknown(instanceRef)
		return nil, nil
	}

	err = c.fillMigInstances(ctx, mig.GceRef())
	if err != nil {
		return nil, err
	}
	// Check in the cache again after it's been refilled
	mig, found, err = c.getCachedMigForInstance(ctx, instanceRef)
	if !found {
		c.cache.MarkInstanceMigUnknown(instanceRef)
	}
	return mig, err

}

func (c *cachingMigInfoProvider) getCachedMigForInstance(ctx context.Context, instanceRef GceRef) (Mig, bool, error) {
	if migRef, found := c.cache.GetMigForInstance(ctx, instanceRef); found {
		mig, found := c.cache.GetMig(migRef)
		if !found {
			return nil, true, fmt.Errorf("instance %v belongs to unregistered mig %v", instanceRef, migRef)
		}
		return mig, true, nil
	} else if c.cache.IsMigUnknownForInstance(ctx, instanceRef) {
		return nil, true, nil
	}
	return nil, false, nil
}

// RegenerateMigInstancesCache regenerates MIGs to instances mapping cache
func (c *cachingMigInfoProvider) RegenerateMigInstancesCache(ctx context.Context) error {
	c.cache.InvalidateAllMigInstances(ctx)
	c.cache.InvalidateAllInstancesToMig(ctx)

	if c.bulkGceMigInstancesListingEnabled {
		return c.bulkListMigInstances(ctx)
	}

	migs := c.migLister.GetMigs()
	errors := make([]error, len(migs))
	workqueue.ParallelizeUntil(context.Background(), c.concurrentGceRefreshes, len(migs), func(piece int) {
		errors[piece] = c.fillMigInstances(ctx, migs[piece].GceRef())
	}, workqueue.WithChunkSize(c.concurrentGceRefreshes))

	for _, err := range errors {
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *cachingMigInfoProvider) bulkListMigInstances(ctx context.Context) error {
	c.cache.InvalidateMigInstancesStateCount()
	err := c.fillMigInfoCache(ctx)
	if err != nil {
		return err
	}
	instances, listErr := c.listInstancesInAllZonesWithMigs(ctx)
	migToInstances := groupInstancesToMigs(instances)
	updateErr := c.updateMigInstancesCache(ctx, migToInstances)

	if listErr != nil {
		return listErr
	}
	return updateErr
}

func (c *cachingMigInfoProvider) listInstancesInAllZonesWithMigs(ctx context.Context) ([]GceInstance, error) {
	var zones []string
	for zone := range c.listAllZonesWithMigs() {
		zones = append(zones, zone)
	}
	var allInstances []GceInstance
	errors := make([]error, len(zones))
	zoneInstances := make([][]GceInstance, len(zones))
	defer metrics.UpdateDurationFromStart(ctx, metrics.BulkListAllGceInstances, time.Now())

	workqueue.ParallelizeUntil(context.Background(), len(zones), len(zones), func(piece int) {
		zoneInstances[piece], errors[piece] = c.gceClient.FetchAllInstances(ctx, c.projectId, zones[piece], "")
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
func (c *cachingMigInfoProvider) updateMigInstancesCache(ctx context.Context, migToInstances map[GceRef][]GceInstance) error {
	logger := klog.FromContext(ctx)
	defer metrics.UpdateDurationFromStart(ctx, metrics.BulkListMigInstances, time.Now())
	inconsistentInstancesMigsCount := 0
	defer func() {
		if inconsistentInstancesMigsCount > 0 {
			logger.Info("Recording inconsistent instances migs count", "migsCount", inconsistentInstancesMigsCount)
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
			if err := c.fillMigInstances(ctx, migRef); err != nil {
				errors = append(errors, err)
			}
			inconsistentInstancesMigsCount += 1
			continue
		}

		// mig instances are re-fetched along with instance.Status.ErrorInfo for migs with
		// instances in creating or deleting state
		if c.isMigCreatingOrDeletingInstances(mig) {
			if err := c.fillMigInstances(ctx, migRef); err != nil {
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

func (c *cachingMigInfoProvider) findMigWithMatchingBasename(ctx context.Context, instanceRef GceRef) Mig {
	for _, mig := range c.migLister.GetMigs() {
		migRef := mig.GceRef()
		basename, err := c.GetMigBasename(ctx, migRef)
		if err == nil && migRef.Project == instanceRef.Project && migRef.Zone == instanceRef.Zone && strings.HasPrefix(instanceRef.Name, basename) {
			return mig
		}
	}
	return nil
}

func (c *cachingMigInfoProvider) fillMigInstances(ctx context.Context, migRef GceRef) error {
	logger := klog.FromContext(ctx)
	if val, ok := c.cache.GetMigInstancesUpdateTime(ctx, migRef); ok {
		// do not regenerate MIG instances cache if last refresh happened recently.
		if c.timeProvider.Now().Sub(val) < c.migInstancesMinRefreshWaitTime {
			logger.V(4).Info("Not regenerating MIG instances cache, as it was refreshed in last MinRefreshWaitTime", "mig", migRef.String(), "minRefreshWaitTime", c.migInstancesMinRefreshWaitTime)
			return nil
		}
	}
	logger.V(4).Info("Regenerating MIG instances cache", "mig", migRef.String())
	instances, err := c.gceClient.FetchMigInstances(ctx, migRef)
	if err != nil {
		c.migLister.HandleMigIssue(migRef, err)
		return err
	}
	// only save information for successful calls, given the errors above may be transient.
	return c.cache.SetMigInstances(migRef, instances, c.timeProvider.Now())
}

func (c *cachingMigInfoProvider) GetMigTargetSize(ctx context.Context, migRef GceRef) (int64, error) {
	c.migInfoMutex.Lock()
	defer c.migInfoMutex.Unlock()

	targetSize, found := c.cache.GetMigTargetSize(ctx, migRef)
	if found {
		return targetSize, nil
	}

	var err error
	if c.cache.IsMigTargetSizeCacheEmpty() {
		// Cache is cold after Refresh() -- list all MIGs and populate the cache.
		err = c.fillMigInfoCache(ctx)
	}

	targetSize, found = c.cache.GetMigTargetSize(ctx, migRef)
	if found && err == nil {
		return targetSize, nil
	}

	// We get here in one of 3 cases:
	//  * InvalidateMigTargetSize was called for this specific mig, so it's not found in cache
	//  * fillMigInfoCache returned an error
	//  * MIG not found
	err = c.fillSingleMigInfo(ctx, migRef)
	if err != nil {
		return 0, err
	}
	targetSize, found = c.cache.GetMigTargetSize(ctx, migRef)
	if !found {
		return 0, fmt.Errorf("target size for %v not found in cache after refresh", migRef)
	}
	return targetSize, nil
}

func (c *cachingMigInfoProvider) GetMigBasename(ctx context.Context, migRef GceRef) (string, error) {
	c.migInfoMutex.Lock()
	defer c.migInfoMutex.Unlock()

	basename, found := c.cache.GetMigBasename(migRef)
	if found {
		return basename, nil
	}

	err := c.fillMigInfoCache(ctx)
	basename, found = c.cache.GetMigBasename(migRef)
	if err == nil && found {
		return basename, nil
	}

	err = c.fillSingleMigInfo(ctx, migRef)
	if err != nil {
		return "", err
	}
	basename, found = c.cache.GetMigBasename(migRef)
	if !found {
		return "", fmt.Errorf("basename for %v not found in cache after refresh", migRef)
	}
	return basename, nil
}

func (c *cachingMigInfoProvider) GetMigInstanceTemplateName(ctx context.Context, migRef GceRef) (InstanceTemplateName, error) {
	c.migInfoMutex.Lock()
	defer c.migInfoMutex.Unlock()

	instanceTemplateName, found := c.cache.GetMigInstanceTemplateName(ctx, migRef)
	if found {
		return instanceTemplateName, nil
	}

	err := c.fillMigInfoCache(ctx)
	instanceTemplateName, found = c.cache.GetMigInstanceTemplateName(ctx, migRef)
	if err == nil && found {
		return instanceTemplateName, nil
	}

	err = c.fillSingleMigInfo(ctx, migRef)
	if err != nil {
		return InstanceTemplateName{}, err
	}
	instanceTemplateName, found = c.cache.GetMigInstanceTemplateName(ctx, migRef)
	if !found {
		return InstanceTemplateName{}, fmt.Errorf("instance template name for %v not found in cache after refresh", migRef)
	}
	return instanceTemplateName, nil
}

func (c *cachingMigInfoProvider) GetMigInstanceTemplate(ctx context.Context, migRef GceRef) (*gce.InstanceTemplate, error) {
	logger := klog.FromContext(ctx)
	instanceTemplateName, err := c.GetMigInstanceTemplateName(ctx, migRef)
	if err != nil {
		return nil, err
	}

	template, found := c.cache.GetMigInstanceTemplate(ctx, migRef)
	if found && template.Name == instanceTemplateName.Name {
		return template, nil
	}
	logger.V(2).Info("Instance template of mig changed", "migName", migRef.Name, "instanceTemplate", instanceTemplateName.Name)
	template, err = c.gceClient.FetchMigTemplate(ctx, migRef, instanceTemplateName.Name, instanceTemplateName.Regional)
	if err != nil {
		return nil, err
	}
	c.cache.SetMigInstanceTemplate(migRef, template)
	return template, nil
}

func (c *cachingMigInfoProvider) GetMigKubeEnv(ctx context.Context, migRef GceRef) (KubeEnv, error) {
	instanceTemplateName, err := c.GetMigInstanceTemplateName(ctx, migRef)
	if err != nil {
		return KubeEnv{}, err
	}

	kubeEnv, kubeEnvFound := c.cache.GetMigKubeEnv(ctx, migRef)
	if kubeEnvFound && kubeEnv.templateName == instanceTemplateName.Name {
		return kubeEnv, nil
	}

	template, err := c.GetMigInstanceTemplate(ctx, migRef)
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
func (c *cachingMigInfoProvider) fillMigInfoCache(ctx context.Context) error {
	logger := klog.FromContext(ctx)
	var zones []string
	for zone := range c.listAllZonesWithMigs() {
		zones = append(zones, zone)
	}

	migs := make([][]*gce.InstanceGroupManager, len(zones))
	errors := make([]error, len(zones))
	workqueue.ParallelizeUntil(context.Background(), len(zones), len(zones), func(piece int) {
		migs[piece], errors[piece] = c.gceClient.FetchAllMigs(ctx, zones[piece])
	})

	failedZones := map[string]error{}
	failedZoneCount := 0
	for idx, err := range errors {
		if err != nil {
			logger.Error(err, "Error listing migs from zone", "zones", zones[idx])
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
			projectId := c.projectId
			if c.multiProjectCachingEnabled {
				var err error
				projectId, err = extractProjectWithRegex(zoneMig.SelfLink)
				if err != nil {
					// At this point we assume its the default project but this could eventually lead to a cache miss
					// if the project information is incorrect.
					projectId = c.projectId
					logger.Error(err, "Unable to extract projectID from MIG self link", "link", zoneMig.SelfLink)
				}
			}
			zoneMigRef := GceRef{
				projectId,
				zone,
				zoneMig.Name,
			}

			if registeredMigRefs[zoneMigRef] {
				c.setMigInfoCache(ctx, zoneMigRef, zoneMig)
			}
		}
	}

	return nil
}

// RefreshMigInfo updates the cached information for a specific MIG without rebuilding the full zone cache
func (c *cachingMigInfoProvider) RefreshMigInfo(ctx context.Context, migRef GceRef) error {
	return c.fillSingleMigInfo(ctx, migRef)
}

func (c *cachingMigInfoProvider) fillSingleMigInfo(ctx context.Context, migRef GceRef) error {
	igm, err := c.gceClient.FetchMig(ctx, migRef)
	if err != nil {
		c.migLister.HandleMigIssue(migRef, err)
		return err
	}
	c.setMigInfoCache(ctx, migRef, igm)
	return nil
}

func (c *cachingMigInfoProvider) setMigInfoCache(ctx context.Context, migRef GceRef, mig *gce.InstanceGroupManager) {
	logger := klog.FromContext(ctx)
	c.cache.SetMigTargetSize(migRef, mig.TargetSize+mig.TargetSuspendedSize)
	c.cache.SetMigBasename(migRef, mig.BaseInstanceName)
	if mig.Status != nil {
		c.cache.SetMigIsStable(migRef, mig.Status.IsStable)
	} else {
		logger.Info("MIG has nil status, assuming isStable=false", "mig", migRef)
		c.cache.SetMigIsStable(migRef, false)
	}
	c.cache.SetListManagedInstancesResults(migRef, mig.ListManagedInstancesResults)
	c.cache.SetMigInstancesStateCount(migRef, createInstancesStateCount(mig.TargetSize, mig.CurrentActions))

	_, templateName := path.Split(mig.InstanceTemplate)
	regional := IsInstanceTemplateRegional(mig.InstanceTemplate)
	c.cache.SetMigInstanceTemplateName(migRef, InstanceTemplateName{templateName, regional})
}

func (c *cachingMigInfoProvider) GetMigIsStable(ctx context.Context, migRef GceRef) (bool, error) {
	c.migInfoMutex.Lock()
	defer c.migInfoMutex.Unlock()

	isStable, found := c.cache.GetMigIsStable(ctx, migRef)
	if found {
		return isStable, nil
	}

	err := c.fillMigInfoCache(ctx)
	isStable, found = c.cache.GetMigIsStable(ctx, migRef)
	if err == nil && found {
		return isStable, nil
	}

	err = c.fillSingleMigInfo(ctx, migRef)
	if err != nil {
		return false, err
	}
	isStable, found = c.cache.GetMigIsStable(ctx, migRef)
	if !found {
		return false, fmt.Errorf("isStable for %v not found in cache after refresh", migRef)
	}
	return isStable, nil
}

// extractProjectWithRegex uses a regular expression to find and return the project name
// from the selfLink of a MIG.
func extractProjectWithRegex(selflink string) (string, error) {
	// FindStringSubmatch returns an array with the full match and all captured groups.
	// matches[0] will be the full matched string (e.g., "/projects/some-project").
	// matches[1] will be the content of the first capturing group (e.g., "some-project").
	matches := migProjectSelfLinkRe.FindStringSubmatch(selflink)
	if len(matches) < 2 {
		return "", fmt.Errorf("could not find project name in self link: %s", selflink)
	}
	return matches[1], nil
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

func (c *cachingMigInfoProvider) GetMigMachineType(ctx context.Context, migRef GceRef) (MachineType, error) {
	template, err := c.GetMigInstanceTemplate(ctx, migRef)
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
		rawMachine, err := c.gceClient.FetchMachineType(ctx, zone, machineName)
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

func (c *cachingMigInfoProvider) GetListManagedInstancesResults(ctx context.Context, migRef GceRef) (string, error) {
	c.migInfoMutex.Lock()
	defer c.migInfoMutex.Unlock()

	listManagedInstancesResults, found := c.cache.GetListManagedInstancesResults(migRef)
	if found {
		return listManagedInstancesResults, nil
	}

	err := c.fillMigInfoCache(ctx)
	listManagedInstancesResults, found = c.cache.GetListManagedInstancesResults(migRef)
	if err == nil && found {
		return listManagedInstancesResults, nil
	}

	err = c.fillSingleMigInfo(ctx, migRef)
	if err != nil {
		return "", err
	}
	listManagedInstancesResults, found = c.cache.GetListManagedInstancesResults(migRef)
	if !found {
		return "", fmt.Errorf("listManagedInstancesResults for %v not found in cache after refresh", migRef)
	}
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
