/*
Copyright 2016 The Kubernetes Authors.

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
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/gce/localssdsize"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/cluster-autoscaler/pkg/cloudprovider"
	"sigs.k8s.io/cluster-autoscaler/pkg/config"
	"sigs.k8s.io/cluster-autoscaler/pkg/config/dynamic"

	apiv1 "k8s.io/api/core/v1"
	provider_gce "k8s.io/cloud-provider-gcp/providers/gce"

	"cloud.google.com/go/compute/metadata"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	gce "google.golang.org/api/compute/v1"
	gcfg "gopkg.in/gcfg.v1"
	klog "k8s.io/klog/v2"
)

const (
	refreshInterval              = 1 * time.Minute
	machinesRefreshInterval      = 1 * time.Hour
	httpTimeout                  = 3 * time.Minute
	scaleToZeroSupported         = true
	autoDiscovererTypeMIG        = "mig"
	migAutoDiscovererKeyPrefix   = "namePrefix"
	migAutoDiscovererKeyMinNodes = "min"
	migAutoDiscovererKeyMaxNodes = "max"
	createInstancesRequestLimit  = 1000
)

var (
	defaultOAuthScopes = []string{
		"https://www.googleapis.com/auth/compute",
		"https://www.googleapis.com/auth/devstorage.read_only",
		"https://www.googleapis.com/auth/service.management.readonly",
		"https://www.googleapis.com/auth/servicecontrol",
	}

	validMIGAutoDiscovererKeys = strings.Join([]string{
		migAutoDiscovererKeyPrefix,
		migAutoDiscovererKeyMinNodes,
		migAutoDiscovererKeyMaxNodes,
	}, ", ")
)

// GceManager handles GCE communication and data caching.
type GceManager interface {
	// Refresh triggers refresh of cached resources.
	Refresh(ctx context.Context) error
	// Cleanup cleans up open resources before the cloud provider is destroyed, i.e. go routines etc.
	Cleanup() error

	// GetMigs returns list of registered MIGs.
	GetMigs() []Mig
	// GetMigNodes returns mig nodes.
	GetMigNodes(ctx context.Context, mig Mig) ([]GceInstance, error)
	// GetMigForInstance returns MIG to which the given instance belongs.
	GetMigForInstance(ctx context.Context, instance GceRef) (Mig, error)
	// GetMigTemplateNode returns a template node for MIG.
	GetMigTemplateNode(ctx context.Context, mig Mig) (*apiv1.Node, error)
	// GetResourceLimiter returns resource limiter.
	GetResourceLimiter() (*cloudprovider.ResourceLimiter, error)
	// GetMigSize gets MIG size.
	GetMigSize(ctx context.Context, mig Mig) (int64, error)
	// GetMigOptions returns MIG's NodeGroupAutoscalingOptions
	GetMigOptions(ctx context.Context, mig Mig, defaults config.NodeGroupAutoscalingOptions) *config.NodeGroupAutoscalingOptions
	// IsMigStable returns whether the MIG is stable. A stable state means that: none of the instances in the managed instance group is currently undergoing any type of change (for example, creation, restart, or deletion); no future changes are scheduled for instances in the managed instance group; and the managed instance group itself is not being modified.
	IsMigStable(mig Mig) (bool, error)

	// SetMigSize sets MIG size.
	SetMigSize(ctx context.Context, mig Mig, size int64) error
	// DeleteInstances deletes the given instances. All instances must be controlled by the same MIG.
	DeleteInstances(ctx context.Context, instances []GceRef) error
	// CreateInstances creates delta new instances in a given mig.
	CreateInstances(ctx context.Context, mig Mig, delta int64) error
}

type gceManagerImpl struct {
	cache                    *GceCache
	lastRefresh              time.Time
	machinesCacheLastRefresh time.Time
	concurrentGceRefreshes   int

	GceService      AutoscalingGceClient
	migInfoProvider MigInfoProvider
	migLister       MigLister

	location                 string
	projectId                string
	domainUrl                string
	templates                *GceTemplateBuilder
	interrupt                chan struct{}
	regional                 bool
	explicitlyConfigured     map[GceRef]bool
	migAutoDiscoverySpecs    []migAutoDiscoveryConfig
	reserved                 *GceReserved
	localSSDDiskSizeProvider localssdsize.LocalSSDSizeProvider
}

// CreateGceManager constructs GceManager object.
func CreateGceManager(configReader io.Reader, discoveryOpts cloudprovider.NodeGroupDiscoveryOptions,
	localSSDDiskSizeProvider localssdsize.LocalSSDSizeProvider,
	regional, bulkGceMigInstancesListingEnabled bool, concurrentGceRefreshes int, userAgent, domainUrl string, migInstancesMinRefreshWaitTime time.Duration) (GceManager, error) {
	// Create Google Compute Engine token.
	var err error
	tokenSource := google.ComputeTokenSource("")
	if len(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")) > 0 {
		tokenSource, err = google.DefaultTokenSource(oauth2.NoContext, gce.ComputeScope)
		if err != nil {
			return nil, err
		}
	}
	var projectId, location string
	if configReader != nil {
		var cfg provider_gce.ConfigFile
		if err := gcfg.ReadInto(&cfg, configReader); err != nil {
			klog.Errorf("Couldn't read config: %v", err)
			return nil, err
		}
		if cfg.Global.TokenURL == "" {
			klog.Warning("Empty tokenUrl in cloud config")
		} else {
			tokenSource = provider_gce.NewAltTokenSource(cfg.Global.TokenURL, cfg.Global.TokenBody)
			klog.V(1).Infof("Using TokenSource from config %#v", tokenSource)
		}
		projectId = cfg.Global.ProjectID
		location = cfg.Global.LocalZone
	} else {
		klog.V(1).Infof("Using default TokenSource %#v", tokenSource)
	}
	if len(projectId) == 0 || len(location) == 0 {
		// XXX: On GKE discoveredProjectId is hosted master project and
		// not the project we want to use, however, zone seems to not
		// be specified in config. For now we can just assume that hosted
		// master project is in the same zone as cluster and only use
		// discoveredZone.
		discoveredProjectId, discoveredLocation, err := getProjectAndLocation(regional)
		if err != nil {
			return nil, err
		}
		if len(projectId) == 0 {
			projectId = discoveredProjectId
		}
		if len(location) == 0 {
			location = discoveredLocation
		}
	}
	klog.V(1).Infof("GCE projectId=%s location=%s", projectId, location)

	// Create Google Compute Engine service.
	client := oauth2.NewClient(context.Background(), tokenSource)
	client.Timeout = httpTimeout
	gceService, err := NewAutoscalingGceClientV1(client, projectId, userAgent)
	if err != nil {
		return nil, err
	}
	cache := NewGceCache()
	migLister := NewMigLister(cache)
	manager := &gceManagerImpl{
		cache:                    cache,
		GceService:               gceService,
		migLister:                migLister,
		migInfoProvider:          NewCachingMigInfoProvider(cache, migLister, gceService, projectId, concurrentGceRefreshes, migInstancesMinRefreshWaitTime, bulkGceMigInstancesListingEnabled, false),
		location:                 location,
		regional:                 regional,
		projectId:                projectId,
		templates:                &GceTemplateBuilder{},
		interrupt:                make(chan struct{}),
		explicitlyConfigured:     make(map[GceRef]bool),
		concurrentGceRefreshes:   concurrentGceRefreshes,
		reserved:                 &GceReserved{},
		domainUrl:                domainUrl,
		localSSDDiskSizeProvider: localSSDDiskSizeProvider,
	}

	if err := manager.fetchExplicitMigs(discoveryOpts.NodeGroupSpecs); err != nil {
		return nil, fmt.Errorf("failed to fetch MIGs: %v", err)
	}
	if manager.migAutoDiscoverySpecs, err = parseMIGAutoDiscoverySpecs(discoveryOpts); err != nil {
		return nil, err
	}

	if err := manager.forceRefresh(context.TODO()); err != nil {
		return nil, err
	}

	go wait.Until(func() {
		if err := manager.migInfoProvider.RegenerateMigInstancesCache(context.TODO()); err != nil {
			klog.Errorf("Error while regenerating Mig cache: %v", err)
		}
	}, time.Hour, manager.interrupt)

	return manager, nil
}

// Cleanup closes the channel to stop the goroutine refreshing cache.
func (m *gceManagerImpl) Cleanup() error {
	close(m.interrupt)
	return nil
}

// registerMig registers mig in GceManager. Returns true if the node group didn't exist before or its config has changed.
func (m *gceManagerImpl) registerMig(ctx context.Context, mig Mig) bool {
	logger := klog.FromContext(ctx)
	changed := m.cache.RegisterMig(ctx, mig)
	if changed {
		// Try to build a node from template to validate that this group
		// can be scaled up from 0 nodes.
		// We may never need to do it, so just log error if it fails.
		if _, err := m.GetMigTemplateNode(ctx, mig); err != nil {
			logger.Error(err, "Can't build node from mig template. Won't be able to scale from 0.", "mig", mig.GceRef())
		}
	}
	return changed
}

// GetMigSize gets MIG size.
func (m *gceManagerImpl) GetMigSize(ctx context.Context, mig Mig) (int64, error) {
	return m.migInfoProvider.GetMigTargetSize(ctx, mig.GceRef())
}

// IsMigStable returns whether the MIG is stable.
func (m *gceManagerImpl) IsMigStable(mig Mig) (bool, error) {
	return m.migInfoProvider.GetMigIsStable(mig.GceRef())
}

// SetMigSize sets MIG size.
func (m *gceManagerImpl) SetMigSize(ctx context.Context, mig Mig, size int64) error {
	logger := klog.FromContext(ctx)
	logger.V(0).Info("Setting mig size", "mig", mig.GceRef(), "size", size)
	m.cache.InvalidateMigTargetSize(ctx, mig.GceRef())
	err := m.GceService.ResizeMig(ctx, mig.GceRef(), size)
	if err != nil {
		return err
	}
	m.cache.SetMigTargetSize(mig.GceRef(), size)
	return nil
}

// DeleteInstances deletes the given instances. All instances must be controlled by the same MIG.
func (m *gceManagerImpl) DeleteInstances(ctx context.Context, instances []GceRef) error {
	if len(instances) == 0 {
		return nil
	}
	commonMig, err := m.GetMigForInstance(ctx, instances[0])
	if err != nil {
		return err
	}
	for _, instance := range instances {
		mig, err := m.GetMigForInstance(ctx, instance)
		if err != nil {
			return err
		}
		if mig != commonMig {
			return fmt.Errorf("cannot delete instances which don't belong to the same MIG.")
		}
	}
	m.cache.InvalidateMigTargetSize(ctx, commonMig.GceRef())
	m.cache.InvalidateMigInstances(ctx, commonMig.GceRef())
	return m.GceService.DeleteInstances(ctx, commonMig.GceRef(), instances)
}

// GetMigs returns list of registered MIGs.
func (m *gceManagerImpl) GetMigs() []Mig {
	return m.migLister.GetMigs()
}

// GetMigForInstance returns MIG to which the given instance belongs.
func (m *gceManagerImpl) GetMigForInstance(ctx context.Context, instance GceRef) (Mig, error) {
	return m.migInfoProvider.GetMigForInstance(ctx, instance)
}

// GetMigNodes returns mig nodes.
func (m *gceManagerImpl) GetMigNodes(ctx context.Context, mig Mig) ([]GceInstance, error) {
	return m.migInfoProvider.GetMigInstances(ctx, mig.GceRef())
}

// Refresh triggers refresh of cached resources.
func (m *gceManagerImpl) Refresh(ctx context.Context) error {
	m.cache.InvalidateAllMigInstances(ctx)
	m.cache.InvalidateAllMigTargetSizes(ctx)
	m.cache.InvalidateAllMigIsStable(ctx)
	m.cache.InvalidateAllMigBasenames()
	m.cache.InvalidateAllListManagedInstancesResults()
	m.cache.InvalidateAllMigInstanceTemplateNames(ctx)
	if m.lastRefresh.Add(refreshInterval).After(time.Now()) {
		return nil
	}

	if err := m.forceRefresh(ctx); err != nil {
		return err
	}

	migs := m.migLister.GetMigs()
	m.cache.DropInstanceTemplatesForMissingMigs(ctx, migs)
	return nil
}

func (m *gceManagerImpl) CreateInstances(ctx context.Context, mig Mig, delta int64) error {
	logger := klog.FromContext(ctx)
	if delta == 0 {
		return nil
	}
	instances, err := m.GetMigNodes(ctx, mig)
	if err != nil {
		return err
	}
	instanceIds := make([]string, 0, len(instances)+int(delta))
	for _, ins := range instances {
		instanceIds = append(instanceIds, ins.Id)
	}
	baseName, err := m.migInfoProvider.GetMigBasename(ctx, mig.GceRef())
	if err != nil {
		return fmt.Errorf("can't upscale %s: failed to collect BaseInstanceName: %w", mig.GceRef(), err)
	}
	m.cache.InvalidateMigTargetSize(ctx, mig.GceRef())
	totalReqs := int((delta + createInstancesRequestLimit - 1) / createInstancesRequestLimit)
	remaining := delta
	for i := 0; i < totalReqs; i++ {
		increment := min(remaining, createInstancesRequestLimit)
		if totalReqs > 1 {
			logger.Info("Sending chunked GCE createInstances request.", "request", fmt.Sprintf("%d/%d", i+1, totalReqs), "requestSize", increment)
		}
		ids, err := m.GceService.CreateInstances(ctx, mig.GceRef(), baseName, increment, instanceIds)
		if err != nil {
			return err
		}
		remaining -= increment
		for _, id := range ids {
			instanceIds = append(instanceIds, id)
		}
	}
	return nil
}

func (m *gceManagerImpl) forceRefresh(ctx context.Context) error {
	logger := klog.FromContext(ctx)
	m.clearMachinesCache(ctx)
	if err := m.fetchAutoMigs(ctx); err != nil {
		logger.Error(err, "Failed to fetch MIGs")
		return err
	}
	m.refreshAutoscalingOptions(ctx)
	m.lastRefresh = time.Now()
	logger.V(2).Info("Refreshed GCE resources", "nextRefreshAfter", m.lastRefresh.Add(refreshInterval))
	return nil
}

func (m *gceManagerImpl) refreshAutoscalingOptions(ctx context.Context) {
	logger := klog.FromContext(ctx)
	for _, mig := range m.migLister.GetMigs() {
		template, err := m.migInfoProvider.GetMigInstanceTemplate(ctx, mig.GceRef())
		if err != nil {
			logger.Info("Not evaluating autoscaling options for mig: failed to find corresponding instance template", "mig", mig.GceRef(), "err", err)
			continue
		}
		if template.Properties == nil {
			logger.Info("Failed to extract autoscaling options from metadata: instance template is incomplete", "templateName", template.Name)
			continue
		}
		kubeEnv, err := m.migInfoProvider.GetMigKubeEnv(ctx, mig.GceRef())
		if err != nil {
			logger.Info("Failed to extract autoscaling options from instance template", "templateName", template.Name, "err", err)
			continue
		}
		options, err := extractAutoscalingOptionsFromKubeEnv(ctx, kubeEnv)
		if err != nil {
			logger.Info("Failed to extract autoscaling options from instance template's metadata", "templateName", template.Name, "err", err)
			continue
		}
		if !reflect.DeepEqual(m.cache.GetAutoscalingOptions(mig.GceRef()), options) {
			logger.V(4).Info("Extracted autoscaling options from instance template KubeEnv", "templateName", template.Name, "options", options)
		}
		m.cache.SetAutoscalingOptions(mig.GceRef(), options)
	}
}

// Fetch explicitly configured MIGs. These MIGs should never be unregistered
// during refreshes, even if they no longer exist in GCE.
func (m *gceManagerImpl) fetchExplicitMigs(specs []string) error {
	changed := false
	for _, spec := range specs {
		mig, err := m.buildMigFromFlag(spec)
		if err != nil {
			return err
		}
		if m.registerMig(context.TODO(), mig) {
			changed = true
		}
		m.explicitlyConfigured[mig.GceRef()] = true
	}

	if changed {
		return m.migInfoProvider.RegenerateMigInstancesCache(context.TODO())
	}
	return nil
}

func (m *gceManagerImpl) buildMigFromFlag(flag string) (Mig, error) {
	s, err := dynamic.SpecFromString(flag, scaleToZeroSupported)
	if err != nil {
		return nil, fmt.Errorf("failed to parse node group spec: %v", err)
	}
	return m.buildMigFromSpec(s)
}

func (m *gceManagerImpl) buildMigFromAutoCfg(link string, cfg migAutoDiscoveryConfig) (Mig, error) {
	s := &dynamic.NodeGroupSpec{
		Name:               link,
		MinSize:            cfg.MinSize,
		MaxSize:            cfg.MaxSize,
		SupportScaleToZero: scaleToZeroSupported,
	}
	return m.buildMigFromSpec(s)
}

func (m *gceManagerImpl) buildMigFromSpec(s *dynamic.NodeGroupSpec) (Mig, error) {
	if err := s.Validate(); err != nil {
		return nil, fmt.Errorf("invalid node group spec: %v", err)
	}
	project, zone, name, err := ParseMigUrl(s.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to parse mig url: %s got error: %v", s.Name, err)
	}
	mig := &gceMig{
		gceRef: GceRef{
			Project: project,
			Name:    name,
			Zone:    zone,
		},
		gceManager: m,
		minSize:    s.MinSize,
		maxSize:    s.MaxSize,
		domainUrl:  m.domainUrl,
	}
	return mig, nil
}

// Fetch automatically discovered MIGs. These MIGs should be unregistered if
// they no longer exist in GCE.
func (m *gceManagerImpl) fetchAutoMigs(ctx context.Context) error {
	logger := klog.FromContext(ctx)
	exists := make(map[GceRef]bool)
	var changed int32 = 0

	toRegister := make([]Mig, 0)
	for _, cfg := range m.migAutoDiscoverySpecs {
		links, err := m.findMigsNamed(ctx, cfg.Re)
		if err != nil {
			return fmt.Errorf("cannot autodiscover managed instance groups: %v", err)
		}

		for _, link := range links {
			mig, err := m.buildMigFromAutoCfg(link, cfg)
			if err != nil {
				return err
			}
			exists[mig.GceRef()] = true
			if m.explicitlyConfigured[mig.GceRef()] {
				// This MIG was explicitly configured, but would also be
				// autodiscovered. We want the explicitly configured min and max
				// nodes to take precedence.
				logger.V(3).Info("Ignoring explicitly configured MIG in autodiscovery", "mig", mig.GceRef())
				continue
			}
			toRegister = append(toRegister, mig)
		}
	}

	workqueue.ParallelizeUntil(ctx, m.concurrentGceRefreshes, len(toRegister), func(piece int) {
		mig := toRegister[piece]
		if m.registerMig(ctx, mig) {
			logger.V(3).Info("Autodiscovered MIG", "mig", mig.GceRef())
			atomic.StoreInt32(&changed, int32(1))
		}
	}, workqueue.WithChunkSize(m.concurrentGceRefreshes))

	for _, mig := range m.GetMigs() {
		if !exists[mig.GceRef()] && !m.explicitlyConfigured[mig.GceRef()] {
			m.cache.UnregisterMig(ctx, mig)
			atomic.StoreInt32(&changed, int32(1))
		}
	}

	if atomic.LoadInt32(&changed) > 0 {
		return m.migInfoProvider.RegenerateMigInstancesCache(ctx)
	}

	return nil
}

// GetResourceLimiter returns resource limiter from cache.
func (m *gceManagerImpl) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	return m.cache.GetResourceLimiter()
}

func (m *gceManagerImpl) clearMachinesCache(ctx context.Context) {
	logger := klog.FromContext(ctx)
	if m.machinesCacheLastRefresh.Add(machinesRefreshInterval).After(time.Now()) {
		return
	}

	m.cache.InvalidateAllMachines()
	nextRefresh := time.Now()
	m.machinesCacheLastRefresh = nextRefresh
	logger.V(2).Info("Cleared machine types cache", "nextClearAfter", nextRefresh)
}

// Code borrowed from gce cloud provider. Reuse the original as soon as it becomes public.
func getProjectAndLocation(regional bool) (string, string, error) {
	result, err := metadata.Get("instance/zone")
	if err != nil {
		return "", "", err
	}
	parts := strings.Split(result, "/")
	if len(parts) != 4 {
		return "", "", fmt.Errorf("unexpected response: %s", result)
	}
	location := parts[3]
	if regional {
		location, err = provider_gce.GetGCERegion(location)
		if err != nil {
			return "", "", err
		}
	}
	projectID, err := metadata.ProjectID()
	if err != nil {
		return "", "", err
	}
	return projectID, location, nil
}

func (m *gceManagerImpl) findMigsNamed(ctx context.Context, name *regexp.Regexp) ([]string, error) {
	if m.regional {
		return m.findMigsInRegion(ctx, m.location, name)
	}
	return m.GceService.FetchMigsWithName(ctx, m.location, name)
}

func (m *gceManagerImpl) getZones(ctx context.Context, region string) ([]string, error) {
	zones, err := m.GceService.FetchZones(ctx, region)
	if err != nil {
		return nil, fmt.Errorf("cannot get zones for GCE region %s: %v", region, err)
	}
	return zones, nil
}

func (m *gceManagerImpl) findMigsInRegion(ctx context.Context, region string, name *regexp.Regexp) ([]string, error) {
	links := make([]string, 0)
	zones, err := m.getZones(ctx, region)
	if err != nil {
		return nil, err
	}

	zoneLinks := make([][]string, len(zones))
	errors := make([]error, len(zones))
	workqueue.ParallelizeUntil(ctx, len(zones), len(zones), func(piece int) {
		zoneLinks[piece], errors[piece] = m.GceService.FetchMigsWithName(ctx, zones[piece], name)
	})

	for _, err := range errors {
		if err != nil {
			return nil, fmt.Errorf("%v", errors)
		}
	}

	for _, zl := range zoneLinks {
		for _, link := range zl {
			links = append(links, link)
		}
	}

	return links, nil
}

// GetMigOptions merges default options with user-provided options as specified in the MIG's instance template metadata
func (m *gceManagerImpl) GetMigOptions(ctx context.Context, mig Mig, defaults config.NodeGroupAutoscalingOptions) *config.NodeGroupAutoscalingOptions {
	migRef := mig.GceRef()
	options := m.cache.GetAutoscalingOptions(migRef)
	if options == nil {
		return &defaults
	}

	if opt, ok := getFloat64Option(ctx, options, migRef.Name, config.DefaultScaleDownUtilizationThresholdKey); ok {
		defaults.ScaleDownUtilizationThreshold = opt
	}
	if opt, ok := getFloat64Option(ctx, options, migRef.Name, config.DefaultScaleDownGpuUtilizationThresholdKey); ok {
		defaults.ScaleDownGpuUtilizationThreshold = opt
	}
	if opt, ok := getDurationOption(ctx, options, migRef.Name, config.DefaultScaleDownUnneededTimeKey); ok {
		defaults.ScaleDownUnneededTime = opt
	}
	if opt, ok := getDurationOption(ctx, options, migRef.Name, config.DefaultScaleDownUnreadyTimeKey); ok {
		defaults.ScaleDownUnreadyTime = opt
	}
	if opt, ok := getDurationOption(ctx, options, migRef.Name, config.DefaultMaxNodeProvisionTimeKey); ok {
		defaults.MaxNodeProvisionTime = opt
	}

	return &defaults
}

// GetMigTemplateNode constructs a node from GCE instance template of the given MIG.
func (m *gceManagerImpl) GetMigTemplateNode(ctx context.Context, mig Mig) (*apiv1.Node, error) {
	template, err := m.migInfoProvider.GetMigInstanceTemplate(ctx, mig.GceRef())
	if err != nil {
		return nil, err
	}
	kubeEnv, err := m.migInfoProvider.GetMigKubeEnv(ctx, mig.GceRef())
	if err != nil {
		return nil, err
	}
	machineType, err := m.migInfoProvider.GetMigMachineType(ctx, mig.GceRef())
	if err != nil {
		return nil, err
	}
	migOsInfo, err := m.templates.MigOsInfo(ctx, mig.Id(), kubeEnv)
	if err != nil {
		return nil, err
	}
	return m.templates.BuildNodeFromTemplate(ctx, mig, migOsInfo, template, kubeEnv, machineType.CPU, machineType.Memory, nil, m.reserved, m.localSSDDiskSizeProvider)
}

// parseMIGAutoDiscoverySpecs returns any provided NodeGroupAutoDiscoverySpecs
// parsed into configuration appropriate for MIG autodiscovery.
func parseMIGAutoDiscoverySpecs(o cloudprovider.NodeGroupDiscoveryOptions) ([]migAutoDiscoveryConfig, error) {
	cfgs := make([]migAutoDiscoveryConfig, len(o.NodeGroupAutoDiscoverySpecs))
	var err error
	for i, spec := range o.NodeGroupAutoDiscoverySpecs {
		cfgs[i], err = parseMIGAutoDiscoverySpec(spec)
		if err != nil {
			return nil, err
		}
	}
	return cfgs, nil
}

// A migAutoDiscoveryConfig specifies how to autodiscover GCE MIGs.
type migAutoDiscoveryConfig struct {
	// Re is a regexp passed using the eq filter to the GCE list API.
	Re *regexp.Regexp
	// MinSize specifies the minimum size for all MIGs that match Re.
	MinSize int
	// MaxSize specifies the maximum size for all MIGs that match Re.
	MaxSize int
}

func parseMIGAutoDiscoverySpec(spec string) (migAutoDiscoveryConfig, error) {
	cfg := migAutoDiscoveryConfig{}

	tokens := strings.Split(spec, ":")
	if len(tokens) != 2 {
		return cfg, fmt.Errorf("spec \"%s\" should be discoverer:key=value,key=value", spec)
	}
	discoverer := tokens[0]
	if discoverer != autoDiscovererTypeMIG {
		return cfg, fmt.Errorf("unsupported discoverer specified: %s", discoverer)
	}

	for _, arg := range strings.Split(tokens[1], ",") {
		kv := strings.Split(arg, "=")
		if len(kv) != 2 {
			return cfg, fmt.Errorf("invalid key=value pair %s", kv)
		}
		k, v := kv[0], kv[1]

		var err error
		switch k {
		case migAutoDiscovererKeyPrefix:
			if cfg.Re, err = regexp.Compile(fmt.Sprintf("^%s.+", v)); err != nil {
				return cfg, fmt.Errorf("invalid instance group name prefix \"%s\" - \"^%s.+\" must be a valid RE2 regexp", v, v)
			}
		case migAutoDiscovererKeyMinNodes:
			if cfg.MinSize, err = strconv.Atoi(v); err != nil {
				return cfg, fmt.Errorf("invalid minimum nodes: %s", v)
			}
		case migAutoDiscovererKeyMaxNodes:
			if cfg.MaxSize, err = strconv.Atoi(v); err != nil {
				return cfg, fmt.Errorf("invalid maximum nodes: %s", v)
			}
		default:
			return cfg, fmt.Errorf("unsupported key \"%s\" is specified for discoverer \"%s\". Supported keys are \"%s\"", k, discoverer, validMIGAutoDiscovererKeys)
		}
	}
	if cfg.Re == nil || cfg.Re.String() == "^.+" {
		return cfg, errors.New("empty instance group name prefix supplied")
	}
	if cfg.MinSize > cfg.MaxSize {
		return cfg, fmt.Errorf("minimum size %d is greater than maximum size %d", cfg.MinSize, cfg.MaxSize)
	}
	if cfg.MaxSize < 1 {
		return cfg, fmt.Errorf("maximum size %d must be at least 1", cfg.MaxSize)
	}
	return cfg, nil
}
