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
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/config/dynamic"
	"k8s.io/client-go/util/workqueue"

	apiv1 "k8s.io/api/core/v1"
	provider_gce "k8s.io/legacy-cloud-providers/gce"

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
	httpTimeout                  = 30 * time.Second
	scaleToZeroSupported         = true
	autoDiscovererTypeMIG        = "mig"
	migAutoDiscovererKeyPrefix   = "namePrefix"
	migAutoDiscovererKeyMinNodes = "min"
	migAutoDiscovererKeyMaxNodes = "max"
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
	Refresh() error
	// Cleanup cleans up open resources before the cloud provider is destroyed, i.e. go routines etc.
	Cleanup() error

	// GetMigs returns list of registered MIGs.
	GetMigs() []Mig
	// GetMigNodes returns mig nodes.
	GetMigNodes(mig Mig) ([]cloudprovider.Instance, error)
	// GetMigForInstance returns MIG to which the given instance belongs.
	GetMigForInstance(instance GceRef) (Mig, error)
	// GetMigTemplateNode returns a template node for MIG.
	GetMigTemplateNode(mig Mig) (*apiv1.Node, error)
	// GetResourceLimiter returns resource limiter.
	GetResourceLimiter() (*cloudprovider.ResourceLimiter, error)
	// GetMigSize gets MIG size.
	GetMigSize(mig Mig) (int64, error)
	// GetMigOptions returns MIG's NodeGroupAutoscalingOptions
	GetMigOptions(mig Mig, defaults config.NodeGroupAutoscalingOptions) *config.NodeGroupAutoscalingOptions

	// SetMigSize sets MIG size.
	SetMigSize(mig Mig, size int64) error
	// DeleteInstances deletes the given instances. All instances must be controlled by the same MIG.
	DeleteInstances(instances []GceRef) error
	// CreateInstances creates delta new instances in a given mig.
	CreateInstances(mig Mig, delta int64) error
}

type gceManagerImpl struct {
	cache                    *GceCache
	lastRefresh              time.Time
	machinesCacheLastRefresh time.Time
	concurrentGceRefreshes   int

	GceService      AutoscalingGceClient
	migInfoProvider MigInfoProvider
	migLister       MigLister

	location              string
	projectId             string
	templates             *GceTemplateBuilder
	interrupt             chan struct{}
	regional              bool
	explicitlyConfigured  map[GceRef]bool
	migAutoDiscoverySpecs []migAutoDiscoveryConfig
	reserved              *GceReserved
}

// CreateGceManager constructs GceManager object.
func CreateGceManager(configReader io.Reader, discoveryOpts cloudprovider.NodeGroupDiscoveryOptions, regional bool, concurrentGceRefreshes int, userAgent string) (GceManager, error) {
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
	client := oauth2.NewClient(oauth2.NoContext, tokenSource)
	client.Timeout = httpTimeout
	gceService, err := NewAutoscalingGceClientV1(client, projectId, userAgent)
	if err != nil {
		return nil, err
	}
	cache := NewGceCache()
	migLister := NewMigLister(cache)
	manager := &gceManagerImpl{
		cache:                  cache,
		GceService:             gceService,
		migLister:              migLister,
		migInfoProvider:        NewCachingMigInfoProvider(cache, migLister, gceService, projectId, concurrentGceRefreshes),
		location:               location,
		regional:               regional,
		projectId:              projectId,
		templates:              &GceTemplateBuilder{},
		interrupt:              make(chan struct{}),
		explicitlyConfigured:   make(map[GceRef]bool),
		concurrentGceRefreshes: concurrentGceRefreshes,
		reserved:               &GceReserved{},
	}

	if err := manager.fetchExplicitMigs(discoveryOpts.NodeGroupSpecs); err != nil {
		return nil, fmt.Errorf("failed to fetch MIGs: %v", err)
	}
	if manager.migAutoDiscoverySpecs, err = parseMIGAutoDiscoverySpecs(discoveryOpts); err != nil {
		return nil, err
	}

	if err := manager.forceRefresh(); err != nil {
		return nil, err
	}

	go wait.Until(func() {
		if err := manager.migInfoProvider.RegenerateMigInstancesCache(); err != nil {
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
func (m *gceManagerImpl) registerMig(mig Mig) bool {
	changed := m.cache.RegisterMig(mig)
	if changed {
		// Try to build a node from template to validate that this group
		// can be scaled up from 0 nodes.
		// We may never need to do it, so just log error if it fails.
		if _, err := m.GetMigTemplateNode(mig); err != nil {
			klog.Errorf("Can't build node from template for %s, won't be able to scale from 0: %v", mig.GceRef().String(), err)
		}
	}
	return changed
}

// GetMigSize gets MIG size.
func (m *gceManagerImpl) GetMigSize(mig Mig) (int64, error) {
	return m.migInfoProvider.GetMigTargetSize(mig.GceRef())
}

// SetMigSize sets MIG size.
func (m *gceManagerImpl) SetMigSize(mig Mig, size int64) error {
	klog.V(0).Infof("Setting mig size %s to %d", mig.Id(), size)
	m.cache.InvalidateMigTargetSize(mig.GceRef())
	err := m.GceService.ResizeMig(mig.GceRef(), size)
	if err != nil {
		return err
	}
	m.cache.SetMigTargetSize(mig.GceRef(), size)
	return nil
}

// DeleteInstances deletes the given instances. All instances must be controlled by the same MIG.
func (m *gceManagerImpl) DeleteInstances(instances []GceRef) error {
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
			return fmt.Errorf("cannot delete instances which don't belong to the same MIG.")
		}
	}
	m.cache.InvalidateMigTargetSize(commonMig.GceRef())
	return m.GceService.DeleteInstances(commonMig.GceRef(), instances)
}

// GetMigs returns list of registered MIGs.
func (m *gceManagerImpl) GetMigs() []Mig {
	return m.migLister.GetMigs()
}

// GetMigForInstance returns MIG to which the given instance belongs.
func (m *gceManagerImpl) GetMigForInstance(instance GceRef) (Mig, error) {
	return m.migInfoProvider.GetMigForInstance(instance)
}

// GetMigNodes returns mig nodes.
func (m *gceManagerImpl) GetMigNodes(mig Mig) ([]cloudprovider.Instance, error) {
	return m.migInfoProvider.GetMigInstances(mig.GceRef())
}

// Refresh triggers refresh of cached resources.
func (m *gceManagerImpl) Refresh() error {
	m.cache.InvalidateAllMigInstances()
	m.cache.InvalidateAllMigTargetSizes()
	m.cache.InvalidateAllMigBasenames()
	m.cache.InvalidateAllMigInstanceTemplateNames()
	if m.lastRefresh.Add(refreshInterval).After(time.Now()) {
		return nil
	}
	return m.forceRefresh()
}

func (m *gceManagerImpl) CreateInstances(mig Mig, delta int64) error {
	if delta == 0 {
		return nil
	}
	instances, err := m.GetMigNodes(mig)
	if err != nil {
		return err
	}
	instancesNames := make([]string, 0, len(instances))
	for _, ins := range instances {
		instancesNames = append(instancesNames, ins.Id)
	}
	baseName, err := m.migInfoProvider.GetMigBasename(mig.GceRef())
	if err != nil {
		return fmt.Errorf("can't upscale %s: failed to collect BaseInstanceName: %w", mig.GceRef(), err)
	}
	m.cache.InvalidateMigTargetSize(mig.GceRef())
	return m.GceService.CreateInstances(mig.GceRef(), baseName, delta, instancesNames)
}

func (m *gceManagerImpl) forceRefresh() error {
	m.clearMachinesCache()
	if err := m.fetchAutoMigs(); err != nil {
		klog.Errorf("Failed to fetch MIGs: %v", err)
		return err
	}
	m.refreshAutoscalingOptions()
	m.lastRefresh = time.Now()
	klog.V(2).Infof("Refreshed GCE resources, next refresh after %v", m.lastRefresh.Add(refreshInterval))
	return nil
}

func (m *gceManagerImpl) refreshAutoscalingOptions() {
	for _, mig := range m.migLister.GetMigs() {
		template, err := m.migInfoProvider.GetMigInstanceTemplate(mig.GceRef())
		if err != nil {
			klog.Warningf("Not evaluating autoscaling options for %q MIG: failed to find corresponding instance template", mig.GceRef(), err)
			continue
		}
		if template.Properties == nil {
			klog.Warningf("Failed to extract autoscaling options from %q metadata: instance template is incomplete", template.Name)
			continue
		}
		kubeEnvValue, err := getKubeEnvValueFromTemplateMetadata(template)
		if err != nil {
			klog.Warningf("Failed to extract autoscaling options from %q instance template's metadata: can't get KubeEnv: %v", template.Name, err)
			continue
		}
		options, err := extractAutoscalingOptionsFromKubeEnv(kubeEnvValue)
		if err != nil {
			klog.Warningf("Failed to extract autoscaling options from %q instance template's metadata: %v", template.Name, err)
			continue
		}
		if !reflect.DeepEqual(m.cache.GetAutoscalingOptions(mig.GceRef()), options) {
			klog.V(4).Infof("Extracted autoscaling options from %q instance template KubeEnv: %v", template.Name, options)
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
		if m.registerMig(mig) {
			changed = true
		}
		m.explicitlyConfigured[mig.GceRef()] = true
	}

	if changed {
		return m.migInfoProvider.RegenerateMigInstancesCache()
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
	}
	return mig, nil
}

// Fetch automatically discovered MIGs. These MIGs should be unregistered if
// they no longer exist in GCE.
func (m *gceManagerImpl) fetchAutoMigs() error {
	exists := make(map[GceRef]bool)
	var changed int32 = 0

	toRegister := make([]Mig, 0)
	for _, cfg := range m.migAutoDiscoverySpecs {
		links, err := m.findMigsNamed(cfg.Re)
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
				klog.V(3).Infof("Ignoring explicitly configured MIG %s in autodiscovery.", mig.GceRef().String())
				continue
			}
			toRegister = append(toRegister, mig)
		}
	}

	workqueue.ParallelizeUntil(context.Background(), m.concurrentGceRefreshes, len(toRegister), func(piece int) {
		mig := toRegister[piece]
		if m.registerMig(mig) {
			klog.V(3).Infof("Autodiscovered MIG %s", mig.GceRef().String())
			atomic.StoreInt32(&changed, int32(1))
		}
	}, workqueue.WithChunkSize(m.concurrentGceRefreshes))

	for _, mig := range m.GetMigs() {
		if !exists[mig.GceRef()] && !m.explicitlyConfigured[mig.GceRef()] {
			m.cache.UnregisterMig(mig)
			atomic.StoreInt32(&changed, int32(1))
		}
	}

	if atomic.LoadInt32(&changed) > 0 {
		return m.migInfoProvider.RegenerateMigInstancesCache()
	}

	return nil
}

// GetResourceLimiter returns resource limiter from cache.
func (m *gceManagerImpl) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	return m.cache.GetResourceLimiter()
}

func (m *gceManagerImpl) clearMachinesCache() {
	if m.machinesCacheLastRefresh.Add(machinesRefreshInterval).After(time.Now()) {
		return
	}

	m.cache.InvalidateAllMachines()
	nextRefresh := time.Now()
	m.machinesCacheLastRefresh = nextRefresh
	klog.V(2).Infof("Cleared machine types cache, next clear after %v", nextRefresh)
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

func (m *gceManagerImpl) findMigsNamed(name *regexp.Regexp) ([]string, error) {
	if m.regional {
		return m.findMigsInRegion(m.location, name)
	}
	return m.GceService.FetchMigsWithName(m.location, name)
}

func (m *gceManagerImpl) getZones(region string) ([]string, error) {
	zones, err := m.GceService.FetchZones(region)
	if err != nil {
		return nil, fmt.Errorf("cannot get zones for GCE region %s: %v", region, err)
	}
	return zones, nil
}

func (m *gceManagerImpl) findMigsInRegion(region string, name *regexp.Regexp) ([]string, error) {
	links := make([]string, 0)
	zones, err := m.getZones(region)
	if err != nil {
		return nil, err
	}

	zoneLinks := make([][]string, len(zones))
	errors := make([]error, len(zones))
	workqueue.ParallelizeUntil(context.Background(), len(zones), len(zones), func(piece int) {
		zoneLinks[piece], errors[piece] = m.GceService.FetchMigsWithName(zones[piece], name)
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
func (m *gceManagerImpl) GetMigOptions(mig Mig, defaults config.NodeGroupAutoscalingOptions) *config.NodeGroupAutoscalingOptions {
	migRef := mig.GceRef()
	options := m.cache.GetAutoscalingOptions(migRef)
	if options == nil {
		return &defaults
	}

	if opt, ok := getFloat64Option(options, migRef.Name, config.DefaultScaleDownUtilizationThresholdKey); ok {
		defaults.ScaleDownUtilizationThreshold = opt
	}
	if opt, ok := getFloat64Option(options, migRef.Name, config.DefaultScaleDownGpuUtilizationThresholdKey); ok {
		defaults.ScaleDownGpuUtilizationThreshold = opt
	}
	if opt, ok := getDurationOption(options, migRef.Name, config.DefaultScaleDownUnneededTimeKey); ok {
		defaults.ScaleDownUnneededTime = opt
	}
	if opt, ok := getDurationOption(options, migRef.Name, config.DefaultScaleDownUnreadyTimeKey); ok {
		defaults.ScaleDownUnreadyTime = opt
	}

	return &defaults
}

// GetMigTemplateNode constructs a node from GCE instance template of the given MIG.
func (m *gceManagerImpl) GetMigTemplateNode(mig Mig) (*apiv1.Node, error) {
	template, err := m.migInfoProvider.GetMigInstanceTemplate(mig.GceRef())
	if err != nil {
		return nil, err
	}
	machineType, err := m.migInfoProvider.GetMigMachineType(mig.GceRef())
	if err != nil {
		return nil, err
	}
	migOsInfo, err := m.templates.MigOsInfo(mig.Id(), template)
	if err != nil {
		return nil, err
	}
	return m.templates.BuildNodeFromTemplate(mig, migOsInfo, template, machineType.CPU, machineType.Memory, nil, m.reserved)
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
