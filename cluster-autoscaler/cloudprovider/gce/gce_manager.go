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
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/opentracing/opentracing-go"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config/dynamic"
	"k8s.io/autoscaler/cluster-autoscaler/utils/units"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	provider_gce "k8s.io/kubernetes/pkg/cloudprovider/providers/gce"

	"cloud.google.com/go/compute/metadata"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	gce "google.golang.org/api/compute/v1"
	gcfg "gopkg.in/gcfg.v1"
	"k8s.io/klog"
)

const (
	refreshInterval         = 1 * time.Minute
	machinesRefreshInterval = 1 * time.Hour
	httpTimeout             = 30 * time.Second
	scaleToZeroSupported    = true
)

var (
	defaultOAuthScopes = []string{
		"https://www.googleapis.com/auth/compute",
		"https://www.googleapis.com/auth/devstorage.read_only",
		"https://www.googleapis.com/auth/service.management.readonly",
		"https://www.googleapis.com/auth/servicecontrol",
	}
)

// GceManager handles GCE communication and data caching.
type GceManager interface {
	// Refresh triggers refresh of cached resources.
	Refresh(ctx context.Context) error
	// Cleanup cleans up open resources before the cloud provider is destroyed, i.e. go routines etc.
	Cleanup(ctx context.Context) error

	// GetMigs returns list of registered MIGs.
	GetMigs(ctx context.Context) []*MigInformation
	// GetMigNodes returns mig nodes.
	GetMigNodes(ctx context.Context, mig Mig) ([]cloudprovider.Instance, error)
	// GetMigForInstance returns MIG to which the given instance belongs.
	GetMigForInstance(ctx context.Context, instance *GceRef) (Mig, error)
	// GetMigTemplateNode returns a template node for MIG.
	GetMigTemplateNode(ctx context.Context, mig Mig) (*apiv1.Node, error)
	// GetResourceLimiter returns resource limiter.
	GetResourceLimiter(ctx context.Context) (*cloudprovider.ResourceLimiter, error)
	// GetMigSize gets MIG size.
	GetMigSize(ctx context.Context, mig Mig) (int64, error)

	// SetMigSize sets MIG size.
	SetMigSize(ctx context.Context, mig Mig, size int64) error
	// DeleteInstances deletes the given instances. All instances must be controlled by the same MIG.
	DeleteInstances(ctx context.Context, instances []*GceRef) error
}

type gceManagerImpl struct {
	cache                    GceCache
	lastRefresh              time.Time
	machinesCacheLastRefresh time.Time

	GceService AutoscalingGceClient

	location              string
	projectId             string
	templates             *GceTemplateBuilder
	interrupt             chan struct{}
	regional              bool
	explicitlyConfigured  map[GceRef]bool
	migAutoDiscoverySpecs []cloudprovider.MIGAutoDiscoveryConfig
}

// CreateGceManager constructs GceManager object.
func CreateGceManager(ctx context.Context, configReader io.Reader, discoveryOpts cloudprovider.NodeGroupDiscoveryOptions, regional bool) (GceManager, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "CreateGceManager")
	defer span.Finish()

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
	gceService, err := NewAutoscalingGceClientV1(client, projectId)
	if err != nil {
		return nil, err
	}
	manager := &gceManagerImpl{
		cache:                NewGceCache(gceService),
		GceService:           gceService,
		location:             location,
		regional:             regional,
		projectId:            projectId,
		templates:            &GceTemplateBuilder{},
		interrupt:            make(chan struct{}),
		explicitlyConfigured: make(map[GceRef]bool),
	}

	if err := manager.fetchExplicitMigs(ctx, discoveryOpts.NodeGroupSpecs); err != nil {
		return nil, fmt.Errorf("failed to fetch MIGs: %v", err)
	}
	if manager.migAutoDiscoverySpecs, err = discoveryOpts.ParseMIGAutoDiscoverySpecs(); err != nil {
		return nil, err
	}

	if err := manager.forceRefresh(ctx); err != nil {
		return nil, err
	}

	go wait.Until(func() {
		if err := manager.cache.RegenerateInstancesCache(ctx); err != nil {
			klog.Errorf("Error while regenerating Mig cache: %v", err)
		}
	}, time.Hour, manager.interrupt)

	return manager, nil
}

// Cleanup closes the channel to stop the goroutine refreshing cache.
func (m *gceManagerImpl) Cleanup(ctx context.Context) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "gceManagerImpl.Cleanup")
	defer span.Finish()

	close(m.interrupt)
	return nil
}

// registerMig registers mig in GceManager. Returns true if the node group didn't exist before or its config has changed.
func (m *gceManagerImpl) registerMig(ctx context.Context, mig Mig) bool {
	span, ctx := opentracing.StartSpanFromContext(ctx, "gceManagerImpl.registerMig")
	defer span.Finish()

	changed := m.cache.RegisterMig(mig)
	if changed {
		// Try to build a node from template to validate that this group
		// can be scaled up from 0 nodes.
		// We may never need to do it, so just log error if it fails.
		if _, err := m.GetMigTemplateNode(ctx, mig); err != nil {
			klog.Errorf("Can't build node from template for %s, won't be able to scale from 0: %v", mig.GceRef().String(), err)
		}
	}
	return changed
}

// GetMigSize gets MIG size.
func (m *gceManagerImpl) GetMigSize(ctx context.Context, mig Mig) (int64, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "gceManagerImpl.GetMigSize")
	defer span.Finish()

	if migSize, found := m.cache.GetMigTargetSize(mig.GceRef()); found {
		return migSize, nil
	}
	targetSize, err := m.GceService.FetchMigTargetSize(ctx, mig.GceRef())
	if err != nil {
		return -1, err
	}
	m.cache.SetMigTargetSize(mig.GceRef(), targetSize)
	return targetSize, nil
}

// SetMigSize sets MIG size.
func (m *gceManagerImpl) SetMigSize(ctx context.Context, mig Mig, size int64) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "gceManagerImpl.SetMigSize")
	defer span.Finish()

	klog.V(0).Infof("Setting mig size %s to %d", mig.Id(), size)
	m.cache.InvalidateTargetSizeCacheForMig(mig.GceRef())
	return m.GceService.ResizeMig(ctx, mig.GceRef(), size)
}

// DeleteInstances deletes the given instances. All instances must be controlled by the same MIG.
func (m *gceManagerImpl) DeleteInstances(ctx context.Context, instances []*GceRef) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "gceManagerImpl.DeleteInstances")
	defer span.Finish()

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
	m.cache.InvalidateTargetSizeCacheForMig(commonMig.GceRef())
	return m.GceService.DeleteInstances(ctx, commonMig.GceRef(), instances)
}

// GetMigs returns list of registered MIGs.
func (m *gceManagerImpl) GetMigs(ctx context.Context) []*MigInformation {
	span, ctx := opentracing.StartSpanFromContext(ctx, "gceManagerImpl.GetMigs")
	defer span.Finish()

	return m.cache.GetMigs(ctx)
}

// GetMigForInstance returns MIG to which the given instance belongs.
func (m *gceManagerImpl) GetMigForInstance(ctx context.Context, instance *GceRef) (Mig, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "gceManagerImpl.GetMigForInstance")
	defer span.Finish()

	return m.cache.GetMigForInstance(ctx, instance)
}

// GetMigNodes returns mig nodes.
func (m *gceManagerImpl) GetMigNodes(ctx context.Context, mig Mig) ([]cloudprovider.Instance, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "gceManagerImpl.GetMigNodes")
	defer span.Finish()

	return m.GceService.FetchMigInstances(ctx, mig.GceRef())
}

// Refresh triggers refresh of cached resources.
func (m *gceManagerImpl) Refresh(ctx context.Context) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "gceManagerImpl.Refresh")
	defer span.Finish()

	m.cache.InvalidateTargetSizeCache()
	if m.lastRefresh.Add(refreshInterval).After(time.Now()) {
		return nil
	}
	return m.forceRefresh(ctx)
}

func (m *gceManagerImpl) forceRefresh(ctx context.Context) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "gceManagerImpl.forceRefresh")
	defer span.Finish()

	m.clearMachinesCache(ctx)
	if err := m.fetchAutoMigs(ctx); err != nil {
		klog.Errorf("Failed to fetch MIGs: %v", err)
		return err
	}
	m.lastRefresh = time.Now()
	klog.V(2).Infof("Refreshed GCE resources, next refresh after %v", m.lastRefresh.Add(refreshInterval))
	return nil
}

// Fetch explicitly configured MIGs. These MIGs should never be unregistered
// during refreshes, even if they no longer exist in GCE.
func (m *gceManagerImpl) fetchExplicitMigs(ctx context.Context, specs []string) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "gceManagerImpl.fetchExplicitMigs")
	defer span.Finish()

	changed := false
	for _, spec := range specs {
		mig, err := m.buildMigFromFlag(ctx, spec)
		if err != nil {
			return err
		}
		if m.registerMig(ctx, mig) {
			changed = true
		}
		m.explicitlyConfigured[mig.GceRef()] = true
	}

	if changed {
		return m.cache.RegenerateInstancesCache(ctx)
	}
	return nil
}

func (m *gceManagerImpl) buildMigFromFlag(ctx context.Context, flag string) (Mig, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "gceManagerImpl.buildMigFromFlag")
	defer span.Finish()

	s, err := dynamic.SpecFromString(flag, scaleToZeroSupported)
	if err != nil {
		return nil, fmt.Errorf("failed to parse node group spec: %v", err)
	}
	return m.buildMigFromSpec(ctx, s)
}

func (m *gceManagerImpl) buildMigFromAutoCfg(ctx context.Context, link string, cfg cloudprovider.MIGAutoDiscoveryConfig) (Mig, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "gceManagerImpl.buildMigFromAutoCfg")
	defer span.Finish()

	s := &dynamic.NodeGroupSpec{
		Name:               link,
		MinSize:            cfg.MinSize,
		MaxSize:            cfg.MaxSize,
		SupportScaleToZero: scaleToZeroSupported,
	}
	return m.buildMigFromSpec(ctx, s)
}

func (m *gceManagerImpl) buildMigFromSpec(ctx context.Context, s *dynamic.NodeGroupSpec) (Mig, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "gceManagerImpl.buildMigFromSpec")
	defer span.Finish()

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
func (m *gceManagerImpl) fetchAutoMigs(ctx context.Context) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "gceManagerImpl.fetchAutoMigs")
	defer span.Finish()

	exists := make(map[GceRef]bool)
	changed := false
	for _, cfg := range m.migAutoDiscoverySpecs {
		links, err := m.findMigsNamed(ctx, cfg.Re)
		if err != nil {
			return fmt.Errorf("cannot autodiscover managed instance groups: %v", err)
		}
		for _, link := range links {
			mig, err := m.buildMigFromAutoCfg(ctx, link, cfg)
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
			if m.registerMig(ctx, mig) {
				klog.V(3).Infof("Autodiscovered MIG %s using regexp %s", mig.GceRef().String(), cfg.Re.String())
				changed = true
			}
		}
	}

	for _, mig := range m.GetMigs(ctx) {
		if !exists[mig.Config.GceRef()] && !m.explicitlyConfigured[mig.Config.GceRef()] {
			m.cache.UnregisterMig(mig.Config)
			changed = true
		}
	}

	if changed {
		return m.cache.RegenerateInstancesCache(ctx)
	}

	return nil
}

// GetResourceLimiter returns resource limiter from cache.
func (m *gceManagerImpl) GetResourceLimiter(ctx context.Context) (*cloudprovider.ResourceLimiter, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "gceManagerImpl.GetResourceLimiter")
	defer span.Finish()

	return m.cache.GetResourceLimiter(ctx)
}

func (m *gceManagerImpl) clearMachinesCache(ctx context.Context) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "gceManagerImpl.clearMachinesCache")
	defer span.Finish()

	if m.machinesCacheLastRefresh.Add(machinesRefreshInterval).After(time.Now()) {
		return
	}

	machinesCache := make(map[MachineTypeKey]*gce.MachineType)
	m.cache.SetMachinesCache(ctx, machinesCache)
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

func (m *gceManagerImpl) findMigsNamed(ctx context.Context, name *regexp.Regexp) ([]string, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "gceManagerImpl.findMigsNamed")
	defer span.Finish()

	if m.regional {
		return m.findMigsInRegion(ctx, m.location, name)
	}
	return m.GceService.FetchMigsWithName(ctx, m.location, name)
}

func (m *gceManagerImpl) getZones(ctx context.Context, region string) ([]string, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "gceManagerImpl.getZones")
	defer span.Finish()

	zones, err := m.GceService.FetchZones(ctx, region)
	if err != nil {
		return nil, fmt.Errorf("cannot get zones for GCE region %s: %v", region, err)
	}
	return zones, nil
}

func (m *gceManagerImpl) findMigsInRegion(ctx context.Context, region string, name *regexp.Regexp) ([]string, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "gceManagerImpl.findMigsInRegion")
	defer span.Finish()

	links := make([]string, 0)
	zones, err := m.getZones(ctx, region)
	if err != nil {
		return nil, err
	}
	for _, z := range zones {
		zl, err := m.GceService.FetchMigsWithName(ctx, z, name)
		if err != nil {
			return nil, err
		}
		for _, link := range zl {
			links = append(links, link)
		}
	}

	return links, nil
}

// GetMigTemplateNode constructs a node from GCE instance template of the given MIG.
func (m *gceManagerImpl) GetMigTemplateNode(ctx context.Context, mig Mig) (*apiv1.Node, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "gceManagerImpl.GetMigTemplateNode")
	defer span.Finish()

	template, err := m.GceService.FetchMigTemplate(ctx, mig.GceRef())
	if err != nil {
		return nil, err
	}
	cpu, mem, err := m.getCpuAndMemoryForMachineType(ctx, template.Properties.MachineType, mig.GceRef().Zone)
	if err != nil {
		return nil, err
	}
	return m.templates.BuildNodeFromTemplate(mig, template, cpu, mem)
}

func (m *gceManagerImpl) getCpuAndMemoryForMachineType(ctx context.Context, machineType string, zone string) (cpu int64, mem int64, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "gceManagerImpl.getCpuAndMemoryForMachineType")
	defer span.Finish()

	if strings.HasPrefix(machineType, "custom-") {
		return parseCustomMachineType(machineType)
	}
	machine := m.cache.GetMachineFromCache(machineType, zone)
	if machine == nil {
		machine, err = m.GceService.FetchMachineType(ctx, zone, machineType)
		if err != nil {
			return 0, 0, err
		}
		m.cache.AddMachineToCache(machineType, zone, machine)
	}
	return machine.GuestCpus, machine.MemoryMb * units.MiB, nil
}

func parseCustomMachineType(machineType string) (cpu, mem int64, err error) {
	// example custom-2-2816
	var count int
	count, err = fmt.Sscanf(machineType, "custom-%d-%d", &cpu, &mem)
	if err != nil {
		return
	}
	if count != 2 {
		return 0, 0, fmt.Errorf("failed to parse all params in %s", machineType)
	}
	// Mb to bytes
	mem = mem * units.MiB
	return
}
